package tests

import (
	"bufio"
	"encoding/json"
	P "factors/pattern"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCountPatterns(t *testing.T) {
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count A -> B -> C, Count:3, OncePerUserCount:2, UserCount:2
	u1CTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	u1ETime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	u2CTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:01:00Z")
	u2ETime, _ := time.Parse(time.RFC3339, "2017-06-01T01:01:00Z")
	u1CTimestamp := u1CTime.Unix()
	u1ETimestamp := u1ETime.Unix()
	u2CTimestamp := u2CTime.Unix()
	u2ETimestamp := u2ETime.Unix()

	eventsInput := []P.CounterEventFormat{
		// User 1.
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "F", EventTimestamp: u1ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "G", EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventTimestamp: u1ETimestamp + (2 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "L", EventTimestamp: u1ETimestamp + (3 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (4 * 60), EventCardinality: uint(5)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventTimestamp: u1ETimestamp + (5 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (6 * 60), EventCardinality: uint(6)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventTimestamp: u1ETimestamp + (7 * 60), EventCardinality: uint(1)},
		// User 2.
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "F", EventTimestamp: u2ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (1 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (2 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "K", EventTimestamp: u2ETimestamp + (3 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp + (4 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "Z", EventTimestamp: u2ETimestamp + (5 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventTimestamp: u2ETimestamp + (6 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (7 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp + (8 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventTimestamp: u2ETimestamp + (9 * 60), EventCardinality: uint(2)},
	}
	eventsInputString := ""
	for _, event := range eventsInput {
		lineBytes, _ := json.Marshal(event)
		line := string(lineBytes)
		eventsInputString += fmt.Sprintf("%s\n", line)
	}

	scanner := bufio.NewScanner(strings.NewReader(eventsInputString))

	pABCEvents := []string{"A", "B", "C"}
	pLen := len(pABCEvents)
	pABC, _ := P.NewPattern(pABCEvents, nil)
	pAB, _ := P.NewPattern([]string{"A", "B"}, nil)
	pBC, _ := P.NewPattern([]string{"B", "C"}, nil)
	pAC, _ := P.NewPattern([]string{"B", "C"}, nil)
	pA, _ := P.NewPattern([]string{"A"}, nil)
	pB, _ := P.NewPattern([]string{"B"}, nil)
	pC, _ := P.NewPattern([]string{"C"}, nil)

	patterns := []*P.Pattern{pABC, pAB, pBC, pAC, pA, pB, pC}
	err := P.CountPatterns(scanner, patterns)
	assert.Nil(t, err)

	// Test pABC output.
	assert.Equal(t, uint(3), pABC.Count)
	assert.Equal(t, uint(2), pABC.UserCount)
	assert.Equal(t, uint(2), pABC.OncePerUserCount)
	assert.Equal(t, pLen, len(pABC.EventNames))
	for i := 0; i < pLen; i++ {
		assert.Equal(t, pABCEvents[i], pABC.EventNames[i])
	}
	// A-B-C occurs twice oncePerUser , with first A occurring after 3720s in User1 and
	// 3660s in User 2.
	// Repeats once before the  next B occurs in User2.
	assert.Equal(t, uint64(2), pABC.CardinalityRepeatTimings.Count())
	assert.Equal(t, float64((2.0+1.0)/2), pABC.CardinalityRepeatTimings.Mean()[0])
	assert.Equal(t, float64((1.0+2.0)/2), pABC.CardinalityRepeatTimings.Mean()[1])
	assert.Equal(t, float64((3720.0+3660.0)/2), pABC.CardinalityRepeatTimings.Mean()[2])

	// A-B-C occurs twice oncePerUser, with first B following first A after 120s in User1 and
	// 180 in User 2.
	// Repeats once before the  next C occurs in User1.
	/* Only start and end event are tracked currently.
	assert.Equal(t, float64(2), pABC.Timings[1].Count())
	assert.Equal(t, float64((120.0+180.0)/2), pABC.Timings[1].Mean())
	assert.Equal(t, float64((5.0+1.0)/2), pABC.EventCardinalities[1].Mean())
	assert.Equal(t, float64((2.0+1.0)/2), pABC.Repeats[1].Mean())*/

	// A-B-C occurs twice oncePerUser, with first C following first B after 180s in User1 and
	// 120s in User 2.
	// Last event always is counted once.
	assert.Equal(t, float64((1.0+1.0)/2), pABC.CardinalityRepeatTimings.Mean()[3])
	assert.Equal(t, float64((1.0+1.0)/2), pABC.CardinalityRepeatTimings.Mean()[4])
	assert.Equal(t, float64((180.0+120.0)/2), pABC.CardinalityRepeatTimings.Mean()[5])

	// Test output on other patterns.
	assert.Equal(t, uint(4), pAB.Count)
	assert.Equal(t, uint(2), pAB.OncePerUserCount)
	assert.Equal(t, uint(2), pAB.UserCount)

	assert.Equal(t, uint(3), pBC.Count)
	assert.Equal(t, uint(2), pBC.OncePerUserCount)
	assert.Equal(t, uint(2), pBC.UserCount)

	assert.Equal(t, uint(3), pAC.Count)
	assert.Equal(t, uint(2), pAC.OncePerUserCount)
	assert.Equal(t, uint(2), pAC.UserCount)

	assert.Equal(t, uint(5), pA.Count)
	assert.Equal(t, uint(2), pA.OncePerUserCount)
	assert.Equal(t, uint(2), pA.UserCount)

	assert.Equal(t, uint(4), pB.Count)
	assert.Equal(t, uint(2), pB.OncePerUserCount)
	assert.Equal(t, uint(2), pB.UserCount)

	assert.Equal(t, uint(3), pC.Count)
	assert.Equal(t, uint(2), pC.OncePerUserCount)
	assert.Equal(t, uint(2), pC.UserCount)
}

func TestGenCandidatesPair(t *testing.T) {
	// Mismatched length patterns.
	p1, _ := P.NewPattern([]string{"A"}, nil)
	p2, _ := P.NewPattern([]string{"A", "B"}, nil)
	c1, c2, ok := P.GenCandidatesPair(p1, p2, nil)
	assert.Nil(t, c1)
	assert.Nil(t, c2)
	assert.Equal(t, false, ok)

	// More than one different element.
	p1, _ = P.NewPattern([]string{"A", "B", "C"}, nil)
	p2, _ = P.NewPattern([]string{"A", "D", "E"}, nil)
	c1, c2, ok = P.GenCandidatesPair(p1, p2, nil)
	assert.Nil(t, c1)
	assert.Nil(t, c2)
	assert.Equal(t, false, ok)

	// Single element candidates.
	p1, _ = P.NewPattern([]string{"A"}, nil)
	p2, _ = P.NewPattern([]string{"B"}, nil)
	c1, c2, ok = P.GenCandidatesPair(p1, p2, nil)
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{"B", "A"}, c1.EventNames)
	assert.Equal(t, []string{"A", "B"}, c2.EventNames)

	// Different at the begining.
	p1, _ = P.NewPattern([]string{"B", "C"}, nil)
	p2, _ = P.NewPattern([]string{"A", "C"}, nil)
	c1, c2, ok = P.GenCandidatesPair(p1, p2, nil)
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{"A", "B", "C"}, c1.EventNames)
	assert.Equal(t, []string{"B", "A", "C"}, c2.EventNames)

	// Different at the end.
	p1, _ = P.NewPattern([]string{"A", "B", "D"}, nil)
	p2, _ = P.NewPattern([]string{"A", "B", "C"}, nil)
	c1, c2, ok = P.GenCandidatesPair(p1, p2, nil)
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{"A", "B", "C", "D"}, c1.EventNames)
	assert.Equal(t, []string{"A", "B", "D", "C"}, c2.EventNames)

	// Different in the middle.
	p1, _ = P.NewPattern([]string{"A", "C", "D"}, nil)
	p2, _ = P.NewPattern([]string{"A", "B", "D"}, nil)
	c1, c2, ok = P.GenCandidatesPair(p1, p2, nil)
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{"A", "B", "C", "D"}, c1.EventNames)
	assert.Equal(t, []string{"A", "C", "B", "D"}, c2.EventNames)
}

func TestGenLenThreeCandidatePatterns(t *testing.T) {
	// Not of length 2.
	pattern, _ := P.NewPattern([]string{"A", "X", "Z"}, nil)
	startPatterns := []*P.Pattern{}
	endPatterns := []*P.Pattern{}
	maxCandidates := 5
	cPatterns, err := P.GenLenThreeCandidatePatterns(
		pattern, startPatterns, endPatterns, maxCandidates, nil)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)

	// Mismatch event.
	pattern, _ = P.NewPattern([]string{"A", "Z"}, nil)
	mismatchPattern, _ := P.NewPattern([]string{"B", "X"}, nil)
	patterns1 := []*P.Pattern{mismatchPattern}
	patterns2 := []*P.Pattern{}
	maxCandidates = 5
	// Mismatch start event.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns1, patterns2, maxCandidates, nil)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)
	// Mismatch end event.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns2, patterns1, maxCandidates, nil)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)

	// Mismatch length.
	pattern, _ = P.NewPattern([]string{"A", "Z"}, nil)
	mismatchPattern, _ = P.NewPattern([]string{"A", "B", "Z"}, nil)
	patterns1 = []*P.Pattern{mismatchPattern}
	patterns2 = []*P.Pattern{}
	maxCandidates = 5
	// Mismatch in startPatterns.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns1, patterns2, maxCandidates, nil)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)
	// Mismatch in endPatterns.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns2, patterns1, maxCandidates, nil)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)

	// Candidate generation.
	pattern, _ = P.NewPattern([]string{"A", "Z"}, nil)
	maxCandidates = 3
	sp1, _ := P.NewPattern([]string{"A", "B"}, nil) // Skipped. BZ not found.
	sp2, _ := P.NewPattern([]string{"A", "Z"}, nil) // Skipped. Same as pattern.
	sp3, _ := P.NewPattern([]string{"A", "C"}, nil) // Skipped. ACZ Repeat.
	sp4, _ := P.NewPattern([]string{"A", "D"}, nil) // Skipped. ADZ Repeat.
	sp5, _ := P.NewPattern([]string{"A", "E"}, nil) // Skipped. AEZ Repeat.
	sp6, _ := P.NewPattern([]string{"A", "F"}, nil) // Ignored. Greater than maxCandidates.
	startPatterns = []*P.Pattern{sp1, sp2, sp3, sp4, sp5, sp6}
	ep1, _ := P.NewPattern([]string{"C", "Z"}, nil) // cPatterns[0] ACZ
	ep2, _ := P.NewPattern([]string{"D", "Z"}, nil) // cPatterns[1] ADZ
	ep3, _ := P.NewPattern([]string{"E", "Z"}, nil) // cPatterns[2] AEZ
	ep4, _ := P.NewPattern([]string{"F", "Z"}, nil) // Ignored. Greater than maxCandidates.
	endPatterns = []*P.Pattern{ep1, ep2, ep3, ep4}
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, startPatterns, endPatterns, maxCandidates, nil)
	assert.Nil(t, err)
	assert.Equal(t, maxCandidates, len(cPatterns))
	// Not expected in order.
	cMap := make(map[string]bool)
	for _, c := range cPatterns {
		cMap[c.String()] = true
	}
	assert.Equal(t, true, cMap["A,C,Z"])
	assert.Equal(t, true, cMap["A,D,Z"])
	assert.Equal(t, true, cMap["A,E,Z"])
}

func TestCollectAndCountEventsWithProperties(t *testing.T) {
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count A -> B -> C, Count:3, OncePerUserCount:2, UserCount:2
	u1CTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	u1ETime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	u2CTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:01:00Z")
	u2ETime, _ := time.Parse(time.RFC3339, "2017-06-01T01:01:00Z")
	u1CTimestamp := u1CTime.Unix()
	u1ETimestamp := u1ETime.Unix()
	u2CTimestamp := u2CTime.Unix()
	u2ETimestamp := u2ETime.Unix()

	eventsInput := []P.CounterEventFormat{
		// User 1.
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "F", EventProperties: map[string]interface{}{"ComNum": 1.0,
			"ComCat": "com1", "IgnoredKey": []string{"check"}}, EventTimestamp: u1ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "G", EventProperties: map[string]interface{}{"ComNum": 2.0,
			"ComCat": "com2", "IgnoredKey": []string{"check"}}, EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 3.0,
			"ComCat": "com3", "IgnoredKey": []string{"check"}, "ANum": 1, "ACat": "acat1"}, EventTimestamp: u1ETimestamp + (2 * 60),
			EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "L", EventProperties: map[string]interface{}{"ComNum": 1.0,
			"ComCat": "com1"}, EventTimestamp: u1ETimestamp + (3 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 2.0,
			"ComCat": "com2", "BNum": 1, "BCat": "bcat1"}, EventTimestamp: u1ETimestamp + (4 * 60), EventCardinality: uint(5)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 3.0,
			"ComCat": "com3", "ANum": 2, "ACat": "acat2"}, EventTimestamp: u1ETimestamp + (5 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 1.0,
			"ComCat": "com1", "BNum": 2, "BCat": "bcat2"}, EventTimestamp: u1ETimestamp + (6 * 60), EventCardinality: uint(6)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 2.0,
			"ComCat": "com2", "CNum": 1.0, "CCat": "ccat1"}, EventTimestamp: u1ETimestamp + (7 * 60), EventCardinality: uint(1)},
		// User 2.
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "F", EventProperties: map[string]interface{}{"ComNum": 3.0,
			"ComCat": "com3", "IgnoredKey": []string{"check"}}, EventTimestamp: u2ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 1.0,
			"ComCat": "com1", "ANum": 1, "ACat": "acat1"}, EventTimestamp: u2ETimestamp + (1 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 2.0,
			"ComCat": "com2", "ANum": 2, "ACat": "acat2"}, EventTimestamp: u2ETimestamp + (2 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "K", EventProperties: map[string]interface{}{"ComNum": 3.0,
			"ComCat": "com3"}, EventTimestamp: u2ETimestamp + (3 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 1.0,
			"ComCat": "com1", "BNum": 1, "BCat": "bcat1"}, EventTimestamp: u2ETimestamp + (4 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "Z", EventProperties: map[string]interface{}{"ComNum": 2.0,
			"ComCat": "com2"}, EventTimestamp: u2ETimestamp + (5 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 3.0,
			"ComCat": "com3", "CNum": 2.0, "CCat": "ccat2"}, EventTimestamp: u2ETimestamp + (6 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 1.0,
			"ComCat": "com1", "ANum": 1.0, "ACat": "acat1"}, EventTimestamp: u2ETimestamp + (7 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 2.0,
			"ComCat": "com2", "BNum": 2, "BCat": "bcat2"}, EventTimestamp: u2ETimestamp + (8 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 3.0,
			"ComCat": "com3", "CNum": 1.0, "CCat": "ccat1"}, EventTimestamp: u2ETimestamp + (9 * 60), EventCardinality: uint(2)},
	}
	eventsInputString := ""
	for _, event := range eventsInput {
		lineBytes, _ := json.Marshal(event)
		line := string(lineBytes)
		eventsInputString += fmt.Sprintf("%s\n", line)
	}

	scanner := bufio.NewScanner(strings.NewReader(eventsInputString))
	actualEventInfoMap := make(map[string]*P.PropertiesInfo)
	// Initialize.
	for _, eventName := range []string{"A", "B", "C", "F", "G", "K", "L", "Z"} {
		// Initialize info.
		actualEventInfoMap[eventName] = &P.PropertiesInfo{
			NumericPropertyKeys:          make(map[string]bool),
			CategoricalPropertyKeyValues: make(map[string]map[string]bool),
		}
	}
	userAndEventsInfo := P.UserAndEventsInfo{
		EventPropertiesInfoMap: &actualEventInfoMap,
	}
	err := P.CollectPropertiesInfo(scanner, &userAndEventsInfo)
	assert.Nil(t, err)

	expectedNumericKeys := map[string][]string{
		"A": []string{"ANum", "ComNum"},
		"B": []string{"BNum", "ComNum"},
		"C": []string{"CNum", "ComNum"},
		"F": []string{"ComNum"},
		"G": []string{"ComNum"},
		"K": []string{"ComNum"},
		"L": []string{"ComNum"},
		"Z": []string{"ComNum"},
	}
	expectedCategoricalKeyValues := map[string]map[string][]string{
		"A": map[string][]string{
			"ACat":   []string{"acat1", "acat2"},
			"ComCat": []string{"com1", "com2", "com3"},
		},
		"B": map[string][]string{
			"BCat":   []string{"bcat1", "bcat2"},
			"ComCat": []string{"com1", "com2"}, // Only those seen with this event.
		},
		"C": map[string][]string{
			"CCat":   []string{"ccat1", "ccat2"},
			"ComCat": []string{"com2", "com3"},
		},
		"F": map[string][]string{
			"ComCat": []string{"com1", "com3"},
		},
		"G": map[string][]string{
			"ComCat": []string{"com2"},
		},
		"K": map[string][]string{
			"ComCat": []string{"com3"},
		},
		"L": map[string][]string{
			"ComCat": []string{"com1"},
		},
		"Z": map[string][]string{
			"ComCat": []string{"com2"},
		},
	}
	// Numeric Keys.
	for e, keys := range expectedNumericKeys {
		eInfo, ok := actualEventInfoMap[e]
		assert.True(t, ok, fmt.Sprintf(
			"Missing event %s, actualEventInfoMap: %v", e, actualEventInfoMap))
		assert.Equal(t, len(keys), len(eInfo.NumericPropertyKeys),
			fmt.Sprintf("Mismatch numeric keys. Expected %v. Actual: %v",
				keys, eInfo.NumericPropertyKeys))
		for _, expectedKey := range keys {
			trueBool, ok := eInfo.NumericPropertyKeys[expectedKey]
			assert.True(t, ok, fmt.Sprintf("event %s, key %s", e, expectedKey))
			assert.True(t, trueBool, fmt.Sprintf("event %s, key %s", e, expectedKey))
		}
	}
	// Categorical key and values.
	for e, keyValues := range expectedCategoricalKeyValues {
		eInfo, ok := actualEventInfoMap[e]
		assert.True(t, ok, fmt.Sprintf("Missing event %s", e))
		assert.Equal(t, len(keyValues), len(eInfo.CategoricalPropertyKeyValues),
			fmt.Sprintf("Mismatch categorical keys. Expected %v. Actual: %v",
				keyValues, eInfo.CategoricalPropertyKeyValues))
		for expectedKey, expectedValues := range keyValues {
			actualValues, ok := eInfo.CategoricalPropertyKeyValues[expectedKey]
			assert.True(t, ok, fmt.Sprintf("event %s, key %s", e, expectedKey))
			assert.Equal(t, len(expectedValues), len(actualValues),
				fmt.Sprintf("event: %s, key: %s, expectedValues: %v, actualValues: %v",
					e, expectedKey, expectedValues, actualValues))
			for _, expectedValue := range expectedValues {
				trueBool, ok := actualValues[expectedValue]
				assert.True(t, ok, fmt.Sprintf("event %s, key %s, value %s",
					e, expectedKey, expectedValue))
				assert.True(t, trueBool, fmt.Sprintf("event %s, key %s, value: %s",
					e, expectedKey, expectedValue))
			}
		}
	}

	// Check counts.
	scanner = bufio.NewScanner(strings.NewReader(eventsInputString))

	pABCEvents := []string{"A", "B", "C"}
	pLen := len(pABCEvents)
	pABC, _ := P.NewPattern(pABCEvents, &userAndEventsInfo)
	pAB, _ := P.NewPattern([]string{"A", "B"}, &userAndEventsInfo)
	pBC, _ := P.NewPattern([]string{"B", "C"}, &userAndEventsInfo)
	pAC, _ := P.NewPattern([]string{"B", "C"}, &userAndEventsInfo)
	pA, _ := P.NewPattern([]string{"A"}, &userAndEventsInfo)
	pB, _ := P.NewPattern([]string{"B"}, &userAndEventsInfo)
	pC, _ := P.NewPattern([]string{"C"}, &userAndEventsInfo)

	patterns := []*P.Pattern{pABC, pAB, pBC, pAC, pA, pB, pC}
	err = P.CountPatterns(scanner, patterns)
	assert.Nil(t, err)

	// Test pABC output.
	assert.Equal(t, uint(3), pABC.Count)
	assert.Equal(t, uint(2), pABC.UserCount)
	assert.Equal(t, uint(2), pABC.OncePerUserCount)
	assert.Equal(t, pLen, len(pABC.EventNames))
	for i := 0; i < pLen; i++ {
		assert.Equal(t, pABCEvents[i], pABC.EventNames[i])
	}
	// A-B-C occurs twice oncePerUser , with first A occurring after 3720s in User1 and
	// 3660s in User 2.
	// Repeats once before the  next B occurs in User2.
	assert.Equal(t, uint64(2), pABC.CardinalityRepeatTimings.Count())
	assert.Equal(t, float64((2.0+1.0)/2), pABC.CardinalityRepeatTimings.Mean()[0])
	assert.Equal(t, float64((1.0+2.0)/2), pABC.CardinalityRepeatTimings.Mean()[1])
	assert.Equal(t, float64((3720.0+3660.0)/2), pABC.CardinalityRepeatTimings.Mean()[2])

	// A-B-C occurs twice oncePerUser, with first C following first B after 180s in User1 and
	// 120s in User 2.
	// Last event always is counted once.
	assert.Equal(t, float64((1.0+1.0)/2), pABC.CardinalityRepeatTimings.Mean()[3])
	assert.Equal(t, float64((1.0+1.0)/2), pABC.CardinalityRepeatTimings.Mean()[4])
	assert.Equal(t, float64((180.0+120.0)/2), pABC.CardinalityRepeatTimings.Mean()[5])

	// A-B-C occurs twice oncePerUser with the following six dimensional numerical
	// distribution.
	// 0.ANum: 1.0 and 1.0
	// 0.ComNum: 3.0 and 1.0
	// 1.BNum: 1.0 and 1.0
	// 1.ComNum: 2.0 and 1.0
	// 2.CNum: 1.0 and 2.0
	// 2.ComNum 2.0 and 3.0
	expectedMeanMap := map[string]float64{
		"0.ANum":   float64((1.0 + 1.0) / 2),
		"0.ComNum": float64((3.0 + 1.0) / 2),
		"1.BNum":   float64((1.0 + 1.0) / 2),
		"1.ComNum": float64((2.0 + 1.0) / 2),
		"2.CNum":   float64((1.0 + 2.0) / 2),
		"2.ComNum": float64((2.0 + 3.0) / 2),
	}
	actualMeanMap := pABC.EventNumericProperties.MeanMap()
	assert.Equal(t, expectedMeanMap, actualMeanMap)

	actualCdf := pABC.EventNumericProperties.CDFFromMap(
		map[string]float64{
			"0.ANum":   2.0,
			"0.ComNum": 2.0,
			"1.ComNum": 1.5,
		})
	assert.InDelta(t, actualCdf, 0.5, 0.01)

	// 0.ACat: "acat1" and "acat1"
	// 0.ComCat: "com3" and "com1"
	// 1.BCat: "bcat1" and "bcat1"
	// 1.ComCat: "com2" and "com1"
	// 2.CCat: "ccat1" and "ccat2"
	// 2.ComCat: "com2" and "com3"
	actualPdf, err := pABC.EventCategoricalProperties.PDFFromMap(
		map[string]string{
			"0.ACat":   "acat1",
			"0.ComCat": "com3",
			"1.BCat":   "bcat1",
			"1.ComCat": "com2",
			"2.CCat":   "ccat1",
			"2.ComCat": "com2",
		})
	assert.Nil(t, err, fmt.Sprintf("Error: %v", err))
	assert.InDelta(t, actualPdf, 0.5, 0.01)

	actualPdf, err = pABC.EventCategoricalProperties.PDFFromMap(
		map[string]string{
			"0.ACat":   "acat1",
			"0.ComCat": "com1",
			"1.BCat":   "bcat1",
			"1.ComCat": "com1",
			"2.CCat":   "ccat2",
			"2.ComCat": "com3",
		})
	assert.Nil(t, err, fmt.Sprintf("Error: %v", err))
	assert.InDelta(t, actualPdf, 0.5, 0.01)

	// Test output on other patterns.
	assert.Equal(t, uint(4), pAB.Count)
	assert.Equal(t, uint(2), pAB.OncePerUserCount)
	assert.Equal(t, uint(2), pAB.UserCount)

	assert.Equal(t, uint(3), pBC.Count)
	assert.Equal(t, uint(2), pBC.OncePerUserCount)
	assert.Equal(t, uint(2), pBC.UserCount)

	assert.Equal(t, uint(3), pAC.Count)
	assert.Equal(t, uint(2), pAC.OncePerUserCount)
	assert.Equal(t, uint(2), pAC.UserCount)

	assert.Equal(t, uint(5), pA.Count)
	assert.Equal(t, uint(2), pA.OncePerUserCount)
	assert.Equal(t, uint(2), pA.UserCount)

	assert.Equal(t, uint(4), pB.Count)
	assert.Equal(t, uint(2), pB.OncePerUserCount)
	assert.Equal(t, uint(2), pB.UserCount)

	assert.Equal(t, uint(3), pC.Count)
	assert.Equal(t, uint(2), pC.OncePerUserCount)
	assert.Equal(t, uint(2), pC.UserCount)

	// Test GetOncePerUserCount with constraints.
	count, err := pABC.GetOncePerUserCount(nil)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), count)

	patternConstraints := make([]P.EventConstraints, 3)
	count, err = pABC.GetOncePerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), count)

	patternConstraints = make([]P.EventConstraints, 3)
	patternConstraints[0] = P.EventConstraints{
		NumericConstraints: []P.NumericConstraint{
			P.NumericConstraint{
				PropertyName: "ANum",
				LowerBound:   -math.MaxFloat64,
				UpperBound:   2.0,
			},
			P.NumericConstraint{
				PropertyName: "ComNum",
				LowerBound:   -math.MaxFloat64,
				UpperBound:   2.0,
			},
		},
	}
	patternConstraints[1] = P.EventConstraints{
		NumericConstraints: []P.NumericConstraint{
			P.NumericConstraint{
				PropertyName: "ComNum",
				LowerBound:   -math.MaxFloat64,
				UpperBound:   2.0,
			},
		},
	}
	count, err = pABC.GetOncePerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)

	patternConstraints = make([]P.EventConstraints, 3)
	patternConstraints[0] = P.EventConstraints{
		NumericConstraints: []P.NumericConstraint{
			P.NumericConstraint{
				PropertyName: "ANum",
				LowerBound:   -0.5,
				UpperBound:   +0.5,
			},
			P.NumericConstraint{
				PropertyName: "ComNum",
				LowerBound:   -0.5,
				UpperBound:   +0.5,
			},
		},
	}
	count, err = pABC.GetOncePerUserCount(patternConstraints)
	assert.Nil(t, err)
	// This combination of 0.Anum=1 and 0.ComNum=1 does not occur together,
	// though they take individually these values.
	assert.Equal(t, uint(0), count)

	patternConstraints = make([]P.EventConstraints, 3)
	// Below categorical combination occurs in the first occurrence.
	patternConstraints[1] = P.EventConstraints{
		CategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "BCat",
				PropertyValue: "bcat1",
			},
		},
	}
	patternConstraints[2] = P.EventConstraints{
		CategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "ComCat",
				PropertyValue: "com2",
			},
		},
	}
	count, err = pABC.GetOncePerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)
}

// TODO(aravind): Add tests for genLenThreeSegmentedCandidates and genSegmentedCandidates in run_pattern_mine.go
