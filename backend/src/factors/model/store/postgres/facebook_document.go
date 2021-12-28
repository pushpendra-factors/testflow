package postgres

import (
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

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	facebookCampaign                                = "campaign"
	facebookAdSet                                   = "ad_set"
	facebookAd                                      = "ad"
	facebookStringColumn                            = "facebook"
	metricsExpressionOfDivisionWithHandleOf0AndNull = "SUM((value->>'%s')::float)*%s/(case when sum((value->>'%s')::float) = 0 then 100000 else NULLIF(sum((value->>'%s')::float), 100000) end)"
)

var facebookDocumentTypeAlias = map[string]int{
	"ad_account":        7,
	"campaign":          1,
	"ad":                2,
	"ad_set":            3,
	"ad_insights":       4,
	"campaign_insights": 5,
	"ad_set_insights":   6,
}

// Note: To change ObjectToValueInFacebookJobsMapping and other constants when we change any of below.
var objectAndPropertyToValueInFacebookReportsMapping = map[string]string{
	"campaign:daily_budget":      "value->>'campaign_daily_budget'",
	"campaign:lifetime_budget":   "value->>'campaign_lifetime_budget'",
	"campaign:configured_status": "value->>'campaign_configured_status'",
	"campaign:effective_status":  "value->>'campaign_effective_status'",
	"campaign:objective":         "value->>'campaign_objective'",
	"campaign:buying_type":       "value->>'campaign_buying_type'",
	"campaign:bid_strategy":      "value->>'campaign_bid_strategy'",
	"campaign:name":              "value->>'campaign_name'",
	"campaign:id":                "campaign_id::bigint",
	"ad_set:daily_budget":        "value->>'ad_set_daily_budget'",
	"ad_set:lifetime_budget":     "value->>'ad_set_lifetime_budget'",
	"ad_set:configured_status":   "value->>'ad_set_configured_status'",
	"ad_set:effective_status":    "value->>'ad_set_effective_status'",
	"ad_set:objective":           "value->>'ad_set_objective'",
	"ad_set:bid_strategy":        "value->>'ad_set_bid_strategy'",
	"ad_set:name":                "value->>'adset_name'",
	"ad_set:id":                  "ad_set_id::bigint",
	"ad:id":                      "ad_id::bigint",
	"ad:name":                    "value->>'ad_name'",
	"ad:configured_status":       "value->>'ad_configured_status'",
	"ad:effective_status":        "value->>'ad_effective_status'",
}

var objectToValueInFacebookFiltersMapping = map[string]string{
	"campaign:daily_budget":      "value->>'campaign_daily_budget'",
	"campaign:lifetime_budget":   "value->>'campaign_lifetime_budget'",
	"campaign:configured_status": "value->>'campaign_configured_status'",
	"campaign:effective_status":  "value->>'campaign_effective_status'",
	"campaign:objective":         "value->>'campaign_objective'",
	"campaign:buying_type":       "value->>'campaign_buying_type'",
	"campaign:bid_strategy":      "value->>'campaign_bid_strategy'",
	"campaign:name":              "value->>'campaign_name'",
	"campaign:id":                "campaign_id",
	"ad_set:daily_budget":        "value->>'ad_set_daily_budget'",
	"ad_set:lifetime_budget":     "value->>'ad_set_lifetime_budget'",
	"ad_set:configured_status":   "value->>'ad_set_configured_status'",
	"ad_set:effective_status":    "value->>'ad_set_effective_status'",
	"ad_set:objective":           "value->>'ad_set_objective'",
	"ad_set:bid_strategy":        "value->>'ad_set_bid_strategy'",
	"ad_set:name":                "value->>'adset_name'",
	"ad_set:id":                  "ad_set_id",
	"ad:id":                      "ad_id::bigint",
	"ad:name":                    "value->>'adset_name'",
	"ad:configured_status":       "value->>'ad_configured_status'",
	"ad:effective_status":        "value->>'ad_effective_status'",
}

var facebookMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions":                   "SUM((value->>'impressions')::float)",
	"clicks":                        "SUM((value->>'clicks')::float)",
	"link_clicks":                   "SUM((value->>'inline_link_clicks')::float)",
	"spend":                         "SUM((value->>'spend')::float)",
	"video_p50_watched_actions":     "SUM((value->>'video_p50_watched_actions')::float)",
	"video_p25_watched_actions":     "SUM((value->>'video_p25_watched_actions')::float)",
	"video_30_sec_watched_actions":  "SUM((value->>'video_30_sec_watched_actions')::float)",
	"video_p100_watched_actions":    "SUM((value->>'video_p100_watched_actions')::float)",
	"video_p75_watched_actions":     "SUM((value->>'video_p75_watched_actions')::float)",
	"cost_per_click":                fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "spend", "1", "clicks", "clicks"),
	"cost_per_link_click":           fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "spend", "1", "inline_link_clicks", "inline_link_clicks"),
	"cost_per_thousand_impressions": fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "spend", "1000", "impressions", "impressions"),
	"click_through_rate":            fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "clicks", "100", "impressions", "impressions"),
	"link_click_through_rate":       fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "inline_link_clicks", "100", "impressions", "impressions"),
	"frequency":                     fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "impressions", "1", "reach", "reach"),
	"reach":                         "SUM((value->>'reach')::float)",
}

const platform = "platform"

var errorEmptyFacebookDocument = errors.New("empty facebook document")

const errorDuplicateFacebookDocument = "pq: duplicate key value violates unique constraint \"facebook_documents_pkey\""

