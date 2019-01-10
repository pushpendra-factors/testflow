package tests

import (
	M "factors/model"
	U "factors/util"
	"math"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndGetEventName(t *testing.T) {
	// Initialize a project for the event.
	randomProjectName := U.RandomLowerAphaNumString(15)
	project, errCode := M.CreateProjectWithDependencies(&M.Project{Name: randomProjectName})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, project)
	projectId := project.ID

	start := time.Now()

	// Test successful create eventName.
	eventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{Name: "test_event", ProjectId: projectId})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, projectId, eventName.ProjectId)
	assert.True(t, eventName.CreatedAt.After(start))
	// Trying to create again should return the old one.
	expectedEventName := &M.EventName{}
	copier.Copy(expectedEventName, eventName)
	retryEventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{Name: "test_event", ProjectId: projectId})
	assert.Equal(t, http.StatusConflict, errCode)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(expectedEventName.CreatedAt.Sub(retryEventName.CreatedAt).Seconds()) < 0.1)
	assert.True(t, math.Abs(expectedEventName.UpdatedAt.Sub(retryEventName.UpdatedAt).Seconds()) < 0.1)
	expectedEventName.CreatedAt = time.Time{}
	retryEventName.CreatedAt = time.Time{}
	expectedEventName.UpdatedAt = time.Time{}
	retryEventName.UpdatedAt = time.Time{}
	assert.Equal(t, expectedEventName, retryEventName)
	// Test Get EventName on the created one.
	expectedEventName = &M.EventName{}
	copier.Copy(expectedEventName, eventName)
	retEventName, errCode := M.GetEventName(expectedEventName.Name, projectId)
	assert.Equal(t, http.StatusFound, errCode)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(expectedEventName.CreatedAt.Sub(retEventName.CreatedAt).Seconds()) < 0.1)
	assert.True(t, math.Abs(expectedEventName.UpdatedAt.Sub(retEventName.UpdatedAt).Seconds()) < 0.1)
	expectedEventName.CreatedAt = time.Time{}
	retEventName.CreatedAt = time.Time{}
	expectedEventName.UpdatedAt = time.Time{}
	retEventName.UpdatedAt = time.Time{}
	assert.Equal(t, expectedEventName, retEventName)

	// Test Get Event on non existent name.
	retEventName, errCode = M.GetEventName("non_existent_event", projectId)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEventName)

	// Test Get Event with only name.
	retEventName, errCode = M.GetEventName(eventName.Name, 0)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, retEventName)

	// Test Get Event with only projectId.
	retEventName, errCode = M.GetEventName("", projectId)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, retEventName)

	// Test Validate type on CreateOrGetUserCreatedEventName.
	randomName := U.RandomLowerAphaNumString(10)
	ucEventName := &M.EventName{Name: randomName, ProjectId: project.ID}
	retEventName, errCode = M.CreateOrGetUserCreatedEventName(ucEventName)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, M.TYPE_USER_CREATED_EVENT_NAME, retEventName.Type)

	// Test Duplicate creation of user created event name. Should be unique by project.
	duplicateEventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{Name: randomName, ProjectId: project.ID})
	assert.Equal(t, http.StatusConflict, errCode) // Should return conflict with the conflicted object.
	assert.Equal(t, M.TYPE_USER_CREATED_EVENT_NAME, retEventName.Type)
	assert.Equal(t, retEventName.ID, duplicateEventName.ID)

	// Test CreateOrGetUserCreatedEventName without ProjectId.
	ucEventName = &M.EventName{Name: U.RandomLowerAphaNumString(10)}
	retEventName, errCode = M.CreateOrGetUserCreatedEventName(ucEventName)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, retEventName)

	// Test CreateOrGetUserCreatedEventName without name.
	ucEventName = &M.EventName{Name: "", ProjectId: project.ID}
	retEventName, errCode = M.CreateOrGetUserCreatedEventName(ucEventName)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, retEventName)

	// Test CreateOrGetUserCreatedEventName with disallowed name.
	ucEventName = &M.EventName{Name: "$name", ProjectId: project.ID}
	retEventName, errCode = M.CreateOrGetUserCreatedEventName(ucEventName)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, retEventName)
}

