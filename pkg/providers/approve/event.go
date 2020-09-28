package approve

import (
	"fmt"
	"strings"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/ngaut/log"
)

const (
	lgtmMsg         = "lgtm"
	lgtmCommand     = "/lgtm"
	approveCommand  = "/approve"
	cancelCommand   = "cancel"
	noAccessComment = "@%s, Thanks for your review, however we are sorry that your vote won't be count."
	lgtmLabelPrefix = "status/LGT"
	releasePrefix   = "release"
)

var lgtmCommands = []string{lgtmMsg, lgtmCommand, approveCommand}

func (a *Approve) ProcessPullRequestReviewEvent(event *github.PullRequestReviewEvent) {
	review := event.GetReview()
	pr := event.GetPullRequest()
	if review == nil || pr == nil {
		return
	}
	reviewer := event.GetSender().GetLogin()
	if reviewer == a.opr.Config.Github.Bot {
		return
	}
	author := pr.GetUser().GetLogin()
	pullNumber := pr.GetNumber()

	base := pr.GetBase().GetRef()
	if a.cfg.ReleaseApproveControl && strings.HasPrefix(base, releasePrefix) && !a.opr.IsAllowed(a.owner, a.repo, reviewer) {
		comment := fmt.Sprintf(noAccessComment, reviewer)
		if err := a.opr.CommentOnGithub(a.owner, a.repo, pullNumber, comment); err != nil {
			util.Error(err)
		}
		return
	}

	switch review.GetState() {
	case "approved":
		{
			a.createApprove(reviewer, author, pullNumber, pr.Labels)
		}
	case "commented":
		{
			approve, cancel := a.distinguishCommontBody(review.GetBody())
			if approve {
				a.createApprove(reviewer, author, pullNumber, pr.Labels)
			} else if cancel {
				a.cancelApprove(reviewer, pullNumber, pr.Labels)
			}
		}
	case "changes_requested":
		{
			a.removeLGTMRecord(reviewer, pullNumber)
			a.correctLGTMLable(pullNumber, pr.Labels)
		}
	}
}

func (a *Approve) distinguishCommontBody(body string) (approve bool, cancel bool) {
	approve = false
	cancel = false
	body = strings.ToLower(body)
	for _, msg := range lgtmCommands {
		if strings.HasPrefix(body, msg) {
			approve = true
			body = strings.TrimLeft(body, msg)
			break
		}
	}
	if !approve {
		return
	}
	body = strings.TrimSpace(body)
	if len(body) == 0 {
		return
	}

	approve = false
	if strings.EqualFold(body, cancelCommand) {
		cancel = true
	}
	return
}

func (a *Approve) ProcessIssueCommentEvent(event *github.IssueCommentEvent) {
	if event.GetAction() != "created" {
		return
	}
	pr := event.GetIssue()
	// if it is not a pull request
	if pr.GetPullRequestLinks() == nil {
		return
	}
	reviewer := event.GetSender().GetLogin()
	if reviewer == a.opr.Config.Github.Bot {
		return
	}
	approve, cancel := a.distinguishCommontBody(event.GetComment().GetBody())
	pullNumber := pr.GetNumber()
	if approve {
		prAuthorID := pr.GetUser().GetLogin()
		a.createApprove(reviewer, prAuthorID, pullNumber, pr.Labels)
	} else if cancel {
		a.cancelApprove(reviewer, pullNumber, pr.Labels)
	}
}

func (a *Approve) createApprove(senderID, prAuthorID string, pullNumber int, labels []*github.Label) {

	comment := "" //fmt.Sprintf("@%s,Thanks for your review.", senderID)""
	defer func() {
		log.Info(a.owner, a.repo, pullNumber, comment)
		if err := a.opr.CommentOnGithub(a.owner, a.repo, pullNumber, comment); err != nil {
			util.Error(err)
		}
	}()

	if senderID == prAuthorID {
		comment = fmt.Sprintf("@%s Sorry, You can’t approve your own PR.", senderID)
		return
	}
	alreadyExist, err := a.addLGTMRecord(senderID, pullNumber, labels)
	if alreadyExist {
		return
	}
	if err != nil {
		comment = fmt.Sprintf("%s", err)
		util.Error(err)
		return
	}
	a.correctLGTMLable(pullNumber, labels)
}

func (a *Approve) cancelApprove(senderID string, pullNumber int, labels []*github.Label) {
	comment := fmt.Sprintf("@%s,cancel success.", senderID)
	defer func() {
		log.Info(senderID, pullNumber, comment)
	}()
	if err := a.removeLGTMRecord(senderID, pullNumber); err != nil {
		util.Error(err)
		comment = fmt.Sprintf("Sorry @%s,cancel failed. %s", senderID, err)
	} else {
		a.correctLGTMLable(pullNumber, labels)
	}

	if err := a.opr.CommentOnGithub(a.owner, a.repo, pullNumber, comment); err != nil {
		util.Error(err)
	}
}
