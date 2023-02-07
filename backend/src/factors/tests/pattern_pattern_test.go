package tests

import (
	"bufio"
	"encoding/json"
	"factors/model/model"
	P "factors/pattern"
	PS "factors/pattern_server/store"
	T "factors/task"
	U "factors/util"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPatternFromFile(t *testing.T) {
	file, err := os.Open("./data/pattern.json")
	assert.Nil(t, err)
	scanner := bufio.NewScanner(file)
	const initBufSize = 100 * 1024 // 100KB
	buf := make([]byte, initBufSize)
	const maxCapacity = 250 * 1024 * 1024
	scanner.Buffer(buf, maxCapacity)

	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}
	err = scanner.Err()
	assert.Nil(t, err)
	file.Close()

	pwm := PS.PatternWithMeta{}
	err = json.Unmarshal([]byte(txtlines[0]), &pwm)
	assert.Nil(t, err)
	p := P.Pattern{}
	err = json.Unmarshal([]byte(pwm.RawPattern), &p)
	assert.Equal(t, 1, len(p.EventNames))

	totalCount, _ := p.GetCount([]P.EventConstraints{P.EventConstraints{}}, P.COUNT_TYPE_PER_USER)
	assert.Equal(t, 893.0, float64(totalCount))

	patternConstraints := []P.EventConstraints{
		P.EventConstraints{
			EPCategoricalConstraints: []P.CategoricalConstraint{
				P.CategoricalConstraint{
					PropertyName:  "$keyword",
					PropertyValue: "lending_company",
				},
			},
		},
	}
	constraintCount, _ := p.GetCount(patternConstraints, P.COUNT_TYPE_PER_USER)
	assert.Equal(t, 31.0, float64(constraintCount))

	patternConstraints = []P.EventConstraints{
		P.EventConstraints{
			EPNumericConstraints: []P.NumericConstraint{
				P.NumericConstraint{
					PropertyName: "$page_load_time",
					LowerBound:   0.78,
					UpperBound:   7.28,
					IsEquality:   false,
				},
			},
		},
	}
	constraintCount, _ = p.GetCount(patternConstraints, P.COUNT_TYPE_PER_USER)
	assert.Equal(t, 304.0, float64(constraintCount))
}

func createEtsStruct(userId string, user1CreatedTime,
	nextEventCreatedTime time.Time, events []string,
	cardinalities []uint) []P.EvSameTs {
	etsList := make([]P.EvSameTs, 0)
	emap := make(map[string]P.CounterEventFormat)
	for idx, et := range events {
		enameList := []string{et}
		t := P.CounterEventFormat{
			UserId:            userId,
			UserJoinTimestamp: user1CreatedTime.Unix(),
			EventName:         et,
			EventTimestamp:    nextEventCreatedTime.Unix(),
			EventCardinality:  cardinalities[idx],
			EventProperties:   nil,
			UserProperties:    nil,
		}
		emap[et] = t
		tmp := P.EvSameTs{EventsNames: enameList, EventTimestamp: nextEventCreatedTime.Unix()}
		tmp.EventsMap = make(map[string]P.CounterEventFormat)
		for k, v := range emap {
			tmp.EventsMap[k] = v
		}
		etsList = append(etsList, tmp)
		nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}

	return etsList

}

