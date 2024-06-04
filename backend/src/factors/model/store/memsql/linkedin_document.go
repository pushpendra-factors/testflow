package memsql

import (
	"database/sql"
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

/*
Points for reviewers to avoid confusion:

In linkedin there are 4 objects
1. Campaign group
2. Campaign
3. Creative
4. Company

With a similar name there are 8 doc types, the ones with insights suffix contain the report data

For filters we use the metadata to fetch the values and for query data we use report data
*/

var LinkedinDocumentTypeAlias = map[string]int{
	"creative":                1,
	"campaign_group":          2,
	"campaign":                3,
	"creative_insights":       4,
	"campaign_group_insights": 5,
	"campaign_insights":       6,
	"ad_account":              7,
	"member_company_insights": 8,
}

var objectAndPropertyToValueInLinkedinReportsMapping = map[string]string{
	"campaign_group:id":                         "campaign_group_id",
	"creative:id":                               "creative_id",
	"campaign:id":                               "campaign_id",
	"campaign_group:name":                       "JSON_EXTRACT_STRING(value, 'campaign_group_name')",
	"campaign:name":                             "JSON_EXTRACT_STRING(value, 'campaign_name')",
	"member_company_insights:vanity_name":       "JSON_EXTRACT_STRING(value, 'vanityName')",
	"member_company_insights:localized_name":    "JSON_EXTRACT_STRING(value, 'localizedName')",
	"member_company_insights:headquarters":      "JSON_EXTRACT_STRING(value, 'companyHeadquarters')",
	"member_company_insights:domain":            "JSON_EXTRACT_STRING(value, 'localizedWebsite')",
	"member_company_insights:preferred_country": "JSON_EXTRACT_STRING(value, 'preferredCountry')",
}

// TODO check
var linkedinMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions": "SUM(JSON_EXTRACT_STRING(value, 'impressions'))",
	"clicks":      "SUM(JSON_EXTRACT_STRING(value, 'clicks'))",
	"spend":       "SUM(JSON_EXTRACT_STRING(value, 'costInLocalCurrency') * inr_value)",
	"conversions": "SUM(JSON_EXTRACT_STRING(value, 'conversionValueInLocalCurrency'))",
	// "cost_per_click": "average_cost",
	// "conversion_rate": "conversion_rate"
}

var objectToValueInLinkedinFiltersMapping = map[string]string{
	"campaign:name":                             "JSON_EXTRACT_STRING(value, 'campaign_name')",
	"campaign_group:name":                       "JSON_EXTRACT_STRING(value, 'campaign_group_name')",
	"campaign:id":                               "campaign_id",
	"campaign_group:id":                         "campaign_group_id",
	"creative:id":                               "creative_id",
	"member_company_insights:vanity_name":       "JSON_EXTRACT_STRING(value, 'vanityName')",
	"member_company_insights:localized_name":    "JSON_EXTRACT_STRING(value, 'localizedName')",
	"member_company_insights:headquarters":      "JSON_EXTRACT_STRING(value, 'companyHeadquarters')",
	"member_company_insights:domain":            "JSON_EXTRACT_STRING(value, 'localizedWebsite')",
	"member_company_insights:preferred_country": "JSON_EXTRACT_STRING(value, 'preferredCountry')",
}
var objectToValueInLinkedinFiltersMappingWithLinkedinDocuments = map[string]string{
	"campaign:name":       "JSON_EXTRACT_STRING(linkedin_documents.value, 'campaign_name')",
	"campaign_group:name": "JSON_EXTRACT_STRING(linkedin_documents.value, 'campaign_group_name')",
	"campaign:id":         "linkedin_documents.campaign_id",
	"campaign_group:id":   "linkedin_documents.campaign_group_id",
	"creative:id":         "linkedin_documents.creative_id",
}
var linkedinMetricsToOperation = map[string]string{
	"impressions": "sum",
	"clicks":      "sum",
	"spend":       "sum",
	"conversions": "sum",
}

var mapOfTypeToLinkedinJobCTEAlias = map[string]string{
	"campaign":       "campaign_cte",
	"campaign_group": "campaign_group_cte",
}

var errorEmptyLinkedinDocument = errors.New("empty linked document")

const linkedinFilterQueryStr = "SELECT DISTINCT(LCASE(JSON_EXTRACT_STRING(value, ?))) as filter_value FROM linkedin_documents WHERE project_id = ? AND" +
	" " + "customer_ad_account_id in ( ? ) AND type = ? AND JSON_EXTRACT_STRING(value, ?) IS NOT NULL AND timestamp BETWEEN ? AND ? LIMIT 5000"

const fromLinkedinDocuments = " FROM linkedin_documents "

const staticWhereStatementForLinkedin = "WHERE project_id = ? AND customer_ad_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "
const staticWhereStatementForLinkedinWithSmartProperty = "WHERE linkedin_documents.project_id = ? AND linkedin_documents.customer_ad_account_id IN ( ? ) AND linkedin_documents.type = ? AND linkedin_documents.timestamp between ? AND ? "

const linkedinAdGroupMetadataFetchQueryStr = "WITH ad_group as (select ad_group_information.campaign_id_1 as campaign_id, ad_group_information.ad_group_id_1 as ad_group_id, ad_group_information.ad_group_name_1 as ad_group_name " +
	"from ( " +
	"select campaign_group_id as campaign_id_1, campaign_id as ad_group_id_1, JSON_EXTRACT_STRING(value, 'name') as ad_group_name_1, timestamp " +
	"from linkedin_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) " +
	") as ad_group_information " +
	"INNER JOIN " +
	"(select campaign_id as ad_group_id_1, max(timestamp) as timestamp " +
	"from linkedin_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) group by ad_group_id_1 " +
	") as ad_group_latest_timestamp_id " +
	"ON ad_group_information.ad_group_id_1 = ad_group_latest_timestamp_id.ad_group_id_1 AND ad_group_information.timestamp = ad_group_latest_timestamp_id.timestamp), " +

	" campaign as (select campaign_information.campaign_id_1 as campaign_id, campaign_information.campaign_name_1 as campaign_name " +
	"from ( " +
	"select campaign_group_id as campaign_id_1, JSON_EXTRACT_STRING(value, 'name') as campaign_name_1, timestamp " +
	"from linkedin_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) " +
	") as campaign_information " +
	"INNER JOIN " +
	"(select campaign_group_id as campaign_id_1, max(timestamp) as timestamp " +
	"from linkedin_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) group by campaign_id_1 " +
	") as campaign_latest_timestamp_id " +
	"ON campaign_information.campaign_id_1 = campaign_latest_timestamp_id.campaign_id_1 AND campaign_information.timestamp = campaign_latest_timestamp_id.timestamp) " +

	"select campaign.campaign_id, campaign.campaign_name, ad_group.ad_group_id, ad_group.ad_group_name " +
	"from campaign join ad_group on ad_group.campaign_id = campaign.campaign_id"

const linkedinCampaignMetadataFetchQueryStr = "select campaign_information.campaign_id_1 as campaign_id, campaign_information.campaign_name_1 as campaign_name " +
	"from ( " +
	"select campaign_group_id as campaign_id_1, JSON_EXTRACT_STRING(value, 'name') as campaign_name_1, timestamp " +
	"from linkedin_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) " +
	") as campaign_information " +
	"INNER JOIN " +
	"(select campaign_group_id as campaign_id_1, max(timestamp) as timestamp " +
	"from linkedin_documents where type = ? AND project_id = ? AND timestamp between ? AND ? AND customer_ad_account_id IN (?) group by campaign_id_1 " +
	") as campaign_latest_timestamp_id " +
	"ON campaign_information.campaign_id_1 = campaign_latest_timestamp_id.campaign_id_1 AND campaign_information.timestamp = campaign_latest_timestamp_id.timestamp "
const insertLinkedinStr = "INSERT INTO linkedin_documents (project_id,customer_ad_account_id,type,timestamp,id,campaign_group_id,campaign_id,creative_id,value,created_at,updated_at,is_backfilled,sync_status) VALUES "
const campaignGroupInfoFetchStr = "With campaign_timestamp as (Select campaign_group_id as c1, max(timestamp) as t1 from linkedin_documents where project_id = ? " +
	"and customer_ad_account_id = ? and type = 2 and timestamp between ? and ? group by campaign_group_id) select * from linkedin_documents inner join " +
	"campaign_timestamp on c1 = campaign_group_id and t1=timestamp where project_id = ? and customer_ad_account_id = ? and type = 2 and timestamp between ? and ?"
const campaignInfoFetchStr = "With campaign_timestamp as (Select campaign_id as c1, max(timestamp) as t1 from linkedin_documents where project_id = ? " +
	"and customer_ad_account_id = ? and type = 3 and timestamp between ? and ? and (JSON_EXTRACT_STRING(value, 'status') = 'ACTIVE' or JSON_EXTRACT_JSON(value, 'changeAuditStamps', 'lastModified', 'time')>= ?) group by campaign_id) select * from linkedin_documents inner join " +
	"campaign_timestamp on c1 = campaign_id and t1=timestamp where project_id = ? and customer_ad_account_id = ? and type = 3 and timestamp between ? and ?"

// we are not deleting many records. Hence taken the direct approach.
const deleteDocumentQuery = "Delete from linkedin_documents where project_id = ? and customer_ad_account_id = ? and type = ? and timestamp = ?"

