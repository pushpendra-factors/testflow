package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
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

// isEnabled, error
func (store *MemSQL) GetFeatureStatusForProjectV2(projectID int64, featureName string, includeProjectSettings bool) (bool, error) {
	logCtx := log.WithField("project_id", projectID)

	featureList, addOns, err := store.GetPlanDetailsAndAddonsForProject(projectID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get feature status for Project ID ", projectID)
	}
	featureStatus := isFeatureAvailableForProject(featureList, addOns, featureName)
	if !featureStatus || !includeProjectSettings {
		return featureStatus, nil
	}
	// project settings check (AND of feature gates and project settings integration)

	switch featureName {
	case model.FEATURE_HUBSPOT:
		return store.IsHubspotIntegrationAvailable(projectID), nil
	case model.FEATURE_SALESFORCE:
		return store.IsSalesforceIntegrationAvailable(projectID), nil
	case model.FEATURE_BING_ADS:
		return store.IsBingIntegrationAvailable(projectID), nil
	case model.FEATURE_MARKETO:
		return store.IsMarketoIntegrationAvailable(projectID), nil
	case model.FEATURE_LINKEDIN:
		return store.IsLinkedInIntegrationAvailable(projectID), nil
	case model.FEATURE_GOOGLE_ADS:
		return store.IsAdwordsIntegrationAvailable(projectID), nil
	default:
		log.Error("Include Project Settings Enabled but Definition is not present for feature ",featureName)
		return featureStatus, nil
	}

	return featureStatus, nil
}

func (store *MemSQL) UpdateFeatureStatusForProject(projectID int64, feature model.FeatureDetails) (string, error) {
	_, addOns, err := store.GetPlanDetailsAndAddonsForProject(projectID)
	if err != nil {
		log.WithError(err).Error("Failed to update feature status for Project ID ", projectID)
		return "Failed to get Plan Details ", err
	}
	for idx, addOn := range addOns {
		if addOn.Name == feature.Name {
			addOns[idx] = feature
		}
	}
	errMsg, err := store.UpdateAddonsForProject(projectID, addOns)
	if err != nil {
		log.WithError(err).Error("Failed to update feature status for Project ID ", projectID)
		return errMsg, err
	}
	return "", nil
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

func (store *MemSQL) GetAllProjectsWithFeatureEnabled(featureName string, includeProjectSettings bool) ([]int64, error) {
	var enabledProjectIds []int64 = make([]int64, 0)

	projectIDs, errCode := store.GetAllProjectIDs()
	if errCode != http.StatusFound {
		err := fmt.Errorf("failed to get all projects ids to query feature enabled flag")
		log.WithField("err_code", err).Error(err)
		return nil, err
	}
	for _, projectId := range projectIDs {
		available, err := store.GetFeatureStatusForProjectV2(projectId, featureName, false)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectId, "feature": featureName}).WithError(err).Error("failed to get feature status for project ID ")
			continue
		}
		if available {
			enabledProjectIds = append(enabledProjectIds, projectId)
		}
	}
	if !includeProjectSettings {
		return enabledProjectIds, nil
	}

	projectIdsMap := make(map[int64]bool)
	for _, projectID := range enabledProjectIds {
		projectIdsMap[projectID] = true
	}
	// project settings/fivetran mappings check

	projectIdsArray := make([]int64, 0)
	switch featureName {
	case model.FEATURE_HUBSPOT:
		settings, status := store.GetAllHubspotProjectSettings()
		if status != http.StatusFound {
			return projectIdsArray, errors.New(fmt.Sprintf("Failed to get %s enabled project settings ", featureName))
		}
		for _, setting := range settings {
			if _, exists := projectIdsMap[setting.ProjectId]; exists {
				projectIdsArray = append(projectIdsArray, setting.ProjectId)
			}
		}
	case model.FEATURE_SALESFORCE:
		settings, status := store.GetAllSalesforceProjectSettings()
		if status != http.StatusFound {
			return projectIdsArray, errors.New(fmt.Sprintf("Failed to get %s enabled project settings ", featureName))
		}
		for _, setting := range settings {
			if _, exists := projectIdsMap[setting.ProjectID]; exists {
				projectIdsArray = append(projectIdsArray, setting.ProjectID)
			}
		}
	case model.FEATURE_LINKEDIN:
		settings, status := store.GetLinkedinEnabledProjectSettings()
		if status != http.StatusFound {
			return projectIdsArray, errors.New(fmt.Sprintf("Failed to get %s enabled project settings ", featureName))
		}
		for _, setting := range settings {
			projectID, _ := strconv.ParseInt(setting.ProjectId, 10, 64)
			if _, exists := projectIdsMap[projectID]; exists {
				projectIdsArray = append(projectIdsArray, projectID)
			}
		}
	case model.FEATURE_GOOGLE_ADS:
		settings, status := store.GetAllIntAdwordsProjectSettings()
		if status != http.StatusFound {
			return projectIdsArray, errors.New(fmt.Sprintf("Failed to get %s enabled project settings ", featureName))
		}
		for _, setting := range settings {
			if _, exists := projectIdsMap[setting.ProjectId]; exists {
				projectIdsArray = append(projectIdsArray, setting.ProjectId)
			}
		}
	default:
		log.Error("Include Project Settings Enabled but Definition is not present for feature ",featureName)
		return enabledProjectIds, nil
	}
	return projectIdsArray, nil

}
