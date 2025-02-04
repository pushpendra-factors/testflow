package delta

import (
	"bufio"
	"encoding/json"
	"factors/filestore"
	"factors/merge"
	M "factors/model/model"
	"factors/model/store"
	"factors/pull"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// get within period insights for a week for custom kpi
func getCustomMetricsInfo(metric string, propFilter []M.KPIFilter, propsToEval []string, projectId int64, periodCode Period, archiveCloudManager, tmpCloudManager, sortedCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, beamConfig *merge.RunBeamConfig, hardPull, useSortedFilesMerge bool, pulledMap map[int64]map[string]bool) (*WithinPeriodInsightsKpi, error) {
	var wpi WithinPeriodInsightsKpi
	wpi.MetricInfo = &MetricInfo{}
	wpi.ScaleInfo = &MetricInfo{}
	var transformation M.CustomMetricTransformation

	//get custom metric details from db
	customMetric, errStr, getStatus := store.GetStore().GetProfileCustomMetricByProjectIdName(projectId, metric)
	if getStatus != http.StatusFound {
		log.WithField("error", errStr).Error("Get custom metrics failed. Not a profile custom metric.")
		return &wpi, fmt.Errorf("%s", errStr)
	}
	err1 := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &transformation)
	if err1 != nil {
		log.WithField("customMetric", customMetric).WithField("err", err1).Warn("Failed in decoding custom Metric")
	}

	//add custom metric filters to propFilter
	newPropFilter := append(propFilter, transformation.Filters...)

	//get file scanner
	scanner, err := GetUserFileScanner(transformation.DateField, projectId, periodCode, archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, hardPull, useSortedFilesMerge, pulledMap)
	if err != nil {
		log.WithError(err).Error("failed getting " + transformation.DateField + " file scanner for custom kpi")
		return &wpi, err
	}

	//get proper function (complex for avg, simple for unique,sum)
	var GetCustomMetric func(scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, metricFunc string, metricProp string, dateField string, startTimestamp, endTimestamp int64) (*MetricInfo, *MetricInfo, error)
	if transformation.AggregateFunction == M.AverageAggregateFunction {
		GetCustomMetric = getCustomMetricsComplex
	} else {
		GetCustomMetric = getCustomMetricsSimple
	}
	if info, scale, err := GetCustomMetric(scanner, newPropFilter, propsToEval, transformation.AggregateFunction, transformation.AggregateProperty, transformation.DateField, periodCode.From, periodCode.To); err != nil {
		log.WithError(err).Error("error GetCustomMetric for kpi " + metric)
		return &wpi, err
	} else {
		wpi.MetricInfo = info
		wpi.ScaleInfo = scale
	}

	return &wpi, nil
}

// get custom kpi values for non-fraction type kpi
func getCustomMetricsSimple(scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, metricFunc string, metricProp string, dateField string, startTimestamp, endTimestamp int64) (*MetricInfo, *MetricInfo, error) {
	var globalVal float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var userDetails pull.CounterUserFormat
		if err := json.Unmarshal([]byte(txtline), &userDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		timestampFloat, _ := U.GetPropertyValueAsFloat64(userDetails.Properties[dateField])
		timestamp := int64(U.GetTimestampInSecs(int(timestampFloat)))
		if !U.IsTimestampInRange(timestamp, startTimestamp, endTimestamp, false) {
			log.WithFields(log.Fields{"line": txtline}).Error("timestamp outside of required range")
			continue
		}

		//check filters
		if ok, err := isUserToBeCounted(userDetails, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScaleUser(&globalScale, scaleMap, propsToEval, userDetails)

		if metricFunc == M.CountAggregateFunction || metricFunc == M.UniqueAggregateFunction {
			addValueToMapForPropsPresent(&globalVal, reqMap, 1, propsToEval, userDetails.Properties, userDetails.Properties)
		} else if metricFunc == M.SumAggregateFunction {
			if val, ok := existsInProps(metricProp, userDetails.Properties, userDetails.Properties, "up"); ok {
				propVal, _ := U.GetFloatValueFromInterface(val)
				addValueToMapForPropsPresent(&globalVal, reqMap, propVal, propsToEval, userDetails.Properties, userDetails.Properties)
			}
		}
	}

	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: globalVal, Features: reqMap}
	return &info, &scale, nil
}

