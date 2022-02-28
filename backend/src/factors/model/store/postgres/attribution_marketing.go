package postgres

import (
	"database/sql"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

func (pg *Postgres) FetchMarketingReports(projectID uint64, q model.AttributionQuery, projectSetting model.ProjectSetting) (*model.MarketingReports, error) {

	data := &model.MarketingReports{}
	var err error

	// Get adwords, facebook, linkedin reports.
	effectiveFrom := q.From
	effectiveTo := q.To

	adwordsCustomerID := ""
	if projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "" {
		adwordsCustomerID = ""
	} else {
		adwordsCustomerID = *projectSetting.IntAdwordsCustomerAccountId
	}
	enableBingAdsAttribution := C.GetConfig().EnableBingAdsAttribution

	var adwordsGCLIDData map[string]model.MarketingData
	var reportType int
	var adwordsCampaignIDData, adwordsAdgroupIDData, adwordsKeywordIDData map[string]model.MarketingData
	var adwordsCampaignAllRows, adwordsAdgroupAllRows, adwordsKeywordAllRows []model.MarketingData
	// Adwords.
	if adwordsCustomerID != "" && model.DoesAdwordsReportExist(q.AttributionKey) {

		reportType = model.AdwordsDocumentTypeAlias[model.CampaignPerformanceReport] // 5
		adwordsCampaignIDData, adwordsCampaignAllRows, err = pg.PullAdwordsMarketingData(projectID, effectiveFrom,
			effectiveTo, adwordsCustomerID, model.AdwordsCampaignID, model.AdwordsCampaignName, model.PropertyValueNone, reportType, model.ReportCampaign, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, v := range adwordsCampaignIDData {
			v.CampaignName = U.IfThenElse(U.IsNonEmptyKey(v.CampaignName), v.CampaignName, v.Name).(string)
			adwordsCampaignIDData[id] = v
		}

		for i, _ := range adwordsCampaignAllRows {
			adwordsCampaignAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(adwordsCampaignAllRows[i].CampaignName), adwordsCampaignAllRows[i].CampaignName, adwordsCampaignAllRows[i].Name).(string)
		}

		reportType = model.AdwordsDocumentTypeAlias[model.AdGroupPerformanceReport] // 10
		adwordsAdgroupIDData, adwordsAdgroupAllRows, err = pg.PullAdwordsMarketingData(projectID, effectiveFrom,
			effectiveTo, adwordsCustomerID, model.AdwordsAdgroupID, model.AdwordsAdgroupName, model.PropertyValueNone, reportType, model.ReportAdGroup, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, value := range adwordsAdgroupIDData {
			value.AdgroupName = U.IfThenElse(U.IsNonEmptyKey(value.AdgroupName), value.AdgroupName, value.Name).(string)
			campID := value.CampaignID
			if U.IsNonEmptyKey(campID) {
				value.CampaignName = U.IfThenElse(U.IsNonEmptyKey(value.CampaignName), value.CampaignName, adwordsCampaignIDData[campID].Name).(string)
				adwordsAdgroupIDData[id] = value
			}
		}
		for i, _ := range adwordsAdgroupAllRows {
			adwordsAdgroupAllRows[i].AdgroupName = U.IfThenElse(U.IsNonEmptyKey(adwordsAdgroupAllRows[i].AdgroupName), adwordsAdgroupAllRows[i].AdgroupName, adwordsAdgroupAllRows[i].Name).(string)
			campID := adwordsAdgroupAllRows[i].CampaignID
			if U.IsNonEmptyKey(campID) {
				adwordsAdgroupAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(adwordsAdgroupAllRows[i].CampaignName), adwordsAdgroupAllRows[i].CampaignName, adwordsCampaignIDData[campID].Name).(string)
			}
		}

		reportType = model.AdwordsDocumentTypeAlias[model.KeywordPerformanceReport] // 8
		adwordsKeywordIDData, adwordsKeywordAllRows, err = pg.PullAdwordsMarketingData(projectID, effectiveFrom,
			effectiveTo, adwordsCustomerID, model.AdwordsKeywordID, model.AdwordsKeywordName, model.AdwordsKeywordMatchType, reportType, model.ReportKeyword, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, value := range adwordsKeywordIDData {
			value.KeywordName = U.IfThenElse(U.IsNonEmptyKey(value.KeywordName), value.KeywordName, value.Name).(string)
			campID := value.CampaignID
			if U.IsNonEmptyKey(campID) {
				value.CampaignName = U.IfThenElse(U.IsNonEmptyKey(value.CampaignName), value.CampaignName, adwordsCampaignIDData[campID].Name).(string)
				adwordsKeywordIDData[id] = value
			}
		}

		for i, _ := range adwordsKeywordAllRows {
			adwordsKeywordAllRows[i].KeywordName = U.IfThenElse(U.IsNonEmptyKey(adwordsKeywordAllRows[i].KeywordName), adwordsKeywordAllRows[i].KeywordName, adwordsKeywordAllRows[i].Name).(string)
			campID := adwordsKeywordAllRows[i].CampaignID
			if U.IsNonEmptyKey(campID) {
				adwordsKeywordAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(adwordsKeywordAllRows[i].CampaignName), adwordsKeywordAllRows[i].CampaignName, adwordsCampaignIDData[campID].Name).(string)
			}
		}
		for id, value := range adwordsKeywordIDData {
			adgroupID := value.AdgroupID
			if U.IsNonEmptyKey(adgroupID) {
				value.AdgroupName = U.IfThenElse(U.IsNonEmptyKey(value.AdgroupName), value.AdgroupName, adwordsAdgroupIDData[adgroupID].Name).(string)
				adwordsKeywordIDData[id] = value
			}
		}
		for i, _ := range adwordsKeywordAllRows {
			adgroupID := adwordsKeywordAllRows[i].AdgroupID
			if U.IsNonEmptyKey(adgroupID) {
				adwordsKeywordAllRows[i].AdgroupName = U.IfThenElse(U.IsNonEmptyKey(adwordsKeywordAllRows[i].AdgroupName), adwordsKeywordAllRows[i].AdgroupName, adwordsAdgroupIDData[adgroupID].Name).(string)
			}
		}

		// Adding 2 days in the effective query range for GCLID report to capture GCLID leakage
		adwordsGCLIDData, err = pg.PullGCLIDReport(projectID, effectiveFrom-(2*model.SecsInADay), effectiveTo+(2*model.SecsInADay), adwordsCustomerID,
			adwordsCampaignIDData, adwordsAdgroupIDData, adwordsKeywordIDData, q.Timezone)
		if err != nil {
			return data, err
		}
	}

	// Facebook.
	var facebookCampaignIDData, facebookAdgroupIDData map[string]model.MarketingData
	var facebookCampaignAllRows, facebookAdgroupAllRows []model.MarketingData
	if projectSetting.IntFacebookAdAccount != "" && model.DoesFBReportExist(q.AttributionKey) {
		facebookCustomerID := projectSetting.IntFacebookAdAccount

		reportType = facebookDocumentTypeAlias["campaign_insights"] // 5
		facebookCampaignIDData, facebookCampaignAllRows, err = pg.PullFacebookMarketingData(projectID, effectiveFrom,
			effectiveTo, facebookCustomerID, model.FacebookCampaignID, model.FacebookCampaignName, model.PropertyValueNone, reportType, model.ReportCampaign, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, v := range facebookCampaignIDData {
			v.CampaignName = U.IfThenElse(U.IsNonEmptyKey(v.CampaignName), v.CampaignName, v.Name).(string)
			facebookCampaignIDData[id] = v
		}
		for i, _ := range facebookCampaignAllRows {
			facebookCampaignAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(facebookCampaignAllRows[i].CampaignName), facebookCampaignAllRows[i].CampaignName, facebookCampaignAllRows[i].Name).(string)
		}

		reportType = facebookDocumentTypeAlias["ad_set_insights"] // 6
		facebookAdgroupIDData, facebookAdgroupAllRows, err = pg.PullFacebookMarketingData(projectID, effectiveFrom,
			effectiveTo, facebookCustomerID, model.FacebookAdgroupID, model.FacebookAdgroupName, model.PropertyValueNone, reportType, model.ReportAdGroup, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, value := range facebookAdgroupIDData {
			value.AdgroupName = U.IfThenElse(U.IsNonEmptyKey(value.AdgroupName), value.AdgroupName, value.Name).(string)
			campID := value.CampaignID
			if U.IsNonEmptyKey(campID) {
				value.CampaignName = U.IfThenElse(U.IsNonEmptyKey(value.CampaignName), value.CampaignName, facebookCampaignIDData[campID].Name).(string)
				facebookAdgroupIDData[id] = value
			}
		}
		for i, _ := range facebookAdgroupAllRows {
			facebookAdgroupAllRows[i].AdgroupName = U.IfThenElse(U.IsNonEmptyKey(facebookAdgroupAllRows[i].AdgroupName), facebookAdgroupAllRows[i].AdgroupName, facebookAdgroupAllRows[i].Name).(string)
			campID := facebookAdgroupAllRows[i].CampaignID
			if U.IsNonEmptyKey(campID) {
				facebookAdgroupAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(facebookAdgroupAllRows[i].CampaignName), facebookAdgroupAllRows[i].CampaignName, facebookCampaignIDData[campID].Name).(string)
			}
		}
	}

	// Linkedin.
	var linkedinCampaignIDData, linkedinAdgroupIDData map[string]model.MarketingData
	var linkedinCampaignAllRows, linkedinAdgroupAllRows []model.MarketingData
	if projectSetting.IntLinkedinAdAccount != "" && model.DoesLinkedinReportExist(q.AttributionKey) {
		linkedinCustomerID := projectSetting.IntLinkedinAdAccount

		reportType = linkedinDocumentTypeAlias["campaign_group_insights"] // 5
		linkedinCampaignIDData, linkedinCampaignAllRows, err = pg.PullLinkedinMarketingData(projectID, effectiveFrom,
			effectiveTo, linkedinCustomerID, model.LinkedinCampaignID, model.LinkedinCampaignName, model.PropertyValueNone, reportType, model.ReportCampaign, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, v := range linkedinCampaignIDData {
			v.CampaignName = U.IfThenElse(U.IsNonEmptyKey(v.CampaignName), v.CampaignName, v.Name).(string)
			linkedinCampaignIDData[id] = v
		}
		for i, _ := range linkedinCampaignAllRows {
			linkedinCampaignAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(linkedinCampaignAllRows[i].CampaignName), linkedinCampaignAllRows[i].CampaignName, linkedinCampaignAllRows[i].Name).(string)
		}

		reportType = linkedinDocumentTypeAlias["campaign_insights"] // 6
		linkedinAdgroupIDData, linkedinAdgroupAllRows, err = pg.PullLinkedinMarketingData(projectID, effectiveFrom,
			effectiveTo, linkedinCustomerID, model.LinkedinAdgroupID, model.LinkedinAdgroupName, model.PropertyValueNone, reportType, model.ReportAdGroup, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, value := range linkedinAdgroupIDData {
			value.AdgroupName = U.IfThenElse(U.IsNonEmptyKey(value.AdgroupName), value.AdgroupName, value.Name).(string)
			campID := value.CampaignID
			if U.IsNonEmptyKey(campID) {
				value.CampaignName = U.IfThenElse(U.IsNonEmptyKey(value.CampaignName), value.CampaignName, linkedinCampaignIDData[campID].Name).(string)
				linkedinAdgroupIDData[id] = value
			}
		}
		for i, _ := range linkedinAdgroupAllRows {
			linkedinAdgroupAllRows[i].AdgroupName = U.IfThenElse(U.IsNonEmptyKey(linkedinAdgroupAllRows[i].AdgroupName), linkedinAdgroupAllRows[i].AdgroupName, linkedinAdgroupAllRows[i].Name).(string)
			campID := linkedinAdgroupAllRows[i].CampaignID
			if U.IsNonEmptyKey(campID) {
				linkedinAdgroupAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(linkedinAdgroupAllRows[i].CampaignName), linkedinAdgroupAllRows[i].CampaignName, linkedinCampaignIDData[campID].Name).(string)
			}
		}
	}

	// Bingads
	var bingadsCampaignIDData, bingadsAdgroupIDData, bingadsKeywordIDData map[string]model.MarketingData
	var bingadsCampaignAllRows, bingadsAdgroupAllRows, bingadsKeywordAllRows []model.MarketingData
	if enableBingAdsAttribution {
		isBingAdsIntegrationDone := pg.IsBingIntegrationAvailable(projectID)
		if isBingAdsIntegrationDone && model.DoesBingAdsReportExist(q.AttributionKey) {
			bingAdsAccountID, _ := pg.getBingAdsAccountId(projectID)

			reportType = model.BingadsDocumentTypeAlias[model.CampaignPerformanceReport] // 4
			bingadsCampaignIDData, bingadsCampaignAllRows, err = pg.PullBingAdsMarketingData(projectID, effectiveFrom,
				effectiveTo, bingAdsAccountID, model.BingadsCampaignID, model.BingadsCampaignName, model.PropertyValueNone, reportType, model.ReportCampaign, q.Timezone)
			if err != nil {
				return data, err
			}
			for id, v := range bingadsCampaignIDData {
				v.CampaignName = U.IfThenElse(U.IsNonEmptyKey(v.CampaignName), v.CampaignName, v.Name).(string)
				bingadsCampaignIDData[id] = v
			}
			for i, _ := range bingadsCampaignAllRows {
				bingadsCampaignAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(bingadsCampaignAllRows[i].CampaignName), bingadsCampaignAllRows[i].CampaignName, bingadsCampaignAllRows[i].Name).(string)
			}

			reportType = model.BingadsDocumentTypeAlias[model.AdGroupPerformanceReport] // 5
			bingadsAdgroupIDData, bingadsAdgroupAllRows, err = pg.PullBingAdsMarketingData(projectID, effectiveFrom,
				effectiveTo, bingAdsAccountID, model.BingadsAdgroupID, model.BingadsAdgroupName, model.PropertyValueNone, reportType, model.ReportAdGroup, q.Timezone)
			if err != nil {
				return data, err
			}
			for id, value := range bingadsAdgroupIDData {
				value.AdgroupName = U.IfThenElse(U.IsNonEmptyKey(value.AdgroupName), value.AdgroupName, value.Name).(string)
				campID := value.CampaignID
				if U.IsNonEmptyKey(campID) {
					value.CampaignName = U.IfThenElse(U.IsNonEmptyKey(value.CampaignName), value.CampaignName, bingadsCampaignIDData[campID].Name).(string)
					bingadsAdgroupIDData[id] = value
				}
			}
			for i, _ := range bingadsAdgroupAllRows {
				bingadsAdgroupAllRows[i].AdgroupName = U.IfThenElse(U.IsNonEmptyKey(bingadsAdgroupAllRows[i].AdgroupName), bingadsAdgroupAllRows[i].AdgroupName, bingadsAdgroupAllRows[i].Name).(string)
				campID := adwordsAdgroupAllRows[i].CampaignID
				if U.IsNonEmptyKey(campID) {
					bingadsAdgroupAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(bingadsAdgroupAllRows[i].CampaignName), bingadsAdgroupAllRows[i].CampaignName, bingadsCampaignIDData[campID].Name).(string)
				}
			}

			reportType = model.BingadsDocumentTypeAlias[model.KeywordPerformanceReport] // 6
			bingadsKeywordIDData, bingadsKeywordAllRows, err = pg.PullBingAdsMarketingData(projectID, effectiveFrom,
				effectiveTo, bingAdsAccountID, model.BingadsKeywordID, model.BingadsKeywordName, model.PropertyValueNone, reportType, model.ReportKeyword, q.Timezone)
			if err != nil {
				return data, err
			}
			for id, value := range bingadsKeywordIDData {
				value.KeywordName = U.IfThenElse(U.IsNonEmptyKey(value.KeywordName), value.KeywordName, value.Name).(string)
				campID := value.CampaignID
				if U.IsNonEmptyKey(campID) {
					value.CampaignName = U.IfThenElse(U.IsNonEmptyKey(value.CampaignName), value.CampaignName, bingadsCampaignIDData[campID].Name).(string)
					bingadsKeywordIDData[id] = value
				}
			}

			for i, _ := range bingadsKeywordAllRows {
				bingadsKeywordAllRows[i].KeywordName = U.IfThenElse(U.IsNonEmptyKey(bingadsKeywordAllRows[i].KeywordName), bingadsKeywordAllRows[i].KeywordName, bingadsKeywordAllRows[i].Name).(string)
				campID := bingadsKeywordAllRows[i].CampaignID
				if U.IsNonEmptyKey(campID) {
					bingadsKeywordAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(bingadsKeywordAllRows[i].CampaignName), bingadsKeywordAllRows[i].CampaignName, bingadsCampaignIDData[campID].Name).(string)
				}
			}
			for id, value := range bingadsKeywordIDData {
				adgroupID := value.AdgroupID
				if U.IsNonEmptyKey(adgroupID) {
					value.AdgroupName = U.IfThenElse(U.IsNonEmptyKey(value.AdgroupName), value.AdgroupName, bingadsAdgroupIDData[adgroupID].Name).(string)
					bingadsKeywordIDData[id] = value
				}
			}
			for i, _ := range bingadsKeywordAllRows {
				adgroupID := bingadsKeywordAllRows[i].AdgroupID
				if U.IsNonEmptyKey(adgroupID) {
					bingadsKeywordAllRows[i].AdgroupName = U.IfThenElse(U.IsNonEmptyKey(bingadsKeywordAllRows[i].AdgroupName), bingadsKeywordAllRows[i].AdgroupName, bingadsAdgroupIDData[adgroupID].Name).(string)
				}
			}
		}
	}
	data.AdwordsGCLIDData = adwordsGCLIDData
	data.AdwordsCampaignIDData = adwordsCampaignIDData
	data.AdwordsCampaignKeyData = model.GetKeyMapToData(model.AttributionKeyCampaign, adwordsCampaignAllRows)

	data.AdwordsAdgroupIDData = adwordsAdgroupIDData
	data.AdwordsAdgroupKeyData = model.GetKeyMapToData(model.AttributionKeyAdgroup, adwordsAdgroupAllRows)

	data.AdwordsKeywordIDData = adwordsKeywordIDData
	data.AdwordsKeywordKeyData = model.GetKeyMapToData(model.AttributionKeyKeyword, adwordsKeywordAllRows)

	data.FacebookCampaignIDData = facebookCampaignIDData
	data.FacebookCampaignKeyData = model.GetKeyMapToData(model.AttributionKeyCampaign, facebookCampaignAllRows)

	data.FacebookAdgroupIDData = facebookAdgroupIDData
	data.FacebookAdgroupKeyData = model.GetKeyMapToData(model.AttributionKeyAdgroup, facebookAdgroupAllRows)

	data.LinkedinCampaignIDData = linkedinCampaignIDData
	data.LinkedinCampaignKeyData = model.GetKeyMapToData(model.AttributionKeyCampaign, linkedinCampaignAllRows)

	data.LinkedinAdgroupIDData = linkedinAdgroupIDData
	data.LinkedinAdgroupKeyData = model.GetKeyMapToData(model.AttributionKeyAdgroup, linkedinAdgroupAllRows)

	return data, err
}

