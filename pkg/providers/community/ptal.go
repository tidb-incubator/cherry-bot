package community

import (
	"context"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/ngaut/log"
)

const (
	CC_CMD    = "/cc"
	CC_CANCEL = "/uncc"
)

func (c *CommunityCmd) ptal(pullNumber int, comment string) error {
	cc := strings.HasPrefix(comment, CC_CMD)
	if cc {
		comment = strings.TrimPrefix(comment, CC_CMD)
	} else if strings.HasPrefix(comment, CC_CANCEL) {
		comment = strings.TrimPrefix(comment, CC_CANCEL)
	} else {
		return nil
	}
	log.Info(cc, comment)
	reviewers := []string{}
	for _, login := range strings.Split(comment, ",") {
		user := strings.TrimSpace(login)
		user = strings.TrimPrefix(user, "@")
		reviewers = append(reviewers, user)
	}
	reviewers_request := github.ReviewersRequest{
		Reviewers: reviewers,
	}
	var res *github.Response
	var err error
	if cc {
		_, res, err = c.opr.Github.PullRequests.RequestReviewers(context.Background(), c.owner, c.repo, pullNumber, reviewers_request)
	} else {
		res, err = c.opr.Github.PullRequests.RemoveReviewers(context.Background(), c.owner, c.repo, pullNumber, reviewers_request)
	}
	if err != nil {
		log.Error(comment, "failed with", err, res, reviewers)
	}
	return nil
}
