package main

import (
	C "factors/config"
	"factors/util"
	"flag"
	"fmt"
	"net/http"

	IntHubspot "factors/integration/hubspot"
	M "factors/model"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", "development", "")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")
	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	errorReportingInterval := flag.Int("error_reporting_interval", 300, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()

	if *env != "development" && *env != "staging" && *env != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *env))
	}

	taskID := "Task#HubspotEnrich"
	defer util.NotifyOnPanic(taskID, *env)

	// init DB, etcd
	config := &C.Configuration{
		AppName: "hubspot_enrich_job",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHost:              *redisHost,
		RedisPort:              *redisPort,
		AWSKey:                 *awsAccessKeyId,
		AWSSecret:              *awsSecretAccessKey,
		AWSRegion:              *awsRegion,
		EmailSender:            *factorsEmailSender,
		ErrorReportingInterval: *errorReportingInterval,
		SentryDSN:              *sentryDSN,
	}

	C.InitConf(config.Env)

	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *dbHost, "port": *dbPort}).Fatal("Failed to initialize DB.")
	}
	db := C.GetServices().Db
	defer db.Close()

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitLogClient(config.Env, config.AppName, config.EmailSender, config.AWSKey,
		config.AWSSecret, config.AWSRegion, config.ErrorReportingInterval, config.SentryDSN)
	C.GetServices().SentryHook.SetTagsContext(map[string]string{
		"JobName": taskID,
	})
	defer C.GetServices().SentryHook.Flush()

	hubspotEnabledProjectSettings, errCode := M.GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		log.Fatal("No projects enabled hubspot integration.")
	}

	statusList := make([]IntHubspot.Status, 0, 0)
	for _, settings := range hubspotEnabledProjectSettings {
		status := IntHubspot.Sync(settings.ProjectId)
		statusList = append(statusList, status...)
	}

	err = util.NotifyThroughSNS("hubspot_enrich", *env, statusList)
	if err != nil {
		log.WithError(err).Fatal("Failed to notify through SNS on hubspot sync.")
	}
}
