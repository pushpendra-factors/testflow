// pattern_service_wrapper
package pattern_service_wrapper

import (
	P "factors/pattern"
	PC "factors/pattern_client"
	U "factors/util"
	"fmt"
	"math"
	"reflect"
	"sort"

	log "github.com/sirupsen/logrus"
)

// Fetches information from pattern server and operations on patterns in local cache.
type PatternServiceWrapperInterface interface {
	GetUserAndEventsInfo() *P.UserAndEventsInfo
	GetPerUserCount(eventNames []string,
		patternConstraints []P.EventConstraints) (uint, bool)
	GetPattern(eventNames []string) *P.Pattern
	GetAllPatterns(
		startEvent string, endEvent string) ([]*P.Pattern, error)
}

type PatternServiceWrapper struct {
	projectId         uint64
	modelId           uint64
	pMap              map[string]*P.Pattern
	userAndEventsInfo *P.UserAndEventsInfo
}

type funnelNodeResult struct {
	Data              []float64           `json:"data"`
	Event             string              `json:"event"`
	EventName         string              `json:"-"`
	Constraints       *P.EventConstraints `json:"-"`
	NodeType          string              `json:"node_type"`
	ConversionPercent float64             `json:"conversion_percent"`
}
type funnelNodeResults []funnelNodeResult
type graphResult struct {
	Type     string                   `json:"type"`
	Header   string                   `json:"header"`
	Labels   []string                 `json:"labels"`
	Datasets []map[string]interface{} `json:"datasets"`
}
type FactorGraphResults struct {
	Charts []graphResult `json:"charts"`
}

func (pw *PatternServiceWrapper) GetUserAndEventsInfo() *P.UserAndEventsInfo {
	return pw.userAndEventsInfo
}

func (pw *PatternServiceWrapper) GetPerUserCount(eventNames []string,
	patternConstraints []P.EventConstraints) (uint, bool) {
	pattern := pw.GetPattern(eventNames)
	if pattern != nil {
		count, err := pattern.GetOncePerUserCount(patternConstraints)
		if err == nil {
			return count, true
		}
	}
	return 0, false
}

func (pw *PatternServiceWrapper) GetPattern(eventNames []string) *P.Pattern {
	var pattern *P.Pattern = nil
	var found bool
	eventsHash := P.EventArrayToString(eventNames)
	if pattern, found = pw.pMap[eventsHash]; !found {
		// Fetch from server.
		patterns, err := PC.GetPatterns(pw.projectId, pw.modelId, [][]string{eventNames})
		if err == nil && len(patterns) == 1 && P.EventArrayToString(patterns[0].EventNames) == eventsHash {
			pattern = patterns[0]
			// Add it to cache.
			pw.pMap[eventsHash] = pattern
		} else {
			log.WithFields(log.Fields{
				"error": err, "modelId": pw.modelId,
				"projectId": pw.projectId, "eventNames": eventNames,
				"returned Patterns": patterns}).Error(
				"Get Patterns failed.")
		}
	}
	return pattern
}

func (pw *PatternServiceWrapper) GetAllPatterns(
	startEvent string, endEvent string) ([]*P.Pattern, error) {
	// Fetch from server.
	patterns, err := PC.GetAllPatterns(pw.projectId, pw.modelId, startEvent, endEvent)
	// Add it to cache.
	for _, p := range patterns {
		pw.pMap[P.EventArrayToString(p.EventNames)] = p
	}
	return patterns, err
}

