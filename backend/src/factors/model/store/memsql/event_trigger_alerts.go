package memsql

import (
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	E "factors/event_match"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"encoding/json"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	ListLimit          = 1000
	AlertCreationLimit = 100
	SortedSetCacheKey  = "ETA:pid"
	CoolDownPrefix     = "ETA:CoolDown"
	oneDayInSeconds    = 24 * 60 * 60
	PoisonTime         = 72             // Hours after which the alert will be paused internally
	DisableTime        = 2 * PoisonTime // Hours after which the alert will not be processed from sdk
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
		log.WithError(err).Error("Failed to fetch rows from event_trigger_alerts table for project")
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
		log.WithError(err).Error("Failed to fetch rows from event_trigger_alerts table for project")
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
			log.WithError(err).Error("Problem deserializing event_trigger_alerts query.")
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
		if alert.Teams {
			if deliveryOption == "" {
				deliveryOption += "Teams"
			} else {
				deliveryOption += "& Teams"
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

func (store *MemSQL) CreateEventTriggerAlert(userID, oldID string, projectID int64, alertConfig *model.EventTriggerAlertConfig, slackTokenUser, teamTokenUser string) (*model.EventTriggerAlert, int, string) {
	logFields := log.Fields{
		"project_id":          projectID,
		"event_trigger_alert": alertConfig,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	var alert model.EventTriggerAlert
	transTime := gorm.NowFunc()
	id := U.GetUUID()

	isValidAlertBody, errCode, errMsg := store.isValidEventTriggerAlertBody(projectID, userID, alertConfig)
	if !isValidAlertBody {
		return nil, errCode, errMsg
	}

	if alertCreationLimitExceeded(projectID) {
		return nil, http.StatusConflict, "Alerts limit reached"
	}

	if isDuplicateAlertTitle(projectID, alertConfig.Title, oldID) {
		return nil, http.StatusConflict, "Alert already exist"
	}

	for _, filter := range (*alertConfig).Filter {
		if filter.Operator == model.InList || filter.Operator == model.NotInList {
			// Get the cloud file that is there for the reference value
			path, file := C.GetCloudManager().GetListReferenceFileNameAndPathFromCloud(projectID, filter.Value)
			reader, err := C.GetCloudManager().Get(path, file)
			if err != nil {
				log.WithFields(logFields).WithError(err).Error("List File Missing")
				return nil, http.StatusInternalServerError, "List File Missing"
			}
			valuesInFile := make([]string, 0)
			data, err := ioutil.ReadAll(reader)
			if err != nil {
				log.WithFields(logFields).WithError(err).Error("File reader failed")
				return nil, http.StatusInternalServerError, "File reader failed"
			}
			err = json.Unmarshal(data, &valuesInFile)
			if err != nil {
				log.WithFields(logFields).WithError(err).Error("list data unmarshall failed")
				return nil, http.StatusInternalServerError, "list data unmarshall failed"
			}
			cacheKeyList, err := model.GetListCacheKey(projectID, filter.Value)
			if err != nil {
				log.WithFields(logFields).WithError(err).Error("get cache key failed")
				return nil, http.StatusInternalServerError, "get cache key failed"
			}
			for _, value := range valuesInFile {
				err = cacheRedis.ZAddPersistent(cacheKeyList, value, 0)
				if err != nil {
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
		ID:                       id,
		ProjectID:                projectID,
		Title:                    alertConfig.Title,
		EventTriggerAlert:        trigger,
		CreatedBy:                userID,
		CreatedAt:                transTime,
		UpdatedAt:                transTime,
		IsDeleted:                false,
		SlackChannelAssociatedBy: slackTokenUser,
		TeamsChannelAssociatedBy: teamTokenUser,
		InternalStatus:           model.Active,
	}

	if err := db.Create(&alert).Error; err != nil {
		log.WithFields(logFields).WithError(err).Error("Create Failed")
		return nil, http.StatusInternalServerError, "Create Failed in db"
	}

	return &alert, http.StatusCreated, ""
}

func (store *MemSQL) isValidEventTriggerAlertBody(projectID int64, agentID string, alert *model.EventTriggerAlertConfig) (bool, int, string) {

	if alert.Title == "" {
		return false, http.StatusBadRequest, "title can not be empty"
	}
	if alert.Event == "" {
		return false, http.StatusBadRequest, "event can not be empty"
	}
	if !alert.Slack && !alert.Webhook && !alert.Teams {
		return false, http.StatusBadRequest, "Choose atleast one delivery option"
	}
	if alert.Slack && (alert.SlackChannels == nil || U.IsEmptyPostgresJsonb(alert.SlackChannels)) {
		return false, http.StatusBadRequest, "Slack channel not selected"
	}
	isSlackIntegrated, errCode := store.IsSlackIntegratedForProject(projectID, agentID)
	if errCode != http.StatusOK {
		log.WithFields(log.Fields{"agentUUID": agentID, "event_trigger_alert": alert}).Error("failed to check slack integration")
		return false, errCode, "failed to check slack integration"
	}
	if alert.Slack && !isSlackIntegrated {
		return false, http.StatusBadRequest, "Slack integration is not enabled for this project"
	}
	isTeamsIntegrated, errCode := store.IsTeamsIntegratedForProject(projectID, agentID)
	if errCode != http.StatusOK {
		log.WithFields(log.Fields{"agentUUID": agentID, "event_trigger_alert": alert}).Error("failed to check teams integration")
		return false, errCode, "failed to check teams integration"
	}
	if alert.Teams && !isTeamsIntegrated {
		return false, http.StatusBadRequest, "Teams integration is not enabled for this project"
	}
	if alert.Webhook && alert.WebhookURL == "" {
		return false, http.StatusBadRequest, "webhook url must not be empty"
	}
	return true, http.StatusOK, ""
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

func (store *MemSQL) GetEventTriggerAlertsByEvent(projectId int64, id string) ([]model.EventTriggerAlert, model.EventName, int) {
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
			return nil, model.EventName{}, http.StatusNotFound
		}

		return nil, model.EventName{}, http.StatusInternalServerError
	}

	if err := db.Where("project_id = ? AND is_deleted = 0", projectId).
		Where("JSON_EXTRACT_STRING(event_trigger_alert, 'event') LIKE ?", eventName.Name).
		Not("internal_status = ?", model.Disabled).
		Find(&eventAlerts).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "event": eventName.Name}).WithError(err).Error(
			"filtering eventName failed on GetFilterEventNamesByEvent")
		if gorm.IsRecordNotFoundError(err) {
			return nil, eventName, http.StatusNotFound
		}

		return nil, eventName, http.StatusInternalServerError
	}

	if len(eventAlerts) == 0 {
		return nil, eventName, http.StatusNotFound
	}

	return eventAlerts, eventName, http.StatusFound
}

func (store *MemSQL) MatchEventTriggerAlertWithTrackPayload(projectId int64, eventNameId string, eventProps, userProps *postgres.Jsonb, UpdatedEventProps *postgres.Jsonb, isUpdate bool) (*[]model.EventTriggerAlert, *model.EventName, int) {
	logFields := log.Fields{
		"project_id":       projectId,
		"event_name":       eventNameId,
		"event_properties": eventProps,
		"user_properties":  userProps,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	alerts, eventName, errCode := store.GetEventTriggerAlertsByEvent(projectId, eventNameId)
	if errCode != http.StatusFound || alerts == nil {
		//log.WithFields(logFields).Error("GetEventTriggerAlertsByEvent failure inside Match function.")
		return nil, nil, errCode
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
			return nil, nil, http.StatusInternalServerError
		}
		messageProperties := make([]model.QueryGroupByProperty, 0)
		if config.MessageProperty != nil {
			err := U.DecodePostgresJsonbToStructType(config.MessageProperty, &messageProperties)
			if err != nil {
				log.WithError(err).Error("Jsonb decoding to struct failure")
				return nil, nil, http.StatusInternalServerError
			}
		}
		if !isUpdate {
			isUpdateOnlyPropertyInMessageBody := false
			for _, msgProp := range messageProperties {
				if eventName.Name == "$session" && U.SESSION_PROPERTIES_SET_IN_UPDATE[msgProp.Property] == true {
					isUpdateOnlyPropertyInMessageBody = true
				}
			}
			if isUpdateOnlyPropertyInMessageBody {
				continue
			}
		}
		if isUpdate {
			if len(*updatedEventProps) == 0 {
				continue
			} else {
				isPropertyInFilterUpdated := false
				isUpdateOnlyPropertyInMessageBody := false
				for _, msgProp := range messageProperties {
					if eventName.Name == "$session" && U.SESSION_PROPERTIES_SET_IN_UPDATE[msgProp.Property] == true {
						isUpdateOnlyPropertyInMessageBody = true
					}
				}
				for _, fil := range config.Filter {
					_, exists := (*updatedEventProps)[fil.Property]
					if fil.Entity == "event" && exists {
						isPropertyInFilterUpdated = true
					}
				}
				if !isPropertyInFilterUpdated && !isUpdateOnlyPropertyInMessageBody {
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
		return nil, nil, http.StatusNotFound
	}
	return &matchedAlerts, &eventName, http.StatusFound
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

func (store *MemSQL) GetMessageAndBreakdownPropertiesMap(event *model.Event, alert *model.EventTriggerAlertConfig, eventName *model.EventName) (U.PropertiesMap, map[string]interface{}, error) {
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

	displayNamesEP := store.getDisplayNamesForEP(event.ProjectId, eventName.Name)
	//log.Info(fmt.Printf("%+v\n", displayNamesEP))

	displayNamesUP := store.getDisplayNamesForUP(event.ProjectId)
	//log.Info(fmt.Printf("%+v\n", displayNamesUP))

	msgPropMap := make(U.PropertiesMap, 0)
	for idx, messageProperty := range messageProperties {
		p := messageProperty.Property
		if messageProperty.Entity == "user" {

			displayName, exists := displayNamesUP[p]
			if !exists {
				displayName = U.CreateVirtualDisplayName(p)
			}
			propVal, exi := (*userPropMap)[p]
			msgPropMap[fmt.Sprintf("%d", idx)] = model.MessagePropMapStruct{
				DisplayName: displayName,
				PropValue:   getDisplayLikePropValue(messageProperty.Type, exi, propVal),
			}

		} else if messageProperty.Entity == "event" {
			displayName, exists := displayNamesEP[p]
			if !exists {
				displayName = U.CreateVirtualDisplayName(p)
			}
			propVal, exi := (*eventPropMap)[p]
			msgPropMap[fmt.Sprintf("%d", idx)] = model.MessagePropMapStruct{
				DisplayName: displayName,
				PropValue:   getDisplayLikePropValue(messageProperty.Type, exi, propVal),
			}
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

		if breakdownProperty.Entity == "user" && uexists {
			value = uval
		} else if breakdownProperty.Entity == "event" && eexists {
			value = eval
		} else {
			log.Warn("can not find the breakdown property in user and event prop sets")
		}
		breakdownPropMap[prop] = value
	}

	projectID := event.ProjectId
	if C.AllowSyncReferenceFields(projectID) {
		breakdownPropMap, err = store.transformBreakdownPropertiesToPropertyLabels(projectID, breakdownPropMap)
		if err != nil {
			log.WithError(err).Error("Failed to get property labels on GetMessageAndBreakdownPropertiesMap")
			return msgPropMap, breakdownPropMap, err
		}
	}

	return msgPropMap, breakdownPropMap, nil
}

func (store *MemSQL) transformBreakdownPropertiesToPropertyLabels(projectID int64, breakdownPropertiesMap map[string]interface{}) (map[string]interface{}, error) {
	if projectID == 0 {
		return breakdownPropertiesMap, errors.New("Invalid parameters.")
	}

	newBreakdownPropertiesMap := make(map[string]interface{}, 0)
	for propertyKey, valueInt := range breakdownPropertiesMap {
		if !U.IsAllowedCRMPropertyPrefix(propertyKey) {
			continue
		}

		source := strings.Split(propertyKey, "_")[0]
		source = strings.TrimPrefix(source, "$")

		value := U.GetPropertyValueAsString(valueInt)

		displayLabel, errCode, err := store.GetDisplayNameLabel(projectID, source, propertyKey, value)
		if errCode != http.StatusFound && errCode != http.StatusNotFound {
			log.WithFields(log.Fields{"project_id": projectID, "source": source,
				"property_key": propertyKey, "value": value}).WithError(err).Error("Failed to get display name label.")
			return breakdownPropertiesMap, err
		}

		if errCode == http.StatusNotFound {
			newBreakdownPropertiesMap[propertyKey] = valueInt
			continue
		}

		newBreakdownPropertiesMap[propertyKey] = displayLabel.Label
	}

	return newBreakdownPropertiesMap, nil
}

func getCacheKeyAndSortedSetTupleAndCheckCoolDownTimeCondition(projectID int64, dontRepeatAlerts bool,
	coolDownTime, unixtime int64, alertID string, breakdownProps *map[string]interface{}) (bool,
	*cacheRedis.Key, cacheRedis.SortedSetKeyValueTuple, error) {

	key, err := model.GetEventTriggerAlertCacheKey(projectID, unixtime, alertID)
	if err != nil {
		log.WithError(err).Error("error while getting cache Key")
		return false, nil, cacheRedis.SortedSetKeyValueTuple{}, err
	}

	check := true
	if dontRepeatAlerts {
		check, err = isCoolDownTimeExhausted(key, coolDownTime, unixtime, breakdownProps)
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

func isCoolDownTimeExhausted(key *cacheRedis.Key, coolDownTime, unixtime int64, breakdownProps *map[string]interface{}) (bool, error) {

	// coolDownKeyCounter structure = ETA:CoolDown:pid:<project_id>:<alert_id>:<prop>:<value>:....
	// remove the unixtime from the alert cache key
	// sort and stringify the breakdownProps
	// create the coolDown cacheKey from the alert key by adding breakdownProps
	// INCR the coolDown cacheKey
	// if the INCR returns value 1 then set the expiry as coolDownTime and return true
	// else return false for isCoolDownTimeExhausted

	suffix := strings.TrimRight(key.Suffix, fmt.Sprintf("%d", unixtime))
	props := make([]string, 0, len(*breakdownProps))
	for p := range *breakdownProps {
		props = append(props, p)
	}
	sort.Strings(props)
	for _, prop := range props {
		suffix = fmt.Sprintf("%s:%s:%v", suffix, prop, (*breakdownProps)[prop])
	}

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

func (store *MemSQL) CacheEventTriggerAlert(alert *model.EventTriggerAlert, event *model.Event, eventName *model.EventName) bool {

	// Adding alert to cache
	// Check for cooldown against the breakdown properties
	// Get sorted set keys from where all the alert keys for a particular projectID are retrieved
	// Get the alert key as well
	// If coolDown is not exhausted then return
	// INCR the counter key
	// If the counterKey is present, continue
	// Else set the counter key with one day of expiry
	// If the counter key has count less than daily limit, then
	// Add the alert key to the sorted set and cache

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
	timestamp := tt.UnixNano()
	date := tt.UTC().Format(U.DATETIME_FORMAT_YYYYMMDD)

	messageProps, breakdownProps, err := store.GetMessageAndBreakdownPropertiesMap(event, &eta, eventName)
	if err != nil {
		log.WithError(err).Error("key and sortedTuple fetching error")
		return false
	}

	isCoolDownTimeExpired, key, sortedSetTuple, err := getCacheKeyAndSortedSetTupleAndCheckCoolDownTimeCondition(
		event.ProjectId, eta.DontRepeatAlerts, eta.CoolDownTime, timestamp, alert.ID, &breakdownProps)
	if err != nil {
		log.WithError(err).Error("key and sortedTuple fetching error")
		return false
	}

	if !isCoolDownTimeExpired {
		log.WithFields(logFields).Info("Alert sending cancelled due to cool down timer")
		return true
	}

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

	if eta.SetAlertLimit && count > eta.AlertLimit {
		log.WithFields(logFields).
			Info("Alert was not sent for current EventTriggerAlert as daily AlertLimit has been reached.")

		return true
	}
	_, err = cacheRedis.ZincrPersistentBatch(true, sortedSetTuple)
	if err != nil {
		log.WithError(err).Error("error while getting zincr")
		return false
	}

	successCode, err := store.AddAlertToCache(&eta, &messageProps, key)
	if err != nil || successCode != http.StatusCreated {
		log.WithFields(logFields).Error("Failed to send alert.")
		return false
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

func (store *MemSQL) UpdateInternalStatusAndGetAlertIDs(projectID int64) ([]string, int, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	alertIDs := make([]string, 0)
	alerts := make([]model.EventTriggerAlert, 0)

	err := db.Where("project_id = ? AND is_deleted = ?", projectID, false).
		Find(&alerts).Error
	if err != nil {
		log.WithFields(logFields).WithError(err).
			Error("Failed to fetch rows from event_trigger_alerts table")
		return alertIDs, http.StatusInternalServerError, err
	}

	for _, alert := range alerts {
		var lastFail model.LastFailDetails
		if alert.LastFailDetails != nil {
			err := U.DecodePostgresJsonbToStructType(alert.LastFailDetails, &lastFail)
			if err != nil {
				log.WithFields(logFields).WithError(err).Error("Error in decoding jsonb to struct type")
				return alertIDs, http.StatusInternalServerError, err
			}
		}

		var lastSentTime time.Time
		if alert.LastAlertAt.IsZero() {
			lastSentTime = alert.CreatedAt
		} else {
			lastSentTime = alert.LastAlertAt
		}
		tt := lastFail.FailTime.Sub(lastSentTime)

		if tt.Hours() >= PoisonTime && alert.InternalStatus != model.Paused {
			updateInternalStatus := map[string]interface{}{
				"internal_status": model.Paused,
			}
			errCode, err := store.UpdateEventTriggerAlertField(projectID, alert.ID, updateInternalStatus)
			if errCode != http.StatusAccepted || err != nil {
				log.WithFields(log.Fields{"project_id": projectID, "alert_id": alert.ID}).WithError(err).
					Error("Failed to update event_trigger_alert row")
			}
			alert.InternalStatus = model.Paused
		}
		if tt.Hours() >= DisableTime && alert.InternalStatus != model.Disabled {
			updateInternalStatus := map[string]interface{}{
				"internal_status": model.Disabled,
			}
			errCode, err := store.UpdateEventTriggerAlertField(projectID, alert.ID, updateInternalStatus)
			if errCode != http.StatusAccepted || err != nil {
				log.WithFields(log.Fields{"project_id": projectID, "alert_id": alert.ID}).WithError(err).
					Error("Failed to update event_trigger_alert row")
			}
			alert.InternalStatus = model.Disabled
		}
		if alert.InternalStatus == model.Paused {
			alertIDs = append(alertIDs, alert.ID)
		}
	}

	return alertIDs, http.StatusOK, nil
}
