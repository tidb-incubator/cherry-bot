package types

import "time"

// PullRequest struct
type PullRequest struct {
	ID         int       `gorm:"column:id"`
	PullNumber int       `gorm:"column:pull_number"`
	Owner      string    `gorm:"column:owner"`
	Repo       string    `gorm:"column:repo"`
	Title      string    `gorm:"column:title"`
	Label      Labels    `gorm:"column:label"`
	Merge      bool      `gorm:"column:merge"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}
