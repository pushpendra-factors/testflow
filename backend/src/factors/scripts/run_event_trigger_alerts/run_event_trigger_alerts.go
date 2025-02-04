package main

import (
	"factors/cache"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	teams "factors/integration/ms_teams"
	"factors/integration/paragon"
	slack "factors/integration/slack"
	webhook "factors/webhooks"

	"encoding/base64"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	SortedSetKeyPrefix           = "ETA"
	FailureSortedSetPrefix       = "ETA:Fail"
	RetryLimit                   = 10
	RejectedQueueLimit           = 3
	RejectedQueueSortedSetPrefix = "ETA:RejectedQueue"
	SLACK                        = "Slack"
	TEAMS                        = "Teams"
	WEBHOOK                      = "WH"
	ParagonUrlRune               = "zeus.useparagon.com"
)

type SendReportLogCount struct {
	SlackSuccess   int
	SlackFail      int
	TeamsSuccess   int
	TeamsFail      int
	WebhookSuccess int
	WebhookFail    int
}

type BlockedAlertList struct {
	alertID map[string]int
	keys    []string
}

func NewBlockedAlertList() *BlockedAlertList {
	return &BlockedAlertList{alertID: make(map[string]int), keys: make([]string, 0)}
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
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
	teamsAppTenantID := flag.String("teams_app_tenant_id", "", "")
	teamsAppClientID := flag.String("teams_app_client_id", "", "")
	teamsAppClientSecret := flag.String("teams_app_client_secret", "", "")
	teamsApplicationID := flag.String("teams_application_id", "", "")
	enableFeatureGatesV2 := flag.Bool("enable_feature_gates_v2", false, "")
	appDomain := flag.String("app_domain", "factors-dev.com:3000", "")
	// blacklistedAlerts := flag.String("blacklisted_alerts", "", "")

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
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:     *primaryDatastore,
		RedisHostPersistent:  *redisHostPersistent,
		RedisPortPersistent:  *redisPortPersistent,
		SentryDSN:            *sentryDSN,
		TeamsAppTenantID:     *teamsAppTenantID,
		TeamsAppClientID:     *teamsAppClientID,
		TeamsAppClientSecret: *teamsAppClientSecret,
		TeamsApplicationID:   *teamsApplicationID,
		EnableFeatureGatesV2: *enableFeatureGatesV2,
		APPDomain:            *appDomain,
	}
	defaultHealthcheckPingID := C.HealthcheckEventTriggerAlertPingID
	highPriorityHealthCheckPingID := C.HealthcheckEventTriggerAlertForHighPriorityPingID
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
	projectIDs, _ := store.GetStore().GetAllProjectIDs()

	tt := U.TimeNowZ()
	hour := tt.Hour()
	min := tt.Minute()

	successfulProjectsCount := 0 //Total number of projects with no failures
	failedProjectsCount := 0     //Total number of projects with atleast one failure

	for _, projectID := range projectIDs {
		available := true
		if *enableFeatureGatesV2 {
			available, err = store.GetStore().GetFeatureStatusForProjectV2(projectID, model.FEATURE_EVENT_BASED_ALERTS, false)
			if err != nil {
				log.WithError(err).Error("Failed to get feature status in event trigger alerts  job for project ID ", projectID)
				finalStatus[fmt.Sprintf("Failure-Feature-Status %v", projectID)] = true
			}
		}

		if !available {
			log.Error("Feature Not Available... Skipping event trigger alerts job for project ID ", projectID)
			return
		}

		blockedAlerts, errMsg, err := store.GetStore().UpdateInternalStatusAndGetAlertIDs(projectID)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID}).Error(errMsg)
		}
		blockedAlertMap := make(map[string]bool)
		for _, id := range blockedAlerts {
			blockedAlertMap[id] = true
		}

		sendReportForProject, blockedAlertList, projectSuccess := EventTriggerAlertsSender(projectID, conf, blockedAlertMap)
		if !projectSuccess {
			log.WithFields(log.Fields{"project_id": projectID}).Error("Event Trigger Alert job failing")
		}

		if projectSuccess {
			successfulProjectsCount++
		} else {
			failedProjectsCount++
		}

		if sendReportForProject.SlackSuccess > 0 {
			finalStatus[fmt.Sprintf("Success-SLACK-%v", projectID)] = sendReportForProject.SlackSuccess
		}
		if sendReportForProject.TeamsSuccess > 0 {
			finalStatus[fmt.Sprintf("Success-TEAMS-%v", projectID)] = sendReportForProject.TeamsSuccess
		}
		if sendReportForProject.WebhookSuccess > 0 {
			finalStatus[fmt.Sprintf("Success-WEBHOOK-%v", projectID)] = sendReportForProject.WebhookSuccess
		}
		if sendReportForProject.SlackFail > 0 {
			finalStatus[fmt.Sprintf("Failure-SLACK-%v", projectID)] = sendReportForProject.SlackFail
		}
		if sendReportForProject.TeamsFail > 0 {
			finalStatus[fmt.Sprintf("Failure-TEAMS-%v", projectID)] = sendReportForProject.TeamsFail
		}
		if sendReportForProject.WebhookFail > 0 {
			finalStatus[fmt.Sprintf("Failure-WEBHOOK-%v", projectID)] = sendReportForProject.WebhookFail
		}
		if len(blockedAlertList.alertID) > 0 {
			finalStatus[fmt.Sprintf("Blocked-alert-ids-%v", projectID)] = blockedAlertList.alertID
			finalStatus[fmt.Sprintf("Blocked-alert-keys-%v", projectID)] = blockedAlertList.keys
		}

		if min < 5 {
			sendReportForProject, blockedAlertList := RetryFailedEventTriggerAlerts(projectID, blockedAlertMap)
			if sendReportForProject.SlackSuccess > 0 {
				finalStatus[fmt.Sprintf("Retry Success-SLACK-%v", projectID)] = sendReportForProject.SlackSuccess
			}
			if sendReportForProject.TeamsSuccess > 0 {
				finalStatus[fmt.Sprintf("Retry Success-TEAMS-%v", projectID)] = sendReportForProject.TeamsSuccess
			}
			if sendReportForProject.WebhookSuccess > 0 {
				finalStatus[fmt.Sprintf("Retry Success-WEBHOOK-%v", projectID)] = sendReportForProject.WebhookSuccess
			}
			if sendReportForProject.SlackFail > 0 {
				finalStatus[fmt.Sprintf("Retry Failure-SLACK-%v", projectID)] = sendReportForProject.SlackFail
			}
			if sendReportForProject.TeamsFail > 0 {
				finalStatus[fmt.Sprintf("Retry Failure-TEAMS-%v", projectID)] = sendReportForProject.TeamsFail
			}
			if sendReportForProject.WebhookFail > 0 {
				finalStatus[fmt.Sprintf("Retry Failure-WEBHOOK-%v", projectID)] = sendReportForProject.WebhookFail
			}
			if len(blockedAlertList.alertID) > 0 {
				alertIds := make([]string, 0)
				for id := range blockedAlertList.alertID {
					alertIds = append(alertIds, id)
				}
				finalStatus[fmt.Sprintf("Retry Blocked-alert-ids-%v", projectID)] = alertIds
				finalStatus[fmt.Sprintf("Retry Blocked-alert-keys-%v", projectID)] = blockedAlertList.keys
			}
		}
		if hour == 12 && min < 5 {
			rejectedQueueReport, _, _ := RejectedQueueAlertSender(projectID)
			if rejectedQueueReport.SlackSuccess > 0 {
				finalStatus[fmt.Sprintf("Rejected Queue Success-SLACK-%v", projectID)] = sendReportForProject.SlackSuccess
			}
			if rejectedQueueReport.TeamsSuccess > 0 {
				finalStatus[fmt.Sprintf("Rejected Queue Success-TEAMS-%v", projectID)] = sendReportForProject.TeamsSuccess
			}
			if rejectedQueueReport.WebhookSuccess > 0 {
				finalStatus[fmt.Sprintf("Rejected Queue Success-WEBHOOK-%v", projectID)] = sendReportForProject.WebhookSuccess
			}
			if rejectedQueueReport.SlackFail > 0 {
				finalStatus[fmt.Sprintf("Rejected Queue Failure-SLACK-%v", projectID)] = sendReportForProject.SlackFail
			}
			if rejectedQueueReport.TeamsFail > 0 {
				finalStatus[fmt.Sprintf("Rejected Queue Failure-TEAMS-%v", projectID)] = sendReportForProject.TeamsFail
			}
			if rejectedQueueReport.WebhookFail > 0 {
				finalStatus[fmt.Sprintf("Rejected Queue Failure-WEBHOOK-%v", projectID)] = sendReportForProject.WebhookFail
			}
		}
	}

	if successfulProjectsCount/3 <= failedProjectsCount {
		C.PingHealthcheckForFailure(healthcheckPingID, finalStatus)
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, finalStatus)
	}

	// Check if the finalStatus map contains atleast one instance of 'Success' key
	var isHighPriorityHealthcheckOK bool
	for key := range finalStatus {
		if strings.Contains(key, "Success") {
			isHighPriorityHealthcheckOK = true
			break
		}
	}
	if isHighPriorityHealthcheckOK {
		C.PingHealthcheckForSuccess(highPriorityHealthCheckPingID, finalStatus)
	}

}

