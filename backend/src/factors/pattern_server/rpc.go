package patternserver

import (
	"encoding/json"
	"errors"
	client "factors/pattern_client"
	store "factors/pattern_server/store"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
)

type FilterPattern func(patternEvents []string, startEvent, endEvent string) bool

func filterByStart(patternEvents []string, startEvent, endEvent string) bool {
	return strings.Compare(startEvent, patternEvents[0]) == 0
}

func filterByEnd(patternEvents []string, startEvent, endEvent string) bool {
	pLen := len(patternEvents)
	return strings.Compare(endEvent, patternEvents[pLen-1]) == 0
}

func filterByStartAndEnd(patternEvents []string, startEvent, endEvent string) bool {
	return filterByStart(patternEvents, startEvent, "") && filterByEnd(patternEvents, "", endEvent)
}

func GetFilter(startEvent, endEvent string) FilterPattern {
	if startEvent != "" && endEvent != "" {
		return filterByStartAndEnd
	}

	if endEvent != "" {
		return filterByEnd
	}

	if startEvent != "" {
		return filterByStart
	}
	return nil
}

func (ps *PatternServer) GetAllPatterns(
	r *http.Request, args *client.GetAllPatternsRequest, result *client.GetAllPatternsResponse) error {
	if args == nil || args.ProjectId == 0 {
		err := errors.New("MissingParams")
		result.Error = err
		return err
	}

	modelId := args.ModelId

	if modelId == 0 {
		latestInterval, err := ps.GetProjectModelLatestInterval(args.ProjectId)
		if err != nil {
			result.Error = err
			return err
		}
		modelId = latestInterval.ModelId
	}

	log.WithFields(log.Fields{
		"pid":           args.ProjectId,
		"mid":           modelId,
		"se (optional)": args.StartEvent,
		"ee (optional)": args.EndEvent,
	}).Debugln("RPC Fetching patterns")

	chunkIds, found := ps.GetProjectModelChunks(args.ProjectId, modelId)
	if !found {
		err := errors.New("ProjectModelChunks not found")
		result.Error = err
		return err
	}

	chunksToServe := make([]string, 0, 0)
	for _, chunkId := range chunkIds {
		if ps.IsProjectModelChunkServable(args.ProjectId, modelId, chunkId) {
			chunksToServe = append(chunksToServe, chunkId)
		}
	}

	if len(chunksToServe) == 0 {
		result.Ignored = true
		return nil
	}

	filterPatterns := GetFilter(args.StartEvent, args.EndEvent)

	patternsToReturn := make([]*json.RawMessage, 0, 0)

	// fetch in go routines to optimize
	for _, chunkId := range chunksToServe {
		patternsWithMeta, err := ps.store.GetPatternsWithMeta(args.ProjectId, modelId, chunkId)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"pid": args.ProjectId,
				"mid": modelId,
				"cid": chunkId,
			}).Error("Failed To get chunk patterns")
			continue
		}
		if filterPatterns == nil {
			rawPatterns := store.GetAllRawPatterns(patternsWithMeta)
			patternsToReturn = append(patternsToReturn, rawPatterns...)
		} else {
			for _, pwm := range patternsWithMeta {
				if filterPatterns(pwm.PatternEvents, args.StartEvent, args.EndEvent) {
					patternsToReturn = append(patternsToReturn, &pwm.RawPattern)
				}
			}
		}
	}

	result.ProjectId = args.ProjectId
	result.ModelId = modelId
	result.Patterns = patternsToReturn

	return nil
}

func (ps *PatternServer) GetPatterns(
	r *http.Request, args *client.GetPatternsRequest, result *client.GetPatternsResponse) error {
	if args == nil || args.ProjectId == 0 {
		err := errors.New("MissingParams")
		result.Error = err
		return err
	}

	modelId := args.ModelId

	if modelId == 0 {
		latestInterval, err := ps.GetProjectModelLatestInterval(args.ProjectId)
		if err != nil {
			result.Error = err
			return err
		}
		modelId = latestInterval.ModelId
	}

	log.WithFields(log.Fields{
		"pid":      args.ProjectId,
		"mid":      modelId,
		"patterns": args.PatternEvents,
	}).Debugln("RPC Fetching patterns")

	chunkIds, found := ps.GetProjectModelChunks(args.ProjectId, modelId)
	if !found {
		err := errors.New("ProjectModelChunks not found")
		result.Error = err
		return err
	}

	chunksToServe := make([]string, 0, 0)
	for _, chunkId := range chunkIds {
		if ps.IsProjectModelChunkServable(args.ProjectId, modelId, chunkId) {
			chunksToServe = append(chunksToServe, chunkId)
		}
	}

	if len(chunksToServe) == 0 {
		result.Ignored = true
		return nil
	}

	patternsToReturn := make([]*json.RawMessage, 0, 0)

	// fetch in go routines to optimize
	for _, chunkId := range chunksToServe {
		patternsWithMeta, err := ps.store.GetPatternsWithMeta(args.ProjectId, modelId, chunkId)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"pid": args.ProjectId,
				"mid": modelId,
				"cid": chunkId,
			}).Error("Failed To get chunk patterns")
			continue
		}

		for _, pwm := range patternsWithMeta {
			for _, pE := range args.PatternEvents {
				if reflect.DeepEqual(pE, pwm.PatternEvents) {
					patternsToReturn = append(patternsToReturn, &pwm.RawPattern)
				}
			}
		}
	}

	result.ProjectId = args.ProjectId
	result.ModelId = modelId
	result.Patterns = patternsToReturn

	return nil
}

