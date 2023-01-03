package delta

import (
	M "factors/model/model"
	"factors/model/store/memsql"
	"fmt"
	"strings"
)

func getRequiredDocumentTypes(channel string) []int {
	switch channel {
	case M.ADWORDS:
		return adwordsRequiredDocumentTypes
	case M.BINGADS:
		return bingadsRequiredDocumentTypes
	case M.FACEBOOK:
		return facebookRequiredDocumentTypes
	case M.LINKEDIN:
		return linkedinRequiredDocumentTypes
	case M.GOOGLE_ORGANIC:
		return googleOrganicRequiredDocumentTypes
	}
	return nil
}

func getConstantsInfo(channel string) map[string]string {
	switch channel {
	case M.ADWORDS:
		return adwordsConstantInfo
	case M.BINGADS:
		return bingadsConstantInfo
	case M.FACEBOOK:
		return facebookConstantInfo
	case M.LINKEDIN:
		return linkedinConstantInfo
	case M.GOOGLE_ORGANIC:
		return googleOrganicConstantInfo
	}
	return nil
}

func getMetricToCalculationinfo(channel string) map[string]MetricCalculationInfo {
	switch channel {
	case M.ADWORDS:
		return adwordsMetricToCalcInfo
	case M.BINGADS:
		return bingadsMetricToCalcInfo
	case M.FACEBOOK:
		return facebookMetricToCalcInfo
	case M.LINKEDIN:
		return linkedinMetricToCalcInfo
	case M.GOOGLE_ORGANIC:
		return googleOrganicMetricToCalcInfo
	}
	return nil
}

func getDocumentTypeAlias(channel string) map[string]int {
	switch channel {
	case M.ADWORDS:
		return M.AdwordsDocumentTypeAlias
	case M.BINGADS:
		return M.BingadsDocumentTypeAlias
	case M.FACEBOOK:
		return memsql.FacebookDocumentTypeAlias
	case M.LINKEDIN:
		return memsql.LinkedinDocumentTypeAlias
	case M.GOOGLE_ORGANIC:
		return map[string]int{"combined_performance_report": 1, "page_performance_report": 2}
	}
	return nil
}

func getFilterPropertyReportName(channel string) func(propName string, objectType string) (string, error) {
	switch channel {
	case M.ADWORDS:
		return getAdwordsFilterPropertyReportName
	case M.BINGADS:
		return getBingadsFilterPropertyReportName
	case M.FACEBOOK:
		return getFacebookFilterPropertyReportName
	case M.LINKEDIN:
		return getLinkedinFilterPropertyReportName
	case M.GOOGLE_ORGANIC:
		return getGoogleOrganicFilterPropertyReportName
	}
	return nil
}

func getPropertyFilterName(channel string) func(prop string) (string, error) {
	switch channel {
	case M.ADWORDS:
		return getAdwordsPropertyFilterName
	case M.BINGADS:
		return getBingadsPropertyFilterName
	case M.FACEBOOK:
		return getFacebookPropertyFilterName
	case M.LINKEDIN:
		return getLinkedinPropertyFilterName
	case M.GOOGLE_ORGANIC:
		return getGoogleOrganicPropertyFilterName
	}
	return nil
}

func getLevelAndtypeOfDoc(docType int, documentTypeAlias map[string]int, infoMap map[string]string) (int, string, bool) {

	for k, val := range documentTypeAlias {

		if val == docType {
			var level int
			var typeOfDoc string

			//level
			for levelStr, levelInt := range levelStrToIntAlias {
				if tmp, ok := infoMap[levelStr]; ok && strings.HasPrefix(k, tmp) {
					level = levelInt
				}
			}
			if level == 0 {
				return 0, "", false
			}

			//typeOfDoc
			if strings.HasSuffix(k, "performance_report") || strings.HasSuffix(k, "insights") {
				typeOfDoc = "insights"
			} else {
				typeOfDoc = "others"
			}
			return level, typeOfDoc, true
		}
	}
	return 0, "", false
}

