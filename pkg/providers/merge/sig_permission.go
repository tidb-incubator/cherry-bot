package merge

import (
	"github.com/google/go-github/v32/github"
	operator "github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

// AutoMergeAllowName define allow name for auto merge

func (m *merge) CanMergeToMaster(repo string, labels []*github.Label, userName string) error {
	return m.opr.HasPermissionToPRWithLables(m.owner, m.repo, labels, userName, operator.MergeRoles)
}
