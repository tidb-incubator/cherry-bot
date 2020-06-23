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

	"github.com/google/go-github/v32/github"
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
	BaseRef   string    `gorm:"column:base_ref"`
	Started   bool      `gorm:"column:started"`
	Status    bool      `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// ReleaseVersion for release records
type ReleaseVersion struct {
	ID      int        `gorm:"column:id"`
	Owner   string     `gorm:"column:owner"`
	Repo    string     `gorm:"column:repo"`
	Branch  string     `gorm:"column:branch"`
	Version string     `gorm:"column:version"`
	Start   *time.Time `gorm:"column:start"`
	End     *time.Time `gorm:"column:end"`
}

// ReleaseMember can merge PR to branches during release time
type ReleaseMember struct {
	ID     int    `gorm:"column:id"`
	Owner  string `gorm:"column:owner"`
	Repo   string `gorm:"column:repo"`
	Branch string `gorm:"column:branch"`
	User   string `gorm:"column:user"`
}

func (m *merge) saveModel(model interface{}) error {
	ctx := context.Background()
	return errors.Wrap(util.RetryOnError(ctx, maxRetryTime, func() error {
		return m.provider.Opr.DB.Save(model).Error
	}), "save auto merge model")
}

func (m *merge) getMergeJobs() []*AutoMerge {
	var mergeJobs []*AutoMerge
	if err := m.provider.Opr.DB.Where("owner = ? and repo = ? and status = ?", m.owner, m.repo,
		false).Order("created_at asc").Find(&mergeJobs).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		util.Error(errors.Wrap(err, "get merge job from DB"))
	}
	return mergeJobs
}

func (m *merge) updateBranch(pr *github.PullRequest) (bool, error) {
	needUpdate, err := m.needUpdateBranch(pr)
	if err != nil {
		return false, errors.Wrap(err, "update branch")
	}
	if !needUpdate {
		return false, nil
	}
	_, _, err = m.provider.Opr.Github.PullRequests.UpdateBranch(context.Background(),
		m.owner, m.repo, *pr.Number, nil)
	if err != nil {
		// break update branch for errors besides `github.AcceptedError`
		if _, ok := err.(*github.AcceptedError); !ok {
			return true, errors.Wrap(err, "update branch")
		}
	}
	return true, nil
}

func (m *merge) getReleaseVersions(base string) ([]*ReleaseVersion, error) {
	var releaseVersions []*ReleaseVersion
	if err := m.provider.Opr.DB.Where("owner = ? and repo = ? and branch = ?", m.owner, m.repo,
		base).Find(&releaseVersions).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "get release versions from DB")
	} else if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return releaseVersions, nil
}

func (m *merge) getReleaseMembers(base string) ([]*ReleaseMember, error) {
	var releaseMembers []*ReleaseMember
	if err := m.provider.Opr.DB.Where("owner = ? and repo = ? and branch = ?", m.owner, m.repo,
		base).Find(&releaseMembers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.Wrap(err, "get release members from DB")
	} else if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return releaseMembers, nil
}

func (m *merge) canMergeReleaseVersion(base, user string) (bool, error) {
	var (
		errMsg = "can merge release version"
		now    = time.Now()
	)
	releaseVersions, err := m.getReleaseVersions(base)
	if err != nil {
		return false, errors.Wrap(err, errMsg)
	}
	var releaseVersion *ReleaseVersion
	for _, r := range releaseVersions {
		if r.Start != nil && r.Start.Before(now) {
			if r.End == nil || r.End.After(now) {
				releaseVersion = r
				break
			}
		}
	}
	if releaseVersion == nil {
		return true, nil
	}
	// this branch's release version is in progress
	// check out if the user has permission to merge it
	members, err := m.getReleaseMembers(base)
	if err != nil {
		return false, errors.Wrap(err, errMsg)
	}
	for _, m := range members {
		if m.User == user {
			return true, nil
		}
	}
	return false, nil
}

func (m *merge) needUpdateBranch(pr *github.PullRequest) (bool, error) {
	baseCommits, _, err := m.provider.Opr.Github.Repositories.ListCommits(context.Background(), m.owner, m.repo, &github.CommitsListOptions{
		SHA: pr.Base.GetRef(),
	})
	if err != nil {
		return false, errors.Wrap(err, "if need update branch")
	}
	headCommits, _, err := m.provider.Opr.Github.Repositories.ListCommits(context.Background(), m.owner, m.repo, &github.CommitsListOptions{
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
		if label.GetName() == m.provider.CanMergeLabel {
			return nil
		}
		labels = append(labels, label.GetName())
	}
	labels = append(labels, m.provider.CanMergeLabel)
	_, _, err := m.provider.Opr.Github.Issues.AddLabelsToIssue(context.Background(),
		m.owner, m.repo, pull.GetNumber(), labels)
	return errors.Wrap(err, "add can merge label")
}

func (m *merge) removeCanMerge(pull *github.PullRequest) error {
	hasLabel := false
	for _, label := range pull.Labels {
		if label.GetName() == m.provider.CanMergeLabel {
			hasLabel = true
		}
	}
	if !hasLabel {
		return nil
	}
	_, err := m.provider.Opr.Github.Issues.RemoveLabelForIssue(context.Background(),
		m.owner, m.repo, pull.GetNumber(), m.provider.CanMergeLabel)
	return errors.Wrap(err, "add can merge label")
}

func (m *merge) queueComment(pull *github.PullRequest) error {
	var (
		baseRef     = pull.GetBase().GetRef()
		jobs        = m.getMergeJobs()
		waitingJobs []*AutoMerge
	)

	for _, job := range jobs {
		if job.PrID != pull.GetNumber() && job.BaseRef == baseRef {
			waitingJobs = append(waitingJobs, job)
		}
	}

	if len(waitingJobs) == 0 {
		return nil
	}
	comment := "Your auto merge job has been accepted, waiting for:\n"

	for _, job := range waitingJobs {
		comment += fmt.Sprintf("* %d \n", job.PrID)
	}

	return errors.Wrap(m.provider.CommentOnGithub(pull.GetNumber(), comment), "queue comment")
}

func (m *merge) failedMergeSlack(pr *github.PullRequest) error {
	msg := fmt.Sprintf("❌ Auto merge #%d failed.\nhttps://github.com/%s/%s/pull/%d",
		*pr.Number, m.owner, m.repo, *pr.Number)
	if err := m.provider.Opr.Slack.SendMessageWithPr(m.provider.GithubBotChannel, msg, pr, "failed"); err != nil {
		return errors.Wrap(err, "failed merge report")
	}
	return nil
}

func (m *merge) successMergeSlack(pr *github.PullRequest) error {
	msg := fmt.Sprintf("✅ Auto merge #%d success.\nhttps://github.com/%s/%s/pull/%d",
		*pr.Number, m.owner, m.repo, *pr.Number)
	if err := m.provider.Opr.Slack.SendMessageWithPr(m.provider.GithubBotChannel, msg, pr, "merged"); err != nil {
		return errors.Wrap(err, "failed merge report")
	}
	return nil
}
