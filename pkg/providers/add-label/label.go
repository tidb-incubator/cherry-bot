package label

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

type Label struct {
	owner   string
	repo    string
	ready   bool
	approve bool
	opr     *operator.Operator
	cfg     *config.RepoConfig
}

// Init create cherry pick middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) *Label {
	n := Label{
		owner:   repo.Owner,
		repo:    repo.Repo,
		ready:   false,
		approve: repo.PullApprove,
		cfg:     repo,
		opr:     opr,
	}
	return &n
}

func (c *Label) Ready() {
	c.ready = true
}
