package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"factors/cache"
	C "factors/config"
	U "factors/util"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"

	log "github.com/sirupsen/logrus"
)

type AttributionQuery struct {
	AnalyzeType            string                     `json:"analyze_type"`
	RunType                string                     `json:"run_type"`
	KPI                    KPIQueryGroup              `json:"kpi_query_group"`
	CampaignMetrics        []string                   `json:"cm"`
	ConversionEvent        QueryEventWithProperties   `json:"ce"`
	ConversionEventCompare QueryEventWithProperties   `json:"ce_c"`
	LinkedEvents           []QueryEventWithProperties `json:"lfe"`
	AttributionKey         string                     `json:"attribution_key"`
	// Dimensions related to key
	AttributionKeyDimension  []string `json:"attribution_key_dimensions"`
	AttributionContentGroups []string `json:"attribution_content_groups"`
	// Custom dimensions related to key
	AttributionKeyCustomDimension []string               `json:"attribution_key_custom_dimensions"`
	AttributionKeyFilter          []AttributionKeyFilter `json:"attribution_key_f"`
	AttributionMethodology        string                 `json:"attribution_methodology"`
	AttributionMethodologyCompare string                 `json:"attribution_methodology_c"`
	LookbackDays                  int                    `json:"lbw"`
	From                          int64                  `json:"from"`
	To                            int64                  `json:"to"`
	QueryType                     string                 `json:"query_type"`
	// Tactic or Offer or TacticOffer
	TacticOfferType string `json:"tactic_offer_type"`
	Timezone        string `json:"time_zone"`
}

type KPIInfo struct {
	KpiID                    string        `json:"kpi_id"`
	KpiGroupID               string        `json:"kpi_group_id"`
	KpiAccountCompanyGroupID string        `json:"kpi_acc_comp_group_id"`
	KpiUserIds               []string      `json:"kpi_users"`
	KpiCoalUserIds           []string      `json:"kpi_coal_users"`
	KpiHeaderNames           []string      `json:"kpi_header_names"`  //  headers (in case of multiple KPIs) (revenue, pipleine, dealWon etc)
	KpiAggFunctionTypes      []string      `json:"kpi_agg_fun_types"` //  Agg function type (in case of multiple KPIs) (sum, unique, sum etc)
	KpiValuesList            []KpiRowValue `json:"kpi_value_list"`    // list of rows as object
}

type KpiRowValue struct {
	TimeString  string
	Timestamp   int64     // unix time
	Values      []float64 // list of values (revenue, pipeline, dealWon etc)
	IsConverted bool
}

type AccountCompUsers struct {
	KpiUserIds     []string `json:"kpi_users"`
	KpiCoalUserIds []string `json:"kpi_coal_users"`
}

const (
	AttributionMethodFirstTouch          = "First_Touch"
	AttributionMethodFirstTouchNonDirect = "First_Touch_ND"
	AttributionMethodLastTouch           = "Last_Touch"
	AttributionMethodLastTouchNonDirect  = "Last_Touch_ND"
	AttributionMethodLastCampaignTouch   = "Last_Campaign_Touch"
	AttributionMethodLinear              = "Linear"
	AttributionMethodUShaped             = "U_Shaped"
	AttributionMethodTimeDecay           = "Time_Decay"
	AttributionMethodInfluence           = "Influence"
	AttributionMethodWShaped             = "W_Shaped"

	AttributionKeyCampaign    = "Campaign"
	AttributionKeySource      = "Source"
	AttributionKeyAdgroup     = "AdGroup"
	AttributionKeyKeyword     = "Keyword"
	AttributionKeyChannel     = "ChannelGroup"
	AttributionKeyLandingPage = "LandingPage"
	AttributionKeyAllPageView = "AllPageView"

	ReportCampaign = "Campaign"
	ReportAdGroup  = "AdGroup"
	ReportKeyword  = "Keyword"

	AttributionQueryTypeConversionBased = "ConversionBased"
	AttributionQueryTypeEngagementBased = "EngagementBased"

	SortASC  = "ASC"
	SortDESC = "DESC"

	AttributionErrorIntegrationNotFound = "no ad-words customer account id found for attribution query"

	AdwordsCampaignID   = "campaign_id"
	AdwordsCampaignName = "campaign_name"

	AdwordsAdgroupID   = "ad_group_id"
	AdwordsAdgroupName = "ad_group_name"

	AdwordsKeywordID        = "id"
	AdwordsKeywordName      = "criteria"
	AdwordsKeywordMatchType = "keyword_match_type"

	BingadsCampaignID   = "campaign_id"
	BingadsCampaignName = "campaign_name"

	BingadsAdgroupID   = "ad_group_id"
	BingadsAdgroupName = "ad_group_name"

	BingadsKeywordID   = "keyword_id"
	BingadsKeywordName = "keyword_name"

	FacebookCampaignID   = "campaign_id"
	FacebookCampaignName = "campaign_name"

	FacebookAdgroupID   = "adset_id"
	FacebookAdgroupName = "adset_name"

	LinkedinCampaignID   = "campaign_group_id"
	LinkedinCampaignName = "campaign_group_name"

	LinkedinAdgroupID   = "campaign_id"
	LinkedinAdgroupName = "campaign_name"

	CustomadsCampaignID   = "campaign_id"
	CustomadsCampaignName = "campaign_name"

	CustomadsAdgroupID   = "ad_group_id"
	CustomadsAdgroupName = "ad_group_name"

	CustomadsKeywordID   = "keyword_id"
	CustomadsKeywordName = "keyword_name"

	KeyDelimiter = ":-:"

	ChannelAdwords    = "adwords"
	ChannelBingads    = "bingads"
	ChannelFacebook   = "facebook"
	ChannelLinkedin   = "linkedin"
	ChannelGoogleAds  = "google_ads"
	ChannelBingAds    = "bingads"
	ChannelCustomAds  = "custom_ads"
	SessionChannelOTP = "OfflineTouchPoint"

	FieldChannelName      = "channel_name"
	FieldCampaignName     = "campaign_name"
	FieldAdgroupName      = "adgroup_name"
	FieldKeywordMatchType = "keyword_match_type"
	FieldKeyword          = "keyword"
	FieldSource           = "source"
	FieldChannelGroup     = "channel_group"
	FieldLandingPageUrl   = "landing_page_url"
	FieldAllPageViewUrl   = "all_page_view"

	EventTypeGoalEvent         = 0
	EventTypeLinkedFunnelEvent = 1

	MarketingEventTypeTactic      = "Tactic"
	MarketingEventTypeOffer       = "Offer"
	MarketingEventTypeTacticOffer = "TacticOffer"

	// query.AnalyzeType
	AnalyzeTypeUsers           = "users"                    // Supports RunTypeUser
	AnalyzeTypeSFOpportunities = "salesforce_opportunities" // Supports RunTypeSFOpportunities, RunTypeSFAccounts
	AnalyzeTypeHSDeals         = "hubspot_deals"            // Supports RunTypeHSDeals, RunTypeHSCompanies
	AnalyzeTypeUserKPI         = "user_kpi"                 // Supports

	// query.RunType
	RunTypeSFOpportunities = "salesforce_opportunities"
	RunTypeSFAccounts      = "salesforce_accounts"
	RunTypeHSDeals         = "hubspot_deals"
	RunTypeHSCompanies     = "hubspot_companies"
	RunTypeUser            = "users"
	RunTypeUserKPI         = "user_kpi"

	HSDealIDProperty        = "$hubspot_deal_hs_object_id"
	SFOpportunityIDProperty = "$salesforce_opportunity_id"
	HSCompanyIDProperty     = "$hubspot_company_hs_object_id"
	SFAccountIDProperty     = "$salesforce_account_id"
)

var AddedKeysForCampaign = []string{"ChannelName"}
var AddedKeysForAdgroup = []string{"ChannelName", "Campaign"}
var AddedKeysForKeyword = []string{"ChannelName", "Campaign", "AdGroup", "MatchType"}
var AttributionFixedHeaders = []string{"Impressions", "Clicks", "Spend", "CTR(%)", "Average CPC", "CPM", "ClickConversionRate(%)"}
var AttributionFixedHeadersPostPostConversion = []string{"Cost Per Conversion", "Compare - Users", "Compare - Users (InfluenceRemove)", "Compare Cost Per Conversion"}
var KeyDimensionToHeaderMap = map[string]string{
	FieldChannelName:      "ChannelName",
	FieldCampaignName:     "Campaign",
	FieldAdgroupName:      "AdGroup",
	FieldKeywordMatchType: "MatchType",
	FieldKeyword:          "Keyword",
	FieldSource:           "Source",
	FieldChannelGroup:     "ChannelGroup",
	FieldLandingPageUrl:   "LandingPage",
	FieldAllPageViewUrl:   "AllPageView",
}
var AttributionFixedHeadersLandingPage []string
var AttributionFixedHeadersPostPostConversionLanding = []string{"Compare - Users", "Compare-Users (InfluenceRemove)"}

// ToDo change here as well.
func (query *AttributionQuery) TransformDateTypeFilters() error {
	err := query.ConversionEvent.TransformDateTypeFilters(query.GetTimeZone())
	if err != nil {
		return err
	}
	err = query.ConversionEventCompare.TransformDateTypeFilters(query.GetTimeZone())
	if err != nil {
		return err
	}
	for _, ewp := range query.LinkedEvents {
		err := ewp.TransformDateTypeFilters(query.GetTimeZone())
		if err != nil {
			return err
		}
	}
	return nil
}

// There can be..
func (query *AttributionQuery) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	err := query.ConversionEvent.ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
	if err != nil {
		return err
	}
	err = query.ConversionEventCompare.ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
	if err != nil {
		return err
	}

	for _, ewp := range query.LinkedEvents {
		err = ewp.ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
		if err != nil {
			return err
		}
	}
	return nil
}

func (query *AttributionQueryV1) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	err := query.ConversionEvent.ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
	if err != nil {
		return err
	}
	err = query.ConversionEventCompare.ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
	if err != nil {
		return err
	}

	for _, ewp := range query.LinkedEvents {
		err = ewp.ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
		if err != nil {
			return err
		}
	}
	return nil
}

func (query *AttributionQuery) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(query.Timezone)
}

type AttributionQueryUnit struct {
	Class string                  `json:"cl"`
	Query *AttributionQuery       `json:"query"`
	Meta  *map[string]interface{} `json:"meta"`
}

type AttributionKeyFilter struct {
	AttributionKey string `json:"attribution_key"`
	// Type: categorical or numerical
	Type      string `json:"ty"`
	Property  string `json:"pr"`
	Operator  string `json:"op"`
	Value     string `json:"va"`
	LogicalOp string `json:"lop"`
}

func (q *AttributionQueryUnit) GetClass() string {
	return q.Class
}

func (q *AttributionQueryUnit) GetQueryDateRange() (from, to int64) {
	return q.Query.From, q.Query.To
}

func (q *AttributionQueryUnit) SetTimeZone(timezoneString U.TimeZoneString) {
	q.Query.Timezone = string(timezoneString)
}

func (q *AttributionQueryV1) SetTimeZone(timezoneString U.TimeZoneString) {
	q.Timezone = string(timezoneString)
}

func (q *AttributionQueryUnit) GetTimeZone() U.TimeZoneString {
	return q.Query.GetTimeZone()
}

func (q *AttributionQueryUnit) SetQueryDateRange(from, to int64) {
	q.Query.From, q.Query.To = from, to
}

func (q *AttributionQueryUnit) GetQueryCacheHashString() (string, error) {
	queryMap, err := U.EncodeStructTypeToMap(q)
	if err != nil {
		return "", err
	}
	delete(queryMap, "meta")
	delete(queryMap["query"].(map[string]interface{}), "from")
	delete(queryMap["query"].(map[string]interface{}), "to")

	queryHash, err := U.GenerateHashStringForStruct(queryMap)
	if err != nil {
		return "", err
	}
	return queryHash, nil
}

func (q *AttributionQueryUnit) GetQueryCacheRedisKey(projectID int64) (*cache.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.Query.From, q.Query.To, U.TimeZoneString(q.Query.Timezone))
	return cache.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func GetStringKeyFromCacheRedisKey(Key *cache.Key) string {

	return fmt.Sprintf("pid:%d:puid:%s:%s:%s", Key.ProjectID, Key.ProjectUID, Key.Prefix, Key.Suffix)
}

func (q *AttributionQueryUnit) GetQueryCacheExpiry(projectID int64) float64 {
	return getQueryCacheResultExpiry(projectID, q.Query.From, q.Query.To, q.Query.Timezone)
}

func (q *AttributionQueryUnit) TransformDateTypeFilters() error {
	return q.Query.TransformDateTypeFilters()
}

func (q *AttributionQueryUnit) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	q.Query.ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone)
	return nil
}

func (q *AttributionQueryUnit) CheckIfNameIsPresent(nameOfQuery string) bool {
	return false
}

func (q *AttributionQueryUnit) SetDefaultGroupByTimestamp() {
	if q.Query.KPI.Class == "" {
		return
	}
	for index := range q.Query.KPI.Queries {
		if q.Query.KPI.Queries[index].GroupByTimestamp == GroupByTimestampHour {
			q.Query.KPI.Queries[index].GroupByTimestamp = GroupByTimestampSecond
		}
	}
}

func (q *AttributionQueryUnit) GetGroupByTimestamps() []string {
	return []string{}
}

type MarketingReports struct {
	AdwordsGCLIDData       map[string]MarketingData
	AdwordsCampaignIDData  map[string]MarketingData
	AdwordsCampaignKeyData map[string]MarketingData

	AdwordsAdgroupIDData  map[string]MarketingData
	AdwordsAdgroupKeyData map[string]MarketingData

	AdwordsKeywordIDData  map[string]MarketingData
	AdwordsKeywordKeyData map[string]MarketingData

	BingAdsCampaignIDData  map[string]MarketingData
	BingAdsCampaignKeyData map[string]MarketingData

	BingAdsAdgroupIDData  map[string]MarketingData
	BingAdsAdgroupKeyData map[string]MarketingData

	BingAdsKeywordIDData  map[string]MarketingData
	BingAdsKeywordKeyData map[string]MarketingData

	FacebookCampaignIDData  map[string]MarketingData
	FacebookCampaignKeyData map[string]MarketingData

	FacebookAdgroupIDData  map[string]MarketingData
	FacebookAdgroupKeyData map[string]MarketingData

	LinkedinCampaignIDData  map[string]MarketingData
	LinkedinCampaignKeyData map[string]MarketingData

	LinkedinAdgroupIDData  map[string]MarketingData
	LinkedinAdgroupKeyData map[string]MarketingData

	CustomAdsCampaignIDData  map[string]MarketingData
	CustomAdsCampaignKeyData map[string]MarketingData

	CustomAdsAdgroupIDData  map[string]MarketingData
	CustomAdsAdgroupKeyData map[string]MarketingData

	CustomAdsKeywordIDData  map[string]MarketingData
	CustomAdsKeywordKeyData map[string]MarketingData

	// id = campaignID + KeyDelimiter + campaignName
	AdwordsCampaignDimensions map[string]MarketingData
	// id = campaignID + KeyDelimiter + campaignName + KeyDelimiter + adgroupID + KeyDelimiter + adgroupName
	AdwordsAdgroupDimensions map[string]MarketingData

	// id = campaignID + KeyDelimiter + campaignName
	BingadsCampaignDimensions map[string]MarketingData
	// id = campaignID + KeyDelimiter + campaignName + KeyDelimiter + adgroupID + KeyDelimiter + adgroupName
	BingadsAdgroupDimensions map[string]MarketingData

	// id = campaignID + KeyDelimiter + campaignName
	FacebookCampaignDimensions map[string]MarketingData
	// id = campaignID + KeyDelimiter + campaignName  + KeyDelimiter + adgroupID + KeyDelimiter + adgroupName
	FacebookAdgroupDimensions map[string]MarketingData

	// id = campaignID + KeyDelimiter + campaignName
	LinkedinCampaignDimensions map[string]MarketingData
	// id = campaignID + KeyDelimiter + campaignName  + KeyDelimiter + adgroupID + KeyDelimiter + adgroupName
	LinkedinAdgroupDimensions map[string]MarketingData

	// id = campaignID + KeyDelimiter + campaignName
	CustomAdsCampaignDimensions map[string]MarketingData
	// id = campaignID + KeyDelimiter + campaignName + KeyDelimiter + adgroupID + KeyDelimiter + adgroupName
	CustomAdsAdgroupDimensions map[string]MarketingData
}

type MarketingData struct {
	// Key is CampaignName + AdgroupName + KeywordName + MatchType (i.e. ExtraValue)
	Key string
	// CampaignID, AdgroupID etc
	ID string
	// CampaignName, AdgroupName etc
	Name string
	// For Adwords Keyword Perf report, it is keyword_match_type, for others it is $none
	Channel               string
	CampaignID            string
	CampaignName          string
	AdgroupID             string
	AdgroupName           string
	KeywordMatchType      string
	KeywordID             string
	KeywordName           string
	AdID                  string
	AdName                string
	Slot                  string
	Source                string
	ChannelGroup          string
	LandingPageUrl        string
	AllPageView           string
	TypeName              string
	Impressions           int64
	Clicks                int64
	Spend                 float64
	CustomDimensions      map[string]interface{}
	ContentGroupValuesMap map[string]string
}

type UserSessionData struct {
	MinTimestamp      int64
	MaxTimestamp      int64
	TimeStamps        []int64
	WithinQueryPeriod bool
	MarketingInfo     MarketingData
}

type ContentGroupNameValue struct {
	name  string
	value string
}

type AttributionData struct {
	AddedKeys                            []string
	Name                                 string
	Channel                              string
	CustomDimensions                     map[string]interface{}
	Impressions                          int64
	Clicks                               int64
	Spend                                float64
	CTR                                  float64
	AvgCPC                               float64
	CPM                                  float64
	ClickConversionRate                  float64
	ConvAggFunctionType                  []string
	ConversionEventCount                 []float64
	ConversionEventCountInfluence        []float64
	CostPerConversion                    []float64
	ConversionEventCompareCount          []float64
	ConversionEventCompareCountInfluence []float64
	CostPerConversionCompareCount        []float64
	LinkedEventsCount                    []float64
	LinkedEventsCountInfluence           []float64
	LinkedCostPerEvents                  []float64
	MarketingInfo                        MarketingData
}

type UserInfo struct {
	CoalUserID string
	Timestamp  int64
}

type UserIDPropID struct {
	UserID    string
	Timestamp int64
}

type UserEventInfo struct {
	CoalUserID string
	EventName  string
	Timestamp  int64
	EventType  int
}

const (
	AdwordsClickReportType = 4
	SecsInADay             = int64(86400)
	LookbackCapInDays      = 370
	UserBatchSize          = 2000
	QueryRangeLimit        = 93
	LookBackWindowLimit    = 370
)

// LookbackAdjustedFrom Returns the effective From timestamp considering lookback days
func LookbackAdjustedFrom(from int64, lookbackDays int) int64 {
	lookbackDaysTimestamp := int64(lookbackDays) * SecsInADay
	if LookbackCapInDays < lookbackDays {
		lookbackDaysTimestamp = int64(LookbackCapInDays) * SecsInADay
	}
	validFrom := from - lookbackDaysTimestamp
	return validFrom
}

// LookbackAdjustedTo Returns the effective To timestamp considering lookback days
func LookbackAdjustedTo(to int64, lookbackDays int) int64 {
	lookbackDaysTimestamp := int64(lookbackDays) * SecsInADay
	if LookbackCapInDays < lookbackDays {
		lookbackDaysTimestamp = int64(LookbackCapInDays) * SecsInADay
	}
	validTo := to + lookbackDaysTimestamp
	return validTo
}

// BuildEventNamesPlaceholder Returns the concatenated list of conversion event + funnel events names
func BuildEventNamesPlaceholder(query *AttributionQuery) []string {
	enames := make([]string, 0)
	enames = append(enames, query.ConversionEvent.Name)
	// add name of compare event if given
	if query.ConversionEventCompare.Name != "" {
		enames = append(enames, query.ConversionEventCompare.Name)
	}
	for _, linkedEvent := range query.LinkedEvents {
		enames = append(enames, linkedEvent.Name)
	}
	return enames
}

