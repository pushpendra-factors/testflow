// pattern_service
package pattern

import (
	"fmt"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

type PatternWrapper struct {
	patterns         []*Pattern
	perUserCountsMap map[string]uint
	countsMap        map[string]uint
}

type PatternService struct {
	patternsMap map[uint64]*PatternWrapper
}

type result struct {
	EventNames     []string  `json:"event_names"`
	Timings        []float64 `json:"timings"`
	Cardinalities  []float64 `json:"cardinalities"`
	Repeats        []float64 `json:"repeats"`
	Counts         []uint    `json:"counts"`
	PerUserCounts  []uint    `json:"per_user_counts"`
	TotalUserCount uint      `json:"total_user_count"`
}

type PatternServiceResults []*result

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
	c, ok := pw.perUserCountsMap[eventArrayToString(eventNames)]
	return c, ok
}

func (ps *PatternService) GetCount(projectId uint64, eventNames []string) (uint, bool) {
	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return 0, false
	}
	c, ok := pw.countsMap[eventArrayToString(eventNames)]
	return c, ok
}

func (ps *PatternService) buildResultsFromPatterns(projectId uint64, patterns []*Pattern) PatternServiceResults {
	results := PatternServiceResults{}
	for _, p := range patterns {
		r := result{
			EventNames:     p.EventNames,
			Timings:        []float64{},
			Cardinalities:  []float64{},
			Repeats:        []float64{},
			Counts:         []uint{},
			PerUserCounts:  []uint{},
			TotalUserCount: p.UserCount,
		}
		for i := 0; i < len(p.EventNames); i++ {
			r.Timings = append(r.Timings, p.Timings[i].Quantile(0.5))
			r.Repeats = append(r.Repeats, p.Repeats[i].Quantile(0.5))
			r.Cardinalities = append(r.Cardinalities, p.EventCardinalities[i].Quantile(0.5))
			subsequenceCount, ok := ps.GetCount(projectId, p.EventNames[:i+1])
			if !ok {
				log.Errorf(fmt.Sprintf(
					"Subsequence %s not as frequent as sequence %s",
					eventArrayToString(p.EventNames[:i+1]), ","), p.String())
				r.Counts = append(r.Counts, p.Count)
			} else {
				r.Counts = append(r.Counts, subsequenceCount)
			}

			subsequencePerUserCount, ok := ps.GetPerUserCount(projectId, p.EventNames[:i+1])
			if !ok {
				log.Errorf(fmt.Sprintf(
					"Subsequence %s not as frequent as sequence %s",
					eventArrayToString(p.EventNames[:i+1]), ","), p.String())
				r.PerUserCounts = append(r.PerUserCounts, p.OncePerUserCount)
			} else {
				r.PerUserCounts = append(r.PerUserCounts, subsequencePerUserCount)
			}
		}
		results = append(results, &r)
	}
	return results
}

func (ps *PatternService) Query(projectId uint64, startEvent string,
	endEvent string) (PatternServiceResults, error) {

	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("No patterns for projectId:%d", projectId))
	}
	if startEvent == "" && endEvent == "" {
		return nil, fmt.Errorf("Invalid Query")
	}
	matchPatterns := []*Pattern{}
	for _, p := range pw.patterns {
		if (startEvent == "" || strings.Compare(startEvent, p.EventNames[0]) == 0) &&
			(endEvent == "" || strings.Compare(endEvent, p.EventNames[len(p.EventNames)-1]) == 0) {
			matchPatterns = append(matchPatterns, p)
		}
	}

	results := ps.buildResultsFromPatterns(projectId, matchPatterns)
	// Sort in decreasing order of per user counts of the sequence.
	sort.SliceStable(results,
		func(i, j int) bool {
			lenI := len(results[i].PerUserCounts)
			lenJ := len(results[j].PerUserCounts)
			return (results[i].PerUserCounts[lenI-1] > results[j].PerUserCounts[lenJ-1])
		})
	maxPatterns := 50
	if len(results) > maxPatterns {
		results = results[:maxPatterns]
	}

	return results, nil
}

func (ps *PatternService) Crunch(projectId uint64, endEvent string) (PatternServiceResults, error) {
	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("No patterns for projectId:%d", projectId))
	}
	if endEvent == "" {
		return nil, fmt.Errorf("Invalid Query")
	}

	iPatterns := []*Pattern{}
	if itree, err := BuildNewItree(endEvent, pw); err != nil {
		log.Error(err)
		return nil, err
	} else {
		for _, node := range itree.Nodes {
			iPatterns = append(iPatterns, node.Pattern)
		}
	}
	results := ps.buildResultsFromPatterns(projectId, iPatterns)

	// Sort in decreasing order of per user counts of the sequence.
	sort.SliceStable(results,
		func(i, j int) bool {
			lenI := len(results[i].PerUserCounts)
			lenJ := len(results[j].PerUserCounts)
			return (results[i].PerUserCounts[lenI-1] > results[j].PerUserCounts[lenJ-1])
		})

	maxPatterns := 50
	if len(results) > maxPatterns {
		results = results[:maxPatterns]
	}

	return results, nil
}
