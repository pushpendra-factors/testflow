package model

import "time"

type ReplicationMetadata struct {
	TableName string    `json:"table_name"`
	LastRunAt time.Time `json:"last_run_at"`
	Count     uint64    `json:"count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