// UpdateSessionsMapWithCoalesceID Clones a new map replacing userId by coalUserId.
func UpdateSessionsMapWithCoalesceID(attributedSessionsByUserID map[string]map[string]UserSessionData,
	usersInfo map[string]UserInfo, sessionMap *map[string]map[string]UserSessionData) {

	for userID, attributionIdMap := range attributedSessionsByUserID {
		if _, exists := usersInfo[userID]; !exists {
			if C.GetAttributionDebug() == 1 {
				log.WithFields(log.Fields{"Method": "UpdateSessionsMapWithCoalesceID", "UserID": userID}).Info("userID not found")
			}
			continue
		}
		userInfo := usersInfo[userID]
		for attributionID, newUserSession := range attributionIdMap {
			if _, ok := (*sessionMap)[userInfo.CoalUserID]; ok {
				if existingUserSession, ok := (*sessionMap)[userInfo.CoalUserID][attributionID]; ok {
					// Update the existing attribution first and last touch.
					existingUserSession.MinTimestamp = U.Min(existingUserSession.MinTimestamp, newUserSession.MinTimestamp)
					existingUserSession.MaxTimestamp = U.Max(existingUserSession.MaxTimestamp, newUserSession.MaxTimestamp)
					// Merging timestamp of same customer having 2 userIds.
					existingUserSession.TimeStamps = append(existingUserSession.TimeStamps, newUserSession.TimeStamps...)
					existingUserSession.WithinQueryPeriod = existingUserSession.WithinQueryPeriod || newUserSession.WithinQueryPeriod
					(*sessionMap)[userInfo.CoalUserID][attributionID] = existingUserSession
					continue
				}
				(*sessionMap)[userInfo.CoalUserID][attributionID] = newUserSession
				continue
			}
			(*sessionMap)[userInfo.CoalUserID] = make(map[string]UserSessionData)
			(*sessionMap)[userInfo.CoalUserID][attributionID] = newUserSession
		}
	}
}

// FilterNoneKeyForKeywordReport removes userSessions with none KeywordName
func FilterNoneKeyForKeywordReport(sessionMap map[string]map[string]UserSessionData, attributionKey string) (map[string]map[string]UserSessionData, []UserSessionData) {

	if attributionKey != AttributionKeyKeyword {
		return sessionMap, nil
	}

	deletedSessions := make([]UserSessionData, 0)
	// Collect the keys to delete from sessionMap
	keysToDelete := make([]struct{ userID, attributionID string }, 0)
	for userID, attributionIdMap := range sessionMap {
		for attributionID, userSession := range attributionIdMap {
			if userSession.MarketingInfo.KeywordName == PropertyValueNone {
				keysToDelete = append(keysToDelete, struct{ userID, attributionID string }{userID, attributionID})
				deletedSessions = append(deletedSessions, userSession)
			}
		}
	}
	// Delete the keys outside the loop to avoid modifying the map while iterating over it
	for _, key := range keysToDelete {
		userID, attributionID := key.userID, key.attributionID
		if attributionIDMap, ok := sessionMap[userID]; ok {
			if _, ok := attributionIDMap[attributionID]; ok {
				// Create a new map without the deleted key
				newAttributionIDMap := make(map[string]UserSessionData)
				for k, v := range attributionIDMap {
					if k != attributionID {
						newAttributionIDMap[k] = v
					}
				}
				// Replace the old inner map with the new one
				sessionMap[userID] = newAttributionIDMap
			}
		}
	}
	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"Attribution": "Debug",
			"Method":          "FilterNoneKeyForKeywordReport",
			"deletedSessions": deletedSessions}).Info("deletedSessions in FilterNoneKeyForKeywordReport")
	}
	return sessionMap, deletedSessions

}

// AddDefaultAnalyzeType adds default Analyze Type as 'users'
func AddDefaultAnalyzeType(query *AttributionQuery) {

	// Default is set as AnalyzeTypeUsers
	if (*query).AnalyzeType == "" {
		(*query).AnalyzeType = AnalyzeTypeUsers
	}
}

// AddDefaultMarketingEventTypeTacticOffer adds default tactic or offer for older queries
func AddDefaultMarketingEventTypeTacticOffer(query *AttributionQuery) {

	// Default is set as Tactic
	if (*query).TacticOfferType == "" {
		(*query).TacticOfferType = MarketingEventTypeTactic
	}
}

