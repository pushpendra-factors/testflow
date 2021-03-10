package pattern

import (
	Hist "factors/histogram"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"math"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Pattern struct {
	EventNames []string `json:"en"`
	// Histograms.
	GenericPropertiesHistogram              *Hist.NumericHistogramStruct     `json:gph`
	PerUserEventNumericProperties           *Hist.NumericHistogramStruct     `json:"enp"`
	PerUserEventCategoricalProperties       *Hist.CategoricalHistogramStruct `json:"ecp"`
	PerUserUserNumericProperties            *Hist.NumericHistogramStruct     `json:"unp"`
	PerUserUserCategoricalProperties        *Hist.CategoricalHistogramStruct `json:"ucp"`
	PerOccurrenceEventNumericProperties     *Hist.NumericHistogramStruct     `json:"oenp"`
	PerOccurrenceEventCategoricalProperties *Hist.CategoricalHistogramStruct `json:"oecp"`
	PerOccurrenceUserNumericProperties      *Hist.NumericHistogramStruct     `json:"ounp"`
	PerOccurrenceUserCategoricalProperties  *Hist.CategoricalHistogramStruct `json:"oucp"`
	// The total number of times this pattern occurs allowing multiple counts
	// per user.
	PerOccurrenceCount uint `json:"c"`
	// Counted only once per user.
	PerUserCount uint `json:"ouc"`
	// Number of users the pattern was counted on.
	TotalUserCount uint `json:"uc"`

	// Private variables.
	waitIndex                       int
	currentUserId                   string
	currentUserJoinTimestamp        int64
	previousEventTimestamp          int64
	currentUserOccurrenceCount      uint
	currentUserEventTimestamps      map[string][]int64
	currentUserEventOccurenceCounts map[string][]uint

	// These are tracked by default for first seen events.
	currentEPropertiesNMap map[string]float64
	currentEPropertiesCMap map[string]string
	currentUPropertiesNMap map[string]float64
	currentUPropertiesCMap map[string]string
}

const num_T_BINS = 20
const num_C_BINS = 10
const num_R_BINS = 10
const num_DEFAULT_MULTI_BINS = 64
const num_NUMERIC_BINS_PER_DIMENSION = 9
const num_MAX_NUMERIC_MULTI_BINS = 128
const num_CATEGORICAL_BINS_PER_DIMENSION = 1
const num_MAX_CATEGORICAL_MULTI_BINS = 6

// 20 MB.
const MAX_PATTERN_BYTES = 20 * 1024 * 1024

const COUNT_TYPE_PER_USER = "ct_per_user"
const COUNT_TYPE_PER_OCCURRENCE = "ct_per_occurrence"
const EQUALS_OPERATOR_CONST = "Equals"
const NOT_EQUALS_OPERATOR_CONST = "NotEquals"

type NumericConstraint struct {
	PropertyName string
	LowerBound   float64
	UpperBound   float64
	IsEquality   bool
	UseBound     string
}
type CategoricalConstraint struct {
	PropertyName  string
	PropertyValue string
	Operator      string
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

	defaultHistTemplate := Hist.NumericHistogramTemplate{}
	for _, propertyName := range U.GENERIC_NUMERIC_USER_PROPERTIES {
		nptu := Hist.NumericHistogramTemplateUnit{
			Name:       propertyName,
			IsRequired: false,
			Default:    0.0,
		}
		defaultHistTemplate = append(defaultHistTemplate, nptu)
	}
	for i, _ := range events {
		for _, propertyName := range U.GENERIC_NUMERIC_EVENT_PROPERTIES {
			// User properties is tracked at each event level.
			nptu := Hist.NumericHistogramTemplateUnit{
				Name:       PatternPropertyKey(i, propertyName),
				IsRequired: false,
				Default:    0.0,
			}
			defaultHistTemplate = append(defaultHistTemplate, nptu)
		}
	}

	defaultHist, err := Hist.NewNumericHistogram(
		num_DEFAULT_MULTI_BINS, len(defaultHistTemplate), &defaultHistTemplate)
	if err != nil {
		return nil, err
	}
	pattern := Pattern{
		EventNames: events,
		// 6 dimensional histogram - Cardinalties, Repeats, Timings of start_event
		// and last_event.
		GenericPropertiesHistogram:              defaultHist,
		PerUserEventNumericProperties:           nil,
		PerUserEventCategoricalProperties:       nil,
		PerUserUserNumericProperties:            nil,
		PerUserUserCategoricalProperties:        nil,
		PerOccurrenceEventNumericProperties:     nil,
		PerOccurrenceEventCategoricalProperties: nil,
		PerOccurrenceUserNumericProperties:      nil,
		PerOccurrenceUserCategoricalProperties:  nil,
		PerOccurrenceCount:                      0,
		PerUserCount:                            0,
		TotalUserCount:                          0,
		waitIndex:                               0,
		currentUserId:                           "",
		currentUserJoinTimestamp:                0,
		previousEventTimestamp:                  0,
		currentUserEventTimestamps:              make(map[string][]int64),
		currentUserEventOccurenceCounts:         make(map[string][]uint),
		currentUserOccurrenceCount:              0,
		currentEPropertiesNMap:                  make(map[string]float64),
		currentEPropertiesCMap:                  make(map[string]string),
		currentUPropertiesNMap:                  make(map[string]float64),
		currentUPropertiesCMap:                  make(map[string]string),
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
		// Restrict between NHIST_MIN_BIN_SIZE to num_MAX_NUMERIC_MULTI_BINS.
		nBinsFloat := math.Min(
			math.Max(
				float64(nDimensions*num_NUMERIC_BINS_PER_DIMENSION),
				float64(Hist.NHIST_MIN_BIN_SIZE)),
			float64(num_MAX_NUMERIC_MULTI_BINS))
		nBins := int(math.Max(1.0, nBinsFloat))
		puNHist, err := Hist.NewNumericHistogram(nBins, nDimensions, userPropertiesNTemplate)
		if err != nil {
			return nil, err
		}
		pattern.PerUserUserNumericProperties = puNHist
		poNHist, err := Hist.NewNumericHistogram(nBins, nDimensions, userPropertiesNTemplate)
		if err != nil {
			return nil, err
		}
		pattern.PerOccurrenceUserNumericProperties = poNHist
	}

	if userPropertiesCTemplate != nil {
		cDimensions := len(*userPropertiesCTemplate)
		cBinsFloat := math.Min(float64(cDimensions*num_CATEGORICAL_BINS_PER_DIMENSION),
			float64(num_MAX_CATEGORICAL_MULTI_BINS))
		cBins := int(math.Max(1.0, cBinsFloat))
		puCHist, err := Hist.NewCategoricalHistogram(cBins, cDimensions, userPropertiesCTemplate)
		if err != nil {
			return nil, err
		}
		pattern.PerUserUserCategoricalProperties = puCHist
		poCHist, err := Hist.NewCategoricalHistogram(cBins, cDimensions, userPropertiesCTemplate)
		if err != nil {
			return nil, err
		}
		pattern.PerOccurrenceUserCategoricalProperties = poCHist
	}

	if eventPropertiesNTemplate != nil {
		nDimensions := len(*eventPropertiesNTemplate)
		nBinsFloat := math.Min(float64(nDimensions*num_NUMERIC_BINS_PER_DIMENSION),
			float64(num_MAX_NUMERIC_MULTI_BINS))
		nBins := int(math.Max(1.0, nBinsFloat))
		puNHist, err := Hist.NewNumericHistogram(nBins, nDimensions, eventPropertiesNTemplate)
		if err != nil {
			return nil, err
		}
		pattern.PerUserEventNumericProperties = puNHist
		poNHist, err := Hist.NewNumericHistogram(nBins, nDimensions, eventPropertiesNTemplate)
		if err != nil {
			return nil, err
		}
		pattern.PerOccurrenceEventNumericProperties = poNHist
	}

	if eventPropertiesCTemplate != nil {
		cDimensions := len(*eventPropertiesCTemplate)
		cBinsFloat := math.Min(float64(cDimensions*num_CATEGORICAL_BINS_PER_DIMENSION),
			float64(num_MAX_CATEGORICAL_MULTI_BINS))
		cBins := int(math.Max(1.0, cBinsFloat))
		puCHist, err := Hist.NewCategoricalHistogram(cBins, cDimensions, eventPropertiesCTemplate)
		if err != nil {
			return nil, err
		}
		pattern.PerUserEventCategoricalProperties = puCHist
		poCHist, err := Hist.NewCategoricalHistogram(cBins, cDimensions, eventPropertiesCTemplate)
		if err != nil {
			return nil, err
		}
		pattern.PerOccurrenceEventCategoricalProperties = poCHist
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

func (p *Pattern) updateGenericHistogram() error {
	if p.currentUserOccurrenceCount > 0 {
		// Update Generic Histogram.
		histMap := map[string]float64{}
		histMap[U.UP_JOIN_TIME] = float64(p.currentUserJoinTimestamp)
		for i, eventName := range p.EventNames {
			ocList := p.currentUserEventOccurenceCounts[eventName]
			histMap[PatternPropertyKey(i, U.EP_FIRST_SEEN_OCCURRENCE_COUNT)] = float64(ocList[0])
			histMap[PatternPropertyKey(i, U.EP_LAST_SEEN_OCCURRENCE_COUNT)] = float64(ocList[len(ocList)-1])

			stList := p.currentUserEventTimestamps[eventName]
			histMap[PatternPropertyKey(i, U.EP_FIRST_SEEN_TIME)] = float64(stList[0])
			histMap[PatternPropertyKey(i, U.EP_LAST_SEEN_TIME)] = float64(stList[len(ocList)-1])
			firstSeenSinceUserJoin := math.Max(0, float64(stList[0]-p.currentUserJoinTimestamp))
			lastSeenSinceUserJoin := math.Max(0, float64(stList[len(stList)-1]-p.currentUserJoinTimestamp))
			histMap[PatternPropertyKey(i, U.EP_FIRST_SEEN_SINCE_USER_JOIN)] = firstSeenSinceUserJoin
			histMap[PatternPropertyKey(i, U.EP_LAST_SEEN_SINCE_USER_JOIN)] = lastSeenSinceUserJoin
		}
		err := p.GenericPropertiesHistogram.AddMap(histMap)
		return err
	}
	return nil
}

func (p *Pattern) ResetAfterLastUser() error {
	err := p.updateGenericHistogram()
	return err
}

func (p *Pattern) ResetForNewUser(userId string, userJoinTimestamp int64) error {
	if userId == "" || userJoinTimestamp <= 0 {
		return fmt.Errorf("Missing userId or userCreatedTime.")
	}

	p.TotalUserCount += 1
	if err := p.updateGenericHistogram(); err != nil {
		return err
	}

	// Reinitialize all private variables maintained per user event stream.
	p.waitIndex = 0
	p.currentUserId = userId
	p.currentUserJoinTimestamp = userJoinTimestamp
	p.previousEventTimestamp = 0
	p.currentUserOccurrenceCount = 0
	p.currentUserEventTimestamps = make(map[string][]int64)
	p.currentUserEventOccurenceCounts = make(map[string][]uint)
	p.currentEPropertiesNMap = make(map[string]float64)
	p.currentEPropertiesCMap = make(map[string]string)
	p.currentUPropertiesNMap = make(map[string]float64)
	p.currentUPropertiesCMap = make(map[string]string)
	return nil
}

func clipCategoricalValue(catValue string) string {
	MAX_CATEGORICAL_STRING_LENGTH := 50
	if len(catValue) < MAX_CATEGORICAL_STRING_LENGTH {
		return catValue
	}
	// If it is a url just use the Hostname+Path.
	// It's a common case.
	u, err := url.Parse(catValue)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"catValue": catValue,
		}).Debug(err)

		return catValue[:MAX_CATEGORICAL_STRING_LENGTH]
	}

	hostPath := fmt.Sprintf("%s%s", u.Hostname(), u.EscapedPath())
	if len(hostPath) > 0 && len(hostPath) < MAX_CATEGORICAL_STRING_LENGTH {
		return hostPath
	}

	return catValue[:MAX_CATEGORICAL_STRING_LENGTH]
}

