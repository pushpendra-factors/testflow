package main

import (
	"encoding/json"
	"flag"
	"fmt"

	U "factors/util"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var memSQLDB *gorm.DB

func initMemSQLDB(env, dsn string) {
	var err error
	// dsn sample admin:LpAHQyAMyI@tcp(svc-2b9e36ee-d5d0-4082-9779-2027e39fcbab-ddl.gcp-virginia-1.db.memsql.com:3306)/factors?charset=utf8mb4&parseTime=True&loc=Local
	memSQLDB, err = gorm.Open("mysql", dsn)
	if err != nil {
		log.WithError(err).Fatal("Failed connecting to memsql.")
	}

	memSQLDB.Callback().Create().Before("gorm:create").Register("cleanup", U.GormCleanupCallback)

	if env == "development" {
		memSQLDB.LogMode(true)
	} else {
		memSQLDB.LogMode(false)
	}
}

type TestCleanup struct {
	ID         string
	Title      string
	Properties postgres.Jsonb
}

func main() {
	memSQLDSN := flag.String(
		"memsql_dsn",
		"root:dbfactors123@tcp(localhost:3306)/factors?charset=utf8mb4&parseTime=True&loc=Local",
		"",
	)
	flag.Parse()
	initMemSQLDB("development", *memSQLDSN)

	input := "This is a string 😄 🐷 with some 👍🏻 🙈 emoji! 🐷 🏃🏿‍♂️"
	output := U.SanitizeStringValueForUnicode(input)
	fmt.Println(output)

	if err := memSQLDB.CreateTable(&TestCleanup{}).Error; err != nil {
		log.WithError(err).Error("TestCleanup table creation failed.")
	}

	id := U.RandomLowerAphaNumString(10)
	m := map[string]string{"emoji_data": "This is a string 😄 🐷 with some 👍🏻 🙈 emoji! 🐷 🏃🏿‍♂️", "single_quote_data": "abc'124"}
	propertiesJSON, err := json.Marshal(m)
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}

	properties := postgres.Jsonb{RawMessage: propertiesJSON}
	title := "Title'123"
	err = memSQLDB.Create(&TestCleanup{ID: id, Title: title, Properties: properties}).Error
	if err != nil {
		log.WithError(err).Fatal("Failed to create.")
	}
}
