package auto_update

import (
	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
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
