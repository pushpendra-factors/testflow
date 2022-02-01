package memsql

import (
	"factors/model/model"
	"net/http"
)

func (store *MemSQL) GetKPIConfigsForAllChannels(projectID uint64, reqID string) (map[string]interface{}, int) {
	return model.GetKPIConfigsForAllChannels(), http.StatusOK
}
