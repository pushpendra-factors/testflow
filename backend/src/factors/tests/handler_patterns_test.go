package tests

import (
	H "factors/handler"
	M "factors/model"
	P "factors/pattern"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertEqualConstraints(
	t *testing.T,
	expectedConstraints P.EventConstraints,
	actualConstraints P.EventConstraints) {

	// Check event properties
	expectedNumericMap := make(map[string]P.NumericConstraint)
	for _, nc := range expectedConstraints.EPNumericConstraints {
		expectedNumericMap[nc.PropertyName] = nc
	}
	expectedCategoricalMap := make(map[string]P.CategoricalConstraint)
	for _, cc := range expectedConstraints.EPCategoricalConstraints {
		expectedCategoricalMap[cc.PropertyName] = cc
	}
	actualNumericMap := make(map[string]P.NumericConstraint)
	for _, nc := range actualConstraints.EPNumericConstraints {
		actualNumericMap[nc.PropertyName] = nc
	}
	actualCategoricalMap := make(map[string]P.CategoricalConstraint)
	for _, cc := range actualConstraints.EPCategoricalConstraints {
		actualCategoricalMap[cc.PropertyName] = cc
	}
	assert.Equal(t, expectedNumericMap, actualNumericMap)
	assert.Equal(t, expectedCategoricalMap, actualCategoricalMap)

	// Check user properties
	expectedNumericMap = make(map[string]P.NumericConstraint)
	for _, nc := range expectedConstraints.UPNumericConstraints {
		expectedNumericMap[nc.PropertyName] = nc
	}
	expectedCategoricalMap = make(map[string]P.CategoricalConstraint)
	for _, cc := range expectedConstraints.UPCategoricalConstraints {
		expectedCategoricalMap[cc.PropertyName] = cc
	}
	actualNumericMap = make(map[string]P.NumericConstraint)
	for _, nc := range actualConstraints.UPNumericConstraints {
		actualNumericMap[nc.PropertyName] = nc
	}
	actualCategoricalMap = make(map[string]P.CategoricalConstraint)
	for _, cc := range actualConstraints.UPCategoricalConstraints {
		actualCategoricalMap[cc.PropertyName] = cc
	}
	assert.Equal(t, expectedNumericMap, actualNumericMap)
	assert.Equal(t, expectedCategoricalMap, actualCategoricalMap)
}

func TestParseFactorQuery(t *testing.T) {
	// No events.
	var query = make(map[string]interface{})
	startEvent, startEventConstraints, endEvent, endEventConstraints, _, err := H.ParseFactorQuery(query)
	assert.NotNil(t, err)
	assert.Equal(t, startEvent, "")
	assert.Nil(t, startEventConstraints)
	assert.Equal(t, endEvent, "")
	assert.Nil(t, endEventConstraints)

	// Only end event.
	event1 := make(map[string]interface{})
	event1["name"] = "endEvent"
	event1["properties"] = []interface{}{}
	query["eventsWithProperties"] = []interface{}{event1}
	query["queryType"] = M.QueryTypeUniqueUsers
	startEvent, startEventConstraints, endEvent, endEventConstraints, _, err = H.ParseFactorQuery(query)
	assert.Nil(t, err)
	assert.Equal(t, startEvent, "")
	assert.Nil(t, startEventConstraints)
	assert.Equal(t, endEvent, "endEvent")
	assert.Equal(t, *endEventConstraints, P.EventConstraints{
		EPNumericConstraints:     []P.NumericConstraint{},
		EPCategoricalConstraints: []P.CategoricalConstraint{},
		UPNumericConstraints:     []P.NumericConstraint{},
		UPCategoricalConstraints: []P.CategoricalConstraint{},
	})

	// Start and end events with properties.
	event1 = make(map[string]interface{})
	event1["name"] = "startEvent"
	// Integer equality
	property1 := make(map[string]interface{})
	property1["property"] = "property1"
	property1["type"] = "numerical"
	property1["value"] = 5.0
	property1["operator"] = "equals"
	// String equality
	property2 := make(map[string]interface{})
	property2["property"] = "property2"
	property2["type"] = "categorical"
	property2["value"] = "property2Value"
	property2["operator"] = "equals"
	event1["properties"] = []interface{}{property1, property2}
	// user properties string equality.
	uProperty1 := make(map[string]interface{})
	uProperty1["property"] = "uProperty1"
	uProperty1["type"] = "categorical"
	uProperty1["value"] = "uProperty1Value"
	uProperty1["operator"] = "equals"
	event1["user_properties"] = []interface{}{uProperty1}
	// Event 2.
	event2 := make(map[string]interface{})
	event2["name"] = "endEvent"
	// Floating point greater than.
	property3 := make(map[string]interface{})
	property3["property"] = "property3"
	property3["type"] = "numerical"
	property3["value"] = 5.001
	property3["operator"] = "greaterThan"
	// Floating point lesser than.
	property4 := make(map[string]interface{})
	property4["property"] = "property4"
	property4["type"] = "numerical"
	property4["value"] = -10.0
	property4["operator"] = "lesserThan"
	// Floating point equality.
	property5 := make(map[string]interface{})
	property5["property"] = "property5"
	property5["type"] = "numerical"
	property5["value"] = 50.01
	property5["operator"] = "equals"
	event2["properties"] = []interface{}{property3, property4, property5}
	// Floating point equality user property.
	uProperty2 := make(map[string]interface{})
	uProperty2["property"] = "uProperty2"
	uProperty2["type"] = "numerical"
	uProperty2["value"] = 50.01
	uProperty2["operator"] = "equals"
	event2["user_properties"] = []interface{}{uProperty2}
	query["eventsWithProperties"] = []interface{}{event1, event2}
	startEvent, startEventConstraints, endEvent, endEventConstraints, _, err = H.ParseFactorQuery(query)
	assert.Nil(t, err)
	assert.Equal(t, startEvent, "startEvent")
	assertEqualConstraints(t, P.EventConstraints{
		EPNumericConstraints: []P.NumericConstraint{
			P.NumericConstraint{
				PropertyName: "property1",
				LowerBound:   4.5,
				UpperBound:   5.5,
				IsEquality:   true,
			},
		},
		EPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "property2",
				PropertyValue: "property2Value",
			},
		},
		UPNumericConstraints: []P.NumericConstraint{},
		UPCategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "uProperty1",
				PropertyValue: "uProperty1Value",
			},
		},
	}, *startEventConstraints)
	assert.Equal(t, endEvent, "endEvent")
	assertEqualConstraints(t,
		P.EventConstraints{
			EPNumericConstraints: []P.NumericConstraint{
				P.NumericConstraint{
					PropertyName: "property3",
					LowerBound:   5.001,
					UpperBound:   math.MaxFloat64,
					IsEquality:   false,
				},
				P.NumericConstraint{
					PropertyName: "property4",
					LowerBound:   -math.MaxFloat64,
					UpperBound:   -10.0,
					IsEquality:   false,
				},
				P.NumericConstraint{
					PropertyName: "property5",
					LowerBound:   50.01 - 0.1,
					UpperBound:   50.01 + 0.1,
					IsEquality:   true,
				},
			},
			EPCategoricalConstraints: []P.CategoricalConstraint{},
			UPNumericConstraints: []P.NumericConstraint{
				P.NumericConstraint{
					PropertyName: "uProperty2",
					LowerBound:   50.01 - 0.1,
					UpperBound:   50.01 + 0.1,
					IsEquality:   true,
				},
			},
			UPCategoricalConstraints: []P.CategoricalConstraint{},
		},
		*endEventConstraints)

	// Three events.
	event1 = make(map[string]interface{})
	event1["name"] = "event1"
	event2 = make(map[string]interface{})
	event2["name"] = "event2"
	event3 := make(map[string]interface{})
	event3["name"] = "event3"
	query["eventsWithProperties"] = []interface{}{event1, event2, event3}
	startEvent, startEventConstraints, endEvent, endEventConstraints, _, err = H.ParseFactorQuery(query)
	assert.NotNil(t, err)
	assert.Equal(t, startEvent, "")
	assert.Nil(t, startEventConstraints)
	assert.Equal(t, endEvent, "")
	assert.Nil(t, endEventConstraints)
}
