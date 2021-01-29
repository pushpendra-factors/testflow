package model

import (
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type Project struct {
	ID   uint64 `gorm:"primary_key:true;" json:"id"`
	Name string `gorm:"not null;" json:"name"`
	// An index created on token.
	Token string `gorm:"size:32" json:"token"`
	// An index created on private_token.
	PrivateToken string          `gorm:"size:32" json:"private_token"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	ProjectURI   string          `json:"project_uri"`
	TimeFormat   string          `json:"time_format"`
	DateFormat   string          `json:"date_format"`
	TimeZone     string          `json:"time_zone"`
	JobsMetadata *postgres.Jsonb `json:"-"`
}

const TOKEN_GEN_RETRY_LIMIT = 5
const DefaultProjectName = "My Project"
const ENABLE_DEFAULT_WEB_ANALYTICS = false

const JobsMetadataKeyNextSessionStartTimestamp = "next_session_start_timestamp"
const JobsMetadataColumnName = "jobs_metadata"

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
	logCtx := log.WithFields(log.Fields{"project": project})

	// Input Validation. (ID is to be auto generated)
	if project.ID > 0 {
		logCtx.Error("CreateProject Failed. ProjectId provided.")
		return nil, http.StatusBadRequest
	}

	// Initialize jobs metadata.
	jobsMetadata := &map[string]interface{}{
		// To start pulling events from the time of project
		// create for adding session for the first time.
		JobsMetadataKeyNextSessionStartTimestamp: U.TimeNowUnix(),
	}
	jobsMetadataJsonb, err := U.EncodeToPostgresJsonb(jobsMetadata)
	if err != nil {
		// Log error and continue to create project.
		logCtx.WithField("jobs_metadata", jobsMetadata).WithError(err).
			Error("Failed to marshal jobs metadata on create project.")
	} else {
		project.JobsMetadata = jobsMetadataJsonb
	}

	db := C.GetServices().Db
	if err := db.Create(project).Error; err != nil {
		logCtx.WithError(err).Error("Create project failed.")
		return nil, http.StatusInternalServerError
	}

	return project, http.StatusCreated
}

func UpdateProject(projectId uint64, project *Project) int {
	db := C.GetServices().Db

	logCtx := log.WithField("project_id", project.ID)
	updateFields := make(map[string]interface{}, 0)
	if project.Name != "" {
		updateFields["name"] = project.Name
	}
	if project.TimeFormat != "" {
		updateFields["time_format"] = project.TimeFormat
	}
	if project.DateFormat != "" {
		updateFields["date_format"] = project.DateFormat
	}
	if project.ProjectURI != "" {
		updateFields["project_uri"] = project.ProjectURI
	}
	if project.TimeZone != "" {
		updateFields["time_zone"] = project.TimeZone
	}
	err := db.Model(&Project{}).Where("id = ?", projectId).Update(updateFields).Error
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to execute query of update project")
		return http.StatusInternalServerError
	}
	return 0
}

func createProjectDependencies(projectID uint64, agentUUID string) int {
	logCtx := log.WithField("project_id", projectID)

	// Associated project setting creation with default state.
	defaultAutoTrackState := true
	defaultExcludebotState := true
	_, errCode := createProjectSetting(&ProjectSetting{ProjectId: projectID,
		AutoTrack: &defaultAutoTrackState, ExcludeBot: &defaultExcludebotState})
	if errCode != http.StatusCreated {
		logCtx.Error("Create project settings failed on create project dependencies.")
		return errCode
	}

	if ENABLE_DEFAULT_WEB_ANALYTICS {
		errCode = createDefaultDashboardsForProject(projectID, agentUUID)
		if errCode != http.StatusCreated {
			logCtx.Error("Create default dashboards failed on create project dependencies.")
			return errCode
		}
	}

	return http.StatusCreated
}

// CreateProjectWithDependencies seperate create method with dependencies to avoid breaking tests.
func CreateProjectWithDependencies(project *Project, agentUUID string,
	agentRole uint64, billingAccountID uint64) (*Project, int) {

	cProject, errCode := createProject(project)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	errCode = createProjectDependencies(cProject.ID, agentUUID)
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
		log.WithField("err_code", errCode).
			Error("CreateDefaultProjectForAgent Failed, billing account error")
		return nil, errCode
	}

	cProject, errCode := CreateProjectWithDependencies(
		&Project{Name: DefaultProjectName},
		agentUUID, ADMIN, billingAcc.ID)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	return cProject, http.StatusCreated
}

func GetProject(id uint64) (*Project, int) {
	db := C.GetServices().Db
	logCtx := log.WithField("project_id", id)

	var project Project
	if err := db.Where("id = ?", id).First(&project).Error; err != nil {
		logCtx.WithError(err).Error("Getting project by id failed")
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
		log.WithField("token", token).WithError(err).Error(
			"Failed to get project by token.")
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
		log.WithError(err).Error("Getting all projects failed")
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
		log.WithError(err).Error("Getting projects using ids failed")
		return nil, http.StatusInternalServerError
	}

	if len(projects) == 0 {
		return projects, http.StatusNoContent
	}

	return projects, http.StatusFound
}

// GetAllProjectIDs Gets the ids of all the existing projects.
func GetAllProjectIDs() ([]uint64, int) {
	projectIds := make([]uint64, 0, 0)

	db := C.GetServices().Db
	rows, err := db.Raw("SELECT id FROM projects").Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get all project ids.")
		return projectIds, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var projectId uint64
		if err = rows.Scan(&projectId); err != nil {
			log.WithError(err).Error("Failed to get all project ids. Scanning failed.")
			return projectIds, http.StatusInternalServerError
		}

		projectIds = append(projectIds, projectId)
	}

	if len(projectIds) == 0 {
		return projectIds, http.StatusNotFound
	}

	return projectIds, http.StatusFound
}

// GetNextSessionStartTimestampForProject - Returns start timestamp for
// pulling events, for add session job, by project.
func GetNextSessionStartTimestampForProject(projectID uint64) (int64, int) {
	logCtx := log.WithField("project_id", projectID)

	if projectID == 0 {
		logCtx.WithField("project_id", projectID).Error("Invalid args to method.")
		return 0, http.StatusBadRequest
	}

	db := C.GetServices().Db
	rows, err := db.Table("projects").Limit(1).Where("id = ?", projectID).
		Select(fmt.Sprintf("%s->>'%s' as session_start_timestamp", JobsMetadataColumnName,
			JobsMetadataKeyNextSessionStartTimestamp)).Rows()
	if err != nil {
		logCtx.WithField("project_id", projectID).WithError(err).
			Error("Failed to get next session start timestamp for project.")
		return 0, http.StatusInternalServerError
	}
	defer rows.Close()

	var sessionStartTimestamp *int64
	for rows.Next() {
		err = rows.Scan(&sessionStartTimestamp)
		if err != nil {
			logCtx.WithError(err).Error("Failed to read next session start timestamp.")
			return 0, http.StatusInternalServerError
		}
	}

	if err := rows.Err(); err != nil {
		logCtx.WithError(err).Error("Failure on rows scanner.")
		return 0, http.StatusInternalServerError
	}

	if sessionStartTimestamp == nil {
		return 0, http.StatusNotFound
	}

	return *sessionStartTimestamp, http.StatusFound
}

// UpdateNextSessionStartTimestampForProject - Updates next session start timestamp
// on project jobs metadata.
func UpdateNextSessionStartTimestampForProject(projectID uint64, timestamp int64) int {
	logCtx := log.WithField("project_id", projectID).WithField("timestamp", timestamp)

	if projectID == 0 || timestamp == 0 {
		logCtx.WithField("project_id", projectID).WithField("timestamp", 0).
			Error("Invalid args to method.")
		return http.StatusBadRequest
	}

	// Updating the add_session JSON field directly, to avoid state corruption
	// because of multiple version of JSON being updaetd by multiple jobs simultaneously.
	query := fmt.Sprintf(`UPDATE projects SET jobs_metadata = jobs_metadata - '%s' || '{"%s": %d}' WHERE id = %d`,
		JobsMetadataKeyNextSessionStartTimestamp, JobsMetadataKeyNextSessionStartTimestamp, timestamp, projectID)
	db := C.GetServices().Db
	err := db.Exec(query).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update next session start timestamp for project.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}
