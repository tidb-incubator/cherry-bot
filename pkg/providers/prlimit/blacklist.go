package prlimit

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// BlackName define black name list database structure
type BlackName struct {
	ID        int       `gorm:"id"`
	Owner     string    `gorm:"owner"`
	Repo      string    `gorm:"repo"`
	Username  string    `gorm:"	username"`
	CreatedAt time.Time `gorm:"created_at"`
}

func (p *prLimit) GetBlackList() ([]string, error) {
	res := []string{}
	var blackNames []*BlackName
	if err := p.opr.DB.Where("owner = ? and repo = ?", p.owner,
		p.repo).Order("created_at asc").Find(&blackNames).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "get blacklist")
	}
	for _, w := range blackNames {
		res = append(res, (*w).Username)
	}
	return res, nil
}

func (p *prLimit) AddBlackList(username string) error {
	model := BlackName{
		Owner:     p.owner,
		Repo:      p.repo,
		Username:  username,
		CreatedAt: time.Now(),
	}

	if err := p.opr.DB.Save(&model).Error; err != nil {
		return errors.Wrap(err, "add black name")
	}
	return nil
}

func (p *prLimit) RemoveBlackList(username string) error {
	if err := p.opr.DB.Where("username = ?", username).Delete(BlackName{}).Error; err != nil {
		return errors.Wrap(err, "remove black name")
	}
	return nil
}
