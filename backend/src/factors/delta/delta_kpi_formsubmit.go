package delta

import (
	"bufio"
	"encoding/json"
	M "factors/model/model"
	P "factors/pattern"
	U "factors/util"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

var formSubmitMetricToFunc = map[string]func(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error){
	M.Count:        GetFormSubmitCount,
	M.UniqueUsers:  GetFormSubmitUniqueUsers,
	M.CountPerUser: GetFormSubmitCountPerUser,
}

func GetFormSubmitMetrics(metricNames []string, queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (map[string]*MetricInfo, error) {
	metricsInfoMap := make(map[string]*MetricInfo)
	// if queryEvent != U.EVENT_NAME_SESSION {
	// 	return nil, fmt.Errorf("wrong query event for session metric: %s", queryEvent)
	// }
	for _, metric := range metricNames {
		if _, ok := formSubmitMetricToFunc[metric]; !ok {
			continue
		}
		if info, err := formSubmitMetricToFunc[metric](queryEvent, scanner, propFilter, propsToEval); err != nil {
			log.WithError(err).Error("error getFormSubmitMetrics")
		} else {
			metricsInfoMap[metric] = info
		}
	}
	return metricsInfoMap, nil
}

func GetFormSubmitCount(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	var count float64
	var reqMap = make(map[string]map[string]float64)
	var info MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return &MetricInfo{}, err
		}
		eventNameString := eventDetails.EventName

		//check if event is session
		if eventNameString != U.EVENT_NAME_FORM_SUBMITTED {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return &MetricInfo{}, err
		} else if !yes {
			continue
		}

		//global
		count++

		//feat
		for _, propWithType := range propsToEval {
			propTypeName := strings.SplitN(propWithType, "#", 2)
			prop := propTypeName[1]
			propType := propTypeName[0]
			if val, ok := ExistsInProps(prop, eventDetails, propType); ok {
				val := fmt.Sprintf("%s", val)
				if _, ok := reqMap[propWithType]; !ok {
					reqMap[propWithType] = make(map[string]float64)
				}
				reqMap[propWithType][val] += 1
			}
		}
	}
	for prop, valMap := range reqMap {
		for val, cnt := range valMap {
			if cnt == 0 {
				delete(reqMap[prop], val)
			}
		}
		if len(reqMap[prop]) == 0 {
			delete(reqMap, prop)
		}
	}
	info = MetricInfo{Global: count, Features: reqMap}

	return &info, nil
}

func GetFormSubmitUniqueUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	var unique float64
	var reqMap = make(map[string]map[string]float64)
	var info MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, err
		}
		eventNameString := eventDetails.EventName
		userId := eventDetails.UserId

		//check if event is session
		if eventNameString != U.EVENT_NAME_FORM_SUBMITTED {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, err
		} else if !yes {
			continue
		}

		//global
		if _, ok := uniqueUsers[userId]; !ok {
			uniqueUsers[userId] = true
			unique++
		}

		//feat
		for _, propWithType := range propsToEval {
			propTypeName := strings.SplitN(propWithType, "#", 2)
			prop := propTypeName[1]
			propType := propTypeName[0]
			if val, ok := ExistsInProps(prop, eventDetails, propType); ok {
				val := fmt.Sprintf("%s", val)
				propWithVal := strings.Join([]string{prop, val}, ":")
				if _, ok := uniqueUsersFeat[propWithVal]; !ok {
					uniqueUsersFeat[propWithVal] = make(map[string]bool)
				}
				if _, ok := uniqueUsersFeat[propWithVal][userId]; !ok {
					uniqueUsersFeat[propWithVal][userId] = true
					if _, ok := reqMap[propWithType]; !ok {
						reqMap[propWithType] = make(map[string]float64)
					}
					reqMap[propWithType][val] += 1
				}
			}
		}
	}
	for prop, valMap := range reqMap {
		for val, cnt := range valMap {
			if cnt == 0 {
				delete(reqMap[prop], val)
			}
		}
		if len(reqMap[prop]) == 0 {
			delete(reqMap, prop)
		}
	}
	info = MetricInfo{Global: unique, Features: reqMap}
	return &info, nil
}

func GetFormSubmitCountPerUser(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	featInfoMap := make(map[string]map[string]Fraction)
	var unique float64
	var count float64
	var countPerUser float64
	var reqMap = make(map[string]map[string]float64)
	var info MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, err
		}
		eventNameString := eventDetails.EventName
		userId := eventDetails.UserId

		//check if event is session
		if eventNameString != U.EVENT_NAME_FORM_SUBMITTED {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, err
		} else if !yes {
			continue
		}

		//global
		count++
		if _, ok := uniqueUsers[userId]; !ok {
			uniqueUsers[userId] = true
			unique++
		}

		//feat
		for _, propWithType := range propsToEval {
			propTypeName := strings.SplitN(propWithType, "#", 2)
			prop := propTypeName[1]
			propType := propTypeName[0]
			if val, ok := ExistsInProps(prop, eventDetails, propType); ok {
				val := fmt.Sprintf("%s", val)
				if _, ok := featInfoMap[propWithType]; !ok {
					featInfoMap[propWithType] = make(map[string]Fraction)
				}
				if frac, ok := featInfoMap[prop][val]; !ok {
					featInfoMap[propWithType][val] = Fraction{}
				} else {
					frac.Denominator += 1
					featInfoMap[propWithType][val] = frac
				}
				propWithVal := strings.Join([]string{prop, val}, ":")
				if _, ok := uniqueUsersFeat[propWithVal]; !ok {
					uniqueUsersFeat[propWithVal] = make(map[string]bool)
				}
				if _, ok := uniqueUsersFeat[propWithVal][userId]; !ok {
					uniqueUsersFeat[propWithVal][userId] = true
					if frac, ok := featInfoMap[propWithType][val]; !ok {
						featInfoMap[propWithType][val] = Fraction{}
					} else {
						frac.Denominator += 1
						featInfoMap[propWithType][val] = frac
					}
				}
			}
		}
	}

	//get sessionsPerUser

	//global
	if unique != 0 {
		countPerUser = count / unique
	}

	//feat

	for prop, valMap := range featInfoMap {
		reqMap[prop] = make(map[string]float64)
		for val, info := range valMap {
			if !(info.Denominator == 0 || info.Numerator == 0) {
				reqMap[prop][val] = info.Numerator / info.Denominator
			}
		}
		if len(reqMap[prop]) == 0 {
			delete(reqMap, prop)
		}
	}

	info = MetricInfo{Global: countPerUser, Features: reqMap}
	return &info, nil
}