func GetRowsByMapsKPILandingPage(contentGroupNamesList []string, attributionData *map[string]*AttributionData, goalEvents []string, isCompare bool) [][]interface{} {

	var defaultMatchingRow []interface{}

	for range goalEvents {
		//ConversionEventCount, ConversionEventCountInfluence,ConversionEventCompareCount,ConversionEventCompareCountInfluence
		defaultMatchingRow = append(defaultMatchingRow, float64(0), float64(0), float64(0), float64(0))
	}
	var contentGroups []interface{}
	for i := 0; i < len(contentGroupNamesList); i++ {
		contentGroups = append(contentGroups, "none")
	}

	nonMatchingRow := []interface{}{"none"}
	nonMatchingRow = append(nonMatchingRow, contentGroups...)
	nonMatchingRow = append(nonMatchingRow, defaultMatchingRow...)

	rows := make([][]interface{}, 0)
	for _, data := range *attributionData {
		attributionIdName := data.MarketingInfo.LandingPageUrl
		if attributionIdName == "" {
			attributionIdName = PropertyValueNone
		}
		if attributionIdName != "" {

			var row []interface{}
			// Add up Name
			row = append(row, attributionIdName)

			// Add up content Groups
			for i := 0; i < len(contentGroups); i++ {
				if v, exists := data.MarketingInfo.ContentGroupValuesMap[contentGroupNamesList[i]]; exists {
					row = append(row, v)
				} else {
					row = append(row, PropertyValueNone)
				}
			}
			// Append fixed Metrics & ConversionEventCount[0] as only one goal event exists for landing page

			for idx, _ := range data.ConversionEventCount {
				row = append(row, data.ConversionEventCount[idx], data.ConversionEventCountInfluence[idx])
				if isCompare && data.ConversionEventCompareCount != nil {

					row = append(row, data.ConversionEventCompareCount[idx])
					row = append(row, data.ConversionEventCompareCountInfluence[idx])
				} else {
					row = append(row, float64(0), float64(0))
				}
			}

			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		// In case of empty result, send a row of zeros
		rows = append(rows, nonMatchingRow)
	}
	return rows
}

// AddDefaultKeyDimensionsToAttributionQuery adds default custom Dimensions for supporting existing old/saved queries
func AddDefaultKeyDimensionsToAttributionQuery(query *AttributionQuery) {

	if (query.AttributionKeyDimension == nil || len(query.AttributionKeyDimension) == 0) &&
		(query.AttributionKeyCustomDimension == nil || len(query.AttributionKeyCustomDimension) == 0) &&
		(query.AttributionContentGroups == nil || len(query.AttributionContentGroups) == 0) {

		switch query.AttributionKey {
		case AttributionKeyCampaign:
			(*query).AttributionKeyDimension = append((*query).AttributionKeyDimension, FieldCampaignName)
		case AttributionKeyAdgroup:
			(*query).AttributionKeyDimension = append((*query).AttributionKeyDimension, FieldCampaignName, FieldAdgroupName)
		case AttributionKeyKeyword:
			(*query).AttributionKeyDimension = append((*query).AttributionKeyDimension, FieldCampaignName, FieldAdgroupName, FieldKeywordMatchType, FieldKeyword)
		case AttributionKeySource:
			(*query).AttributionKeyDimension = append((*query).AttributionKeyDimension, FieldSource)
		case AttributionKeyChannel:
			(*query).AttributionKeyDimension = append((*query).AttributionKeyDimension, FieldChannelGroup)
		case AttributionKeyLandingPage:
			(*query).AttributionKeyDimension = append((*query).AttributionKeyDimension, FieldLandingPageUrl)
		case AttributionKeyAllPageView:
			(*query).AttributionKeyDimension = append((*query).AttributionKeyDimension, FieldAllPageViewUrl)

		}
	}
}

// EnrichUsingMarketingID Enriched Name using ID for Campaign and Adgroup attribution queries
func EnrichUsingMarketingID(attributionKey string, sessionUTMMarketingValue MarketingData, reports *MarketingReports) (string, MarketingData) {

	// Select the best value for attributionKey
	switch attributionKey {
	case AttributionKeyCampaign:

		ID := sessionUTMMarketingValue.CampaignID
		report := reports.AdwordsCampaignIDData
		if v, ok := report[ID]; ok {
			sessionUTMMarketingValue.CampaignName = v.CampaignName
			sessionUTMMarketingValue.Name = v.CampaignName
			sessionUTMMarketingValue.Channel = ChannelAdwords
			return v.CampaignName, sessionUTMMarketingValue
		}

		report = reports.BingAdsCampaignIDData
		if v, ok := report[ID]; ok {
			sessionUTMMarketingValue.CampaignName = v.CampaignName
			sessionUTMMarketingValue.Name = v.CampaignName
			sessionUTMMarketingValue.Channel = ChannelBingads
			return v.CampaignName, sessionUTMMarketingValue
		}

		report = reports.FacebookCampaignIDData
		if v, ok := report[ID]; ok {
			sessionUTMMarketingValue.CampaignName = v.CampaignName
			sessionUTMMarketingValue.Name = v.CampaignName
			sessionUTMMarketingValue.Channel = ChannelFacebook
			return v.CampaignName, sessionUTMMarketingValue
		}

		report = reports.LinkedinCampaignIDData
		if v, ok := report[ID]; ok {

			sessionUTMMarketingValue.CampaignName = v.CampaignName
			sessionUTMMarketingValue.Name = v.CampaignName
			sessionUTMMarketingValue.Channel = ChannelLinkedin
			return v.CampaignName, sessionUTMMarketingValue
		}

		report = reports.CustomAdsCampaignIDData
		if v, ok := report[ID]; ok {
			sessionUTMMarketingValue.CampaignName = v.CampaignName
			sessionUTMMarketingValue.Name = v.CampaignName
			sessionUTMMarketingValue.Channel = v.Channel
			return v.CampaignName, sessionUTMMarketingValue
		}

	case AttributionKeyAdgroup:
		ID := sessionUTMMarketingValue.AdgroupID
		report := reports.AdwordsAdgroupIDData
		if v, ok := report[ID]; ok {
			sessionUTMMarketingValue.AdgroupName = v.AdgroupName
			sessionUTMMarketingValue.Name = v.AdgroupName
			sessionUTMMarketingValue.Channel = ChannelAdwords
			sessionUTMMarketingValue.CampaignID = v.CampaignID
			sessionUTMMarketingValue.CampaignName = v.CampaignName
			return v.AdgroupName, sessionUTMMarketingValue
		}

		report = reports.BingAdsAdgroupIDData
		if v, ok := report[ID]; ok {
			sessionUTMMarketingValue.AdgroupName = v.AdgroupName
			sessionUTMMarketingValue.Name = v.AdgroupName
			sessionUTMMarketingValue.Channel = ChannelBingads
			sessionUTMMarketingValue.CampaignID = v.CampaignID
			sessionUTMMarketingValue.CampaignName = v.CampaignName
			return v.AdgroupName, sessionUTMMarketingValue
		}

		report = reports.FacebookAdgroupIDData
		if v, ok := report[ID]; ok {
			sessionUTMMarketingValue.AdgroupName = v.AdgroupName
			sessionUTMMarketingValue.Name = v.AdgroupName
			sessionUTMMarketingValue.Channel = ChannelFacebook
			sessionUTMMarketingValue.CampaignID = v.CampaignID
			sessionUTMMarketingValue.CampaignName = v.CampaignName
			return v.AdgroupName, sessionUTMMarketingValue
		}

		report = reports.LinkedinAdgroupIDData
		if v, ok := report[ID]; ok {
			sessionUTMMarketingValue.AdgroupName = v.AdgroupName
			sessionUTMMarketingValue.Name = v.AdgroupName
			sessionUTMMarketingValue.Channel = ChannelLinkedin
			sessionUTMMarketingValue.CampaignID = v.CampaignID
			sessionUTMMarketingValue.CampaignName = v.CampaignName
			return v.AdgroupName, sessionUTMMarketingValue
		}

		report = reports.CustomAdsAdgroupIDData
		if v, ok := report[ID]; ok {
			sessionUTMMarketingValue.AdgroupName = v.AdgroupName
			sessionUTMMarketingValue.Name = v.AdgroupName
			sessionUTMMarketingValue.Channel = v.Channel
			sessionUTMMarketingValue.CampaignID = v.CampaignID
			sessionUTMMarketingValue.CampaignName = v.CampaignName
			return v.AdgroupName, sessionUTMMarketingValue
		}
	default:
		// No enrichment for other types using ID
		return PropertyValueNone, sessionUTMMarketingValue
	}
	return PropertyValueNone, sessionUTMMarketingValue
}

// EnrichUsingGCLID Returns the matching value for GCLID, if not found returns $none
func EnrichUsingGCLID(gclIDBasedCampaign *map[string]MarketingData, gclID string,
	attributionKey string, sessionUTMMarketingValue MarketingData) (string, MarketingData) {

	if v, ok := (*gclIDBasedCampaign)[gclID]; ok {

		enrichMarketData(&v, &sessionUTMMarketingValue)

		// Select the best value for attributionKey
		switch attributionKey {
		case AttributionKeyCampaign:
			if U.IsNonEmptyKey(v.CampaignName) {
				return v.CampaignName, sessionUTMMarketingValue
			}
			return v.CampaignID, sessionUTMMarketingValue
		case AttributionKeyAdgroup:
			if U.IsNonEmptyKey(v.AdgroupName) {
				return v.AdgroupName, sessionUTMMarketingValue
			}
			return v.AdgroupID, sessionUTMMarketingValue
		case AttributionKeyKeyword:
			if U.IsNonEmptyKey(v.KeywordName) {
				return v.KeywordName, sessionUTMMarketingValue
			}
			return v.KeywordID, sessionUTMMarketingValue
		default:
			// No enrichment for Source via GCLID
			return PropertyValueNone, sessionUTMMarketingValue
		}
	}
	return PropertyValueNone, sessionUTMMarketingValue
}

func enrichMarketData(v *MarketingData, sessionUTMMarketingValue *MarketingData) {

	if U.IsNonEmptyKey(v.CampaignID) {
		sessionUTMMarketingValue.CampaignID = v.CampaignID
	}
	if U.IsNonEmptyKey(v.CampaignName) {
		sessionUTMMarketingValue.CampaignName = v.CampaignName
	}
	if U.IsNonEmptyKey(v.AdgroupID) {
		sessionUTMMarketingValue.AdgroupID = v.AdgroupID
	}
	if U.IsNonEmptyKey(v.AdgroupName) {
		sessionUTMMarketingValue.AdgroupName = v.AdgroupName
	}
	if U.IsNonEmptyKey(v.KeywordMatchType) {
		sessionUTMMarketingValue.KeywordMatchType = v.KeywordMatchType
	}
	if U.IsNonEmptyKey(v.KeywordID) {
		sessionUTMMarketingValue.KeywordID = v.KeywordID
	}
	if U.IsNonEmptyKey(v.KeywordName) {
		sessionUTMMarketingValue.KeywordName = v.KeywordName
	}
	if U.IsNonEmptyKey(v.AdID) {
		sessionUTMMarketingValue.AdID = v.AdID
	}
	if U.IsNonEmptyKey(v.AdName) {
		sessionUTMMarketingValue.AdName = v.AdName
	}
	if U.IsNonEmptyKey(v.Slot) {
		sessionUTMMarketingValue.Slot = v.Slot
	}
	if U.IsNonEmptyKey(v.Source) {
		sessionUTMMarketingValue.Source = v.Source
	}
	if U.IsNonEmptyKey(v.ChannelGroup) {
		sessionUTMMarketingValue.ChannelGroup = v.ChannelGroup
	}
}

// IsASearchSlotKeyword Returns true if it is not a search click using slot's v
func IsASearchSlotKeyword(gclIDBasedCampaign *map[string]MarketingData, gclID string) bool {
	if value, ok := (*gclIDBasedCampaign)[gclID]; ok {
		if strings.Contains(strings.ToLower(value.Slot), "search") {
			return true
		}
	}
	return false
}

func IsIntegrationNotFoundError(err error) bool {
	return err.Error() == AttributionErrorIntegrationNotFound
}

// GetQuerySessionProperty Maps the {attribution key} to the session properties field
func GetQuerySessionProperty(attributionKey string) (string, error) {
	if attributionKey == AttributionKeyCampaign {
		return U.EP_CAMPAIGN, nil
	} else if attributionKey == AttributionKeySource {
		return U.EP_SOURCE, nil
	} else if attributionKey == AttributionKeyChannel {
		return U.EP_CHANNEL, nil
	} else if attributionKey == AttributionKeyAdgroup {
		return U.EP_ADGROUP, nil
	} else if attributionKey == AttributionKeyKeyword {
		return U.EP_KEYWORD, nil
	} else if attributionKey == AttributionKeyLandingPage {
		return U.UP_INITIAL_PAGE_URL, nil
	} else if attributionKey == AttributionKeyAllPageView {
		return U.EP_PAGE_URL, nil
	}
	return "", errors.New("invalid query properties")
}

// GetAttributionKeyForOffline Maps the {attribution key} to the offline touch points
func GetAttributionKeyForOffline(attributionKey string) (string, error) {
	if attributionKey == AttributionKeyCampaign {
		return U.EP_CAMPAIGN, nil
	} else if attributionKey == AttributionKeySource {
		return U.EP_SOURCE, nil
	} else if attributionKey == AttributionKeyChannel {
		return U.EP_CHANNEL, nil
	} else if attributionKey == AttributionKeyAdgroup {
		return U.EP_ADGROUP, nil
	}
	return "", errors.New("invalid query properties for offline touch point")
}

func addFixedMetrics(result *QueryResult, query *AttributionQuery) {
	result.Headers = append(result.Headers, AttributionFixedHeadersLandingPage...)
	conversionEventUsers := fmt.Sprintf("%s - Users", query.ConversionEvent.Name)
	result.Headers = append(result.Headers, conversionEventUsers)
	conversionEventUsersInfluence := fmt.Sprintf("%s - Users (InfluenceRemove)", query.ConversionEvent.Name)
	result.Headers = append(result.Headers, conversionEventUsersInfluence)
	result.Headers = append(result.Headers, AttributionFixedHeadersPostPostConversionLanding...)
	if len(query.LinkedEvents) > 0 {
		for _, event := range query.LinkedEvents {
			result.Headers = append(result.Headers, fmt.Sprintf("%s - Users", event.Name))
			result.Headers = append(result.Headers, fmt.Sprintf("%s- Users (InfluenceRemove)", event.Name))
		}
	}

}

// AddHeadersByAttributionKey Adds common column names and linked events as header to the result rows.
func AddHeadersByAttributionKey(result *QueryResult, query *AttributionQuery, goalEvents []string, goalEventAggFuncTypes []string) {

	attributionKey := query.AttributionKey
	if attributionKey == AttributionKeyLandingPage || attributionKey == AttributionKeyAllPageView {
		// add up the attribution key
		result.Headers = append(result.Headers, attributionKey)

		// add up content groups
		for _, contentGroupName := range query.AttributionContentGroups {
			result.Headers = append(result.Headers, contentGroupName)
		}

		// add up fixed metrics

		if query.AnalyzeType == AnalyzeTypeHSDeals || query.AnalyzeType == AnalyzeTypeSFOpportunities || query.AnalyzeType == AnalyzeTypeUserKPI {

			result.Headers = append(result.Headers, AttributionFixedHeadersLandingPage...)

			for _, goal := range goalEvents {

				conversion := fmt.Sprintf("%s - Conversion", goal)
				conversionInfluence := fmt.Sprintf("%s - Conversion Value (InfluenceRemove) ", goal)
				result.Headers = append(result.Headers, conversion, conversionInfluence)

				conversionC := fmt.Sprintf("%s - Conversion(compare)", goal)
				conversionCInfluence := fmt.Sprintf("%s - Conversion InfluenceRemove(compare)", goal)
				result.Headers = append(result.Headers, conversionC, conversionCInfluence)
			}

		} else {
			addFixedMetrics(result, query)
		}

	} else if query.AnalyzeType == AnalyzeTypeUsers {

		// Add up for Added Keys {Campaign, Adgroup, Keyword}
		switch attributionKey {
		case AttributionKeyCampaign:
			result.Headers = append(result.Headers, AddedKeysForCampaign...)
		case AttributionKeyAdgroup:
			result.Headers = append(result.Headers, AddedKeysForAdgroup...)
		case AttributionKeyKeyword:
			result.Headers = append(result.Headers, AddedKeysForKeyword...)
		default:
		}

		// add up the attribution key
		result.Headers = append(result.Headers, attributionKey)

		// add up custom dimensions
		for _, key := range query.AttributionKeyCustomDimension {
			result.Headers = append(result.Headers, key)
		}

		// add up fixed metrics
		result.Headers = append(result.Headers, AttributionFixedHeaders...)
		conversionEventUsers := fmt.Sprintf("%s - Users", query.ConversionEvent.Name)
		result.Headers = append(result.Headers, conversionEventUsers)
		conversionEventUsersInfluence := fmt.Sprintf("%s-Users (InfluenceRemove)", query.ConversionEvent.Name)
		result.Headers = append(result.Headers, conversionEventUsersInfluence)
		result.Headers = append(result.Headers, AttributionFixedHeadersPostPostConversion...)
		if len(query.LinkedEvents) > 0 {
			for _, event := range query.LinkedEvents {
				result.Headers = append(result.Headers, fmt.Sprintf("%s - Users", event.Name))
				result.Headers = append(result.Headers, fmt.Sprintf("%s- Users (InfluenceRemove)", event.Name))
				result.Headers = append(result.Headers, fmt.Sprintf("%s - CPC", event.Name))
			}
		}

		// add up key
		result.Headers = append(result.Headers, "key")
	} else if query.AnalyzeType == AnalyzeTypeHSDeals || query.AnalyzeType == AnalyzeTypeSFOpportunities || query.AnalyzeType == AnalyzeTypeUserKPI {

		// Add up for Added Keys {Campaign, Adgroup, Keyword,Source,Channel}
		switch attributionKey {
		case AttributionKeyCampaign:
			result.Headers = append(result.Headers, AddedKeysForCampaign...)
		case AttributionKeyAdgroup:
			result.Headers = append(result.Headers, AddedKeysForAdgroup...)
		case AttributionKeyKeyword:
			result.Headers = append(result.Headers, AddedKeysForKeyword...)
		default:
		}

		// add up the attribution key
		result.Headers = append(result.Headers, attributionKey)

		// add up custom dimensions
		for _, key := range query.AttributionKeyCustomDimension {
			result.Headers = append(result.Headers, key)
		}

		// add up fixed metrics
		result.Headers = append(result.Headers, AttributionFixedHeaders...)

		for idx, goal := range goalEvents {

			if strings.ToLower(goalEventAggFuncTypes[idx]) == "sum" {
				conversion := fmt.Sprintf("%s - Conversion Value", goal)
				conversionInfluence := fmt.Sprintf("%s - Conversion Value (InfluenceRemove) ", goal)
				cpc := fmt.Sprintf("%s - Return on Cost", goal)
				result.Headers = append(result.Headers, conversion, conversionInfluence, cpc)

				conversionC := fmt.Sprintf("%s - Conversion Value(compare)", goal)
				conversionCInfluence := fmt.Sprintf("%s - Conversion Value InfluenceRemove(compare)", goal)
				cpcC := fmt.Sprintf("%s - Return on Cost(compare)", goal)
				result.Headers = append(result.Headers, conversionC, conversionCInfluence, cpcC)
			} else {
				conversion := fmt.Sprintf("%s - Conversion", goal)
				conversionInfluence := fmt.Sprintf("%s - Conversion Value (InfluenceRemove) ", goal)
				cpc := fmt.Sprintf("%s - Cost Per Conversion", goal)
				result.Headers = append(result.Headers, conversion, conversionInfluence, cpc)

				conversionC := fmt.Sprintf("%s - Conversion(compare)", goal)
				conversionCInfluence := fmt.Sprintf("%s - Conversion InfluenceRemove(compare)", goal)
				cpcC := fmt.Sprintf("%s - Cost Per Conversion(compare)", goal)
				result.Headers = append(result.Headers, conversionC, conversionCInfluence, cpcC)
			}
		}

		// add up key
		result.Headers = append(result.Headers, "key")
	}
}

func getLinkedEventColumnAsInterfaceListv1(spend float64, data []float64, dataInfluence []float64, linkedEventCount int) []interface{} {

	var list []interface{}
	// If empty linked events, add 0s
	if len(data) == 0 {
		for i := 0; i < linkedEventCount; i++ {
			list = append(list, 0.0, 0.0, 0.0)
		}
	} else {
		for i := 0; i < len(data) && i < len(dataInfluence); i++ {
			cpc := 0.0
			if data[i] > 0.0 {
				cpc, _ = U.FloatRoundOffWithPrecision(spend/data[i], U.DefaultPrecision)
			}
			list = append(list, data[i], dataInfluence[i], cpc)
		}
	}
	// Each LE should have 3 values, one for conversion, 2nd for conversion Influence and 3rd for conversion cost
	for len(list) < 3*linkedEventCount {
		list = append(list, 0.0)
	}
	return list
}

func getLinkedEventColumnAsInterfaceListLandingPagev1(data []float64, dataInfluence []float64, linkedEventCount int) []interface{} {

	var list []interface{}
	// If empty linked events, add 0s
	if len(data) == 0 {
		for i := 0; i < linkedEventCount; i++ {
			// Each LE should have 2 values, one for conversion and other for conversion Influence
			list = append(list, 0.0, 0.0)
		}
	} else {
		for i := 0; i < len(data) && i < len(dataInfluence); i++ {
			list = append(list, data[i], dataInfluence[i])
		}
	}
	return list
}

func GetKeyIndexOrAddedKeySize(attributionKey string) int {

	addedKeysSize := 0
	// Add up for Added Keys {Campaign, Adgroup, Keyword}
	switch attributionKey {
	case AttributionKeyCampaign:
		addedKeysSize = 1
	case AttributionKeyAdgroup:
		addedKeysSize = 2
	case AttributionKeyKeyword:
		addedKeysSize = 4
	default:
	}
	return addedKeysSize
}

func GetNoneKeyForAttributionType(attributionKey string) string {

	key := ""
	for i := 0; i < GetKeyIndexOrAddedKeySize(attributionKey); i++ {
		key = key + PropertyValueNone + KeyDelimiter
	}
	key = key + PropertyValueNone
	return key

}

func GetConversionIndex(headers []string) int {
	for index, val := range headers {
		// matches the first conversion
		if strings.Contains(val, "- Users") {
			return index
		}
	}
	return -1
}

func GetConversionIndexKPI(headers []string) int {
	for index, val := range headers {
		// matches the first conversion
		if strings.Contains(val, "- Conversion") {
			return index
		}
	}
	return -1
}

func GetSecondConversionIndexKPI(headers []string) int {
	isSecond := false
	for index, val := range headers {
		// matches the second conversion
		if strings.HasSuffix(val, "- Conversion") {
			if isSecond {
				return index
			} else {
				isSecond = true
			}
		}
	}
	return -1
}

func GetLastKeyValueIndex(headers []string) int {
	for index, val := range headers {
		if val == "Impressions" {
			return index - 1
		}
	}
	return -1
}

func GetLastKeyValueIndexLandingPage(headers []string) int {
	for index, val := range headers {
		if strings.Contains(val, "-") {
			return index - 1
		}
	}
	return -1
}

func GetImpressionsIndex(headers []string) int {
	for index, val := range headers {
		if val == "Impressions" {
			return index
		}
	}
	return -1
}

func GetCompareConversionUserCountIndex(headers []string) int {
	for index, val := range headers {
		if val == "Compare - Users" {
			return index
		}
	}
	return -1
}

func getAttributionIdName(data *AttributionData, attributionKey string) string {
	if attributionKey == AttributionKeyAllPageView {
		return data.MarketingInfo.AllPageView
	}
	return data.MarketingInfo.LandingPageUrl
}

func GetSpendIndex(headers []string) int {
	for index, val := range headers {
		if val == "Spend" {
			return index + 1
		}
	}
	return -1
}

func GetRowsByMapsKPIPage(attributionKey string, contentGroupNamesList []string, attributionData *map[string]*AttributionData, goalEvents []string, isCompare bool) [][]interface{} {

	var defaultMatchingRow []interface{}

	for range goalEvents {
		//ConversionEventCount, ConversionEventCountInfluence,ConversionEventCompareCount,ConversionEventCompareCountInfluence
		defaultMatchingRow = append(defaultMatchingRow, float64(0), float64(0), float64(0), float64(0))
	}

	var contentGroups []interface{}
	for i := 0; i < len(contentGroupNamesList); i++ {
		contentGroups = append(contentGroups, "none")
	}

	nonMatchingRow := []interface{}{"none"}
	nonMatchingRow = append(nonMatchingRow, contentGroups...)
	nonMatchingRow = append(nonMatchingRow, defaultMatchingRow...)

	rows := make([][]interface{}, 0)
	for _, data := range *attributionData {
		attributionIdName := getAttributionIdName(data, attributionKey)
		if attributionIdName == "" {
			attributionIdName = PropertyValueNone
		}
		if attributionIdName != "" {

			var row []interface{}
			// Add up Name
			row = append(row, attributionIdName)

			// Add up content Groups
			for i := 0; i < len(contentGroups); i++ {
				if v, exists := data.MarketingInfo.ContentGroupValuesMap[contentGroupNamesList[i]]; exists {
					row = append(row, v)
				} else {
					row = append(row, PropertyValueNone)
				}
			}
			// Append fixed Metrics & ConversionEventCount[0] as only one goal event exists for landing page

			for idx, _ := range data.ConversionEventCount {
				row = append(row, data.ConversionEventCount[idx], data.ConversionEventCountInfluence[idx])
				if isCompare && data.ConversionEventCompareCount != nil {

					row = append(row, data.ConversionEventCompareCount[idx])
					row = append(row, data.ConversionEventCompareCountInfluence[idx])
				} else {
					row = append(row, float64(0), float64(0))
				}
			}

			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		// In case of empty result, send a row of zeros
		rows = append(rows, nonMatchingRow)
	}
	return rows
}

// GetRowsByMaps Returns result in from of metrics. For empty attribution id, the values are accumulated into "$none".
func GetRowsByMaps(attributionKey string, dimensions []string, attributionData *map[string]*AttributionData,
	linkedEvents []QueryEventWithProperties, goalEvents []string, isCompare bool) [][]interface{} {

	// Name, impression, clicks, spend
	defaultMatchingRow := []interface{}{int64(0), int64(0), float64(0),
		// (CTR, AvgCPC, CPM, ClickConversionRate)
		float64(0), float64(0), float64(0), float64(0)}

	doneAddingDefault := false // doing it one time
	noOfGoalEvents := 0

	if *attributionData != nil {
		for _, data := range *attributionData {
			for idx := 0; idx < len(data.ConversionEventCount); idx++ {

				noOfGoalEvents++
				// one for each - ConversionEventCount, ConversionEventCountInfluence, CostPerConversion, ConversionEventCompareCount, ConversionEvenCompareCountInfluence, CostPerConversionCompareCount
				defaultMatchingRow = append(defaultMatchingRow, float64(0), float64(0), float64(0), float64(0), float64(0), float64(0))
				doneAddingDefault = true
			}
			if doneAddingDefault {
				break
			}
		}
	}
	if !doneAddingDefault {
		for range goalEvents {
			defaultMatchingRow = append(defaultMatchingRow, float64(0), float64(0), float64(0), float64(0), float64(0), float64(0))
		}
	}

	var customDims []interface{}
	for i := 0; i < len(dimensions); i++ {
		customDims = append(customDims, "none")
	}

	addedKeysSize := GetKeyIndexOrAddedKeySize(attributionKey)
	nonMatchingRow := []interface{}{"none"}
	// Add up for Added Keys {Campaign, Adgroup, Keyword}
	switch attributionKey {
	case AttributionKeyCampaign:
		nonMatchingRow = append(nonMatchingRow, "none") // channel
		nonMatchingRow = append(nonMatchingRow, customDims...)
		nonMatchingRow = append(nonMatchingRow, defaultMatchingRow...)
	case AttributionKeyAdgroup:
		nonMatchingRow = append(nonMatchingRow, "none", "none") // channel, camp
		nonMatchingRow = append(nonMatchingRow, customDims...)
		nonMatchingRow = append(nonMatchingRow, defaultMatchingRow...)
	case AttributionKeyKeyword:
		nonMatchingRow = append(nonMatchingRow, "none", "none", "none", "none") // channel, camp, adgroup, match_type
		nonMatchingRow = append(nonMatchingRow, customDims...)
		nonMatchingRow = append(nonMatchingRow, defaultMatchingRow...)
	case AttributionKeySource:
		nonMatchingRow = append(nonMatchingRow, defaultMatchingRow...)
	case AttributionKeyChannel:
		nonMatchingRow = append(nonMatchingRow, defaultMatchingRow...)
	default:
		nonMatchingRow = append(nonMatchingRow, defaultMatchingRow...)
	}

	// Add up for linkedEvents for conversion, conversion Influence, CPC
	for i := 0; i < len(linkedEvents); i++ {
		nonMatchingRow = append(nonMatchingRow, float64(0), float64(0), float64(0))
	}
	// Add up for key
	nonMatchingRow = append(nonMatchingRow, "none")

	rows := make([][]interface{}, 0)
	for key, data := range *attributionData {
		attributionIdName := ""
		switch attributionKey {
		case AttributionKeyCampaign:
			attributionIdName = data.MarketingInfo.CampaignName
		case AttributionKeyAdgroup:
			attributionIdName = data.MarketingInfo.AdgroupName
		case AttributionKeyKeyword:
			attributionIdName = data.MarketingInfo.KeywordName
		case AttributionKeySource:
			attributionIdName = data.MarketingInfo.Source
		case AttributionKeyChannel:
			attributionIdName = data.MarketingInfo.ChannelGroup
		case AttributionKeyLandingPage:
			attributionIdName = data.MarketingInfo.LandingPageUrl
		default:
		}
		if attributionIdName == "" {
			attributionIdName = PropertyValueNone
		}
		if attributionIdName != "" {

			var row []interface{}

			// Add up keys
			for i := 0; i < addedKeysSize; i++ {
				if data.AddedKeys != nil && data.AddedKeys[i] != "" {
					row = append(row, data.AddedKeys[i])
				} else {
					row = append(row, PropertyValueNone)
				}
			}

			// Add up Name
			row = append(row, attributionIdName)

			// Add up custom dimensions
			for i := 0; i < len(dimensions); i++ {
				if v, exists := data.CustomDimensions[dimensions[i]]; exists {
					row = append(row, v)
				} else {
					row = append(row, PropertyValueNone)
				}
			}
			// Append fixed Metrics
			row = append(row, data.Impressions, data.Clicks, data.Spend, data.CTR, data.AvgCPC,
				data.CPM, data.ClickConversionRate)

			var cpc []float64

			var cpcCompare []float64

			for len(data.ConversionEventCount) < noOfGoalEvents {
				data.ConversionEventCount = append(data.ConversionEventCount, float64(0))
			}
			for len(data.ConversionEventCountInfluence) < noOfGoalEvents {
				data.ConversionEventCountInfluence = append(data.ConversionEventCountInfluence, float64(0))
			}

			for len(data.ConversionEventCompareCount) < noOfGoalEvents {
				data.ConversionEventCompareCount = append(data.ConversionEventCompareCount, float64(0))
			}
			for len(data.ConversionEventCompareCountInfluence) < noOfGoalEvents {
				data.ConversionEventCompareCountInfluence = append(data.ConversionEventCompareCountInfluence, float64(0))
			}

			for idx := 0; idx < len(data.ConversionEventCount); idx++ {

				functionType := data.ConvAggFunctionType[idx]

				row = append(row, float64(data.ConversionEventCount[idx]))

				row = append(row, float64(data.ConversionEventCountInfluence[idx]))

				cpc = append(cpc, float64(0))

				if strings.ToLower(functionType) == "sum" {
					if data.Spend > 0.0 {
						cpc[idx], _ = U.FloatRoundOffWithPrecision(data.ConversionEventCount[idx]/data.Spend, U.DefaultPrecision)
					}
				} else {
					if data.ConversionEventCount[idx] > 0.0 {
						cpc[idx], _ = U.FloatRoundOffWithPrecision(data.Spend/data.ConversionEventCount[idx], U.DefaultPrecision)
					}
				}

				if isCompare {
					cpcCompare = append(cpcCompare, float64(0))

					if strings.ToLower(functionType) == "sum" {
						if data.Spend > 0.0 {
							cpcCompare[idx], _ = U.FloatRoundOffWithPrecision(data.ConversionEventCompareCount[idx]/data.Spend, U.DefaultPrecision)
						}
					} else {
						if data.ConversionEventCompareCount[idx] > 0.0 {
							cpcCompare[idx], _ = U.FloatRoundOffWithPrecision(data.Spend/data.ConversionEventCompareCount[idx], U.DefaultPrecision)
						}
					}

					row = append(row, cpc[idx])
					row = append(row, float64(data.ConversionEventCompareCount[idx]))
					row = append(row, float64(data.ConversionEventCompareCountInfluence[idx]))
					row = append(row, cpcCompare[idx])
				} else {
					row = append(row, cpc[idx], float64(0), float64(0), float64(0))
				}

			}
			// for linked event considering the data.ConversionEventCount[0] only
			row = append(row, getLinkedEventColumnAsInterfaceListv1(data.Spend, data.LinkedEventsCount, data.LinkedEventsCountInfluence, len(linkedEvents))...)
			// Add up key
			row = append(row, key)
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		// In case of empty result, send a row of zeros
		rows = append(rows, nonMatchingRow)
	}
	return rows
}

// GetRowsByMapsAllPage Returns result in from of metrics. For empty attribution id, the values are accumulated into "$none".
func GetRowsByMapsAllPage(contentGroupNamesList []string, attributionData *map[string]*AttributionData,
	linkedEvents []QueryEventWithProperties, isCompare bool) [][]interface{} {

	var defaultMatchingRow []interface{}

	//ConversionEventCount, ConversionEventCountInfluence,ConversionEventCompareCount,ConversionEventCompareCountInfluence
	defaultMatchingRow = append(defaultMatchingRow, float64(0), float64(0), float64(0), float64(0))

	var contentGroups []interface{}
	for i := 0; i < len(contentGroupNamesList); i++ {
		contentGroups = append(contentGroups, "none")
	}

	nonMatchingRow := []interface{}{"none"}
	nonMatchingRow = append(nonMatchingRow, contentGroups...)
	nonMatchingRow = append(nonMatchingRow, defaultMatchingRow...)

	// Add up for linkedEvents for conversion and conversion rate
	for i := 0; i < len(linkedEvents); i++ {
		nonMatchingRow = append(nonMatchingRow, float64(0), float64(0))
	}
	rows := make([][]interface{}, 0)
	for _, data := range *attributionData {
		attributionIdName := data.MarketingInfo.AllPageView
		if attributionIdName == "" {
			attributionIdName = PropertyValueNone
		}
		if attributionIdName != "" {

			var row []interface{}
			// Add up Name
			row = append(row, attributionIdName)

			// Add up content Groups
			for i := 0; i < len(contentGroups); i++ {
				if v, exists := data.MarketingInfo.ContentGroupValuesMap[contentGroupNamesList[i]]; exists {
					row = append(row, v)
				} else {
					row = append(row, PropertyValueNone)
				}
			}
			// Validate if conversion values are not nil
			if data.ConversionEventCount == nil {
				emptyConversionEventRow := make([]float64, 0)
				emptyConversionEventRow = append(emptyConversionEventRow, 0.0)
				data.ConversionEventCount = emptyConversionEventRow
			}
			if data.ConversionEventCountInfluence == nil {
				emptyConversionEventRowInfluence := make([]float64, 0)
				emptyConversionEventRowInfluence = append(emptyConversionEventRowInfluence, 0.0)
				data.ConversionEventCountInfluence = emptyConversionEventRowInfluence
			}
			// Append fixed Metrics & ConversionEventCount[0] as only one goal event exists for landing page
			row = append(row, data.ConversionEventCount[0], data.ConversionEventCountInfluence[0])

			if isCompare && data.ConversionEventCompareCount != nil {

				row = append(row, data.ConversionEventCompareCount[0])
				row = append(row, data.ConversionEventCompareCountInfluence[0])
			} else {
				row = append(row, float64(0), float64(0))
			}
			row = append(row, getLinkedEventColumnAsInterfaceListLandingPagev1(data.LinkedEventsCount, data.LinkedEventsCountInfluence, len(linkedEvents))...)
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		// In case of empty result, send a row of zeros
		rows = append(rows, nonMatchingRow)
	}
	return rows
}

// GetRowsByMapsLandingPage Returns result in from of metrics. For empty attribution id, the values are accumulated into "$none".
func GetRowsByMapsLandingPage(contentGroupNamesList []string, attributionData *map[string]*AttributionData,
	linkedEvents []QueryEventWithProperties, isCompare bool) [][]interface{} {

	var defaultMatchingRow []interface{}

	//ConversionEventCount, ConversionEventCountInfluence,ConversionEventCompareCount,ConversionEventCompareCountInfluence
	defaultMatchingRow = append(defaultMatchingRow, float64(0), float64(0), float64(0), float64(0))

	var contentGroups []interface{}
	for i := 0; i < len(contentGroupNamesList); i++ {
		contentGroups = append(contentGroups, "none")
	}

	nonMatchingRow := []interface{}{"none"}
	nonMatchingRow = append(nonMatchingRow, contentGroups...)
	nonMatchingRow = append(nonMatchingRow, defaultMatchingRow...)

	// Add up for linkedEvents for conversion and conversion rate
	for i := 0; i < len(linkedEvents); i++ {
		nonMatchingRow = append(nonMatchingRow, float64(0), float64(0))
	}
	rows := make([][]interface{}, 0)
	for _, data := range *attributionData {
		attributionIdName := data.MarketingInfo.LandingPageUrl
		if attributionIdName == "" {
			attributionIdName = PropertyValueNone
		}
		if attributionIdName != "" {

			var row []interface{}
			// Add up Name
			row = append(row, attributionIdName)

			// Add up content Groups
			for i := 0; i < len(contentGroups); i++ {
				if v, exists := data.MarketingInfo.ContentGroupValuesMap[contentGroupNamesList[i]]; exists {
					row = append(row, v)
				} else {
					row = append(row, PropertyValueNone)
				}
			}
			// Validate if conversion values are not nil
			if data.ConversionEventCount == nil {
				emptyConversionEventRow := make([]float64, 0)
				emptyConversionEventRow = append(emptyConversionEventRow, 0.0)
				data.ConversionEventCount = emptyConversionEventRow
			}
			if data.ConversionEventCountInfluence == nil {
				emptyConversionEventRowInfluence := make([]float64, 0)
				emptyConversionEventRowInfluence = append(emptyConversionEventRowInfluence, 0.0)
				data.ConversionEventCountInfluence = emptyConversionEventRowInfluence
			}
			// Append fixed Metrics & ConversionEventCount[0] as only one goal event exists for landing page
			row = append(row, data.ConversionEventCount[0], data.ConversionEventCountInfluence[0])

			if isCompare && data.ConversionEventCompareCount != nil {

				row = append(row, data.ConversionEventCompareCount[0])
				row = append(row, data.ConversionEventCompareCountInfluence[0])
			} else {
				row = append(row, float64(0), float64(0))
			}
			row = append(row, getLinkedEventColumnAsInterfaceListLandingPagev1(data.LinkedEventsCount, data.LinkedEventsCountInfluence, len(linkedEvents))...)
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		// In case of empty result, send a row of zeros
		rows = append(rows, nonMatchingRow)
	}
	return rows
}

func ProcessQueryKPIPageUrl(query *AttributionQuery, attributionData *map[string]*AttributionData, logCtx log.Entry, kpiData map[string]KPIInfo, kpiHeaders []string, kpiAggFunctionType []string, isCompare bool) *QueryResult {
	logFields := log.Fields{"Method": "ProcessQueryKPILandingPageUrl"}
	logCtx = *logCtx.WithFields(logFields)

	var goalEvents []string
	goalEvents = kpiHeaders

	dataRows := GetRowsByMapsKPIPage(query.AttributionKey, query.AttributionContentGroups, attributionData, goalEvents, isCompare)
	logCtx.Info("Done GetRowsByMapsKPIPage")
	result := &QueryResult{}

	AddHeadersByAttributionKey(result, query, goalEvents, kpiAggFunctionType)

	result.Rows = dataRows

	// Update result based on Key Dimensions
	err := GetUpdatedRowsByDimensions(result, query)
	if err != nil {
		return nil
	}
	result.Rows = MergeDataRowsHavingSameKey(result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), query.AttributionKey, query.AnalyzeType, kpiAggFunctionType)
	// sort the rows by conversionEvent
	conversionIndex := GetConversionIndexKPI(result.Headers)
	sort.Slice(result.Rows, func(i, j int) bool {
		if len(result.Rows[i]) < conversionIndex || len(result.Rows[j]) < conversionIndex {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results are rows len mismatch. Ignoring row and continuing.")
			}
			return true
		}
		v1, ok1 := result.Rows[i][conversionIndex].(float64)
		v2, ok2 := result.Rows[j][conversionIndex].(float64)
		if !ok1 || !ok2 {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results cast mismatch. Ignoring row and continuing.")
			}
			return true
		}
		return v1 > v2
	})
	logCtx.Info("MergeDataRowsHavingSameKey")

	result.Rows = AddGrandTotalRowKPILandingPage(result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), goalEvents, query.AttributionMethodology, query.AttributionMethodologyCompare)
	logCtx.Info("Done AddGrandTotal")
	return result

}

// ProcessQueryKPILandingPageUrl converts attribution data into result
func ProcessQueryKPILandingPageUrl(query *AttributionQuery, attributionData *map[string]*AttributionData, logCtx log.Entry, kpiData map[string]KPIInfo, kpiHeaders []string, kpiAggFunctionType []string, isCompare bool) *QueryResult {
	logFields := log.Fields{"Method": "ProcessQueryKPILandingPageUrl"}
	logCtx = *logCtx.WithFields(logFields)

	var goalEvents []string
	goalEvents = kpiHeaders

	dataRows := GetRowsByMapsKPILandingPage(query.AttributionContentGroups, attributionData, goalEvents, isCompare)
	logCtx.Info("Done GetRowsByMapsKPILandingPage")
	result := &QueryResult{}

	AddHeadersByAttributionKey(result, query, goalEvents, kpiAggFunctionType)

	result.Rows = dataRows

	// Update result based on Key Dimensions
	err := GetUpdatedRowsByDimensions(result, query)
	if err != nil {
		return nil
	}
	result.Rows = MergeDataRowsHavingSameKey(result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), query.AttributionKey, query.AnalyzeType, kpiAggFunctionType)
	// sort the rows by conversionEvent
	conversionIndex := GetConversionIndexKPI(result.Headers)
	sort.Slice(result.Rows, func(i, j int) bool {
		if len(result.Rows[i]) < conversionIndex || len(result.Rows[j]) < conversionIndex {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results are rows len mismatch. Ignoring row and continuing.")
			}
			return true
		}
		v1, ok1 := result.Rows[i][conversionIndex].(float64)
		v2, ok2 := result.Rows[j][conversionIndex].(float64)
		if !ok1 || !ok2 {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results cast mismatch. Ignoring row and continuing.")
			}
			return true
		}
		return v1 > v2
	})
	logCtx.Info("MergeDataRowsHavingSameKey")

	result.Rows = AddGrandTotalRowKPILandingPage(result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), goalEvents, query.AttributionMethodology, query.AttributionMethodologyCompare)
	logCtx.Info("Done AddGrandTotal")
	return result

}

