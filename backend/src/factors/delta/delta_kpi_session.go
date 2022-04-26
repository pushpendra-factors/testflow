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

var sessionMetricToFunc = map[string]func(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error){
	M.TotalSessions:          GetSessionTotalSessions,
	M.UniqueUsers:            GetSessionUniqueUsers,
	M.NewUsers:               GetSessionNewUsers,
	M.RepeatUsers:            GetSessionRepeatUsers,
	M.SessionsPerUser:        GetSessionSessionsPerUser,
	M.EngagedSessions:        GetSessionEngagedSessions,
	M.EngagedUsers:           GetSessionEngagedUsers,
	M.EngagedSessionsPerUser: GetSessionEngagedSessionsPerUser,
	M.TotalTimeOnSite:        GetSessionTotalTimeOnSite,
	M.AvgSessionDuration:     GetSessionAvgSessionDuration,
	M.AvgPageViewsPerSession: GetSessionAvgPageViewsPerSession,
	M.AvgInitialPageLoadTime: GetSessionAvgInitialPageLoadTime,
	M.BounceRate:             GetSessionBounceRate,
	M.EngagementRate:         GetSessionEngagementRate,
}

func GetSessionMetrics(metricNames []string, queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (map[string]*MetricInfo, error) {
	metricsInfoMap := make(map[string]*MetricInfo)
	for _, metric := range metricNames {
		if _, ok := sessionMetricToFunc[metric]; !ok {
			continue
		}
		if info, err := sessionMetricToFunc[metric](queryEvent, scanner, propFilter, propsToEval); err != nil {
			log.WithError(err).Error("error getSessionMetrics")
		} else {
			metricsInfoMap[metric] = info
		}
	}
	return metricsInfoMap, nil
}

func GetSessionTotalSessions(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	var sessionsCount float64
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
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return &MetricInfo{}, err
		} else if !yes {
			continue
		}

		//global
		sessionsCount++

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
	info = MetricInfo{Global: sessionsCount, Features: reqMap}

	return &info, nil
}

func GetSessionUniqueUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
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
		if eventNameString != U.EVENT_NAME_SESSION {
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

func GetSessionNewUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	var new float64
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

		//check if event is session
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, err
		} else if !yes {
			continue
		}

		//check if new user
		if first, ok := eventDetails.EventProperties[U.SP_IS_FIRST_SESSION]; ok {
			first := first.(bool)
			if !first {
				continue
			}
		} else {
			continue
		}

		//global
		new++

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
	info = MetricInfo{Global: new, Features: reqMap}

	return &info, nil
}

func GetSessionRepeatUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	var repeat float64
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
		if eventNameString != U.EVENT_NAME_SESSION {
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
			repeat++
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

		//check if new user
		if first, ok := eventDetails.EventProperties[U.SP_IS_FIRST_SESSION]; ok {
			first := first.(bool)
			if !first {
				continue
			}
		} else {
			continue
		}

		//global
		repeat--

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
				reqMap[propWithType][val] -= 1
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

	info = MetricInfo{Global: repeat, Features: reqMap}
	return &info, nil
}

func GetSessionSessionsPerUser(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	featInfoMap := make(map[string]map[string]Fraction)
	var unique float64
	var sessionsCount float64
	var sessionsPerUser float64
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
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, err
		} else if !yes {
			continue
		}

		//global
		sessionsCount++
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
				if frac, ok := featInfoMap[propWithType][val]; !ok {
					featInfoMap[propWithType][val] = Fraction{Numerator: 1}
				} else {
					frac.Numerator += 1
					featInfoMap[propWithType][val] = frac
				}
				propWithVal := strings.Join([]string{prop, val}, ":")
				if _, ok := uniqueUsersFeat[propWithVal]; !ok {
					uniqueUsersFeat[propWithVal] = make(map[string]bool)
				}
				if _, ok := uniqueUsersFeat[propWithVal][userId]; !ok {
					uniqueUsersFeat[propWithVal][userId] = true
					frac := featInfoMap[propWithType][val]
					frac.Denominator += 1
					featInfoMap[propWithType][val] = frac
				}
			}
		}
	}

	//get sessionsPerUser

	//global
	if unique != 0 {
		sessionsPerUser = sessionsCount / unique
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

	info = MetricInfo{Global: sessionsPerUser, Features: reqMap}
	return &info, nil
}

