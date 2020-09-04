package addTemplate

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"github.com/pingcap-incubator/cherry-bot/util"
)

var (
	templatePattern = regexp.MustCompile(`\/info`)
)

func (c *Comment) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if event.GetAction() != "created" {
		return
	}
	if err := c.processComment(event, event.Comment.GetBody()); err != nil {
		util.Error(err)
	}
}

func (c *Comment) processComment(event *github.IssueCommentEvent, comment string) error {
	issueID := event.GetIssue().GetNumber()
	fmt.Println(issueID)
	//fmt.Println(comment)
	temMatches := templatePattern.FindStringSubmatch(comment)
	fmt.Println("temMatches:",temMatches)
	if len(temMatches) > 0 && strings.TrimSpace(temMatches[0]) == "/info" {
		e := c.addTemplate(issueID)
		if e != nil {
			err := errors.Wrap(e, "add template to comment fail")
			return err
		}
	}

	return nil
}

func (c *Comment) addTemplate(issueID int) (err error) {

	b, e := ioutil.ReadFile("template.txt")
	if e != nil {
		err = errors.Wrap(e, "read template file failed")
		return err
	}

	template := string(b)
	//fmt.Println(template)
	e = c.opr.CommentOnGithub(c.owner, c.repo, issueID, template)

	if e != nil {
		err = errors.Wrap(e, "add template failed")
		return err
	}

	return nil
}