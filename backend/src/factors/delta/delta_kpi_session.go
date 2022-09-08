package delta

import (
	"bufio"
	"encoding/json"
	M "factors/model/model"
	P "factors/pattern"
	U "factors/util"
	"fmt"

	log "github.com/sirupsen/logrus"
)

var sessionMetricToFunc = map[string]func(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error){
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

func GetSessionMetrics(metric string, queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*WithinPeriodInsightsKpi, error) {
	var wpi WithinPeriodInsightsKpi

	if _, ok := sessionMetricToFunc[metric]; !ok {
		err := fmt.Errorf("unknown session metric: %s", metric)
		log.WithError(err).Error("error GetSessionMetrics")
		return &wpi, err
	}
	if info, scale, err := sessionMetricToFunc[metric](queryEvent, scanner, propFilter, propsToEval); err != nil {
		log.WithError(err).Error("error GetSessionMetrics")
		return nil, err
	} else {
		wpi.MetricInfo = info
		wpi.ScaleInfo = scale
	}

	return &wpi, nil
}

func GetSessionTotalSessions(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var sessionsCount float64
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

		addValueToMapForPropsPresent(&sessionsCount, reqMap, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
	}
	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: sessionsCount, Features: reqMap}
	return &info, &scale, nil
}

func GetSessionUniqueUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
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

		//check if event is session and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, U.EVENT_NAME_SESSION, propFilter); !ok {
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

func GetSessionNewUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var new float64
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

		//check if new user
		if !checkNew(eventDetails) {
			continue
		}

		addValueToMapForPropsPresent(&new, reqMap, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)

	}
	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: new, Features: reqMap}

	return &info, &scale, nil
}

func GetSessionRepeatUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	var repeat float64
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

		//check if new user
		if checkNew(eventDetails) {
			continue
		}

		addValueToMapForPropsPresentUser(&repeat, reqMap, 1, propsToEval, eventDetails, uniqueUsers, uniqueUsersFeat)
	}

	deleteEntriesWithZeroFreq(reqMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: repeat, Features: reqMap}
	return &info, &scale, nil
}

func GetSessionSessionsPerUser(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	featInfoMap := make(map[string]map[string]Fraction)
	var sessionsPerUserFrac Fraction
	var sessionsPerUser float64
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

		addValuesToFractionForPropsPresentUser(&sessionsPerUserFrac, featInfoMap, 1, 1, propsToEval, eventDetails, uniqueUsers, uniqueUsersFeat)
	}

	//get sessionsPerUser
	sessionsPerUser, reqMap = getFractionValue(&sessionsPerUserFrac, featInfoMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: sessionsPerUser, Features: reqMap}
	return &info, &scale, nil
}

func GetSessionEngagedSessions(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
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

		//check if event is session and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, U.EVENT_NAME_SESSION, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		//check if engaged
		isEngaged := checkEngagedSession(eventDetails)
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

func GetSessionEngagedUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
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

		//check if event is session and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, U.EVENT_NAME_SESSION, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		//check if engaged
		isEngaged := checkEngagedSession(eventDetails)
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

func GetSessionEngagedSessionsPerUser(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	featInfoMap := make(map[string]map[string]Fraction)
	var sessionsPerUserFrac Fraction
	var sessionsPerUser float64
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

		//check if engaged
		isEngaged := checkEngagedSession(eventDetails)
		if !isEngaged {
			continue
		}

		addValuesToFractionForPropsPresentUser(&sessionsPerUserFrac, featInfoMap, 1, 1, propsToEval, eventDetails, uniqueUsers, uniqueUsersFeat)
	}

	//get sessionsPerUser
	sessionsPerUser, reqMap = getFractionValue(&sessionsPerUserFrac, featInfoMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: sessionsPerUser, Features: reqMap}
	return &info, &scale, nil
}

func GetSessionTotalTimeOnSite(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var totalSessionTime float64
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

		if timeSpent, ok := ExistsInProps(U.SP_SPENT_TIME, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
			timeOnSite := timeSpent.(float64)
			addValueToMapForPropsPresent(&totalSessionTime, reqMap, timeOnSite, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
		}
	}

	deleteEntriesWithZeroFreq(reqMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: totalSessionTime, Features: reqMap}

	return &info, &scale, nil
}

