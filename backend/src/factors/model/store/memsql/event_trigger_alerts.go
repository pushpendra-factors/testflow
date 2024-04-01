package memsql

import (
	"encoding/json"
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

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	ETA                                  = "event_based_alert"
	ListLimit                            = 1000
	AlertCreationLimit                   = 200
	SortedSetCacheKey                    = "ETA:pid"
	CoolDownPrefix                       = "ETA:CoolDown"
	oneDayInSeconds                      = 24 * 60 * 60
	PoisonTime                           = 24             // Hours after which the alert will be paused internally
	DisableTime                          = 2 * PoisonTime // Hours after which the alert will not be processed from sdk
	DATETIME_FORMAT_YYYYMMDD_HYPHEN_HHMM = "2006-01-02 15:04"
)

func (store *MemSQL) GetAllEventTriggerAlertsByProject(projectID int64) ([]model.AlertInfo, int) {
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

func (store *MemSQL) GetInternalStatusForEventTriggerAlert(projectID int64, id string) (string, int, error) {
	logFields := log.Fields{
		"id": id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var alert model.EventTriggerAlert
	err := db.Where("project_id = ?", projectID).Where("id = ?", id).
		Where("is_deleted = ?", false).Find(&alert).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", http.StatusNotFound, err
		}
		log.WithError(err).Error("Failed to fetch rows from event_trigger_alerts table for project")
		return "", http.StatusInternalServerError, err
	}

	return alert.InternalStatus, http.StatusFound, nil
}

func (store *MemSQL) convertEventTriggerAlertToEventTriggerAlertInfo(list []model.EventTriggerAlert) []model.AlertInfo {

	res := make([]model.AlertInfo, 0)

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
				deliveryOption += "& Webhook "
			}
		}
		if alert.Teams {
			if deliveryOption == "" {
				deliveryOption += "Teams"
			} else {
				deliveryOption += "& Teams"
			}
		}

		internalStatus := ""
		if obj.InternalStatus == model.Active || obj.InternalStatus == model.Paused {
			internalStatus = model.Active
		} else if obj.InternalStatus == model.Disabled {
			internalStatus = model.Paused
		}

		// not showing any errors in UI if there is a subsequent success
		var lastFailJson *postgres.Jsonb
		var lastFail model.LastFailDetails
		if obj.LastFailDetails != nil {
			err = U.DecodePostgresJsonbToStructType(obj.LastFailDetails, &lastFail)
			if err != nil {
				log.WithError(err).Error("error deserializing last fail details query")
			}
		}

		if !lastFail.IsPausedAutomatically && obj.LastAlertAt.Compare(lastFail.FailTime) <= 0 {
			lastFailJson = obj.LastFailDetails
		} else {
			lastFailJson = nil
		}

		e := model.AlertInfo{
			ID:              obj.ID,
			Title:           obj.Title,
			DeliveryOptions: deliveryOption,
			LastFailDetails: lastFailJson,
			Status:          internalStatus,
			Alert:           obj.EventTriggerAlert,
			Type:            ETA,
			CreatedAt:       obj.CreatedAt,
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

func (store *MemSQL) CreateEventTriggerAlert(userID, oldID string, projectID int64, alertConfig *model.EventTriggerAlertConfig, slackTokenUser, teamTokenUser string, isPausedAlert bool, paragonMetadata *postgres.Jsonb) (*model.EventTriggerAlert, int, string) {
	logFields := log.Fields{
		"project_id":          projectID,
		"event_trigger_alert": alertConfig,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db

	var alert model.EventTriggerAlert
	transTime := time.Now().UTC()
	id := U.GetUUID()

	isValidAlertBody, errCode, errMsg := store.isValidEventTriggerAlertBody(projectID, userID, alertConfig)
	if !isValidAlertBody {
		logCtx.WithError(fmt.Errorf(errMsg)).Error("invalid alert body provided")
		return nil, errCode, errMsg
	}

	if alertCreationLimitExceeded(projectID) {
		logCtx.Error("alerts limit reached")
		return nil, http.StatusConflict, "alerts limit reached"
	}

	if isDuplicateAlertTitle(projectID, alertConfig.Title, oldID) {
		logCtx.Error("An alert with the same name already exists. Please choose a different name before saving.")
		return nil, http.StatusConflict, "alert already exist"
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
				err = cacheRedis.ZAddPersistent(cacheKeyList, strings.TrimSpace(value), 0)
				if err != nil {
					log.WithFields(logFields).WithError(err).Error("failed to add new values to sorted set")
					return nil, http.StatusInternalServerError, "failed to add new values to sorted set"
				}
			}
		}
	}

	trigger, err := U.EncodeStructTypeToPostgresJsonb(*alertConfig)
	if err != nil {
		logCtx.WithError(err).Error("TriggerAlert conversion to Jsonb failed")
		return nil, http.StatusInternalServerError, "TiggerAlert conversion to Jsonb failed"
	}

	internalStatus := model.Active
	if isPausedAlert {
		internalStatus = model.Disabled
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
		InternalStatus:           internalStatus,
		ParagonMetadata:          paragonMetadata,
	}

	if err := db.Create(&alert).Error; err != nil {
		logCtx.WithError(err).Error("Create Failed in db")
		return nil, http.StatusInternalServerError, "Create Failed in db"
	}

	return &alert, http.StatusCreated, ""
}

