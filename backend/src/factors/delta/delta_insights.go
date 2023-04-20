package delta

import (
	"bytes"
	"encoding/json"
	"factors/filestore"
	"factors/merge"
	serviceDisk "factors/services/disk"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	U "factors/util"

	"factors/model/store"

	E "factors/event_match"
	M "factors/model/model"
)

var isDownloaded bool

func StaticDashboardListForMailer() []M.DashboardUnit {
	dashboardUnit := make([]M.DashboardUnit, 0)
	dashboardUnit = append(dashboardUnit, M.DashboardUnit{
		QueryId: 1,
	})
	dashboardUnit = append(dashboardUnit, M.DashboardUnit{
		QueryId: 2,
	})
	dashboardUnit = append(dashboardUnit, M.DashboardUnit{
		QueryId: 3,
	})
	dashboardUnit = append(dashboardUnit, M.DashboardUnit{
		QueryId: 4,
	})
	dashboardUnit = append(dashboardUnit, M.DashboardUnit{
		QueryId: 5,
	})
	dashboardUnit = append(dashboardUnit, M.DashboardUnit{
		QueryId: 6,
	})
	return dashboardUnit
}

func getQueryTypeAndClass(queryId int64) (bool, string, string) {
	if queryId == 1 {
		return false, CRM, M.QueryClassEvents
	}
	if queryId == 2 {
		return true, WebsiteEvent, M.QueryClassEvents
	}
	if queryId == 3 {
		return false, CRM, M.QueryClassEvents
	}
	if queryId == 4 {
		return true, KPI, M.QueryClassKPI
	}
	if queryId == 5 {
		return true, KPI, M.QueryClassKPI
	}
	if queryId == 6 {
		return true, KPI, M.QueryClassKPI
	}
	return false, "", ""
}

func getQueryHeader(queryId int64) string {
	if queryId == 1 {
		return "Website Sessions"
	}
	if queryId == 2 {
		return "Form Submitted"
	}
	if queryId == 3 {
		return "Hubspot Contact Created"
	}
	if queryId == 4 {
		return "All Channel Spend"
	}
	if queryId == 5 {
		return "Website Session - Bounce Rate"
	}
	if queryId == 6 {
		return "Average Session Duration"
	}
	return ""
}

