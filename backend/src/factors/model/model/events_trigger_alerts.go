package model

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	"fmt"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	// DeliveryOptions
	SLACK              = "slack"
	WEBHOOK            = "webhook"
	tableNameforAlerts = "ETA"
	counterIndex       = "Counter"
	cacheExpiry        = 0
	cacheCounterExpiry = 24 * 60 * 60

	// cachekey structure = ETA:pid:<project_id>:<alert_id>:<UnixTime>
	// cacheCounterKey structure = ETA:Counter:pid:<project_id>:<alert_id>:<YYYYMMDD>
	// sortedset key structure = ETA:pid:<project_id>
)

type EventTriggerAlert struct {
	ID                string          `gorm:"column:id; type:uuid; default:uuid_generate_v4()" json:"id"`
	ProjectID         int64           `gorm:"column:project_id; primary_key:true" json:"project_id"`
	Title             string          `gorm:"column:title; not null" json:"title"`
	EventTriggerAlert *postgres.Jsonb `json:"event_trigger_alert"`
	CreatedBy         string          `gorm:"column:created_by" json:"created_by"`
	LastAlertAt       time.Time       `json:"last_alert_at"`
	CreatedAt         time.Time       `gorm:"column:created_at; autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"column:updated_at; autoUpdateTime" json:"updated_at"`
	IsDeleted         bool            `gorm:"column:is_deleted; not null; default:false" json:"is_deleted"`
}

type EventTriggerAlertConfig struct {
	Title           string          `json:"title"`
	Event           string          `json:"event"`
	Filter          []QueryProperty `json:"filter"`
	Message         string          `json:"message"`
	MessageProperty *postgres.Jsonb `json:"message_property"`
	RepeatAlerts    bool            `json:"repeat_alerts"`
	AlertLimit      int64           `json:"alert_limit"`
	Notifications   bool            `json:"notifications"`
	Slack           bool            `json:"slack"`
	SlackChannels   *postgres.Jsonb `json:"slack_channels"`
	Webhook         bool            `json:"webhook"`
	WebhookURL      string          `json:"url"`
}

type EventTriggerAlertInfo struct {
	ID                string                   `json:"id"`
	Title             string                   `json:"title"`
	DeliveryOptions   string                   `json:"delivery_options"`
	EventTriggerAlert *EventTriggerAlertConfig `json:"event_alert"`
}

type CachedEventTriggerAlert struct {
	Message EventTriggerAlertMessage
}

type EventTriggerAlertMessage struct {
	Title           string
	Event           string
	MessageProperty string
	Message         string
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
	prefix := tableNameforAlerts

	log.Info("Fetching redisKey, inside GetEventTriggerAlertCacheKey.")

	key, err := cacheRedis.NewKey(projectId, prefix, suffix)
	if err != nil || key == nil {
		log.WithError(err).Error("cacheKey NewKey function failure")
		return nil, err
	}

	return key, err
}

func GetEventTriggerAlertCacheCounterKey(projectId int64, alertId, date string) (*cacheRedis.Key, error) {

	suffix := fmt.Sprintf("%s:%s", alertId, date)
	prefix := fmt.Sprintf("%s:%s", tableNameforAlerts, counterIndex)

	log.Info("Fetching redisKey, inside GetEventTriggerAlertCacheKey.")

	key, err := cacheRedis.NewKey(projectId, prefix, suffix)
	if err != nil || key == nil {
		log.WithError(err).Error("cacheKey NewKey function failure")
		return nil, err
	}

	return key, err
}
