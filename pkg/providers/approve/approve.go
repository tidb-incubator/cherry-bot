package approve

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

type Approve struct {
	owner   string
	repo    string
	ready   bool
	approve bool
	opr     *operator.Operator
	cfg     *config.RepoConfig
}

// Init create cherry pick middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) *Approve {
	n := Approve{
		owner:   repo.Owner,
		repo:    repo.Repo,
		ready:   false,
		approve: repo.PullApprove,
		opr:     opr,
		cfg:     repo,
	}
	return &n
}

func (c *Approve) Ready() {
	c.ready = true
}
