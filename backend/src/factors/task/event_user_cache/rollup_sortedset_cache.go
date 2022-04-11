package event_user_cache

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	"factors/model/model"
	U "factors/util"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func DoRollUpSortedSet(configs map[string]interface{}) (map[string]interface{}, bool) {
	// Get all projects sorted set
	// Zrange for all the keys
	// Rollup data
	// delete the sorted set

	rollupLookback := configs["rollupLookback"].(int)

	var isCurrentDay bool
	currentDate := U.TimeNowZ()
	for i := 0; i <= rollupLookback; i++ {
		isCurrentDay = i == 0
		currentTimeDatePart := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		uniqueUsersCountKey, err := model.UserCountAnalyticsCacheKey(
			currentTimeDatePart)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key - uniqueEventsCountKey")
			return nil, false
		}
		allProjects, err := cacheRedis.ZrangeWithScoresPersistent(true, uniqueUsersCountKey)
		log.WithField("projects", allProjects).Info("AllProjects")
		if err != nil {
			log.WithError(err).Error("Failed to get projects")
			return nil, false
		}
		for id, _ := range allProjects {
			projId, _ := strconv.Atoi(id)
			projectID := uint64(projId)
			log.WithField("ProjectId", projectID).Info("Starting RollUp")
			eventNamesSmartKeySortedSet, err := model.GetSmartEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectID,
				currentTimeDatePart)
			if err != nil {
				log.WithError(err).Error("Failed to get cache key - events")
				return nil, false
			}
			eventNamesKeySortedSet, err := model.GetEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectID,
				currentTimeDatePart)
			if err != nil {
				log.WithError(err).Error("Failed to get cache key - smart events")
				return nil, false
			}
			propertyCategoryKeySortedSet, err := model.GetPropertiesByEventCategoryCacheKeySortedSet(projectID, currentTimeDatePart)
			if err != nil {
				log.WithError(err).Error("Failed to get cache key - properties")
				return nil, false
			}
			valueKeySortedSet, err := model.GetValuesByEventPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
			if err != nil {
				log.WithError(err).Error("Failed to get cache key - values")
				return nil, false
			}

			userPropertyCategoryKeySortedSet, err := model.GetUserPropertiesCategoryByProjectCacheKeySortedSet(projectID, currentTimeDatePart)
			if err != nil {
				log.WithError(err).Error("Failed to get cache key - property category")
				return nil, false
			}
			userValueKeySortedSet, err := model.GetValuesByUserPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
			if err != nil {
				log.WithError(err).Error("Failed to get cache key - values")
				return nil, false
			}

			smartEvents, err := cacheRedis.ZrangeWithScoresPersistent(false, eventNamesSmartKeySortedSet)
			log.WithField("Count", len(smartEvents)).Info("SmartEventCount")
			events, err := cacheRedis.ZrangeWithScoresPersistent(false, eventNamesKeySortedSet)
			log.WithField("Count", len(events)).Info("EventsCount")
			properties, err := cacheRedis.ZrangeWithScoresPersistent(false, propertyCategoryKeySortedSet)
			log.WithField("Count", len(properties)).Info("PropertiesCount")
			values, err := cacheRedis.ZrangeWithScoresPersistent(false, valueKeySortedSet)
			log.WithField("Count", len(values)).Info("ValuesCount")
			userProperties, err := cacheRedis.ZrangeWithScoresPersistent(false, userPropertyCategoryKeySortedSet)
			log.WithField("Count", len(userProperties)).Info("UserPropertiesCount")
			userValues, err := cacheRedis.ZrangeWithScoresPersistent(false, userValueKeySortedSet)
			log.WithField("Count", len(userValues)).Info("UserValuesCount")

			if len(events) > 0 {
				// Events
				// 1. Create cache object
				// 2. Get the cache key
				// 3. set cache
				eventNamesRollupObj := GetCacheEventObject(events, smartEvents, currentTimeDatePart)
				eventNamesKey, err := model.GetEventNamesOrderByOccurrenceAndRecencyRollUpCacheKey(projectID, currentTimeDatePart)
				enEventCache, err := json.Marshal(eventNamesRollupObj)
				if err != nil {
					log.WithError(err).Error("Failed to marshall event names")
					continue
				}
				err = cacheRedis.SetPersistent(eventNamesKey, string(enEventCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
				if err != nil {
					log.WithError(err).Error("Failed to set cache")
				}

				// event properties
				propertiesMap := make(map[string]map[string]string)
				for property, count := range properties {
					split1 := strings.Split(property, ":SS-EN-PC:")
					eventName := split1[0]
					property := split1[1]
					if propertiesMap[eventName] == nil {
						propertiesMap[eventName] = make(map[string]string)
					}
					propertiesMap[eventName][property] = count
				}
				eventPropertiesToCache := make(map[*cacheRedis.Key]string)
				for eventName, properties := range propertiesMap {
					if len(properties) > 0 {
						eventPropertiesKey, _ := model.GetPropertiesByEventCategoryRollUpCacheKey(projectID, eventName, currentTimeDatePart)
						cacheEventPropertyObject := GetCachePropertyObject(properties, currentTimeDatePart)
						enEventPropertiesCache, err := json.Marshal(cacheEventPropertyObject)
						if err != nil {
							log.WithError(err).Error("Failed to marshall - event properties")
							continue
						}
						eventPropertiesToCache[eventPropertiesKey] = string(enEventPropertiesCache)
					}
				}
				if len(eventPropertiesToCache) > 0 {
					err = cacheRedis.SetPersistentBatch(eventPropertiesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						log.WithError(err).Error("Failed to set cache")
					}
				}

				// user properties
				if len(properties) > 0 {
					userPropertiesRollupObj := GetCachePropertyObject(userProperties, currentTimeDatePart)
					userPropertiesKey, err := model.GetUserPropertiesCategoryByProjectRollUpCacheKey(projectID, currentTimeDatePart)
					usPropertyCache, err := json.Marshal(userPropertiesRollupObj)
					if err != nil {
						log.WithError(err).Error("Failed to marshall user properties")
						continue
					}
					err = cacheRedis.SetPersistent(userPropertiesKey, string(usPropertyCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						log.WithError(err).Error("Failed to set cache")
					}
				}

				// event property values
				propertyValues := make(map[string]map[string]map[string]string)
				for valueKey, count := range values {
					split1 := strings.Split(valueKey, ":SS-EN-PC:")
					eventName := split1[0]
					split2 := strings.Split(split1[1], ":SS-EN-PV:")
					property := split2[0]
					value := split2[1]
					if propertyValues[eventName] == nil {
						propertyValues[eventName] = make(map[string]map[string]string)
					}
					if propertyValues[eventName][property] == nil {
						propertyValues[eventName][property] = make(map[string]string)
					}
					propertyValues[eventName][property][value] = count
				}
				eventPropertyValuesToCache := make(map[*cacheRedis.Key]string)
				for eventName, propertyValue := range propertyValues {
					for property, values := range propertyValue {
						if len(values) > 0 {
							eventPropertyValuesKey, _ := model.GetValuesByEventPropertyRollUpCacheKey(projectID, eventName, property, currentTimeDatePart)
							cacheEventPropertyValueObject := GetCachePropertyValueObject(values, currentTimeDatePart)
							enEventPropertyValuesCache, err := json.Marshal(cacheEventPropertyValueObject)
							if err != nil {
								log.WithError(err).Error("Failed to marshall - property values")
								continue
							}
							eventPropertyValuesToCache[eventPropertyValuesKey] = string(enEventPropertyValuesCache)
						}
					}
				}
				if len(eventPropertyValuesToCache) > 0 {
					err = cacheRedis.SetPersistentBatch(eventPropertyValuesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						log.WithError(err).Error("Failed to set cache")
					}
				}

				// user property values
				userPropertyValues := make(map[string]map[string]string)
				for valueKey, count := range userValues {
					split1 := strings.Split(valueKey, ":SS-US-PV:")
					property := split1[0]
					value := split1[1]
					if userPropertyValues[property] == nil {
						userPropertyValues[property] = make(map[string]string)
					}
					userPropertyValues[property][value] = count

				}
				userPropertyValuesToCache := make(map[*cacheRedis.Key]string)
				for property, values := range userPropertyValues {
					if len(values) > 0 {
						userPropertyValuesKey, _ := model.GetValuesByUserPropertyRollUpCacheKey(projectID, property, currentTimeDatePart)
						cacheUserPropertyValueObject := GetCachePropertyValueObject(values, currentTimeDatePart)
						usEventPropertyValuesCache, err := json.Marshal(cacheUserPropertyValueObject)
						if err != nil {
							log.WithError(err).Error("Failed to marshall - user property values")
							continue
						}
						userPropertyValuesToCache[userPropertyValuesKey] = string(usEventPropertyValuesCache)
					}
				}
				if len(userPropertyValuesToCache) > 0 {
					err = cacheRedis.SetPersistentBatch(userPropertyValuesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						log.WithError(err).Error("Failed to set cache")
					}
				}

				if isCurrentDay == false {
					err = cacheRedis.DelPersistent(
						eventNamesSmartKeySortedSet,
						eventNamesKeySortedSet,
						propertyCategoryKeySortedSet,
						valueKeySortedSet,
						userPropertyCategoryKeySortedSet,
						userValueKeySortedSet,
					)
					if err != nil {
						log.WithError(err).Error("Failed to del cache keys")
						return nil, false
					}
				}
			}

			groupPropertyCategoryKeySortedSet, err := model.GetPropertiesByGroupCategoryCacheKeySortedSet(projectID, currentTimeDatePart)
			if err != nil {
				log.WithError(err).Error("Failed to get cache key - group property category")
				return nil, false
			}

			groupProperties, err := cacheRedis.ZrangeWithScoresPersistent(false, groupPropertyCategoryKeySortedSet)
			log.WithField("Count", len(groupProperties)).Info("GroupPropertiesCount")
			if err != nil {
				log.WithError(err).Error("Failed to get cache redis key - group property category")
				return nil, false
			}

			if len(groupProperties) > 0 {
				groupValueKeySortedSet, err := model.GetValuesByGroupPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
				if err != nil {
					log.WithError(err).Error("Failed to get cache key - group values")
					return nil, false
				}

				groupValues, err := cacheRedis.ZrangeWithScoresPersistent(false, groupValueKeySortedSet)
				log.WithField("Count", len(groupValues)).Info("GroupValuesCount")
				if err != nil {
					log.WithError(err).Error("Failed to get cache redis key - group values")
					return nil, false
				}

				// group properties
				groupPropertiesMap := make(map[string]map[string]string)
				for property, count := range groupProperties {
					split1 := strings.Split(property, ":SS-GN-PC:")
					groupName := split1[0]
					property := split1[1]
					if groupPropertiesMap[groupName] == nil {
						groupPropertiesMap[groupName] = make(map[string]string)
					}
					groupPropertiesMap[groupName][property] = count
				}
				groupPropertiesToCache := make(map[*cacheRedis.Key]string)
				for groupName, properties := range groupPropertiesMap {
					if len(properties) > 0 {
						groupPropertiesKey, _ := model.GetPropertiesByGroupCategoryRollUpCacheKey(projectID, groupName, currentTimeDatePart)
						cacheGroupPropertyObject := GetCachePropertyObject(properties, currentTimeDatePart)
						enGroupPropertiesCache, err := json.Marshal(cacheGroupPropertyObject)
						if err != nil {
							log.WithError(err).Error("Failed to marshall - group properties")
							continue
						}
						groupPropertiesToCache[groupPropertiesKey] = string(enGroupPropertiesCache)
					}
				}
				if len(groupPropertiesToCache) > 0 {
					err = cacheRedis.SetPersistentBatch(groupPropertiesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						log.WithError(err).Error("Failed to set cache")
					}
				}

				// group property values
				groupPropertyValues := make(map[string]map[string]map[string]string)
				for valueKey, count := range groupValues {
					split1 := strings.Split(valueKey, ":SS-GN-PC:")
					groupName := split1[0]
					split2 := strings.Split(split1[1], ":SS-GN-PV:")
					property := split2[0]
					value := split2[1]
					if groupPropertyValues[groupName] == nil {
						groupPropertyValues[groupName] = make(map[string]map[string]string)
					}
					if groupPropertyValues[groupName][property] == nil {
						groupPropertyValues[groupName][property] = make(map[string]string)
					}
					groupPropertyValues[groupName][property][value] = count
				}
				groupPropertyValuesToCache := make(map[*cacheRedis.Key]string)
				for groupName, propertyValue := range groupPropertyValues {
					for property, values := range propertyValue {
						if len(values) > 0 {
							groupPropertyValuesKey, _ := model.GetValuesByGroupPropertyRollUpCacheKey(projectID, groupName, property, currentTimeDatePart)
							cacheGroupPropertyValueObject := GetCachePropertyValueObject(values, currentTimeDatePart)
							enGroupPropertyValuesCache, err := json.Marshal(cacheGroupPropertyValueObject)
							if err != nil {
								log.WithError(err).Error("Failed to marshall - group property values")
								continue
							}
							groupPropertyValuesToCache[groupPropertyValuesKey] = string(enGroupPropertyValuesCache)
						}
					}
				}
				if len(groupPropertyValuesToCache) > 0 {
					err = cacheRedis.SetPersistentBatch(groupPropertyValuesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						log.WithError(err).Error("Failed to set cache")
					}
				}

				if isCurrentDay == false {
					err = cacheRedis.DelPersistent(
						groupPropertyCategoryKeySortedSet,
						groupValueKeySortedSet,
					)
					if err != nil {
						log.WithError(err).Error("Failed to del cache keys")
						return nil, false
					}
				}
			}
		}
	}
	return nil, true
}

