package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesQueriesForeignConstraints(query model.Queries) int {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, errCode := store.GetProject(query.ProjectID)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}

	if query.CreatedBy != "" {
		_, agentErrCode := store.GetAgentByUUID(query.CreatedBy)
		if agentErrCode != http.StatusFound {
			return http.StatusBadRequest
		}
	}
	return http.StatusOK
}

func (store *MemSQL) CreateQuery(projectId int64, query *model.Queries) (*model.Queries, int, string) {
	logFields := log.Fields{
		"project_id": projectId,
		"query":      query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	if projectId == 0 || query.Type == 0 {
		return nil, http.StatusBadRequest, "Invalid request"
	}
	if query.Type == 2 && query.CreatedBy == "" {
		return nil, http.StatusBadRequest, "Need agentUUID for saved query."
	}
	if query.Type == 2 && query.Title == "" {
		return nil, http.StatusBadRequest, "Need title for saved query."
	}

	query.ProjectID = projectId
	if errCode := store.satisfiesQueriesForeignConstraints(*query); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError, "Foreign constraints violation"
	}

	if err := db.Create(&query).Error; err != nil {
		errMsg := "Failed to insert query."
		log.WithFields(log.Fields{"Query": query,
			"project_id": projectId}).WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, errMsg
	}
	return query, http.StatusCreated, ""
}

// GetALLQueriesWithProjectId Get all queries for Saved Reports.
func (store *MemSQL) GetALLQueriesWithProjectId(projectID int64) ([]model.Queries, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	queries := make([]model.Queries, 0, 0)
	err := db.Table("queries").Select("*").
		Where("project_id = ? AND is_deleted = ?", projectID, false).
		Order("created_at DESC").Find(&queries).Error
	if err != nil {
		log.WithError(err).WithField("project_id", projectID).Error("Failed to fetch rows from queries table for project")
		return queries, http.StatusInternalServerError
	}
	if len(queries) == 0 {
		return queries, http.StatusFound
	}
	q, errCode := store.addCreatedByNameInQueries(queries, projectID)
	if errCode != http.StatusFound {
		// logging error but still sending the queries
		log.WithField("project_id", projectID).WithField("err_code", errCode).Error("could not update created " +
			"by name for queries")
		return queries, http.StatusFound
	}
	queriesResponse := make([]model.Queries, 0)
	for _, query := range q {
		exists := existsDashboardUnitForQueryID(projectID, query.ID)
		query.IsDashboardQuery = exists
		queriesResponse = append(queriesResponse, query)
	}
	return queriesResponse, http.StatusFound
}

// fetching deleted, non-deleted queries.
func (store *MemSQL) GetAllNonConvertedQueries(projectID int64) ([]model.Queries, int) {
	db := C.GetServices().Db

	queries := make([]model.Queries, 0, 0)
	err := db.Table("queries").Select("*").
		Where("project_id = ? AND converted = false ", projectID).
		Order("created_at DESC").Find(&queries).Error
	if err != nil {
		log.WithField("project_id", projectID).Error("Failed to fetch rows from queries table for project")
		return queries, http.StatusInternalServerError
	}
	if len(queries) == 0 {
		return queries, http.StatusFound
	}
	q, errCode := store.addCreatedByNameInQueries(queries, projectID)
	if errCode != http.StatusFound {
		// logging error but still sending the queries
		log.WithError(err).WithField("project_id", projectID).Error("could not update created " +
			"by name for queries")
		return queries, http.StatusFound
	}
	return q, http.StatusFound
}

// addCreatedByName Adds agent name in the query.CreatedByName
func (store *MemSQL) addCreatedByNameInQueries(queries []model.Queries, projectID int64) ([]model.Queries, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"queries":    queries,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	agentUUIDs := make([]string, 0, 0)
	for _, q := range queries {
		if q.CreatedBy != "" {
			agentUUIDs = append(agentUUIDs, q.CreatedBy)
		}
	}

	agents, errCode := store.GetAgentsByUUIDs(agentUUIDs)
	if errCode != http.StatusFound {
		log.WithField("project_id", projectID).WithField("err_code", errCode).Error("could not get agents for given agentUUIDs")
		return queries, errCode
	}

	agentUUIDsToName := make(map[string]string)
	agentUUIDsToEmail := make(map[string]string)

	for _, a := range agents {
		agentUUIDsToName[a.UUID] = a.FirstName + " " + a.LastName
		agentUUIDsToEmail[a.UUID] = a.Email
	}

	for i, _ := range queries {
		if _, exists := agentUUIDsToName[queries[i].CreatedBy]; exists {
			queries[i].CreatedByName = agentUUIDsToName[queries[i].CreatedBy]
			queries[i].CreatedByEmail = agentUUIDsToEmail[queries[i].CreatedBy]
		}
	}
	return queries, http.StatusFound
}

