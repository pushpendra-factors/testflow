package model

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type ChannelQuery struct {
	Channel     string   `json:"channel"`
	FilterKey   string   `json:"filter_key"`
	FilterValue string   `json:"filter_value"`
	DateFrom    int64    `json:"date_from"`
	DateTo      int64    `json:"date_to"`
	Status      string   `json:"status"`
	MatchType   string   `json:"match_type"` // optional
	Breakdowns  []string `json:"breakdowns"`
}

type ChannelBreakdownResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}

type ChannelQueryResult struct {
	Metrics          *map[string]interface{} `json:"metrics"`
	MetricsBreakdown *ChannelBreakdownResult `json:"metrics_breakdown"`
}

const CAChannelGoogleAds = "google_ads"

var CAChannels = []string{
	CAChannelGoogleAds,
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

const (
	CAFilterCampaign = "campaign"
	CAFilterAdGroup  = "ad_group"
	CAFilterAd       = "ad"
	CAFilterKeyword  = "keyword"
	CAFilterQuery    = "query"
)

var CAFilters = []string{
	CAFilterCampaign,
	CAFilterAdGroup,
	CAFilterAd,
	CAFilterKeyword,
	CAFilterQuery,
}

func isValidFilterKey(filter string) bool {
	for _, f := range CAFilters {
		if filter == f {
			return true
		}
	}

	return false
}

func isValidChannel(channel string) bool {
	for _, c := range CAChannels {
		if channel == c {
			return true
		}
	}

	return false
}

func GetChannelFilterValues(projectId uint64, channel, filter string) ([]string, int) {
	if !isValidChannel(channel) || !isValidFilterKey(filter) {
		return []string{}, http.StatusBadRequest
	}

	// supports only adwords now.
	docType, err := GetAdwordsDocumentTypeForFilterKey(filter)
	if err != nil {
		return []string{}, http.StatusInternalServerError
	}

	filterValues, errCode := GetAdwordsFilterValuesByType(projectId, docType)
	if errCode != http.StatusFound {
		return []string{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

func ExecuteChannelQuery(projectId uint64, query *ChannelQuery) (*ChannelQueryResult, int) {
	if !isValidChannel(query.Channel) || !isValidFilterKey(query.FilterKey) ||
		query.DateFrom == 0 || query.DateTo == 0 {
		return nil, http.StatusBadRequest
	}

	// supports only adwords now.
	metricKvs, err := GetAdwordsMetricKvs(projectId, query)
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error(
			"Failed to get adowords metric kvs.")
		return nil, http.StatusInternalServerError
	}

	return &ChannelQueryResult{Metrics: metricKvs}, http.StatusOK
}
