package operations

import (
	"data_simulator/config"
	"testing"
	"time"
)

func TestFormatOutput(t *testing.T) {
	var userSegment config.UserSegmentV2
	userSegment.Activity_ticker_in_seconds = 1
	userSegment.Start_Time = time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	result1 := FormatOutput(1258490099, "U1", "E1", nil, nil, nil)
	output1 := "{\"user_id\":\"U1\",\"event_name\":\"E1\",\"timestamp\":1258490099,\"user_properties\":{},\"event_properties\":{},\"smart_event\":\"\"}"
	if result1 != output1 {
		t.Errorf("Expected %v Result %v", output1, result1)
	}

	se := make(map[string]string)
	se["E1"] = "CS"
	attr := make(map[string]string)
	attr["A1"] = "U1"
	result2 := FormatOutput(1258490099, "U1", "E1", attr, attr, se)
	output2 := "{\"user_id\":\"U1\",\"event_name\":\"E1\",\"timestamp\":1258490099,\"user_properties\":{\"A1\":\"U1\"},\"event_properties\":{\"A1\":\"U1\"},\"smart_event\":\"CS\"}"
	if result2 != output2 {
		t.Errorf("Expected %v Result %v", output2, result2)
	}
}
