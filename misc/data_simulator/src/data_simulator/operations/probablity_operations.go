package operations

/*
This file contains methods for doing all probablity based operations
Eg: Precomputing map for each event/activity/attributes
*/

import(
	"data_simulator/utils"
	"data_simulator/config"
	"math"
	Log "data_simulator/logger"
	"fmt"
	"strings"
	"reflect"
	"data_simulator/constants"
)

type RangeMapMultiplierTuple struct {
	probRangeMap utils.RangeMap
	multiplier int
}

type SegmentProbMap struct {
	EventCorrelationMapNormalized map[string]map[string]float64
	EventCorrelationProbMap map[string]RangeMapMultiplierTuple
	EventAttributeRule map[string]map[string]config.AttributeRule
	eventProbMap RangeMapMultiplierTuple
	activityProbMap RangeMapMultiplierTuple
	defaultUserAttrProbMap AttributeProbMap
	customUserAttrProbMap AttributeProbMap
	defaultEventAttrProbMap AttributeProbMap
	customEventAttrProbMap AttributeProbMap
	predefinedEventAttrProbMap map[string]map[string]RangeMapMultiplierTuple
	UserToUserAttributeMap map[string]map[string]string
	EventToEventAttributeMap map[string]map[string]string
	EventDecoratorProbMap map[string]map[string]RangeMapMultiplierTuple
	SeedEventsProbMap RangeMapMultiplierTuple
}

type YesOrNoProbablityMap struct {
	SeedNewUser RangeMapMultiplierTuple
	AddCustomEventAttribute RangeMapMultiplierTuple
	AddCustomUserAttribute RangeMapMultiplierTuple
	BringExistingUser RangeMapMultiplierTuple
}
type ProbMap struct {
	yesOrNoProbMap YesOrNoProbablityMap
	segmentProbMap map[string]SegmentProbMap
}

func ComputeYesOrNoProbablityMap(trueProb float64, tag string)(RangeMapMultiplierTuple)  {
	probMap := make(map[string]float64)
	probMap[constants.YES] = trueProb
	probMap[constants.NO] = (1.0 - trueProb)
	return ComputeRangeMap(probMap, fmt.Sprintf("%s-%s", "YesOrNo", tag))
}

func ComputeRangeMap(probMap map[string]float64, tag string) (RangeMapMultiplierTuple) {

	min := 1.0
	sum := 0.0
	//TODO call this from util once you find a way to iterate values
	for _, element := range probMap {
		sum += element
		if element < min && element != 0{
			min = element
		}
	}

	multiplier := 1.0
	for {
		multiplier = multiplier * 10.0
		temp := (min * multiplier)-(math.Floor(min * multiplier));
		if(temp == 0.0){
			break
		}
	}

	// This is a hack to handle triming issues that is happening in reassign method
	if(!(int(math.Round(sum * multiplier)) >= int(multiplier-5) && int(math.Round(sum * multiplier)) <= int(multiplier+5))){
		Log.Error.Fatal("Probablity Sum != 1 ", tag)
	}

	start := 0
	probRangeMap := utils.RangeMap{}
	for item,element := range probMap {
		rangeEnd := int(math.Round(element * multiplier))
		probRangeMap.Keys = append(probRangeMap.Keys,utils.Range{ start, start+rangeEnd-1 })
		probRangeMap.Values = append(probRangeMap.Values, item)
		start = start + rangeEnd
	}

	return RangeMapMultiplierTuple{ probRangeMap, int(math.Round(sum * multiplier))}
}

