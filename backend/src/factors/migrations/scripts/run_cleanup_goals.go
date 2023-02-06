package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/util"
	U "factors/util"
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func doGoalCleanUp(projectID uint64) int64 {
	db := C.GetServices().Db
	dbObj := db.Where("type = ?", "AT").Where("project_id = ?", projectID).Delete(&model.FactorsGoal{})
	if dbObj.Error != nil {
		log.WithFields(log.Fields{"projectId": projectID}).WithError(db.Error).Error(
			"Deleting from Goal Table failed")
	}
	log.WithField("ProjectId", projectID).WithField("Count", dbObj.RowsAffected).Info("Goals Deleted Count")
	return dbObj.RowsAffected
}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func doTrackedEventsCleanUp(projectID uint64) int64 {
	db := C.GetServices().Db
	var trackedEvent model.FactorsTrackedEvent
	dbObj := db.Where("type = ?", "AT").Where("project_id = ?", projectID).Delete(&trackedEvent)
	if dbObj.Error != nil {
		log.WithFields(log.Fields{"projectId": projectID}).WithError(db.Error).Error(
			"Deleting from TrackedEvents Table failed")
	}
	log.WithField("ProjectId", projectID).WithField("Count", dbObj.RowsAffected).Info("TrackedEvents Deleted Count")
	return dbObj.RowsAffected
}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func doTrackedUserPropertiesCleanUp(projectID uint64) int64 {
	db := C.GetServices().Db
	dbObj := db.Where("type = ?", "AT").Where("project_id = ?", projectID).Delete(&model.FactorsTrackedUserProperty{})
	if dbObj.Error != nil {
		log.WithFields(log.Fields{"projectId": projectID}).WithError(db.Error).Error(
			"Deleting from TrackedUserProperties Table failed")
	}
	log.WithField("ProjectId", projectID).WithField("Count", dbObj.RowsAffected).Info("TrackedUserProperties Deleted Count")
	return dbObj.RowsAffected
}

func main() {

	env := flag.String("env", C.DEVELOPMENT, "")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

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

	appName := "CleanUpGoal"
	config := &C.Configuration{
		AppName:            appName,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
			AppName:  appName,
		},
		PrimaryDatastore: *primaryDatastore,
		SentryDSN:        *sentryDSN,
	}

	C.InitConf(config)
	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 50, 50)
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
		trackedEventsDeleted += doTrackedEventsCleanUp(projectId)
		trackedUserPropertiesDeleted += doTrackedUserPropertiesCleanUp(projectId)
		goalsDeleted += doGoalCleanUp(projectId)
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
