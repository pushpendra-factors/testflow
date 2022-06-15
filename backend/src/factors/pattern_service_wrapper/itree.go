// itree
// Methods to build Insight/Information trees.
package pattern_service_wrapper

import (
	"errors"
	"factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	U "factors/util"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type ItreeNode struct {
	Pattern            *P.Pattern
	PatternConstraints []P.EventConstraints
	NodeType           int
	Index              int
	ParentIndex        int
	// Information at a node with pattern "ABCY" with parent node "ABY" is
	// the measure of amount of impurity of the two partitions of itemsets
	// obtained after splitting only the itemsets containing "AB" with "C"=1
	// and C=0. The class labels are assumed to be Y=1 and Y=0.
	// Let n(ABY) be fpr i.e frequency of parent rule.
	// Let n(AB) be fpp i.e frequency of parent pattern .
	// Let n(ABCY) be fcr i.e frequency of current rule.
	// Let n(ABC) be fcp i.e frequency of current pattern.

	// With C=1 we get all datapoints with pattern ABC as right child.
	// Right child is partitoned by the two subsets ABCY and ABCY'.  (A' is A=0)
	// With C=1 we get the probabilities for RightInformation as p, (1-p) with p = (n(ABCY) / n(ABC)) = fcr / fcp.
	// RightFraction = (n(ABC) / n(AB)) = (fcp / fpp)

	// With C=0 we get all data points with ABC' into left node.
	// The subsets partitioning ABC' are  ABC'YC' and ABC'Y'C'.
	// (It is not ABC'Y because ABYC would be counted in ABC'Y, but not part the superset ABC'.
	// Hence the extra C' at the end to ensure that the subset belongs to ABC')
	// n(ABC'YC') = n(ABY) - n(ABCY) - n(ABYC) + n(ABCYC)
	// n(ABC') = n(AB) - n(ABC)
	// * We are approximating Left information with
	// p = n(ABC'Y) / n(ABC') = (n(ABY) - n(ABCY)) / (n(AB) - n(ABC))
	// p = (fpr - fcr)/(fpp - fcp)
	// Although this number is not guaranteed to be less than 1*
	// With C=0 we get the probabilities for LeftInformation as p, (1-p) with p = (n(ABC'YC') / n(ABC'))
	// LeftFraction = 1 - (fcp/fpp)

	// OverallInformation = RightFraction * RightInformation + LeftFraction * LeftInformation
	// InformationDrop is the amount of impurity gain obtained wrt to impurity in
	// parent among itemsets AB.
	// InformationDrop = parentIndex.RightInformation - OverallInformation
	RightInformation   float64
	RightFraction      float64
	OverallInformation float64
	InformationDrop    float64

	// confidence = fcr / fcp
	Confidence float64
	// confidenceGain = confidence - parentIndex.Confidence
	ConfidenceGain float64
	Fpp            float64
	Fpr            float64
	Fcp            float64
	Fcr            float64
	// Count of number of data points that have other label in graph node.
	// Default 0.0 for other tables.
	OtherFcp        float64
	AddedConstraint P.EventConstraints

	// Used to store graphs.
	PropertyName string
	KLDistances  []KLDistanceUnitInfo
}

type Itree struct {
	Nodes    []*ItreeNode
	EndEvent string
}

type KLDistanceUnitInfo struct {
	Distance      float64
	PropertyValue string
	Fpp           float64
	Fpr           float64
	Fcp           float64
	Fcr           float64
}

const MAX_SEQUENCE_CHILD_NODES = 100
const MAX_PROPERTY_CHILD_NODES = 100
const MAX_NODES_TO_EVALUATE = 500

const NODE_TYPE_ROOT = 0
const NODE_TYPE_SEQUENCE = 1               // A child node that differs from its parent by an event.
const NODE_TYPE_EVENT_PROPERTY = 2         // A child node that differs from its parent by an event property constraint.
const NODE_TYPE_USER_PROPERTY = 3          // A child node that differs from its parent by an user property constraint.
const NODE_TYPE_GRAPH_EVENT_PROPERTIES = 4 // A child graph node that has a different distribution on an event property, compared to parent
const NODE_TYPE_GRAPH_USER_PROPERTIES = 5  // A child graph node that has a different distribution on an user property, compared to parent
const NODE_TYPE_CAMPAIGN = 6

const MAX_PROPERTIES_IN_GRAPH_NODE = 10
const OTHER_PROPERTY_VALUES_LABEL = "Other"
const NONE_PROPERTY_VALUES_LABEL = "None"

var MAX_CAT_PROPERTIES_EVALUATED int
var MAX_CAT_VALUES_EVALUATED int
var MAX_NUM_PROPERTIES_EVALUATED int

var log2Value float64 = math.Log(2)

// events in "*" are blacklisted for any goal

func log2(x float64) float64 {
	if x == 0 {
		return 1
	}
	return math.Log(x) / log2Value
}

func information(probabilities []float64) float64 {
	// The information in a distribution increases with
	// the amount of uncertainity in the distribution.
	// Ex. The sun rises in the east has zero information
	// because it has probability 1.
	// Currently returning entropy as information expressed in bits.
	info := 0.0
	for _, p := range probabilities {
		info += -(p * log2(p))
	}
	return info
}

func computeKLDistanceBits(patternProb float64, ruleProb float64) float64 {
	return -(ruleProb * (log2(patternProb) - log2(ruleProb)))
}

func constructPatternConstraints(
	pLen int, startEventConstraints *P.EventConstraints,
	endEventConstraints *P.EventConstraints) []P.EventConstraints {
	if startEventConstraints == nil && endEventConstraints == nil {
		return nil
	}
	if pLen == 0 {
		return nil
	}
	patternConstraints := make([]P.EventConstraints, pLen)
	for i := 0; i < pLen-1; i++ {
		if i == 0 && startEventConstraints != nil {
			patternConstraints[i] = *startEventConstraints
		} else {
			patternConstraints[i] = P.EventConstraints{}
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
		patternConstraints[pLen-1] = P.EventConstraints{}
	}
	return patternConstraints
}

func (it *Itree) buildRootNode(reqId string, allActiveUsers *P.Pattern,
	pattern *P.Pattern, startEventConstraints *P.EventConstraints,
	endEventConstraints *P.EventConstraints,
	patternWrapper PatternServiceWrapperInterface,
	countType string) (*ItreeNode, error) {
	pLen := len(pattern.EventNames)
	patternConstraints := make([]P.EventConstraints, pLen)
	for i := 0; i < pLen-1; i++ {
		patternConstraints[i] = P.EventConstraints{}
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
	// RightInformation is with distribution p, 1-p
	// RightFraction is 1.0
	// OverallInformation = rightInfo
	// confidence is p.
	// confidenceGain, infoDrop are not defined.
	// parentIndex is set to -1.
	patternCount, err := pattern.GetCount(patternConstraints, countType)
	if err != nil {
		return nil, err
	}
	var p float64
	if len(pattern.EventNames) == 1 {
		if countType == P.COUNT_TYPE_PER_USER {
			p = float64(patternCount) / float64(pattern.TotalUserCount)
		} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
			p = float64(patternCount) / float64(patternWrapper.GetTotalEventCount(reqId))
		}
	} else if len(pattern.EventNames) == 2 {
		c, ok := patternWrapper.GetCount(reqId, pattern.EventNames[:1],
			patternConstraints[:1], countType)
		if !ok {
			return nil, fmt.Errorf(fmt.Sprintf("Frequency missing for startEvent: %s", pattern.String()))
		}
		p = float64(patternCount) / float64(c)
	} else {
		return nil, fmt.Errorf(
			fmt.Sprintf("Unexpected root node pattern %s", pattern.String()))
	}
	var fcp, fcr uint
	if len(pattern.EventNames) == 1 {
		if startEventConstraints != nil {
			fcp, _ = allActiveUsers.GetCount([]P.EventConstraints{*startEventConstraints}, countType)
		} else {
			fcp, _ = allActiveUsers.GetCount([]P.EventConstraints{}, countType)
		}
		fcr, _ = pattern.GetCount(patternConstraints, countType)

	} else if len(pattern.EventNames) == 2 {
		fcp, _ = patternWrapper.GetCount(reqId, pattern.EventNames[:1],
			patternConstraints[:1], countType)
		fcr, _ = pattern.GetCount(patternConstraints, countType)
	}
	info := information([]float64{p, 1.0 - p})

	node := ItreeNode{
		Pattern:            pattern,
		NodeType:           NODE_TYPE_ROOT,
		PatternConstraints: patternConstraints,
		ParentIndex:        -1,
		RightInformation:   info,
		RightFraction:      1.0,
		OverallInformation: info,
		Confidence:         p,
		Fpp:                0.0,
		Fcp:                float64(fcp),
		Fcr:                float64(fcr),
	}
	return &node, nil
}

func (it *Itree) buildCategoricalGraphChildNode(
	childPattern *P.Pattern, categoricalPropertyName string,
	kLDistances []KLDistanceUnitInfo,
	nodeType int, parentIndex int,
	patternWrapper PatternServiceWrapperInterface,
	fpr float64, fpp float64, countType string) (*ItreeNode, error) {

	parentNode := it.Nodes[parentIndex]
	parentPattern := parentNode.Pattern
	parentConstraints := parentNode.PatternConstraints
	childLen := len(childPattern.EventNames)
	parentLen := len(parentPattern.EventNames)

	if parentNode.NodeType != NODE_TYPE_SEQUENCE && parentNode.NodeType != NODE_TYPE_ROOT {
		return nil, fmt.Errorf(fmt.Sprintf("Parent node of unexpected type: %v", parentNode))
	}

	if nodeType != NODE_TYPE_GRAPH_EVENT_PROPERTIES && nodeType != NODE_TYPE_GRAPH_USER_PROPERTIES {
		return nil, fmt.Errorf(fmt.Sprintf("Node of unexpected type: %s, %d", childPattern.String(), nodeType))
	}

	if childLen != parentLen {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Current Graph node with type property and pattern(%s) with propertyName(%s) and parent index(%d), pattern(%s) with constraints(%v) not compatible",
			childPattern.String(), categoricalPropertyName, parentIndex, parentPattern.String(), parentConstraints))
	}

	if parentLen == 1 && countType == P.COUNT_TYPE_PER_OCCURRENCE {
		return nil, fmt.Errorf("can't add graphs for root node of length 1 when counting by occurrence")
	}

	// Sort by decreasing order of absolute KLDistance contribution
	sort.SliceStable(kLDistances,
		func(i, j int) bool {
			distI := math.Abs(kLDistances[i].Distance)
			distJ := math.Abs(kLDistances[j].Distance)
			return (distI > distJ)
		})

	// Trim it to MAX_PROPERTIES_IN_GRAPH_NODE nodes.
	otherFcp := 0.0
	if len(kLDistances) > MAX_PROPERTIES_IN_GRAPH_NODE {
		kLDistances = kLDistances[:MAX_PROPERTIES_IN_GRAPH_NODE-1]
		// Add an other label at the end.
		totalFcp := 0.0
		totalFcr := 0.0
		for _, klu := range kLDistances {
			totalFcp += klu.Fcp
			totalFcr += klu.Fcr
		}
		otherFcp = fpp - totalFcp
		otherFcr := fpr - totalFcr
		otherPatternProb := otherFcp / fpp
		otherRuleProb := otherFcr / fpr

		otherKLDistanceUnit := KLDistanceUnitInfo{
			PropertyValue: OTHER_PROPERTY_VALUES_LABEL,
			Fpp:           fpp,
			Fpr:           fpr,
			Fcp:           otherFcp,
			Fcr:           otherFcr,
			Distance:      computeKLDistanceBits(otherPatternProb, otherRuleProb),
		}
		kLDistances = append(kLDistances, otherKLDistanceUnit)
	}
	totalKLDistance := 0.0
	for _, klu := range kLDistances {
		if klu.PropertyValue == OTHER_PROPERTY_VALUES_LABEL {
			// Do not include Other label in total KL Distance, since it does not add much value.
			continue
		}
		totalKLDistance += klu.Distance
	}
	node := ItreeNode{
		Pattern:            childPattern,
		PatternConstraints: parentConstraints,
		NodeType:           nodeType,
		ParentIndex:        parentIndex,
		InformationDrop:    totalKLDistance,
		Fpp:                fpp,
		Fpr:                fpr,
		OtherFcp:           otherFcp,
		PropertyName:       categoricalPropertyName,
		KLDistances:        kLDistances,
	}
	log.WithFields(log.Fields{"node": node.Pattern.String(), "nodeType": nodeType,
		"parent": parentPattern.String(),
		"fpr":    fpr, "fpp": fpp, "kLDistances": kLDistances,
		"PropertyName": categoricalPropertyName, "infoDrop": totalKLDistance}).Debug(
		"Built candidate graph child node.")
	return &node, nil
}

