package merge

import (
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/ngaut/log"
	operator "github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

func (m *merge) checkLGTM(pullNumber, needLGTM int, userName string) error {
	lgtmNum, err := m.opr.GetLGTMNumForPR(m.owner, m.repo, pullNumber)
	if err != nil {
		log.Error(err)
		return nil
	}

	if lgtmNum < needLGTM {
		return fmt.Errorf("@%s Oops! This PR requires at least %d LGTMs to merge. The current number of `LGTM` is %d", userName, needLGTM, lgtmNum)
	}

	return nil
}

func (m *merge) CanMergeToMaster(pullNumber int, labels []*github.Label, userName string) error {
	err := m.opr.HasPermissionToPRWithLables(m.owner, m.repo, labels, userName, operator.MERGE_ROLES)
	if err != nil {
		err = fmt.Errorf("@%s Oops! auto merge is restricted to Committers of the SIG.%s", userName, err)
		return err
	}
	return m.checkLGTM(pullNumber, m.opr.GetNumberOFLGTMByLable(m.repo, labels), userName)
}

func (m *merge) CanMergeToRelease(pullNumber int, userName string) error {
	return m.checkLGTM(pullNumber, m.cfg.ReleaseLGTMNeed, userName)
}
