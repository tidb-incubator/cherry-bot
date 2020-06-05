package autoupdate

import (
	"sync"

	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"

	"github.com/google/go-github/v31/github"
)

// AutoUpdate defines methods of auto
type AutoUpdate interface {
	ProcessPullRequestEvent(event *github.PullRequestEvent)
	ProcessIssueCommentEvent(event *github.IssueCommentEvent)
}

type autoUpdate struct {
	owner           string
	watchedRepo     string
	updateOwner     string
	updateRepo      string
	targetMap       map[string]string
	updateScript    string
	updateAutoMerge bool
	opr             *operator.Operator
	cfg             *config.RepoConfig
	sync.Mutex
}

// Init create cherry pick middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) AutoUpdate {
	c := autoUpdate{
		owner:           repo.Owner,
		watchedRepo:     repo.Repo,
		updateOwner:     repo.UpdateOwner,
		updateRepo:      repo.UpdateRepo,
		targetMap:       repo.UpdateTargetMap,
		updateScript:    repo.UpdateScript,
		updateAutoMerge: repo.UpdateAutoMerge,
		opr:             opr,
		cfg:             repo,
	}
	return &c
}
