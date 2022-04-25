package model

import (
	U "factors/util"
	"strconv"
	"strings"
	"time"
)

const (
	// Common
	UniqueUsers    = "unique_users"
	EngagedUsers   = "engaged_users"
	EngagementRate = "engagement_rate"

	// Unique Metrics to website session
	TotalSessions          = "total_sessions"
	NewUsers               = "new_users"
	RepeatUsers            = "repeat_users"
	SessionsPerUser        = "sessions_per_user"
	EngagedSessions        = "engaged_sessions"
	EngagedSessionsPerUser = "engaged_sessions_per_user"
	TotalTimeOnSite        = "total_time_on_site"
	AvgSessionDuration     = "average_session_duration"
	AvgPageViewsPerSession = "average_page_views_per_session"
	AvgInitialPageLoadTime = "average_initial_page_load_time"
	BounceRate             = "bounce_rate"

	// Unique Metrics to Page views
	Entrances                = "entrances"
	Exits                    = "exits"
	PageViews                = "page_views"
	PageviewsPerUser         = "page_views_per_user"
	AvgPageLoadTime          = "average_page_load_time"
	AvgVerticalScrollPercent = "average_vertical_scroll_percentage"
	AvgTimeOnPage            = "average_time_on_page"
	EngagedPageViews         = "engaged_page_views"

	// Unique  Metrics to Form Submissions
	Count        = "count"
	CountPerUser = "count_per_user"

	// Common Metrics to Hubspot And Salesforce
	CountOfContactsCreated  = "count_of_contacts_created"
	CountOfContactsUpdated  = "count_of_contacts_updated"
	CountOfCompaniesCreated = "count_of_companies_created"
	CountOfCompaniesUpdated = "count_of_companies_updated"

	// Unique Metrics to hubspot
	CountOfAccountsCreated = "count_of_accounts_created"
	CountOfAccountsUpdated = "count_of_accounts_updated"
	CountOfDealsCreated    = "count_of_deals_created"
	CountOfDealsUpdated    = "count_of_deals_updated"

	// Unique Metrics to Salesforce
	CountOfLeadsCreated         = "count_of_leads_created"
	CountOfLeadsUpdated         = "count_of_leads_updated"
	CountOfOpportunitiesCreated = "count_of_opportunities_created"
	CountOfOpportunitiesUpdated = "count_of_opportunities_updated"

	EventCategory   = "events"
	ChannelCategory = "channels"
	ProfileCategory = "profiles"

	EventEntity = "event"
	UserEntity  = "user"
)

