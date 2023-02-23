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
	LimitPathAnalysisEntityList = 1000
)

func (store *MemSQL) GetAllPathAnalysisEntityByProject(projectID int64) ([]model.PathAnalysisEntityInfo, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	entity := make([]model.PathAnalysis, 0)

	err := db.Table("pathanalysis").
		Where("project_id = ? AND is_deleted = ?", projectID, false).
		Order("created_at DESC").Limit(LimitPathAnalysisEntityList).Find(&entity).Error
	if err != nil {
		log.WithError(err).Error("Failed to fetch rows from pathanalysis table for project")
		return nil, http.StatusInternalServerError
	}

	if len(entity) == 0 {
		return nil, http.StatusFound
	}
	createdByNames, errCode := store.addCreatedByNameInPathAnalysis(entity)
	if errCode != http.StatusFound {
		log.WithFields(logFields).Error("Cannot fetch created_by names")
		return nil, http.StatusInternalServerError
	}

	et := store.convertPathAnalysisToPathAnalysisEntityInfo(entity, createdByNames)
	return et, http.StatusFound
}

func (store *MemSQL) GetPathAnalysisEntity(projectID int64, id string) (model.PathAnalysis, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var entity model.PathAnalysis

	err := db.Model(&model.PathAnalysis{}).
		Where("id = ?", id).
		Where("project_id = ? AND is_deleted = ?", projectID, false).
		Take(&entity).Error
	if err != nil {
		log.WithError(err).Error("Failed to fetch entity from pathanalysis table for project")
		return entity, http.StatusInternalServerError
	}

	return entity, http.StatusFound
}

func (store *MemSQL) convertPathAnalysisToPathAnalysisEntityInfo(list []model.PathAnalysis, names map[string]string) []model.PathAnalysisEntityInfo {

	res := make([]model.PathAnalysisEntityInfo, 0)

	for _, obj := range list {
		var entity model.PathAnalysisQuery
		err := U.DecodePostgresJsonbToStructType(obj.PathAnalysisQuery, &entity)
		if err != nil {
			log.WithError(err).Error("Problem deserializing pathanalysis query.")
			return nil
		}
		e := model.PathAnalysisEntityInfo{
			Id:                obj.ID,
			Title:             obj.Title,
			Status:            obj.Status,
			CreatedBy:         names[obj.CreatedBy],
			Date:              obj.UpdatedAt,
			PathAnalysisQuery: entity,
		}
		res = append(res, e)
	}
	return res
}

