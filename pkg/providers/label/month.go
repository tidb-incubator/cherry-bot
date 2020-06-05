package label

import (
	"fmt"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/pkg/errors"
)

// MonthCheck check a pull request last month
func (l *label) MonthCheck(pr *github.Issue) (string, error) {
	c, err := l.ifCheck(pr)
	if err != nil {
		return "", errors.Wrap(err, "process label check")
	}
	if !c {
		return "", nil
	}

	startTime := time.Now().Add(-maxHistory)
	needCheckLater := time.Now().Add(-checkDuration)

	if pr.CreatedAt.After(startTime) && pr.CreatedAt.Before(needCheckLater) {
		if len(pr.Labels) == 0 {
			return l.formatPR(pr), nil
		}
	}
	return "", nil
}

func (l *label) formatPR(pr *github.Issue) string {
	line1 := fmt.Sprintf("Unlabeled PR: https://github.com/%s/%s/pull/%d", l.owner, l.repo, *pr.Number)
	line2 := fmt.Sprintf("Author: %s", *pr.User.Login)
	return line1 + "\n" + line2
}
