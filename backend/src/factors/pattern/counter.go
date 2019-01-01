package pattern

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type CounterEventFormat struct {
	UserId            string                 `json:"uid"`
	UserJoinTimestamp int64                  `json:"ujt"`
	EventName         string                 `json:"en"`
	EventTimestamp    int64                  `json:"et"`
	EventCardinality  uint                   `json:"ecd"`
	EventProperties   map[string]interface{} `json:"epr"`
	UserProperties    map[string]interface{} `json:"upr"`
}

type PropertiesInfo struct {
	NumericPropertyKeys          map[string]bool
	CategoricalPropertyKeyValues map[string]map[string]bool
}

// Event name to corresponding properties Info.
type UserAndEventsInfo struct {
	UserPropertiesInfo     *PropertiesInfo
	EventPropertiesInfoMap *map[string]*PropertiesInfo
}

func GenCandidatesPair(p1 *Pattern, p2 *Pattern, userAndEventsInfo *UserAndEventsInfo) (*Pattern, *Pattern, bool) {
	p1Len := len(p1.EventNames)
	p2Len := len(p2.EventNames)
	if p1Len != p2Len || p1Len == 0 {
		return nil, nil, false
	}

	numDifferent := 0
	differentIndex := -1
	for i := 0; i < p1Len; i++ {
		if strings.Compare(p1.EventNames[i], p2.EventNames[i]) != 0 {
			numDifferent += 1
			differentIndex = i
			if numDifferent > 1 {
				// Candidates cannot be generated from patterns that differ
				// by more than one event.
				return nil, nil, false
			}
		}
	}

	if numDifferent != 1 {
		return nil, nil, false
	}

	c1String := make([]string, p1Len)
	copy(c1String, p1.EventNames)
	c2String := make([]string, p1Len)
	copy(c2String, p1.EventNames)
	// Insert the different event in p2 before and after differentIndex in p1.
	c1String = append(c1String[:differentIndex], append([]string{p2.EventNames[differentIndex]}, c1String[differentIndex:]...)...)
	c2String = append(c2String[:differentIndex+1], append([]string{p2.EventNames[differentIndex]}, c2String[differentIndex+1:]...)...)
	c1Pattern, _ := NewPattern(c1String, userAndEventsInfo)
	c2Pattern, _ := NewPattern(c2String, userAndEventsInfo)
	return c1Pattern, c2Pattern, true
}

func candidatesMapToSlice(candidatesMap map[string]*Pattern) []*Pattern {
	candidates := []*Pattern{}
	for _, v := range candidatesMap {
		candidates = append(candidates, v)
	}
	return candidates
}

func GenCandidates(currentPatterns []*Pattern, maxCandidates int, userAndEventsInfo *UserAndEventsInfo) (
	[]*Pattern, uint, error) {
	numPatterns := len(currentPatterns)
	var currentMinCount uint

	if numPatterns == 0 {
		return nil, currentMinCount, fmt.Errorf("Zero Patterns")
	}
	// Sort current patterns in decreasing order of frequency.
	sort.Slice(
		currentPatterns,
		func(i, j int) bool {
			return currentPatterns[i].Count > currentPatterns[j].Count
		})
	candidatesMap := make(map[string]*Pattern)
	// Candidates are formed in decreasing order of frequent patterns till maxCandidates.
	for i := 0; i < numPatterns; i++ {
		for j := i + 1; j < numPatterns; j++ {
			if c1, c2, ok := GenCandidatesPair(
				currentPatterns[i], currentPatterns[j], userAndEventsInfo); ok {
				currentMinCount = currentPatterns[j].Count
				candidatesMap[c1.String()] = c1
				if len(candidatesMap) >= maxCandidates {
					return candidatesMapToSlice(candidatesMap), currentMinCount, nil
				}
				candidatesMap[c2.String()] = c2
				if len(candidatesMap) >= maxCandidates {
					return candidatesMapToSlice(candidatesMap), currentMinCount, nil
				}
			}
		}
	}
	if len(candidatesMap) > maxCandidates {
		log.Fatal("More than max candidates generated.")
	}
	return candidatesMapToSlice(candidatesMap), currentMinCount, nil
}

