// itree
// Methods to build Insight/Information trees.
package pattern

import (
	"fmt"
	"math"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ItreeNode struct {
	Pattern            *Pattern
	PatternConstraints []EventConstraints
	NodeType           int
	Index              int
	ParentIndex        int
	// Gini Impurity at a node with pattern "ABCY" with parent node "ABY" is
	// the measure of amount of impurity of the two partitions of itemsets
	// obtained after splitting only the itemsets containing "AB" with "C"=1
	// and C=0. The class labels are assumed to be Y=1 and Y=0.
	// Let n(ABY) be fpr i.e frequency of parent rule.
	// Let n(AB) be fpp i.e frequency of parent pattern .
	// Let n(ABCY) be fcr i.e frequency of current rule.
	// Let n(ABC) be fcp i.e frequency of current pattern.

	// With C=1 we get itemsets with ABCY and ABCY'.  (A' is A=0)
	// With C=1 the GI we get RightGI = (n(ABCY) / n(ABC)) * ( 1 - n(ABCY) / n(ABC))
	// RightGI = (fcr / fcp) * (1 - fcr / fcp)
	// RightFraction = (n(ABC) / n(AB)) = (fcp / fpp)

	// With C=0 we get itemsets ABC'Y and ABC'Y'.
	// With C=0 we get LeftGI = (n(ABC'Y) / n(ABC')) *  ( 1 - n(ABC'Y) / n(ABC'))
	// LeftGI = ((fpr - fcr)/(fpp - fcp)) * [1 - ((fpr - fcr)/(fpp - fcp))]
	// LeftFraction = 1 - (fcp/fpp)

	// OverallGI = RightFraction * RightGI + LeftFraction * LeftGI
	// GiniDrop is the amount of impurity gain obtained wrt to impurity in
	// parent among itemsets AB.
	// GiniDrop = parentIndex.RightGI - OverallGI
	RightGI       float64
	RightFraction float64
	LeftGI        float64
	LeftFraction  float64
	OverallGI     float64
	GiniDrop      float64

	// confidence = fcr / fcp
	Confidence float64
	// confidenceGain = confidence - parentIndex.Confidence
	ConfidenceGain float64
	Fpp            float64
}

type Itree struct {
	Nodes    []*ItreeNode
	EndEvent string
}

const MAX_SEQUENCE_CHILD_NODES = 10
const MAX_PROPERTY_CHILD_NODES = 5
const NODE_TYPE_ROOT = 0
const NODE_TYPE_SEQUENCE = 1 // A child node that differs from its parent by an event.
const NODE_TYPE_PROPERTY = 2 // A child node that differs from its parent by a property constraint.

func constructPatternConstraints(
	pLen int, startEventConstraints *EventConstraints,
	endEventConstraints *EventConstraints) []EventConstraints {
	if startEventConstraints == nil && endEventConstraints == nil {
		return nil
	}
	if pLen == 0 {
		return nil
	}
	patternConstraints := make([]EventConstraints, pLen)
	for i := 0; i < pLen-1; i++ {
		if i == 0 && startEventConstraints != nil {
			patternConstraints[i] = *startEventConstraints
		} else {
			patternConstraints[i] = EventConstraints{}
		}
	}
	if endEventConstraints != nil {
		patternConstraints[pLen-1] = *endEventConstraints
	} else if startEventConstraints != nil && pLen == 1 {
		// When len = 1 startEventConstraints can be passed instead of
		// endEvent as any of the contraints can apply. Higher priority
		// to endEvent
		patternConstraints[pLen-1] = *startEventConstraints
	} else {
		patternConstraints[pLen-1] = EventConstraints{}
	}
	return patternConstraints
}

func (it *Itree) buildRootNode(
	pattern *Pattern, startEventConstraints *EventConstraints,
	endEventConstraints *EventConstraints, patternWrapper *PatternWrapper) (*ItreeNode, error) {
	pLen := len(pattern.EventNames)
	patternConstraints := make([]EventConstraints, pLen)
	for i := 0; i < pLen-1; i++ {
		patternConstraints[i] = EventConstraints{}
	}
	if startEventConstraints != nil {
		patternConstraints[0] = *startEventConstraints
	}
	if endEventConstraints != nil {
		// When len = 1 startEventConstraints can be passed instead of
		// endEvent as any of the contraints can apply. Higher priority
		// to endEvent
		patternConstraints[pLen-1] = *endEventConstraints
	}
	// The pattern for root node is just Y (end_event) without a start_event.
	// If p = n(Y) / totalUsers. n(Y) is number of users with occurrence of Y.
	// Else for root node of type X -> Y
	// p = n(X->Y) / n(X)
	// RightGI is p(1-p)
	// RightFraction is 1.0
	// OverallGI = rightGI
	// confidence is p.
	// confidenceGain, leftGi, leftFraction, giniDrop are not defined.
	// parentIndex is set to -1.
	patternCount, err := pattern.GetOncePerUserCount(patternConstraints)
	if err != nil {
		return nil, err
	}
	var p float64
	if len(pattern.EventNames) == 1 {
		p = float64(patternCount) / float64(pattern.UserCount)
	} else if len(pattern.EventNames) == 2 {
		c, ok := patternWrapper.GetPerUserCount(pattern.EventNames[:1],
			patternConstraints[:1])
		if !ok {
			return nil, fmt.Errorf(fmt.Sprintf("Frequency missing for startEvent: %s", pattern.String()))
		}
		p = float64(patternCount) / float64(c)
	} else {
		return nil, fmt.Errorf(
			fmt.Sprintf("Unexpected root node pattern %s", pattern.String()))
	}

	giniImpurity := p * (1.0 - p)
	node := ItreeNode{
		Pattern:            pattern,
		NodeType:           NODE_TYPE_ROOT,
		PatternConstraints: patternConstraints,
		ParentIndex:        -1,
		RightGI:            giniImpurity,
		RightFraction:      1.0,
		OverallGI:          giniImpurity,
		Confidence:         p,
		Fpp:                0.0,
	}
	log.WithFields(log.Fields{"node": node.Pattern.String(), "constraints": patternConstraints,
		"frequency": node.Pattern.OncePerUserCount}).Debug("Built root node.")
	return &node, nil
}

func (it *Itree) buildChildNode(
	pattern *Pattern, constraints []EventConstraints,
	nodeType int, parentIndex int, fpr float64, fpp float64,
	patternWrapper *PatternWrapper) (*ItreeNode, error) {
	parentNode := it.Nodes[parentIndex]
	parentPattern := parentNode.Pattern
	eLen := len(pattern.EventNames)
	peLen := len(parentPattern.EventNames)
	if parentNode.NodeType != NODE_TYPE_SEQUENCE && parentNode.NodeType != NODE_TYPE_ROOT {
		return nil, fmt.Errorf(fmt.Sprintf("Parent node of unexpected type: %v", parentNode))
	}

	if nodeType == NODE_TYPE_SEQUENCE && peLen != eLen-1 {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Current node with type sequence and pattern(%s) with constraints(%v) and parent index(%d), pattern(%s) with constraints(%v) not compatible.",
			pattern.String(), constraints, parentIndex, parentPattern.String(), parentNode.PatternConstraints))
	}
	if nodeType == NODE_TYPE_PROPERTY && peLen != eLen {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Current node with type property and pattern(%s) with constraints(%v) and parent index(%d), pattern(%s) with constraints(%v) not compatible",
			pattern.String(), constraints, parentIndex, parentPattern.String(), parentNode.PatternConstraints))
	}

	var fcr, fcp float64
	onceCount, err := pattern.GetOncePerUserCount(constraints)
	if err != nil {
		return nil, err
	}
	fcr = float64(onceCount)

	var subConstraints []EventConstraints
	if constraints == nil {
		subConstraints = nil
	} else {
		subConstraints = constraints[:eLen-1]
	}
	c, ok := patternWrapper.GetPerUserCount(pattern.EventNames[:eLen-1], subConstraints)
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("Frequency missing for pattern sequence %s", pattern.String()))
	}
	fcp = float64(c)

	if fcp <= 0.0 || fcr <= 0.0 {
		log.WithFields(log.Fields{"Child": pattern.String(), "constraints": constraints,
			"fcp": fcp, "fcr": fcr}).Debug("Child not frequent enough")
		return nil, nil
	}
	// Expected.
	// fpp >= fpr
	// fcp >= fcr
	// fpp >= fcp
	// fpr >= fcr
	if fpp < fpr || fcp < fcr || fpp < fcp || fpr < fcr {
		log.WithFields(log.Fields{"Child": pattern.String(), "constraints": constraints,
			"fpp": fpp, "fpr": fpr, "fcp": fcp, "fcr": fcr}).Debug("Inconsistent frequencies. Ignoring")
		return nil, nil
	}
	rightGI := (fcr / fcp) * (1.0 - fcr/fcp)
	rightFraction := fcp / fpp
	var leftGI float64
	if (fpp - fcp) > 0.0 {
		leftGI = ((fpr - fcr) / (fpp - fcp)) * (1.0 - ((fpr - fcr) / (fpp - fcp)))
	} else {
		leftGI = 0.0
	}
	var leftFraction float64 = 1.0 - rightFraction
	overallGI := rightFraction*rightGI + leftFraction*leftGI

	giniDrop := parentNode.RightGI - overallGI

	confidence := fcr / fcp
	confidenceGain := confidence - parentNode.Confidence

	node := ItreeNode{
		Pattern:            pattern,
		NodeType:           nodeType,
		PatternConstraints: constraints,
		ParentIndex:        parentIndex,
		RightGI:            rightGI,
		RightFraction:      rightFraction,
		LeftGI:             leftGI,
		LeftFraction:       leftFraction,
		OverallGI:          overallGI,
		GiniDrop:           giniDrop,
		Confidence:         confidence,
		ConfidenceGain:     confidenceGain,
		Fpp:                fpp,
	}
	log.WithFields(log.Fields{"node": node.Pattern.String(), "constraints": constraints,
		"parent": parentPattern.String(), "parentConstraints": parentNode.PatternConstraints,
		"fcr": fcr, "fcp": fcp, "fpr": fpr, "fpp": fpp, "GI": overallGI, "giniDrop": giniDrop}).Debug(
		"Built candidate child node.")
	return &node, nil
}

