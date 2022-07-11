package memsql

import (
	cacheRedis "factors/cache/redis"
	"factors/model/model"
	U "factors/util"
	"fmt"
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
	result := make(map[string][]*model.ProjectAnalytics, 0)
	for i := 0; i < lastNDays; i++ {
		dateKey := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		if result[dateKey] == nil {
			result[dateKey] = make([]*model.ProjectAnalytics, 0)
		}
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
			result[dateKey] = append(result[dateKey], &model.ProjectAnalytics{
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
			})

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
			log.Error(err)
			return 0, err
		}
		return result, nil
	}
	return 0, nil
}
