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
}

func initConf(config *C.Configuration) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)

	C.InitConf(config.Env)
	C.SetIsBeamPipeline()
	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 5, 2)
	if err != nil {
		// TODO(prateek): Check how a panic here will effect the pipeline.
		log.WithError(err).Panic("Failed to initalize db.")
	}
	C.InitRedisConnection(config.RedisHost, config.RedisPort, true, 20, 0)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.KillDBQueriesOnExit()
}

func emitIndividualProjectID(ctx context.Context, projectIDsString string, emit func(string)) {
	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(projectIDsString, "")
	projectIDs := C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)
	if allProjects {
		var errCode int
		projectIDs, errCode = store.GetStore().GetAllProjectIDs()
		if errCode != http.StatusFound {
			return
		}
	}
	for _, pid := range projectIDs {
		emit(fmt.Sprint(pid))
	}
}

func EmitProjectKeyCacheResponse(ctx context.Context, cacheResponse CacheResponse, emit func(uint64, CacheResponse)) {
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
				"Type":            cacheResponse.CacheResponseType,
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
	for values(&cacheResponse) {
		overallStats[cacheResponse.CacheResponseType]["TimeTaken"] += cacheResponse.TimeTaken
		if cacheResponse.ErrorCode == http.StatusOK {
			overallStats[cacheResponse.CacheResponseType]["SuccessCount"]++
		} else {
			overallStats[cacheResponse.CacheResponseType]["FailedCount"]++
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

	log.WithFields(log.Fields{
		"Method": "reportOverallJobSummary",
	}).Info(message)
	if overallStats[BeamCacheTypeDashboardUnits]["FailedCount"] > 0 || overallStats[BeamCacheTypeWebAnalytics]["FailedCount"] > 0 {
		C.PingHealthcheckForFailure(beam.PipelineOptions.Get("HealthchecksPingID"), message)
	} else {
		C.PingHealthcheckForSuccess(beam.PipelineOptions.Get("HealthchecksPingID"), message)
	}
	return message
}

type GetDashboardUnitCachePayloadsFn struct {
	Config *C.Configuration
}

func (f *GetDashboardUnitCachePayloadsFn) StartBundle(ctx context.Context, emit func(model.BeamDashboardUnitCachePayload)) {
	log.Info("Initializing conf from StartBundle getDashboardUnitCachePayloadsFn")
	initConf(f.Config)
}

func (f *GetDashboardUnitCachePayloadsFn) FinishBundle(ctx context.Context, emit func(model.BeamDashboardUnitCachePayload)) {
	log.Info("Closing DB Connection from FinishBundle getDashboardUnitCachePayloadsFn")
	C.GetServices().Db.Close()
	C.SafeFlushAllCollectors()
}

func (f *GetDashboardUnitCachePayloadsFn) ProcessElement(ctx context.Context, projectsToRunString string,
	emit func(model.BeamDashboardUnitCachePayload)) {

	projectIDSplit := strings.Split(projectsToRunString, "|")
	stringProjectIDs := strings.TrimSpace(projectIDSplit[0])
	excludeProjectIDs := strings.TrimSpace(projectIDSplit[1])

	projectIDs := store.GetStore().GetProjectsToRunForIncludeExcludeString(stringProjectIDs, excludeProjectIDs)
	for _, projectID := range projectIDs {
		dashboardUnits, errCode := store.GetStore().GetDashboardUnitsForProjectID(uint64(projectID))
		if errCode != http.StatusFound {
			continue
		}

		for _, dashboardUnit := range dashboardUnits {
			queryClass, errMsg := store.GetStore().GetQueryAndClassFromDashboardUnit(&dashboardUnit)
			if errMsg == "" && queryClass != model.QueryClassWeb {
				for _, rangeFunction := range U.QueryDateRangePresets {
					from, to := rangeFunction()
					cachePayload := model.BeamDashboardUnitCachePayload{
						DashboardUnit: dashboardUnit,
						QueryClass:    queryClass,
						Query:         dashboardUnit.Query,
						From:          from,
						To:            to,
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
	C.GetServices().Db.Close()
	C.SafeFlushAllCollectors()
}

func (f *CacheDashboardUnitDoFn) ProcessElement(ctx context.Context,
	beamCachePayload model.BeamDashboardUnitCachePayload, emit func(CacheResponse)) {

	FB.GetLogContext().WithFields(log.Fields{
		"Time":     time.Now().UnixNano() / 1000,
		"TimeType": "Start",
	}).Info("ProcessElement log cacheDashboardUnitDoFn")

	baseQuery, err := model.DecodeQueryForClass(beamCachePayload.Query, beamCachePayload.QueryClass)
	if err != nil {
		return
	}
	baseQuery.SetQueryDateRange(beamCachePayload.From, beamCachePayload.To)
	cachePayload := model.DashboardUnitCachePayload{
		DashboardUnit: beamCachePayload.DashboardUnit,
		BaseQuery:     baseQuery,
	}
	startTime := U.TimeNowUnix()
	errCode, errMsg := store.GetStore().CacheDashboardUnitForDateRange(cachePayload)
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
	}
	FB.GetLogContext().WithFields(log.Fields{
		"Time":     time.Now().UnixNano() / 1000,
		"TimeType": "End",
	}).Info("ProcessElement log cacheDashboardUnitDoFn")
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
	C.GetServices().Db.Close()
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

		// Get only last 30mins preset for caching.
		from, to := U.WebAnalyticsQueryDateRangePresets[U.DateRangePreset30Minutes]()
		cachePayload := model.WebAnalyticsCachePayload{
			ProjectID:   projectID,
			DashboardID: dashboardID,
			From:        from,
			To:          to,
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
	C.GetServices().Db.Close()
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
	C.GetServices().Db.Close()
	C.SafeFlushAllCollectors()
}

func (f *CacheWebAnalyticsDoFn) ProcessElement(ctx context.Context,
	cachePayload model.WebAnalyticsCachePayload, emit func(CacheResponse)) {

	FB.GetLogContext().WithFields(log.Fields{
		"Time":     time.Now().UnixNano() / 1000,
		"TimeType": "Start",
	}).Info("ProcessElement log cacheWebAnalyticsDoFn")
	startTime := U.TimeNowUnix()
	errCode := store.GetStore().CacheWebsiteAnalyticsForDateRange(cachePayload)
	timeTaken := U.TimeNowUnix() - startTime

	cacheResponse := CacheResponse{
		ProjectID:         cachePayload.ProjectID,
		DashboardID:       cachePayload.DashboardID,
		From:              cachePayload.From,
		To:                cachePayload.To,
		ErrorCode:         errCode,
		TimeTaken:         timeTaken,
		CacheResponseType: BeamCacheTypeWebAnalytics,
	}
	FB.GetLogContext().WithFields(log.Fields{
		"Time":     time.Now().UnixNano() / 1000,
		"TimeType": "End",
	}).Info("ProcessElement log cacheWebAnalyticsDoFn")
	emit(cacheResponse)
}
