package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/pingcap-incubator/cherry-bot/pkg/types"
	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const checkDuration = -30 * 24 * time.Hour

// MonthlyCheck generate report of last 30 days but exclude last day
func (b *bot) MonthlyCheck() (*map[string]*[]string, error) {
	res := make(map[string]*[]string)

	if b.cfg.CherryPick {
		var s []string
		res["cherry"] = &s
	}
	if b.cfg.LabelCheck {
		var s []string
		res["label"] = &s
	}

	var (
		count     int
		startTime = time.Now().Add(checkDuration)
	)
	if err := b.opr.DB.Model(&types.PullRequest{}).Where("owner = ? AND repo = ? AND created_at < ?",
		b.owner, b.repo, startTime).Count(&count).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "monthly check")
	}

	ch := make(chan []*github.PullRequest)
	go b.fetchBatch(count+1, &ch)

	for prList := range ch {
		for _, pr := range prList {
			if b.cfg.CherryPick {
				if r, err := b.Middleware.cherry.MonthCheck(pr); err != nil {
					util.Error(errors.Wrap(err, "monthly check"))
				} else {
					s := append(*res["cherry"], r...)
					res["cherry"] = &s
				}
			}
			// if b.cfg.LabelCheck {
			// 	if r, err := b.Middleware.label.MonthCheck(pr); err != nil {
			// 		util.Error(errors.Wrap(err, "monthly check"))
			// 	} else {
			// 		if r != "" {
			// 			s := append(*res["label"], r)
			// 			res["label"] = &s
			// 		}
			// 	}
			// }
		}
	}

	title := fmt.Sprintf("🔖 %s/%s Monthly Report (%s - %s)", b.owner, b.repo,
		startTime.Format("2006-01-02"), time.Now().Format("2006-01-02"))
	good := "👌 Absolutely perfect!"
	if b.cfg.CherryPick {
		var msg string
		if len(*res["cherry"]) > 0 {
			msg = strings.Join(*res["cherry"], "\n\n")
		} else {
			msg = good
		}
		go util.Error(b.opr.Slack.Report(b.cfg.CherryPickChannel, title, msg))
	}
	if b.cfg.LabelCheck {
		var msg string
		if len(*res["label"]) > 0 {
			msg = strings.Join(*res["label"], "\n\n")
		} else {
			msg = good
		}
		go util.Error(b.opr.Slack.Report(b.cfg.LabelCheckChannel, title, msg))
	}
	return &res, nil
}
