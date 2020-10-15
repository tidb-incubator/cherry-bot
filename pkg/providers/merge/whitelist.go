package merge

import (
	"time"
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
	return m.opr.GetAllowList(m.owner, m.repo)
}

func (m *merge) AddAllowList(username string) error {
	return m.opr.AddAllowList(m.owner, m.repo, username)
}

func (m *merge) RemoveAllowList(username string) error {
	return m.opr.RemoveAllowList(username)
}

func (m *merge) ifInAllowList(username string) bool {
	return m.opr.IsAllowed(m.owner, m.repo, username)
}
