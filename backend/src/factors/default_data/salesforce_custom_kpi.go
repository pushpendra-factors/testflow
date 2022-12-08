package default_data

import (
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var defaultSalesforceCustomKPIs = []model.CustomMetric{
	model.CustomMetric{
		Name:        "Salesforce Leads",
		Description: "All Salesforce Leads timestamped at Lead Create Date",
		TypeOfQuery: 1,
		ObjectType:  model.SalesforceUsersDisplayCategory,
	},
	model.CustomMetric{
		Name:        "Salesforce Opportunities",
		Description: "All Salesforce Opportunities timestamped at Opportunity Create Date",
		TypeOfQuery: 1,
		ObjectType:  model.SalesforceOpportunitiesDisplayCategory,
	},
	model.CustomMetric{
		Name:        "Salesforce Pipeline",
		Description: "Total pipeline generated from all Salesforce Opportunities timestamped at Opportunity Create Date",
		TypeOfQuery: 1,
		ObjectType:  model.SalesforceOpportunitiesDisplayCategory,
	},
	model.CustomMetric{
		Name:        "Salesforce Contacts",
		Description: "All Salesforce Contacts timestamped at Contact Create Date",
		TypeOfQuery: 1,
		ObjectType:  model.SalesforceUsersDisplayCategory,
	},
}

var defaultSalesforceCustomKPITransformations = []model.CustomMetricTransformation{
	model.CustomMetricTransformation{
		AggregateFunction:     model.UniqueAggregateFunction,
		AggregateProperty:     "",
		AggregatePropertyType: "categorical",
		Filters:               []model.KPIFilter{},
		DateField:             "$salesforce_lead_createddate",
		EventName:             "",
		Entity:                model.UserEntity,
	},
	model.CustomMetricTransformation{
		AggregateFunction:     model.UniqueAggregateFunction,
		AggregateProperty:     "",
		AggregatePropertyType: "categorical",
		Filters:               []model.KPIFilter{},
		DateField:             "$salesforce_opportunity_createddate",
		EventName:             "",
		Entity:                model.UserEntity,
	},
	model.CustomMetricTransformation{
		AggregateFunction:     model.SumAggregateFunction,
		AggregateProperty:     "$salesforce_opportunity_amount",
		AggregatePropertyType: "categorical",
		Filters:               []model.KPIFilter{},
		DateField:             "$salesforce_opportunity_createddate",
		EventName:             "",
		Entity:                model.UserEntity,
	},
	model.CustomMetricTransformation{
		AggregateFunction:     model.UniqueAggregateFunction,
		AggregateProperty:     "",
		AggregatePropertyType: "categorical",
		Filters:               []model.KPIFilter{},
		DateField:             "$salesforce_contact_createddate",
		EventName:             "",
		Entity:                model.UserEntity,
	},
}

var defaultSalesforceDerivedKPITransformations = make([]model.KPIQueryGroup, 0)

type BuildDefaultSalesforceCustomKPI struct {
}

func (buildDefault BuildDefaultSalesforceCustomKPI) Build(projectID int64) int {
	if !CheckIfDefaultLeadSquaredDatasAreCorrect() {
		log.Warn("Failed because defaultDatas and transformations are of incorrect length: salesforce.")
		return http.StatusInternalServerError
	}
	return buildForCustomKPI(projectID, defaultSalesforceCustomKPIs, defaultSalesforceCustomKPITransformations, defaultSalesforceDerivedKPITransformations)
}

func CheckIfDefaultSalesforceDatasAreCorrect() bool {
	return len(defaultSalesforceCustomKPIs) == (len(defaultSalesforceCustomKPITransformations) + len(defaultSalesforceDerivedKPITransformations))
}