func (it *Itree) buildChildNode(reqId string,
	childPattern *P.Pattern, constraintToAdd *P.EventConstraints,
	nodeType int, parentIndex int,
	patternWrapper PatternServiceWrapperInterface,
	allActiveUsersPattern *P.Pattern,
	fpr float64, fpp float64, countType string) (*ItreeNode, error) {

	parentNode := it.Nodes[parentIndex]
	parentPattern := parentNode.Pattern
	parentConstraints := parentNode.PatternConstraints
	childLen := len(childPattern.EventNames)
	parentLen := len(parentPattern.EventNames)

	if parentNode.NodeType != NODE_TYPE_SEQUENCE && parentNode.NodeType != NODE_TYPE_ROOT {
		return nil, fmt.Errorf(fmt.Sprintf("Parent node of unexpected type: %v", parentNode))
	}

	if nodeType == NODE_TYPE_SEQUENCE && (parentLen != childLen-1 || parentLen < 1) {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Current node with type sequence and pattern(%s) with constraintToAdd(%v) and parent index(%d), pattern(%s) with constraints(%v) not compatible.",
			childPattern.String(), constraintToAdd, parentIndex, parentPattern.String(), parentConstraints))
	}

	if (nodeType == NODE_TYPE_EVENT_PROPERTY || nodeType == NODE_TYPE_USER_PROPERTY) && childLen != parentLen {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Current node with type property and pattern(%s) with constraintToAdd(%v) and parent index(%d), pattern(%s) with constraints(%v) not compatible",
			childPattern.String(), constraintToAdd, parentIndex, parentPattern.String(), parentConstraints))
	}

	if parentLen == 1 && countType == P.COUNT_TYPE_PER_OCCURRENCE && nodeType != NODE_TYPE_SEQUENCE {
		// Cannot add constraints to AllActiveUsers Pattern when counting by occurrence.
		return nil, nil
	}

	// Build Child constraints.
	childPatternConstraints := make([]P.EventConstraints, childLen)
	if nodeType == NODE_TYPE_SEQUENCE {
		if parentConstraints != nil {
			for i := 0; i < childLen-2; i++ {
				childPatternConstraints[i] = parentNode.PatternConstraints[i]
			}
			// childPattern has an extra event at childLen-2.
			// Minimum expected childLen is 2. Else this will crash.
			if childLen == 2 {
				// If parent is single event, then the constraints are attached to the
				// end event by default.
				childPatternConstraints[1] = parentNode.PatternConstraints[0]
			} else {
				childPatternConstraints[childLen-2] = P.EventConstraints{}
			}
			childPatternConstraints[childLen-1] = parentNode.PatternConstraints[childLen-2]
		} else {
			childPatternConstraints = nil
		}
	} else if nodeType == NODE_TYPE_EVENT_PROPERTY {
		for i := 0; i < childLen; i++ {
			if parentConstraints != nil {
				childPatternConstraints[i] = parentConstraints[i]
			} else {
				childPatternConstraints[i] = P.EventConstraints{
					EPNumericConstraints:     []P.NumericConstraint{},
					EPCategoricalConstraints: []P.CategoricalConstraint{},
					UPNumericConstraints:     []P.NumericConstraint{},
					UPCategoricalConstraints: []P.CategoricalConstraint{},
				}
			}
		}
		// Additional Event constraints are added to N-1st event.
		childPatternConstraints[childLen-2].EPCategoricalConstraints = append(
			childPatternConstraints[childLen-2].EPCategoricalConstraints,
			(*constraintToAdd).EPCategoricalConstraints...)
		childPatternConstraints[childLen-2].EPNumericConstraints = append(
			childPatternConstraints[childLen-2].EPNumericConstraints,
			(*constraintToAdd).EPNumericConstraints...)
	} else if nodeType == NODE_TYPE_USER_PROPERTY {
		for i := 0; i < childLen; i++ {
			if parentConstraints != nil {
				childPatternConstraints[i] = parentConstraints[i]
			} else {
				childPatternConstraints[i] = P.EventConstraints{
					EPNumericConstraints:     []P.NumericConstraint{},
					EPCategoricalConstraints: []P.CategoricalConstraint{},
					UPNumericConstraints:     []P.NumericConstraint{},
					UPCategoricalConstraints: []P.CategoricalConstraint{},
				}
			}
		}
		if childLen > 1 {
			// By default additional constraints are added to N-1st event.
			childPatternConstraints[childLen-2].UPCategoricalConstraints = append(
				childPatternConstraints[childLen-2].UPCategoricalConstraints,
				(*constraintToAdd).UPCategoricalConstraints...)
			childPatternConstraints[childLen-2].UPNumericConstraints = append(
				childPatternConstraints[childLen-2].UPNumericConstraints,
				(*constraintToAdd).UPNumericConstraints...)
		} else {
			// If only one node, additional user property constraints are added
			// to the last node.
			childPatternConstraints[childLen-1].UPCategoricalConstraints = append(
				childPatternConstraints[childLen-1].UPCategoricalConstraints,
				(*constraintToAdd).UPCategoricalConstraints...)
			childPatternConstraints[childLen-1].UPNumericConstraints = append(
				childPatternConstraints[childLen-1].UPNumericConstraints,
				(*constraintToAdd).UPNumericConstraints...)
		}
	}

	// Compute Child frequencies.
	var fcr, fcp float64
	if nodeType == NODE_TYPE_SEQUENCE {
		pCount, _ := childPattern.GetCount(childPatternConstraints, countType)
		fcr = float64(pCount)
		var subConstraints []P.EventConstraints
		if childPatternConstraints == nil {
			subConstraints = nil
		} else {
			subConstraints = childPatternConstraints[:childLen-1]
		}
		c, _ := patternWrapper.GetCount(reqId, childPattern.EventNames[:childLen-1], subConstraints, countType)
		fcp = float64(c)
	} else {
		pCount, _ := childPattern.GetCount(childPatternConstraints, countType)
		fcr = float64(pCount)
		if childLen > 1 {
			var subConstraints []P.EventConstraints
			if childPatternConstraints == nil {
				subConstraints = nil
			} else {
				subConstraints = childPatternConstraints[:childLen-1]
			}
			c, _ := patternWrapper.GetCount(reqId, childPattern.EventNames[:childLen-1], subConstraints, countType)
			fcp = float64(c)
		} else {
			if nodeType != NODE_TYPE_USER_PROPERTY || countType != P.COUNT_TYPE_PER_USER {
				return nil, fmt.Errorf(fmt.Sprintf(
					"Unexpected length of 1 for pattern %s, nodeType %d, constraintToAdd %v, countType: %s",
					childPattern.String(), nodeType, constraintToAdd, countType))
			}
			// frequency of child pattern is computed over allUserEvents after applying constraint.
			allUConstraints := []P.EventConstraints{
				*constraintToAdd,
			}
			c, _ := allActiveUsersPattern.GetPerUserCount(allUConstraints)
			fcp = float64(c)
		}
	}

	if fcp <= 0.0 || fcr <= 0.0 {
		//log.WithFields(log.Fields{"Child": childPattern.String(),
		//	"constraintToAdd": constraintToAdd, "parentConstraints": parentConstraints,
		//	"fcp": fcp, "fcr": fcr}).Debug("Child not frequent enough")
		return nil, nil
	}
	// Expected.
	// fpp >= fpr
	// fcp >= fcr
	// fpp >= fcp
	// fpr >= fcr
	if fpp < fpr || fcp < fcr || fpp < fcp || fpr < fcr {
		log.WithFields(log.Fields{"Child": childPattern.String(), "constraints": childPatternConstraints,
			"fpp": fpp, "fpr": fpr, "fcp": fcp, "fcr": fcr}).Debug("Inconsistent frequencies. Ignoring")
		return nil, nil
	}
	p := fcr / fcp
	rightInfo := information([]float64{p, 1.0 - p})
	rightFraction := fcp / fpp
	// Left information.
	var leftInfo float64
	if (fpp - fcp) > 0.0 {
		leftP := (fpr - fcr) / (fpp - fcp)
		if leftP > 1.0 {
			leftP = 1.0
		}
		leftInfo = information([]float64{leftP, 1.0 - leftP})
	} else {
		leftInfo = 0.0
	}
	overallInfo := rightFraction*rightInfo + (1-rightFraction)*leftInfo

	infoDrop := parentNode.RightInformation - overallInfo

	confidence := fcr / fcp
	confidenceGain := confidence - parentNode.Confidence

	addedConstraint := P.EventConstraints{}
	if constraintToAdd != nil {
		addedConstraint = *constraintToAdd
	}
	node := ItreeNode{
		Pattern:            childPattern,
		NodeType:           nodeType,
		PatternConstraints: childPatternConstraints,
		ParentIndex:        parentIndex,
		RightInformation:   rightInfo,
		RightFraction:      rightFraction,
		OverallInformation: overallInfo,
		InformationDrop:    infoDrop,
		Confidence:         confidence,
		ConfidenceGain:     confidenceGain,
		Fpp:                fpp,
		Fpr:                fpr,
		Fcp:                fcp,
		Fcr:                fcr,
		AddedConstraint:    addedConstraint,
	}
	//log.WithFields(log.Fields{"node": node.Pattern.String(), "constraints": constraints,
	//	"parent": parentPattern.String(), "parentConstraints": parentNode.PatternConstraints,
	//	"fcr": fcr, "fcp": fcp, "fpr": fpr, "fpp": fpp, "OverallInfo": overallInfo, "infoDrop": infoDrop}).Debug(
	//	"Built candidate child node.")
	return &node, nil
}

func (it *Itree) addNode(node *ItreeNode) int {
	node.Index = len(it.Nodes)
	it.Nodes = append(it.Nodes, node)
	log.WithFields(log.Fields{
		"node":            node.Pattern.String(),
		"AddedConstraint": node.AddedConstraint,
		"nodeType":        node.NodeType,
		"constraints":     node.PatternConstraints,
		"index":           node.Index,
		"parentIndex":     node.ParentIndex,
		"Fpp":             node.Fpp,
		"Fpr":             node.Fpr,
		"Fcp":             node.Fcp,
		"Fcr":             node.Fcr,
		"OverallInfo":     node.OverallInformation,
		"infoDrop":        node.InformationDrop}).Info(
		"Added node to Itree.")
	return node.Index
}

func isChildSequenceBlacklisted(crmEvents map[string]bool, ToBeAdded string, preSequence []string, goalEvent string) bool {
	// If goalEvent has restrictions
	// check what all events are in blacklisting and then check the ToBeadded event and then decide
	// If ToBeadded is in restrictions
	// check what all are there in the presequence, then decide if it has to be added

	preSequenceMap := make(map[string]bool)
	for _, eventName := range preSequence {
		preSequenceMap[eventName] = true
	}

	BLACKLISTED_JOURNEYS := map[string][]string{
		// "$hubspot_contact_created": []string{"$hubspot_contact_updated","ANY_CRM"},
		// "$sf_opportunity_created":  []string{"$sf_opportunity_updated","ANY_CRM"},
		// "$sf_lead_created":         []string{"$sf_lead_updated","ANY_CRM"},
		"Deal Won": []string{"Deal Created"},
		"*":        []string{"ANY_CRM"},
	}

	BLACKLISTED_JOURNEYS["*"] = append(BLACKLISTED_JOURNEYS["*"], model.AllowedEventNamesForHubspot...)
	BLACKLISTED_JOURNEYS["*"] = append(BLACKLISTED_JOURNEYS["*"], model.AllowedEventNamesForSalesforce...)

	for _, blacklisted := range BLACKLISTED_JOURNEYS["*"] {
		if ToBeAdded == blacklisted {
			return true
		}
		if blacklisted == "ANY_CRM" && crmEvents[ToBeAdded] {
			return true
		}
	}
	for _, blacklisted := range BLACKLISTED_JOURNEYS[goalEvent] {
		if ToBeAdded == blacklisted {
			return true
		}
		if blacklisted == "ANY_CRM" && crmEvents[ToBeAdded] {
			return true
		}
	}

	for _, blacklisted := range BLACKLISTED_JOURNEYS[ToBeAdded] {
		if preSequenceMap[blacklisted] {
			return true
		}
	}
	return false
}

func isChildSequence(parent []string, child []string, crmEvents map[string]bool) bool {
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
	preSeq := make([]string, 0)
	if pLen > 1 {
		preSeq = parent[0 : pLen-1]
	}
	if isChildSequenceBlacklisted(crmEvents, child[cLen-2], preSeq, parent[pLen-1]) {
		return false
	}
	return true
}

func (it *Itree) buildAndAddSequenceChildNodes(reqId string,
	parentNode *ItreeNode, candidatePattens []*P.Pattern,
	patternWrapper PatternServiceWrapperInterface,
	allActiveUsersPattern *P.Pattern, countType string, projectId uint64) ([]*ItreeNode, error) {

	parentPattern := parentNode.Pattern
	peLen := len(parentPattern.EventNames)
	count, err := parentPattern.GetCount(parentNode.PatternConstraints, countType)
	if err != nil {
		return nil, err
	}
	fpr := float64(count)

	var fpp float64
	if peLen == 1 {
		if countType == P.COUNT_TYPE_PER_USER {
			// Parent is root node. Count is all users.
			fpp = float64(parentPattern.TotalUserCount)
		} else {
			// Parent is root node. Count is all users.
			fpp = float64(patternWrapper.GetTotalEventCount(reqId))
		}
	} else {
		var subParentConstraints []P.EventConstraints
		if parentNode.PatternConstraints == nil {
			subParentConstraints = nil
		} else {
			subParentConstraints = parentNode.PatternConstraints[:peLen-1]
		}
		c, ok := patternWrapper.GetCount(
			reqId, parentPattern.EventNames[:peLen-1], subParentConstraints, countType)
		if !ok {
			return nil, fmt.Errorf(fmt.Sprintf(
				"Frequency missing for pattern sequence %s", parentPattern.String()))
		}
		fpp = float64(c)
	}

	childNodes := []*ItreeNode{}
	if fpp <= 0.0 || fpr <= 0.0 {
		//log.WithFields(log.Fields{"Parent": parentPattern.String(), "fpr": fpr, "fpp": fpp}).Debug(
		//	"Parent not frequent enough")
		return childNodes, nil
	}
	for _, p := range candidatePattens {
		if !isChildSequence(parentNode.Pattern.EventNames, p.EventNames, nil) {
			continue
		}

		if cNode, err := it.buildChildNode(reqId,
			p, nil, NODE_TYPE_SEQUENCE,
			parentNode.Index, patternWrapper, allActiveUsersPattern, fpr, fpp, countType); err != nil {
			log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build child node")
			continue
		} else {
			if cNode == nil {
				continue
			}
			childNodes = append(childNodes, cNode)
		}
	}
	// Sort in decreasing order of infoDrop.
	sort.SliceStable(childNodes,
		func(i, j int) bool {
			return (childNodes[i].InformationDrop > childNodes[j].InformationDrop)
		})

	// Only top MAX_SEQUENCE_CHILD_NODES in order of drop in GiniImpurity are selected.
	if len(childNodes) > MAX_SEQUENCE_CHILD_NODES {
		childNodes = childNodes[:MAX_SEQUENCE_CHILD_NODES]
	}

	addedChildNodes := []*ItreeNode{}
	for _, cNode := range childNodes {
		if cNode.InformationDrop <= 0.0 {
			continue
		}
		it.addNode(cNode)
		addedChildNodes = append(addedChildNodes, cNode)
	}

	return addedChildNodes, nil
}

