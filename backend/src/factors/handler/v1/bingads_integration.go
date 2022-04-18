package v1

import (
	fivetran "factors/fivetran"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func CreateBingAdsIntegration(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	fivetranIntegrations, err := store.GetStore().GetAllActiveFiveTranMapping(projectID, model.BingAdsIntegration)
	if err != nil || len(fivetranIntegrations) > 0 {
		return nil, http.StatusConflict, "", "Integration already exists", true
	}
	statusCode, errMsg, connectorId, schemaId := fivetran.FiveTranCreateBingAdsConnector(projectID)
	if statusCode != http.StatusCreated {
		logCtx.Error("BingAds Connector Creation Failed - " + errMsg)
		return nil, http.StatusInternalServerError, "", "Connector Creation Failed", true
	}
	statusCode, errMsg = fivetran.FiveTranReloadConnectorSchema(connectorId)
	if statusCode != http.StatusOK {
		logCtx.Error("BingAds Connector Reload Schema Failed - " + errMsg)
		return nil, http.StatusInternalServerError, "", "Reload Schema Failed", true
	}
	statusCode, errMsg = fivetran.FiveTranPatchBingAdsConnectorSchema(connectorId, schemaId)
	if statusCode != http.StatusOK {
		logCtx.Error("BingAds Connector Patch schema Failed - " + errMsg)
		return nil, http.StatusInternalServerError, "", "Schema Patch Failed", true
	}
	statusCode, errMsg, redirectUri := fivetran.FiveTranCreateConnectorCard(connectorId)
	if statusCode != http.StatusOK {
		logCtx.Error("BingAds Connector Create Connector Card Failed - " + errMsg)
		return nil, http.StatusInternalServerError, "", "Connector Card Failed", true
	}
	statusCode, errMsg, _, accounts := fivetran.FiveTranGetConnector(connectorId)
	if statusCode == http.StatusOK {
		err = store.GetStore().PostFiveTranMapping(projectID, model.BingAdsIntegration, connectorId, schemaId, accounts)
		if err != nil {
			logCtx.WithError(err).Error("Failed to add connector id to db")
			return nil, http.StatusPartialContent, "", err.Error(), true
		}
	}

	result := IntegrationRedirect{
		RedirectUri: redirectUri,
	}
	return result, http.StatusOK, "", "", false
}

func EnableBingAdsIntegration(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	connectorId, _, err := store.GetStore().GetLatestFiveTranMapping(projectID, model.BingAdsIntegration)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch connector id from db")
		return nil, http.StatusNotFound, "", err.Error(), true
	}
	statusCode, msg := fivetran.FiveTranPatchConnector(connectorId)
	if statusCode == http.StatusOK {
		statusCode, _, status, accounts := fivetran.FiveTranGetConnector(connectorId)
		if statusCode == http.StatusOK && status == true && accounts != "" {
			err := store.GetStore().EnableFiveTranMapping(projectID, model.BingAdsIntegration, connectorId, accounts)
			if err != nil {
				logCtx.WithError(err).Error("Failed to enable connector from db")
				return nil, http.StatusPartialContent, "", err.Error(), true
			}
			status := Status{
				Status: status,
			}
			return status, http.StatusOK, "", "", false
		}
		return nil, http.StatusNotModified, "", "Get connector failed", true
	} else {
		logCtx.Error("BingAds Connector Patch For Enable Failed - " + msg)
		return nil, http.StatusInternalServerError, "", "Connector Update Failed", true
	}
}

func GetBingAdsIntegration(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	connectorId, err := store.GetStore().GetFiveTranMapping(projectID, model.BingAdsIntegration)
	if err != nil {
		return nil, http.StatusNotFound, "", err.Error(), true
	}
	statusCode, errMsg, isActive, accounts := fivetran.FiveTranGetConnector(connectorId)
	if statusCode == http.StatusOK {
		resp := IntegrationResult{
			Status:   isActive,
			Accounts: accounts,
		}
		return resp, http.StatusOK, "", "", false
	} else {
		logCtx.Error("Failed to fetch connector details - " + errMsg)
		if statusCode == http.StatusUnauthorized {
			statusCode = http.StatusBadRequest
		}
		return nil, http.StatusInternalServerError, "", "Connector Fetch Failed", true
	}
}

func DisableBingAdsIntegration(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	connectorId, err := store.GetStore().GetFiveTranMapping(projectID, model.BingAdsIntegration)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch connector id from db")
		return nil, http.StatusNotFound, "", err.Error(), true
	}
	statusCode, msg := fivetran.FiveTranDeleteConnector(connectorId)
	if statusCode != http.StatusOK {
		logCtx.WithError(err).Error(fmt.Sprintf("Fivetran Bingads connector deletion failed %v", projectID))
	}
	if statusCode == http.StatusOK || (statusCode == http.StatusNotFound && msg == "NotFound_Connector") {
		err := store.GetStore().DisableFiveTranMapping(projectID, model.BingAdsIntegration, connectorId)
		if err != nil {
			logCtx.WithError(err).Error("Failed to disable connector from db")
			return nil, http.StatusPartialContent, "", err.Error(), true
		}
		status := Status{
			Status: true,
		}
		return status, http.StatusOK, "", "", false
	} else {
		logCtx.Error("Failed to delete connector - " + msg)
		return nil, http.StatusInternalServerError, "", "Failed to delete connector", true
	}
}

type Status struct {
	Status bool `json:"status"`
}

type IntegrationResult struct {
	Status   bool   `json:"status"`
	Accounts string `json:"accounts"`
}

type IntegrationRedirect struct {
	RedirectUri string `json:"redirect_uri"`
}
