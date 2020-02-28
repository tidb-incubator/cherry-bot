package slack

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v29/github"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
)

const botAvatarURL = "https://avatars.slack-edge.com/2019-07-05/674719764387_89e1530aa496fdb6854b_72.jpg"

func (b *bot) helloMessage() error {
	channel, err := b.GetUserByEmail(b.config.Heartbeat)
	if err != nil {
		return errors.Wrap(err, "hello message")
	}
	if err := b.SendMessage(channel,
		"🍒 Hello, cherry picker is on!\n💉 This is health checking channel."); err != nil {
		return errors.Wrap(err, "hello message")
	}
	if !b.config.Hello {
		return nil
	}
	for _, repo := range b.repos {
		uri := fmt.Sprintf("https://github.com/%s/%s", repo.Owner, repo.Repo)
		if repo.CherryPick && repo.CherryPickChannel != "" {
			msg := fmt.Sprintf("🚀 Repo: %s\n📌 Cherry pick events will be reported in this channel.", uri)
			if err := b.SendMessage(repo.CherryPickChannel, msg); err != nil {
				return errors.Wrap(err, "hello message")
			}
		}
		if repo.LabelCheck && repo.LabelCheckChannel != "" {
			msg := fmt.Sprintf("🚀 Repo: %s\n📌 Label check events will be reported in this channel.", uri)
			if err := b.SendMessage(repo.LabelCheckChannel, msg); err != nil {
				return errors.Wrap(err, "hello message")
			}
		}
	}
	return nil
}

func (b *bot) GetUserByEmail(email string) (string, error) {
	user, err := b.api.GetUserByEmail(email)
	if err != nil {
		return "", errors.Wrap(err, "get user by email")
	}
	return user.ID, nil
}

func (b *bot) makePrAttachment(pr *github.PullRequest, stat string) *slack.Attachment {
	authorLink := fmt.Sprintf("https://github.com/%s", pr.GetUser().GetLogin())
	color := "#f1f8ff"
	if stat == "success" {
		color = "#3ca553"
	} else if stat == "failed" {
		color = "#e53b00"
	} else if stat == "merged" {
		color = "#6f42c1"
	}
	labels := []string{}
	for _, label := range pr.Labels {
		labels = append(labels, label.GetName())
	}
	fields := []slack.AttachmentField{
		slack.AttachmentField{
			Title: "Labels",
			Value: strings.Join(labels, ", "),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Comments",
			Value: strconv.Itoa(pr.GetComments()),
			Short: true,
		},
	}

	attatchment := slack.Attachment{
		Color:         color,
		AuthorName:    pr.GetUser().GetName(),
		AuthorSubname: pr.GetUser().GetLogin(),
		AuthorLink:    authorLink,
		AuthorIcon:    pr.GetUser().GetAvatarURL(),
		Title:         pr.GetTitle(),
		TitleLink:     pr.GetHTMLURL(),
		Text:          pr.GetBody(),
		Fields:        fields,
		Footer:        "Sent by GitHub bot",
		FooterIcon:    botAvatarURL,
	}
	return &attatchment
}

func (b *bot) makeIssueAttachment(issue *github.Issue) *slack.Attachment {
	authorLink := fmt.Sprintf("https://github.com/%s", issue.GetUser().GetLogin())
	color := "#f1f8ff"
	labels := []string{}
	for _, label := range issue.Labels {
		labels = append(labels, label.GetName())
	}
	fields := []slack.AttachmentField{
		slack.AttachmentField{
			Title: "Labels",
			Value: strings.Join(labels, ", "),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Comments",
			Value: strconv.Itoa(issue.GetComments()),
			Short: true,
		},
	}

	attatchment := slack.Attachment{
		Color:         color,
		AuthorName:    issue.GetUser().GetName(),
		AuthorSubname: issue.GetUser().GetLogin(),
		AuthorLink:    authorLink,
		AuthorIcon:    issue.GetUser().GetAvatarURL(),
		Title:         issue.GetTitle(),
		TitleLink:     issue.GetHTMLURL(),
		Text:          issue.GetBody(),
		Fields:        fields,
		Footer:        "Sent by GitHub bot",
		FooterIcon:    botAvatarURL,
	}
	return &attatchment
}

func (b *bot) makeIssueCommentAttachment(issue *github.Issue, issueComment *github.IssueComment) *slack.Attachment {
	authorLink := fmt.Sprintf("https://github.com/%s", issueComment.GetUser().GetLogin())

	attatchment := slack.Attachment{
		Color:         "#f1f8ff",
		AuthorName:    issueComment.GetUser().GetName(),
		AuthorSubname: issueComment.GetUser().GetLogin(),
		AuthorLink:    authorLink,
		AuthorIcon:    issueComment.GetUser().GetAvatarURL(),
		Title:         issue.GetTitle(),
		TitleLink:     issueComment.GetHTMLURL(),
		Text:          issueComment.GetBody(),
		Footer:        "Sent by GitHub bot",
		FooterIcon:    botAvatarURL,
	}
	return &attatchment
}
