package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	IntSalesforce "factors/integration/salesforce"
	M "factors/model"
	"factors/task/event_user_cache"
	TaskSession "factors/task/session"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndGetEventName(t *testing.T) {
	// Initialize a project for the event.
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

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

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	assert.NotNil(t, project)
	projectId := project.ID

	// bad input
	events, errCode := M.GetEventNames(0)
	assert.Equal(t, http.StatusBadRequest, errCode)

	// get events should return not found, no events have been created
	events, errCode = M.GetEventNames(projectId)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Empty(t, events)

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

	assert.True(t, M.IsFilterMatch(U.TokenizeURI("/u1/:v1/u2/:v2"), U.TokenizeURI("/u1/l1/u2/l2")))

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
	filters := map[string]string{"a_u1_u2": "a.com/u1/u2", "u3_v1": "a.com/u3/:v1", "b_u1_u2": "b.com/u1/u2", "only_root": "a.com/"}
	project := setupProjectAndFilters(t, filters)

	// domain only event url should match with root "/" expression.
	onlyDomainEventURL, errCode := M.FilterEventNameByEventURL(project.ID, "a.com")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, onlyDomainEventURL)
	assert.Equal(t, filters["only_root"], onlyDomainEventURL.FilterExpr)

	// Match filter - exact and additional / at the end.
	men, errCode := M.FilterEventNameByEventURL(project.ID, "a.com/u1/u2/")
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

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
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

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
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
	errCode = M.DeleteFilterEventName(project.ID, createdEN.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	eventNames, errCode = M.GetFilterEventNames(project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
}

func TestDBUpdateFilterEventName(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
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
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// Invalid event_name id.
	errCode := M.DeleteFilterEventName(project.ID, 9999)
	assert.Equal(t, http.StatusBadRequest, errCode)

	expr := "a.com/u1/u2/u3"
	name := "login"
	createdEN, errCode := M.CreateOrGetFilterEventName(&M.EventName{
		ProjectId:  project.ID,
		FilterExpr: expr,
		Name:       name,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, createdEN)

	errCode = M.DeleteFilterEventName(project.ID, createdEN.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
}

func TestDBGetEventNamesOrderedByOccurrenceWithLimit(t *testing.T) {
	r := gin.Default()
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)

	user, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.NotNil(t, user)
	assert.Equal(t, http.StatusCreated, errCode)
	rEventName := "event1"
	w := ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "timestamp": %d, "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, user.ID, rEventName, timestamp)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)
	rEventName = "event2"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "timestamp": %d, "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, user.ID, rEventName, timestamp+1)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	assert.Equal(t, http.StatusOK, w.Code)

	rEventName = "event3"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "timestamp": %d, "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, user.ID, rEventName, timestamp+2)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	assert.Equal(t, http.StatusOK, w.Code)

	_, err = TaskSession.AddSession([]uint64{project.ID}, timestamp-60, 0, 0, 1)
	assert.Nil(t, err)

	eventsLimit, propertyLimit, valueLimit, rollBackWindow := 1000, 10000, 10000, 1
	event_user_cache.DoRollUpAndCleanUp(&eventsLimit, &propertyLimit, &valueLimit, &rollBackWindow)
	// with limit.
	getEventNames1, err := M.GetEventNamesOrderedByOccurenceAndRecency(project.ID, 10, 30)
	assert.Equal(t, nil, err)
	assert.Len(t, getEventNames1[U.MostRecent], 4)

	rEventName = "event2"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, user.ID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	assert.Equal(t, http.StatusOK, w.Code)

	event_user_cache.DoRollUpAndCleanUp(&eventsLimit, &propertyLimit, &valueLimit, &rollBackWindow)
	getEventNames2, err := M.GetEventNamesOrderedByOccurenceAndRecency(project.ID, 2, 30)
	assert.Equal(t, nil, err)
	assert.Len(t, getEventNames2[U.MostRecent], 2)
	assert.Equal(t, "$session", getEventNames2[U.MostRecent][0])
	assert.Equal(t, "event2", getEventNames2[U.MostRecent][1])
}

func sendCreateSmartEventFilterReq(r *gin.Engine, projectId uint64, agent *M.Agent, enPayload *map[string]interface{}) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
		return nil
	}

	rb := U.NewRequestBuilder(http.MethodPost, fmt.Sprintf("/projects/%d/v1/smart_event?type=%s", projectId, "crm")).
		WithPostParams(enPayload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating request")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetSmartEventFilterReq(r *gin.Engine, projectId uint64, agent *M.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
		return nil
	}

	rb := U.NewRequestBuilder(http.MethodGet, fmt.Sprintf("/projects/%d/v1/smart_event", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating request")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendUpdateSmartEventFilterReq(r *gin.Engine, projectID uint64, agent *M.Agent, enPayload *map[string]interface{}, filterID uint64) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
		return nil
	}

	rb := U.NewRequestBuilder(http.MethodPut, fmt.Sprintf("/projects/%d/v1/smart_event?type=%s&filter_id=%d", projectID, "crm", filterID)).
		WithPostParams(enPayload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating request")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestSmartCRMFilterCreation(t *testing.T) {

	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// string comparision
	stringComp := &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_email",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "test1@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "test@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	requestPayload := make(map[string]interface{})
	requestPayload["name"] = "smartEventString"
	requestPayload["expr"] = stringComp

	w := sendCreateSmartEventFilterReq(r, project.ID, agent, &requestPayload)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)

	var responsePayload H.APISmartEventFilterResponePayload
	err = json.Unmarshal(jsonResponse, &responsePayload)
	assert.Nil(t, err)
	stringCompEventNameId := responsePayload.EventNameID
	assert.NotEqual(t, 0, stringCompEventNameId)

	// integer comparision
	intComp := &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_count",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "5",
						Operator:      M.COMPARE_GREATER_THAN,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "4",
						Operator:      M.COMPARE_LESS_THAN,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	requestPayload = make(map[string]interface{})
	requestPayload["name"] = "smartEventInt"
	requestPayload["expr"] = intComp

	w = sendCreateSmartEventFilterReq(r, project.ID, agent, &requestPayload)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = sendGetSmartEventFilterReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	var smartCRMEvents []H.APISmartEventFilterResponePayload
	err = json.Unmarshal(jsonResponse, &smartCRMEvents)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(smartCRMEvents))

	// string compare
	currentProperties := make(map[string]interface{})
	prevProperties := make(map[string]interface{})
	currentProperties["$salesforce_contact_email"] = "test1@gmail.com"
	prevProperties["$salesforce_contact_email"] = "test@gmail.com"

	stringFilterIndex := 0
	intFilterIndex := 1
	if smartCRMEvents[1].EventName == "smartEventString" {
		stringFilterIndex = 1
		intFilterIndex = 0
	}

	smartEvent, rPrevProperties, ok := IntSalesforce.GetSalesforceSmartEventPayload(project.ID, smartCRMEvents[stringFilterIndex].EventName, "", "", 0, &currentProperties, &prevProperties, &(smartCRMEvents[stringFilterIndex].FilterExpr))
	assert.Equal(t, true, ok)
	assert.Equal(t, prevProperties, *rPrevProperties)
	assert.NotNil(t, smartEvent)
	assert.Equal(t, "smartEventString", smartEvent.Name)
	assert.Contains(t, smartEvent.Properties, "$prev_salesforce_contact_email", "$curr_salesforce_contact_email")

	// individual properties test
	state := M.CRMFilterEvaluator(project.ID, &currentProperties, nil, &(smartCRMEvents[stringFilterIndex].FilterExpr), M.CompareStateCurr)
	assert.Equal(t, true, state)
	state = M.CRMFilterEvaluator(project.ID, nil, &prevProperties, &(smartCRMEvents[stringFilterIndex].FilterExpr), M.CompareStatePrev)
	assert.Equal(t, true, state)

	// int compare
	currentProperties["$salesforce_contact_count"] = 6
	prevProperties["$salesforce_contact_count"] = 3
	smartEvent, rPrevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, smartCRMEvents[intFilterIndex].EventName, "", "", 0, &currentProperties, &prevProperties, &(smartCRMEvents[intFilterIndex].FilterExpr))
	assert.Equal(t, true, ok)
	assert.Equal(t, prevProperties, *rPrevProperties)
	assert.Contains(t, smartEvent.Properties, "$prev_salesforce_contact_count", "$curr_salesforce_contact_count")

	// overwrite filter exp
	intComp = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_count",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "5",
						Operator:      M.COMPARE_GREATER_THAN,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "4",
						Operator:      M.COMPARE_GREATER_THAN,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	requestPayload = make(map[string]interface{})
	requestPayload["name"] = "smartEventInt"
	requestPayload["expr"] = intComp

	w = sendUpdateSmartEventFilterReq(r, project.ID, agent, &requestPayload, smartCRMEvents[intFilterIndex].EventNameID)
	assert.Equal(t, http.StatusAccepted, w.Code)

	w = sendGetSmartEventFilterReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &smartCRMEvents)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(smartCRMEvents))

	if smartCRMEvents[0].EventName == "smartEventInt" {
		intFilterIndex = 0
	} else {
		intFilterIndex = 1
	}

	assert.Equal(t, intComp, &smartCRMEvents[intFilterIndex].FilterExpr)
}

