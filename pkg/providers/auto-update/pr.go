package autoupdate

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v31/github"
	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
)

const (
	maxRetryTime = 1
	workDir      = "/tmp"
	mergeComment = "/merge"
	testComment  = "/run-all-tests"
)

// SlackUser is slack user table structure
type SlackUser struct {
	ID     int    `gorm:"column:id"`
	Github string `gorm:"column:github"`
	Email  string `gorm:"column:email"`
	Slack  string `gorm:"column:slack"`
}

func (au *autoUpdate) CommitUpdate(pr *github.PullRequest) error {
	if !*pr.Merged {
		return nil
	}

	target, ok := au.targetMap[*pr.Base.Ref]
	if !ok {
		return nil
	}

	for _, t := range strings.Split(target, "|") {
		if err := au.Update(pr, t); err != nil {
			return errors.Wrap(err, "commit label")
		}
	}
	return nil
}

func (au *autoUpdate) Update(pr *github.PullRequest, target string) error {
	au.Lock()
	defer au.Unlock()

	newPr, prepareMessage, err := au.prepareUpdate(pr, target)
	if err != nil {
		util.Error(au.prNotice(false, target, pr, nil, "fail", prepareMessage))
		return errors.Wrap(err, "commit update")
	}
	resPr, _, err := au.submitUpdate(newPr)

	success := false
	if newPr == nil && err == nil {
		// pr already exist
		return nil
	} else if err != nil {
		// pr create failed
		util.Error(au.prNotice(false, target, pr, nil, "fail", "submit PR failed"))
		return errors.Wrap(err, "commit update")
	} else {
		success = true
	}

	if success {
		util.Error(au.addMergeComment(resPr))
	}

	updateResPr, _, err := au.opr.Github.PullRequests.Get(context.Background(),
		au.owner, au.updateRepo, resPr.GetNumber())
	if err != nil {
		updateResPr = resPr
		util.Error(err)
	}
	util.Error(au.prNotice(true, target, pr, updateResPr, "success", ""))

	return nil
}

func do(dir string, c string, args ...string) (string, error) {
	cmd := exec.Command(c, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (au *autoUpdate) prepareUpdate(pr *github.PullRequest, target string) (*github.NewPullRequest, string, error) {
	var newPr github.NewPullRequest
	var message string

	ctx := context.Background()
	err := util.RetryOnError(ctx, maxRetryTime, func() error {
		updateRepo := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", au.opr.Config.Github.Bot,
			au.opr.Config.Github.Token, au.updateOwner, au.updateRepo)
		botRepo := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", au.opr.Config.Github.Bot,
			au.opr.Config.Github.Token, au.opr.Config.Github.Bot, au.updateRepo)
		folder := fmt.Sprintf("%s-%s-%s", au.owner, au.updateRepo, (*pr.MergeCommitSHA)[0:12])
		dir := fmt.Sprintf("%s/%s", workDir, folder)
		newBranch := fmt.Sprintf("%s-%s", target, (*pr.MergeCommitSHA)[0:12])
		commit := fmt.Sprintf("%s: %s", au.watchedRepo, *pr.Title)
		head := fmt.Sprintf("%s:%s", au.opr.Config.Github.Bot, newBranch)
		body := fmt.Sprintf("update %s to include %s/%s#%d for %s", au.watchedRepo, au.owner, au.watchedRepo, pr.GetNumber(), target)
		maintainerCanModify := true
		draft := false

		defer do(workDir, "rm", "-rf", folder)
		if _, err := do(workDir, "git", "clone", updateRepo, folder); err != nil {
			message = "git clone failed"
			return errors.Wrap(err, "clone failed")
		}
		if _, err := do(dir, "git", "checkout", target); err != nil {
			message = fmt.Sprintf("branch %s not exist", target)
			return errors.Wrap(err, "checkout failed")
		}
		if _, err := do(dir, "git", "checkout", "-b", newBranch); err != nil {
			message = "create new branch failed"
			return errors.Wrap(err, "checkout to new branch failed")
		}
		updateScriptCmd := strings.Split(au.updateScript, " ")
		var updateScriptArgs []string
		if len(updateScriptCmd) == 0 {
			message = "invalid update script"
			return errors.New(message)
		} else if len(updateScriptCmd) >= 2 {
			updateScriptArgs = updateScriptCmd[1:]
		}
		if updateMessage, err := do(dir, updateScriptCmd[0], updateScriptArgs...); err != nil {
			message = fmt.Sprintf("make update script failed %s", updateMessage)
			return errors.Wrap(err, "make update script failed")
		}
		if _, err := do(dir, "git", "add", "."); err != nil {
			message = "git add failed"
			return errors.Wrap(err, "git add failed")
		}
		if _, err := do(dir, "git", "commit", "-s", "-m", fmt.Sprintf("\"update %s\"", au.watchedRepo)); err != nil {
			message = "git commit failed"
			return errors.Wrap(err, "git commit failed")
		}
		if _, err := do(dir, "git", "push", botRepo, newBranch); err != nil {
			message = "git push failed"
			return errors.Wrap(err, "git push failed")
		}

		updatePr := github.NewPullRequest{
			Title:               &commit,
			Head:                &head,
			Base:                &target,
			Body:                &body,
			MaintainerCanModify: &maintainerCanModify,
			Draft:               &draft,
		}
		newPr = updatePr
		return nil
	})
	if err != nil {
		return nil, message, errors.Wrap(err, "prepare pull request")
	}
	return &newPr, "", nil
}

