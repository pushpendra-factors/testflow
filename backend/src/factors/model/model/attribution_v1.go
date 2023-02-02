package model

import (
	"database/sql"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
)

type AttributionQueryV1 struct {
	KPIQueries             []AttributionKPIQueries    `json:"kpi_queries"`
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

type AttributionKPIQueries struct {
	KPI         KPIQueryGroup `json:"kpi_query_group"`
	AnalyzeType string        `json:"analyze_type"`
	RunType     string        `json:"run_type"`
}

type AttributionQueryUnitV1 struct {
	Class string                  `json:"cl"`
	Query *AttributionQueryV1     `json:"query"`
	Meta  *map[string]interface{} `json:"meta"`
}

func (q *AttributionQueryUnitV1) GetClass() string {
	return q.Class
}

func (q *AttributionQueryUnitV1) GetQueryDateRange() (int64, int64) {
	return q.Query.From, q.Query.To
}

func (q *AttributionQueryUnitV1) SetQueryDateRange(from, to int64) {
	q.Query.From, q.Query.To = from, to
}

func (q *AttributionQueryUnitV1) GetTimeZone() U.TimeZoneString {
	return q.Query.GetTimeZone()
}

func (q *AttributionQueryUnitV1) SetTimeZone(timezoneString U.TimeZoneString) {
	q.Query.Timezone = string(timezoneString)
}

func (query *AttributionQueryUnitV1) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	query.Query.ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone)
	return nil
}

func (q *AttributionQueryUnitV1) GetQueryCacheHashString() (string, error) {
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

func (q *AttributionQueryUnitV1) GetQueryCacheRedisKey(projectID int64) (*cacheRedis.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.Query.From, q.Query.To, U.TimeZoneString(q.Query.Timezone))
	return cacheRedis.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *AttributionQueryUnitV1) GetQueryCacheExpiry() float64 {
	return getQueryCacheResultExpiry(q.Query.From, q.Query.To, q.Query.Timezone)
}

func (q *AttributionQueryUnitV1) CheckIfNameIsPresent(nameOfQuery string) bool {
	return false
}

func (q *AttributionQueryUnitV1) SetDefaultGroupByTimestamp() {
	if q.Query.KPIQueries[0].KPI.Class == "" {
		return
	}
	for groupIndex := range q.Query.KPIQueries {
		for index := range q.Query.KPIQueries[groupIndex].KPI.Queries {
			if q.Query.KPIQueries[groupIndex].KPI.Queries[index].GroupByTimestamp == GroupByTimestampHour {
				q.Query.KPIQueries[groupIndex].KPI.Queries[index].GroupByTimestamp = GroupByTimestampSecond
			}
		}
	}
}

func (q *AttributionQueryUnitV1) GetGroupByTimestamps() []string {
	return []string{}
}

func (query *AttributionQueryUnitV1) TransformDateTypeFilters() error {
	return query.Query.TransformDateTypeFilters()
}

func (query *AttributionQueryV1) TransformDateTypeFilters() error {
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
func (query *AttributionQueryV1) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(query.Timezone)
}

