package cherry

import (
	"sync"
	"time"

	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"

	"github.com/google/go-github/v32/github"
)

// Cherry defines methods of cherry pick
type Cherry interface {
	Ready()
	ProcessPullRequest(pr *github.PullRequest)
	ProcessPullRequestEvent(event *github.PullRequestEvent)
	ProcessIssueCommentEvent(event *github.IssueCommentEvent)
	MonthCheck(pr *github.PullRequest) ([]string, error)
}

type cherry struct {
	sync.Mutex
	owner                   string
	repo                    string
	ready                   bool
	rule                    string
	release                 string
	typeLabel               string
	ignoreLabel             string
	dryrun                  bool
	forkedRepoCollaborators map[string]struct{}
	collaboratorInvitation  map[string]time.Time
	opr                     *operator.Operator
	cfg                     *config.RepoConfig
}

// Init create cherry pick middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) Cherry {
	c := cherry{
		owner:                   repo.Owner,
		repo:                    repo.Repo,
		ready:                   false,
		rule:                    repo.Rule,
		release:                 repo.Release,
		typeLabel:               repo.TypeLabel,
		ignoreLabel:             repo.IgnoreLabel,
		dryrun:                  repo.Dryrun,
		forkedRepoCollaborators: make(map[string]struct{}),
		collaboratorInvitation:  make(map[string]time.Time),
		opr:                     opr,
		cfg:                     repo,
	}
	go c.runLoadCollaborators()
	return &c
}

func (c *cherry) Ready() {
	c.ready = true
}
