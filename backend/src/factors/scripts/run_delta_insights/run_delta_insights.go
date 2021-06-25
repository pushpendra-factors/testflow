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

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	isWeeklyEnabled := flag.Bool("weekly_enabled", false, "")
	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")
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
	}
	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}

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

	allProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIdFlag, "")
	if allProjects {
		projectIDs, errCode := store.GetStore().GetAllProjectIDs()
		if errCode != http.StatusFound {
			log.Fatal("Failed to get all projects and project_ids set to '*'.")
		}

		projectIdsToRun = make(map[uint64]bool, 0)
		for _, projectID := range projectIDs {
			projectIdsToRun[projectID] = true
		}
	}

	projectIdsArray := make([]uint64, 0)
	for projectId, _ := range projectIdsToRun {
		projectIdsArray = append(projectIdsArray, projectId)
	}

	configs := make(map[string]interface{})
	configs["diskManager"] = diskManager
	configs["cloudManager"] = &cloudManager
	var k int = *kValue // Selecting all top features if k = -1.
	configs["k"] = k

	if *isWeeklyEnabled {
		configs["insightGranularity"] = T.ModelTypeWeek
		status := taskWrapper.TaskFuncWithProjectId("WIWeekly", *lookback, projectIdsArray, D.ComputeDeltaInsights, configs)
		log.Info(status)
	}
}
