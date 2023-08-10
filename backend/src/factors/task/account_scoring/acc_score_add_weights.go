package task

import (
	"encoding/json"
	"factors/model/model"
	M "factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	U "factors/util"

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

	if rules, ok := mweights[M.DEFAULT_EVENT]; ok {
		for _, individual_rule := range rules {
			ruleId, match := ProcessSingleFilterOnEvent(event, individual_rule)
			if match {
				ids = append(ids, ruleId)
			}
		}
	}

	return ids
}

func ProcessSingleFilterOnEvent(event *P.CounterEventFormat, filter M.AccEventWeight) (string, bool) {

	ruleId := filter.WeightId
	ruleResultMap := make(map[int]bool, 0)

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
		if !resVal {
			return "", false
		}
	}

	return ruleId, true
}

func evalProperty(properties map[string]interface{}, propFilter M.WeightKeyValueTuple) bool {

	propFilterMap := make(map[string]bool, 0)
	for _, propertyVal := range propFilter.Value {
		propFilterMap[propertyVal] = true
	}
	propvalKey := properties[propFilter.Key]

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
		if _, ok := propFilterMap[propval]; ok {
			return true
		}
	} else if propFilter.Operator == model.NotContainsOpStr {
		propval := U.GetPropertyValueAsString(propvalKey)
		if _, ok := propFilterMap[propval]; !ok {
			return true
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
