package postgres

import (
	"factors/model/model"
	"net/http"
)

func (pg *Postgres) GetKPIConfigsForGoogleOrganic(projectID uint64, reqID string) (map[string]interface{}, int) {
	settings, errCode := pg.GetIntGoogleOrganicProjectSettingsForProjectID(projectID)
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(settings) == 0 {
		return nil, http.StatusOK
	}
	return model.GetKPIConfigsForGoogleOrganic(), http.StatusOK
}
