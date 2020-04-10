package model

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type ChannelQuery struct {
	Channel     string `json:"channel"`
	FilterKey   string `json:"filter_key"`
	FilterValue string `json:"filter_value"`
	From        int64  `json:"from"` // unix timestamp
	To          int64  `json:"to"`   // unix timestamp
	Status      string `json:"status"`
	MatchType   string `json:"match_type"` // optional
	Breakdown   string `json:"breakdown"`
}

type ChannelBreakdownResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}

type ChannelQueryResultMeta struct {
	Currency string `json:"currency"`
}

type ChannelQueryResult struct {
	Metrics          *map[string]interface{} `json:"metrics"`
	MetricsBreakdown *ChannelBreakdownResult `json:"metrics_breakdown"`
	Meta             *ChannelQueryResultMeta `json:"meta"`
}

type ChannelQueryUnit struct {
	// Json tag should match with Query's class,
	// query dispatched based on this.
	Class string                  `json:"cl"`
	Query *ChannelQuery           `json:"query"`
	Meta  *map[string]interface{} `json:"meta"`
}

const CAChannelGoogleAds = "google_ads"
const CAChannelFacebookAds = "facebook_ads"
const CAChannelGroupKey = "group_key"

var CAChannels = []string{
	CAChannelGoogleAds,
	CAChannelFacebookAds,
}

const CAColumnValueAll = "all"

const (
	CAColumnImpressions          = "impressions"
	CAColumnClicks               = "clicks"
	CAColumnTotalCost            = "total_cost"
	CAColumnConversions          = "conversions"
	CAColumnAllConversions       = "all_conversions"
	CAColumnCostPerClick         = "cost_per_click"
	CAColumnConversionRate       = "conversion_rate"
	CAColumnCostPerConversion    = "cost_per_conversion"
	CAColumnFrequency            = "frequency"
	CAColumnReach                = "reach"
	CAColumnInlinePostEngagement = "inline_post_engagement"
	CAColumnUniqueClicks         = "unique_clicks"
	CAColumnName                 = "name"
	CAColumnPlatform             = "platform"
)

const (
	CAFilterCampaign = "campaign"
	CAFilterAdGroup  = "ad_group"
	CAFilterAd       = "ad"
	CAFilterKeyword  = "keyword"
	CAFilterQuery    = "query"
	CAFilterAdset    = "adset"
)

var CAFilters = []string{
	CAFilterCampaign,
	CAFilterAdGroup,
	CAFilterAd,
	CAFilterKeyword,
	CAFilterQuery,
	CAFilterAdset,
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
		query.From == 0 || query.To == 0 {
		return nil, http.StatusBadRequest
	}

	if query.Channel == "google_ads" {
		result, errCode := ExecuteAdwordsChannelQuery(projectId, query)
		if errCode != http.StatusOK {
			log.WithField("project_id", projectId).Error("Failed to execute adwords channel query.")
			return nil, http.StatusInternalServerError
		}
		return result, http.StatusOK
	}
	if query.Channel == "facebook_ads" {
		result, errCode := ExecuteFacebookChannelQuery(projectId, query)
		if errCode != http.StatusOK {
			log.WithField("project_id", projectId).Error("Failed to execute facebook channel query.")
			return nil, http.StatusInternalServerError
		}
		return result, http.StatusOK
	}
	return nil, http.StatusBadRequest
}
