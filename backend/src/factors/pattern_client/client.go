package patternclient

import (
	"bytes"
	"factors/pattern"
	"fmt"

	"net/http"
	"sync"

	C "factors/config"

	"github.com/gorilla/rpc/json"
	log "github.com/sirupsen/logrus"
)

const (
	RPCServiceName                          = "ps"
	RPCEndpoint                             = "/rpc"
	OperationNameGetAllPatterns             = "GetAllPatterns"
	OperationNameGetPatterns                = "GetPatterns"
	OperationNameGetSeenEventProperties     = "GetSeenEventProperties"
	OperationNameGetSeenEventPropertyValues = "GetSeenEventPropertyValues"
	OperationNameGetCountOfPattern          = "GetCountOfPattern"
	OperationNameGetProjectModelsIntervals  = "GetProjectModelsIntervals"
	OperationNameGetSeenUserProperties      = "GetSeenUserProperties"
	OperationNameGetSeenUserPropertyValues  = "GetSeenUserPropertyValues"
	OperationNameGetUserAndEventsInfo       = "GetUserAndEventsInfo"
	Separator                               = "."
)

type GenericRPCResp struct {
	ProjectId uint64 `json:"pid"`
	ModelId   uint64 `json:"mid"`
	Ignored   bool   `json:"ignored"`
	Error     error  `json:"error"`
}

type GetAllPatternsRequest struct {
	ProjectId  uint64 `json:"pid"`
	ModelId    uint64 `json:"mid"`
	StartEvent string `json:"se"` // optional filter.
	EndEvent   string `json:"ee"` // optional filter.
}

type GetAllPatternsResponse struct {
	GenericRPCResp
	Patterns []*pattern.Pattern `json:"ps"`
}

type GetPatternsRequest struct {
	ProjectId     uint64     `json:"pid"`
	ModelId       uint64     `json:"mid"`
	PatternEvents [][]string `json:"pe"`
}

type GetPatternsResponse struct {
	GenericRPCResp
	Patterns []*pattern.Pattern `json:"ps"`
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

type GetProjectModelIntervalsRequest struct {
	ProjectId uint64 `json:"pid"`
}

type GetProjectModelIntervalsResponse struct {
	GenericRPCResp
	ModelInfo []ModelInfo `json:"intervals"`
}

type GetSeenUserPropertiesRequest struct {
	ProjectId uint64 `json:"pid"`
	ModelId   uint64 `json:"mid"`
}

type GetSeenUserPropertiesResponse struct {
	GenericRPCResp
	UserProperties map[string][]string `json:"ups"`
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

type GetUserAndEventsInfoRequest struct {
	ProjectId uint64 `json:"pid"`
	ModelId   uint64 `json:"mid"` // Optional, if not passed latest modelId will be used
}

type GetUserAndEventsInfoResponse struct {
	GenericRPCResp
	UserAndEventsInfo pattern.UserAndEventsInfo `json:"uei"`
}

type ModelInfo struct {
	ModelId        uint64 `json:"mid"`
	ModelType      string `json:"mt"`
	StartTimestamp int64  `json:"st"`
	EndTimestamp   int64  `json:"et"`
}

func GetSeenUserProperties(reqId string, projectId, modelId uint64) (map[string][]string, error) {
	params := GetSeenUserPropertiesRequest{
		ProjectId: projectId,
		ModelId:   modelId,
	}
	paramBytes, err := json.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetSeenUserProperties, params)
	if err != nil {
		return map[string][]string{}, err
	}
	serverAddrs := C.GetServices().GetPatternServerAddresses()

	gatherResp := make(chan httpResp, len(serverAddrs))
	headers := map[string]string{
		"content-type": "application/json",
		"X-Req-Id":     reqId,
	}

	urls := make([]string, 0, 0)
	for _, serverAddr := range serverAddrs {
		url := fmt.Sprintf("http://%s%s", serverAddr, RPCEndpoint)
		urls = append(urls, url)
	}

	httpDo(http.MethodPost, urls, paramBytes, headers, gatherResp)

	userProps := make(map[string][]string)

	for r := range gatherResp {
		if r.err != nil {
			log.WithError(r.err).Error("Error Ignoring GetSeenUserProperties")
			continue
		}
		var result GetSeenUserPropertiesResponse
		defer r.resp.Body.Close()
		err = json.DecodeClientResponse(r.resp.Body, &result)
		if err != nil {
			log.WithError(err).Error("Error Decoding response Ignoring GetSeenUserProperties")
			continue
		}
		if result.Ignored {
			log.Debugln("Ignoring GetSeenUserProperties")
			continue
		}
		if result.Error != nil {
			log.WithError(result.Error).Error("Error GetSeenUserProperties")
			continue
		}

		for k, v := range result.UserProperties {
			userProps[k] = v // should merge keys rather than over riding keys ?
		}
	}

	return userProps, nil
}

func GetSeenUserPropertyValues(reqId string, projectId, modelId uint64, propertyName string) ([]string, error) {
	params := GetSeenUserPropertyValuesRequest{
		ProjectId:    projectId,
		ModelId:      modelId,
		PropertyName: propertyName,
	}

	paramBytes, err := json.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetSeenUserPropertyValues, params)
	if err != nil {
		return []string{}, err
	}

	serverAddrs := C.GetServices().GetPatternServerAddresses()

	gatherResp := make(chan httpResp, len(serverAddrs))
	headers := map[string]string{
		"content-type": "application/json",
		"X-Req-Id":     reqId,
	}

	urls := make([]string, 0, 0)
	for _, serverAddr := range serverAddrs {
		url := fmt.Sprintf("http://%s%s", serverAddr, RPCEndpoint)
		urls = append(urls, url)
	}

	httpDo(http.MethodPost, urls, paramBytes, headers, gatherResp)

	propValues := make([]string, 0, 0)

	for r := range gatherResp {
		if r.err != nil {
			log.WithError(r.err).Error("Error Ignoring GetSeenUserPropertyValuesResponse")
			continue
		}
		var result GetSeenUserPropertyValuesResponse
		defer r.resp.Body.Close()
		err = json.DecodeClientResponse(r.resp.Body, &result)
		if err != nil {
			log.WithError(err).Error("Error Decoding response Ignoring GetSeenUserPropertyValuesResponse")
			continue
		}
		if result.Ignored {
			log.Debugln("Ignoring GetSeenUserPropertyValuesResponse")
			continue
		}
		if result.Error != nil {
			log.WithError(result.Error).Error("Error GetSeenUserPropertyValuesResponse")
			continue
		}
		propValues = append(propValues, result.PropertyValues...)
	}

	return propValues, nil
}