func isEmptyPostgresJsonb(json *postgres.Jsonb) bool {
	j := string(json.RawMessage)
	return j == "" || j == "null" || j == "[]"
}

func (store *MemSQL) isValidEventTriggerAlertBody(projectID int64, agentID string, alert *model.EventTriggerAlertConfig) (bool, int, string) {

	logCtx := log.WithFields(log.Fields{
		"project_id":   projectID,
		"agent_uuid":   agentID,
		"alert_config": *alert,
	})

	if alert.Title == "" {
		errMsg := "Please provide a name for the alert before saving."
		logCtx.WithError(fmt.Errorf(errMsg)).Error("alert body validation failure")
		return false, http.StatusBadRequest, errMsg
	}
	if alert.Event == "" {
		errMsg := "Please select an event to trigger the alert."
		logCtx.WithError(fmt.Errorf(errMsg)).Error("alert body validation failure")
		return false, http.StatusBadRequest, errMsg
	}
	if alert.DontRepeatAlerts && (alert.BreakdownProperties == nil || isEmptyPostgresJsonb(alert.BreakdownProperties)) {
		errMsg := "Please specify a property to limit alerts under advanced settings."
		logCtx.WithError(fmt.Errorf(errMsg)).Error("alert body validation failure")
		return false, http.StatusBadRequest, errMsg
	}
	if !alert.Slack && !alert.Webhook && !alert.Teams {
		errMsg := "Please choose at least 1 destination to receive the alerts in before saving."
		logCtx.WithError(fmt.Errorf(errMsg)).Error("alert body validation failure")
		return false, http.StatusBadRequest, errMsg
	}
	if alert.Slack && (alert.SlackChannels == nil || U.IsEmptyPostgresJsonb(alert.SlackChannels)) {
		errMsg := "Please select the Slack channel where you want to receive the alert."
		logCtx.WithError(fmt.Errorf(errMsg)).Error("alert body validation failure")
		return false, http.StatusBadRequest, errMsg
	}
	isSlackIntegrated, errCode := store.IsSlackIntegratedForProject(projectID, agentID)
	if errCode != http.StatusOK {
		log.WithFields(log.Fields{"agentUUID": agentID, "event_trigger_alert": alert}).Error("failed to check slack integration")
		return false, errCode, "failed to check slack integration"
	}
	if alert.Slack && !isSlackIntegrated {
		errMsg := "slack integration is not enabled for this project"
		logCtx.WithError(fmt.Errorf(errMsg)).Error("alert body validation failure")
		return false, http.StatusBadRequest, errMsg
	}
	isTeamsIntegrated, errCode := store.IsTeamsIntegratedForProject(projectID, agentID)
	if errCode != http.StatusOK {
		log.WithFields(log.Fields{"agentUUID": agentID, "event_trigger_alert": alert}).Error("failed to check teams integration")
		return false, errCode, "failed to check teams integration"
	}
	if alert.Teams && !isTeamsIntegrated {
		errMsg := "teams integration is not enabled for this project"
		logCtx.WithError(fmt.Errorf(errMsg)).Error("alert body validation failure")
		return false, http.StatusBadRequest, errMsg
	}
	if alert.Webhook && alert.WebhookURL == "" {
		errMsg := "Please enter the Webhook URL to send data through webhook."
		logCtx.WithError(fmt.Errorf(errMsg)).Error("alert body validation failure")
		return false, http.StatusBadRequest, errMsg
	}
	if duplicateMessagePropertiesPresent(alert.MessageProperty) {
		errMsg := "Duplicate properties are selected in the alert configuration. Please remove them before saving."
		logCtx.WithError(fmt.Errorf(errMsg)).Error("alert body validation failure")
		return false, http.StatusBadRequest, errMsg
	}

	return true, http.StatusOK, ""
}

