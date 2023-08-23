package tests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"factors/model/model"
	M "factors/model/model"
	mm "factors/model/store/memsql"
	P "factors/pattern"
	T "factors/task/account_scoring"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestAccScoreDeduplicateWeights(t *testing.T) {

	var finalWeights M.AccWeights

	w0 := M.AccEventWeight{WeightId: "", Weight_value: 5.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w1 := M.AccEventWeight{WeightId: "", Weight_value: 7.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w2 := M.AccEventWeight{WeightId: "", Weight_value: 8.0, Is_deleted: false, EventName: "$session", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w3 := M.AccEventWeight{WeightId: "", Weight_value: 9.0, Is_deleted: false, EventName: "$form_submitted", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w4 := M.AccEventWeight{WeightId: "", Weight_value: 10.0, Is_deleted: false, EventName: "www.acme.com", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w5 := M.AccEventWeight{WeightId: "", Weight_value: 11.0, Is_deleted: false, EventName: "www.acme.com/pricing", Rule: []M.WeightKeyValueTuple{{Key: "$Country", Value: []string{"India"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "categorical"}}}
	w6 := M.AccEventWeight{WeightId: "1", Weight_value: 5.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "user", ValueType: "categorical"}}}

	weightRules := []M.AccEventWeight{w0, w1, w2, w3, w4, w5, w6}
	finalWeights.WeightConfig = weightRules
	finalWeights.SaleWindow = 10

	cr, err := T.DeduplicateWeights(finalWeights)
	assert.Nil(t, err)
	assert.Equal(t, len(weightRules)-1, len(cr.WeightConfig), "duplicate elements are not removed")

}

func TestAccScoreFilterAndCountEvents(t *testing.T) {

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

	// update weights
	w0 := M.AccEventWeight{WeightId: "1", Weight_value: 5.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w1 := M.AccEventWeight{WeightId: "2", Weight_value: 7.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w2 := M.AccEventWeight{WeightId: "3", Weight_value: 8.0, Is_deleted: false, EventName: "$session", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w3 := M.AccEventWeight{WeightId: "4", Weight_value: 9.0, Is_deleted: false, EventName: "$form_submitted", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w4 := M.AccEventWeight{WeightId: "5", Weight_value: 10.0, Is_deleted: false, EventName: "www.acme.com", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w5 := M.AccEventWeight{WeightId: "6", Weight_value: 11.0, Is_deleted: false, EventName: "www.acme.com/pricing", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w6 := M.AccEventWeight{WeightId: "7", Weight_value: 8.0, Is_deleted: false, EventName: "$session"}
	w7 := M.AccEventWeight{WeightId: "8", Weight_value: 11.0, Is_deleted: false, EventName: "", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Kenya"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w8 := M.AccEventWeight{WeightId: "9", Weight_value: 11.0, Is_deleted: false, EventName: "", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Brazil"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}

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

func TestAccScoreFilterAndCountEventsFromFile(t *testing.T) {
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

func TestAccScoreCountingEvents(t *testing.T) {
	config := make(map[string]interface{})

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

	err = T.AggEventsOnUsers(fileEvents, userGroupCount, weightmap, config)

	for k, v := range userGroupCount {
		// log.Debugf("uid :%s", k)
		line, err := json.Marshal(v)
		assert.Nil(t, err)
		log.Debugf("%s, %s", k, string(line))
	}
	assert.Nil(t, err)
}

func TestAccScoreFilterCountAndScoreEvents(t *testing.T) {

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

	projectId := int64(0)
	for _, e := range ev {
		var testev *P.CounterEventFormat
		err := json.Unmarshal([]byte(e), &testev)
		assert.Nil(t, err)
		events = append(events, testev)
	}

	var finalWeights M.AccWeights

	w0 := M.AccEventWeight{WeightId: "1", Weight_value: 1.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w1 := M.AccEventWeight{WeightId: "2", Weight_value: 1.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w2 := M.AccEventWeight{WeightId: "3", Weight_value: 1.0, Is_deleted: false, EventName: "$session", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w3 := M.AccEventWeight{WeightId: "4", Weight_value: 1.0, Is_deleted: false, EventName: "$form_submitted", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w4 := M.AccEventWeight{WeightId: "5", Weight_value: 1.0, Is_deleted: false, EventName: "www.acme.com", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w5 := M.AccEventWeight{WeightId: "6", Weight_value: 1.0, Is_deleted: false, EventName: "www.acme.com/pricing", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w6 := M.AccEventWeight{WeightId: "7", Weight_value: 1.0, Is_deleted: false, EventName: "$session"}
	w7 := M.AccEventWeight{WeightId: "8", Weight_value: 1.0, Is_deleted: false, EventName: "", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Kenya"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w8 := M.AccEventWeight{WeightId: "9", Weight_value: 1.0, Is_deleted: false, EventName: "", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Brazil"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}

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
	countsmap := make(map[string]int64)
	log.Debug(weightmap)
	for idx, ev := range events {
		ids := T.FilterEvents(ev, weightmap)
		log.Debug(ids)
		for _, fids := range ids {
			countsmap[fids] += 1
		}
		assert.ElementsMatch(t, ids, ev_rules[idx], fmt.Sprintf("events :%d", idx))

	}
	assert.Equal(t, 6, len(weightmap))

	// score should be half on mid of sale window
	current_time := time.Now()
	half_num_day := current_time.AddDate(0, 0, -1*int(finalWeights.SaleWindow/2))
	dateString := T.GetDateOnlyFromTimestamp(half_num_day.Unix())
	account_score, counts_map, decayValue, err := mm.ComputeAccountScore(cr, countsmap, dateString)
	s := fmt.Sprintf("acc score : %f , decay_value : %f , counts map :%v", account_score, decayValue, counts_map)
	log.Debugf(s)
	assert.Nil(t, err)
	// expected counts map is all ones
	//countsmap[id]count [1:1 2:1 3:2 4:1 7:2 8:3 9:1]
	assert.Equal(t, account_score, float32(5.500), "score miscalculation")
	assert.Equal(t, decayValue, 0.5, "decay value miscalculation")

	//last day goes to zero as decay is 1 .
	current_time = time.Now()
	onlastsaleday := current_time.AddDate(0, 0, -1*int(finalWeights.SaleWindow))
	dateString = T.GetDateOnlyFromTimestamp(onlastsaleday.Unix())

	countsmapf64 := make(map[string]float64)
	for k, v := range countsmapf64 {
		countsmapf64[k] = v
	}
	account_score_, err := mm.ComputeAccountScoreOnLastEvent(projectId, cr, countsmapf64)
	s = fmt.Sprintf("acc score : %f  , counts map :%v", account_score, countsmapf64)
	log.Debugf(s)
	assert.Nil(t, err)
	// expected counts map is all ones
	//countsmap[id]count [1:1 2:1 3:2 4:1 7:2 8:3 9:1]
	assert.Equal(t, account_score_, float64(0), "score miscalculation")

	// score should be 0 after the sale window has been exceeded
	current_time = time.Now()
	sale_day_exceed := current_time.AddDate(0, 0, -1*int(finalWeights.SaleWindow)-1)
	dateString = T.GetDateOnlyFromTimestamp(sale_day_exceed.Unix())

	countsmapf64 = make(map[string]float64)
	for k, v := range countsmapf64 {
		countsmapf64[k] = v
	}
	account_score_, err = mm.ComputeAccountScoreOnLastEvent(projectId, cr, countsmapf64)
	s = fmt.Sprintf("acc score : %f , counts map :%v", account_score_, counts_map)
	log.Debugf(s)
	assert.Nil(t, err)
	// expected counts map is all ones
	//countsmap[id]count [1:1 2:1 3:2 4:1 7:2 8:3 9:1]
	assert.Equal(t, account_score_, float64(0), "score miscalculation")
}

func TestAccScoreUpdateLastEventsDay(t *testing.T) {

	var prevCountsOfUser map[string]map[string]M.LatestScore = make(map[string]map[string]M.LatestScore)
	users := make(map[string]*T.AggEventsOnUserAndGroup)
	var finalWeights M.AccWeights

	we0 := M.AccEventWeight{WeightId: "a", Weight_value: 1.0, Is_deleted: false, EventName: "a"}
	we1 := M.AccEventWeight{WeightId: "b", Weight_value: 1.0, Is_deleted: false, EventName: "b"}
	we2 := M.AccEventWeight{WeightId: "c", Weight_value: 1.0, Is_deleted: false, EventName: "c"}
	we3 := M.AccEventWeight{WeightId: "d", Weight_value: 1.0, Is_deleted: false, EventName: "d"}
	we4 := M.AccEventWeight{WeightId: "e", Weight_value: 1.0, Is_deleted: false, EventName: "e"}

	finalWeights.WeightConfig = []M.AccEventWeight{we0, we1, we2, we3, we4}
	finalWeights.SaleWindow = 10

	mweights := make(map[string][]M.AccEventWeight, 0)
	for _, v := range finalWeights.WeightConfig {

		mweights[v.WeightId] = make([]M.AccEventWeight, 1)
		mweights[v.WeightId] = append(mweights[v.WeightId], v)
	}

	e1 := M.EventAgg{EventName: "a", EventId: "a", EventCount: 1}
	e2 := M.EventAgg{EventName: "b", EventId: "b", EventCount: 1}
	e3 := M.EventAgg{EventName: "c", EventId: "c", EventCount: 1}
	e4 := M.EventAgg{EventName: "d", EventId: "d", EventCount: 1}
	e5 := M.EventAgg{EventName: "e", EventId: "e", EventCount: 1}

	eventsCount := make(map[string]*M.EventAgg)

	eventsCount["e1"] = &e1
	eventsCount["e2"] = &e2
	eventsCount["e3"] = &e3
	eventsCount["e4"] = &e4
	eventsCount["e5"] = &e5

	propertiesmap := make(map[string]int64)
	propertiesmap["p1"] = 2
	propertiesmap["p2"] = 3
	propertiesmap["p3"] = 4
	propertiesmap["p4"] = 3
	propertiesmap["p5"] = 14
	propertiesmap["p6"] = 13
	propertiesmap["p7"] = 12

	p1 := T.PropAggregate{Name: "$channel", Properties: propertiesmap}
	p1u1 := make(map[string]T.PropAggregate)
	p1u1["$channel"] = p1

	user1 := T.AggEventsOnUserAndGroup{User_id: "1", EventsCount: eventsCount, Is_group: false, Properties: p1u1}
	users["1"] = &user1

	prevCountsOfUser["1"] = make(map[string]M.LatestScore)
	currentTimestamp := time.Now()
	currentTS := currentTimestamp.Unix()
	w3 := make(map[string]float64)
	w3["a"] = 1.0
	w3["b"] = 1.0
	w3["c"] = 1.0
	w3["d"] = 1.0
	w3["e"] = 1.0
	w2 := make(map[string]float64)
	w2["a"] = 1
	w2["b"] = 1
	w2["c"] = 1
	w2["d"] = 1
	w2["e"] = 1
	w2_date_unix := currentTimestamp.AddDate(0, 0, -1*2)
	w3_date_unix := currentTimestamp.AddDate(0, 0, -1*3)

	w2_date_string := T.GetDateOnlyFromTimestamp(w2_date_unix.Unix())
	w3_date_string := T.GetDateOnlyFromTimestamp(w3_date_unix.Unix())
	p1u2 := make(map[string]map[string]int64)
	p1u3 := make(map[string]map[string]int64)
	p1u2["$channel"] = make(map[string]int64)
	p1u2["$channel"]["p1"] = 4
	p1u2["$channel"]["p2"] = 1
	p1u2["$channel"]["p4"] = 1
	p1u2["$channel"]["p5"] = 1

	p1u3["$channel"] = make(map[string]int64)
	p1u3["$channel"]["p1"] = 4
	p1u3["$channel"]["p3"] = 2
	p1u3["$channel"]["p4"] = 1
	p1u3["$channel"]["p5"] = 10

	prevCountsOfUser["1"][w3_date_string] = M.LatestScore{Date: w3_date_unix.Unix(), EventsCount: w3, Properties: p1u2}
	prevCountsOfUser["1"][w2_date_string] = M.LatestScore{Date: w2_date_unix.Unix(), EventsCount: w2, Properties: p1u3}

	dbup := T.UpdateCountsWithDecayToUpdateDB(users, mweights, prevCountsOfUser, finalWeights.SaleWindow, currentTS)
	log.Debugf("result:%v", dbup)
	assert.Equal(t, dbup[0].CurrEventCount.EventsCount["a"], float64(1))
	assert.Equal(t, dbup[0].CurrEventCount.EventsCount["b"], float64(1))
	assert.Equal(t, dbup[0].CurrEventCount.EventsCount["c"], float64(1))
	assert.Equal(t, dbup[0].CurrEventCount.EventsCount["d"], float64(1))
	assert.Equal(t, dbup[0].CurrEventCount.EventsCount["e"], float64(1))

	assert.Equal(t, dbup[0].Lastevent.EventsCount["a"], 2.5)
	assert.Equal(t, dbup[0].Lastevent.EventsCount["b"], 2.5)
	assert.Equal(t, dbup[0].Lastevent.EventsCount["c"], 2.5)
	assert.Equal(t, dbup[0].Lastevent.EventsCount["d"], 2.5)
	assert.Equal(t, dbup[0].Lastevent.EventsCount["e"], 2.5)
	assert.Equal(t, int64(10), dbup[0].Lastevent.Properties["$channel"]["p1"])
	assert.Equal(t, int64(4), dbup[0].Lastevent.Properties["$channel"]["p2"])
	assert.Equal(t, int64(6), dbup[0].Lastevent.Properties["$channel"]["p3"])
	assert.Equal(t, T.TAKETOPK, len(dbup[0].Lastevent.Properties["$channel"]), "Take Topk properties violated")

	assert.Equal(t, 1, len(dbup))
	// assert.Nil(t, err)
}

func TestGenerateDate(t *testing.T) {

	currentTimestamp := time.Now()
	salewindow := int(10)

	ds := mm.GenDateStringsForLastNdays(currentTimestamp.Unix(), int64(salewindow))
	log.Debugf("generated dates : %v", ds)
	assert.Equal(t, salewindow, len(ds))

}

func TestGenerateAccountScores(t *testing.T) {
	var finalWeights M.AccWeights
	projectId := int64(0)
	w0 := M.AccEventWeight{WeightId: "1", Weight_value: 1.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w1 := M.AccEventWeight{WeightId: "2", Weight_value: 1.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w2 := M.AccEventWeight{WeightId: "3", Weight_value: 1.0, Is_deleted: false, EventName: "$session", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w3 := M.AccEventWeight{WeightId: "4", Weight_value: 1.0, Is_deleted: false, EventName: "$form_submitted", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w4 := M.AccEventWeight{WeightId: "5", Weight_value: 1.0, Is_deleted: false, EventName: "www.acme.com", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w5 := M.AccEventWeight{WeightId: "6", Weight_value: 1.0, Is_deleted: false, EventName: "www.acme.com/pricing", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w6 := M.AccEventWeight{WeightId: "7", Weight_value: 1.0, Is_deleted: false, EventName: "$session"}
	w7 := M.AccEventWeight{WeightId: "8", Weight_value: 1.0, Is_deleted: false, EventName: "", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Kenya"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w8 := M.AccEventWeight{WeightId: "9", Weight_value: 1.0, Is_deleted: false, EventName: "", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Brazil"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}

	weightRules := []M.AccEventWeight{w0, w1, w2, w3, w4, w5, w6, w7, w8}
	finalWeights.WeightConfig = weightRules
	finalWeights.SaleWindow = 10

	numDaysToTrend := finalWeights.SaleWindow
	currentTs := time.Now().Unix()
	DateStringToday := T.GetDateOnlyFromTimestamp(currentTs)

	currentDate := model.GetDateFromString(DateStringToday)
	prevDateTotrend := time.Unix(currentDate, 0).AddDate(0, 0, -1*int(numDaysToTrend)).Unix()

	countsMapDays := make(map[string]model.LatestScore)
	e1 := map[string]float64{"1": 1, "2": 1, "3": 1, "4": 1}
	for i := (numDaysToTrend + 10); i > 0; i-- {
		if i%2 == 0 {
			prevDate := time.Unix(currentDate, 0).AddDate(0, 0, -1*int(i)).Unix()
			prevDateString := T.GetDateOnlyFromTimestamp(prevDate)
			countsMapDays[prevDateString] = model.LatestScore{Date: prevDate, EventsCount: e1}
		}
	}

	_, scores, _, err := mm.CalculatescoresPerAccount(projectId, &finalWeights, currentDate, prevDateTotrend, countsMapDays)
	for k1, v1 := range scores {
		_, ok := countsMapDays[k1]
		fmt.Printf("countMapDays : %s , score : %f ,%v \n", k1, v1, ok)
	}
	assert.Nil(t, err)

}

func TestGenerationOfScore(t *testing.T) {

	var finalWeights M.AccWeights
	projectId := int64(0)
	w0 := M.AccEventWeight{WeightId: "1", Weight_value: 1.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w1 := M.AccEventWeight{WeightId: "2", Weight_value: 1.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w2 := M.AccEventWeight{WeightId: "3", Weight_value: 1.0, Is_deleted: false, EventName: "$session", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w3 := M.AccEventWeight{WeightId: "4", Weight_value: 1.0, Is_deleted: false, EventName: "$form_submitted", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}

	weightRules := []M.AccEventWeight{w0, w1, w2, w3}
	finalWeights.WeightConfig = weightRules
	finalWeights.SaleWindow = 30

	counts := make(map[string]float64)
	maxNum := 30

	currentTimestamp := time.Now().AddDate(0, 0, -1*maxNum)
	scores := make([]float64, maxNum)
	counts["1"] += float64(1)
	counts["2"] += float64(1)
	counts["3"] += float64(1)
	counts["4"] += float64(1)

	for idx := 0; idx < maxNum; idx++ {

		day := T.GetDateOnlyFromTimestamp(currentTimestamp.AddDate(0, 0, 1*idx).Unix())
		if idx == 0 {
			idx1 := 0
			counts["1"] += float64(idx1)
			counts["2"] += float64(idx1)
			counts["3"] += float64(idx1)
			counts["4"] += float64(idx1)
		}

		score := mm.ComputeScoreWithWeightsAndCounts(projectId, &finalWeights, counts, day)
		scores[idx] = score

	}
	log.Debugf("Scores : %v", scores)
	score := mm.ComputeScoreWithWeightsAndCounts(projectId, &finalWeights, counts, T.GetDateOnlyFromTimestamp(time.Now().Unix()))
	log.Debugf("Scores on current day  : %v,%f", counts, score)

	assert.IsDecreasing(t, scores)
}

func TestGenerationOfScoresInPeriod(t *testing.T) {

	var finalWeights M.AccWeights
	projectId := int64(0)
	countsOnDays := make(map[string]M.LatestScore)
	w0 := M.AccEventWeight{WeightId: "1", Weight_value: 1.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w1 := M.AccEventWeight{WeightId: "2", Weight_value: 1.0, Is_deleted: false, EventName: "$pageview", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w2 := M.AccEventWeight{WeightId: "3", Weight_value: 1.0, Is_deleted: false, EventName: "$session", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}
	w3 := M.AccEventWeight{WeightId: "4", Weight_value: 1.0, Is_deleted: false, EventName: "$form_submitted", Rule: []M.WeightKeyValueTuple{{Key: "$country", Value: []string{"Australia"}, Operator: M.EqualsOpStr, LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}}}

	weightRules := []M.AccEventWeight{w0, w1, w2, w3}
	finalWeights.WeightConfig = weightRules
	finalWeights.SaleWindow = 10

	counts := make(map[string]float64)
	maxNum := 20

	currentTimestamp := time.Now().AddDate(0, 0, -1*maxNum)
	// scores := make([]float64, maxNum)
	counts["1"] += float64(1)
	counts["2"] += float64(1)
	counts["3"] += float64(1)
	counts["4"] += float64(1)

	for idx := 0; idx < maxNum; idx++ {
		var events M.LatestScore
		countsEvent := make(map[string]float64)
		t := currentTimestamp.AddDate(0, 0, 1*idx).Unix()
		day := T.GetDateOnlyFromTimestamp(t)
		events.Date = t
		idx1 := 1
		if idx > maxNum/2 {
			idx1 = 1
		}
		countsEvent["1"] = float64(idx1)
		countsEvent["2"] = float64(idx1)
		countsEvent["3"] = float64(idx1)
		countsEvent["4"] = float64(idx1)
		events.EventsCount = countsEvent
		countsOnDays[day] = events
	}

	accountScoreMap, err := mm.ComputeTrendWrapper(projectId, time.Now().Unix(), countsOnDays, &finalWeights)
	for d, e := range accountScoreMap {
		log.Debugf("day : %s score :%f", d, e)
	}
	assert.Equal(t, maxNum, len(countsOnDays))
	assert.Nil(t, err)
}