// AddHeadersByAttributionKeyV1 Adds common column names and linked events as header to the result rows.
func AddHeadersByAttributionKeyV1(result *QueryResult, query *AttributionQueryV1, goalEvents []string, goalEventAggFuncTypes []string) {

	attributionKey := query.AttributionKey
	if attributionKey == AttributionKeyLandingPage || attributionKey == AttributionKeyAllPageView {
		// add up the attribution key
		result.Headers = append(result.Headers, attributionKey)

		// add up content groups
		for _, contentGroupName := range query.AttributionContentGroups {
			result.Headers = append(result.Headers, contentGroupName)
		}

		// add up fixed metrics

		result.Headers = append(result.Headers, AttributionFixedHeadersLandingPage...)

		for _, goal := range goalEvents {

			conversion := fmt.Sprintf("%s - Conversion", goal)
			conversionInfluence := fmt.Sprintf("%s - Conversion Value (InfluenceRemove) ", goal)
			result.Headers = append(result.Headers, conversion, conversionInfluence)

			conversionC := fmt.Sprintf("%s - Conversion(compare)", goal)
			conversionC_influence := fmt.Sprintf("%s - Conversion InfluenceRemove(compare)", goal)
			result.Headers = append(result.Headers, conversionC, conversionC_influence)
		}

	} else {

		// Add up for Added Keys {Campaign, Adgroup, Keyword}
		switch attributionKey {
		case AttributionKeyCampaign:
			result.Headers = append(result.Headers, AddedKeysForCampaign...)
		case AttributionKeyAdgroup:
			result.Headers = append(result.Headers, AddedKeysForAdgroup...)
		case AttributionKeyKeyword:
			result.Headers = append(result.Headers, AddedKeysForKeyword...)
		case AttributionKeySource:
			result.Headers = append(result.Headers, AddedKeysForSource...)
		case AttributionKeyChannel:
			result.Headers = append(result.Headers, AddedKeysForChannel...)
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
				conversionC_influence := fmt.Sprintf("%s - Conversion Value InfluenceRemove(compare)", goal)
				cpcC := fmt.Sprintf("%s - Return on Cost(compare)", goal)
				result.Headers = append(result.Headers, conversionC, conversionC_influence, cpcC)
			} else {
				conversion := fmt.Sprintf("%s - Conversion", goal)
				conversionInfluence := fmt.Sprintf("%s - Conversion Value (InfluenceRemove) ", goal)
				cpc := fmt.Sprintf("%s - Cost Per Conversion", goal)
				result.Headers = append(result.Headers, conversion, conversionInfluence, cpc)

				conversionC := fmt.Sprintf("%s - Conversion(compare)", goal)
				conversionC_influence := fmt.Sprintf("%s - Conversion InfluenceRemove(compare)", goal)
				cpcC := fmt.Sprintf("%s - Cost Per Conversion(compare)", goal)
				result.Headers = append(result.Headers, conversionC, conversionC_influence, cpcC)
			}
		}

		// add up key
		result.Headers = append(result.Headers, "key")
	}
}

//MergeTwoDataRowsV1 adds values of two data rows
func MergeTwoDataRowsV1(row1 []interface{}, row2 []interface{}, keyIndex int, attributionKey string, conversionFunTypes []string) []interface{} {

	if attributionKey == AttributionKeyLandingPage {

		row1[keyIndex+1] = row1[keyIndex+1].(float64) + row2[keyIndex+1].(float64) // Conversion.
		row1[keyIndex+2] = row1[keyIndex+2].(float64) + row2[keyIndex+2].(float64) // Conversion Influence - values same as Linear Touch
		row1[keyIndex+3] = row1[keyIndex+3].(float64) + row2[keyIndex+3].(float64) // Compare conversion
		row1[keyIndex+4] = row1[keyIndex+4].(float64) + row2[keyIndex+4].(float64) // Compare conversion-Influence - values same as Linear Touch

		// Remaining linked funnel events & CPCs
		for i := keyIndex + 5; i < len(row1)-1; i += 2 {
			row1[i] = row1[i].(float64) + row2[i].(float64)
			row1[i+1] = row1[i+1].(float64) + row2[i+1].(float64)
		}

		return row1
	} else {

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

	}
}

