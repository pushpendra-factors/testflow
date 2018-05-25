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

	// Create projects table.
	if err := db.CreateTable(&M.Project{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("projects table creation failed.")
	} else {
		log.Info("Created projects table")
	}

	// Create users table.
	if err := db.CreateTable(&M.User{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table creation failed.")
	} else {
		log.Info("Created users table")
	}
	// Add foreign key constraints.
	if err := db.Model(&M.User{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table association with projects table failed.")
	} else {
		log.Info("users table is associated with projects table.")
	}
	// Add unique index on project_id+customer_user_id.
	if err := db.Model(&M.User{}).AddUniqueIndex("project_id_customer_user_id_idx", "project_id", "customer_user_id").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table project_id:customer_user_id unique index failed.")
	} else {
		log.Info("users table project_id:customer_user_id unique index created.")
	}

	// Create event_names table.
	if err := db.CreateTable(&M.EventName{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("event_names table creation failed.")
	} else {
		log.Info("Created event_names table")
	}
	// Add foreign key constraints.
	if err := db.Model(&M.EventName{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("event_names table association with projects table failed.")
	} else {
		log.Info("event_names table is associated with projects table.")
	}

	// Create events table.
	if err := db.CreateTable(&M.Event{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table creation failed.")
	} else {
		log.Info("Created events table")
	}
	// Add foreign key constraints.
	if err := db.Model(&M.Event{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with projects table failed.")
	} else {
		log.Info("events table is associated with projects table.")
	}
	// Adding composite foreign key with users table.
	if err := db.Model(&M.Event{}).AddForeignKey("project_id, user_id", "users(project_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with users table failed.")
	} else {
		log.Info("events table is associated with users table.")
	}
	// Adding composite foreign key with event_names table.
	if err := db.Model(&M.Event{}).AddForeignKey("project_id, event_name", "event_names(project_id, name)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with event_names table failed.")
	} else {
		log.Info("events table is associated with event_names table.")
	}
}