// ProcessQueryPageUrl converts attribution data into result
func ProcessQueryPageUrl(query *AttributionQuery, attributionData *map[string]*AttributionData, logCtx log.Entry, isCompare bool) *QueryResult {
	logFields := log.Fields{"Method": "ProcessQueryLandingPageUrl"}
	logCtx = *logCtx.WithFields(logFields)
	dataRows := GetRowsByMapsAllPage(query.AttributionContentGroups, attributionData, query.LinkedEvents, isCompare)

	result := &QueryResult{}
	AddHeadersByAttributionKey(result, query, nil, nil)

	result.Rows = dataRows

	// Update result based on Key Dimensions
	err := GetUpdatedRowsByDimensions(result, query)
	if err != nil {
		return nil
	}
	result.Rows = MergeDataRowsHavingSameKey(result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), query.AttributionKey, query.AnalyzeType, nil)
	// sort the rows by conversionEvent
	conversionIndex := GetConversionIndex(result.Headers)
	sort.Slice(result.Rows, func(i, j int) bool {
		if len(result.Rows[i]) < conversionIndex || len(result.Rows[j]) < conversionIndex {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results are rows len mismatch. Ignoring row and continuing.")
			}
			return true
		}
		v1, ok1 := result.Rows[i][conversionIndex].(float64)
		v2, ok2 := result.Rows[j][conversionIndex].(float64)
		if !ok1 || !ok2 {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results cast mismatch. Ignoring row and continuing.")
			}
			return true
		}
		return v1 > v2
	})

	result.Rows = AddGrandTotalRowLandingPage(result.Headers, result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), query.AttributionMethodology, query.AttributionMethodologyCompare)
	return result

}

// ProcessQueryLandingPageUrl converts attribution data into result
func ProcessQueryLandingPageUrl(query *AttributionQuery, attributionData *map[string]*AttributionData, logCtx log.Entry, isCompare bool) *QueryResult {
	logFields := log.Fields{"Method": "ProcessQueryLandingPageUrl"}
	logCtx = *logCtx.WithFields(logFields)
	dataRows := GetRowsByMapsLandingPage(query.AttributionContentGroups, attributionData, query.LinkedEvents, isCompare)

	result := &QueryResult{}
	AddHeadersByAttributionKey(result, query, nil, nil)

	result.Rows = dataRows

	// Update result based on Key Dimensions
	err := GetUpdatedRowsByDimensions(result, query)
	if err != nil {
		return nil
	}
	result.Rows = MergeDataRowsHavingSameKey(result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), query.AttributionKey, query.AnalyzeType, nil)
	// sort the rows by conversionEvent
	conversionIndex := GetConversionIndex(result.Headers)
	sort.Slice(result.Rows, func(i, j int) bool {
		if len(result.Rows[i]) < conversionIndex || len(result.Rows[j]) < conversionIndex {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results are rows len mismatch. Ignoring row and continuing.")
			}
			return true
		}
		v1, ok1 := result.Rows[i][conversionIndex].(float64)
		v2, ok2 := result.Rows[j][conversionIndex].(float64)
		if !ok1 || !ok2 {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results cast mismatch. Ignoring row and continuing.")
			}
			return true
		}
		return v1 > v2
	})

	result.Rows = AddGrandTotalRowLandingPage(result.Headers, result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), query.AttributionMethodology, query.AttributionMethodologyCompare)
	return result

}

// ProcessQuery converts attribution data into result
func ProcessQuery(query *AttributionQuery, attributionData *map[string]*AttributionData, marketingReports *MarketingReports, isCompare bool, logCtx log.Entry) *QueryResult {
	logFields := log.Fields{"Method": "ProcessQuery"}
	logCtx = *logCtx.WithFields(logFields)

	var goalEvents []string
	goalEvents = append(goalEvents, query.ConversionEvent.Name)

	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"attributionData": attributionData}).Info(" attributionData before ProcessQuery")
	}
	for key, _ := range *attributionData {
		//add key to attribution data
		addKeyToMarketingInfoForChannelOrSource(attributionData, key, query)
	}
	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"attributionData": attributionData}).Info(" attributionData after  addKeyToMarketingInfoForChannelOrSource")
	}
	// Add additional metrics values
	ComputeAdditionalMetrics(attributionData)
	// Add custom dimensions
	AddCustomDimensions(attributionData, query, marketingReports)

	// Attribution data to rows
	dataRows := GetRowsByMaps(query.AttributionKey, query.AttributionKeyCustomDimension, attributionData, query.LinkedEvents, goalEvents, isCompare)

	result := &QueryResult{}
	AddHeadersByAttributionKey(result, query, nil, nil)
	result.Rows = dataRows
	// get the headers for KPI
	var goalEventAggFuncTypes []string
	for _, value := range *attributionData {
		goalEventAggFuncTypes = value.ConvAggFunctionType
		break
	}
	// Update result based on Key Dimensions
	err := GetUpdatedRowsByDimensions(result, query)
	if err != nil {
		return nil
	}

	result.Rows = MergeDataRowsHavingSameKey(result.Rows, GetLastKeyValueIndex(result.Headers), query.AttributionKey, query.AnalyzeType, goalEventAggFuncTypes)

	// sort the rows by conversionEvent
	conversionIndex := GetConversionIndex(result.Headers)
	sort.Slice(result.Rows, func(i, j int) bool {
		if len(result.Rows[i]) < conversionIndex || len(result.Rows[j]) < conversionIndex {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results are rows len mismatch. Ignoring row and continuing.")
			}
			return true
		}
		v1, ok1 := result.Rows[i][conversionIndex].(float64)
		v2, ok2 := result.Rows[j][conversionIndex].(float64)
		if !ok1 || !ok2 {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results cast mismatch. Ignoring row and continuing.")
			}
			return true
		}
		return v1 > v2
	})

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"result": result}).Info("result before addGrandTotal")
	}
	result.Rows = AddGrandTotalRow(result.Headers, result.Rows, GetLastKeyValueIndex(result.Headers),
		query.AttributionMethodology, query.AttributionMethodologyCompare)
	return result
}

// addKeyToMarketingInfo add keys to marketing info in attributionData
func addKeyToMarketingInfoForChannelOrSource(attributionData *map[string]*AttributionData, key string, query *AttributionQuery) {

	if query.AttributionKey == AttributionKeySource {
		(*attributionData)[key].MarketingInfo.Source = key
	} else if query.AttributionKey == AttributionKeyChannel {
		(*attributionData)[key].MarketingInfo.Channel = key
	}

}

// addKeyToMarketingInfoForChannelOrSourceV1 add keys to marketing info in attributionData
func addKeyToMarketingInfoForChannelOrSourceV1(attributionData *map[string]*AttributionData, key string, query *AttributionQueryV1) {

	if query.AttributionKey == AttributionKeySource {
		(*attributionData)[key].MarketingInfo.Source = key
	} else if query.AttributionKey == AttributionKeyChannel {
		(*attributionData)[key].MarketingInfo.Channel = key
	}

}

// ProcessQueryKPI converts attribution data into result
func ProcessQueryKPI(query *AttributionQuery, attributionData *map[string]*AttributionData,
	marketingReports *MarketingReports, isCompare bool, kpiData map[string]KPIInfo, kpiHeaders []string, kpiAggFunctionType []string) *QueryResult {

	logCtx := log.WithFields(log.Fields{"Method": "ProcessQueryKPI", "KPIAttribution": "Debug"})

	// get the headers for KPI
	var goalEvents []string
	goalEvents = kpiHeaders

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "attributionData": attributionData}).Info("KPI Attribution data 2")
	}
	for key, _ := range *attributionData {
		//add key to attribution data
		addKeyToMarketingInfoForChannelOrSource(attributionData, key, query)
	}

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "attributionData": attributionData}).Info("KPI Attribution data 3")
	}
	// Add additional metrics values
	ComputeAdditionalMetrics(attributionData)

	// Add custom dimensions
	AddCustomDimensions(attributionData, query, marketingReports)

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "attributionData": attributionData}).Info("Done AddTheAddedKeysAndMetrics AddPerformanceData ApplyFilter ComputeAdditionalMetrics AddCustomDimensions")
	}

	for key, _ := range *attributionData {
		(*attributionData)[key].ConvAggFunctionType = kpiAggFunctionType
	}

	// Attribution data to rows
	dataRows := GetRowsByMaps(query.AttributionKey, query.AttributionKeyCustomDimension, attributionData, query.LinkedEvents, goalEvents, isCompare)
	result := &QueryResult{}

	AddHeadersByAttributionKey(result, query, goalEvents, kpiAggFunctionType)
	result.Rows = dataRows
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "result": result}).Info("Done GetRowsByMaps AddHeadersByAttributionKey")
	}
	// Update result based on Key Dimensions
	err := GetUpdatedRowsByDimensions(result, query)
	if err != nil {
		return nil
	}

	result.Rows = MergeDataRowsHavingSameKeyKPI(result.Rows, GetLastKeyValueIndex(result.Headers), query.AttributionKey, query.AnalyzeType, kpiAggFunctionType, *logCtx)
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "Result": result}).Info("KPI Attribution result")
	}

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "Result": result}).Info("KPI Attribution result")
	}

	// sort the rows by conversionEvent
	conversionIndex := GetConversionIndexKPI(result.Headers)
	sort.Slice(result.Rows, func(i, j int) bool {
		if len(result.Rows[i]) < conversionIndex || len(result.Rows[j]) < conversionIndex {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results are rows len mismatch. Ignoring row and continuing.")
			}
			return true
		}
		if len(result.Rows[i]) > conversionIndex && len(result.Rows[j]) > conversionIndex {
			v1, ok1 := result.Rows[i][conversionIndex].(float64)
			v2, ok2 := result.Rows[j][conversionIndex].(float64)
			if !ok1 || !ok2 {
				if C.GetAttributionDebug() == 1 {
					logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results cast mismatch. Ignoring row and continuing.")
				}
				return true
			}
			return v1 > v2

		} else {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "RowI": result.Rows[i], "RowJ": result.Rows[j]}).Info("Bad row in Sorting")
			}
		}
		return true
	})

	result.Rows = AddGrandTotalRowKPI(result.Rows, GetLastKeyValueIndex(result.Headers), kpiAggFunctionType, query.AttributionMethodology, query.AttributionMethodologyCompare)

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "Result": result}).Info("KPI Attribution result AddGrandTotalRow")
	}
	return result
}

// ProcessQueryUserKPI converts attribution data into result
func ProcessQueryUserKPI(query *AttributionQuery, attributionData *map[string]*AttributionData,
	marketingReports *MarketingReports, isCompare bool, kpiData map[string]KPIInfo, kpiHeaders []string, kpiAggFunctionType []string) *QueryResult {

	logCtx := log.WithFields(log.Fields{"Method": "ProcessQueryKPI", "KPIAttribution": "Debug"})

	logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "attributionData": attributionData}).Info("KPI Attribution data")

	// get the headers for KPI
	var goalEvents []string
	goalEvents = kpiHeaders

	for key, _ := range *attributionData {
		//add key to attribution data
		addKeyToMarketingInfoForChannelOrSource(attributionData, key, query)
	}

	// Add additional metrics values
	ComputeAdditionalMetrics(attributionData)

	// Add custom dimensions
	AddCustomDimensions(attributionData, query, marketingReports)

	logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "attributionData": attributionData}).Info("Done AddTheAddedKeysAndMetrics AddPerformanceData ApplyFilter ComputeAdditionalMetrics AddCustomDimensions")

	for key, _ := range *attributionData {
		(*attributionData)[key].ConvAggFunctionType = kpiAggFunctionType
	}

	// Attribution data to rows
	dataRows := GetRowsByMaps(query.AttributionKey, query.AttributionKeyCustomDimension, attributionData, query.LinkedEvents, goalEvents, isCompare)
	result := &QueryResult{}

	AddHeadersByAttributionKey(result, query, goalEvents, kpiAggFunctionType)
	result.Rows = dataRows
	logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "result": result}).Info("Done GetRowsByMaps AddHeadersByAttributionKey")
	// Update result based on Key Dimensions
	err := GetUpdatedRowsByDimensions(result, query)
	if err != nil {
		return nil
	}

	result.Rows = MergeDataRowsHavingSameKeyKPI(result.Rows, GetLastKeyValueIndex(result.Headers), query.AttributionKey, query.AnalyzeType, kpiAggFunctionType, *logCtx)

	logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "Result": result}).Info("KPI Attribution result")

	logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "Result": result}).Info("KPI Attribution result")

	// sort the rows by conversionEvent
	conversionIndex := GetConversionIndexKPI(result.Headers)
	sort.Slice(result.Rows, func(i, j int) bool {
		if len(result.Rows[i]) < conversionIndex || len(result.Rows[j]) < conversionIndex {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results are rows len mismatch. Ignoring row and continuing.")
			}
			return true
		}
		if len(result.Rows[i]) > conversionIndex && len(result.Rows[j]) > conversionIndex {
			v1, ok1 := result.Rows[i][conversionIndex].(float64)
			v2, ok2 := result.Rows[j][conversionIndex].(float64)
			if !ok1 || !ok2 {
				if C.GetAttributionDebug() == 1 {
					logCtx.WithFields(log.Fields{"row1": result.Rows[i], "row2": result.Rows[j]}).Info("final results cast mismatch. Ignoring row and continuing.")
				}
				return true
			}
			return v1 > v2

		} else {
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "RowI": result.Rows[i], "RowJ": result.Rows[j]}).Info("Bad row in Sorting")
			}
		}
		return true
	})

	result.Rows = AddGrandTotalRowKPI(result.Rows, GetLastKeyValueIndex(result.Headers), kpiAggFunctionType, query.AttributionMethodology, query.AttributionMethodologyCompare)

	logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "Result": result}).Info("KPI Attribution result AddGrandTotalRow")

	return result
}

// GetUpdatedRowsByDimensions updated the granular result with reduced dimensions
func GetUpdatedRowsByDimensions(result *QueryResult, query *AttributionQuery) error {

	validHeadersDimensions := make(map[string]int)
	for _, val := range query.AttributionKeyDimension {
		if _, exists := KeyDimensionToHeaderMap[val]; !exists {
			return errors.New("couldn't find the header value for given dimensions value")
		}
		validHeadersDimensions[KeyDimensionToHeaderMap[val]] = 1
	}

	addedKeysSize := GetKeyIndexOrAddedKeySize(query.AttributionKey)
	// Build new header
	var newHeaders []string
	for j, field := range result.Headers {
		// filter out the Added Key Dimensions if reduced
		if j <= addedKeysSize && validHeadersDimensions[field] == 0 {
			continue
		}
		newHeaders = append(newHeaders, field)
	}

	// Build new row
	newRows := make([][]interface{}, 0)
	for _, data := range result.Rows {
		var row []interface{}
		if len(result.Headers) > len(data) {
			for i := len(data); i < len(result.Headers); i++ {
				data = append(data, float64(0))
			}
		}
		for j, field := range result.Headers {
			// filter out the Added Key Dimensions if reduced
			if j <= addedKeysSize && validHeadersDimensions[field] == 0 {
				continue
			}
			row = append(row, data[j])
		}
		newRows = append(newRows, row)
	}

	result.Headers = newHeaders
	result.Rows = newRows
	return nil
}

