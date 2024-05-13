package tests

import (
	"testing"

	"factors/cache"
	cacheRedis "factors/cache/redis"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	key1 := cache.Key{
		ProjectID: project.ID,
		Prefix:    "test1",
	}
	key2 := cache.Key{
		ProjectID: project.ID,
		Prefix:    "test2",
	}

	// Set only one key. Expiry 1 second.
	cacheRedis.Set(&key1, "testValue1", 1)

	exists, err := cacheRedis.Exists(&key1)
	assert.Nil(t, err)
	assert.True(t, exists)

	exists, err = cacheRedis.Exists(&key2)
	assert.Nil(t, err)
	assert.False(t, exists)
}

func TestMGet(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	key1 := cache.Key{
		ProjectID: project.ID,
		Prefix:    "test1",
	}
	key2 := cache.Key{
		ProjectID: project.ID,
		Prefix:    "test2",
	}
	key3 := cache.Key{
		ProjectID: project.ID,
		Prefix:    "test3",
	}

	// Set values for only 2 keys. Expiry 1 second.
	cacheRedis.Set(&key1, "testValue1", 1)
	cacheRedis.Set(&key2, "testValue2", 1)

	values, err := cacheRedis.MGet(&key1, &key2, &key3)
	assert.Nil(t, err)
	assert.Equal(t, "testValue1", values[0])
	assert.Equal(t, "testValue2", values[1])
	assert.Equal(t, "", values[2])
}
