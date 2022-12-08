// pattern_service_wrapper
package pattern_service_wrapper

import (
	"factors/model/model"
	P "factors/pattern"
	PC "factors/pattern_client"
	U "factors/util"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Fetches information from pattern server and operations on patterns in local cache.
type PatternServiceWrapperInterface interface {
	GetPerUserCount(reqId string, eventNames []string,
		patternConstraints []P.EventConstraints) (uint, bool)
	GetPerOccurrenceCount(reqId string, eventNames []string,
		patternConstraints []P.EventConstraints) (uint, bool)
	GetCount(reqId string, eventNames []string,
		patternConstraints []P.EventConstraints, countType string) (uint, bool)
	GetPattern(reqId string, eventNames []string) *P.Pattern
	GetAllPatterns(reqId, startEvent, endEvent string) ([]*P.Pattern, error)
	GetAllContainingPatterns(reqId, event string) ([]*P.Pattern, error)
	GetTotalEventCount(reqId string) uint
	GetUserAndEventsInfo() *P.UserAndEventsInfo
}

type PatternServiceWrapper struct {
	projectId         int64
	modelId           uint64
	pMap              map[string]*P.Pattern
	totalEventCount   uint
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
	userAndEventsInfo, _, _ := PC.GetUserAndEventsInfo("", pw.projectId, pw.modelId)
	return userAndEventsInfo
}

func (pw *PatternServiceWrapper) GetPerUserCount(reqId string, eventNames []string,
	patternConstraints []P.EventConstraints) (uint, bool) {
	pattern := pw.GetPattern(reqId, eventNames)
	if pattern != nil {
		count, err := pattern.GetPerUserCount(patternConstraints)
		if err == nil {
			return count, true
		}
	}
	return 0, false
}

func (pw *PatternServiceWrapper) GetPerOccurrenceCount(reqId string, eventNames []string,
	patternConstraints []P.EventConstraints) (uint, bool) {
	pattern := pw.GetPattern(reqId, eventNames)
	if pattern != nil {
		count, err := pattern.GetPerOccurrenceCount(patternConstraints)
		if err == nil {
			return count, true
		}
	}
	return 0, false
}
func (pw *PatternServiceWrapper) GetCount(reqId string, eventNames []string,
	patternConstraints []P.EventConstraints, countType string) (uint, bool) {
	if countType == P.COUNT_TYPE_PER_USER {
		return pw.GetPerUserCount(reqId, eventNames, patternConstraints)
	} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
		if len(eventNames) == 1 && eventNames[0] == U.SEN_ALL_EVENTS {
			return pw.GetTotalEventCount(reqId), true
		}
		return pw.GetPerOccurrenceCount(reqId, eventNames, patternConstraints)
	}
	log.Errorf(fmt.Sprintf("Unrecognized count type: %s", countType))
	return 0, false
}