// MergeTwoDataRows adds values of two data rows
func MergeTwoDataRows(row1 []interface{}, row2 []interface{}, keyIndex int, attributionKey string, analyzeType string, conversionFunTypes []string) []interface{} {

	if attributionKey == AttributionKeyLandingPage || attributionKey == AttributionKeyAllPageView {

		row1[keyIndex+1] = row1[keyIndex+1].(float64) + row2[keyIndex+1].(float64) // Conversion.
		row1[keyIndex+2] = row1[keyIndex+2].(float64) + row2[keyIndex+2].(float64) // Conversion Influence
		row1[keyIndex+3] = row1[keyIndex+3].(float64) + row2[keyIndex+3].(float64) // Compare conversion
		row1[keyIndex+4] = row1[keyIndex+4].(float64) + row2[keyIndex+4].(float64) // Compare conversion-Influence

		// Remaining linked funnel events & CPCs
		for i := keyIndex + 5; i < len(row1)-1; i += 2 {
			row1[i] = row1[i].(float64) + row2[i].(float64)
			row1[i+1] = row1[i+1].(float64) + row2[i+1].(float64)
		}

		return row1
	} else if analyzeType == AnalyzeTypeHSDeals || analyzeType == AnalyzeTypeSFOpportunities || analyzeType == AnalyzeTypeUserKPI {

		row1[keyIndex+1] = row1[keyIndex+1].(int64) + row2[keyIndex+1].(int64)     // Impressions.
		row1[keyIndex+2] = row1[keyIndex+2].(int64) + row2[keyIndex+2].(int64)     // Clicks.
		row1[keyIndex+3] = row1[keyIndex+3].(float64) + row2[keyIndex+3].(float64) // Spend.

		for idx, _ := range conversionFunTypes {
			nextConPosition := idx * 6
			row1[keyIndex+8+nextConPosition] = row1[keyIndex+8+nextConPosition].(float64) + row2[keyIndex+8+nextConPosition].(float64)    // Conversion.
			row1[keyIndex+9+nextConPosition] = row1[keyIndex+9+nextConPosition].(float64) + row2[keyIndex+9+nextConPosition].(float64)    // Conversion Influence - values same as Linear Touch
			row1[keyIndex+11+nextConPosition] = row1[keyIndex+11+nextConPosition].(float64) + row2[keyIndex+11+nextConPosition].(float64) // Compare Conversion.
			row1[keyIndex+12+nextConPosition] = row1[keyIndex+12+nextConPosition].(float64) + row2[keyIndex+12+nextConPosition].(float64) // Compare Conversion Influence - values same as Linear Touch
		}
		impressions := (row1[keyIndex+1]).(int64)
		clicks := (row1[keyIndex+2]).(int64)
		spend := row1[keyIndex+3].(float64)

		if float64(impressions) > 0 {
			row1[keyIndex+4], _ = U.FloatRoundOffWithPrecision(100*float64(clicks)/float64(impressions), U.DefaultPrecision) // CTR.
			row1[keyIndex+6], _ = U.FloatRoundOffWithPrecision(1000*float64(spend)/float64(impressions), U.DefaultPrecision) // CPM.
		} else {
			row1[keyIndex+4] = float64(0) // CTR.
			row1[keyIndex+6] = float64(0) // CPM.
		}
		if float64(clicks) > 0 {
			row1[keyIndex+5], _ = U.FloatRoundOffWithPrecision(float64(spend)/float64(clicks), U.DefaultPrecision)                          // AvgCPC.
			row1[keyIndex+7], _ = U.FloatRoundOffWithPrecision(100*float64(row1[keyIndex+8].(float64))/float64(clicks), U.DefaultPrecision) // ClickConversionRate.
		} else {
			row1[keyIndex+5] = float64(0) // AvgCPC.
			row1[keyIndex+7] = float64(0) // ClickConversionRate.
		}

		for idx, funcType := range conversionFunTypes {
			nextConPosition := idx * 6
			// Normal conversion [8, 9,10] = [Conversion,Conversion Influence, CPC]
			// Compare conversion [11, 12,13]  = [Conversion,Conversion Influence, CPC, Rate+nextConPosition]
			if strings.ToLower(funcType) == "sum" {

				if spend > 0 {
					row1[keyIndex+10+nextConPosition], _ = U.FloatRoundOffWithPrecision(row1[keyIndex+8+nextConPosition].(float64)/spend, U.DefaultPrecision) // Conversion - CPC.
				} else {
					row1[keyIndex+10+nextConPosition] = float64(0) // Conversion - CPC.
				}

				if spend > 0 {
					row1[keyIndex+13+nextConPosition], _ = U.FloatRoundOffWithPrecision(row1[keyIndex+11+nextConPosition].(float64)/spend, U.DefaultPrecision) // Compare Conversion - CPC.
				} else {
					row1[keyIndex+13+nextConPosition] = float64(0) // Compare Conversion - CPC.
				}

			} else {

				if row1[keyIndex+8+nextConPosition].(float64) > 0 {
					row1[keyIndex+10+nextConPosition], _ = U.FloatRoundOffWithPrecision(spend/row1[keyIndex+8+nextConPosition].(float64), U.DefaultPrecision) // Conversion - CPC.
				} else {
					row1[keyIndex+10+nextConPosition] = float64(0) // Conversion - CPC.
				}

				if row1[keyIndex+11+nextConPosition].(float64) > 0 {
					row1[keyIndex+13+nextConPosition], _ = U.FloatRoundOffWithPrecision(spend/row1[keyIndex+11+nextConPosition].(float64), U.DefaultPrecision) // Compare Conversion - CPC.
				} else {
					row1[keyIndex+13+nextConPosition] = float64(0) // Compare Conversion - CPC.
				}
			}
		}
		return row1

	} else {

		row1[keyIndex+1] = row1[keyIndex+1].(int64) + row2[keyIndex+1].(int64)     // Impressions.
		row1[keyIndex+2] = row1[keyIndex+2].(int64) + row2[keyIndex+2].(int64)     // Clicks.
		row1[keyIndex+3] = row1[keyIndex+3].(float64) + row2[keyIndex+3].(float64) // Spend.

		row1[keyIndex+8] = row1[keyIndex+8].(float64) + row2[keyIndex+8].(float64)    // Conversion.
		row1[keyIndex+9] = row1[keyIndex+9].(float64) + row2[keyIndex+9].(float64)    // Conversion Influence - values same as Linear Touch
		row1[keyIndex+11] = row1[keyIndex+11].(float64) + row2[keyIndex+11].(float64) // Compare Conversion.
		row1[keyIndex+12] = row1[keyIndex+12].(float64) + row2[keyIndex+12].(float64) // Compare Conversion Influence - values same as Linear Touch

		impressions := (row1[keyIndex+1]).(int64)
		clicks := (row1[keyIndex+2]).(int64)
		spend := row1[keyIndex+3].(float64)
		if float64(impressions) > 0 {
			row1[keyIndex+4], _ = U.FloatRoundOffWithPrecision(100*float64(clicks)/float64(impressions), U.DefaultPrecision) // CTR.
			row1[keyIndex+6], _ = U.FloatRoundOffWithPrecision(1000*float64(spend)/float64(impressions), U.DefaultPrecision) // CPM.
		} else {
			row1[keyIndex+4] = float64(0) // CTR.
			row1[keyIndex+6] = float64(0) // CPM.
		}
		if float64(clicks) > 0 {
			row1[keyIndex+5], _ = U.FloatRoundOffWithPrecision(float64(spend)/float64(clicks), U.DefaultPrecision)                          // AvgCPC.
			row1[keyIndex+7], _ = U.FloatRoundOffWithPrecision(100*float64(row1[keyIndex+8].(float64))/float64(clicks), U.DefaultPrecision) // ClickConversionRate.
		} else {
			row1[keyIndex+5] = float64(0) // AvgCPC.
			row1[keyIndex+7] = float64(0) // ClickConversionRate.
		}
		// Normal conversion [8, 9,10] = [Conversion,Conversion Influence, CPC]
		if row1[keyIndex+8].(float64) > 0 {
			row1[keyIndex+10], _ = U.FloatRoundOffWithPrecision(spend/row1[keyIndex+8].(float64), U.DefaultPrecision) // Conversion - CPC.
		} else {
			row1[keyIndex+10] = float64(0) // Conversion - CPC.
		}

		// Compare conversion [11,12,13] = [Conversion,Conversion Influence, CPC]
		if row1[keyIndex+11].(float64) > 0 {
			row1[keyIndex+13], _ = U.FloatRoundOffWithPrecision(spend/row1[keyIndex+11].(float64), U.DefaultPrecision) // Compare Conversion - CPC.
		} else {
			row1[keyIndex+13] = float64(0) // Compare Conversion - CPC.
		}

		// Remaining linked funnel events & CPCs
		for i := keyIndex + 14; i < len(row1)-1; i += 3 {
			// Conversion
			row1[i] = row1[i].(float64) + row2[i].(float64)
			// Influence
			row1[i+1] = row1[i+1].(float64) + row2[i+1].(float64)

			if row1[i].(float64) > 0 && i < len(row1) {
				row1[i+2], _ = U.FloatRoundOffWithPrecision(spend/row1[i].(float64), U.DefaultPrecision) // Funnel - Conversion - CPC. spend/conversion
			} else {
				row1[i+2] = float64(0) // Funnel - Conversion - CPC.
			}
		}
		return row1
	}
}

// SanitizeResult removes unwanted headers which are marked by (remove).
// Ex. For KPIs like Revenue/Pipeline, User conversion rate is not needed.
// It also updates the channel names.
func SanitizeResult(result *QueryResult) {

	// Populating the valid index
	var validIdx []int
	for idx, colName := range result.Headers {
		if !strings.Contains(colName, "(remove)") && !strings.Contains(colName, "InfluenceRemove") {
			validIdx = append(validIdx, idx)
		}
	}

	// Building new headers
	var resultHeader []string
	for _, val := range validIdx {
		resultHeader = append(resultHeader, result.Headers[val])
	}

	// Building new rows
	resultRows := make([][]interface{}, 0)
	for _, row := range result.Rows {

		for len(row) < len(result.Headers) {
			row = append(row, float64(0))
		}
		resultRow := make([]interface{}, 0)
		for _, val := range validIdx {

			resultRow = append(resultRow, row[val])
		}
		resultRows = append(resultRows, resultRow)
	}

	result.Headers = resultHeader
	result.Rows = resultRows
	UpdateChannelName(result, "adwords", "Google Ads")
}

// UpdateChannelName updates the original channel names with the given  channel name
func UpdateChannelName(result *QueryResult, originalName string, newName string) {

	hasChannelName := false
	var channelNameIdx int
	for idx, colName := range result.Headers {
		if colName == "ChannelName" {
			channelNameIdx = idx
			hasChannelName = true
			break
		}
	}

	if !hasChannelName {
		return
	}

	for _, row := range result.Rows {

		if row[channelNameIdx] == originalName {
			row[channelNameIdx] = newName
		}

	}
}

// SanitizeResultForSourceAndChannel removes marketing metrics for source and channel level report
func SanitizeResultForSourceAndChannel(result *QueryResult) {

	// Populating the valid index
	var validIdx []int
	for idx, colName := range result.Headers {
		if colName != "Impressions" && colName != "Clicks" && colName != "Spend" &&
			colName != "CTR(%)" && colName != "Average CPC" && colName != "CPM" && colName != "ClickConversionRate(%)" &&
			!strings.Contains(colName, "Cost Per Conversion") && !strings.Contains(colName, "Return on Cost") {
			validIdx = append(validIdx, idx)
		}
	}

	// Building new headers
	var resultHeader []string
	for _, val := range validIdx {
		resultHeader = append(resultHeader, result.Headers[val])
	}

	// Building new rows
	resultRows := make([][]interface{}, 0)
	for _, row := range result.Rows {

		for len(row) < len(result.Headers) {
			row = append(row, float64(0))
		}
		resultRow := make([]interface{}, 0)
		for _, val := range validIdx {

			resultRow = append(resultRow, row[val])
		}
		resultRows = append(resultRows, resultRow)
	}

	result.Headers = resultHeader
	result.Rows = resultRows
}

// MergeDataRowsHavingSameKey merges rows having same key by adding each column value
func MergeDataRowsHavingSameKey(rows [][]interface{}, keyIndex int, attributionKey string, analyzeType string, conversionFunTypes []string) [][]interface{} {

	rowKeyMap := make(map[string][]interface{})
	maxRowSize := 0
	for _, row := range rows {
		maxRowSize = U.MaxInt(len(row), maxRowSize)
		if len(row) == 0 || len(row) != maxRowSize {
			continue
		}
		// creating a key for using added keys and index
		key := ""
		for j := 0; j <= keyIndex; j++ {
			val, ok := row[j].(string)
			// Ignore row if key is not proper
			if !ok {
				continue
			}
			key = key + val
		}
		if _, exists := rowKeyMap[key]; exists {
			rowKeyMap[key] = MergeTwoDataRows(rowKeyMap[key], row, keyIndex, attributionKey, analyzeType, conversionFunTypes)
		} else {
			rowKeyMap[key] = row
		}
	}
	resultRows := make([][]interface{}, 0)
	for _, mapRow := range rowKeyMap {
		resultRows = append(resultRows, mapRow)
	}
	return resultRows
}

// AddGrandTotalRow adds a row with grand total in report
func AddGrandTotalRow(headers []string, rows [][]interface{}, keyIndex int, method string, methodCompare string) [][]interface{} {

	var grandTotalRow []interface{}

	for j := 0; j <= keyIndex; j++ {
		grandTotalRow = append(grandTotalRow, "Grand Total")
	}
	// Name, impression, clicks, spend
	defaultMatchingRow := []interface{}{int64(0), int64(0), float64(0),
		// (CTR, AvgCPC, CPM, ClickConversionRate)
		float64(0), float64(0), float64(0), float64(0),
		// ConversionEventCount, ConversionEventCountInfluence,CostPerConversion, ConversionEventCompareCount, ConversionEventCompareCountInfluence,CostPerConversionCompareCount
		float64(0), float64(0), float64(0), float64(0), float64(0), float64(0)}

	grandTotalRow = append(grandTotalRow, defaultMatchingRow...)
	// Remaining linked funnel events & CPCs
	for i := keyIndex + 14; i < len(headers)-1; i++ {
		grandTotalRow = append(grandTotalRow, float64(0))
	}

	clicksCTR := int64(0)      //4
	impressionsCTR := int64(0) //4

	conversionsClickConversionRate := float64(0) //7
	clicksClickConversionRate := int64(0)        //7

	spendAvgCPC := float64(0) //5
	clickAvgCPC := int64(0)   //5

	spendCPM := float64(0)     //6
	impressionsCPM := int64(0) //6

	maxRowSize := 0
	for _, row := range rows {

		maxRowSize = U.MaxInt(len(row), maxRowSize)
		if len(row) == 0 || len(row) != maxRowSize {
			continue
		}

		grandTotalRow[keyIndex+1] = grandTotalRow[keyIndex+1].(int64) + row[keyIndex+1].(int64)                                                        // Impressions.
		grandTotalRow[keyIndex+2] = grandTotalRow[keyIndex+2].(int64) + row[keyIndex+2].(int64)                                                        // Clicks.
		grandTotalRow[keyIndex+3], _ = U.FloatRoundOffWithPrecision(grandTotalRow[keyIndex+3].(float64)+row[keyIndex+3].(float64), U.DefaultPrecision) // Spend.

		grandTotalRow[keyIndex+8] = grandTotalRow[keyIndex+8].(float64) + row[keyIndex+8].(float64)    // Conversion.
		grandTotalRow[keyIndex+9] = grandTotalRow[keyIndex+9].(float64) + row[keyIndex+9].(float64)    // Conversion Influence - values same as Linear Touch
		grandTotalRow[keyIndex+11] = grandTotalRow[keyIndex+11].(float64) + row[keyIndex+11].(float64) // Compare Conversion.
		grandTotalRow[keyIndex+12] = grandTotalRow[keyIndex+12].(float64) + row[keyIndex+12].(float64) // Compare Conversion Influence - values same as Linear Touch

		impressions := (row[keyIndex+1]).(int64)
		clicks := (row[keyIndex+2]).(int64)
		spend := row[keyIndex+3].(float64)

		if impressions > 0 {
			clicksCTR = clicksCTR + clicks
			impressionsCTR = impressionsCTR + impressions
			spendCPM = spendCPM + spend
			impressionsCPM = impressionsCPM + impressions
		}

		if clicks > 0 {
			spendAvgCPC = spendAvgCPC + spend
			clickAvgCPC = clickAvgCPC + clicks
			conversionsClickConversionRate = conversionsClickConversionRate + (row[keyIndex+8]).(float64)
			clicksClickConversionRate = clicksClickConversionRate + clicks
		}

		// Remaining linked funnel events & CPCs
		for i := keyIndex + 14; i < len(grandTotalRow)-1; i += 3 {
			grandTotalRow[i] = grandTotalRow[i].(float64) + row[i].(float64)       //LFE Conversion Count
			grandTotalRow[i+1] = grandTotalRow[i+1].(float64) + row[i+1].(float64) // LFE Conversion Influence

		}
		// If attribution method is influence then replacing the value of grand total row of conversion with conversion Influence
		if method == AttributionMethodInfluence {
			grandTotalRow[keyIndex+8] = grandTotalRow[keyIndex+9]
			for i := keyIndex + 14; i < len(grandTotalRow)-1; i += 3 {
				grandTotalRow[i] = grandTotalRow[i+1]

			}
		}
		if methodCompare == AttributionMethodInfluence {
			grandTotalRow[keyIndex+11] = grandTotalRow[keyIndex+12]
		}

	}

	// CTR(%)
	if float64(impressionsCTR) > 0 {
		grandTotalRow[keyIndex+4], _ = U.FloatRoundOffWithPrecision(float64(100*float64(clicksCTR)/float64(impressionsCTR)), U.DefaultPrecision)
	} else {
		grandTotalRow[keyIndex+4] = float64(0)
	}

	// clickAvgCPC
	if float64(clickAvgCPC) > 0 {
		grandTotalRow[keyIndex+5], _ = U.FloatRoundOffWithPrecision(float64(spendAvgCPC)/float64(clickAvgCPC), U.DefaultPrecision)
	} else {
		grandTotalRow[keyIndex+5] = float64(0)
	}

	// CPM
	if float64(impressionsCPM) > 0 {
		grandTotalRow[keyIndex+6], _ = U.FloatRoundOffWithPrecision(float64(1000*float64(spendCPM))/float64(impressionsCPM), U.DefaultPrecision)
	} else {
		grandTotalRow[keyIndex+6] = float64(0)
	}

	// clicksClickConversionRate
	if float64(clicksClickConversionRate) > 0 {
		grandTotalRow[keyIndex+7], _ = U.FloatRoundOffWithPrecision(100*float64(conversionsClickConversionRate)/float64(clicksClickConversionRate), U.DefaultPrecision)
	} else {
		grandTotalRow[keyIndex+7] = float64(0)
	}

	// Conversion - CPC.
	if grandTotalRow[keyIndex+8].(float64) > 0 {
		grandTotalRow[keyIndex+10], _ = U.FloatRoundOffWithPrecision(grandTotalRow[keyIndex+3].(float64)/grandTotalRow[keyIndex+8].(float64), U.DefaultPrecision)
	} else {
		grandTotalRow[keyIndex+10] = float64(0)
	}

	// Compare Conversion - CPC.
	if grandTotalRow[keyIndex+11].(float64) > 0 {
		grandTotalRow[keyIndex+13], _ = U.FloatRoundOffWithPrecision(grandTotalRow[keyIndex+3].(float64)/grandTotalRow[keyIndex+11].(float64), U.DefaultPrecision)
	} else {
		grandTotalRow[keyIndex+13] = float64(0)
	}

	// Remaining linked funnel events & CPCs
	for i := keyIndex + 14; i < len(grandTotalRow)-1; i += 3 {
		if grandTotalRow[i].(float64) > 0 {
			// Funnel - Conversion - CPC.
			grandTotalRow[i+2], _ = U.FloatRoundOffWithPrecision(grandTotalRow[keyIndex+3].(float64)/grandTotalRow[i].(float64), U.DefaultPrecision)
		} else {
			// Funnel - Conversion - CPC.
			grandTotalRow[i+2] = float64(0)
		}

	}

	// concatenated key
	grandTotalRow = append(grandTotalRow, "Grand Total")

	rows = append([][]interface{}{grandTotalRow}, rows...)

	return rows

}