func (it *Itree) addNode(node *ItreeNode) int {
	node.Index = len(it.Nodes)
	it.Nodes = append(it.Nodes, node)
	log.WithFields(log.Fields{
		"node":        node.Pattern.String(),
		"nodeType":    node.NodeType,
		"constraints": node.PatternConstraints,
		"index":       node.Index,
		"parentIndex": node.ParentIndex,
		"GI":          node.OverallGI,
		"giniDrop":    node.GiniDrop}).Info(
		"Added node.")
	return node.Index
}

func isChildPattern(parentNode *ItreeNode, childPattern *Pattern) (bool, []EventConstraints, error) {
	if !isChildSequence(parentNode.Pattern.EventNames, childPattern.EventNames) {
		return false, nil, nil
	}
	var childPatternConstraints []EventConstraints
	if parentNode.PatternConstraints != nil {
		childPLen := len(childPattern.EventNames)
		childPatternConstraints = make([]EventConstraints, childPLen)
		for i := 0; i < childPLen-2; i++ {
			childPatternConstraints[i] = parentNode.PatternConstraints[i]
		}
		// childPattern has an extra event at childPLen-2 if childPLen > 2.
		// Minimum expected childPLen is 2. Else this will crash.
		if childPLen == 2 {
			// If parent is single event, then the constraints are attached to the
			// end event by default.
			childPatternConstraints[1] = parentNode.PatternConstraints[0]
		} else {
			childPatternConstraints[childPLen-2] = EventConstraints{}
		}
		childPatternConstraints[childPLen-1] = parentNode.PatternConstraints[childPLen-2]
	} else {
		childPatternConstraints = nil
	}

	onceCount, err := childPattern.GetOncePerUserCount(childPatternConstraints)
	if err != nil {
		return false, nil, err
	}
	if onceCount < 1 {
		return false, nil, nil
	}
	return true, childPatternConstraints, nil
}

