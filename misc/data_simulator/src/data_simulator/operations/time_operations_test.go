package operations

/*
Test for dataGenV2 operations
*/

import(
	"testing"
	"data_simulator/config"
	"time"
)

func TestComputeActivityTimestamp(t *testing.T){
	config := config.UserSegmentV2{
		Activity_ticker_in_seconds: 5, 
		Start_Time: time.Date(2020, time.Month(1), 1, 0, 0, 0, 0, time.UTC)}
	resultTimeStamp, resultCounter  := ComputeActivityTimestamp(config, 4, 0)
	if(resultCounter != 5){
		t.Errorf("Expected: 4 Result: %v", resultCounter)
	}
	if(resultTimeStamp != 1577836809){
		t.Errorf("Expected: %v Result: %v", 1577836804, 
			resultTimeStamp)
	}
	resultTimeStamp, resultCounter  = ComputeActivityTimestamp(config, 4, 3)
	if(resultCounter != 3){
		t.Errorf("Expected: 3 Result: %v", resultCounter)
	}
	if(resultTimeStamp != 1577836807){
		t.Errorf("Expected: %v Result: %v", 1577836807, 
			resultTimeStamp)
	}
	resultTimeStamp, resultCounter  = ComputeActivityTimestamp(config, 4, 10)
	if(resultCounter != 10){
		t.Errorf("Expected: 10 Result: %v", resultCounter)
	}
	if(resultTimeStamp != 1577836814){
		t.Errorf("Expected: %v Result: %v", 1577836814, 
			resultTimeStamp)
	}
}