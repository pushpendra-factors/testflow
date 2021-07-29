package delta

import (
	"bytes"
	"encoding/json"
	"factors/filestore"
	serviceDisk "factors/services/disk"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	U "factors/util"

	"factors/model/model"
	"factors/model/store"

	M "factors/model/model"
)

var isDownloaded bool

// ComputeDeltaInsights Take details about two periods, and get delta (difference based) insights
func ComputeDeltaInsights(projectId uint64, configs map[string]interface{}) (map[string]interface{}, bool) {

	status := make(map[string]interface{})
	var insightId uint64
	if configs["endTimestamp"].(int64) > U.TimeNowUnix() {
		status["error"] = "invalid end timestamp"
		return status, false
	}
	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	cloudManager := configs["cloudManager"].(*filestore.FileManager)
	log.Info((*cloudManager).GetBucketName())
	k := configs["k"].(int)
	// Cross Period Insights
	periodCodesWithWeekNMinus1 := []Period{
		Period{
			From: configs["startTimestamp"].(int64) - 604800,
			To:   configs["endTimestamp"].(int64) - 604800,
		},
		Period{
			From: configs["startTimestamp"].(int64),
			To:   configs["endTimestamp"].(int64),
		},
	}
	insightGranularity := configs["insightGranularity"].(string)
	log.Info("Reading delta query.")
	computedQueries := make(map[uint64]bool)
	dashboardUnits, _ := store.GetStore().GetDashboardUnitsForProjectID(projectId)
	for _, dashboardUnit := range dashboardUnits {
		queryIdString := fmt.Sprintf("%v", dashboardUnit.QueryId)
		deltaQuery, isEnabled, isEventOccurence := IsDashboardUnitWIEnabled(dashboardUnit)
		// Check if this is a valid query with valid filters
		if isEnabled == false || computedQueries[dashboardUnit.QueryId] == true {
			continue
		}
		// Within Period Insights
		// TODO: $others values are not getting propagated to the second pass. Do that.
		// TODO: This was changed from set to map
		unionOfFeatures := make(map[string]map[string]bool)
		log.Info("1st pass: Scanning events file to get top-k base features for each period.")
		err := processSeparatePeriods(projectId, periodCodesWithWeekNMinus1, cloudManager, diskManager, deltaQuery, k, &unionOfFeatures, 1, insightGranularity, isEventOccurence)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to process wpi pass 1"))
			status["error-wpi-pass1-"+queryIdString] = err.Error()
			continue
		}
		isDownloaded = true
		log.Info("2nd pass: Scanning events file again to compute counts for union of features.")
		err = processSeparatePeriods(projectId, periodCodesWithWeekNMinus1, cloudManager, diskManager, deltaQuery, k, &unionOfFeatures, 2, insightGranularity, isEventOccurence)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to process wpi pass 2"))
			status["error-wpi-pass2-"+queryIdString] = err.Error()
			continue
		}
		log.Info("Computing cross-period insights.")
		err = processCrossPeriods(periodCodesWithWeekNMinus1, diskManager, projectId, k, deltaQuery.Id, cloudManager)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to process wpi pass 1"))
			status["error-cpi-pass1-"+queryIdString] = err.Error()
			continue
		}
		insightId = uint64(U.TimeNowUnix())
		computedQueries[dashboardUnit.QueryId] = true
		errCode, errMsg := store.GetStore().CreateWeeklyInsightsMetadata(&model.WeeklyInsightsMetadata{
			ProjectId:           projectId,
			QueryId:             dashboardUnit.QueryId,
			BaseStartTime:       periodCodesWithWeekNMinus1[1].From,
			BaseEndTime:         periodCodesWithWeekNMinus1[1].To,
			ComparisonStartTime: periodCodesWithWeekNMinus1[0].From,
			ComparisonEndTime:   periodCodesWithWeekNMinus1[0].To,
			InsightType:         "w",
			InsightId:           insightId,
			CreatedAt:           U.TimeNow(),
			UpdatedAt:           U.TimeNow(),
		})
		if errCode != http.StatusCreated {
			log.Error(errMsg)
			status["error"] = errMsg
			return status, false
		}
	}
	status["InsightId"] = insightId
	return status, true
}

