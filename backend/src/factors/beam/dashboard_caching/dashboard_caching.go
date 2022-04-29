package dashboard_caching

import (
	"context"
	FB "factors/beam"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/apache/beam/sdks/go/pkg/beam"
	log "github.com/sirupsen/logrus"
)

// Cache response types to be used in end of the job reporting.
const (
	BeamCacheTypeDashboardUnits string = "Dashboard Units"
	BeamCacheTypeWebAnalytics   string = "Website Analytics"
)

// CacheResponse Struct to store cache response and emit.
type CacheResponse struct {
	ProjectID, DashboardID, DashboardUnitID uint64
	From, To, TimeTaken                     int64
	ErrorCode                               int
	ErrorMessage                            string
	CacheResponseType                       string
	UnitReport                              model.CachingUnitReport
}

func initConf(config *C.Configuration) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)

	C.InitConf(config)
	C.SetIsBeamPipeline()
	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 5, 2)
	if err != nil {
		// TODO(prateek): Check how a panic here will effect the pipeline.
		log.WithError(err).Panic("Failed to initalize db.")
	}
	C.InitRedisConnection(config.RedisHost, config.RedisPort, true, 20, 0)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	model.SetSmartPropertiesReservedNames()
	C.KillDBQueriesOnExit()
}

func EmitProjectKeyCacheResponse(ctx context.Context, cacheResponse CacheResponse, emit func(uint64, CacheResponse)) {
	log.WithFields(log.Fields{
		"Method":          "EmitProjectKeyCacheResponse",
		"ProjectID":       cacheResponse.ProjectID,
		"DashboardID":     cacheResponse.DashboardID,
		"DashboardUnitID": cacheResponse.DashboardUnitID,
		"UnitType":        cacheResponse.CacheResponseType,
		"ErrorMessage":    cacheResponse.ErrorMessage,
	}).Info("EmitProjectKeyCacheResponse in caching.")
	emit(cacheResponse.ProjectID, cacheResponse)
}

func EmitCommonKeyCacheResponse(ctx context.Context, cacheResponse CacheResponse, emit func(uint64, CacheResponse)) {
	emit(0, cacheResponse)
}

func ReportProjectLevelSummary(projectID uint64, values func(*CacheResponse) bool) string {
	var timeTakenForProject int64
	var successCount, failedCount int64

	var cacheResponse CacheResponse
	var cacheResponseType string
	for values(&cacheResponse) {
		cacheResponseType = cacheResponse.CacheResponseType
		timeTakenForProject += cacheResponse.TimeTaken
		if cacheResponse.ErrorCode == http.StatusOK {
			successCount++
		} else {
			failedCount++
			log.WithFields(log.Fields{
				"Method":          "reportProjectLevelSummary",
				"ProjectID":       projectID,
				"DashboardID":     cacheResponse.DashboardID,
				"DashboardUnitID": cacheResponse.DashboardUnitID,
				"UnitType":        cacheResponse.CacheResponseType,
				"ErrorMessage":    cacheResponse.ErrorMessage,
			}).Info("Error in caching.")
		}
	}

	timeTakenString := U.SecondsToHMSString(timeTakenForProject)
	log.WithFields(log.Fields{
		"Method":          "reportProjectLevelSummary",
		"ProjectID":       projectID,
		"TimeTaken":       timeTakenForProject,
		"TimeTakenString": timeTakenString,
		"Success":         successCount,
		"Failed":          failedCount,
	}).Infof("%s: Completed caching for project.", cacheResponseType)
	return fmt.Sprintf("ProjectID: %d, TimeTaken: %s, Success: %d, Failed: %d",
		projectID, timeTakenString, successCount, failedCount)
}

