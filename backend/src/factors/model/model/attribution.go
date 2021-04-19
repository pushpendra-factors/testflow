package model

import (
	cacheRedis "factors/cache/redis"
	U "factors/util"
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

	AttributionQueryTypeConversionBased = "ConversionBased"
	AttributionQueryTypeEngagementBased = "EngagementBased"

	SortASC  = "ASC"
	SortDESC = "DESC"
)

type UserSessionTimestamp struct {
	MinTimestamp      int64
	MaxTimestamp      int64
	TimeStamps        []int64
	WithinQueryPeriod bool
}

var AttributionFixedHeaders = []string{"Impressions", "Clicks", "Spend", "Website Visitors"}

type AttributionData struct {
	Name                        string
	Impressions                 int64
	Clicks                      int64
	Spend                       float64
	WebsiteVisitors             int64
	ConversionEventCount        float64
	LinkedEventsCount           []float64
	ConversionEventCompareCount float64
}

type UserInfo struct {
	CoalUserID   string
	PropertiesID string
	Timestamp    int64
}

type UserIDPropID struct {
	UserID       string
	PropertiesID string
	Timestamp    int64
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
func MergeDataRowsHavingSameKey(rows [][]interface{}) [][]interface{} {

	rowKeyMap := make(map[string][]interface{})
	keyIndex := 0
	for _, row := range rows {
		key := row[keyIndex].(string)
		if _, exists := rowKeyMap[key]; exists {
			seenRow := rowKeyMap[key]
			seenRow[1] = seenRow[1].(int64) + row[1].(int64)     // Impressions.
			seenRow[2] = seenRow[2].(int64) + row[2].(int64)     // Clicks.
			seenRow[3] = seenRow[3].(float64) + row[3].(float64) // Spend.
			seenRow[4] = seenRow[4].(int64) + row[4].(int64)     // Website Visitors.
			seenRow[5] = seenRow[5].(float64) + row[5].(float64) // Conversion.
			seenRow[6] = seenRow[6].(float64) + row[6].(float64) // Conversion - CPC.
			seenRow[7] = seenRow[7].(float64) + row[7].(float64) // Compare Conversion.
			seenRow[8] = seenRow[8].(float64) + row[8].(float64) // Compare Conversion - CPC.
			// Remaining linked funnel events & CPCs
			for i := 9; i < len(seenRow); i++ {
				seenRow[i] = seenRow[i].(float64) + row[i].(float64)
			}
			rowKeyMap[key] = seenRow
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

// GetGCLIDAttributionValue Returns the matching value for GCLID, if not found returns $none
func GetGCLIDAttributionValue(gclIDBasedCampaign map[string]CampaignInfo, gclID string, attributionKey string) string {

	if value, ok := gclIDBasedCampaign[gclID]; ok {
		switch attributionKey {
		case U.EP_ADGROUP:
			if value.AdgroupName != PropertyValueNone {
				return value.AdgroupName
			}
			return value.AdgroupID
		case U.EP_CAMPAIGN:
			if value.CampaignName != PropertyValueNone {
				return value.CampaignName
			}
			return value.CampaignID
		case U.EP_KEYWORD:
			return value.KeywordID
		default:
			// No enrichment for Source via GCLID
			return PropertyValueNone
		}
	}
	return PropertyValueNone
}