func processCrossPeriods(periodCodes []Period, diskManager *serviceDisk.DiskDriver, projectId uint64, k int, queryId int, cloudManager *filestore.FileManager) error {
	for i, periodCode1 := range periodCodes {
		var wpi1 WithinPeriodInsights
		dateString1 := U.GetDateOnlyFromTimestamp(periodCode1.From)
		efTmpPath1, efTmpName1 := diskManager.GetInsightsWpiFilePathAndName(projectId, dateString1, uint64(queryId), k)
		ReadFromJSONFile(efTmpPath1+efTmpName1, &wpi1)
		for j, periodCode2 := range periodCodes {
			if i >= j {
				continue
			}
			var wpi2 WithinPeriodInsights
			dateString2 := U.GetDateOnlyFromTimestamp(periodCode2.From)
			efTmpPath2, efTmpName2 := diskManager.GetInsightsWpiFilePathAndName(projectId, dateString2, uint64(queryId), k)

			err := ReadFromJSONFile(efTmpPath2+efTmpName2, &wpi2)
			if err != nil {
				log.WithError(err).Error(fmt.Sprintf("Failed to read json file "))
				return err
			}
			crossPeriodInsights, err := ComputeCrossPeriodInsights(wpi1, wpi2)
			if err != nil {
				log.WithError(err).Error(fmt.Sprintf("Could not compute cross insights for periods %v VS %v", periodCode1, periodCode2))
				return err
			}
			periodPair := PeriodPair{First: periodCode1, Second: periodCode2}
			crossPeriodInsights.Periods = periodPair
			crossPeriodInsightsBytes, err := json.Marshal(crossPeriodInsights)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Failed to unmarshal events Info.")
				return err
			}
			err = writeCpiPath(projectId, periodPair.Second, uint64(queryId), k, bytes.NewReader(crossPeriodInsightsBytes), *cloudManager)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Failed to write files to cloud")
				return err
			}
		}
	}
	return nil
}

func processSeparatePeriods(projectId uint64, periodCodes []Period, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, deltaQuery Query, k int, unionOfFeatures *(map[string]map[string]bool), passId int, insightGranularity string, isEventOccurence bool) error {
	for _, periodCode := range periodCodes {
		err := processSinglePeriodData(projectId, periodCode, cloudManager, diskManager, deltaQuery, k, unionOfFeatures, passId, insightGranularity, isEventOccurence)
		if err != nil {
			return err
		}
	}
	return nil
}

func processSinglePeriodData(projectId uint64, periodCode Period, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, deltaQuery Query, k int, unionOfFeatures *(map[string]map[string]bool), passId int, insightGranularity string, isEventOccurence bool) error {
	scanner, err := GetEventFileScanner(projectId, periodCode, cloudManager, diskManager, insightGranularity, isDownloaded)
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Scanner initialization failed for period %v", periodCode))
		return err
	}
	withinPeriodInsights, err := ComputeWithinPeriodInsights(scanner, deltaQuery, k, *unionOfFeatures, passId, isEventOccurence)
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Could not mine features for period ", periodCode))
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
			for value, _ := range valueStatus {
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
			log.WithError(err).Error("Failed to unmarshal events Info.")
			return err
		}
		writeWpiPath(projectId, periodCode, uint64(deltaQuery.Id), k, bytes.NewReader(withinPeriodInsightsBytes), *cloudManager)
		dateString := U.GetDateOnlyFromTimestamp(periodCode.From)
		efTmpPath, efTmpName := diskManager.GetInsightsWpiFilePathAndName(projectId, dateString, uint64(deltaQuery.Id), k)
		err = diskManager.Create(efTmpPath, efTmpName, bytes.NewReader(withinPeriodInsightsBytes))
		if err != nil {
			log.WithFields(log.Fields{"err": err, "filePath": efTmpPath,
				"eventFileName": efTmpName}).Error("Failed to write to temp path")
			return err
		}
	}
	return nil
}

func writeWpiPath(projectId uint64, periodCode Period, queryId uint64, k int, events *bytes.Reader,
	cloudManager filestore.FileManager) error {
	dateString := U.GetDateOnlyFromTimestamp(periodCode.From)
	path, name := cloudManager.GetInsightsWpiFilePathAndName(projectId, dateString, queryId, k)
	err := cloudManager.Create(path, name, events)
	if err != nil {
		log.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return err
	}
	return err
}

func writeCpiPath(projectId uint64, periodCode Period, queryId uint64, k int, events *bytes.Reader,
	cloudManager filestore.FileManager) error {
	dateString := U.GetDateOnlyFromTimestamp(periodCode.From)
	path, name := cloudManager.GetInsightsCpiFilePathAndName(projectId, dateString, queryId, k)
	err := cloudManager.Create(path, name, events)
	if err != nil {
		log.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return err
	}
	return err
}

