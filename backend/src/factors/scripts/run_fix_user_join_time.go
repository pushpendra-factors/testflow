package main

// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_fix_user_join_time.go --project_id=1
// Fix join timestamps of users. Make it lesser than the first event seen for the user.

import (
	C "factors/config"
	"factors/util"
	"flag"
	"fmt"

	"github.com/jinzhu/gorm"

	log "github.com/sirupsen/logrus"
)

func fixUserJoinTimestamp(db *gorm.DB, projectId uint64, isDryRun bool) error {

	userRows, err := db.Raw("SELECT id, join_timestamp FROM users WHERE project_id = ?", projectId).Rows()
	defer userRows.Close()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("SQL Query failed.")
		return err
	}
	for userRows.Next() {
		var userId string
		var joinTimestamp int64
		if err = userRows.Scan(&userId, &joinTimestamp); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return err
		}
		type Result struct {
			Timestamp int64
		}
		var result Result
		db.Raw("SELECT MIN(timestamp) as Timestamp FROM events WHERE user_id = ? AND project_id = ?", userId, projectId).Scan(&result)
		if result.Timestamp > 0 && result.Timestamp < joinTimestamp {
			newJoinTimestamp := result.Timestamp - 60
			log.WithFields(log.Fields{
				"userId":            userId,
				"userJoinTimestamp": joinTimestamp,
				"minEventTimestamp": result.Timestamp,
				"newJoinTimestamp":  newJoinTimestamp,
			}).Error("Need to update.")
			if !isDryRun {
				db.Exec("UPDATE users SET join_timestamp=? WHERE project_id=? AND id=?", newJoinTimestamp, projectId, userId)
				log.Info(fmt.Sprintf("Updated %s", userId))
			}
		}
	}
	return nil
}

func main() {
	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")
	dryRunFlag := flag.Bool("dry_run", true, "values are updated only when dry_run is false")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defer util.NotifyOnPanic("Task#FixUserJoinTime", *env)

	config := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}

	C.InitConf(config.Env)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	if *projectIdFlag <= 0 {
		log.Fatal("Failed to pull events. Invalid project_id.")
	}
	if err := fixUserJoinTimestamp(db, *projectIdFlag, *dryRunFlag); err != nil {
		log.Fatal(err)
	}
}
