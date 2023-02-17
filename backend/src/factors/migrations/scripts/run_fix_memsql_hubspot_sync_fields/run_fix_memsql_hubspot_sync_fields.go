package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	"flag"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var appName = "fix_memslq_hubspot_sync_fields"

func main() {
	env := flag.String("env", "development", "")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	wet := flag.Bool("wet", false, "")
	flag.Parse()

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
		PrimaryDatastore: C.DatastoreTypeMemSQL,
	}
	C.InitConf(config)
	log.SetFormatter(&log.JSONFormatter{})

	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 5, 5)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize db.")
	}

	// Log queries.
	C.GetServices().Db.LogMode(true)

	faultyDocuments, err := GetFaultyHubspotContactRecords()
	if err != nil {
		return
	}

	for _, doc := range faultyDocuments {
		logCtx := log.WithField("project_id", doc.ProjectId).WithField("faulty_doc_id", doc.ID)

		if doc.Action == 2 {
			logCtx.Error("Faulty contact updated doc.")
			continue
		}

		syncID, userID, err := GetSyncIdAndUserIdOfHubspotContactCreatedEventUsingUpdatedDocumentID(doc.ProjectId, doc.ID)
		if err != nil {
			continue
		}

		logCtx = logCtx.WithField("sync_id", syncID).WithField("user_id", userID).WithField("is_wet", *wet)
		logCtx.Info("Update info.")

		// Dry run.
		if !*wet {
			continue
		}

		status := UpdateDocumentSyncDetails(doc.ProjectId, doc.ID, 2, syncID, doc.Timestamp, 1, userID)
		if status != http.StatusAccepted {
			logCtx.Error("Failed to update the sync details")
		}
	}
}

func UpdateDocumentSyncDetails(projectID int64, id string, docType int, syncId string,
	timestamp int64, action int, userID string) int {

	updates := map[string]interface{}{
		"sync_id": syncId,
		"user_id": userID,
	}

	db := C.GetServices().Db
	exec := db.Model(&model.HubspotDocument{}).
		Where("project_id = ? AND id = ? AND timestamp= ? AND action = ? AND type= ?",
			projectID, id, timestamp, action, docType).
		Updates(updates)

	log.WithFields(log.Fields{"project_id": projectID, "id": id, "doc_type": docType,
		"sync_id": syncId, "timestamp": timestamp, "action": action, "user_id": userID,
		"rows": exec.RowsAffected}).
		Info("Updating record.")

	if exec.Error != nil {
		log.WithError(exec.Error).Error("Failed to update hubspot document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func GetSyncIdAndUserIdOfHubspotContactCreatedEventUsingUpdatedDocumentID(
	projectID int64, id string) (string, string, error) {

	logCtx := log.WithField("project_id", projectID).WithField("id", id)

	// Get synced updated document of faulty created document.
	var updatedDocument model.HubspotDocument
	db := C.GetServices().Db
	err := db.Limit(1).Order("timestamp").Select("sync_id, user_id").
		Where("project_id = ? and id = ? and synced = true and sync_id IS NOT NULL and type = 2 and action = 2",
			projectID, id).
		Find(&updatedDocument).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to find synced update for the document.")
		return "", "", err
	}

	// Get userId of faulty created_document.
	var userId string
	if updatedDocument.UserId != "" {
		userId = updatedDocument.UserId
	} else {
		// get user_id of $hubspot_contact_updated event.
		event, errCode := store.GetStore().GetEventById(projectID, updatedDocument.SyncId, "")
		if errCode != http.StatusFound {
			logCtx.WithField("status", errCode).Error("Event not found for updated document's sync_id")
			return "", "", err
		}

		userId = event.UserId
	}

	// Get eventId of faulty created_document for updating sync_id
	id, err = GetHubspotContactCreatedEvent(projectID, userId)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get the contact created event.")
		return "", "", err
	}

	return id, userId, nil
}

func GetHubspotContactCreatedEvent(projectID int64, userID string) (string, error) {
	var event model.Event

	db := C.GetServices().Db
	err := db.Limit(1).Select("id").
		Where("project_id=? and user_id=? and event_name_id=(select id from event_names where project_id=? and name=? limit 1)",
			projectID, userID, projectID, util.EVENT_NAME_HUBSPOT_CONTACT_CREATED).
		Find(&event).Error
	if err != nil {
		return "", err
	}

	return event.ID, nil
}

func GetFaultyHubspotContactRecords() ([]model.HubspotDocument, error) {
	db := C.GetServices().Db

	var documents []model.HubspotDocument
	err := db.Select("project_id, id, timestamp").
		Where("created_at < '2021-06-04 20:40:00' and updated_at >= '2021-07-19 00:00:00' and synced = false and type = 2").
		Find(&documents).Error
	if err != nil {
		log.WithError(err).Error("Failed to find faulty documents.")
		return documents, err
	}

	return documents, nil
}
