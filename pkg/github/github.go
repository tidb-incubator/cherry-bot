package github

import (
	"context"

	"github.com/google/go-github/v31/github"
	"github.com/pingcap-incubator/cherry-bot/config"

	"golang.org/x/oauth2"
)

// GetGithubClient return client with auth
func GetGithubClient(githubConfig *config.Github) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubConfig.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	return github.NewClient(tc)
}
