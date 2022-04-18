package postgres

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) CreateSharableURLAudit(sharableURL *model.ShareableURL, agentId string) int {
	if sharableURL == nil || sharableURL.QueryID == "" {
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