func TestDBGetEventNames(t *testing.T) {
	// Initialize a project for the event.
	randomProjectName := U.RandomLowerAphaNumString(15)
	project, errCode := M.CreateProjectWithDependencies(&M.Project{Name: randomProjectName})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, project)
	projectId := project.ID

	// bad input
	events, errCode := M.GetEventNames(0)
	assert.Equal(t, http.StatusBadRequest, errCode)

	// get events should return not found, no events have been created
	events, errCode = M.GetEventNames(projectId)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, events)

	// create events
	eventName1, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{Name: "test_event", ProjectId: projectId})
	assert.Equal(t, http.StatusCreated, errCode)
	eventName2, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{Name: "test_event_1", ProjectId: projectId})
	assert.Equal(t, http.StatusCreated, errCode)

	createdEventsNames := []string{eventName1.Name, eventName2.Name}
	sort.Strings(createdEventsNames)

	// should return events
	events, errCode = M.GetEventNames(projectId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Len(t, events, 2)

	resultEventNames := []string{events[0].Name, events[1].Name}
	sort.Strings(resultEventNames)
	assert.Equal(t, createdEventsNames, resultEventNames)
}

func TestDBIsFilterMatch(t *testing.T) {
	assert.True(t, M.IsFilterMatch(U.TokenizeURI("/u1/u2"), U.TokenizeURI("/u1/u2")))
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/u1/u2"), U.TokenizeURI("/u1")))
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/u1/u2"), U.TokenizeURI("")))

	assert.True(t, M.IsFilterMatch(U.TokenizeURI("/u1/:v1"), U.TokenizeURI("/u1/a1")))
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/u3/:v1"), U.TokenizeURI("/u1/1")))
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/u1/:v1"), U.TokenizeURI("/u1")))

	assert.True(t, M.IsFilterMatch(U.TokenizeURI("/:v1"), U.TokenizeURI("/a1")))
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/:v1"), U.TokenizeURI("/a1/a3")))
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/:v1"), U.TokenizeURI("/")))
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/:v1"), U.TokenizeURI("")))

	assert.True(t, M.IsFilterMatch(U.TokenizeURI("/:v1/u1"), U.TokenizeURI("/a1/u1")))
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/:v1/u1"), U.TokenizeURI("/a1")))
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/:v1/u1"), U.TokenizeURI("/a1/a2/u1")))

	assert.True(t, M.IsFilterMatch(U.TokenizeURI("/u1/:v1/u2"), U.TokenizeURI("/u1/a2/u2")))
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/u1/:v1/u2"), U.TokenizeURI("/u1/a2")))
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/u1/:v1/u2"), U.TokenizeURI("/a2/u2")))

	// Empty filter.
	assert.False(t, M.IsFilterMatch(U.TokenizeURI(""), U.TokenizeURI("/u1")))

	// Root as filter.
	assert.False(t, M.IsFilterMatch(U.TokenizeURI("/"), U.TokenizeURI("/u1")))
	assert.True(t, M.IsFilterMatch(U.TokenizeURI("/"), U.TokenizeURI("/")))
}

func setupProjectAndFilters(t *testing.T, filters map[string]string) *M.Project {
	project, _ := SetupProjectReturnDAO()
	assert.NotNil(t, project)

	for name, fexpr := range filters {
		filterEventName1, errCode := M.CreateOrGetFilterEventName(&M.EventName{ProjectId: project.ID, Name: name, FilterExpr: fexpr})
		assert.NotNil(t, filterEventName1)
		assert.Equal(t, http.StatusCreated, errCode)
	}

	return project
}

