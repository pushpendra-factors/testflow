package delta

import (
	"encoding/json"
	"errors"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	C "factors/config"

	log "github.com/sirupsen/logrus"
)

type WeeklyInsights struct {
	Base     Base            `json:"base"`
	Goal     Base            `json:"goal"`
	Conv     Base            `json:"conv"`
	Insights []ActualMetrics `json:"actual_metrics"`
}
type Base struct {
	W1          float64 `json:"w1"`
	W2          float64 `json:"w2"`
	IsIncreased bool    `json:"isIncrease"`
	Percentage  float64 `json:"percentage"`
}

type ActualMetrics struct {
	Key                string `json:"key"`
	Value              string `json:"value"`
	Entity             string `json:"entity"`
	ActualValues       Base   `json:"actual_values"`
	ChangeInConversion Base   `json:"change_in_conversion"`
	ChangeInPrevalance Base   `json:"change_in_prevalance"`
	Type               string `json:"type"`
}

// temporary structure to hold values
type ValueWithDetails struct {
	Key                string            `json:"key"`
	Value              string            `json:"value"`
	Entity             string            `json:"entity"`
	ActualValues       BaseTargetMetrics `json:"actual_values"`
	ChangeInConversion Base              `json:"change_in_conversion"`
	ChangeInPrevalance Base              `json:"change_in_prevalance"`
	Type               string            `json:"type"`
}
type BaseTargetMetrics struct {
	W1           float64 `json:"w1"`
	W2           float64 `json:"w2"`
	Per          float64 `json:"per"`
	DeltaRatio   float64 `json:"delrat"`
	JSDivergence float64 `json:"jsd"`
}

var numberOfRecordsFromGbp int = 5 // number of records to be fetched from gbp
var propertyMap map[string]bool

