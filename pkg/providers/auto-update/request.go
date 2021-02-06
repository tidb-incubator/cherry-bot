package autoupdate

import (
	"context"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

// AutoUpdateReviewers is a table struct of requested reviewer when auto updating
type AutoUpdateReviewers struct {
	ID          int    `gorm:"column:id"`
	Owner       string `gorm:"column:owner"`
	Repo        string `gorm:"column:repo"`
	UpdateOwner string `gorm:"column:update_owner"`
	UpdateRepo  string `gorm:"column:update_repo"`
	Reviewer    string `gorm:"reviewer"`
}

func (au *autoUpdate) reviewers() (reviewers []*AutoUpdateReviewers, err error) {
	err = au.opr.DB.
		Where("owner = ? AND repo = ? AND update_owner = ? AND update_owner = ?", au.owner, au.watchedRepo, au.updateOwner, au.updateRepo).
		Find(&reviewers).Error
	return
}

func (au *autoUpdate) getReviewersRequest() (*github.ReviewersRequest, error) {
	reviewers, err := au.reviewers()
	if err != nil {
		return nil, err
	}

	request := new(github.ReviewersRequest)
	for _, reviewer := range reviewers {
		request.Reviewers = append(request.Reviewers, reviewer.Reviewer)
	}
	return request, nil
}

func (au *autoUpdate) requestReviewers(pr *github.PullRequest) error {
	request, err := au.getReviewersRequest()
	if err != nil {
		return err
	}
	_, _, err = au.opr.Github.PullRequests.RequestReviewers(context.Background(), au.updateOwner, au.updateRepo, *pr.Number, *request)
	return errors.Wrap(err, "add github requests reviews")
}