func GetSessionEngagedSessions(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	var engaged float64
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

		//check if event is session
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, err
		} else if !yes {
			continue
		}

		//check if engaged
		isEngaged := false
		if timeSpent, ok := ExistsInProps(U.SP_SPENT_TIME, eventDetails, "ep"); ok {
			timeSpent := timeSpent.(float64)
			if timeSpent > 10 {
				isEngaged = true
			}
		}
		if cnt, ok := ExistsInProps(U.SP_PAGE_COUNT, eventDetails, "ep"); ok {
			cnt := int64(cnt.(float64))
			if cnt > 2 {
				isEngaged = true
			}
		}
		if !isEngaged {
			continue
		}

		//global
		engaged++

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
	info = MetricInfo{Global: engaged, Features: reqMap}

	return &info, nil
}

func GetSessionEngagedUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
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
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, err
		} else if !yes {
			continue
		}

		//check if engaged
		isEngaged := false
		if timeSpent, ok := ExistsInProps(U.SP_SPENT_TIME, eventDetails, "ep"); ok {
			timeSpent := timeSpent.(float64)
			if timeSpent > 10 {
				isEngaged = true
			}
		}
		if cnt, ok := ExistsInProps(U.SP_PAGE_COUNT, eventDetails, "ep"); ok {
			cnt := int64(cnt.(float64))
			if cnt > 2 {
				isEngaged = true
			}
		}
		if !isEngaged {
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

func GetSessionEngagedSessionsPerUser(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	featInfoMap := make(map[string]map[string]Fraction)
	var unique float64
	var sessionsCount float64
	var sessionsPerUser float64
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
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, err
		} else if !yes {
			continue
		}

		//check if engaged
		isEngaged := false
		if timeSpent, ok := ExistsInProps(U.SP_SPENT_TIME, eventDetails, "ep"); ok {
			timeSpent := timeSpent.(float64)
			if timeSpent > 10 {
				isEngaged = true
			}
		}
		if cnt, ok := ExistsInProps(U.SP_PAGE_COUNT, eventDetails, "ep"); ok {
			cnt := int64(cnt.(float64))
			if cnt > 2 {
				isEngaged = true
			}
		}
		if !isEngaged {
			continue
		}

		//global
		sessionsCount++
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
				if frac, ok := featInfoMap[propWithType][val]; !ok {
					featInfoMap[propWithType][val] = Fraction{Numerator: 1}
				} else {
					frac.Numerator += 1
					featInfoMap[propWithType][val] = frac
				}
				propWithVal := strings.Join([]string{prop, val}, ":")
				if _, ok := uniqueUsersFeat[propWithVal]; !ok {
					uniqueUsersFeat[propWithVal] = make(map[string]bool)
				}
				if _, ok := uniqueUsersFeat[propWithVal][userId]; !ok {
					uniqueUsersFeat[propWithVal][userId] = true
					frac := featInfoMap[propWithType][val]
					frac.Denominator += 1
					featInfoMap[propWithType][val] = frac
				}
			}
		}
	}

	//get sessionsPerUser

	//global
	if unique != 0 {
		sessionsPerUser = sessionsCount / unique
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

	info = MetricInfo{Global: sessionsPerUser, Features: reqMap}
	return &info, nil
}

func GetSessionTotalTimeOnSite(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	var totalSessionTime float64
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
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return &MetricInfo{}, err
		} else if !yes {
			continue
		}

		if timeSpent, ok := ExistsInProps(U.SP_SPENT_TIME, eventDetails, "ep"); ok {
			timeOnSite := timeSpent.(float64)

			//global
			totalSessionTime += timeOnSite

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
					reqMap[propWithType][val] += timeOnSite
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

	info = MetricInfo{Global: totalSessionTime, Features: reqMap}

	return &info, nil
}

