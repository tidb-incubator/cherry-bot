package label

import (
	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

var ignoreActions = []string{
	"closed",
	"labeled",
	"locked",
}

func (l *label) ProcessPullRequest(pr *github.PullRequest) {
	if m, er := l.getLabelCheck(pr.GetNumber()); er == nil {
		if m.ID != 0 {
			return
		}
	}
	issue, er := l.getIssueByID(pr.GetNumber())
	if er != nil {
		util.Error(errors.Wrap(er, "cherry picker process pull request event"))
	}
	if err := l.processLabelCheck(issue); err != nil {
		util.Error(errors.Wrap(err, "label middleware process pull request"))
	}
	// if err := l.processLabelCheck(pr); err != nil {
	// 	util.Error(errors.Wrap(err, "label middleware process pull request"))
	// }
}

func (l *label) ProcessPullRequestEvent(event *github.PullRequestEvent) {
	if ifIgnoreAction(event.GetAction()) {
		return
	}
	if m, er := l.getLabelCheck(event.GetPullRequest().GetNumber()); er == nil {
		if m.ID != 0 {
			return
		}
	}

	var err error

	if *event.Action == "opened" {
		issue, er := l.getIssueByID(event.GetPullRequest().GetNumber())
		if er != nil {
			util.Error(errors.Wrap(er, "cherry picker process pull request event"))
		}
		err = l.processLabelCheck(issue)
	}

	if err != nil {
		util.Error(errors.Wrap(err, "cherry picker process pull request event"))
	}
}

func (l *label) ProcessIssuesEvent(event *github.IssuesEvent) {
	if event.GetAction() != "opened" {
		return
	}
	util.Println("process issue event", event.GetIssue().GetNumber())

	if err := l.processLabelCheck(event.GetIssue()); err != nil {
		util.Error(errors.Wrap(err, "cherry picker process issue event"))
	}
}

func ifIgnoreAction(action string) bool {
	for _, ignoreAction := range ignoreActions {
		if action == ignoreAction {
			return true
		}
	}
	return false
}
