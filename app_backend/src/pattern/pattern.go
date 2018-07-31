package pattern

import (
	"fmt"
	Hist "gohistogram"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Pattern struct {
	EventNames         []string
	Timings            []Hist.NumericHistogram
	EventCardinalities []Hist.NumericHistogram
	Repeats            []Hist.NumericHistogram
	// The total number of times this pattern occurs allowing multiple counts
	// per user.
	Count uint
	// Counted only once per user.
	OncePerUserCount uint
	// Number of users the pattern was counted on.
	UserCount uint

	// Private variables.
	waitIndex                  int
	currentUserId              string
	currentUserCreatedTime     time.Time
	currentUserOccurrenceCount uint
	currentEventTimes          []time.Time
	currentEventCardinalities  []uint
	currentRepeats             []uint
}

const num_T_BINS = 20
const num_C_BINS = 10
const num_R_BINS = 10

func NewPattern(events []string) (*Pattern, error) {
	pLen := len(events)
	if pLen == 0 {
		err := fmt.Errorf("No events in pattern")
		return nil, err
	}
	if !isEventsUnique(events) {
		err := fmt.Errorf("Events are not unique")
		return nil, err
	}
	pattern := Pattern{
		EventNames:                 events,
		Timings:                    make([]Hist.NumericHistogram, pLen),
		EventCardinalities:         make([]Hist.NumericHistogram, pLen),
		Repeats:                    make([]Hist.NumericHistogram, pLen),
		Count:                      0,
		OncePerUserCount:           0,
		waitIndex:                  0,
		currentUserId:              "",
		currentUserCreatedTime:     time.Time{},
		currentUserOccurrenceCount: 0,
		currentEventTimes:          make([]time.Time, pLen),
		currentEventCardinalities:  make([]uint, pLen),
		currentRepeats:             make([]uint, pLen),
	}
	for i := 0; i < pLen; i++ {
		pattern.Timings[i] = *Hist.NewHistogram(num_T_BINS)
		pattern.EventCardinalities[i] = *Hist.NewHistogram(num_C_BINS)
		pattern.Repeats[i] = *Hist.NewHistogram(num_R_BINS)
	}
	return &pattern, nil
}

func isEventsUnique(eventNames []string) bool {
	pLen := len(eventNames)
	var pMap map[string]bool = make(map[string]bool)
	for i := 0; i < pLen; i++ {
		pMap[eventNames[i]] = true
	}
	return len(pMap) == pLen
}

func (p *Pattern) ResetForNewUser(userId string, userCreatedTime time.Time) error {
	if userId == "" || userCreatedTime.Equal(time.Time{}) {
		return fmt.Errorf("Missing userId or userCreatedTime.")
	}

	p.UserCount += 1
	// Reinitialize all private variables maintained per user event stream.
	p.waitIndex = 0
	p.currentUserId = userId
	p.currentUserCreatedTime = userCreatedTime
	p.currentUserOccurrenceCount = 0
	pLen := len(p.EventNames)
	p.currentEventTimes = make([]time.Time, pLen)
	p.currentRepeats = make([]uint, pLen)
	return nil
}

// If data is visualized in below format, where U(.) are users, E(.) are events,
// T(.) are timestamps.

// U1: E1(T1), E4(T2), E1(T3), E5(T4), E1(T5), E5(T6)
// U2: E3(T7), E4(T8), E5(T9), E1(T10)
// U3: E2(T11), E1(T12), E5(T13)

// The frequency of the event E1 -> E5 is 3 - twice non overlapping
// in U1 and once in U3 - i.e. [U1: E1(T1) -> E5(T4)] [U1: E1(T5) -> E5(T6)] and
// [U3: E1(T12) -> E5(T13)].
// Further the distribution of timestamps, event properties and number of occurrences
// are stored with the patterns.
func (p *Pattern) CountForEvent(eventName string, eventCreatedTime time.Time, eventCardinality uint, userId string, userCreatedTime time.Time) (string, error) {
	if eventName == "" || eventCreatedTime.Equal(time.Time{}) {
		return "", fmt.Errorf("Missing eventId or eventCreatedTime.")
	}

	if userId != p.currentUserId || !p.currentUserCreatedTime.Equal(userCreatedTime) {
		return "", fmt.Errorf(
			fmt.Sprintf("Mismatch in User data. userId: %s, userCreatedTime: %v, pattern userId: %s, pattern userCreatedTime: %v",
				userId, userCreatedTime, p.currentUserId, p.currentUserCreatedTime))
	}

	if p.waitIndex > 0 && eventName == p.EventNames[p.waitIndex-1] {
		// Repeats count the number of times the current event has occurred
		// before seeing the next event being waited upon.
		p.currentRepeats[p.waitIndex-1] += 1
	} else if eventName == p.EventNames[p.waitIndex] {
		// Record the event occurrence and wait on the next one.
		p.currentEventTimes[p.waitIndex] = eventCreatedTime
		p.currentEventCardinalities[p.waitIndex] = eventCardinality
		p.currentRepeats[p.waitIndex] = 1

		p.waitIndex += 1

		pLen := len(p.EventNames)
		if p.waitIndex == pLen {
			// Record the pattern occurrence.
			p.Count += 1
			p.currentUserOccurrenceCount += 1
			if p.currentUserOccurrenceCount == 1 {
				p.OncePerUserCount += 1
				// Update histograms only for the first count per user.
				// Update histograms of timings.
				var duration float64
				for i := 0; i < pLen; i++ {
					if i == 0 {
						duration = p.currentEventTimes[0].Sub(userCreatedTime).Seconds()
						if duration < 0 {
							// Ignoring this error for now, since there are no DB checks to avoid
							// these user input values.
							log.Error(fmt.Sprintf("Event occurs before creation for user:%s", p.currentUserId))
						}
					} else {
						duration = p.currentEventTimes[i].Sub(p.currentEventTimes[i-1]).Seconds()
						if duration < 0 {
							return "", fmt.Errorf("Event Timings not in order")
						}
					}
					p.Timings[i].Add(duration)
				}

				// Update histograms of repeats and Cardinalities.
				for i := 0; i < pLen; i++ {
					p.EventCardinalities[i].Add(float64(p.currentEventCardinalities[i]))
					p.Repeats[i].Add(float64(p.currentRepeats[i]))
				}
			}

			// Reset.
			p.currentEventTimes = make([]time.Time, pLen)
			p.currentEventCardinalities = make([]uint, pLen)
			p.currentRepeats = make([]uint, pLen)
			p.waitIndex = 0
		}
	}
	return p.EventNames[p.waitIndex], nil
}

func (p *Pattern) GetOncePerUserCount(eventCardinalityLowerBound int,
	eventCardinalityUpperBound int) uint {

	if eventCardinalityLowerBound > eventCardinalityUpperBound {
		log.WithFields(log.Fields{
			"eclb": eventCardinalityLowerBound,
			"ecub": eventCardinalityUpperBound}).Error("Unexpected cardinality bounds.")
		return p.OncePerUserCount
	}
	lowerCDF := 0.0
	upperCDF := 1.0
	pLen := len(p.EventNames)
	if eventCardinalityLowerBound > 0 {
		// The bounds are meant for the last event.
		lowerCDF = p.EventCardinalities[pLen-1].CDF(float64(eventCardinalityLowerBound) - 0.5)
	}
	if eventCardinalityUpperBound > 0 {
		upperCDF = p.EventCardinalities[pLen-1].CDF(float64(eventCardinalityUpperBound) + 0.5)
	}
	floatCount := (float64(p.OncePerUserCount) * (upperCDF - lowerCDF))
	if floatCount < 0 {
		log.WithFields(log.Fields{"upperCDF": upperCDF, "lowerCDF": lowerCDF,
			"eelb": eventCardinalityLowerBound,
			"ecub": eventCardinalityUpperBound, "pattern": p.String()}).Fatal(
			"Final count is less than 0.")
	}
	return uint(floatCount)
}

func (p *Pattern) WaitingOn() string {
	return p.EventNames[p.waitIndex]
}

func (p *Pattern) PrevWaitingOn() string {
	if p.waitIndex > 0 {
		return p.EventNames[p.waitIndex-1]
	}
	return ""
}

func (p *Pattern) String() string {
	return eventArrayToString(p.EventNames)
}

func eventArrayToString(eventNames []string) string {
	return strings.Join(eventNames, ",")
}
