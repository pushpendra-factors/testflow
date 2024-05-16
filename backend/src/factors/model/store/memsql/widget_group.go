package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const (
	AccountsWidgetGroupInternal            = "account"
	AccountsWidgetGroup                    = "Account Analysis"
	MarketingEngagementWidgetGroupInternal = "marketing"
	MarketingEngagementWidgetGroup         = "Marketing Engagement Analysis"
	SalesOppWidgetGroupInternal            = "sales"
	SalesOppWidgetGroup                    = "Sales Opportunity Analysis"
	TotalAccountsWidget                    = "Accounts currently in segment"
	HighEngagementAccountsWidget           = "Accounts with High engagement"
	OpportunityCreated                     = "Opportunity Created"
	PipelineCreated                        = "Pipeline Created"
	AverageDealSize                        = "Average Deal Size"
	RevenueBooked                          = "Revenue Booked"
	CloseRate                              = "Close Rate (%)"
	AvgSalesCycleLength                    = "Avg Sales Cycle Length"
	MarketingQualifiedLeads                = "Marketing qualified leads"
	SalesQualifiedLeads                    = "Sales qualified leads"
)

var integrationBasedWidgetGroupNames = []string{MarketingEngagementWidgetGroup, SalesOppWidgetGroup}
var integrationBasedWidgetGroupNamesInternal = []string{MarketingEngagementWidgetGroupInternal, SalesOppWidgetGroupInternal}

var marketingWidgetGroup = map[string][]model.Widget{
	model.HUBSPOT: {
		{
			QueryMetric: model.HubspotMQLDateEntered,
			DisplayName: MarketingQualifiedLeads,
		},
		{
			QueryMetric: model.HubspotSQLDateEntered,
			DisplayName: SalesQualifiedLeads,
		},
		{
			QueryMetric: model.HubspotDeals,
			DisplayName: OpportunityCreated,
		},
		{
			QueryMetric:     model.HubspotPipeline,
			DisplayName:     PipelineCreated,
			QueryMetricType: model.CurrencyBasedMetric,
		},
	},
	model.SALESFORCE: {
		{
			QueryMetric: model.SalesforceLeads,
			DisplayName: MarketingQualifiedLeads,
		},
		{
			QueryMetric: model.SalesforceSQLDateEntered,
			DisplayName: SalesQualifiedLeads,
		},
		{
			QueryMetric: model.SalesforceOpportunities,
			DisplayName: OpportunityCreated,
		},
		{
			QueryMetric:     model.SalesforcePipeline,
			DisplayName:     PipelineCreated,
			QueryMetricType: model.CurrencyBasedMetric,
		},
	},
}

var salesOppWidgetGroup = map[string][]model.Widget{
	model.HUBSPOT: {
		{
			QueryMetric:     model.HubspotAvgDealSize,
			DisplayName:     AverageDealSize,
			QueryMetricType: model.CurrencyBasedMetric,
		},
		{
			QueryMetric:     model.HubspotRevenue,
			DisplayName:     RevenueBooked,
			QueryMetricType: model.CurrencyBasedMetric,
		},
		{
			QueryMetric:     model.HubspotClosedRate,
			DisplayName:     CloseRate,
			QueryMetricType: model.PercentageBasedMetric,
		},
		{
			QueryMetric:     model.HubspotAvgSalesCycleLength,
			DisplayName:     AvgSalesCycleLength,
			QueryMetricType: model.DurationBasedMetric,
		},
	},
	model.SALESFORCE: {
		{
			QueryMetric:     model.SalesforceAvgDealSize,
			DisplayName:     AverageDealSize,
			QueryMetricType: model.CurrencyBasedMetric,
		},
		{
			QueryMetric:     model.SalesforceRevenue,
			DisplayName:     RevenueBooked,
			QueryMetricType: model.CurrencyBasedMetric,
		},
		{
			QueryMetric:     model.SalesforceClosedRate,
			DisplayName:     CloseRate,
			QueryMetricType: model.PercentageBasedMetric,
		},
		{
			QueryMetric:     model.SalesforceAvgSalesCycleLength,
			DisplayName:     AvgSalesCycleLength,
			QueryMetricType: model.DurationBasedMetric,
		},
	},
}

var MapOfWidgetNameToIntegrationToCustomKPI = map[string]map[string][]model.Widget{
	MarketingEngagementWidgetGroup: marketingWidgetGroup,
	SalesOppWidgetGroup:            salesOppWidgetGroup,
}

