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

	var eventNameIDs []uint64
	for i := range eventNameIDSplit {
		eventNameID, err := util.GetPropertyValueAsFloat64(eventNameIDSplit[i])
		if err != nil {
			log.Panic("invalid event name id")
		}

		eventNameIDs = append(eventNameIDs, uint64(eventNameID))
	}

	log.Info(fmt.Sprintf("Running for event_name_id %d", eventNameIDs))
	propertiesUpdateList, eventPropertiesUpdateCount, eventUserPropertiesUpdateCount, latestUserPropertiesUpdateCount, err := beginBackFillDateTimePropertiesByEventNameID(*projectID, eventNameIDs, *from, *to, *wetRun)
	if err != nil {
		log.WithFields(log.Fields{"wet_run": *wetRun}).WithError(err).Error("Failed to update event for datetime properties.")
		os.Exit(1)
	}

	log.WithFields(log.Fields{"properties_update_list": propertiesUpdateList,
		"event_properties_update_count":       eventPropertiesUpdateCount,
		"event_user_properties_update_count":  eventUserPropertiesUpdateCount,
		"latest_user_properties_update_count": latestUserPropertiesUpdateCount}).Info("Updating event properties completed.")
}

// GetEventsBetweenRangeByEventNameIDs return events between range by event_name_id. End time is not inclusive
func GetEventsBetweenRangeByEventNameIDs(projectID uint64, eventNameID []uint64, from, to int64) ([]model.Event, int) {
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

func beginBackFillDateTimePropertiesByEventNameID(projectID uint64, eventNameIDs []uint64, from, to int64, wetRun bool) (*map[string]bool, int, int, int, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_name_ids": eventNameIDs, "from": from, "to": to})
	if projectID == 0 || len(eventNameIDs) < 1 || from == 0 || to == 0 {
		logCtx.Error("Missing fields.")
		return nil, 0, 0, 0, errors.New("missing fields")
	}

	datetimeProperties := make(map[string]bool, 0)
	for i := range eventNameIDs {
		eventName, err := store.GetStore().GetEventNameFromEventNameId(eventNameIDs[i], projectID)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event name.")
			return nil, 0, 0, 0, errors.New("failed to get event name")
		}

		propertyDetails, status := store.GetStore().GetAllPropertyDetailsByProjectID(projectID, eventName.Name, false)
		if status != http.StatusFound {
			logCtx.Error("Failed to get property details.")
			return nil, 0, 0, 0, errors.New("failed to get property details")
		}

		for pName, pType := range *propertyDetails {
			if pType == util.PropertyTypeDateTime {
				datetimeProperties[pName] = true
			}
		}

	}

	if len(datetimeProperties) < 1 {
		logCtx.Error("Empty datetime properties")
		return nil, 0, 0, 0, errors.New("empty datetime properties")
	}

	events, status := GetEventsBetweenRangeByEventNameIDs(projectID, eventNameIDs, from, to)
	if status != http.StatusFound {
		logCtx.Error("Failed to get events between range.")
		return nil, 0, 0, 0, errors.New("failed to get events")
	}

	allPropertiesUpdateList := make(map[string]bool, 0)
	eventPropertieUpdateCount := 0
	eventUserPropertiesUpdateCount := 0
	latestUserPropertiesUpdateCount := 0
	for i := range events {
		propertiesUpdateList, eventPropertiesUpdate, eventUserPropertiesUpdate, err := updateEventPropertiesAndEventUserProperties(projectID, &events[i], &datetimeProperties, wetRun)
		if err != nil {
			logCtx.WithFields(log.Fields{"event_id": events[i].ID}).WithError(err).Error("Failed to update event.")
			return nil, eventPropertieUpdateCount, eventUserPropertiesUpdateCount, latestUserPropertiesUpdateCount, err
		}

		if eventPropertiesUpdate {
			eventPropertieUpdateCount++
		}

		if eventUserPropertiesUpdate {
			eventUserPropertiesUpdateCount++
		}

		if propertiesUpdateList != nil {
			for pName := range *propertiesUpdateList {
				allPropertiesUpdateList[pName] = true
			}
		}

		propertiesUpdateList, err = updateLatesUserPropeties(projectID, events[i].UserId, &datetimeProperties, wetRun)
		if err != nil {
			logCtx.WithFields(log.Fields{"event_id": events[i].ID}).WithError(err).Error("Failed to update latest user properties.")
			return nil, eventPropertieUpdateCount, eventUserPropertiesUpdateCount, latestUserPropertiesUpdateCount, err
		}

		if propertiesUpdateList != nil {
			latestUserPropertiesUpdateCount++
		}

		if propertiesUpdateList != nil {
			for pName := range *propertiesUpdateList {
				allPropertiesUpdateList[pName] = true
			}
		}

	}

	return &allPropertiesUpdateList, eventPropertieUpdateCount, eventUserPropertiesUpdateCount, latestUserPropertiesUpdateCount, nil
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

func backfillEventPropertiesIfRequired(projectID uint64, userID string, eventID string, properties *postgres.Jsonb, datetimeProperties *map[string]bool, wetRun bool) (*map[string]bool, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "event_id": eventID, "datetime_properties": datetimeProperties})
	if projectID == 0 || userID == "" || eventID == "" || properties == nil {
		return nil, errors.New("missing required field")
	}

	var eventProperties map[string]interface{}
	err := json.Unmarshal(properties.RawMessage, &eventProperties)
	if err != nil {
		return nil, err
	}

	if eventProperties == nil || len(eventProperties) < 1 {
		return nil, errors.New("empty map found")
	}

	logCtx = logCtx.WithFields(log.Fields{"event_properties": eventProperties})

	newEventProperties, propertiesUpdateList, isUpdateRequired, err := updateDatetimePropertiesToUnix(&eventProperties, datetimeProperties)
	if err != nil {
		logCtx.WithFields(log.Fields{"event_properties": eventProperties}).WithError(err).Error("Failed to update datetime properties to unix.")
		return nil, err
	}

	if !isUpdateRequired {
		return nil, nil
	}

	updatedEventPropertiesJsonb, err := util.EncodeToPostgresJsonb(newEventProperties)
	if err != nil {
		return nil, err
	}

	if wetRun {
		status := store.GetStore().OverwriteEventProperties(projectID, userID, eventID, updatedEventPropertiesJsonb)
		if status != http.StatusAccepted {
			return nil, errors.New("failed to update event properties")
		}
	}

	return propertiesUpdateList, nil
}

