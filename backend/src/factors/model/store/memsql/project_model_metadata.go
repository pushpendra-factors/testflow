package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
)

func (store *MemSQL) GetProjectModelMetadata(projectId int64) ([]model.ProjectModelMetadata, int, string) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if valid := isValidProjectScope(projectId); !valid {
		return nil, http.StatusBadRequest, "Invalid project"
	}

	modelMetadata := make([]model.ProjectModelMetadata, 0)
	if err := db.Where("project_id = ?", projectId).Find(&modelMetadata).Error; err != nil {
		logCtx.WithError(err).Error("Getting Project metadata failed")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, "no records for the project"
		}
		return nil, http.StatusInternalServerError, err.Error()
	}

	sort.Slice(modelMetadata, func(i, j int) bool {
		return modelMetadata[i].StartTime > modelMetadata[j].StartTime
	})
	// sort model metadata
	return modelMetadata, http.StatusFound, ""
}

func (store *MemSQL) GetAllProjectModelMetadata() ([]model.ProjectModelMetadata, int, string) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	db := C.GetServices().Db

	modelMetadata := make([]model.ProjectModelMetadata, 0)
	if err := db.Find(&modelMetadata).Error; err != nil {
		log.WithError(err).Error("Getting Project metadata failed")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, "no records for the project"
		}
		return nil, http.StatusInternalServerError, err.Error()
	}

	sort.Slice(modelMetadata, func(i, j int) bool {
		return modelMetadata[i].StartTime > modelMetadata[j].StartTime
	})
	// sort model metadata
	return modelMetadata, http.StatusFound, ""
}

func (store *MemSQL) CreateProjectModelMetadata(pmm *model.ProjectModelMetadata) (int, string) {
	logFields := log.Fields{
		"pmm": pmm,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	if valid := isValidProjectScope(pmm.ProjectId); !valid {
		return http.StatusBadRequest, "Invalid Projectid"
	}

	pmm.ID = U.GetUUID()
	if err := db.Create(pmm).Error; err != nil {
		if strings.Contains(err.Error(), "project_model_metadata_project_id_stdate_enddate_unique_idx") {
			updateFields := map[string]interface{}{
				"model_id":   pmm.ModelId,
				"chunks":     pmm.Chunks,
				"updated_at": U.TimeNowZ(),
			}
			query := db.Model(&model.ProjectModelMetadata{}).Where("project_id = ? AND start_time = ? AND end_time = ? AND model_type = ?",
				pmm.ProjectId, pmm.StartTime, pmm.EndTime, pmm.ModelType).Updates(updateFields)
			if err := query.Error; err != nil {
				log.WithError(err).Error("Failed updating model metadata.")
				return http.StatusInternalServerError, err.Error()
			}

			if query.RowsAffected == 0 {
				return http.StatusNotFound, "No record updated"
			}
		} else {
			log.WithFields(log.Fields{"model.ProjectModelMetadata": pmm}).WithError(
				err).Error("Failed creating model.ProjectModelMetadata.")
			return http.StatusInternalServerError, err.Error()
		}
	}

	return http.StatusCreated, ""
}
