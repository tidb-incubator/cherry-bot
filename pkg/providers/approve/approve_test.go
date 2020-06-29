package approve

import (
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/config"
	"github.com/pingcap-incubator/cherry-bot/pkg/db"
	"github.com/pingcap-incubator/cherry-bot/pkg/operator"
)

func initDB() (Approve, LgtmRecord) {
	dbCfg := config.Database{
		Address:  "10.10.10.1",
		Port:     3306,
		Username: "root",
		Password: "",
		Dbname:   "test",
	}
	dbConnect := db.CreateDbConnect(&dbCfg)
	op := operator.Operator{
		DB: dbConnect,
	}
	app := Approve{
		owner:   "tikv",
		repo:    "tikv",
		ready:   true,
		approve: true,
		opr:     &op,
	}
	record := LgtmRecord{
		Owner:      "tikv",
		Repo:       "tikv",
		PullNumber: 8521,
		Github:     "AndreMouche",
		Score:      1,
	}
	return app, record
}
func TestOrm(t *testing.T) {
	app, record := initDB()
	fmt.Println(app.LGTMRecordExist(&record, app.opr.DB.Begin()))

}

//go test -run TestOrmUpdate
func TestOrmUpdate(t *testing.T) {
	app, record := initDB()
	fmt.Println(app.getLGTMNum(record.PullNumber))
	err := app.addLGTMRecord(record.Github, record.PullNumber, []*github.Label{})
	fmt.Println(err)
	// err = app.removeLGTMRecord(record.Github, record.PullNumber)
	// fmt.Println(err)
}

func TestOrmCancel(t *testing.T) {
	app, record := initDB()
	fmt.Println(app.getLGTMNum(record.PullNumber))

	err := app.removeLGTMRecord(record.Github, record.PullNumber)
	fmt.Println(err)
}
