package main

import (
	C "factors/config"
	"factors/delta"

	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	U "factors/util"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "--bucket_name=/usr/local/var/factors/cloud_storage pass bucket name")
	modelBucketNameFlag := flag.String("model_bucket_name", "/usr/local/var/factors/cloud_storage_models", "--bucket_name=/usr/local/var/factors/cloud_storage_models pass models bucket name")
	useBucketV2 := flag.Bool("use_bucket_v2", false, "Whether to use new bucketing system or not")
	hardPull := flag.Bool("hard_pull", false, "replace the files already present")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory.")
	sortOnGroup := flag.Int("sort_group", 0, "sort events based on group (0 for uid)")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	//lookback := flag.Int("lookback", 1, "lookback_for_delta lookup")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	projectIdFlag := flag.String("project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")
	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defer U.NotifyOnPanic("Script#six_signal_report", *env)
	appName := "six_signal_report"
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
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}
	defaultHealthcheckPingID := C.HealthCheckSixSignalReportPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Init failed.")
	}
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	db := C.GetServices().Db
	defer db.Close()
	//Initialized configs
	_, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIdFlag, "")
	projectIdsArray := make([]int64, 0)
	for projectId := range projectIdsToRun {
		projectIdsArray = append(projectIdsArray, projectId)
	}
	// Get All the Projects for which the path analysis has pending items
	configs := make(map[string]interface{})

	if *useBucketV2 {
		var modelCloudManager filestore.FileManager
		if *env == "development" {
			modelCloudManager = serviceDisk.New(*modelBucketNameFlag)
		} else {
			modelCloudManager, err = serviceGCS.New(*modelBucketNameFlag)
			if err != nil {
				log.WithField("error", err).Fatal("Failed to init cloud manager.")
			}
		}
		configs["modelCloudManager"] = &modelCloudManager
	} else {
		var cloudManager filestore.FileManager
		if *env == "development" {
			cloudManager = serviceDisk.New(*bucketNameFlag)
		} else {
			cloudManager, err = serviceGCS.New(*bucketNameFlag)
			if err != nil {
				log.WithField("error", err).Fatal("Failed to init cloud manager.")
			}
		}
		configs["modelCloudManager"] = &cloudManager
	}
	diskManager := serviceDisk.New(*localDiskTmpDirFlag)
	configs["diskManager"] = diskManager
	configs["useBucketV2"] = *useBucketV2
	configs["hardPull"] = *hardPull
	configs["sortOnGroup"] = *sortOnGroup

	log.Info("Hitting the method SixSignalAnalysis")
	_, status := delta.SixSignalAnalysis(projectIdsArray, configs)
	if !status {
		log.Info("Six Signal Analysis status failed")
		C.PingHealthcheckForFailure(healthcheckPingID, status)
	}
	log.Info("Six Signal Analysis status successful")
	C.PingHealthcheckForSuccess(healthcheckPingID, status)

}
