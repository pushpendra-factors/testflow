package postgres

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	campaignPerformanceReport       = "campaign_performance_report"
	adGroupPerformanceReport        = "ad_group_performance_report"
	adPerformanceReport             = "ad_performance_report"
	keywordPerformanceReport        = "keyword_performance_report"
	searchPerformanceReport         = "search_performance_report"
	adwordsCampaign                 = "campaign"
	adwordsAdGroup                  = "ad_group"
	adwordsAd                       = "ad"
	adwordsKeyword                  = "keyword"
	adwordsStringColumn             = "adwords"
	errorDuplicateAdwordsDocument   = "pq: duplicate key value violates unique constraint \"adwords_documents_pkey\""
	filterValueAll                  = "all"
	lastSyncInfoQueryForAllProjects = "SELECT project_id, customer_account_id, type as document_type, max(timestamp) as last_timestamp" +
		" " + "FROM adwords_documents GROUP BY project_id, customer_account_id, type"
	lastSyncInfoForAProject = "SELECT project_id, customer_account_id, type as document_type, max(timestamp) as last_timestamp" +
		" " + "FROM adwords_documents WHERE project_id = ? GROUP BY project_id, customer_account_id, type"
	insertAdwordsStr               = "INSERT INTO adwords_documents (project_id,customer_account_id,type,timestamp,id,campaign_id,ad_group_id,ad_id,keyword_id,value,created_at,updated_at) VALUES "
	adwordsFilterQueryStr          = "SELECT DISTINCT(value->>?) as filter_value FROM adwords_documents WHERE project_id = ? AND" + " " + "customer_account_id = ? AND type = ? AND value->>? IS NOT NULL LIMIT 5000"
	staticWhereStatementForAdwords = "WHERE project_id = ? AND customer_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "
	fromAdwordsDocument            = " FROM adwords_documents "

	impressions                = "impressions"
	shareHigherOrderExpression = "sum(case when value->>'%s' IS NOT NULL THEN (value->>'%s')::float else 0 END)/NULLIF(sum(case when value->>'%s' IS NOT NULL THEN (value->>'%s')::float else 0 END), 0)"
	sumOfFloatExp              = "sum((value->>'%s')::float)"
	approvalStatus             = "approval_status"
	matchType                  = "match_type"
	firstPositionCpc           = "first_position_cpc"
	firstPageCpc               = "first_page_cpc"
	isNegative                 = "is_negative"
	topOfPageCpc               = "top_of_page_cpc"
	qualityScore               = "quality_score"

	clicks                                     = "clicks"
	clickThroughRate                           = "click_through_rate"
	conversion                                 = "conversion"
	conversionRate                             = "conversion_rate"
	costPerClick                               = "cost_per_click"
	costPerConversion                          = "cost_per_conversion"
	searchImpressionShare                      = "search_impression_share"
	searchClickShare                           = "search_click_share"
	searchTopImpressionShare                   = "search_top_impression_share"
	searchAbsoluteTopImpressionShare           = "search_absolute_top_impression_share"
	searchBudgetLostAbsoluteTopImpressionShare = "search_budget_lost_absolute_top_impression_share"
	searchBudgetLostImpressionShare            = "search_budget_lost_impression_share"
	searchBudgetLostTopImpressionShare         = "search_budget_lost_top_impression_share"
	searchRankLostAbsoluteTopImpressionShare   = "search_rank_lost_absolute_top_impression_share"
	searchRankLostImpressionShare              = "search_rank_lost_impression_share"
	searchRankLostTopImpressionShare           = "search_rank_lost_top_impression_share"
	totalSearchImpression                      = "total_search_impression"
	totalSearchClick                           = "total_search_click"
	totalSearchTopImpression                   = "total_search_top_impression"
	totalSearchAbsoluteTopImpression           = "total_search_absolute_top_impression"
	totalSearchBudgetLostAbsoluteTopImpression = "total_search_budget_lost_absolute_top_impression"
	totalSearchBudgetLostImpression            = "total_search_budget_lost_impression"
	totalSearchBudgetLostTopImpression         = "total_search_budget_lost_top_impression"
	totalSearchRankLostAbsoluteTopImpression   = "total_search_rank_lost_absolute_top_impression"
	totalSearchRankLostImpression              = "total_search_rank_lost_impression"
	totalSearchRankLostTopImpression           = "total_search_rank_lost_top_impression"
)

