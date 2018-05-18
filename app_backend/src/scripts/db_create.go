package main

// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/autometa/app_backend/
// go run db_create.go

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
	db.CreateTable(&M.Event{})
	log.Info("Created Event table")
}
