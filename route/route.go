package route

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pingcap-incubator/cherry-bot/pkg/controller"
	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/google/go-github/v32/github"
	"github.com/kataras/iris"
	"github.com/pkg/errors"
)

// HookBody for parsing webhook
type HookBody struct {
	Repository struct {
		FullName string `json:"full_name"`
	}
}

// Wrapper add webhook router
func Wrapper(app *iris.Application, ctl *controller.Controller) {
	// healthy test
	app.Get("/ping", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"message": "pong",
		})
	})

	// Github webhook
	app.Post("/webhook", func(ctx iris.Context) {
		r := ctx.Request()
		body, err := ioutil.ReadAll(r.Body)

		// restore body for iris ReadJSON use
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		hookBody := HookBody{}
		if err := ctx.ReadJSON(&hookBody); err != nil {
			// body parse error
			util.Error(errors.Wrap(err, "webhook post request"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString(err.Error())
			return
		}

		key := strings.Replace(hookBody.Repository.FullName, "/", "-", 1)
		repo := (*ctl).GetRepo(key)
		if repo == nil {
			// repo not in config file
			// util.Error(errors.New("unsupported repo"))
			util.Event("unsupported repo", key)
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		// restore body for github ValidatePayload use
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		payload, err := github.ValidatePayload(r, []byte(repo.WebhookSecret))
		if err != nil {
			// invalid payload
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString(err.Error())
			util.Error(errors.Wrap(err, fmt.Sprintf("%s webhook post request", key)))
			return
		}
		event, err := github.ParseWebHook(github.WebHookType(r), payload)
		//log.Debug("event", event)
		if err != nil {
			// event parse err
			util.Error(errors.Wrap(err, fmt.Sprintf("%s webhook post request", key)))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString(err.Error())
			return
		}

		bot := (*ctl).GetBot(key)
		if bot == nil {
			util.Error(errors.New("bot not found, however config exist"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}
		ctx.WriteString("ok")
		go (*bot).Webhook(event)
	})

	// monthly check
	app.Get("/history/{owner:string}/{repo:string}", func(ctx iris.Context) {
		owner := ctx.Params().Get("owner")
		repo := ctx.Params().Get("repo")
		key := owner + "-" + repo
		secret := ctx.URLParam("secret")

		if !auth(ctl, key, secret) {
			// repo not in config file or auth fail
			util.Event("unsupported repo")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		bot := (*ctl).GetBot(key)
		if bot == nil {
			util.Error(errors.New("bot not found, however config exist"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		res, err := (*bot).MonthlyCheck()
		if err != nil {
			util.Error(errors.Wrap(err, "monthly check"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("check failed")
			return
		}

		ctx.JSON(res)
		return
	})

	// diaplay allowlist
	app.Get("/prlimit/allowlist/{owner:string}/{repo:string}", func(ctx iris.Context) {
		owner := ctx.Params().Get("owner")
		repo := ctx.Params().Get("repo")
		key := owner + "-" + repo
		secret := ctx.URLParam("secret")

		if !auth(ctl, key, secret) {
			// repo not in config file or auth fail
			util.Event("unsupported repo")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		bot := (*ctl).GetBot(key)
		list, err := (*bot).GetMiddleware().Prlimit.GetAllowList()
		if err != nil {
			util.Event(errors.Wrap(err, "get allowname list"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("database query error")
			return
		}
		ctx.JSON(list)
	})

	// add allowname
	app.Post("/prlimit/allowlist/{owner:string}/{repo:string}/{username:string}", func(ctx iris.Context) {
		owner := ctx.Params().Get("owner")
		repo := ctx.Params().Get("repo")
		username := ctx.Params().Get("username")
		key := owner + "-" + repo
		secret := ctx.URLParam("secret")

		if !auth(ctl, key, secret) {
			// repo not in config file or auth fail
			util.Event("unsupported repo")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		bot := (*ctl).GetBot(key)
		err := (*bot).GetMiddleware().Prlimit.AddAllowList(username)
		if err != nil {
			util.Event(errors.Wrap(err, "get allowname list"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("database query error")
			return
		}
		ctx.WriteString("ok")
	})

	// remove allowname
	app.Delete("/prlimit/allowlist/{owner:string}/{repo:string}/{username:string}", func(ctx iris.Context) {
		owner := ctx.Params().Get("owner")
		repo := ctx.Params().Get("repo")
		username := ctx.Params().Get("username")
		key := owner + "-" + repo
		secret := ctx.URLParam("secret")

		if !auth(ctl, key, secret) {
			// repo not in config file or auth fail
			util.Event("unsupported repo")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		bot := (*ctl).GetBot(key)
		err := (*bot).GetMiddleware().Prlimit.RemoveAllowList(username)
		if err != nil {
			util.Event(errors.Wrap(err, "get allowname list"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("database query error")
			return
		}
		ctx.WriteString("ok")
	})

	// diaplay blocklist
	app.Get("/prlimit/blocklist/{owner:string}/{repo:string}", func(ctx iris.Context) {
		owner := ctx.Params().Get("owner")
		repo := ctx.Params().Get("repo")
		key := owner + "-" + repo
		secret := ctx.URLParam("secret")

		if !auth(ctl, key, secret) {
			// repo not in config file or auth fail
			util.Event("unsupported repo")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		bot := (*ctl).GetBot(key)
		list, err := (*bot).GetMiddleware().Prlimit.GetBlockList()
		if err != nil {
			util.Event(errors.Wrap(err, "get blockname list"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("database query error")
			return
		}
		ctx.JSON(list)
	})

	// add blocklist
	app.Post("/prlimit/blocklist/{owner:string}/{repo:string}/{username:string}", func(ctx iris.Context) {
		owner := ctx.Params().Get("owner")
		repo := ctx.Params().Get("repo")
		username := ctx.Params().Get("username")
		key := owner + "-" + repo
		secret := ctx.URLParam("secret")

		if !auth(ctl, key, secret) {
			// repo not in config file or auth fail
			util.Event("unsupported repo")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		bot := (*ctl).GetBot(key)
		err := (*bot).GetMiddleware().Prlimit.AddBlockList(username)
		if err != nil {
			util.Event(errors.Wrap(err, "get blockname list"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("database query error")
			return
		}
		ctx.WriteString("ok")
	})

	// remove blocklist
	app.Delete("/prlimit/blocklist/{owner:string}/{repo:string}/{username:string}", func(ctx iris.Context) {
		owner := ctx.Params().Get("owner")
		repo := ctx.Params().Get("repo")
		username := ctx.Params().Get("username")
		key := owner + "-" + repo
		secret := ctx.URLParam("secret")

		if !auth(ctl, key, secret) {
			// repo not in config file or auth fail
			util.Event("unsupported repo")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		bot := (*ctl).GetBot(key)
		err := (*bot).GetMiddleware().Prlimit.RemoveBlockList(username)
		if err != nil {
			util.Event(errors.Wrap(err, "get blockname list"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("database query error")
			return
		}
		ctx.WriteString("ok")
	})

	// display allowlist
	app.Get("/merge/allowlist/{owner:string}/{repo:string}", func(ctx iris.Context) {
		owner := ctx.Params().Get("owner")
		repo := ctx.Params().Get("repo")
		key := owner + "-" + repo
		secret := ctx.URLParam("secret")

		if !auth(ctl, key, secret) {
			// repo not in config file or auth fail
			util.Event("unsupported repo")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		bot := (*ctl).GetBot(key)
		list, err := (*bot).GetMiddleware().Merge.GetAllowList()
		if err != nil {
			util.Event(errors.Wrap(err, "get allowname list"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("database query error")
			return
		}
		ctx.JSON(list)
	})

	// add allowname
	app.Post("/merge/allowlist/{owner:string}/{repo:string}/{username:string}", func(ctx iris.Context) {
		owner := ctx.Params().Get("owner")
		repo := ctx.Params().Get("repo")
		username := ctx.Params().Get("username")
		key := owner + "-" + repo
		secret := ctx.URLParam("secret")

		if !auth(ctl, key, secret) {
			// repo not in config file or auth fail
			util.Event("unsupported repo")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		bot := (*ctl).GetBot(key)
		err := (*bot).GetMiddleware().Merge.AddAllowList(username)
		if err != nil {
			util.Event(errors.Wrap(err, "get allowname list"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("database query error")
			return
		}
		ctx.WriteString("ok")
	})

	// remove allowname
	app.Delete("/merge/allowlist/{owner:string}/{repo:string}/{username:string}", func(ctx iris.Context) {
		owner := ctx.Params().Get("owner")
		repo := ctx.Params().Get("repo")
		username := ctx.Params().Get("username")
		key := owner + "-" + repo
		secret := ctx.URLParam("secret")

		if !auth(ctl, key, secret) {
			// repo not in config file or auth fail
			util.Event("unsupported repo")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("unsupported repo")
			return
		}

		bot := (*ctl).GetBot(key)
		err := (*bot).GetMiddleware().Merge.RemoveAllowList(username)
		if err != nil {
			util.Event(errors.Wrap(err, "get allowname list"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("database query error")
			return
		}
		ctx.WriteString("ok")
	})

	app.Get("/merge/statistic/{owner:string}/{repo:string}", func(ctx iris.Context) {
		owner := ctx.Params().Get("owner")
		repo := ctx.Params().Get("repo")
		key := owner + "-" + repo

		if (*ctl).GetRepo(key) == nil {
			util.Event("repo not found")
			ctx.StatusCode(iris.StatusNotFound)
			return
		}

		bot := (*ctl).GetBot(key)
		statistic, err := (*bot).GetMiddleware().Merge.StatisticRepo(owner, repo)
		if err != nil {
			util.Event(errors.Wrap(err, "statistic merge status"))
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.WriteString("database query error")
			return
		}
		ctx.JSON(statistic)
	})
}
