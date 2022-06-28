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
	QueryId      uint64          `json:"query_id"`
}
type Base struct {
	W1          float64 `json:"w1"`
	W2          float64 `json:"w2"`
	IsIncreased bool    `json:"isIncrease"`
	Percentage  float64 `json:"percentage"`
}
type BaseWithPerc struct {
	W1          [2]float64 `json:"w1"`
	W2          [2]float64 `json:"w2"`
	IsIncreased [2]bool    `json:"isIncrease"`
	Percentage  [2]float64 `json:"percentage"`
}

type ActualMetrics struct {
	Category             string       `json:"category"`
	Key                  string       `json:"key"`
	Value                string       `json:"value"`
	Entity               string       `json:"entity"`
	VoteStatus           string       `json:"vote_status"`
	ActualValues         Base         `json:"actual_values"`
	ChangeInConversion   Base         `json:"change_in_conversion"`
	ChangeInPrevalance   Base         `json:"change_in_prevalance"`
	ChangeInDistribution Base         `json:"change_in_distribution"`
	ChangeInScale        BaseWithPerc `json:"change_in_scale"`
	Type                 string       `json:"type"`
}

// temporary structure to hold values
type ValueWithDetails struct {
	Category             string            `json:"category"`
	Key                  string            `json:"key"`
	Value                string            `json:"value"`
	Entity               string            `json:"entity"`
	VoteStatus           string            `json:"vote_status"`
	ActualValues         BaseTargetMetrics `json:"actual_values"`
	ChangeInConversion   Base              `json:"change_in_conversion"`
	ChangeInPrevalance   Base              `json:"change_in_prevalance"`
	ChangeInDistribution Base              `json:"change_in_distribution"`
	ChangeInScale        BaseWithPerc      `json:"change_in_scale"`
	Type                 string            `json:"type"`
}
type BaseTargetMetrics struct {
	W1           float64 `json:"w1"`
	W2           float64 `json:"w2"`
	Per          float64 `json:"per"`
	DeltaRatio   float64 `json:"delrat"`
	JSDivergence float64 `json:"jsd"`
}

var numberOfRecordsFromGbp int = 2 // number of records to be fetched from gbp
var increasedRecords int
var decreasedRecords int
var propertyMap map[string]bool

var WebsiteEvent string = "WebsiteEvent"
var CRM string = "CRM"
var Funnel string = "Funnel"
var KPI string = "KPI"

var BlackListedKeys map[string]bool
var WhiteListedKeys map[string]bool
var WhiteListedKeysOtherQuery map[string]bool
var DecreaseBoostKeys map[string]bool
var PriorityKeysConversion map[string]float64
var PriorityKeysDistribution map[string]float64

const (
	Upvoted                = "upvoted"
	UpvotedForOtherQuery   = "upvoted_other_query"
	Downvoted              = "downvoted"
	DownvotedForOtherQuery = "downvoted_other_query"
)

const DistributionChangePer float64 = 5 // x of overall to be comapared with distrubution W1

func getInsightImportance(valWithDetails ValueWithDetails) float64 {
	var cons float64
	w1 := valWithDetails.ActualValues.W1
	w2 := valWithDetails.ActualValues.W2
	dist1 := valWithDetails.ChangeInScale.W1[1]
	dist2 := valWithDetails.ChangeInScale.W2[1]
	if w2 == 0 || w1 == 0 {
		return 0
	}
	if valWithDetails.Category == "kpi_events" {
		cons = (dist1 + dist2)
	} else {
		cons = 1
	}
	return math.Abs(w2-w1) * cons / (w2 + w1)
}

