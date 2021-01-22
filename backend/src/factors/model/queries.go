package model

import (
	C "factors/config"
	"net/http"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const QueryTypeDashboardQuery = 1
const QueryTypeSavedQuery = 2

type Queries struct {
	// Composite primary key, id + project_id.
	ID uint64 `gorm:"primary_key:true" json:"id"`
	// Foreign key queries(project_id) ref projects(id).
	ProjectID     uint64         `gorm:"primary_key:true" json:"project_id"`
	Title         string         `gorm:"not null" json:"title"`
	Query         postgres.Jsonb `gorm:"not null" json:"query"`
	Type          int            `gorm:"not null; primary_key:true" json:"type"`
	IsDeleted     bool           `gorm:"not null;default:false" json:"is_deleted"`
	CreatedBy     string         `gorm:"type:varchar(255)" json:"created_by"`
	CreatedByName string         `gorm:"-" json:"created_by_name"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func CreateQuery(projectId uint64, query *Queries) (*Queries, int, string) {
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

func GetALLQueriesWithProjectId(projectID uint64) ([]Queries, int) {
	db := C.GetServices().Db

	queries := make([]Queries, 0, 0)
	err := db.Table("queries").Select("*").Where("project_id = ? AND is_deleted = ?", projectID, "false").Order("created_at DESC").Find(&queries).Error
	if err != nil {
		log.WithField("project_id", projectID).Error("Failed to fetch rows from queries table for project")
		return queries, http.StatusInternalServerError
	}
	if len(queries) == 0 {
		log.WithField("project_id", projectID).Error("No Saved Queries found")
		return queries, http.StatusFound

	}
	q, errCode := addCreatedByNameInQueries(queries, projectID)
	if errCode != http.StatusFound {
		// logging error but still sending the queries
		log.WithField("project_id", projectID).Error("could not update created " +
			"by name for queries")
		return queries, http.StatusFound
	}
	return q, http.StatusFound
}

// addCreatedByName Adds agent name in the query.CreatedByName
func addCreatedByNameInQueries(queries []Queries, projectID uint64) ([]Queries, int) {

	agentUUIDs := make([]string, 0, 0)
	for _, q := range queries {
		if q.CreatedBy != "" {
			agentUUIDs = append(agentUUIDs, q.CreatedBy)
		}
	}

	agents, errCode := GetAgentsByUUIDs(agentUUIDs)
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
func addCreatedByNameInQuery(query Queries) (Queries, int) {

	agentUUID := query.CreatedBy

	agent, errCode := GetAgentByUUID(agentUUID)
	if errCode != http.StatusFound {
		return query, errCode
	}
	query.CreatedByName = agent.FirstName + " " + agent.LastName
	return query, http.StatusFound
}

// GetDashboardQueryWithQueryId Get query of type DashboardQuery.
func GetDashboardQueryWithQueryId(projectID uint64, queryID uint64) (*Queries, int) {
	return getQueryWithQueryID(projectID, queryID, QueryTypeDashboardQuery)
}

// GetSavedQueryWithQueryId Get query of type SavedQuery.
func GetSavedQueryWithQueryId(projectID uint64, queryID uint64) (*Queries, int) {
	return getQueryWithQueryID(projectID, queryID, QueryTypeSavedQuery)
}

// GetQueryWithQueryId Get query by query id of any type.
func GetQueryWithQueryId(projectID uint64, queryID uint64) (*Queries, int) {
	return getQueryWithQueryID(projectID, queryID, 0)
}

func getQueryWithQueryID(projectID uint64, queryID uint64, queryType int) (*Queries, int) {
	db := C.GetServices().Db
	var query Queries
	var err error
	if queryType == 0 {
		err = db.Table("queries").Where("project_id = ? AND id=? AND is_deleted = ?",
			projectID, queryID, "false").Find(&query).Error
	} else {
		err = db.Table("queries").Where("project_id = ? AND id=? AND type=? AND is_deleted = ?",
			projectID, queryID, queryType, "false").Find(&query).Error
	}
	if err != nil {
		return &Queries{}, http.StatusNotFound
	}

	if queryType == QueryTypeSavedQuery {
		q, errCode := addCreatedByNameInQuery(query)
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

	var dashboardUnits []DashboardUnit
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
func DeleteQuery(projectID uint64, queryID uint64) (int, string) {
	return deleteQuery(projectID, queryID, 0)
}

// DeleteSavedQuery Deletes query of type QueryTypeSavedQuery.
func DeleteSavedQuery(projectID uint64, queryID uint64) (int, string) {
	return deleteQuery(projectID, queryID, QueryTypeSavedQuery)
}

// DeleteDashboardQuery Deletes query of type QueryTypeDashboardQuery.
func DeleteDashboardQuery(projectID uint64, queryID uint64) (int, string) {
	return deleteQuery(projectID, queryID, QueryTypeDashboardQuery)
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
		err = db.Model(&Queries{}).Where("id= ? AND project_id=?", queryID, projectID).
			Update(map[string]interface{}{"is_deleted": true}).Error
	} else {
		err = db.Model(&Queries{}).Where("id= ? AND project_id=? AND type=?", queryID, projectID, queryType).
			Update(map[string]interface{}{"is_deleted": true}).Error
	}
	if err != nil {
		return http.StatusInternalServerError, "Failed to delete saved query"
	}
	return http.StatusAccepted, ""
}

func UpdateSavedQuery(projectID uint64, queryID uint64, query *Queries) (*Queries, int) {
	db := C.GetServices().Db

	if queryID == 0 || query.Title == "" {
		return &Queries{}, http.StatusBadRequest
	}

	err := db.Model(&Queries{}).Where("project_id = ? AND id=? AND type=? AND is_deleted = ?", projectID, queryID, query.Type, "false").Update("title", query.Title).Error
	if err != nil {
		return &Queries{}, http.StatusInternalServerError
	}
	return query, http.StatusAccepted
}

func SearchQueriesWithProjectId(projectID uint64, searchString string) ([]Queries, int) {
	db := C.GetServices().Db

	var queries []Queries
	err := db.Table("queries").Where("project_id = ? AND title ILIKE ? AND is_deleted= ?", projectID, "%"+searchString+"%", "false").Find(&queries).Error
	if err != nil || len(queries) == 0 {
		return nil, http.StatusNotFound
	}
	return queries, http.StatusFound
}
