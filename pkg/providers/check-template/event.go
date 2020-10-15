package check_template

import (
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"
	"io/ioutil"

	"regexp"
	"strings"
)

var (
	templateStr = "## Please edit this comment to complete the following information"
	needMoreInfo = "need-more-info"
	component = "component"
	sig = "sig"
	severity = "severity"
	typeBug = "type/bug"
	typeDuplicate  = "type/duplicate"
	typeNeedMoreInfo = "type/need-more-info"
	templatePattern = regexp.MustCompile(templateStr)

	// mandatory field
	componentLabel = "component(label)"
	severityLabel = "severity(label)"
	RCA = "RCA"
	symptom = "Symptom"
	allTriggerCon = "All Trigger Conditions"
	affectedVersion = "Affected Versions"
	fixedVersions = "Fixed Version"
)

func (c *check) ProcessIssueEvent(event *github.IssueEvent) {
	if event.GetEvent() != "closed" {
		return
	}
	if err := c.processIssue(event); err != nil {
		util.Error(err)
	}

}

func (c *check) processIssue(event *github.IssueEvent,) error {
	isOk,err := c.checkLabel(event.Issue.Labels)
	if err!=nil{
		return err
	}

	// not just closed bug
	if !isOk{
		return nil
	}
	//github.IssueListCommentsOptions{
	//	Sort:        nil,
	//	Direction:   nil,
	//	Since:       nil,
	//	ListOptions: github.ListOptions{},
	//}
	//comments,_,err:=c.opr.Github.Issues.ListComments(nil,c.owner,c.repo,)

	//TODO dummy comments
	comment1 := &github.IssueComment{Body: &typeBug}
	//comment2 := &github.IssueComment{Body: &templateStr}
	comments := []*github.IssueComment{comment1}

	templateComment ,err := c.hasTemplate(comments)
	if err!=nil{
		return err
	}

	if templateComment!="" {
		fields,err := c.getLackMandatoryField(templateComment)
		if err!=nil{
			return err
		}

		// valid
		if fields==nil{
			return nil
		}

		//c.solveMissingField(fields)
	}else{
		c.solveNoTemplate(event.Issue)
	}
	return nil
}

// checkLable check has "type/bug",not has "type/duplicate", "type/need-more-info" issueEvent
func(c *check) checkLabel(labels []*github.Label)(bool, error){
	var isBug,isDuplicate,isNeedMoreInfo=false,false,false
	for i:=0;i< len(labels);i++{
		switch *labels[i].Name{
			case typeBug: isBug = true
			case typeDuplicate: isDuplicate = true
			case typeNeedMoreInfo:isNeedMoreInfo = true
		}
	}
	return isBug&&!isDuplicate&&!isNeedMoreInfo,nil
}

func(c *check) hasTemplate(comments []*github.IssueComment)(string,error){
	//TODO check hash bug template
	for i:=0;i<len(comments);i++ {
		temMatches := templatePattern.FindStringSubmatch(*comments[i].Body)
		if len(temMatches) > 0 && strings.TrimSpace(temMatches[0]) == templateStr {
			return *comments[i].Body, nil
		}
	}
	return "",nil
}


func(c *check) solveNoTemplate(issue *github.Issue) error{
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
	c.opr.Github.Issues.AddLabelsToIssue(nil,c.owner,c.repo,*issue.Number,[]string{needMoreInfo})

	// 3.notify the developer in charge of this bug
	//	send an email
	// TODO Enterprise wechat
	return nil
}


