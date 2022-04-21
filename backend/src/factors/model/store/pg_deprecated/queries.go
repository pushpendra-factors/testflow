package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) CreateQuery(projectId uint64, query *model.Queries) (*model.Queries, int, string) {
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
	if err := db.Create(&query).Error; err != nil {
		errMsg := "Failed to insert query."
		log.WithFields(log.Fields{"Query": query,
			"project_id": projectId}).WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, errMsg
	}

	return query, http.StatusCreated, ""
}

// GetALLQueriesWithProjectId Get all queries for Saved Reports.
func (pg *Postgres) GetALLQueriesWithProjectId(projectID uint64) ([]model.Queries, int) {
	db := C.GetServices().Db

	queries := make([]model.Queries, 0, 0)
	err := db.Table("queries").Select("*").
		Where("project_id = ? AND is_deleted = ?", projectID, "false").
		Order("created_at DESC").Find(&queries).Error
	if err != nil {
		log.WithField("project_id", projectID).Error("Failed to fetch rows from queries table for project")
		return queries, http.StatusInternalServerError
	}
	if len(queries) == 0 {
		return queries, http.StatusFound
	}
	q, errCode := pg.addCreatedByNameInQueries(queries, projectID)
	if errCode != http.StatusFound {
		// logging error but still sending the queries
		log.WithField("project_id", projectID).Error("could not update created " +
			"by name for queries")
		return queries, http.StatusFound
	}
	return q, http.StatusFound
}

func (pg *Postgres) GetAllNonConvertedQueries(projectID uint64) ([]model.Queries, int) {
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
	q, errCode := pg.addCreatedByNameInQueries(queries, projectID)
	if errCode != http.StatusFound {
		// logging error but still sending the queries
		log.WithField("project_id", projectID).Error("could not update created " +
			"by name for queries")
		return queries, http.StatusFound
	}
	return q, http.StatusFound
}

// addCreatedByName Adds agent name in the query.CreatedByName
func (pg *Postgres) addCreatedByNameInQueries(queries []model.Queries, projectID uint64) ([]model.Queries, int) {

	agentUUIDs := make([]string, 0, 0)
	for _, q := range queries {
		if q.CreatedBy != "" {
			agentUUIDs = append(agentUUIDs, q.CreatedBy)
		}
	}

	agents, errCode := pg.GetAgentsByUUIDs(agentUUIDs)
	if errCode != http.StatusFound {
		log.WithField("project_id", projectID).Error("could not get agents for given agentUUIDs")
		return queries, errCode
	}

	agentUUIDsToName := make(map[string]string)

	for _, a := range agents {
		agentUUIDsToName[a.UUID] = a.FirstName + " " + a.LastName
	}

	for i, _ := range queries {
		if _, exists := agentUUIDsToName[queries[i].CreatedBy]; exists {
			queries[i].CreatedByName = agentUUIDsToName[queries[i].CreatedBy]
		}
	}
	return queries, http.StatusFound
}

// addCreatedByName Adds agent name in the query.CreatedByName
func (pg *Postgres) addCreatedByNameInQuery(query model.Queries) (model.Queries, int) {

	agentUUID := query.CreatedBy

	agent, errCode := pg.GetAgentByUUID(agentUUID)
	if errCode != http.StatusFound {
		return query, errCode
	}
	query.CreatedByName = agent.FirstName + " " + agent.LastName
	return query, http.StatusFound
}

// GetDashboardQueryWithQueryId Get query of type DashboardQuery.
func (pg *Postgres) GetDashboardQueryWithQueryId(projectID uint64, queryID uint64) (*model.Queries, int) {
	return pg.getQueryWithQueryID(projectID, queryID, model.QueryTypeDashboardQuery)
}

// GetSavedQueryWithQueryId Get query of type SavedQuery.
func (pg *Postgres) GetSavedQueryWithQueryId(projectID uint64, queryID uint64) (*model.Queries, int) {
	return pg.getQueryWithQueryID(projectID, queryID, model.QueryTypeSavedQuery)
}

// GetQueryWithQueryId Get query by query id of any type.
func (pg *Postgres) GetQueryWithQueryId(projectID uint64, queryID uint64) (*model.Queries, int) {
	return pg.getQueryWithQueryID(projectID, queryID, 0)
}

