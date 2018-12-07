// pattern_service
package pattern

import (
	U "factors/util"
	"fmt"
	"math"
	"sort"

	log "github.com/sirupsen/logrus"
)

type PatternWrapper struct {
	patterns     []*Pattern
	pMap         map[string]*Pattern
	eventInfoMap *EventInfoMap
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

func NewPatternWrapper(patterns []*Pattern, eventInfoMap *EventInfoMap) *PatternWrapper {
	patternWrapper := PatternWrapper{
		patterns: patterns,
	}
	pMap := make(map[string]*Pattern)
	for _, p := range patterns {
		pMap[p.String()] = p
	}
	patternWrapper.pMap = pMap
	patternWrapper.eventInfoMap = eventInfoMap
	return &patternWrapper
}

func (pw *PatternWrapper) GetPerUserCount(eventNames []string,
	patternConstraints []EventConstraints) (uint, bool) {
	if p, ok := pw.pMap[eventArrayToString(eventNames)]; ok {
		count, err := p.GetOncePerUserCount(patternConstraints)
		if err == nil {
			return count, true
		}
	}
	return 0, false
}

func (pw *PatternWrapper) GetPattern(eventNames []string) (*Pattern, bool) {
	p, ok := pw.pMap[eventArrayToString(eventNames)]
	return p, ok
}

func eventStringWithConditions(eventName string, eventConstraints *EventConstraints) string {
	var eventString string = eventName
	if eventConstraints != nil {
		for _, c := range eventConstraints.NumericConstraints {
			if c.PropertyName == "" {
				continue
			}
			hasLowerBound := c.LowerBound > -math.MaxFloat64 && c.LowerBound < math.MaxFloat64
			hasUpperBound := c.UpperBound > -math.MaxFloat64 && c.UpperBound < math.MaxFloat64
			lowerBoundStr := ""
			if hasLowerBound {
				if c.LowerBound == float64(int64(c.LowerBound)) {
					lowerBoundStr = fmt.Sprintf("%d", int(c.LowerBound))
				} else {
					lowerBoundStr = fmt.Sprintf("%.2f", c.LowerBound)
				}
			}
			upperBoundStr := ""
			if hasUpperBound {
				if c.UpperBound == float64(int64(c.UpperBound)) {
					upperBoundStr = fmt.Sprintf("%d", int(c.UpperBound))
				} else {
					upperBoundStr = fmt.Sprintf("%.2f", c.UpperBound)
				}
			}
			midStr := ""
			if hasUpperBound && hasLowerBound {
				midNum := (c.LowerBound + c.UpperBound) / 2.0
				if midNum == float64(int64(midNum)) {
					midStr = fmt.Sprintf("%d", int(midNum))
				} else {
					midStr = fmt.Sprintf("%.2f", midNum)
				}
			}
			if hasLowerBound || hasUpperBound {
				eventString += " ("
				if hasLowerBound && hasUpperBound {
					if c.IsEquality {
						eventString += fmt.Sprintf("%s = %s",
							c.PropertyName, midStr)
					} else {
						eventString += fmt.Sprintf("%s < %s < %s",
							lowerBoundStr, c.PropertyName, upperBoundStr)
					}
				} else if hasLowerBound {
					eventString += fmt.Sprintf("%s > %s", c.PropertyName, lowerBoundStr)
				} else if hasUpperBound {
					eventString += fmt.Sprintf("%s < %s", c.PropertyName, upperBoundStr)
				}
				eventString += ")"
			}
		}
		for _, c := range eventConstraints.CategoricalConstraints {
			if c.PropertyName == "" {
				continue
			}
			eventString += fmt.Sprintf(" (%s is %s)", c.PropertyName, c.PropertyValue)
		}
	}
	return eventString
}

func headerString(funnelEvents []string,
	patternConstraints []EventConstraints, nodeType int,
	funnelConversionPercent float64, baseFunnelConversionPercent float64) string {
	var header string
	pLen := len(funnelEvents)
	if pLen < 2 {
		log.Error(fmt.Sprintf("Unexpected funnel. %s", funnelEvents))
		return header
	}
	if patternConstraints == nil {
		patternConstraints = make([]EventConstraints, len(funnelEvents))
	}
	var impactString string
	if nodeType == NODE_TYPE_SEQUENCE {
		impactString = funnelEvents[pLen-2]
	} else if nodeType == NODE_TYPE_PROPERTY {
		impactString = fmt.Sprintf("%s with %s", funnelEvents[pLen-2],
			eventStringWithConditions("", &patternConstraints[pLen-2]))
	}

	endEventString := eventStringWithConditions(funnelEvents[pLen-1], &patternConstraints[pLen-1])

	otherEventString := ""
	for i := 0; i < pLen-2; i++ {
		if i == 0 {
			otherEventString += "after"
		}
		otherEventString +=  fmt.Sprintf(" %s",
            eventStringWithConditions(funnelEvents[i], &patternConstraints[i]))
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
	header = fmt.Sprintf("Users who have %s %s %s %s", impactString, otherEventString, conversionChangeString, endEventString)
	return header
}

func (pw *PatternWrapper) buildFunnelData(
	p *Pattern, patternConstraints []EventConstraints,
	nodeType int, isBaseFunnel bool) funnelNodeResults {
	pLen := len(p.EventNames)
	funnelData := funnelNodeResults{}
	var referenceFunnelCount uint

	for i := 0; i < pLen; i++ {
		var funnelSubsequencePerUserCount uint
		var found = true
		if i == pLen-1 {
			funnelSubsequencePerUserCount, _ = p.GetOncePerUserCount(patternConstraints)
		} else {
			var funnelConstraints []EventConstraints
			if patternConstraints != nil {
				funnelConstraints = patternConstraints[:i+1]
			}
			funnelSubsequencePerUserCount, found = pw.GetPerUserCount(
				p.EventNames[:i+1], funnelConstraints)
		}
		if !found {
			log.Errorf(fmt.Sprintf(
				"Subsequence %s not as frequent as sequence %s",
				eventArrayToString(p.EventNames[:i+1]), ","), p.String())
			funnelSubsequencePerUserCount, _ = p.GetOncePerUserCount(patternConstraints)
		}
		if i == 0 {
			if (nodeType == NODE_TYPE_SEQUENCE) && ((pLen == 1 && isBaseFunnel) || (pLen == 2 && !isBaseFunnel)) {
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
		if patternConstraints != nil {
			eventString = eventStringWithConditions(p.EventNames[i], &patternConstraints[i])
		} else {
			eventString = eventStringWithConditions(p.EventNames[i], nil)
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

func (pw *PatternWrapper) buildFactorResultsFromPatterns(nodes []*ItreeNode) PatternServiceGraphResults {
	results := PatternServiceGraphResults{Charts: []graphResult{}}
	endEventString := "dummyEvent"
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
				"label": fmt.Sprintf("Average %s", endEventString),
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

	for _, node := range nodes {
		p := node.Pattern
		patternConstraints := node.PatternConstraints
		pLen := len(p.EventNames)
		baseFunnelEvents := []string{}
		baseFunnelConstraints := []EventConstraints{}

		if node.NodeType == NODE_TYPE_SEQUENCE {
			// Skip (n - 1)st element for baseFunnel.
			baseFunnelEvents = append(append([]string(nil), p.EventNames[:pLen-2]...), p.EventNames[pLen-1:]...)
			if patternConstraints != nil {
				baseFunnelConstraints = append(append([]EventConstraints(nil), patternConstraints[:pLen-2]...), patternConstraints[pLen-1:]...)
			}
		} else if node.NodeType == NODE_TYPE_PROPERTY {
			// Skip (n - 1)st constraint for baseFunnel.
			baseFunnelEvents = append(append([]string(nil), p.EventNames...))
			if patternConstraints != nil {
				baseFunnelConstraints = append(append([]EventConstraints(nil), patternConstraints...))
			} else {
				baseFunnelConstraints = make([]EventConstraints, pLen)
			}
			baseFunnelConstraints[pLen-2] = EventConstraints{}
		}
		var baseP *Pattern
		var ok bool
		if baseP, ok = pw.GetPattern(baseFunnelEvents); !ok {
			log.Errorf(fmt.Sprintf("Missing Base Funnel Pattern for %s", p.String()))
			continue
		}
		baseFunnelData := pw.buildFunnelData(baseP, baseFunnelConstraints, node.NodeType, true)
		funnelData := pw.buildFunnelData(p, patternConstraints, node.NodeType, false)

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
			Header: headerString(p.EventNames, patternConstraints, node.NodeType,
				funnelConversionPercent, baseFunnelConversionPercent),
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

func NewPatternService(
	patternsMap map[uint64][]*Pattern,
	projectEventInfoMap map[uint64]*EventInfoMap) (*PatternService, error) {

	patternService := PatternService{patternsMap: map[uint64]*PatternWrapper{}}

	for projectId, patterns := range patternsMap {
		eventInfoMap, _ := projectEventInfoMap[projectId]
		patternWrapper := NewPatternWrapper(patterns, eventInfoMap)
		patternService.patternsMap[projectId] = patternWrapper
	}
	return &patternService, nil
}

func (ps *PatternService) Factor(projectId uint64, startEvent string,
	startEventConstraints *EventConstraints, endEvent string,
	endEventConstraints *EventConstraints) (PatternServiceGraphResults, error) {
	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return PatternServiceGraphResults{}, fmt.Errorf(fmt.Sprintf("No patterns for projectId:%d", projectId))
	}
	if endEvent == "" {
		return PatternServiceGraphResults{}, fmt.Errorf("Invalid Query")
	}

	iPatternNodes := []*ItreeNode{}
	if itree, err := BuildNewItree(startEvent, startEventConstraints,
		endEvent, endEventConstraints, pw); err != nil {
		log.Error(err)
		return PatternServiceGraphResults{}, err
	} else {
		for _, node := range itree.Nodes {
			if node.NodeType == NODE_TYPE_ROOT {
				// Root node.
				continue
			}
			iPatternNodes = append(iPatternNodes, node)
		}
	}

	// Rerank iPatternNodes in descending order of ranked scores.
	sort.SliceStable(iPatternNodes,
		func(i, j int) bool {
			// GiniDrop * parentPatternFrequency is the ranking score for the node.
			scoreI := iPatternNodes[i].GiniDrop * iPatternNodes[i].Fpp
			scoreJ := iPatternNodes[j].GiniDrop * iPatternNodes[j].Fpp
			return (scoreI > scoreJ)
		})

	results := pw.buildFactorResultsFromPatterns(iPatternNodes)

	maxPatterns := 50
	if len(results.Charts) > maxPatterns {
		results.Charts = results.Charts[:maxPatterns]
	}

	return results, nil
}

func (ps *PatternService) GetSeenEventProperties(projectId uint64, eventName string) (map[string][]string, error) {
	// Initialize results.
	results := make(map[string][]string)
	numericalProperties := []string{}
	for _, dnp := range U.DEFAULT_NUMERIC_EVENT_PROPERTIES {
		numericalProperties = append(numericalProperties, dnp)
	}
	categoricalProperties := []string{}
	results["numerical"] = numericalProperties
	results["categorical"] = categoricalProperties
	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return results, fmt.Errorf(fmt.Sprintf("No data for projectId:%d", projectId))
	}
	if eventName == "" {
		return results, fmt.Errorf("Invalid Query")
	}
	if pw.eventInfoMap == nil {
		return results, nil
	}
	eventInfo, _ := (*pw.eventInfoMap)[eventName]
	for nprop, _ := range eventInfo.NumericPropertyKeys {
		numericalProperties = append(numericalProperties, nprop)
	}
	for cprop, _ := range eventInfo.CategoricalPropertyKeyValues {
		categoricalProperties = append(categoricalProperties, cprop)
	}
	results["numerical"] = numericalProperties
	results["categorical"] = categoricalProperties
	return results, nil
}

func (ps *PatternService) GetSeenEventPropertyValues(
	projectId uint64, eventName string, propertyName string) ([]string, error) {
	// Initialize results.
	results := []string{}
	if eventName == "" {
		return results, fmt.Errorf("Invalid Query")
	}
	if propertyName == "" {
		return results, fmt.Errorf("Invalid Query")
	}

	pw, ok := ps.patternsMap[projectId]
	if !ok {
		return results, fmt.Errorf(fmt.Sprintf("No data for projectId:%d", projectId))
	}
	if pw.eventInfoMap == nil {
		return results, nil
	}
	eventInfo, _ := (*pw.eventInfoMap)[eventName]
	propValuesMap, ok := eventInfo.CategoricalPropertyKeyValues[propertyName]
	if !ok {
		log.WithFields(log.Fields{
			"eventName": eventName, "propertyName": propertyName,
			"projectId": projectId}).Info("Property not found.")
		return results, nil
	}
	for k, _ := range propValuesMap {
		results = append(results, k)
	}
	return results, nil
}
