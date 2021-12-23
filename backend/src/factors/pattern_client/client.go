package patternclient

import (
	"bytes"
	"factors/pattern"
	"fmt"
	"time"

	"net/http"
	"sync"

	C "factors/config"

	"encoding/json"

	rpcJson "github.com/gorilla/rpc/json"
	log "github.com/sirupsen/logrus"
)

const (
	RPCServiceName                        = "ps"
	RPCEndpoint                           = "/rpc"
	OperationNameGetAllPatterns           = "GetAllPatterns"
	OperationNameGetAllContainingPatterns = "GetAllContainingPatterns"
	OperationNameGetPatterns              = "GetPatterns"
	OperationNameGetCountOfPattern        = "GetCountOfPattern"
	OperationNameGetTotalEventCount       = "GetTotalEventCount"
	Separator                             = "."
	OperationNameGetUserAndEventsInfo     = "GetUserAndEventsInfo"
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
	Patterns []*json.RawMessage `json:"ps"`
}

type GetAllContainingPatternsRequest struct {
	ProjectId uint64 `json:"pid"`
	ModelId   uint64 `json:"mid"`
	Event     string `json:"en"`
}

type GetAllContainingPatternsResponse struct {
	GenericRPCResp
	Patterns []*json.RawMessage `json:"ps"`
}

type GetPatternsRequest struct {
	ProjectId     uint64     `json:"pid"`
	ModelId       uint64     `json:"mid"`
	PatternEvents [][]string `json:"pe"`
}

type GetPatternsResponse struct {
	GenericRPCResp
	Patterns []*json.RawMessage `json:"ps"`
}

type GetTotalEventCountRequest struct {
	ProjectId uint64 `json:"pid"`
	ModelId   uint64 `json:"mid"` // Optional, if not passed latest modelId will be used
}

type GetTotalEventCountResponse struct {
	GenericRPCResp
	TotalEventCount uint64 `json:"tec"`
}

type ModelInfo struct {
	ModelId        uint64 `json:"mid"`
	ModelType      string `json:"mt"`
	StartTimestamp int64  `json:"st"`
	EndTimestamp   int64  `json:"et"`
}

type GetUserAndEventsInfoRequest struct {
	ProjectId uint64 `json:"pid"`
	ModelId   uint64 `json:"mid"` // Optional, if not passed latest modelId will be used
}

type GetUserAndEventsInfoResponse struct {
	GenericRPCResp
	UserAndEventsInfo pattern.UserAndEventsInfo `json:"uei"`
}

func CreatePatternsFromRawPatterns(rawPatterns []*json.RawMessage) ([]*pattern.Pattern, error) {
	patterns := make([]*pattern.Pattern, 0, 0)
	for _, rp := range rawPatterns {
		var p pattern.Pattern

		err := json.Unmarshal(*rp, &p)
		if err != nil {
			return patterns, err
		}
		p.GenFrequentProperties()

		patterns = append(patterns, &p)
	}

	return patterns, nil
}

func GetAllPatterns(reqId string, projectId, modelId uint64, startEvent, endEvent string) ([]*pattern.Pattern, error) {
	params := GetAllPatternsRequest{
		ProjectId:  projectId,
		ModelId:    modelId,
		StartEvent: startEvent,
		EndEvent:   endEvent,
	}
	paramBytes, err := rpcJson.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetAllPatterns, params)
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
		err = rpcJson.DecodeClientResponse(r.resp.Body, &result)
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

		resultPatterns, err := CreatePatternsFromRawPatterns(result.Patterns)
		if err != nil {
			log.WithError(err).Error("Error Decoding result patterns on response Ignoring GetAllPatternsResponse")
			continue
		}

		patterns = append(patterns, resultPatterns...)
	}

	return patterns, nil
}

func GetAllContainingPatterns(reqId string, projectId, modelId uint64, event string) ([]*pattern.Pattern, error) {
	params := GetAllContainingPatternsRequest{
		ProjectId: projectId,
		ModelId:   modelId,
		Event:     event,
	}
	paramBytes, err := rpcJson.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetAllContainingPatterns, params)
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
			log.WithError(r.err).Error("Error Ignoring GetAllContainingPatternsResponse")
			continue
		}
		var result GetAllContainingPatternsResponse
		defer r.resp.Body.Close()
		err = rpcJson.DecodeClientResponse(r.resp.Body, &result)
		if err != nil {
			log.WithError(err).Error("Error Decoding response Ignoring GetAllContainingPatternsResponse")
			continue
		}
		if result.Ignored {
			log.Debugln("Ignoring GetAllContainingPatternsResponse")
			continue
		}
		if result.Error != nil {
			log.WithError(result.Error).Error("Error GetAllContainingPatternsResponse")
			continue
		}

		resultPatterns, err := CreatePatternsFromRawPatterns(result.Patterns)
		if err != nil {
			log.WithError(err).Error("Error Decoding result patterns on response Ignoring GetAllContainingPatternsResponse")
			continue
		}

		patterns = append(patterns, resultPatterns...)
	}

	return patterns, nil
}

func GetPatterns(reqId string, projectId, modelId uint64, patternEvents [][]string) ([]*pattern.Pattern, error) {
	params := GetPatternsRequest{
		ProjectId:     projectId,
		ModelId:       modelId,
		PatternEvents: patternEvents,
	}
	paramBytes, err := rpcJson.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetPatterns, params)
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
		err = rpcJson.DecodeClientResponse(r.resp.Body, &result)
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

		resultPatterns, err := CreatePatternsFromRawPatterns(result.Patterns)
		if err != nil {
			log.WithError(err).Error("Error Decoding patterns from response Ignoring GetPatternsResponse")
			continue
		}
		patterns = append(patterns, resultPatterns...)
	}

	return patterns, nil
}

func GetUserAndEventsInfo(reqId string, projectId, modelId uint64) (*pattern.UserAndEventsInfo, uint64, error) {

	params := GetUserAndEventsInfoRequest{
		ProjectId: projectId,
		ModelId:   modelId,
	}
	paramBytes, err := rpcJson.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetUserAndEventsInfo, params)
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
		err = rpcJson.DecodeClientResponse(r.resp.Body, &result)
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

func GetTotalEventCount(reqId string, projectId, modelId uint64) (uint64, error) {
	params := GetTotalEventCountRequest{
		ProjectId: projectId,
		ModelId:   modelId,
	}
	paramBytes, err := rpcJson.EncodeClientRequest(RPCServiceName+Separator+OperationNameGetTotalEventCount, params)
	if err != nil {
		return 0, err
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

	var totalEventCount uint64 = 0
	for r := range gatherResp {
		if r.err != nil {
			log.WithError(r.err).Error("Error Ignoring GetTotalEventCountResponse")
			continue
		}

		var result GetTotalEventCountResponse
		defer r.resp.Body.Close()
		err = rpcJson.DecodeClientResponse(r.resp.Body, &result)
		if err != nil {
			result.Error = err
		}
		if result.Ignored {
			log.Debugln("Ignoring GetTotalEventCountResponse")
			continue
		}

		if result.Error != nil {
			log.WithError(result.Error).Error("Error GetTotalEventCountResponse")
			continue
		}

		totalEventCount += result.TotalEventCount
	}

	return totalEventCount, nil
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
			client := http.Client{
				// timeout in one minute.
				Timeout: time.Duration(600 * time.Second),
			}
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
