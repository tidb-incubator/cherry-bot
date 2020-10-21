package checkTemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/PingCAP-QE/libs/extractor"
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

var (
	// label
	needMoreInfo  = "need-more-info"
	typeBug       = "type/bug"
	typeDuplicate = "type/duplicate"
)

var (
	ErrBugNoOwner = errors.New("Bug is no owner")
)

func (c *Check) ProcessIssuesEvent(event *github.IssuesEvent) {
	if event.GetAction() != "closed" {
		return
	}
	if err := c.processIssues(event); err != nil {
		util.Error(err)
	}
}

func (c *Check) processIssues(event *github.IssuesEvent) error {
	// 1.check if just closed bug
	isOk, err := c.checkLabel(event.Issue.Labels)
	if err != nil {
		return err
	}
	if !isOk {
		return nil
	}

	// 2.check comments if have bug template
	if err = c.solveComments(event); err != nil {
		return err
	}
	return nil
}

func (c *Check) solveComments(event *github.IssuesEvent) error {
	var (
		page    = 0
		perpage = 100
		batch   []*github.IssueComment
		//res     *github.Response
		err error
	)

	// possibly,there are more than one bug template in comments, solve the latest one.
	var templates []*github.IssueComment
	// if batch is not filled, this is last page.
	for page == 0 || len(batch) == perpage {
		page++
		if err := util.RetryOnError(context.Background(), 3, func() error {

			batch, _, err = c.opr.Github.Issues.ListComments(context.Background(), c.owner, c.repo, *event.Issue.Number, &github.IssueListCommentsOptions{
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: perpage,
				},
			})
			if err != nil {
				return err
			}

			// TODO Busy waiting waste resources
			// wait batch until written
			time.Sleep(time.Second)

			for i := 0; i < len(batch); i++ {
				if extractor.ContainsBugTemplate(*batch[i].Body) {
					templates = append(templates, batch[i])
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if len(templates) != 0 {
		c.solveTemplate(event, templates[len(templates)-1])
	} else {
		c.solveNoTemplate(event)
	}
	return nil
}

// checkLable check has "type/bug",not has "type/duplicate", "type/need-more-info" issueEvent
func (c *Check) checkLabel(labels []*github.Label) (bool, error) {
	var isBug, isDuplicate, isNeedMoreInfo = false, false, false
	for i := 0; i < len(labels); i++ {
		switch *labels[i].Name {
		case typeBug:
			isBug = true
		case typeDuplicate:
			isDuplicate = true
		case needMoreInfo:
			isNeedMoreInfo = true
		}
	}
	return isBug && !isDuplicate && !isNeedMoreInfo, nil
}

func (c *Check) solveTemplate(event *github.IssuesEvent, comment *github.IssueComment) error {

	bugInfo, err := extractor.ParseCommentBody(*comment.Body)

	// version is invalid
	if err != nil {
		return err
	}

	fields := c.bugInfoIsEmpty(bugInfo)
	if len(fields) != 0 {
		c.solveMissingFields(fields, event)
	}

	return nil
}

func (c *Check) bugInfoIsEmpty(infos *extractor.BugInfos) []string {
	var fields []string
	if infos.Workaround == "" {
		fields = append(fields, "WorkAroud")
	}
	if infos.RCA == "" {
		fields = append(fields, "RCA")
	}
	if infos.AllTriggerConditions == "" {
		fields = append(fields, "AllTriggerConditions")
	}
	if len(infos.FixedVersions) == 0 {
		fields = append(fields, "FixedVersions")
	}
	if len(infos.AffectedVersions) == 0 {
		fields = append(fields, "AffectedVersions")
	}
	if infos.Symptom == "" {
		fields = append(fields, "Symptom")
	}
	return fields
}

func (c *Check) solveNoTemplate(event *github.IssuesEvent) error {
	b, err := ioutil.ReadFile("template.txt")
	if err != nil {
		return err
	}

	template := string(b)
	// 1.add bug template to comments
	err = c.opr.CommentOnGithub(c.owner, c.repo, *event.Issue.Number, template)
	if err != nil {
		return err
	}

	// 2.add need-more-info label on this issue
	c.opr.Github.Issues.AddLabelsToIssue(context.Background(), c.owner, c.repo, *event.Issue.Number, []string{needMoreInfo})

	// 3.notify the developer in charge of this bug
	err = c.notifyBugOwner(event)
	if err != nil {
		return err
	}

	return nil
}

func (c *Check) solveMissingFields(missingFileds []string, event *github.IssuesEvent) error {
	// 1.add need-more-info label on this issue
	c.opr.Github.Issues.AddLabelsToIssue(context.Background(), c.owner, c.repo, *event.Issue.Number, []string{needMoreInfo})

	// 2.add comment "(lack) fields are empty."
	tips := ""
	for i := 0; i < len(missingFileds); i++ {
		tips += missingFileds[i] + " "
	}
	tips = "(" + tips + ") fields are empty."
	err := c.opr.CommentOnGithub(c.owner, c.repo, *event.Issue.Number, tips)
	if err != nil {
		return err
	}

	// 3.notify the developer in charge of this bug
	err = c.notifyBugOwner(event)
	if err != nil {
		return err
	}

	return nil
}

func (c *Check) notifyBugOwner(event *github.IssuesEvent) error {
	//	send an email
	title := "Please fill the bug template"
	body := event.Issue.HTMLURL
	owner, err := c.getBugOwner(event)
	if err != nil {
		return err
	}

	if owner == "" {
		return ErrBugNoOwner
	}

	fmt.Println("notify ", owner)
	gmailAddress, err := c.opr.GetGmailByGithubID(owner)
	if err != nil {
		return err
	}
	err = util.SendMail([]string{gmailAddress}, title, *body)
	if err != nil {
		fmt.Println("fail to send to ", owner, "'s email. err:", err)
		return err
	}

	// TODO Enterprise wechat(maybe not need)
	return nil
}

func (c *Check) getBugOwner(event *github.IssuesEvent) (string, error) {
	//return "jiangyuhan@pingcap.com",nil
	// 1. find pull request that fix bug committer
	if event.Issue.PullRequestLinks != nil {
		type User struct {
			Login string
		}
		type Pull struct {
			User User
		}

		var pull Pull
		// eg:"https://api.github.com/repos/tikv/tikv/pulls/8855"
		// try three times.
		for i := 0; i < 3; i++ {
			resp, err := http.Get(*event.Issue.PullRequestLinks.URL)
			if err != nil {
				continue
			}
			result, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				continue
			}
			err = json.Unmarshal(result, &pull)
			if err != nil {
				continue
			}
			break
		}
		if pull.User.Login != "" {
			isIn, err := c.opr.IsInCompany(pull.User.Login)
			if err != nil {
				return "", err
			}
			if isIn {
				return pull.User.Login, nil
			}
		}
	}

	// 2. find bug assignee
	if event.Issue.Assignee != nil {
		isIn, err := c.opr.IsInCompany(*event.Issue.Assignee.Login)
		if err != nil {
			return "", err
		}
		if isIn {
			return *event.Issue.Assignee.Login, nil
		}
	}

	// 3. find person who close bug
	fmt.Println(*event.GetSender().Login, " closed issue")
	isIn, err := c.opr.IsInCompany(*event.GetSender().Login)
	if err != nil {
		return "", err
	}
	if isIn {
		return *event.GetSender().Login, err
	}

	// 4. find sig/component owner
	// TODO This function is not available for the time being

	// no match return ""
	return "", nil
}
