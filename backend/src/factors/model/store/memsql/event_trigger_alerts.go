package memsql

import (
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"
	"io/ioutil"
	E "factors/event_match"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"encoding/json"
)

const (
	ListLimit          = 1000
	AlertCreationLimit = 100
	SortedSetCacheKey  = "ETA:pid"
	CoolDownPrefix     = "ETA:CoolDown"
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
	transTime := gorm.NowFunc()
	id := U.GetUUID()

	if alertCreationLimitExceeded(projectID) {
		return nil, http.StatusConflict, "Alerts limit reached"
	}

	if isDuplicateAlertTitle(projectID, alertConfig.Title, id) {
		return nil, http.StatusConflict, "Alert already exist"
	}

	for _, filter := range (*alertConfig).Filter {
		if(filter.Operator == model.InList){
			// Get the cloud file that is there for the reference value
			path, file := C.GetCloudManager(projectID, true).GetListReferenceFileNameAndPathFromCloud(projectID, filter.Value)
			reader, err := C.GetCloudManager(projectID, true).Get(path, file)
			if(err != nil){
				log.WithFields(logFields).WithError(err).Error("List File Missing")
				return nil, http.StatusInternalServerError, "List File Missing"
			}
			valuesInFile := make([]string, 0)
			data, err := ioutil.ReadAll(reader)
			if(err != nil){
				log.WithFields(logFields).WithError(err).Error("File reader failed")
				return nil, http.StatusInternalServerError, "File reader failed"
			}
			err = json.Unmarshal(data, &valuesInFile)
			if(err != nil){
				log.WithFields(logFields).WithError(err).Error("list data unmarshall failed")
				return nil, http.StatusInternalServerError, "list data unmarshall failed"
			}
			cacheKeyList, err := model.GetListCacheKey(projectID, filter.Value)
			if(err != nil){
				log.WithFields(logFields).WithError(err).Error("get cache key failed")
				return nil, http.StatusInternalServerError, "get cache key failed"
			}
			for _, value := range valuesInFile {
				err = cacheRedis.ZAddPersistent(cacheKeyList, value, 0)
				if(err != nil){
					log.WithFields(logFields).WithError(err).Error("failed to add new values to sorted set")
					return nil, http.StatusInternalServerError, "failed to add new values to sorted set"
				}
			}
		}
	}

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

func isDuplicateAlertTitle(projectID int64, title, id string) bool {
	logFields := log.Fields{
		"project_id":          projectID,
		"event_trigger_alert": title,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var alerts []model.EventTriggerAlert
	if err := db.Where("project_id = ?", projectID).
		Where("is_deleted = ?", false).
		Not("id = ?", id).
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
		Where("JSON_EXTRACT_STRING(event_trigger_alert, 'event') LIKE ?", eventName.Name).
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

func (store *MemSQL) MatchEventTriggerAlertWithTrackPayload(projectId int64, eventNameId string, eventProps, userProps *postgres.Jsonb, UpdatedEventProps *postgres.Jsonb, isUpdate bool) (*[]model.EventTriggerAlert, int) {
	logFields := log.Fields{
		"project_id":       projectId,
		"event_name":       eventNameId,
		"event_properties": eventProps,
		"user_properties":  userProps,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	alerts, errCode := store.GetEventTriggerAlertsByEvent(projectId, eventNameId)
	if errCode != http.StatusFound || alerts == nil {
		//log.WithFields(logFields).Error("GetEventTriggerAlertsByEvent failure inside Match function.")
		return nil, errCode
	}

	var userPropMap, eventPropMap, updatedEventProps *map[string]interface{}
	if userProps != nil {
		userPropMap, _ = U.DecodePostgresJsonb(userProps)
	}
	if eventProps != nil {
		eventPropMap, _ = U.DecodePostgresJsonb(eventProps)
	}
	if UpdatedEventProps != nil {
		updatedEventProps, _ = U.DecodePostgresJsonb(UpdatedEventProps)
	}

	var matchedAlerts []model.EventTriggerAlert
	for _, alert := range alerts {
		var config model.EventTriggerAlertConfig
		err := U.DecodePostgresJsonbToStructType(alert.EventTriggerAlert, &config)
		if err != nil {
			log.WithError(err).Error("Jsonb decoding to struct failure")
			return nil, http.StatusInternalServerError
		}
		if isUpdate {
			if len(*updatedEventProps) == 0 {
				continue
			} else {
				isPropertyInFilterUpdated := false
				for _, fil := range config.Filter {
					_, exists := (*updatedEventProps)[fil.Property]
					if fil.Entity == "event" && exists {
						isPropertyInFilterUpdated = true
					}
				}
				if !isPropertyInFilterUpdated {
					continue
				}
			}
		}
		if E.EventMatchesFilterCriterionList(projectId, *userPropMap, *eventPropMap, E.MapFilterProperties(config.Filter)) {
			matchedAlerts = append(matchedAlerts, alert)
		}
	}
	if len(matchedAlerts) == 0 {
		log.WithFields(logFields).Info("Match function did not find anything in event_trigger_alerts")
		return nil, http.StatusNotFound
	}
	return &matchedAlerts, http.StatusFound
}

func (store *MemSQL) getDisplayNamesForEP(projectId int64, eventName string) map[string]string {

	_, displayNames := store.GetDisplayNamesForAllEventProperties(projectId, eventName)
	standardPropertiesAllEvent := U.STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES
	displayNamesOp := make(map[string]string)
	for property, displayName := range standardPropertiesAllEvent {
		displayNamesOp[property] = strings.Title(displayName)
	}
	if eventName == U.EVENT_NAME_SESSION {
		standardPropertiesSession := U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES
		for property, displayName := range standardPropertiesSession {
			displayNamesOp[property] = strings.Title(displayName)
		}
	}
	for property, displayName := range displayNames {
		displayNamesOp[property] = strings.Title(displayName)
	}

	_, displayNames = store.GetDisplayNamesForObjectEntities(projectId)
	for property, displayName := range displayNames {
		displayNamesOp[property] = strings.Title(displayName)
	}

	dupCheck := make(map[string]bool)
	for _, name := range displayNamesOp {
		_, exists := dupCheck[name]
		if exists {
			log.Warning(fmt.Sprintf("Duplicate display name %s", name))
		}
		dupCheck[name] = true
	}

	return displayNamesOp
}

func (store *MemSQL) getDisplayNamesForUP(projectId int64) map[string]string {

	_, displayNames := store.GetDisplayNamesForAllUserProperties(projectId)
	standardProperties := U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES
	displayNamesOp := make(map[string]string)
	for property, displayName := range standardProperties {
		displayNamesOp[property] = strings.Title(displayName)
	}
	for property, displayName := range displayNames {
		displayNamesOp[property] = strings.Title(displayName)
	}

	_, displayNames = store.GetDisplayNamesForObjectEntities(projectId)
	for property, displayName := range displayNames {
		displayNamesOp[property] = strings.Title(displayName)
	}

	dupCheck := make(map[string]bool)
	for _, name := range displayNamesOp {
		_, exists := dupCheck[name]
		if exists {
			log.Warningf(fmt.Sprintf("Duplicate display name %s", name))
		}
		dupCheck[name] = true
	}

	return displayNamesOp
}

func (store *MemSQL) AddAlertToCache(alert *model.EventTriggerAlertConfig, msgProps *U.PropertiesMap, key *cacheRedis.Key) (int, error) {
	logFields := log.Fields{
		"event_trigger_alert": alert,
		"CacheKey":            key,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	message := model.EventTriggerAlertMessage{
		Title:           alert.Title,
		Event:           U.CreateVirtualDisplayName(alert.Event),
		MessageProperty: *msgProps,
		Message:         alert.Message,
	}

	cachePackage := model.CachedEventTriggerAlert{
		Message: message,
	}

	err := model.SetCacheForEventTriggerAlert(key, &cachePackage)
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

func getDisplayLikePropValue(typ string, exi bool, value interface{}) interface{} {

	var res interface{}
	if exi {
		if typ == "datetime" {
			val, ok := value.(int64)
			if !ok {
				val = int64(value.(float64))
			}
			res = U.GetDateOnlyHyphenFormatFromTimestampZ(val)
		} else {
			res = U.GetPropertyValueAsString(value)
		}
	}

	return res
}

func (store *MemSQL) GetMessageAndBreakdownPropertiesMap(event *model.Event, alert *model.EventTriggerAlertConfig) (U.PropertiesMap, map[string]interface{}, error) {
	logFields := log.Fields{
		"event_trigger_alert": *alert,
		"event":               *event,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	messageProperties := make([]model.QueryGroupByProperty, 0)
	if alert.MessageProperty != nil {
		err := U.DecodePostgresJsonbToStructType(alert.MessageProperty, &messageProperties)
		if err != nil {
			log.WithError(err).Error("Jsonb decoding to struct failure")
			return nil, nil, err
		}
	}

	var userPropMap, eventPropMap *map[string]interface{}
	var err error
	if event.UserProperties != nil {
		userPropMap, err = U.DecodePostgresJsonb(event.UserProperties)
		if err != nil {
			log.WithError(err).Error("Jsonb decoding to propMap failure")
			return nil, nil, err
		}
	}
	if &event.Properties != nil && len(event.Properties.RawMessage) != 0 {
		eventPropMap, err = U.DecodePostgresJsonb(&event.Properties)
		if err != nil {
			log.WithError(err).Error("Jsonb decoding to propMap failure")
			return nil, nil, err
		}
	}

	displayNamesEP := store.getDisplayNamesForEP(event.ProjectId, event.EventNameId)
	//log.Info(fmt.Printf("%+v\n", displayNamesEP))

	displayNamesUP := store.getDisplayNamesForUP(event.ProjectId)
	//log.Info(fmt.Printf("%+v\n", displayNamesUP))

	msgPropMap := make(U.PropertiesMap, 0)
	for _, messageProperty := range messageProperties {
		p := messageProperty.Property
		if messageProperty.Entity == "user" {

			displayName, exists := displayNamesUP[p]
			if !exists {
				displayName = U.CreateVirtualDisplayName(p)
			}
			propVal, exi := (*userPropMap)[p]
			msgPropMap[displayName] = getDisplayLikePropValue(messageProperty.Type, exi, propVal)

		} else if messageProperty.Entity == "event" {
			displayName, exists := displayNamesEP[p]
			if !exists {
				displayName = U.CreateVirtualDisplayName(p)
			}
			propVal, exi := (*eventPropMap)[p]
			msgPropMap[displayName] = getDisplayLikePropValue(messageProperty.Type, exi, propVal)
		} else {
			log.Warn("can not find the message property in user and event prop sets")
		}
	}

	breakdownPropMap := make(map[string]interface{}, 0)
	var breakdownProperties []model.QueryGroupByProperty
	if alert.BreakdownProperties != nil {
		err = U.DecodePostgresJsonbToStructType(alert.BreakdownProperties, &breakdownProperties)
		if err != nil {
			log.WithError(err).Error("breakdownProperty Jsonb decoding to queryGroupByProperty failure")
			return nil, nil, err
		}
	}

	for _, breakdownProperty := range breakdownProperties {
		prop := breakdownProperty.Property
		var value interface{}
		uval, uexists := (*userPropMap)[prop]
		eval, eexists := (*eventPropMap)[prop]

		if uexists {
			value = uval
		} else if eexists {
			value = eval
		} else {
			log.Warn("can not find the breakdown property in user and event prop sets")
		}
		breakdownPropMap[prop] = value
	}
	return msgPropMap, breakdownPropMap, nil
}

func getCacheKeyAndSortedSetTupleAndCheckCoolDownTimeCondition(projectID int64, dontRepeatAlerts bool,
	coolDownTime, unixtime int64, alertID string, breakdownProps *map[string]interface{}) (bool,
	*cacheRedis.Key, cacheRedis.SortedSetKeyValueTuple, error) {

	key, err := model.GetEventTriggerAlertCacheKey(projectID, unixtime, alertID, breakdownProps)
	if err != nil {
		log.WithError(err).Error("error while getting cache Key")
		return false, nil, cacheRedis.SortedSetKeyValueTuple{}, err
	}

	check := true
	if dontRepeatAlerts {
		check, err = isCoolDownTimeExhausted(key, coolDownTime, unixtime)
		if err != nil {
			log.WithError(err).Error("error while getting coolDownTime diff")
			return false, key, cacheRedis.SortedSetKeyValueTuple{}, nil
		}
	}

	ssKey, err := getSortedSetCacheKey(projectID)
	if err != nil {
		log.WithError(err).Error("error while getting sorted set cache Key")
		return false, key, cacheRedis.SortedSetKeyValueTuple{}, err
	}

	ssValue, err := key.Key()
	if err != nil {
		log.WithError(err).Error("error while converting cache Key to string")
		return false, key, cacheRedis.SortedSetKeyValueTuple{}, err
	}

	sortedSetTuple := cacheRedis.SortedSetKeyValueTuple{
		Key:   ssKey,
		Value: ssValue,
	}

	return check, key, sortedSetTuple, nil
}

func isCoolDownTimeExhausted(key *cacheRedis.Key, coolDownTime, unixtime int64) (bool, error) {

	suffix := strings.TrimRight(key.Suffix, fmt.Sprintf(":%d", unixtime))
	cdKey, err := cacheRedis.NewKey(key.ProjectID, CoolDownPrefix, suffix)
	if err != nil {
		log.WithError(err).Error("error while getting redis key")
		return false, err
	}

	count, err := cacheRedis.IncrPersistentBatch(cdKey)
	if err != nil {
		log.WithError(err).Error("error while getting redis key")
		return false, err
	}

	if count[0] == 1 {
		if coolDownTime == 0 {
			coolDownTime = oneDayInSeconds
		}
		expire, err := cacheRedis.SetExpiryPersistent(cdKey, int(coolDownTime))
		if err != nil || expire != 1 {
			log.WithError(err).Error("cannot set expiry for redis key")
		}
	} else {
		return false, nil
	}
	return true, nil
}

func (store *MemSQL) CacheEventTriggerAlert(alert *model.EventTriggerAlert, event *model.Event) bool {

	//Adding alert to cache
	//INCR the counter key
	//If the counterKey is present, continue
	//Else set the counter key with one day of expiry
	//If the counter key has count less than daily limit, then
	//Get sorted set keys from where all the alert keys for a particular projectID are retrieved
	//Get the alert key as well
	//Add the alert key to the sorted set and cache
	//Else return

	logFields := log.Fields{
		"project_id":          alert.ProjectID,
		"event_trigger_alert": *alert,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
	if eta.SetAlertLimit && count > eta.AlertLimit {
		addToCache = false
	}

	if addToCache {
		messageProps, breakdownProps, err := store.GetMessageAndBreakdownPropertiesMap(event, &eta)
		if err != nil {
			log.WithError(err).Error("key and sortedTuple fetching error")
			return false
		}

		check, key, sortedSetTuple, err := getCacheKeyAndSortedSetTupleAndCheckCoolDownTimeCondition(
			event.ProjectId, eta.DontRepeatAlerts, eta.CoolDownTime, timestamp, alert.ID, &breakdownProps)
		if err != nil {
			log.WithError(err).Error("key and sortedTuple fetching error")
			return false
		}

		if !check {
			log.WithFields(log.Fields{"project_id": event.ProjectId, "event_trigger_alert": alert}).
				Info("Alert sending cancelled due to cool down timer")
			return true
		}
		_, err = cacheRedis.ZincrPersistentBatch(true, sortedSetTuple)
		if err != nil {
			log.WithError(err).Error("error while getting zincr")
			return false
		}

		successCode, err := store.AddAlertToCache(&eta, &messageProps, key)
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

	if title, exist := field["title"]; exist {
		if isDuplicateAlertTitle(projectID, title.(string), id) {
			return http.StatusConflict, nil
		}
	}

	var alert model.EventTriggerAlert

	db := C.GetServices().Db
	if err := db.Model(&alert).Where("id = ? AND project_id = ? AND is_deleted = 0", id, projectID).
		Updates(field).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectID, "event_trigger_alert": id}).WithError(err).
			Error("Failed to update event trigger alert")

		return http.StatusInternalServerError, err
	}
	return http.StatusAccepted, nil
}
