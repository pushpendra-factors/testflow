package model

import "net/http"

type ChannelQuery struct {
	Channel    string   `json:"channel"`
	Campaign   string   `json:"campaign"`
	AdGroup    string   `json:"ad_group"`
	Ad         string   `json:"ad"`
	Keyword    string   `json:"keyword"`
	Query      string   `json:"query"`
	DateFrom   int64    `json:"date_from"`
	DateTo     int64    `json:"date_to"`
	Status     string   `json:"status"`
	MatchType  string   `json:"match_type"` // optional
	Breakdowns []string `json:"breakdowns"`
	ResultCols string   `json:"result_cols"`
}

type ChannelBreakdownResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}

type ChannelQueryResult struct {
	WidgetKvs      map[string]int         `json:"widget_kvs"`
	BreakdownTable ChannelBreakdownResult `json:"breakdown_table"`
}

var Channels = []string{
	"google_ads",
}

const CAColumnValueAll = "all"

const (
	CAColumnImpressions    = "impressions"
	CAColumnClicks         = "clicks"
	CAColumnCostPerClick   = "cost_per_click"
	CAColumnTotalCost      = "total_cost"
	CAColumnConversions    = "conversions"
	CAColumnAllConversions = "all_conversions"
)

func ExecuteChannelQuery(query *ChannelQuery) (*ChannelQueryResult, int) {
	return &ChannelQueryResult{}, http.StatusOK
}