// Each Property could belong to event or user entity based on type of event we consider. Eg - session has os property in event. Mostly used in EventsCategory.
var MapOfKPIPropertyNameToData = map[string]map[string]map[string]string{
	// Session only -  Event Properties.
	U.EP_SOURCE:                 {EventEntity: {"name": U.EP_SOURCE, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.EP_SOURCE], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},
	U.EP_MEDIUM:                 {EventEntity: {"name": U.EP_MEDIUM, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.EP_MEDIUM], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},
	U.EP_CAMPAIGN:               {EventEntity: {"name": U.EP_CAMPAIGN, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.EP_CAMPAIGN], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},
	U.EP_ADGROUP:                {EventEntity: {"name": U.EP_ADGROUP, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.EP_ADGROUP], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},
	U.EP_KEYWORD:                {EventEntity: {"name": U.EP_KEYWORD, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.EP_KEYWORD], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},
	U.EP_CHANNEL:                {EventEntity: {"name": U.EP_CHANNEL, "display_name": U.STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES[U.EP_CHANNEL], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},
	U.EP_CONTENT:                {EventEntity: {"name": U.EP_CONTENT, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.EP_CONTENT], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},
	U.SP_INITIAL_PAGE_URL:       {EventEntity: {"name": U.SP_INITIAL_PAGE_URL, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.SP_INITIAL_PAGE_URL], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},
	U.SP_LATEST_PAGE_URL:        {EventEntity: {"name": U.SP_LATEST_PAGE_URL, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.SP_LATEST_PAGE_URL], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},
	U.EP_PAGE_COUNT:             {EventEntity: {"name": U.EP_PAGE_COUNT, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.EP_PAGE_COUNT], "data_type": U.PropertyTypeNumerical, "entity": EventEntity}},
	U.SP_SPENT_TIME:             {EventEntity: {"name": U.SP_SPENT_TIME, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.SP_SPENT_TIME], "data_type": U.PropertyTypeNumerical, "entity": EventEntity}},
	U.SP_INITIAL_PAGE_LOAD_TIME: {EventEntity: {"name": U.SP_INITIAL_PAGE_LOAD_TIME, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.SP_INITIAL_PAGE_LOAD_TIME], "data_type": U.PropertyTypeNumerical, "entity": EventEntity}},
	U.UP_INITIAL_REFERRER_URL:   {EventEntity: {"name": U.UP_INITIAL_REFERRER_URL, "display_name": U.STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES[U.EP_REFERRER_URL], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},

	// Session and Generic Event - Event Properties
	U.EP_TIMESTAMP: {EventEntity: {"name": U.EP_TIMESTAMP, "display_name": U.STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES[U.EP_TIMESTAMP], "data_type": U.PropertyTypeDateTime, "entity": EventEntity}},

	// Generic Event - Event Properties.
	U.EP_REFERRER_URL:        {EventEntity: {"name": U.EP_REFERRER_URL, "display_name": U.STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES[U.EP_REFERRER_URL], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},
	U.EP_PAGE_URL:            {EventEntity: {"name": U.EP_PAGE_URL, "display_name": U.STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES[U.EP_PAGE_URL], "data_type": U.PropertyTypeCategorical, "entity": EventEntity}},
	U.EP_PAGE_LOAD_TIME:      {EventEntity: {"name": U.EP_PAGE_LOAD_TIME, "display_name": U.STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES[U.EP_PAGE_LOAD_TIME], "data_type": U.PropertyTypeNumerical, "entity": EventEntity}},
	U.EP_PAGE_SPENT_TIME:     {EventEntity: {"name": U.EP_PAGE_SPENT_TIME, "display_name": U.STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES[U.EP_PAGE_SPENT_TIME], "data_type": U.PropertyTypeNumerical, "entity": EventEntity}},
	U.EP_PAGE_SCROLL_PERCENT: {EventEntity: {"name": U.EP_PAGE_SCROLL_PERCENT, "display_name": U.STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES[U.EP_PAGE_SCROLL_PERCENT], "data_type": U.PropertyTypeNumerical, "entity": EventEntity}},

	// Generic Event - User Properties.
	U.UP_DEVICE_TYPE:  {UserEntity: {"name": U.UP_DEVICE_TYPE, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_DEVICE_TYPE], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},
	U.UP_DEVICE_BRAND: {UserEntity: {"name": U.UP_DEVICE_BRAND, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_DEVICE_BRAND], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},
	U.UP_DEVICE_MODEL: {UserEntity: {"name": U.UP_DEVICE_MODEL, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_DEVICE_MODEL], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},
	U.UP_DEVICE_NAME:  {UserEntity: {"name": U.UP_DEVICE_NAME, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_DEVICE_NAME], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},
	U.UP_PLATFORM:     {UserEntity: {"name": U.UP_PLATFORM, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_PLATFORM], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},

	// Session and Generic Event - Properties which could be in event or user entity.
	U.UP_OS: {UserEntity: {"name": U.UP_OS, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_OS], "data_type": U.PropertyTypeCategorical, "entity": UserEntity},
		EventEntity: {"name": U.UP_OS, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.UP_OS], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},
	U.UP_OS_VERSION: {UserEntity: {"name": U.UP_OS_VERSION, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_OS_VERSION], "data_type": U.PropertyTypeCategorical, "entity": UserEntity},
		EventEntity: {"name": U.UP_OS_VERSION, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.UP_OS_VERSION], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},
	U.UP_BROWSER: {UserEntity: {"name": U.UP_BROWSER, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_BROWSER], "data_type": U.PropertyTypeCategorical, "entity": UserEntity},
		EventEntity: {"name": U.UP_BROWSER, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.UP_BROWSER], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},
	U.UP_BROWSER_VERSION: {UserEntity: {"name": U.UP_BROWSER_VERSION, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_BROWSER_VERSION], "data_type": U.PropertyTypeCategorical, "entity": UserEntity},
		EventEntity: {"name": U.UP_BROWSER_VERSION, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.UP_BROWSER_VERSION], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},
	U.UP_COUNTRY: {UserEntity: {"name": U.UP_COUNTRY, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_COUNTRY], "data_type": U.PropertyTypeCategorical, "entity": UserEntity},
		EventEntity: {"name": U.UP_COUNTRY, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.UP_COUNTRY], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},
	U.UP_REGION: {UserEntity: {"name": U.UP_REGION, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_REGION], "data_type": U.PropertyTypeCategorical, "entity": UserEntity},
		EventEntity: {"name": U.UP_REGION, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.UP_REGION], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},
	U.UP_CITY: {UserEntity: {"name": U.UP_CITY, "display_name": U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES[U.UP_CITY], "data_type": U.PropertyTypeCategorical, "entity": UserEntity},
		EventEntity: {"name": U.UP_CITY, "display_name": U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES[U.UP_CITY], "data_type": U.PropertyTypeCategorical, "entity": UserEntity}},
}

