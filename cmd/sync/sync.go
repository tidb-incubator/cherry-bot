package main

import (
	"flag"
	"fmt"

	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/db"
	"github.com/pingcap-incubator/cherry-bot/pkg/providers/merge"
	"github.com/pingcap-incubator/cherry-bot/util"

	_ "github.com/go-sql-driver/mysql"
)

var (
	sourceConfigPath = flag.String("s", "./config.toml", "source config path")
	dstConfigPath    = flag.String("d", "./dst.toml", "destination config path")
)

func main() {
	flag.Parse()

	srcCfg, err := config.GetConfig(sourceConfigPath)
	if err != nil {
		util.Fatal(err)
	}

	dstCfg, err := config.GetConfig(dstConfigPath)
	if err != nil {
		util.Fatal(err)
	}

	srcDb := db.CreateDbConnect(srcCfg.Database)
	dstDb := db.CreateDbConnect(dstCfg.Database)

	if srcDb == nil || dstDb == nil {
		util.Fatal(fmt.Errorf("invalid database config"))
	}

	var merges []*merge.AutoMerge

	if err = srcDb.Where("!synced").Find(&merges).Error; err != nil {
		util.Fatal(err)
	}

	for _, merge := range merges {
		if err = dstDb.Create(merge).Error; err != nil {
			util.Fatal(err)
		}
		merge.Synced = true
		if err = srcDb.Save(merge).Error; err != nil {
			util.Fatal(err)
		}
	}

	var jobs []*merge.TestJob

	if err = srcDb.Where("!synced").Find(&jobs).Error; err != nil {
		util.Fatal(err)
	}

	for _, job := range jobs {
		if err = dstDb.Create(job).Error; err != nil {
			util.Fatal(err)
		}
		job.Synced = true
		if err = srcDb.Save(job).Error; err != nil {
			util.Fatal(err)
		}
	}
}
