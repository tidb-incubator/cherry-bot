package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/BurntSushi/toml"

	"github.com/pkg/errors"
)

const (
	defaultReleaseLGTMNeed = 1
)

// Config is cherry picker config struct
type Config struct {
	Github   *Github
	Slack    *Slack
	Repos    map[string]*RepoConfig
	Database *Database
	Member   Member `toml:"member"`
	Email    Email  `toml:"email"`
	Check    Check  `toml:"check"`
}

// Redeliver is struct of redeliver rule
type Redeliver struct {
	Label   string
	Keyword string
	Exclude string
	Channel string
}

// PullStatusControlEvent is a rule for pull status control
type PullStatusControlEvent struct {
	Duration int    `toml:"duration"`
	Events   string `toml:"events"`
}

// PullStatusControl is pull status control for a specific label
type PullStatusControl struct {
	Label  string                    `toml:"label"`
	Cond   string                    `toml:"cond"`
	Events []*PullStatusControlEvent `toml:"event"`
}

// WatchFile watches file change
type WatchFile struct {
	Path     string   `toml:"path"`
	Branches []string `toml:"branch"`
}

// RepoConfig is single repo config
type RepoConfig struct {
	// common config
	GithubBotChannel string  `toml:"github-bot-channel"`
	Member           *Member `toml:"member"`
	// repo config
	Owner          string `toml:"owner"`
	Repo           string `toml:"repo"`
	Interval       int    `toml:"interval"`
	Fullupdate     int    `toml:"fullupdate"`
	WebhookSecret  string `toml:"webhook-secret"`
	RunTestCommand string `toml:"run-test-command"`
	// cherry picker config
	CherryPick          bool   `toml:"cherry-pick"`
	Dryrun              bool   `toml:"dryrun"`
	Rule                string `toml:"cherry-pick-rule"`
	Release             string `toml:"cherry-pick-release"`
	TypeLabel           string `toml:"cherry-type-label"`
	ReplaceLabel        string `toml:"cherry-replace-label"`
	IgnoreLabel         string `toml:"ignore-label"`
	CherryPickChannel   string `toml:"cherry-pick-channel"`
	CherryConflict      bool   `toml:"cherry-pick-conflict"`
	SquashMerge         bool   `toml:"cherry-pick-squash"`
	CherryPickMilestone bool   `toml:"cherry-pick-milestone"`
	CherryPickAssign    bool   `toml:"cherry-pick-assign"`
	// label check config
	LabelCheck          bool   `toml:"label-check"`
	ShortCheckDuration  int    `toml:"short-check-duration"`
	MediumCheckDuration int    `toml:"medium-check-duration"`
	LongCheckDuration   int    `toml:"long-check-duration"`
	CommonChecker       string `toml:"common-checker"`
	ChiefChecker        string `toml:"chief-checker"`
	LabelCheckChannel   string `toml:"label-check-channel"`
	DefaultChecker      string `toml:"default-checker"`
	InviteCollaborator  bool   `toml:"invite-collaborator"`

	// pr limit config
	PrLimit          bool   `toml:"pr-limit"`
	MaxPrOpened      int    `toml:"max-pr-opened"`
	PrLimitMode      string `toml:"pr-limit-mode"`
	PrLimitOrgs      string `toml:"pr-limit-orgs"`
	PrLimitLabel     string `toml:"pr-limit-label"`
	ContributorLabel string `toml:"contributor-label"`
	// merge config
	Merge                bool   `toml:"auto-merge"`
	CanMergeLabel        string `toml:"can-merge-label"`
	ReleaseAccessControl bool   `toml:"release-access-control"`
	ReleaseLGTMNeed      int    `toml:"release-lgtm-need"`
	SignedOffMessage     bool   `toml:"signed-off-message"`
	MergeSIGControl      bool   `toml:"merge-sig-control"`
	SkipTestCommand      bool   `toml:"skip-test-command"`
	// issue redeliver
	IssueRedeliver bool         `toml:"issue-redeliver"`
	Redeliver      []*Redeliver `toml:"redeliver"`
	// pull request status control
	StatusControl     bool                 `toml:"status-control"`
	LabelOutdated     string               `toml:"label-outdated"`
	NoticeChannel     string               `toml:"notice-channel"`
	PullStatusControl []*PullStatusControl `toml:"pull-status-control"`
	// auto update config
	AutoUpdate        bool              `toml:"auto-update"`
	AutoUpdateChannel string            `toml:"auto-update-channel"`
	UpdateOwner       string            `toml:"update-owner"`
	UpdateRepo        string            `toml:"update-repo"`
	UpdateTargetMap   map[string]string `toml:"update-target-map"`
	UpdateScript      string            `toml:"update-script"`
	MergeLabel        string            `toml:"merge-label"`
	UpdateAutoMerge   bool              `toml:"update-auto-merge"`

	// add label
	DisableAddLabel bool `toml:"disable-add-label"`

	// issue notify
	IssueSlackNotice        bool   `toml:"issue-slack-notice"`
	IssueSlackNoticeChannel string `toml:"issue-slack-notice-channel"`
	IssueSlackNoticeNotify  string `toml:"issue-slack-notice-notify"`
	// approve
	PullApprove bool `toml:"pull-approve"`
	// contributor
	NotifyNewContributorPR bool `toml:"notify-new-contributor-pr"`
	// watch file change
	WatchFiles       []*WatchFile `toml:"watch-file"`
	WatchFileChannel string       `toml:"watch-file-channel"`
}

