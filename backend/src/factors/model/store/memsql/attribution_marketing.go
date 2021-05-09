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

	var adwordsGCLIDData map[string]model.MarketingData
	var reportType int
	var adwordsCampaignIDData, adwordsAdgroupIDData, adwordsKeywordIDData map[string]model.MarketingData
	// Adwords.
	if adwordsCustomerID != "" && model.DoesAdwordsReportExist(q.AttributionKey) {

		reportType = model.AdwordsDocumentTypeAlias[model.CampaignPerformanceReport] // 5
		adwordsCampaignIDData, err = store.PullAdwordsMarketingData(projectID, effectiveFrom,
			effectiveTo, adwordsCustomerID, model.AdwordsCampaignID, model.AdwordsCampaignName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}

		reportType = model.AdwordsDocumentTypeAlias[model.AdGroupPerformanceReport] // 10
		adwordsAdgroupIDData, err = store.PullAdwordsMarketingData(projectID, effectiveFrom,
			effectiveTo, adwordsCustomerID, model.AdwordsAdgroupID, model.AdwordsAdgroupName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, value := range adwordsAdgroupIDData {
			campID := adwordsAdgroupIDData[id].CampaignID
			if U.IsNonEmptyKey(campID) {
				value.CampaignName = adwordsCampaignIDData[campID].Name
				adwordsAdgroupIDData[id] = value
			}
		}

		reportType = model.AdwordsDocumentTypeAlias[model.KeywordPerformanceReport] // 8
		adwordsKeywordIDData, err = store.PullAdwordsMarketingData(projectID, effectiveFrom,
			effectiveTo, adwordsCustomerID, model.AdwordsKeywordID, model.AdwordsKeywordName, model.AdwordsKeywordMatchType, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, value := range adwordsKeywordIDData {
			campID := adwordsKeywordIDData[id].CampaignID
			if U.IsNonEmptyKey(campID) {
				value.CampaignName = adwordsCampaignIDData[campID].Name
				adwordsKeywordIDData[id] = value
			}
		}
		for id, value := range adwordsKeywordIDData {
			adgroupID := adwordsKeywordIDData[id].AdgroupID
			if U.IsNonEmptyKey(adgroupID) {
				value.AdgroupName = adwordsAdgroupIDData[adgroupID].Name
				adwordsKeywordIDData[id] = value
			}
		}

		adwordsGCLIDData, err = store.GetGCLIDBasedCampaignInfo(projectID, effectiveFrom, effectiveTo, adwordsCustomerID,
			adwordsCampaignIDData, adwordsAdgroupIDData, adwordsKeywordIDData)
		if err != nil {
			return data, err
		}
	}

	// Facebook.
	var facebookCampaignIDData, facebookAdgroupIDData map[string]model.MarketingData
	if projectSetting.IntFacebookAdAccount != "" && model.DoesFBReportExist(q.AttributionKey) {
		facebookCustomerID := projectSetting.IntFacebookAdAccount

		reportType = facebookDocumentTypeAlias["campaign_insights"] // 5
		facebookCampaignIDData, err = store.PullFacebookMarketingData(projectID, effectiveFrom,
			effectiveTo, facebookCustomerID, model.FacebookCampaignID, model.FacebookCampaignName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
		reportType = facebookDocumentTypeAlias["ad_set_insights"] // 5
		facebookAdgroupIDData, err = store.PullFacebookMarketingData(projectID, effectiveFrom,
			effectiveTo, facebookCustomerID, model.FacebookAdgroupID, model.FacebookAdgroupName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, value := range facebookAdgroupIDData {
			campID := facebookAdgroupIDData[id].CampaignID
			if U.IsNonEmptyKey(campID) {
				value.CampaignName = facebookCampaignIDData[campID].Name
				facebookAdgroupIDData[id] = value
			}
		}
	}

	// Linkedin.
	var linkedinCampaignIDData, linkedinAdgroupIDData map[string]model.MarketingData
	if projectSetting.IntLinkedinAdAccount != "" && model.DoesLinkedinReportExist(q.AttributionKey) {
		linkedinCustomerID := projectSetting.IntLinkedinAdAccount

		reportType = linkedinDocumentTypeAlias["campaign_group_insights"] // 5
		linkedinCampaignIDData, err = store.PullLinkedinMarketingData(projectID, effectiveFrom,
			effectiveTo, linkedinCustomerID, model.LinkedinCampaignID, model.LinkedinCampaignName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
		reportType = linkedinDocumentTypeAlias["campaign_insights"] // 6
		linkedinAdgroupIDData, err = store.PullLinkedinMarketingData(projectID, effectiveFrom,
			effectiveTo, linkedinCustomerID, model.LinkedinAdgroupID, model.LinkedinAdgroupName, model.PropertyValueNone, reportType, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, value := range linkedinAdgroupIDData {
			campID := linkedinAdgroupIDData[id].CampaignID
			if U.IsNonEmptyKey(campID) {
				value.CampaignName = linkedinCampaignIDData[campID].Name
				linkedinAdgroupIDData[id] = value
			}
		}
	}

	data.AdwordsGCLIDData = adwordsGCLIDData
	data.AdwordsCampaignIDData = adwordsCampaignIDData
	data.AdwordsCampaignKeyData = model.GetKeyMapToData(model.AttributionKeyCampaign, adwordsCampaignIDData)

	data.AdwordsAdgroupIDData = adwordsAdgroupIDData
	data.AdwordsAdgroupKeyData = model.GetKeyMapToData(model.AttributionKeyAdgroup, adwordsAdgroupIDData)

	data.AdwordsKeywordIDData = adwordsKeywordIDData
	data.AdwordsKeywordKeyData = model.GetKeyMapToData(model.AttributionKeyKeyword, adwordsKeywordIDData)

	data.FacebookCampaignIDData = facebookCampaignIDData
	data.FacebookCampaignKeyData = model.GetKeyMapToData(model.AttributionKeyCampaign, facebookCampaignIDData)

	data.FacebookAdgroupIDData = facebookAdgroupIDData
	data.FacebookAdgroupKeyData = model.GetKeyMapToData(model.AttributionKeyAdgroup, facebookAdgroupIDData)

	data.LinkedinCampaignIDData = linkedinCampaignIDData
	data.LinkedinCampaignKeyData = model.GetKeyMapToData(model.AttributionKeyCampaign, linkedinCampaignIDData)

	data.LinkedinAdgroupIDData = linkedinAdgroupIDData
	data.LinkedinAdgroupKeyData = model.GetKeyMapToData(model.AttributionKeyAdgroup, linkedinAdgroupIDData)

	return data, err
}

// PullAdwordsMarketingData Pulls Adds channel data for Adwords.
func (store *MemSQL) PullAdwordsMarketingData(projectID uint64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, timeZone string) (map[string]model.MarketingData, error) {

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
		return nil, err
	}
	defer rows.Close()

	marketingDataIDMap := model.ProcessRow(rows, err, logCtx)
	return marketingDataIDMap, nil
}

// PullFacebookMarketingData Pulls Adds channel data for Facebook.
func (store *MemSQL) PullFacebookMarketingData(projectID uint64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, timeZone string) (map[string]model.MarketingData, error) {

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
		return nil, err
	}
	defer rows.Close()

	marketingDataIDMap := model.ProcessRow(rows, err, logCtx)
	return marketingDataIDMap, nil
}

// PullLinkedinMarketingData Pulls Adds channel data for Linkedin.
func (store *MemSQL) PullLinkedinMarketingData(projectID uint64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, timeZone string) (map[string]model.MarketingData, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})
	customerAccountIDs := strings.Split(customerAccountID, ",")
	performanceQuery := "SELECT campaign_group_id as campaignID, campaign_id as adgroupID, '$none' as keywordID, creative_id as adID, " +
		"JSON_EXTRACT_STRING(value, ?) AS key_id, JSON_EXTRACT_STRING(value, ?) AS key_name, JSON_EXTRACT_STRING(value, ?) AS extra_value1, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'costInLocalCurrency')) AS total_spend FROM linkedin_documents " +
		"where project_id = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, adID, key_id, key_name, extra_value1"

	rows, err := store.ExecQueryWithContext(performanceQuery, []interface{}{keyID, keyName, extraValue1, projectID, customerAccountIDs, reportType,
		U.GetDateAsStringZ(from, U.TimeZoneString(timeZone)), U.GetDateAsStringZ(to, U.TimeZoneString(timeZone))})
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, err
	}
	defer rows.Close()

	marketingDataIDMap := model.ProcessRow(rows, err, logCtx)
	return marketingDataIDMap, nil
}
