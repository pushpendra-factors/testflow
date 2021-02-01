package operations

/*
This file contains all core operations for probablity based event generation
*/

import (
	"data_simulator/config"
	"data_simulator/constants"
	Log "data_simulator/logger"
	"data_simulator/registration"
	"data_simulator/utils"
	"strconv"
	"sync"
)

var userAttributeMutex = &sync.Mutex{}

func OperateV2(env string) {
	//Declaring WaitGroup for SegmentLevel and newUser Concurrency
	var segmentWg sync.WaitGroup
	var newUserWg sync.WaitGroup
	var globalTimerWg sync.WaitGroup
	/*Calculating USERNAME indexing across segments
	Ex: Segment1 has 10 users and Segment2 has 5 users
	Segment1 will have users named U1,U2...U10 and
	Segment2 will have U11... U15
	New seeded users will have name from U16*/
	existingUsers := LoadExistingUsers(env) //thread safe
	// assuming the runs will be one per hour.
	var userCounter int = 1
	userIndex := make(map[string]int) // thread safe
	for item, element := range config.ConfigV2.User_segments {
		userIndex[item] = userCounter
		userCounter = userCounter + element.Number_of_users
	}
	Log.Debug.Printf("UserIndex Map %v", userIndex)
	/* Pre-Computing the following probablityRangeMaps per segment
	1. Activity
	2. Event
	3. Event Correlation
	4. New User seed probablity
	*/
	var probMap ProbMap

	probMap.yesOrNoProbMap = YesOrNoProbablityMap{
		ComputeYesOrNoProbablityMap(config.ConfigV2.New_user_probablity, "NewUser"),
		ComputeYesOrNoProbablityMap(config.ConfigV2.Custom_event_attribute_probablity, "Custom-Event"),
		ComputeYesOrNoProbablityMap(config.ConfigV2.Custom_user_attribute_probablity, "Custom-User"),
		ComputeYesOrNoProbablityMap(config.ConfigV2.Bring_existing_user, "Bring-User")}

	probMap.segmentProbMap = make(map[string]SegmentProbMap)
	for item, element := range config.ConfigV2.User_segments {
		probMap.segmentProbMap[item] = PreComputeRangeMap(element)
		// All the following shortcuts are for handling this error
		//  cannot assign to struct field in map
		// Need to fix it properly : TODO janani
		temp := probMap.segmentProbMap[item]
		temp.EventToEventAttributeMap = PreloadEventAttributes(probMap, element, probMap.segmentProbMap[item])
		temp.UserToUserAttributeMap = make(map[string]map[string]string)
		probMap.segmentProbMap[item] = temp
	}
	Log.Debug.Printf("RangeMaps %v", probMap)

	// Generate events per USER SEGMENT
	// segmentStatus variable is used to check if all the segments are done executing
	segmentStatus := make(map[string]bool)
	for item, element := range config.ConfigV2.User_segments {
		segmentWg.Add(1)
		segmentStatus[item] = false
		go OperateOnSegment(
			&segmentWg,
			probMap,
			item,
			element,
			probMap.segmentProbMap[item],
			userIndex[item],
			userIndex[item]+element.Number_of_users-1,
			segmentStatus,
			existingUsers)
	}

	Log.Debug.Printf("Main: Waiting for All Segments to finish")

	allSegmentsDone := false
	//newUserSegmentStatus is used to check if the new users seeded into the system are done executing
	newUserSegmentStatus := make(map[string]bool)

	// Seeding new users based on the seed probablity till the pre-defined segments executes
	i := userCounter
	globalTimer = false
	globalTimerWg.Add(1)
	go WaitForNSeconds(&globalTimerWg, config.ConfigV2.Activity_time_in_seconds)
	for (allSegmentsDone == false && IsRealTime() == true) || (IsRealTime() == true && globalTimer == false) {

		WaitIfRealTime(config.ConfigV2.New_user_poll_time)
		if SeedUserOrNot(probMap) == true {

			seg := GetRandomSegment()
			end := i + config.ConfigV2.Per_tick_new_user_seed_count - 1
			Log.Debug.Printf("Getting User %v - %v to the system with Segment %s", i, end, seg)
			newUserWg.Add(1)
			go OperateOnSegment(
				&newUserWg,
				probMap,
				seg, config.ConfigV2.User_segments[seg],
				probMap.segmentProbMap[seg],
				i,
				end,
				newUserSegmentStatus,
				existingUsers)
			i = end + 1
			allSegmentsDone = IsAllSegmentsDone(segmentStatus)

		}
	}
	segmentWg.Wait()
	Log.Debug.Printf("All Segments - Done !!!")
	newUserWg.Wait()
	Log.Debug.Printf("New Users - Done !!!")
	globalTimerWg.Wait()
	Log.Debug.Printf("Global Timer - Exit !!!")
	Log.Debug.Printf("Main - Done !!!")
	Log.Debug.Printf("Starting Upload to Cloud storage")
	if env != "development" {
		UploadData()
	}
}

