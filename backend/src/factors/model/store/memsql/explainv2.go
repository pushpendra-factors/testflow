package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"reflect"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const (
	LimitExplainV2EntityList = 1000
)

func (store *MemSQL) GetAllExplainV2EntityByProject(projectID int64) ([]model.ExplainV2EntityInfo, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	entity := make([]model.ExplainV2, 0)

	err := db.Table("explain_v2").
		Where("project_id = ? AND is_deleted = ?", projectID, false).
		Order("created_at DESC").Limit(LimitExplainV2EntityList).Find(&entity).Error
	if err != nil {
		log.WithError(err).Error("Failed to fetch rows from explain table for project")
		return nil, http.StatusInternalServerError
	}

	if len(entity) == 0 {
		return nil, http.StatusFound
	}
	createdByNames, errCode := store.addCreatedByNameInExplainV2(entity)
	if errCode != http.StatusFound {
		log.WithFields(logFields).Error("Cannot fetch created_by names")
		return nil, http.StatusInternalServerError
	}

	et := store.convertExplainV2ToExplainV2EntityInfo(entity, createdByNames)
	return et, http.StatusFound
}

func (store *MemSQL) convertExplainV2ToExplainV2EntityInfo(list []model.ExplainV2, names map[string]string) []model.ExplainV2EntityInfo {

	res := make([]model.ExplainV2EntityInfo, 0)
	var entity model.ExplainV2Query

	for _, obj := range list {
		err := U.DecodePostgresJsonbToStructType(obj.ExplainV2Query, &entity)
		if err != nil {
			log.WithError(err).Error("Problem deserializing explainV2 query.")
			return nil
		}
		e := model.ExplainV2EntityInfo{
			Id:             obj.ID,
			Title:          obj.Title,
			Status:         obj.Status,
			CreatedBy:      names[obj.CreatedBy],
			Date:           obj.UpdatedAt,
			ExplainV2Query: entity.Query,
			ModelID:        obj.ModelID,
			Raw_query:      entity.Raw_query,
		}
		res = append(res, e)
	}
	return res
}

func (store *MemSQL) GetExplainV2Entity(projectID int64, id string) (model.ExplainV2, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var entity model.ExplainV2

	err := db.Table("explain_v2").Model(&model.ExplainV2{}).
		Where("id = ?", id).
		Where("project_id = ? AND is_deleted = ?", projectID, false).
		Take(&entity).Error
	if err != nil {
		log.WithError(err).Error("Failed to fetch entity from ExplainV2 table for project")
		return entity, http.StatusInternalServerError
	}

	return entity, http.StatusFound
}

