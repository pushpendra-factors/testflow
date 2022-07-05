package model

import "time"

type GroupRelationship struct {
	ProjectID        int64     `json:"project_id"`
	LeftGroupNameID  int       `json:"left_group_name_id"`
	LeftGroupUserID  string    `json:"left_group_user_id"`
	RightGroupNameID int       `json:"right_group_name_id"`
	RightGroupUserID string    `json:"group_group_user_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
