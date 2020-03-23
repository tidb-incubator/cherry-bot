package main

import (
	"flag"
	"fmt"

	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/controller"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
	"github.com/pingcap-incubator/cherry-bot/route"
	"github.com/pingcap-incubator/cherry-bot/util"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kataras/iris"
)

var (
	configPath = flag.String("c", "./conf", "config path")
	address    = flag.String("addr", "0.0.0.0", "listen address")
	port       = flag.Int("port", 8080, "listen port")
)

func main() {
	flag.Parse()

	cfg, err := config.GetConfig(configPath)
	if err != nil {
		util.Fatal(err)
	}

	plg := operator.InitOperator(cfg)

	ctl, err := controller.InitController(plg)
	if err != nil {
		util.Fatal(err)
	}

	defer ctl.Close()
	ctl.StartBots()

	// api
	app := iris.Default()
	route.Wrapper(app, &ctl)

	util.Event("Cherry Picker running.")
	listen := fmt.Sprintf("%s:%d", *address, *port)
	if err := app.Run(iris.Addr(listen)); err != nil {
		util.Fatal(err)
	}
}
