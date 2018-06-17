package pattern

import (
	"bufio"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func GenCandidatesPair(p1 *Pattern, p2 *Pattern) (*Pattern, *Pattern, bool) {
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
	c1Pattern, _ := NewPattern(c1String)
	c2Pattern, _ := NewPattern(c2String)
	return c1Pattern, c2Pattern, true
}

func candidatesMapToSlice(candidatesMap map[string]*Pattern) []*Pattern {
	candidates := []*Pattern{}
	for _, v := range candidatesMap {
		candidates = append(candidates, v)
	}
	return candidates
}

func GenCandidates(currentPatterns []*Pattern, maxCandidates int) ([]*Pattern, uint, error) {
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
			if c1, c2, ok := GenCandidatesPair(currentPatterns[i], currentPatterns[j]); ok {
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
		panic(fmt.Errorf("More than max candidates generated."))
	}
	return candidatesMapToSlice(candidatesMap), currentMinCount, nil
}

func CountPatterns(scanner *bufio.Scanner, patterns []*Pattern) error {
	var seenUsers map[string]bool = make(map[string]bool)

	numEventsProcessed := 0
	waitingOnPatternsMap := make(map[string][]*Pattern)
	prevWaitPatternsMap := make(map[string][]*Pattern)
	// Initialize.
	for _, p := range patterns {
		waitEvent := p.EventNames[0]
		if _, ok := waitingOnPatternsMap[waitEvent]; !ok {
			waitingOnPatternsMap[waitEvent] = []*Pattern{}
		}
		waitingOnPatternsMap[waitEvent] = append(waitingOnPatternsMap[waitEvent], p)
	}

	for scanner.Scan() {
		line := scanner.Text()
		splits := strings.Split(line, ",")
		userId, eventName := splits[0], splits[2]
		userCreatedTime, err := time.Parse(time.RFC3339, splits[1])
		if err != nil {
			log.Fatal(err)
		}
		eventCreatedTime, err := time.Parse(time.RFC3339, splits[3])
		if err != nil {
			log.Fatal(err)
		}
		eventCardinality, err := strconv.ParseUint(splits[4], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		numEventsProcessed += 1
		if math.Mod(float64(numEventsProcessed), 1000.0) == 0.0 {
			log.Info(fmt.Sprintf("Processed %d events", numEventsProcessed))
		}

		_, isSeenUser := seenUsers[userId]
		if !isSeenUser {
			for _, p := range patterns {
				if err = p.ResetForNewUser(userId, userCreatedTime); err != nil {
					log.Fatal(err)
				}
			}
		}

		// Count Repeats.
		prevWaitPattens, ok := prevWaitPatternsMap[eventName]
		if ok {
			for _, p := range prevWaitPattens {
				if _, err = p.CountForEvent(eventName, eventCreatedTime, uint(eventCardinality), userId, userCreatedTime); err != nil {
					log.Error(err)
				}
			}
		}

		waitPatterns, _ := waitingOnPatternsMap[eventName]
		waitingOnPatternsMap[eventName] = []*Pattern{}
		prevWaitPatternsMap[eventName] = waitPatterns
		for _, p := range waitPatterns {
			var waitingOnEvent string
			if waitingOnEvent, err = p.CountForEvent(eventName, eventCreatedTime, uint(eventCardinality), userId, userCreatedTime); err != nil || waitingOnEvent == "" {
				log.Error(err)
			}
			waitingOnPatternsMap[waitingOnEvent] = append(waitingOnPatternsMap[waitingOnEvent], p)
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
func GenLenThreeCandidatePattern(pattern *Pattern, startPatterns []*Pattern,
	endPatterns []*Pattern, maxCandidates int) []*Pattern {
	if len(pattern.EventNames) != 2 {
		panic(fmt.Errorf("Unexpected length"))
	}
	sLen := len(startPatterns)
	eLen := len(endPatterns)
	minLen := int(math.Min(float64(sLen), float64(eLen)))

	eventsWithStartMap := make(map[string]bool)
	eventsWithEndMap := make(map[string]bool)
	for i := 0; i < sLen; i++ {
		eventsWithStartMap[startPatterns[i].EventNames[1]] = true
	}
	for i := 0; i < eLen; i++ {
		eventsWithEndMap[endPatterns[i].EventNames[0]] = true
	}

	candidatesMap := make(map[string]*Pattern)
	var err error
	// Alternate between startsWith and endsWith till the end of one.
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
				candidate, err = NewPattern(cString)
			} else {
				if reflect.DeepEqual(pattern.EventNames, endPatterns[j].EventNames) {
					continue
				}
				if found, _ := eventsWithStartMap[endPatterns[j].EventNames[0]]; !found {
					continue
				}
				cString = []string{pattern.EventNames[0]}
				cString = append(cString, endPatterns[j].EventNames...)
				candidate, err = NewPattern(cString)
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
				candidate, err = NewPattern(cString)
			} else {
				if reflect.DeepEqual(pattern.EventNames, endPatterns[j].EventNames) {
					continue
				}
				if found, _ := eventsWithStartMap[endPatterns[j].EventNames[0]]; !found {
					continue
				}
				cString = []string{pattern.EventNames[0]}
				cString = append(cString, endPatterns[j].EventNames...)
				candidate, err = NewPattern(cString)
			}
		}
		if err != nil {
			panic(err)
		}
		candidatesMap[candidate.String()] = candidate
		if len(candidatesMap) >= maxCandidates {
			return candidatesMapToSlice(candidatesMap)
		}
	}
	if err != nil {
		panic(err)
	}
	return candidatesMapToSlice(candidatesMap)
}
