package tests

import (
	"encoding/json"
	IntHubspot "factors/integration/Hubspot"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
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
	_, status := M.CreateUser(&M.User{ProjectId: project.ID, ID: userID1, CustomerUserId: cuid})
	assert.Equal(t, http.StatusCreated, status)
	_, status = M.CreateUser(&M.User{ProjectId: project.ID, ID: userID2, CustomerUserId: cuid})
	assert.Equal(t, http.StatusCreated, status)
	_, status = M.CreateUser(&M.User{ProjectId: project.ID, ID: userID3, CustomerUserId: cuid})
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

	jsonContact := fmt.Sprintf(jsonContactModel, 1, createdAt.Unix(), createdAt.Unix(), updatedDate.Unix(), "lead", cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := M.HubspotDocument{
		TypeAlias: M.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = M.CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = M.UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID1)
	assert.Equal(t, http.StatusAccepted, status)

	// updated to opportunity
	updatedDate = createdAt.AddDate(0, 0, 1)
	jsonContact = fmt.Sprintf(jsonContactModel, 1, createdAt.Unix(), createdAt.Unix(), updatedDate.Unix(), "lead", "test@gmail.com", "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = M.HubspotDocument{
		TypeAlias: M.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = M.CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	filter := M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceHubspot,
		ObjectType:           "contact",
		Description:          "hubspot contact lifecyclestage",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "lifecyclestage",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         "opportunity",
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         "lead",
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
	currentProperties["lifecyclestage"] = "opportunity"
	smartEvent, _, ok := IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", cuid, userID3, hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)
	assert.Equal(t, "test", smartEvent.Name)
	assert.Equal(t, "lead", smartEvent.Properties["$prev_hubspot_contact_lifecyclestage"])
	assert.Equal(t, "opportunity", smartEvent.Properties["$curr_hubspot_contact_lifecyclestage"])

	// updated last synced to customer
	updatedDate = createdAt.AddDate(0, 0, 2)
	jsonContact = fmt.Sprintf(jsonContactModel, 1, createdAt.Unix(), createdAt.Unix(), updatedDate.Unix(), "customer", cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = M.HubspotDocument{
		TypeAlias: M.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = M.CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = M.UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID1)
	assert.Equal(t, http.StatusAccepted, status)

	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", cuid, userID1, hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, false, ok)

	// updated last synced to lead with different user_id having same customer_user_id
	updatedDate = createdAt.AddDate(0, 0, 3)
	jsonContact = fmt.Sprintf(jsonContactModel, 1, createdAt.Unix(), createdAt.Unix(), updatedDate.Unix(), "lead", cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = M.HubspotDocument{
		TypeAlias: M.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = M.CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = M.UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID3)
	assert.Equal(t, http.StatusAccepted, status)

	currentProperties = make(map[string]interface{})
	currentProperties["lifecyclestage"] = "opportunity"
	smartEvent, _, ok = IntHubspot.GetHubspotSmartEventPayload(project.ID, "test", cuid, userID1, hubspotDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)

	// use empty records if no previous record exist
	filter = M.SmartCRMEventFilter{
		Source:               M.SmartCRMEventSourceHubspot,
		ObjectType:           "contact",
		Description:          "hubspot booked",
		FilterEvaluationType: M.FilterEvaluationTypeSpecific,
		Filters: []M.PropertyFilter{
			{
				Name: "lifecyclestage",
				Rules: []M.CRMFilterRule{
					{
						PropertyState: M.CurrentState,
						Value:         M.PROPERTY_VALUE_ANY,
						Operator:      M.COMPARE_EQUAL,
					},
					{
						PropertyState: M.PreviousState,
						Value:         M.PROPERTY_VALUE_ANY,
						Operator:      M.COMPARE_NOT_EQUAL,
					},
				},
				LogicalOp: M.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               M.LOGICAL_OP_AND,
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
	_, errCode := M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntHubspot: &intHubspot, IntHubspotApiKey: "1234",
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	cuID := U.RandomLowerAphaNumString(5) + "@exm.com"
	firstPropTimestamp := time.Now().Unix()
	user, status := M.CreateUser(&M.User{
		ProjectId:      project.ID,
		JoinTimestamp:  firstPropTimestamp,
		CustomerUserId: cuID,
	})
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, user)

	properties := &postgres.Jsonb{RawMessage: []byte(`{"name":"user1","city":"bangalore"}`)}
	_, status = M.UpdateUserProperties(project.ID, user.ID, properties, firstPropTimestamp)
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

	hubspotDocument := M.HubspotDocument{
		TypeAlias: M.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = M.CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	//enrich job, create contact created and contact updated event
	enrichStatus := IntHubspot.Sync(project.ID)
	projectIndex := -1
	for i := range enrichStatus {
		if enrichStatus[i].ProjectId == project.ID {
			projectIndex = i
			break
		}
	}
	assert.Equal(t, project.ID, enrichStatus[projectIndex].ProjectId)
	assert.Equal(t, "success", enrichStatus[projectIndex].Status)

	query := M.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []M.QueryEventWithProperties{
			{
				Name:       "$hubspot_contact_created",
				Properties: []M.QueryProperty{},
			},
		},
		Class: M.QueryClassFunnel,
		GroupByProperties: []M.QueryGroupByProperty{
			{
				Entity:         M.PropertyEntityUser,
				Property:       "city",
				EventName:      "$hubspot_contact_created",
				EventNameIndex: 1,
			},
		},

		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAnyGivenEvent,
	}

	result, status, _ := M.Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "city", result.Headers[0])
	assert.Equal(t, "bangalore", result.Rows[1][0])
	assert.Equal(t, int64(1), result.Rows[1][1])

	query = M.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []M.QueryEventWithProperties{
			{
				Name:       "$hubspot_contact_created",
				Properties: []M.QueryProperty{},
			},
		},
		Class: M.QueryClassFunnel,
		GroupByProperties: []M.QueryGroupByProperty{
			{
				Entity:         M.PropertyEntityUser,
				Property:       "$user_id",
				EventName:      "$hubspot_contact_created",
				EventNameIndex: 1,
			},
		},

		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}

	result, status, _ = M.Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, cuID, result.Rows[1][0])
	assert.Equal(t, int64(1), result.Rows[1][1])
}
