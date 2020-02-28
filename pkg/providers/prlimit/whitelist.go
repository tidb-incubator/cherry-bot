package prlimit

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// WhiteName define white name list database structure
type WhiteName struct {
	ID        int       `gorm:"id"`
	Owner     string    `gorm:"owner"`
	Repo      string    `gorm:"repo"`
	Username  string    `gorm:"	username"`
	CreatedAt time.Time `gorm:"created_at"`
}

func (p *prLimit) GetWhiteList() ([]string, error) {
	res := []string{p.opr.Config.Github.Bot}
	var whiteNames []*WhiteName
	if err := p.opr.DB.Where("owner = ? and repo = ?", p.owner,
		p.repo).Order("created_at asc").Find(&whiteNames).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "get whitelist")
	}
	for _, w := range whiteNames {
		res = append(res, (*w).Username)
	}
	return res, nil
}

func (p *prLimit) AddWhiteList(username string) error {
	model := WhiteName{
		Owner:     p.owner,
		Repo:      p.repo,
		Username:  username,
		CreatedAt: time.Now(),
	}

	if err := p.opr.DB.Save(&model).Error; err != nil {
		return errors.Wrap(err, "add white name")
	}
	return nil
}

func (p *prLimit) RemoveWhiteList(username string) error {
	if err := p.opr.DB.Where("username = ?", username).Delete(WhiteName{}).Error; err != nil {
		return errors.Wrap(err, "remove white name")
	}
	return nil
}
