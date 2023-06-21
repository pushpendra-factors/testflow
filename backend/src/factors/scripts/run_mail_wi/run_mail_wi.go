package main

import (
	C "factors/config"
	T "factors/task"
	taskWrapper "factors/task/task_wrapper"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"

	"factors/model/store"

	D "factors/delta"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "--bucket_name=/usr/local/var/factors/cloud_storage pass bucket name")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	isWeeklyEnabled := flag.Bool("weekly_enabled", false, "")

	projectIdFlag := flag.String("project_id", "", "Comma separated list of project ids to run")
	lookback := flag.Int("lookback", 11, "lookback_for_delta lookup")
	enableDryRunAlerts := flag.Bool("dry_run_alerts", false, "")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")
	projectsFromDB := flag.Bool("projects_from_db", false, "")
	wetRun := flag.Bool("wet_run", false, "")
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
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:   *primaryDatastore,
		EmailSender:        *factorsEmailSender,
		EnableDryRunAlerts: *enableDryRunAlerts,
		AWSKey:             *awsAccessKeyId,
		AWSSecret:          *awsSecretAccessKey,
		AWSRegion:          *awsRegion,
	}
	defaultHealthcheckPingID := C.HealthcheckMailWIPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	C.InitConf(config)
	C.InitSenderEmail(C.GetFactorsSenderEmail())
	C.InitMailClient(config.AWSKey, config.AWSSecret, config.AWSRegion)
	C.InitFilemanager(*bucketNameFlag, *env, config)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()
	//Initialized configs

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

	if *isWeeklyEnabled {
		configs["modelType"] = T.ModelTypeWeek
		configs["wetRun"] = *wetRun
		status := taskWrapper.TaskFuncWithProjectId("WIMailInsights", *lookback, projectIdsArray, D.MailWeeklyInsights, configs)
		log.Info(status)
		if status["err"] != nil {
			C.PingHealthcheckForFailure(healthcheckPingID, status)
		}
		C.PingHealthcheckForSuccess(healthcheckPingID, status)
	}
}
