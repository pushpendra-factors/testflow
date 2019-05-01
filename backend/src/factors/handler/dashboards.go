package handler

import (
	"encoding/json"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm/dialects/postgres"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type DashboardUnitRequestPayload struct {
	Title        string  `json:"title"`
	Query        M.Query `json:"query"`
	Presentation string  `json:"presentation"`
}

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

func GetDashboardUnitsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get dashboard units failed. Invalid project."})
		return
	}

	dashboardId, err := strconv.ParseUint(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		log.WithError(err).Error("Get dashboard units failed. Invalid dashboard.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}

	dashboardUnits, errCode := M.GetDashboardUnits(projectId, dashboardId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get dashboard units failed."})
		return
	}

	c.JSON(http.StatusFound, dashboardUnits)
}

func CreateDashboardUnitHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Get dashboard units failed. Invalid project."})
		return
	}
	logCtx := log.WithField("project_id", projectId)

	dashboardId, err := strconv.ParseUint(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		logCtx.WithError(err).Error("Get dashboard units failed. Invalid dashboard.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}
	logCtx = logCtx.WithField("dashboard_id", dashboardId)

	var requestPayload DashboardUnitRequestPayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Get dashboard units failed. Json decode failed."
		logCtx.WithError(err).Error(errMsg)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	queryValid, errMsg := M.IsValidQuery(&requestPayload.Query)
	if !queryValid {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	queryAsJson, err := json.Marshal(requestPayload.Query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to marshal query.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(projectId, &M.DashboardUnit{
		DashboardId:  dashboardId,
		Query:        postgres.Jsonb{queryAsJson},
		Title:        requestPayload.Title,
		Presentation: requestPayload.Presentation,
	})
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(http.StatusCreated, dashboardUnit)
}
