package approve

import (
	"context"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/ngaut/log"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"github.com/pkg/errors"

	"github.com/google/go-github/v32/github"
)

type LgtmRecord struct {
	Owner      string `gorm:"column:owner"`
	Repo       string `gorm:"column:repo"`
	PullNumber int    `gorm:"column:pull_number"`
	Github     string `gorm:"column:github"`
	Score      int    `gorm:"column:score"`
}

func (a *Approve) addLGTMRecord(login string, pullNumber int, labels []*github.Label) (already_exist bool, err error) {
	already_exist = false
	record := LgtmRecord{
		Owner:      a.owner,
		Repo:       a.repo,
		PullNumber: pullNumber,
		Github:     login,
		Score:      1,
	}
	txn := a.opr.DB.Begin()
	defer func() {
		txn.Commit()
		if txn.Error != nil {
			log.Error("insert lgtm recod failed with err", txn.Error)
			err = errors.New("something is wrong,please try again later.")
		}
	}()

	already_exist, _ = a.LGTMRecordExist(&record, txn)
	if already_exist {
		//err = errors.New("You already give a LGTM to this PR")
		return
	}
	err = a.opr.HasPermissionToPRWithLables(a.owner, a.repo, labels, login, operator.REVIEW_ROLES)
	if err != nil {
		return
	}
	err = txn.Save(&record).Error
	if err != nil {
		log.Warn(err)
		err = txn.Table("lgtm_records").Where("repo=? and owner=? and pull_number=?", record.Repo, record.Owner, record.PullNumber).Update("score", record.Score).Error
		return
	}
	return
}

func (a *Approve) LGTMRecordExist(record *LgtmRecord, txn *gorm.DB) (bool, error) {
	records := []LgtmRecord{}
	terr := txn.Where("score>0 and repo=? and owner=? and pull_number=? and github=?", record.Repo, record.Owner, record.PullNumber, record.Github).Find(&records).Error
	log.Error(len(records), terr)
	if terr == nil || gorm.IsRecordNotFoundError(terr) {
		return len(records) > 0, nil
	}
	//error caused by db
	log.Error(terr)
	return false, terr
}

func (a *Approve) getLGTMNum(pullNumber int) (num int, err error) {
	return a.opr.GetLGTMNumForPR(a.owner, a.repo, pullNumber)
}

func (a *Approve) correctLGTMLable(pullNumber int, labels []*github.Label) {
	lgtmNum, err := a.getLGTMNum(pullNumber)
	//Do not do anything when connect db failed
	if err != nil {
		return
	}

	lgtmPrefixLower := strings.ToLower(lgtmLabelPrefix)
	lgtmLabelLower := fmt.Sprintf("%s%d", lgtmPrefixLower, lgtmNum)
	labelAlreadyExist := false
	for _, label := range labels {
		labelName := strings.ToLower(label.GetName())
		if strings.EqualFold(lgtmLabelLower, labelName) {
			labelAlreadyExist = true
			continue
		}
		if strings.HasPrefix(labelName, lgtmPrefixLower) {
			_, e := a.opr.Github.Issues.RemoveLabelForIssue(context.Background(), a.owner, a.repo, pullNumber, label.GetName())
			if e != nil {
				log.Error(e)
			}
		}
	}
	if labelAlreadyExist || lgtmNum == 0 {
		return
	}
	lgtmLabel := fmt.Sprintf("%s%d", lgtmLabelPrefix, lgtmNum)
	_, _, e := a.opr.Github.Issues.AddLabelsToIssue(context.Background(), a.owner, a.repo, pullNumber, []string{lgtmLabel})
	if e != nil {
		log.Error(e)
	}
	needLGTMNum := a.opr.GetNumberOFLGTMByLable(a.repo, labels)
	if lgtmNum >= needLGTMNum {
		err = a.sendApprove(pullNumber)
	} else {
		err = a.dismissApprove(pullNumber)
	}
	if err != nil {
		log.Error(e)
	}
}

func (a *Approve) removeLGTMRecord(login string, pullNumber int) (err error) {
	record := LgtmRecord{
		Owner:      a.owner,
		Repo:       a.repo,
		PullNumber: pullNumber,
		Github:     login,
		Score:      -1,
	}
	txn := a.opr.DB.Begin()
	defer func() {
		txn.Commit()
		if txn.Error != nil {
			err = errors.New("something is wrong,please try again later")
			log.Error("cancel lgtm recod failed with err", txn.Error)
		}
	}()
	exist, _ := a.LGTMRecordExist(&record, txn)
	if !exist {
		err = errors.New("You never give a LGTM to this PR")
		return err
	}

	return txn.Table("lgtm_records").Where("repo=? and owner=? and pull_number=? and github=?", record.Repo, record.Owner, record.PullNumber, record.Github).Update("score", record.Score).Error
}

func (a *Approve) sendApprove(pullNumber int) error {
	if a.getApproveFromBot(pullNumber) > 0 {
		return nil
	}
	var (
		body  string = "LGTM"
		event string = "APPROVE"
	)
	review := &github.PullRequestReviewRequest{
		Body:  &body,
		Event: &event,
	}
	_, _, err := a.opr.Github.PullRequests.CreateReview(context.Background(), a.owner, a.repo, pullNumber, review)
	return errors.Wrap(err, "send approve")
}

func (a *Approve) getApproveFromBot(pullNumber int) int64 {
	reviews, _, err := a.opr.Github.PullRequests.ListReviews(context.Background(), a.owner, a.repo, pullNumber, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		log.Error(errors.Wrap(err, "dismiss approve"))
		return 0
	}

	for _, review := range reviews {
		if review.GetState() == "APPROVED" && review.GetUser().GetLogin() == a.opr.Config.Github.Bot {
			return review.GetID()
		}
	}
	return 0
}
func (a *Approve) dismissApprove(pullNumber int) error {
	dismissMessage := "approve cancel command"
	reviewID := a.getApproveFromBot(pullNumber)
	if reviewID == 0 {
		return nil
		//return a.addGithubComment(pullNumber, "bot approve review not found")
	}

	_, _, err := a.opr.Github.PullRequests.DismissReview(context.Background(), a.owner, a.repo, pullNumber, reviewID,
		&github.PullRequestReviewDismissalRequest{
			Message: &dismissMessage,
		})

	return errors.Wrap(err, "dismiss approve")
}
