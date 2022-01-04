package main

import (
	"context"
	"factors/model/model"
	"flag"
	"fmt"
	"reflect"

	FB "factors/beam/dashboard_caching"
	C "factors/config"
	U "factors/util"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"

	log "github.com/sirupsen/logrus"
)

/*
Sample query to run on:
Development (Direct Runner):
go run scripts/beam_dashboard_caching.go --project_id='*'

Staging (Dataflow runner):
go run scripts/run_beam_dashboard_caching.go --project_id='2,3' --runner dataflow --project factors-staging \
	--region us-west1 --temp_location gs://factors-staging-misc/beam/tmp/ \
	--staging_location gs://factors-staging-misc/beam/binaries/ \
	--worker_harness_container_image=apache/beam_go_sdk:latest --db_host=10.12.64.2 \
	--db_user autometa_ro --db_pass='<dbPass>' --redis_host='10.0.0.24' \
	--redis_port=8379 --subnetwork='regions/us-west1/subnetworks/us-west-1-factors-staging-subnet-1' \
	--num_workers=1 --max_num_workers=1 --zone='us-west1-b'
*/

var (
	env = flag.String("env", C.DEVELOPMENT, "Environment")

	dbHost = flag.String("db_host", "localhost", "")
	dbPort = flag.Int("db_port", 5432, "")
	dbUser = flag.String("db_user", "autometa", "")
	dbName = flag.String("db_name", "autometa", "")
	dbPass = flag.String("db_pass", "@ut0me7a", "")

	memSQLHost        = flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort        = flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser        = flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName        = flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass        = flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate = flag.String("memsql_cert", "", "")
	primaryDatastore  = flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	redisHost = flag.String("redis_host", "localhost", "")
	redisPort = flag.Int("redis_port", 6379, "")
	sentryDSN = flag.String("sentry_dsn", "", "Sentry DSN")

	projectIDs                      = flag.String("project_id", "3", "Project ids to run for. * for all.")
	excludeProjectIDs               = flag.String("exclude_project_id", "", "Comma separated project ids to exclude for the run")
	multipleTimezoneEnabledProjects = flag.String("timezone_enabled_projects", "", "List of projectIds where multiple timezones are enabled")

	overrideHealthcheckPingID = flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName           = flag.String("app_name", "", "Override default app_name.")
	enableMemSQLRedisWrites   = flag.Bool("enable_mql_redis_writes", false, "To enable redis writes when using MemSQL")
	enableFilterOptimisation  = flag.Bool("enable_filter_optimisation", false,
		"Enables filter optimisation changes for memsql implementation.")
	filterPropertiesStartTimestamp = flag.Int64("filter_properties_start_timestamp", -1,
		"Start timestamp of data available for filtering with parquet on memsql.")
	skipEventNameStepByProjectID        = flag.String("skip_event_name_step_by_project_id", "", "")
	skipUserJoinInEventQueryByProjectID = flag.String("skip_user_join_in_event_query_by_project_id", "", "")
)

func registerStructs() {
	beam.RegisterType(reflect.TypeOf((*model.DashboardUnit)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*model.BeamDashboardUnitCachePayload)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*model.CachingUnitReport)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*model.CachingProjectReport)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*model.FailedDashboardUnitReport)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*FB.CacheResponse)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*FB.GetDashboardUnitCachePayloadsFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*FB.CacheDashboardUnitDoFn)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*model.WebAnalyticsCachePayload)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*FB.GetWebAnalyticsCachePayloadsFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*FB.CacheWebAnalyticsDoFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*FB.GetWebAnalyticsCachePayloadsNowFn)(nil)).Elem())
}

// TODO(prateek): Check a way to add handling for panic and worker errors.
func main() {
	flag.Parse()
	defaultAppName := "beam_dashboard_caching_now"
	defaultHealthcheckPingID := C.HealthcheckBeamDashboardCachingNowPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)
	registerStructs()
	beam.Init()

	// Creating a pipeline.
	p, s := beam.NewPipelineWithRoot()

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:                    *primaryDatastore,
		RedisHost:                           *redisHost,
		RedisPort:                           *redisPort,
		SentryDSN:                           *sentryDSN,
		MultipleTimezoneEnabledProjects:     C.GetTokensFromStringListAsUint64(*multipleTimezoneEnabledProjects),
		EnableFilterOptimisation:            *enableFilterOptimisation,
		FilterPropertiesStartTimestamp:      *filterPropertiesStartTimestamp,
		SkipEventNameStepByProjectID:        *skipEventNameStepByProjectID,
		SkipUserJoinInEventQueryByProjectID: *skipUserJoinInEventQueryByProjectID,
	}
	beam.PipelineOptions.Set("HealthchecksPingID", healthcheckPingID)
	beam.PipelineOptions.Set("StartTime", fmt.Sprint(U.TimeNowUnix()))

	// Create initial PCollection for the projectIDs string passed to be processed.
	projectIDString := beam.Create(s, fmt.Sprintf("%s|%s", *projectIDs, *excludeProjectIDs))

	var webAnalyticsCacheResponses beam.PCollection
	webAnalyticsCachePayloads := beam.ParDo(s, &FB.GetWebAnalyticsCachePayloadsNowFn{Config: config}, projectIDString)
	reShuffledWebAnalyticsCachePayloads := beam.Reshuffle(s, webAnalyticsCachePayloads)

	webAnalyticsCacheResponses = beam.ParDo(s, &FB.CacheWebAnalyticsDoFn{Config: config}, reShuffledWebAnalyticsCachePayloads)

	// Emit (project, cache response) for project level summary.
	webAnalyticsProjectKeyCacheResponses := beam.ParDo(s, FB.EmitProjectKeyCacheResponse, webAnalyticsCacheResponses)
	webAnalyticsProjectGroupedCacheResponses := beam.GroupByKey(s, webAnalyticsProjectKeyCacheResponses)
	_ = beam.ParDo(s, FB.ReportProjectLevelSummary, webAnalyticsProjectGroupedCacheResponses)

	// To log an overall job summary.
	allCacheResponses := webAnalyticsCacheResponses
	allCommonKeyCacheResponses := beam.ParDo(s, FB.EmitCommonKeyCacheResponse, allCacheResponses)
	allCacheResponsesGrouped := beam.GroupByKey(s, allCommonKeyCacheResponses)
	_ = beam.ParDo(s, FB.ReportOverallJobSummary, allCacheResponsesGrouped)

	if err := beamx.Run(context.Background(), p); err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}
}
