package operator

import (
	"context"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

//CreateComment creates a new comment on the specified issue.
//number: issue or pull request's ID
func (o *Operator) CommentOnGithub(owner string, repo string, number int, commentBody string) error {
	if commentBody == "" {
		return nil
	}
	comment := &github.IssueComment{
		Body: &commentBody,
	}
	_, _, err := o.Github.Issues.CreateComment(context.Background(),
		owner, repo, number, comment)
	return errors.Wrap(err, "add github comment")
}

func (o *Operator) ListLabelsOnGithub(owner, repo string) ([]*github.Label, error) {
	var (
		page    = 0
		perPage = 100
		all     []*github.Label
		batch   []*github.Label
		err     error
	)
	for len(all) == page*perPage {
		page++
		batch, _, err = o.Github.Issues.ListLabels(context.Background(), owner, repo, &github.ListOptions{
			Page:    page,
			PerPage: perPage,
		})
		if err != nil {
			return nil, errors.Wrap(err, "list all github labels")
		}
		all = append(all, batch...)
	}
	return all, nil
}