func getPropertyNamesMapFromConstraints(
	patternConstraints []P.EventConstraints) *map[string]bool {
	seenPropertyConstraints := make(map[string]bool)
	// Initialize seen properties with that of parent properties.
	for _, pcs := range patternConstraints {
		for _, ecc := range pcs.EPCategoricalConstraints {
			seenPropertyConstraints[ecc.PropertyName] = true
		}
		for _, enc := range pcs.EPNumericConstraints {
			seenPropertyConstraints[enc.PropertyName] = true
		}
		for _, ucc := range pcs.UPCategoricalConstraints {
			seenPropertyConstraints[ucc.PropertyName] = true
		}
		for _, upc := range pcs.UPNumericConstraints {
			seenPropertyConstraints[upc.PropertyName] = true
		}
	}
	return &seenPropertyConstraints
}

func getPropertyTypeOfAddedConstraint(pcs P.EventConstraints) bool {
	var isCat bool
	for range pcs.EPCategoricalConstraints {
		isCat = true
	}
	for range pcs.EPNumericConstraints {
		isCat = false
	}
	for range pcs.UPCategoricalConstraints {
		isCat = true
	}
	for range pcs.UPNumericConstraints {
		isCat = false
	}
	return isCat
}
func (it *Itree) buildCategoricalPropertyChildNodes(reqId string,
	categoricalPropertyKeyValues []string,
	nodeType int, maxNumProperties int, maxNumValues int,
	parentNode *ItreeNode, patternWrapper PatternServiceWrapperInterface,
	allActiveUsersPattern *P.Pattern, pLen int, fpr float64, fpp float64,
	countType string) []*ItreeNode {
	propertyChildNodes := []*ItreeNode{}
	numP := 0
	seenProperties := getPropertyNamesMapFromConstraints(parentNode.PatternConstraints)
	for _, propertyName := range categoricalPropertyKeyValues {
		if numP > maxNumProperties {
			break
		}
		if _, found := (*seenProperties)[propertyName]; found {
			continue
		}
		if nodeType == NODE_TYPE_EVENT_PROPERTY && U.ShouldIgnoreItreeProperty(propertyName) {
			continue
		}
		if nodeType == NODE_TYPE_USER_PROPERTY && U.ShouldIgnoreItreeProperty(propertyName) {
			continue
		}
		numP++
		numVal := 0
		attributeIndex := int(0)
		if pLen == 1 {
			attributeIndex = 0
		} else {
			attributeIndex = pLen - 2
		}
		// Compute KL Distance of child vs parent.
		seenValues := parentNode.Pattern.GetPerUserUserPropertyValues(attributeIndex, propertyName)
		klDistanceUnits := []KLDistanceUnitInfo{}
		for _, value := range seenValues {
			if numVal > maxNumValues {
				break
			}
			if value == "" || value == "None" {
				continue
			}
			numVal++
			constraintToAdd := P.EventConstraints{
				EPNumericConstraints:     []P.NumericConstraint{},
				EPCategoricalConstraints: []P.CategoricalConstraint{},
				UPNumericConstraints:     []P.NumericConstraint{},
				UPCategoricalConstraints: []P.CategoricalConstraint{},
			}
			categoricalConstraint := []P.CategoricalConstraint{
				{
					PropertyName:  propertyName,
					PropertyValue: value,
				},
			}
			if nodeType == NODE_TYPE_EVENT_PROPERTY {
				constraintToAdd.EPCategoricalConstraints = categoricalConstraint
			} else if nodeType == NODE_TYPE_USER_PROPERTY {
				constraintToAdd.UPCategoricalConstraints = categoricalConstraint
			}

			if cNode, err := it.buildChildNode(reqId,
				parentNode.Pattern, &constraintToAdd, nodeType,
				parentNode.Index, patternWrapper, allActiveUsersPattern, fpr, fpp, countType); err != nil {
				log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build child node")
				continue
			} else {
				if cNode == nil {
					continue
				}
				propertyChildNodes = append(propertyChildNodes, cNode)
				// Distance computation for graph node.
				// Distance is the measure of change in distribution between the pattern and the rule.
				patternProb := cNode.Fcp / fpp
				ruleProb := cNode.Fcr / fpr
				klDistanceUnits = append(klDistanceUnits, KLDistanceUnitInfo{
					PropertyValue: value,
					// The distance / divergence of rule Distribution from pattern Distribution wrt propertyName in bits.
					Distance: computeKLDistanceBits(patternProb, ruleProb),
					Fpp:      cNode.Fpp,
					Fpr:      cNode.Fpr,
					Fcp:      cNode.Fcp,
					Fcr:      cNode.Fcr,
				})
			}
		}

		if len(klDistanceUnits) > 0 {
			// Append None label if there is discrepancy in count.
			totalFcp := 0.0
			totalFcr := 0.0
			for _, klu := range klDistanceUnits {
				totalFcp += klu.Fcp
				totalFcr += klu.Fcr
			}
			noneFcp := fpp - totalFcp
			noneFcr := fpr - totalFcr
			if noneFcp > 0.0 || noneFcr > 0.0 {
				nonePatternProb := noneFcp / fpp
				noneRuleProb := noneFcr / fpr

				noneKLDistanceUnit := KLDistanceUnitInfo{
					PropertyValue: NONE_PROPERTY_VALUES_LABEL,
					Fpp:           fpp,
					Fpr:           fpr,
					Fcp:           noneFcp,
					Fcr:           noneFcr,
					Distance:      computeKLDistanceBits(nonePatternProb, noneRuleProb),
				}
				klDistanceUnits = append(klDistanceUnits, noneKLDistanceUnit)
			}

			/* Not showing graph results for now as it was confusing for users to interpret.
			nodeGraphType := NODE_TYPE_GRAPH_USER_PROPERTIES
			if nodeType == NODE_TYPE_EVENT_PROPERTY {
				nodeGraphType = NODE_TYPE_GRAPH_EVENT_PROPERTIES
			}
			if cGraphNode, err := it.buildCategoricalGraphChildNode(
				parentNode.Pattern, propertyName, klDistanceUnits, nodeGraphType,
				parentNode.Index, patternWrapper, fpr, fpp, countType); err != nil {
				log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build graph child node")
			} else {
				propertyChildNodes = append(propertyChildNodes, cGraphNode)
			}*/
		}
	}
	return propertyChildNodes
}

func (it *Itree) buildNumericalPropertyChildNodes(reqId string,
	numericPropertyKeys []string, nodeType int, maxNumProperties int,
	parentNode *ItreeNode, patternWrapper PatternServiceWrapperInterface,
	allActiveUsersPattern *P.Pattern, pLen int, fpr float64, fpp float64,
	countType string) []*ItreeNode {

	return []*ItreeNode{}
	/*
		parentPattern := parentNode.Pattern
		propertyChildNodes := []*ItreeNode{}
		numP := 0
		seenProperties := getPropertyNamesMapFromConstraints(parentNode.PatternConstraints)
		for _, propertyName := range numericPropertyKeys {
			if numP > maxNumProperties {
				break
			}
			if _, found := (*seenProperties)[propertyName]; found {
				continue
			}
			if nodeType == NODE_TYPE_EVENT_PROPERTY && U.ShouldIgnoreItreeProperty(propertyName) {
				continue
			}
			if nodeType == NODE_TYPE_USER_PROPERTY && U.ShouldIgnoreItreeProperty(propertyName) {
				continue
			}
			numP++
			binRanges := [][2]float64{}
			isPredefined := false
			if nodeType == NODE_TYPE_EVENT_PROPERTY {
				if countType == P.COUNT_TYPE_PER_USER {
					binRanges, isPredefined = parentPattern.GetPerUserEventPropertyRanges(pLen-2, propertyName)
				} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
					binRanges, isPredefined = parentPattern.GetPerOccurrenceEventPropertyRanges(pLen-2, propertyName)
				}
			} else if nodeType == NODE_TYPE_USER_PROPERTY {
				if countType == P.COUNT_TYPE_PER_USER {
					binRanges, isPredefined = parentPattern.GetPerUserUserPropertyRanges(pLen-2, propertyName)
				} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
					binRanges, isPredefined = parentPattern.GetPerOccurrenceUserPropertyRanges(pLen-2, propertyName)
				}
			}

			for _, pRange := range binRanges {
				minValue := pRange[0]
				maxValue := pRange[1]
				// Try three combinations of constraints. v < min, min < v < max, v > max
				constraint1ToAdd := P.EventConstraints{
					EPNumericConstraints:     []P.NumericConstraint{},
					EPCategoricalConstraints: []P.CategoricalConstraint{},
					UPNumericConstraints:     []P.NumericConstraint{},
					UPCategoricalConstraints: []P.CategoricalConstraint{},
				}
				if !isPredefined {
					// Show only interval constraint for predefined bin ranges.
					lesserThanConstraint := []P.NumericConstraint{
						P.NumericConstraint{
							PropertyName: propertyName,
							LowerBound:   -math.MaxFloat64,
							UpperBound:   minValue,
						},
					}
					if nodeType == NODE_TYPE_EVENT_PROPERTY {
						constraint1ToAdd.EPNumericConstraints = lesserThanConstraint
					} else if nodeType == NODE_TYPE_USER_PROPERTY {
						constraint1ToAdd.UPNumericConstraints = lesserThanConstraint
					}
				}

				constraint2ToAdd := P.EventConstraints{
					EPNumericConstraints:     []P.NumericConstraint{},
					EPCategoricalConstraints: []P.CategoricalConstraint{},
					UPNumericConstraints:     []P.NumericConstraint{},
					UPCategoricalConstraints: []P.CategoricalConstraint{},
				}
				intervalConstraint := []P.NumericConstraint{
					P.NumericConstraint{
						PropertyName: propertyName,
						LowerBound:   minValue,
						UpperBound:   maxValue,
					},
				}
				if nodeType == NODE_TYPE_EVENT_PROPERTY {
					constraint2ToAdd.EPNumericConstraints = intervalConstraint
				} else if nodeType == NODE_TYPE_USER_PROPERTY {
					constraint2ToAdd.UPNumericConstraints = intervalConstraint
				}

				constraint3ToAdd := P.EventConstraints{
					EPNumericConstraints:     []P.NumericConstraint{},
					EPCategoricalConstraints: []P.CategoricalConstraint{},
					UPNumericConstraints:     []P.NumericConstraint{},
					UPCategoricalConstraints: []P.CategoricalConstraint{},
				}
				if !isPredefined {
					greaterThanConstraint := []P.NumericConstraint{
						P.NumericConstraint{
							PropertyName: propertyName,
							LowerBound:   maxValue,
							UpperBound:   math.MaxFloat64,
						},
					}
					if nodeType == NODE_TYPE_EVENT_PROPERTY {
						constraint3ToAdd.EPNumericConstraints = greaterThanConstraint
					} else if nodeType == NODE_TYPE_USER_PROPERTY {
						constraint3ToAdd.UPNumericConstraints = greaterThanConstraint
					}
				}

				for _, constraintToAdd := range []P.EventConstraints{
					constraint1ToAdd, constraint2ToAdd, constraint3ToAdd} {
					if cNode, err := it.buildChildNode(reqId,
						parentPattern, &constraintToAdd, nodeType,
						parentNode.Index, patternWrapper, allActiveUsersPattern,
						fpr, fpp, countType); err != nil {
						log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build child node")
						continue
					} else {
						if cNode == nil {
							continue
						}
						propertyChildNodes = append(propertyChildNodes, cNode)
					}
				}
			}
		}
		return propertyChildNodes
	*/
}

