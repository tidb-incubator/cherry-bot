package checkTemplate

import (
	"context"
	"fmt"
	"github.com/PingCAP-QE/libs/extractor"
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"regexp"
)

var (
	unreleased      = "unreleased"
	templatePattern = regexp.MustCompile(templateStr)

	// label
	needMoreInfo     = "need-more-info"
	typeBug          = "type/bug"
	typeDuplicate    = "type/duplicate"
	typeNeedMoreInfo = "type/need-more-info"

	passwordStr = "Alkaid.io1024"
	// template
	templateStr = "## Please edit this comment to complete the following information"
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
	isOk, err := c.checkLabel(event.Issue.Labels)
	if err != nil {
		return err
	}
	if !isOk {
		return nil
	}

	if err = c.solveComments(event.Issue); err != nil {
		return err
	}
	return nil
}

func (c *Check) solveComments(issue *github.Issue) error {
	var (
		page    = 0
		perpage = 100
		batch   []*github.IssueComment
		//res     *github.Response
		err error
	)

	var templates []*github.IssueComment
	for page == 0 || len(batch) == perpage {
		page++
		if err := util.RetryOnError(context.Background(), 3, func() error {

			batch, _, err = c.opr.Github.Issues.ListComments(context.Background(), c.owner, c.repo, *issue.Number, &github.IssueListCommentsOptions{
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: perpage,
				},
			})
			if err != nil {
				//return errors.Wrap()
			}
			// TODO
			// wait batch
			time.Sleep(time.Second)

			for i := 0; i < len(batch); i++ {
				isTemplate := extractor.ContainsBugTemplate(*batch[i].Body)
				if isTemplate {
					templates = append(templates, batch[i])
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if len(templates) != 0 {
		// possibly,there are more than one bug template in comments, solve the latest one.
		fmt.Println(1)
		c.solveTemplate(issue, templates[len(templates)-1])
	} else {
		fmt.Println(2)
		c.solveNoTemplate(issue)
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
		case typeNeedMoreInfo:
			isNeedMoreInfo = true
		}
	}
	return isBug && !isDuplicate && !isNeedMoreInfo, nil
}

func (c *Check) isTemplate(comment *github.IssueComment) (bool, error) {
	//TODO check hash bug template
	temMatches := templatePattern.FindStringSubmatch(*comment.Body)
	if len(temMatches) > 0 && strings.TrimSpace(temMatches[0]) == templateStr {
		return true, nil
	}
	return false, nil
}

func (c *Check) solveTemplate(issue *github.Issue, comment *github.IssueComment) error {

	bugInfo, err := extractor.ParseCommentBody(*comment.Body)
	fmt.Println(bugInfo)
	//TODO
	if err != nil {
		//invalid version
	}

	fields := c.bugInfoIsEmpty(bugInfo)
	if len(fields) != 0 {
		c.solveMissingFields(fields, issue)
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

func (c *Check) solveNoTemplate(issue *github.Issue) error {
	b, e := ioutil.ReadFile("template.txt")
	if e != nil {
		err := errors.Wrap(e, "read template file failed")
		return err
	}

	template := string(b)
	// 1.add bug template to comments
	e = c.opr.CommentOnGithub(c.owner, c.repo, *issue.Number, template)
	if e != nil {
		err := errors.Wrap(e, "add template failed")
		return err
	}
	// 2.add need-more-info label on this issue
	c.opr.Github.Issues.AddLabelsToIssue(nil, c.owner, c.repo, *issue.Number, []string{needMoreInfo})

	// 3.notify the developer in charge of this bug
	//	send an email
	// TODO Enterprise wechat
	return nil
}

func (c *Check) getLackMandatoryField() ([]string, error) {
	return nil, nil
}

func (c *Check) solveMissingFields(missingFileds []string, issue *github.Issue) error {
	fmt.Println(missingFileds)
	// 1.add need-more-info label on this issue
	c.opr.Github.Issues.AddLabelsToIssue(context.Background(), c.owner, c.repo, *issue.Number, []string{needMoreInfo})
	// 2.add comment lack fields are emtpy
	tips := ""
	for i := 0; i < len(missingFileds); i++ {
		tips += missingFileds[i] + " "
	}
	tips = "(" + tips + ") fields are empty."
	err := c.opr.CommentOnGithub(c.owner, c.repo, *issue.Number, tips)
	if err != nil {
		return err
	}
	// 3.notify the developer in charge of this bug
	//	send an email
	title:= "Please fill in the bug template"
	body:= issue.URL
	owner,err :=c.getBugOwnerEmail(issue)
	if err!=nil{
		return err
	}
	c.sendMail([]string{owner},title, *body)
	// TODO Enterprise wechat
	return nil
}

func (c *Check) sendMail(mailTo []string, subject string, body string) error {

	from := "jiangyuhan@pingcap.com"
	// TODO read password.txt
	password := passwordStr
	to := []string{
		"CadmusJiang@gmail.com",
	}
	message := []byte("test")
	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, message)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Email Sent!")
	return nil
}

func (c *Check) getBugOwnerEmail(issue *github.Issue) (string,error){
	//return "jiangyuhan@pingcap.com",nil
	// 1. find pull request that fix bug committer
	resp,err :=http.Get(*issue.PullRequestLinks.URL)
	if err!=nil{
		return "",err
	}
	fmt.Println(resp)

	// 2. find bug assignee
	fmt.Println("3",issue.Assignee)
	if issue.Assignee != nil{
		return "",nil
	}
	// 3. find person who close bug
	fmt.Println("4",*issue.ClosedBy.Name)
	isIn,err := c.isInCompany(*issue.ClosedBy.Name)
	if err!=nil{
		return "",err
	}
	if isIn{
		return "",err
	}
	// 4. find sig/component owner
	// TODO
	return "",nil
}

func (c *Check) isInCompany(person string) (bool, error){
	return true,nil
}