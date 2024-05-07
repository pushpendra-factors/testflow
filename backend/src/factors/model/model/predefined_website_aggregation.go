package model

import (
	U "factors/util"
)

const (

	// Widget names
	PredefWidWebsiteOverview = "website_traffic_overview"
	PredefWidUtmParams       = "web_traffic_by_UTM_parameters"
	PredefWidpageUrl         = "website_sessions_by_page_URL"
	PredefWidPageView        = "top_pages_visited_by_page_URL"
	PredefWidGeography       = "website_sessions_by_geography"
	PredefWidTechnographics  = "website_sessions_by_browser"
	PredefWidFirmographics   = "website_sessions_by_company_firmographics"

	// Widget Display Names
	PredefWidDispWebsiteOverview = "Website traffic overview"
	PredefWidDispUtmParams       = "Web traffic by UTM parameters"
	PredefWidDispPageUrl         = "Website sessions by page URL"
	PredefWidDispPageView        = "Top pages visited by page URL"
	PredefWidDispGeography       = "Website sessions by geography"
	PredefWidDispTechnographics  = "Website sessions by browser"
	PredefWidDispFirmographics   = "Website sessions by company firmographics"

	PredefEventTypeSession   = "session"
	PredefEventTypePageViews = "page_view"

	// Metrics
	PredefTotalSessions      = "total_sessions"
	PredefTotalPageViews     = "total_page_views"
	PredefBounceRate         = "bounce_rate"
	PredefAvgSessionDuration = "avg_session_duration"

	// Metrics Display
	PredefDispTotalSessions      = "Total Sessions"
	PredefDispTotalPageViews     = "Total Page Views"
	PredefDispBounceRate         = "Bounce Rate"
	PredefDispAvgSessionDuration = "Avg Session Duration"

	// Internal Metrics
	PredefSpendTime      = "spent_time"
	PredefCountOfRecords = "count_of_records"

	// GroupByProperties or Global Filters for page
	PredefPropSource              = "source"
	PredefPropMedium              = "medium"
	PredefPropChannel             = "channel"
	PredefPropCampaign            = "campaign"
	PredefPropLandingPageURl      = "landing_page_url"
	PredefPropReferrerUrl         = "referrer_url"
	PredefPropExitPage            = "exit_page"
	PredefPropTopPage             = "top_pages"
	PredefPropCountry             = "country"
	PredefPropCity                = "city"
	PredefPropBrowser             = "browser"
	PredefPropBrowserVersion      = "browser_version"
	PredefPropOs                  = "os"
	PredefPropOsVersion           = "os_version"
	PredefPropDevice              = "device"
	PredefProp6SignalIndustry     = "6signal_industry"
	PredefProp6SignalEmpRange     = "6signal_employee_range"
	PredefProp6SignalRevenueRange = "6signal_revenue_range"
	PredepPropTimestampAtDay      = "timestamp_at_day"

	// GroupByProperties or Global Filters for page Display
	PredefPropDispSource              = "Source"
	PredefPropDispMedium              = "Medium"
	PredefPropDispCampaign            = "Campaign"
	PredefPropDispChannel             = "Channel"
	PredefPropDispLandingPageURl      = "Landing Page URL"
	PredefPropDispReferrerUrl         = "Referrer URL"
	PredefPropDispExitPage            = "Exit page"
	PredefPropDispTopPage             = "Top pages"
	PredefPropDispCountry             = "Country"
	PredefPropDispRegion              = "Region"
	PredefPropDispCity                = "City"
	PredefPropDispBrowser             = "Browser"
	PredefPropDispBrowsVersion        = "Browser version"
	PredefPropDispOs                  = "Os"
	PredefPropDispOsVersion           = "Os version"
	PredefPropDispDevice              = "Device"
	PredefPropDisp6SignalIndustry     = "Company industry"
	PredefPropDisp6SignalEmpRange     = "Company employee range"
	PredefPropDisp6SignalRevenueRange = "Company revenue range"

	// Internal Group by
	PredefPropEventName     = "event_name"
	PredefPropLatestPageUrl = "latest_page_url"
)