// get custom kpi values for fraction type kpi
func getCustomMetricsComplex(scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, metricFunc string, metricProp string, dateField string, startTimestamp, endTimestamp int64) (*MetricInfo, *MetricInfo, error) {
	var globalVal float64
	var globalFrac Fraction
	var globalScale float64
	var featInfoMap = make(map[string]map[string]Fraction)
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var userDetails pull.CounterUserFormat
		if err := json.Unmarshal([]byte(txtline), &userDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		timestampFloat, _ := U.GetPropertyValueAsFloat64(userDetails.Properties[dateField])
		timestamp := int64(U.GetTimestampInSecs(int(timestampFloat)))
		if !U.IsTimestampInRange(timestamp, startTimestamp, endTimestamp, false) {
			log.WithFields(log.Fields{"line": txtline}).Error("timestamp outside of required range")
			continue
		}

		//check filters
		if ok, err := isUserToBeCounted(userDetails, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScaleUser(&globalScale, scaleMap, propsToEval, userDetails)

		if val, ok := existsInProps(metricProp, userDetails.Properties, userDetails.Properties, "up"); ok {
			propVal, _ := U.GetFloatValueFromInterface(val)
			addValuesToFractionForPropsPresent(&globalFrac, featInfoMap, propVal, 1, propsToEval, userDetails.Properties, userDetails.Properties)
		}
	}

	globalVal, reqMap = getFractionValue(&globalFrac, featInfoMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: globalVal, Features: reqMap}

	return &info, &scale, nil
}

// check user properties contains all required properties(satisfies constraints)
func isUserToBeCounted(userDetails pull.CounterUserFormat, propFilter []M.KPIFilter) (bool, error) {

	//check if event contains all requiredProps(constraint)
	if ok, err := userSatisfiesConstraints(userDetails, propFilter); err != nil {
		return false, err
	} else if ok {
		return true, nil
	}

	return false, nil
}

// check if user contains all required properties(satisfies constraints)
func userSatisfiesConstraints(userDetails pull.CounterUserFormat, propFilter []M.KPIFilter) (bool, error) {

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

		if val, ok := existsInProps(propName, userDetails.Properties, nil, "either"); !ok {
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

// add 1 to globalScale and to scaleMap for all values from propsToEval properties found in userDetails
func addToScaleUser(globalScale *float64, scaleMap map[string]map[string]float64, propsToEval []string, userDetails pull.CounterUserFormat) {
	(*globalScale)++
	for _, propWithType := range propsToEval {
		propTypeName := strings.SplitN(propWithType, "#", 2)
		prop := propTypeName[1]
		propType := propTypeName[0]
		if val, ok := existsInProps(prop, userDetails.Properties, userDetails.Properties, propType); ok {
			val := fmt.Sprintf("%s", val)
			if _, ok := scaleMap[propWithType]; !ok {
				scaleMap[propWithType] = make(map[string]float64)
			}
			scaleMap[propWithType][val] += 1
		}
	}
}

// get union of topK properties from both files
func getPropKeysToEvalForCustomMetric(metric string, projectId int64, periodCodes []Period, archiveCloudManager, tmpCloudManager, sortedCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, topK int, beamConfig *merge.RunBeamConfig, hardPull, useSortedFilesMerge bool, pulledMap map[int64]map[string]bool) (map[string]bool, error) {

	var finalProps = make(map[string]bool)

	//get datefield of custom metric to get the name of associated user file
	var datefield string
	{
		var transformation M.CustomMetricTransformation
		customMetric, errStr, getStatus := store.GetStore().GetProfileCustomMetricByProjectIdName(projectId, metric)
		if getStatus != http.StatusFound {
			log.WithField("error", errStr).Error("getPropKeysToEvalForCustomMetric failed. Not a profile custom metric.")
			return nil, fmt.Errorf("%s", errStr)
		}
		err1 := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &transformation)
		if err1 != nil {
			log.WithField("customMetric", customMetric).WithField("err", err1).Warn("Failed in decoding custom Metric")
			return nil, err1
		}
		datefield = transformation.DateField
	}

	//add topK props from second week
	err := addTopKPropKeys(finalProps, datefield, projectId, periodCodes[1], archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, topK, beamConfig, hardPull, useSortedFilesMerge, pulledMap)
	if err != nil {
		log.WithField("err", err).Error("Failed in getting topk keys")
		return nil, err
	}

	//add topK props from first week
	err = addTopKPropKeys(finalProps, datefield, projectId, periodCodes[0], archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, topK, beamConfig, hardPull, useSortedFilesMerge, pulledMap)
	if err != nil {
		log.WithField("err", err).Error("Failed in getting topk keys")
		return nil, err
	}
	return finalProps, nil
}

// get user file and get topK keys(top K keys meaning unique keys from top K [key,value] pairs)
func addTopKPropKeys(finalProps map[string]bool, datefield string, projectId int64, periodCode Period, archiveCloudManager, tmpCloudManager, sortedCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, topK int, beamConfig *merge.RunBeamConfig, hardPull, useSortedFilesMerge bool, pulledMap map[int64]map[string]bool) error {
	//get counts map in proper format to use in functions built for events wi
	var propsPerWeek = make(Level3CatRatioDist)
	{
		scanner, err := GetUserFileScanner(datefield, projectId, periodCode, archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, hardPull, useSortedFilesMerge, pulledMap)
		if err != nil {
			log.WithError(err).Error("failed getting " + datefield + " file scanner for custom kpi")
			return err
		}
		countsMap, err := getCountsMapFromUserScanner(scanner, periodCode.From, periodCode.To, datefield)
		if err != nil {
			log.WithError(err).Error("failed getting countsMap for" + datefield + " file scanner")
			return err
		}

		for k, valMap := range countsMap {
			for val, cnt := range valMap {
				key := "t#" + k
				if _, ok := propsPerWeek[key]; !ok {
					propsPerWeek[key] = make(Level2CatRatioDist)
				}
				if _, ok := propsPerWeek[k][val]; !ok {
					propsPerWeek[key][val] = make(Level1CatRatioDist)
				}
				propsPerWeek[key][val]["#users"] = cnt
			}
		}
	}
	var wpi WithinPeriodInsights
	wpi.Target.FeatureMetrics = propsPerWeek

	PrefilterFeatures(&wpi)
	selectTopKFeatures(&(wpi), topK)

	//add to finalProps (set false in first run and true in second)
	for k := range wpi.Target.FeatureMetrics {
		tmpKey := strings.SplitN(k, "#", 2)
		key := tmpKey[1]
		if _, ok := finalProps[key]; !ok {
			finalProps[key] = false
		} else {
			finalProps[key] = true
		}
	}

	return nil
}

// get map of counts of prop values (map[prop][value] = count) from user file scanner
func getCountsMapFromUserScanner(scanner *bufio.Scanner, startTimestamp, endTimestamp int64, dateField string) (map[string]map[string]float64, error) {

	var countsMap = make(map[string]map[string]float64)
	for scanner.Scan() {
		txtline := scanner.Text()

		var userDetails pull.CounterUserFormat
		if err := json.Unmarshal([]byte(txtline), &userDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, err
		}
		timestampFloat, _ := U.GetPropertyValueAsFloat64(userDetails.Properties[dateField])
		timestamp := int64(timestampFloat)
		if timestamp < startTimestamp || timestamp > endTimestamp {
			continue
		}

		for k, v := range userDetails.Properties {
			if _, ok := countsMap[k]; !ok {
				countsMap[k] = make(map[string]float64)
			}
			val := fmt.Sprintf("%s", v)
			countsMap[k][val] += 1
		}
	}
	return countsMap, nil
}

// get union of topk prop keys from both files and filter through kpiProperties
func getFilteredKpiPropertiesForCustomMetric(kpiProperties []map[string]string, metric string, projectId int64, periodCodes []Period, archiveCloudManager, tmpCloudManager, sortedCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, topK int, beamConfig *merge.RunBeamConfig, hardPull, useSortedFilesMerge bool, pulledMap map[int64]map[string]bool) ([]map[string]string, error) {
	filteredKpiProperties := make([]map[string]string, 0)
	propKeys, err := getPropKeysToEvalForCustomMetric(metric, projectId, periodCodes, archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, topK, beamConfig, hardPull, useSortedFilesMerge, pulledMap)
	if err != nil {
		log.WithError(err).Error("error getPropKeysToEvalForCustomMetric: error getting topK keys from 1st scan")
		return nil, err
	}
	for _, propMap := range kpiProperties {
		//check for true if taking intersection of both week property keys; check ok for union
		if _, ok := propKeys[propMap["name"]]; ok {
			filteredKpiProperties = append(filteredKpiProperties, propMap)
		}
	}

	return filteredKpiProperties, nil
}