func (store *MemSQL) satisfiesLinkedinDocumentForeignConstraints(linkedinDocument model.LinkedinDocument) int {
	logFields := log.Fields{
		"linkedin_document": linkedinDocument,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, errCode := store.GetProject(linkedinDocument.ProjectID)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func (store *MemSQL) satisfiesLinkedinDocumentUniquenessConstraints(linkedinDocument *model.LinkedinDocument) int {
	logFields := log.Fields{
		"linkedin_document": linkedinDocument,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	errCode := store.isLinkedinDocumentExistByPrimaryKey(linkedinDocument)
	if errCode == http.StatusFound {
		return http.StatusConflict
	}
	if errCode == http.StatusNotFound {
		return http.StatusOK
	}
	return errCode
}

// Checks PRIMARY KEY (project_id, customer_ad_account_id, type, timestamp, id)
func (store *MemSQL) isLinkedinDocumentExistByPrimaryKey(document *model.LinkedinDocument) int {
	logFields := log.Fields{
		"document": document,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if document.ProjectID == 0 || document.CustomerAdAccountID == "" || document.Type == 0 ||
		document.Timestamp == 0 || document.ID == "" {

		log.Error("Invalid linkedin document on primary constraint check.")
		return http.StatusBadRequest
	}

	var linkedinDocument model.LinkedinDocument

	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ? AND customer_ad_account_id = ? AND type = ? AND timestamp = ? AND id = ?",
		document.ProjectID, document.CustomerAdAccountID, document.Type, document.Timestamp, document.ID,
	).Select("id").Find(&linkedinDocument).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}

		logCtx.WithError(err).
			Error("Failed getting to check existence linkedin document by primary keys.")
		return http.StatusInternalServerError
	}

	if linkedinDocument.ID == "" {
		logCtx.Error("Invalid id value returned on linkedin document primary key check.")
		return http.StatusInternalServerError
	}

	return http.StatusFound
}

func getLinkedinDocumentTypeAliasByType() map[int]string {

	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range LinkedinDocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

/*
lastSyncInfo {project_id: , ad_account:, doc_type: , doc_type_alias:, last_timestamp:, last_backfill_timestamp:}
lastBackfill timestamp is used in weekly pull and is filled for MEMBER_COMPANY_INSIGHTS only
expected doc types:
ad_account, campaign group meta, campaign meta, camapign group insights, campaign insights, member_company insights (along with last backfill)
*/
func (store *MemSQL) GetLinkedinLastSyncInfo(projectID int64, CustomerAdAccountID string) ([]model.LinkedinLastSyncInfo, int) {
	logFields := log.Fields{
		"project_id":             projectID,
		"customer_ad_account_id": CustomerAdAccountID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	linkedinLastSyncInfos := make([]model.LinkedinLastSyncInfo, 0, 0)

	queryStr := "SELECT project_id, customer_ad_account_id, type as document_type, max(timestamp) as last_timestamp" +
		" FROM linkedin_documents WHERE project_id = ? AND customer_ad_account_id = ?" +
		" GROUP BY project_id, customer_ad_account_id, type "

	rows, err := db.Raw(queryStr, projectID, CustomerAdAccountID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last linkedin documents by type for sync info.")
		return linkedinLastSyncInfos, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var linkedinLastSyncInfo model.LinkedinLastSyncInfo
		if err := db.ScanRows(rows, &linkedinLastSyncInfo); err != nil {
			log.WithError(err).Error("Failed to scan last linkedin documents by type for sync info.")
			return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
		}

		linkedinLastSyncInfos = append(linkedinLastSyncInfos, linkedinLastSyncInfo)
	}
	documentTypeAliasByType := getLinkedinDocumentTypeAliasByType()

	for i := range linkedinLastSyncInfos {
		logCtx := log.WithFields(logFields)
		typeAlias, typeAliasExists := documentTypeAliasByType[linkedinLastSyncInfos[i].DocumentType]
		if !typeAliasExists {
			logCtx.WithField("document_type",
				linkedinLastSyncInfos[i].DocumentType).Error("Invalid document type given. No type alias name.")
			continue
		}

		linkedinLastSyncInfos[i].DocumentTypeAlias = typeAlias
	}

	// backfill part
	currentTime := time.Now()
	timestampBefore45Days, errConv := strconv.ParseInt(currentTime.AddDate(0, 0, -45).Format("20060102"), 10, 64)
	if errConv != nil {
		log.WithError(err).Error("Failed to get timestamp before 45 days")
		return linkedinLastSyncInfos, http.StatusInternalServerError
	}

	backfillTimestampQuery := "SELECT CASE WHEN min(timestamp) is NULL THEN 0 ELSE min(timestamp) END AS " +
		"min_timestamp from linkedin_documents where project_id = ? AND customer_ad_account_id = ? " +
		"and type = 8 and is_backfilled = False and timestamp >= ?"
	var backfillTimestamp interface{}
	rows, err = db.Raw(backfillTimestampQuery, projectID, CustomerAdAccountID, timestampBefore45Days).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get backfill timestamp.")
	}
	for rows.Next() {
		if err := rows.Scan(&backfillTimestamp); err != nil {
			log.WithError(err).Error("Failed to scan backfill timestamp.")
		}
	}

	for i := range linkedinLastSyncInfos {
		logCtx := log.WithFields(logFields)
		if linkedinLastSyncInfos[i].DocumentTypeAlias == "member_company_insights" {
			intBackfillTimestamp, ok := backfillTimestamp.(int64)
			if !ok {
				logCtx.WithField("document_type",
					linkedinLastSyncInfos[i].DocumentType).Error("Failed to convert backfill timestamp to int64")
				continue
			} else {
				linkedinLastSyncInfos[i].LastBackfillTimestamp = intBackfillTimestamp
			}
		}
	}
	return linkedinLastSyncInfos, http.StatusOK
}

/*
When I return getLastSyncInfos
  - We are return array of Objects. But the length of objects should be static.
  - So when we change that, we have to make it as backward compatible.

Had lastSyncInfos been return in following way, it would have been caught
LastSyncInfo {Campaign: 202, Adgroup: 303, CampaignDaily: , Campaignt22: , Campaignt8: }

In this v1 change, we are separating out normal ads data and company enagements data
here we only get last sync info for normal ads data
expected doc type here are: ad_acccount, campaign_group meta, campaign meta, campaign_group insights, campaign insights
*/
func (store *MemSQL) GetLinkedinAdsLastSyncInfoV1(projectID int64, customerAdAccountID string) ([]model.LinkedinLastSyncInfo, int) {
	logFields := log.Fields{
		"project_id":             projectID,
		"customer_ad_account_id": customerAdAccountID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	linkedinLastSyncInfos := make([]model.LinkedinLastSyncInfo, 0)

	queryStr := "SELECT project_id, customer_ad_account_id, type as document_type, max(timestamp) as last_timestamp" +
		" FROM linkedin_documents WHERE project_id = ? AND customer_ad_account_id = ? and type != ?" +
		" GROUP BY project_id, customer_ad_account_id, type "

	rows, err := db.Raw(queryStr, projectID, customerAdAccountID, LinkedinDocumentTypeAlias["member_company_insights"]).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last linkedin documents by type for sync info.")
		return linkedinLastSyncInfos, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var linkedinLastSyncInfo model.LinkedinLastSyncInfo
		if err := db.ScanRows(rows, &linkedinLastSyncInfo); err != nil {
			log.WithError(err).Error("Failed to scan last linkedin documents by type for sync info.")
			return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
		}

		linkedinLastSyncInfos = append(linkedinLastSyncInfos, linkedinLastSyncInfo)
	}
	documentTypeAliasByType := getLinkedinDocumentTypeAliasByType()

	for i := range linkedinLastSyncInfos {
		logCtx := log.WithFields(logFields)
		typeAlias, typeAliasExists := documentTypeAliasByType[linkedinLastSyncInfos[i].DocumentType]
		if !typeAliasExists {
			logCtx.WithField("document_type",
				linkedinLastSyncInfos[i].DocumentType).Error("Invalid document type given. No type alias name.")
			return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
		}

		linkedinLastSyncInfos[i].DocumentTypeAlias = typeAlias
	}

	return linkedinLastSyncInfos, http.StatusOK
}

/*
this is for company engagement
3 types of last sync infos are fetched here
 1. for daily pull, we want to fetch the data from the next day where any type(sync_status=0,1,2) is present
    - 	that's why we return max ingestion timestamp for this
 2. for type1 we get max(timestamp) where sync_status =1
    or 0 in case sync_status is not present
    - We are not worried about timestamp > n days, because here we're only concerned with knowing when the last sync was done
    - All calculations regarding data fetch are done on sync job side
    - it's is possible to have sync_status =2 data, but it is validated on the sync job
 3. for type2 we get max(timestamp) where sync_status =2
    or 0 in case sync_status is not present
    - all the points mentioned above are valid here as well
*/
func (store *MemSQL) GetLinkedinCompanyLastSyncInfoV1(projectID int64, customerAdAccountID string) ([]model.LinkedinLastSyncInfo, int) {
	logFields := log.Fields{
		"project_id":             projectID,
		"customer_ad_account_id": customerAdAccountID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	linkedinLastSyncInfos := make([]model.LinkedinLastSyncInfo, 0)

	queryStr := "SELECT project_id, customer_ad_account_id, type as document_type, max(timestamp) as last_timestamp" +
		" FROM linkedin_documents WHERE project_id = ? AND customer_ad_account_id = ? and type = ?" +
		" GROUP BY project_id, customer_ad_account_id, type "

	rows, err := db.Raw(queryStr, projectID, customerAdAccountID, LinkedinDocumentTypeAlias["member_company_insights"]).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last linkedin documents by type for sync info.")
		return linkedinLastSyncInfos, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var linkedinLastSyncInfo model.LinkedinLastSyncInfo
		if err := db.ScanRows(rows, &linkedinLastSyncInfo); err != nil {
			log.WithError(err).Error("Failed to scan last linkedin documents by type for sync info.")
			return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
		}

		linkedinLastSyncInfos = append(linkedinLastSyncInfos, linkedinLastSyncInfo)
	}
	documentTypeAliasByType := getLinkedinDocumentTypeAliasByType()

	for i := range linkedinLastSyncInfos {
		logCtx := log.WithFields(logFields)
		typeAlias, typeAliasExists := documentTypeAliasByType[linkedinLastSyncInfos[i].DocumentType]
		if !typeAliasExists {
			logCtx.WithField("document_type",
				linkedinLastSyncInfos[i].DocumentType).Error("Invalid document type given. No type alias name.")
			return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
		}

		linkedinLastSyncInfos[i].DocumentTypeAlias = typeAlias
		linkedinLastSyncInfos[i].SyncType = model.CompanySyncJobTypeMap["daily"]
	}

	t8TimestampQuery := "SELECT CASE WHEN max(timestamp) is NULL THEN 0 ELSE max(timestamp) END AS " +
		"max_timestamp from linkedin_documents where project_id = ? AND customer_ad_account_id = ? " +
		"and type = ? and sync_status = 1"
	var t8EndTimestamp interface{}
	rows, err = db.Raw(t8TimestampQuery, projectID, customerAdAccountID, LinkedinDocumentTypeAlias["member_company_insights"]).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get backfill timestamp.")
		return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
	}
	for rows.Next() {
		if err := rows.Scan(&t8EndTimestamp); err != nil {
			log.WithError(err).Error("Failed to scan backfill timestamp.")
			return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
		}
	}
	intT8EndTimestamp, ok := t8EndTimestamp.(int64)
	if !ok {
		log.WithError(err).Error("Failed to convert t22 timestamp into integer.")
		return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
	}
	t8_last_sync_info := model.LinkedinLastSyncInfo{
		ProjectID:           projectID,
		CustomerAdAccountID: customerAdAccountID,
		DocumentTypeAlias:   "member_company_insights",
		SyncType:            model.CompanySyncJobTypeMap["t8"],
		LastTimestamp:       intT8EndTimestamp,
	}
	linkedinLastSyncInfos = append(linkedinLastSyncInfos, t8_last_sync_info)

	t22TimestampQuery := "SELECT CASE WHEN max(timestamp) is NULL THEN 0 ELSE max(timestamp) END AS " +
		"max_timestamp from linkedin_documents where project_id = ? AND customer_ad_account_id = ? " +
		"and type = ? and sync_status = 2"
	var t22EndTimestamp interface{}
	rows, err = db.Raw(t22TimestampQuery, projectID, customerAdAccountID, LinkedinDocumentTypeAlias["member_company_insights"]).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get backfill timestamp.")
		return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
	}
	for rows.Next() {
		if err := rows.Scan(&t22EndTimestamp); err != nil {
			log.WithError(err).Error("Failed to scan backfill timestamp.")
			return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
		}
	}
	intT22EndTimestamp, ok := t22EndTimestamp.(int64)
	if !ok {
		log.WithError(err).Error("Failed to convert t22 timestamp into integer.")
		return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
	}
	t22_last_sync_info := model.LinkedinLastSyncInfo{
		ProjectID:           projectID,
		CustomerAdAccountID: customerAdAccountID,
		DocumentTypeAlias:   "member_company_insights",
		LastTimestamp:       intT22EndTimestamp,
		SyncType:            model.CompanySyncJobTypeMap["t22"],
	}
	linkedinLastSyncInfos = append(linkedinLastSyncInfos, t22_last_sync_info)

	return linkedinLastSyncInfos, http.StatusOK
}