func getSortedSetCacheKey(prefix string, projectId int64) (*cache.Key, error) {
	pre := fmt.Sprintf("%s:pid:%d", prefix, projectId)
	key, err := cache.NewKeyWithOnlyPrefix(pre)
	if err != nil {
		log.WithError(err).Error("Cannot get redis key")
		return nil, err
	}
	return key, err
}

/*
EVENT TRIGGER ALERTS SENDER

# Picks keys from ETA:pid:<project_id> sorted set for sending
# Dispatches keys with other details one by one to send to the delivery options set by the user
# If the sendHelper function return a success, alert key is removed from the cache
# Whether the status received is a success or a failure we remove the entry for that key from the sorted set
*/
func EventTriggerAlertsSender(projectID int64, configs map[string]interface{},
	blockedAlert map[string]bool) (SendReportLogCount, *BlockedAlertList, bool) {

	logFields := log.Fields{
		"project_id": projectID,
	}
	logCtx := log.WithFields(logFields)

	ok := int(0)
	sendReportForProject := SendReportLogCount{}
	blockedAlertList := NewBlockedAlertList()
	ssKey, err := getSortedSetCacheKey(SortedSetKeyPrefix, projectID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch cacheKey for sortedSet")
		return sendReportForProject, blockedAlertList, false
	}

	allKeys, err := cacheRedis.ZrangeWithScoresPersistent(true, ssKey)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get all alert keys for project.")
		return sendReportForProject, blockedAlertList, false
	}

	for key := range allKeys {
		cacheKey, err := cache.KeyFromStringWithPid(key)
		if err != nil {
			logCtx.WithField("alert_key", key).WithError(err).
				Error("Failed to get cacheKey from the key string")
			continue
		}

		alertID := strings.Split(cacheKey.Suffix, ":")[0]
		if blockedAlert[alertID] {
			blockedAlertList.alertID[alertID] += 1
			blockedAlertList.keys = append(blockedAlertList.keys, key)
			SendKeyToRejectedQueue(key, cacheKey, ssKey, projectID)
			ok++
			continue
		}

		cacheStr, err := cacheRedis.GetPersistent(cacheKey)
		if err != nil {
			logCtx.WithField("alert_key", key).WithError(err).
				Error("failed to find value of the cached alert")
			continue
		}
		var msg model.CachedEventTriggerAlert
		err = U.DecodeJSONStringToStructType(cacheStr, &msg)
		if err != nil {
			logCtx.WithField("alert_key", key).WithError(err).
				Error("failed to decode cached JSON alert")
			continue
		}

		totalSuccess := false
		sendReport := SendReportLogCount{}
		if msg.IsWorkflow {
			totalSuccess, _, sendReport = ProcessWorkflow(cacheKey, &msg, alertID, false, "")
		} else {
			totalSuccess, _, sendReport = sendHelperForEventTriggerAlert(cacheKey, &msg, alertID, false, "")
		}

		if totalSuccess {
			err = cacheRedis.DelPersistent(cacheKey)
			if err != nil {
				logCtx.WithField("alert_key", key).WithError(err).Error("failed to remove alert from cache")
			}
			ok++
		}
		cc, err := cacheRedis.ZRemPersistent(ssKey, true, key)
		if err != nil || cc != 1 {
			logCtx.WithField("alert_key", key).WithError(err).Error("failed to remove alert from sorted set")
		}

		sendReportForProject.addToSendReport(sendReport)
	}
	return sendReportForProject, blockedAlertList, ok == len(allKeys)
}

