package model

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	"fmt"
	"time"

	U "factors/util"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	// DeliveryOptions
	SLACK               = "slack"
	WEBHOOK             = "webhook"
	prefixNameforAlerts = "ETA"
	TEAMS               = "teams"
	counterIndex        = "Counter"
	cacheExpiry         = 7 * 24 * 60 * 60
	cacheCounterExpiry  = 24 * 60 * 60
	EventLevelAccount   = "account"
	EventLevelUser      = "user"
	Paused              = "paused"   //Internal status if the difference between failure is greater than success of the alert by 72 hours
	Active              = "active"   //Default internal status
	Disabled            = "disabled" //Internal status if the failures from the poison queue are not resolved for 72 more hours
	ETA_DOMAIN_GROUP_USER_ID = "ignore_eta_d_g_uid" // additonal property which needs to be cached for account urls

	// cachekey structure = ETA:pid:<project_id>:<alert_id>:<UnixTime>
	// cacheCounterKey structure = ETA:Counter:pid:<project_id>:<alert_id>:<YYYYMMDD>
	// sortedset key structure = ETA:pid:<project_id>
	// coolDownKeyCounter structure = ETA:CoolDown:pid:<project_id>:<alert_id>:<prop>:<value>:....:
	// failure sorted set key structure = ETA:Fail:pid:<project_id>
	// failure key = <fail_point>:ETA:pid:<project_id>:<alert_id>:<UnixTime>
	// 		-> fail_point = Slack/WH/Teams
	// Poison queue sorted set cache key = ETA:Poison:pid:<project_id>
)

type EventTriggerAlert struct {
	ID                       string          `gorm:"column:id; type:uuid; default:uuid_generate_v4()" json:"id"`
	ProjectID                int64           `gorm:"column:project_id; primary_key:true" json:"project_id"`
	Title                    string          `gorm:"column:title; not null" json:"title"`
	EventTriggerAlert        *postgres.Jsonb `json:"event_trigger_alert"`
	CreatedBy                string          `gorm:"column:created_by" json:"created_by"`
	SlackChannelAssociatedBy string          `gorm:"column:slack_channel_associated_by" json:"slack_channel_associated_by"`
	TeamsChannelAssociatedBy string          `gorm:"column:teams_channel_associated_by" json:"teams_channel_associated_by"`
	ParagonMetadata          *postgres.Jsonb `gorm:"column:paragon_metadata" json:"paragon_metadata"`
	LastAlertAt              time.Time       `json:"last_alert_at"`
	CreatedAt                time.Time       `gorm:"column:created_at; autoCreateTime" json:"created_at"`
	UpdatedAt                time.Time       `gorm:"column:updated_at; autoUpdateTime" json:"updated_at"`
	LastFailDetails          *postgres.Jsonb `gorm:"column:last_fail_details" json:"last_fail_details"`
	InternalStatus           string          `gorm:"column:internal_status; default:'active'" json:"internal_status"`
	IsDeleted                bool            `gorm:"column:is_deleted; not null; default:false" json:"is_deleted"`
}

type EventTriggerAlertConfig struct {
	Title               string          `json:"title"`
	EventLevel          string          `json:"event_level"`
	Event               string          `json:"event"`
	Filter              []QueryProperty `json:"filter"`
	Message             string          `json:"message"`
	MessageProperty     *postgres.Jsonb `json:"message_property"`
	DontRepeatAlerts    bool            `json:"repeat_alerts"`
	CoolDownTime        int64           `json:"cool_down_time"`
	BreakdownProperties *postgres.Jsonb `json:"breakdown_properties"`
	SetAlertLimit       bool            `json:"notifications"`
	AlertLimit          int64           `json:"alert_limit"`
	Slack               bool            `json:"slack"`
	SlackChannels       *postgres.Jsonb `json:"slack_channels"`
	SlackMentions       *postgres.Jsonb `json:"slack_mentions"`
	SlackFieldsTag      []string        `json:"slack_fields_tag"`
	IsHyperlinkDisabled bool            `json:"is_hyperlink_disabled"`
	Webhook             bool            `json:"webhook"`
	Secret              string          `json:"secret"`
	WebhookURL          string          `json:"url"`
	Teams               bool            `json:"teams"`
	TeamsChannelsConfig *postgres.Jsonb `json:"teams_channels_config"`
}

