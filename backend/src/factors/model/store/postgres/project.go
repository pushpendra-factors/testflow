package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const TOKEN_GEN_RETRY_LIMIT = 5
const ENABLE_DEFAULT_WEB_ANALYTICS = false

// Checks for the existence of token already.
func isTokenExist(token string, private bool) (exists int, err error) {
	db := C.GetServices().Db

	whereCondition := "token = ?"
	if private {
		whereCondition = "private_token = ?"
	}
	var count uint64
	if err := db.Model(&model.Project{}).Where(whereCondition, token).Count(&count).Error; err != nil {
		return -1, err
	}

	if count > 0 {
		return 1, nil
	}

	return 0, nil
}

func generateUniqueToken(private bool) (token string, err error) {
	for tryCount := 0; tryCount < TOKEN_GEN_RETRY_LIMIT; tryCount++ {
		token = U.RandomLowerAphaNumString(32)
		tokenExists, err := isTokenExist(token, private)
		if err != nil {
			return "", err
		}
		// Break generation, if token doesn't exist already.
		if tokenExists == 0 {
			return token, nil
		}
	}
	return "", fmt.Errorf("token generation failed after %d attempts", TOKEN_GEN_RETRY_LIMIT)
}

func createProject(project *model.Project) (*model.Project, int) {
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
		model.JobsMetadataKeyNextSessionStartTimestamp: U.TimeNowUnix(),
	}
	jobsMetadataJsonb, err := U.EncodeToPostgresJsonb(jobsMetadata)
	if err != nil {
		// Log error and continue to create project.
		logCtx.WithField("jobs_metadata", jobsMetadata).WithError(err).
			Error("Failed to marshal jobs metadata on create project.")
	} else {
		project.JobsMetadata = jobsMetadataJsonb
	}

	// Initialize interaction settings.
	settingsJsonb, err := U.EncodeStructTypeToPostgresJsonb(model.DefaultMarketingPropertiesMap())
	if err != nil {
		// Log error and continue to create project.
		logCtx.WithError(err).Error("Failed to marshal InteractionSettings on create project.")
	} else {
		project.InteractionSettings = *settingsJsonb
	}

	//Initialize default channel group rules
	channelGroupRulesJsonb, err := U.EncodeStructTypeToPostgresJsonb(model.DefaultChannelPropertyRules)
	if err != nil {
		// Log error and continue to create project.
		logCtx.WithError(err).Error("Failed to marshal defaultChannelGroupRules on create project.")
	} else {
		project.ChannelGroupRules = *channelGroupRulesJsonb
	}
	// Add project token before create.
	token, err := generateUniqueToken(false)
	if err != nil {
		logCtx.WithError(err).Error("Failed to generate unique token for project token.")
		return nil, http.StatusInternalServerError
	}
	project.Token = token

	// Add project private token before create.
	privateToken, err := generateUniqueToken(true)
	if err != nil {
		logCtx.WithError(err).Error("Failed to generate unique token for private token.")
		return nil, http.StatusInternalServerError
	}
	project.PrivateToken = privateToken
	if project.TimeZone == "" {
		project.TimeZone = string(U.TimeZoneStringIST)
	}
	_, errCode := time.LoadLocation(string(project.TimeZone))
	if errCode != nil {
		log.WithField("projectId", project.ID).Error("This project hasnt been given with wrong timezone")
		project.TimeZone = string(U.TimeZoneStringIST)
	}

	db := C.GetServices().Db
	if err := db.Create(project).Error; err != nil {
		logCtx.WithError(err).Error("Create project failed.")
		return nil, http.StatusInternalServerError
	}

	return project, http.StatusCreated
}

func (pg *Postgres) UpdateProject(projectId uint64, project *model.Project) int {
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
	if project.ProfilePicture != "" {
		updateFields["profile_picture"] = project.ProfilePicture
	}
	_, errCode := time.LoadLocation(string(project.TimeZone))
	if errCode != nil {
		log.WithField("projectId", project.ID).Error("This project hasnt been given with wrong timezone")
		project.TimeZone = string(U.TimeZoneStringIST)
	}

	if !U.IsEmptyPostgresJsonb(&project.InteractionSettings) {
		updateFields["interaction_settings"] = project.InteractionSettings
	}

	if !U.IsEmptyPostgresJsonb(&project.SalesforceTouchPoints) {
		updateFields["salesforce_touch_points"] = project.SalesforceTouchPoints
	}

	if !U.IsEmptyPostgresJsonb(&project.HubspotTouchPoints) {
		updateFields["hubspot_touch_points"] = project.HubspotTouchPoints
	}

	if !U.IsEmptyPostgresJsonb(&project.ChannelGroupRules) {
		isValid := model.ValidateChannelGroupRules(project.ChannelGroupRules)
		if !isValid {
			return http.StatusInternalServerError
		}
		updateFields["channel_group_rules"] = project.ChannelGroupRules
	}

	err := db.Model(&model.Project{}).Where("id = ?", projectId).Update(updateFields).Error
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to execute query of update project")
		return http.StatusInternalServerError
	}
	return 0
}

