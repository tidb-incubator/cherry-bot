package label

import (
	"context"
	"fmt"
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

func (l *Label) checkLabels(labels []string) (legalLabels, illegalLabels []string, err error) {
	repoLabels, _, e := l.provider.ListLabelsOnGithub()
	legalLabels = []string{}
	illegalLabels = []string{}
	if e != nil {
		err = errors.Wrap(e, "list labels failed")
		return
	}
	repoLabel := map[string]bool{}
	for _, l := range repoLabels {
		repoLabel[strings.ToLower(l.GetName())] = true
	}

	for _, label := range labels {
		if repoLabel[strings.ToLower(label)] {
			legalLabels = append(legalLabels, label)
		} else {
			illegalLabels = append(illegalLabels, label)
		}
	}
	return legalLabels, illegalLabels, nil
}

func (l *Label) processLabel(event *github.IssueCommentEvent, raw string) error {
	if !l.provider.IfMember(event.GetSender().GetLogin()) {
		return nil
	}
	issueID := event.GetIssue().GetNumber()

	var labels []string
	for _, label := range strings.Split(raw, ",") {
		labels = append(labels, strings.TrimSpace(label))
	}
	legal, illegal, err := l.checkLabels(labels)
	if err != nil {
		return err
	}
	if len(legal) > 0 {
		_, _, err = l.provider.GithubClient().Issues.AddLabelsToIssue(context.Background(), l.owner, l.repo, issueID, legal)
	}

	if len(illegal) > 0 {
		comment := fmt.Sprintf("These  labels are not found %s.", strings.Join(illegal, ","))
		util.Println("errMsg", comment)
		err = l.provider.CommentOnGithub(issueID, comment)
	}
	if err != nil {
		return errors.Wrap(err, "add labels")
	}
	return nil
}

func (l *Label) processUnlabel(event *github.IssueCommentEvent, raw string) error {
	if !l.provider.IfMember(event.GetSender().GetLogin()) {
		return nil
	}
	issueID := event.GetIssue().GetNumber()

	labels := strings.Split(raw, ",")

	var g errgroup.Group
	for _, label := range labels {
		ll := strings.TrimSpace(label)
		g.Go(func() error {
			_, err := l.provider.GithubClient().Issues.RemoveLabelForIssue(context.Background(), l.owner, l.repo,
				issueID, ll)
			return errors.Wrap(err, "remove labels")
		})
	}

	return g.Wait()
}
