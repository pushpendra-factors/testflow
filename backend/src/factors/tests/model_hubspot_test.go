package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	IntHubspot "factors/integration/Hubspot"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	"factors/task/event_user_cache"
	"factors/util"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestHubspotCRMSmartEvent(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	// contactID := U.RandomLowerAphaNumString(5)
	userID1 := U.RandomLowerAphaNumString(5)
	userID2 := U.RandomLowerAphaNumString(5)
	userID3 := U.RandomLowerAphaNumString(5)
	cuid := U.RandomLowerAphaNumString(5)
	_, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID1, CustomerUserId: cuid})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID2, CustomerUserId: cuid})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID3, CustomerUserId: cuid})
	assert.Equal(t, http.StatusCreated, status)

	createdAt := time.Now().AddDate(0, 0, -11)
	updatedDate := createdAt.AddDate(0, 0, 1)

	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	jsonContact := fmt.Sprintf(jsonContactModel, 1, updatedDate.Unix(), createdAt.Unix(), updatedDate.Unix(), "lead", cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdAt.Unix(), hubspotDocument.Timestamp)
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID1)
	assert.Equal(t, http.StatusAccepted, status)

	filter := model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceHubspot,
		ObjectType:           "contact",
		Description:          "hubspot contact lifecyclestage",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "lifecyclestage",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "opportunity",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "lead",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	currentProperties := make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity"
	smartEvent, _, ok := IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", "1", hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)
	assert.Equal(t, "test", smartEvent.Name)
	assert.Equal(t, "lead", smartEvent.Properties["$prev_hubspot_contact_lifecyclestage"])
	assert.Equal(t, "opportunity", smartEvent.Properties["$curr_hubspot_contact_lifecyclestage"])

	// updated to opportunity
	updatedDate = updatedDate.AddDate(0, 0, 1)
	jsonContact = fmt.Sprintf(jsonContactModel, 1, updatedDate.Unix(), createdAt.Unix(), updatedDate.Unix(), "opportunity", "test@gmail.com", "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, updatedDate.Unix(), hubspotDocument.Timestamp)

	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID2)
	assert.Equal(t, http.StatusAccepted, status)
	// previous rule should fail
	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", "1", hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, false, ok)
	// use any change should also fail
	filter.FilterEvaluationType = model.FilterEvaluationTypeAny
	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", "1", hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, false, ok)

	// updated last synced to customer
	updatedDate = updatedDate.AddDate(0, 0, 2)
	jsonContact = fmt.Sprintf(jsonContactModel, 1, updatedDate.Unix(), createdAt.Unix(), updatedDate.Unix(), "customer", cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, updatedDate.Unix(), hubspotDocument.Timestamp)
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID1)
	assert.Equal(t, http.StatusAccepted, status)

	filter.FilterEvaluationType = model.FilterEvaluationTypeSpecific
	filter.Filters[0].Rules = []model.CRMFilterRule{
		{
			PropertyState: model.CurrentState,
			Value:         "opportunity",
			Operator:      model.COMPARE_EQUAL,
		},
		{
			PropertyState: model.PreviousState,
			Value:         "customer",
			Operator:      model.COMPARE_EQUAL,
		},
	}

	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", "1", hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)
	assert.Equal(t, "test", smartEvent.Name)
	assert.Equal(t, "customer", smartEvent.Properties["$prev_hubspot_contact_lifecyclestage"])
	assert.Equal(t, "opportunity", smartEvent.Properties["$curr_hubspot_contact_lifecyclestage"])

	// updated last synced to lead with different user_id having same customer_user_id should not have affect
	updatedDate = updatedDate.AddDate(0, 0, 3)
	jsonContact = fmt.Sprintf(jsonContactModel, 1, updatedDate.Unix(), createdAt.Unix(), updatedDate.Unix(), "lead", cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, updatedDate.Unix(), hubspotDocument.Timestamp)
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID3)
	assert.Equal(t, http.StatusAccepted, status)

	filter.Filters[0].Rules = []model.CRMFilterRule{
		{
			PropertyState: model.CurrentState,
			Value:         "opportunity",
			Operator:      model.COMPARE_EQUAL,
		},
		{
			PropertyState: model.PreviousState,
			Value:         "lead",
			Operator:      model.COMPARE_EQUAL,
		},
	}

	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", "1", hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)
	assert.Equal(t, "test", smartEvent.Name)
	assert.Equal(t, "lead", smartEvent.Properties["$prev_hubspot_contact_lifecyclestage"])
	assert.Equal(t, "opportunity", smartEvent.Properties["$curr_hubspot_contact_lifecyclestage"])

	// updated last synced to lead with different user_id having same customer_user_id should not have affect
	updatedDate = updatedDate.AddDate(0, 0, 3)
	jsonContact = fmt.Sprintf(jsonContactModel, 2, updatedDate.Unix(), createdAt.Unix(), updatedDate.Unix(), "lead", cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdAt.Unix(), hubspotDocument.Timestamp)
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID3)
	assert.Equal(t, http.StatusAccepted, status)

	// use empty records if no previous record exist
	filter.Filters[0].Rules = []model.CRMFilterRule{
		{
			PropertyState: model.CurrentState,
			Value:         U.PROPERTY_VALUE_ANY,
			Operator:      model.COMPARE_EQUAL,
		},
		{
			PropertyState: model.PreviousState,
			Value:         "lead",
			Operator:      model.COMPARE_EQUAL,
		},
	}

	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity1"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", "2", hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)

	assert.Equal(t, "test", smartEvent.Name)
	assert.Equal(t, "lead", smartEvent.Properties["$prev_hubspot_contact_lifecyclestage"])
	assert.Equal(t, "opportunity1", smartEvent.Properties["$curr_hubspot_contact_lifecyclestage"])

	// negative case for above
	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "lead"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", "2", hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, false, ok)

	//use empty records if no previous record exist
	filter.Filters[0].Rules = []model.CRMFilterRule{
		{
			PropertyState: model.CurrentState,
			Value:         U.PROPERTY_VALUE_ANY,
			Operator:      model.COMPARE_EQUAL,
		},
		{
			PropertyState: model.PreviousState,
			Value:         U.PROPERTY_VALUE_ANY,
			Operator:      model.COMPARE_NOT_EQUAL,
		},
	}

	// no previous record by document id
	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity1"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", "3", hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)
	assert.Nil(t, smartEvent.Properties["$prev_hubspot_contact_lifecyclestage"])
	assert.Equal(t, "opportunity1", smartEvent.Properties["$curr_hubspot_contact_lifecyclestage"])

	//use empty records if no previous record exist
	filter.Filters[0].Rules = []model.CRMFilterRule{
		{
			PropertyState: model.CurrentState,
			Value:         U.PROPERTY_VALUE_ANY,
			Operator:      model.COMPARE_EQUAL,
		},
		{
			PropertyState: model.PreviousState,
			Value:         U.PROPERTY_VALUE_ANY,
			Operator:      model.COMPARE_NOT_EQUAL,
		},
	}

	// if property value nil
	PrevProperties := make(map[string]interface{})
	PrevProperties["lifecyclestage"] = nil
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", "2", hubspotDocument.Type, &currentProperties, &PrevProperties, &filter)
	assert.Equal(t, true, ok)
}

func TestHubspotEventUserPropertiesState(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	intHubspot := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntHubspot: &intHubspot, IntHubspotApiKey: "1234",
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	cuID := U.RandomLowerAphaNumString(5) + "@exm.com"
	firstPropTimestamp := time.Now().Unix()
	createdUserID, status := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		JoinTimestamp:  firstPropTimestamp,
		CustomerUserId: cuID,
	})
	assert.Equal(t, http.StatusCreated, status)
	assert.NotEmpty(t, createdUserID)

	properties := &postgres.Jsonb{RawMessage: []byte(`{"name":"user1","city":"bangalore"}`)}
	_, status = store.GetStore().UpdateUserProperties(project.ID, createdUserID, properties, firstPropTimestamp)
	assert.Equal(t, http.StatusAccepted, status)

	createdDate := time.Now()

	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	jsonContact := fmt.Sprintf(jsonContactModel, 1, createdDate.Unix()*1000, createdDate.Unix()*1000, createdDate.Unix()*1000, "lead", cuID, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdDate.Unix()*1000, hubspotDocument.Timestamp)

	//enrich job, create contact created and contact updated event
	enrichStatus, _ := IntHubspot.Sync(project.ID, 3)
	projectIndex := -1
	for i := range enrichStatus {
		if enrichStatus[i].ProjectId == project.ID {
			projectIndex = i
			break
		}
	}
	assert.Equal(t, project.ID, enrichStatus[projectIndex].ProjectId)
	assert.Equal(t, "success", enrichStatus[projectIndex].Status)

	query := model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "$hubspot_contact_created",
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				Property:       "city",
				EventName:      "$hubspot_contact_created",
				EventNameIndex: 1,
			},
		},

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, status, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "city", result.Headers[0])
	assert.Equal(t, "bangalore", result.Rows[1][0])
	assert.Equal(t, float64(1), result.Rows[1][1])

	query = model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "$hubspot_contact_created",
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				Property:       "$user_id",
				EventName:      "$hubspot_contact_created",
				EventNameIndex: 1,
			},
		},

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, status, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, cuID, result.Rows[1][0])
	assert.Equal(t, float64(1), result.Rows[1][1])
}

func TestHubspotObjectPropertiesAPI(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	property1 := U.RandomLowerAphaNumString(4)
	documentID := 1
	createdAt := time.Now().AddDate(0, 0, -1).Unix() * 1000
	updatedAt := createdAt + 100
	cuid := U.RandomLowerAphaNumString(5)

	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "%s": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	value1 := "val1"
	jsonContact := fmt.Sprintf(jsonContactModel, documentID, updatedAt, createdAt, updatedAt, property1, value1, cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}
	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdAt, hubspotDocument.Timestamp)

	documents, status := store.GetStore().GetHubspotDocumentsByTypeForSync(project.ID, model.HubspotDocumentTypeContact)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 3, len(documents))

	// try reinserting the same record
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusConflict, status)

	documents, status = store.GetStore().GetHubspotDocumentsByTypeForSync(project.ID, model.HubspotDocumentTypeContact)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 3, len(documents))

	// 100 unique values
	limit := 99
	for i := 0; i < limit; i++ {
		updatedAt = updatedAt + 100
		value1 = fmt.Sprintf("%s_%d", property1, i)
		jsonContact = fmt.Sprintf(jsonContactModel, documentID, updatedAt, createdAt, updatedAt, property1, value1, cuid, "123-45")
		contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

		hubspotDocument = model.HubspotDocument{
			TypeAlias: model.HubspotDocumentTypeNameContact,
			Value:     &contactPJson,
		}
		status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, updatedAt, hubspotDocument.Timestamp)
	}

	var property1Values []interface{}
	w := sendGetCRMObjectValuesByPropertyNameReq(r, project.ID, agent, model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContact, property1)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &property1Values)
	assert.Nil(t, err)
	//should contain all values
	for i := 0; i < limit; i++ {
		assert.Contains(t, property1Values, fmt.Sprintf("%s_%d", property1, i))
	}

	// increasing count based on value1
	for i := 0; i < 5; i++ {
		for j := 0; j < i+2; j++ {
			updatedAt = updatedAt + 100
			value1 = fmt.Sprintf("%s_%d", property1, i)
			jsonContact = fmt.Sprintf(jsonContactModel, documentID, updatedAt, createdAt, updatedAt, property1, value1, cuid, "123-45")
			contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

			hubspotDocument = model.HubspotDocument{
				TypeAlias: model.HubspotDocumentTypeNameContact,
				Value:     &contactPJson,
			}
			status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
			assert.Equal(t, http.StatusCreated, status)
			assert.Equal(t, updatedAt, hubspotDocument.Timestamp)
		}
	}

	// 101 unique values
	updatedAt = updatedAt + 100
	value1 = "val3"
	jsonContact = fmt.Sprintf(jsonContactModel, documentID, updatedAt, createdAt, updatedAt, property1, value1, cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, updatedAt, hubspotDocument.Timestamp)

	w = sendGetCRMObjectValuesByPropertyNameReq(r, project.ID, agent, model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContact, property1)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &property1Values)
	assert.Nil(t, err)

	// should come in ordered for top 5
	for i := range property1Values[:6] {
		if i == 0 {
			assert.Equal(t, "$none", property1Values[i])
			continue
		}

		assert.Equal(t, fmt.Sprintf("%s_%d", property1, 5-i), property1Values[i])
	}
}

