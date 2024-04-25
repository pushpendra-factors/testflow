package accountscoring

import (
	"encoding/json"
	"factors/model/model"
	M "factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"

	log "github.com/sirupsen/logrus"

	"strings"
)

// weights :
// create : create a new rule
// update : check if duplicate rules exist and is deleted
// delete : soft delete rule
type weightfilter struct {
	Ename string                  `json:"en"`
	Rule  []M.WeightKeyValueTuple `json:"rule"`
}

const HASHKEYLENGTH = 12

func DeduplicateWeights(weights M.AccWeights) (M.AccWeights, error) {

	var updatedWeights M.AccWeights
	var weightrulelist []M.AccEventWeight = make([]M.AccEventWeight, 0)
	weightIdMap := make(map[string]M.AccEventWeight)
	wt := weights.WeightConfig
	var err error
	for _, wid := range wt {
		var w_id_hash string
		// if event name is empty, set it to all_events
		// filter will be applied to all events
		if wid.EventName == "" {
			wid.EventName = model.DEFAULT_EVENT
		}
		w_val := weightfilter{wid.EventName, wid.Rule}
		w_id_hash = wid.WeightId
		if w_id_hash == "" {
			w_id_hash, err = GenerateHashFromStruct(w_val)
			if err != nil {
				return M.AccWeights{}, err
			}
			log.Debugf("generated hash :%s", w_id_hash)
		}

		if _, ok := weightIdMap[w_id_hash]; !ok {
			wid.WeightId = w_id_hash
			weightIdMap[w_id_hash] = wid
			weightrulelist = append(weightrulelist, wid)
		} else {
			p := weightIdMap[w_id_hash]
			p.Is_deleted = wid.Is_deleted
			log.Debugf("Duplicate detected, not adding to list:%s", w_id_hash)
		}
	}

	updatedWeights.WeightConfig = weightrulelist
	updatedWeights.SaleWindow = weights.SaleWindow
	return updatedWeights, nil
}

func FilterEvents(event *P.CounterEventFormat, mweights map[string][]M.AccEventWeight) []string {
	//basically a rule engine
	// for each rule matched output the rule id

	var ids []string = make([]string, 0)

	event_name := event.EventName

	if mrules, ok := mweights[event_name]; ok {
		for _, individual_rule := range mrules {
			ruleId, match := ProcessSingleFilterOnEvent(event, individual_rule)
			if match {
				ids = append(ids, ruleId)
			}
		}

	}

	ids_special_filters := SpecialFilters(mweights, event)
	ids = append(ids, ids_special_filters...)

	return ids
}

func ProcessSingleFilterOnEvent(event *P.CounterEventFormat, filter M.AccEventWeight) (string, bool) {

	ruleId := filter.WeightId
	ruleResultMap := make(map[int]bool, 0)

	if len(filter.Rule) == 0 {
		if event.EventName == filter.EventName {
			return filter.WeightId, true
		}
	}

	for idx, singleRule := range filter.Rule {
		if singleRule.Type == "event" {
			match := evalProperty(event.EventProperties, singleRule)
			ruleResultMap[idx] = match
		} else if singleRule.Type == "user" {
			match := evalProperty(event.UserProperties, singleRule)
			ruleResultMap[idx] = match
		}
	}

	for _, resVal := range ruleResultMap {
		if resVal {
			return ruleId, true
		}
	}

	return "", false
}

