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

type EventPropInfo struct {
	propFunc     func(eventDetails P.CounterEventFormat, prop string) (float64, bool, error)
	useProp      string
	defaultValue float64
	setBase      bool
}

type EventMetricCalculationInfo struct {
	PropsInfo []EventPropInfo
	useUnique bool
}

func getMetricToCalcInfoMap(queryEvent string) map[string]EventMetricCalculationInfo {
	switch queryEvent {
	case U.EVENT_NAME_SESSION:
		return sessionMetricToCalcInfo
	case U.EVENT_NAME_FORM_SUBMITTED:
		return formSubmitMetricToCalcInfo
	default:
		return pageviewMetricToCalcInfo
	}
}

func getEventMetricsInfo(metric string, queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*WithinPeriodInsightsKpi, error) {
	var wpi WithinPeriodInsightsKpi
	wpi.MetricInfo = &MetricInfo{}
	wpi.ScaleInfo = &MetricInfo{}

	var page string

	metricCalcInfo, ok := getMetricToCalcInfoMap(queryEvent)[metric]
	if !ok {
		err := fmt.Errorf("unknown event metric: %s", metric)
		log.WithError(err).Error("error GetEventMetricsInfo")
		return &wpi, err
	}
	var GetEventMetric func(queryEvent, page string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, propsInfo []EventPropInfo, useUnique bool) (*MetricInfo, *MetricInfo, error)
	if len(metricCalcInfo.PropsInfo) == 1 {
		GetEventMetric = getEventMetricSimple
	} else if len(metricCalcInfo.PropsInfo) == 2 {
		GetEventMetric = getEventMetricComplex
	} else {
		log.Error("error GetEventMetricsInfo: wrong propsInfo")
		return &wpi, fmt.Errorf("incorrect propsInfo for metric: %s", metric)
	}

	if metric == M.Entrances || metric == M.Exits {
		page = queryEvent
		queryEvent = U.EVENT_NAME_SESSION
	}
	if info, scale, err := GetEventMetric(queryEvent, page, scanner, propFilter, propsToEval, metricCalcInfo.PropsInfo, metricCalcInfo.useUnique); err != nil {
		log.WithError(err).Error("error GetEventMetric")
		return &wpi, err
	} else {
		wpi.MetricInfo = info
		wpi.ScaleInfo = scale
	}

	return &wpi, nil
}

func getEventMetricSimple(queryEvent, page string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, propsInfo []EventPropInfo, useUnique bool) (*MetricInfo, *MetricInfo, error) {
	var globalVal float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	//for unique implementation
	var uniqueUsers = make(map[string]bool)
	var uniqueUsersFeat = make(map[string]map[string]bool)

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}

		//check if event is session and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, queryEvent, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		var valueToAdd float64
		propInfo := propsInfo[0]
		if val, ok, err := checkPropInfo(eventDetails, propInfo, page); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		} else {
			valueToAdd = val
		}

		if valueToAdd != 0 {
			if useUnique {
				addValueToMapForPropsPresentUnique(&globalVal, reqMap, valueToAdd, propsToEval, eventDetails, uniqueUsers, uniqueUsersFeat)
			} else {
				addValueToMapForPropsPresent(&globalVal, reqMap, valueToAdd, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
			}
		}

	}
	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: globalVal, Features: reqMap}
	return &info, &scale, nil
}

func getEventMetricComplex(queryEvent, page string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, propsInfo []EventPropInfo, useUnique bool) (*MetricInfo, *MetricInfo, error) {
	var globalFrac Fraction
	var globalVal float64
	var globalScale float64
	var featInfoMap = make(map[string]map[string]Fraction)
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	//for unique implementation
	var uniqueUsers = make(map[string]bool)
	var uniqueUsersFeat = make(map[string]map[string]bool)

	for scanner.Scan() {
		txtline := scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}

		//check if event is session and contains all requiredProps(constraint)
		if ok, err := isEventToBeCounted(eventDetails, queryEvent, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScale(&globalScale, scaleMap, propsToEval, eventDetails)

		var numerValue float64
		var denomValue float64
		baseTrue := true
		for i, propInfo := range propsInfo {
			if val, ok, err := checkPropInfo(eventDetails, propInfo, page); !ok {
				if err != nil {
					return &info, &scale, err
				}
				if propInfo.setBase {
					baseTrue = false
					break
				}
				continue
			} else {
				if i == 0 {
					numerValue = val
				} else {
					denomValue = val
				}
			}
		}

		if baseTrue && !(numerValue == 0 && denomValue == 0) {
			if useUnique {
				addValuesToFractionForPropsPresentUnique(&globalFrac, featInfoMap, numerValue, denomValue, propsToEval, eventDetails, uniqueUsers, uniqueUsersFeat)
			} else {
				addValuesToFractionForPropsPresent(&globalFrac, featInfoMap, numerValue, denomValue, propsToEval, eventDetails.EventProperties, eventDetails.UserProperties)
			}
		}
	}

	// get bounce rate
	globalVal, reqMap = getFractionValue(&globalFrac, featInfoMap)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: globalVal, Features: reqMap}

	return &info, &scale, nil
}

