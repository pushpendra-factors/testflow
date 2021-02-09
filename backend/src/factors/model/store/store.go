package store

import (
	"factors/model"
	storePostgres "factors/model/store/postgres"
)

// GetStore - Should decide on which model implementation to use by
// configuration and return the store.
func GetStore() model.Model {
	// TODO: Use configuration and add memsql
	// store selection based on config here.
	var store model.Model = &storePostgres.Postgres{}
	return store
}
