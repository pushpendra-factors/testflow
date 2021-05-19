package model

import (
	"database/sql"
	"errors"
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

type AttributionQuery struct {
	CampaignMetrics               []string                   `json:"cm"`
	ConversionEvent               QueryEventWithProperties   `json:"ce"`
	ConversionEventCompare        QueryEventWithProperties   `json:"ce_c"`
	LinkedEvents                  []QueryEventWithProperties `json:"lfe"`
	AttributionKey                string                     `json:"attribution_key"`
	AttributionKeyFilter          []AttributionKeyFilter     `json:"attribution_key_f"`
	AttributionMethodology        string                     `json:"attribution_methodology"`
	AttributionMethodologyCompare string                     `json:"attribution_methodology_c"`
	LookbackDays                  int                        `json:"lbw"`
	From                          int64                      `json:"from"`
	To                            int64                      `json:"to"`
	QueryType                     string                     `json:"query_type"`
	Timezone                      string                     `json:"time_zone"`
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

func (q *AttributionQueryUnit) GetQueryCacheRedisKey(projectID uint64) (*cacheRedis.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.Query.From, q.Query.To)
	return cacheRedis.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *AttributionQueryUnit) GetQueryCacheExpiry() float64 {
	return getQueryCacheResultExpiry(q.Query.From, q.Query.To)
}

const (
	AttributionMethodFirstTouch          = "First_Touch"
	AttributionMethodFirstTouchNonDirect = "First_Touch_ND"
	AttributionMethodLastTouch           = "Last_Touch"
	AttributionMethodLastTouchNonDirect  = "Last_Touch_ND"
	AttributionMethodLinear              = "Linear"
	AttributionKeyCampaign               = "Campaign"
	AttributionKeySource                 = "Source"
	AttributionKeyAdgroup                = "AdGroup"
	AttributionKeyKeyword                = "Keyword"

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

	FacebookCampaignID   = "campaign_id"
	FacebookCampaignName = "campaign_name"

	FacebookAdgroupID   = "adset_id"
	FacebookAdgroupName = "adset_name"

	LinkedinCampaignID   = "campaign_group_id"
	LinkedinCampaignName = "campaign_group_name"

	LinkedinAdgroupID   = "campaign_id"
	LinkedinAdgroupName = "campaign_name"

	KeyDelimiter = ":-:"
)

type MarketingReports struct {
	AdwordsGCLIDData      map[string]MarketingData
	AdwordsCampaignIDData map[string]MarketingData
	//AdwordsCampaignNameData map[string]MarketingData
	AdwordsCampaignKeyData map[string]MarketingData

	AdwordsAdgroupIDData map[string]MarketingData
	//AdwordsAdgroupNameData map[string]MarketingData
	AdwordsAdgroupKeyData map[string]MarketingData

	AdwordsKeywordIDData map[string]MarketingData
	//AdwordsKeywordNameData map[string]MarketingData
	AdwordsKeywordKeyData map[string]MarketingData

	FacebookCampaignIDData map[string]MarketingData
	//FacebookCampaignNameData map[string]MarketingData
	FacebookCampaignKeyData map[string]MarketingData

	FacebookAdgroupIDData map[string]MarketingData
	//FacebookAdgroupNameData map[string]MarketingData
	FacebookAdgroupKeyData map[string]MarketingData

	LinkedinCampaignIDData map[string]MarketingData
	//LinkedinCampaignNameData map[string]MarketingData
	LinkedinCampaignKeyData map[string]MarketingData

	LinkedinAdgroupIDData map[string]MarketingData
	//LinkedinAdgroupNameData map[string]MarketingData
	LinkedinAdgroupKeyData map[string]MarketingData
}
type MarketingData struct {
	// Key is CampaignName + AdgroupName + KeywordName + MatchType (i.e. ExtraValue)
	Key string
	// CampaignID, AdgroupID etc
	ID string
	// CampaignName, AdgroupName etc
	Name string
	// For Adwords Keyword Perf report, it is keyword_match_type, for others it is $none
	CampaignID       string
	CampaignName     string
	AdgroupID        string
	AdgroupName      string
	KeywordMatchType string
	KeywordID        string
	KeywordName      string
	AdID             string
	AdName           string
	Slot             string
	Source           string
	Impressions      int64
	Clicks           int64
	Spend            float64
}

type UserSessionData struct {
	MinTimestamp      int64
	MaxTimestamp      int64
	TimeStamps        []int64
	SessionSpentTimes []float64
	PageCounts        []int64
	WithinQueryPeriod bool
	MarketingInfo     MarketingData
}

var AddedKeysForAdgroup = []string{"Campaign"}
var AddedKeysForKeyword = []string{"Campaign", "AdGroup", "MatchType"}

// TODO (Anil) Update Website Visitors to Session back
var AttributionFixedHeaders = []string{"Impressions", "Clicks", "Spend", "CTR", "Average CPC", "CPM", "ConversionRate", "Website Visitors", "Users", "Average Session Time", "PageViews"}

// var AttributionFixedHeaders = []string{"Impressions", "Clicks", "Spend", "CTR", "Average CPC", "CPM", "ConversionRate", "Sessions", "Users", "Average Session Time", "PageViews"}
var AttributionFixedHeadersPostPostConversion = []string{"Cost Per Conversion", "Compare - Users", "Compare Cost Per Conversion"}

type AttributionData struct {
	AddedKeys                     []string
	Name                          string
	Impressions                   int64
	Clicks                        int64
	Spend                         float64
	CTR                           float64
	AvgCPC                        float64
	CPM                           float64
	ConversionRate                float64
	Sessions                      int64
	Users                         int64
	AvgSessionTime                float64
	PageViews                     int64
	ConversionEventCount          float64
	CostPerConversion             float64
	LinkedEventsCount             []float64
	ConversionEventCompareCount   float64
	CostPerConversionCompareCount float64
	MarketingInfo                 MarketingData
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
}

const (
	AdwordsClickReportType = 4
	SecsInADay             = int64(86400)
	LookbackCapInDays      = 180
	UserBatchSize          = 3000
)

// MergeDataRowsHavingSameKey merges rows having same key by adding each column value
func MergeDataRowsHavingSameKey(rows [][]interface{}, keyIndex int) [][]interface{} {

	logCtx := log.WithFields(log.Fields{"MISMATCH": "MergeDataRowsHavingSameKey"})

	rowKeyMap := make(map[string][]interface{})
	for _, row := range rows {
		if len(row) == 0 {
			continue
		}
		// creating a key for using added keys and index
		key := ""
		for j := 0; j <= keyIndex; j++ {
			key = key + row[j].(string)
		}

		if _, exists := rowKeyMap[key]; exists {
			seenRow := rowKeyMap[key]
			// Don't sum up Impressions, Clicks, Spend.
			seenRow[keyIndex+1] = U.Max(seenRow[keyIndex+1].(int64), row[keyIndex+1].(int64))            // Impressions.
			seenRow[keyIndex+2] = U.Max(seenRow[keyIndex+2].(int64), row[keyIndex+2].(int64))            // Clicks.
			seenRow[keyIndex+3] = U.MaxFloat64(seenRow[keyIndex+3].(float64), row[keyIndex+3].(float64)) // Spend.

			seenRow[keyIndex+4] = U.MaxFloat64(seenRow[keyIndex+4].(float64), row[keyIndex+4].(float64)) // CTR.
			seenRow[keyIndex+5] = U.MaxFloat64(seenRow[keyIndex+5].(float64), row[keyIndex+5].(float64)) // AvgCPC.
			seenRow[keyIndex+6] = U.MaxFloat64(seenRow[keyIndex+6].(float64), row[keyIndex+6].(float64)) // CPM.
			seenRow[keyIndex+7] = U.MaxFloat64(seenRow[keyIndex+7].(float64), row[keyIndex+7].(float64)) // ConversionRate.

			seenRow[keyIndex+8] = seenRow[keyIndex+8].(int64) + row[keyIndex+8].(int64) // Sessions.
			seenRow[keyIndex+9] = seenRow[keyIndex+9].(int64) + row[keyIndex+9].(int64) // Users.

			seenRow[keyIndex+10] = U.MaxFloat64(seenRow[keyIndex+10].(float64), row[keyIndex+10].(float64)) // AvgSessionTime.
			seenRow[keyIndex+11] = seenRow[keyIndex+11].(int64) + row[keyIndex+11].(int64)                  // PageViews.

			seenRow[keyIndex+12] = seenRow[keyIndex+12].(float64) + row[keyIndex+12].(float64)              // Conversion.
			seenRow[keyIndex+13] = U.MaxFloat64(seenRow[keyIndex+13].(float64), row[keyIndex+13].(float64)) // Conversion - CPC.
			seenRow[keyIndex+14] = seenRow[keyIndex+14].(float64) + row[keyIndex+14].(float64)              // Compare Conversion.
			seenRow[keyIndex+15] = U.MaxFloat64(seenRow[keyIndex+15].(float64), row[keyIndex+15].(float64)) // Compare Conversion - CPC.
			// Remaining linked funnel events & CPCs
			for i := keyIndex + 16; i < len(seenRow); i++ {
				seenRow[i] = seenRow[i].(float64) + row[i].(float64)
			}
			rowKeyMap[key] = seenRow
		} else {
			rowKeyMap[key] = row
		}

		logCtx.WithFields(log.Fields{
			"Key":         key,
			"Name":        rowKeyMap[key][keyIndex],
			"Impressions": rowKeyMap[key][keyIndex+1],
			"Clicks":      rowKeyMap[key][keyIndex+2],
			"Spend":       rowKeyMap[key][keyIndex+3],
		}).Info("Original Value from ReportKeyData")

	}
	resultRows := make([][]interface{}, 0)
	for _, mapRow := range rowKeyMap {
		resultRows = append(resultRows, mapRow)
	}
	return resultRows
}

// GetGCLIDAttributionValue Returns the matching value for GCLID, if not found returns $none
func GetGCLIDAttributionValue(gclIDBasedCampaign *map[string]MarketingData, gclID string,
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
	} else if attributionKey == AttributionKeyAdgroup {
		return U.EP_ADGROUP, nil
	} else if attributionKey == AttributionKeyKeyword {
		return U.EP_KEYWORD, nil
	}
	return "", errors.New("invalid query properties")
}

