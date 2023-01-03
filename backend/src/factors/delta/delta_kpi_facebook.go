package delta

import (
	M "factors/model/model"
	"factors/model/store/memsql"
	"fmt"
	"strings"
)

var facebookRequiredDocumentTypes = []int{1, 2, 3, 4, 5, 6} //Refer memsql.FacebookDocumentTypeAlias for clarity

var facebookMetricToCalcInfo = map[string]MetricCalculationInfo{
	M.Impressions: {
		Props:     []ChannelPropInfo{{Name: M.Impressions}},
		Operation: "sum",
	},
	M.Clicks: {
		Props:     []ChannelPropInfo{{Name: M.Clicks}},
		Operation: "sum",
	},
	"link_clicks": {
		Props:     []ChannelPropInfo{{Name: "inline_link_clicks"}},
		Operation: "sum",
	},
	"spend": {
		Props:     []ChannelPropInfo{{Name: "spend"}},
		Operation: "sum",
	},
	"video_p50_watched_actions": {
		Props:     []ChannelPropInfo{{Name: "video_p50_watched_actions"}},
		Operation: "sum",
	},
	"video_p25_watched_actions": {
		Props:     []ChannelPropInfo{{Name: "video_p25_watched_actions"}},
		Operation: "sum",
	},
	"video_30_sec_watched_actions": {
		Props:     []ChannelPropInfo{{Name: "video_30_sec_watched_actions"}},
		Operation: "sum",
	},
	"video_p100_watched_actions": {
		Props:     []ChannelPropInfo{{Name: "video_p100_watched_actions"}},
		Operation: "sum",
	},
	"video_p75_watched_actions": {
		Props:     []ChannelPropInfo{{Name: "video_p75_watched_actions"}},
		Operation: "sum",
	},
	"reach": {
		Props:     []ChannelPropInfo{{Name: "reach"}},
		Operation: "sum",
	},
	"fb_pixel_purchase_count": {
		Props:     []ChannelPropInfo{{Name: "actions_offsite_conversion.fb_pixel_purchase"}},
		Operation: "sum",
	},
	"fb_pixel_purchase_revenue": {
		Props:     []ChannelPropInfo{{Name: "action_values_offsite_conversion.fb_pixel_purchase"}},
		Operation: "sum",
	},
	"cost_per_click": {
		Props: []ChannelPropInfo{
			{Name: "spend"},
			{Name: M.Clicks, ReplaceValue: map[float64]float64{0: 100000, 100000: 0}},
		},
		Operation: "quotient",
	},
	"cost_per_link_click": {
		Props: []ChannelPropInfo{
			{Name: "spend"},
			{Name: "inline_link_clicks", ReplaceValue: map[float64]float64{0: 100000, 100000: 0}},
		},
		Operation: "quotient",
	},
	"cost_per_thousand_impressions": {
		Props: []ChannelPropInfo{
			{Name: "spend"},
			{Name: M.Impressions, ReplaceValue: map[float64]float64{0: 100000, 100000: 0}},
		},
		Operation: "quotient",
		Constants: map[string]float64{"product": 1000},
	},
	"click_through_rate": {
		Props: []ChannelPropInfo{
			{Name: M.Clicks},
			{Name: M.Impressions, ReplaceValue: map[float64]float64{0: 100000, 100000: 0}},
		},
		Operation: "quotient",
		Constants: map[string]float64{"product": 100},
	},
	"link_click_through_rate": {
		Props: []ChannelPropInfo{
			{Name: "inline_link_clicks"},
			{Name: M.Impressions, ReplaceValue: map[float64]float64{0: 100000, 100000: 0}},
		},
		Operation: "quotient",
		Constants: map[string]float64{"product": 100},
	},
	"frequency": {
		Props: []ChannelPropInfo{
			{Name: M.Impressions},
			{Name: "reach", ReplaceValue: map[float64]float64{0: 100000, 100000: 0}},
		},
		Operation: "quotient",
	},
	"fb_pixel_purchase_cost_per_action_type": {
		Props: []ChannelPropInfo{
			{Name: "spend"},
			{Name: "actions_offsite_conversion.fb_pixel_purchase", ReplaceValue: map[float64]float64{0: 100000, 100000: 0}},
		},
		Operation: "quotient",
	},
	"fb_pixel_purchase_roas": {
		Props: []ChannelPropInfo{
			{Name: "action_values_offsite_conversion.fb_pixel_purchase"},
			{Name: "spend", ReplaceValue: map[float64]float64{0: 100000, 100000: 0}},
		},
		Operation: "quotient",
	},
}

var facebookConstantInfo = map[string]string{
	memsql.CAFilterCampaign: memsql.FacebookCampaign,
	memsql.CAFilterAdGroup:  memsql.FacebookAdSet,
	memsql.CAFilterAd:       memsql.FacebookAd,
}

func getFacebookFilterPropertyReportName(propName string, objectType string) (string, error) {
	propNameTrimmed := strings.TrimPrefix(propName, objectType+"_")

	if _, ok := facebookConstantInfo[objectType]; !ok {
		return "", fmt.Errorf("unknown object type: %s", objectType)
	}

	if name, ok := M.ObjectToValueInFacebookJobsMapping[fmt.Sprintf("%s:%s", facebookConstantInfo[objectType], propNameTrimmed)]; ok {
		return name, nil
	}
	return "", fmt.Errorf("filter property report name not found for %s", propName)
}

func getFacebookPropertyFilterName(prop string) (string, error) {
	propWithType := strings.SplitN(prop, "#", 2)
	objType := propWithType[0]
	name := propWithType[1]

	if _, ok := facebookConstantInfo[objType]; !ok {
		return prop, fmt.Errorf("unknown object type: %s", objType)
	}

	for k, v := range M.ObjectToValueInFacebookJobsMapping {
		if v == name {
			tmpProp := strings.SplitN(k, ":", 2)
			if tmpProp[0] == facebookConstantInfo[objType] {
				reqName := strings.Join([]string{objType, objType + "_" + tmpProp[1]}, "#")
				return reqName, nil
			}
		}
	}

	return prop, fmt.Errorf("property filter name not found for %s", prop)
}