func (store *MemSQL) GetDomainData(projectID string) ([]model.DomainDataResponse, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	domainDatas := make([]model.DomainDataResponse, 0)

	currentTime := time.Now()
	timestampBefore30Days, err := strconv.ParseInt(currentTime.AddDate(0, 0, -30).Format("20060102"), 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to get timestamp before 31 days")
		return domainDatas, http.StatusInternalServerError
	}

	queryStr := "SELECT project_id, id, timestamp, customer_ad_account_id, JSON_EXTRACT_STRING(value, 'companyHeadquarters') as headquarters, " +
		"JSON_EXTRACT_STRING(value, 'localizedWebsite') as domain, JSON_EXTRACT_STRING(value, 'vanityName') as vanity_name, " +
		"JSON_EXTRACT_STRING(value, 'localizedName') as localized_name, JSON_EXTRACT_STRING(value, 'preferredCountry') as preferred_country, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) as impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) as clicks " +
		"FROM linkedin_documents WHERE " +
		"project_id = ? and type = ? and is_backfilled = TRUE and is_group_user_created != TRUE and domain != '$none' and timestamp >= ? " +
		"group by project_id, id, timestamp, customer_ad_account_id, headquarters, domain, vanity_name, localized_name, preferred_country " +
		"order by timestamp ASC, project_id, id, customer_ad_account_id, headquarters, domain, vanity_name, localized_name, preferred_country"

	rows, err := db.Raw(queryStr, projectID, LinkedinDocumentTypeAlias["member_company_insights"], timestampBefore30Days).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last linkedin documents by type for sync info.")
		return domainDatas, http.StatusInternalServerError
	}
	defer rows.Close()
	for rows.Next() {
		var domainData model.DomainDataResponse
		if err := db.ScanRows(rows, &domainData); err != nil {
			log.WithError(err).Error("Failed to scan last linkedin documents by type for sync info.")
			return []model.DomainDataResponse{}, http.StatusInternalServerError
		}

		domainDatas = append(domainDatas, domainData)
	}
	return domainDatas, http.StatusOK
}

func (store *MemSQL) GetDistinctTimestampsForEventCreationFromLinkedinDocs(projectID string) ([]int64, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	currentTime := time.Now()
	timestampBefore45Days, err := strconv.ParseInt(currentTime.AddDate(0, 0, -45).Format("20060102"), 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to get timestamp before 31 days")
		return make([]int64, 0), http.StatusInternalServerError
	}
	distinctTimestampQueryStr := "SELECT distinct(timestamp) from linkedin_documents where project_id = ? " +
		"and type = 8 and is_group_user_created != TRUE and timestamp >= ? order by timestamp asc"
	rows, err := db.Raw(distinctTimestampQueryStr, projectID, timestampBefore45Days).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get distinct timestamps for domain data.")
		return make([]int64, 0), http.StatusInternalServerError
	}
	defer rows.Close()
	arrOfTimestamps := make([]int64, 0)
	for rows.Next() {
		var timestamp int64
		if err := rows.Scan(&timestamp); err != nil {
			log.WithError(err).Error("Failed to scan  distinct timestamps for domain data.")
			return make([]int64, 0), http.StatusInternalServerError
		}
		arrOfTimestamps = append(arrOfTimestamps, timestamp)
	}
	return arrOfTimestamps, http.StatusOK
}

func (store *MemSQL) GetCompanyDataFromLinkedinDocsForTimestamp(projectID string, timestamp int64) ([]model.DomainDataResponse, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	domainDataSet := make([]model.DomainDataResponse, 0)
	projectIDInt, _ := strconv.ParseInt(projectID, 10, 64)

	fetchDomainDataFieldsQueryStr := "SELECT project_id, id, timestamp, customer_ad_account_id, campaign_id, campaign_group_id, JSON_EXTRACT_STRING(value, 'companyHeadquarters') as headquarters, " +
		"JSON_EXTRACT_STRING(value, 'localizedWebsite') as domain, JSON_EXTRACT_STRING(value, 'vanityName') as vanity_name, " +
		"JSON_EXTRACT_STRING(value, 'localizedName') as localized_name, JSON_EXTRACT_STRING(value, 'preferredCountry') as preferred_country, " +
		"JSON_EXTRACT_STRING(value, 'campaign_group_name') as campaign_group_name, JSON_EXTRACT_STRING(value, 'campaign_name') as campaign_name, " +
		"JSON_EXTRACT_STRING(value, 'impressions') as impressions, JSON_EXTRACT_STRING(value, 'clicks') as clicks " +
		"FROM linkedin_documents WHERE " +
		"project_id = ? and type = ? and is_group_user_created != TRUE and timestamp = ? " +
		"group by project_id, id, campaign_group_id, campaign_id, timestamp, customer_ad_account_id, headquarters, domain, vanity_name, localized_name, preferred_country, impressions, clicks " +
		"order by timestamp ASC, project_id, id, customer_ad_account_id, campaign_group_id, campaign_id, headquarters, domain, vanity_name, localized_name, " +
		"campaign_group_name, preferred_country limit 10000"

	rows, err := db.Raw(fetchDomainDataFieldsQueryStr, projectID, LinkedinDocumentTypeAlias["member_company_insights"], timestamp).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get domain data for given timestamp.")
		return make([]model.DomainDataResponse, 0), http.StatusInternalServerError
	}
	defer rows.Close()
	for rows.Next() {
		var domainData model.DomainDataResponse
		if err := db.ScanRows(rows, &domainData); err != nil {
			log.WithError(err).Error("Failed to scan domain data for given timestamp.")
			return make([]model.DomainDataResponse, 0), http.StatusInternalServerError
		}
		if domainData.Domain != "$none" {
			domainData.RawDomain = domainData.Domain
			domainData.Domain = U.GetDomainGroupDomainName(projectIDInt, domainData.Domain)
		}
		domainData.OrgID = domainData.ID // to avoid ID ambiguity
		domainDataSet = append(domainDataSet, domainData)
	}
	return domainDataSet, http.StatusOK
}

