package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	PW "factors/pattern_service_wrapper"
	U "factors/util"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetExplainV2EntityHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusForbidden, "", "Get Explain V2 enitity failed. Invalid project.", true
	}
	entity, errCode := store.GetStore().GetAllExplainV2EntityByProject(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, "", "Get Saved Queries failed.", true
	}

	return entity, http.StatusOK, "", "", false
}

func GetModelIdforJob(projectId int64, job_id string) (uint64, int) {

	entity, errCode := store.GetStore().GetExplainV2Entity(projectId, job_id)
	if errCode != http.StatusFound {
		return 0, errCode
	}
	log.Info("Got job id :%s, model_id: %d, entity:%v :%d", job_id, entity.ModelID, entity, errCode)
	return entity.ModelID, errCode

}

func GetFactorsHandlerV2(c *gin.Context) {
	log.Infof("Inside get factors handler")
	reqID, _ := getReqIDAndProjectID(c)
	var err error

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// projectId := U.GetScopeBxyKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	logCtx := log.WithFields(log.Fields{
		"projectId":  projectId,
		"request id": reqID,
	})

	jobId := c.Query("job_id")

	model_id, errInt := GetModelIdforJob(projectId, jobId)
	if errInt != http.StatusFound {

		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	log.Infof("inside get factors handler project id :%d : model_id:%d", projectId, model_id)
	patternMode := c.Query("pattern_mode")
	if model_id == 0 {
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	modelId := model_id
	ipParams, err := GetUserDistributionParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ps, err := PW.NewPatternServiceWrapperV2("", projectId, modelId)
	if err != nil {
		logCtx.WithError(err).Error("Pattern Service initialization failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}
	params1, _, _, _ := MapRule(ipParams[0].Rule)
	startConstraints1, endConstraints1 := parseConstraints(params1.Rule)
	params2, _, _, _ := MapRule(ipParams[1].Rule)
	startConstraints2, endConstraints2 := parseConstraints(params2.Rule)
	type EventComparison struct {
		First  uint `json:"first"`
		Second uint `json:"second"`
	}
	type EventDistribution struct {
		Base   EventComparison            `json:"base"`
		After  map[string]EventComparison `json:"after"`
		Before map[string]EventComparison `json:"before"`
	}
	type PropDistribution struct {
		Base EventComparison            `json:"base"`
		Prop map[string]EventComparison `json:"prop"`
	}
	if patternMode == "EventDistribution" {
		base1, before1, after1, _ := PW.BuildUserDistribution("", params1.EndEvent, endConstraints1, ps)
		base2, before2, after2 := uint(0), make(map[string]uint), make(map[string]uint)
		if params2.EndEvent != "" {
			base2, before2, after2, _ = PW.BuildUserDistribution("", params2.EndEvent, endConstraints2, ps)
		}
		finalResult := EventDistribution{
			Base: EventComparison{
				First:  base1,
				Second: base2,
			},
			After:  make(map[string]EventComparison),
			Before: make(map[string]EventComparison),
		}
		for key, value := range before1 {
			data, exists := finalResult.Before[key]
			if !exists {
				finalResult.Before[key] = EventComparison{}
			}
			data.First = value
			finalResult.Before[key] = data
		}
		if params2.EndEvent != "" {
			for key, value := range before2 {
				data, exists := finalResult.Before[key]
				if !exists {
					finalResult.Before[key] = EventComparison{}
				}
				data.Second = value
				finalResult.Before[key] = data
			}
		}
		for key, value := range after1 {
			data, exists := finalResult.After[key]
			if !exists {
				finalResult.After[key] = EventComparison{}
			}
			data.First = value
			finalResult.After[key] = data
		}
		if params2.EndEvent != "" {
			for key, value := range after2 {
				data, exists := finalResult.After[key]
				if !exists {
					finalResult.After[key] = EventComparison{}
				}
				data.Second = value
				finalResult.After[key] = data
			}
		}
		c.JSON(http.StatusOK, finalResult)
		return
	}
	if patternMode == "UserDistribution" {
		events1 := make([]string, 0)
		if params1.StartEvent != "" {
			events1 = append(events1, params1.StartEvent)
		}
		if params1.EndEvent != "" {
			events1 = append(events1, params1.EndEvent)
		}
		overall1, propDist1, _ := PW.BuildUserDistributionWithProperties("", events1, startConstraints1, endConstraints1, ps)
		overall2, propDist2 := uint(0), make(map[string]uint)
		events2 := make([]string, 0)
		if params2.StartEvent != "" {
			events2 = append(events2, params2.StartEvent)
		}
		if params2.EndEvent != "" {
			events2 = append(events2, params2.EndEvent)
		}
		if params2.EndEvent != "" {
			overall2, propDist2, _ = PW.BuildUserDistributionWithProperties("", events2, startConstraints2, endConstraints2, ps)
		}
		finalResult := PropDistribution{
			Base: EventComparison{
				First:  overall1,
				Second: overall2,
			},
			Prop: make(map[string]EventComparison),
		}
		for key, value := range propDist1 {
			data, exists := finalResult.Prop[key]
			if !exists {
				finalResult.Prop[key] = EventComparison{}
			}
			data.First = value
			finalResult.Prop[key] = data
		}
		if params2.EndEvent != "" {
			for key, value := range propDist2 {
				data, exists := finalResult.Prop[key]
				if !exists {
					finalResult.Prop[key] = EventComparison{}
				}
				data.Second = value
				finalResult.Prop[key] = data
			}
		}
		c.JSON(http.StatusOK, finalResult)
		return
	}
}

func PostFactorsHandlerV2(c *gin.Context) {
	log.Infof("Inside post factors handler")
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	reqID, _ := getReqIDAndProjectID(c)

	if projectId == 611 {
		fac := PW.Factors{}
		if err := json.Unmarshal([]byte(returnConstantData()), &fac); err != nil {
			return
		}
		c.JSON(http.StatusOK, fac)
		return
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
		"RequestId": reqID,
	})

	modelId := uint64(0)
	jobId := c.Query("job_id")

	modelId, errInt := GetModelIdforJob(projectId, jobId)
	if errInt != http.StatusFound {
		log.Errorf(" err integer :%d", errInt)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patternMode := c.Query("pattern_mode")
	propertyName := c.Query("debug_property_name")
	propertyValue := c.Query("debug_property_value")
	var err error
	inputType := c.Query("type")

	ipParams, err := GetcreateFactorsGoalParams(c)
	if err != nil {
		log.Errorf("Unable to create goal params")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	params, in_en, in_epr, in_upr := MapRule(ipParams.Rule)
	ps, err := PW.NewPatternServiceWrapperV2("", projectId, modelId)
	if err != nil {
		logCtx.WithError(err).Error("Pattern Service initialization failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}
	startConstraints, endConstraints := parseConstraints(params.Rule)
	if patternMode == "AllPatterns" {
		allEventPatterns := make([]string, 0)
		allPatterns, _ := ps.GetAllPatterns("", params.StartEvent, params.EndEvent)
		for _, eventPattern := range allPatterns {
			pattern := fmt.Sprintf("%v", eventPattern.PerUserCount)
			for _, eventName := range eventPattern.EventNames {
				pattern = pattern + "," + eventName
			}
			allEventPatterns = append(allEventPatterns, pattern)
		}
		c.JSON(http.StatusOK, allEventPatterns)
		return
	}
	if patternMode == "GetCount" {
		var count uint
		if params.StartEvent == "" {
			count, _ = ps.GetPerUserCount("", []string{params.EndEvent}, []P.EventConstraints{*endConstraints})
		} else {
			count, _ = ps.GetPerUserCount("", []string{params.StartEvent, params.EndEvent}, []P.EventConstraints{*startConstraints, *endConstraints})
		}
		c.JSON(http.StatusOK, count)
		return
	}
	if patternMode == "AllProperties" {
		userInfo := ps.GetUserAndEventsInfo()
		c.JSON(http.StatusOK, userInfo)
		return
	}
	if patternMode == "AllHistogram" {
		var patternHistogram *P.Pattern
		if params.StartEvent == "" {
			patternHistogram = ps.GetPattern("", []string{params.EndEvent})
		} else {
			patternHistogram = ps.GetPattern("", []string{params.StartEvent, params.EndEvent})
		}
		c.JSON(http.StatusOK, patternHistogram)
		return
	}
	debugParams := make(map[string]string)
	debugParams["PropertyName"] = propertyName
	debugParams["PropertyValue"] = propertyValue
	if results, err, debugData := PW.FactorV1(reqID,
		projectId, params.StartEvent, startConstraints,
		params.EndEvent, endConstraints, P.COUNT_TYPE_PER_USER, ps, patternMode, debugParams, in_en, in_epr, in_upr); err != nil {
		logCtx.WithError(err).Error("Factors failed.")
		if err.Error() == "root node not found or frequency 0" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No Insights Found"})
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		if patternMode != "" {
			c.JSON(http.StatusOK, debugData)
		}
		results.Type = inputType
		results.GoalRule = params
		c.JSON(http.StatusOK, results)
		return

	}
}

func CreateExplainV2EntityHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	userID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	if projectID == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	log.Info("Create function handler triggered.")

	var entity model.ExplainV2Query
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&entity); err != nil {
		errMsg := "Get Explain v2 failed. Invalid JSON."
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}

	// if len(entity.IncludeEvents) != 0 && len(entity.ExcludeEvents) != 0 {
	// 	return nil, http.StatusBadRequest, PROCESSING_FAILED, "Provide either IncludeEvents or ExcludeEvents", true
	// }

	// err := BeforeCreate(projectID)
	// if err != http.StatusOK {
	// 	return nil, http.StatusBadRequest, PROCESSING_FAILED, "Build limit reached for creating pathanalysis", true
	// }

	_, errCode, errMsg := store.GetStore().CreateExplainV2Entity(userID, projectID, &entity)
	if errCode != http.StatusCreated {
		log.WithFields(log.Fields{"document": entity, "err-message": errMsg}).Error("Failed to Explain v2 query in handler.")
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	return entity, http.StatusCreated, "", "", false
}

func DeleteSavedExplainV2EntityHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Delete explain v2 failed. Invalid project."})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Delete failed. Invalid id provided."})
		return
	}

	errCode, errMsg := store.GetStore().DeleteExplainV2Entity(projectID, id)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(errCode, gin.H{"Status": "OK"})
}