// AddHeadersByAttributionKey Adds common column names and linked events as header to the result rows.
func AddHeadersByAttributionKey(result *QueryResult, query *AttributionQuery) {

	attributionKey := query.AttributionKey

	// Add up for Added Keys {Campaign, Adgroup, Keyword}
	switch attributionKey {
	case AttributionKeyCampaign:
		// No keys are added.
	case AttributionKeyAdgroup:
		result.Headers = append(result.Headers, AddedKeysForAdgroup...)
	case AttributionKeyKeyword:
		result.Headers = append(result.Headers, AddedKeysForKeyword...)
	default:
	}

	result.Headers = append(append(result.Headers, attributionKey), AttributionFixedHeaders...)
	conversionEventUsers := fmt.Sprintf("%s - Users", query.ConversionEvent.Name)
	result.Headers = append(result.Headers, conversionEventUsers)
	result.Headers = append(result.Headers, AttributionFixedHeadersPostPostConversion...)
	if len(query.LinkedEvents) > 0 {
		for _, event := range query.LinkedEvents {
			result.Headers = append(result.Headers, fmt.Sprintf("%s - Users", event.Name))
			result.Headers = append(result.Headers, fmt.Sprintf("%s - CPC", event.Name))
		}
	}
}

// getLinkedEventColumnAsInterfaceList return interface list having linked event count and CPC
func getLinkedEventColumnAsInterfaceList(spend float64, data []float64) []interface{} {
	var list []interface{}
	for _, val := range data {
		cpc := 0.0
		if val != 0.0 {
			cpc, _ = U.FloatRoundOffWithPrecision(spend/val, U.DefaultPrecision)
		}
		list = append(list, val, cpc)
	}
	return list
}

