package delta

import (
	M "factors/model/model"
	"factors/model/store/memsql"
)

var linkedinRequiredDocumentTypes = []int{1, 2, 3, 4, 5, 6} //Refer memsql.LinkedinDocumentTypeAlias for clarity

var linkedinMetricToCalcInfo = map[string]MetricCalculationInfo{
	M.Impressions:                      {Props: []PropInfo{{Name: M.Impressions}}, Operation: "sum"},
	M.Clicks:                           {Props: []PropInfo{{Name: M.Clicks}}, Operation: "sum"},
	"cost":                             {Props: []PropInfo{{Name: "cost"}}, Operation: "sum", Constants: map[string]float64{"product": 100}},
	M.Conversion:                       {Props: []PropInfo{{Name: M.Conversion}}, Operation: "sum"},
	M.ClickThroughRate:                 {Props: []PropInfo{{Name: M.Clicks}, {Name: M.Impressions}}, Operation: "sum", Constants: map[string]float64{"product": 100}},
	M.ConversionRate:                   {Props: []PropInfo{{Name: M.Conversion}, {Name: M.Clicks}}, Operation: "sum", Constants: map[string]float64{"product": 100}},
	M.CostPerClick:                     {Props: []PropInfo{{Name: "cost"}, {Name: M.Clicks}}, Operation: "sum", Constants: map[string]float64{"quotient": 1000000}},
	M.CostPerConversion:                {Props: []PropInfo{{Name: "cost"}, {Name: M.Conversion}}, Operation: "sum", Constants: map[string]float64{"quotient": 1000000}},
	M.SearchImpressionShare:            {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchImpressionShare}, {Name: M.TotalSearchImpression, Dependent: M.SearchImpressionShare}}, Operation: "sum"},
	M.SearchClickShare:                 {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchClickShare}, {Name: M.TotalSearchClick, Dependent: M.SearchClickShare}}, Operation: "sum"},
	M.SearchTopImpressionShare:         {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchTopImpressionShare}, {Name: M.TotalSearchTopImpression, Dependent: M.SearchTopImpressionShare}}, Operation: "sum"},
	M.SearchAbsoluteTopImpressionShare: {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchAbsoluteTopImpressionShare}, {Name: M.TotalSearchAbsoluteTopImpression, Dependent: M.SearchAbsoluteTopImpressionShare}}, Operation: "sum"},
	M.SearchBudgetLostAbsoluteTopImpressionShare: {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchBudgetLostAbsoluteTopImpressionShare}, {Name: M.TotalSearchBudgetLostAbsoluteTopImpression, Dependent: M.SearchBudgetLostAbsoluteTopImpressionShare}}, Operation: "sum"},
	M.SearchBudgetLostImpressionShare:            {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchBudgetLostImpressionShare}, {Name: M.TotalSearchBudgetLostImpression, Dependent: M.SearchBudgetLostImpressionShare}}, Operation: "sum"},
	M.SearchBudgetLostTopImpressionShare:         {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchBudgetLostTopImpressionShare}, {Name: M.TotalSearchBudgetLostTopImpression, Dependent: M.SearchBudgetLostTopImpressionShare}}, Operation: "sum"},
	M.SearchRankLostAbsoluteTopImpressionShare:   {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchRankLostAbsoluteTopImpressionShare}, {Name: M.TotalSearchRankLostAbsoluteTopImpression, Dependent: M.SearchRankLostAbsoluteTopImpressionShare}}, Operation: "sum"},
	M.SearchRankLostImpressionShare:              {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchRankLostImpressionShare}, {Name: M.TotalSearchRankLostImpression, Dependent: M.SearchRankLostImpressionShare}}, Operation: "sum"},
	M.SearchRankLostTopImpressionShare:           {Props: []PropInfo{{Name: M.Impressions, Dependent: M.SearchRankLostTopImpressionShare}, {Name: M.TotalSearchRankLostTopImpression, Dependent: M.SearchRankLostTopImpressionShare}}, Operation: "sum"},
	M.ConversionValue:                            {Props: []PropInfo{{Name: M.ConversionValue}}, Operation: "sum"},
}

var linkedinConstantInfo = map[string]string{
	memsql.CAFilterCampaign: M.LinkedinCampaignGroup,
	memsql.CAFilterAdGroup:  M.LinkedinCampaign,
	memsql.CAFilterKeyword:  M.LinkedinCreative,
	"campaign_id":           M.LinkedinCampaignID,
	"ad_group_id":           M.LinkedinAdgroupID,
	"keyword_id":            M.LinkedinCreative + "_id",
}