func GetInsightsKpi(file CrossPeriodInsightsKpi, numberOfRecords int, QueryClass, KpiType string, isEventWebsite bool) WeeklyInsights {
	var KeyMapForDistribution = make(map[string]bool)
	propertyMap = make(map[string]bool)
	var insights WeeklyInsights
	insights.Insights = make([]ActualMetrics, 0)
	insights.InsightsType = "DistOnly"
	// ZeroFlag := true // flag to check if overall W1||W2 is 0.
	var tmpGlobal Base
	var valWithDetailsArr []ValueWithDetails
	var ActualValuearr []ActualMetrics
	cpiInfo := file.Target
	scaleInfo := file.ScaleInfo

	//get scale
	var globalScaleW1, globalScaleW2 float64
	if scaleInfo.GlobalMetrics.First != nil {
		globalScaleW1 = scaleInfo.GlobalMetrics.First.(float64)
	}
	if scaleInfo.GlobalMetrics.Second != nil {
		globalScaleW2 = scaleInfo.GlobalMetrics.Second.(float64)
	}

	var globalW1, globalW2 float64
	if cpiInfo.GlobalMetrics.First != nil {
		globalW1 = cpiInfo.GlobalMetrics.First.(float64)
	}
	tmpGlobal.W1 = globalW1
	if cpiInfo.GlobalMetrics.Second != nil {
		globalW2 = cpiInfo.GlobalMetrics.Second.(float64)
	}
	tmpGlobal.W2 = globalW2
	tmpGlobal.IsIncreased = cpiInfo.GlobalMetrics.PercentChange > 0
	tmpGlobal.Percentage = cpiInfo.GlobalMetrics.PercentChange
	insights.Goal = tmpGlobal

	for key, valMap := range cpiInfo.FeatureMetrics {
		var tmp ValueWithDetails
		var temp BaseTargetMetrics

		keyNameType := strings.SplitN(key, "#", 2)
		keyType, keyName := keyNameType[0], keyNameType[1]

		for val, diff := range valMap {
			if val == "" { // omitting "" values
				continue
			}
			if KeyMapForDistribution[val] {
				continue
			}
			if BlackListedKeys[keyName] {
				continue
			}
			KeyMapForDistribution[val] = true

			if _, ok := scaleInfo.FeatureMetrics[key][val]; !ok {
				scaleInfo.FeatureMetrics[key] = make(map[string]DiffMetric)
			}
			featScaleW1 := scaleInfo.FeatureMetrics[key][val].First.(float64)
			featScaleW2 := scaleInfo.FeatureMetrics[key][val].Second.(float64)

			featW1 := diff.First.(float64)
			featW2 := diff.Second.(float64)
			tmp.Key = keyName
			tmp.Value = val
			tmp.Entity = keyType
			temp.W1 = featW1
			temp.W2 = featW2
			temp.Per = diff.PercentChange

			if _, exists := file.JSDivergence.Target[key][val]; exists {
				temp.JSDivergence = file.JSDivergence.Target[key][val] * temp.W1
				if _, exists := PriorityKeysDistribution[tmp.Key]; exists {
					temp.JSDivergence = 4
				} else {
					if WhiteListedKeys[tmp.Key] {
						temp.JSDivergence = 3
						tmp.VoteStatus = Upvoted
					} else if WhiteListedKeysOtherQuery[tmp.Key] {
						temp.JSDivergence = 2
						tmp.VoteStatus = UpvotedForOtherQuery
					} else if DecreaseBoostKeys[tmp.Key] {
						temp.JSDivergence = 0
						tmp.VoteStatus = DownvotedForOtherQuery
					} else {
						temp.JSDivergence = 1
					}
				}
			}
			tmp.ActualValues = temp

			tmp.ChangeInScale.W1[0] = featScaleW1
			if globalScaleW1 != 0 {
				tmp.ChangeInScale.W1[1] = featScaleW1 * 100 / globalScaleW1
			}
			tmp.ChangeInScale.W2[0] = featScaleW2
			if globalScaleW2 != 0 {
				tmp.ChangeInScale.W2[1] = featScaleW2 * 100 / globalScaleW2
			}
			tmp.ChangeInScale.IsIncreased[0] = tmp.ChangeInScale.W1[0] < tmp.ChangeInScale.W2[0]
			tmp.ChangeInScale.IsIncreased[1] = tmp.ChangeInScale.W1[1] < tmp.ChangeInScale.W2[1]
			if tmp.ChangeInScale.W1[0] != 0 {
				tmp.ChangeInScale.Percentage[0] = (tmp.ChangeInScale.W2[0] - tmp.ChangeInScale.W1[0]) * 100 / tmp.ChangeInScale.W1[0]
			}
			tmp.ChangeInScale.Percentage[1] = tmp.ChangeInScale.W2[1] - tmp.ChangeInScale.W1[1]

			tmp.Type = "distribution"
			if file.Category == "" { //for older models built without category
				tmp.Category = "kpi_events"
			} else {
				tmp.Category = "kpi_" + file.Category
			}

			if !(CheckPercentageChange(globalScaleW1, featScaleW1) || CheckPercentageChange(globalScaleW2, featScaleW2)) && math.Abs(tmp.ActualValues.Per) > 5 {
				valWithDetailsArr = append(valWithDetailsArr, tmp)
			}
		}
	}
	sort.Slice(valWithDetailsArr, func(i, j int) bool {
		if valWithDetailsArr[i].ActualValues.JSDivergence == valWithDetailsArr[j].ActualValues.JSDivergence {
			return getInsightImportance(valWithDetailsArr[i]) > getInsightImportance(valWithDetailsArr[j])
		}
		return valWithDetailsArr[i].ActualValues.JSDivergence > valWithDetailsArr[j].ActualValues.JSDivergence
	})

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

	keysUsedInInsights := make(map[string]bool)
	for _, data := range valWithDetailsArr {
		if data.Category == "kpi_campaign" && keysUsedInInsights[data.Key] {
			continue
		}
		keysUsedInInsights[data.Key] = true
		var tempActualValue = ActualMetrics{
			ActualValues: Base{
				W1:          data.ActualValues.W1,
				W2:          data.ActualValues.W2,
				IsIncreased: data.ActualValues.Per > 0,
				Percentage:  data.ActualValues.Per,
			},
		}
		tempActualValue.ChangeInScale = data.ChangeInScale
		tempActualValue.Category = data.Category
		propertyMap[data.Key] = true
		tempActualValue.Key = data.Key
		tempActualValue.Value = data.Value
		tempActualValue.Entity = data.Entity
		tempActualValue.VoteStatus = data.VoteStatus
		tempActualValue.Type = data.Type
		if tempActualValue.ActualValues.IsIncreased && increasedRecords > 0 {
			ActualValuearr = append(ActualValuearr, tempActualValue)
			increasedRecords -= 1
		} else if !tempActualValue.ActualValues.IsIncreased && decreasedRecords > 0 {
			ActualValuearr = append(ActualValuearr, tempActualValue)
			decreasedRecords -= 1
		}
	}

	insights.Insights = append(insights.Insights, ActualValuearr...)
	return insights
}

