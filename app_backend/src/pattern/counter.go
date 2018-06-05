package pattern

import (
	"bufio"
	"fmt"
	"sort"
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
	candidates := []*Pattern{}
	// Candidates are formed in decreasing order of frequent patterns till maxCandidates.
	for i := 0; i < numPatterns; i++ {
		for j := i + 1; j < numPatterns; j++ {
			if c1, c2, ok := GenCandidatesPair(currentPatterns[i], currentPatterns[j]); ok {
				candidates = append(candidates, c1, c2)
				currentMinCount = candidates[j].Count
				if len(candidates) >= maxCandidates-1 {
					return candidates, currentMinCount, nil
				}
			}
		}
	}
	if len(candidates) > maxCandidates {
		panic(fmt.Errorf("More than max candidates generated."))
	}
	return candidates, currentMinCount, nil
}

func CountPatterns(scanner *bufio.Scanner, patterns []*Pattern) error {
	var seenUsers map[string]bool = make(map[string]bool)

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

		_, isSeenUser := seenUsers[userId]
		for _, p := range patterns {
			if !isSeenUser {
				if err = p.ResetForNewUser(userId, userCreatedTime); err != nil {
					log.Fatal(err)
				}
			}
			if err = p.CountForEvent(eventName, eventCreatedTime, userId, userCreatedTime); err != nil {
				log.Error(err)
			}
		}
		seenUsers[userId] = true
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return nil
}
