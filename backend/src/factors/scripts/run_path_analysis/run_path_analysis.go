package main

import (
	C "factors/config"
	taskWrapper "factors/task/task_wrapper"
	U "factors/util"
	"flag"
	"fmt"

	D "factors/delta"

	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"

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

	lookback := flag.Int("lookback", 1, "lookback_for_delta lookup")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}
	defer U.NotifyOnPanic("Script#run_wi_alerts", *env)
	appName := "run_wi_mailer"
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
	defaultHealthcheckPingID := C.HealthcheckPathAnalysisPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()
	//Initialized configs

	projectIdsArray := make([]int64, 0)
	projectIdsArray = append(projectIdsArray, 51)
	// Get All the Projects for which the path analysis has pending items
	configs := make(map[string]interface{})
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

	configs["diskManager"] = diskManager
	configs["cloudManager"] = &cloudManager

	status := taskWrapper.TaskFuncWithProjectId("PathAnalysis", *lookback, projectIdsArray, D.PathAnalysis, configs)
	log.Info(status)
	if status["err"] != nil {
		C.PingHealthcheckForFailure(healthcheckPingID, status)
	}
	C.PingHealthcheckForSuccess(healthcheckPingID, status)

}
