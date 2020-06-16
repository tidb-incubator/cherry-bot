package command

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/google/go-github/v32/github"
)

func (c *CommandRedeliver) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	log.Printf("redeliver command: bot process issue event %s/%s #%d\n", c.repo.Owner, c.repo.Repo, event.GetIssue().GetNumber())
	// only PR author can trigger this comment
	if event.GetComment().GetUser().GetLogin() == c.opr.Config.Github.Bot {
		return
	}
	if event.GetIssue().GetUser().GetLogin() != event.GetComment().GetUser().GetLogin() {
		return
	}
	commentPattern := regexp.MustCompile(fmt.Sprintf(`@%s \/((.|\n)*)`, c.opr.Config.Github.Bot))
	// match comment
	m := commentPattern.FindStringSubmatch(event.GetComment().GetBody())
	if len(m) != 3 {
		return
	}
	comment := fmt.Sprintf("/%s", m[1])
	comment = strings.ReplaceAll(comment, "/merge", "")
	comment = strings.ReplaceAll(comment, "/run-auto-merge", "")
	comment = strings.TrimSpace(comment)
	if !strings.Contains(comment, "/run") &&
		!strings.Contains(comment, "/test") &&
		!strings.Contains(comment, "/bench") &&
		!strings.Contains(comment, "/release") {
		comment = ""
	}
	if comment == "" {
		comment = fmt.Sprintf("@%s No command or invalid command", event.GetComment().GetUser().GetLogin())
	}
	githubComment := &github.IssueComment{
		Body: &comment,
	}
	issueInfo := fmt.Sprintf("%s/%s #%d", c.repo.Owner, c.repo.Repo, event.GetIssue().GetNumber())
	if _, _, err := c.opr.Github.Issues.CreateComment(context.Background(),
		c.repo.Owner, c.repo.Repo, event.GetIssue().GetNumber(), githubComment); err != nil {
		log.Printf("error occured when redeliver command in %s, %s\n", issueInfo, err)
	} else {
		log.Printf("redeliver command success, pull %s\n", issueInfo)
	}
}
