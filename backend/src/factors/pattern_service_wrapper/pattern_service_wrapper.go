// pattern_service_wrapper
package pattern_service_wrapper

import (
	"encoding/json"
	P "factors/pattern"
	PC "factors/pattern_client"
	U "factors/util"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

// DeepCopy deepcopies a to b using json marshaling
func deepCopy(a, b interface{}) {
	byt, _ := json.Marshal(a)
	json.Unmarshal(byt, b)
}

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
	Type         string                   `json:"type"`
	Header       string                   `json:"header"`
	Explanations []string                 `json:"explanations"`
	Labels       []string                 `json:"labels"`
	Datasets     []map[string]interface{} `json:"datasets"`
	XLabel       string                   `json:"x_label"`
	YLabel       string                   `json:"y_label"`
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
		constraintStr += " with "
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
	}
	return constraintStr
}

func eventStringWithConditions(eventName string, eventConstraints *P.EventConstraints) string {
	var eventString string = eventName
	if eventString == U.SEN_ALL_ACTIVE_USERS {
		eventString = U.SEN_ALL_ACTIVE_USERS_DISPLAY_STRING
	}
	var seenProperty bool = false
	if eventConstraints != nil {
		for _, nC := range eventConstraints.EPNumericConstraints {
			if nC.PropertyName == "" {
				continue
			}
			if seenProperty {
				eventString += " and"
			}
			eventString += numericConstraintString(nC)
			seenProperty = true
		}
		for _, c := range eventConstraints.EPCategoricalConstraints {
			if c.PropertyName == "" {
				continue
			}
			if seenProperty {
				eventString += " and"
			}
			eventString += fmt.Sprintf(" with %s equals %s", c.PropertyName, c.PropertyValue)
			seenProperty = true
		}
		for _, nC := range eventConstraints.UPNumericConstraints {
			if nC.PropertyName == "" {
				continue
			}
			if seenProperty {
				eventString += " and"
			}
			eventString += numericConstraintString(nC)
			seenProperty = true
		}
		for _, c := range eventConstraints.UPCategoricalConstraints {
			if c.PropertyName == "" {
				continue
			}
			if seenProperty {
				eventString += " and"
			}
			eventString += fmt.Sprintf(" with %s equals %s", c.PropertyName, c.PropertyValue)
			seenProperty = true
		}
	}
	return eventString
}

func headerString(funnelData funnelNodeResults, nodeType int,
	funnelConversionPercent float64, baseFunnelConversionPercent float64) (string, []string) {
	var header string
	pLen := len(funnelData)
	if pLen < 2 {
		log.Error(fmt.Sprintf("Unexpected! Funnel: %s ", funnelData))
		return header, []string{}
	}
	var impactString string
	// Impact event.
	if nodeType == NODE_TYPE_SEQUENCE {
		impactString = fmt.Sprintf("who have %s", funnelData[pLen-2].EventName)
	} else if nodeType == NODE_TYPE_EVENT_PROPERTY || nodeType == NODE_TYPE_USER_PROPERTY {
		if funnelData[pLen-2].EventName == U.SEN_ALL_ACTIVE_USERS {
			impactString = fmt.Sprintf("with %s", eventStringWithConditions("", funnelData[pLen-2].Constraints))
		} else {
			impactString = fmt.Sprintf("who have %s with %s", funnelData[pLen-2].EventName,
				eventStringWithConditions("", funnelData[pLen-2].Constraints))
		}
	}

	endEventString := eventStringWithConditions(
		funnelData[pLen-1].EventName, funnelData[pLen-1].Constraints)

	otherEventString := ""
	for i := 0; i < pLen-2; i++ {
		if i == 0 && funnelData[i].EventName == U.SEN_ALL_ACTIVE_USERS {
			continue
		}
		if otherEventString == "" {
			otherEventString += " after"
		}
		otherEventString += fmt.Sprintf(" %s",
			eventStringWithConditions(funnelData[i].EventName, funnelData[i].Constraints))
		if i < pLen-3 {
			otherEventString += " and"
		}
	}
	conversionChangeString := ", have"
	if funnelConversionPercent > baseFunnelConversionPercent {
		conversionMultiple := funnelConversionPercent / baseFunnelConversionPercent
		conversionChangeString += fmt.Sprintf(" %.1f times higher chance to", conversionMultiple)
	} else {
		conversionMultiple := baseFunnelConversionPercent / funnelConversionPercent
		conversionChangeString += fmt.Sprintf(" %.2f times lower chance to", conversionMultiple)
	}
	header = fmt.Sprintf("Users %s%s%s %s.", impactString, otherEventString, conversionChangeString, endEventString)
	return header, []string{}
}

