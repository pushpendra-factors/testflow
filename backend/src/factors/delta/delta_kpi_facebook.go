package delta

import (
	M "factors/model/model"
	"factors/model/store/memsql"
)

var facebookRequiredDocumentTypes = []int{1, 2, 3, 4, 5, 6} //Refer memsql.FacebookDocumentTypeAlias for clarity

var facebookMetricToCalcInfo = map[string]MetricCalculationInfo{
	M.Impressions:                  {Props: []PropInfo{{Name: M.Impressions}}, Operation: "sum"},
	M.Clicks:                       {Props: []PropInfo{{Name: M.Clicks}}, Operation: "sum"},
	"link_clicks":                  {Props: []PropInfo{{Name: "inline_link_clicks"}}},
	"spend":                        {Props: []PropInfo{{Name: "spend"}}, Operation: "sum", Constants: map[string]float64{"product": 100}},
	"video_p50_watched_actions":    {Props: []PropInfo{{Name: "video_p50_watched_actions"}}, Operation: "sum"},
	"video_p25_watched_actions":    {Props: []PropInfo{{Name: "video_p25_watched_actions"}}, Operation: "sum"},
	"video_30_sec_watched_actions": {Props: []PropInfo{{Name: "video_30_sec_watched_actions"}}, Operation: "sum"},
	"video_p100_watched_actions":   {Props: []PropInfo{{Name: "video_p100_watched_actions"}}, Operation: "sum"},
	"video_p75_watched_actions":    {Props: []PropInfo{{Name: "video_p75_watched_actions"}}, Operation: "sum"},
	"reach":                        {Props: []PropInfo{{Name: "reach"}}, Operation: "sum"},
	"fb_pixel_purchase_count":      {Props: []PropInfo{{Name: "actions_offsite_conversion.fb_pixel_purchase"}}, Operation: "sum"},
	"fb_pixel_purchase_revenue":    {Props: []PropInfo{{Name: "action_values_offsite_conversion.fb_pixel_purchase"}}, Operation: "sum"},
	M.ClickThroughRate:             {Props: []PropInfo{{Name: M.Clicks}, {Name: M.Impressions}}, Operation: "sum", Constants: map[string]float64{"product": 100}},
	M.ConversionRate:               {Props: []PropInfo{{Name: M.Conversion}, {Name: M.Clicks}}, Operation: "sum", Constants: map[string]float64{"product": 100}},
	M.CostPerClick:                 {Props: []PropInfo{{Name: "cost"}, {Name: M.Clicks}}, Operation: "sum", Constants: map[string]float64{"quotient": 1000000}},
	M.CostPerConversion:            {Props: []PropInfo{{Name: "cost"}, {Name: M.Conversion}}, Operation: "sum", Constants: map[string]float64{"quotient": 1000000}},
	M.ConversionValue:              {Props: []PropInfo{{Name: M.ConversionValue}}, Operation: "sum"},
}

var facebookConstantInfo = map[string]string{
	memsql.CAFilterCampaign: memsql.FacebookCampaign,
	memsql.CAFilterAdGroup:  memsql.FacebookAdSet,
	memsql.CAFilterKeyword:  memsql.FacebookAd,
	"campaign_id":           M.FacebookCampaignID,
	"ad_group_id":           M.FacebookAdgroupID,
	"keyword_id":            memsql.FacebookAd + "_id",
}
