package providers

import (
	"context"
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"github.com/pkg/errors"
)

type Provider struct {
	Opr *operator.Operator
	*config.RepoConfig
}

func Init(repo *config.RepoConfig, opr *operator.Operator) *Provider {
	p := Provider{
		opr,
		repo,
	}
	return &p
}

//CreateComment creates a new comment on the specified issue.
//number: issue or pull request's ID
func (p *Provider) CommentOnGithub(number int, commentBody string) error {
	if commentBody == "" {
		return nil
	}
	comment := &github.IssueComment{
		Body: &commentBody,
	}
	_, _, err := p.Opr.Github.Issues.CreateComment(context.Background(),
		p.Owner, p.Repo, number, comment)
	return errors.Wrap(err, "add github comment")
}

func (p *Provider) ListLabelsOnGithub() ([]*github.Label, *github.Response, error) {
	return p.Opr.Github.Issues.ListLabels(context.Background(), p.Owner, p.Repo, nil)
}

func (p *Provider) IfMember(login string) bool {
	return p.Opr.Member.IfMember(login)
}

func (p *Provider) GithubClient() *github.Client {
	return p.Opr.Github
}
