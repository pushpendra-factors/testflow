// itree
// Methods to build Insight/Information trees.
package pattern

import (
	"fmt"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ItreeNode struct {
	Pattern     *Pattern
	Index       int
	ParentIndex int
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
}

type Itree struct {
	Nodes    []*ItreeNode
	EndEvent string
}

const MAX_CHILD_NODES = 3

func (it *Itree) buildRootNode(pattern *Pattern) (*ItreeNode, error) {
	// The pattern for root node is just Y (end_event).
	// If p = n(Y) / totalUsers. n(Y) is number of users with occurrence of Y.
	// RightGI is p(1-p)
	// RightFraction is 1.0
	// OverallGI = rightGI
	// confidence is p.
	// confidenceGain, leftGi, leftFraction, giniDrop are not defined.
	// parentIndex is set to -1.
	p := float64(pattern.OncePerUserCount) / float64(pattern.UserCount)
	giniImpurity := p * (1.0 - p)
	node := ItreeNode{
		Pattern:       pattern,
		ParentIndex:   -1,
		RightGI:       giniImpurity,
		RightFraction: 1.0,
		OverallGI:     giniImpurity,
		Confidence:    p,
	}
	log.WithFields(log.Fields{"node": node}).Info("Built root node.")
	return &node, nil
}

func (it *Itree) buildChildNode(
	pattern *Pattern, parentIndex int, allPatternCountsMap map[string]uint) (*ItreeNode, error) {

	parentNode := it.Nodes[parentIndex]
	parentPattern := parentNode.Pattern

	var fcr, fcp, fpr, fpp float64
	fcr = float64(pattern.OncePerUserCount)

	eLen := len(pattern.EventNames)
	currentPatternString := eventArrayToString(pattern.EventNames[:eLen-1])
	if c, ok := allPatternCountsMap[currentPatternString]; !ok {
		return nil, fmt.Errorf(fmt.Sprintf("Frequency missing for pattern %s", currentPatternString))
	} else {
		fcp = float64(c)
	}

	parentRuleString := eventArrayToString(parentPattern.EventNames)
	if c, ok := allPatternCountsMap[parentRuleString]; !ok {
		return nil, fmt.Errorf(fmt.Sprintf("Frequency missing for pattern %s", parentRuleString))
	} else {
		fpr = float64(c)
	}

	peLen := len(parentPattern.EventNames)
	if peLen == 1 {
		// Parent is root node. Count is all users.
		fpp = float64(parentPattern.UserCount)
	} else {
		parentPatternString := eventArrayToString(parentPattern.EventNames[:peLen-1])
		if c, ok := allPatternCountsMap[parentPatternString]; !ok {
			return nil, fmt.Errorf(fmt.Sprintf("Frequency missing for pattern %s", parentPatternString))
		} else {
			fpp = float64(c)
		}
	}

	if peLen != eLen-1 {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Current pattern(%s) and parent pattern(%s) not compatible", pattern.String(), parentRuleString))
	}

	rightGI := (fcr / fcp) * (1.0 - fcr/fcp)
	rightFraction := fcp / fpp
	leftGI := ((fpr - fcr) / (fpp - fcp)) * (1.0 - ((fpr - fcr) / (fpp - fcp)))
	var leftFraction float64 = 1.0 - rightFraction
	overallGI := rightFraction*rightGI + leftFraction*leftGI

	giniDrop := parentNode.RightGI - overallGI

	confidence := fcr / fcp
	confidenceGain := confidence - parentNode.Confidence

	node := ItreeNode{
		Pattern:        pattern,
		ParentIndex:    parentIndex,
		RightGI:        rightGI,
		RightFraction:  rightFraction,
		LeftGI:         leftGI,
		LeftFraction:   leftFraction,
		OverallGI:      overallGI,
		GiniDrop:       giniDrop,
		Confidence:     confidence,
		ConfidenceGain: confidenceGain,
	}
	log.WithFields(log.Fields{"node": node, "parent": parentPattern.String(),
		"fcr": fcr, "fcp": fcp, "fpr": fpr, "fpp": fpp}).Info("Built candidate child node.")
	return &node, nil
}

func (it *Itree) addNode(node *ItreeNode) int {
	node.Index = len(it.Nodes)
	it.Nodes = append(it.Nodes, node)
	return node.Index
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

func (it *Itree) buildAndAddChildNodes(
	parentNode *ItreeNode, candidatePattens []*Pattern, allPatternCountsMap map[string]uint) ([]*ItreeNode, error) {
	childNodes := []*ItreeNode{}
	parentPattern := parentNode.Pattern
	for _, p := range candidatePattens {
		if isChildSequence(parentPattern.EventNames, p.EventNames) {
			if cNode, err := it.buildChildNode(p, parentNode.Index, allPatternCountsMap); err != nil {
				log.WithFields(log.Fields{"err": err}).Errorf("Couldn't build child node")
				continue
			} else {
				childNodes = append(childNodes, cNode)
			}
		}
	}
	// Sort in decreasing order of giniDrop.
	sort.SliceStable(childNodes,
		func(i, j int) bool {
			return (childNodes[i].GiniDrop > childNodes[j].GiniDrop)
		})

	// Only top MAX_CHILD_NODES in order of drop in GiniImpurity are selected.
	if len(childNodes) > MAX_CHILD_NODES {
		childNodes = childNodes[:MAX_CHILD_NODES]
	}

	for _, cNode := range childNodes {
		it.addNode(cNode)
	}

	return childNodes, nil
}

func BuildNewItree(endEvent string, patternWrapper *PatternWrapper) (*Itree, error) {
	if endEvent == "" {
		return nil, fmt.Errorf("Missing end event")
	}

	itreeNodes := []*ItreeNode{}
	itree := Itree{
		Nodes:    itreeNodes,
		EndEvent: endEvent,
	}

	allPatternCountsMap := patternWrapper.perUserCountsMap

	var rootNodePattern *Pattern = nil
	candidatePatterns := []*Pattern{}
	for _, p := range patternWrapper.patterns {
		pLen := len(p.EventNames)
		if strings.Compare(endEvent, p.EventNames[pLen-1]) == 0 {
			candidatePatterns = append(candidatePatterns, p)
			if pLen == 1 {
				rootNodePattern = p
			}
		}
	}
	if rootNodePattern == nil {
		return nil, fmt.Errorf("Root node not found")
	}
	queue := []*ItreeNode{}

	if rootNode, err := itree.buildRootNode(rootNodePattern); err != nil {
		return nil, err
	} else {
		itree.addNode(rootNode)
		queue = append(queue, rootNode)
	}

	for len(queue) > 0 {
		parentNode := queue[0]
		if childNodes, err := itree.buildAndAddChildNodes(parentNode, candidatePatterns, allPatternCountsMap); err != nil {
			return nil, err
		} else {
			queue = append(queue, childNodes...)
		}
		queue = queue[1:]
	}

	log.WithFields(log.Fields{"itree": itree}).Info("Returning Itree.")
	return &itree, nil
}
