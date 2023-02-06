package main

import (
	C "factors/config"
	"factors/model/store"
	"factors/sdk"
	"factors/util"
	U "factors/util"
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	RedisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	projectIds := flag.String("project_ids", "", "Projects for which the cache is to be refreshed")
	eventRecordsLimit := flag.Int("event_records_limit", 100000, "")
	usersProcessedLimit := flag.Int("users_processed_limit", 10000, "")
	eventsLimit := flag.Int("events_limit", 10000, "")
	propertiesLimit := flag.Int("properties_limit", 5000, "")
	valuesLimit := flag.Int("values_limit", 2000, "")
	lookBackDays := flag.Int("look_back_days", 1, "")

	/* The following three keys are to be used only when there is a need to backfill a specific project in a particular time range
	perQueryPullRange - for how many days is the data need to be backfilled
	overrideDateRangeEnd - backfill range end
	So the backfll duration will be from (overrideDateRangeEnd-perQueryPullRange) to (overrideDateRangeEnd-1)
	This will need the SkipExpiry flag to also set to true since we dont want that data to be deleted from cache because any key older than past 30 days is deleted from cache
	This is primarily used today to backfill demo data where the data is not flowing in continuously instead available for a particular date range and we need that data in cache as well
	Handler also has overrides to pull data from specific range because usually it pull last 30 days data*/
	overrideDateRangeEnd := flag.String("overrride_date_range_end", "", "")
	perQueryPullRange := flag.Int("per_query_pull_range", 0, "")
	skipExpiry := flag.Bool("skip_expiry_for_cache", false, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "Task#InstantiateEventUserCache"
	defer util.NotifyOnPanic(taskID, *env)

	appName := "instantiate_event_user_cache"
	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
			AppName:  appName,
		},
		PrimaryDatastore:    *primaryDatastore,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *RedisPortPersistent,
		SentryDSN:           *sentryDSN,
	}

	C.InitConf(config)

	// Will allow all 50/50 connection to be idle on the pool.
	// As we allow num_routines (per project) as per no.of db connections
	// and will be used continiously.
	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 50, 50)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db in add session.")
	}

	// Cache dependency for requests not using queue.
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	defer C.SafeFlushAllCollectors()

	var startOfCurrentDay time.Time
	if *overrideDateRangeEnd == "" {
		currentTime := U.TimeNowZ()
		startOfCurrentDay = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		startOfCurrentDay, _ = time.Parse(U.DATETIME_FORMAT_YYYYMMDD, *overrideDateRangeEnd)
	}

	projectIdMap := util.GetIntBoolMapFromStringList(projectIds)

	for projectId, _ := range projectIdMap {
		sdk.BackFillEventDataInCacheFromDb(projectId, startOfCurrentDay, *lookBackDays, *eventsLimit, *propertiesLimit, *valuesLimit, *eventRecordsLimit, *perQueryPullRange, *skipExpiry)
		store.GetStore().BackFillUserDataInCacheFromDb(projectId, startOfCurrentDay, *usersProcessedLimit, *propertiesLimit, *valuesLimit, *skipExpiry)
	}
	fmt.Println("Done!!!")
}
