package operations

/*
File containing all the real time based operations
*/

import(
	"time"
	"sync"
	Log "data_simulator/logger"
	"data_simulator/config"
	"strconv"
	"fmt"
)

var globalTimer bool

func WaitForNSeconds(wg *sync.WaitGroup, duration int){
	defer wg.Done()
	Log.Debug.Printf("Waiting for Total Activity Time")
	WaitIfRealTime(duration)
	globalTimer = true
}

func WaitIfRealTime(duration int) {
	if(IsRealTime() == true){
		time.Sleep(time.Duration(duration) * time.Second)
	}
}

func IsRealTime() bool {
	if(config.ConfigV2.Real_Time == true){
		return true
	}
	return false
}

func ComputeActivityTimestamp(segmentConfig config.UserSegmentV2, eventCounter int, realTimeWait int) (int, int){
	
	var counter int
	if(realTimeWait != 0){
		counter = realTimeWait
	} else {
		counter = segmentConfig.Activity_ticker_in_seconds
	}
	timestamp, _ := strconv.Atoi(
		fmt.Sprintf("%v", 
		segmentConfig.Start_Time.Add(
			time.Second * time.Duration(
				eventCounter + counter)).Unix()))
	return timestamp, counter
}