package delta

import (
	"bufio"
	"encoding/json"
	"factors/filestore"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	T "factors/task"
	U "factors/util"
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
)

var SESSION_EVENT_NAME = "$session"
var OTHERS_VALUE string = "$others"
var deltaLog = log.New()
var deltaComputeLog = deltaLog.WithField("prefix", "Delta#Compute")

func updateSessions(preSessionEvents *[]P.CounterEventFormat, sessions *[]Session,
	sessionId int, event P.CounterEventFormat) int {
	if event.EventName == SESSION_EVENT_NAME {
		sessionId++
	}
	if sessionId == -1 {
		*preSessionEvents = append(*preSessionEvents, event)
		return sessionId
	}
	updateSessionsList(sessions, sessionId, event)
	return sessionId
}

func updateSessionsList(sessions *[]Session, sessionId int, event P.CounterEventFormat) {
	if len(*sessions) > sessionId { // Session exists.
		(*sessions)[sessionId].Events = append((*sessions)[sessionId].Events, event)
	} else {
		var events []P.CounterEventFormat
		events = append(events, event)
		session := Session{Events: events}
		(*sessions) = append((*sessions), session)
	}
	// if _, ok := sessions[sessionId]; !ok {
	// 	sessions[sessionId] = []P.CounterEventFormat{}
	// }
	// sessions[sessionId] = append(sessions[sessionId], event)
}

var USER_PROPERTIES_MODE PropertiesMode = "user"
var EVENT_PROPERTIES_MODE PropertiesMode = "event"

func eventMatchesFilterCriterion(event P.CounterEventFormat, filterCriterion EventFilterCriterion) bool {
	filterKey := filterCriterion.Key
	filterValues := filterCriterion.ValueSet
	containmentFlag := filterCriterion.EqualityFlag
	mode := filterCriterion.PropertiesMode
	var props map[string]interface{}
	if mode == USER_PROPERTIES_MODE {
		props = event.UserProperties
	} else if mode == EVENT_PROPERTIES_MODE {
		props = event.EventProperties
	}
	if _, ok := props[filterKey]; !ok {
		fmt.Println("Error: Filter key not found in properties. Defaulting match flag as false.")
		return false
	}
	propertyValue := props[filterKey]
	for _, filterValue := range filterValues {
		// "OR" logic for containment: If there's even a single match when asked for containment, return True.
		// "AND" logic for non-containment: If there's even a single match when asked for non-containment, return False.
		if filterValue == propertyValue {
			return containmentFlag
		}
	}
	// Finally, for containment, if there's no match in the loop above, return False (!containmentFlag)
	// For non-containment, if all are mismatches (in the loop above), return True (!containmentFlag)
	return !containmentFlag
}

func eventMatchesFilterCriterionList(event P.CounterEventFormat, filterCriterionList []EventFilterCriterion) bool {
	for _, fc := range filterCriterionList {
		if !eventMatchesFilterCriterion(event, fc) { // "AND" logic: If even a single filter fails, return False.
			return false
		}
	}
	return true
}

func eventMatchesCriterion(event P.CounterEventFormat, eventCriterion EventCriterion) bool {
	// TODO: Match event filters as well.
	nameMatchFlag := eventCriterion.EqualityFlag == (event.EventName == eventCriterion.Name)
	if !nameMatchFlag {
		return false
	}
	filterMatchFlag := eventMatchesFilterCriterionList(event, eventCriterion.FilterCriterionList)
	return filterMatchFlag
}

func updateCriteriaResult(event P.CounterEventFormat, criteria EventsCriteria, criteriaResult *PerUserCriteriaResult) {
	if criteriaResult.criteriaMatchFlag { // If criteria already met, do nothing.
		return
	}
	if len(criteria.EventCriterionList) == 0 { // If criteria has no event criterion to match
		log.Info("Criteria has empty criterion list. By default, making the user match the first event.")
		(*criteriaResult).anyFlag = true
		(*criteriaResult).allFlag = true
		(*criteriaResult).criteriaMatchFlag = true
		(*criteriaResult).mostRecentEvent = event
		return
	}
	for i, eventCriterion := range criteria.EventCriterionList {
		if (*criteriaResult).criterionResultList[i].matchId != -1 {
			// This criterion has already been matched. So make no more changes.
			// TODO: For now, we are going with "first occurrence" model. Might need to update.
			continue
		}
		if eventMatchesCriterion(event, eventCriterion) {
			(*criteriaResult).criterionResultList[i].matchId = (*criteriaResult).numCriterionMatched
			(*criteriaResult).numCriterionMatched++
			(*criteriaResult).anyFlag = true
			(*criteriaResult).mostRecentEvent = event
			if (*criteriaResult).numCriterionMatched == len(criteria.EventCriterionList) {
				(*criteriaResult).allFlag = true
			}
			if ((criteria.Operator == "And") && (*criteriaResult).allFlag) ||
				((criteria.Operator == "Or") && (*criteriaResult).anyFlag) {
				(*criteriaResult).criteriaMatchFlag = true
				break
			}
		}
	}
}