func PreComputeRangeMap(segment config.UserSegmentV2) (SegmentProbMap) {

	var probMap SegmentProbMap
	probMap.EventCorrelationProbMap = make(map[string]RangeMapMultiplierTuple)
	probMap.EventAttributeRule = make(map[string]map[string]config.AttributeRule)
	probMap.EventCorrelationMapNormalized = make(map[string]map[string]float64)
	probMap.predefinedEventAttrProbMap = make(map[string]map[string]RangeMapMultiplierTuple)
	probMap.EventDecoratorProbMap = make(map[string]map[string]RangeMapMultiplierTuple)
	probMap.SeedEventsProbMap = ComputeRangeMap(segment.Event_probablity_map.Correlation_matrix.Seed_events, "Seed_events")
	for item1, element1 := range segment.Event_probablity_map.Correlation_matrix.Events {
		eventCorrelations := make(map[string]float64)
		probMap.EventAttributeRule[item1] = make(map[string]config.AttributeRule)
		for item2, element2 := range element1 {
			if(reflect.TypeOf(element2).Kind() == reflect.Float64){
				eventCorrelations[item2] = element2.(float64)
			}
			if(reflect.TypeOf(element2).Kind() == reflect.String && strings.HasPrefix(element2.(string), "RULE")){
				eventCorrelations[item2] = segment.Rules[element2.(string)].Overall_probablity
				probMap.EventAttributeRule[item1][item2] = segment.Rules[element2.(string)]
			}
		}
		probMap.EventCorrelationMapNormalized[item1] = eventCorrelations
		probMap.EventCorrelationProbMap[item1] = ComputeRangeMap(eventCorrelations, fmt.Sprintf("%s-%s","Event-Correlation",item1))
	}

	events := make(map[string]float64)
	sum := 0.0
	if segment.Event_probablity_map.Independent_events != nil {
		for item, element := range segment.Event_probablity_map.Independent_events {
			sum += element
			events[item] = element
		}
	}

	events[constants.EVENTCORRELATION] = (1.0 - sum)
	probMap.eventProbMap = ComputeRangeMap(events, "Event")
	probMap.activityProbMap = ComputeRangeMap(segment.Activity_probablity_map, "Actiivity")

	for item1, element1 := range segment.Event_attributes.Predefined{
		probMap.predefinedEventAttrProbMap[item1] = make(map[string]RangeMapMultiplierTuple)
		for item2, element2 := range element1 {
			probMap.predefinedEventAttrProbMap[item1][item2] = 
				ComputeRangeMap(element2, fmt.Sprintf("%s-%s-%s","PredefinedMap",item1, item2))
		}
	}

	for item1, element1 := range segment.Event_decorators{
		probMap.EventDecoratorProbMap[item1] = make(map[string]RangeMapMultiplierTuple)
		for item2, element2 := range element1 {
			probMap.EventDecoratorProbMap[item1][item2] = 
				ComputeRangeMap(element2, fmt.Sprintf("%s-%s-%s","EventDecorator",item1, item2))
		}
	}

	probMap.defaultUserAttrProbMap, probMap.customUserAttrProbMap = 
		PreComputeUserAttributeProbMap(segment.User_attributes)
	probMap.defaultEventAttrProbMap, probMap.customEventAttrProbMap =
		PreComputeEventAttributeProbMap(segment.Event_attributes)
	return probMap
}

func PreComputeUserAttributeProbMap(userAttributes config.UserAttributes)(AttributeProbMap, AttributeProbMap) {
	SortAttributeMap(userAttributes.Default)
	SortAttributeMap(userAttributes.Custom)
	d := GenerateProbablityMapForAttributes(
		userAttributes.Default, "User-Default")
	c := GenerateProbablityMapForAttributes(
			userAttributes.Custom, "User-Custom")
	return d, c
}

func PreComputeEventAttributeProbMap(eventAttributes config.EventAttributes)(AttributeProbMap, AttributeProbMap) {
	SortAttributeMap(eventAttributes.Default)
	SortAttributeMap(eventAttributes.Custom)
	d := GenerateProbablityMapForAttributes(
		eventAttributes.Default, "Event-Default")
	c := GenerateProbablityMapForAttributes(
		eventAttributes.Custom, "Event-Custom")
	return d, c
}