// During website_aggregation, we fetch the properties from events table. Here source represent the details of properties in events table.
// Predefined website aggregation is mostly for session based properties. Hence the source is defined against session data.
var predefinedWebsiteAggregationProperties = []PredefinedDashboardProperty{
	{Name: PredefPropSource, DisplayName: PredefPropDispSource, DataType: "categorical", SourceEventName: U.EVENT_NAME_SESSION, SourceEntity: EventEntity, SourceProperty: U.EP_SOURCE},
	{Name: PredefPropMedium, DisplayName: PredefPropDispMedium, DataType: "categorical", SourceEventName: U.EVENT_NAME_SESSION, SourceEntity: EventEntity, SourceProperty: U.EP_MEDIUM},
	{Name: PredefPropCampaign, DisplayName: PredefPropDispCampaign, DataType: "categorical", SourceEventName: U.EVENT_NAME_SESSION, SourceEntity: EventEntity, SourceProperty: U.EP_CAMPAIGN},

	{Name: PredefPropLandingPageURl, DisplayName: PredefPropDispLandingPageURl, DataType: "categorical", SourceEventName: U.EVENT_NAME_SESSION, SourceEntity: EventEntity, SourceProperty: U.SP_INITIAL_PAGE_URL},
	{Name: PredefPropReferrerUrl, DisplayName: PredefPropDispReferrerUrl, DataType: "categorical", SourceEventName: U.EVENT_NAME_SESSION, SourceEntity: EventEntity, SourceProperty: U.SP_INITIAL_REFERRER_URL},

	{Name: PredefPropCountry, DisplayName: PredefPropDispCountry, DataType: "categorical", SourceEventName: U.EVENT_NAME_SESSION, SourceEntity: EventEntity, SourceProperty: U.UP_COUNTRY},
	{Name: PredefPropCity, DisplayName: PredefPropDispCity, DataType: "categorical", SourceEventName: U.EVENT_NAME_SESSION, SourceEntity: EventEntity, SourceProperty: U.UP_CITY},

	{Name: PredefPropBrowser, DisplayName: PredefPropDispBrowser, DataType: "categorical", SourceEventName: U.EVENT_NAME_SESSION, SourceEntity: EventEntity, SourceProperty: U.UP_BROWSER},

	{Name: PredefPropOs, DisplayName: PredefPropDispOs, DataType: "categorical", SourceEventName: U.EVENT_NAME_SESSION, SourceEntity: EventEntity, SourceProperty: U.UP_OS},

	{Name: PredefProp6SignalIndustry, DisplayName: PredefPropDisp6SignalIndustry, DataType: "categorical", SourceEventName: "", SourceEntity: UserEntity, SourceProperty: U.SIX_SIGNAL_INDUSTRY},
	{Name: PredefProp6SignalEmpRange, DisplayName: PredefPropDisp6SignalEmpRange, DataType: "categorical", SourceEventName: "", SourceEntity: UserEntity, SourceProperty: U.SIX_SIGNAL_EMPLOYEE_RANGE},
	{Name: PredefProp6SignalRevenueRange, DisplayName: PredefPropDisp6SignalRevenueRange, DataType: "categorical", SourceEventName: "", SourceEntity: UserEntity, SourceProperty: U.SIX_SIGNAL_REVENUE_RANGE},
}

