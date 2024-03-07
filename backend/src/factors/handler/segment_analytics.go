package handler


import (
	// "errors"
	// C "factors/config"
	// H "factors/handler/helpers"
	// mid "factors/middleware"
	"factors/model/model"
	// "factors/model/store"
	// U "factors/util"
	// "fmt"
	"net/http"
	// "strconv"
	// "strings"
	// "time"

	"github.com/gin-gonic/gin"
	// log "github.com/sirupsen/logrus"
)

// GetSegmentAnalyticsConfigHandler godoc
// @Summary To get config for the segment analytics.
// @Tags SegmentAnalytics
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} json model.SegmentAnalyticsConfig
// @Router /{project_id}/v1/segments/analytics/config [get]
func GetSegmentAnalyticsConfigHandler(c *gin.Context) {
	results := make([]map[string]interface{}, 0)
	result := map[string]interface{} {
		"wid_g_id":   1,
		"wid_g_d_name": "Marketing Engagement Analysis",
		"wids": []map[string]interface{}{
			{
				"id": "5dc44775-a5aa-43e9-9157-b52575373b79",
				"d_name": "Opportunities",
			},
		},
	}
	results = append(results, result)
	responseData := gin.H{
		"result": results,
	}
	
	// Return JSON response
	c.JSON(http.StatusOK, responseData)
}

// AddNewWidgetToWidgetGroupHandler godoc.
// @Summary To add widget to segment widget group.
// @Tags SegmentAnalytics
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param widget_group_id path integer true "Widget group ID"
// @Param query body model.SegmentAnalyticsWidget true "Segment analytics widget payload."
// @Success 200 {string} json "{"result": model.SegmentAnalyticsWidget}"
// @Router /{project_id}/segments/analytics/widget_group/{widget_group_id}/widgets [post]
func AddNewWidgetToWidgetGroupHandler(c *gin.Context) {
	result := map[string]interface{} {
		"d_name": "Opportunities",
        "id": "5dc44775-a5aa-43e9-9157-b52575373b79",
        "q_ty": model.QueryClassKPI,
        "q_me": "Salesforce Opportunity",
	}
	// TODO - change to object later. Add kpi Query to object as well? 
	// if whole query object will be given. we have to rethink of object. hence keeping it to just these fields.

	responseData := gin.H {
		"result": result,
	}

	// Return JSON response
	c.JSON(http.StatusOK, responseData)
}



// EditSegmentAnalyticsWidgetHandler godoc.
// @Summary To edit segment widget.
// @Tags SegmentAnalytics
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param widget_group_id path integer true "Widget group ID"
// @Param query body model.SegmentAnalyticsWidget true "Segment analytics widget payload."
// @Success 200 {string} json "{"result": model.SegmentAnalyticsWidget}"
// @Router /{project_id}/segments/analytics/widget_group/{widget_group_id}/widgets/{widget_id} [patch]
func EditSegmentAnalyticsWidgetHandler(c *gin.Context) {
	result := map[string]interface{} {
		// ID Shouldnt be passed as request object.
		"d_name": "Opportunities 1",
        "id": "5dc44775-a5aa-43e9-9157-b52575373b79", 
        "q_ty": model.QueryClassKPI, // Not editable
        "q_me": "Hubspot Opportunity",
	}

	responseData := gin.H {
		"result": result,
	}

	// Return JSON response
	c.JSON(http.StatusOK, responseData)
}


// ExecuteSegmentQueryHandler godoc.
// @Summary To run segment widget group query.
// @Tags SegmentAnalytics, Query
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param segment_id path integer true "Segment ID"
// @Param id path integer true "Widget group ID"
// @Param query body model. true "Segment analytics query payload."
// @Success 200 {string} json "{"result": model.SegmentAnalyticsWidget}"
// @Router /{project_id}/segments/{segment_id}/analytics/widget_group/{id}/query [post]
func ExecuteSegmentQueryHandler(c *gin.Context) {

	row := make([]interface{}, 0)
	row = append(row, 365)
	rows := make([][]interface{}, 0)
	rows = append(rows, row)
	results := make([]model.QueryResult, 0)

	result1 := model.QueryResult {}
	result1.Headers = []string{ "hubspot_opportunities" }
	result1.Rows = rows

	results = append(results, result1)

	responseData := gin.H {
		"result": results,
	}

	// Return JSON response
	c.JSON(http.StatusOK, responseData)
}