const facebookFilterQueryStr = "SELECT DISTINCT(LOWER(value->>?)) as filter_value FROM facebook_documents WHERE project_id = ? AND" +
	" " + "customer_ad_account_id IN (?) AND type = ? AND value->>? IS NOT NULL LIMIT 5000"

const fromFacebookDocuments = " FROM facebook_documents "

const staticWhereStatementForFacebook = "WHERE project_id = ? AND customer_ad_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "
const staticWhereStatementForFacebookWithSmartProperty = "WHERE facebook_documents.project_id = ? AND facebook_documents.customer_ad_account_id IN ( ? ) AND facebook_documents.type = ? AND facebook_documents.timestamp between ? AND ? "

const facebookAdGroupMetadataFetchQueryStr = "WITH ad_group AS (select ad_set_id as ad_group_id, value->>'name' as ad_group_name, campaign_id " +
	"from facebook_documents where type = ? AND project_id = ? AND timestamp BETWEEN ? AND ? AND customer_ad_account_id IN (?) " +
	"AND (ad_set_id, timestamp) in (select ad_set_id, max(timestamp) from facebook_documents " +
	"where type = ? AND project_id = ? AND timestamp between ? and ? AND customer_ad_account_id IN (?) group by ad_set_id))" +
	", campaign as (select campaign_id, value->>'name' as campaign_name from facebook_documents where type = ? AND " +
	"project_id = ? AND timestamp BETWEEN ? AND ? AND customer_ad_account_id IN (?) and (campaign_id, timestamp) in " +
	"(select campaign_id, max(timestamp) from facebook_documents where type = ? and project_id = ? and timestamp " +
	"BETWEEN ? and ? AND customer_ad_account_id IN (?) group by campaign_id)) select ad_group_id, ad_group_name, " +
	"ad_group.campaign_id, campaign_name from ad_group join campaign on ad_group.campaign_id = campaign.campaign_id"

const facebookCampaignMetadataFetchQueryStr = "select campaign_id, value->>'name' as campaign_name from facebook_documents where type = ? AND " +
	"project_id = ? and timestamp BETWEEN ? and ? AND customer_ad_account_id IN (?) and (campaign_id, timestamp) " +
	"in (select campaign_id, max(timestamp) from facebook_documents where type = ? " +
	"and project_id = ? and timestamp BETWEEN ? and ? AND customer_ad_account_id IN (?) group by campaign_id)"

func isDuplicateFacebookDocumentError(err error) bool {
	return err.Error() == errorDuplicateFacebookDocument
}

