package handler

import (
	"encoding/json"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

/*
Test Command

Unique User:
curl -i -H 'cookie: factors-sid=eyJhdSI6IjUxMGM1NTg4LWFjMTItNDhkZS1iZDc4LTBlY2MxNjk0NDRmOSIsInBmIjoiTVRVMU1USTBPVFl3TTN4cVZVdDFOVE0zVjNoVVN6UnFhSHBKVmtoNlpWbDVZWG95VGsxZlQyTlpaWEZoWXkwM2FFZzRVekF5WkdaeVlXVnRNR1ZCV1doQmFHcGZVRGxuVjFJMFZreFhNV0ozWm13MGIxQXhlbU5yUFh3RldScEktQVZPVzF4d0FoN2MycTNtaTNsNXF4SWlWRWhhNkpEYlp4RUVVUT09In0%3D' -H "Content-Type: application/json" -i -X POST http://factors-dev.com:8080/projects/2/query -d '{"query":{"type":"unique_users","eventsCondition":"all","from":1393632004,"to":1396310325,"eventsWithProperties":[{"name":"View Project","properties":[{"entity":"event","property":"category","operator":"equals","type":"categorical","value":"Sports"},{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]},{"name":"Fund Project","properties":[{"entity":"event","property":"category","operator":"equals","type":"categorical","value":"Sports"},{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]}],"groupByProperties":[{"property":"$region","entity":"user","index":1}]}}'

Events Occurence:
curl -i -H 'cookie: factors-sid=eyJhdSI6IjUxMGM1NTg4LWFjMTItNDhkZS1iZDc4LTBlY2MxNjk0NDRmOSIsInBmIjoiTVRVMU1USTBPVFl3TTN4cVZVdDFOVE0zVjNoVVN6UnFhSHBKVmtoNlpWbDVZWG95VGsxZlQyTlpaWEZoWXkwM2FFZzRVekF5WkdaeVlXVnRNR1ZCV1doQmFHcGZVRGxuVjFJMFZreFhNV0ozWm13MGIxQXhlbU5yUFh3RldScEktQVZPVzF4d0FoN2MycTNtaTNsNXF4SWlWRWhhNkpEYlp4RUVVUT09In0%3D' -H "Content-Type: application/json" -i -X POST http://factors-dev.com:8080/projects/2/query -d '{"query":{"type":"events_occurrence","eventsCondition":"any","from":1393632004,"to":1396310325,"eventsWithProperties":[{"name":"View Project","properties":[{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]},{"name":"Fund Project","properties":[{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]}],"groupByProperties":[{"property":"$region","entity":"user","index":0},{"property":"category","entity":"event","index":1}]}}'
*/

type QueryRequestPayload struct {
	Query M.Query `json:"query"`
}

func QueryHandler(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	r := c.Request

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Query failed. Invalid project."})
		return
	}

	var requestPayload QueryRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Query failed. Json decode failed."})
		return
	}

	colNames, resultRows, err := M.Analyze(projectId, requestPayload.Query)
	if err != nil {
		logCtx.WithError(err).Error("Analyze query execution failure.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed processing query."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"headers": colNames, "rows": resultRows})
}
