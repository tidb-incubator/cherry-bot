package add_template

import (
"github.com/pingcap-incubator/cherry-bot/config"
"github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

type Comment struct {
	owner string
	repo  string
	opr   *operator.Operator
	cfg   *config.RepoConfig
}

func Init(repoCfg *config.RepoConfig, opr *operator.Operator) *Comment {
	return &Comment{
		owner: repoCfg.Owner,
		repo:  repoCfg.Repo,
		opr:   opr,
		cfg:   repoCfg,
	}
}