func OperateOnSegment(segmentWg *sync.WaitGroup, probMap ProbMap,
	segmentName string, segment config.UserSegmentV2,
	segmentProbMap SegmentProbMap, userRangeStart int,
	userRangeEnd int, segmentStatus map[string]bool,
	existingUsers map[string]map[string]string) {

	defer segmentWg.Done()
	var wg sync.WaitGroup
	var userAttributes map[string]string
	var userId string
	Log.Debug.Printf("Main: Operating on %s with User Range %v - %v", segmentName, userRangeStart, userRangeEnd)
	//Generating events per user in the segment
	for i := userRangeStart; i <= userRangeEnd; i++ {
		wg.Add(1)
		userAttributeMutex.Lock()
		userFound := false
		if BringExistingUserOrNot(probMap) {
			noOfpeekInUserAttributeMap := 0
			for noOfpeekInUserAttributeMap < 5 {
				userId, userAttributes = PickFromExistingUsers(existingUsers)
				if userId != "" && !UserAlreadyExists(userId, segmentProbMap.UserToUserAttributeMap) {
					userFound = true
					break
				}
				noOfpeekInUserAttributeMap++
			}
			Log.Debug.Printf("Getting user %s back to system", userId)
			segmentProbMap.UserToUserAttributeMap[userId] = userAttributes
		}
		if userFound == false {
			userId = config.ConfigV2.User_id_prefix + strconv.Itoa(i)
			segmentProbMap.UserToUserAttributeMap[userId] = make(map[string]string)
			segmentProbMap.UserToUserAttributeMap[userId] = GetUserAttributes(probMap, segmentProbMap, segment)
			segmentProbMap.UserToUserAttributeMap[userId]["UserId"] = userId
			registration.WriterInstance.WriteUserData(FormatUserData(userId, segmentProbMap.UserToUserAttributeMap[userId]))
		}
		userAttributeMutex.Unlock()
		go GenerateEvents(
			&wg,
			probMap,
			segment,
			config.ConfigV2.Activity_time_in_seconds,
			userId,
			segmentProbMap)
	}

	Log.Debug.Printf("Main: Waiting for %s to finish for user Range %v - %v", segmentName, userRangeStart, userRangeEnd)
	wg.Wait()
	segmentStatus[segmentName] = true
	Log.Debug.Printf("Main: %s Completed for user Range %v - %v", segmentName, userRangeStart, userRangeEnd)
}

func GenerateEvents(wg *sync.WaitGroup, probMap ProbMap, segmentConfig config.UserSegmentV2, totalActivityDuration int, userId string, segmentProbMap SegmentProbMap) {

	defer wg.Done()
	var lastKnownGoodState string
	var realTimeWait int
	var decorators map[string]string

	// Setting attributes in output
	userAttributes := SetUserAttributes(segmentProbMap, segmentConfig, userId)

	Log.Debug.Printf("Starting %s for duration %v", userId, totalActivityDuration)
	i := 0
	for i < totalActivityDuration {

		activity := GetRandomActivity(segmentProbMap)
		// TODO: Janani Have enums for these
		if activity == constants.DOSOMETHING {
			event := GetRandomEvent(segmentProbMap)

			if event == constants.EVENTCORRELATION {

				event, realTimeWait = GetRandomEventWithCorrelation(
					&lastKnownGoodState,
					segmentProbMap, userAttributes, segmentConfig)

				decorators = GetEventDecorators(event, segmentProbMap)
				if utils.Contains(segmentConfig.Event_probablity_map.Correlation_matrix.Exit_events, event) {
					Log.Debug.Printf("User %s Exit events: %s", userId, event)
					break
				}
			}

			eventAttributes := SetEventAttributes(segmentProbMap, segmentConfig, event)
			eventAttributesWithDecorators := utils.AppendMaps(eventAttributes, decorators)
			timeStamp, counter := ComputeActivityTimestamp(segmentConfig, i, realTimeWait)
			op := FormatOutput(timeStamp, userId, event, userAttributes, eventAttributesWithDecorators, segmentConfig.Smart_events)
			registration.WriterInstance.WriteOutput(op)
			i = i + counter
			WaitIfRealTime(realTimeWait)

		}
		if activity == constants.EXIT {
			Log.Debug.Printf("Exit %s", userId)
			break
		}
	}
	Log.Debug.Printf("Done %s", userId)
}

func UploadData() {
	utils.CopyFilesToCloud(constants.LOCALOUTPUTFOLDER, constants.UNPROCESSEDFILESCLOUD, constants.BUCKETNAME, true)
}