func QueryEvent(event P.CounterEventFormat, deltaQuery Query, perUserQueryResult *PerUserQueryResult) {
	updateCriteriaResult(event, deltaQuery.Base, &(*perUserQueryResult).baseResult)
	updateCriteriaResult(event, deltaQuery.Target, &(*perUserQueryResult).targetResult)
}

func QuerySession(session Session, deltaQuery Query, perUserQueryResult *PerUserQueryResult) {
	for _, event := range session.Events {
		QueryEvent(event, deltaQuery, perUserQueryResult)
		baseFlag := (*perUserQueryResult).baseResult.criteriaMatchFlag
		targetFlag := (*perUserQueryResult).targetResult.criteriaMatchFlag
		if !baseFlag && targetFlag {
			log.Info("WARNING: Target matched before base!")
		} else if baseFlag && targetFlag {
			break
		}
	}
}

func makeCriteriaResult(criteria EventsCriteria) PerUserCriteriaResult {
	var criterionResultList []CriterionResult
	for range criteria.EventCriterionList {
		result := CriterionResult{matchId: -1}
		criterionResultList = append(criterionResultList, result)
	}
	criteriaResult := PerUserCriteriaResult{criterionResultList: criterionResultList,
		allFlag: false, anyFlag: false, mostRecentEvent: P.CounterEventFormat{},
		numCriterionMatched: 0, criteriaMatchFlag: false}
	return criteriaResult
}

func makePerUserQueryResult(deltaQuery Query) PerUserQueryResult {
	baseCriteriaResult := makeCriteriaResult(deltaQuery.Base)
	targetCriteriaResult := makeCriteriaResult(deltaQuery.Target)
	perUserQueryResult := PerUserQueryResult{baseResult: baseCriteriaResult, targetResult: targetCriteriaResult}
	return perUserQueryResult
}

type Session struct {
	Events []P.CounterEventFormat
}

func QueryUser(preSessionEvents []P.CounterEventFormat, sessions []Session, deltaQuery Query) (PerUserQueryResultSummary, error) {
	var err error = nil
	userResult := makePerUserQueryResult(deltaQuery)
	var extendedSessions []Session // Extended sessions are preSessionEvents + sessions.
	extendedSessions = append(extendedSessions, Session{Events: preSessionEvents})
	extendedSessions = append(extendedSessions, sessions...)
	for _, session := range extendedSessions {
		QuerySession(session, deltaQuery, &userResult)
		if userResult.baseResult.criteriaMatchFlag && userResult.targetResult.criteriaMatchFlag {
			break
		}
	}
	summary := PerUserQueryResultSummary{BaseFlag: userResult.baseResult.criteriaMatchFlag,
		TargetFlag: userResult.targetResult.criteriaMatchFlag, ChosenEvent: P.CounterEventFormat{}}
	summary.BaseAndTargetFlag = summary.BaseFlag && summary.TargetFlag
	if summary.BaseFlag {
		if summary.TargetFlag {
			// If user hits both base and target, then choose the target event (properties) for the analysis.
			summary.ChosenEvent = userResult.targetResult.mostRecentEvent
		} else {
			// If user hits base but misses target, then choose base event (properties) for the analysis.
			summary.ChosenEvent = userResult.baseResult.mostRecentEvent
		}
	} else {
		if summary.TargetFlag {
			summary.ChosenEvent = userResult.targetResult.mostRecentEvent
		}
	}
	return summary, err
}

func getEventMatchFlag(event P.CounterEventFormat, eventCriterion EventCriterion) bool {
	matchFlag := (event.EventName == eventCriterion.Name)
	return matchFlag
}

func getCustomBlacklist() map[string]bool {
	// TODO: This was changed from set to map
	customBlacklistMap := make(map[string]bool)
	for _, featKey := range U.CUSTOM_BLACKLIST_DELTA {
		customBlacklistMap[featKey] = true
	}
	return customBlacklistMap
}