func validateLinkedinDocuments(linkedinDocuments []model.LinkedinDocument) int {
	for index, document := range linkedinDocuments {
		if document.CustomerAdAccountID == "" || document.TypeAlias == "" {
			log.WithField("document", document).Error("Invalid linkedin document.")
			return http.StatusBadRequest
		}
		if document.ProjectID == 0 || document.Timestamp == 0 {
			log.WithField("document", document).Error("Invalid linkedin document.")
			return http.StatusBadRequest
		}

		docType, docTypeExists := LinkedinDocumentTypeAlias[document.TypeAlias]
		if !docTypeExists {
			log.WithField("document", document).Error("Invalid linkedin document.")
			return http.StatusBadRequest
		}
		addLinkedinDocType(&linkedinDocuments[index], docType)
	}
	return http.StatusOK
}

// CreatelinkedinDocument ...
func (store *MemSQL) CreateLinkedinDocument(projectID int64, document *model.LinkedinDocument) int {
	logFields := log.Fields{
		"project_id": projectID,
		"document":   document,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if document.CustomerAdAccountID == "" || document.TypeAlias == "" {
		logCtx.Error("Invalid linkedin document.")
		return http.StatusBadRequest
	}
	if document.ProjectID == 0 || document.Timestamp == 0 {
		logCtx.Error("Invalid linkedin document.")
		return http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", document.TypeAlias)
	docType, docTypeExists := LinkedinDocumentTypeAlias[document.TypeAlias]
	if !docTypeExists {
		logCtx.Error("Invalid type alias.")
		return http.StatusBadRequest
	}
	document.Type = docType

	campaignGroupID, campaignID, creativeID, err := getLinkedinHierarchyColumnsByType(docType, document.Value)
	if err != nil {
		logCtx.WithError(err).Error("Invalid docType alias.")
		return http.StatusBadRequest
	}
	document.CampaignGroupID = campaignGroupID
	document.CampaignID = campaignID
	document.CreativeID = creativeID
	if errCode := store.satisfiesLinkedinDocumentForeignConstraints(*document); errCode != http.StatusOK {
		return http.StatusInternalServerError
	}

	errCode := store.satisfiesLinkedinDocumentUniquenessConstraints(document)
	if errCode != http.StatusOK {
		return errCode
	}

	db := C.GetServices().Db
	err = db.Create(&document).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			logCtx.WithError(err).WithField("id", document.ID).Error(
				"Failed to create an linkedin doc. Duplicate.")
			return http.StatusConflict
		}
		logCtx.WithError(err).WithField("id", document.ID).Error(
			"Failed to create an linkedin doc. Continued inserting other docs.")
		return http.StatusInternalServerError
	}
	UpdateCountCacheByDocumentType(projectID, &document.CreatedAt, "linkedin")
	return http.StatusCreated
}

func addColumnInformationForLinkedinDocuments(linkedinDocuments []model.LinkedinDocument) ([]model.LinkedinDocument, int) {
	logFields := log.Fields{"linkedin_documents": linkedinDocuments}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for index, document := range linkedinDocuments {
		campaignGroupID, campaignID, creativeID, err := getLinkedinHierarchyColumnsByType(document.Type, document.Value)
		if err != nil {
			log.WithError(err).WithField("document", linkedinDocuments[index]).Error("Invalid docType alias.")
			return linkedinDocuments, http.StatusBadRequest
		}
		addColumnInformationForLinkedinDocument(&linkedinDocuments[index], campaignGroupID, campaignID, creativeID)
	}
	return linkedinDocuments, http.StatusOK
}
func addColumnInformationForLinkedinDocument(linkedinDocument *model.LinkedinDocument, campaignGroupID string, campaignID string, creativeID string) {
	linkedinDocument.CampaignGroupID = campaignGroupID
	linkedinDocument.CampaignID = campaignID
	linkedinDocument.CreativeID = creativeID
}

func addLinkedinDocType(linkedinDoc *model.LinkedinDocument, docType int) {
	linkedinDoc.Type = docType
}

func (store *MemSQL) DeleteLinkedinDocuments(deletePayload model.LinkedinDeleteDocumentsPayload) int {
	logFields := log.Fields{"delete_payload_linkedin": deletePayload}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	docType, docTypeExists := LinkedinDocumentTypeAlias[deletePayload.TypeAlias]
	if !docTypeExists {
		logCtx.Error("Invalid type alias.")
		return http.StatusBadRequest
	}
	if docType != 8 {
		return http.StatusForbidden
	}
	query := deleteDocumentQuery
	params := []interface{}{deletePayload.ProjectID, deletePayload.CustomerAdAccountID, docType, deletePayload.Timestamp}
	_, _, err := store.ExecuteSQL(query, params, logCtx)
	if err != nil {
		logCtx.WithError(err).Error("Failed to delete data")
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

func (store *MemSQL) GetCampaignGroupInfoForGivenTimerange(campaignGroupInfoRequestPayload model.LinkedinCampaignGroupInfoRequestPayload) ([]model.LinkedinDocument, int) {
	logFields := log.Fields{"campaign_info_request_payload_linkedin": campaignGroupInfoRequestPayload}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	projectID, adAccountID, startTime, endTime := campaignGroupInfoRequestPayload.ProjectID, campaignGroupInfoRequestPayload.CustomerAdAccountID, campaignGroupInfoRequestPayload.StartTimestamp, campaignGroupInfoRequestPayload.EndTimestamp
	linkedinDocuments := make([]model.LinkedinDocument, 0)

	query := campaignGroupInfoFetchStr
	rows, err := db.Raw(query, projectID, adAccountID, startTime, endTime, projectID, adAccountID, startTime, endTime).Rows()

	if err != nil {
		logCtx.WithError(err).Error("Failed to get campaign group info for given time range.")
		return make([]model.LinkedinDocument, 0), http.StatusInternalServerError
	}
	defer rows.Close()
	for rows.Next() {
		var linkedinDocument model.LinkedinDocument
		if err := db.ScanRows(rows, &linkedinDocument); err != nil {
			logCtx.WithError(err).Error("Failed to scan campaign group info for given time range.")
			return make([]model.LinkedinDocument, 0), http.StatusInternalServerError
		}
		linkedinDocuments = append(linkedinDocuments, linkedinDocument)
	}
	return linkedinDocuments, http.StatusOK
}

func (store *MemSQL) GetCampaignInfoForGivenTimerange(campaignInfoRequestPayload model.LinkedinCampaignGroupInfoRequestPayload) ([]model.LinkedinDocument, int) {
	logFields := log.Fields{"campaign_info_request_payload_linkedin": campaignInfoRequestPayload}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	projectID, adAccountID, startTime, endTime := campaignInfoRequestPayload.ProjectID, campaignInfoRequestPayload.CustomerAdAccountID, campaignInfoRequestPayload.StartTimestamp, campaignInfoRequestPayload.EndTimestamp
	unixForStartTime := (U.GetBeginningoftheDayEpochForDateAndTimezone(startTime, "UTC") - (7 * 86400)) * 1000
	linkedinDocuments := make([]model.LinkedinDocument, 0)

	query := campaignInfoFetchStr
	rows, err := db.Raw(query, projectID, adAccountID, startTime, endTime, unixForStartTime, projectID, adAccountID, startTime, endTime).Rows()

	if err != nil {
		logCtx.WithError(err).Error("Failed to get campaign group info for given time range.")
		return make([]model.LinkedinDocument, 0), http.StatusInternalServerError
	}
	defer rows.Close()
	for rows.Next() {
		var linkedinDocument model.LinkedinDocument
		if err := db.ScanRows(rows, &linkedinDocument); err != nil {
			logCtx.WithError(err).Error("Failed to scan campaign group info for given time range.")
			return make([]model.LinkedinDocument, 0), http.StatusInternalServerError
		}
		linkedinDocuments = append(linkedinDocuments, linkedinDocument)
	}
	return linkedinDocuments, http.StatusOK
}

/*
validation logic for a given timerange
 1. for daily job (sync_status=0), the give timerange should not have any data and this is represented by sync_status > -1
 2. for type1/t8 job (sync_status=1), the give timerange should not have type2 job data, the range can contain daily data
    and this is represented by sync_status > 1
 3. for type2/t22 job (sync_status=2), there are not validations req

**note**; sync_state -> 0 -> dailydata, 1 -> t8 data, 2 -> t22 data
*/
func (store *MemSQL) GetValidationForGivenTimerangeAndJobType(validateRequestPayload model.LinkedinValidationRequestPayload) (bool, int) {
	logFields := log.Fields{"validateRequestPayload": validateRequestPayload}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	projectID, adAccountID, startTime, endTime, syncStatus := validateRequestPayload.ProjectID, validateRequestPayload.CustomerAdAccountID, validateRequestPayload.StartTimestamp, validateRequestPayload.EndTimestamp, validateRequestPayload.SyncStatus

	if syncStatus == model.CompanySyncJobTypeMap["t22"] {
		return true, http.StatusOK
	}
	if syncStatus == model.CompanySyncJobTypeMap["daily"] {
		syncStatus = -1 // changing to satisfy the condition explained above
	}
	query := "SELECT count(*) as count from linkedin_documents where project_id = ? AND customer_ad_account_id = ? " +
		"and type = ? and sync_status > ? and timestamp between ? and ?"

	rows, err := db.Raw(query, projectID, adAccountID, LinkedinDocumentTypeAlias["member_company_insights"], syncStatus, startTime, endTime).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to query count of rows.")
		return false, http.StatusInternalServerError
	}
	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			logCtx.WithError(err).Error("Failed to scan count of rows")
			return false, http.StatusInternalServerError
		}
	}

	if count > 0 {
		return false, http.StatusOK
	}

	return true, http.StatusOK
}

