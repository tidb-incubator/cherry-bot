package label

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

const (
	maxHistory    = 30 * 24 * time.Hour
	checkDuration = 3 * time.Hour
)

func (l *label) restartJobs() {
	util.Println(l.owner, l.repo, "start jobs by ready trigger")
	if jobs, err := l.getRestoreCheck(); err != nil {
		util.Error(errors.Wrap(err, "restart label check jobs"))
	} else {
		for _, job := range jobs {
			if job.HasLabel || job.SendNotice {
				continue
			}
			if issue, er := l.getIssueByID(job.PrID); er != nil {
				util.Error(errors.Wrap(er, "restart label check jobs"))
			} else {
				l.startJob(issue)
			}
		}
	}
}

func (l *label) processLabelCheck(issue *github.Issue) error {
	// c, err := l.ifCheck(issue)
	// if err != nil {
	// 	return errors.Wrap(err, "process label check")
	// }
	// if !c {
	// 	return nil
	// }
	util.Println("before start single job", issue.GetNumber())

	model, err := l.getLabelCheck(issue.GetNumber())
	if err != nil {
		return errors.Wrap(err, "process label check")
	}
	if model.PrID != 0 {
		return nil
	}

	err = l.startJob(issue)

	return errors.Wrap(err, "process label check")
}

func (l *label) startJob(issue *github.Issue) error {
	startTime := time.Now().Add(-maxHistory)
	// needCheckLater := time.Now().Add(-checkDuration)

	util.Println("start single job", issue.GetNumber())

	if issue.GetCreatedAt().After(startTime) && len(issue.Labels) == 0 {
		if err := l.saveLabelCheck(issue, false, false); err != nil {
			return errors.Wrap(err, "process label check")
		}

		durations := []int{
			l.cfg.ShortCheckDuration,
			l.cfg.MediumCheckDuration,
			l.cfg.LongCheckDuration,
		}
		for index, duration := range durations {
			if index >= len(durations)-1 {
				l.checkLater(issue, time.Duration(duration)*time.Hour,
					strings.Split(l.cfg.ChiefChecker, ","), true)
			} else {
				l.checkLater(issue, time.Duration(duration)*time.Hour,
					strings.Split(l.cfg.CommonChecker, ","), false)
			}
		}
	}
	return nil
}

func (l *label) checkLater(issue *github.Issue, duration time.Duration,
	checkSlice []string, finish bool) {
	needCheckLater := time.Now().Add(-duration)

	if issue.GetCreatedAt().After(needCheckLater) {
		// send notice
		// 	if err := l.sendLabelCheckNotice(issue); err != nil {
		// 		util.Error(errors.Wrap(err, "check label later"))
		// 	}
		// } else {
		// check later
		delay := time.Until(issue.GetCreatedAt().Add(duration))
		util.Println(l.owner, l.repo, issue.GetNumber(), delay)
		time.AfterFunc(delay, func() {
			util.Println(l.owner, l.repo, issue.GetNumber(), delay, "on going")
			updateIssue, _, err := l.opr.Github.Issues.Get(context.Background(), l.owner, l.repo, issue.GetNumber())
			if err != nil {
				util.Error(errors.Wrap(err, "process label check"))
			}
			if len(updateIssue.Labels) == 0 && updateIssue.GetState() == "open" {
				// send notice
				if err := l.sendLabelCheckNotice(issue, checkSlice); err != nil {
					util.Error(errors.Wrap(err, "process label check"))
				}
			}
			if err := l.saveLabelCheck(updateIssue, len(updateIssue.Labels) != 0, finish); err != nil {
				util.Error(errors.Wrap(err, "process label check"))
			}
		})
	}
}

func (l *label) ifCheck(issue *github.Issue) (bool, error) {
	match, err := regexp.MatchString(`\(#[0-9]+\)$`, *issue.Title)
	if err != nil {
		return false, errors.Wrap(err, "if check")
	}
	if match {
		return false, nil
	}
	if issue.ClosedAt != nil {
		return false, nil
	}
	return true, nil
}
