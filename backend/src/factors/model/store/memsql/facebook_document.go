package memsql

import (
	"errors"
	C "factors/config"
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
	facebookCampaign                                = "campaign"
	facebookAdSet                                   = "ad_set"
	facebookAd                                      = "ad"
	facebookStringColumn                            = "facebook"
	metricsExpressionOfDivisionWithHandleOf0AndNull = "SUM(JSON_EXTRACT_STRING(value,'%s'))*%s/(case when sum(JSON_EXTRACT_STRING(value,'%s')) = 0 then 100000 else NULLIF(sum(JSON_EXTRACT_STRING(value,'%s')), 100000) end)"
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

var objectAndPropertyToValueInFacebookReportsMapping = map[string]string{
	"campaign:daily_budget":      "JSON_EXTRACT_STRING(value, 'campaign_daily_budget')",
	"campaign:lifetime_budget":   "JSON_EXTRACT_STRING(value, 'campaign_lifetime_budget')",
	"campaign:configured_status": "JSON_EXTRACT_STRING(value, 'campaign_configured_status')",
	"campaign:effective_status":  "JSON_EXTRACT_STRING(value, 'campaign_effective_status')",
	"campaign:objective":         "JSON_EXTRACT_STRING(value, 'campaign_objective')",
	"campaign:buying_type":       "JSON_EXTRACT_STRING(value, 'campaign_buying_type')",
	"campaign:bid_strategy":      "JSON_EXTRACT_STRING(value, 'campaign_bid_strategy')",
	"campaign:name":              "JSON_EXTRACT_STRING(value, 'campaign_name')",
	"campaign:id":                "campaign_id",
	"ad_set:daily_budget":        "JSON_EXTRACT_STRING(value, 'ad_set_daily_budget')",
	"ad_set:lifetime_budget":     "JSON_EXTRACT_STRING(value, 'ad_set_lifetime_budget')",
	"ad_set:configured_status":   "JSON_EXTRACT_STRING(value, 'ad_set_configured_status')",
	"ad_set:effective_status":    "JSON_EXTRACT_STRING(value, 'ad_set_effective_status')",
	"ad_set:objective":           "JSON_EXTRACT_STRING(value, 'ad_set_objective')",
	"ad_set:bid_strategy":        "JSON_EXTRACT_STRING(value, 'ad_set_bid_strategy')",
	"ad_set:name":                "JSON_EXTRACT_STRING(value, 'adset_name')",
	"ad_set:id":                  "ad_set_id",
	"ad:id":                      "ad_id::bigint",
	"ad:name":                    "JSON_EXTRACT_STRING(value, 'ad_name')",
	"ad:configured_status":       "JSON_EXTRACT_STRING(value, 'ad_configured_status')",
	"ad:effective_status":        "JSON_EXTRACT_STRING(value, 'ad_effective_status')",
}

var objectToValueInFacebookFiltersMapping = map[string]string{
	"campaign:daily_budget":      "JSON_EXTRACT_STRING(value,'campaign_daily_budget')",
	"campaign:lifetime_budget":   "JSON_EXTRACT_STRING(value,'campaign_lifetime_budget')",
	"campaign:configured_status": "JSON_EXTRACT_STRING(value,'campaign_configured_status')",
	"campaign:effective_status":  "JSON_EXTRACT_STRING(value,'campaign_effective_status')",
	"campaign:objective":         "JSON_EXTRACT_STRING(value,'campaign_objective')",
	"campaign:buying_type":       "JSON_EXTRACT_STRING(value,'campaign_buying_type')",
	"campaign:bid_strategy":      "JSON_EXTRACT_STRING(value,'campaign_bid_strategy')",
	"campaign:name":              "JSON_EXTRACT_STRING(value,'campaign_name')",
	"campaign:id":                "campaign_id",
	"ad_set:daily_budget":        "JSON_EXTRACT_STRING(value,'ad_set_daily_budget')",
	"ad_set:lifetime_budget":     "JSON_EXTRACT_STRING(value,'ad_set_lifetime_budget')",
	"ad_set:configured_status":   "JSON_EXTRACT_STRING(value,'ad_set_configured_status')",
	"ad_set:effective_status":    "JSON_EXTRACT_STRING(value,'ad_set_effective_status')",
	"ad_set:objective":           "JSON_EXTRACT_STRING(value,'ad_set_objective')",
	"ad_set:bid_strategy":        "JSON_EXTRACT_STRING(value,'ad_set_bid_strategy')",
	"ad_set:name":                "JSON_EXTRACT_STRING(value,'adset_name')",
	"ad_set:id":                  "ad_set_id",
	"ad:id":                      "ad_id",
	"ad:name":                    "JSON_EXTRACT_STRING(value, 'ad_name')",
	"ad:configured_status":       "JSON_EXTRACT_STRING(value, 'ad_configured_status')",
	"ad:effective_status":        "JSON_EXTRACT_STRING(value, 'ad_effective_status')",
}

var objectToValueInFacebookFiltersMappingWithFacebookDocuments = map[string]string{
	"campaign:daily_budget":      "JSON_EXTRACT_STRING(facebook_documents.value,'campaign_daily_budget')",
	"campaign:lifetime_budget":   "JSON_EXTRACT_STRING(facebook_documents.value,'campaign_lifetime_budget')",
	"campaign:configured_status": "JSON_EXTRACT_STRING(facebook_documents.value,'campaign_configured_status')",
	"campaign:effective_status":  "JSON_EXTRACT_STRING(facebook_documents.value,'campaign_effective_status')",
	"campaign:objective":         "JSON_EXTRACT_STRING(facebook_documents.value,'campaign_objective')",
	"campaign:buying_type":       "JSON_EXTRACT_STRING(facebook_documents.value,'campaign_buying_type')",
	"campaign:bid_strategy":      "JSON_EXTRACT_STRING(facebook_documents.value,'campaign_bid_strategy')",
	"campaign:name":              "JSON_EXTRACT_STRING(facebook_documents.value,'campaign_name')",
	"campaign:id":                "facebook_documents.campaign_id",
	"ad_set:daily_budget":        "JSON_EXTRACT_STRING(facebook_documents.value,'ad_set_daily_budget')",
	"ad_set:lifetime_budget":     "JSON_EXTRACT_STRING(facebook_documents.value,'ad_set_lifetime_budget')",
	"ad_set:configured_status":   "JSON_EXTRACT_STRING(facebook_documents.value,'ad_set_configured_status')",
	"ad_set:effective_status":    "JSON_EXTRACT_STRING(facebook_documents.value,'ad_set_effective_status')",
	"ad_set:objective":           "JSON_EXTRACT_STRING(facebook_documents.value,'ad_set_objective')",
	"ad_set:bid_strategy":        "JSON_EXTRACT_STRING(facebook_documents.value,'ad_set_bid_strategy')",
	"ad_set:name":                "JSON_EXTRACT_STRING(facebook_documents.value,'adset_name')",
	"ad_set:id":                  "facebook_documents.ad_set_id",
	"ad:id":                      "facebook_documents.ad_id",
	"ad:name":                    "JSON_EXTRACT_STRING(facebook_documents.value, 'ad_name')",
	"ad:configured_status":       "JSON_EXTRACT_STRING(facebook_documents.value, 'ad_configured_status')",
	"ad:effective_status":        "JSON_EXTRACT_STRING(facebook_documents.value, 'ad_effective_status')",
}

var facebookMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions":                            "SUM(JSON_EXTRACT_STRING(value,'impressions'))",
	"clicks":                                 "SUM(JSON_EXTRACT_STRING(value,'clicks'))",
	"link_clicks":                            "SUM(JSON_EXTRACT_STRING(value,'inline_link_clicks'))",
	"spend":                                  "SUM(JSON_EXTRACT_STRING(value,'spend'))",
	"video_p50_watched_actions":              "SUM(JSON_EXTRACT_STRING(value,'video_p50_watched_actions'))",
	"video_p25_watched_actions":              "SUM(JSON_EXTRACT_STRING(value,'video_p25_watched_actions'))",
	"video_30_sec_watched_actions":           "SUM(JSON_EXTRACT_STRING(value,'video_30_sec_watched_actions'))",
	"video_p100_watched_actions":             "SUM(JSON_EXTRACT_STRING(value,'video_p100_watched_actions'))",
	"video_p75_watched_actions":              "SUM(JSON_EXTRACT_STRING(value,'video_p75_watched_actions'))",
	"cost_per_click":                         fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "spend", "1", "clicks", "clicks"),
	"cost_per_link_click":                    fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "spend", "1", "inline_link_clicks", "inline_link_clicks"),
	"cost_per_thousand_impressions":          fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "spend", "1000", "impressions", "impressions"),
	"click_through_rate":                     fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "clicks", "100", "impressions", "impressions"),
	"link_click_through_rate":                fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "inline_link_clicks", "100", "impressions", "impressions"),
	"frequency":                              fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "impressions", "1", "reach", "reach"),
	"reach":                                  "SUM(JSON_EXTRACT_STRING(value,'reach'))",
	"fb_pixel_purchase_count":                "SUM(JSON_EXTRACT_STRING(value, 'actions_offsite_conversion.fb_pixel_purchase'))",
	"fb_pixel_purchase_revenue":              "SUM(JSON_EXTRACT_STRING(value, 'action_values_offsite_conversion.fb_pixel_purchase'))",
	"fb_pixel_purchase_cost_per_action_type": fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "spend", "1", "actions_offsite_conversion.fb_pixel_purchase", "actions_offsite_conversion.fb_pixel_purchase"),
	"fb_pixel_purchase_roas":                 fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "action_values_offsite_conversion.fb_pixel_purchase", "1", "spend", "spend"),
}

