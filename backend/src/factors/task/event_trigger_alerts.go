package task

import (
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	slack "factors/slack_bot/handler"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	scanCount = 10000
	limit     = 10000
)

func EventTriggerAlertsSender(configs map[string]interface{}) (map[string]interface{}, bool) {
	log.Info("Inside task manager")
	prefix := "ETA:Alert:pid:*"
	allKeys, err := cacheRedis.ScanPersistent(prefix, scanCount, limit)
	if err != nil {
		log.Fatalf("Failed to get all alerts for project")
		return nil, false
	}
	status := make(map[string]interface{})

	for _, key := range allKeys {
		cacheStr, err := cacheRedis.GetPersistent(key)
		if err != nil {
			log.Fatalf("Failed to get alert from the cache")
			return nil, false
		}
		log.Info("Key found: ", key)

		counterKey := cacheRedis.Key{
			ProjectID: key.ProjectID,
			Prefix:    "ETA:Counter",
			Suffix:    key.Suffix,
		}

		log.Info(fmt.Printf("%+v\n", counterKey))

		var alert model.CachedEventTriggerAlert
		err = U.DecodeJSONStringToStructType(cacheStr, &alert)
		if err != nil {
			log.WithError(err).Errorf("failed to decode alert for event_trigger_alert: %s", alert.AlertID)
			status["error"] = err
			return status, false
		}

		log.Info("Proceeding with sendHelper function.")
		success := sendHelperForEventTriggerAlert(key, &counterKey, &alert)

		if success {
			err := cacheRedis.DelPersistent(key)
			if err != nil {
				log.WithError(err).Error("Cannot remove alert from cache")
			}
			log.Info("Alert removed from cache")
		}
	}

	return nil, true
}

func sendHelperForEventTriggerAlert(key, counterKey *cacheRedis.Key, alert *model.CachedEventTriggerAlert) bool {
	var alertConfiguration model.EventTriggerAlertConfig
	var sendSuccess bool

	eta, errCode := store.GetStore().GetEventTriggerAlertByID(alert.AlertID)
	if errCode != http.StatusFound {
		log.WithFields(log.Fields{"event_trigger_alert": alert, log.ErrorKey: errCode}).Error(
			"Failed to decode alert.")
		return false
	}
	err := U.DecodePostgresJsonbToStructType(eta.EventTriggerAlert, &alertConfiguration)
	if err != nil {
		log.WithFields(log.Fields{"event_trigger_alert": alert, log.ErrorKey: err}).Error(
			"Failed to decode alert.")
		return false
	}

	notify := alertConfiguration.Notifications
	if notify {
		msg :=  alert.Message
		if alertConfiguration.Slack {
			log.Info(fmt.Printf("Message to be sent: %s", msg))
			log.Info(fmt.Printf("%+v\n", alertConfiguration))
			sendSuccess = sendSlackAlertForEventTriggerAlert(eta.ProjectID, eta.CreatedBy, msg, alertConfiguration.SlackChannels)
		}
	}

	if sendSuccess {
		status, err := store.GetStore().UpdateEventTriggerAlertField(eta.ProjectID, eta.ID,
			map[string]interface{}{"last_alert_at": time.Now()})
		if status != http.StatusAccepted || err != nil {
			log.Fatalf("Failed to update db field")
		}
		log.Info("Alert update in db successful")
		return true
	}
	return false
}

func sendSlackAlertForEventTriggerAlert(projectID int64, agentUUID string, msg model.EventTriggerAlertMessage, Schannels *postgres.Jsonb) bool {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"agent_uuid": agentUUID,
	})
	var slackChannels []model.SlackChannel

	err := U.DecodePostgresJsonbToStructType(Schannels, &slackChannels)
	if err != nil {
		log.WithError(err).Error("failed to decode slack channels")
		return false
	}

	log.Info("Inside sendSlackAlert function")
	dryRunFlag := C.GetConfig().EnableDryRunAlerts
	if dryRunFlag {
		log.Info("Dry run mode enabled. No alerts will be sent")
		log.Info(msg, projectID)
		return false
	}

	for _, channel := range slackChannels {
		log.Info("Sending alert for slack channel ", channel)

		status, err := slack.SendSlackAlert(projectID, msg.Message, agentUUID, channel)
		if err != nil || !status {
			logCtx.WithError(err).Error("failed to send slack alert ", msg)
			return false
		}

		logCtx.Info("slack alert sent")
	}

	return true
}
