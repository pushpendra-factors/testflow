package patternserver

import (
	"errors"
	"factors/pattern"
	U "factors/util"
	"net/http"
	"time"
)

const (
	RPCServiceName                          = "ps"
	RPCEndpoint                             = "/rpc"
	OperationNameGetPatterns                = "GetPatterns"
	OperationNameGetSeenEventProperties     = "GetSeenEventProperties"
	OperationNameGetSeenEventPropertyValues = "GetSeenEventPropertyValues"
	OperationNameGetCountOfPattern          = "GetCountOfPattern"
	OperationNameGetProjectModelsIntervals  = "GetProjectModelsIntervals"
	Separator                               = "."
)

type GenericRPCResp struct {
	ProjectId uint64 `json:"pid"`
	ModelId   uint64 `json:"mid"`
	Ignored   bool   `json:"ignored"`
	Error     error  `json:"error"`
}

type ListPatternsRequest struct {
	ProjectId  uint64    `json:"pid"`
	ModelId    uint64    `json:"mid"`
	StartEvent time.Time `json:"se"`
	EndEvent   time.Time `json:"ee"`
}

type ListPatternsResponse struct {
	GenericRPCResp
	Patterns []pattern.Pattern `json:"ps"`
}

func (ps *PatternServer) GetPatterns(r *http.Request, args *ListPatternsRequest, result *ListPatternsResponse) error {
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
	EventName    string `json:"event_name"`
	PropertyName string `json:"property_name"`
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
