package memsql

import (
	"encoding/json"
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
	projectIDNameMap := make(map[string]string)
	for _, project := range projects {
		projId, _ := U.GetValueAsString(project.ID)
		projectIDNameMap[projId] = project.Name
	}
	result, err := store.GetProjectAnalyticsData(projectIDNameMap, lastNDays, currentDate, 0)
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
					ProjectID:         projId,
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
					ProjectID:         projId,
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

	projectIDNameMap := make(map[string]string)
	projId, _ := U.GetValueAsString(project.ID)
	projectIDNameMap[projId] = project.Name

	result, err := store.GetProjectAnalyticsData(projectIDNameMap, lastNDays, currentDate, projectId)

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (store *MemSQL) GetProjectAnalyticsData(projectIDNameMap map[string]string, lastNDays int, currentDate time.Time, projectId int64) (map[string][]*model.ProjectAnalytics, error) {

	result := make(map[string][]*model.ProjectAnalytics, 0)

	timezoneString, statusCode := store.GetTimezoneForProject(projectId)
	if statusCode != http.StatusFound {
		log.Errorf("Failed to get project Timezone for %d", projectId)
	}

	from := U.GetBeginningOfDayTimestampIn(currentDate.AddDate(0, 0, -lastNDays+1).Unix(), timezoneString)
	end := U.GetEndOfDayTimestampIn(currentDate.Unix(), timezoneString)

	var err error

	eventData, err := store.GetGlobalProjectAnalyticsEventDataByProjectId(projectId, "daily_login_count", timezoneString, from, end)
	if err != nil {
		return nil, errors.New("failed to event base metrics data")
	}

	if len(eventData) == 0 {
		eventData = make([]map[string]interface{}, lastNDays)
	}

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
			projIdInt64, _ := strconv.ParseInt(projId, 10, 64)
			dateKeyInt, _ := strconv.Atoi(dateKey)
			adwordsEvents, _ := GetEventsFromCacheByDocumentType(projId, "adwords", dateKey)
			facebookEvents, _ := GetEventsFromCacheByDocumentType(projId, "facebook", dateKey)
			hubspotEvents, _ := GetEventsFromCacheByDocumentType(projId, "hubspot", dateKey)
			linkedinEvents, _ := GetEventsFromCacheByDocumentType(projId, "linkedin", dateKey)
			salesforceEvents, _ := GetEventsFromCacheByDocumentType(projId, "salesforce", dateKey)
			sixSignalAPIHits, _ := model.GetFactorsDeanonAPICountResult(projIdInt64, uint64(dateKeyInt))
			sixSignalAPITotalHits, _ := model.GetFactorsDeanonAPITotalHitCountResult(projIdInt64, uint64(dateKeyInt))

			if projectId == 0 {
				result[dateKey] = append(result[dateKey], &model.ProjectAnalytics{
					ProjectID:             projId,
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
					ProjectName:           projectIDNameMap[projId],
					Date:                  dateKey,
				})
			} else if projIdInt64 == projectId {

				entry := model.ProjectAnalytics{
					ProjectID:             projId,
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
					ProjectName:           projectIDNameMap[projId],
					Date:                  dateKey,
					DailyLoginCount:       int64(U.SafeConvertToFloat64(eventData[lastNDays-1-i]["aggregate"])),
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

func (store *MemSQL) GetGlobalProjectAnalyticsDataByProjectId(projectID int64, monthString, agentUUID string) ([]map[string]interface{}, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	result := make([]map[string]interface{}, 0)
	params := make([]interface{}, 0)

	stmt := fmt.Sprintf(` 
	with 
    step_1 as ( select count(*) as users_count from project_agent_mappings where project_id =?),
    step_2 as ( select count(*) as alerts_count from alerts where project_id =? and alert_type != 3 and is_deleted = ? ),
	step_3 as (select count(*) as event_trigger_alerts_count from event_trigger_alerts where project_id =? and internal_status != ? and is_deleted = ?),
    step_4 as ( select count(*) as segment_count_user from segments where project_id =? and JSON_EXTRACT_STRING(query,'%s') = '%s'),
	step_5 as ( select count(*) as segment_count_account from segments where project_id =? and JSON_EXTRACT_STRING(query,'%s') = '%s'),
    step_6 as ( select count(*) as dashboard_count from  dashboards where project_id = ? and is_deleted = ? ),
    step_7 as ( select count(*) as webhooks_count from event_trigger_alerts where project_id = ? and JSON_EXTRACT_STRING(event_trigger_alert,'%s') = ? and internal_status != ? and is_deleted = ? ),
    step_8 as ( select count(*) as report_count from queries where project_id =? and type != 3 and type != 4 and is_deleted = ? ) 
	select * from step_1,step_2,step_3,step_4,step_5,step_6,step_7,step_8;
	`, "caller", "user_profiles", "caller", "account_profiles", model.WEBHOOK)

	params = append(params, projectID, projectID, false, projectID, model.Disabled, false, projectID, projectID, projectID, false, projectID, "true", model.Disabled, false, projectID, false)

	rows, err := db.Raw(stmt, params...).Rows()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var usersCount string
		var alertsCount string
		var peopleSegmentsCount string
		var accountSegmentsCount string
		var dashboardCount string
		var webhooksCount string
		var reportCount string
		var eventAlertsCount string

		if err = rows.Scan(&usersCount, &eventAlertsCount, &alertsCount, &peopleSegmentsCount, &accountSegmentsCount, &dashboardCount,
			&webhooksCount, &reportCount); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return nil, err
		}

		var intgrationCompleted bool
		isExist, _ := store.IsEventExistsWithType(projectID, model.TYPE_AUTO_TRACKED_EVENT_NAME)
		intgrationCompleted = isExist

		timeZoneString, statusCode := store.GetTimezoneForProject(projectID)
		if statusCode != http.StatusFound {
			timeZoneString = U.TimeZoneStringIST
		}

		monthYearString := U.IfThenElse(monthString == "previous", U.GetPreviousMonthYear(timeZoneString), U.GetCurrentMonthYear(timeZoneString))

		identifiedCount, err := model.GetFactorsDeanonMonthlyUniqueEnrichmentCount(projectID, monthYearString.(string))
		if err != nil {
			return nil, errors.New("failed to get six signal count")
		}

		data := map[string]interface{}{
			"user_count":            usersCount,
			"alerts_count":          U.SafeConvertToFloat64(alertsCount) + U.SafeConvertToFloat64(eventAlertsCount),
			"segment_count_user":    peopleSegmentsCount,
			"segment_count_account": accountSegmentsCount,
			"dashboard_count":       dashboardCount,
			"webhooks_count":        webhooksCount,
			"report_count":          reportCount,
			"sdk_int_completed":     intgrationCompleted,
			"identified_count":      identifiedCount,
		}

		startTimestamp := time.Now().AddDate(0, 0, -90)

		for key := range model.ProjectAnalyticsEventSingleQueryStmnt {

			if key == "unique_user_kpi" {
				startTimestamp = time.Now().AddDate(0, 0, -30)
			}

			eventData, err := store.GetGlobalProjectAnalyticsEventDataByProjectId(projectID, key, timeZoneString, startTimestamp.Unix(), time.Now().Unix())
			if len(eventData) == 0 {
				eventData = make([]map[string]interface{}, 2)
			}

			if err != nil {
				return nil, errors.New("failed to event base metrics data")
			}

			val, _ := U.GetValueAsString(eventData[0]["aggregate"])
			data[key], _ = strconv.ParseInt(val, 10, 64)

		}

		sixSignalLimit, err := store.GetFeatureLimitForProject(projectID, FEATURE_FACTORS_DEANONYMISATION)
		if err != nil {
			log.Error("Failed fetching sixsignal quota limit with error ", err)
			sixSignalLimit = 1
		}
		if sixSignalLimit == 0 {
			sixSignalLimit = 1
		}

		data["account_potential"] = U.SafeConvertToFloat64(data["unique_user_kpi"]) * 0.4
		data["account_limit"] = float64(sixSignalLimit)
		data["upsell_potential"] = (U.SafeConvertToFloat64(data["unique_user_kpi"]) * 0.4) / float64(sixSignalLimit)
		result = append(result, data)

	}

	return result, nil
}

func (store *MemSQL) GetGlobalProjectAnalyticsEventDataByProjectId(projectID int64, queryStmntKey string, timeZoneString U.TimeZoneString, startTimestmap, endTimestamp int64) ([]map[string]interface{}, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var query model.Query
	var jsonString string
	var err error

	if queryStmntKey == "daily_login_count" || queryStmntKey == "login_count" {
		jsonString = fmt.Sprintf(model.ProjectAnalyticsEventSingleQueryStmnt[queryStmntKey], projectID, timeZoneString, startTimestmap, endTimestamp)
		err = json.Unmarshal([]byte(jsonString), &query)
		if err != nil {
			return nil, err
		}
		projectID = int64(2)
	} else {
		jsonString = fmt.Sprintf(model.ProjectAnalyticsEventSingleQueryStmnt[queryStmntKey], timeZoneString, startTimestmap, endTimestamp)
		err = json.Unmarshal([]byte(jsonString), &query)
		if err != nil {
			return nil, err
		}
	}

	singleResult, errCode, _ := store.ExecuteEventsQuery(projectID, query, true)

	if errCode != http.StatusOK {
		return nil, nil
	}

	result := make([]map[string]interface{}, 0)
	for _, row := range singleResult.Rows {
		val := make(map[string]interface{})
		for i, key := range singleResult.Headers {
			val[key] = row[i]
		}
		result = append(result, val)
	}

	return result, nil
}

func (store *MemSQL) GetIntegrationStatusesCount(settings model.ProjectSetting, projectID int64, agentUUID string) (map[string]interface{}, map[string]interface{}) {

	connected := make([]string, 0)
	disconnected := make([]string, 0)
	if *settings.IntSegment {
		connected = append(connected, "Segment")
	} else {
		disconnected = append(disconnected, "Segment")
	}
	if *settings.IntDrift {
		connected = append(connected, "Drift")
	} else {
		disconnected = append(disconnected, "Drift")
	}
	if *settings.IntRudderstack {
		connected = append(connected, "Rudderstack")
	} else {
		disconnected = append(disconnected, "Rudderstack")
	}
	if *settings.IntClientSixSignalKey {
		connected = append(connected, "Client 6Signal")
	} else {
		disconnected = append(disconnected, "Client 6Signal")
	}
	if *settings.IntFactorsSixSignalKey {
		connected = append(connected, "Factors 6Signal")
	} else {
		disconnected = append(disconnected, "Factors 6Signal")
	}
	if store.IsHubspotIntegrationAvailable(projectID) {
		connected = append(connected, "Hubspot")
	} else {
		disconnected = append(disconnected, "Hubspot")
	}
	if store.IsSalesforceIntegrationAvailable(projectID) {
		connected = append(connected, "Salesforce")
	} else {
		disconnected = append(disconnected, "Salesforce")
	}
	if store.IsBingIntegrationAvailable(projectID) {
		connected = append(connected, "Bing")
	} else {
		disconnected = append(disconnected, "Bing")
	}
	if store.IsAdwordsIntegrationAvailable(projectID) {
		connected = append(connected, "Adwords")
	} else {
		disconnected = append(disconnected, "Adwords")
	}
	if store.IsFacebookIntegrationAvailable(projectID) {
		connected = append(connected, "Facebook")
	} else {
		disconnected = append(disconnected, "Facebook")
	}
	if store.IsLinkedInIntegrationAvailable(projectID) {
		connected = append(connected, "Linkedin")
	} else {
		disconnected = append(disconnected, "Linkedin")
	}

	if store.IsGoogleOrganicIntegrationAvailable(projectID) {
		connected = append(connected, "Google Search Console")
	} else {
		disconnected = append(disconnected, "Google Search Console")
	}

	if store.IsMarketoIntegrationAvailable(projectID) {
		connected = append(connected, "Marketo")
	} else {
		disconnected = append(disconnected, "Marketo")
	}

	if store.IsG2IntegrationAvailable(projectID) {
		connected = append(connected, "G2")
	} else {
		disconnected = append(disconnected, "G2")
	}

	if ok, _ := store.IsClearbitIntegratedByProjectID(projectID); ok {
		connected = append(connected, "Clearbit")
	} else {
		disconnected = append(disconnected, "Clearbit")
	}
	if store.IsLeadSquaredIntegrationAvailble(projectID) {
		connected = append(connected, "Lead Squared")
	} else {
		disconnected = append(disconnected, "Lead Squared")
	}
	if ok, _ := store.IsSlackIntegratedForProject(projectID, agentUUID); ok {
		connected = append(connected, "Slack")
	} else {
		disconnected = append(disconnected, "Slack")
	}
	if ok, _ := store.IsTeamsIntegrated(projectID); ok {
		connected = append(connected, "Teams")
	} else {
		disconnected = append(disconnected, "Teams")
	}

	return map[string]interface{}{"connected": connected}, map[string]interface{}{"disconnected": disconnected}
}
