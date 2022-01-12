package main

import (
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	projectID := flag.Uint64("project_id", 0, "Project Id.")
	eventNameIDstr := flag.String("event_name_id", "", "Event name Id.")
	from := flag.Int64("start_timestamp", 0, "Staring timestamp from events search.")
	to := flag.Int64("end_timestamp", 0, "Ending timestamp from events search. End timestamp will be excluded")
	wetRun := flag.Bool("wet", false, "Wet run")
	flag.Parse()
	defer util.NotifyOnPanic("Task#run_salesforce_backfill_datetime_properties", *env)

	taskID := "run_salesforce_backfill_datetime_properties"
	if *projectID == 0 {
		log.Error("projectId not provided")
		os.Exit(1)
	}

	if *eventNameIDstr == "" {
		log.Panic("Invalid event_name_id")
	}

	if *from <= 0 || *to <= 0 {
		log.Panic("Invalid range.")
	}

	// verify timezone file exist
	if _, err := os.Stat("/usr/local/go/lib/time/zoneinfo.zip"); os.IsNotExist(err) {
		log.Panic("missing timezone info file.")
	}

	config := &C.Configuration{
		AppName: taskID,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  taskID,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		SentryDSN:           *sentryDSN,
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	if !*wetRun {
		log.Info("Running in dry run")
	} else {
		log.Info("Running in wet run")
	}

	eventNameIDSplit := strings.Split(*eventNameIDstr, ",")

	var eventNameIDs []string
	for i := range eventNameIDSplit {
		eventNameID := eventNameIDSplit[i]

		eventNameIDs = append(eventNameIDs, eventNameID)
	}

	log.Info(fmt.Sprintf("Running for event_name_id %v", eventNameIDs))
	propertiesUpdateList, updateCount, err := beginBackFillDateTimePropertiesByEventNameID(*projectID, eventNameIDs, *from, *to, *wetRun)
	if err != nil {
		log.WithFields(log.Fields{"wet_run": *wetRun}).WithError(err).Error("Failed to update event for datetime properties.")
		os.Exit(1)
	}

	log.WithFields(log.Fields{"properties_update_list": propertiesUpdateList,
		"event_properties_update_count":       updateCount["eventPropertieUpdateCount"],
		"event_user_properties_update_count":  updateCount["eventUserPropertiesUpdateCount"],
		"latest_user_properties_update_count": updateCount["latestUserPropertiesUpdateCount"]}).Info("Updating event properties completed.")
}

// GetEventsBetweenRangeByEventNameIDs return events between range by event_name_id. End time is not inclusive
func GetEventsBetweenRangeByEventNameIDs(projectID uint64, eventNameID []string, from, to int64) ([]model.Event, int) {
	db := C.GetServices().Db

	var events []model.Event
	if err := db.Where("project_id = ?", projectID).Where("event_name_id IN ( ? ) AND timestamp BETWEEN ? AND ?", eventNameID, from, to).Find(&events).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// Do not log error. Log on caller, if needed.
			return nil, http.StatusNotFound
		}

		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error(
			"Failed to GetEventsBetweenRangeByEventNameIDs")
		return nil, http.StatusInternalServerError
	}

	return events, http.StatusFound
}

func beginBackFillDateTimePropertiesByEventNameID(projectID uint64, eventNameIDs []string, from, to int64, wetRun bool) (map[string]bool, map[string]int, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_name_ids": eventNameIDs, "from": from, "to": to})
	if projectID == 0 || len(eventNameIDs) < 1 || from == 0 || to == 0 {
		logCtx.Error("Missing fields.")
		return nil, nil, errors.New("missing fields")
	}

	datetimeProperties := make(map[string]bool, 0)

	propertyDetails, status := store.GetStore().GetAllPropertyDetailsByProjectID(projectID, "", true)
	if status != http.StatusFound {
		logCtx.Error("Failed to get property details.")
		return nil, nil, errors.New("failed to get property details")
	}

	for pName, pType := range *propertyDetails {
		if pType == util.PropertyTypeDateTime && strings.HasPrefix(pName, util.SALESFORCE_PROPERTY_PREFIX) {
			datetimeProperties[pName] = true
		}
	}

	if len(datetimeProperties) < 1 {
		logCtx.Error("Empty datetime properties")
		return nil, nil, errors.New("empty datetime properties")
	}

	events, status := GetEventsBetweenRangeByEventNameIDs(projectID, eventNameIDs, from, to)
	if status != http.StatusFound {
		logCtx.Error("Failed to get events between range.")
		return nil, nil, errors.New("failed to get events")
	}

	allPropertiesUpdateList := make(map[string]bool, 0)
	updateCount := make(map[string]int)
	for i := range events {
		eventPropertiesUpdateList, eventUserPropertiesUpdateList, err := updateEventPropertiesAndEventUserProperties(projectID, &events[i], &datetimeProperties, wetRun)
		if err != nil {
			logCtx.WithFields(log.Fields{"event_id": events[i].ID}).WithError(err).Error("Failed to update event.")
			return nil, updateCount, err
		}

		if len(eventPropertiesUpdateList) > 0 {
			updateCount["eventPropertieUpdateCount"]++
			for pName := range eventPropertiesUpdateList {
				allPropertiesUpdateList[pName] = true
			}
		}

		if len(eventUserPropertiesUpdateList) > 0 {
			updateCount["eventUserPropertiesUpdateCount"]++
			for pName := range eventUserPropertiesUpdateList {
				allPropertiesUpdateList[pName] = true
			}
		}

		latestUserPropertiesUpdateList, err := updateLatesUserPropeties(projectID, events[i].UserId, &datetimeProperties, wetRun)
		if err != nil {
			logCtx.WithFields(log.Fields{"event_id": events[i].ID}).WithError(err).Error("Failed to update latest user properties.")
			return nil, updateCount, err
		}

		if len(latestUserPropertiesUpdateList) > 0 {
			updateCount["latestUserPropertiesUpdateCount"]++
			for pName := range latestUserPropertiesUpdateList {
				allPropertiesUpdateList[pName] = true
			}

		}

	}

	return allPropertiesUpdateList, updateCount, nil
}

