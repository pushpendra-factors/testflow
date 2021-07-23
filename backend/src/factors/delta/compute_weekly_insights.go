package delta

import (
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type WeeklyInsights struct {
	InsightsType string          `json:"insights_type"`
	BaseLine     string          `json:"baseline"`
	Base         Base            `json:"base"`
	Goal         Base            `json:"goal"`
	Conv         Base            `json:"conv"`
	Insights     []ActualMetrics `json:"actual_metrics"`
}
type Base struct {
	W1          float64 `json:"w1"`
	W2          float64 `json:"w2"`
	IsIncreased bool    `json:"isIncrease"`
	Percentage  float64 `json:"percentage"`
}

type ActualMetrics struct {
	Key                  string `json:"key"`
	Value                string `json:"value"`
	Entity               string `json:"entity"`
	ActualValues         Base   `json:"actual_values"`
	ChangeInConversion   Base   `json:"change_in_conversion"`
	ChangeInPrevalance   Base   `json:"change_in_prevalance"`
	ChangeInDistribution Base   `json:"change_in_distribution"`
	Type                 string `json:"type"`
}

// temporary structure to hold values
type ValueWithDetails struct {
	Key                  string            `json:"key"`
	Value                string            `json:"value"`
	Entity               string            `json:"entity"`
	ActualValues         BaseTargetMetrics `json:"actual_values"`
	ChangeInConversion   Base              `json:"change_in_conversion"`
	ChangeInPrevalance   Base              `json:"change_in_prevalance"`
	ChangeInDistribution Base              `json:"change_in_distribution"`
	Type                 string            `json:"type"`
}
type BaseTargetMetrics struct {
	W1           float64 `json:"w1"`
	W2           float64 `json:"w2"`
	Per          float64 `json:"per"`
	DeltaRatio   float64 `json:"delrat"`
	JSDivergence float64 `json:"jsd"`
}

var numberOfRecordsFromGbp int = 5 // number of records to be fetched from gbp
var increasedRecords int
var decreasedRecords int
var propertyMap map[string]bool

var WebsiteEvent string = "WebsiteEvent"
var CRM string = "CRM"
var Funnel string = "Funnel"

func GetInsights(file CrossPeriodInsights, numberOfRecords int, QueryClass, EventType string) WeeklyInsights {
	var KeyMapForConversion = make(map[string]bool)
	var KeyMapForDistribution = make(map[string]bool)
	propertyMap = make(map[string]bool)
	var insights WeeklyInsights
	insights.Insights = make([]ActualMetrics, 0)
	if EventType == Funnel || EventType == WebsiteEvent {
		insights.InsightsType = "ConvAndDist"
		if EventType == WebsiteEvent {
			insights.BaseLine = "$session"
		}
	} else if EventType == CRM {
		insights.InsightsType = "DistOnly"
	}
	if EventType == Funnel || EventType == WebsiteEvent { // change the values
		if _, exists := file.Base.GlobalMetrics["#users"]; exists {

			if file.Base.GlobalMetrics["#users"].First != nil {
				insights.Base.W1 = file.Base.GlobalMetrics["#users"].First.(float64)
			} else {
				insights.Base.W1 = 0
			}
			if file.Base.GlobalMetrics["#users"].Second != nil {
				insights.Base.W2 = file.Base.GlobalMetrics["#users"].Second.(float64)
			} else {
				insights.Base.W2 = 0
			}
			insights.Base.IsIncreased = file.Base.GlobalMetrics["#users"].PercentChange > 0
			insights.Base.Percentage = file.Base.GlobalMetrics["#users"].PercentChange
		}
	}
	// pulling  goal according to type of event
	if EventType == Funnel || EventType == WebsiteEvent { // change the values here
		if _, exists := file.BaseAndTarget.GlobalMetrics["#users"]; exists {
			if file.BaseAndTarget.GlobalMetrics["#users"].First != nil {
				insights.Goal.W1 = file.BaseAndTarget.GlobalMetrics["#users"].First.(float64)
			} else {
				insights.Goal.W1 = 0
			}
			if file.BaseAndTarget.GlobalMetrics["#users"].Second != nil {
				insights.Goal.W2 = file.BaseAndTarget.GlobalMetrics["#users"].Second.(float64)
			} else {
				insights.Goal.W2 = 0
			}
			insights.Goal.IsIncreased = file.BaseAndTarget.GlobalMetrics["#users"].PercentChange > 0
			insights.Goal.Percentage = file.BaseAndTarget.GlobalMetrics["#users"].PercentChange
		}
	} else if EventType == CRM { //  pulling from target for crm type event
		if _, exists := file.Target.GlobalMetrics["#users"]; exists {
			if file.Target.GlobalMetrics["#users"].First != nil {
				insights.Goal.W1 = file.Target.GlobalMetrics["#users"].First.(float64)
			} else {
				insights.Goal.W1 = 0
			}
			if file.Target.GlobalMetrics["#users"].Second != nil {
				insights.Goal.W2 = file.Target.GlobalMetrics["#users"].Second.(float64)
			} else {
				insights.Goal.W2 = 0
			}
			insights.Goal.IsIncreased = file.Target.GlobalMetrics["#users"].PercentChange > 0
			insights.Goal.Percentage = file.Target.GlobalMetrics["#users"].PercentChange
		}
	}
	if EventType == Funnel || EventType == WebsiteEvent {

		insights.Conv.W1 = insights.Goal.W1 / insights.Base.W1 * 100
		insights.Conv.W2 = insights.Goal.W2 / insights.Base.W2 * 100
		insights.Conv.IsIncreased = insights.Conv.W1 < insights.Conv.W2
		insights.Conv.Percentage = ((insights.Conv.W2 - insights.Conv.W1) / insights.Conv.W1) * 100

	}
	var valWithDetailsArr []ValueWithDetails
	// for conversion
	if EventType == Funnel || EventType == WebsiteEvent {
		for keys := range file.BaseAndTarget.FeatureMetrics {
			// filtering keys prefixed with s#
			if strings.HasPrefix(keys, "s#") {
				var value ValueWithDetails
				var temp BaseTargetMetrics
				for keys2 := range file.BaseAndTarget.FeatureMetrics[keys] {
					if keys2 == "" { // omitting "" values
						continue
					}
					if KeyMapForConversion[keys2] { // deduping the results
						continue
					}
					KeyMapForConversion[keys2] = true
					value.Key = keys[5:]
					value.Value = keys2
					value.Entity = keys[2:4]
					if file.BaseAndTarget.FeatureMetrics[keys][keys2].First != nil {
						temp.W1 = file.BaseAndTarget.FeatureMetrics[keys][keys2].First.(float64)
					} else {
						temp.W1 = 0
					}
					if file.BaseAndTarget.FeatureMetrics[keys][keys2].Second != nil {
						temp.W2 = file.BaseAndTarget.FeatureMetrics[keys][keys2].Second.(float64)
					} else {
						temp.W2 = 0
					}
					if _, exists := file.BaseAndTarget.FeatureMetrics[keys][keys2]; exists {
						temp.Per = file.BaseAndTarget.FeatureMetrics[keys][keys2].PercentChange
					}
					if _, exists := file.DeltaRatio[keys][keys2]; exists {
						temp.DeltaRatio = file.DeltaRatio[keys][keys2]
					}

					value.ActualValues = temp

					if file.Base.FeatureMetrics[keys][keys2].First != nil {
						value.ChangeInPrevalance.W1 = file.Base.FeatureMetrics[keys][keys2].First.(float64)
					} else {
						value.ChangeInPrevalance.W1 = 0
					}
					if file.Base.FeatureMetrics[keys][keys2].Second != nil {
						value.ChangeInPrevalance.W2 = file.Base.FeatureMetrics[keys][keys2].Second.(float64)
					} else {
						value.ChangeInPrevalance.W2 = 0
					}
					if _, exists := file.Base.FeatureMetrics[keys][keys2]; exists {
						value.ChangeInPrevalance.IsIncreased = file.Base.FeatureMetrics[keys][keys2].PercentChange > 0
						value.ChangeInPrevalance.Percentage = file.Base.FeatureMetrics[keys][keys2].PercentChange
					}

					value.ChangeInConversion.W1 = value.ActualValues.W1 / value.ChangeInPrevalance.W1 * 100
					value.ChangeInConversion.W2 = value.ActualValues.W2 / value.ChangeInPrevalance.W2 * 100
					value.ChangeInConversion.Percentage = ((value.ChangeInConversion.W2 - value.ChangeInConversion.W1) / value.ChangeInConversion.W1) * 100
					value.ChangeInConversion.IsIncreased = value.ChangeInConversion.Percentage > 0

					value.Type = "conversion"

					valWithDetailsArr = append(valWithDetailsArr, value)
				}
			}
		}
	}
	sort.Slice(valWithDetailsArr, func(i, j int) bool {
		return valWithDetailsArr[j].ActualValues.DeltaRatio < valWithDetailsArr[i].ActualValues.DeltaRatio
	})
	var ActualValuearr []ActualMetrics

	if insights.Goal.IsIncreased {
		// dividing records into 70% and 30%
		floatValue := float64(0.7) * float64(numberOfRecords)
		increasedRecords = int(floatValue)
		decreasedRecords = numberOfRecords - increasedRecords
	} else {
		floatValue := float64(0.7) * float64(numberOfRecords)
		decreasedRecords = int(floatValue)
		increasedRecords = numberOfRecords - increasedRecords
	}
	for _, data := range valWithDetailsArr {
		if increasedRecords == 0 && decreasedRecords == 0 {
			break
		}
		var tempActualValue ActualMetrics
		tempActualValue = ActualMetrics{
			ActualValues: Base{
				W1:          data.ActualValues.W1,
				W2:          data.ActualValues.W2,
				IsIncreased: data.ActualValues.Per > 0,
				Percentage:  data.ActualValues.Per,
			},
		}
		propertyMap[data.Key] = true
		tempActualValue.Key = data.Key
		tempActualValue.Value = data.Value
		tempActualValue.Entity = data.Entity
		tempActualValue.ChangeInConversion = data.ChangeInConversion
		tempActualValue.ChangeInPrevalance = data.ChangeInPrevalance
		tempActualValue.Type = data.Type
		if tempActualValue.ActualValues.IsIncreased && increasedRecords > 0 {
			ActualValuearr = append(ActualValuearr, tempActualValue)
			increasedRecords -= 1
		} else if !tempActualValue.ActualValues.IsIncreased && decreasedRecords > 0 {
			ActualValuearr = append(ActualValuearr, tempActualValue)
			decreasedRecords -= 1
		}
	}
	// sort ActualValuearr based on ActualValuearr.ActualValues.IsIncreased
	sort.Slice(ActualValuearr, func(i, j int) bool {
		return ActualValuearr[i].ActualValues.IsIncreased == insights.Goal.IsIncreased
	})
	insights.Insights = append(insights.Insights, ActualValuearr...)

	//for distribution

	var valWithDetailsArr2 []ValueWithDetails
	if EventType == CRM || EventType == WebsiteEvent {
		for keys := range file.Target.FeatureMetrics {
			if strings.HasPrefix(keys, "t#") {

				var val2 ValueWithDetails
				var temp BaseTargetMetrics

				for keys2 := range file.Target.FeatureMetrics[keys] {
					if keys2 == "" { // omitting "" values
						continue
					}
					if KeyMapForDistribution[keys2] {
						continue
					}
					KeyMapForDistribution[keys2] = true
					val2.Key = keys[5:]
					val2.Value = keys2
					val2.Entity = keys[2:4]
					if file.Target.FeatureMetrics[keys][keys2].First != nil {
						temp.W1 = file.Target.FeatureMetrics[keys][keys2].First.(float64)
					} else {
						temp.W1 = 0
					}
					if file.Target.FeatureMetrics[keys][keys2].Second != nil {
						temp.W2 = file.Target.FeatureMetrics[keys][keys2].Second.(float64)
					} else {
						temp.W2 = 0
					}
					if _, exists := file.Target.FeatureMetrics[keys][keys2]; exists {
						temp.Per = file.Target.FeatureMetrics[keys][keys2].PercentChange
					}
					if _, exists := file.JSDivergence.Base[keys][keys2]; exists {
						temp.JSDivergence = file.JSDivergence.Base[keys][keys2]
					}
					val2.ActualValues = temp

					val2.ChangeInDistribution.W1 = (val2.ActualValues.W1 / insights.Goal.W1) * 100
					val2.ChangeInDistribution.W2 = (val2.ActualValues.W2 / insights.Goal.W2) * 100
					val2.ChangeInDistribution.IsIncreased = val2.ChangeInDistribution.W1 < val2.ChangeInDistribution.W2
					val2.ChangeInDistribution.Percentage = (val2.ChangeInDistribution.W1) - (val2.ChangeInDistribution.W2)

					val2.Type = "distribution"
					valWithDetailsArr2 = append(valWithDetailsArr2, val2)
				}
			}
		}
	} else if EventType == Funnel {
		for keys := range file.BaseAndTarget.FeatureMetrics {
			if strings.HasPrefix(keys, "t#") {

				var val2 ValueWithDetails
				var temp BaseTargetMetrics

				for keys2 := range file.BaseAndTarget.FeatureMetrics[keys] {
					if keys2 == "" { // omitting "" values
						continue
					}
					if KeyMapForDistribution[keys2] {
						continue
					}
					KeyMapForDistribution[keys2] = true
					val2.Key = keys[5:]
					val2.Value = keys2
					val2.Entity = keys[2:4]
					if file.BaseAndTarget.FeatureMetrics[keys][keys2].First != nil {
						temp.W1 = file.BaseAndTarget.FeatureMetrics[keys][keys2].First.(float64)
					} else {
						temp.W1 = 0
					}
					if file.BaseAndTarget.FeatureMetrics[keys][keys2].Second != nil {
						temp.W2 = file.BaseAndTarget.FeatureMetrics[keys][keys2].Second.(float64)
					} else {
						temp.W2 = 0
					}
					if _, exists := file.BaseAndTarget.FeatureMetrics[keys][keys2]; exists {
						temp.Per = file.BaseAndTarget.FeatureMetrics[keys][keys2].PercentChange
					}
					if _, exists := file.JSDivergence.Base[keys][keys2]; exists {
						temp.JSDivergence = file.JSDivergence.Base[keys][keys2]
					}
					val2.ActualValues = temp

					val2.ChangeInDistribution.W1 = (val2.ActualValues.W1 / insights.Goal.W1) * 100
					val2.ChangeInDistribution.W2 = (val2.ActualValues.W2 / insights.Goal.W2) * 100
					val2.ChangeInDistribution.IsIncreased = val2.ChangeInDistribution.W1 < val2.ChangeInDistribution.W2
					val2.ChangeInDistribution.Percentage = (val2.ChangeInDistribution.W1) - (val2.ChangeInDistribution.W2)

					val2.Type = "distribution"
					valWithDetailsArr2 = append(valWithDetailsArr2, val2)
				}
			}
		}
	}
	sort.Slice(valWithDetailsArr2, func(i, j int) bool {
		return valWithDetailsArr2[j].ActualValues.JSDivergence < valWithDetailsArr2[i].ActualValues.JSDivergence
	})
	var ActualValuearr2 []ActualMetrics

	if insights.Goal.IsIncreased {
		// dividing records into 70% and 30%
		floatValue := float64(0.7) * float64(numberOfRecords)
		increasedRecords = int(floatValue)
		decreasedRecords = numberOfRecords - increasedRecords
	} else {
		floatValue := float64(0.7) * float64(numberOfRecords)
		decreasedRecords = int(floatValue)
		increasedRecords = numberOfRecords - increasedRecords
	}
	for _, data := range valWithDetailsArr2 {
		if increasedRecords == 0 && decreasedRecords == 0 {
			break
		}
		var tempActualValue ActualMetrics
		tempActualValue = ActualMetrics{
			ActualValues: Base{
				W1:          data.ActualValues.W1,
				W2:          data.ActualValues.W2,
				IsIncreased: data.ActualValues.Per > 0,
				Percentage:  data.ActualValues.Per,
			},
		}
		tempActualValue.ChangeInDistribution = data.ChangeInDistribution
		propertyMap[data.Key] = true
		tempActualValue.Key = data.Key
		tempActualValue.Value = data.Value
		tempActualValue.Entity = data.Entity
		tempActualValue.Type = data.Type

		if tempActualValue.ActualValues.IsIncreased && increasedRecords > 0 {
			ActualValuearr2 = append(ActualValuearr2, tempActualValue)
			increasedRecords -= 1
		} else if !tempActualValue.ActualValues.IsIncreased && decreasedRecords > 0 {
			ActualValuearr2 = append(ActualValuearr2, tempActualValue)
			decreasedRecords -= 1
		}
	}
	sort.Slice(ActualValuearr2, func(i, j int) bool {
		return ActualValuearr2[i].ActualValues.IsIncreased == insights.Goal.IsIncreased
	})
	insights.Insights = append(insights.Insights, ActualValuearr2...)
	return insights
}
func GetWeeklyInsights(projectId uint64, queryId uint64, baseStartTime *time.Time, compStartTime *time.Time, insightsType string, numberOfRecords int) (interface{}, error) {
	path, file := C.GetCloudManager().GetInsightsCpiFilePathAndName(projectId, U.GetDateOnlyFromTimestamp(baseStartTime.Unix()), queryId, 10)
	fmt.Println(path, file)
	reader, err := C.GetCloudManager().Get(path, file)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return nil, err
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return nil, err
	}

	var insights CrossPeriodInsights
	err = json.Unmarshal(data, &insights)
	if err != nil {
		log.WithError(err).Error("Error unmarshalling response")
		return nil, err
	}
	// finding query class and query object
	QueriesObj, status := store.GetStore().GetQueryWithQueryId(projectId, queryId)
	if status != http.StatusFound {
		log.Error("query not found")
		return nil, errors.New("query not found")
	}
	class, errMsg := store.GetStore().GetQueryClassFromQueries(*QueriesObj)
	if errMsg != "" {
		return nil, errors.New(errMsg)
	}
	var query model.Query
	if class == model.QueryClassFunnel {
		err = U.DecodePostgresJsonbToStructType(&QueriesObj.Query, &query)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	} else if class == model.QueryClassEvents {
		var queryGroup model.QueryGroup
		err = U.DecodePostgresJsonbToStructType(&QueriesObj.Query, &queryGroup)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		query = queryGroup.Queries[0]
	}
	EventType := getEventType(&query, class, projectId)
	insightsObj := GetInsights(insights, numberOfRecords, class, EventType)
	// adding query groups

	gbpInsights := addGroupByProperties(query, EventType, insights, insightsObj)
	// appending at top
	insightsObj.Insights = append(gbpInsights, insightsObj.Insights...)
	removeNegativePercentageFromInsights(&insightsObj)
	return insightsObj, nil
}

