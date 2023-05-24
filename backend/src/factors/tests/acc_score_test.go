package tests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	M "factors/model/model"
	P "factors/pattern"
	T "factors/task"

	"factors/model/store"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

// func TestUpdatingCountsinDB(t *testing.T) {

// 	ev := M.EventsCountScore{}
// 	counts1 := make(map[string]int64, 0)
// 	ev.ProjectId = int64(6000003)
// 	ev.UserId = "justin.miller@ivvy.com"
// 	ev.DateStamp = "20230102"
// 	counts1["session4"] = int64(70)
// 	counts1["session3"] = int64(30)
// 	counts1["session2"] = int64(20)
// 	counts1["session1"] = int64(10)
// 	ev.EventScore = counts1
// 	ti, err := store.GetStore().GetUserScore(ev.ProjectId, ev.UserId, ev.DateStamp, true, true)
// 	log.Infof("data : %v", ti)
// 	assert.Nil(t, err)

// }

func TestGettingAccountscoresinDB(t *testing.T) {

	ev := M.EventsCountScore{}
	counts1 := make(map[string]int64, 0)
	project_id := int64(51)
	ev.UserId = "justin.miller@ivvy.com"
	group_id := int(1)
	ts := "20230102"
	counts1["session4"] = int64(70)
	counts1["session3"] = int64(30)
	counts1["session2"] = int64(20)
	counts1["session1"] = int64(10)
	ev.EventScore = counts1
	ti, err := store.GetStore().GetAccountsScore(project_id, group_id, ts, true)
	log.Infof("data : %v", ti)
	assert.Nil(t, err)

}

func TestGettingAllUsersSCores(t *testing.T) {

	project_id := int64(6000003)
	ti, err := store.GetStore().GetAllUserScore(project_id, false)
	log.Infof("data : %v", ti)
	assert.Nil(t, err)

}

