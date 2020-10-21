package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/util"

	"factors/task/session"
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

	// projectIds: supports * (asterisk) for all projects.
	projectIds := flag.String("project_ids", "", "Allowed projects to create sessions offline.")
	disabledProjectIds := flag.String("disabled_project_ids", "", "Disallowed projects to create sessions offline.")
	numRoutines := flag.Int("num_routines", 1, "Number of routines to use.")
	maxLookbackDays := flag.Int64("max_lookback_days", 0, "Max lookback days to look for session existence.")
	bufferTimeBeforeCreateSessionInMins := flag.Int64("buffer_time_in_mins", 30, "Buffer time to wait before processing an event for session.")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	isRealTimeEventUserCachingEnabled := flag.Bool("enable_real_time_event_user_caching", false, "If the real time caching is enabled")
	realTimeEventUserCachingProjectIds := flag.String("real_time_event_user_caching_project_ids", "", "If the real time caching is enabled and the whitelisted projectids")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "Task#AddSession"
	defer util.NotifyOnPanic(taskID, *env)

	config := &C.Configuration{
		AppName: "add_session",
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
		RedisHostPersistent:                *redisHostPersistent,
		RedisPortPersistent:                *redisPortPersistent,
		SentryDSN:                          *sentryDSN,
		IsRealTimeEventUserCachingEnabled:  *isRealTimeEventUserCachingEnabled,
		RealTimeEventUserCachingProjectIds: *realTimeEventUserCachingProjectIds,
	}

	C.InitConf(config.Env)
	C.InitEventUserRealTimeCachingConfig(config.IsRealTimeEventUserCachingEnabled, config.RealTimeEventUserCachingProjectIds)

	// Will allow all 50/50 connection to be idle on the pool.
	// As we allow num_routines (per project) as per no.of db connections
	// and will be used continiously.
	err := C.InitDBWithMaxIdleAndMaxOpenConn(config.DBInfo, 50, 50)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db in add session.")
	}

	// Cache dependency for requests not using queue.
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	defer C.SafeFlushAllCollectors()

	allowedProjectIds, errCode := session.GetAddSessionAllowedProjects(*projectIds, *disabledProjectIds)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("Failed to get add session allowed project ids.")
		os.Exit(0)
	}

	var maxLookbackTimestamp int64
	if *maxLookbackDays > 0 {
		maxLookbackTimestamp = util.UnixTimeBeforeDuration(time.Hour * 24 * time.Duration(*maxLookbackDays))
	}

	statusMap, _ := session.AddSession(allowedProjectIds, maxLookbackTimestamp,
		*bufferTimeBeforeCreateSessionInMins, *numRoutines)

	modifiedStatusMap := make(map[uint64]session.Status, 0)
	notModifiedProjects := make([]uint64, 0, 0)

	for pid, status := range statusMap {
		if status.Status == session.StatusNotModified {
			notModifiedProjects = append(notModifiedProjects, pid)
			continue
		}
		modifiedStatusMap[pid] = status
	}

	status := map[string]interface{}{
		"no_session_projects": notModifiedProjects,
		"new_session_status":  modifiedStatusMap,
	}

	if err := util.NotifyThroughSNS(taskID, *env, status); err != nil {
		log.Fatalf("Failed to notify status %+v", status)
	}

	log.WithField("no_of_projects", len(allowedProjectIds)).Info("Successfully added sessions.")
}