var predefinedWebsiteAggregationWidgets = []PredefinedWidget{
	{
		InternalID:  1,
		Name:        PredefWidWebsiteOverview,
		DisplayName: PredefWidDispWebsiteOverview,
		Metrics: []PredefinedMetric{
			{Name: PredefTotalSessions, DisplayName: PredefDispTotalSessions, InternalEventType: PredefEventTypeSession},
			{Name: PredefTotalPageViews, DisplayName: PredefDispTotalPageViews, InternalEventType: PredefEventTypePageViews},
			// {Name: PredefBounceRate, DisplayName: PredefDispBounceRate, InternalEventType: PredefEventTypeSession},
			{Name: PredefAvgSessionDuration, DisplayName: PredefDispAvgSessionDuration, InternalEventType: PredefEventTypeSession, Type: MetricsDateType},
		},
		GroupBy: []PredefinedGroupBy{},
		Setting: ChartSetting{Type: ChartTypeLineChart, Presentation: PresentationTypeChart},
	},
	{
		InternalID:  2,
		Name:        PredefWidUtmParams,
		DisplayName: PredefWidDispUtmParams,
		Metrics: []PredefinedMetric{
			{Name: PredefTotalSessions, DisplayName: PredefDispTotalSessions, InternalEventType: PredefEventTypeSession},
		},
		GroupBy: []PredefinedGroupBy{
			{Name: PredefPropSource, DisplayName: PredefPropDispSource},
			{Name: PredefPropMedium, DisplayName: PredefPropDispMedium},
			{Name: PredefPropCampaign, DisplayName: PredefPropDispCampaign},
			{Name: PredefPropChannel, DisplayName: PredefPropDispChannel},
		},
		Setting: ChartSetting{Type: ChartTypeBarChart, Presentation: PresentationTypeChart},
	},
	{
		InternalID:  3,
		Name:        PredefWidpageUrl,
		DisplayName: PredefWidDispPageUrl,
		Metrics: []PredefinedMetric{
			{Name: PredefTotalSessions, DisplayName: PredefDispTotalSessions, InternalEventType: PredefEventTypeSession},
			// {Name: PredefBounceRate, DisplayName: PredefDispBounceRate, InternalEventType: PredefEventTypeSession},
			{Name: PredefAvgSessionDuration, DisplayName: PredefDispAvgSessionDuration, InternalEventType: PredefEventTypeSession, Type: MetricsDateType},
		},
		GroupBy: []PredefinedGroupBy{
			{Name: PredefPropLandingPageURl, DisplayName: PredefPropDispLandingPageURl},
			{Name: PredefPropReferrerUrl, DisplayName: PredefPropDispReferrerUrl},
			{Name: PredefPropExitPage, DisplayName: PredefPropDispExitPage},
		},
		Setting: ChartSetting{Type: ChartTypeHorizontalBarChart, Presentation: PresentationTypeChart},
	},
	{
		InternalID:  4,
		Name:        PredefWidPageView,
		DisplayName: PredefWidDispPageView,
		Metrics: []PredefinedMetric{
			{Name: PredefTotalPageViews, DisplayName: PredefDispTotalPageViews, InternalEventType: PredefEventTypePageViews},
		},
		GroupBy: []PredefinedGroupBy{
			{Name: PredefPropTopPage, DisplayName: PredefPropDispTopPage},
		},
		Setting: ChartSetting{Type: ChartTypeHorizontalBarChart, Presentation: PresentationTypeChart},
	},
	{
		InternalID:  5,
		Name:        PredefWidGeography,
		DisplayName: PredefWidDispGeography,
		Metrics: []PredefinedMetric{
			{Name: PredefTotalSessions, DisplayName: PredefDispTotalSessions, InternalEventType: PredefEventTypeSession},
			// {Name: PredefBounceRate, DisplayName: PredefDispBounceRate, InternalEventType: PredefEventTypeSession},
			{Name: PredefAvgSessionDuration, DisplayName: PredefDispAvgSessionDuration, InternalEventType: PredefEventTypeSession, Type: MetricsDateType},
		},
		GroupBy: []PredefinedGroupBy{
			{Name: PredefPropCountry, DisplayName: PredefPropDispCountry},
			{Name: PredefPropCity, DisplayName: PredefPropDispCity},
		},
		Setting: ChartSetting{Type: ChartTypeBarChart, Presentation: PresentationTypeTable},
	},
	{
		InternalID:  6,
		Name:        PredefWidTechnographics,
		DisplayName: PredefWidDispTechnographics,
		Metrics: []PredefinedMetric{
			{Name: PredefTotalSessions, DisplayName: PredefDispTotalSessions, InternalEventType: PredefEventTypeSession},
			// {Name: PredefBounceRate, DisplayName: PredefDispBounceRate, InternalEventType: PredefEventTypeSession},
			{Name: PredefAvgSessionDuration, DisplayName: PredefDispAvgSessionDuration, InternalEventType: PredefEventTypeSession, Type: MetricsDateType},
		},
		GroupBy: []PredefinedGroupBy{
			{Name: PredefPropBrowser, DisplayName: PredefPropDispBrowser},
			{Name: PredefPropOs, DisplayName: PredefPropDispOs},
			{Name: PredefPropDevice, DisplayName: PredefPropDispDevice},
		},
		Setting: ChartSetting{Type: ChartTypeBarChart, Presentation: PresentationTypeTable},
	},
	{
		InternalID:  7,
		Name:        PredefWidFirmographics,
		DisplayName: PredefWidDispFirmographics,
		Metrics: []PredefinedMetric{
			{Name: PredefTotalSessions, DisplayName: PredefDispTotalSessions, InternalEventType: PredefEventTypeSession},
			// {Name: PredefBounceRate, DisplayName: PredefDispBounceRate, InternalEventType: PredefEventTypeSession},
			{Name: PredefAvgSessionDuration, DisplayName: PredefDispAvgSessionDuration, InternalEventType: PredefEventTypeSession},
		},
		GroupBy: []PredefinedGroupBy{
			{Name: PredefProp6SignalIndustry, DisplayName: PredefPropDisp6SignalIndustry},
			{Name: PredefProp6SignalEmpRange, DisplayName: PredefPropDisp6SignalEmpRange},
			{Name: PredefProp6SignalRevenueRange, DisplayName: PredefPropDisp6SignalRevenueRange},
		},
		Setting: ChartSetting{Type: ChartTypeBarChart, Presentation: PresentationTypeTable},
	},
}

