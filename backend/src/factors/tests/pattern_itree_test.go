package tests

import (
	"bufio"
	"encoding/json"
	P "factors/pattern"
	PS "factors/pattern_server/store"
	PW "factors/pattern_service_wrapper"
	U "factors/util"
	"fmt"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestBuildNewItree(t *testing.T) {
	// Hypothetical data on which Itree is built, ending with Y.
	// User1:  B  C  A  A
	// User2:  A  A  B  C  Y  C
	// User3:  C  A  B
	// User4:  A  B  C  C  Y
	// User5:  A  A  B
	// User6:  A  B  C  Y  A
	// User7:  B  Y  C
	// User8:  C  Y  A
	// User9:  A  B  C
	// User10: A  B  A  C  B  C  Y
	var userCount uint = 10

	type testPatternInfo struct {
		eventNames          []string
		perUserCount, count uint
	}
	// Patterns with zero counts are not listed.
	// Pattern, PerUserCount, Count
	tpis := []testPatternInfo{
		{eventNames: []string{U.SEN_ALL_ACTIVE_USERS}, perUserCount: 10, count: 10},
		{eventNames: []string{"A"}, perUserCount: 9, count: 14},
		{eventNames: []string{"B"}, perUserCount: 9, count: 10},
		{eventNames: []string{"C"}, perUserCount: 9, count: 12},
		{eventNames: []string{"Y"}, perUserCount: 6, count: 6},
		{eventNames: []string{"A", "B"}, perUserCount: 7, count: 8},
		{eventNames: []string{"A", "C"}, perUserCount: 5, count: 5},
		{eventNames: []string{"A", "Y"}, perUserCount: 4, count: 4},
		{eventNames: []string{"B", "A"}, perUserCount: 3, count: 3},
		{eventNames: []string{"B", "C"}, perUserCount: 7, count: 8},
		{eventNames: []string{"B", "Y"}, perUserCount: 5, count: 5},
		{eventNames: []string{"C", "A"}, perUserCount: 4, count: 4},
		{eventNames: []string{"C", "B"}, perUserCount: 2, count: 2},
		{eventNames: []string{"C", "Y"}, perUserCount: 5, count: 5},
		{eventNames: []string{"Y", "A"}, perUserCount: 2, count: 2},
		{eventNames: []string{"Y", "C"}, perUserCount: 2, count: 2},
		{eventNames: []string{"A", "C", "B"}, perUserCount: 1, count: 1},
		{eventNames: []string{"A", "B", "C"}, perUserCount: 5, count: 5},
		{eventNames: []string{"A", "Y", "C"}, perUserCount: 1, count: 1},
		{eventNames: []string{"A", "B", "Y"}, perUserCount: 4, count: 4},
		{eventNames: []string{"A", "C", "Y"}, perUserCount: 4, count: 4},
		{eventNames: []string{"B", "C", "A"}, perUserCount: 2, count: 2},
		{eventNames: []string{"B", "Y", "A"}, perUserCount: 1, count: 1},
		{eventNames: []string{"B", "A", "C"}, perUserCount: 1, count: 1},
		{eventNames: []string{"B", "Y", "C"}, perUserCount: 2, count: 2},
		{eventNames: []string{"B", "A", "Y"}, perUserCount: 1, count: 1},
		{eventNames: []string{"B", "C", "Y"}, perUserCount: 4, count: 4},
		{eventNames: []string{"C", "Y", "A"}, perUserCount: 2, count: 2},
		{eventNames: []string{"C", "A", "B"}, perUserCount: 1, count: 1},
		{eventNames: []string{"C", "B", "Y"}, perUserCount: 1, count: 1},
		{eventNames: []string{"A", "C", "B", "Y"}, perUserCount: 1, count: 1},
		{eventNames: []string{"A", "B", "C", "Y"}, perUserCount: 4, count: 4},
		{eventNames: []string{"A", "B", "Y", "C"}, perUserCount: 1, count: 1},
		{eventNames: []string{"B", "A", "C", "Y"}, perUserCount: 1, count: 1},
		{eventNames: []string{"B", "C", "Y", "A"}, perUserCount: 1, count: 1},
	}
	patterns := []*P.Pattern{}
	for _, tpi := range tpis {
		p, _ := P.NewPattern(tpi.eventNames, nil)
		p.PerOccurrenceCount = tpi.count
		p.PerUserCount = tpi.perUserCount
		p.TotalUserCount = userCount
		patterns = append(patterns, p)
	}
	pw := NewMockPatternServiceWrapper(patterns, nil)
	itree, err := PW.BuildNewItree("", "", nil, "Y", nil, pw, P.COUNT_TYPE_PER_USER, 1)
	assert.Nil(t, err)
	assert.NotNil(t, itree)

	// Expected iTree.
	//
	//                       Y(0)
	//                       |
	//           -------------------------------
	//           |                |             |
	//          AY(1)            BY(2)         CY(3)
	//           |                |             |
	//        -----------      ---------     ------------
	//        |         |      |       |     |           |
	//       ACY(4)    ABY(5) BAY(6)  BCY(7) CAY(Skip)  CBY(8)
	//        |         |
	//        |         |
	//      ACBY(9)    ABCY(10)
	type expectedNode struct {
		patternString                              string
		index, parentIndex                         int
		rightInfo, leftInfo, overallInfo, infoDrop float64
		confidence, confidenceGain                 float64
		fpp, fpr, fcp, fcr                         float64
	}
	node0 := expectedNode{
		patternString: "Y",
		index:         0, parentIndex: -1,
		rightInfo: 0.97095, leftInfo: 0.0,
		overallInfo: 0.97095, confidence: 0.6,
		fpp: 0.0, fpr: 0.0, fcp: 10.0, fcr: 6.0,
	}
	node1 := expectedNode{
		patternString: "A,Y",
		index:         1, parentIndex: 0,
		rightInfo: 0.99108, leftInfo: 0.0,
		overallInfo: 0.89197, infoDrop: 0.07898,
		confidence: 0.444444, confidenceGain: -0.155556,
		fpp: 10.0, fpr: 6.0, fcp: 9.0, fcr: 4.0,
	}
	node2 := expectedNode{
		patternString: "B,Y",
		index:         2, parentIndex: 0,
		rightInfo: 0.99108, leftInfo: 0.0,
		overallInfo: 0.89197, infoDrop: 0.07898,
		confidence: 0.555556, confidenceGain: -0.044444,
		fpp: 10.0, fpr: 6.0, fcp: 9.0, fcr: 5.0,
	}
	node3 := expectedNode{
		patternString: "C,Y",
		index:         3, parentIndex: 0,
		rightInfo: 0.99108, leftInfo: 0.0,
		overallInfo: 0.89197, infoDrop: 0.07898,
		confidence: 0.555556, confidenceGain: -0.044444,
		fpp: 10.0, fpr: 6.0, fcp: 9.0, fcr: 5.0,
	}

	node4 := expectedNode{
		patternString: "A,C,Y",
		index:         4, parentIndex: 1,
		rightInfo: 0.72193, leftInfo: 0.0,
		overallInfo: 0.40107, infoDrop: 0.59001,
		confidence: 0.80, confidenceGain: 0.355556,
		fpp: 9.0, fpr: 4.0, fcp: 5.0, fcr: 4.0,
	}
	node5 := expectedNode{
		patternString: "A,B,Y",
		index:         5, parentIndex: 1,
		rightInfo: 0.98523, leftInfo: 0.0,
		overallInfo: 0.76629, infoDrop: 0.22479,
		confidence: 0.571429, confidenceGain: 0.126985,
		fpp: 9.0, fpr: 4.0, fcp: 7.0, fcr: 4.0,
	}
	node9 := expectedNode{
		patternString: "A,C,B,Y",
		index:         9, parentIndex: 4,
		rightInfo: 0.0, leftInfo: 0.81128,
		overallInfo: 0.64902, infoDrop: 0.07291,
		confidence: 1.0, confidenceGain: 0.2,
		fpp: 5.0, fpr: 4.0, fcp: 1.0, fcr: 1.0,
	}
	node10 := expectedNode{
		patternString: "A,B,C,Y",
		index:         10, parentIndex: 5,
		rightInfo: 0.72193, leftInfo: 0.0,
		overallInfo: 0.51566, infoDrop: 0.46957,
		confidence: 0.8, confidenceGain: 0.228571,
		fpp: 7.0, fpr: 4.0, fcp: 5.0, fcr: 4.0,
	}
	expectedNodes := []*expectedNode{
		&node0, &node1, &node2, &node3, &node4, &node5, &node9, &node10}

	for i, eNode := range expectedNodes {
		aNode := itree.Nodes[eNode.index]
		assert.Equal(t, eNode.patternString, aNode.Pattern.String())
		assert.Equal(t, eNode.parentIndex, aNode.ParentIndex, fmt.Sprintf("Node: %s", eNode.patternString))
		assert.InDelta(t, eNode.rightInfo, aNode.RightInformation, 0.0001, fmt.Sprintf("Node: %s", eNode.patternString))
		assert.InDelta(t, eNode.overallInfo, aNode.OverallInformation, 0.0001, fmt.Sprintf("Node: %s", eNode.patternString))
		assert.InDelta(t, eNode.confidence, aNode.Confidence, 0.0001, fmt.Sprintf("Node: %s", eNode.patternString))
		if i != 0 {
			assert.InDelta(t, eNode.infoDrop, aNode.InformationDrop, 0.0001, fmt.Sprintf("Node: %s", eNode.patternString))
			assert.InDelta(t, eNode.confidenceGain, aNode.ConfidenceGain, 0.0001, fmt.Sprintf("Node: %s", eNode.patternString))
		}
	}

	// Build withs start and end event. Expected tree would be the subtree of above tree with root node AY.
	itree, err = PW.BuildNewItree("", "A", nil, "Y", nil, pw, P.COUNT_TYPE_PER_USER, 1)
	assert.Nil(t, err)
	assert.NotNil(t, itree)
	node0 = expectedNode{
		patternString: "A,Y",
		index:         0, parentIndex: -1,
		rightInfo: 0.99107, overallInfo: 0.99107,
		confidence: 0.444444, fcp: 9.0, fcr: 4.0,
	}
	node1 = expectedNode{
		patternString: "A,C,Y",
		index:         1, parentIndex: 0,
		rightInfo: 0.72193, leftInfo: 0.0,
		overallInfo: 0.40107, infoDrop: 0.59,
		confidence: 0.80, confidenceGain: 0.355556,
		fpp: 9.0, fpr: 4.0, fcp: 5.0, fcr: 4.0,
	}
	node2 = expectedNode{
		patternString: "A,B,Y",
		index:         2, parentIndex: 0,
		rightInfo: 0.98523, leftInfo: 0.0,
		overallInfo: 0.76629, infoDrop: 0.22478,
		confidence: 0.571429, confidenceGain: 0.126985,
		fpp: 9.0, fpr: 4.0, fcp: 7.0, fcr: 4.0,
	}
	node3 = expectedNode{
		patternString: "A,C,B,Y",
		index:         3, parentIndex: 1,
		rightInfo: 0.0, leftInfo: 0.81128,
		overallInfo: 0.64902, infoDrop: 0.07291,
		confidence: 1.0, confidenceGain: 0.2,
		fpp: 5.0, fpr: 4.0, fcp: 1.0, fcr: 1.0,
	}
	node4 = expectedNode{
		patternString: "A,B,C,Y",
		index:         4, parentIndex: 2,
		rightInfo: 0.72193, leftInfo: 0.0,
		overallInfo: 0.51566, infoDrop: 0.46957,
		confidence: 0.8, confidenceGain: 0.228571,
		fpp: 7.0, fpr: 4.0, fcp: 5.0, fcr: 4.0,
	}
	expectedNodes = []*expectedNode{
		&node0, &node1, &node2, &node3, &node4}
	for i, eNode := range expectedNodes {
		aNode := itree.Nodes[eNode.index]
		assert.Equal(t, eNode.patternString, aNode.Pattern.String())
		assert.Equal(t, eNode.parentIndex, aNode.ParentIndex, fmt.Sprintf("Node: %s", eNode.patternString))
		assert.InDelta(t, eNode.rightInfo, aNode.RightInformation, 0.0001, fmt.Sprintf("Node: %s", eNode.patternString))
		assert.InDelta(t, eNode.overallInfo, aNode.OverallInformation, 0.0001, fmt.Sprintf("Node: %s", eNode.patternString))
		assert.InDelta(t, eNode.confidence, aNode.Confidence, 0.0001, fmt.Sprintf("Node: %s", eNode.patternString))
		if i != 0 {
			assert.InDelta(t, eNode.infoDrop, aNode.InformationDrop, 0.0001, fmt.Sprintf("Node: %s", eNode.patternString))
			assert.InDelta(t, eNode.confidenceGain, aNode.ConfidenceGain, 0.0001, fmt.Sprintf("Node: %s", eNode.patternString))
		}
	}
}

func TestFactorV1(t *testing.T) {

	filepath := "data/chunk_data.txt"

	patterns := make([]*P.Pattern, 0)

	file, err := os.Open(filepath)
	assert.Nil(t, err)
	scanner := bufio.NewScanner(file)
	buf := make([]byte, P.MAX_PATTERN_BYTES)
	scanner.Buffer(buf, P.MAX_PATTERN_BYTES)

	for scanner.Scan() {
		line := string(scanner.Text())
		var ptm PS.PatternWithMeta
		var patternDetail P.Pattern
		err := json.Unmarshal([]byte(line), &ptm)
		if err != nil {
			log.Debugf("error : %v", err)
		}
		err = json.Unmarshal([]byte(ptm.RawPattern), &patternDetail)
		assert.Nil(t, err)
		patterns = append(patterns, &patternDetail)
	}

	pw := NewMockPatternServiceWrapper(patterns, nil)
	assert.Equal(t, 86, len(pw.patterns))

	var projectId int64 = int64(1)
	reqId := ""
	startEvent := "$session"
	endEvent := "Schedule A Demo Form"
	var startEventConstraints *P.EventConstraints
	var endEventConstraints *P.EventConstraints
	startEventConstraints, endEventConstraints = createUserEventConstraint()
	countType := P.COUNT_TYPE_PER_USER
	debugKey := ""
	debugParams := make(map[string]string)
	includedEvents := make(map[string]bool, 0)
	includedEventProperties := make(map[string]bool)
	includedUserProperties := make(map[string]bool)

	factors, err, _ := PW.FactorV1(reqId, projectId, startEvent, startEventConstraints, endEvent,
		endEventConstraints, countType, pw, debugKey, debugParams, includedEvents, includedEventProperties,
		includedUserProperties)

	assert.Nil(t, err)

	for _, ft := range factors.Insights {
		log.Debugf("factors insight: %d,%d, %v", len(ft.FactorsInsightsAttribute), len(ft.FactorsSubInsights), ft.FactorsInsightsAttribute)
	}

	assert.Equal(t, len(factors.Insights), 5)

}

func createUserEventConstraint() (*P.EventConstraints, *P.EventConstraints) {

	var startEventConstraints *P.EventConstraints
	var endEventConstraints *P.EventConstraints
	startEventConstraints = &P.EventConstraints{
		EPNumericConstraints:     []P.NumericConstraint{},
		EPCategoricalConstraints: []P.CategoricalConstraint{},
		UPNumericConstraints:     []P.NumericConstraint{},
		UPCategoricalConstraints: []P.CategoricalConstraint{},
	}
	endEventConstraints = &P.EventConstraints{
		EPNumericConstraints:     []P.NumericConstraint{},
		EPCategoricalConstraints: []P.CategoricalConstraint{},
		UPNumericConstraints:     []P.NumericConstraint{},
		UPCategoricalConstraints: []P.CategoricalConstraint{},
	}

	ep1 := P.CategoricalConstraint{PropertyName: "Country", PropertyValue: "US", Operator: P.EQUALS_OPERATOR_CONST}
	// ep2 := P.CategoricalConstraint{PropertyName: "$user_agent", PropertyValue: "Go-http-client/2.0", Operator: P.EQUALS_OPERATOR_CONST}
	startEventConstraints.UPCategoricalConstraints = append(startEventConstraints.UPCategoricalConstraints, ep1)
	endEventConstraints.UPCategoricalConstraints = append(endEventConstraints.UPCategoricalConstraints, ep1)
	// startEventConstraints.UPCategoricalConstraints = append(startEventConstraints.EPCategoricalConstraints, ep2)

	return startEventConstraints, endEventConstraints

}

func TestGenFreqProperties(t *testing.T) {

	patterns := make([]*P.Pattern, 0)
	filepath := "data/chunk_data.txt"

	file, err := os.Open(filepath)
	assert.Nil(t, err)
	scanner := bufio.NewScanner(file)
	buf := make([]byte, P.MAX_PATTERN_BYTES)
	scanner.Buffer(buf, P.MAX_PATTERN_BYTES)

	for scanner.Scan() {
		line := string(scanner.Text())
		var ptm PS.PatternWithMeta
		var patternDetail P.Pattern
		err := json.Unmarshal([]byte(line), &ptm)
		if err != nil {
			log.Debugf("error : %v", err)
		}
		err = json.Unmarshal([]byte(ptm.RawPattern), &patternDetail)
		assert.Nil(t, err)
		patterns = append(patterns, &patternDetail)
	}

	var testPattern *P.Pattern

	for _, pt := range patterns {
		if pt.EventNames[0] == "Schedule A Demo Form" {
			testPattern = pt
		}
	}

	tfr := testPattern.GenFrequentProperties()
	for _, t := range tfr.FrequentItemsets {
		log.Debugf("fitem :%v", t)
	}
	ap, err := tfr.GetPropertiesOfType("event")

	for _, s := range ap {
		log.Debugf("prop string :%s", s)
	}
	assert.Nil(t, err)

	assert.Equal(t, len(testPattern.FreqProps.FrequentItemsets), 302)

}

func TestFormatProperty(t *testing.T) {

	pos := []string{"1.$Country", "1.$country"}
	pos_true := []string{"$Country", "$country"}
	neg := []string{"a1.$Country", "a2.$Country"}

	for idx, sr := range pos {
		tstring := U.FormatProperty(sr)
		assert.Equal(t, pos_true[idx], tstring)
	}
	for _, sr := range neg {
		tstring := U.FormatProperty(sr)
		assert.Equal(t, tstring, sr)
	}
}

func TestFactorV1ExplainV2(t *testing.T) {

	filepath := "/usr/local/var/factors/cloud_storage/projects/1000002/models/1669197440190/chunks/chunk_1.txt"

	patterns := make([]*P.Pattern, 0)

	file, err := os.Open(filepath)
	assert.Nil(t, err)
	scanner := bufio.NewScanner(file)
	buf := make([]byte, P.MAX_PATTERN_BYTES)
	scanner.Buffer(buf, P.MAX_PATTERN_BYTES)

	for scanner.Scan() {
		line := string(scanner.Text())
		var ptm PS.PatternWithMeta
		var patternDetail P.Pattern
		err := json.Unmarshal([]byte(line), &ptm)
		if err != nil {
			log.Debugf("error : %v", err)
		}
		err = json.Unmarshal([]byte(ptm.RawPattern), &patternDetail)
		assert.Nil(t, err)
		patterns = append(patterns, &patternDetail)
	}
	log.Debugf("%v", patterns)

	pw := NewMockPatternServiceWrapper(patterns, nil)
	assert.Equal(t, 167, len(pw.patterns))

	var projectId int64 = int64(1000002)
	reqId := "13"
	startEvent := "www.supportlogic.io"
	endEvent := "www.supportlogic.io/blog"
	var startEventConstraints *P.EventConstraints
	var endEventConstraints *P.EventConstraints
	startEventConstraints, endEventConstraints = createUserEventConstraint()
	countType := P.COUNT_TYPE_PER_USER
	debugKey := ""
	debugParams := make(map[string]string)
	includedEvents := make(map[string]bool, 0)
	includedEventProperties := make(map[string]bool)
	includedUserProperties := make(map[string]bool)

	factors, err, _ := PW.FactorV1(reqId, projectId, startEvent, startEventConstraints, endEvent,
		endEventConstraints, countType, pw, debugKey, debugParams, includedEvents, includedEventProperties,
		includedUserProperties)

	assert.Nil(t, err)

	for _, ft := range factors.Insights {
		log.Debugf("factors insight: %d,%d, %v", len(ft.FactorsInsightsAttribute), len(ft.FactorsSubInsights), ft.FactorsInsightsAttribute)
	}

	assert.Equal(t, len(factors.Insights), 5)

}
