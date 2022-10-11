package memsql

import (
	"database/sql"
	"errors"
	C "factors/config"
	"factors/model/model"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	U "factors/util"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	filterValueAll                                  = "all"
	fromSmartProperty                               = " FROM smart_properties "
	adwordsDocuments                                = "adwords_documents"
	adwordsSelectQueryForSmartPropertyStr           = "SELECT project_id, customer_account_id, value, timestamp, campaign_id, ad_group_id, keyword_id "
	smartPropertyCampaignStaticFilter               = " campaign.object_type = 1 "
	smartPropertyAdGroupStaticFilter                = " ad_group.object_type = 2 "
	staticWhereStatementForAdwordsWithSmartProperty = "WHERE adwords_documents.project_id = ? AND adwords_documents.customer_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "
	lastSyncInfoQueryForAllProjects                 = "SELECT project_id, customer_account_id, type as document_type, max(timestamp) as last_timestamp" +
		" " + "FROM adwords_documents GROUP BY project_id, customer_account_id, type"
	lastSyncInfoForAProject = "SELECT project_id, customer_account_id, type as document_type, max(timestamp) as last_timestamp" +
		" " + "FROM adwords_documents WHERE project_id = ? GROUP BY project_id, customer_account_id, type"
	insertAdwordsStr               = "INSERT INTO adwords_documents (project_id,customer_account_id,type,timestamp,id,campaign_id,ad_group_id,ad_id,keyword_id,value,created_at,updated_at) VALUES "
	adwordsFilterQueryStr          = "SELECT DISTINCT(LCASE(JSON_EXTRACT_STRING(value, ?))) as filter_value FROM adwords_documents WHERE project_id = ? AND" + " " + "customer_account_id IN (?) AND type = ? AND JSON_EXTRACT_STRING(value, ?) IS NOT NULL AND timestamp BETWEEN ? AND ? LIMIT 5000"
	staticWhereStatementForAdwords = "WHERE project_id = ? AND customer_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "
	fromAdwordsDocument            = " FROM adwords_documents "

	shareHigherOrderExpression              = "sum(case when JSON_EXTRACT_STRING(value, '%s') IS NOT NULL THEN (JSON_EXTRACT_STRING(value, '%s')) else 0 END)/NULLIF(sum(case when JSON_EXTRACT_STRING(value, '%s') IS NOT NULL THEN (JSON_EXTRACT_STRING(value, '%s')) else 0 END), 0)"
	shareHigherOrderExpressionWithZeroCheck = "sum(case when JSON_EXTRACT_STRING(value, '%s') IS NOT NULL AND JSON_EXTRACT_STRING(value, '%s') != '0.0' THEN (JSON_EXTRACT_STRING(value, '%s')) else 0 END)/NULLIF(sum(case when JSON_EXTRACT_STRING(value, '%s') IS NOT NULL THEN (JSON_EXTRACT_STRING(value, '%s')) else 0 END), 0)"
	higherOrderExpressionsWithMultiply      = "SUM(JSON_EXTRACT_STRING(value, '%s'))*%s/(COALESCE( NULLIF(sum(JSON_EXTRACT_STRING(value, '%s')), 0), 100000))"
	higherOrderExpressionsWithDiv           = "(SUM(JSON_EXTRACT_STRING(value, '%s'))/1000000)/(COALESCE( NULLIF(sum(JSON_EXTRACT_STRING(value, '%s')), 0), 100000))"

	sumOfFloatExp = "sum((JSON_EXTRACT_STRING(value, '%s')))"

	adwordsAdGroupMetadataFetchQueryStr = "select ad_group_information.ad_group_id, ad_group_information.campaign_id, ad_group_information.ad_group_name, ad_group_information.campaign_name " +
		"from ( " +
		"select ad_group_id, campaign_id, JSON_EXTRACT_STRING(value, 'name') as ad_group_name, JSON_EXTRACT_STRING(value, 'campaign_name') as campaign_name, timestamp " +
		"from adwords_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_account_id IN (?) " +
		") as ad_group_information " +
		"INNER JOIN " +
		"(select ad_group_id, max(timestamp) as timestamp " +
		"from adwords_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_account_id IN (?) group by ad_group_id " +
		") as ad_group_latest_timestamp_id " +
		"ON ad_group_information.ad_group_id = ad_group_latest_timestamp_id.ad_group_id AND ad_group_information.timestamp = ad_group_latest_timestamp_id.timestamp "

	adwordsCampaignMetadataFetchQueryStr = "select campaign_information.campaign_id, campaign_information.campaign_name " +
		"from ( " +
		"select campaign_id, JSON_EXTRACT_STRING(value, 'name') as campaign_name, timestamp " +
		"from adwords_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_account_id IN (?) " +
		") as campaign_information " +
		"INNER JOIN " +
		"(select campaign_id, max(timestamp) as timestamp " +
		"from adwords_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_account_id IN (?) group by campaign_id " +
		") as campaign_latest_timestamp_id " +
		"ON campaign_information.campaign_id = campaign_latest_timestamp_id.campaign_id AND campaign_information.timestamp = campaign_latest_timestamp_id.timestamp "

	adwordsTemplateWeeklyDifferenceAnalysisKeywordsQuerySelectStmnt = "Select %s %s from "
	fixedSelectForBreakdownAnalysisKeyword                          = "ABS((keyword_analysis_last_week.analysis_metric - keyword_analysis_previous_week.analysis_metric)) as abs_change, " +
		"(keyword_analysis_last_week.analysis_metric - keyword_analysis_previous_week.analysis_metric) as absolute_change, " +
		"keyword_analysis_previous_week.analysis_metric as previous_week_value, keyword_analysis_last_week.analysis_metric as last_week_value "
	adwordsTemplateWeeklyDifferenceAnalysisKeywordsQueryJoinStmnt = "keyword_analysis_last_week full outer join keyword_analysis_previous_week on keyword_analysis_last_week.keyword_id = " +
		"keyword_analysis_previous_week.keyword_id and keyword_analysis_last_week.keyword_match_type=keyword_analysis_previous_week.keyword_match_type " +
		"and keyword_analysis_last_week.campaign_id = keyword_analysis_previous_week.campaign_id and keyword_analysis_last_week.keyword_name " +
		"= keyword_analysis_previous_week.keyword_name "
	adwordsTemplateWeeklyDifferenceAnalysisKeywordsQueryWhereStmnt = "where (ABS((((keyword_analysis_last_week.analysis_metric - " +
		"keyword_analysis_previous_week.analysis_metric))*100/(COALESCE(NULLIF(keyword_analysis_previous_week.analysis_metric, 0), 0.0000001)))) >= ? " +
		"AND ABS((keyword_analysis_last_week.analysis_metric - keyword_analysis_previous_week.analysis_metric)) > ?) or " +
		"keyword_analysis_previous_week.analysis_metric is null or keyword_analysis_last_week.analysis_metric is null order by abs_change DESC"

	adwordsTemplatesWeeklyKeywordAnalysisQuery = "select %s as analysis_metric, %s keyword_id, campaign_id, (JSON_EXTRACT_STRING(value,'criteria')) " +
		"as keyword_name, (JSON_EXTRACT_STRING(value,'keyword_match_type')) as keyword_match_type from adwords_documents " +
		"where project_id = ? and customer_account_id in (?) and type = ? and timestamp between ? AND ?  and LCASE(JSON_EXTRACT_STRING(value,'criteria')) " +
		"!= 'automaticcontent' and LCASE(JSON_EXTRACT_STRING(value,'criteria')) != 'automatickeywords'" +
		" group by campaign_id, keyword_id, keyword_name, keyword_match_type"

	adwordsTemplatesWeeklyCampaignAnalysisQuery = "select %s as analysis_metric, " +
		"campaign_id, (JSON_EXTRACT_STRING(value,'campaign_name')) as campaign_name from adwords_documents " +
		"where project_id = ? and customer_account_id in (?) and type = ? and timestamp between ? AND ? and campaign_id in " +
		"(?) and JSON_EXTRACT_STRING(value,'advertising_channel_type') RLIKE 'search' group by campaign_id, campaign_name"
	adwordsTemplatesWeeklyDifferenceCampaignAnalysisQueryQuery = " Select coalesce(campaign_analysis_last_week.campaign_name, campaign_analysis_previous_week.campaign_name) as campaign_name, " +
		"campaign_analysis_previous_week.analysis_metric as previous_week_value, campaign_analysis_last_week.analysis_metric as last_week_value, " +
		"(((campaign_analysis_last_week.analysis_metric - campaign_analysis_previous_week.analysis_metric))*100/(COALESCE(NULLIF(campaign_analysis_previous_week.analysis_metric, 0), 0.0000001))) as percentage_change, " +
		"ABS((campaign_analysis_last_week.analysis_metric - campaign_analysis_previous_week.analysis_metric)) as abs_change, " +
		"(campaign_analysis_last_week.analysis_metric - campaign_analysis_previous_week.analysis_metric) as absolute_change, " +
		"coalesce(campaign_analysis_last_week.campaign_id, campaign_analysis_previous_week.campaign_id) as campaign_id from campaign_analysis_last_week " +
		"full outer join campaign_analysis_previous_week on campaign_analysis_last_week.campaign_id = campaign_analysis_previous_week.campaign_id " +
		"order by abs_change DESC limit 10000"

	semChecklistOverallAnalysisQuery = "select %s from adwords_documents " +
		"where project_id = ? and customer_account_id in (?) and type = ? and timestamp between ? AND ? AND LCASE(JSON_EXTRACT_STRING(value,'advertising_channel_type')) = 'search'"
	adwordsTemplatesExtraSelectWeekAnalysisForRCA      = "%s as impressions, %s as search_impression_share, %s as conversion_rate, %s as click_through_rate, %s as cost_per_click, "
	adwordsTemplatesExtraSelectBreakdownAnalysisForRCA = "%s as impressions, %s as search_impression_share, %s as conversion_rate, %s as click_through_rate, %s as cost_per_click, " +
		"%s as prev_impressions, %s as prev_search_impression_share, %s as prev_conversion_rate,%s as prev_click_through_rate, %s as prev_cost_per_click, " +
		"%s as last_impressions, %s as last_search_impression_share, %s as last_conversion_rate,%s as last_click_through_rate, %s as last_cost_per_click, "
	percentageChangeForSemChecklist = "(((%s.%s - %s.%s))*100/(COALESCE(NULLIF(%s.%s, 0), 0.0000001)))"
	coalesceForChecklists           = "COALESCE(%s.%s, %s.%s) as %s"
)

var templateMetricsToSelectStatement = map[string]string{
	model.Clicks:                "sum(JSON_EXTRACT_STRING(value, 'clicks'))",
	model.Impressions:           "sum(JSON_EXTRACT_STRING(value, 'impressions'))",
	model.ClickThroughRate:      fmt.Sprintf(higherOrderExpressionsWithMultiply, "clicks", "100", "impressions"),
	model.CostPerClick:          fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "clicks"),
	model.SearchImpressionShare: fmt.Sprintf(shareHigherOrderExpression, model.SearchImpressionShare, model.Impressions, model.SearchImpressionShare, model.TotalSearchImpression),
	"cost":                      "sum(JSON_EXTRACT_STRING(value, 'cost'))/1000000",
	model.Conversion:            "sum(JSON_EXTRACT_STRING(value, 'conversions'))",
	"cost_per_lead":             fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions"),
	model.ConversionRate:        fmt.Sprintf(higherOrderExpressionsWithMultiply, "conversions", "100", "clicks"),
}

var templateMetricsToSelectStatementForOverallAnalysis = map[string]string{
	model.Clicks:                "sum(JSON_EXTRACT_STRING(value, 'clicks')) as clicks, sum(JSON_EXTRACT_STRING(value, 'conversions')) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	model.Impressions:           "sum(JSON_EXTRACT_STRING(value, 'impressions')) as impressions, sum(JSON_EXTRACT_STRING(value, 'conversions')) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	model.ClickThroughRate:      fmt.Sprintf(higherOrderExpressionsWithMultiply, "clicks", "100", "impressions") + fmt.Sprintf("as %s, ", model.ClickThroughRate) + "sum(JSON_EXTRACT_STRING(value, 'conversions')) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	model.CostPerClick:          fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "clicks") + fmt.Sprintf("as %s, ", model.CostPerClick) + "sum(JSON_EXTRACT_STRING(value, 'conversions')) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	model.SearchImpressionShare: fmt.Sprintf(shareHigherOrderExpression, model.SearchImpressionShare, model.Impressions, model.SearchImpressionShare, model.TotalSearchImpression) + fmt.Sprintf("as %s, ", model.SearchImpressionShare) + "sum(JSON_EXTRACT_STRING(value, 'conversions')) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	"cost":                      "sum(JSON_EXTRACT_STRING(value, 'cost'))/1000000 as cost, sum(JSON_EXTRACT_STRING(value, 'conversions')) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	model.Conversion:            "sum(JSON_EXTRACT_STRING(value, 'conversions')) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	"cost_per_lead":             fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + fmt.Sprintf(" as %s, ", "cost_per_lead") + "sum(JSON_EXTRACT_STRING(value, 'conversions')) as conversion ",
	model.ConversionRate:        fmt.Sprintf(higherOrderExpressionsWithMultiply, "conversions", "100", "clicks") + fmt.Sprintf("as %s, ", model.ConversionRate) + "sum(JSON_EXTRACT_STRING(value, 'conversions')) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
}

var errorEmptyAdwordsDocument = errors.New("empty adwords document")

var propertiesToBeDividedByMillion = map[string]struct{}{
	model.TopOfPageCpc:     {},
	model.FirstPositionCpc: {},
	model.FirstPageCpc:     {},
}

type metricsAndRelated struct {
	higherOrderExpression    string
	nonHigherOrderExpression string
	externalValue            string
	externalOperation        string // This is not clear. What happens when ctr at higher level is presented?
}

var nonHigherOrderMetrics = map[string]struct{}{
	model.Impressions:     {},
	model.Clicks:          {},
	"cost":                {},
	"conversions":         {},
	model.ConversionValue: {},
}

