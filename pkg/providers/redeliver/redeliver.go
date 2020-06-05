package redeliver

import (
	"sync"

	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"

	"github.com/google/go-github/v32/github"
)

// Redeliver defines methods of issue redeliver
type Redeliver interface {
	Ready()
	ProcessIssuesEvent(event *github.IssuesEvent)
	ProcessIssueCommentEvent(event *github.IssueCommentEvent)
}

type redeliver struct {
	owner string
	repo  string
	ready bool
	opr   *operator.Operator
	cfg   *config.RepoConfig
	sync.Mutex
}

// Init create PR limit middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) Redeliver {
	r := redeliver{
		owner: repo.Owner,
		repo:  repo.Repo,
		ready: false,
		opr:   opr,
		cfg:   repo,
	}
	return &r
}

func (r *redeliver) Ready() {
	r.ready = true
}
