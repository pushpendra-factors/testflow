package delta

import (
	"bufio"
	"bytes"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	"factors/merge"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"os"
	"sort"
	"strings"

	T "factors/task"

	M "factors/model/model"
	"factors/model/store"

	"io/ioutil"

	"net/http"

	E "factors/event_match"

	log "github.com/sirupsen/logrus"
)

const (
	NO_OF_EVENTS_AT_EACH_LEVEL = 10
	PAGE_VIEW_CATEGORY = "Page Views"
	BUTTON_CLICKS_CATEGORY = "Button Clicks"
	CRM_EVENTS_CATEGORY = "CRM Events"
	SESSIONS_CATEGORY = "Sessions"
)

func GetDisplayName(displayNamesFromDB map[string]string, eventName string) string {
	displayNameEvents := make(map[string]string)
	standardEvents := U.STANDARD_EVENTS_DISPLAY_NAMES
	for event, displayName := range standardEvents {
		displayNameEvents[event] = strings.Title(displayName)
	}
	for event, displayName := range displayNamesFromDB {
		displayNameEvents[event] = strings.Title(displayName)
	}
	displayname, exist := displayNameEvents[eventName]
	if exist {
		return displayname
	}
	return U.CreateVirtualDisplayName(eventName)
	
}

func toCountEvent(eventDetails P.CounterEventFormat, nameFilter string, filters []M.QueryProperty) (bool, error) {
	var propFilter = make([]M.KPIFilter, 0)
	for _, filter := range filters {
		var kpiFilter M.KPIFilter
		kpiFilter.Entity = filter.Entity
		kpiFilter.Condition = filter.Operator
		kpiFilter.LogicalOp = filter.LogicalOp
		kpiFilter.PropertyName = filter.Property
		kpiFilter.PropertyDataType = filter.Type
		kpiFilter.Value = filter.Value
		propFilter = append(propFilter, kpiFilter)
	}
	return isEventToBeCounted(eventDetails, nameFilter, propFilter)
}

func appendExpandByCriteria(expandByProperty map[string][]PropertyEntityTypeMap, event P.CounterEventFormat, displayNamesFromDB map[string]string,) string {
	page_views, crm_event, button_click, session_event := false, false, false, false
	_, exist := event.EventProperties["$is_page_view"]
	if(!exist){
		page_views = true
	}
	hasPrefix := false
	for _, prefix := range U.CRMEventPrefixes {
		if(strings.HasPrefix(event.EventName, prefix)){
			hasPrefix = true
		}
	}
	if (hasPrefix == true){
		crm_event = false
	}
	_, exist = event.EventProperties["element_type"]
	if(!exist){
		button_click = true
	}
	if(!(event.EventName == "$session")){
		session_event = true
	}
	expandByProperties := make([]PropertyEntityTypeMap, 0)
	expandByPropertiesEvent := make([]PropertyEntityTypeMap, 0)
	expandByPropertiesUser := make([]PropertyEntityTypeMap, 0)
	properties, existInMap := expandByProperty[PAGE_VIEW_CATEGORY]
	if(page_views == true && existInMap){
		expandByProperties = append(expandByProperties, properties...)
	}
	properties, existInMap = expandByProperty[CRM_EVENTS_CATEGORY]
	if(crm_event == true && existInMap){
		expandByProperties = append(expandByProperties, properties...)
	}
	properties, existInMap = expandByProperty[BUTTON_CLICKS_CATEGORY]
	if(button_click == true && existInMap){
		expandByProperties = append(expandByProperties, properties...)
	}
	properties, existInMap = expandByProperty[SESSIONS_CATEGORY]
	if(session_event == true && existInMap){
		expandByProperties = append(expandByProperties, properties...)
	}
	properties, existInMap = expandByProperty[event.EventName]
	if(existInMap){
		expandByProperties = append(expandByProperties, properties...)
	}
	for _, property := range expandByProperties {
		if(property.Entity == "event"){
			expandByPropertiesEvent = append(expandByPropertiesEvent, property)
		}
		if(property.Entity == "user") {
			expandByPropertiesUser = append(expandByPropertiesUser, property)
		}
	}
	sort.Slice(expandByPropertiesEvent, func(i, j int) bool {
		return expandByPropertiesEvent[i].Property > expandByPropertiesEvent[j].Property
	})
	sort.Slice(expandByPropertiesUser, func(i, j int) bool {
		return expandByPropertiesUser[i].Property > expandByPropertiesUser[j].Property
	})
	finalEventString := GetDisplayName(displayNamesFromDB, event.EventName)
	for _, property := range expandByPropertiesEvent {
		value, exists := event.EventProperties[property.Property]
		valueString := ""
		if(exists){
			valueString = fmt.Sprintf("%v", value)
		} else {
			valueString = "<nil>"
		}
		finalEventString = fmt.Sprintf("%s:%s-%v", finalEventString, property.Property, valueString)
	}
	for _, property := range expandByPropertiesUser {
		value, exists := event.UserProperties[property.Property]
		valueString := ""
		if(exists){
			valueString = fmt.Sprintf("%v", value)
		} else {
			valueString = "<nil>"
		}
		finalEventString = fmt.Sprintf("%s:%s-%v", finalEventString, property.Property, valueString)
	}
	return finalEventString
}

