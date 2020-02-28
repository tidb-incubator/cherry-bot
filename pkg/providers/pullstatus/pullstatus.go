package pullstatus

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"

	"github.com/google/go-github/v29/github"
)

// Label defines methods of label check
type PullStatus interface {
	Ready()
	ProcessPullRequestEvent(event *github.PullRequestEvent)
	ProcessPullRequestReviewEvent(event *github.PullRequestReviewEvent)
	ProcessPullRequestReviewCommentEvent(event *github.PullRequestReviewCommentEvent)
	ProcessIssueCommentEvent(event *github.IssueCommentEvent)
}

type pullStatus struct {
	owner string
	repo  string
	ready bool
	opr   *operator.Operator
	cfg   *config.RepoConfig
}

// Init create label check middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) PullStatus {
	p := pullStatus{
		owner: repo.Owner,
		repo:  repo.Repo,
		ready: false,
		opr:   opr,
		cfg:   repo,
	}

	p.startPolling()
	return &p
}

func (p *pullStatus) Ready() {
	p.ready = true
	//p.restartJobs()
}
