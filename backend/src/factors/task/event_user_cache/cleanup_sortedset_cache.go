package event_user_cache

import(
	cacheRedis "factors/cache/redis"
	"factors/model/model"
	log "github.com/sirupsen/logrus"
	U "factors/util"
	"strconv"
)

func DoCleanUpSortedSet(eventsLimit *int, propertiesLimit *int, valuesLimit *int, rollupLookback *int) map[string]interface{} {
	// Get all projects sorted set
	// zcard for all the keys
	// zremrange for [0 - (count - limit)]
	
	currentTimeDatePart := U.TimeNow().Format(U.DATETIME_FORMAT_YYYYMMDD)
	uniqueUsersCountKey, err := model.UserCountAnalyticsCacheKey(
		currentTimeDatePart)
	if err != nil {
		log.WithError(err).Error("Failed to get cache key - uniqueEventsCountKey")
		return nil
	}
	allProjects, _ := cacheRedis.ZrangeWithScoresPersistent(true, uniqueUsersCountKey)
	log.WithField("projects", allProjects).Info("AllProjects")
	for id, _ := range allProjects {
		projId, _ := strconv.Atoi(id)
		projectID := uint64(projId)
		log.WithField("ProjectId", projectID).Info("Starting CLEANUP")
		eventNamesSmartKeySortedSet, err := model.GetSmartEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectID,
			currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - events")
			return nil
		}
		eventNamesKeySortedSet, err := model.GetEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectID,
				currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - smart events")
			return nil
		}
		propertyCategoryKeySortedSet, err := model.GetPropertiesByEventCategoryCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - properties")
			return nil
		}
		valueKeySortedSet, err := model.GetValuesByEventPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - values")
			return nil
		}

		userPropertyCategoryKeySortedSet, err := model.GetUserPropertiesCategoryByProjectCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - property category")
			return nil
		}
		userValueKeySortedSet, err := model.GetValuesByUserPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - values")
			return nil
		}
		count, err := cacheRedis.ZcardPersistent(eventNamesKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - events")
			return nil
		}
		eventCount := int(count)
		count, err = cacheRedis.ZcardPersistent(eventNamesSmartKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - smart events")
			return nil
		}
		smartEventsCount := int(count)
		if(eventCount + smartEventsCount > *eventsLimit && eventCount > (*eventsLimit - eventCount + smartEventsCount -1)){
			log.WithField("ProjectId", projectID).WithField("Count", eventCount + smartEventsCount -1 - *eventsLimit ).Info("Deleting events")
			_, err := cacheRedis.ZRemRangePersistent(eventNamesKeySortedSet, 0, (eventCount + smartEventsCount -1 - *eventsLimit ))
			if err != nil {
				log.WithError(err).Error("Failed to delete - events")
				return nil
			}
		}
		count, err = cacheRedis.ZcardPersistent(propertyCategoryKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - properties")
			return nil
		}
		propCount := int(count)
		if(propCount > *propertiesLimit){
			log.WithField("ProjectId", projectID).WithField("Count", propCount -1 - *propertiesLimit  ).Info("Deleting event properties")
			_, err := cacheRedis.ZRemRangePersistent(propertyCategoryKeySortedSet, 0, (propCount -1 - *propertiesLimit ))
			if err != nil {
				log.WithError(err).Error("Failed to delete - properties")
				return nil
			}
		}
		count, err = cacheRedis.ZcardPersistent(valueKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count- values")
			return nil
		}
		valueCount := int(count)
		if(valueCount > *valuesLimit){
			log.WithField("ProjectId", projectID).WithField("Count", valueCount -1 - *valuesLimit  ).Info("Deleting event properties values")
			_, err := cacheRedis.ZRemRangePersistent(valueKeySortedSet, 0, (valueCount -1 - *valuesLimit))
			if err != nil {
				log.WithError(err).Error("Failed to delete - values")
				return nil
			}
		}
		count, err = cacheRedis.ZcardPersistent(userPropertyCategoryKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - user property")
			return nil
		}
		userPropCount := int(count)
		if(userPropCount > *propertiesLimit){
			log.WithField("ProjectId", projectID).WithField("Count", userPropCount -1 - *propertiesLimit  ).Info("Deleting user properties")
			_, err := cacheRedis.ZRemRangePersistent(userPropertyCategoryKeySortedSet, 0, (userPropCount -1 - *propertiesLimit ))
			if err != nil {
				log.WithError(err).Error("Failed to delete - user property")
				return nil
			}
		}
		count, err = cacheRedis.ZcardPersistent(userValueKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - user property value")
			return nil
		}
		userValueCount := int(count)
		if(userValueCount > *valuesLimit){
			log.WithField("ProjectId", projectID).WithField("Count", userValueCount -1 - *valuesLimit).Info("Deleting user properties values")
			_, err := cacheRedis.ZRemRangePersistent(userValueKeySortedSet, 0, (userValueCount -1 - *valuesLimit ))
			if err != nil {
				log.WithError(err).Error("Failed to delete - user property value")
				return nil
			}
		}
	}
	return nil
}