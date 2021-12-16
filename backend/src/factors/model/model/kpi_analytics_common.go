package model

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"strings"

	log "github.com/sirupsen/logrus"
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

	EventEntity = "event"
	UserEntity  = "user"
)

type KPIQueryGroup struct {
	Class         string       `json:"cl"`
	Queries       []KPIQuery   `json:"qG"`
	GlobalFilters []KPIFilter  `json:"gFil"`
	GlobalGroupBy []KPIGroupBy `json:"gGBy"`
}

func (q *KPIQueryGroup) GetClass() string {
	return q.Class
}

func (q *KPIQueryGroup) GetQueryDateRange() (from, to int64) {
	if len(q.Queries) > 0 {
		// all queries in query group are expected to run for same time range
		return q.Queries[0].From, q.Queries[0].To
	}
	return 0, 0
}

func (q *KPIQueryGroup) SetQueryDateRange(from, to int64) {
	for index, _ := range q.Queries {
		q.Queries[index].From, q.Queries[index].To = from, to
	}
}

func (q *KPIQueryGroup) SetTimeZone(timezoneString U.TimeZoneString) {
	for index, _ := range q.Queries {
		q.Queries[index].Timezone = string(timezoneString)
	}
}

func (q *KPIQueryGroup) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(q.Queries[0].Timezone)
}

func (q *KPIQueryGroup) GetQueryCacheHashString() (string, error) {
	queryMap, err := U.EncodeStructTypeToMap(q)
	if err != nil {
		return "", err
	}
	queries := queryMap["qG"].([]interface{})
	for _, query := range queries {
		delete(query.(map[string]interface{}), "fr")
		delete(query.(map[string]interface{}), "to")
	}

	queryHash, err := U.GenerateHashStringForStruct(queryMap)
	if err != nil {
		return "", err
	}
	return queryHash, nil
}

func (q *KPIQueryGroup) GetQueryCacheRedisKey(projectID uint64) (*cacheRedis.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.Queries[0].From, q.Queries[0].To, U.TimeZoneString(q.Queries[0].Timezone))
	return cacheRedis.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *KPIQueryGroup) GetQueryCacheExpiry() float64 {
	return getQueryCacheResultExpiry(q.Queries[0].From, q.Queries[0].To, q.Queries[0].Timezone)
}

func (q *KPIQueryGroup) GetGroupByTimestamp() string {
	if q.Queries[0].GroupByTimestamp == "" {
		return ""
	}
	return q.Queries[0].GroupByTimestamp
}

func (q *KPIQueryGroup) TransformDateTypeFilters() error {
	timezoneString := q.GetTimeZone()
	err := transformDateTypeFiltersForKPIFilters(q.GlobalFilters, timezoneString)
	if err != nil {
		return err
	}
	for _, query := range q.Queries {
		err := transformDateTypeFiltersForKPIFilters(query.Filters, timezoneString)
		if err != nil {
			return err
		}
	}
	return nil
}

func transformDateTypeFiltersForKPIFilters(filters []KPIFilter, timezoneString U.TimeZoneString) error {
	for i := range filters {
		err := filters[i].TransformDateTypeFilters(timezoneString)
		if err != nil {
			return err
		}
	}
	return nil
}

type KPIQuery struct {
	Category         string       `json:"ca"`
	DisplayCategory  string       `json:"dc"`
	PageUrl          string       `json:"pgUrl"`
	Metrics          []string     `json:"me"`
	Filters          []KPIFilter  `json:"fil"`
	GroupBy          []KPIGroupBy `json:"gBy"`
	GroupByTimestamp string       `json:"gbt"`
	Timezone         string       `json:"tz"`
	From             int64        `json:"fr"`
	To               int64        `json:"to"`
}

type KPIFilter struct {
	ObjectType       string `json:"objTy"`
	PropertyName     string `json:"prNa"`
	PropertyDataType string `json:"prDaTy"`
	Entity           string `json:"en"`
	Condition        string `json:"co"`
	Value            string `json:"va"`
	LogicalOp        string `json:"lOp"`
}

