package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"reflect"

	FB "factors/beam/dashboard_caching"
	C "factors/config"
	"factors/model/model"
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

	memSQLHost        = flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort        = flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser        = flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName        = flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass        = flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate = flag.String("memsql_cert", "", "")
	primaryDatastore  = flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	memSQLDBMaxOpenConnections = flag.Int("memsql_max_open_connections", 100, "Max no.of open connections allowed on connection pool of memsql")
	memSQLDBMaxIdleConnections = flag.Int("memsql_max_idle_connections", 50, "Max no.of idle connections allowed on connection pool of memsql")

	redisHost = flag.String("redis_host", "localhost", "")
	redisPort = flag.Int("redis_port", 6379, "")
	sentryDSN = flag.String("sentry_dsn", "", "Sentry DSN")

	projectIDs        = flag.String("project_id", "3", "Project ids to run for. * for all.")
	excludeProjectIDs = flag.String("exclude_project_id", "", "Comma separated project ids to exclude for the run")
	onlyWebAnalytics  = flag.Bool("only_web_analytics", false, "Cache only web analytics dashboards.")
	skipWebAnalytics  = flag.Bool("skip_web_analytics", false, "Skip the web analytics and run other.")
	// better to have 0 or 1 values instead of false/true
	onlyAttribution  = flag.Int("only_attribution", 0, "Cache only Attribution dashboards.")
	skipAttribution  = flag.Int("skip_attribution", 0, "Skip the Attribution and run other.")
	runningForMemsql = flag.Int("running_for_memsql", 0, "Disable routines for memsql.")

	overrideHealthcheckPingID = flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName           = flag.String("app_name", "", "Override default app_name.")
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

}

type cachingJobSummary struct {
	TimeTaken                 int64
	SuccessCount, FailedCount int64
}

type jobStatsAccumulator struct {
	WebAnalyticsSummary   map[int64]cachingJobSummary
	DashboardUnitsSummary map[int64]cachingJobSummary
}

type reportJobSummaryCombineFn struct{}

func (f *reportJobSummaryCombineFn) CreateAccumulator() jobStatsAccumulator {
	return jobStatsAccumulator{
		DashboardUnitsSummary: make(map[int64]cachingJobSummary),
		WebAnalyticsSummary:   make(map[int64]cachingJobSummary),
	}
}

func (f *reportJobSummaryCombineFn) AddInput(a jobStatsAccumulator, value interface{}) jobStatsAccumulator {
	_, ok := value.(FB.CacheResponse)
	if ok {
		castedValue := value.(FB.CacheResponse)
		var successCount, failedCount int64
		if castedValue.ErrorCode == http.StatusOK {
			successCount++
		} else {
			failedCount++
		}
		if _, found := a.DashboardUnitsSummary[castedValue.ProjectID]; found {
			a.DashboardUnitsSummary[castedValue.ProjectID] = cachingJobSummary{
				TimeTaken:    a.DashboardUnitsSummary[castedValue.ProjectID].TimeTaken + castedValue.TimeTaken,
				SuccessCount: a.DashboardUnitsSummary[castedValue.ProjectID].SuccessCount + successCount,
				FailedCount:  a.DashboardUnitsSummary[castedValue.ProjectID].FailedCount + failedCount,
			}
		} else {
			a.DashboardUnitsSummary[castedValue.ProjectID] = cachingJobSummary{
				TimeTaken:    castedValue.TimeTaken,
				SuccessCount: successCount,
				FailedCount:  failedCount,
			}
		}
	} else {
		castedValue := value.(FB.CacheResponse)
		var successCount, failedCount int64
		if castedValue.ErrorCode == http.StatusOK {
			successCount++
		} else {
			failedCount++
		}
		if _, found := a.WebAnalyticsSummary[castedValue.ProjectID]; found {
			a.WebAnalyticsSummary[castedValue.ProjectID] = cachingJobSummary{
				TimeTaken:    a.WebAnalyticsSummary[castedValue.ProjectID].TimeTaken + castedValue.TimeTaken,
				SuccessCount: a.WebAnalyticsSummary[castedValue.ProjectID].SuccessCount + successCount,
				FailedCount:  a.WebAnalyticsSummary[castedValue.ProjectID].FailedCount + failedCount,
			}
		} else {
			a.WebAnalyticsSummary[castedValue.ProjectID] = cachingJobSummary{
				TimeTaken:    castedValue.TimeTaken,
				SuccessCount: successCount,
				FailedCount:  failedCount,
			}
		}
	}
	return a
}