func GetSessionAvgSessionDuration(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	var avgSessionDuration float64
	var sessionsCount float64
	var totalSessionTime float64
	var featInfoMap = make(map[string]map[string]Fraction)
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
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return &MetricInfo{}, err
		} else if !yes {
			continue
		}

		if timeSpent, ok := ExistsInProps(U.SP_SPENT_TIME, eventDetails, "ep"); ok {
			timeSpent := timeSpent.(float64)

			timeOnSite := float64(timeSpent)
			//global
			sessionsCount++
			totalSessionTime += timeOnSite

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
					if frac, ok := featInfoMap[propWithType][val]; !ok {
						featInfoMap[propWithType][val] = Fraction{Denominator: 1, Numerator: timeOnSite}
					} else {
						frac.Numerator += timeOnSite
						frac.Denominator += 1
						featInfoMap[propWithType][val] = frac
					}
				}
			}
		}
	}

	//get sessionsPerUser

	//global
	if sessionsCount != 0 {
		avgSessionDuration = totalSessionTime / sessionsCount
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
	info = MetricInfo{Global: avgSessionDuration, Features: reqMap}

	return &info, nil
}

func GetSessionAvgPageViewsPerSession(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	var totalSessionPageCounts float64
	var sessionsCount float64
	var avgPageViewsPerSession float64
	var featInfoMap = make(map[string]map[string]Fraction)
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
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return &MetricInfo{}, err
		} else if !yes {
			continue
		}

		//global
		sessionsCount++

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
				if frac, ok := featInfoMap[propWithType][val]; !ok {
					featInfoMap[propWithType][val] = Fraction{Denominator: 1}
				} else {
					frac.Denominator += 1
					featInfoMap[propWithType][val] = frac
				}
			}
		}

		//check if event has pageview count as property
		if cnt, ok := ExistsInProps(U.SP_PAGE_COUNT, eventDetails, "ep"); !ok {
			continue
		} else {
			cnt := cnt.(float64)
			//global
			totalSessionPageCounts += cnt
			//feat
			for _, propWithType := range propsToEval {
				propTypeName := strings.SplitN(propWithType, "#", 2)
				prop := propTypeName[1]
				propType := propTypeName[0]
				if val, ok := ExistsInProps(prop, eventDetails, propType); ok {
					val := fmt.Sprintf("%s", val)
					frac := featInfoMap[propWithType][val]
					frac.Numerator += cnt
					featInfoMap[propWithType][val] = frac
				}
			}
		}

	}

	// get average PageViews Per Session

	//global
	if sessionsCount != 0 {
		avgPageViewsPerSession = totalSessionPageCounts / sessionsCount
	}

	//feat
	for key, valMap := range featInfoMap {
		reqMap[key] = make(map[string]float64)
		for val, info := range valMap {
			if !(info.Denominator == 0 || info.Numerator == 0) {
				reqMap[key][val] = info.Numerator / info.Denominator
			}
		}
		if len(reqMap[key]) == 0 {
			delete(reqMap, key)
		}
	}

	info = MetricInfo{Global: avgPageViewsPerSession, Features: reqMap}

	return &info, nil
}

func GetSessionAvgInitialPageLoadTime(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	var avgInitialPageLoadTime float64
	var sessionsCount float64
	var initialPageLoadTime float64
	var featInfoMap = make(map[string]map[string]Fraction)
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
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return &MetricInfo{}, err
		} else if !yes {
			continue
		}

		if time, ok := ExistsInProps(U.SP_INITIAL_PAGE_LOAD_TIME, eventDetails, "ep"); ok {
			time := time.(float64)

			//global
			sessionsCount++
			initialPageLoadTime += time

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
					if frac, ok := featInfoMap[propWithType][val]; !ok {
						featInfoMap[propWithType][val] = Fraction{Denominator: 1, Numerator: time}
					} else {
						frac.Numerator += time
						frac.Denominator += 1
						featInfoMap[propWithType][val] = frac
					}
				}
			}
		}
	}

	//get sessionsPerUser

	//global
	if sessionsCount != 0 {
		avgInitialPageLoadTime = initialPageLoadTime / sessionsCount
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
	info = MetricInfo{Global: avgInitialPageLoadTime, Features: reqMap}

	return &info, nil
}

