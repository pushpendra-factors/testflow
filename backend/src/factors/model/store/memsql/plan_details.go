package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// temp
var FEATURE_FACTORS_DEANONYMISATION string = "factors_deanonymisation"

func (store *MemSQL) CreateDefaultProjectPlanMapping(projectID int64, planID int) (int, error) {
	ppMapping := model.ProjectPlanMapping{
		ProjectID: projectID,
		PlanID:    int64(planID),
	}
	db := C.GetServices().Db
	if err := db.Create(&ppMapping).Error; err != nil {

		log.WithError(err).Error("Mapping project to Free Plan Failed")
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}
func (store *MemSQL) GetPlanDetailsForProject(projectID int64) (model.PlanDetails, error) {
	db := C.GetServices().Db
	var planDetails model.PlanDetails
	err := db.Where("project_id = ?", projectID).Find(&planDetails).Error
	if err != nil {
		log.WithError(err).Error("Failed to get plan details for projectID ", projectID)
		return model.PlanDetails{}, errors.New("failed to get plan details")
	}
	return planDetails, nil
}

func (store *MemSQL) GetProjectPlanMappingforProject(projectID int64) (model.ProjectPlanMapping, error) {
	db := C.GetServices().Db
	var projectPlanMapping model.ProjectPlanMapping
	err := db.Where("project_id = ?", projectID).Find(&projectPlanMapping).Error
	if err != nil {
		log.WithError(err).Error("Failed to get project plan mapping for projectID ", projectID)
		return model.ProjectPlanMapping{}, errors.New("Failed to get project plan mapping")
	}
	return projectPlanMapping, nil
}

func (store *MemSQL) GetPlanDetailsFromPlanId(id int64) (model.PlanDetails, int, string, error) {
	logFields := log.Fields{
		"plan_id": id,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var planDetails model.PlanDetails
	db := C.GetServices().Db
	err := db.Where("id = ?", id).Find(&planDetails).Error
	if err != nil {
		errMsg := "failed to fetch plan details"
		logCtx.WithError(err).Error(errMsg)
		return planDetails, http.StatusInternalServerError, errMsg, err
	}

	return planDetails, http.StatusFound, "", nil
}

func (store *MemSQL) GetFeatureListForProject(projectID int64) (*model.DisplayPlanDetails, int,
	string, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	projectPlanMapping, err := store.GetProjectPlanMappingforProject(projectID)
	if err != nil {
		logCtx.Error(err)
		return nil, http.StatusInternalServerError, "Failed to get project plan mapping", err
	}
	planDetails, errCode, errMsg, err := store.GetPlanDetailsFromPlanId(projectPlanMapping.PlanID)
	if err != nil || errCode != http.StatusFound {
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, errMsg, err
	}

	return store.GetDisplayablePlanDetails(projectPlanMapping, planDetails)
}
func (store *MemSQL) GetDisplayablePlanDetails(ppMap model.ProjectPlanMapping, planDetails model.PlanDetails) (
	*model.DisplayPlanDetails, int, string, error) {
	logCtx := log.WithFields(log.Fields{
		"project_id": ppMap.ProjectID,
	})

	var addOns model.OverWrite
	if ppMap.OverWrite != nil {
		err := U.DecodePostgresJsonbToStructType(ppMap.OverWrite, &addOns)
		if err != nil {
			errMsg := "failed to decode postgres jsonb object"
			logCtx.WithError(err).Error(errMsg)
			return nil, http.StatusBadRequest, errMsg, err
		}
	}
	var enabledAddOns model.OverWrite
	for _, feature := range addOns {
		if feature.IsEnabledFeature {
			enabledAddOns = append(enabledAddOns, feature)
		}
	}
	var sixSignalInfo model.SixSignalInfo
	isDeanonymisationEnabled, _, err := store.GetFeatureStatusForProjectV2(ppMap.ProjectID, FEATURE_FACTORS_DEANONYMISATION)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get status for six signal")
		return nil, http.StatusInternalServerError, "Failed to get status for six signal", err
	}
	if isDeanonymisationEnabled {
		sixSignalInfo, err = store.GetSixSignalInfoForProject(ppMap.ProjectID)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get six signal info")
			return nil, http.StatusInternalServerError, "Failed to get six signal info", err
		}
	}

	obj := model.DisplayPlanDetails{
		ProjectID:     ppMap.ProjectID,
		Plan:          planDetails,
		DisplayName:   getDisplayNameForPlan(planDetails.Name),
		AddOns:        enabledAddOns,
		LastRenewedOn: ppMap.LastRenewedOn,
		SixSignalInfo: sixSignalInfo,
	}

	return &obj, http.StatusFound, "", nil
}
func getDisplayNameForPlan(planName string) string {
	return strings.Title(strings.ToLower(planName))
}
func (store *MemSQL) GetSixSignalInfoForProject(projectID int64) (model.SixSignalInfo, error) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	// metering logic here
	timeZoneString, statusCode := store.GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		logCtx.Warn(" Failed to get Timezone for six signal count")
		timeZoneString = U.TimeZoneStringIST
	}
	monthYearString := U.GetCurrentMonthYear(timeZoneString)
	sixSignalCount, err := model.GetSixSignalMonthlyUniqueEnrichmentCount(projectID, monthYearString)
	if err != nil {
		logCtx.WithError(err).Error(" Failed to get six signal count")
		return model.SixSignalInfo{}, errors.New("Failed to get six signal count")
	}
	sixSignalLimit, err := store.GetFeatureLimitForProject(projectID, FEATURE_FACTORS_DEANONYMISATION)
	log.Info("sumit log ", sixSignalCount, sixSignalLimit)
	if err != nil {
		logCtx.WithError(err).Error(" Failed to get six signal limit")
		return model.SixSignalInfo{}, errors.New("Failed to get six signal limit")
	}

	sixSignalInfo := model.SixSignalInfo{
		IsEnabled: true,
		Usage:     sixSignalCount,
		Limit:     sixSignalLimit,
	}
	return sixSignalInfo, nil

}
func (store *MemSQL) GetPlanDetailsAndAddonsForProject(projectID int64) (model.FeatureList, model.OverWrite, error) {

	var featureList model.FeatureList
	var addOns model.OverWrite

	projectPlanMaping, err := store.GetProjectPlanMappingforProject(projectID)
	if err != nil {
		log.Error(err)
		return featureList, addOns, err
	}
	planDetails, _, _, err := store.GetPlanDetailsFromPlanId(projectPlanMaping.PlanID)
	if err != nil {
		log.Error(err)
		return featureList, addOns, err
	}
	if planDetails.FeatureList != nil {
		err = U.DecodePostgresJsonbToStructType(planDetails.FeatureList, &featureList)
		if err != nil && err.Error() != "Empty jsonb object" {
			log.WithError(err).Error("Failed to decode plan details.")
			return featureList, addOns, err
		}
	}

	if projectPlanMaping.OverWrite != nil {
		err = U.DecodePostgresJsonbToStructType(projectPlanMaping.OverWrite, &addOns)
		if err != nil && err.Error() != "Empty jsonb object" {
			log.WithError(err).Error("Failed to decode project plan mapping.")
			return featureList, addOns, err
		}
	}

	return featureList, addOns, nil

}
func (store *MemSQL) GetAllProjectIdsUsingPlanId(id int64) ([]int64, int, string, error) {
	logFields := log.Fields{
		"plan_id": id,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var ppMap []model.ProjectPlanMapping
	pids := make([]int64, 0)
	db := C.GetServices().Db

	err := db.Where("plan_id = ?", id).Find(&ppMap).Error
	if err != nil {
		errMsg := "failed to fetch records from db"
		logCtx.WithError(err).Error(errMsg)
		return pids, http.StatusNotFound, errMsg, err
	}

	for _, pp := range ppMap {
		pids = append(pids, pp.ProjectID)
	}

	return pids, http.StatusFound, "", nil
}

func (store *MemSQL) UpdateProjectPlanMappingField(projectID int64, planType string) (int,
	string, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	planID, err := GetPlanIDFromString(planType)
	if err != nil {
		return http.StatusInternalServerError, "failed to upgrade plan type", err
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	err = db.Table("project_plan_mappings").Where("project_id = ?", projectID).Updates(map[string]interface{}{"plan_id": planID, "last_renewed_on": time.Now().UTC()}).Error
	if err != nil {
		errMsg := "failed to upgrade plan type"
		logCtx.WithError(err).Error(errMsg)
		return http.StatusNotAcceptable, errMsg, err
	}

	if planType == model.PLAN_FREE {
		// making addons nil if plan is moved to free
		addOns := model.OverWrite{}
		_, err := store.UpdateAddonsForProject(projectID, addOns)
		if err != nil {
			errMsg := "failed to remove addons for project"
			logCtx.WithError(err).Error(errMsg)
			return http.StatusNotAcceptable, errMsg, err
		}
	} else if planType == model.PLAN_CUSTOM {
		err = store.CreateAddonsForCustomPlanForProject(projectID)
		if err != nil {
			errMsg := "failed to create addons for custom plan"
			logCtx.WithError(err).Error(errMsg)
			return http.StatusNotAcceptable, errMsg, err
		}
	}

	return http.StatusAccepted, "", nil
}
func (store *MemSQL) UpdateFeaturesForCustomPlan(projectID int64, AccountLimit int64, MtuLimit int64, AvailableFeatuers []string) (int, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	ppMap, err := store.GetProjectPlanMappingforProject(projectID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update features for custom plan")
		return http.StatusInternalServerError, err
	}
	if ppMap.PlanID != model.PLAN_ID_CUSTOM {
		logCtx.Error("This operation is only allowed in cutsom plan")
		return http.StatusBadRequest, err
	}
	featureMap := make(map[string]bool)
	for _, feature := range AvailableFeatuers {
		featureMap[feature] = true
	}
	_, addOns, err := store.GetPlanDetailsAndAddonsForProject(projectID)
	if err != nil {
		logCtx.Error("Failed to update features for custom plan")
		return http.StatusInternalServerError, err
	}
	var updatedFeatureList model.OverWrite
	for _, feature := range addOns {
		if _, exists := featureMap[feature.Name]; !exists {
			feature.IsEnabledFeature = false

		} else {
			feature.IsEnabledFeature = true
		}
		updatedFeatureList = append(updatedFeatureList, feature)
	}
	// TODO : update MTU limit after roshan's changes
	for idx, feature := range updatedFeatureList {
		if feature.Name == model.FEATURE_FACTORS_DEANONYMISATION {

			updatedFeatureList[idx].Limit = AccountLimit
		}
	}
	_, err = store.UpdateAddonsForProject(projectID, updatedFeatureList)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update features for custom plan")
		return http.StatusInternalServerError, err
	}
	return http.StatusAccepted, nil
}
func (store *MemSQL) UpdateAddonsForProject(projectID int64, addons model.OverWrite) (string, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	jsonRaw, err := U.EncodeStructTypeToPostgresJsonb(addons)
	if err != nil {
		errMsg := "failed to encode addons json"
		logCtx.WithError(err).Error(errMsg)
		return errMsg, err
	}
	err = db.Table("project_plan_mappings").Where("project_id = ?", projectID).Updates(map[string]interface{}{"over_write": jsonRaw}).Error
	if err != nil {
		errMsg := "failed to update addons"
		logCtx.WithError(err).Error(errMsg)
		return errMsg, err
	}
	return "", nil
}
func (store *MemSQL) CreateAddonsForCustomPlanForProject(projectID int64) error {
	var addOns model.OverWrite
	featureNames := model.GetAllAvailableFeatures()
	for _, featureName := range featureNames {
		var feature model.FeatureDetails
		feature.Name = featureName
		feature.IsEnabledFeature = true
		if featureName == model.FEATURE_FACTORS_DEANONYMISATION {
			// TODO : change this to const
			feature.Limit = 100
		}
		if featureName == model.INT_FACTORS_DEANONYMISATION {
			feature.IsConnected = true
		}
		addOns = append(addOns, feature)
	}
	_, err := store.UpdateAddonsForProject(projectID, addOns)
	if err != nil {
		log.WithError(err).Error("Failed to create custom plan addons")
		return err
	}
	settings, status := store.GetProjectSetting(projectID)
	if status != http.StatusFound {
		log.WithError(err).Error("Failed to update custom plan addons int status")
		return errors.New("Failed to update custom plan addons int status")
	}
	err = store.UpdateAllFeatureStatusForProject(projectID, *settings)
	if err != nil {
		log.WithError(err).Error("Failed to update custom plan addons int status")
		return err
	}
	return nil
}
func GetPlanIDFromString(planID string) (int, error) {
	switch planID {
	case model.PLAN_FREE:
		return model.PLAN_ID_FREE, nil
	case model.PLAN_STARTUP:
		return model.PLAN_ID_STARTUP, nil
	case model.PLAN_BASIC:
		return model.PLAN_ID_BASIC, nil
	case model.PLAN_PROFESSIONAL:
		return model.PLAN_ID_PROFESSIONAL, nil
	case model.PLAN_CUSTOM:
		return model.PLAN_ID_CUSTOM, nil
	default:
		return 0, errors.New("Plan type does not exist")
	}

}

func (store *MemSQL) PopulatePlanDetailsTable(planDetails model.PlanDetails) (int, error) {
	logFields := log.Fields{
		"plan_details": planDetails,
	}
	logCtx := log.WithFields(logFields)
	model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if planDetails.Name == model.PLAN_FREE {
		planDetails.ID = model.PLAN_ID_FREE
	}
	if planDetails.Name == model.PLAN_CUSTOM {
		planDetails.ID = model.PLAN_ID_CUSTOM
	}

	db := C.GetServices().Db
	err := db.Create(&planDetails).Error
	if err != nil {
		logCtx.WithError(err).Error("failed to insert data")
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

func (store *MemSQL) UpdatePlanDetailsTable(id int64, features []string, add bool) (int, error) {
	logFields := log.Fields{
		"plan_id": id,
	}
	logCtx := log.WithFields(logFields)
	model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	//Check if the feature to be updated is available or not
	for _, newF := range features {
		found := false
		for _, f := range model.GetAllAvailableFeatures() {
			if newF == f {
				found = true
				break
			}
		}
		if !found {
			logCtx.Error("update failure - unknown feature given")
			return http.StatusBadRequest, errors.New("update failure - unknown feature given")
		}
	}

	planDetails, errCode, errMsg, err := store.GetPlanDetailsFromPlanId(id)
	if err != nil || errCode != http.StatusFound {
		logCtx.WithError(err).Error(errMsg)
		return http.StatusNotFound, err
	}

	featureList := make([]string, 0)
	err = U.DecodePostgresJsonbToStructType(planDetails.FeatureList, &featureList)
	if err != nil {
		logCtx.WithError(err).Error("unable to decode jsonb for the desired plan")
		return http.StatusInternalServerError, err
	}
	var featureListJson *postgres.Jsonb
	newFeatureList := make([]string, 0)
	if add {
		newFeatureList = featureList
		// Check to avoid adding duplicate entries for same features
		for _, fl := range featureList {
			found := false
			for _, f := range features {
				if fl == f {
					found = true
					break
				}
			}
			if !found {
				newFeatureList = append(newFeatureList, fl)
			}
		}
	} else {
		for _, fl := range featureList {
			found := false
			for _, f := range features {
				if fl == f {
					found = true
					break
				}
			}
			if !found {
				newFeatureList = append(newFeatureList, fl)
			}
		}
	}
	featureListJson, err = U.EncodeStructTypeToPostgresJsonb(newFeatureList)
	if err != nil {
		logCtx.WithError(err).Error("unable to encode to jsonb for the desired plan")
		return http.StatusInternalServerError, err
	}

	db := C.GetServices().Db

	var planDets model.PlanDetails
	err = db.Model(&planDets).Updates(map[string]interface{}{"feature_list": featureListJson}).Where("id = ?", id).Error
	if err != nil {
		logCtx.WithError(err).Error("failed to insert data")
		return http.StatusInternalServerError, err
	}

	return http.StatusAccepted, nil
}
