package delta

import (
	"bufio"
	"encoding/json"
	"factors/filestore"
	"factors/model/model"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	T "factors/task"
	U "factors/util"
	"fmt"
	"math"
	"strconv"
	"strings"

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
	// If catagorical property then there will be both AND and OR. Though AND doesnt makes sense there is a possiblity
	// Will add log to check if there are any queries like that
	// if numerical, it can never be OR it will always be AND
	// if datetime, it can always be OR again
	// even boolean are getting mapped as categorical but comparison doesnt work directly since the db and event file has it as boolen type
	// if a key is missing in the event, we cant return false since the comparison can be against $none which will be true in case
	filterKey := filterCriterion.Key
	filterValues := filterCriterion.Values
	mode := filterCriterion.PropertiesMode
	var props map[string]interface{}
	if mode == USER_PROPERTIES_MODE {
		props = event.UserProperties
	} else if mode == EVENT_PROPERTIES_MODE {
		props = event.EventProperties
	}

	propertyValue, ok := props[filterKey]
	if filterCriterion.Type == U.PropertyTypeNumerical {
		return matchFitlerValuesForNumerical(propertyValue, ok, filterValues)
	}
	if filterCriterion.Type == U.PropertyTypeCategorical {
		return matchFitlerValuesForCategorical(propertyValue, ok, filterValues)
	}
	if filterCriterion.Type == U.PropertyTypeDateTime {
		return matchFitlerValuesForDatetime(propertyValue, ok, filterValues)
	}
	return false
}

func matchFitlerValuesForCategorical(eventPropValue interface{}, isPresentEventPropValue bool, filterValues []OperatorValueTuple) bool {
	/*
		Categorical	,=, !=, contains, not contains
		A = b	right
		A != b	right
		A= b and A = c	not logical but right
		A = b OR A = c	right
		A != b and A != c 	right
		A != b OR A != c	wrong
		A = b OR A != c	not possible
		A != b OR A = c	not possible
		A = b OR A = c and A = d	ordering is wrong
		A != b OR A != c and A = d	ordering is wrong
		A = b OR A = c and A != d	ordering is wrong
		A != b OR A != c and A != d	ordering is wrong
	*/
	andCount, orCount, containsCount, equalsCount := 0, 0, 0, 0
	for _, value := range filterValues {
		if value.LogicalOp == "AND" {
			andCount++
		}
		if value.LogicalOp == "OR" {
			orCount++
		}
		if value.Operator == model.EqualsOpStr || value.Operator == model.NotEqualOpStr {
			equalsCount++
		}
		if value.Operator == model.ContainsOpStr || value.Operator == model.NotContainsOpStr {
			containsCount++
		}
	}
	/*
		Rejection cases
		With same property in two different rows with atleast one containining multiple	count(AND) > 1 and count(OR) >= 1
		multiple with !=	op = != and count(OR) >= 1
	*/
	if andCount > 1 && orCount >= 1 && equalsCount == 0 && containsCount >= 1 {
		/*
			A = b OR A = c and A = d	ordering is wrong
			A != b OR A != c and A = d	ordering is wrong
			A = b OR A = c and A != d	ordering is wrong
			A != b OR A != c and A != d	ordering is wrong
		*/
		/* in case of the above ones with contains and not contains it is valid. its only a rejection case in = and != combinations */
		return false
		// "Multiple filters with same property and any one with multi select"
	}
	if andCount == 1 && orCount >= 1 && (filterValues[0].Operator == model.NotEqualOpStr || filterValues[0].Operator == model.NotContainsOpStr) {
		//A != b OR A != c	wrong
		return false
		// "Multi select filter with not equals"
	}
	// TODO What to do if there is a misclassification
	results := make(map[int]bool)
	propertyValue := fmt.Sprintf("%v", eventPropValue)

	for i, value := range filterValues {
		if value.Value == model.PropertyValueNone {
			results[i] = handleNoneCase(propertyValue, isPresentEventPropValue, value.Operator)
			continue
		}
		if value.Operator == model.EqualsOpStr {
			results[i] = strings.EqualFold(propertyValue, value.Value)
		}
		if value.Operator == model.NotEqualOpStr {
			results[i] = !strings.EqualFold(propertyValue, value.Value)
		}
		if value.Operator == model.ContainsOpStr {
			results[i] = strings.Contains(strings.ToLower(propertyValue), strings.ToLower(value.Value))
		}
		if value.Operator == model.NotContainsOpStr {
			results[i] = !(strings.Contains(strings.ToLower(propertyValue), strings.ToLower(value.Value)))
		}
	}
	var soFar bool
	var op string
	for i := range filterValues {
		op = filterValues[i].LogicalOp
		if i == 0 {
			soFar = results[i]
		} else {
			if op == "AND" {
				soFar = soFar && results[i]
			}
			if op == "OR" {
				soFar = soFar || results[i]
			}
		}
	}
	return soFar
}

func handleNoneCase(eventPropValue string, isPresentEventPropValue bool, operator string) bool {
	if operator == model.EqualsOpStr || operator == model.ContainsOpStr {
		return !isPresentEventPropValue || eventPropValue == "$none"
	}
	if operator == model.NotEqualOpStr || operator == model.NotContainsOpStr {
		return isPresentEventPropValue && eventPropValue != "$none"
	}
	return false
}

