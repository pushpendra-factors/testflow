package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
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
	ResultCols  string   `json:"result_cols"`
}

type ChannelBreakdownResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}

type ChannelQueryResult struct {
	WidgetKvs      *map[string]interface{} `json:"widget_kvs"`
	BreakdownTable *ChannelBreakdownResult `json:"breakdown_table"`
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

// select value->>'impressions' as impressions, value->>'clicks' as clicks,
// value->>'average_cost' as cost_per_click, value->>'cost' as total_cost,
// value->>'conversions' as conversions, value->>'all_conversions' as all_conversions
// from adwords_documents where type=8 and timestamp between 20191120 and 20191120;
func getAdwordsWidgetKvs(projectId uint64, query *ChannelQuery) (*map[string]interface{}, error) {
	// Todo: Add cost_per_click.
	sqlStmnt := "SELECT SUM((value->>'impressions')::float) as %s, SUM((value->>'clicks')::float) as %s," +
		" " + "SUM((value->>'cost')::float) as %s, SUM((value->>'conversions')::float) as %s," +
		" " + "SUM((value->>'all_conversions')::float) as %s FROM adwords_documents" +
		" " + "WHERE type=? AND timestamp BETWEEN ? and ?"

	stmnt := fmt.Sprintf(sqlStmnt, CAColumnImpressions, CAColumnClicks,
		CAColumnTotalCost, CAColumnAllConversions, CAColumnAllConversions)

	if _, exists := documentTypeByAlias[query.FilterKey]; !exists {
		return nil, errors.New("no matching type alias for filter key in adwords document")
	}

	params := []interface{}{
		documentTypeByAlias[query.FilterKey],
		query.DateFrom,
		query.DateTo,
	}

	db := C.GetServices().Db
	rows, err := db.Raw(stmnt, params...).Rows()
	if err != nil {
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows)
	if err != nil {
		return nil, err
	}

	if len(resultRows) == 0 {
		log.Error("Aggregate query returned zero rows.")
		return nil, errors.New("no rows returned")
	}

	if len(resultRows) > 1 {
		log.Error("Aggregate query returned more than one row on get adwords widget kvs.")
	}

	widgetKvs := make(map[string]interface{})
	for i, k := range resultHeaders {
		widgetKvs[k] = resultRows[0][i]
	}

	return &widgetKvs, nil
}

func ExecuteChannelQuery(projectId uint64, query *ChannelQuery) (*ChannelQueryResult, int) {
	if !isValidChannel(query.Channel) || !isValidFilterKey(query.FilterKey) ||
		query.DateFrom == 0 || query.DateTo == 0 {
		return nil, http.StatusBadRequest
	}

	widgetKvs, err := getAdwordsWidgetKvs(projectId, query)
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error(
			"Failed to get adowords widget kvs.")
		return nil, http.StatusInternalServerError
	}

	return &ChannelQueryResult{WidgetKvs: widgetKvs}, http.StatusOK
}