func ComputeWithinPeriodInsights(scanner *bufio.Scanner, deltaQuery Query, k int, featSoftWhitelist map[string]map[string]bool, passId int) (WithinPeriodInsights, error) {
	var err error
	var wpInsights WithinPeriodInsights
	var prevUserId string = ""
	var matchSummary PerUserQueryResultSummary
	var matchedBaseEvents []P.CounterEventFormat
	var matchedTargetEvents []P.CounterEventFormat
	var matchedBaseAndTargetEvents []P.CounterEventFormat
	var sessions []Session
	var preSessionEvents []P.CounterEventFormat = nil
	customBlacklist := getCustomBlacklist()
	sessionId := -1
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum%1000 == 0 {
			fmt.Printf("%d lines scanned\n", lineNum)
		}
		line := scanner.Text()
		var event P.CounterEventFormat
		json.Unmarshal([]byte(line), &event) // TODO: Add error check.
		filterBlacklist(&event, &customBlacklist)
		sanitizeScreenSize(&event)
		if prevUserId == "" {
			prevUserId = event.UserId
		}
		if event.UserId == prevUserId { // If the same user's events are continuing...
			sessionId = updateSessions(&preSessionEvents, &sessions, sessionId, event)
			if sessionId == -1 {
				continue
			}
		} else { // If a new user's events have started coming...
			matchSummary, err = QueryUser(preSessionEvents, sessions, deltaQuery)
			if err != nil {
				return wpInsights, err
			}
			if matchSummary.BaseFlag {
				matchedBaseEvents = append(matchedBaseEvents, matchSummary.ChosenEvent)
			}
			if matchSummary.TargetFlag {
				matchedTargetEvents = append(matchedTargetEvents, matchSummary.ChosenEvent)
			}
			if matchSummary.BaseAndTargetFlag {
				matchedBaseAndTargetEvents = append(matchedBaseAndTargetEvents, matchSummary.ChosenEvent)
			}
			preSessionEvents = nil
			sessions = nil
			sessionId = -1
			sessionId = updateSessions(&preSessionEvents, &sessions, sessionId, event)
			prevUserId = event.UserId
		}
	}
	wpInsights = translateToWPMetrics(matchedBaseEvents, matchedTargetEvents,
		matchedBaseAndTargetEvents, featSoftWhitelist, passId)
	if passId == 1 {
		prefilterFeatures(&wpInsights)
		selectTopKFeatures(&(wpInsights), k)
	}
	return wpInsights, err
}

// RFDParams stands for Parameters of a Rank-Frequency Distribution
type RFDParams struct {
	numNonNullUniq              int
	numNonNullMultiOccurrence   int
	percNonNullSingleOccurrence float32
	numNonNullSingleOccurrence  int
	totalCount                  int
}

// Analyzes the RFD params of a feature and decides whether or not to retain it.
func analyzeRFDParams(rfd RFDParams) bool {
	if rfd.numNonNullUniq == 1 { // If feature has only one non-null value, reject
		return false
	}
	if rfd.numNonNullMultiOccurrence == 0 { //If feature has no value with > 1 frequency, reject
		return false
	}
	if rfd.numNonNullMultiOccurrence == 1 {
		// If only one value occurs with > 1 frequency, and (of course) others with 1 frequency, reject
		return false
	}
	return true
}

func countRFD(valStats Level2CatRatioDist) RFDParams {
	var numVals int = 0
	var numSingleFreqVals int = 0
	var numMultiFreqVals int = 0
	var numZeroFreqVals int = 0
	var totalCount int = 0
	for _, stats := range valStats {
		count := int(stats["#users"])
		numVals++
		totalCount += count
		if count == 1 {
			numSingleFreqVals++
		} else if count == 0 {
			numZeroFreqVals++
		} else if count > 1 {
			numMultiFreqVals++
		}
	}
	numNonZeroFreqVals := numVals - numZeroFreqVals
	rfd := RFDParams{numNonNullUniq: numNonZeroFreqVals,
		numNonNullMultiOccurrence:  numMultiFreqVals,
		numNonNullSingleOccurrence: numSingleFreqVals,
		totalCount:                 totalCount}
	return rfd
}

func analyzeFeature(key string, valStats Level2CatRatioDist) (bool, RFDParams) {
	rfd := countRFD(valStats)
	numericFlag := IsANumericFeature(key, valStats)
	return numericFlag, rfd
}

func prefilterFeatures(wpInsights *WithinPeriodInsights) {
	prefilterFeaturesProperties(&((*wpInsights).Base.FeatureMetrics))
	prefilterFeaturesProperties(&((*wpInsights).Target.FeatureMetrics))
	prefilterFeaturesProperties(&((*wpInsights).BaseAndTarget.FeatureMetrics))
}

func prefilterFeaturesProperties(featureMetrics *Level3CatRatioDist) {
	for key, valStats := range *featureMetrics {
		keyNumericFlag, keyRFDParams := analyzeFeature(key, valStats)
		if keyNumericFlag || !analyzeRFDParams(keyRFDParams) {
			delete((*featureMetrics), key)
		}
	}
}

func filterBlacklist(event *P.CounterEventFormat, customBlacklist *map[string]bool) {
	filterPropertiesBlacklist(&((*event).EventProperties), customBlacklist)
	filterPropertiesBlacklist(&((*event).UserProperties), customBlacklist)
}

func filterPropertiesBlacklist(properties *map[string]interface{}, customBlacklist *map[string]bool) {
	for key := range *properties {
		if (*customBlacklist)[key] == true {
			delete((*properties), key)
		}
	}
}

func sanitizeScreenSize(event *P.CounterEventFormat) {
	sanitizeScreenSizeProperties(&((*event).EventProperties))
	sanitizeScreenSizeProperties(&((*event).UserProperties))
}

