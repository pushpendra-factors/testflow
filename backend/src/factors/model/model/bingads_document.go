package model

import (
	U "factors/util"
	"fmt"
)

const (
	// Unique Metrics to BingAds
	Conversions = "conversions"
)

var BingadsDocumentTypeAlias = map[string]int{
	"campaigns":               1,
	"ad_groups":               2,
	"keyword":                 3,
	CampaignPerformanceReport: 4,
	AdGroupPerformanceReport:  5,
	KeywordPerformanceReport:  6,
	"account":                 7,
}

var AllAccountQuery = "SELECT DISTINCT id FROM `%s.%s.account_history`"
var BingAdsDocumentToQuery = map[string]string{
	"campaigns": "SELECT id, account_id, budget, status, name, type FROM `%s.%s.campaign_history` WHERE %v",
	"ad_groups": "select DISTINCT ad.id, ad.campaign_id, ad.name, ad.status, ad.bid_option, ad.bid_strategy_type, ch.account_id " +
		"From `%s.%s.ad_group_history` AS ad inner join `%s.%s.campaign_history` AS ch " +
		"on ad.campaign_id = ch.id WHERE %v",
	"keyword": "select DISTINCT kh.id,kh.bid,kh.ad_group_id,kh.name,kh.status,kh.match_type,ad.campaign_id,ch.account_id " +
		"From `%s.%s.keyword_history` AS kh inner join `%s.%s.ad_group_history` AS ad on kh.ad_group_id = ad.id " +
		"inner join `%s.%s.campaign_history` AS ch on ad.campaign_id = ch.id WHERE %v",
	CampaignPerformanceReport: "SELECT SUM(cpd.impressions), SUM(cpd.spend), SUM(cpd.clicks), SUM(cpd.conversions), MAX(cpd.account_id), cpd.campaign_id, MAX(cpd.currency_code), " +
		"MAX(ch.name) AS campaign_name, MAX(ch.status) AS campaign_status, MAX(ch.type) AS campaign_type FROM " +
		"`%s.%s.campaign_performance_daily_report`  AS cpd " +
		" left outer join " +
		"`%s.%s.campaign_history` AS ch " +
		"ON cpd.campaign_id = ch.id WHERE %v GROUP BY cpd.campaign_id",
	AdGroupPerformanceReport: "SELECT SUM(agp.impressions), SUM(agp.spend), SUM(agp.clicks), SUM(agp.conversions), MAX(agp.account_id) , MAX(agp.campaign_id), MAX(agp.currency_code), agp.ad_group_id, " +
		" MAX(ad.name) AS ad_group_name, MAX(ad.status) AS ad_group_status, MAX(ad.bid_strategy_type) AS ad_group_bid_strategy_type, " +
		" MAX(ch.name) AS campaign_name, MAX(ch.status) AS campaign_status, MAX(ch.type) AS campaign_type " +
		" FROM  " +
		" `%s.%s.ad_group_performance_daily_report`  AS agp " +
		" left outer join  " +
		" `%s.%s.ad_group_history` AS ad " +
		" ON agp.ad_group_id = ad.id " +
		" left outer join  " +
		" `%s.%s.campaign_history` AS ch " +
		" ON agp.campaign_id = ch.id WHERE %v GROUP BY agp.ad_group_id",
	KeywordPerformanceReport: "SELECT SUM(kwp.impressions), SUM(kwp.spend), SUM(kwp.clicks), SUM(kwp.conversions), MAX(kwp.account_id),MAX(kwp.campaign_id),MAX(kwp.currency_code),MAX(kwp.ad_group_id),kwp.keyword_id, " +
		" MAX(kw.name) AS keyword_name, MAX(kw.status) AS keyword_status, MAX(kw.match_type) AS keyword_match_type, " +
		" MAX(ad.name) AS ad_group_name ,MAX(ad.status) AS ad_group_status, MAX(ad.bid_strategy_type) AS ad_group_bid_strategy_type, " +
		" MAX(ch.name) AS campaign_name, MAX(ch.status) AS campaign_status ,MAX(ch.type) AS campaign_type " +
		" FROM  " +
		" `%s.%s.keyword_performance_daily_report`  AS kwp " +
		" left outer join  " +
		" `%s.%s.keyword_history` AS kw " +
		" ON kwp.keyword_id = kw.id " +
		" left outer join  " +
		" `%s.%s.ad_group_history` AS ad " +
		" ON kwp.ad_group_id = ad.id " +
		" left outer join  " +
		" `%s.%s.campaign_history` AS ch " +
		" ON kwp.campaign_id = ch.id WHERE %v GROUP BY kwp.keyword_id",
	"account": "SELECT id, name, currency_code, time_zone FROM `%s.%s.account_history` WHERE %v",
}

