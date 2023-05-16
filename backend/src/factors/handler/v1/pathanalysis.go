package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"factors/delta"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type PathAnalysis model.PathAnalysis

const (
	buildLimit = model.BuildLimit
	BUILDING   = "building"
	SAVED      = "saved"
)

func GetPathAnalysisEntityHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusForbidden, "", "Get path analysis enitity failed. Invalid project.", true
	}
	entity, errCode := store.GetStore().GetAllPathAnalysisEntityByProject(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, "", "Get Saved Queries failed.", true
	}

	return entity, http.StatusOK, "", "", false
}

func CreatePathAnalysisEntityHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	userID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	if projectID == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	log.Info("Create function handler triggered.")

	var entity model.PathAnalysisQueryWithReference
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&entity); err != nil {
		errMsg := "Get pathanalysis failed. Invalid JSON."
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}

	if len(entity.Query.IncludeEvents) != 0 && len(entity.Query.ExcludeEvents) != 0 {
		return nil, http.StatusBadRequest, PROCESSING_FAILED, "Provide either IncludeEvents or ExcludeEvents", true
	}

	err := BeforeCreate(projectID)
	if err != http.StatusOK {
		return nil, http.StatusBadRequest, PROCESSING_FAILED, "Build limit reached for creating pathanalysis", true
	}

	_, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(userID, projectID, &entity.Query, entity.ReferenceId)
	if errCode != http.StatusCreated {
		log.WithFields(log.Fields{"document": entity, "err-message": errMsg}).Error("Failed to create path analysis in handler.")
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	return entity, http.StatusCreated, "", "", false
}

// Function triggered before Create handler
func BeforeCreate(projectID int64) int {

	// Checks if the there are already enough projects with BUILDING status
	status := []string{BUILDING, SAVED}
	count, _, _ := store.GetStore().GetProjectCountWithStatus(projectID, status)
	if count >= buildLimit {
		log.WithFields(log.Fields{"project_id": projectID, "err-message": count}).Error("Project BUILDING Limit reached")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func DeleteSavedPathAnalysisEntityHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Delete pathanalaysis failed. Invalid project."})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Delete failed. Invalid id provided."})
		return
	}

	errCode, errMsg := store.GetStore().DeletePathAnalysisEntity(projectID, id)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(errCode, gin.H{"Status": "OK"})
}

func GetPathAnalysisData(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusForbidden, "", "Get path analysis enitity failed. Invalid project.", true
	}
	id := c.Param("id")
	if id == "" {
		return nil, http.StatusForbidden, "", "Get path analysis enitity failed. Invalid query id.", true
	}

	n := c.Query("n")
	version := c.Query("version")
	log.Info(n)
	noOfNodes := int64(20)
	if n != "" {
		noOfNodes, _ = strconv.ParseInt(n, 10, 64)
	}

	query, _ := store.GetStore().GetPathAnalysisEntity(projectID, id)
	var actualQuery model.PathAnalysisQuery
	U.DecodePostgresJsonbToStructType(query.PathAnalysisQuery, &actualQuery)

	result := delta.GetPathAnalysisData(projectID, id)
	finalResult := filterNodes(result, int(noOfNodes), actualQuery.EventType == "startswith")
	if(version == "2"){
		return convertToArray(finalResult, actualQuery.EventType == "startswith"), http.StatusOK, "", "", false
	}
	return finalResult, http.StatusOK, "", "", false
}

type ResultStruct struct {
	Key string `json:"key"`
	Count int `json:"count"`
	Percentage float64 `json:"percentage"`
}

func convertToArray(resultMap map[int]map[string]int, startsWith bool) []ResultStruct {
	resultArray := make([]ResultStruct, 0)
	for i := 1; i <= len(resultMap) ; i++ {
		othersArray := make([]ResultStruct, 0)
		if(resultMap[i] == nil){
			continue
		}
		for key, count := range resultMap[i] {
			percentage := 0.0
			if( i != 1){
				rootElement := returnRootElement(key, startsWith)
				percentage = float64(count) * 100.0 / float64(resultMap[i-1][rootElement])
			} else {
				percentage = float64(100)
			}
			if strings.Contains(key, "OTHERS"){
				othersArray = append(othersArray, ResultStruct{Key: key, Count: count, Percentage: percentage})
			} else {
				resultArray = append(resultArray, ResultStruct{Key: key, Count: count, Percentage: percentage})
			}
		}
		resultArray = append(resultArray, othersArray...)
	}
	return resultArray
}

func filterNodes(result map[int]map[string]int, n int, startsWith bool)map[int]map[string]int {
	type labelCount struct {
		label string
		count int
	}
	finalResult := make(map[int]map[string]int)
	for i := 1; i <= len(result); i++ {
		nodes := result[i]
		if len(nodes) <= n {
			finalResult[i] = nodes
		} else {
			labelCountArray := make([]labelCount, 0)
			for label, count := range nodes {
				labelCountArray = append(labelCountArray, labelCount{
					label: label,
					count: count,
				})
			}
			sort.Slice(labelCountArray, func(i, j int) bool {
				return labelCountArray[i].count > labelCountArray[j].count
			})
			totalSelectedCount := 0
			selectedNodes := make(map[string]int)
			for _, labelCount := range labelCountArray {
				if strings.Contains(labelCount.label, "OTHERS") {
					continue
				} else {
					rootEvent := returnRootElement(labelCount.label, startsWith)
					if totalSelectedCount > n || finalResult[i-1][rootEvent] == 0 {
						continue
					}
					selectedNodes[labelCount.label] = labelCount.count
					totalSelectedCount++
				}
			}
			finalResult[i] = selectedNodes
			if i >= 2 {
				for label, count := range finalResult[i-1] {
					sum := 0
					rootEvent := ""
					if startsWith == true {
						rootEvent = label + ","
					} else {
						rootEvent = "," + label
					}
					if strings.Contains(label, "OTHERS") {
						continue
					}
					for label1, count1 := range finalResult[i] {
						if strings.Contains(label1, "OTHERS") {
							continue
						} else {
							if startsWith == true {
								if strings.HasPrefix(label1, rootEvent) {
									sum += count1
								}
							} else {
								if strings.HasSuffix(label1, rootEvent) {
									sum += count1
								}
							}
						}
					}
					if startsWith == true {
						finalResult[i][rootEvent+fmt.Sprintf("%v:OTHERS", i-1)] = count - sum
					} else {
						finalResult[i][fmt.Sprintf("%v:OTHERS", i-1)+rootEvent] = count - sum
					}
				}
			}
		}
	}
	return finalResult
}

func returnRootElement(label string, startsWith bool)string{
	labelEvents := strings.Split(label, ",")
	rootEvent := ""
	if startsWith == true {
		for it, event := range labelEvents {
			if it == len(labelEvents)-1 {
				break
			}
			if rootEvent == "" {
				rootEvent = event
			} else {
				rootEvent = rootEvent + "," + event
			}
		}
	} else {
		for it, event := range labelEvents {
			if it == 0 {
				continue
			}
			if rootEvent == "" {
				rootEvent = event
			} else {
				rootEvent = rootEvent + "," + event
			}
		}
	}
	return rootEvent
}
