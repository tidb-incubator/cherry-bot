package checkTemplate

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PingCAP-QE/libs/extractor"
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"github.com/pingcap-incubator/cherry-bot/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func initOperator() operator.Operator {
	//dbCfg := config.Database{}
	//dbConnect := db.CreateDbConnect(&dbCfg)
	httpcli := http.Client{}
	cli := github.NewClient(&httpcli)
	op := operator.Operator{
		Github: cli,
	}

	return op
}

func InitCheck() Check {
	v_operator := initOperator()
	v_cfg := config.RepoConfig{}
	cmt := Check{
		owner: "pingcap",
		repo:  "qa",
		opr:   &v_operator,
		cfg:   &v_cfg,
	}
	return cmt
}

func TestGetComments(t *testing.T) {
	c := InitCheck()
	a, _, _ := c.opr.Github.Issues.ListComments(context.Background(), "you06", "tiedb", 26, &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 2,
		},
	})
	time.Sleep(time.Second * 2)
	for i := 0; i < len(a); i++ {
		fmt.Println(a[i].CreatedAt)
	}
}

func TestSendEmail(t *testing.T) {
	title := "Please fill in the bug template"
	body := "http://www.baidu.com"
	owner := "CadmusJiang@gmail.com"
	util.SendEMail([]string{owner}, title, body)
}

func TestPR(t *testing.T) {

	type User struct {
		Login string
	}
	type Pull struct {
		User User
	}

	var pull Pull
	resp, err := http.Get("https://api.github.com/repos/tikv/tikv/pulls/8855")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("2")
	fmt.Println(pull.User)

	result, err := ioutil.ReadAll(resp.Body)
	body := fmt.Sprintf("%s", result)
	fmt.Println(body)
	err = json.Unmarshal(result, &pull)
	fmt.Println(err)
	fmt.Println(pull)
	//fmt.Println(body)
}

func TestDB(t *testing.T) {
	type CompanyEmployees struct {
		GithubID string `gorm:"column:github_id"`
		GMail    string `gorm:"column:gmail"`
	}

	// URL view help document
	url := ""
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{})
	fmt.Println(err)

	fi, err := os.Open("/Users/cadmusjiang/Desktop/employees.csv")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		if string(a) == "" {
			continue
		}
		fields := strings.Split(string(a), ",")
		fmt.Println(fields)
		employee := CompanyEmployees{GithubID: fields[1], GMail: fields[0]}
		result := db.Create(&employee)
		fmt.Println(result.RowsAffected)
	}
}