func deletePatternFromSlice(patternArray []*Pattern, pattern *Pattern) []*Pattern {
	// Delete all occurrences of the pattern.
	j := 0
	for _, p := range patternArray {
		if p != pattern {
			patternArray[j] = p
			j++
		}
	}
	patternArray = patternArray[:j]
	return patternArray
}

func PatternPropertyKey(patternIndex int, propertyName string) string {
	// property names are scoped by index of the event in pattern,
	// since different events can have same properties.
	return fmt.Sprintf("%d.%s", patternIndex, propertyName)
}

// Collects event info for the events initilaized in userAndEventsInfo.
const max_SEEN_PROPERTIES = 10000
const max_SEEN_PROPERTY_VALUES = 1000

func CollectPropertiesInfo(scanner *bufio.Scanner, userAndEventsInfo *UserAndEventsInfo) error {
	lineNum := 0
	userPropertiesInfo := userAndEventsInfo.UserPropertiesInfo
	eventInfoMap := userAndEventsInfo.EventPropertiesInfoMap

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		var eventDetails CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return err
		}
		for key, value := range eventDetails.UserProperties {
			if _, ok := value.(float64); ok {
				if len(userPropertiesInfo.NumericPropertyKeys) > max_SEEN_PROPERTIES {
					continue
				}
				userPropertiesInfo.NumericPropertyKeys[key] = true
			} else if categoricalValue, ok := value.(string); ok {
				if len(userPropertiesInfo.CategoricalPropertyKeyValues) > max_SEEN_PROPERTIES {
					continue
				}
				cmap, ok := userPropertiesInfo.CategoricalPropertyKeyValues[key]
				if !ok {
					cmap = make(map[string]bool)
					userPropertiesInfo.CategoricalPropertyKeyValues[key] = cmap
				}
				if len(categoricalValue) < max_SEEN_PROPERTY_VALUES {
					cmap[categoricalValue] = true
				}
			} else {
				log.WithFields(log.Fields{"property": key, "value": value, "line no": lineNum}).Info(
					"Ignoring non string, non numeric user property.")
			}
		}
		eventName := eventDetails.EventName
		eInfo, ok := (*eventInfoMap)[eventName]
		if !ok {
			log.WithFields(log.Fields{"event": eventName, "line no": lineNum}).Info("Unexpected event. Ignoring")
			continue
		}
		for key, value := range eventDetails.EventProperties {
			if _, ok := value.(float64); ok {
				if len(eInfo.NumericPropertyKeys) > max_SEEN_PROPERTIES {
					continue
				}
				eInfo.NumericPropertyKeys[key] = true
			} else if categoricalValue, ok := value.(string); ok {
				if len(eInfo.CategoricalPropertyKeyValues) > max_SEEN_PROPERTIES {
					continue
				}
				cmap, ok := eInfo.CategoricalPropertyKeyValues[key]
				if !ok {
					cmap = make(map[string]bool)
					eInfo.CategoricalPropertyKeyValues[key] = cmap
				}
				if len(categoricalValue) < max_SEEN_PROPERTY_VALUES {
					cmap[categoricalValue] = true
				}
			} else {
				log.WithFields(log.Fields{"event": eventName, "property": key, "value": value, "line no": lineNum}).Info(
					"Ignoring non string, non numeric event property.")
			}
		}
	}
	return nil
}

