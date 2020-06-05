package redeliver

import (
	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

func (r *redeliver) ProcessIssuesEvent(event *github.IssuesEvent) {
	if event.GetAction() == "closed" || event.GetAction() == "locked" {
		return
	}
	if event.Issue.IsPullRequest() {
		return
	}
	if err := r.checkIssueTitle(event.Issue); err != nil {
		util.Error(errors.Wrap(err, "issue deliver process issue event"))
	}
	if err := r.checkIssueBody(event.Issue); err != nil {
		util.Error(errors.Wrap(err, "issue deliver process issue event"))
	}
	if err := r.checkLabel(event.Issue); err != nil {
		util.Error(errors.Wrap(err, "issue deliver process issue event"))
	}
}

func (r *redeliver) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if event.Issue.IsPullRequest() {
		return
	}
	if event.GetAction() != "created" {
		return
	}
	if err := r.checkComment(event.GetIssue(), event.GetComment()); err != nil {
		util.Error(errors.Wrap(err, "issue deliver process issue comment event"))
	}
	// if err := r.checkFollow(event.GetIssue(), event.GetComment()); err != nil {
	// 	util.Error(errors.Wrap(err, "issue deliver process issue comment event"))
	// }
}
