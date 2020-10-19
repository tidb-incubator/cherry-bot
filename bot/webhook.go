package bot

import (
	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

func (b *bot) Webhook(event interface{}) {
	switch event := event.(type) {
	case *github.PullRequestEvent:
		b.processPullRequestEvent(event)
	case *github.IssueCommentEvent:
		b.processIssueCommentEvent(event)
	case *github.IssuesEvent:
		b.processIssuesEvent(event)
	case *github.PullRequestReviewEvent:
		{
			b.processPullRequestReviewEvent(event)
			//event.GetReview().Body = "LGTM"
			//event.GetReview().GetState() = "approved" | commented|changes_requested
		}
	case *github.PullRequestReviewCommentEvent:
		{
			//log.Info("processPullRequestReviewCommentEvent", event.GetAction, event)
			b.processPullRequestReviewCommentEvent(event)
		}
	}
}

func (b *bot) processPullRequestEvent(event *github.PullRequestEvent) {
	switch *event.Action {
	case "opened", "labeled", "unlabeled":
		util.Error(errors.Wrap(b.createOrUpdatePullRequest(event.PullRequest), "bot process pull request event"))
	}

	if b.cfg.CherryPick {
		b.Middleware.cherry.ProcessPullRequestEvent(event)
	}

	if b.cfg.LabelCheck {
		b.Middleware.label.ProcessPullRequestEvent(event)
	}

	if b.cfg.PrLimit {
		b.Middleware.Prlimit.ProcessPullRequestEvent(event)
	}

	if b.cfg.Merge {
		b.Middleware.Merge.ProcessPullRequestEvent(event)
	}

	if b.cfg.StatusControl {
		b.Middleware.PullStatus.ProcessPullRequestEvent(event)
	}

	if b.cfg.AutoUpdate {
		b.Middleware.AutoUpdate.ProcessPullRequestEvent(event)
	}

	b.Middleware.Contributor.ProcessPullRequestEvent(event)
}

func (b *bot) processIssuesEvent(event *github.IssuesEvent) {
	util.Println("process issue event in bot", event.GetIssue().GetNumber())
	if b.cfg.IssueRedeliver {
		b.Middleware.IssueRedeliver.ProcessIssuesEvent(event)
	}
	if b.cfg.LabelCheck {
		b.Middleware.label.ProcessIssuesEvent(event)
	}
	if b.cfg.IssueSlackNotice {
		b.Middleware.Notify.ProcessIssuesEvent(event)
	}
	b.Middleware.CheckTemplate.ProcessIssuesEvent(event)
}

func (b *bot) processIssueCommentEvent(event *github.IssueCommentEvent) {
	if b.cfg.CherryPick {
		b.Middleware.cherry.ProcessIssueCommentEvent(event)
	}
	b.Middleware.community.ProcessIssueCommentEvent(event)

	if b.cfg.Merge {
		b.Middleware.Merge.ProcessIssueCommentEvent(event)
		b.Middleware.Approve.ProcessIssueCommentEvent(event)
	}

	if b.cfg.IssueRedeliver {
		b.Middleware.IssueRedeliver.ProcessIssueCommentEvent(event)
	}

	if b.cfg.StatusControl {
		b.Middleware.PullStatus.ProcessIssueCommentEvent(event)
	}

	if b.cfg.IssueSlackNotice {
		b.Middleware.Notify.ProcessIssueCommentEvent(event)
	}

	if b.cfg.AutoUpdate {
		b.Middleware.AutoUpdate.ProcessIssueCommentEvent(event)
	}

	b.Middleware.AddLabel.ProcessIssueCommentEvent(event)

	b.Middleware.Assign.ProcessIssueCommentEvent(event)

	b.Middleware.CommandRedeliver.ProcessIssueCommentEvent(event)

	b.Middleware.AddTemplate.ProcessIssueCommentEvent(event)
}

func (b *bot) processPullRequestReviewEvent(event *github.PullRequestReviewEvent) {
	if b.cfg.StatusControl {
		b.Middleware.PullStatus.ProcessPullRequestReviewEvent(event)
	}
	if b.cfg.Merge {
		b.Middleware.Approve.ProcessPullRequestReviewEvent(event)
	}
}

func (b *bot) processPullRequestReviewCommentEvent(event *github.PullRequestReviewCommentEvent) {
	if b.cfg.StatusControl {
		b.Middleware.PullStatus.ProcessPullRequestReviewCommentEvent(event)
	}
}
