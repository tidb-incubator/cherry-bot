package merge

import (
	"context"
	"fmt"
	"time"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
)

const (
	autoMergeCommand = "/run-auto-merge"
	autoMergeAlias   = "/merge"
	noAccessComment  = "Sorry @%s, you don't have permission to trigger auto merge event on this branch."
)

func (m *merge) ProcessPullRequestEvent(event *github.PullRequestEvent) {
	if event.GetSender().GetLogin() == m.opr.Config.Github.Bot {
		return
	}
	if *event.Action == "labeled" && *event.Label.Name == m.cfg.CanMergeLabel {
		pr := event.GetPullRequest()
		login := event.GetSender().GetLogin()
		if !m.ifInWhiteList(login, pr.GetBase().GetRef()) {
			util.Error(m.addGithubComment(pr, fmt.Sprintf(noAccessComment, login)))
			return
		}
		m.processPREvent(event)
	}
}

func (m *merge) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if event.GetAction() != "created" {
		return
	}

	if event.Comment.GetBody() == autoMergeCommand || event.Comment.GetBody() == autoMergeAlias {
		// command only for org members
		login := event.GetSender().GetLogin()
		if !m.opr.Member.IfMember(login) {
			return
		}

		pr, _, err := m.opr.Github.PullRequests.Get(context.Background(),
			m.owner, m.repo, event.Issue.GetNumber())
		if err != nil {
			util.Error(errors.Wrap(err, "issue comment get PR"))
			return
		}

		if !m.ifInWhiteList(login, pr.GetBase().GetRef()) {
			util.Error(m.addGithubComment(pr, fmt.Sprintf(noAccessComment, login)))
			return
		}
		util.Error(m.addCanMerge(pr))
		model := AutoMerge{
			PrID:      pr.GetNumber(),
			Owner:     m.owner,
			Repo:      m.repo,
			Status:    false,
			CreatedAt: time.Now(),
		}
		if err := m.saveModel(&model); err != nil {
			util.Error(errors.Wrap(err, "merge process PR event"))
		} else {
			util.Error(m.queueComment(pr))
		}
	}
}
