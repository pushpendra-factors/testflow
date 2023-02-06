package main

import (
	"context"
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const appName = "ingest_leadgen_data"

type Status struct {
	ProjectID int64  `json:"project_id"`
	ErrCode   int    `json:"err_code"`
	ErrMsg    string `json:"err_msg"`
	Source    string `json:"source"`
}
type Key struct {
	Type                   string `json:"type"`
	ProjectID              string `json:"project_id"`
	PrivateKeyID           string `json:"private_key_id"`
	PrivateKey             string `json:"private_key"`
	ClientEmail            string `json:"client_email"`
	ClientID               string `json:"client_id"`
	AuthURI                string `json:"auth_uri"`
	TokenURI               string `json:"token_uri"`
	AuthProvideX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL      string `json:"client_x509_cert_url"`
}

func main() {
	ctx := context.Background()
	env := flag.String("env", "development", "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	projectIDStringList := flag.String("project_ids", "*", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	jsonKey := flag.String("json_key", "", "")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "*", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	IngestionTimezoneEnabledProjectIDs := flag.String("ingestion_timezone_enabled_projects", "", "List of projectIds whose ingestion timezone is enabled.")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultHealthcheckPingID := C.HealthCheckSmartPropertiesPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)

	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)
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
		PrimaryDatastore:                   *primaryDatastore,
		RedisHost:                          *redisHost,
		RedisPort:                          *redisPort,
		RedisHostPersistent:                *redisHostPersistent,
		RedisPortPersistent:                *redisPortPersistent,
		SentryDSN:                          *sentryDSN,
		CaptureSourceInUsersTable:          *captureSourceInUsersTable,
		IngestionTimezoneEnabledProjectIDs: C.GetTokensFromStringListAsString(*IngestionTimezoneEnabledProjectIDs),
	}
	C.InitConf(config)
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to ingest leadgen data. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()
	log.Info("jsonKey ", *jsonKey)
	key := Key{}
	err = json.Unmarshal([]byte(*jsonKey), &key)
	if err != nil {
		log.Fatalf("error unmarshaling key ", err)
	}
	keyByte, err := json.Marshal(key)
	if err != nil {
		log.Fatalf("error marshaling key ", err)
	}

	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON(keyByte))
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDStringList, "")

	finalLeadgenSettings := make([]model.LeadgenSettings, 0)
	if allProjects {
		leadgenSettings, err := store.GetStore().GetLeadgenSettings()
		if err != nil {
			log.WithError(err).Error("Failed to get leadgen settings for all projects")
		}
		finalLeadgenSettings = append(finalLeadgenSettings, leadgenSettings...)
	} else {
		projectIDs := C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)
		for _, projectID := range projectIDs {
			leadgenSettings, err := store.GetStore().GetLeadgenSettingsForProject(projectID)
			if err != nil {
				log.WithField("project_id", projectID).WithError(err).Error("Failed to get leadgen settings for all projects")
			}
			finalLeadgenSettings = append(finalLeadgenSettings, leadgenSettings...)
		}
	}
	syncStatusSuccesses := make([]Status, 0)
	syncStatusFailures := make([]Status, 0)
	for _, leadgenSetting := range finalLeadgenSettings {
		_, status := store.GetStore().CreateOrGetOfflineTouchPointEventName(leadgenSetting.ProjectID)
		if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
			errMsg := "failed to create event name on leadgen data ingestion"
			log.Error(errMsg)
			syncStatusFailure := Status{
				ProjectID: leadgenSetting.ProjectID,
				ErrCode:   http.StatusInternalServerError,
				Source:    model.SourceAliasMapping[leadgenSetting.Source],
				ErrMsg:    errMsg,
			}
			syncStatusFailures = append(syncStatusFailures, syncStatusFailure)
		} else {
			spreadsheetId := leadgenSetting.SpreadsheetID
			var readRange string
			rowRead := leadgenSetting.RowRead
			readRange = fmt.Sprintf("%s!A%d:J", leadgenSetting.SheetName, leadgenSetting.RowRead+2)
			resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
			if err != nil {
				log.WithField("document", leadgenSetting).Error("Unable to retrieve data from sheet: %v", err)
				syncStatusFailure := Status{
					ProjectID: leadgenSetting.ProjectID,
					ErrCode:   http.StatusInternalServerError,
					Source:    model.SourceAliasMapping[leadgenSetting.Source],
					ErrMsg:    err.Error(),
				}
				syncStatusFailures = append(syncStatusFailures, syncStatusFailure)
			} else {
				trackStatus := ""
				for _, record := range resp.Values {
					eventProperties, userProperties, timestamp, errTransform := model.TransformAndGenerateTrackPayload(record, leadgenSetting.ProjectID, model.SourceAliasMapping[leadgenSetting.Source], leadgenSetting.Timezone)
					if errTransform != nil {
						log.WithFields(log.Fields{"record": record, "document": leadgenSetting}).Error(errTransform)
						trackStatus = "failed"
						err = errTransform
						break
					} else {
						userID, errUser := CreateOrGetUserBySource(eventProperties, userProperties, leadgenSetting.ProjectID, leadgenSetting.Source, timestamp)
						if errUser != nil {
							log.WithFields(log.Fields{"record": record, "document": leadgenSetting}).Error(errUser)
							trackStatus = "failed"
							err = errUser
							break
						}
						payload := &SDK.TrackPayload{
							ProjectId:       leadgenSetting.ProjectID,
							UserId:          userID,
							EventProperties: eventProperties,
							UserProperties:  userProperties,
							RequestSource:   leadgenSetting.Source,
							Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
							Timestamp:       timestamp,
						}
						status, _ := SDK.Track(leadgenSetting.ProjectID, payload, true, "", "")
						if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
							log.WithField("Document", payload).WithError(err).Error(fmt.Errorf("failed to offline touchpoint from leadgen data"))
							trackStatus = "failed"
							err = fmt.Errorf("failed to offline touchpoint from leadgen data")
						}
					}
					if trackStatus != "failed" {
						rowRead = rowRead + 1
					} else {
						break
					}
				}
				if trackStatus == "failed" {
					syncStatusFailure := Status{
						ProjectID: leadgenSetting.ProjectID,
						ErrCode:   http.StatusInternalServerError,
						ErrMsg:    err.Error(),
						Source:    model.SourceAliasMapping[leadgenSetting.Source],
					}
					syncStatusFailures = append(syncStatusFailures, syncStatusFailure)
				} else {
					syncStatusSuccess := Status{
						ProjectID: leadgenSetting.ProjectID,
						ErrCode:   http.StatusOK,
						ErrMsg:    "",
						Source:    model.SourceAliasMapping[leadgenSetting.Source],
					}
					syncStatusSuccesses = append(syncStatusFailures, syncStatusSuccess)
				}
				errCode, err := store.GetStore().UpdateRowRead(leadgenSetting.ProjectID, leadgenSetting.Source, rowRead)
				if errCode != http.StatusAccepted || err != nil {
					log.WithField("document", leadgenSetting).Error("Failed to update row read")
					for errCode != http.StatusAccepted || err != nil {
						errCode, err = store.GetStore().UpdateRowRead(leadgenSetting.ProjectID, leadgenSetting.Source, rowRead)
						if errCode != http.StatusAccepted || err != nil {
							log.WithField("document", leadgenSetting).Error("Failed to update row read on loop")
						}
					}
				}
			}
		}
	}
	log.Warn("End of leadgen ingestion job")
	syncStatus := map[string]interface{}{
		"Success": syncStatusSuccesses,
		"Failure": syncStatusFailures,
	}
	if *env == "production" {
		if len(syncStatusFailures) > 0 {
			C.PingHealthcheckForFailure(healthcheckPingID, syncStatus)
			return
		}
		C.PingHealthcheckForSuccess(healthcheckPingID, syncStatus)
	} else {
		log.Info(syncStatus)
	}
}

