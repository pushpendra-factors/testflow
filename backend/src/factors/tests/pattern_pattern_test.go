package tests

import (
	P "factors/pattern"
	U "factors/util"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPatternCountEvents(t *testing.T) {
	// Count A -> B -> C
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count:3, OncePerUserCount:2, UserCount:2
	pEvents := []string{"A", "B", "C"}
	p, err := P.NewPattern(pEvents, nil)
	assert.Nil(t, err)
	assert.NotNil(t, p)
	pLen := len(pEvents)
	assert.Equal(t, pLen, len(p.EventNames))
	assert.Equal(t, uint(0), p.Count)
	assert.Equal(t, uint(0), p.UserCount)
	assert.Equal(t, uint(0), p.OncePerUserCount)
	// User 1 events.
	userId := "user1"
	userCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	err = p.ResetForNewUser(userId, userCreatedTime)
	assert.Nil(t, err)
	nextEventCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	events := []string{"F", "G", "A", "L", "B", "A", "B", "C"}
	cardinalities := []uint{1, 2, 2, 1, 5, 3, 6, 1}
	for i, event := range events {
		_, err = p.CountForEvent(event, nextEventCreatedTime,
			make(map[string]interface{}), cardinalities[i], userId, userCreatedTime)
		assert.Nil(t, err)
		nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}
	// User 2 events.
	userId = "user2"
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:01:00Z")
	err = p.ResetForNewUser(userId, userCreatedTime)
	assert.Nil(t, err)
	nextEventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:01:00Z")
	events = []string{"F", "A", "A", "K", "B", "Z", "C", "A", "B", "C"}
	cardinalities = []uint{1, 1, 2, 1, 1, 1, 1, 3, 2, 2}
	for i, event := range events {
		_, err = p.CountForEvent(event, nextEventCreatedTime,
			make(map[string]interface{}), cardinalities[i], userId, userCreatedTime)
		assert.Nil(t, err)
		nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}

	assert.Equal(t, uint(3), p.Count)
	assert.Equal(t, uint(2), p.UserCount)
	assert.Equal(t, uint(2), p.OncePerUserCount)
	assert.Equal(t, pLen, len(p.EventNames))
	for i := 0; i < pLen; i++ {
		assert.Equal(t, pEvents[i], p.EventNames[i])
	}
	// A-B-C occurs twice oncePerUser , with first A occurring after 3720s in User1 and
	// 3660s in User 2.
	// Repeats once before the  next B occurs in User2.
	assert.Equal(t, uint64(2), p.CardinalityRepeatTimings.Count())
	assert.Equal(t, float64((2.0+1.0)/2), p.CardinalityRepeatTimings.Mean()[0])
	assert.Equal(t, float64((1.0+2.0)/2), p.CardinalityRepeatTimings.Mean()[1])
	assert.Equal(t, float64((3720.0+3660.0)/2), p.CardinalityRepeatTimings.Mean()[2])

	// A-B-C occurs twice oncePerUser, with first B following first A after 120s in User1 and
	// 180 in User 2.
	// Repeats once before the  next C occurs in User1.
	/* Only start and end event are tracked currently.
	assert.Equal(t, float64(2), p.Timings[1].Count())
	assert.Equal(t, float64((120.0+180.0)/2), p.Timings[1].Mean())
	assert.Equal(t, float64((5.0+1.0)/2), p.EventCardinalities[1].Mean())
	assert.Equal(t, float64((2.0+1.0)/2), p.Repeats[1].Mean())
	*/

	// A-B-C occurs twice oncePerUser, with first C following first B after 180s in User1 and
	// 120s in User 2.
	// Last event always is counted once.
	assert.Equal(t, float64((1.0+1.0)/2), p.CardinalityRepeatTimings.Mean()[3])
	assert.Equal(t, float64((1.0+1.0)/2), p.CardinalityRepeatTimings.Mean()[4])
	assert.Equal(t, float64((180.0+120.0)/2), p.CardinalityRepeatTimings.Mean()[5])
}

func TestPatternGetOncePerUserCount(t *testing.T) {
	// Count A -> B -> C
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count:3, OncePerUserCount:2, UserCount:2
	pEvents := []string{"A", "B", "C"}
	p, err := P.NewPattern(pEvents, nil)
	// User 1 events.
	userId := "user1"
	userCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	err = p.ResetForNewUser(userId, userCreatedTime)
	assert.Nil(t, err)
	nextEventCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	events := []string{"F", "G", "A", "L", "B", "A", "B", "C"}
	cardinalities := []uint{1, 2, 2, 1, 5, 3, 6, 1}
	for i, event := range events {
		_, err = p.CountForEvent(event, nextEventCreatedTime,
			make(map[string]interface{}), cardinalities[i], userId, userCreatedTime)
		assert.Nil(t, err)
		nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}
	// User 2 events.
	userId = "user2"
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:01:00Z")
	err = p.ResetForNewUser(userId, userCreatedTime)
	assert.Nil(t, err)
	nextEventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:01:00Z")
	events = []string{"F", "A", "A", "K", "B", "Z", "C", "A", "B", "C"}
	cardinalities = []uint{1, 1, 2, 1, 1, 1, 2, 3, 2, 3}
	for i, event := range events {
		_, err = p.CountForEvent(event, nextEventCreatedTime,
			make(map[string]interface{}), cardinalities[i], userId, userCreatedTime)
		assert.Nil(t, err)
		nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}

	// A-B-C occurs 2 times, with cardinality of C 1, 2.
	assert.Equal(t, uint(3), p.Count)
	assert.Equal(t, uint(2), p.UserCount)
	assert.Equal(t, uint(2), p.OncePerUserCount)
	assert.Equal(t, float64((1.0+2.0)/2), p.CardinalityRepeatTimings.Mean()[3])
	count, err := p.GetOncePerUserCount(nil)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), count)
	lastEventCardinalityConstraints := []P.EventConstraints{
		P.EventConstraints{},
		P.EventConstraints{},
		P.EventConstraints{},
	}
	count, err = p.GetOncePerUserCount(lastEventCardinalityConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), count)
	lastEventCardinalityConstraints[2].NumericConstraints = []P.NumericConstraint{
		P.NumericConstraint{
			PropertyName: U.EP_OCCURRENCE_COUNT,
			LowerBound:   0.5,
			UpperBound:   math.MaxFloat64,
		},
	}
	count, err = p.GetOncePerUserCount(lastEventCardinalityConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), count)
	lastEventCardinalityConstraints[2].NumericConstraints = []P.NumericConstraint{
		P.NumericConstraint{
			PropertyName: U.EP_OCCURRENCE_COUNT,
			LowerBound:   -math.MaxFloat64,
			UpperBound:   1.5,
		},
	}
	count, err = p.GetOncePerUserCount(lastEventCardinalityConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)
	lastEventCardinalityConstraints[2].NumericConstraints = []P.NumericConstraint{
		P.NumericConstraint{
			PropertyName: U.EP_OCCURRENCE_COUNT,
			LowerBound:   0.5,
			UpperBound:   1.5,
		},
	}
	count, err = p.GetOncePerUserCount(lastEventCardinalityConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)
	lastEventCardinalityConstraints[2].NumericConstraints = []P.NumericConstraint{
		P.NumericConstraint{
			PropertyName: U.EP_OCCURRENCE_COUNT,
			LowerBound:   1.5,
			UpperBound:   2.5,
		},
	}
	count, err = p.GetOncePerUserCount(lastEventCardinalityConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)
	lastEventCardinalityConstraints[2].NumericConstraints = []P.NumericConstraint{
		P.NumericConstraint{
			PropertyName: U.EP_OCCURRENCE_COUNT,
			LowerBound:   1.5,
			UpperBound:   3.5,
		},
	}
	count, err = p.GetOncePerUserCount(lastEventCardinalityConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)

	// A-B-C occurrs twice with cardinality of A 2, 1.
	startEventCardinalityConstraints := []P.EventConstraints{
		P.EventConstraints{},
		P.EventConstraints{},
		P.EventConstraints{},
	}
	startEventCardinalityConstraints[0].NumericConstraints = []P.NumericConstraint{
		P.NumericConstraint{
			PropertyName: U.EP_OCCURRENCE_COUNT,
			LowerBound:   0.5,
			UpperBound:   1.5,
		},
	}
	count, err = p.GetOncePerUserCount(startEventCardinalityConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)
	startEventCardinalityConstraints[0].NumericConstraints = []P.NumericConstraint{
		P.NumericConstraint{
			PropertyName: U.EP_OCCURRENCE_COUNT,
			LowerBound:   -math.MaxFloat64,
			UpperBound:   0.5,
		},
	}
	count, err = p.GetOncePerUserCount(startEventCardinalityConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(0), count)
	startEventCardinalityConstraints[0].NumericConstraints = []P.NumericConstraint{
		P.NumericConstraint{
			PropertyName: U.EP_OCCURRENCE_COUNT,
			LowerBound:   1.5,
			UpperBound:   math.MaxFloat64,
		},
	}
	count, err = p.GetOncePerUserCount(startEventCardinalityConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)
}

