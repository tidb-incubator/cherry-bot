package merge

import (
	"github.com/google/go-github/v32/github"
	operator "github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

// AutoMergeAllowName define allow name for auto merge

func (m *merge) CanMergeToMaster(repo string, labels []*github.Label, userName string) error {
	canMergeRoles := []string{operator.ROLE_COMMITTER, operator.ROLE_COLEADER, operator.ROLE_LEADER}
	return m.provider.Opr.HasPermissionToPRWithLables(repo, labels, userName, canMergeRoles)
}