var selectableMetricsForAdwords = []string{
	conversion,
	clickThroughRate,
	conversionRate,
	costPerClick,
	costPerConversion,
	searchImpressionShare,
	searchClickShare,
	searchTopImpressionShare,
	searchAbsoluteTopImpressionShare,
	searchBudgetLostAbsoluteTopImpressionShare,
	searchBudgetLostImpressionShare,
	searchBudgetLostTopImpressionShare,
	searchRankLostAbsoluteTopImpressionShare,
	searchRankLostImpressionShare,
	searchRankLostTopImpressionShare,
}

var errorEmptyAdwordsDocument = errors.New("empty adwords document")

var objectsForAdwords = []string{adwordsCampaign, adwordsAdGroup, adwordsKeyword}

var mapOfObjectsToPropertiesAndRelated = map[string]map[string]PropertiesAndRelated{
	adwordsCampaign: {
		"id":     PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		"name":   PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		"status": PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
	},
	adwordsAdGroup: {
		"id":     PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		"name":   PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		"status": PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
	},
	adwordsKeyword: {
		"id":             PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		"name":           PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		"status":         PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		approvalStatus:   PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		matchType:        PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		firstPositionCpc: PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		firstPageCpc:     PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		isNegative:       PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		topOfPageCpc:     PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		qualityScore:     PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
	},
}

// AdwordsDocumentTypeAlias ...
var AdwordsDocumentTypeAlias = map[string]int{
	"campaigns":                   1,
	"ads":                         2,
	"ad_groups":                   3,
	"click_performance_report":    4,
	campaignPerformanceReport:     5,
	adPerformanceReport:           6,
	searchPerformanceReport:       7,
	keywordPerformanceReport:      8,
	"customer_account_properties": 9,
	adGroupPerformanceReport:      10,
}

/*
	Map from request Params to Internal Representation is needed, so that validation of params and operating within adwords context becomes easy.
	Map from Internal Representation to Representation within Report/Job as field values can vary.
	Map from Internal Representation to External Representation is needed to expose right column names and also to perform clear operations like union or join.
		We can follow the same representation of external even during cte formation, though used in internal context.
	We might all above complicated transformations in api if we merge all document types i.e.facebook, linkedin etc...
*/
var adwordsExtToInternal = map[string]string{
	"campaign":                       "campaign",
	"ad_group":                       "ad_group",
	"ad":                             "ad",
	"name":                           "name",
	"keyword":                        "keyword",
	"id":                             "id",
	"status":                         "status",
	approvalStatus:                   approvalStatus,
	matchType:                        matchType,
	firstPositionCpc:                 firstPositionCpc,
	firstPageCpc:                     firstPageCpc,
	isNegative:                       isNegative,
	topOfPageCpc:                     topOfPageCpc,
	qualityScore:                     qualityScore,
	impressions:                      impressions,
	clicks:                           clicks,
	"spend":                          "cost",
	conversion:                       "conversions",
	clickThroughRate:                 clickThroughRate,
	conversionRate:                   conversionRate,
	costPerClick:                     costPerClick,
	costPerConversion:                costPerConversion,
	searchImpressionShare:            searchImpressionShare,
	searchClickShare:                 searchClickShare,
	searchTopImpressionShare:         searchTopImpressionShare,
	searchAbsoluteTopImpressionShare: searchAbsoluteTopImpressionShare,
	searchBudgetLostAbsoluteTopImpressionShare: searchBudgetLostAbsoluteTopImpressionShare,
	searchBudgetLostImpressionShare:            searchBudgetLostImpressionShare,
	searchBudgetLostTopImpressionShare:         searchBudgetLostTopImpressionShare,
	searchRankLostAbsoluteTopImpressionShare:   searchRankLostAbsoluteTopImpressionShare,
	searchRankLostImpressionShare:              searchRankLostImpressionShare,
	searchRankLostTopImpressionShare:           searchRankLostTopImpressionShare,
}

var adwordsInternalPropertiesToJobsInternal = map[string]string{
	"campaign:id":                "id",
	"campaign:name":              "name",
	"campaign:status":            "status",
	"ad_group:id":                "id",
	"ad_group:name":              "name",
	"ad_group:status":            "status",
	"ad:id":                      "ad_id",
	"keyword:id":                 "id",
	"keyword:name":               "criteria",
	"keyword:status":             "status",
	"keyword:approval_status":    approvalStatus,
	"keyword:match_type":         "keyword_match_type",
	"keyword:first_position_cpc": firstPositionCpc,
	"keyword:first_page_cpc":     firstPageCpc,
	"keyword:is_negative":        isNegative,
	"keyword:top_of_page_cpc":    topOfPageCpc,
	"keyword:quality_score":      qualityScore,
}