// AddGrandTotalRowKPI adds a row with grand total in report for KPI queries
func AddGrandTotalRowKPI(rows [][]interface{}, keyIndex int, conversionFunTypes []string, method string, methodCompare string) [][]interface{} {

	var grandTotalRow []interface{}

	for j := 0; j <= keyIndex; j++ {
		grandTotalRow = append(grandTotalRow, "Grand Total")
	}
	// Name, impression, clicks, spend
	defaultMatchingRow := []interface{}{int64(0), int64(0), float64(0),
		// (CTR, AvgCPC, CPM, ClickConversionRate)
		float64(0), float64(0), float64(0), float64(0)}

	for idx := 0; idx < len(conversionFunTypes); idx++ {
		// one for each - ConversionEventCount, ConversionEventCountInfluence,CostPerConversion,  ConversionEventCompareCount, ConversionEventCompareCountInfluence, CostPerConversionCompareCount
		defaultMatchingRow = append(defaultMatchingRow, float64(0), float64(0), float64(0), float64(0), float64(0), float64(0))
	}

	grandTotalRow = append(grandTotalRow, defaultMatchingRow...)

	clicksCTR := int64(0)      //4
	impressionsCTR := int64(0) //4

	conversionsClickConversionRate := float64(0) //7
	clicksClickConversionRate := int64(0)        //7

	spendAvgCPC := float64(0) //5
	clickAvgCPC := int64(0)   //5

	spendCPM := float64(0)     //6
	impressionsCPM := int64(0) //6

	maxRowSize := 0
	for _, row := range rows {

		maxRowSize = U.MaxInt(len(row), maxRowSize)
		if len(row) == 0 || len(row) != maxRowSize {
			continue
		}

		grandTotalRow[keyIndex+1] = grandTotalRow[keyIndex+1].(int64) + row[keyIndex+1].(int64)                                                        // Impressions.
		grandTotalRow[keyIndex+2] = grandTotalRow[keyIndex+2].(int64) + row[keyIndex+2].(int64)                                                        // Clicks.
		grandTotalRow[keyIndex+3], _ = U.FloatRoundOffWithPrecision(grandTotalRow[keyIndex+3].(float64)+row[keyIndex+3].(float64), U.DefaultPrecision) // Spend.

		impressions := (row[keyIndex+1]).(int64)
		clicks := (row[keyIndex+2]).(int64)
		spend := row[keyIndex+3].(float64)

		if impressions > 0 {
			clicksCTR = clicksCTR + clicks
			impressionsCTR = impressionsCTR + impressions
			spendCPM = spendCPM + spend
			impressionsCPM = impressionsCPM + impressions
		}

		if clicks > 0 {
			spendAvgCPC = spendAvgCPC + spend
			clickAvgCPC = clickAvgCPC + clicks
			conversionsClickConversionRate = conversionsClickConversionRate + (row[keyIndex+8]).(float64)
			clicksClickConversionRate = clicksClickConversionRate + clicks
		}

		for idx, _ := range conversionFunTypes {
			nextConPosition := idx * 6
			grandTotalRow[keyIndex+8+nextConPosition] = grandTotalRow[keyIndex+8+nextConPosition].(float64) + row[keyIndex+8+nextConPosition].(float64)    // Conversion.
			grandTotalRow[keyIndex+9+nextConPosition] = grandTotalRow[keyIndex+9+nextConPosition].(float64) + row[keyIndex+9+nextConPosition].(float64)    // Conversion Influence - values same as Linear Touch
			grandTotalRow[keyIndex+11+nextConPosition] = grandTotalRow[keyIndex+11+nextConPosition].(float64) + row[keyIndex+11+nextConPosition].(float64) // Compare Conversion.
			grandTotalRow[keyIndex+12+nextConPosition] = grandTotalRow[keyIndex+12+nextConPosition].(float64) + row[keyIndex+12+nextConPosition].(float64) // Compare Conversion Influence - values same as Linear Touch

			if method == AttributionMethodInfluence {
				grandTotalRow[keyIndex+8+nextConPosition] = grandTotalRow[keyIndex+9+nextConPosition]

			}
			if methodCompare == AttributionMethodInfluence {
				grandTotalRow[keyIndex+11+nextConPosition] = grandTotalRow[keyIndex+12+nextConPosition]
			}
		}

	}

	if float64(impressionsCTR) > 0 {
		grandTotalRow[keyIndex+4], _ = U.FloatRoundOffWithPrecision(float64(100*float64(clicksCTR)/float64(impressionsCTR)), U.DefaultPrecision)
	} else {
		grandTotalRow[keyIndex+4] = float64(0)
	}

	if float64(clickAvgCPC) > 0 {
		grandTotalRow[keyIndex+5], _ = U.FloatRoundOffWithPrecision(float64(spendAvgCPC)/float64(clickAvgCPC), U.DefaultPrecision)
	} else {
		grandTotalRow[keyIndex+5] = float64(0)
	}

	if float64(impressionsCPM) > 0 {
		grandTotalRow[keyIndex+6], _ = U.FloatRoundOffWithPrecision(float64(1000*float64(spendCPM))/float64(impressionsCPM), U.DefaultPrecision)
	} else {
		grandTotalRow[keyIndex+6] = float64(0)
	}
	if float64(clicksClickConversionRate) > 0 {
		grandTotalRow[keyIndex+7], _ = U.FloatRoundOffWithPrecision(100*float64(conversionsClickConversionRate)/float64(clicksClickConversionRate), U.DefaultPrecision)
	} else {
		grandTotalRow[keyIndex+7] = float64(0)
	}

	for idx, funcType := range conversionFunTypes {
		nextConPosition := idx * 6
		// Normal conversion [8, 9,10] = [Conversion,Conversion Influence, CPC]
		// Compare conversion [11, 12,13]  = [Conversion, Conversion Influence,CPC]

		if strings.ToLower(funcType) == "sum" {
			// Conversion - CPC
			if grandTotalRow[keyIndex+3].(float64) > 0 {
				grandTotalRow[keyIndex+10+nextConPosition], _ = U.FloatRoundOffWithPrecision(grandTotalRow[keyIndex+8+nextConPosition].(float64)/grandTotalRow[keyIndex+3].(float64), U.DefaultPrecision)
			} else {
				grandTotalRow[keyIndex+10+nextConPosition] = float64(0)
			}

			// Compare Conversion - CPC.
			if grandTotalRow[keyIndex+3].(float64) > 0 {
				grandTotalRow[keyIndex+13+nextConPosition], _ = U.FloatRoundOffWithPrecision(grandTotalRow[keyIndex+11+nextConPosition].(float64)/grandTotalRow[keyIndex+3].(float64), U.DefaultPrecision)
			} else {
				grandTotalRow[keyIndex+13+nextConPosition] = float64(0)
			}

		} else {

			// Conversion - CPC.
			if grandTotalRow[keyIndex+8+nextConPosition].(float64) > 0 {
				grandTotalRow[keyIndex+10+nextConPosition], _ = U.FloatRoundOffWithPrecision(grandTotalRow[keyIndex+3].(float64)/grandTotalRow[keyIndex+8+nextConPosition].(float64), U.DefaultPrecision)
			} else {
				grandTotalRow[keyIndex+10+nextConPosition] = float64(0)
			}
			// Compare Conversion - CPC.
			if grandTotalRow[keyIndex+11+nextConPosition].(float64) > 0 {
				grandTotalRow[keyIndex+13+nextConPosition], _ = U.FloatRoundOffWithPrecision(grandTotalRow[keyIndex+3].(float64)/grandTotalRow[keyIndex+11+nextConPosition].(float64), U.DefaultPrecision)
			} else {
				grandTotalRow[keyIndex+13+nextConPosition] = float64(0)
			}
		}
	}

	rows = append([][]interface{}{grandTotalRow}, rows...)

	return rows

}
func AddGrandTotalRowKPILandingPage(rows [][]interface{}, keyIndex int, goalEvents []string, method string, methodCompare string) [][]interface{} {

	var grandTotalRow []interface{}

	for j := 0; j <= keyIndex; j++ {
		grandTotalRow = append(grandTotalRow, "Grand Total")
	}

	//ConversionEventCount, ConversionEventCountInfluence,ConversionEventCompareCount,ConversionEventCompareCountInfluence

	var defaultMatchingRow []interface{}
	for idx := 0; idx < len(goalEvents); idx++ {
		// one for each - ConversionEventCount, ConversionEventCountInfluence, ConversionEventCompareCount, ConversionEventCompareCountInfluence
		defaultMatchingRow = append(defaultMatchingRow, float64(0), float64(0), float64(0), float64(0))
	}

	grandTotalRow = append(grandTotalRow, defaultMatchingRow...)

	maxRowSize := 0
	for _, row := range rows {

		maxRowSize = U.MaxInt(len(row), maxRowSize)
		if len(row) == 0 || len(row) != maxRowSize {
			continue
		}
		for idx, _ := range goalEvents {
			nextConPosition := idx * 4
			grandTotalRow[keyIndex+1+nextConPosition] = grandTotalRow[keyIndex+1+nextConPosition].(float64) + row[keyIndex+1+nextConPosition].(float64) // Conversion.
			grandTotalRow[keyIndex+2+nextConPosition] = grandTotalRow[keyIndex+2+nextConPosition].(float64) + row[keyIndex+2+nextConPosition].(float64) // Conversion INFLUENCE - values same as Linear Touch
			grandTotalRow[keyIndex+3+nextConPosition] = grandTotalRow[keyIndex+3+nextConPosition].(float64) + row[keyIndex+3+nextConPosition].(float64) // Compare Conversion.
			grandTotalRow[keyIndex+4+nextConPosition] = grandTotalRow[keyIndex+4+nextConPosition].(float64) + row[keyIndex+4+nextConPosition].(float64) // Compare Conversion Influence - values same as Linear Touch

			if method == AttributionMethodInfluence {
				grandTotalRow[keyIndex+1+nextConPosition] = grandTotalRow[keyIndex+2+nextConPosition]
				for i := keyIndex + 4*len(goalEvents) + 1; i < len(grandTotalRow); i += 2 {
					grandTotalRow[i] = grandTotalRow[i+1]

				}
			}
			if methodCompare == AttributionMethodInfluence {
				grandTotalRow[keyIndex+3+nextConPosition] = grandTotalRow[keyIndex+4+nextConPosition]
			}

		}
	}

	rows = append([][]interface{}{grandTotalRow}, rows...)

	return rows

}

func AddGrandTotalRowLandingPage(headers []string, rows [][]interface{}, keyIndex int, method string, methodCompare string) [][]interface{} {

	var grandTotalRow []interface{}

	for j := 0; j <= keyIndex; j++ {
		grandTotalRow = append(grandTotalRow, "Grand Total")
	}

	//ConversionEventCount, ConversionEventCountInfluence,ConversionEventCompareCount,ConversionEventCompareCountInfluence
	defaultMatchingRow := []interface{}{float64(0), float64(0), float64(0), float64(0)}

	grandTotalRow = append(grandTotalRow, defaultMatchingRow...)

	// Remaining linked funnel events
	for i := keyIndex + 5; i < len(headers); i++ {
		grandTotalRow = append(grandTotalRow, float64(0))
	}

	maxRowSize := 0
	for _, row := range rows {

		maxRowSize = U.MaxInt(len(row), maxRowSize)
		if len(row) == 0 || len(row) != maxRowSize {
			continue
		}
		grandTotalRow[keyIndex+1] = grandTotalRow[keyIndex+1].(float64) + row[keyIndex+1].(float64) // Conversion.
		grandTotalRow[keyIndex+2] = grandTotalRow[keyIndex+2].(float64) + row[keyIndex+2].(float64) // Conversion INFLUENCE
		grandTotalRow[keyIndex+3] = grandTotalRow[keyIndex+3].(float64) + row[keyIndex+3].(float64) // Compare Conversion.
		grandTotalRow[keyIndex+4] = grandTotalRow[keyIndex+4].(float64) + row[keyIndex+4].(float64) // Compare Conversion Influence

		// Remaining linked funnel events & Conversion rates
		for i := keyIndex + 5; i < len(grandTotalRow); i++ {
			grandTotalRow[i] = grandTotalRow[i].(float64) + row[i].(float64)
		}
	}
	if method == AttributionMethodInfluence {
		grandTotalRow[keyIndex+1] = grandTotalRow[keyIndex+2]
		for i := keyIndex + 5; i < len(grandTotalRow); i += 2 {
			grandTotalRow[i] = grandTotalRow[i+1]

		}
	}
	if methodCompare == AttributionMethodInfluence {
		grandTotalRow[keyIndex+3] = grandTotalRow[keyIndex+4]
	}

	rows = append([][]interface{}{grandTotalRow}, rows...)

	return rows

}

// FilterRows filters rows based on attribution key. ex. $none exclusion for 'Keyword' type report.
func FilterRows(rows [][]interface{}, attributionKey string, keyIndex int) [][]interface{} {

	// Select the best value for attributionKey
	switch attributionKey {
	case AttributionKeyKeyword:
		filteredRows := make([][]interface{}, 0)
		for _, mapRow := range rows {
			if mapRow[keyIndex].(string) != PropertyValueNone {
				filteredRows = append(filteredRows, mapRow)
			}
		}
		return filteredRows
	default:
	}
	return rows
}

// AddUpConversionEventCount Groups all unique users by attributionId and adds it to attributionData
func AddUpConversionEventCount(usersIdAttributionIdMap map[string][]AttributionKeyWeight, sessionWT map[string][]float64) map[string]*AttributionData {

	attributionData := make(map[string]*AttributionData)

	for userID, attributionKeys := range usersIdAttributionIdMap {

		userIDWeightsForEachGoalEvent := sessionWT[userID] // Revenue, Pipeline, DealValue, etc

		for _, keyWeight := range attributionKeys { // camp1, camp2, camp3, etc

			if _, exists := attributionData[keyWeight.Key]; !exists {
				attributionData[keyWeight.Key] = &AttributionData{}
			}

			for idx := 0; idx < len(userIDWeightsForEachGoalEvent); idx++ {
				// filling additional data if ConversionEventCount is empty
				if len(attributionData[keyWeight.Key].ConversionEventCount) < idx+1 {
					attributionData[keyWeight.Key].ConversionEventCount = append(attributionData[keyWeight.Key].ConversionEventCount, float64(0))
				}

				if len(attributionData[keyWeight.Key].ConversionEventCountInfluence) < idx+1 {
					attributionData[keyWeight.Key].ConversionEventCountInfluence = append(attributionData[keyWeight.Key].ConversionEventCountInfluence, float64(0))
				}

				weightedValue := keyWeight.Weight * userIDWeightsForEachGoalEvent[idx]
				attributionData[keyWeight.Key].ConversionEventCount[idx] = float64(attributionData[keyWeight.Key].ConversionEventCount[idx] + weightedValue)
				attributionData[keyWeight.Key].ConversionEventCountInfluence[idx] = float64(attributionData[keyWeight.Key].ConversionEventCountInfluence[idx] + (weightedValue / float64(len(attributionKeys))))
			}
		}
	}
	// non-used sessionWT rows can be written back to '$none' userID
	return attributionData
}

// AddUpLinkedFunnelEventCount Attribute each user to the conversion event and linked event by attribution Id.
func AddUpLinkedFunnelEventCount(linkedEvents []QueryEventWithProperties,
	attributionData map[string]*AttributionData, linkedUserAttributionData map[string]map[string][]AttributionKeyWeight) {

	linkedEventToPositionMap := make(map[string]int)
	for position, linkedEvent := range linkedEvents {
		linkedEventToPositionMap[linkedEvent.Name] = position
	}
	// fill up all the linked events count with 0 value
	for _, attributionRow := range attributionData {
		if attributionRow != nil {
			for len(attributionRow.LinkedEventsCount) < len(linkedEvents) {
				attributionRow.LinkedEventsCount = append(attributionRow.LinkedEventsCount, 0.0)
			}
			for len(attributionRow.LinkedEventsCountInfluence) < len(linkedEvents) {
				attributionRow.LinkedEventsCountInfluence = append(attributionRow.LinkedEventsCountInfluence, 0.0)
			}
		}
	}
	// Update linked up events with event hit count.
	for linkedEventName, userIdAttributionIdMap := range linkedUserAttributionData {
		for _, attributionKeys := range userIdAttributionIdMap {
			for _, keyWeight := range attributionKeys {
				if attributionData[keyWeight.Key] != nil {
					attributionData[keyWeight.Key].LinkedEventsCount[linkedEventToPositionMap[linkedEventName]] += keyWeight.Weight
					attributionData[keyWeight.Key].LinkedEventsCountInfluence[linkedEventToPositionMap[linkedEventName]] += keyWeight.Weight / float64(len(attributionKeys))
				}
			}
		}
	}
}

func IsValidAttributionKeyValueAND(attributionKeyType string, keyValue string,
	filters []AttributionKeyFilter) bool {

	for _, filter := range filters {
		// supports AND and treats blank operator as AND
		if filter.LogicalOp == "OR" {
			continue
		}
		filterResult := applyOperator(attributionKeyType, keyValue, filter)
		// AND is false for any false.
		if !filterResult {
			return false
		}
	}
	return true
}

func IsValidAttributionKeyValueOR(attributionKeyType string, keyValue string,
	filters []AttributionKeyFilter) bool {

	for _, filter := range filters {
		if filter.LogicalOp != "OR" {
			continue
		}
		filterResult := applyOperator(attributionKeyType, keyValue, filter)
		// OR is true for any true
		if filterResult {
			return true
		}
	}
	return false
}

func applyOperator(attributionKeyType string, keyValue string,
	filter AttributionKeyFilter) bool {

	filterResult := true
	// Currently only supporting matching key filters
	if filter.AttributionKey == attributionKeyType {
		switch filter.Operator {
		case EqualsOpStr:
			if keyValue != filter.Value {
				filterResult = false
			}
		case NotEqualOpStr:
			if keyValue == filter.Value {
				filterResult = false
			}
		case ContainsOpStr:
			if !strings.Contains(keyValue, filter.Value) {
				filterResult = false
			}
		case NotContainsOpStr:
			if strings.Contains(keyValue, filter.Value) {
				filterResult = false
			}
		default:
			filterResult = false
		}
	}
	return filterResult
}

func DoesAdwordsReportExist(attributionKey string) bool {
	// only campaign, adgroup, keyword reports available
	if attributionKey == AttributionKeyCampaign || attributionKey == AttributionKeyAdgroup ||
		attributionKey == AttributionKeyKeyword || attributionKey == AttributionKeySource || attributionKey == AttributionKeyChannel {
		return true
	}
	return false
}
func DoesBingAdsReportExist(attributionKey string) bool {
	// only campaign, adgroup, keyword reports available
	if attributionKey == AttributionKeyCampaign || attributionKey == AttributionKeyAdgroup ||
		attributionKey == AttributionKeyKeyword || attributionKey == AttributionKeySource || attributionKey == AttributionKeyChannel {
		return true
	}
	return false
}
func DoesCustomAdsReportExist(attributionKey string) bool {
	// only campaign, adgroup, keyword reports available
	if attributionKey == AttributionKeyCampaign || attributionKey == AttributionKeyAdgroup ||
		attributionKey == AttributionKeyKeyword || attributionKey == AttributionKeySource || attributionKey == AttributionKeyChannel {
		return true
	}
	return false
}
func DoesFBReportExist(attributionKey string) bool {
	// only campaign, adgroup reports available
	if attributionKey == AttributionKeyCampaign || attributionKey == AttributionKeyAdgroup || attributionKey == AttributionKeySource || attributionKey == AttributionKeyChannel {
		return true
	}
	return false
}

func DoesLinkedinReportExist(attributionKey string) bool {
	// only campaign, adgroup reports available
	if attributionKey == AttributionKeyCampaign || attributionKey == AttributionKeyAdgroup || attributionKey == AttributionKeySource || attributionKey == AttributionKeyChannel {
		return true
	}
	return false
}

func AddTheAddedKeysAndMetrics(attributionData *map[string]*AttributionData, query *AttributionQuery, sessions map[string]map[string]UserSessionData, noOfConversionEvents int) {

	// Extract out key based info
	sessionKeyMarketingInfo := make(map[string]MarketingData)
	for _, value := range sessions {
		for k, v := range value {

			// Runs for each unique userID-Key pair
			sessionKeyMarketingInfo[k] = v.MarketingInfo
		}
	}

	// Creating an empty linked events row.
	emptyConversionEventRow := make([]float64, 0)
	emptyConversionEventRowInfluence := make([]float64, 0)
	for i := 0; i < noOfConversionEvents; i++ {
		emptyConversionEventRow = append(emptyConversionEventRow, float64(0))
		emptyConversionEventRowInfluence = append(emptyConversionEventRowInfluence, float64(0))
	}

	// Creating an empty linked events row.
	emptyLinkedEventRow := make([]float64, 0)
	emptyLinkedEventRowInfluence := make([]float64, 0)
	for i := 0; i < len(query.LinkedEvents); i++ {
		emptyLinkedEventRow = append(emptyLinkedEventRow, float64(0))
		emptyLinkedEventRowInfluence = append(emptyLinkedEventRowInfluence, float64(0))
	}
	for _, attributionIDMap := range sessions {
		for key, sessionTimestamp := range attributionIDMap {
			// Only count sessions that happened during attribution period.
			if sessionTimestamp.WithinQueryPeriod {

				// Create a row in AttributionData if no key is present for this session
				if _, ok := (*attributionData)[key]; !ok {
					(*attributionData)[key] = &AttributionData{}

					(*attributionData)[key].ConversionEventCount = emptyConversionEventRow
					(*attributionData)[key].ConversionEventCountInfluence = emptyConversionEventRowInfluence
					(*attributionData)[key].ConversionEventCompareCount = emptyConversionEventRow
					(*attributionData)[key].ConversionEventCompareCountInfluence = emptyConversionEventRowInfluence
					if len(query.LinkedEvents) > 0 {
						// Init the linked events with 0.0 value.
						tempRow := emptyLinkedEventRow
						tempRowInfluence := emptyConversionEventRowInfluence
						(*attributionData)[key].LinkedEventsCount = tempRow
						(*attributionData)[key].LinkedEventsCountInfluence = tempRowInfluence
					}
				}

				if _, exists := sessionKeyMarketingInfo[key]; exists {
					// Add the marketing info
					(*attributionData)[key].MarketingInfo = sessionKeyMarketingInfo[key]
					switch query.AttributionKey {
					case AttributionKeyCampaign:
						(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, sessionKeyMarketingInfo[key].Channel)
						(*attributionData)[key].Name = sessionKeyMarketingInfo[key].CampaignName
					case AttributionKeyAdgroup:
						(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, sessionKeyMarketingInfo[key].Channel, sessionKeyMarketingInfo[key].CampaignName)
						(*attributionData)[key].Name = sessionKeyMarketingInfo[key].AdgroupName
					case AttributionKeyKeyword:
						(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, sessionKeyMarketingInfo[key].Channel, sessionKeyMarketingInfo[key].CampaignName, sessionKeyMarketingInfo[key].AdgroupName, sessionKeyMarketingInfo[key].KeywordMatchType)
						(*attributionData)[key].Name = sessionKeyMarketingInfo[key].KeywordName
					case AttributionKeySource:
						(*attributionData)[key].Name = sessionKeyMarketingInfo[key].Source
					case AttributionKeyChannel:
						(*attributionData)[key].Name = sessionKeyMarketingInfo[key].ChannelGroup
					case AttributionKeyLandingPage:
						(*attributionData)[key].Name = sessionKeyMarketingInfo[key].LandingPageUrl
					}

				}
			}
		}
	}
}

