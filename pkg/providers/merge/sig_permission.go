package merge

import (
	"github.com/google/go-github/v31/github"
	"github.com/jinzhu/gorm"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

// AutoMergeWhiteName define white name for auto merge
type SigMember struct {
	Sig        string `gorm:"sig"`
	GithubName string `gorm:"github"`
	Level      string `gorm:"level"`
	SigID      int    `gorm:"sig_id"`
}

type SigLabel struct {
	SigName    string `gorm:"sig_name"`
	SigID      int    `gorm:"sig_id"`
	Repo       string `gorm:"repo"`
	Label      string `gorm:"label"`
	ProjectURL string `gorm:"project_url"`
}

func (m *merge) CanMergeToMaster(repo string, labels []*github.Label, userName string) error {
	util.Println("get list,repo", repo, "label", labels, "author", userName)
	// first you should be a committer.
	var sigMembers []*SigMember
	if err := m.opr.DB.Where("github=? and level in('committer','leader','co-leader')", userName).Find(&sigMembers).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		util.Println(err, "get member info failed")
		return errors.Wrap(err, "get member info")
	}
	if len(sigMembers) == 0 {
		util.Println("you are not a committer")
		return errors.New("You are not a committer")
	}

	// get the labels sig info
	labelArgs := []string{}
	for _, label := range labels {
		labelArgs = append(labelArgs, *label.Name)
	}
	var sigLabels []*SigLabel
	if err := m.opr.DB.Where("(label in (?) or label is null) and repo=?", labelArgs, repo).Find(&sigLabels).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		util.Println("get label list failed", err)
		return errors.Wrap(err, "get whitelist")
	}
	util.Println("len", len(sigLabels), "value,", sigLabels)
	if len(sigLabels) == 0 { // any committer can merge this PR
		return nil
	}

	canMergeSigs := map[int]bool{}
	for _, member := range sigMembers {
		canMergeSigs[member.SigID] = true
	}
	for _, sigLabel := range sigLabels {
		util.Println("the label", sigLabel)
		if canMergeSigs[sigLabel.SigID] {
			return nil
		}
	}
	return errors.New("You are not a committer for this part")
}
