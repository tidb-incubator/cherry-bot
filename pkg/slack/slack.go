package slack

import (
	"fmt"
	"time"

	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v29/github"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
)

const (
	interval = 4 * time.Hour
)

// Bot process Slack API events
type Bot interface {
	GetAPI() *slack.Client
	SendMessage(channel string, msg string) error
	NewPR(channel string, owner string, repo string, target string, pr *github.PullRequest, newPr *github.PullRequest, stat string) error
	FailPR(channel string, owner string, repo string, target string, pr *github.PullRequest, message string) error
	Report(channel string, title string, content string) error
	NoticeLabel(channel string, owner string, repo string, number int, author string) error
	helloMessage() error
	GetUserByEmail(email string) (string, error)
	SendMessageWithPr(channel string, msg string, pr *github.PullRequest, stat string) error
	SendMessageWithIssue(channel string, msg string, issue *github.Issue) error
	SendMessageWithIssueComment(channel string, msg string, issue *github.Issue, issueComment *github.IssueComment) error
}

type bot struct {
	config *config.Slack
	repos  map[string]*config.RepoConfig
	api    *slack.Client
}

// GetSlackClient connect to Slack
func GetSlackClient(slackConfig *config.Slack, repos map[string]*config.RepoConfig) (Bot, error) {
	b := bot{
		config: slackConfig,
		repos:  repos,
		api:    slack.New(slackConfig.Token),
	}

	if err := b.helloMessage(); err != nil {
		return nil, errors.Wrap(err, "get slack client")
	}
	go func() {
		for range time.Tick(interval) {
			channel, err := b.GetUserByEmail(b.config.Heartbeat)
			if err != nil {
				util.Error(errors.Wrap(err, "Slack health report"))
				continue
			}
			if err := b.SendMessage(channel, "Cherry Picker is working hard 😣"); err != nil {
				util.Error(errors.Wrap(err, "Slack health report"))
			}
		}
	}()
	return &b, nil
}

// get api for outside usage
func (b *bot) GetAPI() *slack.Client {
	return b.api
}

func (b *bot) SendMessage(channel string, msg string) error {
	if b.config.Mute {
		return nil
	}
	if _, _, err := (*b).api.PostMessage(channel,
		slack.MsgOptionText(msg, true)); err != nil {
		return errors.Wrap(err, "send slack message to "+channel)
	}
	return nil
}

func (b *bot) SendMessageWithPr(channel string, msg string, pr *github.PullRequest, stat string) error {
	if b.config.Mute {
		return nil
	}
	attachment := b.makePrAttachment(pr, stat)
	if _, _, err := (*b).api.PostMessage(channel,
		slack.MsgOptionText(msg, true), slack.MsgOptionAttachments(*attachment)); err != nil {
		return errors.Wrap(err, "send slack message with pr to "+channel)
	}
	return nil
}

func (b *bot) SendMessageWithIssue(channel string, msg string, issue *github.Issue) error {
	if b.config.Mute {
		return nil
	}
	attachment := b.makeIssueAttachment(issue)
	if _, _, err := (*b).api.PostMessage(channel,
		slack.MsgOptionText(msg, true), slack.MsgOptionAttachments(*attachment)); err != nil {
		return errors.Wrap(err, "send slack message with issue to "+channel)
	}
	return nil
}

func (b *bot) SendMessageWithIssueComment(channel string, msg string, issue *github.Issue, issueComment *github.IssueComment) error {
	if b.config.Mute {
		return nil
	}
	attachment := b.makeIssueCommentAttachment(issue, issueComment)
	if _, _, err := (*b).api.PostMessage(channel,
		slack.MsgOptionText(msg, true), slack.MsgOptionAttachments(*attachment)); err != nil {
		return errors.Wrap(err, "send slack message with issue comment to "+channel)
	}
	return nil
}

func (b *bot) NewPR(channel string, owner string, repo string, target string,
	pr *github.PullRequest, newPr *github.PullRequest, stat string) error {
	if channel == "" {
		return nil
	}
	uri := fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, newPr.GetNumber())
	origin := fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, pr.GetNumber())
	msg := fmt.Sprintf("✅ Create cherry pick pull request from `%s` to `%s`\n%s\nFrom: %s",
		pr.GetHead().GetLabel(), target, uri, origin)
	if err := (*b).SendMessageWithPr(channel, msg, newPr, stat); err != nil {
		return errors.Wrap(err, "send success message")
	}
	return nil
}

func (b *bot) FailPR(channel string, owner string, repo string, target string,
	pr *github.PullRequest, message string) error {
	uri := fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, pr.GetNumber())
	msg := fmt.Sprintf("❌ Create cherry pick pull request from `%s` to `%s`\norigin PR\n%s",
		pr.GetHead().GetLabel(), target, uri)
	if message != "" {
		msg = fmt.Sprintf("%s\n%s", msg, message)
	}
	if err := (*b).SendMessage(channel, msg); err != nil {
		return errors.Wrap(err, "send fail message")
	}
	return nil
}

func (b *bot) Report(channel string, title string, content string) error {
	msg := fmt.Sprintf("%s\n\n%s", title, content)
	if err := (*b).SendMessage(channel, msg); err != nil {
		return errors.Wrap(err, "send fail message")
	}
	return nil
}

func (b *bot) NoticeLabel(channel string, owner string, repo string, number int, author string) error {
	uri := fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, number)
	msg := fmt.Sprintf("🕵🏼 Unlabeled issue detected\n👩🏼‍💻 Author: %s\n%s", author, uri)
	if err := (*b).SendMessage(channel, msg); err != nil {
		return errors.Wrap(err, "send label notice")
	}
	return nil
}
