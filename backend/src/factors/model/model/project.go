package model

import (
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Project struct {
	ID   uint64 `gorm:"primary_key:true;" json:"id"`
	Name string `gorm:"not null;" json:"name"`
	// An index created on token.
	Token string `gorm:"size:32" json:"token"`
	// An index created on private_token.
	PrivateToken        string          `gorm:"size:32" json:"private_token"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	ProjectURI          string          `json:"project_uri"`
	TimeFormat          string          `json:"time_format"`
	DateFormat          string          `json:"date_format"`
	TimeZone            string          `json:"time_zone"`
	InteractionSettings postgres.Jsonb  `json:"interaction_settings"`
	JobsMetadata        *postgres.Jsonb `json:"jobs_metadata"`
}

const (
	JobsMetadataKeyNextSessionStartTimestamp = "next_session_start_timestamp"
	JobsMetadataColumnName                   = "jobs_metadata"
)

const DefaultProjectName = "My Project"

type InteractionSettings struct {
	UTMMappings map[string][]string `json:"utm_mapping"`
}

// DefaultMarketingPropertiesMap returns default query params and order (utm then qp) for various event properties.
func DefaultMarketingPropertiesMap() InteractionSettings {

	interactionSettings := InteractionSettings{}
	interactionSettings.UTMMappings = make(map[string][]string)
	interactionSettings.UTMMappings[U.EP_CAMPAIGN] = []string{U.QUERY_PARAM_UTM_PREFIX + "campaign", U.QUERY_PARAM_UTM_PREFIX + "campaign_name"}
	interactionSettings.UTMMappings[U.EP_CAMPAIGN_ID] = []string{U.QUERY_PARAM_UTM_PREFIX + "campaignid", U.QUERY_PARAM_UTM_PREFIX + "campaign_id"}

	interactionSettings.UTMMappings[U.EP_ADGROUP] = []string{U.QUERY_PARAM_UTM_PREFIX + "adgroup", U.QUERY_PARAM_UTM_PREFIX + "ad_group"}
	interactionSettings.UTMMappings[U.EP_ADGROUP_ID] = []string{U.QUERY_PARAM_UTM_PREFIX + "adgroupid", U.QUERY_PARAM_UTM_PREFIX + "adgroup_id", U.QUERY_PARAM_UTM_PREFIX + "ad_group_id"}
	interactionSettings.UTMMappings[U.EP_CREATIVE] = []string{U.QUERY_PARAM_UTM_PREFIX + "creative", U.QUERY_PARAM_UTM_PREFIX + "creative_id", U.QUERY_PARAM_UTM_PREFIX + "creativeid"}

	interactionSettings.UTMMappings[U.EP_SOURCE] = []string{U.QUERY_PARAM_UTM_PREFIX + "source"}
	interactionSettings.UTMMappings[U.EP_MEDIUM] = []string{U.QUERY_PARAM_UTM_PREFIX + "medium"}
	interactionSettings.UTMMappings[U.EP_KEYWORD] = []string{U.QUERY_PARAM_UTM_PREFIX + "keyword", U.QUERY_PARAM_UTM_PREFIX + "key_word"}
	interactionSettings.UTMMappings[U.EP_CONTENT] = []string{U.QUERY_PARAM_UTM_PREFIX + "content", U.QUERY_PARAM_UTM_PREFIX + "utm_content"}
	interactionSettings.UTMMappings[U.EP_GCLID] = []string{U.QUERY_PARAM_PROPERTY_PREFIX + "gclid", U.QUERY_PARAM_PROPERTY_PREFIX + "utm_gclid"}
	interactionSettings.UTMMappings[U.EP_FBCLIID] = []string{U.QUERY_PARAM_PROPERTY_PREFIX + "fbclid", U.QUERY_PARAM_PROPERTY_PREFIX + "utm_fbclid"}

	interactionSettings.UTMMappings[U.EP_TERM] = []string{U.QUERY_PARAM_UTM_PREFIX + "term"}
	interactionSettings.UTMMappings[U.EP_KEYWORD_MATCH_TYPE] = []string{U.QUERY_PARAM_UTM_PREFIX + "matchtype", U.QUERY_PARAM_UTM_PREFIX + "match_type"}

	return interactionSettings
}