func (it *Itree) buildAndAddPropertyChildNodes(reqId string,
	parentNode *ItreeNode, allActiveUsersPattern *P.Pattern,
	patternWrapper PatternServiceWrapperInterface, countType string) ([]*ItreeNode, error) {
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

	userCatProperties, userNumProperties, eventCatProperties, eventNumProperties := extractProperties(parentNode, nil, nil)

	var eventName string
	if pLen == 1 {
		if countType == P.COUNT_TYPE_PER_USER {
			// Parent is root node. Count is all users.
			fpp = float64(parentPattern.TotalUserCount)
		} else {
			// Parent is root node. Count is all users.
			fpp = float64(patternWrapper.GetTotalEventCount(reqId))
		}
	} else {
		eventName = parentPattern.EventNames[pLen-2]
		var subConstraints []P.EventConstraints
		if parentConstraints == nil {
			subConstraints = nil
		} else {
			subConstraints = parentConstraints[:pLen-1]
		}
		c, ok := patternWrapper.GetCount(reqId, parentPattern.EventNames[:pLen-1], subConstraints, countType)
		if !ok {
			return nil, fmt.Errorf(fmt.Sprintf("Frequency missing for pattern sequence %s", parentPattern.String()))
		}
		fpp = float64(c)
	}
	count, err := parentPattern.GetCount(parentConstraints, countType)
	if err != nil {
		return nil, err
	}
	fpr = float64(count)

	// Map out the current constraints applied on N-1st event in the pattern.
	eventConstraints := P.EventConstraints{}
	if parentConstraints != nil && pLen > 1 {
		eventConstraints = parentConstraints[pLen-2]
	}
	if len(eventConstraints.EPCategoricalConstraints) > 0 ||
		len(eventConstraints.EPNumericConstraints) > 0 ||
		len(eventConstraints.UPCategoricalConstraints) > 0 ||
		len(eventConstraints.UPNumericConstraints) > 0 {
		log.Errorf(fmt.Sprintf(
			"Pattern %s already has constraints %v on event %s",
			parentPattern.String(), eventConstraints, eventName))
		return []*ItreeNode{}, nil
	}

	childNodes := []*ItreeNode{}
	if fpp <= 0.0 || fpr <= 0.0 {
		//log.WithFields(log.Fields{"Parent": parentPattern.String(), "parentConstraints": parentConstraints,
		//	"fpr": fpr, "fpp": fpp}).Debug("Parent not frequent enough")
		return childNodes, nil
	}

	// Add children by splitting on constraints on categorical properties.
	pLen = len(parentPattern.EventNames)
	MAX_CAT_PROPERTIES_EVALUATED := 100
	MAX_CAT_VALUES_EVALUATED := 100

	if pLen > 1 {
		childNodes = append(childNodes, it.buildCategoricalPropertyChildNodes(reqId,
			eventCatProperties, NODE_TYPE_EVENT_PROPERTY, MAX_CAT_PROPERTIES_EVALUATED,
			MAX_CAT_VALUES_EVALUATED, parentNode, patternWrapper, allActiveUsersPattern, pLen, fpr, fpp, countType)...)
	}

	if pLen > 1 || countType == P.COUNT_TYPE_PER_USER {
		childNodes = append(childNodes, it.buildCategoricalPropertyChildNodes(reqId,
			userCatProperties, NODE_TYPE_USER_PROPERTY, MAX_CAT_PROPERTIES_EVALUATED,
			MAX_CAT_VALUES_EVALUATED, parentNode, patternWrapper, allActiveUsersPattern, pLen, fpr, fpp, countType)...)
	}

	// Add children by splitting on constraints on categorical properties.
	pLen = len(parentPattern.EventNames)
	MAX_NUM_PROPERTIES_EVALUATED := 100
	if pLen > 1 {
		childNodes = append(childNodes, it.buildNumericalPropertyChildNodes(reqId,
			eventNumProperties, NODE_TYPE_EVENT_PROPERTY, MAX_NUM_PROPERTIES_EVALUATED,
			parentNode, patternWrapper, allActiveUsersPattern, pLen, fpr, fpp, countType)...)
	}
	if pLen > 1 || countType == P.COUNT_TYPE_PER_USER {
		childNodes = append(childNodes, it.buildNumericalPropertyChildNodes(reqId,
			userNumProperties, NODE_TYPE_USER_PROPERTY, MAX_NUM_PROPERTIES_EVALUATED,
			parentNode, patternWrapper, allActiveUsersPattern, pLen, fpr, fpp, countType)...)
	}

	// Sort in decreasing order of infoDrop.
	sort.SliceStable(childNodes,
		func(i, j int) bool {
			return (childNodes[i].InformationDrop > childNodes[j].InformationDrop)
		})

	addedChildNodes := []*ItreeNode{}
	numAddedChildNodes := 0
	// Initialize seen properties with that of parent properties.
	seenPropertyConstraints := getPropertyNamesMapFromConstraints(
		parentNode.PatternConstraints)

	for _, cNode := range childNodes {
		if cNode.InformationDrop <= 0.0 {
			continue
		}
		// Dedup repeated constraints on different values.
		childPropertyConstraintsMap := getPropertyNamesMapFromConstraints(
			[]P.EventConstraints{cNode.AddedConstraint})
		isDup := false
		for childPropertyName, _ := range *childPropertyConstraintsMap {
			if _, found := (*seenPropertyConstraints)[childPropertyName]; found {
				isDup = true
				break
			}
		}
		if isDup {
			continue
		}
		it.addNode(cNode)
		addedChildNodes = append(addedChildNodes, cNode)
		for pc, _ := range *childPropertyConstraintsMap {
			(*seenPropertyConstraints)[pc] = true
		}
		numAddedChildNodes++
		// Only top MAX_PROPERTY_CHILD_NODES in order of drop in GiniImpurity are selected.
		if numAddedChildNodes >= MAX_PROPERTY_CHILD_NODES {
			break
		}
	}
	return addedChildNodes, nil
}

func BuildNewItree(reqId,
	startEvent string, startEventConstraints *P.EventConstraints,
	endEvent string, endEventConstraints *P.EventConstraints,
	patternWrapper PatternServiceWrapperInterface, countType string, projectId uint64) (*Itree, error) {
	if endEvent == "" {
		return nil, fmt.Errorf("missing end event")
	}

	itreeNodes := []*ItreeNode{}
	itree := Itree{
		Nodes:    itreeNodes,
		EndEvent: endEvent,
	}

	var rootNodePattern *P.Pattern = nil
	var allActiveUsersPattern *P.Pattern = nil
	startTime := time.Now().Unix()
	candidatePatterns, err := patternWrapper.GetAllPatterns(reqId, startEvent, endEvent)
	endTime := time.Now().Unix()
	log.WithFields(log.Fields{
		"time_taken": endTime - startTime}).Error("GetAllPatterns Time taken.")
	if err != nil {
		return nil, err
	}
	allActiveUsersPattern = patternWrapper.GetPattern(reqId, []string{U.SEN_ALL_ACTIVE_USERS})
	if allActiveUsersPattern == nil {
		return nil, fmt.Errorf("all active users pattern not found")
	}

	for _, p := range candidatePatterns {
		pLen := len(p.EventNames)
		if (startEvent == "" && pLen == 1) || (startEvent != "" && pLen == 2) {
			if count, err := p.GetCount(constructPatternConstraints(
				pLen, startEventConstraints, endEventConstraints), countType); count > 0 {
				rootNodePattern = p
			} else if err != nil {
				return nil, err
			}
		}
	}

	startTime = time.Now().Unix()
	if rootNodePattern == nil {
		return nil, fmt.Errorf("root node not found or frequency 0")
	}
	queue := []*ItreeNode{}

	rootNode, err := itree.buildRootNode(
		reqId, allActiveUsersPattern, rootNodePattern, startEventConstraints,
		endEventConstraints, patternWrapper, countType)
	if err != nil {
		return nil, err
	}

	itree.addNode(rootNode)
	queue = append(queue, rootNode)
	numNodesEvaluated := 0

	for len(queue) > 0 && numNodesEvaluated < MAX_NODES_TO_EVALUATE {
		numNodesEvaluated++
		parentNode := queue[0]
		if _, err := itree.buildAndAddPropertyChildNodes(reqId,
			parentNode, allActiveUsersPattern, patternWrapper, countType); err != nil {
			log.Errorf(fmt.Sprintf("%s", err))
			return nil, err
		}

		if sequenceChildNodes, err := itree.buildAndAddSequenceChildNodes(reqId, parentNode,
			candidatePatterns, patternWrapper, allActiveUsersPattern, countType, projectId); err != nil {
			return nil, err
		} else {
			// Only sequenceChildNodes are explored further.
			queue = append(queue, sequenceChildNodes...)
		}
		queue = queue[1:]
	}
	endTime = time.Now().Unix()
	log.WithFields(log.Fields{
		"time_taken": endTime - startTime}).Error("Building Tree Time taken.")

	//log.WithFields(log.Fields{"itree": itree}).Info("Returning Itree.")
	return &itree, nil
}

var debugCounts map[string]int

func BuildNewItreeV1(reqId string,
	startEvent string, startEventConstraints *P.EventConstraints,
	endEvent string, endEventConstraints *P.EventConstraints,
	patternWrapper PatternServiceWrapperInterface, countType string, debugKey string, debugParams map[string]string, projectId uint64,
	includedEventProperties map[string]bool, includedUserProperties map[string]bool, includedEvents map[string]bool) (*Itree, error, interface{}) {
	if endEvent == "" {
		return nil, fmt.Errorf("missing end event"), nil
	}

	if projectId == 659 {
		MAX_CAT_PROPERTIES_EVALUATED = 20
		MAX_CAT_VALUES_EVALUATED = 20
		MAX_NUM_PROPERTIES_EVALUATED = 20
	} else {
		MAX_CAT_PROPERTIES_EVALUATED = 100
		MAX_CAT_VALUES_EVALUATED = 100
		MAX_NUM_PROPERTIES_EVALUATED = 100
	}
	debugCounts = make(map[string]int)
	itreeNodes := []*ItreeNode{}
	itree := Itree{
		Nodes:    itreeNodes,
		EndEvent: endEvent,
	}

	events, _ := store.GetStore().GetSmartEventFilterEventNames(projectId, true)
	crmEvents := make(map[string]bool)
	for _, event := range events {
		crmEvents[event.Name] = true
	}
	var rootNodePattern *P.Pattern = nil
	var allActiveUsersPattern *P.Pattern = nil
	startTime := time.Now().Unix()
	candidatePatterns, err := patternWrapper.GetAllPatterns(reqId, startEvent, endEvent)
	endTime := time.Now().Unix()
	log.WithFields(log.Fields{
		"time_taken": endTime - startTime}).Error("GetAllPatterns Time taken.")
	log.WithFields(log.Fields{
		"time_taken": endTime - startTime}).Error("explain_debug_GetAllPatterns")
	debugCounts["total_patterns"] = len(candidatePatterns)

	var len1PatternCount, len2PatternCount, len3PatternCount int
	for _, pattern := range candidatePatterns {
		pattern.GenFrequentProperties()
		if len(pattern.EventNames) == 1 {
			len1PatternCount++
		}
		if len(pattern.EventNames) == 2 {
			len2PatternCount++
		}
		if len(pattern.EventNames) == 3 {
			len3PatternCount++
		}
	}
	debugCounts["len1_patterns"] = len1PatternCount
	debugCounts["len2_patterns"] = len2PatternCount
	debugCounts["len3_patterns"] = len3PatternCount

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error()}).Error("explain_debug_GetAllPatterns")
		return nil, err, nil
	}
	// allActiveUsersPattern = patternWrapper.GetPattern(reqId, []string{U.SEN_ALL_ACTIVE_USERS})
	// if allActiveUsersPattern == nil {
	// 	return nil, fmt.Errorf("all active users pattern not found"), nil
	// }

	allActiveUsersPattern = patternWrapper.GetPattern(reqId, []string{U.SEN_ALL_ACTIVE_USERS})
	if allActiveUsersPattern == nil {
		return nil, fmt.Errorf("all active users pattern not found"), nil
	}

	for _, p := range candidatePatterns {
		pLen := len(p.EventNames)
		if (startEvent == "" && pLen == 1) || (startEvent != "" && pLen == 2) {
			if count, err := p.GetCount(constructPatternConstraints(
				pLen, startEventConstraints, endEventConstraints), countType); count > 0 {
				rootNodePattern = p
			} else if err != nil {
				return nil, err, nil
			}
		}
	}

	startTime = time.Now().Unix()
	if rootNodePattern == nil {
		return nil, fmt.Errorf("root node not found or frequency 0"), nil
	}

	type LevelItreeNodeTuple struct {
		node  *ItreeNode
		level int
	}
	queue := []*LevelItreeNodeTuple{}

	rootNode, err := itree.buildRootNode(
		reqId, allActiveUsersPattern, rootNodePattern, startEventConstraints,
		endEventConstraints, patternWrapper, countType)
	if err != nil {
		return nil, err, nil
	}

	itree.addNode(rootNode)
	queue = append(queue, &LevelItreeNodeTuple{node: rootNode, level: 0})
	numNodesEvaluated := 0
	var debugData interface{}
	var campaignNodesDuration, campaignNodesCount, sequenceNodesDuration, sequenceNodesCount, attributeNodesDuration, attributeNodesCount int
	for len(queue) > 0 && numNodesEvaluated < MAX_NODES_TO_EVALUATE {
		numNodesEvaluated++
		parentNode := queue[0]
		if parentNode.node.NodeType == NODE_TYPE_USER_PROPERTY || parentNode.node.NodeType == NODE_TYPE_EVENT_PROPERTY || parentNode.node.NodeType == NODE_TYPE_ROOT {

			startDateTime := time.Now()
			if attributeChildNodes, err, debugInfo := itree.buildAndAddPropertyChildNodesV1(reqId,
				parentNode.node, allActiveUsersPattern, patternWrapper, countType, parentNode.level, debugKey, debugParams["PropertyName"], debugParams["PropertyValue"], startEventConstraints, endEventConstraints, includedEventProperties, includedUserProperties, projectId); err != nil {
				log.Errorf(fmt.Sprintf("%s", err))
				return nil, err, nil
			} else {
				if debugInfo != nil {
					debugData = debugInfo
				}
				endDateTime := time.Now()
				attributeNodesCount++
				attributeNodesDuration += int(endDateTime.Sub(startDateTime).Milliseconds())
				debugKey := fmt.Sprintf("itree_attribute_count_level%v", parentNode.level)
				debugCounts[debugKey] = debugCounts[debugKey] + len(attributeChildNodes)
				for _, childNode := range attributeChildNodes {
					if parentNode.level < 1 {
						queue = append(queue, &LevelItreeNodeTuple{node: childNode, level: parentNode.level + 1})
					}
				}
			}
		}

		if parentNode.node.NodeType == NODE_TYPE_SEQUENCE || parentNode.node.NodeType == NODE_TYPE_ROOT || parentNode.node.NodeType == NODE_TYPE_CAMPAIGN {
			startDateTime := time.Now()
			if sequenceChildNodes, err := itree.buildAndAddSequenceChildNodesV1(reqId, parentNode.node,
				candidatePatterns, patternWrapper, allActiveUsersPattern, countType, startEventConstraints, endEventConstraints, crmEvents, includedEvents); err != nil {
				return nil, err, nil
			} else {
				endDateTime := time.Now()
				sequenceNodesCount++
				sequenceNodesDuration += int(endDateTime.Sub(startDateTime).Milliseconds())
				debugKey := fmt.Sprintf("itree_sequence_count_level%v", parentNode.level)
				debugCounts[debugKey] = debugCounts[debugKey] + len(sequenceChildNodes)
				for _, childNode := range sequenceChildNodes {
					if parentNode.level < 1 {
						queue = append(queue, &LevelItreeNodeTuple{node: childNode, level: parentNode.level + 1})
					}
				}
			}
		}

		if parentNode.node.NodeType == NODE_TYPE_CAMPAIGN || parentNode.node.NodeType == NODE_TYPE_ROOT {
			startDateTime := time.Now()
			if sequenceChildNodes, err := itree.buildAndAddCampaignChildNodesV1(reqId, parentNode.node,
				candidatePatterns, patternWrapper, allActiveUsersPattern, countType, startEventConstraints, endEventConstraints, crmEvents); err != nil {
				return nil, nil, err
			} else {
				endDateTime := time.Now()
				campaignNodesCount++
				campaignNodesDuration += int(endDateTime.Sub(startDateTime).Milliseconds())
				debugKey := fmt.Sprintf("itree_campaign_count_level%v", parentNode.level)
				debugCounts[debugKey] = debugCounts[debugKey] + len(sequenceChildNodes)
				for _, childNode := range sequenceChildNodes {
					if parentNode.level < 1 {
						queue = append(queue, &LevelItreeNodeTuple{node: childNode, level: parentNode.level + 1})
					}
				}
			}
		}
		queue = queue[1:]
	}
	endTime = time.Now().Unix()
	log.WithFields(log.Fields{
		"time_taken": endTime - startTime}).Error("explain_debug_Building Tree Time taken")
	log.WithFields(log.Fields{
		"time_taken": sequenceNodesDuration, "count": sequenceNodesCount}).Error("explain_debug_sequence")
	log.WithFields(log.Fields{
		"time_taken": campaignNodesDuration, "count": campaignNodesCount}).Error("explain_debug_campaign")
	log.WithFields(log.Fields{
		"time_taken": attributeNodesDuration, "count": attributeNodesCount}).Error("explain_debug_attribute")
	log.WithFields(log.Fields{
		"debug_counts": debugCounts}).Error("explain_debug")
	//log.WithFields(log.Fields{"itree": itree}).Info("Returning Itree.")
	return &itree, nil, debugData
}

