package main

import (
	"bytes"
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/store"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const PARTIAL_LIMIT_THRESHOLD = 75
const LIMIT_PARTIAL_EXHAUSTED = "partial_limit"
const LIMIT_FULL_EXHAUSTED = "full_limit"

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
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	RedisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	slackWebhookURLForInternalAlerts := flag.String("slack_webhook_url_internal", "", "slack webhook url")

	flag.Parse()
	config := &C.Configuration{
		Env:                 *env,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *RedisPortPersistent,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
		},
		PrimaryDatastore:             *primaryDatastore,
		SlackInternalAlertWebhookUrl: *slackWebhookURLForInternalAlerts,
	}

	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	projects, status := store.GetStore().GetAllProjectIDs()
	if status != http.StatusFound {
		log.Error("Failed to get project ids ")
		return
	}

	for _, projectID := range projects {
		info, err := store.GetStore().GetSixSignalInfoForProject(projectID)
		if err != nil {
			log.WithError(err).Error("Failed to get six signal info for project ", projectID)
			continue
		}
		if info.IsEnabled {
			err = HandleAccountLimitExhaustedForInternalAlert(projectID, info.Usage, info.Limit)
			if err != nil {
				log.WithError(err).Error("Failed to handle internal alert for project ", projectID)
				continue
			}
		}

	}
}
func HandleAccountLimitExhaustedForInternalAlert(ProjectID, count, limit int64) error {
	logCtx := log.WithField("project_id", ProjectID)
	if limit == 0 {
		// avoiding divide by zero
		return nil
	}
	percentageExhausted := float64(count) / float64(limit) * 100.0
	if percentageExhausted >= float64(100) {
		err := SendInternalAlertForAccountLimitExhausted(ProjectID, LIMIT_FULL_EXHAUSTED, percentageExhausted)
		if err != nil {
			logCtx.WithError(err).Error("failed to send internal alert")
			return errors.New("failed to send internal alert")
		}
	} else if percentageExhausted >= float64(PARTIAL_LIMIT_THRESHOLD) {
		err := SendInternalAlertForAccountLimitExhausted(ProjectID, LIMIT_PARTIAL_EXHAUSTED, percentageExhausted)
		if err != nil {
			logCtx.WithError(err).Error("failed to send internal alert")
			return errors.New("failed to send internal alert")
		}
	}
	return nil
}

func SendInternalAlertForAccountLimitExhausted(ProjectID int64, limitType string, percentage float64) error {
	var err error
	project, status := store.GetStore().GetProject(ProjectID)
	if status != http.StatusFound {
		return errors.New("Failed to get Project")
	}
	message := fmt.Sprintf(`Project %s With ID %v Has Exhuasted %v %% of their monthly quota`, project.Name, ProjectID, percentage)

	shouldSend, err := shouldSendInternalAlert(ProjectID, limitType)
	if err != nil {
		return err
	}

	if shouldSend {
		err = TriggerWebhookAlertForAccountLimit(ProjectID, message)
		if err != nil {
			return err
		}

		err = setCacheKeyForInternalAlert(ProjectID, limitType)
		if err != nil {
			return err
		}

	}

	return nil
}

func TriggerWebhookAlertForAccountLimit(ProjectID int64, message string) error {
	logCtx := log.WithFields(log.Fields{
		"ProjectID": ProjectID,
		"message":   message,
	})
	url := C.GetSlackWebhookUrlForInternalAlerts()
	reqBody := map[string]interface{}{
		"text": message,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		logCtx.Error("Failed to marshal request body for slack internal alert")
		return err
	}
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "text/plain")
	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		logCtx.Error("Failed to make request to slack for sending alert")
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logCtx.Error("Invalid status code for interal alert ", resp.StatusCode)
		return errors.New("Invalid status code for interal alert")
	}
	return nil
}
func shouldSendInternalAlert(projectID int64, limit string) (bool, error) {
	logCtx := log.WithField("project_id", projectID)

	key, err := getCacheKeyForInternalAlert(projectID, limit)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key for internal alert")
		return false, err
	}
	exists, err := cacheRedis.ExistsPersistent(key)
	if err != nil {
		logCtx.WithError(err).Error("Failed to check existence for cache key for internal alert")
		return false, err
	}
	return !exists, nil
}

func getCacheKeyForInternalAlert(projectID int64, limit string) (*cacheRedis.Key, error) {
	logCtx := log.WithField("project_id", projectID)

	currentMonth := time.Now().Month().String()

	key, err := cacheRedis.NewKey(projectID, limit, currentMonth)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key for internal alert")
		return nil, err
	}

	return key, nil
}

func setCacheKeyForInternalAlert(projectID int64, limit string) error {
	logCtx := log.WithField("project_id", projectID)

	key, err := getCacheKeyForInternalAlert(projectID, limit)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key for internal alert")
		return err
	}
	// one month
	expiry := float64(2678400)
	err = cacheRedis.SetPersistent(key, "true", expiry)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key for internal alert")
		return err
	}
	return nil
}