func numericConstraintString(nC P.NumericConstraint) string {
	constraintStr := ""
	hasLowerBound := nC.LowerBound > -math.MaxFloat64 && nC.LowerBound < math.MaxFloat64
	hasUpperBound := nC.UpperBound > -math.MaxFloat64 && nC.UpperBound < math.MaxFloat64
	lowerBoundStr := ""
	if hasLowerBound {
		if nC.LowerBound == float64(int64(nC.LowerBound)) {
			lowerBoundStr = fmt.Sprintf("%d", int(nC.LowerBound))
		} else {
			lowerBoundStr = fmt.Sprintf("%.2f", nC.LowerBound)
		}
	}
	upperBoundStr := ""
	if hasUpperBound {
		if nC.UpperBound == float64(int64(nC.UpperBound)) {
			upperBoundStr = fmt.Sprintf("%d", int(nC.UpperBound))
		} else {
			upperBoundStr = fmt.Sprintf("%.2f", nC.UpperBound)
		}
	}
	midStr := ""
	if hasUpperBound && hasLowerBound {
		midNum := (nC.LowerBound + nC.UpperBound) / 2.0
		if midNum == float64(int64(midNum)) {
			midStr = fmt.Sprintf("%d", int(midNum))
		} else {
			midStr = fmt.Sprintf("%.2f", midNum)
		}
	}
	if hasLowerBound || hasUpperBound {
		constraintStr += " ("
		if hasLowerBound && hasUpperBound {
			if nC.IsEquality {
				constraintStr += fmt.Sprintf("%s = %s",
					nC.PropertyName, midStr)
			} else {
				constraintStr += fmt.Sprintf("%s < %s < %s",
					lowerBoundStr, nC.PropertyName, upperBoundStr)
			}
		} else if hasLowerBound {
			constraintStr += fmt.Sprintf("%s > %s", nC.PropertyName, lowerBoundStr)
		} else if hasUpperBound {
			constraintStr += fmt.Sprintf("%s < %s", nC.PropertyName, upperBoundStr)
		}
		constraintStr += ")"
	}
	return constraintStr
}

func eventStringWithConditions(eventName string, eventConstraints *P.EventConstraints) string {
	var eventString string = eventName
	if eventConstraints != nil {
		for _, nC := range eventConstraints.EPNumericConstraints {
			if nC.PropertyName == "" {
				continue
			}
			eventString += numericConstraintString(nC)
		}
		for _, c := range eventConstraints.EPCategoricalConstraints {
			if c.PropertyName == "" {
				continue
			}
			eventString += fmt.Sprintf(" (%s is %s)", c.PropertyName, c.PropertyValue)
		}
		for _, nC := range eventConstraints.UPNumericConstraints {
			if nC.PropertyName == "" {
				continue
			}
			eventString += numericConstraintString(nC)
		}
		for _, c := range eventConstraints.UPCategoricalConstraints {
			if c.PropertyName == "" {
				continue
			}
			eventString += fmt.Sprintf(" (%s is %s)", c.PropertyName, c.PropertyValue)
		}
	}
	return eventString
}