var PredefinedDashboardUnitsPosition = map[string]map[string]int{
	"position": {"1": 1, "2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7},
	"size":     {"1": 1, "2": 1, "3": 0, "4": 0, "5": 0, "6": 0, "7": 1},
}

var MapOfPredefPropertyNameToProperties = map[string]PredefinedDashboardProperty{
	predefinedWebsiteAggregationProperties[0].Name:  predefinedWebsiteAggregationProperties[0],
	predefinedWebsiteAggregationProperties[1].Name:  predefinedWebsiteAggregationProperties[1],
	predefinedWebsiteAggregationProperties[2].Name:  predefinedWebsiteAggregationProperties[2],
	predefinedWebsiteAggregationProperties[3].Name:  predefinedWebsiteAggregationProperties[3],
	predefinedWebsiteAggregationProperties[4].Name:  predefinedWebsiteAggregationProperties[4],
	predefinedWebsiteAggregationProperties[5].Name:  predefinedWebsiteAggregationProperties[5],
	predefinedWebsiteAggregationProperties[6].Name:  predefinedWebsiteAggregationProperties[6],
	predefinedWebsiteAggregationProperties[7].Name:  predefinedWebsiteAggregationProperties[7],
	predefinedWebsiteAggregationProperties[8].Name:  predefinedWebsiteAggregationProperties[8],
	predefinedWebsiteAggregationProperties[9].Name:  predefinedWebsiteAggregationProperties[9],
	predefinedWebsiteAggregationProperties[10].Name: predefinedWebsiteAggregationProperties[10],
	predefinedWebsiteAggregationProperties[11].Name: predefinedWebsiteAggregationProperties[11],
}

var MapOfPredefDashboardIDToWidgets = map[int64]PredefinedWidget{
	predefinedWebsiteAggregationWidgets[0].InternalID: predefinedWebsiteAggregationWidgets[0],
	predefinedWebsiteAggregationWidgets[1].InternalID: predefinedWebsiteAggregationWidgets[1],
	predefinedWebsiteAggregationWidgets[2].InternalID: predefinedWebsiteAggregationWidgets[2],
	predefinedWebsiteAggregationWidgets[3].InternalID: predefinedWebsiteAggregationWidgets[3],
	predefinedWebsiteAggregationWidgets[4].InternalID: predefinedWebsiteAggregationWidgets[4],
	predefinedWebsiteAggregationWidgets[5].InternalID: predefinedWebsiteAggregationWidgets[5],
	predefinedWebsiteAggregationWidgets[6].InternalID: predefinedWebsiteAggregationWidgets[6],
}

type PredefWebsiteAggregationQueryGroup struct {
	Class   string                          `json:"cl"`
	Queries []PredefWebsiteAggregationQuery `json:"q_g"`
}

func (q PredefWebsiteAggregationQueryGroup) IsValid() (bool, string) {
	resValid := true
	for _, query := range q.Queries {
		isValid, errMsg := query.IsValid()
		if !isValid {
			return isValid, errMsg
		}
	}
	return resValid, ""
}

type PredefWebsiteAggregationQuery struct {
	Metrics           []PredefinedMetric `json:"me"`
	GroupBy           PredefinedGroupBy  `json:"g_by"`
	Filters           []PredefinedFilter `json:"fil"`
	GroupByTimestamp  string             `json:"gbt"`
	Timezone          string             `json:"tz"`
	From              int64              `json:"fr"`
	To                int64              `json:"to"`
	InternalEventType string             `json:"inter_e_type"`
	WidgetName        string             `json:"wid"`
	WidgetInternalID  int64              `json:"inter_id"`
}

// TODO forward the widget name. Add it as params.
// TODO add dataType categorical Check.
// TODO logicalOperator in filters to be valid.
func (q *PredefWebsiteAggregationQuery) IsValid() (bool, string) {

	if widgetConfig, exists := MapOfPredefDashboardIDToWidgets[q.WidgetInternalID]; exists {

		// if len(widgetConfig.Metrics) != len(q.Metrics) {
		// 	return false, "Number of Metrics sent for this widget are not matching"
		// }

		// Checking for metrics.
		set := make(map[string]bool)
		for _, metric := range widgetConfig.Metrics {
			set[metric.Name] = true
		}

		for _, metric := range q.Metrics {
			exists := set[metric.Name]
			if !exists {
				return false, "Input metrics sent for this widget are wrong"
			}
		}

		if len(widgetConfig.GroupBy) != 0 {
			// Checking for groupBy.
			groupExists := false
			for _, groupBy := range widgetConfig.GroupBy {
				if groupBy.Name == q.GroupBy.Name {
					groupExists = true
				}
			}
			if !groupExists {
				return false, "Invalid group by provided for this dashboard ID"
			}
		}
	} else {
		return false, "Invalid widget internal ID sent"
	}

	// TODO if we should move towards MapOfPredefinedDashboardToPropertyNameToProperties.
	for _, filter := range q.Filters {
		if _, exists := MapOfPredefPropertyNameToProperties[filter.PropertyName]; !exists {
			return false, "Invalid Filters provided for this dashboard ID"
		}
	}
	return true, ""
}

func (q *PredefWebsiteAggregationQuery) GetGroupByTimestamp() string {
	return q.GroupByTimestamp
}

type PredefWebsiteAggregationMetricTransform struct {
	Operation          string
	InternalProperty   string
	ExternalProperty   string
	ArithmeticOperator string
}

var PredefWebMetricToInternalTransformations = map[string][]PredefWebsiteAggregationMetricTransform{
	PredefTotalSessions: {
		{Operation: SumAggregateFunction, InternalProperty: PredefCountOfRecords, ExternalProperty: "", ArithmeticOperator: ""},
	},
	PredefTotalPageViews: {
		{Operation: SumAggregateFunction, InternalProperty: PredefCountOfRecords, ExternalProperty: "", ArithmeticOperator: ""},
	},
	PredefBounceRate: {
		{Operation: SumAggregateFunction, InternalProperty: "", ExternalProperty: "", ArithmeticOperator: ""},
	},
	PredefAvgSessionDuration: {
		{Operation: SumAggregateFunction, InternalProperty: PredefSpendTime, ExternalProperty: PredefSpendTime, ArithmeticOperator: "Division"},
		{Operation: SumAggregateFunction, InternalProperty: PredefCountOfRecords, ExternalProperty: PredefCountOfRecords, ArithmeticOperator: ""},
	},
}

var MapOfPredefWebsiteAggregGroupByExternalToInternal = map[string]string{
	PredefPropTopPage:  PredefPropEventName,
	PredefPropExitPage: PredefPropLatestPageUrl,
}

var MapOfPredefinedWebsiteAggregaFilterExternalToInternal = map[string]string{}

// Convert the object Format i.e.key-value to Array of field key-values that are required.
func convertToArrayPropertiesFormat(properties []PredefinedDashboardProperty) [][]string {

	resultantArray := make([][]string, 0)
	for _, property := range properties {
		currentyArray := []string{property.DisplayName, property.Name, property.DataType}
		resultantArray = append(resultantArray, currentyArray)
	}
	return resultantArray
}
