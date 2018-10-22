package tests

import (
	"bufio"
	P "pattern"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountPatterns(t *testing.T) {
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count A -> B -> C, Count:3, OncePerUserCount:2, UserCount:2

	const eventsInput = ("U1,2017-06-01T00:00:00Z,F,2017-06-01T01:00:00Z,1\n" +
		"U1,2017-06-01T00:00:00Z,G,2017-06-01T01:01:00Z,2\n" +
		"U1,2017-06-01T00:00:00Z,A,2017-06-01T01:02:00Z,2\n" +
		"U1,2017-06-01T00:00:00Z,L,2017-06-01T01:03:00Z,1\n" +
		"U1,2017-06-01T00:00:00Z,B,2017-06-01T01:04:00Z,5\n" +
		"U1,2017-06-01T00:00:00Z,A,2017-06-01T01:05:00Z,3\n" +
		"U1,2017-06-01T00:00:00Z,B,2017-06-01T01:06:00Z,6\n" +
		"U1,2017-06-01T00:00:00Z,C,2017-06-01T01:07:00Z,1\n" +
		"U2,2017-06-01T00:01:00Z,F,2017-06-01T01:01:00Z,1\n" +
		"U2,2017-06-01T00:01:00Z,A,2017-06-01T01:02:00Z,1\n" +
		"U2,2017-06-01T00:01:00Z,A,2017-06-01T01:03:00Z,2\n" +
		"U2,2017-06-01T00:01:00Z,K,2017-06-01T01:04:00Z,1\n" +
		"U2,2017-06-01T00:01:00Z,B,2017-06-01T01:05:00Z,1\n" +
		"U2,2017-06-01T00:01:00Z,Z,2017-06-01T01:06:00Z,1\n" +
		"U2,2017-06-01T00:01:00Z,C,2017-06-01T01:07:00Z,1\n" +
		"U2,2017-06-01T00:01:00Z,A,2017-06-01T01:08:00Z,3\n" +
		"U2,2017-06-01T00:01:00Z,B,2017-06-01T01:09:00Z,2\n" +
		"U2,2017-06-01T00:01:00Z,C,2017-06-01T01:10:00Z,2\n")
	scanner := bufio.NewScanner(strings.NewReader(eventsInput))

	pABCEvents := []string{"A", "B", "C"}
	pLen := len(pABCEvents)
	pABC, _ := P.NewPattern(pABCEvents)
	pAB, _ := P.NewPattern([]string{"A", "B"})
	pBC, _ := P.NewPattern([]string{"B", "C"})
	pAC, _ := P.NewPattern([]string{"B", "C"})
	pA, _ := P.NewPattern([]string{"A"})
	pB, _ := P.NewPattern([]string{"B"})
	pC, _ := P.NewPattern([]string{"C"})

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
	p1, _ := P.NewPattern([]string{"A"})
	p2, _ := P.NewPattern([]string{"A", "B"})
	c1, c2, ok := P.GenCandidatesPair(p1, p2)
	assert.Nil(t, c1)
	assert.Nil(t, c2)
	assert.Equal(t, false, ok)

	// More than one different element.
	p1, _ = P.NewPattern([]string{"A", "B", "C"})
	p2, _ = P.NewPattern([]string{"A", "D", "E"})
	c1, c2, ok = P.GenCandidatesPair(p1, p2)
	assert.Nil(t, c1)
	assert.Nil(t, c2)
	assert.Equal(t, false, ok)

	// Single element candidates.
	p1, _ = P.NewPattern([]string{"A"})
	p2, _ = P.NewPattern([]string{"B"})
	c1, c2, ok = P.GenCandidatesPair(p1, p2)
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{"B", "A"}, c1.EventNames)
	assert.Equal(t, []string{"A", "B"}, c2.EventNames)

	// Different at the begining.
	p1, _ = P.NewPattern([]string{"B", "C"})
	p2, _ = P.NewPattern([]string{"A", "C"})
	c1, c2, ok = P.GenCandidatesPair(p1, p2)
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{"A", "B", "C"}, c1.EventNames)
	assert.Equal(t, []string{"B", "A", "C"}, c2.EventNames)

	// Different at the end.
	p1, _ = P.NewPattern([]string{"A", "B", "D"})
	p2, _ = P.NewPattern([]string{"A", "B", "C"})
	c1, c2, ok = P.GenCandidatesPair(p1, p2)
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{"A", "B", "C", "D"}, c1.EventNames)
	assert.Equal(t, []string{"A", "B", "D", "C"}, c2.EventNames)

	// Different in the middle.
	p1, _ = P.NewPattern([]string{"A", "C", "D"})
	p2, _ = P.NewPattern([]string{"A", "B", "D"})
	c1, c2, ok = P.GenCandidatesPair(p1, p2)
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{"A", "B", "C", "D"}, c1.EventNames)
	assert.Equal(t, []string{"A", "C", "B", "D"}, c2.EventNames)
}

