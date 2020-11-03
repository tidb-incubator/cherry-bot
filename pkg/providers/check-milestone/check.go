package checkMilestone

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

type Check struct {
	owner string
	repo  string
	opr   *operator.Operator
	cfg   *config.RepoConfig
}

func Init(repoCfg *config.RepoConfig, opr *operator.Operator) *Check {

	check := &Check{
		owner: repoCfg.Owner,
		repo:  repoCfg.Repo,
		opr:   opr,
		cfg:   repoCfg,
	}

	repos := check.opr.Config.Check.WhiteList
	for i := 0; i < len(repos); i++ {
		fullName := check.owner + "/" + check.repo
		if fullName == repos[i] {
			go func(repo string) {
				check.loopCheckRepo(repo)
			}(repos[i])
		}
	}

	return check
}
