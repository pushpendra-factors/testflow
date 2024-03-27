package model

import (
	U "factors/util"
)

type AccountAnalyticsQuery struct {
	AggregateFunction string          `json:"agFn"`
	AggregateProperty string          `json:"agPr"`
	Metric            string          `json:"agMe"`
	Filters           []QueryProperty `json:"fil"`
	From              int64           `json:"from"`
	To                int64           `json:"to"`
	Timezone          string          `json:"tz"`
	SegmentID         string          `json:"s_id"`
}

const (
	TotalAccountsMetric       = "total_accounts"
	HighEngagedAccountsMetric = "high_engaged_accounts"
)

// TODO - Add domain if required when hubspot or salesforce specific filters comes in.
var MapOfAccountAnalyticsMetricToFilters = map[string][]QueryProperty{
	TotalAccountsMetric: []QueryProperty{},
	HighEngagedAccountsMetric: []QueryProperty{
		{
			Type:      U.PropertyTypeCategorical,
			Property:  U.DP_ENGAGEMENT_LEVEL,
			Operator:  EqualsOp,
			Value:     ENGAGEMENT_LEVEL_HOT,
			LogicalOp: LOGICAL_OP_AND,
			Entity:    PropertyEntityUserGlobal,
		},
		{
			Type:      U.PropertyTypeCategorical,
			Property:  U.DP_ENGAGEMENT_LEVEL,
			Operator:  EqualsOp,
			Value:     ENGAGEMENT_LEVEL_WARM,
			LogicalOp: LOGICAL_OP_OR,
			Entity:    PropertyEntityUserGlobal,
		},
	},
}
