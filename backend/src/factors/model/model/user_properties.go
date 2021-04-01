package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"factors/config"
	"factors/util"
)

type UserProperties struct {
	// Composite primary key with project_id, user_id and random uuid.
	ID string `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	// (project_id, user_id) -> users(project_id, id)
	ProjectId uint64 `gorm:"primary_key:true;" json:"project_id"`
	UserId    string `gorm:"primary_key:true;" json:"user_id"`

	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Properties       postgres.Jsonb `json:"properties"`
	UpdatedTimestamp int64          `gorm:"not null;default:0" json:"updated_timestamp"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// UserPropertiesMeta is a map for customer_user_id to IdentifyMeta
type UserPropertiesMeta map[string]IdentifyMeta

// IdentifyMeta holds data for overwriting customer_user_id
type IdentifyMeta struct {
	Timestamp int64  `json:"timestamp"`
	PageURL   string `json:"page_url,omitempty"`
	Source    string `json:"source"`
}

type SessionUserProperties struct {
	// Meta
	UserID                string
	SessionEventTimestamp int64

	// Current event user properties.
	EventUserProperties *postgres.Jsonb

	// Properties
	SessionCount         uint64
	SessionPageCount     float64
	SessionPageSpentTime float64
}

// indexed hubspot user property.
const UserPropertyHubspotContactLeadGUID = "$hubspot_contact_lead_guid"
const MaxUsersForPropertiesMerge = 100

var ErrDifferentEmailSeen error = errors.New("different_email_seen_for_customer_user_id")
var ErrDifferentPhoneNoSeen error = errors.New("different_phone_no_seen_for_customer_user_id")

func FillLocationUserProperties(properties *util.PropertiesMap, clientIP string) error {
	geo := config.GetServices().GeoLocation

	// ClientIP unavailable.
	if clientIP == "" {
		return fmt.Errorf("invalid IP, failed adding geolocation properties")
	}

	city, err := geo.City(net.ParseIP(clientIP))
	if err != nil {
		log.WithFields(log.Fields{"clientIP": clientIP}).WithError(err).Error(
			"Failed to get city information from geodb")
		return err
	}

	// Using en -> english name.
	if countryName, ok := city.Country.Names["en"]; ok && countryName != "" {
		if c, ok := (*properties)[util.UP_COUNTRY]; !ok || c == "" {
			(*properties)[util.UP_COUNTRY] = countryName
		}
	}

	if cityName, ok := city.City.Names["en"]; ok && cityName != "" {
		if c, ok := (*properties)[util.UP_CITY]; !ok || c == "" {
			(*properties)[util.UP_CITY] = cityName
		}
	}

	return nil
}

// GetDecodedUserPropertiesIdentifierMetaObject gets the identifier meta data from the user properties
func GetDecodedUserPropertiesIdentifierMetaObject(existingUserProperties *map[string]interface{}) (*UserPropertiesMeta, error) {
	metaObj := make(UserPropertiesMeta)
	if existingUserProperties == nil {
		return &metaObj, nil
	}

	intMetaObj, exists := (*existingUserProperties)[util.UP_META_OBJECT_IDENTIFIER_KEY]
	if !exists {
		return &metaObj, nil
	}

	metaObjMap, ok := intMetaObj.(map[string]interface{})
	if !ok {
		err := json.Unmarshal([]byte(util.GetPropertyValueAsString(intMetaObj)), &metaObj)
		if err != nil {
			log.WithError(err).Errorf("Failed to get meta data from user properties")
		}
		return &metaObj, err
	}

	var enMetaObj []byte
	enMetaObj, err := json.Marshal(metaObjMap)
	if err != nil {
		log.WithError(err).Errorf("Failed to encode meta data from user properties")
		return &metaObj, err
	}

	err = json.Unmarshal(enMetaObj, &metaObj)
	if err != nil {
		log.WithError(err).Errorf("Failed to unmarshal meta data from user properties")
	}

	return &metaObj, err
}