var adwordsInternalPropertiesToReportsInternal = map[string]string{
	"campaign:id":                "campaign_id",
	"campaign:name":              "campaign_name",
	"campaign:status":            "campaign_status",
	"ad_group:id":                "ad_group_id",
	"ad_group:name":              "ad_group_name",
	"ad_group:status":            "ad_group_status",
	"ad:id":                      "ad_id",
	"keyword:id":                 "keyword_id",
	"keyword:name":               "criteria",
	"keyword:status":             "status",
	"keyword:approval_status":    approvalStatus,
	"keyword:match_type":         "keyword_match_type",
	"keyword:first_position_cpc": firstPositionCpc,
	"keyword:first_page_cpc":     firstPageCpc,
	"keyword:is_negative":        isNegative,
	"keyword:top_of_page_cpc":    topOfPageCpc,
	"keyword:quality_score":      qualityScore,
}

var propertiesToBeDividedByMillion = map[string]struct{}{
	topOfPageCpc:     {},
	firstPositionCpc: {},
	firstPageCpc:     {},
}

type metricsAndRelated struct {
	higherOrderExpression    string
	nonHigherOrderExpression string
	externalValue            string
	externalOperation        string // This is not clear. What happens when ctr at higher level is presented?
}

var nonHigherOrderMetrics = map[string]struct{}{
	impressions:   {},
	clicks:        {},
	"cost":        {},
	"conversions": {},
}

