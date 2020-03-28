package assign

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

type Assign struct {
	owner string
	repo  string
	opr   *operator.Operator
	cfg   *config.RepoConfig
}

func Init(repoCfg *config.RepoConfig, opr *operator.Operator) *Assign {
	return &Assign{
		owner: repoCfg.Owner,
		repo:  repoCfg.Repo,
		opr:   opr,
		cfg:   repoCfg,
	}
}
