package prlimit

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// BlockName define block name list database structure
type BlockName struct {
	ID        int       `gorm:"id"`
	Owner     string    `gorm:"owner"`
	Repo      string    `gorm:"repo"`
	Username  string    `gorm:"	username"`
	CreatedAt time.Time `gorm:"created_at"`
}

func (p *prLimit) GetBlockList() ([]string, error) {
	res := []string{}
	var blockNames []*BlockName
	if err := p.opr.DB.Where("owner = ? and repo = ?", p.owner,
		p.repo).Order("created_at asc").Find(&blockNames).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "get blocklist")
	}
	for _, w := range blockNames {
		res = append(res, (*w).Username)
	}
	return res, nil
}

func (p *prLimit) AddBlockList(username string) error {
	model := BlockName{
		Owner:     p.owner,
		Repo:      p.repo,
		Username:  username,
		CreatedAt: time.Now(),
	}

	if err := p.opr.DB.Save(&model).Error; err != nil {
		return errors.Wrap(err, "add block name")
	}
	return nil
}

func (p *prLimit) RemoveBlockList(username string) error {
	if err := p.opr.DB.Where("username = ?", username).Delete(BlockName{}).Error; err != nil {
		return errors.Wrap(err, "remove block name")
	}
	return nil
}