func TestSmartCRMFilterStringCompare(t *testing.T) {

	/* (current email == test1@gmail.com and prev email == test@gmail.com )
	AND (current company == example2 AND  previous company == example) */
	filter := &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_email",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "test1@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "test@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
			{
				Name: "$salesforce_contact_company",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "example2",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "example1",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	currentProperties := make(map[string]interface{})
	prevProperties := make(map[string]interface{})
	currentProperties["$salesforce_contact_email"] = "test1@gmail.com"
	prevProperties["$salesforce_contact_email"] = "test@gmail.com"
	currentProperties["$salesforce_contact_company"] = "example2"
	prevProperties["$salesforce_contact_company"] = "example1"
	_, rPrevProperties, ok := IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, true, ok)
	assert.Equal(t, prevProperties, *rPrevProperties)

	/* (current email == test1@gmail.com OR prev email == test@gmail.com )
	AND (current company == example2 AND  previous company == example) */
	filter = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_email",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "test1@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "test@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_OR,
			},
			{
				Name: "$salesforce_contact_company",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "example2",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "example1",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	currentProperties = make(map[string]interface{})
	prevProperties = make(map[string]interface{})
	currentProperties["$salesforce_contact_email"] = "test1@gmail.com"
	prevProperties["$salesforce_contact_email"] = "fail@gmail.com" // failed value
	currentProperties["$salesforce_contact_company"] = "example2"
	prevProperties["$salesforce_contact_company"] = "example1"
	_, rPrevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, true, ok)
	assert.Equal(t, prevProperties, *rPrevProperties)

	// individual test
	// individual properties test
	state := M.CRMFilterEvaluator(1, &currentProperties, nil, filter, M.CompareStateCurr)
	assert.Equal(t, true, state)
	state = M.CRMFilterEvaluator(1, nil, &prevProperties, filter, M.CompareStatePrev)
	assert.Equal(t, false, state)

	/* Property OR operation
	(current email == test1@gmail.com AND prev email == test@gmail.com )
	OR (current company == example2 AND  previous company == example) */
	filter = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_email",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "test1@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "test@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
			{
				Name: "$salesforce_contact_company",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "example2",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "example1",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_OR,
		TimestampReferenceField: "time",
	}

	currentProperties = make(map[string]interface{})
	prevProperties = make(map[string]interface{})
	currentProperties["$salesforce_contact_email"] = "test1@gmail.com"
	prevProperties["$salesforce_contact_email"] = "fail@gmail.com" // failed value
	currentProperties["$salesforce_contact_company"] = "example2"
	prevProperties["$salesforce_contact_company"] = "example1"
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, true, ok)

	// individual test
	// individual properties test
	state = M.CRMFilterEvaluator(1, &currentProperties, nil, filter, M.CompareStateCurr)
	assert.Equal(t, true, state)
	state = M.CRMFilterEvaluator(1, nil, &prevProperties, filter, M.CompareStatePrev)
	assert.Equal(t, true, state)

	/*
		FAIL TESTS
	*/

	/* (current email == test1@gmail.com and prev email == test@gmail.com )
	AND (current company == example2 AND  previous company == example) */
	filter = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_email",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "test1@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "test@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
			{
				Name: "$salesforce_contact_company",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "example2",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "example1",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	currentProperties = make(map[string]interface{})
	prevProperties = make(map[string]interface{})
	currentProperties["$salesforce_contact_email"] = "test1@gmail.com"
	prevProperties["$salesforce_contact_email"] = "test@gmail.com"
	currentProperties["$salesforce_contact_company"] = "example1" // failed value
	prevProperties["$salesforce_contact_company"] = "example1"
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, false, ok)

	// individual test
	// individual properties test
	state = M.CRMFilterEvaluator(1, &currentProperties, nil, filter, M.CompareStateCurr)
	assert.Equal(t, false, state)
	state = M.CRMFilterEvaluator(1, nil, &prevProperties, filter, M.CompareStatePrev)
	assert.Equal(t, true, state)

	/* (current email == test1@gmail.com and prev email == test@gmail.com )
	OR (current company == example2 AND  previous company == example) */
	filter = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_email",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "test1@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "test@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
			{
				Name: "$salesforce_contact_company",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "example2",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "example1",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	currentProperties = make(map[string]interface{})
	prevProperties = make(map[string]interface{})
	currentProperties["$salesforce_contact_email"] = "failed@gmail.com" //failed value
	prevProperties["$salesforce_contact_email"] = "failed2@gmail.com"   //failed value
	currentProperties["$salesforce_contact_company"] = "failed"         // failed value
	prevProperties["$salesforce_contact_company"] = "failed"            //failed value
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, false, ok)

	// individual test
	// individual properties test
	state = M.CRMFilterEvaluator(1, &currentProperties, nil, filter, M.CompareStateCurr)
	assert.Equal(t, false, state)
	state = M.CRMFilterEvaluator(1, nil, &prevProperties, filter, M.CompareStatePrev)
	assert.Equal(t, false, state)

	/* (current email == test1@gmail.com OR prev email == test@gmail.com )
	OR (current company == example2 OR previous company == example) */
	filter = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_email",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "test1@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "test@gmail.com",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
			{
				Name: "$salesforce_contact_company",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "example2",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "example1",
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	currentProperties = make(map[string]interface{})
	prevProperties = make(map[string]interface{})
	currentProperties["$salesforce_contact_email"] = "failed@gmail.com" //failed value
	prevProperties["$salesforce_contact_email"] = "failed2@gmail.com"   //failed value
	currentProperties["$salesforce_contact_company"] = "failed"         // failed value
	prevProperties["$salesforce_contact_company"] = "failed"            //failed value
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, false, ok)

	// individual test
	// individual properties test
	state = M.CRMFilterEvaluator(1, &currentProperties, nil, filter, M.CompareStateCurr)
	assert.Equal(t, false, state)
	state = M.CRMFilterEvaluator(1, nil, &prevProperties, filter, M.CompareStatePrev)
	assert.Equal(t, false, state)

}

