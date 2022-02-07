package memsql

import (
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForAdwords(projectID uint64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	adwordsSettings, errCode := store.GetIntAdwordsProjectSettingsForProjectID(projectID)
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(adwordsSettings) == 0 {
		return nil, http.StatusOK
	}
	config := model.GetKPIConfigsForAdwords()
	adwordsObjectsAndProperties := store.buildObjectAndPropertiesForAdwords(projectID, model.ObjectsForAdwords)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(adwordsObjectsAndProperties)
	return config, http.StatusOK
}
