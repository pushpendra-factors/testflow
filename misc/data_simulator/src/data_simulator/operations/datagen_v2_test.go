package operations

import(
	"testing"
)

func TestIsAllSegmentsDone(t *testing.T){
	segmentStatus := make(map[string]bool)
	segmentStatus["E1"] = true
	segmentStatus["E2"] = false
	segmentStatus["E3"] = true
	result1 := IsAllSegmentsDone(segmentStatus)
	if(result1 == true){
		t.Errorf("Expected false. Result %v",result1)
	}
	segmentStatus["E2"] = true
	result2 := IsAllSegmentsDone(segmentStatus)
	if(result2 == false){
		t.Errorf("Expected true. Result %v",result2)
	}
}