func GetSessionAvgSessionDuration(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var avgSessionDuration float64
	var avgSessionDurationFrac Fraction
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

		//check if event is session and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, U.EVENT_NAME_SESSION, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		if timeSpent, ok := ExistsInProps(U.SP_SPENT_TIME, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
			timeSpent := timeSpent.(float64)
			timeOnSite := float64(timeSpent)
			addValuesToFractionForPropsPresent(&avgSessionDurationFrac, featInfoMap, timeOnSite, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
		}
	}

	// get average session duration
	avgSessionDuration, reqMap = getFractionValue(&avgSessionDurationFrac, featInfoMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: avgSessionDuration, Features: reqMap}

	return &info, &scale, nil
}

func GetSessionAvgPageViewsPerSession(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var avgPageViewsPerSessionFrac Fraction
	var avgPageViewsPerSession float64
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

		//check if event is session and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, U.EVENT_NAME_SESSION, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		addValuesToFractionForPropsPresent(&avgPageViewsPerSessionFrac, featInfoMap, 0, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)

		//check if event has pageview count as property
		if cnt, ok := ExistsInProps(U.SP_PAGE_COUNT, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
			cnt := cnt.(float64)
			addValuesToFractionForPropsPresent(&avgPageViewsPerSessionFrac, featInfoMap, cnt, 0, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
		}
	}

	// get average PageViews Per Session
	avgPageViewsPerSession, reqMap = getFractionValue(&avgPageViewsPerSessionFrac, featInfoMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: avgPageViewsPerSession, Features: reqMap}

	return &info, &scale, nil
}

func GetSessionAvgInitialPageLoadTime(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var avgInitialPageLoadTime float64
	var avgInitialPageLoadTimeFrac Fraction
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

		//check if event is session and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, U.EVENT_NAME_SESSION, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		if time, ok := ExistsInProps(U.SP_INITIAL_PAGE_LOAD_TIME, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
			time := time.(float64)

			addValuesToFractionForPropsPresent(&avgInitialPageLoadTimeFrac, featInfoMap, time, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
		}
	}

	//get avgInitialPageLoadTime
	avgInitialPageLoadTime, reqMap = getFractionValue(&avgInitialPageLoadTimeFrac, featInfoMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: avgInitialPageLoadTime, Features: reqMap}

	return &info, &scale, nil
}

func GetSessionBounceRate(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var bounceRateFrac Fraction
	var bounceRate float64
	var featInfoMap = make(map[string]map[string]Fraction) //[]string = (bounceSessions,sessionsCount)
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

		addValuesToFractionForPropsPresent(&bounceRateFrac, featInfoMap, 0, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)

		//check if it is a bounced session
		if cnt, ok := ExistsInProps(U.SP_PAGE_COUNT, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
			cnt := int64(cnt.(float64))
			if cnt != 1 {
				continue
			}
		} else {
			continue
		}

		addValuesToFractionForPropsPresent(&bounceRateFrac, featInfoMap, 1, 0, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
	}

	// get bounce rate
	bounceRate, reqMap = getFractionValueForRate(&bounceRateFrac, featInfoMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: bounceRate, Features: reqMap}

	return &info, &scale, nil
}

func GetSessionEngagementRate(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var engagementRate float64
	var engagementRateFrac Fraction
	var globalScale float64
	var featInfoMap = make(map[string]map[string]Fraction) //[]string = (engagedSessions,sessionsCount)
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
		addValuesToFractionForPropsPresent(&engagementRateFrac, featInfoMap, 0, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)

		//check if engaged
		isEngaged := checkEngagedSession(eventDetails)
		if !isEngaged {
			continue
		}

		addValuesToFractionForPropsPresent(&engagementRateFrac, featInfoMap, 1, 0, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
	}

	// get engagement rate
	engagementRate, reqMap = getFractionValueForRate(&engagementRateFrac, featInfoMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: engagementRate, Features: reqMap}

	return &info, &scale, nil
}

func checkEngagedSession(eventDetails P.CounterEventFormat) bool {
	isEngaged := false
	if timeSpent, ok := ExistsInProps(U.SP_SPENT_TIME, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		timeSpent := timeSpent.(float64)
		if timeSpent > 10 {
			isEngaged = true
		}
	}
	if cnt, ok := ExistsInProps(U.SP_PAGE_COUNT, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		cnt := int64(cnt.(float64))
		if cnt > 2 {
			isEngaged = true
		}
	}
	return isEngaged
}

func checkNew(eventDetails P.CounterEventFormat) bool {
	var new bool
	if first, ok := ExistsInProps(U.SP_IS_FIRST_SESSION, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		new = first.(bool)
	}
	return new
}