func TestPatternCountEvents(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	var cAlgoProps P.CountAlgoProperties
	cAlgoProps.Counting_version = 1
	cAlgoProps.Hmine_support = 0.0001
	cAlgoProps.Hmine_persist = 0
	// Count A -> B -> C
	// U1: F, B, A, C, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: B, A, A, C, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count:3, OncePerUserCount:2, UserCount:2
	pCountOccur := []bool{true}
	for _, countOccurFlag := range pCountOccur {

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

		etsList := createEtsStruct(userId, user1CreatedTime, nextEventCreatedTime, events, cardinalities)

		for i, _ := range events {
			err = p.CountForEvent(project.ID, userId, user1CreatedTime.Unix(), countOccurFlag, etsList[i], cAlgoProps)
			assert.Nil(t, err)
		}

		// User 2 events.
		userId = "user2"
		user2CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:01:00Z")
		err = p.ResetForNewUser(userId, user2CreatedTime.Unix())
		assert.Nil(t, err)
		nextEventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:01:00Z")
		events = []string{"B", "A", "A", "C", "B", "Z", "C", "A", "B", "C"}
		cardinalities = []uint{1, 1, 2, 1, 2, 1, 2, 3, 3, 3}
		etsList = createEtsStruct(userId, user1CreatedTime, nextEventCreatedTime, events, cardinalities)

		for i, _ := range events {
			err = p.CountForEvent(project.ID, userId, user2CreatedTime.Unix(), countOccurFlag, etsList[i], cAlgoProps)
			assert.Nil(t, err)
			// nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
		}

		err = p.ResetAfterLastUser()

		assert.Nil(t, err)

		assert.Equal(t, uint(0), p.PerOccurrenceCount)
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
}

func TestPatternGetPerUserCount(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	var cAlgoProps P.CountAlgoProperties
	cAlgoProps.Counting_version = 1
	cAlgoProps.Hmine_support = 0.0001
	cAlgoProps.Hmine_persist = 0
	// Count A -> B -> C
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count:3, OncePerUserCount:2, UserCount:2
	countOccurFlag := true
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

	etsList := createEtsStruct(userId, user1CreatedTime, nextEventCreatedTime, events, cardinalities)

	for i, _ := range events {
		err = p.CountForEvent(project.ID, userId, user1CreatedTime.Unix(), countOccurFlag, etsList[i], cAlgoProps)
		assert.Nil(t, err)
		// nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}
	// User 2 events.
	userId = "user2"
	user2CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:01:00Z")
	err = p.ResetForNewUser(userId, user2CreatedTime.Unix())
	assert.Nil(t, err)
	nextEventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:01:00Z")
	events = []string{"B", "A", "A", "C", "B", "Z", "C", "A", "B", "C"}
	cardinalities = []uint{1, 1, 2, 1, 2, 1, 2, 3, 3, 3}

	etsList = createEtsStruct(userId, user1CreatedTime, nextEventCreatedTime, events, cardinalities)

	for i, _ := range events {
		err = p.CountForEvent(project.ID, userId, user2CreatedTime.Unix(), countOccurFlag, etsList[i], cAlgoProps)
		assert.Nil(t, err)
		// nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
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

	assert.Equal(t, uint(0), p.PerOccurrenceCount)
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
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	countOccurFlag := true
	var cAlgoProps P.CountAlgoProperties
	cAlgoProps.Counting_version = 1
	cAlgoProps.Hmine_support = 0.0001
	cAlgoProps.Hmine_persist = 0
	// Test NewPattern with empty array.
	p, err := P.NewPattern([]string{}, nil)
	assert.NotNil(t, err)
	assert.Nil(t, p)

	// Test NewPattern with repeated elements.
	p, err = P.NewPattern([]string{"A", "B", "A", "C"}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, p)

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
	etsList := createEtsStruct("user1", userCreatedTime, eventCreatedTime, []string{"J"}, []uint{1})
	err = p.CountForEvent(project.ID, "user1", userCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
	assert.NotNil(t, err)

	// Test Count Event, with wrong userId or wrong userCreatedTime.
	p, err = P.NewPattern([]string{"A", "B", "C"}, nil)
	assert.Nil(t, err)
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	eventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	err = p.ResetForNewUser("user1", userCreatedTime.Unix())
	assert.Nil(t, err)
	etsList = createEtsStruct("user1", userCreatedTime, eventCreatedTime, []string{"J"}, []uint{1})
	err = p.CountForEvent(project.ID, "user1", userCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
	assert.Nil(t, err)
	// Wrong userId.
	etsList = createEtsStruct("user2", userCreatedTime, eventCreatedTime, []string{"J"}, []uint{1})
	err = p.CountForEvent(project.ID, "user2", userCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
	assert.NotNil(t, err)
	// Wrong userCreatedTime. Error is ignored.
	etsList = createEtsStruct("user1", eventCreatedTime, eventCreatedTime, []string{"J"}, []uint{1})

	err = p.CountForEvent(project.ID, "user1", eventCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
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
	etsList = createEtsStruct(userId, userCreatedTime, event1CreatedTime, []string{"A"}, []uint{1})
	err = p.CountForEvent(project.ID, userId, userCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
	assert.Nil(t, err)
	etsList = createEtsStruct(userId, userCreatedTime, event2CreatedTime, []string{"B"}, []uint{1})
	err = p.CountForEvent(project.ID, userId, userCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
	assert.NotNil(t, err)

}

func TestAddNumericAndCategoricalProperties(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
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
	P.AddNumericAndCategoricalProperties(project.ID, "event1", 0, properties, nMap, cMap, false)
	assert.Contains(t, cMap, "0.catProperty")
	assert.Contains(t, nMap, "0.numProperty")
	assert.Contains(t, nMap, "0.$qp_utm_campaign")
	assert.Contains(t, cMap, "0.$qp_utm_keyword")
	assert.Contains(t, cMap, "0.$campaign_id")
	assert.Contains(t, nMap, "0.$cost")
	utm_campaign, _ := nMap["0.$qp_utm_campaign"]
	assert.Equal(t, float64(123456), utm_campaign)
	campaign_id, _ := cMap["0.$campaign_id"]
	assert.Equal(t, "23456", campaign_id)
	cost, _ := nMap["0.$cost"]
	assert.Equal(t, 10.0, cost)
	numProperty, _ := nMap["0.numProperty"]
	assert.Equal(t, float64(1), numProperty)
}

func TestPatternCountEventsOccurenceFalse(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	var cAlgoProps P.CountAlgoProperties
	cAlgoProps.Counting_version = 1
	cAlgoProps.Hmine_support = 0.0001
	cAlgoProps.Hmine_persist = 0
	// Count A -> B -> C
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count:3, OncePerUserCount:2, UserCount:2
	countOccurFlag := false
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
	etsList := createEtsStruct(userId, user1CreatedTime, nextEventCreatedTime, events, cardinalities)

	for i, _ := range events {
		err = p.CountForEvent(project.ID, userId, user1CreatedTime.Unix(), countOccurFlag, etsList[i], cAlgoProps)
		assert.Nil(t, err)
		// nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}

	// User 2 events.
	userId = "user2"
	user2CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:01:00Z")
	err = p.ResetForNewUser(userId, user2CreatedTime.Unix())
	assert.Nil(t, err)
	nextEventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:01:00Z")
	events = []string{"B", "A", "A", "C", "B", "Z", "C", "A", "B", "C"}
	cardinalities = []uint{1, 1, 2, 1, 2, 1, 2, 3, 3, 3}
	etsList = createEtsStruct(userId, user2CreatedTime, nextEventCreatedTime, events, cardinalities)

	for i, _ := range events {
		err = p.CountForEvent(project.ID, userId, user2CreatedTime.Unix(), countOccurFlag, etsList[i], cAlgoProps)
		assert.Nil(t, err)
		// nextEventCreatedTime = nextEventCreatedTime.Add(time.Second * 60)
	}

	err = p.ResetAfterLastUser()

	assert.Nil(t, err)

	assert.Equal(t, uint(0), p.PerOccurrenceCount)
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

func TestPatternGetPerUserCountWithOccurenceFalse(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	var cAlgoProps P.CountAlgoProperties
	cAlgoProps.Counting_version = 1
	cAlgoProps.Hmine_support = 0.0001
	cAlgoProps.Hmine_persist = 0
	// Count A -> B -> C
	// U1: F, G, A, L, B, A, B, C   (A(1) -> B(2) -> C(1):1)
	// U2: F, A, A, K, B, Z, C, A, B, C  (A(2,1) -> B (1, 1) -> C(1, 1)
	// Count:3, OncePerUserCount:2, UserCount:2
	countOccurFlag := false

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
	etsList := createEtsStruct(userId, user1CreatedTime, nextEventCreatedTime, events, cardinalities)

	for i, _ := range events {
		err = p.CountForEvent(project.ID, userId, user1CreatedTime.Unix(), countOccurFlag, etsList[i], cAlgoProps)
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
	etsList = createEtsStruct(userId, user2CreatedTime, nextEventCreatedTime, events, cardinalities)

	for i, _ := range events {
		err = p.CountForEvent(project.ID, userId, user2CreatedTime.Unix(), countOccurFlag, etsList[i], cAlgoProps)
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

	assert.Equal(t, uint(0), p.PerOccurrenceCount)
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

func TestPatternEdgeConditionsOccureceFalse(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	countOccurFlag := false

	var cAlgoProps P.CountAlgoProperties
	cAlgoProps.Counting_version = 1
	cAlgoProps.Hmine_support = 0.0001
	cAlgoProps.Hmine_persist = 0

	// Test NewPattern with empty array.
	p, err := P.NewPattern([]string{}, nil)
	assert.NotNil(t, err)
	assert.Nil(t, p)

	// Test NewPattern with repeated elements.
	p, err = P.NewPattern([]string{"A", "B", "A", "C"}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, p)

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
	events := []string{"J"}
	cardinalities := []uint{1}
	userId := "user1"
	etsList := createEtsStruct(userId, userCreatedTime, eventCreatedTime, events, cardinalities)

	err = p.CountForEvent(project.ID, "user1", userCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
	assert.NotNil(t, err)

	// Test Count Event, with wrong userId or wrong userCreatedTime.
	p, err = P.NewPattern([]string{"A", "B", "C"}, nil)
	assert.Nil(t, err)
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	eventCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	err = p.ResetForNewUser("user1", userCreatedTime.Unix())
	assert.Nil(t, err)

	etsList = createEtsStruct(userId, userCreatedTime, eventCreatedTime, events, cardinalities)

	err = p.CountForEvent(project.ID, userId, userCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
	assert.Nil(t, err)
	// Wrong userId.
	events = []string{"J"}
	etsList = createEtsStruct("user2", userCreatedTime, eventCreatedTime, events, cardinalities)
	err = p.CountForEvent(project.ID, "user2", userCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
	assert.NotNil(t, err)
	etsList = createEtsStruct("user1", eventCreatedTime, eventCreatedTime, events, cardinalities)

	// Wrong userCreatedTime. Error is ignored.
	err = p.CountForEvent(project.ID, "user1", eventCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
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
	userId = "user2"
	userCreatedTime, _ = time.Parse(time.RFC3339, "2017-06-01T00:00:00Z")
	event1CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T01:00:00Z")
	event2CreatedTime, _ := time.Parse(time.RFC3339, "2017-06-01T00:59:59Z")
	err = p.ResetForNewUser(userId, userCreatedTime.Unix())
	assert.Nil(t, err)
	etsList = createEtsStruct(userId, userCreatedTime, event1CreatedTime, []string{"A"}, []uint{1})
	err = p.CountForEvent(project.ID, userId, userCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
	assert.Nil(t, err)
	etsList = createEtsStruct(userId, userCreatedTime, event2CreatedTime, []string{"B"}, []uint{1})
	err = p.CountForEvent(project.ID, userId, userCreatedTime.Unix(), countOccurFlag, etsList[0], cAlgoProps)
	assert.NotNil(t, err)

}

func TestPatternFilterTopKpatternTypes(t *testing.T) {

	topUCevents := []string{"uc1", "uc2", "uc3", "uc4", "uc5"}
	topPageViewEvents := []string{"pgv1", "pgv2", "pgv3", "$sp4", "pgv5"}
	topIEEvents := []string{"ie1", "ie2", "ie3", "ie4", "ie5"}
	topSpecEvents := []string{"$sp1", "$sp2", "$sp3", "$sp4", "$sp5"}
	topCampEvents := []string{"$session[campaign=1]", "$session[campaign=2]", "$session[campaign=3]", "$session[campaign=4]", "$session[campaign=5]"}
	var cAlgoProps P.CountAlgoProperties
	cAlgoProps.Counting_version = 1
	cAlgoProps.Hmine_support = 0.0001
	cAlgoProps.Hmine_persist = 0
	patterns := make([]*P.Pattern, 0)
	eNT := make(map[string]string)

	for idx := 0; idx < len(topUCevents); idx++ {

		tmpUC, _ := P.NewPattern([]string{topUCevents[idx]}, nil)
		tmpPv, _ := P.NewPattern([]string{topPageViewEvents[idx]}, nil)
		tmpIE, _ := P.NewPattern([]string{topIEEvents[idx]}, nil)
		tmpSpec, _ := P.NewPattern([]string{topSpecEvents[idx]}, nil)
		tmpCamp, _ := P.NewPattern([]string{topCampEvents[idx]}, nil)

		tmpUC.PerUserCount = uint(idx)
		tmpPv.PerUserCount = uint(idx)
		tmpIE.PerUserCount = uint(idx)
		tmpSpec.PerUserCount = uint(idx)
		tmpCamp.PerUserCount = uint(idx)

		patterns = append(patterns, tmpUC)
		patterns = append(patterns, tmpPv)
		patterns = append(patterns, tmpIE)
		patterns = append(patterns, tmpSpec)
		patterns = append(patterns, tmpCamp)

		eNT[topUCevents[idx]] = model.TYPE_USER_CREATED_EVENT_NAME
		eNT[topPageViewEvents[idx]] = model.TYPE_FILTER_EVENT_NAME
		eNT[topIEEvents[idx]] = model.TYPE_INTERNAL_EVENT_NAME
		eNT[topSpecEvents[idx]] = "specialEvents"
		eNT[topCampEvents[idx]] = "CampaignEvents"
	}
	var ca T.CampaignEventLists
	ca.CampaignList = []string{"$session[campaign=1]", "$session[campaign=2]", "$session[campaign=3]", "$session[campaign=4]", "$session[campaign=5]"}
	filterdPatterns := T.FilterTopKEventsOnTypes(patterns, eNT, 3, 3, 3, ca)
	for _, v := range filterdPatterns {
		fmt.Println(v.EventNames)
	}

	// $sp4 is repeated hence count is reduced by 1 ,
	// campaign type should not counted
	assert.Equal(t, 11, len(filterdPatterns))
}

func TestGenCombinationPatternsEndingWithGoal(t *testing.T) {
	var err bool
	err = false

	allEvents := []string{"uc1", "uc2", "pgv1", "pgv2", "ie1", "ie2", "$sp1", "$sp2"}
	goalEvents := []string{"G1", "G2", "G3", "G4", "G5"}

	allPatterns := make([]*P.Pattern, 0)
	for idx := 0; idx < len(allEvents); idx++ {

		tmpAll, _ := P.NewPattern([]string{allEvents[idx]}, nil)
		tmpAll.PerUserCount = uint(idx)
		allPatterns = append(allPatterns, tmpAll)

	}

	goalPatterns := make([]*P.Pattern, 0)
	for idx := 0; idx < len(goalEvents); idx++ {

		tmpGoal, _ := P.NewPattern([]string{goalEvents[idx]}, nil)
		tmpGoal.PerUserCount = uint(idx)
		goalPatterns = append(goalPatterns, tmpGoal)

	}
	filterdPatterns, _, _ := P.GenCombinationPatternsEndingWithGoal(0, allPatterns, goalPatterns, nil)
	for _, f := range filterdPatterns {
		if f.EventNames[0] == f.EventNames[1] {
			err = true
		}
		assert.Equal(t, err, false, "Both start and goal patterns are equal")
	}

	assert.Equal(t, 40, len(filterdPatterns), "total number of patterns")
}

func TestGenRepeatedEventCandidates(t *testing.T) {
	cycEvents := []string{"A", "B", "C"}
	sp1, _ := P.NewPattern([]string{"A", "B"}, nil)
	sp2, _ := P.NewPattern([]string{"A", "Z"}, nil)
	sp3, _ := P.NewPattern([]string{"A", "A"}, nil)
	sp4, _ := P.NewPattern([]string{"A", "D"}, nil)
	sp5, _ := P.NewPattern([]string{"A", "E"}, nil)
	sp6, _ := P.NewPattern([]string{"A", "F", "B", "C"}, nil)
	startPatterns := []*P.Pattern{sp1, sp2, sp3, sp4, sp5, sp6}
	cMap := make(map[string]bool)
	for _, pt := range startPatterns {
		tmp, _ := P.GenRepeatedEventCandidates(cycEvents, pt, nil)
		for _, v := range tmp {
			cMap[v.String()] = true
		}
	}
	assert.Equal(t, true, cMap["A,A,A"])
	assert.Equal(t, true, cMap["A,A,D"])
	assert.Equal(t, true, cMap["A,A,E"])
	assert.Equal(t, false, cMap["A,B,A"])
	assert.Equal(t, true, cMap["A,B,B"])

	cycEvents = []string{}
	sp7, _ := P.NewPattern([]string{"A"}, nil)
	emptyPt, err := P.GenRepeatedEventCandidates(cycEvents, sp7, nil)
	assert.Equal(t, 0, len(emptyPt))
	assert.Nil(t, err, "Empty cyclic set with no error")

}

func TestGetTopURLs(t *testing.T) {
	url1, _ := P.NewPattern([]string{"http://www.abc1.com"}, nil)
	url2, _ := P.NewPattern([]string{"http://www.abc2.com"}, nil)
	url3, _ := P.NewPattern([]string{"http://www.abc3.com"}, nil)
	url4, _ := P.NewPattern([]string{"http://www.abc4.com"}, nil)
	url5, _ := P.NewPattern([]string{"www.abc5.com"}, nil)
	url6, _ := P.NewPattern([]string{"www.abc6.com"}, nil)
	url7, _ := P.NewPattern([]string{"abc7.com"}, nil)

	UD1, _ := P.NewPattern([]string{"UD1"}, nil)
	UD2, _ := P.NewPattern([]string{"UD2"}, nil)
	UD3, _ := P.NewPattern([]string{"UD3"}, nil)
	p1, _ := P.NewPattern([]string{"$hub1"}, nil)
	p2, _ := P.NewPattern([]string{"$hub2"}, nil)
	p3, _ := P.NewPattern([]string{"$hub3"}, nil)

	url1.PerUserCount = 1
	url2.PerUserCount = 2
	url3.PerUserCount = 3
	url4.PerUserCount = 4
	url5.PerUserCount = 5
	url6.PerUserCount = 6
	url7.PerUserCount = 6
	UD1.PerUserCount = 14
	UD2.PerUserCount = 12
	UD3.PerUserCount = 10

	urlPatterns := []*P.Pattern{url1, url2, url3, url4, url5, url6, url7, UD1, UD2, UD3, p1, p2, p3}
	topk := 4
	filteredPatterns := T.GetTopURLs(urlPatterns, topk)
	assert.Equal(t, topk, len(filteredPatterns))
	resPattern := []*P.Pattern{url7, url6, url4, url5}
	assert.ElementsMatch(t, filteredPatterns, resPattern, "Not all top patterns found")

}

func TestGetTopUDE(t *testing.T) {
	UD1, _ := P.NewPattern([]string{"UD1"}, nil)
	UD2, _ := P.NewPattern([]string{"UD2"}, nil)
	UD3, _ := P.NewPattern([]string{"UD3"}, nil)
	UD4, _ := P.NewPattern([]string{"UD4"}, nil)
	UD5, _ := P.NewPattern([]string{"UD5"}, nil)
	UD6, _ := P.NewPattern([]string{"UD6"}, nil)

	UD1.PerUserCount = 1
	UD2.PerUserCount = 2
	UD3.PerUserCount = 3
	UD4.PerUserCount = 4
	UD5.PerUserCount = 5
	UD6.PerUserCount = 6

	urlPatterns := []*P.Pattern{UD1, UD2, UD3, UD4, UD5, UD6}
	topk := 3
	eventNamesWithType := make(map[string]string, 0)
	eventNamesWithType["UD1"] = "UC"
	eventNamesWithType["UD2"] = "UC"
	eventNamesWithType["UD3"] = "UC"
	eventNamesWithType["UD4"] = "UC"
	eventNamesWithType["UD5"] = "UC"
	eventNamesWithType["UD6"] = "UC"
	filteredPatterns := T.GetTopUDE(urlPatterns, eventNamesWithType, topk)
	assert.Equal(t, topk, len(filteredPatterns))
	resPattern := []*P.Pattern{UD6, UD5, UD4}
	assert.ElementsMatch(t, filteredPatterns, resPattern, "Not all top patterns found")

}

func TestGetTopStandardPatterns(t *testing.T) {
	p1, _ := P.NewPattern([]string{"$hub1"}, nil)
	p2, _ := P.NewPattern([]string{"$hub2"}, nil)
	p3, _ := P.NewPattern([]string{"$hub3"}, nil)
	p4, _ := P.NewPattern([]string{"$hub4"}, nil)
	p5, _ := P.NewPattern([]string{"$hub5"}, nil)
	p6, _ := P.NewPattern([]string{"$hub6"}, nil)

	p1.PerUserCount = 1
	p2.PerUserCount = 2
	p3.PerUserCount = 3
	p4.PerUserCount = 4
	p5.PerUserCount = 5
	p6.PerUserCount = 6

	stanPatterns := []*P.Pattern{p1, p2, p3, p4, p5, p6}
	topk := 3
	filteredPatterns := T.GetTopStandardPatterns(stanPatterns, topk)
	assert.Equal(t, topk, len(filteredPatterns))
	resPattern := []*P.Pattern{p6, p5, p4}
	assert.ElementsMatch(t, filteredPatterns, resPattern, "Not all top patterns found")

	topk = -1
	filteredPatterns = T.GetTopStandardPatterns(stanPatterns, topk)
	assert.Equal(t, len(stanPatterns), len(filteredPatterns))
	resPattern = []*P.Pattern{p6, p5, p4, p3, p2, p1}
	assert.ElementsMatch(t, filteredPatterns, resPattern, "Not all top patterns found")

}

func TestGetTopCampaigns(t *testing.T) {
	p1, _ := P.NewPattern([]string{"$session[campaign=1]"}, nil)
	p2, _ := P.NewPattern([]string{"$session[campaign=2]"}, nil)
	p3, _ := P.NewPattern([]string{"$session[campaign=3]"}, nil)
	p4, _ := P.NewPattern([]string{"$session[campaign=4]"}, nil)
	p5, _ := P.NewPattern([]string{"$session[campaign=5]"}, nil)
	p6, _ := P.NewPattern([]string{"$session[campaign=6]"}, nil)

	p1.PerUserCount = 1
	p2.PerUserCount = 2
	p3.PerUserCount = 3
	p4.PerUserCount = 4
	p5.PerUserCount = 5
	p6.PerUserCount = 6

	urlPatterns := []*P.Pattern{p1, p2, p3, p4, p5, p6}
	topk := 3
	filteredPatterns := T.GetTopCampaigns(urlPatterns, topk)
	assert.Equal(t, topk, len(filteredPatterns))
	resPattern := []*P.Pattern{p6, p5, p4}
	assert.ElementsMatch(t, filteredPatterns, resPattern, "Not all top patterns found")

}

func TestGenMissingJourneyPatterns(t *testing.T) {

	p1, _ := P.NewPattern([]string{"a", "b", "c"}, nil)
	p2, _ := P.NewPattern([]string{"a", "b", "d"}, nil)
	p3, _ := P.NewPattern([]string{"a", "c", "e"}, nil)
	p4, _ := P.NewPattern([]string{"a", "k", "e"}, nil)
	q1, _ := P.NewPattern([]string{"a", "b"}, nil)
	q2, _ := P.NewPattern([]string{"a", "c"}, nil)
	q3, _ := P.NewPattern([]string{"a", "d"}, nil)

	threeLen := []*P.Pattern{p1, p2, p3, p4}
	twoLen := []*P.Pattern{q1, q2, q3}

	pt, _ := T.GenMissingJourneyPatterns(threeLen, twoLen, nil)
	assert.Equal(t, 1, len(pt), "not Counting all missing two Level")

	pt, err := T.GenMissingJourneyPatterns(twoLen, threeLen, nil)
	assert.NotNil(t, err, err)

	p1, _ = P.NewPattern([]string{"a", "a", "c"}, nil)
	p2, _ = P.NewPattern([]string{"a", "a", "d"}, nil)
	p3, _ = P.NewPattern([]string{"a", "a", "e"}, nil)
	p4, _ = P.NewPattern([]string{"a", "k", "e"}, nil)
	q1, _ = P.NewPattern([]string{"a", "a"}, nil)
	q2, _ = P.NewPattern([]string{"a", "k"}, nil)
	q3, _ = P.NewPattern([]string{"a", "d"}, nil)

	threeLen = []*P.Pattern{p1, p2, p3, p4}
	twoLen = []*P.Pattern{q1, q2, q3}

	pt, _ = T.GenMissingJourneyPatterns(threeLen, twoLen, nil)
	assert.Equal(t, 0, len(pt), "not Counting all missing two Level")

	p1, _ = P.NewPattern([]string{"a", "a"}, nil)
	p2, _ = P.NewPattern([]string{"b", "b"}, nil)
	p3, _ = P.NewPattern([]string{"c", "a"}, nil)
	p4, _ = P.NewPattern([]string{"d", "e"}, nil)
	q1, _ = P.NewPattern([]string{"a"}, nil)
	q2, _ = P.NewPattern([]string{"b"}, nil)
	q3, _ = P.NewPattern([]string{"c"}, nil)

	twoLen = []*P.Pattern{p1, p2, p3, p4}
	oneLen := []*P.Pattern{q1, q2, q3}

	pt, _ = T.GenMissingJourneyPatterns(twoLen, oneLen, nil)
	assert.Equal(t, 1, len(pt), "not Counting all missing two Level")

}

func TestGenRepeatedCombinations(t *testing.T) {

	q1, _ := P.NewPattern([]string{"a", "e"}, nil)
	q2, _ := P.NewPattern([]string{"b", "k"}, nil)
	q3, _ := P.NewPattern([]string{"c", "d"}, nil)
	q4, _ := P.NewPattern([]string{"d", "l"}, nil)

	p1, _ := P.NewPattern([]string{"a", "a", "e"}, nil)
	p2, _ := P.NewPattern([]string{"b", "b", "k"}, nil)
	p3, _ := P.NewPattern([]string{"c", "c", "d"}, nil)
	// p4, _ := P.NewPattern([]string{"d", "d", "l"}, nil)

	lenTwoPatt := []*P.Pattern{q1, q2, q3, q4}
	lenTwoMatch := []*P.Pattern{p1, p2, p3}
	// lenTwofail := []*P.Pattern{q4}

	repeaptedEvents := []string{"a", "b", "c"}
	repeatedEventsMap := make(map[string]bool, 0)
	repeatedEventsMap["a"] = true
	repeatedEventsMap["b"] = true
	repeatedEventsMap["c"] = true

	pt, err := T.GenRepeatedCombinations(lenTwoPatt, nil, repeaptedEvents)
	assert.Nil(t, err)

	assert.ElementsMatch(t, pt, lenTwoMatch, "Not all repeated elemets found")
}

func TestGenInterMediateCombinations(t *testing.T) {

	//result will be {"a","b","g"} {"b","a","g"}
	q1, _ := P.NewPattern([]string{"a", "g"}, nil)
	q2, _ := P.NewPattern([]string{"b", "g"}, nil)

	lenTwoPatt := []*P.Pattern{q1, q2}

	patts, err := T.GenInterMediateCombinations(lenTwoPatt, nil)
	assert.Nil(t, err, "erorr is not nil")
	assert.Equal(t, 2, len(patts), "number of patterns generated is not matched")
	for _, v := range patts {
		fmt.Println(v.EventNames)
	}

	var flagCount = true
	for _, v := range patts {
		if len(v.EventNames) != 3 {
			flagCount = false
		}
	}
	assert.True(t, flagCount, "len not equal to three")

}

func TestFilteringPatternsNotMatching(t *testing.T) {

	// var rl model.FactorsGoalRule
	var qr model.ExplainV2Query
	qr.Title = "test"
	qr.StartTimestamp = 1672491600
	qr.EndTimestamp = 1672491601
	eventsList := make([]P.CounterEventFormat, 0)

	eventsListTrue := make([]P.CounterEventFormat, 0)

	file, err := os.Open("./data/events_filter.txt")
	assert.Nil(t, err)

	filter_string := []byte(`{"st_en":"www.acme.com","en_en":"www.acme.com/pricing","rule":{"st_en_ft":[],"en_en_ft":[],"st_us_ft":[{"key":"Country","vl":"US","operator":true,"lower_bound":0,"upper_bound":0,"property_type":"categorical"}],"en_us_ft":[{"key":"Country","vl":"US","operator":true,"lower_bound":0,"upper_bound":0,"property_type":"categorical"}],"ft":[],"in_en":[],"in_epr":null,"in_upr":null},"vs":false}`)
	err = json.Unmarshal(filter_string, &qr.Query)

	assert.Nil(t, err, "Unable to decode filter string")
	qr.Raw_query = string(filter_string)
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println(qr.Query)
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$")
	scanner := bufio.NewScanner(file)

	upg := make(map[string]string)
	epg := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return
		}
		eventsList = append(eventsList, eventDetails)
		if T.FilterEventsOnRule(eventDetails, qr, upg, epg) {
			eventsListTrue = append(eventsListTrue, eventDetails)
			_, _ = fmt.Println(eventDetails)
			fmt.Println("---------")
		}
	}

	assert.Equal(t, 18, len(eventsList))
	assert.Equal(t, 0, len(eventsListTrue))

}

func TestFilteringPatternsMatching(t *testing.T) {

	// var rl model.FactorsGoalRule
	var qr model.ExplainV2Query
	qr.Title = "test"
	qr.StartTimestamp = 1672491600
	qr.EndTimestamp = 1672491601
	eventsList := make([]P.CounterEventFormat, 0)

	eventsListTrue := make([]P.CounterEventFormat, 0)

	file, err := os.Open("./data/events_filter.txt")
	// file, err := os.Open("/Users/vinithkumar/work/data/events_20230101-20230108.txt")
	assert.Nil(t, err)

	filter_string := []byte(`{"st_en":"$session","en_en":"$form_submitted","rule":{"st_en_ft":[],"en_en_ft":[],"st_us_ft":[{"key":"$country","vl":"India","operator":true,"lower_bound":0,"upper_bound":0,"property_type":"categorical"},{"key":"$city","vl":"Chennai","operator":true,"lower_bound":0,"upper_bound":0,"property_type":"categorical"}],"en_us_ft":[{"key":"$country","vl":"United States","operator":false,"lower_bound":0,"upper_bound":0,"property_type":"categorical"}],"ft":[],"in_en":[],"in_epr":null,"in_upr":null},"vs":false}`)
	// filter_string := []byte(`{"en_en":"app.factors.ai","rule":{"en_en_ft":null,"en_us_ft":null,"ft":null,"in_en":["$hubspot_contact_updated","www.factors.ai/features","RUN-QUERY","$hubspot_engagement_email","VIEW_DASHBOARD","www.factors.ai/pricing","www.factors.ai","staging-app.factors.ai","app.factors.ai/analyse"],"in_epr":null,"in_upr":null,"st_en_ft":null,"st_us_ft":null},"st_en":"$session","vs":false}`)
	err = json.Unmarshal(filter_string, &qr.Query)
	assert.Nil(t, err, "Unable to decode filter string")
	qr.Raw_query = string(filter_string)
	fmt.Println(qr.Query)
	scanner := bufio.NewScanner(file)
	emap := make(map[string]bool)
	upg := make(map[string]string)
	epg := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return
		}
		eventsList = append(eventsList, eventDetails)

		if T.FilterEventsOnRule(eventDetails, qr, upg, epg) == true {
			eventsListTrue = append(eventsListTrue, eventDetails)
			_, _ = fmt.Println(eventDetails.EventName)
			emap[eventDetails.EventName] = true
			fmt.Println(eventDetails)
			fmt.Println("====")
		}
	}

	for k, _ := range emap {
		fmt.Println(fmt.Sprintf("key --> %s", k))
	}
	assert.Equal(t, 18, len(eventsList))
	assert.Equal(t, 5, len(eventsListTrue))

}
