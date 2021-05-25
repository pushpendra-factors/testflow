package postgres

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
	errorDuplicateGoogleOrganicDocument          = "pq: duplicate key value violates unique constraint \"google_organic_documents_primary_key\""
	lastSyncInfoQueryForAllProjectsGoogleOrganic = "SELECT project_id, url_prefix, max(timestamp) as last_timestamp" +
		" " + "FROM google_organic_documents GROUP BY project_id, url_prefix"
	lastSyncInfoForAProjectGoogleOrganic = "SELECT project_id, url_prefix, max(timestamp) as last_timestamp" +
		" " + "FROM google_organic_documents WHERE project_id = ? GROUP BY project_id, url_prefix"
	insertGoogleOrganicDocumentsStr = "INSERT INTO google_organic_documents (id,project_id,url_prefix,timestamp,value,created_at,updated_at) VALUES "
	googleOrganicFilterQueryStr     = "SELECT DISTINCT(value->>?) as filter_value FROM google_organic_documents WHERE project_id = ? AND" +
		" " + "value->>? IS NOT NULL LIMIT 5000"
	fromGoogleOrganicDocuments                              = " FROM google_organic_documents "
	staticWhereStatementForGoogleOrganic                    = "WHERE project_id = ? AND url_prefix IN ( ? ) AND timestamp between ? AND ? "
	weightedMetricsExpressionOfDivisionWithHandleOf0AndNull = "SUM(((value->>'%s')::float)*((value->>'%s')::float))/(case when sum((value->>'%s')::float) = 0 then 100000 else NULLIF(sum((value->>'%s')::float), 100000) end)"
)

var googleOrganicMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions":                      "SUM((value->>'impressions')::float)",
	"clicks":                           "SUM((value->>'clicks')::float)",
	model.ClickThroughRate:             fmt.Sprintf(metricsExpressionOfDivisionWithHandleOf0AndNull, "clicks", "100", "impressions", "impressions"),
	"position_avg":                     "AVG((value->>'position')::float)",
	"position_impression_weighted_avg": fmt.Sprintf(weightedMetricsExpressionOfDivisionWithHandleOf0AndNull, "position", "impressions", "impressions", "impressions"),
}

var mapOfObjectsToPropertiesAndRelatedGoogleOrganic = map[string]map[string]PropertiesAndRelated{
	"organic_property": {
		"query":   PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		"page":    PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		"country": PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		"device":  PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
	},
}
var selectableMetricsForGoogleOrganic = []string{"impressions", "clicks", model.ClickThroughRate, "position_avg", "position_impression_weighted_avg"}
var objectsForGoogleOrganic = []string{"organic_property"}

func isDuplicateGoogleOrganicDocumentError(err error) bool {
	return err.Error() == errorDuplicateGoogleOrganicDocument
}

func (pg *Postgres) buildGoogleOrganicChannelConfig() *model.ChannelConfigResult {
	googleOrganicObjectsAndProperties := pg.buildObjectAndPropertiesForGoogleOrganic(objectsForGoogleOrganic)
	selectMetrics := selectableMetricsForGoogleOrganic
	objectsAndProperties := googleOrganicObjectsAndProperties
	return &model.ChannelConfigResult{
		SelectMetrics:        selectMetrics,
		ObjectsAndProperties: objectsAndProperties,
	}
}

