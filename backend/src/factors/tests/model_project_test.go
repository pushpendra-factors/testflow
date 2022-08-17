package tests

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"math"
	"net/http"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndGetProject(t *testing.T) {
	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	billingAccount, errCode := store.GetStore().GetBillingAccountByAgentUUID(agent.UUID)
	assert.Equal(t, http.StatusFound, errCode)

	start := time.Now()

	// Test successful create project.
	projectName := U.RandomLowerAphaNumString(15)
	project, errCode := store.GetStore().CreateProjectWithDependencies(&model.Project{Name: projectName}, agent.UUID, model.ADMIN, billingAccount.ID, true)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, project.ID > 0)
	assert.Equal(t, projectName, project.Name)
	assert.Equal(t, 32, len(project.Token))
	assert.True(t, project.CreatedAt.After(start))
	assert.True(t, project.UpdatedAt.After(start))
	assert.Equal(t, project.CreatedAt, project.UpdatedAt)

	// Test update
	interactionSettings := model.InteractionSettings{}
	interactionSettings.UTMMappings = make(map[string][]string)
	interactionSettings.UTMMappings["Hello"] = []string{"World"}
	val1, _ := U.EncodeStructTypeToPostgresJsonb(interactionSettings)

	/*salesforceTouchPoint := model.SalesforceTouchPoints{}
	salesforceTouchPoint.TouchPointRules = make(map[string][]model.SFTouchPointRule)
	salesforceTouchPoint.TouchPointRules["Sales"] = []model.SFTouchPointRule{model.SFTouchPointRule{TouchPointTimeRef: "Force"}}
	val2, _ := U.EncodeStructTypeToPostgresJsonb(salesforceTouchPoint)

	hubspotTouchPoint := model.HubspotTouchPoints{}
	hubspotTouchPoint.TouchPointRules = make(map[string][]model.HSTouchPointRule)
	hubspotTouchPoint.TouchPointRules["Hub"] = []model.HSTouchPointRule{model.HSTouchPointRule{TouchPointTimeRef: "Spot"}}
	val3, _ := U.EncodeStructTypeToPostgresJsonb(hubspotTouchPoint)*/

	errCode = store.GetStore().UpdateProject(project.ID,
		&model.Project{InteractionSettings: *val1})

	assert.Equal(t, errCode, 0)
	getProject, errCode := store.GetStore().GetProject(project.ID)
	assert.Equal(t, http.StatusFound, errCode)

	valUpdated1 := model.InteractionSettings{}
	_ = U.DecodePostgresJsonbToStructType(&getProject.InteractionSettings, &valUpdated1)
	assert.Equal(t, valUpdated1.UTMMappings["Hello"], []string{"World"})

	/*valUpdated2 := model.SalesforceTouchPoints{}
	_ = U.DecodePostgresJsonbToStructType(&getProject.SalesforceTouchPoints, &valUpdated2)
	assert.Equal(t, valUpdated2.TouchPointRules["Sales"], []model.SFTouchPointRule{model.SFTouchPointRule{TouchPointTimeRef: "Force"}})

	valUpdated3 := model.HubspotTouchPoints{}
	_ = U.DecodePostgresJsonbToStructType(&getProject.HubspotTouchPoints, &valUpdated3)
	assert.Equal(t, valUpdated3.TouchPointRules["Hub"], []model.HSTouchPointRule{model.HSTouchPointRule{TouchPointTimeRef: "Spot"}})*/

	// Test token is overwritten and cannot be provided.
	previousProjectId := project.ID
	// Random Token.
	providedToken := U.RandomLowerAphaNumString(32)
	// Reusing the same name. Name is not meant to be unique.
	project, errCode = store.GetStore().CreateProjectWithDependencies(&model.Project{Name: projectName, Token: providedToken}, agent.UUID, model.ADMIN, billingAccount.ID, true)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, project.ID > previousProjectId)
	assert.Equal(t, projectName, project.Name)
	assert.Equal(t, 32, len(project.Token))
	assert.NotEqual(t, providedToken, project.Token)
	assert.True(t, project.CreatedAt.After(start))
	assert.True(t, project.UpdatedAt.After(start))
	assert.Equal(t, project.CreatedAt, project.UpdatedAt)
	// Test Get Project on the created one.
	getProject, errCode = store.GetStore().GetProject(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(project.CreatedAt.Sub(getProject.CreatedAt).Seconds()) < 0.1)
	assert.True(t, math.Abs(project.UpdatedAt.Sub(getProject.UpdatedAt).Seconds()) < 0.1)
	project.CreatedAt = time.Time{}
	project.UpdatedAt = time.Time{}
	getProject.CreatedAt = time.Time{}
	getProject.UpdatedAt = time.Time{}
	assert.Equal(t, project.ID, getProject.ID)

	// Test Get Project on random id.
	var randomId int64 = int64(U.RandomUint64WithUnixNano())
	getProject, errCode = store.GetStore().GetProject(randomId)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, getProject)

	// Test Bad input by providing id.
	// Reusing the same name. Name is not meant to be unique.
	project, errCode = store.GetStore().CreateProjectWithDependencies(&model.Project{Name: projectName, ID: previousProjectId + 10}, agent.UUID, model.ADMIN, billingAccount.ID, true)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, project)

	// Test Get Project by a token.
	// Bad input.
	project, errCode = store.GetStore().GetProjectByToken("")
	assert.Equal(t, http.StatusBadRequest, errCode)

	// RandomInput
	project, errCode = store.GetStore().GetProjectByToken(U.RandomLowerAphaNumString(32))
	assert.Equal(t, http.StatusNotFound, errCode)

	// Check corresponding project returned with token.
	project, errCode = store.GetStore().CreateProjectWithDependencies(&model.Project{Name: projectName}, agent.UUID, model.ADMIN, billingAccount.ID, true)
	rProject, rErrCode := store.GetStore().GetProjectByToken(project.Token)
	assert.Equal(t, http.StatusFound, rErrCode)
	assert.Equal(t, project.ID, rProject.ID)

	// Test CreateProjectWithDependencies
	start = time.Now()
	projectWithDeps, errCode := store.GetStore().CreateProjectWithDependencies(&model.Project{Name: projectName}, agent.UUID, model.ADMIN, billingAccount.ID, true)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, projectWithDeps.ID > 0)
	assert.Equal(t, 32, len(projectWithDeps.Token))
	assert.True(t, projectWithDeps.CreatedAt.After(start))
	assert.True(t, projectWithDeps.UpdatedAt.After(start))
	assert.Equal(t, projectWithDeps.CreatedAt, projectWithDeps.UpdatedAt)

	// Test depedencies creation - ProjectSettings.
	ps, errCode := store.GetStore().GetProjectSetting(projectWithDeps.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, ps)
	assert.True(t, *ps.AutoTrack)
	assert.True(t, *ps.ExcludeBot)
}

