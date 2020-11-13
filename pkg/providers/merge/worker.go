package merge

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

const (
	pollingInterval = 30 * time.Second
	waitForStatus   = 120 * time.Second
	testCommentBody = "/run-all-tests"
	mergeMessage    = "Ready to merge!"
	mergeMethod     = "squash"
)

func (m *merge) processPREvent(event *github.PullRequestEvent) {
	pr := event.GetPullRequest()
	model := AutoMerge{
		PrID:      *pr.Number,
		Owner:     m.owner,
		Repo:      m.repo,
		BaseRef:   event.GetPullRequest().GetBase().GetRef(),
		Status:    mergeIncomplete,
		CreatedAt: time.Now(),
	}
	if err := m.saveModel(&model); err != nil {
		util.Error(errors.Wrap(err, "merge process PR event"))
	} else {
		util.Error(m.queueComment(pr))
	}
}

func (m *merge) startJob(mergeJob *AutoMerge) error {
	pr, _, err := m.opr.Github.PullRequests.Get(context.Background(), m.owner, m.repo, (*mergeJob).PrID)
	if err != nil {
		return errors.Wrap(err, "start merge job")
	}
	if pr.MergedAt != nil {
		return nil
	}
	needUpdate, err := m.updateBranch(pr)
	util.Println("update branch", needUpdate, err)
	if err != nil {
		if _, ok := err.(*github.AcceptedError); ok {
			// no need for update branch, continue test
			// FIXME: for some cases, we should stop the test
			// like "422 head repository does not exist" which may caused by head repository deleted
		} else {
			return errors.Wrap(err, "start merge job")
		}
	}
	if needUpdate {
		time.Sleep(waitForStatus)
	}

	commentBody := testCommentBody
	conmment := &github.IssueComment{
		Body: &commentBody,
	}
	_, _, err = m.opr.Github.Issues.CreateComment(context.Background(),
		m.owner, m.repo, *pr.Number, conmment)
	if err != nil {
		return errors.Wrap(err, "start merge job")
	}
	mergeJob.Started = true
	if err := m.saveModel(mergeJob); err != nil {
		util.Error(errors.Wrap(err, "start merge job"))
	}
	time.Sleep(waitForStatus)
	return nil
}

func (m *merge) startPolling() {
	ticker := time.NewTicker(pollingInterval)
	go func() {
		for range ticker.C {
			var wg sync.WaitGroup

			jobs := m.getMergeJobs()
			if len(jobs) == 0 {
				continue
			}

			classificationJobsList := m.classifyPR(jobs)

			for _, job := range classificationJobsList {
				wg.Add(1)

				go func(job *AutoMerge) {
					if !job.Started {
						m.startJob(job)
						job.Started = true
					}
					m.checkPR(job)
					if job.Status != mergeIncomplete {
						if err := m.saveModel(job); err != nil {
							util.Error(errors.Wrap(err, "merge polling job"))
						}
					}
					wg.Done()
				}(job)
			}

			wg.Wait()
		}
	}()
}

func (m *merge) classifyPR(jobs []*AutoMerge) (jobListOfPR map[string]*AutoMerge) {
	jobListOfPR = make(map[string]*AutoMerge, 0)
	for _, mergeJob := range jobs {
		baseRef := mergeJob.BaseRef
		if _, ok := jobListOfPR[baseRef]; !ok && mergeJob.Started {
			jobListOfPR[baseRef] = mergeJob
		}
	}

	for _, mergeJob := range jobs {
		baseRef := mergeJob.BaseRef
		if _, ok := jobListOfPR[baseRef]; !ok {
			jobListOfPR[baseRef] = mergeJob
		}
	}
	return
}

func (m *merge) checkPR(mergeJob *AutoMerge) {
	pr, _, err := m.opr.Github.PullRequests.Get(context.Background(), m.owner, m.repo, (*mergeJob).PrID)
	if err != nil {
		util.Error(errors.Wrap(err, "checking PR if can be merged"))
		return
	}
	if pr.MergedAt != nil {
		mergeJob.Status = mergeSuccess
		return
	}
	// check if still have "can merge" label
	ifHasLabel := false
	for _, l := range pr.Labels {
		if *l.Name == m.cfg.CanMergeLabel {
			ifHasLabel = true
		}
	}
	if !ifHasLabel {
		mergeJob.Status = mergeFinish
		return
	}

	// if need update, update branch & re-run test
	needUpdate, err := m.needUpdateBranch(pr)
	if err == nil && needUpdate {
		util.Println("restart job due to branch need update")
		m.startJob(mergeJob)
	}

	success := true

	status, _, err := m.opr.Github.Repositories.GetCombinedStatus(context.Background(), m.owner, m.repo,
		*pr.Head.SHA, nil)
	if err != nil {
		util.Error(errors.Wrap(err, "polling PR status"))
		return
	}
	if *status.State == "failure" || *status.State == "error" {
		success = false
		util.Println("Tests failed in statuses", status)
		if err := m.saveFailTestJob(mergeJob, status); err != nil {
			util.Println("fail to save test jobs: ", err)
		}
	}
	if *status.State == "pending" {
		return
	}

	checks, _, err := m.opr.Github.Checks.ListCheckRunsForRef(context.Background(), m.owner, m.repo,
		*pr.Head.SHA, nil)
	if err != nil {
		util.Error(errors.Wrap(err, "polling PR status"))
		return
	}
	for _, check := range checks.CheckRuns {
		if check.GetStatus() != "completed" {
			return
		}
		conclusion := check.GetConclusion()
		if conclusion != "success" && conclusion != "neutral" {
			success = false
			util.Println("Tests failed in check-runs", checks)
		}
	}

	if success {
		// send success comment and merge it
		// if err := m.addGithubComment(pr, mergeMessage); err != nil {
		// 	util.Error(errors.Wrap(err, "checking PR"))
		// }
		util.Println("Tests test passed", status, checks)
		// compose commit message
		message := " "
		if m.cfg.SignedOffMessage {
			msg, err := m.getMergeMessage(pr.GetNumber())
			if err != nil {
				util.Error(errors.Wrap(err, "merging PR"))
			} else {
				message = msg
			}
		}

		opt := github.PullRequestOptions{
			CommitTitle: fmt.Sprintf("%s (#%d)", pr.GetTitle(), pr.GetNumber()),
			MergeMethod: mergeMethod,
		}
		_, _, err := m.opr.Github.PullRequests.Merge(context.Background(), m.owner, m.repo,
			pr.GetNumber(), message, &opt)
		if err != nil {
			mergeJob.Status = mergeMergeFail
			mergeJob.Err = err.Error()
			util.Error(errors.Wrap(err, "checking PR"))
			if err := m.failedMergeSlack(pr); err != nil {
				util.Error(errors.Wrap(err, "checking PR"))
			}
		} else {
			mergeJob.Status = mergeSuccess
			if err := m.successMergeSlack(pr); err != nil {
				util.Error(errors.Wrap(err, "checking PR"))
			}
		}
	} else {
		mergeJob.Status = mergeTestFail
		// send failure comment
		comment := fmt.Sprintf("@%s merge failed.", *pr.User.Login)
		if err := m.opr.CommentOnGithub(m.owner, m.repo, pr.GetNumber(), comment); err != nil {
			util.Error(errors.Wrap(err, "checking PR"))
		}
		if err := m.failedMergeSlack(pr); err != nil {
			util.Error(errors.Wrap(err, "checking PR"))
		}
	}
}
