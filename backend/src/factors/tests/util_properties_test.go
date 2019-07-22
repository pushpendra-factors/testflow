package tests

import (
	U "factors/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillPropertyKvsFromPropertiesJson(t *testing.T) {
	propertiesKvs := make(map[string]map[interface{}]bool, 0)
	sampleValuesLimit := 1

	// Should add sample values for each key upto limit.
	U.FillPropertyKvsFromPropertiesJson([]byte(`{"prop1":"value11"}`), &propertiesKvs, sampleValuesLimit)
	U.FillPropertyKvsFromPropertiesJson([]byte(`{"prop1":"value12"}`), &propertiesKvs, sampleValuesLimit)
	U.FillPropertyKvsFromPropertiesJson([]byte(`{"prop2":"value2"}`), &propertiesKvs, sampleValuesLimit)
	U.FillPropertyKvsFromPropertiesJson([]byte(`{"prop3": { "subprop": ["subvalue1", "subvalue2"]}}`), &propertiesKvs, sampleValuesLimit)
	U.FillPropertyKvsFromPropertiesJson([]byte(`{"prop4": ["subvalue3", "subvalue4"]}`), &propertiesKvs, sampleValuesLimit)

	assert.Contains(t, propertiesKvs, "prop1")
	assert.Contains(t, propertiesKvs, "prop2")
	assert.NotContains(t, propertiesKvs, "prop3")
	assert.NotContains(t, propertiesKvs, "prop4")
	assert.Len(t, propertiesKvs["prop1"], 1)
	assert.Contains(t, propertiesKvs["prop1"], "value11")
	assert.Contains(t, propertiesKvs["prop2"], "value2")
}
