package prlimit

import (
	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
)

func (p *prLimit) ProcessPullRequestEvent(event *github.PullRequestEvent) {
	var err error

	if *event.Action == "opened" || *event.Action == "reopened" {
		err = p.processOpenedPR(event.PullRequest)
	}

	if err != nil {
		util.Error(errors.Wrap(err, "cherry picker process pull request event"))
	}
}
