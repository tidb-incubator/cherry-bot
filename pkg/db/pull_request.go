package db

import (
	"github.com/google/go-github/v32/github"
	"github.com/jinzhu/gorm"
	"github.com/pingcap-incubator/cherry-bot/pkg/types"
	"github.com/pkg/errors"
)

// CreatePullRequestModel create pull request model from github pull request
func (d *DB) CreatePullRequestModel(repo *types.Repo, pull *github.PullRequest) *types.PullRequest {
	return &types.PullRequest{
		PullNumber: pull.GetNumber(),
		Owner:      repo.GetOwner(),
		Repo:       repo.GetRepo(),
		Title:      pull.GetTitle(),
		Label:      types.ParseGithubLabels(pull.Labels),
		Merge:      pull.GetMerged(),
		CreatedAt:  pull.GetCreatedAt(),
	}
}

// PatchPullRequestModel update pull request model
func (d *DB) PatchPullRequestModel(model *types.PullRequest, repo *types.Repo, pull *github.PullRequest) *types.PullRequest {
	model.PullNumber = pull.GetNumber()
	model.Owner = repo.GetOwner()
	model.Repo = repo.GetRepo()
	model.Title = pull.GetTitle()
	model.Label = types.ParseGithubLabels(pull.Labels)
	model.Merge = pull.GetMerged()
	model.CreatedAt = pull.GetCreatedAt()
	return model
}

// SavePull creates or updates pull request
func (d *DB) SavePull(pull *types.PullRequest) error {
	return d.DB.Save(pull).Error
}

// GetPullByNumber finds pull by pull number
func (d *DB) GetPullByNumber(repo *types.Repo, pullNumber int) (*types.PullRequest, error) {
	var (
		model = types.PullRequest{}
		o     = repo.GetOwner()
		r     = repo.GetRepo()
	)

	if err := d.Where("owner = ? AND repo = ? AND pull_number = ?",
		o, r, pullNumber).First(&model).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "find pull failed")
	}
	return &model, nil
}
