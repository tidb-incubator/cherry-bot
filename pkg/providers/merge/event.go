package merge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

const (
	autoMergeCommand                    = "/run-auto-merge"
	autoMergeAlias                      = "/merge"
	unMergeCommand                      = "/unmerge"
	noAccessComment                     = "Sorry @%s, you don't have permission to trigger auto merge event on this branch."
	needReleaseMaintainerApproveComment = "Sorry @%s, this branch cannot be merged without an approval of release maintainers"
	versionReleaseComment               = "The version releasement is in progress."
)

func (m *merge) ProcessPullRequestEvent(event *github.PullRequestEvent) {
	if event.GetSender().GetLogin() == m.opr.Config.Github.Bot {
		return
	}
	if *event.Action == "labeled" && *event.Label.Name == m.cfg.CanMergeLabel {
		pr := event.GetPullRequest()
		login := event.GetSender().GetLogin()

		if !m.havePermission(login, pr) {
			return
		}
		m.processPREvent(event)
	}
}

func (m *merge) havePermission(username string, pr *github.PullRequest) (permission bool) {
	base := pr.GetBase().GetRef()
	var msg string

	defer func() {
		if msg != "" {
			util.Error(m.opr.CommentOnGithub(m.owner, m.repo, pr.GetNumber(), msg))
		}
	}()

	if username != m.opr.Config.Github.Bot && m.cfg.MergeSIGControl {
		if base == "master" || (strings.HasPrefix(base, "release") && !m.cfg.ReleaseAccessControl) {
			if err := m.SIGAutoMergeCheck(pr.Labels, username); err != nil {
				msg = err.Error()
				return
			}
		}
		if base == "master" {
			if err := m.CanMergeToMaster(pr.GetNumber(), pr.Labels, username); err != nil {
				msg = err.Error()
				return
			}
		}
	}

	if strings.HasPrefix(base, "release") {
		if err := m.CanMergeToRelease(pr.GetNumber(), username); err != nil {
			msg = err.Error()
			return
		}
		currentReleaseVersion, err := m.currentReleaseVersion(base)
		if err != nil {
			util.Error(err)
			return
		}

		if currentReleaseVersion != nil {
			// this branch's release version is in progress
			// check out if the user has permission to merge it
			members, err := m.getReleaseMembers(base)
			if err != nil {
				util.Error(err)
				return
			}

			isReleaseMember := false
			for _, m := range members {
				if m.User == username {
					isReleaseMember = true
				}
			}
			if !isReleaseMember {
				msg = fmt.Sprintf(noAccessComment, username)
				return
			}
		}

		if m.cfg.ReleaseAccessControl {
			reviewers, err := m.opr.GetLGTMReviewers(m.owner, m.repo, *pr.Number)
			if err != nil {
				util.Error(err)
				return
			}

			if !m.opr.IsAllowed(m.owner, m.repo, reviewers...) {
				msg = fmt.Sprintf(needReleaseMaintainerApproveComment, username)
				return
			}
		}
	}

	return true
}

func (m *merge) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if event.GetAction() != "created" {
		return
	}

	body := event.Comment.GetBody()
	if body == autoMergeCommand || body == autoMergeAlias || body == unMergeCommand {
		// command only for org members
		login := event.GetSender().GetLogin()

		pr, _, err := m.opr.Github.PullRequests.Get(context.Background(),
			m.owner, m.repo, event.Issue.GetNumber())
		if err != nil {
			util.Error(errors.Wrap(err, "issue comment get PR"))
			return
		}
		if pr.GetBase().GetRef() != "master" && !m.opr.Member.IfMember(login) {
			return
		}
		if !m.havePermission(login, pr) {
			return
		}
		if body == unMergeCommand {
			util.Error(m.removeCanMerge(pr))
			return
		}
		util.Error(m.addCanMerge(pr))
		model := AutoMerge{
			PrID:      pr.GetNumber(),
			Owner:     m.owner,
			Repo:      m.repo,
			BaseRef:   pr.GetBase().GetRef(),
			Status:    mergeIncomplete,
			CreatedAt: time.Now(),
		}
		if err := m.saveModel(&model); err != nil {
			util.Error(errors.Wrap(err, "merge process PR event"))
		} else {
			util.Error(m.queueComment(pr))
		}
	}
}