func isChildSequence(parent []string, child []string) bool {
	// ABY and ABCY is true.
	// ABY and ACBY is false.
	// ABY and BCY is false.
	cLen := len(child)
	pLen := len(parent)
	if cLen-pLen != 1 {
		return false
	}
	cIndex := 0
	pIndex := 0
	differentChildIndex := -1
	for cIndex < cLen && pIndex < pLen {
		if parent[pIndex] == child[cIndex] {
			pIndex++
			cIndex++
		} else {
			if differentChildIndex != -1 {
				// More than one mismatch.
				return false
			}
			differentChildIndex = cIndex
			cIndex++
		}
	}
	if differentChildIndex != cLen-2 {
		// Parent and child should differ only at the end.
		return false
	}
	return true
}

func (it *Itree) buildAndAddSequenceChildNodes(
	parentNode *ItreeNode, candidatePattens []*Pattern,
	patternWrapper *PatternWrapper) ([]*ItreeNode, error) {

	parentPattern := parentNode.Pattern
	peLen := len(parentPattern.EventNames)
	onceCount, err := parentPattern.GetOncePerUserCount(parentNode.PatternConstraints)
	if err != nil {
		return nil, err
	}
	fpr := float64(onceCount)

	var fpp float64
	if peLen == 1 {
		// Parent is root node. Count is all users.
		fpp = float64(parentPattern.UserCount)
	} else {
		var subParentConstraints []EventConstraints
		if parentNode.PatternConstraints == nil {
			subParentConstraints = nil
		} else {
			subParentConstraints = parentNode.PatternConstraints[:peLen-1]
		}
		c, ok := patternWrapper.GetPerUserCount(parentPattern.EventNames[:peLen-1], subParentConstraints)
		if !ok {
			return nil, fmt.Errorf(fmt.Sprintf("Frequency missing for pattern sequence", parentPattern.String()))
		}
		fpp = float64(c)
	}

	childNodes := []*ItreeNode{}
	if fpp <= 0.0 || fpr <= 0.0 {
		log.WithFields(log.Fields{"Parent": parentPattern.String(), "fpr": fpr, "fpp": fpp}).Debug(
			"Parent not frequent enough")
		return childNodes, nil
	}
	for _, p := range candidatePattens {
		isChild, childPatternConstraints, err := isChildPattern(parentNode, p)
		if err != nil {
			return nil, err
		}
		if isChild {
			if cNode, err := it.buildChildNode(
				p, childPatternConstraints, NODE_TYPE_SEQUENCE,
				parentNode.Index, fpr, fpp, patternWrapper); err != nil {
				log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build child node")
				continue
			} else {
				if cNode == nil {
					continue
				}
				childNodes = append(childNodes, cNode)
			}
		}
	}
	// Sort in decreasing order of giniDrop.
	sort.SliceStable(childNodes,
		func(i, j int) bool {
			return (childNodes[i].GiniDrop > childNodes[j].GiniDrop)
		})

	// Only top MAX_SEQUENCE_CHILD_NODES in order of drop in GiniImpurity are selected.
	if len(childNodes) > MAX_SEQUENCE_CHILD_NODES {
		childNodes = childNodes[:MAX_SEQUENCE_CHILD_NODES]
	}

	addedChildNodes := []*ItreeNode{}
	for _, cNode := range childNodes {
		if cNode.GiniDrop <= 0.0 {
			continue
		}
		it.addNode(cNode)
		addedChildNodes = append(addedChildNodes, cNode)
		log.WithFields(log.Fields{"node": cNode.Pattern.String(),
			"constraints": cNode.PatternConstraints,
			"GI":          cNode.OverallGI, "giniDrop": cNode.GiniDrop}).Debug(
			"Added sequence child node.")
	}

	return addedChildNodes, nil
}

