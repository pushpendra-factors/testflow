package delta

import (
	"bufio"
	"bytes"
	"encoding/json"
	"factors/filestore"
	M "factors/model/model"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func CreateKpiInsights(diskManager *serviceDisk.DiskDriver, cloudManager *filestore.FileManager, periodCodesWithWeekNMinus1 []Period, projectId uint64, queryId uint64, queryGroup M.KPIQueryGroup, insightGranularity string, skipWpi, skipWpi2 bool) error {
	// readEvents := true
	var err error
	var newInsightsList = make([]*WithinPeriodInsightsKpi, 0)
	var oldInsightsList = make([]*WithinPeriodInsightsKpi, 0)

	skipW1 := false
	skipW2 := false
	{
		dateString := U.GetDateOnlyFromTimestampZ(periodCodesWithWeekNMinus1[0].From)
		path, name := (*cloudManager).GetInsightsWpiFilePathAndName(projectId, dateString, queryId, 100)
		if reader, err := (*cloudManager).Get(path, name); err == nil {
			data, err := ioutil.ReadAll(reader)
			if err == nil && skipWpi {
				err := json.Unmarshal(data, &oldInsightsList)
				if err == nil {
					skipW1 = true
				}
			}
		}
		dateString = U.GetDateOnlyFromTimestampZ(periodCodesWithWeekNMinus1[1].From)
		path, name = (*cloudManager).GetInsightsWpiFilePathAndName(projectId, dateString, queryId, 100)
		if reader, err := (*cloudManager).Get(path, name); err == nil {
			data, err := ioutil.ReadAll(reader)
			if err == nil && skipWpi2 {
				err := json.Unmarshal(data, &newInsightsList)
				if err == nil {
					skipW2 = true
				}
			}
		}
	}

	for i, query := range queryGroup.Queries {
		//every query occurs twice so
		if i%2 == 0 {
			continue
		}

		//get global + local constraints
		propFilter := append(queryGroup.GlobalFilters, query.Filters...)

		//get proper props based on category
		var kpiProperties *[]map[string]string
		if query.DisplayCategory == M.WebsiteSessionDisplayCategory {
			kpiProperties = &M.KPIPropertiesForWebsiteSessions
		} else if query.DisplayCategory == M.FormSubmissionsDisplayCategory {
			kpiProperties = &M.KPIPropertiesForFormSubmissions
		} else if query.DisplayCategory == M.PageViewsDisplayCategory {
			kpiProperties = &M.KPIPropertiesForPageViews
		} else {
			log.Errorf("no kpi Insights for category: %s", query.DisplayCategory)
			continue
		}

		//get features for insights as a map
		propsToEval := make([]string, 0)
		for _, propMap := range *kpiProperties {
			if propMap["data_type"] == U.PropertyTypeCategorical {
				var propType string
				if ent, ok := propMap["entity"]; ok && ent == M.UserEntity {
					propType = "up"
				} else { //eventEntity or ["object_type"]
					propType = "ep"
				}
				propName := strings.Join([]string{propType, propMap["name"]}, "#")
				propsToEval = append(propsToEval, propName)
			}
		}

		//get week 2 metrics by reading file
		if !skipW2 {
			if wpi, err := GetMetricsEvaluated(query.DisplayCategory, query.Metrics, query.PageUrl, propFilter, propsToEval, projectId, periodCodesWithWeekNMinus1[1], cloudManager, diskManager, insightGranularity); err != nil {
				return err
			} else {
				newInsightsList = append(newInsightsList, wpi)
			}
		}

		//get week 1 metrics by reading file
		if !skipW1 {
			if wpi, err := GetMetricsEvaluated(query.DisplayCategory, query.Metrics, query.PageUrl, propFilter, propsToEval, projectId, periodCodesWithWeekNMinus1[0], cloudManager, diskManager, insightGranularity); err != nil {
				return err
			} else {
				oldInsightsList = append(oldInsightsList, wpi)
			}
		}
	}
	if !skipW2 {
		wpiBytes, err := json.Marshal(newInsightsList)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("failed to marshal wpi2 Info.")
			return err
		}

		err = WriteWpiPath(projectId, Period(periodCodesWithWeekNMinus1[1]), queryId, 100, bytes.NewReader(wpiBytes), *cloudManager)
		if err != nil {
			log.WithError(err).Error("write WPI error - ", err)
			return err
		}
	}
	if !skipW1 {
		wpiBytes, err := json.Marshal(oldInsightsList)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("failed to marshal wpi1 Info.")
			return err
		}
		err = WriteWpiPath(projectId, Period(periodCodesWithWeekNMinus1[0]), queryId, 100, bytes.NewReader(wpiBytes), *cloudManager)
		if err != nil {
			log.WithError(err).Error("write WPI error - ", err)
			return err
		}
	}

	//get insights between the weeks
	var crossPeriodInsightsList []*CrossPeriodInsightsKpi
	periodPair := PeriodPair{First: periodCodesWithWeekNMinus1[0], Second: periodCodesWithWeekNMinus1[1]}
	if len(newInsightsList) > 0 && len(oldInsightsList) > 0 {
		crossPeriodInsightsList, err = ComputeCrossPeriodKpiInsights(periodPair, newInsightsList, oldInsightsList)
		if err != nil {
			log.WithError(err).Error("compute cpi for kpi error - ", err)
			return err
		}
	}

	//Create cpi file with insights
	if len(crossPeriodInsightsList) > 0 {
		crossPeriodInsightsBytes, err := json.Marshal(crossPeriodInsightsList)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("failed to unmarshal cpi Info.")
			return err
		}
		err = WriteCpiPath(projectId, periodPair.Second, uint64(queryId), 100, bytes.NewReader(crossPeriodInsightsBytes), *cloudManager)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("failed to write cpi files to cloud")
			return err
		}
	}
	return nil
}

