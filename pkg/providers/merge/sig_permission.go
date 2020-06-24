package merge

import (
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/provider"
	providers "github.com/pingcap-incubator/cherry-bot/pkg/providers/provider"
)

// AutoMergeAllowName define allow name for auto merge

func (m *merge) CanMergeToMaster(repo string, labels []*github.Label, userName string) error {
	canMergeRoles := []string{providers.ROLE_COMMITTER, provider.ROLE_COLEADER, provider.ROLE_LEADER}
	return m.provider.HasPermissionToPRWithLables(repo, labels, userName, canMergeRoles)
}
