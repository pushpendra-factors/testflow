package model

import (
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type Project struct {
	ID   uint64 `gorm:"primary_key:true;" json:"id"`
	Name string `gorm:"not null;" json:"name"`
	// An index created on token.
	Token string `gorm:"size:32" json:"token"`
	// An index created on private_token.
	PrivateToken string    `gorm:"size:32" json:"private_token"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

const TOKEN_GEN_RETRY_LIMIT = 5
const DefaultProjectName = "My Project"

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
	// Token creation.
	if token, err := generateUniqueToken(); err != nil {
		return err
	} else {
		project.Token = token
	}

	// PrivateToken creation.
	if privateToken, err := generateUniqueToken(); err != nil {
		return err
	} else {
		project.PrivateToken = privateToken
	}

	return nil
}

func createProject(project *Project) (*Project, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"project": &project}).Info("Creating project")

	// Input Validation. (ID is to be auto generated)
	if project.ID > 0 {
		log.Error("CreateProject Failed. ProjectId provided.")
		return nil, http.StatusBadRequest
	}

	if err := db.Create(project).Error; err != nil {
		log.WithFields(log.Fields{"project": &project}).WithError(err).Error("CreateProject Failed")
		return nil, http.StatusInternalServerError
	}

	return project, http.StatusCreated
}

func createProjectDependencies(projectID uint64) int {
	// Associated project setting creation with default state.
	defaultAutoTrackState := true
	defaultExcludebotState := true

	_, errCode := createProjectSetting(&ProjectSetting{ProjectId: projectID,
		AutoTrack: &defaultAutoTrackState, ExcludeBot: &defaultExcludebotState})
	if errCode != http.StatusCreated {
		log.WithFields(log.Fields{"projectID": projectID}).Error("Creating project_settings failed")
		return errCode
	}

	return http.StatusCreated
}

// CreateProjectWithDependencies seperate create method with dependencies to avoid breaking tests.
func CreateProjectWithDependencies(project *Project, agentUUID string, agentRole uint64, billingAccountID uint64) (*Project, int) {
	cProject, errCode := createProject(project)
	if errCode != http.StatusCreated {
		return nil, errCode
	}
	errCode = createProjectDependencies(cProject.ID)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	_, errCode = CreateProjectAgentMappingWithDependencies(&ProjectAgentMapping{
		ProjectID: cProject.ID,
		AgentUUID: agentUUID,
		Role:      agentRole,
	})
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	_, errCode = createProjectBillingAccountMapping(project.ID, billingAccountID)
	return cProject, errCode
}

// CreateDefaultProjectForAgent creates project for an agent if there is no project
func CreateDefaultProjectForAgent(agentUUID string) (*Project, int) {
	if agentUUID == "" {
		return nil, http.StatusBadRequest
	}

	errCode := DoesAgentHaveProject(agentUUID)
	if errCode == http.StatusFound {
		return nil, http.StatusConflict
	}
	if errCode != http.StatusNotFound {
		return nil, errCode
	}

	billingAcc, errCode := GetBillingAccountByAgentUUID(agentUUID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("CreateDefaultProjectForAgent Failed, billing account error")
		return nil, errCode
	}

	cProject, errCode := CreateProjectWithDependencies(&Project{Name: DefaultProjectName}, agentUUID, ADMIN, billingAcc.ID)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	return cProject, http.StatusCreated
}

func GetProject(id uint64) (*Project, int) {
	db := C.GetServices().Db

	var project Project
	if err := db.Where("id = ?", id).First(&project).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &project, http.StatusFound
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
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &project, http.StatusFound
}

func GetProjectByPrivateToken(privateToken string) (*Project, int) {
	db := C.GetServices().Db

	cleanToken := strings.TrimSpace(privateToken)
	if len(cleanToken) == 0 {
		return nil, http.StatusBadRequest
	}

	var project Project
	if err := db.Where("private_token = ?", cleanToken).First(&project).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		log.WithField("private_token", privateToken).WithError(err).Error(
			"Failed to get project by private token.")
		return nil, http.StatusInternalServerError
	}

	return &project, http.StatusFound
}

func GetProjects() ([]Project, int) {
	db := C.GetServices().Db

	var projects []Project
	if err := db.Find(&projects).Error; err != nil {
		return nil, http.StatusInternalServerError
	}
	if len(projects) == 0 {
		return projects, http.StatusNotFound
	}
	return projects, http.StatusFound
}

// isValidProjectScope return false if projectId is invalid.
func isValidProjectScope(id uint64) bool {
	return id != 0
}

func GetProjectsByIDs(ids []uint64) ([]Project, int) {

	if len(ids) == 0 {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var projects []Project
	if err := db.Limit(len(ids)).Where(ids).Find(&projects).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(projects) == 0 {
		return projects, http.StatusNoContent
	}

	return projects, http.StatusFound
}
