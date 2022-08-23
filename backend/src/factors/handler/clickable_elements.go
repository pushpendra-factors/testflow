package handler

import (
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetClickableElementsHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get clickable elements failed. Invalid project."})
		return
	}

	clickableElements, status := store.GetStore().GetAllClickableElements(projectID)
	if status == http.StatusBadRequest || status == http.StatusInternalServerError {
		c.AbortWithStatusJSON(status, gin.H{"error": "Get clickable elements failed. Invalid project."})
		return
	}

	c.JSON(status, clickableElements)
	return
}

func ToggleClickableElementHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Failed to toggle clickable element enable/disable."})
		return
	}

	elementID := c.Params.ByName("id")

	status := store.GetStore().ToggleEnabledClickableElement(projectID, elementID)
	if status != http.StatusAccepted {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to toggle clickable element enable/disable."})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{})
	return
}
