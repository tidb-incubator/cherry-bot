package checkTemplate

import (
	"context"
	"github.com/PingCAP-QE/libs/extractor"
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
)

func (c *Check) ProcessIssueComment(event *github.IssueCommentEvent) {
	if err := c.processIssueComment(event); err != nil {
		util.Error(err)
	}
}

func (c *Check) processIssueComment(event *github.IssueCommentEvent) error {
	switch *event.Action {
	case "created":
		if extractor.ContainsBugTemplate(*event.Comment.Body) {
			c.solveCreatedBugTemplateComment(event)
		}
		return nil
	case "edited":
		if extractor.ContainsBugTemplate(*event.Comment.Body) {
			c.solveEditedBugTemplateComment(event)
		}
		return nil
	default:
		return nil
	}
}

func (c *Check) solveCreatedBugTemplateComment(event *github.IssueCommentEvent) error {
	bugInfo, err := extractor.ParseCommentBody(*event.Comment.Body)

	// version is invalid
	if err != nil {
		tips := "Affected versions and Fixed versions relations are invalid."
		err := c.opr.CommentOnGithub(c.owner, c.repo, *event.Issue.Number, tips)
		if err != nil {
			return err
		}
	}

	missingFields := c.bugInfoIsEmpty(bugInfo)
	if len(missingFields) != 0 {
		// add comment "(lack) fields are empty."
		tips := ""
		for i := 0; i < len(missingFields); i++ {
			tips += missingFields[i] + " "
		}
		tips = "(" + tips + ") fields are empty."
		err := c.opr.CommentOnGithub(c.owner, c.repo, *event.Issue.Number, tips)
		if err != nil {
			return err
		}
	} else {
		// delete label
		_, err := c.opr.Github.Issues.RemoveLabelForIssue(context.Background(), c.owner, c.repo, *event.Issue.Number, needMoreInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Check) solveEditedBugTemplateComment(event *github.IssueCommentEvent) error {
	return c.solveCreatedBugTemplateComment(event)
}