// Removed constants for hubspot and salesforce kpi metrics in PR - .
// 1 Represents agggregation equivalent to aggregateFunc(1) in sql. For eg - select count(1)
var TransformationOfKPIMetricsToEventAnalyticsQuery = map[string]map[string][]TransformQueryi{
	WebsiteSessionDisplayCategory: {
		TotalSessions: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
		UniqueUsers:   []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
		NewUsers: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
				Filters: []QueryProperty{{Entity: EventEntity, Type: U.PropertyTypeCategorical, Property: U.SP_IS_FIRST_SESSION, LogicalOp: "AND", Operator: EqualsOpStr, Value: "true"}},
			},
		},
		RepeatUsers: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
				Filters: []QueryProperty{{Entity: EventEntity, Type: U.PropertyTypeCategorical, Property: U.SP_IS_FIRST_SESSION, LogicalOp: "AND", Operator: NotEqualOpStr, Value: "true"}},
			},
		},
		SessionsPerUser: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: "Division"},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
		EngagedSessions: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
				Filters: []QueryProperty{
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.SP_SPENT_TIME, LogicalOp: "AND", Operator: GreaterThanOpStr, Value: "10"},
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.SP_PAGE_COUNT, LogicalOp: "OR", Operator: GreaterThanOpStr, Value: "2"},
				},
			},
		},
		EngagedUsers: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
				Filters: []QueryProperty{
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.SP_SPENT_TIME, LogicalOp: "AND", Operator: GreaterThanOpStr, Value: "10"},
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.SP_PAGE_COUNT, LogicalOp: "OR", Operator: GreaterThanOpStr, Value: "2"},
				},
			},
		},
		EngagedSessionsPerUser: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: "Division"},
				Filters: []QueryProperty{
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.SP_SPENT_TIME, LogicalOp: "AND", Operator: GreaterThanOpStr, Value: "10"},
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.SP_PAGE_COUNT, LogicalOp: "OR", Operator: GreaterThanOpStr, Value: "2"},
				},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
				Filters: []QueryProperty{
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.SP_SPENT_TIME, LogicalOp: "AND", Operator: GreaterThanOpStr, Value: "10"},
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.SP_PAGE_COUNT, LogicalOp: "OR", Operator: GreaterThanOpStr, Value: "2"},
				},
			},
		},
		TotalTimeOnSite: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "sum", Entity: EventEntity, Property: U.SP_SPENT_TIME, GroupByType: U.PropertyTypeNumerical, Operator: ""},
			},
		},
		AvgSessionDuration: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "sum", Entity: EventEntity, Property: U.SP_SPENT_TIME, GroupByType: U.PropertyTypeNumerical, Operator: "Division"},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
		AvgPageViewsPerSession: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "sum", Entity: EventEntity, Property: U.SP_PAGE_COUNT, GroupByType: U.PropertyTypeNumerical, Operator: "Division"},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
		AvgInitialPageLoadTime: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "sum", Entity: EventEntity, Property: U.SP_INITIAL_PAGE_LOAD_TIME, GroupByType: U.PropertyTypeNumerical, Operator: "Division"},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
		BounceRate: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: "Division"},
				Filters: []QueryProperty{
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.SP_PAGE_COUNT, LogicalOp: "AND", Operator: EqualsOpStr, Value: "1"},
				},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
		EngagementRate: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: "Division"},
				Filters: []QueryProperty{
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.SP_SPENT_TIME, LogicalOp: "AND", Operator: GreaterThanOpStr, Value: "10"},
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.SP_PAGE_COUNT, LogicalOp: "OR", Operator: GreaterThanOpStr, Value: "2"},
				},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
	},
	PageViewsDisplayCategory: {
		Entrances: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
				Filters: []QueryProperty{{Entity: EventEntity, Type: U.PropertyTypeCategorical, Property: U.SP_INITIAL_PAGE_URL, LogicalOp: "AND", Operator: EqualsOpStr, Value: "true"}},
			},
		},
		Exits: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
				Filters: []QueryProperty{{Entity: EventEntity, Type: U.PropertyTypeCategorical, Property: U.UP_LATEST_PAGE_URL, LogicalOp: "AND", Operator: EqualsOpStr, Value: "true"}},
			},
		},
		PageViews: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
		UniqueUsers: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
		PageviewsPerUser: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: "Division"},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
		EngagedPageViews: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
				Filters: []QueryProperty{
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.EP_PAGE_SPENT_TIME, LogicalOp: "AND", Operator: GreaterThanOpStr, Value: "10"},
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.EP_PAGE_SCROLL_PERCENT, LogicalOp: "OR", Operator: GreaterThanOpStr, Value: "50"},
				},
			},
		},
		EngagedUsers: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
				Filters: []QueryProperty{
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.EP_PAGE_SPENT_TIME, LogicalOp: "AND", Operator: GreaterThanOpStr, Value: "10"},
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.EP_PAGE_SCROLL_PERCENT, LogicalOp: "OR", Operator: GreaterThanOpStr, Value: "50"},
				},
			},
		},
		EngagementRate: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: "Division"},
				Filters: []QueryProperty{
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.EP_PAGE_SPENT_TIME, LogicalOp: "AND", Operator: GreaterThanOpStr, Value: "10"},
					{Entity: EventEntity, Type: U.PropertyTypeNumerical, Property: U.EP_PAGE_SCROLL_PERCENT, LogicalOp: "OR", Operator: GreaterThanOpStr, Value: "50"},
				},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
		AvgPageLoadTime: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "sum", Entity: EventEntity, Property: U.EP_PAGE_LOAD_TIME, GroupByType: U.PropertyTypeNumerical, Operator: "Division"},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
		AvgVerticalScrollPercent: {
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "sum", Entity: EventEntity, Property: U.EP_PAGE_SCROLL_PERCENT, GroupByType: U.PropertyTypeNumerical, Operator: "Division"},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
		AvgTimeOnPage: {
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "sum", Entity: EventEntity, Property: U.EP_PAGE_SPENT_TIME, GroupByType: U.PropertyTypeNumerical, Operator: "Division"},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
	},
	FormSubmissionsDisplayCategory: {
		Count:       []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
		UniqueUsers: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
		CountPerUser: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: "Division"},
			},
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
			},
		},
	},
}

