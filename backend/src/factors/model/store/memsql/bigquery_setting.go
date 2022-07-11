package memsql

import (
	"net/http"

	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesBigquerySettingsForeignConstraints(setting model.BigquerySetting) int {
	logFields := log.Fields{
		"setting": setting,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, errCode := store.GetProject(setting.ProjectID)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

// CreateBigquerySetting Validates and creates a new bigquery entry for the given setting.
func (store *MemSQL) CreateBigquerySetting(setting *model.BigquerySetting) (*model.BigquerySetting, int) {
	logFields := log.Fields{
		"setting": setting,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if setting.ID == "" {
		setting.ID = U.GetUUID()
	}

	logCtx.Info("Creating new BigquerySetting.")
	if setting.ProjectID == 0 {
		logCtx.Error("Invalid project id.")
		return nil, http.StatusBadRequest
	} else if setting.BigqueryProjectID == "" || setting.BigqueryDatasetName == "" ||
		setting.BigqueryCredentialsJSON == "" {
		logCtx.Error("Invalid Biquery credentials.")
		return nil, http.StatusBadRequest
	} else if errCode := store.satisfiesBigquerySettingsForeignConstraints(*setting); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}

	db := C.GetServices().Db
	err := db.Create(&setting).Error
	if err != nil {
		logCtx.WithError(err).Errorf("Failed to create BigquerySetting.")
		return nil, http.StatusInternalServerError
	}

	return setting, http.StatusCreated
}

// UpdateBigquerySettingLastRunAt Updates LastRunAt for a given setting. Other fields are not updated.
func (store *MemSQL) UpdateBigquerySettingLastRunAt(settingID string, lastRunAt int64) (int64, int) {
	logFields := log.Fields{
		"setting_id":  settingID,
		"last_run_at": lastRunAt,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	db = db.Model(&model.BigquerySetting{}).Where("id = ?", settingID).
		Updates(map[string]interface{}{
			"last_run_at": lastRunAt,
		})

	if db.Error != nil {
		log.WithError(db.Error).Error("UpdateBigquerySettingLastRunAt Failed")
		return 0, http.StatusInternalServerError
	}

	return db.RowsAffected, http.StatusAccepted
}

// GetBigquerySettingByProjectID Return bigquery setting for a given project_id of projects table.
func (store *MemSQL) GetBigquerySettingByProjectID(projectID int64) (*model.BigquerySetting, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		log.Error("Invalid project id")
		return nil, http.StatusInternalServerError
	}
	db := C.GetServices().Db

	var setting model.BigquerySetting
	if err := db.Where("project_id = ?", projectID).First(&setting).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		log.WithError(err).Error("Failed to get Bigquery setting.")
		return nil, http.StatusInternalServerError
	}

	return &setting, http.StatusFound
}