func updateDatetimePropertiesToUnix(properties *map[string]interface{}, datetimeProperties *map[string]bool) (*map[string]interface{}, *map[string]bool, bool, error) {
	isUpdateRequired := false
	propertiesUpdateList := make(map[string]bool)
	for pName := range *properties {
		if _, exist := (*datetimeProperties)[pName]; exist {
			value := (*properties)[pName]
			if value == nil || value == "" {
				continue
			}

			if _, err := util.GetPropertyValueAsFloat64(value); err == nil { // if already number ignore
				continue
			}

			unixTimestamp, err := model.GetSalesforceDocumentTimestamp(value)
			if err != nil {
				return nil, nil, false, err
			}

			(*properties)[pName] = unixTimestamp
			isUpdateRequired = true
			propertiesUpdateList[pName] = true
		}
	}

	return properties, &propertiesUpdateList, isUpdateRequired, nil
}

func updateEventPropertiesIfRequired(projectID uint64, eventID, eventUserID string, eventProperties *map[string]interface{}, datetimeProperties *map[string]bool, wetRun bool) (*map[string]bool, error) {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_id": eventID, "event_properties": eventProperties})

	newEventProperties, propertiesUpdateList, isUpdateRequired, err := updateDatetimePropertiesToUnix(eventProperties, datetimeProperties)
	if err != nil {
		logCtx.WithFields(log.Fields{"event_properties": eventProperties}).WithError(err).Error("Failed to update datetime properties to unix.")
		return nil, err
	}

	if !isUpdateRequired {
		return nil, nil
	}

	updatedEventPropertiesJsonb, err := util.EncodeToPostgresJsonb(newEventProperties)
	if err != nil {
		logCtx.WithFields(log.Fields{"event_properties": eventProperties}).WithError(err).Error("Failed to encode event properties.")
		return nil, err
	}

	if wetRun {
		status := store.GetStore().OverwriteEventProperties(projectID, eventUserID, eventID, updatedEventPropertiesJsonb)
		if status != http.StatusAccepted {
			return nil, errors.New("failed to update event properties")
		}
	}

	return propertiesUpdateList, nil

}

func updateEventUserPropertiesIfRequired(projectID uint64, eventID, eventUserID string, eventUserProperties *map[string]interface{}, datetimeProperties *map[string]bool, wetRun bool) (*map[string]bool, error) {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_id": eventID, "event_user_properties": eventUserProperties})

	newEventUserProperties, propertiesUpdateList, isUpdateRequired, err := updateDatetimePropertiesToUnix(eventUserProperties, datetimeProperties)
	if err != nil {
		logCtx.WithFields(log.Fields{"event_user_properties": newEventUserProperties}).WithError(err).Error("Failed to update datetime properties to unix.")
		return nil, err
	}

	if !isUpdateRequired {
		return nil, nil
	}

	updatedEventUserPropertiesJsonb, err := util.EncodeToPostgresJsonb(newEventUserProperties)
	if err != nil {
		return nil, err
	}

	if wetRun {
		status := store.GetStore().OverwriteEventUserPropertiesByID(projectID, eventUserID, eventID, updatedEventUserPropertiesJsonb)
		if status != http.StatusAccepted {
			return nil, errors.New("failed to update event properties")
		}
	}

	return propertiesUpdateList, nil

}