const platform = "platform"

var errorEmptyFacebookDocument = errors.New("empty facebook document")

const facebookFilterQueryStr = "SELECT DISTINCT(LCASE(JSON_EXTRACT_STRING(value, ?))) as filter_value FROM facebook_documents WHERE project_id = ? AND" +
	" " + "customer_ad_account_id IN (?) AND type = ? AND JSON_EXTRACT_STRING(value, ?) IS NOT NULL AND timestamp BETWEEN ? AND ? LIMIT 5000"

const fromFacebookDocuments = " FROM facebook_documents "

const staticWhereStatementForFacebook = "WHERE project_id = ? AND customer_ad_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "
const staticWhereStatementForFacebookWithSmartProperty = "WHERE facebook_documents.project_id = ? AND facebook_documents.customer_ad_account_id IN ( ? ) AND facebook_documents.type = ? AND facebook_documents.timestamp between ? AND ? "

const facebookAdGroupMetadataFetchQueryStr = "WITH ad_group as (select ad_group_information.campaign_id_1 as campaign_id, ad_group_information.ad_group_id_1 as ad_group_id, ad_group_information.ad_group_name_1 as ad_group_name " +
	"from ( " +
	"select campaign_id as campaign_id_1, ad_set_id as ad_group_id_1, JSON_EXTRACT_STRING(value, 'name') as ad_group_name_1, timestamp " +
	"from facebook_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) " +
	") as ad_group_information " +
	"INNER JOIN " +
	"(select ad_set_id as ad_group_id_1, max(timestamp) as timestamp " +
	"from facebook_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) group by ad_group_id_1 " +
	") as ad_group_latest_timestamp_id " +
	"ON ad_group_information.ad_group_id_1 = ad_group_latest_timestamp_id.ad_group_id_1 AND ad_group_information.timestamp = ad_group_latest_timestamp_id.timestamp), " +

	" campaign as (select campaign_information.campaign_id_1 as campaign_id, campaign_information.campaign_name_1 as campaign_name " +
	"from ( " +
	"select campaign_id as campaign_id_1, JSON_EXTRACT_STRING(value, 'name') as campaign_name_1, timestamp " +
	"from facebook_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) " +
	") as campaign_information " +
	"INNER JOIN " +
	"(select campaign_id as campaign_id_1, max(timestamp) as timestamp " +
	"from facebook_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) group by campaign_id_1 " +
	") as campaign_latest_timestamp_id " +
	"ON campaign_information.campaign_id_1 = campaign_latest_timestamp_id.campaign_id_1 AND campaign_information.timestamp = campaign_latest_timestamp_id.timestamp) " +

	"select campaign.campaign_id as campaign_id, campaign.campaign_name as campaign_name, ad_group.ad_group_id as ad_group_id, ad_group.ad_group_name as ad_group_name " +
	"from campaign join ad_group on ad_group.campaign_id = campaign.campaign_id"

