package pattern

import (
	Hist "factors/histogram"
	U "factors/util"
	"fmt"
	"math"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Pattern struct {
	EventNames []string `json:"en"`
	// Histograms.
	CardinalityRepeatTimings *Hist.NumericHistogramStruct     `json:"crt"`
	NumericProperties        *Hist.NumericHistogramStruct     `json:"np"`
	CategoricalProperties    *Hist.CategoricalHistogramStruct `json:"cp"`
	// The total number of times this pattern occurs allowing multiple counts
	// per user.
	Count uint `json:"c"`
	// Counted only once per user.
	OncePerUserCount uint `json:"ouc"`
	// Number of users the pattern was counted on.
	UserCount uint `json:"uc"`

	// Private variables.
	waitIndex                  int
	currentUserId              string
	currentUserCreatedTime     time.Time
	currentUserOccurrenceCount uint
	currentEventTimes          []time.Time
	currentEventCardinalities  []uint
	currentRepeats             []uint
	currentNMap                map[string]float64
	currentCMap                map[string]string
}

const num_T_BINS = 20
const num_C_BINS = 10
const num_R_BINS = 10
const num_DEFAULT_MULTI_BINS = 64
const num_NUMERIC_BINS_PER_DIMENSION = 3
const num_MAX_NUMERIC_MULTI_BINS = 128
const num_CATEGORICAL_BINS_PER_DIMENSION = 1
const num_MAX_CATEGORICAL_MULTI_BINS = 6

type NumericConstraint struct {
	PropertyName string
	LowerBound   float64
	UpperBound   float64
	IsEquality   bool
}
type CategoricalConstraint struct {
	PropertyName  string
	PropertyValue string
}
type EventConstraints struct {
	NumericConstraints     []NumericConstraint
	CategoricalConstraints []CategoricalConstraint
}

func NewPattern(events []string, eventInfoMap *EventInfoMap) (*Pattern, error) {
	pLen := len(events)
	if pLen == 0 {
		err := fmt.Errorf("No events in pattern")
		return nil, err
	}
	if !isEventsUnique(events) {
		err := fmt.Errorf("Events are not unique")
		return nil, err
	}
	defaultHist, err := Hist.NewNumericHistogram(num_DEFAULT_MULTI_BINS, 6, nil)
	if err != nil {
		return nil, err
	}
	pattern := Pattern{
		EventNames: events,
		// 6 dimensional histogram - Cardinalties, Repeats, Timings of start_event
		// and last_event.
		CardinalityRepeatTimings: defaultHist,
		NumericProperties:        nil,
		CategoricalProperties:    nil,
		Count:                      0,
		OncePerUserCount:           0,
		waitIndex:                  0,
		currentUserId:              "",
		currentUserCreatedTime:     time.Time{},
		currentUserOccurrenceCount: 0,
		currentEventTimes:          make([]time.Time, pLen),
		currentEventCardinalities:  make([]uint, pLen),
		currentRepeats:             make([]uint, pLen),
		currentNMap:                make(map[string]float64),
		currentCMap:                make(map[string]string),
	}
	if eventInfoMap != nil {
		nTemplate := Hist.NumericHistogramTemplate{}
		for i := 0; i < pLen; i++ {
			if eventInfo, ok := (*eventInfoMap)[events[i]]; ok {
				nTemplate = append(nTemplate, *eventInfo.NumericPropertiesTemplate...)
			} else {
				return nil, fmt.Errorf(fmt.Sprintf(
					"Missing info for event %s", events[i]))
			}
		}
		nDimensions := len(nTemplate)
		nBinsFloat := math.Min(float64(nDimensions*num_NUMERIC_BINS_PER_DIMENSION),
			float64(num_MAX_NUMERIC_MULTI_BINS))
		nBins := int(math.Max(1.0, nBinsFloat))
		nHist, err := Hist.NewNumericHistogram(nBins, nDimensions, &nTemplate)
		if err != nil {
			return nil, err
		}
		pattern.NumericProperties = nHist

		cTemplate := Hist.CategoricalHistogramTemplate{}
		for i := 0; i < pLen; i++ {
			if eventInfo, ok := (*eventInfoMap)[events[i]]; ok {
				cTemplate = append(cTemplate, *eventInfo.CategoricalPropertiesTemplate...)
			} else {
				return nil, fmt.Errorf(fmt.Sprintf(
					"Missing info for event %s", events[i]))
			}
		}
		cDimensions := len(cTemplate)
		cBinsFloat := math.Min(float64(cDimensions*num_CATEGORICAL_BINS_PER_DIMENSION),
			float64(num_MAX_CATEGORICAL_MULTI_BINS))
		cBins := int(math.Max(1.0, cBinsFloat))
		cHist, err := Hist.NewCategoricalHistogram(cBins, cDimensions, &cTemplate)
		if err != nil {
			return nil, err
		}
		pattern.CategoricalProperties = cHist
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

func addNumericAndCategoricalProperties(
	eventName string, eventProperties map[string]interface{},
	nMap map[string]float64, cMap map[string]string) {

	for key, value := range eventProperties {
		if numericValue, ok := value.(float64); ok {
			nMap[EventPropertyKey(eventName, key)] = numericValue
		} else if categoricalValue, ok := value.(string); ok {
			cMap[EventPropertyKey(eventName, key)] = categoricalValue
		}
	}
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
func (p *Pattern) CountForEvent(
	eventName string, eventCreatedTime time.Time, eventProperties map[string]interface{},
	eventCardinality uint, userId string, userCreatedTime time.Time) (string, error) {

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
		// Update seen properties.
		addNumericAndCategoricalProperties(eventName, eventProperties, p.currentNMap, p.currentCMap)

		p.waitIndex += 1

		pLen := len(p.EventNames)
		if p.waitIndex == pLen {
			// Record the pattern occurrence.
			p.Count += 1
			p.currentUserOccurrenceCount += 1
			if p.currentUserOccurrenceCount == 1 {
				p.OncePerUserCount += 1

				// Check whether events are in order.
				for i := 0; i < pLen; i++ {
					if i == 0 {
						duration := p.currentEventTimes[0].Sub(userCreatedTime).Seconds()
						if duration < 0 {
							// Ignoring this error for now, since there are no DB checks to avoid
							// these user input values.
							log.Error(fmt.Sprintf("Event occurs before creation for user:%s", p.currentUserId))
						}
					} else {
						duration := p.currentEventTimes[i].Sub(p.currentEventTimes[i-1]).Seconds()
						if duration < 0 {
							return "", fmt.Errorf("Event Timings not in order")
						}
					}
				}
				// Update properties histograms.
				if p.NumericProperties != nil {
					if err := p.NumericProperties.AddMap(p.currentNMap); err != nil {
						return "", err
					}
				}
				// Update properties histograms.
				if p.CategoricalProperties != nil {
					if err := p.CategoricalProperties.AddMap(p.currentCMap); err != nil {
						return "", err
					}
				}
				// Update multi histogram of cardinalities, repeats and timings.
				var cardinalityRepeatTimingsVec []float64 = make([]float64, 6)
				cardinalityRepeatTimingsVec[0] = float64(p.currentEventCardinalities[0])
				cardinalityRepeatTimingsVec[1] = float64(p.currentRepeats[0])
				cardinalityRepeatTimingsVec[2] = p.currentEventTimes[0].Sub(userCreatedTime).Seconds()
				cardinalityRepeatTimingsVec[3] = float64(p.currentEventCardinalities[pLen-1])
				cardinalityRepeatTimingsVec[4] = float64(p.currentRepeats[pLen-1])
				if pLen > 1 {
					cardinalityRepeatTimingsVec[5] = p.currentEventTimes[pLen-1].Sub(p.currentEventTimes[pLen-2]).Seconds()
				} else {
					cardinalityRepeatTimingsVec[5] = cardinalityRepeatTimingsVec[2]
				}
				p.CardinalityRepeatTimings.Add(cardinalityRepeatTimingsVec)
			}

			// Reset.
			p.currentEventTimes = make([]time.Time, pLen)
			p.currentEventCardinalities = make([]uint, pLen)
			p.currentRepeats = make([]uint, pLen)
			p.waitIndex = 0
			p.currentNMap = make(map[string]float64)
			p.currentCMap = make(map[string]string)
		}
	}
	return p.EventNames[p.waitIndex], nil
}

func (p *Pattern) GetOncePerUserCount(
	patternConstraints []EventConstraints) (uint, error) {
	pLen := len(p.EventNames)
	if patternConstraints != nil && (len(patternConstraints) != pLen) {
		errorString := fmt.Sprintf(
			"Constraint length %d does not match pattern length %d for pattern %v",
			len(patternConstraints), pLen, p.EventNames)
		log.Error(errorString)
		return 0, fmt.Errorf(errorString)
	}
	crtLowerBounds := make([]float64, 6)
	crtUpperBounds := make([]float64, 6)
	for i := 0; i < 6; i++ {
		crtLowerBounds[i] = math.MaxFloat64
		crtUpperBounds[i] = math.MaxFloat64
	}
	nMapUpperBounds := make(map[string]float64)
	nMapLowerBounds := make(map[string]float64)
	cMapEquality := make(map[string]string)
	hasCrtConstraints := false
	for i, ecs := range patternConstraints {
		eventName := p.EventNames[i]
		for _, ncs := range ecs.NumericConstraints {
			if ncs.PropertyName == U.EP_OCCURRENCE_COUNT {
				hasCrtConstraints = true
				if i == 0 {
					crtLowerBounds[0] = ncs.LowerBound
					crtUpperBounds[0] = ncs.UpperBound
				} else if i == pLen-1 {
					crtLowerBounds[3] = ncs.LowerBound
					crtUpperBounds[3] = ncs.UpperBound
				} else {
					errorString := fmt.Sprintf(
						"Cardinality is not maintained for event %v in pattern %s", eventName, p.EventNames)
					return 0, fmt.Errorf(errorString)
				}
			} else {
				key := EventPropertyKey(eventName, ncs.PropertyName)
				nMapLowerBounds[key] = ncs.LowerBound
				nMapUpperBounds[key] = ncs.UpperBound
			}
		}
		for _, ccs := range ecs.CategoricalConstraints {
			key := EventPropertyKey(eventName, ccs.PropertyName)
			cMapEquality[key] = ccs.PropertyValue
		}
	}

	crtUpperCDF := 1.0
	crtLowerCDF := 0.0
	if hasCrtConstraints {
		crtUpperCDF = p.CardinalityRepeatTimings.CDF(crtUpperBounds)
		crtLowerCDF = p.CardinalityRepeatTimings.CDF(crtLowerBounds)
	}

	numericUpperCDF := 1.0
	numericLowerCDF := 0.0
	if p.NumericProperties != nil && len(nMapLowerBounds) > 0 {
		numericUpperCDF = p.NumericProperties.CDFFromMap(nMapUpperBounds)
		numericLowerCDF = p.NumericProperties.CDFFromMap(nMapLowerBounds)
	}
	categoricalPDF := 1.0
	if p.CategoricalProperties != nil && len(cMapEquality) > 0 {
		var err error
		categoricalPDF, err = p.CategoricalProperties.PDFFromMap(cMapEquality)
		if err != nil {
			return 0, err
		}
	}
	count := (float64(p.OncePerUserCount) *
		(crtUpperCDF - crtLowerCDF) *
		(numericUpperCDF - numericLowerCDF) *
		categoricalPDF)

	if count < 0 {
		log.WithFields(log.Fields{
			"crtUpperCDF":        crtUpperCDF,
			"crtUpperBounds":     crtUpperBounds,
			"crtLowerCDF":        crtLowerCDF,
			"crtLowerBounds":     crtLowerBounds,
			"numericUpperCDF":    numericUpperCDF,
			"numericLowerCDF":    numericLowerCDF,
			"categoricalPDF":     categoricalPDF,
			"pattern":            p.String(),
			"patternConstraints": patternConstraints,
			"patternCount":       p.OncePerUserCount,
			"finalCount":         count,
		}).Info("Computed CDF's and PDF's")
		errorString := "Final count is less than 0."
		log.Error(errorString)
		return 0, fmt.Errorf(errorString)
	}
	return uint(count), nil
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
