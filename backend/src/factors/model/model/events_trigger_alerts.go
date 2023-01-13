package model

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	// DeliveryOptions
	SLACK                              = "slack"
	WEBHOOK                            = "webhook"
	tableNameforAlerts                 = "ETA"
	cachedIndexEventTiggerAlert        = "Alert"
	cachedIndexEventTiggerAlertCounter = "Counter"
	cacheExpiry                        = 0
	cacheCounterExpiry                 = 24 * 60 * 60
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
	AlertID   string
	Timestamp time.Time
	Message   EventTriggerAlertMessage
}

type CacheEventTriggerAlertCounter int

type EventTriggerAlertMessage struct {
	Title           string
	Event           string
	MessageProperty string
	Message         string
}

func SetCacheForEventTriggerAlert(projectId int64, userId string, cacheETA *CachedEventTriggerAlert) error {
	logCtx := log.WithField("project_id", projectId).WithField("user_id", userId)
	if projectId == 0 || userId == "" {
		logCtx.Error("Invalid project or user id")
		return errors.New("invalid project or user id")
	}
	if cacheETA == nil {
		logCtx.Error("Nil cache event on setCacheUserLastEventTriggerAlert")
		return errors.New("nil cache event")
	}

	cacheETAJson, err := json.Marshal(cacheETA)
	if err != nil {
		logCtx.Error("Failed cache event trigger alert json marshal.")
		return err
	}

	date := time.Now().UTC().Format(U.DATETIME_FORMAT_YYYYMMDD)
	key, err := GetEventTriggerAlertCacheKey(projectId, userId, cacheETA.AlertID, date)
	if err != nil || key == nil {
		logCtx.WithError(err).Error("Failed at GetEventTriggerAlertCacheKey.")
		return err
	}

	err = cacheRedis.SetPersistent(key, string(cacheETAJson), float64(cacheExpiry))
	if err != nil {
		logCtx.WithError(err).Error("Failed to set Cache for EventTriggerAlert.")
	}

	log.Info("Adding to cache successful.")
	return err
}

func SetCacheCounterForEventTriggerAlert(projectId int64, userId, alertId string, counterPresent bool) error {
	logCtx := log.WithField("project_id", projectId).WithField("user_id", userId)

	date := time.Now().UTC().Format(U.DATETIME_FORMAT_YYYYMMDD)
	key, err := GetEventTriggerAlertCacheCounterKey(projectId, userId, alertId, date)
	if err != nil || key == nil {
		logCtx.WithError(err).Error("Failed at GetEventTriggerAlertCacheCounterKey.")
		return err
	}

	if counterPresent {
		_, err := cacheRedis.IncrPersistentBatch(key)
		if err != nil {
			log.WithError(err).Error("Cache cannot be updated")
		}
	} else {
		err = cacheRedis.SetPersistent(key, "1", float64(cacheCounterExpiry))
		if err != nil {
			logCtx.WithError(err).Error("Failed to setCacheUserLastEventTriggerAlert.")
		}
	}
	log.Info("Adding to cache successful.")
	return err
}

func GetEventTriggerAlertCacheKey(projectId int64, userId, alertId, date string) (*cacheRedis.Key, error) {

	suffix := fmt.Sprintf("uid:%s:aid:%s:%s", userId, alertId, date)
	prefix := fmt.Sprintf("%s:%s", tableNameforAlerts, cachedIndexEventTiggerAlert)

	log.Info("Fetching redisKey, inside GetEventTriggerAlertCacheKey.")

	key, err := cacheRedis.NewKey(projectId, prefix, suffix)
	if err != nil || key == nil {
		log.WithError(err).Error("cacheKey NewKey function failure")
		return nil, err
	}

	return key, err
}

func GetEventTriggerAlertCacheCounterKey(projectId int64, userId, alertId, date string) (*cacheRedis.Key, error) {

	suffix := fmt.Sprintf("uid:%s:aid:%s:%s", userId, alertId, date)
	prefix := fmt.Sprintf("%s:%s", tableNameforAlerts, cachedIndexEventTiggerAlertCounter)

	log.Info("Fetching redisKey, inside GetEventTriggerAlertCacheCounterKey.")

	key, err := cacheRedis.NewKey(projectId, prefix, suffix)
	if err != nil || key == nil {
		log.WithError(err).Error("cacheKey NewKey function failure for counter")
		return nil, err
	}

	return key, err
}
