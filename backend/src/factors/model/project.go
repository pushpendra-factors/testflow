package model

import (
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type Project struct {
	ID   uint64 `gorm:"primary_key:true;" json:"id"`
	Name string `gorm:"not null;" json:"name"`
	// Token is indexed. Ignored on JSON response.
	Token     string    `gorm:"size:32" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const TOKEN_GEN_RETRY_LIMIT = 5

// Checks for the existence of token already.
func isTokenExist(token string) (exists int, err error) {
	db := C.GetServices().Db

	var count uint64
	if err := db.Model(&Project{}).Where("token = ?", token).Count(&count).Error; err != nil {
		return -1, err
	}

	if count > 0 {
		return 1, nil
	}

	return 0, nil
}

func generateUniqueToken() (token string, err error) {
	for tryCount := 0; tryCount < TOKEN_GEN_RETRY_LIMIT; tryCount++ {
		token = U.RandomLowerAphaNumString(32)
		tokenExists, err := isTokenExist(token)
		if err != nil {
			return "", err
		}
		// Break generation, if token doesn't exist already.
		if tokenExists == 0 {
			return token, nil
		}
	}
	return "", fmt.Errorf("Token generation failed after %d attempts.", TOKEN_GEN_RETRY_LIMIT)
}

func (project *Project) BeforeCreate() (err error) {
	// Unique token assignment.
	var token string
	if token, err = generateUniqueToken(); err != nil {
		return err
	}
	project.Token = token
	return nil
}

func CreateProject(project *Project) (*Project, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"project": &project}).Info("Creating project")

	// Input Validation. (ID is to be auto generated)
	if project.ID > 0 {
		log.Error("CreateProject Failed. ProjectId provided.")
		return nil, http.StatusBadRequest
	}

	if err := db.Create(project).Error; err != nil {
		log.WithFields(log.Fields{"project": &project, "error": err}).Error("CreateProject Failed")
		return nil, http.StatusInternalServerError
	}

	return project, DB_SUCCESS
}

// CreateProjectDependencies bootstraps a project.
func CreateProjectDependencies(project *Project) int {
	// Associated project setting creation.
	if _, errCode := CreateProjectSetting(&ProjectSetting{ProjectId: project.ID}); errCode != DB_SUCCESS {
		log.WithFields(log.Fields{"project": project}).Error("Creating project_settings failed")
		return errCode
	}

	return DB_SUCCESS
}

func GetProject(id uint64) (*Project, int) {
	db := C.GetServices().Db

	var project Project
	if err := db.Where("id = ?", id).First(&project).Error; err != nil {
		return nil, 404
	} else {
		return &project, DB_SUCCESS
	}
}

func GetProjectByToken(token string) (*Project, int) {
	db := C.GetServices().Db

	// Todo(Dinesh): Move to validations.
	cleanToken := strings.TrimSpace(token)
	if len(cleanToken) == 0 {
		return nil, http.StatusBadRequest
	}

	var project Project
	if err := db.Where("token = ?", cleanToken).First(&project).Error; err != nil {
		return nil, http.StatusNotFound
	}

	return &project, DB_SUCCESS
}

func GetProjects() ([]Project, int) {
	db := C.GetServices().Db

	var projects []Project
	if err := db.Find(&projects).Error; err != nil {
		return nil, 404
	} else {
		return projects, DB_SUCCESS
	}
}

// isValidProjectScope return false if projectId is invalid.
func isValidProjectScope(id uint64) bool {
	return id != 0
}
