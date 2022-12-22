package model

import (
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

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

type CaptureClickResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`

	// EventID - ID of event created after enabling.
	EventID string `json:"event_id,omitempty"`
}

type CaptureClickPayload struct {
	// Click Payload.
	DisplayName       string          `json:"display_name"`
	ElementType       string          `json:"element_type"`
	ElementAttributes U.PropertiesMap `json:"element_attributes"`

	// Track Payload.
	UserID          string          `json:"user_id"`
	EventProperties U.PropertiesMap `json:"event_properties"`
	UserProperties  U.PropertiesMap `json:"user_properties"`
	Timestamp       int64           `json:"timestamp"`
	UpdatedAt       *time.Time      `json:"-"`
}

func AddAllowedElementAttributes(projectID int64,
	receivedAttributes U.PropertiesMap, addToMap *U.PropertiesMap) {

	logCtx := log.WithField("project_id", projectID).
		WithField("received_attributes", receivedAttributes).
		WithField("add_to", addToMap)

	// Note: Not adding $prefix as these are event properties
	// added to event created by us, conflicts are not possible.
	var allowedAttributesAsProperties = map[string]bool{
		"display_text": true,
		"element_type": true,
		"class":        true,
		"id":           true,
		"rel":          true,
		"role":         true,
		"target":       true,
		"href":         true,
		"media":        true,
		"type":         true,
		"name":         true,
	}

	isUnsupportedAttributeSeen := false
	for key := range receivedAttributes {
		if _, exists := allowedAttributesAsProperties[key]; exists {
			(*addToMap)[key] = receivedAttributes[key]
		} else {
			isUnsupportedAttributeSeen = true
		}
	}

	if isUnsupportedAttributeSeen {
		logCtx.Error("Received unsupported on click capture.")
	}
}
