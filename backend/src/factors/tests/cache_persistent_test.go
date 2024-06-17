package tests

import (
	"testing"

	"factors/cache"
	cacheDB "factors/cache/db"
	cachePersistent "factors/cache/persistent"
	cacheRedis "factors/cache/redis"
	"factors/model/model"
	U "factors/util"

	"github.com/stretchr/testify/assert"
)

func TestCacheDBSetBatch(t *testing.T) {

	var records = make(map[*cache.Key]string, 0)

	cacheKey1, _ := model.GetValuesByEventPropertyRollUpCacheKey(2, "$session", U.RandomLowerAphaNumString(5), "20240524")
	cacheKey2, _ := model.GetValuesByEventPropertyRollUpCacheKey(2, "$session", U.RandomLowerAphaNumString(5), "20240524")
	cacheKey3, _ := model.GetValuesByEventPropertyRollUpCacheKey(2, "$session", U.RandomLowerAphaNumString(5), "20240524")

	records[cacheKey1] = "value1"
	records[cacheKey2] = "value2"
	records[cacheKey3] = "value3"

	err := cacheDB.SetBatch(records, 7200)
	assert.Nil(t, err)

	v1, err := cacheDB.Get(cacheKey1)
	assert.Nil(t, err)

	v2, err := cacheDB.Get(cacheKey2)
	assert.Nil(t, err)

	v3, err := cacheDB.Get(cacheKey3)
	assert.Nil(t, err)

	assert.Equal(t, "value1", v1)
	assert.Equal(t, "value2", v2)
	assert.Equal(t, "value3", v3)

	// Update same records and it should be updated without duplicate error.

	records[cacheKey1] = "value11"
	records[cacheKey2] = "value22"
	records[cacheKey3] = "value33"

	err = cacheDB.SetBatch(records, 7200)
	assert.Nil(t, err)

	v1, err = cacheDB.Get(cacheKey1)
	assert.Nil(t, err)

	v2, err = cacheDB.Get(cacheKey2)
	assert.Nil(t, err)

	v3, err = cacheDB.Get(cacheKey3)
	assert.Nil(t, err)

	assert.Equal(t, "value11", v1)
	assert.Equal(t, "value22", v2)
	assert.Equal(t, "value33", v3)
}

func TestCacheDBFallback(t *testing.T) {

	// set cache only on redis
	cacheKey1, _ := model.GetValuesByEventPropertyRollUpCacheKey(2, "$session", U.RandomLowerAphaNumString(5), "20240524")
	err := cacheRedis.SetPersistent(cacheKey1, "value1", 7200)
	assert.Nil(t, err)

	v, b, err := cacheRedis.GetIfExistsPersistent(cacheKey1)
	assert.Nil(t, err)
	assert.True(t, b)
	assert.Equal(t, "value1", v)

	v, b, err = cachePersistent.GetIfExists(cacheKey1, true)
	assert.Nil(t, err)
	assert.True(t, b)
	assert.NotEmpty(t, "value1", v)

	v, err = cachePersistent.Get(cacheKey1, true)
	assert.Nil(t, err)
	assert.True(t, b)
	assert.NotEmpty(t, "value1", v)
}
