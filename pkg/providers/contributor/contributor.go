package contributor

import (
	"sync"

	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

// Contributor struct
type Contributor struct {
	sync.Mutex
	owner string
	repo  string
	opr   *operator.Operator
	cfg   *config.RepoConfig
}

// Init create contributor middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) *Contributor {
	c := Contributor{
		owner: repo.Owner,
		repo:  repo.Repo,
		opr:   opr,
		cfg:   repo,
	}
	return &c
}

func (c *Contributor) Ready() {
}
