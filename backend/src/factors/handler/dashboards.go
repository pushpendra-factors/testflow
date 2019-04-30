package handler

import (
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetDashboardsHanlder(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get dashboards failed. Invalid project."})
		return
	}

	dashboards, errCode := M.GetDashboards(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get dashboards failed."})
		return
	}

	c.JSON(http.StatusFound, dashboards)
}
