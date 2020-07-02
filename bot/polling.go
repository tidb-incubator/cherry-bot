package bot

import (
	"context"
	"time"

	"github.com/pingcap-incubator/cherry-bot/pkg/types"
	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// fetch batch size
const (
	perPage      = 100
	maxRetryTime = 1
)

// StartPolling init the polling job of a repository
func (b *bot) StartPolling() {
	// remove polling here
	// TODO: remove the ready state and re-design polling function
	b.ready()
	//if !b.cfg.CherryPick && !b.cfg.LabelCheck {
	//	return
	//}
	//b.fetchDuration(30 * 24 * time.Hour)
	//b.ready()
	//util.Println(b.owner, "/", b.repo, "ready")
	//// b.fetchDuration(7 * 24 * time.Hour)
	//
	//duration := time.Duration(b.interval) * time.Millisecond
	//fullupdateInterval := time.Duration(b.fullupdateInterval) * time.Millisecond
	//nextFetch := time.Now().Add(duration)
	//nextFullupdate := time.Now().Add(fullupdateInterval)
	//
	//tick := time.Tick(time.Second)
	//
	//go func() {
	//	for {
	//		select {
	//		case t := <-tick:
	//			if t.After(nextFetch) {
	//				nextFetch = nextFetch.Add(duration)
	//				b.fetchLatest()
	//			}
	//			if t.After(nextFullupdate) {
	//				nextFullupdate = nextFullupdate.Add(fullupdateInterval)
	//				b.fetchDuration(7 * 24 * time.Hour)
	//			}
	//		}
	//	}
	//}()
}

func (b *bot) fetchLatest() {
	var count int
	if err := b.opr.DB.Model(&types.PullRequest{}).Where("owner = ? AND repo = ?",
		b.owner, b.repo).Count(&count).Error; err != nil {
		util.Error(errors.Wrap(err, "fetch latest"))
	}
	b.fetchHistory(count + 1)
}

func (b *bot) fetchDuration(d time.Duration) {
	var count int
	startTime := time.Now().Add(-d)
	if err := b.opr.DB.Model(&types.PullRequest{}).Where("owner = ? AND repo = ? AND created_at < ?",
		b.owner, b.repo, startTime).Count(&count).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		util.Error(errors.Wrap(err, "fetch duration"))
		return
	}
	b.fetchHistory(count + 1)
}

func (b *bot) fetchHistory(startID int) {
	ch := make(chan []*github.PullRequest)
	go b.fetchBatch(startID, &ch)
	for prList := range ch {
		for _, pr := range prList {
			if err := b.createOrUpdatePullRequest(pr); err != nil {
				util.Error(errors.Wrap(err, "fetch history"))
				continue
			}
			b.Middleware.cherry.ProcessPullRequest(pr)
			// if b.cfg.CherryPick {

			// }
			// if b.cfg.LabelCheck {
			// 	b.Middleware.label.ProcessPullRequest(pr)
			// }
		}
	}
}

func (b *bot) fetchBatch(startID int, ch *chan []*github.PullRequest) {
	ctx := context.Background()
	opt := &github.PullRequestListOptions{
		State: "all",
		Sort:  "created",
		ListOptions: github.ListOptions{
			Page:    1 + startID/perPage,
			PerPage: perPage,
		},
	}

	util.RetryOnError(ctx, maxRetryTime, func() error {
		prList, _, err := b.opr.Github.PullRequests.List(ctx, b.owner, b.repo, opt)
		if err != nil {
			return err
		}
		*ch <- prList
		if len(prList) < perPage {
			close(*ch)
		} else {
			go b.fetchBatch(startID+perPage, ch)
		}
		return nil
	})
}