func TestSmartCRMFilterContains(t *testing.T) {
	/* (current $description  contains "greetings" and prev $$description contains "greetings" ) */
	filter := &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_description",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "greetings",
						Operator:      M.COMPARE_CONTAINS,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "greetings",
						Operator:      M.COMPARE_CONTAINS,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	currentProperties := make(map[string]interface{})
	prevProperties := make(map[string]interface{})
	currentProperties["$salesforce_contact_description"] = "greetings from example.com"
	prevProperties["$salesforce_contact_description"] = "will be providing greetings"
	_, _, ok := IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, true, ok)

	/* (current $description  not contains "greetings" and prev $$description not contains "greetings" ) */
	filter = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_description",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "greetings",
						Operator:      M.COMPARE_NOT_CONTAINS,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "greetings",
						Operator:      M.COMPARE_NOT_CONTAINS,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, false, ok)
}

func TestSmartCRMFilterInteger(t *testing.T) {

	/* (current page_spent_time  > 5 and prev page_spent_time < 3 ) */
	filter := &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_page_spent_time",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         5,
						Operator:      M.COMPARE_GREATER_THAN,
					},
					{
						PropertyState: M.PreviousState,
						Value:         3,
						Operator:      M.COMPARE_LESS_THAN,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	currentProperties := make(map[string]interface{})
	prevProperties := make(map[string]interface{})
	currentProperties["$salesforce_contact_page_spent_time"] = 7
	prevProperties["$salesforce_contact_page_spent_time"] = 2
	_, _, ok := IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, true, ok)

	// individual test
	// individual properties test
	state := M.CRMFilterEvaluator(1, &currentProperties, nil, filter, M.CompareStateCurr)
	assert.Equal(t, true, state)
	state = M.CRMFilterEvaluator(1, nil, &prevProperties, filter, M.CompareStatePrev)
	assert.Equal(t, true, state)

	// Fail test
	currentProperties["$salesforce_contact_page_spent_time"] = 3
	prevProperties["$salesforce_contact_page_spent_time"] = 2
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, false, ok)

	/* (current page_spent_time  == 5 and prev page_spent_time == 3 )
	OR (current page_count == 10 AND  previous page_count == 7) */
	filter = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_page_spent_time",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         5,
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         3,
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
			{
				Name: "$salesforce_contact_page_spent_count",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         10,
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         7,
						Operator:      M.COMPARE_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	currentProperties = make(map[string]interface{})
	prevProperties = make(map[string]interface{})
	currentProperties["$salesforce_contact_page_spent_time"] = 5
	prevProperties["$salesforce_contact_page_spent_time"] = 3
	currentProperties["$salesforce_contact_page_spent_count"] = 10
	prevProperties["$salesforce_contact_page_spent_count"] = 7
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, true, ok)

	// individual test
	// individual properties test
	state = M.CRMFilterEvaluator(1, &currentProperties, nil, filter, M.CompareStateCurr)
	assert.Equal(t, true, state)
	state = M.CRMFilterEvaluator(1, nil, &prevProperties, filter, M.CompareStatePrev)
	assert.Equal(t, true, state)

	/* (current page_spent_time  == 5 and prev page_spent_time == 3 )
	OR (current page_count == 10 AND  previous page_count == 7) */
	filter = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_page_spent_time",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         5,
						Operator:      M.COMPARE_GREATER_THAN,
					},
					{
						PropertyState: M.PreviousState,
						Value:         3,
						Operator:      M.COMPARE_LESS_THAN,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
			{
				Name: "$salesforce_contact_page_spent_count",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         10,
						Operator:      M.COMPARE_LESS_THAN,
					},
					{
						PropertyState: M.PreviousState,
						Value:         7,
						Operator:      M.COMPARE_GREATER_THAN,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	currentProperties = make(map[string]interface{})
	prevProperties = make(map[string]interface{})
	currentProperties["$salesforce_contact_page_spent_time"] = 7
	prevProperties["$salesforce_contact_page_spent_time"] = 2
	currentProperties["$salesforce_contact_page_spent_count"] = 6
	prevProperties["$salesforce_contact_page_spent_count"] = 8
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, true, ok)

	// individual test
	// individual properties test
	state = M.CRMFilterEvaluator(1, &currentProperties, nil, filter, M.CompareStateCurr)
	assert.Equal(t, true, state)
	state = M.CRMFilterEvaluator(1, nil, &prevProperties, filter, M.CompareStatePrev)
	assert.Equal(t, true, state)
}

