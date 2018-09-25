// pattern_service
package pattern

import (
	"fmt"
	"math"
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

type funnelNodeResult struct {
	Data              []float64 `json:"data"`
	Event             string    `json:"event"`
	NodeType          string    `json:"node_type"`
	ConversionPercent float64   `json:"conversion_percent"`
}
type funnelNodeResults []funnelNodeResult
type graphResult struct {
	Type     string                   `json:"type"`
	Header   string                   `json:"header"`
	Labels   []string                 `json:"labels"`
	Datasets []map[string]interface{} `json:"datasets"`
}
type PatternServiceGraphResults struct {
	Charts []graphResult `json:"charts"`
}

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

func (pw *PatternWrapper) GetPattern(eventNames []string) (*Pattern, bool) {
	p, ok := pw.pMap[eventArrayToString(eventNames)]
	return p, ok
}

func eventStringWithConditions(event string, eCardLowerBound int, eCardUpperBound int) string {
	var eventString string = event
	if eCardLowerBound > 0 || eCardUpperBound > 0 {
		eventString += "("
		if eCardLowerBound > 0 && eCardUpperBound > 0 {
			eventString += fmt.Sprintf("%d to %d time", eCardLowerBound, eCardUpperBound)
		} else if eCardLowerBound > 0 {
			eventString += fmt.Sprintf("> %d time", eCardLowerBound)
		} else if eCardUpperBound > 0 {
			eventString += fmt.Sprintf("< %d time", eCardUpperBound)
		}
		eventString += ")"
	}
	return eventString
}

func headerString(funnelEvents []string,
	eCardLowerBound int, eCardUpperBound int, funnelConversionPercent float64,
	baseFunnelConversionPercent float64) string {

	var header string
	pLen := len(funnelEvents)
	if pLen < 2 {
		log.Error(fmt.Sprintf("Unexpected funnel. %s", funnelEvents))
		return header
	}
	impactEvent := funnelEvents[pLen-2]
	endEventString := eventStringWithConditions(funnelEvents[pLen-1], eCardLowerBound, eCardUpperBound)

	otherEventString := ""
	for i := 0; i < pLen-2; i++ {
		if i == 0 {
			otherEventString += "after"
		}
		otherEventString += fmt.Sprintf(" %s", funnelEvents[i])
		if i < pLen-3 {
			otherEventString += " and"
		}
	}
	conversionChangeString := " have"
	if funnelConversionPercent > baseFunnelConversionPercent {
		conversionMultiple := funnelConversionPercent / baseFunnelConversionPercent
		conversionChangeString += fmt.Sprintf(" %.1f times higher chance to", conversionMultiple)
	} else {
		conversionMultiple := baseFunnelConversionPercent / funnelConversionPercent
		conversionChangeString += fmt.Sprintf(" %.2f times lower chance to", conversionMultiple)
	}
	header = fmt.Sprintf("Users who have %s %s %s %s", impactEvent, otherEventString, conversionChangeString, endEventString)
	return header
}

func (pw *PatternWrapper) buildFunnelData(
	p *Pattern, endEvent string, eCardLowerBound int,
	eCardUpperBound int, isBaseFunnel bool) funnelNodeResults {
	pLen := len(p.EventNames)
	funnelData := funnelNodeResults{}
	var referenceFunnelCount uint

	for i := 0; i < pLen; i++ {
		var funnelSubsequencePerUserCount uint
		var found bool = true
		if i == pLen-1 {
			funnelSubsequencePerUserCount = p.GetOncePerUserCount(eCardLowerBound, eCardUpperBound)
		} else {
			funnelSubsequencePerUserCount, found = pw.GetPerUserCount(p.EventNames[:i+1], -1, -1)
		}
		if !found {
			log.Errorf(fmt.Sprintf(
				"Subsequence %s not as frequent as sequence %s",
				eventArrayToString(p.EventNames[:i+1]), ","), p.String())
			funnelSubsequencePerUserCount = p.GetOncePerUserCount(eCardLowerBound, eCardUpperBound)
		}
		if i == 0 {
			if (pLen == 1 && isBaseFunnel) || (pLen == 2 && !isBaseFunnel) {
				// Reference is total users.
				referenceFunnelCount = p.UserCount
				// If basefunnel has length 1 we prefix an initial node with all users for better comparision.
				node := funnelNodeResult{
					Data:  []float64{float64(referenceFunnelCount), 0.0},
					Event: fmt.Sprintf("AllActiveUsers (%d)", referenceFunnelCount),
				}
				funnelData = append(funnelData, node)
			} else {
				referenceFunnelCount = funnelSubsequencePerUserCount
			}
		}
		var eventString string
		if i == (pLen - 1) {
			// Conditions are implicitly only on end event currently.
			eventString = eventStringWithConditions(p.EventNames[i], eCardLowerBound, eCardUpperBound)
		} else {
			eventString = eventStringWithConditions(p.EventNames[i], -1, -1)
		}
		node := funnelNodeResult{
			Data:  []float64{float64(funnelSubsequencePerUserCount), float64(referenceFunnelCount - funnelSubsequencePerUserCount)},
			Event: fmt.Sprintf("%s (%d)", eventString, funnelSubsequencePerUserCount),
		}
		funnelData = append(funnelData, node)
	}
	funnelLength := len(funnelData)
	funnelConversionPercent := float64(funnelData[funnelLength-1].Data[0]*100.0) / funnelData[funnelLength-2].Data[0]
	if funnelConversionPercent > 0.1 {
		// Round it to nearest one digit.
		funnelConversionPercent = math.Round(funnelConversionPercent*10) / 10.0
	}
	funnelData[funnelLength-2].ConversionPercent = funnelConversionPercent
	return funnelData
}

func (pw *PatternWrapper) buildFactorResultsFromPatterns(
	patterns []*Pattern, endEvent string,
	eCardLowerBound int, eCardUpperBound int) PatternServiceGraphResults {
	results := PatternServiceGraphResults{Charts: []graphResult{}}
	endEventString := eventStringWithConditions(endEvent, eCardLowerBound, eCardUpperBound)
	// Dummy Line Chart.
	chart := graphResult{
		Type:   "line",
		Header: fmt.Sprintf("Average %s per month", endEventString),
		Labels: []string{"January", "February", "March", "April", "May", "June", "July"},
		Datasets: []map[string]interface{}{
			map[string]interface{}{
				"label": "Users with country:US",
				"data":  []float64{65, 59, 80, 81, 56, 55, 40},
			},
			map[string]interface{}{
				"label": "All Users",
				"data":  []float64{45, 50, 70, 101, 95, 80, 64},
			},
		},
	}
	results.Charts = append(results.Charts, chart)
	// Dummy Bar Chart.
	chart = graphResult{
		Type:   "bar",
		Header: fmt.Sprintf("Users with country US have 30%% higher average %s than others.", endEventString),
		Labels: []string{"All Users", "US", "India", "UK", "Australia", "Egypt", "Iran"},
		Datasets: []map[string]interface{}{
			map[string]interface{}{
				"label": fmt.Sprintf("Average %s", endEvent),
				"data":  []float64{65, 59, 80, 81, 56, 55, 40},
			},
			map[string]interface{}{
				"label": "All Users",
				"data":  []float64{45, 50, 70, 101, 95, 80, 64},
			},
		},
	}
	results.Charts = append(results.Charts, chart)
	// Actual funnel results.
	for _, p := range patterns {
		pLen := len(p.EventNames)
		if pLen == 1 {
			continue
		}

		// Skip (n - 1)st element for baseFunnel.
		baseFunnelEvents := append(append([]string(nil), p.EventNames[:pLen-2]...), p.EventNames[pLen-1:]...)
		var baseP *Pattern
		var ok bool
		if baseP, ok = pw.GetPattern(baseFunnelEvents); !ok {
			log.Errorf(fmt.Sprintf("Missing Base Funnel Pattern for %s", p.String()))
			continue
		}
		baseFunnelData := pw.buildFunnelData(
			baseP, endEvent, eCardLowerBound, eCardUpperBound, true)
		funnelData := pw.buildFunnelData(
			p, endEvent, eCardLowerBound, eCardUpperBound, false)

		baseFunnelLength := len(baseFunnelData)
		baseFunnelConversionPercent := baseFunnelData[baseFunnelLength-2].ConversionPercent
		funnelLength := len(funnelData)
		funnelConversionPercent := funnelData[funnelLength-2].ConversionPercent
		if funnelConversionPercent > baseFunnelConversionPercent {
			funnelData[funnelLength-2].NodeType = "positive"
		} else if funnelConversionPercent < baseFunnelConversionPercent {
			funnelData[funnelLength-2].NodeType = "negative"
		}

		chart = graphResult{
			Type: "funnel",
			Header: headerString(p.EventNames, eCardLowerBound,
				eCardUpperBound, funnelConversionPercent,
				baseFunnelConversionPercent),
			Datasets: []map[string]interface{}{
				map[string]interface{}{
					"base_funnel_data": baseFunnelData,
					"funnel_data":      funnelData,
				},
			},
		}
		results.Charts = append(results.Charts, chart)
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

func (ps *PatternService) Factor(projectId uint64, endEvent string,
	eCardLowerBound int, ecardUpperBound int) (PatternServiceGraphResults, error) {
	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return PatternServiceGraphResults{}, fmt.Errorf(fmt.Sprintf("No patterns for projectId:%d", projectId))
	}
	if endEvent == "" {
		return PatternServiceGraphResults{}, fmt.Errorf("Invalid Query")
	}

	iPatterns := []*Pattern{}
	if itree, err := BuildNewItree(endEvent, eCardLowerBound, ecardUpperBound, pw); err != nil {
		log.Error(err)
		return PatternServiceGraphResults{}, err
	} else {
		for _, node := range itree.Nodes {
			iPatterns = append(iPatterns, node.Pattern)
		}
	}
	results := pw.buildFactorResultsFromPatterns(
		iPatterns, endEvent, eCardLowerBound, ecardUpperBound)

	maxPatterns := 50
	if len(results.Charts) > maxPatterns {
		results.Charts = results.Charts[:maxPatterns]
	}

	return results, nil
}

func (ps *PatternService) FrequentPaths(projectId uint64, startEvent string,
	endEvent string, eCardLowerBound int, ecardUpperBound int) (PatternServiceGraphResults, error) {

	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return PatternServiceGraphResults{}, fmt.Errorf(fmt.Sprintf("No patterns for projectId:%d", projectId))
	}
	if startEvent == "" && endEvent == "" {
		return PatternServiceGraphResults{}, fmt.Errorf("Invalid Query")
	}
	matchPatterns := []*Pattern{}
	for _, p := range pw.patterns {
		if (startEvent == "" || strings.Compare(startEvent, p.EventNames[0]) == 0) &&
			(endEvent == "" || strings.Compare(endEvent, p.EventNames[len(p.EventNames)-1]) == 0) {
			matchPatterns = append(matchPatterns, p)
		}
	}

	// Sort in decreasing order of per user counts of the sequence.
	sort.SliceStable(matchPatterns,
		func(i, j int) bool {
			return (matchPatterns[i].OncePerUserCount > matchPatterns[j].OncePerUserCount)
		})
	maxPatterns := 50
	if len(matchPatterns) > maxPatterns {
		matchPatterns = matchPatterns[:maxPatterns]
	}

	results := pw.buildFactorResultsFromPatterns(
		matchPatterns, endEvent, eCardLowerBound, ecardUpperBound)

	return results, nil
}