func sanitizeScreenSizeProperties(properties *(map[string]interface{})) {
	if sw, ok1 := (*properties)["$screen_width"]; ok1 {
		if sh, ok2 := (*properties)["$screen_height"]; ok2 {
			(*properties)["$screen_size"] = fmt.Sprint(sw) + "x" + fmt.Sprint(sh)
			delete((*properties), "$screen_width")
			delete((*properties), "$screen_height")
		}
	}
}

func eventsToMetrics(events []P.CounterEventFormat, featSoftWhitelist map[string]map[string]bool, passId int) WithinPeriodMetrics {
	globalMetrics := make(Level1CatFreqDist)
	featureMetrics := make(Level3CatRatioDist)
	for _, event := range events {
		updateLevel1CatFreqDist(&globalMetrics, "#users", 1)
		updateMetricsWithProperties(&featureMetrics, event.UserProperties, "up#", featSoftWhitelist, passId)
		updateMetricsWithProperties(&featureMetrics, event.EventProperties, "ep#", featSoftWhitelist, passId)
	}
	populatePrevalence(&featureMetrics, float64(globalMetrics["#users"]))
	metrics := WithinPeriodMetrics{GlobalMetrics: globalMetrics, FeatureMetrics: featureMetrics}
	return metrics
}

func populatePrevalence(featureMetrics *Level3CatRatioDist, totalUserCount float64) {
	for key, valStats := range *featureMetrics {
		for val, stats := range valStats {
			featureUserCount := stats["#users"]
			prev := SmartDivide(featureUserCount, totalUserCount)
			(*featureMetrics)[key][val]["prev"] = prev
		}
	}
}

func updateMetricsWithProperties(featureMetrics *Level3CatRatioDist, properties map[string]interface{}, keyPrefix string, featSoftWhitelist map[string]map[string]bool, passId int) {
	for key, val := range properties {
		// TODO: As of now, we are treating all (user/event) properties as categorical variables.
		// Need to support numerical ones.
		keyStr := keyPrefix + key
		valStr := fmt.Sprintf("%v", val)
		if passId == 2 { // In passId 1, we would have featSoftWhitelist as empty.
			if featSoftWhitelist[keyStr] != nil && featSoftWhitelist[keyStr][valStr] == true {
				continue
			}
		}
		// This increments the frequency of #users by 1.
		updateLevel3CatRatioDist(featureMetrics, keyStr, valStr, "#users", 1.0)
	}
}

func updateLevel1CatFreqDist(freqDist *Level1CatFreqDist, key string, count int) {
	if _, ok := (*freqDist)[key]; !ok {
		(*freqDist)[key] = count
	} else {
		(*freqDist)[key] += count
	}
}

func updateLevel2CatFreqDist(freqDist *Level2CatFreqDist, key1, key2 string, count int) {
	if _, ok := (*freqDist)[key1]; !ok {
		(*freqDist)[key1] = make(Level1CatFreqDist)
	}
	level1FreqDist := (*freqDist)[key1]
	updateLevel1CatFreqDist(&level1FreqDist, key2, count)
	(*freqDist)[key1] = level1FreqDist
}

func updateLevel3CatFreqDist(freqDist *Level3CatFreqDist, key1, key2, key3 string, count int) {
	if _, ok := (*freqDist)[key1]; !ok {
		(*freqDist)[key1] = make(Level2CatFreqDist)
	}
	level2FreqDist := (*freqDist)[key1]
	updateLevel2CatFreqDist(&level2FreqDist, key2, key3, count)
	(*freqDist)[key1] = level2FreqDist
}

func updateLevel1CatRatioDist(freqDist *Level1CatRatioDist, key string, ratio float64) {
	if _, ok := (*freqDist)[key]; !ok {
		(*freqDist)[key] = ratio
	} else {
		(*freqDist)[key] += ratio
	}
}

func updateLevel2CatRatioDist(freqDist *Level2CatRatioDist, key1, key2 string, ratio float64) {
	if _, ok := (*freqDist)[key1]; !ok {
		(*freqDist)[key1] = make(Level1CatRatioDist)
	}
	level1RatioDist := (*freqDist)[key1]
	updateLevel1CatRatioDist(&level1RatioDist, key2, ratio)
	(*freqDist)[key1] = level1RatioDist
}

func updateLevel3CatRatioDist(freqDist *Level3CatRatioDist, key1, key2, key3 string, ratio float64) {
	if _, ok := (*freqDist)[key1]; !ok {
		(*freqDist)[key1] = make(Level2CatRatioDist)
	}
	level2RatioDist := (*freqDist)[key1]
	updateLevel2CatRatioDist(&level2RatioDist, key2, key3, ratio)
	(*freqDist)[key1] = level2RatioDist
}