func CheckIfPropertyIsPresentInStaticKPIPropertyList(inputProperty string) bool {
	_, exists := MapOfKPIPropertyNameToData[inputProperty]
	return exists
}

func GetDirectDerviableQueryPropsFromKPI(kpiQuery KPIQuery) Query {
	var query Query
	query.Class = "events"
	query.GroupByTimestamp = kpiQuery.GroupByTimestamp
	query.EventsCondition = "each_given_event"
	query.Timezone = kpiQuery.Timezone
	query.From = kpiQuery.From
	query.To = kpiQuery.To
	return query
}

func BuildFiltersAndGroupByBasedOnKPIQuery(query Query, kpiQuery KPIQuery, metric string) Query {
	objectType := GetObjectTypeForQueryExecute(kpiQuery.DisplayCategory, metric, kpiQuery.PageUrl)
	query.EventsWithProperties, query.GlobalUserProperties = getFilterEventsForEventAnalytics(kpiQuery.Filters, objectType)
	query.GroupByProperties = getGroupByEventsForEventsAnalytics(kpiQuery.GroupBy, objectType)
	return query
}

func GetObjectTypeForQueryExecute(displayCategory string, metric string, pageUrl string) string {
	metricsData := MapOfMetricsToData[displayCategory][metric]
	var objectType string
	if displayCategory != PageViewsDisplayCategory {
		objectType = metricsData["object_type"]
	} else if displayCategory == PageViewsDisplayCategory && U.ContainsStringInArray([]string{Entrances, Exits}, metric) {
		objectType = U.EVENT_NAME_SESSION
	} else {
		objectType = pageUrl
	}
	return objectType
}

