package postgres

import (
	"factors/model/model"
	"net/http"
)

// TODO: all channels response given even if no channel integration is there.
func (pg *Postgres) GetKPIConfigsForAllChannels(projectID uint64, reqID string) (map[string]interface{}, int) {
	return model.GetKPIConfigsForAllChannels(), http.StatusOK
}
