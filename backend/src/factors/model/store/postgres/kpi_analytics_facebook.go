package postgres

import (
	"factors/model/model"
	"net/http"
)

func (pg *Postgres) GetKPIConfigsForFacebook(projectID uint64, reqID string) (map[string]interface{}, int) {
	_, settings, errCode := pg.GetFacebookEnabledIDsAndProjectSettingsForProject([]uint64{projectID})
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(settings) == 0 {
		return nil, http.StatusOK
	}
	return model.GetKPIConfigsForLinkedin(), http.StatusOK
}
