package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type SDKButtonElementAttributes struct {
	DisplayText string `json:"display_text"`
	Class       string `json:"class"`
	Id          string `json:"id"`
	Rel         string `json:"rel"`
	Role        string `json:"role"`
	Target      string `json:"target"`
	Href        string `json:"href"`
	Media       string `json:"media"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Timestamp   int64  `json:"timestamp"`
}

type SDKButtonElementAttributesPayload struct {
	DisplayName       string                     `json:"display_name"`
	ElementType       string                     `json:"element_type"`
	ElementAttributes SDKButtonElementAttributes `json:"element_attributes"`
}

type ClickableElements struct {
	ProjectID         int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	Id                string          `gorm:"primary_key:true;type:varchar(255)" json:"id"`
	DisplayName       string          `json:"display_name"`
	ElementType       string          `json:"element_type"`
	ElementAttributes *postgres.Jsonb `json:"element_attributes"`
	ClickCount        uint            `json:"click_count"`
	Enabled           bool            `json:"enabled"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type SDKButtonElementAttributesResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