func TestDBGetProjectByIDs(t *testing.T) {
	t.Run("NoProjects", func(t *testing.T) {
		randIds := []int64{
			int64(U.RandomUint64WithUnixNano()),
			int64(U.RandomUint64WithUnixNano()),
		}
		proj, errCode := store.GetStore().GetProjectsByIDs(randIds)
		assert.Equal(t, 0, len(proj))
		assert.Equal(t, http.StatusNoContent, errCode)
	})

	t.Run("MissingParams", func(t *testing.T) {
		randIds := []int64{}
		_, errCode := store.GetStore().GetProjectsByIDs(randIds)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("FetchProjects", func(t *testing.T) {
		noOfProjects := int(U.RandomUint64()%5 + 2)
		idsToFetch := make([]int64, 0, 0)
		for i := 0; i < noOfProjects; i++ {
			project, err := SetupProjectReturnDAO()
			assert.Nil(t, err)
			idsToFetch = append(idsToFetch, project.ID)
		}
		retProjects, errCode := store.GetStore().GetProjectsByIDs(idsToFetch)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, noOfProjects, len(retProjects))
	})
}

func TestCreateDefaultProjectForAgent(t *testing.T) {
	t.Run("CreateDefaultProjectForAgent", func(t *testing.T) {
		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
		assert.Equal(t, http.StatusCreated, errCode)

		project, errCode := store.GetStore().CreateDefaultProjectForAgent(agent.UUID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, project)
		assert.Equal(t, model.DefaultProjectName, project.Name)
		assert.NotNil(t, project.InteractionSettings)
	})

	t.Run("CreateDefaultProjectForAgent:AgentAlreadyWithProject", func(t *testing.T) {
		_, agent, err := SetupProjectWithAgentDAO()
		assert.Nil(t, err)

		// should not create if agent has a project associated.
		_, errCode := store.GetStore().CreateDefaultProjectForAgent(agent.UUID)
		assert.Equal(t, http.StatusConflict, errCode)
	})

	t.Run("CreateDefaultProjectForAgent:Invalid", func(t *testing.T) {
		project, errCode := store.GetStore().CreateDefaultProjectForAgent("")
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, project)
	})
}

func TestNextSessionStartTimestampForProject(t *testing.T) {
	// Create project without resetting next session start timestamp.
	createAgentParams := &model.CreateAgentParams{Agent: &model.Agent{FirstName: getRandomName(),
		LastName: getRandomName(), Email: getRandomEmail(), Phone: "123456789"}, PlanCode: model.FreePlanCode}
	createdAgent, errCode := store.GetStore().CreateAgentWithDependencies(createAgentParams)
	assert.Equal(t, http.StatusCreated, errCode)
	billingAccount, errCode := store.GetStore().GetBillingAccountByAgentUUID(createdAgent.Agent.UUID)
	assert.Equal(t, http.StatusFound, errCode)
	project, errCode := store.GetStore().CreateProjectWithDependencies(&model.Project{Name: U.RandomLowerAphaNumString(15)},
		createdAgent.Agent.UUID, model.ADMIN, billingAccount.ID, true)
	assert.Equal(t, http.StatusCreated, errCode)

	assert.NotNil(t, project.JobsMetadata)
	jobsMetadata, err := U.DecodePostgresJsonb(project.JobsMetadata)
	assert.Nil(t, err)
	assert.NotNil(t, (*jobsMetadata)[model.JobsMetadataKeyNextSessionStartTimestamp])
	assert.NotZero(t, (*jobsMetadata)[model.JobsMetadataKeyNextSessionStartTimestamp])
	timestamp := (*jobsMetadata)[model.JobsMetadataKeyNextSessionStartTimestamp]

	gotTimestamp, errCode := store.GetStore().GetNextSessionStartTimestampForProject(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, timestamp, float64(gotTimestamp))

	newTimestamp := gotTimestamp + 10
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, newTimestamp)
	assert.Equal(t, http.StatusAccepted, errCode)

	gotTimestampAfterUpdate, errCode := store.GetStore().GetNextSessionStartTimestampForProject(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, newTimestamp, gotTimestampAfterUpdate)
}

func TestProjectSettingIngestionTimezoneFetch(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	projectSetting := model.ProjectSetting{}
	db := C.GetServices().Db
	db.Table("project_settings").Where("project_id = ?", project.ID).First(&projectSetting)

	projectSetting.IntGoogleIngestionTimezone = "Australia"
	db.Save(projectSetting)
	_, projectSettings, _ := store.GetStore().GetFacebookEnabledIDsAndProjectSettingsForProject([]int64{project.ID})
	log.Warn(projectSettings)
}
