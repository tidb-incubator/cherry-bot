package file

import (
	"sync"
	"time"

	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"github.com/pingcap-incubator/cherry-bot/util"
)

type checkPoint struct {
	time time.Time
	sha  string
}

// Watcher struct
type Watcher struct {
	sync.Mutex
	repo        *config.RepoConfig
	opr         *operator.Operator
	jobs        []job
	checkPoints map[string]map[string]checkPoint
}

// Init create cherry pick middleware instance
func Init(repo *config.RepoConfig, opr *operator.Operator) *Watcher {
	watcher := &Watcher{
		repo:        repo,
		opr:         opr,
		checkPoints: make(map[string]map[string]checkPoint),
	}
	if repo.WatchFiles == nil {
		return watcher
	}

	go func() {
		util.Error(watcher.composeJobs())
		util.Error(watcher.initCheckPoints())
		util.Println("init complete", watcher.checkPoints)
		go watcher.polling()
	}()

	return watcher
}