func staticQueriesForMailer(queryId int64) (Query, MultiFunnelQuery, M.KPIQueryGroup, bool, bool, bool, map[string]bool) {
	if queryId == 1 {
		return Query{
			Id: 1,
			Base: E.EventsCriteria{
				Id:       0,
				Operator: "And",
				EventCriterionList: []E.EventCriterion{
					{
						Id:                  0,
						Name:                "$session",
						EqualityFlag:        true,
						FilterCriterionList: nil,
					},
				},
			},
			Target: E.EventsCriteria{
				Id:       0,
				Operator: "And",
				EventCriterionList: []E.EventCriterion{
					{
						Id:                  0,
						Name:                "$session",
						EqualityFlag:        true,
						FilterCriterionList: nil,
					},
				},
			},
		}, MultiFunnelQuery{}, M.KPIQueryGroup{}, true, true, false, make(map[string]bool)
	}
	if queryId == 2 {
		return Query{
			Id: 2,
			Base: E.EventsCriteria{
				Id:       0,
				Operator: "And",
				EventCriterionList: []E.EventCriterion{
					{
						Id:                  0,
						Name:                "$session",
						EqualityFlag:        true,
						FilterCriterionList: nil,
					},
				},
			},
			Target: E.EventsCriteria{
				Id:       0,
				Operator: "And",
				EventCriterionList: []E.EventCriterion{
					{
						Id:                  0,
						Name:                "$form_submitted",
						EqualityFlag:        true,
						FilterCriterionList: nil,
					},
				},
			},
		}, MultiFunnelQuery{}, M.KPIQueryGroup{}, true, true, false, make(map[string]bool)
	}
	if queryId == 3 {
		return Query{
			Id: 3,
			Base: E.EventsCriteria{
				Id:       0,
				Operator: "And",
				EventCriterionList: []E.EventCriterion{
					{
						Id:                  0,
						Name:                "$session",
						EqualityFlag:        true,
						FilterCriterionList: nil,
					},
				},
			},
			Target: E.EventsCriteria{
				Id:       0,
				Operator: "And",
				EventCriterionList: []E.EventCriterion{
					{
						Id:           0,
						Name:         "$hubspot_contact_created",
						EqualityFlag: true,
						FilterCriterionList: []E.EventFilterCriterion{
							{
								Key:            "$hubspot_contact_hs_analytics_source",
								Type:           "categorical",
								PropertiesMode: "user",
								Values: []E.OperatorValueTuple{
									{
										Operator:  "notEqual",
										Value:     "OFFLINE",
										LogicalOp: "AND",
									},
								},
							},
						},
					},
				},
			},
		}, MultiFunnelQuery{}, M.KPIQueryGroup{}, true, true, false, make(map[string]bool)
	}
	if queryId == 4 {
		return Query{}, MultiFunnelQuery{}, M.KPIQueryGroup{
			Class: "kpi",
			Queries: []M.KPIQuery{
				{
					Category:        "channels",
					DisplayCategory: "all_channels_metrics",
				},
				{
					Category:        "channels",
					DisplayCategory: "all_channels_metrics",
					Metrics: []string{
						"spend",
					},
				},
			},
			GlobalFilters: []M.KPIFilter{},
			GlobalGroupBy: []M.KPIGroupBy{},
		}, true, false, false, make(map[string]bool)
	}
	if queryId == 5 {
		return Query{}, MultiFunnelQuery{}, M.KPIQueryGroup{
			Class: "kpi",
			Queries: []M.KPIQuery{
				{
					Category:        "events",
					DisplayCategory: "website_session",
					Metrics: []string{
						"bounce_rate",
					},
				},
				{
					Category:        "events",
					DisplayCategory: "website_session",
					Metrics: []string{
						"bounce_rate",
					},
				},
			},
			GlobalFilters: []M.KPIFilter{},
			GlobalGroupBy: []M.KPIGroupBy{},
		}, true, false, false, make(map[string]bool)
	}
	if queryId == 6 {
		return Query{}, MultiFunnelQuery{}, M.KPIQueryGroup{
			Class: "kpi",
			Queries: []M.KPIQuery{
				{
					Category:        "events",
					DisplayCategory: "website_session",
					Metrics: []string{
						"average_session_duration",
					},
				},
				{
					Category:        "events",
					DisplayCategory: "website_session",
					Metrics: []string{
						"average_session_duration",
					},
				},
			},
			GlobalFilters: []M.KPIFilter{},
			GlobalGroupBy: []M.KPIGroupBy{},
		}, true, false, false, make(map[string]bool)
	}
	return Query{}, MultiFunnelQuery{}, M.KPIQueryGroup{}, false, false, false, nil
}

