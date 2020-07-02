package cherry

import (
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

// MonthCheck check a pull request last month
func (cherry *cherry) MonthCheck(pr *github.PullRequest) ([]string, error) {
	var res []string

	// not merged yet
	if pr.MergedAt == nil {
		return res, nil
	}

	// PR in expected duration
	for _, label := range pr.Labels {
		// if label format cherry-pick rule
		target, _, err := cherry.getTarget(*label.Name)
		if err != nil {
			// no match, skip
			continue
		}
		// if cherry-pick PR already exist
		m, err := cherry.getCherryPick(*pr.Number, target)
		if err != nil {
			return nil, errors.Wrap(err, "commit cherry pick")
		}
		if m.PrID == 0 {
			// cherry-pick PR not in database
			msg := cherry.formatPR(*pr.Number, target, "fetch PR failed, no record database")
			res = append(res, msg)
			continue
		}
		if !m.Success {
			// submit cherry-pick PR failed
			msg := cherry.formatPR(*pr.Number, target, "submit PR failed")
			res = append(res, msg)
			continue
		}
	}
	return res, nil
}

func (cherry *cherry) formatPR(from int, target string, msg string) string {
	line1 := fmt.Sprintf("From #%d to branch %s", from, target)
	line2 := fmt.Sprintf("Origin PR: https://github.com/%s/%s/pull/%d", cherry.owner, cherry.repo, from)
	line3 := fmt.Sprintf("Message: %s", msg)
	return line1 + "\n" + line2 + "\n" + line3
}
