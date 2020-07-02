package notify

import (
	"fmt"
	"log"

	"github.com/google/go-github/v32/github"
)

func (n *Notify) ProcessIssuesEvent(event *github.IssuesEvent) {
	if event.GetAction() != "opened" {
		return
	}
	preText, text := buildIssueMsg(event)
	if err := n.sendMsg(preText, text); err != nil {
		log.Println(err)
	}
}

func (n *Notify) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if event.Issue != nil && event.Issue.PullRequestLinks != nil {
		return
	}
	preText, text := buildIssueCommentMsg(event)
	if err := n.sendMsg(preText, text); err != nil {
		log.Println(err)
	}
}

func buildIssueMsg(event *github.IssuesEvent) (preText, text string) {
	issue := event.Issue
	preText = fmt.Sprintf("New Issue <%s|%s> by user: <%s|%s>",
		issue.GetHTMLURL(), issue.GetTitle(), issue.GetUser().GetHTMLURL(), issue.GetUser().GetLogin())
	text = issue.GetBody()
	return
}

func buildIssueCommentMsg(event *github.IssueCommentEvent) (preText, text string) {
	issue := event.Issue
	comment := event.Comment
	preText = fmt.Sprintf("New comment for <%s|%s> by user: <%s|%s>",
		comment.GetHTMLURL(), issue.GetTitle(), comment.GetUser().GetHTMLURL(), comment.GetUser().GetLogin())
	text = comment.GetBody()
	return
}