func GetCacheEventObject(events map[string]string, smartEvents map[string]string, date string) model.CacheEventNamesWithTimestamp {
	eventNames := make(map[string]U.CountTimestampTuple)
	for eventName, count := range events {
		eventCount, _ := strconv.Atoi(count)
		keyDate, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, date)
		eventNameCacheObj := U.CountTimestampTuple{
			LastSeenTimestamp: keyDate.Unix(),
			Count:             int64(eventCount),
		}

		eventNames[eventName] = eventNameCacheObj
	}
	for eventName, count := range smartEvents {
		eventCount, _ := strconv.Atoi(count)
		keyDate, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, date)
		eventNameCacheObj := U.CountTimestampTuple{
			LastSeenTimestamp: keyDate.Unix(),
			Count:             int64(eventCount),
			Type:              model.EVENT_NAME_TYPE_SMART_EVENT,
		}
		eventNames[eventName] = eventNameCacheObj
	}
	cacheEventNames := model.CacheEventNamesWithTimestamp{
		EventNames: eventNames}
	return cacheEventNames
}

func GetCachePropertyValueObject(values map[string]string, date string) U.CachePropertyValueWithTimestamp {
	propertyValues := make(map[string]U.CountTimestampTuple)
	for value, count := range values {
		valueCount, _ := strconv.Atoi(count)
		keyDate, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, date)
		valueCacheObj := U.CountTimestampTuple{
			LastSeenTimestamp: keyDate.Unix(),
			Count:             int64(valueCount),
		}
		propertyValues[value] = valueCacheObj
	}
	cachePropertyValues := U.CachePropertyValueWithTimestamp{
		PropertyValue: propertyValues}
	return cachePropertyValues
}

