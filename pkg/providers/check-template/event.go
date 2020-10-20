package checkTemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"time"

	"github.com/PingCAP-QE/libs/extractor"
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

var (

	// label
	needMoreInfo     = "need-more-info"
	typeBug          = "type/bug"
	typeDuplicate    = "type/duplicate"
	typeNeedMoreInfo = "type/need-more-info"

	// gmail pwd
	specialPasswordStr = "jxjtwfjrakukiwiq"
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

	// possibly,there are more than one bug template in comments, solve the latest one.
	var templates []*github.IssueComment
	// if batch is not filled, this is last page.
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
				return err
			}

			// TODO Busy waiting waste resources
			// wait batch until written
			time.Sleep(time.Second)

			for i := 0; i < len(batch); i++ {
				if extractor.ContainsBugTemplate(*batch[i].Body){
					templates = append(templates, batch[i])
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if len(templates) != 0 {
		c.solveTemplate(issue, templates[len(templates)-1])
	} else {
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


func (c *Check) solveTemplate(issue *github.Issue, comment *github.IssueComment) error {

	bugInfo, err := extractor.ParseCommentBody(*comment.Body)

	// version is invalid
	if err != nil {
		return err
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
	b, err := ioutil.ReadFile("template.txt")
	if err != nil {
		return err
	}

	template := string(b)
	// 1.add bug template to comments
	err = c.opr.CommentOnGithub(c.owner, c.repo, *issue.Number, template)
	if err != nil {
		return err
	}

	// 2.add need-more-info label on this issue
	c.opr.Github.Issues.AddLabelsToIssue(nil, c.owner, c.repo, *issue.Number, []string{needMoreInfo})

	// 3.notify the developer in charge of this bug
	err = c.notifyBugOwner(issue)
	if err != nil {
		return err
	}

	return nil
}


func (c *Check) solveMissingFields(missingFileds []string, issue *github.Issue) error {
	// 1.add need-more-info label on this issue
	c.opr.Github.Issues.AddLabelsToIssue(context.Background(), c.owner, c.repo, *issue.Number, []string{needMoreInfo})

	// 2.add comment "(lack) fields are empty."
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
	err = c.notifyBugOwner(issue)
	if err != nil {
		return err
	}

	return nil
}

func (c *Check) notifyBugOwner(issue *github.Issue) error{
	//	send an email
	title := "Please fill in the bug template"
	body := issue.URL
	owner, err := c.getBugOwnerEmail(issue)
	if err != nil {
		return err
	}

	if owner==""{
		return ErrBugNoOwner
	}

	err = c.sendMail([]string{owner}, title, *body)
	if err != nil {
		return err
	}

	// TODO Enterprise wechat
}

func (c *Check) sendMail(mailTo []string, subject string, body string) error {

	// TODO read password.txt
	from := "jiangyuhan@pingcap.com"
	header := make(map[string]string)
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Body"] = body
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	auth := smtp.PlainAuth("", from,specialPasswordStr, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, mailTo, []byte(message))
	if err != nil {
		return err
	}
	fmt.Println("Send one email to ",mailTo)
	return nil
}

func (c *Check) getBugOwnerEmail(issue *github.Issue) (string, error) {
	//return "jiangyuhan@pingcap.com",nil
	// 1. find pull request that fix bug committer
	if issue.PullRequestLinks != nil{
		// eg:"https://api.github.com/repos/tikv/tikv/pulls/8855"
		resp,err :=http.Get(*issue.PullRequestLinks.URL)
		type User struct{
			Login string
		}
		type Pull struct {
			User User
		}

		var pull Pull
		if err != nil {
			return "",err
		}
		result, err := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(result, &pull)
		if err!=nil{
			return "",err
		}

		isIn, err := c.opr.IsInCompany(pull.User.Login)
		if err!=nil{
			return "",err
		}
		if isIn{
			return pull.User.Login,nil
		}
	}

	// 2. find bug assignee
	if issue.Assignee != nil {

		isIn, err := c.opr.IsInCompany(*issue.Assignee.Login)
		if err!=nil{
			return "",err
		}
		if isIn{
			return *issue.Assignee.Login,nil
		}
	}

	// 3. find person who close bug
	isIn, err := c.opr.IsInCompany(*issue.ClosedBy.Login)
	if err != nil {
		return "", err
	}
	if isIn {
		return *issue.ClosedBy.Login, err
	}

	// 4. find sig/component owner
	// TODO This function is not available for the time being

	// no match return ""
	return "", nil
}