// Duplicate code present between QueryProperty and KPIFilter
func (qp *KPIFilter) TransformDateTypeFilters(timezoneString U.TimeZoneString) error {
	if qp.PropertyDataType == U.PropertyTypeDateTime && (qp.Condition == InLastStr || qp.Condition == NotInLastStr) {
		dateTimeValue, err := DecodeDateTimePropertyValue(qp.Value)
		if err != nil {
			log.WithError(err).Error("Failed reading timestamp on user join query.")
			return err
		}
		lastXthDay := U.GetDateBeforeXPeriod(dateTimeValue.Number, dateTimeValue.Granularity, timezoneString)
		dateTimeValue.From = lastXthDay
		transformedValue, _ := json.Marshal(dateTimeValue)
		qp.Value = string(transformedValue)
	}
	return nil
}

type KPIGroupBy struct {
	ObjectType       string `json:"objTy"`
	PropertyName     string `json:"prNa"`
	PropertyDataType string `json:"prDaTy"`
	GroupByType      string `json:"gbty"`
	Entity           string `json:"en"`
	Granularity      string `json:"gr"`
}

var MapOfMetricsToData = map[string]map[string]map[string]string{
	WebsiteSessionDisplayCategory: {
		TotalSessions:          {"display_name": "Total Sessions", "object_type": U.EVENT_NAME_SESSION},
		UniqueUsers:            {"display_name": "Unique Users", "object_type": U.EVENT_NAME_SESSION},
		NewUsers:               {"display_name": "New Users", "object_type": U.EVENT_NAME_SESSION},
		RepeatUsers:            {"display_name": "Repeat Users", "object_type": U.EVENT_NAME_SESSION},
		SessionsPerUser:        {"display_name": "Session Per User", "object_type": U.EVENT_NAME_SESSION},
		EngagedSessions:        {"display_name": "Engaged Sessions", "object_type": U.EVENT_NAME_SESSION},
		EngagedUsers:           {"display_name": "Engaged Users", "object_type": U.EVENT_NAME_SESSION},
		EngagedSessionsPerUser: {"display_name": "Engaged Sessions per user", "object_type": U.EVENT_NAME_SESSION},
		TotalTimeOnSite:        {"display_name": "Total time on site", "object_type": U.EVENT_NAME_SESSION},
		AvgSessionDuration:     {"display_name": "Avg session duration", "object_type": U.EVENT_NAME_SESSION},
		AvgPageViewsPerSession: {"display_name": "Avg page views per session", "object_type": U.EVENT_NAME_SESSION},
		AvgInitialPageLoadTime: {"display_name": "Avg initial page load time", "object_type": U.EVENT_NAME_SESSION},
		BounceRate:             {"display_name": "Bounce rate", "object_type": U.EVENT_NAME_SESSION},
		EngagementRate:         {"display_name": "Engagement rate", "object_type": U.EVENT_NAME_SESSION},
	},
	PageViewsDisplayCategory: {
		Entrances:                {"display_name": "Entrances", "object_type": U.EVENT_NAME_SESSION},
		Exits:                    {"display_name": "Exits", "object_type": U.EVENT_NAME_SESSION},
		PageViews:                {"display_name": "Page Views"},
		UniqueUsers:              {"display_name": "Unique users"},
		PageviewsPerUser:         {"display_name": "Page views per user"},
		AvgPageLoadTime:          {"display_name": "Avg page load time"},
		AvgVerticalScrollPercent: {"display_name": "Avg vertical scroll percent"},
		AvgTimeOnPage:            {"display_name": "Avg time on page"},
		EngagedPageViews:         {"display_name": "Engaged page views"},
		EngagedUsers:             {"display_name": "Engaged Users"},
		EngagementRate:           {"display_name": "Engagement rate"},
	},
	FormSubmissionsDisplayCategory: {
		Count:        {"display_name": "Count", "object_type": U.EVENT_NAME_FORM_SUBMITTED},
		UniqueUsers:  {"display_name": "Unique users", "object_type": U.EVENT_NAME_FORM_SUBMITTED},
		CountPerUser: {"display_name": "Count per user", "object_type": U.EVENT_NAME_FORM_SUBMITTED},
	},
	HubspotContactsDisplayCategory: {
		CountOfContactsCreated: {"display_name": "Contacts created", "object_type": U.EVENT_NAME_HUBSPOT_CONTACT_CREATED},
		CountOfContactsUpdated: {"display_name": "Contacts updated", "object_type": U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED},
	},
	HubspotCompaniesDisplayCategory: {
		CountOfCompaniesCreated: {"display_name": "Companies created", "object_type": U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED},
		CountOfCompaniesUpdated: {"display_name": "Companies updated", "object_type": U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED},
	},
	// HubspotDealsDisplayCategory: {
	// 	CountOfContactsCreated: {"display_name": "Contacts created", "object_type": U.EVENT_NAME_HUBSPOT_CONTACT_CREATED},
	// 	CountOfContactsUpdated: {"display_name": "Contacts updated", "object_type": U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED},
	// },
	SalesforceUsersDisplayCategory: {
		CountOfContactsCreated: {"display_name": "Contacts created", "object_type": U.EVENT_NAME_SALESFORCE_CONTACT_CREATED},
		CountOfContactsUpdated: {"display_name": "Contacts updated", "object_type": U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED},
	},
	SalesforceAccountsDisplayCategory: {
		CountOfLeadsCreated: {"display_name": "Leads created", "object_type": U.EVENT_NAME_SALESFORCE_LEAD_CREATED},
		CountOfLeadsUpdated: {"display_name": "Leads updated", "object_type": U.EVENT_NAME_SALESFORCE_LEAD_UPDATED},
	},
	SalesforceOpportunitiesDisplayCategory: {
		CountOfOpportunitiesCreated: {"display_name": "Opportunities created", "object_type": U.EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED},
		CountOfOpportunitiesUpdated: {"display_name": "Opportunities updated", "object_type": U.EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED},
	},
	AllChannelsDisplayCategory: {
		"impressions": {"display_name": "Impressions"},
		"clicks":      {"display_name": "Clicks"},
		"spend":       {"display_name": "Spend"},
	},
	AdwordsDisplayCategory: {
		Conversion:                                 {"display_name": "Conversion"},
		ClickThroughRate:                           {"display_name": "Click through rate"},
		ConversionRate:                             {"display_name": "Conversion rate"},
		CostPerClick:                               {"display_name": "Cost per click"},
		CostPerConversion:                          {"display_name": "Cost per conversion"},
		SearchImpressionShare:                      {"display_name": "Search Impr. share"},
		SearchClickShare:                           {"display_name": "Search click share"},
		SearchTopImpressionShare:                   {"display_name": "Search top Impr. share"},
		SearchAbsoluteTopImpressionShare:           {"display_name": "Search abs. top Impr. share"},
		SearchBudgetLostAbsoluteTopImpressionShare: {"display_name": "Search budget lost abs top impr. share"},
		SearchBudgetLostImpressionShare:            {"display_name": "Search budget lost Impr. share"},
		SearchBudgetLostTopImpressionShare:         {"display_name": "Search budget lost top Impr. share"},
		SearchRankLostAbsoluteTopImpressionShare:   {"display_name": "Search rank lost abs. top Impr. share"},
		SearchRankLostImpressionShare:              {"display_name": "Search rank lost Impr. share"},
		SearchRankLostTopImpressionShare:           {"display_name": "Search rank lost top Impr. share"},
	},
	FacebookDisplayCategory: {
		"video_p50_watched_actions":     {"display_name": "Video p50 watched actions"},
		"video_p25_watched_actions":     {"display_name": "Video p25 watched actions"},
		"video_30_sec_watched_actions":  {"display_name": "Video 30 sec watched actions"},
		"video_p100_watched_actions":    {"display_name": "Video p100 watched actions"},
		"video_p75_watched_actions":     {"display_name": "Video p75 watched actions"},
		"cost_per_click":                {"display_name": "Cost per click"},
		"cost_per_link_click":           {"display_name": "Cost per link click"},
		"cost_per_thousand_impressions": {"display_name": "Cost per thousand impressions"},
		"click_through_rate":            {"display_name": "Click through rate"},
		"link_click_through_rate":       {"display_name": "Link click through rate"},
		"link_clicks":                   {"display_name": "Link clicks"},
		"frequency":                     {"display_name": "frequency"},
		"reach":                         {"display_name": "reach"},
	},
}

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
				Filters: []QueryProperty{{Entity: UserEntity, Type: U.PropertyTypeCategorical, Property: U.SP_INITIAL_PAGE_URL, LogicalOp: "AND", Operator: EqualsOpStr, Value: "true"}},
			},
		},
		Exits: []TransformQueryi{
			{
				Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: EventEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""},
				Filters: []QueryProperty{{Entity: UserEntity, Type: U.PropertyTypeCategorical, Property: U.SP_LATEST_PAGE_URL, LogicalOp: "AND", Operator: EqualsOpStr, Value: "true"}},
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
	HubspotContactsDisplayCategory: {
		CountOfContactsCreated: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
		CountOfContactsUpdated: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
	},
	// HubspotCompaniesDisplayCategory: {
	// 	CountOfContactsCreated: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
	// 	CountOfContactsUpdated: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
	// },
	SalesforceUsersDisplayCategory: {
		CountOfContactsCreated: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
		CountOfContactsUpdated: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
	},
	// SalesforceAccountsDisplayCategory: {
	// 	CountOfLeadsCreated: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
	// 	CountOfLeadsUpdated: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
	// },
	// SalesforceOpportunitiesDisplayCategory: {
	// 	CountOfOpportunitiesCreated: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
	// 	CountOfOpportunitiesUpdated: []TransformQueryi{{Metrics: KpiToEventMetricRepr{Aggregation: "count", Entity: UserEntity, Property: "1", GroupByType: U.PropertyTypeCategorical, Operator: ""}}},
	// },
}

