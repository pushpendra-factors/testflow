package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type IntegrationDocument struct {
	DocumentId        string          `json:"document_id"`
	ProjectID         int64           `json:"project_id"`
	CustomerAccountID string          `json:"customer_account_id"`
	Source            string          `json:"source"`
	DocumentType      int             `json:"document_type"`
	Timestamp         int64           `json:"timestamp"`
	Value             *postgres.Jsonb `json:"value"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}