func extractProperty(constraint P.EventConstraints) map[string]string {
	properties := make(map[string]string)
	for _, property := range constraint.EPNumericConstraints {
		properties[property.PropertyName] = ""
	}
	for _, property := range constraint.EPCategoricalConstraints {
		properties[property.PropertyName] = property.PropertyValue
	}
	for _, property := range constraint.UPNumericConstraints {
		properties[property.PropertyName] = ""
	}
	for _, property := range constraint.UPCategoricalConstraints {
		properties[property.PropertyName] = property.PropertyValue
	}
	return properties
}

func extractProperties(parentNode *ItreeNode, includedEventProperties map[string]bool, includedUserProperties map[string]bool) ([]string, []string, []string, []string) {

	type PropertyCountTuple struct {
		Property string
		Count    uint64
	}
	var err error
	userPropertiesCount := make(map[string]uint64)
	userCatProperties := make([]string, 0)
	userCatPropertiesSorted := make([]string, 0)
	userPropertiesCountTuple := make([]PropertyCountTuple, 0)
	if parentNode.Pattern.PatternVersion < 2 {
		if parentNode.Pattern.PerUserUserCategoricalProperties != nil {
			for _, template := range *parentNode.Pattern.PerUserUserCategoricalProperties.Template {
				userCatProperties = append(userCatProperties, formatProperty(template.Name))
			}
			for _, bin := range parentNode.Pattern.PerUserUserCategoricalProperties.Bins {
				for index, freqMap := range bin.FrequencyMaps {
					userPropertiesCount[userCatProperties[index]] += freqMap.Count
				}
			}
		}

	} else {
		userPropertiesCount = parentNode.Pattern.FreqProps.GetPropertiesWithCount("user")
	}
	for property, count := range userPropertiesCount {
		userPropertiesCountTuple = append(userPropertiesCountTuple, PropertyCountTuple{
			Property: property,
			Count:    count,
		})
	}
	sort.Slice(userPropertiesCountTuple, func(i, j int) bool {
		return userPropertiesCountTuple[i].Count > userPropertiesCountTuple[j].Count
	})
	for _, property := range userPropertiesCountTuple {
		userCatPropertiesSorted = append(userCatPropertiesSorted, property.Property)
	}

	userNumProperties := make([]string, 0)
	if parentNode.Pattern.PatternVersion < 2 {
		if parentNode.Pattern.PerUserUserNumericProperties != nil {
			for _, template := range *parentNode.Pattern.PerUserUserNumericProperties.Template {
				userNumProperties = append(userNumProperties, formatProperty(template.Name))
			}
		}
	} else {
		userNumProperties = []string{}
	}

	eventCatProperties := make([]string, 0)
	if parentNode.Pattern.PatternVersion < 2 {
		if parentNode.Pattern.PerUserEventCategoricalProperties != nil {
			for _, template := range *parentNode.Pattern.PerUserEventCategoricalProperties.Template {
				eventCatProperties = append(eventCatProperties, formatProperty(template.Name))
			}
		}
	} else {
		eventCatProperties, err = parentNode.Pattern.FreqProps.GetPropertiesOfType("event")
		if err != nil {
			eventCatProperties = []string{}
		}
	}

	eventNumProperties := make([]string, 0)
	if parentNode.Pattern.PatternVersion < 2 {
		if parentNode.Pattern.PerUserEventNumericProperties != nil {
			for _, template := range *parentNode.Pattern.PerUserEventNumericProperties.Template {
				eventNumProperties = append(eventNumProperties, formatProperty(template.Name))
			}
		}
	} else {
		eventNumProperties = []string{}
	}

	return changePropertiesOrder(userCatPropertiesSorted, includedUserProperties), changePropertiesOrder(userNumProperties, includedUserProperties), changePropertiesOrder(eventCatProperties, includedEventProperties), changePropertiesOrder(eventNumProperties, includedEventProperties)
}

func changePropertiesOrder(properties []string, included map[string]bool) []string {
	if included == nil {
		return properties
	}
	first := make([]string, 0)
	second := make([]string, 0)
	for _, property := range properties {
		if included[property] == true {
			first = append(first, property)
		} else {
			second = append(second, property)
		}
	}
	return append(first, second...)
}

func formatProperty(property string) string {
	return strings.SplitN(property, ".", 2)[1]
}

func (it *Itree) buildAndAddPropertyChildNodesV1(reqId string,
	parentNode *ItreeNode, allActiveUsersPattern *P.Pattern,
	patternWrapper PatternServiceWrapperInterface, countType string, level int, debugKey string, debugPropertyName string, debugPropertyValue string,
	startEventConstraints *P.EventConstraints, endEventConstraints *P.EventConstraints, includedEventProperties, includedUserProperties map[string]bool, projectId uint64) ([]*ItreeNode, error, interface{}) {
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
	baseProperty := extractProperty(parentConstraints[len(parentConstraints)-1])
	var fpr, fpp float64
	pLen := len(parentPattern.EventNames)
	userCatProperties, userNumProperties, eventCatProperties, eventNumProperties := extractProperties(parentNode, includedEventProperties, includedUserProperties)
	if pLen == 1 {
		if countType == P.COUNT_TYPE_PER_USER {
			// Parent is root node. Count is all users.
			fpp = float64(parentPattern.TotalUserCount)
		} else {
			// Parent is root node. Count is all users.
			fpp = float64(patternWrapper.GetTotalEventCount(reqId))
		}
	} else {
		var subConstraints []P.EventConstraints
		if parentConstraints == nil {
			subConstraints = nil
		} else {
			subConstraints = parentConstraints[:pLen-1]
		}
		c, ok := patternWrapper.GetCount(reqId, parentPattern.EventNames[:pLen-1], subConstraints, countType)
		if !ok {
			return nil, fmt.Errorf(fmt.Sprintf("Frequency missing for pattern sequence %s", parentPattern.String())), nil
		}
		fpp = float64(c)
	}
	count, err := parentPattern.GetCount(parentConstraints, countType)
	if err != nil {
		return nil, err, nil
	}
	fpr = float64(count)

	childNodes := []*ItreeNode{}
	if fpp <= 0.0 || fpr <= 0.0 {
		//log.WithFields(log.Fields{"Parent": parentPattern.String(), "parentConstraints": parentConstraints,
		//	"fpr": fpr, "fpp": fpp}).Debug("Parent not frequent enough")
		return childNodes, nil, nil
	}

	// Add children by splitting on constraints on categorical properties.
	pLen = len(parentPattern.EventNames)
	CATEGORICAL_PROPERTY_REPETATION_THRESHOLD := int64(3)
	NUMERICAL_PROPERTY_REPETATION_THRESHOLD := int64(1)
	debugKey = fmt.Sprintf("itree_attribute_totaleventCategorical_level%v", level)
	debugCounts[debugKey] = debugCounts[debugKey] + len(eventCatProperties)
	debugKey = fmt.Sprintf("itree_attribute_totaleventNumerical_level%v", level)
	debugCounts[debugKey] = debugCounts[debugKey] + len(eventNumProperties)
	debugKey = fmt.Sprintf("itree_attribute_totaluserCategorical_level%v", level)
	debugCounts[debugKey] = debugCounts[debugKey] + len(userCatProperties)
	debugKey = fmt.Sprintf("itree_attribute_totaluserNumerical_level%v", level)
	debugCounts[debugKey] = debugCounts[debugKey] + len(userNumProperties)

	debugData := make(map[string][][]P.EventConstraints)
	debugData["UPCat"] = make([][]P.EventConstraints, 0)
	debugData["EPCat"] = make([][]P.EventConstraints, 0)
	debugData["UPNum"] = make([][]P.EventConstraints, 0)
	debugData["EPNum"] = make([][]P.EventConstraints, 0)
	if pLen > 1 || countType == P.COUNT_TYPE_PER_USER {
		child := it.buildCategoricalPropertyChildNodesV1(reqId,
			userCatProperties, NODE_TYPE_USER_PROPERTY, MAX_CAT_PROPERTIES_EVALUATED,
			MAX_CAT_VALUES_EVALUATED, parentNode, patternWrapper, allActiveUsersPattern, pLen, fpr, fpp, countType, startEventConstraints, endEventConstraints)
		debugKey = fmt.Sprintf("itree_attribute_totaluserCategorical_filtered_level%v", level)
		debugCounts[debugKey] = debugCounts[debugKey] + len(child)
		for _, node := range child {
			debugData["UPCat"] = append(debugData["UPCat"], node.PatternConstraints)
		}
		childNodes = append(childNodes, child...)
	}
	// Add children by splitting on constraints on categorical properties.
	pLen = len(parentPattern.EventNames)

	if pLen > 1 || countType == P.COUNT_TYPE_PER_USER {
		child := it.buildNumericalPropertyChildNodesV1(reqId,
			userNumProperties, NODE_TYPE_USER_PROPERTY, MAX_NUM_PROPERTIES_EVALUATED,
			parentNode, patternWrapper, allActiveUsersPattern, pLen, fpr, fpp, countType, startEventConstraints, endEventConstraints)
		debugKey = fmt.Sprintf("itree_attribute_totaluserNumerical_filtered_level%v", level)
		debugCounts[debugKey] = debugCounts[debugKey] + len(child)
		for _, node := range child {
			debugData["UPNum"] = append(debugData["UPNum"], node.PatternConstraints)
		}
		childNodes = append(childNodes, child...)
	}

	if pLen > 1 {
		child := it.buildCategoricalPropertyChildNodesV1(reqId,
			eventCatProperties, NODE_TYPE_EVENT_PROPERTY, MAX_CAT_PROPERTIES_EVALUATED,
			MAX_CAT_VALUES_EVALUATED, parentNode, patternWrapper, allActiveUsersPattern, pLen, fpr, fpp, countType, startEventConstraints, endEventConstraints)
		debugKey = fmt.Sprintf("itree_attribute_totaleventCategorical_filtered_level%v", level)
		debugCounts[debugKey] = debugCounts[debugKey] + len(child)
		for _, node := range child {
			debugData["EPCat"] = append(debugData["EPCat"], node.PatternConstraints)
		}
		childNodes = append(childNodes, child...)
	}

	if pLen > 1 {
		child := it.buildNumericalPropertyChildNodesV1(reqId,
			eventNumProperties, NODE_TYPE_EVENT_PROPERTY, MAX_NUM_PROPERTIES_EVALUATED,
			parentNode, patternWrapper, allActiveUsersPattern, pLen, fpr, fpp, countType, startEventConstraints, endEventConstraints)
		debugKey = fmt.Sprintf("itree_attribute_totaleventNumerical_filtered_level%v", level)
		debugCounts[debugKey] = debugCounts[debugKey] + len(child)
		for _, node := range child {
			debugData["EPNum"] = append(debugData["EPNum"], node.PatternConstraints)
		}
		childNodes = append(childNodes, child...)
	}
	// Sort in decreasing order of infoDrop.
	sort.SliceStable(childNodes,
		func(i, j int) bool {
			return (childNodes[i].InformationDrop > childNodes[j].InformationDrop)
		})

	addedChildNodes := []*ItreeNode{}
	numAddedChildNodes := 0
	// Initialize seen properties with that of parent properties.
	seenPropertyConstraints := getPropertyNamesMapFromConstraints(
		parentNode.PatternConstraints)
	seenPropertyConstraintsByCount := make(map[string]int64)
	for _, cNode := range childNodes {
		if cNode.InformationDrop <= 0.0 {
			debugKey := fmt.Sprintf("filterreason:itree_attribute_informationdrop_level%v", level)
			debugCounts[debugKey] = debugCounts[debugKey] + 1
			continue
		}
		// Dedup repeated constraints on different values.
		childPropertyConstraintsMap := getPropertyNamesMapFromConstraints(
			[]P.EventConstraints{cNode.AddedConstraint})

		isCat := getPropertyTypeOfAddedConstraint(cNode.AddedConstraint)
		isDup := false
		for childPropertyName, _ := range *childPropertyConstraintsMap {
			if _, found := (*seenPropertyConstraints)[childPropertyName]; found || isCat && seenPropertyConstraintsByCount[childPropertyName] >= CATEGORICAL_PROPERTY_REPETATION_THRESHOLD || !isCat && seenPropertyConstraintsByCount[childPropertyName] >= NUMERICAL_PROPERTY_REPETATION_THRESHOLD {
				isDup = true
				break
			}
		}
		if isDup {
			continue
		}
		it.addNode(cNode)
		addedChildNodes = append(addedChildNodes, cNode)
		for pc, _ := range *childPropertyConstraintsMap {
			seenPropertyConstraintsByCount[pc]++
		}
		numAddedChildNodes++
		// Only top MAX_PROPERTY_CHILD_NODES in order of drop in GiniImpurity are selected.
		if numAddedChildNodes >= MAX_PROPERTY_CHILD_NODES {
			break
		}
	}
	if debugKey == "Level0Attribute" {
		if level == 0 {
			return addedChildNodes, nil, debugData
		}
	}
	if debugKey == "Level1Attribute" {
		if level == 1 {
			for property, value := range baseProperty {
				if property == debugPropertyName && (value == debugPropertyValue || value == "") {
					return addedChildNodes, nil, debugData
				}
			}
		}
	}
	return addedChildNodes, nil, nil
}