func IsDashboardUnitWIEnabled(dashboardUnit M.DashboardUnit) (Query, bool, bool) {
	var deltaQuery Query
	queryClass, _ := store.GetStore().GetQueryAndClassFromDashboardUnit(&dashboardUnit)
	if queryClass == model.QueryClassEvents {
		var queryGroup M.QueryGroup
		U.DecodePostgresJsonbToStructType(&dashboardUnit.Query, &queryGroup)
		query := queryGroup.Queries[0]
		if query.Type == model.QueryTypeUniqueUsers || query.Type == model.QueryTypeEventsOccurrence {
			isEventOccurence := query.Type == model.QueryTypeEventsOccurrence
			if (query.EventsCondition == model.EventCondAnyGivenEvent || query.EventsCondition == model.EventCondAllGivenEvent) || (query.EventsCondition == model.EventCondEachGivenEvent && len(query.EventsWithProperties) == 1) {
				deltaQuery = Query{Id: int(dashboardUnit.QueryId),
					Base: EventsCriteria{
						Operator: "And",
						EventCriterionList: []EventCriterion{EventCriterion{
							Name:         "$session",
							EqualityFlag: true,
						}}},
					Target: EventsCriteria{
						EventCriterionList: []EventCriterion{},
					}}
				for _, event := range query.EventsWithProperties {
					if query.EventsCondition == model.EventCondEachGivenEvent {
						deltaQuery.Target.Operator = "And"
						deltaQuery.Target.EventCriterionList = append(deltaQuery.Target.EventCriterionList, EventCriterion{
							Name:                event.Name,
							EqualityFlag:        true,
							FilterCriterionList: MapFilterProperties(event.Properties),
						})
					}
					if query.EventsCondition == model.EventCondAllGivenEvent {
						deltaQuery.Target.Operator = "And"
						deltaQuery.Target.EventCriterionList = append(deltaQuery.Target.EventCriterionList, EventCriterion{
							Name:                event.Name,
							EqualityFlag:        true,
							FilterCriterionList: MapFilterProperties(event.Properties),
						})
					}
					if query.EventsCondition == model.EventCondAnyGivenEvent {
						deltaQuery.Target.Operator = "Or"
						deltaQuery.Target.EventCriterionList = append(deltaQuery.Target.EventCriterionList, EventCriterion{
							Name:                event.Name,
							EqualityFlag:        true,
							FilterCriterionList: MapFilterProperties(event.Properties),
						})
					}
				}
				return deltaQuery, true, isEventOccurence
			}
		}
	}
	if queryClass == model.QueryClassFunnel {
		var query M.Query
		U.DecodePostgresJsonbToStructType(&dashboardUnit.Query, &query)
		if query.Type == model.QueryTypeUniqueUsers {
			if query.EventsCondition == model.EventCondAnyGivenEvent {
				if len(query.EventsWithProperties) == 2 {
					deltaQuery = Query{Id: int(dashboardUnit.QueryId),
						Base: EventsCriteria{
							Operator: "And",
							EventCriterionList: []EventCriterion{EventCriterion{
								Name:                query.EventsWithProperties[0].Name,
								EqualityFlag:        true,
								FilterCriterionList: MapFilterProperties(query.EventsWithProperties[0].Properties),
							}}},
						Target: EventsCriteria{
							Operator: "And",
							EventCriterionList: []EventCriterion{EventCriterion{
								Name:                query.EventsWithProperties[1].Name,
								EqualityFlag:        true,
								FilterCriterionList: MapFilterProperties(query.EventsWithProperties[1].Properties),
							}},
						}}
					return deltaQuery, true, false
				}
			}
		}
	}
	return deltaQuery, false, false
}

func MapFilterProperties(qp []model.QueryProperty) []EventFilterCriterion {
	filters := make(map[string]EventFilterCriterion)
	for _, prop := range qp {
		filterProp := EventFilterCriterion{
			Key: prop.Property,
		}
		filterProp.Type = prop.Type
		if prop.Entity == "user" {
			filterProp.PropertiesMode = "user"
		} else if prop.Entity == "event" {
			filterProp.PropertiesMode = "event"
		} else {
			log.Error("Incorrect entity type")
			return nil
		}
		keyString := fmt.Sprintf("%s-%s", prop.Entity, prop.Property)
		propertyInMap, exists := filters[keyString]
		var values []OperatorValueTuple
		if exists == false {
			values = make([]OperatorValueTuple, 0)
		} else {
			values = propertyInMap.Values
		}
		values = append(values, OperatorValueTuple{
			Operator:  prop.Operator,
			Value:     prop.Value,
			LogicalOp: prop.LogicalOp,
		})
		filterProp.Values = values
		filters[keyString] = filterProp
	}
	criterias := make([]EventFilterCriterion, 0)
	for _, criteria := range filters {
		criterias = append(criterias, criteria)
	}
	return criterias
}
