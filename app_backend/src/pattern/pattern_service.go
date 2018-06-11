// pattern_service
package pattern

import (
	"fmt"
	"sort"
	"strings"
)

type PatternService struct {
	patterns []*Pattern
}

func NewPatternService(patterns []*Pattern) (*PatternService, error) {
	patternService := PatternService{
		patterns: patterns,
	}
	return &patternService, nil
}

func (ps *PatternService) Query(startEvent string, endEvent string) ([]*Pattern, error) {
	if startEvent == "" && endEvent == "" {
		return nil, fmt.Errorf("Invalid Query")
	}
	resPatterns := []*Pattern{}
	for _, p := range ps.patterns {
		if (startEvent == "" || strings.Compare(startEvent, p.EventNames[0]) == 0) &&
			(endEvent == "" || strings.Compare(endEvent, p.EventNames[len(p.EventNames)-1]) == 0) {
			resPatterns = append(resPatterns, p)
		}
	}
	// Sort in decreasing order of counts.
	sort.SliceStable(resPatterns,
		func(i, j int) bool {
			return resPatterns[i].Count > resPatterns[j].Count
		})
	return resPatterns, nil
}
