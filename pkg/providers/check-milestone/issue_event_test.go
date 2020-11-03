package checkMilestone

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v32/github"
	cfg "github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"github.com/pingcap-incubator/cherry-bot/util"
)

func initOperator() operator.Operator {
	//dbCfg := cfg.Database{}
	//dbConnect := db.CreateDbConnect(&dbCfg)
	httpcli := http.Client{}
	cli := github.NewClient(&httpcli)
	op := operator.Operator{
		Github: cli,
		//DB:     dbConnect,
	}

	return op
}

func InitCheck() Check {
	v_operator := initOperator()
	v_cfg := cfg.RepoConfig{}
	cmt := Check{
		owner: "pingcap",
		repo:  "tidb",
		opr:   &v_operator,
		cfg:   &v_cfg,
	}
	return cmt
}

func TestSearch(t *testing.T) {
	c := InitCheck()
	var (
		page    = 0
		perpage = 100
		batch   []*github.Issue
		//res     *github.Response
		err error
	)

	// if batch is not filled, this is last page.
	for page == 0 || len(batch) == perpage {
		page++
		if err := util.RetryOnError(context.Background(), 3, func() error {

			query := "milestone:v4.0.8 repo:pingcap/tidb"
			opt := &github.SearchOptions{
				Sort:      "",
				Order:     "",
				TextMatch: false,
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: perpage,
				},
			}
			result, _, _ := c.opr.Github.Search.Issues(context.Background(), query, opt)
			if err != nil {
				return err
			}
			batch := result.Issues
			fmt.Println(batch[0].User)
			// TODO Busy waiting waste resources
			// wait batch until written
			time.Sleep(time.Second * 5)
			fmt.Println(*result.Total)
			//for i := 0; i < len(batch); i++ {
			//	fmt.Println(batch[i])
			//}
			return nil
		}); err != nil {

		}
	}
}

func TestLoopCheckRepos(t *testing.T) {
	//repos := []string{"you06/tiedb"}
	//repos := []string{"pingcap/tidb", "you06/tiedb"}
	c := InitCheck()
	go c.loopCheckRepo("pingcap/tidb")
	for {

	}
}

func TestCheckMileStone(t *testing.T) {
	c := InitCheck()
	c.checkMileStone("pingcap/tidb", "v4.0.8")
}

func TestIsNeed(t *testing.T) {
	c := InitCheck()
	fmt.Println(c.isVersionNeedCheck("v4.0.8"))
	fmt.Println(c.isVersionNeedCheck("Requirement pool"))
	fmt.Println(c.isVersionNeedCheck("v5.0.0-alpha"))
	fmt.Println(c.isVersionNeedCheck("v3.0.20"))
	fmt.Println(c.isVersionNeedCheck("v300.0.20"))
}

func TestAppendLog(t *testing.T) {
	c := InitCheck()
	err := c.appendLog("123", 1)
	fmt.Println(err)
}

func TestInit(t *testing.T) {
	httpcli := http.Client{}
	cli := github.NewClient(&httpcli)
	v_cfg := &cfg.RepoConfig{}
	config := &cfg.Config{
		Check: cfg.Check{
			WhiteList: []string{"pingcap/tidb", "you06/tiedb"},
		},
	}

	op := &operator.Operator{
		Github: cli,
		Config: config,
	}

	c := Init(v_cfg, op)
	c.appendLog("123")
}

func TestIsTimeNeedCheck(t *testing.T) {
	c := InitCheck()
	isOk, _ := c.isTimeNeedCheck(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+2, 0, 0, 0, 0, time.Local))
	fmt.Println(isOk)
	isOk, _ = c.isTimeNeedCheck(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+3, 0, 0, 0, 0, time.Local))
	fmt.Println(isOk)
	isOk, _ = c.isTimeNeedCheck(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+7, 0, 0, 0, 0, time.Local))
	fmt.Println(isOk)
}