func backFillEventUserPropertiesIfRequired(projectID uint64, userID string, eventUserPropertiesID string, datetimeProperties *map[string]bool, wetRun bool) (*map[string]bool, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "event_properties_id": eventUserPropertiesID})
	if projectID == 0 || userID == "" || eventUserPropertiesID == "" {
		logCtx.Error("Missing required fields.")
		return nil, errors.New("missing required field")
	}

	properties, status := store.GetStore().GetUserProperties(projectID, userID, eventUserPropertiesID)
	if status != http.StatusFound {
		logCtx.Error("Failed to get event user properties.")
		return nil, errors.New("failed to get event user properties")
	}

	var eventUserProperties map[string]interface{}
	err := json.Unmarshal(properties.RawMessage, &eventUserProperties)
	if err != nil {
		return nil, err
	}

	if eventUserProperties == nil || len(eventUserProperties) < 1 {
		logCtx.Error("Empty map found.")
		return nil, errors.New("empty map found")
	}

	newEventUserProperties, propertiesUpdateList, isUpdateRequired, err := updateDatetimePropertiesToUnix(&eventUserProperties, datetimeProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update datetime properties to unix.")
		return nil, err
	}

	if !isUpdateRequired {
		return nil, nil
	}

	eventUserPropertiesJsonb, err := util.EncodeToPostgresJsonb(newEventUserProperties)
	if err != nil {
		return nil, err
	}

	if wetRun {
		status = store.GetStore().OverwriteUserProperties(projectID, userID, eventUserPropertiesID, eventUserPropertiesJsonb)
		if status != http.StatusAccepted {
			return nil, errors.New("failed to overwrite event user properties")
		}
	}

	return propertiesUpdateList, nil
}

func updateEventPropertiesAndEventUserProperties(projectID uint64, event *model.Event, datetimeProperties *map[string]bool, wetRun bool) (*map[string]bool, bool, bool, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "datetime_properties": datetimeProperties, "event_id": event.ID})
	if projectID == 0 || event == nil || datetimeProperties == nil || len(*datetimeProperties) < 1 {
		logCtx.Error("Missing required fields.")
		return nil, false, false, errors.New("missing required fields")
	}

	allPropertiesUpdateList := make(map[string]bool)
	propertiesUpdateList, err := backfillEventPropertiesIfRequired(projectID, event.UserId, event.ID, &event.Properties, datetimeProperties, wetRun)
	if err != nil {
		logCtx.WithError(err).Error("Failed to backfill event propeties.")
		return nil, false, false, err
	}

	eventPropertiesUpdated := propertiesUpdateList != nil

	if propertiesUpdateList != nil {
		for pName := range *propertiesUpdateList {
			allPropertiesUpdateList[pName] = true
		}
	}

	propertiesUpdateList, err = backFillEventUserPropertiesIfRequired(projectID, event.UserId, event.UserPropertiesId, datetimeProperties, wetRun)
	if err != nil {
		logCtx.WithError(err).Error("Failed to backfill event user propeties.")
		return nil, false, false, err
	}

	eventUserPropertiesUpdated := propertiesUpdateList != nil

	if propertiesUpdateList != nil {
		for pName := range *propertiesUpdateList {
			allPropertiesUpdateList[pName] = true
		}
	}

	return &allPropertiesUpdateList, eventPropertiesUpdated, eventUserPropertiesUpdated, nil
}

func updateLatesUserPropeties(projectID uint64, userID string, datetimeProperties *map[string]bool, wetRun bool) (*map[string]bool, error) {
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
		status = store.GetStore().OverwriteUserProperties(projectID, userID, user.PropertiesId, userPropertiesJsonb)
		if status != http.StatusAccepted {
			return nil, errors.New("failed to overwrite event user properties")
		}
	}

	return propertiesUpdateList, nil
}