// ComputeDeltaInsights Take details about two periods, and get delta (difference based) insights
func ComputeDeltaInsights(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	status := make(map[string]interface{})
	var insightId uint64
	if configs["endTimestamp"].(int64) > U.TimeNowUnix() {
		status["error"] = "invalid end timestamp"
		return status, false
	}
	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	modelCloudManager := configs["modelCloudManager"].(*filestore.FileManager)
	archiveCloudManager := configs["archiveCloudManager"].(*filestore.FileManager)
	tmpCloudManager := configs["tmpCloudManager"].(*filestore.FileManager)
	sortedCloudManager := configs["sortedCloudManager"].(*filestore.FileManager)
	useBucketV2 := configs["useBucketV2"].(bool)
	hardPull := configs["hardPull"].(bool)
	pulledMap := make(map[int64]map[string]bool)

	beamConfig := configs["beamConfig"].(*merge.RunBeamConfig)

	whitelistedDashboardUnits := configs["whitelistedDashboardUnits"].(map[string]bool)
	log.Info("model bucket : ", (*modelCloudManager).GetBucketName())
	log.Info("sorted data bucket : ", (*sortedCloudManager).GetBucketName())
	log.Info("archive bucket : ", (*archiveCloudManager).GetBucketName())
	k := configs["k"].(int)
	// Cross Period Insights
	periodCodesWithWeekNMinus1 := []Period{
		{
			From: configs["startTimestamp"].(int64) - 604800,
			To:   configs["endTimestamp"].(int64) - 604800,
		},
		{
			From: configs["startTimestamp"].(int64),
			To:   configs["endTimestamp"].(int64),
		},
	}
	for _, period := range periodCodesWithWeekNMinus1 {
		pulledMap[period.From] = make(map[string]bool)
	}

	mailerRun := false
	if configs["run_type"] != nil && configs["run_type"].(string) == "mailer" {
		mailerRun = true
	}
	skipWpi := configs["skipWpi"].(bool)
	skipWpi2 := configs["skipWpi2"].(bool)
	log.Info("Reading delta query.")
	computedQueries := make(map[int64]bool)
	var dashboardUnits []M.DashboardUnit
	if mailerRun == true {
		// Janani Add a static list
		dashboardUnits = StaticDashboardListForMailer()
	} else {
		dashboardUnits, _ = store.GetStore().GetDashboardUnitsForProjectID(projectId)
	}
	isDownloaded = false
	totalProperties := make(map[string]bool)
	for _, dashboardUnit := range dashboardUnits {
		if !(whitelistedDashboardUnits["*"] || whitelistedDashboardUnits[fmt.Sprintf("%v", dashboardUnit.ID)]) {
			continue
		}
		queryIdString := fmt.Sprintf("%v", dashboardUnit.QueryId)
		deltaQuery, multiStepQuery, kpiQuery, isEnabled, isEventOccurence, isMultiStep, properties := Query{}, MultiFunnelQuery{}, M.KPIQueryGroup{}, false, false, false, make(map[string]bool)
		if mailerRun == true {
			deltaQuery, multiStepQuery, kpiQuery, isEnabled, isEventOccurence, isMultiStep, properties = staticQueriesForMailer(dashboardUnit.QueryId)
		} else {
			deltaQuery, multiStepQuery, kpiQuery, isEnabled, isEventOccurence, isMultiStep, properties = IsDashboardUnitWIEnabled(dashboardUnit)
		} // Check if this is a valid query with valid filters
		if !isEnabled || computedQueries[dashboardUnit.QueryId] {
			continue
		}

		if len(kpiQuery.Queries) > 0 {
			if !configs["runKpi"].(bool) {
				continue
			}
			if err := createKpiInsights(diskManager, archiveCloudManager, tmpCloudManager, sortedCloudManager, modelCloudManager, periodCodesWithWeekNMinus1, projectId,
				dashboardUnit.QueryId, kpiQuery, k, skipWpi, skipWpi2, mailerRun, beamConfig, useBucketV2, hardPull, pulledMap, status); err != nil {
				deltaComputeLog.WithError(err).Error("KPI insights error query: ", dashboardUnit.QueryId)
				status["error-kpi-query-"+queryIdString] = err
				continue
			}
		} else {
			// Within Period Insights
			// TODO: $others values are not getting propagated to the second pass. Do that.
			// TODO: This was changed from set to map
			unionOfFeatures := make(map[string]map[string]bool)
			deltaComputeLog.Info("1st pass: Scanning events file to get top-k base features for each period.")
			err := processSeparatePeriods(projectId, periodCodesWithWeekNMinus1, archiveCloudManager, tmpCloudManager, sortedCloudManager, modelCloudManager, diskManager, deltaQuery, multiStepQuery, k, &unionOfFeatures, 1, isEventOccurence, isMultiStep, skipWpi, skipWpi2, mailerRun, beamConfig, useBucketV2, hardPull, pulledMap)
			if err != nil {
				deltaComputeLog.WithError(err).Error("failed to process wpi pass 1")
				status["error-wpi-pass1-"+queryIdString] = err.Error()
				continue
			}
			isDownloaded = true
			deltaComputeLog.Info("2nd pass: Scanning events file again to compute counts for union of features.")
			err = processSeparatePeriods(projectId, periodCodesWithWeekNMinus1, archiveCloudManager, tmpCloudManager, sortedCloudManager, modelCloudManager, diskManager, deltaQuery, multiStepQuery, k, &unionOfFeatures, 2, isEventOccurence, isMultiStep, skipWpi, skipWpi2, mailerRun, beamConfig, useBucketV2, hardPull, pulledMap)
			if err != nil {
				deltaComputeLog.WithError(err).Error("failed to process wpi pass 2")
				status["error-wpi-pass2-"+queryIdString] = err.Error()
				continue
			}
			deltaComputeLog.Info("Computing cross-period insights.")
			var queryId int64
			if deltaQuery.Id == 0 {
				queryId = multiStepQuery.Id
			} else {
				queryId = deltaQuery.Id
			}
			err = processCrossPeriods(periodCodesWithWeekNMinus1, diskManager, projectId, k, queryId, modelCloudManager, mailerRun)
			if err != nil {
				deltaComputeLog.WithError(err).Error("failed to process cpi")
				status["error-cpi-"+queryIdString] = err.Error()
				continue
			}
		}

		insightId = uint64(U.TimeNowUnix())
		computedQueries[dashboardUnit.QueryId] = true
		errCode, errMsg := store.GetStore().CreateWeeklyInsightsMetadata(&M.WeeklyInsightsMetadata{
			ProjectId:           projectId,
			QueryId:             dashboardUnit.QueryId,
			BaseStartTime:       periodCodesWithWeekNMinus1[1].From,
			BaseEndTime:         periodCodesWithWeekNMinus1[1].To,
			ComparisonStartTime: periodCodesWithWeekNMinus1[0].From,
			ComparisonEndTime:   periodCodesWithWeekNMinus1[0].To,
			InsightType:         "w",
			InsightId:           insightId,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		})
		if errCode != http.StatusCreated {
			deltaComputeLog.Error(errMsg)
			status["error"] = errMsg
			return status, false
		}
		for property := range properties {
			totalProperties[property] = true
		}
	}
	if mailerRun == false {
		writePropertiesPath(projectId, totalProperties, *modelCloudManager)
	}
	status["InsightId"] = insightId
	return status, true
}

