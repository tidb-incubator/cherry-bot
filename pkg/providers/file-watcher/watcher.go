package file

import (
	"context"
	"math/rand"
	"time"

	"github.com/pingcap-incubator/cherry-bot/util"
	"github.com/pkg/errors"

	"github.com/google/go-github/v32/github"
)

// DURATION for polling
const DURATION = 20 * time.Minute

func (w *Watcher) initCheckPoints() error {
	var (
		errs []error
	)

	var jobs []job

	for _, job := range w.jobs {
		lastUpdate, sha, err := w.getLastUpdate(job.path, job.branch)
		errs = append(errs, err)
		if sha == "" || err != nil {
			continue
		}
		// accept job exists
		jobs = append(jobs, job)
		util.Println("init job", w.repo.Owner, w.repo.Repo, job.path, job.branch, lastUpdate, sha)

		w.Lock()
		if w.checkPoints[job.path] == nil {
			w.checkPoints[job.path] = make(map[string]checkPoint)
		}
		w.checkPoints[job.path][job.branch] = checkPoint{
			time: lastUpdate,
			sha:  sha,
		}
		w.Unlock()

		errs = append(errs, err)
	}

	// apply valid jobs
	w.jobs = jobs

	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Watcher) polling() {
	time.Sleep(time.Duration(rand.Intn(20)) * time.Minute)
	for range time.NewTicker(DURATION).C {
		w.checkOnce()
	}
}

func (w *Watcher) checkOnce() {
	for _, job := range w.jobs {
		util.Error(w.checkFile(job.path, job.branch))
	}
}

func (w *Watcher) getLastUpdate(path, branch string) (time.Time, string, error) {
	commits, _, err := w.opr.Github.Repositories.ListCommits(context.Background(),
		w.repo.Owner, w.repo.Repo,
		&github.CommitsListOptions{
			SHA:  branch,
			Path: path,
		})

	if err != nil {
		return time.Time{}, "", errors.Wrap(err, "list commits")
	}
	if len(commits) == 0 {
		return time.Now(), "", nil
	}

	return commits[0].GetCommit().GetCommitter().GetDate(), commits[0].GetSHA(), nil
}

func (w *Watcher) checkFile(path, branch string) error {
	lastUpdate, sha, err := w.getLastUpdate(path, branch)
	if err != nil {
		return errors.Wrap(err, "check file, get last update")
	}
	if lastUpdate.After(w.checkPoints[path][branch].time) &&
		sha != w.checkPoints[path][branch].sha {
		util.Println("before notice", w.repo.Owner, w.repo.Repo, w.checkPoints[path][branch].time,
			w.checkPoints[path][branch].sha, lastUpdate, sha)
		err := w.noticeUpdate(path, branch, w.checkPoints[path][branch].sha, sha)
		if err == nil {
			w.Lock()
			w.checkPoints[path][branch] = checkPoint{
				time: lastUpdate,
				sha:  sha,
			}
			w.Unlock()
		}
		util.Println("ater notice", w.checkPoints[path][branch].time, w.checkPoints[path][branch].sha)
		return err
	}
	return nil
}