// EnrichRequestUsingAttributionConfigV1 adds some fields in attribution query using attribution config
func EnrichRequestUsingAttributionConfigV1(projectID int64, query *AttributionQueryV1, settings *ProjectSetting, logCtx *log.Entry) error {

	attributionConfig, err1 := decodeAttributionConfig(settings.AttributionConfig)
	if err1 != nil {
		return errors.New("failed to decode attribution config from project settings")
	}

	query.LookbackDays = attributionConfig.AttributionWindow
	if attributionConfig.QueryType == "" {
		query.QueryType = AttributionQueryTypeEngagementBased
	} else {
		query.QueryType = attributionConfig.QueryType
	}

	for index := range query.KPIQueries {
		switch query.KPIQueries[index].AnalyzeType {

		case AnalyzeTypeUsers:
			query.KPIQueries[index].RunType = RunTypeUser
			return nil
		case AnalyzeTypeUserKPI:
			query.KPIQueries[index].RunType = RunTypeUserKPI
			return nil
		case AnalyzeTypeHSDeals:
			if &attributionConfig != nil && attributionConfig.AnalyzeTypeHSCompaniesEnabled == true {
				query.KPIQueries[index].RunType = RunTypeHSCompanies
				return nil
			} else if &attributionConfig != nil && attributionConfig.AnalyzeTypeHSDealsEnabled == true {
				query.KPIQueries[index].RunType = RunTypeHSDeals
				return nil
			} else {
				logCtx.WithFields(log.Fields{"Query": query, "AttributionConfig": attributionConfig}).Error("Failed to set analyze type")
				return errors.New("invalid config/query. Failed to set analyze type from attribution config & project settings")
			}
		case AnalyzeTypeSFOpportunities:
			if &attributionConfig != nil && attributionConfig.AnalyzeTypeSFAccountsEnabled == true {
				query.KPIQueries[index].RunType = RunTypeSFAccounts
				return nil
			} else if &attributionConfig != nil && attributionConfig.AnalyzeTypeSFOpportunitiesEnabled == true {
				query.KPIQueries[index].RunType = RunTypeSFOpportunities
				return nil
			} else {
				logCtx.WithFields(log.Fields{"Query": query, "AttributionConfig": attributionConfig}).Error("Failed to set analyze type")
				return errors.New("invalid config/query. Failed to set analyze type from attribution config & project settings")
			}
		default:
			query.KPIQueries[index].AnalyzeType = AnalyzeTypeUsers
			query.KPIQueries[index].RunType = RunTypeUser
			return nil
			// logCtx.WithFields(log.Fields{"Query": query, "AttributionConfig": attributionConfig}).Error("Failed to set analyze type")
			// return errors.New("invalid config/query. Failed to set analyze type from attribution config & project settings")
		}
	}
	return nil
	// return errors.New("invalid config/query. Failed to set analyze type from attribution config & project settings")
}

