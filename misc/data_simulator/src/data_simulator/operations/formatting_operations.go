package operations

/*
This is to format the output in required format
*/

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

func FormatOutput(timeStamp int, userId string, event string, userAttributes map[string]string, eventAttributes map[string]string) string {
	var op EventOutput
	op.UserId = userId
	op.Event = event
	op.Timestamp = timeStamp
	op.UserAttributes = convertStringToInterface(userAttributes)
	op.EventAttributes = convertStringToInterface(eventAttributes)
	e, _ := json.Marshal(&op)
	return string(e)
}

func FormatUserData(userId string, attributes map[string]string) string {
	var op UserDataOutput
	op.UserId = userId
	op.UserAttributes = convertStringToInterface(attributes)
	e, _ := json.Marshal(&op)
	return string(e)
}

func convertStringToInterface(ip map[string]string) map[string]interface{} {
	op := make(map[string]interface{})
	for key, value := range ip {
		valueNum, err := strconv.Atoi(value)
		if err != nil {
			op[key] = value
		} else {
			op[key] = valueNum
		}
	}
	return op
}

func ConvertInterfaceToString(ip map[string]interface{}) map[string]string {
	op := make(map[string]string)
	for key, value := range ip {
		if reflect.TypeOf(value).Kind() == reflect.String {
			op[key] = value.(string)
		} else {
			op[key] = fmt.Sprintf("%v", value)
		}
	}
	return op
}
