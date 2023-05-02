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

	slack "factors/integration/slack"
	teams "factors/integration/ms_teams"
	webhook "factors/webhooks"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	SortedSetKeyPrefix     = "ETA"
	FailureSortedSetPrefix = "ETA:Fail"
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

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")

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
		PrimaryDatastore:    *primaryDatastore,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		SentryDSN:           *sentryDSN,
	}
	defaultHealthcheckPingID := C.HealthcheckEventTriggerAlertPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
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
	finalStatus := make(map[string]interface{})
	success := true
	successCount, failureCount := 0, 0
	projectIDs, _ := store.GetStore().GetAllProjectIDs()
	for _, projectID := range projectIDs {
		projectSuccess := false
		successCount, failureCount, projectSuccess = EventTriggerAlertsSender(projectID, conf)
		if !projectSuccess {
			log.Error("Event Trigger Alert job failing for projectID: ", projectID)
		}
		success = success && projectSuccess
		if successCount > 0 {
			finalStatus[fmt.Sprintf("Success-%v", projectID)] = successCount
		}
		if failureCount > 0 {
			finalStatus[fmt.Sprintf("Failure-%v", projectID)] = failureCount
		}
	}
	if !success {
		C.PingHealthcheckForFailure(healthcheckPingID, finalStatus)
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, finalStatus)
	}
}

func getSortedSetCacheKey(prefix string, projectId int64) (*cacheRedis.Key, error) {
	pre := fmt.Sprintf("%s:pid:%d", prefix, projectId)
	key, err := cacheRedis.NewKeyWithOnlyPrefix(pre)
	if err != nil {
		log.WithError(err).Error("Cannot get redis key")
		return nil, err
	}
	return key, err
}

func EventTriggerAlertsSender(projectID int64, configs map[string]interface{}) (int, int, bool) {

	ok := int(0)

	ssKey, err := getSortedSetCacheKey(SortedSetKeyPrefix, projectID)
	if err != nil {
		log.WithError(err).Error("Failed to fetch cacheKey for sortedSet")
		return -1, -1, false
	}

	allKeys, err := cacheRedis.ZrangeWithScoresPersistent(true, ssKey)
	if err != nil {
		log.WithError(err).Error("Failed to get all alert keys for project: ", projectID)
		return -1, -1, false
	}

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
			continue
		}

		log.Info("Proceeding with sendHelper function.")
		success := sendHelperForEventTriggerAlert(cacheKey, &msg, alertID)

		if success {
			err = cacheRedis.DelPersistent(cacheKey)
			if err != nil {
				log.WithError(err).Error("Cannot remove alert from cache")
			}
			ok++
		}
		cc, err := cacheRedis.ZRemPersistent(ssKey, true, key)
		if err != nil || cc != 1 {
			log.WithError(err).Error("Cannot remove alert by zrem")
		}
	}
	return ok, len(allKeys) - ok, ok == len(allKeys)
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

	slackSuccess := true
	teamsSuccess := true
	whSuccess := true

	msg := alert.Message
	if alertConfiguration.Slack {
		slackSuccess = sendSlackAlertForEventTriggerAlert(eta.ProjectID, eta.SlackChannelAssociatedBy, msg, alertConfiguration.SlackChannels)
		if !slackSuccess {
			err := AddKeyToFailureSet(key, eta.ProjectID, "Slack")
			if err != nil {
				log.WithError(err).Error("failed to put key in FailureSortedSet")
			}
		}
	}
	if alertConfiguration.Teams {
		teamsSuccess = sendTeamsAlertForEventTriggerAlert(eta.ProjectID, eta.CreatedBy, msg, alertConfiguration.TeamsChannelsConfig)
		if !teamsSuccess {
			err := AddKeyToFailureSet(key, eta.ProjectID, "Teams")
			if err != nil {
				log.WithError(err).Error("failed to put key in FailureSortedSet")
			}
		}
	}
	if alertConfiguration.Webhook {
		response, err := webhook.DropWebhook(alertConfiguration.WebhookURL, alertConfiguration.Secret, alert.Message)
		if err != nil {
			log.WithFields(log.Fields{"alert_id": alertID, "server_response": response}).
				WithError(err).Error("Webhook failure")
		}
		log.Info(fmt.Printf("Webhook dropped for alert: %s. RESPONSE: %+v", alertID, response))
		stat := response["status"]
		if stat != "ok" {
			log.Error("Error details: ", stat, response)
			err := AddKeyToFailureSet(key, eta.ProjectID, "WH")
			if err != nil {
				log.WithError(err).Error("failed to put key in FailureSortedSet")
			}
			whSuccess = false
		}
	}

	if slackSuccess || whSuccess || teamsSuccess {
		status, err := store.GetStore().UpdateEventTriggerAlertField(eta.ProjectID, eta.ID,
			map[string]interface{}{"last_alert_at": time.Now()})
		if status != http.StatusAccepted || err != nil {
			log.Fatalf("Failed to update db field")
		}
	}

	return slackSuccess && whSuccess && teamsSuccess
}

