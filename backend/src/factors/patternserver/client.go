package patternserver

import (
	"bytes"
	"fmt"

	"net/http"
	"sync"

	C "factors/config"

	"github.com/gorilla/rpc/json"
	log "github.com/sirupsen/logrus"
)

func GetProjectModelIntervals(projectId uint64) ([]ModelInfo, error) {
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
		err = json.DecodeClientResponse(r.resp.Body, &result)
		if err != nil {
			log.WithError(r.err).Error("Error Ignoring GetProjectModelIntervalsResponse")
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

func GetSeenEventPropertyValues(projectId, modelId uint64, eventName, propertyName string) ([]string, error) {
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

func GetSeenEventProperties(projectId, modelId uint64, eventName string) (map[string][]string, error) {

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
			resp[k] = v // should merge keys rather than over riding keys
		}
	}

	return resp, nil
}

type httpResp struct {
	resp *http.Response
	err  error
}

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
