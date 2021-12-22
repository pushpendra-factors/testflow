package operations

/*
This file contains all rand based operations.
Eg: Get random event with probablity, Get random attribute set based on probablity
*/

import (
	"data_simulator/config"
	"data_simulator/constants"
	Log "data_simulator/logger"
	"data_simulator/utils"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
)

type valueBoolTuple struct {
	value  float64
	change bool
}

func SeedUserOrNot(probMap ProbMap) bool {
	return DecideYesOrNo(probMap.yesOrNoProbMap.SeedNewUser, "Seed New User")
}

func AddCustomEventAttributeOrNot(probMap ProbMap) bool {
	return DecideYesOrNo(probMap.yesOrNoProbMap.AddCustomEventAttribute, "Add custom Event Attribute")
}

func AddCustomUserAttributeOrNot(probMap ProbMap) bool {
	return DecideYesOrNo(probMap.yesOrNoProbMap.AddCustomUserAttribute, "Add Custom User Attribute")
}

func BringExistingUserOrNot(probMap ProbMap) bool {
	return DecideYesOrNo(probMap.yesOrNoProbMap.BringExistingUser, "Bring Existing User")
}

func DecideYesOrNo(rangeMap RangeMapMultiplierTuple, tag string) bool {
	yesOrNo := GetRandomValueWithProbablity(rangeMap, tag)
	if yesOrNo == constants.YES {
		return true
	}
	return false
}

func GetRandomSegment() string {
	segmentKeys := reflect.ValueOf(config.ConfigV2.User_segments).MapKeys()
	seg := (segmentKeys[rand.Intn(len(segmentKeys))].Interface()).(string)
	return seg
}

func GetRandomActivity(probMap SegmentProbMap) string {
	return GetRandomValueWithProbablity(probMap.activityProbMap, "Activity")
}

func GetRandomEvent(probMap SegmentProbMap) string {
	return GetRandomValueWithProbablity(probMap.eventProbMap, "Event")
}

func GetRandomEventWithCorrelation(lastKnownGoodState *string, probMap SegmentProbMap, userAttributes map[string]string, segmentConfig config.UserSegmentV2) (string, int) {

	if *lastKnownGoodState == "" {
		*lastKnownGoodState = GetRandomValueWithProbablity(probMap.SeedEventsProbMap, "Seed-Events")
		return *lastKnownGoodState, 0
	}
	eventCorrelationRangeMap, realTimeWaitMap := GetAttributesBasedProbablityMap(lastKnownGoodState, probMap, userAttributes)
	*lastKnownGoodState = GetRandomValueWithProbablity(
		eventCorrelationRangeMap, fmt.Sprintf("EventWithCorrelation:%s", *lastKnownGoodState))
	if realTimeWaitMap != nil {
		return *lastKnownGoodState, realTimeWaitMap[*lastKnownGoodState]
	} else {
		return *lastKnownGoodState, 0
	}
}

func GetAttributesBasedProbablityMap(lastKnownGoodState *string, probMap SegmentProbMap, userAttributes map[string]string) (RangeMapMultiplierTuple, map[string]int) {
	// This method compares the user attributes against the attribute probablity map
	// Returns the adjusted probablity map
	newEventCorrelationMap := make(map[string]valueBoolTuple)
	realTimeWaitMap := make(map[string]int)
	if probMap.EventAttributeRule[*lastKnownGoodState] != nil {
		for item, element := range probMap.EventCorrelationMapNormalized[*lastKnownGoodState] {
			value, state, realTimeWait := IsProbablityOverridden(probMap.EventAttributeRule[*lastKnownGoodState][item], userAttributes)
			if state == true {
				newEventCorrelationMap[item] = valueBoolTuple{value, state}
			} else {
				newEventCorrelationMap[item] = valueBoolTuple{element, state}
			}
			realTimeWaitMap[item] = realTimeWait
		}
	} else {
		return probMap.EventCorrelationProbMap[*lastKnownGoodState], nil
	}
	return ComputeRangeMap(ReassignProbablity(newEventCorrelationMap, *lastKnownGoodState), fmt.Sprintf("%s-%s-%s", "Reassignment", "Event-Correlation", *lastKnownGoodState)), realTimeWaitMap
}

func IsProbablityOverridden(attributeRule config.AttributeRule, userAttributes map[string]string) (float64, bool, int) {
	// If there are multiple matches from attribute probablity map
	// Pick the highest
	highest := 0.0
	probablityModified := false
	for _, element1 := range attributeRule.Attribute_weights {
		attributesMatch := true
		for item2, element2 := range element1.Attributes {
			if !utils.Contains(element2, userAttributes[item2]) {
				attributesMatch = false
				break
			}
		}
		if attributesMatch == true {
			if highest < element1.Probablity {
				highest = element1.Probablity
				probablityModified = true
			}
		}
	}
	return highest, probablityModified, attributeRule.Real_time_wait
}

func GetEventDecorators(eventName string, segmentProbMap SegmentProbMap) map[string]string {
	decorators := make(map[string]string)
	if segmentProbMap.EventDecoratorProbMap[eventName] != nil {
		for item, element := range segmentProbMap.EventDecoratorProbMap[eventName] {
			decorators[item] = GetRandomValueWithProbablity(element, fmt.Sprintf("Decorator-%s-%s", eventName, item))
		}
	}
	if strings.HasPrefix(eventName, "www") {
		decorators["$page_url"] = strings.Split(eventName, "?")[0]
	}
	return decorators
}

func GetUserDecorators(eventName string, segmentProbMap SegmentProbMap) map[string]string {
	decorators := make(map[string]string)
	if segmentProbMap.UserDecoratorProbMap[eventName] != nil {
		for item, element := range segmentProbMap.UserDecoratorProbMap[eventName] {
			decorators[item] = GetRandomValueWithProbablity(element, fmt.Sprintf("Decorator-%s-%s", eventName, item))
		}
	}
	return decorators
}

func GetRandomValueWithProbablity(rangeMap RangeMapMultiplierTuple, tag string) string {
	r := rand.Intn(rangeMap.multiplier)
	value, state := rangeMap.probRangeMap.Get(r)
	if state == false {
		Log.Error.Fatal(fmt.Sprintf("Tag: %s, RangeMap: Key not found %v", tag, r))
	}
	return value
}

func PickFromExistingUsers(users map[string]map[string]string) (string, map[string]string) {
	keys := utils.GetMapKeys(users)
	if len(keys) != 0 {
		r := rand.Intn(len(keys))
		return keys[r], users[keys[r]]
	}
	return "", nil
}