func (pg *Postgres) getBingAdsAccountId(projectID uint64) (string, error) {
	ftMapping, _ := pg.GetActiveFiveTranMapping(projectID, model.BingAdsIntegration)
	return ftMapping.Accounts, nil
}

// PullAdwordsMarketingData Pulls Adds channel data for Adwords.
func (pg *Postgres) PullAdwordsMarketingData(projectID uint64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, reportName string, timeZone string) (map[string]model.MarketingData, []model.MarketingData, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})
	customerAccountIDs := strings.Split(customerAccountID, ",")
	performanceQuery := "SELECT campaign_id as campaignID, ad_group_id as adgroupID, keyword_id as keywordID, ad_id as adID, " +
		"value->>? AS key_id,  value->>? AS key_name, value->>? AS extra_value1," +
		"SUM((value->>'impressions')::float) AS impressions, SUM((value->>'clicks')::float) AS clicks, " +
		"SUM((value->>'cost')::float)/1000000 AS total_cost FROM adwords_documents " +
		"where project_id = ? AND customer_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, adID, key_id, key_name, extra_value1"

	params := []interface{}{keyID, keyName, extraValue1, projectID, customerAccountIDs, reportType,
		U.GetDateAsStringIn(from, U.TimeZoneString(timeZone)), U.GetDateAsStringIn(to, U.TimeZoneString(timeZone))}
	rows, tx, err := pg.ExecQueryWithContext(performanceQuery, params)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	marketingDataIDMap, allRows := model.ProcessRow(rows, reportName, logCtx, model.ChannelAdwords)
	return marketingDataIDMap, allRows, nil
}