func (pg *Postgres) GetQueryWithQueryIdString(projectID uint64, queryIDString string) (*model.Queries, int) {
	db := C.GetServices().Db
	var query model.Queries
	var err error
	err = db.Table("queries").Where("project_id = ? AND id_text=? AND is_deleted = ?",
		projectID, queryIDString, false).Find(&query).Error
	if err != nil {
		return &model.Queries{}, http.StatusNotFound
	}
	return pg.getQueryWithQueryID(projectID, query.ID, model.QueryTypeAllQueries)
}

func (pg *Postgres) getQueryWithQueryID(projectID uint64, queryID uint64, queryType int) (*model.Queries, int) {
	db := C.GetServices().Db
	var query model.Queries
	var err error
	if queryType == model.QueryTypeAllQueries {
		err = db.Table("queries").Where("project_id = ? AND id=? AND is_deleted = ?",
			projectID, queryID, "false").Find(&query).Error
	} else {
		err = db.Table("queries").Where("project_id = ? AND id=? AND type=? AND is_deleted = ?",
			projectID, queryID, queryType, "false").Find(&query).Error
	}
	if err != nil {
		return &model.Queries{}, http.StatusNotFound
	}

	if queryType == model.QueryTypeSavedQuery {
		q, errCode := pg.addCreatedByNameInQuery(query)
		if errCode != http.StatusFound {
			// logging error but still sending the queries
			log.WithField("project_id", projectID).WithField("query_id",
				queryID).Error("could not update created by name for queries")
			return &query, http.StatusFound
		}
		return &q, http.StatusFound
	}
	return &query, http.StatusFound
}

// existsDashboardUnitForQueryID checks if dashboard unit exists for given queryID
func existsDashboardUnitForQueryID(projectID uint64, queryID uint64) bool {
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

// DeleteQuery To delete query of any type.
func (pg *Postgres) DeleteQuery(projectID uint64, queryID uint64) (int, string) {
	return deleteQuery(projectID, queryID, 0)
}

// DeleteSavedQuery Deletes query of type QueryTypeSavedQuery.
func (pg *Postgres) DeleteSavedQuery(projectID uint64, queryID uint64) (int, string) {
	return deleteQuery(projectID, queryID, model.QueryTypeSavedQuery)
}

// DeleteDashboardQuery Deletes query of type QueryTypeDashboardQuery.
func (pg *Postgres) DeleteDashboardQuery(projectID uint64, queryID uint64) (int, string) {
	return deleteQuery(projectID, queryID, model.QueryTypeDashboardQuery)
}

func deleteQuery(projectID uint64, queryID uint64, queryType int) (int, string) {
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

func (pg *Postgres) UpdateSavedQuery(projectID uint64, queryID uint64, query *model.Queries) (*model.Queries, int) {
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

	err := db.Model(&model.Queries{}).Where("project_id = ? AND id=? AND is_deleted = ?",
		projectID, queryID, "false").Update(updateFields).Error
	if err != nil {
		return &model.Queries{}, http.StatusInternalServerError
	}
	return query, http.StatusAccepted
}

func (pg *Postgres) UpdateQueryIDsWithNewIDs(projectID uint64, shareableURLs []string) int {
	db := C.GetServices().Db
	statusCode := http.StatusAccepted
	for _, idText := range shareableURLs {
		if err := db.Table("queries").Where("project_id = ? AND id_text = ? AND is_deleted = ?", projectID, idText, "false").
			Update("id_text", U.RandomStringForSharableQuery(50)).Error; err != nil {
			statusCode = http.StatusPartialContent
		}
	}
	return statusCode
}

func (pg *Postgres) SearchQueriesWithProjectId(projectID uint64, searchString string) ([]model.Queries, int) {
	db := C.GetServices().Db

	var queries []model.Queries
	err := db.Table("queries").Where("project_id = ? AND title ILIKE ? AND is_deleted= ?", projectID, "%"+searchString+"%", "false").Find(&queries).Error
	if err != nil || len(queries) == 0 {
		return nil, http.StatusNotFound
	}
	return queries, http.StatusFound
}
