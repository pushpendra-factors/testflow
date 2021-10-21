package postgres

import (
	"database/sql"
	"errors"
	C "factors/config"
	Const "factors/constants"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	errorDuplicateAdwordsDocument                   = "pq: duplicate key value violates unique constraint \"adwords_documents_pkey\""
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
	insertAdwordsStr                   = "INSERT INTO adwords_documents (project_id,customer_account_id,type,timestamp,id,campaign_id,ad_group_id,ad_id,keyword_id,value,created_at,updated_at) VALUES "
	adwordsFilterQueryStr              = "SELECT DISTINCT(LOWER(value->>?)) as filter_value FROM adwords_documents WHERE project_id = ? AND customer_account_id IN ( ? ) AND type = ? AND value->>? IS NOT NULL LIMIT 5000"
	staticWhereStatementForAdwords     = "WHERE project_id = ? AND customer_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "
	fromAdwordsDocument                = " FROM adwords_documents "
	shareHigherOrderExpression         = "sum(case when value->>'%s' IS NOT NULL THEN (value->>'%s')::float else 0 END)/NULLIF(sum(case when value->>'%s' IS NOT NULL THEN (value->>'%s')::float else 0 END), 0)"
	higherOrderExpressionsWithMultiply = "SUM(COALESCE((value->>'%s')::float, 0))*%s/(COALESCE( NULLIF(sum(COALESCE((value->>'%s')::float, 0)), 0), 100000))"
	higherOrderExpressionsWithDiv      = "(SUM(COALESCE((value->>'%s')::float, 0))/1000000)/(COALESCE( NULLIF(sum(COALESCE((value->>'%s')::float, 0)), 0), 100000))"
	sumOfFloatExp                      = "sum((value->>'%s')::float)"
	adwordsAdGroupMetdataFetchQueryStr = "select ad_group_id::text, campaign_id::text, value->>'name' as ad_group_name, " +
		"value->>'campaign_name' as campaign_name from adwords_documents where type = ? AND project_id = ? AND " +
		"timestamp BETWEEN ? AND ? AND customer_account_id in (?) " +
		"and (ad_group_id, timestamp) in (select ad_group_id, max(timestamp) from adwords_documents where type = ?" +
		" AND project_id = ? AND timestamp BETWEEN ? AND ? AND customer_account_id in (?) group by ad_group_id)"

	adwordsCampaignMetadataFetchQueryStr = "select campaign_id::text, value->>'name' as campaign_name from adwords_documents where type = ? AND " +
		"project_id = ? AND timestamp BETWEEN ? and ? AND customer_account_id in (?) and (campaign_id, timestamp) in " +
		"(select campaign_id, max(timestamp) from adwords_documents where type = ? " +
		"and project_id = ? and timestamp BETWEEN ? and ? AND customer_account_id in (?) group by campaign_id)"
	adwordsTemplateWeeklyDifferenceAnalysisKeywordsQuerySelectStmnt = "Select %s %s from "
	fixedSelectForBreakdownAnalysisKeyword                          = "ABS((keyword_analysis_last_week.analysis_metric - keyword_analysis_previous_week.analysis_metric)::float) as abs_change, " +
		"(keyword_analysis_last_week.analysis_metric - keyword_analysis_previous_week.analysis_metric)::float as absolute_change, " +
		"keyword_analysis_previous_week.analysis_metric as previous_week_value, keyword_analysis_last_week.analysis_metric as last_week_value "
	adwordsTemplateWeeklyDifferenceAnalysisKeywordsQueryJoinStmnt = "keyword_analysis_last_week full outer join keyword_analysis_previous_week on keyword_analysis_last_week.keyword_id = " +
		"keyword_analysis_previous_week.keyword_id and keyword_analysis_last_week.keyword_match_type=keyword_analysis_previous_week.keyword_match_type " +
		"and keyword_analysis_last_week.campaign_id = keyword_analysis_previous_week.campaign_id and keyword_analysis_last_week.keyword_name " +
		"= keyword_analysis_previous_week.keyword_name "
	adwordsTemplateWeeklyDifferenceAnalysisKeywordsQueryWhereStmnt = "where (ABS((((keyword_analysis_last_week.analysis_metric - " +
		"keyword_analysis_previous_week.analysis_metric)::float)*100/(COALESCE(NULLIF(keyword_analysis_previous_week.analysis_metric::float, 0), 0.0000001)))) >= ? " +
		"AND ABS((keyword_analysis_last_week.analysis_metric - keyword_analysis_previous_week.analysis_metric)) > ?) or " +
		"keyword_analysis_previous_week.analysis_metric is null or keyword_analysis_last_week.analysis_metric is null order by abs_change DESC"

	adwordsTemplatesWeeklyKeywordAnalysisQuery = "select %s as analysis_metric, %s keyword_id, campaign_id, value->>'criteria' as keyword_name, " +
		"value->>'keyword_match_type' as keyword_match_type from adwords_documents where project_id = ? and customer_account_id " +
		"in (?) and type = ? and timestamp between ? AND ? and lower(value->>'criteria') != 'automaticcontent' and lower(value->>'criteria') != 'automatickeywords'" +
		" group by campaign_id, keyword_id, keyword_name, keyword_match_type"

	adwordsTemplatesWeeklyCampaignAnalysisQuery = "select %s as analysis_metric, campaign_id, value->>'campaign_name' " +
		"as campaign_name from adwords_documents where project_id = ? and customer_account_id in (?) and type = ? and timestamp " +
		"between ? AND ? and campaign_id in (?) and value->>'advertising_channel_type' ILIKE 'search' group by campaign_id, campaign_name"
	adwordsTemplatesWeeklyDifferenceCampaignAnalysisQueryQuery = " Select coalesce(campaign_analysis_last_week.campaign_name, campaign_analysis_previous_week.campaign_name) as campaign_name, " +
		"campaign_analysis_previous_week.analysis_metric as previous_week_value, campaign_analysis_last_week.analysis_metric as last_week_value, " +
		"(((campaign_analysis_last_week.analysis_metric - campaign_analysis_previous_week.analysis_metric)::float)*100/(COALESCE(NULLIF(campaign_analysis_previous_week.analysis_metric::float, 0), 0.0000001))) as percentage_change, " +
		"ABS((campaign_analysis_last_week.analysis_metric - campaign_analysis_previous_week.analysis_metric)::float) as abs_change, " +
		"(campaign_analysis_last_week.analysis_metric - campaign_analysis_previous_week.analysis_metric)::float as absolute_change, " +
		"coalesce(campaign_analysis_last_week.campaign_id, campaign_analysis_previous_week.campaign_id) as campaign_id from campaign_analysis_last_week " +
		"full outer join campaign_analysis_previous_week on campaign_analysis_last_week.campaign_id = campaign_analysis_previous_week.campaign_id " +
		"order by abs_change DESC limit 10000"
	semChecklistOverallAnalysisQuery = "select %s from adwords_documents " +
		"where project_id = ? and customer_account_id in (?) and type = ? and timestamp between ? AND ? AND lower(value->>'advertising_channel_type') = 'search'"
	adwordsTemplatesExtraSelectWeekAnalysisForRCA      = "%s as impressions, %s as search_impression_share, %s as conversion_rate, %s as click_through_rate, %s as cost_per_click, "
	adwordsTemplatesExtraSelectBreakdownAnalysisForRCA = "%s as impressions, %s as search_impression_share, %s as conversion_rate, %s as click_through_rate, %s as cost_per_click, " +
		"%s as prev_impressions, %s as prev_search_impression_share, %s as prev_conversion_rate,%s as prev_click_through_rate, %s as prev_cost_per_click, " +
		"%s as last_impressions, %s as last_search_impression_share, %s as last_conversion_rate,%s as last_click_through_rate, %s as last_cost_per_click, "
	percentageChangeForSemChecklist = "(((%s.%s - %s.%s)::float)*100/(COALESCE(NULLIF(%s.%s::float, 0), 0.0000001)))"
	coalesceForChecklists           = "COALESCE(%s.%s, %s.%s) as %s"
)