type TransformQueryi struct {
	Metrics KpiToEventMetricRepr
	Filters []QueryProperty
}

type KpiToEventMetricRepr struct {
	Aggregation string
	Entity      string
	Property    string
	Operator    string
	GroupByType string
}

// Util/Common Methods.
func GetMetricsForDisplayCategory(category string) []map[string]string {
	resultMetrics := []map[string]string{}
	mapOfMetricsToData := MapOfMetricsToData[category]
	for metricName, data := range mapOfMetricsToData {
		currentMetrics := map[string]string{}
		currentMetrics["name"] = metricName
		currentMetrics["display_name"] = data["display_name"]
		resultMetrics = append(resultMetrics, currentMetrics)
	}
	return resultMetrics
}

func AddObjectTypeToProperties(kpiConfig map[string]interface{}, value string) map[string]interface{} {
	properties := kpiConfig["properties"].([]map[string]string)
	for index := range properties {
		properties[index]["object_type"] = value
	}
	kpiConfig["properties"] = properties
	return kpiConfig
}

func TransformCRMPropertiesToKPIConfigProperties(properties map[string][]string, propertiesToDisplayNames map[string]string, prefix string) []map[string]string {
	var resultantKPIConfigProperties []map[string]string
	var tempKPIConfigProperty map[string]string
	for data_type, propertyNames := range properties {
		for _, propertyName := range propertyNames {
			if strings.HasPrefix(propertyName, prefix) {
				var displayName string
				displayName, exists := propertiesToDisplayNames[propertyName]
				if !exists {
					displayName = propertyName
				}
				tempKPIConfigProperty = map[string]string{
					"name":         propertyName,
					"display_name": displayName,
					"data_type":    data_type,
					"entity":       UserEntity,
				}
				resultantKPIConfigProperties = append(resultantKPIConfigProperties, tempKPIConfigProperty)
			}
		}
	}
	if resultantKPIConfigProperties == nil {
		return make([]map[string]string, 0)
	}
	return resultantKPIConfigProperties
}

