package model

import (
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
	Metric string `json:"metric"`
	Type   int    `json:"type"`
	From   int64  `json:"from"`
	To     int64  `json:"to"`
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
	SubLevelData     []SubLevelData `json:"sub_level_data"`
}
type SubLevelData struct {
	Name             string            `json:"name"`
	PercentageChange float64           `json:"percentage_change"`
	AbsoluteChange   float64           `json:"absolute_change"`
	PreviousValue    float64           `json:"previous_value"`
	LastValue        float64           `json:"last_value"`
	RootCauseMetrics []RootCauseMetric `json:"root_cause_metrics"`
}

type RootCauseMetric struct {
	Metric           string  `json:"metric"`
	PercentageChange float64 `json:"percentage_change"`
}

type TemplateResponse struct {
	BreakdownAnalysis TemplateData         `json:"breakdown_analysis"`
	Meta              TemplateResponseMeta `json:"meta"`
}
type TemplateData struct {
	OverallChangesData []RootCauseMetric  `json:"overall_changes_data"`
	PrimaryLevelData   []PrimaryLevelData `json:"primary_level_data"`
}
type TemplateThreshold struct {
	Metric           string  `json:"metric"`
	PercentageChange float64 `json:"percentage_change"`
	AbsoluteChange   float64 `json:"absolute_change"`
}

type TemplateConfig struct {
	Metrics    []string            `json:"metrics"`
	Thresholds []TemplateThreshold `json:"thresholds"`
}

var mockMeta = TemplateResponseMeta{
	PrimaryLevel: LevelMeta{
		ColumnName: "campaign",
	},
	SubLevel: LevelMeta{
		ColumnName: "keyword",
	},
}
var mockPrimaryData = []PrimaryLevelData{
	{
		Name:             "campaign_1",
		PercentageChange: 20,
		AbsoluteChange:   20,
		PreviousValue:    100,
		LastValue:        120,
		SubLevelData: []SubLevelData{
			{
				Name:             "keyword_1",
				PercentageChange: 50,
				AbsoluteChange:   25,
				PreviousValue:    50,
				LastValue:        75,
				RootCauseMetrics: []RootCauseMetric{},
			},
			{
				Name:             "keyword_2",
				PercentageChange: -10,
				AbsoluteChange:   -5,
				PreviousValue:    50,
				LastValue:        45,
				RootCauseMetrics: []RootCauseMetric{},
			},
		},
	},
	{
		Name:             "campaign_2",
		PercentageChange: -20,
		AbsoluteChange:   -20,
		PreviousValue:    100,
		LastValue:        80,
		SubLevelData: []SubLevelData{
			{
				Name:             "keyword_3",
				PercentageChange: -20,
				AbsoluteChange:   -20,
				PreviousValue:    100,
				LastValue:        80,
				RootCauseMetrics: []RootCauseMetric{},
			},
		},
	},
}
var mockPrimaryDataForLeads = []PrimaryLevelData{
	{
		Name:             "campaign_1",
		PercentageChange: 20,
		AbsoluteChange:   20,
		PreviousValue:    100,
		LastValue:        120,
		SubLevelData: []SubLevelData{
			{
				Name:             "keyword_1",
				PercentageChange: 50,
				AbsoluteChange:   25,
				PreviousValue:    50,
				LastValue:        75,
				RootCauseMetrics: []RootCauseMetric{
					{
						Metric:           "impressions",
						PercentageChange: 20,
					},
					{
						Metric:           "si_share",
						PercentageChange: 10,
					},
				},
			},
			{
				Name:             "keyword_2",
				PercentageChange: -10,
				AbsoluteChange:   -5,
				PreviousValue:    50,
				LastValue:        45,
				RootCauseMetrics: []RootCauseMetric{
					{
						Metric:           "impressions",
						PercentageChange: -10,
					},
					{
						Metric:           "si_share",
						PercentageChange: -5,
					},
				},
			},
		},
	},
	{
		Name:             "campaign_2",
		PercentageChange: -20,
		AbsoluteChange:   -20,
		PreviousValue:    100,
		LastValue:        80,
		SubLevelData: []SubLevelData{
			{
				Name:             "keyword_1",
				PercentageChange: -20,
				AbsoluteChange:   -20,
				PreviousValue:    100,
				LastValue:        80,
				RootCauseMetrics: []RootCauseMetric{
					{
						Metric:           "impressions",
						PercentageChange: -20,
					},
					{
						Metric:           "si_share",
						PercentageChange: -10,
					},
				},
			},
		},
	},
}
var mockHeaderData = []RootCauseMetric{
	{
		Metric:           "si_share",
		PercentageChange: -10,
	},
	{
		Metric:           "leads",
		PercentageChange: 10,
	},
}
var mockData = TemplateData{
	OverallChangesData: mockHeaderData,
	PrimaryLevelData:   mockPrimaryData,
}
var mockDataForLeads = TemplateData{
	OverallChangesData: mockHeaderData,
	PrimaryLevelData:   mockPrimaryDataForLeads,
}
var MockResponse = TemplateResponse{
	BreakdownAnalysis: mockData,
	Meta:              mockMeta,
}
var MockResponseLeads = TemplateResponse{
	BreakdownAnalysis: mockDataForLeads,
	Meta:              mockMeta,
}