// Removed conversions metric on pr - 2089.
// CreateFacebookDocument ...
func (pg *Postgres) CreateFacebookDocument(projectID uint64, document *model.FacebookDocument) int {
	logCtx := log.WithField("customer_acc_id", document.CustomerAdAccountID).WithField(
		"project_id", document.ProjectID)

	if document.CustomerAdAccountID == "" || document.TypeAlias == "" {
		logCtx.Error("Invalid facebook document.")
		return http.StatusBadRequest
	}
	if document.ProjectID == 0 || document.Timestamp == 0 || document.Platform == "" {
		logCtx.Error("Invalid facebook document.")
		return http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", document.TypeAlias)
	docType, docTypeExists := facebookDocumentTypeAlias[document.TypeAlias]
	if !docTypeExists {
		logCtx.Error("Invalid type alias.")
		return http.StatusBadRequest
	}
	document.Type = docType

	campaignIDValue, adSetID, adID, error := getFacebookHierarchyColumnsByType(docType, document.Value)
	if error != nil {
		logCtx.Error("Invalid docType alias.")
		return http.StatusBadRequest
	}
	document.CampaignID = campaignIDValue
	document.AdSetID = adSetID
	document.AdID = adID

	db := C.GetServices().Db
	err := db.Create(&document).Error
	if err != nil {
		if isDuplicateFacebookDocumentError(err) {
			logCtx.WithError(err).WithField("id", document.ID).WithField("platform", document.Platform).Error(
				"Failed to create an facebook doc. Duplicate.")
			return http.StatusConflict
		}
		logCtx.WithError(err).WithField("id", document.ID).WithField("platform", document.Platform).Error(
			"Failed to create an facebook doc. Continued inserting other docs.")
		return http.StatusInternalServerError
	}
	UpdateCountCacheByDocumentType(projectID, &document.CreatedAt, "facebook")
	return http.StatusCreated
}

func getFacebookHierarchyColumnsByType(docType int, valueJSON *postgres.Jsonb) (string, string, string, error) {
	if docType > len(facebookDocumentTypeAlias) {
		return "", "", "", errors.New("invalid document type")
	}

	valueMap, err := U.DecodePostgresJsonb(valueJSON)
	if err != nil {
		return "", "", "", err
	}

	if len(*valueMap) == 0 {
		return "", "", "", errorEmptyFacebookDocument
	}

	switch docType {
	case 1:
		return U.GetStringFromMapOfInterface(*valueMap, "id", ""), "", "", nil
	case 2:
		return U.GetStringFromMapOfInterface(*valueMap, "campaign_id", ""), U.GetStringFromMapOfInterface(*valueMap, "adset_id", ""), U.GetStringFromMapOfInterface(*valueMap, "id", ""), nil
	case 3:
		return U.GetStringFromMapOfInterface(*valueMap, "campaign_id", ""), U.GetStringFromMapOfInterface(*valueMap, "id", ""), "", nil
	case 4, 5, 6:
		return U.GetStringFromMapOfInterface(*valueMap, "campaign_id", ""), U.GetStringFromMapOfInterface(*valueMap, "adset_id", ""), U.GetStringFromMapOfInterface(*valueMap, "ad_id", ""), nil
	default:
		return "", "", "", nil
	}
}

func getFacebookDocumentTypeAliasByType() map[int]string {
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range facebookDocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

// @TODO Kark v1
func (pg *Postgres) buildFbChannelConfig(projectID uint64) *model.ChannelConfigResult {
	facebookObjectsAndProperties := pg.buildObjectAndPropertiesForFacebook(projectID, model.ObjectsForFacebook)
	selectMetrics := append(selectableMetricsForAllChannels, model.SelectableMetricsForFacebook...)
	objectsAndProperties := facebookObjectsAndProperties

	return &model.ChannelConfigResult{
		SelectMetrics:        selectMetrics,
		ObjectsAndProperties: objectsAndProperties,
	}
}

// current
func (pg *Postgres) buildObjectAndPropertiesForFacebook(projectID uint64, objects []string) []model.ChannelObjectAndProperties {
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0)
	for _, currentObject := range objects {
		// to do: check if normal properties present then only smart properties will be there
		propertiesAndRelated, isPresent := model.MapOfFacebookObjectsToPropertiesAndRelated[currentObject]
		var currentProperties []model.ChannelProperty
		var currentPropertiesSmart []model.ChannelProperty
		if isPresent {
			currentProperties = buildProperties(propertiesAndRelated)
			smartProperty := pg.GetSmartPropertyAndRelated(projectID, currentObject, "facebook")
			currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		} else {
			currentProperties = buildProperties(allChannelsPropertyToRelated)
			smartProperty := pg.GetSmartPropertyAndRelated(projectID, currentObject, "facebook")
			currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		}
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
}

// GetFacebookFilterValues - @TODO Kark v1
func (pg *Postgres) GetFacebookFilterValues(projectID uint64, requestFilterObject string,
	requestFilterProperty string, reqID string) ([]interface{}, int) {

	_, isPresent := Const.SmartPropertyReservedNames[requestFilterProperty]
	if !isPresent {
		filterValues, errCode := pg.getSmartPropertyFilterValues(projectID, requestFilterObject, requestFilterProperty, "facebook", reqID)
		if errCode != http.StatusFound {
			return []interface{}{}, http.StatusInternalServerError
		}
		return filterValues, http.StatusFound
	}
	facebookInternalFilterProperty, docType, err := getFilterRelatedInformationForFacebook(
		requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return make([]interface{}, 0, 0), http.StatusBadRequest
	}
	filterValues, errCode := pg.getFacebookFilterValuesByType(projectID, docType,
		facebookInternalFilterProperty, reqID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

// GetFacebookSQLQueryAndParametersForFilterValues - @TODO Kark v1
func (pg *Postgres) GetFacebookSQLQueryAndParametersForFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int) {
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	facebookInternalFilterProperty, docType, err := getFilterRelatedInformationForFacebook(requestFilterObject,
		requestFilterProperty)
	if err != http.StatusOK {
		return "", make([]interface{}, 0, 0), http.StatusBadRequest
	}
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("failed to fetch Project Setting in facebook filter values.")
		return "", make([]interface{}, 0, 0), http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntFacebookAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		logCtx.Info(integrationNotAvailable)
		return "", make([]interface{}, 0, 0), http.StatusNotFound
	}
	customerAccountIDs := strings.Split(customerAccountID, ",")
	params := []interface{}{facebookInternalFilterProperty, projectID, customerAccountIDs,
		docType, facebookInternalFilterProperty}

	return "(" + facebookFilterQueryStr + ")", params, http.StatusFound
}

func getFilterRelatedInformationForFacebook(requestFilterObject string,
	requestFilterProperty string) (string, int, int) {

	facebookInternalFilterObject, isPresent := model.FacebookExternalRepresentationToInternalRepresentation[requestFilterObject]
	if !isPresent {
		log.Error("Invalid facebook filter object.")
		return "", 0, http.StatusBadRequest
	}
	facebookInternalFilterProperty, isPresent := model.FacebookExternalRepresentationToInternalRepresentation[requestFilterProperty]
	if !isPresent {
		log.Error("Invalid facebook filter property.")
		return "", 0, http.StatusBadRequest
	}
	docType := facebookDocumentTypeAlias[facebookInternalFilterObject]

	return facebookInternalFilterProperty, docType, http.StatusOK
}

// @TODO Kark v1
func (pg *Postgres) getFacebookFilterValuesByType(projectID uint64, docType int, property string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID).WithField("doc_type", docType).WithField("property", property)
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("failed to fetch project setting in facebook filter values.")
		return []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntFacebookAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		logCtx.Info(integrationNotAvailable)
		return nil, http.StatusNotFound
	}
	customerAccountIDs := strings.Split(customerAccountID, ",")

	logCtx = logCtx.WithField("doc_type", docType)
	params := []interface{}{property, projectID, customerAccountIDs, docType, property}
	_, resultRows, err := pg.ExecuteSQL(facebookFilterQueryStr, params, logCtx)
	if err != nil {
		logCtx.WithError(err).WithField("query", facebookFilterQueryStr).WithField("params", params).Error(model.FacebookSpecificError)
		return make([]interface{}, 0, 0), http.StatusInternalServerError
	}
	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

// ExecuteFacebookChannelQueryV1 - @Kark TODO v1
// In this flow, Job represents the meta data associated with particular object type. Reports represent data with metrics and few filters.
// TODO - Duplicate code/flow in facebook and adwords.
func (pg *Postgres) ExecuteFacebookChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	var fetchSource = false
	logCtx := log.WithField("xreq_id", reqID)
	if query.GroupByTimestamp == "" {
		sql, params, selectKeys, selectMetrics, errCode := pg.GetSQLQueryAndParametersForFacebookQueryV1(projectID,
			query, reqID, fetchSource, " LIMIT 10000", false, nil)
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := pg.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.FacebookSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	} else {
		sql, params, selectKeys, selectMetrics, errCode := pg.GetSQLQueryAndParametersForFacebookQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 100", false, nil)
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := pg.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.FacebookSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		groupByCombinations := model.GetGroupByCombinationsForChannelAnalytics(columns, resultMetrics)
		// log.Fatal(groupByCombinations)
		sql, params, selectKeys, selectMetrics, errCode = pg.GetSQLQueryAndParametersForFacebookQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 10000", true, groupByCombinations)
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err = pg.ExecuteSQL(sql, params, logCtx)
		columns = append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.FacebookSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	}
}

