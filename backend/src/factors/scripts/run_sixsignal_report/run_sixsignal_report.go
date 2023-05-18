package main

import (
	C "factors/config"
	"factors/filestore"
	"factors/integration/six_signal"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	U "factors/util"
	"flag"
	"fmt"

	T "factors/task"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	modelBucketNameFlag := flag.String("model_bucket_name", "/usr/local/var/factors/cloud_storage_models", "--bucket_name=/usr/local/var/factors/cloud_storage_models pass models bucket name")

	hardPull := flag.Bool("hard_pull", false, "replace the files already present")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory.")
	sortOnGroup := flag.Int("sort_group", 0, "sort events based on group (0 for uid)")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	appDomain := flag.String("app_domain", "factors-dev.com:3000", "")
	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")
	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	//lookback := flag.Int("lookback", 1, "lookback_for_delta lookup")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	//projectIdFlag := flag.String("project_ids", "",
	//	"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")
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
		APPDomain:           *appDomain,
		AWSKey:              *awsAccessKeyId,
		AWSSecret:           *awsSecretAccessKey,
		AWSRegion:           *awsRegion,
		EmailSender:         *factorsEmailSender,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}
	defaultHealthcheckPingID := C.HealthCheckSixSignalReportPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	C.InitConf(config)
	C.InitSenderEmail(C.GetFactorsSenderEmail())
	C.InitMailClient(config.AWSKey, config.AWSSecret, config.AWSRegion)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Init failed.")
	}
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	db := C.GetServices().Db
	defer db.Close()

	projectIdsArray := store.GetStore().GetProjectIDsWithSixSignalEnabled()

	//Initialized configs
	configs := make(map[string]interface{})

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

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)
	configs["diskManager"] = diskManager
	configs["hardPull"] = *hardPull
	configs["sortOnGroup"] = *sortOnGroup

	//Higher level map to contain run data from both SixSignalAnalysis and SendSixSignalEmailForSubscribe
	jobReport := make(map[string]interface{})

	//Generating the sixsignal report
	SixSignalReportFailures := T.SixSignalAnalysis(projectIdsArray, configs)

	//Sending emails for weekly report generated for subscribed reports
	SendEmailReportFailures := six_signal.SendSixSignalEmailForSubscribe(projectIdsArray)

	jobReport["Six Signal Analysis Job Report"] = SixSignalReportFailures
	jobReport["Send Six Signal Email Report"] = SendEmailReportFailures

	if len(SixSignalReportFailures) > 0 || len(SendEmailReportFailures) > 0 {
		C.PingHealthcheckForFailure(healthcheckPingID, jobReport)
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, jobReport)
	}
}
