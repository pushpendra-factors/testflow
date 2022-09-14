package delta

import (
	M "factors/model/model"
	"strings"
)

var googleOrganicRequiredDocumentTypes = []int{2} // 1:combined_performance_report, 2:page_performance_report

var googleOrganicMetricToCalcInfo = map[string]MetricCalculationInfo{
	M.Impressions: {
		Props:     []PropInfo{{Name: M.Impressions}},
		Operation: "sum",
	},
	M.Clicks: {
		Props:     []PropInfo{{Name: M.Clicks}},
		Operation: "sum",
	},
	M.ClickThroughRate: {
		Props: []PropInfo{
			{Name: "clicks"},
			{Name: "impressions", ReplaceValue: map[float64]float64{0: 100000, 100000: 0}},
		},
		Operation: "quotient",
		Constants: map[string]float64{"product": 100},
	},
	"position_avg": {
		Props:     []PropInfo{{Name: "position"}},
		Operation: "average",
	},
	"position_impression_weighted_avg": {
		Props: []PropInfo{
			{Name: "position", DependentKey: "impressions"},
			{Name: "impressions", ReplaceValue: map[float64]float64{0: 100000, 100000: 0}},
		},
		Operation: "quotient",
	},
}

var googleOrganicConstantInfo = map[string]string{
	"organic_property": "page",
	// "campaign_id":           "id",
	// "ad_group_id":           "id",
	// "keyword_id":            "",
}

func getGoogleOrganicFilterPropertyReportName(propName string, objectType string) (string, error) {
	propNameTrimmed := strings.TrimPrefix(propName, objectType+"_")

	return propNameTrimmed, nil
}

func getGoogleOrganicPropertyFilterName(prop string) (string, error) {
	propWithType := strings.SplitN(prop, "#", 2)
	objType := propWithType[0]
	name := propWithType[1]

	reqName := strings.Join([]string{objType, objType + "_" + name}, "#")
	return reqName, nil
}
