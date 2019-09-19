package tests

import (
	P "factors/pattern"
	U "factors/util"
	"fmt"
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
	assert.Equal(t, uint(0), p.PerOccurrenceCount)
	assert.Equal(t, uint(0), p.TotalUserCount)
	assert.Equal(t, uint(0), p.PerUserCount)
	// User 1 events.
	userId := "user1"
	user1CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	err = p.ResetForNewUser(userId, user1CreatedTime.Unix())
	assert.Nil(t, err)
	nextEventCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	events := []string{"F", "B", "A", "C", "B", "A", "B", "C"}
	cardinalities := []uint{1, 4, 2, 1, 5, 3, 6, 2}
	for i, event := range events {
		err = p.CountForEvent(event, nextEventCreatedTime.Unix(), make(map[string]interface{}),
			make(map[string]interface{}), cardinalities[i], userId, user1CreatedTime.Unix())
		assert.Nil(t, err)
		nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}
	// User 2 events.
	userId = "user2"
	user2CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:01:00Z")
	err = p.ResetForNewUser(userId, user2CreatedTime.Unix())
	assert.Nil(t, err)
	nextEventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:01:00Z")
	events = []string{"B", "A", "A", "C", "B", "Z", "C", "A", "B", "C"}
	cardinalities = []uint{1, 1, 2, 1, 2, 1, 2, 3, 3, 3}
	for i, event := range events {
		err = p.CountForEvent(event, nextEventCreatedTime.Unix(), make(map[string]interface{}),
			make(map[string]interface{}), cardinalities[i], userId, user2CreatedTime.Unix())
		assert.Nil(t, err)
		nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}

	err = p.ResetAfterLastUser()
	assert.Nil(t, err)

	assert.Equal(t, uint(3), p.PerOccurrenceCount)
	assert.Equal(t, uint(2), p.TotalUserCount)
	assert.Equal(t, uint(2), p.PerUserCount)
	assert.Equal(t, pLen, len(p.EventNames))
	for i := 0; i < pLen; i++ {
		assert.Equal(t, pEvents[i], p.EventNames[i])
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
	assert.Equal(t, uint64(2), p.GenericPropertiesHistogram.Count())
	expectedMeanMap := map[string]float64{
		U.UP_JOIN_TIME: float64((user1CreatedTime.Unix() + user2CreatedTime.Unix()) / 2.0),
		// Event A Generic Properties.
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 1.0) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((3.0 + 3.0) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 120 + user2CreatedTime.Unix() + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 300 + user2CreatedTime.Unix() + 3600 + 420) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 120 + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 300 + 3600 + 420) / 2),

		// Event B Generic Properties.
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((5.0 + 2.0) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((6.0 + 3.0) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 240 + user2CreatedTime.Unix() + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 360 + user2CreatedTime.Unix() + 3600 + 480) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 240 + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 360 + 3600 + 480) / 2),

		// Event C Generic Properties.
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 2.0) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((2.0 + 3.0) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 420 + user2CreatedTime.Unix() + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 420 + user2CreatedTime.Unix() + 3600 + 540) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 420 + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 420 + 3600 + 540) / 2),
	}
	actualMeanMap := p.GenericPropertiesHistogram.MeanMap()
	for k, expectedMean := range expectedMeanMap {
		assert.Equal(t, expectedMean, actualMeanMap[k], fmt.Sprintf("Failed for Key: %s", k))
	}
}