func (pw *PatternServiceWrapper) GetPattern(reqId string, eventNames []string) *P.Pattern {
	var pattern *P.Pattern = nil
	var found bool
	eventsHash := P.EventArrayToString(eventNames)
	if pattern, found = pw.pMap[eventsHash]; !found {
		// Fetch from server.
		patterns, err := PC.GetPatterns(reqId, pw.projectId, pw.modelId, [][]string{eventNames})
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

func (pw *PatternServiceWrapper) GetAllPatterns(reqId,
	startEvent, endEvent string) ([]*P.Pattern, error) {
	// Fetch from server.
	patterns, err := PC.GetAllPatterns(reqId, pw.projectId, pw.modelId, startEvent, endEvent)
	// Add it to cache.
	for _, p := range patterns {
		pw.pMap[P.EventArrayToString(p.EventNames)] = p
	}
	return patterns, err
}

func (pw *PatternServiceWrapper) GetAllContainingPatterns(reqId, event string) ([]*P.Pattern, error) {
	// Fetch from server.
	patterns, err := PC.GetAllContainingPatterns(reqId, pw.projectId, pw.modelId, event)
	// Add it to cache.
	for _, p := range patterns {
		pw.pMap[P.EventArrayToString(p.EventNames)] = p
	}
	return patterns, err
}

func (pw *PatternServiceWrapper) GetTotalEventCount(reqId string) uint {
	if pw.totalEventCount > 0 {
		// Fetch from cache.
		return pw.totalEventCount
	}
	if tc, err := PC.GetTotalEventCount(reqId, pw.projectId, pw.modelId); err != nil {
		return 0
	} else {
		pw.totalEventCount = uint(tc)
		return pw.totalEventCount
	}
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
	if eventString == U.SEN_ALL_EVENTS {
		eventString = U.SEN_ALL_EVENTS_DISPLAY_STRING
	}
	// Change URL events names to Visited URL for better readability.
	if strings.HasSuffix(eventString, "/") && !strings.HasPrefix(eventString, "Visited") {
		eventString = "Visited " + eventString
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

func funnelHeaderString(funnelData funnelNodeResults, nodeType int,
	funnelConversionPercent float64, baseFunnelConversionPercent float64,
	countType string) (string, []string) {
	var header string
	pLen := len(funnelData)
	if pLen < 2 {
		log.Error(fmt.Sprintf("Unexpected! Funnel: %s ", funnelData))
		return header, []string{}
	}
	var impactString string
	// Impact event.
	if nodeType == NODE_TYPE_SEQUENCE {
		if countType == P.COUNT_TYPE_PER_USER {
			impactString = fmt.Sprintf("who have %s", eventStringWithConditions(
				funnelData[pLen-2].EventName, nil))
		} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
			impactString = fmt.Sprintf("of %s", eventStringWithConditions(funnelData[pLen-2].EventName, nil))
		}
	} else if nodeType == NODE_TYPE_EVENT_PROPERTY || nodeType == NODE_TYPE_USER_PROPERTY {
		if funnelData[pLen-2].EventName == U.SEN_ALL_ACTIVE_USERS {
			impactString = fmt.Sprintf("with %s", eventStringWithConditions("", funnelData[pLen-2].Constraints))
		} else {
			if countType == P.COUNT_TYPE_PER_USER {
				impactString = fmt.Sprintf("who have %s with %s", funnelData[pLen-2].EventName,
					eventStringWithConditions("", funnelData[pLen-2].Constraints))
			} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
				impactString = fmt.Sprintf("of %s with %s", funnelData[pLen-2].EventName,
					eventStringWithConditions("", funnelData[pLen-2].Constraints))
			}
		}
	}

	endEventString := eventStringWithConditions(
		funnelData[pLen-1].EventName, funnelData[pLen-1].Constraints)

	otherEventString := ""
	for i := 0; i < pLen-2; i++ {
		if i == 0 && (funnelData[i].EventName == U.SEN_ALL_ACTIVE_USERS ||
			funnelData[i].EventName == U.SEN_ALL_EVENTS) {
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
	if countType == P.COUNT_TYPE_PER_USER {
		header = fmt.Sprintf("Users %s%s%s %s.", impactString, otherEventString, conversionChangeString, endEventString)
	} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
		header = fmt.Sprintf("Occurrences %s%s%s convert to %s.", impactString, otherEventString, conversionChangeString, endEventString)
	}
	return header, []string{}
}

func barGraphHeaderString(
	patternEvents []string, patternConstraints []P.EventConstraints,
	propertyName string, propertyValues []string,
	patternsLabel string,
	patternPercentages map[string]float64,
	rulesLabel string,
	rulePercentages map[string]float64,
	isIncrement bool,
	countType string) (string, []string) {
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
		if countType == P.COUNT_TYPE_PER_USER {
			impactString = fmt.Sprintf("who have %s with %s in %s",
				eventStringWithConditions(patternEvents[pLen-2], nil),
				propertyName, propertyValuesString)
		} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
			impactString = fmt.Sprintf("of %s with %s in %s",
				eventStringWithConditions(patternEvents[pLen-2], nil),
				propertyName, propertyValuesString)
		}
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
	header := ""
	if countType == P.COUNT_TYPE_PER_USER {
		header = fmt.Sprintf("Users %s%s%s %s.", impactString, otherEventString, conversionChangeString, endEventString)
	} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
		header = fmt.Sprintf("Occurrences %s%s%s convert to %s.", impactString, otherEventString, conversionChangeString, endEventString)
	}

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
			rulesLabel,
			propertyName,
		))

	headerExplanation = append(
		headerExplanation,
		fmt.Sprintf(
			"%0.1f%% of %s, have these %s.",
			totalPatternPercentage,
			patternsLabel,
			propertyName,
		))

	//log.WithFields(log.Fields{"events": patternEvents,
	//	"patternConstraints": patternConstraints,
	//	"endEventString":     endEventString, "pLen": pLen}).Debug("Graph results.")
	return header, headerExplanation
}

