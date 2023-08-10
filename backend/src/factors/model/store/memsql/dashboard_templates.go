package memsql

import (
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var categoryList = []string{model.CategoryWebAnalytics,
	model.CategoryPaidMarketing,
	model.CategoryOrganicPerformance,
	model.CategoryLandingPageAnalysis,
	model.CategoryCRMInsights,
	model.CategoryFullFunnelMarketing,
	model.CategoryExecutiveDashboards}

var requiredIntegrations = []string{model.IntegrationWebsiteSDK,
	model.IntegrationSegment,
	model.IntegrationMarketo,
	model.IntegrationHubspot,
	model.IntegrationSalesforce,
	model.IntegrationAdwords,
	model.IntegrationFacebook,
	model.IntegrationLinkedin,
	model.IntegrationGoogleSearchConsole,
	model.IntegrationBing,
	model.IntegrationClearbit,
	model.IntegrationLeadsquared,
	model.Integration6Signal}

func (store *MemSQL) CreateTemplate(template *model.DashboardTemplate) (*model.DashboardTemplate, int, string) {
	db := C.GetServices().Db

	if template.ID == "" {
		template.ID = U.GetUUID()
	}

	if templateCategoryValid, errMsg := isTemplateCategoryValid(template); !templateCategoryValid {
		return nil, http.StatusInternalServerError, errMsg
	}

	if templateRequiredIntegrationValid, errMsg := isTemplateRequiredIntegrationValid(template); !templateRequiredIntegrationValid {
		return nil, http.StatusInternalServerError, errMsg
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
	if err != nil {
		log.WithField("id", templateId).Error("Failed to fetch the template with given id")
		return dashboardTemplate, http.StatusInternalServerError
	}
	return dashboardTemplate, http.StatusFound
}

func (store *MemSQL) SearchTemplateWithTemplateDetails(templateID string) (model.DashboardTemplate, int) {
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

func (store *MemSQL) DeleteTemplate(templateID string) int {
	db := C.GetServices().Db

	err := db.Model(&model.DashboardUnit{}).Where("id = ?", templateID).Update(map[string]interface{}{"is_deleted": true}).Error
	if err != nil {
		log.WithFields(log.Fields{"id": templateID}).WithError(err).Error("Failed to delete dashboard unit.")
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

func isTemplateCategoryValid(template *model.DashboardTemplate) (bool, string) {
	templateCategory := make([]string, 0)

	if err := json.Unmarshal(template.Categories.RawMessage, &templateCategory); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warn("Cannot unmarshal JSON")
		errMsg := "Failed to unmarshal JSON Template Categories"
		return false, errMsg
	}

	if diff := U.StringSliceDiff(templateCategory, categoryList); len(diff) > 0 {
		log.WithField("Difference in Category List", diff)
		errMsg := "Dashboard Template category list contains foreign element"
		return false, errMsg
	}

	return true, ""
}

func isTemplateRequiredIntegrationValid(template *model.DashboardTemplate) (bool, string) {
	templateRequiredIntegration := make([]string, 0)

	if err := json.Unmarshal(template.RequiredIntegrations.RawMessage, &templateRequiredIntegration); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warn("Cannot unmarshal JSON")
		errMsg := "Failed to unmarshal JSON Template Required Integration"
		return false, errMsg
	}

	if diff := U.StringSliceDiff(templateRequiredIntegration, requiredIntegrations); len(diff) > 0 {
		log.WithField("Difference in Required Integration List", diff)
		errMsg := "Dashboard Template Required Integration contains foreign element"
		return false, errMsg
	}

	return true, ""
}

func (store *MemSQL) GenerateDashboardFromTemplate(projectID int64, agentUUID string, templateID string) (*model.Dashboard, int, error) {

	logCtx := log.WithFields(log.Fields{
		"projectID": projectID, "templateID": templateID,
	})

	dashboardTemplate, errCode := store.SearchTemplateWithTemplateID(templateID)
	if errCode != http.StatusFound {
		logCtx.Error("Search template using templateId failed. Invalid template Id.")
		return nil, errCode, errors.New("search template using templateId failed. Invalid template Id")
	}

	// Create a blank Dashboard
	var _dashboardDetails model.Dashboard
	err := json.Unmarshal(dashboardTemplate.Dashboard.RawMessage, &_dashboardDetails)
	if err != nil {
		logCtx.WithFields(log.Fields{"DashboardData": dashboardTemplate.Dashboard}).Error("Template has bad dashboard data. Exiting.")
		return nil, http.StatusBadRequest, errors.New("template has bad dashboard data. Exiting")
	}
	dashboardRequest := &model.Dashboard{
		Name:        _dashboardDetails.Name,
		Description: _dashboardDetails.Description,
		Type:        _dashboardDetails.Type,
		Settings:    _dashboardDetails.Settings,
	}

	dashboard, errCode := store.CreateDashboard(projectID, agentUUID, dashboardRequest)
	if errCode != http.StatusCreated {
		logCtx.WithFields(log.Fields{"DashboardData": _dashboardDetails, "ErrorCode": errCode}).Error("Failed to create Dashboard for given data.")
		return nil, errCode, errors.New("failed to create Dashboard for given data")
	}

	var unitsArray []model.UnitInfo
	err = json.Unmarshal(dashboardTemplate.Units.RawMessage, &unitsArray)
	if err != nil {
		logCtx.WithFields(log.Fields{"DashboardUnitData": dashboardTemplate.Units}).Error("Template has bad dashboard units data. Exiting.")
		return nil, http.StatusBadRequest, errors.New("template has bad dashboard units data")
	}

	// Creating queries from UnitInfo
	var dashQueries []model.Queries
	for _, unit := range unitsArray {
		queryRequest := &model.Queries{
			Query:     unit.Query,
			Title:     unit.Title,
			Type:      unit.QueryType,
			CreatedBy: agentUUID,
			// To support empty settings value.
			Settings: unit.QuerySettings,
			IdText:   U.RandomStringForSharableQuery(50),
		}
		query, errCode, errMsg := store.CreateQuery(projectID, queryRequest)
		if errCode != http.StatusCreated {
			logCtx.WithFields(log.Fields{"UnitData": unit, "Error": errMsg, "ErrorCode": errCode}).Error("Failed to create query for given Unit data.")
			return nil, errCode, errors.New("failed to create query for given unit data")
		}
		dashQueries = append(dashQueries, *query)
	}
	// Creating units from UnitInfo
	var dashUnits []model.DashboardUnit

	for idx, unit := range unitsArray {

		requestPayload := model.DashboardUnitRequestPayload{
			Description:  unit.Description,
			Presentation: unit.Presentation,
			QueryId:      dashQueries[idx].ID,
		}

		_dUnit, errCode, errMsg := store.CreateDashboardUnit(projectID, agentUUID,
			&model.DashboardUnit{
				DashboardId:  dashboard.ID,
				Presentation: requestPayload.Presentation,
				QueryId:      requestPayload.QueryId,
			})
		if errCode != http.StatusCreated {
			logCtx.WithFields(log.Fields{"UnitData": unit, "Error": errMsg, "ErrorCode": errCode}).Error("Failed to create query for given Unit data.")
			return nil, errCode, errors.New("failed to create query for given Unit data")
		}
		dashUnits = append(dashUnits, *_dUnit)
	}

	_dashboardDetails.Name = dashboardTemplate.Title
	_dashboardDetails.AgentUUID = agentUUID
	_dashboardDetails.Description = dashboardTemplate.Description
	_dashboardDetails.IsDeleted = false
	_dashboardDetails.ProjectId = projectID

	_unitPosition := make(map[string]map[int64]int, 0)
	_unitPosition["position"] = make(map[int64]int, 0)
	_unitPosition["size"] = make(map[int64]int, 0)
	for idx, unit := range dashUnits {
		pos := unitsArray[idx].Position
		size := unitsArray[idx].Size
		_unitPosition["position"][unit.ID] = pos
		_unitPosition["size"][unit.ID] = size
	}

	unitPos, _ := json.Marshal(_unitPosition)
	_dashboardDetails.UnitsPosition = &postgres.Jsonb{json.RawMessage(unitPos)}

	var requestPayload model.UpdatableDashboard
	requestPayload.Name = _dashboardDetails.Name
	requestPayload.Type = _dashboardDetails.Type
	requestPayload.Description = _dashboardDetails.Description
	requestPayload.UnitsPosition = &_unitPosition
	requestPayload.Settings = &_dashboardDetails.Settings

	errCode = store.UpdateDashboard(projectID, agentUUID, dashboard.ID, &requestPayload)
	if errCode != http.StatusAccepted {
		errMsg := "Update dashboard failed."
		logCtx.WithFields(log.Fields{"DashboardUpdateData": requestPayload, "Error": errMsg, "ErrorCode": errCode}).Error("Update dashboard failed")
		return nil, errCode, errors.New("update dashboard failed")
	}

	return dashboard, http.StatusCreated, nil

}
