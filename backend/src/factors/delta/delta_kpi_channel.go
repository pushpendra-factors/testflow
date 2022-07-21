package delta

import (
	"bufio"
	"encoding/json"
	"factors/filestore"
	M "factors/model/model"
	"factors/model/store/memsql"
	serviceDisk "factors/services/disk"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type PropInfo struct {
	Name               string
	DependentKey       string
	DependentValue     float64
	DependentOperation string
	ReplaceValue       map[float64]float64
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

var ChannelValueFilterName = map[string]string{
	M.ADWORDS:        "Google Ads",
	M.BINGADS:        "Bing Ads",
	M.LINKEDIN:       "LinkedIn Ads",
	M.FACEBOOK:       "Facebook Ads",
	M.GOOGLE_ORGANIC: "Google Ads",
}

var levelStrToIntAlias = map[string]int{
	memsql.CAFilterCampaign: 3,
	memsql.CAFilterAdGroup:  2,
	memsql.CAFilterKeyword:  1,
	memsql.CAFilterAd:       1,
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

var getFilterPropertyReportName = map[string]func(propName string, objectType string) (string, error){
	M.ADWORDS:        getAdwordsFilterPropertyReportName,
	M.BINGADS:        getBingadsFilterPropertyReportName,
	M.FACEBOOK:       getFacebookFilterPropertyReportName,
	M.LINKEDIN:       getLinkedinFilterPropertyReportName,
	M.GOOGLE_ORGANIC: getGoogleOrganicFilterPropertyReportName,
}

var getPropertyFilterName = map[string]func(prop string) (string, error){
	M.ADWORDS:        getAdwordsPropertyFilterName,
	M.BINGADS:        getBingadsPropertyFilterName,
	M.FACEBOOK:       getFacebookPropertyFilterName,
	M.LINKEDIN:       getLinkedinPropertyFilterName,
	M.GOOGLE_ORGANIC: getGoogleOrganicPropertyFilterName,
}

func getLevelAndtypeOfDoc(docType int, documentTypeAlias map[string]int, infoMap map[string]string) (int, string, bool) {

	for k, val := range documentTypeAlias {
		if val == docType {
			var level int
			var typeOfDoc string

			//level
			for levelStr, levelInt := range levelStrToIntAlias {
				if strings.HasPrefix(k, infoMap[levelStr]) {
					if levelInt > level {
						level = levelInt
					}
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
		} else if filter.ObjectType == "organic_property" {
			return 3, nil
		} else {
			return 0, fmt.Errorf("error getQueryLevel: unknown filter object type - %s", filter.ObjectType)
		}
	}
	return level, nil
}

func GetCampaignMetricsInfo(metricNames []string, channel string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*WithinPeriodInsightsKpi, error) {
	var wpi WithinPeriodInsightsKpi
	wpi.MetricInfo = &MetricInfo{}
	wpi.ScaleInfo = &MetricInfo{}

	var newPropFilter []M.KPIFilter
	for _, filter := range propFilter {
		name, err := getFilterPropertyReportName[channel](filter.PropertyName, filter.ObjectType)
		if err != nil {
			return nil, err
		}
		filter.PropertyName = name
		newPropFilter = append(newPropFilter, filter)
	}

	infoMap := *(constantsInfo[channel])

	var propsToEvalPerQuery []string
	queryLevel, err := getQueryLevel(newPropFilter)
	if err != nil {
		return nil, err
	}
	for _, prop := range propsToEval {
		propWithLevel := strings.SplitN(prop, "#", 2)
		level := levelStrToIntAlias[propWithLevel[0]]
		if level <= queryLevel {
			propsToEvalPerQuery = append(propsToEvalPerQuery, prop)
		}
	}

	metricToCalcinfo := *(metricToCalculationinfo[channel])
	docTypeAlias := documentTypeAlias[channel]
	requiredDocTypes := *requiredDocumentTypes[channel]

	for i, metric := range metricNames {
		if i == 1 {
			break
		}
		if channel == M.ADWORDS {
			if metricInt, ok := M.AdwordsExtToInternal[metric]; ok {
				metric = metricInt
			}
		}
		var GetCampaignMetric func(metric string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, queryLevel int, metricCalcInfo MetricCalculationInfo, docTypeAlias map[string]int, requiredDocTypes []int, infoMap map[string]string) (*MetricInfo, *MetricInfo, error)
		if info, ok := metricToCalcinfo[metric]; ok {
			if info.Operation == "sum" {
				GetCampaignMetric = GetCampaignMetricSimple
			} else {
				GetCampaignMetric = GetCampaignMetricComplex
			}
		} else {
			log.Error("error metric calculation info not available for " + metric)
			continue
		}
		if info, scale, err := GetCampaignMetric(metric, scanner, newPropFilter, propsToEvalPerQuery, queryLevel, metricToCalcinfo[metric], docTypeAlias, requiredDocTypes, infoMap); err != nil {
			log.WithError(err).Error("error GetCampaignMetric for kpi " + metric)
			return nil, err
		} else {
			for key, valMap := range info.Features {
				newKey, _ := getPropertyFilterName[channel](key)
				delete(info.Features, key)
				info.Features[newKey] = valMap
			}
			wpi.MetricInfo = info
			wpi.ScaleInfo = scale
		}
	}
	return &wpi, nil
}

func GetCampaignMetricSimple(metric string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, queryLevel int, metricCalcInfo MetricCalculationInfo, docTypeAlias map[string]int, requiredDocTypes []int, infoMap map[string]string) (*MetricInfo, *MetricInfo, error) {
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
		docLevel, typeOfDoc, ok := getLevelAndtypeOfDoc(campaignDetails.Doctype, docTypeAlias, infoMap)
		if !ok {
			continue
		}
		if docLevel > queryLevel {
			continue
		}
		level_id := campaignDetails.Id
		if typeOfDoc == "others" {
			addToAssociatedProps(campaignDetails, associatedProps, docLevel, level_id)
			continue
		}

		extraProps, _ := getExtraProps(campaignDetails, level_id, associatedProps, docLevel)

		propsToEvalPerDoc := getDocLevelProps(propsToEval, docLevel, queryLevel)
		if len(propsToEvalPerDoc) == 0 {
			continue
		}

		var propVal float64
		var depVal float64
		// check dependent prop
		if val, ok, err := checkDependentProp(metricCalcInfo.Props[0], campaignDetails, extraProps); !ok {
			if err != nil {
				return nil, nil, err
			}
			continue
		} else {
			depVal = val
		}

		//check metric prop
		if val, ok, err := getPropValue(metricCalcInfo.Props[0], campaignDetails, extraProps); !ok {
			if err != nil {
				return nil, nil, err
			}
			continue
		} else {
			propVal = val
		}

		if metricCalcInfo.Props[0].DependentOperation == "product" {
			if val, err := performOperation(metricCalcInfo.Props[0].DependentOperation, propVal, depVal); err == nil {
				propVal = val
			}
		}

		var useless float64
		if propVal != 0 {
			if docLevel == queryLevel {
				addValueToMapForPropsPresent(&globalVal, reqMap, propVal, propsToEvalPerDoc, campaignDetails.Value, extraProps)
			} else {
				addValueToMapForPropsPresent(&useless, reqMap, propVal, propsToEvalPerDoc, campaignDetails.Value, extraProps)
			}
		}

	}

	//TODO: add replace values and constants logic (if a metric having non-nil ReplaceValues OR Constants exists for simple case)
	globalVal, reqMap = getMetricValue(globalVal, reqMap, metricCalcInfo)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: globalVal, Features: reqMap}
	return &info, &scale, nil
}

func GetCampaignMetricComplex(metric string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, queryLevel int, metricCalcInfo MetricCalculationInfo, docTypeAlias map[string]int, requiredDocTypes []int, infoMap map[string]string) (*MetricInfo, *MetricInfo, error) {
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

		docLevel, typeOfDoc, ok := getLevelAndtypeOfDoc(campaignDetails.Doctype, docTypeAlias, infoMap)
		if !ok {
			continue
		}

		if docLevel > queryLevel {
			continue
		}

		level_id := campaignDetails.Id

		if typeOfDoc == "others" {
			addToAssociatedProps(campaignDetails, associatedProps, docLevel, level_id)
			continue
		}

		extraProps, _ := getExtraProps(campaignDetails, level_id, associatedProps, docLevel)

		propsToEvalPerDoc := getDocLevelProps(propsToEval, docLevel, queryLevel)
		if len(propsToEvalPerDoc) == 0 {
			continue
		}

		var numerValue float64
		var denomValue float64
		for i, propInfo := range metricCalcInfo.Props {
			var propVal float64
			var depVal float64
			// check dependent prop
			if val, ok, err := checkDependentProp(propInfo, campaignDetails, extraProps); !ok {
				if err != nil {
					return nil, nil, err
				}
				continue
			} else {
				depVal = val
			}

			//check metric prop
			if val, ok, err := getPropValue(propInfo, campaignDetails, extraProps); !ok {
				if err != nil {
					return nil, nil, err
				}
				continue
			} else {
				propVal = val
			}

			if propInfo.DependentOperation == "product" {
				if val, err := performOperation(propInfo.DependentOperation, propVal, depVal); err == nil {
					propVal = val
				}
			}

			if i == 0 {
				numerValue = propVal
				if metricCalcInfo.Operation == "average" {
					denomValue = 1
				}
			} else {
				denomValue = propVal
			}
		}

		if !(numerValue == 0 && denomValue == 0) {
			if docLevel == queryLevel {
				addValuesToFractionForPropsPresent(&globalFrac, featInfoMap, numerValue, denomValue, propsToEvalPerDoc, campaignDetails.Value, extraProps)
			} else {
				addValuesToFractionForPropsPresent(&Fraction{}, featInfoMap, numerValue, denomValue, propsToEvalPerDoc, campaignDetails.Value, extraProps)
			}
		}
	}
	globalVal, reqMap = getMetricValueFromFraction(&globalFrac, featInfoMap, metricCalcInfo)

	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: globalVal, Features: reqMap}
	return &info, &scale, nil
}

//check dependent prop
func checkDependentProp(propInfo PropInfo, campaignDetails CounterCampaignFormat, extraProps map[string]interface{}) (float64, bool, error) {
	var reqVal float64
	key := propInfo.DependentKey
	val := propInfo.DependentValue
	op := propInfo.DependentOperation
	if key != "" {
		if docVal, ok := ExistsInProps(key, campaignDetails.Value, extraProps, "either"); !ok {
			return 0, false, nil
		} else if op != "" {
			docVal, err := getFloatValueFromInterface(docVal)
			if err != nil {
				return 0, false, err
			}
			if op == M.NotEqualOp && docVal == val {
				return docVal, false, nil
			} else if op == M.EqualsOp && docVal != val {
				return docVal, false, nil
			}
			reqVal = docVal
		}
	}

	return reqVal, true, nil
}

//check metric prop
func getPropValue(propInfo PropInfo, campaignDetails CounterCampaignFormat, extraProps map[string]interface{}) (float64, bool, error) {
	if val, ok := ExistsInProps(propInfo.Name, campaignDetails.Value, extraProps, "either"); ok {
		if floatVal, err := getFloatValueFromInterface(val); err != nil {
			return 0, false, err
		} else {
			return floatVal, true, nil
		}
	} else {
		return 0, false, nil
	}
}

//get smart props + associated props
func getExtraProps(campaignDetails CounterCampaignFormat, level_id string, associatedProps map[int]map[string]map[string]interface{}, docLevel int) (map[string]interface{}, error) {
	var extraProps = make(map[string]interface{})
	for k, v := range campaignDetails.SmartProps {
		extraProps[k] = v
	}
	if level_id != "" && level_id != "0" {
		for k, v := range associatedProps[docLevel][level_id] {
			extraProps[k] = v
		}
	}
	return extraProps, nil
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
	case int:
		return float64(val), nil
	}

	return 0, fmt.Errorf("interface type unknown - cant convert to float")
}