func GetMetricsEvaluated(category string, metricNames []string, queryEvent string, propFilter []M.KPIFilter, propsToEval []string, projectId uint64, periodCode Period, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, insightGranularity string) (*WithinPeriodInsightsKpi, error) {

	var insights *WithinPeriodInsightsKpi
	var scanner *bufio.Scanner
	var err error
	if scanner, err = GetEventFileScanner(projectId, periodCode, cloudManager, diskManager, insightGranularity, true); err != nil {
		log.WithError(err).Error("failed getting event file scanner")
		return nil, err
	}
	var GetMetrics func(metricNames []string, queryEvent string, scanner *bufio.Scanner, propFilter []M.KPIFilter, propsToEval []string) (*WithinPeriodInsightsKpi, error)
	if category == M.WebsiteSessionDisplayCategory {
		GetMetrics = GetSessionMetrics
	} else if category == M.FormSubmissionsDisplayCategory {
		GetMetrics = GetFormSubmitMetrics
	} else if category == M.PageViewsDisplayCategory {
		GetMetrics = GetPageViewMetrics
	} else {
		err := fmt.Errorf("no kpi Insights for category: %s", category)
		log.WithError(err).Error("not computing insights for this category")
		return insights, err
	}
	insights, err = GetMetrics(metricNames, queryEvent, scanner, propFilter, propsToEval)
	return insights, err
}

//check if prop exists in user or event props
func ExistsInProps(prop string, eventDetails P.CounterEventFormat, entity string) (interface{}, bool) {
	if entity == "ep" || entity == "both" {
		if val, ok := eventDetails.EventProperties[prop]; ok {
			return val, true
		}
	} else if entity == "up" || entity == "both" {
		if val, ok := eventDetails.UserProperties[prop]; ok {
			return val, true
		}
	}
	return nil, false
}

//check if event contains all required properties(constraints)
func eventSatisfiesConstraints(eventDetails P.CounterEventFormat, propFilter []M.KPIFilter) (bool, error) {

	for _, filter := range propFilter {
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

		if val, ok := ExistsInProps(propName, eventDetails, propType); !ok {
			return false, nil
		} else {
			eventVal = val
		}

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
				return false, fmt.Errorf("")
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
				return false, fmt.Errorf("")
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
			return false, fmt.Errorf("property type unknown for %s", propName)
		} else {
			return false, fmt.Errorf("strange property type: %s", filter.PropertyDataType)
		}
	}
	return true, nil
}

