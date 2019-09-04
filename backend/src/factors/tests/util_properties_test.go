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

func TestGetPropertyKeyValueType(t *testing.T) {
	assert.Equal(t, U.PropertyTypeCategorical, U.GetPropertyKeyValueType("testKey", "10.24string"))
	assert.Equal(t, U.PropertyTypeCategorical, U.GetPropertyKeyValueType("testKey", "10.24"))
	assert.Equal(t, U.PropertyTypeNumerical, U.GetPropertyKeyValueType("testKey", 10.24))
	assert.Equal(t, U.PropertyTypeCategorical, U.GetPropertyKeyValueType("$qp_utm_campaignid", "10.24"))
	assert.Equal(t, U.PropertyTypeCategorical, U.GetPropertyKeyValueType("$qp_utm_adgroupid", "10.35"))
	assert.Equal(t, U.PropertyTypeCategorical, U.GetPropertyKeyValueType("utm_creative", "10"))
	assert.Equal(t, U.PropertyTypeNumerical, U.GetPropertyKeyValueType("testKey", 10.24))
	assert.Equal(t, U.PropertyTypeUnknown, U.GetPropertyKeyValueType("testKey", true))
	assert.Equal(t, U.PropertyTypeUnknown, U.GetPropertyKeyValueType("testKey", []string{"value1", "value2"}))
}

func TestGetCleanPropertyValue(t *testing.T) {
	assert.Equal(t, "value with?reserved+characters$are(properties)",
		U.GetUnEscapedPropertyValue("value%20with%3Freserved%2Bcharacters%24are%28properties%29"))
}