func (pg *Postgres) createProjectDependencies(projectID uint64, agentUUID string, createDashboard bool) int {
	logCtx := log.WithField("project_id", projectID)

	// Associated project setting creation with default state.
	defaultAutoTrackState := true
	defaultExcludebotState := true
	defaultDriftIntegrationState := false
	defaultClearBitIntegrationState := false
	_, errCode := createProjectSetting(&model.ProjectSetting{ProjectId: projectID,
		AutoTrack: &defaultAutoTrackState, ExcludeBot: &defaultExcludebotState, IntDrift: &defaultDriftIntegrationState, IntClearBit: &defaultClearBitIntegrationState})
	if errCode != http.StatusCreated {
		logCtx.Error("Create project settings failed on create project dependencies.")
		return errCode
	}

	if ENABLE_DEFAULT_WEB_ANALYTICS {
		if createDashboard {
			errCode = pg.createDefaultDashboardsForProject(projectID, agentUUID)
			if errCode != http.StatusCreated {
				logCtx.Error("Create default dashboards failed on create project dependencies.")
				return errCode
			}
		}
	}

	return http.StatusCreated
}

// CreateProjectWithDependencies separate create method with dependencies to avoid breaking tests.
func (pg *Postgres) CreateProjectWithDependencies(project *model.Project, agentUUID string,
	agentRole uint64, billingAccountID string, createDashboard bool) (*model.Project, int) {

	cProject, errCode := createProject(project)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	errCode = pg.createProjectDependencies(cProject.ID, agentUUID, createDashboard)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	_, errCode = pg.CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
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
func (pg *Postgres) CreateDefaultProjectForAgent(agentUUID string) (*model.Project, int) {
	if agentUUID == "" {
		return nil, http.StatusBadRequest
	}

	errCode := pg.DoesAgentHaveProject(agentUUID)
	if errCode == http.StatusFound {
		return nil, http.StatusConflict
	}
	if errCode != http.StatusNotFound {
		return nil, errCode
	}

	billingAcc, errCode := pg.GetBillingAccountByAgentUUID(agentUUID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).
			Error("CreateDefaultProjectForAgent Failed, billing account error")
		return nil, errCode
	}

	cProject, errCode := pg.CreateProjectWithDependencies(
		&model.Project{Name: model.DefaultProjectName},
		agentUUID, model.ADMIN, billingAcc.ID, true)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	return cProject, http.StatusCreated
}