func isPropertyGoalEventBlacklisted(goalEvent string, propertyName string) bool {

	BLACKLISTED_PROPERTIES := map[string][]string{
		"*": {"$hubspot_", "$salesforce_"},
	}

	for _, propertyPrefix := range BLACKLISTED_PROPERTIES["*"] {
		if strings.HasPrefix(propertyName, propertyPrefix) {
			return true
		}
	}
	return false
}
func (it *Itree) buildCategoricalPropertyChildNodesV1(reqId string,
	categoricalPropertyKeys []string,
	nodeType int, maxNumProperties int, maxNumValues int,
	parentNode *ItreeNode, patternWrapper PatternServiceWrapperInterface,
	allActiveUsersPattern *P.Pattern, pLen int, fpr float64, fpp float64,
	countType string, startEventConstraints *P.EventConstraints, endEventConstraints *P.EventConstraints) []*ItreeNode {
	propertyChildNodes := []*ItreeNode{}
	numP := 0
	seenProperties := getPropertyNamesMapFromConstraints(parentNode.PatternConstraints)
	for _, propertyName := range categoricalPropertyKeys {
		if isPropertyGoalEventBlacklisted(parentNode.Pattern.EventNames[len(parentNode.Pattern.EventNames)-1], propertyName) {
			continue
		}
		if numP > maxNumProperties {
			break
		}
		if _, found := (*seenProperties)[propertyName]; found {
			continue
		}
		if nodeType == NODE_TYPE_EVENT_PROPERTY && U.ShouldIgnoreItreeProperty(propertyName) {
			continue
		}
		if nodeType == NODE_TYPE_USER_PROPERTY && U.ShouldIgnoreItreeProperty(propertyName) {
			continue
		}
		numP++
		numVal := 0
		// Compute KL Distance of child vs parent.
		klDistanceUnits := []KLDistanceUnitInfo{}
		attributeIndex := int(0)
		if pLen == 1 {
			attributeIndex = 0
		} else {
			attributeIndex = pLen - 2
		}
		seenValues := parentNode.Pattern.GetPerUserUserPropertyValues(attributeIndex, propertyName)
		for _, value := range seenValues {
			if numVal > maxNumValues {
				break
			}
			if value == "" {
				continue
			}
			numVal++
			constraintToAdd := P.EventConstraints{
				EPNumericConstraints:     []P.NumericConstraint{},
				EPCategoricalConstraints: []P.CategoricalConstraint{},
				UPNumericConstraints:     []P.NumericConstraint{},
				UPCategoricalConstraints: []P.CategoricalConstraint{},
			}
			categoricalConstraint := []P.CategoricalConstraint{
				{
					PropertyName:  propertyName,
					PropertyValue: value,
					Operator:      P.EQUALS_OPERATOR_CONST,
				},
			}
			if nodeType == NODE_TYPE_EVENT_PROPERTY {
				constraintToAdd.EPCategoricalConstraints = categoricalConstraint
			} else if nodeType == NODE_TYPE_USER_PROPERTY {
				constraintToAdd.UPCategoricalConstraints = categoricalConstraint
			}

			if cNode, err := it.buildChildNodeV1(reqId,
				parentNode.Pattern, &constraintToAdd, nodeType,
				parentNode.Index, patternWrapper, allActiveUsersPattern, fpr, fpp, countType, startEventConstraints, endEventConstraints); err != nil {
				log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build child node")
				continue
			} else {
				if cNode == nil {
					continue
				}
				propertyChildNodes = append(propertyChildNodes, cNode)
				// Distance computation for graph node.
				// Distance is the measure of change in distribution between the pattern and the rule.
				patternProb := cNode.Fcp / fpp
				ruleProb := cNode.Fcr / fpr
				klDistanceUnits = append(klDistanceUnits, KLDistanceUnitInfo{
					PropertyValue: value,
					// The distance / divergence of rule Distribution from pattern Distribution wrt propertyName in bits.
					Distance: computeKLDistanceBits(patternProb, ruleProb),
					Fpp:      cNode.Fpp,
					Fpr:      cNode.Fpr,
					Fcp:      cNode.Fcp,
					Fcr:      cNode.Fcr,
				})
			}
		}

		if len(klDistanceUnits) > 0 {
			// Append None label if there is discrepancy in count.
			totalFcp := 0.0
			totalFcr := 0.0
			for _, klu := range klDistanceUnits {
				totalFcp += klu.Fcp
				totalFcr += klu.Fcr
			}
			noneFcp := fpp - totalFcp
			noneFcr := fpr - totalFcr
			if noneFcp > 0.0 || noneFcr > 0.0 {
				nonePatternProb := noneFcp / fpp
				noneRuleProb := noneFcr / fpr

				noneKLDistanceUnit := KLDistanceUnitInfo{
					PropertyValue: NONE_PROPERTY_VALUES_LABEL,
					Fpp:           fpp,
					Fpr:           fpr,
					Fcp:           noneFcp,
					Fcr:           noneFcr,
					Distance:      computeKLDistanceBits(nonePatternProb, noneRuleProb),
				}
				klDistanceUnits = append(klDistanceUnits, noneKLDistanceUnit)
			}

			/* Not showing graph results for now as it was confusing for users to interpret.
			nodeGraphType := NODE_TYPE_GRAPH_USER_PROPERTIES
			if nodeType == NODE_TYPE_EVENT_PROPERTY {
				nodeGraphType = NODE_TYPE_GRAPH_EVENT_PROPERTIES
			}
			if cGraphNode, err := it.buildCategoricalGraphChildNode(
				parentNode.Pattern, propertyName, klDistanceUnits, nodeGraphType,
				parentNode.Index, patternWrapper, fpr, fpp, countType); err != nil {
				log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build graph child node")
			} else {
				propertyChildNodes = append(propertyChildNodes, cGraphNode)
			}*/
		}
	}
	return propertyChildNodes
}

func (it *Itree) buildAndAddSequenceChildNodesV1(reqId string,
	parentNode *ItreeNode, candidatePattens []*P.Pattern,
	patternWrapper PatternServiceWrapperInterface,
	allActiveUsersPattern *P.Pattern, countType string,
	startEventConstraints *P.EventConstraints, endEventConstraints *P.EventConstraints,
	crmEvents map[string]bool, includedEvents map[string]bool) ([]*ItreeNode, error) {

	parentPattern := parentNode.Pattern
	peLen := len(parentPattern.EventNames)
	count, err := parentPattern.GetCount(parentNode.PatternConstraints, countType)
	if err != nil {
		return nil, err
	}
	fpr := float64(count)

	var fpp float64
	if peLen == 1 {
		if countType == P.COUNT_TYPE_PER_USER {
			// Parent is root node. Count is all users.
			fpp = float64(parentPattern.TotalUserCount)
		} else {
			// Parent is root node. Count is all users.
			fpp = float64(patternWrapper.GetTotalEventCount(reqId))
		}
	} else {
		var subParentConstraints []P.EventConstraints
		if parentNode.PatternConstraints == nil {
			subParentConstraints = nil
		} else {
			subParentConstraints = parentNode.PatternConstraints[:peLen-1]
		}
		c, ok := patternWrapper.GetCount(
			reqId, parentPattern.EventNames[:peLen-1], subParentConstraints, countType)
		if !ok {
			return nil, fmt.Errorf(fmt.Sprintf(
				"Frequency missing for pattern sequence %s", parentPattern.String()))
		}
		fpp = float64(c)
	}

	childNodes := []*ItreeNode{}
	if fpp <= 0.0 || fpr <= 0.0 {
		//log.WithFields(log.Fields{"Parent": parentPattern.String(), "fpr": fpr, "fpp": fpp}).Debug(
		//	"Parent not frequent enough")
		return childNodes, nil
	}
	for _, p := range candidatePattens {
		if !isChildSequence(parentNode.Pattern.EventNames, p.EventNames, crmEvents) || P.IsEncodedEvent(p.EventNames[len(p.EventNames)-2]) {
			continue
		}
		if cNode, err := it.buildChildNodeV1(reqId,
			p, nil, NODE_TYPE_SEQUENCE,
			parentNode.Index, patternWrapper, allActiveUsersPattern, fpr, fpp, countType, startEventConstraints, endEventConstraints); err != nil {
			log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build child node")
			continue
		} else {
			if cNode == nil {
				continue
			}
			childNodes = append(childNodes, cNode)
		}
	}
	// Sort in decreasing order of infoDrop.
	sort.SliceStable(childNodes,
		func(i, j int) bool {
			return (childNodes[i].InformationDrop > childNodes[j].InformationDrop)
		})

	// Only top MAX_SEQUENCE_CHILD_NODES in order of drop in GiniImpurity are selected.
	childNodesSelected := childNodes
	if len(childNodes) > MAX_SEQUENCE_CHILD_NODES {
		childNodesSelected = childNodes[:MAX_SEQUENCE_CHILD_NODES]
		if includedEvents != nil && len(includedEvents) > 0 {
			for _, cNode := range childNodes[MAX_SEQUENCE_CHILD_NODES:] {
				for _, event := range cNode.Pattern.EventNames {
					if includedEvents[event] == true {
						childNodesSelected = append(childNodesSelected, cNode)
						break
					}
				}
			}
		}
	}

	addedChildNodes := []*ItreeNode{}
	for _, cNode := range childNodesSelected {
		if cNode.InformationDrop <= 0.0 {
			continue
		}
		it.addNode(cNode)
		addedChildNodes = append(addedChildNodes, cNode)
	}

	return addedChildNodes, nil
}

