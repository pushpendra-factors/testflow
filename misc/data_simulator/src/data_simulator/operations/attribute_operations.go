package operations

/*
File with all attribute level operations
*/

import(
	"sort"
	"data_simulator/config"
	"fmt"
	Log "data_simulator/logger"
	"data_simulator/utils"
	"strconv"
	"strings"
	"math/rand"
	"data_simulator/constants"
	"math"
)


type AttributeProbMap struct{
	Attributes_Order1 map[string]RangeMapMultiplierTuple
	Attributes_OrderOver1 map[string]map[string]RangeMapMultiplierTuple
}

func Convert1(input map[string]interface{}) map[string]float64{
	op := make(map[string]float64)
	for k, v := range input {
		op[k] = v.(float64)
	}
	return op
}

func Convert2(input map[interface{}]interface{}) map[string]float64{
	op := make(map[string]float64)
	for k, v := range input {
		op[k.(string)] = v.(float64)
	}
	return op
}

func SortAttributeMap(attributes []config.AttributeData){
	Log.Debug.Printf("Attributes before sorting: %v", attributes)
	sort.Slice(attributes, func(i,j int) bool {
		return attributes[i].Order_Level < attributes[j].Order_Level
	})
	Log.Debug.Printf("User Attributes After sorting: %v", attributes)
}

func GenerateProbablityMapForAttributes(attributes []config.AttributeData, tag string) (AttributeProbMap){
	var attributeProbMap AttributeProbMap
	attributeProbMap.Attributes_Order1 = make(map[string]RangeMapMultiplierTuple)
	attributeProbMap.Attributes_OrderOver1 = make(map[string]map[string]RangeMapMultiplierTuple)
	for _, element := range attributes {
		if(element.Order_Level == 1){
			attributeProbMap.Attributes_Order1[element.Key] = 
				ComputeRangeMap(Convert1(element.Values), fmt.Sprintf("Attributes-%s", element.Key))
		}
		if(element.Order_Level > 1){
			attributeProbMap.Attributes_OrderOver1[element.Key] = make(map[string]RangeMapMultiplierTuple)
			for key, value := range element.Values {
				attributeProbMap.Attributes_OrderOver1[element.Key][key] = 
					ComputeRangeMap(Convert2(value.(map[interface{}]interface{})),fmt.Sprintf("Attributes-%s-%s", element.Key, key))
			}
		}
	}
	Log.Debug.Printf("UserAttributes Probablity Map: \n %v", attributeProbMap)
	return attributeProbMap
}

func PickAttributes(attributes []config.AttributeData, attributeProbMap AttributeProbMap)(map[string]string){
// Compulsory choose each default attributes
// iterate over each attribute Eg: gender, age
// For orderLevel = 1 choose rand from probablity map directly
// For orderLevel = 2 choose the value of orderLevel = 1 attribute and go one level down 
// to get the actual value of the required attribute from the probablity map
// OrderLevel can be to any level
// but key is currently assumed to be only one and not multiple

// for custom attributes
// check if yes/no
// check how many attributes to pick (random again) - for now we will assume all
//append both and return
// have a global map for user to attributes map
	attr := make(map[string]string)
	for _, element := range attributes {
		if(element.Order_Level == 1) {
			attr[element.Key] = GetRandomValueWithProbablity(
				attributeProbMap.Attributes_Order1[element.Key], 
				fmt.Sprintf("UserAttribute:%s",element.Key))
		}
		if(element.Order_Level > 1) {
			attr[element.Key] = GetRandomValueWithProbablity(
				attributeProbMap.Attributes_OrderOver1[element.Key][attr[element.Dependency]], 
				fmt.Sprintf("UserAttribute:%s",element.Dependency))
		}
		if(element.Data_type == constants.NUMERICAL){
			attr[element.Key] = fmt.Sprintf("%v", GetValueFromRange(attr[element.Key]))
		}
			
	}
	return attr
}

func GetValueFromRange(value string) int {
	num := strings.Split(value, "-")
	if(len(num) != 2){
		Log.Error.Fatal("Incorrect numerical range", value)
	}
	start, _ := strconv.Atoi(num[0])
	end, _ := strconv.Atoi(num[1])
	return start + rand.Intn(end-start+1)
}