func TestDBFilterEventNameByEventURL(t *testing.T) {
	filters := map[string]string{"a_u1_u2": "a.com/u1/u2", "u3_v1": "a.com/u3/:v1", "b_u1_u2": "b.com/u1/u2"}
	project := setupProjectAndFilters(t, filters)

	// Match filter - exact.
	men, errCode := M.FilterEventNameByEventURL(project.ID, "a.com/u1/u2")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men)
	assert.Equal(t, filters["a_u1_u2"], men.FilterExpr)

	// Match filter - prefix.
	men1, errCode := M.FilterEventNameByEventURL(project.ID, "a.com/u1/u2/u3/u4/u5")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men1)
	assert.Equal(t, filters["a_u1_u2"], men1.FilterExpr)

	// Match filter with property - exact.
	men2, errCode := M.FilterEventNameByEventURL(project.ID, "a.com/u3/1")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men2)
	assert.Equal(t, filters["u3_v1"], men2.FilterExpr)

	// Match filter with property - prefix.
	men3, errCode := M.FilterEventNameByEventURL(project.ID, "a.com/u3/1/u1/u2")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men3)
	assert.Equal(t, filters["u3_v1"], men3.FilterExpr)

	// Match by domain scope.
	men4, errCode := M.FilterEventNameByEventURL(project.ID, "b.com/u1/u2")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men4)
	assert.Equal(t, filters["b_u1_u2"], men4.FilterExpr)

	// Test priority with similar prefix.
	filters1 := map[string]string{"u1_u2": "a.com/u1/u2", "u1_u2_u3": "a.com/u1/u2/u3"}
	project1 := setupProjectAndFilters(t, filters1)

	men11, errCode := M.FilterEventNameByEventURL(project1.ID, "a.com/u1/u2")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men11)
	assert.Equal(t, filters1["u1_u2"], men11.FilterExpr)

	men12, errCode := M.FilterEventNameByEventURL(project1.ID, "a.com/u1/u2/u3")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men12)
	assert.Equal(t, filters1["u1_u2_u3"], men12.FilterExpr)

	men13, errCode := M.FilterEventNameByEventURL(project1.ID, "a.com/u1/u2/u3/u4")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men13)
	assert.Equal(t, filters1["u1_u2_u3"], men13.FilterExpr)

	men14, errCode := M.FilterEventNameByEventURL(project1.ID, "a.com/u1/u2/u4")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men14)
	assert.Equal(t, filters1["u1_u2"], men14.FilterExpr)

	men15, errCode := M.FilterEventNameByEventURL(project1.ID, "a.com/u3/u1/u2")
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, men15)

	// Test definity score priority based matching.
	filters2 := map[string]string{"u1_v1": "a.com/u1/:v1", "u1_u2": "a.com/u1/u2", "u1_v1_u2": "a.com/u1/:v1/u2", "u1_u2_v1": "a.com/u1/u2/:v1", "u1_u2_u3": "a.com/u1/u2/u3", "u1_v1_v2": "a.com/u1/:v1:/:v2"}
	project2 := setupProjectAndFilters(t, filters2)

	men20, errCode := M.FilterEventNameByEventURL(project2.ID, "a.com/u1/u2")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men20)
	assert.Equal(t, filters2["u1_u2"], men20.FilterExpr)

	men21, errCode := M.FilterEventNameByEventURL(project2.ID, "a.com/u1/i1")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men21)
	assert.Equal(t, filters2["u1_v1"], men21.FilterExpr)

	men22, errCode := M.FilterEventNameByEventURL(project2.ID, "a.com/u1/i1/i2")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men22)
	assert.Equal(t, filters2["u1_v1_v2"], men22.FilterExpr)

	men23, errCode := M.FilterEventNameByEventURL(project2.ID, "a.com/u1/i1/u2")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men23)
	assert.Equal(t, filters2["u1_v1_u2"], men23.FilterExpr)

	men24, errCode := M.FilterEventNameByEventURL(project2.ID, "a.com/u1/u2/u3")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men24)
	assert.Equal(t, filters2["u1_u2_u3"], men24.FilterExpr)

	men25, errCode := M.FilterEventNameByEventURL(project2.ID, "a.com/u1/u2/i1")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, men25)
	assert.Equal(t, filters2["u1_u2_v1"], men25.FilterExpr)
}

