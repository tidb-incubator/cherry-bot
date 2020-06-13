package prlimit

import (
	"context"
	"strings"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

const (
	maxRetryTime = 2
	perPage      = 50
)

func (p *prLimit) processOpenedPR(openedPr *github.PullRequest) error {
	author := *openedPr.User.Login
	inOrg := false
	var err error
	for _, org := range strings.Split(p.cfg.PrLimitOrgs, ",") {
		isMember, _, er := p.opr.Github.Organizations.IsMember(context.Background(), org, author)
		if er == nil && isMember {
			inOrg = true
		}
		if er != nil {
			err = errors.Wrap(er, "process opened PR")
		}
	}
	if err != nil {
		return err
	}
	if canApprove, err := p.canApprove(author); err != nil {
		return errors.Wrap(err, "process opened PR")
	} else if inOrg && canApprove {
		inOrg = false
	}
	if !inOrg {
		return errors.Wrap(p.labelPr(openedPr, p.cfg.ContributorLabel), "process opened PR")
	}
	// no limit for bot itself
	if author == p.opr.Config.Github.Bot {
		return nil
	}
	if p.cfg.PrLimitMode == "allowlist" {
		// no limit for AllowList users
		AllowList, err := p.GetAllowList()
		if err != nil {
			util.Error(errors.Wrap(err, "prlimit process opened PR"))
			return nil
		}
		for _, name := range AllowList {
			if author == name {
				return nil
			}
		}
	}
	if p.cfg.PrLimitMode == "blocklist" {
		// no limit for user not in blocklist
		blocklist, err := p.GetBlockList()
		if err != nil {
			util.Error(errors.Wrap(err, "prlimit process opened PR"))
			return nil
		}
		inBlockList := false
		for _, name := range blocklist {
			if author == name {
				inBlockList = true
			}
		}
		if !inBlockList {
			return nil
		}
	}

	// no limit for branches beside master
	if *openedPr.Base.Ref != "master" {
		return nil
	}

	// check if PR need to be closed
	return errors.Wrap(p.checkPr(openedPr), "prlimit process opened PR")
}

func (p *prLimit) checkPr(openedPr *github.PullRequest) error {
	author := *openedPr.User.Login
	var openedPrSlice []*github.PullRequest
	count := 0

	ch := make(chan []*github.PullRequest)
	go p.fetchBatch(0, &ch)
	for prList := range ch {
		for _, pr := range prList {
			if pr.GetUser().GetLogin() == author && pr.GetNumber() != openedPr.GetNumber() && pr.GetBase().GetRef() == "master" {
				ifIgnore := false
				for _, label := range pr.Labels {
					for _, ignoreLabel := range strings.Split(util.LimitIgnoreLabels, ",") {
						if label.GetName() == ignoreLabel {
							ifIgnore = true
						}
					}
				}
				if !ifIgnore {
					count++
					openedPrSlice = append(openedPrSlice, pr)
				}
			}
		}
	}

	if len(openedPrSlice) < p.cfg.MaxPrOpened {
		return nil
	}

	err := util.RetryOnError(context.Background(), maxRetryTime, func() error {
		return errors.Wrap(p.commentPr(openedPr, openedPrSlice), "close PR")
	})
	if err != nil {
		return err
	}
	err = util.RetryOnError(context.Background(), maxRetryTime, func() error {
		return errors.Wrap(p.labelPr(openedPr, p.cfg.PrLimitLabel), "close PR")
	})
	if err != nil {
		return err
	}
	err = util.RetryOnError(context.Background(), maxRetryTime, func() error {
		return errors.Wrap(p.closePr(openedPr), "close PR")
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *prLimit) fetchBatch(startID int, ch *chan []*github.PullRequest) {
	ctx := context.Background()
	opt := &github.PullRequestListOptions{
		State: "open",
		Sort:  "created",
		ListOptions: github.ListOptions{
			Page:    1 + int(startID/perPage),
			PerPage: perPage,
		},
	}

	util.RetryOnError(ctx, maxRetryTime, func() error {
		prList, _, err := p.opr.Github.PullRequests.List(ctx, p.owner, p.repo, opt)
		if err != nil {
			return err
		}
		(*ch) <- prList
		if len(prList) < perPage {
			close(*ch)
		} else {
			p.fetchBatch(startID+perPage, ch)
		}
		return nil
	})
}