func (store *MemSQL) GetWidgetGroupAndWidgetsForConfig(projectID int64) ([]model.WidgetGroup, string, int) {
	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	var widgetGroups []model.WidgetGroup
	err := db.Order("display_name asc").Where("project_id = ?", projectID).Find(&widgetGroups).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).WithField("projectID", projectID).Warn("Failed while retrieving widget groups.")
			return make([]model.WidgetGroup, 0), "Invalid project ID for widget group", http.StatusNotFound
		}
		return make([]model.WidgetGroup, 0), "Error during get of widget groups", http.StatusInternalServerError
	}

	// widgetGroup.DecodeWidgetsAndSetDecodedWidgets()
	return widgetGroups, "", http.StatusFound
}

func (store *MemSQL) CreateWidgetGroups(projectID int64) ([]model.WidgetGroup, int) {

	resWidgetGroups := make([]model.WidgetGroup, 0)
	for index, widgetGroupName := range integrationBasedWidgetGroupNames {
		widgetGroup := model.WidgetGroup{}
		widgetGroup.ProjectID = projectID
		widgetGroup.DisplayName = widgetGroupName
		widgetGroup.Name = integrationBasedWidgetGroupNamesInternal[index]
		widgetGroup.ID = uuid.New().String()
		widgetGroup.CreatedAt = time.Now()
		widgetGroup.UpdatedAt = time.Now()
		widgetGroup.CreateWidgetJsonWithNoElements()

		widgetGroup, statusCode := store.CreateWidgetGroup(widgetGroup)
		if statusCode != http.StatusCreated {
			return resWidgetGroups, statusCode
		}
		resWidgetGroups = append(resWidgetGroups, widgetGroup)
	}

	// Creating account based widgetGroup
	accountsTotalWidget := model.Widget{
		ID:              uuid.New().String(),
		DisplayName:     TotalAccountsWidget,
		QueryMetric:     model.TotalAccountsMetric,
		QueryMetricType: "",
		IsNonEditable:   true,
		QueryType:       model.QueryClassAccounts,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	highEngagementAccountsWidget := model.Widget{
		ID:              uuid.New().String(),
		DisplayName:     HighEngagementAccountsWidget,
		QueryMetric:     model.HighEngagedAccountsMetric,
		QueryMetricType: "",
		IsNonEditable:   true,
		QueryType:       model.QueryClassAccounts,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	accountWidgets := []model.Widget{
		accountsTotalWidget, highEngagementAccountsWidget,
	}
	encodedAccountsWidgets, _ := U.EncodeStructTypeToPostgresJsonb(accountWidgets)

	accountWidgetGroup := model.WidgetGroup{
		ProjectID:       projectID,
		DisplayName:     AccountsWidgetGroup,
		Name:            AccountsWidgetGroupInternal,
		IsNonComparable: true,
		ID:              uuid.New().String(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Widgets:         encodedAccountsWidgets,
	}

	widgetGroup, statusCode := store.CreateWidgetGroup(accountWidgetGroup)
	if statusCode != http.StatusCreated {
		return resWidgetGroups, statusCode
	}
	resWidgetGroups = append(resWidgetGroups, widgetGroup)

	return resWidgetGroups, http.StatusCreated
}

func (store *MemSQL) CreateWidgetGroup(widgetGroup model.WidgetGroup) (model.WidgetGroup, int) {

	logCtx := log.WithFields(log.Fields{"widget_group": widgetGroup})
	logCtx.Warn("Widget Group created")

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)
	db := C.GetServices().Db

	if err := db.Create(&widgetGroup).Error; err != nil {
		logCtx.WithError(err).Error("Failed to create widget group.")
		return widgetGroup, http.StatusInternalServerError
	}

	return widgetGroup, http.StatusCreated
}

// Just checking one of the widgetGroups.
func (store *MemSQL) AreWidgetsAddedToWidgetGroup(projectID int64) (bool, string, int) {
	widgetGroup, errMsg, statusCode := store.GetWidgetGroupByName(projectID, MarketingEngagementWidgetGroup)
	if statusCode != http.StatusFound {
		return false, errMsg, statusCode
	}
	return widgetGroup.WidgetsAdded, "", http.StatusFound
}

// Used to add CRM based widgets alone
func (store *MemSQL) AddWidgetsToWidgetGroup(projectID int64, widgetGroupName, integrationName string) (model.WidgetGroup, string, int) {
	db := C.GetServices().Db

	widgetGroup, errMsg, statusCode := store.GetWidgetGroupByName(projectID, widgetGroupName)
	if statusCode != http.StatusFound {
		errMsg = errMsg + " Adding multiple Widgets - Failed during widget group by name: " + integrationName + " : "
		return model.WidgetGroup{}, errMsg, statusCode
	}
	logFields := log.Fields{"project_id": projectID, "widget_group_id": widgetGroup.ID}
	logCtx := log.WithFields(logFields)

	widgets := MapOfWidgetNameToIntegrationToCustomKPI[widgetGroupName][integrationName]
	resWidgets := make([]model.Widget, len(widgets))
	U.DeepCopy(&widgets, &resWidgets)
	for index, _ := range resWidgets {
		resWidgets[index].ID = uuid.New().String()
		resWidgets[index].QueryType = model.QueryClassKPI
		resWidgets[index].CreatedAt = time.Now()
		resWidgets[index].UpdatedAt = time.Now()
	}

	encodedWidgets, _ := U.EncodeStructTypeToPostgresJsonb(resWidgets)
	widgetGroup.Widgets = encodedWidgets

	widgetGroup.WidgetsAdded = true
	widgetGroup.UpdatedAt = time.Now()

	log.WithField("widgetGroup", widgetGroup).Warn("Add widgets to widget group")
	err := db.Save(&widgetGroup).Error
	if err != nil {
		logCtx.WithField("err", err).Warn("Failed during insert of widget group" + integrationName + " ")
		return widgetGroup, "Failed during insert of widget group", http.StatusInternalServerError
	}
	return widgetGroup, "", http.StatusCreated
}

func (store *MemSQL) GetWidgetGroups(projectID int64) ([]model.WidgetGroup, string, int) {
	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	if projectID == 0 {
		return make([]model.WidgetGroup, 0), "Invalid project ID for widget group", http.StatusBadRequest
	}
	var widgetGroups []model.WidgetGroup
	err := db.Where("project_id = ?", projectID).Find(&widgetGroups).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).WithField("projectID", projectID).Warn("Failed while retrieving widget groups.")
			return make([]model.WidgetGroup, 0), "Invalid project ID for widget group", http.StatusInternalServerError
		}
		return make([]model.WidgetGroup, 0), "Invalid ID for widget group", http.StatusNotFound
	}
	for index := range widgetGroups {
		widgetGroups[index].DecodeWidgetsAndSetDecodedWidgets()
	}
	return widgetGroups, "", http.StatusFound
}