func (it *Itree) buildAndAddPropertyChildNodes(
	parentNode *ItreeNode, patternWrapper *PatternWrapper) ([]*ItreeNode, error) {
	// The top child nodes are obtained by adding constraints on the (N-1) event
	// of parent pattern.
	// i.e If parrent pattern is A -> B -> C -> Y with
	// some predefined constraints {c1, ... ck} on any of the events.
	// then an additional constraint is applied on event C, say ck+1.
	// The ones that causes the highest gini drop when all users who have
	// A->B->C with constraints are split by the additional constraint ck+1
	// taking Y (with constraint if any) as the classification label.
	parentPattern := parentNode.Pattern
	parentConstraints := parentNode.PatternConstraints
	var fpr, fpp float64
	pLen := len(parentPattern.EventNames)
	if pLen < 2 {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Pattern size should be atleast 2. %s", parentPattern.String()))
	}
	if patternWrapper.userAndEventsInfo == nil {
		// No event properties available.
		return []*ItreeNode{}, nil
	}
	if patternWrapper.userAndEventsInfo.EventPropertiesInfoMap == nil {
		// No event properties available.
		return []*ItreeNode{}, nil
	}
	eventName := parentPattern.EventNames[pLen-2]
	eventInfo, ok := (*patternWrapper.userAndEventsInfo.EventPropertiesInfoMap)[eventName]
	if !ok {
		// No properties.
		return []*ItreeNode{}, nil
	}

	onceCount, err := parentPattern.GetOncePerUserCount(parentConstraints)
	if err != nil {
		return nil, err
	}
	fpr = float64(onceCount)
	var subConstraints []EventConstraints
	if parentConstraints == nil {
		subConstraints = nil
	} else {
		subConstraints = parentConstraints[:pLen-1]
	}
	c, ok := patternWrapper.GetPerUserCount(parentPattern.EventNames[:pLen-1], subConstraints)
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("Frequency missing for pattern sequence %s", parentPattern.String()))
	}
	fpp = float64(c)

	// Map out the current constraints applied on N-1st event in the pattern.
	eventConstraints := EventConstraints{}
	if parentConstraints != nil {
		eventConstraints = parentConstraints[pLen-2]
	}
	if len(eventConstraints.EPCategoricalConstraints) > 0 ||
		len(eventConstraints.EPNumericConstraints) > 0 {
		log.Errorf(fmt.Sprintf(
			"Pattern %s already has constraints %v on event %s",
			parentPattern.String(), eventConstraints, eventName))
		return []*ItreeNode{}, nil
	}

	childNodes := []*ItreeNode{}
	if fpp <= 0.0 || fpr <= 0.0 {
		log.WithFields(log.Fields{"Parent": parentPattern.String(), "parentConstraints": parentConstraints,
			"fpr": fpr, "fpp": fpp}).Debug("Parent not frequent enough")
		return childNodes, nil
	}

	// Add children by splitting on constraints on categorical properties.
	pLen = len(parentPattern.EventNames)
	MAX_CAT_PROPERTIES_EVALUATED := 500
	MAX_CAT_VALUES_EVALUATED := 1000
	numP := 0
	for propertyName, seenValues := range eventInfo.CategoricalPropertyKeyValues {
		if numP > MAX_CAT_PROPERTIES_EVALUATED {
			break
		}
		numP++
		numVal := 0
		for value, _ := range seenValues {
			if numVal > MAX_CAT_VALUES_EVALUATED {
				break
			}
			numVal++
			childPatternConstraints := make([]EventConstraints, pLen)
			for i := 0; i < pLen; i++ {
				if parentConstraints != nil {
					childPatternConstraints[i] = parentConstraints[i]
				} else {
					childPatternConstraints[i] = EventConstraints{
						EPNumericConstraints:     []NumericConstraint{},
						EPCategoricalConstraints: []CategoricalConstraint{},
						UPNumericConstraints:     []NumericConstraint{},
						UPCategoricalConstraints: []CategoricalConstraint{},
					}
				}
			}
			// Add constraint to N-1st event. Assumed to be empty.
			// Empty check is done earlier.
			childPatternConstraints[pLen-2].EPCategoricalConstraints = []CategoricalConstraint{
				CategoricalConstraint{
					PropertyName:  propertyName,
					PropertyValue: value,
				},
			}

			// The sequence of event does not change for child.
			childPattern := parentNode.Pattern
			if cNode, err := it.buildChildNode(
				childPattern, childPatternConstraints, NODE_TYPE_PROPERTY,
				parentNode.Index, fpr, fpp, patternWrapper); err != nil {
				log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build child node")
				continue
			} else {
				if cNode == nil {
					continue
				}
				childNodes = append(childNodes, cNode)
			}
		}
	}

	// Add children by splitting on constraints on categorical properties.
	pLen = len(parentPattern.EventNames)
	MAX_NUM_PROPERTIES_EVALUATED := 500
	numP = 0
	for propertyName, _ := range eventInfo.NumericPropertyKeys {
		if numP > MAX_NUM_PROPERTIES_EVALUATED {
			break
		}
		numP++
		for _, pRange := range parentPattern.GetEventPropertyRanges(pLen-2, propertyName) {
			minValue := pRange[0]
			maxValue := pRange[1]
			// Try three combinations of constraints. v < min, min < v < max, v > max
			childPatternConstraints1 := make([]EventConstraints, pLen)
			childPatternConstraints2 := make([]EventConstraints, pLen)
			childPatternConstraints3 := make([]EventConstraints, pLen)
			for i := 0; i < pLen; i++ {
				if parentConstraints != nil {
					childPatternConstraints1[i] = parentConstraints[i]
					childPatternConstraints2[i] = parentConstraints[i]
					childPatternConstraints3[i] = parentConstraints[i]
				} else {
					childPatternConstraints1[i] = EventConstraints{
						EPNumericConstraints:     []NumericConstraint{},
						EPCategoricalConstraints: []CategoricalConstraint{},
						UPNumericConstraints:     []NumericConstraint{},
						UPCategoricalConstraints: []CategoricalConstraint{},
					}
					childPatternConstraints2[i] = EventConstraints{
						EPNumericConstraints:     []NumericConstraint{},
						EPCategoricalConstraints: []CategoricalConstraint{},
						UPNumericConstraints:     []NumericConstraint{},
						UPCategoricalConstraints: []CategoricalConstraint{},
					}
					childPatternConstraints3[i] = EventConstraints{
						EPNumericConstraints:     []NumericConstraint{},
						EPCategoricalConstraints: []CategoricalConstraint{},
						UPNumericConstraints:     []NumericConstraint{},
						UPCategoricalConstraints: []CategoricalConstraint{},
					}
				}
			}
			// Add constraint to N-1st event. Assumed to be empty.
			// Empty check is done earlier.
			childPatternConstraints1[pLen-2].EPNumericConstraints = []NumericConstraint{
				NumericConstraint{
					PropertyName: propertyName,
					LowerBound:   -math.MaxFloat64,
					UpperBound:   minValue,
				},
			}
			childPatternConstraints2[pLen-2].EPNumericConstraints = []NumericConstraint{
				NumericConstraint{
					PropertyName: propertyName,
					LowerBound:   minValue,
					UpperBound:   maxValue,
				},
			}
			childPatternConstraints3[pLen-2].EPNumericConstraints = []NumericConstraint{
				NumericConstraint{
					PropertyName: propertyName,
					LowerBound:   maxValue,
					UpperBound:   math.MaxFloat64,
				},
			}

			for _, cConstraint := range [][]EventConstraints{
				childPatternConstraints1, childPatternConstraints2, childPatternConstraints3} {
				if cNode, err := it.buildChildNode(
					parentPattern, cConstraint, NODE_TYPE_PROPERTY,
					parentNode.Index, fpr, fpp, patternWrapper); err != nil {
					log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build child node")
					continue
				} else {
					if cNode == nil {
						continue
					}
					childNodes = append(childNodes, cNode)
				}
			}
		}
	}

	// Sort in decreasing order of giniDrop.
	sort.SliceStable(childNodes,
		func(i, j int) bool {
			return (childNodes[i].GiniDrop > childNodes[j].GiniDrop)
		})

	// Only top MAX_PROPERTY_CHILD_NODES in order of drop in GiniImpurity are selected.
	if len(childNodes) > MAX_PROPERTY_CHILD_NODES {
		childNodes = childNodes[:MAX_PROPERTY_CHILD_NODES]
	}

	addedChildNodes := []*ItreeNode{}
	for _, cNode := range childNodes {
		if cNode.GiniDrop <= 0.0 {
			continue
		}
		it.addNode(cNode)
		addedChildNodes = append(addedChildNodes, cNode)
		log.WithFields(log.Fields{"node": cNode.Pattern.String(),
			"constraints": cNode.PatternConstraints,
			"GI":          cNode.OverallGI, "giniDrop": cNode.GiniDrop}).Debug(
			"Added property child node.")
	}
	return addedChildNodes, nil
}

