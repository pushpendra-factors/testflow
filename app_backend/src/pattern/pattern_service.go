// pattern_service
package pattern

import (
	"fmt"
	"sort"
	"strings"
)

type PatternWrapper struct {
	patterns         []*Pattern
	perUserCountsMap map[string]uint
	countsMap        map[string]uint
}

type PatternService struct {
	patternsMap map[uint64]*PatternWrapper
}

func NewPatternService(patternsMap map[uint64][]*Pattern) (*PatternService, error) {
	patternService := PatternService{patternsMap: map[uint64]*PatternWrapper{}}

	for projectId, patterns := range patternsMap {
		patternWrapper := PatternWrapper{
			patterns: patterns,
		}
		perUserCountsMap := make(map[string]uint)
		countsMap := make(map[string]uint)
		for _, p := range patterns {
			perUserCountsMap[p.String()] = p.OncePerUserCount
			countsMap[p.String()] = p.Count
		}
		patternWrapper.perUserCountsMap = perUserCountsMap
		patternWrapper.countsMap = countsMap
		patternService.patternsMap[projectId] = &patternWrapper
	}
	return &patternService, nil
}

func (ps *PatternService) GetPerUserCount(projectId uint64, eventNames []string) (uint, bool) {
	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return 0, false
	}
	c, ok := pw.perUserCountsMap[strings.Join(eventNames, ",")]
	return c, ok
}

func (ps *PatternService) GetCount(projectId uint64, eventNames []string) (uint, bool) {
	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return 0, false
	}
	c, ok := pw.countsMap[strings.Join(eventNames, ",")]
	return c, ok
}

func (ps *PatternService) Query(projectId uint64, startEvent string, endEvent string) ([]*Pattern, error) {
	maxPatterns := 50
	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("No patterns for projectId:%d", projectId))
	}
	if startEvent == "" && endEvent == "" {
		return nil, fmt.Errorf("Invalid Query")
	}
	resPatterns := []*Pattern{}
	for _, p := range pw.patterns {
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
	if len(resPatterns) > maxPatterns {
		resPatterns = resPatterns[:maxPatterns]
	}
	return resPatterns, nil
}