func (store *MemSQL) CreateExplainV2Entity(userID string, projectId int64, entity *model.ExplainV2Query) (*model.ExplainV2, int, string) {
	logFields := log.Fields{
		"project_id": projectId,
		"ExplainV2":  entity,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	log.Info("memsql Create function triggered.")

	if entity.Query.StartEvent == "" && entity.Query.EndEvent == "" {
		return nil, http.StatusBadRequest, "Both startevent and endevent are empty"

	}

	if isDuplicateTitleExplainV2(projectId, entity) {
		return nil, http.StatusConflict, "Please provide a different title"
	}

	if status, errMsg := store.isRuleValidv2(entity.Query, projectId); status == false {
		return nil, http.StatusBadRequest, errMsg
	}

	if isDulplicateExplainV2Query(projectId, entity) {
		return nil, http.StatusConflict, "Query already exists"
	}

	transTime := gorm.NowFunc()
	id := U.GetUUID()

	query, err := U.EncodeStructTypeToPostgresJsonb(entity)
	if err != nil {
		log.WithField("entity", entity).WithError(err).Error("ExplainV2Query conversion to Jsonb failed")
		return nil, http.StatusInternalServerError, "ExplainV2 conversion to Jsonb failed"
	}

	obj := model.ExplainV2{
		ID:             id,
		ProjectID:      projectId,
		Title:          entity.Title,
		Status:         "saved",
		CreatedBy:      userID,
		CreatedAt:      transTime,
		UpdatedAt:      transTime,
		ExplainV2Query: query,
		IsDeleted:      false,
		ModelID:        0,
	}

	if err := db.Table("explain_v2").Create(&obj).Error; err != nil {
		log.WithField("entity", entity).WithError(err).Error("Create Failed")
		return nil, http.StatusInternalServerError, "Create Failed in db"
	}

	return &obj, http.StatusCreated, ""
}

func (store *MemSQL) DeleteExplainV2Entity(projectID int64, id string) (int, string) {
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
	entity, errCode := store.GetExplainV2Entity(projectID, id)
	if errCode != http.StatusFound {
		return http.StatusBadRequest, "Invalid id"
	}

	err := db.Table("explain_v2").Model(&model.ExplainV2{}).Where("id = ? AND project_id = ?", entity.ID, projectID).
		Update(map[string]interface{}{"is_deleted": true, "updated_at": modTime}).Error

	if err != nil {
		return http.StatusInternalServerError, "Failed to delete saved entity"
	}
	return http.StatusAccepted, ""
}

func (store *MemSQL) GetExplainV2ProjectCountWithStatus(projectID int64, status []string) (int, int, string) {

	db := C.GetServices().Db
	if projectID == 0 {
		return 0, http.StatusBadRequest, "Invalid project ID"
	}

	var count int

	err := db.Table("explain_v2").Model(&model.ExplainV2{}).
		Where(`project_id=?`, projectID).
		Where(`status = ? OR status = ?`, status[0], status[1]).
		Where(`is_deleted = ?`, false).Count(&count)

	if err != nil {
		return count, http.StatusInternalServerError, "Error caused on count memsql"
	}
	return count, http.StatusAccepted, ""
}

func (store *MemSQL) addCreatedByNameInExplainV2(obj []model.ExplainV2) (map[string]string, int) {
	logFields := log.Fields{
		"ExplainV2": obj,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	agentUUIDs := make([]string, 0)
	for _, q := range obj {
		if q.CreatedBy != "" {
			agentUUIDs = append(agentUUIDs, q.CreatedBy)
		}
	}

	agents, errCode := store.GetAgentsByUUIDs(agentUUIDs)
	if errCode != http.StatusFound {
		log.WithFields(logFields).Error("could not get agents for given agentUUIDs")
		return nil, errCode
	}

	agentUUIDsToName := make(map[string]string)

	for _, a := range agents {
		agentUUIDsToName[a.UUID] = a.FirstName + " " + a.LastName
	}

	return agentUUIDsToName, http.StatusFound
}

func isDulplicateExplainV2Query(ProjectID int64, query *model.ExplainV2Query) bool {
	logFields := log.Fields{
		"project_id": ProjectID,
		"ExplainV2":  query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var objects []model.ExplainV2
	if err := db.Table("explain_v2").Where("project_id = ?", ProjectID).
		Where("is_deleted = ?", false).
		Where("JSON_EXTRACT_DOUBLE(query, 'st_en') = ?", query.Query.StartEvent).
		Where("JSON_EXTRACT_STRING(query, 'et_en') = ?", query.Query.EndEvent).
		Where("JSON_EXTRACT_STRING(query, 'ie_en') = ?", query.Query.Rule.IncludedEvents).
		Find(&objects).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false
		}
	}

	for _, obj := range objects {
		var res model.ExplainV2Query
		if err := U.DecodePostgresJsonbToStructType(obj.ExplainV2Query, &res); err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed to decode explainV2 query")
			continue
		}

		equal := (res.EndTimestamp == query.EndTimestamp) &&
			(res.StartTimestamp == query.StartTimestamp) &&
			reflect.DeepEqual(res.Query.Rule.IncludedEvents, query.Query.Rule.IncludedEvents)

		if equal {
			log.WithFields(logFields).Error("Same explainV2 features request")
			return true
		}
	}
	return false
}

func isDuplicateTitleExplainV2(projectId int64, entity *model.ExplainV2Query) bool {
	logFields := log.Fields{
		"project_id": projectId,
		"explainv2":  entity,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	title := entity.Title
	var objects []model.ExplainV2
	if err := db.Table("explain_v2").Where("project_id = ?", projectId).
		Where("is_deleted = ?", false).
		Where("title = ?", title).
		Find(&objects).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false
		}
	}
	for _, obj := range objects {
		if obj.Title == title {
			return true
		}
	}
	return false
}

func (store *MemSQL) GetAllSavedExplainV2EntityByProject(projectID int64) ([]model.ExplainV2, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	entity := make([]model.ExplainV2, 0)

	err := db.Table("explain_v2").
		Where("project_id = ? AND is_deleted = ? AND status = ?", projectID, false, model.SAVED).
		Order("created_at ASC").Limit(LimitExplainV2EntityList).Find(&entity).Error
	if err != nil {
		log.WithError(err).Error("Failed to fetch rows from queries table for project")
		return nil, http.StatusInternalServerError
	}

	if len(entity) == 0 {
		return nil, http.StatusFound
	}

	return entity, http.StatusFound
}

