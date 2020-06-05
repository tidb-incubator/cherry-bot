package label

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"

	"github.com/google/go-github/v32/github"
)

// Label defines methods of label check
type Label interface {
	Ready()
	ProcessPullRequest(pr *github.PullRequest)
	ProcessPullRequestEvent(event *github.PullRequestEvent)
	ProcessIssuesEvent(event *github.IssuesEvent)
	MonthCheck(pr *github.Issue) (string, error)
}

type label struct {
	owner   string
	repo    string
	ready   bool
	rule    string
	release string
	dryrun  bool
	opr     *operator.Operator
	cfg     *config.RepoConfig
}

// Init create label check middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) Label {
	l := label{
		owner:   repo.Owner,
		repo:    repo.Repo,
		ready:   false,
		rule:    repo.Rule,
		release: repo.Release,
		dryrun:  repo.Dryrun,
		opr:     opr,
		cfg:     repo,
	}
	return &l
}

func (l *label) Ready() {
	l.ready = true
	l.restartJobs()
}
