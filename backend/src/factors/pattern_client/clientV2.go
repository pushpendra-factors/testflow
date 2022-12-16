package patternclient

import (
	"bufio"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	"factors/pattern"
	P "factors/pattern"
	"reflect"
	"strings"

	PS "factors/pattern_server/store"
)

func GetAllPatternsV2(reqId string, projectId int64, modelId uint64, startEvent, endEvent string) ([]*P.Pattern, error) {

	var patterns []*P.Pattern
	patterns = make([]*P.Pattern, 0)
	patternsToReturn := make([]*json.RawMessage, 0, 0)

	fm := C.GetCloudManager(projectId)

	patternsWithMeta, err := getExplainV2FromFileManager(fm, projectId, modelId, "chunk_1.txt")
	if err != nil {
		return nil, err
	}

	filterPatterns := GetFilter(startEvent, endEvent, "")
	for _, pwm := range patternsWithMeta {
		if filterPatterns(pwm.PatternEvents, startEvent, endEvent, "") {
			patternsToReturn = append(patternsToReturn, &pwm.RawPattern)
		}
	}
	patterns, err = CreatePatternsFromRawPatterns(patternsToReturn)
	if err != nil {
		return nil, err
	}
	return patterns, nil
}

func GetAllContainingPatternsV2(reqId string, projectId int64, modelId uint64, event string) ([]*P.Pattern, error) {

	var patterns []*P.Pattern
	patterns = make([]*P.Pattern, 0)
	patternsToReturn := make([]*json.RawMessage, 0, 0)

	fm := C.GetCloudManager(projectId)

	patternsWithMeta, err := getExplainV2FromFileManager(fm, projectId, modelId, "chunk_1.txt")
	if err != nil {
		return nil, err
	}

	filterPatterns := GetFilter("", "", event)
	for _, pwm := range patternsWithMeta {
		if filterPatterns(pwm.PatternEvents, "", "", event) {
			patternsToReturn = append(patternsToReturn, &pwm.RawPattern)
		}
	}
	patterns, err = CreatePatternsFromRawPatterns(patternsToReturn)
	if err != nil {
		return nil, err
	}
	return patterns, nil
}

func GetPatternsV2(reqId string, projectId int64, modelId uint64, patternEvents [][]string) ([]*P.Pattern, error) {

	var patterns []*P.Pattern
	patterns = make([]*P.Pattern, 0)
	patternsToReturn := make([]*json.RawMessage, 0, 0)

	fm := C.GetCloudManager(projectId)

	patternsWithMeta, err := getExplainV2FromFileManager(fm, projectId, modelId, "chunk_1.txt")
	if err != nil {
		return nil, err
	}

	for _, pwm := range patternsWithMeta {
		for _, pE := range patternEvents {
			if reflect.DeepEqual(pE, pwm.PatternEvents) {
				patternsToReturn = append(patternsToReturn, &pwm.RawPattern)
			}
		}
	}
	patterns, err = CreatePatternsFromRawPatterns(patternsToReturn)
	if err != nil {
		return nil, err
	}
	return patterns, nil
}

func GetUserAndEventsInfoV2(reqId string, projectId int64, modelId uint64) (*P.UserAndEventsInfo, error) {
	fm := C.GetCloudManager(projectId)
	userEventsInfo, err := getModelEventInfoFromFileManager(fm, projectId, modelId)
	if err != nil {
		return nil, err
	}
	return &userEventsInfo, nil
}

func GetTotalEventCountV2(reqId string, projectId int64, modelId uint64) (uint64, error) {

	var totalEventCount uint64 = 0
	singleEventRawPatterns := make([]*json.RawMessage, 0, 0)

	fm := C.GetCloudManager(projectId)

	patternsWithMeta, err := getExplainV2FromFileManager(fm, projectId, modelId, "chunk_1.txt")
	if err != nil {
		return 0, err
	}

	for _, pwm := range patternsWithMeta {

		if len(pwm.PatternEvents) == 1 {
			singleEventRawPatterns = append(singleEventRawPatterns, &pwm.RawPattern)
		}

	}

	if singleEventPatterns, err := CreatePatternsFromRawPatterns(singleEventRawPatterns); err != nil {
		return 0, err
	} else {
		for _, p := range singleEventPatterns {
			totalEventCount += uint64(p.PerOccurrenceCount)
		}
	}
	return totalEventCount, nil

}

func getModelEventInfoFromFileManager(fm filestore.FileManager, projectId int64, modelId uint64) (P.UserAndEventsInfo, error) {
	path, fName := fm.GetModelEventInfoFilePathAndName(projectId, modelId)
	modelEventInfoFile, err := fm.Get(path, fName)
	if err != nil {
		return pattern.UserAndEventsInfo{}, err
	}
	defer modelEventInfoFile.Close()

	scanner := bufio.NewScanner(modelEventInfoFile)

	patternEventInfo, err := PS.CreatePatternEventInfoFromScanner(scanner)

	return patternEventInfo, err
}

func getExplainV2FromFileManager(fm filestore.FileManager, projectId int64, modelId uint64, chunkId string) ([]*PS.PatternWithMeta, error) {
	path, fName := fm.GetExplainV2ModelPath(modelId, projectId)
	patternsReader, err := fm.Get(path, fName)
	if err != nil {
		return nil, err
	}
	defer patternsReader.Close()
	scanner := PS.CreateScannerFromReader(patternsReader)
	patterns, err := PS.CreatePatternsWithMetaFromScanner(scanner)
	return patterns, err
}

type FilterPattern func(patternEvents []string, startEvent, endEvent, anyEvent string) bool

func filterByStart(patternEvents []string, startEvent, endEvent, anyEvent string) bool {
	return strings.Compare(startEvent, patternEvents[0]) == 0
}

func filterByEnd(patternEvents []string, startEvent, endEvent, anyEvent string) bool {
	pLen := len(patternEvents)
	return strings.Compare(endEvent, patternEvents[pLen-1]) == 0
}

func filterByStartAndEnd(patternEvents []string, startEvent, endEvent, anyEvent string) bool {
	return filterByStart(patternEvents, startEvent, "", "") && filterByEnd(patternEvents, "", endEvent, "")
}

func filterByContaining(patternEvents []string, startEvent, endEvent, anyEvent string) bool {
	pLen := len(patternEvents)
	for i := 0; i < pLen; i++ {
		if strings.Compare(anyEvent, patternEvents[i]) == 0 {
			return true
		}
	}
	return false
}

func GetFilter(startEvent, endEvent, anyEvent string) FilterPattern {
	if startEvent != "" && endEvent != "" {
		return filterByStartAndEnd
	}

	if anyEvent != "" {
		return filterByContaining
	}
	if endEvent != "" {
		return filterByEnd
	}

	if startEvent != "" {
		return filterByStart
	}
	return nil
}