// PullFacebookMarketingData Pulls Adds channel data for Facebook.
func (pg *Postgres) PullFacebookMarketingData(projectID uint64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, reportName string, timeZone string) (map[string]model.MarketingData, []model.MarketingData, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})
	customerAccountIDs := strings.Split(customerAccountID, ",")
	performanceQuery := "SELECT campaign_id as campaignID, ad_set_id as adgroupID, '$none' as keywordID, ad_id as adID, " +
		"value->>? AS key_id,  value->>? AS key_name, value->>? AS extra_value1," +
		"SUM((value->>'impressions')::float) AS impressions, SUM((value->>'clicks')::float) AS clicks, " +
		"SUM((value->>'spend')::float) AS total_cost FROM facebook_documents " +
		"where project_id = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, adID, key_id, key_name, extra_value1"

	params := []interface{}{keyID, keyName, extraValue1, projectID, customerAccountIDs, reportType,
		U.GetDateAsStringIn(from, U.TimeZoneString(timeZone)), U.GetDateAsStringIn(to, U.TimeZoneString(timeZone))}
	rows, tx, err := pg.ExecQueryWithContext(performanceQuery, params)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	marketingDataIDMap, allRows := model.ProcessRow(rows, reportName, logCtx, model.ChannelFacebook)
	return marketingDataIDMap, allRows, nil
}