// addCreatedByName Adds agent name in the query.CreatedByName
func (store *MemSQL) addCreatedByNameInQuery(query model.Queries) (model.Queries, int) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	agentUUID := query.CreatedBy

	agent, errCode := store.GetAgentByUUID(agentUUID)
	if errCode != http.StatusFound {
		return query, errCode
	}
	query.CreatedByName = agent.FirstName + " " + agent.LastName
	return query, http.StatusFound
}

// GetDashboardQueryWithQueryId Get query of type DashboardQuery.
func (store *MemSQL) GetDashboardQueryWithQueryId(projectID int64, queryID int64) (*model.Queries, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.getQueryWithQueryID(projectID, queryID, model.QueryTypeDashboardQuery)
}

// GetSavedQueryWithQueryId Get query of type SavedQuery.
func (store *MemSQL) GetSavedQueryWithQueryId(projectID int64, queryID int64) (*model.Queries, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.getQueryWithQueryID(projectID, queryID, model.QueryTypeSavedQuery)
}

// GetQueryWithQueryId Get query by query id of any type.
func (store *MemSQL) GetQueryWithQueryId(projectID int64, queryID int64) (*model.Queries, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.getQueryWithQueryID(projectID, queryID, model.QueryTypeAllQueries)
}

// GetSixSignalQueryWithQueryID Get query by query id of type SixSignalQuery
func (store *MemSQL) GetSixSignalQueryWithQueryId(projectID int64, queryID int64) (*model.Queries, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.getQueryWithQueryID(projectID, queryID, model.QueryTypeSixSignalQuery)
}