func ApplyFilter(attributionData *map[string]*AttributionData, query *AttributionQuery) {
	// Filter out the key values from query (apply filter after performance enrichment)
	for key, value := range *attributionData {
		attributionId := value.Name
		if !IsValidAttributionKeyValueAND(query.AttributionKey,
			attributionId, query.AttributionKeyFilter) && !IsValidAttributionKeyValueOR(query.AttributionKey,
			attributionId, query.AttributionKeyFilter) {
			delete(*attributionData, key)
		}
	}
}

func AddPerformanceData(attributionData *map[string]*AttributionData, attributionKey string, marketingData *MarketingReports, noOfConversionEvents int) {

	AddAdwordsPerformanceReportInfo(attributionData, attributionKey, marketingData, noOfConversionEvents)
	AddFacebookPerformanceReportInfo(attributionData, attributionKey, marketingData, noOfConversionEvents)
	AddLinkedinPerformanceReportInfo(attributionData, attributionKey, marketingData, noOfConversionEvents)
	AddBingAdsPerformanceReportInfo(attributionData, attributionKey, marketingData, noOfConversionEvents)
	AddCustomAdsPerformanceReportInfo(attributionData, attributionKey, marketingData, noOfConversionEvents)
}

func AddAdwordsPerformanceReportInfo(attributionData *map[string]*AttributionData, attributionKey string, marketingData *MarketingReports, noOfConversionEvents int) {

	switch attributionKey {
	case AttributionKeyCampaign:
		addMetricsFromReport(attributionData, marketingData.AdwordsCampaignKeyData, attributionKey, ChannelAdwords, noOfConversionEvents)
	case AttributionKeyAdgroup:
		addMetricsFromReport(attributionData, marketingData.AdwordsAdgroupKeyData, attributionKey, ChannelAdwords, noOfConversionEvents)
	case AttributionKeyKeyword:
		addMetricsFromReport(attributionData, marketingData.AdwordsKeywordKeyData, attributionKey, ChannelAdwords, noOfConversionEvents)
	default:
		// no enrichment for any other type
		return
	}
}

func AddBingAdsPerformanceReportInfo(attributionData *map[string]*AttributionData, attributionKey string, marketingData *MarketingReports, noOfConversionEvents int) {

	switch attributionKey {
	case AttributionKeyCampaign:
		addMetricsFromReport(attributionData, marketingData.BingAdsCampaignKeyData, attributionKey, ChannelBingAds, noOfConversionEvents)
	case AttributionKeyAdgroup:
		addMetricsFromReport(attributionData, marketingData.BingAdsAdgroupKeyData, attributionKey, ChannelBingAds, noOfConversionEvents)
	case AttributionKeyKeyword:
		addMetricsFromReport(attributionData, marketingData.BingAdsKeywordKeyData, attributionKey, ChannelBingAds, noOfConversionEvents)
	default:
		return
	}
}

func AddCustomAdsPerformanceReportInfo(attributionData *map[string]*AttributionData, attributionKey string, marketingData *MarketingReports, noOfConversionEvents int) {

	switch attributionKey {
	case AttributionKeyCampaign:
		addMetricsFromReport(attributionData, marketingData.CustomAdsCampaignKeyData, attributionKey, ChannelCustomAds, noOfConversionEvents)
	case AttributionKeyAdgroup:
		addMetricsFromReport(attributionData, marketingData.CustomAdsAdgroupKeyData, attributionKey, ChannelCustomAds, noOfConversionEvents)
	case AttributionKeyKeyword:
		addMetricsFromReport(attributionData, marketingData.CustomAdsKeywordKeyData, attributionKey, ChannelCustomAds, noOfConversionEvents)
	default:
		return
	}
}

func AddFacebookPerformanceReportInfo(attributionData *map[string]*AttributionData, attributionKey string, marketingData *MarketingReports, noOfConversionEvents int) {

	switch attributionKey {
	case AttributionKeyCampaign:
		addMetricsFromReport(attributionData, marketingData.FacebookCampaignKeyData, attributionKey, ChannelFacebook, noOfConversionEvents)
	case AttributionKeyAdgroup:
		addMetricsFromReport(attributionData, marketingData.FacebookAdgroupKeyData, attributionKey, ChannelFacebook, noOfConversionEvents)
	case AttributionKeyKeyword:
		// No keyword report for fb.
		return
	default:
		// no enrichment for any other type
		return
	}
}

func AddLinkedinPerformanceReportInfo(attributionData *map[string]*AttributionData, attributionKey string, marketingData *MarketingReports, noOfConversionEvents int) {

	switch attributionKey {
	case AttributionKeyCampaign:
		addMetricsFromReport(attributionData, marketingData.LinkedinCampaignKeyData, attributionKey, ChannelLinkedin, noOfConversionEvents)
	case AttributionKeyAdgroup:
		addMetricsFromReport(attributionData, marketingData.LinkedinAdgroupKeyData, attributionKey, ChannelLinkedin, noOfConversionEvents)
	case AttributionKeyKeyword:
		// No keyword report for Linkedin.
		return
	default:
		// no enrichment for any other type
		return
	}
}

func addMetricsFromReport(attributionData *map[string]*AttributionData, reportKeyData map[string]MarketingData, attributionKey string, channel string, noOfConversionEvents int) {

	// Creating an empty linked events row.
	emptyConversionEventRow := make([]float64, 0)
	emptyConversionEventRowInfluence := make([]float64, 0)
	for i := 0; i < noOfConversionEvents; i++ {
		emptyConversionEventRow = append(emptyConversionEventRow, float64(0))
		emptyConversionEventRowInfluence = append(emptyConversionEventRowInfluence, float64(0))
	}

	for key, value := range reportKeyData {

		if value.Impressions == 0 && value.Clicks == 0 && value.Spend == 0 {
			// ignore ZERO valued keys
			continue
		}
		// Create a new record if not found
		if _, found := (*attributionData)[key]; !found {

			(*attributionData)[key] = &AttributionData{}
			(*attributionData)[key].MarketingInfo = reportKeyData[key]
			switch attributionKey {
			case AttributionKeyCampaign:
				(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, reportKeyData[key].Channel)
				(*attributionData)[key].Name = reportKeyData[key].CampaignName
			case AttributionKeyAdgroup:
				(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, reportKeyData[key].Channel, reportKeyData[key].CampaignName)
				(*attributionData)[key].Name = reportKeyData[key].AdgroupName
			case AttributionKeyKeyword:
				(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, reportKeyData[key].Channel, reportKeyData[key].CampaignName, reportKeyData[key].AdgroupName, reportKeyData[key].KeywordMatchType)
				(*attributionData)[key].Name = reportKeyData[key].KeywordName
			case AttributionKeySource:
				(*attributionData)[key].Name = reportKeyData[key].Source
			case AttributionKeyChannel:
				(*attributionData)[key].Name = reportKeyData[key].ChannelGroup
			case AttributionKeyLandingPage:
				(*attributionData)[key].Name = reportKeyData[key].LandingPageUrl
			}
			(*attributionData)[key].ConversionEventCount = emptyConversionEventRow
			(*attributionData)[key].ConversionEventCountInfluence = emptyConversionEventRowInfluence
			(*attributionData)[key].ConversionEventCompareCount = emptyConversionEventRow
			(*attributionData)[key].ConversionEventCompareCountInfluence = emptyConversionEventRowInfluence
		}

		if (*attributionData)[key].CustomDimensions == nil {
			(*attributionData)[key].CustomDimensions = make(map[string]interface{})
		}
		if channel == ChannelCustomAds {
			(*attributionData)[key].Channel = reportKeyData[key].Channel
		} else {
			(*attributionData)[key].Channel = channel
		}
		(*attributionData)[key].Impressions = value.Impressions
		(*attributionData)[key].Clicks = value.Clicks
		(*attributionData)[key].Spend = value.Spend

		// replacing the marketing info on key match
		(*attributionData)[key].MarketingInfo = value

	}
}

func ComputeAdditionalMetrics(attributionData *map[string]*AttributionData) {

	for k, v := range *attributionData {
		(*attributionData)[k].CTR = 0
		(*attributionData)[k].CPM = 0
		(*attributionData)[k].AvgCPC = 0
		(*attributionData)[k].ClickConversionRate = 0
		if v.Impressions > 0 {
			(*attributionData)[k].CTR, _ = U.FloatRoundOffWithPrecision(100*float64(v.Clicks)/float64(v.Impressions), U.DefaultPrecision)
			(*attributionData)[k].CPM, _ = U.FloatRoundOffWithPrecision(1000*float64(v.Spend)/float64(v.Impressions), U.DefaultPrecision)
		}
		if v.Clicks > 0 {
			(*attributionData)[k].AvgCPC, _ = U.FloatRoundOffWithPrecision(float64(v.Spend)/float64(v.Clicks), U.DefaultPrecision)
			if (*attributionData)[k].ConversionEventCount == nil || len((*attributionData)[k].ConversionEventCount) == 0 {
				(*attributionData)[k].ConversionEventCount = append((*attributionData)[k].ConversionEventCount, float64(0))
			}
			(*attributionData)[k].ClickConversionRate, _ = U.FloatRoundOffWithPrecision(100*float64((*attributionData)[k].ConversionEventCount[0])/float64(v.Clicks), U.DefaultPrecision)
		}
	}
}

func GetMarketingDataKey(attributionKey string, data MarketingData) string {

	key := ""
	switch attributionKey {
	case AttributionKeyCampaign:
		// we know we get campaignIDReport here
		key = key + data.Channel + KeyDelimiter + U.IfThenElse(data.Name != "" && data.Name != PropertyValueNone, data.Name, data.CampaignName).(string)
	case AttributionKeyAdgroup:
		key = key + data.Channel + KeyDelimiter + data.CampaignName + KeyDelimiter + U.IfThenElse(data.Name != "" && data.Name != PropertyValueNone, data.Name, data.AdgroupName).(string)
	case AttributionKeyKeyword:
		// we know we get keywordIDReport here
		key = key + data.Channel + KeyDelimiter + data.CampaignName + KeyDelimiter + data.AdgroupName + KeyDelimiter + data.KeywordMatchType + KeyDelimiter + U.IfThenElse(data.Name != "" && data.Name != PropertyValueNone, data.Name, data.KeywordName).(string)
	case AttributionKeySource:
		key = key + U.IfThenElse(data.Name != "" && data.Name != PropertyValueNone, data.Name, data.Source).(string)
	case AttributionKeyChannel:
		key = key + U.IfThenElse(data.Name != "" && data.Name != PropertyValueNone, data.Name, data.ChannelGroup).(string)
	case AttributionKeyLandingPage:
		key = key + U.IfThenElse(data.Name != "" && data.Name != PropertyValueNone, data.Name, data.LandingPageUrl).(string)
	case AttributionKeyAllPageView:
		key = key + U.IfThenElse(data.Name != "" && data.Name != PropertyValueNone, data.Name, data.AllPageView).(string)
	default:
		key = key + data.Name
	}
	return key
}

func GetKeyMapToData(attributionKey string, allRows []MarketingData, idMarketingDataMap map[string]MarketingData) map[string]MarketingData {

	keyToData := make(map[string]MarketingData)
	for i, v := range allRows {
		switch attributionKey {
		case AttributionKeyCampaign:
			v.CampaignName = idMarketingDataMap[v.ID].CampaignName
			v.Name = v.CampaignName
			allRows[i] = v
		case AttributionKeyAdgroup:
			v.AdgroupName = idMarketingDataMap[v.ID].AdgroupName
			v.Name = v.AdgroupName
			allRows[i] = v
		case AttributionKeyKeyword:
			v.KeywordName = idMarketingDataMap[v.ID].KeywordName
			v.Name = v.KeywordName
			allRows[i] = v
		}

		key := GetMarketingDataKey(attributionKey, v)
		if _, ok := keyToData[key]; ok {
			v = mergeMarketingData(keyToData[key], v)
		}
		keyToData[key] = v
		val := MarketingData{}
		U.DeepCopy(&v, &val)
		val.Key = key
		keyToData[key] = val
	}
	return keyToData
}

func ProcessOTPEventRows(rows *sql.Rows, query *AttributionQuery,
	logCtx log.Entry, queryID string) (map[string]map[string]UserSessionData, []string, error) {

	attributedSessionsByUserId := make(map[string]map[string]UserSessionData)
	userIdMap := make(map[string]bool)
	var userIdsWithSession []string

	startReadTime := time.Now()
	for rows.Next() {
		var userIDNull sql.NullString
		var campaignIDNull sql.NullString
		var campaignNameNull sql.NullString
		var adgroupIDNull sql.NullString
		var adgroupNameNull sql.NullString
		var sourceNameNull sql.NullString
		var channelGroupNull sql.NullString
		var typeNull sql.NullString
		var attributionIdNull sql.NullString
		var timestampNull sql.NullInt64

		if err := rows.Scan(&userIDNull, &campaignIDNull, &campaignNameNull, &adgroupIDNull, &adgroupNameNull,
			&sourceNameNull, &channelGroupNull, &typeNull, &attributionIdNull, &timestampNull); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row (OTP). Continuing")
			continue
		}

		var userID string
		var campaignID string
		var campaignName string
		var adgroupID string
		var adgroupName string
		var sourceName string
		var channelGroup string
		var typeName string
		var attributionKeyName string
		var timestamp int64

		userID = U.IfThenElse(userIDNull.Valid, userIDNull.String, PropertyValueNone).(string)
		campaignID = U.IfThenElse(campaignIDNull.Valid, campaignIDNull.String, PropertyValueNone).(string)
		campaignName = U.IfThenElse(campaignNameNull.Valid, campaignNameNull.String, PropertyValueNone).(string)
		adgroupID = U.IfThenElse(adgroupIDNull.Valid, adgroupIDNull.String, PropertyValueNone).(string)
		adgroupName = U.IfThenElse(adgroupNameNull.Valid, adgroupNameNull.String, PropertyValueNone).(string)
		sourceName = U.IfThenElse(sourceNameNull.Valid, sourceNameNull.String, PropertyValueNone).(string)
		channelGroup = U.IfThenElse(channelGroupNull.Valid, channelGroupNull.String, PropertyValueNone).(string)
		typeName = U.IfThenElse(typeNull.Valid, typeNull.String, PropertyValueNone).(string)
		attributionKeyName = U.IfThenElse(attributionIdNull.Valid, attributionIdNull.String, PropertyValueNone).(string)
		timestamp = U.IfThenElse(timestampNull.Valid, timestampNull.Int64, int64(0)).(int64)

		// apply filter at extracting session level itself
		if !IsValidAttributionKeyValueAND(query.AttributionKey,
			attributionKeyName, query.AttributionKeyFilter) && !IsValidAttributionKeyValueOR(query.AttributionKey,
			attributionKeyName, query.AttributionKeyFilter) {
			continue
		}

		// Exclude for non-matching tactic or offer type but include the none values
		if query.TacticOfferType != typeName && typeName != PropertyValueNone && query.TacticOfferType != MarketingEventTypeTacticOffer {
			continue
		}
		if _, ok := userIdMap[userID]; !ok {
			userIdsWithSession = append(userIdsWithSession, userID)
			userIdMap[userID] = true
		}
		marketingValues := MarketingData{Channel: SessionChannelOTP, CampaignID: campaignID, CampaignName: campaignName, AdgroupID: adgroupID, AdgroupName: adgroupName,
			KeywordName: PropertyValueNone, KeywordMatchType: PropertyValueNone,
			Source: sourceName, TypeName: typeName, ChannelGroup: channelGroup}

		// Name
		marketingValues.Name = attributionKeyName
		// Add the unique attributionKey key
		marketingValues.Key = GetMarketingDataKey(query.AttributionKey, marketingValues)
		uniqueAttributionKey := marketingValues.Key
		// add session info uniquely for user-attributionKeyName pair
		if _, ok := attributedSessionsByUserId[userID]; ok {

			if userSessionData, ok := attributedSessionsByUserId[userID][uniqueAttributionKey]; ok {
				userSessionData.MinTimestamp = U.Min(userSessionData.MinTimestamp, timestamp)
				userSessionData.MaxTimestamp = U.Max(userSessionData.MaxTimestamp, timestamp)
				userSessionData.TimeStamps = append(userSessionData.TimeStamps, timestamp)
				userSessionData.WithinQueryPeriod = userSessionData.WithinQueryPeriod || isSessionWithinQueryPeriod(query.QueryType, query.LookbackDays, query.From, query.To, timestamp)
				attributedSessionsByUserId[userID][uniqueAttributionKey] = userSessionData
			} else {
				userSessionDataNew := UserSessionData{MinTimestamp: timestamp,
					MaxTimestamp: timestamp, TimeStamps: []int64{timestamp},
					WithinQueryPeriod: isSessionWithinQueryPeriod(query.QueryType, query.LookbackDays, query.From, query.To, timestamp),
					MarketingInfo:     marketingValues}
				attributedSessionsByUserId[userID][uniqueAttributionKey] = userSessionDataNew
			}
		} else {
			attributedSessionsByUserId[userID] = make(map[string]UserSessionData)
			userSessionDataNew := UserSessionData{MinTimestamp: timestamp,
				MaxTimestamp: timestamp, TimeStamps: []int64{timestamp},
				WithinQueryPeriod: isSessionWithinQueryPeriod(query.QueryType, query.LookbackDays, query.From, query.To, timestamp),
				MarketingInfo:     marketingValues}
			attributedSessionsByUserId[userID][uniqueAttributionKey] = userSessionDataNew
		}
	}
	err := rows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in ProcessOTPEventRows")
		return nil, nil, err
	}
	U.LogReadTimeWithQueryRequestID(startReadTime, queryID, &log.Fields{"query": query})

	return attributedSessionsByUserId, userIdsWithSession, nil
}

