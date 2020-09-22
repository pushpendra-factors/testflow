package main

import (
	C "factors/config"
	"factors/util"
	"flag"
	"fmt"
	"net/http"

	M "factors/model"

	IntSalesforce "factors/integration/salesforce"

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
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")
	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	errorReportingInterval := flag.Int("error_reporting_interval", 300, "")
	isRealTimeEventUserCachingEnabled := flag.Bool("enable_real_time_event_user_caching", false, "If the real time caching is enabled")
	realTimeEventUserCachingProjectIds := flag.String("real_time_event_user_caching_project_ids", "", "If the real time caching is enabled and the whitelisted projectids")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()

	if *env != "development" && *env != "staging" && *env != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *env))
	}

	taskID := "Task#SalesforceEnrich"
	defer util.NotifyOnPanic(taskID, *env)

	// init DB, etcd
	config := &C.Configuration{
		AppName: "salesforce_enrich_job",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHost:                          *redisHost,
		RedisPort:                          *redisPort,
		AWSKey:                             *awsAccessKeyId,
		AWSSecret:                          *awsSecretAccessKey,
		AWSRegion:                          *awsRegion,
		EmailSender:                        *factorsEmailSender,
		ErrorReportingInterval:             *errorReportingInterval,
		RedisHostPersistent:                *redisHostPersistent,
		RedisPortPersistent:                *redisPortPersistent,
		SentryDSN:                          *sentryDSN,
		IsRealTimeEventUserCachingEnabled:  *isRealTimeEventUserCachingEnabled,
		RealTimeEventUserCachingProjectIds: *realTimeEventUserCachingProjectIds,
	}

	C.InitConf(config.Env)
	C.InitEventUserRealTimeCachingConfig(config.IsRealTimeEventUserCachingEnabled, config.RealTimeEventUserCachingProjectIds)
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	defer C.SafeFlushSentryHook()

	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *dbHost, "port": *dbPort}).Fatal("Failed to initialize DB.")
	}
	db := C.GetServices().Db
	defer db.Close()

	salesforceEnabledProjects, status := M.GetAllSalesforceProjectSettings()
	if status != http.StatusFound {
		log.Fatal("No projects enabled salesforce integration.")
	}

	statusList := make([]IntSalesforce.Status, 0, 0)
	for _, settings := range salesforceEnabledProjects {
		status := IntSalesforce.Sync(settings.ProjectId)
		statusList = append(statusList, status...)
	}
}
