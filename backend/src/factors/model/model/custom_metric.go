package model

import (
	U "factors/util"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type CustomKPI interface {
	ContainsNameInInternalTransformation(input string) bool
	ValidateFilterAndGroupBy() bool
}

type CustomMetricConfig struct {
	AggregateFunctions      []string                              `json:"agFn"`
	ObjectTypeAndProperties []CustomMetricObjectTypeAndProperties `json:"objTyAndProp"`
}

type CustomMetricObjectTypeAndProperties struct {
	ObjectType string              `json:"objTy"`
	Properties []map[string]string `json:"properties"`
}

type CustomMetricConfigV1 struct {
	ObjectType         string              `json:"obj_ty"`
	TypeOfQuery        int                 `json:"type_of_query"`
	AggregateFunctions []string            `json:"agFn"`
	Properties         []map[string]string `json:"properties"`
}

const (
	SumAggregateFunction     = "sum"
	UniqueAggregateFunction  = "unique"
	AverageAggregateFunction = "average"
	CountAggregateFunction   = "count"
	Derived                  = "derived"
	DerivedMetrics           = "derived_metrics"
	CustomEventType          = "custom_events"
)

var (
	CustomMetricProfilesAggregateFunctions = []string{SumAggregateFunction, UniqueAggregateFunction, AverageAggregateFunction}
	CustomEventsAggregateFunctions         = []string{SumAggregateFunction, UniqueAggregateFunction, AverageAggregateFunction, CountAggregateFunction}
	CustomKPIProfileObjectCategories       = []string{HubspotContactsDisplayCategory, HubspotCompaniesDisplayCategory, HubspotDealsDisplayCategory,
		SalesforceUsersDisplayCategory, SalesforceAccountsDisplayCategory, SalesforceOpportunitiesDisplayCategory, MarketoLeadsDisplayCategory, LeadSquaredLeadsDisplayCategory}
	ProfileQueryType    = 1
	DerivedQueryType    = 2
	EventBasedQueryType = 3
)

var customMetricGroupNameByObjectType = map[string]string{
	GROUP_NAME_HUBSPOT_COMPANY:        HubspotCompaniesDisplayCategory,
	GROUP_NAME_HUBSPOT_DEAL:           HubspotDealsDisplayCategory,
	GROUP_NAME_SALESFORCE_OPPORTUNITY: SalesforceOpportunitiesDisplayCategory,
	GROUP_NAME_SALESFORCE_ACCOUNT:     SalesforceAccountsDisplayCategory,
}

type CustomMetric struct {
	ProjectID       int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID              string          `gorm:"primary_key:true;type:varchar(255)" json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	TypeOfQuery     int             `json:"type_of_query"`
	Transformations *postgres.Jsonb `json:"transformations"`
	ObjectType      string          `json:"obj_ty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

var KpiDerivedMetricsConfig = map[string]interface{}{
	"category":         Derived,
	"display_category": DerivedMetrics,
}

func (customMetric *CustomMetric) GetKPIConfig() map[string]string {
	currentMetric := make(map[string]string)
	currentMetric["name"] = customMetric.Name
	currentMetric["display_name"] = customMetric.Name
	currentMetric["type"] = ""
	return currentMetric
}

func GetKPIConfig(customMetrics []CustomMetric) []map[string]string {
	rCustomMetrics := make([]map[string]string, 0)

	for _, customMetric := range customMetrics {
		rCustomMetrics = append(rCustomMetrics, customMetric.GetKPIConfig())
	}
	return rCustomMetrics
}

func (customMetric *CustomMetric) IsValid() (bool, string) {
	if customMetric.TypeOfQuery == ProfileQueryType || customMetric.TypeOfQuery == EventBasedQueryType {

		var customMetricTransformation CustomMetricTransformation
		err := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &customMetricTransformation)
		if err != nil {
			return false, "Error during decode of custom metrics transformations - custom_metrics handler."
		}
		if !customMetricTransformation.IsValid(customMetric.TypeOfQuery) {
			return false, "Error with values passed in transformations - custom_metrics handler."
		}
		return true, ""

	} else if customMetric.TypeOfQuery == DerivedQueryType {

		var derivedMetricTransformation KPIQueryGroup
		err := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &derivedMetricTransformation)
		if err != nil {
			return false, "Error during decode of derived metrics transformations - custom_metrics handler."
		}
		if strings.Contains(derivedMetricTransformation.Formula, " ") {
			return false, "No empty space allowed in formula field"
		}
		isValidDerivedKPI, errMsg := derivedMetricTransformation.IsValidDerivedKPI()
		if !isValidDerivedKPI {
			return false, errMsg
		}
		return true, ""
	} else {
		return false, "Wrong query type is forwarded"
	}

}