type PropertyEntityTypeMap struct {
	Property string
	Entity string
}
func PathAnalysis(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	// Get the Queries that are to be computed
	// Evaluate queries one by one
	// Take the timestamp from the query
	// Take the files for those days
	// Read it sequentially
	// Check the number of steps from the steps
	// Keep n-steps always in memory (to support before queries)
	// If include events is populated, continue for all the non-events (! included)
	// If exclude events is populated, continue for all the excluded events
	// Keep writing each path to the file
	// After all the path are shorlisted, make a counter for level 1, level 2 and so on (5 at each level depending on count)
	// Write the final paths to a file

	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	archiveCloudManager := configs["archiveCloudManager"].(*filestore.FileManager)
	sortedCloudManager := configs["sortedCloudManager"].(*filestore.FileManager)
	modelCloudManager := configs["modelCloudManager"].(*filestore.FileManager)
	tmpCloudManager := configs["tmpCloudManager"].(*filestore.FileManager)
	hardPull := configs["hardPull"].(bool)
	beamConfig := configs["beamConfig"].(*merge.RunBeamConfig)
	useSortedFilesMerge := configs["useSortedFilesMerge"].(bool)
	blacklistedQueries := configs["blacklistedQueries"].([]string)
	blacklistedQueriesMap := make(map[string]bool)
	for _, id := range blacklistedQueries {
		blacklistedQueriesMap[id] = true
	}

	finalStatus := make(map[string]interface{})
	processedQueries := make([]string, 0)
	queries, _ := store.GetStore().GetAllSavedPathAnalysisEntityByProject(projectId)
	queryCountMap := make(map[string]int)
	
	_, displayNames := store.GetStore().GetDisplayNamesForAllEvents(projectId)
	for _, query := range queries {
		_, isBlacklisted := blacklistedQueriesMap[query.ID]
		if(isBlacklisted){
			finalStatus[query.ID] = "skipped - blacklisted"
			continue
		}
		timeNow := U.TimeNowZ().Unix()
		lastUpdated := query.UpdatedAt.Unix()
		lastCreated := query.CreatedAt.Unix()
		if (lastUpdated-lastCreated > 3600) && ((lastUpdated + 14400) > timeNow) {
			continue
		}
		store.GetStore().UpdatePathAnalysisEntity(projectId, query.ID, M.BUILDING)
		var actualQuery M.PathAnalysisQuery
		U.DecodePostgresJsonbToStructType(query.PathAnalysisQuery, &actualQuery)
		expandedProperty := make(map[string][]PropertyEntityTypeMap)
		for _, event := range actualQuery.IncludeEvents {
			_, exists := expandedProperty[event.Label] 
			if(!exists){
				expandedProperty[event.Label] = make([]PropertyEntityTypeMap, 0)
			}
			for _, prop := range event.ExpandProperty {
				expandedProperty[event.Label] = append(expandedProperty[event.Label], PropertyEntityTypeMap{Property: prop.Property, Entity: prop.Entity})
			}
		}
		for _, event := range actualQuery.IncludeGroup {
			_, exists := expandedProperty[event.Label] 
			if(!exists){
				expandedProperty[event.Label] = make([]PropertyEntityTypeMap, 0)
			}
			for _, prop := range event.ExpandProperty {
				expandedProperty[event.Label] = append(expandedProperty[event.Label], PropertyEntityTypeMap{Property: prop.Property, Entity: prop.Entity})
			}
		}
		if len(actualQuery.IncludeEvents) > 0 {
			actualQuery.IncludeEvents = append(actualQuery.IncludeEvents, actualQuery.Event)
		}
		groupId := int(0)
		if actualQuery.Group != "" && actualQuery.Group != "users" {
			eventNamesObj, eventnameerr := store.GetStore().GetEventName(actualQuery.Event.Label, projectId)
			if eventnameerr != http.StatusFound {
				store.GetStore().UpdatePathAnalysisEntity(projectId, query.ID, M.SAVED)
				finalStatus[query.ID] = "Failed to get event name"
				log.Error("Failed to get event name")
				return finalStatus, false
			}
			groupNameFromDb, _ := store.GetStore().IsGroupEventName(projectId, actualQuery.Event.Label, eventNamesObj.ID)
			if groupNameFromDb != "" {
				if actualQuery.Group != groupNameFromDb {
					store.GetStore().UpdatePathAnalysisEntity(projectId, query.ID, M.SAVED)
					finalStatus[query.ID] = "group names mismatch"
					log.Error("group names mismatch", actualQuery.Group, groupNameFromDb)
					return finalStatus, false
				} else {
					groupId = int(0)
				}
			} else {
				groupDetails, groupErr := store.GetStore().GetGroup(projectId, actualQuery.Group)
				if groupErr != http.StatusFound {
					store.GetStore().UpdatePathAnalysisEntity(projectId, query.ID, M.SAVED)
					finalStatus[query.ID] = "Failed to get group details"
					log.Error("Failed to get group details")
					return finalStatus, false
				}
				groupId = groupDetails.ID
			}
		}

		cfCloudPath, cfCloudName, err := merge.MergeAndWriteSortedFile(projectId, U.DataTypeEvent, "", actualQuery.StartTimestamp, actualQuery.EndTimestamp,
			archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, hardPull, groupId, useSortedFilesMerge, true, true)
		if err != nil {
			store.GetStore().UpdatePathAnalysisEntity(projectId, query.ID, M.SAVED)
			finalStatus[query.ID] = err.Error()
			log.WithError(err).Error("Failed creating events file")
			return finalStatus, false
		}

		log.Info("Processing Query ID: ", query.ID, " query: ", actualQuery)
		log.Info("Starting cloud events file get")
		// "projects/2251799829000005/", "events.txt"

		eReader, err := (*sortedCloudManager).Get(cfCloudPath, cfCloudName)
		if err != nil {
			store.GetStore().UpdatePathAnalysisEntity(projectId, query.ID, M.SAVED)
			log.WithFields(log.Fields{"err": err, "eventFilePath": cfCloudPath,
				"eventFileName": cfCloudName}).Error("Failed downloading  file from cloud.")
			finalStatus[query.ID] = "Failed downloading  file from cloud."
			return finalStatus, false
		}
		scanner := bufio.NewScanner(eReader)
		const maxCapacity = 30 * 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)
		pathanalysisTempPath, pathanalysisTempName := diskManager.GetPathAnalysisTempFilePathAndName(query.ID, projectId)
		log.Info("creating path analysis temp file. Path: ", pathanalysisTempPath, " Name: ", pathanalysisTempName)
		if err := os.MkdirAll(pathanalysisTempPath, os.ModePerm); err != nil {
			log.WithError(err).Error("Failed creating path analysis file")
		}
		currentFile, err := os.Create(pathanalysisTempPath + pathanalysisTempName)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Failed creating pathanalysis temp file")
		}
		resultFile, err1 := os.Create(pathanalysisTempPath + "result.txt")
		if err1 != nil {
			log.WithFields(log.Fields{"err": err}).Error("Failed creating pathanalysis temp file")
		}
		AllFilters := append(actualQuery.Event.Filter, actualQuery.Filter...)
		queryCriteria := E.EventCriterion{
			Name:                actualQuery.Event.Label,
			EqualityFlag:        true,
			FilterCriterionList: E.MapFilterProperties(AllFilters),
		}
		log.Info("Transformed Query: ", queryCriteria)
		finalEvents := make([]P.CounterEventFormat, 0)
		shouldStopForThisUser := false
		stepsProcessed := 0
		prevId := ""
		matched := false
		lineNo := 0
		queryCount := 0
		for scanner.Scan() {
			txtline := scanner.Text()
			lineNo++
			if lineNo%1000 == 0 {
				log.Info("Scanned ", lineNo, " lines")
			}
			var event P.CounterEventFormat
			err := json.Unmarshal([]byte(txtline), &event) // TODO: Add error check.
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Failed unmarshaling file")
			}
			if event.EventTimestamp < actualQuery.StartTimestamp || event.EventTimestamp > actualQuery.EndTimestamp {
				continue
			}
			if ok, err := toCountEvent(event, actualQuery.Event.Label, append(actualQuery.Filter, actualQuery.Event.Filter...)); err != nil {
				log.WithError(err).Error("error counting")
			} else if ok {
				queryCount++
			}
			currentId := event.UserId
			if groupId == 1 {
				currentId = event.Group1UserId
			} else if groupId == 2 {
				currentId = event.Group2UserId
			} else if groupId == 3 {
				currentId = event.Group3UserId
			} else if groupId == 4 {
				currentId = event.Group4UserId
			}
			if currentId == "" {
				continue
			}
			if prevId == "" {
				prevId = currentId
			}
			if currentId != prevId {
				if len(finalEvents) > 0 && matched == true {
					pwmBytes, _ := json.Marshal(finalEvents)
					pString := string(pwmBytes)
					pString = pString + "\n"
					pBytes := []byte(pString)
					_, err = currentFile.Write(pBytes)
				}
				finalEvents = make([]P.CounterEventFormat, 0)
				shouldStopForThisUser = false
				stepsProcessed = 0
				matched = false
				// Write the existing list and reset
				// Reset all the previous events and after events
			}
			if shouldStopForThisUser && currentId == prevId {
				continue
			}
			prevId = currentId
			isIncludePresent := false
			if(len(actualQuery.IncludeEvents) == 0 && len(actualQuery.IncludeGroup) == 0){
				isIncludePresent = true
			}
			if len(actualQuery.IncludeEvents) > 0 {
				if StringIn(actualQuery.IncludeEvents, event.EventName) {
					isIncludePresent = true
				}
			}
			if len(actualQuery.ExcludeEvents) > 0 {
				if StringIn(actualQuery.ExcludeEvents, event.EventName) {
					continue
				}
			}
			if(len(actualQuery.IncludeGroup) > 0){
				for _, group := range actualQuery.IncludeGroup {
					allFiltersForIncludeGroup := append(group.Filter, actualQuery.Filter...)
					matchesFilter := E.EventMatchesFilterCriterionList(projectId, event.UserProperties, event.EventProperties, E.MapFilterProperties(allFiltersForIncludeGroup))
					if(group.Label == PAGE_VIEW_CATEGORY){
						_, exist := event.EventProperties["$is_page_view"]
						if(exist && matchesFilter){
							isIncludePresent = true
						}
					}
					if(group.Label == CRM_EVENTS_CATEGORY){
						hasPrefix := false
						for _, prefix := range U.CRMEventPrefixes {
							if(strings.HasPrefix(event.EventName, prefix)){
								hasPrefix = true
							}
						}
						if (hasPrefix == true && matchesFilter){
							isIncludePresent = true
						}
					}
					if(group.Label == BUTTON_CLICKS_CATEGORY){
						_, exist := event.EventProperties["element_type"]
						if(exist && matchesFilter){
							isIncludePresent = true
						}
					}
					if(group.Label == SESSIONS_CATEGORY){
						if((event.EventName == "$session") && matchesFilter){
							isIncludePresent = true
						}
					}
				}
			}
			if(isIncludePresent == false){
				continue
			}
			if actualQuery.EventType == M.STARTSWITH {
				if E.EventMatchesCriterion(projectId, event.EventName, event.UserProperties, event.EventProperties, queryCriteria) { // check if a particular event matches the request criteria
					matched = true
				}
				if matched == true {
					if actualQuery.AvoidRepeatedEvents == true && IfStringArrayContains(finalEvents, event) {
						continue
					}
					finalEvents = append(finalEvents, event)
					stepsProcessed++
				}
				if stepsProcessed == actualQuery.NumberOfSteps+1 {
					shouldStopForThisUser = true
				}
			}
			if actualQuery.EventType == M.ENDSWITH {
				if actualQuery.AvoidRepeatedEvents == true && IfStringArrayContains(finalEvents, event) {
					finalEvents = RemoveFromArray(finalEvents, event)
				}
				finalEvents = EnqueueWithSize(finalEvents, event, actualQuery.NumberOfSteps+1)
				if E.EventMatchesCriterion(projectId, event.EventName, event.UserProperties, event.EventProperties, queryCriteria) { // check if a particular event matches the request criteria
					matched = true
				}
				// Check this has to be the last event or the first event
				if matched == true {
					shouldStopForThisUser = true
				}
			}
		}
		err = eReader.Close()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Failed closing events reader.")
		}
		queryCountMap[query.Title] = queryCount
		log.Infof("queryCount: %d", queryCount)
		scanner, err = T.OpenEventFileAndGetScanner(pathanalysisTempPath + pathanalysisTempName)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Failed opening file and getting scanner of patterns file.")
		}
		steps := make(map[string]int)
		for scanner.Scan() {
			txtline := scanner.Text()
			events := make([]P.CounterEventFormat, 0)
			err := json.Unmarshal([]byte(txtline), &events) // TODO: Add error check.
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Failed unmarshaling file")
			}
			soFar := ""
			if actualQuery.EventType == M.STARTSWITH {
				for _, event := range events {
					if soFar == "" {
						soFar = appendExpandByCriteria(expandedProperty, event, displayNames)
					} else	 {
						soFar = soFar + "," + appendExpandByCriteria(expandedProperty, event, displayNames)
					}
					steps[soFar]++
				}
			}
			if actualQuery.EventType == M.ENDSWITH {
				for i := len(events) - 1; i >= 0; i-- {
					if soFar == "" {
						soFar = appendExpandByCriteria(expandedProperty, events[i], displayNames)
					} else {
						soFar = soFar + "," + appendExpandByCriteria(expandedProperty, events[i], displayNames)
					}
					steps[soFar]++
				}
			}
		}
		maxcount := make(map[int]map[string]int)
		for event, count := range steps {
			noOfEvents := strings.Split(event, ",")
			_, exist := maxcount[len(noOfEvents)]
			if !exist {
				maxcount[len(noOfEvents)] = make(map[string]int)
			}
			maxcount[len(noOfEvents)][event] = count
		}
		shortlistedEvents := make(map[int]map[string]int)
		for i := 1; i <= actualQuery.NumberOfSteps+1; i++ {
			if i == 1 {
				shortlistedEvents[1] = maxcount[1]
			} else {
				for rootEvent, rootCount := range shortlistedEvents[i-1] {
					if strings.Contains(rootEvent, "OTHERS") {
						continue
					}
					eventCountArray := make([]eventCount, 0)
					eventCountArrayTrimmed := make([]eventCount, 0)
					for event, count := range maxcount[i] {
						rootEventWithComma := rootEvent + ","
						if strings.HasPrefix(event, rootEventWithComma) {
							eventCountArray = append(eventCountArray, eventCount{event: event, count: count})
						}
					}
					sort.Slice(eventCountArray, func(i, j int) bool {
						return eventCountArray[i].count > eventCountArray[j].count
					})
					eventCountArrayTrimmed = eventCountArray
					if len(eventCountArray) > NO_OF_EVENTS_AT_EACH_LEVEL {
						eventCountArrayTrimmed = eventCountArray[0:NO_OF_EVENTS_AT_EACH_LEVEL]
					}
					selectedCount := 0
					for _, eventsTrimmed := range eventCountArrayTrimmed {
						selectedCount += eventsTrimmed.count
					}
					eventCountArrayTrimmed = append(eventCountArrayTrimmed, eventCount{event: rootEvent + ",OTHERS", count: rootCount - selectedCount})
					_, exists := shortlistedEvents[i]
					if !exists {
						shortlistedEvents[i] = make(map[string]int)
					}
					for _, eventcount := range eventCountArrayTrimmed {
						shortlistedEvents[i][eventcount.event] = eventcount.count
					}

				}
			}
			// Prefixing with step number
			eventsFormatted := make(map[string]int)
			for event, count := range shortlistedEvents[i] {
				events := strings.Split(event, ",")
				eventNameAppend := ""
				for i, eventname := range events {
					if eventNameAppend == "" {
						eventNameAppend = fmt.Sprintf("%v:%v", i, eventname)
					} else {
						eventNameAppend = fmt.Sprintf("%v,%v:%v", eventNameAppend, i, eventname)
					}
				}
				eventsFormatted[eventNameAppend] = count
			}
			// Reversing the order for ENDSWITH
			eventsOrdered := make(map[string]int)
			if actualQuery.EventType == M.ENDSWITH {
				for event, count := range eventsFormatted {
					events := strings.Split(event, ",")
					eventTag := ""
					for i := len(events) - 1; i >= 0; i-- {
						if eventTag == "" {
							eventTag = events[i]
						} else {
							eventTag = eventTag + "," + events[i]
						}
					}
					eventsOrdered[eventTag] = count
				}
			} else {
				eventsOrdered = eventsFormatted
			}
			pwmBytes, _ := json.Marshal(eventsOrdered)
			pString := string(pwmBytes)
			pString = pString + "\n"
			pBytes := []byte(pString)
			_, err = resultFile.Write(pBytes)
		}
		WriteResultsToCloud(diskManager, modelCloudManager, query.ID, projectId)
		processedQueries = append(processedQueries, query.ID)
		if len(processedQueries) > 0 {
			finalStatus["PROCESSED QUERIES"] = processedQueries
		}
		store.GetStore().UpdatePathAnalysisEntity(projectId, query.ID, M.ACTIVE)
	}
	log.Infof("queryCountMap: %v, projectID; %d", queryCountMap, projectId)
	return finalStatus, true
}