func duplicateMessagePropertiesPresent(mp *postgres.Jsonb) bool {

	//Note to anyone reading this code. Always check for empty Jsonb and nil pointers
	if mp == nil || isEmptyPostgresJsonb(mp) {
		return false
	}

	props := make([]model.QueryGroupByProperty, 0)
	err := U.DecodePostgresJsonbToStructType(mp, &props)
	if err != nil {
		return true
	}

	for i := 0; i < len(props)-1; i++ {
		for j := i + 1; j < len(props); j++ {
			// timestamp property can be selected for multiple granularities like day, hour, and week
			if strings.EqualFold(props[i].Property, props[j].Property) && (props[i].Entity == props[j].Entity) && (props[i].Property != "$timestamp") {
				return true
			}
		}
	}
	return false
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

func (store *MemSQL) GetEventTriggerAlertsForTheCurrentEvent(projectId int64, id string) ([]model.EventTriggerAlert, model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"prefix":     id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var eventAlerts []model.EventTriggerAlert
	var eventName *model.EventName
	var errCode int
	var err error

	if id != "" {
		eventName, errCode, err = store.GetEventNameByID(projectId, id)
		if errCode != http.StatusFound || err != nil {
			log.WithFields(logFields).WithError(err).Error("event_name not found")
			return nil, model.EventName{}, http.StatusInternalServerError
		}
	} else {
		eventName = &model.EventName{
			Name: U.EVENT_NAME_PAGE_VIEW,
		}
	}

	eventAlerts, errCode, err = store.GetEventTriggerAlertsByEventName(projectId, eventName.Name)
	if (errCode != http.StatusFound && errCode != http.StatusNotFound) || err != nil {
		log.WithFields(logFields).WithError(err).Error("event_name not found")
		return nil, model.EventName{}, errCode
	}

	if len(eventAlerts) == 0 {
		return nil, *eventName, http.StatusNotFound
	}

	return eventAlerts, *eventName, http.StatusFound
}

func (store *MemSQL) GetEventTriggerAlertsByEventName(projectID int64, eventName string) ([]model.EventTriggerAlert, int, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	eventAlerts := make([]model.EventTriggerAlert, 0)
	db := C.GetServices().Db

	if err := db.Where("project_id = ? AND is_deleted = 0", projectID).
		Where("JSON_EXTRACT_STRING(event_trigger_alert, 'event') LIKE ?", eventName).
		Not("internal_status = ?", model.Disabled).
		Find(&eventAlerts).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, err
		}
		log.WithFields(logFields).WithError(err).
			Error("can not get event trigger alerts by event names")
		return nil, http.StatusInternalServerError, err
	}

	return eventAlerts, http.StatusFound, nil
}