func evalProperty(properties map[string]interface{}, propFilter M.WeightKeyValueTuple) bool {

	propFilterMap := make(map[string]bool, 0)
	for _, propertyVal := range propFilter.Value {
		propFilterMap[propertyVal] = true
	}
	propvalKey, prop_ok := properties[propFilter.Key]

	if propFilter.Operator == model.IsKnown {
		return prop_ok
	} else if propFilter.Operator == model.IsUnKnown {
		return !prop_ok
	}

	if propFilter.ValueType == U.PropertyTypeCategorical {

		if propFilter.Operator == model.EqualsOpStr {
			propval := U.GetPropertyValueAsString(propvalKey)
			filterval := propFilter.Value[0]
			if strings.Compare(propval, filterval) == 0 {
				return true
			}
		} else if propFilter.Operator == model.NotEqualOpStr {
			propval := U.GetPropertyValueAsString(propvalKey)
			filterval := propFilter.Value[0]
			if strings.Compare(propval, filterval) != 0 {
				return true
			}
		} else if propFilter.Operator == model.ContainsOpStr {
			propval := U.GetPropertyValueAsString(propvalKey)
			for propFilterKey, _ := range propFilterMap {
				if strings.Contains(propval, propFilterKey) == true {
					return true
				}
			}
		} else if propFilter.Operator == model.NotContainsOpStr {
			propval := U.GetPropertyValueAsString(propvalKey)
			for propFilterKey, _ := range propFilterMap {
				if strings.Contains(propval, propFilterKey) == false {
					return true
				}
			}
		} else if propFilter.Operator == model.InList {
			propval := U.GetPropertyValueAsString(propvalKey)
			_, inListok := propFilterMap[propval]
			return inListok
		} else if propFilter.Operator == model.NotInList {
			propval := U.GetPropertyValueAsString(propvalKey)
			_, inListok := propFilterMap[propval]
			return !inListok
		}

	} else if propFilter.ValueType == U.PropertyTypeNumerical {

		if propFilter.Operator == model.EqualsOp {
			propval, err := U.GetPropertyValueAsFloat64(propvalKey)
			if err != nil {
				log.WithField("key", propvalKey).Error("Unable to decode property key")
				return false
			}
			filterVal := propFilter.LowerBound
			return propval == filterVal
		} else if propFilter.Operator == model.NotEqualOp {
			propval, err := U.GetPropertyValueAsFloat64(propvalKey)
			if err != nil {
				log.WithField("key", propvalKey).Error("Unable to decode property key")
				return false
			}
			filterVal := propFilter.LowerBound
			return propval != filterVal
		} else if propFilter.Operator == model.GreaterThanOrEqualOpStr {
			propval, err := U.GetPropertyValueAsFloat64(propvalKey)
			if err != nil {
				log.WithField("key", propvalKey).Error("Unable to decode property key")
				return false
			}
			filterVal := propFilter.LowerBound
			return propval >= filterVal
		} else if propFilter.Operator == model.GreaterThanOpStr {
			propval, err := U.GetPropertyValueAsFloat64(propvalKey)
			if err != nil {
				log.WithField("key", propvalKey).Error("Unable to decode property key")
				return false
			}
			filterVal := propFilter.LowerBound
			return propval > filterVal
		} else if propFilter.Operator == model.LesserThanOrEqualOpStr {
			propval, err := U.GetPropertyValueAsFloat64(propvalKey)
			if err != nil {
				log.WithField("key", propvalKey).Error("Unable to decode property key")
				return false
			}
			filterVal := propFilter.LowerBound
			return propval <= filterVal
		} else if propFilter.Operator == model.LesserThanOpStr {
			propval, err := U.GetPropertyValueAsFloat64(propvalKey)
			if err != nil {
				log.WithField("key", propvalKey).Error("Unable to decode property key")
				return false
			}
			filterVal := propFilter.LowerBound
			return propval < filterVal
		} else if propFilter.Operator == model.BetweenStr {
			propval, err := U.GetPropertyValueAsFloat64(propvalKey)
			if err != nil {
				log.WithField("key", propvalKey).Error("Unable to decode property key")
				return false
			}
			if propval >= propFilter.LowerBound && propval <= propFilter.UpperBound {
				return true
			}
		} else if propFilter.Operator == model.NotInBetweenStr {
			propval, err := U.GetPropertyValueAsFloat64(propvalKey)
			if err != nil {
				log.WithField("key", propvalKey).Error("Unable to decode property key")
				return false
			}
			if !(propval >= propFilter.LowerBound && propval <= propFilter.UpperBound) {
				return true
			}
		}

	} else if propFilter.ValueType == U.PropertyTypeDateTime {
		// expecting all date time values in int64

		if propFilter.Operator == model.BetweenStr {

			propval, err := U.GetPropertyValueAsInt64(propvalKey)
			if err != nil {
				log.WithField("key", propvalKey).Error("Unable to decode property key")
				return false
			}
			if propval >= int64(propFilter.LowerBound) && propval <= int64(propFilter.UpperBound) {
				return true
			}
			return false
		} else if propFilter.Operator == model.NotInBetweenStr {

			propval, err := U.GetPropertyValueAsInt64(propvalKey)
			if err != nil {
				log.WithField("key", propvalKey).Error("Unable to decode property key")
				return false
			}
			if !(propval >= int64(propFilter.LowerBound) && propval <= int64(propFilter.UpperBound)) {
				return true
			}
			return false
		}

	}

	return false
}

