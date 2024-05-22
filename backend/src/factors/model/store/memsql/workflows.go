package memsql

import (
	"factors/cache"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	E "factors/event_match"
	"factors/model/model"
	"strings"

	U "factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetAllWorkflowTemplates() ([]model.AlertTemplate, int) {

	db := C.GetServices().Db

	var alertTemplates []model.AlertTemplate
	err := db.Where("is_deleted = ?", false).
		Where("is_workflow = ?", true).
		Order("id").Find(&alertTemplates).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return alertTemplates, http.StatusNotFound
		}
		log.WithError(err).Error("Failed to get workflow templates.")
		return alertTemplates, http.StatusInternalServerError
	}

	return alertTemplates, http.StatusOK
}

func (store *MemSQL) GetWorklfowUrlFromTemplate(id int) (string, int) {

	db := C.GetServices().Db

	var alertTemplate model.AlertTemplate
	err := db.Where("is_deleted = ?", false).
		Where("is_workflow = ?", true).
		Where("id = ?", id).Find(&alertTemplate).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", http.StatusNotFound
		}
		log.WithError(err).Error("Failed to get workflow templates.")
		return "", http.StatusInternalServerError
	}

	templateConstants, err := U.DecodePostgresJsonb(alertTemplate.TemplateConstants)
	if err != nil {
		log.WithError(err).Error("Failed to decode template constants.")
		return "", http.StatusInternalServerError
	}

	if templateConstants == nil {
		log.WithError(err).Error("No template constants identified for the required template.")
		return "", http.StatusInternalServerError
	}

	url, exists := (*templateConstants)["url"]
	if !exists {
		return "", http.StatusInternalServerError
	}

	return U.GetPropertyValueAsString(url), http.StatusOK
}

func (store *MemSQL) GetAllWorklfowsByProject(projectID int64) ([]model.WorkflowDisplayableInfo, int, error) {
	if projectID == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	workflows := make([]model.Workflow, 0)
	wfAlerts := make([]model.WorkflowDisplayableInfo, 0)
	err := db.Where("project_id = ?", projectID).Where("is_deleted = ?", false).
		Order("created_at DESC").Limit(ListLimit).Find(&workflows).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return wfAlerts, http.StatusNotFound, err
		}
		log.WithError(err).Error("Failed to fetch rows of workflows")
		return nil, http.StatusInternalServerError, err
	}

	//Transform into displayable workflow object
	for _, wf := range workflows {
		alert := model.WorkflowDisplayableInfo{
			ID:        wf.ID,
			Title:     wf.Name,
			Status:    wf.InternalStatus,
			AlertBody: wf.AlertBody,
			CreatedAt: wf.CreatedAt,
		}
		wfAlerts = append(wfAlerts, alert)
	}

	return wfAlerts, http.StatusFound, nil
}

func (store *MemSQL) GetWorkflowById(projectID int64, id string) (*model.Workflow, int, error) {
	if projectID == 0 || id == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id":  projectID,
		"workflow_id": id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	var workflow model.Workflow

	err := db.Where("project_id = ?", projectID).
		Where("id = ?", id).
		Where("is_deleted = ?", false).
		Order("created_at DESC").Find(&workflow).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, err
		}
		log.WithError(err).Error("Failed to fetch requested workflows")
		return nil, http.StatusInternalServerError, err
	}

	return &workflow, http.StatusFound, nil
}