func TestDeduplicateWeights(t *testing.T) {

	var finalWeights M.AccWeights

	w0 := M.AccEventWeight{WeightId: "", Weight_value: 5.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w1 := M.AccEventWeight{WeightId: "", Weight_value: 7.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w2 := M.AccEventWeight{WeightId: "", Weight_value: 8.0, Is_deleted: false, EventName: "$session", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w3 := M.AccEventWeight{WeightId: "", Weight_value: 9.0, Is_deleted: false, EventName: "$form_submitted", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w4 := M.AccEventWeight{WeightId: "", Weight_value: 10.0, Is_deleted: false, EventName: "www.acme.com", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w5 := M.AccEventWeight{WeightId: "", Weight_value: 11.0, Is_deleted: false, EventName: "www.acme.com/pricing", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w6 := M.AccEventWeight{WeightId: "1", Weight_value: 5.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "user", ValueType: "categorical"}}}

	weightRules := []M.AccEventWeight{w0, w1, w2, w3, w4, w5, w6}
	finalWeights.WeightConfig = weightRules
	finalWeights.SaleWindow = 10

	cr, err := T.DeduplicateWeights(finalWeights)
	assert.Nil(t, err)
	assert.Equal(t, len(weightRules)-1, len(cr.WeightConfig), "duplicate elements are not removed")

}

func TestFilterAndCountEvents(t *testing.T) {

	e1 := string(`{"uid": "pp5", "ujt": 1651209637, "en": "$session", "et": 165120963, "ecd": 1, "epr": {"$browser": "Chrome", "$channel": "Paid Search", "$city": "Queanbeyan", "$country": "Australia"}, "upr": {}, "g1ui": "t1", "g3ui": "t2", "g3ui": "t3"}`)
	e2 := string(`{"uid": "pp5", "ujt": 1651209637, "en": "$pageview", "et": 165120963, "ecd": 1, "epr": {"$browser": "Chrome",  "$channel": "Paid Search", "$city": "Queanbeyan", "$country": "Australia"}, "upr": {}, "g1ui": "t1", "g3ui": "t2", "g3ui": "t3"}`)
	e3 := string(`{"uid": "pp5", "ujt": 1651209637, "en": "$form_submitted", "et": 165120963, "ecd": 1, "epr": {"$browser": "Chrome", "$channel": "Paid Search", "$city": "Queanbeyan", "$country": "Australia"}, "upr": {}, "g1ui": "t1", "g3ui": "t2", "g3ui": "t3"}`)
	e4 := string(`{"uid": "pp5", "ujt": 1651209637, "en": "$session", "et": 165120963, "ecd": 1, "epr": {"$browser": "Chrome",  "$channel": "Paid Search", "$city": "Queanbeyan", "$country": "Australia"}, "upr": {}, "g1ui": "t1", "g3ui": "t2", "g3ui": "t3"}`)
	e5 := string(`{"uid": "pp5", "ujt": 1651209637, "en": "$session_1", "et": 165120963, "ecd": 1, "epr": {"$browser": "Chrome", "$channel": "Paid Search", "$city": "Queanbeyan", "$country": "Kenya"}, "upr": {}, "g1ui": "t1", "g3ui": "t2", "g3ui": "t3"}`)
	e6 := string(`{"uid": "pp5", "ujt": 1651209637, "en": "$pageview_1", "et": 165120963, "ecd": 1, "epr": {"$browser": "Chrome",  "$channel": "Paid Search", "$city": "Queanbeyan", "$country": "Kenya"}, "upr": {}, "g1ui": "t1", "g3ui": "t2", "g3ui": "t3"}`)
	e7 := string(`{"uid": "pp5", "ujt": 1651209637, "en": "$form_submitted_1", "et": 165120963, "ecd": 1, "epr": {"$browser": "Chrome", "$channel": "Paid Search", "$city": "Queanbeyan", "$country": "Kenya"}, "upr": {}, "g1ui": "t1", "g3ui": "t2", "g3ui": "t3"}`)
	e8 := string(`{"uid": "pp5", "ujt": 1651209637, "en": "$session_3", "et": 165120963, "ecd": 1, "epr": {"$browser": "Chrome",  "$channel": "Paid Search", "$city": "Queanbeyan", "$country": "Brazil"}, "upr": {}, "g1ui": "t1", "g3ui": "t2", "g3ui": "t3"}`)

	ev := []string{e1, e2, e3, e4, e5, e6, e7, e8}
	var events []*P.CounterEventFormat = make([]*P.CounterEventFormat, 0)

	for _, e := range ev {
		var testev *P.CounterEventFormat
		err := json.Unmarshal([]byte(e), &testev)
		assert.Nil(t, err)
		events = append(events, testev)
	}

	var finalWeights M.AccWeights

	w0 := M.AccEventWeight{WeightId: "1", Weight_value: 5.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w1 := M.AccEventWeight{WeightId: "2", Weight_value: 7.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w2 := M.AccEventWeight{WeightId: "3", Weight_value: 8.0, Is_deleted: false, EventName: "$session", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w3 := M.AccEventWeight{WeightId: "4", Weight_value: 9.0, Is_deleted: false, EventName: "$form_submitted", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w4 := M.AccEventWeight{WeightId: "5", Weight_value: 10.0, Is_deleted: false, EventName: "www.acme.com", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w5 := M.AccEventWeight{WeightId: "6", Weight_value: 11.0, Is_deleted: false, EventName: "www.acme.com/pricing", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w6 := M.AccEventWeight{WeightId: "7", Weight_value: 8.0, Is_deleted: false, EventName: "$session"}
	w7 := M.AccEventWeight{WeightId: "8", Weight_value: 11.0, Is_deleted: false, EventName: "", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Kenya"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w8 := M.AccEventWeight{WeightId: "9", Weight_value: 11.0, Is_deleted: false, EventName: "", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Brazil"}, Operator: true, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}

	weightRules := []M.AccEventWeight{w0, w1, w2, w3, w4, w5, w6, w7, w8}
	finalWeights.WeightConfig = weightRules
	finalWeights.SaleWindow = 10

	ev_r1 := []string{"3", "7"}
	ev_r2 := []string{"1", "2"}
	ev_r3 := []string{"4"}
	ev_r4 := []string{"3", "7"}
	ev_r5 := []string{"8"}
	ev_r6 := []string{"8"}
	ev_r7 := []string{"8"}
	ev_r8 := []string{"9"}

	ev_rules := [][]string{ev_r1, ev_r2, ev_r3, ev_r4, ev_r5, ev_r6, ev_r7, ev_r8}
	cr, err := T.DeduplicateWeights(finalWeights) //new ids will not be added, as weight id is already filled
	assert.Nil(t, err)
	for _, w := range cr.WeightConfig {
		log.Debug(w)
	}
	weightmap, _ := T.CreateweightMap(&cr)
	log.Debug(weightmap)
	for idx, ev := range events {
		ids := T.FilterEvents(ev, weightmap)
		log.Debug(ids)
		assert.ElementsMatch(t, ids, ev_rules[idx], fmt.Sprintf("events :%d", idx))

	}
	assert.Equal(t, 6, len(weightmap))
}

func TestFilterAndCountEventsFromFile(t *testing.T) {
	events := make([]P.CounterEventFormat, 0)

	fileEvents, err := os.Open("./data/events_score_data.txt")
	assert.Nil(t, err)
	scanner := bufio.NewScanner(fileEvents)
	line_count := 0
	for scanner.Scan() {
		line := scanner.Text()
		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return
		}

		events = append(events, eventDetails)
		line_count += 1
		if line_count%10 == 0 {
			fmt.Printf(fmt.Sprintf("Read lines :%d", line_count))
		}
	}

	fileWeights, err := os.Open("./data/events_score_weights.txt")
	assert.Nil(t, err)
	scannerW := bufio.NewScanner(fileWeights)
	var weightsRequest M.AccWeights
	var weights M.AccWeights

	for scannerW.Scan() {
		line := scannerW.Text()
		if err := json.Unmarshal([]byte(line), &weightsRequest); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return
		}
	}
	weights.SaleWindow = weightsRequest.SaleWindow

	for _, wtVal := range weightsRequest.WeightConfig {
		var r M.AccEventWeight
		r.EventName = wtVal.EventName
		r.Is_deleted = wtVal.Is_deleted
		r.Rule = wtVal.Rule
		r.WeightId = wtVal.WeightId
		r.Weight_value = wtVal.Weight_value

		weights.WeightConfig = append(weights.WeightConfig, r)
	}

	cr, err := T.DeduplicateWeights(weights) //new ids will not be added, as weight id is already filled
	assert.Nil(t, err)

	weightmap, _ := T.CreateweightMap(&cr)
	log.Debug(weightmap)
	for idx, ev := range events {
		ids := T.FilterEvents(&ev, weightmap)
		log.Debugf("%d, %v", idx, ids)
	}
	assert.Equal(t, 5, len(weightmap))
}

func TestCountingEvents(t *testing.T) {
	fileWeights, err := os.Open("./data/events_score_weights.txt")
	assert.Nil(t, err)
	scannerW := bufio.NewScanner(fileWeights)
	var weightsRequest M.AccWeights
	var weights M.AccWeights

	for scannerW.Scan() {
		line := scannerW.Text()
		if err := json.Unmarshal([]byte(line), &weightsRequest); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return
		}
	}
	weights.SaleWindow = weightsRequest.SaleWindow

	for _, wtVal := range weightsRequest.WeightConfig {
		var r M.AccEventWeight
		r.EventName = wtVal.EventName
		r.Is_deleted = wtVal.Is_deleted
		r.Rule = wtVal.Rule
		r.WeightId = wtVal.WeightId
		r.Weight_value = wtVal.Weight_value

		weights.WeightConfig = append(weights.WeightConfig, r)
	}

	cr, err := T.DeduplicateWeights(weights) //new ids will not be added, as weight id is already filled
	assert.Nil(t, err)

	fileEvents, err := os.Open("./data/events_score_data.txt")
	assert.Nil(t, err)

	var userGroupCount map[string]*T.AggEventsOnUserAndGroup = make(map[string]*T.AggEventsOnUserAndGroup)
	weightmap, _ := T.CreateweightMap(&cr)

	err = T.AggEventsOnUsers(fileEvents, userGroupCount, weightmap)

	for k, v := range userGroupCount {
		// log.Debugf("uid :%s", k)
		line, err := json.Marshal(v)
		assert.Nil(t, err)
		log.Debugf("%s, %s", k, string(line))
	}
	assert.Nil(t, err)
}
