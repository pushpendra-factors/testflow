package model

import "time"

const ALL_BOARDS_FOLDER = "All Boards"

type DashboardFolders struct {
	Id              string    `json:"id"`
	Name            string    `json:"name"`
	ProjectId       int64     `json:"project_id"`
	IsDeleted       bool      `json:"is_deleted"`
	IsDefaultFolder bool      `json:"can_be_deleted"` //IsDefaultFolder: true for All Boards folder, in other case false
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type DashboardFoldersRequestPayload struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	ProjectId   int64  `json:"project_id"`
	DashboardId int64  `json:"dashboard_id"`
}

type UpdatableDashboardFolder struct {
	Name string `json:"name"`
}
