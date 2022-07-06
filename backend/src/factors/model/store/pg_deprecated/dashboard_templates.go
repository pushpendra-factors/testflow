package postgres

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)


func (pg *Postgres) CreateTemplate(template *model.DashboardTemplate) (*model.DashboardTemplate, int, string) {
	db := C.GetServices().Db

	// if template.ID == 0 {
	// 	return nil, http.StatusBadRequest, "Invalid template request."
	// }

	if err := db.Create(template).Error; err != nil {
		errMsg := "Failed to insert template."
		log.WithField("error", err).Error("Failed to add the template to dashboard templates table")
		return nil, http.StatusInternalServerError, errMsg
	}

	return template, http.StatusCreated, ""
}

func (pg *Postgres) SearchTemplateWithTemplateID(templateId string) (model.DashboardTemplate, int) {
	db := C.GetServices().Db
	var dashboardTemplate model.DashboardTemplate
	err := db.Table("dashboard_templates").Where("id = ? AND is_deleted = ?", templateId, false).Find(&dashboardTemplate).Error
	if(err != nil){
		log.WithField("id", templateId).Error("Failesd to fetch the template with given id")
		return dashboardTemplate, http.StatusInternalServerError
	}
	return dashboardTemplate, http.StatusFound
}

func (pg *Postgres) DeleteTemplate(templateID string) int {
	// make it similar to other delete methods
	db := C.GetServices().Db

	err := db.Model(&model.DashboardTemplate{}).Where("id = ?", templateID).Update(map[string]interface{}{"is_deleted": true}).Error

	if(err != nil){
		log.WithField("id", templateID).Error("Failed to delete template.")
		return http.StatusInternalServerError
	}

	return http.StatusFound
}

func (pg *Postgres) SearchTemplateWithTemplateDetails(templateID string) (model.DashboardTemplate, int){
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

func (pg *Postgres) GetAllTemplates() ([]model.DashboardTemplate, int) {
	db := C.GetServices().Db

	var dashboardTemplates []model.DashboardTemplate

	err := db.Order("created_at ASC").Where("is_deleted = ?", false).Find(&dashboardTemplates).Error
	if err != nil {
		log.WithError(err).Error("Failed to get dashboard templates.")
		return dashboardTemplates, http.StatusInternalServerError
	}

	return dashboardTemplates, http.StatusFound
}