// Same structure is being used for internal operations and external.
var adwordsInternalMetricsToAllRep = map[string]metricsAndRelated{
	impressions: {
		nonHigherOrderExpression: "sum((value->>'impressions')::float)",
		externalValue:            impressions,
		externalOperation:        "sum",
	},
	clicks: {
		nonHigherOrderExpression: "sum((value->>'clicks')::float)",
		externalValue:            clicks,
		externalOperation:        "sum",
	},
	"cost": {
		nonHigherOrderExpression: "sum((value->>'cost')::float)/1000000",
		externalValue:            "spend",
		externalOperation:        "sum",
	},
	"conversions": {
		nonHigherOrderExpression: "sum((value->>'conversions')::float)",
		externalValue:            conversion,
		externalOperation:        "sum",
	},
	clickThroughRate: {
		higherOrderExpression:    "sum((value->>'clicks')::float)*100/NULLIF(sum((value->>'impressions')::float), 0)",
		nonHigherOrderExpression: "sum((value->>'clicks')::float)*100",
		externalValue:            clickThroughRate,
		externalOperation:        "sum",
	},
	conversionRate: {
		higherOrderExpression:    "sum((value->>'conversions')::float)*100/NULLIF(sum((value->>'clicks')::float), 0)",
		nonHigherOrderExpression: "sum((value->>'conversions')::float)*100",
		externalValue:            conversionRate,
		externalOperation:        "sum",
	},
	costPerClick: {
		higherOrderExpression:    "(sum((value->>'cost')::float)/1000000)/NULLIF(sum((value->>'clicks')::float), 0)",
		nonHigherOrderExpression: "(sum((value->>'cost')::float)/1000000)",
		externalValue:            costPerClick,
		externalOperation:        "sum",
	},
	costPerConversion: {
		higherOrderExpression:    "(sum((value->>'cost')::float)/1000000)/NULLIF(sum((value->>'conversions')::float), 0)",
		nonHigherOrderExpression: "(sum((value->>'cost')::float)/1000000)",
		externalValue:            costPerConversion,
		externalOperation:        "sum",
	},
	searchImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, searchImpressionShare, impressions, searchImpressionShare, totalSearchImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, totalSearchImpression),
		externalValue:            searchImpressionShare,
		externalOperation:        "sum",
	},
	searchClickShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, searchClickShare, impressions, searchClickShare, totalSearchClick),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, totalSearchClick),
		externalValue:            searchClickShare,
		externalOperation:        "sum",
	},
	searchTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, searchTopImpressionShare, impressions, searchTopImpressionShare, totalSearchTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, totalSearchTopImpression),
		externalValue:            searchTopImpressionShare,
		externalOperation:        "sum",
	},
	searchAbsoluteTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, searchAbsoluteTopImpressionShare, impressions, searchAbsoluteTopImpressionShare, totalSearchAbsoluteTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, totalSearchAbsoluteTopImpression),
		externalValue:            searchAbsoluteTopImpressionShare,
		externalOperation:        "sum",
	},
	searchBudgetLostAbsoluteTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, searchBudgetLostAbsoluteTopImpressionShare, impressions, searchBudgetLostAbsoluteTopImpressionShare, totalSearchBudgetLostAbsoluteTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, totalSearchBudgetLostAbsoluteTopImpression),
		externalValue:            searchBudgetLostAbsoluteTopImpressionShare,
		externalOperation:        "sum",
	},
	searchBudgetLostImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, searchBudgetLostImpressionShare, impressions, searchBudgetLostImpressionShare, totalSearchBudgetLostImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, totalSearchBudgetLostImpression),
		externalValue:            searchBudgetLostImpressionShare,
		externalOperation:        "sum",
	},
	searchBudgetLostTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, searchBudgetLostTopImpressionShare, impressions, searchBudgetLostTopImpressionShare, totalSearchBudgetLostTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, totalSearchBudgetLostTopImpression),
		externalValue:            searchBudgetLostTopImpressionShare,
		externalOperation:        "sum",
	},
	searchRankLostAbsoluteTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, searchRankLostAbsoluteTopImpressionShare, impressions, searchRankLostAbsoluteTopImpressionShare, totalSearchRankLostAbsoluteTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, totalSearchRankLostAbsoluteTopImpression),
		externalValue:            searchRankLostAbsoluteTopImpressionShare,
		externalOperation:        "sum",
	},
	searchRankLostImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, searchRankLostImpressionShare, impressions, searchRankLostImpressionShare, totalSearchRankLostImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, totalSearchRankLostImpression),
		externalValue:            searchRankLostImpressionShare,
		externalOperation:        "sum",
	},
	searchRankLostTopImpressionShare: {
		higherOrderExpression:    fmt.Sprintf(shareHigherOrderExpression, searchRankLostTopImpressionShare, impressions, searchRankLostTopImpressionShare, totalSearchRankLostTopImpression),
		nonHigherOrderExpression: fmt.Sprintf(sumOfFloatExp, totalSearchRankLostTopImpression),
		externalValue:            searchRankLostTopImpressionShare,
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
	if docType == AdwordsDocumentTypeAlias[keywordPerformanceReport] || docType == AdwordsDocumentTypeAlias[searchPerformanceReport] {
		idStr = U.GetUUID()
	}

	value1, value2, value3, value4 := getAdwordsHierarchyColumnsByType(valueMap, docType)

	// ID as string always.
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
			log.WithError(dbc.Error).Error("Failed to create an adwords doc. Duplicate.")
			return http.StatusConflict
		}
	}

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

	return sanitizedLastSyncInfos(adwordsLastSyncInfos, adwordsSettings)
}

// GetAllAdwordsLastSyncInfoByProjectCustomerAccountAndType - @TODO Kark v1
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

	return sanitizedLastSyncInfos(adwordsLastSyncInfos, adwordsSettings)
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
func sanitizedLastSyncInfos(adwordsLastSyncInfos []model.AdwordsLastSyncInfo, adwordsSettings []model.AdwordsProjectSettings) ([]model.AdwordsLastSyncInfo, int) {

	adwordsSettingsByProjectAndCustomerAccount := make(map[uint64]map[string]*model.AdwordsProjectSettings, 0)

	for i := range adwordsSettings {
		customerAccountIDs := strings.Split(adwordsSettings[i].CustomerAccountId, ",")
		adwordsSettingsByProjectAndCustomerAccount[adwordsSettings[i].ProjectId] = make(map[string]*model.AdwordsProjectSettings)
		for j := range customerAccountIDs {
			var setting model.AdwordsProjectSettings
			setting.ProjectId = adwordsSettings[i].ProjectId
			setting.AgentUUID = adwordsSettings[i].AgentUUID
			setting.RefreshToken = adwordsSettings[i].RefreshToken
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
			logCtx.Error("Adwords project settings not found for customer account adwords synced earlier.")
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
					}

					selectedLastSyncInfos = append(selectedLastSyncInfos, syncInfo)
				}
			}
		}

	}

	return selectedLastSyncInfos, http.StatusOK
}