func SendKeyToRejectedQueue(strKey string, key, ssKey *cache.Key, projectID int64) error {

	logFields := log.Fields{
		"project_id":     projectID,
		"sorted_set_key": *ssKey,
		"alert_key":      key,
	}
	logCtx := log.WithFields(logFields)
	cc, err := cacheRedis.ZRemPersistent(ssKey, true, strKey)
	if err != nil || cc != 1 {
		logCtx.WithError(err).Error("Cannot remove alert by zrem")
		return err
	}

	keyStr, err := key.Key()
	if err != nil {
		logCtx.WithError(err).Error("Cannot get stringified key from redis key")
		return err
	}

	rqSSKey, err := getSortedSetCacheKey(RejectedQueueSortedSetPrefix, projectID)
	if err != nil {
		logCtx.WithError(err).Error("Cannot get sorted set key")
		return err
	}

	ssTuple := cacheRedis.SortedSetKeyValueTuple{
		Key:   rqSSKey,
		Value: keyStr,
	}
	_, err = cacheRedis.ZincrPersistentBatch(true, ssTuple)
	if err != nil {
		logCtx.WithError(err).Error("failed to update sortedSet in cache")
		return err
	}
	return nil
}

func (projReport *SendReportLogCount) addToSendReport(alertReport SendReportLogCount) {
	projReport.SlackFail += alertReport.SlackFail
	projReport.SlackSuccess += alertReport.SlackSuccess
	projReport.TeamsFail += alertReport.TeamsFail
	projReport.TeamsSuccess += alertReport.TeamsSuccess
	projReport.WebhookFail += alertReport.WebhookFail
	projReport.WebhookSuccess += alertReport.WebhookSuccess
}

/*
# Total success is if all the failures are zero
# Partial success is if atleast one send is a success
*/
func findTotalAndPartialSuccess(report SendReportLogCount) (bool, bool) {
	total := report.SlackFail + report.TeamsFail + report.WebhookFail
	partial := report.SlackSuccess + report.TeamsSuccess + report.WebhookSuccess

	return total == 0, partial > 0
}

func GetSlackIdForCorrespondingHubspotOwnerIds(projectID int64, agentID string, fieldTagsMap map[string]string) []string {
	logFields := log.Fields{
		"project_id": projectID,
		"agent_uuid": agentID,
	}
	logCtx := log.WithFields(logFields)

	fieldTagsWithEmailId := make(map[string]string)

	for tag, ownerId := range fieldTagsMap {
		email, errCode, err := store.GetStore().GetHubspotOwnerEmailFromOwnerId(projectID, ownerId)
		if errCode != http.StatusFound || err != nil || email == "" {
			if err != nil {
				logCtx.WithField("owner_id", ownerId).WithError(err).Error("fetch email from hubspot document failed")
				continue
			}
			logCtx.WithField("owner_id", ownerId).Warn("No email available for owner_id")
			continue
		}
		fieldTagsWithEmailId[tag] = email
	}

	slackUsersList := make([]model.SlackMember, 0)
	var err error
	var errCode int
	// get slack users
	if fieldTagsMap != nil {
		slackUsersList, errCode, err = store.GetStore().GetSlackUsersListFromDb(projectID, agentID)
		if err != nil || errCode != http.StatusFound || slackUsersList == nil {
			logCtx.WithError(err).Error("failed to fetch slack users")
			return nil
		}
	}

	slackIdsToBeTagged := make([]string, 0)
	for _, email := range fieldTagsWithEmailId {
		for _, user := range slackUsersList {
			if !user.Deleted && user.Profile.Email == email {
				slackIdsToBeTagged = append(slackIdsToBeTagged, user.Id)
			}
		}
	}
	return slackIdsToBeTagged
}

