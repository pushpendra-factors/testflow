package event_user_cache

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	"factors/config"
	"factors/model/model"
	U "factors/util"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func DoRollUpSortedSet(configs map[string]interface{}) (map[string]interface{}, bool) {
	rollupLookback := 0
	if _, exists := configs["rollupLookback"]; exists {
		rollupLookback = configs["rollupLookback"].(int)
	}

	deleteRollupAfterAddingToAggregate := 0
	if _, exists := configs["deleteRollupAfterAddingToAggregate"]; exists {
		deleteRollupAfterAddingToAggregate = configs["deleteRollupAfterAddingToAggregate"].(int)
	}

	log.WithField("config", configs).Info("Starting rollup sorted set job.")

	var isCurrentDay bool
	currentDate := U.TimeNowZ()

	var rollupMinTimestamp int64
	if rollupLookback > 0 {
		rollupMinTimestamp = currentDate.AddDate(0, 0, -rollupLookback).UTC().Unix()
	}

	logCtx := log.WithField("rollup_lookback", rollupLookback).WithField("rollup_min_timestamp", rollupMinTimestamp)

	allProjects := map[string]bool{}
	for i := 0; i <= rollupLookback; i++ {
		currentTimeDatePart := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		uniqueUsersCountKey, err := model.UserCountAnalyticsCacheKey(
			currentTimeDatePart)
		logCtx.WithError(err).Info("Got unique users count key.")
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache key - uniqueEventsCountKey")
			return nil, false
		}

		projects, err := cacheRedis.ZrangeWithScoresPersistent(true, uniqueUsersCountKey)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get projects")
			return nil, false
		}

		for k := range projects {
			allProjects[k] = true
		}
	}

	logCtx.WithField("projects", allProjects).Info("Got all projects.")

	for id := range allProjects {
		projId, _ := strconv.Atoi(id)
		projectID := int64(projId)

		logCtx := logCtx.WithField("project_id", projectID).WithField("tag", "rollup")

		logCtx.Info("Starting rollup for project.")

		// eventNameAndPropertyKeysCacheByDate contains []U.CachePropertyValueWithTimestamp of multiple dates per event_name+property_key.
		// U.CachePropertyValueWithTimestamp contains map of multiple values of a specific day.
		eventNameAndPropertyKeysCacheByDate := make(map[string]map[string][]U.CachePropertyValueWithTimestamp)

		rollupsAddedToAggregate := make([]*cacheRedis.Key, 0)
		for i := 0; i <= rollupLookback; i++ {
			isCurrentDay = i == 0
			currentTimeDatePart := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)

			logCtx.Info("Starting rollup for a project and rollback.")

			eventNamesSmartKeySortedSet, err := model.GetSmartEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectID,
				currentTimeDatePart)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get cache key - smart events")
				continue
			}
			eventNamesPageViewSortedSet, err := model.GetPageViewEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectID,
				currentTimeDatePart)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get cache key - pageView events")
				continue
			}
			eventNamesKeySortedSet, err := model.GetEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectID,
				currentTimeDatePart)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get cache key - events")
				continue
			}
			propertyCategoryKeySortedSet, err := model.GetPropertiesByEventCategoryCacheKeySortedSet(projectID, currentTimeDatePart)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get cache key - properties")
				return nil, false
			}
			valueKeySortedSet, err := model.GetValuesByEventPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get cache key - values")
				continue
			}

			userPropertyCategoryKeySortedSet, err := model.GetUserPropertiesCategoryByProjectCacheKeySortedSet(projectID, currentTimeDatePart)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get cache key - property category")
				continue
			}
			userValueKeySortedSet, err := model.GetValuesByUserPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get cache key - values")
				continue
			}

			smartEvents, err := cacheRedis.ZrangeWithScoresPersistent(false, eventNamesSmartKeySortedSet)
			logCtx.WithField("Count", len(smartEvents)).Info("Got SmartEventCount")
			pageViewEvents, err := cacheRedis.ZrangeWithScoresPersistent(false, eventNamesPageViewSortedSet)
			logCtx.WithField("Count", len(pageViewEvents)).Info("Got PageViewEventsCount")

			events, err := cacheRedis.ZrangeWithScoresPersistent(false, eventNamesKeySortedSet)
			logCtx.WithField("Count", len(events)).Info("Got EventsCount")
			properties, err := cacheRedis.ZrangeWithScoresPersistent(false, propertyCategoryKeySortedSet)
			logCtx.WithField("Count", len(properties)).Info("Got PropertiesCount")
			values, err := cacheRedis.ZrangeWithScoresPersistent(false, valueKeySortedSet)
			logCtx.WithField("Count", len(values)).Info("Got ValuesCount")

			userProperties, err := cacheRedis.ZrangeWithScoresPersistent(false, userPropertyCategoryKeySortedSet)
			logCtx.WithField("Count", len(userProperties)).Info("Got UserPropertiesCount")
			userValues, err := cacheRedis.ZrangeWithScoresPersistent(false, userValueKeySortedSet)
			logCtx.WithField("Count", len(userValues)).Info("Got UserValuesCount")

			logCtx.Info("Received all range counts.")

			if len(events) > 0 {
				// Events
				// 1. Create cache object
				// 2. Get the cache key
				// 3. set cache
				eventNamesRollupObj := GetCacheEventObject(events, smartEvents, pageViewEvents, currentTimeDatePart)
				eventNamesKey, err := model.GetEventNamesOrderByOccurrenceAndRecencyRollUpCacheKey(projectID, currentTimeDatePart)
				enEventCache, err := json.Marshal(eventNamesRollupObj)
				if err != nil {
					logCtx.WithError(err).Error("Failed to marshall event names")
					continue
				}
				err = cacheRedis.SetPersistent(eventNamesKey, string(enEventCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
				if err != nil {
					logCtx.WithError(err).Error("Failed to set events rollup cache")
				}

				logCtx.Info("Cached events rollup.")

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
							logCtx.WithError(err).Error("Failed to marshall - event properties")
							continue
						}
						eventPropertiesToCache[eventPropertiesKey] = string(enEventPropertiesCache)
					}
				}
				if len(eventPropertiesToCache) > 0 {
					err = cacheRedis.SetPersistentBatch(eventPropertiesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						logCtx.WithError(err).Error("Failed to set cache")
					}

					logCtx.Info("Cached event_properties rollup.")
				}

				// user properties
				if len(properties) > 0 {
					userPropertiesRollupObj := GetCachePropertyObject(userProperties, currentTimeDatePart)
					userPropertiesKey, err := model.GetUserPropertiesCategoryByProjectRollUpCacheKey(projectID, currentTimeDatePart)
					usPropertyCache, err := json.Marshal(userPropertiesRollupObj)
					if err != nil {
						logCtx.WithError(err).Error("Failed to marshall user properties")
						continue
					}
					err = cacheRedis.SetPersistent(userPropertiesKey, string(usPropertyCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						logCtx.WithError(err).Error("Failed to set cache user properties rollup.")
					}

					logCtx.Info("Cached user_properties rollup.")
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
					if _, isExists := eventNameAndPropertyKeysCacheByDate[eventName]; config.IsAggrEventPropertyValuesCacheEnabled(projectID) && !isExists {
						eventNameAndPropertyKeysCacheByDate[eventName] = make(map[string][]U.CachePropertyValueWithTimestamp)
					}

					for property, values := range propertyValue {
						if len(values) > 0 {
							eventPropertyValuesKey, _ := model.GetValuesByEventPropertyRollUpCacheKey(projectID, eventName, property, currentTimeDatePart)
							cacheEventPropertyValueObject := GetCachePropertyValueObject(values, currentTimeDatePart)
							enEventPropertyValuesCache, err := json.Marshal(cacheEventPropertyValueObject)
							if err != nil {
								logCtx.WithError(err).Error("Failed to marshall - property values")
								continue
							}
							eventPropertyValuesToCache[eventPropertyValuesKey] = string(enEventPropertyValuesCache)

							// Adding to aggregate will work till prev day.
							if config.IsAggrEventPropertyValuesCacheEnabled(projectID) && !isCurrentDay {
								if _, isExsits := eventNameAndPropertyKeysCacheByDate[eventName][property]; !isExsits {
									eventNameAndPropertyKeysCacheByDate[eventName][property] = make([]U.CachePropertyValueWithTimestamp, 0)
								}
								eventNameAndPropertyKeysCacheByDate[eventName][property] = append(
									eventNameAndPropertyKeysCacheByDate[eventName][property],
									cacheEventPropertyValueObject,
								)
								rollupsAddedToAggregate = append(rollupsAddedToAggregate, eventPropertyValuesKey)
							}
						}
					}
				}
				if len(eventPropertyValuesToCache) > 0 {
					err = cacheRedis.SetPersistentBatch(eventPropertyValuesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						logCtx.WithError(err).Error("Failed to set cache event property values rollup")
					}
				}

				logCtx.Info("Cached event property values rollup.")

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
							logCtx.WithError(err).Error("Failed to marshall - user property values")
							continue
						}
						userPropertyValuesToCache[userPropertyValuesKey] = string(usEventPropertyValuesCache)
					}
				}
				if len(userPropertyValuesToCache) > 0 {
					err = cacheRedis.SetPersistentBatch(userPropertyValuesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						logCtx.WithError(err).Error("Failed to set cache user property values.")
					}
				}

				if isCurrentDay == false {
					err = cacheRedis.DelPersistent(
						eventNamesSmartKeySortedSet,
						eventNamesPageViewSortedSet,
						eventNamesKeySortedSet,
						propertyCategoryKeySortedSet,
						valueKeySortedSet,
						userPropertyCategoryKeySortedSet,
						userValueKeySortedSet,
					)
					if err != nil {
						logCtx.WithError(err).Error("Failed to del cache keys")
						return nil, false
					}
				}
			}

			groupPropertyCategoryKeySortedSet, err := model.GetPropertiesByGroupCategoryCacheKeySortedSet(projectID, currentTimeDatePart)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get cache key - group property category")
				return nil, false
			}

			groupProperties, err := cacheRedis.ZrangeWithScoresPersistent(false, groupPropertyCategoryKeySortedSet)
			logCtx.WithField("Count", len(groupProperties)).Info("GroupPropertiesCount")
			if err != nil {
				logCtx.WithError(err).Error("Failed to get cache redis key - group property category")
				return nil, false
			}

			if len(groupProperties) > 0 {
				groupValueKeySortedSet, err := model.GetValuesByGroupPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
				if err != nil {
					logCtx.WithError(err).Error("Failed to get cache key - group values")
					return nil, false
				}

				groupValues, err := cacheRedis.ZrangeWithScoresPersistent(false, groupValueKeySortedSet)
				logCtx.WithField("Count", len(groupValues)).Info("GroupValuesCount")
				if err != nil {
					logCtx.WithError(err).Error("Failed to get cache redis key - group values")
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
							logCtx.WithError(err).Error("Failed to marshall - group properties")
							continue
						}
						groupPropertiesToCache[groupPropertiesKey] = string(enGroupPropertiesCache)
					}
				}
				if len(groupPropertiesToCache) > 0 {
					err = cacheRedis.SetPersistentBatch(groupPropertiesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						logCtx.WithError(err).Error("Failed to set cache group properties rollup in batch.")
					}
				}

				logCtx.Info("Cached group properties rollup.")

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
								logCtx.WithError(err).Error("Failed to marshall - group property values")
								continue
							}
							groupPropertyValuesToCache[groupPropertyValuesKey] = string(enGroupPropertyValuesCache)
						}
					}
				}
				if len(groupPropertyValuesToCache) > 0 {
					err = cacheRedis.SetPersistentBatch(groupPropertyValuesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
					if err != nil {
						logCtx.WithError(err).Error("Failed to set cache group property values rollup in batch.")
					}
				}

				logCtx.Info("Cached group property values rollup.")

				if isCurrentDay == false {
					err = cacheRedis.DelPersistent(
						groupPropertyCategoryKeySortedSet,
						groupValueKeySortedSet,
					)
					if err != nil {
						logCtx.WithError(err).Error("Failed to del cache keys")
						return nil, false
					}
				}
			}
		}

		if !config.IsAggrEventPropertyValuesCacheEnabled(projectID) {
			continue
		}

		logCtx = logCtx.WithField("tag", "aggregate_cache")
		logCtx.Info("Starting aggregate event property values.")

		// Run through event property values rollups of all recent dates and
		// add to aggregated cache excluding the current date.
		eventPropertyValuesAggregateCache := make(map[*cacheRedis.Key]string)
		for eventName, propertyWithValuesByDate := range eventNameAndPropertyKeysCacheByDate {
			for property, valuesByDate := range propertyWithValuesByDate {
				logCtx = logCtx.WithField("property", property).WithField("event_name", eventName)

				eventPropertyValuesAggCacheKey, _ := model.GetValuesByEventPropertyRollUpAggregateCacheKey(
					projectID, eventName, property)

				// Transform aggregate to tuple to allow aggregation with new rollups.
				existingAggCache, aggCacheExists, err := cacheRedis.GetIfExistsPersistent(eventPropertyValuesAggCacheKey)
				if err != nil {
					logCtx.WithError(err).Error("Failed to get existing values cache aggregate.")
					continue
				}

				var existingAggregate U.CacheEventPropertyValuesAggregate
				if aggCacheExists {
					err = json.Unmarshal([]byte(existingAggCache), &existingAggregate)
					if err != nil {
						logCtx.WithError(err).Error("Failed to unmarshal values cache aggregate.")
						continue
					}
				}

				valuesListFromAggMap := make(map[string]U.CountTimestampTuple)
				for i := range existingAggregate.NameCountTimestampCategoryList {
					exa := existingAggregate.NameCountTimestampCategoryList[i]
					valuesListFromAggMap[exa.Name] = U.CountTimestampTuple{LastSeenTimestamp: exa.Timestamp, Count: exa.Count}
				}

				// Add existing aggregate to next compute.
				valuesListFromAgg := U.CachePropertyValueWithTimestamp{PropertyValue: valuesListFromAggMap}
				valuesList := make([]U.CachePropertyValueWithTimestamp, 0)
				valuesList = append(valuesList, valuesByDate...)
				valuesList = append(valuesList, valuesListFromAgg)

				aggregatedValues := U.AggregatePropertyValuesAcrossDate(valuesList, true, rollupMinTimestamp)
				aggregatedValuesCache := U.CacheEventPropertyValuesAggregate{
					NameCountTimestampCategoryList: aggregatedValues,
				}

				eventPropertyValuesAggCache, err := json.Marshal(aggregatedValuesCache)
				if err != nil {
					logCtx.WithError(err).Error("Failed to marshal aggregate values cache for event property values.")
					continue
				}
				eventPropertyValuesAggregateCache[eventPropertyValuesAggCacheKey] = string(eventPropertyValuesAggCache)
			}
		}

		if len(eventPropertyValuesAggregateCache) > 0 {
			err := cacheRedis.SetPersistentBatch(eventPropertyValuesAggregateCache, 0)
			if err != nil {
				logCtx.WithError(err).Error("Failed to set aggregate cache for event properties.")
				continue
			}
		}

		// Delete rollups only if enabled for backward compatibility on rollback.
		if deleteRollupAfterAddingToAggregate == 1 {
			err := cacheRedis.DelPersistent(
				rollupsAddedToAggregate...,
			)
			if err != nil {
				logCtx.WithError(err).Error("Failed to delete the rollup after added to aggregate.")
				continue
			}
		}
	}

	return nil, true
}

func GetCacheEventObject(events map[string]string, smartEvents map[string]string, pageViewEvents map[string]string, date string) model.CacheEventNamesWithTimestamp {
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
	for eventName, count := range pageViewEvents {
		eventCount, _ := strconv.Atoi(count)
		keyDate, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, date)
		eventNameCacheObj := U.CountTimestampTuple{
			LastSeenTimestamp: keyDate.Unix(),
			Count:             int64(eventCount),
			Type:              model.EVENT_NAME_TYPE_PAGE_VIEW_EVENT,
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