func TestDBFillEventPropertiesByFilterExpr(t *testing.T) {
	props := U.PropertiesMap{}
	M.FillEventPropertiesByFilterExpr(&props, "a.com/:v1", "a.com/i1")
	assert.NotNil(t, props["v1"])
	assert.Equal(t, "i1", props["v1"])

	props1 := U.PropertiesMap{}
	M.FillEventPropertiesByFilterExpr(&props1, "a.com/u1/:v1", "a.com/u1/i1")
	assert.NotNil(t, props1["v1"])
	assert.Equal(t, "i1", props1["v1"])

	// multiple values
	props2 := U.PropertiesMap{}
	M.FillEventPropertiesByFilterExpr(&props2, "a.com/u1/:v1/u2/:v2", "a.com/u1/i1/u2/i2")
	assert.NotNil(t, props2["v1"])
	assert.NotNil(t, props2["v2"])
	assert.Equal(t, "i1", props2["v1"])
	assert.Equal(t, "i2", props2["v2"])

	// continuous multiple values
	props3 := U.PropertiesMap{}
	M.FillEventPropertiesByFilterExpr(&props3, "a.com/u1/:v1/:v2", "a.com/u1/i1/i2")
	assert.NotNil(t, props3["v1"])
	assert.NotNil(t, props3["v2"])
	assert.Equal(t, "i1", props3["v1"])
	assert.Equal(t, "i2", props3["v2"])

	props4 := U.PropertiesMap{}
	M.FillEventPropertiesByFilterExpr(&props4, "a.com/u1/:v1/u2", "https://a.com/u1/i1/u2")
	assert.NotNil(t, props4["v1"])
	assert.Equal(t, "i1", props4["v1"])
}

func TestDBCreateOrGetFilterEventName(t *testing.T) {
	randomProjectName := U.RandomLowerAphaNumString(15)
	project, errCode := M.CreateProjectWithDependencies(&M.Project{Name: randomProjectName})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, project)

	expr := "a.com/u1/u2/u3"
	name := "login"
	eventName, errCode := M.CreateOrGetFilterEventName(&M.EventName{
		ProjectId:  project.ID,
		FilterExpr: expr,
		Name:       name,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, eventName)
	assert.NotZero(t, eventName.ID)
	assert.Equal(t, name, eventName.Name)
	assert.Equal(t, expr, eventName.FilterExpr)
	assert.Equal(t, M.TYPE_FILTER_EVENT_NAME, eventName.Type)

	// only domain as expr.
	expr = "b.com"
	name = "root"
	eventName, errCode = M.CreateOrGetFilterEventName(&M.EventName{
		ProjectId:  project.ID,
		Name:       name,
		FilterExpr: expr,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, eventName)
	assert.NotZero(t, eventName.ID)
	assert.Equal(t, name, eventName.Name)
	assert.Equal(t, "b.com/", eventName.FilterExpr) // only domain. root as expr.
	assert.Equal(t, M.TYPE_FILTER_EVENT_NAME, eventName.Type)

	// Test property and sanitization of expr.
	expr = "https://a.com/u1/:v1?q=10"
	name = "login2"
	eventName, errCode = M.CreateOrGetFilterEventName(&M.EventName{
		ProjectId:  project.ID,
		Name:       name,
		FilterExpr: expr,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, eventName)
	assert.NotZero(t, eventName.ID)
	assert.Equal(t, name, eventName.Name)
	assert.Equal(t, "a.com/u1/:v1", eventName.FilterExpr) // sanitized expr.
	assert.Equal(t, M.TYPE_FILTER_EVENT_NAME, eventName.Type)

	expr = ""
	name = "login2"
	eventName, errCode = M.CreateOrGetFilterEventName(&M.EventName{
		ProjectId:  project.ID,
		Name:       name,
		FilterExpr: expr,
	})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, eventName)

	expr = "a.com/u1/u2"
	name = ""
	eventName, errCode = M.CreateOrGetFilterEventName(&M.EventName{
		ProjectId:  project.ID,
		Name:       name,
		FilterExpr: expr,
	})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, eventName)

	// Test expr without domain.
	expr = "/u1/u2"
	name = "u1_u2"
	eventName, errCode = M.CreateOrGetFilterEventName(&M.EventName{
		ProjectId:  project.ID,
		Name:       name,
		FilterExpr: expr,
	})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, eventName)
}

