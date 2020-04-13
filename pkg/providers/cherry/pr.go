package cherry

import (
	"context"
	"fmt"
	"time"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
)

const day = 24 * time.Hour

func (cherry *cherry) commitLabel(pr *github.PullRequest, label string) error {
	model, err := cherry.getPullRequest(*pr.Number)
	if err != nil {
		return errors.Wrap(err, "commit label")
	}

	// if pr not exist
	if model.ID == 0 {
		util.Println("commit label, found model not exist", *pr.Number, model)
		return errors.New("pull request not exist when add label")
	}

	target, version, err := cherry.getTarget(label)
	if err != nil {
		// label parse failed, skip it
		return nil
	}

	// add label to database
	ifHasLabel, err := hasLabel(label, model.Label.String())
	if err != nil {
		return errors.Wrap(err, "commit label")
	}
	if !ifHasLabel {
		if err := cherry.addLabel(model, label); err != nil {
			return errors.Wrap(err, "commit label")
		}
	}

	// unmerged issue will be cherry picked later when it merged
	if model.Merge {
		if err := cherry.cherryPick(pr, target, version, true); err != nil {
			return errors.Wrap(err, "commit label")
		}
	}
	return nil
}

func (cherry *cherry) commitMerge(pr *github.PullRequest) error {
	model, err := cherry.getPullRequest(pr.GetNumber())
	if err != nil {
		return errors.Wrap(err, "commit merge")
	}

	merge := false
	if pr.MergedAt != nil {
		merge = true
	}
	if model.Merge != merge {
		model.Merge = merge
		if err := cherry.saveModel(model); err != nil {
			return errors.Wrap(err, "commit merge")
		}
		if merge {
			// commit pr
			// labels := labelSlice{}
			// if err := json.Unmarshal([]byte(model.Label.String()), &labels); err != nil {
			// 	return errors.Wrap(err, "commit merge")
			// }
			for _, l := range model.Label {
				target, version, err := cherry.getTarget(l)
				if err != nil {
					util.Error(errors.Wrap(err, "commit merge"))
				} else if version != "" {
					if err := cherry.cherryPick(pr, target, version, true); err != nil {
						util.Error(errors.Wrap(err, "commit merge"))
					}
				}
			}
		}
	}
	return nil
}

func (cherry *cherry) cherryPick(pr *github.PullRequest, target string, version string, byLabel bool) error {
	if !cherry.ready {
		util.Printf("%s/%s#%d cherry pick not ready", cherry.owner, cherry.repo, pr.GetNumber())
		return nil
	}
	// cherry pick from a branch to itself
	if pr.GetBase().GetRef() == target {
		return nil
	}
	// hotfix for the empty target
	// TODO: find out the real problem
	if target == "" {
		return nil
	}
	// if pr.MergedAt != nil {
	// 	util.Println("try creating PR, ID:", *pr.Number, ", merged at", *pr.MergedAt)
	// }
	if byLabel && pr.MergedAt != nil && pr.MergedAt.Before(time.Now().Add(-2*day)) {
		return nil
	}

	util.Printf("%s/%s#%d cherry pick locked", cherry.owner, cherry.repo, pr.GetNumber())
	cherry.Lock()
	defer cherry.Unlock()

	title := fmt.Sprintf("%s (#%d)", pr.GetTitle(), pr.GetNumber())
	from := pr.GetNumber()

	model, err := cherry.getCherryPick(from, target)
	if err != nil {
		return errors.Wrap(err, "commit cherry pick")
	}

	if model.FromPr != 0 {
		if model.Success {
			// cherry pick pull request already created
			return nil
		}
		if byLabel && model.TryTime >= maxRetryTime {
			// fail over max try times, cancel pull request plan
			return nil
			// return errors.New(fmt.Sprintf("create cherry pick PR failed over %d times", model.TryTime))
		}
	}

	newPr, prepareMessage, err := cherry.prepareCherryPick(pr, target)
	if err != nil {
		model.Owner = cherry.owner
		model.Repo = cherry.repo
		model.FromPr = from
		model.Base = target
		model.TryTime = model.TryTime + 1
		model.Success = false
		cherry.saveModel(model)
		// util.Error(cherry.prNotice(false, pr, target, 0, from, prepareMessage))
		util.Error(cherry.prNotice(false, target, pr, nil, "fail", prepareMessage))
		util.Error(cherry.addGithubReadyComment(pr, false, target, 0))
		return errors.Wrap(err, "commit cherry pick")
	}
	resPr, tryTime, err := cherry.submitCherryPick(newPr)

	success := false
	if resPr == nil && err == nil {
		// pr already exist
		return nil
	} else if err != nil {
		// pr create failed
		// util.Error(cherry.prNotice(false, pr, target, 0, from, "submit PR failed"))
		util.Error(cherry.prNotice(false, target, pr, nil, "fail", "submit PR failed"))
		util.Error(cherry.addGithubReadyComment(pr, false, target, 0))
		// util.Error(cherry.opr.Slack.FailPR(cherry.cfg.CherryPickChannel,
		// 	cherry.owner, cherry.repo, *pr.Head.Label, target, *pr.Number))
		return errors.Wrap(err, "commit cherry pick")
	} else {
		success = true
	}

	model.PrID = *resPr.Number
	model.FromPr = from
	model.Owner = cherry.owner
	model.Repo = cherry.repo
	model.Title = title
	model.Head = *pr.Head.Label
	model.Base = target
	model.Body = *newPr.Body
	model.CreatedByBot = true
	model.Success = success
	model.TryTime = model.TryTime + tryTime

	err = cherry.saveModel(model)
	if err != nil {
		return errors.Wrap(err, "commit cherry pick")
	}

	if success {
		if cherry.cfg.RunTestCommand != "" {
			util.Error(cherry.addGithubTestComment(resPr))
		}
		util.Error(cherry.addGithubReadyComment(pr, true, target, *resPr.Number))
		util.Error(cherry.addGithubLabel(resPr, pr, version))
		util.Error(cherry.addGithubRequestReviews(resPr, cherry.getReviewers(pr)))
		util.Error(cherry.replaceGithubLabel(pr, version))
	}

	updateResPr, _, err := cherry.opr.Github.PullRequests.Get(context.Background(),
		cherry.owner, cherry.repo, resPr.GetNumber())
	if err != nil {
		updateResPr = resPr
		util.Error(err)
	}
	util.Error(cherry.prNotice(true, target, pr, updateResPr, "success", ""))

	return nil
}