func buildFunnelData(reqId string,
	funnelEvents []string, funnelConstraints []P.EventConstraints,
	node *ItreeNode, isBaseFunnel bool, countType string,
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
			funnelSubsequencePerUserCount, found = pw.GetCount(reqId,
				funnelEvents[:i+1], funnelConstraints[:i+1], countType)
		}
		if !found {
			log.Errorf(fmt.Sprintf(
				"Subsequence %s not as frequent as sequence %s",
				P.EventArrayToString(funnelEvents[:i+1]), ","), funnelEvents)
			funnelSubsequencePerUserCount, _ = pw.GetCount(reqId, funnelEvents, funnelConstraints, countType)
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
func getCountConstraints(constraint P.EventConstraints) int {
	return len(constraint.EPCategoricalConstraints) +
		len(constraint.EPNumericConstraints) +
		len(constraint.UPCategoricalConstraints) +
		len(constraint.UPNumericConstraints)
}

func buildFunnelFormats(node *ItreeNode, countType string) (
	[]string, []P.EventConstraints, []string, []P.EventConstraints) {
	var funnelConstraints []P.EventConstraints
	// https://stackoverflow.com/questions/46790190/quicker-way-to-deepcopy-objects-in-golang
	funnelEvents := append(make([]string, 0, len(node.Pattern.EventNames)), node.Pattern.EventNames...)
	if node.PatternConstraints != nil {
		funnelConstraints = make([]P.EventConstraints, len(node.PatternConstraints))
		U.DeepCopy(&node.PatternConstraints, &funnelConstraints)
	} else {
		funnelConstraints = make([]P.EventConstraints, len(funnelEvents))
	}
	funnelConstraints = append([]P.EventConstraints(nil), funnelConstraints...)
	var baseFunnelEvents []string
	var baseFunnelConstraints []P.EventConstraints
	if node.NodeType == NODE_TYPE_SEQUENCE || node.NodeType == NODE_TYPE_CAMPAIGN {
		pLen := len(funnelEvents)
		if pLen == 2 {
			if countType == P.COUNT_TYPE_PER_USER {
				// Prepend AllActiveUsers Event at the begining for readability.
				funnelEvents = append([]string{U.SEN_ALL_ACTIVE_USERS}, funnelEvents...)
				funnelConstraints = append([]P.EventConstraints{P.EventConstraints{}}, funnelConstraints...)
				pLen++
			} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
				// Prepend AllEvents  at the begining for readability.
				funnelEvents = append([]string{U.SEN_ALL_EVENTS}, funnelEvents...)
				funnelConstraints = append([]P.EventConstraints{P.EventConstraints{}}, funnelConstraints...)
				pLen++
			}
		}
		// Skip (n - 1)st element and constraint for baseFunnel.
		baseFunnelEvents = append(append([]string(nil), funnelEvents[:pLen-2]...), funnelEvents[pLen-1:]...)
		baseFunnelConstraints = append(append([]P.EventConstraints(nil), funnelConstraints[:pLen-2]...), funnelConstraints[pLen-1:]...)
	} else if node.NodeType == NODE_TYPE_EVENT_PROPERTY {
		pLen := len(funnelEvents)
		// Base funnel events are the same.
		baseFunnelEvents = append(make([]string, 0, len(funnelEvents)), funnelEvents...)
		baseFunnelConstraints = make([]P.EventConstraints, len(funnelConstraints))
		U.DeepCopy(&funnelConstraints, &baseFunnelConstraints)
		// Remove pLen-2 constraints.
		baseFunnelConstraints[pLen-2] = P.EventConstraints{}
	} else if node.NodeType == NODE_TYPE_USER_PROPERTY {
		pLen := len(funnelEvents)
		if pLen == 1 {
			if countType == P.COUNT_TYPE_PER_USER {
				allConstraintsCount := getCountConstraints(funnelConstraints[0])
				if allConstraintsCount > 1 {
					funnelEvents = append([]string{U.SEN_ALL_ACTIVE_USERS}, funnelEvents...)
					var baseNodeConstraint P.EventConstraints
					funnelConstraints = append(funnelConstraints, baseNodeConstraint)
					pLen++
				} else {
					// Prepend AllActiveUsers to the begining.
					// When length 1, the added constraint is collapsed on endEvent and needs to
					// be removed from endEvent and moved to AllActiveUsers.
					funnelEvents = append([]string{U.SEN_ALL_ACTIVE_USERS}, funnelEvents...)
					var baseNodeConstraint P.EventConstraints
					funnelConstraints = append(funnelConstraints, baseNodeConstraint)
					pLen++
				}
			} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
				// Prepend AllEvents  at the begining for readability.
				funnelEvents = append([]string{U.SEN_ALL_EVENTS}, funnelEvents...)
				funnelConstraints = append([]P.EventConstraints{P.EventConstraints{}}, funnelConstraints...)
				pLen++
			}
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
		U.DeepCopy(&funnelConstraints, &baseFunnelConstraints)
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

func buildFunnelGraphResult(reqId string,
	node *ItreeNode, funnelEvents []string, funnelConstraints []P.EventConstraints,
	baseFunnelEvents []string, baseFunnelConstraints []P.EventConstraints,
	countType string, pw PatternServiceWrapperInterface) (
	*graphResult, error) {
	baseFunnelData := buildFunnelData(reqId, baseFunnelEvents, baseFunnelConstraints, node, true, countType, pw)
	funnelData := buildFunnelData(reqId, funnelEvents, funnelConstraints, node, false, countType, pw)

	baseFunnelLength := len(baseFunnelData)
	baseFunnelConversionPercent := baseFunnelData[baseFunnelLength-2].ConversionPercent
	funnelLength := len(funnelData)
	funnelConversionPercent := funnelData[funnelLength-2].ConversionPercent
	if funnelConversionPercent > baseFunnelConversionPercent {
		funnelData[funnelLength-2].NodeType = "positive"
	} else if funnelConversionPercent < baseFunnelConversionPercent {
		funnelData[funnelLength-2].NodeType = "negative"
	}

	header, explanations := funnelHeaderString(funnelData, node.NodeType,
		funnelConversionPercent, baseFunnelConversionPercent, countType)
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

func buildBarGraphResult(node *ItreeNode, countType string) (*graphResult, error) {
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

		MAX_LABEL_LENGTH := 30
		if len(node.KLDistances[i].PropertyValue) > MAX_LABEL_LENGTH {
			node.KLDistances[i].PropertyValue = node.KLDistances[i].PropertyValue[:MAX_LABEL_LENGTH] + "..."
		}
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
	patternsLabel := ""
	if pLen == 1 {
		if countType == P.COUNT_TYPE_PER_USER {
			patternsLabel += U.SEN_ALL_ACTIVE_USERS_DISPLAY_STRING
		} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
			patternsLabel += U.SEN_ALL_ACTIVE_USERS_DISPLAY_STRING
		}
	} else {
		for i := 0; i < pLen-3; i++ {
			patternsLabel += eventStringWithConditions(
				node.Pattern.EventNames[i], &node.PatternConstraints[i])
			patternsLabel += " and "
		}
		patternsLabel += eventStringWithConditions(
			node.Pattern.EventNames[pLen-2], &node.PatternConstraints[pLen-2])
	}
	rulesLabel := ""
	if pLen == 1 {
		rulesLabel += eventStringWithConditions(
			node.Pattern.EventNames[0], &node.PatternConstraints[0])
	} else {
		rulesLabel += patternsLabel + " and " + eventStringWithConditions(
			node.Pattern.EventNames[pLen-1], &node.PatternConstraints[pLen-1])
	}
	var headerString string
	var explanations []string
	if len(increasedValues) > 0 && len(decreasedValues) > 0 {
		// Look at average gain.
		if (percentageGain / float64(len(increasedValues))) > (percentageLoss / float64(len(decreasedValues))) {
			headerString, explanations = barGraphHeaderString(node.Pattern.EventNames, patternConstraints,
				node.PropertyName, increasedValues, patternsLabel, patternPercentages,
				rulesLabel, rulePercentages, true, countType)
		} else {
			headerString, explanations = barGraphHeaderString(node.Pattern.EventNames, patternConstraints,
				node.PropertyName, decreasedValues, patternsLabel, patternPercentages,
				rulesLabel, rulePercentages, false, countType)
		}
	} else if percentageGain >= 0.0 && len(increasedValues) > 0 {
		headerString, explanations = barGraphHeaderString(node.Pattern.EventNames, patternConstraints,
			node.PropertyName, increasedValues, patternsLabel, patternPercentages,
			rulesLabel, rulePercentages, true, countType)
	} else if percentageLoss >= 0.0 && len(decreasedValues) > 0 {
		headerString, explanations = barGraphHeaderString(node.Pattern.EventNames, patternConstraints,
			node.PropertyName, decreasedValues, patternsLabel, patternPercentages,
			rulesLabel, rulePercentages, false, countType)
	} else {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Unclear graph. increasedValues:%v, decreasedValues:%v, percentageGain: %f, percentageLoss:%f",
			increasedValues, decreasedValues, percentageGain, percentageLoss))
	}

	yLabel := ""
	if countType == P.COUNT_TYPE_PER_USER {
		yLabel = "Percentage of users"
	} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
		yLabel = "Percentage of occurrences"
	}

	// Compute percentages.
	totalPatternCounts := 0.0
	for _, c := range patternCounts {
		totalPatternCounts += float64(c)
	}
	patternPercentagesArray := []float64{}
	for _, c := range patternCounts {
		x := float64(c) * 100 / totalPatternCounts
		patternPercentagesArray = append(
			patternPercentagesArray,
			math.Floor(x*10)/10.0) // Round to one decimal point.
	}

	totalRuleCounts := 0.0
	for _, c := range ruleCounts {
		totalRuleCounts += float64(c)
	}
	rulePercentagesArray := []float64{}
	for _, c := range ruleCounts {
		x := float64(c) * 100 / totalRuleCounts

		rulePercentagesArray = append(
			rulePercentagesArray,
			math.Floor(x*10)/10.0) // Round to one decimal point.
	}

	// Bar Chart.
	chart := &graphResult{
		Type:         "bar",
		Header:       headerString,
		Explanations: explanations,
		Labels:       propertyValues,
		Datasets: []map[string]interface{}{
			map[string]interface{}{
				"label": patternsLabel,
				"data":  patternPercentagesArray,
			},
			map[string]interface{}{
				"label": rulesLabel,
				"data":  rulePercentagesArray,
			},
		},
		XLabel: node.PropertyName,
		YLabel: yLabel,
	}
	return chart, nil
}

