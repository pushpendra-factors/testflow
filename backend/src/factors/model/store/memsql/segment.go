package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"

	"errors"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateSegment(projectId int64, segmentPayload *model.SegmentPayload) (model.Segment, int, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"name":       segmentPayload.Name,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectId == 0 {
		logCtx.Error("segment creation failed. invalid projectId")
		return model.Segment{}, http.StatusBadRequest, errors.New("segment creation failed. invalid project_id")
	}

	isSegmentValid, err := isValidSegment(segmentPayload)
	if !isSegmentValid {
		return model.Segment{}, http.StatusBadRequest, err
	}

	if store.IsDuplicateSegmentNameCheck(projectId, segmentPayload.Name) {
		logCtx.Error("segment creation failed. Duplicate Name")
		return model.Segment{}, http.StatusBadRequest, errors.New("segment creation failed. Duplicate Name")
	}

	querySegment, err := U.EncodeStructTypeToPostgresJsonb(segmentPayload.Query)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed to encode segment query while segment creation")
		return model.Segment{}, http.StatusInternalServerError, err
	}
	segment := model.Segment{
		ProjectID:   projectId,
		Id:          U.GetUUID(),
		Name:        segmentPayload.Name,
		Description: segmentPayload.Description,
		Query:       querySegment,
		Type:        segmentPayload.Type,
		UpdatedAt:   U.TimeNowZ(),
	}

	db := C.GetServices().Db
	dbx := db.Create(&segment)
	if dbx.Error != nil {
		if IsDuplicateRecordError(dbx.Error) {
			return model.Segment{}, http.StatusConflict, errors.New("failed to create a segment. Duplicate Record")
		}
		logCtx.WithError(dbx.Error).Error("Failed to create a segment.")
		return model.Segment{}, http.StatusInternalServerError, errors.New("failed to create a segment")
	}

	return segment, http.StatusCreated, nil
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

	lastRunTime, lastRunStatusCode := store.GetMarkerLastForAllAccounts(projectId)
	// for case - segment is updated but all_run for the day is yet to run

	db := C.GetServices().Db
	var segments []model.Segment
	err := db.Table("segments").Where("project_id = ?", projectId).Find(&segments).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed while getting all segments by ProjectId.")
		return nil, http.StatusInternalServerError
	}
	allSegmentsMap := make(map[string][]model.Segment, 0)
	for _, segment := range segments {

		if lastRunStatusCode != http.StatusFound || segment.UpdatedAt.After(lastRunTime) {
			segment.IsLongRunComplete = false
		} else {
			segment.IsLongRunComplete = true
		}
		if _, ok := allSegmentsMap[segment.Type]; !ok {
			allSegmentsMap[segment.Type] = make([]model.Segment, 0)
		}
		allSegmentsMap[segment.Type] = append(allSegmentsMap[segment.Type], segment)
	}
	return allSegmentsMap, http.StatusFound
}

func (store *MemSQL) GetSegmentByGivenIds(projectId int64, segmentIds []string) (map[string][]model.Segment, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"ids":        segmentIds,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectId == 0 || len(segmentIds) == 0 {
		logCtx.Error("Failed to get segment by IDs. Invalid parameters.")
		return map[string][]model.Segment{}, http.StatusBadRequest
	}

	var segments []model.Segment
	db := C.GetServices().Db
	err := db.Where("project_id = ? AND id IN (?) ",
		projectId, segmentIds).Find(&segments).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return map[string][]model.Segment{}, http.StatusNotFound
		}

		logCtx.WithError(err).Error(
			"Failed at getting segment on GetSegmentByGivenIds.")
		return map[string][]model.Segment{}, http.StatusInternalServerError
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