func (store *MemSQL) CreateWorkflow(projectID int64, agentID, oldIDIfEdit string, alertBody model.WorkflowAlertBody) (*model.Workflow, int, error) {
	if projectID == 0 || agentID == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id": projectID,
		"agent_id":   agentID,
		"workflow":   alertBody,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	if isValid, errMsg := store.isValidWorkflowAlertBody(projectID, oldIDIfEdit, alertBody); !isValid {
		return nil, http.StatusBadRequest, fmt.Errorf(errMsg)
	}

	var workflow model.Workflow
	transTime := U.TimeNowZ()
	id := U.GetUUID()

	alertJson, err := U.EncodeStructTypeToPostgresJsonb(alertBody)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode workflow body")
		return nil, http.StatusInternalServerError, err
	}

	url, errCode := store.GetWorklfowUrlFromTemplate(alertBody.TemplateID)
	if errCode != http.StatusOK {
		logCtx.WithError(err).Error("Failed to assign workflow url.")
		return nil, http.StatusInternalServerError, fmt.Errorf("no url for template")
	}

	workflow = model.Workflow{
		ID:             id,
		ProjectID:      projectID,
		Name:           alertBody.Title,
		AlertBody:      alertJson,
		CreatedBy:      agentID,
		CreatedAt:      transTime,
		UpdatedAt:      transTime,
		InternalStatus: model.ACTIVE,
		WorkflowUrl:    url,
		IsDeleted:      false,
	}

	if err := db.Create(&workflow).Error; err != nil {
		logCtx.WithError(err).Error("Failed to create workflow")
		return nil, http.StatusInternalServerError, err
	}

	return &workflow, http.StatusFound, nil
}

