package model

import "time"

type SegmentFolder struct {
	Id	string 	`json:"id"`
	Name string	`json:"name"`
	ProjectId       int64     `json:"project_id"`
	FolderType		string	  `json:"folder_type"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SegmentFolderPayload struct {
	Name	string	`json:"name"`
}

type MoveSegmentFolderItemPayload struct {
	FolderID	string	`json:"folder_id"`
}