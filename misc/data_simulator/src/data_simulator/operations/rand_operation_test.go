package operations

import(
	"testing"
	"data_simulator/config"
)


func TestGetRandomSegment(t *testing.T){
	config.ConfigV2.User_segments = make(map[string]config.UserSegmentV2)
	config.ConfigV2.User_segments["Segment1"] = config.UserSegmentV2{}
	result1 := GetRandomSegment()
	if(result1 != "Segment1"){
		t.Errorf("Expected Segment1. Result %v", result1)
	}
	config.ConfigV2.User_segments["Segment2"] = config.UserSegmentV2{}
	result2 := GetRandomSegment()
	if(!(result2 == "Segment1" || result2 == "Segment2")){
		t.Errorf("Expected Segment1 || segment2. Result %v", result2)
	}
}