func (au *autoUpdate) submitUpdate(newPr *github.NewPullRequest) (*github.PullRequest, int, error) {
	var (
		resPr   *github.PullRequest
		tryTime int
	)

	ctx := context.Background()
	err := util.RetryOnError(ctx, maxRetryTime, func() error {
		p, _, err := au.opr.Github.PullRequests.Create(context.Background(),
			au.owner, au.updateRepo, newPr)
		if err != nil {
			util.Error(err)
			if er, ok := err.(*github.ErrorResponse); ok {
				if er.Message == "Validation Failed" {
					// pull request already exist
					newPr = nil
					tryTime = 1
					return nil
				}
				return errors.Wrap(err, "create github PR failed")
			}
			return nil
		}
		resPr = p
		return nil
	})

	if err != nil {
		return nil, tryTime, errors.Wrap(err, "create github PR failed")
	}
	return resPr, tryTime, nil
}

// use /merge command instead of adding label
func (au *autoUpdate) addCanMergeLabel(res *github.PullRequest, from *github.PullRequest) error {
	_, _, err := au.opr.Github.Issues.AddLabelsToIssue(context.Background(),
		au.owner, au.updateRepo, *res.Number, []string{""})
	if err != nil {
		return errors.Wrap(err, "add github label")
	}
	return nil
}

func (au *autoUpdate) prNotice(success bool, target string,
	pr *github.PullRequest, newPr *github.PullRequest, stat string, message string) error {
	if pr == nil || pr.User == nil {
		return errors.Wrap(errors.New("nil pull request"), "send pr notice")
	}

	var channels []string

	slack := au.getSlackByGithub(*pr.User.Login)
	if slack == "" {
		for _, e := range strings.Split(au.cfg.DefaultChecker, ",") {
			if e != "" {
				channel := au.getSlackByGithub(e)
				if channel != "" {
					channels = append(channels, channel)
				}
			}
		}
	} else {
		channels = append(channels, slack)
	}

	if success {
		uri := fmt.Sprintf("https://github.com/%s/%s/pull/%d", au.owner, au.updateRepo, newPr.GetNumber())
		origin := fmt.Sprintf("https://github.com/%s/%s/pull/%d", au.owner, au.watchedRepo, pr.GetNumber())
		msg := fmt.Sprintf("✅ Create an update %s pull request for `%s` to `%s`", au.watchedRepo, uri, origin)

		if err := au.opr.Slack.SendMessageWithPr(au.cfg.AutoUpdateChannel,
			msg, newPr, "success"); err != nil {
			return errors.Wrap(err, "send pr notice")
		}
	} else {
		origin := fmt.Sprintf("https://github.com/%s/%s/pull/%d", au.owner, au.watchedRepo, pr.GetNumber())
		msg := fmt.Sprintf("❌ Create an update %s pull request for `%s`", au.watchedRepo, origin)
		if message != "" {
			msg = fmt.Sprintf("%s\n%s", msg, message)
		}
		if err := au.opr.Slack.SendMessage(au.cfg.AutoUpdateChannel, msg); err != nil {
			return errors.Wrap(err, "send success message")
		}

		for _, channel := range channels {
			err := au.opr.Slack.SendMessage(channel, msg)
			if err != nil {
				return errors.Wrap(err, "send pr notice")
			}
		}
	}
	return nil
}

func (au *autoUpdate) addMergeComment(pr *github.PullRequest) error {
	commentBody := ""
	if au.updateAutoMerge {
		commentBody = mergeComment
	} else {
		commentBody = testComment
	}
	comment := &github.IssueComment{
		Body: &commentBody,
	}
	_, _, err := au.opr.Github.Issues.CreateComment(context.Background(),
		au.owner, au.updateRepo, *pr.Number, comment)
	return errors.Wrap(err, "add github test comment")
}

// slack request
func (au *autoUpdate) getSlackByGithub(github string) string {
	model := SlackUser{}
	if err := au.opr.DB.Where("github = ?",
		github).First(&model).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return ""
	}
	if model.Slack != "" {
		return model.Slack
	}
	if model.Email != "" {
		slack, err := au.opr.Slack.GetUserByEmail(model.Email)
		if err != nil {
			return ""
		}
		return slack
	}
	return ""
}