// CreateMultipleLinkedinDocument ...
func (store *MemSQL) CreateMultipleLinkedinDocument(linkedinDocuments []model.LinkedinDocument) int {
	logFields := log.Fields{"linkedin_documents": linkedinDocuments}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	status := validateLinkedinDocuments(linkedinDocuments)
	if status != http.StatusOK {
		return status
	}
	linkedinDocuments, status = addColumnInformationForLinkedinDocuments(linkedinDocuments)
	if status != http.StatusOK {
		return status
	}

	db := C.GetServices().Db

	insertStatement := insertLinkedinStr
	insertValuesStatement := make([]string, 0)
	insertValues := make([]interface{}, 0)
	for _, linkedinDoc := range linkedinDocuments {
		insertValuesStatement = append(insertValuesStatement, fmt.Sprintf("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"))
		insertValues = append(insertValues, linkedinDoc.ProjectID, linkedinDoc.CustomerAdAccountID,
			linkedinDoc.Type, linkedinDoc.Timestamp, linkedinDoc.ID, linkedinDoc.CampaignGroupID, linkedinDoc.CampaignID,
			linkedinDoc.CreativeID, linkedinDoc.Value, time.Now(), time.Now(), linkedinDoc.IsBackfilled, linkedinDoc.SyncStatus)
		UpdateCountCacheByDocumentType(linkedinDoc.ProjectID, &linkedinDoc.CreatedAt, "linkedin")
	}
	insertStatement += joinWithComma(insertValuesStatement...)
	rows, err := db.Raw(insertStatement, insertValues...).Rows()

	if err != nil {
		if IsDuplicateRecordError(err) {
			log.WithError(err).WithField("linkedinDocuments", linkedinDocuments).Error("Failed to create an linkedin doc. Duplicate.")
			return http.StatusConflict
		} else {
			log.WithError(err).WithField("linkedinDocuments", linkedinDocuments).Error(
				"Failed to create an linkedin doc. Continued inserting other docs.")
			return http.StatusInternalServerError
		}
	}
	defer rows.Close()

	if status != http.StatusOK {
		return status
	}
	return http.StatusCreated
}
func getLinkedinHierarchyColumnsByType(docType int, valueJSON *postgres.Jsonb) (string, string, string, error) {
	logFields := log.Fields{
		"doc_type":   docType,
		"value_json": valueJSON,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if docType > len(LinkedinDocumentTypeAlias) {
		return "", "", "", errors.New("invalid document type")
	}

	valueMap, err := U.DecodePostgresJsonb(valueJSON)
	if err != nil {
		return "", "", "", err
	}

	if len(*valueMap) == 0 {
		return "", "", "", errorEmptyLinkedinDocument
	}

	return U.GetStringFromMapOfInterface(*valueMap, "campaign_group_id", ""), U.GetStringFromMapOfInterface(*valueMap, "campaign_id", ""), U.GetStringFromMapOfInterface(*valueMap, "creative_id", ""), nil
}

func (store *MemSQL) IsLinkedInIntegrationAvailable(projectID int64) bool {
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return false
	}

	if projectSetting.IntLinkedinAdAccount == "" {
		return false
	}
	return true
}

func (store *MemSQL) UpdateSyncStatusLinkedinDocs(domainData model.DomainDataResponse) error {
	db := C.GetServices().Db
	var err error
	if domainData.CampaignGroupID == "" {
		err = db.Table("linkedin_documents").Where("project_id = ? and customer_ad_account_id = ? and timestamp = ? and id = ? and type = 8", domainData.ProjectID, domainData.CustomerAdAccountID, domainData.Timestamp, domainData.ID).Updates(map[string]interface{}{"is_group_user_created": true, "updated_at": time.Now()}).Error
	} else {
		err = db.Table("linkedin_documents").Where("project_id = ? and customer_ad_account_id = ? and timestamp = ? and id = ? and campaign_group_id = ? and type = 8", domainData.ProjectID, domainData.CustomerAdAccountID, domainData.Timestamp, domainData.ID, domainData.CampaignGroupID).Updates(map[string]interface{}{"is_group_user_created": true, "updated_at": time.Now()}).Error
	}
	if err != nil {
		return err
	}
	return nil
}

// v1 Api
func (store *MemSQL) buildLinkedinChannelConfig(projectID int64) *model.ChannelConfigResult {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	linkedinObjectsAndProperties := store.buildObjectAndPropertiesForLinkedin(projectID, model.ObjectsForLinkedin)
	objectsAndProperties := append(linkedinObjectsAndProperties)

	return &model.ChannelConfigResult{
		SelectMetrics:        SelectableMetricsForAllChannels,
		ObjectsAndProperties: objectsAndProperties,
	}
}

func (store *MemSQL) buildObjectAndPropertiesForLinkedin(projectID int64, objects []string) []model.ChannelObjectAndProperties {
	logFields := log.Fields{
		"project_id": projectID,
		"objects":    objects,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0, 0)
	for _, currentObject := range objects {
		var currentProperties []model.ChannelProperty
		var currentPropertiesSmart []model.ChannelProperty
		currentProperties = buildProperties(allChannelsPropertyToRelated)
		smartProperty := store.GetSmartPropertyAndRelated(projectID, currentObject, "linkedin")
		currentPropertiesSmart = buildProperties(smartProperty)
		currentProperties = append(currentProperties, currentPropertiesSmart...)
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
}

/*
This func uses metadata to get the values, as paused camapigns and deleted campaigns won't show up in report data and
if someone wants to look at old data it'll cause issues
*/
func (store *MemSQL) GetLinkedinFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, isPresent := model.SmartPropertyReservedNames[requestFilterProperty]
	if !isPresent {
		filterValues, errCode := store.getSmartPropertyFilterValues(projectID, requestFilterObject, requestFilterProperty, "linkedin", reqID)
		if errCode != http.StatusFound {
			return []interface{}{}, http.StatusInternalServerError
		}
		return filterValues, http.StatusFound
	}
	linkedinInternalFilterProperty, docType, err := getFilterRelatedInformationForLinkedin(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return make([]interface{}, 0, 0), http.StatusBadRequest
	}

	filterValues, errCode := store.getLinkedinFilterValuesByType(projectID, docType, linkedinInternalFilterProperty, reqID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

func getFilterRelatedInformationForLinkedin(requestFilterObject string, requestFilterProperty string) (string, int, int) {
	logFields := log.Fields{
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	linkedinInternalFilterObject, isPresent := model.LinkedinExternalRepresentationToInternalRepresentation[requestFilterObject]
	if !isPresent {
		log.Error("Invalid linkedin filter object.")
		return "", 0, http.StatusBadRequest
	}
	linkedinInternalFilterProperty, isPresent := model.LinkedinExternalRepresentationToInternalRepresentation[requestFilterProperty]
	if !isPresent {
		log.Error("Invalid linkedin filter property.")
		return "", 0, http.StatusBadRequest
	}
	docType := LinkedinDocumentTypeAlias[linkedinInternalFilterObject]

	return linkedinInternalFilterProperty, docType, http.StatusOK
}

func (store *MemSQL) getLinkedinFilterValuesByType(projectID int64, docType int, property string, reqID string) ([]interface{}, int) {
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
		logCtx.WithField("err_code", errCode).Error("Failed to fetch Project Setting in linkedin filter values.")
		return []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntLinkedinAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		logCtx.Info(integrationNotAvailable)
		return []interface{}{}, http.StatusNotFound
	}
	customerAccountIDs := strings.Split(customerAccountID, ",")
	logCtx = log.WithField("project_id", projectID).WithField("doc_type", docType).WithField("req_id", reqID)
	from, to := model.GetFromAndToDatesForFilterValues()
	params := []interface{}{property, projectID, customerAccountIDs, docType, property, from, to}
	_, resultRows, err := store.ExecuteSQL(linkedinFilterQueryStr, params, logCtx)
	if err != nil {
		logCtx.WithError(err).WithField("query", linkedinFilterQueryStr).WithField("params", params).Error(model.LinkedinSpecificError)
		return make([]interface{}, 0), http.StatusInternalServerError
	}

	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

func (store *MemSQL) GetLinkedinSQLQueryAndParametersForFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	linkedinInternalFilterProperty, docType, err := getFilterRelatedInformationForLinkedin(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return "", make([]interface{}, 0, 0), http.StatusBadRequest
	}
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("failed to fetch Project Setting in linkedin filter values.")
		return "", make([]interface{}, 0, 0), http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntLinkedinAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		logCtx.Info(integrationNotAvailable)
		return "", nil, http.StatusNotFound
	}
	customerAccountIDs := strings.Split(customerAccountID, ",")
	from, to := model.GetFromAndToDatesForFilterValues()
	params := []interface{}{linkedinInternalFilterProperty, projectID, customerAccountIDs, docType, linkedinInternalFilterProperty, from, to}

	return "(" + linkedinFilterQueryStr + ")", params, http.StatusFound
}

func (store *MemSQL) ExecuteLinkedinChannelQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	fetchSource := false
	logCtx := log.WithFields(logFields)
	limitString := fmt.Sprintf(" LIMIT %d", model.ChannelsLimit)
	if query.GroupByTimestamp == "" {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForLinkedinQueryV1(projectID,
			query, reqID, fetchSource, limitString, false, nil)
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
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.LinkedinSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	} else {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForLinkedinQueryV1(
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
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.LinkedinSpecificError)
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		groupByCombinations := model.GetGroupByCombinationsForChannelAnalytics(columns, resultMetrics)
		sql, params, selectKeys, selectMetrics, errCode = store.GetSQLQueryAndParametersForLinkedinQueryV1(
			projectID, query, reqID, fetchSource, limitString, true, groupByCombinations)
		if errCode != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err = store.ExecuteSQL(sql, params, logCtx)
		columns = append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.LinkedinSpecificError)
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	}
}