// AddDefaultKeyDimensionsToAttributionQueryV1 adds default custom Dimensions for supporting existing old/saved queries
func AddDefaultKeyDimensionsToAttributionQueryV1(query *AttributionQueryV1) {

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

// AddDefaultMarketingEventTypeTacticOfferV1 adds default tactic or offer for older queries
func AddDefaultMarketingEventTypeTacticOfferV1(query *AttributionQueryV1) {

	// Default is set as Tactic
	if (*query).TacticOfferType == "" {
		(*query).TacticOfferType = MarketingEventTypeTactic
	}
}

// BuildEventNamesPlaceholderV1 Returns the concatenated list of conversion event + funnel events names
func BuildEventNamesPlaceholderV1(query *AttributionQueryV1) []string {
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

// ProcessOTPEventRowsV1 returns OTP session data for each userId, attributionKey pair
func ProcessOTPEventRowsV1(rows *sql.Rows, query *AttributionQueryV1,
	logCtx log.Entry, queryID string) (map[string]map[string]UserSessionData, []string, error) {

	attributedSessionsByUserId := make(map[string]map[string]UserSessionData)
	userIdMap := make(map[string]bool)
	var userIdsWithSession []string

	startReadTime := time.Now()
	for rows.Next() {
		var userIDNull sql.NullString
		var campaignIDNull sql.NullString
		var campaignNameNull sql.NullString
		var sourceNameNull sql.NullString
		var channelGroupNull sql.NullString
		var typeNull sql.NullString
		var attributionIdNull sql.NullString
		var timestampNull sql.NullInt64

		if err := rows.Scan(&userIDNull, &campaignIDNull, &campaignNameNull, &sourceNameNull, &channelGroupNull, &typeNull, &attributionIdNull, &timestampNull); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row (OTP). Continuing")
			continue
		}

		var userID string
		var campaignID string
		var campaignName string
		var sourceName string
		var channelGroup string
		var typeName string
		var attributionKeyName string
		var timestamp int64

		userID = U.IfThenElse(userIDNull.Valid, userIDNull.String, PropertyValueNone).(string)
		campaignID = U.IfThenElse(campaignIDNull.Valid, campaignIDNull.String, PropertyValueNone).(string)
		campaignName = U.IfThenElse(campaignNameNull.Valid, campaignNameNull.String, PropertyValueNone).(string)
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
		marketingValues := MarketingData{Channel: SessionChannelOTP, CampaignID: campaignID, CampaignName: campaignName, AdgroupID: PropertyValueNone, AdgroupName: PropertyValueNone, KeywordName: PropertyValueNone, KeywordMatchType: PropertyValueNone,
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

// AddTheAddedKeysAndMetricsV1 adds addedKeys and few metrics in attribution data
func AddTheAddedKeysAndMetricsV1(attributionData *map[string]*AttributionData, query *AttributionQueryV1, sessions map[string]map[string]UserSessionData, noOfConversionEvents int) {

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
						(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, sessionKeyMarketingInfo[key].CampaignName)
						(*attributionData)[key].Name = sessionKeyMarketingInfo[key].Source
					case AttributionKeyChannel:
						(*attributionData)[key].AddedKeys = append((*attributionData)[key].AddedKeys, sessionKeyMarketingInfo[key].CampaignName)
						(*attributionData)[key].Name = sessionKeyMarketingInfo[key].ChannelGroup
					case AttributionKeyLandingPage:
						(*attributionData)[key].Name = sessionKeyMarketingInfo[key].LandingPageUrl
					}

				}
			}
		}
	}
}

// ProcessQueryKPILandingPageUrlV1 converts attribution data into result
func ProcessQueryKPILandingPageUrlV1(query *AttributionQueryV1, attributionData *map[string]*AttributionData, logCtx log.Entry, kpiData map[string]KPIInfo, isCompare bool) *QueryResult {
	logFields := log.Fields{"Method": "ProcessQueryKPILandingPageUrl"}
	logCtx = *logCtx.WithFields(logFields)
	dataRows := GetRowsByMapsKPILandingPage(query.AttributionContentGroups, attributionData, isCompare)
	logCtx.Info("Done GetRowsByMapsKPILandingPage")
	result := &QueryResult{}
	var goalEvents []string
	for _, value := range kpiData {
		goalEvents = value.KpiHeaderNames
	}

	AddHeadersByAttributionKeyV1(result, query, goalEvents, nil)

	result.Rows = dataRows

	// Update result based on Key Dimensions
	err := GetUpdatedRowsByDimensionsV1(result, query, logCtx)
	if err != nil {
		return nil
	}
	result.Rows = MergeDataRowsHavingSameKeyV1(result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), query.AttributionKey, nil, logCtx)
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

	result.Rows = AddGrandTotalRowKPILandingPage(result.Headers, result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), goalEvents, query.AttributionMethodology, query.AttributionMethodologyCompare)
	logCtx.Info("Done AddGrandTotal")
	return result

}