var BingAdsDataObjectFilters = map[string]string{
	"campaigns":               "DATE(%v) = '%v'",
	"ad_groups":               "DATE(%v) = '%v'",
	"keyword":                 "DATE(%v) = '%v'",
	CampaignPerformanceReport: "%v = '%v'",
	AdGroupPerformanceReport:  "%v = '%v'",
	KeywordPerformanceReport:  "%v = '%v'",
	"account":                 "DATE(%v) = '%v'",
}

var BingAdsDataObjectColumnsInValue = map[string]map[string]int{
	"campaigns":               {"id": 0, "account_id": 1, "budget": 2, "status": 3, "name": 4, "type": 5},
	"ad_groups":               {"id": 0, "campaign_id": 1, "name": 2, "status": 3, "bid_option": 4, "bid_strategy_type": 5, "account_id": 6},
	"keyword":                 {"id": 0, "bid": 1, "ad_group_id": 2, "name": 3, "status": 4, "match_type": 5, "campaign_id": 6, "account_id": 7},
	CampaignPerformanceReport: {"impressions": 0, "spend": 1, "clicks": 2, "conversions": 3, "account_id": 4, "campaign_id": 5, "currency_code": 6, "campaign_name": 7, "campaign_status": 8, "campaign_type": 9},
	AdGroupPerformanceReport:  {"impressions": 0, "spend": 1, "clicks": 2, "conversions": 3, "account_id": 4, "campaign_id": 5, "currency_code": 6, "ad_group_id": 7, "ad_group_name": 8, "ad_group_status": 9, "ad_group_bid_strategy_type": 10, "campaign_name": 11, "campaign_status": 12, "campaign_type": 13},
	KeywordPerformanceReport:  {"impressions": 0, "spend": 1, "clicks": 2, "conversions": 3, "account_id": 4, "campaign_id": 5, "currency_code": 6, "ad_group_id": 7, "keyword_id": 8, "keyword_name": 9, "keyword_status": 10, "keyword_match_type": 11, "ad_group_name": 12, "ad_group_status": 13, "ad_group_bid_strategy_type": 14, "campaign_name": 15, "campaign_status": 16, "campaign_type": 17},
	"account":                 {"id": 0, "name": 1, "currency_code": 2, "time_zone": 3},
}

var BingAdsDataObjectFiltersColumn = map[string]string{
	"campaigns":               "_fivetran_synced",
	"ad_groups":               "_fivetran_synced",
	"keyword":                 "_fivetran_synced",
	CampaignPerformanceReport: "date",
	AdGroupPerformanceReport:  "date",
	KeywordPerformanceReport:  "date",
	"account":                 "_fivetran_synced",
}

func GetBingAdsDocumentDocumentId(documentType int, data []string) string {
	if documentType == BingadsDocumentTypeAlias["campaigns"] {
		return data[BingAdsDataObjectColumnsInValue["campaigns"]["id"]]
	} else if documentType == BingadsDocumentTypeAlias["ad_groups"] {
		return data[BingAdsDataObjectColumnsInValue["ad_groups"]["id"]]
	} else if documentType == BingadsDocumentTypeAlias["keyword"] {
		return data[BingAdsDataObjectColumnsInValue["keyword"]["id"]]
	} else if documentType == BingadsDocumentTypeAlias[CampaignPerformanceReport] {
		return data[BingAdsDataObjectColumnsInValue[CampaignPerformanceReport]["campaign_id"]]
	} else if documentType == BingadsDocumentTypeAlias[AdGroupPerformanceReport] {
		return data[BingAdsDataObjectColumnsInValue[AdGroupPerformanceReport]["ad_group_id"]]
	} else if documentType == BingadsDocumentTypeAlias[KeywordPerformanceReport] {
		return data[BingAdsDataObjectColumnsInValue[KeywordPerformanceReport]["keyword_id"]]
	} else if documentType == BingadsDocumentTypeAlias["account"] {
		return data[BingAdsDataObjectColumnsInValue["account"]["id"]]
	}
	return ""
}

