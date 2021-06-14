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
	ChannelGroupRules   postgres.Jsonb  `json:"channel_group_rules"`
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
	interactionSettings.UTMMappings[U.EP_CAMPAIGN_ID] = []string{U.QUERY_PARAM_UTM_PREFIX + "campaignid", U.QUERY_PARAM_UTM_PREFIX + "campaign_id",
		U.QUERY_PARAM_UTM_PREFIX + "hsa_cam"}

	interactionSettings.UTMMappings[U.EP_ADGROUP] = []string{U.QUERY_PARAM_UTM_PREFIX + "adgroup", U.QUERY_PARAM_UTM_PREFIX + "ad_group"}
	interactionSettings.UTMMappings[U.EP_ADGROUP_ID] = []string{U.QUERY_PARAM_UTM_PREFIX + "adgroupid", U.QUERY_PARAM_UTM_PREFIX + "adgroup_id",
		U.QUERY_PARAM_UTM_PREFIX + "ad_group_id", U.QUERY_PARAM_UTM_PREFIX + "hsa_grp"}

	interactionSettings.UTMMappings[U.EP_AD] = []string{U.QUERY_PARAM_UTM_PREFIX + "ad"}
	interactionSettings.UTMMappings[U.EP_AD_ID] = []string{U.QUERY_PARAM_UTM_PREFIX + "ad_id", U.QUERY_PARAM_UTM_PREFIX + "adid",
		U.QUERY_PARAM_UTM_PREFIX + "hsa_ad"}

	interactionSettings.UTMMappings[U.EP_CREATIVE] = []string{U.QUERY_PARAM_UTM_PREFIX + "creative", U.QUERY_PARAM_UTM_PREFIX + "creative_id",
		U.QUERY_PARAM_UTM_PREFIX + "creativeid"}

	interactionSettings.UTMMappings[U.EP_SOURCE] = []string{U.QUERY_PARAM_UTM_PREFIX + "source"}
	interactionSettings.UTMMappings[U.EP_MEDIUM] = []string{U.QUERY_PARAM_UTM_PREFIX + "medium"}
	interactionSettings.UTMMappings[U.EP_KEYWORD] = []string{U.QUERY_PARAM_UTM_PREFIX + "keyword", U.QUERY_PARAM_UTM_PREFIX + "key_word"}
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

func GetCacheProjectIDByToken(token string) (uint64, int) {
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

	projectID, err := strconv.ParseUint(projectIDAsString, 10, 64)
	if err != nil {
		logCtx.WithError(err).Error("Failed to convert project_id to uint64 in GetCacheProjectIDByToken.")
		return 0, http.StatusInternalServerError
	}

	return projectID, http.StatusFound
}

func SetCacheProjectIDByToken(token string, projectID uint64) int {
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