func (customMetric *CustomMetric) SetDefaultGroupIfRequired() {
	if customMetric.TypeOfQuery == DerivedQueryType {
		var derivedMetricTransformation KPIQueryGroup
		U.DecodePostgresJsonbToStructType(customMetric.Transformations, &derivedMetricTransformation)
		previousDisplayCategory := ""
		for _, kpiQuery := range derivedMetricTransformation.Queries {
			if previousDisplayCategory == "" {
				previousDisplayCategory = kpiQuery.DisplayCategory
			} else if kpiQuery.DisplayCategory != previousDisplayCategory {
				previousDisplayCategory = OthersDisplayCategory
				break
			}
		}
		customMetric.ObjectType = previousDisplayCategory
	}
}

type CustomMetricTransformation struct {
	AggregateFunction     string      `json:"agFn"`
	AggregateProperty     string      `json:"agPr"`
	AggregatePropertyType string      `json:"agPrTy"`
	Filters               []KPIFilter `json:"fil"`
	DateField             string      `json:"daFie"`
	EventName             string      `json:"evNm"`
	Entity                string      `json:"en"`
}

func (c *CustomMetricTransformation) ContainsNameInInternalTransformation(input string) bool {
	return false
}

func (transformation *CustomMetricTransformation) ValidateFilterAndGroupBy() bool {
	return true
}

// Check if filter is being passed with objectType in create Custom metric.
func (transformation *CustomMetricTransformation) IsValid(queryType int) bool {

	if queryType == ProfileQueryType {
		if !U.ContainsStringInArray(CustomMetricProfilesAggregateFunctions, transformation.AggregateFunction) || strings.Contains(transformation.AggregateProperty, " ") {
			return false
		}
	} else if queryType == EventBasedQueryType {
		if !U.ContainsStringInArray(CustomEventsAggregateFunctions, transformation.AggregateFunction) || strings.Contains(transformation.AggregateProperty, " ") || 
			transformation.EventName == "" || ((transformation.AggregateFunction == SumAggregateFunction || transformation.AggregateFunction == AverageAggregateFunction) && transformation.Entity == "") {
			return false
		}
	} else {
		// invalid query type
		return false
	}

	for _, filter := range transformation.Filters {
		if !filter.IsValid() {
			return false
		}
	}

	return true
}

// invalid name - duplicate and blacklist.
func ValidateCustomMetric(customMetric CustomMetric) (string, bool) {
	if customMetric.ProjectID == 0 {
		return "Invalid project ID for custom metric", false
	}

	if customMetric.Name == "" {
		return "Invalid Name for custom metric", false
	}
	return "", true
}

func GetGroupNameByMetricObjectType(objectType string) string {
	for groupName, metricObjectType := range customMetricGroupNameByObjectType {
		if metricObjectType == objectType {
			return groupName
		}
	}
	return ""
}

// TODO handle errors
func DecodeCustomMetricsTransformation(customMetric CustomMetric) CustomKPI {
	if customMetric.TypeOfQuery == ProfileQueryType {
		var customMetricTransformation CustomMetricTransformation
		_ = U.DecodePostgresJsonbToStructType(customMetric.Transformations, &customMetricTransformation)
		return &customMetricTransformation
	} else if customMetric.TypeOfQuery == DerivedQueryType {
		var derivedMetricTransformation KPIQueryGroup
		_ = U.DecodePostgresJsonbToStructType(customMetric.Transformations, &derivedMetricTransformation)
		return &derivedMetricTransformation
	}
	return nil
}