func getStringValueFromInterface(inter interface{}) (string, error) {
	if inter == nil {
		return "", nil
	}
	switch val := inter.(type) {
	case float64:
		return fmt.Sprintf("%f", val), nil
	case string:
		return val, nil
	case int:
		return fmt.Sprintf("%d", val), nil
	}
	return "", fmt.Errorf("interface type unknown - cant convert to string")
}

func performOperation(operation string, val1 float64, val2 float64) (float64, error) {
	var reqVal float64
	if operation == "sum" {
		reqVal = val1 + val2
	} else if operation == "quotient" {
		reqVal = val1 / val2
	} else if operation == "product" {
		reqVal = val1 * val2
	} else if operation == "difference" {
		reqVal = val1 - val2
	} else {
		return 0, fmt.Errorf("unknown operation : %s", operation)
	}
	return reqVal, nil
}

func GetAllChannelMetricsInfo(metricNames []string, channel string, propFilter []M.KPIFilter, propsToEval []string, projectId int64, periodCode Period, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, insightGranularity string) (*WithinPeriodInsightsKpi, error) {
	var wpi WithinPeriodInsightsKpi
	wpi.MetricInfo = &MetricInfo{}
	wpi.ScaleInfo = &MetricInfo{}
	wpi.ScaleInfo.Features = make(map[string]map[string]float64)

	for _, channel := range []string{M.ADWORDS, M.BINGADS, M.LINKEDIN, M.FACEBOOK, M.GOOGLE_ORGANIC} {
		passFilter := true
		var newPropFilter []M.KPIFilter
		for _, filter := range propFilter {
			if filter.ObjectType == "channel" {
				if ok, _ := checkValSatisfiesFilterCondition(filter, ChannelValueFilterName[channel]); !ok {
					passFilter = false
				}
			} else {
				newPropFilter = append(newPropFilter, filter)
			}
		}
		if !passFilter {
			continue
		}

		scanner, err := GetChannelFileScanner(channel, projectId, periodCode, cloudManager, diskManager, insightGranularity, true)
		if err != nil {
			log.WithError(err).Error("failed getting " + channel + " file scanner for all channel kpi")
			continue
		}

		var newPropsToEval []string
		for _, prop := range propsToEval {
			propWithType := strings.SplitN(prop, "#", 2)
			objType := propWithType[0]
			propName := propWithType[1]
			name, err := getFilterPropertyReportName[channel](propName, objType)
			if err != nil {
				log.WithError(err).Error("error getting property name for channel " + channel + " for all channel kpi")
				continue
			}
			newName := strings.Join([]string{objType, name}, "#")
			newPropsToEval = append(newPropsToEval, newName)
		}
		wpiTmp, err := GetCampaignMetricsInfo(metricNames, channel, scanner, newPropFilter, newPropsToEval)
		if err != nil {
			log.WithError(err).Error("error GetCampaignMetricInfo for all channel kpi for source " + channel)
			continue
		} else {
			wpi.MetricInfo = addMetricInfoStructForSource(channel, wpi.MetricInfo, wpiTmp.MetricInfo)
			// wpi.ScaleInfo = addMetricInfoStruct(wpi.ScaleInfo, wpiTmp.ScaleInfo)
		}
	}
	return &wpi, nil
}

func addMetricInfoStructForSource(source string, baseInfo *MetricInfo, info2add *MetricInfo) *MetricInfo {
	if info2add == nil {
		return baseInfo
	}
	info := *baseInfo
	info.Global += info2add.Global
	info.Features = make(map[string]map[string]float64)
	for key, valMap := range info2add.Features {
		if _, ok := info.Features[key]; !ok {
			info.Features[key] = make(map[string]float64)
		}
		for val, cnt := range valMap {
			info.Features[key][val] += cnt
		}
	}
	if _, ok := info.Features["channel#channel_name"]; !ok {
		info.Features["channel#channel_name"] = make(map[string]float64)
	}
	info.Features["channel#channel_name"][ChannelValueFilterName[source]] = info2add.Global
	return &info
}
