package tests

import (
	"testing"

	"factors/cache"
	cacheDB "factors/cache/db"
	"factors/model/model"
	U "factors/util"

	"github.com/stretchr/testify/assert"
)

func TestCacheDBSetBatch(t *testing.T) {

	var records = make(map[*cache.Key]string, 0)

	cacheKey1, _ := model.GetValuesByEventPropertyRollUpCacheKey(2, "$session", U.RandomLowerAphaNumString(5), "20240524")
	cacheKey2, _ := model.GetValuesByEventPropertyRollUpCacheKey(2, "$session", U.RandomLowerAphaNumString(5), "20240524")

	records[cacheKey1] = "value1"
	records[cacheKey2] = "value2"

	err := cacheDB.SetBatch(records, 7200)
	assert.Nil(t, err)

	v1, err := cacheDB.Get(cacheKey1)
	assert.Nil(t, err)

	v2, err := cacheDB.Get(cacheKey2)
	assert.Nil(t, err)

	assert.Equal(t, "value1", v1)
	assert.Equal(t, "value2", v2)

	// Update same records and it should be updated without duplicate error.

	records[cacheKey1] = "value3"
	records[cacheKey2] = "value4"

	err = cacheDB.SetBatch(records, 7200)
	assert.Nil(t, err)

	v1, err = cacheDB.Get(cacheKey1)
	assert.Nil(t, err)

	v2, err = cacheDB.Get(cacheKey2)
	assert.Nil(t, err)

	assert.Equal(t, "value3", v1)
	assert.Equal(t, "value4", v2)

}