func GetInsights(file CrossPeriodInsights, numberOfRecords int, QueryClass, EventType string, isEventWebsite bool) WeeklyInsights {
	var KeyMapForConversion = make(map[string]bool)
	var KeyMapForDistribution = make(map[string]bool)
	propertyMap = make(map[string]bool)
	var insights WeeklyInsights
	insights.Insights = make([]ActualMetrics, 0)
	ZeroFlag := true // flag to check if overall W1||W2 is 0.
	if EventType == Funnel || EventType == WebsiteEvent {
		insights.InsightsType = "ConvAndDist"
		if EventType == WebsiteEvent {
			insights.BaseLine = "$session"
		}
	} else if EventType == CRM || EventType == KPI {
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
		if insights.Base.W1 != float64(0) {
			insights.Conv.W1 = insights.Goal.W1 / insights.Base.W1 * 100
		}
		if insights.Base.W2 != float64(0) {
			insights.Conv.W2 = insights.Goal.W2 / insights.Base.W2 * 100
		}
		insights.Conv.IsIncreased = insights.Conv.W1 < insights.Conv.W2
		if insights.Conv.W1 != float64(0) {
			insights.Conv.Percentage = ((insights.Conv.W2 - insights.Conv.W1) / insights.Conv.W1) * 100
		}
	}
	var valWithDetailsArr []ValueWithDetails
	if insights.Goal.W1 == float64(0) || insights.Goal.W2 == float64(0) {
		ZeroFlag = false
	}
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
					if BlackListedKeys[value.Key] {
						continue
					}
					if file.BaseAndTarget.FeatureMetrics[keys][keys2].First != nil {
						temp.W1 = file.BaseAndTarget.FeatureMetrics[keys][keys2].First.(float64)
						if temp.W1 == float64(0) && ZeroFlag {
							continue
						}
					} else {
						continue
					}
					if file.BaseAndTarget.FeatureMetrics[keys][keys2].Second != nil {
						temp.W2 = file.BaseAndTarget.FeatureMetrics[keys][keys2].Second.(float64)
						if temp.W2 == float64(0) && ZeroFlag {
							continue
						}
					} else {
						continue
					}
					if _, exists := file.BaseAndTarget.FeatureMetrics[keys][keys2]; exists {
						temp.Per = file.BaseAndTarget.FeatureMetrics[keys][keys2].PercentChange
					}
					if _, exists := file.DeltaRatio[keys][keys2]; exists {
						temp.DeltaRatio = file.DeltaRatio[keys][keys2] * temp.W1
						if factor, exists := PriorityKeysConversion[value.Key]; exists {
							temp.DeltaRatio *= factor
						} else {
							if WhiteListedKeys[value.Key] {
								temp.DeltaRatio *= 2 // boosting the sorting factor if upvoted
								value.VoteStatus = Upvoted
							} else if WhiteListedKeysOtherQuery[value.Key] {
								temp.DeltaRatio *= 2
								value.VoteStatus = UpvotedForOtherQuery
							} else if DecreaseBoostKeys[value.Key] {
								temp.DeltaRatio *= 0.5 // reverse
								value.VoteStatus = DownvotedForOtherQuery
							}
						}

					}
					if isEventWebsite {
						if value.Entity == "ep" && !(strings.Contains(value.Key, "hubspot") || strings.Contains(value.Key, "salesforce")) {
							temp.DeltaRatio *= 2
						} else if value.Entity == "up" && !(strings.Contains(value.Key, "hubspot") || strings.Contains(value.Key, "salesforce")) {
							temp.DeltaRatio *= 1.1
						} else {
							temp.DeltaRatio *= 0.8
						}
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
					if value.ChangeInPrevalance.W1 != float64(0) {
						value.ChangeInConversion.W1 = value.ActualValues.W1 / value.ChangeInPrevalance.W1 * 100
					}
					if value.ChangeInPrevalance.W2 != float64(0) {
						value.ChangeInConversion.W2 = value.ActualValues.W2 / value.ChangeInPrevalance.W2 * 100
					}
					if value.ChangeInConversion.W1 != float64(0) {
						value.ChangeInConversion.Percentage = ((value.ChangeInConversion.W2 - value.ChangeInConversion.W1) / value.ChangeInConversion.W1) * 100
					}
					value.ChangeInConversion.IsIncreased = value.ChangeInConversion.Percentage > 0

					value.Type = "conversion"
					value.Category = "events"

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
		var tempActualValue = ActualMetrics{
			ActualValues: Base{
				W1:          data.ActualValues.W1,
				W2:          data.ActualValues.W2,
				IsIncreased: data.ActualValues.Per > 0,
				Percentage:  data.ActualValues.Per,
			},
		}
		propertyMap[data.Key] = true
		tempActualValue.Category = data.Category
		tempActualValue.Key = data.Key
		tempActualValue.Value = data.Value
		tempActualValue.Entity = data.Entity
		tempActualValue.VoteStatus = data.VoteStatus
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
					if BlackListedKeys[val2.Key] {
						continue
					}
					if file.Target.FeatureMetrics[keys][keys2].First != nil {
						temp.W1 = file.Target.FeatureMetrics[keys][keys2].First.(float64)
						if (temp.W1 == float64(0) && ZeroFlag) || CheckPercentageChange(insights.Goal.W1, temp.W1) {
							continue
						}
					} else {
						continue
					}
					if file.Target.FeatureMetrics[keys][keys2].Second != nil {
						temp.W2 = file.Target.FeatureMetrics[keys][keys2].Second.(float64)
						if temp.W2 == float64(0) && ZeroFlag {
							continue
						}
					} else {
						continue
					}
					if _, exists := file.Target.FeatureMetrics[keys][keys2]; exists {
						temp.Per = file.Target.FeatureMetrics[keys][keys2].PercentChange
					}
					if _, exists := file.JSDivergence.Target[keys][keys2]; exists {
						temp.JSDivergence = file.JSDivergence.Target[keys][keys2] * temp.W1
						if factor, exists := PriorityKeysDistribution[val2.Key]; exists {
							temp.JSDivergence *= factor
						} else {
							if WhiteListedKeys[val2.Key] {
								temp.JSDivergence *= 2 // boosting 2X
								val2.VoteStatus = Upvoted
							} else if WhiteListedKeysOtherQuery[val2.Key] {
								temp.JSDivergence *= 2
								val2.VoteStatus = UpvotedForOtherQuery
							} else if DecreaseBoostKeys[val2.Key] {
								temp.JSDivergence *= 0.5
								val2.VoteStatus = DownvotedForOtherQuery
							}
						}

					}
					val2.ActualValues = temp
					if insights.Goal.W1 != float64(0) {
						val2.ChangeInDistribution.W1 = (val2.ActualValues.W1 / insights.Goal.W1) * 100
					}
					if insights.Goal.W2 != float64(0) {
						val2.ChangeInDistribution.W2 = (val2.ActualValues.W2 / insights.Goal.W2) * 100
					}
					val2.ChangeInDistribution.IsIncreased = val2.ChangeInDistribution.W1 < val2.ChangeInDistribution.W2
					val2.ChangeInDistribution.Percentage = (val2.ChangeInDistribution.W1) - (val2.ChangeInDistribution.W2)

					val2.Type = "distribution"
					val2.Category = "events"
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
					if BlackListedKeys[val2.Key] {
						continue
					}
					if file.BaseAndTarget.FeatureMetrics[keys][keys2].First != nil {
						temp.W1 = file.BaseAndTarget.FeatureMetrics[keys][keys2].First.(float64)
						if (temp.W1 == float64(0) && ZeroFlag) || CheckPercentageChange(insights.Goal.W1, temp.W1) {
							continue
						}
					} else {
						continue
					}
					if file.BaseAndTarget.FeatureMetrics[keys][keys2].Second != nil {
						temp.W2 = file.BaseAndTarget.FeatureMetrics[keys][keys2].Second.(float64)
						if temp.W2 == float64(0) && ZeroFlag {
							continue
						}
					} else {
						continue
					}
					if _, exists := file.BaseAndTarget.FeatureMetrics[keys][keys2]; exists {
						temp.Per = file.BaseAndTarget.FeatureMetrics[keys][keys2].PercentChange
					}
					if _, exists := file.JSDivergence.Target[keys][keys2]; exists {
						temp.JSDivergence = file.JSDivergence.Target[keys][keys2] * temp.W1
						if factor, exists := PriorityKeysDistribution[val2.Key]; exists {
							temp.JSDivergence *= factor
						} else {
							if WhiteListedKeys[val2.Key] {
								temp.JSDivergence *= 2
								val2.VoteStatus = Upvoted
							} else if WhiteListedKeysOtherQuery[val2.Key] {
								temp.JSDivergence *= 2
								val2.VoteStatus = UpvotedForOtherQuery
							} else if DecreaseBoostKeys[val2.Key] {
								temp.JSDivergence *= 0.5
								val2.VoteStatus = DownvotedForOtherQuery
							}
						}

					}
					val2.ActualValues = temp
					if insights.Goal.W1 != float64(0) {
						val2.ChangeInDistribution.W1 = (val2.ActualValues.W1 / insights.Goal.W1) * 100
					}
					if insights.Goal.W2 != float64(0) {
						val2.ChangeInDistribution.W2 = (val2.ActualValues.W2 / insights.Goal.W2) * 100
					}
					val2.ChangeInDistribution.IsIncreased = val2.ChangeInDistribution.W1 < val2.ChangeInDistribution.W2
					val2.ChangeInDistribution.Percentage = (val2.ChangeInDistribution.W1) - (val2.ChangeInDistribution.W2)

					val2.Type = "distribution"
					val2.Category = "events"
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
		var tempActualValue = ActualMetrics{
			ActualValues: Base{
				W1:          data.ActualValues.W1,
				W2:          data.ActualValues.W2,
				IsIncreased: data.ActualValues.Per > 0,
				Percentage:  data.ActualValues.Per,
			},
		}
		tempActualValue.ChangeInDistribution = data.ChangeInDistribution
		tempActualValue.Category = data.Category
		propertyMap[data.Key] = true
		tempActualValue.Key = data.Key
		tempActualValue.Value = data.Value
		tempActualValue.Entity = data.Entity
		tempActualValue.VoteStatus = data.VoteStatus
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

func GetWeeklyInsights(projectId uint64, agentUUID string, queryId uint64, baseStartTime *time.Time, compStartTime *time.Time, insightsType string, numberOfRecords int) (interface{}, error) {
	k := make(map[uint64]int)
	k[399] = 100
	k[594] = 100
	k[559] = 100
	k[628] = 100
	k[616] = 100
	k[519] = 100
	kValue, ok := k[projectId]
	if !ok {
		kValue = 100
	}
	path, file := C.GetCloudManager().GetInsightsCpiFilePathAndName(projectId, U.GetDateOnlyFromTimestampZ(baseStartTime.Unix()), queryId, kValue)
	fmt.Println("path/file:", path, file)
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
	var isEventOccurence bool
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
		isEventOccurence = (query.Type == model.QueryTypeEventsOccurrence)
	} else if class == model.QueryClassKPI {
		var KpiQueryGroup model.KPIQueryGroup
		err = U.DecodePostgresJsonbToStructType(&QueriesObj.Query, &KpiQueryGroup)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	var EventType string
	var isEventWebsite bool
	EventType = getEventType(&query, class, projectId)
	if EventType == Funnel || EventType == WebsiteEvent {
		NewEventType := GetEventTypeForFunnelOrWebsite(&query, projectId)
		if NewEventType == WebsiteEvent {
			isEventWebsite = true
		}
	}
	if isEventOccurence {
		EventType = CRM
	}

	var insights CrossPeriodInsights
	var insightsKpiList []CrossPeriodInsightsKpi
	var insightsKpi CrossPeriodInsightsKpi
	if class == model.QueryClassKPI {
		err = json.Unmarshal(data, &insightsKpiList)
		if len(insightsKpiList) > 0 {
			insightsKpi = insightsKpiList[0]
		}
	} else {
		err = json.Unmarshal(data, &insights)
	}
	if err != nil {
		log.WithError(err).Error("error unmarshalling response")
		return nil, err
	}
	var insightsObj WeeklyInsights
	PriorityKeysDistribution = GetPriorityKeysMapDistribution()
	WhiteListedKeys = make(map[string]bool)
	WhiteListedKeysOtherQuery = make(map[string]bool)
	BlackListedKeys = GetBlackListedKeys()
	PriorityKeysConversion = GetPriorityKeysMapConversion()
	CaptureBlackListedAndWhiteListedKeys(projectId, agentUUID, queryId)
	if class == model.QueryClassKPI {
		insightsObj = GetInsightsKpi(insightsKpi, numberOfRecords, class, EventType, isEventWebsite)
	} else {
		insightsObj = GetInsights(insights, numberOfRecords, class, EventType, isEventWebsite)
		// adding query groups
		gbpInsights := addGroupByProperties(query, EventType, insights, insightsObj, isEventWebsite)
		// appending at top
		insightsObj.Insights = append(gbpInsights, insightsObj.Insights...)
	}

	removeNegativePercentageFromInsights(&insightsObj)
	insightsObj.QueryId = queryId
	return insightsObj, nil
}

func addGroupByProperties(query model.Query, EventType string, file CrossPeriodInsights, insights WeeklyInsights, isEventWebsite bool) []ActualMetrics {
	ActualMetricsArr := make([]ActualMetrics, 0)
	if EventType == KPI {
		return ActualMetricsArr
	}

	ZeroFlag := true
	if insights.Goal.W1 == float64(0) || insights.Goal.W2 == float64(0) {
		ZeroFlag = false
	}
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
			if EventType == Funnel || EventType == WebsiteEvent {
				for _, property := range properties {
					for values := range file.BaseAndTarget.FeatureMetrics[property] { // conversion
						var newData ValueWithDetails
						var temp BaseTargetMetrics
						newData.Key = gbp.Property
						newData.Value = values
						newData.Entity = gbp.Entity
						if BlackListedKeys[newData.Key] {
							continue
						}
						if file.BaseAndTarget.FeatureMetrics[property][values].First != nil {
							temp.W1 = file.BaseAndTarget.FeatureMetrics[property][values].First.(float64)
							if temp.W1 == float64(0) && ZeroFlag {
								continue
							}
						} else {
							continue
						}
						if file.BaseAndTarget.FeatureMetrics[property][values].Second != nil {
							temp.W2 = file.BaseAndTarget.FeatureMetrics[property][values].Second.(float64)
							if temp.W2 == float64(0) && ZeroFlag {
								continue
							}
						} else {
							continue
						}
						if _, exists := file.BaseAndTarget.FeatureMetrics[property][values]; exists {
							temp.Per = file.BaseAndTarget.FeatureMetrics[property][values].PercentChange
						}
						if _, exists := file.DeltaRatio[property][values]; exists {
							temp.DeltaRatio = file.DeltaRatio[property][values] * temp.W1
							if factor, exists := PriorityKeysConversion[property]; exists {
								temp.DeltaRatio *= factor
							} else {
								if WhiteListedKeys[property] {
									temp.DeltaRatio *= 2
									newData.VoteStatus = Upvoted
								} else if WhiteListedKeysOtherQuery[property] {
									temp.DeltaRatio *= 2
									newData.VoteStatus = UpvotedForOtherQuery
								} else if DecreaseBoostKeys[property] {
									temp.DeltaRatio *= 0.5
									newData.VoteStatus = DownvotedForOtherQuery
								}
							}

						}
						if isEventWebsite {
							if newData.Entity == model.PropertyEntityEvent && !(strings.Contains(newData.Key, "hubspot") || strings.Contains(newData.Key, "salesforce")) {
								temp.DeltaRatio *= 2
							} else if newData.Entity == model.PropertyEntityUser && !(strings.Contains(newData.Key, "hubspot") || strings.Contains(newData.Key, "salesforce")) {
								temp.DeltaRatio *= 1.1
							} else {
								temp.DeltaRatio *= 0.8
							}
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
						if newData.ChangeInPrevalance.W1 != float64(0) {
							newData.ChangeInConversion.W1 = newData.ActualValues.W1 / newData.ChangeInPrevalance.W1 * 100
						}
						if newData.ChangeInPrevalance.W2 != float64(0) {
							newData.ChangeInConversion.W2 = newData.ActualValues.W2 / newData.ChangeInPrevalance.W2 * 100
						}
						if newData.ChangeInConversion.W1 != float64(0) {
							newData.ChangeInConversion.Percentage = ((newData.ChangeInConversion.W2 - newData.ChangeInConversion.W1) / newData.ChangeInConversion.W1) * 100
						}
						newData.ChangeInConversion.IsIncreased = newData.ChangeInConversion.Percentage > 0
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
							var tempActualValue = ActualMetrics{
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
							tempActualValue.VoteStatus = data.VoteStatus
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
						if BlackListedKeys[newData.Key] {
							continue
						}
						if file.Target.FeatureMetrics[property][values].First != nil {
							temp.W1 = file.Target.FeatureMetrics[property][values].First.(float64)
							if (temp.W1 == float64(0) && ZeroFlag) || CheckPercentageChange(insights.Goal.W1, temp.W1) {
								continue
							}
						} else {
							continue
						}
						if file.Target.FeatureMetrics[property][values].Second != nil {
							temp.W2 = file.Target.FeatureMetrics[property][values].Second.(float64)
							if temp.W2 == float64(0) && ZeroFlag {
								continue
							}
						} else {
							continue
						}
						if _, exists := file.Target.FeatureMetrics[property][values]; exists {
							temp.Per = file.Target.FeatureMetrics[property][values].PercentChange
						}
						if _, exists := file.JSDivergence.Target[property][values]; exists {
							temp.JSDivergence = file.JSDivergence.Target[property][values] * temp.W1
							if factor, exists := PriorityKeysDistribution[property]; exists {
								temp.JSDivergence *= factor
							} else {
								if WhiteListedKeys[property] {
									temp.JSDivergence *= 2
									newData.VoteStatus = Upvoted
								} else if WhiteListedKeysOtherQuery[property] {
									temp.JSDivergence *= 2
									newData.VoteStatus = UpvotedForOtherQuery
								} else if DecreaseBoostKeys[property] {
									temp.JSDivergence *= 0.5
									newData.VoteStatus = DownvotedForOtherQuery
								}
							}

						}
						newData.ActualValues = temp
						if insights.Goal.W1 != float64(0) {
							newData.ChangeInDistribution.W1 = (newData.ActualValues.W1 / insights.Goal.W1) * 100
						}
						if insights.Goal.W2 != float64(0) {
							newData.ChangeInDistribution.W2 = (newData.ActualValues.W2 / insights.Goal.W2) * 100
						}
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
						if BlackListedKeys[newData.Key] {
							continue
						}
						if file.BaseAndTarget.FeatureMetrics[property][values].First != nil {
							temp.W1 = file.BaseAndTarget.FeatureMetrics[property][values].First.(float64)
							if (temp.W1 == float64(0) && ZeroFlag) || CheckPercentageChange(insights.Goal.W1, temp.W1) {
								continue
							}
						} else {
							continue
						}
						if file.BaseAndTarget.FeatureMetrics[property][values].Second != nil {
							temp.W2 = file.BaseAndTarget.FeatureMetrics[property][values].Second.(float64)
							if temp.W2 == float64(0) && ZeroFlag {
								continue
							}
						} else {
							continue
						}
						if _, exists := file.BaseAndTarget.FeatureMetrics[property][values]; exists {
							temp.Per = file.BaseAndTarget.FeatureMetrics[property][values].PercentChange
						}
						if _, exists := file.JSDivergence.Base[property][values]; exists {
							temp.JSDivergence = file.JSDivergence.Base[property][values]
							if factor, exists := PriorityKeysDistribution[property]; exists {
								temp.JSDivergence *= factor
							} else {
								if WhiteListedKeys[property] {
									temp.JSDivergence *= 2
									newData.VoteStatus = Upvoted
								} else if WhiteListedKeysOtherQuery[property] {
									temp.JSDivergence *= 2
									newData.VoteStatus = UpvotedForOtherQuery
								} else if DecreaseBoostKeys[property] {
									temp.JSDivergence *= 0.5
									newData.VoteStatus = DownvotedForOtherQuery
								}
							}

						}
						newData.ActualValues = temp

						if insights.Goal.W1 != float64(0) {
							newData.ChangeInDistribution.W1 = (newData.ActualValues.W1 / insights.Goal.W1) * 100
						}
						if insights.Goal.W2 != float64(0) {
							newData.ChangeInDistribution.W2 = (newData.ActualValues.W2 / insights.Goal.W2) * 100
						}
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
				var tempActualValue = ActualMetrics{
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
				tempActualValue.VoteStatus = data.VoteStatus
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
	} else if QueryClass == model.QueryClassKPI {
		return KPI
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

func GetEventTypeForFunnelOrWebsite(query *model.Query, project_id uint64) string {
	EventType := ""
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
			break
		} else {
			EventType = WebsiteEvent
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
		insightsObj.Insights[index].ChangeInConversion.Percentage = math.Abs(insightsObj.Insights[index].ChangeInConversion.Percentage)
		insightsObj.Insights[index].ChangeInPrevalance.Percentage = math.Abs(insightsObj.Insights[index].ChangeInPrevalance.Percentage)
		insightsObj.Insights[index].ChangeInDistribution.Percentage = math.Abs(insightsObj.Insights[index].ChangeInDistribution.Percentage)
	}
}

func CaptureBlackListedAndWhiteListedKeys(projectID uint64, agentUUID string, queryID uint64) {
	records, err := store.GetStore().GetRecordsFromFeedback(projectID, agentUUID)
	if err != nil {
		log.Error(err)
	}
	DecreaseBoostKeys = make(map[string]bool)

	for _, record := range records {
		bytes, err := json.Marshal(record.Property)
		if err != nil {
			log.Error(err)
			continue
		}
		var property model.WeeklyInsightsProperty
		json.Unmarshal(bytes, &property)
		if record.VoteType == model.VOTE_TYPE_UPVOTE { // upvote
			if property.QueryID == queryID {
				WhiteListedKeys[property.Key] = true
			} else {
				WhiteListedKeysOtherQuery[property.Key] = true
			}

		} else { // downvote
			if property.QueryID == queryID {
				BlackListedKeys[property.Key] = true
			} else {
				DecreaseBoostKeys[property.Key] = true
			}
		}

	}
}

func CheckPercentageChange(overall, week float64) bool {
	//filtering  if week1 data is less than x % of overall w1 data
	actual := (DistributionChangePer / 100) * overall
	return week < actual
}
