package redeliver

import (
	"fmt"
	"strings"

	"github.com/pingcap-incubator/cherry-bot/config"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

type parser func(content interface{}, rule *config.Redeliver)

func (r *redeliver) checkLabel(issue *github.Issue) error {
	var err error
	r.checkRules(issue, func(issue interface{}, rule *config.Redeliver) {
		if v, ok := issue.(*github.Issue); ok {
			for _, label := range v.Labels {
				if rule.Label == label.GetName() {
					// send Slack message
					msg := fmt.Sprintf("label \"%s\" detected", rule.Label)
					if er := r.sendNotice(v, rule.Channel, msg); er != nil {
						err = errors.Wrap(er, "check label")
					}
				}
			}
		}
	})
	return err
}

func (r *redeliver) checkIssueBody(issue *github.Issue) error {
	var err error
	r.checkRules(issue, func(comment interface{}, rule *config.Redeliver) {
		if v, ok := comment.(*github.Issue); ok {
			targetKeyword := strings.ToLower(rule.Keyword)
			excludeKeyword := strings.ToLower(rule.Exclude)
			commentBody := strings.ToLower(v.GetBody())
			commentBody = strings.ReplaceAll(commentBody, excludeKeyword, "")
			if strings.Contains(commentBody, targetKeyword) {
				// send Slack message
				msg := fmt.Sprintf("%s in issue body detected", rule.Keyword)
				if er := r.sendNotice(v, rule.Channel, msg); er != nil {
					err = errors.Wrap(er, "check issue body")
				}
			}
		}
	})
	return err
}

func (r *redeliver) checkIssueTitle(issue *github.Issue) error {
	var err error
	r.checkRules(issue, func(comment interface{}, rule *config.Redeliver) {
		if v, ok := comment.(*github.Issue); ok {
			targetKeyword := strings.ToLower(rule.Keyword)
			excludeKeyword := strings.ToLower(rule.Exclude)
			commentTitle := strings.ToLower(v.GetTitle())
			commentTitle = strings.ReplaceAll(commentTitle, excludeKeyword, "")
			if strings.Contains(commentTitle, targetKeyword) {
				// send Slack message
				msg := fmt.Sprintf("%s in issue title detected", rule.Keyword)
				if er := r.sendNotice(v, rule.Channel, msg); er != nil {
					err = errors.Wrap(er, "check issue body")
				}
			}
		}
	})
	return err
}

func (r *redeliver) checkComment(issue *github.Issue, comment *github.IssueComment) error {
	var err error
	r.checkRules(comment, func(comment interface{}, rule *config.Redeliver) {
		if v, ok := comment.(*github.IssueComment); ok {
			targetKeyword := strings.ToLower(rule.Keyword)
			excludeKeyword := strings.ToLower(rule.Exclude)
			commentBody := strings.ToLower(v.GetBody())
			commentBody = strings.ReplaceAll(commentBody, excludeKeyword, "")
			if strings.Contains(commentBody, targetKeyword) {
				// send Slack message
				msg := fmt.Sprintf("%s in comment detected", rule.Keyword)
				if er := r.sendNotice(issue, rule.Channel, msg); er != nil {
					err = errors.Wrap(er, "check comment")
				}
			}
		}
	})
	return err
}

func (r *redeliver) checkFollow(issue *github.Issue, comment *github.IssueComment) error {
	var err error
	sentChannels := []string{}
	r.checkRules(issue, func(issue interface{}, rule *config.Redeliver) {
		if v, ok := issue.(*github.Issue); ok {
			redeliverModel, er := r.getRedeliver(v.GetNumber(), rule.Channel)
			if err != nil {
				err = errors.Wrap(er, "check follow")
			} else {
				if redeliverModel.ID != 0 {
					for _, sentChannel := range sentChannels {
						if sentChannel == rule.Channel {
							return
						}
					}
					msg := fmt.Sprintf("Comment in followed issue detected")
					if er := r.sendCommentNotice(rule.Channel, msg, v, comment); er != nil {
						err = errors.Wrap(er, "check comment")
					} else {
						sentChannels = append(sentChannels, rule.Channel)
					}
				}
			}
		}
	})

	return err
}

func (r *redeliver) checkRules(content interface{}, fn parser) {
	for _, rule := range r.cfg.Redeliver {
		fn(content, rule)
	}
}
