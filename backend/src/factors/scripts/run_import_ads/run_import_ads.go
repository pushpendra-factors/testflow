package main

import (
	C "factors/config"
	"flag"

	"factors/model/store"

	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"

	T "factors/task"

	taskWrapper "factors/task/task_wrapper"

	"factors/filestore"

	log "github.com/sirupsen/logrus"
)

func main() {

	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production.")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	lookback := flag.Int("lookback", 1, "lookback for job")
	appName := "import_ads"
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")

	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "--bucket_name=/usr/local/var/factors/cloud_storage pass bucket name")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory.")

	flag.Parse()
	defaultHealthcheckPingID := C.HealthcheckLeadSquaredPullEventsPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	defer C.PingHealthcheckForPanic(appName, *envFlag, healthcheckPingID)

	config := &C.Configuration{
		Env: *envFlag,
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

	err = C.InitDBWithMaxIdleAndMaxOpenConn(*config, 200, 100)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *envFlag,
			"host": *memSQLHost, "port": *memSQLPort}).Panic("Failed to initialize DB.")
	}
	db := C.GetServices().Db
	defer db.Close()

	projectIdsArray := make([]int64, 0)
	mappings, err := store.GetStore().GetAllAdsImportEnabledProjects()
	if err != nil {
		C.PingHealthcheckForFailure(healthcheckPingID, "Failed to get Ads Import Enabled Projects")
	}
	for id, _ := range mappings {
		projectIdsArray = append(projectIdsArray, id)
	}

	// Init cloud manager.
	var cloudManager filestore.FileManager
	if *envFlag == "development" {
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
	status := taskWrapper.TaskFuncWithProjectId("AdsImport", *lookback, projectIdsArray, T.AdsImport, configs)
	log.Info(status)
	C.PingHealthcheckForSuccess(healthcheckPingID, status)
}
