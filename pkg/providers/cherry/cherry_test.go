package cherry

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/v29/github"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

func getBot() (*Cherry, *gorm.DB, error) {
	repo := &config.RepoConfig{
		Owner:         "owner",
		Repo:          "repo",
		Interval:      10000000,
		Fullupdate:    10000000,
		WebhookSecret: "secret",
		Rule:          "needs-cherry-pick-([0-9.]+)",
		Release:       "release-[version]",
	}

	connect := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		"root", "", "127.0.0.1", 3306, "cherry_picker")
	db, err := gorm.Open("mysql", connect)
	if err != nil {
		return nil, nil, errors.Wrap(err, "create bot")
	}

	opr := &operator.Operator{
		Config: nil,
		DB:     db,
		Github: nil,
		Slack:  nil,
	}

	bot := InitCherry(repo, opr)
	return &bot, db, nil
}

func getPr() *github.PullRequest {
	number := 10
	title := "this is test pr (#1)"
	createdAt := time.Now()
	merged := time.Now().Add(time.Hour)
	head := "hotfix"
	base := "master"
	body := "body content"
	pr := github.PullRequest{
		Number:    &number,
		Title:     &title,
		CreatedAt: &createdAt,
		MergedAt:  &merged,
		Head: &github.PullRequestBranch{
			Label: &head,
		},
		Base: &github.PullRequestBranch{
			Ref: &base,
		},
		Body: &body,
	}
	return &pr
}

func createPr(bot *Cherry) error {
	pr := getPr()

	if err := (*bot).createPullRequest(pr); err != nil {
		return errors.Wrap(err, "insert pull request")
	}
	return nil
}

func deletePr(bot *Cherry, db *gorm.DB) error {
	pr := *getPr()
	err := db.Where("pull_number = ? and owner = ? and repo = ? ",
		pr.Number, "owner", "repo").Delete(PullRequest{}).Error
	if err != nil {
		return errors.Wrap(err, "delete PR")
	}
	return nil
}

func TestCreateCherryPick(t *testing.T) {
	log.Println("create bot...")
	bot, db, err := getBot()
	if err != nil {
		t.Errorf("create bot failed, %+v", err)
	}
	if err := createPr(bot); err != nil {
		t.Errorf("insert pull request failed, %+v", err)
	}
	if err := deletePr(bot, db); err != nil {
		t.Errorf("delete pull request failed, %+v", err)
	}
}
