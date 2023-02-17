package memsql

import (
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const TOKEN_GEN_RETRY_LIMIT = 5
const ENABLE_DEFAULT_WEB_ANALYTICS = false

// Checks for the existence of token already.
func isTokenExist(token string, private bool) (exists int, err error) {
	logFields := log.Fields{
		"token":   token,
		"private": private,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	whereCondition := "token = ?"
	if private {
		whereCondition = "private_token = ?"
	}
	var count int64
	if err := db.Model(&model.Project{}).Where(whereCondition, token).Count(&count).Error; err != nil {
		return -1, err
	}

	if count > 0 {
		return 1, nil
	}

	return 0, nil
}

func generateUniqueToken(private bool) (token string, err error) {
	logFields := log.Fields{
		"private": private,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"project": project,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)

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

	// Initialize default channel group rules.
	channelGroupRulesJsonb, err := U.EncodeStructTypeToPostgresJsonb(model.DefaultChannelPropertyRules)
	if err != nil {
		// Log error and continue to create project.
		logCtx.WithError(err).Error("Failed to marshal defaultChannelGroupRules on create project.")
	} else {
		project.ChannelGroupRules = *channelGroupRulesJsonb
	}

	// Add project token before create.
	// Unique (token).
	token, err := generateUniqueToken(false)
	if err != nil {
		logCtx.WithError(err).Error("Failed to generate unique token for project token.")
		return nil, http.StatusInternalServerError
	}
	project.Token = token
	if project.TimeZone == "" {
		project.TimeZone = string(U.TimeZoneStringIST)
	}
	_, errCode := time.LoadLocation(string(project.TimeZone))
	if errCode != nil {
		log.WithField("projectId", project.ID).Error("This project hasnt been given with wrong timezone")
		project.TimeZone = string(U.TimeZoneStringIST)
	}

	// Add project private token before create.
	// Unique (private_token).
	privateToken, err := generateUniqueToken(true)
	if err != nil {
		logCtx.WithError(err).Error("Failed to generate unique token for private token.")
		return nil, http.StatusInternalServerError
	}
	project.PrivateToken = privateToken

	db := C.GetServices().Db
	var errCount int64
	maxRetries := 700
	for i := 0; i < maxRetries; i++ {
		if err := db.Create(project).Error; err != nil {
			if IsDuplicateRecordError(err) {
				logCtx.WithError(err).Error("Duplicate primary key error for memsql projects")

				errCount++
				if errCount >= int64(maxRetries) {
					logCtx.WithError(err).Error(fmt.Sprintf("Failed to create project after '%d' retries", maxRetries))
					return nil, http.StatusInternalServerError
				} else {
					continue
				}
			}
			logCtx.WithError(err).Error("Create project failed.")
			return nil, http.StatusInternalServerError
		} else {
			break
		}
	}

	project.HubspotTouchPoints = postgres.Jsonb{}
	project.SalesforceTouchPoints = postgres.Jsonb{}
	return project, http.StatusCreated
}

func (store *MemSQL) UpdateProject(projectId int64, project *model.Project) int {
	logFields := log.Fields{
		"project_id": projectId,
		"project":    project,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	logCtx := log.WithFields(logFields)
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
	delProjectTimezoneCacheForID(projectId)
	return 0
}

func (store *MemSQL) createProjectDependencies(projectID int64, agentUUID string) int {
	logFields := log.Fields{
		"project_id": projectID,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	// Associated project setting creation with default state.
	defaultAutoTrackState := true
	defaultAutoFormCapture := true
	defaultAutoFormFillsCapture := true
	defaultExcludebotState := true
	defaultDriftIntegrationState := false
	defaultClearBitIntegrationState := false

	//default timeline config
	timelinesConfig := model.TimelinesConfig{
		DisabledEvents: []string{"Contact Updated", "Campaign Member Updated", "Engagement Meeting Updated", "Engagement Call Updated"}, //Display Names. Used on FE only.
		UserConfig: model.UserConfig{
			TableProps:    []string{U.UP_COUNTRY},
			LeftpaneProps: []string{U.UP_EMAIL, U.UP_COUNTRY, U.UP_PAGE_COUNT},
		},
		AccountConfig: model.AccountConfig{
			TableProps:    []string{U.GP_HUBSPOT_COMPANY_COUNTRY, U.GP_HUBSPOT_COMPANY_NUM_ASSOCIATED_CONTACTS},
			LeftpaneProps: []string{U.GP_HUBSPOT_COMPANY_INDUSTRY, U.GP_HUBSPOT_COMPANY_COUNTRY, U.GP_HUBSPOT_COMPANY_NUMBEROFEMPLOYEES},
			UserProp:      U.UP_USER_ID,
		},
	}
	tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
	if err != nil {
		logCtx.Error("Default Timelines Config Encode Failed.")
	}

	_, errCode := store.createProjectSetting(&model.ProjectSetting{
		ProjectId:            projectID,
		AutoTrack:            &defaultAutoTrackState,
		AutoFormCapture:      &defaultAutoFormCapture,
		AutoCaptureFormFills: &defaultAutoFormFillsCapture,
		ExcludeBot:           &defaultExcludebotState,
		IntDrift:             &defaultDriftIntegrationState,
		IntClearBit:          &defaultClearBitIntegrationState,
		IntegrationBits:      model.DEFAULT_STRING_WITH_ZEROES_32BIT,
		AutoClickCapture:     &model.AutoClickCaptureDefault,
		TimelinesConfig:      tlConfigEncoded,
	})
	if errCode != http.StatusCreated {
		logCtx.Error("Create project settings failed on create project dependencies.")
		return errCode
	}

	if ENABLE_DEFAULT_WEB_ANALYTICS {
		errCode = store.createDefaultDashboardsForProject(projectID, agentUUID)
		if errCode != http.StatusCreated {
			logCtx.Error("Create default dashboards failed on create project dependencies.")
			return errCode
		}
	}

	status, err := store.CreateDefaultFeatureGatesConfigForProject(projectID)
	if status != http.StatusCreated {
		logCtx.Error("Create default feature gates failed on create project dependencies for project ID ", projectID)
		return errCode
	}

	return http.StatusCreated
}

// CreateProjectWithDependencies seperate create method with dependencies to avoid breaking tests.
func (store *MemSQL) CreateProjectWithDependencies(project *model.Project, agentUUID string,
	agentRole uint64, billingAccountID string, createDashboard bool) (*model.Project, int) {
	logFields := log.Fields{
		"project":            project,
		"agent_uuid":         agentUUID,
		"agent_role":         agentRole,
		"billing_account_id": billingAccountID,
		"create_dashboard":   createDashboard,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	cProject, errCode := createProject(project)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	errCode = store.createProjectDependencies(cProject.ID, agentUUID)
	if errCode != http.StatusCreated {
		return nil, errCode
	}
	if createDashboard {
		_, errCode = store.CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
			ProjectID: cProject.ID,
			AgentUUID: agentUUID,
			Role:      agentRole,
		})
		if errCode != http.StatusCreated {
			return nil, errCode
		}
	} else {
		_, errCode = store.CreateProjectAgentMappingWithDependenciesWithoutDashboard(&model.ProjectAgentMapping{
			ProjectID: cProject.ID,
			AgentUUID: agentUUID,
			Role:      agentRole,
		})
		if errCode != http.StatusCreated {
			return nil, errCode
		}
	}
	_, errCode = store.createProjectBillingAccountMapping(project.ID, billingAccountID)
	return cProject, errCode
}

// CreateDefaultProjectForAgent creates project for an agent if there is no project
func (store *MemSQL) CreateDefaultProjectForAgent(agentUUID string) (*model.Project, int) {
	logFields := log.Fields{
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if agentUUID == "" {
		return nil, http.StatusBadRequest
	}

	errCode := store.DoesAgentHaveProject(agentUUID)
	if errCode == http.StatusFound {
		return nil, http.StatusConflict
	}
	if errCode != http.StatusNotFound {
		return nil, errCode
	}

	billingAcc, errCode := store.GetBillingAccountByAgentUUID(agentUUID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).
			Error("CreateDefaultProjectForAgent Failed, billing account error")
		return nil, errCode
	}

	cProject, errCode := store.CreateProjectWithDependencies(
		&model.Project{Name: model.DefaultProjectName},
		agentUUID, model.ADMIN, billingAcc.ID, true)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	return cProject, http.StatusCreated
}

func (store *MemSQL) GetProject(id int64) (*model.Project, int) {
	logFields := log.Fields{
		"id": id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

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

func (store *MemSQL) GetProjectByToken(token string) (*model.Project, int) {
	logFields := log.Fields{
		"token": token,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) GetProjectByPrivateToken(privateToken string) (*model.Project, int) {
	logFields := log.Fields{
		"private_token": privateToken,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) GetProjects() ([]model.Project, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
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
func isValidProjectScope(id int64) bool {
	logFields := log.Fields{
		"id": id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return id != 0
}

func (store *MemSQL) GetProjectsByIDs(ids []int64) ([]model.Project, int) {
	logFields := log.Fields{
		"ids": ids,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
func (store *MemSQL) GetAllProjectIDs() ([]int64, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	projectIds := make([]int64, 0, 0)

	db := C.GetServices().Db
	rows, err := db.Raw("SELECT id FROM projects").Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get all project ids.")
		return projectIds, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var projectId int64
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
func (store *MemSQL) GetNextSessionStartTimestampForProject(projectID int64) (int64, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 {
		logCtx.WithField("project_id", projectID).Error("Invalid args to method.")
		return 0, http.StatusBadRequest
	}

	db := C.GetServices().Db
	rows, err := db.Table("projects").Limit(1).Where("id = ?", projectID).
		Select(fmt.Sprintf("%s::%s as session_start_timestamp",
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

func (store *MemSQL) GetTimezoneForProject(projectID int64) (U.TimeZoneString, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	project, statusCode := store.GetProject(projectID)
	if statusCode != http.StatusFound {
		return U.TimeZoneStringIST, statusCode
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
		return U.TimeZoneString(project.TimeZone), statusCode
	}
}

func (store *MemSQL) GetTimezoneByIDWithCache(projectID int64) (U.TimeZoneString, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resTimezone, errCode := getProjectTimezoneCache(projectID)
	if errCode == http.StatusFound {
		return U.TimeZoneString(resTimezone), http.StatusFound
	}

	resTimezone, statusCode := store.GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		return "", statusCode
	}

	// add to cache.
	setCacheForProjectTimezone(projectID, resTimezone)

	_, errCode2 := time.LoadLocation(string(resTimezone))
	if errCode2 != nil {
		log.WithField("projectId", projectID).Error("This project has been given with wrong timezone")
		return "", http.StatusNotFound
	}
	return resTimezone, statusCode
}

// UpdateNextSessionStartTimestampForProject - Updates next session start timestamp
// on project jobs metadata.
func (store *MemSQL) UpdateNextSessionStartTimestampForProject(projectID int64, timestamp int64) int {
	logFields := log.Fields{
		"project_id": projectID,
		"timestamp":  timestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || timestamp == 0 {
		logCtx.WithField("project_id", projectID).WithField("timestamp", 0).
			Error("Invalid args to method.")
		return http.StatusBadRequest
	}

	// Updating the add_session JSON field directly, to avoid state corruption
	// because of multiple version of JSON being updaetd by multiple jobs simultaneously.
	query := fmt.Sprintf(`UPDATE projects SET jobs_metadata = JSON_SET_DOUBLE(jobs_metadata, '%s', %d) WHERE id = %d`,
		model.JobsMetadataKeyNextSessionStartTimestamp, timestamp, projectID)
	db := C.GetServices().Db
	rows, err := db.Raw(query).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to update next session start timestamp for project.")
		return http.StatusInternalServerError
	}
	defer rows.Close()

	return http.StatusAccepted
}

// GetProjectsToRunForIncludeExcludeString For a given list of include / exclude comma separated strings,
// returns a list of project ids handling * and other cases.
func (store *MemSQL) GetProjectsToRunForIncludeExcludeString(projectIDs, excludeProjectIDs string) []int64 {
	logFields := log.Fields{
		"project_ids":         projectIDs,
		"exclude_project_ids": excludeProjectIDs,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var projectIDsToRun []int64
	allProjects, projectIDsMap, excludeProjectIDsMap := C.GetProjectsFromListWithAllProjectSupport(
		projectIDs, excludeProjectIDs)
	projectIDsToRun = C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)

	if allProjects {
		var errCode int
		allProjectIDs, errCode := store.GetAllProjectIDs()
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
func (store *MemSQL) FillNextSessionStartTimestampForProject(projectID int64, timestamp int64) int {
	logFields := log.Fields{
		"project_id": projectID,
		"timestamp":  timestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || timestamp == 0 {
		logCtx.WithField("project_id", projectID).WithField("timestamp", 0).
			Error("Invalid args to method.")
		return http.StatusBadRequest
	}

	query := fmt.Sprintf(`UPDATE projects SET jobs_metadata = '{"%s": %d}' WHERE id = %d`,
		model.JobsMetadataKeyNextSessionStartTimestamp, timestamp, projectID)
	db := C.GetServices().Db
	rows, err := db.Raw(query).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to update next session start timestamp for project.")
		return http.StatusInternalServerError
	}
	defer rows.Close()

	return http.StatusAccepted
}

func (store *MemSQL) GetProjectsWithoutWebAnalyticsDashboard(onlyProjectsMap map[int64]bool) (projectIds []int64, errCode int) {
	logFields := log.Fields{
		"only_projects_map": onlyProjectsMap,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	onlyProjectIds := make([]int64, 0, len(onlyProjectsMap))
	for k := range onlyProjectsMap {
		onlyProjectIds = append(onlyProjectIds, k)
	}

	projectIds = make([]int64, 0, 0)

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
		var projectId int64

		if err = rows.Scan(&projectId); err != nil {
			logCtx.WithError(err).
				Error("Failed to scan rows on getProjectIdsWithoutWebAnalyticsDashboard.")
			return projectIds, http.StatusInternalServerError
		}

		projectIds = append(projectIds, projectId)
	}

	return projectIds, http.StatusFound
}

func (store *MemSQL) GetProjectIDByToken(token string) (int64, int) {
	logFields := log.Fields{
		"token": token,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	projectID, errCode := model.GetCacheProjectIDByToken(token)
	if errCode == http.StatusFound {
		return projectID, errCode
	}

	project, errCode := store.GetProjectByToken(token)
	if errCode != http.StatusFound {
		return 0, errCode
	}

	model.SetCacheProjectIDByToken(token, project.ID)
	return project.ID, errCode
}

func getProjectTimezoneCache(projectID int64) (U.TimeZoneString, int) {
	logCtx := log.WithField("projectID", projectID)
	if projectID == 0 {
		return "", http.StatusBadRequest
	}

	key, err := getProjectTimezoneCacheKey(projectID)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get project key on getProjectCache")
		return "", http.StatusInternalServerError
	}

	resTimezone, err := cacheRedis.Get(key)
	if err != nil {
		if err == redis.ErrNil {
			return "", http.StatusNotFound
		}

		logCtx.WithError(err).Error(
			"Failed to get json from cache on getProjectCache.")
		return "", http.StatusInternalServerError
	}

	return U.TimeZoneString(resTimezone), http.StatusFound
}

func setCacheForProjectTimezone(projectID int64, timezoneString U.TimeZoneString) int {
	logFields := log.Fields{"ID": projectID}
	logCtx := log.WithFields(logFields)

	key, err := getProjectTimezoneCacheKey(projectID)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get project settings by token cache key on setCacheProject")
		return http.StatusInternalServerError
	}

	var expiryInSecs float64 = 60 * 60
	err = cacheRedis.Set(key, string(timezoneString), expiryInSecs)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache on setCacheProjectSetting")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func delProjectTimezoneCacheForID(ID int64) int64 {
	logCtx := log.WithField("ID", ID)
	if ID == 0 {
		return http.StatusBadRequest
	}
	key, err := getProjectTimezoneCacheKey(ID)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get project settings by token cache key on delCacheProjectSetting")
		return http.StatusInternalServerError
	}

	err = cacheRedis.Del(key)
	if err != nil && err != redis.ErrNil {
		logCtx.WithError(err).Error("Failed to del cache on delCacheProjectSetting")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func getProjectTimezoneCacheKey(ID int64) (*cacheRedis.Key, error) {
	prefix := "project_tz"
	return cacheRedis.NewKeyWithProjectUID(string(ID), prefix, "")
}
