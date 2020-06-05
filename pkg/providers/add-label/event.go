package label

import (
	"context"
	"regexp"
	"strings"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var (
	labelPattern   = regexp.MustCompile(`\/label ([a-zA-Z0-9\/_\- ,]*)`)
	unlabelPattern = regexp.MustCompile(`\/unlabel ([a-zA-Z0-9\/_\- ,]*)`)
)

func (l *Label) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if event.GetAction() != "created" {
		return
	}
	if err := l.processComment(event, event.Comment.GetBody()); err != nil {
		util.Error(err)
	}
}

func (l *Label) processComment(event *github.IssueCommentEvent, comment string) error {
	labelMatches := labelPattern.FindStringSubmatch(comment)
	if len(labelMatches) == 2 {
		return errors.Wrap(l.processLabel(event, labelMatches[1]), "label issue")
	}
	unlabelMatches := unlabelPattern.FindStringSubmatch(comment)
	if len(unlabelMatches) == 2 {
		return errors.Wrap(l.processUnlabel(event, unlabelMatches[1]), "unlabel issue")
	}
	return nil
}

func (l *Label) processLabel(event *github.IssueCommentEvent, raw string) error {
	if !l.opr.Member.IfMember(event.GetSender().GetLogin()) {
		return nil
	}

	var labels []string
	for _, label := range strings.Split(raw, ",") {
		labels = append(labels, strings.TrimSpace(label))
	}
	_, _, err := l.opr.Github.Issues.AddLabelsToIssue(context.Background(), l.owner, l.repo,
		event.GetIssue().GetNumber(), labels)

	return errors.Wrap(err, "add labels")
}

func (l *Label) processUnlabel(event *github.IssueCommentEvent, raw string) error {
	if !l.opr.Member.IfMember(event.GetSender().GetLogin()) {
		return nil
	}

	labels := strings.Split(raw, ",")

	var g errgroup.Group
	for _, label := range labels {
		ll := strings.TrimSpace(label)
		g.Go(func() error {
			_, err := l.opr.Github.Issues.RemoveLabelForIssue(context.Background(), l.owner, l.repo,
				event.GetIssue().GetNumber(), ll)
			return errors.Wrap(err, "remove labels")
		})
	}

	return g.Wait()
}