/*
SEND HELPER FOR EVENT TRIGGER ALERT

# Get the alert from db.
# If the alert is from the first sorted set or the rejected queue sorted set, then find the alert delivery
configuration and pass it to the block executing the send for that option
# Else if the alert is from the retry then check the sendTo variable and send it to the block which is
sending for that delivery option
# For every block sending to the corresponding delivery option, we collect the success and failure report
error messages, and delivery failures.
# From the report we find total and partial success of the run using the findTotalAndPartialSuccess func
# In case of partialSuccess we update the last_alert_sent column in the db
# In case the send for the alert was NOT a total success we execute the
EventTriggerDeliveryFailureExecution func with the errorMessage and deliveryFailure collected
*/
func sendHelperForEventTriggerAlert(key *cache.Key, alert *model.CachedEventTriggerAlert,
	alertID string, retry bool, sendTo string) (totalSuccess bool, partialSuccess bool, sendReport SendReportLogCount) {

	logCtx := log.WithFields(log.Fields{
		"key":      key,
		"alert":    alert,
		"alert_id": alertID,
		"retry":    retry,
		"send_to":  sendTo,
	})

	errMessage := make([]string, 0)
	deliveryFailures := make([]string, 0)
	rejectedQueue := false

	eta, errCode := store.GetStore().GetEventTriggerAlertByID(alertID)
	if errCode != http.StatusFound {
		if errCode == http.StatusNotFound {
			logCtx.Info("Alert definition not found.")
			return true, true, sendReport
		}
		logCtx.WithField("err_code", errCode).Error("Failed to fetch alert from db.")
		return false, false, sendReport
	}

	var alertConfiguration model.EventTriggerAlertConfig
	err := U.DecodePostgresJsonbToStructType(eta.EventTriggerAlert, &alertConfiguration)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode Jsonb to struct type")
		return false, false, sendReport
	}

	if sendTo == "RejectedQueue" {
		rejectedQueue = true
	}

	var accountUrl, hubspotAccountUrl, salesforceAccountUrl string
	isAccounAlert := alertConfiguration.EventLevel == model.EventLevelAccount
	if isAccounAlert {
		if _, exists := alert.Message.MessageProperty[model.ETA_DOMAIN_GROUP_USER_ID]; exists {
			// factors account url
			groupDomainUserID := alert.Message.MessageProperty[model.ETA_DOMAIN_GROUP_USER_ID].(string)
			accountUrl = BuildAccountURL(groupDomainUserID)
			delete(alert.Message.MessageProperty, model.ETA_DOMAIN_GROUP_USER_ID)
		}
		// hubspot object url
		if hubspotUrlInProperties, doesHubspotUrlExist := alert.Message.MessageProperty[model.ETA_ENRICHED_HUBSPOT_COMPANY_OBJECT_URL]; doesHubspotUrlExist {
			hubspotAccountUrl = hubspotUrlInProperties.(string)
			delete(alert.Message.MessageProperty, model.ETA_ENRICHED_HUBSPOT_COMPANY_OBJECT_URL)
		}

		// salesforce object url
		if salesforceUrlInProperties, doesSalesforceUrlExist := alert.Message.MessageProperty[model.ETA_ENRICHED_SALESFORCE_ACCOUNT_OBJECT_URL]; doesSalesforceUrlExist {
			salesforceAccountUrl = salesforceUrlInProperties.(string)
			delete(alert.Message.MessageProperty, model.ETA_ENRICHED_SALESFORCE_ACCOUNT_OBJECT_URL)
		}
	}

	// If retry is true and sendTo var is set for Slack option
	// OR
	// If the retry is off, meaning the alert is from the first try or rejected queue and configuration
	//		for alert is set for Slack option
	if (retry && strings.EqualFold(SLACK, sendTo)) || (!retry && alertConfiguration.Slack) {

		isSlackIntergrated, errCode := store.GetStore().IsSlackIntegratedForProject(eta.ProjectID,
			eta.SlackChannelAssociatedBy)
		if errCode != http.StatusOK {
			logCtx.WithFields(log.Fields{"agentID": eta.SlackChannelAssociatedBy, "event_trigger_alert_id": eta.ID}).
				Error("failed to check slack integration")
		}
		if isSlackIntergrated {
			partialSlackSuccess, _, errMsg := sendSlackAlertForEventTriggerAlert(eta.ProjectID,
				eta.SlackChannelAssociatedBy, alert, alertConfiguration.SlackChannels, alertConfiguration.SlackMentions, alertConfiguration.IsHyperlinkDisabled, isAccounAlert, accountUrl, hubspotAccountUrl, salesforceAccountUrl)
			log.WithFields(log.Fields{
				"project_id": eta.ProjectID,
				"alert_id":   eta.ID,
				"mode":       SLACK,
				"retry":      retry,
				"is_success": partialSlackSuccess,
				"tag":        "alert_tracker",
			}).Info("ALERT TRACKER.")
			if !partialSlackSuccess {
				sendReport.SlackFail++
				errMessage = append(errMessage, errMsg)
				deliveryFailures = append(deliveryFailures, SLACK)

			} else {
				sendReport.SlackSuccess++
			}
		} else {
			logCtx.WithFields(log.Fields{"alert_id": alertID}).Error("integration not found for slack configuration")
		}

	}

	// If retry is true and sendTo var is set for Teams option
	// OR
	// If the retry is off, meaning the alert is from the first try or rejected queue and configuration
	//		for alert is set for Teams option
	if (retry && strings.EqualFold(TEAMS, sendTo)) || (!retry && alertConfiguration.Teams) {

		isTeamsIntergrated, errCode := store.GetStore().IsTeamsIntegratedForProject(eta.ProjectID,
			eta.TeamsChannelAssociatedBy)
		if errCode != http.StatusOK {
			logCtx.WithFields(log.Fields{"agentID": eta.TeamsChannelAssociatedBy, "event_trigger_alert": alert}).
				Error("failed to check teams integration")
		}
		if isTeamsIntergrated {
			teamsSuccess, errMsg := sendTeamsAlertForEventTriggerAlert(eta.ProjectID,
				eta.TeamsChannelAssociatedBy, alert.Message, alertConfiguration.TeamsChannelsConfig, isAccounAlert, accountUrl)
			log.WithFields(log.Fields{
				"project_id": eta.ProjectID,
				"alert_id":   eta.ID,
				"mode":       TEAMS,
				"retry":      retry,
				"is_success": teamsSuccess,
				"tag":        "alert_tracker",
			}).Info("ALERT TRACKER.")
			if !teamsSuccess {
				sendReport.TeamsFail++
				errMessage = append(errMessage, errMsg)
				deliveryFailures = append(deliveryFailures, TEAMS)
			} else {
				sendReport.TeamsSuccess++
			}
		} else {
			logCtx.WithFields(log.Fields{"alert_id": alertID}).Error("integration not found for teams configuration")
		}

	}

	// If retry is true and sendTo var is set for Webhook option
	// OR
	// If the retry is off, meaning the alert is from the first try or rejected queue and configuration
	//		for alert is set for Webhook option
	if (retry && strings.EqualFold(WEBHOOK, sendTo)) || (!retry && alertConfiguration.Webhook) {

		var response = make(map[string]interface{})

		//Adding factors account URL to webhook payload
		if !isAccounAlert {
			accountUrl = "https://app.factors.ai/profiles/people"
		} else if isAccounAlert && accountUrl == "" {
			accountUrl = "https://app.factors.ai"
		}

		if alertConfiguration.IsFactorsUrlInPayload {
			idx := len(alert.Message.MessageProperty)
			alert.Message.MessageProperty[fmt.Sprintf("%d", idx)] = model.MessagePropMapStruct{
				DisplayName: "Factors Activity URL",
				PropValue:   accountUrl,
			}
		}

		if strings.Contains(alertConfiguration.WebhookURL, ParagonUrlRune) {
			response, err = paragon.SendPayloadToParagonForTheAlert(eta.ProjectID, eta.ID, &alertConfiguration, alert)
			if err != nil {
				logCtx.WithFields(log.Fields{"alert_id": alertID, "server_response": response}).
					WithError(err).Error("Paragon event failure")
			}
			logCtx.WithField("response", response).Info("paragon response")
		} else {
			response, err = webhook.DropWebhook(alertConfiguration.WebhookURL, alertConfiguration.Secret, alert.Message)
			if err != nil {
				logCtx.WithFields(log.Fields{"alert_id": alertID, "server_response": response}).
					WithError(err).Error("Webhook failure")
			}
		}
		logCtx.WithField("alert", alert).WithField("response", response).Info("Webhook dropped for alert.")
		stat := response["status"]
		//if atleast one property field is not null in payload the payload is considered not null
		isPayloadNull := true
		for _, val := range alert.Message.MessageProperty {
			if val != nil {
				isPayloadNull = false
			}
		}
		log.WithFields(log.Fields{
			"project_id":      eta.ProjectID,
			"alert_id":        eta.ID,
			"mode":            WEBHOOK,
			"retry":           retry,
			"is_success":      stat == "success",
			"tag":             "alert_tracker",
			"is_payload_null": isPayloadNull,
		}).Info("ALERT TRACKER.")

		if response["error"] == "<nil>" {
			response["error"] = "an"
		}
		if stat != "success" {
			log.WithField("status", stat).WithField("response", response).Error("Web hook error details")
			sendReport.WebhookFail++
			errMessage = append(errMessage, fmt.Sprintf("Webhook host reported %v error", response["error"]))
			deliveryFailures = append(deliveryFailures, WEBHOOK)

		} else {
			sendReport.WebhookSuccess++
		}

	}

	totalSuccess, partialSuccess = findTotalAndPartialSuccess(sendReport)
	// not total success means there has been atleast one failure
	if !totalSuccess {
		err := EventTriggerDeliveryFailureExecution(key, eta, deliveryFailures, errMessage, rejectedQueue, partialSuccess)
		if err != nil {
			logCtx.WithError(err).Error("failed while updating teams-fail flow")
		}
	}

	// partial success means there has been atleast one success
	if partialSuccess {
		status, err := store.GetStore().UpdateEventTriggerAlertField(eta.ProjectID, eta.ID,
			map[string]interface{}{"last_alert_at": U.TimeNowZ()})
		if status != http.StatusAccepted || err != nil {
			logCtx.WithError(err).Error("Failed to update db field")
		}
	}

	return totalSuccess, partialSuccess, sendReport
}

