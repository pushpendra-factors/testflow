package main

// Pull events that needs to be processed and write to file.
// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_pull_events.go --project_id=1 --output_dir="" --end_time=""

import (
	C "factors/config"
	"factors/filestore"
	M "factors/model/model"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	taskWrapper "factors/task/task_wrapper"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	archiveBucketNameFlag := flag.String("archive_bucket_name", "/usr/local/var/factors/cloud_storage_archive", "--bucket_name=/usr/local/var/factors/cloud_storage_archive pass archive bucket name")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	hardPull := flag.Bool("hard_pull", false, "replace the files already present")
	pullEventsDaily := flag.Bool("pull_events_daily", false, "run PullEventsDaily as well")
	sortOnTimestamp := flag.Bool("sort_on_timestamp", false, "whether to sort in db (for memory spike cases)")

	fileTypesFlag := flag.String("file_types", "*",
		"Optional: file type. A comma separated list of file types and supports '*' for all files. ex: 1,2,6,9") //refer to pull.FileType map
	projectIdFlag := flag.String("project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")
	eventSplitRangeProjectIdFlag := flag.String("event_split_project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids where query range is spli nto multiple parts and supports '*' for all projects. ex: 1,2,6,9")
	userSplitRangeProjectIdFlag := flag.String("user_split_project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids where query range is spli nto multiple parts and supports '*' for all projects. ex: 1,2,6,9")
	noOfSplits := flag.Int("number_splits", 1, "number of parts to split the range into for db query")
	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")
	projectsFromDB := flag.Bool("projects_from_db", false, "")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defer U.NotifyOnPanic("Task#PullEvents", *env)

	appName := "pull_events_job"
	healthcheckPingID := C.HealthcheckPullEventsPingID
	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:    *primaryDatastore,
		SentryDSN:           *sentryDSN,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to pull data. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	fileTypesList := strings.TrimSpace(*fileTypesFlag)
	var fileTypes []int64
	if fileTypesList == "*" {
		fileTypes = []int64{1, 2, 3, 4, 5, 6, 7}
	} else {
		fileTypes = C.GetTokensFromStringListAsUint64(fileTypesList)
	}
	fileTypesMap := make(map[int64]bool)
	for _, i := range fileTypes {
		fileTypesMap[i] = true
	}

	eventsProjectIdsArray := make([]int64, 0)
	allDataProjectIdsArray := make([]int64, 0)
	{
		projectIdsToRun := make(map[int64]bool)
		if *projectsFromDB {
			if path_analysis_projects, err := store.GetStore().GetAllProjectsWithFeatureEnabled(M.FEATURE_PATH_ANALYSIS, false); err != nil {
				log.WithError(err).Fatal("failed to get path analysis enabled projects")
			} else {
				for _, id := range path_analysis_projects {
					projectIdsToRun[id] = false
				}
			}
			if explain_projects, err := store.GetStore().GetAllProjectsWithFeatureEnabled(M.FEATURE_EXPLAIN, false); err != nil {
				log.WithError(err).Fatal("failed to get explain enabled projects")
			} else {
				for _, id := range explain_projects {
					projectIdsToRun[id] = false
				}
			}
			if acc_scoring_projects, err := store.GetStore().GetAllProjectsWithFeatureEnabled(M.FEATURE_ACCOUNT_SCORING, false); err != nil {
				log.WithError(err).Fatal("failed to get account scoring enabled projects")
			} else {
				for _, id := range acc_scoring_projects {
					projectIdsToRun[id] = false
				}
			}
			if wi_projects, err := store.GetStore().GetAllProjectsWithFeatureEnabled(M.FEATURE_WEEKLY_INSIGHTS, false); err != nil {
				log.WithError(err).Fatal("failed to get weekly insights enabled projects")
			} else {
				for _, id := range wi_projects {
					projectIdsToRun[id] = true
				}
			}
		}
		{
			allProjects, projectIdsFromList, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIdFlag, "")
			if allProjects {
				projectIDs, errCode := store.GetStore().GetAllProjectIDs()
				if errCode != http.StatusFound {
					log.Fatal("Failed to get all projects and project_ids set to '*'.")
				}
				for _, projectID := range projectIDs {
					projectIdsFromList[projectID] = true
				}
			}
			for projectId := range projectIdsFromList {
				projectIdsToRun[projectId] = true
			}
		}
		for projectId, yesData := range projectIdsToRun {
			eventsProjectIdsArray = append(eventsProjectIdsArray, projectId)
			if yesData {
				allDataProjectIdsArray = append(allDataProjectIdsArray, projectId)
			}
		}
	}

	eventSplitRangeProjectIds := make([]int64, 0)
	{
		allProjects, splitRangeProjectIdsMap, _ := C.GetProjectsFromListWithAllProjectSupport(*eventSplitRangeProjectIdFlag, "")
		if allProjects {
			for _, projectID := range eventsProjectIdsArray {
				splitRangeProjectIdsMap[projectID] = true
			}
		}

		for projectId := range splitRangeProjectIdsMap {
			eventSplitRangeProjectIds = append(eventSplitRangeProjectIds, projectId)
		}
	}

	userSplitRangeProjectIds := make([]int64, 0)
	{
		allProjects, splitRangeProjectIdsMap, _ := C.GetProjectsFromListWithAllProjectSupport(*userSplitRangeProjectIdFlag, "")
		if allProjects {
			for _, projectID := range eventsProjectIdsArray {
				splitRangeProjectIdsMap[projectID] = true
			}
		}

		for projectId := range splitRangeProjectIdsMap {
			userSplitRangeProjectIds = append(userSplitRangeProjectIds, projectId)
		}
	}

	configs := make(map[string]interface{})
	// Init cloud manager.
	var archiveCloudManager filestore.FileManager
	if *env == "development" {
		archiveCloudManager = serviceDisk.New(*archiveBucketNameFlag)
	} else {
		archiveCloudManager, err = serviceGCS.New(*archiveBucketNameFlag)
		if err != nil {
			log.WithField("error", err).Fatal("Failed to init archive cloud manager")
		}
	}
	configs["cloudManager"] = &archiveCloudManager

	configs["hardPull"] = hardPull
	configs["eventSplitRangeProjectIds"] = eventSplitRangeProjectIds
	configs["userSplitRangeProjectIds"] = userSplitRangeProjectIds
	configs["noOfSplits"] = *noOfSplits
	configs["sortOnTimestamp"] = *sortOnTimestamp

	var statusEvents map[string]interface{}
	if *pullEventsDaily {
		fileTypesMapOnlyEvents := make(map[int64]bool)
		fileTypesMapOnlyEvents[1] = true
		configs["fileTypes"] = fileTypesMapOnlyEvents
		C.PingHealthcheckForStart(healthcheckPingID)
		statusEvents = taskWrapper.TaskFuncWithProjectId("PullEventsDaily", *lookback, eventsProjectIdsArray, T.PullAllDataV2, configs)
		log.Info("PullEventsDaily: ", statusEvents)
		C.PingHealthCheckBasedOnStatus(statusEvents, healthcheckPingID)
	}

	if len(fileTypesMap) != 0 {
		configs["fileTypes"] = fileTypesMap
		C.PingHealthcheckForStart(healthcheckPingID)
		status := taskWrapper.TaskFuncWithProjectId("PullDataDaily", *lookback, allDataProjectIdsArray, T.PullAllDataV2, configs)
		log.Info("PullDataDaily: ", status)
		log.Info("PullEventsDaily: ", statusEvents)
		C.PingHealthCheckBasedOnStatus(status, healthcheckPingID)
	}
}
