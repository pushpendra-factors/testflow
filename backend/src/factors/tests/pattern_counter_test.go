package tests

import (
	"bufio"
	"encoding/json"
	P "factors/pattern"
	U "factors/util"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const INFHEX = 0x8000000000000000

func TestCountPatterns(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count A -> B -> C, Count:3, OncePerUserCount:2, UserCount:2
	countOccurFlag := true
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
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(4)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventTimestamp: u1ETimestamp + (2 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventTimestamp: u1ETimestamp + (3 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (4 * 60), EventCardinality: uint(5)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventTimestamp: u1ETimestamp + (5 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (6 * 60), EventCardinality: uint(6)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventTimestamp: u1ETimestamp + (7 * 60), EventCardinality: uint(2)},
		// User 2.
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (1 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (2 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventTimestamp: u2ETimestamp + (3 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp + (4 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "Z", EventTimestamp: u2ETimestamp + (5 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventTimestamp: u2ETimestamp + (6 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (7 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp + (8 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventTimestamp: u2ETimestamp + (9 * 60), EventCardinality: uint(3)},
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
	pAC, _ := P.NewPattern([]string{"A", "C"}, nil)
	pA, _ := P.NewPattern([]string{"A"}, nil)
	pB, _ := P.NewPattern([]string{"B"}, nil)
	pC, _ := P.NewPattern([]string{"C"}, nil)

	patterns := []*P.Pattern{pABC, pAB, pBC, pAC, pA, pB, pC}

	var cProps P.CountAlgoProperties
	cProps.Counting_version = 1
	erronFalse := P.CountPatterns(project.ID, scanner, patterns, countOccurFlag, cProps)
	assert.Nil(t, erronFalse)

	// Test pABC output.
	assert.Equal(t, uint(0), pABC.PerOccurrenceCount, "pABC.PerOccurrenceCount")
	assert.Equal(t, uint(2), pABC.TotalUserCount, "pABC.TotalUserCount")
	assert.Equal(t, uint(2), pABC.PerUserCount, "pABC.PerUserCount")
	assert.Equal(t, pLen, len(pABC.EventNames))
	for i := 0; i < pLen; i++ {
		assert.Equal(t, pABCEvents[i], pABC.EventNames[i])
	}

	// A-B-C occurs twice OncePerUser with the following Generic Properties.

	// A: firstSeenOccurrenceCount -> 2 and 1.
	// A: lastSeenOccurrenceCount -> 3 and 3.
	// A: firstSeenTime -> user1CreatedTime+1hour+120seconds and user2CreatedTime+1hour+60seconds.
	// A: lastSeenTime -> user1CreatedTime+1hour+300seconds and user2CreatedTime+1hour+420seconds.
	// A: firstSeenSinceUserJoin -> 1hour+120seconds and 1hour+60seconds.
	// A: lastSeenSinceUserJoin -> 1hour+300seconds and 1hour+420seconds.

	// B: firstSeenOccurrenceCount -> 5 and 2.
	// B: lastSeenOccurrenceCount -> 6 and 3.
	// B: firstSeenTime -> user1CreatedTime+1hour+240seconds and user2CreatedTime+1hour+240seconds.
	// B: lastSeenTime -> user1CreatedTime+1hour+360seconds and user2CreatedTime+1hour+480seconds.
	// B: firstSeenSinceUserJoin -> 1hour+240seconds and 1hour+240seconds.
	// B: lastSeenSinceUserJoin -> 1hour+360seconds and 1hour+480seconds.

	// C: firstSeenOccurrenceCount -> 2 and 2.
	// C: lastSeenOccurrenceCount -> 2 and 3.
	// C: firstSeenTime -> user1CreatedTime+1hour+420seconds and user2CreatedTime+1hour+360seconds.
	// C: lastSeenTime -> user1CreatedTime+1hour+420seconds and user2CreatedTime+1hour+540seconds.
	// C: firstSeenSinceUserJoin -> 1hour+240seconds and 1hour+240seconds.
	// C: lastSeenSinceUserJoin -> 1hour+360seconds and 1hour+540seconds.
	assert.Equal(t, uint64(2), pABC.GenericPropertiesHistogram.Count())
	expectedMeanMap := map[string]float64{
		U.UP_JOIN_TIME: float64((u1CTimestamp + u2CTimestamp) / 2.0),
		// Event A Generic Properties.
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 1.0) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((3.0 + 3.0) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 120 + u2CTimestamp + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 300 + u2CTimestamp + 3600 + 420) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 120 + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 300 + 3600 + 420) / 2),

		// Event B Generic Properties.
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((5.0 + 2.0) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((6.0 + 3.0) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 240 + u2CTimestamp + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 360 + u2CTimestamp + 3600 + 480) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 240 + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 360 + 3600 + 480) / 2),

		// Event C Generic Properties.
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 2.0) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((2.0 + 3.0) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 420 + u2CTimestamp + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 420 + u2CTimestamp + 3600 + 540) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 420 + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 420 + 3600 + 540) / 2),
	}
	actualMeanMap := pABC.GenericPropertiesHistogram.MeanMap()
	for k, expectedMean := range expectedMeanMap {
		assert.Equal(t, expectedMean, actualMeanMap[k], fmt.Sprintf("Failed for Key: %s", k))
	}

	// Test output on other patterns.
	assert.Equal(t, uint(0), pAB.PerOccurrenceCount)
	assert.Equal(t, uint(2), pAB.PerUserCount)
	assert.Equal(t, uint(2), pAB.TotalUserCount)

	assert.Equal(t, uint(0), pBC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pBC.PerUserCount)
	assert.Equal(t, uint(2), pBC.TotalUserCount)

	assert.Equal(t, uint(0), pAC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pAC.PerUserCount)
	assert.Equal(t, uint(2), pAC.TotalUserCount)

	assert.Equal(t, uint(0), pA.PerOccurrenceCount)
	assert.Equal(t, uint(2), pA.PerUserCount)
	assert.Equal(t, uint(2), pA.TotalUserCount)

	assert.Equal(t, uint(0), pB.PerOccurrenceCount)
	assert.Equal(t, uint(2), pB.PerUserCount)
	assert.Equal(t, uint(2), pB.TotalUserCount)

	assert.Equal(t, uint(0), pC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pC.PerUserCount)
	assert.Equal(t, uint(2), pC.TotalUserCount)

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
		pattern, startPatterns, endPatterns, maxCandidates, nil, nil)
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
		pattern, patterns1, patterns2, maxCandidates, nil, nil)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)
	// Mismatch end event.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns2, patterns1, maxCandidates, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, cPatterns)

	// Mismatch length.
	pattern, _ = P.NewPattern([]string{"A", "Z"}, nil)
	mismatchPattern, _ = P.NewPattern([]string{"A", "B", "Z"}, nil)
	patterns1 = []*P.Pattern{mismatchPattern}
	patterns2 = []*P.Pattern{}
	maxCandidates = 5
	// Mismatch in startPatterns.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns1, patterns2, maxCandidates, nil, nil)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)
	// Mismatch in endPatterns.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns2, patterns1, maxCandidates, nil, nil)
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
		pattern, startPatterns, endPatterns, maxCandidates, nil, nil)
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

func TestGenLenThreeCandidateCyclic(t *testing.T) {
	// Not of length 2.
	pattern, _ := P.NewPattern([]string{"A", "X", "Z"}, nil)
	cycEvents := []string{"A", "B", "D"}
	startPatterns := []*P.Pattern{}
	endPatterns := []*P.Pattern{}
	maxCandidates := 5
	cPatterns, err := P.GenLenThreeCandidatePatterns(
		pattern, startPatterns, endPatterns, maxCandidates, nil, cycEvents)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)

	// two Candidate
	pattern, _ = P.NewPattern([]string{"A", "Z"}, nil)
	cycEvents = []string{"A", "Z"}
	startPatterns = []*P.Pattern{}
	endPatterns = []*P.Pattern{}
	maxCandidates = 5
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, startPatterns, endPatterns, maxCandidates, nil, cycEvents)
	assert.Nil(t, err)
	assert.NotNil(t, cPatterns)
	assert.Equal(t, 2, len(cPatterns))
	cMap := make(map[string]bool)
	for _, c := range cPatterns {
		cMap[c.String()] = true
	}
	assert.Equal(t, true, cMap["A,A,Z"])
	assert.Equal(t, true, cMap["A,Z,Z"])

	// Mismatch event.
	pattern, _ = P.NewPattern([]string{"A", "Z"}, nil)
	mismatchPattern, _ := P.NewPattern([]string{"B", "X"}, nil)
	patterns1 := []*P.Pattern{mismatchPattern}
	patterns2 := []*P.Pattern{}
	maxCandidates = 5
	// Mismatch start event.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns1, patterns2, maxCandidates, nil, cycEvents)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)
	// Mismatch end event.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns2, patterns1, maxCandidates, nil, cycEvents)
	assert.Nil(t, err)
	assert.NotNil(t, cPatterns)

	// Mismatch length.
	pattern, _ = P.NewPattern([]string{"A", "Z"}, nil)
	mismatchPattern, _ = P.NewPattern([]string{"A", "B", "Z"}, nil)
	patterns1 = []*P.Pattern{mismatchPattern}
	patterns2 = []*P.Pattern{}
	maxCandidates = 5
	// Mismatch in startPatterns.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns1, patterns2, maxCandidates, nil, cycEvents)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)
	// Mismatch in endPatterns.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns2, patterns1, maxCandidates, nil, cycEvents)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)

	// Candidate generation.
	pattern, _ = P.NewPattern([]string{"A", "Z"}, nil)
	cycEvents = []string{"A", "B", "Z"}
	maxCandidates = 5
	sp1, _ := P.NewPattern([]string{"A", "D"}, nil) // Skipped. BZ not found.
	sp2, _ := P.NewPattern([]string{"A", "Z"}, nil) // Skipped. Same as pattern.
	sp3, _ := P.NewPattern([]string{"A", "C"}, nil) // Skipped. ACZ Repeat.
	sp4, _ := P.NewPattern([]string{"A", "D"}, nil) // Skipped. ADZ Repeat.
	sp5, _ := P.NewPattern([]string{"A", "E"}, nil) // Skipped. AEZ Repeat.
	sp6, _ := P.NewPattern([]string{"A", "F"}, nil) // Ignored. Greater than maxCandidates.
	startPatterns = []*P.Pattern{sp1, sp2, sp3, sp4, sp5, sp6}
	ep1, _ := P.NewPattern([]string{"C", "Z"}, nil) // cPatterns[0] ACZ
	ep2, _ := P.NewPattern([]string{"D", "Z"}, nil) // cPatterns[1] ADZ
	ep3, _ := P.NewPattern([]string{"E", "Z"}, nil) // cPatterns[2] AEZ
	ep4, _ := P.NewPattern([]string{"B", "Z"}, nil) // Ignored. Greater than maxCandidates.
	endPatterns = []*P.Pattern{ep1, ep2, ep3, ep4}
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, startPatterns, endPatterns, maxCandidates, nil, cycEvents)
	assert.Nil(t, err)
	assert.Equal(t, maxCandidates, len(cPatterns))
	// Not expected in order.
	cMap = make(map[string]bool)
	for _, c := range cPatterns {
		cMap[c.String()] = true
	}
	assert.Equal(t, true, cMap["A,A,Z"])
	assert.Equal(t, true, cMap["A,Z,Z"])
	assert.Equal(t, true, cMap["A,E,Z"])
	assert.Equal(t, true, cMap["A,D,Z"])
	assert.Equal(t, true, cMap["A,C,Z"])

}

