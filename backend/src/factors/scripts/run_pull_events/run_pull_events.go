package main

// Pull events that needs to be processed and write to file.
// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_pull_events.go --project_id=1 --output_dir="" --end_time=""

import (
	C "factors/config"
	"factors/filestore"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"strings"

	taskWrapper "factors/task/task_wrapper"

	log "github.com/sirupsen/logrus"
)

const (
	DayInSecs   = 24 * 60 * 60
	MonthInSecs = 31 * DayInSecs
	WeekInSecs  = 7 * DayInSecs
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "--bucket_name=/usr/local/var/factors/cloud_storage pass bucket name")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory.")

	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	isWeeklyEnabled := flag.Bool("weekly_enabled", false, "")
	isMonthlyEnabled := flag.Bool("monthly_enabled", false, "")
	isQuarterlyEnabled := flag.Bool("quarterly_enabled", false, "")

	hardPull := flag.Bool("hard_pull", false, "replace the files already present")

	fileTypesFlag := flag.String("file_types", "*",
		"Optional: file type. A comma separated list of file types and supports '*' for all files. ex: 1,2,6,9") //refer to T.fileType map
	projectIdFlag := flag.String("project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")

	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")
	projectsFromDB := flag.Bool("projects_from_db", false, "")
	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defer util.NotifyOnPanic("Task#PullEvents", *env)

	appName := "pull_events_job"
	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)

	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	fileTypesList := strings.TrimSpace(*fileTypesFlag)
	var fileTypes []int64
	if fileTypesList == "*" {
		fileTypes = []int64{1, 2, 3, 4, 5, 6}
	} else {
		fileTypes = C.GetTokensFromStringListAsUint64(fileTypesList)
	}
	fileTypesMap := make(map[int64]bool)
	for i := range fileTypes {
		fileTypesMap[fileTypes[i]] = true
	}

	projectIdsToRun := make(map[int64]bool, 0)
	if *projectsFromDB {
		wi_projects, _ := store.GetStore().GetAllWeeklyInsightsEnabledProjects()
		explain_projects, _ := store.GetStore().GetAllExplainEnabledProjects()
		for _, id := range wi_projects {
			projectIdsToRun[id] = true
		}
		for _, id := range explain_projects {
			projectIdsToRun[id] = true
		}

	} else {
		var allProjects bool
		allProjects, projectIdsToRun, _ = C.GetProjectsFromListWithAllProjectSupport(*projectIdFlag, "")
		if allProjects {
			projectIDs, errCode := store.GetStore().GetAllProjectIDs()
			if errCode != http.StatusFound {
				log.Fatal("Failed to get all projects and project_ids set to '*'.")
			}
			for _, projectID := range projectIDs {
				projectIdsToRun[projectID] = true
			}
		}
	}

	projectIdsArray := make([]int64, 0)
	for projectId, _ := range projectIdsToRun {
		projectIdsArray = append(projectIdsArray, projectId)
	}
	// Init cloud manager.
	var cloudManager filestore.FileManager
	if *env == "development" {
		cloudManager = serviceDisk.New(*bucketNameFlag)
	} else {
		cloudManager, err = serviceGCS.New(*bucketNameFlag)
		if err != nil {
			log.WithField("error", err).Fatal("Failed to init cloud manager.")
		}
	}

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	configs := make(map[string]interface{})
	configs["diskManager"] = diskManager
	configs["cloudManager"] = &cloudManager
	configs["hardPull"] = hardPull

	fileTypesMapOnlyEvents := make(map[int64]bool)
	fileTypesMapOnlyEvents[1] = true
	if *isWeeklyEnabled {
		configs["fileTypes"] = fileTypesMapOnlyEvents
		configs["modelType"] = T.ModelTypeWeek
		status := taskWrapper.TaskFuncWithProjectId("PullEventsWeeklyOnlyEvents", *lookback, projectIdsArray, T.PullAllData, configs)
		log.Info(status)
	}

	if *isWeeklyEnabled {
		configs["fileTypes"] = fileTypesMap
		configs["modelType"] = T.ModelTypeWeek
		status := taskWrapper.TaskFuncWithProjectId("PullEventsWeekly", *lookback, projectIdsArray, T.PullAllData, configs)
		log.Info(status)
	}

	if *isMonthlyEnabled {
		configs["fileTypes"] = fileTypesMapOnlyEvents
		configs["modelType"] = T.ModelTypeMonth
		status := taskWrapper.TaskFuncWithProjectId("PullEventsMonthly", *lookback, projectIdsArray, T.PullAllData, configs)
		log.Info(status)
	}

	if *isQuarterlyEnabled {
		configs["fileTypes"] = fileTypesMapOnlyEvents
		configs["modelType"] = T.ModelTypeQuarter
		status := taskWrapper.TaskFuncWithProjectId("PullEventsQuarterly", *lookback, projectIdsArray, T.PullAllData, configs)
		log.Info(status)
	}
}