func ReportOverallJobSummary(commonKey uint64, values func(*CacheResponse) bool) string {
	overallStats := make(map[string]map[string]int64)
	overallStats[BeamCacheTypeDashboardUnits] = make(map[string]int64)
	overallStats[BeamCacheTypeWebAnalytics] = make(map[string]int64)

	var cacheResponse CacheResponse
	var allUnitReports []model.CachingUnitReport
	for values(&cacheResponse) {
		overallStats[cacheResponse.CacheResponseType]["TimeTaken"] += cacheResponse.TimeTaken
		if cacheResponse.ErrorCode == http.StatusOK {
			overallStats[cacheResponse.CacheResponseType]["SuccessCount"]++
		} else {
			overallStats[cacheResponse.CacheResponseType]["FailedCount"]++
		}
		if cacheResponse.UnitReport.UnitID != 0 {
			allUnitReports = append(allUnitReports, cacheResponse.UnitReport)
		}
	}
	unitsTimeTakenString := U.SecondsToHMSString(overallStats[BeamCacheTypeDashboardUnits]["TimeTaken"])
	webTimeTakenString := U.SecondsToHMSString(overallStats[BeamCacheTypeWebAnalytics]["TimeTaken"])

	jobStartTime, _ := strconv.ParseInt(beam.PipelineOptions.Get("StartTime"), 10, 64)
	jobEndTime := U.TimeNowUnix()
	overallTimeTaken := U.SecondsToHMSString(jobEndTime - jobStartTime)

	unitsSummary := fmt.Sprintf("Time taken for dashboard units: %s, Success: %d, Failed: %d", unitsTimeTakenString,
		overallStats[BeamCacheTypeDashboardUnits]["SuccessCount"], overallStats[BeamCacheTypeDashboardUnits]["FailedCount"])
	webSummary := fmt.Sprintf("Time taken for web analytics: %s, Success: %d, Failed: %d", webTimeTakenString,
		overallStats[BeamCacheTypeWebAnalytics]["SuccessCount"], overallStats[BeamCacheTypeWebAnalytics]["FailedCount"])
	message := "Overall time taken: " + overallTimeTaken + " | " + unitsSummary + " | " + webSummary

	slowUnits := model.GetNSlowestUnits(allUnitReports, 3)
	failedUnits := model.GetFailedUnitsByProject(allUnitReports)
	slowProjects := model.GetNSlowestProjects(allUnitReports, 5)
	failed, passed, notComputed := model.GetTotalFailedComputedNotComputed(allUnitReports)

	status := map[string]interface{}{
		"Summary":              message,
		"TotalFailed":          failed,
		"TotalPassed":          passed,
		"TotalNotComputed":     notComputed,
		"Top3SlowUnits":        slowUnits,
		"FailedUnitsByProject": failedUnits,
		"Top5SlowProjects":     slowProjects,
	}

	log.WithFields(log.Fields{"Method": "reportOverallJobSummary", "Report": "finalReporting"}).Info(status)
	if overallStats[BeamCacheTypeDashboardUnits]["FailedCount"] > 0 || overallStats[BeamCacheTypeWebAnalytics]["FailedCount"] > 0 {
		C.PingHealthcheckForFailure(beam.PipelineOptions.Get("HealthchecksPingID"), status)
	} else {
		C.PingHealthcheckForSuccess(beam.PipelineOptions.Get("HealthchecksPingID"), status)
	}

	slowUnits = model.GetNSlowestUnits(allUnitReports, 20)
	failedUnits = model.GetFailedUnitsByProject(allUnitReports)
	slowProjects = model.GetNSlowestProjects(allUnitReports, 10)
	log.WithFields(log.Fields{
		"Summary":                       message,
		"TimeTakenForNormalUnits":       unitsTimeTakenString,
		"TimeTakenForWebAnalyticsUnits": webTimeTakenString,
		"TotalFailed":                   failed,
		"TotalPassed":                   passed,
		"TotalNotComputed":              notComputed,
		"Top3SlowUnits":                 slowUnits,
		"FailedUnitsByProject":          failedUnits,
		"Top5SlowProjects":              slowProjects}).Info("Final Caching Job Report")

	return message
}

// CachingJobProps Job level properties to store custom flags for the run.
type CachingJobProps struct {
	OnlyAttribution    int
	SkipAttribution    int
	IsRunningForMemsql int
}

type GetDashboardUnitCachePayloadsFn struct {
	Config   *C.Configuration
	JobProps *CachingJobProps
}

func (f *GetDashboardUnitCachePayloadsFn) StartBundle(ctx context.Context, emit func(model.BeamDashboardUnitCachePayload)) {
	log.Info("Initializing conf from StartBundle getDashboardUnitCachePayloadsFn")
	initConf(f.Config)
}

func (f *GetDashboardUnitCachePayloadsFn) FinishBundle(ctx context.Context, emit func(model.BeamDashboardUnitCachePayload)) {
	log.Info("Closing DB Connection from FinishBundle getDashboardUnitCachePayloadsFn")
	err := C.GetServices().Db.Close()
	if err != nil {
		log.Info("error while closing the Closing DB Connection from FinishBundle GetDashboardUnitCachePayloadsFn")
	}
	C.SafeFlushAllCollectors()
}