//compute cross period using within period infos
func ComputeCrossPeriodKpiInsights(periodPair PeriodPair, newInsightsList, oldInsightsList []*WithinPeriodInsightsKpi) ([]*CrossPeriodInsightsKpi, error) {
	crossPeriodInsightsList := make([]*CrossPeriodInsightsKpi, 0)
	for i := range newInsightsList {
		var crossPeriodInsights CpiMetricInfo
		newInsights := newInsightsList[i]
		oldInsights := *oldInsightsList[i]
		oldInfo := oldInsights.MetricInfo
		newInfo := newInsights.MetricInfo

		//get union of props
		var allProps = make(map[string]map[string]bool)
		for key, valMap := range oldInfo.Features {
			allProps[key] = make(map[string]bool)
			for val := range valMap {
				allProps[key][val] = true
			}
		}
		for key, valMap := range newInfo.Features {
			if _, ok := allProps[key]; !ok {
				allProps[key] = make(map[string]bool)
			}
			for val := range valMap {
				allProps[key][val] = true
			}
		}

		//global
		first := oldInfo.Global
		second := newInfo.Global
		var percentChange, factor float64
		if first != 0 {
			percentChange = 100 * float64(second-first) / float64(first)
			factor = float64(second) / float64(first)
		} else {
			percentChange = 100
			factor = float64(second)
		}
		crossPeriodInsights.GlobalMetrics = DiffMetric{First: first, Second: second, PercentChange: percentChange, FactorChange: factor}

		//features
		crossPeriodInsights.FeatureMetrics = make(map[string]map[string]DiffMetric)
		for key, valMap := range allProps {
			if _, ok := newInfo.Features[key]; !ok {
				newInfo.Features[key] = make(map[string]float64)
			}
			if _, ok := oldInfo.Features[key]; !ok {
				oldInfo.Features[key] = make(map[string]float64)
			}
			for val := range valMap {
				first := oldInfo.Features[key][val]
				second := newInfo.Features[key][val]
				var percentChange, factor float64
				if first != 0 {
					percentChange = 100 * (second - first) / first
					factor = second / first
				} else {
					percentChange = 100
					factor = second
				}
				if _, ok := crossPeriodInsights.FeatureMetrics[key]; !ok {
					crossPeriodInsights.FeatureMetrics[key] = make(map[string]DiffMetric)
				}
				crossPeriodInsights.FeatureMetrics[key][val] = DiffMetric{First: first, Second: second, PercentChange: percentChange, FactorChange: factor}
			}
		}

		var scaleInfo CpiMetricInfo
		oldScale := oldInsights.ScaleInfo
		newScale := newInsights.ScaleInfo

		//global
		first = oldScale.Global
		second = newScale.Global
		if first != 0 {
			percentChange = 100 * float64(second-first) / float64(first)
			factor = float64(second) / float64(first)
		} else {
			percentChange = 100
			factor = float64(second)
		}
		scaleInfo.GlobalMetrics = DiffMetric{First: first, Second: second, PercentChange: percentChange, FactorChange: factor}

		//features
		scaleInfo.FeatureMetrics = make(map[string]map[string]DiffMetric)
		for key, valMap := range allProps {
			if _, ok := newScale.Features[key]; !ok {
				newScale.Features[key] = make(map[string]float64)
			}
			if _, ok := oldScale.Features[key]; !ok {
				oldScale.Features[key] = make(map[string]float64)
			}
			for val := range valMap {
				first := oldScale.Features[key][val]
				second := newScale.Features[key][val]
				var percentChange, factor float64
				if first != 0 {
					percentChange = 100 * (second - first) / first
					factor = second / first
				} else {
					percentChange = 100
					factor = second
				}
				if _, ok := scaleInfo.FeatureMetrics[key]; !ok {
					scaleInfo.FeatureMetrics[key] = make(map[string]DiffMetric)
				}
				scaleInfo.FeatureMetrics[key][val] = DiffMetric{First: first, Second: second, PercentChange: percentChange, FactorChange: factor}
			}
		}
		var cpiInsightsKpi CrossPeriodInsightsKpi
		cpiInsightsKpi.Periods = periodPair
		cpiInsightsKpi.Target = &crossPeriodInsights
		cpiInsightsKpi.BaseAndTarget = &crossPeriodInsights
		cpiInsightsKpi.ScaleInfo = &scaleInfo
		// cpiInsightsKpi.JSDivergence = JSDType{Target: MultipleJSDivergenceKpi(oldInfo, newInfo, allProps)}
		crossPeriodInsightsList = append(crossPeriodInsightsList, &cpiInsightsKpi)
	}
	return crossPeriodInsightsList, nil
}

func MultipleJSDivergenceKpi(metricInfo1, metricInfo2 *MetricInfo, allProps map[string]map[string]bool) Level2CatRatioDist {
	jsdMetrics := make(Level2CatRatioDist)
	for key, valMap := range metricInfo1.Features {
		jsdMetrics[key] = make(Level1CatRatioDist)
		for val, _ := range valMap {
			prev1 := metricInfo1.Features[key][val] / metricInfo1.Global
			prev2 := metricInfo2.Features[key][val] / metricInfo2.Global
			jsd := SingleJSDivergence(prev1, prev2)
			jsdMetrics[key][val] = jsd
		}
	}
	return jsdMetrics
}

func addToScale(globalScale *float64, scaleMap map[string]map[string]float64, propsToEval []string, eventDetails P.CounterEventFormat) {
	(*globalScale)++
	for _, propWithType := range propsToEval {
		propTypeName := strings.SplitN(propWithType, "#", 2)
		prop := propTypeName[1]
		propType := propTypeName[0]
		if val, ok := ExistsInProps(prop, eventDetails, propType); ok {
			val := fmt.Sprintf("%s", val)
			if _, ok := scaleMap[propWithType]; !ok {
				scaleMap[propWithType] = make(map[string]float64)
			}
			scaleMap[propWithType][val] += 1
		}
	}
}
