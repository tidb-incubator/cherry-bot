package merge

import (
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

const (
	autoMergeCommand = "/run-auto-merge"
	autoMergeAlias   = "/merge"
	noAccessComment  = "Sorry @%s, you don't have permission to trigger auto merge event on this branch."
)

var accessAssociation = []string{"COLLABORATOR", "MEMBER", "OWNER"}

func (m *merge) processLabelEvent(event *github.PullRequestEvent) {
	if event.GetLabel().GetName() != m.cfg.CanMergeLabel {
		return
	}
	access, pull, err := m.permissionCheck(event.GetSender().GetLogin(), event)
	if err != nil {
		util.Error(errors.Wrap(err, "process label event"))
		return
	}
}

func (m *merge) processIssueCommentCreatedEvent(event *github.IssueCommentEvent) {

}

func (m *merge) createAutoMergeJob(pull *github.PullRequest) {

}

func (m *merge) permissionCheck(user string, event interface{}) (bool, *github.PullRequest, error) {
	switch event.(type) {
	case *github.PullRequestEvent:
		{
			pullEvent, _ := event.(*github.PullRequestEvent)
			if pullEvent.GetSender().GetLogin() == m.opr.Config.Github.Bot {
				return false, nil, nil
			}
			pull := pullEvent.GetPullRequest()
			if pull.GetBase().GetRef() == "master" {
				return true, pull, nil
			} else {

			}
		}
	}
	return false, nil, nil
}

func (m *merge) queueComment() {

}
