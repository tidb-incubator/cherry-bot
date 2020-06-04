package merge

import (
	"context"
	"time"

	"github.com/google/go-github/v29/github"
	"github.com/jinzhu/gorm"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

// AutoMergeWhiteName define white name for auto merge
type AutoMergeWhiteName struct {
	ID        int       `gorm:"id"`
	Owner     string    `gorm:"owner"`
	Repo      string    `gorm:"repo"`
	Username  string    `gorm:"	username"`
	CreatedAt time.Time `gorm:"created_at"`
}

func (m *merge) GetWhiteList() ([]string, error) {
	res := []string{m.opr.Config.Github.Bot}
	var whiteNames []*AutoMergeWhiteName
	if err := m.opr.DB.Where("owner = ? and repo = ?", m.owner,
		m.repo).Order("created_at asc").Find(&whiteNames).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "get whitelist")
	}
	for _, w := range whiteNames {
		res = append(res, (*w).Username)
	}
	return res, nil
}

func (m *merge) AddWhiteList(username string) error {
	model := AutoMergeWhiteName{
		Owner:     m.owner,
		Repo:      m.repo,
		Username:  username,
		CreatedAt: time.Now(),
	}

	if err := m.opr.DB.Save(&model).Error; err != nil {
		return errors.Wrap(err, "add white name")
	}
	return nil
}

func (m *merge) RemoveWhiteList(username string) error {
	if err := m.opr.DB.Where("username = ?", username).Delete(AutoMergeWhiteName{}).Error; err != nil {
		return errors.Wrap(err, "remove white name")
	}
	return nil
}

func (m *merge) ifInWhiteList(username string, pr *github.PullRequest) bool {
	base := pr.GetBase().GetRef()
	if base == "master" {
		err := m.CanMergeToMaster(m.repo, pr.Labels, username)
		util.Println("can merge to master,err", err)
		if err != nil {
			return false
		} else {
			return true
		}
	}
	if !m.cfg.ReleaseAccessControl {
		return true
	}

	whitelist, err := m.GetWhiteList()
	util.Println(username, whitelist)
	if err != nil {
		util.Error(errors.Wrap(err, "if in white list"))
	} else {
		for _, whitename := range whitelist {
			if username == whitename {
				return true
			}
		}
	}
	// FIXME: should not hard code
	team, _, err := m.opr.Github.Teams.GetTeamBySlug(context.Background(), "pingcap", "owners")
	if err == nil {
		isMember, _, er := m.opr.Github.Teams.IsTeamMember(context.Background(), team.GetID(), username)
		if er == nil {
			return isMember
		}
	}
	return false
}