func ReassignProbablity(probMap map[string]valueBoolTuple, tag string) (map[string]float64){
	// Compute new probablities - Should be straight forward
	// Generate new range map
	// check 
	reassignedProbMap := make(map[string]float64)
	sum_changed := 0.0
	sum_unchanged := 0.0
	for item, element := range probMap {
		if(element.change == true){
			sum_changed += element.value
			reassignedProbMap[item] = element.value
		}
		if(element.change == false){
			sum_unchanged += element.value
		}
	}
	if(sum_changed >= 1.0){
		Log.Error.Fatal("Sum Probablities >= 1.0. Cannot Reassign ", tag)
	}
	remainingProb := 1.0 - sum_changed
	for item, element := range probMap {
		if(element.change == false){
			reassignedProbMap[item] = math.Round(((remainingProb/sum_unchanged) * element.value)*10000)/10000
		}
	}
	return reassignedProbMap
}

func PreloadEventAttributes(probMap ProbMap, segmentConfig config.UserSegmentV2, segmentProbMap SegmentProbMap)(map[string]map[string]string){
	eventToEventAttributes := make(map[string]map[string]string)
	for item, _ := range segmentConfig.Event_probablity_map.Correlation_matrix.Events{
		eventToEventAttributes[item] = GetEventAttributes(
			probMap,
			segmentProbMap,
			segmentConfig,
			item)
	}
	for item, _ := range segmentConfig.Event_probablity_map.Independent_events{
		eventToEventAttributes[item] = GetEventAttributes(
			probMap,
			segmentProbMap,
			segmentConfig,
			item)
	}
	return eventToEventAttributes
}

func GetUserAttributes(probMap ProbMap, segmentProbMap SegmentProbMap, segmentConfig config.UserSegmentV2) map[string]string{
	userAttr := make(map[string]string)
	userAttr = PickAttributes(
		segmentConfig.User_attributes.Default,
		segmentProbMap.defaultUserAttrProbMap)
	if(AddCustomUserAttributeOrNot(probMap)){
		userAttr = utils.AppendMaps(userAttr, PickAttributes(
			segmentConfig.User_attributes.Custom,
			segmentProbMap.customUserAttrProbMap))
	}
	return userAttr
}

func SetUserAttributes(segmentProbMap SegmentProbMap, segmentConfig config.UserSegmentV2, userId string) map[string]string{	
	attr := make(map[string]string)
	userAttributeMutex.Lock()
	if(segmentConfig.Set_attributes == true){
		attr = segmentProbMap.UserToUserAttributeMap[userId]
	}
	userAttributeMutex.Unlock()
	return attr 	
}

func GetEventAttributes(probMap ProbMap, segmentProbMap SegmentProbMap, segmentConfig config.UserSegmentV2,eventName string) map[string]string{
	eventAttr := make(map[string]string)
	eventAttr = PickPredefinedAttributes(segmentProbMap.predefinedEventAttrProbMap[eventName])
	eventAttr = utils.AppendMaps(eventAttr, PickAttributes(
		segmentConfig.Event_attributes.Default,
		segmentProbMap.defaultEventAttrProbMap))
	if(AddCustomEventAttributeOrNot(probMap)){
		eventAttr = utils.AppendMaps(eventAttr, PickAttributes(
			segmentConfig.Event_attributes.Custom,
			segmentProbMap.customEventAttrProbMap))
	}
	return eventAttr
}

func SetEventAttributes(segmentProbMap SegmentProbMap, segmentConfig config.UserSegmentV2, event string) map[string]string{
	if(segmentConfig.Set_attributes == true){
		return segmentProbMap.EventToEventAttributeMap[event]
	}
	return make(map[string]string)
}

func PickPredefinedAttributes(predefined map[string]RangeMapMultiplierTuple) map[string]string{
	op := make(map[string]string)
	for item, element := range predefined {
		op[item] = GetRandomValueWithProbablity(element, item)
	}
	return op
}