func TestCollectAndCountEventsWithProperties(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count A -> B -> C, Count:3, OncePerUserCount:2, UserCount:2
	// Add False pcountOccur TestCase

	countOccurFlag := true
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
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "F", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1", "IgnoredKey": []string{"check"}}, EventTimestamp: u1ETimestamp, EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2", "IgnoredKey": []string{"check"}}, EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(4),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3", "IgnoredKey": []string{"check"}, "ANum": 1, "ACat": "acat1"}, EventTimestamp: u1ETimestamp + (2 * 60),
			EventCardinality: uint(2), UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1"}, EventTimestamp: u1ETimestamp + (3 * 60), EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2", "BNum": 1, "BCat": "bcat1"}, EventTimestamp: u1ETimestamp + (4 * 60), EventCardinality: uint(5),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3", "ANum": 2, "ACat": "acat2"}, EventTimestamp: u1ETimestamp + (5 * 60), EventCardinality: uint(3),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1", "BNum": 2, "BCat": "bcat2"}, EventTimestamp: u1ETimestamp + (6 * 60), EventCardinality: uint(6),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2", "CNum": 1.0, "CCat": "ccat1"}, EventTimestamp: u1ETimestamp + (7 * 60), EventCardinality: uint(2),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Paid", "age": 20.0},
		},
		// User 2.
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3", "IgnoredKey": []string{"check"}}, EventTimestamp: u2ETimestamp, EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1", "ANum": 1, "ACat": "acat1"}, EventTimestamp: u2ETimestamp + (1 * 60), EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2", "ANum": 2, "ACat": "acat2"}, EventTimestamp: u2ETimestamp + (2 * 60), EventCardinality: uint(2),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3"}, EventTimestamp: u2ETimestamp + (3 * 60), EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0}},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1", "BNum": 1, "BCat": "bcat1"}, EventTimestamp: u2ETimestamp + (4 * 60), EventCardinality: uint(2),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "Z", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2"}, EventTimestamp: u2ETimestamp + (5 * 60), EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3", "CNum": 2.0, "CCat": "ccat2"}, EventTimestamp: u2ETimestamp + (6 * 60), EventCardinality: uint(2),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Paid", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1", "ANum": 1.0, "ACat": "acat1"}, EventTimestamp: u2ETimestamp + (7 * 60), EventCardinality: uint(3),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Paid", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2", "BNum": 2, "BCat": "bcat2"}, EventTimestamp: u2ETimestamp + (8 * 60), EventCardinality: uint(3),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Paid", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3", "CNum": 1.0, "CCat": "ccat1"}, EventTimestamp: u2ETimestamp + (9 * 60), EventCardinality: uint(3),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Paid", "age": 30.0},
		},
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
		UserPropertiesInfo: &P.PropertiesInfo{
			NumericPropertyKeys:          make(map[string]bool),
			CategoricalPropertyKeyValues: make(map[string]map[string]bool),
		},
		EventPropertiesInfoMap: &actualEventInfoMap,
	}
	_, err = P.CollectPropertiesInfo(project.ID, scanner, &userAndEventsInfo)
	assert.Nil(t, err)

	expectedNumericKeys := map[string][]string{
		"A": []string{"ANum", "ComNum"},
		"B": []string{"BNum", "ComNum"},
		"C": []string{"CNum", "ComNum"},
		"F": []string{"ComNum"},
		"Z": []string{"ComNum"},
	}
	expectedCategoricalKeyValues := map[string]map[string][]string{
		"A": map[string][]string{
			"ACat":   []string{"acat1", "acat2"},
			"ComCat": []string{"com1", "com2", "com3"},
		},
		"B": map[string][]string{
			"BCat":   []string{"bcat1", "bcat2"},
			"ComCat": []string{"com1", "com2", "com3"},
		},
		"C": map[string][]string{
			"CCat":   []string{"ccat1", "ccat2"},
			"ComCat": []string{"com1", "com2", "com3"},
		},
		"F": map[string][]string{
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
			fmt.Sprintf("Mismatch numeric keys. event: %s, Expected %v. Actual: %v",
				e, keys, eInfo.NumericPropertyKeys))
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
	pAC, _ := P.NewPattern([]string{"A", "C"}, &userAndEventsInfo)
	pA, _ := P.NewPattern([]string{"A"}, &userAndEventsInfo)
	pB, _ := P.NewPattern([]string{"B"}, &userAndEventsInfo)
	pC, _ := P.NewPattern([]string{"C"}, &userAndEventsInfo)

	patterns := []*P.Pattern{pABC, pAB, pBC, pAC, pA, pB, pC}
	var cProps P.CountAlgoProperties
	cProps.Counting_version = 1
	erronFalse := P.CountPatterns(project.ID, scanner, patterns, countOccurFlag, cProps)
	assert.Nil(t, erronFalse)

	// A-B-C occurs twice PerUser with the following Generic Properties.

	// A: firstSeenOccurrenceCount -> 2 and 1.
	// A: lastSeenOccurrenceCount -> 3 and 3.
	// A: firstSeenTime -> user1CreatedTime+1hour+120seconds and user2CreatedTime+1hour+60seconds.
	// A: lastSeenTime -> user1CreatedTime+1hour+300seconds and user2CreatedTime+1hour+420seconds.
	// A: firstSeenSinceUserJoin -> 1hour+120seconds and 1hour+60seconds.
	// A: lastSeenSinceUserJoin -> 1hour+300seconds and 1hour+420seconds.

	// B: firstSeenOccurrenceCount -> 5 and 2.
	// B: lastSeenOccurrenceCount -> 6 and 3.
	// B: firstSeenTime -> user1CreatedTime+1hour+240seconds and user2CreatedTime+1hour+240seconds.
	// B: lastSeenTime -> user1CreatedTime+1hour+360seconds and user2CreatedTime+1hour+480seconds.
	// B: firstSeenSinceUserJoin -> 1hour+240seconds and 1hour+240seconds.
	// B: lastSeenSinceUserJoin -> 1hour+360seconds and 1hour+480seconds.

	// C: firstSeenOccurrenceCount -> 2 and 2.
	// C: lastSeenOccurrenceCount -> 2 and 3.
	// C: firstSeenTime -> user1CreatedTime+1hour+420seconds and user2CreatedTime+1hour+360seconds.
	// C: lastSeenTime -> user1CreatedTime+1hour+420seconds and user2CreatedTime+1hour+540seconds.
	// C: firstSeenSinceUserJoin -> 1hour+240seconds and 1hour+240seconds.
	// C: lastSeenSinceUserJoin -> 1hour+360seconds and 1hour+540seconds.
	assert.Equal(t, uint(0), pABC.PerOccurrenceCount, "pABC.PerOccurrenceCount")
	assert.Equal(t, uint(2), pABC.TotalUserCount)
	assert.Equal(t, uint(2), pABC.PerUserCount)
	assert.Equal(t, pLen, len(pABC.EventNames))
	for i := 0; i < pLen; i++ {
		assert.Equal(t, pABCEvents[i], pABC.EventNames[i])
	}
	assert.Equal(t, uint64(2), pABC.GenericPropertiesHistogram.Count())

	expectedMeanMap := map[string]float64{
		U.UP_JOIN_TIME: float64((u1CTimestamp + u2CTimestamp) / 2.0),
		// Event A Generic Properties.
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 1.0) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((3.0 + 3.0) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 120 + u2CTimestamp + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 300 + u2CTimestamp + 3600 + 420) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 120 + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 300 + 3600 + 420) / 2),

		// Event B Generic Properties.
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((5.0 + 2.0) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((6.0 + 3.0) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 240 + u2CTimestamp + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 360 + u2CTimestamp + 3600 + 480) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 240 + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 360 + 3600 + 480) / 2),

		// Event C Generic Properties.
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 2.0) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((2.0 + 3.0) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 420 + u2CTimestamp + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 420 + u2CTimestamp + 3600 + 540) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 420 + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 420 + 3600 + 540) / 2),
	}
	actualMeanMap := pABC.GenericPropertiesHistogram.MeanMap()

	for k, expectedMean := range expectedMeanMap {
		assert.Equal(t, expectedMean, actualMeanMap[k], fmt.Sprintf("Failed for Key: %s", k))
	}
	// A-B-C occurs twice oncePerUser with the following six dimensional event numerical
	// distribution.
	// 0.ANum: 1.0 and 1.0
	// 0.ComNum: 3.0 and 1.0
	// 1.BNum: 1.0 and 1.0
	// 1.ComNum: 2.0 and 1.0
	// 2.CNum: 1.0 and 2.0
	// 2.ComNum 2.0 and 3.0
	expectedMeanMap = map[string]float64{
		"0.ANum":   float64((1.0 + 1.0) / 2),
		"0.ComNum": float64((3.0 + 1.0) / 2),
		"1.BNum":   float64((1.0 + 1.0) / 2),
		"1.ComNum": float64((2.0 + 1.0) / 2),
		"2.CNum":   float64((1.0 + 2.0) / 2),
		"2.ComNum": float64((2.0 + 3.0) / 2),
	}
	actualMeanMap = pABC.PerUserEventNumericProperties.MeanMap()
	assert.Equal(t, expectedMeanMap, actualMeanMap)

	actualCdf := pABC.PerUserEventNumericProperties.CDFFromMap(
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
	actualPdf, err := pABC.PerUserEventCategoricalProperties.PDFFromMap(
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

	// A-B-C occurs twice oncePerUser with the following six dimensional user numerical
	// distribution.
	// 0.age: 20.0 and 30.0
	// 1.age: 20.0 and 30.0
	// 2.age: 20.0 and 30.0
	expectedMeanMap = map[string]float64{
		"0.age": float64((20.0 + 30.0) / 2),
		"1.age": float64((20.0 + 30.0) / 2),
		"2.age": float64((20.0 + 30.0) / 2),
	}
	actualMeanMap = pABC.PerUserUserNumericProperties.MeanMap()
	assert.Equal(t, expectedMeanMap, actualMeanMap)

	actualCdf = pABC.PerUserUserNumericProperties.CDFFromMap(
		map[string]float64{
			"0.age": 25.0,
			"1.age": 25.0,
			"2.age": 25.0,
		})
	assert.InDelta(t, actualCdf, 0.5, 0.01)

	// ABC occurs twice with U1 country India and U2 country USA.
	// Payment status changes from Free to Paid on first occurrence of C.
	// 0.$country: "India" and "USA"
	// 0.paymentStatus: "Free" and "Free"
	// 1.$country: "India" and "USA"
	// 1.paymentStatus: "Free" and "Free"
	// 2.$country: "India" and "USA"
	// 2.paymentStatus: "Paid" and "Paid"
	actualPdf, err = pABC.PerUserUserCategoricalProperties.PDFFromMap(
		map[string]string{
			"0.$country":      "USA",
			"0.paymentStatus": "Free",
			"1.$country":      "USA",
			"1.paymentStatus": "Free",
			"2.$country":      "USA",
			"2.paymentStatus": "Paid",
		})
	assert.Nil(t, err, fmt.Sprintf("Error: %v", err))
	assert.InDelta(t, actualPdf, 0.5, 0.01)

	// A-B-C occurs thrice across users with the following six dimensional event numerical
	// distribution.
	// 0.ANum: 1.0 and 1.0 and 1.0
	// 0.ComNum: 3.0 and 1.0 and 1.0
	// 1.BNum: 1.0 and 1.0 and 2.0
	// 1.ComNum: 2.0 and 1.0 and 2.0
	// 2.CNum: 1.0 and 2.0 and 1.0
	// 2.ComNum 2.0 and 3.0 and 3.0
	expectedMeanMap = map[string]float64{
		"0.ANum":   float64((1.0 + 1.0 + 1.0) / 3),
		"0.ComNum": float64((3.0 + 1.0 + 1.0) / 3),
		"1.BNum":   float64((1.0 + 1.0 + 2.0) / 3),
		"1.ComNum": float64((2.0 + 1.0 + 2.0) / 3),
		"2.CNum":   float64((1.0 + 2.0 + 1.0) / 3),
		"2.ComNum": float64((2.0 + 3.0 + 3.0) / 3),
	}
	fmt.Println(pABC)
	actualMeanMap = pABC.PerOccurrenceEventNumericProperties.MeanMap()

	// assert.Equal(t, expectedMeanMap, actualMeanMap)
	actualCdf = pABC.PerOccurrenceEventNumericProperties.CDFFromMap(
		map[string]float64{
			"0.ANum":   2.0,
			"0.ComNum": 2.0,
			"1.ComNum": 1.5,
		})
	// assert.InDelta(t, actualCdf, 0.33, 0.01)// removed as per Occurence counts are not counted

	// 0.ACat: "acat1" and "acat1" and "acat1"
	// 0.ComCat: "com3" and "com1" and "com1"
	// 1.BCat: "bcat1" and "bcat1" and "bcat2"
	// 1.ComCat: "com2" and "com1" and "com2"
	// 2.CCat: "ccat1" and "ccat2" and "ccat1"
	// 2.ComCat: "com2" and "com3" and "com3"

	actualPdf, err = pABC.PerOccurrenceEventCategoricalProperties.PDFFromMap(
		map[string]string{
			"0.ACat":   "acat1",
			"0.ComCat": "com3",
			"1.BCat":   "bcat1",
			"1.ComCat": "com2",
			"2.CCat":   "ccat1",
			"2.ComCat": "com2",
		})
	assert.Nil(t, err, fmt.Sprintf("Error: %v", err))
	// assert.InDelta(t, actualPdf, 0.33, 0.01) // removed as per Occurence counts are not counted

	// A-B-C occurs thrice with the following six dimensional user numerical
	// distribution.
	// 0.age: 20.0 and 30.0 and 30.0
	// 1.age: 20.0 and 30.0 and 30.0
	// 2.age: 20.0 and 30.0 and 30.0
	expectedMeanMap = map[string]float64{
		"0.age": float64((20.0 + 30.0 + 30.0) / 3),
		"1.age": float64((20.0 + 30.0 + 30.0) / 3),
		"2.age": float64((20.0 + 30.0 + 30.0) / 3),
	}
	actualMeanMap = pABC.PerOccurrenceUserNumericProperties.MeanMap()
	// assert.Equal(t, expectedMeanMap, actualMeanMap) // removed as per Occurence counts are not counted

	actualCdf = pABC.PerOccurrenceUserNumericProperties.CDFFromMap(
		map[string]float64{
			"0.age": 25.0,
			"1.age": 25.0,
			"2.age": 25.0,
		})
	// assert.InDelta(t, actualCdf, 0.33, 0.01) // removed as per Occurence counts are not counted

	// ABC occurs thrice, once with U1 country India and twice with U2 country USA.
	// Payment status changes from Free to Paid on first occurrence of C.
	// 0.$country: "India" and "USA" and "USA"
	// 0.paymentStatus: "Free" and "Free" and "Paid"
	// 1.$country: "India" and "USA" and "USA"
	// 1.paymentStatus: "Free" and "Free" and "Paid"
	// 2.$country: "India" and "USA" and "USA"
	// 2.paymentStatus: "Paid" and "Paid" and "Paid"

	actualPdf, err = pABC.PerOccurrenceUserCategoricalProperties.PDFFromMap(
		map[string]string{
			"0.$country":      "USA",
			"0.paymentStatus": "Free",
			"1.$country":      "USA",
			"1.paymentStatus": "Free",
			"2.$country":      "USA",
			"2.paymentStatus": "Paid",
		})

	assert.Nil(t, err, fmt.Sprintf("Error: %v", err))
	// assert.InDelta(t, actualPdf, 0.33, 0.01) // removed as per Occurence counts are not counted

	// Test output on other patterns.
	assert.Equal(t, uint(0), pAB.PerOccurrenceCount)
	assert.Equal(t, uint(2), pAB.PerUserCount)
	assert.Equal(t, uint(2), pAB.TotalUserCount)

	assert.Equal(t, uint(0), pBC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pBC.PerUserCount)
	assert.Equal(t, uint(2), pBC.TotalUserCount)

	assert.Equal(t, uint(0), pAC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pAC.PerUserCount)
	assert.Equal(t, uint(2), pAC.TotalUserCount)

	assert.Equal(t, uint(0), pA.PerOccurrenceCount)
	assert.Equal(t, uint(2), pA.PerUserCount)
	assert.Equal(t, uint(2), pA.TotalUserCount)

	assert.Equal(t, uint(0), pB.PerOccurrenceCount)
	assert.Equal(t, uint(2), pB.PerUserCount)
	assert.Equal(t, uint(2), pB.TotalUserCount)

	assert.Equal(t, uint(0), pC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pC.PerUserCount)
	assert.Equal(t, uint(2), pC.TotalUserCount)

	// Test GetPerUserCount and GetPerOccurrenceCount with properties constraints.
	count, err := pABC.GetPerUserCount(nil)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), count, "PerUserCount")
	count, err = pABC.GetPerOccurrenceCount(nil)
	assert.Nil(t, err)
	assert.Equal(t, uint(0), count, "Per occurence Count")

	patternConstraints := make([]P.EventConstraints, 3)
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), count, "Per User count")
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(0), count, "PerOccurence count")

	patternConstraints = make([]P.EventConstraints, 3)
	patternConstraints[0] = P.EventConstraints{
		EPNumericConstraints: []P.NumericConstraint{
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
		EPNumericConstraints: []P.NumericConstraint{
			P.NumericConstraint{
				PropertyName: "ComNum",
				LowerBound:   -math.MaxFloat64,
				UpperBound:   2.0,
			},
		},
	}
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count, "pABC.GetPerUserCount")
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	// assert.Equal(t, uint(2), count, "pABC.GetPerOccurrenceCount") // removed as per Occurence counts are not counted

	patternConstraints = make([]P.EventConstraints, 3)
	patternConstraints[0] = P.EventConstraints{
		EPNumericConstraints: []P.NumericConstraint{
			P.NumericConstraint{
				PropertyName: "ANum",
				LowerBound:   0.5,
				UpperBound:   1.5,
			},
			P.NumericConstraint{
				PropertyName: "ComNum",
				LowerBound:   0.5,
				UpperBound:   1.5,
			},
		},
	}
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count, "pABC.GetPerUserCount(patternConstraints)")
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	// assert.Equal(t, uint(2), count) // removed as per Occurence counts are not counted

	patternConstraints = make([]P.EventConstraints, 3)
	// Below categorical combination occurs in the first occurrence.
	patternConstraints[1] = P.EventConstraints{
		EPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "BCat",
				PropertyValue: "bcat1",
			},
		},
	}
	patternConstraints[2] = P.EventConstraints{
		EPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "ComCat",
				PropertyValue: "com2",
			},
		},
	}
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(0), count)

	// User properties constraints.
	patternConstraints = make([]P.EventConstraints, 3)
	patternConstraints[0] = P.EventConstraints{
		UPNumericConstraints: []P.NumericConstraint{
			P.NumericConstraint{
				PropertyName: "age",
				LowerBound:   10.0,
				UpperBound:   25.0,
			},
		},
	}
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	// U1 is age 20.0.
	assert.Equal(t, uint(1), count)
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	// assert.Equal(t, uint(0), count) // removed as per Occurence counts are not counted

	patternConstraints = make([]P.EventConstraints, 3)
	// Below categorical combination occurs in the first occurrence.
	patternConstraints[0] = P.EventConstraints{
		UPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "$country",
				PropertyValue: "India",
			},
		},
	}
	patternConstraints[1] = P.EventConstraints{
		UPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "$paymentStatus",
				PropertyValue: "Free",
			},
		},
	}
	patternConstraints[2] = P.EventConstraints{
		UPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "$paymentStatus",
				PropertyValue: "Paid",
			},
		},
	}
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(0), count) // was 1 previously
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(0), count) //was 1 previously

}

