package model

import (
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Template struct {
	ProjectID  int64           `json:"project_id"`
	Type       int             `json:"type"`
	Thresholds *postgres.Jsonb `json:"thresholds"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type TemplateQuery struct {
	Metric     string            `json:"metric"`
	Type       int               `json:"type"`
	From       int64             `json:"from"`
	To         int64             `json:"to"`
	PrevFrom   int64             `json:"prev_from"`
	PrevTo     int64             `json:"prev_to"`
	Thresholds RequestThresholds `json:"thresholds"`
	Class      string            `json:"cl"`
	Timezone   string            `json:"time_zone"`
}

type RequestThresholds struct {
	PercentageChange float64 `json:"percentage_change"`
	AbsoluteChange   float64 `json:"absolute_change"`
}

func (q *TemplateQuery) GetClass() string {
	return q.Class
}

func (q *TemplateQuery) GetQueryDateRange() (from, to int64) {
	return q.From, q.To
}

func (q *TemplateQuery) SetQueryDateRange(from, to int64) {
	q.From, q.To = from, to
}

func (q *TemplateQuery) GetQueryCacheHashString() (string, error) {
	queryMap, err := U.EncodeStructTypeToMap(q)
	if err != nil {
		return "", err
	}
	delete(queryMap, "from")
	delete(queryMap, "to")
	delete(queryMap, "prev_from")
	delete(queryMap, "prev_to")
	queryHash, err := U.GenerateHashStringForStruct(queryMap)
	if err != nil {
		return "", err
	}
	return queryHash, nil
}

func (q *TemplateQuery) GetQueryCacheRedisKey(projectID int64) (*cacheRedis.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := fmt.Sprintf("%s:from:%d:to:%d:prev_from:%d:prev_to:%d", hashString, q.From, q.To, q.PrevFrom, q.PrevTo)
	return cacheRedis.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *TemplateQuery) GetQueryCacheExpiry(projectID int64) float64 {
	return getQueryCacheResultExpiry(projectID, q.From, q.To, string(q.Timezone))
}

func (q *TemplateQuery) TransformDateTypeFilters() error {
	return nil
}

func (q *TemplateQuery) SetTimeZone(timezoneString U.TimeZoneString) {
	q.Timezone = string(timezoneString)
}

func (q *TemplateQuery) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(q.Timezone)
}

func (q *TemplateQuery) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	return nil
}

func (query *TemplateQuery) CheckIfNameIsPresent(nameOfQuery string) bool {
	return false
}

func (query *TemplateQuery) SetDefaultGroupByTimestamp() {
}

func (query *TemplateQuery) GetGroupByTimestamps() []string {
	return []string{}
}

var DefaultThresholds = RequestThresholds{
	PercentageChange: 10,
	AbsoluteChange:   0,
}

type TemplateResponseMeta struct {
	PrimaryLevel LevelMeta `json:"primary_level"`
	SubLevel     LevelMeta `json:"sub_level"`
}
type LevelMeta struct {
	ColumnName string `json:"column_name"`
}

type PrimaryLevelData struct {
	Name             string         `json:"name"`
	PercentageChange float64        `json:"percentage_change"`
	AbsoluteChange   float64        `json:"absolute_change"`
	PreviousValue    float64        `json:"previous_value"`
	LastValue        float64        `json:"last_value"`
	IsInfinity       bool           `json:"is_infinity"`
	SubLevelData     []SubLevelData `json:"sub_level_data"`
}
type SubLevelData struct {
	Name             string            `json:"name"`
	PercentageChange float64           `json:"percentage_change"`
	AbsoluteChange   float64           `json:"absolute_change"`
	PreviousValue    float64           `json:"previous_value"`
	LastValue        float64           `json:"last_value"`
	IsInfinity       bool              `json:"is_infinity"`
	RootCauseMetrics []RootCauseMetric `json:"root_cause_metrics"`
}

type RootCauseMetric struct {
	Metric           string  `json:"metric"`
	PercentageChange float64 `json:"percentage_change"`
	IsInfinity       bool    `json:"is_infinity"`
}
type OverallChanges struct {
	Metric           string  `json:"metric"`
	PercentageChange float64 `json:"percentage_change"`
	PreviousValue    float64 `json:"previous_value"`
	LastValue        float64 `json:"last_value"`
	IsInfinity       bool    `json:"is_infinity"`
}

type TemplateResponse struct {
	BreakdownAnalysis TemplateData         `json:"breakdown_analysis"`
	Meta              TemplateResponseMeta `json:"meta"`
}
type TemplateData struct {
	OverallChangesData []OverallChanges   `json:"overall_changes_data"`
	PrimaryLevelData   []PrimaryLevelData `json:"primary_level_data"`
}
type TemplateThreshold struct {
	Metric           string  `json:"metric"`
	PercentageChange float64 `json:"percentage_change"`
	AbsoluteChange   float64 `json:"absolute_change"`
}

type TemplateConfig struct {
	Metrics    []TemplateMetricWithDisplayName `json:"metrics"`
	Thresholds []TemplateThreshold             `json:"thresholds"`
}
type TemplateMetricWithDisplayName struct {
	Metric      string `json:"metric"`
	DisplayName string `json:"display_name"`
}

const (
	SEMChecklist = "sem_checklist"
)

var TemplateAliasToType = map[string]int{
	SEMChecklist: 1,
}
var TemplateAdwordsMetricsMapForAdwords = map[string]bool{
	Clicks:                true,
	Impressions:           true,
	ClickThroughRate:      true,
	CostPerClick:          true,
	SearchImpressionShare: true,
	"cost":                true,
	Conversion:            true,
	"cost_per_lead":       true,
	ConversionRate:        true,
}

var TemplateMetricsForAdwordsWithDisplayName = []TemplateMetricWithDisplayName{
	{
		Metric:      SearchImpressionShare,
		DisplayName: "SI Share",
	},
	{
		Metric:      Impressions,
		DisplayName: "Impr.",
	},
	{
		Metric:      Clicks,
		DisplayName: "Clicks",
	},
	{
		Metric:      ClickThroughRate,
		DisplayName: "CTR",
	},
	{
		Metric:      "cost",
		DisplayName: "Cost",
	},
	{
		Metric:      CostPerClick,
		DisplayName: "Avg. CPC",
	},
	{
		Metric:      Conversion,
		DisplayName: "Conv.",
	},
	{
		Metric:      "cost_per_lead",
		DisplayName: "Cost/Conv",
	},
	{
		Metric:      ConversionRate,
		DisplayName: "Conv. Rate",
	},
}
var TemplateMetricsForAdwordsMap = map[string]bool{
	Clicks:                true,
	Impressions:           true,
	ClickThroughRate:      true,
	CostPerClick:          true,
	SearchImpressionShare: true,
	"cost":                true,
	Conversion:            true,
	"cost_per_lead":       true,
	ConversionRate:        true,
}

//validates if the thresholds metric is part of allowed metrics and is not repeated. e.g: [{metric: clicks}, {metric: clicks}] not allowed
func ValidateTemplateThresholds(thresholds []TemplateThreshold) bool {
	metricsCountMap := make(map[string]int)
	for _, threshold := range thresholds {
		_, isExistsMetric := TemplateAdwordsMetricsMapForAdwords[threshold.Metric]
		if !isExistsMetric {
			return false
		}
		metricsCountMap[threshold.Metric]++
	}
	for _, count := range metricsCountMap {
		if count > 1 {
			return false
		}
	}
	return true
}

func GetInputOrDefaultTimestampsForTemplateQueryWithDays(query TemplateQuery, timezoneString U.TimeZoneString, days int) (int64, int64, int64, int64, error) {
	location, err := time.LoadLocation(string(timezoneString))
	if err != nil {
		return 0, 0, 0, 0, err
	}
	lastWeekFromTime := time.Unix(query.From, 0).In(location)
	lastWeekToTime := time.Unix(query.To, 0).In(location)
	var prevWeekFromTime, prevWeekToTime time.Time
	if query.PrevFrom == 0 {
		prevWeekFromTime = lastWeekFromTime.AddDate(0, 0, -days)
	} else {
		prevWeekFromTime = time.Unix(query.PrevFrom, 0).In(location)
	}
	if query.PrevTo == 0 {
		prevWeekToTime = lastWeekToTime.AddDate(0, 0, -days)
	} else {
		prevWeekToTime = time.Unix(query.PrevTo, 0).In(location)
	}
	lastWeekFromTimestamp, _ := strconv.ParseInt(lastWeekFromTime.Format("20060102"), 10, 64)
	lastWeekToTimestamp, _ := strconv.ParseInt(lastWeekToTime.Format("20060102"), 10, 64)
	prevWeekFromTimestamp, _ := strconv.ParseInt(prevWeekFromTime.Format("20060102"), 10, 64)
	prevWeekToTimestamp, _ := strconv.ParseInt(prevWeekToTime.Format("20060102"), 10, 64)
	return lastWeekFromTimestamp, lastWeekToTimestamp, prevWeekFromTimestamp, prevWeekToTimestamp, nil
}

func calcTotalAdImpressions(siShare float64, impressions float64) float64 {
	if siShare == 0 {
		return 0
	} else {
		return (impressions * 100 / siShare)
	}
}
func buildRootCauseMetricDirectRelation(metric string, percentageChangeAnalysisMetric float64, percentageChangeRootCause float64, previousValue float64, lastValue float64) (RootCauseMetric, bool) {
	rootCauseMetric := RootCauseMetric{Metric: metric, PercentageChange: percentageChangeRootCause}
	if previousValue < 0.1 {
		if lastValue < 0.1 {
			rootCauseMetric.PercentageChange = math.Round(rootCauseMetric.PercentageChange)
		}
		if lastValue > 0 {
			rootCauseMetric.IsInfinity = true
		}
	}
	if math.Round(rootCauseMetric.PercentageChange) != 0 {
		if percentageChangeAnalysisMetric > 0 && percentageChangeRootCause > 0 {
			return rootCauseMetric, true
		}
		if percentageChangeAnalysisMetric < 0 && percentageChangeRootCause < 0 {
			return rootCauseMetric, true
		}
	}
	return rootCauseMetric, false
}
func buildRootCauseMetricInverseRelation(metric string, percentageChangeAnalysisMetric float64, percentageChangeRootCause float64, previousValue float64, lastValue float64) (RootCauseMetric, bool) {
	rootCauseMetric := RootCauseMetric{Metric: metric, PercentageChange: percentageChangeRootCause}
	if previousValue < 0.1 {
		if lastValue < 0.1 {
			rootCauseMetric.PercentageChange = math.Round(rootCauseMetric.PercentageChange)
		}
		if lastValue > 0 {
			rootCauseMetric.IsInfinity = true
		}
	}
	if math.Round(rootCauseMetric.PercentageChange) != 0 {
		if percentageChangeAnalysisMetric < 0 && percentageChangeRootCause > 0 {
			return rootCauseMetric, true
		}
		if percentageChangeAnalysisMetric > 0 && percentageChangeRootCause < 0 {
			return rootCauseMetric, true
		}
	}
	return rootCauseMetric, false
}
func BuildRootCauseFromWeeklyDifferenceForLeads(query TemplateQuery, keywordAnalysis KeywordAnalysis) []RootCauseMetric {
	rootCauseMetrics := make([]RootCauseMetric, 0)
	percentageChangeTotalAdImpressions := 0.0
	prevTotalAdImpressions := calcTotalAdImpressions(keywordAnalysis.PrevSearchImpressionShare, keywordAnalysis.PrevImpressions)
	lastTotalAdImpressions := calcTotalAdImpressions(keywordAnalysis.LastSearchImpressionShare, keywordAnalysis.LastImpressions)
	if prevTotalAdImpressions == 0 {
		percentageChangeTotalAdImpressions = (lastTotalAdImpressions - prevTotalAdImpressions) * 100 / 0.0000001
	} else {
		percentageChangeTotalAdImpressions = (lastTotalAdImpressions - prevTotalAdImpressions) * 100 / prevTotalAdImpressions
	}
	// sanitising null values due to full outer join
	if keywordAnalysis.ClickThroughRate == 0 {
		keywordAnalysis.ClickThroughRate = calcPercentagesForTemplates(keywordAnalysis.LastClickThroughRate, keywordAnalysis.PrevClickThroughRate)
	}
	if keywordAnalysis.SearchImpressionShare == 0 {
		keywordAnalysis.SearchImpressionShare = calcPercentagesForTemplates(keywordAnalysis.LastSearchImpressionShare, keywordAnalysis.PrevSearchImpressionShare)
	}
	if keywordAnalysis.ConversionRate == 0 {
		keywordAnalysis.ConversionRate = calcPercentagesForTemplates(keywordAnalysis.LastConversionRate, keywordAnalysis.PrevConversionRate)
	}
	//
	rootCauseMetric, shouldAppend := buildRootCauseMetricDirectRelation(ClickThroughRate, keywordAnalysis.PercentageChange, keywordAnalysis.ClickThroughRate, keywordAnalysis.PrevClickThroughRate, keywordAnalysis.LastClickThroughRate)
	if shouldAppend {
		rootCauseMetrics = append(rootCauseMetrics, rootCauseMetric)
	}

	rootCauseMetric, shouldAppend = buildRootCauseMetricDirectRelation("Total Ad Impr.", keywordAnalysis.PercentageChange, percentageChangeTotalAdImpressions, prevTotalAdImpressions, lastTotalAdImpressions)
	if shouldAppend {
		rootCauseMetrics = append(rootCauseMetrics, rootCauseMetric)
	}
	rootCauseMetric, shouldAppend = buildRootCauseMetricDirectRelation(SearchImpressionShare, keywordAnalysis.PercentageChange, keywordAnalysis.SearchImpressionShare, keywordAnalysis.PrevSearchImpressionShare, keywordAnalysis.LastSearchImpressionShare)
	if shouldAppend {
		rootCauseMetrics = append(rootCauseMetrics, rootCauseMetric)
	}
	if query.Metric == Conversion {
		rootCauseMetric, shouldAppend = buildRootCauseMetricDirectRelation(ConversionRate, keywordAnalysis.PercentageChange, keywordAnalysis.ConversionRate, keywordAnalysis.PrevConversionRate, keywordAnalysis.LastConversionRate)
		if shouldAppend {
			rootCauseMetrics = append(rootCauseMetrics, rootCauseMetric)
		}
	} else {
		rootCauseMetric, shouldAppend = buildRootCauseMetricInverseRelation(ConversionRate, keywordAnalysis.PercentageChange, keywordAnalysis.ConversionRate, keywordAnalysis.PrevConversionRate, keywordAnalysis.LastConversionRate)
		if shouldAppend {
			rootCauseMetrics = append(rootCauseMetrics, rootCauseMetric)
		}
	}
	return rootCauseMetrics
}
func BuildRootCauseFromWeeklyDifferenceForCostPerLead(query TemplateQuery, keywordAnalysis KeywordAnalysis) []RootCauseMetric {
	rootCauseMetrics := BuildRootCauseFromWeeklyDifferenceForLeads(query, keywordAnalysis)
	if keywordAnalysis.CostPerClick == 0 {
		keywordAnalysis.CostPerClick = calcPercentagesForTemplates(keywordAnalysis.LastCostPerClick, keywordAnalysis.PrevCostPerClick)
	}
	rootCauseMetric, shouldAppend := buildRootCauseMetricDirectRelation(CostPerClick, keywordAnalysis.PercentageChange, keywordAnalysis.CostPerClick, keywordAnalysis.PrevCostPerClick, keywordAnalysis.LastCostPerClick)
	if shouldAppend {
		rootCauseMetrics = append(rootCauseMetrics, rootCauseMetric)
	}
	return rootCauseMetrics
}

func TransformKeywordAnalysisToTemplateSubLevelData(query TemplateQuery, keywordAnalysis KeywordAnalysis) SubLevelData {
	//rounding of values < 0.1
	if keywordAnalysis.PreviousWeekValue < 0.1 {
		keywordAnalysis.PreviousWeekValue = math.Round(keywordAnalysis.PreviousWeekValue)
	}
	if keywordAnalysis.LastWeekValue < 0.1 {
		keywordAnalysis.LastWeekValue = math.Round(keywordAnalysis.LastWeekValue)
	}
	var subLevelData SubLevelData
	switch keywordAnalysis.KeywordMatchType {
	case "Exact":
		subLevelData.Name = "[" + keywordAnalysis.KeywordName + "]"
	case "Phrase":
		subLevelData.Name = `"` + keywordAnalysis.KeywordName + `"`
	default:
		subLevelData.Name = keywordAnalysis.KeywordName
	}
	subLevelData.PercentageChange = keywordAnalysis.PercentageChange
	subLevelData.AbsoluteChange = keywordAnalysis.AbsoluteChange
	subLevelData.PreviousValue = keywordAnalysis.PreviousWeekValue
	subLevelData.LastValue = keywordAnalysis.LastWeekValue
	if keywordAnalysis.PreviousWeekValue == 0 && keywordAnalysis.LastWeekValue != 0 {
		subLevelData.IsInfinity = true
	}
	if query.Metric == Conversion {
		rootCauseMetrics := BuildRootCauseFromWeeklyDifferenceForLeads(query, keywordAnalysis)
		subLevelData.RootCauseMetrics = rootCauseMetrics
	}
	if query.Metric == "cost_per_lead" {
		rootCauseMetrics := BuildRootCauseFromWeeklyDifferenceForCostPerLead(query, keywordAnalysis)
		subLevelData.RootCauseMetrics = rootCauseMetrics
	}

	return subLevelData
}
func TransfromCampaignLevelDataToTemplatePrimaryLevelData(campaignAnalysisRow CampaignAnalysis, campaignIDToSubLevelDataMap map[string][]SubLevelData) PrimaryLevelData {
	//rounding of values < 0.1
	if campaignAnalysisRow.PreviousWeekValue < 0.1 {
		campaignAnalysisRow.PreviousWeekValue = math.Round(campaignAnalysisRow.PreviousWeekValue)
	}
	if campaignAnalysisRow.LastWeekValue < 0.1 {
		campaignAnalysisRow.LastWeekValue = math.Round(campaignAnalysisRow.LastWeekValue)
	}
	var primaryLevelData PrimaryLevelData
	primaryLevelData.Name = campaignAnalysisRow.CampaignName
	primaryLevelData.PreviousValue = campaignAnalysisRow.PreviousWeekValue
	primaryLevelData.LastValue = campaignAnalysisRow.LastWeekValue
	primaryLevelData.PercentageChange = campaignAnalysisRow.PercentageChange
	primaryLevelData.AbsoluteChange = campaignAnalysisRow.AbsoluteChange
	primaryLevelData.SubLevelData = campaignIDToSubLevelDataMap[campaignAnalysisRow.CampaignID]
	if campaignAnalysisRow.PreviousWeekValue == 0 && campaignAnalysisRow.LastWeekValue != 0 {
		primaryLevelData.IsInfinity = true
	}

	return primaryLevelData
}
func calcPercentagesForTemplates(lastWeekValue float64, previousWeekValue float64) float64 {
	if previousWeekValue < 0.1 {
		return (lastWeekValue - previousWeekValue) * 100 / 0.0000001
	}
	return (lastWeekValue - previousWeekValue) * 100 / previousWeekValue
}
func SanitiseKeywordsAnalysisResult(query TemplateQuery, keywordAnalysisResult []KeywordAnalysis) []KeywordAnalysis {
	sanitisedKeywordsAnalysisResult := make([]KeywordAnalysis, 0)
	for _, keywordAnalysisRow := range keywordAnalysisResult {
		absoluteChange, percentageChange, shouldAppend := sanitiseNullValues(query, keywordAnalysisRow.AbsoluteChange, keywordAnalysisRow.PercentageChange, keywordAnalysisRow.LastWeekValue, keywordAnalysisRow.PreviousWeekValue)
		if shouldAppend {
			keywordAnalysisRow.AbsoluteChange = absoluteChange
			keywordAnalysisRow.PercentageChange = percentageChange
			sanitisedKeywordsAnalysisResult = append(sanitisedKeywordsAnalysisResult, keywordAnalysisRow)
		}
	}
	return sanitisedKeywordsAnalysisResult
}
func SanitiseCampaignAnalysisResult(query TemplateQuery, campaignAnalysisResult []CampaignAnalysis) []CampaignAnalysis {
	sanitisedCampaignsAnalysisResult := make([]CampaignAnalysis, 0)
	for _, campaignAnalysisRow := range campaignAnalysisResult {
		absoluteChange, percentageChange, shouldAppend := sanitiseNullValues(query, campaignAnalysisRow.AbsoluteChange, campaignAnalysisRow.PercentageChange, campaignAnalysisRow.LastWeekValue, campaignAnalysisRow.PreviousWeekValue)
		if shouldAppend {
			campaignAnalysisRow.AbsoluteChange = absoluteChange
			campaignAnalysisRow.PercentageChange = percentageChange
			sanitisedCampaignsAnalysisResult = append(sanitisedCampaignsAnalysisResult, campaignAnalysisRow)
		}
	}
	return sanitisedCampaignsAnalysisResult
}
func sanitiseNullValues(query TemplateQuery, absoluteChange float64, percentageChange float64, lastWeekValue float64, previousWeekValue float64) (float64, float64, bool) {
	absoluteChange = math.Abs(absoluteChange)
	if absoluteChange == 0 {
		absoluteChange = math.Abs(lastWeekValue - previousWeekValue)
		percentageChange = calcPercentagesForTemplates(lastWeekValue, previousWeekValue)
		if lastWeekValue < 0.1 && previousWeekValue < 0.1 {
			percentageChange = math.Round(percentageChange)
		}
	}
	if math.Abs(percentageChange) >= query.Thresholds.PercentageChange && absoluteChange > query.Thresholds.AbsoluteChange {
		return absoluteChange, percentageChange, true
	}
	return 0, 0, false
}

type KeywordAnalysis struct {
	KeywordID                 int64   `json:"keyword_id"`
	KeywordName               string  `json:"keyword_name"`
	PreviousWeekValue         float64 `json:"previous_week_value"`
	LastWeekValue             float64 `json:"last_week_value"`
	PercentageChange          float64 `json:"percentage_change"`
	AbsoluteChange            float64 `json:"absolute_change"`
	CampaignID                string  `json:"campaign_id"`
	KeywordMatchType          string  `json:"keyword_match_type"`
	Impressions               float64 `json:"impressions"`
	SearchImpressionShare     float64 `json:"search_impression_share"`
	ConversionRate            float64 `json:"conversion_rate"`
	ClickThroughRate          float64 `json:"click_through_rate"`
	CostPerClick              float64 `json:"cost_per_click"`
	PrevImpressions           float64 `json:"prev_impressions"`
	PrevSearchImpressionShare float64 `json:"prev_search_impression_share"`
	PrevConversionRate        float64 `json:"prev_conversion_rate"`
	PrevCostPerClick          float64 `json:"prev_cost_per_click"`
	PrevClickThroughRate      float64 `json:"prev_click_through_rate"`
	LastImpressions           float64 `json:"last_impressions"`
	LastSearchImpressionShare float64 `json:"last_search_impression_share"`
	LastConversionRate        float64 `json:"last_conversion_rate"`
	LastCostPerClick          float64 `json:"last_cost_per_click"`
	LastClickThroughRate      float64 `json:"last_click_through_rate"`
}

type CampaignAnalysis struct {
	CampaignName      string  `json:"campaign_name"`
	PreviousWeekValue float64 `json:"previous_week_value"`
	LastWeekValue     float64 `json:"last_week_value"`
	PercentageChange  float64 `json:"percentage_change"`
	AbsoluteChange    float64 `json:"absolute_change"`
	CampaignID        string  `json:"campaign_id"`
}
