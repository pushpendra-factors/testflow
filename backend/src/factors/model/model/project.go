package model

import (
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type Project struct {
	ID             int64  `gorm:"primary_key:true;" json:"id"`
	Name           string `gorm:"not null;" json:"name"`
	ProfilePicture string `json:"profile_picture"`
	// An index created on token.
	Token string `gorm:"size:32" json:"token"`
	// An index created on private_token.
	PrivateToken          string          `gorm:"size:32" json:"private_token"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	ProjectURI            string          `json:"project_uri"`
	TimeFormat            string          `json:"time_format"`
	DateFormat            string          `json:"date_format"`
	TimeZone              string          `json:"time_zone"`
	InteractionSettings   postgres.Jsonb  `json:"interaction_settings"`
	SalesforceTouchPoints postgres.Jsonb  `json:"salesforce_touch_points"`
	HubspotTouchPoints    postgres.Jsonb  `json:"hubspot_touch_points"`
	JobsMetadata          *postgres.Jsonb `json:"jobs_metadata"`
	ChannelGroupRules     postgres.Jsonb  `json:"channel_group_rules"`
}

type ProjectString struct {
	ID             string `gorm:"primary_key:true;" json:"id"`
	Name           string `gorm:"not null;" json:"name"`
	ProfilePicture string `json:"profile_picture"`
	// An index created on token.
	Token string `gorm:"size:32" json:"token"`
	// An index created on private_token.
	PrivateToken          string          `gorm:"size:32" json:"private_token"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	ProjectURI            string          `json:"project_uri"`
	TimeFormat            string          `json:"time_format"`
	DateFormat            string          `json:"date_format"`
	TimeZone              string          `json:"time_zone"`
	InteractionSettings   postgres.Jsonb  `json:"interaction_settings"`
	SalesforceTouchPoints postgres.Jsonb  `json:"salesforce_touch_points"`
	HubspotTouchPoints    postgres.Jsonb  `json:"hubspot_touch_points"`
	JobsMetadata          *postgres.Jsonb `json:"jobs_metadata"`
	ChannelGroupRules     postgres.Jsonb  `json:"channel_group_rules"`
}

const (
	JobsMetadataKeyNextSessionStartTimestamp = "next_session_start_timestamp"
	JobsMetadataColumnName                   = "jobs_metadata"
	/*LastModifiedTimeRef                      = "LAST_MODIFIED_TIME_REF"
	TouchPointPropertyValueAsProperty        = "Property"
	TouchPointPropertyValueAsConstant        = "Constant"

	TouchPointRuleTypeEmails   = "Emails"
	TouchPointRuleTypeMeetings = "Meetings"
	TouchPointRuleTypeCalls    = "Calls"
	TouchPointRuleTypeForms    = "Form_Submissions"*/
)

const DefaultProjectName = "My Project"

type InteractionSettings struct {
	UTMMappings map[string][]string `json:"utm_mapping"`
}

type SalesforceTouchPoints struct {
	TouchPointRules map[string][]SFTouchPointRule `json:"sf_touch_point_rules"`
}

type SFTouchPointRule struct {
	Filters           []TouchPointFilter                 `json:"filters"`
	TouchPointTimeRef string                             `json:"touch_point_time_ref"`
	PropertiesMap     map[string]TouchPointPropertyValue `json:"properties_map"`
}

/*type TouchPointFilter struct {
	Property string `json:"pr"`
	// Entity: user or event.
	Entity string `json:"en"`
	// Type: categorical or numerical
	Type      string `json:"ty"`
	Operator  string `json:"op"`
	Value     string `json:"va"`
	LogicalOp string `json:"lop"`
}*/

// DefaultSalesforceTouchPointsRules returns default query params and order (utm then qp) for various event properties.
func DefaultSalesforceTouchPointsRules() SalesforceTouchPoints {

	rules := SalesforceTouchPoints{}
	rules.TouchPointRules = make(map[string][]SFTouchPointRule)
	return rules
}

type HubspotTouchPoints struct {
	TouchPointRules map[string][]HSTouchPointRule `json:"hs_touch_point_rules"`
}

type HSTouchPointRule struct {
	RuleType          string                             `json:"rule_type"`
	Filters           []TouchPointFilter                 `json:"filters"`
	TouchPointTimeRef string                             `json:"touch_point_time_ref"`
	PropertiesMap     map[string]TouchPointPropertyValue `json:"properties_map"`
}

/*type TouchPointPropertyValue struct {
	Type  string `json:"ty"`
	Value string `json:"va"`
}*/

// DefaultHubspotTouchPointsRules returns default query params and order (utm then qp) for various event properties.
func DefaultHubspotTouchPointsRules() HubspotTouchPoints {

	rules := HubspotTouchPoints{}
	rules.TouchPointRules = make(map[string][]HSTouchPointRule)
	return rules
}

// DefaultMarketingPropertiesMap returns default query params and order (utm then qp) for various event properties.
func DefaultMarketingPropertiesMap() InteractionSettings {

	interactionSettings := InteractionSettings{}
	interactionSettings.UTMMappings = make(map[string][]string)
	interactionSettings.UTMMappings[U.EP_CAMPAIGN] = []string{U.QUERY_PARAM_UTM_PREFIX + "campaign", U.QUERY_PARAM_UTM_PREFIX + "campaign_name"}
	interactionSettings.UTMMappings[U.EP_CAMPAIGN_ID] = []string{U.QUERY_PARAM_UTM_PREFIX + "campaignid", U.QUERY_PARAM_UTM_PREFIX + "campaign_id",
		U.QUERY_PARAM_UTM_PREFIX + "hsa_cam", "$qp_hsa_cam"}

	interactionSettings.UTMMappings[U.EP_ADGROUP] = []string{U.QUERY_PARAM_UTM_PREFIX + "adgroup", U.QUERY_PARAM_UTM_PREFIX + "ad_group"}
	interactionSettings.UTMMappings[U.EP_ADGROUP_ID] = []string{U.QUERY_PARAM_UTM_PREFIX + "adgroupid", U.QUERY_PARAM_UTM_PREFIX + "adgroup_id",
		U.QUERY_PARAM_UTM_PREFIX + "ad_group_id", U.QUERY_PARAM_UTM_PREFIX + "hsa_grp", "$qp_hsa_grp"}

	interactionSettings.UTMMappings[U.EP_AD] = []string{U.QUERY_PARAM_UTM_PREFIX + "ad"}
	interactionSettings.UTMMappings[U.EP_AD_ID] = []string{U.QUERY_PARAM_UTM_PREFIX + "ad_id", U.QUERY_PARAM_UTM_PREFIX + "adid",
		U.QUERY_PARAM_UTM_PREFIX + "hsa_ad", "$qp_hsa_ad"}

	interactionSettings.UTMMappings[U.EP_CREATIVE] = []string{U.QUERY_PARAM_UTM_PREFIX + "creative", U.QUERY_PARAM_UTM_PREFIX + "creative_id",
		U.QUERY_PARAM_UTM_PREFIX + "creativeid"}

	interactionSettings.UTMMappings[U.EP_SOURCE] = []string{U.QUERY_PARAM_UTM_PREFIX + "source"}
	interactionSettings.UTMMappings[U.EP_MEDIUM] = []string{U.QUERY_PARAM_UTM_PREFIX + "medium"}
	interactionSettings.UTMMappings[U.EP_KEYWORD] = []string{U.QUERY_PARAM_UTM_PREFIX + "keyword", U.QUERY_PARAM_UTM_PREFIX + "key_word", "$qp_hsa_kw"}
	interactionSettings.UTMMappings[U.EP_CONTENT] = []string{U.QUERY_PARAM_UTM_PREFIX + "content", U.QUERY_PARAM_UTM_PREFIX + "utm_content"}
	interactionSettings.UTMMappings[U.EP_GCLID] = []string{U.QUERY_PARAM_PROPERTY_PREFIX + "gclid", U.QUERY_PARAM_PROPERTY_PREFIX + "utm_gclid",
		U.QUERY_PARAM_PROPERTY_PREFIX + "wbraid", U.QUERY_PARAM_PROPERTY_PREFIX + "gbraid"}
	interactionSettings.UTMMappings[U.EP_FBCLID] = []string{U.QUERY_PARAM_PROPERTY_PREFIX + "fbclid", U.QUERY_PARAM_PROPERTY_PREFIX + "utm_fbclid"}

	interactionSettings.UTMMappings[U.EP_TERM] = []string{U.QUERY_PARAM_UTM_PREFIX + "term"}
	interactionSettings.UTMMappings[U.EP_KEYWORD_MATCH_TYPE] = []string{U.QUERY_PARAM_UTM_PREFIX + "matchtype", U.QUERY_PARAM_UTM_PREFIX + "match_type"}

	return interactionSettings
}

// DefaultURLPropertiesToMarketingPropertiesMap returns default query params map reverse of DefaultMarketingPropertiesMap
func DefaultURLPropertiesToMarketingPropertiesMap() map[string]string {

	urlToEventPropMap := make(map[string]string)
	interactionSettings := DefaultMarketingPropertiesMap()
	for key, value := range interactionSettings.UTMMappings {
		for _, v := range value {
			urlToEventPropMap[v] = key
		}
	}
	return urlToEventPropMap
}

func ValidateChannelGroupRules(rules postgres.Jsonb) bool {
	var channelPropertyRules []ChannelPropertyRule
	err := U.DecodePostgresJsonbToStructType(&rules, &channelPropertyRules)
	if err != nil {
		return false
	}
	for _, rule := range channelPropertyRules {
		if rule.Channel == "" || len(rule.Conditions) == 0 {
			return false
		}
		for _, filter := range rule.Conditions {
			if filter.Condition == "" || filter.Property == "" || filter.LogicalOp == "" {
				return false
			}
		}
	}
	return true
}
func getCacheKeyForProjectIDByToken(token string) (*cacheRedis.Key, error) {
	return cacheRedis.NewKeyWithProjectUID(token, "projects_token_id", "")
}

func GetCacheProjectIDByToken(token string) (int64, int) {
	logCtx := log.WithField("token", token)

	if token == "" {
		logCtx.Error("Invalid params on GetCacheProjectIDByToken.")
		return 0, http.StatusInternalServerError
	}

	key, err := getCacheKeyForProjectIDByToken(token)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key GetCacheProjectIDByToken")
		return 0, http.StatusInternalServerError
	}

	projectIDAsString, err := cacheRedis.Get(key)
	if err != nil {
		if err == redis.ErrNil {
			return 0, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to GetCacheProjectIDByToken.")
		return 0, http.StatusInternalServerError
	}

	projectID, err := strconv.ParseInt(projectIDAsString, 10, 64)
	if err != nil {
		logCtx.WithError(err).Error("Failed to convert project_id to uint64 in GetCacheProjectIDByToken.")
		return 0, http.StatusInternalServerError
	}

	return projectID, http.StatusFound
}

func SetCacheProjectIDByToken(token string, projectID int64) int {
	logCtx := log.WithField("token", token).WithField("project_id", projectID)

	if token == "" || projectID == 0 {
		logCtx.Error("Invalid params on SetCacheProjectIDByToken.")
		return http.StatusInternalServerError
	}

	key, err := getCacheKeyForProjectIDByToken(token)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key for SetCacheProjectIDByToken.")
		return http.StatusInternalServerError
	}

	var expiryInSecs float64 = 60 * 60 * 24 // one day
	err = cacheRedis.Set(key, fmt.Sprintf("%d", projectID), expiryInSecs)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache on SetCacheProjectIDByToken")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