// ProcessQueryKPIPageUrlV1 converts attribution data into result
func ProcessQueryKPIPageUrlV1(query *AttributionQueryV1, attributionData *map[string]*AttributionData, logCtx log.Entry, kpiData map[string]KPIInfo, isCompare bool) *QueryResult {
	logFields := log.Fields{"Method": "ProcessQueryKPILandingPageUrl"}
	logCtx = *logCtx.WithFields(logFields)
	dataRows := GetRowsByMapsKPIPage(query.AttributionKey, query.AttributionContentGroups, attributionData, isCompare)
	logCtx.Info("Done GetRowsByMapsKPIPage")
	result := &QueryResult{}
	var goalEvents []string
	for _, value := range kpiData {
		goalEvents = value.KpiHeaderNames
	}

	AddHeadersByAttributionKeyV1(result, query, goalEvents, nil)

	result.Rows = dataRows

	// Update result based on Key Dimensions
	err := GetUpdatedRowsByDimensionsV1(result, query, logCtx)
	if err != nil {
		return nil
	}
	result.Rows = MergeDataRowsHavingSameKeyV1(result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), query.AttributionKey, nil, logCtx)
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

	result.Rows = AddGrandTotalRowKPILandingPage(result.Headers, result.Rows, GetLastKeyValueIndexLandingPage(result.Headers), goalEvents, query.AttributionMethodology, query.AttributionMethodologyCompare)
	logCtx.Info("Done AddGrandTotal")
	return result

}

// MergeDataRowsHavingSameKeyV1 merges rows having same key by adding each column value
func MergeDataRowsHavingSameKeyV1(rows [][]interface{}, keyIndex int, attributionKey string, conversionFunTypes []string, logCtx log.Entry) [][]interface{} {

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
			rowKeyMap[key] = MergeTwoDataRowsV1(rowKeyMap[key], row, keyIndex, attributionKey, conversionFunTypes)
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

// GetUpdatedRowsByDimensionsV1 updated the granular result with reduced dimensions
func GetUpdatedRowsByDimensionsV1(result *QueryResult, query *AttributionQueryV1, logCtx log.Entry) error {

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

// ProcessQueryKPIV1 converts attribution data into result
func ProcessQueryKPIV1(query *AttributionQueryV1, attributionData *map[string]*AttributionData,
	marketingReports *MarketingReports, isCompare bool, kpiData map[string]KPIInfo) *QueryResult {

	logCtx := log.WithFields(log.Fields{"Method": "ProcessQueryKPI", "KPIAttribution": "Debug"})

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "attributionData": attributionData}).Info("KPI Attribution data")
	}

	// add CampaignData result based on Key Dimensions
	AddCampaignDataForSourceV1(*attributionData, marketingReports, query)

	// add CampaignData result based on Key Dimensions
	AddCampaignDataForChannelGroupV1(*attributionData, marketingReports, query)

	for key, _ := range *attributionData {
		//add key to attribution data
		addKeyToMarketingInfoForChannelOrSourceV1(attributionData, key, query)
	}

	// Add additional metrics values
	ComputeAdditionalMetrics(attributionData)

	// Add custom dimensions
	AddCustomDimensionsV1(attributionData, query, marketingReports)

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "attributionData": attributionData}).Info("Done AddTheAddedKeysAndMetrics AddPerformanceData ApplyFilter ComputeAdditionalMetrics AddCustomDimensions")
	}
	// for KPI queries, use the kpiData.KpiAggFunctionTypes as ConvAggFunctionType
	var convAggFunctionType []string
	for _, val := range kpiData {
		if len(val.KpiAggFunctionTypes) > 0 {
			convAggFunctionType = val.KpiAggFunctionTypes
			break
		}
	}
	for key, _ := range *attributionData {
		(*attributionData)[key].ConvAggFunctionType = convAggFunctionType
	}

	// Attribution data to rows
	dataRows := GetRowsByMaps(query.AttributionKey, query.AttributionKeyCustomDimension, attributionData, query.LinkedEvents, isCompare)
	result := &QueryResult{}

	// get the headers for KPI
	var goalEvents []string
	var goalEventAggFuncTypes []string
	for _, value := range kpiData {
		goalEvents = value.KpiHeaderNames
		goalEventAggFuncTypes = value.KpiAggFunctionTypes
		break
	}

	AddHeadersByAttributionKeyV1(result, query, goalEvents, goalEventAggFuncTypes)
	result.Rows = dataRows
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "result": result}).Info("Done GetRowsByMaps AddHeadersByAttributionKey")
	}
	// Update result based on Key Dimensions
	err := GetUpdatedRowsByDimensionsV1(result, query, *logCtx)
	if err != nil {
		return nil
	}

	result.Rows = MergeDataRowsHavingSameKeyKPIV1(result.Rows, GetLastKeyValueIndex(result.Headers), query.AttributionKey, goalEventAggFuncTypes, *logCtx)
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "Result": result}).Info("KPI Attribution result")
	}
	// Additional filtering based on AttributionKey.
	result.Rows = FilterRows(result.Rows, query.AttributionKey, GetLastKeyValueIndex(result.Headers))

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

	result.Rows = AddGrandTotalRowKPI(result.Headers, result.Rows, GetLastKeyValueIndex(result.Headers), goalEventAggFuncTypes, query.AttributionMethodology, query.AttributionMethodologyCompare)

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "Result": result}).Info("KPI Attribution result AddGrandTotalRow")
	}
	return result
}

