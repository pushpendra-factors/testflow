package handler

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const FACTORS_CLEARBIT = "factors_clearbit"
const FACTORS_SIXSIGNAL = "factors_6Signal"

func UpdateFactorsDeanonProvider(c *gin.Context) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
	})

	if projectId == 0 {
		logCtx.Error("Failed to get projectId.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project id."})
		return
	}

	name := c.Params.ByName("name")

	var errCode int
	var errMsg string

	if name == FACTORS_CLEARBIT {
		errCode, errMsg = UpdateClearbitAsDeanonProvider(projectId)
		if errCode != http.StatusOK {
			logCtx.Error(errMsg)
			c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
			return
		}
	} else if name == FACTORS_SIXSIGNAL {
		errCode, errMsg = Update6SignalAsDeanonProvider(projectId)
		if errCode != http.StatusOK {
			logCtx.Error(errMsg)
			c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
			return
		}
	} else {
		logCtx.Error("Parameter is incorrect")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Parameter is incorrect"})
		return
	}

	c.JSON(errCode, errMsg)
}

func UpdateClearbitAsDeanonProvider(projectId int64) (int, string) {

	errCode, errMsg := store.GetStore().ProvisionClearbitAccountByAdminEmailAndDomain(projectId)
	if errCode != http.StatusOK {
		return errCode, errMsg
	}

	factorsDeanonConfig, err := json.Marshal(model.FactorsDeanonConfig{Clearbit: model.DeanonVendorConfig{TrafficFraction: 1.0}, SixSignal: model.DeanonVendorConfig{TrafficFraction: 0.0}})
	if err != nil {
		return http.StatusInternalServerError, "Failed Json Marshal of Deanon Config"
	}
	factorsDeanonConfigJson := postgres.Jsonb{RawMessage: factorsDeanonConfig}
	_, errCode = store.GetStore().UpdateProjectSettings(projectId, &model.ProjectSetting{FactorsDeanonConfig: &factorsDeanonConfigJson})
	if errCode != http.StatusAccepted {
		return errCode, "Failed update project settings"
	}

	return http.StatusOK, "Clearbit successfully setup as deanonymisation provider"

}

func Update6SignalAsDeanonProvider(projectId int64) (int, string) {

	factorsDeanonConfig, err := json.Marshal(model.FactorsDeanonConfig{Clearbit: model.DeanonVendorConfig{TrafficFraction: 0.0}, SixSignal: model.DeanonVendorConfig{TrafficFraction: 1.0}})
	if err != nil {
		return http.StatusInternalServerError, "Failed Json Marshal of Deanon Config"
	}
	factorsDeanonConfigJson := postgres.Jsonb{RawMessage: factorsDeanonConfig}
	_, errCode := store.GetStore().UpdateProjectSettings(projectId, &model.ProjectSetting{FactorsDeanonConfig: &factorsDeanonConfigJson})
	if errCode != http.StatusAccepted {
		return errCode, "Failed update project settings"
	}

	return http.StatusOK, "6Signal successfully setup as deanonymisation provider"

}
