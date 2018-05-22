package main

// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/autometa/app_backend/
// go run db_delete.go

import (
	C "config"
	M "model"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Initialize configs and connections.
	err := C.Init()
	if err != nil {
		log.Error("Failed to initialize.")
		return
	}

	if C.GetConfig().Env != C.DEVELOPMENT {
		log.Error("Not Development Environment. Aborting")
		return
	}

	db := C.GetServices().Db
	defer db.Close()

	// Drop events table.
	if err := db.DropTableIfExists(&M.Event{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table deletion failed.")
	} else {
		log.Info("Dropped events table")
	}

	// Drop event_names Table.
	if err := db.DropTableIfExists(&M.EventName{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("event_names table deletion failed.")
	} else {
		log.Info("Dropped event_names table")
	}

	// Drop users Table.
	if err := db.DropTableIfExists(&M.User{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table deletion failed.")
	} else {
		log.Info("Dropped users table")
	}

	// Drop accounts Table.
	if err := db.DropTableIfExists(&M.Account{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("accounts table deletion failed.")
	} else {
		log.Info("Dropped accounts table")
	}
}
