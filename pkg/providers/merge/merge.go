package merge

import (
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers"
)

// Merge defines methods of auto merge
type Merge interface {
	Ready()
	ProcessPullRequestEvent(event *github.PullRequestEvent)
	ProcessIssueCommentEvent(event *github.IssueCommentEvent)
	GetAllowList() ([]string, error)
	AddAllowList(username string) error
	RemoveAllowList(username string) error
}

type merge struct {
	owner    string
	repo     string
	ready    bool
	provider *providers.Provider
	list     []*github.PullRequest
}

// Init create PR limit middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) Merge {
	m := merge{
		owner:    repo.Owner,
		repo:     repo.Repo,
		ready:    false,
		provider: providers.Init(repo, opr),
		list:     []*github.PullRequest{},
	}
	m.startPolling()
	return &m
}

func (m *merge) Ready() {
	m.ready = true
}
