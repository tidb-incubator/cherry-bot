package operator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/jinzhu/gorm"
	"github.com/ngaut/log"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

type SigMember struct {
	Sig        string `gorm:"sig"`
	GithubName string `gorm:"github"`
	Level      string `gorm:"level"`
	SigID      int    `gorm:"sig_id"`
}

type Sig struct {
	SigName    string `gorm:"sig_name"`
	SigID      int    `gorm:"sig_id"`
	Repo       string `gorm:"repo"`
	Label      string `gorm:"label"`
	ProjectURL string `gorm:"project_url"`
	SigUrl     string `gorm:"sig_url"`
	Channel    string `gorm:"channel"`
	Lgtm       int    `gorm:"column:lgtm"`
}

const (
	ROLE_PMC         = "pmc"
	ROLE_MAMINTAINER = "maintainer"
	ROLE_LEADER      = "leader"
	ROLE_COLEADER    = "co-leader"
	ROLE_COMMITTER   = "committer"
	ROLE_REVIEWER    = "reviewer"
)

const (
	LABEL_REQUIRE_LGT = "require-LGT"
)

var (
	MERGE_ROLES              = []string{ROLE_COMMITTER, ROLE_COLEADER, ROLE_LEADER}
	REVIEW_ROLES             = []string{ROLE_REVIEWER, ROLE_COMMITTER, ROLE_COLEADER, ROLE_LEADER}
	LABEL_REQUIRE_LGTM_LOWER = strings.ToLower(LABEL_REQUIRE_LGT)
)

func LabelsToStrArr(labels []*github.Label) []string {
	labelsArr := make([]string, len(labels))
	for _, label := range labels {
		labelsArr = append(labelsArr, label.GetName())
	}
	return labelsArr
}

func (o *Operator) ListSIGByLabel(repo string, labels []*github.Label) (sigs []*Sig, err error) {
	lablesArr := LabelsToStrArr(labels)
	err = o.DB.Where("(label in (?) or label is null) and repo=?", lablesArr, repo).Find(&sigs).Error
	if err == nil || gorm.IsRecordNotFoundError(err) {
		return sigs, nil
	}
	log.Error("get sig list failed", err)
	err = errors.Wrap(err, "get siglist")
	return
}

func (o *Operator) GetNumberOFLGTMByLable(repo string, labels []*github.Label) int {
	// check if there is a require lgtm label
	for _, label := range labels {
		name := strings.ToLower(label.GetName())
		if !strings.HasPrefix(name, LABEL_REQUIRE_LGTM_LOWER) {
			continue
		}
		lgtm_num_str := strings.TrimPrefix(name, LABEL_REQUIRE_LGTM_LOWER)
		num, err := strconv.Atoi(lgtm_num_str)
		if err != nil || num == 0 {
			log.Error("parse require lgtm failed", name, err)
			continue
		}
		log.Info("there is a require lgtm label", name, num)
		return num
	}

	sigs, err := o.ListSIGByLabel(repo, labels)
	if err != nil {
		log.Error(err)
		return 2
	}
	lgtm := 2
	for _, sig := range sigs {
		log.Info(sig.Label, sig.Lgtm)
		if sig.Lgtm < lgtm {
			lgtm = sig.Lgtm
		}
	}
	return lgtm
}

func (o *Operator) GetRolesInSigByGithubID(githubID string) (members []*SigMember, err error) {
	if err = o.DB.Where("github=?", githubID).Find(&members).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		util.Println("get roles in sig failed", err)
		err = errors.Wrap(err, "get roles in sig")
	}
	return
}

func (o *Operator) HasPermissionToPRWithLables(owner, repo string, labels []*github.Label, githubID string, roles []string) error {
	sigLabels, err := o.ListSIGByLabel(repo, labels)
	if err != nil {
		return err
	}

	legallRoles := map[string]bool{}
	for _, role := range roles {
		legallRoles[role] = true
	}
	// check members.
	members, err := o.GetRolesInSigByGithubID(githubID)
	if err != nil {
		return err
	}
	canEditSigs := map[int]bool{}
	for _, member := range members {
		if member.Level == ROLE_MAMINTAINER || member.Level == ROLE_PMC {
			return nil // the PMC or maintainer can do anything
		}
		if legallRoles[member.Level] {
			canEditSigs[member.SigID] = true
		}
	}
	// TODO this shouldn't happen in the future
	if len(sigLabels) == 0 {
		if len(canEditSigs) == 0 {
			return errors.New(fmt.Sprintf("You are not a %s.", strings.Join(roles, " or ")))
		} else {
			// when a pr doesn't belong to any sig, anyone in roles have pessimisstion to this pr now.
			return nil
		}
	}
	// check if there is one of current pr's sigs that can be editted by the user
	for _, sig := range sigLabels {
		if canEditSigs[sig.SigID] == true {
			return nil
		}
	}

	// prepare error messages
	sig_infos := []string{}
	visitedSIGs := map[int]bool{}
	for _, sig := range sigLabels {
		if visitedSIGs[sig.SigID] {
			continue
		}
		visitedSIGs[sig.SigID] = true
		sig_infos = append(sig_infos, fmt.Sprintf("[%s](%s)([slack](%s))", sig.SigName, sig.SigUrl, sig.Channel))
	}

	sigStr := "SIG"
	if len(sig_infos) > 1 {
		sigStr = "SIGs"
	}

	errMsg := fmt.Sprintf("See the corresponding SIG page for more information. Related %s: %s.", sigStr, strings.Join(sig_infos, ","))
	return errors.New(errMsg)
}

func (o *Operator) GetLGTMNumForPR(owner, repo string, pullNumber int) (num int, err error) {
	err = o.DB.Table("lgtm_records").Where("score>0 and repo=? and owner=? and pull_number=?", repo, owner, pullNumber).Count(&num).Error
	return num, err
}