func TestHubspotDocumentTimestamp(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	documentID := 1
	createdDate := time.Now().AddDate(0, 0, -1).Unix() * 1000
	cuid := U.RandomLowerAphaNumString(5)

	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	// document first created
	jsonContact := fmt.Sprintf(jsonContactModel, documentID, createdDate, createdDate, createdDate, "lead", cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdDate, hubspotDocument.Timestamp)

	document, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"1"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 2, len(document))
	createdIndex := -1
	updatedIndex := -1
	if document[0].Action == model.HubspotDocumentActionCreated {
		createdIndex = 0
		updatedIndex = 1
	} else {
		createdIndex = 1
		updatedIndex = 0
	}

	assert.Equal(t, model.HubspotDocumentActionCreated, document[createdIndex].Action)
	assert.Equal(t, model.HubspotDocumentActionUpdated, document[updatedIndex].Action)
	assert.Greater(t, document[updatedIndex].CreatedAt.UnixNano(), document[createdIndex].CreatedAt.UnixNano())

	// document updated
	updatedTime := createdDate + 100
	jsonContact = fmt.Sprintf(jsonContactModel, documentID, updatedTime, createdDate, updatedTime, "lead", cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, updatedTime, hubspotDocument.Timestamp)

	// new document missing createddate should use fallback key
	jsonContactModel = `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	documentID = 2
	cuid = U.RandomLowerAphaNumString(5)
	updatedTime = createdDate + 100
	jsonContact = fmt.Sprintf(jsonContactModel, documentID, createdDate, updatedTime, "lead", cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdDate, hubspotDocument.Timestamp)
}

func TestHubspotDocumentDelete(t *testing.T) {
	// The test case first creates the document in the system, and then processes the delete contact
	// operation for the same
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	documentID := 1
	createdDate := time.Now().AddDate(0, 0, -1).Unix() * 1000
	cuid := U.RandomLowerAphaNumString(5)

	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	jsonContact := fmt.Sprintf(jsonContactModel, documentID, createdDate, createdDate, createdDate, "lead", cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	createdDate = time.Now().AddDate(0, 0, -1).Unix() * 1000
	lastmodifieddate := time.Now().Format(model.HubspotDateTimeLayout)

	jsonContactModel = `{
		"id": %d,
		"properties": {
			"createdate": "%d",
			"email": "test453@test.com",
			"firstname": "11",
			"hs_object_id": "451",
			"lastmodifieddate": "%s",
			"lastname": "11"
		},
		"createdAt": "%s",
		"updatedAt": "%s", 
		"archived": true,
		"archivedAt": "%s"
	}`

	jsonContact = fmt.Sprintf(jsonContactModel, documentID, createdDate, lastmodifieddate, lastmodifieddate, lastmodifieddate, lastmodifieddate)
	var myStoredVariable map[string]interface{}
	json.Unmarshal([]byte(jsonContact), &myStoredVariable)
	tempJson := myStoredVariable["properties"].((map[string]interface{}))
	get_lastmodifieddate := tempJson["lastmodifieddate"].(string)
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionDeleted,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, model.HubspotDocumentActionDeleted, hubspotDocument.Action)
	tm, err := time.Parse(model.HubspotDateTimeLayout, U.GetPropertyValueAsString(get_lastmodifieddate))
	if err != nil {
		log.WithError(err).Error("Error while parsing lastmodifieddate to HubspotDateTimeLayout.")
	}
	assert.Equal(t, tm.UnixNano()/int64(time.Millisecond), hubspotDocument.Timestamp)
	var contact map[string]interface{}
	err = json.Unmarshal(hubspotDocument.Value.RawMessage, &contact)
	assert.Nil(t, err)
	assert.Equal(t, true, reflect.DeepEqual(myStoredVariable, contact))

	// A negative test case, where the document which not present in our system, should not be inserted.
	jsonContactModel = `{
		"id": %d,
		"properties": {
			"createdate": "%d",
			"email": "test453@test.com",
			"firstname": "11",
			"hs_object_id": "451",
			"lastmodifieddate": "%s",
			"lastname": "11"
		},
		"createdAt": "%s",
		"updatedAt": "%s", 
		"archived": true,
		"archivedAt": "%s"
	}`

	documentID = rand.Intn(100)
	createdDate = time.Now().AddDate(0, 0, -1).Unix() * 1000
	lastmodifieddate = time.Now().Format(model.HubspotDateTimeLayout)
	jsonContact = fmt.Sprintf(jsonContactModel, documentID, createdDate, lastmodifieddate, lastmodifieddate, lastmodifieddate, lastmodifieddate)
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionDeleted,
		Value:     &contactPJson,
	}

	project.ID = uint64(rand.Intn(10000))
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusOK, status)
	document, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", documentID)}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionDeleted})
	assert.NotEqual(t, http.StatusFound, status)
	assert.Equal(t, 0, len(document))
}

func TestHubspotSyncJobDocumentDeleteAndMerge(t *testing.T) {
	// Test case to process deleted contact. A new contact is created and the corresponding deletion request is generated.
	// Later, the request is been processed by sync job. Thereafter, the existence of deleted-user-property
	// "$hubspot_contact_deleted" is verified.
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	deletionDocumentID := rand.Intn(100)
	createdDate := time.Now().AddDate(0, 0, -1).Unix() * 1000
	cuid := getRandomEmail()
	leadguid := U.RandomString(5)

	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	jsonContact := fmt.Sprintf(jsonContactModel, deletionDocumentID, createdDate, createdDate, createdDate, "lead", cuid, leadguid)
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionCreated,
		Value:     &contactPJson,
	}

	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	createdDate = time.Now().AddDate(0, 0, -1).Unix() * 1000
	lastmodifieddateHSLayout := time.Now().UTC().Format(model.HubspotDateTimeLayout)

	jsonContactModel = `{
		"id": %d,
		"properties": {
			"createdate": "%d",
			"email": "test453@test.com",
			"firstname": "11",
			"hs_object_id": "451",
			"lastmodifieddate": "%s",
			"lastname": "11"
		},
		"createdAt": "%s",
		"updatedAt": "%s", 
		"archived": true,
		"archivedAt": "%s"
	}`

	jsonContact = fmt.Sprintf(jsonContactModel, deletionDocumentID, createdDate, lastmodifieddateHSLayout, createdDate, createdDate, createdDate)
	var myStoredVariable map[string]interface{}
	json.Unmarshal([]byte(jsonContact), &myStoredVariable)
	tempJson := myStoredVariable["properties"].((map[string]interface{}))
	get_lastmodifieddate := tempJson["lastmodifieddate"].(string)
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionDeleted,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, model.HubspotDocumentActionDeleted, hubspotDocument.Action)
	tm, err := time.Parse(model.HubspotDateTimeLayout, U.GetPropertyValueAsString(get_lastmodifieddate))
	if err != nil {
		log.WithError(err).Error("Error while parsing lastmodifieddate to HubspotDateTimeLayout.")
	}
	assert.Equal(t, tm.UnixNano()/int64(time.Millisecond), hubspotDocument.Timestamp)
	var contact map[string]interface{}
	err = json.Unmarshal(hubspotDocument.Value.RawMessage, &contact)
	assert.Nil(t, err)
	assert.Equal(t, true, reflect.DeepEqual(myStoredVariable, contact))

	// Test case to process contact merge contact. Three new contacts are created. First 2 been the contacts that
	// needs to be merged, and the third contact being the primary contact. Later, the request is been processed
	// by sync job. Thereafter, the existence of merge-user-properties "$hubspot_contact_merged" and
	// "$hubspot_contact_primary_contact" are verified.
	firstDocumentID := rand.Intn(100) + 100
	cuid_first := getRandomEmail()
	leadguid = U.RandomString(5)
	lastmodifieddate := time.Now().UTC().Unix() * 1000

	jsonContactModel = `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	jsonContact = fmt.Sprintf(jsonContactModel, firstDocumentID, createdDate, createdDate, lastmodifieddate, "lead", cuid_first, leadguid)
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionCreated,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", firstDocumentID)}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)

	secondDocumentID := rand.Intn(100) + 200
	cuid = getRandomEmail()
	leadguid = U.RandomString(5)
	lastmodifieddate = time.Now().UTC().Unix() * 1000

	jsonContactModel = `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	jsonContact = fmt.Sprintf(jsonContactModel, secondDocumentID, createdDate, createdDate, lastmodifieddate, "lead", cuid, leadguid)
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionCreated,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", secondDocumentID)}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)

	jsonContactModel = `{
		"vid": %d,
		"addedAt": %d,
		"canonical-vid": %d,
		"merged-vids": [%d,%d,%d],
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	}`

	primaryDocumentID := rand.Intn(100) + 300
	cuid = getRandomEmail()
	leadguid = U.RandomString(5)
	lastmodifieddate = time.Now().UTC().Unix() * 1000
	jsonContact = fmt.Sprintf(jsonContactModel, primaryDocumentID, createdDate, primaryDocumentID, primaryDocumentID, firstDocumentID, secondDocumentID, createdDate, lastmodifieddate, "lead", cuid, leadguid)
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionCreated,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", primaryDocumentID)}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	json.Unmarshal([]byte(jsonContact), &myStoredVariable)
	err = json.Unmarshal(hubspotDocument.Value.RawMessage, &contact)
	assert.Nil(t, err)
	assert.Equal(t, true, reflect.DeepEqual(myStoredVariable, contact))

	// Test case which creates a random user with same cuid as that of cuid of firstDocumentID. Later, the request
	// is been processed by sync job. Thereafter, the NON-existence of merge-user-properties "$hubspot_contact_merged"
	// and "$hubspot_contact_primary_contact" are verified.
	randomUserDocumentID := rand.Intn(100) + 400
	leadguid = U.RandomString(5)
	lastmodifieddate = time.Now().UTC().Unix() * 1000

	jsonContactModel = `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	jsonContact = fmt.Sprintf(jsonContactModel, randomUserDocumentID, createdDate, createdDate, lastmodifieddate, "lead", cuid_first, leadguid)
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionCreated,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// Processing the sync job altogether for all the test cases.
	enrichStatus, _ := IntHubspot.Sync(project.ID, 3)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	// Verification for contact delete test case.
	deleteDocument, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", deletionDocumentID)}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	user, status := store.GetStore().GetUser(project.ID, deleteDocument[0].UserId)
	assert.Equal(t, http.StatusFound, status)
	properitesMap := make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	assert.Equal(t, true, properitesMap[model.UserPropertyHubspotContactDeleted])

	// Verification for contact merge test case.
	firstDocument, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", firstDocumentID)}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, firstDocument[0].UserId)
	assert.Equal(t, http.StatusFound, status)
	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	assert.Equal(t, true, properitesMap[model.UserPropertyHubspotContactMerged])
	assert.Equal(t, primaryDocumentID, int(properitesMap[model.UserPropertyHubspotContactPrimaryContact].(float64)))

	secondDocument, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", secondDocumentID)}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, secondDocument[0].UserId)
	assert.Equal(t, http.StatusFound, status)
	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	assert.Equal(t, true, properitesMap[model.UserPropertyHubspotContactMerged])
	assert.Equal(t, primaryDocumentID, int(properitesMap[model.UserPropertyHubspotContactPrimaryContact].(float64)))

	// Verification for random user test case with same cuid.
	randomUserDocument, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", randomUserDocumentID)}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, randomUserDocument[0].UserId)
	assert.Equal(t, http.StatusFound, status)
	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	_, exists := (properitesMap)[model.UserPropertyHubspotContactMerged]
	assert.Equal(t, false, exists)
	_, exists = (properitesMap)[model.UserPropertyHubspotContactPrimaryContact]
	assert.Equal(t, false, exists)
}