func TestSmartCRMFilterAnyChange(t *testing.T) {

	/* any change in $page_spent_time */
	filter := &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeAny,
		Filters: []M.PropertyFilter{
			{
				Name:  "$salesforce_contact_page_spent_time",
				Rules: []M.CRMFilterRule{},
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	currentProperties := make(map[string]interface{})
	prevProperties := make(map[string]interface{})
	currentProperties["$salesforce_contact_page_spent_time"] = 7
	prevProperties["$salesforce_contact_page_spent_time"] = 2

	smartEvent, rPrevProperties, ok := IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, true, ok)
	assert.Equal(t, prevProperties, *rPrevProperties)
	assert.Contains(t, smartEvent.Properties, "$curr_salesforce_contact_page_spent_time")
	assert.Contains(t, smartEvent.Properties, "$prev_salesforce_contact_page_spent_time")

	ok = M.CRMFilterEvaluator(1, &currentProperties, nil, filter, M.CompareStateCurr)
	assert.Equal(t, true, ok)
	// same value
	prevProperties["$salesforce_contact_page_spent_time"] = 7
	_, rPrevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, false, ok)
	assert.Equal(t, prevProperties, *rPrevProperties)

	/* any change in $page_spent_time OR $count */
	filter = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeAny,
		Filters: []M.PropertyFilter{
			{
				Name:  "$salesforce_contact_page_spent_time",
				Rules: []M.CRMFilterRule{},
			},
			{
				Name:  "$salesforce_contact_count",
				Rules: []M.CRMFilterRule{},
			},
		},
		LogicalOp:               M.LOGICAL_OP_OR,
		TimestampReferenceField: "time",
	}

	currentProperties = make(map[string]interface{})
	prevProperties = make(map[string]interface{})
	currentProperties["$salesforce_contact_page_spent_time"] = 2
	prevProperties["$salesforce_contact_page_spent_time"] = 10
	currentProperties["$salesforce_contact_count"] = 2
	prevProperties["$salesforce_contact_count"] = 2
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, true, ok)

	// fail on no change
	currentProperties["$salesforce_contact_page_spent_time"] = 2
	prevProperties["$salesforce_contact_page_spent_time"] = 2
	currentProperties["$salesforce_contact_count"] = 2
	prevProperties["$salesforce_contact_count"] = 2
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, false, ok)

	/*
		prev $page_spent_time equals anything and curr $page_spent_time = 10
	*/
	filter = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeAny,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_page_spent_time",
				Rules: []M.CRMFilterRule{
					{
						Operator:      M.COMPARE_EQUAL,
						Value:         M.PROPERTY_VALUE_ANY,
						PropertyState: M.PreviousState,
					},
					{
						Operator:      M.COMPARE_EQUAL,
						Value:         10,
						PropertyState: M.CurrentState,
					},
				},
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	currentProperties["$salesforce_contact_page_spent_time"] = 10
	prevProperties["$salesforce_contact_page_spent_time"] = 2
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, true, ok)

	// fail on same value
	currentProperties["$salesforce_contact_page_spent_time"] = 10
	prevProperties["$salesforce_contact_page_spent_time"] = 10
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, false, ok)

	/*
		prev $page_spent_time equals 10 and curr $page_spent_time equal anything
	*/
	filter = &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeAny,
		Filters: []M.PropertyFilter{
			{
				Name: "$salesforce_contact_page_spent_time",
				Rules: []M.CRMFilterRule{
					{
						Operator:      M.COMPARE_EQUAL,
						Value:         M.PROPERTY_VALUE_ANY,
						PropertyState: M.CurrentState,
					},
					{
						Operator:      M.COMPARE_EQUAL,
						Value:         10,
						PropertyState: M.PreviousState,
					},
				},
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	currentProperties["$salesforce_contact_page_spent_time"] = 2
	prevProperties["$salesforce_contact_page_spent_time"] = 10
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, true, ok)

	// fail on same value
	currentProperties["$salesforce_contact_page_spent_time"] = 10
	prevProperties["$salesforce_contact_page_spent_time"] = 10
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(1, "test", "", "", 0, &currentProperties, &prevProperties, filter)
	assert.Equal(t, false, ok)

}

