package main

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	M "factors/model"
	U "factors/util"
	"flag"
	"fmt"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	RedisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")
	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	errorReportingInterval := flag.Int("error_reporting_interval", 300, "")

	eventsLimit := flag.Int("events_limit", 4, "")
	propertiesLimit := flag.Int("properties_limit", 10, "")
	valuesLimit := flag.Int("values_limit", 10, "")
	// This is in days
	rollupLookback := flag.Int("rollup_lookback", 1, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "Task#InstantiateEventUserCache"
	defer U.NotifyOnPanic(taskID, *env)

	config := &C.Configuration{
		AppName: "instantiate_event_user_cacheß",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHostPersistent:    *redisHostPersistent,
		RedisPortPersistent:    *RedisPortPersistent,
		AWSKey:                 *awsAccessKeyId,
		AWSSecret:              *awsSecretAccessKey,
		AWSRegion:              *awsRegion,
		EmailSender:            *factorsEmailSender,
		ErrorReportingInterval: *errorReportingInterval,
		SentryDSN:              *sentryDSN,
	}

	C.InitConf(config.Env)

	// Cache dependency for requests not using queue.
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	C.InitLogClient(config.Env, config.AppName, config.EmailSender, config.AWSKey,
		config.AWSSecret, config.AWSRegion, config.ErrorReportingInterval, config.SentryDSN)

	for i := 1; i <= *rollupLookback; i++ {
		date := U.TimeNow().AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
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
		userPropertyValuesKeys, _, err := getAllProjectUserPropertyValueCountKeys(date)
		if err != nil {
			log.WithError(err).Error("Error Getting keys")
		}
		for _, eventKey := range eventCountKeys {
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
					return
				}
				log.Info("RollUp:EN")
				err = cacheRedis.SetPersistent(eventNamesKey, string(enEventCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
				if err != nil {
					log.WithError(err).Error("Failed to set cache")
					return
				}
				err = cacheRedis.DelPersistent(eventKeys...)
				if err != nil {
					log.WithError(err).Error("Failed to del cache keys")
					return
				}
				for eventName, _ := range cacheEventObject.EventNames {
					eventPropertiesInCacheKey, err := M.GetPropertiesByEventCategoryCacheKey(eventKey.ProjectID, eventName, "*", "*", date)
					if err != nil {
						log.WithError(err).Error("Error Getting cache keys")
					}
					eventPropertyKeys, eventPropertyCount, err := getAllKeys(eventPropertiesInCacheKey)
					if err != nil {
						log.WithError(err).Error("Error Getting keys")
					}
					if len(eventPropertyKeys) > 0 {
						cacheEventPropertyObject := M.GetCachePropertyObject(eventPropertyKeys, eventPropertyCount)
						eventPropertiesKey, err := M.GetPropertiesByEventCategoryRollUpCacheKey(eventKey.ProjectID, eventName, date)
						enEventPropertiesCache, err := json.Marshal(cacheEventPropertyObject)
						if err != nil {
							log.WithError(err).Error("Failed to marshall - event properties")
							return
						}
						log.Info("RollUp:EP")
						err = cacheRedis.SetPersistent(eventPropertiesKey, string(enEventPropertiesCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
						if err != nil {
							log.WithError(err).Error("Failed to set cache")
							return
						}
						err = cacheRedis.DelPersistent(eventPropertyKeys...)
						if err != nil {
							log.WithError(err).Error("Failed to del cache keys")
							return
						}
						for propertyName, _ := range cacheEventPropertyObject.Property {
							eventvaluesInCacheKey, _ := M.GetValuesByEventPropertyCacheKey(eventKey.ProjectID, eventName, propertyName, "*", date)
							if err != nil {
								log.WithError(err).Error("Error Getting cache keys")
							}
							eventValueKeys, eventValueCount, err := getAllKeys(eventvaluesInCacheKey)
							if err != nil {
								log.WithError(err).Error("Error Getting keys")
							}
							if len(eventValueKeys) > 0 {
								cacheEventPropertyValueObject := M.GetCachePropertyValueObject(eventValueKeys, eventValueCount)
								eventPropertyValuesKey, _ := M.GetValuesByEventPropertyRollUpCacheKey(eventKey.ProjectID, eventName, propertyName, date)
								enEventPropertyValuesCache, err := json.Marshal(cacheEventPropertyValueObject)
								if err != nil {
									log.WithError(err).Error("Failed to marshall - property values")
									return
								}
								log.Info("RollUp:EV")
								err = cacheRedis.SetPersistent(eventPropertyValuesKey, string(enEventPropertyValuesCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
								if err != nil {
									log.WithError(err).Error("Failed to set cache")
									return
								}
								err = cacheRedis.DelPersistent(eventValueKeys...)
								if err != nil {
									log.WithError(err).Error("Failed to del cache keys")
									return
								}
							}
						}
					}
				}
			}
		}
		if len(eventCountKeys) > 0 {
			err = cacheRedis.DelPersistent(eventCountKeys...)
			if err != nil {
				log.WithError(err).Error("Failed to del cache keys")
				return
			}
		}
		if len(eventPropertyCountKeys) > 0 {
			err = cacheRedis.DelPersistent(eventPropertyCountKeys...)
			if err != nil {
				log.WithError(err).Error("Failed to del cache keys")
				return
			}
		}
		if len(eventPropertyValuesCountKeys) > 0 {
			err = cacheRedis.DelPersistent(eventPropertyValuesCountKeys...)
			if err != nil {
				log.WithError(err).Error("Failed to del cache keys")
				return
			}
		}
		for _, property := range userPropertyCountKeys {
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
				log.Info("RollUp:UP")
				err = cacheRedis.SetPersistent(propertyCacheKey, string(enPropertiesCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
				if err != nil {
					log.WithError(err).Error("Failed to set cache")
					return
				}
				err = cacheRedis.DelPersistent(userPropertyKeys...)
				if err != nil {
					log.WithError(err).Error("Failed to del cache keys")
					return
				}
				for propertyName, _ := range cacheUserPropertyObject.Property {
					uservaluesInCacheKey, _ := M.GetValuesByUserPropertyCacheKey(property.ProjectID, propertyName, "*", date)
					if err != nil {
						log.WithError(err).Error("Error Getting cache keys")
					}
					userValueKeys, userValueCount, err := getAllKeys(uservaluesInCacheKey)
					if err != nil {
						log.WithError(err).Error("Error Getting keys")
					}
					if len(userValueKeys) > 0 {
						cacheUserPropertyValueObject := M.GetCachePropertyValueObject(userValueKeys, userValueCount)
						PropertyValuesKey, err := M.GetValuesByUserPropertyRollUpCacheKey(property.ProjectID, propertyName, date)
						enPropertyValuesCache, err := json.Marshal(cacheUserPropertyValueObject)
						if err != nil {
							log.WithError(err).Error("Failed to marshal property value - getvaluesbyuserproperty")
						}
						log.Info("RollUp:UV")
						err = cacheRedis.SetPersistent(PropertyValuesKey, string(enPropertyValuesCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
						if err != nil {
							log.WithError(err).Error("Failed to set cache")
							return
						}
						err = cacheRedis.DelPersistent(userValueKeys...)
						if err != nil {
							log.WithError(err).Error("Failed to del cache keys")
							return
						}
					}
				}
			}
		}
		if len(userPropertyCountKeys) > 0 {
			err = cacheRedis.DelPersistent(userPropertyCountKeys...)
			if err != nil {
				log.WithError(err).Error("Failed to del cache keys")
				return
			}
		}
		if len(userPropertyValuesKeys) > 0 {
			err = cacheRedis.DelPersistent(userPropertyValuesKeys...)
			if err != nil {
				log.WithError(err).Error("Failed to del cache keys")
				return
			}
		}
	}

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

	fmt.Println("Done!!!")
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
	log.WithField("Count", len(cacheKeysToBeDeleted)).Info("DeletedKeys")
	return int64(len(cacheKeysToBeDeleted)), nil
}
