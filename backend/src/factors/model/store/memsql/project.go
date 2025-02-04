package memsql

import (
	billing "factors/billing/chargebee"
	"factors/cache"
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
		updatedTimeZone := project.TimeZone
		_, errCode := time.LoadLocation(string(updatedTimeZone))
		if errCode != nil {
			log.WithField("projectId", project.ID).Error("This project hasnt been given with wrong timezone")
			updatedTimeZone = string(U.TimeZoneStringIST)
		}
		updateFields["time_zone"] = updatedTimeZone
	}
	if project.ProfilePicture != "" {
		updateFields["profile_picture"] = project.ProfilePicture
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

	if project.EnableBilling {
		updateFields["enable_billing"] = true

	}
	if project.BillingAccountID != "" {
		updateFields["billing_account_id"] = project.BillingAccountID
	}

	if project.BillingSubscriptionID != "" {
		updateFields["billing_subscription_id"] = project.BillingSubscriptionID
	}

	if !project.BillingLastSyncedAt.IsZero() {
		updateFields["billing_last_synced_at"] = project.BillingLastSyncedAt
	}

	if project.ClearbitDomain != "" {
		updateFields["clearbit_domain"] = project.ClearbitDomain
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

func (store *MemSQL) createProjectDependencies(projectID int64, agentUUID string, billingEnabled bool) int {
	logFields := log.Fields{
		"project_id": projectID,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	// Associated project setting creation with default state.
	defaultAutoTrackState := true
	defaultAutoFormCapture := true
	defaultExcludebotState := true
	defaultAutoFormFillsCapture := false
	defaultDriftIntegrationState := false
	defaultClearBitIntegrationState := false
	deafultSixSignalIntegrationState := true

	tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(model.DefaultTimelineConfig)
	if err != nil {
		logCtx.WithError(err).Error("Default Timelines Config Encode Failed.")
	}

	//default filter_ips
	filterIps := model.FilterIps{
		BlockIps: []string{},
	}
	filtersIpsEncoded, err := U.EncodeStructTypeToPostgresJsonb(filterIps)
	if err != nil {
		logCtx.WithError(err).Error("Default Filter IPs Encode Failed.")
	}
	_, errCode := store.createProjectSetting(&model.ProjectSetting{
		ProjectId:              projectID,
		AutoTrack:              &defaultAutoTrackState,
		AutoFormCapture:        &defaultAutoFormCapture,
		AutoCaptureFormFills:   &defaultAutoFormFillsCapture,
		ExcludeBot:             &defaultExcludebotState,
		IntDrift:               &defaultDriftIntegrationState,
		IntClearBit:            &defaultClearBitIntegrationState,
		IntFactorsSixSignalKey: &deafultSixSignalIntegrationState,
		IntegrationBits:        model.DEFAULT_STRING_WITH_ZEROES_32BIT,
		AutoClickCapture:       &model.AutoClickCaptureDefault,
		TimelinesConfig:        tlConfigEncoded,
		FilterIps:              filtersIpsEncoded,
	})
	if errCode != http.StatusCreated {
		logCtx.WithField("err_code", errCode).Error("Create project settings failed on create project dependencies.")
		return errCode
	}

	if ENABLE_DEFAULT_WEB_ANALYTICS {
		errCode = store.createDefaultDashboardsForProject(projectID, agentUUID)
		if errCode != http.StatusCreated {
			logCtx.WithField("err_code", errCode).Error("Create default dashboards failed on create project dependencies.")
			return errCode
		}
	}

	// creating free subscription
	if billingEnabled {
		// creating subscription on chargebee
		agent, status := store.GetAgentByUUID(agentUUID)
		if status != http.StatusFound {
			logCtx.WithField("err_code", status).Error("Failed to get agent while creating default subscription")
			return errCode
		}

		subscription, status, err := billing.CreateChargebeeSubscriptionForCustomer(projectID, agent.BillingCustomerID, model.FREE_PLAN_ITEM_PRICE_ID)
		if err != nil || status != http.StatusCreated {
			logCtx.WithField("err_code", status).WithError(err).Error("Failed to create default subscription for agent")
			return errCode
		}

		bAccount, status := store.GetBillingAccountByAgentUUID(agentUUID)
		if status != http.StatusFound {
			logCtx.WithField("err_code", status).WithError(err).Error("Failed to get billing account agent")
			return errCode
		}
		// update subscription id on project
		updatedProject := model.Project{
			BillingAccountID:      bAccount.ID,
			EnableBilling:         true,
			BillingSubscriptionID: subscription.Id,
			BillingLastSyncedAt:   time.Now(),
		}

		status = store.UpdateProject(projectID, &updatedProject)
		if status != 0 {
			logCtx.WithField("err_code", status).WithError(err).Error("Failed to update subscription id on project")
			return http.StatusInternalServerError
		}

	}

	// inserting project into free plan by default
	status, err := store.CreateDefaultProjectPlanMapping(projectID, model.DEFAULT_PLAN_ID, model.DEFAULT_PLAN_ITEM_PRICE_ID)
	if status != http.StatusCreated {
		logCtx.WithField("err", err).Error("Create default project plan mapping failed on create project dependencies for project ID ", projectID)
		return errCode
	}

	// Create Default Segment - Visited Website
	status, err = store.CreateDefaultSegment(projectID, model.PropertyEntityUser, false)
	if status != http.StatusCreated {
		log.WithError(err).Error("Failed to create default segment - \"Visited Website\".")
	}

	err = store.OnboardingAccScoring(projectID)
	if err != nil {
		log.WithError(err).WithField("projectId", projectID).Error("Unable to add default weights")

	}

	// statusCode := store.CreatePredefinedDashboards(projectID, agentUUID)
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

	agent, errCode := store.GetAgentByUUID(agentUUID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).
			Error("CreateProjectForAgent Failed, get agent by uuid error")
		return nil, errCode
	}

	project.EnableBilling = agent.BillingCustomerID != ""

	errCode = store.createProjectDependencies(cProject.ID, agentUUID, project.EnableBilling)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	_, errCode = store.CreateAllBoardsDashboardFolder(cProject.ID)
	if errCode != http.StatusCreated {
		log.WithField("err_code", errCode).Error("CreateProject Failed, All Boards creation failed.")
		return nil, errCode
	}

	//Generate Website Visitor Identification Dashboard
	visitorIdentificationTemplate, errCode := store.SearchTemplateWithType(model.WEBSITE_VISITOR_IDENTIFICATION)
	if errCode != http.StatusFound {
		log.WithField("errCode", errCode).Error("Failed to fetch dashboard template id for website visitor identification template")
	}

	_, errCode, errMsg := store.GenerateDashboardFromTemplate(cProject.ID, agentUUID, visitorIdentificationTemplate.ID)
	if errCode != http.StatusCreated {
		log.WithField("error", errMsg).Error("Website Visitor Identification dashboard creation failed")
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

	agent, errCode := store.GetAgentByUUID(agentUUID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).
			Error("CreateDefaultProjectForAgent Failed, get agent by uuid error")
		return nil, errCode
	}

	enableBilling := agent.BillingCustomerID != ""

	cProject, errCode := store.CreateProjectWithDependencies(
		&model.Project{Name: model.DefaultProjectName, EnableBilling: enableBilling, BillingAccountID: billingAcc.ID},
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
	if err := db.Limit(1).Where("id = ?", id).Find(&project).Error; err != nil {
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

func (store *MemSQL) GetProjectsInfoByIDs(ids []int64) ([]model.ProjectInfo, int) {
	logFields := log.Fields{
		"ids": ids,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if len(ids) == 0 {
		return nil, http.StatusBadRequest
	}

	var projects []model.ProjectInfo

	db := C.GetServices().Db
	if err := db.Table("projects").Limit(len(ids)).Where(ids).Find(&projects).Error; err != nil {
		log.WithError(err).Error("Getting projects info using ids failed")
		return nil, http.StatusInternalServerError
	}

	if len(projects) == 0 {
		return projects, http.StatusNoContent
	}

	// add login method to project infos

	for idx, project := range projects {
		settings, status := store.GetProjectSetting(project.ID)
		if status != http.StatusFound {
			log.WithField("project_id", project.ID).Error("Failed to fetch project settings")
			return nil, http.StatusInternalServerError
		}
		projects[idx].LoginMethod = settings.SSOState
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

// GetProjectIDsWithSixSignalEnabled Gets the project_ids of projects for which 6Signal is enabled.
func (store *MemSQL) GetProjectIDsWithSixSignalEnabled() []int64 {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	projectIds := make([]int64, 0, 0)

	db := C.GetServices().Db
	rows, err := db.Raw("SELECT project_id from project_settings WHERE int_factors_six_signal_key=1 OR int_client_six_signal_key=1").Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get projectIds for which sixsignal is enabled. ")
		return projectIds
	}
	defer rows.Close()

	for rows.Next() {
		var projectId int64
		if err = rows.Scan(&projectId); err != nil {
			log.WithError(err).Error("Failed to get project ids for which sixsignal is enabled. Scanning failed.")
			return projectIds
		}

		projectIds = append(projectIds, projectId)
	}

	return projectIds
}

/*
GetProjectsToRunForVisitorIdentificationReport

1. An external function GetProjectsFromListWithAllProjectSupport() is called which returns true if projectIdFlag is '*'.

2. if allIdentificationEnabledProjects is true, projects id is fetched for the projects for which FEATURE_SIX_SIGNAL_REPORT is enabled.

3. else it returns the list of projectIds from the projectIDsMap returned from GetProjectsFromListWithAllProjectSupport()
*/
func (store *MemSQL) GetProjectsToRunForVisitorIdentificationReport(projectIdFlag, excludeProjectIDs string) []int64 {

	var projectIDsToRun []int64
	allIdentificationEnabledProjects, projectIDsMap, excludeProjectIDsMap := C.GetProjectsFromListWithAllProjectSupport(
		projectIdFlag, excludeProjectIDs)
	projectIDsToRun = C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)

	if allIdentificationEnabledProjects {
		projectIDs, err := store.GetAllProjectsWithFeatureEnabled(model.FEATURE_SIX_SIGNAL_REPORT, false)
		if err != nil {
			return projectIDsToRun
		}

		for _, projectID := range projectIDs {
			if _, found := excludeProjectIDsMap[projectID]; !found {
				projectIDsToRun = append(projectIDsToRun, projectID)
			}
		}
	}
	return projectIDsToRun
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
			log.WithField("projectId", project.ID).WithField("err_code", errCode).Error("This project has been given with wrong timezone")
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
		log.WithField("projectId", projectID).WithField("err_code", errCode2).Error("This project has been given with wrong timezone")
		return "", http.StatusNotFound
	}
	return resTimezone, statusCode
}

func (store *MemSQL) GetIDAndTimezoneForAllProjects() ([]model.Project, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	db := C.GetServices().Db

	var projects []model.Project
	if err := db.Find(&projects).Error; err != nil {
		log.WithError(err).Error("Getting all projects failed")
		return nil, http.StatusInternalServerError
	}

	for index := range projects {
		if projects[index].TimeZone == "" {
			projects[index].TimeZone = string(U.TimeZoneStringIST)
		}
	}
	return projects, http.StatusFound
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

func (store *MemSQL) GetProjectIDByBillingSubscriptionID(id string) (int64, int) {
	logFields := log.Fields{
		"subscription_id": id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var project model.Project
	if err := db.Where("billing_subscription_id = ?", id).First(&project).Error; err != nil {
		log.WithFields(logFields).WithError(err).Error("Getting project by subscription id failed")
		if gorm.IsRecordNotFoundError(err) {
			return 0, http.StatusNotFound
		}
		return 0, http.StatusInternalServerError
	}
	return project.ID, http.StatusFound
}

// // TODO Add default positions and sizes. Response is not giving all dashboards.
// // TODO Check if dashboards are being picked in caching. They shouldnt be.
// func (store *MemSQL) createPredefinedDashboards(projectID int64, predefinedDashboards []model.PredefinedDashboard, agentUUID string) ([]*model.Dashboard, int) {
// 	dashboards := make([]*model.Dashboard, 0)
// 	errCode := http.StatusCreated
// 	for _, predefinedDashboard := range predefinedDashboards {
// 		_, errCode := store.CreateDashboard(
// 			projectID, agentUUID,
// 			&model.Dashboard{
// 				Name:        predefinedDashboard.DisplayName,
// 				Description: predefinedDashboard.Description,
// 				Type:        model.DashboardTypeProjectVisible,
// 				Class:       model.DashboardClassPredefined,
// 				InternalID:  predefinedDashboard.InternalID,
// 			})
// 		if errCode != http.StatusCreated {
// 			return dashboards, errCode
// 		}
// 	}
// 	return dashboards, errCode
// }

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

func getProjectTimezoneCacheKey(ID int64) (*cache.Key, error) {
	prefix := "project_tz"
	return cache.NewKeyWithProjectUID(string(ID), prefix, "")
}

func (store *MemSQL) OnboardingAccScoring(projectId int64) error {

	logCtx := log.WithField("project id", projectId)
	// update project with default account score weights
	weights := model.GetDefaultAccScoringWeights()
	dedupWeights, _, err := model.UpdateWeights(projectId, weights, "adding acount scoring - Onboarding project")
	if err != nil {
		logCtx.WithError(err).Errorf("Unable to add default weights - onboarding")
	} else {
		err = store.UpdateAccScoreWeights(projectId, dedupWeights)
		if err != nil {
			logCtx.WithField("project id", projectId).WithError(err).Errorf("Unable to add default weights -onboarding")
		}
	}

	return nil
}

func (store *MemSQL) CreateAllBoardsDashboardFolder(projectId int64) (*model.DashboardFolders, int) {

	logCtx := log.WithField("project id", projectId)
	folder := &model.DashboardFolders{Name: model.ALL_BOARDS_FOLDER, IsDefaultFolder: true}

	folder, errCode := store.CreateDashboardFolder(projectId, folder)
	if errCode != http.StatusCreated {
		logCtx.Error("Failed creating All Boards folder.")
		return nil, errCode
	}
	return folder, http.StatusCreated

}