// PullLinkedinMarketingData Pulls Adds channel data for Linkedin.
func (pg *Postgres) PullLinkedinMarketingData(projectID uint64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, reportName string, timeZone string) (map[string]model.MarketingData, []model.MarketingData, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})
	customerAccountIDs := strings.Split(customerAccountID, ",")
	performanceQuery := "SELECT campaign_group_id as campaignID, campaign_id as adgroupID, '$none' as keywordID, creative_id as adID, " +
		"value->>? AS key_id,  value->>? AS key_name, value->>? AS extra_value1," +
		"SUM((value->>'impressions')::float) AS impressions, SUM((value->>'clicks')::float) AS clicks, " +
		"SUM((value->>'costInLocalCurrency')::float) AS total_cost FROM linkedin_documents " +
		"where project_id = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, adID, key_id, key_name, extra_value1"

	params := []interface{}{keyID, keyName, extraValue1, projectID, customerAccountIDs, reportType,
		U.GetDateAsStringIn(from, U.TimeZoneString(timeZone)), U.GetDateAsStringIn(to, U.TimeZoneString(timeZone))}
	rows, tx, err := pg.ExecQueryWithContext(performanceQuery, params)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	marketingDataIDMap, allRows := model.ProcessRow(rows, reportName, logCtx, model.ChannelLinkedin)
	return marketingDataIDMap, allRows, nil
}
func (pg *Postgres) PullBingAdsMarketingData(projectID uint64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, reportName string, timeZone string) (map[string]model.MarketingData, []model.MarketingData, error) {
	logFields := log.Fields{
		"project_id":   projectID,
		"from":         from,
		"to":           to,
		"account_id":   customerAccountID,
		"key_id":       keyID,
		"key_name":     keyName,
		"extra_value1": extraValue1,
		"report_name":  reportName,
		"time_zone":    timeZone,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	customerAccountIDs := strings.Split(customerAccountID, ",")
	performanceQuery := "SELECT JSON_EXTRACT_STRING(value, 'campaign_id')  as campaignID, JSON_EXTRACT_STRING(value, 'ad_group_id') as adgroupID, JSON_EXTRACT_STRING(value, 'keyword_id') as keywordID, " +
		"JSON_EXTRACT_STRING(value, ?) AS key_id, JSON_EXTRACT_STRING(value, ?) AS key_name, JSON_EXTRACT_STRING(value, ?) AS extra_value1, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'spend')) AS total_spend FROM integration_documents " +
		"where project_id = ? AND source = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, key_id, key_name, extra_value1"

	params := []interface{}{keyID, keyName, extraValue1, projectID, model.BingAdsIntegration, customerAccountIDs, reportType,
		U.GetDateAsStringIn(from, U.TimeZoneString(timeZone)), U.GetDateAsStringIn(to, U.TimeZoneString(timeZone))}
	rows, tx, err := pg.ExecQueryWithContext(performanceQuery, params)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	marketingDataIDMap, allRows := model.ProcessRow(rows, reportName, logCtx, model.BingAdsIntegration)
	return marketingDataIDMap, allRows, nil
}
func (pg *Postgres) PullCustomDimensionData(projectID uint64, attributionKey string, marketingReport *model.MarketingReports) error {

	// Custom Dimensions are support only for Campaign and Adgroup currently
	if attributionKey != model.AttributionKeyCampaign && attributionKey != model.AttributionKeyAdgroup {
		return nil
	}

	var err error
	switch attributionKey {
	case model.AttributionKeyCampaign:

		marketingReport.AdwordsCampaignDimensions, err = pg.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, model.ChannelAdwords, 1, attributionKey)
		if err != nil {
			return err
		}
		marketingReport.FacebookCampaignDimensions, err = pg.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, model.ChannelFacebook, 1, attributionKey)
		if err != nil {
			return err
		}
		marketingReport.LinkedinCampaignDimensions, err = pg.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, model.ChannelLinkedin, 1, attributionKey)
		if err != nil {
			return err
		}
		marketingReport.BingadsCampaignDimensions, err = pg.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, model.ChannelBingads, 1, attributionKey)
		if err != nil {
			return err
		}
	case model.FieldAdgroupName:
		marketingReport.AdwordsAdgroupDimensions, err = pg.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, model.ChannelAdwords, 2, attributionKey)
		if err != nil {
			return err
		}
		marketingReport.FacebookAdgroupDimensions, err = pg.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, model.ChannelFacebook, 2, attributionKey)
		if err != nil {
			return err
		}
		marketingReport.LinkedinAdgroupDimensions, err = pg.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, model.ChannelLinkedin, 2, attributionKey)
		if err != nil {
			return err
		}
		marketingReport.BingadsAdgroupDimensions, err = pg.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, model.ChannelBingads, 2, attributionKey)
		if err != nil {
			return err
		}
	}
	return nil
}

