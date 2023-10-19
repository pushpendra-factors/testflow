package memsql

import (
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetEventUserCountsOfAllProjects(lastNDays int) (map[string][]*model.ProjectAnalytics, error) {
	logFields := log.Fields{
		"last_n_days": lastNDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	currentDate := time.Now().UTC()
	projects, _ := store.GetProjects()
	projectIDNameMap := make(map[int64]string)
	for _, project := range projects {
		projectIDNameMap[project.ID] = project.Name
	}
	result, err := GetProjectAnalyticsData(projectIDNameMap, lastNDays, currentDate, 0)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (store *MemSQL) GetEventUserCountsMerged(projectIdsList []int64, lastNDays int, currentDate time.Time) (map[int64]*model.ProjectAnalytics, error) {
	logFields := log.Fields{
		"last_n_days": lastNDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	//TODO: modify this date to first day of the week
	projectIDNameMap := make(map[int64]string)
	for _, project := range projectIdsList {
		projectIDNameMap[project] = "" // TODO : add project name here
	}
	result := make(map[int64]*model.ProjectAnalytics, 0)
	for i := 0; i < lastNDays; i++ {
		dateKey := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		// if result[dateKey] == nil {
		// 	result[dateKey] = make([]*model.ProjectAnalytics, 0)
		// }
		totalUniqueUsersKey, err := model.UserCountAnalyticsCacheKey(dateKey)
		if err != nil {
			return nil, err
		}
		users, err := cacheRedis.ZrangeWithScoresPersistent(true, totalUniqueUsersKey)
		if err != nil {
			return nil, err
		}
		totalUniqueEventsKey, err := model.UniqueEventNamesAnalyticsCacheKey(dateKey)
		if err != nil {
			return nil, err
		}
		uniqueEvents, err := cacheRedis.ZrangeWithScoresPersistent(true, totalUniqueEventsKey)
		if err != nil {
			return nil, err
		}
		totalEventsKey, err := model.EventsCountAnalyticsCacheKey(dateKey)
		if err != nil {
			return nil, err
		}
		totalEvents, err := cacheRedis.ZrangeWithScoresPersistent(true, totalEventsKey)
		if err != nil {
			return nil, err
		}
		for projId, count := range users {
			uniqueUsers, _ := strconv.Atoi(count)
			totalEvents, _ := strconv.Atoi(totalEvents[projId])
			uniqueEvents, _ := strconv.Atoi(uniqueEvents[projId])
			projIdInt, _ := strconv.Atoi(projId)
			adwordsEvents, _ := GetEventsFromCacheByDocumentType(projId, "adwords", dateKey)
			facebookEvents, _ := GetEventsFromCacheByDocumentType(projId, "facebook", dateKey)
			hubspotEvents, _ := GetEventsFromCacheByDocumentType(projId, "hubspot", dateKey)
			linkedinEvents, _ := GetEventsFromCacheByDocumentType(projId, "linkedin", dateKey)
			salesforceEvents, _ := GetEventsFromCacheByDocumentType(projId, "salesforce", dateKey)
			if result[int64(projIdInt)] == nil {
				firstEntry := model.ProjectAnalytics{
					ProjectID:         int64(projIdInt),
					TotalEvents:       uint64(totalEvents),
					TotalUniqueEvents: uint64(uniqueEvents),
					TotalUniqueUsers:  uint64(uniqueUsers),
					AdwordsEvents:     uint64(adwordsEvents),
					FacebookEvents:    uint64(facebookEvents),
					HubspotEvents:     uint64(hubspotEvents),
					LinkedinEvents:    uint64(linkedinEvents),
					SalesforceEvents:  uint64(salesforceEvents),
					ProjectName:       projectIDNameMap[int64(projIdInt)],
				}
				result[int64(projIdInt)] = &firstEntry
			} else {
				old := result[int64(projIdInt)]
				new := model.ProjectAnalytics{
					ProjectID:         int64(projIdInt),
					TotalEvents:       old.TotalEvents + uint64(totalEvents),
					TotalUniqueEvents: uint64(uniqueEvents),
					TotalUniqueUsers:  uint64(uniqueUsers),
					AdwordsEvents:     uint64(adwordsEvents),
					FacebookEvents:    uint64(facebookEvents),
					HubspotEvents:     uint64(hubspotEvents),
					LinkedinEvents:    uint64(linkedinEvents),
					SalesforceEvents:  uint64(salesforceEvents),
					ProjectName:       projectIDNameMap[int64(projIdInt)],
					Date:              dateKey,
				}
				result[int64(projIdInt)] = &new
			}

		}

	}

	return result, nil
}

func (store *MemSQL) GetEventUserCountsByProjectID(projectId int64, lastNDays int) (map[string][]*model.ProjectAnalytics, error) {
	logFields := log.Fields{
		"last_n_days": lastNDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	currentDate := time.Now().UTC()
	project, _ := store.GetProject(projectId)

	projectIDNameMap := make(map[int64]string)
	projectIDNameMap[project.ID] = project.Name

	result, err := GetProjectAnalyticsData(projectIDNameMap, lastNDays, currentDate, projectId)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetProjectAnalyticsData(projectIDNameMap map[int64]string, lastNDays int, currentDate time.Time, projectId int64) (map[string][]*model.ProjectAnalytics, error) {

	result := make(map[string][]*model.ProjectAnalytics, 0)

	for i := 0; i < lastNDays; i++ {
		dateKey := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)

		totalUniqueUsersKey, err := model.UserCountAnalyticsCacheKey(dateKey)
		if err != nil {
			return nil, err
		}
		users, err := cacheRedis.ZrangeWithScoresPersistent(true, totalUniqueUsersKey)
		if err != nil {
			return nil, err
		}
		totalUniqueEventsKey, err := model.UniqueEventNamesAnalyticsCacheKey(dateKey)
		if err != nil {
			return nil, err
		}
		uniqueEvents, err := cacheRedis.ZrangeWithScoresPersistent(true, totalUniqueEventsKey)
		if err != nil {
			return nil, err
		}
		totalEventsKey, err := model.EventsCountAnalyticsCacheKey(dateKey)
		if err != nil {
			return nil, err
		}
		totalEvents, err := cacheRedis.ZrangeWithScoresPersistent(true, totalEventsKey)
		if err != nil {
			return nil, err
		}
		for projId, count := range users {
			uniqueUsers, _ := strconv.Atoi(count)
			totalEvents, _ := strconv.Atoi(totalEvents[projId])
			uniqueEvents, _ := strconv.Atoi(uniqueEvents[projId])
			projIdInt, _ := strconv.Atoi(projId)
			dateKeyInt, _ := strconv.Atoi(dateKey)
			adwordsEvents, _ := GetEventsFromCacheByDocumentType(projId, "adwords", dateKey)
			facebookEvents, _ := GetEventsFromCacheByDocumentType(projId, "facebook", dateKey)
			hubspotEvents, _ := GetEventsFromCacheByDocumentType(projId, "hubspot", dateKey)
			linkedinEvents, _ := GetEventsFromCacheByDocumentType(projId, "linkedin", dateKey)
			salesforceEvents, _ := GetEventsFromCacheByDocumentType(projId, "salesforce", dateKey)
			sixSignalAPIHits := model.GetSixSignalAPICountCacheResult(int64(projIdInt), uint64(dateKeyInt))
			sixSignalAPITotalHits := model.GetSixSignalAPITotalHitCountCacheResult(int64(projIdInt), uint64(dateKeyInt))

			if projectId == 0 {
				result[dateKey] = append(result[dateKey], &model.ProjectAnalytics{
					ProjectID:             int64(projIdInt),
					TotalEvents:           uint64(totalEvents),
					TotalUniqueEvents:     uint64(uniqueEvents),
					TotalUniqueUsers:      uint64(uniqueUsers),
					AdwordsEvents:         uint64(adwordsEvents),
					FacebookEvents:        uint64(facebookEvents),
					HubspotEvents:         uint64(hubspotEvents),
					LinkedinEvents:        uint64(linkedinEvents),
					SalesforceEvents:      uint64(salesforceEvents),
					SixSignalAPIHits:      uint64(sixSignalAPIHits),
					SixSignalAPITotalHits: uint64(sixSignalAPITotalHits),
					ProjectName:           projectIDNameMap[int64(projIdInt)],
					Date:                  dateKey,
				})
			} else if int64(projIdInt) == projectId {

				entry := model.ProjectAnalytics{
					ProjectID:         int64(projIdInt),
					TotalEvents:       uint64(totalEvents),
					TotalUniqueEvents: uint64(uniqueEvents),
					TotalUniqueUsers:  uint64(uniqueUsers),
					AdwordsEvents:     uint64(adwordsEvents),
					FacebookEvents:    uint64(facebookEvents),
					HubspotEvents:     uint64(hubspotEvents),
					LinkedinEvents:    uint64(linkedinEvents),
					SalesforceEvents:  uint64(salesforceEvents),
					ProjectName:       projectIDNameMap[int64(projIdInt)],
					Date:              dateKey,
				}
				result[projId] = append(result[projId], &entry)

			}
		}

	}

	return result, nil
}

func UpdateCountCacheByDocumentType(projectID int64, time *time.Time, documentType string) (status bool) {
	logFields := log.Fields{
		"project_id":    projectID,
		"time":          time,
		"document_type": documentType,
	}
	logCtx := log.WithFields(logFields)
	timeDatePart := time.Format(U.DATETIME_FORMAT_YYYYMMDD)
	key, err := model.EventCountKeyByDocumentType(documentType, timeDatePart)
	if err != nil {
		return false
	}
	keysToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	keysToIncrSortedSet = append(keysToIncrSortedSet, cacheRedis.SortedSetKeyValueTuple{
		Key:   key,
		Value: fmt.Sprintf("%v", projectID)})
	if len(keysToIncrSortedSet) > 0 {
		_, err = cacheRedis.ZincrPersistentBatch(true, keysToIncrSortedSet...)
		if err != nil {
			logCtx.WithError(err).Error("Failed to increment keys")
			return false
		}
	}
	return true
}
func GetEventsFromCacheByDocumentType(projectID, documentType, dateKey string) (documentEvents uint64, err error) {
	logFields := log.Fields{
		"project_id":    projectID,
		"document_type": documentType,
		"date_key":      dateKey,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	typeEvents, err := model.EventCountKeyByDocumentType(documentType, dateKey)
	logCtx := log.WithFields(logFields)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch keys")
		return 0, err
	}
	events, err := cacheRedis.ZrangeWithScoresPersistent(true, typeEvents)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch data from keys")
		return 0, err
	}
	if events[projectID] != "" {
		result, err := strconv.ParseUint(events[projectID], 10, 64)
		if err != nil {
			log.WithError(err).Error("Failure in GetEventsFromCacheByDocumentType")
			return 0, err
		}
		return result, nil
	}
	return 0, nil
}

func (store *MemSQL) GetGlobalProjectAnalyticsDataByProjectId(projectID int64, monthString string) ([]map[string]interface{}, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	result := make([]map[string]interface{}, 0)
	params := make([]interface{}, 0)

	stmt := fmt.Sprintf(` 
	with 
    step_1 as ( select count(*) as users_count from users where project_id =?),
    step_2 as ( select count(*) as alerts_count from alerts where project_id =? and is_deleted = ? ),
    step_3 as ( select count(*) as segments_count from segments where project_id =? ),
    step_4 as ( select count(*) as dashboard_count from  dashboards where project_id = ? and is_deleted = ? ),
    step_5 as ( select count(*) as webhooks_count from event_trigger_alerts where project_id = ? and JSON_EXTRACT_STRING(event_trigger_alert,'%s') = ? and is_deleted = ? and internal_status = ? ),
    step_6 as ( select count(*) as report_count from queries where project_id =? and is_deleted = ? ),  
    select * from step_1,step_2,step_3,step_4,step_5,step_6;
	`, model.WEBHOOK)

	params = append(params, projectID, projectID, false, projectID, projectID, false, projectID, true, false, model.ACTIVE, projectID, false)

	rows, err := db.Raw(stmt, params).Rows()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var usersCount string
		var alertsCount string
		var segmentsCount string
		var dashboardCount string
		var webhooksCount string
		var reportCount string

		if err = rows.Scan(&usersCount, &alertsCount, &segmentsCount, &dashboardCount,
			&webhooksCount, &reportCount); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return nil, err
		}

		var intgrationCompleted bool
		isExist, _ := store.IsEventExistsWithType(projectID, model.TYPE_AUTO_TRACKED_EVENT_NAME)
		intgrationCompleted = isExist

		var settings *model.ProjectSetting
		var errCode int
		settings, errCode = store.GetProjectSetting(projectID)
		if errCode != http.StatusFound {
			return nil, err
		}

		integrationConnectedCount, integrationDisconnectedCount, integrationErrCount := getIntegrationStatusesCount(*settings)

		// metering logic here
		timeZoneString, statusCode := store.GetTimezoneForProject(projectID)
		if statusCode != http.StatusFound {
			timeZoneString = U.TimeZoneStringIST
		}

		monthYearString := U.IfThenElse(monthString == "previous", U.GetPreviousMonthYear(timeZoneString), U.GetCurrentMonthYear(timeZoneString))

		identifiedCount, err := model.GetSixSignalMonthlyUniqueEnrichmentCount(projectID, monthYearString.(string))
		if err != nil {
			return nil, errors.New("failed to get six signal count")
		}

		data := map[string]interface{}{
			"user_count":               usersCount,
			"alerts_count":             alertsCount,
			"segments_count":           segmentsCount,
			"dashboard_count":          dashboardCount,
			"webhooks_count":           webhooksCount,
			"report_count":             reportCount,
			"sdk_int_completed":        intgrationCompleted,
			"identified_count":         identifiedCount,
			"integration_connected":    integrationConnectedCount,
			"integration_disconnected": integrationDisconnectedCount,
			"integration_err":          integrationErrCount,
		}

		result = append(result, data)

	}

	return result, nil
}

func getIntegrationStatusesCount(settings model.ProjectSetting) (int, int, int) {
	var connected, disconnected, error int

	if *settings.IntHubspot && settings.IntHubspotApiKey != "" {
		connected++
	} else if *settings.IntHubspot && settings.IntHubspotApiKey == "" {
		error++
	} else {
		disconnected++
	}

	if *settings.IntSegment {
		connected++
	} else {
		disconnected++
	}

	if *settings.IntDrift {
		connected++
	} else {
		disconnected++
	}

	if *settings.IntClearBit && settings.ClearbitKey != "" {
		connected++
	} else if *settings.IntHubspot && settings.ClearbitKey == "" {
		error++
	} else {
		disconnected++
	}

	if *settings.IntRudderstack {
		connected++
	} else {
		disconnected++
	}

	if *settings.IntG2 && settings.IntG2ApiKey != "" {
		connected++
	} else if *settings.IntG2 && settings.IntG2ApiKey == "" {
		error++
	} else {
		disconnected++
	}

	if (*settings.IntClientSixSignalKey && settings.Client6SignalKey != "") ||
		(*settings.IntFactorsSixSignalKey && settings.Factors6SignalKey != "") {
		connected++
	} else if (*settings.IntClientSixSignalKey && settings.Client6SignalKey == "") ||
		(*settings.IntFactorsSixSignalKey && settings.Factors6SignalKey == "") {

		error++
	} else {
		disconnected++
	}

	return connected, disconnected, error
}
