package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// temp
var FEATURE_FACTORS_DEANONYMISATION string = "factors_deanonymisation"

func (store *MemSQL) CreateDefaultProjectPlanMapping(projectID int64, planID int, billingPlanPriceID string) (int, error) {
	ppMapping := model.ProjectPlanMapping{
		ProjectID:           projectID,
		PlanID:              int64(planID),
		BillingPlanID:       billingPlanPriceID,
		BillingLastSyncedAt: time.Now(),
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
		if err != nil && err.Error() != "Empty jsonb object" {
			errMsg := "failed to decode postgres jsonb object"
			logCtx.WithError(err).Error(errMsg)
			return nil, http.StatusBadRequest, errMsg, err
		}
	}
	enabledAddOns := make(map[string]model.FeatureDetails)

	for featureName, feature := range addOns {
		// filter disabled features
		if feature.IsEnabledFeature {
			enabledAddOns[featureName] = feature
		}
	}

	var sixSignalInfo model.SixSignalInfo
	isDeanonymisationEnabled, err := store.GetFeatureStatusForProjectV2(ppMap.ProjectID, FEATURE_FACTORS_DEANONYMISATION, false)
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

	var featureList model.FeatureList
	if planDetails.FeatureList != nil {
		err = U.DecodePostgresJsonbToStructType(planDetails.FeatureList, &featureList)
		if err != nil && err.Error() != "Empty jsonb object" {
			log.WithError(err).Error("Failed to decode plan details.")
			return nil, http.StatusInternalServerError, "Failed to decode feature list json", err
		}
	}

	// filter disabled features
	for fname, feature := range featureList {
		if !feature.IsEnabledFeature {
			delete(featureList, fname)
		}
	}

	// transform plan details to array to support backward compatibility
	transFormedFeatureList := model.TransformFeatureListMaptoFeatureListArray(featureList)
	// encode to json
	transformedFeatureListJson, err := U.EncodeStructTypeToPostgresJsonb(transFormedFeatureList)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode transformed feature list to json")
		return nil, http.StatusInternalServerError, "Failed to encode transformed feature list to json", err
	}

	transformedOverwrites := model.TransformFeatureListMaptoFeatureListArray(model.FeatureList(enabledAddOns))

	planDetails.FeatureList = transformedFeatureListJson

	obj := model.DisplayPlanDetails{
		ProjectID:     fmt.Sprint(ppMap.ProjectID),
		Plan:          planDetails,
		DisplayName:   getDisplayNameForPlan(planDetails.Name),
		AddOns:        transformedOverwrites,
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

func (store *MemSQL) GetBillingAddonsForProject(projectID int64) (model.BillingAddons, error) {
	projectPlanMaping, err := store.GetProjectPlanMappingforProject(projectID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var billingAddOns model.BillingAddons
	if projectPlanMaping.BillingAddons != nil {
		err = U.DecodePostgresJsonbToStructType(projectPlanMaping.BillingAddons, &billingAddOns)
		if err != nil && err.Error() != "Empty jsonb object" {
			log.WithError(err).Error("Failed to decode project plan mapping billing addons.")
			return nil, err
		}
	}
	return billingAddOns, nil
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
func (store *MemSQL) UpdateProjectPlanMapping(projectID int64, planMapping *model.ProjectPlanMapping) int {

	updateFields := make(map[string]interface{}, 0)
	db := C.GetServices().Db
	if planMapping.BillingPlanID != "" {
		updateFields["plan_id"] = planMapping.PlanID
		updateFields["billing_plan_id"] = planMapping.BillingPlanID
		updateFields["billing_last_synced_at"] = planMapping.BillingLastSyncedAt
		updateFields["billing_addons"] = planMapping.BillingAddons

	}

	if planMapping.PlanID != model.PLAN_ID_CUSTOM {
		planMapping.OverWrite = nil
		updateFields["over_write"] = planMapping.OverWrite
	}

	ppMapping, err := store.GetProjectPlanMappingforProject(projectID)
	if err != nil {
		log.WithError(err).Error(
			"Failed to get project plan mappings ")
		return http.StatusInternalServerError
	}

	err = db.Model(&model.ProjectPlanMapping{}).Where("project_id = ?", projectID).Update(updateFields).Error
	if err != nil {
		log.WithError(err).Error(
			"Failed to execute query of update project plan mappings ")
		return http.StatusInternalServerError
	}

	if ppMapping.PlanID != model.PLAN_ID_CUSTOM {
		if planMapping.PlanID == model.PLAN_ID_CUSTOM {
			err = store.CreateAddonsForCustomPlanForProject(projectID)
			if err != nil {
				log.WithError(err).Error(
					"Failed to create default addons for custom project ")
				return http.StatusInternalServerError
			}
		}
	}

	return http.StatusOK
}
func (store *MemSQL) UpdateFeaturesForCustomPlan(projectID int64, AccountLimit int64, MtuLimit int64, AvailableFeatures []string) (int, error) {
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
		logCtx.Error("This operation is only allowed in custom plan")
		return http.StatusBadRequest, err
	}
	featureMap := make(map[string]bool)
	for _, feature := range AvailableFeatures {
		featureMap[feature] = true
	}

	allFeatures := model.GetAllAvailableFeatures()
	updatedFeatureList := make(map[string]model.FeatureDetails)
	for _, featureName := range allFeatures {
		var feature model.FeatureDetails
		if _, exists := featureMap[featureName]; !exists {
			feature.IsEnabledFeature = false
		} else {
			feature.IsEnabledFeature = true
		}
		updatedFeatureList[featureName] = feature
	}
	// TODO : update MTU limit after roshan's changes
	for fname := range updatedFeatureList {
		if fname == model.FEATURE_FACTORS_DEANONYMISATION {
			var tempFeatureDetails model.FeatureDetails
			tempFeatureDetails.Limit = AccountLimit
			tempFeatureDetails.Expiry = updatedFeatureList[fname].Expiry
			tempFeatureDetails.IsEnabledFeature = updatedFeatureList[fname].IsEnabledFeature
			tempFeatureDetails.Granularity = updatedFeatureList[fname].Granularity

			updatedFeatureList[fname] = tempFeatureDetails
		}
	}

	_, existingFeatures, err := store.GetPlanDetailsAndAddonsForProject(projectID)
	if err != nil {
		logCtx.WithError(err)
	}

	updatedFeatures := filterUpdatedFeatures(existingFeatures, updatedFeatureList)

	_, err = store.UpdateAddonsForProject(projectID, updatedFeatureList)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update features for custom plan")
		return http.StatusInternalServerError, err
	}

	// Call Action on Feature Update
	err = store.OnFeatureEnableOrDisableHook(projectID, updatedFeatures)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update configs on plan update")
	}

	return http.StatusAccepted, nil
}