func GetObjectTypeForFilterValues(displayCategory string, metric string) string {
	var objectType string
	if displayCategory == WebsiteSessionDisplayCategory {
		objectType = U.EVENT_NAME_SESSION
	} else if displayCategory == FormSubmissionsDisplayCategory {
		objectType = U.EVENT_NAME_FORM_SUBMITTED
	} else if U.ContainsStringInArray([]string{HubspotContactsDisplayCategory, HubspotCompaniesDisplayCategory, SalesforceUsersDisplayCategory,
		SalesforceAccountsDisplayCategory, SalesforceOpportunitiesDisplayCategory}, displayCategory) {
		metricsData := MapOfMetricsToData[displayCategory][metric]
		objectType = metricsData["object_type"]
	} else { // pageViews case as default.
		objectType = displayCategory
	}

	return objectType
}

func getFilterEventsForEventAnalytics(filters []KPIFilter, objectType string) ([]QueryEventWithProperties, []QueryProperty) {
	var filterForEventEventAnalytics QueryEventWithProperties
	var filterForUserPropertiesEventAnalytics []QueryProperty
	var currentFilterProperties QueryProperty
	filterForEventEventAnalytics.Name = objectType
	if len(filters) == 0 {
		return []QueryEventWithProperties{filterForEventEventAnalytics}, filterForUserPropertiesEventAnalytics
	}

	for _, filter := range filters {
		currentFilterProperties.Entity = filter.Entity
		currentFilterProperties.Type = filter.PropertyDataType
		currentFilterProperties.Property = filter.PropertyName
		currentFilterProperties.Operator = filter.Condition
		currentFilterProperties.Value = filter.Value
		currentFilterProperties.LogicalOp = filter.LogicalOp
		filterForEventEventAnalytics.Properties = append(filterForEventEventAnalytics.Properties, currentFilterProperties)
	}
	return []QueryEventWithProperties{filterForEventEventAnalytics}, filterForUserPropertiesEventAnalytics
}

func getGroupByEventsForEventsAnalytics(groupBys []KPIGroupBy, objectType string) []QueryGroupByProperty {
	var groupBysForEventAnalytics []QueryGroupByProperty
	var currentGroupByProperty QueryGroupByProperty

	for _, kpiGroupBy := range groupBys {
		currentGroupByProperty = QueryGroupByProperty{}
		currentGroupByProperty.Property = kpiGroupBy.PropertyName
		currentGroupByProperty.Type = kpiGroupBy.PropertyDataType
		currentGroupByProperty.GroupByType = kpiGroupBy.GroupByType //Raw or bucketed
		// currentGroupByProperty.Index = index
		currentGroupByProperty.EventName = objectType
		currentGroupByProperty.EventNameIndex = 1
		currentGroupByProperty.Granularity = kpiGroupBy.Granularity
		currentGroupByProperty.Entity = kpiGroupBy.Entity

		groupBysForEventAnalytics = append(groupBysForEventAnalytics, currentGroupByProperty)
	}
	return groupBysForEventAnalytics
}