func TestHubspotPropertyDetails(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	refreshToken := U.RandomLowerAphaNumString(5)
	instancURL := U.RandomLowerAphaNumString(5)
	errCode := store.GetStore().UpdateAgentIntSalesforce(agent.UUID,
		refreshToken,
		instancURL,
	)
	assert.Equal(t, http.StatusAccepted, errCode)

	status := IntHubspot.CreateOrGetHubspotEventName(project.ID)
	assert.Equal(t, http.StatusOK, status)

	createdDate := time.Now().Unix()
	eventNameCreated := U.EVENT_NAME_HUBSPOT_CONTACT_CREATED

	// datetime property detail
	eventNameUpdated := U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED
	dtPropertyName1 := "last_visit"
	dtPropertyValue1 := createdDate * 1000
	dtPropertyName2 := "next_visit"
	dtPropertyValue2 := createdDate * 1000

	// numerical property detail
	numPropertyName1 := "vists"
	numPropertyValue1 := 15
	numPropertyName2 := "views"
	numPropertyValue2 := 10

	// datetime property
	dtEnKey1 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact,
		U.GetPropertyValueAsString(dtPropertyName1),
	)
	dtEnKey2 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact,
		U.GetPropertyValueAsString(dtPropertyName2),
	)

	// numerical property
	numEnKey1 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact,
		U.GetPropertyValueAsString(numPropertyName1),
	)
	numEnKey2 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact,
		U.GetPropertyValueAsString(numPropertyName2),
	)

	// datetime property details
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, dtEnKey1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, dtEnKey2, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, "", dtEnKey1, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", dtEnKey2, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, dtEnKey1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, dtEnKey2, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	// numerical property details
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, numEnKey1, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, numEnKey2, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, "", numEnKey1, U.PropertyTypeNumerical, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", numEnKey2, U.PropertyTypeNumerical, true, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, numEnKey1, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, numEnKey2, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)

	// new document missing createddate should use fallback key
	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		"createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" },
		  "%s":{"value":"%d"},
		  "%s":{"value":"%d"},
		  "%s":{"value":"%d"},
		  "%s":{"value":"%d"}
		},
		"identity-profiles": [
		  {
			"vid": %d,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	documentID := 2
	cuid := U.RandomLowerAphaNumString(5)
	updatedTime := createdDate*1000 + 100
	jsonContact := fmt.Sprintf(jsonContactModel, documentID, createdDate*1000, createdDate*1000, updatedTime, "lead", dtPropertyName1, dtPropertyValue1, dtPropertyName2, dtPropertyValue2, numPropertyName1, numPropertyValue1, numPropertyName2, numPropertyValue2, documentID, cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdDate*1000, hubspotDocument.Timestamp)

	allStatus, _ := IntHubspot.Sync(project.ID, 3)
	for i := range allStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, allStatus[i].Status)
	}

	configs := make(map[string]interface{})
	configs["rollupLookback"] = 1
	event_user_cache.DoRollUpSortedSet(configs)
	properties, err := store.GetStore().GetPropertiesByEvent(project.ID, eventNameCreated, 2500, 1)
	assert.Nil(t, err)
	assert.Contains(t, properties[U.PropertyTypeDateTime], dtEnKey1, dtEnKey2)
	assert.Contains(t, properties[U.PropertyTypeNumerical], numEnKey1, numEnKey2)

	properties, err = store.GetStore().GetUserPropertiesByProject(project.ID, 100, 10)
	assert.Nil(t, err)
	assert.Contains(t, properties[U.PropertyTypeDateTime], dtEnKey1, dtEnKey2)
	assert.Contains(t, properties[U.PropertyTypeNumerical], numEnKey1, numEnKey2)

	query := model.Query{
		From: createdDate - 500,
		To:   updatedTime + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: "$hubspot_contact_created",
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:   model.PropertyEntityEvent,
				Property: dtEnKey1,
			},
			{
				Entity:   model.PropertyEntityEvent,
				Property: dtEnKey2,
			},
			{
				Entity:   model.PropertyEntityEvent,
				Property: numEnKey1,
			},
			{
				Entity:   model.PropertyEntityEvent,
				Property: numEnKey2,
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, status, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, result.Headers, dtEnKey1, dtEnKey2, numEnKey1, numEnKey2)
	count := 0
	for i := range result.Headers[:len(result.Headers)-1] {
		if result.Headers[i] == dtEnKey1 || result.Headers[i] == dtEnKey2 {
			assert.Equal(t, fmt.Sprint(createdDate), result.Rows[0][i])
			count++
		}
		if result.Headers[i] == numEnKey1 {
			assert.Equal(t, fmt.Sprint(numPropertyValue1), result.Rows[0][i])
			count++
		}

		if result.Headers[i] == numEnKey2 {
			assert.Equal(t, fmt.Sprint(numPropertyValue2), result.Rows[0][i])
			count++
		}
	}
	assert.Equal(t, 4, count)
}

func sendCreateHubspotDocumentRequest(projectID uint64, r *gin.Engine, agent *model.Agent, documentType string, documentValue *map[string]interface{}) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
		return nil
	}

	payload := map[string]interface{}{
		"project_id": projectID,
		"type_alias": documentType,
		"value":      documentValue,
	}

	rb := U.NewRequestBuilder(http.MethodPost, "http://localhost:8089/data_service/hubspot/documents/add").
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		}).WithPostParams(payload)

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating request")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHubspotAPINullCharacter(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	createdAt := fmt.Sprint(time.Now().AddDate(0, 0, -11).Unix())
	cuid := U.RandomLowerAphaNumString(5)

	jsonContactMap := map[string]interface{}{
		"vid":     1,
		"addedAt": createdAt,
		"properties": map[string]map[string]interface{}{
			"createdate":       {"value": createdAt},
			"lastmodifieddate": {"value": createdAt},
			"lifecyclestage":   {"value": "lead\u0000"}, // unicode null
		},
		"identity-profiles": []map[string]interface{}{
			{
				"vid": 1,
				"identities": []map[string]interface{}{
					{
						"type":  "EMAIL",
						"value": cuid,
					},
					{
						"type":  "LEAD_GUID",
						"value": "123-45",
					},
				},
			},
		},
	}
	w := sendCreateHubspotDocumentRequest(project.ID, r, agent, model.HubspotDocumentTypeNameContact, &jsonContactMap)
	assert.Equal(t, http.StatusCreated, w.Code)

	document, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"1"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 1, len(document))

	var contact IntHubspot.Contact
	err = json.Unmarshal(document[0].Value.RawMessage, &contact)
	assert.Nil(t, err)
	assert.Equal(t, "lead ", contact.Properties["lifecyclestage"].Value)

	updateDate := fmt.Sprint(time.Now().AddDate(0, 0, -10).Unix())
	jsonContactMap["vid"] = 2
	jsonContactMap["properties"] = map[string]map[string]interface{}{
		"createdate":       {"value": createdAt},
		"lastmodifieddate": {"value": updateDate},
		"lifecyclestage":   {"value": "lead\x00"}, // utf null
	}

	w = sendCreateHubspotDocumentRequest(project.ID, r, agent, model.HubspotDocumentTypeNameContact, &jsonContactMap)
	assert.Equal(t, http.StatusCreated, w.Code)

	document, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"2"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 1, len(document))

	err = json.Unmarshal(document[0].Value.RawMessage, &contact)
	assert.Nil(t, err)
	assert.Equal(t, "lead ", contact.Properties["lifecyclestage"].Value)

	// test GetFilteredNullCharacterBytes
	alteredNullcharacterBytes := make([]byte, len(U.NullcharBytes)*2)
	for i := 0; i < len(U.NullcharBytes); i++ {
		alteredNullcharacterBytes[i*2] = 0x22
		alteredNullcharacterBytes[i*2+1] = U.NullcharBytes[i]
	}

	newBytes := U.RemoveNullCharacterBytes(alteredNullcharacterBytes)
	assert.Equal(t, alteredNullcharacterBytes, newBytes)
}

func TestHubspotCreateActionUpdatedOnCreate(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	documentID := 1
	createdDate := time.Now().AddDate(0, 0, -1).Unix() * 1000
	cuid := U.RandomLowerAphaNumString(5)

	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	// document first created, should add updated document
	jsonContact := fmt.Sprintf(jsonContactModel, documentID, createdDate, createdDate, createdDate, "lead", cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdDate, hubspotDocument.Timestamp)

	//enrich job, create contact created and contact updated event
	enrichStatus, _ := IntHubspot.Sync(project.ID, 3)
	projectIndex := -1
	for i := range enrichStatus {
		if enrichStatus[i].ProjectId == project.ID {
			projectIndex = i
			break
		}
	}
	assert.Equal(t, project.ID, enrichStatus[projectIndex].ProjectId)
	assert.Equal(t, "success", enrichStatus[projectIndex].Status)

	query := []model.Query{
		{
			From: createdDate/1000 - 500,
			To:   createdDate/1000 + 500,
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
				},
				{
					Name: U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
		},
		{
			From: createdDate/1000 - 500,
			To:   createdDate/1000 + 500,
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
				},
				{
					Name: U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
		},
	}

	result, status := store.GetStore().RunEventsGroupQuery(query, project.ID)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, 2, len(result.Results[0].Rows))
	for i := range result.Results { // two events, one on each
		assert.Contains(t, []string{U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED, U.EVENT_NAME_HUBSPOT_CONTACT_CREATED}, result.Results[i].Rows[0][1])
		assert.Equal(t, float64(1), result.Results[i].Rows[0][2])
	}

	// One unique user
	query = []model.Query{
		{
			From: createdDate/1000 - 500,
			To:   createdDate/1000 + 500,
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
				},
				{
					Name: U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAnyGivenEvent,
		},
	}

	result, status = store.GetStore().RunEventsGroupQuery(query, project.ID)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, 1, len(result.Results[0].Rows))
	assert.Equal(t, float64(1), result.Results[0].Rows[0][0])
}

