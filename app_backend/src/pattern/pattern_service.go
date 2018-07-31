// pattern_service
package pattern

import (
	"fmt"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

type PatternWrapper struct {
	patterns []*Pattern
	pMap     map[string]*Pattern
}

type PatternService struct {
	patternsMap map[uint64]*PatternWrapper
}

type result struct {
	EventNames     []string  `json:"event_names"`
	Timings        []float64 `json:"timings"`
	Cardinalities  []float64 `json:"cardinalities"`
	Repeats        []float64 `json:"repeats"`
	PerUserCounts  []uint    `json:"per_user_counts"`
	TotalUserCount uint      `json:"total_user_count"`
}

type PatternServiceResults []*result

func NewPatternWrapper(patterns []*Pattern) *PatternWrapper {
	patternWrapper := PatternWrapper{
		patterns: patterns,
	}
	pMap := make(map[string]*Pattern)
	for _, p := range patterns {
		pMap[p.String()] = p
	}
	patternWrapper.pMap = pMap
	return &patternWrapper
}

func (pw *PatternWrapper) GetPerUserCount(eventNames []string,
	eCardLowerBound int, ecardUpperBound int) (uint, bool) {
	if p, ok := pw.pMap[eventArrayToString(eventNames)]; ok {
		return p.GetOncePerUserCount(eCardLowerBound, ecardUpperBound), true
	}
	return 0, false
}

func (pw *PatternWrapper) buildResultsFromPatterns(patterns []*Pattern,
	eCardLowerBound int, ecardUpperBound int) PatternServiceResults {
	results := PatternServiceResults{}

	for _, p := range patterns {
		r := result{
			EventNames:     p.EventNames,
			Timings:        []float64{},
			Cardinalities:  []float64{},
			Repeats:        []float64{},
			PerUserCounts:  []uint{},
			TotalUserCount: p.UserCount,
		}
		pLen := len(p.EventNames)
		for i := 0; i < pLen; i++ {
			r.Timings = append(r.Timings, p.Timings[i].Quantile(0.5))
			r.Repeats = append(r.Repeats, p.Repeats[i].Quantile(0.5))
			r.Cardinalities = append(r.Cardinalities, p.EventCardinalities[i].Quantile(0.5))

			var subsequencePerUserCount uint
			var found bool = true
			if i == pLen-1 {
				subsequencePerUserCount = p.GetOncePerUserCount(eCardLowerBound, ecardUpperBound)
			} else {
				subsequencePerUserCount, found = pw.GetPerUserCount(p.EventNames[:i+1], -1, -1)
			}
			if !found {
				log.Errorf(fmt.Sprintf(
					"Subsequence %s not as frequent as sequence %s",
					eventArrayToString(p.EventNames[:i+1]), ","), p.String())
				r.PerUserCounts = append(r.PerUserCounts, p.GetOncePerUserCount(eCardLowerBound, ecardUpperBound))
			} else {
				r.PerUserCounts = append(r.PerUserCounts, subsequencePerUserCount)
			}
		}
		results = append(results, &r)
	}
	return results
}

func NewPatternService(patternsMap map[uint64][]*Pattern) (*PatternService, error) {
	patternService := PatternService{patternsMap: map[uint64]*PatternWrapper{}}

	for projectId, patterns := range patternsMap {
		patternWrapper := NewPatternWrapper(patterns)
		patternService.patternsMap[projectId] = patternWrapper
	}
	return &patternService, nil
}

func (ps *PatternService) Query(projectId uint64, startEvent string,
	endEvent string, eCardLowerBound int, ecardUpperBound int) (PatternServiceResults, error) {

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

	results := pw.buildResultsFromPatterns(matchPatterns, eCardLowerBound, ecardUpperBound)
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

func (ps *PatternService) Crunch(projectId uint64, endEvent string,
	eCardLowerBound int, ecardUpperBound int) (PatternServiceResults, error) {
	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("No patterns for projectId:%d", projectId))
	}
	if endEvent == "" {
		return nil, fmt.Errorf("Invalid Query")
	}

	iPatterns := []*Pattern{}
	if itree, err := BuildNewItree(endEvent, eCardLowerBound, ecardUpperBound, pw); err != nil {
		log.Error(err)
		return nil, err
	} else {
		for _, node := range itree.Nodes {
			iPatterns = append(iPatterns, node.Pattern)
		}
	}
	results := pw.buildResultsFromPatterns(iPatterns, eCardLowerBound, ecardUpperBound)

	maxPatterns := 50
	if len(results) > maxPatterns {
		results = results[:maxPatterns]
	}

	return results, nil
}
