package postgres

import (
	"factors/model/model"
	"net/http"
)

func (pg *Postgres) GetKPIConfigsForLinkedin(projectID uint64, reqID string) (map[string]interface{}, int) {
	projectIDInString := []string{string(projectID)}
	settings, errCode := pg.GetLinkedinEnabledProjectSettingsForProjects(projectIDInString)
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(settings) == 0 {
		return nil, http.StatusOK
	}
	config := model.GetKPIConfigsForLinkedin()
	linkedinObjectsAndProperties := pg.buildObjectAndPropertiesForLinkedin(projectID, model.ObjectsForLinkedin)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(linkedinObjectsAndProperties)
	return config, http.StatusOK
}
