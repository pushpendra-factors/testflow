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

var pageViewMetricToFunc = map[string]func(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error){
	M.Entrances:                GetPageviewEntrances,
	M.Exits:                    GetPageviewExits,
	M.PageViews:                GetPageviewPageViews,
	M.UniqueUsers:              GetPageviewUniqueUsers,
	M.PageviewsPerUser:         GetPageviewPageviewsPerUser,
	M.AvgPageLoadTime:          GetPageviewAvgPageLoadTime,
	M.AvgVerticalScrollPercent: GetPageviewAvgVerticalScrollPercent,
	M.AvgTimeOnPage:            GetPageviewAvgTimeOnPage,
	M.EngagedPageViews:         GetPageviewEngagedPageViews,
	M.EngagedUsers:             GetPageviewEngagedUsers,
	M.EngagementRate:           GetPageviewEngagementRate,
}

func GetPageViewMetrics(metricNames []string, queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*WithinPeriodInsightsKpi, error) {
	var wpi WithinPeriodInsightsKpi
	for i, metric := range metricNames {
		if i == 1 {
			break
		}
		if _, ok := pageViewMetricToFunc[metric]; !ok {
			continue
		}
		if info, scale, err := pageViewMetricToFunc[metric](queryEvent, scanner, propFilter, propsToEval); err != nil {
			log.WithError(err).Error("error getPageViewMetrics")
		} else {
			wpi.MetricInfo = info
			wpi.ScaleInfo = scale
		}
	}
	return &wpi, nil
}

func GetPageviewEntrances(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var entrances float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName

		//check if event is session
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, nil, err
		} else if !yes {
			continue
		}

		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		isEntrance := false
		if url, ok := ExistsInProps(U.SP_INITIAL_PAGE_URL, eventDetails, "ep"); ok {
			if url := url.(string); url == queryEvent {
				isEntrance = true
			}
		}
		if !isEntrance {
			continue
		}

		//global
		entrances++

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
	info = MetricInfo{Global: entrances, Features: reqMap}

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	return &info, &scale, nil
}

func GetPageviewExits(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var exits float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName

		//check if event is session
		if eventNameString != U.EVENT_NAME_SESSION {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, nil, err
		} else if !yes {
			continue
		}

		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		isExit := false
		if url, ok := ExistsInProps(U.SP_LATEST_PAGE_URL, eventDetails, "ep"); ok {
			if url := url.(string); url == queryEvent {
				isExit = true
			}
		}
		if !isExit {
			continue
		}

		//global
		exits++

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
	info = MetricInfo{Global: exits, Features: reqMap}

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	return &info, &scale, nil
}

func GetPageviewPageViews(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var pageviews float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName

		//check if event is session
		if eventNameString != queryEvent {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, nil, err
		} else if !yes {
			continue
		}

		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		//global
		pageviews++

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
	info = MetricInfo{Global: pageviews, Features: reqMap}

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	return &info, &scale, nil
}

func GetPageviewUniqueUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	var unique float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName
		userId := eventDetails.UserId

		//check if event is session
		if eventNameString != queryEvent {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, nil, err
		} else if !yes {
			continue
		}

		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

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
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	return &info, &scale, nil
}

func GetPageviewPageviewsPerUser(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	featInfoMap := make(map[string]map[string]Fraction)
	var unique float64
	var pageviews float64
	var PageviewsPerUser float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName
		userId := eventDetails.UserId

		//check if event is session
		if eventNameString != queryEvent {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, nil, err
		} else if !yes {
			continue
		}

		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		//global
		pageviews++
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
		PageviewsPerUser = pageviews / unique
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

	info = MetricInfo{Global: PageviewsPerUser, Features: reqMap}
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	return &info, &scale, nil
}

func GetPageviewAvgPageLoadTime(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var avgPageLoadTime float64
	var pageviews float64
	var totalPageLoadTime float64
	var featInfoMap = make(map[string]map[string]Fraction)
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName

		//check if event is session
		if eventNameString != queryEvent {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, nil, err
		} else if !yes {
			continue
		}

		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		if loadTime, ok := ExistsInProps(U.EP_PAGE_LOAD_TIME, eventDetails, "ep"); ok {
			loadTime := loadTime.(float64)

			//global
			pageviews++
			totalPageLoadTime += loadTime

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
						featInfoMap[propWithType][val] = Fraction{Denominator: 1, Numerator: loadTime}
					} else {
						frac.Numerator += loadTime
						frac.Denominator += 1
						featInfoMap[propWithType][val] = frac
					}
				}
			}
		}
	}

	//get sessionsPerUser

	//global
	if pageviews != 0 {
		avgPageLoadTime = totalPageLoadTime / pageviews
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
	info = MetricInfo{Global: avgPageLoadTime, Features: reqMap}

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	return &info, &scale, nil
}

