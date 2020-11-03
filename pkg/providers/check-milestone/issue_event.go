package checkMilestone

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
)

var (
	ErrBugNoOwner = errors.New("Bug is no owner")
	ErrNetWork    = errors.New("network is not good,try three times still fail")
)

type MileStone struct {
	TotalCount int
}

type SampleMileStone struct {
	URL   string    `json:"url"`
	Title string    `json:"title"`
	DueOn time.Time `json:"due_on"`
}

type Notification struct {
	IssueURL  string
	Receivers []string
}

func (c *Check) loopCheckRepo(repo string) {
	// start at 10:00 am
	year := time.Now().Year()
	month := time.Now().Month()
	day := time.Now().Day()
	now := time.Now()
	var timeChan <-chan time.Time
	todayTen := time.Date(year, month, day, 10, 0, 0, 0, time.Local)
	tomorrowTen := todayTen.Add(time.Hour * 24)
	if time.Now().Before(todayTen) {
		// start at today
		timeChan = time.After(todayTen.Sub(now))
	} else {
		// start at tomorrow
		timeChan = time.After(tomorrowTen.Sub(now))
	}
	<-timeChan

	// check every day
	t := time.NewTicker(time.Hour * 24)
	defer t.Stop()

	// call once immediately
	c.appendLog("Time:", time.Now())
	c.checkRepo(repo)
	for {
		select {
		case <-t.C:
			c.appendLog("Time:", time.Now())
			c.checkRepo(repo)
		}
	}
}

func (c *Check) checkRepo(repo string) error {
	var milestones []SampleMileStone
	URL := "https://api.github.com/repos/" + repo + "/milestones"
	var isSuccess bool
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", URL, nil)
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			continue
		}
		result, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		err = json.Unmarshal(result, &milestones)
		if err != nil {
			continue
		}
		isSuccess = true
		break
	}
	if !isSuccess {
		c.appendLog("repo", repo, ErrNetWork)
		return ErrNetWork
	}

	c.appendLog("repo", repo)
	// check every milestone
	for i := 0; i < len(milestones); i++ {
		isNeed, err := c.isVersionNeedCheck(milestones[i].Title)
		if err != nil {
			return err
		}
		if isNeed {
			// TODO how long check
			if milestones[i].DueOn.IsZero() {
				continue
			}
			// According to the zero hour of the day
			targetDate := time.Date(milestones[i].DueOn.Year(), milestones[i].DueOn.Month(), milestones[i].DueOn.Day(), 0, 0, 0, 0, time.Local)
			isNeed, err = c.isTimeNeedCheck(targetDate)
			if err != nil {
				return err
			}
			if isNeed {
				c.appendLog("milestone", milestones[i].Title)
				c.checkMileStone(repo, milestones[i].Title)
			}
		}
	}

	return nil
}

func (c Check) isTimeNeedCheck(targetDay time.Time) (bool, error) {
	checkDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
	duration := targetDay.Sub(checkDay)
	// One week left
	if duration.Milliseconds() == 7*24*60*60*1000 {
		return true, nil
	}
	// Three Day left
	if duration.Milliseconds() == 3*24*60*60*1000 {
		return true, nil
	}
	// Two Day left
	if duration.Milliseconds() == 2*24*60*60*1000 {
		return true, nil
	}
	return false, nil
}

func (c *Check) isVersionNeedCheck(milestone string) (bool, error) {
	r, err := regexp.Compile("v(\\d*)\\.(\\d*)\\.(\\d*)$")
	if err != nil {
		return false, nil
	}
	return r.MatchString(milestone), nil
}

func (c *Check) checkMileStone(repo string, milestone string) error {
	bugIssues, err := c.getAllOpenBugIssues(repo, milestone)
	c.appendLog("OpenedBugIssues num:", len(bugIssues))
	if err != nil {
		return err
	}
	var notifications []Notification
	for i := 0; i < len(bugIssues); i++ {
		notification, err := c.solveBugIssue(bugIssues[i])
		if err != nil {
			notifications = append(notifications, Notification{IssueURL: *bugIssues[i].HTMLURL})
			continue
		}
		notifications = append(notifications, notification)
	}
	for i := 0; i < len(notifications); i++ {
		title := milestone + ": One Bug need to fix"
		body := notifications[i].IssueURL
		if len(notifications[i].Receivers) == 0 {
			c.appendLog(notifications[i].IssueURL, "no assignees")
			title += ",no assignees"
			util.SendEMail(c.opr.Config.Email.DefaultReceiverAddr, title, body)
		} else {
			c.appendLog(notifications[i].IssueURL, notifications[i].Receivers)
			util.SendEMail(c.opr.Config.Email.DefaultReceiverAddr, title, body)
		}
	}
	return nil
}

func (c *Check) solveBugIssue(issue *github.Issue) (Notification, error) {
	assignees := issue.Assignees
	var notification Notification
	notification.IssueURL = *issue.HTMLURL
	for i := 0; i < len(assignees); i++ {
		isIn, err := c.opr.IsInCompany(*assignees[i].Login)
		if err != nil {
			return notification, err
		}
		if isIn {
			addr, err := c.opr.GetGmailByGithubID(*assignees[i].Login)
			if err != nil {
				return notification, err
			}
			notification.Receivers = append(notification.Receivers, addr)
		}
	}
	return notification, nil
}

func (c *Check) getAllOpenBugIssues(repo string, milestone string) ([]*github.Issue, error) {
	var (
		page    = 0
		perpage = 300
		batch   []*github.Issue
		//res     *github.Response
		//err error
	)

	var bugIssues []*github.Issue
	// if batch is not filled, this is last page.
	for page == 0 || len(batch) == perpage {
		page++
		if err := util.RetryOnError(context.Background(), 3, func() error {

			query := "milestone:" + milestone + " " + "repo:" + repo + " " + "state:open"
			opt := &github.SearchOptions{
				Sort:      "",
				Order:     "",
				TextMatch: false,
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: perpage,
				},
			}
			result, _, err := c.opr.Github.Search.Issues(context.Background(), query, opt)
			if err != nil {
				return err
			}
			time.Sleep(time.Second * 3)
			// wait batch until written
			// TODO Busy waiting waste resources
			c.appendLog("issues", *result.Total)
			batch := result.Issues
			for i := 0; i < len(batch); i++ {
				isBug := false
				for j := 0; j < len(batch[i].Labels); j++ {
					if *batch[i].Labels[j].Name == "type/bug" {
						isBug = true
						break
					}
				}
				if isBug {
					bugIssues = append(bugIssues, batch[i])
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return bugIssues, nil
}

func (c *Check) appendLog(args ...interface{}) error {
	fd, err := os.OpenFile("/root/github-bot/logs/check_milestone.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	message := fmt.Sprintln(args...)
	buf := []byte(message)
	_, err = fd.Write(buf)
	if err != nil {
		return err
	}
	return fd.Close()
}