func (it *Itree) buildAndAddCampaignChildNodesV1(reqId string,
	parentNode *ItreeNode, candidatePattens []*P.Pattern,
	patternWrapper PatternServiceWrapperInterface,
	allActiveUsersPattern *P.Pattern, countType string, startEventConstraints *P.EventConstraints, endEventConstraints *P.EventConstraints, crmEvents map[string]bool) ([]*ItreeNode, error) {

	parentPattern := parentNode.Pattern
	peLen := len(parentPattern.EventNames)
	count, err := parentPattern.GetCount(parentNode.PatternConstraints, countType)
	if err != nil {
		return nil, err
	}
	fpr := float64(count)

	var fpp float64
	if peLen == 1 {
		if countType == P.COUNT_TYPE_PER_USER {
			// Parent is root node. Count is all users.
			fpp = float64(parentPattern.TotalUserCount)
		} else {
			// Parent is root node. Count is all users.
			fpp = float64(patternWrapper.GetTotalEventCount(reqId))
		}
	} else {
		var subParentConstraints []P.EventConstraints
		if parentNode.PatternConstraints == nil {
			subParentConstraints = nil
		} else {
			subParentConstraints = parentNode.PatternConstraints[:peLen-1]
		}
		c, ok := patternWrapper.GetCount(
			reqId, parentPattern.EventNames[:peLen-1], subParentConstraints, countType)
		if !ok {
			return nil, fmt.Errorf(fmt.Sprintf(
				"Frequency missing for pattern sequence %s", parentPattern.String()))
		}
		fpp = float64(c)
	}

	childNodes := []*ItreeNode{}
	if fpp <= 0.0 || fpr <= 0.0 {
		//log.WithFields(log.Fields{"Parent": parentPattern.String(), "fpr": fpr, "fpp": fpp}).Debug(
		//	"Parent not frequent enough")
		return childNodes, nil
	}
	subPattersnCount := 0
	filteredPatternsCount := 0
	for _, p := range candidatePattens {
		if !isChildSequence(parentNode.Pattern.EventNames, p.EventNames, crmEvents) || !P.IsEncodedEvent(p.EventNames[len(p.EventNames)-2]) {
			continue
		}
		if P.ExtractCampaignName(p.EventNames[len(p.EventNames)-2]) == "" {
			continue
		}
		subPattersnCount++
		if cNode, err := it.buildChildNodeV1(reqId,
			p, nil, NODE_TYPE_CAMPAIGN,
			parentNode.Index, patternWrapper, allActiveUsersPattern, fpr, fpp, countType, startEventConstraints, endEventConstraints); err != nil {
			log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build child node")
			continue
		} else {
			if cNode == nil || cNode.Fcp == cNode.Fpp {
				continue
			}
			filteredPatternsCount++
			childNodes = append(childNodes, cNode)
		}
	}
	log.WithFields(log.Fields{
		"matching_campaign_child_nodes": subPattersnCount}).Info("explain_debug")
	log.WithFields(log.Fields{
		"filtered_campaign_child_nodes": filteredPatternsCount}).Info("explain_debug")
	// Sort in decreasing order of infoDrop.
	sort.SliceStable(childNodes,
		func(i, j int) bool {
			return (childNodes[i].InformationDrop > childNodes[j].InformationDrop)
		})

	// Only top MAX_SEQUENCE_CHILD_NODES in order of drop in GiniImpurity are selected.
	if len(childNodes) > MAX_SEQUENCE_CHILD_NODES {
		childNodes = childNodes[:MAX_SEQUENCE_CHILD_NODES]
	}

	addedChildNodes := []*ItreeNode{}
	for _, cNode := range childNodes {
		it.addNode(cNode)
		addedChildNodes = append(addedChildNodes, cNode)
	}

	return addedChildNodes, nil
}

func (it *Itree) buildChildNodeV1(reqId string,
	childPattern *P.Pattern, constraintToAdd *P.EventConstraints,
	nodeType int, parentIndex int,
	patternWrapper PatternServiceWrapperInterface,
	allActiveUsersPattern *P.Pattern,
	fpr float64, fpp float64, countType string, startEventConstraints *P.EventConstraints, endEventConstraints *P.EventConstraints) (*ItreeNode, error) {

	parentNode := it.Nodes[parentIndex]
	parentPattern := parentNode.Pattern
	parentConstraints := parentNode.PatternConstraints
	childLen := len(childPattern.EventNames)
	parentLen := len(parentPattern.EventNames)

	if (nodeType == NODE_TYPE_CAMPAIGN || nodeType == NODE_TYPE_SEQUENCE) && (parentLen != childLen-1 || parentLen < 1) {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Current node with type sequence and pattern(%s) with constraintToAdd(%v) and parent index(%d), pattern(%s) with constraints(%v) not compatible.",
			childPattern.String(), constraintToAdd, parentIndex, parentPattern.String(), parentConstraints))
	}

	if (nodeType == NODE_TYPE_EVENT_PROPERTY || nodeType == NODE_TYPE_USER_PROPERTY) && childLen != parentLen {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Current node with type property and pattern(%s) with constraintToAdd(%v) and parent index(%d), pattern(%s) with constraints(%v) not compatible",
			childPattern.String(), constraintToAdd, parentIndex, parentPattern.String(), parentConstraints))
	}

	if parentLen == 1 && countType == P.COUNT_TYPE_PER_OCCURRENCE && nodeType != NODE_TYPE_SEQUENCE {
		// Cannot add constraints to AllActiveUsers Pattern when counting by occurrence.
		return nil, nil
	}

	// Build Child constraints.
	childPatternConstraints := make([]P.EventConstraints, childLen)
	if nodeType == NODE_TYPE_SEQUENCE || nodeType == NODE_TYPE_CAMPAIGN {
		if parentConstraints != nil {
			for i := 0; i < childLen-2; i++ {
				childPatternConstraints[i] = parentNode.PatternConstraints[i]
			}
			// childPattern has an extra event at childLen-2.
			// Minimum expected childLen is 2. Else this will crash.
			if childLen == 2 {
				// If parent is single event, then the constraints are attached to the
				// end event by default.
				childPatternConstraints[1] = parentNode.PatternConstraints[0]
			} else {
				childPatternConstraints[childLen-2] = P.EventConstraints{}
			}
			childPatternConstraints[childLen-1] = parentNode.PatternConstraints[childLen-2]
		} else {
			childPatternConstraints = nil
		}
	} else if nodeType == NODE_TYPE_EVENT_PROPERTY {
		for i := 0; i < childLen; i++ {
			if parentConstraints != nil {
				childPatternConstraints[i] = parentConstraints[i]
			} else {
				childPatternConstraints[i] = P.EventConstraints{
					EPNumericConstraints:     []P.NumericConstraint{},
					EPCategoricalConstraints: []P.CategoricalConstraint{},
					UPNumericConstraints:     []P.NumericConstraint{},
					UPCategoricalConstraints: []P.CategoricalConstraint{},
				}
			}
		}
		// Additional Event constraints are added to N-1st event.
		childPatternConstraints[childLen-2].EPCategoricalConstraints = append(
			childPatternConstraints[childLen-2].EPCategoricalConstraints,
			(*constraintToAdd).EPCategoricalConstraints...)
		childPatternConstraints[childLen-2].EPNumericConstraints = append(
			childPatternConstraints[childLen-2].EPNumericConstraints,
			(*constraintToAdd).EPNumericConstraints...)
	} else if nodeType == NODE_TYPE_USER_PROPERTY {
		for i := 0; i < childLen; i++ {
			if parentConstraints != nil {
				childPatternConstraints[i] = parentConstraints[i]
			} else {
				childPatternConstraints[i] = P.EventConstraints{
					EPNumericConstraints:     []P.NumericConstraint{},
					EPCategoricalConstraints: []P.CategoricalConstraint{},
					UPNumericConstraints:     []P.NumericConstraint{},
					UPCategoricalConstraints: []P.CategoricalConstraint{},
				}
			}
		}
		if childLen > 1 {
			// By default additional constraints are added to N-1st event.
			childPatternConstraints[childLen-2].UPCategoricalConstraints = append(
				childPatternConstraints[childLen-2].UPCategoricalConstraints,
				(*constraintToAdd).UPCategoricalConstraints...)
			childPatternConstraints[childLen-2].UPNumericConstraints = append(
				childPatternConstraints[childLen-2].UPNumericConstraints,
				(*constraintToAdd).UPNumericConstraints...)
		} else {
			// If only one node, additional user property constraints are added
			// to the last node.
			childPatternConstraints[childLen-1].UPCategoricalConstraints = append(
				childPatternConstraints[childLen-1].UPCategoricalConstraints,
				(*constraintToAdd).UPCategoricalConstraints...)
			childPatternConstraints[childLen-1].UPNumericConstraints = append(
				childPatternConstraints[childLen-1].UPNumericConstraints,
				(*constraintToAdd).UPNumericConstraints...)
		}
	}

	// Compute Child frequencies.
	var fcr, fcp float64
	if nodeType == NODE_TYPE_SEQUENCE || nodeType == NODE_TYPE_CAMPAIGN {
		pCount, _ := childPattern.GetCount(childPatternConstraints, countType)
		fcr = float64(pCount)
		var subConstraints []P.EventConstraints
		if childPatternConstraints == nil {
			subConstraints = nil
		} else {
			subConstraints = childPatternConstraints[:childLen-1]
		}
		c, _ := patternWrapper.GetCount(reqId, childPattern.EventNames[:childLen-1], subConstraints, countType)
		fcp = float64(c)
	} else {
		pCount, _ := childPattern.GetCount(childPatternConstraints, countType)
		fcr = float64(pCount)
		if childLen > 1 {
			var subConstraints []P.EventConstraints
			if childPatternConstraints == nil {
				subConstraints = nil
			} else {
				subConstraints = childPatternConstraints[:childLen-1]
			}
			c, _ := patternWrapper.GetCount(reqId, childPattern.EventNames[:childLen-1], subConstraints, countType)
			fcp = float64(c)
		} else {
			if nodeType != NODE_TYPE_USER_PROPERTY || countType != P.COUNT_TYPE_PER_USER {
				return nil, fmt.Errorf(fmt.Sprintf(
					"Unexpected length of 1 for pattern %s, nodeType %d, constraintToAdd %v, countType: %s",
					childPattern.String(), nodeType, constraintToAdd, countType))
			}
			var allUConstraints []P.EventConstraints
			if len(parentConstraints[0].EPCategoricalConstraints) > 0 ||
				len(parentConstraints[0].EPNumericConstraints) > 0 ||
				len(parentConstraints[0].UPCategoricalConstraints) > 0 ||
				len(parentConstraints[0].UPNumericConstraints) > 0 {
				// frequency of child pattern is computed over allUserEvents after applying constraint.
				allUConstraints = []P.EventConstraints{
					*constraintToAdd,
				}
				nonEndConstraints := P.EventConstraints{
					EPNumericConstraints:     []P.NumericConstraint{},
					EPCategoricalConstraints: []P.CategoricalConstraint{},
					UPNumericConstraints:     []P.NumericConstraint{},
					UPCategoricalConstraints: []P.CategoricalConstraint{},
				}
				for _, constraint := range parentConstraints[0].EPCategoricalConstraints {
					toBeAdded := true
					for _, endConstraint := range endEventConstraints.EPCategoricalConstraints {
						if reflect.DeepEqual(constraint, endConstraint) {
							toBeAdded = false
							break
						}
					}
					if toBeAdded {
						nonEndConstraints.EPCategoricalConstraints = append(nonEndConstraints.EPCategoricalConstraints, constraint)
					}
				}
				for _, constraint := range parentConstraints[0].EPNumericConstraints {
					toBeAdded := true
					for _, endConstraint := range endEventConstraints.EPNumericConstraints {
						if reflect.DeepEqual(constraint, endConstraint) {
							toBeAdded = false
							break
						}
					}
					if toBeAdded {
						nonEndConstraints.EPNumericConstraints = append(nonEndConstraints.EPNumericConstraints, constraint)
					}
				}
				for _, constraint := range parentConstraints[0].UPCategoricalConstraints {
					toBeAdded := true
					for _, endConstraint := range endEventConstraints.UPCategoricalConstraints {
						if reflect.DeepEqual(constraint, endConstraint) {
							toBeAdded = false
							break
						}
					}
					if toBeAdded {
						nonEndConstraints.UPCategoricalConstraints = append(nonEndConstraints.UPCategoricalConstraints, constraint)
					}
				}
				for _, constraint := range parentConstraints[0].UPNumericConstraints {
					toBeAdded := true
					for _, endConstraint := range endEventConstraints.UPNumericConstraints {
						if reflect.DeepEqual(constraint, endConstraint) {
							toBeAdded = false
							break
						}
					}
					if toBeAdded {
						nonEndConstraints.UPNumericConstraints = append(nonEndConstraints.UPNumericConstraints, constraint)
					}
				}
				// For patternLength = 1 we dont need to consider endevent constraints for fcp calculation we just need startEvent constraints
				if len(startEventConstraints.EPCategoricalConstraints) > 0 ||
					len(startEventConstraints.EPNumericConstraints) > 0 ||
					len(startEventConstraints.UPCategoricalConstraints) > 0 ||
					len(startEventConstraints.UPNumericConstraints) > 0 {
					allUConstraints[0].EPCategoricalConstraints = append(allUConstraints[0].EPCategoricalConstraints, startEventConstraints.EPCategoricalConstraints...)
					allUConstraints[0].EPNumericConstraints = append(allUConstraints[0].EPNumericConstraints, startEventConstraints.EPNumericConstraints...)
					allUConstraints[0].UPCategoricalConstraints = append(allUConstraints[0].UPCategoricalConstraints, startEventConstraints.UPCategoricalConstraints...)
					allUConstraints[0].UPNumericConstraints = append(allUConstraints[0].UPNumericConstraints, startEventConstraints.UPNumericConstraints...)
				}

				if len(nonEndConstraints.EPCategoricalConstraints) > 0 ||
					len(nonEndConstraints.EPNumericConstraints) > 0 ||
					len(nonEndConstraints.UPCategoricalConstraints) > 0 ||
					len(nonEndConstraints.UPNumericConstraints) > 0 {
					allUConstraints[0].EPCategoricalConstraints = append(allUConstraints[0].EPCategoricalConstraints, nonEndConstraints.EPCategoricalConstraints...)
					allUConstraints[0].EPNumericConstraints = append(allUConstraints[0].EPNumericConstraints, nonEndConstraints.EPNumericConstraints...)
					allUConstraints[0].UPCategoricalConstraints = append(allUConstraints[0].UPCategoricalConstraints, nonEndConstraints.UPCategoricalConstraints...)
					allUConstraints[0].UPNumericConstraints = append(allUConstraints[0].UPNumericConstraints, nonEndConstraints.UPNumericConstraints...)
				}
				// Take all the event constraints which are part of start and not part of end and remove those from calculating fcp
			} else {
				allUConstraints = []P.EventConstraints{
					*constraintToAdd,
				}
			}
			c, _ := allActiveUsersPattern.GetPerUserCount(allUConstraints)
			fcp = float64(c)
		}
	}

	if fcp <= 0.0 || fcr <= 0.0 {
		//log.WithFields(log.Fields{"Child": childPattern.String(),
		//	"constraintToAdd": constraintToAdd, "parentConstraints": parentConstraints,
		//	"fcp": fcp, "fcr": fcr}).Debug("Child not frequent enough")
		return nil, nil
	}
	// Expected.
	// fpp >= fpr
	// fcp >= fcr
	// fpp >= fcp
	// fpr >= fcr
	if fpp < fpr || fcp < fcr || fpp < fcp || fpr < fcr {
		log.WithFields(log.Fields{"Child": childPattern.String(), "constraints": childPatternConstraints,
			"fpp": fpp, "fpr": fpr, "fcp": fcp, "fcr": fcr, "nodeType": nodeType}).Debug("Inconsistent frequencies. Ignoring")
		return nil, nil
	}
	p := fcr / fcp
	rightInfo := information([]float64{p, 1.0 - p})
	rightFraction := fcp / fpp
	// Left information.
	var leftInfo float64
	if (fpp - fcp) > 0.0 {
		leftP := (fpr - fcr) / (fpp - fcp)
		if leftP > 1.0 {
			leftP = 1.0
		}
		leftInfo = information([]float64{leftP, 1.0 - leftP})
	} else {
		leftInfo = 0.0
	}
	overallInfo := rightFraction*rightInfo + (1-rightFraction)*leftInfo

	infoDrop := parentNode.RightInformation - overallInfo

	confidence := fcr / fcp
	confidenceGain := confidence - parentNode.Confidence

	addedConstraint := P.EventConstraints{}
	if constraintToAdd != nil {
		addedConstraint = *constraintToAdd
	}
	node := ItreeNode{
		Pattern:            childPattern,
		NodeType:           nodeType,
		PatternConstraints: childPatternConstraints,
		ParentIndex:        parentIndex,
		RightInformation:   rightInfo,
		RightFraction:      rightFraction,
		OverallInformation: overallInfo,
		InformationDrop:    infoDrop,
		Confidence:         confidence,
		ConfidenceGain:     confidenceGain,
		Fpp:                fpp,
		Fpr:                fpr,
		Fcp:                fcp,
		Fcr:                fcr,
		AddedConstraint:    addedConstraint,
	}
	//log.WithFields(log.Fields{"node": node.Pattern.String(), "constraints": constraints,
	//	"parent": parentPattern.String(), "parentConstraints": parentNode.PatternConstraints,
	//	"fcr": fcr, "fcp": fcp, "fpr": fpr, "fpp": fpp, "OverallInfo": overallInfo, "infoDrop": infoDrop}).Debug(
	//	"Built candidate child node.")
	return &node, nil
}

