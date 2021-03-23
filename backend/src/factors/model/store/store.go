package store

import (
	"factors/config"
	"factors/model"
	storeMemSQL "factors/model/store/memsql"
	storePostgres "factors/model/store/postgres"
)

// GetStore - Should decide on which model implementation to use by
// configuration and return the store.
func GetStore() model.Model {
	var store model.Model
	if config.UseMemSQLDatabaseStore() {
		store = &storeMemSQL.MemSQL{}
	} else {
		store = &storePostgres.Postgres{}
	}
	return store
}
