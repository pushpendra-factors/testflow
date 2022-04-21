package store

import (
	"factors/model"
	storeMemSQL "factors/model/store/memsql"
)

// GetStore - Should decide on which model implementation to use by
// configuration and return the store.
func GetStore() model.Model {
	var store model.Model
	store = &storeMemSQL.MemSQL{}
	return store
}