func matchFitlerValuesForNumerical(eventPropValue interface{}, isPresentEventPropValue bool, filterValues []OperatorValueTuple) bool {
	/*
		Numerical	<, <=, >, >=, = , !=
		A = 1 and A  = 2	right
		a = 1 or a= 2	not possible
	*/
	// TODO What to do if there is a misclassification
	results := make(map[int]bool)
	propertyValue := fmt.Sprintf("%v", eventPropValue)
	eventPropertyValue, err := strconv.ParseFloat(propertyValue, 64)
	if err != nil {
		return false
	}
	for i, value := range filterValues {
		filterValue, err := strconv.ParseFloat(value.Value, 64)
		if err != nil {
			return false
		}
		if value.Operator == model.EqualsOpStr {
			results[i] = eventPropertyValue == filterValue
		}
		if value.Operator == model.NotEqualOpStr {
			results[i] = eventPropertyValue != filterValue
		}
		if value.Operator == model.GreaterThanOpStr {
			results[i] = eventPropertyValue > filterValue
		}
		if value.Operator == model.LesserThanOpStr {
			results[i] = eventPropertyValue < filterValue
		}
		if value.Operator == model.GreaterThanOrEqualOpStr {
			results[i] = eventPropertyValue >= filterValue
		}
		if value.Operator == model.LesserThanOrEqualOpStr {
			results[i] = eventPropertyValue <= filterValue
		}
	}
	var soFar bool
	var op string
	for i := range filterValues {
		op = filterValues[i].LogicalOp

		if i == 0 {
			soFar = results[i]
		} else {
			if op == "AND" {
				soFar = soFar && results[i]
			}
			if op == "OR" {
				soFar = soFar || results[i]
			}
		}
	}
	return soFar
}

func matchFitlerValuesForDatetime(eventPropValue interface{}, isPresentEventPropValue bool, filterValues []OperatorValueTuple) bool {
	results := make(map[int]bool)
	propertyValue := fmt.Sprintf("%v", eventPropValue)
	eventPropertyValue, err := strconv.ParseInt(propertyValue, 10, 64)
	if err != nil {
		return false
	}
	for i, value := range filterValues {
		if value.Operator == model.EqualsOpStr {
			dateTimeFilter, err := model.DecodeDateTimePropertyValue(value.Value)
			if err != nil {
				return false
			}
			if eventPropertyValue >= dateTimeFilter.From && eventPropertyValue <= dateTimeFilter.To {
				results[i] = true
			} else {
				results[i] = false
			}
		}
	}
	var soFar bool
	var op string
	for i := range filterValues {
		op = filterValues[i].LogicalOp
		if i == 0 {
			soFar = results[i]
		} else {
			if op == "AND" {
				soFar = soFar && results[i]
			}
			if op == "OR" {
				soFar = soFar || results[i]
			}
		}
	}
	return soFar
}

func eventMatchesFilterCriterionList(event P.CounterEventFormat, filterCriterionList []EventFilterCriterion) bool {
	for _, fc := range filterCriterionList {
		// Today we dont support OR across filters. So retaining it this way. Its always a AND
		if !eventMatchesFilterCriterion(event, fc) { // "AND" logic: If even a single filter fails, return False.
			return false
		}
	}
	return true
}

func EventMatchesCriterion(event P.CounterEventFormat, eventCriterion EventCriterion) bool {
	// TODO: Match event filters as well.
	nameMatchFlag := eventCriterion.EqualityFlag == (event.EventName == eventCriterion.Name)
	if !nameMatchFlag {
		return false
	}
	filterMatchFlag := eventMatchesFilterCriterionList(event, eventCriterion.FilterCriterionList)
	return filterMatchFlag
}

func updateCriteriaResult(event P.CounterEventFormat, criteria EventsCriteria, criteriaResult *PerUserCriteriaResult, isBase bool) bool {
	customBlacklist := GetCustomBlacklist()
	if criteriaResult.criteriaMatchFlag && isBase { // If criteria already met, do nothing.
		return false
	}
	if len(criteria.EventCriterionList) == 0 { // If criteria has no event criterion to match
		log.Info("Criteria has empty criterion list. By default, making the user match the first event.")
		(*criteriaResult).anyFlag = true
		(*criteriaResult).allFlag = true
		FilterBlacklist(&event, &customBlacklist)
		if !(*criteriaResult).criteriaMatchFlag {
			(*criteriaResult).firstEvent = event
		}
		(*criteriaResult).criteriaMatchFlag = true
		(*criteriaResult).mostRecentEvent = event
		return true
	}
	for i, eventCriterion := range criteria.EventCriterionList {
		if EventMatchesCriterion(event, eventCriterion) {
			(*criteriaResult).criterionResultList[i].matchId = (*criteriaResult).numCriterionMatched
			(*criteriaResult).numCriterionMatched++
			(*criteriaResult).anyFlag = true
			FilterBlacklist(&event, &customBlacklist)
			(*criteriaResult).mostRecentEvent = event
			if (*criteriaResult).numCriterionMatched == len(criteria.EventCriterionList) {
				(*criteriaResult).allFlag = true
			}
			if ((criteria.Operator == "And") && (*criteriaResult).allFlag) ||
				((criteria.Operator == "Or") && (*criteriaResult).anyFlag) {
				if !(*criteriaResult).criteriaMatchFlag {
					(*criteriaResult).firstEvent = event
				}
				(*criteriaResult).criteriaMatchFlag = true
				return true
			}
		}
	}
	return false
}