func GetAllPatterns(reqId string, projectId, modelId uint64, startEvent, endEvent string) ([]*pattern.Pattern, error) {
	params := GetAllPatternsRequest{
		ProjectId:  projectId,
		ModelId:    modelId,
		StartEvent: startEvent,
		EndEvent:   endEvent,
	}
	paramBytes, err := json.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetAllPatterns, params)
	if err != nil {
		return []*pattern.Pattern{}, err
	}
	serverAddrs := C.GetServices().GetPatternServerAddresses()

	gatherResp := make(chan httpResp, len(serverAddrs))
	headers := map[string]string{
		"content-type": "application/json",
		"X-Req-Id":     reqId,
	}

	urls := make([]string, 0, 0)
	for _, serverAddr := range serverAddrs {
		url := fmt.Sprintf("http://%s%s", serverAddr, RPCEndpoint)
		urls = append(urls, url)
	}

	httpDo(http.MethodPost, urls, paramBytes, headers, gatherResp)

	patterns := make([]*pattern.Pattern, 0, 0)

	for r := range gatherResp {
		if r.err != nil {
			log.WithError(r.err).Error("Error Ignoring GetAllPatternsResponse")
			continue
		}
		var result GetAllPatternsResponse
		defer r.resp.Body.Close()
		err = json.DecodeClientResponse(r.resp.Body, &result)
		if err != nil {
			log.WithError(err).Error("Error Decoding response Ignoring GetAllPatternsResponse")
			continue
		}
		if result.Ignored {
			log.Debugln("Ignoring GetAllPatternsResponse")
			continue
		}
		if result.Error != nil {
			log.WithError(result.Error).Error("Error GetAllPatternsResponse")
			continue
		}
		patterns = append(patterns, result.Patterns...)
	}

	return patterns, nil
}

