package delta

import (
	"bufio"
	"encoding/json"
	M "factors/model/model"
	"factors/model/store/memsql"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type PropInfo struct {
	Name      string
	Dependent string
}

type MetricCalculationInfo struct {
	Props     []PropInfo
	Operation string
	Constants map[string]float64
}

type CounterCampaignFormat struct {
	Id         string                 `json:"id"`
	Channel    string                 `json:"source"`
	Doctype    int                    `json:"type"`
	Timestamp  int64                  `json:"timestamp"`
	Value      map[string]interface{} `json:"value"`
	SmartProps map[string]interface{} `json:"sp"`
}

var requiredDocumentTypes = map[string]*[]int{
	M.ADWORDS:        &adwordsRequiredDocumentTypes,
	M.BINGADS:        &bingadsRequiredDocumentTypes,
	M.FACEBOOK:       &facebookRequiredDocumentTypes,
	M.LINKEDIN:       &linkedinRequiredDocumentTypes,
	M.GOOGLE_ORGANIC: &googleOrganicRequiredDocumentTypes,
}

var constantsInfo = map[string]*map[string]string{
	M.ADWORDS:        &adwordsConstantInfo,
	M.BINGADS:        &bingadsConstantInfo,
	M.FACEBOOK:       &facebookConstantInfo,
	M.LINKEDIN:       &linkedinConstantInfo,
	M.GOOGLE_ORGANIC: &googleOrganicConstantInfo,
}

var metricToCalculationinfo = map[string]*map[string]MetricCalculationInfo{
	M.ADWORDS:        &adwordsMetricToCalcInfo,
	M.BINGADS:        &bingadsMetricToCalcInfo,
	M.FACEBOOK:       &facebookMetricToCalcInfo,
	M.LINKEDIN:       &linkedinMetricToCalcInfo,
	M.GOOGLE_ORGANIC: &googleOrganicMetricToCalcInfo,
}

var documentTypeAlias = map[string]map[string]int{
	M.ADWORDS:        M.AdwordsDocumentTypeAlias,
	M.BINGADS:        M.BingadsDocumentTypeAlias,
	M.FACEBOOK:       memsql.FacebookDocumentTypeAlias,
	M.LINKEDIN:       memsql.LinkedinDocumentTypeAlias,
	M.GOOGLE_ORGANIC: {"combined_performance_report": 1, "page_performance_report": 2},
}

func getLevelAndtypeOfDoc(docType int, infoMap map[string]string, documentTypeAlias map[string]int, levelStrToInt map[string]int) (int, string, bool) {

	for k, val := range documentTypeAlias {
		if val == docType {
			var level int
			var typeOfDoc string

			//level
			if strings.HasPrefix(k, infoMap[memsql.CAFilterCampaign]) {
				level = levelStrToInt[infoMap[memsql.CAFilterCampaign]]
			} else if strings.HasPrefix(k, infoMap[memsql.CAFilterAdGroup]) {
				level = levelStrToInt[infoMap[memsql.CAFilterAdGroup]]
			} else if strings.HasPrefix(k, infoMap[memsql.CAFilterKeyword]) {
				level = levelStrToInt[infoMap[memsql.CAFilterKeyword]]
			} else {
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

func getQueryLevel(propFilter []M.KPIFilter, infoMap map[string]string, levelStrToInt map[string]int) int {
	level := levelStrToInt[infoMap[memsql.CAFilterCampaign]]
	for _, filter := range propFilter {
		if strings.HasPrefix(filter.PropertyName, infoMap[memsql.CAFilterAdGroup]) {
			level = levelStrToInt[infoMap[memsql.CAFilterAdGroup]]
		} else if strings.HasPrefix(filter.PropertyName, infoMap[memsql.CAFilterKeyword]) {
			level = levelStrToInt[infoMap[memsql.CAFilterKeyword]]
			break
		}
	}
	return level
}

func GetCampaignMetricsInfo(metricNames []string, channel string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*WithinPeriodInsightsKpi, error) {
	var wpi WithinPeriodInsightsKpi
	wpi.MetricInfo = &MetricInfo{}
	wpi.ScaleInfo = &MetricInfo{}

	infoMap := *(constantsInfo[channel])

	var levelStrToInt = make(map[string]int)
	levelStrToInt[infoMap[memsql.CAFilterCampaign]] = 3
	levelStrToInt[infoMap[memsql.CAFilterAdGroup]] = 2
	levelStrToInt[infoMap[memsql.CAFilterKeyword]] = 1

	var propsToEvalFiltered []string
	queryLevel := getQueryLevel(propFilter, infoMap, levelStrToInt)
	for _, prop := range propsToEval {
		propWithLevel := strings.SplitN(prop, "#", 2)
		level := levelStrToInt[propWithLevel[0]]
		if level <= queryLevel {
			propsToEvalFiltered = append(propsToEvalFiltered, prop)
		}
	}

	metricToCalcinfo := *(metricToCalculationinfo[channel])
	docTypeAlias := documentTypeAlias[channel]
	requiredDocTypes := *requiredDocumentTypes[channel]

	for i, metric := range metricNames {
		if i == 1 {
			break
		}
		var GetCampaignMetric func(metric string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, queryLevel int, metricCalcInfo MetricCalculationInfo, infoMap map[string]string, docTypeAlias map[string]int, levelStrToInt map[string]int, requiredDocTypes []int) (*MetricInfo, *MetricInfo, error)
		if info, ok := metricToCalcinfo[metric]; ok {
			if len(info.Props) == 1 {
				GetCampaignMetric = GetCampaignMetricSimple
			} else {
				GetCampaignMetric = GetCampaignMetricComplex
			}
		} else {
			log.Error("error metric calculation info not available for " + metric)
			continue
		}
		if info, scale, err := GetCampaignMetric(metric, scanner, propFilter, propsToEvalFiltered, queryLevel, metricToCalcinfo[metric], infoMap, docTypeAlias, levelStrToInt, requiredDocTypes); err != nil {
			log.WithError(err).Error("error GetCampaignMetric for kpi " + metric)
			return nil, err
		} else {
			wpi.MetricInfo = info
			wpi.ScaleInfo = scale
		}
	}
	return &wpi, nil
}

func GetCampaignMetricSimple(metric string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, queryLevel int, metricCalcInfo MetricCalculationInfo, infoMap map[string]string, docTypeAlias map[string]int, levelStrToInt map[string]int, requiredDocTypes []int) (*MetricInfo, *MetricInfo, error) {
	var globalVal float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo
	var associatedProps = make(map[int]map[string]map[string]interface{})

	for scanner.Scan() {
		txtline := scanner.Text()

		var campaignDetails CounterCampaignFormat
		if err := json.Unmarshal([]byte(txtline), &campaignDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}

		//check document type and filters
		if ok, err := isCampaignToBeCounted(campaignDetails, propFilter, requiredDocTypes); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		docLevel, typeOfDoc, ok := getLevelAndtypeOfDoc(campaignDetails.Doctype, infoMap, docTypeAlias, levelStrToInt)
		if !ok {
			continue
		}
		if docLevel > queryLevel {
			continue
		}
		level_id := campaignDetails.Id
		if typeOfDoc == "others" {
			if _, ok := associatedProps[docLevel]; !ok {
				associatedProps[docLevel] = make(map[string]map[string]interface{})
			}
			if level_id != "" && level_id != "0" {
				associatedProps[docLevel][level_id] = campaignDetails.Value
			}
			continue
		}

		var extraProps = make(map[string]interface{})
		for k, v := range campaignDetails.SmartProps {
			extraProps[k] = v
		}
		if level_id != "" && level_id != "0" {
			for k, v := range associatedProps[docLevel][level_id] {
				extraProps[k] = v
			}
		}

		// check dependent prop
		if metricCalcInfo.Props[0].Dependent != "" {
			if _, ok := ExistsInProps(metricCalcInfo.Props[0].Dependent, campaignDetails.Value, extraProps, "either"); !ok {
				continue
			}
		}

		var propVal float64
		//check metric prop
		if val, ok := ExistsInProps(metricCalcInfo.Props[0].Name, campaignDetails.Value, extraProps, "either"); ok {
			if floatVal, err := getFloatValueFromInterface(val); err != nil {
				return nil, nil, err
			} else {
				propVal = floatVal
			}
		}

		propsToEvalTmp := make([]string, 0)
		for _, prop := range propsToEval {
			propWithLevel := strings.SplitN(prop, "#", 2)
			level := levelStrToInt[propWithLevel[0]]
			if level == docLevel {
				propsToEvalTmp = append(propsToEvalTmp, prop)
			}
		}

		var useless float64
		if docLevel == queryLevel {
			addValueToMapForPropsPresent(&globalVal, reqMap, propVal, propsToEvalTmp, campaignDetails.Value, extraProps)
		} else {
			addValueToMapForPropsPresent(&useless, reqMap, propVal, propsToEvalTmp, campaignDetails.Value, extraProps)
		}

	}

	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: globalVal, Features: reqMap}
	return &info, &scale, nil
}

func GetCampaignMetricComplex(metric string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, queryLevel int, metricCalcInfo MetricCalculationInfo, infoMap map[string]string, docTypeAlias map[string]int, levelStrToInt map[string]int, requiredDocTypes []int) (*MetricInfo, *MetricInfo, error) {
	var globalVal float64
	var globalFrac Fraction
	var globalScale float64
	var featInfoMap = make(map[string]map[string]Fraction)
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo
	var associatedProps = make(map[int]map[string]map[string]interface{})

	for scanner.Scan() {
		txtline := scanner.Text()

		var campaignDetails CounterCampaignFormat
		if err := json.Unmarshal([]byte(txtline), &campaignDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}

		//check if event is session and contains all requiredProps(constraint)
		if ok, err := isCampaignToBeCounted(campaignDetails, propFilter, requiredDocTypes); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}

		docLevel, typeOfDoc, ok := getLevelAndtypeOfDoc(campaignDetails.Doctype, infoMap, docTypeAlias, levelStrToInt)
		if !ok {
			continue
		}

		if docLevel > queryLevel {
			continue
		}
		level_id := campaignDetails.Id
		if typeOfDoc == "others" {
			if _, ok := associatedProps[docLevel]; !ok {
				associatedProps[docLevel] = make(map[string]map[string]interface{})
			}
			if level_id != "" && level_id != "0" {
				associatedProps[docLevel][level_id] = campaignDetails.Value
			}
			continue
		}

		var extraProps = make(map[string]interface{})
		for k, v := range campaignDetails.SmartProps {
			extraProps[k] = v
		}
		if level_id != "" && level_id != "0" {
			for k, v := range associatedProps[docLevel][level_id] {
				extraProps[k] = v
			}
		}

		propsToEvalTmp := make([]string, 0)
		for _, prop := range propsToEval {
			propWithLevel := strings.SplitN(prop, "#", 2)
			level := levelStrToInt[propWithLevel[0]]
			if level == docLevel {
				propsToEvalTmp = append(propsToEvalTmp, prop)
			}
		}

		var propVal float64
		// check dependent prop
		if metricCalcInfo.Props[0].Dependent != "" {
			if _, ok := ExistsInProps(metricCalcInfo.Props[0].Dependent, campaignDetails.Value, extraProps, "either"); !ok {
				continue
			}
		}
		//check metric prop
		if val, ok := ExistsInProps(metricCalcInfo.Props[0].Name, campaignDetails.Value, extraProps, "either"); ok {
			if floatVal, err := getFloatValueFromInterface(val); err != nil {
				return nil, nil, err
			} else {
				propVal = floatVal
			}
		}
		if docLevel == queryLevel {
			addValuesToFractionForPropsPresent(&globalFrac, featInfoMap, propVal, 0, propsToEvalTmp, campaignDetails.Value, extraProps)
		} else {
			addValuesToFractionForPropsPresent(&Fraction{}, featInfoMap, propVal, 0, propsToEvalTmp, campaignDetails.Value, extraProps)
		}

		// check dependent prop
		if metricCalcInfo.Props[1].Dependent != "" {
			if _, ok := ExistsInProps(metricCalcInfo.Props[1].Dependent, campaignDetails.Value, extraProps, "either"); !ok {
				continue
			}
		}
		//check metric prop
		if val, ok := ExistsInProps(metricCalcInfo.Props[1].Name, campaignDetails.Value, extraProps, "either"); ok {
			if floatVal, err := getFloatValueFromInterface(val); err != nil {
				return nil, nil, err
			} else {
				propVal = floatVal
			}
		}

		if docLevel == queryLevel {
			addValuesToFractionForPropsPresent(&globalFrac, featInfoMap, 0, propVal, propsToEvalTmp, campaignDetails.Value, extraProps)
		} else {
			addValuesToFractionForPropsPresent(&Fraction{}, featInfoMap, 0, propVal, propsToEvalTmp, campaignDetails.Value, extraProps)
		}

	}
	globalVal, reqMap = getMetricValue(&globalFrac, featInfoMap, metricCalcInfo)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: globalVal, Features: reqMap}
	return &info, &scale, nil
}

func getMetricValue(globalFrac *Fraction, featInfoMap map[string]map[string]Fraction, metricCalcInfo MetricCalculationInfo) (float64, map[string]map[string]float64) {
	var globalVal float64
	if globalFrac.Denominator != 0 {
		globalVal = getValueFromCalcInfo(globalFrac.Numerator, globalFrac.Denominator, metricCalcInfo)
	}
	reqMap := make(map[string]map[string]float64)
	for prop, valMap := range featInfoMap {
		reqMap[prop] = make(map[string]float64)
		for val, info := range valMap {
			if !(info.Denominator == 0 || info.Numerator == 0) {
				reqMap[prop][val] = getValueFromCalcInfo(info.Numerator, info.Denominator, metricCalcInfo)
			}
		}
		if len(reqMap[prop]) == 0 {
			delete(reqMap, prop)
		}
	}
	return globalVal, reqMap
}

func getValueFromCalcInfo(firstVal float64, secondVal float64, metricCalcInfo MetricCalculationInfo) float64 {
	var reqVal float64
	reqVal = firstVal / secondVal
	for op, cons := range metricCalcInfo.Constants {
		if op == "sum" {
			reqVal = reqVal + cons
		} else if op == "quotient" {
			reqVal = reqVal / cons
		} else if op == "product" {
			reqVal = reqVal * cons
		} else if op == "difference" {
			reqVal = reqVal - cons
		}
	}
	return reqVal
}

func getFloatValueFromInterface(inter interface{}) (float64, error) {
	if inter == nil {
		return 0, nil
	}
	switch val := inter.(type) {
	case float64:
		return val, nil
	case string:
		interFloat, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, err
		}
		return interFloat, nil
	}
	return 0, fmt.Errorf("interface type unknown - cant convert to float")
}

// func getDocLevelId(typeOfDoc string, docLevel int, campaignDetails CounterCampaignFormat, infoMap map[string]string) (float64, error) {
// 	var idKey string
// 	if typeOfDoc == "others" {
// 		idKey = infoMap["id"]
// 	}
// 	if docLevel == 3 {
// 		idKey = infoMap["campaign_id"]
// 	} else if docLevel == 2 {
// 		idKey = infoMap["ad_group_id"]
// 	} else if docLevel == 1 {
// 		idKey = infoMap["keyword_id"]
// 	}
// 	return getFloatValueFromInterface(campaignDetails.Value[idKey])
// }
