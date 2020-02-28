package merge

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v29/github"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	maxRetryTime   = 1
	signedOffRegex = "^(Signed-off-by:.*)$"
)

// AutoMerge define merge database structure
type AutoMerge struct {
	ID        int       `gorm:"column:id"`
	PrID      int       `gorm:"column:pull_number"`
	Owner     string    `gorm:"column:owner"`
	Repo      string    `gorm:"column:repo"`
	Started   bool      `gorm:"column:started"`
	Status    bool      `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (m *merge) saveModel(model interface{}) error {
	ctx := context.Background()
	return errors.Wrap(util.RetryOnError(ctx, maxRetryTime, func() error {
		return m.opr.DB.Save(model).Error
	}), "save auto merge model")
}

func (m *merge) getMergeJobs() []*AutoMerge {
	var mergeJobs []*AutoMerge
	if err := m.opr.DB.Where("owner = ? and repo = ? and status = ?", m.owner, m.repo,
		false).Order("created_at asc").Find(&mergeJobs).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		util.Error(errors.Wrap(err, "get merge job from DB"))
	}
	return mergeJobs
}

func (m *merge) addGithubComment(pr *github.PullRequest, commentBody string) error {
	comment := &github.IssueComment{
		Body: &commentBody,
	}
	_, _, err := m.opr.Github.Issues.CreateComment(context.Background(),
		m.owner, m.repo, *pr.Number, comment)
	return errors.Wrap(err, "add github comment")
}

func (m *merge) updateBranch(pr *github.PullRequest) (bool, error) {
	needUpdate, err := m.needUpdateBranch(pr)
	if err != nil {
		return false, errors.Wrap(err, "update branch")
	}
	if !needUpdate {
		return false, nil
	}
	_, _, err = m.opr.Github.PullRequests.UpdateBranch(context.Background(),
		m.owner, m.repo, *pr.Number, nil)
	if err != nil {
		// return true, errors.Wrap(err, "start merge job")
		if _, ok := err.(*github.AcceptedError); ok {
			// no need for update branch, continue test
		} else {
			return true, errors.Wrap(err, "update branch")
		}
	}
	return true, nil
}

func (m *merge) needUpdateBranch(pr *github.PullRequest) (bool, error) {
	baseCommits, _, err := m.opr.Github.Repositories.ListCommits(context.Background(), m.owner, m.repo, &github.CommitsListOptions{
		SHA: pr.Base.GetRef(),
	})
	if err != nil {
		return false, errors.Wrap(err, "if need update branch")
	}
	headCommits, _, err := m.opr.Github.Repositories.ListCommits(context.Background(), m.owner, m.repo, &github.CommitsListOptions{
		SHA: pr.Head.GetSHA(),
	})
	if err != nil {
		return false, errors.Wrap(err, "if need update branch")
	}
	baseSHA := baseCommits[0].GetSHA()
	for _, commit := range headCommits {
		if baseSHA != "" && commit.GetSHA() == baseSHA {
			return false, nil
		}
	}
	return true, nil
}

func (m *merge) getMergeMessage(ID int) (string, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/pull/%d.patch", m.owner, m.repo, ID)
	res, err := http.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "ger merge message")
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", errors.Wrap(err, "ger merge message")
	}
	lines := strings.Split(string(b), "\n")

	r, _ := regexp.Compile(signedOffRegex)
	signedOffs := []string{}
	for _, line := range lines {
		m := r.FindStringSubmatch(line)
		// log.Println(line, m)
		if len(m) >= 2 {
			hasSignOff := false
			for _, signOff := range signedOffs {
				if signOff == m[1] {
					hasSignOff = true
				}
			}
			if !hasSignOff {
				signedOffs = append(signedOffs, m[1])
			}
		}
	}
	return strings.Join(signedOffs, "\n"), nil
}

func (m *merge) addCanMerge(pull *github.PullRequest) error {
	var labels []string
	for _, label := range pull.Labels {
		if label.GetName() == m.cfg.CanMergeLabel {
			return nil
		}
		labels = append(labels, label.GetName())
	}
	labels = append(labels, m.cfg.CanMergeLabel)
	_, _, err := m.opr.Github.Issues.AddLabelsToIssue(context.Background(),
		m.owner, m.repo, pull.GetNumber(), labels)
	return errors.Wrap(err, "add can merge label")
}

func (m *merge) queueComment(pull *github.PullRequest) error {
	var queue []string
	for _, job := range m.getMergeJobs() {
		if job.PrID == pull.GetNumber() {
			break
		}
		queue = append(queue, fmt.Sprintf("%d", job.PrID))
	}
	if len(queue) == 0 {
		return nil
	}
	comment := fmt.Sprintf("Your auto merge job has been accepted, waiting for %s",
		strings.Join(queue, ", "))
	return errors.Wrap(m.addGithubComment(pull, comment), "queue comment")
}

func (m *merge) failedMergeSlack(pr *github.PullRequest) error {
	msg := fmt.Sprintf("❌ Auto merge #%d failed.\nhttps://github.com/%s/%s/pull/%d",
		*pr.Number, m.owner, m.repo, *pr.Number)
	if err := m.opr.Slack.SendMessageWithPr(m.cfg.GithubBotChannel, msg, pr, "failed"); err != nil {
		return errors.Wrap(err, "failed merge report")
	}
	return nil
}

func (m *merge) successMergeSlack(pr *github.PullRequest) error {
	msg := fmt.Sprintf("✅ Auto merge #%d success.\nhttps://github.com/%s/%s/pull/%d",
		*pr.Number, m.owner, m.repo, *pr.Number)
	if err := m.opr.Slack.SendMessageWithPr(m.cfg.GithubBotChannel, msg, pr, "merged"); err != nil {
		return errors.Wrap(err, "failed merge report")
	}
	return nil
}