func (f *reportJobSummaryCombineFn) MergeAccumulators(a, b jobStatsAccumulator) jobStatsAccumulator {
	mergedAccumulator := jobStatsAccumulator{
		DashboardUnitsSummary: make(map[int64]cachingJobSummary),
		WebAnalyticsSummary:   make(map[int64]cachingJobSummary),
	}

	for _, accum := range append([]jobStatsAccumulator{a}, b) {
		for projectID, jobSummary := range accum.DashboardUnitsSummary {
			mergedAccumulator.DashboardUnitsSummary[projectID] = jobSummary
		}

		for projectID, jobSummary := range accum.WebAnalyticsSummary {
			mergedAccumulator.WebAnalyticsSummary[projectID] = jobSummary
		}
	}
	return mergedAccumulator
}

func (f *reportJobSummaryCombineFn) ExtractOutput(a jobStatsAccumulator) {
	var overallTimeTaken, overallSuccessCount, overallFailedCount int64
	for projectID, jobSummary := range a.DashboardUnitsSummary {
		log.WithFields(log.Fields{
			"ProjectID":       projectID,
			"TimeTaken":       jobSummary.TimeTaken,
			"TimeTakenString": U.SecondsToHMSString(jobSummary.TimeTaken),
		}).Info("Completed dashboard unit caching for project %d. Success: %d, Failed: %d",
			projectID, jobSummary.SuccessCount, jobSummary.FailedCount)

		overallTimeTaken += jobSummary.TimeTaken
		overallSuccessCount += jobSummary.SuccessCount
		overallFailedCount += jobSummary.FailedCount
	}
	overallSummary := fmt.Sprintf("Time taken for dashboard units: %s (Success: %d, Failed: %d).",
		U.SecondsToHMSString(overallTimeTaken), overallSuccessCount, overallFailedCount)

	var webOverallTimeTaken, webOverallSuccessCount, webOverallFailedCount int64
	for projectID, jobSummary := range a.WebAnalyticsSummary {
		log.WithFields(log.Fields{
			"ProjectID":       projectID,
			"TimeTaken":       jobSummary.TimeTaken,
			"TimeTakenString": U.SecondsToHMSString(jobSummary.TimeTaken),
		}).Info("Completed web analytics caching for project %d. Success: %d, Failed: %d",
			projectID, jobSummary.SuccessCount, jobSummary.FailedCount)

		webOverallTimeTaken += jobSummary.TimeTaken
		webOverallSuccessCount += jobSummary.SuccessCount
		webOverallFailedCount += jobSummary.FailedCount
	}
	webOverallSummary := fmt.Sprintf("Time taken for web analytics: %s (Success: %d, Failed: %d).",
		U.SecondsToHMSString(webOverallTimeTaken), webOverallSuccessCount, webOverallFailedCount)

	message := overallSummary + " " + webOverallSummary
	if overallFailedCount > 0 || webOverallFailedCount > 0 {
		C.PingHealthcheckForFailure(beam.PipelineOptions.Get("HealthchecksPingID"), message)
	} else {
		C.PingHealthcheckForSuccess(beam.PipelineOptions.Get("HealthchecksPingID"), message)
	}
}

