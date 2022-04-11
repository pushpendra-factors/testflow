package event_user_cache

import (
	cacheRedis "factors/cache/redis"
	"factors/model/model"
	U "factors/util"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func DoCleanUpSortedSet(configs map[string]interface{}) (map[string]interface{}, bool) {
	// Get all projects sorted set
	// zcard for all the keys
	// zremrange for [0 - (count - limit)]

	eventsLimit := configs["eventsLimit"].(int)
	propertiesLimit := configs["propertiesLimit"].(int)
	valuesLimit := configs["valuesLimit"].(int)

	currentTimeDatePart := U.TimeNowZ().Format(U.DATETIME_FORMAT_YYYYMMDD)
	uniqueUsersCountKey, err := model.UserCountAnalyticsCacheKey(
		currentTimeDatePart)
	if err != nil {
		log.WithError(err).Error("Failed to get cache key - uniqueEventsCountKey")
		return nil, false
	}
	allProjects, err := cacheRedis.ZrangeWithScoresPersistent(true, uniqueUsersCountKey)
	if err != nil {
		log.WithError(err).Error("Failed to get projects")
		return nil, false
	}
	log.WithField("projects", allProjects).Info("AllProjects")
	for id, _ := range allProjects {
		projId, _ := strconv.Atoi(id)
		projectID := uint64(projId)
		log.WithField("ProjectId", projectID).Info("Starting CLEANUP")

		// Event name cleanup.
		eventNamesKeySortedSet, err := model.GetEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectID,
			currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - smart events")
			return nil, false
		}
		count, err := cacheRedis.ZcardPersistent(eventNamesKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - events")
			return nil, false
		}
		eventCount := int(count)

		eventNamesSmartKeySortedSet, err := model.GetSmartEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectID,
			currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - events")
			return nil, false
		}
		count, err = cacheRedis.ZcardPersistent(eventNamesSmartKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - smart events")
			return nil, false
		}
		smartEventsCount := int(count)
		if eventCount+smartEventsCount > eventsLimit && eventCount > (eventsLimit-eventCount+smartEventsCount-1) {
			log.WithField("ProjectId", projectID).WithField("Count", eventCount+smartEventsCount-1-eventsLimit).Info("Deleting events")
			_, err := cacheRedis.ZRemRangePersistent(eventNamesKeySortedSet, 0, (eventCount + smartEventsCount - 1 - eventsLimit))
			if err != nil {
				log.WithError(err).Error("Failed to delete - events")
				return nil, false
			}
		}

		// Event property name cleanup.
		propertyCategoryKeySortedSet, err := model.GetPropertiesByEventCategoryCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - properties")
			return nil, false
		}
		count, err = cacheRedis.ZcardPersistent(propertyCategoryKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - properties")
			return nil, false
		}
		propCount := int(count)
		if propCount > propertiesLimit {
			log.WithField("ProjectId", projectID).WithField("Count", propCount-1-propertiesLimit).Info("Deleting event properties")
			_, err := cacheRedis.ZRemRangePersistent(propertyCategoryKeySortedSet, 0, (propCount - 1 - propertiesLimit))
			if err != nil {
				log.WithError(err).Error("Failed to delete - properties")
				return nil, false
			}
		}

		// Event property value cleanup.
		valueKeySortedSet, err := model.GetValuesByEventPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - values")
			return nil, false
		}
		count, err = cacheRedis.ZcardPersistent(valueKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count- values")
			return nil, false
		}
		valueCount := int(count)
		if valueCount > valuesLimit {
			log.WithField("ProjectId", projectID).WithField("Count", valueCount-1-valuesLimit).Info("Deleting event properties values")
			_, err := cacheRedis.ZRemRangePersistent(valueKeySortedSet, 0, (valueCount - 1 - valuesLimit))
			if err != nil {
				log.WithError(err).Error("Failed to delete - values")
				return nil, false
			}
		}

		// User property name cleanup.
		userPropertyCategoryKeySortedSet, err := model.GetUserPropertiesCategoryByProjectCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - property category")
			return nil, false
		}
		count, err = cacheRedis.ZcardPersistent(userPropertyCategoryKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - user property")
			return nil, false
		}
		userPropCount := int(count)
		if userPropCount > propertiesLimit {
			log.WithField("ProjectId", projectID).WithField("Count", userPropCount-1-propertiesLimit).Info("Deleting user properties")
			_, err := cacheRedis.ZRemRangePersistent(userPropertyCategoryKeySortedSet, 0, (userPropCount - 1 - propertiesLimit))
			if err != nil {
				log.WithError(err).Error("Failed to delete - user property")
				return nil, false
			}
		}

		// User property value cleanup.
		userValueKeySortedSet, err := model.GetValuesByUserPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - values")
			return nil, false
		}
		count, err = cacheRedis.ZcardPersistent(userValueKeySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - user property value")
			return nil, false
		}
		userValueCount := int(count)
		if userValueCount > valuesLimit {
			log.WithField("ProjectId", projectID).WithField("Count", userValueCount-1-valuesLimit).Info("Deleting user properties values")
			_, err := cacheRedis.ZRemRangePersistent(userValueKeySortedSet, 0, (userValueCount - 1 - valuesLimit))
			if err != nil {
				log.WithError(err).Error("Failed to delete - user property value")
				return nil, false
			}
		}

		// Group property name cleanup.
		groupPropertySortedSet, err := model.GetPropertiesByGroupCategoryCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - group property category")
			return nil, false
		}
		count, err = cacheRedis.ZcardPersistent(groupPropertySortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - group property")
			return nil, false
		}
		groupPropertyCount := int(count)
		if groupPropertyCount > propertiesLimit {
			log.WithField("ProjectId", projectID).WithField("Count", groupPropertyCount-1-propertiesLimit).Info("Deleting group properties")
			_, err := cacheRedis.ZRemRangePersistent(groupPropertySortedSet, 0, (groupPropertyCount - 1 - propertiesLimit))
			if err != nil {
				log.WithError(err).Error("Failed to delete - group property")
				return nil, false
			}
		}

		// Group property value cleanup.
		groupPropertyValueSortedSet, err := model.GetValuesByGroupPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - group property values")
			return nil, false
		}
		count, err = cacheRedis.ZcardPersistent(groupPropertyValueSortedSet)
		if err != nil {
			log.WithError(err).Error("Failed to get count - group property value")
			return nil, false
		}
		groupPropertyValueCount := int(count)
		if groupPropertyValueCount > valuesLimit {
			log.WithField("ProjectId", projectID).WithField("Count", groupPropertyValueCount-1-valuesLimit).Info("Deleting group properties values")
			_, err := cacheRedis.ZRemRangePersistent(groupPropertyValueSortedSet, 0, (groupPropertyValueCount - 1 - valuesLimit))
			if err != nil {
				log.WithError(err).Error("Failed to delete - group property value")
				return nil, false
			}
		}
	}
	return nil, true

}