func SplitKPIQueryToInternalKPIQueries(query Query, kpiQuery KPIQuery, metric string, transformations []TransformQueryi) []Query {
	var finalResultantQueries []Query
	for _, metricTransformation := range transformations {
		currentQuery := query
		if metricTransformation.Metrics.Entity == EventEntity {
			currentQuery.Type = "events_occurrence"
		} else {
			currentQuery.Type = "unique_users"
		}
		currentQuery.AggregateFunction = metricTransformation.Metrics.Aggregation
		currentQuery.AggregateProperty = metricTransformation.Metrics.Property
		currentQuery.AggregateEntity = metricTransformation.Metrics.Entity
		currentQuery.AggregatePropertyType = metricTransformation.Metrics.GroupByType
		currentQuery.EventsWithProperties = prependEventFiltersBasedOnInternalTransformation(metricTransformation.Filters, query.EventsWithProperties, kpiQuery, metric)
		currentQuery.GlobalUserProperties = prependUserFiltersBasedOnInternalTransformation(metricTransformation.Filters, query.GlobalUserProperties, kpiQuery, metric)
		finalResultantQueries = append(finalResultantQueries, currentQuery)
	}
	return finalResultantQueries
}

func prependEventFiltersBasedOnInternalTransformation(filters []QueryProperty, eventsWithProperties []QueryEventWithProperties, kpiQuery KPIQuery, metric string) []QueryEventWithProperties {
	resultantEventsWithProperties := make([]QueryEventWithProperties, 1)
	var filtersBasedOnMetric []QueryProperty
	if kpiQuery.DisplayCategory == PageViewsDisplayCategory && U.ContainsStringInArray([]string{Entrances, Exits}, metric) {
		for _, filter := range filters {
			filtersBasedOnMetric = append(filtersBasedOnMetric, QueryProperty{
				Entity:    filter.Entity,
				Type:      filter.Type,
				Property:  filter.Property,
				Operator:  filter.Operator,
				LogicalOp: filter.LogicalOp,
				Value:     kpiQuery.PageUrl,
			})
		}
	} else {
		filtersBasedOnMetric = filters
	}
	resultantEventsWithProperties[0].Name = eventsWithProperties[0].Name
	resultantEventsWithProperties[0].AliasName = eventsWithProperties[0].AliasName
	resultantEventsWithProperties[0].Properties = append(filtersBasedOnMetric, eventsWithProperties[0].Properties...)
	return resultantEventsWithProperties
}

func prependUserFiltersBasedOnInternalTransformation(filters []QueryProperty, userProperties []QueryProperty, kpiQuery KPIQuery, metric string) []QueryProperty {
	return make([]QueryProperty, 0)
}

// Functions supporting transforming eventResults to KPIresults
// Note: Considering the format to be generally... event_index, event_name,..., count.
func TransformResultsToKPIResults(results []*QueryResult, hasGroupByTimestamp bool, hasAnyGroupBy bool, displayCategory string, timezoneString string) []*QueryResult {
	resultantResults := make([]*QueryResult, 0)
	for _, result := range results {
		var tmpResult *QueryResult
		tmpResult = &QueryResult{}

		tmpResult.Headers = getTransformedHeaders(result.Headers, hasGroupByTimestamp, hasAnyGroupBy, displayCategory)
		tmpResult.Rows = GetTransformedRows(tmpResult.Headers, result.Rows, hasGroupByTimestamp, hasAnyGroupBy, len(result.Headers), timezoneString)
		resultantResults = append(resultantResults, tmpResult)
	}
	return resultantResults
}

func getTransformedHeaders(headers []string, hasGroupByTimestamp bool, hasAnyGroupBy bool, displayCategory string) []string {
	currentHeaders := make([]string, 0)
	if hasAnyGroupBy && hasGroupByTimestamp {
		currentHeaders = append(headers[1:2], headers[3:]...)
	} else if !hasAnyGroupBy && hasGroupByTimestamp {
		headers[1] = AliasAggr
		currentHeaders = headers
	} else {
		currentHeaders = headers[2:]
	}
	return currentHeaders
}

func GetTransformedRows(headers []string, rows [][]interface{}, hasGroupByTimestamp bool, hasAnyGroupBy bool, headersLen int, timezoneString string) [][]interface{} {
	var currentRows [][]interface{}
	currentRows = make([][]interface{}, 0)
	if len(rows) == 0 {
		return currentRows
	}

	for _, row := range rows {
		var currentRow []interface{}
		if len(row) == 0 {
			currentRow = make([]interface{}, headersLen)
			for index := range currentRow[:headersLen-1] {
				currentRow[index] = ""
			}
			currentRow[headersLen-1] = 0
		} else {
			currentRow = row
		}
		if hasAnyGroupBy && hasGroupByTimestamp {
			currentRow = append(currentRow[1:2], currentRow[3:]...)
			currentRows = append(currentRows, currentRow)
		} else if !hasAnyGroupBy && hasGroupByTimestamp {
			currentRows = append(currentRows, currentRow)
		} else {
			currentRows = append(currentRows, currentRow[2:])
		}
	}

	currentRows = TransformDateTypeValueForEventsKPI(headers, currentRows, hasGroupByTimestamp, timezoneString)
	return currentRows
}

