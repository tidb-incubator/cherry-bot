package bot

import (
	"github.com/pingcap-incubator/cherry-bot/pkg/types"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

func (b *bot) createOrUpdatePullRequest(pull *github.PullRequest) error {
	var (
		model *types.PullRequest
		err   error
	)

	model, err = b.opr.DB.GetPullByNumber(&b.Repo, pull.GetNumber())
	if err != nil {
		return errors.Wrap(err, "create pull request")
	}

	// save new pull request
	if model.ID == 0 {
		model = b.opr.DB.CreatePullRequestModel(&b.Repo, pull)
		return errors.Wrap(b.opr.DB.SavePull(model), "create pull request")
	}

	// pull request already exist
	model = b.opr.DB.PatchPullRequestModel(model, &b.Repo, pull)
	return errors.Wrap(b.opr.DB.SavePull(model), "create pull request")
}
