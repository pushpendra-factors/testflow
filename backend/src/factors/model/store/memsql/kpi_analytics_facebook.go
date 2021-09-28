package memsql

import (
	"factors/model/model"
	"net/http"
)

func (store *MemSQL) GetKPIConfigsForFacebook(projectID uint64, reqID string) (map[string]interface{}, int) {
	_, settings, errCode := store.GetFacebookEnabledIDsAndProjectSettingsForProject([]uint64{projectID})
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(settings) == 0 {
		return nil, http.StatusOK
	}
	return model.GetKPIConfigsForLinkedin(), http.StatusOK
}
