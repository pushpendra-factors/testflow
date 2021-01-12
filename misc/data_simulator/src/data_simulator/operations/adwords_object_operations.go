package operations

import (
	"math/rand"
	"strconv"
	"time"
)

const campaignPerformanceReportTypeAlias = "campaign_performance_report"
const campaignPerformanceReportType = 5

func GetCampaignPerfReportAdwordsDoc(yesterday time.Time, projectIDStage uint64, adwordsCustomerAccountIDStage, campaignName string, totalEvents int) AdwordsDocument {

	values := map[string]string{}

	min := 1.2
	budgetPerClick := 1000000 // with conversionFactor of 100k
	probability := min + rand.Float64()
	// 1.5 is upper limit
	for probability > 1.5 {
		probability = min + rand.Float64()
	}
	impressionsP := rand.Intn(5) + 8
	campaignId := strconv.Itoa(int(Hash(campaignName)))
	cost := strconv.FormatInt(int64(probability*float64(budgetPerClick)*float64(totalEvents)), 10)
	clicks := strconv.Itoa(int(probability * float64(totalEvents)))
	impressions := strconv.Itoa(totalEvents * impressionsP)

	values["cost"] = cost
	values["amount"] = "20000000"
	values["clicks"] = clicks
	values["budget_id"] = "991039825"
	values["start_date"] = "2016-12-09"
	values["bounce_rate"] = "0.00%"
	values["campaign_id"] = campaignId
	values["conversions"] = "0.00"
	values["engagements"] = "0"
	values["gmail_saves"] = "0"
	values["impressions"] = impressions
	values["video_views"] = "0"
	values["average_cost"] = "0"
	values["interactions"] = "0"
	values["campaign_name"] = campaignName
	values["gmail_forwards"] = "0"
	values["invalid_clicks"] = "0"
	values["all_conversions"] = "0.00"
	values["campaign_status"] = "active"
	values["conversion_rate"] = "0.00%"
	values["video_view_rate"] = "0.00%"
	values["base_campaign_id"] = campaignId
	values["conversion_value"] = "0.00"
	values["all_conversion_rate"] = "0.00%"
	values["campaign_trial_type"] = "base campaign"
	values["cost_per_conversion"] = "0"
	values["all_conversion_value"] = "0.00"
	values["average_time_on_site"] = "0"
	values["gmail_secondary_clicks"] = "0"
	values["video_quartile_25_rate"] = "0.00%"
	values["video_quartile_50_rate"] = "0.00%"
	values["video_quartile_75_rate"] = "0.00%"
	values["active_view_impressions"] = "0"
	values["active_view_viewability"] = "0.00%"
	values["video_quartile_100_rate"] = "0.00%"
	values["value_per_all_conversion"] = "0.00"
	values["view_through_conversions"] = "0"
	values["active_view_measurability"] = "0.00%"
	values["click_assisted_conversions"] = "0"
	values["active_view_measurable_cost"] = "0"
	values["is_budget_explicitly_shared"] = "false"
	values["impression_assisted_conversions"] = "0"
	values["active_view_measurable_impressions"] = "0"

	adwordsDoc := AdwordsDocument{}
	adwordsDoc.ProjectID = projectIDStage
	adwordsDoc.CustomerAccountID = adwordsCustomerAccountIDStage
	adwordsDoc.TypeAlias = campaignPerformanceReportTypeAlias // Type = 5
	timeStr := yesterday.Format("20060102")
	intDate64, _ := strconv.ParseInt(timeStr, 10, 64)
	adwordsDoc.Timestamp = intDate64
	adwordsDoc.ID = campaignId
	adwordsDoc.Value = values

	return adwordsDoc
}