/*
EVENT TRIGGER DELIVERY FAILURE EXECUTION

# In case of failure while sending the alert at any point
-> For failure in first sending trial, alert key is added for each failure point in the FailureSortedSet
-> For failure in retry sending, count for the same alert key is incremented in the FailureSortedSet
-> For failure in rejected queue sending, count for the alert is incremented once in the rejected sorted set
for all the delivery failures
-> For partial success in the rejected queue sending, the alert is sent to FailureSortedSet with the poin
it failed while sending

# Update the last_fail_details column in the DB
*/
func EventTriggerDeliveryFailureExecution(key *cache.Key, eta *model.EventTriggerAlert,
	deliveryFailures, errMsg []string, rejected, partialSuccess bool) error {

	logFields := log.Fields{
		"alert_id":  eta.ID,
		"alert_key": key,
	}
	logCtx := log.WithFields(logFields)

	if rejected && !partialSuccess {
		err := AddKeyToSortedSet(key, eta.ProjectID, "RejectedQueue", rejected, partialSuccess)
		if err != nil {
			logCtx.WithError(err).Error("failed to put key in FailureSortedSet")
			return err
		}
	} else {
		for _, failPoint := range deliveryFailures {
			err := AddKeyToSortedSet(key, eta.ProjectID, failPoint, rejected, partialSuccess)
			if err != nil {
				logCtx.WithError(err).Error("failed to put key in FailureSortedSet")
				return err
			}
		}
	}

	errDetails := model.LastFailDetails{
		FailTime: U.TimeNowZ(),
		FailedAt: deliveryFailures,
		Details:  errMsg,
	}
	errJson, err := U.EncodeStructTypeToPostgresJsonb(errDetails)
	if err != nil {
		logCtx.WithError(err).Error("failed to encode struct to jsonb")
		return err
	}

	status, err := store.GetStore().UpdateEventTriggerAlertField(eta.ProjectID, eta.ID,
		map[string]interface{}{"last_fail_details": errJson})
	if status != http.StatusAccepted || err != nil {
		logCtx.WithError(err).Error("Failed to update db field")
		return err
	}

	return nil
}

/*
ADD KEY TO SORTED SET

# Rejected queues are retried but they are updated once for all the runs
# Retries are updated once for each delivery option failure
*/
func AddKeyToSortedSet(key *cache.Key, projectID int64, failPoint string, rejected, partialSuccess bool) error {

	logFields := log.Fields{
		"project_id": projectID,
		"alert_key":  key,
	}
	logCtx := log.WithFields(logFields)

	prefix := ""
	if rejected && !partialSuccess {
		prefix = RejectedQueueSortedSetPrefix

	} else {
		prefix = FailureSortedSetPrefix
	}

	ssKey, err := getSortedSetCacheKey(prefix, projectID)
	if err != nil {
		logCtx.WithError(err).Error("failed to fetch sorted set key for failure set")
		return err
	}

	val, err := key.Key()
	if err != nil {
		logCtx.WithError(err).Error("cannot find str value for cache key")
		return err
	}

	ssTuple := cacheRedis.SortedSetKeyValueTuple{
		Key: ssKey,
	}
	if rejected && !partialSuccess {
		ssTuple.Value = val
	} else {
		ssTuple.Value = fmt.Sprintf("%s:%s", failPoint, val)
	}

	_, err = cacheRedis.ZincrPersistentBatch(true, ssTuple)
	if err != nil {
		logCtx.WithError(err).Error("failed to update sortedSet in cache")
		return err
	}

	return nil
}

