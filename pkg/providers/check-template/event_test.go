package checkTemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"net/http"
	"testing"
	"time"
)

var (
	comments_json = `
	{
		"url": "https://api.github.com/repos/pingcap/tidb/issues/comments/147186180",
		"html_url": "https://github.com/pingcap/tidb/pull/351#issuecomment-147186180",
		"issue_url": "https://api.github.com/repos/pingcap/tidb/issues/351",
		"id": 147186180,
		"node_id": "MDEyOklzc3VlQ29tbWVudDE0NzE4NjE4MA==",
		"user": {
		    "login": "qiuyesuifeng",
			"id": 1953644,
			"node_id": "MDQ6VXNlcjE5NTM2NDQ=",
			"avatar_url": "https://avatars1.githubusercontent.com/u/1953644?v=4",
			"gravatar_id": "",
			"url": "https://api.github.com/users/qiuyesuifeng",
			"html_url": "https://github.com/qiuyesuifeng",
			"followers_url": "https://api.github.com/users/qiuyesuifeng/followers",
			"following_url": "https://api.github.com/users/qiuyesuifeng/following{/other_user}",
			"gists_url": "https://api.github.com/users/qiuyesuifeng/gists{/gist_id}",
			"starred_url": "https://api.github.com/users/qiuyesuifeng/starred{/owner}{/repo}",
			"subscriptions_url": "https://api.github.com/users/qiuyesuifeng/subscriptions",
			"organizations_url": "https://api.github.com/users/qiuyesuifeng/orgs",
			"repos_url": "https://api.github.com/users/qiuyesuifeng/repos",
			"events_url": "https://api.github.com/users/qiuyesuifeng/events{/privacy}",
			"received_events_url": "https://api.github.com/users/qiuyesuifeng/received_events",
			"type": "User",
			"site_admin": false
	    },
		"created_at": "2015-10-11T11:59:29Z",
		"updated_at": "2015-10-11T11:59:29Z",
		"author_association": "MEMBER",
		"body": "LGTM\n",
		"performed_via_github_app": null
	}`
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

//func GetTestConfig() (*config.Config, error) {
//	_, localFile, _, _ := runtime.Caller(0)
//	pathStr := path.Join(path.Dir(localFile), "config.example.toml")
//	cfg, err := config.GetConfig(&pathStr)
//	if err != nil {
//		return nil, err
//	}
//	return cfg, nil
//}

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

func InitIssueComment() github.IssueComment {
	comment := github.IssueComment{}
	err := json.Unmarshal([]byte(comments_json), &comment)
	if err != nil {
		fmt.Println(err)
	}
	return comment
}

func InitEvent() *github.IssueCommentEvent {
	action := "opened"
	issue_json := `
	{
		"url": "https://api.github.com/repos/pingcap/tidb/issues/351",
		"repository_url": "https://api.github.com/repos/pingcap/tidb",
		"labels_url": "https://api.github.com/repos/pingcap/tidb/issues/351/labels{/name}",
		"comments_url": "https://api.github.com/repos/pingcap/tidb/issues/351/comments",
		"events_url": "https://api.github.com/repos/pingcap/tidb/issues/351/events",
		"html_url": "https://github.com/pingcap/tidb/pull/351",
		"id": 110841121,
		"node_id": "MDExOlB1bGxSZXF1ZXN0NDczNjIxNzI=",
		"number": 351,
		"title": "*: Check float length in parser"
	}`
	comment := InitIssueComment()
	issue := github.Issue{}
	err := json.Unmarshal([]byte(issue_json), &issue)
	if err != nil {
		fmt.Println(err)
	}

	empty_json := `{}`
	change := github.EditChange{}
	err = json.Unmarshal([]byte(empty_json), &change)
	if err != nil {
		fmt.Println(err)
	}

	repo := github.Repository{}
	err = json.Unmarshal([]byte(empty_json), &repo)
	if err != nil {
		fmt.Println(err)
	}

	sender := github.User{}
	err = json.Unmarshal([]byte(empty_json), &sender)
	if err != nil {
		fmt.Println(err)
	}

	installation := github.Installation{}
	err = json.Unmarshal([]byte(empty_json), &installation)
	if err != nil {
		fmt.Println(err)
	}

	event := github.IssueCommentEvent{&action, &issue, &comment, &change,
		&repo, &sender, &installation}
	return &event

}

//func TestProcessComment(t *testing.T) {
//
//	c := InitComment()
//
//	event := InitEvent()
//	c.processComment(event, comments_json)
//}

//func TestCheckLabel(t *testing.T) {
//	c := InitCheck()
//	correctLabels := []*github.Label{&github.Label{Name: &bug}}
//	wrongLabels1 := []*github.Label{&github.Label{Name: &bug}, &github.Label{Name: &duplicate}}
//	wrongLabels2 := []*github.Label{&github.Label{Name: &bug}, &github.Label{Name: &needMoreInfo}}
//	wrongLabels3 := []*github.Label{&github.Label{Name: &bug}, &github.Label{Name: &duplicate}, &github.Label{Name: &needMoreInfo}}
//	wrongLabels4 := []*github.Label{&github.Label{Name: &duplicate}, &github.Label{Name: &needMoreInfo}}
//	isOk, err := c.checkLabel(correctLabels)
//	if err != nil {
//		t.Error("checkLabel err")
//	}
//	assert.Equal(t, true, isOk)
//
//	isOk, err = c.checkLabel(wrongLabels1)
//	if err != nil {
//		t.Error("checkLabel err")
//	}
//	assert.Equal(t, false, isOk)
//
//	isOk, err = c.checkLabel(wrongLabels2)
//	if err != nil {
//		t.Error("checkLabel err")
//	}
//	assert.Equal(t, false, isOk)
//
//	isOk, err = c.checkLabel(wrongLabels3)
//	if err != nil {
//		t.Error("checkLabel err")
//	}
//	assert.Equal(t, false, isOk)
//
//	isOk, err = c.checkLabel(wrongLabels4)
//	if err != nil {
//		t.Error("checkLabel err")
//	}
//	assert.Equal(t, false, isOk)
//}
//func TestHasTemplate(t *testing.T) {
//	c := InitCheck()
//	comment1 := &github.IssueComment{Body: &bug}
//	comment2 := &github.IssueComment{Body: &templateStr}
//
//	hasComments := []*github.IssueComment{comment1, comment2}
//	notHasComments := []*github.IssueComment{comment1}
//
//	template, err := c.hasTemplate(hasComments)
//	if err != nil {
//		t.Error("hasTemplate err")
//	}
//	assert.NotEqual(t, "", template)
//
//	template, err = c.hasTemplate(notHasComments)
//	if err != nil {
//		t.Error("hasTemplate err")
//	}
//	assert.Equal(t, "", template)
//}

func TestInit(t *testing.T) {
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

func TestInit1(t *testing.T) {
	//title:= "Please fill in the bug template"
	//body:= "http://www.baidu.com"
	//c := InitCheck()
	//issue := github.Issue{}
	//owner,_ :=c.getBugOwnerEmail(&issue)
	//c.sendMail([]string{owner},title, body)

	resp,err :=http.Get()
	if err!=nil{
		fmt.Println(err)
	}
	fmt.Println(resp)
}