// TODO: we need a way to figure if something is getting skipped. Here timezoneString could error out.
func (f *GetDashboardUnitCachePayloadsFn) ProcessElement(ctx context.Context, projectsToRunString string,
	emit func(model.BeamDashboardUnitCachePayload)) {

	projectIDSplit := strings.Split(projectsToRunString, "|")
	stringProjectIDs := strings.TrimSpace(projectIDSplit[0])
	excludeProjectIDs := strings.TrimSpace(projectIDSplit[1])

	logCtx := FB.GetLogContext().WithFields(log.Fields{
		"LogType":           "MemSqlDebug",
		"Method":            "GetDashboardUnitCachePayloadsFn",
		"stringProjectIDs":  stringProjectIDs,
		"excludeProjectIDs": excludeProjectIDs,
	})
	logCtx.Info("Running caching for projects")

	projectIDs := store.GetStore().GetProjectsToRunForIncludeExcludeString(stringProjectIDs, excludeProjectIDs)
	for _, projectID := range projectIDs {
		timezoneString, statusCode := store.GetStore().GetTimezoneForProject(projectID)
		if statusCode != http.StatusFound {
			log.Errorf("Failed to get project Timezone for %d", projectID)
			continue
		}
		dashboardUnits, errCode := store.GetStore().GetDashboardUnitsForProjectID(uint64(projectID))
		if errCode != http.StatusFound {
			continue
		}
		for _, dashboardUnit := range dashboardUnits {
			queryClass, queryInfo, errMsg := store.GetStore().GetQueryAndClassFromDashboardUnit(&dashboardUnit)
			if errMsg == "" && queryClass != model.QueryClassWeb {
				for preset, rangeFunction := range U.QueryDateRangePresets {
					fr, t, errCode := rangeFunction(timezoneString)
					if errCode != nil {
						log.Errorf("Failed to get proper project Timezone for %d", projectID)
						continue
					}
					// Filtering queries on type and range for attribution query
					shouldCache, from, to := model.ShouldCacheUnitForTimeRange(queryClass, preset, fr, t, f.JobProps.OnlyAttribution, f.JobProps.SkipAttribution)
					if !shouldCache {
						continue
					}
					cachePayload := model.BeamDashboardUnitCachePayload{
						DashboardUnit: dashboardUnit,
						QueryClass:    queryClass,
						Query:         queryInfo.Query,
						From:          from,
						To:            to,
						TimeZone:      timezoneString,
					}
					emit(cachePayload)
				}
			}
		}
	}
}

type CacheDashboardUnitDoFn struct {
	Config *C.Configuration
}

func (f *CacheDashboardUnitDoFn) StartBundle(ctx context.Context, emit func(CacheResponse)) {
	log.Info("Initializing conf from StartBundle cacheDashboardUnitDoFn")
	initConf(f.Config)
}

func (f *CacheDashboardUnitDoFn) FinishBundle(ctx context.Context, emit func(CacheResponse)) {
	log.Info("Closing DB Connection from FinishBundle cacheDashboardUnitDoFn")
	err := C.GetServices().Db.Close()
	if err != nil {
		log.Info("error while closing the Closing DB Connection from FinishBundle CacheDashboardUnitDoFn")
	}
	C.SafeFlushAllCollectors()
}