// TODO: add changes in this method acc to excel
func addGroupByProperties(query model.Query, EventType string, file CrossPeriodInsights, insights WeeklyInsights) []ActualMetrics {
	var ActualMetricsArr []ActualMetrics
	for _, gbp := range query.GroupByProperties {
		var properties []string
		if gbp.Entity == model.PropertyEntityUser {
			properties = append(properties, "t#up#"+gbp.Property)
			properties = append(properties, "s#up#"+gbp.Property)
		} else if gbp.Entity == model.PropertyEntityEvent {
			properties = append(properties, "t#ep#"+gbp.Property)
			properties = append(properties, "s#ep#"+gbp.Property)
		}
		if !propertyMap[gbp.Property] {
			var valWithDetailsArr []ValueWithDetails
			if EventType == Funnel || EventType == CRM {
				for _, property := range properties {
					for values := range file.BaseAndTarget.FeatureMetrics[property] { // conversion
						var newData ValueWithDetails
						var temp BaseTargetMetrics
						newData.Key = gbp.Property
						newData.Value = values
						newData.Entity = gbp.Entity

						if file.BaseAndTarget.FeatureMetrics[property][values].First != nil {
							temp.W1 = file.BaseAndTarget.FeatureMetrics[property][values].First.(float64)
						}
						if file.BaseAndTarget.FeatureMetrics[property][values].Second != nil {
							temp.W2 = file.BaseAndTarget.FeatureMetrics[property][values].Second.(float64)
						}
						if _, exists := file.BaseAndTarget.FeatureMetrics[property][values]; exists {
							temp.Per = file.BaseAndTarget.FeatureMetrics[property][values].PercentChange
						}
						if _, exists := file.DeltaRatio[property][values]; exists {
							temp.DeltaRatio = file.DeltaRatio[property][values]
						}

						if file.Conversion.FeatureMetrics[property][values].First != nil {
							newData.ChangeInConversion.W1 = file.Conversion.FeatureMetrics[property][values].First.(float64)
						}
						if file.Conversion.FeatureMetrics[property][values].Second != nil {
							newData.ChangeInConversion.W2 = file.Conversion.FeatureMetrics[property][values].Second.(float64)
						}
						if _, exists := file.Conversion.FeatureMetrics[property][values]; exists {
							newData.ChangeInConversion.IsIncreased = file.Conversion.FeatureMetrics[property][values].PercentChange > 0
							newData.ChangeInConversion.Percentage = file.Conversion.FeatureMetrics[property][values].PercentChange
						}

						if file.Base.FeatureMetrics[property][values].First != nil {
							newData.ChangeInPrevalance.W1 = file.Base.FeatureMetrics[property][values].First.(float64)
						}
						if file.Base.FeatureMetrics[property][values].Second != nil {
							newData.ChangeInPrevalance.W2 = file.Base.FeatureMetrics[property][values].Second.(float64)
						}
						if _, exists := file.Base.FeatureMetrics[property][values]; exists {
							newData.ChangeInPrevalance.IsIncreased = file.Base.FeatureMetrics[property][values].PercentChange > 0
							newData.ChangeInPrevalance.Percentage = file.Base.FeatureMetrics[property][values].PercentChange
						}
						newData.ActualValues = temp
						newData.Type = "conversion"

						valWithDetailsArr = append(valWithDetailsArr, newData)

						sort.Slice(valWithDetailsArr, func(i, j int) bool {
							return valWithDetailsArr[j].ActualValues.DeltaRatio < valWithDetailsArr[i].ActualValues.DeltaRatio
						})

						for index, data := range valWithDetailsArr {
							if index >= numberOfRecordsFromGbp {
								break
							}
							var tempActualValue ActualMetrics
							tempActualValue = ActualMetrics{
								ActualValues: Base{
									W1:          data.ActualValues.W1,
									W2:          data.ActualValues.W2,
									IsIncreased: data.ActualValues.Per > 0,
									Percentage:  data.ActualValues.Per,
								},
							}

							tempActualValue.Key = data.Key
							tempActualValue.Value = data.Value
							tempActualValue.Entity = data.Entity
							tempActualValue.ChangeInConversion = data.ChangeInConversion
							tempActualValue.ChangeInPrevalance = data.ChangeInPrevalance
							tempActualValue.Type = data.Type
							ActualMetricsArr = append(ActualMetricsArr, tempActualValue)

						}
					}
				}
			}
			var valWithDetailsArr2 []ValueWithDetails
			if EventType == CRM || EventType == WebsiteEvent {
				for _, property := range properties {
					for values := range file.Target.FeatureMetrics[property] { // distribution
						var newData ValueWithDetails
						var temp BaseTargetMetrics
						newData.Key = gbp.Property
						newData.Value = values
						newData.Entity = gbp.Entity
						if file.Target.FeatureMetrics[property][values].First != nil {
							temp.W1 = file.Target.FeatureMetrics[property][values].First.(float64)
						}
						if file.Target.FeatureMetrics[property][values].Second != nil {
							temp.W2 = file.Target.FeatureMetrics[property][values].Second.(float64)
						}
						if _, exists := file.Target.FeatureMetrics[property][values]; exists {
							temp.Per = file.Target.FeatureMetrics[property][values].PercentChange
						}
						if _, exists := file.JSDivergence.Base[property][values]; exists {
							temp.JSDivergence = file.JSDivergence.Base[property][values]

						}
						newData.ActualValues = temp
						newData.ChangeInDistribution.W1 = (newData.ActualValues.W1 / insights.Goal.W1) * 100
						newData.ChangeInDistribution.W2 = (newData.ActualValues.W2 / insights.Goal.W2) * 100
						newData.ChangeInDistribution.IsIncreased = newData.ChangeInDistribution.W1 < newData.ChangeInDistribution.W2
						newData.ChangeInDistribution.Percentage = (newData.ChangeInDistribution.W1) - (newData.ChangeInDistribution.W2)
						newData.Type = "distribution"

						valWithDetailsArr2 = append(valWithDetailsArr2, newData)
					}
				}
			} else if EventType == Funnel {
				for _, property := range properties {
					for values := range file.Target.FeatureMetrics[property] { // distribution
						var newData ValueWithDetails
						var temp BaseTargetMetrics
						newData.Key = gbp.Property
						newData.Value = values
						newData.Entity = gbp.Entity
						if file.BaseAndTarget.FeatureMetrics[property][values].First != nil {
							temp.W1 = file.BaseAndTarget.FeatureMetrics[property][values].First.(float64)
						}
						if file.BaseAndTarget.FeatureMetrics[property][values].Second != nil {
							temp.W2 = file.BaseAndTarget.FeatureMetrics[property][values].Second.(float64)
						}
						if _, exists := file.BaseAndTarget.FeatureMetrics[property][values]; exists {
							temp.Per = file.BaseAndTarget.FeatureMetrics[property][values].PercentChange
						}
						if _, exists := file.JSDivergence.Base[property][values]; exists {
							temp.JSDivergence = file.JSDivergence.Base[property][values]

						}
						newData.ActualValues = temp

						newData.ChangeInDistribution.W1 = (newData.ActualValues.W1 / insights.Goal.W1) * 100
						newData.ChangeInDistribution.W2 = (newData.ActualValues.W2 / insights.Goal.W2) * 100
						newData.ChangeInDistribution.IsIncreased = newData.ChangeInDistribution.W1 < newData.ChangeInDistribution.W2
						newData.ChangeInDistribution.Percentage = (newData.ChangeInDistribution.W1) - (newData.ChangeInDistribution.W2)

						newData.Type = "distribution"

						valWithDetailsArr2 = append(valWithDetailsArr2, newData)
					}
				}
			}
			sort.Slice(valWithDetailsArr2, func(i, j int) bool {
				return valWithDetailsArr2[j].ActualValues.JSDivergence < valWithDetailsArr2[i].ActualValues.JSDivergence
			})
			for index, data := range valWithDetailsArr2 {
				if index >= numberOfRecordsFromGbp {
					break
				}
				var tempActualValue ActualMetrics
				tempActualValue = ActualMetrics{
					ActualValues: Base{
						W1:          data.ActualValues.W1,
						W2:          data.ActualValues.W2,
						IsIncreased: data.ActualValues.Per > 0,
						Percentage:  data.ActualValues.Per,
					},
				}
				tempActualValue.ChangeInDistribution = data.ChangeInDistribution
				propertyMap[data.Key] = true
				tempActualValue.Key = data.Key
				tempActualValue.Value = data.Value
				tempActualValue.Entity = data.Entity
				tempActualValue.Type = data.Type
				ActualMetricsArr = append(ActualMetricsArr, tempActualValue)

			}
		}

	}

	return ActualMetricsArr
}
func getEventType(query *model.Query, QueryClass string, project_id uint64) string {
	EventType := ""
	if QueryClass == model.QueryClassFunnel {
		return Funnel
	} else if QueryClass == model.QueryClassEvents {

		ewp := query.EventsWithProperties
		for _, data := range ewp {
			name := data.Name
			eventNameObj, status := store.GetStore().GetEventName(name, project_id)
			if status != http.StatusFound {
				log.Error("Not found "+name+" ", project_id)
				continue
			}
			if eventNameObj.Type == model.EVENT_NAME_TYPE_SMART_EVENT || eventNameObj.Type == model.TYPE_CRM_SALESFORCE || eventNameObj.Type == model.TYPE_CRM_HUBSPOT || strings.HasPrefix(eventNameObj.Name, "$hubspot") || strings.HasPrefix(eventNameObj.Name, "$sf") || strings.HasPrefix(eventNameObj.Name, "$session") {
				EventType = CRM
			} else {
				EventType = WebsiteEvent
				break
			}
		}

	}
	return EventType

}
func removeNegativePercentageFromInsights(insightsObj *WeeklyInsights) {
	insightsObj.Base.Percentage = math.Abs(insightsObj.Base.Percentage)
	insightsObj.Goal.Percentage = math.Abs(insightsObj.Goal.Percentage)
	insightsObj.Conv.Percentage = math.Abs(insightsObj.Conv.Percentage)

	for index := range insightsObj.Insights {
		insightsObj.Insights[index].ActualValues.Percentage = math.Abs(insightsObj.Insights[index].ActualValues.Percentage)
		insightsObj.Insights[index].ChangeInConversion.Percentage = math.Abs(insightsObj.Insights[index].ActualValues.Percentage)
		insightsObj.Insights[index].ChangeInPrevalance.Percentage = math.Abs(insightsObj.Insights[index].ChangeInPrevalance.Percentage)
		insightsObj.Insights[index].ChangeInDistribution.Percentage = math.Abs(insightsObj.Insights[index].ChangeInDistribution.Percentage)
	}

}
