package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateWebsiteAggregation(websiteAggregation model.WebsiteAggregation) (model.WebsiteAggregation, string, int) {
	logFields := log.Fields{
		"website_aggregation": websiteAggregation,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if websiteAggregation.ProjectID == 0 {
		logCtx.Warn("Invalid project ID for website aggregation")
		return model.WebsiteAggregation{}, "Invalid project ID for website aggregation", http.StatusBadRequest
	}

	db := C.GetServices().Db
	if err := db.Create(&websiteAggregation).Error; err != nil {
		logCtx.WithError(err).Error("Failed while creating website aggregation")
		return model.WebsiteAggregation{}, err.Error(), http.StatusInternalServerError
	}

	return websiteAggregation, "", http.StatusCreated

}