func AddNumericAndCategoricalProperties(projectID uint64, eventName string,
	eventIndex int, properties map[string]interface{},
	nMap map[string]float64, cMap map[string]string, isUserProperty bool) {
	for key, value := range properties {
		propertyType := store.GetStore().GetPropertyTypeByKeyValue(projectID, eventName, key, value, isUserProperty)
		if propertyType == U.PropertyTypeNumerical {
			numValue, _ := U.GetPropertyValueAsFloat64(value)
			nMap[PatternPropertyKey(eventIndex, key)] = float64(numValue)
		} else if propertyType == U.PropertyTypeCategorical {
			categoricalValue := U.GetPropertyValueAsString(value)
			categoricalValue = clipCategoricalValue(categoricalValue)
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
func (p *Pattern) CountForEvent(projectID uint64,
	eventName string, eventTimestamp int64, eventProperties map[string]interface{},
	userProperties map[string]interface{}, eventCardinality uint, userId string,
	userJoinTimestamp int64, shouldCountOccurence bool) error {

	if eventName == "" || eventTimestamp <= 0 {
		return fmt.Errorf("Missing eventId or eventTimestamp.")
	}

	if userId != p.currentUserId {
		return fmt.Errorf(
			fmt.Sprintf("Mismatch in User data. userId: %s, userJoinTime: %v, pattern userId: %s, pattern userJoinTime: %d",
				userId, userJoinTimestamp, p.currentUserId, p.currentUserJoinTimestamp))
	}

	if p.currentUserJoinTimestamp != userJoinTimestamp {
		// Can happen when multiple userId's have the same customerUserId.
		minJoinTimestamp := int64(math.Max(math.Min(
			float64(userJoinTimestamp), float64(p.currentUserJoinTimestamp)), 0.0))
		log.Debug(fmt.Sprintf("Mismatch in User data.userJoinTime: %v, pattern userJoinTime: %d. Pattern timestamp will change to %d",
			userJoinTimestamp, p.currentUserJoinTimestamp, minJoinTimestamp))
		p.currentUserJoinTimestamp = minJoinTimestamp
	}

	// Check whether events are in order.
	if eventTimestamp < p.previousEventTimestamp {
		return fmt.Errorf(
			"Event Timings not in order. user: %s, event %s, userJoinTime:%d, eventTimestamp :%d, previousEventTimestamp: %d",
			userId, eventName, p.currentUserJoinTimestamp, eventTimestamp, p.previousEventTimestamp)
	}
	if eventTimestamp < p.currentUserJoinTimestamp {
		// Ignoring this error for now, since there are no DB checks to avoid
		// these user input values.
		log.Debug(fmt.Sprintf("Event occurs before creation for user:%s", p.currentUserId))
	}
	p.previousEventTimestamp = eventTimestamp

	// Start collecting timestamps and occurrences of an event
	// only after we have seen atleast one occurrence of
	// preceding events in sequence.
	if len(p.currentUserEventTimestamps[eventName]) == 0 {
		if eventName == p.EventNames[p.waitIndex] {
			p.currentUserEventTimestamps[eventName] = []int64{eventTimestamp}
		}
	} else {
		p.currentUserEventTimestamps[eventName] = append(
			p.currentUserEventTimestamps[eventName], eventTimestamp)
	}
	if len(p.currentUserEventOccurenceCounts[eventName]) == 0 {
		if eventName == p.EventNames[p.waitIndex] {
			p.currentUserEventOccurenceCounts[eventName] = []uint{eventCardinality}
		}
	} else {
		p.currentUserEventOccurenceCounts[eventName] = append(
			p.currentUserEventOccurenceCounts[eventName], eventCardinality)
	}

	if eventName == p.EventNames[p.waitIndex] {
		// Update seen properties.
		AddNumericAndCategoricalProperties(projectID, eventName,
			p.waitIndex, eventProperties, p.currentEPropertiesNMap, p.currentEPropertiesCMap, false)
		AddNumericAndCategoricalProperties(projectID, "",
			p.waitIndex, userProperties, p.currentUPropertiesNMap, p.currentUPropertiesCMap, true)

		p.waitIndex += 1

		pLen := len(p.EventNames)
		if p.waitIndex == pLen {
			p.currentUserOccurrenceCount += 1

			// Record the pattern occurrence.
			if shouldCountOccurence == true {
				p.PerOccurrenceCount += 1
				// Update properties histograms.
				if p.PerOccurrenceEventNumericProperties != nil {
					if err := p.PerOccurrenceEventNumericProperties.AddMap(p.currentEPropertiesNMap); err != nil {
						return err
					}
				}
				if p.PerOccurrenceEventCategoricalProperties != nil {
					if err := p.PerOccurrenceEventCategoricalProperties.AddMap(p.currentEPropertiesCMap); err != nil {
						return err
					}
				}
				if p.PerOccurrenceUserNumericProperties != nil {
					if err := p.PerOccurrenceUserNumericProperties.AddMap(p.currentUPropertiesNMap); err != nil {
						return err
					}
				}
				if p.PerOccurrenceUserCategoricalProperties != nil {
					if err := p.PerOccurrenceUserCategoricalProperties.AddMap(p.currentUPropertiesCMap); err != nil {
						return err
					}
				}
			}
			if p.currentUserOccurrenceCount == 1 {
				p.PerUserCount += 1

				// Update properties histograms.
				if p.PerUserEventNumericProperties != nil {
					if err := p.PerUserEventNumericProperties.AddMap(p.currentEPropertiesNMap); err != nil {
						return err
					}
				}
				if p.PerUserEventCategoricalProperties != nil {
					if err := p.PerUserEventCategoricalProperties.AddMap(p.currentEPropertiesCMap); err != nil {
						return err
					}
				}
				if p.PerUserUserNumericProperties != nil {
					if err := p.PerUserUserNumericProperties.AddMap(p.currentUPropertiesNMap); err != nil {
						return err
					}
				}
				if p.PerUserUserCategoricalProperties != nil {
					if err := p.PerUserUserCategoricalProperties.AddMap(p.currentUPropertiesCMap); err != nil {
						return err
					}
				}
			}

			// Reset.
			p.waitIndex = 0
			p.currentEPropertiesNMap = make(map[string]float64)
			p.currentEPropertiesCMap = make(map[string]string)
			p.currentUPropertiesNMap = make(map[string]float64)
			p.currentUPropertiesCMap = make(map[string]string)
		}
	}
	return nil
}

func (p *Pattern) GetPerUserCount(
	patternConstraints []EventConstraints) (uint, error) {
	pLen := len(p.EventNames)
	if patternConstraints != nil && (len(patternConstraints) != pLen) {
		errorString := fmt.Sprintf(
			"Constraint length %d does not match pattern length %d for pattern %v",
			len(patternConstraints), pLen, p.EventNames)
		log.Error(errorString)
		return 0, fmt.Errorf(errorString)
	}
	EPNMapUpperBounds := make(map[string]float64)
	EPNMapLowerBounds := make(map[string]float64)
	EPCMapEquality := make(map[string]string)
	UPNMapUpperBounds := make(map[string]float64)
	UPNMapLowerBounds := make(map[string]float64)
	UPCMapEquality := make(map[string]string)
	GPMapUpperBounds := make(map[string]float64)
	GPMapLowerBounds := make(map[string]float64)
	for i, ecs := range patternConstraints {
		for _, ncs := range ecs.EPNumericConstraints {
			if U.IsGenericEventProperty(&ncs.PropertyName) {
				key := PatternPropertyKey(i, ncs.PropertyName)
				GPMapLowerBounds[key] = ncs.LowerBound
				GPMapUpperBounds[key] = ncs.UpperBound
			} else {
				key := PatternPropertyKey(i, ncs.PropertyName)
				EPNMapLowerBounds[key] = ncs.LowerBound
				EPNMapUpperBounds[key] = ncs.UpperBound
			}
		}
		for _, ccs := range ecs.EPCategoricalConstraints {
			key := PatternPropertyKey(i, ccs.PropertyName)
			EPCMapEquality[key] = appendPropertyValues(EPCMapEquality[key], ccs.PropertyValue, ccs.Operator)
		}

		for _, ncs := range ecs.UPNumericConstraints {
			if U.IsGenericUserProperty(&ncs.PropertyName) {
				key := ncs.PropertyName
				GPMapLowerBounds[key] = ncs.LowerBound
				GPMapUpperBounds[key] = ncs.UpperBound
			} else {
				key := PatternPropertyKey(i, ncs.PropertyName)
				UPNMapLowerBounds[key] = ncs.LowerBound
				UPNMapUpperBounds[key] = ncs.UpperBound
			}
		}
		for _, ccs := range ecs.UPCategoricalConstraints {
			key := PatternPropertyKey(i, ccs.PropertyName)
			UPCMapEquality[key] = appendPropertyValues(UPCMapEquality[key], ccs.PropertyValue, ccs.Operator)
		}
	}

	GPNumericUpperCDF := 1.0
	GPNumericLowerCDF := 0.0
	if p.GenericPropertiesHistogram != nil && len(GPMapLowerBounds) > 0 {
		GPNumericUpperCDF = p.GenericPropertiesHistogram.CDFFromMap(GPMapUpperBounds)
		GPNumericLowerCDF = p.GenericPropertiesHistogram.CDFFromMap(GPMapLowerBounds)
	}

	EPNumericUpperCDF := 1.0
	EPNumericLowerCDF := 0.0
	if p.PerUserEventNumericProperties != nil && len(EPNMapLowerBounds) > 0 {
		EPNumericUpperCDF = p.PerUserEventNumericProperties.CDFFromMap(EPNMapUpperBounds)
		EPNumericLowerCDF = p.PerUserEventNumericProperties.CDFFromMap(EPNMapLowerBounds)
	}
	EPCategoricalPDF := 1.0
	if p.PerUserEventCategoricalProperties != nil && len(EPCMapEquality) > 0 {
		var err error
		EPCategoricalPDF, err = p.PerUserEventCategoricalProperties.PDFFromMap(EPCMapEquality)
		if err != nil {
			return 0, err
		}
	}

	UPNumericUpperCDF := 1.0
	UPNumericLowerCDF := 0.0
	if p.PerUserUserNumericProperties != nil && len(UPNMapLowerBounds) > 0 {
		UPNumericUpperCDF = p.PerUserUserNumericProperties.CDFFromMap(UPNMapUpperBounds)
		UPNumericLowerCDF = p.PerUserUserNumericProperties.CDFFromMap(UPNMapLowerBounds)
	}
	UPCategoricalPDF := 1.0
	if p.PerUserUserCategoricalProperties != nil && len(UPCMapEquality) > 0 {
		var err error
		UPCategoricalPDF, err = p.PerUserUserCategoricalProperties.PDFFromMap(UPCMapEquality)
		if err != nil {
			return 0, err
		}
	}

	count := (float64(p.PerUserCount) *
		(GPNumericUpperCDF - GPNumericLowerCDF) *
		(EPNumericUpperCDF - EPNumericLowerCDF) *
		EPCategoricalPDF *
		(UPNumericUpperCDF - UPNumericLowerCDF) *
		UPCategoricalPDF)

	if count < 0 {
		log.WithFields(log.Fields{
			"GPNumericUpperCDF":  GPNumericUpperCDF,
			"GPNumericLowerCDF":  GPNumericLowerCDF,
			"EPNumericUpperCDF":  EPNumericUpperCDF,
			"EPNumericLowerCDF":  EPNumericLowerCDF,
			"EPCategoricalPDF":   EPCategoricalPDF,
			"UPNumericUpperCDF":  UPNumericUpperCDF,
			"UPNumericLowerCDF":  UPNumericLowerCDF,
			"UPCategoricalPDF":   UPCategoricalPDF,
			"pattern":            p.String(),
			"patternConstraints": patternConstraints,
			"patternCount":       p.PerUserCount,
			"finalCount":         count,
		}).Info("Computed CDF's and PDF's")
		errorString := "Final count is less than 0."
		log.Error(errorString)
		return 0, fmt.Errorf(errorString)
	}
	return uint(count), nil
}

func appendPropertyValues(existingValue string, addedValue string, operator string) string {
	if existingValue == "" {
		if operator == NOT_EQUALS_OPERATOR_CONST {
			return "!=" + "," + addedValue
		}
		return addedValue
	}
	return (existingValue + "," + addedValue)
}

func (p *Pattern) GetPerOccurrenceCount(
	patternConstraints []EventConstraints) (uint, error) {
	pLen := len(p.EventNames)
	if patternConstraints != nil && (len(patternConstraints) != pLen) {
		errorString := fmt.Sprintf(
			"Constraint length %d does not match pattern length %d for pattern %v",
			len(patternConstraints), pLen, p.EventNames)
		log.Error(errorString)
		return 0, fmt.Errorf(errorString)
	}
	if p.PerOccurrenceEventCategoricalProperties == nil ||
		p.PerOccurrenceEventNumericProperties == nil ||
		p.PerOccurrenceUserCategoricalProperties == nil ||
		p.PerOccurrenceUserNumericProperties == nil {
		return 0, fmt.Errorf("Unsupported pattern for per occurrence count")
	}
	EPNMapUpperBounds := make(map[string]float64)
	EPNMapLowerBounds := make(map[string]float64)
	EPCMapEquality := make(map[string]string)
	UPNMapUpperBounds := make(map[string]float64)
	UPNMapLowerBounds := make(map[string]float64)
	UPCMapEquality := make(map[string]string)
	for i, ecs := range patternConstraints {
		for _, ncs := range ecs.EPNumericConstraints {
			key := PatternPropertyKey(i, ncs.PropertyName)
			EPNMapLowerBounds[key] = ncs.LowerBound
			EPNMapUpperBounds[key] = ncs.UpperBound
		}
		for _, ccs := range ecs.EPCategoricalConstraints {
			key := PatternPropertyKey(i, ccs.PropertyName)
			EPCMapEquality[key] = appendPropertyValues(EPCMapEquality[key], ccs.PropertyValue, ccs.Operator)
		}

		for _, ncs := range ecs.UPNumericConstraints {
			key := PatternPropertyKey(i, ncs.PropertyName)
			UPNMapLowerBounds[key] = ncs.LowerBound
			UPNMapUpperBounds[key] = ncs.UpperBound
		}
		for _, ccs := range ecs.UPCategoricalConstraints {
			key := PatternPropertyKey(i, ccs.PropertyName)
			UPCMapEquality[key] = appendPropertyValues(UPCMapEquality[key], ccs.PropertyValue, ccs.Operator)

		}
	}

	EPNumericUpperCDF := 1.0
	EPNumericLowerCDF := 0.0
	if p.PerOccurrenceEventNumericProperties != nil && len(EPNMapLowerBounds) > 0 {
		EPNumericUpperCDF = p.PerOccurrenceEventNumericProperties.CDFFromMap(EPNMapUpperBounds)
		EPNumericLowerCDF = p.PerOccurrenceEventNumericProperties.CDFFromMap(EPNMapLowerBounds)
	}
	EPCategoricalPDF := 1.0
	if p.PerOccurrenceEventCategoricalProperties != nil && len(EPCMapEquality) > 0 {
		var err error
		EPCategoricalPDF, err = p.PerOccurrenceEventCategoricalProperties.PDFFromMap(EPCMapEquality)
		if err != nil {
			return 0, err
		}
	}

	UPNumericUpperCDF := 1.0
	UPNumericLowerCDF := 0.0
	if p.PerOccurrenceUserNumericProperties != nil && len(UPNMapLowerBounds) > 0 {
		UPNumericUpperCDF = p.PerOccurrenceUserNumericProperties.CDFFromMap(UPNMapUpperBounds)
		UPNumericLowerCDF = p.PerOccurrenceUserNumericProperties.CDFFromMap(UPNMapLowerBounds)
	}
	UPCategoricalPDF := 1.0
	if p.PerOccurrenceUserCategoricalProperties != nil && len(UPCMapEquality) > 0 {
		var err error
		UPCategoricalPDF, err = p.PerOccurrenceUserCategoricalProperties.PDFFromMap(UPCMapEquality)
		if err != nil {
			return 0, err
		}
	}

	count := (float64(p.PerOccurrenceCount) *
		(EPNumericUpperCDF - EPNumericLowerCDF) *
		EPCategoricalPDF *
		(UPNumericUpperCDF - UPNumericLowerCDF) *
		UPCategoricalPDF)

	if count < 0 {
		log.WithFields(log.Fields{
			"EPNumericUpperCDF":      EPNumericUpperCDF,
			"EPNumericLowerCDF":      EPNumericLowerCDF,
			"EPCategoricalPDF":       EPCategoricalPDF,
			"UPNumericUpperCDF":      UPNumericUpperCDF,
			"UPNumericLowerCDF":      UPNumericLowerCDF,
			"UPCategoricalPDF":       UPCategoricalPDF,
			"pattern":                p.String(),
			"patternConstraints":     patternConstraints,
			"patternOccurrenceCount": p.PerOccurrenceCount,
			"finalCount":             count,
		}).Info("Computed CDF's and PDF's")
		errorString := "Final count is less than 0."
		log.Error(errorString)
		return 0, fmt.Errorf(errorString)
	}
	return uint(count), nil
}

func (p *Pattern) GetCount(patternConstraints []EventConstraints, countType string) (uint, error) {
	if countType == COUNT_TYPE_PER_USER {
		return p.GetPerUserCount(patternConstraints)
	} else if countType == COUNT_TYPE_PER_OCCURRENCE {
		return p.GetPerOccurrenceCount(patternConstraints)
	}
	return 0, fmt.Errorf(fmt.Sprintf("Unrecognized count type: %s", countType))
}

func (p *Pattern) GetPerUserEventPropertyRanges(
	eventIndex int, propertyName string) ([][2]float64, bool) {
	if predefinedBinRanges, found := U.GetPredefinedBinRanges(propertyName); found {
		return predefinedBinRanges, true
	}
	// Return the ranges of the bin [min, max], in which the numeric values for the event property occurr.
	return p.PerUserEventNumericProperties.GetBinRanges(PatternPropertyKey(eventIndex, propertyName)), false
}

func (p *Pattern) GetPerUserUserPropertyRanges(
	eventIndex int, propertyName string) ([][2]float64, bool) {
	if predefinedBinRanges, found := U.GetPredefinedBinRanges(propertyName); found {
		return predefinedBinRanges, true
	}
	// Return the ranges of the bin [min, max], in which the numeric values for the event property occurr.
	return p.PerUserUserNumericProperties.GetBinRanges(PatternPropertyKey(eventIndex, propertyName)), false
}

func (p *Pattern) GetPerOccurrenceEventPropertyRanges(
	eventIndex int, propertyName string) ([][2]float64, bool) {
	if predefinedBinRanges, found := U.GetPredefinedBinRanges(propertyName); found {
		return predefinedBinRanges, true
	}
	// Return the ranges of the bin [min, max], in which the numeric values for the event property occurr.
	return p.PerOccurrenceEventNumericProperties.GetBinRanges(PatternPropertyKey(eventIndex, propertyName)), false
}

func (p *Pattern) GetPerOccurrenceUserPropertyRanges(
	eventIndex int, propertyName string) ([][2]float64, bool) {
	if predefinedBinRanges, found := U.GetPredefinedBinRanges(propertyName); found {
		return predefinedBinRanges, true
	}
	// Return the ranges of the bin [min, max], in which the numeric values for the event property occurr.
	return p.PerOccurrenceUserNumericProperties.GetBinRanges(PatternPropertyKey(eventIndex, propertyName)), false
}

func (p *Pattern) GetPerUserEventPropertyValues(
	eventIndex int, propertyName string) []string {
	// Return the ranges of the bin [min, max], in which the numeric values for the event property occurr.
	return p.PerUserEventCategoricalProperties.GetBinValues(PatternPropertyKey(eventIndex, propertyName))
}

func (p *Pattern) GetPerUserUserPropertyValues(
	eventIndex int, propertyName string) []string {
	// Return the ranges of the bin [min, max], in which the numeric values for the event property occurr.
	return p.PerUserUserCategoricalProperties.GetBinValues(PatternPropertyKey(eventIndex, propertyName))
}

func (p *Pattern) GetPerOccurrenceEventPropertyValues(
	eventIndex int, propertyName string) []string {
	// Return the ranges of the bin [min, max], in which the numeric values for the event property occurr.
	return p.PerOccurrenceEventCategoricalProperties.GetBinValues(PatternPropertyKey(eventIndex, propertyName))
}

func (p *Pattern) GetPerOccurrenceUserPropertyValues(
	eventIndex int, propertyName string) []string {
	// Return the ranges of the bin [min, max], in which the numeric values for the event property occurr.
	return p.PerOccurrenceUserCategoricalProperties.GetBinValues(PatternPropertyKey(eventIndex, propertyName))
}

func (p *Pattern) String() string {
	return EventArrayToString(p.EventNames)
}

func EventArrayToString(eventNames []string) string {
	return strings.Join(eventNames, ",")
}

func IsEncodedEvent(eventName string) bool {
	return U.IsItreeCampaignEvent(eventName)
}

func ExtractCampaignName(eventName string) string {
	if strings.HasPrefix(eventName, "$session[campaign:") {
		prefix := strings.Split(eventName, "$session[campaign:")
		campaignName := strings.Split(prefix[1], "]")
		return "Campaign = " + campaignName[0]
	}
	if strings.HasPrefix(eventName, "$session[source:") {
		prefix := strings.Split(eventName, "$session[source:")
		campaignName := strings.Split(prefix[1], "]")
		return "Source = " + campaignName[0]
	}
	if strings.HasPrefix(eventName, "$session[medium:") {
		prefix := strings.Split(eventName, "$session[medium:")
		campaignName := strings.Split(prefix[1], "]")
		return "Medium = " + campaignName[0]
	}
	return ""
}