func BuildNewItree(
	startEvent string, startEventConstraints *EventConstraints,
	endEvent string, endEventConstraints *EventConstraints,
	patternWrapper *PatternWrapper) (*Itree, error) {
	if endEvent == "" {
		return nil, fmt.Errorf("Missing end event")
	}

	itreeNodes := []*ItreeNode{}
	itree := Itree{
		Nodes:    itreeNodes,
		EndEvent: endEvent,
	}

	var rootNodePattern *Pattern = nil
	candidatePatterns := []*Pattern{}
	for _, p := range patternWrapper.patterns {
		pLen := len(p.EventNames)
		if (strings.Compare(endEvent, p.EventNames[pLen-1]) == 0) &&
			(startEvent == "" || strings.Compare(startEvent, p.EventNames[0]) == 0) {
			candidatePatterns = append(candidatePatterns, p)
			if (startEvent == "" && pLen == 1) || (startEvent != "" && pLen == 2) {
				if count, err := p.GetOncePerUserCount(constructPatternConstraints(
					pLen, startEventConstraints, endEventConstraints)); count > 0 {
					rootNodePattern = p
				} else if err != nil {
					return nil, err
				}
			}
		}
	}
	if rootNodePattern == nil {
		return nil, fmt.Errorf("Root node not found or frequency 0")
	}
	queue := []*ItreeNode{}

	rootNode, err := itree.buildRootNode(
		rootNodePattern, startEventConstraints, endEventConstraints, patternWrapper)
	if err != nil {
		return nil, err
	}

	itree.addNode(rootNode)
	queue = append(queue, rootNode)

	for len(queue) > 0 {
		parentNode := queue[0]
		if len(parentNode.Pattern.EventNames) >= 2 {
			if _, err := itree.buildAndAddPropertyChildNodes(
				parentNode, patternWrapper); err != nil {
				return nil, err
			}
		}

		if sequenceChildNodes, err := itree.buildAndAddSequenceChildNodes(parentNode,
			candidatePatterns, patternWrapper); err != nil {
			return nil, err
		} else {
			// Only sequenceChildNodes are explored further.
			queue = append(queue, sequenceChildNodes...)
		}
		queue = queue[1:]
	}

	//log.WithFields(log.Fields{"itree": itree}).Info("Returning Itree.")
	return &itree, nil
}