func TestHubspotUseLastModifiedTimestampAsDefault(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	documentID := 1
	createdDate := time.Now().AddDate(0, 0, -1).Unix() * 1000
	lastModifiedDate := createdDate + 100

	email := getRandomEmail()

	// First contact created with createdate in document timestamp
	contact := IntHubspot.Contact{
		Vid: int64(documentID),
		Properties: map[string]IntHubspot.Property{
			"createdate":       {Value: fmt.Sprintf("%d", createdDate)},
			"lastmodifieddate": {Value: fmt.Sprintf("%d", lastModifiedDate)},
			"lifecyclestage":   {Value: "lead"},
		},
		IdentityProfiles: []IntHubspot.ContactIdentityProfile{
			{
				Identities: []IntHubspot.ContactIdentity{
					{
						Type:  "EMAIL",
						Value: email,
					},
					{
						Type:  "LEAD_GUID",
						Value: "123-45",
					},
				},
			},
		},
	}

	enJSON, err := json.Marshal(contact)
	assert.Nil(t, err)
	contactPJson := postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}
	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdDate, hubspotDocument.Timestamp)

	// updated contact with lifecyclestage to customer, to_customer_timestamp is set
	toCustomerTimestamp := lastModifiedDate + 100
	contact = IntHubspot.Contact{
		Vid: int64(documentID),
		Properties: map[string]IntHubspot.Property{
			"createdate":            {Value: fmt.Sprintf("%d", createdDate)},
			"lastmodifieddate":      {Value: fmt.Sprintf("%d", toCustomerTimestamp-1)},
			"lifecyclestage":        {Value: "customer"},
			"to_customer_timestamp": {Value: fmt.Sprintf("%d", toCustomerTimestamp)},
		},
		IdentityProfiles: []IntHubspot.ContactIdentityProfile{
			{
				Identities: []IntHubspot.ContactIdentity{
					{
						Type:  "EMAIL",
						Value: email,
					},
					{
						Type:  "LEAD_GUID",
						Value: "123-45",
					},
				},
			},
		},
	}

	enJSON, err = json.Marshal(contact)
	assert.Nil(t, err)
	contactPJson = postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, toCustomerTimestamp-1, hubspotDocument.Timestamp)

	// updated contact with lifecyclestage to junk, to_junk_time is missing
	toJunkTimestamp := toCustomerTimestamp + 500
	contact = IntHubspot.Contact{
		Vid: int64(documentID),
		Properties: map[string]IntHubspot.Property{
			"createdate":            {Value: fmt.Sprintf("%d", createdDate)},
			"lastmodifieddate":      {Value: fmt.Sprintf("%d", toJunkTimestamp-1)},
			"lifecyclestage":        {Value: "junk"},
			"to_customer_timestamp": {Value: fmt.Sprintf("%d", toCustomerTimestamp)},
		},
		IdentityProfiles: []IntHubspot.ContactIdentityProfile{
			{
				Identities: []IntHubspot.ContactIdentity{
					{
						Type:  "EMAIL",
						Value: email,
					},
					{
						Type:  "LEAD_GUID",
						Value: "123-45",
					},
				},
			},
		},
	}

	enJSON, err = json.Marshal(contact)
	assert.Nil(t, err)
	contactPJson = postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, toJunkTimestamp-1, hubspotDocument.Timestamp)

	smartEventRule := &model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceHubspot,
		ObjectType:           model.HubspotDocumentTypeNameContact,
		Description:          "hubspot contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "lifecyclestage",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "lead",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "lead",
						Operator:      model.COMPARE_NOT_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: model.TimestampReferenceTypeDocument,
	}

	eventNameLifecycleStageLead := "lifecyclestage_lead"
	requestPayload := make(map[string]interface{})
	requestPayload["name"] = eventNameLifecycleStageLead
	requestPayload["expr"] = smartEventRule

	w := sendCreateSmartEventFilterReq(r, project.ID, agent, &requestPayload)
	assert.Equal(t, http.StatusCreated, w.Code)

	smartEventRule.Filters[0].Rules[0].Value = "customer" // current
	smartEventRule.Filters[0].Rules[1].Value = "customer" // previous
	smartEventRule.TimestampReferenceField = "to_customer_timestamp"
	eventNameLifecycleStageCustomer := "lifecyclestage_customer"
	requestPayload = make(map[string]interface{})
	requestPayload["name"] = eventNameLifecycleStageCustomer
	requestPayload["expr"] = smartEventRule

	w = sendCreateSmartEventFilterReq(r, project.ID, agent, &requestPayload)
	assert.Equal(t, http.StatusCreated, w.Code)

	smartEventRule.Filters[0].Rules[0].Value = "junk"       // current
	smartEventRule.Filters[0].Rules[1].Value = "junk"       // previous
	smartEventRule.TimestampReferenceField = "to_junk_time" // doest not exist, should use lastmodified timestamp
	eventNameLifecycleStageJunk := "lifecyclestage_junk"
	requestPayload = make(map[string]interface{})
	requestPayload["name"] = eventNameLifecycleStageJunk
	requestPayload["expr"] = smartEventRule

	w = sendCreateSmartEventFilterReq(r, project.ID, agent, &requestPayload)
	assert.Equal(t, http.StatusCreated, w.Code)

	//enrich job, create contact created and contact updated event
	enrichStatus, _ := IntHubspot.Sync(project.ID, 1)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectId)
	assert.Equal(t, util.CRM_SYNC_STATUS_SUCCESS, enrichStatus[0].Status)
	assert.Equal(t, util.CRM_SYNC_STATUS_SUCCESS, enrichStatus[1].Status)
	assert.Equal(t, util.CRM_SYNC_STATUS_SUCCESS, enrichStatus[2].Status)

	query := model.Query{
		From: createdDate/1000 - 500,
		To:   toJunkTimestamp/1000 + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: eventNameLifecycleStageLead,
			},
			{
				Name: eventNameLifecycleStageCustomer,
			},
			{
				Name: eventNameLifecycleStageJunk,
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:   model.PropertyEntityEvent,
				Property: util.EP_TIMESTAMP,
			},
		},
	}

	result, status, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, 3, len(result.Rows))
	eventNameTimestamp := make(map[string]int64)
	for i := range result.Rows {
		timestamp, _ := util.GetPropertyValueAsFloat64(result.Rows[i][1])
		eventNameTimestamp[result.Rows[i][0].(string)] = int64(timestamp)
	}
	assert.Equal(t, lastModifiedDate/1000+1, eventNameTimestamp[eventNameLifecycleStageLead]) // timestamp+1
	assert.Equal(t, toCustomerTimestamp/1000, eventNameTimestamp[eventNameLifecycleStageCustomer])
	assert.Equal(t, (toJunkTimestamp-1)/1000+1, eventNameTimestamp[eventNameLifecycleStageJunk]) // timestamp+1
}

func sendGetHubspotSyncInfo(r *gin.Engine, isFirstTime bool) *httptest.ResponseRecorder {

	url := "http://localhost:8089/data_service/hubspot/documents/sync_info?is_first_time="
	if isFirstTime {
		url = url + "true"
	} else {
		url = url + "false"
	}

	rb := U.NewRequestBuilder(http.MethodGet, url)
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating request")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendUpdateHubspotSyncInfo(r *gin.Engine, updateInfo map[string]interface{}, isFirstTime bool) *httptest.ResponseRecorder {

	url := "http://localhost:8089/data_service/hubspot/documents/sync_info?is_first_time="
	if isFirstTime {
		url = url + "true"
	} else {
		url = url + "false"
	}
	rb := U.NewRequestBuilder(http.MethodPost, fmt.Sprintf(url)).
		WithPostParams(updateInfo)
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating request")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHubspotFirstSyncStatus(t *testing.T) {
	project1, agent1, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)
	H.InitDataServiceRoutes(r)

	project2, agent2, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	w := sendUpdateProjectSettingReq(r, project1.ID, agent1, map[string]interface{}{
		"int_hubspot_api_key": "1234", "int_hubspot": true,
	})
	assert.Equal(t, http.StatusOK, w.Code)

	w = sendUpdateProjectSettingReq(r, project2.ID, agent2, map[string]interface{}{
		"int_hubspot_api_key": "1234", "int_hubspot": true,
	})
	assert.Equal(t, http.StatusOK, w.Code)

	project, status := store.GetStore().GetProjectSetting(project1.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, false, project.IntHubspotFirstTimeSynced)
	project, status = store.GetStore().GetProjectSetting(project2.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, false, project.IntHubspotFirstTimeSynced)

	w = sendGetHubspotSyncInfo(r, true)
	assert.Equal(t, http.StatusFound, w.Code)

	var jsonResponseMap map[string]map[string]map[string]interface{}
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.Equal(t, float64(0), jsonResponseMap["last_sync_info"][fmt.Sprintf("%d", project1.ID)]["company"])
	assert.Equal(t, float64(0), jsonResponseMap["last_sync_info"][fmt.Sprintf("%d", project1.ID)]["contact"])
	assert.Equal(t, float64(0), jsonResponseMap["last_sync_info"][fmt.Sprintf("%d", project1.ID)]["deal"])

	assert.Equal(t, float64(0), jsonResponseMap["last_sync_info"][fmt.Sprintf("%d", project2.ID)]["company"])
	assert.Equal(t, float64(0), jsonResponseMap["last_sync_info"][fmt.Sprintf("%d", project2.ID)]["contact"])
	assert.Equal(t, float64(0), jsonResponseMap["last_sync_info"][fmt.Sprintf("%d", project2.ID)]["deal"])

	payload := map[string]interface{}{
		"status": "success",
		"success": []map[string]interface{}{
			{
				"project_id": project1.ID,
				"doc_type":   "contact",
				"status":     "success",
			},
			{
				"project_id": project1.ID,
				"doc_type":   "company",
				"status":     "success",
			},
			{
				"project_id": project1.ID,
				"doc_type":   "deals",
				"status":     "success",
			},
			{
				"project_id": project2.ID,
				"doc_type":   "contact",
				"status":     "success",
			},
			{
				"project_id": project2.ID,
				"doc_type":   "company",
				"status":     "success",
			},
			{
				"project_id": project2.ID,
				"doc_type":   "deals",
				"status":     "success",
			},
		},
	}
	w = sendUpdateHubspotSyncInfo(r, payload, true)
	assert.Equal(t, http.StatusOK, w.Code)
	project, status = store.GetStore().GetProjectSetting(project1.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, true, project.IntHubspotFirstTimeSynced)
	project, status = store.GetStore().GetProjectSetting(project2.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, true, project.IntHubspotFirstTimeSynced)

}

