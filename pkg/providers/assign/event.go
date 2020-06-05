package assign

import (
	"context"
	"github.com/google/go-github/v31/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
	"strings"
)

const (
	assignCommand       = "/assign"
	assignCancelCommand = "/assign cancel"
)

func (assign *Assign) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if event.GetAction() != "created" {
		return
	}
	opType := strings.TrimSpace(event.GetComment().GetBody())
	comment := ""
	if err := assign.do(event, opType); err != nil {
		util.Error(err)
		comment = "Assign failed."
	}
	if err := assign.addGithubComment(event.GetIssue().GetNumber(), comment); err != nil {
		util.Error(err)
	}
}

func (assign *Assign) do(event *github.IssueCommentEvent, opType string) (err error) {
	assignees := []string{event.GetComment().GetUser().GetLogin()}
	switch opType {
	case assignCommand:
		_, _, err = assign.opr.Github.Issues.AddAssignees(context.Background(), assign.owner, assign.repo,
			event.GetIssue().GetNumber(), assignees)
	case assignCancelCommand:
		_, _, err = assign.opr.Github.Issues.RemoveAssignees(context.Background(), assign.owner, assign.repo,
			event.GetIssue().GetNumber(), assignees)
	}
	return errors.Wrap(err, "send assign")
}

func (assign *Assign) addGithubComment(pullNumber int, commentBody string) error {
	if commentBody == "" {
		return nil
	}
	comment := &github.IssueComment{
		Body: &commentBody,
	}
	_, _, err := assign.opr.Github.Issues.CreateComment(context.Background(),
		assign.owner, assign.repo, pullNumber, comment)
	return errors.Wrap(err, "add github comment")
}