func (f *CacheDashboardUnitDoFn) ProcessElement(ctx context.Context,
	beamCachePayload model.BeamDashboardUnitCachePayload, emit func(CacheResponse)) {

	FB.GetLogContext().WithFields(log.Fields{
		"Time":        time.Now().UnixNano() / 1000,
		"TimeType":    "Start",
		"Project":     beamCachePayload.DashboardUnit.ProjectID,
		"DashboardId": beamCachePayload.DashboardUnit.DashboardId,
		"UnitId":      beamCachePayload.DashboardUnit.ID,
		"QueryClass":  beamCachePayload.QueryClass,
	}).Info("ProcessElement log cacheDashboardUnitDoFn")

	baseQuery, err := model.DecodeQueryForClass(beamCachePayload.Query, beamCachePayload.QueryClass)
	logCtx := FB.GetLogContext().WithFields(log.Fields{
		"Project":     beamCachePayload.DashboardUnit.ProjectID,
		"DashboardId": beamCachePayload.DashboardUnit.DashboardId,
		"UnitId":      beamCachePayload.DashboardUnit.ID,
		"Query":       beamCachePayload.Query,
		"QueryClass":  beamCachePayload.QueryClass,
	})
	if err != nil {
		logCtx.Info("dashboard caching - DecodeQueryForClass failed")
		return
	}

	baseQuery.SetQueryDateRange(beamCachePayload.From, beamCachePayload.To)
	baseQuery.SetTimeZone(beamCachePayload.TimeZone)
	err = baseQuery.TransformDateTypeFilters()
	if err != nil {
		logCtx.WithField("transformed BaseQuery", baseQuery).Info("Error during transformation of DateType filters")
		return
	}
	cachePayload := model.DashboardUnitCachePayload{
		DashboardUnit: beamCachePayload.DashboardUnit,
		BaseQuery:     baseQuery,
	}
	startTime := U.TimeNowUnix()
	errCode, errMsg, cachingReport := store.GetStore().CacheDashboardUnitForDateRange(cachePayload)

	timeTaken := U.TimeNowUnix() - startTime

	dashboardUnit := cachePayload.DashboardUnit
	from, to := cachePayload.BaseQuery.GetQueryDateRange()
	cacheResponse := CacheResponse{
		ProjectID:         dashboardUnit.ProjectID,
		DashboardID:       dashboardUnit.DashboardId,
		DashboardUnitID:   dashboardUnit.ID,
		From:              from,
		To:                to,
		ErrorCode:         errCode,
		ErrorMessage:      errMsg,
		TimeTaken:         timeTaken,
		CacheResponseType: BeamCacheTypeDashboardUnits,
		UnitReport:        cachingReport,
	}
	FB.GetLogContext().WithFields(log.Fields{
		"Time":          time.Now().UnixNano() / 1000,
		"TimeType":      "End",
		"ProjectID":     dashboardUnit.ProjectID,
		"DashboardId":   dashboardUnit.DashboardId,
		"UnitID":        dashboardUnit.ID,
		"cacheResponse": cacheResponse,
	}).Info("ProcessElement log cacheDashboardUnitDoFn")
	if errCode != http.StatusOK {
		logCtx.WithField("transformed  BaseQuery", baseQuery).Info("dashboard caching - couldn't run the caching query")
	}
	emit(cacheResponse)
}

type GetWebAnalyticsCachePayloadsNowFn struct {
	Config *C.Configuration
}

func (f *GetWebAnalyticsCachePayloadsNowFn) StartBundle(ctx context.Context, emit func(model.WebAnalyticsCachePayload)) {
	log.Info("Initializing conf from StartBundle getWebAnalyticsCachePayloadsFn")
	initConf(f.Config)
}

func (f *GetWebAnalyticsCachePayloadsNowFn) FinishBundle(ctx context.Context, emit func(model.WebAnalyticsCachePayload)) {
	log.Info("Closing DB Connection from FinishBundle getWebAnalyticsCachePayloadsFn")
	err := C.GetServices().Db.Close()
	if err != nil {
		log.Info("error while closing the Closing DB Connection from FinishBundle GetWebAnalyticsCachePayloadsNowFn")
	}
	C.SafeFlushAllCollectors()
}

func (f *GetWebAnalyticsCachePayloadsNowFn) ProcessElement(ctx context.Context, projectsToRunString string,
	emit func(model.WebAnalyticsCachePayload)) {

	projectIDSplit := strings.Split(projectsToRunString, "|")
	stringProjectIDs := strings.TrimSpace(projectIDSplit[0])
	excludeProjectIDs := strings.TrimSpace(projectIDSplit[1])

	projectIDs := store.GetStore().GetWebAnalyticsEnabledProjectIDsFromList(stringProjectIDs, excludeProjectIDs)
	for _, projectID := range projectIDs {
		dashboardID, webAnalyticsQueries, errCode := store.GetStore().GetWebAnalyticsQueriesFromDashboardUnits(projectID)
		if errCode == http.StatusNotFound {
			continue
		} else if errCode != http.StatusFound {
			log.Errorf("Failed to get web analytics queries for project %d", projectID)
			continue
		}
		timezoneString, statusCode := store.GetStore().GetTimezoneForProject(projectID)
		if statusCode != http.StatusFound {
			log.Errorf("Failed to get project Timezone for %d", projectID)
			continue
		}
		// Get only last 30mins preset for caching.
		from, to, err := U.WebAnalyticsQueryDateRangePresets[U.DateRangePreset30Minutes](timezoneString)
		if err != nil {
			log.Errorf("Failed to get proper project Timezone for %d", projectID)
			continue
		}
		cachePayload := model.WebAnalyticsCachePayload{
			ProjectID:   projectID,
			DashboardID: dashboardID,
			From:        from,
			To:          to,
			Timezone:    timezoneString,
			Queries:     webAnalyticsQueries,
		}
		emit(cachePayload)
	}
}