func AddKeyToFailureSet(key *cacheRedis.Key, projectID int64, failPoint string) error {
	failureKey, err := getSortedSetCacheKey(FailureSortedSetPrefix, projectID)
	if err != nil {
		log.WithError(err).Error("failed to fetch sorted set key for failure set")
		return err
	}

	val, err := key.Key()
	if err != nil {
		log.WithError(err).Error("cannot find str value for cache key")
		return err
	}

	failureSS := cacheRedis.SortedSetKeyValueTuple{
		Key:   failureKey,
		Value: fmt.Sprintf("%s:%s", failPoint, val),
	}

	_, err = cacheRedis.ZincrPersistentBatch(true, failureSS)
	if err != nil {
		log.WithError(err).Error("failed to update failureSortedSet")
		return err
	}

	return nil
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

	wetRun := true
	if wetRun {
		for _, channel := range slackChannels {

			status, err := slack.SendSlackAlert(projectID, getSlackMsgBlock(msg), agentUUID, channel)
			if err != nil || !status {
				logCtx.WithError(err).Error("failed to send slack alert ", msg)
				return false
			}
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
	for i := 0; i < len(propMap); i++ {
		pp := propMap[fmt.Sprintf("%d", i)]
		var mp model.MessagePropMapStruct
		if pp != nil {
			trans, ok := pp.(map[string]interface{})
			if !ok {
				log.Warn("cannot convert interface to map[string]interface{} type")
				continue
			}
			err := U.DecodeInterfaceMapToStructType(trans, &mp)
			if err != nil {
				log.Warn("cannot convert interface map to struct type")
				continue
			}
		}

		key := mp.DisplayName
		prop := mp.PropValue
		if prop == "" {
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
						"type": "mrkdwn",
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
func getPropsBlockV2(propMap U.PropertiesMap) string {

	var propBlock string
	count := 0
	for i := 0; i < len(propMap); i++ {
		pp := propMap[fmt.Sprintf("%d", i)]
		var mp model.MessagePropMapStruct
		if pp != nil {
			trans, ok := pp.(map[string]interface{})
			if !ok {
				log.Warn("cannot convert interface to map[string]interface{} type")
				continue
			}
			err := U.DecodeInterfaceMapToStructType(trans, &mp)
			if err != nil {
				log.Warn("cannot convert interface map to struct type")
				continue
			}
		}
		var key1, key2 string
		var prop1, prop2 interface{}
		count++

		if count == 1 {
			key1 = mp.DisplayName
			prop1 = mp.PropValue
			if prop1 == "" {
				prop1 = "<nil>"
			}
		} else if count == 2 {
			key2 = mp.DisplayName
			prop2 = mp.PropValue
			if prop2 == "" {
				prop2 = "<nil>"
			}
		}
		// as slack template support only 2 columns hence adding check for count 2
		if count == 2 {
			propBlock += fmt.Sprintf(
				`{
					"type": "section",
					"fields": [
						{
							"type": "mrkdwn",
							"text": "%s \n %v"
						},
						{
							"type": "mrkdwn",
							"text": "%s \n %v",
						}
					]
				},
				{
					"type": "divider"
				},`, key1, strings.Replace(fmt.Sprintf("%v", prop1), "\"", "", -1), key2, strings.Replace(fmt.Sprintf("%v", prop2), "\"", "", -1))
			count = 0
		}
	}
	return propBlock
}
func getSlackMsgBlock(msg model.EventTriggerAlertMessage) string {

	propBlock := getPropsBlockV2(msg.MessageProperty)

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
			"type": "section",
						"text": {
							"type": "mrkdwn",
							"text": "*<https://app.factors.ai/profiles/people|Know More>*"
						}
		}
	]`, strings.Replace(msg.Title, "\"", "", -1), strings.Replace(msg.Message, "\"", "", -1), propBlock)

	return mainBlock
}
func sendTeamsAlertForEventTriggerAlert(projectID int64, agentUUID string, msg model.EventTriggerAlertMessage, Tchannels *postgres.Jsonb) bool {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"agent_uuid": agentUUID,
	})
	var teamsChannels model.Team

	err := U.DecodePostgresJsonbToStructType(Tchannels, &teamsChannels)
	if err != nil {
		log.WithError(err).Error("failed to decode teams channels")
		return false
	}

	wetRun := true
	if wetRun {

		for _, channel := range teamsChannels.TeamsChannelList {
			message := getTeamsMessageTemp(msg)
			err := teams.SendTeamsMessage(projectID, agentUUID, teamsChannels.TeamsId, channel.ChannelId, message)
			if err != nil {
				logCtx.WithError(err).Error("failed to send teams message ", msg)
				return false
			}

			logCtx.Info("teams alert sent: ", channel, message)
		}

	} else {
		log.Info("Dry run mode enabled. No alerts will be sent")
		log.Info("*****", msg, projectID)
		return true
	}

	return true
}

func getPropsJsonForTeams(propMap U.PropertiesMap) string {
	propsBlock := ``
	for i := 0; i < len(propMap); i++ {
		pp := propMap[fmt.Sprintf("%d", i)]
		var mp model.MessagePropMapStruct
		if pp != nil {
			trans, ok := pp.(map[string]interface{})
			if !ok {
				log.Warn("cannot convert interface to map[string]interface{} type")
				continue
			}
			err := U.DecodeInterfaceMapToStructType(trans, &mp)
			if err != nil {
				log.Warn("cannot convert interface map to struct type")
				continue
			}
			key := mp.DisplayName
			prop := mp.PropValue
			if prop == "" {
				prop = "<nil>"
			}

			propsBlock += fmt.Sprintf(`{
				"name": %s,
				"value": %v"
			}`, key, strings.Replace(fmt.Sprintf("%v", prop), "\"", "", -1))

			if i != len(propMap)-1 {
				propsBlock += ","
			}
		}
	}
	return propsBlock
}

func getTeamsMessageJson(message model.EventTriggerAlertMessage) string {
	msg := fmt.Sprintf(`{
		"@type": "MessageCard",
		"@context": "http://schema.org/extensions",
		"themeColor": "0076D7",
		"summary": %s,
		"sections": [{
			"activityTitle": %s,
			"activitySubtitle": %s,
			"activityImage": "https://teamsnodesample.azurewebsites.net/static/img/image5.png",
			"facts": [%s],
		"potentialAction": [{
			"@type": "OpenUri",
			"name": "Know More",
			"targets": [{
				"os": "default",
				"uri": "https://app.factors.ai/profiles/people"
			}]
		}]
	}`, strings.Replace(message.Message, "\"", "", -1), strings.Replace(message.Title, "\"", "", -1), strings.Replace(message.Message, "\"", "", -1), getPropsJsonForTeams(message.MessageProperty))
	return msg
}

func getTeamsMessageTemp(message model.EventTriggerAlertMessage) string {
	msg := fmt.Sprintf(`
		%s 
		%s 
	`,strings.Replace(message.Title, "\"", "", -1),strings.Replace(message.Message, "\"", "", -1))
	propMap := message.MessageProperty
	for i := 0; i < len(propMap); i++ {
		pp := propMap[fmt.Sprintf("%d", i)]
		var mp model.MessagePropMapStruct
		if pp != nil {
			trans, ok := pp.(map[string]interface{})
			if !ok {
				log.Warn("cannot convert interface to map[string]interface{} type")
				continue
			}
			err := U.DecodeInterfaceMapToStructType(trans, &mp)
			if err != nil {
				log.Warn("cannot convert interface map to struct type")
				continue
			}
			key := mp.DisplayName
			prop := mp.PropValue
			if prop == "" {
				prop = "<nil>"
			}

			msg += fmt.Sprintf(`{
				"name": %s,
				"value": %v"
			}`, key, strings.Replace(fmt.Sprintf("%v", prop), "\"", "", -1))

			if i != len(propMap)-1 {
				msg += ","
			}
		}
	}
	return msg
}