func CountPatterns(scanner *bufio.Scanner, patterns []*Pattern) error {
	var seenUsers map[string]bool = make(map[string]bool)

	numEventsProcessed := 0
	waitingOnPatternsMap := make(map[string][]*Pattern)
	prevWaitPatternsMap := make(map[string][]*Pattern)
	// Initialize.
	for _, p := range patterns {
		waitEvent := p.WaitingOn()
		if _, ok := waitingOnPatternsMap[waitEvent]; !ok {
			waitingOnPatternsMap[waitEvent] = []*Pattern{}
		}
		waitingOnPatternsMap[waitEvent] = append(waitingOnPatternsMap[waitEvent], p)
	}

	for scanner.Scan() {
		line := scanner.Text()
		var eventDetails CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return err
		}
		userId := eventDetails.UserId
		eventName := eventDetails.EventName
		eventProperties := eventDetails.EventProperties
		userProperties := eventDetails.UserProperties
		userJoinTimestamp := eventDetails.UserJoinTimestamp
		eventTimestamp := eventDetails.EventTimestamp
		eventCardinality := eventDetails.EventCardinality

		numEventsProcessed += 1
		if math.Mod(float64(numEventsProcessed), 1000.0) == 0.0 {
			log.Info(fmt.Sprintf("Processed %d events", numEventsProcessed))
		}

		_, isSeenUser := seenUsers[userId]
		if !isSeenUser {
			// Reinitialize.
			waitingOnPatternsMap = make(map[string][]*Pattern)
			prevWaitPatternsMap = make(map[string][]*Pattern)
			for _, p := range patterns {
				if err := p.ResetForNewUser(userId, userJoinTimestamp); err != nil {
					log.Fatal(err)
				}
				waitEvent := p.WaitingOn()
				if _, ok := waitingOnPatternsMap[waitEvent]; !ok {
					waitingOnPatternsMap[waitEvent] = []*Pattern{}
				}
				waitingOnPatternsMap[waitEvent] = append(waitingOnPatternsMap[waitEvent], p)
			}
		}

		// Count Repeats.
		prevWaitPattens, ok := prevWaitPatternsMap[eventName]
		if ok {
			for _, p := range prevWaitPattens {
				if _, err := p.CountForEvent(eventName, eventTimestamp, eventProperties,
					userProperties, uint(eventCardinality), userId, userJoinTimestamp); err != nil {
					log.Error(err)
				}
			}
		}

		waitPatterns, _ := waitingOnPatternsMap[eventName]
		waitingOnPatternsMap[eventName] = []*Pattern{}
		for _, p := range waitPatterns {
			waitingOn1 := p.WaitingOn()
			prevWaitingOn1 := p.PrevWaitingOn()
			if strings.Compare(waitingOn1, eventName) != 0 {
				log.Fatal(fmt.Errorf(
					"Pattern %s assumed to wait on %s but actually waiting on %s. Line %s",
					p.String(), eventName, waitingOn1, line))
			}
			waitingOn2, err := p.CountForEvent(eventName, eventTimestamp, eventProperties,
				userProperties, uint(eventCardinality), userId, userJoinTimestamp)
			if err != nil || waitingOn2 == "" {
				log.Error(err)
			}
			prevWaitingOn2 := p.PrevWaitingOn()
			if strings.Compare(prevWaitingOn1, prevWaitingOn2) != 0 {
				if prevWaitingOn1 != "" {
					if pwArray1, ok := prevWaitPatternsMap[prevWaitingOn1]; ok {
						pwArray1 = deletePatternFromSlice(pwArray1, p)
						prevWaitPatternsMap[prevWaitingOn1] = pwArray1
					}
				}
				if prevWaitingOn2 != "" {
					if pwArray2, ok := prevWaitPatternsMap[prevWaitingOn2]; ok {
						pwArray2 = deletePatternFromSlice(pwArray2, p)
						// Add the pattern.
						pwArray2 = append(pwArray2, p)
						prevWaitPatternsMap[prevWaitingOn2] = pwArray2
					} else {
						prevWaitPatternsMap[prevWaitingOn2] = []*Pattern{p}
					}
				}
			}
			if strings.Compare(waitingOn1, waitingOn2) == 0 && len(p.EventNames) != 1 {
				log.Fatal(fmt.Errorf(
					"Pattern %s waiting on %s did not get updated. Line %s",
					p.String(), waitingOn1, line))
			} else {
				waitingOnPatternsMap[waitingOn2] = append(waitingOnPatternsMap[waitingOn2], p)
			}
		}
		seenUsers[userId] = true
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return nil
}