type AlertInfo struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	DeliveryOptions string          `json:"delivery_options"`
	Status          string          `json:"status"`
	LastFailDetails *postgres.Jsonb `json:"last_fail_details"`
	Alert           *postgres.Jsonb `json:"alert"`
	Type            string          `json:"type"`
	CreatedAt       time.Time       `json:"created_at"`
}

type CachedEventTriggerAlert struct {
	Message   EventTriggerAlertMessage
	FieldTags map[string]string
}

type EventTriggerAlertMessage struct {
	Title           string
	Event           string
	MessageProperty U.PropertiesMap
	Message         string
}

type EventTriggerWebhook struct {
	Title           string          `json:"title"`
	Event           string          `json:"event"`
	MessageProperty *postgres.Jsonb `json:"message_property"`
	Message         string          `json:"message"`
	Url             string          `json:"url"`
	Secret          string          `json:"secret"`
}

type MessagePropMapStruct struct {
	DisplayName string
	PropValue   interface{}
}

type LastFailDetails struct {
	FailTime time.Time `json:"fail_time"`
	FailedAt []string  `json:"failed_at"`
	Details  []string  `json:"details"`
}

var ValidAlertTagsForHubspotOwners = map[string]string{
	"#contact_owner": "$hubspot_contact_contact_owner_fs_",
	"#deal_owner":    "$hubspot_deal_hubspot_owner_id",
	"#company_owner": "$hubspot_company_hubspot_owner_id",
}

var SlackErrorStates = map[string]string{
	"channel_not_found":      "Channel not found. Please check the channel is not archived.",
	"is_archived":            "The channel has been archived.",
	"message_limit_exceeded": "Members on this team are sending too many messages.",
	"msg_too_long":           "Message text is too long. Consider shortening it.",
	"no_text":                "No message text provided.",
	"rate_limited":           "Application has posted too many message. Try using more filters to prevent spamming.",
	"too_many_attachments":   "Too many attachments were provided with this message. A maximum of 100 attachments are allowed on a message.",
	"access_denied":          "Access to a resource specified in the request is denied.",
	"account_inactive":       "Authentication token is for a deleted user or workspace when using a bot token. Please re-authenticate.",
	"invalid_auth":           "Authentication cannot be validated",
	"missing_scope":          "The token used is not granted the specific scope permissions required to complete this request.",
	"token_expired":          "Authentication token has expired",
	"token_revoked":          "Authentication token is for a deleted user or workspace or the app has been removed when using a user token.",
	"request_timeout":        "The method was called via a POST request, but the POST data was either missing or truncated.",
}

var TeamsErrorStates = map[string]string{
	"Unauthorized":      "Please re-integrate teams.",
	"Forbidden":         "Please re-integrate teams with work account.",
	"Too Many Requests": "Too many messages in a short while. Try refining the alerts filter",
}

func SetCacheForEventTriggerAlert(key *cacheRedis.Key, cacheETA *CachedEventTriggerAlert) error {
	if cacheETA == nil {
		log.Error("Nil cache event on setCacheUserLastEventTriggerAlert")
		return errors.New("nil cache event")
	}

	cacheETAJson, err := json.Marshal(cacheETA)
	if err != nil {
		log.Error("Failed cache event trigger alert json marshal.")
		return err
	}

	err = cacheRedis.SetPersistent(key, string(cacheETAJson), float64(cacheExpiry))
	if err != nil {
		log.WithError(err).Error("Failed to set Cache for EventTriggerAlert.")
	}

	log.Info("Adding to cache successful.")
	return err
}

func GetEventTriggerAlertCacheKey(projectId, timestamp int64, alertID string) (*cacheRedis.Key, error) {

	suffix := fmt.Sprintf("%s:%d", alertID, timestamp)
	prefix := prefixNameforAlerts

	key, err := cacheRedis.NewKey(projectId, prefix, suffix)
	if err != nil || key == nil {
		log.WithError(err).Error("cacheKey NewKey function failure")
		return nil, err
	}

	return key, err
}

func GetEventTriggerAlertCacheCounterKey(projectId int64, alertId, date string) (*cacheRedis.Key, error) {

	suffix := fmt.Sprintf("%s:%s", alertId, date)
	prefix := fmt.Sprintf("%s:%s", prefixNameforAlerts, counterIndex)

	log.Info("Fetching redisKey, inside GetEventTriggerAlertCacheKey.")

	key, err := cacheRedis.NewKey(projectId, prefix, suffix)
	if err != nil || key == nil {
		log.WithError(err).Error("cacheKey NewKey function failure")
		return nil, err
	}

	return key, err
}
