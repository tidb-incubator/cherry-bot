// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package addTemplate

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"net/http"
	"testing"
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

func InitComment() Comment {
	v_operator := initOperator()
	v_cfg := config.RepoConfig{}
	cmt := Comment{
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

func TestProcessComment(t *testing.T) {

	c := InitComment()

	event := InitEvent()
	c.processComment(event, comments_json)
}

//
//func TestAddComments(t *testing.T){
//	c := InitComment()
//	e := c.addTemplate(351)
//	assert.Equal(t,nil,e,"pass")
//}

//
//func TestAddTemplate(t *testing.T) {
//   event := InitEvent()
//   c := InitComment()
//   c.ProcessIssueCommentEvent(event)
//
//}
