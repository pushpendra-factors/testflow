package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

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