func GetSessionBounceRate(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	var bounceSessions float64
	var sessionsCount float64
	var bounceRate float64
	var featInfoMap = make(map[string]map[string]Fraction) //[]string = (bounceSessions,sessionsCount)
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
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return &MetricInfo{}, err
		} else if !yes {
			continue
		}

		//global
		sessionsCount++

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
				if frac, ok := featInfoMap[propWithType][val]; !ok {
					featInfoMap[propWithType][val] = Fraction{Denominator: 1}
				} else {
					frac.Denominator += 1
					featInfoMap[propWithType][val] = frac
				}
			}
		}

		//check if it is a bounced session
		if cnt, ok := ExistsInProps(U.SP_PAGE_COUNT, eventDetails, "ep"); ok {
			cnt := int64(cnt.(float64))
			if cnt != 1 {
				continue
			}
		} else {
			continue
		}

		//global
		bounceSessions++

		//feat
		for _, propWithType := range propsToEval {
			propTypeName := strings.SplitN(propWithType, "#", 2)
			prop := propTypeName[1]
			propType := propTypeName[0]
			if val, ok := ExistsInProps(prop, eventDetails, propType); ok {
				val := fmt.Sprintf("%s", val)
				frac := featInfoMap[propWithType][val]
				frac.Numerator += 1
				featInfoMap[propWithType][val] = frac
			}
		}
	}

	// get bounce rate

	//global
	if sessionsCount != 0 {
		bounceRate = bounceSessions / sessionsCount
	}

	//feat
	for key, valMap := range featInfoMap {
		reqMap[key] = make(map[string]float64)
		for val, info := range valMap {
			if !(info.Denominator == 0 || info.Numerator == 0) {
				reqMap[key][val] = info.Numerator / info.Denominator
			}
		}
		if len(reqMap[key]) == 0 {
			delete(reqMap, key)
		}
	}

	info = MetricInfo{Global: bounceRate, Features: reqMap}

	return &info, nil
}

func GetSessionEngagementRate(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, error) {
	var engagedSessions float64
	var sessionsCount float64
	var engagementRate float64
	var featInfoMap = make(map[string]map[string]Fraction) //[]string = (bounceSessions,sessionsCount)
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
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return &MetricInfo{}, err
		} else if !yes {
			continue
		}

		//global
		sessionsCount++

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
				if frac, ok := featInfoMap[propWithType][val]; !ok {
					featInfoMap[propWithType][val] = Fraction{Denominator: 1}
				} else {
					frac.Denominator += 1
					featInfoMap[propWithType][val] = frac
				}
			}
		}

		//check if engaged
		isEngaged := false
		if timeSpent, ok := ExistsInProps(U.SP_SPENT_TIME, eventDetails, "ep"); ok {
			timeSpent := timeSpent.(float64)
			if timeSpent > 10 {
				isEngaged = true
			}
		}
		if cnt, ok := ExistsInProps(U.SP_PAGE_COUNT, eventDetails, "ep"); ok {
			cnt := int64(cnt.(float64))
			if cnt > 2 {
				isEngaged = true
			}
		}
		if !isEngaged {
			continue
		}

		//global
		engagedSessions++

		//feat
		for _, propWithType := range propsToEval {
			propTypeName := strings.SplitN(propWithType, "#", 2)
			prop := propTypeName[1]
			propType := propTypeName[0]
			if val, ok := ExistsInProps(prop, eventDetails, propType); ok {
				val := fmt.Sprintf("%s", val)
				frac := featInfoMap[propWithType][val]
				frac.Numerator += 1
				featInfoMap[propWithType][val] = frac
			}
		}
	}

	// get engagement rate

	//global
	if sessionsCount != 0 {
		engagementRate = engagedSessions / sessionsCount
	}

	//feat
	for key, valMap := range featInfoMap {
		reqMap[key] = make(map[string]float64)
		for val, info := range valMap {
			if !(info.Denominator == 0 || info.Numerator == 0) {
				reqMap[key][val] = info.Numerator / info.Denominator
			}
		}
		if len(reqMap[key]) == 0 {
			delete(reqMap, key)
		}
	}

	info = MetricInfo{Global: engagementRate, Features: reqMap}

	return &info, nil
}
