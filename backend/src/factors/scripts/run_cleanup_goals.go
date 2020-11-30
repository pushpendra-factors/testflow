package main

import (
	C "factors/config"
	"factors/util"
	U "factors/util"
	"flag"
	"fmt"
	"time"

	cleanup "factors/task/goal_cleanup"

	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")
	projectIds := flag.String("project_ids", "", "Projects for which the cache is to be refreshed")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "Task#CleanUpGoal"
	defer U.NotifyOnPanic(taskID, *env)

	config := &C.Configuration{
		AppName:            "CleanUpGoal",
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		SentryDSN: *sentryDSN,
	}

	C.InitConf(config.Env)
	err := C.InitDBWithMaxIdleAndMaxOpenConn(config.DBInfo, 50, 50)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db in add session.")
	}
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	trackedEventsDeleted := int64(0)
	trackedUserPropertiesDeleted := int64(0)
	goalsDeleted := int64(0)
	projectIdMap := util.GetIntBoolMapFromStringList(projectIds)
	for projectId, _ := range projectIdMap {
		trackedEventsDeleted += cleanup.DoTrackedEventsCleanUp(projectId)
		trackedUserPropertiesDeleted += cleanup.DoTrackedUserPropertiesCleanUp(projectId)
		goalsDeleted += cleanup.DoGoalCleanUp(projectId)
	}

	status := map[string]interface{}{
		"no_of_tracked_events_deleted":          trackedEventsDeleted,
		"no_of_tracked_user_properties_deleted": trackedUserPropertiesDeleted,
		"no_of_goals_deleted":                   goalsDeleted,
	}
	if err := U.NotifyThroughSNS(taskID, *env, status); err != nil {
		log.Fatalf("Failed to notify status %+v", status)
	}
	log.Info("Done!!!")
}
