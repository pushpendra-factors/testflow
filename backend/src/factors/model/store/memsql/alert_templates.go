package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetAlertTemplates() ([]model.AlertTemplate, int) {
	db := C.GetServices().Db
	var alertTemplates []model.AlertTemplate
	err := db.Where("is_deleted = ?", false).Where("is_workflow = ?", false).Order("id").Find(&alertTemplates).Error
	if err != nil {
		log.WithError(err).Error("Failed to get alert templates.")
		return alertTemplates, http.StatusInternalServerError
	}
	return alertTemplates, http.StatusOK
}

func (store *MemSQL) DeleteAlertTemplate(id int) error {
	db := C.GetServices().Db
	err := db.Model(&model.AlertTemplate{}).Where("id = ?", id).Update("is_deleted", true).Error
	if err != nil {
		log.WithError(err).Error("Failed to Delete alert templates.")
		return err
	}
	return nil
}
