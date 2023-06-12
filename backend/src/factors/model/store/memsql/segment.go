package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"

	"errors"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateSegment(projectId int64, segmentPayload *model.SegmentPayload) (int, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"name":       segmentPayload.Name,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectId == 0 {
		logCtx.Error("segment creation failed. invalid projectId")
		return http.StatusBadRequest, errors.New("segment creation failed. invalid project_id")
	}

	isValidSegment, err := isValidSegment(segmentPayload)
	if !isValidSegment {
		return http.StatusBadRequest, err
	}

	if store.IsDuplicateSegmentNameCheck(projectId, segmentPayload.Name) {
		logCtx.Error("segment creation failed. Duplicate Name")
		return http.StatusBadRequest, errors.New("segment creation failed. Duplicate Name")
	}

	querySegment, err := U.EncodeStructTypeToPostgresJsonb(segmentPayload.Query)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed to encode segment query while segment creation")
		return http.StatusInternalServerError, err
	}
	segment := model.Segment{
		ProjectID:   projectId,
		Id:          U.GetUUID(),
		Name:        segmentPayload.Name,
		Description: segmentPayload.Description,
		Query:       querySegment,
		Type:        segmentPayload.Type,
	}

	db := C.GetServices().Db
	dbx := db.Create(&segment)
	if dbx.Error != nil {
		if IsDuplicateRecordError(dbx.Error) {
			return http.StatusConflict, errors.New("failed to create a segment. Duplicate Record")
		}
		logCtx.WithError(dbx.Error).Error("Failed to create a segment.")
		return http.StatusInternalServerError, errors.New("failed to create a segment")
	}

	return http.StatusCreated, nil
}

func isValidSegment(segmentPayload *model.SegmentPayload) (bool, error) {
	logFields := log.Fields{
		"segment": segmentPayload,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if segmentPayload.Type == "" {
		logCtx.Error("segment creation failed. No Type Added.")
		return false, errors.New("segment creation failed. Analyze Section Empty")
	}
	if segmentPayload.Name == "" {
		logCtx.Error("segment creation failed. No Name Added.")
		return false, errors.New("segment creation failed. Name Field Empty")
	}
	if len(segmentPayload.Query.EventsWithProperties) == 0 && len(segmentPayload.Query.GlobalUserProperties) == 0 {
		logCtx.Error("segment creation failed. No Query Added.")
		return false, errors.New("segment creation failed. Query Section Empty")
	}
	return true, nil
}

func (store *MemSQL) IsDuplicateSegmentNameCheck(projectID int64, name string) bool {
	segmentGroups, _ := store.GetAllSegments(projectID)
	for _, segments := range segmentGroups {
		for _, segment := range segments {
			if segment.Name == name {
				return true
			}
		}
	}
	return false
}

func (store *MemSQL) GetAllSegments(projectId int64) (map[string][]model.Segment, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectId == 0 {
		logCtx.Error("Failed to get all segments by ProjectId. Invalid projectId.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	var segments []model.Segment
	err := db.Table("segments").Where("project_id = ?", projectId).Find(&segments).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed while getting all segments by ProjectId.")
		return nil, http.StatusInternalServerError
	}
	allSegmentsMap := make(map[string][]model.Segment, 0)
	for _, segment := range segments {
		if _, ok := allSegmentsMap[segment.Type]; !ok {
			allSegmentsMap[segment.Type] = make([]model.Segment, 0)
		}
		allSegmentsMap[segment.Type] = append(allSegmentsMap[segment.Type], segment)
	}
	return allSegmentsMap, http.StatusFound
}

func (store *MemSQL) GetSegmentById(projectId int64, segmentId string) (*model.Segment, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"id":         segmentId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectId == 0 || segmentId == "" {
		logCtx.Error("Failed to get segment by ID. Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	var segment model.Segment
	db := C.GetServices().Db
	err := db.Limit(1).Where("project_id = ? AND id = ? ",
		projectId, segmentId).Find(&segment).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error(
			"Failed at getting segment on GetSegmentById.")
		return nil, http.StatusInternalServerError
	}

	return &segment, http.StatusFound
}

func (store *MemSQL) UpdateSegmentById(projectId int64, id string, segmentPayload model.SegmentPayload) (error, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"id":         id,
		"name":       segmentPayload.Name,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectId == 0 || id == "" {
		logCtx.Error("Failed to update segment by ID. Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	updateFields := make(map[string]interface{}, 0)

	if segmentPayload.Name != "" {
		updateFields["name"] = segmentPayload.Name
	}

	if segmentPayload.Description != "" {
		updateFields["description"] = segmentPayload.Description
	}

	updatedQuery := segmentPayload.Query
	if segmentPayload.Type != "" {
		if len(updatedQuery.EventsWithProperties) == 0 && len(updatedQuery.GlobalUserProperties) == 0 {
			logCtx.Error("Failed to update segment by ID. Query is empty.")
			return nil, http.StatusBadRequest
		}
		updateFields["type"] = segmentPayload.Type
	}

	if len(updatedQuery.EventsWithProperties) > 0 || len(updatedQuery.GlobalUserProperties) > 0 {
		querySegment, err := U.EncodeStructTypeToPostgresJsonb(segmentPayload.Query)
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed to encode segment query while segment updation")
			return err, http.StatusInternalServerError
		}
		updateFields["query"] = querySegment
	}

	err := db.Model(&model.Segment{}).Where("project_id = ? AND id = ? ",
		projectId, id).Update(updateFields).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return err, http.StatusNotFound
		}

		logCtx.WithError(err).Error(
			"Failed while updating segment on UpdateSegmentById.")
		return err, http.StatusInternalServerError
	}

	return nil, http.StatusOK
}

func (store *MemSQL) DeleteSegmentById(projectId int64, segmentId string) (int, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"id":         segmentId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectId == 0 || segmentId == "" {
		logCtx.Error("Failed to delete segment by ID. Invalid parameters.")
		return http.StatusBadRequest, nil
	}

	db := C.GetServices().Db
	dbx := db.Table("segments").Limit(1).Where("project_id = ? AND id = ? ",
		projectId, segmentId).Delete(&model.Segment{})

	if err := dbx.Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, err
		}
		logCtx.WithError(err).Error(
			"Failed to delete segment by ID.")
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
