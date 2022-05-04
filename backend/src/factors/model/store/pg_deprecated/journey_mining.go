package postgres

import (
	"bufio"
	"encoding/json"
	"factors/model/model"
	"fmt"
	"runtime"
	osDebug "runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"factors/filestore"
	serviceGCS "factors/services/gcstorage"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

// General idea of plotting Journey Map would be as follows.
// Consider events E1 to E20 and journey is to be mapped for E1 to E10 with E10 as goal event.
//     1. Calculate all paths and count of users for each path.
//     2. Each user should be present in exactly one path. Longest possible one.
//            i. If user has path E1-E2-E3-E4. He will not be counted in sub paths E2-E3-E4 or E3-E4.
//     3. Keep one month or configurable lookback period (Not supported right now).
//            i. But goal event must have happened in the start and end time range.
//     4. For a path E1-E3-E7-E10 to be counted for a user, no other event should have occurred,
//        neither from E1 to E10 set nor outside E10.
//     5. For a path E1-E3-E7-E10, older events are allowed to repeat as we move forward. Ex:
//            i. User with path E1-E3-E1-E1-E3-E7-E3-E10 is valid and will be truncated to E1-E3-E7-E10.
//     6. For a path E1-E3-E7-E10, events outside our universe are allowed to appear. Ex:
//            i. User with path E1-E3-E79-E7-E10 with E79 is valid and will be counted in E1-E3-E7-E10.
//
// Approach:
//     1. Take files from build seq or archive job for coalesced and denormalized events data with user properties.
//     2. If there are weekly files, parse them all. Create buckets of events by user like map[userID]EventsList.
//     3. While adding to bucket, keep track of if goal event occurred for user. Will be helpful to filter.
//     4. Parse user events bucket following the path guidelines. Reject any unwanted (repeated / outside universe) events.
//     5. From the pruned events, build path strings and group similar path user together.

var debug = false
var debugUserID = ""
var allowRepeatedGoals = true

// JourneyPatternMap holds journey as E1 -> E2 -> E3 along with users and count.
type JourneyPatternMap struct {
	JourneyPath string
	UserIDs     []string
	Count       int64
}

// ArchiveEventFormat Copy of struct present in model/archival with extra fields.
type ArchiveEventFormat struct {
	EventID            string                 `json:"event_id" bigquery:"event_id"`
	UserID             string                 `json:"user_id" bigquery:"user_id"`
	UserJoinTimestamp  int64                  `json:"user_join_timestamp" bigquery:"user_join_timestamp"`
	EventName          string                 `json:"event_name" bigquery:"event_name"`
	EventTimestamp     time.Time              `json:"event_timestamp" bigquery:"event_timestamp"`
	SessionID          string                 `json:"session_id" bigquery:"session_id"`
	EventProperties    string                 `json:"event_properties" bigquery:"event_properties"`
	UserProperties     string                 `json:"user_properties" bigquery:"user_properties"`
	EventPropertiesMap map[string]interface{} `json:"-"`
	UserPropertiesMap  map[string]interface{} `json:"-"`
	EventTimestampUnix int64                  `json:"-"`
}

func getGraphKeyAliasForIndex(index int) string {
	return fmt.Sprintf("E%d", index)
}

func getEventAsString(event interface{}) string {
	eventJSON, _ := json.Marshal(event)
	return string(eventJSON)
}

func getValueFromProperties(propertyMap map[string]interface{}, propertyName string) interface{} {
	value, found := propertyMap[propertyName]
	if !found || value == nil || fmt.Sprintf("%v", value) == "" {
		return model.PropertyValueNone
	}
	return fmt.Sprintf("%v", value)
}

func updateEventWithParsedFields(denormalizedEvent *ArchiveEventFormat) error {
	var userPropertiesMap, eventPropertiesMap map[string]interface{}
	if err := json.Unmarshal([]byte(denormalizedEvent.UserProperties), &userPropertiesMap); err != nil {
		log.WithError(err).Errorf("failed to decode user properties: %s", denormalizedEvent.UserProperties)
		return err
	} else if err := json.Unmarshal([]byte(denormalizedEvent.EventProperties), &eventPropertiesMap); err != nil {
		log.WithError(err).Errorf("failed to decode event properties: %s", denormalizedEvent.EventProperties)
		return err
	}
	denormalizedEvent.UserPropertiesMap = userPropertiesMap
	denormalizedEvent.EventPropertiesMap = eventPropertiesMap
	denormalizedEvent.EventTimestampUnix = denormalizedEvent.EventTimestamp.In(
		U.GetTimeLocationFor(U.TimeZoneStringIST)).Unix()
	return nil
}

// Specfic to chargebee monthly report analysis.
func getUserEventsFromFileChargebee(fileName string, segmentAGoals, segmentBGoals, segmentCGoals []model.QueryEventWithProperties,
	userEventsMap *map[string][]ArchiveEventFormat, segmentAUsers, segmentBUsers, segmentCUsers *map[string]bool,
	journeyEvents []model.QueryEventWithProperties, startTime, endTime int64, cloudManager filestore.FileManager) error {
	logCtx := log.WithFields(log.Fields{
		"Method": "getUserEventsFromFile",
	})

	logCtx.Infof("Scanning events from file %s", fileName)
	fileNameSplit := strings.Split(fileName, "/")
	filePath, fileName := strings.Join(fileNameSplit[:len(fileNameSplit)-1], "/"), fileNameSplit[len(fileNameSplit)-1]
	fileReader, err := cloudManager.Get(filePath, fileName)
	if err != nil {
		logCtx.WithError(err).Errorf("error opening file")
		return err
	}

	defer func() {
		if err := fileReader.Close(); err != nil {
			logCtx.WithError(err).Errorf("error while closing file")
		}
	}()

	eventCount := 0
	scanner := bufio.NewScanner(fileReader)
	for scanner.Scan() {
		eventCount++
		var denormalizedEvent ArchiveEventFormat
		jsonString := scanner.Text()
		if err = U.DecodeJSONStringToStructType(jsonString, &denormalizedEvent); err != nil || denormalizedEvent.UserID == "" {
			logCtx.WithError(err).Errorf("error decoding %s", jsonString)
			continue
		} else if err = updateEventWithParsedFields(&denormalizedEvent); err != nil {
			logCtx.WithError(err).Errorf("error while parsing properties for event %v", denormalizedEvent)
			continue
		} else if denormalizedEvent.EventTimestampUnix < startTime || denormalizedEvent.EventTimestampUnix > endTime {
			continue
		}

		userID := denormalizedEvent.UserID
		(*userEventsMap)[userID] = append((*userEventsMap)[userID], denormalizedEvent)

		var satisfiesGoalEvent bool
		if satisfiesGoalEvent, _ = satisfiesAnyJourneyEvent(denormalizedEvent, segmentAGoals, nil); satisfiesGoalEvent {
			(*segmentAUsers)[denormalizedEvent.UserID] = true
		} else if satisfiesGoalEvent, _ = satisfiesAnyJourneyEvent(denormalizedEvent, segmentBGoals, nil); satisfiesGoalEvent {
			(*segmentBUsers)[denormalizedEvent.UserID] = true
		} else if satisfiesGoalEvent, _ = satisfiesAnyJourneyEvent(denormalizedEvent, segmentCGoals, nil); satisfiesGoalEvent {
			(*segmentCUsers)[denormalizedEvent.UserID] = true
		}
	}

	err = scanner.Err()
	if err != nil {
		return err
	}
	logCtx.Infof("%d events scanned from file", eventCount)
	return nil
}

// getUserEventsFromFile Returns events bucketed by userID.
// TODO(prateek): De-dup based on event id if required.
func getUserEventsFromFile(fileName string, goalEvents []model.QueryEventWithProperties,
	userEventsMap *map[string][]ArchiveEventFormat, userCustomerUserIDMap map[string]string,
	goalFulfilledUsersMap *map[string]bool, journeyEvents []model.QueryEventWithProperties,
	startTime, endTime int64, cloudManager filestore.FileManager) error {

	logCtx := log.WithFields(log.Fields{
		"Method": "getUserEventsFromFile",
	})

	logCtx.Infof("Scanning events from file %s", fileName)
	fileNameSplit := strings.Split(fileName, "/")
	filePath, fileName := strings.Join(fileNameSplit[:len(fileNameSplit)-1], "/"), fileNameSplit[len(fileNameSplit)-1]
	fileReader, err := cloudManager.Get(filePath, fileName)
	if err != nil {
		logCtx.WithError(err).Errorf("error opening file")
		return err
	}

	defer func() {
		if err := fileReader.Close(); err != nil {
			logCtx.WithError(err).Errorf("error while closing file")
		}
	}()

	eventCount := 0
	scanner := bufio.NewScanner(fileReader)
	for scanner.Scan() {
		eventCount++
		var denormalizedEvent ArchiveEventFormat
		jsonString := scanner.Text()
		if err = U.DecodeJSONStringToStructType(jsonString, &denormalizedEvent); err != nil || denormalizedEvent.UserID == "" {
			logCtx.WithError(err).Errorf("error decoding %s", jsonString)
			continue
		} else if err = updateEventWithParsedFields(&denormalizedEvent); err != nil {
			logCtx.WithError(err).Errorf("error while parsing properties for event %v", denormalizedEvent)
			continue
		} else if denormalizedEvent.EventTimestampUnix < startTime || denormalizedEvent.EventTimestampUnix > endTime {
			continue
		} else if satisfies, _ := satisfiesAnyJourneyEvent(denormalizedEvent, journeyEvents, nil); !satisfies {
			// TODO(prateek): Remove this if want to analyze for non completed users as well.
			// Added to save memory on local run.
			continue
		}

		var userID string
		if customerUserID, found := userCustomerUserIDMap[denormalizedEvent.UserID]; found && customerUserID != "" {
			userID = customerUserID
		} else {
			userID = denormalizedEvent.UserID
		}
		(*userEventsMap)[userID] = append((*userEventsMap)[userID], denormalizedEvent)

		if val, found := (*goalFulfilledUsersMap)[userID]; !found || !val {
			if satisfies, _ := satisfiesAnyJourneyEvent(denormalizedEvent, goalEvents, nil); satisfies &&
				(denormalizedEvent.EventTimestampUnix >= startTime && denormalizedEvent.EventTimestampUnix <= endTime) {
				(*goalFulfilledUsersMap)[userID] = true
			} else {
				(*goalFulfilledUsersMap)[userID] = false
			}
		}

	}

	err = scanner.Err()
	if err != nil {
		return err
	}
	logCtx.Infof("%d events scanned from file", eventCount)
	return nil
}

func getUserCustomerUserIDMapFromFile(fileName string, userCustomerUserIDMap *map[string]string, cloudManager filestore.FileManager) error {
	log.Infof("Scanning events from file %s", fileName)
	fileNameSplit := strings.Split(fileName, "/")
	filePath, fileName := strings.Join(fileNameSplit[:len(fileNameSplit)-1], "/"), fileNameSplit[len(fileNameSplit)-1]
	fileReader, err := cloudManager.Get(filePath, fileName)
	if err != nil {
		log.WithError(err).Errorf("Error opening file")
		return err
	}

	defer func() {
		if err := fileReader.Close(); err != nil {
			log.WithError(err).Errorf("Error while closing file")
		}
	}()

	userCount := 0
	scanner := bufio.NewScanner(fileReader)
	for scanner.Scan() {
		userCount++
		var entry model.ArchiveUsersTableFormat
		jsonString := scanner.Text()
		if err = U.DecodeJSONStringToStructType(jsonString, &entry); err != nil || entry.UserID == "" {
			log.WithError(err).Errorf("Error decoding %s", jsonString)
			continue
		}
		(*userCustomerUserIDMap)[entry.UserID] = entry.CustomerUserID
	}

	err = scanner.Err()
	if err != nil {
		return err
	}
	log.Infof("%d users scanned from file", userCount)
	return nil
}

// Compares `value1` wrt to `value2` based on property type and operator.
// `value1` - New value to be compared.
// `value2` - Original value.
// Ex:
//    For operator contains, it will check if value2 contains value1.
//    For operator greaterThan, it will check value1 > value2.
// TODO(prateek): Add tests for this. Move to model.go model.
func areEqualPropertyValues(value1, value2 interface{}, propertyType, operator string) bool {
	if propertyType == U.PropertyTypeCategorical {
		// String comparisons for categorical values.
		switch operator {
		case model.EqualsOpStr:
			return value2.(string) == value1.(string)
		case model.NotEqualOpStr:
			return value2.(string) != value1.(string)
		case "contains":
			return strings.Contains(value1.(string), value2.(string))
		case "notContains":
			return !strings.Contains(value1.(string), value2.(string))
		default:
			return false
		}
	} else if propertyType == U.PropertyTypeNumerical {
		// Float64 comparisons for numerical values.
		switch operator {
		case model.EqualsOpStr:
			return value2.(float64) == value1.(float64)
		case model.NotEqualOpStr:
			return value2.(float64) != value1.(float64)
		case "greaterThan":
			return value1.(float64) > value2.(float64)
		case "lesserThan":
			return value1.(float64) < value2.(float64)
		case "greaterThanOrEqual":
			return value1.(float64) >= value2.(float64)
		case "lesserThanOrEqual":
			return value1.(float64) <= value2.(float64)
		default:
			return false
		}
	}
	return false
}

// areEqualEventsWithFilters Checks if user event satisfies criteria of the event from journey universe.
func areEqualEventsWithFilters(denormalizedEvent ArchiveEventFormat, journeyEvent model.QueryEventWithProperties) bool {
	// Remove any traling / in event name before matching.
	denormalizedEventName := strings.TrimSuffix(denormalizedEvent.EventName, "/")
	if journeyEvent.Name == denormalizedEventName {
		passedAllFilters := true
		for index, property := range journeyEvent.Properties {
			requiredName := property.Property
			var requiredValue interface{}
			requiredValue = property.Value
			if property.Type == U.PropertyTypeNumerical {
				// Numerical properties are also passed as string in event filters.
				// Convert to float before comparing.
				floatValue, err := strconv.ParseFloat(property.Value, 64)
				if err != nil {
					requiredValue = floatValue
				}
			}

			var propertiesToFilterOn map[string]interface{}
			if property.Entity == model.PropertyEntityEvent {
				propertiesToFilterOn = denormalizedEvent.EventPropertiesMap
			} else {
				propertiesToFilterOn = denormalizedEvent.UserPropertiesMap
			}

			eventValue, found := propertiesToFilterOn[requiredName]
			if !found || eventValue == nil || eventValue == "" {
				// If empty, consider as $none before comparison.
				eventValue = model.PropertyValueNone
			}
			passedThisFilter := areEqualPropertyValues(eventValue, requiredValue, property.Type, property.Operator)
			if index >= 0 {
				// LogicalOp for first event property is passed as AND. For reset all apply as passed.
				if property.LogicalOp == "AND" {
					passedAllFilters = passedAllFilters && passedThisFilter
				} else {
					passedAllFilters = passedAllFilters || passedThisFilter
				}
			}
		}
		return passedAllFilters
	}
	return false
}

func satisfiesAnyJourneyEvent(denormalizedEvent ArchiveEventFormat, journeyEvents []model.QueryEventWithProperties,
	visitedEventsMap map[string]bool) (bool, model.QueryEventWithProperties) {
	for _, journeyEvent := range journeyEvents {
		if areEqualEventsWithFilters(denormalizedEvent, journeyEvent) {
			if visitedEventsMap != nil {
				journeyEventString := getEventAsString(journeyEvent)
				if _, alreadyVisited := visitedEventsMap[journeyEventString]; alreadyVisited {
					// This event is already visited. Try matching with other journey events.
					continue
				}
			}
			return true, journeyEvent
		}
	}
	return false, model.QueryEventWithProperties{}
}

// pruneUserEventsPath Makes events for the users ready for path analysis.
//     1. Prune user events list as soon as goal event is achieved. Later events are not of interest.
//     2. Removes any older repeated events from user event list. Ex:
//            E1->E2->E2->E1->E3 will be reduced to E1->E2->E3
//     4. Removes any event which are not part of the journey universe.
//     5. Special event $session, if part of the journey, is allowed to repeat.
func pruneUserEventsPath(userEventsMap *map[string][]ArchiveEventFormat,
	goalEvents []model.QueryEventWithProperties, journeyEvents []model.QueryEventWithProperties,
	eventAliasLegend map[string]string, sessionProperty string) map[string]*JourneyPatternMap {

	allPatternsMap := make(map[string]*JourneyPatternMap)
	logCtx := log.WithFields(log.Fields{"Method": "pruneUserEventsPath"})
	for userID, events := range *userEventsMap {
		if debug {
			logCtx.Infof("%d events for user id %s", len(events), userID)
		}
		debugInfo := debug && debugUserID == userID
		visitedEventsMap := make(map[string]bool) // Map of journey events already visited.
		// visitedSessionEventsMap := make(map[int64]bool) // Map of visited session events to remove duplicates with exact timestamp.
		duplicateAllowedVisitedMap := make(map[string]bool)
		prunedEventsList := make([]ArchiveEventFormat, 0, 0)
		prunedEventsListAliases := []string{}
		lastGoalEventInJourneyIndex := 0
		for _, event := range events {
			if debugInfo {
				logCtx.Infof("Parsing event %s with timestamp %d", event.EventName, event.EventTimestampUnix)
			}
			isPartOfJourneyUniverse, journeyEvent := satisfiesAnyJourneyEvent(
				event, journeyEvents, visitedEventsMap)

			// $session event if present, allowed to be repeated during the journey.
			duplicateAllowed := journeyEvent.Name == U.EVENT_NAME_SESSION
			journeyEventString := getEventAsString(journeyEvent)
			if _, found := visitedEventsMap[journeyEventString]; (found && !duplicateAllowed) || !isPartOfJourneyUniverse {
				// Is already visited once or outside of the journey events universe. Ignore.
				if debugInfo {
					logCtx.Infof("Skipping event %s", event.EventName)
				}
				continue
			}

			// if isSessionEvent {
			// 	if _, found := visitedSessionEventsMap[event.EventTimestampUnix]; found {
			// 		// If exact same event visited before, skip to avoid unwanted duplicates.
			// 		continue
			// 	}
			// 	visitedSessionEventsMap[event.EventTimestampUnix] = true
			// }

			eventAlias := eventAliasLegend[journeyEventString]
			if duplicateAllowed {
				if journeyEvent.Name == U.EVENT_NAME_SESSION {
					value, found := event.EventPropertiesMap[sessionProperty]
					if found && value != nil {
						eventAlias = fmt.Sprintf("%s (%v)", eventAlias, value)
					}
				} else if journeyEvent.Name == "$hubspot_contact_updated" {
					var propSuffix []string
					prop1, found1 := event.EventPropertiesMap["$hubspot_contact_hs_email_last_email_name"]
					if found1 && prop1 != nil {
						propSuffix = append(propSuffix, "email = "+prop1.(string))
					}

					prop1, found1 = event.EventPropertiesMap["$hubspot_contact_lifecyclestage"]
					if found1 && prop1 != nil {
						propSuffix = append(propSuffix, "stage = "+prop1.(string))
					}

					if strings.Join(propSuffix, ", ") != "" {
						eventAlias = fmt.Sprintf("%s (%v)", eventAlias, strings.Join(propSuffix, ", "))
					}
				}

				if _, alreadyVisited := duplicateAllowedVisitedMap[eventAlias]; alreadyVisited {
					continue
				}
				duplicateAllowedVisitedMap[eventAlias] = true
			}
			visitedEventsMap[journeyEventString] = true
			prunedEventsList = append(prunedEventsList, event)
			prunedEventsListAliases = append(prunedEventsListAliases, eventAlias)

			// In case of analyzing users who have not fulfilled goals, it is still safe to check this.
			// They will not have any goal events so this will not break.
			if satisfies, satisfiedGoalEvent := satisfiesAnyJourneyEvent(event, goalEvents, nil); satisfies {
				// TODO(prateek): Special case added for $hubspot_contact_updated analysis. Consider removing.
				eventAlias = eventAliasLegend[getEventAsString(satisfiedGoalEvent)]
				prunedEventsListAliases = append(prunedEventsListAliases, eventAlias)
				lastGoalEventInJourneyIndex = len(prunedEventsListAliases)

				// Goal is reached. Break the parsing here.
				if !allowRepeatedGoals {
					break
				}
			}
		}
		// Update users event list with the pruned list for pattern
		(*userEventsMap)[userID] = prunedEventsList
		if debug {
			logCtx.Infof("Pruned %d events for user %s to %d", len(events), userID, len(prunedEventsList))
		}

		var journeyPath string
		if lastGoalEventInJourneyIndex > 0 {
			journeyPath = strings.Join(prunedEventsListAliases[:lastGoalEventInJourneyIndex], " -> ")
		} else {
			// When analysing for non completed or user as done only goal event from journey,
			// lastGoalEventInJourneyIndex would be zero. Use entire list of events.
			journeyPath = strings.Join(prunedEventsListAliases, " -> ")
		}
		if _, found := allPatternsMap[journeyPath]; found {
			allPatternsMap[journeyPath].Count++
			allPatternsMap[journeyPath].UserIDs = append(allPatternsMap[journeyPath].UserIDs, userID)
		} else {
			allPatternsMap[journeyPath] = &JourneyPatternMap{
				JourneyPath: journeyPath,
				Count:       1,
				UserIDs:     []string{userID},
			}
		}
	}
	return allPatternsMap
}

func dedupJourneyEvents(journeyEvents []model.QueryEventWithProperties) []model.QueryEventWithProperties {
	journeyEventsMap := make(map[string]bool)
	dedupedEventsList := make([]model.QueryEventWithProperties, 0, 0)
	for _, event := range journeyEvents {
		eventString := getEventAsString(event)
		if _, found := journeyEventsMap[eventString]; !found {
			dedupedEventsList = append(dedupedEventsList, event)
		}
	}
	return dedupedEventsList
}

func createEventAliasLegendForJourneyEvents(journeyEvents []model.QueryEventWithProperties) (
	eventAliasLegend, reverseEventAliasLegend map[string]string) {

	eventAliasLegend = make(map[string]string)
	reverseEventAliasLegend = make(map[string]string)
	for index, eventName := range journeyEvents {
		eventKey := getGraphKeyAliasForIndex(index + 1)
		eventAliasLegend[getEventAsString(eventName)] = eventKey
		reverseEventAliasLegend[eventKey] = getEventAsString(eventName)
	}
	return
}

func prettyPrintJourneyPatterns(projectID uint64, allPatternsMap map[string]*JourneyPatternMap,
	outputFileContents string, journeyEvents []model.QueryEventWithProperties,
	eventAliasLegend, reverseEventAliasLegend map[string]string) {

	allPaterns := []string{}
	for pattern := range allPatternsMap {
		allPaterns = append(allPaterns, pattern)
	}

	// Sort by decreasing order of pattern counts.
	sort.Slice(allPaterns, func(i, j int) bool {
		return allPatternsMap[allPaterns[i]].Count > allPatternsMap[allPaterns[j]].Count
	})

	for _, pattern := range allPaterns {
		outputFileContents = fmt.Sprintf("%s\nPattern: %s, Count: %d", outputFileContents, pattern, allPatternsMap[pattern].Count)
	}

	outputFileContents = fmt.Sprintf("%s\n\nWhere", outputFileContents)
	for i := 1; i <= len(journeyEvents); i++ {
		aliasKey := getGraphKeyAliasForIndex(i)
		outputFileContents = fmt.Sprintf("%s\n%s = %s", outputFileContents, aliasKey, reverseEventAliasLegend[aliasKey])
	}
	fmt.Print(outputFileContents)

	// TODO(prateek): Move this to filemanager once it is finalized for production.
	// TODO(prateek): Make bucket configurable.
	storageBucketName := "factors-misc"
	storageFilePath := "JourneyMiningPOC/"
	storageFileName := fmt.Sprintf("%d_%d_journey.txt", projectID, time.Now().UnixNano())
	cloudManager, err := serviceGCS.New(storageBucketName)
	if err != nil {
		log.WithError(err).Error("error creating bucket")
		return
	}
	err = cloudManager.Create(storageFilePath, storageFileName, strings.NewReader(outputFileContents))
	if err != nil {
		log.WithError(err).Error("error while creating journey file")
	}
}

func filterUsersForGoalCompleted(userEventsMap map[string][]ArchiveEventFormat,
	goalFulfilledUsersMap map[string]bool) map[string][]ArchiveEventFormat {
	filteredUserEventsMap := make(map[string][]ArchiveEventFormat)
	for userID, goalFulfilled := range goalFulfilledUsersMap {
		if goalFulfilled {
			filteredUserEventsMap[userID] = userEventsMap[userID]
		}
	}
	return filteredUserEventsMap
}

func filterUsersForGoalNotCompleted(userEventsMap map[string][]ArchiveEventFormat,
	goalFulfilledUsersMap map[string]bool) map[string][]ArchiveEventFormat {
	filteredUserEventsMap := make(map[string][]ArchiveEventFormat)
	for userID, goalFulfilled := range goalFulfilledUsersMap {
		if !goalFulfilled {
			filteredUserEventsMap[userID] = userEventsMap[userID]
		}
	}
	return filteredUserEventsMap
}

// GetWeightedJourneyMatrix Gets user's journey for the given events in given range.
func (pg *Postgres) GetWeightedJourneyMatrix(projectID uint64, journeyEvents []model.QueryEventWithProperties,
	goalEvents []model.QueryEventWithProperties, startTime, endTime, lookbackDays int64, eventFiles,
	userFiles string, includeSession bool, sessionProperty string, cloudManager filestore.FileManager) {
	logCtx := log.WithFields(log.Fields{
		"Method":       "GetWeightedJourneyMatrix",
		"ProjectID":    projectID,
		"Period":       fmt.Sprintf("%d-%d", startTime, endTime),
		"LookbackDays": lookbackDays})
	userEventsMap := make(map[string][]ArchiveEventFormat)
	goalFulfilledUsersMap := make(map[string]bool)

	if includeSession {
		logCtx.Infof("Adding $session event to journey")
		journeyEvents = append(
			[]model.QueryEventWithProperties{model.QueryEventWithProperties{Name: U.EVENT_NAME_SESSION}},
			journeyEvents...)
	}

	// Dedup events just in case user has entered same event twice.
	journeyEvents = dedupJourneyEvents(journeyEvents)
	journeyEvents = append(journeyEvents, goalEvents...)
	eventAliasLegend, reverseEventAliasLegend := createEventAliasLegendForJourneyEvents(journeyEvents)

	userCustomerUserIDMap := make(map[string]string)
	userFileNames := strings.Split(userFiles, ",")
	for _, fileName := range userFileNames {
		getUserCustomerUserIDMapFromFile(fileName, &userCustomerUserIDMap, cloudManager)
	}

	eventFileNames := strings.Split(eventFiles, ",")
	for _, fileName := range eventFileNames {
		fileName = strings.TrimSpace(fileName)
		getUserEventsFromFile(fileName, goalEvents, &userEventsMap, userCustomerUserIDMap,
			&goalFulfilledUsersMap, journeyEvents, startTime, endTime, cloudManager)
		runtime.GC()
		osDebug.FreeOSMemory()
	}
	outputFileContents := fmt.Sprintf("Total number of users: %d", len(userEventsMap))

	// Sort stable events in increasing order of timestamp.
	for userID := range userEventsMap {
		sort.SliceStable(userEventsMap[userID], func(i, j int) bool {
			return userEventsMap[userID][i].EventTimestampUnix < userEventsMap[userID][j].EventTimestampUnix
		})
	}

	goalCompletedUserEventsMap := filterUsersForGoalCompleted(userEventsMap, goalFulfilledUsersMap)
	goalNotCompletedUserEventsMap := filterUsersForGoalNotCompleted(userEventsMap, goalFulfilledUsersMap)

	completedFileContents := fmt.Sprintf("%s\nUsers who reached goal: %d", outputFileContents, len(goalCompletedUserEventsMap))
	allPatternsMap := pruneUserEventsPath(&goalCompletedUserEventsMap, goalEvents, journeyEvents, eventAliasLegend, sessionProperty)
	prettyPrintJourneyPatterns(projectID, allPatternsMap, completedFileContents, journeyEvents,
		eventAliasLegend, reverseEventAliasLegend)

	notCompletedFileContents := fmt.Sprintf("%s\nUsers who did not reach goal: %d", outputFileContents, len(goalNotCompletedUserEventsMap))
	allPatternsMap = pruneUserEventsPath(&goalNotCompletedUserEventsMap, goalEvents, journeyEvents, eventAliasLegend, sessionProperty)
	prettyPrintJourneyPatterns(projectID, allPatternsMap, notCompletedFileContents, journeyEvents,
		eventAliasLegend, reverseEventAliasLegend)
}

func computeInsights5(userEventsMap map[string][]ArchiveEventFormat, segmentUsers map[string]bool) {
	trialSignupEvent := model.QueryEventWithProperties{Name: "www.chargebee.com/trial-signup"}
	scheduleADemoEvent := model.QueryEventWithProperties{Name: "www.chargebee.com/schedule-a-demo"}

	trialSignupBreakDown := make(map[string]map[string]int64)
	trialSignupBreakDown[U.UP_INITIAL_CAMPAIGN] = make(map[string]int64)
	trialSignupBreakDown[U.UP_INITIAL_PAGE_URL] = make(map[string]int64)

	scheduleADemoBreakDown := make(map[string]map[string]int64)
	scheduleADemoBreakDown[U.UP_INITIAL_CAMPAIGN] = make(map[string]int64)
	scheduleADemoBreakDown[U.UP_INITIAL_PAGE_URL] = make(map[string]int64)

	for userID := range segmentUsers {
		latestDenormalizedEvent := userEventsMap[userID][len(userEventsMap[userID])-1]
		initialPageURL := getValueFromProperties(latestDenormalizedEvent.UserPropertiesMap, U.UP_INITIAL_PAGE_URL).(string)
		initialCampaign := getValueFromProperties(latestDenormalizedEvent.UserPropertiesMap, U.UP_INITIAL_CAMPAIGN).(string)

		for i := len(userEventsMap[userID]) - 1; i >= 0; i-- {
			denormalizedEvent := userEventsMap[userID][i]
			if areEqualEventsWithFilters(denormalizedEvent, trialSignupEvent) {
				trialSignupBreakDown[U.UP_INITIAL_CAMPAIGN][initialCampaign]++
				trialSignupBreakDown[U.UP_INITIAL_PAGE_URL][initialPageURL]++
				break
			} else if areEqualEventsWithFilters(denormalizedEvent, scheduleADemoEvent) {
				scheduleADemoBreakDown[U.UP_INITIAL_CAMPAIGN][initialCampaign]++
				scheduleADemoBreakDown[U.UP_INITIAL_PAGE_URL][initialPageURL]++
				break
			}
		}
	}

	for property, propertyCountsMap := range trialSignupBreakDown {
		fmt.Printf("\t\t%s : %s\n", trialSignupEvent.Name, property)
		allBreakDowns := []string{}
		for breakDown := range propertyCountsMap {
			allBreakDowns = append(allBreakDowns, breakDown)
		}
		sort.Slice(allBreakDowns, func(i, j int) bool {
			return propertyCountsMap[allBreakDowns[i]] > propertyCountsMap[allBreakDowns[j]]
		})

		for _, breakdown := range allBreakDowns {
			fmt.Printf("\t\t\t%s: %d\n", breakdown, propertyCountsMap[breakdown])
		}
	}

	for property, propertyCountsMap := range scheduleADemoBreakDown {
		fmt.Printf("\t\t%s : %s\n", scheduleADemoEvent.Name, property)
		allBreakDowns := []string{}
		for breakDown := range propertyCountsMap {
			allBreakDowns = append(allBreakDowns, breakDown)
		}
		sort.Slice(allBreakDowns, func(i, j int) bool {
			return propertyCountsMap[allBreakDowns[i]] > propertyCountsMap[allBreakDowns[j]]
		})

		for _, breakdown := range allBreakDowns {
			fmt.Printf("\t\t\t%s: %d\n", breakdown, propertyCountsMap[breakdown])
		}
	}
}

func computeInsights4(userEventsMap map[string][]ArchiveEventFormat, segmentUsers map[string]bool,
	segmentVisitedPages []model.QueryEventWithProperties) {

	totalSegmentUsers := len(segmentUsers)

	visitedPageUsersMap := make(map[string]map[string]bool)
	visitedPageURLs := make([]string, 0, 0)
	for _, pageEvent := range segmentVisitedPages {
		pageEventString := pageEvent.Name
		visitedPageUsersMap[pageEventString] = make(map[string]bool)
		visitedPageURLs = append(visitedPageURLs, pageEventString)
	}

	for userID := range segmentUsers {
		for _, denormalizedEvent := range userEventsMap[userID] {
			for _, pageEvent := range segmentVisitedPages {
				if areEqualEventsWithFilters(denormalizedEvent, pageEvent) {
					pageEventString := pageEvent.Name
					visitedPageUsersMap[pageEventString][userID] = true
				}
			}
		}
	}

	sort.Slice(visitedPageURLs, func(i, j int) bool {
		return len(visitedPageUsersMap[visitedPageURLs[i]]) > len(visitedPageUsersMap[visitedPageURLs[j]])
	})

	for _, pageEventString := range visitedPageURLs {
		pageEventCount := len(visitedPageUsersMap[pageEventString])
		usersPercentage := float64(pageEventCount) / float64(totalSegmentUsers) * 100
		fmt.Printf("\t\t%s: %d / %d, %0.2f%%\n", pageEventString, pageEventCount, totalSegmentUsers, usersPercentage)
	}
}

func computeInsights3(userEventsMap map[string][]ArchiveEventFormat, segmentUsers map[string]bool) {
	sessionsByCampaignInitialPageURL := make(map[string]map[string]int64)
	sessionEvent := model.QueryEventWithProperties{Name: U.EVENT_NAME_SESSION}

	for userID := range segmentUsers {
		var initialPageURL interface{}
		latestDenormalizedEvent := userEventsMap[userID][len(userEventsMap[userID])-1]
		initialPageURL = getValueFromProperties(latestDenormalizedEvent.UserPropertiesMap, U.UP_INITIAL_PAGE_URL)
		for _, denormalizedEvent := range userEventsMap[userID] {
			if areEqualEventsWithFilters(denormalizedEvent, sessionEvent) {
				campaign := getValueFromProperties(denormalizedEvent.EventPropertiesMap, U.EP_CAMPAIGN).(string)
				if _, found := sessionsByCampaignInitialPageURL[campaign]; !found {
					sessionsByCampaignInitialPageURL[campaign] = make(map[string]int64)
				}
				sessionsByCampaignInitialPageURL[campaign][initialPageURL.(string)]++
			}
		}
	}

	allCampaigns := []string{}
	for campaign := range sessionsByCampaignInitialPageURL {
		allCampaigns = append(allCampaigns, campaign)
	}
	sort.Slice(allCampaigns, func(i, j int) bool {
		var iCampaignCounts, jCampaignCounts int64
		for _, count := range sessionsByCampaignInitialPageURL[allCampaigns[i]] {
			iCampaignCounts += count
		}
		for _, count := range sessionsByCampaignInitialPageURL[allCampaigns[j]] {
			jCampaignCounts += count
		}
		return iCampaignCounts > jCampaignCounts
	})

	for _, campaign := range allCampaigns {
		fmt.Printf("\t\tCampaign: %s\n", campaign)
		initialPageURLMap := sessionsByCampaignInitialPageURL[campaign]
		allURLs := []string{}
		for URL := range initialPageURLMap {
			allURLs = append(allURLs, URL)
		}
		sort.Slice(allURLs, func(i, j int) bool {
			return initialPageURLMap[allURLs[i]] > initialPageURLMap[allURLs[j]]
		})
		for _, URL := range allURLs {
			fmt.Printf("\t\t\t%s: %d\n", URL, initialPageURLMap[URL])
		}
	}
}

func computeInsights2(userEventsMap map[string][]ArchiveEventFormat, segmentUsers map[string]bool) {
	latestPageURLMap := make(map[string]int64)
	latestPageURLs := make([]string, 0, 0)

	for userID := range segmentUsers {
		latestDenormalizedEvent := userEventsMap[userID][len(userEventsMap[userID])-1]
		userPorperties := latestDenormalizedEvent.UserPropertiesMap
		pageURL, found := userPorperties[U.UP_LATEST_PAGE_URL]
		if !found || pageURL == nil || pageURL.(string) == "" {
			pageURL = model.PropertyValueNone
		}
		if _, found := latestPageURLMap[pageURL.(string)]; !found {
			latestPageURLs = append(latestPageURLs, pageURL.(string))
		}
		latestPageURLMap[pageURL.(string)]++
	}

	// Sort in decreasing order of counts.
	sort.Slice(latestPageURLs, func(i, j int) bool {
		return latestPageURLMap[latestPageURLs[i]] > latestPageURLMap[latestPageURLs[j]]
	})

	for _, pageURL := range latestPageURLs {
		fmt.Printf("\t\t%s: %d\n", pageURL, latestPageURLMap[pageURL])
	}
}

func computeInsights1(userEventsMap map[string][]ArchiveEventFormat, segmentUsers map[string]bool) {
	propertyCounts := make(map[string][]float64)
	noneCounts := make(map[string]int64)
	properties := []string{U.UP_PAGE_COUNT, U.UP_TOTAL_SPENT_TIME}
	sessionEvent := model.QueryEventWithProperties{Name: U.EVENT_NAME_SESSION}

	for _, property := range properties {
		propertyCounts[property] = []float64{}
		noneCounts[property] = 0
	}

	for userID := range segmentUsers {
		userCounts := map[string]float64{
			U.UP_PAGE_COUNT:       0,
			U.UP_TOTAL_SPENT_TIME: 0,
		}
		// Alternative logic to compute page_count and session_spent_time from $session events.
		// for _, denormalizedEvent := range userEventsMap[userID] {
		// 	if areEqualEventsWithFilters(denormalizedEvent, sessionEvent) {
		// 		userCounts[U.UP_SESSION_COUNT]++
		// 		eventProperties := denormalizedEvent.EventProperties
		// 		log.Infof("%s: This is session event ...", userID)

		// 		log.Infof("%s: %v", userID, eventProperties[U.SP_PAGE_COUNT])
		// 		if eventPageCount, found := eventProperties[U.SP_PAGE_COUNT]; found &&
		// 			eventPageCount != nil && eventPageCount.(float64) != 0 {
		// 			userCounts[U.UP_PAGE_COUNT] += eventPageCount.(float64)
		// 		}

		// 		log.Infof("%s: %v", userID, eventProperties[U.SP_SPENT_TIME])
		// 		if eventSessionSpentTime, found := eventProperties[U.SP_SPENT_TIME]; found &&
		// 			eventSessionSpentTime != nil && eventSessionSpentTime.(float64) != 0 {
		// 			userCounts[U.UP_TOTAL_SPENT_TIME] += eventSessionSpentTime.(float64)
		// 		}
		// 	}
		// }
		for _, denormalizedEvent := range userEventsMap[userID] {
			if strings.HasPrefix(denormalizedEvent.EventName, "www.") {
				userCounts[U.UP_PAGE_COUNT]++
				eventProperties := denormalizedEvent.EventPropertiesMap

				if eventSessionSpentTime, found := eventProperties[U.EP_PAGE_SPENT_TIME]; found &&
					eventSessionSpentTime != nil && eventSessionSpentTime.(float64) != 0 {
					userCounts[U.UP_TOTAL_SPENT_TIME] += eventSessionSpentTime.(float64)
				}
			}
		}
		for _, property := range properties {
			if userCounts[property] == 0 {
				noneCounts[property]++
			} else {
				propertyCounts[property] = append(propertyCounts[property], float64(int64(userCounts[property])))
			}
		}
	}

	for _, property := range properties {
		// Sort before calculating median, 75th and 95th percentile.
		counts := propertyCounts[property]
		sort.Slice(counts, func(i, j int) bool {
			return counts[i] < counts[j]
		})

		var median, percentile70, percentile90 float64
		if len(counts) > 0 {
			mid := len(counts) / 2
			if len(counts)%2 == 0 {
				median = (counts[mid] + counts[mid-1]) / 2
			} else {
				median = counts[mid]
			}
			percentile70 = counts[int(float64(len(counts))*0.7)]
			percentile90 = counts[int(float64(len(counts))*0.9)]
		}
		fmt.Printf("\t\t%s, $none: %d, Median: %d, 70th percentile: %d, 90th percentile: %d\n",
			property, noneCounts[property], int64(median), int64(percentile70), int64(percentile90))
	}
}

// ChargebeeAnalysis To run analysis for chargebee as per the requirements on issue #515.
func ChargebeeAnalysis(segmentAGoals, segmentBGoals, segmentCGoals,
	segmentVisitedPages []model.QueryEventWithProperties, sourceFileNames string,
	startTime, endTime int64, cloudManager filestore.FileManager) {

	userEventsMap := make(map[string][]ArchiveEventFormat)

	segmentAUsers := make(map[string]bool)
	segmentBUsers := make(map[string]bool)
	segmentCUsers := make(map[string]bool)

	fileNames := strings.Split(sourceFileNames, ",")
	for _, fileName := range fileNames {
		fileName = strings.TrimSpace(fileName)
		getUserEventsFromFileChargebee(fileName, segmentAGoals, segmentBGoals, segmentCGoals, &userEventsMap,
			&segmentAUsers, &segmentBUsers, &segmentCUsers, []model.QueryEventWithProperties{},
			startTime, endTime, cloudManager)
		runtime.GC()
		osDebug.FreeOSMemory()
	}

	// Sort stable events in increasing order of timestamp.
	for userID := range userEventsMap {
		sort.SliceStable(userEventsMap[userID], func(i, j int) bool {
			return userEventsMap[userID][i].EventTimestampUnix < userEventsMap[userID][j].EventTimestampUnix
		})
	}
	fmt.Printf("Total number of users: %d\n", len(userEventsMap))
	originalSegmentCUsersCounts := len(segmentCUsers)

	for userID := range segmentCUsers {
		_, foundInA := segmentAUsers[userID]
		_, foundInB := segmentBUsers[userID]
		if foundInA || foundInB {
			delete(segmentCUsers, userID)
		}
	}

	fmt.Printf("Segment A: %v\n", segmentAGoals)
	fmt.Printf("Total number of users in segment A: %d\n", len(segmentAUsers))
	fmt.Printf("\tInsights 1\n")
	computeInsights1(userEventsMap, segmentAUsers)
	fmt.Printf("\tInsights 2\n")
	computeInsights2(userEventsMap, segmentAUsers)
	fmt.Printf("\tInsights 3\n")
	computeInsights3(userEventsMap, segmentAUsers)
	fmt.Printf("\tInsights 4\n")
	computeInsights4(userEventsMap, segmentAUsers, segmentVisitedPages)

	fmt.Printf("Segment B: %v\n", segmentBGoals)
	fmt.Printf("Total number of users in segment B: %d\n", len(segmentBUsers))
	fmt.Printf("\tInsights 1\n")
	computeInsights1(userEventsMap, segmentBUsers)
	fmt.Printf("\tInsights 2\n")
	computeInsights2(userEventsMap, segmentBUsers)
	fmt.Printf("\tInsights 3\n")
	computeInsights3(userEventsMap, segmentBUsers)
	fmt.Printf("\tInsights 4\n")
	computeInsights4(userEventsMap, segmentBUsers, segmentVisitedPages)

	fmt.Printf("Segment C: %v\n", segmentCGoals)
	fmt.Printf("Total number of users in segment C: %d\n", originalSegmentCUsersCounts)
	fmt.Printf("Total number of users in segment C after filtering A and B: %d\n", len(segmentCUsers))
	fmt.Printf("\tInsights 1\n")
	computeInsights1(userEventsMap, segmentCUsers)
	fmt.Printf("\tInsights 2\n")
	computeInsights2(userEventsMap, segmentCUsers)
	fmt.Printf("\tInsights 3\n")
	computeInsights3(userEventsMap, segmentCUsers)
	fmt.Printf("\tInsights 4\n")
	computeInsights4(userEventsMap, segmentCUsers, segmentVisitedPages)
	fmt.Printf("\tInsights 5\n")
	computeInsights5(userEventsMap, segmentCUsers)
}