func ValidateKPIQueryMetricsForAnyEventType(kpiQueryMetrics []string, mapOfMetrics map[string]map[string]string) bool {
	for _, metric := range kpiQueryMetrics {
		if _, exists := mapOfMetrics[metric]; !exists {
			return false
		}
	}
	return true
}
func ValidateKPIQueryFiltersForAnyEventType(kpiQueryFilters []KPIFilter, configPropertiesData []map[string]string) bool {
	mapOfPropertyName := make(map[string]struct{})
	for _, propertyData := range configPropertiesData {
		mapOfPropertyName[propertyData["name"]] = struct{}{}
	}
	for _, filter := range kpiQueryFilters {
		if _, exists := mapOfPropertyName[filter.PropertyName]; !exists {
			return false
		}
	}
	return true
}

func ValidateKPIQueryGroupByForAnyEventType(kpiQueryGroupBys []KPIGroupBy, configPropertiesData []map[string]string) bool {
	mapOfPropertyName := make(map[string]struct{})
	for _, propertyData := range configPropertiesData {
		mapOfPropertyName[propertyData["name"]] = struct{}{}
	}
	for _, groupBy := range kpiQueryGroupBys {
		if _, exists := mapOfPropertyName[groupBy.PropertyName]; !exists {
			return false
		}
	}
	return true
}

func GetTransformedHeadersForChannels(headers []string) []string {
	currentHeaders := headers
	size := len(currentHeaders)
	currentHeaders[size-1] = AliasAggr
	return currentHeaders
}
