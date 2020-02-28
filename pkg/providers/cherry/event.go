package cherry

import (
	"context"
	"regexp"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
)

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
	// case "labeled":
	// 	{
	// 		err = cherry.commitLabel(event.GetPullRequest(), *event.Label.Name)
	// 	}
	// case "unlabeled":
	// 	{
	// 		{
	// 			err = cherry.removeLabel(event.GetPullRequest(), event.GetLabel().GetName())
	// 		}
	// 	}
	case "closed":
		{
			err = cherry.commitMerge(event.PullRequest)
		}
	}

	if err != nil {
		util.Error(errors.Wrap(err, "cherry picker process pull request event"))
	}
}

func (cherry *cherry) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if *event.Comment.Body != "/run-cherry-picker" {
		return
	}
	isMember, _, err := cherry.opr.Github.Organizations.IsMember(context.Background(),
		"pingcap", *event.Comment.User.Login)
	if err != nil {
		isMember = false
	}
	if isMember || *event.Issue.User.Login == *event.Comment.User.Login {
		pr, _, err := cherry.opr.Github.PullRequests.Get(context.Background(),
			cherry.owner, cherry.repo, *event.Issue.Number)
		if err != nil {
			util.Error(errors.Wrap(err, "issue comment get PR"))
			return
		}
		if pr.MergedAt == nil {
			return
		}
		for _, label := range pr.Labels {
			target, version, err := cherry.getTarget(*label.Name)
			util.Println("label is", *label.Name)
			if err == nil {
				util.Println("ready to cherry pick via command", target, version)
				if err := cherry.cherryPick(pr, target, version, false); err != nil {
					util.Error(errors.Wrap(err, "commit label"))
				}
			}
		}
	}
}
