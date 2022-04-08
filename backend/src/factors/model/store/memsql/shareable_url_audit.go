package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateSharableURLAudit(sharableURL *model.ShareableURL, agentId string) int {
	logFields := log.Fields{
		"agent_uuid":    agentId,
		"shareable_url": sharableURL,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if sharableURL == nil || sharableURL.QueryID == "" {
		logCtx.Error("Invalid sharable url")
		return http.StatusBadRequest
	}

	audit := &model.ShareableURLAudit{
		ShareID:    sharableURL.ID,
		ProjectID:  sharableURL.ProjectID,
		QueryID:    sharableURL.QueryID,
		EntityType: sharableURL.EntityType,
		ShareType:  sharableURL.ShareType,
		EntityID:   sharableURL.EntityID,
		IsDeleted:  sharableURL.IsDeleted,
		ExpiresAt:  sharableURL.ExpiresAt,
		AccessedBy: agentId,
	}

	db := C.GetServices().Db
	if err := db.Create(&audit).Error; err != nil {
		log.WithError(err).Error("CreateSharableURLAudit Failed")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