var templateMetricsToSelectStatement = map[string]string{
	model.Clicks:                "sum((value->>'clicks')::float)",
	model.Impressions:           "sum((value->>'impressions')::float)",
	model.ClickThroughRate:      fmt.Sprintf(higherOrderExpressionsWithMultiply, "clicks", "100", "impressions"),
	model.CostPerClick:          fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "clicks"),
	model.SearchImpressionShare: fmt.Sprintf(shareHigherOrderExpression, model.SearchImpressionShare, model.Impressions, model.SearchImpressionShare, model.TotalSearchImpression),
	"cost":                      "sum((value->>'cost')::float)/1000000",
	model.Conversion:            "sum((value->>'conversions')::float)",
	"cost_per_lead":             fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions"),
	"click_to_lead_rate":        fmt.Sprintf(higherOrderExpressionsWithMultiply, "conversions", "100", "clicks"),
	model.ConversionRate:        fmt.Sprintf(higherOrderExpressionsWithMultiply, "conversions", "100", "clicks"),
}

var templateMetricsToSelectStatementForOverallAnalysis = map[string]string{
	model.Clicks:                "sum((value->>'clicks')::float) as clicks, sum((value->>'conversions')::float) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	model.Impressions:           "sum((value->>'impressions')::float) as impressions, sum((value->>'conversions')::float) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	model.ClickThroughRate:      fmt.Sprintf(higherOrderExpressionsWithMultiply, "clicks", "100", "impressions") + fmt.Sprintf("as %s, ", model.ClickThroughRate) + "sum((value->>'conversions')::float) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	model.CostPerClick:          fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "clicks") + fmt.Sprintf("as %s, ", model.CostPerClick) + "sum((value->>'conversions')::float) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	model.SearchImpressionShare: fmt.Sprintf(shareHigherOrderExpression, model.SearchImpressionShare, model.Impressions, model.SearchImpressionShare, model.TotalSearchImpression) + fmt.Sprintf("as %s, ", model.SearchImpressionShare) + "sum((value->>'conversions')::float) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	"cost":                      "sum((value->>'cost')::float)/1000000 as cost, sum((value->>'conversions')::float) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	model.Conversion:            "sum((value->>'conversions')::float) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	"cost_per_lead":             fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + fmt.Sprintf(" as %s, ", "cost_per_lead") + "sum((value->>'conversions')::float) as conversion ",
	"click_to_lead_rate":        fmt.Sprintf(higherOrderExpressionsWithMultiply, "conversions", "100", "clicks") + fmt.Sprintf("as %s, ", "click_to_lead_rate") + "sum((value->>'conversions')::float) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
	model.ConversionRate:        fmt.Sprintf(higherOrderExpressionsWithMultiply, "conversions", "100", "clicks") + fmt.Sprintf("as %s, ", model.ConversionRate) + "sum((value->>'conversions')::float) as conversion, " + fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions") + " as cost_per_lead",
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
	model.Impressions: {},
	model.Clicks:      {},
	"cost":            {},
	"conversions":     {},
}

// Later Handle divide by zero separately.
// Same structure is being used for internal operations and external.
var adwordsInternalMetricsToAllRep = map[string]metricsAndRelated{
	model.Impressions: {
		nonHigherOrderExpression: "sum((value->>'impressions')::float)",
		externalValue:            model.Impressions,
		externalOperation:        "sum",
	},
	model.Clicks: {
		nonHigherOrderExpression: "sum((value->>'clicks')::float)",
		externalValue:            model.Clicks,
		externalOperation:        "sum",
	},
	"cost": {
		nonHigherOrderExpression: "sum((value->>'cost')::float)/1000000",
		externalValue:            "spend",
		externalOperation:        "sum",
	},
	"conversions": {
		nonHigherOrderExpression: "sum((value->>'conversions')::float)",
		externalValue:            model.Conversion,
		externalOperation:        "sum",
	},
	model.ClickThroughRate: {
		higherOrderExpression:    fmt.Sprintf(higherOrderExpressionsWithMultiply, "clicks", "100", "impressions"),
		nonHigherOrderExpression: "sum((value->>'clicks')::float)*100",
		externalValue:            model.ClickThroughRate,
		externalOperation:        "sum",
	},
	model.ConversionRate: {
		higherOrderExpression:    fmt.Sprintf(higherOrderExpressionsWithMultiply, "conversions", "100", "clicks"),
		nonHigherOrderExpression: "sum((value->>'conversions')::float)*100",
		externalValue:            model.ConversionRate,
		externalOperation:        "sum",
	},
	model.CostPerClick: {
		higherOrderExpression:    fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "clicks"),
		nonHigherOrderExpression: "(sum((value->>'cost')::float)/1000000)",
		externalValue:            model.CostPerClick,
		externalOperation:        "sum",
	},
	model.CostPerConversion: {
		higherOrderExpression:    fmt.Sprintf(higherOrderExpressionsWithDiv, "cost", "conversions"),
		nonHigherOrderExpression: "(sum((value->>'cost')::float)/1000000)",
		externalValue:            model.CostPerConversion,
		externalOperation:        "sum",
	},
	model.SearchImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, model.SearchImpressionShare, model.Impressions, model.SearchImpressionShare, model.TotalSearchImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchImpression),
		externalValue:            model.SearchImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchClickShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, model.SearchClickShare, model.Impressions, model.SearchClickShare, model.TotalSearchClick),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchClick),
		externalValue:            model.SearchClickShare,
		externalOperation:        "sum",
	},
	model.SearchTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, model.SearchTopImpressionShare, model.Impressions, model.SearchTopImpressionShare, model.TotalSearchTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchTopImpression),
		externalValue:            model.SearchTopImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchAbsoluteTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, model.SearchAbsoluteTopImpressionShare, model.Impressions, model.SearchAbsoluteTopImpressionShare, model.TotalSearchAbsoluteTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchAbsoluteTopImpression),
		externalValue:            model.SearchAbsoluteTopImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchBudgetLostAbsoluteTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, model.SearchBudgetLostAbsoluteTopImpressionShare, model.Impressions, model.SearchBudgetLostAbsoluteTopImpressionShare, model.TotalSearchBudgetLostAbsoluteTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchBudgetLostAbsoluteTopImpression),
		externalValue:            model.SearchBudgetLostAbsoluteTopImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchBudgetLostImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, model.SearchBudgetLostImpressionShare, model.Impressions, model.SearchBudgetLostImpressionShare, model.TotalSearchBudgetLostImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchBudgetLostImpression),
		externalValue:            model.SearchBudgetLostImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchBudgetLostTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, model.SearchBudgetLostTopImpressionShare, model.Impressions, model.SearchBudgetLostTopImpressionShare, model.TotalSearchBudgetLostTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchBudgetLostTopImpression),
		externalValue:            model.SearchBudgetLostTopImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchRankLostAbsoluteTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, model.SearchRankLostAbsoluteTopImpressionShare, model.Impressions, model.SearchRankLostAbsoluteTopImpressionShare, model.TotalSearchRankLostAbsoluteTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchRankLostAbsoluteTopImpression),
		externalValue:            model.SearchRankLostAbsoluteTopImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchRankLostImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, model.SearchRankLostImpressionShare, model.Impressions, model.SearchRankLostImpressionShare, model.TotalSearchRankLostImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchRankLostImpression),
		externalValue:            model.SearchRankLostImpressionShare,
		externalOperation:        "sum",
	},
	model.SearchRankLostTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, model.SearchRankLostTopImpressionShare, model.Impressions, model.SearchRankLostTopImpressionShare, model.TotalSearchRankLostTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, model.TotalSearchRankLostTopImpression),
		externalValue:            model.SearchRankLostTopImpressionShare,
		externalOperation:        "sum",
	},
}

type fields struct {
	selectExpressions []string
	values            []string
}

func isDuplicateAdwordsDocumentError(err error) bool {
	return err.Error() == errorDuplicateAdwordsDocument
}

func getAdwordsIDFieldNameByType(docType int) string {
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
	// Todo: Add timezone support using util.getTimeFromUnixTimestampWithZone.
	return time.Unix(unixTimestamp, 0).UTC().Format("20060102")
}

func getAdwordsIDAndHeirarchyColumnsByType(docType int, valueJSON *postgres.Jsonb) (string, int64, int64, int64, int64, error) {
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

	// KeyID as string always.
	return idStr, value1, value2, value3, value4, nil
}