func WriteResultsToCloud(diskManager *serviceDisk.DiskDriver, cloudManager *filestore.FileManager, query_id string, project_id int64) error {

	path, _ := diskManager.GetPathAnalysisTempFilePathAndName(query_id, project_id)
	resultLocalPath := path + "result.txt"
	scanner, err := T.OpenEventFileAndGetScanner(resultLocalPath)
	result := make(map[int]map[string]int)
	i := 0
	for scanner.Scan() {
		txtline := scanner.Text()
		i++
		events := make(map[string]int)
		json.Unmarshal([]byte(txtline), &events) // TODO: Add error check.
		result[i] = events
	}
	path, _ = (*cloudManager).GetPathAnalysisTempFilePathAndName(query_id, project_id)
	resultJson, err := json.Marshal(result)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("failed to unmarshal result Info.")
		return err
	}
	err = (*cloudManager).Create(path, "result.txt", bytes.NewReader(resultJson))
	if err != nil {
		log.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return err
	}
	return err
}

type eventCount struct {
	event string
	count int
}

func StringIn(events []M.PathAnalysisEvent, key string) bool {
	for _, event := range events {
		if key == event.Label {
			return true
		}
	}
	return false
}

func StringNotIn(events []M.PathAnalysisEvent, key string) bool {
	for _, event := range events {
		if key == event.Label {
			return false
		}
	}
	return true
}