// TODO(prateek): Check a way to add handling for panic and worker errors.
func main() {
	flag.Parse()
	defaultAppName := "beam_dashboard_caching"
	defaultHealthcheckPingID := C.HealthcheckBeamDashboardCachingPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)
	registerStructs()
	beam.Init()

	if *skipWebAnalytics && *onlyWebAnalytics {
		log.Fatal("Both skip and only web analytics can not be set")
	}

	// Creating a pipeline.
	p, s := beam.NewPipelineWithRoot()

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

			MaxOpenConnections:     *memSQLDBMaxOpenConnections,
			MaxIdleConnections:     *memSQLDBMaxIdleConnections,
			UseExactConnFromConfig: true,
		},
		PrimaryDatastore:                    *primaryDatastore,
		RedisHost:                           *redisHost,
		RedisPort:                           *redisPort,
		SentryDSN:                           *sentryDSN,
		EnableFilterOptimisation:            *enableFilterOptimisation,
		FilterPropertiesStartTimestamp:      *filterPropertiesStartTimestamp,
		SkipAttributionDashboardCaching:     *skipAttribution,
		OnlyAttributionDashboardCaching:     *onlyAttribution,
		IsRunningForMemsql:                  *runningForMemsql,
		SkipEventNameStepByProjectID:        *skipEventNameStepByProjectID,
		SkipUserJoinInEventQueryByProjectID: *skipUserJoinInEventQueryByProjectID,
	}
	beam.PipelineOptions.Set("HealthchecksPingID", healthcheckPingID)
	beam.PipelineOptions.Set("StartTime", fmt.Sprint(U.TimeNowUnix()))

	// Create initial PCollection for the projectIDs string passed to be processed.
	projectIDString := beam.Create(s, fmt.Sprintf("%s|%s", *projectIDs, *excludeProjectIDs))

	dashboardJobProps := &FB.CachingJobProps{OnlyAttribution: *onlyAttribution, SkipAttribution: *skipAttribution, IsRunningForMemsql: *runningForMemsql}

	var cacheResponses, webAnalyticsCacheResponses beam.PCollection
	if !*onlyWebAnalytics {
		// Fetches dashboard_units for all project_id and emits a unit of work along with query and date range.
		cachePayloads := beam.ParDo(s, &FB.GetDashboardUnitCachePayloadsFn{Config: config, JobProps: dashboardJobProps}, projectIDString)
		reShuffledCachePayloads := beam.Reshuffle(s, cachePayloads)
		// Processes each DashboardUnitCachePayload and emits a cache response which includes errCode, errMsg, time etc.
		cacheResponses = beam.ParDo(s, &FB.CacheDashboardUnitDoFn{Config: config}, reShuffledCachePayloads)

		projectKeyCacheResponses := beam.ParDo(s, FB.EmitProjectKeyCacheResponse, cacheResponses)
		projectGroupedCacheResponses := beam.GroupByKey(s, projectKeyCacheResponses)
		_ = beam.ParDo(s, FB.ReportProjectLevelSummary, projectGroupedCacheResponses)
	}

	if !*skipWebAnalytics {
		webAnalyticsCachePayloads := beam.ParDo(s, &FB.GetWebAnalyticsCachePayloadsFn{Config: config}, projectIDString)
		reShuffledWebAnalyticsCachePayloads := beam.Reshuffle(s, webAnalyticsCachePayloads)

		webAnalyticsCacheResponses = beam.ParDo(s, &FB.CacheWebAnalyticsDoFn{Config: config}, reShuffledWebAnalyticsCachePayloads)

		// Emit (project, cache response) for project level summary.
		webAnalyticsProjectKeyCacheResponses := beam.ParDo(s, FB.EmitProjectKeyCacheResponse, webAnalyticsCacheResponses)
		webAnalyticsProjectGroupedCacheResponses := beam.GroupByKey(s, webAnalyticsProjectKeyCacheResponses)
		_ = beam.ParDo(s, FB.ReportProjectLevelSummary, webAnalyticsProjectGroupedCacheResponses)
	}

	// To log an overall job summary.
	var allCacheResponses beam.PCollection
	if *onlyWebAnalytics {
		allCacheResponses = webAnalyticsCacheResponses
	} else if *skipWebAnalytics {
		allCacheResponses = cacheResponses
	} else {
		allCacheResponses = beam.Flatten(s, cacheResponses, webAnalyticsCacheResponses)
	}
	allCommonKeyCacheResponses := beam.ParDo(s, FB.EmitCommonKeyCacheResponse, allCacheResponses)
	allCacheResponsesGrouped := beam.GroupByKey(s, allCommonKeyCacheResponses)
	_ = beam.ParDo(s, FB.ReportOverallJobSummary, allCacheResponsesGrouped)

	if err := beamx.Run(context.Background(), p); err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}
}
