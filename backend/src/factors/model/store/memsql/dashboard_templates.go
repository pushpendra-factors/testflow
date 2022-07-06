package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)


func (store *MemSQL) CreateTemplate(template *model.DashboardTemplate) (*model.DashboardTemplate, int, string) {
	db := C.GetServices().Db

	if template.ID == "" {
		template.ID = U.GetUUID()
	}

	if err := db.Create(&template).Error; err != nil {
		errMsg := "Failed to insert template."
		log.WithField("template", template)
		log.WithField("error", err).Error("Failed to add the template to dashboard templates table")
		return nil, http.StatusInternalServerError, errMsg
	}

	return template, http.StatusCreated, ""
}

func (store *MemSQL) SearchTemplateWithTemplateID(templateId string) (model.DashboardTemplate, int) {
	db := C.GetServices().Db
	var dashboardTemplate model.DashboardTemplate
	err := db.Table("dashboard_templates").Where("id = ? AND is_deleted = ?", templateId, false).Find(&dashboardTemplate).Error
	if(err != nil){
		log.WithField("id", templateId).Error("Failed to fetch the template with given id")
		return dashboardTemplate, http.StatusInternalServerError
	}
	return dashboardTemplate, http.StatusFound
}

func (store *MemSQL) SearchTemplateWithTemplateDetails(templateID string) (model.DashboardTemplate, int){
	db := C.GetServices().Db

	var template model.DashboardTemplate
	if templateID == "" {
		log.WithField("Failed to search, Invalid template ID.", templateID)
		return template, http.StatusBadRequest
	}

	err := db.Table("dashboard_queries").Where("id = ?", templateID).Find(&template).Error
	if err != nil {
		return template, http.StatusNotFound
	}
	return template, http.StatusFound
}

func (store *MemSQL) GetAllTemplates() ([]model.DashboardTemplate, int) {
	db := C.GetServices().Db

	var dashboardTemplates []model.DashboardTemplate

	err := db.Order("created_at ASC").Where("is_deleted = ?", false).Find(&dashboardTemplates).Error
	if err != nil {
		log.WithError(err).Error("Failed to get dashboard templates.")
		return dashboardTemplates, http.StatusInternalServerError
	}

	return dashboardTemplates, http.StatusFound
}

func (store *MemSQL) DeleteTemplate(templateID string) int{
	db := C.GetServices().Db

	err := db.Model(&model.DashboardUnit{}).Where("id = ?", templateID).Update(map[string]interface{}{"is_deleted": true}).Error
	if err != nil {
		log.WithFields(log.Fields{"id": templateID}).WithError(err).Error("Failed to delete dashboard unit.")
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}