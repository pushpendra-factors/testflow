package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type CustomMetricConfig struct {
	AggregateFunctions      []string                              `json:"agFn"`
	ObjectTypeAndProperties []CustomMetricObjectTypeAndProperties `json:"objTyAndProp"`
}

type CustomMetricObjectTypeAndProperties struct {
	ObjectType string              `json:"objTy"`
	Properties []map[string]string `json:"properties"`
}

const (
	SumAggregateFunction      = "sum"
	UniqueAggregationFunction = "unique"
)

// var CustomMetricAggregateFunctions = []string{Count, SumAggregateFunction, UniqueAggregationFunction}
var CustomMetricAggregateFunctions = []string{UniqueAggregationFunction}
var CustomMetricObjectTypeNames = []string{HubspotContactsDisplayCategory, SalesforceUsersDisplayCategory}
var ProfileQueryType = 1

type CustomMetric struct {
	ProjectID       uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID              string          `gorm:"primary_key:true;type:varchar(255)" json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	TypeOfQuery     int             `json:"type_of_query"`
	Transformations *postgres.Jsonb `json:"transformations"`
	ObjectType      string          `json:"objTy"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type CustomMetricTransformation struct {
	AggregateFunction     string      `json:"agFn"`
	AggregateProperty     string      `json:"agPr"`
	AggregatePropertyType string      `json:"agPrTy"`
	Filters               []KPIFilter `json:"fil"`
	DateField             string      `json:"daFie"`
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
