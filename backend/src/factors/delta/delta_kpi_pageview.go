package delta

import (
	"bufio"
	"encoding/json"
	M "factors/model/model"
	P "factors/pattern"
	U "factors/util"

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

		//check if event is session and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, U.EVENT_NAME_SESSION, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}

		if !checkEntrance(eventDetails, queryEvent) {
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		addValueToMapForPropsPresent(&entrances, reqMap, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
	}
	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: entrances, Features: reqMap}

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

		//check if event is session and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, U.EVENT_NAME_SESSION, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		if !checkExit(eventDetails, queryEvent) {
			continue
		}

		addValueToMapForPropsPresent(&exits, reqMap, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
	}
	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: exits, Features: reqMap}

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

		//check if page is correct and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, queryEvent, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		addValueToMapForPropsPresent(&pageviews, reqMap, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
	}
	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: pageviews, Features: reqMap}

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

		//check if page is correct and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, queryEvent, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)
		addValueToMapForPropsPresentUser(&unique, reqMap, 1, propsToEval, eventDetails, uniqueUsers, uniqueUsersFeat)
	}
	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: unique, Features: reqMap}
	return &info, &scale, nil
}

func GetPageviewPageviewsPerUser(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	featInfoMap := make(map[string]map[string]Fraction)
	var globalScale float64
	var pageviewsPerUser float64
	var pageviewsPerUserFrac Fraction
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

		//check if page is correct and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, queryEvent, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		addValuesToFractionForPropsPresentUser(&pageviewsPerUserFrac, featInfoMap, 1, 1, propsToEval, eventDetails, uniqueUsers, uniqueUsersFeat)
	}

	//feat
	pageviewsPerUser, reqMap = getFractionValue(&pageviewsPerUserFrac, featInfoMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: pageviewsPerUser, Features: reqMap}
	return &info, &scale, nil
}

func GetPageviewAvgPageLoadTime(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var avgPageLoadTime float64
	var avgPageLoadTimeFrac Fraction
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

		//check if page is correct and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, queryEvent, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		if loadTime, ok := ExistsInProps(U.EP_PAGE_LOAD_TIME, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
			loadTime := loadTime.(float64)
			addValuesToFractionForPropsPresent(&avgPageLoadTimeFrac, featInfoMap, loadTime, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
		}
	}

	avgPageLoadTime, reqMap = getFractionValue(&avgPageLoadTimeFrac, featInfoMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: avgPageLoadTime, Features: reqMap}

	return &info, &scale, nil
}

func GetPageviewAvgVerticalScrollPercent(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var avgVerticalScrollPercentFrac Fraction
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

		//check if page is correct and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, queryEvent, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		addValuesToFractionForPropsPresent(&avgVerticalScrollPercentFrac, featInfoMap, 0, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)

		if scrollPerc, ok := ExistsInProps(U.EP_PAGE_SCROLL_PERCENT, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
			scrollPerc := scrollPerc.(float64)
			addValuesToFractionForPropsPresent(&avgVerticalScrollPercentFrac, featInfoMap, scrollPerc, 0, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
		}

	}

	// get average vertical scroll percent
	avgVerticalScrollPercent, reqMap = getFractionValue(&avgVerticalScrollPercentFrac, featInfoMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: avgVerticalScrollPercent, Features: reqMap}

	return &info, &scale, nil
}

func GetPageviewAvgTimeOnPage(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var avgTimeOnPage float64
	var avgTimeOnPageFrac Fraction
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

		//check if page is correct and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, queryEvent, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		if time, ok := ExistsInProps(U.EP_PAGE_SPENT_TIME, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
			time := time.(float64)

			addValuesToFractionForPropsPresent(&avgTimeOnPageFrac, featInfoMap, time, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
		}
	}

	//feat
	avgTimeOnPage, reqMap = getFractionValue(&avgTimeOnPageFrac, featInfoMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: avgTimeOnPage, Features: reqMap}

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

		//check if page is correct and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, queryEvent, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		//check if engaged
		isEngaged := checkEngagedPageview(eventDetails)
		if !isEngaged {
			continue
		}

		addValueToMapForPropsPresent(&engaged, reqMap, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
	}
	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: engaged, Features: reqMap}

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

		//check if page is correct and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, queryEvent, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)
		//check if engaged
		isEngaged := checkEngagedPageview(eventDetails)
		if !isEngaged {
			continue
		}
		addValueToMapForPropsPresentUser(&unique, reqMap, 1, propsToEval, eventDetails, uniqueUsers, uniqueUsersFeat)
	}
	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: unique, Features: reqMap}
	return &info, &scale, nil
}

func GetPageviewEngagementRate(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var engagementRateFrac Fraction
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

		//check if page is correct and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, queryEvent, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		addValuesToFractionForPropsPresent(&engagementRateFrac, featInfoMap, 0, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)

		//check if engaged
		isEngaged := checkEngagedPageview(eventDetails)
		if !isEngaged {
			continue
		}

		addValuesToFractionForPropsPresent(&engagementRateFrac, featInfoMap, 1, 0, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
	}

	engagementRate, reqMap = getFractionValue(&engagementRateFrac, featInfoMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: engagementRate, Features: reqMap}

	return &info, &scale, nil
}

func checkEngagedPageview(eventDetails P.CounterEventFormat) bool {
	isEngaged := false
	if spentTime, ok := ExistsInProps(U.EP_PAGE_SPENT_TIME, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		spentTime := spentTime.(float64)
		if spentTime > 10 {
			isEngaged = true
		}
	}
	if scrollPerc, ok := ExistsInProps(U.EP_PAGE_SCROLL_PERCENT, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		scrollPerc := scrollPerc.(float64)
		if scrollPerc > 50 {
			isEngaged = true
		}
	}
	return isEngaged
}

func checkEntrance(eventDetails P.CounterEventFormat, pageName string) bool {
	isEntrance := false
	if url, ok := ExistsInProps(U.SP_INITIAL_PAGE_URL, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		if url := url.(string); url == pageName {
			isEntrance = true
		}
	}
	return isEntrance
}

func checkExit(eventDetails P.CounterEventFormat, pageName string) bool {
	isEntrance := false
	if url, ok := ExistsInProps(U.SP_LATEST_PAGE_URL, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		if url := url.(string); url == pageName {
			isEntrance = true
		}
	}
	return isEntrance
}