func GetPatterns(reqId string, projectId, modelId uint64, patternEvents [][]string) ([]*pattern.Pattern, error) {
	params := GetPatternsRequest{
		ProjectId:     projectId,
		ModelId:       modelId,
		PatternEvents: patternEvents,
	}
	paramBytes, err := json.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetPatterns, params)
	if err != nil {
		return []*pattern.Pattern{}, err
	}
	serverAddrs := C.GetServices().GetPatternServerAddresses()

	gatherResp := make(chan httpResp, len(serverAddrs))
	headers := map[string]string{
		"content-type": "application/json",
		"X-Req-Id":     reqId,
	}

	urls := make([]string, 0, 0)
	for _, serverAddr := range serverAddrs {
		url := fmt.Sprintf("http://%s%s", serverAddr, RPCEndpoint)
		urls = append(urls, url)
	}

	httpDo(http.MethodPost, urls, paramBytes, headers, gatherResp)

	patterns := make([]*pattern.Pattern, 0, 0)

	for r := range gatherResp {
		if r.err != nil {
			log.WithError(r.err).Error("Error Ignoring GetPatternsResponse")
			continue
		}
		defer r.resp.Body.Close()
		var result GetPatternsResponse
		err = json.DecodeClientResponse(r.resp.Body, &result)
		if err != nil {
			log.WithError(err).Error("Error Decoding response Ignoring GetPatternsResponse")
			continue
		}
		if result.Ignored {
			log.Debugln("Ignoring GetPatternsResponse")
			continue
		}
		if result.Error != nil {
			log.WithError(result.Error).Error("Error GetPatternsResponse")
			continue
		}
		patterns = append(patterns, result.Patterns...)
	}

	return patterns, nil
}

func GetProjectModelIntervals(reqId string, projectId uint64) ([]ModelInfo, error) {
	params := GetProjectModelIntervalsRequest{
		ProjectId: projectId,
	}
	paramBytes, err := json.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetProjectModelsIntervals, params)
	if err != nil {
		return []ModelInfo{}, err
	}

	serverAddrs := C.GetServices().GetPatternServerAddresses()

	gatherResp := make(chan httpResp, len(serverAddrs))
	headers := map[string]string{
		"content-type": "application/json",
		"X-Req-Id":     reqId,
	}

	urls := make([]string, 0, 0)
	for _, serverAddr := range serverAddrs {
		url := fmt.Sprintf("http://%s%s", serverAddr, RPCEndpoint)
		urls = append(urls, url)
	}

	httpDo(http.MethodPost, urls, paramBytes, headers, gatherResp)
	modelInfo := make([]ModelInfo, 0, 0)

	for r := range gatherResp {
		if r.err != nil {
			log.WithError(r.err).Error("Error Ignoring GetProjectModelIntervalsResponse")
			continue
		}

		var result GetProjectModelIntervalsResponse
		defer r.resp.Body.Close()
		err = json.DecodeClientResponse(r.resp.Body, &result)
		if err != nil {
			log.WithError(err).Error("Error Ignoring GetProjectModelIntervalsResponse")
			result.Error = err
			continue
		}
		if result.Ignored {
			log.Debugln("Ignoring GetProjectModelIntervalsResponse")
			continue
		}

		if result.Error != nil {
			log.WithError(result.Error).Error("Error GetProjectModelIntervalsResponse")
			continue
		}
		modelInfo = append(modelInfo, result.ModelInfo...)
	}

	return modelInfo, nil
}

func GetSeenEventPropertyValues(reqId string, projectId, modelId uint64, eventName, propertyName string) ([]string, error) {
	params := GetSeenEventPropertyValuesRequest{
		ProjectId:    projectId,
		ModelId:      modelId,
		EventName:    eventName,
		PropertyName: propertyName,
	}
	paramBytes, err := json.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetSeenEventPropertyValues, params)
	if err != nil {
		return []string{}, err
	}
	serverAddrs := C.GetServices().GetPatternServerAddresses()

	gatherResp := make(chan httpResp, len(serverAddrs))
	headers := map[string]string{
		"content-type": "application/json",
		"X-Req-Id":     reqId,
	}

	urls := make([]string, 0, 0)
	for _, serverAddr := range serverAddrs {
		url := fmt.Sprintf("http://%s%s", serverAddr, RPCEndpoint)
		urls = append(urls, url)
	}

	httpDo(http.MethodPost, urls, paramBytes, headers, gatherResp)

	resp := make([]string, 0, 0)
	for r := range gatherResp {
		if r.err != nil {
			log.WithError(r.err).Error("Error Ignoring GetSeenEventPropertyValuesResponse")
			continue
		}

		var result GetSeenEventPropertyValuesResponse
		defer r.resp.Body.Close()
		err = json.DecodeClientResponse(r.resp.Body, &result)
		if err != nil {
			result.Error = err
		}
		if result.Ignored {
			log.Debugln("Ignoring GetSeenEventPropertyValuesResponse")
			continue
		}

		if result.Error != nil {
			log.WithError(result.Error).Error("Error GetSeenEventPropertyValuesResponse")
			continue
		}
		resp = append(resp, result.PropertyValues...)
	}

	return resp, nil
}

