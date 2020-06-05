package merge

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"

	"github.com/google/go-github/v31/github"
)

// Merge defines methods of auto merge
type Merge interface {
	Ready()
	ProcessPullRequestEvent(event *github.PullRequestEvent)
	ProcessIssueCommentEvent(event *github.IssueCommentEvent)
	GetWhiteList() ([]string, error)
	AddWhiteList(username string) error
	RemoveWhiteList(username string) error
}

type merge struct {
	owner string
	repo  string
	ready bool
	opr   *operator.Operator
	cfg   *config.RepoConfig
	list  []*github.PullRequest
}

// Init create PR limit middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) Merge {
	m := merge{
		owner: repo.Owner,
		repo:  repo.Repo,
		ready: false,
		opr:   opr,
		cfg:   repo,
		list:  []*github.PullRequest{},
	}
	m.startPolling()
	return &m
}

func (m *merge) Ready() {
	m.ready = true
}
