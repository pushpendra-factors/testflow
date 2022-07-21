package memsql

import (
	"database/sql"
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const (
	lastSyncInfoQueryForAllProjectsGoogleOrganic = "SELECT project_id, url_prefix, max(timestamp) as last_timestamp, type" +
		" " + "FROM google_organic_documents GROUP BY project_id, url_prefix, type"
	lastSyncInfoForAProjectGoogleOrganic = "SELECT project_id, url_prefix, max(timestamp) as last_timestamp, type" +
		" " + "FROM google_organic_documents WHERE project_id = ? GROUP BY project_id, url_prefix, type"
	insertGoogleOrganicDocumentsStr = "INSERT INTO google_organic_documents (id,project_id,url_prefix,timestamp,value,type,created_at,updated_at) VALUES "
	googleOrganicFilterQueryStr     = "SELECT DISTINCT LCASE(JSON_EXTRACT_STRING(value, ?)) as filter_value FROM google_organic_documents WHERE project_id = ? AND" +
		" " + "JSON_EXTRACT_STRING(value, ?) IS NOT NULL AND timestamp BETWEEN ? AND ? LIMIT 5000"
	fromGoogleOrganicDocuments                              = " FROM google_organic_documents "
	staticWhereStatementForGoogleOrganic                    = "WHERE project_id = ? AND url_prefix IN ( ? ) AND timestamp between ? AND ? "
	weightedMetricsExpressionOfDivisionWithHandleOf0AndNull = "SUM(JSON_EXTRACT_STRING(value, '%s')*JSON_EXTRACT_STRING(value, '%s'))/(case when sum(JSON_EXTRACT_STRING(value, '%s')) = 0 then 100000 else NULLIF(sum(JSON_EXTRACT_STRING(value, '%s')), 100000) end)"
)

var googleOrganicMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions":                      "SUM(JSON_EXTRACT_STRING(value, 'impressions'))",
	"clicks":                           "SUM(JSON_EXTRACT_STRING(value, 'clicks'))",
	model.ClickThroughRate:             fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "clicks", "100", "impressions", "impressions"),
	"position_avg":                     "AVG(JSON_EXTRACT_STRING(value, 'position'))",
	"position_impression_weighted_avg": fmt.Sprintf(weightedMetricsExpressionOfDivisionWithHandleOf0AndNull, "position", "impressions", "impressions", "impressions"),
}

func (store *MemSQL) buildGoogleOrganicChannelConfig() *model.ChannelConfigResult {

	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	googleOrganicObjectsAndProperties := store.buildObjectAndPropertiesForGoogleOrganic(model.ObjectsForGoogleOrganic)
	selectMetrics := model.SelectableMetricsForGoogleOrganic
	objectsAndProperties := googleOrganicObjectsAndProperties
	return &model.ChannelConfigResult{
		SelectMetrics:        selectMetrics,
		ObjectsAndProperties: objectsAndProperties,
	}
}

func (store *MemSQL) buildObjectAndPropertiesForGoogleOrganic(objects []string) []model.ChannelObjectAndProperties {
	logFields := log.Fields{
		"objects": objects,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0)
	for _, currentObject := range objects {
		propertiesAndRelated, isPresent := model.MapOfObjectsToPropertiesAndRelatedGoogleOrganic[currentObject]
		var currentProperties []model.ChannelProperty
		if isPresent {
			currentProperties = buildProperties(propertiesAndRelated)
		} else {
			return make([]model.ChannelObjectAndProperties, 0)
		}
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
}

func (store *MemSQL) IsGoogleOrganicIntegrationAvailable(projectID int64) bool {
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return false
	}
	urlPrefix := projectSetting.IntGoogleOrganicURLPrefixes

	if urlPrefix == nil || *urlPrefix == "" {
		return false
	}
	return true
}
func (store *MemSQL) GetGoogleOrganicFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to fetch Project Setting in searcch console filter values.")
		return []interface{}{}, http.StatusInternalServerError
	}
	urlPrefix := projectSetting.IntGoogleOrganicURLPrefixes

	if urlPrefix == nil || *urlPrefix == "" {
		logCtx.Info(integrationNotAvailable)
		return []interface{}{}, http.StatusNotFound
	}

	from, to := model.GetFromAndToDatesForFilterValues()
	logCtx = log.WithField("project_id", projectID).WithField("req_id", reqID)
	params := []interface{}{requestFilterProperty, projectID, requestFilterProperty, from, to}
	_, resultRows, err := store.ExecuteSQL(googleOrganicFilterQueryStr, params, logCtx)
	if err != nil {
		logCtx.WithError(err).WithField("query", googleOrganicFilterQueryStr).WithField("params", params).Error(model.GoogleOrganicSpecificError)
		return make([]interface{}, 0), http.StatusInternalServerError
	}
	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