func isLevel1Node(node *ItreeNode,
	seenPropertyConstraints *map[string]bool,
	seenEvents *map[string]bool) bool {
	if len(node.Pattern.EventNames) == 1 || (node.NodeType == NODE_TYPE_SEQUENCE && len(node.Pattern.EventNames) == 2) {
		// Ignore, AllActiveUsers results, since they are captured with session.
		return true
	}
	return false
}

func shouldFilterResult(node *ItreeNode,
	seenPropertyConstraints *map[string]bool,
	seenEvents *map[string]bool) bool {
	sessionIndex := 0
	for i, e := range node.Pattern.EventNames {
		if e == U.EVENT_NAME_SESSION {
			sessionIndex = i
		}
	}
	if sessionIndex > 0 {
		// Session event occurring in between a pattern is non intuitive.
		return true
	}
	return false
}

func isDuplicate(node *ItreeNode,
	seenPropertyConstraints *map[string]bool,
	seenEvents *map[string]bool) bool {
	isDup := false
	if node.NodeType == NODE_TYPE_SEQUENCE {
		nodePLen := len(node.Pattern.EventNames)
		addedEvent := node.Pattern.EventNames[nodePLen-2]
		if _, found := (*seenEvents)[addedEvent]; found {
			isDup = true
		}
		// Updates seen events.
		(*seenEvents)[addedEvent] = true
	} else if node.NodeType == NODE_TYPE_EVENT_PROPERTY || node.NodeType == NODE_TYPE_USER_PROPERTY {
		// Not deduping on graph results.
		propertyConstraints := []string{}
		for _, c := range node.AddedConstraint.EPCategoricalConstraints {
			propertyConstraints = append(propertyConstraints, fmt.Sprintf("%s:%s", c.PropertyName, c.PropertyValue))
		}
		for _, c := range node.AddedConstraint.EPNumericConstraints {
			propertyConstraints = append(propertyConstraints, c.PropertyName)
		}
		for _, c := range node.AddedConstraint.UPCategoricalConstraints {
			propertyConstraints = append(propertyConstraints, fmt.Sprintf("%s:%s", c.PropertyName, c.PropertyValue))
		}
		for _, c := range node.AddedConstraint.UPNumericConstraints {
			propertyConstraints = append(propertyConstraints, c.PropertyName)
		}
		for _, pc := range propertyConstraints {
			if _, found := (*seenPropertyConstraints)[pc]; found {
				isDup = true
				break
			}
		}
		for _, pc := range propertyConstraints {
			(*seenPropertyConstraints)[pc] = true
		}
	} else if node.NodeType == NODE_TYPE_GRAPH_EVENT_PROPERTIES || node.NodeType == NODE_TYPE_GRAPH_USER_PROPERTIES {
		if _, found := (*seenPropertyConstraints)[node.PropertyName]; found {
			isDup = true
		} else {
			(*seenPropertyConstraints)[node.PropertyName] = true
		}
	}
	return isDup
}

