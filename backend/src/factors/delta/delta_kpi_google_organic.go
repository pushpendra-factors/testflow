package delta

import (
	M "factors/model/model"
	"factors/model/store/memsql"
)

var googleOrganicRequiredDocumentTypes = []int{1, 2} // 1:combined_performance_report, 2:page_performance_report

var googleOrganicMetricToCalcInfo = map[string]MetricCalculationInfo{
	M.Impressions:                        {Props: []PropInfo{{Name: M.Impressions}}, Operation: "sum"},
	M.Clicks:                             {Props: []PropInfo{{Name: M.Clicks}}, Operation: "sum"},
	M.ClickThroughRate:                   {Props: []PropInfo{{Name: M.ClickThroughRate, Dependent: M.SearchBudgetLostAbsoluteTopImpressionShare}, {Name: M.TotalSearchBudgetLostAbsoluteTopImpression, Dependent: M.SearchBudgetLostAbsoluteTopImpressionShare}}, Operation: "sum"},
	"position_avg":                       {Props: []PropInfo{{Name: "position"}}, Operation: "sum"},
	M.SearchBudgetLostTopImpressionShare: {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchBudgetLostTopImpressionShare}, {Name: M.TotalSearchBudgetLostTopImpression, Dependent: M.SearchBudgetLostTopImpressionShare}}, Operation: "sum"},
	M.SearchRankLostAbsoluteTopImpressionShare: {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchRankLostAbsoluteTopImpressionShare}, {Name: M.TotalSearchRankLostAbsoluteTopImpression, Dependent: M.SearchRankLostAbsoluteTopImpressionShare}}, Operation: "sum"},
	M.SearchRankLostImpressionShare:            {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchRankLostImpressionShare}, {Name: M.TotalSearchRankLostImpression, Dependent: M.SearchRankLostImpressionShare}}, Operation: "sum"},
	M.SearchRankLostTopImpressionShare:         {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchRankLostTopImpressionShare}, {Name: M.TotalSearchRankLostTopImpression, Dependent: M.SearchRankLostTopImpressionShare}}, Operation: "sum"},
	M.ConversionValue:                          {Props: []PropInfo{{Name: M.ConversionValue}}, Operation: "sum"},
}

var googleOrganicConstantInfo = map[string]string{
	memsql.CAFilterCampaign: "combined",
	memsql.CAFilterAdGroup:  "page",
	memsql.CAFilterKeyword:  "",
	"campaign_id":           "id",
	"ad_group_id":           "id",
	"keyword_id":            "",
}