func getQueryLevel(propFilter []M.KPIFilter) (int, error) {
	level := 3
	for _, filter := range propFilter {
		if tmp, ok := levelStrToIntAlias[filter.ObjectType]; ok {
			if tmp < level {
				level = tmp
			}
		} else {
			return 0, fmt.Errorf("error getQueryLevel: unknown filter object type - %s", filter.ObjectType)
		}
	}
	return level, nil
}

func addToAssociatedProps(campaignDetails CounterCampaignFormat, associatedProps map[int]map[string]map[string]interface{}, docLevel int, level_id string) error {
	if _, ok := associatedProps[docLevel]; !ok {
		associatedProps[docLevel] = make(map[string]map[string]interface{})
	}
	if level_id != "" && level_id != "0" {
		associatedProps[docLevel][level_id] = campaignDetails.Value
	}
	return nil
}

func getDocLevelProps(propsToEval []string, docLevel int, queryLevel int) []string {
	propsToEvalPerDoc := make([]string, 0)
	for _, prop := range propsToEval {
		propWithLevel := strings.SplitN(prop, "#", 2)
		level := levelStrToIntAlias[propWithLevel[0]]
		if level == docLevel {
			propsToEvalPerDoc = append(propsToEvalPerDoc, prop)
		}
	}
	return propsToEvalPerDoc
}

func getMetricValueFromFraction(globalFrac *Fraction, featInfoMap map[string]map[string]Fraction, metricCalcInfo MetricCalculationInfo) (float64, map[string]map[string]float64) {
	var globalVal float64
	globalVal, _ = getValueFromCalcInfo(globalFrac.Numerator, globalFrac.Denominator, metricCalcInfo)
	reqMap := make(map[string]map[string]float64)
	for prop, valMap := range featInfoMap {
		reqMap[prop] = make(map[string]float64)
		for val, info := range valMap {
			reqVal, _ := getValueFromCalcInfo(info.Numerator, info.Denominator, metricCalcInfo)
			if reqVal != 0 {
				reqMap[prop][val] = reqVal
			}
		}
		if len(reqMap[prop]) == 0 {
			delete(reqMap, prop)
		}
	}
	return globalVal, reqMap
}

func getMetricValue(global float64, featInfoMap map[string]map[string]float64, metricCalcInfo MetricCalculationInfo) (float64, map[string]map[string]float64) {
	var globalVal float64
	reqMap := make(map[string]map[string]float64)

	globalVal, _ = getValueFromCalcInfo(global, 1, metricCalcInfo)

	for prop, valMap := range featInfoMap {
		reqMap[prop] = make(map[string]float64)
		for val, freq := range valMap {
			reqVal, _ := getValueFromCalcInfo(freq, 1, metricCalcInfo)
			if reqVal != 0 {
				reqMap[prop][val] = reqVal
			}
		}
		if len(reqMap[prop]) == 0 {
			delete(reqMap, prop)
		}
	}
	return globalVal, reqMap
}

func getValueFromCalcInfo(firstVal float64, secondVal float64, metricCalcInfo MetricCalculationInfo) (float64, error) {
	if val, ok := metricCalcInfo.Props[0].ReplaceValue[firstVal]; ok {
		firstVal = val
	}
	if len(metricCalcInfo.Props) > 1 {
		if val, ok := metricCalcInfo.Props[1].ReplaceValue[secondVal]; ok {
			secondVal = val
		}
	}

	var reqVal float64
	if metricCalcInfo.Operation == "sum" {
		reqVal = firstVal
	} else {
		if firstVal == 0 || secondVal == 0 {
			reqVal = 0
		} else {
			reqVal = firstVal / secondVal
		}
	}

	for op, cons := range metricCalcInfo.Constants {
		if val, err := performOperation(op, reqVal, cons); err != nil {
			return 0, err
		} else {
			reqVal = val
		}
	}
	return reqVal, nil
}
