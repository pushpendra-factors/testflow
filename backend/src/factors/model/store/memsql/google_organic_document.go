package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"

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
		" " + "JSON_EXTRACT_STRING(value, ?) IS NOT NULL LIMIT 5000"
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
	googleOrganicObjectsAndProperties := store.buildObjectAndPropertiesForGoogleOrganic(model.ObjectsForGoogleOrganic)
	selectMetrics := model.SelectableMetricsForGoogleOrganic
	objectsAndProperties := googleOrganicObjectsAndProperties
	return &model.ChannelConfigResult{
		SelectMetrics:        selectMetrics,
		ObjectsAndProperties: objectsAndProperties,
	}
}

func (store *MemSQL) buildObjectAndPropertiesForGoogleOrganic(objects []string) []model.ChannelObjectAndProperties {
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

func (store *MemSQL) GetGoogleOrganicFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("projectID", projectID).WithField("req_id", reqID)
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
	logCtx = log.WithField("project_id", projectID).WithField("req_id", reqID)
	params := []interface{}{requestFilterProperty, projectID, requestFilterProperty}
	_, resultRows, err := store.ExecuteSQL(googleOrganicFilterQueryStr, params, logCtx)
	if err != nil {
		logCtx.WithError(err).WithField("query", googleOrganicFilterQueryStr).WithField("params", params).Error(model.GoogleOrganicSpecificError)
		return make([]interface{}, 0), http.StatusInternalServerError
	}
	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

func (store *MemSQL) GetAllGoogleOrganicLastSyncInfoForAllProjects() ([]model.GoogleOrganicLastSyncInfo, int) {
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

func (store *MemSQL) GetGoogleOrganicLastSyncInfoForProject(projectID uint64) ([]model.GoogleOrganicLastSyncInfo, int) {
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

	googleOrganicSettingsByProjectAndURL := make(map[uint64]map[string]*model.GoogleOrganicProjectSettings)

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
	existingProjectAndURLWithTypes := make(map[uint64]map[string]map[int64]bool)
	selectedLastSyncInfos := make([]model.GoogleOrganicLastSyncInfo, 0)

	for i := range googleOrganicLastSyncInfos {
		logCtx := log.WithFields(
			log.Fields{"project_id": googleOrganicLastSyncInfos[i].ProjectId,
				"url": googleOrganicLastSyncInfos[i].URLPrefix})

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
	logCtx := log.WithField("url", googleOrganicDocument.URLPrefix).WithField(
		"project_id", googleOrganicDocument.ProjectID)

	if googleOrganicDocument.URLPrefix == "" {
		logCtx.Error("Invalid search console document.")
		return http.StatusBadRequest
	}
	return http.StatusOK
}

// Assigning id, campaignId columns with values from json...
func addColumnInformationForGoogleOrganicDocuments(googleOrganicDocuments []model.GoogleOrganicDocument) ([]model.GoogleOrganicDocument, int) {
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
	currentTime := gorm.NowFunc()
	googleOrganicDocument.CreatedAt = currentTime
	googleOrganicDocument.UpdatedAt = currentTime

	return http.StatusOK
}

func (store *MemSQL) ExecuteGoogleOrganicChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	fetchSource := false
	logCtx := log.WithField("xreq_id", reqID)
	if query.GetGroupByTimestamp() == "" {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForGoogleOrganicQueryV1(projectID, query, reqID, fetchSource, " LIMIT 10000", false, nil)
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
func getGroupByCombinationsForSearchConsole(columns []string, resultMetrics [][]interface{}) []map[string]interface{} {
	groupByCombinations := make([]map[string]interface{}, 0)
	for _, resultRow := range resultMetrics {
		groupByCombination := make(map[string]interface{})
		for index, column := range columns {
			if strings.HasPrefix(column, "organic_property_") {
				dimension := strings.TrimPrefix(column, "organic_property_")
				groupByCombination[dimension] = resultRow[index]
			}
		}
		groupByCombinations = append(groupByCombinations, groupByCombination)
	}
	return groupByCombinations
}
func buildWhereConditionForGBTForSearchConsole(groupByCombinations []map[string]interface{}) (string, []interface{}) {
	whereConditionForGBT := ""
	params := make([]interface{}, 0)
	for _, groupByCombination := range groupByCombinations {
		whereConditionForEachCombination := ""
		for dimension, value := range groupByCombination {
			if whereConditionForEachCombination == "" {
				whereConditionForEachCombination = fmt.Sprintf("JSON_EXTRACT_STRING(value, '%s') = ? ", dimension)
				params = append(params, value)
			} else {
				whereConditionForEachCombination += fmt.Sprintf(" AND JSON_EXTRACT_STRING(value, '%s') = ? ", dimension)
				params = append(params, value)
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

// GetSQLQueryAndParametersForGoogleOrganicQueryV1 ...
func (store *MemSQL) GetSQLQueryAndParametersForGoogleOrganicQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, int) {
	var selectMetrics []string
	var sql string
	var selectKeys []string
	var params []interface{}
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
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

func (store *MemSQL) transFormRequestFieldsAndFetchRequiredFieldsForGoogleOrganic(projectID uint64, query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, string, error) {
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

func buildGoogleOrganicQueryV1(query *model.ChannelQueryV1, projectID uint64, urlPrefixes string, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string) {
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
func (store *MemSQL) DeleteGoogleOrganicIntegration(projectID uint64) (int, error) {
	db := C.GetServices().Db
	projectSetting := model.ProjectSetting{}

	err := db.Model(&model.ProjectSetting{}).Where("project_id = ?", projectID).Find(&projectSetting).Error
	if err != nil {
		return http.StatusInternalServerError, err
	}
	agentUpdateValues := make(map[string]interface{})
	agentUpdateValues["int_google_organic_refresh_token"] = nil
	err = db.Model(&model.Agent{}).Where("uuid = ?", *projectSetting.IntGoogleOrganicEnabledAgentUUID).Update(agentUpdateValues).Error
	if err != nil {
		return http.StatusInternalServerError, err
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
