package postgres

import (
	"factors/model/model"
	"net/http"
)

// TODO: all channels response given even if no channel integration is there.
func (pg *Postgres) GetKPIConfigsForAllChannels(projectID uint64, reqID string) (map[string]interface{}, int) {
	config := model.GetKPIConfigsForAllChannels()
	objectsAndProperties := pg.buildObjectAndPropertiesForAllChannel(projectID, ObjectsForAllChannels)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(objectsAndProperties)
	return config, http.StatusOK
}