func extractCategoryProperty(propertyCategory string) (string, string) {
	splits := strings.Split(propertyCategory, ":")
	category := splits[0]
	property := splits[1]
	return category, property
}

func GetCachePropertyObject(propertiesCategory map[string]string, date string) U.CachePropertyWithTimestamp {
	var dateKeyInTime time.Time
	eventProperties := make(map[string]U.PropertyWithTimestamp)
	propertyCategory := make(map[string]map[string]int64)
	for prCat, count := range propertiesCategory {
		cat, pr := extractCategoryProperty(prCat)
		dateKeyInTime, _ = time.Parse(U.DATETIME_FORMAT_YYYYMMDD, date)
		if propertyCategory[pr] == nil {
			propertyCategory[pr] = make(map[string]int64)
		}
		catCount, _ := strconv.Atoi(count)
		propertyCategory[pr][cat] = int64(catCount)
	}
	for pr, catCount := range propertyCategory {
		cwc := make(map[string]int64)
		totalCount := int64(0)
		for cat, catCount := range catCount {
			cwc[cat] = catCount
			totalCount += catCount
		}
		prWithTs := U.PropertyWithTimestamp{CategorywiseCount: cwc,
			CountTime: U.CountTimestampTuple{Count: totalCount, LastSeenTimestamp: dateKeyInTime.Unix()}}
		eventProperties[pr] = prWithTs
	}
	cacheProperties := U.CachePropertyWithTimestamp{
		Property: eventProperties}
	return cacheProperties
}
