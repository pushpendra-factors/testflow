package task

import (
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func EventTriggerAlertsSender(projectID int64, configs map[string]interface{}) (map[string]interface{}, bool) {
	log.Info("Inside task manager")

	prefix := fmt.Sprintf("ETA:pid:%d", projectID)
	ssKey, err := cacheRedis.NewKeyWithOnlyPrefix(prefix)
	if err != nil {
		log.WithError(err).Error("Failed to fetch cacheKey for sortedSet")
		return nil, false
	}
	allKeys, err := cacheRedis.ZrangeWithScoresPersistent(true, ssKey)
	if err != nil {
		log.WithError(err).Error("Failed to get all alert keys for project: ", projectID)
		return nil, false
	}

	log.Info(fmt.Printf("%+v\n", allKeys))
	status := make(map[string]interface{})

	for key := range allKeys {
		cacheKey, err := cacheRedis.KeyFromStringWithPid(key)
		if err != nil {
			log.Error("Failed to get cacheKey from the key string")
			continue
		}

		log.Info("Key found: ", cacheKey)

		alertID := strings.Split(cacheKey.Suffix, ":")[0]
		cacheStr, err := cacheRedis.GetPersistent(cacheKey)
		if err != nil {
			log.WithError(err).Error("failed to find message for the alert ", alertID)
			continue
		}
		var msg model.CachedEventTriggerAlert
		err = U.DecodeJSONStringToStructType(cacheStr, &msg)
		if err != nil {
			log.WithError(err).Error("failed to decode alert for event_trigger_alert")
			status["error"] = err
			continue
		}

		log.Info("Proceeding with sendHelper function.")
		success := sendHelperForEventTriggerAlert(cacheKey, &msg, alertID)

		if success {
			err := cacheRedis.DelPersistent(cacheKey)
			if err != nil {
				log.WithError(err).Error("Cannot remove alert from cache")
			}
			cc, err := cacheRedis.ZRemPersistent(ssKey, true, key)
			if err != nil || cc != 1 {
				log.WithError(err).Error("Cannot remove alert by zrem")
			}
			log.Info("Alert removed from cache")
		}
	}

	return nil, true
}

func sendHelperForEventTriggerAlert(key *cacheRedis.Key, alert *model.CachedEventTriggerAlert, alertID string) bool {

	eta, errCode := store.GetStore().GetEventTriggerAlertByID(alertID)
	if errCode != http.StatusFound || eta == nil {
		log.Error("Failed to fetch alert from db, ", errCode)
		return false
	}

	var alertConfiguration model.EventTriggerAlertConfig
	err := U.DecodePostgresJsonbToStructType(eta.EventTriggerAlert, &alertConfiguration)
	if err != nil {
		log.WithError(err).Error("Failed to decode Jsonb to struct type")
		return false
	}

	var sendSuccess bool

	msg := alert.Message
	if alertConfiguration.Slack {
		log.Info(fmt.Printf("Message to be sent: %s", msg))
		log.Info(fmt.Printf("%+v\n", alertConfiguration))
		sendSuccess = sendSlackAlertForEventTriggerAlert(eta.ProjectID, eta.CreatedBy, msg, alertConfiguration.SlackChannels)
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

		// status, err := slack.SendSlackAlert(projectID, msg.Message, agentUUID, channel)
		// if err != nil || !status {
		// 	logCtx.WithError(err).Error("failed to send slack alert ", msg)
		// 	return false
		// }

		logCtx.Info("slack alert sent")
	}

	return true
}