// Minimum Frequency of child rule to be considered as a valuable insight.
const MIN_FCR = 5
const MIN_FCP = 100

func buildFactorResultsFromPatterns(reqId string, nodes []*ItreeNode,
	countType string, pw PatternServiceWrapperInterface) FactorGraphResults {
	results := FactorGraphResults{Charts: []graphResult{}}
	seenPropertyConstraints := make(map[string]bool)
	seenEvents := make(map[string]bool)

	for _, node := range nodes {
		var chart *graphResult = nil
		if node.NodeType == NODE_TYPE_SEQUENCE || node.NodeType == NODE_TYPE_EVENT_PROPERTY ||
			node.NodeType == NODE_TYPE_USER_PROPERTY {
			// Dedup results to show more novel results as user scrolls down.
			if isDuplicate(node, &seenPropertyConstraints, &seenEvents) || isLevel1Node(node, &seenPropertyConstraints, &seenEvents) || shouldFilterResult(node, &seenPropertyConstraints, &seenEvents) {
				continue
			}
			if node.Fcr < MIN_FCR && node.Fcp < MIN_FCP {
				continue
			}
			funnelEvents, funnelConstraints, baseFunnelEvents, baseFunnelConstraints := buildFunnelFormats(node, countType)
			if c, err := buildFunnelGraphResult(reqId, node, funnelEvents, funnelConstraints,
				baseFunnelEvents, baseFunnelConstraints, countType, pw); err != nil {
				log.Error(err)
				continue
			} else {
				chart = c
			}
		} else if node.NodeType == NODE_TYPE_GRAPH_EVENT_PROPERTIES || node.NodeType == NODE_TYPE_GRAPH_USER_PROPERTIES {
			if isDuplicate(node, &seenPropertyConstraints, &seenEvents) || isLevel1Node(node, &seenPropertyConstraints, &seenEvents) || shouldFilterResult(node, &seenPropertyConstraints, &seenEvents) {
				continue
			}
			if c, err := buildBarGraphResult(node, countType); err != nil {
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

func NewPatternServiceWrapper(reqId string, projectId int64, modelId uint64) (*PatternServiceWrapper, error) {
	pMap := make(map[string]*P.Pattern)
	patternServiceWrapper := PatternServiceWrapper{
		projectId: projectId,
		modelId:   modelId,
		pMap:      pMap,
	}

	return &patternServiceWrapper, nil
}

func NewPatternServiceWrapperV2(reqId string, projectId int64, modelId uint64) (*PatternServiceWrapperV2, error) {
	pMap := make(map[string]*P.Pattern)
	patternServiceWrapper := PatternServiceWrapperV2{
		projectId: projectId,
		modelId:   modelId,
		pMap:      pMap,
	}

	return &patternServiceWrapper, nil
}

func Factor(reqId string, projectId int64, startEvent string,
	startEventConstraints *P.EventConstraints, endEvent string,
	endEventConstraints *P.EventConstraints, countType string,
	pw PatternServiceWrapperInterface) (FactorGraphResults, error) {
	if countType != P.COUNT_TYPE_PER_OCCURRENCE && countType != P.COUNT_TYPE_PER_USER {
		err := fmt.Errorf(fmt.Sprintf("Unknown count type: %s, for req: %s", countType, reqId))
		log.Error(err)
		return FactorGraphResults{}, err
	}
	iPatternNodes := []*ItreeNode{}
	if itree, err := BuildNewItree(reqId, startEvent, startEventConstraints,
		endEvent, endEventConstraints, pw, countType, projectId); err != nil {
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

	results := buildFactorResultsFromPatterns(reqId, iPatternNodes, countType, pw)

	maxPatterns := 150
	if len(results.Charts) > maxPatterns {
		results.Charts = results.Charts[:maxPatterns]
	}

	return results, nil
}

// Factors Constants
var CAMPAIGNTYPE string = "campaign"
var ATTRIBUTETYPE string = "attribute"
var JOURNEYTYPE string = "journey"

// Factors object
type Factors struct {
	GoalRule          model.FactorsGoalRule `json:"goal"`
	Insights          []*FactorsInsights    `json:"insights"`
	GoalUserCount     float64               `json:"goal_user_count"`
	TotalUsersCount   float64               `json:"total_users_count"`
	OverallPercentage float64               `json:"overall_percentage"`
	OverallMultiplier float64               `json:"overall_multiplier"`
	Type              string                `json:"type"`
}

// FactorsAttributeTuple object
type FactorsAttributeTuple struct {
	FactorsAttributeKey        string  `json:"factors_attribute_key"`
	FactorsAttributeValue      string  `json:"factors_attribute_value"`
	FactorsAttributeUpperBound float64 `json:"factors_attribute_upper_bound"`
	FactorsAttributeLowerBound float64 `json:"factors_attribute_lower_bound"`
	FactorsAttributeEquality   bool    `json:"factors_attribute_equality"`
	FactorsAttributeUseBound   string  `json:"factors_attribute_use_bound"`
}

// FactorsInsights object
type FactorsInsights struct {
	FactorsInsightsAttribute      []FactorsAttributeTuple `json:"factors_insights_attribute"`
	FactorsInsightsKey            string                  `json:"factors_insights_key"`
	FactorsInsightsMultiplier     float64                 `json:"factors_insights_multiplier"`
	FactorsInsightsPercentage     float64                 `json:"factors_insights_percentage"`
	FactorsInsightsUsersCount     float64                 `json:"factors_insights_users_count"`
	FactorsGoalUsersCount         float64                 `json:"factors_goal_users_count"`
	FactorsMultiplierIncreaseFlag bool                    `json:"factors_multiplier_increase_flag"`
	FactorsInsightsType           string                  `json:"factors_insights_type"`
	FactorsSubInsights            []*FactorsInsights      `json:"factors_sub_insights"`
	FactorsInsightsRank           uint64                  `json:"factors_insights_rank"`
}

func buildFactorResultsFromPatternsV1(reqId string, nodes []*ItreeNode, Level0GoalPercentage float64,
	countType string, pw PatternServiceWrapperInterface) []*FactorsInsights {
	seenPropertyConstraints := make(map[string]bool)
	seenEvents := make(map[string]bool)
	type parentInsightsTuple struct {
		parentIndex int
		index       int
		insights    FactorsInsights
	}
	indexLevelInsightsMap := make(map[int]FactorsInsights)
	indexLevelInsightsMap[0] = FactorsInsights{
		FactorsSubInsights: make([]*FactorsInsights, 0),
	}
	levelInsightsMap := make(map[int][]parentInsightsTuple)
	indexLevelMap := make(map[int]int)
	indexLevelMap[0] = 0

	for _, node := range nodes {
		indexLevelMap[node.Index] = indexLevelMap[node.ParentIndex] + 1
	}
	for rank, node := range nodes {
		// Dedup results to show more novel results as user scrolls down.
		if shouldFilterResult(node, &seenPropertyConstraints, &seenEvents) {
			continue
		}
		if node.Fcr < MIN_FCR && node.Fcp < MIN_FCP {
			continue
		}
		if node.NodeType == NODE_TYPE_EVENT_PROPERTY && indexLevelMap[node.Index] == 1 {
			continue
		}
		if node.NodeType == NODE_TYPE_SEQUENCE || node.NodeType == NODE_TYPE_EVENT_PROPERTY || node.NodeType == NODE_TYPE_USER_PROPERTY || node.NodeType == NODE_TYPE_CAMPAIGN {
			PLen := len(node.Pattern.EventNames)
			var insights FactorsInsights
			insights.FactorsInsightsRank = uint64(rank) // adding rank
			if node.NodeType == NODE_TYPE_EVENT_PROPERTY || node.NodeType == NODE_TYPE_USER_PROPERTY {
				attributes := make([]FactorsAttributeTuple, 0)
				for _, attribute := range node.AddedConstraint.EPNumericConstraints {
					attributes = append(attributes, FactorsAttributeTuple{
						FactorsAttributeKey:        attribute.PropertyName,
						FactorsAttributeLowerBound: attribute.LowerBound,
						FactorsAttributeUpperBound: attribute.UpperBound,
						FactorsAttributeEquality:   attribute.IsEquality,
						FactorsAttributeUseBound:   attribute.UseBound,
					})
				}
				for _, attribute := range node.AddedConstraint.EPCategoricalConstraints {
					attributes = append(attributes, FactorsAttributeTuple{
						FactorsAttributeKey:   attribute.PropertyName,
						FactorsAttributeValue: attribute.PropertyValue,
					})
				}
				for _, attribute := range node.AddedConstraint.UPNumericConstraints {
					attributes = append(attributes, FactorsAttributeTuple{
						FactorsAttributeKey:        attribute.PropertyName,
						FactorsAttributeLowerBound: attribute.LowerBound,
						FactorsAttributeUpperBound: attribute.UpperBound,
						FactorsAttributeEquality:   attribute.IsEquality,
						FactorsAttributeUseBound:   attribute.UseBound,
					})
				}
				for _, attribute := range node.AddedConstraint.UPCategoricalConstraints {
					attributes = append(attributes, FactorsAttributeTuple{
						FactorsAttributeKey:   attribute.PropertyName,
						FactorsAttributeValue: attribute.PropertyValue,
					})
				}
				insights.FactorsInsightsAttribute = attributes
				insights.FactorsInsightsType = ATTRIBUTETYPE
			}
			if node.NodeType == NODE_TYPE_SEQUENCE {
				insights.FactorsInsightsKey = node.Pattern.EventNames[PLen-2]
				insights.FactorsInsightsType = JOURNEYTYPE
			}
			if node.NodeType == NODE_TYPE_CAMPAIGN {
				insights.FactorsInsightsKey = P.ExtractCampaignName(node.Pattern.EventNames[PLen-2])
				insights.FactorsInsightsType = CAMPAIGNTYPE
			}
			insights.FactorsGoalUsersCount = node.Fcr
			insights.FactorsInsightsUsersCount = node.Fcp
			if node.Fcp > 0 {
				insights.FactorsInsightsPercentage = roundTo1Decimal(node.Fcr * 100 / node.Fcp)
			}
			insights.FactorsSubInsights = make([]*FactorsInsights, 0)
			indexLevelInsightsMap[node.Index] = insights
			levelInsightsMap[indexLevelMap[node.Index]] = append(levelInsightsMap[indexLevelMap[node.Index]], parentInsightsTuple{
				parentIndex: node.ParentIndex,
				index:       node.Index,
				insights:    insights,
			})
		}
	}
	indexLevelInsightsMap[0] = FactorsInsights{
		FactorsSubInsights:        make([]*FactorsInsights, 0),
		FactorsInsightsPercentage: Level0GoalPercentage,
	}
	for i := 2; i >= 1; i-- {
		prevLevelMap := make(map[interface{}]bool)
		for _, insight := range levelInsightsMap[i-1] {
			insig := indexLevelInsightsMap[insight.index]
			if insig.FactorsInsightsType == "" {
				continue
			} else if insig.FactorsInsightsType == ATTRIBUTETYPE {
				prevLevelMap[insig.FactorsInsightsAttribute[len(insig.FactorsInsightsAttribute)-1]] = true
			} else {
				prevLevelMap[insig.FactorsInsightsKey] = true
			}
		}

		for _, insight := range levelInsightsMap[i] {
			parent := indexLevelInsightsMap[insight.parentIndex]
			child := indexLevelInsightsMap[insight.index]
			if child.FactorsInsightsPercentage > parent.FactorsInsightsPercentage {
				child.FactorsMultiplierIncreaseFlag = true
			} else {
				child.FactorsMultiplierIncreaseFlag = false
			}
			if parent.FactorsInsightsPercentage != 0 {
				child.FactorsInsightsMultiplier = roundTo1Decimal(child.FactorsInsightsPercentage / parent.FactorsInsightsPercentage)
			} else {
				child.FactorsInsightsMultiplier = roundTo1Decimal(child.FactorsInsightsPercentage / 0.01)
			}
			if parent.FactorsInsightsType == "" || isValidInsightTransition(parent.FactorsInsightsType, child.FactorsInsightsType) {
				subInsights := parent.FactorsSubInsights
				subInsights = append(subInsights, trimChildNode(parent.FactorsInsightsType, child.FactorsInsightsType, parent, child))
				parent.FactorsSubInsights = rearrangeSubinsights(subInsights, prevLevelMap)
				indexLevelInsightsMap[insight.parentIndex] = parent
			}
		}
	}
	return indexLevelInsightsMap[0].FactorsSubInsights
}

func rearrangeSubinsights(subInsights []*FactorsInsights, prevLevelMap map[interface{}]bool) []*FactorsInsights {

	inPrevLevel := func(insi *FactorsInsights) bool {
		inPrevLevelbool := false
		if insi.FactorsInsightsType == ATTRIBUTETYPE {
			inPrevLevelbool = prevLevelMap[insi.FactorsInsightsAttribute[0]]
		} else {
			inPrevLevelbool = prevLevelMap[insi.FactorsInsightsKey]
		}
		return inPrevLevelbool
	}

	numerical := make([]*FactorsInsights, 0)
	withMultiplier1 := make([]*FactorsInsights, 0)
	others := make([]*FactorsInsights, 0)
	inPrev := make([]*FactorsInsights, 0)

	for _, insight := range subInsights {
		if inPrevLevel(insight) {
			inPrev = append(inPrev, insight)
		} else {
			if insight.FactorsInsightsMultiplier == 1 {
				withMultiplier1 = append(withMultiplier1, insight)
			} else if insight.FactorsInsightsType == ATTRIBUTETYPE && insight.FactorsInsightsAttribute[0].FactorsAttributeValue == "" {
				numerical = append(numerical, insight)
			} else {
				others = append(others, insight)
			}
		}
	}

	final := make([]*FactorsInsights, 0)
	final = append(final, others...)
	final = append(final, withMultiplier1...)
	final = append(final, numerical...)
	final = append(final, inPrev...)
	reRankOnWeightsExplainProperties(final)
	return final
}

func reRankOnWeightsExplainProperties(subInsights []*FactorsInsights) {

	log.Debug("reranking properties")
	type propertyRank struct {
		rank    uint64
		insight *FactorsInsights
		count   int64
		weight  float64
	}

	props := make([]propertyRank, 0)

	if len(subInsights) > 0 {
		for _, fi := range subInsights {
			if len(fi.FactorsInsightsAttribute) > 0 {
				w := U.GetExplainPropertyWeights(fi.FactorsInsightsAttribute[0].FactorsAttributeKey)
				pr := propertyRank{fi.FactorsInsightsRank, fi, int64(fi.FactorsInsightsUsersCount), float64(int64(fi.FactorsInsightsUsersCount)) * w}
				props = append(props, pr)
			}
		}
		if len(props) > 0 {
			sort.Slice(props, func(i, j int) bool {
				return props[i].weight > props[j].weight
			})
		}

		for rk, pr := range props {
			pr.insight.FactorsInsightsRank = uint64(rk)
		}
	}

}

func isValidInsightTransition(parentType string, childType string) bool {
	if parentType == JOURNEYTYPE {
		return true
	}
	if parentType == ATTRIBUTETYPE && childType == ATTRIBUTETYPE {
		return true
	}
	if (parentType == CAMPAIGNTYPE && childType == CAMPAIGNTYPE) || parentType == CAMPAIGNTYPE && childType == JOURNEYTYPE {
		return true
	}
	return false
}

func trimChildNode(parentType string, childType string, parent FactorsInsights, child FactorsInsights) *FactorsInsights {
	if parentType == ATTRIBUTETYPE && childType == ATTRIBUTETYPE {
		attributes := make([]FactorsAttributeTuple, 0)
		for _, childNodeAttribute := range child.FactorsInsightsAttribute {
			match := false
			for _, parentNodeAttribute := range parent.FactorsInsightsAttribute {
				if parentNodeAttribute.FactorsAttributeKey == "" {
					continue
				}
				if parentNodeAttribute == childNodeAttribute {
					match = true
				}
			}
			if match == false {
				attributes = append(attributes, childNodeAttribute)
			}
		}
		child.FactorsInsightsAttribute = attributes
	}
	return &child
}

func roundTo1Decimal(value float64) float64 {
	return math.Floor(value*10) / 10
}

func FactorV1(reqId string, projectId int64, startEvent string,
	startEventConstraints *P.EventConstraints, endEvent string,
	endEventConstraints *P.EventConstraints, countType string,
	pw PatternServiceWrapperInterface, debugKey string, debugParams map[string]string, includedEvents map[string]bool,
	includedEventProperties map[string]bool, includedUserProperties map[string]bool) (Factors, error, interface{}) {

	if countType != P.COUNT_TYPE_PER_OCCURRENCE && countType != P.COUNT_TYPE_PER_USER {
		err := fmt.Errorf(fmt.Sprintf("Unknown count type: %s, for req: %s", countType, reqId))
		log.Error(err)
		return Factors{}, err, nil
	}
	rootNode := &ItreeNode{}
	iPatternNodesUnsorted := []*ItreeNode{}
	var debugData interface{}
	if itree, err, debugInfo := BuildNewItreeV1(reqId, startEvent, startEventConstraints,
		endEvent, endEventConstraints, pw, countType, debugKey, debugParams, projectId, includedEventProperties, includedUserProperties, includedEvents); err != nil {
		log.Error(err)
		return Factors{}, err, nil
	} else {
		debugData = debugInfo
		for _, node := range itree.Nodes {
			if node.NodeType == NODE_TYPE_ROOT {

				rootNode = node
				// Root node.
				continue
			}
			iPatternNodesUnsorted = append(iPatternNodesUnsorted, node)
		}
	}

	v1FactorResult := Factors{}

	startDateTime := time.Now()
	goalUsersCount, totalUsersCount, goalsUsersPercentage := rootNode.Fcr, rootNode.Fcp, float64(0)
	if rootNode.Fcp > 0 {
		goalsUsersPercentage = roundTo1Decimal(rootNode.Fcr * 100 / rootNode.Fcp)
	}
	insights := buildFactorResultsFromPatternsV1(reqId, iPatternNodesUnsorted, goalsUsersPercentage, countType, pw)
	v1FactorResult.GoalUserCount = goalUsersCount
	v1FactorResult.OverallPercentage = goalsUsersPercentage
	v1FactorResult.OverallMultiplier = 1
	v1FactorResult.TotalUsersCount = totalUsersCount
	v1FactorResult.Insights = insights
	endDateTime := time.Now()
	log.WithFields(log.Fields{
		"time_taken": endDateTime.Sub(startDateTime).Milliseconds()}).Error("explain_debug_buildResults")
	return v1FactorResult, nil, debugData
}