func TestLibsCheck(t *testing.T) {
	test := "## Please edit this comment to complete the following information\n\n### Not a bug\n\n1. Remove the 'type/bug' label\n2. Add notes to indicate why it is not a bug\n\n### Duplicate bug\n\n1. Add the 'type/duplicate' label\n2. Add the link to the original bug\n\n### Bug\n\nNote: Make Sure that 'component', and 'severity' labels are added\nExample for how to fill out the template: https://github.com/pingcap/tidb/issues/20100\n\n#### 1. Root Cause Analysis (RCA)\n<!-- Write down the reason why this bug occurs -->\n\n#### 2. Symptom\n\n<!-- What will the user see when this bug occurs. The error message may be in the terminal, log or monitoring -->\n\n#### 3. All Trigger Conditions\n\n<!-- All the user scenarios that may trigger this bug -->\n\n#### 4. Workaround (optional)\n\n#### 5. Affected versions\n[v4.0.1:v4.1.5]\n<!--\nIn the format of [start_version:end_version], multiple version ranges are\naccepted. If the bug only affects the unreleased version, please input:\n\"unreleased\". For example:\n\nNotes:\n  1. Do not use any white spaces in '[]'.\n  2. The range in '[]' is a closed interval\n  3. The version format is `v$Major.$Minor.$Patch`, the $Majoy and $Minor\n     number in a version range should be the same. [v3.0.1:v3.1.2] is\n     invalid because the $Minor number of the version range is different.\n\nExample 1: [v3.0.1:v3.0.5], [v4.0.1:v4.0.5]\nExample 2: unreleased\n-->\n\n#### 6. Fixed versions\n[v4.0.7]\n<!--\nThe first released version that contains this fix in each minor version. If the bug's affected version has been released, the fixed version should be a detailed version number; If the bug doesn't affect any released version, the fixed version can be \"master\". \n\nExample 1: v3.0.13, v4.0.5\nExample 2: master\n-->"
	_, errMaps := extractor.ParseCommentBody(test)
	fmt.Println(errMaps)
	test = "## Please edit this comment to complete the following information\n\n### Not a bug\n\n1. Remove the 'type/bug' label\n2. Add notes to indicate why it is not a bug\n\n### Duplicate bug\n\n1. Add the 'type/duplicate' label\n2. Add the link to the original bug\n\n### Bug\n\nNote: Make Sure that 'component', and 'severity' labels are added\nExample for how to fill out the template: https://github.com/pingcap/tidb/issues/20100\n\n#### 1. Root Cause Analysis (RCA)\n<!-- Write down the reason why this bug occurs -->\n\n#### 2. Symptom\n\n<!-- What will the user see when this bug occurs. The error message may be in the terminal, log or monitoring -->\n\n#### 3. All Trigger Conditions\n\n<!-- All the user scenarios that may trigger this bug -->\n\n#### 4. Workaround (optional)\n\n#### 5. Affected versions\nsdaff\n<!--\nIn the format of [start_version:end_version], multiple version ranges are\naccepted. If the bug only affects the unreleased version, please input:\n\"unreleased\". For example:\n\nNotes:\n  1. Do not use any white spaces in '[]'.\n  2. The range in '[]' is a closed interval\n  3. The version format is `v$Major.$Minor.$Patch`, the $Majoy and $Minor\n     number in a version range should be the same. [v3.0.1:v3.1.2] is\n     invalid because the $Minor number of the version range is different.\n\nExample 1: [v3.0.1:v3.0.5], [v4.0.1:v4.0.5]\nExample 2: unreleased\n-->\n\n#### 6. Fixed versions\nsdfsafsffa\n<!--\nThe first released version that contains this fix in each minor version. If the bug's affected version has been released, the fixed version should be a detailed version number; If the bug doesn't affect any released version, the fixed version can be \"master\". \n\nExample 1: v3.0.13, v4.0.5\nExample 2: master\n-->"
	_, errMaps = extractor.ParseCommentBody(test)
	fmt.Println(errMaps)
	test = "## Please edit this comment to complete the following information\n\n### Not a bug\n\n1. Remove the 'type/bug' label\n2. Add notes to indicate why it is not a bug\n\n### Duplicate bug\n\n1. Add the 'type/duplicate' label\n2. Add the link to the original bug\n\n### Bug\n\nNote: Make Sure that 'component', and 'severity' labels are added\nExample for how to fill out the template: https://github.com/pingcap/tidb/issues/20100\n\n#### 1. Root Cause Analysis (RCA)\n<!-- Write down the reason why this bug occurs -->\n\n#### 2. Symptom\nsdgsegs\n<!-- What will the user see when this bug occurs. The error message may be in the terminal, log or monitoring -->\n\n#### 3. All Trigger Conditions\n\n<!-- All the user scenarios that may trigger this bug -->\n\n#### 4. Workaround (optional)\n\n#### 5. Affected versions\n24234\n<!--\nIn the format of [start_version:end_version], multiple version ranges are\naccepted. If the bug only affects the unreleased version, please input:\n\"unreleased\". For example:\n\nNotes:\n  1. Do not use any white spaces in '[]'.\n  2. The range in '[]' is a closed interval\n  3. The version format is `v$Major.$Minor.$Patch`, the $Majoy and $Minor\n     number in a version range should be the same. [v3.0.1:v3.1.2] is\n     invalid because the $Minor number of the version range is different.\n\nExample 1: [v3.0.1:v3.0.5], [v4.0.1:v4.0.5]\nExample 2: unreleased\n-->\n\n#### 6. Fixed versions\n123124\n<!--\nThe first released version that contains this fix in each minor version. If the bug's affected version has been released, the fixed version should be a detailed version number; If the bug doesn't affect any released version, the fixed version can be \"master\". \n\nExample 1: v3.0.13, v4.0.5\nExample 2: master\n-->"
	_, errMaps = extractor.ParseCommentBody(test)
	fmt.Println(errMaps)
}

