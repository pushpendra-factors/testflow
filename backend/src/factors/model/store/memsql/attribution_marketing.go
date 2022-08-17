package memsql

import (
	"database/sql"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) FetchMarketingReports(projectID int64, q model.AttributionQuery, projectSetting model.ProjectSetting) (*model.MarketingReports, error) {
	logFields := log.Fields{
		"project_id":      projectID,
		"q":               q,
		"project_setting": projectSetting,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	enableBingAdsAttribution := C.GetConfig().EnableBingAdsAttribution
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
	var adwordsGCLIDData map[string]model.MarketingData
	var reportType int
	var adwordsCampaignIDData, adwordsAdgroupIDData, adwordsKeywordIDData map[string]model.MarketingData
	var adwordsCampaignAllRows, adwordsAdgroupAllRows, adwordsKeywordAllRows []model.MarketingData
	// Adwords.
	if adwordsCustomerID != "" && model.DoesAdwordsReportExist(q.AttributionKey) {

		reportType = model.AdwordsDocumentTypeAlias[model.CampaignPerformanceReport] // 5
		adwordsCampaignIDData, adwordsCampaignAllRows, err = store.PullAdwordsMarketingData(projectID, effectiveFrom,
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
		adwordsAdgroupIDData, adwordsAdgroupAllRows, err = store.PullAdwordsMarketingData(projectID, effectiveFrom,
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
		adwordsKeywordIDData, adwordsKeywordAllRows, err = store.PullAdwordsMarketingData(projectID, effectiveFrom,
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
		adwordsGCLIDData, err = store.PullGCLIDReport(projectID, effectiveFrom-(2*model.SecsInADay), effectiveTo+(2*model.SecsInADay), adwordsCustomerID,
			adwordsCampaignIDData, adwordsAdgroupIDData, adwordsKeywordIDData, q.Timezone)
		if err != nil {
			return data, err
		}

		log.WithFields(log.Fields{"size of adwordsGCLIDData map": len(adwordsGCLIDData)}).Info("Attribution keyword razorpay debug")
	}

	// Facebook.
	var facebookCampaignIDData, facebookAdgroupIDData map[string]model.MarketingData
	var facebookCampaignAllRows, facebookAdgroupAllRows []model.MarketingData
	if projectSetting.IntFacebookAdAccount != "" && model.DoesFBReportExist(q.AttributionKey) {
		facebookCustomerID := projectSetting.IntFacebookAdAccount

		reportType = FacebookDocumentTypeAlias["campaign_insights"] // 5
		facebookCampaignIDData, facebookCampaignAllRows, err = store.PullFacebookMarketingData(projectID, effectiveFrom,
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

		reportType = FacebookDocumentTypeAlias["ad_set_insights"] // 6
		facebookAdgroupIDData, facebookAdgroupAllRows, err = store.PullFacebookMarketingData(projectID, effectiveFrom,
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

		reportType = LinkedinDocumentTypeAlias["campaign_group_insights"] // 5
		linkedinCampaignIDData, linkedinCampaignAllRows, err = store.PullLinkedinMarketingData(projectID, effectiveFrom,
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

		reportType = LinkedinDocumentTypeAlias["campaign_insights"] // 6
		linkedinAdgroupIDData, linkedinAdgroupAllRows, err = store.PullLinkedinMarketingData(projectID, effectiveFrom,
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
		isBingAdsIntegrationDone := store.IsBingIntegrationAvailable(projectID)
		if isBingAdsIntegrationDone && model.DoesBingAdsReportExist(q.AttributionKey) {
			bingAdsAccountID, _ := store.getBingAdsAccountId(projectID)

			reportType = model.BingadsDocumentTypeAlias[model.CampaignPerformanceReport] // 4
			bingadsCampaignIDData, bingadsCampaignAllRows, err = store.PullBingAdsMarketingData(projectID, effectiveFrom,
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
			bingadsAdgroupIDData, bingadsAdgroupAllRows, err = store.PullBingAdsMarketingData(projectID, effectiveFrom,
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
				campID := bingadsAdgroupAllRows[i].CampaignID
				if U.IsNonEmptyKey(campID) {
					bingadsAdgroupAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(bingadsAdgroupAllRows[i].CampaignName), bingadsAdgroupAllRows[i].CampaignName, bingadsCampaignIDData[campID].Name).(string)
				}
			}

			reportType = model.BingadsDocumentTypeAlias[model.KeywordPerformanceReport] // 6
			bingadsKeywordIDData, bingadsKeywordAllRows, err = store.PullBingAdsMarketingData(projectID, effectiveFrom,
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

	// CustomAds
	var customadsCampaignIDData, customadsAdgroupIDData, customadsKeywordIDData map[string]model.MarketingData
	var customadsCampaignAllRows, customadsAdgroupAllRows, customadsKeywordAllRows []model.MarketingData
	isCustomAdsIntegrationDone := store.IsCustomAdsAvailable(projectID)
	sources, _ := store.GetCustomAdsSourcesByProject(projectID)
	if isCustomAdsIntegrationDone && model.DoesCustomAdsReportExist(q.AttributionKey) {
		customAdsAccountID, _ := store.GetCustomAdsAccountsByProject(projectID, sources)

		reportType = model.CustomadsDocumentTypeAlias[model.CampaignPerformanceReport] // 4
		customadsCampaignIDData, customadsCampaignAllRows, err = store.PullCustomAdsMarketingData(projectID, effectiveFrom,
			effectiveTo, customAdsAccountID, model.CustomadsCampaignID, model.CustomadsCampaignName, model.PropertyValueNone, reportType, model.ReportCampaign, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, v := range customadsCampaignIDData {
			v.CampaignName = U.IfThenElse(U.IsNonEmptyKey(v.CampaignName), v.CampaignName, v.Name).(string)
			customadsCampaignIDData[id] = v
		}
		for i, _ := range customadsCampaignAllRows {
			customadsCampaignAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(customadsCampaignAllRows[i].CampaignName), customadsCampaignAllRows[i].CampaignName, customadsCampaignAllRows[i].Name).(string)
		}

		reportType = model.CustomadsDocumentTypeAlias[model.AdGroupPerformanceReport] // 5
		customadsAdgroupIDData, customadsAdgroupAllRows, err = store.PullCustomAdsMarketingData(projectID, effectiveFrom,
			effectiveTo, customAdsAccountID, model.CustomadsAdgroupID, model.CustomadsAdgroupName, model.PropertyValueNone, reportType, model.ReportAdGroup, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, value := range customadsAdgroupIDData {
			value.AdgroupName = U.IfThenElse(U.IsNonEmptyKey(value.AdgroupName), value.AdgroupName, value.Name).(string)
			campID := value.CampaignID
			if U.IsNonEmptyKey(campID) {
				value.CampaignName = U.IfThenElse(U.IsNonEmptyKey(value.CampaignName), value.CampaignName, customadsCampaignIDData[campID].Name).(string)
				customadsAdgroupIDData[id] = value
			}
		}
		for i, _ := range customadsAdgroupAllRows {
			customadsAdgroupAllRows[i].AdgroupName = U.IfThenElse(U.IsNonEmptyKey(customadsAdgroupAllRows[i].AdgroupName), customadsAdgroupAllRows[i].AdgroupName, customadsAdgroupAllRows[i].Name).(string)
			campID := customadsAdgroupAllRows[i].CampaignID
			if U.IsNonEmptyKey(campID) {
				customadsAdgroupAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(customadsAdgroupAllRows[i].CampaignName), customadsAdgroupAllRows[i].CampaignName, customadsCampaignIDData[campID].Name).(string)
			}
		}

		reportType = model.CustomadsDocumentTypeAlias[model.KeywordPerformanceReport] // 6
		customadsKeywordIDData, customadsKeywordAllRows, err = store.PullCustomAdsMarketingData(projectID, effectiveFrom,
			effectiveTo, customAdsAccountID, model.CustomadsKeywordID, model.CustomadsKeywordName, model.PropertyValueNone, reportType, model.ReportKeyword, q.Timezone)
		if err != nil {
			return data, err
		}
		for id, value := range customadsKeywordIDData {
			value.KeywordName = U.IfThenElse(U.IsNonEmptyKey(value.KeywordName), value.KeywordName, value.Name).(string)
			campID := value.CampaignID
			if U.IsNonEmptyKey(campID) {
				value.CampaignName = U.IfThenElse(U.IsNonEmptyKey(value.CampaignName), value.CampaignName, customadsCampaignIDData[campID].Name).(string)
				customadsKeywordIDData[id] = value
			}
		}

		for i, _ := range customadsKeywordAllRows {
			customadsKeywordAllRows[i].KeywordName = U.IfThenElse(U.IsNonEmptyKey(customadsKeywordAllRows[i].KeywordName), customadsKeywordAllRows[i].KeywordName, customadsKeywordAllRows[i].Name).(string)
			campID := customadsKeywordAllRows[i].CampaignID
			if U.IsNonEmptyKey(campID) {
				customadsKeywordAllRows[i].CampaignName = U.IfThenElse(U.IsNonEmptyKey(customadsKeywordAllRows[i].CampaignName), customadsKeywordAllRows[i].CampaignName, customadsCampaignIDData[campID].Name).(string)
			}
		}
		for id, value := range customadsKeywordIDData {
			adgroupID := value.AdgroupID
			if U.IsNonEmptyKey(adgroupID) {
				value.AdgroupName = U.IfThenElse(U.IsNonEmptyKey(value.AdgroupName), value.AdgroupName, customadsAdgroupIDData[adgroupID].Name).(string)
				customadsKeywordIDData[id] = value
			}
		}
		for i, _ := range customadsKeywordAllRows {
			adgroupID := customadsKeywordAllRows[i].AdgroupID
			if U.IsNonEmptyKey(adgroupID) {
				customadsKeywordAllRows[i].AdgroupName = U.IfThenElse(U.IsNonEmptyKey(customadsKeywordAllRows[i].AdgroupName), customadsKeywordAllRows[i].AdgroupName, customadsAdgroupIDData[adgroupID].Name).(string)
			}
		}
	}

	data.AdwordsGCLIDData = adwordsGCLIDData
	data.AdwordsCampaignIDData = adwordsCampaignIDData
	data.AdwordsCampaignKeyData = model.GetKeyMapToData(model.AttributionKeyCampaign, adwordsCampaignAllRows, data.AdwordsCampaignIDData)

	data.AdwordsAdgroupIDData = adwordsAdgroupIDData
	data.AdwordsAdgroupKeyData = model.GetKeyMapToData(model.AttributionKeyAdgroup, adwordsAdgroupAllRows, data.AdwordsAdgroupIDData)

	data.AdwordsKeywordIDData = adwordsKeywordIDData
	data.AdwordsKeywordKeyData = model.GetKeyMapToData(model.AttributionKeyKeyword, adwordsKeywordAllRows, data.AdwordsKeywordIDData)

	data.BingAdsCampaignIDData = bingadsCampaignIDData
	data.BingAdsCampaignKeyData = model.GetKeyMapToData(model.AttributionKeyCampaign, bingadsCampaignAllRows, data.BingAdsCampaignIDData)

	data.BingAdsAdgroupIDData = bingadsAdgroupIDData
	data.BingAdsAdgroupKeyData = model.GetKeyMapToData(model.AttributionKeyAdgroup, bingadsAdgroupAllRows, data.BingAdsAdgroupIDData)

	data.BingAdsKeywordIDData = bingadsKeywordIDData
	data.BingAdsKeywordKeyData = model.GetKeyMapToData(model.AttributionKeyKeyword, bingadsKeywordAllRows, data.BingAdsKeywordIDData)

	data.FacebookCampaignIDData = facebookCampaignIDData
	data.FacebookCampaignKeyData = model.GetKeyMapToData(model.AttributionKeyCampaign, facebookCampaignAllRows, data.FacebookCampaignIDData)

	data.FacebookAdgroupIDData = facebookAdgroupIDData
	data.FacebookAdgroupKeyData = model.GetKeyMapToData(model.AttributionKeyAdgroup, facebookAdgroupAllRows, data.FacebookAdgroupIDData)

	data.LinkedinCampaignIDData = linkedinCampaignIDData
	data.LinkedinCampaignKeyData = model.GetKeyMapToData(model.AttributionKeyCampaign, linkedinCampaignAllRows, data.LinkedinCampaignIDData)

	data.LinkedinAdgroupIDData = linkedinAdgroupIDData
	data.LinkedinAdgroupKeyData = model.GetKeyMapToData(model.AttributionKeyAdgroup, linkedinAdgroupAllRows, data.LinkedinAdgroupIDData)

	data.CustomAdsCampaignIDData = customadsCampaignIDData
	data.CustomAdsCampaignKeyData = model.GetKeyMapToData(model.AttributionKeyCampaign, customadsCampaignAllRows, data.CustomAdsCampaignIDData)

	data.CustomAdsAdgroupIDData = customadsAdgroupIDData
	data.CustomAdsAdgroupKeyData = model.GetKeyMapToData(model.AttributionKeyAdgroup, customadsAdgroupAllRows, data.CustomAdsAdgroupIDData)

	data.CustomAdsKeywordIDData = customadsKeywordIDData
	data.CustomAdsKeywordKeyData = model.GetKeyMapToData(model.AttributionKeyKeyword, customadsKeywordAllRows, data.CustomAdsKeywordIDData)

	return data, err
}
func (store *MemSQL) getBingAdsAccountId(projectID int64) (string, error) {
	ftMapping, err := store.GetActiveFiveTranMapping(projectID, model.BingAdsIntegration)
	if err == nil {
		return ftMapping.Accounts, nil
	} else {
		return "", nil
	}
}

// PullAdwordsMarketingData Pulls Adds channel data for Adwords.
func (store *MemSQL) PullAdwordsMarketingData(projectID int64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, reportName string, timeZone string) (map[string]model.MarketingData, []model.MarketingData, error) {
	logFields := log.Fields{
		"project_id":          projectID,
		"from":                from,
		"to":                  to,
		"customer_account_id": customerAccountID,
		"key_id":              keyID,
		"key_name":            keyName,
		"extra_value1":        extraValue1,
		"report_name":         reportName,
		"time_zone":           timeZone,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	customerAccountIDs := strings.Split(customerAccountID, ",")
	performanceQuery := "SELECT campaign_id as campaignID, ad_group_id as adgroupID, keyword_id as keywordID, ad_id as adID, " +
		"JSON_EXTRACT_STRING(value, ?) AS key_id, JSON_EXTRACT_STRING(value, ?) AS key_name, JSON_EXTRACT_STRING(value, ?) AS extra_value1, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'cost'))/1000000 AS total_cost FROM adwords_documents " +
		"where project_id = ? AND customer_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, adID, key_id, key_name, extra_value1 " + "order by timestamp"

	params := []interface{}{keyID, keyName, extraValue1, projectID, customerAccountIDs, reportType,
		U.GetDateAsStringIn(from, U.TimeZoneString(timeZone)), U.GetDateAsStringIn(to, U.TimeZoneString(timeZone))}
	rows, tx, err, reqID := store.ExecQueryWithContext(performanceQuery, params)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	marketingDataIDMap, allRows := model.ProcessRow(rows, reportName, logCtx, model.ChannelAdwords, reqID)
	return marketingDataIDMap, allRows, nil
}

// PullFacebookMarketingData Pulls Adds channel data for Facebook.
func (store *MemSQL) PullFacebookMarketingData(projectID int64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, reportName string, timeZone string) (map[string]model.MarketingData, []model.MarketingData, error) {
	logFields := log.Fields{
		"project_id":          projectID,
		"from":                from,
		"to":                  to,
		"customer_account_id": customerAccountID,
		"key_id":              keyID,
		"key_name":            keyName,
		"extra_value1":        extraValue1,
		"report_name":         reportName,
		"time_zone":           timeZone,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	customerAccountIDs := strings.Split(customerAccountID, ",")
	performanceQuery := "SELECT campaign_id as campaignID, ad_set_id as adgroupID, '$none' as keywordID, ad_id as adID, " +
		"JSON_EXTRACT_STRING(value, ?) AS key_id, JSON_EXTRACT_STRING(value, ?) AS key_name, JSON_EXTRACT_STRING(value, ?) AS extra_value1, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'inline_link_clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'spend')) AS total_cost FROM facebook_documents " +
		"where project_id = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, adID, key_id, key_name, extra_value1 " + "order by timestamp"

	params := []interface{}{keyID, keyName, extraValue1, projectID, customerAccountIDs, reportType,
		U.GetDateAsStringIn(from, U.TimeZoneString(timeZone)), U.GetDateAsStringIn(to, U.TimeZoneString(timeZone))}

	rows, tx, err, reqID := store.ExecQueryWithContext(performanceQuery, params)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	marketingDataIDMap, allRows := model.ProcessRow(rows, reportName, logCtx, model.ChannelFacebook, reqID)
	return marketingDataIDMap, allRows, nil
}

// PullLinkedinMarketingData Pulls Adds channel data for Linkedin.
func (store *MemSQL) PullLinkedinMarketingData(projectID int64, from, to int64, customerAccountID string, keyID string,
	keyName string, extraValue1 string, reportType int, reportName string, timeZone string) (map[string]model.MarketingData, []model.MarketingData, error) {
	logFields := log.Fields{
		"project_id":          projectID,
		"from":                from,
		"to":                  to,
		"customer_account_id": customerAccountID,
		"key_id":              keyID,
		"key_name":            keyName,
		"extra_value1":        extraValue1,
		"report_name":         reportName,
		"time_zone":           timeZone,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	customerAccountIDs := strings.Split(customerAccountID, ",")
	performanceQuery := "SELECT campaign_group_id as campaignID, campaign_id as adgroupID, '$none' as keywordID, creative_id as adID, " +
		"JSON_EXTRACT_STRING(value, ?) AS key_id, JSON_EXTRACT_STRING(value, ?) AS key_name, JSON_EXTRACT_STRING(value, ?) AS extra_value1, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'costInLocalCurrency')) AS total_spend FROM linkedin_documents " +
		"where project_id = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, adID, key_id, key_name, extra_value1 " + "order by timestamp"

	params := []interface{}{keyID, keyName, extraValue1, projectID, customerAccountIDs, reportType,
		U.GetDateAsStringIn(from, U.TimeZoneString(timeZone)), U.GetDateAsStringIn(to, U.TimeZoneString(timeZone))}
	rows, tx, err, reqID := store.ExecQueryWithContext(performanceQuery, params)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	marketingDataIDMap, allRows := model.ProcessRow(rows, reportName, logCtx, model.ChannelLinkedin, reqID)
	return marketingDataIDMap, allRows, nil
}
func (store *MemSQL) PullBingAdsMarketingData(projectID int64, from, to int64, customerAccountID string, keyID string,
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
		"'$none' as adId, JSON_EXTRACT_STRING(value, ?) AS key_id, JSON_EXTRACT_STRING(value, ?) AS key_name, JSON_EXTRACT_STRING(value, ?) AS extra_value1, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'spend')) AS total_spend FROM integration_documents " +
		"where project_id = ? AND source = ? AND customer_account_id IN (?) AND document_type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, key_id, key_name, extra_value1 " + "order by timestamp"

	params := []interface{}{keyID, keyName, extraValue1, projectID, model.BingAdsIntegration, customerAccountIDs, reportType,
		U.GetDateAsStringIn(from, U.TimeZoneString(timeZone)), U.GetDateAsStringIn(to, U.TimeZoneString(timeZone))}
	rows, tx, err, reqID := store.ExecQueryWithContext(performanceQuery, params)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	marketingDataIDMap, allRows := model.ProcessRow(rows, reportName, logCtx, model.BingAdsIntegration, reqID)
	return marketingDataIDMap, allRows, nil
}

func (store *MemSQL) PullCustomAdsMarketingData(projectID int64, from, to int64, customerAccountID []string, keyID string,
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
	performanceQuery := "SELECT JSON_EXTRACT_STRING(value, 'campaign_id')  as campaignID, JSON_EXTRACT_STRING(value, 'ad_group_id') as adgroupID, JSON_EXTRACT_STRING(value, 'keyword_id') as keywordID, " +
		"'$none' as adId, JSON_EXTRACT_STRING(value, ?) AS key_id, JSON_EXTRACT_STRING(value, ?) AS key_name, JSON_EXTRACT_STRING(value, ?) AS extra_value1, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'spend')) AS total_spend, source FROM integration_documents " +
		"where project_id = ? AND source IN (?) AND customer_account_id IN (?) AND document_type = ? AND timestamp between ? AND ? " +
		"group by campaignID, adgroupID, keywordID, key_id, key_name, extra_value1, source"

	sources, _ := store.GetCustomAdsSourcesByProject(projectID)
	params := []interface{}{keyID, keyName, extraValue1, projectID, sources, customerAccountID, reportType,
		U.GetDateAsStringIn(from, U.TimeZoneString(timeZone)), U.GetDateAsStringIn(to, U.TimeZoneString(timeZone))}
	rows, tx, err, reqID := store.ExecQueryWithContext(performanceQuery, params)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	marketingDataIDMap, allRows := model.ProcessRow(rows, reportName, logCtx, model.CustomAdsIntegration, reqID)
	return marketingDataIDMap, allRows, nil
}

func (store *MemSQL) PullCustomDimensionData(projectID int64, attributionKey string, marketingReport *model.MarketingReports, logCtx log.Entry) error {

	// Custom Dimensions are support only for Campaign and Adgroup currently
	if attributionKey != model.AttributionKeyCampaign && attributionKey != model.AttributionKeyAdgroup {
		return nil
	}
	enableBingAdsAttribution := C.GetConfig().EnableBingAdsAttribution
	isCustomAdsIntegrationDone := store.IsCustomAdsAvailable(projectID)
	sources, _ := store.GetCustomAdsSourcesByProject(projectID)

	var err error
	switch attributionKey {
	case model.AttributionKeyCampaign:

		marketingReport.AdwordsCampaignDimensions, err = store.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, []string{model.ChannelAdwords}, 1, attributionKey, logCtx)
		if err != nil {
			return err
		}
		marketingReport.FacebookCampaignDimensions, err = store.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, []string{model.ChannelFacebook}, 1, attributionKey, logCtx)
		if err != nil {
			return err
		}
		marketingReport.LinkedinCampaignDimensions, err = store.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, []string{model.ChannelLinkedin}, 1, attributionKey, logCtx)
		if err != nil {
			return err
		}
		if enableBingAdsAttribution {
			marketingReport.BingadsCampaignDimensions, err = store.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, []string{model.ChannelBingads}, 1, attributionKey, logCtx)
			if err != nil {
				return err
			}
		}
		if isCustomAdsIntegrationDone {
			marketingReport.CustomAdsCampaignDimensions, err = store.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, sources, 1, attributionKey, logCtx)
			if err != nil {
				return err
			}
		}
	case model.FieldAdgroupName:
		marketingReport.AdwordsAdgroupDimensions, err = store.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, []string{model.ChannelAdwords}, 2, attributionKey, logCtx)
		if err != nil {
			return err
		}
		marketingReport.FacebookAdgroupDimensions, err = store.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, []string{model.ChannelFacebook}, 2, attributionKey, logCtx)
		if err != nil {
			return err
		}
		marketingReport.LinkedinAdgroupDimensions, err = store.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, []string{model.ChannelLinkedin}, 2, attributionKey, logCtx)
		if err != nil {
			return err
		}
		if enableBingAdsAttribution {
			marketingReport.BingadsAdgroupDimensions, err = store.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, []string{model.ChannelBingads}, 2, attributionKey, logCtx)
			if err != nil {
				return err
			}
		}
		if isCustomAdsIntegrationDone {
			marketingReport.CustomAdsAdgroupDimensions, err = store.PullSmartProperties(projectID, model.SmartPropertyCampaignID, model.SmartPropertyCampaignName, model.SmartPropertyAdGroupID, model.SmartPropertyAdGroupName, sources, 2, attributionKey, logCtx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// PullSmartProperties Pulls Smart Properties
func (store *MemSQL) PullSmartProperties(projectID int64, campaignIDPlaceHolder string, campaignNamePlaceHolder string, adgroupIDPlaceHolder string, adgroupNamePlaceHolder string, sourceChannelPlaceHolder []string, objectType int, attributionKey string, logCtx log.Entry) (map[string]model.MarketingData, error) {
	logFields := log.Fields{
		"campaign_id_place_holder":    campaignIDPlaceHolder,
		"campaign_name_place_holder":  campaignNamePlaceHolder,
		"adgroup_id_place_holder":     adgroupIDPlaceHolder,
		"ad_group_name_place_holder":  adgroupNamePlaceHolder,
		"source_channel_place_holder": sourceChannelPlaceHolder,
		"object_type":                 objectType,
		"attribution_key":             attributionKey,
	}

	// GetEventsWithoutPropertiesAndWithPropertiesByNameForYourStory
	logCtx1 := logCtx.WithFields(logFields)
	stmt := "SELECT JSON_EXTRACT_STRING(object_property, ?) AS campaignID,  JSON_EXTRACT_STRING(object_property, ?) AS campaignName, " +
		"JSON_EXTRACT_STRING(object_property, ?) AS adgroupID,  JSON_EXTRACT_STRING(object_property, ?) AS adgroupName, " +
		"properties FROM smart_properties " +
		"where project_id = ? AND source IN (?) AND object_type = ?"

	params := []interface{}{campaignIDPlaceHolder, campaignNamePlaceHolder, adgroupIDPlaceHolder, adgroupNamePlaceHolder, projectID, sourceChannelPlaceHolder, objectType}
	rows, tx, err, reqID := store.ExecQueryWithContext(stmt, params)
	if err != nil {
		logCtx1.WithError(err).Error("SQL Query failed")
		return nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	startReadTime := time.Now()
	dataKeyDimensions := make(map[string]model.MarketingData)
	for rows.Next() {
		var campaignIDNull sql.NullString
		var campaignNameNull sql.NullString
		var adgroupIDNull sql.NullString
		var adgroupNameNull sql.NullString
		var properties postgres.Jsonb

		if err := rows.Scan(&campaignIDNull, &campaignNameNull, &adgroupIDNull, &adgroupNameNull, &properties); err != nil {
			logCtx1.WithError(err).Error("Bad row. Ignoring row and continuing")
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
			logCtx1.WithError(err).Error("Failed to decode smart properties. Ignoring row and continuing")
			continue
		}
		key := model.GetKeyForCustomDimensionsName(_campaignID, _campaignName, _adgroupID, _adgroupName, attributionKey)

		if key == "" {
			continue
		}
		// added custom dimensions
		(*propertiesMap)["campaign_id"] = _campaignID
		(*propertiesMap)["campaign_name"] = _campaignName
		(*propertiesMap)["adgroup_id"] = _adgroupID
		(*propertiesMap)["adgroup_name"] = _adgroupName

		dataKeyDimensions[key] = model.MarketingData{
			Key:              key,
			CampaignID:       _campaignID,
			CampaignName:     _campaignName,
			AdgroupID:        _adgroupID,
			AdgroupName:      _adgroupName,
			CustomDimensions: *propertiesMap,
		}

	}
	logCtx1.WithFields(log.Fields{"CustomDebug": "True", "ProjectId": projectID,
		"Source": sourceChannelPlaceHolder}).
		Info("Pull Smart Properties")
	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &logFields)

	return dataKeyDimensions, nil
}