// GetSQLQueryAndParametersForFacebookQueryV1 ...
func (pg *Postgres) GetSQLQueryAndParametersForFacebookQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, int) {
	var selectMetrics []string
	var selectKeys []string
	var sql string
	var params []interface{}
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	transformedQuery, customerAccountID, err := pg.transFormRequestFieldsAndFetchRequiredFieldsForFacebook(
		projectID, *query, reqID)
	if err != nil && err.Error() == integrationNotAvailable {
		logCtx.WithError(err).Info(model.FacebookSpecificError)
		return "", nil, make([]string, 0, 0), make([]string, 0, 0), http.StatusNotFound
	}
	if err != nil {
		logCtx.WithError(err).Error(model.FacebookSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusBadRequest
	}
	isSmartPropertyPresent := checkSmartProperty(query.Filters, query.GroupBy)
	if isSmartPropertyPresent {
		sql, params, selectKeys, selectMetrics, err = buildFacebookQueryWithSmartPropertyV1(transformedQuery, projectID, customerAccountID, fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
		if err != nil {
			return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
		}
		return sql, params, selectKeys, selectMetrics, http.StatusOK
	}

	sql, params, selectKeys, selectMetrics, err = buildFacebookQueryV1(transformedQuery, projectID, customerAccountID, fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
	}
	return sql, params, selectKeys, selectMetrics, http.StatusOK
}

func (pg *Postgres) transFormRequestFieldsAndFetchRequiredFieldsForFacebook(projectID uint64,
	query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, string, error) {

	var transformedQuery model.ChannelQueryV1
	var err error
	logCtx := log.WithField("req_id", reqID)
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return &model.ChannelQueryV1{}, "", errors.New("Project setting not found")
	}
	customerAccountID := projectSetting.IntFacebookAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		return &model.ChannelQueryV1{}, "", errors.New(integrationNotAvailable)
	}

	transformedQuery, err = convertFromRequestToFacebookSpecificRepresentation(query)
	if err != nil {
		logCtx.Warn("Request failed in validation: ", err)
		return &model.ChannelQueryV1{}, "", err
	}
	return &transformedQuery, customerAccountID, nil
}