func (store *MemSQL) CreatePathAnalysisEntity(userID string, projectId int64, entity *model.PathAnalysisQuery) (*model.PathAnalysis, int, string) {
	logFields := log.Fields{
		"project_id":   projectId,
		"pathanalysis": entity,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	log.Info("memsql Create function triggered.")

	if isDuplicateTitle(projectId, entity) {
		return nil, http.StatusConflict, "Please provide a different title"
	}

	if isDulplicatePathAnalysisQuery(projectId, entity) {
		return nil, http.StatusConflict, "Query already exists"
	}

	transTime := gorm.NowFunc()
	id := U.GetUUID()

	query, err := U.EncodeStructTypeToPostgresJsonb(entity)
	if err != nil {
		log.WithField("entity", entity).WithError(err).Error("PathAnalysisQuery conversion to Jsonb failed")
		return nil, http.StatusInternalServerError, "PathAnalysisQuery conversion to Jsonb failed"
	}

	obj := model.PathAnalysis{
		ID:                id,
		ProjectID:         projectId,
		Title:             entity.Title,
		Status:            "saved",
		CreatedBy:         userID,
		CreatedAt:         transTime,
		UpdatedAt:         transTime,
		PathAnalysisQuery: query,
		IsDeleted:         false,
	}

	if err := db.Create(&obj).Error; err != nil {
		log.WithField("entity", entity).WithError(err).Error("Create Failed")
		return nil, http.StatusInternalServerError, "Create Failed in db"
	}

	return &obj, http.StatusCreated, ""
}

func (store *MemSQL) DeletePathAnalysisEntity(projectID int64, id string) (int, string) {
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
	entity, errCode := store.GetPathAnalysisEntity(projectID, id)
	if errCode != http.StatusFound {
		return http.StatusBadRequest, "Invalid id"
	}

	err := db.Model(&model.PathAnalysis{}).Where("id = ? AND project_id = ?", entity.ID, projectID).
		Update(map[string]interface{}{"is_deleted": true, "updated_at": modTime}).Error

	if err != nil {
		return http.StatusInternalServerError, "Failed to delete saved entity"
	}
	return http.StatusAccepted, ""
}

func (store *MemSQL) GetProjectCountWithStatus(projectID int64, status []string) (int, int, string) {

	db := C.GetServices().Db
	if projectID == 0 {
		return 0, http.StatusBadRequest, "Invalid project ID"
	}

	var count int

	err := db.Model(&model.PathAnalysis{}).
		Where(`project_id=?`, projectID).
		Where(`status = ? OR status = ?`, status[0], status[1]).
		Where(`is_deleted = ?`, false).Count(&count)

	if err != nil {
		return count, http.StatusInternalServerError, "Error caused on count memsql"
	}
	return count, http.StatusAccepted, ""
}

func (store *MemSQL) addCreatedByNameInPathAnalysis(obj []model.PathAnalysis) (map[string]string, int) {
	logFields := log.Fields{
		"pathanalysis": obj,
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
		log.WithFields(logFields).WithField("err_code", errCode).Error("could not get agents for given agentUUIDs")
		return nil, errCode
	}

	agentUUIDsToName := make(map[string]string)

	for _, a := range agents {
		agentUUIDsToName[a.UUID] = a.FirstName + " " + a.LastName
	}

	return agentUUIDsToName, http.StatusFound
}

func isDulplicatePathAnalysisQuery(ProjectID int64, query *model.PathAnalysisQuery) bool {
	logFields := log.Fields{
		"project_id":   ProjectID,
		"pathanalysis": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var objects []model.PathAnalysis
	if err := db.Where("project_id = ?", ProjectID).
		Where("is_deleted = ?", false).
		Where("JSON_EXTRACT_DOUBLE(query, 'steps') = ?", query.NumberOfSteps).
		Where("JSON_EXTRACT_STRING(query, 'event_type') = ?", query.EventType).
		Find(&objects).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false
		}
	}

	for _, obj := range objects {
		var res model.PathAnalysisQuery
		if err := U.DecodePostgresJsonbToStructType(obj.PathAnalysisQuery, &res); err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed to decode pathanalysis query")
			continue
		}

		equal := (res.AvoidRepeatedEvents == query.AvoidRepeatedEvents) &&
			(res.EndTimestamp == query.EndTimestamp) &&
			(res.StartTimestamp == query.StartTimestamp) &&
			reflect.DeepEqual(res.Event, query.Event) &&
			reflect.DeepEqual(res.ExcludeEvents, query.ExcludeEvents) &&
			reflect.DeepEqual(res.IncludeEvents, query.IncludeEvents) &&
			reflect.DeepEqual(res.Filter, query.Filter)

		if equal {
			log.WithFields(logFields).Error("Same pathanalysis features request")
			return true
		}
	}
	return false
}

func isDuplicateTitle(projectId int64, entity *model.PathAnalysisQuery) bool {
	logFields := log.Fields{
		"project_id":   projectId,
		"pathanalysis": entity,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	title := entity.Title
	var objects []model.PathAnalysis
	if err := db.Where("project_id = ?", projectId).
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

func (store *MemSQL) GetAllSavedPathAnalysisEntityByProject(projectID int64) ([]model.PathAnalysis, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	entity := make([]model.PathAnalysis, 0)

	err := db.Table("pathanalysis").
		Where("project_id = ? AND is_deleted = ? AND status = ?", projectID, false, model.SAVED).
		Order("created_at ASC").Limit(LimitPathAnalysisEntityList).Find(&entity).Error
	if err != nil {
		log.WithError(err).Error("Failed to fetch rows from queries table for project")
		return nil, http.StatusInternalServerError
	}

	if len(entity) == 0 {
		return nil, http.StatusFound
	}

	return entity, http.StatusFound
}

func (store *MemSQL) UpdatePathAnalysisEntity(projectID int64, id string, status string) (int, string) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	updatedFields := make(map[string]interface{})
	updatedFields["status"] = status
	updatedFields["updated_at"] = gorm.NowFunc()

	dbErr := db.Table("pathanalysis").Where("project_id = ? AND id = ?", projectID, id).Update(updatedFields).Error
	if dbErr != nil {
		logCtx.WithError(dbErr).Error("updating pathanalysis failed")
		return http.StatusInternalServerError, dbErr.Error()
	}
	return http.StatusOK, ""
}
