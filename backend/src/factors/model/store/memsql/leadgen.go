package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"
)

func (store *MemSQL) GetLeadgenSettingsForProject(projectID uint64) ([]model.LeadgenSettings, error) {
	db := C.GetServices().Db
	leadgenSettings := make([]model.LeadgenSettings, 0)
	err := db.Model(model.LeadgenSettings{}).Where("project_id = ?", projectID).Find(&leadgenSettings).Error
	return leadgenSettings, err
}

func (store *MemSQL) GetLeadgenSettings() ([]model.LeadgenSettings, error) {
	db := C.GetServices().Db
	leadgenSettings := make([]model.LeadgenSettings, 0)
	err := db.Model(model.LeadgenSettings{}).Find(&leadgenSettings).Error
	return leadgenSettings, err
}
func (store *MemSQL) UpdateRowRead(projectID uint64, source int, rowRead int64) (int, error) {
	db := C.GetServices().Db
	updateMap := map[string]interface{}{
		"row_read":   rowRead,
		"updated_at": time.Now().UTC(),
	}
	err := db.Model(model.LeadgenSettings{}).Where("project_id = ? and source = ?", projectID, source).Updates(updateMap).Error
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusAccepted, nil
}