func (store *MemSQL) UpdateWorkflow(projectID int64, id, agentID string, fieldsToUpdate map[string]interface{}) (int, error) {
	if projectID == 0 || id == "" {
		return http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id":  projectID,
		"workflow_id": id,
		"agent_id":    agentID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	var workflow model.Workflow
	transTime := U.TimeNowZ()
	fieldsToUpdate["updated_at"] = transTime

	if err := db.Model(&workflow).Where("project_id = ? AND is_deleted = 0", projectID).
		Where("id = ?", id).Updates(fieldsToUpdate).Error; err != nil {
		logCtx.WithError(err).Error("Failed to update workflow.")
		return http.StatusInternalServerError, err
	}

	return http.StatusAccepted, nil
}

func (store *MemSQL) DeleteWorkflow(projectID int64, id, agentID string) (int, error) {
	if projectID == 0 || id == "" || agentID == "" {
		return http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id":  projectID,
		"workflow_id": id,
		"agent_id":    agentID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	transTime := U.TimeNowZ()

	err := db.Model(&model.Workflow{}).
		Where("id = ?", id).
		Where("project_id = ?", projectID).
		Updates(map[string]interface{}{"is_deleted": true, "updated_at": transTime}).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to delete workflow.")
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (store *MemSQL) isValidWorkflowAlertBody(projectID int64, id string, body model.WorkflowAlertBody) (bool, string) {
	if body.Title == "" {
		return false, "Please set title for the workflow. Title can not be empty."
	}
	if store.isDuplicateWorkflowTitle(id, body.Title, projectID) {
		return false, "Title already in use. Please provide a different title."
	}
	return true, ""
}

func (store *MemSQL) isDuplicateWorkflowTitle(id, title string, projectID int64) bool {
	if id == "" {
		return false
	}

	logFields := log.Fields{
		"project_id":  projectID,
		"workflow_id": id,
		"title":       title,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	var workflow model.Workflow

	err := db.Where("project_id = ?", projectID).
		Where("title = ?", title).
		Where("is_deleted = ?", false).
		Not("id = ?", id).Find(&workflow).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false
		}
		logCtx.WithError(err).Error("Failed to fetch requested workflows")
		return true
	}

	return false
}

func (store *MemSQL) checkAndCacheIfAnyEventBasedOWorkflowsSetForTheCurrentEvent(event *model.Event, updatedEventPropertiesOnly *postgres.Jsonb, isUpdate bool) {

	eventNameId := event.EventNameId

	alerts, eventName, updatedUserProps, ErrCode := store.MatchEventTriggerAlertWithTrackPayload(event.ProjectId, eventNameId, event.UserId, &event.Properties, event.UserProperties, updatedEventPropertiesOnly, isUpdate)
	if ErrCode == http.StatusFound && alerts != nil {
		// log.WithFields(log.Fields{"project_id": event.ProjectId,
		// 	"event_trigger_alerts": *alerts}).Info("EventTriggerAlert found. Caching Alert.")

		for _, alert := range *alerts {
			success := store.CacheEventTriggerAlert(&alert, event, eventName, updatedUserProps)
			if !success {
				log.WithFields(log.Fields{"project_id": event.ProjectId,
					"event_trigger_alert": alert}).Error("Caching alert failure")
			}
		}
	}

	//Check for alerts set on All Page view event
	eventPropMap, err := U.DecodePostgresJsonbAsPropertiesMap(&event.Properties)
	if err == nil {
		if (*eventPropMap)[U.EP_IS_PAGE_VIEW] == true {
			eventNameId = ""
			alerts, eventName, updatedUserProps, ErrCode := store.MatchEventTriggerAlertWithTrackPayload(event.ProjectId, eventNameId, event.UserId, &event.Properties, event.UserProperties, updatedEventPropertiesOnly, isUpdate)
			if ErrCode == http.StatusFound && alerts != nil {
				// log.WithFields(log.Fields{"project_id": event.ProjectId,
				// 	"event_trigger_alerts": *alerts}).Info("EventTriggerAlert found. Caching Alert.")

				for _, alert := range *alerts {
					success := store.CacheEventTriggerAlert(&alert, event, eventName, updatedUserProps)
					if !success {
						log.WithFields(log.Fields{"project_id": event.ProjectId,
							"event_trigger_alert": alert}).Error("Caching alert failure")
					}
				}
			}
		}
	}

	store.FindAndCacheWorkflowsWithFiltersMatchingProperties(event.ProjectId, event, updatedEventPropertiesOnly, isUpdate)
}

func (store *MemSQL) FindAndCacheWorkflowsWithFiltersMatchingProperties(projectId int64, event *model.Event, UpdatedEventProps *postgres.Jsonb, isUpdate bool) {
	logFields := log.Fields{
		"project_id":       projectId,
		"event_name":       event.EventNameId,
		"event_properties": event.Properties,
		"user_properties":  event.UserProperties,
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

	workflows, _ := store.GetWorkflowsForTheCurrentEvent(projectId, event.EventNameId)
	if workflows == nil {
		return
	}

	userPropMap, eventPropMap, updatedEventProps := getMappedPropertiesFromJsonbOfEventObject(event.UserProperties, &event.Properties, UpdatedEventProps)

	for _, workflow := range workflows {
		var alertBody model.WorkflowAlertBody
		err := U.DecodePostgresJsonbToStructType(workflow.AlertBody, &alertBody)
		if err != nil {
			logCtx.WithError(err).Error("Jsonb decoding to struct failure")
			continue
		}
		if isUpdateOnlySessionPropertyPresentInMessagePropertyOrFilterProperty(isUpdate, alertBody, updatedEventProps, logCtx) {
			continue
		}

		userPropMap := store.addAllGroupPropertiesToUserPropertiesIfRequired(projectId, event.UserId, alertBody, userPropMap, logCtx)

		modifiedFilters := getModifiedFiltersForInPropertiesDefaultQueryMap(alertBody.Filters)

		criteria := E.MapFilterProperties(modifiedFilters)
		if E.EventMatchesFilterCriterionList(projectId, *userPropMap, *eventPropMap, criteria) {
			store.CacheWorkflowToBeSent(&workflow, event, userPropMap)
		}
	}
}

func (store *MemSQL) addAllGroupPropertiesToUserPropertiesIfRequired(projectId int64, userId string, alertBody model.WorkflowAlertBody, userProps *map[string]interface{}, logCtx *log.Entry) *map[string]interface{} {

	if alertBody.EventLevel == model.EventLevelAccount {
		domainsGroupUserID, domainsGroupID, err := store.getDomainsGroupUserIDForUser(projectId, userId)
		if err != nil {
			logCtx.Warn("no domains group for the alert")
			return userProps
		}
		groupProps := store.GetAllGroupPropertiesForGivenDomainGroupUserID(projectId, domainsGroupID, domainsGroupUserID)
		if groupProps != nil {
			updateUserPropMapWithGroupProperties(userProps, groupProps, logCtx)
		}
	}

	return userProps
}

func isUpdateOnlySessionPropertyPresentInMessagePropertyOrFilterProperty(isUpdate bool, alertBody model.WorkflowAlertBody, updatedEventProps *map[string]interface{}, logCtx *log.Entry) bool {
	var messageProperties model.WorkflowMessageProperties
	if alertBody.MessageProperties != nil {
		err := U.DecodePostgresJsonbToStructType(alertBody.MessageProperties, &messageProperties)
		if err != nil {
			logCtx.WithError(err).Error("Jsonb decoding to struct failure")
			return true
		}
	}

	mandatoryMessageProperties := messageProperties.MandatoryPropertiesCompany
	additionalMessageProperties := messageProperties.AdditionalPropertiesCompany

	if !isUpdate {
		isUpdateOnlyPropertyInMessageBody := false

		for _, msgProp := range mandatoryMessageProperties {
			if alertBody.Event == U.EVENT_NAME_SESSION && U.SESSION_PROPERTIES_SET_IN_UPDATE[msgProp.Factors] {
				isUpdateOnlyPropertyInMessageBody = true
			}
		}
		for _, msgProp := range additionalMessageProperties {
			if alertBody.Event == U.EVENT_NAME_SESSION && U.SESSION_PROPERTIES_SET_IN_UPDATE[msgProp.Factors] {
				isUpdateOnlyPropertyInMessageBody = true
			}
		}
		if isUpdateOnlyPropertyInMessageBody {
			return true
		}
	}
	if isUpdate {
		if len(*updatedEventProps) == 0 {
			return true
		} else {
			isPropertyInFilterUpdated := false
			isUpdateOnlyPropertyInMessageBody := false
			for _, msgProp := range mandatoryMessageProperties {
				if alertBody.Event == U.EVENT_NAME_SESSION && U.SESSION_PROPERTIES_SET_IN_UPDATE[msgProp.Factors] {
					isUpdateOnlyPropertyInMessageBody = true
				}
			}
			for _, msgProp := range additionalMessageProperties {
				if alertBody.Event == U.EVENT_NAME_SESSION && U.SESSION_PROPERTIES_SET_IN_UPDATE[msgProp.Factors] {
					isUpdateOnlyPropertyInMessageBody = true
				}
			}
			for _, fil := range alertBody.Filters {
				_, exists := (*updatedEventProps)[fil.Property]
				if fil.Entity == model.PropertyEntityEvent && exists {
					isPropertyInFilterUpdated = true
				}
			}
			if !isPropertyInFilterUpdated && !isUpdateOnlyPropertyInMessageBody {
				return true
			}
		}
	}

	return false
}

func getModifiedFiltersForInPropertiesDefaultQueryMap(filters []model.QueryProperty) []model.QueryProperty {
	//IN_PROPERTIES_DEFAULT_QUERY_MAP selected properties check
	for i, filter := range filters {
		_logicalOp := filter.LogicalOp
		if q, exists := model.IN_PROPERTIES_DEFAULT_QUERY_MAP[filter.Property]; exists {
			if filter.Value == "true" {
				filters[i] = q
			} else if filter.Value == "false" || filter.Value == "$none" {
				filters[i] = q
				filters[i].Operator = model.EqualsOpStr
			}
		}

		filters[i].LogicalOp = _logicalOp
		if filters[i].LogicalOp == "" {
			filters[i].LogicalOp = "AND"
		}
	}

	return filters
}

func getMappedPropertiesFromJsonbOfEventObject(userProps, eventProps, UpdatedEventProps *postgres.Jsonb) (userPropMap *map[string]interface{}, eventPropMap *map[string]interface{}, updatedEventProps *map[string]interface{}) {

	if userProps != nil {
		userPropMap, _ = U.DecodePostgresJsonb(userProps)
	}
	if eventProps != nil {
		eventPropMap, _ = U.DecodePostgresJsonb(eventProps)
	}
	if UpdatedEventProps != nil {
		updatedEventProps, _ = U.DecodePostgresJsonb(UpdatedEventProps)
	}

	return userPropMap, eventPropMap, updatedEventProps
}

func (store *MemSQL) GetWorkflowsForTheCurrentEvent(projectId int64, id string) ([]model.Workflow, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"prefix":     id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var workflows []model.Workflow
	var eventName *model.EventName
	var errCode int
	var err error

	if id != "" {
		eventName, errCode, err = store.GetEventNameByID(projectId, id)
		if errCode != http.StatusFound || err != nil {
			log.WithFields(logFields).WithError(err).Error("event_name not found")
			return nil, http.StatusInternalServerError
		}
	} else {
		eventName = &model.EventName{
			Name: U.EVENT_NAME_PAGE_VIEW,
		}
	}

	workflows, errCode, err = store.GetWorkflowsSetOnCurrentEventName(projectId, eventName.Name)
	if (errCode != http.StatusFound && errCode != http.StatusNotFound) || err != nil {
		log.WithFields(logFields).WithError(err).Error("workflows not found")
		return nil, errCode
	}

	if len(workflows) == 0 {
		return nil, http.StatusNotFound
	}

	return workflows, http.StatusFound
}

func (store *MemSQL) GetWorkflowsSetOnCurrentEventName(projectID int64, eventName string) ([]model.Workflow, int, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	workflows := make([]model.Workflow, 0)
	db := C.GetServices().Db

	if err := db.Where("project_id = ? AND is_deleted = 0", projectID).
		Where("JSON_EXTRACT_STRING(alert_body, 'event') LIKE ?", eventName).
		Not("internal_status = ?", model.Disabled).
		Find(&workflows).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, err
		}
		log.WithFields(logFields).WithError(err).
			Error("can not get event trigger alerts by event names")
		return nil, http.StatusInternalServerError, err
	}

	return workflows, http.StatusFound, nil
}

func (store *MemSQL) CacheWorkflowToBeSent(workflow *model.Workflow, event *model.Event, allUserProperties *map[string]interface{}) bool {

	logFields := log.Fields{
		"project_id": workflow.ProjectID,
		"workflow":   *workflow,
		"event":      *event,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var wfBody model.WorkflowAlertBody
	if err := U.DecodePostgresJsonbToStructType(workflow.AlertBody, &wfBody); err != nil {
		logCtx.WithError(err).Error("Error in decoding jsonb to workflow body type.")
		return false
	}

	tt := time.Now()
	timestamp := tt.UnixNano()
	date := tt.UTC().Format(U.DATETIME_FORMAT_YYYYMMDD)

	messageProps, breakdownProps, err := store.GetWorkflowMessageAndBreakdownPropertiesMap(workflow.ProjectID, event, &wfBody, allUserProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get message and/or breakdown properties map.")
		return false
	}

	isCoolDownTimeExpired, key, sortedSetTuple, err := getCacheKeyAndSortedSetTupleAndCheckCoolDownTimeCondition(
		event.ProjectId, wfBody.DontRepeatAlerts, wfBody.CoolDownTime, timestamp, workflow.ID, &breakdownProps)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch caching key and/or sorted set tuple.")
		return false
	}

	if !isCoolDownTimeExpired {
		logCtx.WithFields(logFields).Info("Workflow caching cancelled due to cool down timer")
		return true
	}

	counterKey, err := model.GetEventTriggerAlertCacheCounterKey(event.ProjectId, workflow.ID, date)
	if err != nil {
		logCtx.WithError(err).Error("Failed to construct counter caching key.")
		return false
	}

	cc, err := cacheRedis.IncrPersistentBatch(counterKey)
	if err != nil {
		logCtx.WithError(err).Error("Failed while getting count value from the counter key.")
		return false
	}
	count := cc[0]

	if count == 1 {
		_, err := cacheRedis.SetExpiryPersistent(counterKey, oneDayInSeconds)
		if err != nil {
			logCtx.WithError(err).Error("Failed to set expiry for the counter key.")
			return false
		}
	}

	if wfBody.SetAlertLimit && count > wfBody.AlertLimit {
		logCtx.WithFields(logFields).
			Info("Alert was not cached for current workflow as daily AlertLimit has been reached.")

		return true
	}
	_, err = cacheRedis.ZincrPersistentBatch(true, sortedSetTuple)
	if err != nil {
		logCtx.WithError(err).Error("error while getting zincr")
		return false
	}

	logCtx.WithFields(log.Fields{
		"breakdown_props": breakdownProps,
		"message_props":   messageProps,
		"workflow_id":     workflow.ID,
		"counter_key":     *counterKey,
		"cache_key":       *key,
	}).Info("$$Check workflow message props.")

	successCode, err := store.AddWorkflowToCache(workflow, &messageProps, key)
	if err != nil || successCode != http.StatusCreated {
		logCtx.WithError(err).Error("Failed to add workflow payload to cache.")
		return false
	}

	return true
}

func (store *MemSQL) AddWorkflowToCache(workflow *model.Workflow, msgProps *U.PropertiesMap, key *cache.Key) (int, error) {
	logFields := log.Fields{
		"workflow":  *workflow,
		"cache_key": key,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	message := model.EventTriggerAlertMessage{
		Title:           workflow.Name,
		MessageProperty: *msgProps,
		Message:         workflow.WorkflowUrl,
	}

	cachePackage := model.CachedEventTriggerAlert{
		Message:    message,
		IsWorkflow: true,
	}

	err := model.SetCacheForEventTriggerAlert(key, &cachePackage)
	if err != nil {
		logCtx.WithFields(logFields).WithError(err).Error("Failed to set cache for workflow.")
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

func (store *MemSQL) GetWorkflowMessageAndBreakdownPropertiesMap(projectID int64, event *model.Event, workflowBody *model.WorkflowAlertBody, updatedUserProps *map[string]interface{}) (U.PropertiesMap, map[string]interface{}, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"workflow":   *workflowBody,
		"event":      *event,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var messageProperties model.WorkflowMessageProperties
	if workflowBody.MessageProperties != nil {
		err := U.DecodePostgresJsonbToStructType(workflowBody.MessageProperties, &messageProperties)
		if err != nil {
			logCtx.WithError(err).Error("Jsonb decoding to struct failure")
			return nil, nil, err
		}
	}

	var eventPropMap *map[string]interface{}
	var err error
	if &event.Properties != nil && len(event.Properties.RawMessage) != 0 {
		eventPropMap, err = U.DecodePostgresJsonb(&event.Properties)
		if err != nil {
			logCtx.WithError(err).Error("Jsonb decoding to propMap failure")
			return nil, nil, err
		}
	}

	allPropertiesCombined := updatedUserProps
	updateUserPropMapWithGroupProperties(allPropertiesCombined, eventPropMap, logCtx)
	if allPropertiesCombined == nil {
		logCtx.Error("No properties found for the event and user.")
		return nil, nil, fmt.Errorf("no properties found")
	}

	msgPropMap := store.getWorkflowMessageProperties(projectID, messageProperties, allPropertiesCombined)
	if msgPropMap == nil {
		logCtx.Error("Nil payload map found for the Workflow")
		return nil, nil, fmt.Errorf("nil received for payload properties map")
	}

	if workflowBody.EventLevel == model.EventLevelAccount && event.UserId != "" {
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

	breakdownPropMap := getWorkflowBreakdownProperties(workflowBody, eventPropMap, updatedUserProps)

	return msgPropMap, breakdownPropMap, nil
}

func getWorkflowBreakdownProperties(workflow *model.WorkflowAlertBody, eventPropMap, updatedUserProps *map[string]interface{}) map[string]interface{} {
	breakdownPropMap := make(map[string]interface{}, 0)

	for _, breakdownProperty := range workflow.BreakdownProperties {
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

	return breakdownPropMap
}

func (store *MemSQL) getWorkflowMessageProperties(projectID int64,
	messagePropertiesQuery model.WorkflowMessageProperties, allProperties *map[string]interface{}) U.PropertiesMap {

	mandatoryProps := messagePropertiesQuery.MandatoryPropertiesCompany
	mandatoryPropsPayload := make(model.WorkflowPayloadProperties)

	for _, mp := range mandatoryProps {
		propertyValue := store.getPropertiesPayloadForTheGivenProperty(projectID, mp, allProperties)
		mandatoryPropsPayload[mp.Others] = propertyValue
	}

	additionalProps := messagePropertiesQuery.AdditionalPropertiesCompany
	additionalPropsPayload := make(model.WorkflowPayloadProperties)
	for _, mp := range additionalProps {
		propertyValue := store.getPropertiesPayloadForTheGivenProperty(projectID, mp, allProperties)
		additionalPropsPayload[mp.Others] = propertyValue
	}

	additionalContactProps := messagePropertiesQuery.AdditionalPropertiesContact
	additionalContactPropsPayload := make(model.WorkflowPayloadProperties)
	for _, acp := range additionalContactProps {
		propertyValue := store.getPropertiesPayloadForTheGivenProperty(projectID, acp, allProperties)
		additionalPropsPayload[acp.Others] = propertyValue
	}

	payloadProperties := model.WorkflowParagonPayload{
		MandatoryPropsCompany:  mandatoryPropsPayload,
		AdditionalPropsCompany: additionalPropsPayload,
		AdditionalPropsContact: additionalContactPropsPayload,
	}

	msgPropMap, err := U.EncodeStructTypeToMap(payloadProperties)
	if err != nil {
		log.WithError(err).Error("Failed to encode struct to map.")
	}

	log.WithFields(log.Fields{
		"project_id":      projectID,
		"user_properties": allProperties,
		"msg_prop_map":    msgPropMap,
		"payload":         payloadProperties,
	}).Info("$$Check message properties in meth.")

	return msgPropMap
}

func (store *MemSQL) getPropertiesPayloadForTheGivenProperty(projectID int64, propertyMapping model.WorkflowPropertiesMapping, allProperties *map[string]interface{}) string {
	var propVal interface{}
	var exi bool

	if allProperties != nil {
		propVal, exi = (*allProperties)[propertyMapping.Factors]
	}

	//Serve values for properties which need to be modified internally
	if filter, exists := model.IN_PROPERTIES_DEFAULT_QUERY_MAP[propertyMapping.Factors]; !exi && exists {
		if satisfiesInternalProperty(filter.Value, (*allProperties)[filter.Property], filter.Operator) {
			propVal = true
			exi = true
		}
	}
	// check and get for display name labels for crm property keys
	if value := U.GetPropertyValueAsString(propVal); U.IsAllowedCRMPropertyPrefix(propertyMapping.Factors) && exi && value != "" {
		propertyLabel, exist := store.getDisplayNameLabelForThisProperty(projectID, propertyMapping.Factors, value)
		if exist {
			propVal = propertyLabel
		}
	}

	return getDisplayLikePropertyValueForWorkflowProperty(propertyMapping.Factors, propVal)
}

func getDisplayLikePropertyValueForWorkflowProperty(property string, value interface{}) string {
	if strings.Contains(property, "time") {
		res, _ := U.GetPropertyValueAsInt64(value)
		displayValue := U.GetDateOnlyHyphenFormatFromTimestampZ(res)
		return displayValue
	}

	displayValue := U.GetPropertyValueAsString(value)
	return displayValue
}