func (pg *Postgres) buildObjectAndPropertiesForGoogleOrganic(objects []string) []model.ChannelObjectAndProperties {
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0)
	for _, currentObject := range objects {
		propertiesAndRelated, isPresent := mapOfObjectsToPropertiesAndRelatedGoogleOrganic[currentObject]
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

func (pg *Postgres) GetGoogleOrganicFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("projectID", projectID).WithField("req_id", reqID)
	projectSetting, errCode := pg.GetProjectSetting(projectID)
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
	_, resultRows, err := pg.ExecuteSQL(googleOrganicFilterQueryStr, params, logCtx)
	if err != nil {
		logCtx.WithError(err).WithField("query", googleOrganicFilterQueryStr).WithField("params", params).Error(model.GoogleOrganicSpecificError)
		return make([]interface{}, 0), http.StatusInternalServerError
	}
	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

func (pg *Postgres) GetAllGoogleOrganicLastSyncInfoForAllProjects() ([]model.GoogleOrganicLastSyncInfo, int) {
	params := make([]interface{}, 0)
	googleOrganicLastSyncInfos, status := getGoogleOrganicLastSyncInfo(lastSyncInfoQueryForAllProjectsGoogleOrganic, params)
	if status != http.StatusOK {
		return googleOrganicLastSyncInfos, status
	}

	googleOrganicSettings, errCode := pg.GetAllIntGoogleOrganicProjectSettings()
	if errCode != http.StatusOK {
		return []model.GoogleOrganicLastSyncInfo{}, errCode
	}
	log.Info("All settings ", googleOrganicSettings)

	return sanitizedLastSyncInfosGoogleOrganic(googleOrganicLastSyncInfos, googleOrganicSettings)
}

func (pg *Postgres) GetGoogleOrganicLastSyncInfoForProject(projectID uint64) ([]model.GoogleOrganicLastSyncInfo, int) {
	params := []interface{}{projectID}
	googleOrganicLastSyncInfos, status := getGoogleOrganicLastSyncInfo(lastSyncInfoForAProjectGoogleOrganic, params)
	if status != http.StatusOK {
		return googleOrganicLastSyncInfos, status
	}
	googleOrganicSettings, errCode := pg.GetIntGoogleOrganicProjectSettingsForProjectID(projectID)
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
	existingProjectAndURLWithTypes := make(map[uint64]map[string]bool)
	selectedLastSyncInfos := make([]model.GoogleOrganicLastSyncInfo, 0)

	for i := range googleOrganicLastSyncInfos {
		logCtx := log.WithFields(
			log.Fields{"project_id": googleOrganicLastSyncInfos[i].ProjectId,
				"url": googleOrganicLastSyncInfos[i].URLPrefix})

		settings, exists := googleOrganicSettingsByProjectAndURL[googleOrganicLastSyncInfos[i].ProjectId][googleOrganicLastSyncInfos[i].URLPrefix]
		if !exists {
			logCtx.Error("GoogleOrganic project settings not found for url googleOrganic synced earlier.")
		}

		if settings == nil {
			logCtx.Info("GoogleOrganic disabled for project.")
			continue
		}

		googleOrganicLastSyncInfos[i].RefreshToken = settings.RefreshToken

		selectedLastSyncInfos = append(selectedLastSyncInfos, googleOrganicLastSyncInfos[i])

		if _, projectWithURLExists := existingProjectAndURLWithTypes[googleOrganicLastSyncInfos[i].ProjectId][googleOrganicLastSyncInfos[i].URLPrefix]; !projectWithURLExists {
			if _, projectExists := existingProjectAndURLWithTypes[googleOrganicLastSyncInfos[i].ProjectId]; !projectExists {
				existingProjectAndURLWithTypes[googleOrganicLastSyncInfos[i].ProjectId] = make(map[string]bool)
			}
			existingProjectAndURLWithTypes[googleOrganicLastSyncInfos[i].ProjectId][googleOrganicLastSyncInfos[i].URLPrefix] = true
		}
	}

	// add all types for missing projects and
	// add missing types for existing projects.
	for i := range googleOrganicSettings {
		URLs := strings.Split(googleOrganicSettings[i].URLPrefix, ",")
		for _, URL := range URLs {
			_, urlExists := existingProjectAndURLWithTypes[googleOrganicSettings[i].ProjectID][URL]
			if !urlExists {
				syncInfo := model.GoogleOrganicLastSyncInfo{
					ProjectId:     googleOrganicSettings[i].ProjectID,
					RefreshToken:  googleOrganicSettings[i].RefreshToken,
					URLPrefix:     URL,
					LastTimestamp: 0, // no sync yet.
				}
				selectedLastSyncInfos = append(selectedLastSyncInfos, syncInfo)
			}
		}

	}

	return selectedLastSyncInfos, http.StatusOK
}

// CreateGoogleOrganicDocument ...
func (pg *Postgres) CreateGoogleOrganicDocument(googleOrganicDoc *model.GoogleOrganicDocument) int {
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
		if isDuplicateGoogleOrganicDocumentError(dbc.Error) {
			log.WithError(dbc.Error).WithField("googleOrganicDocuments", googleOrganicDoc).Warn("Failed to create search console doc. Duplicate")
			return http.StatusConflict
		}
		log.WithError(dbc.Error).WithField("googleOrganicDocuments", googleOrganicDoc).Error("Failed to create search console doc")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

// CreateMultipleGoogleOrganicDocument ...
func (pg *Postgres) CreateMultipleGoogleOrganicDocument(googleOrganicDocuments []model.GoogleOrganicDocument) int {
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
		insertValuesStatement = append(insertValuesStatement, "(?, ?, ?, ?, ?, ?, ?)")
		insertValues = append(insertValues, googleOrganicDoc.ID, googleOrganicDoc.ProjectID, googleOrganicDoc.URLPrefix,
			googleOrganicDoc.Timestamp, googleOrganicDoc.Value, googleOrganicDoc.CreatedAt, googleOrganicDoc.UpdatedAt)
	}
	insertStatement += joinWithComma(insertValuesStatement...)
	rows, err := db.Raw(insertStatement, insertValues...).Rows()

	if err != nil {
		if isDuplicateGoogleOrganicDocumentError(err) {
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

func (pg *Postgres) ExecuteGoogleOrganicChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
	fetchSource := false
	logCtx := log.WithField("xreq_id", reqID)
	sql, params, selectKeys, selectMetrics, errCode := pg.GetSQLQueryAndParametersForGoogleOrganicQueryV1(projectID, query, reqID, fetchSource)
	if errCode != http.StatusOK {
		return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
	}
	_, resultMetrics, err := pg.ExecuteSQL(sql, params, logCtx)
	columns := append(selectKeys, selectMetrics...)
	if err != nil {
		logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.GoogleOrganicSpecificError)
		return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
	}
	return columns, resultMetrics, http.StatusOK
}

// GetSQLQueryAndParametersForGoogleOrganicQueryV1 ...
func (pg *Postgres) GetSQLQueryAndParametersForGoogleOrganicQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool) (string, []interface{}, []string, []string, int) {
	var selectMetrics []string
	var sql string
	var selectKeys []string
	var params []interface{}
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	transformedQuery, urlPrefix, err := pg.transFormRequestFieldsAndFetchRequiredFieldsForGoogleOrganic(projectID, *query, reqID)
	if err != nil && err.Error() == integrationNotAvailable {
		logCtx.WithError(err).Info(model.GoogleOrganicSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusNotFound
	}
	if err != nil {
		logCtx.WithError(err).Error(model.GoogleOrganicSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusBadRequest
	}

	sql, params, selectKeys, selectMetrics = buildGoogleOrganicQueryV1(transformedQuery, projectID, urlPrefix)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
	}
	return sql, params, selectKeys, selectMetrics, http.StatusOK
}

func (pg *Postgres) transFormRequestFieldsAndFetchRequiredFieldsForGoogleOrganic(projectID uint64, query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, string, error) {
	query.From = U.GetDateAsStringZ(query.From, U.TimeZoneString(query.Timezone))
	query.To = U.GetDateAsStringZ(query.To, U.TimeZoneString(query.Timezone))
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return &model.ChannelQueryV1{}, "", errors.New("Project setting not found")
	}
	url_prefix := projectSetting.IntGoogleOrganicURLPrefixes
	if url_prefix == nil || *url_prefix == "" {
		return &model.ChannelQueryV1{}, "", errors.New(integrationNotAvailable)
	}
	return &query, *url_prefix, nil
}

//todo @ashhhar: rebase to adwords
func buildGoogleOrganicQueryV1(query *model.ChannelQueryV1, projectID uint64, urlPrefixes string) (string, []interface{}, []string, []string) {
	customerUrlPrefixes := strings.Split(urlPrefixes, ",")
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	finalSelectKeys := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

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
		value := fmt.Sprintf("value->>'%s' as %s", groupBy.Property, groupBy.Object+"_"+groupBy.Property)
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
	orderByQuery := "ORDER BY " + getOrderByClause(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters := getGoogleOrganicFiltersWhereStatement(query.Filters)

	resultSQLStatement := selectQuery + fromGoogleOrganicDocuments + staticWhereStatementForGoogleOrganic + whereConditionForFilters
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + channeAnalyticsLimit + ";"
	staticWhereParams := []interface{}{projectID, customerUrlPrefixes, query.From, query.To}
	return resultSQLStatement, staticWhereParams, responseSelectKeys, responseSelectMetrics
}

func getGoogleOrganicFiltersWhereStatement(filters []model.ChannelFilterV1) string {
	resultStatement := ""
	var filterValue string
	for index, filter := range filters {
		currentFilterStatement := ""
		if filter.LogicalOp == "" {
			filter.LogicalOp = "AND"
		}
		filterOperator := getOp(filter.Condition)
		if filter.Condition == model.ContainsOpStr || filter.Condition == model.NotContainsOpStr {
			filterValue = fmt.Sprintf("%%%s%%", filter.Value)
		} else {
			filterValue = filter.Value
		}
		currentFilterStatement = fmt.Sprintf("value->>'%s' %s '%s' ", filter.Property, filterOperator, filterValue)
		if index == 0 {
			resultStatement = " AND " + currentFilterStatement
		} else {
			resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement
}