// GetGCLIDBasedCampaignInfo - It returns GCLID based campaign info ( Adgroup, Campaign and Ad) for given time range and adwords account
func (pg *Postgres) GetGCLIDBasedCampaignInfo(projectID uint64, from, to int64, adwordsAccountIDs string) (map[string]model.CampaignInfo, error) {

	logCtx := log.WithFields(log.Fields{"ProjectID": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})
	adGroupNameCase := "CASE WHEN value->>'ad_group_name' IS NULL THEN ? " +
		" WHEN value->>'ad_group_name' = '' THEN ? ELSE value->>'ad_group_name' END AS ad_group_name"
	campaignNameCase := "CASE WHEN value->>'campaign_name' IS NULL THEN ? " +
		" WHEN value->>'campaign_name' = '' THEN ? ELSE value->>'campaign_name' END AS campaign_name"
	adIDCase := "CASE WHEN value->>'creative_id' IS NULL THEN ? " +
		" WHEN value->>'creative_id' = '' THEN ? ELSE value->>'creative_id' END AS creative_id"

	performanceQuery := "SELECT id, " + adGroupNameCase + ", " + campaignNameCase + ", " + adIDCase +
		" FROM adwords_documents where project_id = ? AND customer_account_id IN (?) AND type = ? AND timestamp between ? AND ? "
	customerAccountIDs := strings.Split(adwordsAccountIDs, ",")
	rows, err := pg.ExecQueryWithContext(performanceQuery, []interface{}{model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone,
		model.PropertyValueNone, model.PropertyValueNone, projectID, customerAccountIDs, AdwordsClickReportType, U.GetDateOnlyFromTimestamp(from),
		U.GetDateOnlyFromTimestamp(to)})
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, err
	}
	defer rows.Close()
	gclIDBasedCampaign := make(map[string]model.CampaignInfo)
	for rows.Next() {
		var gclID string
		var adgroupName string
		var campaignName string
		var adID string
		if err = rows.Scan(&gclID, &adgroupName, &campaignName, &adID); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		gclIDBasedCampaign[gclID] = model.CampaignInfo{
			AdgroupName:  adgroupName,
			CampaignName: campaignName,
			AdID:         adID,
		}
	}
	return gclIDBasedCampaign, nil
}

// @TODO Kark v1
func buildAdwordsChannelConfig() *model.ChannelConfigResult {
	adwordsObjectsAndProperties := buildObjectAndPropertiesForAdwords(objectsForAdwords)
	selectMetrics := append(selectableMetricsForAllChannels, selectableMetricsForAdwords...)
	objectsAndProperties := append(adwordsObjectsAndProperties)
	return &model.ChannelConfigResult{
		SelectMetrics:        selectMetrics,
		ObjectsAndProperties: objectsAndProperties,
	}
}

func buildObjectAndPropertiesForAdwords(objects []string) []model.ChannelObjectAndProperties {
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0, 0)
	for _, currentObject := range objects {
		propertiesAndRelated, isPresent := mapOfObjectsToPropertiesAndRelated[currentObject]
		var currentProperties []model.ChannelProperty
		if isPresent {
			currentProperties = buildProperties(propertiesAndRelated)
		} else {
			currentProperties = buildProperties(allChannelsPropertyToRelated)
		}
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
}

// GetAdwordsFilterValues - @TODO Kark v1
func (pg *Postgres) GetAdwordsFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
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
		logCtx.Error("adwords integration is not available.")
		return "", []interface{}{}, http.StatusInternalServerError
	}

	params := []interface{}{adwordsInternalFilterProperty, projectID, customerAccountID, docType, adwordsInternalFilterProperty}
	return "(" + adwordsFilterQueryStr + ")", params, http.StatusFound
}