func translateToWPMetrics(baseEvents, targetEvents, baseAndTargetEvents []P.CounterEventFormat, featSoftWhitelist map[string]map[string]bool, passId int) WithinPeriodInsights {
	// TODO: Assert that in baseEvents, each user has only one row, not more. Same for targetEvents.
	baseMetrics := eventsToMetrics(baseEvents, featSoftWhitelist, passId)
	targetMetrics := eventsToMetrics(targetEvents, featSoftWhitelist, passId)
	baseAndTargetMetrics := eventsToMetrics(baseAndTargetEvents, featSoftWhitelist, passId)
	convMetrics := WithinPeriodRatioMetrics{}
	if passId == 2 {
		convMetrics = computeConversionMetrics(baseMetrics, targetMetrics, baseAndTargetMetrics)
	}
	wpInsights := WithinPeriodInsights{Base: baseMetrics, Target: targetMetrics,
		BaseAndTarget: baseAndTargetMetrics, Conversion: convMetrics}
	return wpInsights
}

func computeConversionMetrics(baseMetrics, targetMetrics, baseAndTargetMetrics WithinPeriodMetrics) WithinPeriodRatioMetrics {
	globalConvRatio := SmartDivide(float64(baseAndTargetMetrics.GlobalMetrics["#users"]), float64(baseMetrics.GlobalMetrics["#users"]))
	globalMetrics := Level1CatRatioDist{"ratio": globalConvRatio}
	featureMetrics := make(Level2CatRatioDist)
	for key, valBaseCounts := range baseMetrics.FeatureMetrics {
		if _, ok := baseAndTargetMetrics.FeatureMetrics[key]; !ok {
			continue
		}
		featureMetrics[key] = make(Level1CatRatioDist)
		valBaseAndTargetCounts := baseAndTargetMetrics.FeatureMetrics[key]
		for val, baseStats := range valBaseCounts {
			baseCount := baseStats["#users"]
			if _, ok := valBaseAndTargetCounts[val]; !ok {
				continue
			}
			baseAndTargetCount := valBaseAndTargetCounts[val]["#users"]
			convRate := SmartDivide(float64(baseAndTargetCount), float64(baseCount))
			featureMetrics[key][val] = convRate
		}
	}
	convMetrics := WithinPeriodRatioMetrics{GlobalMetrics: globalMetrics, FeatureMetrics: featureMetrics}
	return convMetrics
}

func findTopKCountThres(featureMetrics Level3CatRatioDist, k int) int {
	var counts []int = nil
	var tempCounts []int = nil
	for _, valStats := range featureMetrics {
		for _, stats := range valStats {
			count := int(stats["#users"])
			counts = append(counts, count)
			tempCounts = append(tempCounts, count)
		}
	}
	kthMaxCount := findKthMaxVal(tempCounts, k)
	return kthMaxCount
}

func selectTopKFeatures(wpInsights *WithinPeriodInsights, k int) {
	if k == -1 {
		return
	}
	numBase := len((*wpInsights).Base.FeatureMetrics)
	if k < numBase {
		kthMaxCountBase := findTopKCountThres((*wpInsights).Base.FeatureMetrics, k)
		totalUserCount := (*wpInsights).Base.GlobalMetrics["#users"]
		filterFeatureCounts(&(*wpInsights).Base.FeatureMetrics, totalUserCount, kthMaxCountBase)
	}
	numBaseAndTarget := len((*wpInsights).BaseAndTarget.FeatureMetrics)
	if k < numBaseAndTarget {
		kthMaxCountTarget := findTopKCountThres((*wpInsights).BaseAndTarget.FeatureMetrics, k)
		totalUserCount := (*wpInsights).BaseAndTarget.GlobalMetrics["#users"]
		filterFeatureCounts(&(*wpInsights).BaseAndTarget.FeatureMetrics, totalUserCount, kthMaxCountTarget)
	}
	// TODO: Should we compute this for (*wpInsights).Target.FeatureMetrics as well?
}

func filterFeatureCounts(featureMetrics *Level3CatRatioDist, totalCount int, kthMaxCount int) {
	var othersCount int = 0
	for key, valStats := range *featureMetrics {
		retainKey := false
		othersFlag := false
		othersCount = 0
		for val, stats := range valStats {
			count := int(stats["#users"])
			if count < kthMaxCount {
				delete((*featureMetrics)[key], val)
				othersCount += count
				othersFlag = true
			} else {
				retainKey = true
			}
		}
		if retainKey {
			if othersFlag {
				(*featureMetrics)[key][OTHERS_VALUE] = make(Level1CatRatioDist)
				(*featureMetrics)[key][OTHERS_VALUE]["#users"] = float64(othersCount)
				(*featureMetrics)[key][OTHERS_VALUE]["prev"] = SmartDivide(float64(othersCount), float64(totalCount))
			}
		} else {
			// TODO: When all values of a key are infrequent (or not in the whitelist), we delete the key itself. Have to see what to do when we have "Others".
			delete((*featureMetrics), key)
		}
	}
}

