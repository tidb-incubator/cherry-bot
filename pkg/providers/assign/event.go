package assign

import (
	"context"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/ngaut/log"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

const (
	assignCommand       = "/assign"
	assignCancelCommand = "/assign cancel"
	unAssignCommand     = "/unassign"
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
	if err := assign.opr.CommentOnGithub(assign.owner, assign.repo, event.GetIssue().GetNumber(), comment); err != nil {
		util.Error(err)
	}
}

func (a *Assign) do(event *github.IssueCommentEvent, comment string) (err error) {
	assign := strings.HasPrefix(comment, assignCommand)
	if assign {
		comment = strings.TrimPrefix(comment, assignCommand)
	} else if strings.HasPrefix(comment, unAssignCommand) {
		comment = strings.TrimPrefix(comment, unAssignCommand)
	} else if strings.HasPrefix(comment, assignCancelCommand) {
		comment = strings.TrimPrefix(comment, assignCancelCommand)
	} else {
		return nil
	}
	log.Info(assign, comment)
	assignees := []string{}
	for _, login := range strings.Split(comment, ",") {
		user := strings.TrimSpace(login)
		user = strings.TrimPrefix(user, "@")
		assignees = append(assignees, user)
	}
	if len(assignees) == 0 {
		assignees = []string{event.GetComment().GetUser().GetLogin()}
	}
	// TODO:check login's permission
	if assign {
		_, _, err = a.opr.Github.Issues.AddAssignees(context.Background(), a.owner, a.repo,
			event.GetIssue().GetNumber(), assignees)
	} else {
		_, _, err = a.opr.Github.Issues.RemoveAssignees(context.Background(), a.owner, a.repo,
			event.GetIssue().GetNumber(), assignees)
	}
	return errors.Wrap(err, "send assign")
}
