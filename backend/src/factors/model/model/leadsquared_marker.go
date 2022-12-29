package model

import "time"

type LeadsquaredMarker struct {
	ProjectID   int64     `json:"project_id"`
	Delta       int64     `json:"delta"`
	Document    string    `json:"document"`
	IndexNumber int       `json:"index_number"`
	NoOfRetries int       `json:"no_of_retries"`
	Tag         string    `json:"tag"`
	IsDone		bool	  `json:"is_done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
