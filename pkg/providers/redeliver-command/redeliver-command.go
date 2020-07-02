package command

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

type Redeliver struct {
	repo *config.RepoConfig
	opr  *operator.Operator
}

// Init create cherry pick middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) *Redeliver {
	return &Redeliver{
		repo: repo,
		opr:  opr,
	}
}