func getFilterRelatedInformationForAdwords(requestFilterObject string, requestFilterProperty string) (string, int, int) {
	adwordsInternalFilterObject, isPresent := adwordsExtToInternal[requestFilterObject]
	if !isPresent {
		log.Error("Invalid adwords filter object.")
		return "", 0, http.StatusBadRequest
	}
	docType := getAdwordsDocumentTypeForFilterKeyV1(adwordsInternalFilterObject)

	adwordsInternalFilterProperty, isPresent := adwordsExtToInternal[requestFilterProperty]
	if !isPresent {
		log.Error("Invalid adwords filter property.")
		return "", 0, http.StatusBadRequest
	}
	keyForJobInternalRepresentation := fmt.Sprintf("%s:%s", adwordsInternalFilterObject, adwordsInternalFilterProperty)
	adwordsInternalPropertyOfJob, isPresent := adwordsInternalPropertiesToJobsInternal[keyForJobInternalRepresentation]
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
		logCtx.Error("adwords Integration is not available.")
		return []interface{}{}, http.StatusInternalServerError
	}

	logCtx = log.WithField("doc_type", docType)
	params := []interface{}{property, projectID, customerAccountID, docType, property}
	_, resultRows, _ := pg.ExecuteSQL(adwordsFilterQueryStr, params, logCtx)

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
func (pg *Postgres) ExecuteAdwordsChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, error) {
	fetchSource := false
	logCtx := log.WithField("xreq_id", reqID)
	sql, params, selectKeys, selectMetrics, err := pg.GetSQLQueryAndParametersForAdwordsQueryV1(
		projectID, query, reqID, fetchSource)
	if err != nil {
		return make([]string, 0, 0), make([][]interface{}, 0, 0), err
	}
	_, resultMetrics, err := pg.ExecuteSQL(sql, params, logCtx)
	columns := append(selectKeys, selectMetrics...)
	return columns, resultMetrics, err
}

// GetSQLQueryAndParametersForAdwordsQueryV1 - @Kark TODO v1
// TODO Understand null cases.
func (pg *Postgres) GetSQLQueryAndParametersForAdwordsQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool) (string, []interface{}, []string, []string, error) {
	var selectMetrics []string
	var selectKeys []string
	var sql string
	var params []interface{}
	transformedQuery, customerAccountID, err := pg.transFormRequestFieldsAndFetchRequiredFieldsForAdwords(projectID, *query, reqID)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), err
	}
	sql, params, selectKeys, selectMetrics = buildAdwordsSimpleQueryV2(transformedQuery, projectID, *customerAccountID, reqID, fetchSource)
	return sql, params, selectKeys, selectMetrics, nil
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
		return &model.ChannelQueryV1{}, nil, errors.New("adwords document integration not available for this project.")
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
	transformedQuery.From = U.GetDateAsStringZ(query.From, U.TimeZoneString(query.Timezone))
	transformedQuery.To = U.GetDateAsStringZ(query.To, U.TimeZoneString(query.Timezone))
	transformedQuery.Timezone = query.Timezone
	transformedQuery.GroupByTimestamp = query.GroupByTimestamp

	return transformedQuery, nil
}

