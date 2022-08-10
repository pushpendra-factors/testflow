package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"
)

func (store *MemSQL) GetLeadgenSettingsForProject(projectID int64) ([]model.LeadgenSettings, error) {
	db := C.GetServices().Db
	leadgenSettings := make([]model.LeadgenSettings, 0)
	err := db.Model(model.LeadgenSettings{}).Where("project_id = ?", projectID).Find(&leadgenSettings).Error
	if err != nil {
		return leadgenSettings, err
	}

	projects, _ := store.GetProjectsByIDs([]int64{projectID})
	for _, project := range projects {
		for index := range leadgenSettings {
			if leadgenSettings[index].ProjectID == project.ID {
				if leadgenSettings[index].Timezone == "" {
					leadgenSettings[index].Timezone = project.TimeZone
				}
			}
		}
	}
	return leadgenSettings, err
}

func (store *MemSQL) GetLeadgenSettings() ([]model.LeadgenSettings, error) {
	db := C.GetServices().Db
	leadgenSettings := make([]model.LeadgenSettings, 0)
	err := db.Model(model.LeadgenSettings{}).Find(&leadgenSettings).Error
	if err != nil {
		return leadgenSettings, err
	}

	projectIDsMap := make(map[int64]bool)
	for _, leadgenSetting := range leadgenSettings {
		projectIDsMap[leadgenSetting.ProjectID] = true
	}
	projectIDs := make([]int64, 0)
	for key, _ := range projectIDsMap {
		projectIDs = append(projectIDs, key)
	}

	projects, _ := store.GetProjectsByIDs(projectIDs)
	for _, project := range projects {
		for index := range leadgenSettings {
			if leadgenSettings[index].ProjectID == project.ID {
				if leadgenSettings[index].Timezone == "" {
					leadgenSettings[index].Timezone = project.TimeZone
				}
			}
		}
	}
	return leadgenSettings, err
}
func (store *MemSQL) UpdateRowRead(projectID int64, source int, rowRead int64) (int, error) {
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