func (store *MemSQL) UpdateMarkerRunSegment(projectID int64, ids []string, updateTime time.Time) int {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         ids,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectID == 0 || len(ids) == 0 {
		logCtx.Error("Failed to update marker_run_segment by ID. Invalid parameters.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db

	query := "UPDATE segments SET marker_run_segment = ? WHERE project_id = ? AND id IN (?);"

	params := []interface{}{updateTime, projectID, ids}

	err := db.Exec(query, params...).Error

	if err != nil {
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// NOTE: used for testing only.
func (store *MemSQL) GetSegmentByName(projectId int64, name string) (*model.Segment, int) {
	if projectId == 0 || name == "" {
		return nil, http.StatusBadRequest
	}

	var segment model.Segment
	db := C.GetServices().Db
	err := db.Limit(1).Where("project_id = ? AND name = ? ",
		projectId, name).Find(&segment).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		return nil, http.StatusInternalServerError
	}

	return &segment, http.StatusFound
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
	err := db.Limit(1).Where("project_id = ? AND id = ?",
		projectId, segmentId).Find(&segment).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error(
			"Failed at getting segment on GetSegmentById.")
		return nil, http.StatusInternalServerError
	}

	lastRunTime, lastRunStatusCode := store.GetMarkerLastForAllAccounts(projectId)
	if lastRunStatusCode != http.StatusFound || segment.UpdatedAt.After(lastRunTime) {
		segment.IsLongRunComplete = false
	} else {
		segment.IsLongRunComplete = true
	}

	return &segment, http.StatusFound
}

func (store *MemSQL) UpdateSegmentById(projectId int64, id string, segmentPayload model.SegmentPayload) (int, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"id":         id,
		"name":       segmentPayload.Name,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectId == 0 || id == "" {
		logCtx.Error("Failed to update segment by ID. Invalid parameters.")
		return http.StatusBadRequest, fmt.Errorf("failed to update segment. Invalid parameters")
	}

	db := C.GetServices().Db

	var segment model.Segment

	if segmentPayload.Name != "" {
		segment.Name = segmentPayload.Name
	}

	if segmentPayload.Description != "" {
		segment.Description = segmentPayload.Description
	}

	updatedQuery := segmentPayload.Query
	if segmentPayload.Type != "" {
		if len(updatedQuery.EventsWithProperties) == 0 && len(updatedQuery.GlobalUserProperties) == 0 {
			logCtx.Error("Failed to update segment by ID. Query is empty.")
			return http.StatusBadRequest, fmt.Errorf("failed to update segment. Query is empty")
		}
		segment.Type = segmentPayload.Type
	}

	if len(updatedQuery.EventsWithProperties) > 0 || len(updatedQuery.GlobalUserProperties) > 0 {
		querySegment, err := U.EncodeStructTypeToPostgresJsonb(segmentPayload.Query)
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed to encode segment query while segment updation")
			return http.StatusInternalServerError, err
		}

		segment.Query = querySegment
		segment.UpdatedAt = U.TimeNowZ()
	}

	err := db.Model(&model.Segment{}).Where("project_id = ? AND id = ? ",
		projectId, id).UpdateColumns(segment).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, err
		}

		logCtx.WithError(err).Error(
			"Failed while updating segment on UpdateSegmentById.")
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
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

func (store *MemSQL) ModifySegment(projectID int64, segment model.Segment) (int, error) {
	segmentQuery := model.Query{}
	err := U.DecodePostgresJsonbToStructType(segment.Query, &segmentQuery)
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID, "segment_id": segment.Id}).
			Error("Unable to decode segment query")
		return http.StatusInternalServerError, err
	}

	// nothing to update
	if len(segmentQuery.EventsWithProperties) == 0 {
		return http.StatusOK, nil
	}

	for index := range segmentQuery.EventsWithProperties {
		segmentQuery.EventsWithProperties[index].IsEventPerformed = true
		segmentQuery.EventsWithProperties[index].Range = 180
		segmentQuery.EventsWithProperties[index].Frequency = "0"
		segmentQuery.EventsWithProperties[index].FrequencyOperator = model.GreaterThanOpStr
	}

	encodedQuery, err := U.EncodeStructTypeToPostgresJsonb(segmentQuery)
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID, "segment_id": segment.Id}).
			Error("Unable to encode segment query")
		return http.StatusInternalServerError, err
	}

	segment.Query = encodedQuery

	db := C.GetServices().Db

	query := `UPDATE segments
	SET query = ?
	WHERE project_id = ? AND id = ?;`

	params := []interface{}{encodedQuery, projectID, segment.Id}

	err = db.Exec(query, params...).Error

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
