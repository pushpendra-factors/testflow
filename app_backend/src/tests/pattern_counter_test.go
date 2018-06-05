package tests

import (
	P "pattern"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

// TODO: Add tests for GenCandidates.
