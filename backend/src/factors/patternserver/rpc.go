package patternserver

import (
	"errors"
	"factors/pattern"
	U "factors/util"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	RPCServiceName                          = "ps"
	RPCEndpoint                             = "/rpc"
	OperationNameGetPatterns                = "GetPatterns"
	OperationNameGetSeenEventProperties     = "GetSeenEventProperties"
	OperationNameGetSeenEventPropertyValues = "GetSeenEventPropertyValues"
	OperationNameGetCountOfPattern          = "GetCountOfPattern"
	OperationNameGetProjectModelsIntervals  = "GetProjectModelsIntervals"
	OperationNameGetSeenUserProperties      = "GetSeenUserProperties"
	OperationNameGetSeenUserPropertyValues  = "GetSeenUserPropertyValues"
	Separator                               = "."
)

type GenericRPCResp struct {
	ProjectId uint64 `json:"pid"`
	ModelId   uint64 `json:"mid"`
	Ignored   bool   `json:"ignored"`
	Error     error  `json:"error"`
}

type GetPatternsRequest struct {
	ProjectId  uint64 `json:"pid"`
	ModelId    uint64 `json:"mid"`
	StartEvent string `json:"se"` // optional
	EndEvent   string `json:"ee"` // optional
}

type GetPatternsResponse struct {
	GenericRPCResp
	Patterns []*pattern.Pattern `json:"ps"`
}

type filterPattern func(p *pattern.Pattern, startEvent, endEvent string) bool

func filterByStart(p *pattern.Pattern, startEvent, endEvent string) bool {
	return strings.Compare(startEvent, p.EventNames[0]) == 0
}

func filterByEnd(p *pattern.Pattern, startEvent, endEvent string) bool {
	pLen := len(p.EventNames)
	return strings.Compare(endEvent, p.EventNames[pLen-1]) == 0
}

func filterByStartAndEnd(p *pattern.Pattern, startEvent, endEvent string) bool {
	return filterByStart(p, startEvent, "") && filterByEnd(p, "", endEvent)
}

func getFilter(startEvent, endEvent string) filterPattern {
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

func (ps *PatternServer) GetPatterns(r *http.Request, args *GetPatternsRequest, result *GetPatternsResponse) error {
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

	filterPatterns := getFilter(args.StartEvent, args.EndEvent)

	patternsToReturn := make([]*pattern.Pattern, 0, 0)

	// fetch in go routines to optimize
	for _, chunkId := range chunksToServe {
		chunkPatterns, err := ps.store.GetPatterns(args.ProjectId, modelId, chunkId)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"pid": args.ProjectId,
				"mid": modelId,
				"cid": chunkId,
			}).Error("Failed To get chunk patterns")
			continue
		}
		if filterPatterns == nil {
			patternsToReturn = append(patternsToReturn, chunkPatterns...)
		} else {
			for _, chunkP := range chunkPatterns {
				if filterPatterns(chunkP, args.StartEvent, args.EndEvent) {
					patternsToReturn = append(patternsToReturn, chunkP)
				}
			}
		}
	}

	result.ProjectId = args.ProjectId
	result.ModelId = modelId
	result.Patterns = patternsToReturn

	return nil
}

type GetSeenEventPropertiesRequest struct {
	ProjectId uint64 `json:"pid"`
	ModelId   uint64 `json:"mid"` // Optional, if not passed latest modelId will be used
	EventName string `json:"en"`
}

type GetSeenEvenPropertiesResponse struct {
	GenericRPCResp
	EventProperties map[string][]string `json:"eps"`
}

func (ps *PatternServer) GetSeenEventProperties(r *http.Request, args *GetSeenEventPropertiesRequest, result *GetSeenEvenPropertiesResponse) error {

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

	eventInfo, _ := (*userAndEventsInfo.EventPropertiesInfoMap)[args.EventName]
	for nprop := range eventInfo.NumericPropertyKeys {
		numericalProperties = append(numericalProperties, nprop)
	}
	for cprop, _ := range eventInfo.CategoricalPropertyKeyValues {
		categoricalProperties = append(categoricalProperties, cprop)
	}

	resp := make(map[string][]string)
	resp["numerical"] = numericalProperties
	resp["categorical"] = categoricalProperties
	result.ProjectId = args.ProjectId
	result.ModelId = modelId
	result.EventProperties = resp
	return nil
}

type GetSeenEventPropertyValuesRequest struct {
	ProjectId    uint64 `json:"pid"`
	ModelId      uint64 `json:"mid"` // Optional, if not passed latest modelId will be used
	EventName    string `json:"en"`
	PropertyName string `json:"pn"`
}

type GetSeenEventPropertyValuesResponse struct {
	GenericRPCResp
	PropertyValues []string `json:"pvs"`
}

func (ps *PatternServer) GetSeenEventPropertyValues(r *http.Request, args *GetSeenEventPropertyValuesRequest, result *GetSeenEventPropertyValuesResponse) error {

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

	eventInfo, _ := (*userAndEventsInfoMap.EventPropertiesInfoMap)[args.EventName]
	propValuesMap, ok := eventInfo.CategoricalPropertyKeyValues[args.PropertyName]
	if !ok {
		err := errors.New("PropertyValues not found")
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

type GetProjectModelIntervalsRequest struct {
	ProjectId uint64 `json:"pid"`
}

type GetProjectModelIntervalsResponse struct {
	GenericRPCResp
	ModelInfo []ModelInfo `json:"intervals"`
}

func (ps *PatternServer) GetProjectModelsIntervals(r *http.Request, args *GetProjectModelIntervalsRequest, result *GetProjectModelIntervalsResponse) error {
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
	GenericRPCResp
}

func (ps *PatternServer) GetCountOfPattern(r *http.Request, args *GetCountOfPatternRequest, result *GetCountOfPatternResponse) error {
	return nil
}

type GetSeenUserPropertiesRequest struct {
	ProjectId uint64 `json:"pid"`
	ModelId   uint64 `json:"mid"`
}

type GetSeenUserPropertiesResponse struct {
	GenericRPCResp
	UserProperties map[string][]string `json:"ups"`
}

func (ps *PatternServer) GetSeenUserProperties(r *http.Request, args *GetSeenUserPropertiesRequest, result *GetSeenUserPropertiesResponse) error {

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

	for nprop := range userAndEventsInfo.UserPropertiesInfo.NumericPropertyKeys {
		numericalProperties = append(numericalProperties, nprop)
	}
	for cprop := range userAndEventsInfo.UserPropertiesInfo.CategoricalPropertyKeyValues {
		categoricalProperties = append(categoricalProperties, cprop)
	}

	props := make(map[string][]string)
	props["numerical"] = numericalProperties
	props["categorical"] = categoricalProperties

	result.ProjectId = args.ProjectId
	result.ModelId = modelId
	result.UserProperties = props
	return nil
}

type GetSeenUserPropertyValuesRequest struct {
	ProjectId    uint64 `json:"pid"`
	ModelId      uint64 `json:"mid"`
	PropertyName string `json:"pn"`
}
type GetSeenUserPropertyValuesResponse struct {
	GenericRPCResp
	PropertyValues []string `json:"pvs"`
}

func (ps *PatternServer) GetSeenUserPropertyValues(r *http.Request, args *GetSeenUserPropertyValuesRequest, result *GetSeenUserPropertyValuesResponse) error {
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

	propValuesMap, ok := userAndEventsInfoMap.UserPropertiesInfo.CategoricalPropertyKeyValues[args.PropertyName]
	if !ok {
		err := errors.New("PropertyValues not found")
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
