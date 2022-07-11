package delta

import (
	M "factors/model/model"
	"factors/model/store/memsql"
)

var bingadsRequiredDocumentTypes = []int{1, 2, 3, 4, 5, 6} //Refer M.BingadsDocumentTypeAlias for clarity

var bingadsMetricToCalcInfo = map[string]MetricCalculationInfo{
	M.Impressions: {Props: []PropInfo{{Name: M.Impressions}}, Operation: "sum"},
	M.Clicks:      {Props: []PropInfo{{Name: M.Clicks}}, Operation: "sum"},
	"spend":       {Props: []PropInfo{{Name: "spend"}}, Operation: "sum"},
	M.Conversions: {Props: []PropInfo{{Name: M.Conversions}}, Operation: "sum"},
}

var bingadsConstantInfo = map[string]string{
	memsql.CAFilterCampaign: M.FilterCampaign,
	memsql.CAFilterAdGroup:  M.FilterAdGroup,
	memsql.CAFilterKeyword:  M.FilterKeyword,
	"campaign_id":           M.BingadsCampaignID,
	"ad_group_id":           M.BingadsAdgroupID,
	"keyword_id":            M.BingadsKeywordID,
}
