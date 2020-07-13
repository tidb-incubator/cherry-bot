package merge

import (
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/ngaut/log"
	operator "github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

// AutoMergeAllowName define allow name for auto merge

func (m *merge) CanMergeToMaster(pullNumber int, labels []*github.Label, userName string) error {
	err := m.opr.HasPermissionToPRWithLables(m.owner, m.repo, labels, userName, operator.MERGE_ROLES)
	if err != nil {
		err = fmt.Errorf("@%s Oops! auto merge is restricted to Committers of the SIG.%s", userName, err)
		return err
	}
	lgtmNum, err := m.opr.GetLGTMNumForPR(m.owner, m.repo, pullNumber)
	if err != nil {
		log.Error(err)
		return nil
	}
	needLGTMNum := m.opr.GetNumberOFLGTMByLable(m.repo, labels)
	if lgtmNum < needLGTMNum {
		return fmt.Errorf("@%s Oops! This PR requires at least %v LGTMs to merge. The current number of `LGTM` is %v.", userName, needLGTMNum, lgtmNum)
	}
	return nil
}