func (store *MemSQL) GetAllGoogleOrganicLastSyncInfoForAllProjects() ([]model.GoogleOrganicLastSyncInfo, int) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	params := make([]interface{}, 0)
	googleOrganicLastSyncInfos, status := getGoogleOrganicLastSyncInfo(lastSyncInfoQueryForAllProjectsGoogleOrganic, params)
	if status != http.StatusOK {
		return googleOrganicLastSyncInfos, status
	}

	googleOrganicSettings, errCode := store.GetAllIntGoogleOrganicProjectSettings()
	if errCode != http.StatusOK {
		return []model.GoogleOrganicLastSyncInfo{}, errCode
	}
	log.Info("All settings ", googleOrganicSettings)

	return sanitizedLastSyncInfosGoogleOrganic(googleOrganicLastSyncInfos, googleOrganicSettings)
}

func (store *MemSQL) GetGoogleOrganicLastSyncInfoForProject(projectID int64) ([]model.GoogleOrganicLastSyncInfo, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	params := []interface{}{projectID}
	googleOrganicLastSyncInfos, status := getGoogleOrganicLastSyncInfo(lastSyncInfoForAProjectGoogleOrganic, params)
	if status != http.StatusOK {
		return googleOrganicLastSyncInfos, status
	}
	googleOrganicSettings, errCode := store.GetIntGoogleOrganicProjectSettingsForProjectID(projectID)
	if errCode != http.StatusOK {
		return []model.GoogleOrganicLastSyncInfo{}, errCode
	}

	return sanitizedLastSyncInfosGoogleOrganic(googleOrganicLastSyncInfos, googleOrganicSettings)
}

func getGoogleOrganicLastSyncInfo(query string, params []interface{}) ([]model.GoogleOrganicLastSyncInfo, int) {
	logFields := log.Fields{
		"query":  query,
		"params": params,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	googleOrganicLastSyncInfos := make([]model.GoogleOrganicLastSyncInfo, 0)

	rows, err := db.Raw(query, params).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last googleOrganic documents by type for sync info.")
		return googleOrganicLastSyncInfos, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var googleOrganicLastSyncInfo model.GoogleOrganicLastSyncInfo
		if err := db.ScanRows(rows, &googleOrganicLastSyncInfo); err != nil {
			log.WithError(err).Error("Failed to scan last googleOrganic documents by type for sync info.")
			return []model.GoogleOrganicLastSyncInfo{}, http.StatusInternalServerError
		}

		googleOrganicLastSyncInfos = append(googleOrganicLastSyncInfos, googleOrganicLastSyncInfo)
	}

	return googleOrganicLastSyncInfos, http.StatusOK
}

