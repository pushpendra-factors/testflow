package model

import (
	"net/http"
	"time"

	C "factors/config"
	U "factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// BigquerySetting Model bigquery_settings table.
type BigquerySetting struct {
	ID                      string    `gorm:"primary_key:true;auto_increment:false" json:"id"`
	ProjectID               uint64    `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	BigqueryProjectID       string    `gorm:"column:bq_project_id" json:"bq_project_id"`
	BigqueryDatasetName     string    `gorm:"column:bq_dataset_name" json:"bq_dataset_name"`
	BigqueryCredentialsJSON string    `gorm:"column:bq_credentials_json" json:"bq_credentials_json"`
	LastRunAt               int64     `json:"last_run_at"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// CreateBigquerySetting Validates and creates a new bigquery entry for the given setting.
func CreateBigquerySetting(setting *BigquerySetting) (*BigquerySetting, int) {
	logCtx := log.WithFields(log.Fields{
		"Prefix":            "Model#BigquerySetting",
		"ProjectID":         setting.ProjectID,
		"BigqueryProjectId": setting.BigqueryProjectID,
	})

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
func UpdateBigquerySettingLastRunAt(settingID string, lastRunAt int64) (int64, int) {
	db := C.GetServices().Db
	db = db.Model(&BigquerySetting{}).Where("id = ?", settingID).
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
func GetBigquerySettingByProjectID(projectID uint64) (*BigquerySetting, int) {
	if projectID == 0 {
		log.Error("Invalid project id")
		return nil, http.StatusInternalServerError
	}
	db := C.GetServices().Db

	var setting BigquerySetting
	if err := db.Where("project_id = ?", projectID).First(&setting).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		log.WithError(err).Error("Failed to get Bigquery setting.")
		return nil, http.StatusInternalServerError
	}

	return &setting, http.StatusFound
}
