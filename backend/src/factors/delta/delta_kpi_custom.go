package delta

import (
	"bufio"
	"encoding/json"
	"factors/filestore"
	M "factors/model/model"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type CounterUserFormat struct {
	Id            string
	Properties    map[string]interface{}
	Is_Anonymous  bool
	JoinTimestamp int64
}

func GetCustomMetricsInfo(metric string, propFilter []M.KPIFilter, propsToEval []string, projectId int64, periodCode Period, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, insightGranularity string) (*WithinPeriodInsightsKpi, error) {
	var wpi WithinPeriodInsightsKpi
	var transformation M.CustomMetricTransformation
	customMetric, errStr, getStatus := store.GetStore().GetKpiRelatedCustomMetricsByName(projectId, metric)
	if getStatus != http.StatusFound {
		log.WithField("error", errStr).Error("Get custom metrics failed.")
		return nil, fmt.Errorf("%s", errStr)
	}
	err1 := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &transformation)
	if err1 != nil {
		log.WithField("customMetric", customMetric).WithField("err", err1).Warn("Failed in decoding custom Metric")
	}

	newPropFilter := append(propFilter, transformation.Filters...)

	scanner, err := GetUserFileScanner(transformation.DateField, projectId, periodCode, cloudManager, diskManager, insightGranularity, true)
	if err != nil {
		log.WithError(err).Error("failed getting " + transformation.DateField + " file scanner for custom kpi")
		return nil, err
	}

	var GetCustomMetric func(metric string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, metricFunc string, metricProp string) (*MetricInfo, *MetricInfo, error)
	if transformation.AggregateFunction == M.AverageAggregateFunction {
		GetCustomMetric = GetCustomMetricsComplex
	} else {
		GetCustomMetric = GetCustomMetricsSimple
	}
	if info, scale, err := GetCustomMetric(metric, scanner, newPropFilter, propsToEval, transformation.AggregateFunction, transformation.AggregateProperty); err != nil {
		log.WithError(err).Error("error GetCustomMetric for kpi " + metric)
		return nil, err
	} else {
		wpi.MetricInfo = info
		wpi.ScaleInfo = scale
	}
	return &wpi, nil
}

func GetCustomMetricsSimple(metric string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, metricFunc string, metricProp string) (*MetricInfo, *MetricInfo, error) {
	var globalVal float64
	var globalScale float64
	var reqMap = make(map[string]map[string]float64)
	var scaleMap = make(map[string]map[string]float64)
	var info MetricInfo
	var scale MetricInfo

	for scanner.Scan() {
		txtline := scanner.Text()

		var userDetails CounterUserFormat
		if err := json.Unmarshal([]byte(txtline), &userDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}

		//check filters
		if ok, err := isUserToBeCounted(userDetails, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScaleUser(&globalScale, scaleMap, propsToEval, userDetails)

		if metricFunc == M.UniqueAggregateFunction {
			addValueToMapForPropsPresent(&globalVal, reqMap, 1, propsToEval, userDetails.Properties, userDetails.Properties)
		} else if metricFunc == M.SumAggregateFunction {
			var propVal float64
			if val, ok := ExistsInProps(metricProp, userDetails.Properties, userDetails.Properties, "up"); ok {
				propVal = val.(float64)
			} else {
				continue
			}
			addValueToMapForPropsPresent(&globalVal, reqMap, propVal, propsToEval, userDetails.Properties, userDetails.Properties)
		}
	}

	deleteEntriesWithZeroFreq(reqMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: globalVal, Features: reqMap}
	return &info, &scale, nil
}

func GetCustomMetricsComplex(metric string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string, metricFunc string, metricProp string) (*MetricInfo, *MetricInfo, error) {
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

		var userDetails CounterUserFormat
		if err := json.Unmarshal([]byte(txtline), &userDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}

		//check filters
		if ok, err := isUserToBeCounted(userDetails, propFilter); !ok {
			if err != nil {
				return &MetricInfo{}, &MetricInfo{}, err
			}
			continue
		}
		addToScaleUser(&globalScale, scaleMap, propsToEval, userDetails)

		var propVal float64
		if val, ok := ExistsInProps(metricProp, userDetails.Properties, userDetails.Properties, "up"); ok {
			propVal = val.(float64)
		} else {
			continue
		}
		addValuesToFractionForPropsPresent(&globalFrac, featInfoMap, propVal, 1, propsToEval, userDetails.Properties, userDetails.Properties)
	}

	globalVal, reqMap = getFractionValue(&globalFrac, featInfoMap)
	scale = MetricInfo{Global: globalScale, Features: scaleMap}
	info = MetricInfo{Global: globalVal, Features: reqMap}

	return &info, &scale, nil
}

