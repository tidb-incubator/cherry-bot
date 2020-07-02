package pullstatus

import (
	"context"
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

func (p *pullStatus) noticePingAuthor(pull *github.PullRequest) error {
	comment := fmt.Sprintf("@%s, please update your pull request.", pull.GetUser().GetLogin())
	return errors.Wrap(p.addComment(pull, comment), "notice ping author")
}

func (p *pullStatus) noticeClosePull(pull *github.PullRequest) error {
	comment := fmt.Sprintf("@%s PR closed due to no update for a long time.", pull.GetUser().GetLogin())
	comment = fmt.Sprintf("%s Feel free to reopen it anytime.", comment)
	if err := p.addComment(pull, comment); err != nil {
		return errors.Wrap(err, "notice close pull")
	}
	return errors.Wrap(p.closeGithubPull(pull), "notice close pull")
}

func (p *pullStatus) noticeLabelOutdated(pull *github.PullRequest) error {
	return errors.Wrap(p.addLabel(pull, p.cfg.LabelOutdated), "notice label outdated")
}

func (p *pullStatus) noticeCommentOutdated(pull *github.PullRequest) error {
	comment := "No updates for a long time, close PR."
	return errors.Wrap(p.addComment(pull, comment), "notice ping author")
}

func (p *pullStatus) noticePingReviewer(pull *github.PullRequest) error {
	reviewers := p.getReviewers(pull)
	if len(reviewers) == 0 {
		return nil
	}
	comment := ""
	for _, reviewer := range reviewers {
		comment += fmt.Sprintf("@%s, ", reviewer)
	}
	comment += "PTAL."
	return errors.Wrap(p.addComment(pull, comment), "notice ping author")
}

func (p *pullStatus) noticeSlackDirect(pull *github.PullRequest) error {
	pullURL := fmt.Sprintf("https://github.com/%s/%s/pull/%d", p.owner, p.repo, pull.GetNumber())
	reviewers := p.getReviewers(pull)
	if len(reviewers) == 0 {
		message := fmt.Sprintf("PR no reviewers\n%s", pullURL)
		if err := p.sendSlackMessage(p.cfg.NoticeChannel, message); err != nil {
			return errors.Wrap(err, "notice slack direct")
		}
	}

	var channels []string
	for _, reviewer := range reviewers {
		channel := p.getSlackByGithub(reviewer)
		if channel != "" {
			channels = append(channels, channel)
		}
	}
	for _, channel := range channels {
		message := fmt.Sprintf("PR waiting for review\n%s", pullURL)
		if err := p.sendSlackMessage(channel, message); err != nil {
			return errors.Wrap(err, "notice slack direct")
		}
	}

	if len(reviewers) == 0 {
		infoCollectMessage := "Send GitHub ID, Slack ID and email to Tong Mu"
		message := fmt.Sprintf("PR waiting for review, Reviewer's Slack not found\n%s\n%s",
			infoCollectMessage, pullURL)
		if err := p.sendSlackMessage(p.cfg.NoticeChannel, message); err != nil {
			return errors.Wrap(err, "notice slack direct")
		}
	}
	return nil
}

func (p *pullStatus) noticeSlackChannel(pull *github.PullRequest) error {
	pullURL := fmt.Sprintf("https://github.com/%s/%s/pull/%d", p.owner, p.repo, pull.GetNumber())
	message := fmt.Sprintf("PR waiting for review\n%s", pullURL)
	if err := p.sendSlackMessage(p.cfg.NoticeChannel, message); err != nil {
		return errors.Wrap(err, "notice slack direct")
	}
	return nil
}

func (p *pullStatus) noticeRedPacket(pull *github.PullRequest) error {
	pullURL := fmt.Sprintf("https://github.com/%s/%s/pull/%d", p.owner, p.repo, pull.GetNumber())
	redPacketMessage := "Reviewers should give red packet!"
	message := fmt.Sprintf("PR waiting for review for a long time\n%s\n%s", redPacketMessage, pullURL)
	if err := p.sendSlackMessage(p.cfg.NoticeChannel, message); err != nil {
		return errors.Wrap(err, "notice red packet")
	}
	return nil
}

func (p *pullStatus) askForReviewer(pull *github.PullRequest) error {
	reviewerIds := make(map[int64]struct{})

	for _, reviewer := range pull.RequestedReviewers {
		reviewerIds[*reviewer.ID] = struct{}{}
	}
	if len(reviewerIds) >= 2 {
		return nil
	}

	opt := github.ListOptions{}
	reviews, _, err := p.opr.Github.PullRequests.ListReviews(context.Background(), p.owner, p.repo, pull.GetNumber(), &opt)
	if err != nil {
		return err
	}
	for _, review := range reviews {
		reviewerIds[*review.User.ID] = struct{}{}
	}
	delete(reviewerIds, *pull.User.ID)
	if len(reviewerIds) >= 2 {
		return nil
	}

	message := fmt.Sprintf("PR has no reviewer!\n%s", *pull.HTMLURL)
	if err := p.sendSlackMessage(p.cfg.NoticeChannel, message); err != nil {
		return errors.Wrap(err, "ask for reviewer")
	}
	return nil
}
