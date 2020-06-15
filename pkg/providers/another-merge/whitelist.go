package merge

import (
	"context"
	"time"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/jinzhu/gorm"
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
		return nil, errors.Wrap(err, "get allowlist")
	}
	for _, w := range allowNames {
		res = append(res, (*w).Username)
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

func (m *merge) ifInAllowList(username, base string) bool {
	if !m.cfg.ReleaseAccessControl {
		return true
	}
	if base == "master" {
		return true
	}
	allowlist, err := m.GetAllowList()
	util.Println(username, allowlist)
	if err != nil {
		util.Error(errors.Wrap(err, "if in allow list"))
	} else {
		for _, allowname := range allowlist {
			if username == allowname {
				return true
			}
		}
	}
	team, _, err := m.opr.Github.Teams.GetTeamBySlug(context.Background(), "pingcap", "owners")
	if err == nil {
		memberShip, _, er := m.opr.Github.Teams.GetTeamMembership(context.Background(), team.GetID(), username)
		if er == nil {
			role := memberShip.GetRole()
			return role == "member" || role == "maintainer"
		}
	}
	return false
}
