package prlimit

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"

	"github.com/google/go-github/v32/github"
)

// PrLimit defines methods of PR limit module
type PrLimit interface {
	Ready()
	ProcessPullRequestEvent(event *github.PullRequestEvent)
	GetAllowList() ([]string, error)
	AddAllowList(username string) error
	RemoveAllowList(username string) error
	GetBlockList() ([]string, error)
	AddBlockList(username string) error
	RemoveBlockList(username string) error
}

type prLimit struct {
	owner string
	repo  string
	ready bool
	opr   *operator.Operator
	cfg   *config.RepoConfig
}

// Init create PR limit middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) PrLimit {
	p := prLimit{
		owner: repo.Owner,
		repo:  repo.Repo,
		ready: false,
		opr:   opr,
		cfg:   repo,
	}
	return &p
}

func (p *prLimit) Ready() {
	p.ready = true
}
