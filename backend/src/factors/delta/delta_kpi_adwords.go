package delta

import (
	M "factors/model/model"
	"factors/model/store/memsql"
	"fmt"
	"strings"
)

var adwordsRequiredDocumentTypes = []int{1, 3, 5, 8, 10} //Refer M.AdwordsDocumentTypeAlias for clarity

var adwordsMetricToCalcInfo = map[string]MetricCalculationInfo{
	M.Impressions: {
		Props:     []ChannelPropInfo{{Name: M.Impressions}},
		Operation: "sum",
	},
	M.Clicks: {
		Props:     []ChannelPropInfo{{Name: M.Clicks}},
		Operation: "sum",
	},
	"cost": {
		Props:     []ChannelPropInfo{{Name: "cost"}},
		Operation: "sum",
		Constants: map[string]float64{"quotient": 1000000},
	},
	M.Conversions: {
		Props: []ChannelPropInfo{
			{Name: M.Conversions},
		},
		Operation: "sum",
	},
	M.ClickThroughRate: {
		Props: []ChannelPropInfo{
			{Name: M.Clicks},
			{Name: M.Impressions, ReplaceValue: map[float64]float64{0: 100000}},
		},
		Operation: "sum",
		Constants: map[string]float64{"product": 100},
	},
	M.ConversionRate: {
		Props: []ChannelPropInfo{
			{Name: M.Conversions},
			{Name: M.Clicks, ReplaceValue: map[float64]float64{0: 100000}},
		},
		Operation: "quotient",
		Constants: map[string]float64{"product": 100},
	},
	M.CostPerClick: {
		Props: []ChannelPropInfo{
			{Name: "cost"},
			{Name: M.Clicks, ReplaceValue: map[float64]float64{0: 100000}},
		},
		Operation: "quotient",
		Constants: map[string]float64{"quotient": 1000000},
	},
	M.CostPerConversion: {
		Props: []ChannelPropInfo{
			{Name: "cost"},
			{Name: M.Conversions, ReplaceValue: map[float64]float64{0: 100000}},
		},
		Operation: "quotient",
		Constants: map[string]float64{"quotient": 1000000},
	},
	M.SearchImpressionShare: {
		Props: []ChannelPropInfo{
			{Name: M.Impressions, DependentKey: M.SearchImpressionShare, DependentValue: 0, DependentOperation: "!="},
			{Name: M.TotalSearchImpression, DependentKey: M.SearchImpressionShare},
		},
		Operation: "quotient",
	},
	M.SearchClickShare: {
		Props: []ChannelPropInfo{
			{Name: M.Impressions, DependentKey: M.SearchClickShare, DependentValue: 0, DependentOperation: "!="},
			{Name: M.TotalSearchClick, DependentKey: M.SearchClickShare},
		},
		Operation: "quotient",
	},
	M.SearchTopImpressionShare: {
		Props: []ChannelPropInfo{
			{Name: M.TopImpressions, DependentKey: M.SearchTopImpressionShare, DependentValue: 0, DependentOperation: "!="},
			{Name: M.TotalTopImpressions, DependentKey: M.SearchTopImpressionShare},
		},
		Operation: "quotient",
	},
	M.SearchAbsoluteTopImpressionShare: {
		Props: []ChannelPropInfo{
			{Name: M.AbsoluteTopImpressions, DependentKey: M.SearchAbsoluteTopImpressionShare, DependentValue: 0, DependentOperation: "!="},
			{Name: M.TotalTopImpressions, DependentKey: M.SearchAbsoluteTopImpressionShare},
		},
		Operation: "quotient",
	},
	M.SearchBudgetLostAbsoluteTopImpressionShare: {
		Props: []ChannelPropInfo{
			{Name: M.AbsoluteTopImpressionLostDueToBudget, DependentKey: M.SearchBudgetLostAbsoluteTopImpressionShare, DependentValue: 0, DependentOperation: "!="},
			{Name: M.TotalTopImpressions, DependentKey: M.SearchBudgetLostAbsoluteTopImpressionShare},
		},
		Operation: "quotient",
	},
	M.SearchBudgetLostImpressionShare: {
		Props: []ChannelPropInfo{
			{Name: M.ImpressionLostDueToBudget, DependentKey: M.SearchBudgetLostImpressionShare, DependentValue: 0, DependentOperation: "!="},
			{Name: M.TotalSearchImpression, DependentKey: M.SearchBudgetLostImpressionShare},
		},
		Operation: "quotient",
	},
	M.SearchBudgetLostTopImpressionShare: {
		Props: []ChannelPropInfo{
			{Name: M.TopImpressionLostDueToBudget, DependentKey: M.SearchBudgetLostTopImpressionShare, DependentValue: 0, DependentOperation: "!="},
			{Name: M.TotalTopImpressions, DependentKey: M.SearchBudgetLostTopImpressionShare},
		},
		Operation: "quotient",
	},
	M.SearchRankLostAbsoluteTopImpressionShare: {
		Props: []ChannelPropInfo{
			{Name: M.AbsoluteTopImpressionLostDueToRank, DependentKey: M.SearchRankLostAbsoluteTopImpressionShare, DependentValue: 0, DependentOperation: "!="},
			{Name: M.TotalTopImpressions, DependentKey: M.SearchRankLostAbsoluteTopImpressionShare},
		},
		Operation: "quotient",
	},
	M.SearchRankLostImpressionShare: {
		Props: []ChannelPropInfo{
			{Name: M.ImpressionLostDueToRank, DependentKey: M.SearchRankLostImpressionShare, DependentValue: 0, DependentOperation: "!="},
			{Name: M.TotalSearchImpression, DependentKey: M.SearchRankLostImpressionShare},
		},
		Operation: "quotient",
	},
	M.SearchRankLostTopImpressionShare: {
		Props: []ChannelPropInfo{
			{Name: M.TopImpressionLostDueToRank, DependentKey: M.SearchRankLostTopImpressionShare, DependentValue: 0, DependentOperation: "!="},
			{Name: M.TotalTopImpressions, DependentKey: M.SearchRankLostTopImpressionShare},
		},
		Operation: "quotient",
	},
	M.ConversionValue: {
		Props:     []ChannelPropInfo{{Name: M.ConversionValue}},
		Operation: "sum",
	},
}

