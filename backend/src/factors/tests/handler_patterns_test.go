package tests

import (
	H "factors/handler"
	P "factors/pattern"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFactorQuery(t *testing.T) {
	// No events.
	var query = make(map[string]interface{})
	startEvent, startEventConstraints, endEvent, endEventConstraints, err := H.ParseFactorQuery(query)
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
	startEvent, startEventConstraints, endEvent, endEventConstraints, err = H.ParseFactorQuery(query)
	assert.Nil(t, err)
	assert.Equal(t, startEvent, "")
	assert.Nil(t, startEventConstraints)
	assert.Equal(t, endEvent, "endEvent")
	assert.Equal(t, *endEventConstraints, P.EventConstraints{
		NumericConstraints:     []P.NumericConstraint{},
		CategoricalConstraints: []P.CategoricalConstraint{},
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
	query["eventsWithProperties"] = []interface{}{event1, event2}
	startEvent, startEventConstraints, endEvent, endEventConstraints, err = H.ParseFactorQuery(query)
	assert.Nil(t, err)
	assert.Equal(t, startEvent, "startEvent")
	assert.Equal(t, *startEventConstraints, P.EventConstraints{
		NumericConstraints: []P.NumericConstraint{
			P.NumericConstraint{
				PropertyName: "property1",
				LowerBound:   4.5,
				UpperBound:   5.5,
				IsEquality:   true,
			},
		},
		CategoricalConstraints: []P.CategoricalConstraint{
			P.CategoricalConstraint{
				PropertyName:  "property2",
				PropertyValue: "property2Value",
			}},
	})
	assert.Equal(t, endEvent, "endEvent")
	assert.Equal(t, *endEventConstraints, P.EventConstraints{
		NumericConstraints: []P.NumericConstraint{
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
		CategoricalConstraints: []P.CategoricalConstraint{},
	})

	// Three events.
	event1 = make(map[string]interface{})
	event1["name"] = "event1"
	event2 = make(map[string]interface{})
	event2["name"] = "event2"
	event3 := make(map[string]interface{})
	event3["name"] = "event3"
	query["eventsWithProperties"] = []interface{}{event1, event2, event3}
	startEvent, startEventConstraints, endEvent, endEventConstraints, err = H.ParseFactorQuery(query)
	assert.NotNil(t, err)
	assert.Equal(t, startEvent, "")
	assert.Nil(t, startEventConstraints)
	assert.Equal(t, endEvent, "")
	assert.Nil(t, endEventConstraints)
}