func TestHubspotSyncInfo(t *testing.T) {
	project1, agent1, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)
	H.InitDataServiceRoutes(r)

	project2, agent2, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	t.Run("HubspotBeforeIntegrationResponse", func(t *testing.T) {
		w := sendGetProjectSettingsReq(r, project1.ID, agent1)
		assert.Equal(t, http.StatusOK, w.Code)
		var jsonResponseMap map[string]interface{}
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		err = json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.Nil(t, err)
		assert.Nil(t, jsonResponseMap["int_hubspot_api_key"])
		assert.Equal(t, false, jsonResponseMap["int_hubspot"])
		assert.Nil(t, jsonResponseMap["int_hubspot_first_time_synced"])
		assert.Nil(t, jsonResponseMap["int_hubspot_portal_id"])
		assert.Nil(t, jsonResponseMap["int_hubspot_sync_info"])

		w = sendGetProjectSettingsReq(r, project2.ID, agent2)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ = ioutil.ReadAll(w.Body)
		err = json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.Nil(t, err)
		assert.Nil(t, jsonResponseMap["int_hubspot_api_key"])
		assert.Equal(t, false, jsonResponseMap["int_hubspot"])
		assert.Nil(t, jsonResponseMap["int_hubspot_first_time_synced"])
		assert.Nil(t, jsonResponseMap["int_hubspot_portal_id"])
		assert.Nil(t, jsonResponseMap["int_hubspot_sync_info"])
	})

	t.Run("HubspotAfterIntegrationResponse", func(t *testing.T) {
		w := sendUpdateProjectSettingReq(r, project1.ID, agent1, map[string]interface{}{
			"int_hubspot_api_key": "1234", "int_hubspot": true,
		})
		assert.Equal(t, http.StatusOK, w.Code)

		w = sendUpdateProjectSettingReq(r, project2.ID, agent2, map[string]interface{}{
			"int_hubspot_api_key": "1234", "int_hubspot": true,
		})

		w = sendGetProjectSettingsReq(r, project1.ID, agent1)
		assert.Equal(t, http.StatusOK, w.Code)
		var jsonResponseMap map[string]interface{}
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		err = json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.Nil(t, err)
		assert.Equal(t, "1234", jsonResponseMap["int_hubspot_api_key"])
		assert.Equal(t, true, jsonResponseMap["int_hubspot"])
		assert.Equal(t, float64(0), jsonResponseMap["int_hubspot_portal_id"])
		assert.Nil(t, jsonResponseMap["int_hubspot_sync_info"])

		w = sendGetProjectSettingsReq(r, project2.ID, agent2)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ = ioutil.ReadAll(w.Body)
		err = json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.Nil(t, err)
		assert.Equal(t, "1234", jsonResponseMap["int_hubspot_api_key"])
		assert.Equal(t, true, jsonResponseMap["int_hubspot"])
		assert.Equal(t, float64(0), jsonResponseMap["int_hubspot_portal_id"])
		assert.Nil(t, jsonResponseMap["int_hubspot_sync_info"])
	})

	t.Run("HubSyncInfoBeforeFirstRun", func(t *testing.T) {
		w := sendGetHubspotSyncInfo(r, true)
		assert.Equal(t, http.StatusFound, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var hubspotSyncInfo model.HubspotSyncInfo
		err = json.Unmarshal(jsonResponse, &hubspotSyncInfo)
		assert.Nil(t, err)
		assert.Equal(t, int64(0), hubspotSyncInfo.LastSyncInfo[project1.ID]["contact"])
		assert.Equal(t, int64(0), hubspotSyncInfo.LastSyncInfo[project1.ID]["deals"])
		assert.Equal(t, int64(0), hubspotSyncInfo.LastSyncInfo[project1.ID]["companies"])

		assert.Equal(t, int64(0), hubspotSyncInfo.LastSyncInfo[project2.ID]["contact"])
		assert.Equal(t, int64(0), hubspotSyncInfo.LastSyncInfo[project2.ID]["deals"])
		assert.Equal(t, int64(0), hubspotSyncInfo.LastSyncInfo[project2.ID]["companies"])

		w = sendGetHubspotSyncInfo(r, false)
		assert.Equal(t, http.StatusFound, w.Code)
		jsonResponse, _ = ioutil.ReadAll(w.Body)
		var hubspotSyncInfo2 model.HubspotSyncInfo
		err = json.Unmarshal(jsonResponse, &hubspotSyncInfo2)
		assert.Nil(t, err)
		assert.Nil(t, hubspotSyncInfo2.LastSyncInfo[project1.ID])

		assert.Nil(t, hubspotSyncInfo2.LastSyncInfo[project2.ID])
	})

	t.Run("HubSyncInfoFirstRun", func(t *testing.T) {
		payload := map[string]interface{}{
			"status": "success",
			"success": []map[string]interface{}{
				{
					"project_id": project1.ID,
					"doc_type":   "contact",
					"status":     "success",
					"timestamp":  123,
				},
				{
					"project_id": project1.ID,
					"doc_type":   "company",
					"status":     "success",
					"timestamp":  1234,
				},
				{
					"project_id": project1.ID,
					"doc_type":   "deals",
					"status":     "success",
					"timestamp":  12345,
				},
				{
					"project_id": project2.ID,
					"doc_type":   "contact",
					"status":     "success",
					"timestamp":  123456,
				},
				{
					"project_id": project2.ID,
					"doc_type":   "company",
					"status":     "success",
					"timestamp":  1234567,
				},
				{
					"project_id": project2.ID,
					"doc_type":   "deals",
					"status":     "success",
					"timestamp":  12345678,
				},
			},
		}

		w := sendUpdateHubspotSyncInfo(r, payload, true)
		assert.Equal(t, http.StatusOK, w.Code)
		w = sendGetProjectSettingsReq(r, project1.ID, agent1)
		assert.Equal(t, http.StatusOK, w.Code)
		var jsonResponseMap map[string]interface{}
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		err = json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.Nil(t, err)
		assert.Equal(t, true, jsonResponseMap["int_hubspot_first_time_synced"])
		projectSyncInfo := jsonResponseMap["int_hubspot_sync_info"].(map[string]interface{})
		assert.Equal(t, float64(123), projectSyncInfo["contact"])
		assert.Equal(t, float64(1234), projectSyncInfo["company"])
		assert.Equal(t, float64(12345), projectSyncInfo["deals"])

		w = sendGetProjectSettingsReq(r, project2.ID, agent2)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponseMap = map[string]interface{}{}
		jsonResponse, _ = ioutil.ReadAll(w.Body)
		err = json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.Nil(t, err)
		assert.Equal(t, true, jsonResponseMap["int_hubspot_first_time_synced"])
		projectSyncInfo = jsonResponseMap["int_hubspot_sync_info"].(map[string]interface{})
		assert.Equal(t, float64(123456), projectSyncInfo["contact"])
		assert.Equal(t, float64(1234567), projectSyncInfo["company"])
		assert.Equal(t, float64(12345678), projectSyncInfo["deals"])

		w = sendGetHubspotSyncInfo(r, true)
		assert.Equal(t, http.StatusFound, w.Code)
		jsonResponse, _ = ioutil.ReadAll(w.Body)
		hubspotSyncInfo := model.HubspotSyncInfo{}
		err = json.Unmarshal(jsonResponse, &hubspotSyncInfo)
		assert.Nil(t, err)
		assert.Nil(t, hubspotSyncInfo.LastSyncInfo[project1.ID])

		assert.Nil(t, hubspotSyncInfo.LastSyncInfo[project2.ID])

		w = sendGetHubspotSyncInfo(r, false)
		assert.Equal(t, http.StatusFound, w.Code)
		jsonResponse, _ = ioutil.ReadAll(w.Body)
		hubspotSyncInfo = model.HubspotSyncInfo{}
		err = json.Unmarshal(jsonResponse, &hubspotSyncInfo)
		assert.Nil(t, err)
		assert.NotNil(t, hubspotSyncInfo.LastSyncInfo[project1.ID])

		assert.NotNil(t, hubspotSyncInfo.LastSyncInfo[project2.ID])

		assert.Equal(t, int64(123), hubspotSyncInfo.LastSyncInfo[project1.ID]["contact"])
		assert.Equal(t, int64(1234), hubspotSyncInfo.LastSyncInfo[project1.ID]["company"])
		assert.Equal(t, int64(12345), hubspotSyncInfo.LastSyncInfo[project1.ID]["deals"])

		assert.Equal(t, int64(123456), hubspotSyncInfo.LastSyncInfo[project2.ID]["contact"])
		assert.Equal(t, int64(1234567), hubspotSyncInfo.LastSyncInfo[project2.ID]["company"])
		assert.Equal(t, int64(12345678), hubspotSyncInfo.LastSyncInfo[project2.ID]["deals"])
	})

	t.Run("HubSyncInfoRecentRun", func(t *testing.T) {
		payload := map[string]interface{}{
			"status": "success",
			"success": []map[string]interface{}{
				{
					"project_id": project1.ID,
					"doc_type":   "contact",
					"status":     "success",
					"timestamp":  1234,
				},
				{
					"project_id": project1.ID,
					"doc_type":   "company",
					"status":     "success",
					"timestamp":  1233, // should not update since old timestamp
				},
				{
					"project_id": project1.ID,
					"doc_type":   "deals",
					"status":     "success",
					"timestamp":  12346,
				},
				{
					"project_id": project2.ID,
					"doc_type":   "contact",
					"status":     "success",
					"timestamp":  123455, // should not update
				},
				{
					"project_id": project2.ID,
					"doc_type":   "company",
					"status":     "success",
					"timestamp":  1234567,
				},
				{
					"project_id": project2.ID,
					"doc_type":   "deals",
					"status":     "success",
					"timestamp":  12345678,
				},
			},
		}

		w := sendUpdateHubspotSyncInfo(r, payload, false)
		assert.Equal(t, http.StatusOK, w.Code)

		w = sendGetHubspotSyncInfo(r, false)
		assert.Equal(t, http.StatusFound, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		hubspotSyncInfo := model.HubspotSyncInfo{}
		err = json.Unmarshal(jsonResponse, &hubspotSyncInfo)
		assert.Nil(t, err)
		assert.Equal(t, int64(1234), hubspotSyncInfo.LastSyncInfo[project1.ID]["contact"])
		assert.Equal(t, int64(1234), hubspotSyncInfo.LastSyncInfo[project1.ID]["company"]) // same as before
		assert.Equal(t, int64(12346), hubspotSyncInfo.LastSyncInfo[project1.ID]["deals"])

		assert.Equal(t, int64(123456), hubspotSyncInfo.LastSyncInfo[project2.ID]["contact"]) // same as before
		assert.Equal(t, int64(1234567), hubspotSyncInfo.LastSyncInfo[project2.ID]["company"])
		assert.Equal(t, int64(12345678), hubspotSyncInfo.LastSyncInfo[project2.ID]["deals"])
	})

	t.Run("HubspotSyncFallbackToDocumentTimestamp", func(t *testing.T) {
		jsonContactModel := `{
			"vid": %d,
			"addedAt": %d,
			"properties": {
			  "createdate": { "value": "%d" },
			  "lastmodifieddate": { "value": "%d" },
			  "lifecyclestage": { "value": "%s" }
			},
			"identity-profiles": [
			  {
				"vid": 1,
				"identities": [
				  {
					"type": "EMAIL",
					"value": "%s"
				  },
				  {
					"type": "LEAD_GUID",
					"value": "%s"
				  }
				]
			  }
			]
		  }`

		jsonContact := fmt.Sprintf(jsonContactModel, 1, 111, 112, 111, "lead", "123@124.com", "123-45")
		contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

		hubspotDocument := model.HubspotDocument{
			TypeAlias: model.HubspotDocumentTypeNameContact,
			Value:     &contactPJson,
		}

		status := store.GetStore().CreateHubspotDocument(project1.ID, &hubspotDocument)
		assert.Equal(t, http.StatusCreated, status)

		newSyncInfoMap := map[string]int64{
			"company": 1234,
			"deals":   123,
		}
		enNewSyncInfoMap, err := json.Marshal(newSyncInfoMap)
		assert.Nil(t, err)
		store.GetStore().UpdateProjectSettings(project1.ID, &model.ProjectSetting{IntHubspotSyncInfo: &postgres.Jsonb{enNewSyncInfoMap}})

		w := sendGetHubspotSyncInfo(r, false)
		assert.Equal(t, http.StatusFound, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		hubspotSyncInfo := model.HubspotSyncInfo{}
		err = json.Unmarshal(jsonResponse, &hubspotSyncInfo)
		assert.Nil(t, err)
		assert.Equal(t, int64(112), hubspotSyncInfo.LastSyncInfo[project1.ID]["contact"])
		assert.Equal(t, int64(1234), hubspotSyncInfo.LastSyncInfo[project1.ID]["company"]) // same as before
		assert.Equal(t, int64(123), hubspotSyncInfo.LastSyncInfo[project1.ID]["deals"])
	})
}

func TestHubspotLatestUserProperties(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	createdAt := time.Now().AddDate(0, 0, -11)
	updatedDate := createdAt.AddDate(0, 0, 1)

	email := getRandomEmail()
	contact := IntHubspot.Contact{
		Vid: int64(1),
		Properties: map[string]IntHubspot.Property{
			"createdate":       {Value: fmt.Sprintf("%d", createdAt.Unix()*1000)},
			"lastmodifieddate": {Value: fmt.Sprintf("%d", updatedDate.Unix()*1000)},
			"lifecyclestage":   {Value: "lead"},
		},
		IdentityProfiles: []IntHubspot.ContactIdentityProfile{
			{
				Identities: []IntHubspot.ContactIdentity{
					{
						Type:  "EMAIL",
						Value: email,
					},
					{
						Type:  "LEAD_GUID",
						Value: "123-45",
					},
				},
			},
		},
	}

	enJSON, err := json.Marshal(contact)
	assert.Nil(t, err)
	contactPJson := postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}
	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdAt.Unix()*1000, hubspotDocument.Timestamp)

	companyCreatedDate := createdAt.AddDate(0, 0, -1)
	companyUpdatedDate := companyCreatedDate.AddDate(0, 0, 1)
	company := IntHubspot.Company{
		CompanyId:  1,
		ContactIds: []int64{1},
		Properties: map[string]IntHubspot.Property{
			"createdate":             {Value: fmt.Sprintf("%d", companyCreatedDate.Unix()*1000)},
			"hs_lastmodifieddate":    {Value: fmt.Sprintf("%d", companyUpdatedDate.Unix()*1000)},
			"company_lifecyclestage": {Value: "lead"},
			"name": {
				Value:     "testcompany",
				Timestamp: companyCreatedDate.Unix() * 1000,
			},
		},
	}

	enJSON, err = json.Marshal(company)
	assert.Nil(t, err)
	companyPJson := postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameCompany,
		Value:     &companyPJson,
	}
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, companyCreatedDate.Unix()*1000, hubspotDocument.Timestamp)

	IntHubspot.Sync(project.ID, 3)

	query := model.Query{
		From: createdAt.Unix() - 500,
		To:   updatedDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "$hubspot_contact_created",
				Properties: []model.QueryProperty{},
			},
		},

		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:   model.PropertyEntityUser,
				Property: "$hubspot_contact_lifecyclestage",
			},
		},
	}

	result, status, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, 1, len(result.Rows))
	assert.Equal(t, "$hubspot_contact_created", result.Rows[0][0])
	assert.Equal(t, "lead", result.Rows[0][1])
	assert.Equal(t, float64(1), result.Rows[0][2])

	updatedDate = updatedDate.AddDate(0, 0, 1)
	contact = IntHubspot.Contact{
		Vid: int64(1),
		Properties: map[string]IntHubspot.Property{
			"createdate":       {Value: fmt.Sprintf("%d", createdAt.Unix()*1000)},
			"lastmodifieddate": {Value: fmt.Sprintf("%d", updatedDate.Unix()*1000)},
			"lifecyclestage":   {Value: "customer"},
		},
		IdentityProfiles: []IntHubspot.ContactIdentityProfile{
			{
				Identities: []IntHubspot.ContactIdentity{
					{
						Type:  "EMAIL",
						Value: email,
					},
					{
						Type:  "LEAD_GUID",
						Value: "123-45",
					},
				},
			},
		},
	}

	enJSON, err = json.Marshal(contact)
	assert.Nil(t, err)
	contactPJson = postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, updatedDate.Unix()*1000, hubspotDocument.Timestamp)
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	IntHubspot.Sync(project.ID, 3)

	result, status, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, 1, len(result.Rows))
	assert.Equal(t, "$hubspot_contact_created", result.Rows[0][0])
	assert.Equal(t, "customer", result.Rows[0][1])
	assert.Equal(t, float64(1), result.Rows[0][2])

}