func (store *MemSQL) GetSQLQueryAndParametersForLinkedinQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string, int) {
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
	var sql string
	var selectKeys []string
	var params []interface{}
	logCtx := log.WithFields(logFields)
	transformedQuery, customerAccountID, projectCurrency, err := store.transFormRequestFieldsAndFetchRequiredFieldsForLinkedin(projectID, *query, reqID)
	if err != nil && err.Error() == integrationNotAvailable {
		logCtx.WithError(err).Info(model.LinkedinSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusNotFound
	}
	if err != nil {
		logCtx.WithError(err).Error(model.LinkedinSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusBadRequest
	}
	isSmartPropertyPresent := checkSmartProperty(query.Filters, query.GroupBy)
	dataCurrency := ""
	if projectCurrency != "" {
		dataCurrency = store.GetDataCurrencyForLinkedin(projectID)
	}
	if isSmartPropertyPresent {
		sql, params, selectKeys, selectMetrics, err = buildLinkedinQueryWithSmartPropertyV1(transformedQuery, projectID, customerAccountID, fetchSource,
			limitString, isGroupByTimestamp, groupByCombinationsForGBT, dataCurrency, projectCurrency)
		if err != nil {
			return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
		}
		return sql, params, selectKeys, selectMetrics, http.StatusOK
	}

	sql, params, selectKeys, selectMetrics, err = buildLinkedinQueryV1(transformedQuery, projectID, customerAccountID, fetchSource,
		limitString, isGroupByTimestamp, groupByCombinationsForGBT, dataCurrency, projectCurrency)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
	}
	return sql, params, selectKeys, selectMetrics, http.StatusOK
}

func (store *MemSQL) transFormRequestFieldsAndFetchRequiredFieldsForLinkedin(projectID int64, query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, string, string, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	query.From = U.GetDateAsStringIn(query.From, U.TimeZoneString(query.Timezone))
	query.To = U.GetDateAsStringIn(query.To, U.TimeZoneString(query.Timezone))
	var err error
	logCtx := log.WithFields(logFields)
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return &model.ChannelQueryV1{}, "", "", errors.New("Project setting not found")
	}
	customerAccountID := projectSetting.IntLinkedinAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		return &model.ChannelQueryV1{}, "", "", errors.New(integrationNotAvailable)
	}

	query, err = convertFromRequestToLinkedinSpecificRepresentation(query)
	if err != nil {
		logCtx.Warn("Request failed in validation: ", err)
		return &model.ChannelQueryV1{}, "", "", err
	}
	return &query, customerAccountID, projectSetting.ProjectCurrency, nil
}

func (store *MemSQL) GetDataCurrencyForLinkedin(projectId int64) string {
	query := "select JSON_EXTRACT_STRING(value, 'currency') from linkedin_documents where project_id = ? and type = 7 order by created_at desc limit 1"
	db := C.GetServices().Db

	params := make([]interface{}, 0)
	params = append(params, projectId)
	rows, err := db.Raw(query, params).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get currency code.")
	}
	defer rows.Close()

	var currency string
	for rows.Next() {

		if err := rows.Scan(&currency); err != nil {
			log.WithError(err).Error("Failed to get currency details for linkedin")
		}
	}

	return currency
}

func convertFromRequestToLinkedinSpecificRepresentation(query model.ChannelQueryV1) (model.ChannelQueryV1, error) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var err1, err2, err3 error
	query.SelectMetrics, err1 = getLinkedinSpecificMetrics(query.SelectMetrics)
	query.Filters, err2 = getLinkedinSpecificFilters(query.Filters)
	query.GroupBy, err3 = getLinkedinSpecificGroupBy(query.GroupBy)
	if err1 != nil {
		return query, err1
	}
	if err2 != nil {
		return query, err2
	}
	if err3 != nil {
		return query, err3
	}
	return query, nil
}

// @Kark TODO v1
func getLinkedinSpecificMetrics(requestSelectMetrics []string) ([]string, error) {
	logFields := log.Fields{
		"request_select_metrics": requestSelectMetrics,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultMetrics := make([]string, 0, 0)
	for _, requestMetric := range requestSelectMetrics {
		metric, isPresent := model.LinkedinExternalRepresentationToInternalRepresentation[requestMetric]
		if !isPresent {
			return make([]string, 0, 0), errors.New("Invalid metric found for document type")
		}
		resultMetrics = append(resultMetrics, metric)
	}
	return resultMetrics, nil
}

// @Kark TODO v1
func getLinkedinSpecificFilters(requestFilters []model.ChannelFilterV1) ([]model.ChannelFilterV1, error) {
	logFields := log.Fields{
		"request_filters": requestFilters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	filters := make([]model.ChannelFilterV1, 0)
	for _, requestFilter := range requestFilters {
		filterObject, isPresent := model.LinkedinExternalRepresentationToInternalRepresentation[requestFilter.Object]
		if !isPresent {
			return make([]model.ChannelFilterV1, 0, 0), errors.New("Invalid filter key found for document type")

		}
		filters = append(filters, model.ChannelFilterV1{Object: filterObject, Property: requestFilter.Property, Condition: requestFilter.Condition,
			Value: requestFilter.Value, LogicalOp: requestFilter.LogicalOp})
	}
	return filters, nil
}

// @Kark TODO v1
func getLinkedinSpecificGroupBy(requestGroupBys []model.ChannelGroupBy) ([]model.ChannelGroupBy, error) {

	resultGroupBys := make([]model.ChannelGroupBy, 0, 0)
	for _, requestGroupBy := range requestGroupBys {
		var resultGroupBy model.ChannelGroupBy
		groupByObject, isPresent := model.LinkedinExternalRepresentationToInternalRepresentation[requestGroupBy.Object]
		if !isPresent {
			return make([]model.ChannelGroupBy, 0, 0), errors.New("Invalid groupby key found for document type")
		}
		resultGroupBy = requestGroupBy
		resultGroupBy.Object = groupByObject
		resultGroupBys = append(resultGroupBys, resultGroupBy)
	}
	return resultGroupBys, nil
}

func buildLinkedinQueryWithSmartPropertyV1(query *model.ChannelQueryV1, projectID int64, customerAccountID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}, dataCurrency string, projectCurrency string) (string, []interface{}, []string, []string, error) {
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
	lowestHierarchyLevel := getLowestHierarchyLevelForLinkedin(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_insights"
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromLinkedinWithSmartPropertyReports(query, projectID, query.From, query.To, customerAccountID, LinkedinDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT, dataCurrency, projectCurrency)
	return sql, params, selectKeys, selectMetrics, nil
}
func buildLinkedinQueryV1(query *model.ChannelQueryV1, projectID int64, customerAccountID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}, dataCurrency string, projectCurrency string) (string, []interface{}, []string, []string, error) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"customer_account_id":           customerAccountID,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	lowestHierarchyLevel := getLowestHierarchyLevelForLinkedin(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_insights"
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromLinkedinReports(query, projectID, query.From, query.To, customerAccountID, LinkedinDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT, dataCurrency, projectCurrency)
	return sql, params, selectKeys, selectMetrics, nil
}

// Added case when statement for NULL value and empty value for group bys
// Added case when statement for NULL value for smart properties. Didn't add for empty values as such case will not be present
func getSQLAndParamsFromLinkedinWithSmartPropertyReports(query *model.ChannelQueryV1, projectID int64, from, to int64, linkedinAccountIDs string, docType int,
	fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}, dataCurrency string, projectCurrency string) (string, []interface{}, []string, []string) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"from":                          from,
		"to":                            to,
		"linkedin_account_ids":          linkedinAccountIDs,
		"doc_type":                      docType,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	customerAccountIDs := strings.Split(linkedinAccountIDs, ",")
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
			if groupBy.Object == "campaign_group" {
				value := fmt.Sprintf("Case when JSON_EXTRACT_STRING(campaign.properties, '%s') is null then '$none' else JSON_EXTRACT_STRING(campaign.properties, '%s') END as campaign_%s", groupBy.Property, groupBy.Property, groupBy.Property)
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, fmt.Sprintf("campaign_%s", groupBy.Property))

				groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, fmt.Sprintf("campaign_%s", groupBy.Property))
			} else {
				value := fmt.Sprintf("Case when JSON_EXTRACT_STRING(ad_group.properties,'%s') is null then '$none' else JSON_EXTRACT_STRING(ad_group.properties,'%s') END as ad_group_%s", groupBy.Property, groupBy.Property, groupBy.Property)
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, fmt.Sprintf("ad_group_%s", groupBy.Property))

				groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, fmt.Sprintf("ad_group_%s", groupBy.Property))
			}
		} else {
			key := groupBy.Object + ":" + groupBy.Property
			if groupBy.Object == CAFilterChannel {
				value := fmt.Sprintf("'LinkedIn Ads' as %s", model.LinkedinInternalRepresentationToExternalRepresentation[key])
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, model.LinkedinInternalRepresentationToExternalRepresentation[key])
			} else {
				value := fmt.Sprintf("CASE WHEN %s IS NULL THEN '$none' WHEN %s = '' THEN '$none' ELSE %s END as %s", objectAndPropertyToValueInLinkedinReportsMapping[key], objectAndPropertyToValueInLinkedinReportsMapping[key], objectAndPropertyToValueInLinkedinReportsMapping[key], model.LinkedinInternalRepresentationToExternalRepresentation[key])
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, model.LinkedinInternalRepresentationToExternalRepresentation[key])
			}

			groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, model.LinkedinInternalGroupByRepresentation[key])
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
		value := fmt.Sprintf("%s as %s", linkedinMetricsToAggregatesInReportsMapping[selectMetric], model.LinkedinInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = model.LinkedinInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters, filterParams, err := getFilterPropertiesForLinkedinReportsNew(query.Filters)
	if err != nil {
		return "", nil, nil, nil
	}

	finalFilterStatement := joinWithWordInBetween("AND", staticWhereStatementForLinkedinWithSmartProperty, whereConditionForFilters)

	fromStatement := getLinkedinFromStatementWithJoins(query.Filters, query.GroupBy)
	finalParams := make([]interface{}, 0)
	if (dataCurrency != "" && projectCurrency != "") && U.ContainsStringInArray(query.SelectMetrics, "spend") {
		finalParams = append(finalParams, projectCurrency, dataCurrency)
	}
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, filterParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForLinkedin(groupByCombinationsForGBT)
		finalFilterStatement += " AND (" + whereConditionForGBT + ") "
		finalParams = append(finalParams, whereParams...)
	}
	resultSQLStatement := ""
	if (dataCurrency != "" && projectCurrency != "") && U.ContainsStringInArray(query.SelectMetrics, "spend") {
		resultSQLStatement = selectQuery + fromStatement + currencyQuery + finalFilterStatement
	} else {
		selectQuery = strings.Replace(selectQuery, "* inr_value", "", -1)
		resultSQLStatement = selectQuery + fromStatement + finalFilterStatement
	}
	if len(groupByStatement) != 0 {
		resultSQLStatement += " GROUP BY " + groupByStatement
	}

	resultSQLStatement += " " + orderByQuery + limitString + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

