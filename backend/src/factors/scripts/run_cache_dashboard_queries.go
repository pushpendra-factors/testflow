package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	C "factors/config"
	M "factors/model"
	"factors/util"

	log "github.com/sirupsen/logrus"
)

func main() {
	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production")
	projectIDFlag := flag.String("project_id", "", "Comma separated project ids to run for. * to run for all")
	excludeProjectIDFlag := flag.String("exclude_project_id", "", "Comma separated project ids to exclude for the run")
	numRoutinesFlag := flag.Int("num_routines", 4, "Number of dashboard units to sync in parallel. Each dashboard unit runs 4 queries")
	numRoutinesForWebAnalyticsFlag := flag.Int("num_routines_for_web_analytics", 1,
		"No.of routines to use for web analytics dashboard caching.")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	onlyWebAnalytics := flag.Bool("only_web_analytics", false, "Cache only web analytics dashboards.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	flag.Parse()
	taskID := "dashboard_caching"
	healthcheckPingID := C.HealthcheckDashboardCachingPingID
	defer C.PingHealthcheckForPanic(taskID, *envFlag, healthcheckPingID)
	logCtx := log.WithFields(log.Fields{"Prefix": taskID})

	if *envFlag != C.DEVELOPMENT && *envFlag != C.STAGING && *envFlag != C.PRODUCTION {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == "" {
		panic(fmt.Errorf("Invalid project id %s", *projectIDFlag))
	} else if *numRoutinesFlag == 0 {
		panic(fmt.Errorf("Num routines must at least be 1"))
	}

	logCtx.Info("Starting to initialize database.")
	config := &C.Configuration{
		AppName:            taskID,
		Env:                *envFlag,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  taskID,
		},
		RedisHost: *redisHost,
		RedisPort: *redisPort,
		SentryDSN: *sentryDSN,
	}
	C.InitConf(config.Env)

	err := C.InitDB(config.DBInfo)
	if err != nil {
		logCtx.WithError(err).Panic("Failed to initialize DB")
	}
	C.InitRedisPersistent(config.RedisHost, config.RedisPort)

	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	logCtx = logCtx.WithFields(log.Fields{
		"Env":              *envFlag,
		"ProjectID":        *projectIDFlag,
		"ExcludeProjectID": *excludeProjectIDFlag,
		"NumRoutines":      *numRoutinesFlag,
	})

	var notifyMessage string
	var waitGroup sync.WaitGroup
	var timeTaken sync.Map

	if !*onlyWebAnalytics {
		waitGroup.Add(1)
		go cacheDashboardUnitsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesFlag, &timeTaken, &waitGroup)
	}

	waitGroup.Add(1)
	go cacheWebsiteAnalyticsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesForWebAnalyticsFlag, &timeTaken, &waitGroup)

	waitGroup.Wait()
	timeTakenString, _ := timeTaken.Load("all")
	timeTakenStringWeb, _ := timeTaken.Load("web")
	notifyMessage = fmt.Sprintf("Caching successful for %s - %s projects. Time taken: %+v. Time taken for web analytics: %+v",
		*projectIDFlag, *excludeProjectIDFlag, timeTakenString, timeTakenStringWeb)
	C.PingHealthcheckForSuccess(healthcheckPingID, notifyMessage)
}

func cacheDashboardUnitsForProjects(projectIDs, excludeProjectIDs string, numRoutines int, timeTaken *sync.Map, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	startTime := util.TimeNowUnix()
	M.CacheDashboardUnitsForProjects(projectIDs, excludeProjectIDs, numRoutines)
	timeTakenString := util.SecondsToHMSString(util.TimeNowUnix() - startTime)
	timeTaken.Store("all", timeTakenString)
}

func cacheWebsiteAnalyticsForProjects(projectIDs, excludeProjectIDs string, numRoutines int, timeTaken *sync.Map, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	startTime := util.TimeNowUnix()
	M.CacheWebsiteAnalyticsForProjects(projectIDs, excludeProjectIDs, numRoutines)
	timeTakenStringWeb := util.SecondsToHMSString(util.TimeNowUnix() - startTime)
	timeTaken.Store("web", timeTakenStringWeb)
}