func (ps *PatternServer) GetSeenEventProperties(
	r *http.Request, args *client.GetSeenEventPropertiesRequest,
	result *client.GetSeenEvenPropertiesResponse) error {

	if args == nil || args.ProjectId == 0 || args.EventName == "" {
		err := errors.New("MissingParams")
		result.Error = err
		return err
	}

	modelId := args.ModelId

	if modelId == 0 {
		latestInterval, err := ps.GetProjectModelLatestInterval(args.ProjectId)
		if err != nil {
			result.Error = err
			return err
		}
		modelId = latestInterval.ModelId
	}

	if !ps.IsProjectModelServable(args.ProjectId, modelId) {
		result.Ignored = true
		return nil
	}

	userAndEventsInfo, err := ps.GetModelEventInfo(args.ProjectId, modelId)
	if err != nil {
		result.Error = err
		return err
	}
	numericalProperties := []string{}
	categoricalProperties := []string{}

	for _, dnp := range U.VISIBLE_DEFAULT_NUMERIC_EVENT_PROPERTIES {
		numericalProperties = append(numericalProperties, dnp)
	}

	if eventInfo, exists := (*userAndEventsInfo.EventPropertiesInfoMap)[args.EventName]; exists {
		for nprop := range eventInfo.NumericPropertyKeys {
			numericalProperties = append(numericalProperties, nprop)
		}
		for cprop := range eventInfo.CategoricalPropertyKeyValues {
			categoricalProperties = append(categoricalProperties, cprop)
		}
	}

	resp := make(map[string][]string)
	resp["numerical"] = numericalProperties
	resp["categorical"] = categoricalProperties
	result.ProjectId = args.ProjectId
	result.ModelId = modelId
	result.EventProperties = resp
	return nil
}

func (ps *PatternServer) GetSeenEventPropertyValues(
	r *http.Request, args *client.GetSeenEventPropertyValuesRequest,
	result *client.GetSeenEventPropertyValuesResponse) error {

	if args == nil || args.ProjectId == 0 || args.EventName == "" || args.PropertyName == "" {
		err := errors.New("MissingParams")
		result.Error = err
		return err
	}

	modelId := args.ModelId
	if modelId == 0 {
		latestInterval, err := ps.GetProjectModelLatestInterval(args.ProjectId)
		if err != nil {
			result.Error = err
			return err
		}
		modelId = latestInterval.ModelId
	}

	if !ps.IsProjectModelServable(args.ProjectId, modelId) {
		result.Ignored = true
		return nil
	}

	userAndEventsInfoMap, err := ps.GetModelEventInfo(args.ProjectId, modelId)
	if err != nil {
		result.Error = err
		return err
	}

	eventInfo, exists := (*userAndEventsInfoMap.EventPropertiesInfoMap)[args.EventName]
	if !exists {
		err := fmt.Errorf("EventInfo not found for EventName: %s", args.EventName)
		result.Error = err
		return err
	}

	propValuesMap, exists := eventInfo.CategoricalPropertyKeyValues[args.PropertyName]
	if !exists {
		err := fmt.Errorf("PropertyValues not found for EventName: %s, PropertyName: %s", args.EventName, args.PropertyName)
		result.Error = err
		return err
	}

	propValues := []string{}

	for value := range propValuesMap {
		propValues = append(propValues, value)
	}
	result.ProjectId = args.ProjectId
	result.ModelId = modelId
	result.PropertyValues = propValues
	return nil
}

func (ps *PatternServer) GetProjectModelsIntervals(
	r *http.Request, args *client.GetProjectModelIntervalsRequest,
	result *client.GetProjectModelIntervalsResponse) error {
	if args == nil || args.ProjectId == 0 {
		err := errors.New("MissingParams")
		result.Error = err
		return err
	}

	if !ps.IsProjectServable(args.ProjectId) {
		result.Ignored = true
		return nil
	}

	modelInfos, err := ps.GetProjectModelIntervals(args.ProjectId)
	if err != nil {
		result.Error = err
		return err
	}

	result.ProjectId = args.ProjectId
	result.ModelInfo = modelInfos
	return nil
}