func TestGenLenThreeCandidatePatterns(t *testing.T) {
	// Not of length 2.
	pattern, _ := P.NewPattern([]string{"A", "X", "Z"})
	startPatterns := []*P.Pattern{}
	endPatterns := []*P.Pattern{}
	maxCandidates := 5
	cPatterns, err := P.GenLenThreeCandidatePatterns(
		pattern, startPatterns, endPatterns, maxCandidates)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)

	// Mismatch event.
	pattern, _ = P.NewPattern([]string{"A", "Z"})
	mismatchPattern, _ := P.NewPattern([]string{"B", "X"})
	patterns1 := []*P.Pattern{mismatchPattern}
	patterns2 := []*P.Pattern{}
	maxCandidates = 5
	// Mismatch start event.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns1, patterns2, maxCandidates)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)
	// Mismatch end event.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns2, patterns1, maxCandidates)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)

	// Mismatch length.
	pattern, _ = P.NewPattern([]string{"A", "Z"})
	mismatchPattern, _ = P.NewPattern([]string{"A", "B", "Z"})
	patterns1 = []*P.Pattern{mismatchPattern}
	patterns2 = []*P.Pattern{}
	maxCandidates = 5
	// Mismatch in startPatterns.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns1, patterns2, maxCandidates)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)
	// Mismatch in endPatterns.
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, patterns2, patterns1, maxCandidates)
	assert.NotNil(t, err)
	assert.Nil(t, cPatterns)

	// Candidate generation.
	pattern, _ = P.NewPattern([]string{"A", "Z"})
	maxCandidates = 3
	sp1, _ := P.NewPattern([]string{"A", "B"}) // Skipped. BZ not found.
	sp2, _ := P.NewPattern([]string{"A", "Z"}) // Skipped. Same as pattern.
	sp3, _ := P.NewPattern([]string{"A", "C"}) // Skipped. ACZ Repeat.
	sp4, _ := P.NewPattern([]string{"A", "D"}) // Skipped. ADZ Repeat.
	sp5, _ := P.NewPattern([]string{"A", "E"}) // Skipped. AEZ Repeat.
	sp6, _ := P.NewPattern([]string{"A", "F"}) // Ignored. Greater than maxCandidates.
	startPatterns = []*P.Pattern{sp1, sp2, sp3, sp4, sp5, sp6}
	ep1, _ := P.NewPattern([]string{"C", "Z"}) // cPatterns[0] ACZ
	ep2, _ := P.NewPattern([]string{"D", "Z"}) // cPatterns[1] ADZ
	ep3, _ := P.NewPattern([]string{"E", "Z"}) // cPatterns[2] AEZ
	ep4, _ := P.NewPattern([]string{"F", "Z"}) // Ignored. Greater than maxCandidates.
	endPatterns = []*P.Pattern{ep1, ep2, ep3, ep4}
	cPatterns, err = P.GenLenThreeCandidatePatterns(
		pattern, startPatterns, endPatterns, maxCandidates)
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

// TODO: Add tests for genLenThreeSegmentedCandidates and genSegmentedCandidates in run_pattern_mine.go
