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

type CustomMetricConfigV1 struct {
	ObjectType             string              `json:"obj_ty"`
	SectionDisplayCategory string              `json:"section_display_category"`
	TypeOfQuery            int                 `json:"type_of_query"`
	AggregateFunctions     []string            `json:"agFn"`
	Properties             []map[string]string `json:"properties"`
	TypeOfQueryDisplayName string              `json:"type_of_query_display_name"`
	MetricTypes            []string            `json:"me_ty"`
}

const (
	SumAggregateFunction           = "sum"
	UniqueAggregateFunction        = "unique"
	AverageAggregateFunction       = "average"
	CountAggregateFunction         = "count"
	Derived                        = "derived"
	DerivedMetrics                 = "derived_metrics"
	CustomEventType                = "custom_events"
	ProfileQueryTypeDisplayName    = "Profile Based"
	DerivedQueryTypeDisplayName    = "Derived"
	EventBasedQueryTypeDisplayName = "Event Based"
	DateTypeDiffMetricType         = "date_type_diff_metric"
)

const (
	HubspotContacts                  = "Hubspot Contacts"
	HubspotMQLDateEntered            = "HubSpot MQLs (Date Entered Stage)"
	HubspotSQLDateEntered            = "HubSpot SQLs (Date Entered Stage)"
	HubspotContactsCreatedDate       = "HubSpot MQLs (Contact Create Date)"
	HubspotContactSQLCreatedDate     = "HubSpot SQLs (Contact Create Date)"
	HubspotOppCreatedDate            = "HubSpot Opps (Contact Create Date)"
	HubspotCustomersCreatedDate      = "HubSpot Customers (Contact Create Date)"
	HubspotMQLMqlDate                = "HubSpot MQLs (MQL Date)"
	HubspotSQLSqlDate                = "HubSpot SQLs (SQL Date)"
	HubspotOppsOppDate               = "HubSpot Opps (Opp Date)"
	HubspotCustomersCustomerDate     = "HubSpot Customers (Customer Date)"
	HubspotDeals                     = "Deals"
	HubspotPipeline                  = "Pipeline"
	HubspotRevenue                   = "Revenue"
	HubspotAvgDealSize               = "Avg Deal size"
	HubspotClosedWonDeals            = "Closed won Deals"
	HubspotClosedWonDealsCreatedDate = "Closed won Deals (Create Date)"
	HubspotMQLToSQL                  = "Hubspot MQLs to SQLs (Contact Create Date)"
	HubspotSQLToOpp                  = "Hubspot SQLs to Opps (Contact Create Date)"
	HubspotOppsToCustomer            = "Hubspot Opps to Customer (Contact Create Date)"
	HubspotClosedRate                = "Hubspot Closed rate (%)"
	HubspotAvgSalesCycleLength       = "Hubspot Avg Sales Cycle Length"

	SalesforceLeads                   = "Salesforce Leads"
	SalesforceSQLDateEntered          = "Salesforce SQLs (Date Entered Stage)"
	SalesforceOpportunities           = "Salesforce Opportunities"
	SalesforcePipeline                = "Salesforce Pipeline"
	SalesforceRevenue                 = "Salesforce Revenue"
	SalesforceAvgDealSize             = "Salesforce Avg Deal size"
	SalesforceClosedWonOpportunities  = "Salesforce Closed won Opportunities"
	SalesforceClosedWonOppsCreateDate = "Salesforce Closed won Opportunities (Create Date)"
	SalesforceClosedRate              = "Salesforce Closed rate (%)"
	SalesforceAvgSalesCycleLength     = "Salesforce Avg Sales Cycle Length"

	CurrencyBasedMetric   = "currency"
	PercentageBasedMetric = "percentage"
	DurationBasedMetric   = "duration"
)

var (
	// Shouldnt have unique in dateTypeDiff. How do I present that.
	CustomMetricProfilesAggregateFunctions   = []string{SumAggregateFunction, UniqueAggregateFunction, AverageAggregateFunction}
	CustomEventsAggregateFunctions           = []string{SumAggregateFunction, UniqueAggregateFunction, AverageAggregateFunction, CountAggregateFunction}
	CustomKPIProfileSectionDisplayCategories = []string{HubspotContactsDisplayCategory, HubspotCompaniesDisplayCategory, HubspotDealsDisplayCategory,
		SalesforceUsersDisplayCategory, SalesforceAccountsDisplayCategory, SalesforceOpportunitiesDisplayCategory, MarketoLeadsDisplayCategory, LeadSquaredLeadsDisplayCategory}
	CustomKPIProfilesMetricTypes = []string{DateTypeDiffMetricType}
	MapOfCustomMetricTypeToOp    = map[string]string{DateTypeDiffMetricType: "-"}
	ProfileQueryType             = 1
	DerivedQueryType             = 2
	EventBasedQueryType          = 3
)

var customMetricGroupNameBySectionDisplayCategory = map[string]string{
	GROUP_NAME_HUBSPOT_COMPANY:        HubspotCompaniesDisplayCategory,
	GROUP_NAME_HUBSPOT_DEAL:           HubspotDealsDisplayCategory,
	GROUP_NAME_SALESFORCE_OPPORTUNITY: SalesforceOpportunitiesDisplayCategory,
	GROUP_NAME_SALESFORCE_ACCOUNT:     SalesforceAccountsDisplayCategory,
}

