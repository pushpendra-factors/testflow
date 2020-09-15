package main

import (
	cacheRedis "factors/cache/redis"
	C "factors/config"
	M "factors/model"
	"factors/util"
	"flag"
	"fmt"
	"sort"
	"strconv"
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

	eventsLimit := flag.Int("events_limit", 1, "")
	propertiesLimit := flag.Int("properties_limit", 10, "")
	valuesLimit := flag.Int("values_limit", 10, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "Task#InstantiateEventUserCache"
	defer util.NotifyOnPanic(taskID, *env)

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

	today := U.TimeNow().Format(U.DATETIME_FORMAT_YYYYMMDD)
	yesterday := U.TimeNow().AddDate(0, 0, -1).Format(U.DATETIME_FORMAT_YYYYMMDD)
	dayBeforeYesterday := U.TimeNow().AddDate(0, 0, -2).Format(U.DATETIME_FORMAT_YYYYMMDD)

	

	type tuple struct {
		Key   *cacheRedis.Key
		Value int64
	}

	eventCount := make([]tuple, 0)
	eventsCountCacheKey, _ := M.GetEventNamesOrderByOccurrenceAndRecencyCountCacheKey(0, "*")
	eventsCountCacheKeyString, _ := eventsCountCacheKey.KeyWithAllProjectsSupport()
	eventCountKeys, _ := cacheRedis.ScanPersistent(eventsCountCacheKeyString, 1000, -1)
	eventCountsPerProject, _ := cacheRedis.MGetPersistent(eventCountKeys...)
	cacheKeysToDecrement := make(map[*cacheRedis.Key]int64)
	// Implement Rollup
	for projIndex, eventsCount := range eventCountsPerProject {
		count, _ := strconv.Atoi(eventsCount)
		if count > *eventsLimit {
			eventsInCacheTodayCacheKey, _ := M.GetEventNamesOrderByOccurrenceAndRecencyCacheKey(eventCountKeys[projIndex].ProjectID, "*", eventCountKeys[projIndex].Suffix)
			cacheKey, _ := eventsInCacheTodayCacheKey.Key()
			eventsPerProject, _ := cacheRedis.ScanPersistent(cacheKey, 1000, -1)
			events, _ := cacheRedis.MGetPersistent(eventsPerProject...)
			for index, event := range eventsPerProject {
				eCount, _ := strconv.Atoi(events[index])
				eventCount = append(eventCount, tuple{
					Key:   event,
					Value: int64(eCount)})
			}
			sort.Slice(eventCount, func(i, j int) bool {
				return eventCount[i].Value > eventCount[j].Value
			})
			cacheKeysToBeDeleted := make([]*cacheRedis.Key, 0)
			for _, event := range eventCount[*eventsLimit:len(eventCount)] {
				cacheKeysToBeDeleted = append(cacheKeysToBeDeleted, event.Key)
			}
			cacheRedis.DelPersistent(cacheKeysToBeDeleted...)
			cacheKeysToDecrement[eventCountKeys[projIndex]] = int64(len(cacheKeysToBeDeleted))
		}
	}
	cacheRedis.DecrByBatchPersistent(cacheKeysToDecrement)

	propertyCount := make([]tuple, 0)
	propertiesCountCacheKey, _ := M.GetPropertiesByEventCategoryCountCacheKey(0, "*")
	propertiesCountCacheKeyString, _ := propertiesCountCacheKey.KeyWithAllProjectsSupport()
	propertyCountKeys, _ := cacheRedis.ScanPersistent(propertiesCountCacheKeyString, 1000, -1)
	propertyCountsPerProject, _ := cacheRedis.MGetPersistent(propertyCountKeys...)
	cacheKeysToDecrement = make(map[*cacheRedis.Key]int64)
	// Implement Rollup
	for projIndex, propertiesCount := range propertyCountsPerProject {
		count, _ := strconv.Atoi(propertiesCount)
		if count > *propertiesLimit {
			propertiesInCacheTodayCacheKey, _ := M.GetPropertiesByEventCategoryCacheKey(propertyCountKeys[projIndex].ProjectID, "*", "*", "*", propertyCountKeys[projIndex].Suffix)
			cacheKey, _ := propertiesInCacheTodayCacheKey.Key()
			propertiesPerProject, _ := cacheRedis.ScanPersistent(cacheKey, 1000, -1)
			properties, _ := cacheRedis.MGetPersistent(propertiesPerProject...)
			for index, property := range propertiesPerProject {
				pCount, _ := strconv.Atoi(properties[index])
				propertyCount = append(propertyCount, tuple{
					Key:   property,
					Value: int64(pCount)})
			}
			sort.Slice(propertyCount, func(i, j int) bool {
				return propertyCount[i].Value > propertyCount[j].Value
			})
			cacheKeysToBeDeleted := make([]*cacheRedis.Key, 0)
			for _, property := range propertyCount[*propertiesLimit:len(propertyCount)] {
				cacheKeysToBeDeleted = append(cacheKeysToBeDeleted, property.Key)
			}
			cacheRedis.DelPersistent(cacheKeysToBeDeleted...)
			cacheKeysToDecrement[propertyCountKeys[projIndex]] = int64(len(cacheKeysToBeDeleted))
		}
	}
	cacheRedis.DecrByBatchPersistent(cacheKeysToDecrement)

	valueCount := make([]tuple, 0)
	valuesCountCacheKey, _ := M.GetValuesByEventPropertyCountCacheKey(0, "*")
	valuesCountCacheKeyString, _ := valuesCountCacheKey.KeyWithAllProjectsSupport()
	valueCountKeys, _ := cacheRedis.ScanPersistent(valuesCountCacheKeyString, 1000, -1)
	valueCountsPerProject, _ := cacheRedis.MGetPersistent(valueCountKeys...)
	cacheKeysToDecrement = make(map[*cacheRedis.Key]int64)
	// Implement Rollup
	for projIndex, valuesCount := range valueCountsPerProject {
		count, _ := strconv.Atoi(valuesCount)
		if count > *valuesLimit {
			valuesInCacheTodayCacheKey, _ := M.GetValuesByEventPropertyCacheKey(valueCountKeys[projIndex].ProjectID, "*", "*", "*", valueCountKeys[projIndex].Suffix)
			cacheKey, _ := valuesInCacheTodayCacheKey.Key()
			valuesPerProject, _ := cacheRedis.ScanPersistent(cacheKey, 1000, -1)
			values, _ := cacheRedis.MGetPersistent(valuesPerProject...)
			for index, value := range valuesPerProject {
				vCount, _ := strconv.Atoi(values[index])
				valueCount = append(valueCount, tuple{
					Key:   value,
					Value: int64(vCount)})
			}
			sort.Slice(valueCount, func(i, j int) bool {
				return valueCount[i].Value > valueCount[j].Value
			})
			cacheKeysToBeDeleted := make([]*cacheRedis.Key, 0)
			for _, value := range valueCount[*valuesLimit:len(valueCount)] {
				cacheKeysToBeDeleted = append(cacheKeysToBeDeleted, value.Key)
			}
			cacheRedis.DelPersistent(cacheKeysToBeDeleted...)
			cacheKeysToDecrement[valueCountKeys[projIndex]] = int64(len(cacheKeysToBeDeleted))
		}
	}
	cacheRedis.DecrByBatchPersistent(cacheKeysToDecrement)

	userpropertyCount := make([]tuple, 0)
	userpropertiesCountCacheKey, _ := M.GetUserPropertiesCategoryByProjectCountCacheKey(0, "*")
	userpropertiesCountCacheKeyString, _ := userpropertiesCountCacheKey.KeyWithAllProjectsSupport()
	userpropertyCountKeys, _ := cacheRedis.ScanPersistent(userpropertiesCountCacheKeyString, 1000, -1)
	userpropertyCountsPerProject, _ := cacheRedis.MGetPersistent(userpropertyCountKeys...)
	cacheKeysToDecrement = make(map[*cacheRedis.Key]int64)
	// Implement Rollup
	for projIndex, userpropertiesCount := range userpropertyCountsPerProject {
		count, _ := strconv.Atoi(userpropertiesCount)
		if count > *propertiesLimit {
			userpropertiesInCacheTodayCacheKey, _ := M.GetUserPropertiesCategoryByProjectCacheKey(userpropertyCountKeys[projIndex].ProjectID, "*", "*", userpropertyCountKeys[projIndex].Suffix)
			cacheKey, _ := userpropertiesInCacheTodayCacheKey.Key()
			userpropertiesPerProject, _ := cacheRedis.ScanPersistent(cacheKey, 1000, -1)
			userproperties, _ := cacheRedis.MGetPersistent(userpropertiesPerProject...)
			for index, property := range userpropertiesPerProject {
				pCount, _ := strconv.Atoi(userproperties[index])
				userpropertyCount = append(userpropertyCount, tuple{
					Key:   property,
					Value: int64(pCount)})
			}
			sort.Slice(userpropertyCount, func(i, j int) bool {
				return userpropertyCount[i].Value > userpropertyCount[j].Value
			})
			cacheKeysToBeDeleted := make([]*cacheRedis.Key, 0)
			for _, property := range userpropertyCount[*propertiesLimit:len(userpropertyCount)] {
				cacheKeysToBeDeleted = append(cacheKeysToBeDeleted, property.Key)
			}
			cacheRedis.DelPersistent(cacheKeysToBeDeleted...)
			cacheKeysToDecrement[userpropertyCountKeys[projIndex]] = int64(len(cacheKeysToBeDeleted))
		}
	}
	cacheRedis.DecrByBatchPersistent(cacheKeysToDecrement)

	uservalueCount := make([]tuple, 0)
	uservaluesCountCacheKey, _ := M.GetValuesByUserPropertyCountCacheKey(0, "*")
	uservaluesCountCacheKeyString, _ := uservaluesCountCacheKey.KeyWithAllProjectsSupport()
	uservalueCountKeys, _ := cacheRedis.ScanPersistent(uservaluesCountCacheKeyString, 1000, -1)
	uservalueCountsPerProject, _ := cacheRedis.MGetPersistent(uservalueCountKeys...)
	cacheKeysToDecrement = make(map[*cacheRedis.Key]int64)
	// Implement Rollup
	for projIndex, uservaluesCount := range uservalueCountsPerProject {
		count, _ := strconv.Atoi(uservaluesCount)
		if count > *valuesLimit {
			uservaluesInCacheTodayCacheKey, _ := M.GetValuesByUserPropertyCacheKey(uservalueCountKeys[projIndex].ProjectID, "*", "*", uservalueCountKeys[projIndex].Suffix)
			cacheKey, _ := uservaluesInCacheTodayCacheKey.Key()
			uservaluesPerProject, _ := cacheRedis.ScanPersistent(cacheKey, 1000, -1)
			values, _ := cacheRedis.MGetPersistent(uservaluesPerProject...)
			for index, value := range uservaluesPerProject {
				vCount, _ := strconv.Atoi(values[index])
				uservalueCount = append(uservalueCount, tuple{
					Key:   value,
					Value: int64(vCount)})
			}
			sort.Slice(uservalueCount, func(i, j int) bool {
				return uservalueCount[i].Value > uservalueCount[j].Value
			})
			cacheKeysToBeDeleted := make([]*cacheRedis.Key, 0)
			for _, value := range uservalueCount[*valuesLimit:len(uservalueCount)] {
				cacheKeysToBeDeleted = append(cacheKeysToBeDeleted, value.Key)
			}
			cacheRedis.DelPersistent(cacheKeysToBeDeleted...)
			cacheKeysToDecrement[uservalueCountKeys[projIndex]] = int64(len(cacheKeysToBeDeleted))
		}
	}
	cacheRedis.DecrByBatchPersistent(cacheKeysToDecrement)

	fmt.Println("Done!!!")
}
