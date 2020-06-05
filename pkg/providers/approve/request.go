package approve

import (
	"context"

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

func (a *Approve) canApprove(login string) (bool, error) {
	if a.opr.Member.IfMember(login) {
		return true, nil
	}
	model := &ApproveRecord{}
	if err := a.opr.DB.Where("owner = ? AND repo = ? AND github = ?",
		a.owner, a.repo, login).First(model).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "query can approve failed")
	}
	return true, nil
}

func (a *Approve) sendApprove(pullNumber int) error {
	var (
		body  string = "LGTM"
		event string = "APPROVE"
	)
	review := &github.PullRequestReviewRequest{
		Body:  &body,
		Event: &event,
	}
	_, _, err := a.opr.Github.PullRequests.CreateReview(context.Background(), a.owner, a.repo, pullNumber, review)
	return errors.Wrap(err, "send approve")
}

func (a *Approve) dismissApprove(pullNumber int) error {
	reviews, _, err := a.opr.Github.PullRequests.ListReviews(context.Background(), a.owner, a.repo, pullNumber, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		return errors.Wrap(err, "dismiss approve")
	}

	var (
		reviewID       int64
		dismissMessage = "approve cancel command"
	)
	for _, review := range reviews {
		if review.GetState() == "APPROVED" && review.GetUser().GetLogin() == a.opr.Config.Github.Bot {
			reviewID = review.GetID()
		}
	}

	if reviewID == 0 {
		return a.addGithubComment(pullNumber, "bot approve review not found")
	}

	_, _, err = a.opr.Github.PullRequests.DismissReview(context.Background(), a.owner, a.repo, pullNumber, reviewID,
		&github.PullRequestReviewDismissalRequest{
			Message: &dismissMessage,
		})

	return errors.Wrap(err, "dismiss approve")
}

func (a *Approve) addGithubComment(pullNumber int, commentBody string) error {
	if commentBody == "" {
		return nil
	}
	comment := &github.IssueComment{
		Body: &commentBody,
	}
	_, _, err := a.opr.Github.Issues.CreateComment(context.Background(),
		a.owner, a.repo, pullNumber, comment)
	return errors.Wrap(err, "add github comment")
}
