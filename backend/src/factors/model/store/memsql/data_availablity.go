package memsql

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	C "factors/config"
	"factors/model/model"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) IsDataAvailable(project_id uint64, integration string, timestamp uint64) bool {
	data, _ := store.GetLatestDataStatus([]string{integration}, project_id, false)
	if data[integration].LatestData >= timestamp {
		return true
	}
	return false
}

func shouldProcessIntegration(allIntegration bool, integrationMap map[string]bool, integration string) bool {
	if allIntegration == true || integrationMap[integration] == true {
		return true
	}
	return false
}

func shouldProcessIntegrationFromSource(integration string, existingRecordFromDb map[string]model.DataAvailability, hardRefresh bool) bool {
	dataFromDb, exists := existingRecordFromDb[integration]
	if hardRefresh == true || exists == false || dataFromDb.LastPolled.Add(time.Minute*time.Duration(C.GetConfig().DataAvailabilityExpiry)).Unix() < U.TimeNowZ().Unix() {
		return true
	}
	return false
}
func lastProcessedFromDb(integration string, existingRecordFromDb map[string]model.DataAvailability) uint64 {
	dataFromDb, exists := existingRecordFromDb[integration]
	if exists {
		return uint64(dataFromDb.LatestDataTimestamp)
	}
	return 0

}
func (store *MemSQL) IsIntegrationAvailable(projectID uint64) map[string]bool {
	result := make(map[string]bool)
	for integration, _ := range model.INTEGRATIONS {
		if integration == model.HUBSPOT {
			result[integration] = store.IsHubspotIntegrationAvailable(projectID)
		}
		if integration == model.SALESFORCE {
			result[integration] = store.IsSalesforceIntegrationAvailable(projectID)
		}
		if integration == model.BINGADS {
			result[integration] = store.IsBingIntegrationAvailable(projectID)
		}
		if integration == model.ADWORDS {
			result[integration] = store.IsAdwordsIntegrationAvailable(projectID)
		}
		if integration == model.FACEBOOK {
			result[integration] = store.IsFacebookIntegrationAvailable(projectID)
		}
		if integration == model.LINKEDIN {
			result[integration] = store.IsLinkedInIntegrationAvailable(projectID)
		}
		if integration == model.GOOGLE_ORGANIC {
			result[integration] = store.IsGoogleOrganicIntegrationAvailable(projectID)
		}
		if integration == model.MARKETO {
			result[integration] = store.IsMarketoIntegrationAvailable(projectID)
		}
		if integration == model.SESSION {
			result[integration] = true
		}
	}
	return result
}

func transformTimestampValue(integration string, timestamp int64) int64 {
	if integration == model.HUBSPOT {
		return timestamp / int64(time.Microsecond)
	}
	if integration == model.ADWORDS || integration == model.FACEBOOK || integration == model.GOOGLE_ORGANIC || integration == model.BINGADS || integration == model.LINKEDIN {
		timestampInUnix, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, fmt.Sprintf("%v", timestamp))
		return timestampInUnix.Unix()
	}
	return timestamp
}

func (store *MemSQL) FindLatestProcessedStatus(integration string, projectID uint64) uint64 {
	db := C.GetServices().Db
	query := model.INTEGRATIONS_QUERY[integration]
	rows, err := db.Raw(query, projectID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get latest timestamp")
		return 0
	}
	defer rows.Close()
	var latestData int64
	for rows.Next() {
		if err := rows.Scan(&latestData); err == nil {
			return uint64(transformTimestampValue(integration, latestData))
		} else {
			return 0
		}
	}
	return 0
}
func (store *MemSQL) GetLatestDataStatus(integrations []string, project_id uint64, hardRefresh bool) (map[string]model.DataAvailabilityStatus, error) {
	allIntegration := false
	integrationMap := make(map[string]bool)
	for _, integration := range integrations {
		if integration == "*" {
			allIntegration = true
		} else if model.INTEGRATIONS[integration] == false {
			log.Error("Invalid integration - check model.INTEGRATIONS for the valid list", integration)
			return nil, errors.New("Invalid Integration")
		} else {
			integrationMap[integration] = true
		}
	}
	// Get The already existing status from DB
	// Check what all integrations have to be run
	// Check if integration exists for this project
	integrationStatusFromDb, err := store.GetIntegrationStatusByProjectId(project_id)
	if err != http.StatusOK {
		return nil, errors.New("Failed to fetch existing integration status")
	}
	projectIntegrations := store.IsIntegrationAvailable(project_id)
	result := make(map[string]model.DataAvailabilityStatus)
	for integration, _ := range model.INTEGRATIONS {
		if shouldProcessIntegration(allIntegration, integrationMap, integration) {
			latestProcessedTimestamp := uint64(0)
			if projectIntegrations[integration] == true {
				if shouldProcessIntegrationFromSource(integration, integrationStatusFromDb, hardRefresh) {
					latestProcessedTimestamp = store.FindLatestProcessedStatus(integration, project_id)
					store.AddIntegrationStatusByProjectId(project_id, integration, latestProcessedTimestamp, "rerun")
				} else {
					latestProcessedTimestamp = lastProcessedFromDb(integration, integrationStatusFromDb)
				}
				result[integration] = model.DataAvailabilityStatus{
					IntegrationStatus: projectIntegrations[integration],
					LatestData:        latestProcessedTimestamp,
				}
			}
		}
	}
	return result, nil
}

func (store *MemSQL) GetIntegrationStatusByProjectId(project_id uint64) (map[string]model.DataAvailability, int) {
	db := C.GetServices().Db
	var dataAvailability []model.DataAvailability
	if err := db.Model(&model.DataAvailability{}).Where("project_id = ?", project_id).Find(&dataAvailability).Error; err != nil {
		log.WithError(err).Error("Failed to fetch data availability from db ")
		return nil, http.StatusInternalServerError
	}
	result := make(map[string]model.DataAvailability)
	for _, data := range dataAvailability {
		result[data.Integration] = data
	}
	return result, http.StatusOK
}

func (store *MemSQL) AddIntegrationStatusByProjectId(project_id uint64, integration string, latest_data uint64, source string) int {
	db := C.GetServices().Db
	dataAvailability := model.DataAvailability{
		ProjectID:           project_id,
		Integration:         integration,
		LatestDataTimestamp: int64(latest_data),
		LastPolled:          U.TimeNowZ(),
		Source:              source,
		CreatedAt:           U.TimeNowZ(),
		UpdatedAt:           U.TimeNowZ(),
	}
	if project_id == 0 || integration == "" || source == "" || latest_data == 0 {
		return http.StatusBadRequest
	}
	if err := db.Create(dataAvailability).Error; err != nil {
		if IsDuplicateRecordError(err) {
			updateValues := make(map[string]interface{})
			updateValues["integration"] = integration
			updateValues["latest_data_timestamp"] = latest_data
			updateValues["source"] = source
			updateValues["last_polled"] = U.TimeNowZ()
			err := db.Model(&model.DataAvailability{}).Where("project_id = ?", project_id).Update(updateValues).Error
			if err != nil {
				log.WithError(err).Error("Failed to create integration status")
				return http.StatusInternalServerError
			}
		}
		log.WithFields(log.Fields{"dataavailability": dataAvailability,
			"project_id": project_id}).WithError(err).Error("Failed to create integration status")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