// @Kark TODO v1
func getAdwordsSpecificMetrics(requestSelectMetrics []string) ([]string, error) {
	resultMetrics := make([]string, 0, 0)
	for _, requestMetric := range requestSelectMetrics {
		metric, isPresent := adwordsExtToInternal[requestMetric]
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
		filterObject, isPresent := adwordsExtToInternal[requestFilter.Object]
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
		groupByObject, isPresent := adwordsExtToInternal[requestGroupBy.Object]
		if !isPresent {
			return make([]model.ChannelGroupBy, 0, 0), errors.New("Invalid groupby key found for document type")
		}
		resultGroupBy = requestGroupBy
		resultGroupBy.Object = groupByObject
		resultGroupBys = append(resultGroupBys, resultGroupBy)
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
func buildAdwordsSimpleQueryV2(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, reqID string, fetchSource bool) (string, []interface{}, []string, []string) {
	lowestHierarchyLevel := getLowestHierarchyLevelForAdwords(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_performance_report"
	return getSQLAndParamsForAdwordsV2(query, projectID, query.From, query.To, customerAccountID, AdwordsDocumentTypeAlias[lowestHierarchyReportLevel], fetchSource)
}

func getSQLAndParamsForAdwordsV2(query *model.ChannelQueryV1, projectID uint64, from, to int64, customerAccountID string,
	docType int, fetchSource bool) (string, []interface{}, []string, []string) {
	computeHigherOrderMetricsHere := !fetchSource
	customerAccountIDs := strings.Split(customerAccountID, ",")
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	filterPropertiesStatement := ""
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
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
	if fetchSource {
		internalValue := adwordsStringColumn
		externalValue := source
		expression := fmt.Sprintf("'%s' as %s", internalValue, externalValue)
		dimensions.selectExpressions = append(dimensions.selectExpressions, expression)
		dimensions.values = append(dimensions.values, externalValue)
	}
	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		internalValue := adwordsInternalPropertiesToReportsInternal[key]
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
		if selectMetric == impressions {
			toFetchImpressionsForHigherOrderMetric = false
			break
		} else if !isNonHigherOrderMetric && !computeHigherOrderMetricsHere {
			toFetchImpressionsForHigherOrderMetric = true
		}
	}

	if toFetchImpressionsForHigherOrderMetric {
		internalValue := adwordsInternalMetricsToAllRep[impressions].nonHigherOrderExpression
		externalValue := adwordsInternalMetricsToAllRep[impressions].externalValue
		expression := fmt.Sprintf("%s as %s", internalValue, externalValue)
		metrics.selectExpressions = append(metrics.selectExpressions, expression)
		metrics.values = append(metrics.values, externalValue)
	}

	// Filters
	filterPropertiesStatement, filterParams := getFilterPropertiesForAdwordsReports(query.Filters)
	finalWhereStatement = joinWithWordInBetween("AND", staticWhereStatementForAdwords, filterPropertiesStatement)
	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, filterParams...)

	finalGroupByKeys = dimensions.values
	if len(finalGroupByKeys) != 0 {
		finalGroupByStatement = " GROUP BY " + joinWithComma(finalGroupByKeys...)
	}

	// orderBy
	finalOrderByKeys = appendSuffix("DESC", metrics.values...)
	if len(finalOrderByKeys) != 0 {
		finalOrderByStatement = " ORDER BY " + joinWithComma(finalOrderByKeys...)
	}

	finalSelectKeys = append(finalSelectKeys, dimensions.selectExpressions...)
	finalSelectKeys = append(finalSelectKeys, metrics.selectExpressions...)
	finalSelectStatement = "SELECT " + joinWithComma(finalSelectKeys...)

	// finalSQL
	resultantSQLStatement = finalSelectStatement + fromAdwordsDocument + finalWhereStatement +
		finalGroupByStatement + finalOrderByStatement + channeAnalyticsLimit
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
		filterOperator := getOp(filter.Condition)
		if filter.Condition == model.ContainsOpStr || filter.Condition == model.NotContainsOpStr {
			filterValue = fmt.Sprintf("%%%s%%", filter.Value)
		} else {
			filterValue = filter.Value
		}
		key := fmt.Sprintf("%s:%s", filter.Object, filter.Property)
		currentFilterProperty = adwordsInternalPropertiesToReportsInternal[key]
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

// @TODO Kark v0
func (pg *Postgres) GetAdwordsChannelResultMeta(projectID uint64, customerAccountID string,
	query *model.ChannelQuery) (*model.ChannelQueryResultMeta, error) {

	customerAccountIDArray := strings.Split(customerAccountID, ",")
	stmnt := "SELECT value->>'currency_code' as currency FROM adwords_documents" +
		" " + "WHERE project_id=? AND customer_account_id IN (?) AND type=? AND timestamp BETWEEN ? AND ?" +
		" " + "ORDER BY timestamp DESC LIMIT 1"

	logCtx := log.WithField("project_id", projectID)

	rows, err := pg.ExecQueryWithContext(stmnt, []interface{}{projectID, customerAccountIDArray,
		model.AdwordsDocumentTypeAlias["customer_account_properties"],
		GetAdwordsDateOnlyTimestamp(query.From),
		GetAdwordsDateOnlyTimestamp(query.To)})
	if err != nil {
		logCtx.WithError(err).Error("Failed to build meta for channel query result.")
		return nil, err
	}
	defer rows.Close()

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
		if key == impressions {
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
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return []string{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId

	db := C.GetServices().Db
	logCtx := log.WithField("project_id", projectID).WithField("doc_type", docType)

	filterValueKey, err := GetAdwordsFilterPropertyKeyByType(docType)
	if err != nil {
		logCtx.WithError(err).Error("Unknown doc type for get adwords filter key.")
		return []string{}, http.StatusBadRequest
	}

	queryStr := "SELECT DISTINCT(value->>?) as filter_value FROM adwords_documents WHERE project_id = ? AND" +
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

	rows, err := pg.ExecQueryWithContext(stmnt, params)
	if err != nil {
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows)

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

	rows, err := pg.ExecQueryWithContext(stmnt, params)
	if err != nil {
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows)
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