func (pg *Postgres) GetProject(id uint64) (*model.Project, int) {
	db := C.GetServices().Db
	logCtx := log.WithField("project_id", id)

	var project model.Project
	if err := db.Where("id = ?", id).First(&project).Error; err != nil {
		logCtx.WithError(err).Error("Getting project by id failed")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &project, http.StatusFound
}

func (pg *Postgres) GetProjectByToken(token string) (*model.Project, int) {
	db := C.GetServices().Db

	// Todo(Dinesh): Move to validations.
	cleanToken := strings.TrimSpace(token)
	if len(cleanToken) == 0 {
		return nil, http.StatusBadRequest
	}

	var project model.Project
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

func (pg *Postgres) GetProjectByPrivateToken(privateToken string) (*model.Project, int) {
	db := C.GetServices().Db

	cleanToken := strings.TrimSpace(privateToken)
	if len(cleanToken) == 0 {
		return nil, http.StatusBadRequest
	}

	var project model.Project
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

func (pg *Postgres) GetProjects() ([]model.Project, int) {
	db := C.GetServices().Db

	var projects []model.Project
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

func (pg *Postgres) GetProjectsByIDs(ids []uint64) ([]model.Project, int) {
	if len(ids) == 0 {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var projects []model.Project
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
func (pg *Postgres) GetAllProjectIDs() ([]uint64, int) {
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
func (pg *Postgres) GetNextSessionStartTimestampForProject(projectID uint64) (int64, int) {
	logCtx := log.WithField("project_id", projectID)

	if projectID == 0 {
		logCtx.WithField("project_id", projectID).Error("Invalid args to method.")
		return 0, http.StatusBadRequest
	}

	db := C.GetServices().Db
	rows, err := db.Table("projects").Limit(1).Where("id = ?", projectID).
		Select(fmt.Sprintf("%s->>'%s' as session_start_timestamp",
			model.JobsMetadataColumnName,
			model.JobsMetadataKeyNextSessionStartTimestamp,
		)).Rows()
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

func (pg *Postgres) GetTimezoneForProject(projectID uint64) (U.TimeZoneString, int) {
	project, statusCode := pg.GetProject(projectID)
	if statusCode != http.StatusFound {
		return U.TimeZoneStringIST, statusCode
	}
	if !C.IsMultipleProjectTimezoneEnabled(projectID) {
		return U.TimeZoneStringIST, http.StatusFound
	}
	if project.TimeZone == "" {
		log.WithField("projectId", project.ID).Error("This project has been set with no timezone")
		return U.TimeZoneStringIST, http.StatusFound
	} else {
		_, errCode := time.LoadLocation(string(project.TimeZone))
		if errCode != nil {
			log.WithField("projectId", project.ID).Error("This project has been given with wrong timezone")
			return "", http.StatusNotFound
		}
		return U.TimeZoneString(project.TimeZone), http.StatusFound
	}
}

// UpdateNextSessionStartTimestampForProject - Updates next session start timestamp
// on project jobs metadata.
func (pg *Postgres) UpdateNextSessionStartTimestampForProject(projectID uint64, timestamp int64) int {
	logCtx := log.WithField("project_id", projectID).WithField("timestamp", timestamp)

	if projectID == 0 || timestamp == 0 {
		logCtx.WithField("project_id", projectID).WithField("timestamp", 0).
			Error("Invalid args to method.")
		return http.StatusBadRequest
	}

	// Updating the add_session JSON field directly, to avoid state corruption
	// because of multiple version of JSON being updaetd by multiple jobs simultaneously.
	query := fmt.Sprintf(`UPDATE projects SET jobs_metadata = jobs_metadata - '%s' || '{"%s": %d}' WHERE id = %d`,
		model.JobsMetadataKeyNextSessionStartTimestamp, model.JobsMetadataKeyNextSessionStartTimestamp, timestamp, projectID)
	db := C.GetServices().Db
	err := db.Exec(query).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update next session start timestamp for project.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// GetProjectsToRunForIncludeExcludeString For a given list of include / exclude comma separated strings,
// returns a list of project ids handling * and other cases.
func (pg *Postgres) GetProjectsToRunForIncludeExcludeString(projectIDs, excludeProjectIDs string) []uint64 {
	var projectIDsToRun []uint64
	allProjects, projectIDsMap, excludeProjectIDsMap := C.GetProjectsFromListWithAllProjectSupport(
		projectIDs, excludeProjectIDs)
	projectIDsToRun = C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)

	if allProjects {
		var errCode int
		allProjectIDs, errCode := pg.GetAllProjectIDs()
		if errCode != http.StatusFound {
			return projectIDsToRun
		}
		for _, projectID := range allProjectIDs {
			if _, found := excludeProjectIDsMap[projectID]; !found {
				projectIDsToRun = append(projectIDsToRun, projectID)
			}
		}
	}
	return projectIDsToRun
}

// FillNextSessionStartTimestampForProject - Fills the initial next session start timestamp.
// Postgres only implementation.
func (pg *Postgres) FillNextSessionStartTimestampForProject(projectID uint64, timestamp int64) int {
	logCtx := log.WithField("project_id", projectID).WithField("timestamp", timestamp)

	if projectID == 0 || timestamp == 0 {
		logCtx.WithField("project_id", projectID).WithField("timestamp", 0).
			Error("Invalid args to method.")
		return http.StatusBadRequest
	}

	query := fmt.Sprintf(`UPDATE projects SET jobs_metadata = '{"%s": %d}' WHERE id = %d`,
		model.JobsMetadataKeyNextSessionStartTimestamp, timestamp, projectID)
	db := C.GetServices().Db
	err := db.Exec(query).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update next session start timestamp for project.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (pg *Postgres) GetProjectsWithoutWebAnalyticsDashboard(onlyProjectsMap map[uint64]bool) (projectIds []uint64, errCode int) {

	logCtx := log.WithField("projects", onlyProjectsMap)

	onlyProjectIds := make([]uint64, 0, len(onlyProjectsMap))
	for k := range onlyProjectsMap {
		onlyProjectIds = append(onlyProjectIds, k)
	}

	projectIds = make([]uint64, 0, 0)

	db := C.GetServices().Db
	queryStmnt := "SELECT id FROM projects WHERE id not in (SELECT distinct(project_id) FROM dashboards WHERE dashboards.name = '" + model.DefaultDashboardWebsiteAnalytics + "')"

	//TODO(Maisa): create util function for joining []uint64
	inProjectIds := ""
	for i, opid := range onlyProjectIds {
		inProjectIds = inProjectIds + fmt.Sprintf("%d", opid)

		if i < len(onlyProjectIds)-1 {
			inProjectIds = inProjectIds + ","
		}
	}

	if len(onlyProjectIds) > 0 {
		queryStmnt = queryStmnt + " " + fmt.Sprintf("AND id IN (%s)", inProjectIds)
	}

	rows, err := db.Raw(queryStmnt).Rows()
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to get projectIds on getProjectIdsWithoutWebAnalyticsDashboard.")
		return projectIds, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var projectId uint64

		if err = rows.Scan(&projectId); err != nil {
			logCtx.WithError(err).
				Error("Failed to scan rows on getProjectIdsWithoutWebAnalyticsDashboard.")
			return projectIds, http.StatusInternalServerError
		}

		projectIds = append(projectIds, projectId)
	}

	return projectIds, http.StatusFound
}

func (pg *Postgres) GetProjectIDByToken(token string) (uint64, int) {
	projectID, errCode := model.GetCacheProjectIDByToken(token)
	if errCode == http.StatusFound {
		return projectID, errCode
	}

	project, errCode := pg.GetProjectByToken(token)
	if errCode == http.StatusFound {
		return project.ID, errCode
	}

	return 0, errCode
}
