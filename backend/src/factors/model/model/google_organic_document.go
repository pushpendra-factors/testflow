package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type GoogleOrganicDocument struct {
	ID        string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	ProjectID uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	URLPrefix string          `gorm:"primary_key:true;auto_increment:false" json:"url_prefix"`
	Timestamp int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	Value     *postgres.Jsonb `json:"value"`
	Type      int             `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type GoogleOrganicLastSyncInfo struct {
	ProjectId     uint64 `json:"project_id"`
	URLPrefix     string `json:"url_prefix"`
	RefreshToken  string `json:"refresh_token"`
	LastTimestamp int64  `json:"last_timestamp"`
	Type          int64  `json:"type"`
}
type GoogleOrganicLastSyncInfoPayload struct {
	ProjectId uint64 `json:"project_id"`
}

var GoogleOrganicTypes = []int64{CombinedLevelData, PageLevelData}

const (
	GoogleOrganicSpecificError = "Failed in search console with the error."
	CombinedLevelData          = int64(1)
	PageLevelData              = int64(2)
)