const facebookCampaignMetadataFetchQueryStr = "select campaign_information.campaign_id as campaign_id, campaign_information.campaign_name as campaign_name " +
	"from ( " +
	"select campaign_id, JSON_EXTRACT_STRING(value, 'name') as campaign_name, timestamp " +
	"from facebook_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) " +
	") as campaign_information " +
	"INNER JOIN " +
	"(select campaign_id, max(timestamp) as timestamp " +
	"from facebook_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) group by campaign_id " +
	") as campaign_latest_timestamp_id " +
	"ON campaign_information.campaign_id = campaign_latest_timestamp_id.campaign_id AND campaign_information.timestamp = campaign_latest_timestamp_id.timestamp "

func (store *MemSQL) satisfiesFacebookDocumentForeignConstraints(facebookDocument model.FacebookDocument) int {
	logFields := log.Fields{
		"facebook_document": facebookDocument,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, errCode := store.GetProject(facebookDocument.ProjectID)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func (store *MemSQL) satisfiesFacebookDocumentUniquenessConstraints(facebookDocument *model.FacebookDocument) int {
	logFields := log.Fields{
		"facebook_document": facebookDocument,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	errCode := store.isFacebookDocumentExistByPrimaryKey(facebookDocument)
	if errCode == http.StatusFound {
		return http.StatusConflict
	}
	if errCode == http.StatusNotFound {
		return http.StatusOK
	}
	return errCode
}

// Checks PRIMARY KEY (project_id, customer_ad_account_id, platform, type, timestamp, id)
func (store *MemSQL) isFacebookDocumentExistByPrimaryKey(document *model.FacebookDocument) int {
	logFields := log.Fields{
		"document": document,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if document.ProjectID == 0 || document.CustomerAdAccountID == "" || document.Platform == "" ||
		document.Type == 0 || document.Timestamp == 0 || document.ID == "" {

		log.Error("Invalid facebook document on primary constraint check.")
		return http.StatusBadRequest
	}

	var facebookDocument model.FacebookDocument

	db := C.GetServices().Db
	if err := db.Limit(1).Where(
		"project_id = ? AND customer_ad_account_id = ? AND platform = ? AND type = ? AND timestamp = ? AND id = ?",
		document.ProjectID, document.CustomerAdAccountID, document.Platform, document.Type, document.Timestamp, document.ID,
	).Select("id").Find(&facebookDocument).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}

		logCtx.WithError(err).
			Error("Failed getting to check existence facebook document by primary keys.")
		return http.StatusInternalServerError
	}

	if facebookDocument.ID == "" {
		logCtx.Error("Invalid id value returned on facebook document primary key check.")
		return http.StatusInternalServerError
	}

	return http.StatusFound
}

// CreateFacebookDocument ...
func (store *MemSQL) CreateFacebookDocument(projectID int64, document *model.FacebookDocument) int {
	logFields := log.Fields{
		"document":   document,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

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
	if errCode := store.satisfiesFacebookDocumentForeignConstraints(*document); errCode != http.StatusOK {
		return http.StatusInternalServerError
	}

	errCode := store.satisfiesFacebookDocumentUniquenessConstraints(document)
	if errCode != http.StatusOK {
		return errCode
	}

	db := C.GetServices().Db
	err := db.Create(&document).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
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
	logFields := log.Fields{
		"doc_type":   docType,
		"value_json": valueJSON,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range facebookDocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

// @TODO Kark v1
func (store *MemSQL) buildFbChannelConfig(projectID int64) *model.ChannelConfigResult {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	facebookObjectsAndProperties := store.buildObjectAndPropertiesForFacebook(projectID, model.ObjectsForFacebook)
	selectMetrics := append(SelectableMetricsForAllChannels, model.SelectableMetricsForFacebook...)
	objectsAndProperties := facebookObjectsAndProperties

	return &model.ChannelConfigResult{
		SelectMetrics:        selectMetrics,
		ObjectsAndProperties: objectsAndProperties,
	}
}

func (store *MemSQL) buildObjectAndPropertiesForFacebook(projectID int64, objects []string) []model.ChannelObjectAndProperties {
	logFields := log.Fields{
		"project_id": projectID,
		"objects":    objects,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0, 0)
	for _, currentObject := range objects {
		// to do: check if normal properties present then only smart properties will be there
		propertiesAndRelated, isPresent := model.MapOfFacebookObjectsToPropertiesAndRelated[currentObject]
		var currentProperties []model.ChannelProperty
		var currentPropertiesSmart []model.ChannelProperty
		if isPresent {
			currentProperties = buildProperties(propertiesAndRelated)
			smartProperty := store.GetSmartPropertyAndRelated(projectID, currentObject, "facebook")
			currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		} else {
			currentProperties = buildProperties(allChannelsPropertyToRelated)
			smartProperty := store.GetSmartPropertyAndRelated(projectID, currentObject, "facebook")
			currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		}
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
}

// GetFacebookFilterValues - @TODO Kark v1
func (store *MemSQL) GetFacebookFilterValues(projectID int64, requestFilterObject string,
	requestFilterProperty string, reqID string) ([]interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	_, isPresent := model.SmartPropertyReservedNames[requestFilterProperty]
	if !isPresent {
		filterValues, errCode := store.getSmartPropertyFilterValues(projectID, requestFilterObject, requestFilterProperty, "facebook", reqID)
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
	filterValues, errCode := store.getFacebookFilterValuesByType(projectID, docType,
		facebookInternalFilterProperty, reqID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

func (store *MemSQL) IsFacebookIntegrationAvailable(projectID int64) bool {
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return false
	}
	customerAccountID := projectSetting.IntFacebookAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		return false
	}
	return true
}

// GetFacebookSQLQueryAndParametersForFilterValues - @TODO Kark v1
func (store *MemSQL) GetFacebookSQLQueryAndParametersForFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	facebookInternalFilterProperty, docType, err := getFilterRelatedInformationForFacebook(requestFilterObject,
		requestFilterProperty)
	if err != http.StatusOK {
		return "", make([]interface{}, 0, 0), http.StatusBadRequest
	}
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("failed to fetch Project Setting in facebook filter values.")
		return "", make([]interface{}, 0, 0), http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntFacebookAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		logCtx.Error(integrationNotAvailable)
		return "", make([]interface{}, 0, 0), http.StatusNotFound
	}
	from, to := model.GetFromAndToDatesForFilterValues()
	customerAccountIDs := strings.Split(customerAccountID, ",")
	params := []interface{}{facebookInternalFilterProperty, projectID, customerAccountIDs,
		docType, facebookInternalFilterProperty, from, to}

	return "(" + facebookFilterQueryStr + ")", params, http.StatusFound
}

func getFilterRelatedInformationForFacebook(requestFilterObject string,
	requestFilterProperty string) (string, int, int) {
	logFields := log.Fields{
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
func (store *MemSQL) getFacebookFilterValuesByType(projectID int64, docType int, property string, reqID string) ([]interface{}, int) {
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
		logCtx.Error("failed to fetch project setting in facebook filter values.")
		return []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntFacebookAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		logCtx.Error(integrationNotAvailable)
		return nil, http.StatusNotFound
	}
	customerAccountIDs := strings.Split(customerAccountID, ",")

	logCtx = logCtx.WithField("doc_type", docType)
	from, to := model.GetFromAndToDatesForFilterValues()
	params := []interface{}{property, projectID, customerAccountIDs, docType, property, from, to}
	_, resultRows, err := store.ExecuteSQL(facebookFilterQueryStr, params, logCtx)
	if err != nil {
		logCtx.WithError(err).WithField("query", facebookFilterQueryStr).WithField("params", params).Error(model.FacebookSpecificError)
		return make([]interface{}, 0, 0), http.StatusInternalServerError
	}
	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

// ExecuteFacebookChannelQueryV1 - @Kark TODO v1
// In this flow, Job represents the meta data associated with particular object type. Reports represent data with metrics and few filters.
// TODO - Duplicate code/flow in facebook and adwords.
func (store *MemSQL) ExecuteFacebookChannelQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	var fetchSource = false
	logCtx := log.WithFields(logFields)
	if query.GroupByTimestamp == "" {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForFacebookQueryV1(projectID,
			query, reqID, fetchSource, " LIMIT 10000", false, nil)
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
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.FacebookSpecificError)
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	} else {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForFacebookQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 100", false, nil)
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
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.FacebookSpecificError)
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		groupByCombinations := model.GetGroupByCombinationsForChannelAnalytics(columns, resultMetrics)
		sql, params, selectKeys, selectMetrics, errCode = store.GetSQLQueryAndParametersForFacebookQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 10000", true, groupByCombinations)
		if errCode != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err = store.ExecuteSQL(sql, params, logCtx)
		columns = append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.FacebookSpecificError)
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	}
}

// GetSQLQueryAndParametersForFacebookQueryV1 ...
func (store *MemSQL) GetSQLQueryAndParametersForFacebookQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, int) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"fetch_source":                  fetchSource,
		"req_id":                        reqID,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var selectMetrics []string
	var selectKeys []string
	var sql string
	var params []interface{}
	logCtx := log.WithFields(logFields)
	transformedQuery, customerAccountID, err := store.transFormRequestFieldsAndFetchRequiredFieldsForFacebook(
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

func (store *MemSQL) transFormRequestFieldsAndFetchRequiredFieldsForFacebook(projectID int64,
	query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, string, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var transformedQuery model.ChannelQueryV1
	var err error
	logCtx := log.WithFields(logFields)
	projectSetting, errCode := store.GetProjectSetting(projectID)
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
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"request_select_metrics": requestSelectMetrics,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"request_filters": requestFilters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

	resultGroupBys := make([]model.ChannelGroupBy, 0, 0)
	for _, requestGroupBy := range requestGroupBys {
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

func buildFacebookQueryV1(query *model.ChannelQueryV1, projectID int64, customerAccountID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, error) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"fetch_source":                  fetchSource,
		"customer_account_id":           customerAccountID,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	lowestHierarchyLevel := getLowestHierarchyLevelForFacebook(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_insights"
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromFacebookReports(query, projectID, query.From, query.To, customerAccountID, facebookDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	return sql, params, selectKeys, selectMetrics, nil
}
func buildFacebookQueryWithSmartPropertyV1(query *model.ChannelQueryV1, projectID int64, customerAccountID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, error) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"fetch_source":                  fetchSource,
		"customer_account_id":           customerAccountID,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	lowestHierarchyLevel := getLowestHierarchyLevelForFacebook(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_insights"
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromFacebookReportsWithSmartProperty(query, projectID, query.From, query.To, customerAccountID, facebookDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	return sql, params, selectKeys, selectMetrics, nil
}

func getSQLAndParamsFromFacebookReportsWithSmartProperty(query *model.ChannelQueryV1, projectID int64, from, to int64, facebookAccountIDs string,
	docType int, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"fetch_source":                  fetchSource,
		"from":                          from,
		"to":                            to,
		"doc_type":                      docType,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
		"facebook_account_ids":          facebookAccountIDs,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	customerAccountIDs := strings.Split(facebookAccountIDs, ",")
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	finalSelectKeys := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	// Group By and select keys
	for _, groupBy := range query.GroupBy {
		_, isPresent := model.SmartPropertyReservedNames[groupBy.Property]
		isSmartProperty := !isPresent
		if isSmartProperty {
			if groupBy.Object == model.AdwordsCampaign {

				value := fmt.Sprintf("JSON_EXTRACT_STRING(campaign.properties, '%s') as campaign_%s", groupBy.Property, groupBy.Property)
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, fmt.Sprintf("campaign_%s", groupBy.Property))

				groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, fmt.Sprintf("campaign_%s", groupBy.Property))
			} else {
				value := fmt.Sprintf("JSON_EXTRACT_STRING(ad_group.properties,'%s') as ad_group_%s", groupBy.Property, groupBy.Property)
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, fmt.Sprintf("ad_group_%s", groupBy.Property))

				groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, fmt.Sprintf("ad_group_%s", groupBy.Property))
			}
		} else {
			key := groupBy.Object + ":" + groupBy.Property
			if groupBy.Object == CAFilterChannel {
				value := fmt.Sprintf("'Facebook Ads' as %s", model.FacebookInternalRepresentationToExternalRepresentation[key])
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, model.FacebookInternalRepresentationToExternalRepresentation[key])
			} else {
				value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInFacebookReportsMapping[key], model.FacebookInternalRepresentationToExternalRepresentation[key])
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, model.FacebookInternalRepresentationToExternalRepresentation[key])
			}

			groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, model.FacebookInternalRepresentationToExternalRepresentation[key])
		}
	}

	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, model.AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
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
	whereConditionForFilters := getFacebookFiltersWhereStatementWithSmartProperty(query.Filters)
	filterStatementForSmartPropertyGroupBy := getNotNullFilterStatementForSmartPropertyGroupBys(query.GroupBy)
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
	logFields := log.Fields{
		"filters":   filters,
		"group_bys": groupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
func getSQLAndParamsFromFacebookReports(query *model.ChannelQueryV1, projectID int64, from, to int64, facebookAccountIDs string,
	docType int, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"fetch_source":                  fetchSource,
		"from":                          from,
		"to":                            to,
		"doc_type":                      docType,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
		"facebook_account_ids":          facebookAccountIDs,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
		if groupBy.Object == CAFilterChannel {
			value := fmt.Sprintf("'Facebook Ads' as %s", model.FacebookInternalRepresentationToExternalRepresentation[key])
			selectKeys = append(selectKeys, value)
			responseSelectKeys = append(responseSelectKeys, model.FacebookInternalRepresentationToExternalRepresentation[key])
		} else {
			value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInFacebookReportsMapping[key], model.FacebookInternalRepresentationToExternalRepresentation[key])
			selectKeys = append(selectKeys, value)
			responseSelectKeys = append(responseSelectKeys, model.FacebookInternalRepresentationToExternalRepresentation[key])
		}
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
	logFields := log.Fields{
		"filters": filters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultStatement := ""
	var filterValue string
	for index, filter := range filters {
		currentFilterStatement := ""
		if filter.LogicalOp == "" {
			filter.LogicalOp = "AND"
		}
		filterOperator := getOp(filter.Condition, "categorical")
		if filter.Condition == model.ContainsOpStr || filter.Condition == model.NotContainsOpStr {
			filterValue = fmt.Sprintf("%s", filter.Value)
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
	logFields := log.Fields{
		"group_by_combinations": groupByCombinations,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	whereConditionForGBT := ""
	params := make([]interface{}, 0)
	filterStringSmartPropertiesCampaign := "campaign.properties"
	filterStringSmartPropertiesAdGroup := "ad_group.properties"
	for _, groupByCombination := range groupByCombinations {
		whereConditionForEachCombination := ""
		for dimension, value := range groupByCombination {
			filterString := ""
			if strings.HasPrefix(dimension, model.CampaignPrefix) {
				key := fmt.Sprintf(`%s:%s`, "campaign", strings.TrimPrefix(dimension, model.CampaignPrefix))
				currentFilterKey, isPresent := objectToValueInFacebookFiltersMappingWithFacebookDocuments[key]
				if isPresent {
					filterString = currentFilterKey
				} else {
					filterString = fmt.Sprintf("JSON_EXTRACT_STRING(%s,'%s')", filterStringSmartPropertiesCampaign, strings.TrimPrefix(dimension, model.CampaignPrefix))
				}
			} else if strings.HasPrefix(dimension, model.AdgroupPrefix) {
				key := fmt.Sprintf(`%s:%s`, "ad_set", strings.TrimPrefix(dimension, model.AdgroupPrefix))
				currentFilterKey, isPresent := objectToValueInFacebookFiltersMappingWithFacebookDocuments[key]
				if isPresent {
					filterString = currentFilterKey
				} else {
					filterString = fmt.Sprintf("JSON_EXTRACT_STRING(%s,'%s')", filterStringSmartPropertiesAdGroup, strings.TrimPrefix(dimension, model.AdgroupPrefix))
				}
			} else {
				key := fmt.Sprintf(`%s:%s`, "ad", strings.TrimPrefix(dimension, model.KeywordPrefix))
				currentFilterKey := objectToValueInFacebookFiltersMappingWithFacebookDocuments[key]
				filterString = currentFilterKey
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

func getFacebookFiltersWhereStatementWithSmartProperty(filters []model.ChannelFilterV1) string {
	logFields := log.Fields{
		"filters": filters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
			filterValue = fmt.Sprintf("%s", filter.Value)
		} else {
			filterValue = filter.Value
		}
		_, isPresent := model.SmartPropertyReservedNames[filter.Property]
		if isPresent {
			currentFilterStatement = fmt.Sprintf("%s %s '%s' ", objectToValueInFacebookFiltersMapping[filter.Object+":"+filter.Property], filterOperator, filterValue)
			if index == 0 {
				resultStatement = " AND " + currentFilterStatement
			} else {
				resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
			}
		} else {
			currentFilterStatement = fmt.Sprintf("JSON_EXTRACT_STRING(%s.properties, '%s') %s '%s'", model.FacebookObjectMapForSmartProperty[filter.Object], filter.Property, filterOperator, filterValue)
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
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
func (store *MemSQL) GetFacebookLastSyncInfo(projectID int64, CustomerAdAccountID string) ([]model.FacebookLastSyncInfo, int) {
	logFields := log.Fields{
		"project_id":             projectID,
		"customer_ad_account_id": CustomerAdAccountID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
		logCtx := log.WithFields(logFields)
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
func (store *MemSQL) ExecuteFacebookChannelQuery(projectID int64,
	query *model.ChannelQuery) (*model.ChannelQueryResult, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || query == nil {
		logCtx.Error("Invalid project_id or query on execute facebook channel query.")
		return nil, http.StatusInternalServerError
	}

	projectSetting, errCode := store.GetProjectSetting(projectID)
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
	result, err := store.GetFacebookChannelResult(projectID, projectSetting.IntFacebookAdAccount, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get facebook query result.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult = result
	if query.Breakdown == "" {
		return queryResult, http.StatusOK
	}

	metricBreakDown, err := store.GetFacebookMetricBreakdown(projectID, projectSetting.IntFacebookAdAccount, query)
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
func (store *MemSQL) GetFacebookMetricBreakdown(projectID int64, customerAccountID string,
	query *model.ChannelQuery) (*model.ChannelBreakdownResult, error) {
	logFields := log.Fields{
		"project_id":          projectID,
		"query":               query,
		"customer_account_id": customerAccountID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	sqlQuery, documentType := getFacebookMetricsQuery(query, true)

	rows, tx, err, reqID := store.ExecQueryWithContext(sqlQuery, []interface{}{projectID, customerAccountID,
		query.From,
		query.To,
		documentType})
	if err != nil {
		logCtx.WithError(err).Error("Failed to build channel query result.")
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows, tx, reqID)
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
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

	logFields := log.Fields{
		"with_breakdown": withBreakdown,
		"query":          query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	documentType := getFacebookDocumentType(query)

	selectColstWithoutAlias := "SUM(JSON_EXTRACT_STRING(value, 'impressions')) as %s , SUM(JSON_EXTRACT_STRING(value, 'clicks')) as %s," +
		" " + "SUM(JSON_EXTRACT_STRING(value, 'spend')) as %s," +
		" " + "SUM(JSON_EXTRACT_STRING(value, 'unique_clicks')) as %s," +
		" " + "SUM(JSON_EXTRACT_STRING(value, 'reach')) as %s, AVG(JSON_EXTRACT_STRING(value, 'frequency')) as %s, " +
		" " + "SUM(JSON_EXTRACT_STRING(value, 'inline_post_engagement')) as %s," +
		" " + "AVG(JSON_EXTRACT_STRING(value, 'cpc')) as %s"

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
			firstValue := "JSON_EXTRACT_STRING(value, '%s_name') as name, "
			firstValue = fmt.Sprintf(firstValue, query.Breakdown)
			selectCols = firstValue + selectCols
			strmntGroupBy = "GROUP BY id, JSON_EXTRACT_STRING(value, '%s_name')"
			strmntGroupBy = fmt.Sprintf(strmntGroupBy, query.Breakdown)
		}
	}

	sqlQuery := "SELECT" + " " + selectCols + " " + "FROM facebook_documents" + " " + strmntWhere +
		" " + strmntGroupBy
	return sqlQuery, documentType
}

// @TODO Kark v0
func (store *MemSQL) GetFacebookChannelResult(projectID int64, customerAccountID string,
	query *model.ChannelQuery) (*model.ChannelQueryResult, error) {
	logFields := log.Fields{
		"project_id":          projectID,
		"query":               query,
		"customer_account_id": customerAccountID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	sqlQuery, documentType := getFacebookMetricsQuery(query, false)

	queryResult := &model.ChannelQueryResult{}
	rows, tx, err, reqID := store.ExecQueryWithContext(sqlQuery, []interface{}{projectID, customerAccountID,
		query.From,
		query.To,
		documentType})
	if err != nil {
		logCtx.WithError(err).Error("Failed to build channel query result.")
		return queryResult, err
	}
	resultHeaders, resultRows, err := U.DBReadRows(rows, tx, reqID)
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

func (store *MemSQL) GetLatestMetaForFacebookForGivenDays(projectID int64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields) {
	logFields := log.Fields{
		"project_id": projectID,
		"days":       days,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	channelDocumentsCampaign := make([]model.ChannelDocumentsWithFields, 0)
	channelDocumentsAdGroup := make([]model.ChannelDocumentsWithFields, 0)

	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		log.Error("Failed to get project settings")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	if projectSetting.IntFacebookAdAccount == "" {
		log.WithField("projectID", projectID).Error("Integration of facebook is not available for this project.")
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
	query := facebookAdGroupMetadataFetchQueryStr
	params := []interface{}{facebookDocumentTypeAlias[facebookAdSet], projectID, from, to,
		customerAccountIDs, facebookDocumentTypeAlias[facebookAdSet], projectID, from, to, customerAccountIDs,
		facebookDocumentTypeAlias[facebookCampaign], projectID, from, to, customerAccountIDs,
		facebookDocumentTypeAlias[facebookCampaign], projectID, from, to, customerAccountIDs}

	rows1, tx1, err, queryID1 := store.ExecQueryWithContext(query, params)
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d ad_group meta for facebook", days)
		log.WithField("error string", err).Error(errString)
		U.CloseReadQuery(rows1, tx1)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	startReadTime1 := time.Now()
	for rows1.Next() {
		currentRecord := model.ChannelDocumentsWithFields{}
		rows1.Scan(&currentRecord.CampaignID, &currentRecord.CampaignName, &currentRecord.AdGroupID, &currentRecord.AdGroupName)
		channelDocumentsAdGroup = append(channelDocumentsAdGroup, currentRecord)
	}
	U.CloseReadQuery(rows1, tx1)
	U.LogReadTimeWithQueryRequestID(startReadTime1, queryID1, &logFields)

	query = facebookCampaignMetadataFetchQueryStr
	params = []interface{}{facebookDocumentTypeAlias[facebookCampaign], projectID, from, to,
		customerAccountIDs, facebookDocumentTypeAlias[facebookCampaign], projectID, from, to, customerAccountIDs}
	rows2, tx2, err, queryID2 := store.ExecQueryWithContext(query, params)
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d campaign meta for facebook", days)
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

func (store *MemSQL) DeleteFacebookIntegration(projectID int64) (int, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
