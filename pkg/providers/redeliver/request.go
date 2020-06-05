package redeliver

import (
	"context"
	"fmt"
	"time"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	maxRetryTime = 2
)

// IssueRedeliver define issue redeliver database structure
type IssueRedeliver struct {
	ID        int       `gorm:"column:id"`
	IssueID   int       `gorm:"column:issue_number"`
	Owner     string    `gorm:"column:owner"`
	Repo      string    `gorm:"column:repo"`
	Channel   string    `gorm:"column:channel"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (r *redeliver) saveModel(model interface{}) error {
	ctx := context.Background()
	return errors.Wrap(util.RetryOnError(ctx, maxRetryTime, func() error {
		return r.opr.DB.Save(model).Error
	}), "save auto merge model")
}

func (r *redeliver) getRedeliver(issueID int, channel string) (*IssueRedeliver, error) {
	model := &IssueRedeliver{}
	if err := r.opr.DB.Where("owner = ? AND repo = ? AND issue_number = ? AND channel = ?",
		r.owner, r.repo, issueID, channel).First(model).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "query cherry pick failed")
	}
	return model, nil
}

func (r *redeliver) sendSlackMessage(issue *github.Issue, msg string, channel string) error {
	issueURL := fmt.Sprintf("https://github.com/%s/%s/issues/%d", r.owner, r.repo, issue.GetNumber())
	message := fmt.Sprintf("🕵🏼 %s\n%s", msg, issueURL)
	if err := r.opr.Slack.SendMessageWithIssue(channel, message, issue); err != nil {
		return errors.Wrap(err, "redeliver send slack message")
	}
	return nil
}

func (r *redeliver) sendNotice(issue *github.Issue, channel string, msg string) error {
	r.Lock()
	defer r.Unlock()
	issueID := issue.GetNumber()
	model, err := r.getRedeliver(issueID, channel)
	if err != nil {
		return errors.Wrap(err, "redeliver send notice")
	}

	if model.ID != 0 {
		// already sent notice
		return nil
	}

	if err = r.sendSlackMessage(issue, msg, channel); err != nil {
		return errors.Wrap(err, "redeliver send notice")
	}

	model.IssueID = issueID
	model.Owner = r.owner
	model.Repo = r.repo
	model.Channel = channel

	if err = r.saveModel(model); err != nil {
		return errors.Wrap(err, "redeliver send notice")
	}

	return nil
}

func (r *redeliver) sendCommentNotice(channel string, msg string, issue *github.Issue, comment *github.IssueComment) error {
	err := r.opr.Slack.SendMessageWithIssueComment(channel, msg, issue, comment)
	return errors.Wrap(err, "send comment notice")
}
