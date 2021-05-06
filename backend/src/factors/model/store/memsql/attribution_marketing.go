package memsql

import (
	"factors/model/model"
	U "factors/util"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) FetchMarketingReports(projectID uint64, q model.AttributionQuery, projectSetting model.ProjectSetting) (*model.MarketingReports, error) {

	data := &model.MarketingReports{}
	var err error

	// Get adwords, facebook, linkedin reports.
	effectiveFrom := lookbackAdjustedFrom(q.From, q.LookbackDays)
	effectiveTo := q.To
	// Extend the campaign window for engagement based attribution.
	if q.QueryType == model.AttributionQueryTypeEngagementBased {
		effectiveFrom = lookbackAdjustedFrom(q.From, q.LookbackDays)
		effectiveTo = lookbackAdjustedTo(q.To, q.LookbackDays)
	}

	adwordsCustomerID := *projectSetting.IntAdwordsCustomerAccountId

	var adwordsGCLIDData map[string]model.CampaignInfo
	var reportType int
	var adwordsCampaignIDData, adwordsCampaignNameData, adwordsAdgroupIDData, adwordsAdgroupNameData, adwordsKeywordIDData, adwordsKeywordNameData map[string]model.MarketingData
	// Adwords.
	if adwordsCustomerID != "" && model.DoesAdwordsReportExist(q.AttributionKey) {

		adwordsGCLIDData, err = store.GetGCLIDBasedCampaignInfo(projectID, effectiveFrom, effectiveTo, adwordsCustomerID)
		if err != nil {
			return data, err
		}

		reportType = model.AdwordsDocumentTypeAlias[model.CampaignPerformanceReport] // 5
		adwordsCampaignIDData, adwordsCampaignNameData, err = store.PullAdwordsMarketingData(projectID, effectiveFrom,
			effectiveTo, adwordsCustomerID, model.AdwordsCampaignID, model.AdwordsCampaignName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
		reportType = model.AdwordsDocumentTypeAlias[model.AdGroupPerformanceReport] // 10
		adwordsAdgroupIDData, adwordsAdgroupNameData, err = store.PullAdwordsMarketingData(projectID, effectiveFrom,
			effectiveTo, adwordsCustomerID, model.AdwordsAdgroupID, model.AdwordsAdgroupName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
		reportType = model.AdwordsDocumentTypeAlias[model.KeywordPerformanceReport] // 8
		adwordsKeywordIDData, adwordsKeywordNameData, err = store.PullAdwordsMarketingData(projectID, effectiveFrom,
			effectiveTo, adwordsCustomerID, model.AdwordsKeywordID, model.AdwordsKeywordName, model.AdwordsKeywordMatchType, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
	}

	// Facebook.
	var facebookCampaignIDData, facebookCampaignNameData, facebookAdgroupIDData, facebookAdgroupNameData map[string]model.MarketingData
	if projectSetting.IntFacebookAdAccount != "" && model.DoesFBReportExist(q.AttributionKey) {
		facebookCustomerID := projectSetting.IntFacebookAdAccount

		reportType = facebookDocumentTypeAlias["campaign_insights"] // 5
		facebookCampaignIDData, facebookCampaignNameData, err = store.PullFacebookMarketingData(projectID, effectiveFrom,
			effectiveTo, facebookCustomerID, model.FacebookCampaignID, model.FacebookCampaignName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
		reportType = facebookDocumentTypeAlias["ad_set_insights"] // 5
		facebookAdgroupIDData, facebookAdgroupNameData, err = store.PullFacebookMarketingData(projectID, effectiveFrom,
			effectiveTo, facebookCustomerID, model.FacebookAdgroupID, model.FacebookAdgroupName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
	}

	// Linkedin.
	var linkedinCampaignIDData, linkedinCampaignNameData, linkedinAdgroupIDData, linkedinAdgroupNameData map[string]model.MarketingData
	if projectSetting.IntLinkedinAdAccount != "" && model.DoesLinkedinReportExist(q.AttributionKey) {
		linkedinCustomerID := projectSetting.IntLinkedinAdAccount

		reportType = linkedinDocumentTypeAlias["campaign_group_insights"] // 5
		linkedinCampaignIDData, linkedinCampaignNameData, err = store.PullLinkedinMarketingData(projectID, effectiveFrom,
			effectiveTo, linkedinCustomerID, model.LinkedinCampaignID, model.LinkedinCampaignName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
		reportType = linkedinDocumentTypeAlias["campaign_insights"] // 6
		linkedinAdgroupIDData, linkedinAdgroupNameData, err = store.PullLinkedinMarketingData(projectID, effectiveFrom,
			effectiveTo, linkedinCustomerID, model.LinkedinAdgroupID, model.LinkedinAdgroupName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
	}

	data.AdwordsGCLIDData = adwordsGCLIDData
	data.AdwordsCampaignIDData = adwordsCampaignIDData
	data.AdwordsCampaignNameData = adwordsCampaignNameData

	data.AdwordsAdgroupIDData = adwordsAdgroupIDData
	data.AdwordsAdgroupNameData = adwordsAdgroupNameData

	data.AdwordsKeywordIDData = adwordsKeywordIDData
	data.AdwordsKeywordNameData = adwordsKeywordNameData

	data.FacebookCampaignIDData = facebookCampaignIDData
	data.FacebookCampaignNameData = facebookCampaignNameData

	data.FacebookAdgroupIDData = facebookAdgroupIDData
	data.FacebookAdgroupNameData = facebookAdgroupNameData

	data.LinkedinCampaignIDData = linkedinCampaignIDData
	data.LinkedinCampaignNameData = linkedinCampaignNameData

	data.LinkedinAdgroupIDData = linkedinAdgroupIDData
	data.LinkedinAdgroupNameData = linkedinAdgroupNameData

	return data, err
}

// PullAdwordsMarketingData Pulls Adds channel data for Adwords.
func (store *MemSQL) PullAdwordsMarketingData(projectID uint64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, timeZone string) (map[string]model.MarketingData, map[string]model.MarketingData, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})
	customerAccountIDs := strings.Split(customerAccountID, ",")
	performanceQuery := "SELECT campaign_id as campaignID, ad_group_id as adgroupID, keyword_id as keywordID, ad_id as adID, " +
		"JSON_EXTRACT_STRING(value, ?) AS key_id, JSON_EXTRACT_STRING(value, ?) AS key_name, JSON_EXTRACT_STRING(value, ?) AS extra_value1, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'cost'))/1000000 AS total_cost FROM adwords_documents " +
		"where project_id = ? AND customer_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, adID, key_id, key_name, extra_value1"

	rows, err := store.ExecQueryWithContext(performanceQuery, []interface{}{keyID, keyName, extraValue1, projectID, customerAccountIDs, reportType,
		U.GetDateAsStringZ(from, U.TimeZoneString(timeZone)), U.GetDateAsStringZ(to, U.TimeZoneString(timeZone))})
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer rows.Close()

	marketingDataKeyIdMap, marketingDataKeyNameMap := model.ProcessRow(rows, err, logCtx)
	return marketingDataKeyIdMap, marketingDataKeyNameMap, nil
}

// PullFacebookMarketingData Pulls Adds channel data for Facebook.
func (store *MemSQL) PullFacebookMarketingData(projectID uint64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, timeZone string) (map[string]model.MarketingData, map[string]model.MarketingData, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})
	customerAccountIDs := strings.Split(customerAccountID, ",")
	performanceQuery := "SELECT campaign_id as campaignID, ad_set_id as adgroupID, '$none' as keywordID, ad_id as adID, " +
		"JSON_EXTRACT_STRING(value, ?) AS key_id, JSON_EXTRACT_STRING(value, ?) AS key_name, JSON_EXTRACT_STRING(value, ?) AS extra_value1, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'spend')) AS total_cost FROM facebook_documents " +
		"where project_id = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, adID, key_id, key_name, extra_value1"

	rows, err := store.ExecQueryWithContext(performanceQuery, []interface{}{keyID, keyName, extraValue1, projectID, customerAccountIDs, reportType,
		U.GetDateAsStringZ(from, U.TimeZoneString(timeZone)), U.GetDateAsStringZ(to, U.TimeZoneString(timeZone))})
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer rows.Close()

	marketingDataKeyIdMap, marketingDataKeyNameMap := model.ProcessRow(rows, err, logCtx)
	return marketingDataKeyIdMap, marketingDataKeyNameMap, nil
}

// PullLinkedinMarketingData Pulls Adds channel data for Linkedin.
func (store *MemSQL) PullLinkedinMarketingData(projectID uint64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, timeZone string) (map[string]model.MarketingData, map[string]model.MarketingData, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})
	customerAccountIDs := strings.Split(customerAccountID, ",")
	performanceQuery := "SELECT campaign_group_id as campaignID, campaign_id as adgroupID, '$none' as keywordID, creative_id as adID, " +
		"JSON_EXTRACT_STRING(value, ?) AS key_id, JSON_EXTRACT_STRING(value, ?) AS key_name, JSON_EXTRACT_STRING(value, ?) AS extra_value1, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM((JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'costInLocalCurrency')) AS total_spend FROM linkedin_documents " +
		"where project_id = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, adID, key_id, key_name, extra_value1"

	rows, err := store.ExecQueryWithContext(performanceQuery, []interface{}{keyID, keyName, extraValue1, projectID, customerAccountIDs, reportType,
		U.GetDateAsStringZ(from, U.TimeZoneString(timeZone)), U.GetDateAsStringZ(to, U.TimeZoneString(timeZone))})
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer rows.Close()

	marketingDataKeyIdMap, marketingDataKeyNameMap := model.ProcessRow(rows, err, logCtx)
	return marketingDataKeyIdMap, marketingDataKeyNameMap, nil
}
