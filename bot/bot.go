package bot

import (
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"github.com/pingcap-incubator/cherry-bot/pkg/types"

	addLabel "github.com/pingcap-incubator/cherry-bot/pkg/providers/add-label"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/add-template"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/approve"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/assign"
	autoUpdate "github.com/pingcap-incubator/cherry-bot/pkg/providers/auto-update"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/cherry"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/community"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/contributor"
	fileWatcher "github.com/pingcap-incubator/cherry-bot/pkg/providers/file-watcher"
	notify "github.com/pingcap-incubator/cherry-bot/pkg/providers/issue-notify"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/label"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/merge"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/prlimit"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/pullstatus"
	issueRedeliver "github.com/pingcap-incubator/cherry-bot/pkg/providers/redeliver"
	command "github.com/pingcap-incubator/cherry-bot/pkg/providers/redeliver-command"
)

// Bot contains main polling process
type Bot interface {
	StartPolling()
	Webhook(event interface{})
	MonthlyCheck() (*map[string]*[]string, error)
	GetMiddleware() Middleware
}

// Middleware defines middleware struct
type Middleware struct {
	cherry           cherry.Cherry
	label            label.Label
	Prlimit          prlimit.PrLimit
	Merge            merge.Merge
	IssueRedeliver   issueRedeliver.Redeliver
	PullStatus       pullstatus.PullStatus
	AutoUpdate       autoUpdate.AutoUpdate
	CommandRedeliver *command.CommandRedeliver
	Notify           *notify.Notify
	Approve          *approve.Approve
	Contributor      *contributor.Contributor
	FileWatcher      *fileWatcher.Watcher
	Assign           *assign.Assign
	AddLabel         *addLabel.Label
	AddTemplate      *addTemplate.Comment
	community        *community.CommunityCmd
}

type bot struct {
	// TODO: remove owner and repo field
	owner              string
	repo               string
	interval           int
	fullupdateInterval int
	rule               string
	release            string
	dryrun             bool
	opr                *operator.Operator
	cfg                *config.RepoConfig
	Repo               types.Repo
	Middleware         Middleware
}

// InitBot return bot instance
func InitBot(repo *config.RepoConfig, opr *operator.Operator) Bot {
	return &bot{
		owner:              repo.Owner,
		repo:               repo.Repo,
		interval:           repo.Interval,
		fullupdateInterval: repo.Fullupdate,
		rule:               repo.Rule,
		release:            repo.Release,
		dryrun:             repo.Dryrun,
		opr:                opr,
		cfg:                repo,
		Repo: types.Repo{
			Owner: repo.Owner,
			Repo:  repo.Repo,
		},
		Middleware: Middleware{
			cherry:           cherry.Init(repo, opr),
			label:            label.Init(repo, opr),
			Prlimit:          prlimit.Init(repo, opr),
			Merge:            merge.Init(repo, opr),
			community:        community.Init(repo, opr),
			IssueRedeliver:   issueRedeliver.Init(repo, opr),
			PullStatus:       pullstatus.Init(repo, opr),
			AutoUpdate:       autoUpdate.Init(repo, opr),
			CommandRedeliver: command.Init(repo, opr),
			Notify:           notify.Init(repo, opr),
			Approve:          approve.Init(repo, opr),
			Contributor:      contributor.Init(repo, opr),
			FileWatcher:      fileWatcher.Init(repo, opr),
			Assign:           assign.Init(repo, opr),
			AddLabel:         addLabel.Init(repo, opr),
			AddTemplate:      addTemplate.Init(repo, opr),
		},
	}
}

func (b *bot) ready() {
	b.Middleware.cherry.Ready()
	b.Middleware.label.Ready()
}

func (b *bot) GetMiddleware() Middleware {
	return b.Middleware
}
