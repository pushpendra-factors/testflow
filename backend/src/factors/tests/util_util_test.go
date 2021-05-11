package tests

import (
	U "factors/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertIntToUUID(t *testing.T) {
	uuid, err := U.ConvertIntToUUID(1000)
	assert.Nil(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000001000", uuid)
	assert.True(t, U.IsValidUUID(uuid))

	// 10 digits
	uuid, err = U.ConvertIntToUUID(123456789)
	assert.Nil(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000123456789", uuid)
	assert.True(t, U.IsValidUUID(uuid))

	// 12 digits
	uuid, err = U.ConvertIntToUUID(123456789876)
	assert.Nil(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-123456789876", uuid)
	assert.True(t, U.IsValidUUID(uuid))

	// 13 digits
	uuid, err = U.ConvertIntToUUID(1234567898765)
	assert.NotNil(t, err)
	assert.Empty(t, uuid)
}