func TestPatternEdgeConditions(t *testing.T) {
	// Test NewPattern with empty array.
	p, err := P.NewPattern([]string{}, nil)
	assert.NotNil(t, err)
	assert.Nil(t, p)

	// Test NewPattern with repeated elements.
	p, err = P.NewPattern([]string{"A", "B", "A", "C"}, nil)
	assert.NotNil(t, err)
	assert.Nil(t, p)

	// Test Empty Pattern Creation.
	p, err = P.NewPattern([]string{"A", "B", "C"}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, uint(0), p.Count)
	assert.Equal(t, uint(0), p.UserCount)
	assert.Equal(t, uint(0), p.OncePerUserCount)

	// Test ResetForNewUser without time or Id.
	p, err = P.NewPattern([]string{"A", "B", "C"}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, p)
	userCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	err = p.ResetForNewUser("", userCreatedTime)
	assert.NotNil(t, err)
	err = p.ResetForNewUser("user1", time.Time{})
	assert.NotNil(t, err)

	// Test Count Event with User not initialized.
	p, err = P.NewPattern([]string{"A", "B", "C"}, nil)
	assert.Nil(t, err)
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	eventCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	_, err = p.CountForEvent("J", eventCreatedTime,
		make(map[string]interface{}), 1, "user1", userCreatedTime)
	assert.NotNil(t, err)

	// Test Count Event, with wrong userId or wrong userCreatedTime.
	p, err = P.NewPattern([]string{"A", "B", "C"}, nil)
	assert.Nil(t, err)
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	eventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	err = p.ResetForNewUser("user1", userCreatedTime)
	assert.Nil(t, err)
	_, err = p.CountForEvent("J", eventCreatedTime, make(map[string]interface{}),
		1, "user1", userCreatedTime)
	assert.Nil(t, err)
	// Wrong userId.
	_, err = p.CountForEvent("J", eventCreatedTime,
		make(map[string]interface{}), 1, "user2", userCreatedTime)
	assert.NotNil(t, err)
	// Wrong userCreatedTime
	_, err = p.CountForEvent("J", eventCreatedTime,
		make(map[string]interface{}), 1, "user1", eventCreatedTime)
	assert.NotNil(t, err)

	// Test Events out of order. Out of order events are noticed only when
	// the whole pattern is observed.
	p, err = P.NewPattern([]string{"A", "B"}, nil)
	assert.Nil(t, err)
	// Event1 and userCreated are out of order.
	// Ignoring this error for now, since there are no DB checks to avoid
	// these user input values.
	/*userId := "user1"
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	event1CreatedTime, _ := time.Parse(time.RFC3339, "2017-05-30T01:00:00Z")
	event2CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	err = p.ResetForNewUser(userId, userCreatedTime)
	assert.Nil(t, err)
	_, err = p.CountForEvent("A", event1CreatedTime,
	 make(map[string]interface{}), 1, userId, userCreatedTime)
	assert.Nil(t, err)
	_, err = p.CountForEvent("B", event2CreatedTime,
	make(map[string]interface{}), 1, userId, userCreatedTime)
	assert.NotNil(t, err)*/

	// Event2 and Event1 are out of order.
	userId := "user2"
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	event1CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	event2CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:59:59Z")
	err = p.ResetForNewUser(userId, userCreatedTime)
	assert.Nil(t, err)
	_, err = p.CountForEvent("A", event1CreatedTime,
		make(map[string]interface{}), 1, userId, userCreatedTime)
	assert.Nil(t, err)
	_, err = p.CountForEvent("B", event2CreatedTime,
		make(map[string]interface{}), 1, userId, userCreatedTime)
	assert.NotNil(t, err)
}