func CreateOrGetUserBySource(eventProperties map[string]interface{}, userProperties map[string]interface{}, projectID int64,
	source int, timestamp int64) (string, error) {
	email, phone, code := "", "", 0
	customerUserID, userID := "", ""
	var err error
	emailInterface, emailExists := userProperties[U.UP_EMAIL]
	if emailExists {
		email = fmt.Sprintf("%v", emailInterface)
	}
	phoneInterface, phoneExists := userProperties[U.UP_PHONE]
	if phoneExists {
		phone = fmt.Sprintf("%v", phoneInterface)
	}
	if email != "" {
		customerUserID = email
	} else if phone != "" {
		customerUserID = phone
	}

	if customerUserID == "" {
		err = errors.New("both phone and email not present")
		return userID, err
	} else {
		user, code := store.GetStore().GetUserLatestByCustomerUserId(projectID, customerUserID, source)
		if code == http.StatusFound {
			return user.ID, nil
		}
	}

	userID, code = store.GetStore().CreateUser(&model.User{
		ProjectId:      projectID,
		JoinTimestamp:  timestamp,
		CustomerUserId: customerUserID,
		Source:         &source})
	if code != http.StatusCreated {
		err = errors.New("failed to create user")
		return userID, err
	}
	return userID, nil
}
