package memsql

import (
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	ListLimit          = 1000
	AlertCreationLimit = 100
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
		Order("created_at").Limit(ListLimit).Find(&alerts).Error
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
		Order("created_at").Limit(ListLimit).Find(&alert).Error
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
	var alert model.EventTriggerAlertConfig

	for _, obj := range list {
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

	trigger, err := U.EncodeStructTypeToPostgresJsonb(alertConfig)
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
		Where("JSON_EXTRACT_STRING(event_trigger_alert, 'event') LIKE ?", fmt.Sprintf("%s%%", eventName.Name)).
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

func MapFilterPropertiesToEventFilterCriterion(qp []model.QueryProperty) []U.PropertiesMap {
	log.Info("Converting filters to propertiesMap")
	filterMap := make([]U.PropertiesMap, 0)
	FilterEq := make(U.PropertiesMap)
	FilterNotEq := make(U.PropertiesMap)
	FilterCont := make(U.PropertiesMap)
	FilterNotCont := make(U.PropertiesMap)

	for _, prop := range qp {
		if prop.Operator == "equals" {
			FilterEq[prop.Property] = prop.Value
		} else if prop.Operator == "notEqual" {
			FilterNotEq[prop.Property] = prop.Value
		} else if prop.Operator == "contains" {
			FilterCont[prop.Property] = prop.Value
		} else if prop.Operator == "notContains" {
			FilterNotCont[prop.Property] = prop.Value
		}
	}
	filterMap = append(filterMap, FilterEq, FilterNotEq, FilterCont, FilterCont)
	return filterMap
}

func matchFilterProps(filter []U.PropertiesMap, userProps, eventProps U.PropertiesMap) bool {
	log.Info("inside matchProps function")
	if len(filter) == 0 {
		return true
	}
	log.Info(fmt.Printf("%+v\n", filter))
	for key, prop := range filter[0] {
		if userProps[key] == prop || eventProps[key] == prop {
			log.Info("matchProps function found a mismatch in one of the property. Returning.")
			return false
		}
	}
	for key, prop := range filter[1] {
		if userProps[key] != prop && eventProps[key] != prop {
			log.Info("matchProps function found a mismatch in one of the property. Returning.")
			return false
		}
	}
	//TODO: Match 'contains' and 'notContains' operator type filters
	return true
}

func (store *MemSQL) MatchEventTriggerAlertWithTrackPayload(projectId int64, eventNameId string, eventProps, userProps *postgres.Jsonb) (*model.EventTriggerAlert, int) {
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
		log.WithFields(logFields).Error("GetEventTriggerAlertsByEvent failure inside Match function.")
		return nil, errCode
	}

	var userPropMap, eventPropMap U.PropertiesMap
	err := U.DecodePostgresJsonbToStructType(eventProps, &userPropMap)
	if err != nil {
		log.WithError(err).Error("Jsonb decoding to struct failure")
		return nil, http.StatusInternalServerError
	}
	err = U.DecodePostgresJsonbToStructType(userProps, &eventPropMap)
	if err != nil {
		log.WithError(err).Error("Jsonb decoding to struct failure")
		return nil, http.StatusInternalServerError
	}

	log.Info(fmt.Printf("%+v\n", userPropMap))
	log.Info(fmt.Printf("%+v\n", eventPropMap))

	for _, alert := range alerts {
		var config model.EventTriggerAlertConfig
		err := U.DecodePostgresJsonbToStructType(alert.EventTriggerAlert, &config)
		if err != nil {
			log.WithError(err).Error("Jsonb decoding to struct failure")
			return nil, http.StatusInternalServerError
		}
		filterProps := MapFilterPropertiesToEventFilterCriterion(config.Filter)
		matched := matchFilterProps(filterProps, userPropMap, eventPropMap)
		if matched {
			log.WithFields(logFields).Info("Match found for the event_trigger_alerts")
			return &alert, http.StatusFound
		}
	}
	log.WithFields(logFields).Info("Match function did not find anything in event_trigger_alerts")
	return nil, http.StatusNotFound
}

func AddAlertToCache(projectID int64, userID string, alert *model.EventTriggerAlert, counterPresent bool) (int, error) {
	logFields := log.Fields{
		"project_id":          projectID,
		"event_trigger_alert": alert,
	}
	log.Info("Inside AddAlertToCache function.")
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var config model.EventTriggerAlertConfig
	err := U.DecodePostgresJsonbToStructType(alert.EventTriggerAlert, &config)
	if err != nil {
		log.WithError(err).Error("Unable to decide json into struct type")
		return http.StatusInternalServerError, err
	}

	message := model.EventTriggerAlertMessage{
		Title:           alert.Title,
		Event:           config.Event,
		MessageProperty: string(config.MessageProperty.RawMessage),
		Message:         config.Message,
	}

	cachePackage := model.CachedEventTriggerAlert{
		AlertID:   alert.ID,
		Timestamp: time.Now(),
		Message:   message,
	}

	log.WithFields(logFields).Info("SetCacheForEventTriggerAlert function inside AddAlertToCache.")
	err = model.SetCacheForEventTriggerAlert(projectID, userID, &cachePackage)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("setting cache failed inside AddAlertToCache")
		return http.StatusInternalServerError, err
	}

	err = model.SetCacheCounterForEventTriggerAlert(projectID, userID, alert.ID, counterPresent)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("setting cache failed inside AddAlertToCache")
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

func GetCacheKeyForAlert(alert *model.EventTriggerAlert, counter bool) (*cacheRedis.Key, error) {
	log.Info("Inside GetCacheKey function")

	var cacheKey *cacheRedis.Key
	var err error
	date := time.Now().UTC().Format(U.DATETIME_FORMAT_YYYYMMDD)

	if counter {
		cacheKey, err = model.GetEventTriggerAlertCacheCounterKey(alert.ProjectID, alert.CreatedBy, alert.ID, date)
		if err != nil {
			log.WithError(err).Error("Counter CacheKey getting function failure.")
			return nil, err
		}
	} else {
		cacheKey, err = model.GetEventTriggerAlertCacheKey(alert.ProjectID, alert.CreatedBy, alert.ID, date)
		if err != nil {
			log.WithError(err).Error("CacheKey getting function failure.")
			return nil, err
		}
	}
	return cacheKey, nil
}

func checkCacheForAlert(alert *model.EventTriggerAlert) (*cacheRedis.Key, bool) {
	log.Info("Inside checkCacheForAlert function")

	cacheKey, err := GetCacheKeyForAlert(alert, true)
	log.Info("Control returning to checkCacheForAlert")
	if err != nil {
		log.WithError(err).Error("CacheKey getting function failure.")
		return nil, false
	}
	kk, err := cacheKey.Key()
	if err != nil {
		log.Info("cachekey string conversion failure")
	}
	log.Info("CacheKey: ", kk)
	exists, err := cacheRedis.ExistsPersistent(cacheKey)
	if err != nil {
		log.WithError(err).Error("CacheKey existence failure.")
		return nil, false
	}
	log.Info("cacheRedisKey successfully retrieved. Key exists is ", exists)
	return cacheKey, exists
}

func (store *MemSQL) SendEventTriggerAlert(alert *model.EventTriggerAlert, projectId int64, name, userID string) bool {

	log.Info("Inside SendEventTriggerAlert function.")
	log.Info(fmt.Printf("%+v\n", *alert))

	var eta model.EventTriggerAlertConfig
	if err := U.DecodePostgresJsonbToStructType(alert.EventTriggerAlert, &eta); err != nil {
		log.WithError(err).Error("Error in decoding jsonb to struct type. InternalServerError")
		return false
	}

	log.Info("Checking cache for alert's existence.")
	cacheKey, status := checkCacheForAlert(alert)
	var count int64

	if status {
		limit, err := cacheRedis.GetPersistent(cacheKey)
		if err != nil {
			log.WithError(err).Error("Redis error while fetching counter in event_trigger_alerts")
		}

		count, err = strconv.ParseInt(limit, 10, 64)
		if err != nil {
			count = 0
			log.WithError(err).Error("string conversion to int64 failure. Defaulting to 0.")
		}
	}

	if count <= eta.AlertLimit {
		successCode, err := AddAlertToCache(projectId, userID, alert, status)
		if err != nil || successCode != http.StatusCreated {
			log.WithFields(log.Fields{"project_id": projectId,
				"event_trigger_alert": alert, log.ErrorKey: err}).Error("Failed to send alert.")
			return false
		}
	} else {
		log.WithFields(log.Fields{"project_id": projectId,
			"event_trigger_alert": alert}).Error("Alert was not sent for current EventTriggerAlert as daily AlertLimit has been reached.")
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
