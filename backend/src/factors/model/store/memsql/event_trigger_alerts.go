package memsql

import (
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	E "factors/event_match"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	ListLimit          = 1000
	AlertCreationLimit = 100
	SortedSetCacheKey  = "ETA:pid"
	oneDayInSeconds    = 24 * 60 * 60
)

func (store *MemSQL) GetAllEventTriggerAlertsByProject(projectID int64) ([]model.EventTriggerAlertInfo, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	alerts := make([]model.EventTriggerAlert, 0)

	err := db.Table("event_trigger_alerts").
		Where("project_id = ? AND is_deleted = ?", projectID, false).
		Order("created_at DESC").Limit(ListLimit).Find(&alerts).Error
	if err != nil {
		log.WithError(err).Error("Failed to fetch rows from pathanalysis table for project")
		return nil, http.StatusInternalServerError
	}

	if len(alerts) == 0 {
		return nil, http.StatusFound
	}

	alertArray := store.convertEventTriggerAlertToEventTriggerAlertInfo(alerts)
	return alertArray, http.StatusFound
}

func (store *MemSQL) GetEventTriggerAlertByID(id string) (*model.EventTriggerAlert, int) {
	logFields := log.Fields{
		"id": id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var alert model.EventTriggerAlert
	err := db.Table("event_trigger_alerts").
		Where("id = ? AND is_deleted = ?", id, false).
		Order("created_at DESC").Limit(ListLimit).Find(&alert).Error
	if err != nil {
		log.WithError(err).Error("Failed to fetch rows from pathanalysis table for project")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &alert, http.StatusFound
}

func (store *MemSQL) convertEventTriggerAlertToEventTriggerAlertInfo(list []model.EventTriggerAlert) []model.EventTriggerAlertInfo {

	res := make([]model.EventTriggerAlertInfo, 0)

	for _, obj := range list {
		var alert model.EventTriggerAlertConfig
		err := U.DecodePostgresJsonbToStructType(obj.EventTriggerAlert, &alert)
		if err != nil {
			log.WithError(err).Error("Problem deserializing pathanalysis query.")
			return nil
		}
		deliveryOption := ""
		if alert.Slack {
			deliveryOption += "Slack "
		}
		if alert.Webhook {
			if deliveryOption == "" {
				deliveryOption += "Webhook"
			} else {
				deliveryOption += "& Webhook"
			}
		}
		e := model.EventTriggerAlertInfo{
			ID:                obj.ID,
			Title:             obj.Title,
			DeliveryOptions:   deliveryOption,
			EventTriggerAlert: &alert,
		}
		res = append(res, e)
	}
	return res
}

func (store *MemSQL) DeleteEventTriggerAlert(projectID int64, id string) (int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	if projectID == 0 {
		return http.StatusBadRequest, "Invalid project ID"
	}
	modTime := gorm.NowFunc()

	err := db.Model(&model.EventTriggerAlert{}).Where("id = ? AND project_id = ?", id, projectID).
		Update(map[string]interface{}{"is_deleted": true, "updated_at": modTime}).Error

	if err != nil {
		return http.StatusInternalServerError, "Failed to delete saved entity"
	}
	return http.StatusAccepted, ""
}

func (store *MemSQL) CreateEventTriggerAlert(userID string, projectID int64, alertConfig *model.EventTriggerAlertConfig) (*model.EventTriggerAlert, int, string) {
	logFields := log.Fields{
		"project_id":          projectID,
		"event_trigger_alert": alertConfig,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	var alert model.EventTriggerAlert

	if alertCreationLimitExceeded(projectID) {
		return nil, http.StatusConflict, "Alerts limit reached"
	}

	if isDuplicateAlertTitle(projectID, alertConfig.Title) {
		return nil, http.StatusConflict, "Alert already exist"
	}

	transTime := gorm.NowFunc()
	id := U.GetUUID()

	trigger, err := U.EncodeStructTypeToPostgresJsonb(*alertConfig)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("TriggerAlert conversion to Jsonb failed")
		return nil, http.StatusInternalServerError, "TiggerAlert conversion to Jsonb failed"
	}

	alert = model.EventTriggerAlert{
		ID:                id,
		ProjectID:         projectID,
		Title:             alertConfig.Title,
		EventTriggerAlert: trigger,
		CreatedBy:         userID,
		CreatedAt:         transTime,
		UpdatedAt:         transTime,
		IsDeleted:         false,
	}

	if err := db.Create(&alert).Error; err != nil {
		log.WithFields(logFields).WithError(err).Error("Create Failed")
		return nil, http.StatusInternalServerError, "Create Failed in db"
	}

	return &alert, http.StatusCreated, ""
}

func alertCreationLimitExceeded(projectID int64) bool {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var count int
	if err := db.Model(&model.EventTriggerAlert{}).Where("project_id = ?", projectID).
		Where("is_deleted = ?", false).
		Count(&count).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false
		}
	}
	return count > AlertCreationLimit
}
func isDuplicateAlertTitle(projectID int64, title string) bool {
	logFields := log.Fields{
		"project_id":          projectID,
		"event_trigger_alert": title,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var alerts []model.EventTriggerAlert
	if err := db.Where("project_id = ?", projectID).
		Where("is_deleted = ?", false).
		Where("title = ?", title).
		Find(&alerts).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false
		}
	}
	for _, obj := range alerts {
		if obj.Title == title {
			return true
		}
	}
	return false
}

