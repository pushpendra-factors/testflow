package pattern

import (
	"fmt"
	"time"

	Hist "github.com/VividCortex/gohistogram"
)

type Pattern struct {
	EventNames []string
	Timings    []Hist.NumericHistogram
	Repeats    []Hist.NumericHistogram
	// The total number of times this pattern occurs allowing multiple counts
	// per user.
	Count uint
	// Counted only once per user.
	OncePerUserCount uint
	// Number of users the pattern was counted on.
	UserCount uint
	// Number of events the pattern was counted on.
	EventCount uint

	// Private variables.
	waitIndex                  int
	currentUserId              string
	currentUserCreatedTime     time.Time
	currentUserOccurrenceCount uint
	currentEventTimes          []time.Time
	currentRepeats             []uint
}

const num_T_BINS = 100
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
		Repeats:                    make([]Hist.NumericHistogram, pLen),
		Count:                      0,
		OncePerUserCount:           0,
		EventCount:                 0,
		waitIndex:                  0,
		currentUserId:              "",
		currentUserCreatedTime:     time.Time{},
		currentUserOccurrenceCount: 0,
		currentEventTimes:          make([]time.Time, pLen),
		currentRepeats:             make([]uint, pLen),
	}
	for i := 0; i < pLen; i++ {
		pattern.Timings[i] = *Hist.NewHistogram(num_T_BINS)
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

func (p *Pattern) CountForEvent(eventName string, eventCreatedTime time.Time, userId string, userCreatedTime time.Time) error {
	if eventName == "" || eventCreatedTime.Equal(time.Time{}) {
		return fmt.Errorf("Missing eventId or eventCreatedTime.")
	}

	if userId != p.currentUserId || !p.currentUserCreatedTime.Equal(userCreatedTime) {
		return fmt.Errorf("Mismatch in User data.")
	}

	p.EventCount += 1

	if p.waitIndex > 0 && eventName == p.EventNames[p.waitIndex-1] {
		// Repeats count the number of times the current event has occurred
		// before seeing the next event being waited upon.
		p.currentRepeats[p.waitIndex-1] += 1
	} else if eventName == p.EventNames[p.waitIndex] {
		// Record the event occurrence and wait on the next one.
		p.currentEventTimes[p.waitIndex] = eventCreatedTime
		p.currentRepeats[p.waitIndex] = 1

		p.waitIndex += 1

		pLen := len(p.EventNames)
		if p.waitIndex == pLen {
			// Record the pattern occurrence.
			p.Count += 1
			p.currentUserOccurrenceCount += 1
			if p.currentUserOccurrenceCount == 1 {
				p.OncePerUserCount += 1
			}

			// Update histograms of timings.
			var duration float64
			for i := 0; i < pLen; i++ {
				if i == 0 {
					duration = p.currentEventTimes[0].Sub(userCreatedTime).Seconds()
				} else {
					duration = p.currentEventTimes[i].Sub(p.currentEventTimes[i-1]).Seconds()
				}
				if duration < 0 {
					return fmt.Errorf("Event Timings not in order")
				}
				p.Timings[i].Add(duration)
			}

			// Update histograms of repeats.
			for i := 0; i < pLen; i++ {
				p.Repeats[i].Add(float64(p.currentRepeats[i]))
			}

			// Reset.
			p.currentEventTimes = make([]time.Time, pLen)
			p.currentRepeats = make([]uint, pLen)
			p.waitIndex = 0
		}
	}

	return nil
}
