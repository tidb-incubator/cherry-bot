package merge

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

// AutoMergeAllowName define allow name for auto merge
type AutoMergeAllowName struct {
	ID        int       `gorm:"id"`
	Owner     string    `gorm:"owner"`
	Repo      string    `gorm:"repo"`
	Username  string    `gorm:"	username"`
	CreatedAt time.Time `gorm:"created_at"`
}

func (m *merge) GetAllowList() ([]string, error) {
	res := []string{m.opr.Config.Github.Bot}
	var allowNames []*AutoMergeAllowName
	if err := m.opr.DB.Where("owner = ? and repo = ?", m.owner,
		m.repo).Order("created_at asc").Find(&allowNames).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "get allowList")
	}
	for _, w := range allowNames {
		res = append(res, w.Username)
	}
	return res, nil
}

func (m *merge) AddAllowList(username string) error {
	model := AutoMergeAllowName{
		Owner:     m.owner,
		Repo:      m.repo,
		Username:  username,
		CreatedAt: time.Now(),
	}

	if err := m.opr.DB.Save(&model).Error; err != nil {
		return errors.Wrap(err, "add allow name")
	}
	return nil
}

func (m *merge) RemoveAllowList(username string) error {
	if err := m.opr.DB.Where("username = ?", username).Delete(AutoMergeAllowName{}).Error; err != nil {
		return errors.Wrap(err, "remove allow name")
	}
	return nil
}

func (m *merge) ifInAllowList(username string) bool {
	if !m.cfg.ReleaseAccessControl {
		return true
	}

	allowList, err := m.GetAllowList()
	util.Println(username, allowList)
	if err != nil {
		util.Error(errors.Wrap(err, "if in allow list"))
	} else {
		for _, allowname := range allowList {
			if username == allowname {
				return true
			}
		}
	}
	// FIXME: should not hard code
	// the following code is outdated
	// and should be removed
	// team, _, err := m.opr.Github.Teams.GetTeamBySlug(context.Background(), "pingcap", "owners")
	// if err == nil {
	// 	isMember, _, er := m.opr.Github.Teams.IsTeamMember(context.Background(), team.GetID(), username)
	// 	if er == nil {
	// 		return isMember
	// 	}
	// }
	return false
}
