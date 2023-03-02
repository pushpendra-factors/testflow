package main

import (
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	slack "factors/slack_bot/handler"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultAppName := "event_trigger_alerts"
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:              *primaryDatastore,
		RedisHostPersistent:           *redisHostPersistent,
		RedisPortPersistent:           *redisPortPersistent,
		SentryDSN:                     *sentryDSN,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	db := C.GetServices().Db
	defer db.Close()

	conf := make(map[string]interface{})
	projectIDs, _ := store.GetStore().GetAllProjectIDs()
	for _, projectID := range projectIDs {
		status, success := EventTriggerAlertsSender(projectID, conf)
		log.Info(status)
		if status["err"] != nil || !success {
			log.Error("Event Trigger Alert job failing for projectID: ", projectID)
		}
	}
}

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

	wetRun := true
	if wetRun {
		for _, channel := range slackChannels {
			log.Info("Sending alert for slack channel ", channel)

			status, err := slack.SendSlackAlert(projectID, getSlackMsgBlock(msg), agentUUID, channel)
			if err != nil || !status {
				logCtx.WithError(err).Error("failed to send slack alert ", msg)
				return false
			}

			logCtx.Info("slack alert sent")
		}
	} else {
		log.Info("Dry run mode enabled. No alerts will be sent")
		log.Info("*****", msg, projectID)
		return true
	}

	return true
}

func returnSlackMessage(actualmsg string) string {
	template := fmt.Sprintf(`
		[
			{
				"type": "section",
				"text": {
					"type": "plain_text",
					"text": "%s",
					"emoji": true
				}
			}
		]
	`, actualmsg)
	return template
}

func getPropsBlock(propMap U.PropertiesMap) string {
	var propBlock string
	for key, prop := range propMap {
		if(prop == ""){
			prop = "<nil>"
		}
		propBlock += fmt.Sprintf(
			`{
				"type": "section",
				"fields": [
					{
						"type": "mrkdwn",
						"text": "%s"
					},
					{
						"type": "plain_text",
						"text": "%v",
					}
				]
			},
			{
				"type": "divider"
			},`, key, strings.Replace(fmt.Sprintf("%v", prop), "\"", "", -1))
	}
	return propBlock
}

func getSlackMsgBlock(msg model.EventTriggerAlertMessage) string {

	propBlock := getPropsBlock(msg.MessageProperty)

	mainBlock := fmt.Sprintf(`[
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "%s\n*%s*\n"
			}
		},
		%s
		{
			"type": "actions",
			"elements": [
				{
					"type": "button",
					"text": {
						"type": "plain_text",
						"text": "Know more",
						"emoji": true
					},
					"value": "click_me_123",
					"url": "https://app.factors.ai/profiles/people",
					"action_id": "button-action"
				}
			]
		}
	]`, strings.Replace(msg.Title, "\"", "", -1), strings.Replace(msg.Message, "\"", "", -1), propBlock)

	return mainBlock
}
