package autoupdate

import (
	"context"
	"strings"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/ngaut/log"
	"github.com/pkg/errors"
)

const (
	updateCommand = "/auto-update"
)

func (au *autoUpdate) ProcessPullRequestEvent(event *github.PullRequestEvent) {
	var err error

	if *event.Action == "closed" {
		err = au.CommitUpdate(event.PullRequest)
	}

	if err != nil {
		util.Error(errors.Wrap(err, "auto update process pull request event"))
	}
}

func (au *autoUpdate) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	var (
		login = event.GetSender().GetLogin()
		issue = event.GetIssue()
	)
	if login == au.opr.Config.Github.Bot {
		return
	}
	if strings.Trim(event.GetComment().GetBody(), " ") != updateCommand {
		return
	}
	var (
		err  error
		pull *github.PullRequest
	)

	if !au.opr.Member.IfMember(login) {
		log.Infof("%s#%d %s is not a member", au.watchedRepo, issue.GetNumber(), login)
		return
	}

	pull, _, err = au.opr.Github.PullRequests.Get(context.Background(), au.owner, au.watchedRepo, issue.GetNumber())
	if err != nil {
		util.Error(errors.Wrap(err, "auto update process pull request event"))
		return
	}

	err = au.CommitUpdate(pull)
	// switch *event.Action {
	// case "closed":
	// 	{
	// 		err = au.CommitUpdate(event.PullRequest)
	// 	}
	// }

	if err != nil {
		util.Error(errors.Wrap(err, "auto update process pull request event"))
	}
}
