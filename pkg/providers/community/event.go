package community

import (
	"github.com/google/go-github/v32/github"
	"github.com/ngaut/log"
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

type CommunityCmd struct {
	owner string
	repo  string
	ready bool
	opr   *operator.Operator
	cfg   *config.RepoConfig
}

// Init create cherry pick middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) *CommunityCmd {
	n := CommunityCmd{
		owner: repo.Owner,
		repo:  repo.Repo,
		ready: false,
		cfg:   repo,
		opr:   opr,
	}
	return &n
}

func (c *CommunityCmd) Ready() {
	c.ready = true
}

func (c *CommunityCmd) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if event.GetAction() != "created" {
		return
	}
	if err := c.processComment(event, event.Comment.GetBody()); err != nil {
		log.Error("process community command failed", event.Comment.GetBody(), err)
	} else {
		log.Info("process community command success", event.Comment.GetBody())
	}
}

func (c *CommunityCmd) processComment(event *github.IssueCommentEvent, comment string) error {
	pr := event.GetIssue()
	if pr.GetPullRequestLinks() != nil {
		err := c.ptal(pr.GetNumber(), comment)
		if err != nil {
			return err
		}
	}
	return nil
}
