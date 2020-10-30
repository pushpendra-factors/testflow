package event_user_cache

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	M "factors/model"
	U "factors/util"
	"fmt"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func DoRollUpAndCleanUp(eventsLimit *int, propertiesLimit *int, valuesLimit *int, rollupLookback *int) map[string]interface{} {
	eventsRollup := 0
	eventPropertiesRollup := 0
	eventPropertiesValuesRollup := 0
	userPropertiesRollup := 0
	userPropertiesValuesRollup := 0
	projectsEventsTimmed := 0
	projectsEventPropertiesTrimmed := 0
	projectsEventPropertyValuesTrimmed := 0
	projectsUserPropertiesTrimmed := 0
	projectsUserPropertyValuesTrimmed := 0

	currentDate := U.TimeNow()
	// Cleaning up events
	currentdate := U.TimeNow().Format(U.DATETIME_FORMAT_YYYYMMDD)
	cacheKeysToDecrement := make(map[*cacheRedis.Key]int64)
	eventCountKeys, eventCountsPerProject, err := getAllProjectEventCountKeys(currentdate)
	if err != nil {
		log.WithError(err).Error("Error Getting keys")
	}
	for projIndex, eventsCount := range eventCountsPerProject {
		count, _ := strconv.Atoi(eventsCount)
		if count > *eventsLimit {
			projectsEventsTimmed++
			eventsInCacheTodayCacheKey, err := M.GetEventNamesOrderByOccurrenceAndRecencyCacheKey(eventCountKeys[projIndex].ProjectID, "*", eventCountKeys[projIndex].Suffix)
			if err != nil {
				log.WithError(err).Error("Error Getting cache keys")
			}
			deletedCount, err := delLeastOccuringKeys(eventsInCacheTodayCacheKey, *eventsLimit)
			if err != nil {
				log.WithError(err).Error("Error deleting additional keys")
			}
			cacheKeysToDecrement[eventCountKeys[projIndex]] = deletedCount
		}
	}
	if len(cacheKeysToDecrement) > 0 {
		err = cacheRedis.DecrByBatchPersistent(cacheKeysToDecrement)
		if err != nil {
			log.WithError(err).Error("Error Getting keys")
		}
	}

	// cleaning up event properties
	cacheKeysToDecrement = make(map[*cacheRedis.Key]int64)
	eventPropertyCountKeys, eventPropertyCountsPerProject, err := getAllProjectEventPropertyCountKeys(currentdate)
	if err != nil {
		log.WithError(err).Error("Error Getting keys")
	}
	for projIndex, propertiesCount := range eventPropertyCountsPerProject {
		count, _ := strconv.Atoi(propertiesCount)
		if count > *propertiesLimit {
			projectsEventPropertiesTrimmed++
			eventPropertiesInCacheTodayCacheKey, err := M.GetPropertiesByEventCategoryCacheKey(eventPropertyCountKeys[projIndex].ProjectID, "*", "*", "*", eventPropertyCountKeys[projIndex].Suffix)
			if err != nil {
				log.WithError(err).Error("Error Getting cache keys")
			}
			deletedCount, err := delLeastOccuringKeys(eventPropertiesInCacheTodayCacheKey, *propertiesLimit)
			if err != nil {
				log.WithError(err).Error("Error deleting additional keys")
			}
			cacheKeysToDecrement[eventPropertyCountKeys[projIndex]] = deletedCount
		}
	}
	if len(cacheKeysToDecrement) > 0 {
		err = cacheRedis.DecrByBatchPersistent(cacheKeysToDecrement)
		if err != nil {
			log.WithError(err).Error("Error Getting keys")
		}
	}

	// cleaning up event property values
	cacheKeysToDecrement = make(map[*cacheRedis.Key]int64)
	eventPropertyValuesCountKeys, eventPropertyValuesCountsPerProject, err := getAllProjectEventPropertyValueCountKeys(currentdate)
	if err != nil {
		log.WithError(err).Error("Error Getting keys")
	}

	for projIndex, valuesCount := range eventPropertyValuesCountsPerProject {
		count, _ := strconv.Atoi(valuesCount)
		if count > *valuesLimit {
			projectsEventPropertyValuesTrimmed++
			valuesInCacheTodayCacheKey, _ := M.GetValuesByEventPropertyCacheKey(eventPropertyValuesCountKeys[projIndex].ProjectID, "*", "*", "*", eventPropertyValuesCountKeys[projIndex].Suffix)
			if err != nil {
				log.WithError(err).Error("Error Getting cache keys")
			}
			deletedCount, err := delLeastOccuringKeys(valuesInCacheTodayCacheKey, *valuesLimit)
			if err != nil {
				log.WithError(err).Error("Error deleting additional keys")
			}
			cacheKeysToDecrement[eventPropertyValuesCountKeys[projIndex]] = deletedCount
		}
	}
	if len(cacheKeysToDecrement) > 0 {
		err = cacheRedis.DecrByBatchPersistent(cacheKeysToDecrement)
		if err != nil {
			log.WithError(err).Error("Error Getting keys")
		}
	}

	// cleaning up user property keys
	cacheKeysToDecrement = make(map[*cacheRedis.Key]int64)
	userPropertyCountKeys, userPropertyKeysCountsPerProject, err := getAllProjectUserPropertyCountKeys(currentdate)
	if err != nil {
		log.WithError(err).Error("Error Getting keys")
	}
	for projIndex, userpropertiesCount := range userPropertyKeysCountsPerProject {
		count, _ := strconv.Atoi(userpropertiesCount)
		if count > *propertiesLimit {
			projectsUserPropertiesTrimmed++
			userpropertiesInCacheTodayCacheKey, err := M.GetUserPropertiesCategoryByProjectCacheKey(userPropertyCountKeys[projIndex].ProjectID, "*", "*", userPropertyCountKeys[projIndex].Suffix)
			if err != nil {
				log.WithError(err).Error("Error Getting cache keys")
			}
			deletedCount, err := delLeastOccuringKeys(userpropertiesInCacheTodayCacheKey, *propertiesLimit)
			if err != nil {
				log.WithError(err).Error("Error deleting additional keys")
			}
			cacheKeysToDecrement[userPropertyCountKeys[projIndex]] = deletedCount
		}
	}
	if len(cacheKeysToDecrement) > 0 {
		err = cacheRedis.DecrByBatchPersistent(cacheKeysToDecrement)
		if err != nil {
			log.WithError(err).Error("Error Getting keys")
		}
	}

	// cleaning up user property values key
	cacheKeysToDecrement = make(map[*cacheRedis.Key]int64)
	userPropertyValuesKeys, userPropertyValuesKeysCountsPerProject, err := getAllProjectUserPropertyValueCountKeys(currentdate)
	if err != nil {
		log.WithError(err).Error("Error Getting keys")
	}
	for projIndex, uservaluesCount := range userPropertyValuesKeysCountsPerProject {
		count, _ := strconv.Atoi(uservaluesCount)
		if count > *valuesLimit {
			projectsUserPropertyValuesTrimmed++
			uservaluesInCacheTodayCacheKey, err := M.GetValuesByUserPropertyCacheKey(userPropertyValuesKeys[projIndex].ProjectID, "*", "*", userPropertyValuesKeys[projIndex].Suffix)
			if err != nil {
				log.WithError(err).Error("Error Getting cache keys")
			}
			deletedCount, err := delLeastOccuringKeys(uservaluesInCacheTodayCacheKey, *valuesLimit)
			if err != nil {
				log.WithError(err).Error("Error deleting additional keys")
			}
			cacheKeysToDecrement[userPropertyValuesKeys[projIndex]] = deletedCount
		}
	}
	if len(cacheKeysToDecrement) > 0 {
		err = cacheRedis.DecrByBatchPersistent(cacheKeysToDecrement)
		if err != nil {
			log.WithError(err).Error("Error Getting keys")
		}
	}

	var isCurrentDay bool
	for i := 0; i <= *rollupLookback; i++ {
		if i == 0 {
			isCurrentDay = true
		} else {
			isCurrentDay = false
		}
		date := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		// Getting all the Count Keys - We use only the event level count key for processing finding events/prop/values
		// But all of it is fetched for cleaning it up at the end of the day
		eventCountKeys, _, err := getAllProjectEventCountKeys(date)
		if err != nil {
			log.WithError(err).Error("Error Getting keys")
		}
		eventPropertyCountKeys, _, err := getAllProjectEventPropertyCountKeys(date)
		if err != nil {
			log.WithError(err).Error("Error Getting keys")
		}
		eventPropertyValuesCountKeys, _, err := getAllProjectEventPropertyValueCountKeys(date)
		if err != nil {
			log.WithError(err).Error("Error Getting keys")
		}
		userPropertyCountKeys, _, err := getAllProjectUserPropertyCountKeys(date)
		if err != nil {
			log.WithError(err).Error("Error Getting keys")
		}
		userPropertyValueCountKeys, _, err := getAllProjectUserPropertyValueCountKeys(date)
		if err != nil {
			log.WithError(err).Error("Error Getting keys")
		}
		for _, eventKey := range eventCountKeys {
			log.WithField("ProjectId", eventKey.ProjectID).Info("Starting RollUp")
			// Events
			eventsInCacheKey, err := M.GetEventNamesOrderByOccurrenceAndRecencyCacheKey(eventKey.ProjectID, "*", date)
			if err != nil {
				log.WithError(err).Error("Error Getting cache keys")
			}
			eventKeys, eventCounts, err := getAllKeys(eventsInCacheKey)
			if err != nil {
				log.WithError(err).Error("Error Getting keys")
			}

			if len(eventKeys) > 0 {
				cacheEventObject := M.GetCacheEventObject(eventKeys, eventCounts)
				eventNamesKey, err := M.GetEventNamesOrderByOccurrenceAndRecencyRollUpCacheKey(eventKey.ProjectID, date)
				enEventCache, err := json.Marshal(cacheEventObject)
				if err != nil {
					log.WithError(err).Error("Failed to marshall event names")
					return nil
				}
				eventsRollup++
				err = cacheRedis.SetPersistent(eventNamesKey, string(enEventCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
				if err != nil {
					log.WithError(err).Error("Failed to set cache")
					return nil
				}
			}

			// eventProperties
			type keyValue struct {
				key   *cacheRedis.Key
				value string
			}
			distinctEventsInPropertiesKey := make(map[string][]keyValue)
			eventPropertiesInCacheKey, err := M.GetPropertiesByEventCategoryCacheKey(eventKey.ProjectID, "*", "*", "*", date)
			if err != nil {
				log.WithError(err).Error("Error Getting cache keys")
			}
			eventPropertyKeys, eventPropertyCount, err := getAllKeys(eventPropertiesInCacheKey)
			if err != nil {
				log.WithError(err).Error("Error Getting keys")
			}
			allEventsInPropertyCache := make(map[string]bool)
			for index, property := range eventPropertyKeys {
				eventName := extractEventNameFromPropertyKey(property.Prefix)
				allEventsInPropertyCache[eventName] = true
				distinctEventsInPropertiesKey[eventName] = append(distinctEventsInPropertiesKey[eventName], keyValue{
					key:   eventPropertyKeys[index],
					value: eventPropertyCount[index]})
			}
			eventPropertiesToCache := make(map[*cacheRedis.Key]string)
			for eventName, propertyValues := range distinctEventsInPropertiesKey {
				propertyNames := make([]*cacheRedis.Key, 0)
				propertyCounts := make([]string, 0)
				for _, property := range propertyValues {
					propertyNames = append(propertyNames, property.key)
					propertyCounts = append(propertyCounts, property.value)
				}
				cacheEventPropertyObject := M.GetCachePropertyObject(propertyNames, propertyCounts)
				eventPropertiesKey, err := M.GetPropertiesByEventCategoryRollUpCacheKey(eventKey.ProjectID, eventName, date)
				enEventPropertiesCache, err := json.Marshal(cacheEventPropertyObject)
				if err != nil {
					log.WithError(err).Error("Failed to marshall - event properties")
					return nil
				}
				eventPropertiesRollup++
				eventPropertiesToCache[eventPropertiesKey] = string(enEventPropertiesCache)
			}
			if len(eventPropertiesToCache) > 0 {
				err = cacheRedis.SetPersistentBatch(eventPropertiesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
				if err != nil {
					log.WithError(err).Error("Failed to set cache")
					return nil
				}
			}

			// event property values
			distinctPropertiesInValuesKey := make(map[string]map[string][]keyValue)
			eventvaluesInCacheKey, _ := M.GetValuesByEventPropertyCacheKey(eventKey.ProjectID, "*", "*", "*", date)
			if err != nil {
				log.WithError(err).Error("Error Getting cache keys")
			}
			eventValueKeys, eventValueCount, err := getAllKeys(eventvaluesInCacheKey)
			if err != nil {
				log.WithError(err).Error("Error Getting keys")
			}

			eventPropertyValuesToCache := make(map[*cacheRedis.Key]string)
			for eventName, _ := range allEventsInPropertyCache {
				distinctPropertiesInValuesKey[eventName] = make(map[string][]keyValue)
				for index, value := range eventValueKeys {
					propertyName := extractPropertyKeyFromValueKey(value.Prefix, eventName)
					if propertyName != "" {
						distinctPropertiesInValuesKey[eventName][propertyName] = append(distinctPropertiesInValuesKey[eventName][propertyName], keyValue{
							key:   eventValueKeys[index],
							value: eventValueCount[index]})
					}
				}
			}
			for eventName, propertyDetails := range distinctPropertiesInValuesKey {
				for property, propertyValueDetails := range propertyDetails {
					valueNames := make([]*cacheRedis.Key, 0)
					valueCounts := make([]string, 0)
					for _, value := range propertyValueDetails {
						valueNames = append(valueNames, value.key)
						valueCounts = append(valueCounts, value.value)
					}
					cacheEventPropertyValueObject := M.GetCachePropertyValueObject(valueNames, valueCounts)
					eventPropertyValuesKey, _ := M.GetValuesByEventPropertyRollUpCacheKey(eventKey.ProjectID, eventName, property, date)
					enEventPropertyValuesCache, err := json.Marshal(cacheEventPropertyValueObject)
					if err != nil {
						log.WithError(err).Error("Failed to marshall - property values")
						return nil
					}
					eventPropertiesValuesRollup++
					eventPropertyValuesToCache[eventPropertyValuesKey] = string(enEventPropertyValuesCache)
				}
			}

			if len(eventPropertyValuesToCache) > 0 {
				err = cacheRedis.SetPersistentBatch(eventPropertyValuesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
				if err != nil {
					log.WithError(err).Error("Failed to set cache")
					return nil
				}
			}
			if len(eventValueKeys) > 0 && isCurrentDay == false {
				log.WithField("ProjectId", eventKey.ProjectID).WithField("length", len(eventValueKeys)).Info("DEL:EN:PV")
				err = cacheRedis.DelPersistent(eventValueKeys...)
				if err != nil {
					log.WithError(err).Error("Failed to del cache keys")
					return nil
				}
			}
			if len(eventPropertyKeys) > 0 && isCurrentDay == false {
				log.WithField("ProjectId", eventKey.ProjectID).WithField("length", len(eventPropertyKeys)).Info("DEL:EN:PC")
				err = cacheRedis.DelPersistent(eventPropertyKeys...)
				if err != nil {
					log.WithError(err).Error("Failed to del cache keys")
					return nil
				}
			}
			if len(eventKeys) > 0 && isCurrentDay == false {
				log.WithField("ProjectId", eventKey.ProjectID).WithField("length", len(eventKeys)).Info("DEL:EN")
				err = cacheRedis.DelPersistent(eventKeys...)
				if err != nil {
					log.WithError(err).Error("Failed to del cache keys")
					return nil
				}
			}
		}

		if len(eventCountKeys) > 0 && isCurrentDay == false {
			log.WithField("length", len(eventCountKeys)).Info("DEL:C:EN")
			err = cacheRedis.DelPersistent(eventCountKeys...)
			if err != nil {
				log.WithError(err).Error("Failed to del cache keys")
				return nil
			}
		}
		if len(eventPropertyCountKeys) > 0 && isCurrentDay == false {
			log.WithField("length", len(eventPropertyCountKeys)).Info("DEL:C:EN:PC")
			err = cacheRedis.DelPersistent(eventPropertyCountKeys...)
			if err != nil {
				log.WithError(err).Error("Failed to del cache keys")
				return nil
			}
		}
		if len(eventPropertyValuesCountKeys) > 0 && isCurrentDay == false {
			log.WithField("length", len(eventPropertyValuesCountKeys)).Info("DEL:C:EN:PV")
			err = cacheRedis.DelPersistent(eventPropertyValuesCountKeys...)
			if err != nil {
				log.WithError(err).Error("Failed to del cache keys")
				return nil
			}
		}
		for _, property := range userPropertyCountKeys {
			log.WithField("project_id", property.ProjectID).Info("Starting Rollup for UserProperties")
			userPropertiesInCacheKey, err := M.GetUserPropertiesCategoryByProjectCacheKey(property.ProjectID, "*", "*", date)
			if err != nil {
				log.WithError(err).Error("Error Getting cache keys")
			}
			userPropertyKeys, userPropertyCount, err := getAllKeys(userPropertiesInCacheKey)
			if err != nil {
				log.WithError(err).Error("Error Getting keys")
			}
			if len(userPropertyKeys) > 0 {
				cacheUserPropertyObject := M.GetCachePropertyObject(userPropertyKeys, userPropertyCount)
				propertyCacheKey, err := M.GetUserPropertiesCategoryByProjectRollUpCacheKey(property.ProjectID, date)
				enPropertiesCache, err := json.Marshal(cacheUserPropertyObject)
				if err != nil {
					log.WithError(err).Error("Failed to marshal property key - getuserpropertiesbyproject")
				}
				userPropertiesRollup++
				err = cacheRedis.SetPersistent(propertyCacheKey, string(enPropertiesCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
				if err != nil {
					log.WithError(err).Error("Failed to set cache")
					return nil
				}
			}
			type keyValue struct {
				key   *cacheRedis.Key
				value string
			}
			uservaluesInCacheKey, _ := M.GetValuesByUserPropertyCacheKey(property.ProjectID, "*", "*", date)
			if err != nil {
				log.WithError(err).Error("Error Getting cache keys")
			}
			userValueKeys, userValueCount, err := getAllKeys(uservaluesInCacheKey)
			if err != nil {
				log.WithError(err).Error("Error Getting keys")
			}

			distinctPropertiesInValuesKey := make(map[string][]keyValue)
			for index, value := range userValueKeys {
				propertyName := extractUserPropertyKeyFromValueKey(value.Prefix)
				if propertyName != "" {
					distinctPropertiesInValuesKey[propertyName] = append(distinctPropertiesInValuesKey[propertyName], keyValue{
						key:   userValueKeys[index],
						value: userValueCount[index]})
				}
			}
			userPropertiesToCache := make(map[*cacheRedis.Key]string)
			for propertyName, propertyValueDetails := range distinctPropertiesInValuesKey {
				valueNames := make([]*cacheRedis.Key, 0)
				valueCounts := make([]string, 0)
				for _, value := range propertyValueDetails {
					valueNames = append(valueNames, value.key)
					valueCounts = append(valueCounts, value.value)
				}
				cacheUserPropertyValueObject := M.GetCachePropertyValueObject(valueNames, valueCounts)
				PropertyValuesKey, err := M.GetValuesByUserPropertyRollUpCacheKey(property.ProjectID, propertyName, date)
				enPropertyValuesCache, err := json.Marshal(cacheUserPropertyValueObject)
				if err != nil {
					log.WithError(err).Error("Failed to marshal property value - getvaluesbyuserproperty")
				}
				userPropertiesValuesRollup++
				userPropertiesToCache[PropertyValuesKey] = string(enPropertyValuesCache)
			}
			if len(userPropertiesToCache) > 0 {
				err = cacheRedis.SetPersistentBatch(userPropertiesToCache, U.EVENT_USER_CACHE_EXPIRY_SECS)
				if err != nil {
					log.WithError(err).Error("Failed to set cache")
					return nil
				}
			}
			if len(userValueKeys) > 0 && isCurrentDay == false {
				log.WithField("ProjectId", property.ProjectID).WithField("length", len(userValueKeys)).Info("DEL:US:PV")
				err = cacheRedis.DelPersistent(userValueKeys...)
				if err != nil {
					log.WithError(err).Error("Failed to del cache keys")
					return nil
				}
			}

			if len(userPropertyKeys) > 0 && isCurrentDay == false {
				log.WithField("ProjectId", property.ProjectID).WithField("length", len(userPropertyKeys)).Info("DEL:US:PC")
				err = cacheRedis.DelPersistent(userPropertyKeys...)
				if err != nil {
					log.WithError(err).Error("Failed to del cache keys")
					return nil
				}
			}
		}
		if len(userPropertyCountKeys) > 0 && isCurrentDay == false {
			log.WithField("length", len(userPropertyCountKeys)).Info("DEL:C:US:PC")
			err = cacheRedis.DelPersistent(userPropertyCountKeys...)
			if err != nil {
				log.WithError(err).Error("Failed to del cache keys")
				return nil
			}
		}
		if len(userPropertyValueCountKeys) > 0 && isCurrentDay == false {
			log.WithField("length", len(userPropertyValueCountKeys)).Info("DEL:C:US:PV")
			err = cacheRedis.DelPersistent(userPropertyValueCountKeys...)
			if err != nil {
				log.WithError(err).Error("Failed to del cache keys")
				return nil
			}
		}
	}

	status := map[string]interface{}{
		"no_of_events_rollup":                 eventsRollup,
		"no_of_event_properties_rollup":       eventPropertiesRollup,
		"no_of_event_property_values_rollup":  eventPropertiesValuesRollup,
		"no_of_user_properties_rollup":        userPropertiesRollup,
		"no_of_user_property_values_rollup":   userPropertiesValuesRollup,
		"no_of_events_trimmed":                projectsEventsTimmed,
		"no_of_event_properties_trimmed":      projectsEventPropertiesTrimmed,
		"no_of_event_property_values_trimmed": projectsEventPropertyValuesTrimmed,
		"no_of_user_properties_trimmed":       projectsUserPropertiesTrimmed,
		"no_of_user_property_values_trimmed":  projectsUserPropertyValuesTrimmed,
	}

	return status
}

func getAllProjectEventCountKeys(dateKey string) ([]*cacheRedis.Key, []string, error) {
	eventsCountCacheKey, err := M.GetEventNamesOrderByOccurrenceAndRecencyCountCacheKey(0, dateKey)
	if err != nil {
		return nil, nil, err
	}
	return getAllKeysWithAllProjectSupport(eventsCountCacheKey)
}

func getAllProjectEventPropertyCountKeys(dateKey string) ([]*cacheRedis.Key, []string, error) {
	propertiesCountCacheKey, err := M.GetPropertiesByEventCategoryCountCacheKey(0, dateKey)
	if err != nil {
		return nil, nil, err
	}
	return getAllKeysWithAllProjectSupport(propertiesCountCacheKey)
}

func getAllProjectEventPropertyValueCountKeys(dateKey string) ([]*cacheRedis.Key, []string, error) {
	valuesCountCacheKey, err := M.GetValuesByEventPropertyCountCacheKey(0, dateKey)
	if err != nil {
		return nil, nil, err
	}
	return getAllKeysWithAllProjectSupport(valuesCountCacheKey)
}

func getAllProjectUserPropertyCountKeys(dateKey string) ([]*cacheRedis.Key, []string, error) {
	userpropertiesCountCacheKey, err := M.GetUserPropertiesCategoryByProjectCountCacheKey(0, dateKey)
	if err != nil {
		return nil, nil, err
	}
	return getAllKeysWithAllProjectSupport(userpropertiesCountCacheKey)
}

func getAllProjectUserPropertyValueCountKeys(dateKey string) ([]*cacheRedis.Key, []string, error) {
	uservaluesCountCacheKey, err := M.GetValuesByUserPropertyCountCacheKey(0, dateKey)
	if err != nil {
		return nil, nil, err
	}
	return getAllKeysWithAllProjectSupport(uservaluesCountCacheKey)
}

func getAllKeysWithAllProjectSupport(key *cacheRedis.Key) ([]*cacheRedis.Key, []string, error) {
	keyString, err := key.KeyWithAllProjectsSupport()
	if err != nil {
		return nil, nil, err
	}
	countKeys, err := cacheRedis.ScanPersistent(keyString, 1000, -1)
	if err != nil {
		return nil, nil, err
	}
	countsPerProject := make([]string, 0)
	if len(countKeys) > 0 {
		countsPerProject, err = cacheRedis.MGetPersistent(countKeys...)
		if err != nil {
			return nil, nil, err
		}
	}
	return countKeys, countsPerProject, nil
}

func getAllKeys(key *cacheRedis.Key) ([]*cacheRedis.Key, []string, error) {
	cacheKey, err := key.Key()
	if err != nil {
		return nil, nil, err
	}
	perProjectKeys, err := cacheRedis.ScanPersistent(cacheKey, 1000, -1)
	if err != nil {
		return nil, nil, err
	}
	keyValues := make([]string, 0)
	if len(perProjectKeys) > 0 {
		keyValues, err = cacheRedis.MGetPersistent(perProjectKeys...)
		if err != nil {
			return nil, nil, err
		}
	}
	return perProjectKeys, keyValues, nil
}

func delLeastOccuringKeys(key *cacheRedis.Key, limit int) (int64, error) {
	type tuple struct {
		Key   *cacheRedis.Key
		Value int64
	}
	keyCount := make([]tuple, 0)
	perProjectKeys, keyValues, err := getAllKeys(key)
	if err != nil {
		return 0, err
	}
	for index, key := range perProjectKeys {
		count, err := strconv.Atoi(keyValues[index])
		if err != nil {
			return 0, err
		}
		keyCount = append(keyCount, tuple{
			Key:   key,
			Value: int64(count)})
	}
	sort.Slice(keyCount, func(i, j int) bool {
		return keyCount[i].Value > keyCount[j].Value
	})
	cacheKeysToBeDeleted := make([]*cacheRedis.Key, 0)
	for _, keyValue := range keyCount[limit:len(keyCount)] {
		cacheKeysToBeDeleted = append(cacheKeysToBeDeleted, keyValue.Key)
	}
	err = cacheRedis.DelPersistent(cacheKeysToBeDeleted...)
	if err != nil {
		return 0, err
	}
	log.WithField("key", key).WithField("Count", len(cacheKeysToBeDeleted)).Info("DeletedKeys")
	return int64(len(cacheKeysToBeDeleted)), nil
}

func extractEventNameFromPropertyKey(key string) string {
	values := strings.SplitN(key, ":", 3)
	return values[2]
}

func extractPropertyKeyFromValueKey(key string, eventName string) string {
	values := strings.SplitN(key, fmt.Sprintf("%s:", eventName), 2)
	if len(values) > 1 {
		return values[1]
	}
	return ""
}

func extractUserPropertyKeyFromValueKey(key string) string {
	values := strings.SplitN(key, ":", 3)
	return values[2]
}