func updateCriteriaResultEventOccurence(event P.CounterEventFormat, criteria EventsCriteria, criteriaResult *PerUserCriteriaResult, isBase bool) bool {
	customBlacklist := GetCustomBlacklist()
	if len(criteria.EventCriterionList) == 0 { // If criteria has no event criterion to match
		log.Info("Criteria has empty criterion list. By default, making the user match the first event.")
		(*criteriaResult).anyFlag = true
		(*criteriaResult).allFlag = true
		FilterBlacklist(&event, &customBlacklist)
		if !(*criteriaResult).criteriaMatchFlag {
			(*criteriaResult).firstEvent = event
		}
		(*criteriaResult).criteriaMatchFlag = true
		(*criteriaResult).mostRecentEvent = event
		return true
	}
	for i, eventCriterion := range criteria.EventCriterionList {
		if EventMatchesCriterion(event, eventCriterion) {
			(*criteriaResult).criterionResultList[i].matchId = (*criteriaResult).numCriterionMatched
			(*criteriaResult).numCriterionMatched++
			(*criteriaResult).anyFlag = true
			FilterBlacklist(&event, &customBlacklist)
			(*criteriaResult).mostRecentEvent = event
			if (*criteriaResult).numCriterionMatched == len(criteria.EventCriterionList) {
				(*criteriaResult).allFlag = true
			}
			if ((criteria.Operator == "And") && (*criteriaResult).allFlag) ||
				((criteria.Operator == "Or") && (*criteriaResult).anyFlag) {
				if !(*criteriaResult).criteriaMatchFlag {
					(*criteriaResult).firstEvent = event
				}
				(*criteriaResult).criteriaMatchFlag = true
				return true
			}
		}
	}
	return false
}

func QueryEvent(event P.CounterEventFormat, deltaQuery Query, perUserQueryResult *PerUserQueryResult) (bool, bool) {
	base := updateCriteriaResult(event, deltaQuery.Base, &(*perUserQueryResult).baseResult, true)
	target := updateCriteriaResult(event, deltaQuery.Target, &(*perUserQueryResult).targetResult, false)
	return base, target
}

func QueryEventEventOccurence(event P.CounterEventFormat, deltaQuery Query, perUserQueryResult *PerUserQueryResult) bool {
	target := updateCriteriaResultEventOccurence(event, deltaQuery.Target, &(*perUserQueryResult).targetResult, false)
	return target
}

func QuerySessionMultiStepFunnel(session Session, multiStageFunnel MultiFunnelQuery, timestamp *map[int][]int64, index *map[int][]int, i int) {
	it := i
	intermediateLength := len(multiStageFunnel.Intermediate)
	for _, event := range session.Events {
		criteriaResult := makeCriteriaResult(multiStageFunnel.Base)
		base := updateCriteriaResult(event, multiStageFunnel.Base, &criteriaResult, true)
		if (*timestamp)[0] == nil {
			(*timestamp)[0] = make([]int64, 0)
		}
		if (*index)[0] == nil {
			(*index)[0] = make([]int, 0)
		}
		if base {
			(*timestamp)[0] = append((*timestamp)[0], event.EventTimestamp)
			(*index)[0] = append((*index)[0], it)
		}
		for iteratorIndex, intermediate := range multiStageFunnel.Intermediate {
			criteriaResult = makeCriteriaResult(intermediate)
			intermediate := updateCriteriaResult(event, intermediate, &criteriaResult, false)
			if (*timestamp)[iteratorIndex+1] == nil {
				(*timestamp)[iteratorIndex+1] = make([]int64, 0)
			}
			if (*index)[iteratorIndex+1] == nil {
				(*index)[iteratorIndex+1] = make([]int, 0)
			}
			if intermediate {
				(*timestamp)[iteratorIndex+1] = append((*timestamp)[iteratorIndex+1], event.EventTimestamp)
				(*index)[iteratorIndex+1] = append((*index)[iteratorIndex+1], it)
			}
		}
		criteriaResult = makeCriteriaResult(multiStageFunnel.Target)
		target := updateCriteriaResult(event, multiStageFunnel.Target, &criteriaResult, false)
		if (*timestamp)[intermediateLength+1] == nil {
			(*timestamp)[intermediateLength+1] = make([]int64, 0)
		}
		if (*index)[intermediateLength+1] == nil {
			(*index)[intermediateLength+1] = make([]int, 0)
		}
		if target {
			(*timestamp)[intermediateLength+1] = append((*timestamp)[intermediateLength+1], event.EventTimestamp)
			(*index)[intermediateLength+1] = append((*index)[intermediateLength+1], it)
		}
		it++
	}
}
func QuerySession(session Session, deltaQuery Query, perUserQueryResult *PerUserQueryResult, baseTimestamp *int64, targetTimestamp *int64, baseIndex *int, targetIndex *int, i int) {
	index := i
	for _, event := range session.Events {
		base, target := QueryEvent(event, deltaQuery, perUserQueryResult)
		if base {
			(*baseTimestamp) = event.EventTimestamp
			(*baseIndex) = index
		}
		if target {
			(*targetTimestamp) = event.EventTimestamp
			(*targetIndex) = index
		}
		if *baseTimestamp <= *targetTimestamp && *baseTimestamp != -1 && *targetTimestamp != -1 && *baseIndex != *targetIndex {
			break
		}
		index++
	}
}

