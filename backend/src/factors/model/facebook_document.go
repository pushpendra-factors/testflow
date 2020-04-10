package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type FacebookDocument struct {
	ProjectId           uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	CustomerAdAccountId string          `gorm:"primary_key:true;auto_increment:false" json:"customer_ad_account_id"`
	Platform            string          `gorm:"primary_key:true;auto_increment:false" json:"platform`
	TypeAlias           string          `gorm:"-" json:"type_alias"`
	Type                int             `gorm:"primary_key:true;auto_increment:false" json:"-"`
	Timestamp           int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	ID                  string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Value               *postgres.Jsonb `json:"value"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

var facebookDocumentTypeAlias = map[string]int{
	"ad_account":        0,
	"campaign":          1,
	"ad":                2,
	"ad_set":            3,
	"ad_insights":       4,
	"campaign_insights": 5,
	"ad_set_insights":   6,
}

const platform = "platform"

const error_DuplicateFacebookDocument = "pq: duplicate key value violates unique constraint \"facebook_documents_pkey\""

func isDuplicateFacebookDocumentError(err error) bool {
	return err.Error() == error_DuplicateFacebookDocument
}

func getFacebookId(valueJson *postgres.Jsonb) (string, error) {

	valueMap, err := U.DecodePostgresJsonb(valueJson)
	if err != nil {
		return "", err
	}
	id, exists := (*valueMap)["id"]
	if !exists {
		return "", fmt.Errorf("id field %s does not exist")
	}

	if id == nil {
		return "", fmt.Errorf("id field %s has empty value")
	}

	idStr, err := U.GetValueAsString(id)
	if err != nil {
		return "", err
	}

	// Id as string always.
	return idStr, nil
}

func CreateFacebookDocument(projectId uint64, document *FacebookDocument) int {
	logCtx := log.WithField("customer_acc_id", document.CustomerAdAccountId).WithField(
		"project_id", document.ProjectId)

	if document.CustomerAdAccountId == "" || document.TypeAlias == "" {
		logCtx.Error("Invalid facebook document.")
		return http.StatusBadRequest
	}
	if document.ProjectId == 0 || document.Timestamp == 0 || document.Platform == "" {
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

	db := C.GetServices().Db
	err := db.Create(&document).Error
	if err != nil {
		if isDuplicateFacebookDocumentError(err) {
			logCtx.WithError(err).WithField("id", document.ID).WithField("plateform", document.Platform).Error(
				"Failed to create an facebook doc. Duplicate.")
			return http.StatusConflict

			logCtx.WithError(err).WithField("id", document.ID).WithField("plateform", document.Platform).Error(
				"Failed to create an facebook doc. Continued inserting other docs.")
			return http.StatusInternalServerError
		}
	}

	return http.StatusCreated
}

type FacebookLastSyncInfo struct {
	ProjectId           uint64 `json:"project_id"`
	CustomerAdAccountId string `json:"customer_ad_acc_id"`
	Platform            string `json:"platform"`
	DocumentType        int    `json:"-"`
	DocumentTypeAlias   string `json:"type_alias"`
	LastTimestamp       int64  `json:"last_timestamp"`
}

func getFacebookDocumentTypeAliasByType() map[int]string {
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range facebookDocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

func GetFacebookLastSyncInfo() ([]FacebookLastSyncInfo, int) {
	db := C.GetServices().Db

	facebookLastSyncInfos := make([]FacebookLastSyncInfo, 0, 0)

	queryStr := "SELECT project_id, customer_ad_account_id, platform, type as document_type, max(timestamp) as last_timestamp" +
		" " + "FROM facebook_documents GROUP BY project_id, customer_ad_account_id, platform, type "

	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last facebook documents by type for sync info.")
		return facebookLastSyncInfos, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var facebookLastSyncInfo FacebookLastSyncInfo
		if err := db.ScanRows(rows, &facebookLastSyncInfo); err != nil {
			log.WithError(err).Error("Failed to scan last facebook documents by type for sync info.")
			return []FacebookLastSyncInfo{}, http.StatusInternalServerError
		}

		facebookLastSyncInfos = append(facebookLastSyncInfos, facebookLastSyncInfo)
	}
	documentTypeAliasByType := getFacebookDocumentTypeAliasByType()

	for i := range facebookLastSyncInfos {
		logCtx := log.WithField("project_id", facebookLastSyncInfos[i].ProjectId)
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

func ExecuteFacebookChannelQuery(projectId uint64, query *ChannelQuery) (*ChannelQueryResult, int) {
	logCtx := log.WithField("project_id", projectId).WithField("query", query)

	if projectId == 0 || query == nil {
		logCtx.Error("Invalid project_id or query on execute facebook channel query.")
		return nil, http.StatusInternalServerError
	}

	projectSetting, errCode := GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	if projectSetting.IntFacebookAdAccount == "" {
		logCtx.Error("Execute facebook channel query failed. No customer account id.")
		return nil, http.StatusInternalServerError
	}
	queryResult := &ChannelQueryResult{}
	result, err := getFacebookChannelResult(projectId, projectSetting.IntFacebookAdAccount, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get facebook query result.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult = result
	if query.Breakdown == "" {
		return queryResult, http.StatusOK
	}

	metricBreakDown, err := getFacebookMetricBreakdown(projectId, projectSetting.IntFacebookAdAccount, query)
	queryResult.MetricsBreakdown = metricBreakDown
	return queryResult, http.StatusOK
}
func getFacebookMetricBreakdown(projectId uint64, customerAccountId string, query *ChannelQuery) (*ChannelBreakdownResult, error) {
	logCtx := log.WithField("project_id", projectId).WithField("customer_account_id", customerAccountId)

	sqlQuery, documentType := getFacebookMetricsQuery(query, true)

	db := C.GetServices().Db
	rows, err := db.Raw(sqlQuery, projectId, customerAccountId,
		query.From,
		query.To,
		documentType).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build channel query result.")
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows)
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

	return &ChannelBreakdownResult{Headers: resultHeaders, Rows: resultRows}, nil
}
func getFacebookDocumentType(query *ChannelQuery) int {
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

func getFacebookMetricsQuery(query *ChannelQuery, withBreakdown bool) (string, int) {

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

	strmntWhere := "WHERE project_id= ? AND customer_ad_account_id = ? AND timestamp>? AND timestamp<? AND type=?"

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

	sqlQuery := "SELECT" + " " + selectCols + " " + "FROM facebook_documents" + " " + strmntWhere + " " + strmntGroupBy
	return sqlQuery, documentType
}

func getFacebookChannelResult(projectId uint64, customerAccountId string, query *ChannelQuery) (*ChannelQueryResult, error) {

	logCtx := log.WithField("project_id", projectId)

	sqlQuery, documentType := getFacebookMetricsQuery(query, false)

	queryResult := &ChannelQueryResult{}
	db := C.GetServices().Db
	rows, err := db.Raw(sqlQuery, projectId, customerAccountId,
		query.From,
		query.To,
		documentType).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build channel query result.")
		return queryResult, err
	}
	resultHeaders, resultRows, err := U.DBReadRows(rows)
	if err != nil {
		return nil, err
	}
	if len(resultRows) == 0 {
		log.Error("Aggregate query returned zero rows.")
		return nil, errors.New("no rows returned")
	}

	if len(resultRows) > 1 {
		log.Error("Aggregate query returned more than one row on get adwords metric kvs.")
	}

	metricKvs := make(map[string]interface{})
	for i, k := range resultHeaders {
		metricKvs[k] = resultRows[0][i]
	}

	queryResult.Metrics = &metricKvs
	return queryResult, nil
}