type GetCountOfPatternRequest struct {
}

type GetCountOfPatternResponse struct {
	client.GenericRPCResp
}

func (ps *PatternServer) GetCountOfPattern(r *http.Request, args *GetCountOfPatternRequest, result *GetCountOfPatternResponse) error {
	return nil
}

func (ps *PatternServer) GetSeenUserProperties(
	r *http.Request, args *client.GetSeenUserPropertiesRequest,
	result *client.GetSeenUserPropertiesResponse) error {

	if args == nil || args.ProjectId == 0 {
		err := errors.New("MissingParams")
		result.Error = err
		return err
	}
	modelId := args.ModelId

	if modelId == 0 {
		latestInterval, err := ps.GetProjectModelLatestInterval(args.ProjectId)
		if err != nil {
			result.Error = err
			return err
		}
		modelId = latestInterval.ModelId
	}

	if !ps.IsProjectModelServable(args.ProjectId, modelId) {
		result.Ignored = true
		return nil
	}

	userAndEventsInfo, err := ps.GetModelEventInfo(args.ProjectId, modelId)
	if err != nil {
		result.Error = err
		return err
	}
	numericalProperties := []string{}
	categoricalProperties := []string{}

	if userAndEventsInfo.UserPropertiesInfo != nil {
		for nprop := range userAndEventsInfo.UserPropertiesInfo.NumericPropertyKeys {
			numericalProperties = append(numericalProperties, nprop)
		}
		for cprop := range userAndEventsInfo.UserPropertiesInfo.CategoricalPropertyKeyValues {
			categoricalProperties = append(categoricalProperties, cprop)
		}
	}

	props := make(map[string][]string)
	props["numerical"] = numericalProperties
	props["categorical"] = categoricalProperties

	result.ProjectId = args.ProjectId
	result.ModelId = modelId
	result.UserProperties = props
	return nil
}

func (ps *PatternServer) GetSeenUserPropertyValues(
	r *http.Request, args *client.GetSeenUserPropertyValuesRequest,
	result *client.GetSeenUserPropertyValuesResponse) error {
	if args == nil || args.ProjectId == 0 || args.PropertyName == "" {
		err := errors.New("MissingParams")
		result.Error = err
		return err
	}

	modelId := args.ModelId
	if modelId == 0 {
		latestInterval, err := ps.GetProjectModelLatestInterval(args.ProjectId)
		if err != nil {
			result.Error = err
			return err
		}
		modelId = latestInterval.ModelId
	}

	if !ps.IsProjectModelServable(args.ProjectId, modelId) {
		result.Ignored = true
		return nil
	}

	userAndEventsInfoMap, err := ps.GetModelEventInfo(args.ProjectId, modelId)
	if err != nil {
		result.Error = err
		return err
	}

	if userAndEventsInfoMap.UserPropertiesInfo == nil {
		err := errors.New("UserPropertyValues not found")
		result.Error = err
		return err
	}

	propValuesMap, ok := userAndEventsInfoMap.UserPropertiesInfo.CategoricalPropertyKeyValues[args.PropertyName]
	if !ok {
		err := errors.New("UserPropertyValues not found")
		result.Error = err
		return err
	}

	values := make([]string, 0, 0)
	for k := range propValuesMap {
		values = append(values, k)
	}

	result.ProjectId = args.ProjectId
	result.ModelId = modelId
	result.PropertyValues = values
	return nil
}

func (ps *PatternServer) GetUserAndEventsInfo(
	r *http.Request, args *client.GetUserAndEventsInfoRequest,
	result *client.GetUserAndEventsInfoResponse) error {

	if args == nil || args.ProjectId == 0 {
		err := errors.New("MissingParams")
		result.Error = err
		return err
	}

	modelId := args.ModelId

	if modelId == 0 {
		latestInterval, err := ps.GetProjectModelLatestInterval(args.ProjectId)
		if err != nil {
			result.Error = err
			return err
		}
		modelId = latestInterval.ModelId
	}

	if !ps.IsProjectModelServable(args.ProjectId, modelId) {
		result.Ignored = true
		return nil
	}

	userAndEventsInfo, err := ps.GetModelEventInfo(args.ProjectId, modelId)
	if err != nil {
		result.Error = err
		return err
	}
	result.ProjectId = args.ProjectId
	result.ModelId = modelId
	result.UserAndEventsInfo = userAndEventsInfo
	return nil
}