func ProcessEventRows(rows *sql.Rows, query *AttributionQuery, reports *MarketingReports,
	contentGroupNamesList []string, attributedSessionsByUserId *map[string]map[string]UserSessionData,
	userIdsWithSession *[]string, logCtx log.Entry, queryID string) error {

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	userIdMap := make(map[string]bool)

	type MissingCollection struct {
		AttributionKey string
		GCLID          string
		CampaignID     string
		AdgroupID      string
	}
	var missingIDs []MissingCollection
	countEnrichedGclid := 0
	countEnrichedMarketingId := 0

	startReadTime := time.Now()
	for rows.Next() {
		var userIDNull sql.NullString
		var campaignIDNull sql.NullString
		var campaignNameNull sql.NullString
		var adgroupIDNull sql.NullString
		var adgroupNameNull sql.NullString
		var keywordNameNull sql.NullString
		var keywordMatchTypeNull sql.NullString
		var sourceNameNull sql.NullString
		var channelGroupNull sql.NullString
		var attributionIdNull sql.NullString
		var gclIDNull sql.NullString
		var landingPageUrlNull sql.NullString
		var allPageViewUrlNull sql.NullString
		var timestampNull sql.NullInt64
		contentGroupValuesListNull := make([]sql.NullString, len(contentGroupNamesList))

		var fields []interface{}
		fields = append(fields, &userIDNull, &campaignIDNull, &campaignNameNull,
			&adgroupIDNull, &adgroupNameNull, &keywordNameNull, &keywordMatchTypeNull, &sourceNameNull, &channelGroupNull,
			&attributionIdNull, &gclIDNull, &landingPageUrlNull, &allPageViewUrlNull)

		// contentGroupValuesListNull wil be empty for queries where property is not "Landing page url"
		for i := 0; i < len(contentGroupValuesListNull); i++ {
			fields = append(fields, &contentGroupValuesListNull[i])
		}

		fields = append(fields, &timestampNull)

		if err := rows.Scan(fields...); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
			continue
		}

		var userID string
		var campaignID string
		var campaignName string
		var adgroupID string
		var adgroupName string
		var keywordName string
		var keywordMatchType string
		var sourceName string
		var channelGroup string
		var attributionKeyName string
		var gclID string
		var landingPageUrl string
		var allPageViewUrl string
		var timestamp int64
		contentGroupValuesMap := make(map[string]string)

		userID = U.IfThenElse(userIDNull.Valid, userIDNull.String, PropertyValueNone).(string)
		campaignID = U.IfThenElse(campaignIDNull.Valid, campaignIDNull.String, PropertyValueNone).(string)
		campaignName = U.IfThenElse(campaignNameNull.Valid, campaignNameNull.String, PropertyValueNone).(string)
		adgroupID = U.IfThenElse(adgroupIDNull.Valid, adgroupIDNull.String, PropertyValueNone).(string)
		adgroupName = U.IfThenElse(adgroupNameNull.Valid, adgroupNameNull.String, PropertyValueNone).(string)
		keywordName = U.IfThenElse(keywordNameNull.Valid, keywordNameNull.String, PropertyValueNone).(string)
		keywordMatchType = U.IfThenElse(keywordMatchTypeNull.Valid, keywordMatchTypeNull.String, PropertyValueNone).(string)
		sourceName = U.IfThenElse(sourceNameNull.Valid, sourceNameNull.String, PropertyValueNone).(string)
		channelGroup = U.IfThenElse(channelGroupNull.Valid, channelGroupNull.String, PropertyValueNone).(string)
		attributionKeyName = U.IfThenElse(attributionIdNull.Valid, attributionIdNull.String, PropertyValueNone).(string)
		gclID = U.IfThenElse(gclIDNull.Valid, gclIDNull.String, PropertyValueNone).(string)
		landingPageUrl = U.IfThenElse(landingPageUrlNull.Valid, landingPageUrlNull.String, PropertyValueNone).(string)
		allPageViewUrl = U.IfThenElse(allPageViewUrlNull.Valid, allPageViewUrlNull.String, PropertyValueNone).(string)
		timestamp = U.IfThenElse(timestampNull.Valid, timestampNull.Int64, int64(0)).(int64)
		for i, val := range contentGroupValuesListNull {
			contentGroupValuesMap[contentGroupNamesList[i]] = U.IfThenElse(val.Valid, val.String, PropertyValueNone).(string)
		}

		// apply filter at extracting session level itself
		if !IsValidAttributionKeyValueAND(query.AttributionKey,
			attributionKeyName, query.AttributionKeyFilter) && !IsValidAttributionKeyValueOR(query.AttributionKey,
			attributionKeyName, query.AttributionKeyFilter) {
			continue
		}
		if _, ok := userIdMap[userID]; !ok {
			*userIdsWithSession = append(*userIdsWithSession, userID)
			userIdMap[userID] = true
		}
		marketingValues := MarketingData{Channel: PropertyValueNone, CampaignID: campaignID, CampaignName: campaignName, AdgroupID: adgroupID,
			AdgroupName: adgroupName, KeywordName: keywordName, KeywordMatchType: keywordMatchType, Source: sourceName, ChannelGroup: channelGroup,
			LandingPageUrl: landingPageUrl, AllPageView: allPageViewUrl, ContentGroupValuesMap: contentGroupValuesMap}
		// Override GCLID based campaign info if presents
		if gclID != PropertyValueNone && !(query.AttributionKey == AttributionKeyKeyword && !IsASearchSlotKeyword(&(*reports).AdwordsGCLIDData, gclID)) {
			countEnrichedGclid++
			var attributionIdBasedOnGclID string
			attributionIdBasedOnGclID, marketingValues = EnrichUsingGCLID(&(*reports).AdwordsGCLIDData, gclID, query.AttributionKey, marketingValues)
			marketingValues.Channel = ChannelAdwords
			// In cases where GCLID is present in events, but not in adwords report (as users tend to bookmark expired URLs),
			// fallback is attributionId
			if U.IsNonEmptyKey(attributionIdBasedOnGclID) {
				attributionKeyName = attributionIdBasedOnGclID
			} else {
				missingIDs = append(missingIDs, MissingCollection{AttributionKey: query.AttributionKey, GCLID: gclID})
			}
		}
		// Even after the data is enriched gclid, for latest name, enrich it using campaign/adgroup report
		if (query.AttributionKey == AttributionKeyCampaign && U.IsNonEmptyKey(campaignID)) ||
			(query.AttributionKey == AttributionKeyAdgroup && U.IsNonEmptyKey(adgroupID)) {
			// enrich for campaign/adgroup based session having campaign_id/adgroup_id
			countEnrichedMarketingId++
			var attributionIdBasedOnEnrichment string
			attributionIdBasedOnEnrichment, marketingValues = EnrichUsingMarketingID(query.AttributionKey, marketingValues, reports)
			if U.IsNonEmptyKey(attributionIdBasedOnEnrichment) {
				attributionKeyName = attributionIdBasedOnEnrichment
			} else {
				missingIDs = append(missingIDs, MissingCollection{AttributionKey: query.AttributionKey, CampaignID: campaignID, AdgroupID: adgroupID})
			}
		}

		if sourceName == "bing" {
			marketingValues.Channel = ChannelBingAds
		}
		// Name
		marketingValues.Name = attributionKeyName
		// Add the unique attributionKey key
		marketingValues.Key = GetMarketingDataKey(query.AttributionKey, marketingValues)
		uniqueAttributionKey := marketingValues.Key
		// add session info uniquely for user-attributionId pair
		if _, ok := (*attributedSessionsByUserId)[userID]; ok {

			if userSessionData, ok := (*attributedSessionsByUserId)[userID][uniqueAttributionKey]; ok {
				userSessionData.MinTimestamp = U.Min(userSessionData.MinTimestamp, timestamp)
				userSessionData.MaxTimestamp = U.Max(userSessionData.MaxTimestamp, timestamp)
				userSessionData.TimeStamps = append(userSessionData.TimeStamps, timestamp)
				userSessionData.WithinQueryPeriod = userSessionData.WithinQueryPeriod || isSessionWithinQueryPeriod(query.QueryType, query.LookbackDays, query.From, query.To, timestamp)
				(*attributedSessionsByUserId)[userID][uniqueAttributionKey] = userSessionData
			} else {
				userSessionDataNew := UserSessionData{MinTimestamp: timestamp,
					MaxTimestamp: timestamp, TimeStamps: []int64{timestamp},
					WithinQueryPeriod: isSessionWithinQueryPeriod(query.QueryType, query.LookbackDays, query.From, query.To, timestamp),
					MarketingInfo:     marketingValues}
				(*attributedSessionsByUserId)[userID][uniqueAttributionKey] = userSessionDataNew
			}
		} else {
			(*attributedSessionsByUserId)[userID] = make(map[string]UserSessionData)
			userSessionDataNew := UserSessionData{MinTimestamp: timestamp,
				MaxTimestamp: timestamp, TimeStamps: []int64{timestamp},
				WithinQueryPeriod: isSessionWithinQueryPeriod(query.QueryType, query.LookbackDays, query.From, query.To, timestamp), MarketingInfo: marketingValues}
			(*attributedSessionsByUserId)[userID][uniqueAttributionKey] = userSessionDataNew
		}
	}
	err := rows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in ProcessEventRows")
		return err
	}
	logCtx.WithFields(log.Fields{"AttributionKey": query.AttributionKey}).
		Info("no document was found in any of the reports for ID. Logging and continuing %+v",
			missingIDs[:U.MinInt(100, len(missingIDs))])
	U.LogReadTimeWithQueryRequestID(startReadTime, queryID, &log.Fields{})
	return nil
}

func ProcessRow(rows *sql.Rows, reportName string, logCtx *log.Entry,
	channel string, queryID string) (map[string]MarketingData, []MarketingData) {

	// ID is CampaignID, AdgroupID, KeywordID etc
	marketingDataIDMap := make(map[string]MarketingData)
	var allRows []MarketingData

	startReadTime := time.Now()
	for rows.Next() {
		var campaignIDNull sql.NullString
		var adgroupIDNull sql.NullString
		var keywordIDNull sql.NullString
		var adIDNull sql.NullString
		var keyIDNull sql.NullString
		var keyNameNull sql.NullString
		var extraValue1Null sql.NullString
		var impressionsNull sql.NullFloat64
		var clicksNull sql.NullFloat64
		var spendNull sql.NullFloat64
		var sourceNull sql.NullString
		if channel == CustomAdsIntegration {
			if err := rows.Scan(&campaignIDNull, &adgroupIDNull, &keywordIDNull, &adIDNull, &keyIDNull, &keyNameNull, &extraValue1Null,
				&impressionsNull, &clicksNull, &spendNull, &sourceNull); err != nil {
				logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
				continue
			}
		} else {
			if err := rows.Scan(&campaignIDNull, &adgroupIDNull, &keywordIDNull, &adIDNull, &keyIDNull, &keyNameNull, &extraValue1Null,
				&impressionsNull, &clicksNull, &spendNull); err != nil {
				logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
				continue
			}
		}
		if !keyNameNull.Valid || !keyIDNull.Valid {
			continue
		}
		ID, data := getMarketingDataFromValues(campaignIDNull, adgroupIDNull, keywordIDNull, adIDNull,
			keyIDNull, keyNameNull, extraValue1Null, impressionsNull, clicksNull, spendNull, reportName)
		if ID == "" {
			continue
		}
		if channel == CustomAdsIntegration {
			data.Channel = U.IfThenElse(sourceNull.String != "", sourceNull.String, PropertyValueNone).(string)
		} else {
			data.Channel = channel
		}
		allRows = append(allRows, data)
		if _, ok := marketingDataIDMap[ID]; ok {
			data = mergeMarketingData(marketingDataIDMap[ID], data)
		}
		marketingDataIDMap[ID] = data
	}
	err := rows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in ProcessRow")
		return nil, nil
	}
	U.LogReadTimeWithQueryRequestID(startReadTime, queryID, &log.Fields{})
	return marketingDataIDMap, allRows
}

func getMarketingDataFromValues(campaignIDNull sql.NullString, adgroupIDNull sql.NullString, keywordIDNull sql.NullString,
	adIDNull sql.NullString, IDNull sql.NullString, nameNull sql.NullString, extraValue1Null sql.NullString, impressionsNull sql.NullFloat64,
	clicksNull sql.NullFloat64, spendNull sql.NullFloat64, reportName string) (string, MarketingData) {

	campaignID := PropertyValueNone
	adgroupID := PropertyValueNone
	keywordID := PropertyValueNone
	adID := PropertyValueNone
	extraValue1 := PropertyValueNone
	var impressions float64
	var clicks float64
	var spend float64
	name := nameNull.String
	ID := IDNull.String
	impressions = 0
	clicks = 0
	spend = 0
	if impressionsNull.Valid {
		impressions = impressionsNull.Float64
	}
	if clicksNull.Valid {
		clicks = clicksNull.Float64
	}
	if spendNull.Valid {
		spend = spendNull.Float64
	}
	if impressions == 0 && clicks == 0 && spend == 0 {
		return "", MarketingData{}
	}
	if extraValue1Null.Valid {
		extraValue1 = U.IfThenElse(extraValue1Null.String != "", extraValue1Null.String, PropertyValueNone).(string)
	}
	if campaignIDNull.Valid {
		campaignID = U.IfThenElse(campaignIDNull.String != "", campaignIDNull.String, PropertyValueNone).(string)
	}
	if adgroupIDNull.Valid {
		adgroupID = U.IfThenElse(adgroupIDNull.String != "", adgroupIDNull.String, PropertyValueNone).(string)
	}
	if keywordIDNull.Valid {
		keywordID = U.IfThenElse(keywordIDNull.String != "", keywordIDNull.String, PropertyValueNone).(string)
	}
	if adIDNull.Valid {
		adID = U.IfThenElse(adIDNull.String != "", adIDNull.String, PropertyValueNone).(string)
	}

	// Only fill IDs. Key and Names would be set separately.
	data := MarketingData{
		Key:              "",
		ID:               ID,
		Name:             name,
		CampaignID:       campaignID,
		CampaignName:     PropertyValueNone,
		AdgroupID:        adgroupID,
		AdgroupName:      PropertyValueNone,
		KeywordMatchType: extraValue1,
		KeywordName:      PropertyValueNone,
		KeywordID:        keywordID,
		AdName:           PropertyValueNone,
		AdID:             adID,
		Slot:             PropertyValueNone,
		Impressions:      int64(impressions),
		Clicks:           int64(clicks),
		Spend:            spend}

	switch reportName {
	case ReportCampaign:
		data.CampaignName = name
	case ReportAdGroup:
		data.AdgroupName = name
	case ReportKeyword:
		data.KeywordName = name
	}
	return ID, data
}

// mergeMarketingData combines values in two MarketingData rows having same marketing id but different names
func mergeMarketingData(marketingDataOld MarketingData, marketingDataNew MarketingData) MarketingData {

	data := marketingDataNew
	data.Impressions = marketingDataOld.Impressions + marketingDataNew.Impressions
	data.Clicks = marketingDataOld.Clicks + marketingDataNew.Clicks
	data.Spend = marketingDataOld.Spend + marketingDataNew.Spend
	return data
}

func AddCustomDimensions(attributionData *map[string]*AttributionData, query *AttributionQuery, reports *MarketingReports) {

	// Custom Dimensions are support only for Campaign,Source,Adgroup currently
	if query.AttributionKey != AttributionKeyCampaign && query.AttributionKey != AttributionKeyAdgroup {
		return
	}

	// Return if extra Custom Dimensions not required
	if !isExtraDimensionRequired(query) {
		return
	}

	if query.AttributionKey == AttributionKeyCampaign {
		enrichDimensionsWithoutChannel(attributionData, query.AttributionKeyCustomDimension, reports.AdwordsCampaignDimensions, reports.FacebookCampaignDimensions, reports.LinkedinCampaignDimensions, reports.BingadsCampaignDimensions, reports.CustomAdsCampaignDimensions, query.AttributionKey)
	} else if query.AttributionKey == AttributionKeyAdgroup {
		enrichDimensionsWithoutChannel(attributionData, query.AttributionKeyCustomDimension, reports.AdwordsAdgroupDimensions, reports.FacebookAdgroupDimensions, reports.LinkedinAdgroupDimensions, reports.BingadsAdgroupDimensions, reports.CustomAdsAdgroupDimensions, query.AttributionKey)
	}
}

func enrichDimensionsWithoutChannel(attributionData *map[string]*AttributionData, dimensions []string, adwordsData, fbData,
	linkedinData, bingadsData, customAdsData map[string]MarketingData, attributionKey string) {

	for _, dim := range dimensions {
		for k, v := range *attributionData {

			if (*attributionData)[k].CustomDimensions == nil {
				(*attributionData)[k].CustomDimensions = make(map[string]interface{})
			}
			(*attributionData)[k].CustomDimensions[dim] = PropertyValueNone

			customDimKey := GetKeyForCustomDimensionsName(v.MarketingInfo.CampaignID, v.MarketingInfo.AdgroupID, attributionKey)
			if customDimKey == "" {
				continue
			}

			if d, exists := adwordsData[customDimKey]; exists {
				if val, found := d.CustomDimensions[dim]; found {
					(*attributionData)[k].CustomDimensions[dim] = val
					continue
				}
			}

			if d, exists := fbData[customDimKey]; exists {
				if val, found := d.CustomDimensions[dim]; found {
					(*attributionData)[k].CustomDimensions[dim] = val
					continue
				}
			}

			if d, exists := linkedinData[customDimKey]; exists {
				if val, found := d.CustomDimensions[dim]; found {
					(*attributionData)[k].CustomDimensions[dim] = val
					continue
				}
			}

			if d, exists := bingadsData[customDimKey]; exists {
				if val, found := d.CustomDimensions[dim]; found {
					(*attributionData)[k].CustomDimensions[dim] = val
					continue
				}
			}

			if d, exists := customAdsData[customDimKey]; exists {
				if val, found := d.CustomDimensions[dim]; found {
					(*attributionData)[k].CustomDimensions[dim] = val
					continue
				}
			}

		}
	}
}

func GetKeyForCustomDimensionsName(campaignId, adgroupId, attributionKey string) string {

	key := ""
	if attributionKey == AttributionKeyCampaign {
		key = campaignId
	} else if attributionKey == AttributionKeyAdgroup {
		key = adgroupId
	}
	return key
}

func isExtraDimensionRequired(query *AttributionQuery) bool {
	defaultDimensionsMap := make(map[string]int)
	defaultDimensionsMap[FieldCampaignName] = 1
	defaultDimensionsMap[FieldAdgroupName] = 1
	extraDimensionsRequired := false
	for _, dim := range query.AttributionKeyCustomDimension {
		if defaultDimensionsMap[dim] == 0 {
			extraDimensionsRequired = true
			break
		}
	}
	return extraDimensionsRequired
}

func IfValidGetValElseNone(value sql.NullString) string {

	if value.Valid && value.String != "" {
		return value.String
	}
	return PropertyValueNone
}

// MergeUsersToBeAttributed merges users to be attributed for goal and linked event
func MergeUsersToBeAttributed(goalEventUsers *[]UserEventInfo, funnelEventUsers []UserEventInfo) {

	goalHitTime := make(map[string]int64)

	for _, userInfo := range *goalEventUsers {
		goalHitTime[userInfo.CoalUserID] = userInfo.Timestamp
	}

	for _, userInfo := range funnelEventUsers {
		if _, exists := goalHitTime[userInfo.CoalUserID]; exists && userInfo.Timestamp >= goalHitTime[userInfo.CoalUserID] {
			*goalEventUsers = append(*goalEventUsers, userInfo)
		}
	}
}

func GetContentGroupNamesToDummyNamesMap(contentGroupNamesList []string) map[string]string {
	contentGroupNamesToDummyNamesMap := make(map[string]string)
	for index, contentGroupName := range contentGroupNamesList {
		contentGroupNamesToDummyNamesMap[contentGroupName] = "contentGroup_" + fmt.Sprintf("%d", index)
	}
	return contentGroupNamesToDummyNamesMap
}
func isSessionWithinQueryPeriod(queryType string,
	lookBackWindow int, from int64, to int64, timestamp int64) bool {
	lookbackPeriod := int64(lookBackWindow) * SecsInADay
	switch queryType {

	case AttributionQueryTypeConversionBased:
		if timestamp >= from-lookbackPeriod && timestamp <= to {
			return true
		}
		return false

	case AttributionQueryTypeEngagementBased:
		if timestamp >= from && timestamp <= to {
			return true
		}
		return false
	}
	return false
}

func EnrichRequestUsingAttributionConfig(query *AttributionQuery, settings *ProjectSetting, logCtx *log.Entry) error {

	attributionConfig, err1 := decodeAttributionConfig(settings.AttributionConfig)
	if err1 != nil {
		return errors.New("failed to decode attribution config from project settings")
	}

	//Todo (Anil) Add enrichment of attribution Window, handle case of 'Entire User Journey'

	switch query.AnalyzeType {

	case AnalyzeTypeUsers:
		query.RunType = RunTypeUser
		return nil
	case AnalyzeTypeUserKPI:
		query.RunType = RunTypeUserKPI
		return nil
	case AnalyzeTypeHSDeals:
		if &attributionConfig != nil && attributionConfig.AnalyzeTypeHSCompaniesEnabled == true {
			query.RunType = RunTypeHSCompanies
			return nil
		} else if &attributionConfig != nil && attributionConfig.AnalyzeTypeHSDealsEnabled == true {
			query.RunType = RunTypeHSDeals
			return nil
		} else {
			logCtx.WithFields(log.Fields{"Query": query, "AttributionConfig": attributionConfig}).Error("Failed to set analyze type")
			return errors.New("invalid config/query. Failed to set analyze type from attribution config & project settings")
		}
	case AnalyzeTypeSFOpportunities:
		if &attributionConfig != nil && attributionConfig.AnalyzeTypeSFAccountsEnabled == true {
			query.RunType = RunTypeSFAccounts
			return nil
		} else if &attributionConfig != nil && attributionConfig.AnalyzeTypeSFOpportunitiesEnabled == true {
			query.RunType = RunTypeSFOpportunities
			return nil
		} else {
			logCtx.WithFields(log.Fields{"Query": query, "AttributionConfig": attributionConfig}).Error("Failed to set analyze type")
			return errors.New("invalid config/query. Failed to set analyze type from attribution config & project settings")
		}
	default:
		query.AnalyzeType = AnalyzeTypeUsers
		query.RunType = RunTypeUser
		return nil
		// logCtx.WithFields(log.Fields{"Query": query, "AttributionConfig": attributionConfig}).Error("Failed to set analyze type")
		// return errors.New("invalid config/query. Failed to set analyze type from attribution config & project settings")
	}
}

// decodeAttributionConfig decode attribution config from project settings to map
func decodeAttributionConfig(config *postgres.Jsonb) (AttributionConfig, error) {
	attributionConfig := AttributionConfig{}
	if config == nil {
		return attributionConfig, nil
	}

	err := json.Unmarshal(config.RawMessage, &attributionConfig)
	if err != nil {
		return attributionConfig, err
	}

	return attributionConfig, nil
}
