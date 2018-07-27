package tests

import (
	"fmt"
	P "pattern"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildNewItree(t *testing.T) {
	// Hypothetical data on which Itree is built, ending with Y.
	// User1:  B  C  A  A
	// User2:  A  A  B  C  Y  C
	// User3:  C  A  B
	// User4:  A  B  C  C  Y
	// User5:  A  A  B
	// User6:  A  B  C  Y  A
	// User7:  B  Y  C
	// User8:  C  Y  A
	// User9:  A  B  C
	// User10: A  B  A  C  B  C  Y
	var userCount uint = 10

	type testPatternInfo struct {
		eventNames          []string
		perUserCount, count uint
	}
	// Patterns with zero counts are not listed.
	// Pattern, PerUserCount, Count
	tpis := []testPatternInfo{
		testPatternInfo{eventNames: []string{"A"}, perUserCount: 9, count: 14},
		testPatternInfo{eventNames: []string{"B"}, perUserCount: 9, count: 10},
		testPatternInfo{eventNames: []string{"C"}, perUserCount: 9, count: 12},
		testPatternInfo{eventNames: []string{"Y"}, perUserCount: 6, count: 6},
		testPatternInfo{eventNames: []string{"A", "B"}, perUserCount: 7, count: 8},
		testPatternInfo{eventNames: []string{"A", "C"}, perUserCount: 5, count: 5},
		testPatternInfo{eventNames: []string{"A", "Y"}, perUserCount: 4, count: 4},
		testPatternInfo{eventNames: []string{"B", "A"}, perUserCount: 3, count: 3},
		testPatternInfo{eventNames: []string{"B", "C"}, perUserCount: 7, count: 8},
		testPatternInfo{eventNames: []string{"B", "Y"}, perUserCount: 5, count: 5},
		testPatternInfo{eventNames: []string{"C", "A"}, perUserCount: 4, count: 4},
		testPatternInfo{eventNames: []string{"C", "B"}, perUserCount: 2, count: 2},
		testPatternInfo{eventNames: []string{"C", "Y"}, perUserCount: 5, count: 5},
		testPatternInfo{eventNames: []string{"Y", "A"}, perUserCount: 2, count: 2},
		testPatternInfo{eventNames: []string{"Y", "C"}, perUserCount: 2, count: 2},
		testPatternInfo{eventNames: []string{"A", "C", "B"}, perUserCount: 1, count: 1},
		testPatternInfo{eventNames: []string{"A", "B", "C"}, perUserCount: 5, count: 5},
		testPatternInfo{eventNames: []string{"A", "Y", "C"}, perUserCount: 1, count: 1},
		testPatternInfo{eventNames: []string{"A", "B", "Y"}, perUserCount: 4, count: 4},
		testPatternInfo{eventNames: []string{"A", "C", "Y"}, perUserCount: 4, count: 4},
		testPatternInfo{eventNames: []string{"B", "C", "A"}, perUserCount: 2, count: 2},
		testPatternInfo{eventNames: []string{"B", "Y", "A"}, perUserCount: 1, count: 1},
		testPatternInfo{eventNames: []string{"B", "A", "C"}, perUserCount: 1, count: 1},
		testPatternInfo{eventNames: []string{"B", "Y", "C"}, perUserCount: 2, count: 2},
		testPatternInfo{eventNames: []string{"B", "A", "Y"}, perUserCount: 1, count: 1},
		testPatternInfo{eventNames: []string{"B", "C", "Y"}, perUserCount: 4, count: 4},
		testPatternInfo{eventNames: []string{"C", "Y", "A"}, perUserCount: 2, count: 2},
		testPatternInfo{eventNames: []string{"C", "A", "B"}, perUserCount: 1, count: 1},
		testPatternInfo{eventNames: []string{"C", "B", "Y"}, perUserCount: 1, count: 1},
		testPatternInfo{eventNames: []string{"A", "C", "B", "Y"}, perUserCount: 1, count: 1},
		testPatternInfo{eventNames: []string{"A", "B", "C", "Y"}, perUserCount: 4, count: 4},
		testPatternInfo{eventNames: []string{"A", "B", "Y", "C"}, perUserCount: 1, count: 1},
		testPatternInfo{eventNames: []string{"B", "A", "C", "Y"}, perUserCount: 1, count: 1},
		testPatternInfo{eventNames: []string{"B", "C", "Y", "A"}, perUserCount: 1, count: 1},
	}
	patterns := []*P.Pattern{}
	for _, tpi := range tpis {
		p, _ := P.NewPattern(tpi.eventNames)
		p.Count = tpi.count
		p.OncePerUserCount = tpi.perUserCount
		p.UserCount = userCount
		patterns = append(patterns, p)
	}
	pw := P.NewPatternWrapper(patterns)
	itree, err := P.BuildNewItree("Y", pw)
	assert.Nil(t, err)
	assert.NotNil(t, itree)

	// Expected iTree.
	//
	//                       Y(0)
	//                       |
	//           -------------------------------
	//           |                |             |
	//          AY(1)            BY(2)         CY(3)
	//           |                |             |
	//        -----------      ---------     ------------
	//        |         |      |       |     |           |
	//       ACY(4)    ABY(5) BAY(6)  BCY(7) CAY(Skip)  CBY(8)
	//        |         |
	//        |         |
	//      ACBY(9)    ABCY(10)
	type expectedNode struct {
		patternString                                            string
		index, parentIndex                                       int
		rightGi, overallGi, giniDrop, confidence, confidenceGain float64
	}
	node0 := expectedNode{
		patternString: "Y",
		index:         0, parentIndex: -1,
		rightGi: 0.24, overallGi: 0.24,
		confidence: 0.6,
	}
	node1 := expectedNode{
		patternString: "A,Y",
		index:         1, parentIndex: 0,
		rightGi: 0.246914, overallGi: 0.022222, giniDrop: 0.217778,
		confidence: 0.444444, confidenceGain: -0.155556,
	}
	node2 := expectedNode{
		patternString: "B,Y",
		index:         2, parentIndex: 0,
		rightGi: 0.246914, overallGi: 0.222222, giniDrop: 0.017778,
		confidence: 0.555556, confidenceGain: -0.044444,
	}
	node3 := expectedNode{
		patternString: "C,Y",
		index:         3, parentIndex: 0,
		rightGi: 0.246914, overallGi: 0.222222, giniDrop: 0.017778,
		confidence: 0.555556, confidenceGain: -0.044444,
	}

	node4 := expectedNode{
		patternString: "A,C,Y",
		index:         4, parentIndex: 1,
		rightGi: 0.16, overallGi: 0.088889, giniDrop: 0.158025,
		confidence: 0.80, confidenceGain: 0.355556,
	}
	node5 := expectedNode{
		patternString: "A,B,Y",
		index:         5, parentIndex: 1,
		rightGi: 0.244898, overallGi: 0.190476, giniDrop: 0.056438,
		confidence: 0.571429, confidenceGain: 0.126985,
	}
	node9 := expectedNode{
		patternString: "A,C,B,Y",
		index:         9, parentIndex: 4,
		rightGi: 0.0, overallGi: 0.15, giniDrop: 0.01,
		confidence: 1.0, confidenceGain: 0.2,
	}
	node10 := expectedNode{
		patternString: "A,B,C,Y",
		index:         10, parentIndex: 5,
		rightGi: 0.16, overallGi: 0.114286, giniDrop: 0.130612,
		confidence: 0.8, confidenceGain: 0.228571,
	}
	expectedNodes := []*expectedNode{
		&node0, &node1, &node2, &node3, &node4, &node5, &node9, &node10}

	for i, eNode := range expectedNodes {
		aNode := itree.Nodes[eNode.index]
		assert.Equal(t, eNode.patternString, aNode.Pattern.String())
		assert.Equal(t, eNode.parentIndex, aNode.ParentIndex, fmt.Sprintf("Node: %s", eNode.patternString))
		assert.InDelta(t, eNode.rightGi, aNode.RightGI, 0.000001, fmt.Sprintf("Node: %s", eNode.patternString))
		assert.InDelta(t, eNode.overallGi, aNode.OverallGI, 0.000001, fmt.Sprintf("Node: %s", eNode.patternString))
		assert.InDelta(t, eNode.confidence, aNode.Confidence, 0.000001, fmt.Sprintf("Node: %s", eNode.patternString))
		if i != 0 {
			assert.InDelta(t, eNode.giniDrop, aNode.GiniDrop, 0.000001, fmt.Sprintf("Node: %s", eNode.patternString))
			assert.InDelta(t, eNode.confidenceGain, aNode.ConfidenceGain, 0.000001, fmt.Sprintf("Node: %s", eNode.patternString))
		}
	}

}