func processCrossPeriods(periodCodes []Period, diskManager *serviceDisk.DiskDriver, projectId int64, k int, queryId int64, cloudManager *filestore.FileManager, mailerRun bool) error {
	for i, periodCode1 := range periodCodes {
		for j, periodCode2 := range periodCodes {
			if i >= j {
				continue
			}
			var wpi1 WithinPeriodInsights
			dateString1 := U.GetDateOnlyFromTimestampZ(periodCode1.From)
			efTmpPath1, efTmpName1 := diskManager.GetInsightsWpiFilePathAndName(projectId, dateString1, queryId, k, mailerRun)
			err := ReadFromJSONFile(efTmpPath1+efTmpName1, &wpi1)
			if err != nil {
				deltaComputeLog.WithError(err).Error("failed to read json file")
				return err
			}
			if err := os.Remove(efTmpPath1 + efTmpName1); err != nil {
				deltaComputeLog.WithError(err).Error("failed to remove local wpi file")
				return err
			}
			var wpi2 WithinPeriodInsights
			dateString2 := U.GetDateOnlyFromTimestampZ(periodCode2.From)
			efTmpPath2, efTmpName2 := diskManager.GetInsightsWpiFilePathAndName(projectId, dateString2, queryId, k, mailerRun)

			err = ReadFromJSONFile(efTmpPath2+efTmpName2, &wpi2)
			if err != nil {
				deltaComputeLog.WithError(err).Error("failed to read json file")
				return err
			}
			if err := os.Remove(efTmpPath2 + efTmpName2); err != nil {
				deltaComputeLog.WithError(err).Error("failed to remove local wpi file")
				return err
			}
			crossPeriodInsights, err := ComputeCrossPeriodInsights(wpi1, wpi2)
			if err != nil {
				deltaComputeLog.WithError(err).Error(fmt.Sprintf("Could not compute cross insights for periods %v VS %v", periodCode1, periodCode2))
				return err
			}
			periodPair := PeriodPair{First: periodCode1, Second: periodCode2}
			crossPeriodInsights.Periods = periodPair
			crossPeriodInsightsBytes, err := json.Marshal(crossPeriodInsights)
			if err != nil {
				deltaComputeLog.WithFields(log.Fields{"err": err}).Error("failed to unmarshal events Info.")
				return err
			}
			err = writeCpiPath(projectId, periodPair.Second, queryId, k, bytes.NewReader(crossPeriodInsightsBytes), *cloudManager, mailerRun)
			if err != nil {
				deltaComputeLog.WithFields(log.Fields{"err": err}).Error("failed to write files to cloud")
				return err
			}
		}
	}
	return nil
}