func getPropKeysToEvalForCustomMetric(metric string, projectId int64, periodCodes []Period, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, topK int, insightGranularity string) (map[string]bool, error) {

	var finalProps = make(map[string]bool)
	var datefield string
	{
		var transformation M.CustomMetricTransformation
		customMetric, errStr, getStatus := store.GetStore().GetKpiRelatedCustomMetricsByName(projectId, metric)
		if getStatus != http.StatusFound {
			log.WithField("error", errStr).Error("Get custom metrics failed.")
			return nil, fmt.Errorf("%s", errStr)
		}
		err1 := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &transformation)
		if err1 != nil {
			log.WithField("customMetric", customMetric).WithField("err", err1).Warn("Failed in decoding custom Metric")
			return nil, err1
		}
		datefield = transformation.DateField
	}

	err := addTopKPropKeys(finalProps, datefield, projectId, periodCodes[1], cloudManager, diskManager, topK, insightGranularity)
	if err != nil {
		log.WithField("err", err).Error("Failed in getting topk keys")
		return nil, err
	}

	err = addTopKPropKeys(finalProps, datefield, projectId, periodCodes[0], cloudManager, diskManager, topK, insightGranularity)
	if err != nil {
		log.WithField("err", err).Error("Failed in getting topk keys")
		return nil, err
	}
	return finalProps, nil

}

// top K keys meaning unique keys from top K (key,value) pairs
func addTopKPropKeys(finalProps map[string]bool, datefield string, projectId int64, periodCode Period, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, topK int, insightGranularity string) error {
	var propsPerWeek = make(Level3CatRatioDist)
	{
		scanner, err := GetUserFileScanner(datefield, projectId, periodCode, cloudManager, diskManager, insightGranularity, true)
		if err != nil {
			log.WithError(err).Error("failed getting " + datefield + " file scanner for custom kpi")
			return err
		}
		countsMap, err := getCountsMapFromUserScanner(scanner)
		if err != nil {
			log.WithError(err).Error("failed getting countsMap for" + datefield + " file scanner")
			return err
		}

		//get counts map in proper format to use in functions ahead
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

func getCountsMapFromUserScanner(scanner *bufio.Scanner) (map[string]map[string]float64, error) {

	var countsMap = make(map[string]map[string]float64)
	for scanner.Scan() {
		txtline := scanner.Text()

		var userDetails CounterUserFormat
		if err := json.Unmarshal([]byte(txtline), &userDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, err
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

func getFilteredKpiPropertiesForCustomMetric(kpiProperties []map[string]string, metric string, projectId int64, periodCodes []Period, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, topK int, insightGranularity string) ([]map[string]string, error) {
	filteredKpiProperties := make([]map[string]string, 0)
	propKeys, err := getPropKeysToEvalForCustomMetric(metric, projectId, periodCodes, cloudManager, diskManager, topK, insightGranularity)
	if err != nil {
		err := fmt.Errorf("error getting topK keys from 1st scan")
		log.WithError(err).Error("error getPropKeysToEvalForCustomMetric")
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
