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
	ProjectID uint64         `gorm:"primary_key:true" json:"project_id"`
	Title     string         `gorm:"not null" json:"title"`
	Query     postgres.Jsonb `gorm:"not null" json:"query"`
	Type      int            `gorm:"not null; primary_key:true" json:"type"`
	IsDeleted bool           `gorm:"not null;default:false" json:"is_deleted"`
	CreatedBy string         `gorm:"type:varchar(255);primary_key:true;" json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
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

func GetSavedQueriesWithProjectId(projectID uint64) ([]Queries, int) {
	db := C.GetServices().Db

	queries := make([]Queries, 0, 0)
	err := db.Table("queries").Select("*").Where("project_id = ? AND type = ? AND is_deleted = ?", projectID, QueryTypeSavedQuery, "false").Find(&queries).Error
	if err != nil {
		log.WithField("project_id", projectID).Error("Failed to fetch rows from queries table for project")
		return queries, http.StatusInternalServerError
	}
	return queries, http.StatusFound
}

func DeleteSavedQuery(projectID uint64, queryID uint64) (int, string) {
	db := C.GetServices().Db
	if projectID == 0 {
		return http.StatusBadRequest, "Invalid project ID"
	}
	if queryID == 0 {
		return http.StatusBadRequest, "Invalid query ID"
	}
	err := db.Model(&Queries{}).Where("id= ? AND project_id=? AND type=?", queryID, projectID, QueryTypeSavedQuery).Update("is_deleted", true).Error
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
func GetQueryWithQueryId(projectID uint64, queryID uint64) (*Queries, int) {
	db := C.GetServices().Db

	var query Queries
	err := db.Table("queries").Where("project_id = ? AND id=? AND type=? AND is_deleted = ?", projectID, queryID, QueryTypeSavedQuery, "false").Find(&query).Error

	if err != nil {
		return &Queries{}, http.StatusNotFound
	}
	return &query, http.StatusFound
}
func SearchQueriesWithProjectId(projectID uint64, searchString string) ([]Queries, int) {
	db := C.GetServices().Db

	var queries []Queries
	err := db.Table("queries").Where("project_id = ? AND title LIKE ? AND is_deleted= ?", projectID, "%"+searchString+"%", "false").Find(&queries).Error
	if err != nil || len(queries) == 0 {
		return nil, http.StatusNotFound
	}
	return queries, http.StatusFound
}