func (store *MemSQL) MatchEventTriggerAlertWithTrackPayload(projectId int64, eventNameId, userID string, eventProps, userProps *postgres.Jsonb, UpdatedEventProps *postgres.Jsonb, isUpdate bool) (*[]model.EventTriggerAlert, *model.EventName, *map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id":       projectId,
		"event_name":       eventNameId,
		"event_properties": eventProps,
		"user_properties":  userProps,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	alerts, eventName, errCode := store.GetEventTriggerAlertsForTheCurrentEvent(projectId, eventNameId)
	if errCode != http.StatusFound || alerts == nil {
		//log.WithFields(logFields).Error("GetEventTriggerAlertsByEvent failure inside Match function.")
		return nil, nil, nil, errCode
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
			logCtx.WithError(err).Error("Jsonb decoding to struct failure")
			return nil, nil, nil, http.StatusInternalServerError
		}
		messageProperties := make([]model.QueryGroupByProperty, 0)
		if config.MessageProperty != nil {
			err := U.DecodePostgresJsonbToStructType(config.MessageProperty, &messageProperties)
			if err != nil {
				logCtx.WithError(err).Error("Jsonb decoding to struct failure")
				return nil, nil, nil, http.StatusInternalServerError
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

		var groupProps *map[string]interface{}
		isGroupPropertyRequired := false
		for _, fil := range config.Filter {
			if model.AllowedGroupNames[fil.GroupName] || fil.GroupName == model.GROUP_NAME_DOMAINS {
				isGroupPropertyRequired = true
				break
			}
		}

		if config.EventLevel == model.EventLevelAccount && isGroupPropertyRequired {
			domainsGroupUserID, domainsGroupID, err := store.getDomainsGroupUserIDForUser(projectId, userID)
			if err != nil {
				logCtx.Warn("no domains group for the alert")
				continue
			}
			groupProps = store.GetAllGroupPropertiesForGivenDomainGroupUserID(projectId, domainsGroupID, domainsGroupUserID)
			if groupProps != nil {
				updateUserPropMapWithGroupProperties(userPropMap, groupProps, log.WithFields(logFields))
			}
		}

		//IN_PROPERTIES_DEFAULT_QUERY_MAP selected properties check
		for i, filter := range config.Filter {
			_logicalOp := filter.LogicalOp
			if q, exists := model.IN_PROPERTIES_DEFAULT_QUERY_MAP[filter.Property]; exists {
				if filter.Value == "true" {
					config.Filter[i] = q
				} else if filter.Value == "false" || filter.Value == "$none" {
					config.Filter[i] = q
					config.Filter[i].Operator = model.EqualsOpStr
				}
			}

			config.Filter[i].LogicalOp = _logicalOp
			if config.Filter[i].LogicalOp == "" {
				config.Filter[i].LogicalOp = "AND"
			}
		}

		criteria := E.MapFilterProperties(config.Filter)
		if E.EventMatchesFilterCriterionList(projectId, *userPropMap, *eventPropMap, criteria) {
			matchedAlerts = append(matchedAlerts, alert)
		}
	}
	if len(matchedAlerts) == 0 {
		log.WithFields(logFields).Info("Match function did not find anything in event_trigger_alerts")
		return nil, nil, nil, http.StatusNotFound
	}
	return &matchedAlerts, &eventName, userPropMap, http.StatusFound
}

func updateUserPropMapWithGroupProperties(userPropMap, groupProps *map[string]interface{}, logCtx *log.Entry) {
	if userPropMap == nil || groupProps == nil {
		logCtx.Warn("empty Prop map found")
		return
	}

	for key, value := range *groupProps {
		if value != nil {
			(*userPropMap)[key] = value
		}
	}
}

func (store *MemSQL) getDomainsGroupUserIDForUser(projectID int64, userID string) (string, int, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	user, errCode := store.GetUser(projectID, userID)
	if errCode != http.StatusFound {
		logCtx.Error("user not found")
		return "", 0, fmt.Errorf("user not found")
	}

	domainsGroup, errCode := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound || domainsGroup == nil {
		logCtx.Error("no domains group found for project")
		return "", 0, fmt.Errorf("")
	}

	domainsGroupUserID, err := model.GetUserGroupUserID(user, domainsGroup.ID)
	if err != nil || domainsGroupUserID == "" {
		logCtx.WithField("domain_id", domainsGroup.ID).Error("no domains group for the given id found for project")
		return "", domainsGroup.ID, fmt.Errorf("no domains group found for project")
	}

	return domainsGroupUserID, domainsGroup.ID, nil
}

func (store *MemSQL) GetAllGroupPropertiesForGivenDomainGroupUserID(projectID int64, domainsGroupID int, domainsGroupUserID string) *map[string]interface{} {
	logFields := log.Fields{
		"project_id": projectID,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	groupUsers, errCode := store.GetAllGroupUsersByDomainsGroupUserID(projectID, domainsGroupID, domainsGroupUserID)
	if errCode != http.StatusFound || len(groupUsers) == 0 {
		logCtx.Error("no group user found")
		return nil
	}

	groupPropsMap := make(map[string]interface{})

	// Get user properties from all groups and add them to groupPropsMap
	for _, gpUser := range groupUsers {
		getUserPropertiesFromGroupUser(&gpUser, &groupPropsMap, logCtx)
	}

	//Get $domains user to get all account properties
	domainsUser, errCode := store.GetUser(projectID, domainsGroupUserID)
	if errCode != http.StatusFound {
		logCtx.Error("failed to get domains user")
		return &groupPropsMap
	}
	// Get user properties from $domains user and add them to groupPropsMap
	getUserPropertiesFromGroupUser(domainsUser, &groupPropsMap, logCtx)

	return &groupPropsMap
}

func getUserPropertiesFromGroupUser(user *model.User, groupPropMap *map[string]interface{}, logCtx *log.Entry) {

	if isEmptyPostgresJsonb(&user.Properties) {
		logCtx.WithField("group_user", user).Info("no properties for user")
		return
	}

	propMap, err := U.DecodePostgresJsonb(&user.Properties)
	if err != nil {
		logCtx.WithError(err).Error("unable to decode postgres jsonb for properties")
		return
	}

	// Update groupPropMap with properties from propMap
	updateUserPropMapWithGroupProperties(groupPropMap, propMap, logCtx)

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

func (store *MemSQL) AddAlertToCache(alert *model.EventTriggerAlertConfig, msgProps *U.PropertiesMap, fieldTagsMap map[string]string, key *cacheRedis.Key) (int, error) {
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
		Message:   message,
		FieldTags: fieldTagsMap,
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

func (store *MemSQL) getDisplayLikePropValue(projectID int64, typ string, exi bool, prop string, value interface{}, granularity string) (string, interface{}) {

	var res interface{}
	if exi {
		if typ == "datetime" {
			val, err := U.GetPropertyValueAsInt64(value)
			if err != nil {
				log.WithFields(log.Fields{
					"property_type":  typ,
					"property_value": val,
				}).Error("datetime property value could not parsed as int")
				return prop, ""
			}

			displayName, displayPropVal := store.getDisplayLikeNameAndPropValForDatetimeType(projectID, prop, granularity, val)
			if strings.EqualFold(prop, "Timestamp") {
				prop = displayName
			}
			res = displayPropVal

		} else {
			res = U.GetPropertyValueAsString(value)
		}
	}

	return prop, res
}

func satisfiesInternalProperty(filterVal string, propVal interface{}, op string) bool {
	if op == model.NotEqualOpStr && filterVal == "$none" {
		propVal = U.GetPropertyValueAsString(propVal)
		if propVal != filterVal && propVal != "" {
			return true
		}
	} else if op == model.EqualsOpStr && filterVal == "$none" {
		propVal = U.GetPropertyValueAsString(propVal)
		if propVal == filterVal || propVal == "" {
			return true
		}
	} else if op == model.GreaterThanOpStr && filterVal == "0" {
		intfilterVal, err := U.GetPropertyValueAsInt64(filterVal)
		if err != nil {
			log.Error("can get property value as int64")
			return false
		}
		intPropVal, err := U.GetPropertyValueAsInt64(propVal)
		if err != nil {
			log.Error("can get property value as int64")
			return false
		}
		return intPropVal > intfilterVal
	}

	return false
}

func (store *MemSQL) GetMessageAndBreakdownPropertiesAndFieldsTagMap(projectID int64, event *model.Event, alert *model.EventTriggerAlertConfig, eventName *model.EventName, updatedUserProps *map[string]interface{}) (U.PropertiesMap, map[string]interface{}, map[string]string, error) {
	logFields := log.Fields{
		"project_id":          projectID,
		"event_trigger_alert": *alert,
		"event":               *event,
		"event_name":          *eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	messageProperties := make([]model.QueryGroupByProperty, 0)
	if alert.MessageProperty != nil {
		err := U.DecodePostgresJsonbToStructType(alert.MessageProperty, &messageProperties)
		if err != nil {
			log.WithError(err).Error("Jsonb decoding to struct failure")
			return nil, nil, nil, err
		}
	}

	var eventPropMap *map[string]interface{}
	var err error
	if &event.Properties != nil && len(event.Properties.RawMessage) != 0 {
		eventPropMap, err = U.DecodePostgresJsonb(&event.Properties)
		if err != nil {
			log.WithError(err).Error("Jsonb decoding to propMap failure")
			return nil, nil, nil, err
		}
	}

	var displayNamesEP, displayNamesUP map[string]string
	if eventName.Name != "" {
		displayNamesEP = store.getDisplayNamesForEP(projectID, eventName.Name)
		//log.Info(fmt.Printf("%+v\n", displayNamesEP))
	}

	displayNamesUP = store.getDisplayNamesForUP(projectID)
	//log.Info(fmt.Printf("%+v\n", displayNamesUP))

	msgPropMap := store.getMessageProperties(projectID, displayNamesEP, displayNamesUP, messageProperties, eventPropMap, updatedUserProps)

	if alert.EventLevel == model.EventLevelAccount && event.UserId != "" {
		groupDomainUserID, _, _ := store.getDomainsGroupUserIDForUser(projectID, event.UserId)
		msgPropMap[model.ETA_DOMAIN_GROUP_USER_ID] = groupDomainUserID

		// for hubspot company url
		if hsUrl, exists := (*updatedUserProps)[U.ENRICHED_HUBSPOT_COMPANY_OBJECT_URL]; exists {
			msgPropMap[model.ETA_ENRICHED_HUBSPOT_COMPANY_OBJECT_URL] = hsUrl
		}

		// for salesforce company url
		if sfUrl, exists := (*updatedUserProps)[U.ENRICHED_SALESFORCE_ACCOUNT_OBJECT_URL]; exists {
			msgPropMap[model.ETA_ENRICHED_SALESFORCE_ACCOUNT_OBJECT_URL] = sfUrl
		}
	}

	breakdownPropMap := make(map[string]interface{}, 0)
	var breakdownProperties []model.QueryGroupByProperty
	if alert.BreakdownProperties != nil {
		err = U.DecodePostgresJsonbToStructType(alert.BreakdownProperties, &breakdownProperties)
		if err != nil {
			log.WithError(err).Error("breakdownProperty Jsonb decoding to queryGroupByProperty failure")
			return nil, nil, nil, err
		}
	}

	for _, breakdownProperty := range breakdownProperties {
		prop := breakdownProperty.Property
		var value interface{}
		uval, uexists := (*updatedUserProps)[prop]

		var eval interface{}
		var eexists bool
		if eventPropMap != nil {
			eval, eexists = (*eventPropMap)[prop]
		}

		if breakdownProperty.Entity == "user" && uexists {
			value = uval
		} else if breakdownProperty.Entity == "event" && eexists {
			value = eval
		} else {
			log.Warn("can not find the breakdown property in user and event prop sets")
		}
		breakdownPropMap[prop] = value
	}

	fieldTagsMap := make(map[string]string)
	if alert.EventLevel == model.EventLevelAccount && alert.SlackFieldsTag != nil {
		for _, tag := range alert.SlackFieldsTag {
			if ownerField := model.ValidAlertTagsForHubspotOwners[tag]; ownerField != "" {
				ownerId := U.GetPropertyValueAsString((*updatedUserProps)[ownerField])
				if ownerId != "" {
					log.Warn("Unable to find owner id for the field tag")
				}
				fieldTagsMap[tag] = ownerId
			}
		}
	}

	return msgPropMap, breakdownPropMap, fieldTagsMap, nil
}

func (store *MemSQL) getMessageProperties(projectID int64, displayNamesEP, displayNamesUP map[string]string,
	messageProperties []model.QueryGroupByProperty, eventPropMap, updatedUserProps *map[string]interface{}) U.PropertiesMap {

	msgPropMap := make(U.PropertiesMap, 0)
	for idx, messageProperty := range messageProperties {
		p := messageProperty.Property
		var propVal interface{}
		var exi bool
		if messageProperty.Entity == "user" || messageProperty.Entity == model.PropertyEntityUserGlobal {

			displayName, exists := displayNamesUP[p]
			if !exists {
				displayName = U.CreateVirtualDisplayName(p)
			}

			if updatedUserProps != nil {
				propVal, exi = (*updatedUserProps)[p]
			}

			//Serve values for properties which need to be modified internally
			if filter, exists := model.IN_PROPERTIES_DEFAULT_QUERY_MAP[p]; !exi && exists {
				if satisfiesInternalProperty(filter.Value, (*updatedUserProps)[filter.Property], filter.Operator) {
					propVal = true
					exi = true
				}
			}
			// check and get for display name labels for crm property keys
			if value := U.GetPropertyValueAsString(propVal); U.IsAllowedCRMPropertyPrefix(p) && exi && value != "" {
				propertyLabel, exist := store.getDisplayNameLabelForThisProperty(projectID, p, value)
				if exist {
					propVal = propertyLabel
				}
			}

			displayName, displayPropVal := store.getDisplayLikePropValue(projectID, messageProperty.Type, exi, displayName, propVal, messageProperty.Granularity)
			msgPropMap[fmt.Sprintf("%d", idx)] = model.MessagePropMapStruct{
				DisplayName: displayName,
				PropValue:   displayPropVal,
			}

		} else if messageProperty.Entity == "event" {
			displayName, exists := displayNamesEP[p]
			if !exists {
				displayName = U.CreateVirtualDisplayName(p)
			}

			if eventPropMap != nil {
				propVal, exi = (*eventPropMap)[p]
			}

			displayName, displayPropVal := store.getDisplayLikePropValue(projectID, messageProperty.Type, exi, displayName, propVal, messageProperty.Granularity)

			msgPropMap[fmt.Sprintf("%d", idx)] = model.MessagePropMapStruct{
				DisplayName: displayName,
				PropValue:   displayPropVal,
			}
		} else {
			log.Warn("can not find the message property in user and event prop sets")
		}
	}

	return msgPropMap
}
func (store *MemSQL) getDisplayNameLabelForThisProperty(projectID int64, propertyKey, value string) (string, bool) {
	source := strings.Split(propertyKey, "_")[0]
	source = strings.TrimPrefix(source, "$")

	displayLabel, errCode, err := store.GetDisplayNameLabel(projectID, source, propertyKey, value)
	if (errCode != http.StatusFound && errCode != http.StatusNotFound) || err != nil {
		log.WithFields(log.Fields{"project_id": projectID, "source": source,
			"property_key": propertyKey, "value": value}).WithError(err).Error("Failed to get display name label.")
		return "", false
	}

	if errCode == http.StatusNotFound {
		return value, false
	}

	return displayLabel.Label, true
}

func (store *MemSQL) getDisplayLikeNameAndPropValForDatetimeType(projectID int64, displayName, granularity string, propVal interface{}) (string, interface{}) {

	var displayPropVal interface{}
	timezoneString, errCode := store.GetTimezoneForProject(projectID)
	if errCode != http.StatusFound {
		log.WithField("project_id", projectID).Error("failed to get Timezone from project")
	}
	timeLocation := U.GetTimeLocationFor(timezoneString)

	if granularity == "hour" {
		displayName += " - Hour"
		displayPropVal = time.Unix(propVal.(int64), 0).In(timeLocation).Format(DATETIME_FORMAT_YYYYMMDD_HYPHEN_HHMM)
	} else if granularity == "week" {
		displayName += " - Week"
		_, displayPropVal = time.Unix(propVal.(int64), 0).In(timeLocation).ISOWeek()
	} else if granularity == "month" {
		displayName += " - Month"
		displayPropVal = time.Unix(propVal.(int64), 0).In(timeLocation).Month()
	} else {
		displayName += " - Date"
		displayPropVal = time.Unix(propVal.(int64), 0).In(timeLocation).Format(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN)
	}

	return displayName, displayPropVal
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

func (store *MemSQL) CacheEventTriggerAlert(alert *model.EventTriggerAlert, event *model.Event, eventName *model.EventName, updatedUserProps *map[string]interface{}) bool {

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

	messageProps, breakdownProps, fieldsTagMap, err := store.GetMessageAndBreakdownPropertiesAndFieldsTagMap(alert.ProjectID, event, &eta, eventName, updatedUserProps)
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

	successCode, err := store.AddAlertToCache(&eta, &messageProps, fieldsTagMap, key)
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
			//update the is_paused_automatically fields to true in last_fail_details
			lastFail.IsPausedAutomatically = true
			lastFailJson, err := U.EncodeStructTypeToPostgresJsonb(lastFail)
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectID, "alert_id": alert.ID}).WithError(err).
					Error("unable to decode lastFailDetails json")
			}

			updateInternalStatus := map[string]interface{}{
				"internal_status":   model.Disabled,
				"last_fail_details": lastFailJson,
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

func (store *MemSQL) GetParagonMetadataForEventTriggerAlert(projectID int64, alertID string) (map[string]interface{}, int, error) {
	if projectID == 0 || alertID == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid parameters")
	}

	logFields := log.Fields{
		"project_id": projectID,
		"alert_id":   alertID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	var alert model.EventTriggerAlert
	db := C.GetServices().Db
	err := db.Where("project_id = ?", projectID).Where("id = ?", alertID).Find(&alert).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, err
		}
		logCtx.WithError(err).Error("failed to fetch alert for the given params")
		return nil, http.StatusInternalServerError, err
	}

	if alert.ParagonMetadata == nil {
		return nil, http.StatusNotFound, fmt.Errorf("no metadata available")
	}
	metadata, err := U.DecodePostgresJsonb(alert.ParagonMetadata)
	if err != nil {
		logCtx.WithError(err).Error("failed to decode metadata json")
		return nil, http.StatusInternalServerError, err
	}

	return *metadata, http.StatusFound, nil
}

func (store *MemSQL) FindAndCacheAlertForCurrentSegment(projectID int64, segmentID, domainID, actionPerformed string) (int, error) {
	logFields := log.Fields{
		"project_id":       projectID,
		"segment_id":       segmentID,
		"domain_id":        domainID,
		"action_performed": actionPerformed,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	defer func() {
		if err := recover(); err != nil {
			logCtx.WithFields(log.Fields{
				"err": err,
			}).Error("Panic occured.")
		}
	}()

	//Find if any alert is present for the current segment
	alerts, errCode := store.GetEventTriggerAlertsBySegmentID(projectID, segmentID, actionPerformed)
	if errCode != http.StatusFound {
		if errCode == http.StatusNotFound {
			logCtx.Info("No alert found for the matching conditions.")
			return errCode, fmt.Errorf("no alert set for the current segment")
		}
		logCtx.Error("Failure while fetching alerts.")
		return errCode, fmt.Errorf("failure while fetching alerts")
	}

	timestamp := U.TimeNowZ().UnixNano()
	fail := false

	// get $domains id column number
	domainsGroup, errCode := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound || domainsGroup == nil {
		logCtx.Warn("No domains group found for project.")
		return errCode, fmt.Errorf("no domains group found for project")
	}

	// get all group and account properties
	groupUserProps := store.GetAllGroupPropertiesForGivenDomainGroupUserID(projectID, domainsGroup.ID, domainID)

	for _, alert := range alerts {
		// check for empty jsonb
		if alert.EventTriggerAlert == nil {
			logCtx.WithField("alert_id", alerts).Error("Alert config absent for the provided alert.")
			continue
		}

		var alertConfig model.EventTriggerAlertConfig
		err := U.DecodePostgresJsonbToStructType(alert.EventTriggerAlert, &alertConfig)
		if err != nil {
			logCtx.WithError(err).Error("Failed to deserialize json to struct.")
			continue
		}

		messageProps, breakdownProps, fieldTagsMap, err := store.GetMessageAndBreakdownPropertiesAndFieldsTagMap(projectID, &model.Event{}, &alertConfig, &model.EventName{}, groupUserProps)
		if err != nil {
			logCtx.WithError(err).Error("Failed to deserialize json to struct.")
			continue
		}

		// get cache keys
		isCoolDownTimeExpired, key, sortedSetTuple, err := getCacheKeyAndSortedSetTupleAndCheckCoolDownTimeCondition(
			projectID, alertConfig.DontRepeatAlerts, alertConfig.CoolDownTime, timestamp, alert.ID, &breakdownProps)
		if err != nil {
			logCtx.WithError(err).Error("Failed to fetch key and sorted tuple.")
			continue
		}

		if !isCoolDownTimeExpired {
			logCtx.WithFields(logFields).Info("Alert sending cancelled due to cool down timer.")
			continue
		}

		// set key in sorted set
		_, err = cacheRedis.ZincrPersistentBatch(true, sortedSetTuple)
		if err != nil {
			logCtx.WithError(err).Error("Error while getting zincr")
			continue
		}

		// set cache for the alert
		successCode, err := store.AddAlertToCache(&alertConfig, &messageProps, fieldTagsMap, key)
		if err != nil || successCode != http.StatusCreated {
			fail = true
			logCtx.WithField("errCode", successCode).WithError(err).Error("Failed to cache alert")
			continue
		}
	}

	if fail {
		logCtx.Error("Failed to cache one or more alerts.")
		return http.StatusPartialContent, fmt.Errorf("failed to cache one or more alerts")
	}

	return http.StatusAccepted, nil
}

func (store *MemSQL) GetEventTriggerAlertsBySegmentID(projectID int64, segmentID, actionPerformed string) ([]model.EventTriggerAlert, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"segment_id":       segmentID,
		"action_performed": actionPerformed,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	var alerts []model.EventTriggerAlert
	if err := db.Where("project_id = ? AND is_deleted = 0", projectID).
		Where("JSON_EXTRACT_STRING(event_trigger_alert, 'action_performed') LIKE ?", actionPerformed).
		Where("JSON_EXTRACT_STRING(event_trigger_alert, 'event') LIKE ?", segmentID).
		Not("internal_status = ?", model.Disabled).
		Find(&alerts).Limit(ListLimit).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			logCtx.Info("No alert found for the matching conditions.")
			return nil, http.StatusNotFound
		}
		logCtx.WithError(err).Error("Failed to fetch alert for the given segment conditions.")
		return nil, http.StatusInternalServerError
	}

	if len(alerts) == 0 {
		logCtx.Info("No alert found for the matching conditions.")
		return alerts, http.StatusNotFound
	}

	return alerts, http.StatusFound
}