func TestHubspotCustomerUserIDChange(t *testing.T) {

	r := gin.Default()
	H.InitDataServiceRoutes(r)
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	createdAt := time.Now().AddDate(0, 0, -11)
	createdAtStr := fmt.Sprint(createdAt.Unix() * 1000)
	email1 := getRandomEmail()

	jsonContactMap := map[string]interface{}{
		"vid":     1,
		"addedAt": createdAtStr,
		"properties": map[string]map[string]interface{}{
			"createdate":       {"value": createdAtStr},
			"lastmodifieddate": {"value": createdAtStr},
			"lifecyclestage":   {"value": "lead"},
		},
		"identity-profiles": []map[string]interface{}{
			{
				"vid": 1,
				"identities": []map[string]interface{}{
					{
						"type":  "EMAIL",
						"value": email1,
					},
					{
						"type":  "LEAD_GUID",
						"value": "123-45",
					},
				},
			},
		},
	}
	w := sendCreateHubspotDocumentRequest(project.ID, r, agent, model.HubspotDocumentTypeNameContact, &jsonContactMap)
	assert.Equal(t, http.StatusCreated, w.Code)
	email2 := getRandomEmail()
	lastModifiedAt := createdAt.AddDate(0, 0, 1)
	lastModifiedAtStr := fmt.Sprint(lastModifiedAt.Unix() * 1000)
	jsonContactMap = map[string]interface{}{
		"vid":     1,
		"addedAt": lastModifiedAtStr,
		"properties": map[string]map[string]interface{}{
			"createdate":       {"value": createdAtStr},
			"lastmodifieddate": {"value": lastModifiedAtStr},
			"lifecyclestage":   {"value": "customer"},
		},
		"identity-profiles": []map[string]interface{}{
			{
				"vid": 1,
				"identities": []map[string]interface{}{
					{
						"type":  "EMAIL",
						"value": email2,
					},
					{
						"type":  "LEAD_GUID",
						"value": "123-45",
					},
				},
			},
		},
	}
	w = sendCreateHubspotDocumentRequest(project.ID, r, agent, model.HubspotDocumentTypeNameContact, &jsonContactMap)
	assert.Equal(t, http.StatusCreated, w.Code)

	lastModifiedAt = lastModifiedAt.AddDate(0, 0, 1)
	lastModifiedAtStr = fmt.Sprint(lastModifiedAt.Unix() * 1000)
	jsonContactMap = map[string]interface{}{
		"vid":     1,
		"addedAt": lastModifiedAtStr,
		"properties": map[string]map[string]interface{}{
			"createdate":       {"value": createdAtStr},
			"lastmodifieddate": {"value": lastModifiedAtStr},
			"lifecyclestage":   {"value": "customer"},
		},
		"identity-profiles": []map[string]interface{}{
			{
				"vid": 1,
				"identities": []map[string]interface{}{
					{
						"type":  "EMAIL",
						"value": email2,
					},
					{
						"type":  "LEAD_GUID",
						"value": "123-45",
					},
				},
			},
		},
	}
	w = sendCreateHubspotDocumentRequest(project.ID, r, agent, model.HubspotDocumentTypeNameContact, &jsonContactMap)
	assert.Equal(t, http.StatusCreated, w.Code)

	smartEventRule := &model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceHubspot,
		ObjectType:           model.HubspotDocumentTypeNameContact,
		Description:          "hubspot contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "lifecyclestage",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "customer",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "customer",
						Operator:      model.COMPARE_NOT_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: model.TimestampReferenceTypeDocument,
	}

	eventNameLifecycleStageLead := "lifecyclestage_customer"
	requestPayload := make(map[string]interface{})
	requestPayload["name"] = eventNameLifecycleStageLead
	requestPayload["expr"] = smartEventRule

	w = sendCreateSmartEventFilterReq(r, project.ID, agent, &requestPayload)
	assert.Equal(t, http.StatusCreated, w.Code)

	IntHubspot.Sync(project.ID, 3)

	query := model.Query{
		From: createdAt.Unix() - 500,
		To:   lastModifiedAt.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "lifecyclestage_customer",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassInsights,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, status, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "count", result.Headers[0])
	assert.Equal(t, float64(1), result.Rows[0][0])
}

