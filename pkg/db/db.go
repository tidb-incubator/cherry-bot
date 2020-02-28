package db

import (
	"fmt"
	"time"

	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/util"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// DB struct offer database connection and some common table manipulations
type DB struct {
	*gorm.DB
}

// CreateDbConnect create database connect to TiDB or MySQL
func CreateDbConnect(config *config.Database) *DB {
	connect := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.Address, config.Port, config.Dbname)
	db, err := gorm.Open("mysql", connect)
	if err != nil {
		// connect error
		util.Fatal(errors.Wrap(err, "create db connect"))
		return nil
	}
	// TODO: make log mode this configurable
	// db.LogMode(true)
	db.DB().SetConnMaxLifetime(time.Minute)
	db.DB().SetMaxIdleConns(0)
	db.DB().SetMaxOpenConns(5)

	return &DB{db}
}