func QuerySessionEventOccurence(session Session, deltaQuery Query, perUserQueryResult *PerUserQueryResult, targetTimestamp *[]int64, targetIndex *[]int, i int) {
	index := i
	for _, event := range session.Events {
		target := QueryEventEventOccurence(event, deltaQuery, perUserQueryResult)
		if target {
			(*targetTimestamp) = append((*targetTimestamp), event.EventTimestamp)
			(*targetIndex) = append((*targetIndex), index)
		}
		index++
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

func MakePerUserQueryResult(deltaQuery Query) PerUserQueryResult {
	baseCriteriaResult := makeCriteriaResult(deltaQuery.Base)
	targetCriteriaResult := makeCriteriaResult(deltaQuery.Target)
	perUserQueryResult := PerUserQueryResult{baseResult: baseCriteriaResult, targetResult: targetCriteriaResult}
	return perUserQueryResult
}

type Session struct {
	Events []P.CounterEventFormat
}

func QueryUser(preSessionEvents []P.CounterEventFormat, sessions []Session, deltaQuery Query) (PerEventProperties, error) {
	isSessionTarget := false
	for _, target := range deltaQuery.Target.EventCriterionList {
		if target.Name == "$session" {
			isSessionTarget = true
		}
	}
	baseTimestamp, targetTimestamp := int64(-1), int64(-1)
	baseIndex, targetIndex := int(-1), int(-1)
	i := int(0)
	var err error = nil
	userResult := MakePerUserQueryResult(deltaQuery)
	var extendedSessions []Session // Extended sessions are preSessionEvents + sessions.
	extendedSessions = append(extendedSessions, Session{Events: preSessionEvents})
	extendedSessions = append(extendedSessions, sessions...)
	for _, session := range extendedSessions {
		QuerySession(session, deltaQuery, &userResult, &baseTimestamp, &targetTimestamp, &baseIndex, &targetIndex, i)
		if userResult.baseResult.criteriaMatchFlag && userResult.targetResult.criteriaMatchFlag && baseTimestamp <= targetTimestamp && baseIndex != targetIndex {
			break
		}
		i = i + len(session.Events)
	}
	summary := PerEventProperties{BaseFlag: userResult.baseResult.criteriaMatchFlag,
		TargetFlag: userResult.targetResult.criteriaMatchFlag}
	summary.BaseAndTargetFlag = userResult.baseResult.criteriaMatchFlag && userResult.targetResult.criteriaMatchFlag && baseTimestamp <= targetTimestamp && baseIndex != targetIndex
	summary.EventProperties = make(map[string]interface{})
	summary.UserProperties = make(map[string]interface{})
	if summary.BaseAndTargetFlag {
		// both source and target properties
		combineSourceAndTargetProperties(userResult.baseResult.mostRecentEvent, userResult.targetResult.mostRecentEvent, &summary.EventProperties, &summary.UserProperties, isSessionTarget)
	}
	if summary.BaseFlag {
		combineSourceAndTargetProperties(userResult.baseResult.mostRecentEvent, P.CounterEventFormat{}, &summary.EventProperties, &summary.UserProperties, isSessionTarget)
	}
	if summary.TargetFlag {
		combineSourceAndTargetProperties(P.CounterEventFormat{}, userResult.targetResult.firstEvent, &summary.EventProperties, &summary.UserProperties, isSessionTarget)
	}
	return summary, err
}

func QueryUserMultiStepFunnel(preSessionEvents []P.CounterEventFormat, sessions []Session, deltaQuery MultiFunnelQuery) (PerEventProperties, error) {
	isSessionTarget := false
	for _, target := range deltaQuery.Target.EventCriterionList {
		if target.Name == "$session" {
			isSessionTarget = true
		}
	}
	totalFunnelEvents := 2 + len(deltaQuery.Intermediate)
	timestamp := make(map[int][]int64)
	index := make(map[int][]int)
	i := int(0)
	var err error = nil
	var extendedSessions []Session // Extended sessions are preSessionEvents + sessions.
	extendedSessions = append(extendedSessions, Session{Events: preSessionEvents})
	extendedSessions = append(extendedSessions, sessions...)
	isEntireFunnelFound := false
	for _, session := range extendedSessions {
		QuerySessionMultiStepFunnel(session, deltaQuery, &timestamp, &index, i)
		if timestamp[0] == nil || len(timestamp[0]) <= 0 {
			continue
		}
		maxTimestamp := timestamp[0][0]
		maxIndex := index[0][0]
		for it := 1; it < totalFunnelEvents; it++ {
			isLevelFunnelFound := false
			if timestamp[it] == nil || len(timestamp[it]) <= 0 {
				break
			}
			for iteratorI := range timestamp[it] {
				if timestamp[it][iteratorI] >= maxTimestamp && index[it][iteratorI] != maxIndex {
					maxTimestamp = timestamp[it][iteratorI]
					maxIndex = index[it][iteratorI]
					isLevelFunnelFound = true
				}
				if isLevelFunnelFound {
					break
				}
			}
			if !isLevelFunnelFound {
				break
			}
			if it == totalFunnelEvents-1 {
				isEntireFunnelFound = true
			}
		}
		if isEntireFunnelFound {
			break
		}
		i = i + len(session.Events)
	}
	events := make([]P.CounterEventFormat, 0)
	for _, session := range extendedSessions {
		events = append(events, session.Events...)
	}
	summary := PerEventProperties{}
	summary.EventProperties = make(map[string]interface{})
	summary.UserProperties = make(map[string]interface{})
	if isEntireFunnelFound {
		// both source and target properties
		summary.BaseAndTargetFlag = true
		combineSourceAndTargetProperties(events[index[0][0]], events[index[totalFunnelEvents-1][len(index[totalFunnelEvents-1])-1]], &summary.EventProperties, &summary.UserProperties, isSessionTarget)
	}
	if timestamp[0] != nil && len(timestamp[0]) > 0 {
		summary.BaseFlag = true
		combineSourceAndTargetProperties(events[index[0][0]], P.CounterEventFormat{}, &summary.EventProperties, &summary.UserProperties, isSessionTarget)
	}
	if timestamp[totalFunnelEvents-1] != nil && len(timestamp[totalFunnelEvents-1]) > 0 {
		summary.TargetFlag = true
		combineSourceAndTargetProperties(P.CounterEventFormat{}, events[index[totalFunnelEvents-1][len(index[totalFunnelEvents-1])-1]], &summary.EventProperties, &summary.UserProperties, isSessionTarget)
	}
	return summary, err
}
func QueryUserEventOccurence(preSessionEvents []P.CounterEventFormat, sessions []Session, deltaQuery Query) ([]PerEventProperties, error) {
	isSessionTarget := false
	for _, target := range deltaQuery.Target.EventCriterionList {
		if target.Name == "$session" {
			isSessionTarget = true
		}
	}
	targetTimestamp := make([]int64, 0)
	targetIndex := make([]int, 0)
	i := int(0)
	var err error = nil
	userResult := MakePerUserQueryResult(deltaQuery)
	var extendedSessions []Session // Extended sessions are preSessionEvents + sessions.
	extendedSessions = append(extendedSessions, Session{Events: preSessionEvents})
	extendedSessions = append(extendedSessions, sessions...)
	for _, session := range extendedSessions {
		QuerySessionEventOccurence(session, deltaQuery, &userResult, &targetTimestamp, &targetIndex, i)
		i = i + len(session.Events)
	}
	events := make([]P.CounterEventFormat, 0)
	for _, session := range extendedSessions {
		events = append(events, session.Events...)
	}
	summary := make([]PerEventProperties, 0)
	for _, selectedIndex := range targetIndex {
		ep, up := make(map[string]interface{}), make(map[string]interface{})
		combineSourceAndTargetProperties(P.CounterEventFormat{}, events[selectedIndex], &ep, &up, isSessionTarget)
		summary = append(summary, PerEventProperties{EventProperties: ep, UserProperties: up})
	}
	return summary, err
}

func combineSourceAndTargetProperties(base P.CounterEventFormat, target P.CounterEventFormat, eventProperties *map[string]interface{}, userProperties *map[string]interface{}, isSessionTarget bool) {
	if base.EventProperties != nil && !isSessionTarget {
		for key, value := range base.EventProperties {
			propertyKey := fmt.Sprintf("s#%s", key)
			(*eventProperties)[propertyKey] = value
		}
	}
	if target.EventProperties != nil {
		for key, value := range target.EventProperties {
			propertyKey := fmt.Sprintf("s#%s", key)
			_, ok := (*eventProperties)[propertyKey]
			if ok {
				continue
			}
			propertyKey = fmt.Sprintf("t#%s", key)
			(*eventProperties)[propertyKey] = value
		}
	}
	if base.UserProperties != nil && !isSessionTarget {
		for key, value := range base.UserProperties {
			propertyKey := fmt.Sprintf("s#%s", key)
			(*userProperties)[propertyKey] = value
		}
	}
	if target.UserProperties != nil {
		for key, value := range target.UserProperties {
			propertyKey := fmt.Sprintf("s#%s", key)
			_, ok := (*userProperties)[propertyKey]
			if ok {
				continue
			}
			propertyKey = fmt.Sprintf("t#%s", key)
			(*userProperties)[propertyKey] = value
		}
	}
}

func getEventMatchFlag(event P.CounterEventFormat, eventCriterion EventCriterion) bool {
	matchFlag := (event.EventName == eventCriterion.Name)
	return matchFlag
}

func GetCustomBlacklist() map[string]bool {
	// TODO: This was changed from set to map
	customBlacklistMap := make(map[string]bool)
	for _, featKey := range U.CUSTOM_BLACKLIST_DELTA {
		customBlacklistMap[featKey] = true
	}
	return customBlacklistMap
}

func ComputeWithinPeriodInsights(scanner *bufio.Scanner, deltaQuery Query, multiStepQuery MultiFunnelQuery, k int, featSoftWhitelist map[string]map[string]bool, passId int, isEventOccurence bool, isMultistepFunnel bool) (WithinPeriodInsights, error) {
	var err error
	var wpInsights WithinPeriodInsights
	var prevUserId string = ""
	var matchSummary PerEventProperties
	var matchedBaseEvents []PerEventProperties
	var matchedTargetEvents []PerEventProperties
	var matchedBaseAndTargetEvents []PerEventProperties
	var sessions []Session
	var preSessionEvents []P.CounterEventFormat = nil
	sessionId := -1
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum%10000 == 0 {
			fmt.Printf("%d lines scanned\n", lineNum)
		}
		line := scanner.Text()
		var event P.CounterEventFormat
		json.Unmarshal([]byte(line), &event) // TODO: Add error check.
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
			if isMultistepFunnel {
				matchSummary, err = QueryUserMultiStepFunnel(preSessionEvents, sessions, multiStepQuery)
				if err != nil {
					return wpInsights, err
				}
				if matchSummary.BaseFlag {
					matchedBaseEvents = append(matchedBaseEvents, matchOnlyPrefixedPropertiesWithPrefix(matchSummary, "s#"))
				}
				if matchSummary.TargetFlag {
					matchedTargetEvents = append(matchedTargetEvents, matchOnlyPrefixedPropertiesWithPrefix(matchSummary, "t#"))
				}
				if matchSummary.BaseAndTargetFlag {
					matchedBaseAndTargetEvents = append(matchedBaseAndTargetEvents, matchSummary)
				}
			} else if !isEventOccurence {
				matchSummary, err = QueryUser(preSessionEvents, sessions, deltaQuery)
				if err != nil {
					return wpInsights, err
				}
				if matchSummary.BaseFlag {
					matchedBaseEvents = append(matchedBaseEvents, matchOnlyPrefixedPropertiesWithPrefix(matchSummary, "s#"))
				}
				if matchSummary.TargetFlag {
					matchedTargetEvents = append(matchedTargetEvents, matchOnlyPrefixedPropertiesWithPrefix(matchSummary, "t#"))
				}
				if matchSummary.BaseAndTargetFlag {
					matchedBaseAndTargetEvents = append(matchedBaseAndTargetEvents, matchSummary)
				}
			} else {
				matchSummaries, err := QueryUserEventOccurence(preSessionEvents, sessions, deltaQuery)
				if err != nil {
					return wpInsights, err
				}
				for _, matchSummary := range matchSummaries {
					matchedTargetEvents = append(matchedTargetEvents, matchOnlyPrefixedPropertiesWithPrefix(matchSummary, "t#"))
				}
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
		PrefilterFeatures(&wpInsights)
		selectTopKFeatures(&(wpInsights), k)
	}
	return wpInsights, err
}

func matchOnlyPrefixedPropertiesWithPrefix(event PerEventProperties, prefix string) PerEventProperties {
	changedEvent := PerEventProperties{}
	changedEvent.EventProperties = make(map[string]interface{})
	changedEvent.UserProperties = make(map[string]interface{})
	for key, value := range event.UserProperties {
		if strings.HasPrefix(key, prefix) {
			changedEvent.UserProperties[key] = value
		}
	}
	for key, value := range event.EventProperties {
		if strings.HasPrefix(key, prefix) {
			changedEvent.EventProperties[key] = value
		}
	}
	return changedEvent
}

// RFDParams stands for Parameters of a Rank-Frequency Distribution
type RFDParams struct {
	numNonNullUniq            int
	numNonNullMultiOccurrence int
	// percNonNullSingleOccurrence float32
	numNonNullSingleOccurrence int
	totalCount                 int
}

// Analyzes the RFD params of a feature and decides whether or not to retain it.
func AnalyzeRFDParams(rfd RFDParams) bool {
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

func AnalyzeFeature(key string, valStats Level2CatRatioDist) (bool, RFDParams) {
	rfd := countRFD(valStats)
	numericFlag := IsANumericFeature(key, valStats)
	return numericFlag, rfd
}

func PrefilterFeatures(wpInsights *WithinPeriodInsights) {
	PrefilterFeaturesProperties(&((*wpInsights).Base.FeatureMetrics))
	PrefilterFeaturesProperties(&((*wpInsights).Target.FeatureMetrics))
	PrefilterFeaturesProperties(&((*wpInsights).BaseAndTarget.FeatureMetrics))
}

func PrefilterFeaturesProperties(featureMetrics *Level3CatRatioDist) {
	for key, valStats := range *featureMetrics {
		keyNumericFlag, keyRFDParams := AnalyzeFeature(key, valStats)
		if keyNumericFlag || !AnalyzeRFDParams(keyRFDParams) {
			delete((*featureMetrics), key)
		}
	}
}

func FilterBlacklist(event *P.CounterEventFormat, customBlacklist *map[string]bool) {
	filterPropertiesBlacklist(&((*event).EventProperties), customBlacklist)
	filterPropertiesBlacklist(&((*event).UserProperties), customBlacklist)
}

func filterPropertiesBlacklist(properties *map[string]interface{}, customBlacklist *map[string]bool) {
	for key := range *properties {
		if (*customBlacklist)[key] {
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

func eventsToMetrics(events []PerEventProperties, featSoftWhitelist map[string]map[string]bool, passId int) WithinPeriodMetrics {
	globalMetrics := make(Level1CatFreqDist)
	featureMetrics := make(Level3CatRatioDist)
	for _, event := range events {
		UpdateLevel1CatFreqDist(&globalMetrics, "#users", 1)
		eventProperties := make(map[string]interface{})
		userProperties := make(map[string]interface{})
		for key, value := range event.UserProperties {
			splits := strings.SplitN(key, "#", 2)
			userProperties[fmt.Sprintf("%v#%v#%v", splits[0], "up", splits[1])] = value
		}
		for key, value := range event.EventProperties {
			splits := strings.SplitN(key, "#", 2)
			eventProperties[fmt.Sprintf("%v#%v#%v", splits[0], "ep", splits[1])] = value
		}
		updateMetricsWithProperties(&featureMetrics, userProperties, featSoftWhitelist, passId)
		updateMetricsWithProperties(&featureMetrics, eventProperties, featSoftWhitelist, passId)
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

func updateMetricsWithProperties(featureMetrics *Level3CatRatioDist, properties map[string]interface{}, featSoftWhitelist map[string]map[string]bool, passId int) {
	for key, val := range properties {
		// TODO: As of now, we are treating all (user/event) properties as categorical variables.
		// Need to support numerical ones.
		valStr := fmt.Sprintf("%v", val)
		if passId == 2 { // In passId 1, we would have featSoftWhitelist as empty.
			if !(featSoftWhitelist[key] != nil && featSoftWhitelist[key][valStr]) {
				continue
			}
		}
		// This increments the frequency of #users by 1.
		UpdateLevel3CatRatioDist(featureMetrics, key, valStr, "#users", 1.0)
	}
}

func UpdateLevel1CatFreqDist(freqDist *Level1CatFreqDist, key string, count int) {
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
	UpdateLevel1CatFreqDist(&level1FreqDist, key2, count)
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

func UpdateLevel1CatRatioDist(freqDist *Level1CatRatioDist, key string, ratio float64) {
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
	UpdateLevel1CatRatioDist(&level1RatioDist, key2, ratio)
	(*freqDist)[key1] = level1RatioDist
}

func UpdateLevel3CatRatioDist(freqDist *Level3CatRatioDist, key1, key2, key3 string, ratio float64) {
	if _, ok := (*freqDist)[key1]; !ok {
		(*freqDist)[key1] = make(Level2CatRatioDist)
	}
	level2RatioDist := (*freqDist)[key1]
	updateLevel2CatRatioDist(&level2RatioDist, key2, key3, ratio)
	(*freqDist)[key1] = level2RatioDist
}

func translateToWPMetrics(baseEvents, targetEvents, baseAndTargetEvents []PerEventProperties, featSoftWhitelist map[string]map[string]bool, passId int) WithinPeriodInsights {
	// TODO: Assert that in baseEvents, each user has only one row, not more. Same for targetEvents.
	baseMetrics := eventsToMetrics(baseEvents, featSoftWhitelist, passId)
	targetMetrics := eventsToMetrics(targetEvents, featSoftWhitelist, passId)
	baseAndTargetMetrics := eventsToMetrics(baseAndTargetEvents, featSoftWhitelist, passId)
	// TODO: commenting out conversions for now
	/*convMetrics := WithinPeriodRatioMetrics{}
	if passId == 2 {
		convMetrics = computeConversionMetrics(baseMetrics, targetMetrics, baseAndTargetMetrics)
	}*/
	wpInsights := WithinPeriodInsights{Base: baseMetrics, Target: targetMetrics,
		BaseAndTarget: baseAndTargetMetrics}
	return wpInsights
}

// func computeConversionMetrics(baseMetrics, targetMetrics, baseAndTargetMetrics WithinPeriodMetrics) WithinPeriodRatioMetrics {
// 	globalConvRatio := SmartDivide(float64(baseAndTargetMetrics.GlobalMetrics["#users"]), float64(baseMetrics.GlobalMetrics["#users"]))
// 	globalMetrics := Level1CatRatioDist{"ratio": globalConvRatio}
// 	featureMetrics := make(Level2CatRatioDist)
// 	for key, valBaseCounts := range baseMetrics.FeatureMetrics {
// 		if _, ok := baseAndTargetMetrics.FeatureMetrics[key]; !ok {
// 			continue
// 		}
// 		featureMetrics[key] = make(Level1CatRatioDist)
// 		valBaseAndTargetCounts := baseAndTargetMetrics.FeatureMetrics[key]
// 		for val, baseStats := range valBaseCounts {
// 			baseCount := baseStats["#users"]
// 			if _, ok := valBaseAndTargetCounts[val]; !ok {
// 				continue
// 			}
// 			baseAndTargetCount := valBaseAndTargetCounts[val]["#users"]
// 			convRate := SmartDivide(float64(baseAndTargetCount), float64(baseCount))
// 			featureMetrics[key][val] = convRate
// 		}
// 	}
// 	convMetrics := WithinPeriodRatioMetrics{GlobalMetrics: globalMetrics, FeatureMetrics: featureMetrics}
// 	return convMetrics
// }

func FindTopKCountThres(featureMetrics Level3CatRatioDist, k int) (int, int) {
	baseMetrics := Level3CatRatioDist{}
	targetMetrics := Level3CatRatioDist{}
	for key, valStats := range featureMetrics {
		if strings.HasPrefix(key, "s#") {
			baseMetrics[key] = valStats
		}
		if strings.HasPrefix(key, "t#") {
			targetMetrics[key] = valStats
		}
	}
	kthMaxCountBase, kthMaxCountTarget := int(0), int(0)
	if len(baseMetrics) > 0 {
		// var counts []int = nil
		var tempCounts []int = nil
		for _, valStats := range baseMetrics {
			for _, stats := range valStats {
				count := int(stats["#users"])
				// counts = append(counts, count)
				tempCounts = append(tempCounts, count)
			}
		}
		kthMaxCountBase = findKthMaxVal(tempCounts, k)
	}

	if len(targetMetrics) > 0 {
		// var counts []int = nil
		var tempCounts []int = nil
		for _, valStats := range targetMetrics {
			for _, stats := range valStats {
				count := int(stats["#users"])
				// counts = append(counts, count)
				tempCounts = append(tempCounts, count)
			}
		}
		kthMaxCountTarget = findKthMaxVal(tempCounts, k)
	}

	return kthMaxCountBase, kthMaxCountTarget
}

func selectTopKFeatures(wpInsights *WithinPeriodInsights, k int) {
	if k == -1 {
		return
	}
	numBase := len((*wpInsights).Base.FeatureMetrics)
	if k < numBase {
		kthMaxCountBase, _ := FindTopKCountThres((*wpInsights).Base.FeatureMetrics, k)
		totalUserCount := (*wpInsights).Base.GlobalMetrics["#users"]
		FilterFeatureCounts(&(*wpInsights).Base.FeatureMetrics, totalUserCount, kthMaxCountBase, 0)
	}
	numBaseAndTarget := len((*wpInsights).BaseAndTarget.FeatureMetrics)
	if k < numBaseAndTarget {
		kthMaxCountBase, kthMaxCountTarget := FindTopKCountThres((*wpInsights).BaseAndTarget.FeatureMetrics, k)
		totalUserCount := (*wpInsights).BaseAndTarget.GlobalMetrics["#users"]
		FilterFeatureCounts(&(*wpInsights).BaseAndTarget.FeatureMetrics, totalUserCount, kthMaxCountBase, kthMaxCountTarget)
	}
	numTarget := len((*wpInsights).Target.FeatureMetrics)
	if k < numTarget {
		_, kthMaxCountTarget := FindTopKCountThres((*wpInsights).Target.FeatureMetrics, k)
		totalUserCount := (*wpInsights).Target.GlobalMetrics["#users"]
		FilterFeatureCounts(&(*wpInsights).Target.FeatureMetrics, totalUserCount, 0, kthMaxCountTarget)
	}
}

func FilterFeatureCounts(featureMetrics *Level3CatRatioDist, totalCount int, kthMaxCountBase int, kthMaxCountTarget int) {
	baseMetrics := Level3CatRatioDist{}
	targetMetrics := Level3CatRatioDist{}
	for key, valStats := range *featureMetrics {
		if strings.HasPrefix(key, "s#") {
			baseMetrics[key] = valStats
		}
		if strings.HasPrefix(key, "t#") {
			targetMetrics[key] = valStats
		}
	}
	for key, valStats := range baseMetrics {
		retainKey := false
		othersFlag := false
		othersCount := 0
		for val, stats := range valStats {
			count := int(stats["#users"])
			if count < kthMaxCountBase {
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
	for key, valStats := range targetMetrics {
		retainKey := false
		othersFlag := false
		othersCount := 0
		for val, stats := range valStats {
			count := int(stats["#users"])
			if count < kthMaxCountTarget {
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

// func updateCountList(listToBeUpdated, updaterList []int64) {
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
	// TODO : janani uncomment this
	//convMetrics := ComputeCrossPeriodRatioMetrics(wpi1.Conversion, wpi2.Conversion)
	deltaRatioMetrics := ComputeDeltaRatioMetrics(wpi1, wpi2)
	jsdMetrics := ComputeJSDMetrics(wpi1, wpi2)
	cpInsights := CrossPeriodInsights{Base: baseMetrics, Target: targetMetrics, BaseAndTarget: baseAndTargetMetrics,
		DeltaRatio: deltaRatioMetrics, JSDivergence: jsdMetrics}
	return cpInsights, err
}

func ComputeDeltaRatioMetrics(wpi1, wpi2 WithinPeriodInsights) Level2CatRatioDist {
	drMetrics := make(Level2CatRatioDist)
	u1 := float64(wpi1.Base.GlobalMetrics["#users"])
	u2 := float64(wpi2.Base.GlobalMetrics["#users"])
	m1 := float64(wpi1.BaseAndTarget.GlobalMetrics["#users"])
	m2 := float64(wpi2.BaseAndTarget.GlobalMetrics["#users"])
	for key, valStats1 := range wpi1.Base.FeatureMetrics {
		drMetrics[key] = make(Level1CatRatioDist)
		for val := range valStats1 {
			f1 := wpi1.Base.FeatureMetrics[key][val]["#users"]
			f2 := wpi2.Base.FeatureMetrics[key][val]["#users"]
			fm1 := wpi1.BaseAndTarget.FeatureMetrics[key][val]["#users"]
			fm2 := wpi2.BaseAndTarget.FeatureMetrics[key][val]["#users"]
			crf1 := SmartDivide(float64(fm1), float64(f1))
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
func GetEventFileScanner(projectId int64, periodCode Period, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, insightGranularity string, isDownloaded bool) (*bufio.Scanner, error) {
	var err error
	localFile := true
	efTmpPath, efTmpName := diskManager.GetModelEventsFilePathAndName(projectId, periodCode.From, insightGranularity)
	// TODO: Change this to efTmpPath and efTmpName and write the code to download an events file from cloud to local.
	// We already have the logic to dump an events file from DB to local and from local to cloud, but not from cloud to local.
	// The following eventsFilePath will run fine if the cloud_path is a local one. But not if it's an actual remote cloud path.
	eventsFilePath := efTmpPath + efTmpName
	fmt.Println(eventsFilePath)
	_, err = T.OpenEventFileAndGetScanner(eventsFilePath)
	if err != nil {
		localFile = false
		deltaComputeLog.WithFields(log.Fields{"err": err,
			"eventsFilePath": eventsFilePath}).Error("Failed opening event file and getting scanner.")
	}
	if !isDownloaded || !localFile {
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
	scanner, err := T.OpenEventFileAndGetScanner(eventsFilePath)
	if err != nil {
		deltaComputeLog.WithFields(log.Fields{"err": err,
			"eventsFilePath": eventsFilePath}).Error("Failed opening event file and getting scanner.")
	}

	return scanner, err
}

// GetEventFileScanner Return handle to events file scanner
func GetChannelFileScanner(channel string, projectId int64, periodCode Period, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, insightGranularity string, isDownloaded bool) (*bufio.Scanner, error) {
	var err error
	localFile := true
	cfTmpPath, cfTmpName := diskManager.GetModelChannelFilePathAndName(channel, projectId, periodCode.From, insightGranularity)
	// TODO: Change this to cfTmpPath and cfTmpName and write the code to download an channel file from cloud to local.
	// We already have the logic to dump an channel file from DB to local and from local to cloud, but not from cloud to local.
	// The following channelFilePath will run fine if the cloud_path is a local one. But not if it's an actual remote cloud path.
	channelFilePath := cfTmpPath + cfTmpName
	fmt.Println(channelFilePath)
	_, err = T.OpenEventFileAndGetScanner(channelFilePath)
	if err != nil {
		localFile = false
		deltaComputeLog.WithFields(log.Fields{"err": err,
			"channelFilePath": channelFilePath}).Error("Failed opening event file and getting scanner.")
	}
	if !isDownloaded || !localFile {
		cfCloudPath, cfCloudName := (*cloudManager).GetModelChannelFilePathAndName(channel, projectId, periodCode.From, insightGranularity)
		deltaComputeLog.WithFields(log.Fields{"eventFileCloudPath": cfCloudPath,
			"channelFileCloudName": cfCloudName}).Info("Downloading " + channel + " file from cloud.")
		eReader, err := (*cloudManager).Get(cfCloudPath, cfCloudName)
		if err != nil {
			deltaComputeLog.WithFields(log.Fields{"err": err, "eventFilePath": cfCloudPath,
				"eventFileName": cfCloudName}).Error("Failed downloading " + channel + " file from cloud.")
			return nil, err
		}
		err = diskManager.Create(cfTmpPath, cfTmpName, eReader)
		if err != nil {
			deltaComputeLog.WithFields(log.Fields{"err": err, "eventFilePath": cfCloudPath,
				"eventFileName": cfCloudName}).Error("Failed downloading " + channel + " file from cloud.")
			return nil, err
		}
	}
	scanner, err := T.OpenEventFileAndGetScanner(channelFilePath)
	if err != nil {
		deltaComputeLog.WithFields(log.Fields{"err": err,
			"channelFilePath": channelFilePath}).Error("Failed opening event file and getting scanner.")
	}

	return scanner, err
}
