package memsql

import (
	"factors/model/model"
	"net/http"
)

func (store *MemSQL) GetKPIConfigsForGoogleOrganic(projectID uint64, reqID string) (map[string]interface{}, int) {
	settings, errCode := store.GetIntGoogleOrganicProjectSettingsForProjectID(projectID)
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(settings) == 0 {
		return nil, http.StatusOK
	}
	config := model.GetKPIConfigsForGoogleOrganic()
	organicObjectsAndProperties := store.buildObjectAndPropertiesForGoogleOrganic(model.ObjectsForGoogleOrganic)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(organicObjectsAndProperties)
	return config, http.StatusOK
}
