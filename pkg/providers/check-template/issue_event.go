package checkTemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PingCAP-QE/libs/extractor"
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

var (
	// label
	needMoreInfo    = "need-more-info"
	typeBug         = "type/bug"
	componentPrefix = "component/"
	sigPrefix       = "sig/"
	labelsFilter    = []string{"type/duplicate", "type/wontfix", "status/won't-fix", "status/can't-reproduce", "need-more-info"}
)

var (
	ErrBugNoOwner = errors.New("Bug is no owner")
)

func (c *Check) ProcessIssuesEvent(event *github.IssuesEvent) {
	// white list
	isNeed, err := c.isNeedCheck(*event.Repo.FullName)
	if err != nil {
		return
	}
	if !isNeed {
		return
	}

	// just check when issue closed
	if event.GetAction() != "closed" {
		return
	}
	if err = c.processIssues(event); err != nil {
		util.Error(err)
	}
}

func (c *Check) processIssues(event *github.IssuesEvent) error {
	// checkLable check has "type/bug",not has label that in labelsFilter issueEvent
	isOk := c.checkLabel(event.Issue.Labels)
	if !isOk {
		return nil
	}

	// 2.check comments if have bug template
	if err := c.solveComments(event); err != nil {
		return err
	}
	return nil
}

func (c *Check) isNeedCheck(repo string) (bool, error) {
	b, e := ioutil.ReadFile("/root/github-bot/need_check.txt")
	if e != nil {
		err := errors.Wrap(e, "read template file failed")
		return true, err
	}

	filters := string(b)
	repos := strings.Split(filters, ",")
	for i := 0; i < len(repos); i++ {
		if repos[i] == repo {
			return true, nil
		}
	}
	return false, nil
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

func (c *Check) checkLabel(labels []*github.Label) bool {
	var isNeed, isNotNeed = false, false
	for i := 0; i < len(labels); i++ {
		if *labels[i].Name == typeBug {
			isNeed = true
		}
		for j := 0; j < len(labelsFilter); j++ {
			if *labels[i].Name == labelsFilter[j] {
				isNotNeed = true
			}
		}
	}
	return isNeed && !isNotNeed
}

func (c *Check) solveTemplate(event *github.IssuesEvent, comment *github.IssueComment) error {
	_, errMaps := extractor.ParseCommentBody(*comment.Body)
	var emptyFields []string
	emptyFields = append(emptyFields, c.getMissingLabels(event.Issue.Labels)...)
	tmpEmptyFields, incorrectFields := c.getErrorsFields(errMaps)
	emptyFields = append(emptyFields, tmpEmptyFields...)

	if len(emptyFields) != 0 || len(incorrectFields) != 0 {
		c.solveInvalidTemplate(emptyFields, incorrectFields, event)
	}
	// valid nothing

	return nil
}

func (c *Check) getMissingLabels(labels []*github.Label) []string {
	var fields []string
	componentOrSig := false
	severity := false
	for i := 0; i < len(labels); i++ {
		if strings.HasPrefix(*labels[i].Name, componentPrefix) || strings.HasPrefix(*labels[i].Name, sigPrefix) {
			componentOrSig = true
		}
		if *labels[i].Name == "severity" {
			severity = true
		}
	}
	if !componentOrSig {
		fields = append(fields, "component or sig(label)")
	}
	if !severity {
		fields = append(fields, "severity(label)")
	}
	return fields
}

func (c *Check) getErrorsFields(errMaps map[string][]error) ([]string, []string) {

	fields := []string{"RCA", "AllTriggerConditions", "FixedVersions", "AffectedVersions", "Symptom"}
	var emptyFields []string
	var incorrectFields []string
	for i := 0; i < len(fields); i++ {
		errors := errMaps[fields[i]]
		for j := 0; j < len(errors); j++ {
			switch errors[j] {
			case extractor.ErrFieldEmpty:
				emptyFields = append(emptyFields, fields[i])
			default:
				incorrectFields = append(incorrectFields, fields[i])
			}
		}
	}
	return emptyFields, incorrectFields
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

func (c *Check) solveInvalidTemplate(emptyFields []string, incorrectFields []string, event *github.IssuesEvent) error {
	// 1.add need-more-info label on this issue
	c.opr.Github.Issues.AddLabelsToIssue(context.Background(), c.owner, c.repo, *event.Issue.Number, []string{needMoreInfo})

	// 2.add comment
	comment := c.generateComment(emptyFields, incorrectFields)
	err := c.opr.CommentOnGithub(c.owner, c.repo, *event.Issue.Number, comment)
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

func (c *Check) generateComment(emptyFields []string, incorrectFields []string) string {
	/*
		(...) fields are empty.
		the values in (...) fields are incorrect.
	*/
	emptyTips := ""
	incorrectTips := ""
	comment := ""
	if len(emptyFields) != 0 {
		for i := 0; i < len(emptyFields); i++ {
			emptyTips += emptyFields[i] + " "
		}
		comment = "**( " + emptyTips + ")** fields are empty.\n"
	}
	if len(incorrectFields) != 0 {
		for i := 0; i < len(incorrectFields); i++ {
			incorrectTips += incorrectFields[i] + " "
		}
		comment += "The values in **( " + incorrectTips + ")** fields are incorrect."
	}

	return comment
}

func (c *Check) notifyBugOwner(event *github.IssuesEvent) error {
	owners, err := c.getBugOwners(event)
	if err != nil {
		return err
	}

	//TODO if no valid owner, write it in log
	if len(owners) == 0 {
		c.appendLog("no-valid ")
		c.appendLog(*event.Issue.HTMLURL + "\n")
		return ErrBugNoOwner
	}
	c.appendLog(*event.Issue.HTMLURL + "\n")

	var gmailAddresses []string
	for i := 0; i < len(owners); i++ {
		address, err := c.opr.GetGmailByGithubID(owners[i])
		if err != nil {
			return err
		}
		gmailAddresses = append(gmailAddresses, address)
	}

	// send an email
	title := "Please fill the bug template"
	body := event.Issue.HTMLURL
	err = util.SendEMail(gmailAddresses, title, *body)
	if err != nil {
		fmt.Println("fail to send to ", gmailAddresses, " err:", err)
		return err
	}

	// TODO Enterprise wechat(maybe not need)
	return nil
}

func (c *Check) getBugOwners(event *github.IssuesEvent) ([]string, error) {
	timeObj := time.Now()
	var timeStr = timeObj.Format("2006/01/02 15:04:05")
	c.appendLog(timeStr + " ")

	var owners []string

	// 1. find pull request that will close this tissue
	type PullRequest struct {
		URL string `json:"url"`
	}
	type Issue struct {
		PullRequest PullRequest `json:"pull_request"`
	}
	type Source struct {
		Issue Issue `json:"issue"`
	}
	type TimeLine struct {
		Source Source `json:"source"`
	}
	timelinesURL := "https://api.github.com/repos/" + *event.Repo.FullName + "/issues/" + strconv.Itoa(*event.Issue.Number) + "/timeline"
	var timeLines []TimeLine
	// try 3 times for timeLines
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", timelinesURL, nil)
		req.Header.Set("Accept", "application/vnd.github.mockingbird-preview+json")
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			continue
		}
		result, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		err = json.Unmarshal(result, &timeLines)
		if err != nil {
			continue
		}
		// get success
		break
	}

	// the last one have pull_request url is the latest
	var pullRequestURL string
	for i := 0; i < len(timeLines); i++ {
		if timeLines[i].Source.Issue.PullRequest.URL != "" {
			pullRequestURL = timeLines[i].Source.Issue.PullRequest.URL
		}
	}

	if pullRequestURL != "" {
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
			resp, err := http.Get(pullRequestURL)
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
			// get success
			break
		}
		if pull.User.Login != "" {
			isIn, err := c.opr.IsInCompany(pull.User.Login)
			if err != nil {
				return []string{}, err
			}
			if isIn {
				owners = append(owners, pull.User.Login)
				return owners, nil
			}
		}
	}

	// no pr write in log
	c.appendLog("no-pr ")

	// 2. find bug assignees
	if event.Issue.Assignees != nil {
		for i := 0; i < len(event.Issue.Assignees); i++ {
			isIn, err := c.opr.IsInCompany(*event.Issue.Assignees[i].Login)
			if err != nil {
				return []string{}, err
			}
			if isIn {
				owners = append(owners, *event.Issue.Assignees[i].Login)
			}
		}
		if len(owners) != 0 {
			return owners, nil
		}
	}

	// 3. find person who close bug
	isIn, err := c.opr.IsInCompany(*event.GetSender().Login)
	if err != nil {
		return []string{}, err
	}
	if isIn {
		owners = append(owners, *event.GetSender().Login)
		return owners, err
	}

	// 4. find sig/component owner
	// TODO This function is not available for the time being

	// no match write in log

	return []string{}, nil
}

func (c *Check) appendLog(message string) error {
	fd, err := os.OpenFile("/root/github-bot/check_log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	buf := []byte(message)
	_, err = fd.Write(buf)
	if err != nil {
		return err
	}
	return fd.Close()
}
