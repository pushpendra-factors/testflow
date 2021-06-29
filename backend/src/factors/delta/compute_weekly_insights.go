package delta

import (
	"encoding/json"
	U "factors/util"
	"fmt"
	"io/ioutil"
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
	W1          uint64  `json:"w1"`
	W2          uint64  `json:"w2"`
	IsIncreased bool    `json:"isIncrease"`
	Percentage  float64 `json:"percentage"`
}

type ActualMetrics struct {
	Key                string `json:"key"`
	Value              string `json:"value"`
	ActualValues       Base   `json:"actual_values"`
	ChangeInConversion Base   `json:"change_in_conversion"`
	ChangeInPrevalance Base   `json:"change_in_prevalance"`
	Type               string `json:"type"`
}

// temporary structure to hold values
type ValueWithDetails struct {
	Key                string            `json:"key"`
	Value              string            `json:"value"`
	ActualValues       BaseTargetMetrics `json:"actual_values"`
	ChangeInConversion Base              `json:"change_in_conversion"`
	ChangeInPrevalance Base              `json:"change_in_prevalance"`
	Type               string            `json:"type"`
}
type BaseTargetMetrics struct {
	W1           uint64  `json:"w1"`
	W2           uint64  `json:"w2"`
	Per          float64 `json:"per"`
	DeltaRatio   float64 `json:"delrat"`
	JSDivergence float64 `json:"jsd"`
}

func GetInsights(file CrossPeriodInsights, numberOfRecords int) interface{} {
	var insights WeeklyInsights

	if _, exists := file.Base.GlobalMetrics["#users"]; exists {

		if file.Base.GlobalMetrics["#users"].First != nil {
			insights.Base.W1 = uint64(file.Base.GlobalMetrics["#users"].First.(float64))
		} else {
			insights.Base.W1 = 0
		}
		if file.Base.GlobalMetrics["#users"].Second != nil {
			insights.Base.W2 = uint64(file.Base.GlobalMetrics["#users"].Second.(float64))
		} else {
			insights.Base.W2 = 0
		}
		insights.Base.IsIncreased = file.Base.GlobalMetrics["#users"].PercentChange > 0
		insights.Base.Percentage = file.Base.GlobalMetrics["#users"].PercentChange
	}
	if _, exists := file.Target.GlobalMetrics["#users"]; exists {
		if file.Target.GlobalMetrics["#users"].First != nil {
			insights.Goal.W1 = uint64(file.Target.GlobalMetrics["#users"].First.(float64))
		} else {
			insights.Goal.W1 = 0
		}
		if file.Target.GlobalMetrics["#users"].Second != nil {
			insights.Goal.W2 = uint64(file.Target.GlobalMetrics["#users"].Second.(float64))
		} else {
			insights.Goal.W2 = 0
		}
		insights.Goal.IsIncreased = file.Target.GlobalMetrics["#users"].PercentChange > 0
		insights.Goal.Percentage = file.Target.GlobalMetrics["#users"].PercentChange
	}
	if _, exists := file.Conversion.GlobalMetrics["ratio"]; exists {
		if file.Conversion.GlobalMetrics["ratio"].First != nil {
			insights.Conv.W1 = uint64(file.Conversion.GlobalMetrics["ratio"].First.(float64))
		} else {
			insights.Conv.W1 = 0
		}
		if file.Conversion.GlobalMetrics["ratio"].Second != nil {
			insights.Conv.W2 = uint64(file.Conversion.GlobalMetrics["ratio"].Second.(float64))
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
			value.Key = keys
			value.Value = keys2
			if file.BaseAndTarget.FeatureMetrics[keys][keys2].First != nil {
				temp.W1 = uint64(file.BaseAndTarget.FeatureMetrics[keys][keys2].First.(float64))
			} else {
				temp.W1 = 0
			}
			if file.BaseAndTarget.FeatureMetrics[keys][keys2].Second != nil {
				temp.W2 = uint64(file.BaseAndTarget.FeatureMetrics[keys][keys2].Second.(float64))
			} else {
				temp.W2 = 0
			}
			if _, exists := file.BaseAndTarget.FeatureMetrics[keys][keys2]; exists {
				temp.Per = file.BaseAndTarget.FeatureMetrics[keys][keys2].PercentChange
			}
			if _, exists := file.DeltaRatio[keys][keys2]; exists {
				temp.DeltaRatio = file.DeltaRatio[keys][keys2]
			}
			if _, exists := file.JSDivergence.Base[keys][keys2]; exists {
				temp.JSDivergence = file.JSDivergence.Base[keys][keys2]
			}

			value.ActualValues = temp

			if file.Conversion.FeatureMetrics[keys][keys2].First != nil {
				value.ChangeInConversion.W1 = uint64(file.Conversion.FeatureMetrics[keys][keys2].First.(float64))
			} else {
				value.ChangeInConversion.W1 = 0
			}
			if file.Conversion.FeatureMetrics[keys][keys2].Second != nil {
				value.ChangeInConversion.W2 = uint64(file.Conversion.FeatureMetrics[keys][keys2].Second.(float64))
			} else {
				value.ChangeInConversion.W2 = 0
			}
			if _, exists := file.Conversion.FeatureMetrics[keys][keys2]; exists {
				value.ChangeInConversion.IsIncreased = file.Conversion.FeatureMetrics[keys][keys2].PercentChange > 0
				value.ChangeInConversion.Percentage = file.Conversion.FeatureMetrics[keys][keys2].PercentChange
			}

			if file.Base.FeatureMetrics[keys][keys2].First != nil {
				value.ChangeInPrevalance.W1 = uint64(file.Base.FeatureMetrics[keys][keys2].First.(float64))
			} else {
				value.ChangeInPrevalance.W1 = 0
			}
			if file.Base.FeatureMetrics[keys][keys2].Second != nil {
				value.ChangeInPrevalance.W2 = uint64(file.Base.FeatureMetrics[keys][keys2].Second.(float64))
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
		tempActualValue =ActualMetrics{
			ActualValues:Base{
				W1: data.ActualValues.W1,
				W2: data.ActualValues.W2,
				IsIncreased: data.ActualValues.Per>0,
				Percentage: data.ActualValues.Per,
			},			
			
		}
		tempActualValue.Key = data.Key
		tempActualValue.Value = data.Value
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

			val2.Key = keys
			val2.Value = keys2
			if file.Target.FeatureMetrics[keys][keys2].First != nil {
				temp.W1 = uint64(file.Target.FeatureMetrics[keys][keys2].First.(float64))
			} else {
				temp.W1 = 0
			}
			if file.Target.FeatureMetrics[keys][keys2].Second != nil {
				temp.W2 = uint64(file.Target.FeatureMetrics[keys][keys2].Second.(float64))
			} else {
				temp.W2 = 0
			}
			if _, exists := file.Target.FeatureMetrics[keys][keys2]; exists {
				temp.Per = file.Target.FeatureMetrics[keys][keys2].PercentChange
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
		tempActualValue =ActualMetrics{
			ActualValues:Base{
				W1: data.ActualValues.W1,
				W2: data.ActualValues.W2,
				IsIncreased: data.ActualValues.Per>0,
				Percentage: data.ActualValues.Per,
			},			
			
		}
		tempActualValue.Key = data.Key
		tempActualValue.Value = data.Value
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
	insightsObj := GetInsights(insights, numberOfRecords)
	return insightsObj, nil
}
