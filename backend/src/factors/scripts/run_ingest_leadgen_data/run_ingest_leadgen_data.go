package main

import (
	"context"
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
	ProjectID uint64 `json:"project_id"`
	ErrCode   int    `json:"err_code"`
	ErrMsg    string `json:"err_msg"`
	Source    string `json:"source"`
}

func main() {
	ctx := context.Background()
	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	projectIDStringList := flag.String("project_ids", "*", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")
	jsonKey := flag.String("json_key", "", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")

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
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
		},
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
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		SentryDSN:           *sentryDSN,
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
	jsonKeyArray := []byte(*jsonKey)
	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON(jsonKeyArray))
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
			}
			trackStatus := ""
			for _, record := range resp.Values {
				eventProperties, userProperties, errTransform := model.TransformAndGenerateTrackPayload(record, leadgenSetting.ProjectID, model.SourceAliasMapping[leadgenSetting.Source])
				if errTransform != nil {
					log.WithFields(log.Fields{"record": record, "document": leadgenSetting}).Error(err)
					trackStatus = "failed"
					err = errTransform
					break
				} else {
					payload := &SDK.TrackPayload{
						ProjectId:       leadgenSetting.ProjectID,
						EventProperties: eventProperties,
						UserProperties:  userProperties,
						RequestSource:   leadgenSetting.Source,
						Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
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