// func TestGetEventNamesfromFile(t *testing.T) {
// 	filepath := "./data/eventname.txt"
// 	scanner, err := T.OpenEventFileAndGetScanner(filepath)
// 	assert.Nil(t, err)
// 	tmpProjectID := uint64(123)
// 	eventNames, err := model.GetEventNamesFromFile(scanner, tmpProjectID)
// 	assert.Equal(t, 2, len(eventNames))
// 	assert.Nil(t, err)

// }

func TestCollectAndCountEventsWithPropertiesWithOccurenceFalse(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count A -> B -> C, Count:3, OncePerUserCount:2, UserCount:2
	// Add False pcountOccur TestCase

	countOccurFlag := false
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
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "F", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1", "IgnoredKey": []string{"check"}}, EventTimestamp: u1ETimestamp, EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2", "IgnoredKey": []string{"check"}}, EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(4),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3", "IgnoredKey": []string{"check"}, "ANum": 1, "ACat": "acat1"}, EventTimestamp: u1ETimestamp + (2 * 60),
			EventCardinality: uint(2), UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1"}, EventTimestamp: u1ETimestamp + (3 * 60), EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2", "BNum": 1, "BCat": "bcat1"}, EventTimestamp: u1ETimestamp + (4 * 60), EventCardinality: uint(5),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3", "ANum": 2, "ACat": "acat2"}, EventTimestamp: u1ETimestamp + (5 * 60), EventCardinality: uint(3),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1", "BNum": 2, "BCat": "bcat2"}, EventTimestamp: u1ETimestamp + (6 * 60), EventCardinality: uint(6),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Free", "age": 20.0},
		},
		P.CounterEventFormat{
			UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2", "CNum": 1.0, "CCat": "ccat1"}, EventTimestamp: u1ETimestamp + (7 * 60), EventCardinality: uint(2),
			UserProperties: map[string]interface{}{"$country": "India", "paymentStatus": "Paid", "age": 20.0},
		},
		// User 2.
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3", "IgnoredKey": []string{"check"}}, EventTimestamp: u2ETimestamp, EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1", "ANum": 1, "ACat": "acat1"}, EventTimestamp: u2ETimestamp + (1 * 60), EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2", "ANum": 2, "ACat": "acat2"}, EventTimestamp: u2ETimestamp + (2 * 60), EventCardinality: uint(2),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3"}, EventTimestamp: u2ETimestamp + (3 * 60), EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0}},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1", "BNum": 1, "BCat": "bcat1"}, EventTimestamp: u2ETimestamp + (4 * 60), EventCardinality: uint(2),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "Z", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2"}, EventTimestamp: u2ETimestamp + (5 * 60), EventCardinality: uint(1),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Free", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3", "CNum": 2.0, "CCat": "ccat2"}, EventTimestamp: u2ETimestamp + (6 * 60), EventCardinality: uint(2),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Paid", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventProperties: map[string]interface{}{"ComNum": 1.0,
				"ComCat": "com1", "ANum": 1.0, "ACat": "acat1"}, EventTimestamp: u2ETimestamp + (7 * 60), EventCardinality: uint(3),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Paid", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventProperties: map[string]interface{}{"ComNum": 2.0,
				"ComCat": "com2", "BNum": 2, "BCat": "bcat2"}, EventTimestamp: u2ETimestamp + (8 * 60), EventCardinality: uint(3),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Paid", "age": 30.0},
		},
		P.CounterEventFormat{
			UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventProperties: map[string]interface{}{"ComNum": 3.0,
				"ComCat": "com3", "CNum": 1.0, "CCat": "ccat1"}, EventTimestamp: u2ETimestamp + (9 * 60), EventCardinality: uint(3),
			UserProperties: map[string]interface{}{"$country": "USA", "paymentStatus": "Paid", "age": 30.0},
		},
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
		UserPropertiesInfo: &P.PropertiesInfo{
			NumericPropertyKeys:          make(map[string]bool),
			CategoricalPropertyKeyValues: make(map[string]map[string]bool),
		},
		EventPropertiesInfoMap: &actualEventInfoMap,
	}
	_, err = P.CollectPropertiesInfo(project.ID, scanner, &userAndEventsInfo)
	assert.Nil(t, err)

	expectedNumericKeys := map[string][]string{
		"A": []string{"ANum", "ComNum"},
		"B": []string{"BNum", "ComNum"},
		"C": []string{"CNum", "ComNum"},
		"F": []string{"ComNum"},
		"Z": []string{"ComNum"},
	}
	expectedCategoricalKeyValues := map[string]map[string][]string{
		"A": map[string][]string{
			"ACat":   []string{"acat1", "acat2"},
			"ComCat": []string{"com1", "com2", "com3"},
		},
		"B": map[string][]string{
			"BCat":   []string{"bcat1", "bcat2"},
			"ComCat": []string{"com1", "com2", "com3"},
		},
		"C": map[string][]string{
			"CCat":   []string{"ccat1", "ccat2"},
			"ComCat": []string{"com1", "com2", "com3"},
		},
		"F": map[string][]string{
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
			fmt.Sprintf("Mismatch numeric keys. event: %s, Expected %v. Actual: %v",
				e, keys, eInfo.NumericPropertyKeys))
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
	pAC, _ := P.NewPattern([]string{"A", "C"}, &userAndEventsInfo)
	pA, _ := P.NewPattern([]string{"A"}, &userAndEventsInfo)
	pB, _ := P.NewPattern([]string{"B"}, &userAndEventsInfo)
	pC, _ := P.NewPattern([]string{"C"}, &userAndEventsInfo)

	patterns := []*P.Pattern{pABC, pAB, pBC, pAC, pA, pB, pC}
	var cProps P.CountAlgoProperties
	cProps.Counting_version = 1

	erronFalse := P.CountPatterns(project.ID, scanner, patterns, countOccurFlag, cProps)
	assert.Nil(t, erronFalse)

	// A-B-C occurs twice PerUser with the following Generic Properties.

	// A: firstSeenOccurrenceCount -> 2 and 1.
	// A: lastSeenOccurrenceCount -> 3 and 3.
	// A: firstSeenTime -> user1CreatedTime+1hour+120seconds and user2CreatedTime+1hour+60seconds.
	// A: lastSeenTime -> user1CreatedTime+1hour+300seconds and user2CreatedTime+1hour+420seconds.
	// A: firstSeenSinceUserJoin -> 1hour+120seconds and 1hour+60seconds.
	// A: lastSeenSinceUserJoin -> 1hour+300seconds and 1hour+420seconds.

	// B: firstSeenOccurrenceCount -> 5 and 2.
	// B: lastSeenOccurrenceCount -> 6 and 3.
	// B: firstSeenTime -> user1CreatedTime+1hour+240seconds and user2CreatedTime+1hour+240seconds.
	// B: lastSeenTime -> user1CreatedTime+1hour+360seconds and user2CreatedTime+1hour+480seconds.
	// B: firstSeenSinceUserJoin -> 1hour+240seconds and 1hour+240seconds.
	// B: lastSeenSinceUserJoin -> 1hour+360seconds and 1hour+480seconds.

	// C: firstSeenOccurrenceCount -> 2 and 2.
	// C: lastSeenOccurrenceCount -> 2 and 3.
	// C: firstSeenTime -> user1CreatedTime+1hour+420seconds and user2CreatedTime+1hour+360seconds.
	// C: lastSeenTime -> user1CreatedTime+1hour+420seconds and user2CreatedTime+1hour+540seconds.
	// C: firstSeenSinceUserJoin -> 1hour+240seconds and 1hour+240seconds.
	// C: lastSeenSinceUserJoin -> 1hour+360seconds and 1hour+540seconds.
	assert.Equal(t, uint(0), pABC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pABC.TotalUserCount)
	assert.Equal(t, uint(2), pABC.PerUserCount)
	assert.Equal(t, pLen, len(pABC.EventNames))
	for i := 0; i < pLen; i++ {
		assert.Equal(t, pABCEvents[i], pABC.EventNames[i])
	}
	assert.Equal(t, uint64(2), pABC.GenericPropertiesHistogram.Count())

	expectedMeanMap := map[string]float64{
		U.UP_JOIN_TIME: float64((u1CTimestamp + u2CTimestamp) / 2.0),
		// Event A Generic Properties.
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 1.0) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((3.0 + 3.0) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 120 + u2CTimestamp + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 300 + u2CTimestamp + 3600 + 420) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 120 + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 300 + 3600 + 420) / 2),

		// Event B Generic Properties.
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((5.0 + 2.0) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((6.0 + 3.0) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 240 + u2CTimestamp + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 360 + u2CTimestamp + 3600 + 480) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 240 + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 360 + 3600 + 480) / 2),

		// Event C Generic Properties.
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 2.0) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((2.0 + 3.0) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 420 + u2CTimestamp + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 420 + u2CTimestamp + 3600 + 540) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 420 + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 420 + 3600 + 540) / 2),
	}
	actualMeanMap := pABC.GenericPropertiesHistogram.MeanMap()

	for k, expectedMean := range expectedMeanMap {
		assert.Equal(t, expectedMean, actualMeanMap[k], fmt.Sprintf("Failed for Key: %s", k))
	}
	// A-B-C occurs twice oncePerUser with the following six dimensional event numerical
	// distribution.
	// 0.ANum: 1.0 and 1.0
	// 0.ComNum: 3.0 and 1.0
	// 1.BNum: 1.0 and 1.0
	// 1.ComNum: 2.0 and 1.0
	// 2.CNum: 1.0 and 2.0
	// 2.ComNum 2.0 and 3.0
	expectedMeanMap = map[string]float64{
		"0.ANum":   float64((1.0 + 1.0) / 2),
		"0.ComNum": float64((3.0 + 1.0) / 2),
		"1.BNum":   float64((1.0 + 1.0) / 2),
		"1.ComNum": float64((2.0 + 1.0) / 2),
		"2.CNum":   float64((1.0 + 2.0) / 2),
		"2.ComNum": float64((2.0 + 3.0) / 2),
	}
	actualMeanMap = pABC.PerUserEventNumericProperties.MeanMap()
	assert.Equal(t, expectedMeanMap, actualMeanMap)

	actualCdf := pABC.PerUserEventNumericProperties.CDFFromMap(
		map[string]float64{
			"0.ANum":   0.0,
			"0.ComNum": 0.0,
			"1.ComNum": 0.0,
		})
	assert.InDelta(t, actualCdf, 0.0, 0.01)
	// assert.True(t, math.IsNaN(actualCdf))
	// 0.ACat: "acat1" and "acat1"
	// 0.ComCat: "com3" and "com1"
	// 1.BCat: "bcat1" and "bcat1"
	// 1.ComCat: "com2" and "com1"
	// 2.CCat: "ccat1" and "ccat2"
	// 2.ComCat: "com2" and "com3"
	actualPdf, err := pABC.PerUserEventCategoricalProperties.PDFFromMap(
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

	// A-B-C occurs twice oncePerUser with the following six dimensional user numerical
	// distribution.
	// 0.age: 20.0 and 30.0
	// 1.age: 20.0 and 30.0
	// 2.age: 20.0 and 30.0
	expectedMeanMap = map[string]float64{
		"0.age": float64((20.0 + 30.0) / 2),
		"1.age": float64((20.0 + 30.0) / 2),
		"2.age": float64((20.0 + 30.0) / 2),
	}
	actualMeanMap = pABC.PerUserUserNumericProperties.MeanMap()
	assert.Equal(t, expectedMeanMap, actualMeanMap)

	actualCdf = pABC.PerUserUserNumericProperties.CDFFromMap(
		map[string]float64{
			"0.age": 25.0,
			"1.age": 25.0,
			"2.age": 25.0,
		})
	assert.InDelta(t, actualCdf, 0.5, 0.01)

	// ABC occurs twice with U1 country India and U2 country USA.
	// Payment status changes from Free to Paid on first occurrence of C.
	// 0.$country: "India" and "USA"
	// 0.paymentStatus: "Free" and "Free"
	// 1.$country: "India" and "USA"
	// 1.paymentStatus: "Free" and "Free"
	// 2.$country: "India" and "USA"
	// 2.paymentStatus: "Paid" and "Paid"
	actualPdf, err = pABC.PerUserUserCategoricalProperties.PDFFromMap(
		map[string]string{
			"0.$country":      "USA",
			"0.paymentStatus": "Free",
			"1.$country":      "USA",
			"1.paymentStatus": "Free",
			"2.$country":      "USA",
			"2.paymentStatus": "Paid",
		})
	assert.Nil(t, err, fmt.Sprintf("Error: %v", err))
	assert.InDelta(t, actualPdf, 0.5, 0.01)

	// A-B-C occurs thrice across users with the following six dimensional event numerical
	// distribution.
	// 0.ANum: 1.0 and 1.0 and 1.0
	// 0.ComNum: 3.0 and 1.0 and 1.0
	// 1.BNum: 1.0 and 1.0 and 2.0
	// 1.ComNum: 2.0 and 1.0 and 2.0
	// 2.CNum: 1.0 and 2.0 and 1.0
	// 2.ComNum 2.0 and 3.0 and 3.0
	expectedMeanMap = map[string]float64{
		"0.ANum":   float64(0),
		"0.ComNum": float64(0),
		"1.BNum":   float64(0),
		"1.ComNum": float64(0),
		"2.CNum":   float64(0),
		"2.ComNum": float64(0),
	}
	fmt.Println(pABC)
	actualMeanMap = pABC.PerOccurrenceEventNumericProperties.MeanMap()

	assert.Equal(t, expectedMeanMap, actualMeanMap)
	actualCdf = pABC.PerOccurrenceEventNumericProperties.CDFFromMap(
		map[string]float64{
			"0.ANum":   2.0,
			"0.ComNum": 2.0,
			"1.ComNum": 1.5,
		})
	assert.True(t, math.IsNaN(actualCdf))

	// 0.ACat: "acat1" and "acat1" and "acat1"
	// 0.ComCat: "com3" and "com1" and "com1"
	// 1.BCat: "bcat1" and "bcat1" and "bcat2"
	// 1.ComCat: "com2" and "com1" and "com2"
	// 2.CCat: "ccat1" and "ccat2" and "ccat1"
	// 2.ComCat: "com2" and "com3" and "com3"

	actualPdf, err = pABC.PerOccurrenceEventCategoricalProperties.PDFFromMap(
		map[string]string{
			"0.ACat":   "acat1",
			"0.ComCat": "com3",
			"1.BCat":   "bcat1",
			"1.ComCat": "com2",
			"2.CCat":   "ccat1",
			"2.ComCat": "com2",
		})
	assert.Nil(t, err, fmt.Sprintf("Error: %v", err))
	assert.InDelta(t, actualPdf, 0.0, 0.01)

	// A-B-C occurs thrice with the following six dimensional user numerical
	// distribution.
	// 0.age: 20.0 and 30.0 and 30.0
	// 1.age: 20.0 and 30.0 and 30.0
	// 2.age: 20.0 and 30.0 and 30.0
	expectedMeanMap = map[string]float64{
		"0.age": float64(0),
		"1.age": float64(0),
		"2.age": float64(0),
	}
	actualMeanMap = pABC.PerOccurrenceUserNumericProperties.MeanMap()
	assert.Equal(t, expectedMeanMap, actualMeanMap)

	actualCdf = pABC.PerOccurrenceUserNumericProperties.CDFFromMap(
		map[string]float64{
			"0.age": 25.0,
			"1.age": 25.0,
			"2.age": 25.0,
		})
	assert.True(t, math.IsNaN(actualCdf))

	// ABC occurs thrice, once with U1 country India and twice with U2 country USA.
	// Payment status changes from Free to Paid on first occurrence of C.
	// 0.$country: "India" and "USA" and "USA"
	// 0.paymentStatus: "Free" and "Free" and "Paid"
	// 1.$country: "India" and "USA" and "USA"
	// 1.paymentStatus: "Free" and "Free" and "Paid"
	// 2.$country: "India" and "USA" and "USA"
	// 2.paymentStatus: "Paid" and "Paid" and "Paid"

	actualPdf, err = pABC.PerOccurrenceUserCategoricalProperties.PDFFromMap(
		map[string]string{
			"0.$country":      "USA",
			"0.paymentStatus": "Free",
			"1.$country":      "USA",
			"1.paymentStatus": "Free",
			"2.$country":      "USA",
			"2.paymentStatus": "Paid",
		})

	assert.Nil(t, err, fmt.Sprintf("Error: %v", err))
	assert.InDelta(t, actualPdf, 0.0, 0.01)

	// Test output on other patterns.
	assert.Equal(t, uint(0), pAB.PerOccurrenceCount)
	assert.Equal(t, uint(2), pAB.PerUserCount)
	assert.Equal(t, uint(2), pAB.TotalUserCount)

	assert.Equal(t, uint(0), pBC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pBC.PerUserCount)
	assert.Equal(t, uint(2), pBC.TotalUserCount)

	assert.Equal(t, uint(0), pAC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pAC.PerUserCount)
	assert.Equal(t, uint(2), pAC.TotalUserCount)

	assert.Equal(t, uint(0), pA.PerOccurrenceCount)
	assert.Equal(t, uint(2), pA.PerUserCount)
	assert.Equal(t, uint(2), pA.TotalUserCount)

	assert.Equal(t, uint(0), pB.PerOccurrenceCount)
	assert.Equal(t, uint(2), pB.PerUserCount)
	assert.Equal(t, uint(2), pB.TotalUserCount)

	assert.Equal(t, uint(0), pC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pC.PerUserCount)
	assert.Equal(t, uint(2), pC.TotalUserCount)

	// Test GetPerUserCount and GetPerOccurrenceCount with properties constraints.
	count, err := pABC.GetPerUserCount(nil)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), count, "PerUserCount")
	count, err = pABC.GetPerOccurrenceCount(nil)
	assert.Nil(t, err)
	assert.Equal(t, uint(0), count, "Per occurence Count")

	patternConstraints := make([]P.EventConstraints, 3)
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), count, "Per User count")
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(0), count, "PerOccurence count")

	patternConstraints = make([]P.EventConstraints, 3)
	patternConstraints[0] = P.EventConstraints{
		EPNumericConstraints: []P.NumericConstraint{
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
		EPNumericConstraints: []P.NumericConstraint{
			P.NumericConstraint{
				PropertyName: "ComNum",
				LowerBound:   -math.MaxFloat64,
				UpperBound:   2.0,
			},
		},
	}
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count, "pABC.GetPerUserCount")
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(INFHEX), count, "pABC.GetPerOccurrenceCount")

	patternConstraints = make([]P.EventConstraints, 3)
	patternConstraints[0] = P.EventConstraints{
		EPNumericConstraints: []P.NumericConstraint{
			P.NumericConstraint{
				PropertyName: "ANum",
				LowerBound:   0.5,
				UpperBound:   1.5,
			},
			P.NumericConstraint{
				PropertyName: "ComNum",
				LowerBound:   0.5,
				UpperBound:   1.5,
			},
		},
	}
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count, "pABC.GetPerUserCount(patternConstraints)")
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(INFHEX), count)

	patternConstraints = make([]P.EventConstraints, 3)
	// Below categorical combination occurs in the first occurrence.
	patternConstraints[1] = P.EventConstraints{
		EPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "BCat",
				PropertyValue: "bcat1",
			},
		},
	}
	patternConstraints[2] = P.EventConstraints{
		EPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "ComCat",
				PropertyValue: "com2",
			},
		},
	}
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(0), count)

	// User properties constraints.
	patternConstraints = make([]P.EventConstraints, 3)
	patternConstraints[0] = P.EventConstraints{
		UPNumericConstraints: []P.NumericConstraint{
			P.NumericConstraint{
				PropertyName: "age",
				LowerBound:   10.0,
				UpperBound:   25.0,
			},
		},
	}
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	// U1 is age 20.0.
	assert.Equal(t, uint(1), count)
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(INFHEX), count)

	patternConstraints = make([]P.EventConstraints, 3)
	// Below categorical combination occurs in the first occurrence.
	patternConstraints[0] = P.EventConstraints{
		UPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "$country",
				PropertyValue: "India",
			},
		},
	}
	patternConstraints[1] = P.EventConstraints{
		UPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "$paymentStatus",
				PropertyValue: "Free",
			},
		},
	}
	patternConstraints[2] = P.EventConstraints{
		UPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "$paymentStatus",
				PropertyValue: "Paid",
			},
		},
	}
	count, err = pABC.GetPerUserCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(0), count) // was 1 previously
	count, err = pABC.GetPerOccurrenceCount(patternConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(0), count)

}

