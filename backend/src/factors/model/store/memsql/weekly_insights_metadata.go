package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateWeeklyInsightsMetadata(wim *model.WeeklyInsightsMetadata) (int, string) {
	logFields := log.Fields{
		"wim": wim,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	if valid := isValidProjectScope(wim.ProjectId); !valid {
		return http.StatusBadRequest, "Invalid Projectid"
	}

	wim.ID = U.GetUUID()
	if err := db.Create(wim).Error; err != nil {
		if strings.Contains(err.Error(), "weekly_insights_metadata_project_id_stdate_enddate_unique_idx") {
			updateFields := map[string]interface{}{
				"insight_id": wim.InsightId,
				"updated_at": time.Now(),
			}
			query := db.Model(&model.WeeklyInsightsMetadata{}).Where("project_id = ? AND base_start_time = ? AND base_end_time = ? AND comparison_start_time = ? AND comparison_end_time = ? AND query_id = ?",
				wim.ProjectId, wim.BaseStartTime, wim.BaseEndTime, wim.ComparisonStartTime, wim.ComparisonEndTime, wim.QueryId).Updates(updateFields)
			if err := query.Error; err != nil {
				log.WithError(err).Error("Failed updating weekly insights metadata.")
				return http.StatusInternalServerError, err.Error()
			}

			if query.RowsAffected == 0 {
				return http.StatusNotFound, "No record updated"
			}
		} else {
			log.WithFields(log.Fields{"model.WeeklyInsightsMetadata": wim}).WithError(
				err).Error("Failed creating model.WeeklyInsightsMetadata.")
			return http.StatusInternalServerError, err.Error()
		}
	}

	return http.StatusCreated, ""
}

func (store *MemSQL) GetWeeklyInsightsMetadata(projectId int64) ([]model.WeeklyInsightsMetadata, int, string) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if valid := isValidProjectScope(projectId); !valid {
		return nil, http.StatusBadRequest, "Invalid project"
	}

	metadata := make([]model.WeeklyInsightsMetadata, 0)
	if err := db.Where("project_id = ?", projectId).Find(&metadata).Error; err != nil {
		logCtx.WithError(err).Error("Getting weekly insights metadata failed")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, "no records for the project"
		}
		return nil, http.StatusInternalServerError, err.Error()
	}
	return metadata, http.StatusFound, ""
}
