package v1

import (
	C "factors/config"
	DD "factors/default_data"
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

func CreateMarketoIntegration(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	fivetranIntegrations, err := store.GetStore().GetAllActiveFiveTranMapping(projectID, model.MarketoIntegration)
	if err != nil || len(fivetranIntegrations) > 0 {
		return nil, http.StatusConflict, "", "Integration already exists", true
	}
	statusCode, errMsg, connectorId, schemaId := fivetran.FiveTranCreateMarketoConnector(projectID)
	if statusCode != http.StatusCreated {
		logCtx.Error("BingAds Connector Creation Failed - " + errMsg)
		return nil, http.StatusInternalServerError, "", "Connector Creation Failed", true
	}
	statusCode, errMsg, redirectUri := fivetran.FiveTranCreateConnectorCard(connectorId)
	if statusCode != http.StatusOK {
		logCtx.Error("BingAds Connector Create Connector Card Failed - " + errMsg)
		return nil, http.StatusInternalServerError, "", "Connector Card Failed", true
	}
	statusCode, errMsg, _, accounts := fivetran.FiveTranGetConnector(connectorId)
	if statusCode == http.StatusOK {
		err = store.GetStore().PostFiveTranMapping(projectID, model.MarketoIntegration, connectorId, schemaId, accounts)
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

func EnableMarketoIntegration(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	connectorId, schemaId, err := store.GetStore().GetLatestFiveTranMapping(projectID, model.MarketoIntegration)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch connector id from db")
		return nil, http.StatusNotFound, "", err.Error(), true
	}
	statusCode, errMsg := fivetran.FiveTranReloadConnectorSchema(connectorId)
	if statusCode != http.StatusOK {
		logCtx.Error("Marketo Connector Reload Schema Failed - " + errMsg)
		return nil, http.StatusInternalServerError, "", "Reload Schema Failed", true
	}
	statusCode, errMsg = fivetran.FiveTranPatchMarketoConnectorSchema(connectorId, schemaId)
	if statusCode != http.StatusOK {
		logCtx.Error("Marketo Connector Patch schema Failed - " + errMsg)
		return nil, http.StatusInternalServerError, "", "Schema Patch Failed", true
	}
	statusCode, msg := fivetran.FiveTranPatchConnector(connectorId)
	if statusCode == http.StatusOK {
		statusCode, _, _, accounts := fivetran.FiveTranGetConnector(connectorId)
		if statusCode == http.StatusOK {
			err := store.GetStore().EnableFiveTranMapping(projectID, model.MarketoIntegration, connectorId, accounts)
			if err != nil {
				logCtx.WithError(err).Error("Failed to enable connector from db")
				return nil, http.StatusPartialContent, "", err.Error(), true
			}
			status := Status{
				Status: true,
			}

			isFirstTimeIntegrationDone, statusCode := DD.CheckIfFirstTimeIntegrationDone(projectID, DD.MarketoIntegrationName)
			if statusCode != http.StatusFound {
				errMsg := fmt.Sprintf("Failed during first time integration check marketo: %v", projectID)
				C.PingHealthcheckForFailure(C.HealthCheckPreBuiltCustomKPIPingID, errMsg)
			}

			if !isFirstTimeIntegrationDone {
				factory := DD.GetDefaultDataCustomKPIFactory(DD.MarketoIntegrationName)
				statusCode2 := factory.Build(projectID)
				if statusCode2 != http.StatusOK {
					errMsg := fmt.Sprintf("Failed during prebuilt marketo custom KPI creation: %v", projectID)
					C.PingHealthcheckForFailure(C.HealthCheckPreBuiltCustomKPIPingID, errMsg)
				} else {
					statusCode3 := DD.SetFirstTimeIntegrationDone(projectID, DD.MarketoIntegrationName)
					if statusCode3 != http.StatusOK {
						errMsg := fmt.Sprintf("Failed during setting first time integration done marketo: %v", projectID)
						C.PingHealthcheckForFailure(C.HealthCheckPreBuiltCustomKPIPingID, errMsg)
					}
				}
			}

			return status, http.StatusOK, "", "", false
		}
		return nil, http.StatusNotModified, "", "Get connector failed", true
	} else {
		logCtx.Error("Marketo Connector Patch For Enable Failed - " + msg)
		return nil, http.StatusInternalServerError, "", "Connector Update Failed", true
	}
}

func GetMarketoIntegration(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	connectorId, err := store.GetStore().GetFiveTranMapping(projectID, model.MarketoIntegration)
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

func DisableMarketoIntegration(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	connectorId, err := store.GetStore().GetFiveTranMapping(projectID, model.MarketoIntegration)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch connector id from db")
		return nil, http.StatusNotFound, "", err.Error(), true
	}
	statusCode, msg := fivetran.FiveTranDeleteConnector(connectorId)
	if statusCode != http.StatusOK {
		logCtx.WithError(err).Error(fmt.Sprintf("Fivetran Marketo connector deletion failed %v", projectID))
	}
	if statusCode == http.StatusOK || (statusCode == http.StatusNotFound && msg == "NotFound_Connector") {
		err := store.GetStore().DisableFiveTranMapping(projectID, model.MarketoIntegration, connectorId)
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
