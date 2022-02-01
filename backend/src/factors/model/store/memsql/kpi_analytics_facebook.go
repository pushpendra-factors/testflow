package memsql

import (
	"factors/model/model"
	"net/http"
	"time"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForFacebook(projectID uint64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id": reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, settings, errCode := store.GetFacebookEnabledIDsAndProjectSettingsForProject([]uint64{projectID})
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(settings) == 0 {
		return nil, http.StatusOK
	}
	config := model.GetKPIConfigsForFacebook()
	facebookObjectsAndProperties := store.buildObjectAndPropertiesForFacebook(projectID, model.ObjectsForFacebook)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(facebookObjectsAndProperties)
	return config, http.StatusOK
}