func GetKeyIndexOrAddedKeySize(attributionKey string) int {

	addedKeysSize := 0
	// Add up for Added Keys {Campaign, Adgroup, Keyword}
	switch attributionKey {
	case AttributionKeyCampaign:
		addedKeysSize = 0
	case AttributionKeyAdgroup:
		addedKeysSize = 1
	case AttributionKeyKeyword:
		addedKeysSize = 3
	default:
	}
	return addedKeysSize
}

func GetConversionIndex(headers []string) int {
	for index, val := range headers {
		if val == "PageViews" {
			return index + 1
		}
	}
	return 0
}

func GetSpendIndex(headers []string) int {
	for index, val := range headers {
		if val == "Spend" {
			return index + 1
		}
	}
	return 0
}

// GetRowsByMaps Returns result in from of metrics. For empty attribution id, the values are accumulated into "$none".
func GetRowsByMaps(attributionKey string, attributionData *map[string]*AttributionData,
	linkedEvents []QueryEventWithProperties, isCompare bool) [][]interface{} {

	// Name, impression, clicks, spend
	defaultMatchingRow := []interface{}{"none", int64(0), int64(0), float64(0),
		// (CTR, AvgCPC, CPM, ConversionRate)
		float64(0), float64(0), float64(0), float64(0),
		// Sessions, (users), (AvgSessionTime), (pageViews),
		int64(0), int64(0), float64(0), int64(0),
		// ConversionEventCount, CostPerConversion, ConversionEventCompareCount, CostPerConversionCompareCount
		float64(0), float64(0), float64(0), float64(0)}
	var nonMatchingRow []interface{}

	addedKeysSize := 0
	// Add up for Added Keys {Campaign, Adgroup, Keyword}
	switch attributionKey {
	case AttributionKeyCampaign:
		nonMatchingRow = defaultMatchingRow
	case AttributionKeyAdgroup:
		addedKeysSize = 1
		nonMatchingRow = append([]interface{}{"none"}, defaultMatchingRow...)
	case AttributionKeyKeyword:
		addedKeysSize = 3
		nonMatchingRow = append([]interface{}{"none", "none", "none"}, defaultMatchingRow...)
	default:
	}

	// Add up for linkedEvents for conversion and CPC
	for i := 0; i < len(linkedEvents); i++ {
		nonMatchingRow = append(nonMatchingRow, float64(0), float64(0))
	}
	rows := make([][]interface{}, 0)
	for key, data := range *attributionData {
		attributionIdName := ""
		switch attributionKey {
		case AttributionKeyCampaign:
			attributionIdName = key
		case AttributionKeyAdgroup:
			attributionIdName = data.MarketingInfo.AdgroupName
		case AttributionKeyKeyword:
			attributionIdName = data.MarketingInfo.KeywordName
		case AttributionKeySource:
			attributionIdName = data.MarketingInfo.Source
		default:
		}
		if attributionIdName == "" {
			attributionIdName = PropertyValueNone
		}
		if attributionIdName != "" {
			var row []interface{}
			for i := 0; i < addedKeysSize; i++ {
				if data.AddedKeys != nil && data.AddedKeys[i] != "" {
					row = append(row, data.AddedKeys[i])
				} else {
					row = append(row, PropertyValueNone)
				}
			}
			row = append(row, attributionIdName, data.Impressions, data.Clicks, data.Spend,
				data.CTR, data.AvgCPC, data.CPM, data.ConversionRate, data.Sessions, data.Users, data.AvgSessionTime, data.PageViews, data.ConversionEventCount)
			cpc := 0.0
			if data.ConversionEventCount != 0.0 {
				cpc, _ = U.FloatRoundOffWithPrecision(data.Spend/data.ConversionEventCount, U.DefaultPrecision)
			}
			if isCompare {
				cpcCompare := 0.0
				if data.ConversionEventCompareCount != 0.0 {
					cpcCompare, _ = U.FloatRoundOffWithPrecision(data.Spend/data.ConversionEventCompareCount, U.DefaultPrecision)
				}
				row = append(row, cpc, data.ConversionEventCompareCount, cpcCompare)
			} else {
				row = append(row, cpc, float64(0), float64(0))
			}
			row = append(row, getLinkedEventColumnAsInterfaceList(data.Spend, data.LinkedEventsCount)...)
			rows = append(rows, row)
		}
	}
	rows = append(rows, nonMatchingRow)
	return rows
}