func TransformDateTypeValueForEventsKPI(headers []string, rows [][]interface{}, groupByTimestampPresent bool, timezoneString string) [][]interface{} {
	indexForDateTime := -1
	if !groupByTimestampPresent {
		return rows
	}
	for index, header := range headers {
		if header == "datetime" {
			indexForDateTime = index
			break
		}
	}

	for index, row := range rows {
		currentValueInTimeFormat, _ := row[indexForDateTime].(time.Time)
		rows[index][indexForDateTime] = U.GetTimestampAsStrWithTimezone(currentValueInTimeFormat, timezoneString)
	}
	return rows
}

// Each KPI metric is internally converted to event analytics.
// Considering all rows to be equal in size because of analytics response.
// resultAsMap - key with groupByColumns, value as row.
func HandlingEventResultsByApplyingOperations(results []*QueryResult, transformations []TransformQueryi, timezone string, isTimezoneEnabled bool) QueryResult {
	resultKeys := getAllKeysFromResults(results)
	var finalResult QueryResult
	finalResultRows := make([][]interface{}, 0)
	for index, result := range results {
		if index == 0 {
			resultKeys = addValuesToHashMap(resultKeys, result.Rows)
		} else {
			for _, row := range result.Rows {
				key := U.GetkeyFromRow(row)
				value1 := resultKeys[key]
				value2 := row[len(row)-1]
				operator := transformations[index-1].Metrics.Operator
				result := getValueFromValuesAndOperator(value1, value2, operator)
				resultKeys[key] = result
			}
		}
	}

	for key, value := range resultKeys {
		row := make([]interface{}, 0)
		columns := strings.Split(key, ":;")
		for _, column := range columns[:len(columns)-1] {
			if strings.HasPrefix(column, "dat$") {
				unixValue, _ := strconv.ParseInt(strings.TrimPrefix(column, "dat$"), 10, 64)
				columnValue, _ := U.GetTimeFromUnixTimestampWithZone(unixValue, timezone, isTimezoneEnabled)
				row = append(row, columnValue)
			} else {
				row = append(row, column)
			}
		}
		row = append(row, value)
		finalResultRows = append(finalResultRows, row)
	}
	finalResultRows = U.GetSorted2DArrays(finalResultRows)
	finalResult.Headers = results[0].Headers
	finalResult.Rows = finalResultRows
	return finalResult
}

func getAllKeysFromResults(results []*QueryResult) map[string]interface{} {
	resultKeys := make(map[string]interface{}, 0)
	var key string
	for _, result := range results {
		for _, row := range result.Rows {
			key = U.GetkeyFromRow(row)
			resultKeys[key] = 0
		}
	}
	return resultKeys
}

func addValuesToHashMap(resultKeys map[string]interface{}, rows [][]interface{}) map[string]interface{} {
	for _, row := range rows {
		key := U.GetkeyFromRow(row)
		resultKeys[key] = row[len(row)-1]
	}
	return resultKeys
}

func getValueFromValuesAndOperator(value1 interface{}, value2 interface{}, operator string) float64 {
	var result float64
	value1InFloat := U.SafeConvertToFloat64(value1)
	value2InFloat := U.SafeConvertToFloat64(value2)
	if operator == "Division" {
		if value2InFloat == 0 {
			result = 0
		} else {
			result = value1InFloat / value2InFloat
		}
	}
	return result
}

func makeHashWithKeyAsGroupBy(rows [][]interface{}) map[string][]interface{} {
	var hashMap map[string][]interface{} = make(map[string][]interface{})
	for _, row := range rows {
		key := U.GetkeyFromRow(row)
		hashMap[key] = row
	}
	return hashMap
}