func TestLibsContains(t *testing.T) {
	test := "## Please edit this comment to complete the following information\n\n### Not a bug\n\n1. Remove the 'type/bug' label\n2. Add notes to indicate why it is not a bug\n\n### Duplicate bug\n\n1. Add the 'type/duplicate' label\n2. Add the link to the original bug\n\n### Bug\n\nNote: Make Sure that 'component', and 'severity' labels are added\nExample for how to fill out the template: https://github.com/pingcap/tidb/issues/20100\n\n#### 1. Root Cause Analysis (RCA)\n<!-- Write down the reason why this bug occurs -->\n\n#### 2. Symptom\nsdgsegs\n<!-- What will the user see when this bug occurs. The error message may be in the terminal, log or monitoring -->\n\n#### 3. All Trigger Conditions\n\n<!-- All the user scenarios that may trigger this bug -->\n\n#### 4. Workaround (optional)\n\n#### 5. Affected versions\n24234\n<!--\nIn the format of [start_version:end_version], multiple version ranges are\naccepted. If the bug only affects the unreleased version, please input:\n\"unreleased\". For example:\n\nNotes:\n  1. Do not use any white spaces in '[]'.\n  2. The range in '[]' is a closed interval\n  3. The version format is `v$Major.$Minor.$Patch`, the $Majoy and $Minor\n     number in a version range should be the same. [v3.0.1:v3.1.2] is\n     invalid because the $Minor number of the version range is different.\n\nExample 1: [v3.0.1:v3.0.5], [v4.0.1:v4.0.5]\nExample 2: unreleased\n-->\n\n#### 6. Fixed versions\n123124\n<!--\nThe first released version that contains this fix in each minor version. If the bug's affected version has been released, the fixed version should be a detailed version number; If the bug doesn't affect any released version, the fixed version can be \"master\". \n\nExample 1: v3.0.13, v4.0.5\nExample 2: master\n-->"
	isHave := extractor.ContainsBugTemplate(test)
	fmt.Println(isHave)
}

func TestPRURL(t *testing.T) {
	type PullRequest struct {
		URL string `json:"url"`
	}
	type Issue struct {
		PullRequest PullRequest `json:"pull_request"`
	}
	type Source struct {
		Issue Issue `json:"issue"`
	}
	type TimeLine struct {
		Source Source `json:"source"`
	}
	var timeLines []TimeLine
	timelinesURL := "https://api.github.com/repos/you06/tiedb/issues/42/timeline"
	for i := 0; i < 3; i++ {

		req, _ := http.NewRequest("GET", timelinesURL, nil)
		req.Header.Set("Accept", "application/vnd.github.mockingbird-preview+json")
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			continue
		}
		result, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		err = json.Unmarshal(result, &timeLines)
		if err != nil {
			continue
		}
		break
	}
	var pullRequestURL string
	for i := 0; i < len(timeLines); i++ {
		if timeLines[i].Source.Issue.PullRequest.URL != "" {
			pullRequestURL = timeLines[i].Source.Issue.PullRequest.URL
		}
	}
	fmt.Println(timeLines)
	fmt.Println(pullRequestURL)
}

func TestTime(c *testing.T) {
	timeObj := time.Now()
	var timeStr = timeObj.Format("2006/01/02 15:04:05")
	fmt.Println(timeStr)
}
