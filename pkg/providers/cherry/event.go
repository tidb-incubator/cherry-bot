package cherry

import (
	"context"
	"regexp"
	"strings"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
)

const cherryPickTrigger = "/run-cherry-picker"

func (cherry *cherry) ProcessPullRequest(pr *github.PullRequest) {
	// status update
	match, err := regexp.MatchString(`\(#[0-9]+\)$`, *pr.Title)
	if err != nil {
		util.Error(errors.Wrap(err, "process cherry pick"))
	} else {
		if match {
			util.Error(cherry.createCherryPick(pr))
		} else {
			for _, label := range pr.Labels {
				util.Error(cherry.commitLabel(pr, *label.Name))
			}
			if pr.MergedAt != nil {
				util.Error(cherry.commitMerge(pr))
			}
		}
	}
}

func (cherry *cherry) ProcessPullRequestEvent(event *github.PullRequestEvent) {
	var err error

	switch *event.Action {
	case "labeled":
		{
			err = cherry.commitLabel(event.GetPullRequest(), *event.Label.Name)
		}
	case "unlabeled":
		{
			{
				err = cherry.removeLabel(event.GetPullRequest(), event.GetLabel().GetName())
			}
		}
	case "closed":
		{
			util.Printf("process cp closed event %s/%s#%d", cherry.owner, cherry.repo, event.GetPullRequest().GetNumber())
			err = cherry.commitMerge(event.PullRequest)
		}
	}

	if err != nil {
		util.Error(errors.Wrap(err, "cherry picker process pull request event"))
	}
}

func (cherry *cherry) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if strings.Trim(event.GetComment().GetBody(), " ") != cherryPickTrigger {
		return
	}

	var (
		login  = event.GetSender().GetLogin()
		number = event.GetIssue().GetNumber()
	)
	if cherry.opr.Member.IfMember(login) || event.GetIssue().GetUser().GetLogin() == event.GetComment().GetUser().GetLogin() {
		pr, _, err := cherry.opr.Github.PullRequests.Get(context.Background(),
			cherry.owner, cherry.repo, number)
		if err != nil {
			util.Error(errors.Wrap(err, "issue comment get PR"))
			return
		}
		if pr.MergedAt == nil {
			return
		}
		for _, label := range pr.Labels {
			target, version, err := cherry.getTarget(label.GetName())
			util.Println("label is", label.GetName())
			if err == nil {
				util.Println("ready to cherry pick via command", target, version)
				if err := cherry.cherryPick(pr, target, version, false); err != nil {
					util.Error(errors.Wrap(err, "commit label"))
				}
			}
		}
	} else {
		util.Printf("%s/%s#%d %s don't have access to run %s", cherry.owner, cherry.repo, number, login, cherryPickTrigger)
	}
}
