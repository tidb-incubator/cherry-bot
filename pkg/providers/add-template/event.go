package add_template

import (
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
	"io/ioutil"
	"regexp"
	"strings"
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
	temMatches := templatePattern.FindStringSubmatch(comment)
	if strings.TrimSpace(temMatches[0]) == "/info" {
		return errors.Wrap(c.addTemplate(issueID), "label issue")
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
	e = c.opr.CommentOnGithub(c.owner, c.repo, issueID, template)

	if e != nil {
		err = errors.Wrap(e, "add template failed")
		return err
	}

	return nil
}
