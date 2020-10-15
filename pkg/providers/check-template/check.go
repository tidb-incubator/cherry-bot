package check_template

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

type check struct {
	owner string
	repo  string
	opr   *operator.Operator
	cfg   *config.RepoConfig
}

func Init(repoCfg *config.RepoConfig, opr *operator.Operator) *check {
	return &check{
		owner: repoCfg.Owner,
		repo:  repoCfg.Repo,
		opr:   opr,
		cfg:   repoCfg,
	}
}