func GenerateHashFromStruct(w weightfilter) (string, error) {
	queryCacheBytes, err := json.Marshal(w)
	if err != nil {
		return "", err
	}

	hstringLowerCase := strings.ToLower(string(queryCacheBytes))
	hstring, err := GenerateHash(hstringLowerCase)
	if err != nil {
		return "", err
	}
	return hstring, err

}
func GenerateHash(bytes string) (string, error) {
	uuid := U.HashKeyUsingSha256Checksum(string(bytes))[:HASHKEYLENGTH]
	return uuid, nil
}

func OnOnboardingAccScoring(projectId int64) error {
	weights := model.GetDefaultAccScoringWeights()
	err := store.GetStore().UpdateAccScoreWeights(projectId, weights)
	if err != nil {
		log.Errorf("Unable to add default weights")
		return err
	}
	return nil
}

func ValidateAndUpdateEngagementLevel(projectId int64, buckets M.BucketRanges) (bool, int, error) {

	status, err := ValidateEngagementLevel(projectId, buckets)
	if err != nil {
		return false, http.StatusBadRequest, err
	}

	if status {
		err = store.GetStore().UpdateEngagementLevel(projectId, buckets)
		if err != nil {
			return false, http.StatusInternalServerError, err
		}
	}

	return true, http.StatusCreated, nil
}

func ValidateEngagementLevel(projectId int64, engagementLevel M.BucketRanges) (bool, error) {
	//check of string are repeated
	//check if ranges are overlapping

	engagementStringMap := make(map[string]int)

	for _, bucket := range engagementLevel.Ranges {
		if _, ok := engagementStringMap[bucket.Name]; !ok {
			engagementStringMap[bucket.Name] += 1
		} else {
			err := fmt.Errorf("engagementlevel name - %s  is  repeated", bucket.Name)
			return false, err
		}
	}

	sort.Slice(engagementLevel.Ranges, func(i, j int) bool {
		return engagementLevel.Ranges[i].Low < engagementLevel.Ranges[j].Low
	})

	for idx := 1; idx < len(engagementLevel.Ranges); idx++ {
		lv1 := engagementLevel.Ranges[idx-1].Name
		lv2 := engagementLevel.Ranges[idx].Name
		if engagementLevel.Ranges[idx-1].High > engagementLevel.Ranges[idx].Low {
			err := fmt.Errorf("ranges of %s and %s are overlapping", lv1, lv2)
			return false, err
		}
	}

	return true, nil

}

func SpecialFilters(mweights map[string][]M.AccEventWeight, event *P.CounterEventFormat) []string {

	var ids []string = make([]string, 0)

	if rules, ok := mweights[M.DEFAULT_EVENT]; ok {
		for _, individual_rule := range rules {
			ruleId, match := ProcessSingleFilterOnEvent(event, individual_rule)
			if match {
				ids = append(ids, ruleId)
			}
		}
	}

	if rules, ok := mweights[U.EVENT_NAME_PAGE_VIEW]; ok {
		if event.EventProperties[U.EP_IS_PAGE_VIEW] == true {
			for _, individual_rule := range rules {
				ruleId, match := ProcessSingleFilterOnEvent(event, individual_rule)
				if match {
					ids = append(ids, ruleId)
				}
			}
		}
	}

	return ids

}
