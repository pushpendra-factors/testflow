package model

import (
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"strings"
	"time"
)

const ProfileQueryClass = "profiles"

// From and to refer to JoinTime
type ProfileQueryGroup struct {
	Class          string                 `json:"cl"`
	Queries        []ProfileQuery         `json:"queries"`
	GlobalFilters  []QueryProperty        `json:"gup"`
	GlobalGroupBys []QueryGroupByProperty `json:"gbp"`
	From           int64                  `json:"from"`
	To             int64                  `json:"to"`
	Timezone       string                 `json:"tz"`
	DateField      string                 `json:"daFie"`
	GroupAnalysis  string                 `json:"grpa"`
}

func (q *ProfileQueryGroup) GetClass() string {
	return q.Class
}

func (q *ProfileQueryGroup) GetQueryDateRange() (from, to int64) {
	return 0, 0
}

func (q *ProfileQueryGroup) SetQueryDateRange(from, to int64) {
	q.From, q.To = 0, 0
}

func (q *ProfileQueryGroup) GetQueryCacheHashString() (string, error) {
	queryMap, err := U.EncodeStructTypeToMap(q)
	if err != nil {
		return "", err
	}
	delete(queryMap, "from")
	delete(queryMap, "to")
	queryHash, err := U.GenerateHashStringForStruct(queryMap)
	if err != nil {
		return "", err
	}
	return queryHash, nil
}

func (q *ProfileQueryGroup) GetQueryCacheRedisKey(projectID uint64) (*cacheRedis.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := fmt.Sprintf("%s:from:%d:to:%d", hashString, q.From, q.To)
	return cacheRedis.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *ProfileQueryGroup) GetQueryCacheExpiry() float64 {
	return 86400
}

func (q *ProfileQueryGroup) TransformDateTypeFilters() error {
	for index := range q.Queries {
		err := q.Queries[index].TransformDateTypeFilters()
		if err != nil {
			return err
		}
	}
	return nil
}
func (q *ProfileQueryGroup) SetTimeZone(timezoneString U.TimeZoneString) {
	q.Timezone = string(timezoneString)
}
func (q *ProfileQueryGroup) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(q.Timezone)
}

func (q *ProfileQueryGroup) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	for i := range q.GlobalFilters {
		q.GlobalFilters[i].ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
	}
	for i := range q.Queries {
		for j := range q.Queries[i].Filters {
			q.Queries[i].Filters[j].ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
		}
	}
	return nil
}

type ProfileQuery struct {
	Type          string                 `json:"ty"` // all_users, hubspot_events, etc
	Filters       []QueryProperty        `json:"pr"`
	GroupBys      []QueryGroupByProperty `json:"group_bys"`
	From          int64                  `json:"from"`
	To            int64                  `json:"to"`
	Timezone      string                 `json:"tz"`
	GroupAnalysis string                 `json:"grpa"`
	GroupId       int                    `json:"grpid"`
	AliasName     string                 `json:"an"`

	// For specific case of KPI - single eventType
	AggregateFunction     string `json:"agFn"`
	AggregateProperty     string `json:"agPr"`
	AggregatePropertyType string `json:"agPrTy"`
	DateField             string `json:"daFie"` // Currently used for replacement of jointimestamp in filters.
}

func (q *ProfileQuery) SetTimeZone(timezoneString U.TimeZoneString) {
	q.Timezone = string(timezoneString)
}

func (q *ProfileQuery) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(q.Timezone)
}
func (query *ProfileQuery) TransformDateTypeFilters() error {
	timezoneString := query.GetTimeZone()
	for i, _ := range query.Filters {
		err := query.Filters[i].TransformDateTypeFilters(timezoneString)
		if err != nil {
			return err
		}
	}
	return nil
}

const (
	DefaultSelectForAllUsers               = "COUNT(DISTINCT(COALESCE(users.customer_user_id, users.id))) as " + AliasAggr
	DefaultSelectForProfilesWithProperties = "%s(CASE WHEN JSON_EXTRACT_STRING(properties, '%s') IS NOT NULL THEN (JSON_EXTRACT_STRING(properties, '%s')) ELSE 0) as " + AliasAggr
)

func SanitizeDateTypeRowsProfiles(result *QueryResult, query *ProfileQuery) {
	headerIndexMap := make(map[string][]int)
	for index, header := range result.Headers {
		// If same group by is added twice, it will appear twice in headers.
		// Keep as a list to sanitize both indexes.
		headerIndexMap[header] = append(headerIndexMap[header], index)
	}

	alreadySanitizedProperties := make(map[string]bool)
	for _, gbp := range query.GroupBys {
		if gbp.Type == U.PropertyTypeDateTime {
			if _, sanitizedAlready := alreadySanitizedProperties[gbp.Property]; sanitizedAlready {
				continue
			}
			indexesToSanitize := headerIndexMap[gbp.Property]
			for _, indexToSanitize := range indexesToSanitize {
				sanitizeDateTypeForSpecificIndexProfiles(query, result.Rows, indexToSanitize)
			}
			alreadySanitizedProperties[gbp.Property] = true
		}
	}
}

func sanitizeDateTypeForSpecificIndexProfiles(query *ProfileQuery, rows [][]interface{}, indexToSanitize int) {
	for _, row := range rows {
		if row[indexToSanitize].(string) == "" || row[indexToSanitize].(string) == PropertyValueNone {
			continue
		}
		currentValueInTimeFormat, _ := time.Parse(U.DATETIME_FORMAT_DB, row[indexToSanitize].(string))
		row[indexToSanitize] = U.GetTimestampAsStrWithTimezone(currentValueInTimeFormat, query.Timezone)
	}
}

func AddQueryIndexToResult(result *QueryResult, queryIndex int) {
	result.Headers = append([]string{"query_index"}, result.Headers...)
	for index, row := range result.Rows {
		result.Rows[index] = append([]interface{}{queryIndex}, row...)
	}
}

func TransformProfilesQuery(query ProfileQuery) ProfileQuery {
	transformedFilters := make([]QueryProperty, 0)
	for _, filter := range query.Filters {
		filter.Entity = PropertyEntityUserGlobal
		transformedFilters = append(transformedFilters, filter)
	}
	query.Filters = transformedFilters
	return query
}

func IsValidProfileQueryGroupName(source string) bool {
	_, exists := AllowedGroupNames[source]
	if exists || source == USERS {
		return true
	} else {
		return false
	}
}

func GetSourceFromQueryTypeOrGroupName(query ProfileQuery) int {
	source, exists := UserSourceMap[query.Type]
	if exists {
		return source
	}
	if strings.HasPrefix(query.GroupAnalysis, U.HUBSPOT_PROPERTY_PREFIX) {
		return UserSourceHubspot
	}
	if strings.HasPrefix(query.GroupAnalysis, U.SALESFORCE_PROPERTY_PREFIX) {
		return UserSourceSalesforce
	}
	return 0
}