// Database is db connect config
type Database struct {
	Address  string
	Port     int
	Username string
	Password string
	Dbname   string
}

// Github config
type Github struct {
	Token string
	Bot   string
}

// Slack config
type Slack struct {
	Token     string
	Heartbeat string
	Mute      bool
	Hello     bool
}

// Member config
type Member struct {
	Orgs  []string `toml:"orgs"`
	Users []string `toml:"users"`
}

// email config
type Email struct {
	SenderAddr          string   `toml:"sender_addr"`
	DefaultReceiverAddr []string `toml:"default_receiver_addr"`
	SpecialPassWord     string   `toml:"special_password"`
}

// check config
type Check struct {
	WhiteList []string `toml:"white_list"`
}

type rawConfig struct {
	Github   *Github
	Slack    *Slack
	Repos    []*RepoConfig `toml:"repo"`
	Database *Database
	Member   Member
	Email    Email
	Check    Check
	Include  string
}

// GetConfig read config file
func GetConfig(configPath *string) (*Config, error) {
	rawCfg, err := readConfigFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "get config")
	}
	repos := make(map[string]*RepoConfig)
	for _, repo := range rawCfg.Repos {
		if repo.ReleaseLGTMNeed == 0 {
			repo.ReleaseLGTMNeed = defaultReleaseLGTMNeed
		}
		key := fmt.Sprintf("%s-%s", repo.Owner, repo.Repo)
		repos[key] = repo
	}
	return &Config{
		Github:   rawCfg.Github,
		Slack:    rawCfg.Slack,
		Repos:    repos,
		Database: rawCfg.Database,
		Member:   rawCfg.Member,
		Email:    rawCfg.Email,
		Check:    rawCfg.Check,
	}, nil
}
func readConfigFile(configPath *string) (*rawConfig, error) {
	var rawCfg rawConfig
	// read main config file.
	mainFileByte, err := ioutil.ReadFile(*configPath)
	if err != nil {
		return nil, errors.Wrap(err, "read main config file")
	}
	if _, err := toml.Decode(string(mainFileByte), &rawCfg); err != nil {
		return nil, errors.Wrap(err, "decode main config file")
	}
	// if no sub config file
	if rawCfg.Include == "" {
		return &rawCfg, nil
	}
	// read sub config files.
	dir, err := ioutil.ReadDir(rawCfg.Include)
	if err != nil {
		return nil, errors.Wrap(err, "read sub config file directory")
	}
	confBuffer := bytes.NewBuffer(mainFileByte)
	for _, f := range dir {
		if !f.IsDir() {
			realPath := path.Join(rawCfg.Include, f.Name())
			fileByte, err := ioutil.ReadFile(realPath)
			if err != nil {
				return nil, errors.Wrap(err, "read sub config file")
			}
			// merge config
			confBuffer.WriteString("\n" + string(fileByte))
		}
	}
	if _, err := toml.Decode(confBuffer.String(), &rawCfg); err != nil {
		return nil, errors.Wrap(err, "decode config file")
	}
	return &rawCfg, nil
}