func TestHubspotParallelProcessingByDocumentID(t *testing.T) {
	/*
		generate per day time series -> {Day1,Day2}, {Day2,Day3},{Day3,Day4} upto current day
	*/
	startTimestamp := time.Now().AddDate(0, 0, -10) // 10 days excluding today
	startDate := time.Date(startTimestamp.Year(), startTimestamp.Month(), startTimestamp.Day(), 0, 0, 0, 0, time.UTC)
	expectedTimeSeries := [][]int64{}
	for i := 0; i < 11; i++ {
		expectedTimeSeries = append(expectedTimeSeries, []int64{startDate.AddDate(0, 0, i).Unix() * 1000, startDate.AddDate(0, 0, i+1).Unix() * 1000})
	}

	resultTimeSeries := model.GetCRMTimeSeriesByStartTimestamp(1, startTimestamp.Unix()*1000, model.SmartCRMEventSourceHubspot)
	assert.Equal(t, 11, len(resultTimeSeries)) // expected length 11

	for i := 0; i < 11; i++ {
		if i == 0 {
			assert.Equal(t, startTimestamp.Unix()*1000, resultTimeSeries[i][0])
		} else {
			assert.Equal(t, expectedTimeSeries[i][0], resultTimeSeries[i][0])
		}

		assert.Equal(t, expectedTimeSeries[i][1], resultTimeSeries[i][1])
	}

	/*
		Split documents to batches. Mainting order
	*/
	documents := []model.HubspotDocument{}
	for i := 0; i < 10; i++ {
		documents = append(documents, model.HubspotDocument{
			ID: fmt.Sprintf("%d", i),
		})
	}

	batchedDocuments := IntHubspot.GetBatchedOrderedDocumentsByID(documents, 4)
	for i := 0; i < 3; i++ {
		for docID := 4 * i; docID < 4*(i+1); docID++ {
			if docID > 9 {
				break
			}
			assert.NotNil(t, documents[docID].ID, batchedDocuments[i][fmt.Sprintf("%d", docID)])
		}
	}

	r := gin.Default()
	H.InitDataServiceRoutes(r)
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	intHubspot := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntHubspot: &intHubspot, IntHubspotApiKey: "1234",
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	createdAt := time.Now().AddDate(0, 0, -11)
	emails := []string{getRandomEmail(), getRandomEmail(), getRandomEmail()}
	leadGUIDS := []string{U.RandomString(5), U.RandomString(5), U.RandomString(5)}

	// created 3 unique document_id with 11 updates
	for contactID := 1; contactID < 4; contactID++ {
		lastModified := createdAt
		for i := 0; i < 10; i++ {
			jsonContactMap := map[string]interface{}{
				"vid":     contactID,
				"addedAt": U.GetPropertyValueAsString(createdAt.Unix() * 1000),
				"properties": map[string]map[string]interface{}{
					"createdate":       {"value": U.GetPropertyValueAsString(createdAt.Unix() * 1000)},
					"lastmodifieddate": {"value": U.GetPropertyValueAsString(lastModified.Unix() * 1000)},
					"lifecyclestage":   {"value": "lead"},
					"count":            {"value": U.GetPropertyValueAsString(i)},
				},
				"identity-profiles": []map[string]interface{}{
					{
						"vid": 1,
						"identities": []map[string]interface{}{
							{
								"type":  "EMAIL",
								"value": emails[contactID-1],
							},
							{
								"type":  "LEAD_GUID",
								"value": leadGUIDS[contactID-1],
							},
						},
					},
				},
			}
			w := sendCreateHubspotDocumentRequest(project.ID, r, agent, model.HubspotDocumentTypeNameContact, &jsonContactMap)
			assert.Equal(t, http.StatusCreated, w.Code)

			lastModified = lastModified.AddDate(0, 0, 1)
		}
	}

	resultTimeSeries = model.GetCRMTimeSeriesByStartTimestamp(1, createdAt.Unix()*1000, model.SmartCRMEventSourceHubspot)
	assert.Equal(t, 12, len(resultTimeSeries))

	for i := 0; i < 10; i++ {
		documents, _ := store.GetStore().GetHubspotDocumentsByTypeANDRangeForSync(project.ID, model.HubspotDocumentTypeContact, resultTimeSeries[i][0], resultTimeSeries[i][1])
		if i == 0 {
			assert.Equal(t, 6, len(documents))
		} else {
			assert.Equal(t, 3, len(documents))
		}

		var contact IntHubspot.Contact
		json.Unmarshal(documents[0].Value.RawMessage, &contact)
		assert.Equal(t, fmt.Sprintf("%d", i), contact.Properties["count"].Value)
	}

	var companyUpdatedDate time.Time
	for companyID := int64(1); companyID < 4; companyID++ {
		companyCreatedDate := createdAt
		companyUpdatedDate = companyCreatedDate
		for i := 0; i < 9; i++ {
			company := IntHubspot.Company{
				CompanyId:  companyID,
				ContactIds: []int64{companyID},
				Properties: map[string]IntHubspot.Property{
					"createdate":             {Value: fmt.Sprintf("%d", companyCreatedDate.Unix()*1000)},
					"hs_lastmodifieddate":    {Value: fmt.Sprintf("%d", companyUpdatedDate.Unix()*1000)},
					"company_lifecyclestage": {Value: "lead"},
					"name": {
						Value:     "testcompany",
						Timestamp: companyCreatedDate.Unix() * 1000,
					},
					"count": {Value: fmt.Sprintf("%d", i)},
				},
			}

			enJSON, err := json.Marshal(company)
			assert.Nil(t, err)
			companyPJson := postgres.Jsonb{json.RawMessage(enJSON)}
			hubspotDocument := model.HubspotDocument{
				TypeAlias: model.HubspotDocumentTypeNameCompany,
				Value:     &companyPJson,
			}
			status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
			assert.Equal(t, http.StatusCreated, status)
			if i == 0 {
				assert.Equal(t, companyCreatedDate.Unix()*1000, hubspotDocument.Timestamp)
			} else {
				assert.Equal(t, companyUpdatedDate.Unix()*1000, hubspotDocument.Timestamp)
			}
			companyUpdatedDate = companyUpdatedDate.AddDate(0, 0, 1)
		}

	}

	numParallelDocuments := 3
	enrichStatus, _ := IntHubspot.Sync(project.ID, numParallelDocuments)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	query := model.Query{
		From: createdAt.AddDate(0, 0, -1).Unix(),
		To:   companyUpdatedDate.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
				Properties: []model.QueryProperty{},
			},
		},

		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondEachGivenEvent,
		Class:           model.QueryClassEvents,
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				Property:       "$hubspot_contact_lastmodifieddate",
				EventName:      U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
				EventNameIndex: 1,
			},
			{
				Entity:         model.PropertyEntityUser,
				Property:       "$hubspot_company_hs_lastmodifieddate",
				EventName:      U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
				EventNameIndex: 1,
			},
		},
	}

	result, status := store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID)
	assert.Equal(t, http.StatusOK, status)

	rows := result.Results[0].Rows
	sort.Slice(rows, func(i, j int) bool {
		p1, _ := U.GetPropertyValueAsFloat64(rows[i][2])
		p2, _ := U.GetPropertyValueAsFloat64(rows[j][2])
		return p1 < p2
	})
	contactTimestamp := createdAt
	companyTimestamp := createdAt
	for i := 0; i < 10; i++ {
		if i == 0 {
			assert.Equal(t, fmt.Sprintf("%d", contactTimestamp.Unix()), rows[i][2])
			assert.Equal(t, "$none", rows[0][3])
		} else {
			assert.Equal(t, fmt.Sprintf("%d", contactTimestamp.Unix()), rows[i][2])
			assert.Equal(t, fmt.Sprintf("%d", companyTimestamp.Unix()*1000), rows[i][3])
			companyTimestamp = companyTimestamp.AddDate(0, 0, 1)
		}
		contactTimestamp = contactTimestamp.AddDate(0, 0, 1)
	}

	// query unqiue users and total events
	query = model.Query{
		From: createdAt.AddDate(0, 0, -1).Unix(),
		To:   companyUpdatedDate.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
				Properties: []model.QueryProperty{},
			},
		},

		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondEachGivenEvent,
		Class:           model.QueryClassEvents,
	}
	result, status = store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID)
	assert.Equal(t, http.StatusOK, status)
	count := 0
	for i := range result.Results[0].Rows {
		if result.Results[0].Rows[i][1] == U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED {
			assert.Equal(t, float64(30), result.Results[0].Rows[i][2])
			count++
		}
		if result.Results[0].Rows[i][1] == U.EVENT_NAME_HUBSPOT_CONTACT_CREATED {
			assert.Equal(t, float64(3), result.Results[0].Rows[i][2])
			count++
		}
	}
	assert.Equal(t, 2, count)

	query = model.Query{
		From: createdAt.AddDate(0, 0, -1).Unix(),
		To:   companyUpdatedDate.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
				Properties: []model.QueryProperty{},
			},
		},

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
		Class:           model.QueryClassEvents,
	}
	result, status = store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID)
	assert.Equal(t, http.StatusOK, status)
	count = 0
	for i := range result.Results[0].Rows {
		if result.Results[0].Rows[i][1] == U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED {
			assert.Equal(t, float64(3), result.Results[0].Rows[i][2])
			count++
		}
		if result.Results[0].Rows[i][1] == U.EVENT_NAME_HUBSPOT_CONTACT_CREATED {
			assert.Equal(t, float64(3), result.Results[0].Rows[i][2])
			count++
		}
	}
	assert.Equal(t, 2, count)
}