func sendSlackAlertForEventTriggerAlert(projectID int64, agentUUID string,
	alert *model.CachedEventTriggerAlert, Schannels, sMentions *postgres.Jsonb, isHyperlinkDisabled, isAccountAlert bool, accountUrl, hsAccUrl, sfAccUrl string) (partialSuccess bool, channelSuccess []bool, errMessage string) {
	logCtx := log.WithFields(log.Fields{
		"project_id":  projectID,
		"agent_uuid":  agentUUID,
		"alert_title": alert.Message.Title,
	})
	var slackChannels []model.SlackChannel
	partialSuccess = false

	if Schannels == nil {
		errMsg := "no slack channels provided"
		errMessage += errMsg
		logCtx.Error(errMsg)
		return false, channelSuccess, errMessage
	}

	err := U.DecodePostgresJsonbToStructType(Schannels, &slackChannels)
	if err != nil {
		errMsg := "failed to decode slack channels"
		errMessage += errMsg
		logCtx.WithError(err).Error(errMsg)
		return false, channelSuccess, errMessage
	}

	slackMentions := make([]model.SlackMember, 0)
	if sMentions != nil {
		if err := U.DecodePostgresJsonbToStructType(sMentions, &slackMentions); err != nil {
			errMsg := "failed to decode slack mentions"
			errMessage += errMsg
			logCtx.WithError(err).Error(errMsg)
			return false, channelSuccess, errMessage
		}
	}

	slackTags := make([]string, 0)
	if alert.FieldTags != nil {
		slackTags = GetSlackIdForCorrespondingHubspotOwnerIds(projectID, agentUUID, alert.FieldTags)
	}

	wetRun := true
	if wetRun {
		for _, channel := range slackChannels {
			errMsg := "successfully sent"
			var blockMessage, slackMentionStr string
			if slackMentions != nil || slackTags != nil {
				slackMentionStr = model.GetSlackMentionsStr(slackMentions, slackTags)
			}
			if !isHyperlinkDisabled {
				blockMessage = model.GetSlackMsgBlock(alert.Message, slackMentionStr, isAccountAlert, accountUrl, hsAccUrl, sfAccUrl)
			} else {
				blockMessage = model.GetSlackMsgBlockWithoutHyperlinks(alert.Message, slackMentionStr)
			}

			response, status, err := slack.SendSlackAlert(projectID, blockMessage, agentUUID, channel)
			partialSuccess = partialSuccess || status
			if err != nil || !status {
				if response["error"] != nil {
					errMsg = response["error"].(string)
				}
				slackErr, exists := model.SlackErrorStates[errMsg]
				if !exists {
					errMessage += fmt.Sprintf("Slack reported %s for %s channel\n\n", errMsg, channel.Name)
				} else {
					errMessage += fmt.Sprintf("%s Error for %s channel\n\n", slackErr, channel.Name)
				}
				channelSuccess = append(channelSuccess, false)
				logCtx.WithField("channel", channel).WithError(fmt.Errorf("%v", response)).
					Error("failed to send slack alert")
				continue
			}
			channelSuccess = append(channelSuccess, true)
		}
	} else {
		channelSuccess = append(channelSuccess, true)
		errMessage = ""
		log.Info("Dry run mode enabled. No alerts will be sent")
		log.Info("*****", alert.Message, projectID)
		return true, channelSuccess, errMessage
	}

	return partialSuccess, channelSuccess, errMessage
}

/*
OBSOLETE CODE
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
*/

func sendTeamsAlertForEventTriggerAlert(projectID int64, agentUUID string,
	msg model.EventTriggerAlertMessage, Tchannels *postgres.Jsonb, isAccountAlert bool, accountUrl string) (success bool, errMessage string) {
	logCtx := log.WithFields(log.Fields{
		"project_id":  projectID,
		"agent_uuid":  agentUUID,
		"alert_title": msg.Title,
	})
	var teamsChannels model.Team
	if Tchannels == nil {
		errMsg := "no teams channels found"
		logCtx.Error(errMsg)
		return false, errMsg
	}

	err := U.DecodePostgresJsonbToStructType(Tchannels, &teamsChannels)
	if err != nil {
		errMsg := err.Error()
		logCtx.WithError(err).Error(errMsg)
		return false, errMsg
	}

	wetRun := true
	if wetRun {

		for _, channel := range teamsChannels.TeamsChannelList {
			message := model.GetTeamsMsgBlock(msg, isAccountAlert, accountUrl)
			response, err := teams.SendTeamsMessage(projectID, agentUUID, teamsChannels.TeamsId,
				channel.ChannelId, message)
			if err != nil {
				// errMsg := err.Error()
				errorCode, ok := response["error"].(map[string]interface{})["code"].(string)
				teamsErr, exists := model.TeamsErrorStates[errorCode]
				if !ok || !exists {
					errMessage = "Teams reported an error."
				} else {
					errMessage += fmt.Sprintf("%s Error for %s channel\n\n", teamsErr, channel.ChannelName)
				}
				logCtx.WithFields(log.Fields{
					"err_response": response,
					"error_code":   errorCode,
				}).WithError(err).Error("Failed to send teams message for event alert.")
				return false, errMessage
			}

			logCtx.Info("teams alert sent: ", channel, message)
		}

	} else {
		log.Info("Dry run mode enabled. No alerts will be sent")
		log.Info("*****", msg, projectID)
		return true, ""
	}

	return true, ""
}

/*
OBSOLETE CODE
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
*/