func processSeparatePeriods(projectId int64, periodCodes []Period, archiveCloudManager, tmpCloudManager, sortedCloudManager, modelCloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, deltaQuery Query, multiStepQuery MultiFunnelQuery, k int, unionOfFeatures *(map[string]map[string]bool), passId int, isEventOccurence bool, isMultiStep bool, skipWpi bool, skipWpi2 bool, mailerRun bool, beamConfig *merge.RunBeamConfig, useBucketV2, hardPull bool, pulledMap map[int64]map[string]bool) error {
	earlierWeekMap := make(map[int64]bool)
	earlierWeekMap[periodCodes[0].From] = periodCodes[0].From < periodCodes[1].From
	earlierWeekMap[periodCodes[1].From] = periodCodes[0].From > periodCodes[1].From
	for _, periodCode := range periodCodes {
		fileDownloaded := false
		if !skipWpi || (!earlierWeekMap[periodCode.From] && !skipWpi2) {
			err := processSinglePeriodData(projectId, periodCode, archiveCloudManager, tmpCloudManager, sortedCloudManager, modelCloudManager, diskManager, deltaQuery, multiStepQuery, k, unionOfFeatures, passId, isEventOccurence, isMultiStep, mailerRun, beamConfig, useBucketV2, hardPull, pulledMap)
			if err != nil {
				return err
			}
		} else {
			dateString := U.GetDateOnlyFromTimestampZ(periodCode.From)
			if deltaQuery.Id == 0 {
				deltaQuery.Id = multiStepQuery.Id
			}
			path, name := (*modelCloudManager).GetInsightsWpiFilePathAndName(projectId, dateString, deltaQuery.Id, k, mailerRun)
			reader, err := (*modelCloudManager).Get(path, name)
			if err != nil {
				log.WithFields(log.Fields{"err": err, "filePath": path,
					"eventFileName": name}).Info("Failed to skip wpi: failed to fetch from cloud path")
			} else {
				efTmpPath, efTmpName := (*diskManager).GetInsightsWpiFilePathAndName(projectId, dateString, deltaQuery.Id, k, mailerRun)
				err = (*diskManager).Create(efTmpPath, efTmpName, reader)
				if err != nil {
					log.WithFields(log.Fields{"err": err, "filePath": efTmpPath,
						"eventFileName": efTmpName}).Error("Failed to skip wpi: failed to write to temp path")
				} else {
					fileDownloaded = true
				}
			}
			if !fileDownloaded {
				err := processSinglePeriodData(projectId, periodCode, archiveCloudManager, tmpCloudManager, sortedCloudManager, modelCloudManager, diskManager, deltaQuery, multiStepQuery, k, unionOfFeatures, passId, isEventOccurence, isMultiStep, mailerRun, beamConfig, useBucketV2, hardPull, pulledMap)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func processSinglePeriodData(projectId int64, periodCode Period, archiveCloudManager, tmpCloudManager, sortedCloudManager, modelCloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, deltaQuery Query, multiStepQuery MultiFunnelQuery, k int, unionOfFeatures *(map[string]map[string]bool), passId int, isEventOccurence bool, isMultiStep bool, mailerRun bool, beamConfig *merge.RunBeamConfig, useBucketV2, hardPull bool, pulledMap map[int64]map[string]bool) error {
	scanner, err := GetEventFileScanner(projectId, periodCode, archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, useBucketV2, hardPull, pulledMap)
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Scanner initialization failed for period %v", periodCode))
		return err
	}
	withinPeriodInsights, err := ComputeWithinPeriodInsights(scanner, deltaQuery, multiStepQuery, k, *unionOfFeatures, passId, isEventOccurence, isMultiStep, mailerRun)
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Could not mine features for period %v", periodCode))
		return err
	}
	withinPeriodInsights.Period = Period(periodCode)
	if passId == 1 {
		features := make(map[string]map[string]bool)
		for key, valCounts := range withinPeriodInsights.Base.FeatureMetrics {
			for val := range valCounts {
				if features[key] == nil {
					features[key] = make(map[string]bool)
				}
				features[key][val] = true
			}
		}
		for key, valCounts := range withinPeriodInsights.BaseAndTarget.FeatureMetrics {
			for val := range valCounts {
				if features[key] == nil {
					features[key] = make(map[string]bool)
				}
				features[key][val] = true
			}
		}
		for key, valCounts := range withinPeriodInsights.Target.FeatureMetrics {
			for val := range valCounts {
				if features[key] == nil {
					features[key] = make(map[string]bool)
				}
				features[key][val] = true
			}
		}
		for key, valueStatus := range features {
			for value := range valueStatus {
				if (*unionOfFeatures)[key] == nil {
					(*unionOfFeatures)[key] = make(map[string]bool)
				}
				(*unionOfFeatures)[key][value] = true
			}
		}
	}
	if passId == 2 {
		withinPeriodInsightsBytes, err := json.Marshal(withinPeriodInsights)
		if err != nil {
			log.WithError(err).Error("Failed to marshal wpi")
			return err
		}
		if deltaQuery.Id == 0 {
			deltaQuery.Id = multiStepQuery.Id
		}
		writeWpiPath(projectId, periodCode, deltaQuery.Id, k, bytes.NewReader(withinPeriodInsightsBytes), *modelCloudManager, mailerRun)
		dateString := U.GetDateOnlyFromTimestampZ(periodCode.From)
		efTmpPath, efTmpName := (*diskManager).GetInsightsWpiFilePathAndName(projectId, dateString, deltaQuery.Id, k, mailerRun)
		err = (*diskManager).Create(efTmpPath, efTmpName, bytes.NewReader(withinPeriodInsightsBytes))
		if err != nil {
			log.WithFields(log.Fields{"err": err, "filePath": efTmpPath,
				"eventFileName": efTmpName}).Error("Failed to write to temp path")
			return err
		}
	}
	return nil
}

func writeWpiPath(projectId int64, periodCode Period, queryId int64, k int, reader *bytes.Reader,
	cloudManager filestore.FileManager, mailerRun bool) error {
	dateString := U.GetDateOnlyFromTimestampZ(periodCode.From)
	path, name := cloudManager.GetInsightsWpiFilePathAndName(projectId, dateString, queryId, k, mailerRun)
	err := cloudManager.Create(path, name, reader)
	if err != nil {
		log.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return err
	}
	return err
}

func writeCpiPath(projectId int64, periodCode Period, queryId int64, k int, reader *bytes.Reader,
	cloudManager filestore.FileManager, mailerRun bool) error {
	dateString := U.GetDateOnlyFromTimestampZ(periodCode.From)
	path, name := cloudManager.GetInsightsCpiFilePathAndName(projectId, dateString, queryId, k, mailerRun)
	err := cloudManager.Create(path, name, reader)
	if err != nil {
		log.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return err
	}
	return err
}

func writePropertiesPath(projectId int64, properties map[string]bool, cloudManager filestore.FileManager) error {
	path, name := cloudManager.GetWIPropertiesPathAndName(projectId)
	propertiesJson, err := json.Marshal(properties)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("failed to unmarshal properties Info.")
		return err
	}
	err = cloudManager.Create(path, name, bytes.NewReader(propertiesJson))
	if err != nil {
		log.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return err
	}
	return err
}

func IsDashboardUnitWIEnabled(dashboardUnit M.DashboardUnit) (Query, MultiFunnelQuery, M.KPIQueryGroup, bool, bool, bool, map[string]bool) {
	properties := make(map[string]bool)
	var deltaQuery Query
	queryClass, queryInfo, errMsg := store.GetStore().GetQueryAndClassFromDashboardUnit(&dashboardUnit)
	if errMsg != "" {
		log.Info("errMsg: ", errMsg)
		return Query{}, MultiFunnelQuery{}, M.KPIQueryGroup{}, false, false, false, properties
	}

	if queryClass == M.QueryClassEvents {
		var queryGroup M.QueryGroup
		U.DecodePostgresJsonbToStructType(&queryInfo.Query, &queryGroup)
		for _, query := range queryGroup.Queries {
			for _, qwp := range query.EventsWithProperties {
				for _, qp := range qwp.Properties {
					properties[qp.Property] = true
				}
			}
			for _, qp := range query.GlobalUserProperties {
				properties[qp.Property] = true
			}
			for _, qp := range query.GroupByProperties {
				properties[qp.Property] = true
			}
		}
		query := queryGroup.Queries[0]
		if query.Type == M.QueryTypeUniqueUsers || query.Type == M.QueryTypeEventsOccurrence {
			isEventOccurence := query.Type == M.QueryTypeEventsOccurrence
			if (query.EventsCondition == M.EventCondAnyGivenEvent || query.EventsCondition == M.EventCondAllGivenEvent) || (query.EventsCondition == M.EventCondEachGivenEvent && len(query.EventsWithProperties) == 1) {
				deltaQuery = Query{Id: dashboardUnit.QueryId,
					Base: E.EventsCriteria{
						Operator: "And",
						EventCriterionList: []E.EventCriterion{{
							Name:         "$session",
							EqualityFlag: true,
						}}},
					Target: E.EventsCriteria{
						EventCriterionList: []E.EventCriterion{},
					}}
				for _, event := range query.EventsWithProperties {
					event.Properties = append(event.Properties, query.GlobalUserProperties...)
					if query.EventsCondition == M.EventCondEachGivenEvent {
						deltaQuery.Target.Operator = "And"
						deltaQuery.Target.EventCriterionList = append(deltaQuery.Target.EventCriterionList, E.EventCriterion{
							Name:                event.Name,
							EqualityFlag:        true,
							FilterCriterionList: E.MapFilterProperties(event.Properties),
						})
					}
					if query.EventsCondition == M.EventCondAllGivenEvent {
						deltaQuery.Target.Operator = "And"
						deltaQuery.Target.EventCriterionList = append(deltaQuery.Target.EventCriterionList, E.EventCriterion{
							Name:                event.Name,
							EqualityFlag:        true,
							FilterCriterionList: E.MapFilterProperties(event.Properties),
						})
					}
					if query.EventsCondition == M.EventCondAnyGivenEvent {
						deltaQuery.Target.Operator = "Or"
						deltaQuery.Target.EventCriterionList = append(deltaQuery.Target.EventCriterionList, E.EventCriterion{
							Name:                event.Name,
							EqualityFlag:        true,
							FilterCriterionList: E.MapFilterProperties(event.Properties),
						})
					}
				}
				return deltaQuery, MultiFunnelQuery{}, M.KPIQueryGroup{}, true, isEventOccurence, false, properties
			}
		}
	}
	if queryClass == M.QueryClassFunnel {
		var query M.Query
		U.DecodePostgresJsonbToStructType(&queryInfo.Query, &query)
		for _, qwp := range query.EventsWithProperties {
			for _, qp := range qwp.Properties {
				properties[qp.Property] = true
			}
		}
		for _, qp := range query.GlobalUserProperties {
			properties[qp.Property] = true
		}
		for _, qp := range query.GroupByProperties {
			properties[qp.Property] = true
		}
		if query.Type == M.QueryTypeUniqueUsers {
			if query.EventsCondition == M.EventCondAnyGivenEvent {
				if len(query.EventsWithProperties) == 2 {
					deltaQuery = Query{Id: dashboardUnit.QueryId,
						Base: E.EventsCriteria{
							Operator: "And",
							EventCriterionList: []E.EventCriterion{{
								Name:                query.EventsWithProperties[0].Name,
								EqualityFlag:        true,
								FilterCriterionList: E.MapFilterProperties(append(query.EventsWithProperties[0].Properties, query.GlobalUserProperties...)),
							}}},
						Target: E.EventsCriteria{
							Operator: "And",
							EventCriterionList: []E.EventCriterion{{
								Name:                query.EventsWithProperties[1].Name,
								EqualityFlag:        true,
								FilterCriterionList: E.MapFilterProperties(append(query.EventsWithProperties[1].Properties, query.GlobalUserProperties...)),
							}},
						}}
					return deltaQuery, MultiFunnelQuery{}, M.KPIQueryGroup{}, true, false, false, properties
				} else {
					multiStepFunnel := MultiFunnelQuery{Id: dashboardUnit.QueryId,
						Base: E.EventsCriteria{
							Operator: "And",
							EventCriterionList: []E.EventCriterion{{
								Name:                query.EventsWithProperties[0].Name,
								EqualityFlag:        true,
								FilterCriterionList: E.MapFilterProperties(append(query.EventsWithProperties[0].Properties, query.GlobalUserProperties...)),
							}}},
						Intermediate: make([]E.EventsCriteria, 0),
						Target: E.EventsCriteria{
							Operator: "And",
							EventCriterionList: []E.EventCriterion{{
								Name:                query.EventsWithProperties[len(query.EventsWithProperties)-1].Name,
								EqualityFlag:        true,
								FilterCriterionList: E.MapFilterProperties(append(query.EventsWithProperties[len(query.EventsWithProperties)-1].Properties, query.GlobalUserProperties...)),
							}},
						}}
					for i := 1; i <= len(query.EventsWithProperties)-2; i++ {
						criteria := E.EventsCriteria{
							Operator: "And",
							EventCriterionList: []E.EventCriterion{{
								Name:                query.EventsWithProperties[i].Name,
								EqualityFlag:        true,
								FilterCriterionList: E.MapFilterProperties(append(query.EventsWithProperties[i].Properties, query.GlobalUserProperties...)),
							}},
						}
						multiStepFunnel.Intermediate = append(multiStepFunnel.Intermediate, criteria)
					}
					return Query{}, multiStepFunnel, M.KPIQueryGroup{}, true, false, true, properties
				}

			}
		}
	}

	if queryClass == M.QueryClassKPI {
		var kpiQuery M.KPIQueryGroup
		err := json.Unmarshal(queryInfo.Query.RawMessage, &kpiQuery)
		for _, query := range kpiQuery.Queries {
			for _, filter := range query.Filters {
				properties[filter.PropertyName] = true
			}
			for _, gbp := range query.GroupBy {
				properties[gbp.PropertyName] = true
			}
		}
		for _, filter := range kpiQuery.GlobalFilters {
			properties[filter.PropertyName] = true
		}
		for _, gbp := range kpiQuery.GlobalGroupBy {
			properties[gbp.PropertyName] = true
		}
		if err == nil {
			return deltaQuery, MultiFunnelQuery{}, kpiQuery, true, false, false, properties
		} else {
			log.WithError(err).Error("error getting kpiQuery")
		}
		// query := queryGroup.Queries[0]
		// log.Infof("query: %v", query)
	}
	return deltaQuery, MultiFunnelQuery{}, M.KPIQueryGroup{}, false, false, false, properties
}
