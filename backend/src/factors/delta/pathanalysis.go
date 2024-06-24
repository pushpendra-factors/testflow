package delta

import (
	"encoding/json"
	C "factors/config"
	P "factors/pattern"
	U "factors/util"
	"fmt"
	"sort"
	"strings"

	M "factors/model/model"

	"io"

	log "github.com/sirupsen/logrus"
)

const (
	NO_OF_EVENTS_AT_EACH_LEVEL = 10
	PAGE_VIEW_CATEGORY         = "Page Views"
	BUTTON_CLICKS_CATEGORY     = "Button Clicks"
	CRM_EVENTS_CATEGORY        = "CRM Events"
	SESSIONS_CATEGORY          = "Sessions"
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

func ToCountEvent(eventDetails P.CounterEventFormat, nameFilter string, filters []M.QueryProperty) (bool, error) {
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

func AppendExpandByCriteria(expandByProperty map[string][]PropertyEntityTypeMap, event P.CounterEventFormat, displayNamesFromDB map[string]string) string {
	page_views, crm_event, button_click, session_event := false, false, false, false
	_, exist := event.EventProperties["$is_page_view"]
	if !exist {
		page_views = true
	}
	hasPrefix := false
	for _, prefix := range U.CRMEventPrefixes {
		if strings.HasPrefix(event.EventName, prefix) {
			hasPrefix = true
		}
	}
	if hasPrefix == true {
		crm_event = false
	}
	_, exist = event.EventProperties["element_type"]
	if !exist {
		button_click = true
	}
	if !(event.EventName == "$session") {
		session_event = true
	}
	expandByProperties := make([]PropertyEntityTypeMap, 0)
	expandByPropertiesEvent := make([]PropertyEntityTypeMap, 0)
	expandByPropertiesUser := make([]PropertyEntityTypeMap, 0)
	properties, existInMap := expandByProperty[PAGE_VIEW_CATEGORY]
	if page_views == true && existInMap {
		expandByProperties = append(expandByProperties, properties...)
	}
	properties, existInMap = expandByProperty[CRM_EVENTS_CATEGORY]
	if crm_event == true && existInMap {
		expandByProperties = append(expandByProperties, properties...)
	}
	properties, existInMap = expandByProperty[BUTTON_CLICKS_CATEGORY]
	if button_click == true && existInMap {
		expandByProperties = append(expandByProperties, properties...)
	}
	properties, existInMap = expandByProperty[SESSIONS_CATEGORY]
	if session_event == true && existInMap {
		expandByProperties = append(expandByProperties, properties...)
	}
	properties, existInMap = expandByProperty[event.EventName]
	if existInMap {
		expandByProperties = append(expandByProperties, properties...)
	}
	for _, property := range expandByProperties {
		if property.Entity == "event" {
			expandByPropertiesEvent = append(expandByPropertiesEvent, property)
		}
		if property.Entity == "user" {
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
		if exists {
			valueString = fmt.Sprintf("%v", value)
		} else {
			valueString = "<nil>"
		}
		finalEventString = fmt.Sprintf("%s:%s-%v", finalEventString, property.Property, valueString)
	}
	for _, property := range expandByPropertiesUser {
		value, exists := event.UserProperties[property.Property]
		valueString := ""
		if exists {
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
	Entity   string
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
	data, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return nil
	}
	err = json.Unmarshal(data, &result)
	return result
}
