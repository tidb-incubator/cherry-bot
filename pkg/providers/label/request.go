package label

import (
	"context"
	"time"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v29/github"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	maxRetryTime = 1
	oneWeek      = -7 * 24 * time.Hour
)

// LabelCheck is label check table structure
type LabelCheck struct {
	ID         int       `gorm:"column:id"`
	PrID       int       `gorm:"column:pull_number"`
	Owner      string    `gorm:"column:owner"`
	Repo       string    `gorm:"column:repo"`
	Title      string    `gorm:"column:title"`
	HasLabel   bool      `gorm:"column:has_label"`
	SendNotice bool      `gorm:"column:send_notice"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

// SlackUser is slack user table structure
type SlackUser struct {
	ID     int    `gorm:"id"`
	Github string `gorm:"github"`
	Email  string `gorm:"email"`
	Slack  string `gorm:"slack"`
}

func (l *label) getLabelCheck(number int) (*LabelCheck, error) {
	model := &LabelCheck{}
	if err := l.opr.DB.Where("pull_number = ? AND owner = ? AND repo = ?",
		number, l.owner, l.repo).First(model).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "query label failed")
	}
	return model, nil
}

func (l *label) getRestoreCheck() ([]*LabelCheck, error) {
	lastWeek := time.Now().Add(oneWeek)
	var models []*LabelCheck
	if err := l.opr.DB.Where("owner = ? AND repo = ? AND has_label = ? AND send_notice = ? AND created_at > ?",
		l.owner, l.repo, false, false, lastWeek).Find(&models).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return models, errors.Wrap(err, "query restore checks failed")
	}
	return models, nil
}

func (l *label) saveModel(model interface{}) error {
	return errors.Wrap(util.RetryOnError(context.Background(), maxRetryTime, func() error {
		return l.opr.DB.Save(model).Error
	}), "save label model")
}

func (l *label) getSlackByGithub(github string) string {
	model := SlackUser{}
	if err := l.opr.DB.Where("github = ?",
		github).First(&model).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return ""
	}
	if model.Slack != "" {
		return model.Slack
	}
	if model.Email != "" {
		slack, err := l.opr.Slack.GetUserByEmail(model.Email)
		if err != nil {
			return ""
		}
		return slack
	}
	return ""
}

func (l *label) getIssueByID(id int) (*github.Issue, error) {
	var (
		issue *github.Issue
		err   error
	)
	err = util.RetryOnError(context.Background(), 2, func() error {
		iss, _, err := l.opr.Github.Issues.Get(context.Background(), l.owner, l.repo, id)
		issue = iss
		return err
	})
	if err != nil {
		return nil, errors.Wrap(err, "get issue by id")
	}
	return issue, nil
}

func (l *label) saveLabelCheck(issue *github.Issue, label bool, success bool) error {
	old, err := l.getLabelCheck(issue.GetNumber())
	if err != nil {
		return errors.Wrap(err, "save label check")
	}
	model := LabelCheck{
		ID:         old.ID,
		PrID:       issue.GetNumber(),
		Owner:      l.owner,
		Repo:       l.repo,
		Title:      issue.GetTitle(),
		HasLabel:   label,
		SendNotice: success,
		CreatedAt:  issue.GetCreatedAt(),
	}
	if err := l.saveModel(&model); err != nil {
		return errors.Wrap(err, "save label check")
	}
	return nil
}

func (l *label) sendLabelCheckNotice(issue *github.Issue, checkSlice []string) error {
	var channels []string
	if issue == nil || issue.User == nil {
		return errors.Wrap(errors.New("nil pull request"), "send label check")
	}

	for _, e := range checkSlice {
		if e != "" {
			channel := l.getSlackByGithub(e)
			if channel != "" {
				channels = append(channels, channel)
			}
		}
	}
	slack := l.getSlackByGithub(issue.GetUser().GetLogin())
	if slack != "" {
		channels = append(channels, slack)
	}
	// if slack == "" {
	// 	for _, e := range strings.Split(l.cfg.DefaultChecker, ",") {
	// 		if e != "" {
	// 			channel := l.getSlackByGithub(e)
	// 			if channel != "" {
	// 				channels = append(channels, channel)
	// 			}
	// 		}
	// 	}
	// } else {
	// 	channels = append(channels, slack)
	// }

	for _, channel := range channels {
		if err := l.opr.Slack.NoticeLabel(channel, l.owner, l.repo,
			issue.GetNumber(), issue.GetUser().GetLogin()); err != nil {
			return errors.Wrap(err, "send label check")
		}
	}
	// if err := l.saveLabelCheck(issue, false, true); err != nil {
	// 	return errors.Wrap(err, "send label check")
	// }
	return nil
}
