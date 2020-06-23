package label

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers"
)

type Label struct {
	owner    string
	repo     string
	ready    bool
	approve  bool
	provider *providers.Provider
}

// Init create cherry pick middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) *Label {
	n := Label{
		owner:    repo.Owner,
		repo:     repo.Repo,
		ready:    false,
		approve:  repo.PullApprove,
		provider: providers.Init(repo, opr),
	}
	return &n
}

func (c *Label) Ready() {
	c.ready = true
}