func barGraphHeaderString(
	patternEvents []string, patternConstraints []P.EventConstraints,
	propertyName string, propertyValues []string,
	patternUsersLabel string,
	patternPercentages map[string]float64,
	ruleUsersLabel string,
	rulePercentages map[string]float64,
	isIncrement bool) (string, []string) {
	pLen := len(patternEvents)
	if pLen < 1 {
		log.Error(fmt.Sprintf("Unexpected! Pattern: %s ", patternEvents))
		return "", []string{}
	}
	var impactString string
	propertyValuesString := strings.Join(propertyValues, ", ")
	if len(propertyValuesString) > 100 {
		propertyValuesString = propertyValuesString[:100] + "..."
	}
	if pLen == 1 {
		impactString = fmt.Sprintf("with %s in %s", propertyName, propertyValuesString)
	} else {
		impactString = fmt.Sprintf("who have %s with %s in %s", patternEvents[pLen-2],
			propertyName, propertyValuesString)
	}

	endEventString := eventStringWithConditions(
		patternEvents[pLen-1], &patternConstraints[pLen-1])

	otherEventString := ""
	for i := 0; i < pLen-2; i++ {
		if i == 0 {
			otherEventString += " after"
		}
		otherEventString += fmt.Sprintf(" %s",
			eventStringWithConditions(patternEvents[i], &patternConstraints[i]))
		if i < pLen-3 {
			otherEventString += " and"
		}
	}
	conversionChangeString := ", have"
	if isIncrement {
		conversionChangeString += " a higher chance to"
	} else {
		conversionChangeString += " a lower chance to"
	}
	header := fmt.Sprintf("Users %s%s%s %s.", impactString, otherEventString, conversionChangeString, endEventString)

	totalPatternPercentage := 0.0
	totalRulePercentage := 0.0
	for _, v := range propertyValues {
		if pp, ok := patternPercentages[v]; ok {
			totalPatternPercentage += pp
		}
		if rp, ok := rulePercentages[v]; ok {
			totalRulePercentage += rp
		}
	}
	headerExplanation := []string{}
	headerExplanation = append(
		headerExplanation,
		fmt.Sprintf(
			"%0.1f%% of %s, have these %s.",
			totalRulePercentage,
			ruleUsersLabel,
			propertyName,
		))

	headerExplanation = append(
		headerExplanation,
		fmt.Sprintf(
			"%0.1f%% of %s, have these %s.",
			totalPatternPercentage,
			patternUsersLabel,
			propertyName,
		))

	//log.WithFields(log.Fields{"events": patternEvents,
	//	"patternConstraints": patternConstraints,
	//	"endEventString":     endEventString, "pLen": pLen}).Debug("Graph results.")
	return header, headerExplanation
}

