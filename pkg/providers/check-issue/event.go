package checkIssue

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"unicode"

	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

func (c *Check) ProcessIssueEvent(event *github.IssuesEvent) {
	if event.GetAction() != "opened" {
		return
	}
	if err := c.processIssue(event.GetIssue().GetHTMLURL(), event.GetIssue().GetNumber(), event.GetIssue().GetTitle(), event.GetIssue().GetBody()); err != nil {
		util.Error(err)
	}
}

func (c *Check) ProcessPREvent(event *github.PullRequestEvent) {
	if event.GetAction() != "opened" {
		return
	}
	if err := c.processIssue(event.GetPullRequest().GetHTMLURL(), event.GetPullRequest().GetNumber(), event.GetPullRequest().GetTitle(), event.GetPullRequest().GetBody()); err != nil {
		util.Error(err)
	}
}

func (c *Check) processIssue(URL string, issueID int, title string, body string) error {
	if c.IsIncludeChinese(title) || c.IsIncludeChinese(body) {
		e := c.addTemplate(issueID)
		if e != nil {
			err := errors.Wrap(e, "add template to comment fail")
			return err
		}
		e = c.SendMessage(URL + " include chinese.")
		if e != nil {
			err := errors.Wrap(e, "send wechat message fail")
			return err
		}

	}
	return nil
}

func (c *Check) addTemplate(issueID int) (err error) {
	template := "This issue contains Chinese, please use English."
	err = c.opr.CommentOnGithub(c.owner, c.repo, issueID, template)
	if err != nil {
		err = errors.Wrap(err, "add template failed")
		return err
	}

	return nil
}

func (c *Check) IsIncludeChinese(str string) bool {
	var count int
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			count++
			break
		}
	}
	return count > 0
}

func httpPostJson(url string, data map[string]interface{}) (map[string]interface{}, error) {
	xxx, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(xxx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Check) SendMessage(content string) error {
	req := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": content,
		},
	}
	data, err := ioutil.ReadFile("/root/github-bot/webhook.txt")
	if err != nil {
		return err
	}

	// remove \n
	webhook := string(data)[0 : len(data)-1]
	_, err = httpPostJson(webhook, req)
	if err != nil {
		return err
	}
	return nil
}
