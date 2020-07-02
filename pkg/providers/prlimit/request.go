package prlimit

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type ApproveRecord struct {
	ID        int    `gorm:"column:id"`
	Owner     string `gorm:"column:owner"`
	Repo      string `gorm:"column:repo"`
	Github    string `gorm:"column:github"`
	CreatedAt string `gorm:"column:created_at"`
}

func (p *prLimit) canApprove(login string) (bool, error) {
	model := &ApproveRecord{}
	if err := p.opr.DB.Where("owner = ? AND repo = ? AND github = ?",
		p.owner, p.repo, login).First(model).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "query can approve failed")
	}
	return true, nil
}

func (p *prLimit) commentPr(openedPr *github.PullRequest, openedPrSlice []*github.PullRequest) error {
	commentBody := fmt.Sprintf("Thanks for your PR.\nThis PR will be closed by bot because you already had %d opened PRs",
		len(openedPrSlice))
	commentBody += ", close or merge them before submitting a new one.\n"

	var prLinkSlice []string
	for _, pr := range openedPrSlice {
		prLinkSlice = append(prLinkSlice, fmt.Sprintf("https://github.com/%s/%s/pull/%d",
			p.owner, p.repo, *pr.Number))
	}
	commentBody += strings.Join(prLinkSlice, " , ")

	comment := &github.IssueComment{
		Body: &commentBody,
	}
	_, _, err := p.opr.Github.Issues.CreateComment(context.Background(),
		p.owner, p.repo, *openedPr.Number, comment)
	return errors.Wrap(err, "add github test comment")
}

func (p *prLimit) closePr(openedPr *github.PullRequest) error {
	state := "closed"
	updatePr := &github.PullRequest{
		State: &state,
	}

	_, _, err := p.opr.Github.PullRequests.Edit(context.Background(), p.owner, p.repo, *openedPr.Number, updatePr)

	return errors.Wrap(err, "close PR")
}

func (p *prLimit) labelPr(openedPr *github.PullRequest, label string) error {
	if label == "" {
		return nil
	}
	var labels []string

	hasTargetLabelLabel := false
	for _, l := range openedPr.Labels {
		labels = append(labels, *l.Name)
		if *l.Name == label {
			hasTargetLabelLabel = true
		}
	}
	if !hasTargetLabelLabel {
		labels = append(labels, label)
	}

	_, _, err := p.opr.Github.Issues.AddLabelsToIssue(context.Background(),
		p.owner, p.repo, *openedPr.Number, labels)
	return errors.Wrap(err, "label PR")
}
