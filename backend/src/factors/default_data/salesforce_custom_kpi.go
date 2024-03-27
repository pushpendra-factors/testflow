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
	{
		Name:        model.SalesforceSQLDateEntered,
		Description: "Salesforce SQLs (Contact Entered Stage)",
		TypeOfQuery: 1,
		ObjectType:  model.SalesforceUsersDisplayCategory,
	},
	model.CustomMetric{
		Name:        model.SalesforceOpportunities,
		Description: "All Salesforce Opportunities timestamped at Opportunity Create Date",
		TypeOfQuery: 1,
		ObjectType:  model.SalesforceOpportunitiesDisplayCategory,
	},
	model.CustomMetric{
		Name:        model.SalesforcePipeline,
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
	{
		Name:        model.SalesforceRevenue,
		Description: "Salesforce Revenue",
		TypeOfQuery: 1,
		ObjectType:  model.SalesforceOpportunitiesDisplayCategory,
	},
	{
		Name:        model.SalesforceAvgDealSize,
		Description: "Salesforce Avg deal size",
		TypeOfQuery: 1,
		ObjectType:  model.SalesforceOpportunitiesDisplayCategory,
	},
	{
		Name:        model.SalesforceClosedWonDeals,
		Description: "Salesforce Closed Won Opportunities",
		TypeOfQuery: 1,
		ObjectType:  model.SalesforceOpportunitiesDisplayCategory,
	},
	{
		Name:        model.SalesforceAvgSalesCycleLength,
		Description: "Salesforce Avg Sales Cycle Length",
		TypeOfQuery: 1,
		MetricType:  model.DateTypeDiffMetricType,
		ObjectType:  model.SalesforceOpportunitiesDisplayCategory,
	},
	{
		Name:        model.SalesforceClosedRate,
		Description: "For segment level KPIs (Salesforce)",
		TypeOfQuery: 2,
		ObjectType:  model.SalesforceOpportunitiesDisplayCategory,
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
		Filters: []model.KPIFilter{
			{
				ObjectType:       "",
				PropertyName:     "$salesforce_lead_status",
				PropertyDataType: "categorical",
				Entity:           model.UserEntity,
				Condition:        model.EqualsOpStr,
				Value:            "Qualified",
				LogicalOp:        "AND",
			},
		},
		DateField: "$salesforce_lead_createddate",
		EventName: "",
		Entity:    model.UserEntity,
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
	{
		AggregateFunction:     model.SumAggregateFunction,
		AggregateProperty:     "$salesforce_opportunity_amount",
		AggregatePropertyType: "numerical",
		Filters: []model.KPIFilter{
			{
				ObjectType:       "",
				PropertyName:     "$salesforce_opportunity_stagename",
				PropertyDataType: "categorical",
				Entity:           model.UserEntity,
				Condition:        model.EqualsOpStr,
				Value:            "Closed Won",
				LogicalOp:        "AND",
			},
		},
		DateField: "$salesforce_opportunity_createddate",
		EventName: "",
		Entity:    model.UserEntity,
	},
	{
		AggregateFunction:     model.AverageAggregateFunction,
		AggregateProperty:     "$salesforce_opportunity_amount",
		AggregatePropertyType: "numerical",
		Filters:               []model.KPIFilter{},
		DateField:             "$salesforce_opportunity_createddate",
		EventName:             "",
		Entity:                model.UserEntity,
	},
	{
		AggregateFunction:     model.UniqueAggregateFunction,
		AggregateProperty:     "",
		AggregatePropertyType: "numerical",
		Filters: []model.KPIFilter{
			{
				ObjectType:       "",
				PropertyName:     "$salesforce_opportunity_stagename",
				PropertyDataType: "categorical",
				Entity:           model.UserEntity,
				Condition:        model.EqualsOpStr,
				Value:            "Closed Won",
				LogicalOp:        "AND",
			},
		},
		DateField: "$salesforce_opportunity_createddate",
		EventName: "",
		Entity:    model.UserEntity,
	},
	{
		AggregateFunction:      model.AverageAggregateFunction,
		AggregateProperty:      "$salesforce_opportunity_createddate",
		AggregatePropertyType:  "datetime",
		AggregateProperty2:     "$salesforce_opportunity_closedate",
		AggregatePropertyType2: "datetime",
		Filters:                []model.KPIFilter{},
		DateField:              "$salesforce_opportunity_createddate",
		EventName:              "",
		Entity:                 model.UserEntity,
	},
}

var defaultSalesforceDerivedKPITransformations = []model.KPIQueryGroup{
	{
		Class: model.QueryClassKPI,
		Queries: []model.KPIQuery{
			{
				Category:        model.ProfileCategory,
				DisplayCategory: model.SalesforceOpportunitiesDisplayCategory,
				// There?
				Metrics:          []string{model.SalesforceClosedWonDeals},
				Filters:          []model.KPIFilter{},
				GroupBy:          []model.KPIGroupBy{},
				GroupByTimestamp: "",
				Timezone:         "",
				Operator:         "",
				QueryType:        "",
				Name:             "a",
			},
			{
				Category:         model.ProfileCategory,
				DisplayCategory:  model.SalesforceOpportunitiesDisplayCategory,
				Metrics:          []string{model.SalesforceOpportunities},
				Filters:          []model.KPIFilter{},
				GroupBy:          []model.KPIGroupBy{},
				GroupByTimestamp: "",
				Timezone:         "",
				Operator:         "",
				QueryType:        "",
				Name:             "b",
			},
		},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
		Formula:       "a/b",
	},
}

type BuildDefaultSalesforceCustomKPI struct {
}

func (buildDefault BuildDefaultSalesforceCustomKPI) Build(projectID int64) int {
	if !CheckIfDefaultSalesforceDatasAreCorrect() {
		log.Warn("Failed because defaultDatas and transformations are of incorrect length: salesforce.")
		return http.StatusInternalServerError
	}
	return buildForCustomKPI(projectID, defaultSalesforceCustomKPIs, defaultSalesforceCustomKPITransformations, defaultSalesforceDerivedKPITransformations)
}

func CheckIfDefaultSalesforceDatasAreCorrect() bool {
	return len(defaultSalesforceCustomKPIs) == (len(defaultSalesforceCustomKPITransformations) + len(defaultSalesforceDerivedKPITransformations))
}
