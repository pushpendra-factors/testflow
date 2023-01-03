package delta

import (
	M "factors/model/model"
	P "factors/pattern"
	U "factors/util"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

//check if prop exists in props based on entity and returns the interface value (put eventprops in first map, user in second)
func ExistsInProps(prop string, firstMap map[string]interface{}, secondMap map[string]interface{}, entity string) (interface{}, bool) {
	if firstMap != nil && (entity == "ep" || entity == "either") {
		if val, ok := firstMap[prop]; ok {
			return val, true
		}
	}
	if secondMap != nil && (entity == "up" || entity == "either") {
		if val, ok := secondMap[prop]; ok {
			return val, true
		}
	}
	return nil, false
}

func addValueToMapForPropsPresent(globalVal *float64, featMap map[string]map[string]float64, valueToBeAdded float64, propsToEval []string, propMap1 map[string]interface{}, propMap2 map[string]interface{}) {
	(*globalVal) += valueToBeAdded
	for _, propWithType := range propsToEval {
		propTypeName := strings.SplitN(propWithType, "#", 2)
		prop := propTypeName[1]
		propType := propTypeName[0]
		var pt string
		if propType == "up" || propType == "ep" {
			pt = propType
		} else {
			pt = "either"
		}
		if val, ok := ExistsInProps(prop, propMap1, propMap2, pt); ok {
			val, err := getStringValueFromInterface(val)
			if err != nil {
				log.WithError(err).Errorf("error getStringValueFromInterface for key %s and val %s", prop, val)
			}
			if _, ok := featMap[propWithType]; !ok {
				featMap[propWithType] = make(map[string]float64)
			}
			featMap[propWithType][val] += valueToBeAdded
		}
	}
}

func addValueToMapForPropsPresentUnique(globalVal *float64, featMap map[string]map[string]float64, valueToBeAdded float64, propsToEval []string, eventDetails P.CounterEventFormat, uniqueUsersGlobal map[string]bool, uniqueUsersFeat map[string]map[string]bool) {
	uid := eventDetails.UserId
	if _, ok := uniqueUsersGlobal[uid]; !ok {
		uniqueUsersGlobal[uid] = true
		*globalVal += valueToBeAdded
	}
	for _, propWithType := range propsToEval {
		propTypeName := strings.SplitN(propWithType, "#", 2)
		prop := propTypeName[1]
		propType := propTypeName[0]
		if val, ok := ExistsInProps(prop, eventDetails.EventProperties, eventDetails.UserProperties, propType); ok {
			val, err := getStringValueFromInterface(val)
			if err != nil {
				log.WithError(err).Errorf("error getStringValueFromInterface for key %s and val %s", prop, val)
			}
			propWithVal := strings.Join([]string{prop, val}, ":")
			if _, ok := uniqueUsersFeat[propWithVal]; !ok {
				uniqueUsersFeat[propWithVal] = make(map[string]bool)
			}
			if _, ok := uniqueUsersFeat[propWithVal][uid]; !ok {
				uniqueUsersFeat[propWithVal][uid] = true
				if _, ok := featMap[propWithType]; !ok {
					featMap[propWithType] = make(map[string]float64)
				}
				featMap[propWithType][val] += valueToBeAdded
			}
		}
	}
}

func addValuesToFractionForPropsPresent(globalVal *Fraction, featMap map[string]map[string]Fraction, numerValueToBeAdded float64, denomValueToBeAdded float64, propsToEval []string, firstMap map[string]interface{}, secondMap map[string]interface{}) {
	(*globalVal).Numerator += numerValueToBeAdded
	(*globalVal).Denominator += denomValueToBeAdded
	for _, propWithType := range propsToEval {
		propTypeName := strings.SplitN(propWithType, "#", 2)
		prop := propTypeName[1]
		propType := propTypeName[0]
		var pt string
		if propType == "up" || propType == "ep" {
			pt = propType
		} else {
			pt = "either"
		}
		if val, ok := ExistsInProps(prop, firstMap, secondMap, pt); ok {
			val, err := getStringValueFromInterface(val)
			if err != nil {
				log.WithError(err).Errorf("error getStringValueFromInterface for key %s and val %s", prop, val)
			}
			if _, ok := featMap[propWithType]; !ok {
				featMap[propWithType] = make(map[string]Fraction)
			}
			if frac, ok := featMap[propWithType][val]; !ok {
				featMap[propWithType][val] = Fraction{Numerator: numerValueToBeAdded, Denominator: denomValueToBeAdded}
			} else {
				frac.Numerator += numerValueToBeAdded
				frac.Denominator += denomValueToBeAdded
				featMap[propWithType][val] = frac
			}
		}
	}
}

func addValuesToFractionForPropsPresentUnique(globalVal *Fraction, featMap map[string]map[string]Fraction, numerValueToBeAdded float64, denomValueToBeAdded float64, propsToEval []string, eventDetails P.CounterEventFormat, uniqueUsersGlobal map[string]bool, uniqueUsersFeat map[string]map[string]bool) {
	uid := eventDetails.UserId
	(*globalVal).Numerator += numerValueToBeAdded
	if _, ok := uniqueUsersGlobal[uid]; !ok {
		uniqueUsersGlobal[uid] = true
		(*globalVal).Denominator += denomValueToBeAdded
	}
	for _, propWithType := range propsToEval {
		propTypeName := strings.SplitN(propWithType, "#", 2)
		prop := propTypeName[1]
		propType := propTypeName[0]
		if val, ok := ExistsInProps(prop, eventDetails.EventProperties, eventDetails.UserProperties, propType); ok {
			val, err := getStringValueFromInterface(val)
			if err != nil {
				log.WithError(err).Errorf("error getStringValueFromInterface for key %s and val %s", prop, val)
			}
			if _, ok := featMap[propWithType]; !ok {
				featMap[propWithType] = make(map[string]Fraction)
			}
			if frac, ok := featMap[propWithType][val]; !ok {
				featMap[propWithType][val] = Fraction{Numerator: numerValueToBeAdded}
			} else {
				frac.Numerator += numerValueToBeAdded
				featMap[propWithType][val] = frac
			}
			propWithVal := strings.Join([]string{prop, val}, ":")
			if _, ok := uniqueUsersFeat[propWithVal]; !ok {
				uniqueUsersFeat[propWithVal] = make(map[string]bool)
			}
			if _, ok := uniqueUsersFeat[propWithVal][uid]; !ok {
				uniqueUsersFeat[propWithVal][uid] = true
				frac := featMap[propWithType][val]
				frac.Denominator += denomValueToBeAdded
				featMap[propWithType][val] = frac
			}
		}
	}
}

//delete keys and values with zero frequency
func deleteEntriesWithZeroFreq(reqMap map[string]map[string]float64) {
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
}

func getFractionValue(globalFrac *Fraction, featInfoMap map[string]map[string]Fraction) (float64, map[string]map[string]float64) {
	var globalVal float64
	if globalFrac.Denominator != 0 {
		globalVal = globalFrac.Numerator / globalFrac.Denominator
	}
	reqMap := make(map[string]map[string]float64)
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
	return globalVal, reqMap
}

func getFractionValueForRate(globalFrac *Fraction, featInfoMap map[string]map[string]Fraction) (float64, map[string]map[string]float64) {
	var globalVal float64
	if globalFrac.Denominator != 0 {
		globalVal = globalFrac.Numerator * 100 / globalFrac.Denominator
	}
	reqMap := make(map[string]map[string]float64)
	for prop, valMap := range featInfoMap {
		reqMap[prop] = make(map[string]float64)
		for val, info := range valMap {
			if !(info.Denominator == 0 || info.Numerator == 0) {
				reqMap[prop][val] = info.Numerator * 100 / info.Denominator
			}
		}
		if len(reqMap[prop]) == 0 {
			delete(reqMap, prop)
		}
	}
	return globalVal, reqMap
}

func checkValSatisfiesFilterCondition(filter M.KPIFilter, eventVal interface{}) (bool, error) {
	if filter.PropertyDataType == U.PropertyTypeCategorical {
		eventVal := eventVal.(string)
		if filter.Condition == M.EqualsOpStr {
			if eventVal != filter.Value {
				return false, nil
			}
		} else if filter.Condition == M.NotEqualOpStr {
			if eventVal == filter.Value {
				return false, nil
			}
		} else if filter.Condition == M.ContainsOpStr {
			if !strings.Contains(eventVal, filter.Value) {
				return false, nil
			}
		} else if filter.Condition == M.NotContainsOpStr {
			if strings.Contains(eventVal, filter.Value) {
				return false, nil
			}
		} else {
			return false, fmt.Errorf("unknown filter condition - %s", filter.Condition)
		}
	} else if filter.PropertyDataType == U.PropertyTypeNumerical {
		eventVal := eventVal.(float64)
		filterVal, err := strconv.ParseFloat(filter.Value, 64)
		if err != nil {
			log.WithError(err).Error("error Decoding Float64 filter value")
			return false, err
		}
		if filter.Condition == M.EqualsOp {
			if eventVal != filterVal {
				return false, nil
			}
		} else if filter.Condition == M.NotEqualOp {
			if eventVal == filterVal {
				return false, nil
			}
		} else if filter.Condition == M.LesserThanOpStr {
			if eventVal >= filterVal {
				return false, nil
			}
		} else if filter.Condition == M.LesserThanOrEqualOpStr {
			if eventVal > filterVal {
				return false, nil
			}
		} else if filter.Condition == M.GreaterThanOpStr {
			if eventVal <= filterVal {
				return false, nil
			}
		} else if filter.Condition == M.GreaterThanOrEqualOpStr {
			if eventVal < filterVal {
				return false, nil
			}
		} else {
			return false, fmt.Errorf("unknown filter condition - %s", filter.Condition)
		}
	} else if filter.PropertyDataType == U.PropertyTypeDateTime {
		eventVal := eventVal.(float64)

		dateTimeFilter, err := M.DecodeDateTimePropertyValue(filter.Value)
		if err != nil {
			log.WithError(err).Error("error Decoding filter value")
			return false, err
		}
		propVal := fmt.Sprintf("%v", int64(eventVal))
		propertyValue, _ := strconv.ParseInt(propVal, 10, 64)
		if !(propertyValue >= dateTimeFilter.From && propertyValue <= dateTimeFilter.To) {
			return false, nil
		}
	} else if filter.PropertyDataType == U.PropertyTypeUnknown {
		return false, fmt.Errorf("property type unknown for %s", filter.PropertyName)
	} else {
		return false, fmt.Errorf("strange property type: %s", filter.PropertyDataType)
	}
	return true, nil
}

func MultipleJSDivergenceKpi(metricInfo1, metricInfo2 *MetricInfo, allProps map[string]map[string]bool) Level2CatRatioDist {
	jsdMetrics := make(Level2CatRatioDist)
	for key, valMap := range metricInfo1.Features {
		jsdMetrics[key] = make(Level1CatRatioDist)
		for val := range valMap {
			prev1 := metricInfo1.Features[key][val] / metricInfo1.Global
			prev2 := metricInfo2.Features[key][val] / metricInfo2.Global
			jsd := SingleJSDivergence(prev1, prev2)
			jsdMetrics[key][val] = jsd
		}
	}
	return jsdMetrics
}
