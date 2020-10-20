package prlimit

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

const (
	maxRetryTime = 2
	perPage      = 50
	prPattern    = `\(#(\d+)\)`
)

var (
	prRegexp = regexp.MustCompile(prPattern)
)

func (p *prLimit) findOriginPr(pickPr *github.PullRequest) *github.PullRequest {
	rawID := prRegexp.FindString(*pickPr.Title)
	if rawID == "" {
		return nil
	}

	originID, err := strconv.Atoi(rawID)
	if err != nil {
		util.Error(errors.Wrap(err, fmt.Sprintf("fail to parse id of origin pull request #%s", rawID)))
		return nil
	}

	origin, _, err := p.opr.Github.PullRequests.Get(context.Background(), p.owner, p.repo, originID)
	if err != nil {
		util.Error(errors.Wrap(err, fmt.Sprintf("fail to get pull request #%d", originID)))
		return nil
	}

	return origin
}

func (p *prLimit) processOpenedPR(openedPr *github.PullRequest) error {
	author := *openedPr.User.Login
	if author == p.opr.Config.Github.Bot {
		originPr := p.findOriginPr(openedPr)
		if originPr == nil {
			return nil
		}

		author = *originPr.User.Login
	}

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

	// no limit for branches beside master or release
	if *openedPr.Base.Ref != "master" && *openedPr.Base.Ref != "release" {
		return nil
	}

	// check if PR need to be closed
	return errors.Wrap(p.checkPr(openedPr, author), "prlimit process opened PR")
}

func (p *prLimit) checkPr(openedPr *github.PullRequest, originAuthor string) error {
	var openedPrSlice []*github.PullRequest
	var cherryPickSlice []*github.PullRequest
	count := 0
	ch := make(chan []*github.PullRequest)

	go p.fetchBatch(0, &ch)

	for prList := range ch {
		for _, pr := range prList {
			ifIgnore := false
			for _, label := range pr.Labels {
				for _, ignoreLabel := range strings.Split(util.LimitIgnoreLabels, ",") {
					if label.GetName() == ignoreLabel {
						ifIgnore = true
					}
				}
			}

			if ifIgnore {
				continue
			}

			if pr.GetUser().GetLogin() == originAuthor {
				count += 2
				openedPrSlice = append(openedPrSlice, pr)
			} else if pr.GetUser().GetLogin() == p.opr.Config.Github.Bot {
				if originPr := p.findOriginPr(pr); originPr != nil && originPr.GetUser().GetLogin() == originAuthor {
					count++
					cherryPickSlice = append(cherryPickSlice, pr)
				}
			}
		}
	}

	if count <= p.cfg.MaxPrOpened {
		return nil
	}

	err := util.RetryOnError(context.Background(), maxRetryTime, func() error {
		return errors.Wrap(p.commentPr(openedPr, openedPrSlice, cherryPickSlice), "close PR")
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