func backfillEventIfRequired(projectID uint64, userID string, eventID string, properties *postgres.Jsonb, userProperties *postgres.Jsonb, datetimeProperties *map[string]bool, wetRun bool) (*map[string]bool, *map[string]bool, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "event_id": eventID, "datetime_properties": datetimeProperties})
	if projectID == 0 || userID == "" || eventID == "" || properties == nil {
		return nil, nil, errors.New("missing required field")
	}

	var eventProperties map[string]interface{}
	err := json.Unmarshal(properties.RawMessage, &eventProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal event properties")
		return nil, nil, err
	}

	if eventProperties == nil || len(eventProperties) < 1 {
		return nil, nil, errors.New("empty map found")
	}

	eventPropertiesUpdateList, err := updateEventPropertiesIfRequired(projectID, eventID, userID, &eventProperties, datetimeProperties, wetRun)
	if err != nil {
		logCtx.WithFields(log.Fields{"event_properties": eventProperties}).WithError(err).
			Error("Failed to update event properties.")
		return nil, nil, err
	}

	var eventUserProperties map[string]interface{}
	err = json.Unmarshal(userProperties.RawMessage, &eventUserProperties)
	if err != nil {
		logCtx.WithFields(log.Fields{"event_user_properties": eventUserProperties}).WithError(err).
			Error("Failed to unmarshal event user properties.")
		return nil, nil, err
	}

	if eventUserProperties == nil || len(eventUserProperties) < 1 {
		return nil, nil, errors.New("empty map found")
	}

	eventUserPropertiesUpdateList, err := updateEventUserPropertiesIfRequired(projectID, eventID, userID, &eventUserProperties, datetimeProperties, wetRun)
	if err != nil {
		logCtx.WithFields(log.Fields{"event_properties": eventProperties}).WithError(err).
			Error("Failed to update event user properties.")
		return nil, nil, err
	}

	return eventPropertiesUpdateList, eventUserPropertiesUpdateList, nil
}

func updateEventPropertiesAndEventUserProperties(projectID uint64, event *model.Event, datetimeProperties *map[string]bool, wetRun bool) (map[string]bool, map[string]bool, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "datetime_properties": datetimeProperties, "event_id": event.ID})
	if projectID == 0 || event == nil || datetimeProperties == nil || len(*datetimeProperties) < 1 {
		logCtx.Error("Missing required fields.")
		return nil, nil, errors.New("missing required fields")
	}

	allEventPropertiesUpdateList := make(map[string]bool)
	allEventUserPropertiesUpdateList := make(map[string]bool)
	eventPropertiesUpdateList, eventUserPropertiesUpdateList, err := backfillEventIfRequired(projectID, event.UserId, event.ID, &event.Properties, event.UserProperties, datetimeProperties, wetRun)
	if err != nil {
		logCtx.WithError(err).Error("Failed to backfill event propeties.")
		return nil, nil, err
	}

	if eventPropertiesUpdateList != nil {
		for pName := range *eventPropertiesUpdateList {
			allEventPropertiesUpdateList[pName] = true
		}
	}

	if eventUserPropertiesUpdateList != nil {
		for pName := range *eventUserPropertiesUpdateList {
			allEventUserPropertiesUpdateList[pName] = true
		}
	}

	return allEventPropertiesUpdateList, allEventUserPropertiesUpdateList, nil
}

func updateLatesUserPropeties(projectID uint64, userID string, datetimeProperties *map[string]bool, wetRun bool) (map[string]bool, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": userID})

	if projectID == 0 || userID == "" {
		logCtx.Error("Missing required fields.")
		return nil, errors.New("missing required fields")
	}

	user, status := store.GetStore().GetUser(projectID, userID)
	if status != http.StatusFound {
		return nil, errors.New("failed to get latest user properties")
	}

	var userProperties map[string]interface{}
	err := json.Unmarshal(user.Properties.RawMessage, &userProperties)
	if err != nil {
		return nil, err
	}

	if userProperties == nil || len(userProperties) < 1 {
		return nil, errors.New("empty map found")
	}

	newUserProperties, propertiesUpdateList, isUpdateRequired, err := updateDatetimePropertiesToUnix(&userProperties, datetimeProperties)
	if err != nil {
		logCtx.WithFields(log.Fields{"event_properties": newUserProperties}).WithError(err).Error("Failed to update datetime properties.")
		return nil, err
	}

	if !isUpdateRequired {
		return nil, nil
	}

	userPropertiesJsonb, err := util.EncodeToPostgresJsonb(newUserProperties)
	if err != nil {
		return nil, err
	}

	if wetRun {
		status = store.GetStore().OverwriteUserPropertiesByID(projectID, userID, userPropertiesJsonb, false, 0, "")
		if status != http.StatusAccepted {
			return nil, errors.New("failed to overwrite event user properties")
		}
	}

	return *propertiesUpdateList, nil
}
