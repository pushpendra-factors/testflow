package main

// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_compute_streaming_metrics.go --env=development --disk_dir=/usr/local/var/factors/local_disk --s3_region=us-east-1 --s3=/usr/local/var/factors/cloud_storage --project_id=<projectId> --model_id=<modelId>

// go run run_compute_streaming_metrics.go --project_id=<projectId> --model_id=<modelId>
// Also would take default flag values to connect with db similar to run_cache_dashboard_queries.go

import (
	C "factors/config"
	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	"factors/util"
	"flag"
	"fmt"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {

	// Define the necessary flags as in run_pattern_mine.go

	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")
	modelIdFlag := flag.Uint64("model_id", 0, "Model Id")

	envFlag := flag.String("env", "development", "")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory")
	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	flag.Parse()

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	defer util.NotifyOnPanic("Task#ComputeStreamingMetrics", *envFlag)

	// init Config and DB.
	appName := "compute_streaming_metrics"
	config := &C.Configuration{
		AppName: appName,
		Env:     *envFlag,
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
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}
	db := C.GetServices().Db
	defer db.Close()
	// Connect to cloud storage to fetch the required flag to local disk.
	var cloudManager filestore.FileManager
	if *envFlag == "development" {
		cloudManager = serviceDisk.New(*bucketName)
	} else {
		cloudManager, err = serviceGCS.New(*bucketName)
		if err != nil {
			log.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
	}
	diskManager := serviceDisk.New(*localDiskTmpDirFlag)
	// Initialize other services base on requirements.

	// Add checks for flags and input variables.
	if *projectIdFlag <= 0 || *modelIdFlag <= 0 {
		log.Fatal("project_id and model_id are required.")
	}

	// Call the task. Task can be managed from a task manager in a different script.
	err = T.ComputeStreamingMetrics(db, &cloudManager, diskManager, *bucketName, int64(*projectIdFlag), *modelIdFlag)
	if err != nil {
		log.WithError(err).Fatal("Compute Streaming Metrics failed")
	}
}
