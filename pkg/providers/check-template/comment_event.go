package checkTemplate

import (
	"context"
	"github.com/PingCAP-QE/libs/extractor"
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
)

func (c *Check) ProcessIssueComment(event *github.IssueCommentEvent) {
	//white list
	isNeed, err := c.isNeedCheck(*event.Repo.FullName)
	if err != nil {
		return
	}
	if !isNeed {
		return
	}

	// bot create comments dont't need check
	if *event.GetSender().Login == "ti-srebot" {
		return
	}

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
	_, errMaps := extractor.ParseCommentBody(*event.Comment.Body)

	var emptyFields []string
	emptyFields = append(emptyFields, c.getMissingLabels(event.Issue.Labels)...)
	tmpEmptyFields, incorrectFields := c.getErrorsFields(errMaps)
	emptyFields = append(emptyFields, tmpEmptyFields...)

	// check bug template is valid
	if len(emptyFields) != 0 || len(incorrectFields) != 0 {
		comment := c.generateComment(emptyFields, incorrectFields)
		err := c.opr.CommentOnGithub(c.owner, c.repo, *event.Issue.Number, comment)
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