type GetWebAnalyticsCachePayloadsFn struct {
	Config *C.Configuration
}

func (f *GetWebAnalyticsCachePayloadsFn) StartBundle(ctx context.Context, emit func(model.WebAnalyticsCachePayload)) {
	log.Info("Initializing conf from StartBundle getWebAnalyticsCachePayloadsFn")
	initConf(f.Config)
}

func (f *GetWebAnalyticsCachePayloadsFn) FinishBundle(ctx context.Context, emit func(model.WebAnalyticsCachePayload)) {
	log.Info("Closing DB Connection from FinishBundle getWebAnalyticsCachePayloadsFn")
	err := C.GetServices().Db.Close()
	if err != nil {
		log.Info("error while closing the Closing DB Connection from FinishBundle GetWebAnalyticsCachePayloadsFn")
	}
	C.SafeFlushAllCollectors()
}

func (f *GetWebAnalyticsCachePayloadsFn) ProcessElement(ctx context.Context, projectsToRunString string,
	emit func(model.WebAnalyticsCachePayload)) {

	projectIDSplit := strings.Split(projectsToRunString, "|")
	stringProjectIDs := strings.TrimSpace(projectIDSplit[0])
	excludeProjectIDs := strings.TrimSpace(projectIDSplit[1])

	projectIDs := store.GetStore().GetWebAnalyticsEnabledProjectIDsFromList(stringProjectIDs, excludeProjectIDs)
	for _, projectID := range projectIDs {
		cachePayloads, errCode, errMsg := store.GetStore().GetWebAnalyticsCachePayloadsForProject(projectID)
		if errCode != http.StatusFound {
			log.Error("Error getting web analytics cache payloads ", errMsg)
			return
		}
		for index := range cachePayloads {
			emit(cachePayloads[index])
		}
	}
}

type CacheWebAnalyticsDoFn struct {
	Config *C.Configuration
}

func (f *CacheWebAnalyticsDoFn) StartBundle(ctx context.Context, emit func(CacheResponse)) {
	log.Info("Initializing conf from StartBundle cacheWebAnalyticsDoFn")
	initConf(f.Config)
}

func (f *CacheWebAnalyticsDoFn) FinishBundle(ctx context.Context, emit func(CacheResponse)) {
	log.Info("Closing DB Connection from FinishBundle cacheWebAnalyticsDoFn")
	err := C.GetServices().Db.Close()
	if err != nil {
		log.Info("error while closing the Closing DB Connection from FinishBundle CacheWebAnalyticsDoFn")
	}
	C.SafeFlushAllCollectors()
}

func (f *CacheWebAnalyticsDoFn) ProcessElement(ctx context.Context,
	cachePayload model.WebAnalyticsCachePayload, emit func(CacheResponse)) {

	FB.GetLogContext().WithFields(log.Fields{
		"Time":     time.Now().UnixNano() / 1000,
		"TimeType": "Start",
	}).Info("ProcessElement log cacheWebAnalyticsDoFn")
	startTime := U.TimeNowUnix()
	errCode, cachingUnitReport := store.GetStore().CacheWebsiteAnalyticsForDateRange(cachePayload)
	timeTaken := U.TimeNowUnix() - startTime

	cacheResponse := CacheResponse{
		ProjectID:         cachePayload.ProjectID,
		DashboardID:       cachePayload.DashboardID,
		From:              cachePayload.From,
		To:                cachePayload.To,
		ErrorCode:         errCode,
		TimeTaken:         timeTaken,
		CacheResponseType: BeamCacheTypeWebAnalytics,
		UnitReport:        cachingUnitReport,
	}
	FB.GetLogContext().WithFields(log.Fields{
		"Time":     time.Now().UnixNano() / 1000,
		"TimeType": "End",
	}).Info("ProcessElement log cacheWebAnalyticsDoFn")
	emit(cacheResponse)
}
