package event_match

import (
	U "factors/util"
	"fmt"
	"strings"
	"strconv"
	M "factors/model/model"
	log "github.com/sirupsen/logrus"
)
type PropertiesMode string
type BooleanOperator string

var USER_PROPERTIES_MODE PropertiesMode = "user"
var EVENT_PROPERTIES_MODE PropertiesMode = "event"

// EventFilterCriterion represents a filter-criterion that an event has to follow.
// It is composed of a feature key, its allowed values, and an equality flag.
// If equality is True, it translates to "key in values", otherwise, "key not in values".
// In other words, if equality flag is True, then key "can take any of the values" (OR),
// and if it is False, key "cannot take any of the values" (AND).
type EventFilterCriterion struct {
	Id             int    `json:"id"`
	Key            string `json:"key"`
	Type           string
	Values         []OperatorValueTuple
	PropertiesMode PropertiesMode `json:"propmode"`
}

type OperatorValueTuple struct {
	Operator  string
	Value     string
	LogicalOp string
}

// EventCriterion abstracts a single event-criterion that users/events have to adhere to.
// More specifically, it contains an event name, equality flag, and a list of filter-criterion.
type EventCriterion struct {
	Id                  int                    `json:"id"`
	Name                string                 `json:"en"`
	EqualityFlag        bool                   `json:"eq"`
	FilterCriterionList []EventFilterCriterion `json:"filters"`
}

// EventsCriteria represents one period's event-criteria that users/events have to adhere to.
// It is basically an AND or an OR of multiple event-criterion.
type EventsCriteria struct {
	Id                 int              `json:"id"`
	Operator           BooleanOperator  `json:"op"`
	EventCriterionList []EventCriterion `json:"events"`
}

func EventMatchesCriterion(eventName string, userProperties, eventProperties  map[string]interface{}, eventCriterion EventCriterion) bool {
	// TODO: Match event filters as well.
	nameMatchFlag := eventCriterion.EqualityFlag == (eventName == eventCriterion.Name)
	if !nameMatchFlag {
		return false
	}
	filterMatchFlag := EventMatchesFilterCriterionList(userProperties, eventProperties, eventCriterion.FilterCriterionList)
	return filterMatchFlag
}

func EventMatchesFilterCriterionList(userProperties, eventProperties  map[string]interface{},  filterCriterionList []EventFilterCriterion) bool {
	for _, fc := range filterCriterionList {
		// Today we dont support OR across filters. So retaining it this way. Its always a AND
		if !eventMatchesFilterCriterion(userProperties, eventProperties, fc) { // "AND" logic: If even a single filter fails, return False.
			return false
		}
	}
	return true
}

func eventMatchesFilterCriterion(userProperties, eventProperties  map[string]interface{},  filterCriterion EventFilterCriterion) bool {
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
		props = userProperties
	} else if mode == EVENT_PROPERTIES_MODE {
		props = eventProperties
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
		if value.Operator == M.EqualsOpStr || value.Operator == M.NotEqualOpStr {
			equalsCount++
		}
		if value.Operator == M.ContainsOpStr || value.Operator == M.NotContainsOpStr {
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
	if andCount == 1 && orCount >= 1 && (filterValues[0].Operator == M.NotEqualOpStr || filterValues[0].Operator == M.NotContainsOpStr) {
		//A != b OR A != c	wrong
		return false
		// "Multi select filter with not equals"
	}
	// TODO What to do if there is a misclassification
	results := make(map[int]bool)
	propertyValue := fmt.Sprintf("%v", eventPropValue)

	for i, value := range filterValues {
		if value.Value == M.PropertyValueNone {
			results[i] = handleNoneCase(propertyValue, isPresentEventPropValue, value.Operator)
			continue
		}
		if value.Operator == M.EqualsOpStr {
			results[i] = strings.EqualFold(propertyValue, value.Value)
		}
		if value.Operator == M.NotEqualOpStr {
			results[i] = !strings.EqualFold(propertyValue, value.Value)
		}
		if value.Operator == M.ContainsOpStr {
			results[i] = strings.Contains(strings.ToLower(propertyValue), strings.ToLower(value.Value))
		}
		if value.Operator == M.NotContainsOpStr {
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
	if operator == M.EqualsOpStr || operator == M.ContainsOpStr {
		return !isPresentEventPropValue || eventPropValue == "$none"
	}
	if operator == M.NotEqualOpStr || operator == M.NotContainsOpStr {
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
		if value.Operator == M.EqualsOpStr {
			results[i] = eventPropertyValue == filterValue
		}
		if value.Operator == M.NotEqualOpStr {
			results[i] = eventPropertyValue != filterValue
		}
		if value.Operator == M.GreaterThanOpStr {
			results[i] = eventPropertyValue > filterValue
		}
		if value.Operator == M.LesserThanOpStr {
			results[i] = eventPropertyValue < filterValue
		}
		if value.Operator == M.GreaterThanOrEqualOpStr {
			results[i] = eventPropertyValue >= filterValue
		}
		if value.Operator == M.LesserThanOrEqualOpStr {
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
		if value.Operator == M.EqualsOpStr {
			dateTimeFilter, err := M.DecodeDateTimePropertyValue(value.Value)
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

func MapFilterProperties(qp []M.QueryProperty) []EventFilterCriterion {
	filters := make(map[string]EventFilterCriterion)
	for _, prop := range qp {
		filterProp := EventFilterCriterion{
			Key: prop.Property,
		}
		filterProp.Type = prop.Type
		if prop.Entity == "user" || prop.Entity == "user_g" {
			filterProp.PropertiesMode = "user"
		} else if prop.Entity == "event" {
			filterProp.PropertiesMode = "event"
		} else {
			log.Error("Incorrect entity type")
			return nil
		}
		keyString := fmt.Sprintf("%s-%s", prop.Entity, prop.Property)
		propertyInMap, exists := filters[keyString]
		var values []OperatorValueTuple
		if !exists {
			values = make([]OperatorValueTuple, 0)
		} else {
			values = propertyInMap.Values
		}
		values = append(values, OperatorValueTuple{
			Operator:  prop.Operator,
			Value:     prop.Value,
			LogicalOp: prop.LogicalOp,
		})
		filterProp.Values = values
		filters[keyString] = filterProp
	}
	criterias := make([]EventFilterCriterion, 0)
	for _, criteria := range filters {
		criterias = append(criterias, criteria)
	}
	return criterias
}