func (store *MemSQL) GetEventTriggerAlertsByEvent(projectId int64, id string) ([]model.EventTriggerAlert, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"prefix":     id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var eventAlerts []model.EventTriggerAlert
	var eventName model.EventName

	db := C.GetServices().Db

	if err := db.Where("project_id = ? AND id = ?", projectId, id).
		Find(&eventName).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "id": id}).WithError(err).Error(
			"event_name not found")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		return nil, http.StatusInternalServerError
	}

	if err := db.Where("project_id = ? AND is_deleted = 0", projectId).
		Where("JSON_EXTRACT_STRING(event_trigger_alert, 'event') LIKE ?", fmt.Sprintf("%s", eventName.Name)).
		Find(&eventAlerts).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "event": eventName.Name}).WithError(err).Error(
			"filtering eventName failed on GetFilterEventNamesByEvent")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		return nil, http.StatusInternalServerError
	}

	if len(eventAlerts) == 0 {
		return nil, http.StatusNotFound
	}

	return eventAlerts, http.StatusFound
}

func (store *MemSQL) MatchEventTriggerAlertWithTrackPayload(projectId int64, eventNameId string, eventProps, userProps *postgres.Jsonb) (*[]model.EventTriggerAlert, int) {
	logFields := log.Fields{
		"project_id":       projectId,
		"event_name":       eventNameId,
		"event_properties": eventProps,
		"user_properties":  userProps,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	log.Info("Inside Match function of event_trigger_alerts.")
	alerts, errCode := store.GetEventTriggerAlertsByEvent(projectId, eventNameId)
	if errCode != http.StatusFound || alerts == nil {
		//log.WithFields(logFields).Error("GetEventTriggerAlertsByEvent failure inside Match function.")
		return nil, errCode
	}

	var userPropMap, eventPropMap *map[string]interface{}
	if userProps != nil {
		userPropMap, _ = U.DecodePostgresJsonb(userProps)
	}
	if eventProps != nil {
		eventPropMap, _ = U.DecodePostgresJsonb(eventProps)
	}

	var matchedAlerts []model.EventTriggerAlert
	for _, alert := range alerts {
		var config model.EventTriggerAlertConfig
		err := U.DecodePostgresJsonbToStructType(alert.EventTriggerAlert, &config)
		if err != nil {
			log.WithError(err).Error("Jsonb decoding to struct failure")
			return nil, http.StatusInternalServerError
		}

		if E.EventMatchesFilterCriterionList(*userPropMap, *eventPropMap, E.MapFilterProperties(config.Filter)) {
			matchedAlerts = append(matchedAlerts, alert)
		}
	}
	if len(matchedAlerts) == 0 {
		log.WithFields(logFields).Info("Match function did not find anything in event_trigger_alerts")
		return nil, http.StatusNotFound
	}
	return &matchedAlerts, http.StatusFound
}

func AddAlertToCache(alert *model.EventTriggerAlertConfig, event *model.Event, key *cacheRedis.Key) (int, error) {
	logFields := log.Fields{
		"event_trigger_alert": alert,
		"CacheKey":            key,
	}
	log.Info("Inside AddAlertToCache function.")
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	messageProperties := make([]model.QueryGroupByProperty, 0)
	err := U.DecodePostgresJsonbToStructType(alert.MessageProperty, &messageProperties)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Jsonb decoding to struct failure")
		return http.StatusInternalServerError, err
	}
	var userPropMap, eventPropMap *map[string]interface{}
	if event.UserProperties != nil {
		userPropMap, _ = U.DecodePostgresJsonb(event.UserProperties)
	}
	if len(event.Properties.RawMessage) != 0 {
		eventPropMap, _ = U.DecodePostgresJsonb(&event.Properties)
	}

	propMap := make(U.PropertiesMap, 0)
	for _, messageProperty := range messageProperties {
		p := messageProperty.Property
		if messageProperty.Entity == "user" {
			propMap[p] = (*userPropMap)[p]
		} else {
			propMap[p] = (*eventPropMap)[p]
		}
	}

	message := model.EventTriggerAlertMessage{
		Title:           alert.Title,
		Event:           alert.Event,
		MessageProperty: propMap,
		Message:         alert.Message,
	}
	cachePackage := model.CachedEventTriggerAlert{
		Message: message,
	}

	log.WithFields(logFields).Info("SetCacheForEventTriggerAlert function inside AddAlertToCache.")
	err = model.SetCacheForEventTriggerAlert(key, &cachePackage)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("setting cache failed inside AddAlertToCache")
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

func getSortedSetCacheKey(projectId int64) (*cacheRedis.Key, error) {
	prefix := fmt.Sprintf("%s:%d", SortedSetCacheKey, projectId)
	key, err := cacheRedis.NewKeyWithOnlyPrefix(prefix)
	if err != nil {
		log.WithError(err).Error("Cannot adding only prefix key")
		return nil, err
	}
	return key, err
}

func (store *MemSQL) CacheEventTriggerAlert(alert *model.EventTriggerAlert, event *model.Event) bool {

	//Adding alert to cache
	//If the counterKey is present, then
	//INCR the counter key
	//Else set the counter key with one day of expiry
	//If the counter key has count less than daily limit, then
	//Get sorted set keys from where all the alert keys for a particular projectID are retrieved
	//Get the alert and counter key as well
	//Add the alert key to the sorted set and cache
	//Else return

	log.Info("Inside CacheEventTriggerAlert function.")
	log.Info(fmt.Printf("%+v\n", *alert))

	var eta model.EventTriggerAlertConfig
	if err := U.DecodePostgresJsonbToStructType(alert.EventTriggerAlert, &eta); err != nil {
		log.WithError(err).Error("Error in decoding jsonb to struct type. InternalServerError")
		return false
	}

	tt := time.Now()
	timestamp := tt.Unix()
	date := tt.UTC().Format(U.DATETIME_FORMAT_YYYYMMDD)

	counterKey, err := model.GetEventTriggerAlertCacheCounterKey(event.ProjectId, alert.ID, date)
	if err != nil {
		log.WithError(err).Error("error while getting cache counter Key")
		return false
	}

	cc, err := cacheRedis.IncrPersistentBatch(counterKey)
	if err != nil {
		log.WithError(err).Error("error while getting count from cache counter Key")
		return false
	}
	count := cc[0]

	if count == 1 {
		_, err := cacheRedis.SetExpiryPersistent(counterKey, oneDayInSeconds)
		if err != nil {
			log.WithError(err).Error("error while setting expiry for cache counter Key")
			return false
		}
	}

	addToCache := true
	if (eta.RepeatAlerts && count > 1) || (eta.Notifications && count > eta.AlertLimit) {
		addToCache = false
	}

	if addToCache {
		ssKey, err := getSortedSetCacheKey(event.ProjectId)
		if err != nil {
			log.WithError(err).Error("error while getting sorted set cache Key")
			return false
		}

		key, err := model.GetEventTriggerAlertCacheKey(event.ProjectId, timestamp, alert.ID)
		if err != nil {
			log.WithError(err).Error("error while getting cache Key")
			return false
		}

		ssValue, err := key.Key()
		if err != nil {
			log.WithError(err).Error("error while getting cache Key to string")
			return false
		}

		sortedSetTuple := cacheRedis.SortedSetKeyValueTuple{
			Key:   ssKey,
			Value: ssValue,
		}

		_, err = cacheRedis.ZincrPersistentBatch(true, sortedSetTuple)
		if err != nil {
			log.WithError(err).Error("error while getting INCR value")
		}

		successCode, err := AddAlertToCache(&eta, event, key)
		if err != nil || successCode != http.StatusCreated {
			log.WithFields(log.Fields{"project_id": event.ProjectId,
				"event_trigger_alert": alert, log.ErrorKey: err}).Error("Failed to send alert.")
			return false
		}
	} else {
		log.WithFields(log.Fields{"project_id": event.ProjectId, "event_trigger_alert": *alert}).
			Info("Alert was not sent for current EventTriggerAlert as daily AlertLimit has been reached.")
	}

	return true
}

func (store *MemSQL) UpdateEventTriggerAlertField(projectID int64, id string, field map[string]interface{}) (int, error) {
	logFields := log.Fields{
		"project_id":             projectID,
		"event_trigger_alert_id": id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var alert model.EventTriggerAlert

	db := C.GetServices().Db
	if err := db.Model(&alert).Where("id = ? AND project_id = ? AND is_deleted = 0", id, projectID).
		Update(field).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectID, "event_trigger_alert": id}).WithError(err).Error(
			"Failed to fetch the required event trigger alert")

		return http.StatusInternalServerError, err
	}
	return http.StatusAccepted, nil
}