func buildFunnelData(
	funnelEvents []string, funnelConstraints []P.EventConstraints,
	node *ItreeNode, isBaseFunnel bool,
	pw PatternServiceWrapperInterface) funnelNodeResults {
	pLen := len(funnelEvents)
	funnelData := funnelNodeResults{}
	var referenceFunnelCount uint

	for i := 0; i < pLen; i++ {
		var funnelSubsequencePerUserCount uint
		var found = true
		if i == pLen-2 {
			// Pick the counts from node.
			if isBaseFunnel {
				funnelSubsequencePerUserCount = uint(node.Fpp)
			} else {
				funnelSubsequencePerUserCount = uint(node.Fcp)
			}
		} else if i == pLen-1 {
			// Pick the counts from node.
			if isBaseFunnel {
				funnelSubsequencePerUserCount = uint(node.Fpr)
			} else {
				funnelSubsequencePerUserCount = uint(node.Fcr)
			}
		} else {
			funnelSubsequencePerUserCount, found = pw.GetPerUserCount(
				funnelEvents[:i+1], funnelConstraints[:i+1])
		}
		if !found {
			log.Errorf(fmt.Sprintf(
				"Subsequence %s not as frequent as sequence %s",
				P.EventArrayToString(funnelEvents[:i+1]), ","), funnelEvents)
			funnelSubsequencePerUserCount, _ = pw.GetPerUserCount(funnelEvents, funnelConstraints)
		}
		eventString := eventStringWithConditions(funnelEvents[i], &funnelConstraints[i])
		if i == 0 {
			referenceFunnelCount = funnelSubsequencePerUserCount
		}
		node := funnelNodeResult{
			Data:        []float64{float64(funnelSubsequencePerUserCount), float64(referenceFunnelCount - funnelSubsequencePerUserCount)},
			Event:       fmt.Sprintf("%s (%d)", eventString, funnelSubsequencePerUserCount),
			EventName:   funnelEvents[i],
			Constraints: &funnelConstraints[i],
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

func buildFunnelFormats(node *ItreeNode) (
	[]string, []P.EventConstraints, []string, []P.EventConstraints) {
	var funnelConstraints []P.EventConstraints
	// https://stackoverflow.com/questions/46790190/quicker-way-to-deepcopy-objects-in-golang
	funnelEvents := append(make([]string, 0, len(node.Pattern.EventNames)), node.Pattern.EventNames...)
	if node.PatternConstraints != nil {
		funnelConstraints = make([]P.EventConstraints, len(node.PatternConstraints))
		deepCopy(&node.PatternConstraints, &funnelConstraints)
	} else {
		funnelConstraints = make([]P.EventConstraints, len(funnelEvents))
	}
	funnelConstraints = append([]P.EventConstraints(nil), funnelConstraints...)
	var baseFunnelEvents []string
	var baseFunnelConstraints []P.EventConstraints
	if node.NodeType == NODE_TYPE_SEQUENCE {
		pLen := len(funnelEvents)
		if pLen == 2 {
			// Prepend AllActiveUsers Event at the begining for readability.
			funnelEvents = append([]string{U.SEN_ALL_ACTIVE_USERS}, funnelEvents...)
			funnelConstraints = append([]P.EventConstraints{P.EventConstraints{}}, funnelConstraints...)
			pLen++
		}
		// Skip (n - 1)st element and constraint for baseFunnel.
		baseFunnelEvents = append(append([]string(nil), funnelEvents[:pLen-2]...), funnelEvents[pLen-1:]...)
		baseFunnelConstraints = append(append([]P.EventConstraints(nil), funnelConstraints[:pLen-2]...), funnelConstraints[pLen-1:]...)
	} else if node.NodeType == NODE_TYPE_EVENT_PROPERTY {
		pLen := len(funnelEvents)
		// Base funnel events are the same.
		baseFunnelEvents = append(make([]string, 0, len(funnelEvents)), funnelEvents...)
		baseFunnelConstraints = make([]P.EventConstraints, len(funnelConstraints))
		deepCopy(&funnelConstraints, &baseFunnelConstraints)
		// Remove pLen-2 constraints.
		baseFunnelConstraints[pLen-2] = P.EventConstraints{}
	} else if node.NodeType == NODE_TYPE_USER_PROPERTY {
		pLen := len(funnelEvents)
		if pLen == 1 {
			// Prepend AllActiveUsers to the begining.
			// When length 1, the added constraint is collapsed on endEvent and needs to
			// be removed from endEvent and moved to AllActiveUsers.
			funnelEvents = append([]string{U.SEN_ALL_ACTIVE_USERS}, funnelEvents...)
			funnelConstraints = append([]P.EventConstraints{node.AddedConstraint}, funnelConstraints...)
			pLen++
			// Remove addedConstraint from endEvent.
			for j, pNConstraint := range funnelConstraints[pLen-1].UPNumericConstraints {
				for _, aNConstraint := range node.AddedConstraint.UPNumericConstraints {
					if reflect.DeepEqual(pNConstraint, aNConstraint) {
						funnelConstraints[pLen-1].UPNumericConstraints[j] = P.NumericConstraint{}
					}
				}
			}
			for j, pCConstraint := range funnelConstraints[pLen-1].UPCategoricalConstraints {
				for _, aCConstraint := range node.AddedConstraint.UPCategoricalConstraints {
					if reflect.DeepEqual(pCConstraint, aCConstraint) {
						funnelConstraints[pLen-1].UPCategoricalConstraints[j] = P.CategoricalConstraint{}
					}
				}
			}
		}
		// Base funnel events are the same.
		baseFunnelEvents = append(make([]string, 0, len(funnelEvents)), funnelEvents...)
		baseFunnelConstraints = make([]P.EventConstraints, len(funnelConstraints))
		deepCopy(&funnelConstraints, &baseFunnelConstraints)
		// Remove addedConstraint from baseFunnelConstraint .
		for j, pNConstraint := range baseFunnelConstraints[pLen-2].UPNumericConstraints {
			for _, aNConstraint := range node.AddedConstraint.UPNumericConstraints {
				if reflect.DeepEqual(pNConstraint, aNConstraint) {
					baseFunnelConstraints[pLen-2].UPNumericConstraints[j] = P.NumericConstraint{}
				}
			}
		}
		for j, pCConstraint := range baseFunnelConstraints[pLen-2].UPCategoricalConstraints {
			for _, aCConstraint := range node.AddedConstraint.UPCategoricalConstraints {
				if reflect.DeepEqual(pCConstraint, aCConstraint) {
					baseFunnelConstraints[pLen-2].UPCategoricalConstraints[j] = P.CategoricalConstraint{}
				}
			}
		}
	}
	return funnelEvents, funnelConstraints, baseFunnelEvents, baseFunnelConstraints
}

func buildFunnelGraphResult(
	node *ItreeNode, funnelEvents []string, funnelConstraints []P.EventConstraints,
	baseFunnelEvents []string, baseFunnelConstraints []P.EventConstraints,
	pw PatternServiceWrapperInterface) (
	*graphResult, error) {
	baseFunnelData := buildFunnelData(baseFunnelEvents, baseFunnelConstraints, node, true, pw)
	funnelData := buildFunnelData(funnelEvents, funnelConstraints, node, false, pw)

	baseFunnelLength := len(baseFunnelData)
	baseFunnelConversionPercent := baseFunnelData[baseFunnelLength-2].ConversionPercent
	funnelLength := len(funnelData)
	funnelConversionPercent := funnelData[funnelLength-2].ConversionPercent
	if funnelConversionPercent > baseFunnelConversionPercent {
		funnelData[funnelLength-2].NodeType = "positive"
	} else if funnelConversionPercent < baseFunnelConversionPercent {
		funnelData[funnelLength-2].NodeType = "negative"
	}

	header, explanations := headerString(funnelData, node.NodeType,
		funnelConversionPercent, baseFunnelConversionPercent)
	chart := graphResult{
		Type:         "funnel",
		Header:       header,
		Explanations: explanations,
		Datasets: []map[string]interface{}{
			map[string]interface{}{
				"base_funnel_data": baseFunnelData,
				"funnel_data":      funnelData,
			},
		},
	}
	return &chart, nil
}

func buildBarGraphResult(node *ItreeNode) (*graphResult, error) {
	pLen := len(node.Pattern.EventNames)
	patternConstraints := node.PatternConstraints
	if patternConstraints == nil {
		patternConstraints = make([]P.EventConstraints, pLen)
	}
	propertyValues := []string{}
	increasedValues := []string{}
	decreasedValues := []string{}
	patternCounts := []float64{}
	patternPercentages := map[string]float64{}
	ruleCounts := []float64{}
	rulePercentages := map[string]float64{}
	percentageGain := 0.0
	percentageLoss := 0.0
	for i := 0; i < len(node.KLDistances); i++ {
		if node.KLDistances[i].Fcp <= 0 && node.KLDistances[i].Fcr <= 0 {
			log.WithFields(log.Fields{
				"KLNode": node.KLDistances[i], "Node on Property:": node.PropertyName}).Debug(
				"Skipping KLNode")
			continue
		}

		patternPercentage := node.KLDistances[i].Fcp / node.KLDistances[i].Fpp
		// In percentage rounded to two decimals.
		patternPercentage = math.Round(patternPercentage*1000) / 10

		rulePercentage := node.KLDistances[i].Fcr / node.KLDistances[i].Fpr
		// In percentage rounded to two decimals.
		rulePercentage = math.Round(rulePercentage*1000) / 10

		patternCounts = append(patternCounts, node.KLDistances[i].Fcp)
		patternPercentages[node.KLDistances[i].PropertyValue] = patternPercentage

		ruleCounts = append(ruleCounts, node.KLDistances[i].Fcr)
		rulePercentages[node.KLDistances[i].PropertyValue] = rulePercentage

		propertyValues = append(propertyValues, node.KLDistances[i].PropertyValue)
		if node.KLDistances[i].PropertyValue != OTHER_PROPERTY_VALUES_LABEL &&
			node.KLDistances[i].PropertyValue != NONE_PROPERTY_VALUES_LABEL {
			if rulePercentage > patternPercentage {
				percentageGain += (rulePercentage - patternPercentage)
				increasedValues = append(increasedValues, node.KLDistances[i].PropertyValue)
			} else if rulePercentage < patternPercentage {
				percentageLoss += (patternPercentage - rulePercentage)
				decreasedValues = append(decreasedValues, node.KLDistances[i].PropertyValue)
			}
		}
	}
	if len(increasedValues) == 0 && len(decreasedValues) == 0 {
		return nil, fmt.Errorf("No increase or decrease in graphs.")
	}
	patternUsersLabel := ""
	if pLen == 1 {
		patternUsersLabel += U.SEN_ALL_ACTIVE_USERS_DISPLAY_STRING
	} else {
		patternUsersLabel += "Users who "
		for i := 0; i < pLen-3; i++ {
			patternUsersLabel += eventStringWithConditions(
				node.Pattern.EventNames[i], &node.PatternConstraints[i])
			patternUsersLabel += " and "
		}
		patternUsersLabel += eventStringWithConditions(
			node.Pattern.EventNames[pLen-2], &node.PatternConstraints[pLen-2])
	}
	ruleUsersLabel := ""
	if pLen == 1 {
		ruleUsersLabel += "Users who " + eventStringWithConditions(
			node.Pattern.EventNames[0], &node.PatternConstraints[0])
	} else {
		ruleUsersLabel += patternUsersLabel + " and " + eventStringWithConditions(
			node.Pattern.EventNames[pLen-1], &node.PatternConstraints[pLen-1])
	}
	var headerString string
	var explanations []string
	if len(increasedValues) > 0 && len(decreasedValues) > 0 {
		// Look at average gain.
		if (percentageGain / float64(len(increasedValues))) > (percentageLoss / float64(len(decreasedValues))) {
			headerString, explanations = barGraphHeaderString(node.Pattern.EventNames, patternConstraints,
				node.PropertyName, increasedValues, patternUsersLabel, patternPercentages,
				ruleUsersLabel, rulePercentages, true)
		} else {
			headerString, explanations = barGraphHeaderString(node.Pattern.EventNames, patternConstraints,
				node.PropertyName, decreasedValues, patternUsersLabel, patternPercentages,
				ruleUsersLabel, rulePercentages, false)
		}
	} else if percentageGain >= 0.0 {
		headerString, explanations = barGraphHeaderString(node.Pattern.EventNames, patternConstraints,
			node.PropertyName, increasedValues, patternUsersLabel, patternPercentages,
			ruleUsersLabel, rulePercentages, true)
	} else {
		headerString, explanations = barGraphHeaderString(node.Pattern.EventNames, patternConstraints,
			node.PropertyName, decreasedValues, patternUsersLabel, patternPercentages,
			ruleUsersLabel, rulePercentages, false)
	}

	// Bar Chart.
	chart := &graphResult{
		Type:         "bar",
		Header:       headerString,
		Explanations: explanations,
		Labels:       propertyValues,
		Datasets: []map[string]interface{}{
			map[string]interface{}{
				"label": patternUsersLabel,
				"data":  patternCounts,
			},
			map[string]interface{}{
				"label": ruleUsersLabel,
				"data":  ruleCounts,
			},
		},
		XLabel: node.PropertyName,
		YLabel: "Number of users",
	}
	return chart, nil
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
		var chart *graphResult = nil
		if node.NodeType == NODE_TYPE_SEQUENCE || node.NodeType == NODE_TYPE_EVENT_PROPERTY ||
			node.NodeType == NODE_TYPE_USER_PROPERTY {
			funnelEvents, funnelConstraints, baseFunnelEvents, baseFunnelConstraints := buildFunnelFormats(node)
			if c, err := buildFunnelGraphResult(node, funnelEvents, funnelConstraints, baseFunnelEvents, baseFunnelConstraints, pw); err != nil {
				log.Error(err)
				continue
			} else {
				chart = c
			}
		} else if node.NodeType == NODE_TYPE_GRAPH_EVENT_PROPERTIES || node.NodeType == NODE_TYPE_GRAPH_USER_PROPERTIES {
			if c, err := buildBarGraphResult(node); err != nil {
				log.Error(err)
				continue
			} else {
				chart = c
			}
		}
		if chart != nil {
			results.Charts = append(results.Charts, *chart)
		}
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
			// InformationDrop * parentPatternFrequency is the ranking score for the node.
			scoreI := iPatternNodes[i].InformationDrop * (iPatternNodes[i].Fpp - iPatternNodes[i].OtherFcp)
			scoreJ := iPatternNodes[j].InformationDrop * (iPatternNodes[j].Fpp - iPatternNodes[j].OtherFcp)
			return (scoreI > scoreJ)
		})

	results := buildFactorResultsFromPatterns(iPatternNodes, pw)

	maxPatterns := 50
	if len(results.Charts) > maxPatterns {
		results.Charts = results.Charts[:maxPatterns]
	}

	return results, nil
}