// Special candidate generation method that generates upto maxCandidates with
// events that start with and end with the two event patterns.
func GenLenThreeCandidatePatterns(pattern *Pattern, startPatterns []*Pattern,
	endPatterns []*Pattern, maxCandidates int, userAndEventsInfo *UserAndEventsInfo) ([]*Pattern, error) {
	if len(pattern.EventNames) != 2 {
		return nil, fmt.Errorf(fmt.Sprintf("Pattern %s length is not two.", pattern.String()))
	}
	sLen := len(startPatterns)
	eLen := len(endPatterns)
	minLen := int(math.Min(float64(sLen), float64(eLen)))

	eventsWithStartMap := make(map[string]bool)
	eventsWithEndMap := make(map[string]bool)
	for i := 0; i < sLen; i++ {
		if len(startPatterns[i].EventNames) != 2 {
			return nil, fmt.Errorf("Start pattern %s of not length two.",
				startPatterns[i].String())
		}
		if strings.Compare(
			startPatterns[i].EventNames[0], pattern.EventNames[0]) != 0 {
			return nil, fmt.Errorf("Pattern %s does not match start event of %s",
				startPatterns[i].String(), pattern.String())
		}
		eventsWithStartMap[startPatterns[i].EventNames[1]] = true
	}
	for i := 0; i < eLen; i++ {
		if len(endPatterns[i].EventNames) != 2 {
			return nil, fmt.Errorf("End pattern %s of not length two.",
				endPatterns[i].String())
		}
		if strings.Compare(
			endPatterns[i].EventNames[len(endPatterns[i].EventNames)-1],
			pattern.EventNames[1]) != 0 {
			return nil, fmt.Errorf("Pattern %s does not match end event of %s",
				endPatterns[i].String(), pattern.String())
		}
		eventsWithEndMap[endPatterns[i].EventNames[0]] = true
	}

	candidatesMap := make(map[string]*Pattern)
	var err error
	// Alternate between startsWith and endsWith till the end of one.
	// The ordering of the patterns should be taken care by the caller.
	// The ones at the beginning are given higher priority.
	for i := 0; i < (sLen + eLen); i++ {
		var candidate *Pattern
		cString := make([]string, 2)
		if i < 2*minLen {
			j := int(i / 2)
			if math.Mod(float64(i), 2) < 1.0 {
				if reflect.DeepEqual(pattern.EventNames, startPatterns[j].EventNames) {
					continue
				}
				if found, _ := eventsWithEndMap[startPatterns[j].EventNames[1]]; !found {
					continue
				}
				copy(cString, startPatterns[j].EventNames)
				cString = append(cString, pattern.EventNames[1])
				candidate, err = NewPattern(cString, userAndEventsInfo)
			} else {
				if reflect.DeepEqual(pattern.EventNames, endPatterns[j].EventNames) {
					continue
				}
				if found, _ := eventsWithStartMap[endPatterns[j].EventNames[0]]; !found {
					continue
				}
				cString = []string{pattern.EventNames[0]}
				cString = append(cString, endPatterns[j].EventNames...)
				candidate, err = NewPattern(cString, userAndEventsInfo)
			}
		} else {
			j := i - minLen
			if sLen > eLen {
				if reflect.DeepEqual(pattern.EventNames, startPatterns[j].EventNames) {
					continue
				}
				if found, _ := eventsWithEndMap[startPatterns[j].EventNames[1]]; !found {
					continue
				}
				copy(cString, startPatterns[j].EventNames)
				cString = append(cString, pattern.EventNames[1])
				candidate, err = NewPattern(cString, userAndEventsInfo)
			} else {
				if reflect.DeepEqual(pattern.EventNames, endPatterns[j].EventNames) {
					continue
				}
				if found, _ := eventsWithStartMap[endPatterns[j].EventNames[0]]; !found {
					continue
				}
				cString = []string{pattern.EventNames[0]}
				cString = append(cString, endPatterns[j].EventNames...)
				candidate, err = NewPattern(cString, userAndEventsInfo)
			}
		}
		if err != nil {
			return nil, err
		}
		candidatesMap[candidate.String()] = candidate
		if len(candidatesMap) >= maxCandidates {
			return candidatesMapToSlice(candidatesMap), nil
		}
	}
	if err != nil {
		return nil, err
	}
	return candidatesMapToSlice(candidatesMap), nil
}