func TestGetHubspotContactCreatedSyncIDAndUserID(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)
	H.InitAppRoutes(r)
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	intHubspot := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntHubspot: &intHubspot, IntHubspotApiKey: "1234",
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	createdAt := time.Now().AddDate(0, 0, -11)
	updatedDate := createdAt.AddDate(0, 0, 1)
	cuid := getRandomEmail()
	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	jsonContact := fmt.Sprintf(jsonContactModel, 1, updatedDate.Unix(), createdAt.Unix(), updatedDate.Unix(), "lead", cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdAt.Unix(), hubspotDocument.Timestamp)
	eventID := "123-45"
	userID1 := "456-12"
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, eventID, hubspotDocument.Timestamp, model.HubspotDocumentActionCreated, userID1)
	assert.Equal(t, http.StatusAccepted, status)

	documents, status := store.GetStore().GetHubspotContactCreatedSyncIDAndUserID(project.ID, hubspotDocument.ID)
	assert.Equal(t, eventID, documents[0].SyncId)
	assert.Equal(t, userID1, documents[0].UserId)
	assert.Equal(t, createdAt.Unix(), documents[0].Timestamp)
	assert.Equal(t, 0, documents[0].Action)
	assert.Nil(t, documents[0].Value)
	assert.Equal(t, 0, documents[0].Type)
	assert.Equal(t, "", documents[0].ID)

	document, status := store.GetStore().GetLastSyncedHubspotDocumentByID(project.ID, "1", model.HubspotDocumentTypeContact)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, "1", document.ID)
	assert.Equal(t, createdAt.Unix(), document.Timestamp)
	assert.Equal(t, model.HubspotDocumentActionCreated, document.Action)
	assert.Equal(t, model.HubspotDocumentTypeContact, document.Type)

	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, eventID, updatedDate.Unix(), model.HubspotDocumentActionUpdated, userID1)
	assert.Equal(t, http.StatusAccepted, status)

	document, status = store.GetStore().GetLastSyncedHubspotDocumentByID(project.ID, "1", model.HubspotDocumentTypeContact)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, "1", document.ID)
	assert.Equal(t, updatedDate.Unix(), document.Timestamp)
	assert.Equal(t, model.HubspotDocumentActionUpdated, model.HubspotDocumentTypeContact)

	document, status = store.GetStore().GetLastSyncedHubspotDocumentByID(project.ID, "2", model.HubspotDocumentTypeContact)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestHubspotCompanyGroups(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)
	H.InitAppRoutes(r)
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	intHubspot := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntHubspot: &intHubspot, IntHubspotApiKey: "1234",
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	// company with 4 contacts
	companyID := int64(5)
	companyContact := []int64{1, 2, 3, 4}
	companyCreatedDate := time.Now().AddDate(0, 0, -5)
	companyUpdatedDate := companyCreatedDate.AddDate(0, 0, 1)
	company := IntHubspot.Company{
		CompanyId:  companyID,
		ContactIds: companyContact,
		Properties: map[string]IntHubspot.Property{
			"createdate":             {Value: fmt.Sprintf("%d", companyCreatedDate.Unix()*1000)},
			"hs_lastmodifieddate":    {Value: fmt.Sprintf("%d", companyUpdatedDate.Unix()*1000)},
			"company_lifecyclestage": {Value: "lead"},
			"name": {
				Value:     "testcompany",
				Timestamp: companyCreatedDate.Unix() * 1000,
			},
			"domain": {
				Value:     "testcompany.com",
				Timestamp: companyCreatedDate.Unix() * 1000,
			},
		},
	}
	enJSON, err := json.Marshal(company)
	assert.Nil(t, err)
	companyPJson := postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameCompany,
		Value:     &companyPJson,
	}
	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// extra company creation for go routines test
	for _, companyID := range []int{1, 2, 3} {
		company.CompanyId = int64(companyID)
		company.ContactIds = nil
		enJSON, err = json.Marshal(company)
		assert.Nil(t, err)
		companyPJson = postgres.Jsonb{json.RawMessage(enJSON)}
		hubspotDocument := model.HubspotDocument{
			TypeAlias: model.HubspotDocumentTypeNameCompany,
			Value:     &companyPJson,
		}
		status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	// contacts for company
	for i := range companyContact {
		contact := IntHubspot.Contact{
			Vid: companyContact[i],
			Properties: map[string]IntHubspot.Property{
				"createdate":       {Value: fmt.Sprintf("%d", companyCreatedDate.Add(100*time.Minute).Unix()*1000)},
				"lastmodifieddate": {Value: fmt.Sprintf("%d", companyCreatedDate.Add(100*time.Minute).Unix()*1000)},
				"lifecyclestage":   {Value: "lead"},
			},
			IdentityProfiles: []IntHubspot.ContactIdentityProfile{
				{
					Identities: []IntHubspot.ContactIdentity{
						{
							Type:  "LEAD_GUID",
							Value: getRandomAgentUUID(),
						},
					},
				},
			},
		}

		enJSON, err = json.Marshal(contact)
		assert.Nil(t, err)
		contactPJson := postgres.Jsonb{json.RawMessage(enJSON)}
		hubspotDocument := model.HubspotDocument{
			TypeAlias: model.HubspotDocumentTypeNameContact,
			Value:     &contactPJson,
		}
		status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	enrichStatus, _ := IntHubspot.Sync(project.ID, 3)

	assert.Equal(t, project.ID, enrichStatus[0].ProjectId)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)

	// Verfiying contact to company association
	contactIDS := []string{}
	for i := range companyContact {
		contactIDS = append(contactIDS, fmt.Sprintf("%d", companyContact[i]))
	}
	companyIDstring := fmt.Sprintf("%d", companyID)
	companyDocuments, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{companyIDstring}, model.HubspotDocumentTypeCompany, []int{model.HubspotDocumentActionCreated})

	contactDocuments, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, contactIDS, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, contactDocuments, 4)
	for i := range contactDocuments {
		contactUser, status := store.GetStore().GetUser(project.ID, contactDocuments[i].UserId)
		assert.Equal(t, http.StatusFound, status)
		// verify group_1_id is company unique id and group_1_user_id is company user_id
		assert.Equal(t, true, assertUserGroupValueByColumnName(contactUser, "group_1_id", "testcompany"))
		assert.Equal(t, true, assertUserGroupValueByColumnName(contactUser, "group_1_user_id", companyDocuments[0].UserId))
	}

	/*
		Contact moving to different company will not be updated
	*/
	company.CompanyId = 2
	company.ContactIds = companyContact
	company.Properties = map[string]IntHubspot.Property{
		"createdate":             {Value: fmt.Sprintf("%d", companyCreatedDate.Unix()*1000)},
		"hs_lastmodifieddate":    {Value: fmt.Sprintf("%d", companyUpdatedDate.Add(100*time.Minute).Unix()*1000)},
		"company_lifecyclestage": {Value: "lead"},
		"name": {
			Value:     "testcompany",
			Timestamp: companyCreatedDate.Unix() * 1000,
		},
	}
	enJSON, err = json.Marshal(company)
	assert.Nil(t, err)
	companyPJson = postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameCompany,
		Value:     &companyPJson,
	}
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	enrichStatus, _ = IntHubspot.Sync(project.ID, 3)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectId)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)

	// verify user still associated with previous company
	for i := range contactDocuments {
		contactUser, status := store.GetStore().GetUser(project.ID, contactDocuments[i].UserId)
		assert.Equal(t, http.StatusFound, status)
		assert.Equal(t, true, assertUserGroupValueByColumnName(contactUser, "group_1_id", "testcompany"))
		assert.Equal(t, true, assertUserGroupValueByColumnName(contactUser, "group_1_user_id", companyDocuments[0].UserId))
	}

	// total company events
	query := model.Query{
		From: companyCreatedDate.Unix() - 500,
		To:   companyUpdatedDate.AddDate(0, 0, 1).Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				Property:       "$hubspot_company_name",
				EventName:      U.EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				EventNameIndex: 1,
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondEachGivenEvent,
	}

	result, status := store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, U.EVENT_NAME_HUBSPOT_COMPANY_CREATED, result.Results[0].Rows[0][1])
	assert.Equal(t, "testcompany", result.Results[0].Rows[0][2])
	assert.Equal(t, float64(4), result.Results[0].Rows[0][3])

	// total users
	query = model.Query{
		From: companyCreatedDate.Unix() - 500,
		To:   companyUpdatedDate.AddDate(0, 0, 1).Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}

	result, status = store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, float64(4), result.Results[0].Rows[0][2])

	/*
		Test use company domain name if company name not available
	*/

	companyID = 10
	company = IntHubspot.Company{
		CompanyId: companyID,
		Properties: map[string]IntHubspot.Property{
			"createdate":          {Value: fmt.Sprintf("%d", companyCreatedDate.Unix()*1000)},
			"hs_lastmodifieddate": {Value: fmt.Sprintf("%d", companyUpdatedDate.Unix()*1000)},
			"lifecyclestage":      {Value: "lead"},
			"name": {
				Value:     "",
				Timestamp: companyCreatedDate.Unix() * 1000,
			},
			"domain": {
				Value:     "testcompany2.com",
				Timestamp: companyCreatedDate.Unix() * 1000,
			},
		},
	}

	enJSON, err = json.Marshal(company)
	assert.Nil(t, err)
	companyPJson = postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameCompany,
		Value:     &companyPJson,
	}
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	enrichStatus, _ = IntHubspot.Sync(project.ID, 3)

	assert.Equal(t, project.ID, enrichStatus[0].ProjectId)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)

	companyDocuments, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%d", companyID)}, model.HubspotDocumentTypeCompany,
		[]int{model.HubspotDocumentActionCreated})
	user, status := store.GetStore().GetUser(project.ID, companyDocuments[0].UserId)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, "testcompany2.com", user.Group1ID)
	userProperties, err := util.DecodePostgresJsonb(&user.Properties)
	assert.Equal(t, "lead", (*userProperties)["$hubspot_company_lifecyclestage"])
}

func TestHubspotOfflineTouchPoint(t *testing.T) {

	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	_, status := store.GetStore().CreateOrGetOfflineTouchPointEventName(project.ID)
	if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
		fmt.Println("failed to create event name on SF for offline touch point")
		return
	}

	documentID := 1
	createdDate := time.Now().AddDate(0, 0, -1).Unix() * 1000
	cuid := U.RandomLowerAphaNumString(5)
	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	// document first created
	jsonContact := fmt.Sprintf(jsonContactModel, documentID, createdDate, createdDate, createdDate, "lead", cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
		Action:    model.HubspotDocumentActionUpdated,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdDate, hubspotDocument.Timestamp)
	assert.Nil(t, err)

	enProperties, _, err := IntHubspot.GetContactProperties(project.ID, &hubspotDocument)
	assert.Nil(t, err)
	(*enProperties)["$hubspot_campaign_name"] = "Webinar"

	trackPayload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: *enProperties,
		UserProperties:  *enProperties,
		Name:            U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
		Timestamp:       getEventTimestamp(hubspotDocument.Timestamp),
	}
	userID1 := U.RandomLowerAphaNumString(5)
	createdUserID, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID1, CustomerUserId: cuid, JoinTimestamp: getEventTimestamp(hubspotDocument.Timestamp)})
	assert.Equal(t, http.StatusCreated, status)

	trackPayload.UserId = createdUserID

	filter1 := model.TouchPointFilter{
		Property:  "$hubspot_campaign_name",
		Operator:  "contains",
		Value:     "Webinar",
		LogicalOp: "AND",
	}

	rulePropertyMap := make(map[string]model.HSTouchPointPropertyValue)
	rulePropertyMap["$campaign"] = model.HSTouchPointPropertyValue{Type: model.HSTouchPointPropertyValueAsProperty, Value: "$hubspot_campaign_name"}
	rulePropertyMap["$channel"] = model.HSTouchPointPropertyValue{Type: model.HSTouchPointPropertyValueAsConstant, Value: "Other"}

	rule := model.HSTouchPointRule{
		Filters:           []model.TouchPointFilter{filter1},
		TouchPointTimeRef: model.LastModifiedTimeRef,
		PropertiesMap:     rulePropertyMap,
	}

	var defaultSmartEventTimestamp int64
	if timestamp, err := model.GetHubspotDocumentUpdatedTimestamp(&hubspotDocument); err != nil {
		defaultSmartEventTimestamp = hubspotDocument.Timestamp
	} else {
		defaultSmartEventTimestamp = timestamp
	}

	trackResponse, err := IntHubspot.CreateTouchPointEvent(project, trackPayload, &hubspotDocument, rule, defaultSmartEventTimestamp)
	assert.Nil(t, err)
	assert.NotNil(t, trackResponse)

	event, errCode := store.GetStore().GetEventById(project.ID, trackResponse.EventId, trackResponse.UserId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, event)
	eventPropertiesBytes, err := event.Properties.Value()
	var eventPropertiesMap map[string]interface{}
	_ = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
	assert.Equal(t, eventPropertiesMap["$campaign"], "Webinar")
}

func TestHubspotOfflineTouchPointDecode(t *testing.T) {

	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	filter1 := model.TouchPointFilter{
		Property:  "$hubspot_campaign_name",
		Operator:  "contains",
		Value:     "Webinar",
		LogicalOp: "AND",
	}

	rulePropertyMap := make(map[string]model.HSTouchPointPropertyValue)
	rulePropertyMap["$campaign"] = model.HSTouchPointPropertyValue{Type: model.HSTouchPointPropertyValueAsProperty, Value: "$hubspot_campaign_type"}
	rulePropertyMap["$channel"] = model.HSTouchPointPropertyValue{Type: model.HSTouchPointPropertyValueAsConstant, Value: "Other"}

	rule := model.HSTouchPointRule{
		Filters:           []model.TouchPointFilter{filter1},
		TouchPointTimeRef: model.LastModifiedTimeRef,
		PropertiesMap:     rulePropertyMap,
	}

	// creating manual rule
	touchPointRules := make(map[string][]model.HSTouchPointRule)
	touchPointRules["hs_touch_point_rules"] = []model.HSTouchPointRule{rule}

	// adding json rule
	project.HubspotTouchPoints = postgres.Jsonb{RawMessage: json.RawMessage(`{"hs_touch_point_rules":[{"filters":[{"pr":"$hubspot_campaign_type","op":"equals","va":"Field Events","lop":"AND"},{"pr":"$hubspot_campaign_name","op":"contains","va":"Sendoso","lop":"AND"}],"touch_point_time_ref":"LAST_MODIFIED_TIME_REF","properties_map":{"$campaign_name":{"type":"property","value":"$hubspot_campaign_name"},"$source":{"type":"constant","value":"Source1"},"$channel":{"type":"constant","value":"CRM"},"$type":{"type":"constant","value":"Offer"}}}]}`)}
	store.GetStore().UpdateProject(project.ID, project)

	project, errCode := store.GetStore().GetProject(project.ID)
	if errCode != http.StatusFound {
		return
	}
	if &project.HubspotTouchPoints != nil && !U.IsEmptyPostgresJsonb(&project.HubspotTouchPoints) {

		var touchPointRules map[string][]model.HSTouchPointRule
		err := U.DecodePostgresJsonbToStructType(&project.HubspotTouchPoints, &touchPointRules)
		assert.Nil(t, err)

		rules := touchPointRules["hs_touch_point_rules"]

		assert.Equal(t, len(rules), 1)
	}
}

func getEventTimestamp(timestamp int64) int64 {
	if timestamp == 0 {
		return 0
	}
	return timestamp / 1000
}
