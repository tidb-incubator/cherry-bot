package merge

import (
	"errors"
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/ngaut/log"
	operator "github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

// AutoMergeAllowName define allow name for auto merge

func (m *merge) CanMergeToMaster(pullNumber int, labels []*github.Label, userName string) error {
	err := m.opr.HasPermissionToPRWithLables(m.owner, m.repo, labels, userName, operator.MERGE_ROLES)
	if err != nil {
		return err
	}
	lgtmNum, err := m.opr.GetLGTMNumForPR(m.owner, m.repo, pullNumber)
	if err != nil {
		log.Error(err)
		return nil
	}
	if lgtmNum < 2 {
		return errors.New(fmt.Sprintf("The number of `LGTM` for this PR is %v while it need 2", lgtmNum))
	}
	return nil
}