func GetInsights(file CrossPeriodInsights, numberOfRecords int, QueryClass string) WeeklyInsights {
	var KeyMapForConversion = make(map[string]bool)
	var KeyMapForDistribution = make(map[string]bool)
	propertyMap = make(map[string]bool)
	var insights WeeklyInsights
	insights.Insights = make([]ActualMetrics, 0)
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
	if QueryClass == model.QueryClassEvents { // pulling according to type of event
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
	} else if QueryClass == model.QueryClassFunnel {
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
	}

	if _, exists := file.Conversion.GlobalMetrics["ratio"]; exists {
		if file.Conversion.GlobalMetrics["ratio"].First != nil {
			insights.Conv.W1 = file.Conversion.GlobalMetrics["ratio"].First.(float64)
		} else {
			insights.Conv.W1 = 0
		}
		if file.Conversion.GlobalMetrics["ratio"].Second != nil {
			insights.Conv.W2 = file.Conversion.GlobalMetrics["ratio"].Second.(float64)
		} else {
			insights.Conv.W2 = 0
		}
		insights.Conv.IsIncreased = file.Conversion.GlobalMetrics["ratio"].PercentChange > 0
		insights.Conv.Percentage = file.Conversion.GlobalMetrics["ratio"].PercentChange
	}

	var valWithDetailsArr []ValueWithDetails
	// for conversion
	for keys := range file.BaseAndTarget.FeatureMetrics {
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
			value.Key = keys[3:]
			value.Value = keys2
			value.Entity = keys[0:2]
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

			if file.Conversion.FeatureMetrics[keys][keys2].First != nil {
				value.ChangeInConversion.W1 = file.Conversion.FeatureMetrics[keys][keys2].First.(float64)
			} else {
				value.ChangeInConversion.W1 = 0
			}
			if file.Conversion.FeatureMetrics[keys][keys2].Second != nil {
				value.ChangeInConversion.W2 = file.Conversion.FeatureMetrics[keys][keys2].Second.(float64)
			} else {
				value.ChangeInConversion.W2 = 0
			}
			if _, exists := file.Conversion.FeatureMetrics[keys][keys2]; exists {
				value.ChangeInConversion.IsIncreased = file.Conversion.FeatureMetrics[keys][keys2].PercentChange > 0
				value.ChangeInConversion.Percentage = file.Conversion.FeatureMetrics[keys][keys2].PercentChange
			}

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
			value.Type = "conversion"

			valWithDetailsArr = append(valWithDetailsArr, value)
		}
	}
	sort.Slice(valWithDetailsArr, func(i, j int) bool {
		return valWithDetailsArr[j].ActualValues.DeltaRatio < valWithDetailsArr[i].ActualValues.DeltaRatio
	})
	var ActualValuearr []ActualMetrics
	for index, data := range valWithDetailsArr {
		if index >= numberOfRecords {
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
		ActualValuearr = append(ActualValuearr, tempActualValue)

	}
	insights.Insights = append(insights.Insights, ActualValuearr...)

	//for distribution

	var valWithDetailsArr2 []ValueWithDetails
	for keys := range file.Target.FeatureMetrics {

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
			val2.Key = keys[3:]
			val2.Value = keys2
			val2.Entity = keys[0:2]
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
			val2.Type = "distribution"
			valWithDetailsArr2 = append(valWithDetailsArr2, val2)
		}
	}
	sort.Slice(valWithDetailsArr2, func(i, j int) bool {
		return valWithDetailsArr2[j].ActualValues.JSDivergence < valWithDetailsArr2[i].ActualValues.JSDivergence
	})
	var ActualValuearr2 []ActualMetrics
	for index, data := range valWithDetailsArr2 {
		if index >= numberOfRecords {
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
		tempActualValue.Type = data.Type
		ActualValuearr2 = append(ActualValuearr2, tempActualValue)

	}
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
		fmt.Println(query.GroupByProperties)
	} else if class == model.QueryClassEvents {
		var queryGroup model.QueryGroup
		err = U.DecodePostgresJsonbToStructType(&QueriesObj.Query, &queryGroup)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		query = queryGroup.Queries[0]
	}
	fmt.Println(class)
	insightsObj := GetInsights(insights, numberOfRecords, class)

	// adding query groups

	gbpInsights := addGroupByProperties(query, insights)
	insightsObj.Insights = append(insightsObj.Insights, gbpInsights...)
	return insightsObj, nil
}
func addGroupByProperties(query model.Query, insights CrossPeriodInsights) []ActualMetrics {
	var ActualMetricsArr []ActualMetrics
	for _, gbp := range query.GroupByProperties {
		var property string
		if gbp.Entity == model.PropertyEntityUser {
			property = "up#" + gbp.Property
		} else if gbp.Entity == model.PropertyEntityEvent {
			property = "ep#" + gbp.Property
		}
		if !propertyMap[gbp.Property] {
			var valWithDetailsArr []ValueWithDetails
			for values := range insights.BaseAndTarget.FeatureMetrics[property] { // conversion
				var newData ValueWithDetails
				var temp BaseTargetMetrics
				newData.Key = gbp.Property
				newData.Value = values
				newData.Entity = gbp.Entity

				if insights.BaseAndTarget.FeatureMetrics[property][values].First != nil {
					temp.W1 = insights.BaseAndTarget.FeatureMetrics[property][values].First.(float64)
				}
				if insights.BaseAndTarget.FeatureMetrics[property][values].Second != nil {
					temp.W2 = insights.BaseAndTarget.FeatureMetrics[property][values].Second.(float64)
				}
				if _, exists := insights.BaseAndTarget.FeatureMetrics[property][values]; exists {
					temp.Per = insights.BaseAndTarget.FeatureMetrics[property][values].PercentChange
				}
				if _, exists := insights.DeltaRatio[property][values]; exists {
					temp.DeltaRatio = insights.DeltaRatio[property][values]
				}

				if insights.Conversion.FeatureMetrics[property][values].First != nil {
					newData.ChangeInConversion.W1 = insights.Conversion.FeatureMetrics[property][values].First.(float64)
				}
				if insights.Conversion.FeatureMetrics[property][values].Second != nil {
					newData.ChangeInConversion.W2 = insights.Conversion.FeatureMetrics[property][values].Second.(float64)
				}
				if _, exists := insights.Conversion.FeatureMetrics[property][values]; exists {
					newData.ChangeInConversion.IsIncreased = insights.Conversion.FeatureMetrics[property][values].PercentChange > 0
					newData.ChangeInConversion.Percentage = insights.Conversion.FeatureMetrics[property][values].PercentChange
				}

				if insights.Base.FeatureMetrics[property][values].First != nil {
					newData.ChangeInPrevalance.W1 = insights.Base.FeatureMetrics[property][values].First.(float64)
				}
				if insights.Base.FeatureMetrics[property][values].Second != nil {
					newData.ChangeInPrevalance.W2 = insights.Base.FeatureMetrics[property][values].Second.(float64)
				}
				if _, exists := insights.Base.FeatureMetrics[property][values]; exists {
					newData.ChangeInPrevalance.IsIncreased = insights.Base.FeatureMetrics[property][values].PercentChange > 0
					newData.ChangeInPrevalance.Percentage = insights.Base.FeatureMetrics[property][values].PercentChange
				}
				newData.ActualValues = temp
				newData.Type = "conversion"

				valWithDetailsArr = append(valWithDetailsArr, newData)
			}
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

			var valWithDetailsArr2 []ValueWithDetails
			for values := range insights.Target.FeatureMetrics[property] { // distribution
				var newData ValueWithDetails
				var temp BaseTargetMetrics
				newData.Key = gbp.Property
				newData.Value = values
				newData.Entity = gbp.Entity
				if insights.BaseAndTarget.FeatureMetrics[property][values].First != nil {
					temp.W1 = insights.Target.FeatureMetrics[property][values].First.(float64)
				}
				if insights.Target.FeatureMetrics[property][values].Second != nil {
					temp.W2 = insights.BaseAndTarget.FeatureMetrics[property][values].Second.(float64)
				}
				if _, exists := insights.Target.FeatureMetrics[property][values]; exists {
					temp.Per = insights.BaseAndTarget.FeatureMetrics[property][values].PercentChange
				}
				if _, exists := insights.JSDivergence.Base[property][values]; exists {
					temp.JSDivergence = insights.JSDivergence.Base[property][values]

				}
				newData.ActualValues = temp
				newData.Type = "distribution"

				valWithDetailsArr2 = append(valWithDetailsArr2, newData)
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
