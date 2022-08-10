package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type AdsImport struct {
	ID                 string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	ProjectID          int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	LastProcessedIndex *postgres.Jsonb `json:"last_processed_index"`
	Status             bool            `json:"status"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type LastProcessedAdsImport struct {
	LineNo  int `json:"line_no"`
	ChunkNo int `json:"chunk_no"`
}
