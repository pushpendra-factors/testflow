package model

import "time"

type ProjectAgentMapping struct {
	// Composite primary key with project_id and agent_uuid
	AgentUUID string `gorm:"primary_key:true;type:varchar(255)" json:"agent_uuid"`
	ProjectID int64  `gorm:"primary_key:true" json:"project_id"`

	// Foreign key constraints added in creation script
	// project_id -> projects(id)
	// agent_uuid -> agents(uuid)
	// invited_by -> agents(uuid)

	Role uint64 `json:"role"`

	// Created as pointer to allow storing NULL in db
	InvitedBy *string `gorm:"type:varchar(255)" json:"invited_by"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectAgentMappingString struct {
	// Composite primary key with project_id and agent_uuid
	AgentUUID string `gorm:"primary_key:true;type:varchar(255)" json:"agent_uuid"`
	ProjectID string `gorm:"primary_key:true" json:"project_id"`

	// Foreign key constraints added in creation script
	// project_id -> projects(id)
	// agent_uuid -> agents(uuid)
	// invited_by -> agents(uuid)

	Role uint64 `json:"role"`

	// Created as pointer to allow storing NULL in db
	InvitedBy *string `gorm:"type:varchar(255)" json:"invited_by"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	AGENT                  = 1
	ADMIN                  = 2
	MAX_AGENTS_PER_PROJECT = 500
)