func TestDBGetFilterEventNames(t *testing.T) {
	randomProjectName := U.RandomLowerAphaNumString(15)
	project, errCode := M.CreateProjectWithDependencies(&M.Project{Name: randomProjectName})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, project)

	// No filter event_names available.
	eventNames, errCode := M.GetFilterEventNames(project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.NotNil(t, eventNames)
	assert.Zero(t, len(eventNames))

	// Create filter_event_name.
	expr := "a.com/u1/u2/u3"
	name := "login"
	createdEN, errCode := M.CreateOrGetFilterEventName(&M.EventName{
		ProjectId:  project.ID,
		FilterExpr: expr,
		Name:       name,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, createdEN)

	eventNames, errCode = M.GetFilterEventNames(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, eventNames)
	assert.Equal(t, 1, len(eventNames))
	assert.Equal(t, createdEN.ID, eventNames[0].ID)

	// Should not return deleted.
	_, errCode = M.DeleteFilterEventName(project.ID, createdEN.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	eventNames, errCode = M.GetFilterEventNames(project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
}

func TestDBUpdateFilterEventName(t *testing.T) {
	randomProjectName := U.RandomLowerAphaNumString(15)
	project, errCode := M.CreateProjectWithDependencies(&M.Project{Name: randomProjectName})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, project)

	// Invalid event_name id.
	eventName, errCode := M.UpdateFilterEventName(project.ID, 9999, &M.EventName{Name: U.RandomLowerAphaNumString(5)})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, eventName)

	expr := "a.com/u1/u2/u3"
	name := "login"
	createdEN, errCode := M.CreateOrGetFilterEventName(&M.EventName{
		ProjectId:  project.ID,
		FilterExpr: expr,
		Name:       name,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, createdEN)

	// Try updating expr.
	newExpr := "/new/expr"
	eventName, errCode = M.UpdateFilterEventName(project.ID, createdEN.ID, &M.EventName{Name: "login", FilterExpr: newExpr})
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.NotNil(t, eventName)
	assert.NotEqual(t, eventName.FilterExpr, newExpr) // not updated.

	// Happy path.
	newName := U.RandomLowerAphaNumString(5)
	eventName, errCode = M.UpdateFilterEventName(project.ID, createdEN.ID, &M.EventName{Name: newName})
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.NotNil(t, eventName)
	assert.Equal(t, newName, eventName.Name)

	// Invalid project_id.
	eventName, errCode = M.UpdateFilterEventName(999999, createdEN.ID, &M.EventName{Name: U.RandomLowerAphaNumString(5)})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, eventName)
}

func TestDBDeleteFilterEventName(t *testing.T) {
	randomProjectName := U.RandomLowerAphaNumString(15)
	project, errCode := M.CreateProjectWithDependencies(&M.Project{Name: randomProjectName})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, project)

	// Invalid event_name id.
	eventName, errCode := M.DeleteFilterEventName(project.ID, 9999)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, eventName)

	expr := "a.com/u1/u2/u3"
	name := "login"
	createdEN, errCode := M.CreateOrGetFilterEventName(&M.EventName{
		ProjectId:  project.ID,
		FilterExpr: expr,
		Name:       name,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, createdEN)

	eventName, errCode = M.DeleteFilterEventName(project.ID, createdEN.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.NotNil(t, eventName)
}