var adwordsConstantInfo = map[string]string{
	memsql.CAFilterCampaign: M.AdwordsCampaign,
	memsql.CAFilterAdGroup:  M.AdwordsAdGroup,
	memsql.CAFilterKeyword:  M.AdwordsKeyword,
}

func getAdwordsFilterPropertyReportName(propName string, objectType string) (string, error) {
	propNameTrimmed := strings.TrimPrefix(propName, objectType+"_")

	if _, ok := adwordsConstantInfo[objectType]; !ok {
		return "", fmt.Errorf("unknown object type: %s", objectType)
	}
	if name, ok := M.AdwordsInternalPropertiesToReportsInternal[fmt.Sprintf("%s:%s", adwordsConstantInfo[objectType], propNameTrimmed)]; ok {
		return name, nil
	}
	return "", fmt.Errorf("filter property report name not found for %s", propName)
}

func getAdwordsPropertyFilterName(prop string) (string, error) {
	propWithType := strings.SplitN(prop, "#", 2)
	objType := propWithType[0]
	name := propWithType[1]

	if _, ok := adwordsConstantInfo[objType]; !ok {
		return prop, fmt.Errorf("unknown object type: %s", objType)
	}

	for k, v := range M.AdwordsInternalPropertiesToReportsInternal {
		if v == name {
			tmpProp := strings.SplitN(k, ":", 2)
			if tmpProp[0] == adwordsConstantInfo[objType] {
				reqName := strings.Join([]string{objType, objType + "_" + tmpProp[1]}, "#")
				return reqName, nil
			}
		}
	}

	return prop, fmt.Errorf("property filter name not found for %s", prop)
}
