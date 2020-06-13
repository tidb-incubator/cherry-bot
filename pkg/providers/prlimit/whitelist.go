package prlimit

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// AllowName define allow name list database structure
type AllowName struct {
	ID        int       `gorm:"id"`
	Owner     string    `gorm:"owner"`
	Repo      string    `gorm:"repo"`
	Username  string    `gorm:"	username"`
	CreatedAt time.Time `gorm:"created_at"`
}

func (p *prLimit) GetAllowList() ([]string, error) {
	res := []string{p.opr.Config.Github.Bot}
	var allowNames []*AllowName
	if err := p.opr.DB.Where("owner = ? and repo = ?", p.owner,
		p.repo).Order("created_at asc").Find(&allowNames).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "get AllowList")
	}
	for _, w := range allowNames {
		res = append(res, (*w).Username)
	}
	return res, nil
}

func (p *prLimit) AddAllowList(username string) error {
	model := AllowName{
		Owner:     p.owner,
		Repo:      p.repo,
		Username:  username,
		CreatedAt: time.Now(),
	}

	if err := p.opr.DB.Save(&model).Error; err != nil {
		return errors.Wrap(err, "add allow name")
	}
	return nil
}

func (p *prLimit) RemoveAllowList(username string) error {
	if err := p.opr.DB.Where("username = ?", username).Delete(AllowName{}).Error; err != nil {
		return errors.Wrap(err, "remove allow name")
	}
	return nil
}