func TestPatternGetPerUserCount(t *testing.T) {
	// Count A -> B -> C
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count:3, OncePerUserCount:2, UserCount:2
	pEvents := []string{"A", "B", "C"}
	p, err := P.NewPattern(pEvents, nil)
	// User 1 events.
	userId := "user1"
	user1CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	err = p.ResetForNewUser(userId, user1CreatedTime.Unix())
	assert.Nil(t, err)
	nextEventCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	events := []string{"F", "B", "A", "C", "B", "A", "B", "C"}
	cardinalities := []uint{1, 4, 2, 1, 5, 3, 6, 2}
	for i, event := range events {
		err = p.CountForEvent(event, nextEventCreatedTime.Unix(), make(map[string]interface{}),
			make(map[string]interface{}), cardinalities[i], userId, user1CreatedTime.Unix())
		assert.Nil(t, err)
		nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}
	// User 2 events.
	userId = "user2"
	user2CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:01:00Z")
	err = p.ResetForNewUser(userId, user2CreatedTime.Unix())
	assert.Nil(t, err)
	nextEventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:01:00Z")
	events = []string{"B", "A", "A", "C", "B", "Z", "C", "A", "B", "C"}
	cardinalities = []uint{1, 1, 2, 1, 2, 1, 2, 3, 3, 3}
	for i, event := range events {
		err = p.CountForEvent(event, nextEventCreatedTime.Unix(), make(map[string]interface{}),
			make(map[string]interface{}), cardinalities[i], userId, user2CreatedTime.Unix())
		assert.Nil(t, err)
		nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}

	err = p.ResetAfterLastUser()
	assert.Nil(t, err)

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

	assert.Equal(t, uint(3), p.PerOccurrenceCount)
	assert.Equal(t, uint(2), p.TotalUserCount)
	assert.Equal(t, uint(2), p.PerUserCount)

	assert.Equal(t, uint64(2), p.GenericPropertiesHistogram.Count())
	expectedMeanMap := map[string]float64{
		U.UP_JOIN_TIME: float64((user1CreatedTime.Unix() + user2CreatedTime.Unix()) / 2.0),
		// Event A Generic Properties.
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 1.0) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((3.0 + 3.0) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 120 + user2CreatedTime.Unix() + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 300 + user2CreatedTime.Unix() + 3600 + 420) / 2),
		P.PatternPropertyKey(0, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 120 + 3600 + 60) / 2),
		P.PatternPropertyKey(0, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 300 + 3600 + 420) / 2),

		// Event B Generic Properties.
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((5.0 + 2.0) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((6.0 + 3.0) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 240 + user2CreatedTime.Unix() + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 360 + user2CreatedTime.Unix() + 3600 + 480) / 2),
		P.PatternPropertyKey(1, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 240 + 3600 + 240) / 2),
		P.PatternPropertyKey(1, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 360 + 3600 + 480) / 2),

		// Event C Generic Properties.
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_OCCURRENCE_COUNT): float64((2.0 + 2.0) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_OCCURRENCE_COUNT):  float64((2.0 + 3.0) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 420 + user2CreatedTime.Unix() + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_TIME): float64(
			(user1CreatedTime.Unix() + 3600 + 420 + user2CreatedTime.Unix() + 3600 + 540) / 2),
		P.PatternPropertyKey(2, U.EP_FIRST_SEEN_SINCE_USER_JOIN): float64((3600 + 420 + 3600 + 360) / 2),
		P.PatternPropertyKey(2, U.EP_LAST_SEEN_SINCE_USER_JOIN):  float64((3600 + 420 + 3600 + 540) / 2),
	}
	actualMeanMap := p.GenericPropertiesHistogram.MeanMap()
	for k, expectedMean := range expectedMeanMap {
		assert.Equal(t, expectedMean, actualMeanMap[k], fmt.Sprintf("Failed for Key: %s", k))
	}

	count, err := p.GetPerUserCount(nil)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), count)
	genericPropertiesConstraints := []P.EventConstraints{
		P.EventConstraints{},
		P.EventConstraints{},
		P.EventConstraints{},
	}
	count, err = p.GetPerUserCount(genericPropertiesConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), count)

	genericPropertiesConstraints = []P.EventConstraints{
		P.EventConstraints{},
		P.EventConstraints{},
		P.EventConstraints{
			UPNumericConstraints: []P.NumericConstraint{
				P.NumericConstraint{
					PropertyName: U.UP_JOIN_TIME,
					LowerBound:   -math.MaxFloat64,
					UpperBound:   float64((user1CreatedTime.Unix() + user2CreatedTime.Unix()) / 2.0),
				},
			},
		},
	}
	count, err = p.GetPerUserCount(genericPropertiesConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)

	genericPropertiesConstraints = []P.EventConstraints{
		P.EventConstraints{
			EPNumericConstraints: []P.NumericConstraint{
				P.NumericConstraint{
					PropertyName: U.EP_FIRST_SEEN_OCCURRENCE_COUNT,
					LowerBound:   1.5,
					UpperBound:   math.MaxFloat64,
				},
			},
		},
		P.EventConstraints{},
		P.EventConstraints{},
	}
	count, err = p.GetPerUserCount(genericPropertiesConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)

	genericPropertiesConstraints = []P.EventConstraints{
		P.EventConstraints{},
		P.EventConstraints{
			EPNumericConstraints: []P.NumericConstraint{
				P.NumericConstraint{
					PropertyName: U.EP_LAST_SEEN_OCCURRENCE_COUNT,
					LowerBound:   4.0,
					UpperBound:   7.0,
				},
			},
		},
		P.EventConstraints{},
	}
	count, err = p.GetPerUserCount(genericPropertiesConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)

	genericPropertiesConstraints = []P.EventConstraints{
		P.EventConstraints{},
		P.EventConstraints{
			EPNumericConstraints: []P.NumericConstraint{
				P.NumericConstraint{
					PropertyName: U.EP_FIRST_SEEN_TIME,
					LowerBound:   float64(user1CreatedTime.Unix() + 3600 + 230),
					UpperBound:   float64(user1CreatedTime.Unix() + 3600 + 250),
				},
			},
		},
		P.EventConstraints{},
	}
	count, err = p.GetPerUserCount(genericPropertiesConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)

	genericPropertiesConstraints = []P.EventConstraints{
		P.EventConstraints{},
		P.EventConstraints{
			EPNumericConstraints: []P.NumericConstraint{
				P.NumericConstraint{
					PropertyName: U.EP_LAST_SEEN_TIME,
					LowerBound:   float64(user1CreatedTime.Unix() + 3600 + 350),
					UpperBound:   float64(user1CreatedTime.Unix() + 3600 + 370),
				},
			},
		},
		P.EventConstraints{},
	}
	count, err = p.GetPerUserCount(genericPropertiesConstraints)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), count)

	genericPropertiesConstraints = []P.EventConstraints{
		P.EventConstraints{
			EPNumericConstraints: []P.NumericConstraint{
				P.NumericConstraint{
					PropertyName: U.EP_FIRST_SEEN_SINCE_USER_JOIN,
					LowerBound:   float64(3600 + 110),
					UpperBound:   float64(3600 + 130),
				},
			},
		},
		P.EventConstraints{},
		P.EventConstraints{
			EPNumericConstraints: []P.NumericConstraint{
				P.NumericConstraint{
					PropertyName: U.EP_LAST_SEEN_SINCE_USER_JOIN,
					LowerBound:   float64(3600 + 410),
					UpperBound:   float64(3600 + 430),
				},
			},
		},
	}
	count, err = p.GetPerUserCount(genericPropertiesConstraints)
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
	assert.Equal(t, uint(0), p.PerOccurrenceCount)
	assert.Equal(t, uint(0), p.TotalUserCount)
	assert.Equal(t, uint(0), p.PerUserCount)

	// Test ResetForNewUser without time or Id.
	p, err = P.NewPattern([]string{"A", "B", "C"}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, p)
	userCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	err = p.ResetForNewUser("", userCreatedTime.Unix())
	assert.NotNil(t, err)
	err = p.ResetForNewUser("user1", 0)
	assert.NotNil(t, err)

	// Test Count Event with User not initialized.
	p, err = P.NewPattern([]string{"A", "B", "C"}, nil)
	assert.Nil(t, err)
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	eventCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	err = p.CountForEvent("J", eventCreatedTime.Unix(), make(map[string]interface{}),
		make(map[string]interface{}), 1, "user1", userCreatedTime.Unix())
	assert.NotNil(t, err)

	// Test Count Event, with wrong userId or wrong userCreatedTime.
	p, err = P.NewPattern([]string{"A", "B", "C"}, nil)
	assert.Nil(t, err)
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	eventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	err = p.ResetForNewUser("user1", userCreatedTime.Unix())
	assert.Nil(t, err)
	err = p.CountForEvent("J", eventCreatedTime.Unix(), make(map[string]interface{}),
		make(map[string]interface{}), 1, "user1", userCreatedTime.Unix())
	assert.Nil(t, err)
	// Wrong userId.
	err = p.CountForEvent("J", eventCreatedTime.Unix(), make(map[string]interface{}),
		make(map[string]interface{}), 1, "user2", userCreatedTime.Unix())
	assert.NotNil(t, err)
	// Wrong userCreatedTime. Error is ignored.
	err = p.CountForEvent("J", eventCreatedTime.Unix(), make(map[string]interface{}),
		make(map[string]interface{}), 1, "user1", eventCreatedTime.Unix())
	assert.Nil(t, err)

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
	err = p.ResetForNewUser(userId, userCreatedTime.Unix())
	assert.Nil(t, err)
	_, err = p.CountForEvent("A", event1CreatedTime.Unix(), make(map[string]interface{}),
	 make(map[string]interface{}), 1, userId, userCreatedTime.Unix())
	assert.Nil(t, err)
	_, err = p.CountForEvent("B", event2CreatedTime, make(map[string]interface{}),
	make(map[string]interface{}), 1, userId, userCreatedTime.Unix())
	assert.NotNil(t, err)*/

	// Event2 and Event1 are out of order.
	userId := "user2"
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	event1CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	event2CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:59:59Z")
	err = p.ResetForNewUser(userId, userCreatedTime.Unix())
	assert.Nil(t, err)
	err = p.CountForEvent("A", event1CreatedTime.Unix(), make(map[string]interface{}),
		make(map[string]interface{}), 1, userId, userCreatedTime.Unix())
	assert.Nil(t, err)
	err = p.CountForEvent("B", event2CreatedTime.Unix(), make(map[string]interface{}),
		make(map[string]interface{}), 1, userId, userCreatedTime.Unix())
	assert.NotNil(t, err)
}

func TestAddNumericAndCategoricalProperties(t *testing.T) {
	properties := map[string]interface{}{
		"catProperty":      "1",
		"numProperty":      1,
		"$qp_utm_campaign": 123456,
		"$qp_utm_keyword":  "analytics",
		"$campaign_id":     23456,
		"$cost":            10.0,
	}
	nMap := make(map[string]float64)
	cMap := make(map[string]string)
	P.AddNumericAndCategoricalProperties(0, properties, nMap, cMap)
	assert.Contains(t, cMap, "0.catProperty")
	assert.Contains(t, nMap, "0.numProperty")
	assert.Contains(t, cMap, "0.$qp_utm_campaign")
	assert.Contains(t, cMap, "0.$qp_utm_keyword")
	assert.Contains(t, cMap, "0.$campaign_id")
	assert.Contains(t, nMap, "0.$cost")

	utm_campaign, _ := cMap["0.$qp_utm_campaign"]
	assert.Equal(t, "123456", utm_campaign)
	campaign_id, _ := cMap["0.$campaign_id"]
	assert.Equal(t, "23456", campaign_id)
	cost, _ := nMap["0.$cost"]
	assert.Equal(t, 10.0, cost)
}
