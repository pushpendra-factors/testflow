package main

// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_db_create.go

import (
	C "factors/config"
	M "factors/model"

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
	// Add unique index on project tokens.
	if err := db.Exec("CREATE UNIQUE INDEX token_unique_idx ON projects(token);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("projects table token unique indexing failed.")
	} else {
		log.Info("projects table token unique index created.")
	}

	// Create project settings table.
	if err := db.CreateTable(&M.ProjectSetting{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_settings table creation failed.")
	} else {
		log.Info("Created project_settings table.")
	}
	// Add foreign key constraint by project.
	if err := db.Model(&M.ProjectSetting{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_settings table association with projects table failed.")
	} else {
		log.Info("project_settings table is associated with projects table.")
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
	if err := db.Model(&M.Event{}).AddForeignKey("project_id, event_name_id", "event_names(project_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with event_names table failed.")
	} else {
		log.Info("events table is associated with event_names table.")
	}
	// Add sort index on project_id and created_at.
	if err := db.Exec("CREATE INDEX project_id_user_id_created_at_idx ON events (project_id, user_id, created_at NULLS FIRST) ;").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table project_id:user_id:created_at sort index failed.")
	} else {
		log.Info("events table project_id:user_id:created_at sort index created.")
	}
}