func (store *MemSQL) UpdateExplainV2EntityStatus(projectID int64, id string, status string, model_id uint64) (int, string) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	updatedFields := make(map[string]interface{})
	updatedFields["status"] = status
	updatedFields["model_id"] = model_id
	updatedFields["updated_at"] = gorm.NowFunc()

	dbErr := db.Table("explain_v2").Where("project_id = ? AND id = ?", projectID, id).Update(updatedFields).Error
	if dbErr != nil {
		logCtx.WithError(dbErr).Error("updating ExplainV2 failed")
		return http.StatusInternalServerError, dbErr.Error()
	}
	return http.StatusOK, ""
}

func (store *MemSQL) isRuleValidv2(Rule model.FactorsGoalRule, ProjectID int64) (bool, string) {
	logFields := log.Fields{
		"project_id": ProjectID,
		"rule":       Rule,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	res, msg := store.isEventObjectValidv2(Rule.EndEvent, Rule.Rule.EndEnEventFitler, ProjectID)
	if res == false {
		return res, msg
	}
	if Rule.StartEvent != "" {
		res, msg := store.isEventObjectValidv2(Rule.StartEvent, Rule.Rule.StartEnEventFitler, ProjectID)
		if res == false {
			return res, msg
		}
	}
	userProperties := make([]string, 0)
	for _, filter := range Rule.Rule.GlobalFilters {
		userProperties = append(userProperties, filter.Key)
	}
	res, msg = store.isUserPropertiesValidv2(ProjectID, userProperties)
	if res == false {
		logCtx.Error(msg)
		return false, msg
	}
	userProperties = make([]string, 0)
	for _, filter := range Rule.Rule.StartEnUserFitler {
		userProperties = append(userProperties, filter.Key)
	}
	res, msg = store.isUserPropertiesValidv2(ProjectID, userProperties)
	if res == false {
		logCtx.Error(msg)
		return false, msg
	}
	userProperties = make([]string, 0)
	for _, filter := range Rule.Rule.EndEnUserFitler {
		userProperties = append(userProperties, filter.Key)
	}
	res, msg = store.isUserPropertiesValidv2(ProjectID, userProperties)
	if res == false {
		logCtx.Error(msg)
		return false, msg
	}
	return true, ""
}

func (store *MemSQL) isUserPropertiesValidv2(ProjectID int64, UserProperties []string) (bool, string) {
	logFields := log.Fields{
		"project_id":      ProjectID,
		"user_properties": UserProperties,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	allUserPropertiesByCategory, err := store.GetUserPropertiesByProject(ProjectID, 10000, 30)
	if err != nil {
		logCtx.WithError(err).Error("Get user Properties from cache failed")
		return false, "user proeprties missing"
	}
	userPropertiesMap := make(map[string]bool)
	for _, properties := range allUserPropertiesByCategory {
		for _, property := range properties {
			userPropertiesMap[property] = true
		}
	}
	for _, userProperty := range UserProperties {
		if userPropertiesMap[userProperty] == false {
			logCtx.Error("User Property not associated with project")
			return false, "user property not associated to this project"
		}
	}
	return true, ""
}

func (store *MemSQL) isEventPropertiesValidV2(ProjectID int64, EventName string, EventProperties []string) (bool, string) {
	logFields := log.Fields{
		"project_id":       ProjectID,
		"event_name":       EventName,
		"event_properties": EventProperties,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	allEventPropertiesByCategory, err := store.GetPropertiesByEvent(ProjectID, EventName, 10000, 30)
	if err != nil {
		logCtx.WithError(err).Error("Get event Properties from cache failed")
		return false, "event proeprties missing"
	}
	eventPropertiesMap := make(map[string]bool)
	for _, properties := range allEventPropertiesByCategory {
		for _, property := range properties {
			eventPropertiesMap[property] = true
		}
	}
	for _, eventProperty := range EventProperties {
		if eventPropertiesMap[eventProperty] == false {
			logCtx.Error("event Property not associated with project")
			return false, "event property not associated to this project"
		}
	}
	return true, ""
}

func (store *MemSQL) isEventObjectValidv2(event string, eventFilters []model.KeyValueTuple, ProjectID int64) (bool, string) {
	logFields := log.Fields{
		"project_id":    ProjectID,
		"event":         event,
		"event_filters": eventFilters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	_, err := store.GetEventName(event, ProjectID)
	if err != http.StatusFound {
		logCtx.Error("Get Event details failed")
		return false, "event doesnt exist"
	}
	eventProperties := make([]string, 0)
	for _, filter := range eventFilters {
		eventProperties = append(eventProperties, filter.Key)
	}
	res, msg := store.isEventPropertiesValid(ProjectID, event, eventProperties)
	if res == false {
		logCtx.Error(msg)
		return false, msg
	}
	return true, ""
}