func IfStringArrayContains(events []P.CounterEventFormat, key P.CounterEventFormat) bool {
	for _, event := range events {
		if event.EventName == key.EventName {
			return true
		}
	}
	return false
}

func EnqueueWithSize(events []P.CounterEventFormat, key P.CounterEventFormat, size int) []P.CounterEventFormat {
	if len(events) < size {
		events = append(events, key)
		return events
	}
	newEvents := make([]P.CounterEventFormat, 0)
	for i, event := range events {
		if i <= len(events)-size {
			continue
		}
		newEvents = append(newEvents, event)
	}
	newEvents = append(newEvents, key)
	return newEvents
}

func RemoveFromArray(events []P.CounterEventFormat, key P.CounterEventFormat) []P.CounterEventFormat {
	newEvents := make([]P.CounterEventFormat, 0)
	for _, event := range events {
		if event.EventName != key.EventName {
			newEvents = append(newEvents, event)
		}
	}
	return newEvents
}

func GetPathAnalysisData(projectId int64, id string) map[int]map[string]int {
	path, _ := C.GetCloudManager().GetPathAnalysisTempFilePathAndName(id, projectId)
	fmt.Println(path)
	reader, err := C.GetCloudManager().Get(path, "result.txt")
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return nil
	}
	result := make(map[int]map[string]int)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return nil
	}
	err = json.Unmarshal(data, &result)
	return result
}
