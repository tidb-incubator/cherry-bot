package autoupdate

import (
	"context"
	"strings"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
)

const (
	updateCommand = "/auto-update"
)

func (au *autoUpdate) ProcessPullRequestEvent(event *github.PullRequestEvent) {
	var err error

	switch *event.Action {
	case "closed":
		{
			err = au.CommitUpdate(event.PullRequest)
		}
	}

	if err != nil {
		util.Error(errors.Wrap(err, "auto update process pull request event"))
	}
}

func (au *autoUpdate) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if strings.Trim(event.GetComment().GetBody(), " ") != updateCommand {
		return
	}
	var (
		err  error
		pull *github.PullRequest
	)

	pull, _, err = au.opr.Github.PullRequests.Get(context.Background(), au.owner, au.watchedRepo, event.GetIssue().GetNumber())
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
