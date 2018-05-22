package main

// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/autometa/app_backend/
// go run db_create.go

import (
	C "config"
	M "model"

	_ "github.com/jinzhu/gorm"
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

	// Create accounts table.
	if err := db.CreateTable(&M.Account{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("accounts table creation failed.")
	} else {
		log.Info("Created accounts table")
	}

	// Create users table.
	if err := db.CreateTable(&M.User{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table creation failed.")
	} else {
		log.Info("Created users table")
	}
	// Add foreign key constraints.
	if err := db.Model(&M.User{}).AddForeignKey("account_id", "accounts(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table association with accounts table failed.")
	} else {
		log.Info("users table is associated with accounts table.")
	}

	// Create event_names table.
	if err := db.CreateTable(&M.EventName{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("event_names table creation failed.")
	} else {
		log.Info("Created event_names table")
	}
	// Add foreign key constraints.
	if err := db.Model(&M.EventName{}).AddForeignKey("account_id", "accounts(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("event_names table association with accounts table failed.")
	} else {
		log.Info("event_names table is associated with accounts table.")
	}

	// Create events table.
	if err := db.CreateTable(&M.Event{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table creation failed.")
	} else {
		log.Info("Created events table")
	}
	// Add foreign key constraints.
	if err := db.Model(&M.Event{}).AddForeignKey("account_id", "accounts(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with accounts table failed.")
	} else {
		log.Info("events table is associated with accounts table.")
	}
	// Adding composite foreign key with users table.
	if err := db.Model(&M.Event{}).AddForeignKey("account_id, user_id", "users(account_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with users table failed.")
	} else {
		log.Info("events table is associated with users table.")
	}
	// Adding composite foreign key with event_names table.
	if err := db.Model(&M.Event{}).AddForeignKey("account_id, event_name", "event_names(account_id, name)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with event_names table failed.")
	} else {
		log.Info("events table is associated with event_names table.")
	}
}