// @Kark TODO v1
// Currently, this relies on assumption of Object across different filterObjects. Change when we need robust.
func convertFromRequestToFacebookSpecificRepresentation(query model.ChannelQueryV1) (model.ChannelQueryV1, error) {
	var transformedQuery model.ChannelQueryV1
	var err1, err2, err3 error
	transformedQuery.SelectMetrics, err1 = getFacebookSpecificMetrics(query.SelectMetrics)
	transformedQuery.Filters, err2 = getFacebookSpecificFilters(query.Filters)
	transformedQuery.GroupBy, err3 = getFacebookSpecificGroupBy(query.GroupBy)
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
func getFacebookSpecificMetrics(requestSelectMetrics []string) ([]string, error) {
	resultMetrics := make([]string, 0, 0)
	for _, requestMetric := range requestSelectMetrics {
		metric, isPresent := model.FacebookExternalRepresentationToInternalRepresentation[requestMetric]
		if !isPresent {
			return make([]string, 0, 0), errors.New("Invalid metric found for document type")
		}
		resultMetrics = append(resultMetrics, metric)
	}
	return resultMetrics, nil
}

// @Kark TODO v1
func getFacebookSpecificFilters(requestFilters []model.ChannelFilterV1) ([]model.ChannelFilterV1, error) {
	resultFilters := make([]model.ChannelFilterV1, 0, 0)
	for _, requestFilter := range requestFilters {
		var resultFilter model.ChannelFilterV1
		filterObject, isPresent := model.FacebookExternalRepresentationToInternalRepresentation[requestFilter.Object]
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
func getFacebookSpecificGroupBy(requestGroupBys []model.ChannelGroupBy) ([]model.ChannelGroupBy, error) {
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

	resultGroupBys := make([]model.ChannelGroupBy, 0, 0)
	for _, requestGroupBy := range sortedGroupBys {
		var resultGroupBy model.ChannelGroupBy
		groupByObject, isPresent := model.FacebookExternalRepresentationToInternalRepresentation[requestGroupBy.Object]
		if !isPresent {
			return make([]model.ChannelGroupBy, 0, 0), errors.New("Invalid groupby key found for document type")
		}
		resultGroupBy = requestGroupBy
		resultGroupBy.Object = groupByObject
		resultGroupBys = append(resultGroupBys, resultGroupBy)
	}
	return resultGroupBys, nil
}

func buildFacebookQueryV1(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, error) {
	lowestHierarchyLevel := getLowestHierarchyLevelForFacebook(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_insights"
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromFacebookReports(query, projectID, query.From, query.To, customerAccountID, facebookDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	return sql, params, selectKeys, selectMetrics, nil
}
func buildFacebookQueryWithSmartPropertyV1(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, error) {
	lowestHierarchyLevel := getLowestHierarchyLevelForFacebook(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_insights"
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromFacebookReportsWithSmartProperty(query, projectID, query.From, query.To, customerAccountID, facebookDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	return sql, params, selectKeys, selectMetrics, nil
}

func getSQLAndParamsFromFacebookReportsWithSmartProperty(query *model.ChannelQueryV1, projectID uint64, from, to int64, facebookAccountIDs string,
	docType int, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string) {
	customerAccountIDs := strings.Split(facebookAccountIDs, ",")
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	finalSelectKeys := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	smartPropertyCampaignGroupBys := make([]model.ChannelGroupBy, 0, 0)
	smartPropertyAdGroupGroupBys := make([]model.ChannelGroupBy, 0, 0)
	facebookGroupBys := make([]model.ChannelGroupBy, 0, 0)
	// Group By
	for _, groupBy := range query.GroupBy {
		_, isPresent := Const.SmartPropertyReservedNames[groupBy.Property]
		if !isPresent {
			if groupBy.Object == model.AdwordsCampaign {
				smartPropertyCampaignGroupBys = append(smartPropertyCampaignGroupBys, groupBy)
				groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, fmt.Sprintf("campaign_%s", groupBy.Property))
			} else {
				smartPropertyAdGroupGroupBys = append(smartPropertyAdGroupGroupBys, groupBy)
				groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, fmt.Sprintf("ad_group_%s", groupBy.Property))
			}
		} else {
			key := groupBy.Object + ":" + groupBy.Property
			groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, model.FacebookInternalRepresentationToExternalRepresentation[key])
			facebookGroupBys = append(facebookGroupBys, groupBy)
		}
	}
	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, model.AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	// SelectKeys

	for _, groupBy := range facebookGroupBys {
		key := groupBy.Object + ":" + groupBy.Property
		value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInFacebookReportsMapping[key], model.FacebookInternalRepresentationToExternalRepresentation[key])
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, model.FacebookInternalRepresentationToExternalRepresentation[key])
	}
	for _, groupBy := range smartPropertyCampaignGroupBys {
		value := fmt.Sprintf("campaign.properties->>'%s' as campaign_%s", groupBy.Property, groupBy.Property)
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, fmt.Sprintf("campaign_%s", groupBy.Property))
	}
	for _, groupBy := range smartPropertyAdGroupGroupBys {
		value := fmt.Sprintf("ad_group.properties->>'%s' as ad_group_%s", groupBy.Property, groupBy.Property)
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, fmt.Sprintf("ad_group_%s", groupBy.Property))
	}

	finalSelectKeys = append(finalSelectKeys, selectKeys...)
	if isGroupByTimestamp {
		finalSelectKeys = append(finalSelectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}

	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", facebookMetricsToAggregatesInReportsMapping[selectMetric], model.FacebookInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = model.FacebookInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters := getFacebookFiltersWhereStatementWithSmartProperty(query.Filters, smartPropertyCampaignGroupBys, smartPropertyAdGroupGroupBys)
	filterStatementForSmartPropertyGroupBy := getNotNullFilterStatementForSmartPropertyGroupBys(smartPropertyCampaignGroupBys, smartPropertyAdGroupGroupBys)
	finalFilterStatement := joinWithWordInBetween("AND", staticWhereStatementForFacebookWithSmartProperty, whereConditionForFilters, filterStatementForSmartPropertyGroupBy)
	finalParams := make([]interface{}, 0)
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForFacebook(groupByCombinationsForGBT)
		finalFilterStatement += (" AND (" + whereConditionForGBT + ")")
		finalParams = append(finalParams, whereParams...)
	}
	fromStatement := getFacebookFromStatementWithJoins(query.Filters, query.GroupBy)
	resultSQLStatement := selectQuery + fromStatement + finalFilterStatement
	if len(groupByStatement) != 0 {
		resultSQLStatement += " GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + limitString + ";"

	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

func getFacebookFromStatementWithJoins(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) string {
	isPresentCampaignSmartProperty, isPresentAdGroupSmartProperty := checkSmartPropertyWithTypeAndSource(filters, groupBys, "facebook")
	fromStatement := fromFacebookDocuments
	if isPresentAdGroupSmartProperty {
		fromStatement += "inner join smart_properties ad_group on ad_group.project_id = facebook_documents.project_id and ad_group.object_id = ad_set_id "
	}
	if isPresentCampaignSmartProperty {
		fromStatement += "inner join smart_properties campaign on campaign.project_id = facebook_documents.project_id and campaign.object_id = campaign_id "
	}
	return fromStatement
}
func getSQLAndParamsFromFacebookReports(query *model.ChannelQueryV1, projectID uint64, from, to int64, facebookAccountIDs string,
	docType int, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string) {
	customerAccountIDs := strings.Split(facebookAccountIDs, ",")
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	finalSelectKeys := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	// Group By
	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, model.FacebookInternalRepresentationToExternalRepresentation[key])
	}
	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, model.AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	// SelectKeys

	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInFacebookReportsMapping[key], model.FacebookInternalRepresentationToExternalRepresentation[key])
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, model.FacebookInternalRepresentationToExternalRepresentation[key])
	}

	finalSelectKeys = append(finalSelectKeys, selectKeys...)
	if isGroupByTimestamp {
		finalSelectKeys = append(finalSelectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}

	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", facebookMetricsToAggregatesInReportsMapping[selectMetric], model.FacebookInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = model.FacebookInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters := getFacebookFiltersWhereStatement(query.Filters)
	finalParams := make([]interface{}, 0)
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForFacebook(groupByCombinationsForGBT)
		whereConditionForFilters += (" AND (" + whereConditionForGBT + ")")
		finalParams = append(finalParams, whereParams...)
	}

	resultSQLStatement := selectQuery + fromFacebookDocuments + staticWhereStatementForFacebook + whereConditionForFilters
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + limitString + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}
func getFacebookFiltersWhereStatement(filters []model.ChannelFilterV1) string {
	resultStatement := ""
	var filterValue string
	for index, filter := range filters {
		currentFilterStatement := ""
		if filter.LogicalOp == "" {
			filter.LogicalOp = "AND"
		}
		filterOperator := getOp(filter.Condition, "categorical")
		if filter.Condition == model.ContainsOpStr || filter.Condition == model.NotContainsOpStr {
			filterValue = fmt.Sprintf("%%%s%%", filter.Value)
		} else {
			filterValue = filter.Value
		}
		currentFilterStatement = fmt.Sprintf("%s %s '%s' ", objectToValueInFacebookFiltersMapping[filter.Object+":"+filter.Property], filterOperator, filterValue)
		if index == 0 {
			resultStatement = " AND " + currentFilterStatement
		} else {
			resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement
}
func buildWhereConditionForGBTForFacebook(groupByCombinations []map[string]interface{}) (string, []interface{}) {
	whereConditionForGBT := ""
	params := make([]interface{}, 0)
	filterStringFacebook := "facebook_documents"
	filterStringSmartPropertiesCampaign := "campaign.properties"
	filterStringSmartPropertiesAdGroup := "ad_group.properties"
	for _, groupByCombination := range groupByCombinations {
		whereConditionForEachCombination := ""
		for dimension, value := range groupByCombination {
			filterString := ""
			if strings.HasPrefix(dimension, model.CampaignPrefix) {
				key := fmt.Sprintf(`%s:%s`, "campaign", strings.TrimPrefix(dimension, model.CampaignPrefix))
				currentFilterKey, isPresent := objectToValueInFacebookFiltersMapping[key]
				if isPresent {
					filterString = fmt.Sprintf("%s.%s", filterStringFacebook, currentFilterKey)
				} else {
					filterString = fmt.Sprintf("%s->>'%s'", filterStringSmartPropertiesCampaign, strings.TrimPrefix(dimension, model.CampaignPrefix))
				}
			} else if strings.HasPrefix(dimension, model.AdgroupPrefix) {
				key := fmt.Sprintf(`%s:%s`, "ad_set", strings.TrimPrefix(dimension, model.AdgroupPrefix))
				currentFilterKey, isPresent := objectToValueInFacebookFiltersMapping[key]
				if isPresent {
					filterString = fmt.Sprintf("%s.%s", filterStringFacebook, currentFilterKey)
				} else {
					filterString = fmt.Sprintf("%s->>'%s'", filterStringSmartPropertiesAdGroup, strings.TrimPrefix(dimension, model.AdgroupPrefix))
				}
			} else {
				key := fmt.Sprintf(`%s:%s`, "ad", strings.TrimPrefix(dimension, model.KeywordPrefix))
				currentFilterKey := objectToValueInFacebookFiltersMapping[key]
				filterString = fmt.Sprintf("%s.%s", filterStringFacebook, currentFilterKey)
			}
			if whereConditionForEachCombination == "" {
				if value != nil {
					whereConditionForEachCombination = fmt.Sprintf("%s = ? ", filterString)
					params = append(params, value)
				} else {
					whereConditionForEachCombination = fmt.Sprintf("%s is null ", filterString)
				}
			} else {
				if value != nil {
					whereConditionForEachCombination += fmt.Sprintf(" AND %s = ? ", filterString)
					params = append(params, value)
				} else {
					whereConditionForEachCombination += fmt.Sprintf(" AND %s is null ", filterString)
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

func getFacebookFiltersWhereStatementWithSmartProperty(filters []model.ChannelFilterV1, smartPropertyCampaignGroupBys []model.ChannelGroupBy, smartPropertyAdGroupGroupBys []model.ChannelGroupBy) string {
	resultStatement := ""
	var filterValue string
	campaignFilter := ""
	adGroupFilter := ""
	for index, filter := range filters {
		currentFilterStatement := ""
		if filter.LogicalOp == "" {
			filter.LogicalOp = "AND"
		}
		filterOperator := getOp(filter.Condition, "categorical")
		if filter.Condition == model.ContainsOpStr || filter.Condition == model.NotContainsOpStr {
			filterValue = fmt.Sprintf("%%%s%%", filter.Value)
		} else {
			filterValue = filter.Value
		}
		_, isPresent := Const.SmartPropertyReservedNames[filter.Property]
		if isPresent {
			currentFilterStatement = fmt.Sprintf("%s %s '%s' ", objectToValueInFacebookFiltersMapping[filter.Object+":"+filter.Property], filterOperator, filterValue)
			if index == 0 {
				resultStatement = " AND " + currentFilterStatement
			} else {
				resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
			}
		} else {
			currentFilterStatement = fmt.Sprintf("%s.properties->>'%s' %s '%s'", model.FacebookObjectMapForSmartProperty[filter.Object], filter.Property, filterOperator, filterValue)
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
	if resultStatement == "" {
		return resultStatement
	}
	return resultStatement + ")"
}

// @TODO Kark v1
// Complexity consideration - Having at max of 20 filters and 20 group by should be fine.
// change algo/strategy the filters and group by goes beyond 100.
func getLowestHierarchyLevelForFacebook(query *model.ChannelQueryV1) string {
	// Fetch the propertyNames
	var objectNames []string
	for _, filter := range query.Filters {
		objectNames = append(objectNames, filter.Object)
	}

	for _, groupBy := range query.GroupBy {
		objectNames = append(objectNames, groupBy.Object)
	}

	// Check if present
	for _, objectName := range objectNames {
		if objectName == facebookAd {
			return facebookAd
		}
	}

	for _, objectName := range objectNames {
		if objectName == facebookAdSet {
			return facebookAdSet
		}
	}

	for _, objectName := range objectNames {
		if objectName == facebookCampaign {
			return facebookCampaign
		}
	}
	return facebookCampaign
}

// GetFacebookLastSyncInfo ...
func (pg *Postgres) GetFacebookLastSyncInfo(projectID uint64, CustomerAdAccountID string) ([]model.FacebookLastSyncInfo, int) {
	db := C.GetServices().Db

	facebookLastSyncInfos := make([]model.FacebookLastSyncInfo, 0, 0)

	queryStr := "SELECT project_id, customer_ad_account_id, type as document_type, max(timestamp) as last_timestamp" +
		" FROM facebook_documents WHERE project_id = ? AND customer_ad_account_id = ?" +
		" GROUP BY project_id, customer_ad_account_id, type "

	rows, err := db.Raw(queryStr, projectID, CustomerAdAccountID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last facebook documents by type for sync info.")
		return facebookLastSyncInfos, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var facebookLastSyncInfo model.FacebookLastSyncInfo
		if err := db.ScanRows(rows, &facebookLastSyncInfo); err != nil {
			log.WithError(err).Error("Failed to scan last facebook documents by type for sync info.")
			return []model.FacebookLastSyncInfo{}, http.StatusInternalServerError
		}

		facebookLastSyncInfos = append(facebookLastSyncInfos, facebookLastSyncInfo)
	}
	documentTypeAliasByType := getFacebookDocumentTypeAliasByType()

	for i := range facebookLastSyncInfos {
		logCtx := log.WithField("project_id", facebookLastSyncInfos[i].ProjectID)
		typeAlias, typeAliasExists := documentTypeAliasByType[facebookLastSyncInfos[i].DocumentType]
		if !typeAliasExists {
			logCtx.WithField("document_type",
				facebookLastSyncInfos[i].DocumentType).Error("Invalid document type given. No type alias name.")
			continue
		}

		facebookLastSyncInfos[i].DocumentTypeAlias = typeAlias
	}
	return facebookLastSyncInfos, http.StatusOK
}

// ExecuteFacebookChannelQuery - @TODO Kark v0
func (pg *Postgres) ExecuteFacebookChannelQuery(projectID uint64,
	query *model.ChannelQuery) (*model.ChannelQueryResult, int) {
	logCtx := log.WithField("project_id", projectID).WithField("query", query)

	if projectID == 0 || query == nil {
		logCtx.Error("Invalid project_id or query on execute facebook channel query.")
		return nil, http.StatusInternalServerError
	}

	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	if projectSetting.IntFacebookAdAccount == "" {
		logCtx.Error("Execute facebook channel query failed. No customer account id.")
		return nil, http.StatusInternalServerError
	}

	query.From = ChangeUnixTimestampToDate(query.From)
	query.To = ChangeUnixTimestampToDate(query.To)
	queryResult := &model.ChannelQueryResult{}
	result, err := pg.GetFacebookChannelResult(projectID, projectSetting.IntFacebookAdAccount, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get facebook query result.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult = result
	if query.Breakdown == "" {
		return queryResult, http.StatusOK
	}

	metricBreakDown, err := pg.GetFacebookMetricBreakdown(projectID, projectSetting.IntFacebookAdAccount, query)
	queryResult.MetricsBreakdown = metricBreakDown

	// sort only if the impression is there as column
	impressionsIndex := 0
	for _, key := range queryResult.MetricsBreakdown.Headers {
		if key == "impressions" {
			// sort the rows by impressions count in descending order
			sort.Slice(queryResult.MetricsBreakdown.Rows, func(i, j int) bool {
				return queryResult.MetricsBreakdown.Rows[i][impressionsIndex].(float64) >
					queryResult.MetricsBreakdown.Rows[j][impressionsIndex].(float64)
			})
			break
		}
		impressionsIndex++
	}
	return queryResult, http.StatusOK
}

// @TODO Kark v0
func (pg *Postgres) GetFacebookMetricBreakdown(projectID uint64, customerAccountID string,
	query *model.ChannelQuery) (*model.ChannelBreakdownResult, error) {

	logCtx := log.WithField("project_id", projectID).WithField("customer_account_id", customerAccountID)

	sqlQuery, documentType := getFacebookMetricsQuery(query, true)

	rows, tx, err := pg.ExecQueryWithContext(sqlQuery, []interface{}{projectID, customerAccountID,
		query.From,
		query.To,
		documentType})
	if err != nil {
		logCtx.WithError(err).Error("Failed to build channel query result.")
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows, tx)
	if err != nil {
		return nil, err
	}

	for i := range resultHeaders {
		if resultHeaders[i] == CAChannelGroupKey {
			resultHeaders[i] = query.Breakdown
		}
	}

	for ri := range resultRows {
		for ci := range resultRows[ri] {
			if ci > 0 && resultRows[ri][ci] == nil {
				resultRows[ri][ci] = 0
			}
		}
	}

	return &model.ChannelBreakdownResult{Headers: resultHeaders, Rows: resultRows}, nil
}

// @TODO Kark v0
func getFacebookDocumentType(query *model.ChannelQuery) int {
	var documentType int
	if query.FilterKey == "ad" {
		documentType = 4
	}
	if query.FilterKey == "campaign" {
		documentType = 5
	}
	if query.FilterKey == "adset" {
		documentType = 6
	}
	return documentType
}

// @TODO Kark v0
func getFacebookMetricsQuery(query *model.ChannelQuery, withBreakdown bool) (string, int) {

	documentType := getFacebookDocumentType(query)

	selectColstWithoutAlias := "SUM((value->>'impressions')::float) as %s , SUM((value->>'clicks')::float) as %s," +
		" " + "SUM((value->>'spend')::float) as %s," +
		" " + "SUM((value->>'unique_clicks')::float) as %s," +
		" " + "SUM((value->>'reach')::float) as %s, AVG((value->>'frequency')::float) as %s, " +
		" " + "SUM((value->>'inline_post_engagement')::float) as %s," +
		" " + "AVG((value->>'cpc')::float) as %s"

	selectCols := fmt.Sprintf(selectColstWithoutAlias, CAColumnImpressions, CAColumnClicks,
		CAColumnTotalCost, CAColumnUniqueClicks, CAColumnReach,
		CAColumnFrequency, CAColumnInlinePostEngagement, CAColumnCostPerClick)

	strmntWhere := "WHERE project_id= ? AND customer_ad_account_id = ? AND timestamp BETWEEN ? AND ? AND type=? and platform!='facebook_all'"

	strmntGroupBy := ""
	if withBreakdown {
		if query.Breakdown == platform {
			selectCols = platform + ", " + selectCols
			strmntGroupBy = "GROUP BY " + platform
		} else {
			firstValue := "(value->>'%s_name') as name, "
			firstValue = fmt.Sprintf(firstValue, query.Breakdown)
			selectCols = firstValue + selectCols
			strmntGroupBy = "GROUP BY id, (value->>'%s_name')"
			strmntGroupBy = fmt.Sprintf(strmntGroupBy, query.Breakdown)
		}
	}

	sqlQuery := "SELECT" + " " + selectCols + " " + "FROM facebook_documents" + " " + strmntWhere +
		" " + strmntGroupBy
	return sqlQuery, documentType
}

// @TODO Kark v0
func (pg *Postgres) GetFacebookChannelResult(projectID uint64, customerAccountID string,
	query *model.ChannelQuery) (*model.ChannelQueryResult, error) {

	logCtx := log.WithField("project_id", projectID)

	sqlQuery, documentType := getFacebookMetricsQuery(query, false)

	queryResult := &model.ChannelQueryResult{}
	rows, tx, err := pg.ExecQueryWithContext(sqlQuery, []interface{}{projectID, customerAccountID,
		query.From,
		query.To,
		documentType})
	if err != nil {
		logCtx.WithError(err).Error("Failed to build channel query result.")
		return queryResult, err
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
		log.Warn("Aggregate query returned more than one row on get facebook metric kvs.")
	}

	metricKvs := make(map[string]interface{})
	for i, k := range resultHeaders {
		metricKvs[k] = resultRows[0][i]
	}

	queryResult.Metrics = &metricKvs
	return queryResult, nil
}

func (pg *Postgres) GetLatestMetaForFacebookForGivenDays(projectID uint64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields) {
	db := C.GetServices().Db

	channelDocumentsCampaign := make([]model.ChannelDocumentsWithFields, 0)
	channelDocumentsAdGroup := make([]model.ChannelDocumentsWithFields, 0)

	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		log.Error("Failed to get project settings")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	if projectSetting.IntFacebookAdAccount == "" {
		log.Error("Failed to get custtomer account ids")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	customerAccountIDs := strings.Split(projectSetting.IntFacebookAdAccount, ",")

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

	err = db.Raw(facebookAdGroupMetadataFetchQueryStr, facebookDocumentTypeAlias[facebookAdSet], projectID, from, to,
		customerAccountIDs, facebookDocumentTypeAlias[facebookAdSet], projectID, from, to, customerAccountIDs,
		facebookDocumentTypeAlias[facebookCampaign], projectID, from, to, customerAccountIDs,
		facebookDocumentTypeAlias[facebookCampaign], projectID, from, to, customerAccountIDs).Find(&channelDocumentsAdGroup).Error
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d ad_group meta for facebook", days)
		log.Error(errString)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	err = db.Raw(facebookCampaignMetadataFetchQueryStr, facebookDocumentTypeAlias[facebookCampaign], projectID, from, to,
		customerAccountIDs, facebookDocumentTypeAlias[facebookCampaign], projectID, from, to,
		customerAccountIDs).Find(&channelDocumentsCampaign).Error
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d campaign meta for facebook", days)
		log.Error(errString)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	return channelDocumentsCampaign, channelDocumentsAdGroup
}

func (pg *Postgres) DeleteFacebookIntegration(projectID uint64) (int, error) {
	db := C.GetServices().Db
	updateValues := make(map[string]interface{})
	updateValues["int_facebook_access_token"] = nil
	updateValues["int_facebook_email"] = nil
	updateValues["int_facebook_user_id"] = nil
	updateValues["int_facebook_ad_account"] = nil
	updateValues["int_facebook_agent_uuid"] = nil

	err := db.Model(&model.ProjectSetting{}).Where("project_id = ?", projectID).Update(updateValues).Error
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}
