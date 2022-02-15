package model

import (
	"fmt"
	"sort"
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

var BingAdsDocumentToQuery = map[string]string{
	"campaigns":               "SELECT %v FROM `%s.%s.campaign_history` WHERE %v",
	"ad_groups":               "select DISTINCT %v From `%v.%v.ad_group_history` AS ad inner join `%s.%s.campaign_history` AS ch on ad.campaign_id = ch.id WHERE %v",
	"keyword":                 "select DISTINCT %v  From `%v.%v.keyword_history` AS kh inner join `%v.%v.ad_group_history` AS ad on kh.ad_group_id = ad.id inner join `%v.%v.campaign_history` AS ch on ad.campaign_id = ch.id WHERE %v",
	CampaignPerformanceReport: "SELECT %v FROM `%s.%s.campaign_performance_daily_report` WHERE %v GROUP BY campaign_id",
	AdGroupPerformanceReport:  "SELECT %v FROM `%s.%s.ad_group_performance_daily_report` WHERE %v GROUP BY ad_group_id",
	KeywordPerformanceReport:  "SELECT %v FROM `%s.%s.keyword_performance_daily_report` WHERE %v GROUP BY keyword_id",
	"account":                 "SELECT %v FROM `%s.%s.account_history` WHERE %v",
}

var BingAdsDocumentToTable = map[string]string{
	"campaigns":               "campaign_history",
	"ad_groups":               "ad_group_history",
	"keyword":                 "keyword_history",
	CampaignPerformanceReport: "campaign_performance_daily_report",
	AdGroupPerformanceReport:  "ad_group_performance_daily_report",
	KeywordPerformanceReport:  "keyword_performance_daily_report",
	"account":                 "account_history",
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

var BingAdsDataObjectColumns = map[string]map[string]int{
	"campaigns":               {"id": 0, "account_id": 1, "budget": 2, "status": 3, "name": 4, "type": 5},
	"ad_groups":               {"id": 0, "campaign_id": 1, "name": 2, "status": 3, "bid_option": 4, "bid_strategy_type": 5, "account_id": 6},
	"keyword":                 {"id": 0, "bid": 1, "ad_group_id": 2, "name": 3, "status": 4, "match_type": 5, "campaign_id": 6, "account_id": 7},
	CampaignPerformanceReport: {"impressions": 0, "spend": 1, "clicks": 2, "conversions": 3, "account_id": 4, "campaign_id": 5, "currency_code": 6},
	AdGroupPerformanceReport:  {"impressions": 0, "spend": 1, "clicks": 2, "conversions": 3, "account_id": 4, "campaign_id": 5, "ad_group_id": 6, "currency_code": 7},
	KeywordPerformanceReport:  {"impressions": 0, "spend": 1, "clicks": 2, "conversions": 3, "account_id": 4, "campaign_id": 5, "ad_group_id": 6, "keyword_id": 7, "currency_code": 8},
	"account":                 {"id": 0, "name": 1, "currency_code": 2, "time_zone": 3},
}

var BingAdsDataObjectColumnsAgg = map[string]map[string]string{
	CampaignPerformanceReport: {"impressions": "sum", "spend": "sum", "clicks": "sum", "conversions": "sum", "account_id": "max", "campaign_id": "", "currency_code": "max"},
	AdGroupPerformanceReport:  {"impressions": "sum", "spend": "sum", "clicks": "sum", "conversions": "sum", "account_id": "max", "campaign_id": "max", "ad_group_id": "", "currency_code": "max"},
	KeywordPerformanceReport:  {"impressions": "sum", "spend": "sum", "clicks": "sum", "conversions": "sum", "account_id": "max", "campaign_id": "max", "ad_group_id": "max", "keyword_id": "", "currency_code": "max"},
}
var BingAdsDataObjectColumnsInValue = map[string]map[string]int{
	"campaigns":               {"budget": 2, "status": 3, "name": 4, "type": 5},
	"ad_groups":               {"campaign_id": 1, "name": 2, "status": 3, "bid_option": 4, "bid_strategy_type": 5},
	"keyword":                 {"bid": 1, "ad_group_id": 2, "name": 3, "status": 4, "match_type": 5, "campaign_id": 6},
	CampaignPerformanceReport: {"impressions": 0, "spend": 1, "clicks": 2, "conversions": 3, "currency_code": 6},
	AdGroupPerformanceReport:  {"impressions": 0, "spend": 1, "clicks": 2, "conversions": 3, "campaign_id": 5, "currency_code": 7},
	KeywordPerformanceReport:  {"impressions": 0, "spend": 1, "clicks": 2, "conversions": 3, "campaign_id": 5, "ad_group_id": 6, "currency_code": 8},
	"account":                 {"name": 1, "currency_code": 2, "time_zone": 3},
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
		return data[BingAdsDataObjectColumns["campaigns"]["id"]]
	} else if documentType == BingadsDocumentTypeAlias["ad_groups"] {
		return data[BingAdsDataObjectColumns["ad_groups"]["id"]]
	} else if documentType == BingadsDocumentTypeAlias["keyword"] {
		return data[BingAdsDataObjectColumns["keyword"]["id"]]
	} else if documentType == BingadsDocumentTypeAlias[CampaignPerformanceReport] {
		return data[BingAdsDataObjectColumns[CampaignPerformanceReport]["campaign_id"]]
	} else if documentType == BingadsDocumentTypeAlias[AdGroupPerformanceReport] {
		return data[BingAdsDataObjectColumns[AdGroupPerformanceReport]["ad_group_id"]]
	} else if documentType == BingadsDocumentTypeAlias[KeywordPerformanceReport] {
		return data[BingAdsDataObjectColumns[KeywordPerformanceReport]["keyword_id"]]
	} else if documentType == BingadsDocumentTypeAlias["account"] {
		return data[BingAdsDataObjectColumns["account"]["id"]]
	}
	return ""
}

func GetBingAdsDocumentAccountId(documentType int, data []string) string {
	if documentType == BingadsDocumentTypeAlias["campaigns"] {
		return data[BingAdsDataObjectColumns["campaigns"]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias["ad_groups"] {
		return data[BingAdsDataObjectColumns["ad_groups"]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias["keyword"] {
		return data[BingAdsDataObjectColumns["keyword"]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias[CampaignPerformanceReport] {
		return data[BingAdsDataObjectColumns[CampaignPerformanceReport]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias[AdGroupPerformanceReport] {
		return data[BingAdsDataObjectColumns[AdGroupPerformanceReport]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias[KeywordPerformanceReport] {
		return data[BingAdsDataObjectColumns[KeywordPerformanceReport]["account_id"]]
	} else if documentType == BingadsDocumentTypeAlias["account"] {
		return data[BingAdsDataObjectColumns["account"]["id"]]
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

func ReturnBingAdsDocumentCommaSeperatedColumns(ColumnsMap map[string]int, aggMap map[string]string, addPrefix bool, prefix string, n int) string {
	type columnIndexTuple struct {
		id    string
		index int
	}
	columnsArray := make([]columnIndexTuple, 0)
	for column_id, column_index := range ColumnsMap {
		columnsArray = append(columnsArray, columnIndexTuple{id: column_id, index: column_index})
	}
	sort.Slice(columnsArray, func(i, j int) bool {
		return columnsArray[j].index > columnsArray[i].index
	})
	columnString := ""
	for i, columnObj := range columnsArray {
		prefixedColumId := ""
		if addPrefix == true && i < n {
			prefixedColumId = fmt.Sprintf("%v.%v", prefix, columnObj.id)
		} else {
			prefixedColumId = columnObj.id
		}
		if aggMap != nil {
			if aggMap[columnObj.id] != "" {
				prefixedColumId = fmt.Sprintf("%v(%v)", aggMap[columnObj.id], prefixedColumId)
			}
		}
		if columnString == "" {
			columnString = prefixedColumId
		} else {
			columnString = fmt.Sprintf("%v,%v", columnString, prefixedColumId)
		}
	}
	return columnString
}

func GetBingAdsDocumentQuery(bigQueryProjectId string, schemaId string, docType string, executionDate string) string {

	if docType == "ad_groups" {
		columns := ReturnBingAdsDocumentCommaSeperatedColumns(BingAdsDataObjectColumns[docType], BingAdsDataObjectColumnsAgg[docType], true, "ad", 6)
		return fmt.Sprintf(BingAdsDocumentToQuery[docType], columns, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, GetBingAdsDocumentFilterCondition(docType, true, "ad", executionDate))
	} else if docType == "keyword" {
		columns := ReturnBingAdsDocumentCommaSeperatedColumns(BingAdsDataObjectColumns[docType], BingAdsDataObjectColumnsAgg[docType], true, "kh", 6)
		return fmt.Sprintf(BingAdsDocumentToQuery[docType], columns, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, GetBingAdsDocumentFilterCondition(docType, true, "kh", executionDate))
	} else {
		columns := ReturnBingAdsDocumentCommaSeperatedColumns(BingAdsDataObjectColumns[docType], BingAdsDataObjectColumnsAgg[docType], false, "", 0)
		return fmt.Sprintf(BingAdsDocumentToQuery[docType], columns, bigQueryProjectId, schemaId, GetBingAdsDocumentFilterCondition(docType, false, "", executionDate))
	}
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