func getTeamsMessageTemp(message model.EventTriggerAlertMessage) string {
	msg := fmt.Sprintf(`
		%s 
		%s 
	`, strings.Replace(message.Title, "\"", "", -1), strings.Replace(message.Message, "\"", "", -1))
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

/*
RETRY FAILED EVENT TRIGGER ALERTS

# Gets all keys from FailureSortedSet

# For each and its corresponding count we:

	-> Convert the count to int64
	-> The Failure key is of the form <fail_point>:ETA:pid:<project_id>:<alert_id>:<UnixTime>
	-> We segregate the fail_point and original alert key
	-> Convert the original alert key from string to *cache.Key type
	-> Find the alertId from the alert key
	-> Check if the alertId is in blacklisted alerts then don't process it, instead send it to the rejected queue
	-> Calculate the backoff time from the alert's count and catch time (unix_time) of the alert
	-> If the BACKOFF TIME condition is satisfied then move forward otherwise skip for the current alert
	-> Get the cached alert from the alert key
	-> Dispatch the alert to sendHelperForEventTriggerAlert with the fail_point and other values

# If the send is a success (total and partial success in this context is same) then:

	-> Remove the alert from the FailureSortedSet
	-> If there are no more alerts present for that alertID, then only remove the alert key from the cache

# Else for a failure, we simply increase the count for that key (done by sendHelperForEventTriggerAlert)

# If the alert has exhausted the retry limit then remove the alert from the SortedSet and cache both

# BACKOFF TIME CALCULATION

	-> cc = current_count in the sorted set for the key
	-> (current time) - (time at which alert was caught) > (cc * (cc+1)/2)
	-> This means that the alert will be retried one hour from the catch time,
	then two hours after the first retry
	then three hours after the second retry and so on.
*/
func RetryFailedEventTriggerAlerts(projectID int64, blockedAlerts map[string]bool) (SendReportLogCount, *BlockedAlertList) {

	logFields := log.Fields{
		"project_id": projectID,
	}
	logCtx := log.WithFields(logFields)

	sendReportForProject := SendReportLogCount{}
	blockedAlertList := NewBlockedAlertList()
	ssKey, err := getSortedSetCacheKey(FailureSortedSetPrefix, projectID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch cacheKey for sortedSet")
		return sendReportForProject, blockedAlertList
	}

	allKeys, err := cacheRedis.ZrangeWithScoresPersistent(true, ssKey)
	if err != nil {
		logCtx.WithField("sorted_set_key", *ssKey).WithError(err).Error("Failed to get all alert keys from sorted set")
		return sendReportForProject, blockedAlertList
	}

	for key, count := range allKeys {
		//Get the key from failed set
		cc, err := strconv.ParseInt(count, 0, 64)
		if err != nil {
			logCtx.WithField("sorted_set_key", *ssKey).WithError(err).
				Error("unable to parse int in event_trigger_alerts_job")
		}

		orgKey := strings.SplitAfterN(key, ":", 2)

		cacheKey, err := cache.KeyFromStringWithPid(orgKey[1])
		if err != nil {
			logCtx.WithFields(log.Fields{"sorted_set_key": *ssKey, "alert_key": orgKey[1]}).
				Error("failed to get cacheKey from the key string, retry failed")
			continue
		}

		cacheKeySplit := strings.Split(cacheKey.Suffix, ":")
		alertID := cacheKeySplit[0]
		if blockedAlerts[alertID] {
			blockedAlertList.alertID[alertID] += 1
			blockedAlertList.keys = append(blockedAlertList.keys, orgKey[1])
			SendKeyToRejectedQueue(key, cacheKey, ssKey, projectID)
			continue
		}

		firstTry := cacheKeySplit[len(cacheKeySplit)-1]
		retryTime, err := strconv.ParseInt(firstTry, 0, 64)
		if err != nil {
			logCtx.WithFields(log.Fields{"sorted_set_key": *ssKey, "alert_key": orgKey[1]}).
				WithError(err).Error("unable to parse int from string in event_trigger_alerts_job")
		}

		now := U.TimeNowZ().UnixNano()
		expBackoff := cc * (cc + 1) / 2
		if now-retryTime < expBackoff*60*60*1000000000 {
			log.Info("Skipping retry for alert: ", orgKey[1], ", because retry coolDown condition is false")
			continue
		}

		cacheStr, err := cacheRedis.GetPersistent(cacheKey)
		if err != nil {
			logCtx.WithFields(log.Fields{"sorted_set_key": *ssKey, "alert_key": orgKey[1]}).
				WithError(err).Error("failed to find message for the alert, retry failed")
			continue
		}
		//Get the cached alert
		var msg model.CachedEventTriggerAlert
		err = U.DecodeJSONStringToStructType(cacheStr, &msg)
		if err != nil {
			logCtx.WithFields(log.Fields{"sorted_set_key": *ssKey, "alert_key": orgKey[1]}).
				WithError(err).Error("failed to decode alert for event_trigger_alert, retry failed")
			continue
		}

		sendTo := ""
		if strings.Contains(orgKey[0], "Slack") {
			sendTo = "Slack"
		}
		if strings.Contains(orgKey[0], "WH") {
			sendTo = "WH"
		}
		if strings.Contains(orgKey[0], "Teams") {
			sendTo = "Teams"
		}

		var totalSuccess bool
		sendReport := SendReportLogCount{}
		if msg.IsWorkflow {
			if cc <= 3 {
				totalSuccess, _, sendReport = ProcessWorkflow(cacheKey, &msg, alertID, true, sendTo)
			}
		} else {
			totalSuccess, _, sendReport = sendHelperForEventTriggerAlert(cacheKey, &msg, alertID, true, sendTo)
		}

		sendReportForProject.addToSendReport(sendReport)

		if totalSuccess {
			cc, err := cacheRedis.ZRemPersistent(ssKey, true, key)
			if err != nil || cc != 1 {
				logCtx.WithFields(log.Fields{"sorted_set_key": *ssKey, "alert_key": key}).
					WithError(err).Error("Cannot remove alert by zrem")
			}
			allKeys[key] = "-1"
			areMoreAlertPresent := false
			for kk := range allKeys {
				if kk != key && strings.Contains(kk, alertID) && allKeys[kk] != "-1" {
					areMoreAlertPresent = true
				}
			}
			if !areMoreAlertPresent {
				err = cacheRedis.DelPersistent(cacheKey)
				if err != nil {
					logCtx.WithFields(log.Fields{"sorted_set_key": *ssKey, "alert_key": orgKey[1]}).
						WithError(err).Error("Cannot remove alert from cache")
				}
			}
		}

		if allKeys[key] == fmt.Sprintf("%d", RetryLimit-1) {
			cc, err := cacheRedis.ZRemPersistent(ssKey, true, key)
			if err != nil || cc != 1 {
				logCtx.WithFields(log.Fields{"sorted_set_key": *ssKey, "alert_key": key}).
					WithError(err).Error("Cannot remove alert by zrem")
			}
			logCtx.WithFields(log.Fields{"sorted_set_key": *ssKey, "alert_key": key}).
				Error("Retry limit reached. Removing the key completely from the cache")

			err = cacheRedis.DelPersistent(cacheKey)
			if err != nil {
				logCtx.WithFields(log.Fields{"sorted_set_key": *ssKey, "alert_key": orgKey[1]}).
					WithError(err).Error("Cannot remove alert from cache")
			}
		}
	}
	return sendReportForProject, blockedAlertList
}

/*
REJECTED ALERT SENDER FUNCTION

# Gets all keys from Rejected queue
# Rejected queue is the queue of the cached alerts for alertIDs which have a
difference between last_failtime and last_alert_sent as 72 hours or more

# Tries to send the keys one by one

If the send is a partialSuccess then,

	-> Remove the alert key from the sorted set
	-> Update the alert internal status to active
	-> Send the alert key to the failureSortedSet (This happens from the sendHelperForEventTriggerAlert func)

Else If the send is a totalSuccess then,

	-> Remove the alert key from the cache

Else if the send is a failure then,

	-> Increase the count in the sorted set (This happens from the sendHelperForEventTriggerAlert func)
*/
func RejectedQueueAlertSender(projectID int64) (SendReportLogCount, string, error) {

	report := SendReportLogCount{}
	logCtx := log.WithField("project_id", projectID)

	ssKey, err := getSortedSetCacheKey(RejectedQueueSortedSetPrefix, projectID)
	if err != nil {
		errMsg := "Failed to fetch cacheKey for sortedSet"
		logCtx.WithError(err).Error(errMsg)
		return report, errMsg, err
	}

	allKeys, err := cacheRedis.ZrangeWithScoresPersistent(true, ssKey)
	if err != nil {
		errMsg := "Failed to get all alert keys from cache"
		logCtx.WithError(err).Error(errMsg)
		return report, errMsg, err
	}

	for key := range allKeys {
		cacheKey, err := cache.KeyFromStringWithPid(key)
		if err != nil {
			errMsg := "failure while finding cacheKey from string key"
			logCtx.WithField("alert_key", key).WithError(err).Error(errMsg)
			continue
		}

		alertID := strings.Split(cacheKey.Suffix, ":")[0]

		cacheStr, err := cacheRedis.GetPersistent(cacheKey)
		if err != nil {
			errMsg := "failed to find message for the alert"
			logCtx.WithField("alert_key", key).WithError(err).Error(errMsg)
			continue
		}

		var msg model.CachedEventTriggerAlert
		err = U.DecodeJSONStringToStructType(cacheStr, &msg)
		if err != nil {
			errMsg := "failed to decode cached alert for event_trigger_alert"
			logCtx.WithField("alert_key", key).WithError(err).Error(errMsg)
			continue
		}

		totalSuccess, partialSuccess, sendReport := sendHelperForEventTriggerAlert(cacheKey, &msg, alertID, false, "RejectedQueue")
		report.addToSendReport(sendReport)

		if partialSuccess {
			cc, err := cacheRedis.ZRemPersistent(ssKey, true, key)
			if err != nil || cc != 1 {
				errMsg := "cannot remove alert from rejected queue set"
				logCtx.WithField("alert_key", key).WithError(err).Error(errMsg)
			}
			internalStatus := map[string]interface{}{
				"internal_status": model.Active,
			}
			errCode, err := store.GetStore().UpdateEventTriggerAlertField(projectID, alertID, internalStatus)
			if errCode != http.StatusAccepted || err != nil {
				errMsg := "cannot update internal status"
				logCtx.WithField("alert_id", alertID).WithError(err).Error(errMsg)
			}
			allKeys[key] = "-1"
		}

		if totalSuccess {
			err = cacheRedis.DelPersistent(cacheKey)
			if err != nil {
				errMsg := "cannot remove alert from cache"
				logCtx.WithField("alert_key", key).WithError(err).Error(errMsg)
			}
		}

	}
	RemoveContinuouslyFailingRejectedKeys(allKeys, ssKey)
	return report, "", nil
}

func RemoveContinuouslyFailingRejectedKeys(keys map[string]string, ssKey *cache.Key) {
	for key, count := range keys {
		cc, err := strconv.ParseInt(count, 0, 64)
		if err != nil {
			log.WithError(err).Error("unable to convert string to int")
		}
		if cc > RejectedQueueLimit {
			_, err := cacheRedis.ZRemPersistent(ssKey, true, key)
			if err != nil {
				log.WithFields(log.Fields{"sorted_set_key": *ssKey, "alert_key": key}).WithError(err).
					Error("cannot remove key from sorted set in cache")
			}
			cacheKey, err := cache.KeyFromStringWithPid(key)
			if err != nil {
				log.WithFields(log.Fields{"alert_key": key}).WithError(err).
					Error("unable to convert string to cacheRedis key")
			}
			err = cacheRedis.DelPersistent(cacheKey)
			if err != nil {
				log.WithFields(log.Fields{"alert_key": key}).WithError(err).
					Error("Cannot remove alert from cache")
			}
		}

	}
}

func BuildAccountURL(groupDomainID string) string {
	if groupDomainID == "" {
		return ""
	}
	gdIdBytes := []byte(groupDomainID)
	urlHash := base64.StdEncoding.EncodeToString(gdIdBytes)
	return fmt.Sprintf(C.GetProtocol()+C.GetAPPDomain()+"/profiles/accounts/%s?group=$domains&view=birdview", urlHash)
}