func TestCountPatternsWithOccurenceFalse(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count A -> B -> C, Count:3, OncePerUserCount:2, UserCount:2
	// Add False test Case
	countOccurFlag := false
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
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(4)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventTimestamp: u1ETimestamp + (2 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventTimestamp: u1ETimestamp + (3 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (4 * 60), EventCardinality: uint(5)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventTimestamp: u1ETimestamp + (5 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (6 * 60), EventCardinality: uint(6)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventTimestamp: u1ETimestamp + (7 * 60), EventCardinality: uint(2)},
		// User 2.
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (1 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (2 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventTimestamp: u2ETimestamp + (3 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp + (4 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "Z", EventTimestamp: u2ETimestamp + (5 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventTimestamp: u2ETimestamp + (6 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (7 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp + (8 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventTimestamp: u2ETimestamp + (9 * 60), EventCardinality: uint(3)},
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
	pAC, _ := P.NewPattern([]string{"A", "C"}, nil)
	pA, _ := P.NewPattern([]string{"A"}, nil)
	pB, _ := P.NewPattern([]string{"B"}, nil)
	pC, _ := P.NewPattern([]string{"C"}, nil)

	patterns := []*P.Pattern{pABC, pAB, pBC, pAC, pA, pB, pC}
	var cProps P.CountAlgoProperties
	cProps.Counting_version = 1

	erronFalse := P.CountPatterns(project.ID, scanner, patterns, countOccurFlag, cProps)
	assert.Nil(t, erronFalse)

	// Test pABC output.
	assert.Equal(t, uint(0), pABC.PerOccurrenceCount, "pABC.PerOccurrenceCount")
	assert.Equal(t, uint(2), pABC.TotalUserCount, "pABC.TotalUserCount")
	assert.Equal(t, uint(2), pABC.PerUserCount, "pABC.PerUserCount")
	assert.Equal(t, pLen, len(pABC.EventNames))
	for i := 0; i < pLen; i++ {
		assert.Equal(t, pABCEvents[i], pABC.EventNames[i])
	}

	// A-B-C occurs twice OncePerUser with the following Generic Properties.

	// A: firstSeenOccurrenceCount -> 2 and 1.
	// A: lastSeenOccurrenceCount -> 3 and 3.
	// A: firstSeenTime -> user1CreatedTime+1hour+120seconds and user2CreatedTime+1hour+60seconds.
	// A: lastSeenTime -> user1CreatedTime+1hour+300seconds and user2CreatedTime+1hour+420seconds.
	// A: firstSeenSinceUserJoin -> 1hour+120seconds and 1hour+60seconds.
	// A: lastSeenSinceUserJoin -> 1hour+300seconds and 1hour+420seconds.

	// B: firstSeenOccurrenceCount -> 5 and 2.
	// B: lastSeenOccurrenceCount -> 6 and 3.
	// B: firstSeenTime -> user1CreatedTime+1hour+240seconds and user2CreatedTime+1hour+240seconds.
	// B: lastSeenTime -> user1CreatedTime+1hour+360seconds and user2CreatedTime+1hour+480seconds.
	// B: firstSeenSinceUserJoin -> 1hour+240seconds and 1hour+240seconds.
	// B: lastSeenSinceUserJoin -> 1hour+360seconds and 1hour+480seconds.

	// C: firstSeenOccurrenceCount -> 2 and 2.
	// C: lastSeenOccurrenceCount -> 2 and 3.
	// C: firstSeenTime -> user1CreatedTime+1hour+420seconds and user2CreatedTime+1hour+360seconds.
	// C: lastSeenTime -> user1CreatedTime+1hour+420seconds and user2CreatedTime+1hour+540seconds.
	// C: firstSeenSinceUserJoin -> 1hour+240seconds and 1hour+240seconds.
	// C: lastSeenSinceUserJoin -> 1hour+360seconds and 1hour+540seconds.
	assert.Equal(t, uint64(2), pABC.GenericPropertiesHistogram.Count())
	expectedMeanMap := map[string]float64{
		U.UP_JOIN_TIME: float64((u1CTimestamp + u2CTimestamp) / 2.0),
		// Event A Generic Properties.
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 1.0) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((3.0 + 3.0) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 120 + u2CTimestamp + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 300 + u2CTimestamp + 3600 + 420) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 120 + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 300 + 3600 + 420) / 2),

		// Event B Generic Properties.
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((5.0 + 2.0) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((6.0 + 3.0) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 240 + u2CTimestamp + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 360 + u2CTimestamp + 3600 + 480) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 240 + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 360 + 3600 + 480) / 2),

		// Event C Generic Properties.
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 2.0) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((2.0 + 3.0) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 420 + u2CTimestamp + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_TIME): float64(
			(u1CTimestamp + 3600 + 420 + u2CTimestamp + 3600 + 540) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 420 + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 420 + 3600 + 540) / 2),
	}
	actualMeanMap := pABC.GenericPropertiesHistogram.MeanMap()
	for k, expectedMean := range expectedMeanMap {
		assert.Equal(t, expectedMean, actualMeanMap[k], fmt.Sprintf("Failed for Key: %s", k))
	}

	// Test output on other patterns.
	assert.Equal(t, uint(0), pAB.PerOccurrenceCount)
	assert.Equal(t, uint(2), pAB.PerUserCount)
	assert.Equal(t, uint(2), pAB.TotalUserCount)

	assert.Equal(t, uint(0), pBC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pBC.PerUserCount)
	assert.Equal(t, uint(2), pBC.TotalUserCount)

	assert.Equal(t, uint(0), pAC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pAC.PerUserCount)
	assert.Equal(t, uint(2), pAC.TotalUserCount)

	assert.Equal(t, uint(0), pA.PerOccurrenceCount)
	assert.Equal(t, uint(2), pA.PerUserCount)
	assert.Equal(t, uint(2), pA.TotalUserCount)

	assert.Equal(t, uint(0), pB.PerOccurrenceCount)
	assert.Equal(t, uint(2), pB.PerUserCount)
	assert.Equal(t, uint(2), pB.TotalUserCount)

	assert.Equal(t, uint(0), pC.PerOccurrenceCount)
	assert.Equal(t, uint(2), pC.PerUserCount)
	assert.Equal(t, uint(2), pC.TotalUserCount)

}

func TestGenSegmentsForRepeatedEvents(t *testing.T) {

	pA, _ := P.NewPattern([]string{"A"}, nil)
	pB, _ := P.NewPattern([]string{"B"}, nil)
	pC, _ := P.NewPattern([]string{"C"}, nil)
	pD, _ := P.NewPattern([]string{"D"}, nil)
	pAB, _ := P.NewPattern([]string{"AB", "BC"}, nil)

	currentPatterns := []*P.Pattern{pA, pB, pC, pD}
	repeatedPatterns := []*P.Pattern{pA, pB, pC}
	rPatterns, _, err := P.GenSegmentsForRepeatedEvents(currentPatterns, nil, repeatedPatterns)
	assert.Nil(t, err, "Error is not nil")
	for _, v := range rPatterns {
		assert.Equal(t, v.EventNames[0], v.EventNames[1], "Pattern EventNames are not Equal")
	}

	currentPatterns = []*P.Pattern{}
	repeatedPatterns = []*P.Pattern{pA, pB, pC}
	rPatterns, _, err = P.GenSegmentsForRepeatedEvents(currentPatterns, nil, repeatedPatterns)
	assert.NotNil(t, err, err)

	currentPatterns = []*P.Pattern{pAB}
	repeatedPatterns = []*P.Pattern{pAB}
	rPatterns, _, err = P.GenSegmentsForRepeatedEvents(currentPatterns, nil, repeatedPatterns)
	assert.NotNil(t, err, "length pattern eventNames greater than 1 ")

}

// TODO(aravind): Add tests for genLenThreeSegmentedCandidates and genSegmentedCandidates in run_pattern_mine.go

func TestGenCandidatesForGoals(t *testing.T) {

	pA, _ := P.NewPattern([]string{"A", "C"}, nil)
	pB, _ := P.NewPattern([]string{"B", "C"}, nil)

	patt, err := P.GenCandidatesForGoals(pA, pB, nil)
	assert.NotNil(t, patt, "Patterns cannot be empty")
	assert.Equal(t, 2, len(patt), "number of patterns generated should be 2")
	assert.Nil(t, err)
	res := [][]string{{"A", "B", "C"}, {"B", "A", "C"}}
	var testVar2 [][]string
	for _, v := range patt {
		testVar2 = append(testVar2, v.EventNames)
	}
	assert.ElementsMatch(t, res, testVar2, "Elements are not matching")

	pA, _ = P.NewPattern([]string{"A", "B", "D"}, nil)
	pB, _ = P.NewPattern([]string{"A", "C", "D"}, nil)
	patt, err = P.GenCandidatesForGoals(pA, pB, nil)
	res = [][]string{{"A", "B", "C", "D"}, {"A", "C", "B", "D"}}
	assert.NotNil(t, patt, "Patterns cannot be empty")
	assert.Equal(t, 2, len(patt), "number of patterns generated should be 2")
	assert.Nil(t, err)
	var testVar3 [][]string
	for _, v := range patt {
		testVar3 = append(testVar3, v.EventNames)
	}
	assert.ElementsMatch(t, res, testVar3, "Elements are not matching")

	// test cases that should give empty res for genCandidatesForGoal
	pA, _ = P.NewPattern([]string{"A", "C"}, nil)
	pB, _ = P.NewPattern([]string{"A", "C"}, nil)

	patt, err = P.GenCandidatesForGoals(pA, pB, nil)
	assert.Nil(t, patt, "Patterns should be empty")
	assert.Nil(t, err)

	pA, _ = P.NewPattern([]string{"A", "B", "C"}, nil)
	pB, _ = P.NewPattern([]string{"A", "B", "C"}, nil)

	patt, err = P.GenCandidatesForGoals(pA, pB, nil)
	assert.Nil(t, patt, "Patterns should be empty")
	assert.Nil(t, err)

	pA, _ = P.NewPattern([]string{"A", "D", "C"}, nil)
	pB, _ = P.NewPattern([]string{"B", "E", "C"}, nil)

	patt, err = P.GenCandidatesForGoals(pA, pB, nil)
	assert.Nil(t, patt, "Patterns should be empty")
	assert.Nil(t, err)

	pA, _ = P.NewPattern([]string{"A", "D", "B", "C"}, nil)
	pB, _ = P.NewPattern([]string{"B", "E", "K", "C"}, nil)
	patt, err = P.GenCandidatesForGoals(pA, pB, nil)
	assert.Nil(t, patt, "Patterns should be empty")
	assert.Nil(t, err)

}

func TestCountPatternsWithSameTimeStamp(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	// PerOccurrence cout is set to 0 as we not counting occurence count
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count A -> B -> C, Count:3, OncePerUserCount:2, UserCount:2
	countOccurFlag := true
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
		// P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "F", EventTimestamp: u1ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(4)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (4 * 60), EventCardinality: uint(5)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventTimestamp: u1ETimestamp + (5 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (6 * 60), EventCardinality: uint(6)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventTimestamp: u1ETimestamp + (7 * 60), EventCardinality: uint(2)},
		// // User 2.
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (1 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (1 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventTimestamp: u2ETimestamp + (3 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp + (4 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "Z", EventTimestamp: u2ETimestamp + (5 * 60), EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp + (6 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp + (7 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp + (8 * 60), EventCardinality: uint(3)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventTimestamp: u2ETimestamp + (9 * 60), EventCardinality: uint(3)},
	}
	eventsInputString := ""
	for _, event := range eventsInput {
		lineBytes, _ := json.Marshal(event)
		line := string(lineBytes)
		eventsInputString += fmt.Sprintf("%s\n", line)
	}

	scanner := bufio.NewScanner(strings.NewReader(eventsInputString))

	pABCEvents := []string{"A", "B", "C"}
	pCABEvents := []string{"C", "A", "B"}
	pLen := len(pABCEvents)
	pABC, _ := P.NewPattern(pABCEvents, nil)
	pAB, _ := P.NewPattern([]string{"A", "B"}, nil)
	pCAB, _ := P.NewPattern([]string{"C", "A", "B"}, nil)
	pACB, _ := P.NewPattern([]string{"A", "C", "B"}, nil)
	pBA, _ := P.NewPattern([]string{"B", "A"}, nil)
	pBC, _ := P.NewPattern([]string{"B", "C"}, nil)
	pAC, _ := P.NewPattern([]string{"A", "C"}, nil)
	pA, _ := P.NewPattern([]string{"A"}, nil)
	pB, _ := P.NewPattern([]string{"B"}, nil)
	pC, _ := P.NewPattern([]string{"C"}, nil)

	patterns := []*P.Pattern{pABC, pCAB, pACB, pAB, pBA, pBC, pAC, pA, pB, pC}
	var cProps P.CountAlgoProperties
	cProps.Counting_version = 1
	erronFalse := P.CountPatterns(project.ID, scanner, patterns, countOccurFlag, cProps)
	assert.Nil(t, erronFalse)

	// Test pABC output.
	assert.Equal(t, uint(0), pABC.PerOccurrenceCount, "pABC.PerOccurrenceCount")
	assert.Equal(t, uint(2), pABC.TotalUserCount, "pABC.TotalUserCount")
	assert.Equal(t, uint(2), pABC.PerUserCount, "pABC.PerUserCount")
	assert.Equal(t, pLen, len(pABC.EventNames))
	for i := 0; i < pLen; i++ {
		assert.Equal(t, pABCEvents[i], pABC.EventNames[i])
	}
	// Test pCAB output.
	assert.Equal(t, uint(0), pCAB.PerOccurrenceCount, "pCAB.PerOccurrenceCount")
	assert.Equal(t, uint(2), pCAB.TotalUserCount, "pCAB.TotalUserCount")
	assert.Equal(t, uint(2), pCAB.PerUserCount, "pCAB.PerUserCount")
	assert.Equal(t, pLen, len(pCAB.EventNames))
	for i := 0; i < pLen; i++ {
		assert.Equal(t, pCABEvents[i], pCAB.EventNames[i])
	}

	assert.Equal(t, uint(0), pACB.PerOccurrenceCount, "pACB.PerOccurrenceCount")
	assert.Equal(t, uint(2), pACB.TotalUserCount, "pACB.TotalUserCount")
	assert.Equal(t, uint(2), pACB.PerUserCount, "pACB.PerUserCount")
}

func TestCountPatternsWithSameTimeStamp2(t *testing.T) {
	// all three events with same timestamp -> checking counts of all combinations

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	countOccurFlag := false
	u1CTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	u1ETime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")

	u1CTimestamp := u1CTime.Unix()
	u1ETimestamp := u1ETime.Unix()

	eventsInput := []P.CounterEventFormat{
		// User 1.
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventTimestamp: u1ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventTimestamp: u1ETimestamp, EventCardinality: uint(1)},
	}
	eventsInputString := ""
	for _, event := range eventsInput {
		lineBytes, _ := json.Marshal(event)
		line := string(lineBytes)
		eventsInputString += fmt.Sprintf("%s\n", line)
	}

	scanner := bufio.NewScanner(strings.NewReader(eventsInputString))

	pABC, _ := P.NewPattern([]string{"A", "B", "C"}, nil)
	pAB, _ := P.NewPattern([]string{"A", "B"}, nil)
	pCAB, _ := P.NewPattern([]string{"C", "A", "B"}, nil)
	pCBA, _ := P.NewPattern([]string{"C", "B", "A"}, nil)

	pACB, _ := P.NewPattern([]string{"A", "C", "B"}, nil)
	pBAC, _ := P.NewPattern([]string{"B", "A", "C"}, nil)
	pBCA, _ := P.NewPattern([]string{"B", "C", "A"}, nil)

	pBA, _ := P.NewPattern([]string{"B", "A"}, nil)
	pBC, _ := P.NewPattern([]string{"B", "C"}, nil)
	pAC, _ := P.NewPattern([]string{"A", "C"}, nil)
	pCA, _ := P.NewPattern([]string{"C", "A"}, nil)
	pCB, _ := P.NewPattern([]string{"C", "B"}, nil)

	pA, _ := P.NewPattern([]string{"A"}, nil)
	pB, _ := P.NewPattern([]string{"B"}, nil)
	pC, _ := P.NewPattern([]string{"C"}, nil)

	patterns := []*P.Pattern{pA, pB, pC,
		pAB, pAC, pBA, pBC, pCA, pCB, pABC,
		pACB, pBAC, pBCA, pCAB, pCBA}
	var cProps P.CountAlgoProperties
	cProps.Counting_version = 1
	erronFalse := P.CountPatterns(project.ID, scanner, patterns, countOccurFlag, cProps)
	assert.Nil(t, erronFalse)

	for _, v := range patterns {
		txt := fmt.Sprintf("PerUserCount : %v", v.EventNames)
		assert.Equal(t, uint(1), v.PerUserCount, txt)

	}

}

func TestCountPatternsWithSameTimeStamp3(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	countOccurFlag := false
	u1CTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	u1ETime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")

	u1CTimestamp := u1CTime.Unix()
	u1ETimestamp := u1ETime.Unix()

	eventsInput := []P.CounterEventFormat{
		// User 1.
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventTimestamp: u1ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventTimestamp: u1ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "D", EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(1)},
	}
	eventsInputString := ""
	for _, event := range eventsInput {
		lineBytes, _ := json.Marshal(event)
		line := string(lineBytes)
		eventsInputString += fmt.Sprintf("%s\n", line)
	}

	scanner := bufio.NewScanner(strings.NewReader(eventsInputString))

	pABC, _ := P.NewPattern([]string{"A", "B", "C"}, nil)
	pAB, _ := P.NewPattern([]string{"A", "B"}, nil)
	pCAB, _ := P.NewPattern([]string{"C", "A", "B"}, nil)
	pCBA, _ := P.NewPattern([]string{"C", "B", "A"}, nil)

	pACB, _ := P.NewPattern([]string{"A", "C", "B"}, nil)
	pBAC, _ := P.NewPattern([]string{"B", "A", "C"}, nil)
	pBCA, _ := P.NewPattern([]string{"B", "C", "A"}, nil)

	pBA, _ := P.NewPattern([]string{"B", "A"}, nil)
	pBC, _ := P.NewPattern([]string{"B", "C"}, nil)
	pAC, _ := P.NewPattern([]string{"A", "C"}, nil)
	pCA, _ := P.NewPattern([]string{"C", "A"}, nil)
	pCB, _ := P.NewPattern([]string{"C", "B"}, nil)

	pA, _ := P.NewPattern([]string{"A"}, nil)
	pB, _ := P.NewPattern([]string{"B"}, nil)
	pC, _ := P.NewPattern([]string{"C"}, nil)
	pD, _ := P.NewPattern([]string{"D"}, nil)

	pABD, _ := P.NewPattern([]string{"A", "B", "D"}, nil)
	pACD, _ := P.NewPattern([]string{"A", "C", "D"}, nil)
	pBCD, _ := P.NewPattern([]string{"B", "C", "D"}, nil)

	pAD, _ := P.NewPattern([]string{"A", "D"}, nil)
	pCD, _ := P.NewPattern([]string{"C", "D"}, nil)
	pBD, _ := P.NewPattern([]string{"B", "D"}, nil)

	patterns := []*P.Pattern{pA, pB, pC, pAB, pAC, pBA, pBC, pCA, pCB, pABC, pACB,
		pBAC, pBCA, pCAB, pCBA, pD, pABD, pACD, pBCD, pAD, pBD, pCD}

	var cProps P.CountAlgoProperties
	cProps.Counting_version = 1

	erronFalse := P.CountPatterns(project.ID, scanner, patterns, countOccurFlag, cProps)
	assert.Nil(t, erronFalse)

	for _, v := range patterns {
		txt := fmt.Sprintf("PerUserCount : %v", v.EventNames)
		assert.Equal(t, uint(1), v.PerUserCount, txt)

	}

}

func TestCountPatternsWithSameTimeStamp4(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count A -> B -> C, Count:3, OncePerUserCount:2, UserCount:2
	countOccurFlag := true
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
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "A", EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(4)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "B", EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(2)},
		P.CounterEventFormat{UserId: "U1", UserJoinTimestamp: u1CTimestamp, EventName: "C", EventTimestamp: u1ETimestamp + (1 * 60), EventCardinality: uint(1)},

		// // User 2.
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "A", EventTimestamp: u2ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "B", EventTimestamp: u2ETimestamp, EventCardinality: uint(1)},
		P.CounterEventFormat{UserId: "U2", UserJoinTimestamp: u2CTimestamp, EventName: "C", EventTimestamp: u2ETimestamp, EventCardinality: uint(2)},
	}
	eventsInputString := ""
	for _, event := range eventsInput {
		lineBytes, _ := json.Marshal(event)
		line := string(lineBytes)
		eventsInputString += fmt.Sprintf("%s\n", line)
	}

	scanner := bufio.NewScanner(strings.NewReader(eventsInputString))

	pABC, _ := P.NewPattern([]string{"A", "B", "C"}, nil)
	pAB, _ := P.NewPattern([]string{"A", "B"}, nil)
	pCAB, _ := P.NewPattern([]string{"C", "A", "B"}, nil)
	pCBA, _ := P.NewPattern([]string{"C", "B", "A"}, nil)

	pACB, _ := P.NewPattern([]string{"A", "C", "B"}, nil)
	pBAC, _ := P.NewPattern([]string{"B", "A", "C"}, nil)
	pBCA, _ := P.NewPattern([]string{"B", "C", "A"}, nil)

	pBA, _ := P.NewPattern([]string{"B", "A"}, nil)
	pBC, _ := P.NewPattern([]string{"B", "C"}, nil)
	pAC, _ := P.NewPattern([]string{"A", "C"}, nil)
	pCA, _ := P.NewPattern([]string{"C", "A"}, nil)
	pCB, _ := P.NewPattern([]string{"C", "B"}, nil)

	pA, _ := P.NewPattern([]string{"A"}, nil)
	pB, _ := P.NewPattern([]string{"B"}, nil)
	pC, _ := P.NewPattern([]string{"C"}, nil)

	patterns := []*P.Pattern{pA, pB, pC,
		pAB, pAC, pBA, pBC, pCA, pCB, pABC,
		pACB, pBAC, pBCA, pCAB, pCBA}

	var cProps P.CountAlgoProperties
	cProps.Counting_version = 1
	erronFalse := P.CountPatterns(project.ID, scanner, patterns, countOccurFlag, cProps)
	assert.Nil(t, erronFalse)

	for _, v := range patterns {
		txt := fmt.Sprintf("PerUserCount : %v", v.EventNames)
		assert.Equal(t, uint(2), v.PerUserCount, txt)
		txt = fmt.Sprintf("TotalUserCount : %v", v.EventNames)
		assert.Equal(t, uint(2), v.TotalUserCount, txt)

	}
}

func TestCountPatternWithProperties(t *testing.T) {
	// project, err := SetupProjectReturnDAO()
	projectId := int64(1)
	var cAlgoProps P.CountAlgoProperties
	cAlgoProps.Counting_version = 3
	cAlgoProps.Hmine_support = 0.0001
	cAlgoProps.Hmine_persist = 0
	shouldCountOccurence := false

	p1, _ := P.NewPattern([]string{"$session", "$form_submitted"}, nil)
	// p2, _ := P.NewPattern([]string{"$form_submitted", "$session"}, nil)

	patterns := []*P.Pattern{p1}
	file, err := os.Open("./data/events_test.txt")
	assert.Nil(t, err)
	scanner := bufio.NewScanner(file)

	err = P.CountPatterns(projectId, scanner, patterns, shouldCountOccurence, cAlgoProps)
	assert.Nil(t, err)

	var patsSlice []string = make([]string, 0)
	for _, p := range patterns {
		channel_count := 0
		for _, pats := range p.EventPropertiesPatterns {
			pats_string := strings.Join(pats.Items, "_")
			if strings.Contains(pats_string, "$channel") {
				fmt.Println(pats_string)
				channel_count += 1
			}
			patsSlice = append(patsSlice, pats_string)
		}
		assert.Equal(t, 3, channel_count)

	}

	p := patterns[0]
	eps := p.PerUserCount
	assert.Equal(t, uint(1), eps, fmt.Sprintf("count of %v", p.EventNames))

}

func TestPatternPropertyKeyNumerical(t *testing.T) {

	key := "A"
	val := float64(0.0045)
	tString := P.PatternPropertyKeyValueNumerical(key, val)
	assert.Equal(t, "A::0.0045", tString)
	val = float64(0.00000045)
	tString = P.PatternPropertyKeyValueNumerical(key, val)
	assert.Equal(t, "A::0.00000045", tString)

}

func TestPatternPropertyKeyCategorical(t *testing.T) {

	key := "Country"
	categoricalValue := "US"
	tString := P.PatternPropertyKeyCategorical(key, categoricalValue)
	assert.Equal(t, "Country::US", tString)

}

func TestPropertyCategoricalValue(t *testing.T) {

	propKV := []string{"0.$Country::US", "1.$PageView::0.045"}
	pK, pV := P.PropertySplitKeyValue(propKV[0])
	assert.Equal(t, pK, "0.$Country")
	assert.Equal(t, pV, "US")
	pK, pV = P.PropertySplitKeyValue(propKV[1])
	assert.Equal(t, pK, "1.$PageView")
	assert.Equal(t, pV, "0.045")

}