func GetPageviewAvgVerticalScrollPercent(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var totalVerticalScrollPercent float64
	var pageviews float64
	var avgVerticalScrollPercent float64
	var featInfoMap = make(map[string]map[string]Fraction)
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName

		//check if event is session
		if eventNameString != queryEvent {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, nil, err
		} else if !yes {
			continue
		}

		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		//global
		pageviews++

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

		if scrollPerc, ok := ExistsInProps(U.EP_PAGE_SCROLL_PERCENT, eventDetails, "ep"); !ok {
			continue
		} else {
			scrollPerc := scrollPerc.(float64)
			//global
			totalVerticalScrollPercent += scrollPerc
			//feat
			for _, propWithType := range propsToEval {
				propTypeName := strings.SplitN(propWithType, "#", 2)
				prop := propTypeName[1]
				propType := propTypeName[0]
				if val, ok := ExistsInProps(prop, eventDetails, propType); ok {
					val := fmt.Sprintf("%s", val)
					frac := featInfoMap[propWithType][val]
					frac.Numerator += scrollPerc
					featInfoMap[propWithType][val] = frac
				}
			}
		}

	}

	// get average vertical scroll percent

	//global
	if pageviews != 0 {
		avgVerticalScrollPercent = totalVerticalScrollPercent / pageviews
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

	info = MetricInfo{Global: avgVerticalScrollPercent, Features: reqMap}

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	return &info, &scale, nil
}

func GetPageviewAvgTimeOnPage(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var avgTimeOnPage float64
	var pageviews float64
	var totalTimeOnPage float64
	var featInfoMap = make(map[string]map[string]Fraction)
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName

		//check if event is session
		if eventNameString != queryEvent {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, nil, err
		} else if !yes {
			continue
		}

		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		if time, ok := ExistsInProps(U.EP_PAGE_SPENT_TIME, eventDetails, "ep"); ok {
			time := time.(float64)

			//global
			pageviews++
			totalTimeOnPage += time

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
	if pageviews != 0 {
		avgTimeOnPage = totalTimeOnPage / pageviews
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
	info = MetricInfo{Global: avgTimeOnPage, Features: reqMap}

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	return &info, &scale, nil
}

func GetPageviewEngagedPageViews(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var engaged float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName

		//check if event is session
		if eventNameString != queryEvent {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, nil, err
		} else if !yes {
			continue
		}

		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		//check if engaged
		isEngaged := false
		if spentTime, ok := ExistsInProps(U.EP_PAGE_SPENT_TIME, eventDetails, "ep"); ok {
			spentTime := spentTime.(float64)
			if spentTime > 10 {
				isEngaged = true
			}
		}
		if scrollPerc, ok := ExistsInProps(U.EP_PAGE_SCROLL_PERCENT, eventDetails, "ep"); ok {
			scrollPerc := scrollPerc.(float64)
			if scrollPerc > 50 {
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

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	return &info, &scale, nil
}

func GetPageviewEngagedUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	var unique float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName
		userId := eventDetails.UserId

		//check if event is session
		if eventNameString != queryEvent {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, nil, err
		} else if !yes {
			continue
		}

		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		//check if engaged
		isEngaged := false
		if spentTime, ok := ExistsInProps(U.EP_PAGE_SPENT_TIME, eventDetails, "ep"); ok {
			spentTime := spentTime.(float64)
			if spentTime > 10 {
				isEngaged = true
			}
		}
		if scrollPerc, ok := ExistsInProps(U.EP_PAGE_SCROLL_PERCENT, eventDetails, "ep"); ok {
			scrollPerc := scrollPerc.(float64)
			if scrollPerc > 50 {
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
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	return &info, &scale, nil
}

func GetPageviewEngagementRate(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var engagedPageviews float64
	var pageviews float64
	var engagementRate float64
	var featInfoMap = make(map[string]map[string]Fraction)
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName

		//check if event is session
		if eventNameString != queryEvent {
			continue
		}

		//check if event contains all requiredProps(constraint)
		if yes, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
			return nil, nil, err
		} else if !yes {
			continue
		}

		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		//global
		pageviews++

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
					featInfoMap[propWithType][val] = Fraction{}
				} else {
					frac.Denominator += 1
					featInfoMap[propWithType][val] = frac
				}
			}
		}

		//check if engaged
		isEngaged := false
		if spentTime, ok := ExistsInProps(U.EP_PAGE_SPENT_TIME, eventDetails, "ep"); ok {
			spentTime := spentTime.(float64)
			if spentTime > 10 {
				isEngaged = true
			}
		}
		if scrollPerc, ok := ExistsInProps(U.EP_PAGE_SCROLL_PERCENT, eventDetails, "ep"); ok {
			scrollPerc := scrollPerc.(float64)
			if scrollPerc > 50 {
				isEngaged = true
			}
		}
		if !isEngaged {
			continue
		}

		//global
		engagedPageviews++

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

	// get engagemene rate

	//global
	engagementRate = SmartDivide(engagedPageviews, pageviews)

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

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	return &info, &scale, nil
}
