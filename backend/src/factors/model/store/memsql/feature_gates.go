package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// Not in use
func (store *MemSQL) GetFeaturesForProject(projectID int64) (model.FeatureGate, error) {
	var featureGate model.FeatureGate
	db := C.GetServices().Db
	err := db.Where("project_id = ?", projectID).Find(&featureGate).Error
	if err != nil {
		log.WithError(err).Error("Failed to get feature gates for projectID ", projectID)
		return model.FeatureGate{}, errors.New("Failed to get feature gates")
	}
	return featureGate, nil
}

// not in use
func (store *MemSQL) GetFeatureStatusForProject(projectID int64, featureName string) (int, error) {
	var status int
	db := C.GetServices().Db
	sqlQuery := fmt.Sprintf("SELECT %s from feature_gates where project_id = %d LIMIT 1", featureName, projectID)
	rows, err := db.Raw(sqlQuery).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to execute query on get feature status for project ", projectID)
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := db.ScanRows(rows, &status); err != nil {
			log.WithError(err).Error("Failed scanning rows on getSegmentDuplicateUsers")
			return 0, err
		}
	}
	return status, nil
}
func (store *MemSQL) GetFeatureStatusForProjectV2(projectID int64, featureName string) (bool, error) {
	if C.IsEnabledFeatureGatesV2() {
		featureList, addOns, err := store.GetPlanDetailsAndAddonsForProject(projectID)
		if err != nil {
			log.WithError(err).Error("Failed to get feature status for Project ID ", projectID)
		}
		status := isFeatureAvailableForProject(featureList, addOns, featureName)
		return status, nil
	}
	return true, nil
}

func (store *MemSQL) GetFeatureLimitForProject(projectID int64, featureName string) (int64, error) {
	featureList, addOns, err := store.GetPlanDetailsAndAddonsForProject(projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get feature limit for Project ID ", projectID)
		return 0, err
	}

	var limit int64
	isEnabled := false
	for _, feature := range featureList {
		if featureName == feature.Name {
			isEnabled = true
			limit += feature.Limit
		}
	}

	for _, feature := range addOns {
		if featureName == feature.Name {
			if !feature.IsEnabledFeature {
				return 0, errors.New("Feature is disabled for this project")
			} else {
				isEnabled = true
			}
			limit += feature.Limit
		}
	}
	if !isEnabled {
		return 0, errors.New("Feature not enabled for this project")
	}

	return limit, nil

}
func isFeatureAvailableForProject(featureList model.FeatureList, addOns model.OverWrite, featureName string) bool {
	for _, feature := range featureList {
		if featureName == feature.Name {
			return feature.IsEnabledFeature
		}
	}

	for _, feature := range addOns {
		if featureName == feature.Name {
			return feature.IsEnabledFeature
		}
	}

	return false
}

// not in use
func (store *MemSQL) UpdateStatusForFeature(projectID int64, featureName string, updateValue int) (int, error) {
	if _, ok := model.FeatureStatusTypeAlias[updateValue]; !ok {
		return http.StatusBadRequest, errors.New("undefined status")
	}
	db := C.GetServices().Db
	updatedFields := map[string]interface{}{
		featureName: updateValue,
	}
	err := db.Table("feature_gates").Where("project_id = ?", projectID).Update(updatedFields)
	if err != nil {
		log.Error("Failed to get feature gates for projectID ", projectID)
		return http.StatusInternalServerError, errors.New("Failed to get feature gates")
	}
	return http.StatusAccepted, nil
}

// not in use
func (*MemSQL) CreateDefaultFeatureGatesConfigForProject(ProjectID int64) (int, error) {
	db := C.GetServices().Db
	var featureGate model.FeatureGate
	featureGate.ProjectID = ProjectID

	err := db.Create(featureGate).Error

	if err != nil {
		log.WithError(err).Error("Failed to create feature gates dependency for Project ID ", ProjectID)
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil

}