func GetBingAdsDocumentAccountId(documentType int, data []string) string {
	if documentType == BingadsDocumentTypeAlias["campaigns"] {
		return data[BingAdsDataObjectColumnsInValue["campaigns"]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias["ad_groups"] {
		return data[BingAdsDataObjectColumnsInValue["ad_groups"]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias["keyword"] {
		return data[BingAdsDataObjectColumnsInValue["keyword"]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias[CampaignPerformanceReport] {
		return data[BingAdsDataObjectColumnsInValue[CampaignPerformanceReport]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias[AdGroupPerformanceReport] {
		return data[BingAdsDataObjectColumnsInValue[AdGroupPerformanceReport]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias[KeywordPerformanceReport] {
		return data[BingAdsDataObjectColumnsInValue[KeywordPerformanceReport]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias["account"] {
		return data[BingAdsDataObjectColumnsInValue["account"]["id"]]
	}
	return ""
}

func GetBingAdsDocumentValues(docType string, data []string) map[string]interface{} {
	values := make(map[string]interface{})
	for key, index := range BingAdsDataObjectColumnsInValue[docType] {
		values[key] = data[index]
	}
	return values
}
func GetBingAdsDocumentDocumentType(documentTypeString string) int {
	docTypeId, exists := BingadsDocumentTypeAlias[documentTypeString]
	if exists {
		return docTypeId
	}
	return 0
}

func GetBingAdsDocumentQuery(bigQueryProjectId string, schemaId string, baseQuery string, executionDate string, docType string) string {

	if docType == "campaigns" {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, GetBingAdsDocumentFilterCondition(docType, false, "", executionDate))
	} else if docType == "ad_groups" {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, GetBingAdsDocumentFilterCondition(docType, true, "ad", executionDate))
	} else if docType == "keyword" {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, GetBingAdsDocumentFilterCondition(docType, true, "kh", executionDate))
	} else if docType == CampaignPerformanceReport {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, GetBingAdsDocumentFilterCondition(docType, true, "cpd", executionDate))
	} else if docType == AdGroupPerformanceReport {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, GetBingAdsDocumentFilterCondition(docType, true, "agp", executionDate))
	} else if docType == KeywordPerformanceReport {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, GetBingAdsDocumentFilterCondition(docType, true, "kwp", executionDate))
	} else if docType == "account" {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, GetBingAdsDocumentFilterCondition(docType, false, "", executionDate))
	}
	return ""
}

func GetBingAdsDocumentFilterCondition(docType string, addPrefix bool, prefix string, executionDate string) string {
	filterColumn := ""
	if addPrefix {
		filterColumn = fmt.Sprintf("%v.%v", prefix, BingAdsDataObjectFiltersColumn[docType])
	} else {
		filterColumn = BingAdsDataObjectFiltersColumn[docType]
	}
	filterCondition := fmt.Sprintf(BingAdsDataObjectFilters[docType], filterColumn, executionDate)

	return filterCondition
}

const (
	BingSpecificError = "Failed in bingads with the error."
)

var ObjectsForBingads = []string{FilterCampaign, FilterAdGroup, FilterKeyword}

// for bing ads
const (
	FilterCampaign = "campaign"
	FilterAdGroup  = "ad_group"
	FilterKeyword  = "keyword"
)

var MapOfBingAdsObjectsToPropertiesAndRelated = map[string]map[string]PropertiesAndRelated{
	FilterCampaign: {
		"name":   PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"id":     PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"status": PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"type":   PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
	},
	FilterAdGroup: {
		"name":              PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"id":                PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"status":            PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"bid_strategy_type": PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
	},
	FilterKeyword: {
		"name":       PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"id":         PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"status":     PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"match_type": PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
	},
}
var BingAdsInternalRepresentationToExternalRepresentation = map[string]string{
	"campaigns.id":                "id",
	"campaigns.status":            "status",
	"campaigns.name":              "name",
	"campaigns.type":              "type",
	"ad_groups.id":                "id",
	"ad_groups.status":            "status",
	"ad_groups.name":              "name",
	"ad_groups.bid_strategy_type": "bid_strategy_type",
	"keyword.id":                  "id",
	"keyword.name":                "name",
	"keyword.status":              "status",
	"keyword.match_type":          "match_type",
	"impressions":                 "impressions",
	"clicks":                      "clicks",
	"spend":                       "spend",
	"conversions":                 "conversions",
	"channel.name":                "channel_name",
}

var BingAdsInternalRepresentationToExternalRepresentationForReports = map[string]string{
	"campaigns.id":                "campaign_id",
	"campaigns.status":            "campaign_status",
	"campaigns.name":              "campaign_name",
	"campaigns.type":              "campaign_type",
	"ad_groups.id":                "ad_group_id",
	"ad_groups.status":            "ad_group_status",
	"ad_groups.name":              "ad_group_name",
	"ad_groups.bid_strategy_type": "ad_group_bid_strategy_type",
	"keyword.id":                  "keyword_id",
	"keyword.name":                "keyword_name",
	"keyword.status":              "keyword_status",
	"keyword.match_type":          "keyword_match_type",
	"impressions":                 "impressions",
	"clicks":                      "clicks",
	"spend":                       "spend",
	"conversions":                 "conversions",
	"channel.name":                "channel_name",
}

var BingAdsObjectInternalRepresentationToExternalRepresentation = map[string]string{
	FilterCampaign: "campaigns",
	FilterAdGroup:  "ad_groups",
	FilterKeyword:  "keyword",
	"channel":      "channel",
}

var BingAdsObjectToPerfomanceReportRepresentation = map[string]string{
	"campaigns": CampaignPerformanceReport,
	"ad_groups": AdGroupPerformanceReport,
	"keyword":   KeywordPerformanceReport,
}

func GetAllAccountsQuery(bigQueryProjectId string, schemaId string) string {
	return fmt.Sprintf(AllAccountQuery, bigQueryProjectId, schemaId)
}