func getLinkedinFromStatementWithJoins(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) string {
	logFields := log.Fields{
		"filters":   filters,
		"group_bys": groupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	isPresentCampaignSmartProperty, isPresentAdGroupSmartProperty := checkSmartPropertyWithTypeAndSource(filters, groupBys, "linkedin")
	fromStatement := fromLinkedinDocuments
	if isPresentAdGroupSmartProperty {
		fromStatement += "left join smart_properties ad_group on ad_group.project_id = linkedin_documents.project_id and ad_group.object_id = campaign_id "
	}
	if isPresentCampaignSmartProperty {
		fromStatement += "left join smart_properties campaign on campaign.project_id = linkedin_documents.project_id and campaign.object_id = campaign_group_id "
	}
	return fromStatement
}

// Added case when statement for NULL value and empty value for group bys
func getSQLAndParamsFromLinkedinReports(query *model.ChannelQueryV1, projectID int64, from, to int64, linkedinAccountIDs string, docType int,
	fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}, dataCurrency string, projectCurrency string) (string, []interface{}, []string, []string) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"from":                          from,
		"to":                            to,
		"linkedin_account_ids":          linkedinAccountIDs,
		"doc_type":                      docType,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	customerAccountIDs := strings.Split(linkedinAccountIDs, ",")
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
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, model.LinkedinInternalGroupByRepresentation[key])
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
			value := fmt.Sprintf("'LinkedIn Ads' as %s", model.LinkedinInternalRepresentationToExternalRepresentation[key])
			selectKeys = append(selectKeys, value)
			responseSelectKeys = append(responseSelectKeys, model.LinkedinInternalRepresentationToExternalRepresentation[key])
		} else {
			value := fmt.Sprintf("CASE WHEN %s IS NULL THEN '$none' WHEN %s = '' THEN '$none' ELSE %s END as %s", objectAndPropertyToValueInLinkedinReportsMapping[key], objectAndPropertyToValueInLinkedinReportsMapping[key], objectAndPropertyToValueInLinkedinReportsMapping[key], model.LinkedinInternalRepresentationToExternalRepresentation[key])
			selectKeys = append(selectKeys, value)
			responseSelectKeys = append(responseSelectKeys, model.LinkedinInternalRepresentationToExternalRepresentation[key])
		}
	}

	finalSelectKeys = append(finalSelectKeys, selectKeys...)
	if isGroupByTimestamp {
		finalSelectKeys = append(finalSelectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}

	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", linkedinMetricsToAggregatesInReportsMapping[selectMetric], model.LinkedinInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = model.LinkedinInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters, filterParams, err := getFilterPropertiesForLinkedinReportsNew(query.Filters)
	if err != nil {
		return "", nil, nil, nil
	}
	if whereConditionForFilters != "" {
		whereConditionForFilters = " AND " + whereConditionForFilters
	}
	finalFilterStatement := whereConditionForFilters
	finalParams := make([]interface{}, 0)
	if (dataCurrency != "" && projectCurrency != "") && U.ContainsStringInArray(query.SelectMetrics, "spend") {
		finalParams = append(finalParams, projectCurrency, dataCurrency)
	}
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, filterParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForLinkedin(groupByCombinationsForGBT)
		finalFilterStatement += " AND (" + whereConditionForGBT + ") "
		finalParams = append(finalParams, whereParams...)
	}

	resultSQLStatement := ""
	if (dataCurrency != "" && projectCurrency != "") && U.ContainsStringInArray(query.SelectMetrics, "spend") {
		resultSQLStatement = selectQuery + fromLinkedinDocuments + currencyQuery + staticWhereStatementForLinkedin + finalFilterStatement
	} else {
		selectQuery = strings.Replace(selectQuery, "* inr_value", "", -1)
		resultSQLStatement = selectQuery + fromLinkedinDocuments + staticWhereStatementForLinkedin + finalFilterStatement
	}

	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + limitString + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}