func findKthMaxVal(list []int, k int) int {
	kInd := k - 1
	n := len(list)
	if kInd > n-1 {
		kInd = n - 1
	} else if kInd == 0 {
		return 0
	}
	var lim int = 0
	if kInd == n {
		lim = kInd - 1
	} else {
		lim = kInd
	}
	for i := 0; i <= lim; i++ {
		for j := i + 1; j < n; j++ {
			if list[i] < list[j] {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	kthMaxVal := list[kInd]
	return kthMaxVal
}

func updateKeyValFlagMap(mapToBeUpdated, updaterMap KeyValCountTable) {
	for key, valCountMap := range updaterMap {
		if _, ok := mapToBeUpdated[key]; ok {
			updateValFlagMap(mapToBeUpdated[key], valCountMap)
		} else {
			mapToBeUpdated[key] = valCountMap
		}
	}
}

func updateValFlagMap(mapToBeUpdated, updaterMap ValCountTable) {
	for val := range updaterMap {
		if val == "" {
			continue
		}
		if _, ok := mapToBeUpdated[val]; !ok {
			mapToBeUpdated[val] = 1
		}
	}
}

func updateKeyValCountMap(mapToBeUpdated, updaterMap KeyValCountTable) {
	for key, valCountMap := range updaterMap {
		if _, ok := mapToBeUpdated[key]; ok {
			updateValCountMap(mapToBeUpdated[key], valCountMap)
		} else {
			mapToBeUpdated[key] = valCountMap
		}
	}
}

func updateValCountMap(mapToBeUpdated, updaterMap ValCountTable) {
	for val, count := range updaterMap {
		if val == "" {
			continue
		}
		if _, ok := mapToBeUpdated[val]; ok {
			mapToBeUpdated[val] += count
		} else {
			mapToBeUpdated[val] = count
		}
	}
}

// func updateCountList(listToBeUpdated, updaterList []uint64) {
// 	for i, item := range updaterList {
// 		listToBeUpdated[i] += item
// 	}
// }

func processUserSessionEvents(userSessionEvents []([]P.CounterEventFormat)) KeyValCountTable {
	// TODO: Process users more sophisticatedly to handle all feature-value combinations.
	if len(userSessionEvents) == 0 {
		return nil
	}
	// TODO: For now, only processing first session per user. Make appropriate updates to handle all sessions.
	firstSessionEvents := userSessionEvents[0]
	userMap := processSessionEvents(firstSessionEvents)
	return userMap
}

func processSessionEvents(sessionEvents []P.CounterEventFormat) KeyValCountTable {
	sessionMap := make(KeyValCountTable)
	for _, event := range sessionEvents {
		// eventName := event.EventName
		processSessionEvent(sessionMap, event)
	}
	return sessionMap
}

func processSessionEvent(sessionMap KeyValCountTable, event P.CounterEventFormat) {
	eventProperties := event.EventProperties
	userProperties := event.UserProperties
	mergedProperties := mergeProperties(eventProperties, userProperties)
	updateKeyValFlagMap(sessionMap, mergedProperties)
}

func makeCountMapFromProperties(propMap map[string]interface{}) KeyValCountTable {
	countMap := make(KeyValCountTable)
	for key, val := range propMap {
		valString := fmt.Sprintf("%v", val)
		countMap[key] = ValCountTable{valString: 1}
	}
	return countMap
}

func mergeProperties(eprMap, uprMap map[string]interface{}) KeyValCountTable {
	mergedMap := make(KeyValCountTable)

	eprCountMap := makeCountMapFromProperties(eprMap)
	updateKeyValFlagMap(mergedMap, eprCountMap)

	uprCountMap := makeCountMapFromProperties(uprMap)
	updateKeyValFlagMap(mergedMap, uprCountMap)

	return mergedMap
}

// SurfaceAndSortTopInsights selects top k insights and ranks them based on a ranking metric.
// func SurfaceAndSortTopInsights(insights CrossPeriodInsights, rankingMetricName string, k int) (CrossPeriodInsights, error) {
// 	var err error
// 	insightsList := insights.Insights
// 	sort.Slice(insightsList[:], func(i, j int) bool {
// 		iMetric := insightsList[i].Metrics[rankingMetricName]
// 		jMetric := insightsList[j].Metrics[rankingMetricName]
// 		return iMetric.Value < jMetric.Value
// 	})
// 	topSortedInsightsList := insightsList[:k]
// 	var topSortedInsights CrossPeriodInsights
// 	topSortedInsights.Periods = insights.Periods
// 	topSortedInsights.Insights = topSortedInsightsList
// 	return topSortedInsights, err
// }

func ComputeSingleDiffMetric(v1, v2 float64) DiffMetric {
	dm := DiffMetric{First: v1, Second: v2, PercentChange: SmartChange(v1, v2, "perc"), FactorChange: SmartChange(v1, v2, "factor")}
	return dm
}

func ComputeCrossPeriodMetrics(wpMetrics1, wpMetrics2 WithinPeriodMetrics) CrossPeriodMetrics {
	globalMetrics := ComputeCrossPeriodGlobalMetrics(wpMetrics1.GlobalMetrics, wpMetrics2.GlobalMetrics)
	featureMetrics := ComputeCrossPeriodFeatureMetrics(wpMetrics1.FeatureMetrics, wpMetrics2.FeatureMetrics)
	cpm := CrossPeriodMetrics{GlobalMetrics: globalMetrics, FeatureMetrics: featureMetrics}
	return cpm
}

func ComputeCrossPeriodRatioMetrics(wpMetrics1, wpMetrics2 WithinPeriodRatioMetrics) CrossPeriodMetrics {
	globalMetrics := ComputeCrossPeriodGlobalRatioMetrics(wpMetrics1.GlobalMetrics, wpMetrics2.GlobalMetrics)
	featureMetrics := ComputeCrossPeriodFeatureRatioMetrics(wpMetrics1.FeatureMetrics, wpMetrics2.FeatureMetrics)
	cprm := CrossPeriodMetrics{GlobalMetrics: globalMetrics, FeatureMetrics: featureMetrics}
	return cprm
}

func ComputeCrossPeriodGlobalRatioMetrics(wpGM1, wpGM2 Level1CatRatioDist) Level1CatDiffDist {
	metricName := "ratio"
	diffMetric := ComputeSingleDiffMetric(wpGM1[metricName], wpGM2[metricName])
	globalMetrics := Level1CatDiffDist{metricName: diffMetric}
	return globalMetrics
}

func ComputeCrossPeriodFeatureRatioMetrics(wpFM1, wpFM2 Level2CatRatioDist) Level2CatDiffDist {
	featureMetrics := make(Level2CatDiffDist)
	for key, valCounts1 := range wpFM1 {
		featureMetrics[key] = make(Level1CatDiffDist)
		valCounts2 := wpFM2[key]
		for val, count1 := range valCounts1 {
			count2 := valCounts2[val]
			diffMetric := ComputeSingleDiffMetric((count1), (count2))
			featureMetrics[key][val] = diffMetric
		}
	}
	return featureMetrics
}

func ComputeCrossPeriodGlobalMetrics(wpGM1, wpGM2 Level1CatFreqDist) Level1CatDiffDist {
	metricName := "#users"
	diffMetric := ComputeSingleDiffMetric(float64(wpGM1[metricName]), float64(wpGM2[metricName]))
	globalMetrics := Level1CatDiffDist{metricName: diffMetric}
	return globalMetrics
}

func ComputeCrossPeriodFeatureMetrics(wpFM1, wpFM2 Level3CatRatioDist) Level2CatDiffDist {
	featureMetrics := make(Level2CatDiffDist)
	for key, valStats1 := range wpFM1 {
		featureMetrics[key] = make(Level1CatDiffDist)
		valStats2 := wpFM2[key]
		for val, stats1 := range valStats1 {
			count1 := stats1["#users"]
			stats2 := valStats2[val]
			count2 := stats2["#users"]
			diffMetric := ComputeSingleDiffMetric(float64(count1), float64(count2))
			featureMetrics[key][val] = diffMetric
		}
	}
	return featureMetrics
}

// GetDeltaInsights takes two feature-count tables and computes delta insights.
func ComputeCrossPeriodInsights(wpi1, wpi2 WithinPeriodInsights) (CrossPeriodInsights, error) {
	var err error
	baseMetrics := ComputeCrossPeriodMetrics(wpi1.Base, wpi2.Base)
	targetMetrics := ComputeCrossPeriodMetrics(wpi1.Target, wpi2.Target)
	baseAndTargetMetrics := ComputeCrossPeriodMetrics(wpi1.BaseAndTarget, wpi2.BaseAndTarget)
	convMetrics := ComputeCrossPeriodRatioMetrics(wpi1.Conversion, wpi2.Conversion)
	deltaRatioMetrics := ComputeDeltaRatioMetrics(wpi1, wpi2)
	jsdMetrics := ComputeJSDMetrics(wpi1, wpi2)
	cpInsights := CrossPeriodInsights{Base: baseMetrics, Target: targetMetrics, BaseAndTarget: baseAndTargetMetrics,
		Conversion: convMetrics, DeltaRatio: deltaRatioMetrics, JSDivergence: jsdMetrics}
	return cpInsights, err
}

func ComputeDeltaRatioMetrics(wpi1, wpi2 WithinPeriodInsights) Level2CatRatioDist {
	drMetrics := make(Level2CatRatioDist)
	u1 := float64(wpi1.Base.GlobalMetrics["#users"])
	u2 := float64(wpi2.Base.GlobalMetrics["#users"])
	m1 := float64(wpi1.Target.GlobalMetrics["#users"])
	m2 := float64(wpi2.Target.GlobalMetrics["#users"])
	for key, valStats1 := range wpi1.Base.FeatureMetrics {
		drMetrics[key] = make(Level1CatRatioDist)
		for val, _ := range valStats1 {
			f1 := wpi1.Base.FeatureMetrics[key][val]["#users"]
			f2 := wpi2.Base.FeatureMetrics[key][val]["#users"]
			fm1 := wpi1.Target.FeatureMetrics[key][val]["#users"]
			fm2 := wpi2.Target.FeatureMetrics[key][val]["#users"]
			crf1 := wpi1.Conversion.FeatureMetrics[key][val]
			notf1 := u1 - f1
			notfm1 := m1 - fm1
			notf2 := u2 - f2
			notfm2 := m2 - fm2
			cr_notf1 := SmartDivide(notfm1, notf1)
			fm2_pred := f2 * crf1
			notfm2_pred := notf2 * cr_notf1
			D_fm := fm2_pred - fm2
			D_notfm := notfm2_pred - notfm2
			delRatio := SmartDivide(math.Abs(D_fm), math.Abs(D_fm)+math.Abs(D_notfm))
			drMetrics[key][val] = delRatio
		}
	}
	return drMetrics
}

func ComputeJSDMetrics(wp1, wp2 WithinPeriodInsights) JSDType {
	baseJSDMetrics := MultipleJSDivergence(wp1.Base.FeatureMetrics, wp2.Base.FeatureMetrics)
	targetJSDMetrics := MultipleJSDivergence(wp1.Target.FeatureMetrics, wp2.Target.FeatureMetrics)
	jsdMetrics := JSDType{Base: baseJSDMetrics, Target: targetJSDMetrics}
	return jsdMetrics
}

func MultipleJSDivergence(featureMetrics1, featureMetrics2 Level3CatRatioDist) Level2CatRatioDist {
	jsdMetrics := make(Level2CatRatioDist)
	for key, valStats1 := range featureMetrics1 {
		jsdMetrics[key] = make(Level1CatRatioDist)
		for val, stats1 := range valStats1 {
			prev1 := stats1["prev"]
			prev2 := featureMetrics2[key][val]["prev"]
			jsd := SingleJSDivergence(prev1, prev2)
			jsdMetrics[key][val] = jsd
		}
	}
	return jsdMetrics
}

// SingleJSDivergence computes JSD between distributions [p, 1-p] and [q, 1-q]
func SingleJSDivergence(p, q float64) float64 {
	m := 0.5 * (p + q)
	kld1 := SingleKLDivergence(p, m)
	kld2 := SingleKLDivergence(q, m)
	jsd := 0.5 * (kld1 + kld2)
	if math.IsNaN(jsd) {
		fmt.Printf("p=%f\nq=%f\nkld1=%f\nkld2=%f\njsd=%f", p, q, kld1, kld2, jsd)
	}
	return jsd
}

// SingleKLDivergence computes KLD between distributions [a, 1-b] and [a, 1-b]
func SingleKLDivergence(a, b float64) float64 {
	kld := SmartALog2BByC(a, a, b) + SmartALog2BByC(1-a, 1-a, 1-b)
	return kld
}

func SmartALog2BByC(a, b, c float64) float64 {
	if a == 0.0 || c == 0 {
		return 0.0
	}
	return a * math.Log2(SmartDivide(b, c))
}

// PublishDeltaInsights stores delta insights into a file from where frontend can read it.
func PublishDeltaInsights(topSortedInsights CrossPeriodInsights, filePath string) error {
	var err error
	writeAsJSON(topSortedInsights, filePath)
	return err
}

// GetEventFileScanner Return handle to events file scanner
func GetEventFileScanner(projectId uint64, periodCode Period, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, insightGranularity string, isDownloaded bool) (*bufio.Scanner, error) {
	var err error
	efTmpPath, efTmpName := diskManager.GetModelEventsFilePathAndName(projectId, periodCode.From, insightGranularity)
	if isDownloaded == false {
		efCloudPath, efCloudName := (*cloudManager).GetModelEventsFilePathAndName(projectId, periodCode.From, insightGranularity)
		deltaComputeLog.WithFields(log.Fields{"eventFileCloudPath": efCloudPath,
			"eventFileCloudName": efCloudName}).Info("Downloading events file from cloud.")
		eReader, err := (*cloudManager).Get(efCloudPath, efCloudName)
		if err != nil {
			deltaComputeLog.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
				"eventFileName": efCloudName}).Error("Failed downloading events file from cloud.")
			return nil, err
		}
		err = diskManager.Create(efTmpPath, efTmpName, eReader)
		if err != nil {
			deltaComputeLog.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
				"eventFileName": efCloudName}).Error("Failed downloading events file from cloud.")
			return nil, err
		}
	}
	// TODO: Change this to efTmpPath and efTmpName and write the code to download an events file from cloud to local.
	// We already have the logic to dump an events file from DB to local and from local to cloud, but not from cloud to local.
	// The following eventsFilePath will run fine if the cloud_path is a local one. But not if it's an actual remote cloud path.
	eventsFilePath := efTmpPath + efTmpName
	fmt.Println(eventsFilePath)
	scanner, err := T.OpenEventFileAndGetScanner(eventsFilePath)
	if err != nil {
		deltaComputeLog.WithFields(log.Fields{"err": err,
			"eventsFilePath": eventsFilePath}).Error("Failed opening event file and getting scanner.")
	}
	return scanner, err
}
