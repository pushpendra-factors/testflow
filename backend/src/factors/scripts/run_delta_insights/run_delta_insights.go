package main

// Sample usage in terminal.
// export GOPATH=~/factors/backend/
// go run run_delta_insights.go --project_id=<projectId> --model_id1=<modelId1> --model_id2=<modelId2> --env=development --local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp --bucket_name=<bucketName>

import (
	"factors/filestore"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	"factors/util"
	"flag"
	"fmt"
	"net/http"

	C "factors/config"

	T "factors/task"

	taskWrapper "factors/task/task_wrapper"

	D "factors/delta"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {
	mainRunDeltaInsights()
}

func mainRunDeltaInsights() {

	projectIdFlag := flag.String("project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")

	envFlag := flag.String("env", "development", "")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory")
	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "")
	kValue := flag.Int("k", -1, "--k=10")
	whitelistedDashboardIds := flag.String("whitelisted_dashboard_ids", "*", "")
	skipWpi := flag.Bool("skip_wpi", false, "")
	skipWpi2 := flag.Bool("skip_wpi2", false, "")
	runKpi := flag.Bool("run_kpi", false, "")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	lookbackWindowForEventUserCache := flag.Int("lookback_window_event_user_cache",
		30, "look back window in cache for event/user cache(to get props for custom kpis")

	isWeeklyEnabled := flag.Bool("weekly_enabled", false, "")
	isMailerRun := flag.Bool("is_mailer_run", false, "")
	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")
	projectsFromDB := flag.Bool("projects_from_db", false, "")
	flag.Parse()

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	defer util.NotifyOnPanic("Task#ComputeStreamingMetrics", *envFlag)

	// init Config and DB.
	config := &C.Configuration{
		AppName: "compute_delta_insights",
		Env:     *envFlag,
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
			AppName:     "compute_delta_insights",
		},
		PrimaryDatastore:                *primaryDatastore,
		RedisHost:                       *redisHost,
		RedisPort:                       *redisPort,
		RedisHostPersistent:             *redisHostPersistent,
		RedisPortPersistent:             *redisPortPersistent,
		LookbackWindowForEventUserCache: *lookbackWindowForEventUserCache,
	}

	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	log.WithFields(log.Fields{
		"Env":             *envFlag,
		"localDiskTmpDir": *localDiskTmpDirFlag,
		"Bucket":          *bucketName,
	}).Infoln("Initialising")

	var cloudManager filestore.FileManager
	if *envFlag == "development" {
		cloudManager = serviceDisk.New(*bucketName)
	} else {
		log.Info("initializing cloud bucket")
		cloudManager, err = serviceGCS.New(*bucketName)
		if err != nil {
			log.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
	}
	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	projectIdsToRun := make(map[int64]bool, 0)
	if *projectsFromDB {
		wi_projects, _ := store.GetStore().GetAllWeeklyInsightsEnabledProjects()
		for _, id := range wi_projects {
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

	configs := make(map[string]interface{})
	configs["diskManager"] = diskManager
	configs["cloudManager"] = &cloudManager

	allDashboard, allDashboards, _ := C.GetProjectsFromListWithAllProjectSupport(*whitelistedDashboardIds, "")
	whitelistedIds := make(map[string]bool)
	if allDashboard {
		whitelistedIds["*"] = true
	} else {
		for id, _ := range allDashboards {
			whitelistedIds[fmt.Sprintf("%v", id)] = true
		}
	}
	configs["whitelistedDashboardUnits"] = whitelistedIds
	var k int = *kValue // Selecting all top features if k = -1.
	configs["k"] = k
	configs["skipWpi"] = (*skipWpi)
	configs["skipWpi2"] = (*skipWpi2)
	configs["runKpi"] = (*runKpi)
	if *isWeeklyEnabled && !(*isMailerRun) {
		configs["insightGranularity"] = T.ModelTypeWeek
		status := taskWrapper.TaskFuncWithProjectId("WIWeekly", *lookback, projectIdsArray, D.ComputeDeltaInsights, configs)
		log.Info(status)
	}
	if *isWeeklyEnabled && *isMailerRun {
		configs["insightGranularity"] = T.ModelTypeWeek
		configs["run_type"] = "mailer"
		status := taskWrapper.TaskFuncWithProjectId("WIWeeklyMailer", *lookback, projectIdsArray, D.ComputeDeltaInsights, configs)
		log.Info(status)
	}
}
