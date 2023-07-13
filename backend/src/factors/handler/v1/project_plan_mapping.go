package v1

import (
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetPlanDetailsForProjectHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		errMsg := "request made for invalid project_id"
		log.Error(errMsg)
		return nil, http.StatusBadRequest, INVALID_PROJECT, errMsg, true
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	planDetails, errCode, errMsg, err := store.GetStore().GetFeatureListForProject(projectID)
	if err != nil || errCode != http.StatusFound {
		logCtx.WithError(err).Error(errMsg)
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	return planDetails, http.StatusFound, "", "", false
}

type UpgradePlanPayload struct {
	PlanType string `json:"plan_type"`
}
type UpdateCustomPlanPayload struct {
	AccountLimit      int64    `json:"account_limit"`
	MtuLimit          int64    `json:"mtu_limit"`
	ActivatedFeatures []string `json:"activated_features"`
}

func UpdateProjectPlanMappingFieldHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		errMsg := "request made for invalid project_id"
		log.Error(errMsg)
		return nil, http.StatusBadRequest, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	planType := UpgradePlanPayload{}
	err := c.BindJSON(&planType)
	if err != nil {
		return nil, http.StatusBadRequest, PROCESSING_FAILED, "Failed to decode request", true
	}
	errCode, errMsg, err := store.GetStore().UpdateProjectPlanMappingField(projectID, planType.PlanType)
	if err != nil || errCode != http.StatusAccepted {
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, PROCESSING_FAILED, errMsg, true
	}

	return nil, http.StatusOK, "", "", false
}

func UpdateCustomPlanHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		errMsg := "request made for invalid project_id"
		log.Error(errMsg)
		return nil, http.StatusBadRequest, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	payload := UpdateCustomPlanPayload{}
	err := c.BindJSON(&payload)
	if err != nil {
		log.Error(err)
		return nil, http.StatusBadRequest, PROCESSING_FAILED, "Failed to decode request", true
	}
	status, err := store.GetStore().UpdateFeaturesForCustomPlan(projectID, payload.AccountLimit, payload.MtuLimit, payload.ActivatedFeatures)
	if status != http.StatusAccepted {
		log.Error(err)
		return nil, status, PROCESSING_FAILED, "Failed to update features", true
	}
	return nil, status, "", "", false
}