// Same structure is being used for internal operations and external.
var adwordsInternalMetricsToAllRep = map[string]metricsAndRelated{
	model.Impressions: {
		nonHigherOrderExpression: "sum(JSON_EXTRACT_STRING(value, 'impressions'))",
		externalValue:            model.Impressions,
		externalOperation:        "sum",
	},
	model.Clicks: {
		nonHigherOrderExpression: "sum(JSON_EXTRACT_STRING(value, 'clicks'))",
		externalValue:            model.Clicks,
		externalOperation:        "sum",
	},
	"cost": {
		nonHigherOrderExpression: "sum(JSON_EXTRACT_STRING(value, 'cost'))/1000000",
		externalValue:            "spend",
		externalOperation:        "sum",
	},
	"conversions": {
		nonHigherOrderExpression: "sum(JSON_EXTRACT_STRING(value, 'conversions'))",
		externalValue:            model.Conversion,
		externalOperation:        "sum",
	},
	model.ClickThroughRate: {
		higherOrderExpression:    fmt.Sprintf(higherOrderExpressionsWithMultiply, "clicks", "100", "impressions"),
		nonHigherOrderExpression: "sum(JSON_EXTRACT_STRING(value, 'clicks'))*100",
		externalValue:            model.ClickThroughRate,
		externalOperation:        "sum",
	},
	model.ConversionRate: {
		higherOrderExpression:    fmt.Sprintf(higherOrderExpressionsWithMultiply, "conversions", "100", "clicks"),
		nonHigherOrderExpression: "sum(JSON_EXTRACT_STRING(value, 'conversions'))*100",
		externalValue:            model.ConversionRate,
		externalOperation:        "sum",
	},
	model.CostPerClick: {
		higherOrderExpression:    fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "clicks"),
		nonHigherOrderExpression: "(sum((value, 'cost'))/1000000)",
		externalValue:            model.CostPerClick,
		externalOperation:        "sum",
	},
	model.CostPerConversion: {
		higherOrderExpression:    fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions"),
		nonHigherOrderExpression: "(sum(JSON_EXTRACT_STRING(value, 'cost'))/1000000)",
		externalValue:            model.CostPerConversion,
		externalOperation:        "sum",
	},
	model.SearchImpressionShare: {
		// For campaign_type != 'Search', in adwords api Si share is Null but in ads api it's 0.0, excluding those cases here in Case When with shareHigherOrderExpressionWithZeroCheck
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpressionWithZeroCheck, model.SearchImpressionShare, model.SearchImpressionShare, model.Impressions, model.SearchImpressionShare, model.TotalSearchImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchImpression),
		externalValue:            model.SearchImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchClickShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpressionWithZeroCheck, model.SearchClickShare, model.SearchClickShare, model.Impressions, model.SearchClickShare, model.TotalSearchClick),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchClick),
		externalValue:            model.SearchClickShare,
		externalOperation:        "sum",
	},
	model.SearchTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpressionWithZeroCheck, model.SearchTopImpressionShare, model.SearchTopImpressionShare, model.TopImpressions, model.SearchTopImpressionShare, model.TotalTopImpressions),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchTopImpression),
		externalValue:            model.SearchTopImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchAbsoluteTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpressionWithZeroCheck, model.SearchAbsoluteTopImpressionShare, model.SearchAbsoluteTopImpressionShare, model.AbsoluteTopImpressions, model.SearchAbsoluteTopImpressionShare, model.TotalTopImpressions),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchAbsoluteTopImpression),
		externalValue:            model.SearchAbsoluteTopImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchBudgetLostAbsoluteTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpressionWithZeroCheck, model.SearchBudgetLostAbsoluteTopImpressionShare, model.SearchBudgetLostAbsoluteTopImpressionShare, model.AbsoluteTopImpressionLostDueToBudget, model.SearchBudgetLostAbsoluteTopImpressionShare, model.TotalTopImpressions),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchBudgetLostAbsoluteTopImpression),
		externalValue:            model.SearchBudgetLostAbsoluteTopImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchBudgetLostImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpressionWithZeroCheck, model.SearchBudgetLostImpressionShare, model.SearchBudgetLostImpressionShare, model.ImpressionLostDueToBudget, model.SearchBudgetLostImpressionShare, model.TotalSearchImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchBudgetLostImpression),
		externalValue:            model.SearchBudgetLostImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchBudgetLostTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpressionWithZeroCheck, model.SearchBudgetLostTopImpressionShare, model.SearchBudgetLostTopImpressionShare, model.TopImpressionLostDueToBudget, model.SearchBudgetLostTopImpressionShare, model.TotalTopImpressions),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchBudgetLostTopImpression),
		externalValue:            model.SearchBudgetLostTopImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchRankLostAbsoluteTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpressionWithZeroCheck, model.SearchRankLostAbsoluteTopImpressionShare, model.SearchRankLostAbsoluteTopImpressionShare, model.AbsoluteTopImpressionLostDueToRank, model.SearchRankLostAbsoluteTopImpressionShare, model.TotalTopImpressions),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchRankLostAbsoluteTopImpression),
		externalValue:            model.SearchRankLostAbsoluteTopImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchRankLostImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpressionWithZeroCheck, model.SearchRankLostImpressionShare, model.SearchRankLostImpressionShare, model.ImpressionLostDueToRank, model.SearchRankLostImpressionShare, model.TotalSearchImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchRankLostImpression),
		externalValue:            model.SearchRankLostImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchRankLostTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpressionWithZeroCheck, model.SearchRankLostTopImpressionShare, model.SearchRankLostTopImpressionShare, model.TopImpressionLostDueToRank, model.SearchRankLostTopImpressionShare, model.TotalTopImpressions),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchRankLostTopImpression),
		externalValue:            model.SearchRankLostTopImpressionShare,
		externalOperation:        "sum",
	},
	model.ConversionValue: {
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.ConversionValue),
		externalValue:            model.ConversionValue,
		externalOperation:        "sum",
	},
}

type fields struct {
	selectExpressions []string
	values            []string
}

func (store *MemSQL) satisfiesAdwordsDocumentForeignConstraints(adwordsDocument model.AdwordsDocument) int {
	logFields := log.Fields{"adwords_document": adwordsDocument}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, errCode := store.GetProject(adwordsDocument.ProjectID)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}

	return http.StatusOK
}

func (store *MemSQL) satisfiesAdwordsDocumentUniquenessConstraints(adwordsDocument *model.AdwordsDocument) int {
	logFields := log.Fields{"adwords_document": adwordsDocument}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	errCode := store.isAdwordsDocumentExistByPrimaryKey(adwordsDocument)
	if errCode == http.StatusFound {
		return http.StatusConflict
	}
	if errCode == http.StatusNotFound {
		return http.StatusOK
	}
	return errCode
}

// Checks PRIMARY KEY constraint (project_id, customer_account_id, type, timestamp, id)
func (store *MemSQL) isAdwordsDocumentExistByPrimaryKey(document *model.AdwordsDocument) int {
	logFields := log.Fields{"document": document}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if document.ProjectID == 0 || document.CustomerAccountID == "" || document.Type == 0 ||
		document.Timestamp == 0 || document.ID == "" {

		log.Error("Invalid document on primary constraint check.")
		return http.StatusBadRequest
	}

	var adwordsDocument model.AdwordsDocument

	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ? AND customer_account_id = ? AND type = ? AND timestamp = ? AND id = ?",
		document.ProjectID, document.CustomerAccountID, document.Type, document.Timestamp, document.ID,
	).Select("id").Find(&adwordsDocument).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}

		logCtx.WithError(err).
			Error("Failed getting to check existence adwords document by primary keys.")
		return http.StatusInternalServerError
	}

	if adwordsDocument.ID == "" {
		logCtx.Error("Invalid id value returned on adwords document primary key check.")
		return http.StatusInternalServerError
	}

	return http.StatusFound
}

func getAdwordsIDFieldNameByType(docType int) string {
	logFields := log.Fields{"doc_type": docType}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	switch docType {
	case 4: // click_performance_report
		return "gcl_id"
	case 5: // campaign_performance_report
		return "campaign_id"
	case 7: // search_performance_report
		return "query"
	case 9: // customer_account_properties
		return "customer_id"
	case 10: // ad_group_performance_report
		return "ad_group_id"
	default: // others
		return "id"
	}
}