func (it *Itree) buildNumericalPropertyChildNodesV1(reqId string,
	numericPropertyKeys []string, nodeType int, maxNumProperties int,
	parentNode *ItreeNode, patternWrapper PatternServiceWrapperInterface,
	allActiveUsersPattern *P.Pattern, pLen int, fpr float64, fpp float64,
	countType string, startEventConstraints *P.EventConstraints, endEventConstraints *P.EventConstraints) []*ItreeNode {

	parentPattern := parentNode.Pattern
	propertyChildNodes := []*ItreeNode{}
	numP := 0
	seenProperties := getPropertyNamesMapFromConstraints(parentNode.PatternConstraints)
	for _, propertyName := range numericPropertyKeys {
		if numP > maxNumProperties {
			break
		}
		if _, found := (*seenProperties)[propertyName]; found {
			continue
		}
		if nodeType == NODE_TYPE_EVENT_PROPERTY && U.ShouldIgnoreItreeProperty(propertyName) {
			continue
		}
		if nodeType == NODE_TYPE_USER_PROPERTY && U.ShouldIgnoreItreeProperty(propertyName) {
			continue
		}
		if U.ShouldIgnoreItreeNumericalProperty(propertyName) {
			continue
		}
		numP++
		binRanges := [][2]float64{}
		isPredefined := false
		if nodeType == NODE_TYPE_EVENT_PROPERTY {
			if countType == P.COUNT_TYPE_PER_USER {
				binRanges, isPredefined = parentPattern.GetPerUserEventPropertyRanges(pLen-2, propertyName)
			} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
				binRanges, isPredefined = parentPattern.GetPerOccurrenceEventPropertyRanges(pLen-2, propertyName)
			}
		} else if nodeType == NODE_TYPE_USER_PROPERTY {
			if countType == P.COUNT_TYPE_PER_USER {
				binRanges, isPredefined = parentPattern.GetPerUserUserPropertyRanges(pLen-2, propertyName)
			} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
				binRanges, isPredefined = parentPattern.GetPerOccurrenceUserPropertyRanges(pLen-2, propertyName)
			}
		}

		for _, pRange := range binRanges {
			minValue := pRange[0]
			maxValue := pRange[1]
			// Try three combinations of constraints. v < min, min < v < max, v > max
			constraint1ToAdd := P.EventConstraints{
				EPNumericConstraints:     []P.NumericConstraint{},
				EPCategoricalConstraints: []P.CategoricalConstraint{},
				UPNumericConstraints:     []P.NumericConstraint{},
				UPCategoricalConstraints: []P.CategoricalConstraint{},
			}
			if !isPredefined {
				// Show only interval constraint for predefined bin ranges.
				lesserThanConstraint := []P.NumericConstraint{
					{
						PropertyName: propertyName,
						LowerBound:   -math.MaxFloat64,
						UpperBound:   minValue,
						UseBound:     "OnlyUpper",
					},
				}
				if nodeType == NODE_TYPE_EVENT_PROPERTY {
					constraint1ToAdd.EPNumericConstraints = lesserThanConstraint
				} else if nodeType == NODE_TYPE_USER_PROPERTY {
					constraint1ToAdd.UPNumericConstraints = lesserThanConstraint
				}
			}

			constraint2ToAdd := P.EventConstraints{
				EPNumericConstraints:     []P.NumericConstraint{},
				EPCategoricalConstraints: []P.CategoricalConstraint{},
				UPNumericConstraints:     []P.NumericConstraint{},
				UPCategoricalConstraints: []P.CategoricalConstraint{},
			}
			intervalConstraint := []P.NumericConstraint{
				{
					PropertyName: propertyName,
					LowerBound:   minValue,
					UpperBound:   maxValue,
					UseBound:     "Both",
				},
			}
			if nodeType == NODE_TYPE_EVENT_PROPERTY {
				constraint2ToAdd.EPNumericConstraints = intervalConstraint
			} else if nodeType == NODE_TYPE_USER_PROPERTY {
				constraint2ToAdd.UPNumericConstraints = intervalConstraint
			}

			constraint3ToAdd := P.EventConstraints{
				EPNumericConstraints:     []P.NumericConstraint{},
				EPCategoricalConstraints: []P.CategoricalConstraint{},
				UPNumericConstraints:     []P.NumericConstraint{},
				UPCategoricalConstraints: []P.CategoricalConstraint{},
			}
			if !isPredefined {
				greaterThanConstraint := []P.NumericConstraint{
					{
						PropertyName: propertyName,
						LowerBound:   maxValue,
						UpperBound:   math.MaxFloat64,
						UseBound:     "OnlyLower",
					},
				}
				if nodeType == NODE_TYPE_EVENT_PROPERTY {
					constraint3ToAdd.EPNumericConstraints = greaterThanConstraint
				} else if nodeType == NODE_TYPE_USER_PROPERTY {
					constraint3ToAdd.UPNumericConstraints = greaterThanConstraint
				}
			}

			for _, constraintToAdd := range []P.EventConstraints{
				constraint2ToAdd} {
				if cNode, err := it.buildChildNodeV1(reqId,
					parentPattern, &constraintToAdd, nodeType,
					parentNode.Index, patternWrapper, allActiveUsersPattern,
					fpr, fpp, countType, startEventConstraints, endEventConstraints); err != nil {
					log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build child node")
					continue
				} else {
					if cNode == nil {
						continue
					}
					if len(extractProperty(cNode.AddedConstraint)) <= 0 {
						continue
					}
					propertyChildNodes = append(propertyChildNodes, cNode)
				}
			}
		}
	}
	return propertyChildNodes

}

func sortMap(data map[string]uint) interface{} {
	type tuple struct {
		key   string
		value uint
	}
	arr := make([]tuple, 0)
	for key, value := range data {
		arr = append(arr, tuple{
			key:   key,
			value: value,
		})
	}
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].value > arr[j].value
	})
	res := make([]string, 0)
	for _, element := range arr {
		res = append(res, fmt.Sprintf("%v ----> %v", element.key, element.value))
	}
	return res
}

func BuildUserDistribution(reqId string, event string,
	constraints *P.EventConstraints, patternWrapper PatternServiceWrapperInterface) (uint, map[string]uint, map[string]uint, error) {
	if event == "" {
		return 0, nil, nil, fmt.Errorf("Missing event")
	}

	patterns, err := patternWrapper.GetAllContainingPatterns(reqId, event)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("error fetching patterns")
	}
	before := make(map[string]uint)
	after := make(map[string]uint)
	base := uint(0)
	isConstraintPresent := true
	if len(constraints.EPNumericConstraints) == 0 && len(constraints.EPCategoricalConstraints) == 0 && len(constraints.UPNumericConstraints) == 0 && len(constraints.UPCategoricalConstraints) == 0 {
		isConstraintPresent = false
	}
	for _, pattern := range patterns {
		if len(pattern.EventNames) == 1 {
			if isConstraintPresent == true {
				patternDetails := patternWrapper.GetPattern(reqId, pattern.EventNames)
				count, err := patternDetails.GetCount(constructPatternConstraints(
					1, constraints, constraints), P.COUNT_TYPE_PER_USER)
				if err == nil {
					base = uint(count)
				}
			} else {
				base = pattern.PerUserCount
			}
		}
		if len(pattern.EventNames) == 2 {
			patternDetails := patternWrapper.GetPattern(reqId, pattern.EventNames)
			if pattern.EventNames[0] == event && pattern.EventNames[1] != event {
				if isConstraintPresent == true {
					count, err := patternDetails.GetCount(constructPatternConstraints(
						2, constraints, constraints), P.COUNT_TYPE_PER_USER)
					if err == nil {
						after[pattern.EventNames[1]] += uint(count)
					}
				} else {
					after[pattern.EventNames[1]] += pattern.PerUserCount
				}
			}
			if pattern.EventNames[1] == event && pattern.EventNames[0] != event {
				if isConstraintPresent == true {
					count, err := patternDetails.GetCount(constructPatternConstraints(
						2, constraints, constraints), P.COUNT_TYPE_PER_USER)
					if err == nil {
						before[pattern.EventNames[0]] += uint(count)
					}
				} else {
					before[pattern.EventNames[0]] += pattern.PerUserCount
				}
			}
		}
	}
	return base, before, after, nil
}

func BuildUserDistributionWithProperties(reqId string, events []string, startConstraints *P.EventConstraints, endConstraints *P.EventConstraints, patternWrapper PatternServiceWrapperInterface) (uint, map[string]uint, error) {
	if len(events) == 0 {
		events = []string{U.SEN_ALL_ACTIVE_USERS}
	}
	if len(events) > 2 {
		return 0, nil, errors.New("Invalid number of events")
	}
	overall, target := uint(0), make(map[string]uint)
	pattern := patternWrapper.GetPattern(reqId, events)
	if pattern == nil {
		return 0, nil, fmt.Errorf("error fetching patterns")
	}
	existingConstraints := make(map[string]bool)
	for _, prop := range endConstraints.UPCategoricalConstraints {
		existingConstraints[prop.PropertyName] = true
	}
	if count, _ := pattern.GetCount(constructPatternConstraints(
		1, startConstraints, endConstraints), P.COUNT_TYPE_PER_USER); count > 0 {
		overall = count
		allUserCatProperties, _, _, _ := extractProperties(&ItreeNode{Pattern: pattern}, nil, nil)
		distributionProperty := make(map[string]string)
		for _, property := range allUserCatProperties {
			if existingConstraints[property] != true {
				distributionProperty[property] = "categorical"
			}
		}
		for property, propertyCategory := range distributionProperty {
			if propertyCategory == "categorical" {
				values := pattern.GetPerUserUserPropertyValues(0, property)
				for _, value := range values {
					constraints := P.EventConstraints{
						EPNumericConstraints:     endConstraints.EPNumericConstraints,
						EPCategoricalConstraints: endConstraints.EPCategoricalConstraints,
						UPNumericConstraints:     endConstraints.UPNumericConstraints,
						UPCategoricalConstraints: endConstraints.UPCategoricalConstraints,
					}
					constraints.UPCategoricalConstraints = append(constraints.UPCategoricalConstraints,
						P.CategoricalConstraint{PropertyName: property, PropertyValue: value, Operator: P.EQUALS_OPERATOR_CONST})
					if count, _ := pattern.GetCount(constructPatternConstraints(
						1, startConstraints, &constraints), P.COUNT_TYPE_PER_USER); count > 0 {
						target[property+"::"+value] = count
					}
				}
			}
		}
	}
	return overall, target, nil
}
