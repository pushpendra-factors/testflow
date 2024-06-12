package handler

import (
	"factors/config"
	"factors/integration/linkedin_capi"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLinkedinCAPIConversionsList(c *gin.Context) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{"error": " Invalid project."})
		return
	}

	settings, errCode := store.GetStore().GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Failed to get project settings."})
		return
	}

	config := model.LinkedinCAPIConfig{
		LinkedInAccessToken: settings.IntLinkedinAccessToken,
		LinkedInAdAccounts:  config.GetTokensFromStringListAsString(settings.IntLinkedinAdAccount),
	}
	if len(config.LinkedInAdAccounts) == 0 || config.LinkedInAccessToken == "" {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "no linked user found account found"})
		return
	}

	response, err := linkedin_capi.GetConversionFromLinkedCAPI(config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to get conversions "})
		return
	}

	if len(response.LinkedInCAPIConversionsResponseList) == 0 {
		c.AbortWithStatusJSON(http.StatusExpectationFailed, gin.H{"error": "nil response from paragon"})
		return
	}

	c.JSON(http.StatusOK, response)
	return

}