type CustomMetric struct {
	ProjectID              int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID                     string          `gorm:"primary_key:true;type:varchar(255)" json:"id"`
	Name                   string          `json:"name"`
	Description            string          `json:"description"`
	TypeOfQuery            int             `json:"type_of_query"`
	MetricType             string          `json:"metric_type"` // represents a time difference based kpi ...
	Transformations        *postgres.Jsonb `json:"transformations"`
	ObjectType             string          `json:"obj_ty"`                             // Previously used as KPI Display Category for the metric
	SectionDisplayCategory string          `gorm:"-"  json:"section_display_category"` // To be used as KPI Display Category for the metric
	DisplayResultAs        string          `json:"display_result_as"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
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
	if customMetric.TypeOfQuery == ProfileQueryType || customMetric.TypeOfQuery == DerivedQueryType {
		currentMetric["type"] = customMetric.DisplayResultAs
	}
	return currentMetric
}

// This is used to fetch event name from transformation only. This is applicable for custom KPI - events.
func (customMetric *CustomMetric) GetEventName() (string, string) {
	if customMetric.TypeOfQuery == EventBasedQueryType {

		var customMetricTransformation CustomMetricTransformation
		err := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &customMetricTransformation)
		if err != nil {
			return "", "Error during decode of custom metrics transformations - custom_metrics handler."
		}
		return customMetricTransformation.EventName, ""

	} else {
		return "", "Wrong query type is forwarded"
	}
}

func (customMetric *CustomMetric) IsValid() (bool, string) {
	if customMetric.TypeOfQuery == ProfileQueryType || customMetric.TypeOfQuery == EventBasedQueryType {

		var customMetricTransformation CustomMetricTransformation
		err := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &customMetricTransformation)
		if err != nil {
			return false, "Error during decode of custom metrics transformations - custom_metrics handler."
		}
		if !customMetricTransformation.IsValid(customMetric.TypeOfQuery, customMetric.MetricType) {
			return false, "Error with values passed in transformations - custom_metrics handler."
		}

		if strings.Contains(customMetricTransformation.DateField, " ") {
			return false, "Error in date field - Contains space "
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
		if customMetric.DisplayResultAs != MetricsPercentageType && customMetric.DisplayResultAs != "" {
			return false, "Invalid metric type - custom_metrics handler."
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
		customMetric.SectionDisplayCategory = previousDisplayCategory
	}
}

func (customMetric *CustomMetric) SetDisplayResultAsIfRequired() {
	if customMetric.TypeOfQuery == ProfileQueryType {
		var customMetricTransformation CustomMetricTransformation
		U.DecodePostgresJsonbToStructType(customMetric.Transformations, &customMetricTransformation)

		if customMetric.MetricType == DateTypeDiffMetricType {
			customMetric.DisplayResultAs = MetricsDateType
		}
	}
}

func GetKPIConfig(customMetrics []CustomMetric) []map[string]string {
	rCustomMetrics := make([]map[string]string, 0)

	for _, customMetric := range customMetrics {
		rCustomMetrics = append(rCustomMetrics, customMetric.GetKPIConfig())
	}
	return rCustomMetrics
}

type CustomMetricTransformation struct {
	AggregateFunction      string      `json:"agFn"`
	AggregateProperty      string      `json:"agPr"`
	AggregateProperty2     string      `json:"agPr2"` // Currently used in dateType based metrics.
	AggregatePropertyType  string      `json:"agPrTy"`
	AggregatePropertyType2 string      `json:"agPrTy2"`
	Filters                []KPIFilter `json:"fil"`
	DateField              string      `json:"daFie"`
	EventName              string      `json:"evNm"`
	Entity                 string      `json:"en"`
}

func (c *CustomMetricTransformation) ContainsNameInInternalTransformation(input string) bool {
	return false
}

func (transformation *CustomMetricTransformation) ValidateFilterAndGroupBy() bool {
	return true
}

// Check if filter is being passed with objectType in create Custom metric.
func (transform *CustomMetricTransformation) IsValid(queryType int, metricType string) bool {
	if strings.Contains(transform.AggregateProperty, " ") {
		return false
	}

	for _, transformation := range transform.Filters {
		if strings.Contains(transformation.PropertyName, " ") {
			return false
		}
	}

	if queryType == ProfileQueryType {
		if metricType == "" {
			if !U.ContainsStringInArray(CustomMetricProfilesAggregateFunctions, transform.AggregateFunction) {
				return false
			}

			if U.ContainsStringInArray([]string{SumAggregateFunction, AverageAggregateFunction}, transform.AggregateFunction) {
				if transform.AggregateProperty == "" {
					return false
				}
				if transform.AggregatePropertyType == "" {
					return false
				}
				if transform.AggregatePropertyType != U.PropertyTypeNumerical {
					return false
				}
			}
		} else if U.ContainsStringInArray(CustomKPIProfilesMetricTypes, metricType) {
			if !U.ContainsStringInArray([]string{SumAggregateFunction, AverageAggregateFunction}, transform.AggregateFunction) {
				return false
			}

			if transform.AggregateProperty == "" || transform.AggregateProperty2 == "" ||
				transform.AggregatePropertyType == "" || transform.AggregatePropertyType2 == "" {
				return false
			}
			if metricType == DateTypeDiffMetricType {
				if transform.AggregatePropertyType != U.PropertyTypeDateTime &&
					transform.AggregatePropertyType2 != U.PropertyTypeDateTime {
					return false
				}
			}
		} else {
			return false
		}

	} else if queryType == EventBasedQueryType {
		if !U.ContainsStringInArray(CustomEventsAggregateFunctions, transform.AggregateFunction) || transform.EventName == "" ||
			((transform.AggregateFunction == SumAggregateFunction ||
				transform.AggregateFunction == AverageAggregateFunction) && transform.Entity == "") {
			return false
		}
	}

	for _, filter := range transform.Filters {
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

func GetGroupNameByMetricSectionDisplayCategory(sectionDisplayCategory string) string {
	for groupName, metricSectionDisplayCategory := range customMetricGroupNameBySectionDisplayCategory {
		if metricSectionDisplayCategory == sectionDisplayCategory {
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
