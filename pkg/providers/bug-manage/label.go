package bugManage

import (
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
)

func (m *Manage) solveUnLabeled(event *github.IssuesEvent) error {
	fmt.Println("unlabeled")
	isBug := m.isBug(event.Issue.Labels)
	isNeedMoreInfo := m.isNeedMoreInfo(event.Issue.Labels)
	if !isBug && isNeedMoreInfo {
		_, err := m.opr.Github.Issues.RemoveLabelForIssue(context.Background(), m.owner, m.repo, *event.Issue.Number, "need-more-info")
		if err != nil {
			return err
		}
	}
	return nil
}
func (m *Manage) solveLabeled(event *github.IssuesEvent) error {

	if m.isNeedMoreInfo(event.Issue.Labels) && m.isBug(event.Issue.Labels) {
		isEnd := m.isEnd(event.Issue.Labels)
		if isEnd {
			_, err := m.opr.Github.Issues.RemoveLabelForIssue(context.Background(), m.owner, m.repo, *event.Issue.Number, "need-more-info")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manage) isBug(labels []*github.Label) bool {
	var isBug bool
	for i := 0; i < len(labels); i++ {
		if *labels[i].Name == "type/bug" {
			isBug = true
		}
	}
	return isBug
}
func (m *Manage) isNeedMoreInfo(labels []*github.Label) bool {
	var isBug bool
	for i := 0; i < len(labels); i++ {
		if *labels[i].Name == "need-more-info" {
			isBug = true
		}
	}
	return isBug
}

func (m *Manage) isEnd(labels []*github.Label) bool {
	var isEnd bool
	ends := []string{"type/duplicate", "status/won't-fix", "status/can't-reproduce", "type/wontfix"}
	for i := 0; i < len(labels); i++ {
		for j := 0; j < len(ends); j++ {
			if *labels[i].Name == ends[j] {
				isEnd = true
				return isEnd
			}
		}
	}
	return isEnd
}
