package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	IntHubspot "factors/integration/Hubspot"
	"factors/model/model"
	"factors/model/store"
	"factors/task/event_user_cache"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID1)
	assert.Equal(t, http.StatusAccepted, status)

	// updated to opportunity
	updatedDate = createdAt.AddDate(0, 0, 1)
	jsonContact = fmt.Sprintf(jsonContactModel, 1, updatedDate.Unix(), createdAt.Unix(), updatedDate.Unix(), "lead", "test@gmail.com", "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, updatedDate.Unix(), hubspotDocument.Timestamp)

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
	smartEvent, _, ok := IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", cuid, userID3, hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)
	assert.Equal(t, "test", smartEvent.Name)
	assert.Equal(t, "lead", smartEvent.Properties["$prev_hubspot_contact_lifecyclestage"])
	assert.Equal(t, "opportunity", smartEvent.Properties["$curr_hubspot_contact_lifecyclestage"])

	// updated last synced to customer
	updatedDate = createdAt.AddDate(0, 0, 2)
	jsonContact = fmt.Sprintf(jsonContactModel, 1, updatedDate.Unix(), createdAt.Unix(), updatedDate.Unix(), "customer", cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, updatedDate.Unix(), hubspotDocument.Timestamp)
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID1)
	assert.Equal(t, http.StatusAccepted, status)

	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", cuid, userID1, hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, false, ok)

	// updated last synced to lead with different user_id having same customer_user_id
	updatedDate = createdAt.AddDate(0, 0, 3)
	jsonContact = fmt.Sprintf(jsonContactModel, 1, updatedDate.Unix(), createdAt.Unix(), updatedDate.Unix(), "lead", cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, updatedDate.Unix(), hubspotDocument.Timestamp)
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID3)
	assert.Equal(t, http.StatusAccepted, status)

	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", cuid, userID1, hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)

	// use empty records if no previous record exist
	filter = model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceHubspot,
		ObjectType:           "contact",
		Description:          "hubspot booked",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "lifecyclestage",
				Rules: []model.CRMFilterRule{
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
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	cuid = "123-456-789" // new customer user id
	userID4 := "1230234" // new user id no previous record
	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity1"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", cuid, userID4, hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)

	// if property value nil
	PrevProperties := make(map[string]interface{})
	PrevProperties["lifecyclestage"] = nil
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", cuid, userID4, hubspotDocument.Type, &currentProperties, &PrevProperties, &filter)
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
	user, status := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		JoinTimestamp:  firstPropTimestamp,
		CustomerUserId: cuID,
	})
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, user)

	properties := &postgres.Jsonb{RawMessage: []byte(`{"name":"user1","city":"bangalore"}`)}
	_, status = store.GetStore().UpdateUserProperties(project.ID, user.ID, properties, firstPropTimestamp)
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
	enrichStatus, _ := IntHubspot.Sync(project.ID)
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
	assert.Equal(t, int64(1), result.Rows[1][1])

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
	assert.Equal(t, int64(1), result.Rows[1][1])
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
		for j := 0; j < i+1; j++ {
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

	allStatus, _ := IntHubspot.Sync(project.ID)
	for i := range allStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, allStatus[i].Status)
	}

	eventsLimit, propertyLimit, valueLimit, rollBackWindow := 1000, 10000, 10000, 1
	event_user_cache.DoRollUpAndCleanUp(&eventsLimit, &propertyLimit, &valueLimit, &rollBackWindow)
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
	enrichStatus, _ := IntHubspot.Sync(project.ID)
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
		assert.Equal(t, int64(1), result.Results[i].Rows[0][2])
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
	assert.Equal(t, int64(1), result.Results[0].Rows[0][0])
}