// Returns campaign_id, ad_group_id, ad_id, keyword_id
func getAdwordsHierarchyColumnsByType(valueMap *map[string]interface{}, docType int) (int64, int64, int64, int64) {
	logFields := log.Fields{
		"value_map": valueMap,
		"doc_type":  docType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	switch docType {
	case model.AdwordsDocumentTypeAlias["campaigns"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "id", 0), 0, 0, 0
	case model.AdwordsDocumentTypeAlias["ad_groups"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "id", 0), 0, 0
	case model.AdwordsDocumentTypeAlias["click_performance_report"], model.AdwordsDocumentTypeAlias["search_performance_report"], model.AdwordsDocumentTypeAlias["ad_group_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "ad_group_id", 0), 0, 0
	case model.AdwordsDocumentTypeAlias["campaign_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), 0, 0, 0
	case model.AdwordsDocumentTypeAlias["ad_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "ad_group_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "id", 0), 0
	case model.AdwordsDocumentTypeAlias["keyword_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "ad_group_id", 0), 0, U.GetInt64FromMapOfInterface(*valueMap, "id", 0)
	case model.AdwordsDocumentTypeAlias["customer_account_properties"]:
		return 0, 0, 0, 0
	default:
		return 0, 0, 0, 0
	}
}

// GetAdwordsDateOnlyTimestamp - Date only timestamp to query adwords documents.
func GetAdwordsDateOnlyTimestamp(unixTimestamp int64) string {
	logFields := log.Fields{"unix_timestamp": unixTimestamp}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// Todo: Add timezone support using util.getTimeFromUnixTimestampWithZone.
	return time.Unix(unixTimestamp, 0).UTC().Format("20060102")
}

func getAdwordsIDAndHeirarchyColumnsByType(docType int, valueJSON *postgres.Jsonb) (string, int64, int64, int64, int64, error) {
	logFields := log.Fields{
		"doc_type":  docType,
		"valueJSON": valueJSON,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if docType > len(model.AdwordsDocumentTypeAlias) {
		return "", 0, 0, 0, 0, errors.New("invalid document type")
	}

	valueMap, err := U.DecodePostgresJsonb(valueJSON)
	if err != nil {
		return "", 0, 0, 0, 0, err
	}

	if len(*valueMap) == 0 {
		return "", 0, 0, 0, 0, errorEmptyAdwordsDocument
	}
	idFieldName := getAdwordsIDFieldNameByType(docType)
	id, exists := (*valueMap)[idFieldName]
	if !exists {
		return "", 0, 0, 0, 0, fmt.Errorf("id field %s does not exist on doc of type %v", idFieldName, docType)
	}

	if id == nil {
		return "", 0, 0, 0, 0, fmt.Errorf("id field %s has empty value on doc of type %v", idFieldName, docType)
	}
	idStr, err := U.GetValueAsString(id)
	if err != nil {
		return "", 0, 0, 0, 0, err
	}
	if docType == model.AdwordsDocumentTypeAlias[model.KeywordPerformanceReport] || docType == model.AdwordsDocumentTypeAlias[model.SearchPerformanceReport] {
		idStr = U.GetUUID()
	}

	value1, value2, value3, value4 := getAdwordsHierarchyColumnsByType(valueMap, docType)

	// ID as string always.
	return idStr, value1, value2, value3, value4, nil
}

// CreateAdwordsDocument ...
func (store *MemSQL) CreateAdwordsDocument(adwordsDoc *model.AdwordsDocument) int {
	logFields := log.Fields{
		"adwords_doc": adwordsDoc,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	status := validateAdwordsDocument(adwordsDoc)
	if status != http.StatusOK {
		return status
	}

	status = addColumnInformationForAdwordsDocument(adwordsDoc)
	if status != http.StatusOK {
		return status
	} else if errCode := store.satisfiesAdwordsDocumentForeignConstraints(*adwordsDoc); errCode != http.StatusOK {
		return http.StatusInternalServerError
	}

	errCode := store.satisfiesAdwordsDocumentUniquenessConstraints(adwordsDoc)
	if errCode != http.StatusOK {
		return errCode
	}

	db := C.GetServices().Db
	dbc := db.Create(adwordsDoc)
	if dbc.Error != nil {
		if IsDuplicateRecordError(dbc.Error) {
			log.WithError(dbc.Error).WithField("adwordsDocuments", adwordsDoc).Error("Failed to create an adwords doc. Duplicate.")
			return http.StatusConflict
		}
		log.WithError(dbc.Error).WithField("adwordsDocuments", adwordsDoc).Error(
			"Failed to create an adwords doc.")
		return http.StatusInternalServerError
	}
	UpdateCountCacheByDocumentType(adwordsDoc.ProjectID, &adwordsDoc.CreatedAt, "adwords")
	return http.StatusCreated
}

// CreateMultipleAdwordsDocument ...
func (store *MemSQL) CreateMultipleAdwordsDocument(adwordsDocuments []model.AdwordsDocument) int {
	logFields := log.Fields{"adwords_documents": adwordsDocuments}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	status := validateAdwordsDocuments(adwordsDocuments)
	if status != http.StatusOK {
		return status
	}
	adwordsDocuments, status = addColumnInformationForAdwordsDocuments(adwordsDocuments)
	if status != http.StatusOK {
		return status
	}

	// uniqueDocuments := make([]model.AdwordsDocument, 0)
	// for _, document := range adwordsDocuments {
	// 	statusCode := store.satisfiesAdwordsDocumentUniquenessConstraints(&document)
	// 	if statusCode != http.StatusOK {
	// 		status = statusCode
	// 		log.WithField("adwordsDocuments", document).Error("Failed to create an adwords doc. Duplicate.")
	// 	} else {
	// 		uniqueDocuments = append(uniqueDocuments, document)
	// 	}
	// }
	// adwordsDocuments = uniqueDocuments
	db := C.GetServices().Db

	insertStatement := insertAdwordsStr
	insertValuesStatement := make([]string, 0, 0)
	insertValues := make([]interface{}, 0, 0)
	for _, adwordsDoc := range adwordsDocuments {
		insertValuesStatement = append(insertValuesStatement, fmt.Sprintf("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"))
		insertValues = append(insertValues, adwordsDoc.ProjectID, adwordsDoc.CustomerAccountID,
			adwordsDoc.Type, adwordsDoc.Timestamp, adwordsDoc.ID, adwordsDoc.CampaignID, adwordsDoc.AdGroupID, adwordsDoc.AdID, adwordsDoc.KeywordID, adwordsDoc.Value, adwordsDoc.CreatedAt, adwordsDoc.UpdatedAt)
		UpdateCountCacheByDocumentType(adwordsDoc.ProjectID, &adwordsDoc.CreatedAt, "adwords")
	}
	insertStatement += joinWithComma(insertValuesStatement...)
	rows, err := db.Raw(insertStatement, insertValues...).Rows()

	if err != nil {
		if IsDuplicateRecordError(err) {
			log.WithError(err).WithField("adwordsDocuments", adwordsDocuments).Error("Failed to create an adwords doc. Duplicate.")
			return http.StatusConflict
		} else {
			log.WithError(err).WithField("adwordsDocuments", adwordsDocuments).Error(
				"Failed to create an adwords doc. Continued inserting other docs.")
			return http.StatusInternalServerError
		}
	}
	defer rows.Close()

	if status != http.StatusOK {
		return status
	}
	return http.StatusCreated
}

func validateAdwordsDocuments(adwordsDocuments []model.AdwordsDocument) int {
	logFields := log.Fields{"adwords_documents": adwordsDocuments}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for index, _ := range adwordsDocuments {
		status := validateAdwordsDocument(&adwordsDocuments[index])
		if status != http.StatusOK {
			log.WithField("index", index).Error("Failed in this index")
			return status
		}
	}
	return http.StatusOK
}

func validateAdwordsDocument(adwordsDocument *model.AdwordsDocument) int {
	logFields := log.Fields{"adwords_document": adwordsDocument}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithField("customer_acc_id", adwordsDocument.CustomerAccountID).WithField(
		"project_id", adwordsDocument.ProjectID)

	if adwordsDocument.CustomerAccountID == "" || adwordsDocument.TypeAlias == "" {
		logCtx.Error("Invalid adwords document.")
		return http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", adwordsDocument.TypeAlias)
	docType, docTypeExists := model.AdwordsDocumentTypeAlias[adwordsDocument.TypeAlias]
	if !docTypeExists {
		logCtx.Error("Invalid type alias.")
		return http.StatusBadRequest
	}
	adwordsDocument.Type = docType
	return http.StatusOK
}

// Assigning id, campaignId columns with values from json...
func addColumnInformationForAdwordsDocuments(adwordsDocuments []model.AdwordsDocument) ([]model.AdwordsDocument, int) {
	logFields := log.Fields{"adwords_documents": adwordsDocuments}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for index, _ := range adwordsDocuments {
		status := addColumnInformationForAdwordsDocument(&adwordsDocuments[index])
		if status != http.StatusOK {
			log.WithField("index", index).Error("Failed in this index")
			return adwordsDocuments, status
		}
	}
	return adwordsDocuments, http.StatusOK
}

func addColumnInformationForAdwordsDocument(adwordsDocument *model.AdwordsDocument) int {
	logFields := log.Fields{"adwords_document": adwordsDocument}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	adwordsDocID, campaignIDValue, adGroupIDValue, adIDValue,
		keywordIDValue, err := getAdwordsIDAndHeirarchyColumnsByType(adwordsDocument.Type, adwordsDocument.Value)
	if err != nil {
		if err == errorEmptyAdwordsDocument {
			// Using UUID to allow storing empty response. To avoid downloading reports again for the same timerange.
			adwordsDocID = U.GetUUID()
		} else {
			logCtx.WithError(err).Error("Failed to get id by adowords doc type.")
			return http.StatusInternalServerError
		}
	}

	currentTime := gorm.NowFunc()
	adwordsDocument.ID = adwordsDocID
	adwordsDocument.CampaignID = campaignIDValue
	adwordsDocument.AdGroupID = adGroupIDValue
	adwordsDocument.AdID = adIDValue
	adwordsDocument.KeywordID = keywordIDValue
	adwordsDocument.CreatedAt = currentTime
	adwordsDocument.UpdatedAt = currentTime

	return http.StatusOK
}

func getDocumentTypeAliasByType() map[int]string {

	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range model.AdwordsDocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

func (store *MemSQL) GetAdwordsLastSyncInfoForProject(projectID int64) ([]model.AdwordsLastSyncInfo, int) {
	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	params := []interface{}{projectID}
	adwordsLastSyncInfos, status := getAdwordsLastSyncInfo(lastSyncInfoForAProject, params)
	if status != http.StatusOK {
		return adwordsLastSyncInfos, status
	}
	adwordsSettings, errCode := store.GetIntAdwordsProjectSettingsForProjectID(projectID)
	if errCode != http.StatusOK {
		return []model.AdwordsLastSyncInfo{}, errCode
	}

	return store.sanitizedLastSyncInfos(adwordsLastSyncInfos, adwordsSettings)
}

// GetAllAdwordsLastSyncInfoForAllProjects - @TODO Kark v1
func (store *MemSQL) GetAllAdwordsLastSyncInfoForAllProjects() ([]model.AdwordsLastSyncInfo, int) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	params := make([]interface{}, 0, 0)
	adwordsLastSyncInfos, status := getAdwordsLastSyncInfo(lastSyncInfoQueryForAllProjects, params)
	if status != http.StatusOK {
		return adwordsLastSyncInfos, status
	}

	adwordsSettings, errCode := store.GetAllIntAdwordsProjectSettings()
	if errCode != http.StatusOK {
		return []model.AdwordsLastSyncInfo{}, errCode
	}

	return store.sanitizedLastSyncInfos(adwordsLastSyncInfos, adwordsSettings)
}

func getAdwordsLastSyncInfo(query string, params []interface{}) ([]model.AdwordsLastSyncInfo, int) {
	logFields := log.Fields{
		"query":  query,
		"params": params,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	adwordsLastSyncInfos := make([]model.AdwordsLastSyncInfo, 0, 0)

	rows, err := db.Raw(query, params).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last adwords documents by type for sync info.")
		return adwordsLastSyncInfos, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var adwordsLastSyncInfo model.AdwordsLastSyncInfo
		if err := db.ScanRows(rows, &adwordsLastSyncInfo); err != nil {
			log.WithError(err).Error("Failed to scan last adwords documents by type for sync info.")
			return []model.AdwordsLastSyncInfo{}, http.StatusInternalServerError
		}

		adwordsLastSyncInfos = append(adwordsLastSyncInfos, adwordsLastSyncInfo)
	}

	return adwordsLastSyncInfos, http.StatusOK
}

// This method handles adding additionalInformation to lastSyncInfo, Skipping inactive Projects and adding missed LastSync.
func (store *MemSQL) sanitizedLastSyncInfos(adwordsLastSyncInfos []model.AdwordsLastSyncInfo, adwordsSettings []model.AdwordsProjectSettings) ([]model.AdwordsLastSyncInfo, int) {
	logFields := log.Fields{
		"adwords_last_sync_infos": adwordsLastSyncInfos,
		"adword_settings":         adwordsSettings,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	adwordsSettingsByProjectAndCustomerAccount := make(map[int64]map[string]*model.AdwordsProjectSettings, 0)
	projectIDs := make([]int64, 0, 0)

	// Forming MapOfProjectToCustomerAccToManagerAcc
	projectToCustomerAccToManagerAccMap := make(map[int64]map[string]string)
	for i := range adwordsSettings {
		customerAccToManagerAccMap := make(map[string]string)
		if adwordsSettings[i].IntAdwordsClientManagerMap != nil {
			err := U.DecodePostgresJsonbToStructType(adwordsSettings[i].IntAdwordsClientManagerMap, &customerAccToManagerAccMap)
			if err != nil {
				log.Warn(err)
			} else {
				projectToCustomerAccToManagerAccMap[adwordsSettings[i].ProjectId] = customerAccToManagerAccMap
			}
		}

		customerAccountIDs := strings.Split(adwordsSettings[i].CustomerAccountId, ",")
		for j := range customerAccountIDs {
			_, isProjectExists := projectToCustomerAccToManagerAccMap[adwordsSettings[i].ProjectId]
			if !isProjectExists {
				projectToCustomerAccToManagerAccMap[adwordsSettings[i].ProjectId] = make(map[string]string)
			}
			if _, customerAccountExists := projectToCustomerAccToManagerAccMap[adwordsSettings[i].ProjectId][customerAccountIDs[j]]; !customerAccountExists {
				projectToCustomerAccToManagerAccMap[adwordsSettings[i].ProjectId][customerAccountIDs[j]] = ""
			}
		}
	}

	// Forming the MapOfProjectIdCustomerAccountToData.
	for i := range adwordsSettings {
		customerAccountIDs := strings.Split(adwordsSettings[i].CustomerAccountId, ",")
		adwordsSettingsByProjectAndCustomerAccount[adwordsSettings[i].ProjectId] = make(map[string]*model.AdwordsProjectSettings)
		projectIDs = append(projectIDs, adwordsSettings[i].ProjectId)
		for j := range customerAccountIDs {
			var setting model.AdwordsProjectSettings
			setting.ProjectId = adwordsSettings[i].ProjectId
			setting.AgentUUID = adwordsSettings[i].AgentUUID
			setting.RefreshToken = adwordsSettings[i].RefreshToken
			setting.IntGoogleIngestionTimezone = adwordsSettings[i].IntGoogleIngestionTimezone
			setting.CustomerAccountId = customerAccountIDs[j]
			adwordsSettingsByProjectAndCustomerAccount[adwordsSettings[i].ProjectId][customerAccountIDs[j]] = &setting
		}
	}
	documentTypeAliasByType := getDocumentTypeAliasByType()

	// add settings for project_id existing on adwords documents.
	existingProjectAndCustomerAccountWithTypes := make(map[int64]map[string]map[string]bool, 0)
	selectedLastSyncInfos := make([]model.AdwordsLastSyncInfo, 0, 0)

	for i := range adwordsLastSyncInfos {
		logCtx := log.WithFields(
			log.Fields{"project_id": adwordsLastSyncInfos[i].ProjectId,
				"customer_account_id": adwordsLastSyncInfos[i].CustomerAccountId})

		projectToCustomerAccMap, isExists := projectToCustomerAccToManagerAccMap[adwordsLastSyncInfos[i].ProjectId]
		if !isExists {
			projectToCustomerAccMap = make(map[string]string)
		}

		settings, exists := adwordsSettingsByProjectAndCustomerAccount[adwordsLastSyncInfos[i].ProjectId][adwordsLastSyncInfos[i].CustomerAccountId]
		if !exists {
			logCtx.Warn("Adwords project settings not found for customer account adwords synced earlier.")
		}

		if settings == nil {
			logCtx.Info("Adwords disabled for project.")
			continue
		}

		typeAlias, typeAliasExists := documentTypeAliasByType[adwordsLastSyncInfos[i].DocumentType]
		if !typeAliasExists {
			logCtx.WithField("document_type",
				adwordsLastSyncInfos[i].DocumentType).Error("Invalid document type given. No type alias name.")
			continue
		}

		adwordsLastSyncInfos[i].DocumentTypeAlias = typeAlias // map the type to type alias name.
		adwordsLastSyncInfos[i].RefreshToken = settings.RefreshToken
		adwordsLastSyncInfos[i].Timezone = settings.IntGoogleIngestionTimezone
		managerID, isExists := projectToCustomerAccMap[adwordsLastSyncInfos[i].CustomerAccountId]
		if isExists {
			adwordsLastSyncInfos[i].ManagerID = managerID
		} else {
			adwordsLastSyncInfos[i].ManagerID = ""
		}

		selectedLastSyncInfos = append(selectedLastSyncInfos, adwordsLastSyncInfos[i])

		if _, projectWithCustomerAccountExists := existingProjectAndCustomerAccountWithTypes[adwordsLastSyncInfos[i].ProjectId][adwordsLastSyncInfos[i].CustomerAccountId]; !projectWithCustomerAccountExists {
			if _, projectExists := existingProjectAndCustomerAccountWithTypes[adwordsLastSyncInfos[i].ProjectId]; !projectExists {
				existingProjectAndCustomerAccountWithTypes[adwordsLastSyncInfos[i].ProjectId] = make(map[string]map[string]bool, 0)
			}
			existingProjectAndCustomerAccountWithTypes[adwordsLastSyncInfos[i].ProjectId][adwordsLastSyncInfos[i].CustomerAccountId] = make(map[string]bool, 0)
		}

		existingProjectAndCustomerAccountWithTypes[adwordsLastSyncInfos[i].ProjectId][adwordsLastSyncInfos[i].CustomerAccountId][adwordsLastSyncInfos[i].DocumentTypeAlias] = true
	}

	// add all types for missing projects and
	// add missing types for existing projects.
	for i := range adwordsSettings {
		customerAccountIDs := strings.Split(adwordsSettings[i].CustomerAccountId, ",")
		for _, accountID := range customerAccountIDs {
			existingTypesForAccount, accountExists := existingProjectAndCustomerAccountWithTypes[adwordsSettings[i].ProjectId][accountID]
			for docTypeAlias := range model.AdwordsDocumentTypeAlias {
				if !accountExists || (accountExists && !existingTypesForAccount[docTypeAlias]) {
					syncInfo := model.AdwordsLastSyncInfo{
						ProjectId:         adwordsSettings[i].ProjectId,
						RefreshToken:      adwordsSettings[i].RefreshToken,
						CustomerAccountId: accountID,
						LastTimestamp:     0, // no sync yet.
						DocumentTypeAlias: docTypeAlias,
						Timezone:          adwordsSettings[i].IntGoogleIngestionTimezone,
						ManagerID:         projectToCustomerAccToManagerAccMap[adwordsSettings[i].ProjectId][accountID],
					}

					selectedLastSyncInfos = append(selectedLastSyncInfos, syncInfo)
				}
			}
		}

	}

	projects, _ := store.GetProjectsByIDs(projectIDs)
	for _, project := range projects {
		for index := range selectedLastSyncInfos {
			if selectedLastSyncInfos[index].ProjectId == project.ID {
				if selectedLastSyncInfos[index].Timezone == "" {
					selectedLastSyncInfos[index].Timezone = project.TimeZone
				}
			}
		}
	}

	return selectedLastSyncInfos, http.StatusOK
}

// PullGCLIDReport - It returns GCLID based campaign info ( Adgroup, Campaign and Ad) for given time range and adwords account
func (store *MemSQL) PullGCLIDReport(projectID int64, from, to int64, adwordsAccountIDs string,
	campaignIDReport, adgroupIDReport, keywordIDReport map[string]model.MarketingData, timeZone string) (map[string]model.MarketingData, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"from":       from,
		"to":         to,
		"method":     "PullGCLIDReport",
		"time_zone":  timeZone,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	adGroupNameCase := "CASE WHEN JSON_EXTRACT_STRING(value, 'ad_group_name') IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(value, 'ad_group_name') = '' THEN ? ELSE JSON_EXTRACT_STRING(value, 'ad_group_name') END AS ad_group_name"
	adGroupIDCase := "CASE WHEN JSON_EXTRACT_STRING(value, 'ad_group_id') IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(value, 'ad_group_id') = '' THEN ? ELSE JSON_EXTRACT_STRING(value, 'ad_group_id') END AS ad_group_id"
	campaignNameCase := "CASE WHEN JSON_EXTRACT_STRING(value, 'campaign_name')  IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(value, 'campaign_name') = '' THEN ? ELSE JSON_EXTRACT_STRING(value, 'campaign_name') END AS campaign_name"
	campaignIDCase := "CASE WHEN JSON_EXTRACT_STRING(value, 'campaign_id') IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(value, 'campaign_id') = '' THEN ? ELSE JSON_EXTRACT_STRING(value, 'campaign_id') END AS campaign_id"
	adIDCase := "CASE WHEN JSON_EXTRACT_STRING(value, 'creative_id') IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(value, 'creative_id') = '' THEN ? ELSE JSON_EXTRACT_STRING(value, 'creative_id') END AS creative_id"
	keywordNameCase := "CASE WHEN JSON_EXTRACT_STRING(value, 'criteria_name') IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(value, 'criteria_name') = '' THEN ? ELSE JSON_EXTRACT_STRING(value, 'criteria_name') END AS criteria_name"
	keywordIDCase := "CASE WHEN JSON_EXTRACT_STRING(value, 'criteria_id') IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(value, 'criteria_id') = '' THEN ? ELSE JSON_EXTRACT_STRING(value, 'criteria_id') END AS criteria_id"
	keywordNameCase2 := "CASE WHEN JSON_EXTRACT_STRING(value, 'keyword_name') IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(value, 'keyword_name') = '' THEN ? ELSE JSON_EXTRACT_STRING(value, 'keyword_name') END AS keyword_name"
	keywordIDCase2 := "CASE WHEN JSON_EXTRACT_STRING(value, 'keyword_id') IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(value, 'keyword_id') = '' THEN ? ELSE JSON_EXTRACT_STRING(value, 'keyword_id') END AS keyword_id"
	slotCase := "CASE WHEN JSON_EXTRACT_STRING(value, 'slot') IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(value, 'slot') = '' THEN ? ELSE JSON_EXTRACT_STRING(value, 'slot') END AS slot"

	performanceQuery := "SELECT id, " + adGroupNameCase + ", " + adGroupIDCase + ", " + campaignNameCase + ", " +
		campaignIDCase + ", " + adIDCase + ", " + keywordNameCase + ", " + keywordIDCase + ", " + keywordNameCase2 + ", " + keywordIDCase2 + ", " + slotCase +
		" FROM adwords_documents where project_id = ? AND customer_account_id IN (?) AND type = ? AND timestamp between ? AND ? "
	customerAccountIDs := strings.Split(adwordsAccountIDs, ",")
	rows, tx, err, reqID := store.ExecQueryWithContext(performanceQuery, []interface{}{model.PropertyValueNone, model.PropertyValueNone,
		model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone,
		model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone,
		model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone,
		model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone,
		model.PropertyValueNone, model.PropertyValueNone,
		projectID, customerAccountIDs, model.AdwordsClickReportType, U.GetDateAsStringIn(from, U.TimeZoneString(timeZone)),
		U.GetDateAsStringIn(to, U.TimeZoneString(timeZone))})
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, err
	}
	defer U.CloseReadQuery(rows, tx)
	gclidBasedMarketData := make(map[string]model.MarketingData)
	startReadTime := time.Now()
	for rows.Next() {
		var gclIDTmp sql.NullString
		var adgroupNameTmp sql.NullString
		var adgroupIDTmp sql.NullString
		var campaignNameTmp sql.NullString
		var campaignIDTmp sql.NullString
		var adIDTmp sql.NullString
		var keywordNameTmp sql.NullString
		var keywordIDTmp sql.NullString
		var keywordNameTmp2 sql.NullString
		var keywordIDTmp2 sql.NullString
		var slotTmp sql.NullString
		if err = rows.Scan(&gclIDTmp, &adgroupNameTmp, &adgroupIDTmp, &campaignNameTmp, &campaignIDTmp, &adIDTmp, &keywordNameTmp, &keywordIDTmp, &keywordNameTmp2, &keywordIDTmp2, &slotTmp); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
			continue
		}
		if !gclIDTmp.Valid {
			continue
		}
		var gclID string
		var adgroupName string
		var adgroupID string
		var campaignName string
		var campaignID string
		var adID string
		var keywordName string
		var keywordID string
		var keywordName2 string
		var keywordID2 string
		var slot string
		gclID = gclIDTmp.String
		adgroupName = U.IfThenElse(adgroupNameTmp.Valid == true, adgroupNameTmp.String, model.PropertyValueNone).(string)
		adgroupID = U.IfThenElse(adgroupIDTmp.Valid == true, adgroupIDTmp.String, model.PropertyValueNone).(string)
		campaignName = U.IfThenElse(campaignNameTmp.Valid == true, campaignNameTmp.String, model.PropertyValueNone).(string)
		campaignID = U.IfThenElse(campaignIDTmp.Valid == true, campaignIDTmp.String, model.PropertyValueNone).(string)
		adID = U.IfThenElse(adIDTmp.Valid == true, adIDTmp.String, model.PropertyValueNone).(string)
		keywordName = U.IfThenElse(keywordNameTmp.Valid == true, keywordNameTmp.String, model.PropertyValueNone).(string)
		keywordID = U.IfThenElse(keywordIDTmp.Valid == true, keywordIDTmp.String, model.PropertyValueNone).(string)
		keywordName2 = U.IfThenElse(keywordNameTmp2.Valid == true, keywordNameTmp2.String, model.PropertyValueNone).(string)
		keywordID2 = U.IfThenElse(keywordIDTmp2.Valid == true, keywordIDTmp2.String, model.PropertyValueNone).(string)
		slot = U.IfThenElse(slotTmp.Valid == true, slotTmp.String, model.PropertyValueNone).(string)

		// Enriching GCLID report using other reports
		if U.IsNonEmptyKey(campaignID) {
			if val, exists := campaignIDReport[campaignID]; exists {
				campaignName = val.Name
			}
		}
		if U.IsNonEmptyKey(adgroupID) {
			if val, exists := adgroupIDReport[adgroupID]; exists {
				adgroupName = val.Name
			}
		}
		keywordMatchType := ""
		if U.IsNonEmptyKey(keywordID) {
			if val, exists := keywordIDReport[keywordID]; exists {
				keywordName = val.Name
				keywordMatchType = val.KeywordMatchType
			}
		}
		if U.IsNonEmptyKey(keywordID2) {
			if val, exists := keywordIDReport[keywordID2]; exists {
				keywordName = val.Name
				keywordMatchType = val.KeywordMatchType
			}
		}
		if !U.IsNonEmptyKey(keywordName) {
			keywordName = keywordName2
		}
		if !U.IsNonEmptyKey(keywordID) {
			keywordID = keywordID2
		}

		gclidBasedMarketData[gclID] = model.MarketingData{
			ID:               gclID,
			AdgroupName:      adgroupName,
			AdgroupID:        adgroupID,
			CampaignName:     campaignName,
			CampaignID:       campaignID,
			AdID:             adID,
			KeywordID:        keywordID,
			KeywordName:      keywordName,
			KeywordMatchType: keywordMatchType,
			Slot:             slot,
		}
	}
	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &logFields)
	return gclidBasedMarketData, nil
}

// @TODO Kark v1
func (store *MemSQL) buildAdwordsChannelConfig(projectID int64) *model.ChannelConfigResult {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	adwordsObjectsAndProperties := store.buildObjectAndPropertiesForAdwords(projectID, model.ObjectsForAdwords)
	selectMetrics := append(SelectableMetricsForAllChannels, model.SelectableMetricsForAdwords...)
	objectsAndProperties := adwordsObjectsAndProperties
	return &model.ChannelConfigResult{
		SelectMetrics:        selectMetrics,
		ObjectsAndProperties: objectsAndProperties,
	}
}

func (store *MemSQL) buildObjectAndPropertiesForAdwords(projectID int64, objects []string) []model.ChannelObjectAndProperties {
	logFields := log.Fields{
		"project_id": projectID,
		"objects":    objects,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0, 0)
	for _, currentObject := range objects {
		// to do: check if normal properties present then only smart properties will be there
		propertiesAndRelated, isPresent := model.MapOfAdwordsObjectsToPropertiesAndRelated[currentObject]
		var currentProperties []model.ChannelProperty
		var currentPropertiesSmart []model.ChannelProperty
		if isPresent {
			currentProperties = buildProperties(propertiesAndRelated)
			smartProperty := store.GetSmartPropertyAndRelated(projectID, currentObject, "adwords")
			currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		} else {
			currentProperties = buildProperties(allChannelsPropertyToRelated)
			smartProperty := store.GetSmartPropertyAndRelated(projectID, currentObject, "adwords")
			currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		}
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
}

type LatestTimestamp struct {
	Timestamp int64 `json:"timestamp"`
}

// GetAdwordsFilterValues - @TODO Kark v1
func (store *MemSQL) GetAdwordsFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, isPresent := model.AdwordsExtToInternal[requestFilterProperty]
	if !isPresent {
		filterValues, errCode := store.getSmartPropertyFilterValues(projectID, requestFilterObject, requestFilterProperty, "adwords", reqID)
		if errCode != http.StatusFound {
			return []interface{}{}, http.StatusInternalServerError
		}
		return filterValues, http.StatusFound
	}
	adwordsInternalFilterProperty, docType, err := getFilterRelatedInformationForAdwords(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return make([]interface{}, 0, 0), http.StatusBadRequest
	}
	filterValues, errCode := store.getAdwordsFilterValuesByType(projectID, docType, adwordsInternalFilterProperty, reqID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}
func (store *MemSQL) getSmartPropertyFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, source string, reqID string) ([]interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"source":                  source,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	objectType, isExists := model.SmartPropertyRulesTypeAliasToType[requestFilterObject]
	if !isExists {
		logCtx.Error("Invalid filter object")
		return make([]interface{}, 0, 0), http.StatusBadRequest
	}
	smartPropertyRule := model.SmartPropertyRules{}
	filterValues := make([]interface{}, 0, 0)
	db := C.GetServices().Db
	err := db.Table("smart_property_rules").Where("project_id = ? AND type = ? AND name = ?",
		projectID, objectType, requestFilterProperty).Find(&smartPropertyRule).Error
	if err != nil {
		return make([]interface{}, 0, 0), http.StatusNotFound
	}
	propertiesValueMap := make(map[string]bool)
	var rules []model.Rule
	err = U.DecodePostgresJsonbToStructType(smartPropertyRule.Rules, &rules)
	if err != nil {
		return make([]interface{}, 0, 0), http.StatusNotFound
	}

	for _, rule := range rules {
		if rule.Source == "all" || rule.Source == source {
			propertiesValueMap[rule.Value] = true
		}
	}
	for key, _ := range propertiesValueMap {
		filterValues = append(filterValues, key)
	}

	return filterValues, http.StatusFound
}

func (store *MemSQL) IsAdwordsIntegrationAvailable(projectID int64) bool {
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return false
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId
	if customerAccountID == nil || len(*customerAccountID) == 0 {
		return false
	}
	return true
}

// GetAdwordsSQLQueryAndParametersForFilterValues - @TODO Kark v1
// Currently, properties in object dont vary with Object.
func (store *MemSQL) GetAdwordsSQLQueryAndParametersForFilterValues(projectID int64,
	requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	adwordsInternalFilterProperty, docType, err := getFilterRelatedInformationForAdwords(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return "", make([]interface{}, 0, 0), http.StatusBadRequest
	}
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("failed to fetch Project Setting in adwords filter values.")
		return "", []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId
	if customerAccountID == nil || len(*customerAccountID) == 0 {
		logCtx.Info(integrationNotAvailable)
		return "", []interface{}{}, http.StatusNotFound
	}
	customerAccountIDs := strings.Split(*customerAccountID, ",")
	from, to := model.GetFromAndToDatesForFilterValues()

	params := []interface{}{adwordsInternalFilterProperty, projectID, customerAccountIDs, docType, adwordsInternalFilterProperty, from, to}
	return "(" + adwordsFilterQueryStr + ")", params, http.StatusFound
}

func getFilterRelatedInformationForAdwords(requestFilterObject string, requestFilterProperty string) (string, int, int) {
	logFields := log.Fields{
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	adwordsInternalFilterObject, isPresent := model.AdwordsExtToInternal[requestFilterObject]
	if !isPresent {
		log.Error("Invalid adwords filter object.")
		return "", 0, http.StatusBadRequest
	}
	docType := getAdwordsDocumentTypeForFilterKeyV1(adwordsInternalFilterObject)

	adwordsInternalFilterProperty, isPresent := model.AdwordsExtToInternal[requestFilterProperty]
	if !isPresent {
		log.Error("Invalid adwords filter property.")
		return "", 0, http.StatusBadRequest
	}
	keyForJobInternalRepresentation := fmt.Sprintf("%s:%s", adwordsInternalFilterObject, adwordsInternalFilterProperty)
	adwordsInternalPropertyOfJob, isPresent := model.AdwordsInternalPropertiesToJobsInternal[keyForJobInternalRepresentation]
	if !isPresent {
		log.Error("Invalid adwords filter property for given object type.")
		return "", 0, http.StatusBadRequest
	}

	return adwordsInternalPropertyOfJob, docType, http.StatusOK
}

// @TODO Kark v1
// Not considering timezone since the impact is going to be very less.
func (store *MemSQL) getAdwordsFilterValuesByType(projectID int64, docType int, property string, reqID string) ([]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"doc_type":   docType,
		"property":   property,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("failed to fetch Project Setting in adwords filter values.")
		return []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId
	if customerAccountID == nil || len(*customerAccountID) == 0 {
		logCtx.Info(integrationNotAvailable)
		return []interface{}{}, http.StatusInternalServerError
	}
	var customerAccountIDs []string
	customerAccountIDs = strings.Split(*customerAccountID, ",")

	from, to := model.GetFromAndToDatesForFilterValues()
	logCtx = log.WithField("doc_type", docType)
	params := []interface{}{property, projectID, customerAccountIDs, docType, property, from, to}
	_, resultRows, err := store.ExecuteSQL(adwordsFilterQueryStr, params, logCtx)
	if err != nil {
		logCtx.WithError(err).WithField("query", adwordsFilterQueryStr).WithField("params", params).Error(model.AdwordsSpecificError)
		return make([]interface{}, 0, 0), http.StatusInternalServerError
	}
	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

// @TODO Kark v1
// This method uses internal filterObject as input param and not request filterObject.
// Note: method not to be used without proper validation of request params.
func getAdwordsDocumentTypeForFilterKeyV1(filterObject string) int {
	logFields := log.Fields{
		"filter_object": filterObject,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var docType int

	switch filterObject {
	case model.AdwordsCampaign:
		docType = model.AdwordsDocumentTypeAlias[model.AdwordsCampaign+"s"]
	case model.AdwordsAdGroup:
		docType = model.AdwordsDocumentTypeAlias[model.AdwordsAdGroup+"s"]
	case model.AdwordsAd:
		docType = model.AdwordsDocumentTypeAlias[model.AdwordsAd+"s"]
	case model.AdwordsKeyword:
		docType = model.AdwordsDocumentTypeAlias[model.KeywordPerformanceReport]
	}

	return docType
}

// ExecuteAdwordsmodel.ChannelQueryV1 - @TODO Kark v1.
// Job represents the meta data associated with particular object type. Reports represent data with metrics and few filters.
// TODO - Duplicate code/flow in facebook and adwords.
func (store *MemSQL) ExecuteAdwordsChannelQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	fetchSource := false
	logCtx := log.WithFields(logFields)
	if query.GroupByTimestamp == "" {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForAdwordsQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 10000", false, nil)
		if errCode == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if errCode != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := store.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.AdwordsSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	} else {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForAdwordsQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 1000", false, nil)
		if errCode == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if errCode != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := store.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.AdwordsSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		groupByCombinations := model.GetGroupByCombinationsForChannelAnalytics(columns, resultMetrics)
		sql, params, selectKeys, selectMetrics, errCode = store.GetSQLQueryAndParametersForAdwordsQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 10000", true, groupByCombinations)
		if errCode != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err = store.ExecuteSQL(sql, params, logCtx)
		columns = append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.AdwordsSpecificError)
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	}
}

// GetSQLQueryAndParametersForAdwordsQueryV1 - @Kark TODO v1
// TODO Understand null cases.
func (store *MemSQL) GetSQLQueryAndParametersForAdwordsQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string, int) {
	// NOTE: This contains 1000 rows. Should we remove from log group_by_combinations_for_gbt.
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"req_id":                        reqID,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_time_stamp":        isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var selectMetrics []string
	var selectKeys []string
	var sql string
	var params []interface{}
	logCtx := log.WithFields(logFields)
	transformedQuery, customerAccountID, err := store.transFormRequestFieldsAndFetchRequiredFieldsForAdwords(projectID, *query, reqID)
	if err != nil && err.Error() == integrationNotAvailable {
		logCtx.WithError(err).Info(model.AdwordsSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusNotFound
	}
	if err != nil {
		logCtx.WithError(err).Error(model.AdwordsSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusBadRequest
	}
	isSmartPropertyPresent := checkSmartProperty(query.Filters, query.GroupBy)
	if isSmartPropertyPresent {
		sql, params, selectKeys, selectMetrics = buildAdwordsSimpleQueryWithSmartPropertyV2(transformedQuery, projectID, *customerAccountID, reqID, fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
		return sql, params, selectKeys, selectMetrics, http.StatusOK
	}
	sql, params, selectKeys, selectMetrics = buildAdwordsSimpleQueryV2(transformedQuery, projectID, *customerAccountID, reqID, fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	return sql, params, selectKeys, selectMetrics, http.StatusOK
}

// @Kark TODO v1
func (store *MemSQL) transFormRequestFieldsAndFetchRequiredFieldsForAdwords(projectID int64, query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, *string, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var transformedQuery model.ChannelQueryV1
	logCtx := log.WithFields(logFields)
	var err error
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return &model.ChannelQueryV1{}, nil, errors.New("Project setting not found")
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId
	if customerAccountID == nil || len(*customerAccountID) == 0 {
		return &model.ChannelQueryV1{}, nil, errors.New(integrationNotAvailable)
	}

	transformedQuery, err = convertFromRequestToAdwordsSpecificRepresentation(query)
	if err != nil {
		logCtx.Warn("Request failed in validation: ", err)
		return &model.ChannelQueryV1{}, nil, err
	}
	return &transformedQuery, customerAccountID, nil
}

// @Kark TODO v1
// Currently, this relies on assumption of Object across different filterObjects. Change when we need robust.
func convertFromRequestToAdwordsSpecificRepresentation(query model.ChannelQueryV1) (model.ChannelQueryV1, error) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var transformedQuery model.ChannelQueryV1
	var err1, err2, err3 error
	transformedQuery.SelectMetrics, err1 = getAdwordsSpecificMetrics(query.SelectMetrics)
	transformedQuery.Filters, err2 = getAdwordsSpecificFilters(query.Filters)
	transformedQuery.GroupBy, err3 = getAdwordsSpecificGroupBy(query.GroupBy)
	if err1 != nil {
		return query, err1
	}
	if err2 != nil {
		return query, err2
	}
	if err3 != nil {
		return query, err3
	}
	transformedQuery.From = U.GetDateAsStringIn(query.From, U.TimeZoneString(query.Timezone))
	transformedQuery.To = U.GetDateAsStringIn(query.To, U.TimeZoneString(query.Timezone))
	transformedQuery.Timezone = query.Timezone
	transformedQuery.GroupByTimestamp = query.GroupByTimestamp

	return transformedQuery, nil
}

// @Kark TODO v1
func getAdwordsSpecificMetrics(requestSelectMetrics []string) ([]string, error) {
	logFields := log.Fields{
		"request_select_metrics": requestSelectMetrics,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultMetrics := make([]string, 0, 0)
	for _, requestMetric := range requestSelectMetrics {
		metric, isPresent := model.AdwordsExtToInternal[requestMetric]
		if !isPresent {
			return make([]string, 0, 0), errors.New("Invalid metric key found for document type")
		}
		resultMetrics = append(resultMetrics, metric)
	}
	return resultMetrics, nil
}

// @Kark TODO v1
func getAdwordsSpecificFilters(requestFilters []model.ChannelFilterV1) ([]model.ChannelFilterV1, error) {
	logFields := log.Fields{
		"request_filters": requestFilters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultFilters := make([]model.ChannelFilterV1, 0, 0)
	for _, requestFilter := range requestFilters {
		var resultFilter model.ChannelFilterV1
		filterObject, isPresent := model.AdwordsExtToInternal[requestFilter.Object]
		if !isPresent {
			return make([]model.ChannelFilterV1, 0, 0), errors.New("Invalid filter key found for document type")
		}
		resultFilter = requestFilter
		resultFilter.Object = filterObject
		resultFilters = append(resultFilters, resultFilter)
	}
	return resultFilters, nil
}

// @Kark TODO v1
func getAdwordsSpecificGroupBy(requestGroupBys []model.ChannelGroupBy) ([]model.ChannelGroupBy, error) {
	logFields := log.Fields{
		"request_group_by": requestGroupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	resultGroupBys := make([]model.ChannelGroupBy, 0, 0)
	for _, requestGroupBy := range requestGroupBys {
		var resultGroupBy model.ChannelGroupBy
		if requestGroupBy.Object == model.AdwordsSmartProperty {
			resultGroupBys = append(resultGroupBys, resultGroupBy)
		} else {
			groupByObject, isPresent := model.AdwordsExtToInternal[requestGroupBy.Object]
			if !isPresent {
				return make([]model.ChannelGroupBy, 0, 0), errors.New("Invalid groupby key found for document type")
			}
			resultGroupBy = requestGroupBy
			resultGroupBy.Object = groupByObject
			resultGroupBys = append(resultGroupBys, resultGroupBy)
		}
	}
	return resultGroupBys, nil
}

// @TODO Kark v1
// Complexity consideration - Having at max of 20 filters and 20 group by should be fine.
// change algo/strategy the filters and group by goes beyond 100.
func getLowestHierarchyLevelForAdwords(query *model.ChannelQueryV1) string {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// Fetch the propertyNames
	return getLowestHierarchyLevelForAdwordsFiltersAndGroupBy(query.Filters, query.GroupBy)
}

// @TODO Kark v1
func getLowestHierarchyLevelForAdwordsFiltersAndGroupBy(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) string {
	logFields := log.Fields{
		"filters":   filters,
		"group_bys": groupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var objectNames []string
	for _, filter := range filters {
		objectNames = append(objectNames, filter.Object)
	}

	for _, groupBy := range groupBys {
		objectNames = append(objectNames, groupBy.Object)
	}

	// Check if present
	for _, objectName := range objectNames {
		if objectName == model.AdwordsAd {
			return model.AdwordsAd
		}
	}

	for _, objectName := range objectNames {
		if objectName == model.AdwordsKeyword {
			return model.AdwordsKeyword
		}
	}

	for _, objectName := range objectNames {
		if objectName == model.AdwordsAdGroup {
			return model.AdwordsAdGroup
		}
	}

	for _, objectName := range objectNames {
		if objectName == model.AdwordsCampaign {
			return model.AdwordsCampaign
		}
	}
	return model.AdwordsCampaign
}

/*
SELECT JSON_EXTRACT_STRING(value, 'campaign_name') as campaign_name, date_trunc('day', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE 'UTC') as datetime,
SUM(JSON_EXTRACT_STRING(value, 'impressions')) as impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) as clicks FROM adwords_documents WHERE project_id = '2' AND
customer_account_id IN ( '2368493227' ) AND type = '5' AND timestamp between '20200331' AND '20200401'
AND JSON_EXTRACT_STRING(value, 'campaign_name') RLIKE '%Brand - BLR - New_Aug_Desktop_RLSA%' GROUP BY campaign_name, datetime
ORDER BY impressions DESC, clicks DESC LIMIT 2500 ;
*/
// - For reference of complex joins, PR which removed older/QueryV1 adwords is 1437.
func buildAdwordsSimpleQueryV2(query *model.ChannelQueryV1, projectID int64, customerAccountID string, reqID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string) {
	logFields := log.Fields{
		"query":                         query,
		"project_id":                    projectID,
		"customer_account_id":           customerAccountID,
		"req_id":                        reqID,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	lowestHierarchyLevel := getLowestHierarchyLevelForAdwords(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_performance_report"
	return getSQLAndParamsForAdwordsV2(query, projectID, query.From, query.To, customerAccountID, model.AdwordsDocumentTypeAlias[lowestHierarchyReportLevel], fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
}

func buildAdwordsSimpleQueryWithSmartPropertyV2(query *model.ChannelQueryV1, projectID int64, customerAccountID string, reqID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string) {
	logFields := log.Fields{
		"query":                         query,
		"project_id":                    projectID,
		"customer_account_id":           customerAccountID,
		"req_id":                        reqID,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	lowestHierarchyLevel := getLowestHierarchyLevelForAdwords(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_performance_report"
	return getSQLAndParamsForAdwordsWithSmartPropertyV2(query, projectID, query.From, query.To, customerAccountID, model.AdwordsDocumentTypeAlias[lowestHierarchyReportLevel], fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
}
func getSQLAndParamsForAdwordsWithSmartPropertyV2(query *model.ChannelQueryV1, projectID int64, from, to int64, customerAccountID string,
	docType int, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string) {
	logFields := log.Fields{
		"query":                         query,
		"project_id":                    projectID,
		"from":                          from,
		"to":                            to,
		"customer_account_id":           customerAccountID,
		"doc_type":                      docType,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	computeHigherOrderMetricsHere := !fetchSource
	customerAccountIDs := strings.Split(customerAccountID, ",")
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	filterPropertiesStatementBasedOnRequestFilters := ""
	// isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	toFetchImpressionsForHigherOrderMetric := false

	finalParams := make([]interface{}, 0, 0)
	finalWhereStatement := ""
	finalGroupByKeys := make([]string, 0, 0)
	finalGroupByStatement := ""
	finalSelectKeys := make([]string, 0, 0)
	finalSelectStatement := ""
	finalOrderByKeys := make([]string, 0, 0)
	finalOrderByStatement := ""
	resultantSQLStatement := ""

	adwordsGroupBys := make([]model.ChannelGroupBy, 0, 0)

	// Group By
	dimensions := fields{}
	for _, groupBy := range query.GroupBy {
		_, isPresent := model.SmartPropertyReservedNames[groupBy.Property]
		isSmartProperty := !isPresent
		if isSmartProperty {
			if groupBy.Object == model.AdwordsCampaign {
				expression := fmt.Sprintf(`%s as %s`, fmt.Sprintf("JSON_EXTRACT_STRING(campaign.properties, '%s')", groupBy.Property), model.CampaignPrefix+groupBy.Property)
				dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
				dimensions.values = append(dimensions.values, model.CampaignPrefix+groupBy.Property)
			} else {
				expression := fmt.Sprintf(`%s as %s`, fmt.Sprintf("JSON_EXTRACT_STRING(ad_group.properties, '%s')", groupBy.Property), model.AdgroupPrefix+groupBy.Property)
				dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
				dimensions.values = append(dimensions.values, model.AdgroupPrefix+groupBy.Property)
			}
		} else {
			if groupBy.Object == CAFilterChannel {
				externalValue := groupBy.Object + "_" + groupBy.Property
				expression := fmt.Sprintf("'Google Ads' as %s", externalValue)
				dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
				dimensions.values = append(dimensions.values, externalValue)
			} else {
				key := groupBy.Object + ":" + groupBy.Property
				internalValue := model.AdwordsInternalPropertiesToReportsInternal[key]
				externalValue := groupBy.Object + "_" + groupBy.Property
				var expression string
				if groupBy.Property == "id" {
					expression = fmt.Sprintf("%s as %s", internalValue, externalValue)
				} else if _, ok := propertiesToBeDividedByMillion[groupBy.Property]; ok {
					expression = fmt.Sprintf("((JSON_EXTRACT_STRING(value, '%s')))/1000000 as %s", internalValue, externalValue)
				} else {
					expression = fmt.Sprintf("JSON_EXTRACT_STRING(value, '%s') as %s", internalValue, externalValue)
				}
				dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
				dimensions.values = append(dimensions.values, externalValue)
			}
		}
	}

	if isGroupByTimestamp {
		internalValue := getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone)
		externalValue := model.AliasDateTime
		expression := fmt.Sprintf("%s as %s", internalValue, externalValue)
		dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
		dimensions.values = append(dimensions.values, externalValue)
	}

	// select Keys
	// TODO Later: Issue for conversion_rate or click_through_rate nonHigherOrder as they dont have impressions.
	metrics := fields{}
	for _, selectMetric := range query.SelectMetrics {
		var internalValue, externalValue string
		_, isNotHigherOrderMetric := nonHigherOrderMetrics[selectMetric]
		if isNotHigherOrderMetric || !computeHigherOrderMetricsHere {
			internalValue = adwordsInternalMetricsToAllRep[selectMetric].nonHigherOrderExpression
			externalValue = adwordsInternalMetricsToAllRep[selectMetric].externalValue
		} else {
			internalValue = adwordsInternalMetricsToAllRep[selectMetric].higherOrderExpression
			externalValue = adwordsInternalMetricsToAllRep[selectMetric].externalValue
		}
		expression := fmt.Sprintf("%s as %s", internalValue, externalValue)
		metrics.selectExpressions = append(metrics.selectExpressions, expression)
		metrics.values = append(metrics.values, externalValue)
	}

	for _, selectMetric := range query.SelectMetrics {
		_, isNonHigherOrderMetric := nonHigherOrderMetrics[selectMetric]
		if selectMetric == model.Impressions {
			toFetchImpressionsForHigherOrderMetric = false
			break
		} else if !isNonHigherOrderMetric && !computeHigherOrderMetricsHere {
			toFetchImpressionsForHigherOrderMetric = true
		}
	}

	if toFetchImpressionsForHigherOrderMetric {
		internalValue := adwordsInternalMetricsToAllRep[model.Impressions].nonHigherOrderExpression
		externalValue := adwordsInternalMetricsToAllRep[model.Impressions].externalValue
		expression := fmt.Sprintf("%s as %s", internalValue, externalValue)
		metrics.selectExpressions = append(metrics.selectExpressions, expression)
		metrics.values = append(metrics.values, externalValue)
	}

	// Filters
	filterPropertiesStatementBasedOnRequestFilters, filterParams := getFilterPropertiesForAdwordsReportsAndSmartProperty(query.Filters)
	filterStatementForSmartPropertyGroupBy := getNotNullFilterStatementForSmartPropertyGroupBys(adwordsGroupBys)
	finalWhereStatement = joinWithWordInBetween("AND", staticWhereStatementForAdwordsWithSmartProperty, filterPropertiesStatementBasedOnRequestFilters, filterStatementForSmartPropertyGroupBy)
	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, filterParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForAdwords(groupByCombinationsForGBT)
		finalWhereStatement += " AND (" + whereConditionForGBT + ") "
		finalParams = append(finalParams, whereParams...)
	}
	finalGroupByKeys = dimensions.values
	if len(finalGroupByKeys) != 0 {
		finalGroupByStatement = " GROUP BY " + joinWithComma(finalGroupByKeys...)
	}

	// orderBy
	if isGroupByTimestamp {
		finalOrderByKeys = appendSuffix("ASC", model.AliasDateTime)
	} else {
		finalOrderByKeys = appendSuffix("DESC", metrics.values...)
	}
	if len(finalOrderByKeys) != 0 {
		finalOrderByStatement = " ORDER BY " + joinWithComma(finalOrderByKeys...)
	}

	finalSelectKeys = append(finalSelectKeys, dimensions.selectExpressions...)
	finalSelectKeys = append(finalSelectKeys, metrics.selectExpressions...)
	finalSelectStatement = "SELECT " + joinWithComma(finalSelectKeys...)

	fromStatement := getAdwordsFromStatementWithJoins(query.Filters, query.GroupBy)
	// finalSQL
	resultantSQLStatement = finalSelectStatement + fromStatement + finalWhereStatement +
		finalGroupByStatement + finalOrderByStatement + limitString

	return resultantSQLStatement, finalParams, dimensions.values, metrics.values
}

func buildWhereConditionForGBTForAdwords(groupByCombinations map[string][]interface{}) (string, []interface{}) {
	logFields := log.Fields{
		"group_by_combinations": groupByCombinations,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultantWhereCondition := ""
	resultantInClauses := make([]string, 0)
	params := make([]interface{}, 0)

	for dimension, values := range groupByCombinations {
		currentInClause := ""

		valuesInString := make([]string, 0)
		for _, value := range values {
			valuesInString = append(valuesInString, "?")
			params = append(params, value)
		}
		currentInClause = joinWithComma(valuesInString...)

		resultantInClauses = append(resultantInClauses, fmt.Sprintf("%s IN (", dimension)+currentInClause+") ")
	}

	resultantWhereCondition = joinWithWordInBetween("AND", resultantInClauses...)

	return resultantWhereCondition, params
}

// request has dimension - campaign_name.
// response has adwords_documents.value, 'campaign_name'.
func GetFilterObjectAndExtractPropertyKeyForChannelAdwords(dimension string) (string, string) {
	filterObjectAdwords := "adwords_documents.value"
	filterObjectForSmartPropertiesCampaign := "campaign.properties"
	filterObjectForSmartPropertiesAdGroup := "ad_group.properties"

	filterObject := ""
	filterProperty := ""
	isNotSmartProperty := false
	if strings.HasPrefix(dimension, model.CampaignPrefix) {
		filterProperty, isNotSmartProperty = model.GetReportPropertyIfPresentForAdwords("campaign", dimension, model.CampaignPrefix)
		if isNotSmartProperty {
			filterObject = filterObjectAdwords
		} else {
			filterObject = filterObjectForSmartPropertiesCampaign
			filterProperty = strings.TrimPrefix(dimension, model.CampaignPrefix)
		}
	} else if strings.HasPrefix(dimension, model.AdgroupPrefix) {
		filterProperty, isNotSmartProperty = model.GetReportPropertyIfPresentForAdwords("ad_group", dimension, model.AdgroupPrefix)
		if isNotSmartProperty {
			filterObject = filterObjectAdwords
		} else {
			filterObject = filterObjectForSmartPropertiesAdGroup
			filterProperty = strings.TrimPrefix(dimension, model.CampaignPrefix)
		}
	} else {
		filterObject = filterObjectAdwords
		filterProperty, _ = model.GetReportPropertyIfPresentForAdwords("keyword", dimension, model.KeywordPrefix)
	}
	return filterObject, filterProperty
}

func getAdwordsFromStatementWithJoins(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) string {
	logFields := log.Fields{
		"filters":   filters,
		"group_bys": groupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	isPresentCampaignSmartProperty, isPresentAdGroupSmartProperty := checkSmartPropertyWithTypeAndSource(filters, groupBys, "adwords")
	fromStatement := fromAdwordsDocument
	if isPresentAdGroupSmartProperty {
		fromStatement += "inner join smart_properties ad_group on ad_group.project_id = adwords_documents.project_id and ad_group.object_id = ad_group_id "
	}
	if isPresentCampaignSmartProperty {
		fromStatement += "inner join smart_properties campaign on campaign.project_id = adwords_documents.project_id and campaign.object_id = campaign_id "
	}
	return fromStatement
}

func getSQLAndParamsForAdwordsV2(query *model.ChannelQueryV1, projectID int64, from, to int64, customerAccountID string,
	docType int, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string) {
	logFields := log.Fields{
		"query":                         query,
		"project_id":                    projectID,
		"customer_account_id":           customerAccountID,
		"doc_type":                      docType,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	computeHigherOrderMetricsHere := !fetchSource
	customerAccountIDs := strings.Split(customerAccountID, ",")
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	filterPropertiesStatementBasedOnRequestFilters := ""
	toFetchImpressionsForHigherOrderMetric := false

	finalParams := make([]interface{}, 0, 0)
	finalWhereStatement := ""
	finalGroupByKeys := make([]string, 0, 0)
	finalGroupByStatement := ""
	finalSelectKeys := make([]string, 0, 0)
	finalSelectStatement := ""
	finalOrderByKeys := make([]string, 0, 0)
	finalOrderByStatement := ""
	resultantSQLStatement := ""

	// Group By
	dimensions := fields{}

	for _, groupBy := range query.GroupBy {
		if groupBy.Object == CAFilterChannel {
			externalValue := groupBy.Object + "_" + groupBy.Property
			expression := fmt.Sprintf("'Google Ads' as %s", externalValue)
			dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
			dimensions.values = append(dimensions.values, externalValue)
		} else {
			key := groupBy.Object + ":" + groupBy.Property
			internalValue := model.AdwordsInternalPropertiesToReportsInternal[key]
			externalValue := groupBy.Object + "_" + groupBy.Property
			var expression string
			if groupBy.Property == "id" {
				expression = fmt.Sprintf("%s as %s", internalValue, externalValue)
			} else if _, ok := propertiesToBeDividedByMillion[groupBy.Property]; ok {
				expression = fmt.Sprintf("((JSON_EXTRACT_STRING(value, '%s')))/1000000 as %s", internalValue, externalValue)
			} else {
				expression = fmt.Sprintf("JSON_EXTRACT_STRING(value, '%s') as %s", internalValue, externalValue)
			}
			dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
			dimensions.values = append(dimensions.values, externalValue)
		}
	}
	if isGroupByTimestamp {
		internalValue := getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone)
		externalValue := model.AliasDateTime
		expression := fmt.Sprintf("%s as %s", internalValue, externalValue)
		dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
		dimensions.values = append(dimensions.values, externalValue)
	}

	// select Keys
	// TODO Later: Issue for conversion_rate or click_through_rate nonHigherOrder as they dont have impressions.
	metrics := fields{}
	for _, selectMetric := range query.SelectMetrics {
		var internalValue, externalValue string
		_, isNotHigherOrderMetric := nonHigherOrderMetrics[selectMetric]
		if isNotHigherOrderMetric || !computeHigherOrderMetricsHere {
			internalValue = adwordsInternalMetricsToAllRep[selectMetric].nonHigherOrderExpression
			externalValue = adwordsInternalMetricsToAllRep[selectMetric].externalValue
		} else {
			internalValue = adwordsInternalMetricsToAllRep[selectMetric].higherOrderExpression
			externalValue = adwordsInternalMetricsToAllRep[selectMetric].externalValue
		}
		expression := fmt.Sprintf("%s as %s", internalValue, externalValue)
		metrics.selectExpressions = append(metrics.selectExpressions, expression)
		metrics.values = append(metrics.values, externalValue)
	}

	for _, selectMetric := range query.SelectMetrics {
		_, isNonHigherOrderMetric := nonHigherOrderMetrics[selectMetric]
		if selectMetric == model.Impressions {
			toFetchImpressionsForHigherOrderMetric = false
			break
		} else if !isNonHigherOrderMetric && !computeHigherOrderMetricsHere {
			toFetchImpressionsForHigherOrderMetric = true
		}
	}

	if toFetchImpressionsForHigherOrderMetric {
		internalValue := adwordsInternalMetricsToAllRep[model.Impressions].nonHigherOrderExpression
		externalValue := adwordsInternalMetricsToAllRep[model.Impressions].externalValue
		expression := fmt.Sprintf("%s as %s", internalValue, externalValue)
		metrics.selectExpressions = append(metrics.selectExpressions, expression)
		metrics.values = append(metrics.values, externalValue)
	}

	// Filters
	filterPropertiesStatementBasedOnRequestFilters, filterParams := getFilterPropertiesForAdwordsReports(query.Filters)
	finalWhereStatement = joinWithWordInBetween("AND", staticWhereStatementForAdwords, filterPropertiesStatementBasedOnRequestFilters)
	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, filterParams...)
	if groupByCombinationsForGBT != nil && len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForAdwords(groupByCombinationsForGBT)
		finalWhereStatement += " AND (" + whereConditionForGBT + ") "
		finalParams = append(finalParams, whereParams...)
	}

	finalGroupByKeys = dimensions.values
	if len(finalGroupByKeys) != 0 {
		finalGroupByStatement = " GROUP BY " + joinWithComma(finalGroupByKeys...)
	}

	// orderBy
	if isGroupByTimestamp {
		finalOrderByKeys = appendSuffix("ASC", model.AliasDateTime)
	} else {
		finalOrderByKeys = appendSuffix("DESC", metrics.values...)
	}
	if len(finalOrderByKeys) != 0 {
		finalOrderByStatement = " ORDER BY " + joinWithComma(finalOrderByKeys...)
	}

	finalSelectKeys = append(finalSelectKeys, dimensions.selectExpressions...)
	finalSelectKeys = append(finalSelectKeys, metrics.selectExpressions...)
	finalSelectStatement = "SELECT " + joinWithComma(finalSelectKeys...)

	// finalSQL
	resultantSQLStatement = finalSelectStatement + fromAdwordsDocument + finalWhereStatement +
		finalGroupByStatement + finalOrderByStatement + limitString
	return resultantSQLStatement, finalParams, dimensions.values, metrics.values
}

// @Kark TODO v1
// TODO Check if we have none operator
func getFilterPropertiesForAdwordsReports(filters []model.ChannelFilterV1) (string, []interface{}) {
	logFields := log.Fields{
		"filters": filters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultStatement := ""
	var filterValue string
	params := make([]interface{}, 0, 0)
	if len(filters) == 0 {
		return resultStatement, params
	}
	for index, filter := range filters {
		currentFilterStatement := ""
		currentFilterProperty := ""
		if filter.LogicalOp == "" {
			filter.LogicalOp = "AND"
		}
		filterOperator := getOp(filter.Condition, "categorical")
		if filter.Condition == model.ContainsOpStr || filter.Condition == model.NotContainsOpStr {
			filterValue = fmt.Sprintf("%s", filter.Value)
		} else {
			filterValue = filter.Value
		}
		key := fmt.Sprintf("%s:%s", filter.Object, filter.Property)
		currentFilterProperty = model.AdwordsInternalPropertiesToReportsInternal[key]
		if strings.Contains(filter.Property, ("id")) {
			currentFilterStatement = fmt.Sprintf("%s %s ?", currentFilterProperty, filterOperator)
		} else {
			currentFilterStatement = fmt.Sprintf("JSON_EXTRACT_STRING(value, '%s') %s ?", currentFilterProperty, filterOperator)
		}
		params = append(params, filterValue)
		if index == 0 {
			resultStatement = fmt.Sprintf("(%s", currentFilterStatement)
		} else {
			resultStatement = fmt.Sprintf("%s %s %s", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement + ")", params
}
func getFilterPropertiesForAdwordsReportsAndSmartProperty(filters []model.ChannelFilterV1) (string, []interface{}) {
	logFields := log.Fields{
		"filters": filters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultStatement := ""
	var filterValue string
	params := make([]interface{}, 0, 0)
	if len(filters) == 0 {
		return resultStatement, params
	}
	campaignFilter := ""
	adGroupFilter := ""
	for index, filter := range filters {
		currentFilterStatement := ""
		currentFilterProperty := ""
		if filter.LogicalOp == "" {
			filter.LogicalOp = "AND"
		}
		filterOperator := getOp(filter.Condition, "categorical")
		if filter.Condition == model.ContainsOpStr || filter.Condition == model.NotContainsOpStr {
			filterValue = fmt.Sprintf("%s", filter.Value)
		} else {
			filterValue = filter.Value
		}
		_, isPresent := model.AdwordsExtToInternal[filter.Property]
		if isPresent {
			key := fmt.Sprintf("%s:%s", filter.Object, filter.Property)
			currentFilterProperty = model.AdwordsInternalPropertiesToReportsInternal[key]
			if strings.Contains(filter.Property, ("id")) {
				currentFilterStatement = fmt.Sprintf("%s.%s %s ?", adwordsDocuments, currentFilterProperty, filterOperator)
			} else {
				currentFilterStatement = fmt.Sprintf("JSON_EXTRACT_STRING(%s.value, '%s') %s ?", adwordsDocuments, currentFilterProperty, filterOperator)
			}
			params = append(params, filterValue)
			if index == 0 {
				resultStatement = fmt.Sprintf("(%s", currentFilterStatement)
			} else {
				resultStatement = fmt.Sprintf("%s %s %s", resultStatement, filter.LogicalOp, currentFilterStatement)
			}
		} else {
			currentFilterStatement = fmt.Sprintf("JSON_EXTRACT_STRING(%s.properties, '%s') %s '%s'", filter.Object, filter.Property, filterOperator, filterValue)
			if index == 0 {
				resultStatement = fmt.Sprintf("(%s", currentFilterStatement)
			} else {
				resultStatement = fmt.Sprintf("%s %s %s", resultStatement, filter.LogicalOp, currentFilterStatement)
			}
			if filter.Object == "campaign" {
				campaignFilter = smartPropertyCampaignStaticFilter
			} else {
				adGroupFilter = smartPropertyAdGroupStaticFilter
			}
		}
	}
	if campaignFilter != "" {
		resultStatement += (" AND " + campaignFilter)
	}
	if adGroupFilter != "" {
		resultStatement += (" AND " + adGroupFilter)
	}

	return resultStatement + ")", params
}
func getNotNullFilterStatementForSmartPropertyGroupBys(groupBys []model.ChannelGroupBy) string {
	logFields := log.Fields{
		"group_bys": groupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultStatement := ""
	for _, groupBy := range groupBys {
		_, isPresent := model.SmartPropertyReservedNames[groupBy.Property]
		isSmartProperty := !isPresent
		if isSmartProperty {
			if groupBy.Object == model.AdwordsCampaign || groupBy.Object == model.BingAdsCampaign {
				if resultStatement == "" {
					resultStatement += fmt.Sprintf("( JSON_EXTRACT_STRING(campaign.properties, '%s') IS NOT NULL ", groupBy.Property)
				} else {
					resultStatement += fmt.Sprintf("AND JSON_EXTRACT_STRING(campaign.properties, '%s') IS NOT NULL ", groupBy.Property)
				}
			} else {
				if resultStatement == "" {
					resultStatement += fmt.Sprintf("( JSON_EXTRACT_STRING(ad_group.properties,'%s') IS NOT NULL ", groupBy.Property)
				} else {
					resultStatement += fmt.Sprintf("AND JSON_EXTRACT_STRING(ad_group.properties,'%s') IS NOT NULL ", groupBy.Property)
				}
			}

		}
	}

	if resultStatement == "" {
		return resultStatement
	}
	return resultStatement + ")"
}

func (store *MemSQL) GetLatestMetaForAdwordsForGivenDays(projectID int64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields) {
	logFields := log.Fields{
		"days":       days,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	channelDocumentsCampaign := make([]model.ChannelDocumentsWithFields, 0, 0)
	channelDocumentsAdGroup := make([]model.ChannelDocumentsWithFields, 0, 0)

	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		log.Error("Failed to get project settings")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	if projectSetting.IntAdwordsCustomerAccountId == nil || *(projectSetting.IntAdwordsCustomerAccountId) == "" {
		log.WithField("projectID", projectID).Error("Integration of adwords is not available for this project.")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	customerAccountIDs := strings.Split(*(projectSetting.IntAdwordsCustomerAccountId), ",")

	to, err := strconv.ParseUint(time.Now().Format("20060102"), 10, 64)
	if err != nil {
		log.Error("Failed to parse to timestamp")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	from, err := strconv.ParseUint(time.Now().AddDate(0, 0, -days).Format("20060102"), 10, 64)
	if err != nil {
		log.Error("Failed to parse from timestamp")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	query := adwordsAdGroupMetadataFetchQueryStr
	params := []interface{}{model.AdwordsDocumentTypeAlias["ad_groups"], projectID, from, to, customerAccountIDs,
		model.AdwordsDocumentTypeAlias["ad_groups"], projectID, from, to, customerAccountIDs}

	rows1, tx1, err, queryID1 := store.ExecQueryWithContext(query, params)
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d ad_group meta for adwords", days)
		log.WithField("error string", err).Error(errString)
		U.CloseReadQuery(rows1, tx1)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	startReadTime1 := time.Now()
	for rows1.Next() {
		currentRecord := model.ChannelDocumentsWithFields{}
		rows1.Scan(&currentRecord.AdGroupID, &currentRecord.CampaignID, &currentRecord.AdGroupName, &currentRecord.CampaignName)
		channelDocumentsAdGroup = append(channelDocumentsAdGroup, currentRecord)
	}
	U.CloseReadQuery(rows1, tx1)
	U.LogReadTimeWithQueryRequestID(startReadTime1, queryID1, &logFields)

	query = adwordsCampaignMetadataFetchQueryStr
	params = []interface{}{model.AdwordsDocumentTypeAlias["campaigns"], projectID, from, to, customerAccountIDs, model.AdwordsDocumentTypeAlias["campaigns"], projectID, from, to, customerAccountIDs}

	rows2, tx2, err, queryID2 := store.ExecQueryWithContext(query, params)
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d campaign meta for adwords", days)
		log.WithField("error string", err).Error(errString)
		U.CloseReadQuery(rows2, tx2)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	startReadTime2 := time.Now()
	for rows2.Next() {
		currentRecord := model.ChannelDocumentsWithFields{}
		rows2.Scan(&currentRecord.CampaignID, &currentRecord.CampaignName)
		channelDocumentsCampaign = append(channelDocumentsCampaign, currentRecord)
	}
	U.CloseReadQuery(rows2, tx2)
	U.LogReadTimeWithQueryRequestID(startReadTime2, queryID2, &logFields)

	return channelDocumentsCampaign, channelDocumentsAdGroup
}

func (store *MemSQL) ExecuteAdwordsSEMChecklistQuery(projectID int64, query model.TemplateQuery, reqID string) (model.TemplateResponse, int) {
	logFields := log.Fields{
		"query":      query,
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	customerAccountID, err := store.validateIntegratonAndMetricsForAdwordsSEMChecklist(projectID, query, reqID)
	if err != nil && err.Error() == integrationNotAvailable {
		logCtx.WithError(err).Info(model.AdwordsSpecificError)
		return model.TemplateResponse{}, http.StatusNotFound
	}
	customerAccountIDs := strings.Split(*customerAccountID, ",")
	templateQueryResponse, err := store.getAdwordsSEMChecklistQueryData(query, projectID, customerAccountIDs, reqID)
	if err != nil {
		logCtx.Error("Failed to get template query response. Error: ", err.Error())
		return model.TemplateResponse{}, http.StatusNotFound
	}
	return templateQueryResponse, http.StatusOK
}
func (store *MemSQL) validateIntegratonAndMetricsForAdwordsSEMChecklist(projectID int64, query model.TemplateQuery, reqID string) (*string, error) {
	logFields := log.Fields{
		"query":      query,
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	var err error
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, errors.New("Project setting not found")
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId
	if customerAccountID == nil || len(*customerAccountID) == 0 {
		return nil, errors.New(integrationNotAvailable)
	}

	_, isValidMetric := model.TemplateAdwordsMetricsMapForAdwords[query.Metric]
	if !isValidMetric {
		logCtx.Warn("Request failed in validation: ", err)
		return nil, err
	}
	return customerAccountID, nil
}

var coalesceSelectsForKeywords = []string{"keyword_name", "keyword_id", "campaign_id", "keyword_match_type"}

func buildSelectStmnForKeywordTemplates() string {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	selectStmnt := ""
	for _, value := range coalesceSelectsForKeywords {
		selectStmnt += (fmt.Sprintf(coalesceForChecklists, "keyword_analysis_last_week", value, "keyword_analysis_previous_week", value, value) + ", ")
	}
	selectStmnt += fmt.Sprintf("%s as %s, ", fmt.Sprintf(percentageChangeForSemChecklist, "keyword_analysis_last_week", "analysis_metric", "keyword_analysis_previous_week", "analysis_metric", "keyword_analysis_previous_week", "analysis_metric"), "percentage_change")
	selectStmnt += fixedSelectForBreakdownAnalysisKeyword
	return selectStmnt
}
func (store *MemSQL) getKeywordLevelDataForTemplates(projectID int64, customerAccountID []string, query model.TemplateQuery) ([]model.KeywordAnalysis, error) {
	logFields := log.Fields{
		"query":               query,
		"project_id":          projectID,
		"customer_account_id": customerAccountID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	finalKeywordQuery := ""
	lastWeekAnalysisQuery := ""
	previousWeekAnalysisQuery := ""
	weeklyDifferenceAnalysisQuery := ""
	params := make([]interface{}, 0)
	staticParamsForKeywordsTemplates := []interface{}{projectID, customerAccountID,
		model.AdwordsDocumentTypeAlias["keyword_performance_report"]}
	if query.Metric != model.Conversion && query.Metric != "cost_per_lead" {
		previousWeekAnalysisQuery = fmt.Sprintf(adwordsTemplatesWeeklyKeywordAnalysisQuery, templateMetricsToSelectStatement[query.Metric], "")
		params = append(params, staticParamsForKeywordsTemplates...)
		params = append(params, query.PrevFrom, query.PrevTo)
		lastWeekAnalysisQuery = fmt.Sprintf(adwordsTemplatesWeeklyKeywordAnalysisQuery, templateMetricsToSelectStatement[query.Metric], "")
		params = append(params, staticParamsForKeywordsTemplates...)
		params = append(params, query.From, query.To)
		weeklyDifferenceAnalysisSelectQuery := fmt.Sprintf(adwordsTemplateWeeklyDifferenceAnalysisKeywordsQuerySelectStmnt, "", buildSelectStmnForKeywordTemplates())
		weeklyDifferenceAnalysisQuery = weeklyDifferenceAnalysisSelectQuery + adwordsTemplateWeeklyDifferenceAnalysisKeywordsQueryJoinStmnt + adwordsTemplateWeeklyDifferenceAnalysisKeywordsQueryWhereStmnt

	} else {
		extraSelectWeekAnalysisForRCA := fmt.Sprintf(adwordsTemplatesExtraSelectWeekAnalysisForRCA,
			templateMetricsToSelectStatement[model.Impressions], templateMetricsToSelectStatement[model.SearchImpressionShare],
			templateMetricsToSelectStatement[model.ConversionRate], templateMetricsToSelectStatement[model.ClickThroughRate], templateMetricsToSelectStatement[model.CostPerClick],
		)
		extraSelectBreakdownAnalysisForRCA := fmt.Sprintf(adwordsTemplatesExtraSelectBreakdownAnalysisForRCA,
			fmt.Sprintf(percentageChangeForSemChecklist, "keyword_analysis_last_week", model.Impressions, "keyword_analysis_previous_week", model.Impressions, "keyword_analysis_previous_week", model.Impressions),
			fmt.Sprintf(percentageChangeForSemChecklist, "keyword_analysis_last_week", model.SearchImpressionShare, "keyword_analysis_previous_week", model.SearchImpressionShare, "keyword_analysis_previous_week", model.SearchImpressionShare),
			fmt.Sprintf(percentageChangeForSemChecklist, "keyword_analysis_last_week", model.ConversionRate, "keyword_analysis_previous_week", model.ConversionRate, "keyword_analysis_previous_week", model.ConversionRate),
			fmt.Sprintf(percentageChangeForSemChecklist, "keyword_analysis_last_week", model.ClickThroughRate, "keyword_analysis_previous_week", model.ClickThroughRate, "keyword_analysis_previous_week", model.ClickThroughRate),
			fmt.Sprintf(percentageChangeForSemChecklist, "keyword_analysis_last_week", model.CostPerClick, "keyword_analysis_previous_week", model.CostPerClick, "keyword_analysis_previous_week", model.CostPerClick),
			"keyword_analysis_previous_week.impressions", "keyword_analysis_previous_week.search_impression_share", "keyword_analysis_previous_week.conversion_rate", "keyword_analysis_previous_week.click_through_rate", "keyword_analysis_previous_week.cost_per_click",
			"keyword_analysis_last_week.impressions", "keyword_analysis_last_week.search_impression_share", "keyword_analysis_last_week.conversion_rate", "keyword_analysis_last_week.click_through_rate", "keyword_analysis_last_week.cost_per_click")
		previousWeekAnalysisQuery = fmt.Sprintf(adwordsTemplatesWeeklyKeywordAnalysisQuery, templateMetricsToSelectStatement[query.Metric], extraSelectWeekAnalysisForRCA)
		params = append(params, staticParamsForKeywordsTemplates...)
		params = append(params, query.PrevFrom, query.PrevTo)
		lastWeekAnalysisQuery = fmt.Sprintf(adwordsTemplatesWeeklyKeywordAnalysisQuery, templateMetricsToSelectStatement[query.Metric], extraSelectWeekAnalysisForRCA)
		params = append(params, staticParamsForKeywordsTemplates...)
		params = append(params, query.From, query.To)
		weeklyDifferenceAnalysisSelectQuery := fmt.Sprintf(adwordsTemplateWeeklyDifferenceAnalysisKeywordsQuerySelectStmnt, extraSelectBreakdownAnalysisForRCA, buildSelectStmnForKeywordTemplates())
		weeklyDifferenceAnalysisQuery = weeklyDifferenceAnalysisSelectQuery + adwordsTemplateWeeklyDifferenceAnalysisKeywordsQueryJoinStmnt + adwordsTemplateWeeklyDifferenceAnalysisKeywordsQueryWhereStmnt

	}
	params = append(params, query.Thresholds.PercentageChange, query.Thresholds.AbsoluteChange)
	finalKeywordQuery = "With keyword_analysis_previous_week as (" + previousWeekAnalysisQuery + "), keyword_analysis_last_week as (" + lastWeekAnalysisQuery + ") " + weeklyDifferenceAnalysisQuery

	var keywordAnalysisResult []model.KeywordAnalysis
	db := C.GetServices().Db
	rows, tx, err, reqID := store.ExecQueryWithContext(finalKeywordQuery, params)
	if err != nil {
		return make([]model.KeywordAnalysis, 0), err
	}
	defer U.CloseReadQuery(rows, tx)

	startReadTime := time.Now()
	for rows.Next() {
		var keywordAnalysisRow model.KeywordAnalysis
		err := db.ScanRows(rows, &keywordAnalysisRow)
		if err != nil {
			return make([]model.KeywordAnalysis, 0), err
		}
		keywordAnalysisResult = append(keywordAnalysisResult, keywordAnalysisRow)
	}
	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &logFields)

	cleanKeywordAnalysisResult := model.SanitiseKeywordsAnalysisResult(query, keywordAnalysisResult)
	sort.SliceStable(cleanKeywordAnalysisResult, func(i, j int) bool {
		return cleanKeywordAnalysisResult[i].AbsoluteChange > cleanKeywordAnalysisResult[j].AbsoluteChange
	})
	return cleanKeywordAnalysisResult, nil
}
func (store *MemSQL) getCampaignLevelDataForTemplates(projectID int64, customerAccountID []string, campaignArray []string, query model.TemplateQuery) ([]model.CampaignAnalysis, error) {
	logFields := log.Fields{
		"query":               query,
		"project_id":          projectID,
		"customer_account_id": customerAccountID,
		"campaign_array":      campaignArray,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	finalCampaignQuery := ""
	lastWeekAnalysisQuery := ""
	previousWeekAnalysisQuery := ""
	weeklyDifferenceAnalysisQuery := ""
	params := make([]interface{}, 0)
	staticParamsForCampaignsTemplates := []interface{}{projectID, customerAccountID,
		model.AdwordsDocumentTypeAlias["campaign_performance_report"]}
	previousWeekAnalysisQuery = fmt.Sprintf(adwordsTemplatesWeeklyCampaignAnalysisQuery, templateMetricsToSelectStatement[query.Metric])
	params = append(params, staticParamsForCampaignsTemplates...)
	params = append(params, query.PrevFrom, query.PrevTo, campaignArray)
	lastWeekAnalysisQuery = fmt.Sprintf(adwordsTemplatesWeeklyCampaignAnalysisQuery, templateMetricsToSelectStatement[query.Metric])
	params = append(params, staticParamsForCampaignsTemplates...)
	params = append(params, query.From, query.To, campaignArray)
	weeklyDifferenceAnalysisQuery = adwordsTemplatesWeeklyDifferenceCampaignAnalysisQueryQuery
	finalCampaignQuery = "With campaign_analysis_previous_week as (" + previousWeekAnalysisQuery + "), campaign_analysis_last_week as (" + lastWeekAnalysisQuery + ") " + weeklyDifferenceAnalysisQuery

	var campaignAnalysisResult []model.CampaignAnalysis
	db := C.GetServices().Db
	rows, tx, err, reqID := store.ExecQueryWithContext(finalCampaignQuery, params)
	if err != nil {
		return make([]model.CampaignAnalysis, 0), err
	}
	defer U.CloseReadQuery(rows, tx)

	startReadTime := time.Now()
	for rows.Next() {
		var campaignAnalysisRow model.CampaignAnalysis
		err := db.ScanRows(rows, &campaignAnalysisRow)
		if err != nil {
			return make([]model.CampaignAnalysis, 0), err
		}
		campaignAnalysisResult = append(campaignAnalysisResult, campaignAnalysisRow)
	}
	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &logFields)

	cleanCampaignAnalysisResult := model.SanitiseCampaignAnalysisResult(query, campaignAnalysisResult)
	sort.SliceStable(cleanCampaignAnalysisResult, func(i, j int) bool {
		return cleanCampaignAnalysisResult[i].AbsoluteChange > cleanCampaignAnalysisResult[j].AbsoluteChange
	})
	return cleanCampaignAnalysisResult, nil
}

func (store *MemSQL) getOverallChangesDataForTemplates(projectID int64, customerAccountID []string, query model.TemplateQuery) ([]model.OverallChanges, error) {
	logFields := log.Fields{
		"query":               query,
		"project_id":          projectID,
		"customer_account_id": customerAccountID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	overallChangesResult := make([]model.OverallChanges, 0)
	overallAnalysisQuery := fmt.Sprintf(semChecklistOverallAnalysisQuery, templateMetricsToSelectStatementForOverallAnalysis[query.Metric])
	staticParamsForOverallAnaysis := []interface{}{projectID, customerAccountID,
		model.AdwordsDocumentTypeAlias["campaign_performance_report"]}
	paramsPrevWeek := make([]interface{}, 0)
	paramsPrevWeek = append(paramsPrevWeek, staticParamsForOverallAnaysis...)
	paramsPrevWeek = append(paramsPrevWeek, query.PrevFrom, query.PrevTo)
	paramsLastWeek := make([]interface{}, 0)
	paramsLastWeek = append(paramsLastWeek, staticParamsForOverallAnaysis...)
	paramsLastWeek = append(paramsLastWeek, query.From, query.To)

	rows, tx, err, queryID1 := store.ExecQueryWithContext(overallAnalysisQuery, paramsLastWeek)
	if err != nil {
		return make([]model.OverallChanges, 0), err
	}

	resultHeadersLastWeek, resultRowsLastWeek, err := U.DBReadRows(rows, tx, queryID1)
	if err != nil {
		return make([]model.OverallChanges, 0), err
	}

	rows, tx, err, queryID2 := store.ExecQueryWithContext(overallAnalysisQuery, paramsPrevWeek)
	if err != nil {
		return make([]model.OverallChanges, 0), err
	}

	_, resultRowsPreviousWeek, err := U.DBReadRows(rows, tx, queryID2)
	if err != nil {
		return make([]model.OverallChanges, 0), err
	}

	for index, value := range resultHeadersLastWeek {
		var percentageChange float64
		var previousValue, lastValue float64
		if resultRowsPreviousWeek[0][index] == nil {
			previousValue = 0
		} else {
			previousValue = resultRowsPreviousWeek[0][index].(float64)
		}
		if resultRowsLastWeek[0][index] == nil {
			lastValue = 0
		} else {
			lastValue = resultRowsLastWeek[0][index].(float64)
		}
		if previousValue == 0 && lastValue == 0 {
			percentageChange = 0
		} else if previousValue == 0 && lastValue != 0 {
			percentageChange = lastValue * 100 / 0.0000001
		} else {
			percentageChange = (lastValue - previousValue) * 100 / previousValue
		}
		var overallChangesData model.OverallChanges
		overallChangesData.Metric = value
		overallChangesData.PercentageChange = percentageChange
		overallChangesData.LastValue = lastValue
		overallChangesData.PreviousValue = previousValue
		if overallChangesData.PreviousValue == 0 && overallChangesData.LastValue != 0 {
			overallChangesData.IsInfinity = true
		}
		overallChangesResult = append(overallChangesResult, overallChangesData)
	}
	return overallChangesResult, nil
}

func (store *MemSQL) getAdwordsSEMChecklistQueryData(query model.TemplateQuery, projectID int64, customerAccountID []string,
	reqID string) (model.TemplateResponse, error) {
	logFields := log.Fields{
		"query":               query,
		"project_id":          projectID,
		"customer_account_id": customerAccountID,
		"req_id":              reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var result model.TemplateResponse
	keywordAnalysisResult, err := store.getKeywordLevelDataForTemplates(projectID, customerAccountID, query)
	if err != nil {
		return model.TemplateResponse{}, err
	}
	campaignIDToSubLevelDataMap := make(map[string][]model.SubLevelData)
	for _, keywordAnalysis := range keywordAnalysisResult {
		// if both previous value and last value are less than 0.1, we ignore and don't show them
		if keywordAnalysis.PreviousWeekValue < 0.1 && keywordAnalysis.LastWeekValue < 0.1 {
			continue
		}
		// transforming db response to api response format
		subLevelData := model.TransformKeywordAnalysisToTemplateSubLevelData(query, keywordAnalysis)
		campaignIDToSubLevelDataMap[keywordAnalysis.CampaignID] = append(campaignIDToSubLevelDataMap[keywordAnalysis.CampaignID], subLevelData)
	}
	campaignArray := make([]string, 0)
	for key := range campaignIDToSubLevelDataMap {
		campaignArray = append(campaignArray, key)
	}
	// to avoid error in campaign analysis, if the keyword analysis doesn't return any campaigns we return empty response
	if len(campaignArray) == 0 {
		return model.TemplateResponse{}, nil
	}
	campaignAnalysisResult, err := store.getCampaignLevelDataForTemplates(projectID, customerAccountID, campaignArray, query)
	if err != nil {
		return model.TemplateResponse{}, err
	}
	for _, campaignAnalysisRow := range campaignAnalysisResult {
		// if both previous value and last value are less than 0.1, we ignore and don't show them
		if campaignAnalysisRow.PreviousWeekValue < 0.1 && campaignAnalysisRow.LastWeekValue < 0.1 {
			continue
		}
		// transforming db response to api response format
		primaryLevelData := model.TransfromCampaignLevelDataToTemplatePrimaryLevelData(campaignAnalysisRow, campaignIDToSubLevelDataMap)
		result.BreakdownAnalysis.PrimaryLevelData = append(result.BreakdownAnalysis.PrimaryLevelData, primaryLevelData)
	}
	overallChangesResult, err := store.getOverallChangesDataForTemplates(projectID, customerAccountID, query)
	if err != nil {
		return model.TemplateResponse{}, err
	}
	result.BreakdownAnalysis.OverallChangesData = overallChangesResult

	result.Meta = model.TemplateResponseMeta{
		PrimaryLevel: model.LevelMeta{
			ColumnName: "campaign",
		},
		SubLevel: model.LevelMeta{
			ColumnName: "keyword",
		},
	}
	return result, nil
}
func (store *MemSQL) DeleteAdwordsIntegration(projectID int64) (int, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	projectSetting := model.ProjectSetting{}

	err := db.Model(&model.ProjectSetting{}).Where("project_id = ?", projectID).Find(&projectSetting).Error
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if projectSetting.IntAdwordsEnabledAgentUUID != nil {
		agentUpdateValues := make(map[string]interface{})
		agentUpdateValues["int_adwords_refresh_token"] = nil
		err = db.Model(&model.Agent{}).Where("uuid = ?", *projectSetting.IntAdwordsEnabledAgentUUID).Update(agentUpdateValues).Error
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	projectSettingUpdateValues := make(map[string]interface{})
	projectSettingUpdateValues["int_adwords_customer_account_id"] = nil
	projectSettingUpdateValues["int_adwords_enabled_agent_uuid"] = nil
	err = db.Model(&model.ProjectSetting{}).Where("project_id = ?", projectID).Update(projectSettingUpdateValues).Error
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// PullAdwordsRows - Function to pull adwords campaign data
// Selecting VALUE, TIMESTAMP, TYPE from adwords_documents and PROPERTIES, OBJECT_TYPE from smart_properties
// Left join smart_properties filtered by project_id and source=adwords
// where adwords_documents.value["campaign_id"] = smart_properties.object_id (when smart_properties.object_type = 1)
//	 or adwords_documents.value["ad_group_id"] = smart_properties.object_id (when smart_properties.object_type = 2)
// [make sure there aren't multiple smart_properties rows for a particular object,
// or weekly insights for adwords would show incorrect data.]
// TODO(anshul) : [all channels]check for index support for faster query
func (store *MemSQL) PullAdwordsRows(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
		"end_time":   endTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	year, month, date := time.Unix(startTime, 0).Date()
	start := year*10000 + int(month)*100 + date + 1

	year, month, date = time.Unix(endTime, 0).Date()
	end := year*10000 + int(month)*100 + date

	rawQuery := fmt.Sprintf("SELECT adwDocs.id, adwDocs.value, adwDocs.timestamp, adwDocs.type, sp.properties FROM adwords_documents adwDocs "+
		"LEFT JOIN smart_properties sp ON sp.project_id = %d AND sp.source = '%s' AND "+
		"((COALESCE(sp.object_type,1) = 1 AND (sp.object_id = JSON_EXTRACT_STRING(adwDocs.value, 'campaign_id') OR sp.object_id = JSON_EXTRACT_STRING(adwDocs.value, 'base_campaign_id'))) OR "+
		"(COALESCE(sp.object_type,2) = 2 AND (sp.object_id = JSON_EXTRACT_STRING(adwDocs.value, 'ad_group_id') OR sp.object_id = JSON_EXTRACT_STRING(adwDocs.value, 'base_ad_group_id')))) "+
		"WHERE adwDocs.project_id = %d AND adwDocs.timestamp BETWEEN %d AND %d "+
		"ORDER BY adwDocs.type, adwDocs.timestamp LIMIT %d",
		projectID, model.ChannelAdwords, projectID, start, end, model.AdwordsPullLimit+1)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}