// PullSmartProperties Pulls Smart Properties
func (pg *Postgres) PullSmartProperties(projectID uint64, campaignIDPlaceHolder string, campaignNamePlaceHolder string, adgroupIDPlaceHolder string, adgroupNamePlaceHolder string, sourceChannelPlaceHolder string, objectType int, attributionKey string) (map[string]model.MarketingData, error) {

	// GetEventsWithoutPropertiesAndWithPropertiesByNameForYourStory
	logCtx := log.WithFields(log.Fields{"ProjectId": projectID, "UnitType": objectType, "Source": sourceChannelPlaceHolder})
	stmt := "SELECT object_property->>? AS campaignID,  object_property->>? AS campaignName, " +
		"object_property->>? AS adgroupID,  object_property->>? AS adgroupName, " +
		"properties FROM smart_properties " +
		"where project_id = ? AND source = ? AND object_type = ?"

	params := []interface{}{campaignIDPlaceHolder, campaignNamePlaceHolder, adgroupIDPlaceHolder, adgroupNamePlaceHolder, projectID, sourceChannelPlaceHolder, objectType}
	rows, tx, err := pg.ExecQueryWithContext(stmt, params)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, err
	}
	defer U.CloseReadQuery(rows, tx)
	dataKeyDimensions := make(map[string]model.MarketingData)
	for rows.Next() {
		var campaignIDNull sql.NullString
		var campaignNameNull sql.NullString
		var adgroupIDNull sql.NullString
		var adgroupNameNull sql.NullString
		var properties postgres.Jsonb

		if err := rows.Scan(&campaignIDNull, &campaignNameNull, &adgroupIDNull, &adgroupNameNull, &properties); err != nil {
			logCtx.WithError(err).Error("Bad row. Ignoring row and continuing")
			continue
		}
		if !campaignIDNull.Valid && !adgroupIDNull.Valid {
			continue
		}
		_campaignID := model.IfValidGetValElseNone(campaignIDNull)
		_campaignName := model.IfValidGetValElseNone(campaignNameNull)
		_adgroupID := model.IfValidGetValElseNone(adgroupIDNull)
		_adgroupName := model.IfValidGetValElseNone(adgroupNameNull)

		propertiesMap, err := U.DecodePostgresJsonb(&properties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to decode smart properties. Ignoring row and continuing")
			continue
		}
		marketData := model.MarketingData{
			Name:    _campaignName,
			ID:      _campaignID,
			Channel: sourceChannelPlaceHolder,
		}
		key := model.GetKeyForCustomDimensions(_campaignID, _campaignName, _adgroupID, _adgroupName, attributionKey)
		if key == "" {
			continue
		}
		if objectType == model.SmartPropertyRulesTypeAliasToType["ad_group"] { // 1: "campaign", 2: "ad_group"
			marketData.Name = _adgroupName
			marketData.ID = _adgroupID
		}
		// added custom dimensions
		(*propertiesMap)["campaign_id"] = _campaignID
		(*propertiesMap)["campaign_name"] = _campaignName
		(*propertiesMap)["adgroup_id"] = _adgroupID
		(*propertiesMap)["adgroup_name"] = _adgroupName
		(*propertiesMap)["channel_name"] = sourceChannelPlaceHolder

		dataKeyDimensions[key] = model.MarketingData{
			Key:              key,
			CampaignID:       _campaignID,
			CampaignName:     _campaignName,
			AdgroupID:        _adgroupID,
			AdgroupName:      _adgroupName,
			CustomDimensions: *propertiesMap,
		}

	}
	log.WithFields(log.Fields{"CustomDebug": "True", "ProjectId": projectID, "UnitType": objectType, "Source": sourceChannelPlaceHolder, "DataKeyDimensions": dataKeyDimensions}).Info("Pull Smart Properties")
	return dataKeyDimensions, nil
}
