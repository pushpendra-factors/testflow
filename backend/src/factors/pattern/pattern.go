package pattern

import (
	Hist "factors/histogram"
	U "factors/util"
	"fmt"
	"math"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Pattern struct {
	EventNames []string `json:"en"`
	// Histograms.
	CardinalityRepeatTimings   *Hist.NumericHistogramStruct     `json:"crt"`
	EventNumericProperties     *Hist.NumericHistogramStruct     `json:"enp"`
	EventCategoricalProperties *Hist.CategoricalHistogramStruct `json:"ecp"`
	UserNumericProperties      *Hist.NumericHistogramStruct     `json:"unp"`
	UserCategoricalProperties  *Hist.CategoricalHistogramStruct `json:"ucp"`
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
	currentUserJoinTimestamp   int64
	currentUserOccurrenceCount uint
	currentEventTimestamps     []int64
	currentEventCardinalities  []uint
	currentRepeats             []uint
	currentEPropertiesNMap     map[string]float64
	currentEPropertiesCMap     map[string]string
	currentUPropertiesNMap     map[string]float64
	currentUPropertiesCMap     map[string]string
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
	EPNumericConstraints     []NumericConstraint
	EPCategoricalConstraints []CategoricalConstraint
	UPNumericConstraints     []NumericConstraint
	UPCategoricalConstraints []CategoricalConstraint
}

func NewPattern(events []string, userAndEventsInfo *UserAndEventsInfo) (*Pattern, error) {
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
		CardinalityRepeatTimings:   defaultHist,
		EventNumericProperties:     nil,
		EventCategoricalProperties: nil,
		UserNumericProperties:      nil,
		UserCategoricalProperties:  nil,
		Count:                      0,
		OncePerUserCount:           0,
		waitIndex:                  0,
		currentUserId:              "",
		currentUserJoinTimestamp:   0,
		currentUserOccurrenceCount: 0,
		currentEventTimestamps:     make([]int64, pLen),
		currentEventCardinalities:  make([]uint, pLen),
		currentRepeats:             make([]uint, pLen),
		currentEPropertiesNMap:     make(map[string]float64),
		currentEPropertiesCMap:     make(map[string]string),
		currentUPropertiesNMap:     make(map[string]float64),
		currentUPropertiesCMap:     make(map[string]string),
	}
	if userAndEventsInfo == nil {
		return &pattern, nil
	}
	userPropertiesNTemplate, userPropertiesCTemplate, eventPropertiesNTemplate, eventPropertiesCTemplate, err :=
		buildPropertiesHistogramTemplates(events, userAndEventsInfo)
	if err != nil {
		return nil, err
	}
	// Setup Histograms.
	if userPropertiesNTemplate != nil {
		nDimensions := len(*userPropertiesNTemplate)
		nBinsFloat := math.Min(float64(nDimensions*num_NUMERIC_BINS_PER_DIMENSION),
			float64(num_MAX_NUMERIC_MULTI_BINS))
		nBins := int(math.Max(1.0, nBinsFloat))
		nHist, err := Hist.NewNumericHistogram(nBins, nDimensions, userPropertiesNTemplate)
		if err != nil {
			return nil, err
		}
		pattern.UserNumericProperties = nHist
	}

	if userPropertiesCTemplate != nil {
		cDimensions := len(*userPropertiesCTemplate)
		cBinsFloat := math.Min(float64(cDimensions*num_CATEGORICAL_BINS_PER_DIMENSION),
			float64(num_MAX_CATEGORICAL_MULTI_BINS))
		cBins := int(math.Max(1.0, cBinsFloat))
		cHist, err := Hist.NewCategoricalHistogram(cBins, cDimensions, userPropertiesCTemplate)
		if err != nil {
			return nil, err
		}
		pattern.UserCategoricalProperties = cHist
	}

	if eventPropertiesNTemplate != nil {
		nDimensions := len(*eventPropertiesNTemplate)
		nBinsFloat := math.Min(float64(nDimensions*num_NUMERIC_BINS_PER_DIMENSION),
			float64(num_MAX_NUMERIC_MULTI_BINS))
		nBins := int(math.Max(1.0, nBinsFloat))
		nHist, err := Hist.NewNumericHistogram(nBins, nDimensions, eventPropertiesNTemplate)
		if err != nil {
			return nil, err
		}
		pattern.EventNumericProperties = nHist
	}

	if eventPropertiesCTemplate != nil {
		cDimensions := len(*eventPropertiesCTemplate)
		cBinsFloat := math.Min(float64(cDimensions*num_CATEGORICAL_BINS_PER_DIMENSION),
			float64(num_MAX_CATEGORICAL_MULTI_BINS))
		cBins := int(math.Max(1.0, cBinsFloat))
		cHist, err := Hist.NewCategoricalHistogram(cBins, cDimensions, eventPropertiesCTemplate)
		if err != nil {
			return nil, err
		}
		pattern.EventCategoricalProperties = cHist
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

func buildPropertiesHistogramTemplates(
	events []string, userAndEventsInfo *UserAndEventsInfo) (
	*Hist.NumericHistogramTemplate, *Hist.CategoricalHistogramTemplate,
	*Hist.NumericHistogramTemplate, *Hist.CategoricalHistogramTemplate,
	error) {
	userPropertiesNTemplate := Hist.NumericHistogramTemplate{}
	userPropertiesCTemplate := Hist.CategoricalHistogramTemplate{}
	eventPropertiesNTemplate := Hist.NumericHistogramTemplate{}
	eventPropertiesCTemplate := Hist.CategoricalHistogramTemplate{}
	if userAndEventsInfo == nil {
		return &userPropertiesNTemplate, &userPropertiesCTemplate,
			&eventPropertiesNTemplate, &eventPropertiesCTemplate, nil
	}

	pLen := len(events)
	userPropertiesInfo := userAndEventsInfo.UserPropertiesInfo
	if userPropertiesInfo != nil {
		for i := 0; i < pLen; i++ {
			for propertyName, _ := range userAndEventsInfo.UserPropertiesInfo.NumericPropertyKeys {
				// User properties is tracked at each event level.
				nptu := Hist.NumericHistogramTemplateUnit{
					Name:       PatternPropertyKey(i, propertyName),
					IsRequired: false,
					Default:    0.0,
				}
				userPropertiesNTemplate = append(userPropertiesNTemplate, nptu)
			}

			for propertyName, _ := range userAndEventsInfo.UserPropertiesInfo.CategoricalPropertyKeyValues {
				// User properties is tracked at each event level.
				cptu := Hist.CategoricalHistogramTemplateUnit{
					Name:       PatternPropertyKey(i, propertyName),
					IsRequired: false,
					Default:    "",
				}
				userPropertiesCTemplate = append(userPropertiesCTemplate, cptu)
			}
		}
	}

	eventInfoMap := userAndEventsInfo.EventPropertiesInfoMap
	if eventInfoMap != nil {
		for i := 0; i < pLen; i++ {
			if eventInfo, ok := (*eventInfoMap)[events[i]]; ok {
				for propertyName, _ := range eventInfo.NumericPropertyKeys {
					// Event properties of corresponding event.
					nptu := Hist.NumericHistogramTemplateUnit{
						Name:       PatternPropertyKey(i, propertyName),
						IsRequired: false,
						Default:    0.0,
					}
					eventPropertiesNTemplate = append(eventPropertiesNTemplate, nptu)
				}

				for propertyName, _ := range eventInfo.CategoricalPropertyKeyValues {
					// User properties is tracked at each event level.
					cptu := Hist.CategoricalHistogramTemplateUnit{
						Name:       PatternPropertyKey(i, propertyName),
						IsRequired: false,
						Default:    "",
					}
					eventPropertiesCTemplate = append(eventPropertiesCTemplate, cptu)
				}
			} else {
				log.Error(fmt.Sprintf(
					"Missing info for event %s in pattern %s. Not building event histogram templates.",
					events[i], events))
				return &userPropertiesNTemplate, &userPropertiesCTemplate, nil, nil, nil
			}
		}
	}
	return &userPropertiesNTemplate, &userPropertiesCTemplate,
		&eventPropertiesNTemplate, &eventPropertiesCTemplate, nil
}

func (p *Pattern) ResetForNewUser(userId string, userJoinTimestamp int64) error {
	if userId == "" || userJoinTimestamp <= 0 {
		return fmt.Errorf("Missing userId or userCreatedTime.")
	}

	p.UserCount += 1
	// Reinitialize all private variables maintained per user event stream.
	p.waitIndex = 0
	p.currentUserId = userId
	p.currentUserJoinTimestamp = userJoinTimestamp
	p.currentUserOccurrenceCount = 0
	pLen := len(p.EventNames)
	p.currentEventTimestamps = make([]int64, pLen)
	p.currentRepeats = make([]uint, pLen)
	return nil
}

func addNumericAndCategoricalProperties(
	eventIndex int, properties map[string]interface{},
	nMap map[string]float64, cMap map[string]string) {

	for key, value := range properties {
		if numericValue, ok := value.(float64); ok {
			nMap[PatternPropertyKey(eventIndex, key)] = numericValue
		} else if categoricalValue, ok := value.(string); ok {
			cMap[PatternPropertyKey(eventIndex, key)] = categoricalValue
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
	eventName string, eventTimestamp int64, eventProperties map[string]interface{},
	userProperties map[string]interface{}, eventCardinality uint, userId string,
	userJoinTimestamp int64) (string, error) {

	if eventName == "" || eventTimestamp <= 0 {
		return "", fmt.Errorf("Missing eventId or eventTimestamp.")
	}

	if userId != p.currentUserId || p.currentUserJoinTimestamp != userJoinTimestamp {
		return "", fmt.Errorf(
			fmt.Sprintf("Mismatch in User data. userId: %s, userJoinTime: %v, pattern userId: %s, pattern userJoinTime: %d",
				userId, userJoinTimestamp, p.currentUserId, p.currentUserJoinTimestamp))
	}

	if p.waitIndex > 0 && eventName == p.EventNames[p.waitIndex-1] {
		// Repeats count the number of times the current event has occurred
		// before seeing the next event being waited upon.
		p.currentRepeats[p.waitIndex-1] += 1
	} else if eventName == p.EventNames[p.waitIndex] {
		// Record the event occurrence and wait on the next one.
		p.currentEventTimestamps[p.waitIndex] = eventTimestamp
		p.currentEventCardinalities[p.waitIndex] = eventCardinality
		p.currentRepeats[p.waitIndex] = 1
		// Update seen properties.
		addNumericAndCategoricalProperties(
			p.waitIndex, eventProperties, p.currentEPropertiesNMap, p.currentEPropertiesCMap)
		addNumericAndCategoricalProperties(
			p.waitIndex, userProperties, p.currentUPropertiesNMap, p.currentUPropertiesCMap)

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
						duration := p.currentEventTimestamps[0] - userJoinTimestamp
						if duration < 0 {
							// Ignoring this error for now, since there are no DB checks to avoid
							// these user input values.
							log.Debug(fmt.Sprintf("Event occurs before creation for user:%s", p.currentUserId))
						}
					} else {
						duration := p.currentEventTimestamps[i] - p.currentEventTimestamps[i-1]
						if duration < 0 {
							return "", fmt.Errorf("Event Timings not in order")
						}
					}
				}
				// Update properties histograms.
				if p.EventNumericProperties != nil {
					if err := p.EventNumericProperties.AddMap(p.currentEPropertiesNMap); err != nil {
						return "", err
					}
				}
				if p.EventCategoricalProperties != nil {
					if err := p.EventCategoricalProperties.AddMap(p.currentEPropertiesCMap); err != nil {
						return "", err
					}
				}
				if p.UserNumericProperties != nil {
					if err := p.UserNumericProperties.AddMap(p.currentUPropertiesNMap); err != nil {
						return "", err
					}
				}
				if p.UserCategoricalProperties != nil {
					if err := p.UserCategoricalProperties.AddMap(p.currentUPropertiesCMap); err != nil {
						return "", err
					}
				}
				// Update multi histogram of cardinalities, repeats and timings.
				var cardinalityRepeatTimingsVec []float64 = make([]float64, 6)
				cardinalityRepeatTimingsVec[0] = float64(p.currentEventCardinalities[0])
				cardinalityRepeatTimingsVec[1] = float64(p.currentRepeats[0])
				cardinalityRepeatTimingsVec[2] = math.Max(float64(p.currentEventTimestamps[0]-userJoinTimestamp), 0)
				cardinalityRepeatTimingsVec[3] = float64(p.currentEventCardinalities[pLen-1])
				cardinalityRepeatTimingsVec[4] = float64(p.currentRepeats[pLen-1])
				if pLen > 1 {
					cardinalityRepeatTimingsVec[5] = float64(p.currentEventTimestamps[pLen-1] - p.currentEventTimestamps[pLen-2])
				} else {
					cardinalityRepeatTimingsVec[5] = cardinalityRepeatTimingsVec[2]
				}
				p.CardinalityRepeatTimings.Add(cardinalityRepeatTimingsVec)
			}

			// Reset.
			p.currentEventTimestamps = make([]int64, pLen)
			p.currentEventCardinalities = make([]uint, pLen)
			p.currentRepeats = make([]uint, pLen)
			p.waitIndex = 0
			p.currentEPropertiesNMap = make(map[string]float64)
			p.currentEPropertiesCMap = make(map[string]string)
			p.currentUPropertiesNMap = make(map[string]float64)
			p.currentUPropertiesCMap = make(map[string]string)
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
	EPNMapUpperBounds := make(map[string]float64)
	EPNMapLowerBounds := make(map[string]float64)
	EPCMapEquality := make(map[string]string)
	UPNMapUpperBounds := make(map[string]float64)
	UPNMapLowerBounds := make(map[string]float64)
	UPCMapEquality := make(map[string]string)
	hasCrtConstraints := false
	for i, ecs := range patternConstraints {
		eventName := p.EventNames[i]
		for _, ncs := range ecs.EPNumericConstraints {
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
				key := PatternPropertyKey(i, ncs.PropertyName)
				EPNMapLowerBounds[key] = ncs.LowerBound
				EPNMapUpperBounds[key] = ncs.UpperBound
			}
		}
		for _, ccs := range ecs.EPCategoricalConstraints {
			key := PatternPropertyKey(i, ccs.PropertyName)
			EPCMapEquality[key] = ccs.PropertyValue
		}

		for _, ncs := range ecs.UPNumericConstraints {
			key := PatternPropertyKey(i, ncs.PropertyName)
			UPNMapLowerBounds[key] = ncs.LowerBound
			UPNMapUpperBounds[key] = ncs.UpperBound
		}
		for _, ccs := range ecs.UPCategoricalConstraints {
			key := PatternPropertyKey(i, ccs.PropertyName)
			UPCMapEquality[key] = ccs.PropertyValue
		}
	}

	crtUpperCDF := 1.0
	crtLowerCDF := 0.0
	if hasCrtConstraints {
		crtUpperCDF = p.CardinalityRepeatTimings.CDF(crtUpperBounds)
		crtLowerCDF = p.CardinalityRepeatTimings.CDF(crtLowerBounds)
	}

	EPNumericUpperCDF := 1.0
	EPNumericLowerCDF := 0.0
	if p.EventNumericProperties != nil && len(EPNMapLowerBounds) > 0 {
		EPNumericUpperCDF = p.EventNumericProperties.CDFFromMap(EPNMapUpperBounds)
		EPNumericLowerCDF = p.EventNumericProperties.CDFFromMap(EPNMapLowerBounds)
	}
	EPCategoricalPDF := 1.0
	if p.EventCategoricalProperties != nil && len(EPCMapEquality) > 0 {
		var err error
		EPCategoricalPDF, err = p.EventCategoricalProperties.PDFFromMap(EPCMapEquality)
		if err != nil {
			return 0, err
		}
	}

	UPNumericUpperCDF := 1.0
	UPNumericLowerCDF := 0.0
	if p.UserNumericProperties != nil && len(UPNMapLowerBounds) > 0 {
		UPNumericUpperCDF = p.UserNumericProperties.CDFFromMap(UPNMapUpperBounds)
		UPNumericLowerCDF = p.UserNumericProperties.CDFFromMap(UPNMapLowerBounds)
	}
	UPCategoricalPDF := 1.0
	if p.UserCategoricalProperties != nil && len(UPCMapEquality) > 0 {
		var err error
		UPCategoricalPDF, err = p.UserCategoricalProperties.PDFFromMap(UPCMapEquality)
		if err != nil {
			return 0, err
		}
	}

	count := (float64(p.OncePerUserCount) *
		(crtUpperCDF - crtLowerCDF) *
		(EPNumericUpperCDF - EPNumericLowerCDF) *
		EPCategoricalPDF *
		(UPNumericUpperCDF - UPNumericLowerCDF) *
		UPCategoricalPDF)

	if count < 0 {
		log.WithFields(log.Fields{
			"crtUpperCDF":        crtUpperCDF,
			"crtUpperBounds":     crtUpperBounds,
			"crtLowerCDF":        crtLowerCDF,
			"crtLowerBounds":     crtLowerBounds,
			"EPNumericUpperCDF":  EPNumericUpperCDF,
			"EPNumericLowerCDF":  EPNumericLowerCDF,
			"EPCategoricalPDF":   EPCategoricalPDF,
			"UPNumericUpperCDF":  UPNumericUpperCDF,
			"UPNumericLowerCDF":  UPNumericLowerCDF,
			"UPCategoricalPDF":   UPCategoricalPDF,
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

func (p *Pattern) GetEventPropertyRanges(
	eventIndex int, propertyName string) [][2]float64 {
	// Return the ranges of the bin [min, max], in which the numeric values for the event property occurr.
	return p.EventNumericProperties.GetBinRanges(PatternPropertyKey(eventIndex, propertyName))
}

func (p *Pattern) GetUserPropertyRanges(
	eventIndex int, propertyName string) [][2]float64 {
	// Return the ranges of the bin [min, max], in which the numeric values for the event property occurr.
	return p.UserNumericProperties.GetBinRanges(PatternPropertyKey(eventIndex, propertyName))
}

func (p *Pattern) String() string {
	return EventArrayToString(p.EventNames)
}

func EventArrayToString(eventNames []string) string {
	return strings.Join(eventNames, ",")
}
