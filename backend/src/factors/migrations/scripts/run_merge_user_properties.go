package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm/dialects/postgres"

	C "factors/config"
	"factors/model/store"
	"factors/util"

	log "github.com/sirupsen/logrus"
)

func main() {
	envFlag := flag.String("env", "development", "Environment. Could be development|staging|production")
	projectIDFlag := flag.Uint64("project_id", 0, "Project id to be run for")
	userIDFlag := flag.String("user_id", "", "If required to be run for a particular customer_user_id")
	dryRunFlag := flag.Bool("dryrun", true, "Whether to run in dry run mode. Won't make database changes.")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	flag.Parse()
	defer util.NotifyOnPanic("Script#MergeUserProperties", *envFlag)
	logCtx := log.WithFields(log.Fields{"Prefix": "Script#MergeUserProperties"})

	if *envFlag != C.DEVELOPMENT && *envFlag != C.STAGING && *envFlag != C.PRODUCTION {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == 0 {
		panic(fmt.Errorf("Invalid project id %d", *projectIDFlag))
	}

	logCtx.Info("Starting to initialize database.")
	config := &C.Configuration{
		AppName: "script_push_to_bigquery",
		Env:     *envFlag,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}
	C.InitConf(config.Env)

	err := C.InitDB(config.DBInfo)
	if err != nil {
		logCtx.WithError(err).Fatal("Failed to initialize DB")
	}

	logCtx.WithFields(log.Fields{
		"Env":       *envFlag,
		"ProjectID": *projectIDFlag,
		"UserID":    *userIDFlag,
		"Dryrun":    *dryRunFlag,
	}).Infof("Starting merge job")

	var errCode int
	if *userIDFlag != "" && *projectIDFlag != 0 {
		_, errCode = store.GetStore().MergeUserPropertiesForUserID(*projectIDFlag,
			*userIDFlag, postgres.Jsonb{}, "", util.TimeNowUnix(), *dryRunFlag, true)
	} else if *projectIDFlag != 0 {
		errCode = store.GetStore().MergeUserPropertiesForProjectID(*projectIDFlag, *dryRunFlag)
	}

	if errCode == http.StatusNotModified {
		logCtx.Info("User properties not modified")
	} else if errCode == http.StatusCreated {
		logCtx.Info("User properties merge successful")
	} else {
		logCtx.Error("Error while merging user properties")
	}
}
