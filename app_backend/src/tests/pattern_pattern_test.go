package tests

import (
	P "pattern"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPatternCountEvents(t *testing.T) {
	// Count A -> B -> C
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count:3, OncePerUserCount:2, EventCount: 18, UserCount:2
	pEvents := []string{"A", "B", "C"}
	p, err := P.NewPattern(pEvents)
	assert.Nil(t, err)
	assert.NotNil(t, p)
	pLen := len(pEvents)
	assert.Equal(t, pLen, len(p.EventNames))
	assert.Equal(t, pLen, len(p.Timings))
	assert.Equal(t, pLen, len(p.Repeats))
	assert.Equal(t, uint(0), p.Count)
	assert.Equal(t, uint(0), p.EventCount)
	assert.Equal(t, uint(0), p.UserCount)
	assert.Equal(t, uint(0), p.OncePerUserCount)
	// User 1 events.
	userId := "user1"
	userCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	err = p.ResetForNewUser(userId, userCreatedTime)
	assert.Nil(t, err)
	nextEventCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	events := []string{"F", "G", "A", "L", "B", "A", "B", "C"}
	for _, event := range events {
		err = p.CountForEvent(event, nextEventCreatedTime, userId, userCreatedTime)
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
	for _, event := range events {
		err = p.CountForEvent(event, nextEventCreatedTime, userId, userCreatedTime)
		assert.Nil(t, err)
		nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}

	assert.Equal(t, uint(3), p.Count)
	assert.Equal(t, uint(18), p.EventCount)
	assert.Equal(t, uint(2), p.UserCount)
	assert.Equal(t, uint(2), p.OncePerUserCount)
	assert.Equal(t, pLen, len(p.EventNames))
	for i := 0; i < pLen; i++ {
		assert.Equal(t, pEvents[i], p.EventNames[i])
	}
	assert.Equal(t, pLen, len(p.Timings))
	assert.Equal(t, pLen, len(p.Repeats))
	// A-B-C occurs 3 times, with first A occurring after 3720s in User1 and
	// 3660 and 4020s in User 2.
	// Repeats once before the  next B occurs in User2.
	assert.Equal(t, float64(3), p.Timings[0].Count())
	assert.Equal(t, float64((3720.0+3660.0+4020.0)/3), p.Timings[0].Mean())
	assert.Equal(t, float64((1.0+2.0+1.0)/3), p.Repeats[0].Mean())
	// A-B-C occurs 3 times, with first B following first A after 120s in User1 and
	// 180 and 60s in User 2.
	// Repeats once before the  next C occurs in User1.
	assert.Equal(t, float64(3), p.Timings[0].Count())
	assert.Equal(t, float64((120.0+180.0+60.0)/3), p.Timings[1].Mean())
	assert.Equal(t, float64((2.0+1.0+1.0)/3), p.Repeats[1].Mean())
	// A-B-C occurs 3 times, with first C following first B after 180s in User1 and
	// 120 and 60s in User 2.
	// Last event always is counted once.
	assert.Equal(t, float64(3), p.Timings[0].Count())
	assert.Equal(t, float64((180.0+120.0+60.0)/3), p.Timings[2].Mean())
	assert.Equal(t, float64((1.0+1.0+1.0)/3), p.Repeats[2].Mean())

}

func TestPatternEdgeConditions(t *testing.T) {
	// Test NewPattern with empty array.
	p, err := P.NewPattern([]string{})
	assert.NotNil(t, err)
	assert.Nil(t, p)

	// Test NewPattern with repeated elements.
	p, err = P.NewPattern([]string{"A", "B", "A", "C"})
	assert.NotNil(t, err)
	assert.Nil(t, p)

	// Test Empty Pattern Creation.
	p, err = P.NewPattern([]string{"A", "B", "C"})
	assert.Nil(t, err)
	assert.NotNil(t, p)
	pLen := 3
	assert.Equal(t, pLen, len(p.EventNames))
	assert.Equal(t, pLen, len(p.Timings))
	assert.Equal(t, pLen, len(p.Repeats))
	assert.Equal(t, uint(0), p.Count)
	assert.Equal(t, uint(0), p.EventCount)
	assert.Equal(t, uint(0), p.UserCount)
	assert.Equal(t, uint(0), p.OncePerUserCount)

	// Test ResetForNewUser without time or Id.
	p, err = P.NewPattern([]string{"A", "B", "C"})
	assert.Nil(t, err)
	assert.NotNil(t, p)
	userCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	err = p.ResetForNewUser("", userCreatedTime)
	assert.NotNil(t, err)
	err = p.ResetForNewUser("user1", time.Time{})
	assert.NotNil(t, err)

	// Test Count Event with User not initialized.
	p, err = P.NewPattern([]string{"A", "B", "C"})
	assert.Nil(t, err)
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	eventCreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	err = p.CountForEvent("J", eventCreatedTime, "user1", userCreatedTime)
	assert.NotNil(t, err)

	// Test Count Event, with wrong userId or wrong userCreatedTime.
	p, err = P.NewPattern([]string{"A", "B", "C"})
	assert.Nil(t, err)
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	eventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	err = p.ResetForNewUser("user1", userCreatedTime)
	assert.Nil(t, err)
	err = p.CountForEvent("J", eventCreatedTime, "user1", userCreatedTime)
	assert.Nil(t, err)
	// Wrong userId.
	err = p.CountForEvent("J", eventCreatedTime, "user2", userCreatedTime)
	assert.NotNil(t, err)
	// Wrong userCreatedTime
	err = p.CountForEvent("J", eventCreatedTime, "user1", eventCreatedTime)
	assert.NotNil(t, err)

	// Test Events out of order. Out of order events are noticed only when
	// the whole pattern is observed.
	p, err = P.NewPattern([]string{"A", "B"})
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
	err = p.CountForEvent("A", event1CreatedTime, userId, userCreatedTime)
	assert.Nil(t, err)
	err = p.CountForEvent("B", event2CreatedTime, userId, userCreatedTime)
	assert.NotNil(t, err)*/

	// Event2 and Event1 are out of order.
	userId := "user2"
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	event1CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	event2CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:59:59Z")
	err = p.ResetForNewUser(userId, userCreatedTime)
	assert.Nil(t, err)
	err = p.CountForEvent("A", event1CreatedTime, userId, userCreatedTime)
	assert.Nil(t, err)
	err = p.CountForEvent("B", event2CreatedTime, userId, userCreatedTime)
	assert.NotNil(t, err)
}
