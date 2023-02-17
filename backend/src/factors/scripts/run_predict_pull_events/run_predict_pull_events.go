package main

import (
	C "factors/config"
	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	"flag"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "--bucket_name=/usr/local/var/factors/cloud_storage pass bucket name")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory.")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	projectToRun := flag.Int64("project_id", -1, "project id to run ")
	targetTime := flag.Int64("target_time", 0, "time stamp of target event")
	startTime := flag.Int64("start_time", 0, "time stamp of base event")
	targetEvent := flag.String("target_event", "", "target event")
	baseEvent := flag.String("base_event", "", "start event")
	targetProp := flag.String("target_prop", "", "target event property to use")
	bufferTime := flag.Int64("buffer_weeks", 0, "buffer time of base event in weeks")

	flag.Parse()

	appName := "predict_pull_events_job"
	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
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

	// init disk manager
	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	// init redis
	projectIdsArray := make([]int64, 0)

	configs := make(map[string]interface{})
	configs["diskManager"] = diskManager
	configs["cloudManager"] = &cloudManager

	// project
	configs["project_id"] = *projectToRun
	configs["target_event"] = *targetEvent
	configs["base_event"] = *baseEvent
	configs["target_prop"] = *targetProp
	configs["target_time"] = *targetTime
	configs["start_time"] = *startTime
	configs["buffer_time"] = *bufferTime

	projectIdsArray = append(projectIdsArray, *projectToRun)

	err = T.PredictPullData(configs)
	if err != nil {
		log.Panicf("unable to predict pull events job complete job :%v", err)
	}
}
