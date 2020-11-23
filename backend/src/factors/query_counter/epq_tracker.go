package querycounter

import (
	M "factors/model"
	P "factors/pattern"
	"factors/util"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// EventsPropertiesQueryTracker tracks if the set of events and filters are satisfied, based on type of events condition.
type EventsPropertiesQueryTracker struct {
	eventsWithProperties []M.QueryEventWithProperties
	// all, any or each.
	eventsCondition string

	// private variables.
	eventWithPropertyFound []bool
}

func NewEventPropertiesQueryTracker(
	eventsWithProperties []M.QueryEventWithProperties,
	eventsCondition string) (*EventsPropertiesQueryTracker, error) {
	// Checking for errors.
	eLen := len(eventsWithProperties)
	if eLen == 0 {
		return nil, fmt.Errorf("No events to track")
	}
	if eventsCondition != M.EventCondAnyGivenEvent &&
		eventsCondition != M.EventCondAllGivenEvent &&
		eventsCondition != M.EventCondEachGivenEvent {
		return nil, fmt.Errorf(fmt.Sprintf(
			"Unknown  condition %s in %s", eventsCondition, eventsWithProperties))
	}

	// Initialize.
	var epqTracker EventsPropertiesQueryTracker
	epqTracker.eventsWithProperties = eventsWithProperties
	epqTracker.eventsCondition = eventsCondition
	epqTracker.eventWithPropertyFound = make([]bool, eLen)
	for i, _ := range epqTracker.eventsWithProperties {
		epqTracker.eventWithPropertyFound[i] = false
	}
	return &epqTracker, nil
}

func (epqTracker *EventsPropertiesQueryTracker) Reset() {
	for i, _ := range epqTracker.eventsWithProperties {
		epqTracker.eventWithPropertyFound[i] = false
	}
}

// Returns if query tracking is completed.
// Warning: If the given query track is completed, it might keep returning true, based on old state.
// Responsible of the caller to reset the state once track is complete.
func (epqTracker *EventsPropertiesQueryTracker) IsTrackCompleteOnData(eventDetails *P.CounterEventFormat) bool {
	for i, ewp := range epqTracker.eventsWithProperties {
		if ewp.Name == eventDetails.EventName {
			if isAllPropertyFiltersMatched(ewp.Properties, eventDetails) {
				epqTracker.eventWithPropertyFound[i] = true
				if epqTracker.eventsCondition == M.EventCondEachGivenEvent {
					return true
				}
			}
		}
	}
	if epqTracker.eventsCondition == M.EventCondAnyGivenEvent {
		tracked := false
		for _, trackedI := range epqTracker.eventWithPropertyFound {
			tracked = tracked || trackedI
		}
		return tracked
	}
	if epqTracker.eventsCondition == M.EventCondAllGivenEvent {
		tracked := true
		for _, trackedI := range epqTracker.eventWithPropertyFound {
			tracked = tracked && trackedI
		}
		return tracked
	}
	return false
}

func isAllPropertyFiltersMatched(propertiesToMatch []M.QueryProperty, eventDetails *P.CounterEventFormat) bool {
	// By default we assume it is an AND operator on all the filter conditions. property.LogicalOp
	// is also currently default AND in normal queries too.
	for _, ptm := range propertiesToMatch {
		if string(ptm.Value) == M.PropertyValueNone {
			// Not handling $none values yet.
			continue
		}
		var propertiesMap map[string]interface{}

		if ptm.Entity == M.PropertyEntityUser {
			propertiesMap = eventDetails.UserProperties

		} else if ptm.Entity == M.PropertyEntityEvent {
			propertiesMap = eventDetails.EventProperties
		}

		dataVal, found := propertiesMap[ptm.Property]
		if !found {
			return false
		}
		if ptm.Type == util.PropertyTypeDateTime {
			queryValueDateTime, err := M.DecodeDateTimePropertyValue(ptm.Value)
			if err != nil {
				log.WithError(err).Error("Failed reading timestamp on datetime request.")
				return false
			}
			dataValTimestamp := int64(util.SafeConvertToFloat64(dataVal))
			if !(dataValTimestamp >= queryValueDateTime.From && dataValTimestamp <= queryValueDateTime.To) {
				return false
			}
		} else if ptm.Type == util.PropertyTypeNumerical {
			queryValFloat64 := util.SafeConvertToFloat64(ptm.Value)
			dataValFloat64 := util.SafeConvertToFloat64(dataVal)
			if ptm.Operator == M.EqualsOpStr && dataValFloat64 != queryValFloat64 {
				return false
			} else if ptm.Operator == M.NotEqualOpStr && dataValFloat64 == queryValFloat64 {
				return false
			} else if ptm.Operator == M.NotEqualOpStr && dataValFloat64 == queryValFloat64 {
				return false
			} else if ptm.Operator == M.GreaterThanOpStr && dataValFloat64 <= queryValFloat64 {
				return false
			} else if ptm.Operator == M.GreaterThanOrEqualOpStr && dataValFloat64 < queryValFloat64 {
				return false
			} else if ptm.Operator == M.LesserThanOpStr && dataValFloat64 >= queryValFloat64 {
				return false
			} else if ptm.Operator == M.LesserThanOrEqualOpStr && dataValFloat64 > queryValFloat64 {
				return false
			} else {
				log.Error(fmt.Sprintf("Unknown numerical operator %s. No match", ptm.Operator))
				return false
			}
		} else {
			// Default is categorical.
			queryValStr, err := util.GetValueAsString(ptm.Value)
			if err != nil {
				log.WithError(err).Error("Failed converting property value to string.")
				return false
			}
			dataValStr, err := util.GetValueAsString(dataVal)
			if err != nil {
				log.WithError(err).Error("Failed converting property value to string.")
				return false
			}
			if ptm.Operator == M.EqualsOpStr && dataValStr != queryValStr {
				return false
			} else if ptm.Operator == M.NotEqualOpStr && dataValStr == queryValStr {
				return false
			} else if ptm.Operator == M.ContainsOpStr && !strings.Contains(dataValStr, queryValStr) {
				return false
			} else if ptm.Operator == M.NotContainsOpStr && strings.Contains(dataValStr, queryValStr) {
				return false
			} else {
				log.Error(fmt.Sprintf("Unknown numerical operator %s. No match", ptm.Operator))
				return false
			}
		}
	}
	return true
}
