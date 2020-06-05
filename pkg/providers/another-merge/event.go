package merge

import (
	"github.com/google/go-github/v31/github"
)

func (m *merge) ProcessPullRequestEvent(event *github.PullRequestEvent) {
	switch event.GetAction() {
	case "labeled":
		{
			m.processLabelEvent(event)
		}
	}
}

func (m *merge) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	switch event.GetAction() {
	case "created":
		{
			m.processIssueCommentCreatedEvent(event)
		}
	}
}
