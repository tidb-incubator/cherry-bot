package file

import (
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/stretchr/testify/assert"
)

var (
	master    = "master"
	release10 = "release-1.0"
	release20 = "release-2.0"
	release30 = "release-3.0"
	release31 = "release-3.1"
	release40 = "release-4.0"
	perf      = "perf"
	branches  = []*github.Branch{
		{Name: &master},
		{Name: &release10},
		{Name: &release20},
		{Name: &release30},
		{Name: &release31},
		{Name: &release40},
		{Name: &perf},
	}
)

func TestComposeJob(t *testing.T) {
	var (
		watchFileList = []*config.WatchFile{
			{Path: "/README.md", Branches: []string{"master", "release-3.0"}},
			{Path: "/go.mod", Branches: []string{"master", "release-*"}},
		}
		jobList = []job{
			{path: "/README.md", branch: "master"},
			{path: "/README.md", branch: "release-3.0"},
			{path: "/go.mod", branch: "master"},
			{path: "/go.mod", branch: "release-1.0"},
			{path: "/go.mod", branch: "release-2.0"},
			{path: "/go.mod", branch: "release-3.0"},
			{path: "/go.mod", branch: "release-3.1"},
			{path: "/go.mod", branch: "release-4.0"},
		}
	)

	assert.Equal(t, matchJobs(watchFileList, branches), jobList)
}