func GetSeenEventProperties(reqId string, projectId, modelId uint64, eventName string) (map[string][]string, error) {

	params := GetSeenEventPropertiesRequest{
		ProjectId: projectId,
		ModelId:   modelId,
		EventName: eventName,
	}
	paramBytes, err := json.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetSeenEventProperties, params)
	if err != nil {
		return make(map[string][]string), err
	}

	serverAddrs := C.GetServices().GetPatternServerAddresses()

	gatherResp := make(chan httpResp, len(serverAddrs))
	headers := map[string]string{
		"content-type": "application/json",
		"X-Req-Id":     reqId,
	}

	urls := make([]string, 0, 0)
	for _, serverAddr := range serverAddrs {
		url := fmt.Sprintf("http://%s%s", serverAddr, RPCEndpoint)
		urls = append(urls, url)
	}

	httpDo(http.MethodPost, urls, paramBytes, headers, gatherResp)

	resp := make(map[string][]string)
	for r := range gatherResp {
		if r.err != nil {
			log.WithError(r.err).Error("Error Ignoring GetSeenEvenPropertiesResponse")
			continue
		}

		var result GetSeenEvenPropertiesResponse
		defer r.resp.Body.Close()
		err = json.DecodeClientResponse(r.resp.Body, &result)
		if err != nil {
			result.Error = err
		}
		if result.Ignored {
			log.Debugln("Ignoring GetSeenEvenPropertiesResponse")
			continue
		}

		if result.Error != nil {
			log.WithError(result.Error).Error("Error GetSeenEvenPropertiesResponse")
			continue
		}

		for k, v := range result.EventProperties {
			resp[k] = v // should merge keys rather than over riding keys ?
		}
	}

	return resp, nil
}

func GetUserAndEventsInfo(reqId string, projectId, modelId uint64) (*pattern.UserAndEventsInfo, uint64, error) {

	params := GetUserAndEventsInfoRequest{
		ProjectId: projectId,
		ModelId:   modelId,
	}
	paramBytes, err := json.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetUserAndEventsInfo, params)
	if err != nil {
		return nil, modelId, err
	}

	serverAddrs := C.GetServices().GetPatternServerAddresses()

	gatherResp := make(chan httpResp, len(serverAddrs))
	headers := map[string]string{
		"content-type": "application/json",
		"X-Req-Id":     reqId,
	}

	urls := make([]string, 0, 0)
	for _, serverAddr := range serverAddrs {
		url := fmt.Sprintf("http://%s%s", serverAddr, RPCEndpoint)
		urls = append(urls, url)
	}

	httpDo(http.MethodPost, urls, paramBytes, headers, gatherResp)

	var respUserAndEventsInfo pattern.UserAndEventsInfo
	// If modelId is not provided respond with the userAndEventsInfo of the latest modelId.
	var respModelId uint64
	for r := range gatherResp {
		if r.err != nil {
			log.WithError(r.err).Error("Error Ignoring GetUserAndEventsInfoResponse")
			continue
		}

		var result GetUserAndEventsInfoResponse
		defer r.resp.Body.Close()
		err = json.DecodeClientResponse(r.resp.Body, &result)
		if err != nil {
			result.Error = err
		}
		if result.Ignored {
			log.Debugln("Ignoring GetUserAndEventsInfoResponse")
			continue
		}

		if result.Error != nil {
			log.WithError(result.Error).Error("Error GetUserAndEventsInfoResponse")
			continue
		}
		respUserAndEventsInfo = result.UserAndEventsInfo
		respModelId = result.ModelId
	}

	return &respUserAndEventsInfo, respModelId, nil
}

type httpResp struct {
	resp *http.Response
	err  error
}

// TODO(Ankit):
// Do not create new httpClient for each request

// Caller must take care of closing resp.Body
func httpDo(method string, urls []string, paramBytes []byte, headers map[string]string, gatherResp chan httpResp) {
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, url := range urls {
		go func(u string) {
			defer func() {
				wg.Done()
			}()
			resp := httpResp{}
			req, err := http.NewRequest(http.MethodPost, u, bytes.NewBuffer(paramBytes))
			if err != nil {
				resp.err = err
				gatherResp <- resp
				return
			}
			for k, v := range headers {
				req.Header.Add(k, v)
			}
			client := new(http.Client)
			r, err := client.Do(req)
			resp.err = err
			resp.resp = r
			gatherResp <- resp
			return
		}(url)
	}
	wg.Wait()
	close(gatherResp)
}
