package approve

import (
	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
)

const (
	approveCommand       = "/approve"
	approveCancelCommand = "/approve cancel"
)

func (a *Approve) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if event.GetAction() != "created" {
		return
	}
	switch event.Comment.GetBody() {
	case approveCommand:
		a.createApprove(event)
	case approveCancelCommand:
		a.cancelApprove(event)
	}
}

func (a *Approve) createApprove(event *github.IssueCommentEvent) {
	if event.GetIssue().GetPullRequestLinks() == nil {
		return
	}
	canApprove, err := a.canApprove(event.GetSender().GetLogin())
	if err != nil {
		util.Error(err)
		return
	}
	if !canApprove {
		return
	}

	comment := ""
	if event.GetSender().GetLogin() == event.GetIssue().GetUser().GetLogin() {
		comment = "Can not self-approve."
	} else if err := a.sendApprove(event.GetIssue().GetNumber()); err != nil {
		util.Error(err)
		comment = "Approve failed."
	}
	if err := a.addGithubComment(event.GetIssue().GetNumber(), comment); err != nil {
		util.Error(err)
	}
}

func (a *Approve) cancelApprove(event *github.IssueCommentEvent) {
	if event.GetIssue().GetPullRequestLinks() == nil {
		return
	}
	canApprove, err := a.canApprove(event.GetSender().GetLogin())
	if err != nil {
		util.Error(err)
		return
	}
	if !canApprove {
		return
	}

	comment := ""
	if err := a.dismissApprove(event.GetIssue().GetNumber()); err != nil {
		util.Error(err)
		comment = "Cancel approve failed."
	}
	if err := a.addGithubComment(event.GetIssue().GetNumber(), comment); err != nil {
		util.Error(err)
	}
}
