package delta

import (
	"bufio"
	"encoding/json"
	M "factors/model/model"
	"factors/model/store/memsql"
	"factors/pull"
	U "factors/util"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ChannelPropInfo struct {
	Name               string
	DependentKey       string
	DependentValue     float64
	DependentOperation string
	ReplaceValue       map[float64]float64
}

type ChannelMetricCalculationInfo struct {
	Props     []ChannelPropInfo
	Operation string
	Constants map[string]float64
}

var levelStrToIntAlias = map[string]int{
	"organic_property":      3,
	memsql.CAFilterCampaign: 3,
	memsql.CAFilterAdGroup:  2,
	memsql.CAFilterKeyword:  1,
	memsql.CAFilterAd:       1,
}

func getCampaignMetricsInfo(metric string, channel string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*WithinPeriodInsightsKpi, error) {
	var wpi WithinPeriodInsightsKpi
	wpi.MetricInfo = &MetricInfo{}
	wpi.ScaleInfo = &MetricInfo{}

	var newPropFilter []M.KPIFilter
	for _, filter := range propFilter {
		name, err := getFilterPropertyReportName(channel)(filter.PropertyName, filter.ObjectType)
		if err != nil {
			return nil, err
		}
		filter.PropertyName = name
		newPropFilter = append(newPropFilter, filter)
	}

	infoMap := getConstantsInfo(channel)

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

	metricToCalcinfo := getMetricToCalculationinfo(channel)
	docTypeAlias := getDocumentTypeAlias(channel)
	requiredDocTypes := getRequiredDocumentTypes(channel)

	if channel == M.ADWORDS {
		if metricInt, ok := M.AdwordsExtToInternal[metric]; ok {
			metric = metricInt
		}
	}
	var GetCampaignMetric func(scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, queryLevel int, metricCalcInfo ChannelMetricCalculationInfo, docTypeAlias map[string]int, requiredDocTypes []int, infoMap map[string]string) (*MetricInfo, *MetricInfo, error)
	if info, ok := metricToCalcinfo[metric]; ok {
		if info.Operation == "sum" {
			GetCampaignMetric = GetCampaignMetricSimple
		} else {
			GetCampaignMetric = GetCampaignMetricComplex
		}
	} else {
		err := fmt.Errorf("error metric calculation info not available for %s", metric)
		log.WithError(err).Error("error GetCampaignMetricsInfo")
		return nil, err
	}
	if info, scale, err := GetCampaignMetric(scanner, newPropFilter, propsToEvalPerQuery, queryLevel, metricToCalcinfo[metric], docTypeAlias, requiredDocTypes, infoMap); err != nil {
		log.WithError(err).Error("error GetCampaignMetricsInfo for kpi " + metric)
		return nil, err
	} else {
		for key, valMap := range info.Features {
			newKey, _ := getPropertyFilterName(channel)(key)
			delete(info.Features, key)
			info.Features[newKey] = valMap
		}
		wpi.MetricInfo = info
		wpi.ScaleInfo = scale
	}

	return &wpi, nil
}

func GetCampaignMetricSimple(scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, queryLevel int, metricCalcInfo ChannelMetricCalculationInfo, docTypeAlias map[string]int, requiredDocTypes []int, infoMap map[string]string) (*MetricInfo, *MetricInfo, error) {
	var globalVal float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo
	var associatedProps = make(map[int]map[string]map[string]interface{})

	for scanner.Scan() {
		txtline := scanner.Text()

		var campaignDetails pull.CounterCampaignFormat
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

		if val, ok, err := checkDependentProp(metricCalcInfo.Props[0], campaignDetails, extraProps); !ok {
			if err != nil {
				return nil, nil, err
			}
			continue
		} else {
			depVal = val
		}

		if val, ok, err := getPropValueCampaign(metricCalcInfo.Props[0], campaignDetails, extraProps); !ok {
			if err != nil {
				return nil, nil, err
			}
			continue
		} else {
			propVal = val
		}

		//(remove if condition if there is a value other than "product")
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

func GetCampaignMetricComplex(scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, queryLevel int, metricCalcInfo ChannelMetricCalculationInfo, docTypeAlias map[string]int, requiredDocTypes []int, infoMap map[string]string) (*MetricInfo, *MetricInfo, error) {
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

		var campaignDetails pull.CounterCampaignFormat
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
			if val, ok, err := getPropValueCampaign(propInfo, campaignDetails, extraProps); !ok {
				if err != nil {
					return nil, nil, err
				}
				continue
			} else {
				propVal = val
			}

			//(remove if condition if there is a value other than "product")
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

// check if event name is correct and contains all required properties(satisfies constraints)
func isCampaignToBeCounted(campaignDetails pull.CounterCampaignFormat, propFilter []M.KPIFilter, requiredDocTypes []int) (bool, error) {

	allowed := false
	for _, dtype := range requiredDocTypes {
		if campaignDetails.Doctype == dtype {
			allowed = true
			break
		}
	}
	if !allowed {
		return false, nil
	}

	//check if event contains all requiredProps(constraint)
	if ok, err := campaignSatisfiesConstraints(campaignDetails, propFilter); err != nil {
		return false, err
	} else if ok {
		return true, nil
	}

	return false, nil
}

// check if campaign report contains all required properties(satisfies constraints)
func campaignSatisfiesConstraints(campaignDetails pull.CounterCampaignFormat, propFilter []M.KPIFilter) (bool, error) {

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
		propName := filter.PropertyName

		if val, ok := existsInProps(propName, campaignDetails.Value, campaignDetails.SmartProps, "either"); !ok {
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

// check dependent prop
func checkDependentProp(propInfo ChannelPropInfo, campaignDetails pull.CounterCampaignFormat, extraProps map[string]interface{}) (float64, bool, error) {
	var reqVal float64
	key := propInfo.DependentKey
	val := propInfo.DependentValue
	op := propInfo.DependentOperation
	if key != "" {
		if docVal, ok := existsInProps(key, campaignDetails.Value, extraProps, "either"); !ok {
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

// check prop exists and get float value
func getPropValueCampaign(propInfo ChannelPropInfo, campaignDetails pull.CounterCampaignFormat, extraProps map[string]interface{}) (float64, bool, error) {
	if val, ok := existsInProps(propInfo.Name, campaignDetails.Value, extraProps, "either"); ok {
		if floatVal, err := getFloatValueFromInterface(val); err != nil {
			return 0, false, err
		} else {
			return floatVal, true, nil
		}
	} else {
		return 0, false, nil
	}
}

// get smart props + associated props
func getExtraProps(campaignDetails pull.CounterCampaignFormat, level_id string, associatedProps map[int]map[string]map[string]interface{}, docLevel int) (map[string]interface{}, error) {
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
