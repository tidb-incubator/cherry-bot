package bugManage

import (
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

func (m *Manage) ProcessBugIssuesEvent(event *github.IssuesEvent) {
	if *event.GetSender().Login == "ti-srebot" {
		return
	}
	var err error
	switch *event.Action {
	case "unlabeled":
		err = m.solveUnLabeled(event)
	case "labeled":
		err = m.solveLabeled(event)
	}
	if err != nil {
		util.Error(errors.Wrap(err, "cherry picker process issue event"))
	}
}
