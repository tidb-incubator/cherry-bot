package pullstatus

import (
	"context"
	"time"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v29/github"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	maxRetryTime = 3
)

// PullStatusControl is pull status control table structure
type PullStatusControl struct {
	ID         int       `gorm:"column:id"`
	PullID     int       `gorm:"column:pull_number"`
	Owner      string    `gorm:"column:owner"`
	Repo       string    `gorm:"column:repo"`
	Label      string    `gorm:"column:label"`
	Status     bool      `gorm:"column:status"`
	LastUpdate time.Time `gorm:"column:last_update"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

// PullStatusCheck is pull status check table structure
type PullStatusCheck struct {
	ID        int       `gorm:"column:id"`
	PullID    int       `gorm:"column:pull_number"`
	Owner     string    `gorm:"column:owner"`
	Repo      string    `gorm:"column:repo"`
	Label     string    `gorm:"column:label"`
	Duration  int       `gorm:"column:duration"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// SlackUser is slack user table structure
type SlackUser struct {
	ID     int    `gorm:"id"`
	Github string `gorm:"github"`
	Email  string `gorm:"email"`
	Slack  string `gorm:"slack"`
}

// database request
func (p *pullStatus) getPullStatusControl(pullNumber int, label string) (*PullStatusControl, error) {
	model := &PullStatusControl{}
	if err := p.opr.DB.Where("owner = ? AND repo = ? AND pull_number = ? AND label = ?",
		p.owner, p.repo, pullNumber, label).First(model).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "query pull status control failed")
	}
	return model, nil
}

func (p *pullStatus) getPullStatusControls() ([]*PullStatusControl, error) {
	var models []*PullStatusControl
	if err := p.opr.DB.Where("owner = ? AND repo = ? AND status = ?",
		p.owner, p.repo, false).Find(&models).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return models, errors.Wrap(err, "query pull status controls failed")
	}
	return models, nil
}

func (p *pullStatus) createPullStatusControl(pullNumber int, label string) error {
	model := &PullStatusControl{
		PullID:     pullNumber,
		Owner:      p.owner,
		Repo:       p.repo,
		Label:      label,
		Status:     false,
		LastUpdate: time.Now(),
	}
	return errors.Wrap(p.saveModel(model), "create pull status control")
}

func (p *pullStatus) saveModel(model interface{}) error {
	return errors.Wrap(util.RetryOnError(context.Background(), maxRetryTime, func() error {
		return p.opr.DB.Save(model).Error
	}), "save label model")
}

func (p *pullStatus) updatePull(pullNumber int, label string) error {
	model, err := p.getPullStatusControl(pullNumber, label)
	if err != nil {
		return errors.Wrap(err, "update pull")
	}
	if model.ID == 0 {
		return nil
	}
	model.LastUpdate = time.Now()
	return errors.Wrap(p.saveModel(model), "update pull")
}

func (p *pullStatus) closePull(pull *github.PullRequest) error {
	model := &PullStatusControl{}
	if err := p.opr.DB.Model(model).Where("owner = ? AND repo = ? AND pull_number = ?",
		p.owner, p.repo, pull.GetNumber()).Update("status", true).Error; err != nil {
		return errors.Wrap(err, "close pull")
	}
	return nil
}

func (p *pullStatus) openPull(pull *github.PullRequest) error {
	model := &PullStatusControl{}
	if err := p.opr.DB.Model(model).Where("owner = ? AND repo = ? AND pull_number = ?",
		p.owner, p.repo, pull.GetNumber()).Update("status", false).Error; err != nil {
		return errors.Wrap(err, "open pull")
	}
	return nil
}

func (p *pullStatus) getPullStatusCheck(pullNumber int, label string, duration int, updatedAt time.Time) (*PullStatusCheck, error) {
	model := &PullStatusCheck{}
	if err := p.opr.DB.Where("owner = ? AND repo = ? AND pull_number = ? AND label = ? AND updated_at = ? AND duration = ?",
		p.owner, p.repo, pullNumber, label, updatedAt, duration).First(model).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "query pull status check failed")
	}
	return model, nil
}

func (p *pullStatus) createPullStatusCheck(pullNumber int, label string, duration int, updatedAt time.Time) error {
	model := &PullStatusCheck{
		PullID:    pullNumber,
		Owner:     p.owner,
		Repo:      p.repo,
		Label:     label,
		Duration:  duration,
		UpdatedAt: updatedAt,
	}
	return errors.Wrap(p.saveModel(model), "create pull status check")
}

// GitHub request
func (p *pullStatus) getPullRequest(number int) (*github.PullRequest, error) {
	pull, _, err := p.opr.Github.PullRequests.Get(context.Background(), p.owner, p.repo, number)
	if err != nil {
		return nil, errors.Wrap(err, "get pull request")
	}
	return pull, nil
}

func (p *pullStatus) getReviewers(pull *github.PullRequest) []string {
	author := pull.GetUser().GetLogin()
	reviewers := []string{}

	if reviews, _, err := p.opr.Github.PullRequests.ListReviews(context.Background(),
		p.owner, p.repo, pull.GetNumber(), nil); err == nil {
		for _, review := range reviews {
			username := review.GetUser().GetLogin()
			if username != author {
				if !checkExist(username, reviewers) {
					reviewers = append(reviewers, username)
				}
			}
		}
	}

	for _, reviewer := range pull.RequestedReviewers {
		username := reviewer.GetLogin()
		if !checkExist(username, reviewers) {
			reviewers = append(reviewers, username)
		}
	}

	return reviewers
}

func (p *pullStatus) addComment(pull *github.PullRequest, comment string) error {
	issueComment := &github.IssueComment{
		Body: &comment,
	}
	_, _, err := p.opr.Github.Issues.CreateComment(context.Background(),
		p.owner, p.repo, pull.GetNumber(), issueComment)
	return errors.Wrap(err, "add github comment")
}

func (p *pullStatus) closeGithubPull(pull *github.PullRequest) error {
	state := "closed"
	updatePull := &github.PullRequest{
		State: &state,
	}
	_, _, err := p.opr.Github.PullRequests.Edit(context.Background(),
		p.owner, p.repo, pull.GetNumber(), updatePull)
	return errors.Wrap(err, "close pull")
}

func (p *pullStatus) addLabel(pull *github.PullRequest, label string) error {
	if label == "" {
		return nil
	}

	var labels []string
	for _, l := range pull.Labels {
		if l.GetName() == label {
			return nil
		}
		labels = append(labels, l.GetName())
	}
	labels = append(labels, label)

	_, _, err := p.opr.Github.Issues.AddLabelsToIssue(context.Background(),
		p.owner, p.repo, pull.GetNumber(), labels)
	return errors.Wrap(err, "add label")
}

// slack request
func (p *pullStatus) getSlackByGithub(github string) string {
	model := SlackUser{}
	if err := p.opr.DB.Where("github = ?",
		github).First(&model).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return ""
	}
	if model.Slack != "" {
		return model.Slack
	}
	if model.Email != "" {
		slack, err := p.opr.Slack.GetUserByEmail(model.Email)
		if err != nil {
			return ""
		}
		return slack
	}
	return ""
}

func (p *pullStatus) sendSlackMessage(channel, message string) error {
	if err := p.opr.Slack.SendMessage(channel, message); err != nil {
		return errors.Wrap(err, "send slack message")
	}
	return nil
}

// utils
func checkExist(item string, slice []string) bool {
	for _, sliceItem := range slice {
		if item == sliceItem {
			return true
		}
	}
	return false
}