// CreateAdwordsDocument ...
func (pg *Postgres) CreateAdwordsDocument(adwordsDoc *model.AdwordsDocument) int {
	status := validateAdwordsDocument(adwordsDoc)
	if status != http.StatusOK {
		return status
	}

	status = addColumnInformationForAdwordsDocument(adwordsDoc)
	if status != http.StatusOK {
		return status
	}

	db := C.GetServices().Db
	dbc := db.Create(adwordsDoc)

	if dbc.Error != nil {
		if isDuplicateAdwordsDocumentError(dbc.Error) {
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
func (pg *Postgres) CreateMultipleAdwordsDocument(adwordsDocuments []model.AdwordsDocument) int {
	status := validateAdwordsDocuments(adwordsDocuments)
	if status != http.StatusOK {
		return status
	}
	adwordsDocuments, status = addColumnInformationForAdwordsDocuments(adwordsDocuments)
	if status != http.StatusOK {
		return status
	}

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
		if isDuplicateAdwordsDocumentError(err) {
			log.WithError(err).WithField("adwordsDocuments", adwordsDocuments).Error("Failed to create an adwords doc. Duplicate.")
			return http.StatusConflict
		} else {
			log.WithError(err).WithField("adwordsDocuments", adwordsDocuments).Error(
				"Failed to create an adwords doc. Continued inserting other docs.")
			return http.StatusInternalServerError
		}
	}
	defer rows.Close()

	return http.StatusCreated
}

func validateAdwordsDocuments(adwordsDocuments []model.AdwordsDocument) int {
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
	logCtx := log.WithField("customer_acc_id", adwordsDocument.CustomerAccountID).WithField(
		"project_id", adwordsDocument.ProjectID)
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
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range model.AdwordsDocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

func (pg *Postgres) GetAdwordsLastSyncInfoForProject(projectID uint64) ([]model.AdwordsLastSyncInfo, int) {
	params := []interface{}{projectID}
	adwordsLastSyncInfos, status := getAdwordsLastSyncInfo(lastSyncInfoForAProject, params)
	if status != http.StatusOK {
		return adwordsLastSyncInfos, status
	}
	adwordsSettings, errCode := pg.GetIntAdwordsProjectSettingsForProjectID(projectID)
	if errCode != http.StatusOK {
		return []model.AdwordsLastSyncInfo{}, errCode
	}

	return pg.sanitizedLastSyncInfos(adwordsLastSyncInfos, adwordsSettings)
}

// GetAllAdwordsLastSyncInfoForAllProjects - @TODO Kark v1
func (pg *Postgres) GetAllAdwordsLastSyncInfoForAllProjects() ([]model.AdwordsLastSyncInfo, int) {
	params := make([]interface{}, 0, 0)
	adwordsLastSyncInfos, status := getAdwordsLastSyncInfo(lastSyncInfoQueryForAllProjects, params)
	if status != http.StatusOK {
		return adwordsLastSyncInfos, status
	}

	adwordsSettings, errCode := pg.GetAllIntAdwordsProjectSettings()
	if errCode != http.StatusOK {
		return []model.AdwordsLastSyncInfo{}, errCode
	}

	return pg.sanitizedLastSyncInfos(adwordsLastSyncInfos, adwordsSettings)
}

func getAdwordsLastSyncInfo(query string, params []interface{}) ([]model.AdwordsLastSyncInfo, int) {
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
func (pg *Postgres) sanitizedLastSyncInfos(adwordsLastSyncInfos []model.AdwordsLastSyncInfo, adwordsSettings []model.AdwordsProjectSettings) ([]model.AdwordsLastSyncInfo, int) {

	adwordsSettingsByProjectAndCustomerAccount := make(map[uint64]map[string]*model.AdwordsProjectSettings, 0)
	projectIDs := make([]uint64, 0, 0)

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
	existingProjectAndCustomerAccountWithTypes := make(map[uint64]map[string]map[string]bool, 0)
	selectedLastSyncInfos := make([]model.AdwordsLastSyncInfo, 0, 0)

	for i := range adwordsLastSyncInfos {
		logCtx := log.WithFields(
			log.Fields{"project_id": adwordsLastSyncInfos[i].ProjectId,
				"customer_account_id": adwordsLastSyncInfos[i].CustomerAccountId})

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
				if !accountExists || (accountExists && existingTypesForAccount[docTypeAlias] == false) {
					syncInfo := model.AdwordsLastSyncInfo{
						ProjectId:         adwordsSettings[i].ProjectId,
						RefreshToken:      adwordsSettings[i].RefreshToken,
						CustomerAccountId: accountID,
						LastTimestamp:     0, // no sync yet.
						DocumentTypeAlias: docTypeAlias,
						Timezone:          adwordsSettings[i].IntGoogleIngestionTimezone,
					}

					selectedLastSyncInfos = append(selectedLastSyncInfos, syncInfo)
				}
			}
		}

	}

	projects, _ := pg.GetProjectsByIDs(projectIDs)
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

// PullGCLIDReport - It returns GCLID based campaign info for given time range and adwords account
func (pg *Postgres) PullGCLIDReport(projectID uint64, from, to int64, adwordsAccountIDs string,
	campaignIDReport, adgroupIDReport, keywordIDReport map[string]model.MarketingData, timeZone string) (map[string]model.MarketingData, error) {

	logCtx := log.WithFields(log.Fields{"ProjectID": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})
	adGroupNameCase := "CASE WHEN value->>'ad_group_name' IS NULL THEN ? " +
		" WHEN value->>'ad_group_name' = '' THEN ? ELSE value->>'ad_group_name' END AS ad_group_name"
	adGroupIDCase := "CASE WHEN value->>'ad_group_id' IS NULL THEN ? " +
		" WHEN value->>'ad_group_id' = '' THEN ? ELSE value->>'ad_group_id' END AS ad_group_id"
	campaignNameCase := "CASE WHEN value->>'campaign_name' IS NULL THEN ? " +
		" WHEN value->>'campaign_name' = '' THEN ? ELSE value->>'campaign_name' END AS campaign_name"
	campaignIDCase := "CASE WHEN value->>'campaign_id' IS NULL THEN ? " +
		" WHEN value->>'campaign_id' = '' THEN ? ELSE value->>'campaign_id' END AS campaign_id"
	adIDCase := "CASE WHEN value->>'creative_id' IS NULL THEN ? " +
		" WHEN value->>'creative_id' = '' THEN ? ELSE value->>'creative_id' END AS creative_id"
	keywordNameCase := "CASE WHEN value->>'criteria_name' IS NULL THEN ? " +
		" WHEN value->>'criteria_name' = '' THEN ? ELSE value->>'criteria_name' END AS criteria_name"
	keywordIDCase := "CASE WHEN value->>'criteria_id' IS NULL THEN ? " +
		" WHEN value->>'criteria_id' = '' THEN ? ELSE value->>'criteria_id' END AS criteria_id"
	slotCase := "CASE WHEN value->>'slot' IS NULL THEN ? " +
		" WHEN value->>'slot' = '' THEN ? ELSE value->>'slot' END AS slot"

	performanceQuery := "SELECT id, " + adGroupNameCase + ", " + adGroupIDCase + ", " + campaignNameCase + ", " +
		campaignIDCase + ", " + adIDCase + ", " + keywordNameCase + ", " + keywordIDCase + ", " + slotCase +
		" FROM adwords_documents where project_id = ? AND customer_account_id IN (?) AND type = ? AND timestamp between ? AND ? "
	customerAccountIDs := strings.Split(adwordsAccountIDs, ",")

	params := []interface{}{model.PropertyValueNone, model.PropertyValueNone,
		model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone,
		model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone,
		model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone,
		model.PropertyValueNone, model.PropertyValueNone,
		projectID, customerAccountIDs, model.AdwordsClickReportType, U.GetDateAsStringIn(from, U.TimeZoneString(timeZone)),
		U.GetDateAsStringIn(to, U.TimeZoneString(timeZone))}
	rows, tx, err := pg.ExecQueryWithContext(performanceQuery, params)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, err
	}
	defer U.CloseReadQuery(rows, tx)
	gclidBasedMarketData := make(map[string]model.MarketingData)
	for rows.Next() {
		var gclIDTmp sql.NullString
		var adgroupNameTmp sql.NullString
		var adgroupIDTmp sql.NullString
		var campaignNameTmp sql.NullString
		var campaignIDTmp sql.NullString
		var adIDTmp sql.NullString
		var keywordNameTmp sql.NullString
		var keywordIDTmp sql.NullString
		var slotTmp sql.NullString
		if err = rows.Scan(&gclIDTmp, &adgroupNameTmp, &adgroupIDTmp, &campaignNameTmp, &campaignIDTmp, &adIDTmp, &keywordNameTmp, &keywordIDTmp, &slotTmp); err != nil {
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
		var slot string
		gclID = gclIDTmp.String
		adgroupName = U.IfThenElse(adgroupNameTmp.Valid == true, adgroupNameTmp.String, model.PropertyValueNone).(string)
		adgroupID = U.IfThenElse(adgroupIDTmp.Valid == true, adgroupIDTmp.String, model.PropertyValueNone).(string)
		campaignName = U.IfThenElse(campaignNameTmp.Valid == true, campaignNameTmp.String, model.PropertyValueNone).(string)
		campaignID = U.IfThenElse(campaignIDTmp.Valid == true, campaignIDTmp.String, model.PropertyValueNone).(string)
		adID = U.IfThenElse(adIDTmp.Valid == true, adIDTmp.String, model.PropertyValueNone).(string)
		keywordName = U.IfThenElse(keywordNameTmp.Valid == true, keywordNameTmp.String, model.PropertyValueNone).(string)
		keywordID = U.IfThenElse(keywordIDTmp.Valid == true, keywordIDTmp.String, model.PropertyValueNone).(string)
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
	return gclidBasedMarketData, nil
}

// @TODO Kark v1
func (pg *Postgres) buildAdwordsChannelConfig(projectID uint64) *model.ChannelConfigResult {
	adwordsObjectsAndProperties := pg.buildObjectAndPropertiesForAdwords(projectID, model.ObjectsForAdwords)
	selectMetrics := append(selectableMetricsForAllChannels, model.SelectableMetricsForAdwords...)
	objectsAndProperties := adwordsObjectsAndProperties
	return &model.ChannelConfigResult{
		SelectMetrics:        selectMetrics,
		ObjectsAndProperties: objectsAndProperties,
	}
}

func (pg *Postgres) buildObjectAndPropertiesForAdwords(projectID uint64, objects []string) []model.ChannelObjectAndProperties {
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0, 0)
	for _, currentObject := range objects {
		// to do: check if normal properties present then only smart properties will be there
		propertiesAndRelated, isPresent := model.MapOfAdwordsObjectsToPropertiesAndRelated[currentObject]
		var currentProperties []model.ChannelProperty
		var currentPropertiesSmart []model.ChannelProperty
		if isPresent {
			currentProperties = buildProperties(propertiesAndRelated)
			smartProperty := pg.GetSmartPropertyAndRelated(projectID, currentObject, "adwords")
			currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		} else {
			currentProperties = buildProperties(allChannelsPropertyToRelated)
			smartProperty := pg.GetSmartPropertyAndRelated(projectID, currentObject, "adwords")
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
func (pg *Postgres) GetAdwordsFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
	_, isPresent := model.AdwordsExtToInternal[requestFilterProperty]
	if !isPresent {
		filterValues, errCode := pg.getSmartPropertyFilterValues(projectID, requestFilterObject, requestFilterProperty, "adwords", reqID)
		if errCode != http.StatusFound {
			return []interface{}{}, http.StatusInternalServerError
		}
		return filterValues, http.StatusFound
	}
	adwordsInternalFilterProperty, docType, err := getFilterRelatedInformationForAdwords(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return make([]interface{}, 0, 0), http.StatusBadRequest
	}
	filterValues, errCode := pg.getAdwordsFilterValuesByType(projectID, docType, adwordsInternalFilterProperty, reqID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}
func (pg *Postgres) getSmartPropertyFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, source string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID).WithField("smart_property_name", requestFilterProperty)
	objectType, isExists := model.SmartPropertyRulesTypeAliasToType[requestFilterObject]
	if !isExists {
		logCtx.Error("Invalid filter object")
		return make([]interface{}, 0, 0), http.StatusBadRequest
	}
	smartPropertyRule := model.SmartPropertyRules{}
	filterValues := make([]interface{}, 0, 0)
	db := C.GetServices().Db
	err := db.Table("smart_property_rules").Where("project_id = ? AND type = ? AND name = ?", projectID, objectType, requestFilterProperty).Find(&smartPropertyRule).Error
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

// GetAdwordsSQLQueryAndParametersForFilterValues - @TODO Kark v1
// Currently, properties in object dont vary with Object.
func (pg *Postgres) GetAdwordsSQLQueryAndParametersForFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int) {
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	adwordsInternalFilterProperty, docType, err := getFilterRelatedInformationForAdwords(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return "", make([]interface{}, 0, 0), http.StatusBadRequest
	}
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("failed to fetch Project Setting in adwords filter values.")
		return "", []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId
	if customerAccountID == nil || len(*customerAccountID) == 0 {
		logCtx.Info(integrationNotAvailable)
		return "", []interface{}{}, http.StatusInternalServerError
	}
	customerAccountIDs := strings.Split(*customerAccountID, ",")

	params := []interface{}{adwordsInternalFilterProperty, projectID, customerAccountIDs, docType, adwordsInternalFilterProperty}
	return "(" + adwordsFilterQueryStr + ")", params, http.StatusFound
}

func getFilterRelatedInformationForAdwords(requestFilterObject string, requestFilterProperty string) (string, int, int) {
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
func (pg *Postgres) getAdwordsFilterValuesByType(projectID uint64, docType int, property string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID)
	projectSetting, errCode := pg.GetProjectSetting(projectID)
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

	logCtx = log.WithField("doc_type", docType)
	params := []interface{}{property, projectID, customerAccountIDs, docType, property}
	_, resultRows, err := pg.ExecuteSQL(adwordsFilterQueryStr, params, logCtx)
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
// else condition used when there is groupby timestamp, used to extract top 100 values and use them to get data with date.
func (pg *Postgres) ExecuteAdwordsChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	fetchSource := false
	logCtx := log.WithField("xreq_id", reqID)
	if query.GroupByTimestamp == "" {
		sql, params, selectKeys, selectMetrics, errCode := pg.GetSQLQueryAndParametersForAdwordsQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 10000", false, nil)
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := pg.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("channel_query", *query).WithField("query", sql).WithField("params", params).Error(model.AdwordsSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	} else {
		sql, params, selectKeys, selectMetrics, errCode := pg.GetSQLQueryAndParametersForAdwordsQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 100", false, nil)
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := pg.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("channel_query", *query).WithField("query", sql).WithField("params", params).Error(model.AdwordsSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		groupByCombinations := model.GetGroupByCombinationsForChannelAnalytics(columns, resultMetrics)
		sql, params, selectKeys, selectMetrics, errCode = pg.GetSQLQueryAndParametersForAdwordsQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 10000", true, groupByCombinations)
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err = pg.ExecuteSQL(sql, params, logCtx)
		columns = append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("channel_query", *query).WithField("query", sql).WithField("params", params).Error(model.AdwordsSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	}

}

// GetSQLQueryAndParametersForAdwordsQueryV1 - @Kark TODO v1
// TODO Understand null cases.
func (pg *Postgres) GetSQLQueryAndParametersForAdwordsQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, int) {
	var selectMetrics []string
	var selectKeys []string
	var sql string
	var params []interface{}
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	transformedQuery, customerAccountID, err := pg.transFormRequestFieldsAndFetchRequiredFieldsForAdwords(projectID, *query, reqID)
	if projectID == 595 {
		logCtx.WithField("query", transformedQuery).Error("Testing for query failure1")
	}

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
func (pg *Postgres) transFormRequestFieldsAndFetchRequiredFieldsForAdwords(projectID uint64, query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, *string, error) {
	var transformedQuery model.ChannelQueryV1
	logCtx := log.WithField("req_id", reqID)
	var err error
	projectSetting, errCode := pg.GetProjectSetting(projectID)
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
	sortedGroupBys := make([]model.ChannelGroupBy, 0, 0)
	for _, groupBy := range requestGroupBys {
		if groupBy.Object == CAFilterCampaign {
			sortedGroupBys = append(sortedGroupBys, groupBy)
		}
	}

	for _, groupBy := range requestGroupBys {
		if groupBy.Object == CAFilterAdGroup {
			sortedGroupBys = append(sortedGroupBys, groupBy)
		}
	}

	for _, groupBy := range requestGroupBys {
		if groupBy.Object == CAFilterAd {
			sortedGroupBys = append(sortedGroupBys, groupBy)
		}
	}

	for _, groupBy := range requestGroupBys {
		if groupBy.Object == CAFilterKeyword {
			sortedGroupBys = append(sortedGroupBys, groupBy)
		}
	}

	resultGroupBys := make([]model.ChannelGroupBy, 0, 0)
	for _, requestGroupBy := range sortedGroupBys {
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
	// Fetch the propertyNames
	return getLowestHierarchyLevelForAdwordsFiltersAndGroupBy(query.Filters, query.GroupBy)
}

// @TODO Kark v1
func getLowestHierarchyLevelForAdwordsFiltersAndGroupBy(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) string {
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
SELECT value->>'campaign_name' as campaign_name, date_trunc('day', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE 'UTC') as datetime,
SUM((value->>'impressions')::float) as impressions, SUM((value->>'clicks')::float) as clicks FROM adwords_documents WHERE project_id = '2' AND
customer_account_id IN ( '2368493227' ) AND type = '5' AND timestamp between '20200331' AND '20200401'
AND value->>'campaign_name' ILIKE '%Brand - BLR - New_Aug_Desktop_RLSA%' GROUP BY campaign_name, datetime
ORDER BY impressions DESC, clicks DESC LIMIT 2500 ;
*/
// - For reference of complex joins, PR which removed older/QueryV1 adwords is 1437.
func buildAdwordsSimpleQueryV2(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, reqID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string) {
	lowestHierarchyLevel := getLowestHierarchyLevelForAdwords(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_performance_report"
	return getSQLAndParamsForAdwordsV2(query, projectID, query.From, query.To, customerAccountID, model.AdwordsDocumentTypeAlias[lowestHierarchyReportLevel], fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
}

func buildAdwordsSimpleQueryWithSmartPropertyV2(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, reqID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string) {
	lowestHierarchyLevel := getLowestHierarchyLevelForAdwords(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_performance_report"
	return getSQLAndParamsForAdwordsWithSmartPropertyV2(query, projectID, query.From, query.To, customerAccountID, model.AdwordsDocumentTypeAlias[lowestHierarchyReportLevel], fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
}

func getSQLAndParamsForAdwordsWithSmartPropertyV2(query *model.ChannelQueryV1, projectID uint64, from, to int64, customerAccountID string,
	docType int, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string) {
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

	smartPropertyCampaignGroupBys := make([]model.ChannelGroupBy, 0, 0)
	smartPropertyAdGroupGroupBys := make([]model.ChannelGroupBy, 0, 0)
	adwordsGroupBys := make([]model.ChannelGroupBy, 0, 0)

	for _, groupBy := range query.GroupBy {
		_, isPresent := Const.SmartPropertyReservedNames[groupBy.Property]
		if !isPresent {
			if groupBy.Object == "campaign" {
				smartPropertyCampaignGroupBys = append(smartPropertyCampaignGroupBys, groupBy)
			} else {
				smartPropertyAdGroupGroupBys = append(smartPropertyAdGroupGroupBys, groupBy)
			}
		} else {
			adwordsGroupBys = append(adwordsGroupBys, groupBy)
		}
	}
	if projectID == 595 {
		log.WithField("query", *query).WithField("adwordsGroupBys", adwordsGroupBys).WithField("smartPropertyAdGroupGroupBys", smartPropertyAdGroupGroupBys).
			WithField("smartPropertyCampaignGroupBys", smartPropertyCampaignGroupBys).Error("Testing for query failure1")
	}

	// Group By
	dimensions := fields{}

	for _, groupBy := range adwordsGroupBys {
		key := groupBy.Object + ":" + groupBy.Property
		internalValue := model.AdwordsInternalPropertiesToReportsInternal[key]
		externalValue := groupBy.Object + "_" + groupBy.Property
		var expression string
		if groupBy.Property == "id" {
			expression = fmt.Sprintf("%s as %s", internalValue, externalValue)
		} else if _, ok := propertiesToBeDividedByMillion[groupBy.Property]; ok {
			expression = fmt.Sprintf("((value->>'%s')::float)/1000000 as %s", internalValue, externalValue)
		} else {
			expression = fmt.Sprintf("value->>'%s' as %s", internalValue, externalValue)
		}
		dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
		dimensions.values = append(dimensions.values, externalValue)
	}
	for _, groupBy := range smartPropertyCampaignGroupBys {
		expression := fmt.Sprintf(`%s as %s`, fmt.Sprintf("campaign.properties->>'%s'", groupBy.Property), model.CampaignPrefix+groupBy.Property)
		dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
		dimensions.values = append(dimensions.values, model.CampaignPrefix+groupBy.Property)
	}
	for _, groupBy := range smartPropertyAdGroupGroupBys {
		expression := fmt.Sprintf(`%s as %s`, fmt.Sprintf("ad_group.properties->>'%s'", groupBy.Property), model.AdgroupPrefix+groupBy.Property)
		dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
		dimensions.values = append(dimensions.values, model.AdgroupPrefix+groupBy.Property)
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
	filterStatementForSmartPropertyGroupBy := getNotNullFilterStatementForSmartPropertyGroupBys(smartPropertyCampaignGroupBys, smartPropertyAdGroupGroupBys)
	finalWhereStatement = joinWithWordInBetween("AND", staticWhereStatementForAdwordsWithSmartProperty, filterPropertiesStatementBasedOnRequestFilters, filterStatementForSmartPropertyGroupBy)
	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, filterParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForAdwords(groupByCombinationsForGBT)
		finalWhereStatement += (" AND (" + whereConditionForGBT + ")")
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
func getAdwordsFromStatementWithJoins(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) string {
	isPresentCampaignSmartProperty, isPresentAdGroupSmartProperty := checkSmartPropertyWithTypeAndSource(filters, groupBys, "adwords")
	fromStatement := fromAdwordsDocument
	if isPresentAdGroupSmartProperty {
		fromStatement += "inner join smart_properties ad_group on ad_group.project_id = adwords_documents.project_id and ad_group.object_id = ad_group_id::text "
	}
	if isPresentCampaignSmartProperty {
		fromStatement += "inner join smart_properties campaign on campaign.project_id = adwords_documents.project_id and campaign.object_id = campaign_id::text "
	}
	return fromStatement
}
func getSQLAndParamsForAdwordsV2(query *model.ChannelQueryV1, projectID uint64, from, to int64, customerAccountID string,
	docType int, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string) {
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
		key := groupBy.Object + ":" + groupBy.Property
		internalValue := model.AdwordsInternalPropertiesToReportsInternal[key]
		externalValue := groupBy.Object + "_" + groupBy.Property
		var expression string
		if groupBy.Property == "id" {
			expression = fmt.Sprintf("%s as %s", internalValue, externalValue)
		} else if _, ok := propertiesToBeDividedByMillion[groupBy.Property]; ok {
			expression = fmt.Sprintf("((value->>'%s')::float)/1000000 as %s", internalValue, externalValue)
		} else {
			expression = fmt.Sprintf("value->>'%s' as %s", internalValue, externalValue)
		}
		dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
		dimensions.values = append(dimensions.values, externalValue)
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
		finalWhereStatement += (" AND (" + whereConditionForGBT + ")")
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
			filterValue = fmt.Sprintf("%%%s%%", filter.Value)
		} else {
			filterValue = filter.Value
		}
		key := fmt.Sprintf("%s:%s", filter.Object, filter.Property)
		currentFilterProperty = model.AdwordsInternalPropertiesToReportsInternal[key]
		if strings.Contains(filter.Property, ("id")) {
			currentFilterStatement = fmt.Sprintf("%s %s ?", currentFilterProperty, filterOperator)
		} else {
			currentFilterStatement = fmt.Sprintf("value->>'%s' %s ?", currentFilterProperty, filterOperator)
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

// this adds additional filters in case of group by timestamp, using the previously extarcted top 100 values
func buildWhereConditionForGBTForAdwords(groupByCombinations []map[string]interface{}) (string, []interface{}) {
	whereConditionForGBT := ""
	params := make([]interface{}, 0)
	filterStringAdwords := "adwords_documents.value"
	filterStringSmartPropertiesCampaign := "campaign.properties"
	filterStringSmartPropertiesAdGroup := "ad_group.properties"
	for _, groupByCombination := range groupByCombinations {
		whereConditionForEachCombination := ""
		for dimension, value := range groupByCombination {
			filterKey := ""
			filterString := ""
			if strings.HasPrefix(dimension, model.CampaignPrefix) {
				filterString, filterKey = model.GetFilterStringAndFilterKeyForAdwordsGBT(dimension, model.CampaignPrefix, "campaign", filterStringSmartPropertiesCampaign)
			} else if strings.HasPrefix(dimension, model.AdgroupPrefix) {
				filterString, filterKey = model.GetFilterStringAndFilterKeyForAdwordsGBT(dimension, model.AdgroupPrefix, "ad_group", filterStringSmartPropertiesAdGroup)
			} else {
				key := fmt.Sprintf(`%s:%s`, "keyword", strings.TrimPrefix(dimension, model.KeywordPrefix))
				filterString = filterStringAdwords
				currentFilterKey := model.AdwordsInternalPropertiesToReportsInternal[key]
				filterKey = currentFilterKey
			}
			if whereConditionForEachCombination == "" {
				if value != nil {
					whereConditionForEachCombination = fmt.Sprintf("%s->>'%s' = ? ", filterString, filterKey)
					params = append(params, value)
				} else {
					whereConditionForEachCombination = fmt.Sprintf("%s->>'%s' is null ", filterString, filterKey)
				}
			} else {
				if value != nil {
					whereConditionForEachCombination += fmt.Sprintf(" AND %s->>'%s' = ? ", filterString, filterKey)
					params = append(params, value)
				} else {
					whereConditionForEachCombination += fmt.Sprintf(" AND %s->>'%s' is null ", filterString, filterKey)
				}
			}
		}
		if whereConditionForGBT == "" {
			if whereConditionForEachCombination != "" {
				whereConditionForGBT = "(" + whereConditionForEachCombination + ")"
			}
		} else {
			if whereConditionForEachCombination != "" {
				whereConditionForGBT += (" OR (" + whereConditionForEachCombination + ")")
			}
		}
	}

	return whereConditionForGBT, params
}
func getFilterPropertiesForAdwordsReportsAndSmartProperty(filters []model.ChannelFilterV1) (string, []interface{}) {
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
			filterValue = fmt.Sprintf("%%%s%%", filter.Value)
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
				currentFilterStatement = fmt.Sprintf("%s.value->>'%s' %s ?", adwordsDocuments, currentFilterProperty, filterOperator)
			}
			params = append(params, filterValue)
			if index == 0 {
				resultStatement = fmt.Sprintf("(%s", currentFilterStatement)
			} else {
				resultStatement = fmt.Sprintf("%s %s %s", resultStatement, filter.LogicalOp, currentFilterStatement)
			}
		} else {
			currentFilterStatement = fmt.Sprintf("%s.properties->>'%s' %s '%s'", filter.Object, filter.Property, filterOperator, filterValue)
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
func getNotNullFilterStatementForSmartPropertyGroupBys(smartPropertyCampaignGroupBys []model.ChannelGroupBy, smartPropertyAdGroupGroupBys []model.ChannelGroupBy) string {
	resultStatement := ""
	for _, smartPropertyGroupBy := range smartPropertyCampaignGroupBys {
		if resultStatement == "" {
			resultStatement += fmt.Sprintf("( campaign.properties->>'%s' IS NOT NULL ", smartPropertyGroupBy.Property)
		} else {
			resultStatement += fmt.Sprintf("AND campaign.properties->>'%s' IS NOT NULL ", smartPropertyGroupBy.Property)
		}
	}
	for _, smartPropertyGroupBy := range smartPropertyAdGroupGroupBys {
		if resultStatement == "" {
			resultStatement += fmt.Sprintf("( ad_group.properties->>'%s' IS NOT NULL ", smartPropertyGroupBy.Property)
		} else {
			resultStatement += fmt.Sprintf("AND ad_group.properties->>'%s' IS NOT NULL ", smartPropertyGroupBy.Property)
		}
	}
	if resultStatement == "" {
		return resultStatement
	}
	return resultStatement + ")"
}

// @TODO Kark v0
func (pg *Postgres) GetAdwordsChannelResultMeta(projectID uint64, customerAccountID string,
	query *model.ChannelQuery) (*model.ChannelQueryResultMeta, error) {

	customerAccountIDArray := strings.Split(customerAccountID, ",")
	stmnt := "SELECT value->>'currency_code' as currency FROM adwords_documents" +
		" " + "WHERE project_id=? AND customer_account_id IN (?) AND type=? AND timestamp BETWEEN ? AND ?" +
		" " + "ORDER BY timestamp DESC LIMIT 1"

	logCtx := log.WithField("project_id", projectID)

	rows, tx, err := pg.ExecQueryWithContext(stmnt, []interface{}{projectID, customerAccountIDArray,
		model.AdwordsDocumentTypeAlias["customer_account_properties"],
		GetAdwordsDateOnlyTimestamp(query.From),
		GetAdwordsDateOnlyTimestamp(query.To)})
	if err != nil {
		logCtx.WithError(err).Error("Failed to build meta for channel query result.")
		return nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	var currency string
	for rows.Next() {
		rows.Scan(&currency)
	}

	err = rows.Err()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build meta for channel query result.")
		return nil, err
	}

	return &model.ChannelQueryResultMeta{Currency: currency}, nil
}

// ExecuteAdwordsChannelQuery - @TODO Kark v0
func (pg *Postgres) ExecuteAdwordsChannelQuery(projectID uint64, query *model.ChannelQuery) (*model.ChannelQueryResult, int) {

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	logCtx := log.WithField("project_id", projectID).WithField("query", query)
	if projectID == 0 || query == nil {
		logCtx.Error("Invalid project_id or query on execute adwords channel query.")
		return nil, http.StatusInternalServerError
	}

	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	if projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "" {
		logCtx.Error("Execute adwords channel query failed. No customer account id.")
		return nil, http.StatusInternalServerError
	}

	queryResult := &model.ChannelQueryResult{}
	meta, err := pg.GetAdwordsChannelResultMeta(projectID,
		*projectSetting.IntAdwordsCustomerAccountId, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get adwords channel result meta.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult.Meta = meta

	metricKvs, err := pg.getAdwordsMetrics(projectID, *projectSetting.IntAdwordsCustomerAccountId, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get adwords metric kvs.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult.Metrics = metricKvs

	// Return, if no breakdown.
	if query.Breakdown == "" {
		return queryResult, http.StatusOK
	}

	metricBreakdown, err := pg.getAdwordsMetricsBreakdown(projectID,
		*projectSetting.IntAdwordsCustomerAccountId, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get adwords metric breakdown.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult.MetricsBreakdown = metricBreakdown

	// sort only if the impression is there as column
	impressionsIndex := 0
	for _, key := range queryResult.MetricsBreakdown.Headers {
		if key == model.Impressions {
			// sort the rows by impressions count in descending order
			sort.Slice(queryResult.MetricsBreakdown.Rows, func(i, j int) bool {
				return queryResult.MetricsBreakdown.Rows[i][impressionsIndex].(float64) > queryResult.MetricsBreakdown.Rows[j][impressionsIndex].(float64)
			})
			break
		}
		impressionsIndex++
	}
	return queryResult, http.StatusOK
}

// GetAdwordsFilterPropertyKeyByType - @TODO Kark v0
func GetAdwordsFilterPropertyKeyByType(docType int) (string, error) {
	filterKeyByType := map[int]string{
		5:  "campaign_name",
		10: "ad_group_name",
		8:  "criteria",
		6:  "id",
	}

	filterKey, filterKeyExists := filterKeyByType[docType]
	if !filterKeyExists {
		return "", errors.New("no filter key found for document type")
	}

	return filterKey, nil
}

// GetAdwordsFilterValuesByType - @TODO Kark v0
func (pg *Postgres) GetAdwordsFilterValuesByType(projectID uint64, docType int) ([]string, int) {
	logCtx := log.WithField("projectID", projectID)
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return []string{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId
	if customerAccountID == nil || len(*customerAccountID) == 0 {
		logCtx.Error(integrationNotAvailable)
		return nil, http.StatusNotFound
	}

	db := C.GetServices().Db
	logCtx = log.WithField("project_id", projectID).WithField("doc_type", docType)

	filterValueKey, err := GetAdwordsFilterPropertyKeyByType(docType)
	if err != nil {
		logCtx.WithError(err).Error("Unknown doc type for get adwords filter key.")
		return []string{}, http.StatusBadRequest
	}

	queryStr := "SELECT DISTINCT(LOWER(value->>?)) as filter_value FROM adwords_documents WHERE project_id = ? AND" +
		" " + "customer_account_id = ? AND type = ? LIMIT 5000"
	rows, err := db.Raw(queryStr, filterValueKey, projectID, customerAccountID, docType).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to distinct filter values by type from adwords documents.")
		return []string{}, http.StatusInternalServerError
	}
	defer rows.Close()

	filterValues := make([]string, 0, 0)
	for rows.Next() {
		var filterValue string
		if err := rows.Scan(&filterValue); err != nil {
			logCtx.WithError(err).Error("Failed to distinct filter values by type from adwords documents.")
			return filterValues, http.StatusInternalServerError
		}

		filterValues = append(filterValues, filterValue)
	}

	return filterValues, http.StatusFound
}

// GetAdwordsDocumentTypeForFilterKey - @TODO Kark v0
func GetAdwordsDocumentTypeForFilterKey(filter string) (int, error) {
	var docType int

	switch filter {
	case CAFilterCampaign:
		docType = model.AdwordsDocumentTypeAlias["campaign_performance_report"]
	case CAFilterAd:
		docType = model.AdwordsDocumentTypeAlias["ad_performance_report"]
	case CAFilterKeyword:
		docType = model.AdwordsDocumentTypeAlias["keyword_performance_report"]
	case CAFilterAdGroup:
		docType = model.AdwordsDocumentTypeAlias["ad_group_performance_report"]
	}

	if docType == 0 {
		return docType, errors.New("no adwords document type for filter")
	}

	return docType, nil
}

/*
GetAdwordsMetricsQuery
SELECT value->>'criteria', SUM((value->>'impressions')::float) as impressions, SUM((value->>'clicks')::float) as clicks,
SUM((value->>'cost')::float) as total_cost, SUM((value->>'conversions')::float) as all_conversions,
SUM((value->>'all_conversions')::float) as all_conversions FROM adwords_documents
WHERE type='5' AND timestamp BETWEEN '20191122' and '20191129' AND value->>'campaign_name'='Desktop Only'
GROUP BY value->>'criteria';
*/
func (pg *Postgres) getAdwordsMetricsQuery(projectID uint64, customerAccountID string, query *model.ChannelQuery,
	withBreakdown bool) (string, []interface{}, error) {

	customerAccountIDArray := strings.Split(customerAccountID, ",")
	// select handling.
	selectColstWithoutAlias := "SUM((value->>'impressions')::float) as %s, SUM((value->>'clicks')::float) as %s," +
		" " + "SUM((value->>'cost')::float)/1000000 as %s, SUM((value->>'conversions')::float) as %s," +
		" " + "SUM((value->>'all_conversions')::float) as %s," +
		" " + "SUM((value->>'cost')::float)/NULLIF(SUM((value->>'clicks')::float), 0)/1000000 as %s," +
		" " + "SUM((value->>'clicks')::float * regexp_replace(value->>'conversion_rate', ?, '')::float)/NULLIF(SUM((value->>'clicks')::float), 0) as %s," +
		" " + "SUM((value->>'cost')::float)/NULLIF(SUM((value->>'conversions')::float), 0)/1000000 as %s"
	selectCols := fmt.Sprintf(selectColstWithoutAlias, CAColumnImpressions, CAColumnClicks,
		CAColumnTotalCost, CAColumnConversions, CAColumnAllConversions,
		CAColumnCostPerClick, CAColumnConversionRate, CAColumnCostPerConversion)

	paramsSelect := make([]interface{}, 0, 0)

	// Where handling.
	stmntWhere := "WHERE project_id=? AND customer_account_id IN (?) AND type=? AND timestamp BETWEEN ? AND ?"
	paramsWhere := make([]interface{}, 0, 0)

	docType, err := GetAdwordsDocumentTypeForFilterKey(query.FilterKey)
	if err != nil {
		return "", []interface{}{}, err
	}

	paramsWhere = append(paramsWhere, projectID, customerAccountIDArray, docType,
		GetAdwordsDateOnlyTimestamp(query.From), GetAdwordsDateOnlyTimestamp(query.To))

	isWhereByFilterRequired := query.FilterValue != filterValueAll
	if isWhereByFilterRequired {
		stmntWhere = stmntWhere + " " + "AND" + " " + "value->>?=?"

		filterKey, err := GetAdwordsFilterPropertyKeyByType(docType)
		if err != nil {
			return "", []interface{}{}, err
		}

		paramsWhere = append(paramsWhere, filterKey, query.FilterValue)
	}

	// group by handling.
	var stmntGroupBy string
	paramsGroupBy := make([]interface{}, 0, 0)
	if withBreakdown {
		// Todo: Use a seperate or a generic method for getting property key to group by
		// for a specific key type. Now using method which does same for filterKey
		// for breakdownKey. Say campaigns, group by campaign_name.
		docType, err := GetAdwordsDocumentTypeForFilterKey(query.Breakdown)
		if err != nil {
			log.WithError(err).Error("Failed to get adwords doc type by filter key.")
			return "", []interface{}{}, err
		}
		propertyKey, err := GetAdwordsFilterPropertyKeyByType(docType)
		if err != nil {
			log.WithError(err).Error("Failed to get filter propery key by type.")
			return "", []interface{}{}, err
		}

		// prepend group by col on select.
		selectCols = "value->>? as %s" + ", " + selectCols
		selectCols = fmt.Sprintf(selectCols, CAChannelGroupKey)
		paramsSelect = append(paramsSelect, propertyKey)

		stmntGroupBy = "GROUP BY" + " " + "%s"
		stmntGroupBy = fmt.Sprintf(stmntGroupBy, CAChannelGroupKey)
	}

	// Using prepared statement to replace '%', to avoid
	// query breakage with "!%(MISSING)" on gorm.
	paramsSelect = append(paramsSelect, "%")

	params := make([]interface{}, 0, 0)

	stmnt := "SELECT" + " " + selectCols + " " + "FROM adwords_documents" + " " + stmntWhere + " " + stmntGroupBy
	params = append(params, paramsSelect...)
	params = append(params, paramsWhere...)
	params = append(params, paramsGroupBy...)

	return stmnt, params, nil
}

// @TODO Kark v0
func (pg *Postgres) getAdwordsMetrics(projectID uint64, customerAccountID string,
	query *model.ChannelQuery) (*map[string]interface{}, error) {

	stmnt, params, err := pg.getAdwordsMetricsQuery(projectID, customerAccountID, query, false)
	if err != nil {
		return nil, err
	}

	rows, tx, err := pg.ExecQueryWithContext(stmnt, params)
	if err != nil {
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows, tx)

	if err != nil {
		return nil, err
	}
	if len(resultRows) == 0 {
		log.Warn("Aggregate query returned zero rows.")
		return nil, errors.New("no rows returned")
	}

	if len(resultRows) > 1 {
		log.Warn("Aggregate query returned more than one row on get adwords metric kvs.")
	}

	metricKvs := make(map[string]interface{})
	for i, k := range resultHeaders {
		metricKvs[k] = resultRows[0][i]
	}

	return &metricKvs, nil
}

// @TODO Kark v0
func (pg *Postgres) getAdwordsMetricsBreakdown(projectID uint64, customerAccountID string,
	query *model.ChannelQuery) (*model.ChannelBreakdownResult, error) {

	logCtx := log.WithField("project_id", projectID).WithField("customer_account_id", customerAccountID)

	stmnt, params, err := pg.getAdwordsMetricsQuery(projectID, customerAccountID, query, true)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get adwords metrics query.")
		return nil, err
	}

	rows, tx, err := pg.ExecQueryWithContext(stmnt, params)
	if err != nil {
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows, tx)
	if err != nil {
		return nil, err
	}

	// Translate group key.
	for i := range resultHeaders {
		if resultHeaders[i] == CAChannelGroupKey {
			resultHeaders[i] = query.Breakdown
		}
	}

	// Fill null with zero for aggr.
	// Do I need to show this as NA?
	for ri := range resultRows {
		for ci := range resultRows[ri] {
			// if not group key and nil: zero.
			if ci > 0 && resultRows[ri][ci] == nil {
				resultRows[ri][ci] = float64(0)
			}
		}
	}

	return &model.ChannelBreakdownResult{Headers: resultHeaders, Rows: resultRows}, nil
}

func (pg *Postgres) GetLatestMetaForAdwordsForGivenDays(projectID uint64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields) {
	db := C.GetServices().Db

	channelDocumentsCampaign := make([]model.ChannelDocumentsWithFields, 0)
	channelDocumentsAdGroup := make([]model.ChannelDocumentsWithFields, 0)

	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		log.Error("Failed to get project settings")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	if projectSetting.IntAdwordsCustomerAccountId == nil || *(projectSetting.IntAdwordsCustomerAccountId) == "" {
		log.Error("Failed to get custtomer account ids")
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

	err = db.Raw(adwordsAdGroupMetdataFetchQueryStr, model.AdwordsDocumentTypeAlias["ad_groups"], projectID, from, to, customerAccountIDs,
		model.AdwordsDocumentTypeAlias["ad_groups"], projectID, from, to, customerAccountIDs).Find(&channelDocumentsAdGroup).Error
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d ad_group meta for adwords", days)
		log.Error(errString)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	err = db.Raw(adwordsCampaignMetadataFetchQueryStr, model.AdwordsDocumentTypeAlias["campaigns"], projectID, from, to, customerAccountIDs, model.AdwordsDocumentTypeAlias["campaigns"], projectID, from, to, customerAccountIDs).Find(&channelDocumentsCampaign).Error
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d campaign meta for adwords", days)
		log.Error(errString)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	return channelDocumentsCampaign, channelDocumentsAdGroup
}

func (pg *Postgres) ExecuteAdwordsSEMChecklistQuery(projectID uint64, query model.TemplateQuery, reqID string) (model.TemplateResponse, int) {
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	customerAccountID, err := pg.validateIntegratonAndMetricsForAdwordsSEMChecklist(projectID, query, reqID)
	if err != nil && err.Error() == integrationNotAvailable {
		logCtx.WithError(err).Info(model.AdwordsSpecificError)
		return model.TemplateResponse{}, http.StatusNotFound
	}
	customerAccountIDs := strings.Split(*customerAccountID, ",")
	templateQueryResponse, err := pg.getAdwordsSEMChecklistQueryData(query, projectID, customerAccountIDs, reqID)
	if err != nil {
		logCtx.Error("Failed to get template query response. Error: ", err.Error())
		return model.TemplateResponse{}, http.StatusNotFound
	}
	return templateQueryResponse, http.StatusOK
}
func (pg *Postgres) validateIntegratonAndMetricsForAdwordsSEMChecklist(projectID uint64, query model.TemplateQuery, reqID string) (*string, error) {
	logCtx := log.WithField("req_id", reqID)
	var err error
	projectSetting, errCode := pg.GetProjectSetting(projectID)
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
	selectStmnt := ""
	for _, value := range coalesceSelectsForKeywords {
		selectStmnt += (fmt.Sprintf(coalesceForChecklists, "keyword_analysis_last_week", value, "keyword_analysis_previous_week", value, value) + ", ")
	}
	selectStmnt += fmt.Sprintf("%s as %s, ", fmt.Sprintf(percentageChangeForSemChecklist, "keyword_analysis_last_week", "analysis_metric", "keyword_analysis_previous_week", "analysis_metric", "keyword_analysis_previous_week", "analysis_metric"), "percentage_change")
	selectStmnt += fixedSelectForBreakdownAnalysisKeyword
	return selectStmnt
}
func (pg *Postgres) getKeywordLevelDataForTemplates(projectID uint64, customerAccountID []string, query model.TemplateQuery) ([]model.KeywordAnalysis, error) {
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
	rows, tx, err := pg.ExecQueryWithContext(finalKeywordQuery, params)
	if err != nil {
		return make([]model.KeywordAnalysis, 0), err
	}
	defer U.CloseReadQuery(rows, tx)
	for rows.Next() {
		var keywordAnalysisRow model.KeywordAnalysis
		err := db.ScanRows(rows, &keywordAnalysisRow)
		if err != nil {
			return make([]model.KeywordAnalysis, 0), err
		}
		keywordAnalysisResult = append(keywordAnalysisResult, keywordAnalysisRow)
	}
	cleanKeywordAnalysisResult := model.SanitiseKeywordsAnalysisResult(query, keywordAnalysisResult)
	sort.SliceStable(cleanKeywordAnalysisResult, func(i, j int) bool {
		return cleanKeywordAnalysisResult[i].AbsoluteChange > cleanKeywordAnalysisResult[j].AbsoluteChange
	})
	return cleanKeywordAnalysisResult, nil
}
func (pg *Postgres) getCampaignLevelDataForTemplates(projectID uint64, customerAccountID []string, campaignArray []string, query model.TemplateQuery) ([]model.CampaignAnalysis, error) {
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
	rows, tx, err := pg.ExecQueryWithContext(finalCampaignQuery, params)
	if err != nil {
		return make([]model.CampaignAnalysis, 0), err
	}
	defer U.CloseReadQuery(rows, tx)
	for rows.Next() {
		var campaignAnalysisRow model.CampaignAnalysis
		err := db.ScanRows(rows, &campaignAnalysisRow)
		if err != nil {
			return make([]model.CampaignAnalysis, 0), err
		}
		campaignAnalysisResult = append(campaignAnalysisResult, campaignAnalysisRow)
	}
	cleanCampaignAnalysisResult := model.SanitiseCampaignAnalysisResult(query, campaignAnalysisResult)
	sort.SliceStable(cleanCampaignAnalysisResult, func(i, j int) bool {
		return cleanCampaignAnalysisResult[i].AbsoluteChange > cleanCampaignAnalysisResult[j].AbsoluteChange
	})
	return cleanCampaignAnalysisResult, nil
}
func (pg *Postgres) getOverallChangesDataForTemplates(projectID uint64, customerAccountID []string, query model.TemplateQuery) ([]model.OverallChanges, error) {
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

	rows, tx, err := pg.ExecQueryWithContext(overallAnalysisQuery, paramsLastWeek)
	if err != nil {
		return make([]model.OverallChanges, 0), err
	}

	resultHeadersLastWeek, resultRowsLastWeek, err := U.DBReadRows(rows, tx)
	if err != nil {
		return make([]model.OverallChanges, 0), err
	}

	rows, tx, err = pg.ExecQueryWithContext(overallAnalysisQuery, paramsPrevWeek)
	if err != nil {
		return make([]model.OverallChanges, 0), err
	}

	_, resultRowsPreviousWeek, err := U.DBReadRows(rows, tx)

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
func (pg *Postgres) getAdwordsSEMChecklistQueryData(query model.TemplateQuery, projectID uint64, customerAccountID []string,
	reqID string) (model.TemplateResponse, error) {
	var result model.TemplateResponse
	keywordAnalysisResult, err := pg.getKeywordLevelDataForTemplates(projectID, customerAccountID, query)
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

	campaignAnalysisResult, err := pg.getCampaignLevelDataForTemplates(projectID, customerAccountID, campaignArray, query)
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
	overallChangesResult, err := pg.getOverallChangesDataForTemplates(projectID, customerAccountID, query)
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