// AddUpConversionEventCount Groups all unique users by attributionId and adds it to attributionData
func AddUpConversionEventCount(usersIdAttributionIdMap map[string][]string) map[string]*AttributionData {
	attributionData := make(map[string]*AttributionData)
	for _, attributionKeys := range usersIdAttributionIdMap {
		weight := 1 / float64(len(attributionKeys))
		for _, key := range attributionKeys {
			if _, exists := attributionData[key]; !exists {
				attributionData[key] = &AttributionData{}
			}
			attributionData[key].ConversionEventCount += weight
		}
	}
	return attributionData
}

// AddUpLinkedFunnelEventCount Attribute each user to the conversion event and linked event by attribution Id.
func AddUpLinkedFunnelEventCount(linkedEvents []QueryEventWithProperties,
	attributionData map[string]*AttributionData, linkedUserAttributionData map[string]map[string][]string) {

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
		}
	}
	// Update linked up events with event hit count.
	for linkedEventName, userIdAttributionIdMap := range linkedUserAttributionData {
		for _, attributionKeys := range userIdAttributionIdMap {
			weight := 1 / float64(len(attributionKeys))
			for _, key := range attributionKeys {
				if attributionData[key] != nil {
					attributionData[key].LinkedEventsCount[linkedEventToPositionMap[linkedEventName]] += weight
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
		attributionKey == AttributionKeyKeyword {
		return true
	}
	return false
}

func DoesFBReportExist(attributionKey string) bool {
	// only campaign, adgroup reports available
	if attributionKey == AttributionKeyCampaign || attributionKey == AttributionKeyAdgroup {
		return true
	}
	return false
}

func DoesLinkedinReportExist(attributionKey string) bool {
	// only campaign, adgroup reports available
	if attributionKey == AttributionKeyCampaign || attributionKey == AttributionKeyAdgroup {
		return true
	}
	return false
}

func AddTheAddedKeysAndMetrics(attributionData *map[string]*AttributionData, attributionKey string, sessions map[string]map[string]UserSessionData) {

	// Extract out key based info
	sessionKeyMarketingInfo := make(map[string]MarketingData)
	sessionKeySessionTimes := make(map[string][]float64)
	sessionKeyPageCounts := make(map[string][]int64)
	sessionKeyUserCount := make(map[string]int64)
	for _, value := range sessions {
		userKeyPairCounted := false
		for k, v := range value {
			sessionKeyMarketingInfo[k] = v.MarketingInfo
			sessionKeySessionTimes[k] = append(sessionKeySessionTimes[k], v.SessionSpentTimes...)
			sessionKeyPageCounts[k] = append(sessionKeyPageCounts[k], v.PageCounts...)
			// Any one instance of this user
			if !userKeyPairCounted {
				sessionKeyUserCount[k] = sessionKeyUserCount[k] + 1
				userKeyPairCounted = true
			}
		}
	}

	for key, _ := range *attributionData {
		if _, exists := sessionKeyMarketingInfo[key]; exists {
			// Add the marketing info
			(*attributionData)[key].MarketingInfo = sessionKeyMarketingInfo[key]
			switch attributionKey {
			case AttributionKeyCampaign:
				(*attributionData)[key].Name = sessionKeyMarketingInfo[key].CampaignName
			case AttributionKeyAdgroup:
				(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, sessionKeyMarketingInfo[key].CampaignName)
				(*attributionData)[key].Name = sessionKeyMarketingInfo[key].AdgroupName
			case AttributionKeyKeyword:
				(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, sessionKeyMarketingInfo[key].CampaignName)
				(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, sessionKeyMarketingInfo[key].AdgroupName)
				(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, sessionKeyMarketingInfo[key].KeywordMatchType)
				(*attributionData)[key].Name = sessionKeyMarketingInfo[key].KeywordName
			case AttributionKeySource:
				(*attributionData)[key].Name = sessionKeyMarketingInfo[key].Source
			}
			// Add AvgSessionTime
			totalTime := 0.0
			for _, v := range sessionKeySessionTimes[key] {
				totalTime = totalTime + v
			}
			if totalTime != 0 && len(sessionKeySessionTimes[key]) != 0 {
				(*attributionData)[key].AvgSessionTime = totalTime / float64(len(sessionKeySessionTimes[key]))
			}
			// Add PageViews
			totalPageCount := int64(0)
			for _, v := range sessionKeyPageCounts[key] {
				totalPageCount = totalPageCount + v
			}
			(*attributionData)[key].PageViews = totalPageCount

			// Add Unique user count
			(*attributionData)[key].Users = sessionKeyUserCount[key]

		}

	}
}

func AddPerformanceData(attributionData *map[string]*AttributionData, attributionKey string, marketingData *MarketingReports) {

	AddAdwordsPerformanceReportInfo(attributionData, attributionKey, marketingData)
	AddFacebookPerformanceReportInfo(attributionData, attributionKey, marketingData)
	AddLinkedinPerformanceReportInfo(attributionData, attributionKey, marketingData)
}

func AddAdwordsPerformanceReportInfo(attributionData *map[string]*AttributionData, attributionKey string, marketingData *MarketingReports) {

	switch attributionKey {
	case AttributionKeyCampaign:
		addMetricsFromReport(attributionData, marketingData.AdwordsCampaignKeyData, attributionKey)
	case AttributionKeyAdgroup:
		addMetricsFromReport(attributionData, marketingData.AdwordsAdgroupKeyData, attributionKey)
	case AttributionKeyKeyword:
		addMetricsFromReport(attributionData, marketingData.AdwordsKeywordKeyData, attributionKey)
	default:
		// no enrichment for any other type
		return
	}
}

func AddFacebookPerformanceReportInfo(attributionData *map[string]*AttributionData, attributionKey string, marketingData *MarketingReports) {

	switch attributionKey {
	case AttributionKeyCampaign:
		addMetricsFromReport(attributionData, marketingData.FacebookCampaignKeyData, attributionKey)
	case AttributionKeyAdgroup:
		addMetricsFromReport(attributionData, marketingData.FacebookAdgroupKeyData, attributionKey)
	case AttributionKeyKeyword:
		// No keyword report for fb.
		return
	default:
		// no enrichment for any other type
		return
	}
}

func AddLinkedinPerformanceReportInfo(attributionData *map[string]*AttributionData, attributionKey string, marketingData *MarketingReports) {

	switch attributionKey {
	case AttributionKeyCampaign:
		addMetricsFromReport(attributionData, marketingData.LinkedinCampaignKeyData, attributionKey)
	case AttributionKeyAdgroup:
		addMetricsFromReport(attributionData, marketingData.LinkedinAdgroupKeyData, attributionKey)
	case AttributionKeyKeyword:
		// No keyword report for Linkedin.
		return
	default:
		// no enrichment for any other type
		return
	}
}

func addMetricsFromReport(attributionData *map[string]*AttributionData, reportKeyData map[string]MarketingData, attributionKey string) {

	for key, value := range reportKeyData {

		if value.Impressions == 0 && value.Clicks == 0 && value.Spend == 0 {
			// ignore ZERO valued keys
			continue
		}
		if _, found := (*attributionData)[key]; found {

			(*attributionData)[key].Impressions = value.Impressions
			(*attributionData)[key].Clicks = value.Clicks
			(*attributionData)[key].Spend = value.Spend

		} else {
			(*attributionData)[key] = &AttributionData{}
			(*attributionData)[key].MarketingInfo = reportKeyData[key]
			switch attributionKey {
			case AttributionKeyCampaign:
				(*attributionData)[key].Name = reportKeyData[key].CampaignName
			case AttributionKeyAdgroup:
				(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, reportKeyData[key].CampaignName)
				(*attributionData)[key].Name = reportKeyData[key].AdgroupName
			case AttributionKeyKeyword:
				(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, reportKeyData[key].CampaignName)
				(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, reportKeyData[key].AdgroupName)
				(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, reportKeyData[key].KeywordMatchType)
				(*attributionData)[key].Name = reportKeyData[key].KeywordName
			case AttributionKeySource:
				(*attributionData)[key].Name = reportKeyData[key].Source
			}
		}
	}
}

func GetKeyByAttributionData(value *AttributionData) interface{} {

	key := ""
	for i := 0; i < len(value.AddedKeys); i++ {
		key = key + value.AddedKeys[i] + KeyDelimiter
	}
	key = key + value.Name
	return key
}

func ComputeAdditionalMetrics(attributionData *map[string]*AttributionData) {

	for k, v := range *attributionData {
		(*attributionData)[k].CTR = 0
		(*attributionData)[k].CPM = 0
		(*attributionData)[k].AvgCPC = 0
		(*attributionData)[k].ConversionRate = 0
		if v.Impressions > 0 {
			(*attributionData)[k].CTR = float64(v.Clicks) / float64(v.Impressions)
			(*attributionData)[k].CPM = 1000 * float64(v.Spend) / float64(v.Impressions)
		}
		if v.Clicks > 0 {
			(*attributionData)[k].AvgCPC = float64(v.Spend) / float64(v.Clicks)
			(*attributionData)[k].ConversionRate = float64(v.ConversionEventCount) / float64(v.Clicks)
		}
	}
}

func GetMarketingDataKey(attributionKey string, data MarketingData) string {

	key := ""
	switch attributionKey {
	case AttributionKeyCampaign:
		// we know we get campaignIDReport here
		key = key + U.IfThenElse(data.Name != "" && data.Name != PropertyValueNone, data.Name, data.CampaignName).(string)
	case AttributionKeyAdgroup:
		key = key + data.CampaignName + KeyDelimiter + U.IfThenElse(data.Name != "" && data.Name != PropertyValueNone, data.Name, data.AdgroupName).(string)
	case AttributionKeyKeyword:
		// we know we get keywordIDReport here
		key = key + data.CampaignName + KeyDelimiter + data.AdgroupName + KeyDelimiter + data.KeywordMatchType + KeyDelimiter + U.IfThenElse(data.Name != "" && data.Name != PropertyValueNone, data.Name, data.KeywordName).(string)
	case AttributionKeySource:
		key = key + U.IfThenElse(data.Name != "" && data.Name != PropertyValueNone, data.Name, data.Source).(string)
	default:
		key = key + data.Name
	}
	return key
}

func GetKeyMapToData(attributionKey string, allRows []MarketingData) map[string]MarketingData {

	keyToData := make(map[string]MarketingData)
	for _, v := range allRows {
		key := GetMarketingDataKey(attributionKey, v)
		val := MarketingData{}
		U.DeepCopy(&v, &val)
		val.Key = key
		keyToData[key] = val
	}
	return keyToData
}

func ProcessRow(rows *sql.Rows, reportName string, logCtx *log.Entry) (map[string]MarketingData, []MarketingData) {

	// ID is CampaignID, AdgroupID, KeywordID etc
	marketingDataIDMap := make(map[string]MarketingData)
	var allRows []MarketingData

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
		if err := rows.Scan(&campaignIDNull, &adgroupIDNull, &keywordIDNull, &adIDNull, &keyIDNull, &keyNameNull, &extraValue1Null,
			&impressionsNull, &clicksNull, &spendNull); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
			continue
		}
		if !keyNameNull.Valid || !keyIDNull.Valid {
			continue
		}
		ID, data := getMarketingDataFromValues(campaignIDNull, adgroupIDNull, keywordIDNull, adIDNull,
			keyIDNull, keyNameNull, extraValue1Null, impressionsNull, clicksNull, spendNull, reportName)
		if ID == "" {
			continue
		}
		marketingDataIDMap[ID] = data
		allRows = append(allRows, data)
	}
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