func (store *MemSQL) GetQueryWithDashboardUnitIdString(projectID int64, dashboardUnitId int64) (*model.Queries, int) {
	logFields := log.Fields{
		"project_id":               projectID,
		"dashboard_unit_id_string": dashboardUnitId,
	}
	dashboardUnit, errCode := store.GetDashboardUnitByUnitID(projectID, dashboardUnitId)
	if errCode != http.StatusFound {
		return &model.Queries{}, http.StatusNotFound
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	var query model.Queries
	var err error
	err = db.Table("queries").Where("project_id = ? AND id_text=? AND is_deleted = ?",
		projectID, dashboardUnit.QueryId, false).Find(&query).Error
	if err != nil {
		return &model.Queries{}, http.StatusNotFound
	}
	return store.getQueryWithQueryID(projectID, query.ID, model.QueryTypeAllQueries)
}

func (store *MemSQL) GetQueryWithQueryIdString(projectID int64, queryIDString string) (*model.Queries, int) {
	logFields := log.Fields{
		"project_id":      projectID,
		"query_id_string": queryIDString,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	var query model.Queries
	var err error
	err = db.Table("queries").Where("project_id = ? AND id_text=? AND is_deleted = ?",
		projectID, queryIDString, false).Find(&query).Error
	if err != nil {
		return &model.Queries{}, http.StatusNotFound
	}
	return store.getQueryWithQueryID(projectID, query.ID, model.QueryTypeAllQueries)
}

func (store *MemSQL) getQueryWithQueryID(projectID int64, queryID int64, queryType int) (*model.Queries, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
		"query_type": queryType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	var query model.Queries
	var err error
	if queryType == model.QueryTypeAllQueries {
		err = db.Table("queries").Where("project_id = ? AND id=? AND is_deleted = ?",
			projectID, queryID, false).Find(&query).Error
	} else {
		err = db.Table("queries").Where("project_id = ? AND id=? AND type=? AND is_deleted = ?",
			projectID, queryID, queryType, false).Find(&query).Error
	}
	if err != nil {
		return &model.Queries{}, http.StatusNotFound
	}

	if queryType == model.QueryTypeSavedQuery {
		q, errCode := store.addCreatedByNameInQuery(query)
		if errCode != http.StatusFound {
			// logging error but still sending the queries
			log.WithField("project_id", projectID).WithField("query_id",
				queryID).WithField("err_code", errCode).Error("could not update created by name for queries")
			return &query, http.StatusFound
		}
		return &q, http.StatusFound
	}
	return &query, http.StatusFound
}

// existsDashboardUnitForQueryID checks if dashboard unit exists for given queryID
func existsDashboardUnitForQueryID(projectID int64, queryID int64) bool {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var dashboardUnits []model.DashboardUnit
	if err := db.Where("project_id = ? AND query_id = ? AND is_deleted = ?", projectID, queryID, false).
		Find(&dashboardUnits).Error; err != nil {
		log.WithError(err).Errorf("Failed to get dashboard units for projectID %d", projectID)
		// in case of failure allow delete
		return false
	}

	if len(dashboardUnits) == 0 {
		return false
	}
	return true
}

// existsDashboardUnitForQueryID checks if dashboard unit exists for given queryID
func (store *MemSQL) GetDashboardUnitForQueryID(projectID int64, queryID int64) []model.DashboardUnit {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var dashboardUnits []model.DashboardUnit
	if err := db.Where("project_id = ? AND query_id = ? AND is_deleted = ?", projectID, queryID, false).
		Find(&dashboardUnits).Error; err != nil {
		log.WithError(err).Errorf("Failed to get dashboard units for projectID %d", projectID)
		// in case of failure allow delete
		return nil
	}

	if len(dashboardUnits) == 0 {
		return nil
	}
	return dashboardUnits
}

// DeleteQuery To delete query of any type.
func (store *MemSQL) DeleteQuery(projectID int64, queryID int64) (int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return deleteQuery(projectID, queryID, 0)
}

// DeleteSavedQuery Deletes query of type QueryTypeSavedQuery.
func (store *MemSQL) DeleteSavedQuery(projectID int64, queryID int64) (int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return deleteQuery(projectID, queryID, model.QueryTypeSavedQuery)
}

// DeleteDashboardQuery Deletes query of type QueryTypeDashboardQuery.
func (store *MemSQL) DeleteDashboardQuery(projectID int64, queryID int64) (int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return deleteQuery(projectID, queryID, model.QueryTypeDashboardQuery)
}

func deleteQuery(projectID int64, queryID int64, queryType int) (int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
		"query_type": queryType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	if projectID == 0 {
		return http.StatusBadRequest, "Invalid project ID"
	}
	if queryID == 0 {
		return http.StatusBadRequest, "Invalid query ID"
	}
	if existsDashboardUnitForQueryID(projectID, queryID) {
		return http.StatusNotAcceptable, "Query in use: One or more dashboard widgets exists for this query"
	}

	var err error
	if queryType == 0 {
		// Delete any query irrespective of type.
		err = db.Model(&model.Queries{}).Where("id= ? AND project_id=?", queryID, projectID).
			Update(map[string]interface{}{"is_deleted": true}).Error
	} else {
		err = db.Model(&model.Queries{}).Where("id= ? AND project_id=? AND type=?", queryID, projectID, queryType).
			Update(map[string]interface{}{"is_deleted": true}).Error
	}
	if err != nil {
		return http.StatusInternalServerError, "Failed to delete saved query"
	}
	return http.StatusAccepted, ""
}

func (store *MemSQL) UpdateSavedQuery(projectID int64, queryID int64, query *model.Queries) (*model.Queries, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
		"query":      query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	if queryID == 0 || query.Title == "" {
		return &model.Queries{}, http.StatusBadRequest
	}

	if query.Type != 0 && query.Type != model.QueryTypeDashboardQuery && query.Type != model.QueryTypeSavedQuery {
		return &model.Queries{}, http.StatusBadRequest
	}
	// update allowed fields.
	updateFields := make(map[string]interface{}, 0)
	if query.Title != "" {
		updateFields["title"] = query.Title
	}

	if !U.IsEmptyPostgresJsonb(&query.Settings) {
		updateFields["settings"] = query.Settings
	}

	if query.Type == model.QueryTypeDashboardQuery || query.Type == model.QueryTypeSavedQuery {
		updateFields["type"] = query.Type
	}

	if !U.IsEmptyPostgresJsonb(&query.Query) {
		updateFields["query"] = query.Query
	}

	updateFields["locked_for_cache_invalidation"] = query.LockedForCacheInvalidation

	err := db.Model(&model.Queries{}).Where("project_id = ? AND id=? AND is_deleted = ?",
		projectID, queryID, false).Update(updateFields).Error
	if err != nil {
		return &model.Queries{}, http.StatusInternalServerError
	}
	return query, http.StatusAccepted
}

func (store *MemSQL) UpdateQueryIDsWithNewIDs(projectID int64, shareableURLs []string) int {
	logFields := log.Fields{
		"project_id":     projectID,
		"shareable_urls": shareableURLs,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db
	statusCode := http.StatusAccepted
	for _, idText := range shareableURLs {
		if err := db.Table("queries").Where("project_id = ? AND id_text = ? AND is_deleted = ?", projectID, idText, "false").
			Update("id_text", U.RandomStringForSharableQuery(50)).Error; err != nil {
			logCtx.Error("Failed to update query id_text: ", idText)
			statusCode = http.StatusPartialContent
		}
	}
	return statusCode
}

func (store *MemSQL) SearchQueriesWithProjectId(projectID int64, searchString string) ([]model.Queries, int) {
	logFields := log.Fields{
		"project_id":    projectID,
		"search_string": searchString,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var queries []model.Queries
	err := db.Table("queries").Where("project_id = ? AND title RLIKE ? AND is_deleted= ?", projectID, searchString, false).Find(&queries).Error
	if err != nil || len(queries) == 0 {
		return nil, http.StatusNotFound
	}
	return queries, http.StatusFound
}