func(c *check) getLackMandatoryField(labels []*github.Label,template string) ([]string,error){
	result := []string{}
	// 1.component(label)
	// vaild: 'component' or 'sig' is exist
	// invalid: 'component' and 'sig' are not exist
	var isComponent,isSig = false,false
	for i:=0;i< len(labels);i++{
		switch *labels[i].Name{
		case component : isComponent = true
		case sig: isSig = true
		}
	}
	if!(isComponent||isSig){
		result = append(result, componentLabel)
	}


	// 2.severity(label)
	// valid: label exist
	// invalid: label not exist
	isSeverity := false
	for i:=0;i< len(labels);i++{
		switch *labels[i].Name{
		case severity : isSeverity= true
		}
	}
	if!(isSeverity){
		result = append(result,severityLabel)
	}

	// 3.Root Cause Analysis (RCA)
	// valid: from '#### 1. Root Cause Analysis (RCA)' to '#### 2. Symptom' not empty
	// invalid:	from '#### 1. Root Cause Analysis (RCA)' to '#### 2. Symptom' empty
	start := strings.Index(template,"#### 1. Root Cause Analysis (RCA)" )
	end := strings.Index(template,"#### 2. Symptom" )
	if strings.TrimSpace(template[start:end])==""{
		result = append(result,RCA)
	}
	// 4.Symptom
	// valid: from '#### 2. Symptom' to '#### 3. All Trigger Conditions' not empty
	// invalid: from '#### 2. Symptom'	to '#### 3. All Trigger Conditions' empty
	start = strings.Index(template,"#### 2. Symptom" )
	end = strings.Index(template,"#### 3. All Trigger Conditions" )
	if strings.TrimSpace(template[start:end])==""{
		result = append(result,symptom)
	}

	// 5.All Trigger Conditions
	// valid: from '#### 3. All Trigger Conditions' to '#### 4. Workaround (optional)' not empty
	// invalid:	from '#### 3. All Trigger Conditions' to '#### 4. Workaround (optional)' empty
	start = strings.Index(template,"#### 3. All Trigger Conditions" )
	end = strings.Index(template,"#### 4. Workaround (optional)" )
	if strings.TrimSpace(template[start:end])==""{
		result = append(result,allTriggerCon)
	}

	// 6.Affected versions
	// valid: from '#### 5. Affected versions' to '#### 6. Fixed versions' not empty
	// 版本格式为 `v$Major.$Minor.$Patch`；版本必选用 [] 括起来，括号内的两个版本的 $Majoy 和 $Minor 值要求相等；可以有多组有效值，以 ‘,’ 分割
	// 版本有效值可以为‘unreleased‘
	// invalid: from '#### 5. Affected versions' to '#### 6. Fixed versions' empty || version is wrong

	start = strings.Index(template,"#### 5. Affected versions" )
	end = strings.Index(template,"#### 6. Fixed versions" )
	if strings.TrimSpace(template[start:end])==""{
		result = append(result,affectedVersion)
	}

	//TODO check version


	// 7.Fixed versions
	// valid: from '#### 6. Fixed versions' to end is not empty
	//2. 版本格式为 `v$Major.$Minor.$Patch` 或者 ‘unplanned’
	//3. 可以有多个有效值，以 ‘,’ 分割
	// invalid: from '#### 6. Fixed versions' to end is empty|| version is wrong
	start = strings.Index(template,"#### 6. Fixed versions" )
	end = len(template)
	if strings.TrimSpace(template[start:end])==""{
		result = append(result,allTriggerCon)
	}

	//TODO check version
	return result,nil
}

func(c *check) solveMissingFields(missingFileds []string, issue *github.Issue) error{
	// 1.add need-more-info label on this issue
	c.opr.Github.Issues.AddLabelsToIssue(nil,c.owner,c.repo,*issue.Number,[]string{needMoreInfo})
	// 2.add comment lack fields are emtpy
	sum:=""
	for i:=0;i<len(missingFileds);i++ {
		sum += missingFileds[i]+" "
	}
	err := c.opr.CommentOnGithub(c.owner, c.repo, *issue.Number,sum)
	if err!=nil{
		return err
	}
	// 3.notify the developer in charge of this bug
	//	send an email
	// TODO Enterprise wechat
}