func buildWhereConditionForGBTForLinkedin(groupByCombinations map[string][]interface{}) (string, []interface{}) {
	logFields := log.Fields{
		"group_by_combinations": groupByCombinations,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultantWhereCondition := ""
	resultantInClauses := make([]string, 0)
	params := make([]interface{}, 0)

	for dimension, values := range groupByCombinations {
		currentInClause := ""

		jsonExtractExpression := GetFilterObjectExpressionForChannelLinkedin(dimension)

		valuesInString := make([]string, 0)
		for _, value := range values {
			valuesInString = append(valuesInString, "?")
			params = append(params, value)
		}
		currentInClause = joinWithComma(valuesInString...)

		resultantInClauses = append(resultantInClauses, jsonExtractExpression+" IN ("+currentInClause+") ")
	}
	resultantWhereCondition = joinWithWordInBetween("AND", resultantInClauses...)

	return resultantWhereCondition, params
}

func GetFilterObjectExpressionForChannelLinkedin(dimension string) string {
	filterObjectForSmartPropertiesCampaign := "campaign.properties"
	filterObjectForSmartPropertiesAdGroup := "ad_group.properties"

	filterExpression := ""
	isNotSmartProperty := false
	if strings.HasPrefix(dimension, model.CampaignPrefix) {
		filterExpression, isNotSmartProperty = GetReportExpressionIfPresentForLinkedin("campaign_group", dimension, model.CampaignPrefix)
		if !isNotSmartProperty {
			filterExpression = fmt.Sprintf("JSON_EXTRACT_STRING(%s,'%s')", filterObjectForSmartPropertiesCampaign, strings.TrimPrefix(dimension, model.CampaignPrefix))
		}
	} else if strings.HasPrefix(dimension, model.AdgroupPrefix) {
		filterExpression, isNotSmartProperty = GetReportExpressionIfPresentForLinkedin("campaign", dimension, model.AdgroupPrefix)
		if !isNotSmartProperty {
			filterExpression = fmt.Sprintf("JSON_EXTRACT_STRING(%s,'%s')", filterObjectForSmartPropertiesAdGroup, strings.TrimPrefix(dimension, model.AdgroupPrefix))
		}
	} else {
		filterExpression, _ = GetReportExpressionIfPresentForLinkedin("creative", dimension, model.KeywordPrefix)
	}
	return filterExpression
}

// Input: objectType - campaign, dimension - , prefix - . TODO
func GetReportExpressionIfPresentForLinkedin(objectType, dimension, prefix string) (string, bool) {
	key := fmt.Sprintf(`%s:%s`, objectType, strings.TrimPrefix(dimension, prefix))
	reportProperty, isPresent := objectToValueInLinkedinFiltersMappingWithLinkedinDocuments[key]
	return reportProperty, isPresent
}

func getFilterPropertiesForLinkedinReportsNew(filters []model.ChannelFilterV1) (rStmnt string, rParams []interface{}, err error) {
	logFields := log.Fields{
		"filters": filters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	campaignFilter := ""
	adGroupFilter := ""
	filtersLen := len(filters)
	if filtersLen == 0 {
		return rStmnt, rParams, nil
	}

	rParams = make([]interface{}, 0)
	groupedProperties := model.GetChannelFiltersGrouped(filters)

	for indexOfGroup, currentGroupedProperties := range groupedProperties {
		var currentGroupStmnt, pStmnt string
		for indexOfProperty, p := range currentGroupedProperties {

			if p.LogicalOp == "" {
				p.LogicalOp = "AND"
			}

			if !isValidLogicalOp(p.LogicalOp) {
				return rStmnt, rParams, errors.New("invalid logical op on where condition")
			}
			pStmnt = ""
			propertyOp := getOp(p.Condition, "categorical")
			// categorical property type.
			pValue := ""
			if p.Condition == model.ContainsOpStr || p.Condition == model.NotContainsOpStr {
				pValue = fmt.Sprintf("%s", p.Value)
			} else {
				pValue = p.Value
			}
			_, isPresent := model.SmartPropertyReservedNames[p.Property]
			if isPresent {
				key := fmt.Sprintf("%s:%s", p.Object, p.Property)
				pFilter := objectToValueInLinkedinFiltersMapping[key]

				if p.Value != model.PropertyValueNone {
					pStmnt = fmt.Sprintf("%s %s '%s' ", pFilter, propertyOp, pValue)
				} else {
					// where condition for $none value.
					if propertyOp == model.EqualsOp || propertyOp == model.RLikeOp {
						pStmnt = fmt.Sprintf("(%s IS NULL OR %s = '')", pFilter, pFilter)
					} else if propertyOp == model.NotEqualOp || propertyOp == model.NotRLikeOp {
						pStmnt = fmt.Sprintf("(%s IS NOT NULL OR %s != '')", pFilter, pFilter)
					} else {
						return "", nil, fmt.Errorf("unsupported opertator %s for property value none", propertyOp)
					}
				}
			} else {
				if p.Value != model.PropertyValueNone {
					pStmnt = fmt.Sprintf("JSON_EXTRACT_STRING(%s.properties, '%s') %s ?", model.LinkedinObjectMapForSmartProperty[p.Object], p.Property, propertyOp)
					rParams = append(rParams, pValue)
				} else {
					if propertyOp == model.EqualsOp || propertyOp == model.RLikeOp {
						pStmnt = fmt.Sprintf("(JSON_EXTRACT_STRING(%s.properties, '%s') IS NULL OR JSON_EXTRACT_STRING(%s.properties, '%s') = '')", model.LinkedinObjectMapForSmartProperty[p.Object], p.Property, model.LinkedinObjectMapForSmartProperty[p.Object], p.Property)
					} else if propertyOp == model.NotEqualOp || propertyOp == model.NotRLikeOp {
						pStmnt = fmt.Sprintf("(JSON_EXTRACT_STRING(%s.properties, '%s') IS NOT NULL AND JSON_EXTRACT_STRING(%s.properties, '%s') != '')", model.LinkedinObjectMapForSmartProperty[p.Object], p.Property, model.LinkedinObjectMapForSmartProperty[p.Object], p.Property)
					} else {
						return "", nil, fmt.Errorf("unsupported opertator %s for property value none", propertyOp)
					}
				}
				if p.Object == "campaign_group" {
					campaignFilter = smartPropertyCampaignStaticFilter
				} else {
					adGroupFilter = smartPropertyAdGroupStaticFilter
				}
			}
			if indexOfProperty == 0 {
				currentGroupStmnt = pStmnt
			} else {
				currentGroupStmnt = fmt.Sprintf("%s %s %s", currentGroupStmnt, p.LogicalOp, pStmnt)
			}
		}
		if indexOfGroup == 0 {
			rStmnt = fmt.Sprintf("(%s)", currentGroupStmnt)
		} else {
			rStmnt = fmt.Sprintf("%s AND (%s)", rStmnt, currentGroupStmnt)
		}

	}
	if campaignFilter != "" {
		rStmnt += (" AND " + campaignFilter)
	}
	if adGroupFilter != "" {
		rStmnt += (" AND " + adGroupFilter)
	}
	return rStmnt, rParams, nil
}

func getNotNullFilterStatementForSmartPropertyLinkedinGroupBys(groupBys []model.ChannelGroupBy) string {
	logFields := log.Fields{
		"group_bys": groupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultStatement := ""
	for _, groupBy := range groupBys {
		_, isPresent := model.SmartPropertyReservedNames[groupBy.Property]
		isSmartProperty := !isPresent
		if isSmartProperty {
			if groupBy.Object == model.LinkedinCampaignGroup {
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

/*
campaign group, campaign, creative are in a heirarchy but each level contains the data of the higher heirarchy.
E.g campaign_insights would contain the data of campaign group
If there's a combination of filters at a different heirarchy,
we need to decide what is the lowest level of heirarchy, that'll have all the data we require.

In contrast to the above, company enagements data is stored at a single level
It contains both company and campaign group data. Hence if a query is fired for company enagements data,
the following function should point to member company insights

ObjectType is used to define which record type the query is supposed to be executed.
    - if Campaign is there, we would earlier fetch from "campaign_insights". But with company engagements, we might need to query other record type.
    "member_company_insights", this is decided with query.Channel property of query object. We couldn't add more categories as it would involve a lot of changes.
	Using quuery.Channel property is a hacky but an accurate solution. It reduces the number of changes and keeps the complications to a minimum.

The reason we have avoided directly associating object type to record type is avoid SQL query complications
*/

func getLowestHierarchyLevelForLinkedin(query *model.ChannelQueryV1) string {
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

	// pointing to member_company_insights incase of company enagements query
	if query.Channel == model.LinkedinCompanyEngagementsDisplayCategory {
		return model.LinkedInMemberCompany
	}

	// Check if present
	for _, objectName := range objectNames {
		if objectName == model.LinkedinCreative {
			return model.LinkedinCreative
		}
	}

	for _, objectName := range objectNames {
		if objectName == model.LinkedinCampaign {
			return model.LinkedinCampaign
		}
	}

	for _, objectName := range objectNames {
		if objectName == model.LinkedinCampaignGroup {
			return model.LinkedinCampaignGroup
		}
	}

	return model.LinkedinCampaignGroup
}

// Since we dont have a way to store raw format, we are going with the approach of joins on query.
func (store *MemSQL) GetLatestMetaForLinkedinForGivenDays(projectID int64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields) {
	logFields := log.Fields{
		"project_id": projectID,
		"days":       days,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	channelDocumentsCampaign := make([]model.ChannelDocumentsWithFields, 0, 0)
	channelDocumentsAdGroup := make([]model.ChannelDocumentsWithFields, 0, 0)

	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("Failed to get project settings")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	if projectSetting.IntLinkedinAdAccount == "" {
		log.WithField("projectID", projectID).Error("Integration of linkedin is not available for this project.")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	customerAccountIDs := strings.Split(projectSetting.IntLinkedinAdAccount, ",")

	to, err := strconv.ParseUint(time.Now().Format("20060102"), 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to parse to timestamp")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	from, err := strconv.ParseUint(time.Now().AddDate(0, 0, -days).Format("20060102"), 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to parse from timestamp")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	query := linkedinAdGroupMetadataFetchQueryStr
	params := []interface{}{LinkedinDocumentTypeAlias["campaign"], projectID, from, to,
		customerAccountIDs, LinkedinDocumentTypeAlias["campaign"], projectID, from, to,
		customerAccountIDs, LinkedinDocumentTypeAlias["campaign_group"], projectID, from, to,
		customerAccountIDs, LinkedinDocumentTypeAlias["campaign_group"], projectID, from, to,
		customerAccountIDs}

	rows1, tx1, err, queryID1 := store.ExecQueryWithContext(query, params)
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d ad_group meta for facebook", days)
		log.WithError(err).WithField("error string", err).Error(errString)
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

	query = linkedinCampaignMetadataFetchQueryStr
	params = []interface{}{LinkedinDocumentTypeAlias["campaign_group"], projectID, from, to,
		customerAccountIDs, LinkedinDocumentTypeAlias["campaign_group"], projectID, from, to,
		customerAccountIDs}

	rows2, tx2, err, queryID2 := store.ExecQueryWithContext(query, params)
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d campaign meta for Linkedin", days)
		log.WithError(err).WithField("error string", err).Error(errString)
		U.CloseReadQuery(rows2, tx2)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	startReadTime := time.Now()
	for rows2.Next() {
		currentRecord := model.ChannelDocumentsWithFields{}
		rows2.Scan(&currentRecord.CampaignID, &currentRecord.CampaignName)
		channelDocumentsCampaign = append(channelDocumentsCampaign, currentRecord)
	}
	U.CloseReadQuery(rows2, tx2)
	U.LogReadTimeWithQueryRequestID(startReadTime, queryID2, &logFields)

	return channelDocumentsCampaign, channelDocumentsAdGroup
}

func (store *MemSQL) DeleteLinkedinIntegration(projectID int64) (int, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	updateValues := make(map[string]interface{})
	updateValues["int_linkedin_ad_account"] = nil
	updateValues["int_linkedin_access_token"] = nil
	updateValues["int_linkedin_refresh_token"] = nil
	updateValues["int_linkedin_refresh_token_expiry"] = nil
	updateValues["int_linkedin_access_token_expiry"] = nil

	err := db.Model(&model.ProjectSetting{}).Where("project_id = ?", projectID).Update(updateValues).Error
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// PullLinkedInRows - Function to pull LinkedIn campaign data
// Selecting VALUE, TIMESTAMP, TYPE from linkedin_documents and PROPERTIES, OBJECT_TYPE from smart_properties
// Left join smart_properties filtered by project_id and source=linkedin
// where linkedin_documents.value["campaign_id"] = smart_properties.object_id (when smart_properties.object_type = 1)
//
//	or linkedin_documents.value["ad_group_id"] = smart_properties.object_id (when smart_properties.object_type = 2)
//
// [make sure there aren't multiple smart_properties rows for a particular object,
// or weekly insights for LinkedIn would show incorrect data.]
func (store *MemSQL) PullLinkedInRowsV2(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
		"end_time":   endTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rawQuery := fmt.Sprintf("SELECT linDocs.id, linDocs.value, linDocs.timestamp, linDocs.type, sp.properties FROM linkedin_documents linDocs "+
		"LEFT JOIN smart_properties sp ON sp.project_id = %d AND sp.source = '%s' AND ((COALESCE(sp.object_type,1) = 1 AND sp.object_id = linDocs.campaign_group_id) OR (COALESCE(sp.object_type,2) = 2 AND sp.object_id = linDocs.campaign_id))"+
		"WHERE linDocs.project_id = %d AND UNIX_TIMESTAMP(linDocs.created_at) BETWEEN %d AND %d "+
		"LIMIT %d",
		projectID, model.ChannelLinkedin, projectID, startTime, endTime, model.AdReportsPullLimit+1)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}
