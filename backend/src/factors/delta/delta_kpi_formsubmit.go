package delta

import (
	"bufio"
	"encoding/json"
	M "factors/model/model"
	P "factors/pattern"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

var formSubmitMetricToFunc = map[string]func(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error){
	M.Count:        GetFormSubmitCount,
	M.UniqueUsers:  GetFormSubmitUniqueUsers,
	M.CountPerUser: GetFormSubmitCountPerUser,
}

func GetFormSubmitMetrics(metricNames []string, queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*WithinPeriodInsightsKpi, error) {
	var wpi WithinPeriodInsightsKpi
	for i, metric := range metricNames {
		if i == 1 {
			break
		}
		if _, ok := formSubmitMetricToFunc[metric]; !ok {
			continue
		}
		if info, scale, err := formSubmitMetricToFunc[metric](queryEvent, scanner, propFilter, propsToEval); err != nil {
			log.WithError(err).Error("error getFormSubmitMetrics")
		} else {
			wpi.MetricInfo = info
			wpi.ScaleInfo = scale
		}
	}
	return &wpi, nil
}

func GetFormSubmitCount(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	var count float64
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
		//check if event is FormSubmit and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, U.EVENT_NAME_FORM_SUBMITTED, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		addValueToMapForPropsPresent(&count, reqMap, 1, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
	}
	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: count, Features: reqMap}

	return &info, &scale, nil
}

func GetFormSubmitUniqueUsers(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
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

		//check if event is FormSubmit and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, U.EVENT_NAME_FORM_SUBMITTED, propFilter); !ok {
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

func GetFormSubmitCountPerUser(queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*MetricInfo, *MetricInfo, error) {
	uniqueUsers := make(map[string]bool)
	uniqueUsersFeat := make(map[string]map[string]bool)
	featInfoMap := make(map[string]map[string]Fraction)
	var countPerUserFrac Fraction
	var countPerUser float64
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

		//check if event is FormSubmit and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, U.EVENT_NAME_FORM_SUBMITTED, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		addValuesToFractionForPropsPresentUser(&countPerUserFrac, featInfoMap, 1, 1, propsToEval, eventDetails, uniqueUsers, uniqueUsersFeat)
	}

	countPerUser, reqMap = getFractionValue(&countPerUserFrac, featInfoMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: countPerUser, Features: reqMap}
	return &info, &scale, nil
}
