package memsql

import (
	"factors/model/model"
	U "factors/util"
	"net/http"
)

func (store *MemSQL) GetKPIConfigsForFormSubmissions(projectID uint64, reqID string) (map[string]interface{}, int) {
	config := model.KPIConfigForFormSubmissions
	config["metrics"] = model.GetMetricsForDisplayCategory(model.FormSubmissionsDisplayCategory)
	return model.AddObjectTypeToProperties(config, U.EVENT_NAME_FORM_SUBMITTED), http.StatusOK
}
