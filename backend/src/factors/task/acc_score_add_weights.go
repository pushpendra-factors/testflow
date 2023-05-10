package task

import (
	"encoding/json"
	"factors/model/model"
	M "factors/model/model"
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
	Ename string                `json:"en"`
	Rule  M.WeightKeyValueTuple `json:"rule"`
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

	// currently not adding
	if rules, ok := mweights[event_name]; ok {

		for _, individual_rule := range rules {

			r := individual_rule.Rule

			if r.Type == "event" {
				id_, match := evalProperty(event.EventProperties, individual_rule)
				if match {
					ids = append(ids, id_)
				}
			} else if r.Type == "user" {

				id_, match := evalProperty(event.UserProperties, individual_rule)
				if match {
					ids = append(ids, id_)
				}

			}

			if r.Key == "" {
				ids = append(ids, individual_rule.WeightId)
			}

		}

	}

	if rules, ok := mweights[M.DEFAULT_EVENT]; ok {
		for _, individual_rule := range rules {

			r := individual_rule.Rule

			if r.Type == "event" {
				id_, match := evalProperty(event.EventProperties, individual_rule)
				if match {
					ids = append(ids, id_)
				}
			} else if r.Type == "user" {

				id_, match := evalProperty(event.UserProperties, individual_rule)
				if match {
					ids = append(ids, id_)
				}

			}

			if r.Key == "" {
				ids = append(ids, individual_rule.WeightId)
			}

		}

	}

	return ids
}

func evalProperty(properties map[string]interface{}, rule M.AccEventWeight) (string, bool) {

	r := rule.Rule
	propval_ := properties[r.Key]
	propval := U.GetPropertyValueAsString(propval_)
	if r.Operator {
		if strings.Compare(propval, r.Value) == 0 {
			return rule.WeightId, true
		}
	} else {
		if strings.Compare(propval, r.Value) != 0 {
			return rule.WeightId, true
		}
	}

	return "", false
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