func (store *MemSQL) GetWidgetGroupByName(projectID int64, name string) (model.WidgetGroup, string, int) {
	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	if projectID == 0 {
		return model.WidgetGroup{}, "Invalid project ID for widget group", http.StatusBadRequest
	}
	var widgetGroup model.WidgetGroup
	err := db.Where("project_id = ? AND display_name = ?", projectID, name).Find(&widgetGroup).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).WithField("projectID", projectID).WithField("name", name).Warn("Failed while retrieving widget groups.")
			return model.WidgetGroup{}, "Invalid project ID for widget group", http.StatusInternalServerError
		}
		return model.WidgetGroup{}, "Invalid ID for widget group", http.StatusBadRequest
	}
	widgetGroup.DecodeWidgetsAndSetDecodedWidgets()
	return widgetGroup, "", http.StatusFound
}

func (store *MemSQL) GetWidgetGroupByID(projectID int64, ID string) (model.WidgetGroup, string, int) {
	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	if projectID == 0 {
		return model.WidgetGroup{}, "Invalid project ID for widget group", http.StatusBadRequest
	}
	var widgetGroup model.WidgetGroup
	err := db.Where("project_id = ? AND id = ?", projectID, ID).Find(&widgetGroup).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).WithField("projectID", projectID).WithField("ID", ID).Warn("Failed while retrieving widget groups.")
			return model.WidgetGroup{}, "Processing error", http.StatusInternalServerError
		}
		return model.WidgetGroup{}, "Invalid ID for widget group", http.StatusBadRequest
	}
	widgetGroup.DecodeWidgetsAndSetDecodedWidgets()
	return widgetGroup, "", http.StatusFound
}