func filterUpdatedFeatures(existingFeaturesMap map[string]model.FeatureDetails,
	newFeaturesMap map[string]model.FeatureDetails) map[string]model.FeatureDetails {
	allFeatures := model.GetAllAvailableFeatures()
	updatedFeatures := make(map[string]model.FeatureDetails)
	for _, featureName := range allFeatures {
		if existingFeaturesMap[featureName].IsEnabledFeature != newFeaturesMap[featureName].IsEnabledFeature {
			updatedFeatures[featureName] = newFeaturesMap[featureName]
		}
	}
	return updatedFeatures
}

func (store *MemSQL) OnFeatureEnableOrDisableHook(projectID int64, updatedFeatures map[string]model.FeatureDetails) error {

	var enabledFeatures, disabledFeatures []string

	for featureName, config := range updatedFeatures {
		switch config.IsEnabledFeature {
		case true:
			enabledFeatures = append(enabledFeatures, featureName)
			break
		case false:
			disabledFeatures = append(disabledFeatures, featureName)
			break
		default:
			break
		}
	}

	for _, featureName := range enabledFeatures {
		switch featureName {
		case model.FEATURE_ACCOUNT_SCORING:
			err := store.UpdateTimelineConfigForEngagementScoring(projectID, true)
			if err != nil {
				return err
			}
			break
		default:
			break
		}
	}
	for _, featureName := range disabledFeatures {
		switch featureName {
		case model.FEATURE_ACCOUNT_SCORING:
			err := store.UpdateTimelineConfigForEngagementScoring(projectID, false)
			if err != nil {
				return err
			}
			break
		default:
			break
		}
	}
	return nil
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
	logFields := log.Fields{
		"project_id": projectID,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// enable free plan features
	basicPlanDetails, errCode, errMsg, err := store.GetPlanDetailsFromPlanId(model.PLAN_ID_BASIC)
	if err != nil || errCode != http.StatusFound {
		logCtx.WithError(err).Error(errMsg)
		return err
	}
	var basicFeatureList model.FeatureList
	if basicPlanDetails.FeatureList != nil {
		err = U.DecodePostgresJsonbToStructType(basicPlanDetails.FeatureList, &basicFeatureList)
		if err != nil && err.Error() != "Empty jsonb object" {
			logCtx.WithError(err).Error("Failed to decode plan details.")
			return err
		}
	}

	overWrite := make(map[string]model.FeatureDetails)
	for fname := range basicFeatureList {
		overWrite[fname] = basicFeatureList[fname]
		if fname == model.FEATURE_FACTORS_DEANONYMISATION {
			updatedFeature := overWrite[fname]
			updatedFeature.Limit = 500
			overWrite[fname] = updatedFeature
		}
	}

	_, err = store.UpdateAddonsForProject(projectID, overWrite)
	if err != nil {
		log.WithError(err).Error("Failed to create custom plan addons")
		return err
	}
	return nil
}
func GetPlanIDFromString(planID string) (int, error) {
	switch planID {
	case model.PLAN_FREE:
		return model.PLAN_ID_FREE, nil
	case model.PLAN_GROWTH:
		return model.PLAN_ID_GROWTH, nil
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

	planID, err := GetPlanIDFromString(planDetails.Name)
	if err != nil {
		logCtx.WithError(err).Error("Failed to assign plan id")
	}
	planDetails.ID = int64(planID)
	db := C.GetServices().Db
	err = db.Create(&planDetails).Error
	if err != nil {
		logCtx.WithError(err).Error("failed to insert data")
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

func (store *MemSQL) GetAllProjectIdsUsingPaidPlan() ([]int64, int, string, error) {

	var ppMap []model.ProjectPlanMapping
	pids := make([]int64, 0)
	db := C.GetServices().Db

	err := db.Where("plan_id IN (?,?,?,?)", model.PLAN_ID_BASIC, model.PLAN_ID_GROWTH, model.PLAN_ID_PROFESSIONAL, model.PLAN_ID_CUSTOM).Find(&ppMap).Error
	if err != nil {
		errMsg := "failed to fetch records from db"
		log.WithError(err).Error(errMsg)
		return pids, http.StatusNotFound, errMsg, err
	}

	for _, pp := range ppMap {
		pids = append(pids, pp.ProjectID)
	}

	return pids, http.StatusFound, "", nil
}
