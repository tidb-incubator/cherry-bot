package pullstatus

import (
	"github.com/google/go-github/v29/github"
)

func (p *pullStatus) ProcessPullRequestEvent(event *github.PullRequestEvent) {
	switch event.GetAction() {
	case "labeled":
		{
			p.processPullStatusControl(event.GetPullRequest())
		}
	case "synchronize":
		{
			p.processSynchronize(event.GetPullRequest())
		}
	case "closed":
		{
			p.processClosed(event.GetPullRequest())
		}
	case "reopened":
		{
			p.processReopened(event.GetPullRequest())
		}
	}
}

func (p *pullStatus) ProcessPullRequestReviewEvent(event *github.PullRequestReviewEvent) {
	if event.GetSender().GetLogin() == p.opr.Config.Github.Bot {
		return
	}
	switch event.GetAction() {
	case "submitted":
		{
			p.processReviewSubmitted(event.GetPullRequest())
		}
	}
}

// handle comment event, but do nothing
// since the comment created event seems always sent
// with review submitted event at the same time
func (p *pullStatus) ProcessPullRequestReviewCommentEvent(event *github.PullRequestReviewCommentEvent) {
}

func (p *pullStatus) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if event.GetSender().GetLogin() == p.opr.Config.Github.Bot {
		return
	}
	switch event.GetAction() {
	case "created":
		{
			p.processIssueComment(event.GetSender(), event.GetComment(), event.GetIssue())
		}
	}
}
