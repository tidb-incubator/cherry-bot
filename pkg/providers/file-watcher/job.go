package file

import (
	"context"
	"regexp"
	"strings"

	"github.com/pingcap-incubator/cherry-bot/config"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

type job struct {
	path   string
	branch string
}

func (w *Watcher) composeJobs() error {
	branches, _, err := w.opr.Github.Repositories.ListBranches(context.Background(), w.repo.Owner, w.repo.Repo, &github.BranchListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return errors.Wrap(err, "compose jobs from config, get branch")
	}

	w.jobs = matchJobs(w.repo.WatchFiles, branches)

	return nil
}

func matchJobs(files []*config.WatchFile, branches []*github.Branch) []job {
	var jobs []job
	for _, file := range files {
		for _, cfgBranch := range file.Branches {
			r := regexp.MustCompile(strings.ReplaceAll(cfgBranch, "*", ".*"))
			for _, branch := range branches {
				name := branch.GetName()
				if r.MatchString(name) {
					jobs = append(jobs, job{
						path:   file.Path,
						branch: name,
					})
				}
			}
		}
	}
	return jobs
}
