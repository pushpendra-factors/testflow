package patternserver

import (
	"encoding/json"
	"errors"
	client "factors/pattern_client"
	store "factors/pattern_server/store"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	E "github.com/pkg/errors"
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
		err := E.Wrap(errors.New("MissingParams"), "GetAllPatterns missing param projectID")
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

	startTime := time.Now()
	chunkIds, found := ps.GetProjectModelChunks(args.ProjectId, modelId)
	if !found {
		err := E.Wrap(errors.New("ProjectModelChunks Not Found"), fmt.Sprintf("GetAllPatterns failed to fetch ProjectModelChunks, ProjectID: %d, ModelID: %d", args.ProjectId, modelId))
		result.Error = err
		return err
	}
	endTime := time.Now()
	log.WithFields(log.Fields{
		"time_taken": endTime.Sub(startTime).Milliseconds()}).Info("debug_time_GetProjectModelChunks")

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
		startTime := time.Now()
		patternsWithMeta, err := ps.store.GetPatternsWithMeta(args.ProjectId, modelId, chunkId)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"pid": args.ProjectId,
				"mid": modelId,
				"cid": chunkId,
			}).Error("Failed To get chunk patterns")
			continue
		}
		endTime := time.Now()
		log.WithFields(log.Fields{
			"time_taken": endTime.Sub(startTime).Milliseconds()}).Info("debug_time_GetPatternsWithMeta")
		if filterPatterns == nil {
			startTime := time.Now()
			rawPatterns := store.GetAllRawPatterns(patternsWithMeta)
			endTime := time.Now()
			log.WithFields(log.Fields{
				"time_taken": endTime.Sub(startTime).Milliseconds()}).Info("debug_time_GetAllRawPatterns")
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
		err := E.Wrap(errors.New("MissingParams"), fmt.Sprintf("%s", "GetPatterns missing param projectID"))
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
		err := E.Wrap(errors.New("ProjectModelChunks Not Found"), fmt.Sprintf("GetAllPatterns failed to fetch ProjectModelChunks, ProjectID: %d, ModelID: %d", args.ProjectId, modelId))
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

func (ps *PatternServer) GetProjectModelsIntervals(
	r *http.Request, args *client.GetProjectModelIntervalsRequest,
	result *client.GetProjectModelIntervalsResponse) error {
	if args == nil || args.ProjectId == 0 {
		err := E.Wrap(errors.New("MissingParams"), "GetProjectModelsIntervals Missing Param")
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

func (ps *PatternServer) GetTotalEventCount(
	r *http.Request, args *client.GetTotalEventCountRequest,
	result *client.GetTotalEventCountResponse) error {
	if args == nil || args.ProjectId == 0 {
		err := E.Wrap(errors.New("MissingParams"), "GetTotalEventCount missing param projectID")
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
		"pid": args.ProjectId,
		"mid": modelId,
	}).Debugln("RPC Fetching Total Event COunt")

	chunkIds, found := ps.GetProjectModelChunks(args.ProjectId, modelId)
	if !found {
		err := E.Wrap(
			errors.New("ProjectModelChunks Not Found"),
			fmt.Sprintf(
				"GetTotalEventCount failed to fetch ProjectModelChunks, ProjectID: %d, ModelID: %d",
				args.ProjectId, modelId))
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

	singleEventRawPatterns := make([]*json.RawMessage, 0, 0)
	var totalEventCount uint64 = 0

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
			if len(pwm.PatternEvents) == 1 {
				singleEventRawPatterns = append(singleEventRawPatterns, &pwm.RawPattern)
			}
		}
	}

	if singleEventPatterns, err := client.CreatePatternsFromRawPatterns(singleEventRawPatterns); err != nil {
		result.Error = err
		return err
	} else {
		for _, p := range singleEventPatterns {
			totalEventCount += uint64(p.PerOccurrenceCount)
		}
	}

	result.ProjectId = args.ProjectId
	result.ModelId = modelId
	result.TotalEventCount = totalEventCount

	return nil
}