// check if event name is correct and contains all required properties(satisfies constraints)
func isEventToBeCounted(eventDetails P.CounterEventFormat, nameFilter string, propFilter []M.KPIFilter) (bool, error) {
	eventNameString := eventDetails.EventName

	//check if event name is correct
	if eventNameString != nameFilter {
		return false, nil
	}

	//check if event contains all requiredProps(constraint)
	if ok, err := eventSatisfiesConstraints(eventDetails, propFilter); err != nil {
		return false, err
	} else if ok {
		return true, nil
	}

	return false, nil
}

// check if event contains all required properties(satisfies constraints)
func eventSatisfiesConstraints(eventDetails P.CounterEventFormat, propFilter []M.KPIFilter) (bool, error) {
	passFilter := true
	for _, filter := range propFilter {

		if filter.LogicalOp == "AND" {
			if !passFilter {
				return false, nil
			}
			passFilter = false
		} else if filter.LogicalOp == "OR" {
			if passFilter {
				continue
			}
		} else {
			return false, fmt.Errorf("unknown logical operation: %s", filter.LogicalOp)
		}

		var eventVal interface{}
		var propType string

		propName := filter.PropertyName
		if filter.Entity == M.EventEntity {
			propType = "ep"
		} else if filter.Entity == M.UserEntity {
			propType = "up"
		} else {
			return false, fmt.Errorf("strange entity of filter property %s - %s", propName, filter.Entity)
		}

		if val, ok := existsInProps(propName, eventDetails.EventProperties, eventDetails.UserProperties, propType); !ok {
			notOp, _, _ := U.StringIn(notOperations, filter.Condition)
			if notOp {
				passFilter = true
			}
			continue
		} else {
			eventVal = val
		}

		ok, err := checkValSatisfiesFilterCondition(filter, eventVal)
		if err != nil {
			return false, err
		}
		if ok {
			passFilter = true
		}
	}
	if !passFilter {
		return false, nil
	}
	return true, nil
}

// add 1 to globalScale and to scaleMap for all values from propsToEval properties found in eventDetails
func addToScale(globalScale *float64, scaleMap map[string]map[string]float64, propsToEval []string, eventDetails P.CounterEventFormat) {
	(*globalScale)++
	for _, propWithType := range propsToEval {
		propTypeName := strings.SplitN(propWithType, "#", 2)
		prop := propTypeName[1]
		propType := propTypeName[0]
		if val, ok := existsInProps(prop, eventDetails.EventProperties, eventDetails.UserProperties, propType); ok {
			val := fmt.Sprintf("%s", val)
			if _, ok := scaleMap[propWithType]; !ok {
				scaleMap[propWithType] = make(map[string]float64)
			}
			scaleMap[propWithType][val] += 1
		}
	}
}

func checkPropInfo(eventDetails P.CounterEventFormat, propInfo EventPropInfo, page string) (float64, bool, error) {
	yes := true
	getValue := false
	valueToAdd := propInfo.defaultValue
	if propInfo.propFunc != nil {
		var prop string
		if propInfo.useProp == "page" {
			prop = page
		} else if propInfo.useProp != "" {
			getValue = true
			prop = propInfo.useProp
		}
		val, ok, err := propInfo.propFunc(eventDetails, prop)
		if err != nil {
			return valueToAdd, false, err
		}
		yes = ok
		if getValue && ok {
			valueToAdd = val
		}
	}
	return valueToAdd, yes, nil
}

// check prop exists and get float value
func getPropValueEvents(eventDetails P.CounterEventFormat, propName string) (float64, bool, error) {
	if val, ok := existsInProps(propName, eventDetails.EventProperties, eventDetails.UserProperties, "either"); ok {
		if floatVal, err := getFloatValueFromInterface(val); err != nil {
			return 0, false, err
		} else {
			return floatVal, true, nil
		}
	} else {
		return 0, false, nil
	}
}
