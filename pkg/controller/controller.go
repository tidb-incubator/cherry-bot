package controller

import (
	"github.com/pingcap-incubator/cherry-bot/bot"
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"log"
)

// Controller is cherry picker controller interface
type Controller interface {
	GetRepo(key string) *config.RepoConfig
	GetBot(key string) *bot.Bot
	StartBots()
	Close()
}

type controller struct {
	Operator *operator.Operator
	Bots     map[string]*bot.Bot
}

// InitController create controller from plugin
func InitController(opr *operator.Operator) (Controller, error) {
	return &controller{
		Operator: opr,
		Bots:     initBots(opr),
	}, nil
}

func initBots(opr *operator.Operator) map[string]*bot.Bot {
	bots := make(map[string]*bot.Bot)
	for key, repo := range opr.Config.Repos {
		// key := fmt.Sprintf("%s-%s", repo.Owner, repo.Repo)
		log.Println("init", key)
		b := bot.InitBot(repo, opr)
		bots[key] = &b
	}
	return bots
}

// StartBots run polling job
func (ctl *controller) StartBots() {
	for _, bot := range (*ctl).Bots {
		go (*bot).StartPolling()
	}
}

// Close turn off db connect
func (ctl *controller) Close() {
	(*ctl.Operator).DB.Close()
}

// GetRepo return config of specific repo
func (ctl *controller) GetRepo(key string) *config.RepoConfig {
	return (*ctl.Operator).Config.Repos[key]
}

// GetBot return Bot instance of specific repo
func (ctl *controller) GetBot(key string) *bot.Bot {
	return (*ctl).Bots[key]
}