func headerString(funnelData funnelNodeResults, nodeType int,
	funnelConversionPercent float64, baseFunnelConversionPercent float64) string {
	var header string
	pLen := len(funnelData)
	if pLen < 2 {
		log.Error(fmt.Sprintf("Unexpected! Funnel: %s ", funnelData))
		return header
	}
	var impactString string
	// Impact event.
	if nodeType == NODE_TYPE_SEQUENCE {
		impactString = funnelData[pLen-2].EventName
	} else if nodeType == NODE_TYPE_EVENT_PROPERTY || nodeType == NODE_TYPE_USER_PROPERTY {
		impactString = fmt.Sprintf("%s with %s", funnelData[pLen-2].EventName,
			eventStringWithConditions("", funnelData[pLen-2].Constraints))
	}

	endEventString := eventStringWithConditions(
		funnelData[pLen-1].EventName, funnelData[pLen-1].Constraints)

	otherEventString := ""
	for i := 0; i < pLen-2; i++ {
		if i == 0 {
			otherEventString += "after"
		}
		otherEventString += fmt.Sprintf(" %s",
			eventStringWithConditions(funnelData[i].EventName, funnelData[i].Constraints))
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

func buildFunnelData(
	p *P.Pattern, patternConstraints []P.EventConstraints,
	nodeType int, addedConstraint P.EventConstraints,
	isBaseFunnel bool, pw PatternServiceWrapperInterface) funnelNodeResults {
	pLen := len(p.EventNames)
	funnelData := funnelNodeResults{}
	var referenceFunnelCount uint

	for i := 0; i < pLen; i++ {
		var funnelSubsequencePerUserCount uint
		var found = true
		if i == pLen-1 {
			funnelSubsequencePerUserCount, _ = p.GetOncePerUserCount(patternConstraints)
		} else {
			var funnelConstraints []P.EventConstraints
			if patternConstraints != nil {
				funnelConstraints = patternConstraints[:i+1]
			}
			funnelSubsequencePerUserCount, found = pw.GetPerUserCount(
				p.EventNames[:i+1], funnelConstraints)
		}
		if !found {
			log.Errorf(fmt.Sprintf(
				"Subsequence %s not as frequent as sequence %s",
				P.EventArrayToString(p.EventNames[:i+1]), ","), p.String())
			funnelSubsequencePerUserCount, _ = p.GetOncePerUserCount(patternConstraints)
		}
		if i == 0 {
			if (nodeType == NODE_TYPE_SEQUENCE) && ((pLen == 1 && isBaseFunnel) || (pLen == 2 && !isBaseFunnel)) {
				// Reference is total users.
				referenceFunnelCount = p.UserCount
				// If basefunnel has length 1 we prefix an initial node with all users for better comparision.
				node := funnelNodeResult{
					Data:        []float64{float64(referenceFunnelCount), 0.0},
					Event:       fmt.Sprintf("%s (%d)", U.SEN_ALL_ACTIVE_USERS, referenceFunnelCount),
					EventName:   U.SEN_ALL_ACTIVE_USERS,
					Constraints: nil,
				}
				funnelData = append(funnelData, node)
			} else if nodeType == NODE_TYPE_USER_PROPERTY && pLen == 1 {
				// If length 1 the addedConstraint is part of patternConstraints.
				// Remove it from patternConstraint. addedConstraint is used as a constraint for
				// AllActiveUsers
				for j, pNConstraint := range patternConstraints[i].UPNumericConstraints {
					for _, aNConstraint := range addedConstraint.UPNumericConstraints {
						if reflect.DeepEqual(pNConstraint, aNConstraint) {
							patternConstraints[i].UPNumericConstraints[j] = P.NumericConstraint{}
						}
					}
				}
				for j, pCConstraint := range patternConstraints[i].UPCategoricalConstraints {
					for _, aCConstraint := range addedConstraint.UPCategoricalConstraints {
						if reflect.DeepEqual(pCConstraint, aCConstraint) {
							patternConstraints[i].UPCategoricalConstraints[j] = P.CategoricalConstraint{}
						}
					}
				}
				node := funnelNodeResult{}
				if isBaseFunnel {
					referenceFunnelCount, _ = pw.GetPerUserCount(
						[]string{U.SEN_ALL_ACTIVE_USERS}, nil)
					node.Data = []float64{float64(referenceFunnelCount), 0.0}
					node.Event = fmt.Sprintf("%s (%d)", eventStringWithConditions(U.SEN_ALL_ACTIVE_USERS,
						nil), referenceFunnelCount)
					node.EventName = U.SEN_ALL_ACTIVE_USERS
					node.Constraints = nil
				} else {
					tmpConstraints := []P.EventConstraints{addedConstraint}
					referenceFunnelCount, _ = pw.GetPerUserCount(
						[]string{U.SEN_ALL_ACTIVE_USERS}, tmpConstraints)
					node.Data = []float64{float64(referenceFunnelCount), 0.0}
					node.Event = fmt.Sprintf("%s (%d)", eventStringWithConditions(U.SEN_ALL_ACTIVE_USERS,
						&addedConstraint), referenceFunnelCount)
					node.EventName = U.SEN_ALL_ACTIVE_USERS
					node.Constraints = &addedConstraint
				}
				funnelData = append(funnelData, node)
			} else {
				referenceFunnelCount = funnelSubsequencePerUserCount
			}
		}
		var eventString string
		var eventConstraints *P.EventConstraints
		if patternConstraints != nil {
			eventString = eventStringWithConditions(p.EventNames[i], &patternConstraints[i])
			eventConstraints = &patternConstraints[i]
		} else {
			eventString = eventStringWithConditions(p.EventNames[i], nil)
			eventConstraints = nil
		}

		node := funnelNodeResult{
			Data:        []float64{float64(funnelSubsequencePerUserCount), float64(referenceFunnelCount - funnelSubsequencePerUserCount)},
			Event:       fmt.Sprintf("%s (%d)", eventString, funnelSubsequencePerUserCount),
			EventName:   p.EventNames[i],
			Constraints: eventConstraints,
		}
		funnelData = append(funnelData, node)
	}
	funnelLength := len(funnelData)
	funnelConversionPercent := 0.0
	if funnelData[funnelLength-2].Data[0] > 0.0 {
		funnelConversionPercent = float64(funnelData[funnelLength-1].Data[0]*100.0) / funnelData[funnelLength-2].Data[0]
	}
	if funnelConversionPercent > 0.1 {
		// Round it to nearest one digit.
		funnelConversionPercent = math.Round(funnelConversionPercent*10) / 10.0
	}
	funnelData[funnelLength-2].ConversionPercent = funnelConversionPercent
	return funnelData
}

func buildFactorResultsFromPatterns(nodes []*ItreeNode, pw PatternServiceWrapperInterface) FactorGraphResults {
	results := FactorGraphResults{Charts: []graphResult{}}
	/*
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
	*/
	// Actual funnel results.

	for _, node := range nodes {
		p := node.Pattern
		patternConstraints := node.PatternConstraints
		pLen := len(p.EventNames)
		baseFunnelEvents := []string{}
		baseFunnelConstraints := []P.EventConstraints{}

		if node.NodeType == NODE_TYPE_SEQUENCE {
			// Skip (n - 1)st element for baseFunnel.
			baseFunnelEvents = append(append([]string(nil), p.EventNames[:pLen-2]...), p.EventNames[pLen-1:]...)
			if patternConstraints != nil {
				baseFunnelConstraints = append(append([]P.EventConstraints(nil), patternConstraints[:pLen-2]...), patternConstraints[pLen-1:]...)
			}
		} else if node.NodeType == NODE_TYPE_EVENT_PROPERTY || node.NodeType == NODE_TYPE_USER_PROPERTY {
			// Skip (n - 1)st constraint for baseFunnel.
			baseFunnelEvents = append(append([]string(nil), p.EventNames...))
			if patternConstraints != nil {
				baseFunnelConstraints = append(append([]P.EventConstraints(nil), patternConstraints...))
			} else {
				baseFunnelConstraints = make([]P.EventConstraints, pLen)
			}
			if pLen > 1 {
				baseFunnelConstraints[pLen-2] = P.EventConstraints{}
			} else {
				// Skip constraints for last node if Len1.
				baseFunnelConstraints[pLen-1] = P.EventConstraints{}
			}
		}
		var baseP *P.Pattern
		if baseP = pw.GetPattern(baseFunnelEvents); baseP == nil {
			log.Errorf(fmt.Sprintf("Missing Base Funnel Pattern for %s", p.String()))
			continue
		}
		baseFunnelData := buildFunnelData(baseP, baseFunnelConstraints, node.NodeType, node.AddedConstraint, true, pw)
		funnelData := buildFunnelData(p, patternConstraints, node.NodeType, node.AddedConstraint, false, pw)

		baseFunnelLength := len(baseFunnelData)
		baseFunnelConversionPercent := baseFunnelData[baseFunnelLength-2].ConversionPercent
		funnelLength := len(funnelData)
		funnelConversionPercent := funnelData[funnelLength-2].ConversionPercent
		if funnelConversionPercent > baseFunnelConversionPercent {
			funnelData[funnelLength-2].NodeType = "positive"
		} else if funnelConversionPercent < baseFunnelConversionPercent {
			funnelData[funnelLength-2].NodeType = "negative"
		}

		chart := graphResult{
			Type: "funnel",
			Header: headerString(funnelData, node.NodeType,
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

func NewPatternServiceWrapper(projectId uint64, modelId uint64) (*PatternServiceWrapper, error) {
	userAndEventsInfo, respModelId, err := PC.GetUserAndEventsInfo(projectId, modelId)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "projectId": projectId}).Error(
			"GetUserAndEventsInfo failed")
		return nil, err
	}

	pMap := make(map[string]*P.Pattern)
	patternServiceWrapper := PatternServiceWrapper{
		projectId:         projectId,
		modelId:           respModelId,
		userAndEventsInfo: userAndEventsInfo,
		pMap:              pMap,
	}

	return &patternServiceWrapper, nil
}

func Factor(projectId uint64, startEvent string,
	startEventConstraints *P.EventConstraints, endEvent string,
	endEventConstraints *P.EventConstraints,
	pw PatternServiceWrapperInterface) (FactorGraphResults, error) {
	iPatternNodes := []*ItreeNode{}
	if itree, err := BuildNewItree(startEvent, startEventConstraints,
		endEvent, endEventConstraints, pw); err != nil {
		log.Error(err)
		return FactorGraphResults{}, err
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

	results := buildFactorResultsFromPatterns(iPatternNodes, pw)

	maxPatterns := 50
	if len(results.Charts) > maxPatterns {
		results.Charts = results.Charts[:maxPatterns]
	}

	return results, nil
}
