package model

import (
	U "factors/util"
	"strconv"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Template struct {
	ProjectID  uint64          `json:"project_id"`
	Type       int             `json:"type"`
	Thresholds *postgres.Jsonb `json:"thresholds"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type TemplateQuery struct {
	Metric   string `json:"metric"`
	Type     int    `json:"type"`
	From     int64  `json:"from"`
	To       int64  `json:"to"`
	Timezone string `json:"timezone"`
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

var TemplateAliasToType = map[string]int{
	"sem_checklist": 1,
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

func GetTimestampsForTemplateQueryWithDays(query TemplateQuery, days int) (int64, int64, int64, int64) {
	var timeZoneString U.TimeZoneString
	if len(query.Timezone) < 1 {
		timeZoneString = U.TimeZoneStringIST
	}
	location, _ := time.LoadLocation(string(timeZoneString))
	lastWeekFromTime := time.Unix(query.From, 0).In(location)
	lastWeekToTime := time.Unix(query.To, 0).In(location)
	prevWeekFromTime := lastWeekFromTime.AddDate(0, 0, -days)
	prevWeekToTime := lastWeekToTime.AddDate(0, 0, -days)
	lastWeekFromTimestamp, _ := strconv.ParseInt(lastWeekFromTime.Format("20060102"), 10, 64)
	lastWeekToTimestamp, _ := strconv.ParseInt(lastWeekToTime.Format("20060102"), 10, 64)
	prevWeekFromTimestamp, _ := strconv.ParseInt(prevWeekFromTime.Format("20060102"), 10, 64)
	prevWeekToTimestamp, _ := strconv.ParseInt(prevWeekToTime.Format("20060102"), 10, 64)
	return lastWeekFromTimestamp, lastWeekToTimestamp, prevWeekFromTimestamp, prevWeekToTimestamp
}