// AddCustomDimensionsV1 adds custom dimensions in attribution data
func AddCustomDimensionsV1(attributionData *map[string]*AttributionData, query *AttributionQueryV1, reports *MarketingReports) {

	// Custom Dimensions are support only for Campaign and Adgroup currently
	if query.AttributionKey != AttributionKeyCampaign && query.AttributionKey != AttributionKeyAdgroup {
		return
	}

	// Return if extra Custom Dimensions not required
	if !isExtraDimensionRequiredV1(query) {
		return
	}

	if query.AttributionKey == AttributionKeyCampaign {
		enrichDimensionsWithoutChannel(attributionData, query.AttributionKeyCustomDimension, reports.AdwordsCampaignDimensions, reports.FacebookCampaignDimensions, reports.LinkedinCampaignDimensions, reports.BingadsCampaignDimensions, reports.CustomAdsCampaignDimensions, query.AttributionKey)
	} else if query.AttributionKey == AttributionKeyAdgroup {
		enrichDimensionsWithoutChannel(attributionData, query.AttributionKeyCustomDimension, reports.AdwordsAdgroupDimensions, reports.FacebookAdgroupDimensions, reports.LinkedinAdgroupDimensions, reports.BingadsAdgroupDimensions, reports.CustomAdsAdgroupDimensions, query.AttributionKey)
	}
}

//isExtraDimensionRequiredV1 checks if extra dimensions are required
func isExtraDimensionRequiredV1(query *AttributionQueryV1) bool {
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

func ApplyFilterV1(attributionData *map[string]*AttributionData, query *AttributionQueryV1) {
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

// ProcessEventRowsV1 returns session data for each userId, attributionKey pair
func ProcessEventRowsV1(rows *sql.Rows, query *AttributionQueryV1, reports *MarketingReports,
	contentGroupNamesList []string, attributedSessionsByUserId *map[string]map[string]UserSessionData,
	userIdsWithSession *[]string, logCtx log.Entry, queryID string) error {

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	userIdMap := make(map[string]bool)
	reports.CampaignSourceMapping = make(map[string]string)
	reports.CampaignChannelGroupMapping = make(map[string]string)

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
		} else if query.AttributionKey == AttributionKeySource && U.IsNonEmptyKey(sourceName) {

			if _, ok := reports.CampaignSourceMapping[marketingValues.CampaignName]; !ok {
				reports.CampaignSourceMapping[marketingValues.CampaignName] = sourceName
			}
		} else if query.AttributionKey == AttributionKeyChannel && U.IsNonEmptyKey(channelGroup) {

			if _, ok := reports.CampaignChannelGroupMapping[marketingValues.CampaignName]; !ok {
				reports.CampaignChannelGroupMapping[marketingValues.CampaignName] = channelGroup
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
