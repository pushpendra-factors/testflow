package main

import (
	C "factors/config"
	"factors/metrics"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"time"

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
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	isRealTimeEventUserCachingEnabled := flag.Bool("enable_real_time_event_user_caching", false, "If the real time caching is enabled")
	realTimeEventUserCachingProjectIds := flag.String("real_time_event_user_caching_project_ids", "",
		"If the real time caching is enabled and the whitelisted projectids")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	flag.Parse()

	if *env != "development" && *env != "staging" && *env != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *env))
	}

	taskID := "Task#HubspotEnrich"
	defer util.NotifyOnPanic(taskID, *env)

	// init DB, etcd
	config := &C.Configuration{
		AppName:            "hubspot_enrich_job",
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHost:                          *redisHost,
		RedisPort:                          *redisPort,
		RedisHostPersistent:                *redisHostPersistent,
		RedisPortPersistent:                *redisPortPersistent,
		SentryDSN:                          *sentryDSN,
		IsRealTimeEventUserCachingEnabled:  *isRealTimeEventUserCachingEnabled,
		RealTimeEventUserCachingProjectIds: *realTimeEventUserCachingProjectIds,
	}

	C.InitConf(config.Env)
	C.InitEventUserRealTimeCachingConfig(config.IsRealTimeEventUserCachingEnabled, config.RealTimeEventUserCachingProjectIds)

	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *dbHost, "port": *dbPort}).Fatal("Failed to initialize DB.")
	}
	db := C.GetServices().Db
	defer db.Close()

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

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
	metrics.Increment(metrics.IncrCronHubspotEnrichSuccess)
}