func (store *MemSQL) AddWidgetToWidgetGroup(widgetGroup model.WidgetGroup, inputWidget model.Widget) (model.Widget, string, int) {
	logFields := log.Fields{"project_id": widgetGroup.ProjectID, "widget_group_id": widgetGroup.ID, "input_widget": inputWidget}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	inputWidget.ID = uuid.New().String()
	inputWidget.CreatedAt = time.Now()
	inputWidget.UpdatedAt = time.Now()

	widgetGroup.DecodedWidgets = append(widgetGroup.DecodedWidgets, inputWidget)
	encodedWidgets, _ := U.EncodeStructTypeToPostgresJsonb(widgetGroup.DecodedWidgets)
	widgetGroup.Widgets = encodedWidgets

	err := db.Save(&widgetGroup).Error
	if err != nil {
		logCtx.WithField("err", err).Warn("Failed during insert of widget group")
		return inputWidget, "Failed during insert of widget group", http.StatusInternalServerError
	}
	return inputWidget, "", http.StatusCreated
}

func (store *MemSQL) GetWidgetAndWidgetGroupByWidgetID(projectID int64, widgetGroupID string, ID string) (model.WidgetGroup, model.Widget, string, int) {
	widgetGroup, errMsg, statusCode := store.GetWidgetGroupByID(projectID, widgetGroupID)
	if statusCode != http.StatusFound {
		return model.WidgetGroup{}, model.Widget{}, errMsg, statusCode
	}
	widget, statusCode2 := widgetGroup.GetWidget(ID)
	if statusCode2 != http.StatusFound {
		return model.WidgetGroup{}, model.Widget{}, "Invalid ID for widget", http.StatusBadRequest
	}
	return widgetGroup, widget, "", http.StatusFound
}

func (store *MemSQL) IsCustomMetricPresentInWidgetGroups(projectID int64, queryMetric string) (bool, int) {
	widgetGroups, err, statusCode := store.GetWidgetGroups(projectID)
	if statusCode == http.StatusNotFound {
		return false, http.StatusNotFound
	}
	if statusCode != http.StatusFound {
		log.WithField("projectId", projectID).WithField("err", err).Warn("Failed with the following error in IsCustomMetricPresentInWidgetGroups")
		return false, http.StatusInternalServerError
	}

	for _, widgetGroup := range widgetGroups {
		for _, widget := range widgetGroup.DecodedWidgets {
			if widget.QueryMetric == queryMetric {
				return true, http.StatusFound
			}
		}
	}

	return false, http.StatusNotFound
}

func (store *MemSQL) UpdateWidgetToWidgetGroup(widgetGroup model.WidgetGroup, inputWidget model.Widget) (model.Widget, string, int) {
	logFields := log.Fields{"project_id": widgetGroup.ProjectID, "widget_group_id": widgetGroup.ID, "input_widget": inputWidget}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	inputWidget.UpdatedAt = time.Now()

	widgetGroup.UpdateWidget(inputWidget)
	encodedWidgets, _ := U.EncodeStructTypeToPostgresJsonb(widgetGroup.DecodedWidgets)
	widgetGroup.Widgets = encodedWidgets

	// err := db.Where("project_id = ? AND id = ?", widgetGroup.ProjectID, widgetGroup.ID).Find(&widgetGroup).Error
	err := db.Save(&widgetGroup).Error
	if err != nil {
		logCtx.WithField("err", err).Warn("Failed during insert of widget group")
		return inputWidget, "Failed during insert of widget group", http.StatusInternalServerError
	}
	return inputWidget, "", http.StatusOK
}

func (store *MemSQL) DeleteWidgetFromWidgetGroup(projectID int64, widgetGroupID string, widgetID string) (string, int) {
	logFields := log.Fields{"projectID": projectID, "widget_group_id": widgetGroupID, "widget_id": widgetID}
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	widgetGroup, errMsg, statusCode := store.GetWidgetGroupByID(projectID, widgetGroupID)
	if statusCode != http.StatusFound {
		return errMsg, statusCode
	}
	widgetGroup.DeleteWidget(widgetID)
	encodedWidgets, _ := U.EncodeStructTypeToPostgresJsonb(widgetGroup.DecodedWidgets)
	widgetGroup.Widgets = encodedWidgets
	err := db.Save(&widgetGroup).Error
	if err != nil {
		logCtx.WithField("err", err).Warn("Failed while deleting of widget group")
		return "Failed while deleting of widget group", http.StatusInternalServerError
	}
	return "", http.StatusAccepted
}
