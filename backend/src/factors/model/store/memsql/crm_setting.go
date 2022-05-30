package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetAllCRMSetting() ([]model.CRMSetting, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	var crmSettings []model.CRMSetting
	db := C.GetServices().Db
	err := db.Model(model.CRMSetting{}).Find(&crmSettings).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		log.WithError(err).Error("Failed to get all crm settings.")
		return nil, http.StatusInternalServerError
	}

	if len(crmSettings) == 0 {
		return nil, http.StatusNotFound
	}

	return crmSettings, http.StatusFound
}

func (store *MemSQL) GetCRMSetting(projectID uint64) (*model.CRMSetting, int) {
	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 {
		logCtx.Error("Invalid paremeters")
		return nil, http.StatusBadRequest
	}

	var crmSetting model.CRMSetting

	db := C.GetServices().Db
	err := db.Model(model.CRMSetting{}).Where("project_id = ? ", projectID).Find(&crmSetting).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to update crm settngs.")
		return nil, http.StatusInternalServerError
	}

	return &crmSetting, http.StatusFound
}

func (store *MemSQL) UpdateCRMSetting(projectID uint64, option model.CRMSettingOption) int {
	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 {
		logCtx.Error("Invalid paremeters")
		return http.StatusBadRequest
	}

	updates := map[string]interface{}{}
	option(updates)

	db := C.GetServices().Db

	err := db.Model(model.CRMSetting{}).Where("project_id = ? ", projectID).Update(updates).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update crm settngs.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (store *MemSQL) CreateCRMSetting(projectID uint64, crmSetting *model.CRMSetting) int {
	logFields := log.Fields{"project_id": projectID, "crm_setting": crmSetting}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || crmSetting == nil {
		logCtx.Error("Invalid parameters.")
		return http.StatusBadRequest
	}

	crmSetting.ProjectID = projectID

	db := C.GetServices().Db
	err := db.Create(&crmSetting).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to create crm settngs.")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func (store *MemSQL) CreateOrUpdateCRMSetting(projectID uint64, crmSetting *model.CRMSetting) int {
	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	_, status := store.GetCRMSetting(projectID)
	if status != http.StatusFound {
		if status != http.StatusNotFound {
			logCtx.Error("Failed to get crm settings on CreateOrUpdateCRMSetting.")
			return status
		}

		return store.CreateCRMSetting(projectID, crmSetting)
	}
	return store.UpdateCRMSetting(projectID, model.HubspotEnrichHeavy(crmSetting.HubspotEnrichHeavy))
}