// This method handles adding additionalInformation to lastSyncInfo, Skipping inactive Projects and adding missed LastSync.
func sanitizedLastSyncInfosGoogleOrganic(googleOrganicLastSyncInfos []model.GoogleOrganicLastSyncInfo, googleOrganicSettings []model.GoogleOrganicProjectSettings) ([]model.GoogleOrganicLastSyncInfo, int) {
	logFields := log.Fields{
		"google_organic_last_sync_infos": googleOrganicLastSyncInfos,
		"google_organ_settings":          googleOrganicSettings,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	googleOrganicSettingsByProjectAndURL := make(map[int64]map[string]*model.GoogleOrganicProjectSettings)

	for i := range googleOrganicSettings {
		URLs := strings.Split(googleOrganicSettings[i].URLPrefix, ",")
		googleOrganicSettingsByProjectAndURL[googleOrganicSettings[i].ProjectID] = make(map[string]*model.GoogleOrganicProjectSettings)
		for j := range URLs {
			var setting model.GoogleOrganicProjectSettings
			setting.ProjectID = googleOrganicSettings[i].ProjectID
			setting.AgentUUID = googleOrganicSettings[i].AgentUUID
			setting.RefreshToken = googleOrganicSettings[i].RefreshToken
			setting.URLPrefix = URLs[j]
			googleOrganicSettingsByProjectAndURL[googleOrganicSettings[i].ProjectID][URLs[j]] = &setting
		}
	}

	// add settings for project_id existing on googleOrganic documents.
	existingProjectAndURLWithTypes := make(map[int64]map[string]map[int64]bool)
	selectedLastSyncInfos := make([]model.GoogleOrganicLastSyncInfo, 0)

	for i := range googleOrganicLastSyncInfos {
		logCtx := log.WithFields(logFields)

		settings, exists := googleOrganicSettingsByProjectAndURL[googleOrganicLastSyncInfos[i].ProjectId][googleOrganicLastSyncInfos[i].URLPrefix]
		if !exists {
			logCtx.Warn("GoogleOrganic project settings not found for url googleOrganic synced earlier.")
		}

		if settings == nil {
			logCtx.Info("GoogleOrganic disabled for project.")
			continue
		}

		googleOrganicLastSyncInfos[i].RefreshToken = settings.RefreshToken

		selectedLastSyncInfos = append(selectedLastSyncInfos, googleOrganicLastSyncInfos[i])

		if _, projectWithURLExists := existingProjectAndURLWithTypes[googleOrganicLastSyncInfos[i].ProjectId][googleOrganicLastSyncInfos[i].URLPrefix]; !projectWithURLExists {
			if _, projectExists := existingProjectAndURLWithTypes[googleOrganicLastSyncInfos[i].ProjectId]; !projectExists {
				existingProjectAndURLWithTypes[googleOrganicLastSyncInfos[i].ProjectId] = make(map[string]map[int64]bool)
			}
			existingProjectAndURLWithTypes[googleOrganicLastSyncInfos[i].ProjectId][googleOrganicLastSyncInfos[i].URLPrefix] = make(map[int64]bool)
		}
		existingProjectAndURLWithTypes[googleOrganicLastSyncInfos[i].ProjectId][googleOrganicLastSyncInfos[i].URLPrefix][googleOrganicLastSyncInfos[i].Type] = true
	}

	// add all types for missing projects and
	// add missing types for existing projects.
	for i := range googleOrganicSettings {
		URLs := strings.Split(googleOrganicSettings[i].URLPrefix, ",")
		for _, URL := range URLs {
			existingTypeForURL, urlExists := existingProjectAndURLWithTypes[googleOrganicSettings[i].ProjectID][URL]
			for _, docType := range model.GoogleOrganicTypes {
				if !urlExists || (urlExists && !existingTypeForURL[docType]) {
					syncInfo := model.GoogleOrganicLastSyncInfo{
						ProjectId:     googleOrganicSettings[i].ProjectID,
						RefreshToken:  googleOrganicSettings[i].RefreshToken,
						URLPrefix:     URL,
						LastTimestamp: 0, // no sync yet.
						Type:          docType,
					}
					selectedLastSyncInfos = append(selectedLastSyncInfos, syncInfo)
				}
			}
		}

	}

	return selectedLastSyncInfos, http.StatusOK
}

// CreateGoogleOrganicDocument ...
func (store *MemSQL) CreateGoogleOrganicDocument(googleOrganicDoc *model.GoogleOrganicDocument) int {
	logFields := log.Fields{
		"gooogle_organic_doc": googleOrganicDoc,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	status := validateGoogleOrganicDocument(googleOrganicDoc)
	if status != http.StatusOK {
		return status
	}

	status = addColumnInformationForGoogleOrganicDocument(googleOrganicDoc)
	if status != http.StatusOK {
		return status
	}

	db := C.GetServices().Db
	dbc := db.Table("google_organic_documents").Create(googleOrganicDoc)

	if dbc.Error != nil {
		if IsDuplicateRecordError(dbc.Error) {
			log.WithError(dbc.Error).WithField("googleOrganicDocuments", googleOrganicDoc).Warn("Failed to create an search console doc. Duplicate")
			return http.StatusConflict
		}
		log.WithError(dbc.Error).WithField("googleOrganicDocuments", googleOrganicDoc).Error("Failed to create an search console doc.")
		return http.StatusInternalServerError
	}
	return http.StatusCreated
}

// CreateMultipleGoogleOrganicDocument ...
func (store *MemSQL) CreateMultipleGoogleOrganicDocument(googleOrganicDocuments []model.GoogleOrganicDocument) int {
	logFields := log.Fields{
		"google_organic_documents": googleOrganicDocuments,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	status := validateGoogleOrganicDocuments(googleOrganicDocuments)
	if status != http.StatusOK {
		return status
	}
	googleOrganicDocuments, status = addColumnInformationForGoogleOrganicDocuments(googleOrganicDocuments)
	if status != http.StatusOK {
		return status
	}

	db := C.GetServices().Db

	insertStatement := insertGoogleOrganicDocumentsStr
	insertValuesStatement := make([]string, 0)
	insertValues := make([]interface{}, 0)
	for _, googleOrganicDoc := range googleOrganicDocuments {
		insertValuesStatement = append(insertValuesStatement, "(?, ?, ?, ?, ?, ?, ?, ?)")
		insertValues = append(insertValues, googleOrganicDoc.ID, googleOrganicDoc.ProjectID, googleOrganicDoc.URLPrefix,
			googleOrganicDoc.Timestamp, googleOrganicDoc.Value, googleOrganicDoc.Type, googleOrganicDoc.CreatedAt, googleOrganicDoc.UpdatedAt)
	}
	insertStatement += joinWithComma(insertValuesStatement...)
	rows, err := db.Raw(insertStatement, insertValues...).Rows()

	if err != nil {
		if IsDuplicateRecordError(err) {
			log.WithError(err).WithField("googleOrganicDocuments", googleOrganicDocuments).Warn(
				"Failed to create an googleOrganic doc. Duplicate. Continued inserting other docs.")
			return http.StatusConflict
		}
		log.WithError(err).WithField("googleOrganicDocuments", googleOrganicDocuments).Error(
			"Failed to create an googleOrganic doc. Continued inserting other docs.")
		return http.StatusInternalServerError
	}
	defer rows.Close()

	return http.StatusCreated
}

func validateGoogleOrganicDocuments(googleOrganicDocuments []model.GoogleOrganicDocument) int {
	logFields := log.Fields{
		"google_organic_documents": googleOrganicDocuments,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for index := range googleOrganicDocuments {
		status := validateGoogleOrganicDocument(&googleOrganicDocuments[index])
		if status != http.StatusOK {
			log.WithField("index", index).Error("Failed in this index")
			return status
		}
	}
	return http.StatusOK
}

func validateGoogleOrganicDocument(googleOrganicDocument *model.GoogleOrganicDocument) int {
	logFields := log.Fields{
		"google_organic_document": googleOrganicDocument,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if googleOrganicDocument.URLPrefix == "" {
		logCtx.Error("Invalid search console document.")
		return http.StatusBadRequest
	}
	return http.StatusOK
}

// Assigning id, campaignId columns with values from json...
func addColumnInformationForGoogleOrganicDocuments(googleOrganicDocuments []model.GoogleOrganicDocument) ([]model.GoogleOrganicDocument, int) {
	logFields := log.Fields{
		"google_organic_documents": googleOrganicDocuments,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for index := range googleOrganicDocuments {
		status := addColumnInformationForGoogleOrganicDocument(&googleOrganicDocuments[index])
		if status != http.StatusOK {
			log.WithField("index", index).Error("Failed in this index")
			return googleOrganicDocuments, status
		}
	}
	return googleOrganicDocuments, http.StatusOK
}
func addColumnInformationForGoogleOrganicDocument(googleOrganicDocument *model.GoogleOrganicDocument) int {
	logFields := log.Fields{
		"google_organic_document": googleOrganicDocument,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	currentTime := gorm.NowFunc()
	googleOrganicDocument.CreatedAt = currentTime
	googleOrganicDocument.UpdatedAt = currentTime

	return http.StatusOK
}

func (store *MemSQL) ExecuteGoogleOrganicChannelQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	fetchSource := false
	logCtx := log.WithFields(logFields)
	if query.GetGroupByTimestamp() == "" {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForGoogleOrganicQueryV1(projectID, query, reqID, fetchSource, " LIMIT 10000", false, nil)
		if errCode == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := store.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.GoogleOrganicSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	} else {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForGoogleOrganicQueryV1(projectID, query, reqID, fetchSource, " LIMIT 100", false, nil)
		if errCode == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := store.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.GoogleOrganicSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		groupByCombinations := getGroupByCombinationsForSearchConsole(columns, resultMetrics)
		sql, params, selectKeys, selectMetrics, errCode = store.GetSQLQueryAndParametersForGoogleOrganicQueryV1(projectID, query, reqID, fetchSource, " LIMIT 10000", true, groupByCombinations)
		if errCode == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err = store.ExecuteSQL(sql, params, logCtx)
		columns = append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.GoogleOrganicSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	}
}

// Edge Case: When there is no data for a day, we dont want to include it i.e. value == null.
func getGroupByCombinationsForSearchConsole(columns []string, resultMetrics [][]interface{}) map[string][]interface{} {
	logFields := log.Fields{
		"columns":        columns,
		"result_metrics": resultMetrics,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	groupByCombinations := make(map[string][]interface{}, 0)

	for _, resultRow := range resultMetrics {
		for index, column := range columns {
			if resultRow[index] != nil {
				if strings.HasPrefix(column, "organic_property_") {
					propertyWithoutOrganicPrefix := strings.TrimPrefix(column, "organic_property_")
					if value, exists := groupByCombinations[propertyWithoutOrganicPrefix]; exists {
						value = append(value, resultRow[index])
						groupByCombinations[propertyWithoutOrganicPrefix] = value
					} else {
						value = make([]interface{}, 0)
						value = append(value, resultRow[index])
						groupByCombinations[propertyWithoutOrganicPrefix] = value
					}
				}
			}
		}
	}

	return groupByCombinations
}

func buildWhereConditionForGBTForSearchConsole(groupByCombinations map[string][]interface{}) (string, []interface{}) {
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
		resultantInClauses = append(resultantInClauses, fmt.Sprintf("JSON_EXTRACT_STRING(value, '%s') IN (", dimension)+currentInClause+") ")
	}

	resultantWhereCondition = joinWithWordInBetween("AND", resultantInClauses...)

	return resultantWhereCondition, params
}

// GetSQLQueryAndParametersForGoogleOrganicQueryV1 ...
func (store *MemSQL) GetSQLQueryAndParametersForGoogleOrganicQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string, int) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"req_id":                        reqID,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by+combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var selectMetrics []string
	var sql string
	var selectKeys []string
	var params []interface{}
	logCtx := log.WithFields(logFields)
	transformedQuery, urlPrefix, err := store.transFormRequestFieldsAndFetchRequiredFieldsForGoogleOrganic(projectID, *query, reqID)
	if err != nil && err.Error() == integrationNotAvailable {
		logCtx.WithError(err).Info(model.GoogleOrganicSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusNotFound
	}
	if err != nil {
		logCtx.WithError(err).Error(model.GoogleOrganicSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusBadRequest
	}

	sql, params, selectKeys, selectMetrics = buildGoogleOrganicQueryV1(transformedQuery, projectID, urlPrefix, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
	}
	return sql, params, selectKeys, selectMetrics, http.StatusOK
}

func (store *MemSQL) transFormRequestFieldsAndFetchRequiredFieldsForGoogleOrganic(projectID int64, query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, string, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	query.From = U.GetDateAsStringIn(query.From, U.TimeZoneString(query.Timezone))
	query.To = U.GetDateAsStringIn(query.To, U.TimeZoneString(query.Timezone))
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return &model.ChannelQueryV1{}, "", errors.New("Project setting not found")
	}
	url_prefix := projectSetting.IntGoogleOrganicURLPrefixes
	if url_prefix == nil || *url_prefix == "" {
		return &model.ChannelQueryV1{}, "", errors.New(integrationNotAvailable)
	}
	return &query, *url_prefix, nil
}

func checkIfOnlyPageExistsInFiltersAndGroupBys(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) bool {
	logFields := log.Fields{
		"filters":   filters,
		"group_bys": groupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for _, filter := range filters {
		if filter.Property != "page" {
			return false
		}
	}
	for _, groupBy := range groupBys {
		if groupBy.Property != "page" {
			return false
		}
	}
	return true
}

func buildGoogleOrganicQueryV1(query *model.ChannelQueryV1, projectID int64, urlPrefixes string, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"url_prefixes":                  urlPrefixes,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by+combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	customerUrlPrefixes := strings.Split(urlPrefixes, ",")
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	finalSelectKeys := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	isPageLevelDataReq := checkIfOnlyPageExistsInFiltersAndGroupBys(query.Filters, query.GroupBy)
	// Group By
	for _, groupBy := range query.GroupBy {
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, groupBy.Object+"_"+groupBy.Property)
	}
	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, model.AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	for _, groupBy := range query.GroupBy {
		value := fmt.Sprintf("JSON_EXTRACT_STRING(value, '%s') as %s", groupBy.Property, groupBy.Object+"_"+groupBy.Property)
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, groupBy.Object+"_"+groupBy.Property)
	}

	finalSelectKeys = append(finalSelectKeys, selectKeys...)
	if isGroupByTimestamp {
		finalSelectKeys = append(finalSelectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}

	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", googleOrganicMetricsToAggregatesInReportsMapping[selectMetric], selectMetric)
		selectMetrics = append(selectMetrics, value)

		value = selectMetric
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClauseForSearchConsole(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters := getGoogleOrganicFiltersWhereStatement(query.Filters)
	if isPageLevelDataReq {
		whereConditionForFilters += " AND type = 2 "
	} else {
		whereConditionForFilters += " AND type = 1 "
	}

	resultSQLStatement := selectQuery + fromGoogleOrganicDocuments + staticWhereStatementForGoogleOrganic + whereConditionForFilters
	whereConditionForGBT, whereParams := "", make([]interface{}, 0)
	if groupByCombinationsForGBT != nil && len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams = buildWhereConditionForGBTForSearchConsole(groupByCombinationsForGBT)
		if whereConditionForGBT != "" {
			resultSQLStatement += (" AND (" + whereConditionForGBT + ") ")
		}
	}
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + limitString + ";"
	staticWhereParams := []interface{}{projectID, customerUrlPrefixes, query.From, query.To}
	finalParams := staticWhereParams
	if len(whereParams) != 0 {
		finalParams = append(staticWhereParams, whereParams...)
	}
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

func getGoogleOrganicFiltersWhereStatement(filters []model.ChannelFilterV1) string {
	logFields := log.Fields{
		"filters": filters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultStatement := ""
	var filterValue string
	if len(filters) == 0 {
		return resultStatement
	}
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
		currentFilterStatement = fmt.Sprintf("JSON_EXTRACT_STRING(value, '%s') %s '%s' ", filter.Property, filterOperator, filterValue)
		if index == 0 {
			resultStatement = " AND ( " + currentFilterStatement
		} else {
			resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement + " )"

}
func (store *MemSQL) DeleteGoogleOrganicIntegration(projectID int64) (int, error) {
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
	if projectSetting.IntGoogleOrganicEnabledAgentUUID != nil {
		agentUpdateValues := make(map[string]interface{})
		agentUpdateValues["int_google_organic_refresh_token"] = nil
		err = db.Model(&model.Agent{}).Where("uuid = ?", *projectSetting.IntGoogleOrganicEnabledAgentUUID).Update(agentUpdateValues).Error
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	projectSettingUpdateValues := make(map[string]interface{})
	projectSettingUpdateValues["int_google_organic_url_prefixes"] = nil
	projectSettingUpdateValues["int_google_organic_enabled_agent_uuid"] = nil
	err = db.Model(&model.ProjectSetting{}).Where("project_id = ?", projectID).Update(projectSettingUpdateValues).Error
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// PullGoogleOrganicRows - Function to pull GoogleOrganic campaign data
// Selecting VALUE, TIMESTAMP, TYPE from google_organic_documents and PROPERTIES, OBJECT_TYPE from smart_properties
// Left join smart_properties filtered by project_id and source=google_ads
// where google_organic_documents.value["campaign_id"] = smart_properties.object_id (when smart_properties.object_type = 1)
//	 or google_organic_documents.value["ad_group_id"] = smart_properties.object_id (when smart_properties.object_type = 2)
// [make sure there aren't multiple smart_properties rows for a particular object,
// or weekly insights for google_organic would show incorrect data.]
func (store *MemSQL) PullGoogleOrganicRows(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error) {
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

	rawQuery := fmt.Sprintf("SELECT id, gooDocs.value, gooDocs.timestamp, gooDocs.type, sp.properties FROM google_organic_documents gooDocs "+
		"LEFT JOIN smart_properties sp ON sp.project_id = %d AND sp.source = '%s' AND sp.object_id = JSON_EXTRACT_STRING(gooDocs.value, 'id')"+
		"WHERE gooDocs.project_id = %d AND gooDocs.timestamp BETWEEN %d AND %d "+
		"ORDER BY gooDocs.type, timestamp LIMIT %d",
		projectID, model.ChannelGoogleAds, projectID, start, end, model.GoggleOrganicPullLimit+1)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}