func TestPrioritizeSmartEventNames(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	filter := &M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: M.FilterEvaluationTypeAny,
		Filters: []M.PropertyFilter{
			{
				Name:  "page_spent_time",
				Rules: []M.CRMFilterRule{},
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	//Smart event names
	smartEventNames := make([]M.EventName, 0)
	for i := 0; i < 5; i++ {
		filter.Filters[0].Name = fmt.Sprintf("property %d", i)
		eventName, status := M.CreateOrGetCRMSmartEventFilterEventName(project.ID, &M.EventName{ProjectId: project.ID, Name: fmt.Sprintf("Smart Event Name %d", i)}, filter)
		assert.Equal(t, http.StatusCreated, status)
		smartEventNames = append(smartEventNames, *eventName)
	}

	// Normal event names
	eventNames := make([]M.EventName, 0)
	for i := 0; i < 5; i++ {
		eventName, status := M.CreateOrGetEventName(&M.EventName{ProjectId: project.ID, Name: fmt.Sprintf("Event Name %d", i), Type: M.TYPE_USER_CREATED_EVENT_NAME})
		assert.Equal(t, http.StatusCreated, status)
		eventNames = append(eventNames, *eventName)
	}

	user, status := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, status)

	// creating multiple normal events
	for i := 0; i < 100; i++ {
		_, status := M.CreateEvent(&M.Event{
			EventNameId: eventNames[i%5].ID,
			ProjectId:   project.ID,
			UserId:      user.ID,
			Timestamp:   U.TimeNowUnix(),
		})
		assert.Equal(t, http.StatusCreated, status)
	}

	// creating less smart events
	for i := 0; i < 10; i++ {
		_, status := M.CreateEvent(&M.Event{
			EventNameId: smartEventNames[i%5].ID,
			ProjectId:   project.ID,
			UserId:      user.ID,
			Timestamp:   U.TimeNowUnix(),
		})
		assert.Equal(t, http.StatusCreated, status)
	}

	eventsLimit, propertyLimit, valueLimit, rollBackWindow := 1000, 10000, 10000, 1
	event_user_cache.DoRollUpAndCleanUp(&eventsLimit, &propertyLimit, &valueLimit, &rollBackWindow)

	getEventNames, err := M.GetEventNamesOrderedByOccurenceAndRecency(project.ID, 10, 30)
	assert.Equal(t, nil, err)
	responseSmartEventNames := getEventNames[U.MostRecent][:5]
	//check top 5 are smart event names
	for i := 1; i < 5; i++ {
		assert.Contains(t, responseSmartEventNames, fmt.Sprintf("Smart Event Name %d", i))
	}

}
