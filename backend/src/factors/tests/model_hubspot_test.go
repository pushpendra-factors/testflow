package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	IntHubspot "factors/integration/hubspot"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	"factors/task/event_user_cache"
	"factors/task/hubspot_enrich"
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

func TestHubspotEngagements(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	jsonContactModelMeetings := `{
			"engagement": {
				"id": 49861280153,
				"portalId": 5928728,
				"active": true,
				"createdAt": 1579771558604,
				"lastUpdated": 1626648055847,
				"createdBy": 9765292,
				"modifiedBy": 9765292,
				"ownerId": 42479827,
				"type": "MEETING",
				"uid": "s16eeoebshn9mda18kdba4tt010",
				"timestamp": 1579837500000,
				"teamId": "3a81141",
				"allAccessibleTeamIds": [381141],
				"queueMembershipIds": [],
				"bodyPreviewIsTruncated": true,
				"gdprDeleted": false,
				"source": "engage",
				"active": "true"
			},
			"associations": {
				"contactIds": [54051],
				"companyIds": [],
				"dealIds": [],
				"ownerIds": [],
				"workflowIds": [],
				"ticketIds": [],
				"contentIds": [],
				"quoteIds": []
			},
			"attachments": [],
			"scheduledTasks": [{
				"engagementId": 4986280153,
				"portalId": 5928728,
				"engagementType": "MEETING",
				"taskType": "PRE_MEETING_NOTIFICATION",
				"timestamp": 1579835700000,
				"uuid": "MEETING:8e1628saa68-d93c-41ff-9c02-2a17659e987f"
			}],
			"metadata": {
				"startTime": 1579837500000,
				"endTime": 1579838400000,
				"title": "abc",
				"source": "MEETINGS_PUBLIC",
				"sourceId": "s16eeoebhasdn9mda18kdba4tt010",
				"createdFromLinkId": 852169,
				"preMeetingProspectReminders": [],
				"attendeeOwnerIds": [],
				"meetingOutcome": "nope"
			}
	}
	`
	jsonContactModelCalls := `{
			  "engagement": {
				"id": 4709059,
				"portalId": 62515,
				"active": true,
				"createdAt": 1428586724779,
				"lastUpdated": 1428586724779,
				"createdBy": 215482,
				"modifiedBy": 215482,
				"ownerId": 70,
				"type": "CALL",
				"timestamp": 1428565020000,
				"source": "engage",
				"activityType": "calls"
			  },
			  "associations": {
				"contactIds": [
				  54051
				],
				"companyIds": [
				  8347
				],
				"dealIds": [
				  
				],
				"ownerIds": [
				  
				],
				"workflowIds": [
				  
				]
			  },
			  "attachments": [
				
			  ],
			  "metadata": {
				"durationMilliseconds": 24000,
				"body": "test call",
				"disposition": "decent",
				"status": "ok",
				"title": "call"
			  }
	}`

	jsonContactEmail := `{
		"engagement":{
			"id":12134,
			"createdAt":1428586724780,
			"lastUpdated":1428586724781,
			"type":"EMAIL",
			"teamId":"",
			"ownerId":"",
			"active":"",
			"timestamp":1322343000,
			"source":""
		},
		"metadata":{
			"from":{
				"email":"abcd@xyz.com",
				"contactId":54051
			},
			"to":[
				{
					"email":"abcd1@xyz.com",
					"contactId":54051
				}
			],
			"subject":"",
			"sentVia":""
		}
	}`

	jsonContactIncomingEmail := `{
		"engagement":{
			"id":12135,
			"createdAt":1428586724780,
			"lastUpdated":1428586724781,
			"type": "INCOMING_EMAIL",
			"teamId":"",
			"ownerId":"",
			"active":"",
			"timestamp":1322343000,
			"source":""
		},
		"metadata":{
			"from":{
				"email":"abcd@xyz.com",
				"contactId":54051
			},
			"to":[
				{
					"email":"abcd1@xyz.com",
					"contactId":54051
				}
			],
			"subject":"",
			"sentVia":""
		}
	}`

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

	jsonContactModelMeetingsWithOutContactId := `{
		"engagement": {
			"id": 49861280154,
			"portalId": 5928728,
			"active": true,
			"createdAt": 1579771558604,
			"lastUpdated": 1626648055847,
			"createdBy": 9765292,
			"modifiedBy": 9765292,
			"ownerId": 42479827,
			"type": "MEETING",
			"uid": "s16eeoebshn9mda18kdba4tt010",
			"timestamp": 1579837500000,
			"teamId": "3a81141",
			"allAccessibleTeamIds": [381141],
			"queueMembershipIds": [],
			"bodyPreviewIsTruncated": true,
			"gdprDeleted": false,
			"source": "engage",
			"active": "true"
		},
		"associations": {
			"contactIds": [],
			"companyIds": [],
			"dealIds": [],
			"ownerIds": [],
			"workflowIds": [],
			"ticketIds": [],
			"contentIds": [],
			"quoteIds": []
		},
		"attachments": [],
		"scheduledTasks": [{
			"engagementId": 4986280153,
			"portalId": 5928728,
			"engagementType": "MEETING",
			"taskType": "PRE_MEETING_NOTIFICATION",
			"timestamp": 1579835700000,
			"uuid": "MEETING:8e1628saa68-d93c-41ff-9c02-2a17659e987f"
		}],
		"metadata": {
			"startTime": 1579837500000,
			"endTime": 1579838400000,
			"title": "abc",
			"source": "MEETINGS_PUBLIC",
			"sourceId": "s16eeoebhasdn9mda18kdba4tt010",
			"createdFromLinkId": 852169,
			"preMeetingProspectReminders": [],
			"attendeeOwnerIds": [],
			"meetingOutcome": "nope"
		}
}
`
	contactPJsonMeetingsWithOutContactId := postgres.Jsonb{json.RawMessage(jsonContactModelMeetingsWithOutContactId)}
	hubspotDocumentMeetingsWithOutContactId := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameEngagement,
		Value:     &contactPJsonMeetingsWithOutContactId,
	}
	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocumentMeetingsWithOutContactId)
	assert.Equal(t, http.StatusCreated, status)

	jsonContact := fmt.Sprintf(jsonContactModel, 54051, 1428586724779, 1428586724779, 1428586724779, "lead", "a", "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	contactPJsonMeetings := postgres.Jsonb{json.RawMessage(jsonContactModelMeetings)}
	contactPJsonCalls := postgres.Jsonb{json.RawMessage(jsonContactModelCalls)}
	contactPJsonEmail := postgres.Jsonb{json.RawMessage(jsonContactEmail)}
	contactPJsonIncomingEmail := postgres.Jsonb{json.RawMessage(jsonContactIncomingEmail)}

	hubspotDocumentMeetings := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameEngagement,
		Value:     &contactPJsonMeetings,
	}
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocumentMeetings)
	assert.Equal(t, http.StatusCreated, status)

	hubspotDocumentCalls := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameEngagement,
		Value:     &contactPJsonCalls,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocumentCalls)
	assert.Equal(t, http.StatusCreated, status)

	hubspotDocumentEmail := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameEngagement,
		Value:     &contactPJsonEmail,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocumentEmail)
	assert.Equal(t, http.StatusCreated, status)

	hubspotDocumentIncomingEmail := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameEngagement,
		Value:     &contactPJsonIncomingEmail,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocumentIncomingEmail)
	assert.Equal(t, http.StatusCreated, status)

	enrichStatus, _ := IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	docMeetingsWithoutContactId, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"49861280154"}, model.HubspotDocumentTypeEngagement, []int{model.HubspotDocumentActionCreated})
	for _, document := range docMeetingsWithoutContactId {
		assert.Equal(t, document.Synced, true)
	}

	docMeetings, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"54051"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	eventNameObjMeetingCreated, status := store.GetStore().GetEventName(U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED, project.ID)
	eventsMeetingsCreated, status := store.GetStore().GetUserEventsByEventNameId(project.ID, docMeetings[0].UserId, eventNameObjMeetingCreated.ID)
	assert.Len(t, eventsMeetingsCreated, 1)
	eventNameObjMeetingUpdated, status := store.GetStore().GetEventName(U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED, project.ID)
	eventsMeetingsUpdated, status := store.GetStore().GetUserEventsByEventNameId(project.ID, docMeetings[0].UserId, eventNameObjMeetingUpdated.ID)
	assert.Len(t, eventsMeetingsUpdated, 2)

	docCalls, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"54051"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	eventNameObjCallCreated, status := store.GetStore().GetEventName(U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED, project.ID)
	eventsCallCreated, status := store.GetStore().GetUserEventsByEventNameId(project.ID, docCalls[0].UserId, eventNameObjCallCreated.ID)
	assert.Len(t, eventsCallCreated, 1)
	eventNameObjCallUpdated, status := store.GetStore().GetEventName(U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED, project.ID)
	eventsCallUpdated, status := store.GetStore().GetUserEventsByEventNameId(project.ID, docCalls[0].UserId, eventNameObjCallUpdated.ID)
	assert.Len(t, eventsCallUpdated, 1)

	docEmail, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"54051"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	eventNameObjEmail, status := store.GetStore().GetEventName(U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL, project.ID)
	eventsEmail, status := store.GetStore().GetUserEventsByEventNameId(project.ID, docEmail[0].UserId, eventNameObjEmail.ID)
	assert.Len(t, eventsEmail, 2)

	propertyValuesMeetingCreated := make(map[string]interface{})
	err = json.Unmarshal(eventsMeetingsCreated[0].Properties.RawMessage, &propertyValuesMeetingCreated)
	assert.Nil(t, err)
	assert.Equal(t, "49861280153", propertyValuesMeetingCreated["$hubspot_engagement_id"])
	assert.Equal(t, float64(1579837500), propertyValuesMeetingCreated["$hubspot_engagement_timestamp"])
	assert.Equal(t, "MEETING", propertyValuesMeetingCreated["$hubspot_engagement_type"])
	assert.Equal(t, "engage", propertyValuesMeetingCreated["$hubspot_engagement_source"])
	assert.Equal(t, "true", propertyValuesMeetingCreated["$hubspot_engagement_active"])
	assert.Equal(t, float64(1579837500), propertyValuesMeetingCreated["$hubspot_engagement_starttime"])
	assert.Equal(t, float64(1579838400), propertyValuesMeetingCreated["$hubspot_engagement_endtime"])
	assert.Equal(t, "abc", propertyValuesMeetingCreated["$hubspot_engagement_title"])
	assert.Equal(t, "nope", propertyValuesMeetingCreated["$hubspot_engagement_meetingoutcome"])

	propertyValuesMeetingUpdatedFirst := make(map[string]interface{})
	err = json.Unmarshal(eventsMeetingsUpdated[0].Properties.RawMessage, &propertyValuesMeetingUpdatedFirst)
	assert.Nil(t, err)
	assert.Equal(t, "49861280153", propertyValuesMeetingUpdatedFirst["$hubspot_engagement_id"])
	assert.Equal(t, float64(1579837500), propertyValuesMeetingUpdatedFirst["$hubspot_engagement_timestamp"])
	assert.Equal(t, "MEETING", propertyValuesMeetingUpdatedFirst["$hubspot_engagement_type"])
	assert.Equal(t, "engage", propertyValuesMeetingUpdatedFirst["$hubspot_engagement_source"])
	assert.Equal(t, "true", propertyValuesMeetingUpdatedFirst["$hubspot_engagement_active"])
	assert.Equal(t, float64(1579837500), propertyValuesMeetingUpdatedFirst["$hubspot_engagement_starttime"])
	assert.Equal(t, float64(1579838400), propertyValuesMeetingUpdatedFirst["$hubspot_engagement_endtime"])
	assert.Equal(t, "abc", propertyValuesMeetingUpdatedFirst["$hubspot_engagement_title"])
	assert.Equal(t, "nope", propertyValuesMeetingUpdatedFirst["$hubspot_engagement_meetingoutcome"])

	propertyValuesMeetingUpdatedSecond := make(map[string]interface{})
	err = json.Unmarshal(eventsMeetingsUpdated[1].Properties.RawMessage, &propertyValuesMeetingUpdatedSecond)
	assert.Nil(t, err)
	assert.Equal(t, "49861280153", propertyValuesMeetingUpdatedSecond["$hubspot_engagement_id"])
	assert.Equal(t, float64(1579837500), propertyValuesMeetingUpdatedSecond["$hubspot_engagement_timestamp"])
	assert.Equal(t, "MEETING", propertyValuesMeetingUpdatedSecond["$hubspot_engagement_type"])
	assert.Equal(t, "engage", propertyValuesMeetingUpdatedSecond["$hubspot_engagement_source"])
	assert.Equal(t, "true", propertyValuesMeetingUpdatedSecond["$hubspot_engagement_active"])
	assert.Equal(t, float64(1579837500), propertyValuesMeetingUpdatedSecond["$hubspot_engagement_starttime"])
	assert.Equal(t, float64(1579838400), propertyValuesMeetingUpdatedSecond["$hubspot_engagement_endtime"])
	assert.Equal(t, "abc", propertyValuesMeetingUpdatedSecond["$hubspot_engagement_title"])
	assert.Equal(t, "nope", propertyValuesMeetingUpdatedSecond["$hubspot_engagement_meetingoutcome"])

	propertyValuesCallsCreated := make(map[string]interface{})
	err = json.Unmarshal(eventsCallCreated[0].Properties.RawMessage, &propertyValuesCallsCreated)
	assert.Nil(t, err)
	assert.Equal(t, "4709059", propertyValuesCallsCreated["$hubspot_engagement_id"])
	assert.Equal(t, float64(1428565020), propertyValuesCallsCreated["$hubspot_engagement_timestamp"])
	assert.Equal(t, "CALL", propertyValuesCallsCreated["$hubspot_engagement_type"])
	assert.Equal(t, "engage", propertyValuesCallsCreated["$hubspot_engagement_source"])
	assert.Equal(t, "calls", propertyValuesCallsCreated["$hubspot_engagement_activitytype"])
	assert.Equal(t, float64(24000), propertyValuesCallsCreated["$hubspot_engagement_durationmilliseconds"])
	assert.Equal(t, "decent", propertyValuesCallsCreated["$hubspot_engagement_disposition"])
	assert.Equal(t, "ok", propertyValuesCallsCreated["$hubspot_engagement_status"])
	assert.Equal(t, "call", propertyValuesCallsCreated["$hubspot_engagement_title"])

	propertyValuesCallsUpdated := make(map[string]interface{})
	err = json.Unmarshal(eventsCallUpdated[0].Properties.RawMessage, &propertyValuesCallsUpdated)
	assert.Nil(t, err)
	assert.Equal(t, "4709059", propertyValuesCallsUpdated["$hubspot_engagement_id"])
	assert.Equal(t, float64(1428565020), propertyValuesCallsUpdated["$hubspot_engagement_timestamp"])
	assert.Equal(t, "CALL", propertyValuesCallsUpdated["$hubspot_engagement_type"])
	assert.Equal(t, "engage", propertyValuesCallsUpdated["$hubspot_engagement_source"])
	assert.Equal(t, "calls", propertyValuesCallsUpdated["$hubspot_engagement_activitytype"])
	assert.Equal(t, float64(24000), propertyValuesCallsUpdated["$hubspot_engagement_durationmilliseconds"])
	assert.Equal(t, "decent", propertyValuesCallsUpdated["$hubspot_engagement_disposition"])
	assert.Equal(t, "ok", propertyValuesCallsUpdated["$hubspot_engagement_status"])
	assert.Equal(t, "call", propertyValuesCallsUpdated["$hubspot_engagement_title"])

	propertyValuesEmailOne := make(map[string]interface{})
	err = json.Unmarshal(eventsEmail[0].Properties.RawMessage, &propertyValuesEmailOne)
	assert.Nil(t, err)
	if propertyValuesEmailOne["$hubspot_engagement_type"] == "EMAIL" {
		assert.Equal(t, "12134", propertyValuesEmailOne["$hubspot_engagement_id"])
		assert.Equal(t, "EMAIL", propertyValuesEmailOne["$hubspot_engagement_type"])
	} else {
		assert.Equal(t, "12135", propertyValuesEmailOne["$hubspot_engagement_id"])
		assert.Equal(t, "INCOMING_EMAIL", propertyValuesEmailOne["$hubspot_engagement_type"])
	}

	assert.Equal(t, float64(1428586724780), propertyValuesEmailOne["$hubspot_engagement_createdat"])
	assert.Equal(t, float64(1428586724781), propertyValuesEmailOne["$hubspot_engagement_lastupdated"])
	assert.Equal(t, "", propertyValuesEmailOne["$hubspot_engagement_teamid"])
	assert.Equal(t, "", propertyValuesEmailOne["$hubspot_engagement_ownerid"])
	assert.Equal(t, "", propertyValuesEmailOne["$hubspot_engagement_active"])
	assert.Equal(t, float64(1322343), propertyValuesEmailOne["$hubspot_engagement_timestamp"])
	assert.Equal(t, "", propertyValuesEmailOne["$hubspot_engagement_source"])
	assert.Equal(t, "abcd@xyz.com", propertyValuesEmailOne["$hubspot_engagement_from"])
	assert.Equal(t, "abcd1@xyz.com", propertyValuesEmailOne["$hubspot_engagement_to"])
	assert.Equal(t, "", propertyValuesEmailOne["$hubspot_engagement_subject"])
	assert.Equal(t, "", propertyValuesEmailOne["$hubspot_engagement_sentvia"])

	propertyValuesEmailTwo := make(map[string]interface{})
	err = json.Unmarshal(eventsEmail[1].Properties.RawMessage, &propertyValuesEmailTwo)
	assert.Nil(t, err)
	if propertyValuesEmailTwo["$hubspot_engagement_type"] == "INCOMING_EMAIL" {
		assert.Equal(t, "12135", propertyValuesEmailTwo["$hubspot_engagement_id"])
		assert.Equal(t, "INCOMING_EMAIL", propertyValuesEmailTwo["$hubspot_engagement_type"])
	} else {
		assert.Equal(t, "12134", propertyValuesEmailTwo["$hubspot_engagement_id"])
		assert.Equal(t, "EMAIL", propertyValuesEmailTwo["$hubspot_engagement_type"])
	}
	assert.Equal(t, float64(1428586724780), propertyValuesEmailTwo["$hubspot_engagement_createdat"])
	assert.Equal(t, float64(1428586724781), propertyValuesEmailTwo["$hubspot_engagement_lastupdated"])
	assert.Equal(t, "", propertyValuesEmailTwo["$hubspot_engagement_teamid"])
	assert.Equal(t, "", propertyValuesEmailTwo["$hubspot_engagement_ownerid"])
	assert.Equal(t, "", propertyValuesEmailTwo["$hubspot_engagement_active"])
	assert.Equal(t, float64(1322343), propertyValuesEmailTwo["$hubspot_engagement_timestamp"])
	assert.Equal(t, "", propertyValuesEmailTwo["$hubspot_engagement_source"])
	assert.Equal(t, "abcd@xyz.com", propertyValuesEmailTwo["$hubspot_engagement_from"])
	assert.Equal(t, "abcd1@xyz.com", propertyValuesEmailTwo["$hubspot_engagement_to"])
	assert.Equal(t, "", propertyValuesEmailTwo["$hubspot_engagement_subject"])
	assert.Equal(t, "", propertyValuesEmailTwo["$hubspot_engagement_sentvia"])
}
func TestHubspotContactFormSubmission(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	createdAt := time.Now().Unix() * 1000
	lastModified := createdAt + 100

	jsonContactModel := `{
		"vid": %d,
		"addedAt":1647500074000,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "lead" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "abc@xyz.com"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "123-456"
			  }
			]
		  }
		],
		"form-submissions": [
			{
			  "conversion-id": "1d379075-bc57-4d45-80d2-5004e6ad9c44",
			  "form-id": "k61337ec-9102-441d-a7af-cf9eaa2d0774",
			  "form-type": "FACEBOOK_LEAD_AD",
			  "meta-data": [
				
			  ],
			  "page-title": "LinkedIn Lead Generation Ad",
			  "page-url": "https://www.abc.com/ad/portal/500811370/leadgen/view/5371576?hsa_acc=500811370&hsa_cam=619271286&hsa_grp=175608976&hsa_ad=157523466&hsa_src=&utm_campaign=US%257CTravel%2526HospitalityWebinar%257C20thJan2021%257CInmail%257COpen&hsa_la=true&hsa_ol=false&hsa_net=linkedin&hsa_ver=3&utm_source=linkedin&utm_medium=paid",
			  "portal-id": 2361873,
			  "timestamp": 1647393874000,
			  "title": " Webinar 20th Jan 2021"
			},
			{
				"conversion-id": "2d379075-bc57-4d45-80d2-5004e6ad9c44",
				"form-id": "i61337ec-9102-441d-a7af-cf9eaa2d0774",
				"form-type": "FACEBOOK_LEAD_AD",
				"meta-data": [
				  
				],
				"page-title": "Facebook Lead Generation Ad",
				"page-url": "https://www.adb.com/ad/portal/500811370/leadgen/view/5371576?hsa_acc=500811370&hsa_cam=619271286&hsa_grp=175608976&hsa_ad=157523466&hsa_src=&utm_campaign=US%257CTravel%2526HospitalityWebinar%257C20thJan2021%257CInmail%257COpen&hsa_la=true&hsa_ol=false&hsa_net=linkedin&hsa_ver=3&utm_source=linkedin&utm_medium=paid",
				"portal-id": 2361873,
				"timestamp": 1647393874010,
				"title": " Webinar 20th Jan 2021"
		}
		]
		
	}`

	jsonContact := fmt.Sprintf(jsonContactModel, 2, createdAt, lastModified)
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}
	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	enrichStatus, _ := IntHubspot.Sync(project.ID, 1, time.Now().UTC().Unix(), nil, "", 50)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	doc, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"2"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	events, status := store.GetStore().GetHubspotFormEvents(project.ID, doc[0].UserId, []interface{}{1647393874})

	assert.Len(t, events, 2)

	propertyValues := make(map[string]interface{})
	err = json.Unmarshal(events[1].Properties.RawMessage, &propertyValues)
	assert.Nil(t, err)

	assert.Equal(t, "2d379075-bc57-4d45-80d2-5004e6ad9c44", propertyValues["$hubspot_form_submission_conversion-id"])
	assert.Equal(t, "i61337ec-9102-441d-a7af-cf9eaa2d0774", propertyValues["$hubspot_form_submission_form-id"])
	assert.Equal(t, "FACEBOOK_LEAD_AD", propertyValues["$hubspot_form_submission_form-type"])
	assert.Equal(t, "Facebook Lead Generation Ad", propertyValues["$hubspot_form_submission_page-title"])
	assert.Equal(t, (float64)(2361873), propertyValues["$hubspot_form_submission_portal-id"])
	pageURL := "https://www.adb.com/ad/portal/500811370/leadgen/view/5371576?hsa_acc=500811370&hsa_cam=619271286&hsa_grp=175608976&hsa_ad=157523466&hsa_src=&utm_campaign=US%257CTravel%2526HospitalityWebinar%257C20thJan2021%257CInmail%257COpen&hsa_la=true&hsa_ol=false&hsa_net=linkedin&hsa_ver=3&utm_source=linkedin&utm_medium=paid"
	urlParameters := IntHubspot.GetURLParameterAsMap(pageURL)
	assert.Equal(t, urlParameters["utm_source"], propertyValues["utm_source"])
	assert.Equal(t, urlParameters["utm_medium"], propertyValues["utm_medium"])
	assert.Equal(t, (float64)(1647393874), propertyValues["$hubspot_form_submission_timestamp"])
	assert.Equal(t, " Webinar 20th Jan 2021", propertyValues["$hubspot_form_submission_title"])
	assert.Equal(t, "www.adb.com/ad/portal/500811370/leadgen/view/5371576", propertyValues["$hubspot_form_submission_page-url-no-qp"])

	propertyValues = make(map[string]interface{})
	err = json.Unmarshal(events[0].Properties.RawMessage, &propertyValues)
	assert.Nil(t, err)

	assert.Equal(t, "1d379075-bc57-4d45-80d2-5004e6ad9c44", propertyValues["$hubspot_form_submission_conversion-id"])
	assert.Equal(t, "k61337ec-9102-441d-a7af-cf9eaa2d0774", propertyValues["$hubspot_form_submission_form-id"])
	assert.Equal(t, "FACEBOOK_LEAD_AD", propertyValues["$hubspot_form_submission_form-type"])
	assert.Equal(t, "LinkedIn Lead Generation Ad", propertyValues["$hubspot_form_submission_page-title"])
	pageURL = "https://www.abc.com/ad/portal/500811370/leadgen/view/5371576?hsa_acc=500811370&hsa_cam=619271286&hsa_grp=175608976&hsa_ad=157523466&hsa_src=&utm_campaign=US%257CTravel%2526HospitalityWebinar%257C20thJan2021%257CInmail%257COpen&hsa_la=true&hsa_ol=false&hsa_net=linkedin&hsa_ver=3&utm_source=linkedin&utm_medium=paid"
	urlParameters = IntHubspot.GetURLParameterAsMap(pageURL)
	assert.Equal(t, urlParameters["utm_source"], propertyValues["utm_source"])
	assert.Equal(t, urlParameters["utm_medium"], propertyValues["utm_medium"])
	assert.Equal(t, (float64)(2361873), propertyValues["$hubspot_form_submission_portal-id"])
	assert.Equal(t, (float64)(1647393874), propertyValues["$hubspot_form_submission_timestamp"])
	assert.Equal(t, " Webinar 20th Jan 2021", propertyValues["$hubspot_form_submission_title"])
	assert.Equal(t, "www.abc.com/ad/portal/500811370/leadgen/view/5371576", propertyValues["$hubspot_form_submission_page-url-no-qp"])

	decodedString := IntHubspot.GetDecodedValue("Danny%2520%2526%2520Co", 2)
	assert.Equal(t, "Danny & Co", decodedString)
	decodedString = IntHubspot.GetDecodedValue("Danny%2520%2526%2520Co", 3)
	assert.Equal(t, "Danny & Co", decodedString)
}

func TestHubspotCRMSmartEvent(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	// contactID := U.RandomLowerAphaNumString(5)
	userID1 := U.RandomLowerAphaNumString(5)
	userID2 := U.RandomLowerAphaNumString(5)
	userID3 := U.RandomLowerAphaNumString(5)
	cuid := U.RandomLowerAphaNumString(5)
	_, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID1, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID2, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID3, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
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
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID1, "")
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

	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID2, "")
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
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID1, "")
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
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID3, "")
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
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID3, "")
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
		Source:         model.GetRequestSourcePointer(model.UserSourceHubspot),
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
	enrichStatus, _ := IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
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

	result, status, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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

	result, status, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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

	documents, status := store.GetStore().GetHubspotDocumentsByTypeForSync(project.ID, model.HubspotDocumentTypeContact, time.Now().Unix())
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 3, len(documents))

	// try reinserting the same record
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusConflict, status)

	documents, status = store.GetStore().GetHubspotDocumentsByTypeForSync(project.ID, model.HubspotDocumentTypeContact, time.Now().Unix())
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
	get_lastmodifieddate := tempJson[U.PROPERTY_KEY_LAST_MODIFIED_DATE].(string)
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

	project.ID = int64(rand.Intn(10000))
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
	get_lastmodifieddate := tempJson[U.PROPERTY_KEY_LAST_MODIFIED_DATE].(string)
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
	enrichStatus, _ := IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
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

	allStatus, _ := IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
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

	result, status, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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

func sendCreateHubspotDocumentRequest(projectID int64, r *gin.Engine, agent *model.Agent, documentType string, documentValue *map[string]interface{}) *httptest.ResponseRecorder {
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
	enrichStatus, _ := IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
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

	result, status := store.GetStore().RunEventsGroupQuery(query, project.ID, C.EnableOptimisedFilterOnEventUserQuery())
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

	result, status = store.GetStore().RunEventsGroupQuery(query, project.ID, C.EnableOptimisedFilterOnEventUserQuery())
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
	enrichStatus, _ := IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectId)
	assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[0].Status)
	assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[1].Status)
	assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[2].Status)

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
				Property: U.EP_TIMESTAMP,
			},
		},
	}

	result, status, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, 3, len(result.Rows))
	eventNameTimestamp := make(map[string]int64)
	for i := range result.Rows {
		timestamp, _ := U.GetPropertyValueAsFloat64(result.Rows[i][1])
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

	IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)

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

	result, status, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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

	IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)

	result, status, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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

	IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)

	query := model.Query{
		From: createdAt.Unix() - 500,
		To:   lastModifiedAt.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "lifecyclestage_customer",
				Properties: []model.QueryProperty{},
			},
		},
		Class:             model.QueryClassInsights,
		Type:              model.QueryTypeEventsOccurrence,
		EventsCondition:   model.EventCondAnyGivenEvent,
		AggregateFunction: model.DefaultAggrFunc,
	}

	result, status, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "aggregate", result.Headers[0])
	assert.Equal(t, float64(1), result.Rows[0][0])
}

func TestHubspotParallelProcessingByDocumentID(t *testing.T) {
	/*
		generate per day time series -> {Day1,Day2}, {Day2,Day3},{Day3,Day4} upto current day
	*/
	startTimestamp := time.Now().AddDate(0, 0, -10) // 10 days excluding today
	startDate := time.Date(startTimestamp.UTC().Year(), startTimestamp.UTC().Month(), startTimestamp.UTC().Day(), 0, 0, 0, 0, time.UTC)
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
		documents, _ := store.GetStore().GetHubspotDocumentsByTypeANDRangeForSync(project.ID, model.HubspotDocumentTypeContact, resultTimeSeries[i][0], resultTimeSeries[i][1], time.Now().Unix())
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
	enrichStatus, _ := IntHubspot.Sync(project.ID, numParallelDocuments, time.Now().Unix(), nil, "", 100)
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
		},
	}

	result, status := store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID, C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, status)

	rows := result.Results[0].Rows
	sort.Slice(rows, func(i, j int) bool {
		p1, _ := U.GetPropertyValueAsFloat64(rows[i][2])
		p2, _ := U.GetPropertyValueAsFloat64(rows[j][2])
		return p1 < p2
	})

	contactTimestamp := createdAt
	for i := 0; i < 10; i++ {
		if i == 0 {
			assert.Equal(t, fmt.Sprintf("%d", contactTimestamp.Unix()), rows[i][2])
		} else {
			assert.Equal(t, fmt.Sprintf("%d", contactTimestamp.Unix()), rows[i][2])
		}
		contactTimestamp = contactTimestamp.AddDate(0, 0, 1)
	}

	// Verfiying contact to company association
	for i := 1; i <= 3; i++ {
		companyID := int64(i)
		companyContact := []int64{int64(i)}
		contactIDS := []string{}
		for i := range companyContact {
			contactIDS = append(contactIDS, fmt.Sprintf("%d", companyContact[i]))
		}
		companyIDstring := fmt.Sprintf("%d", companyID)
		companyDocuments, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{companyIDstring}, model.HubspotDocumentTypeCompany, []int{model.HubspotDocumentActionCreated})

		contactDocuments, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, contactIDS, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
		assert.Equal(t, http.StatusFound, status)
		assert.Len(t, contactDocuments, 1)
		for i := range contactDocuments {
			contactUser, status := store.GetStore().GetUser(project.ID, contactDocuments[i].UserId)
			assert.Equal(t, http.StatusFound, status)
			// verify group_1_id is company unique id and group_1_user_id is company user_id
			assert.Equal(t, true, assertUserGroupValueByColumnName(contactUser, "group_1_id", companyDocuments[0].ID))
			assert.Equal(t, true, assertUserGroupValueByColumnName(contactUser, "group_1_user_id", companyDocuments[0].GroupUserId))
		}
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
	result, status = store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID, C.EnableOptimisedFilterOnEventUserQuery())
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
	result, status = store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID, C.EnableOptimisedFilterOnEventUserQuery())
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

func TestHubspotGetHubspotContactCreatedSyncIDAndUserID(t *testing.T) {
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
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, eventID, hubspotDocument.Timestamp, model.HubspotDocumentActionCreated, userID1, "")
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

	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, eventID, updatedDate.Unix(), model.HubspotDocumentActionUpdated, userID1, "")
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

	// company with 4 contacts
	company1ID := int64(1)
	company2ID := int64(2)
	company3ID := int64(3)
	company4ID := int64(4)

	company1Contact := []int64{1, 2, 3, 4}
	companyCreatedDate := time.Now().AddDate(0, 0, -5)
	companyUpdatedDate := companyCreatedDate.AddDate(0, 0, 1)
	company := IntHubspot.Company{
		CompanyId:  company1ID,
		ContactIds: company1Contact,
		Properties: map[string]IntHubspot.Property{
			"createdate":             {Value: fmt.Sprintf("%d", companyCreatedDate.Unix()*1000)},
			"hs_lastmodifieddate":    {Value: fmt.Sprintf("%d", companyUpdatedDate.Unix()*1000)},
			"company_lifecyclestage": {Value: "lead"},
			"name": {
				Value:     "testcompany",
				Timestamp: companyCreatedDate.Unix() * 1000,
			},
			"company_id": {
				Value: fmt.Sprintf("%d", company1ID),
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
	status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeCompany, []*model.HubspotDocument{&hubspotDocument}, 1)
	assert.Equal(t, http.StatusCreated, status)

	// extra company creation for go routines test
	for _, companyID := range []int64{company2ID, company3ID, company4ID} {
		company.CompanyId = companyID
		company.ContactIds = nil
		enJSON, err = json.Marshal(company)
		assert.Nil(t, err)
		companyPJson = postgres.Jsonb{json.RawMessage(enJSON)}
		hubspotDocument := model.HubspotDocument{
			TypeAlias: model.HubspotDocumentTypeNameCompany,
			Value:     &companyPJson,
		}
		status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeCompany, []*model.HubspotDocument{&hubspotDocument}, 1)
		assert.Equal(t, http.StatusCreated, status)
	}

	// contacts for company
	for i := range company1Contact {
		contact := IntHubspot.Contact{
			Vid: company1Contact[i],
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
		status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeContact, []*model.HubspotDocument{&hubspotDocument}, 1)
		assert.Equal(t, http.StatusCreated, status)
	}

	deal1 := int64(1)
	deal2 := int64(2)
	deal3 := int64(3)
	deal4 := int64(4)

	dealIds := []int64{deal1, deal2, deal3, deal4}
	dealCompanyAssociations := [][]int64{{company1ID}, {company2ID, company3ID}, {}, {}}
	dealContactAssociations := [][]int64{{}, {company1Contact[1]}, {company1Contact[2]}, {company1Contact[0], company1Contact[2]}}
	dealStartTimestamp := time.Now().AddDate(0, 0, -1)
	for i := range dealIds {
		deal := IntHubspot.Deal{
			DealId: dealIds[i],
			Properties: map[string]IntHubspot.Property{
				"hs_createdate": {
					Value: fmt.Sprintf("%d", dealStartTimestamp.Add(time.Duration(dealIds[i])*time.Hour).Unix()*1000),
				},
				"hs_lastmodifieddate": {
					Value: fmt.Sprintf("%d", dealStartTimestamp.Add(time.Duration(dealIds[i])*time.Hour).Add(20*time.Minute).Unix()*1000),
				},
				"stage": {
					Value: fmt.Sprintf("deal%d In Progress", dealIds[i]),
				},
			},
			Associations: IntHubspot.Associations{
				AssociatedCompanyIds: dealCompanyAssociations[i],
				AssociatedContactIds: dealContactAssociations[i],
			},
		}

		enJSON, err = json.Marshal(deal)
		assert.Nil(t, err)
		dealPJson := postgres.Jsonb{json.RawMessage(enJSON)}
		hubspotDocument = model.HubspotDocument{
			TypeAlias: model.HubspotDocumentTypeNameDeal,
			Value:     &dealPJson,
		}

		status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeDeal, []*model.HubspotDocument{&hubspotDocument}, 1)
		assert.Equal(t, http.StatusCreated, status)
	}

	enrichStatus, _ := IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50)

	assert.Equal(t, project.ID, enrichStatus[0].ProjectId)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)

	// Verfiying contact to company association
	contactIDS := []string{}
	for i := range company1Contact {
		contactIDS = append(contactIDS, fmt.Sprintf("%d", company1Contact[i]))
	}
	companyIDstring := fmt.Sprintf("%d", company1ID)
	companyDocuments, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{companyIDstring}, model.HubspotDocumentTypeCompany, []int{model.HubspotDocumentActionCreated})

	contactDocuments, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, contactIDS, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, contactDocuments, 4)
	for i := range contactDocuments {
		contactUser, status := store.GetStore().GetUser(project.ID, contactDocuments[i].UserId)
		assert.Equal(t, http.StatusFound, status)
		// verify group_1_id is company unique id and group_1_user_id is company user_id
		assert.Equal(t, true, assertUserGroupValueByColumnName(contactUser, "group_1_id", companyDocuments[0].ID))
		assert.Equal(t, true, assertUserGroupValueByColumnName(contactUser, "group_1_user_id", companyDocuments[0].GroupUserId))
	}
	company1GroupUserID := companyDocuments[0].GroupUserId
	var company2GroupUserID, company3GroupUserID string
	for _, companyID := range []int64{company2ID, company3ID} {
		companyDocuments, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%d", companyID)}, model.HubspotDocumentTypeCompany, []int{model.HubspotDocumentActionCreated})
		assert.Equal(t, http.StatusFound, status)
		if companyID == company2ID {
			company2GroupUserID = companyDocuments[0].GroupUserId
		}
		if companyID == company3ID {
			company3GroupUserID = companyDocuments[0].GroupUserId
		}
	}
	/*
		Contact moving to different company will not be updated
	*/
	company.CompanyId = 2
	company.ContactIds = company1Contact
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
	status = store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeCompany, []*model.HubspotDocument{&hubspotDocument}, 1)
	assert.Equal(t, http.StatusCreated, status)

	enrichStatus, _ = IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectId)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)

	// verify user still associated with previous company
	for i := range contactDocuments {
		contactUser, status := store.GetStore().GetUser(project.ID, contactDocuments[i].UserId)
		assert.Equal(t, http.StatusFound, status)
		assert.Equal(t, true, assertUserGroupValueByColumnName(contactUser, "group_1_id", companyDocuments[0].ID))
		assert.Equal(t, true, assertUserGroupValueByColumnName(contactUser, "group_1_user_id", companyDocuments[0].GroupUserId))
	}

	// total company events
	query := model.Query{
		From: companyCreatedDate.Unix() - 500,
		To:   companyUpdatedDate.AddDate(0, 0, 1).Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				Property:       "$hubspot_company_name",
				EventName:      U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				EventNameIndex: 1,
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondEachGivenEvent,
	}

	result, status := store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID, C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED, result.Results[0].Rows[0][1])
	assert.Equal(t, "testcompany", result.Results[0].Rows[0][2])
	assert.Equal(t, float64(4), result.Results[0].Rows[0][3])

	// total users
	query = model.Query{
		From: companyCreatedDate.Unix() - 500,
		To:   companyUpdatedDate.AddDate(0, 0, 1).Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}

	result, status = store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID, C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, float64(4), result.Results[0].Rows[0][2])

	/*
		Test use company domain name if company name not available
	*/

	companyID := int64(10)
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
	status = store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeCompany, []*model.HubspotDocument{&hubspotDocument}, 1)
	assert.Equal(t, http.StatusCreated, status)

	enrichStatus, _ = IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50)

	assert.Equal(t, project.ID, enrichStatus[0].ProjectId)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)

	companyDocuments, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%d", companyID)}, model.HubspotDocumentTypeCompany,
		[]int{model.HubspotDocumentActionCreated})
	user, status := store.GetStore().GetUser(project.ID, companyDocuments[0].GroupUserId)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, companyDocuments[0].ID, user.Group1ID)
	assert.Equal(t, model.UserSourceHubspot, *user.Source)
	userProperties, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Equal(t, "lead", (*userProperties)["$hubspot_company_lifecyclestage"])

	// verify deal groups
	var deal1GroupUserID, deal2GroupUserID, deal3GroupUserID, deal4GroupUserID string
	for i := range dealIds {
		documents, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%d", dealIds[i])}, model.HubspotDocumentTypeDeal, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
		assert.Equal(t, http.StatusFound, status)
		assert.Len(t, documents, 3)
		assert.NotEqual(t, "", documents[0].GroupUserId)
		assert.Equal(t, documents[0].GroupUserId, documents[1].GroupUserId)
		if dealIds[i] == 1 {
			deal1GroupUserID = documents[0].GroupUserId
		}
		if dealIds[i] == 2 {
			deal2GroupUserID = documents[0].GroupUserId
		}

		if dealIds[i] == 3 {
			deal3GroupUserID = documents[0].GroupUserId
		}
		if dealIds[i] == 4 {
			deal4GroupUserID = documents[0].GroupUserId
		}
	}

	//deal1
	groupRelationship, status := store.GetStore().GetGroupRelationshipByUserID(project.ID, deal1GroupUserID)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, groupRelationship, 1)
	assert.Equal(t, groupRelationship[0].LeftGroupUserID, deal1GroupUserID)
	assert.Equal(t, groupRelationship[0].RightGroupUserID, company1GroupUserID)

	//deal2
	groupRelationship, status = store.GetStore().GetGroupRelationshipByUserID(project.ID, deal2GroupUserID)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, groupRelationship, 2) // 2 company associated
	assert.Equal(t, groupRelationship[0].LeftGroupUserID, deal2GroupUserID)
	assert.Equal(t, groupRelationship[1].LeftGroupUserID, deal2GroupUserID)
	if groupRelationship[0].RightGroupUserID != company2GroupUserID {
		assert.Equal(t, groupRelationship[0].RightGroupUserID, company3GroupUserID)
	}
	companyContacts, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%d", company1Contact[1])}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	user, _ = store.GetStore().GetUser(project.ID, companyContacts[0].UserId)
	assert.True(t, assertUserGroupValueByColumnName(user, "group_2_user_id", deal2GroupUserID))

	// deal3
	groupRelationship, status = store.GetStore().GetGroupRelationshipByUserID(project.ID, deal3GroupUserID)
	assert.Equal(t, http.StatusNotFound, status)
	companyContacts, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%d", company1Contact[2])}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	user, _ = store.GetStore().GetUser(project.ID, companyContacts[0].UserId)
	assert.True(t, assertUserGroupValueByColumnName(user, "group_2_user_id", deal3GroupUserID))

	//deal4
	groupRelationship, status = store.GetStore().GetGroupRelationshipByUserID(project.ID, deal4GroupUserID)
	assert.Equal(t, http.StatusNotFound, status)
	companyContacts, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%d", company1Contact[0])}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	user, _ = store.GetStore().GetUser(project.ID, companyContacts[0].UserId)
	assert.True(t, assertUserGroupValueByColumnName(user, "group_2_user_id", deal4GroupUserID))
	companyContacts, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%d", company1Contact[2])}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	user, _ = store.GetStore().GetUser(project.ID, companyContacts[0].UserId)
	assert.False(t, assertUserGroupValueByColumnName(user, "group_2_user_id", deal4GroupUserID))

	// deal1 later getting associated to contact2 and company2
	// deal1 existing mapping company - > company1ID  contact -> nil
	// new  company - > company1ID,company2ID  contact - > company1Contact[3]

	// verify contact not associated to any
	documents, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%d", company1Contact[3])}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 2)
	assert.Equal(t, "", documents[0].GroupUserId)
	user, status = store.GetStore().GetUser(project.ID, documents[0].UserId)
	assert.Equal(t, http.StatusFound, status)
	assert.True(t, assertUserGroupValueByColumnName(user, "group_2_user_id", ""))

	deal := IntHubspot.Deal{
		DealId: dealIds[0],
		Properties: map[string]IntHubspot.Property{
			"hs_createdate": {
				Value: fmt.Sprintf("%d", dealStartTimestamp.Add(time.Duration(dealIds[0])*time.Hour).Unix()*1000),
			},
			"hs_lastmodifieddate": {
				Value: fmt.Sprintf("%d", dealStartTimestamp.Add(time.Duration(dealIds[0])*time.Hour).Add(30*time.Minute).Unix()*1000),
			},
			"stage": {
				Value: fmt.Sprintf("deal%d In Progress", dealIds[0]),
			},
		},
		Associations: IntHubspot.Associations{
			AssociatedCompanyIds: []int64{dealCompanyAssociations[0][0], company2ID},
			AssociatedContactIds: []int64{company1Contact[3]},
		},
	}

	enJSON, err = json.Marshal(deal)
	assert.Nil(t, err)
	dealPJson := postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameDeal,
		Value:     &dealPJson,
	}

	status = store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeDeal, []*model.HubspotDocument{&hubspotDocument}, 1)
	assert.Equal(t, http.StatusCreated, status)

	enrichStatus, _ = IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectId)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)

	documents, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%d", dealIds[0])}, model.HubspotDocumentTypeDeal, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 4)
	for i := range documents {
		assert.Equal(t, documents[i].GroupUserId, deal1GroupUserID)
	}

	documents, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%d", company1Contact[3])}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 2)
	assert.Equal(t, "", documents[0].GroupUserId)
	user, status = store.GetStore().GetUser(project.ID, documents[0].UserId)
	assert.Equal(t, http.StatusFound, status)
	assert.True(t, assertUserGroupValueByColumnName(user, "group_2_user_id", deal1GroupUserID))

	groupRelationship, status = store.GetStore().GetGroupRelationshipByUserID(project.ID, deal1GroupUserID)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, groupRelationship, 2)
	assert.NotEqual(t, groupRelationship[0].RightGroupUserID, groupRelationship[1].RightGroupUserID)
	if groupRelationship[0].RightGroupUserID != company1GroupUserID {
		assert.Equal(t, groupRelationship[0].RightGroupUserID, company2GroupUserID)
	}

	// deal3 getting associated to company2 but without updated timestamp
	// should create new record of action associationupdated with timestamp = prevtimestamp +1

	deal = IntHubspot.Deal{
		DealId: dealIds[2],
		Properties: map[string]IntHubspot.Property{
			"hs_createdate": {
				Value: fmt.Sprintf("%d", dealStartTimestamp.Add(time.Duration(dealIds[2])*time.Hour).Unix()*1000),
			},
			"hs_lastmodifieddate": {
				Value: fmt.Sprintf("%d", dealStartTimestamp.Add(time.Duration(dealIds[2])*time.Hour).Add(20*time.Minute).Unix()*1000),
			},
			"stage": {
				Value: fmt.Sprintf("deal%d In Progress", dealIds[2]),
			},
		},
		Associations: IntHubspot.Associations{
			AssociatedCompanyIds: append(dealCompanyAssociations[2], company2ID),
			AssociatedContactIds: dealContactAssociations[2],
		},
	}

	enJSON, err = json.Marshal(deal)
	assert.Nil(t, err)
	dealPJson = postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameDeal,
		Value:     &dealPJson,
	}

	status = store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeDeal, []*model.HubspotDocument{&hubspotDocument}, 1)
	assert.Equal(t, http.StatusCreated, status)
	// inserting again should return status conflict
	status = store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeDeal, []*model.HubspotDocument{&hubspotDocument}, 1)
	assert.Equal(t, http.StatusCreated, status)
	documents, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID,
		[]string{U.GetPropertyValueAsString(dealIds[2])}, model.HubspotDocumentTypeDeal, []int{model.HubspotDocumentActionAssociationsUpdated})
	assert.Len(t, documents, 1)

	enrichStatus, _ = IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)
	groupRelationship, status = store.GetStore().GetGroupRelationshipByUserID(project.ID, deal3GroupUserID)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, groupRelationship, 1)
	assert.Equal(t, company2GroupUserID, groupRelationship[0].RightGroupUserID)
	assert.Equal(t, 1, groupRelationship[0].RightGroupNameID)
	user, _ = store.GetStore().GetUser(project.ID, companyContacts[0].UserId)
	assert.True(t, assertUserGroupValueByColumnName(user, "group_2_user_id", deal3GroupUserID))
}

/*func TestHubspotOfflineTouchPoint(t *testing.T) {

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
		RequestSource:   model.UserSourceHubspot,
	}
	userID1 := U.RandomLowerAphaNumString(5)
	createdUserID, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID1, CustomerUserId: cuid, JoinTimestamp: getEventTimestamp(hubspotDocument.Timestamp), Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
	assert.Equal(t, http.StatusCreated, status)

	trackPayload.UserId = createdUserID

	filter1 := model.TouchPointFilter{
		Property:  "$hubspot_campaign_name",
		Operator:  "contains",
		Value:     "Webinar",
		LogicalOp: "AND",
	}

	rulePropertyMap := make(map[string]model.TouchPointPropertyValue)
	rulePropertyMap["$campaign"] = model.TouchPointPropertyValue{Type: model.TouchPointPropertyValueAsProperty, Value: "$hubspot_campaign_name"}
	rulePropertyMap["$channel"] = model.TouchPointPropertyValue{Type: model.TouchPointPropertyValueAsConstant, Value: "Other"}

	f, _ := json.Marshal([]model.TouchPointFilter{filter1})
	rPM, _ := json.Marshal(rulePropertyMap)

	rule := model.OTPRule{
		Filters:           postgres.Jsonb{json.RawMessage(f)},
		TouchPointTimeRef: model.LastModifiedTimeRef,
		PropertiesMap:     postgres.Jsonb{json.RawMessage(rPM)},
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

	_, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	filter1 := model.TouchPointFilter{
		Property:  "$hubspot_campaign_name",
		Operator:  "contains",
		Value:     "Webinar",
		LogicalOp: "AND",
	}

	rulePropertyMap := make(map[string]model.TouchPointPropertyValue)
	rulePropertyMap["$campaign"] = model.TouchPointPropertyValue{Type: model.TouchPointPropertyValueAsProperty, Value: "$hubspot_campaign_type"}
	rulePropertyMap["$channel"] = model.TouchPointPropertyValue{Type: model.TouchPointPropertyValueAsConstant, Value: "Other"}

	f, _ := json.Marshal([]model.TouchPointFilter{filter1})
	rPM, _ := json.Marshal(rulePropertyMap)

	rule := model.OTPRule{
		Filters:           postgres.Jsonb{json.RawMessage(f)},
		TouchPointTimeRef: model.LastModifiedTimeRef,
		PropertiesMap:     postgres.Jsonb{json.RawMessage(rPM)},
	}
	fmt.Println(rule)

}
*/
func getEventTimestamp(timestamp int64) int64 {
	if timestamp == 0 {
		return 0
	}
	return timestamp / 1000
}

func TestHubspotUserPropertiesOverwrite(t *testing.T) {
	// Initialize the project and the user. Also capture currentTimestamp, futureTimestamp & middleTimestamp.
	currentTimestamp := time.Now().Unix()
	futureTimestamp := currentTimestamp + 10000
	middleTimestamp := currentTimestamp + 1000
	fmt.Printf("\ncurrentTimestamp : %d\nfutureTimestamp : %d\nmiddleTimestamp : %d\n", currentTimestamp, futureTimestamp, middleTimestamp)
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.Properties)
	_, errCode := store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)

	// Update user properties lastmodifieddate as middleTimestamp, PropertiesUpdatedTimestamp
	// as futureTimestamp.
	newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
		`{"country": "india", "age": 30.1, "paid": true, "$hubspot_contact_lastmodifieddate": %d}`, middleTimestamp)))}
	_, status := store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, futureTimestamp)
	assert.Equal(t, http.StatusAccepted, status)
	storedUser, errCode := store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, futureTimestamp, storedUser.PropertiesUpdatedTimestamp)
	var propertiesMap map[string]interface{}
	err = json.Unmarshal((storedUser.Properties).RawMessage, &propertiesMap)
	assert.Nil(t, err)
	storedLastModifiedDate, err := U.GetPropertyValueAsFloat64(propertiesMap["$hubspot_contact_lastmodifieddate"])
	assert.Nil(t, err)
	assert.Equal(t, middleTimestamp, int64(storedLastModifiedDate))

	// Update user property lastmodifieddate as futureTimestamp and PropertiesUpdatedTimestamp as currentTimestamp.
	// Since the source and object-type are blank, the property value and PropertiesUpdatedTimestamp should not get
	// updated.
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
		`{"$hubspot_contact_lastmodifieddate": %d}`, futureTimestamp)))}
	_, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, currentTimestamp)
	assert.Equal(t, http.StatusAccepted, status)
	storedUser, errCode = store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, futureTimestamp, storedUser.PropertiesUpdatedTimestamp)
	var updatedPropertiesMap map[string]interface{}
	err = json.Unmarshal((storedUser.Properties).RawMessage, &updatedPropertiesMap)
	assert.Nil(t, err)
	storedLastModifiedDate, err = U.GetPropertyValueAsFloat64(updatedPropertiesMap["$hubspot_contact_lastmodifieddate"])
	assert.Nil(t, err)
	assert.Equal(t, middleTimestamp, int64(storedLastModifiedDate))

	// Get oldTimestamp, before the futureTimestamp.
	oldTimestamp := futureTimestamp - 1000
	fmt.Printf("\noldTimestamp : %d\n", oldTimestamp)

	// Update user properties lastmodifieddate as futureTimestamp, PropertiesUpdatedTimestamp as oldTimestamp.
	// lastmodifieddate should get updated with futureTimestamp, but PropertiesUpdatedTimestamp should remain unchanged.
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
		`{"country": "india", "age": 30.1, "paid": true, "$hubspot_contact_lastmodifieddate": %d}`, futureTimestamp)))}
	_, status = store.GetStore().UpdateUserPropertiesV2(project.ID, user.ID,
		newProperties, oldTimestamp, SDK.SourceHubspot, model.HubspotDocumentTypeNameContact)
	assert.Equal(t, http.StatusAccepted, status)
	storedUser, errCode = store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, futureTimestamp, storedUser.PropertiesUpdatedTimestamp)
	err = json.Unmarshal((storedUser.Properties).RawMessage, &propertiesMap)
	assert.Nil(t, err)
	storedLastModifiedDate, err = U.GetPropertyValueAsFloat64(propertiesMap["$hubspot_contact_lastmodifieddate"])
	assert.Nil(t, err)
	assert.Equal(t, futureTimestamp, int64(storedLastModifiedDate))

	// hubspot record test -> Testing single user
	project, _, err = SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	createDocumentID := rand.Intn(100)
	timestampT1 := time.Now().AddDate(0, 0, -1).Unix() * 1000
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

	// create contact record with (properties->‘lastmodified’:timestampT1)
	jsonContact := fmt.Sprintf(jsonContactModel, createDocumentID, timestampT1, timestampT1, timestampT1, "lead", cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionCreated,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// Execute sync job to process the contact created above
	enrichStatus, _ := IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	// Verification for contact creation.
	createDocument, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", createDocumentID)}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, createDocument[0].UserId)
	assert.Equal(t, http.StatusFound, status)
	properitesMap := make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)

	// Verify hubspot_contact_lastmodifieddate is set to timestampT1
	lastmodifieddateProperty := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContact,
		U.PROPERTY_KEY_LAST_MODIFIED_DATE)
	userPropertyValue, err := U.GetPropertyValueAsFloat64(properitesMap[lastmodifieddateProperty])
	assert.Equal(t, err, nil)
	assert.Equal(t, timestampT1, int64(userPropertyValue)*1000)

	// Update user properties (“a”:1) with timestampT3
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"a": 1}`))}
	timestampT3 := timestampT1 + 10000
	_, status = store.GetStore().UpdateUserPropertiesV2(project.ID, user.ID,
		newProperties, timestampT3, SDK.SourceHubspot, model.HubspotDocumentTypeNameContact)
	assert.Equal(t, http.StatusAccepted, status)
	timestampT2 := timestampT1 + 1000

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

	// update contact record with (properties->‘lastmodified’:timestampT2)
	jsonContact = fmt.Sprintf(jsonContactModel, createDocumentID, timestampT1, timestampT1, timestampT2, "lead", cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionUpdated,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// Execute sync job to process the contact updated above
	enrichStatus, _ = IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	// Verify hubspot_contact_lastmodifieddate is set to timestampT2 and PropertiesUpdatedTimestamp to timestampT3.
	updateDocument, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", createDocumentID)}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, updateDocument[0].UserId)
	assert.Equal(t, http.StatusFound, status)
	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	lastmodifieddateProperty = model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContact,
		U.PROPERTY_KEY_LAST_MODIFIED_DATE)
	userPropertyValue, err = U.GetPropertyValueAsFloat64(properitesMap[lastmodifieddateProperty])
	assert.Equal(t, err, nil)
	assert.Equal(t, timestampT2, int64(userPropertyValue)*1000)
	assert.Equal(t, timestampT1, user.PropertiesUpdatedTimestamp*1000)

	// hubspot record test -> Testing multi-user by customer-user-id
	project, _, err = SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	createDocumentIDU1 := rand.Intn(100)
	timestampT1 = time.Now().AddDate(0, 0, -1).Unix() * 1000
	cuid_first := getRandomEmail()

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

	// create contact record createDocumentIDU1 with (properties->‘lastmodified’:timestampT1) and email property
	// ("email": cuid_first)
	jsonContact = fmt.Sprintf(jsonContactModel, createDocumentIDU1, timestampT1, timestampT1, timestampT1, "lead", cuid_first, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionCreated,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// Execute sync job to process the contact created above
	enrichStatus, _ = IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	// Create normal user U2 (createUserU2) with same email property as that of createDocumentIDU1 ("email": cuid_first)
	userU2, errCode1 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: cuid_first, JoinTimestamp: timestampT3, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
	assert.Equal(t, http.StatusCreated, errCode1)

	// Verify lastmodifieddate user property of userU2 to be timestampT1, which is same as createDocumentIDU1
	user, status = store.GetStore().GetUser(project.ID, userU2)
	assert.Equal(t, http.StatusFound, status)
	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	lastmodifieddateProperty = model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContact,
		U.PROPERTY_KEY_LAST_MODIFIED_DATE)
	userPropertyValue, err = U.GetPropertyValueAsFloat64(properitesMap[lastmodifieddateProperty])
	assert.Equal(t, err, nil)
	assert.Equal(t, timestampT1, int64(userPropertyValue)*1000)

	// Update user properties (“a”:1) with timestampT3 for userU2
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"a": 1}`))}
	timestampT3 = timestampT1 + 10000
	_, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, timestampT3)
	assert.Equal(t, http.StatusAccepted, status)

	timestampT2 = timestampT1 + 1000

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

	// create contact updated record for createDocumentIDU1 with (properties->‘lastmodified’:timestampT2)
	jsonContact = fmt.Sprintf(jsonContactModel, createDocumentIDU1, timestampT1, timestampT1, timestampT2, "lead", cuid_first, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionUpdated,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// Execute sync job to process the contact updated above
	enrichStatus, _ = IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	// Verify hubspot_contact_lastmodifieddate is set to timestampT2 for both createDocumentIDU1 and userU2.
	// Verify PropertiesUpdatedTimestamp is set to timestampT3 for both createDocumentIDU1 and userU2.
	updateDocument, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", createDocumentIDU1)}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, updateDocument[0].UserId)
	assert.Equal(t, http.StatusFound, status)
	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	lastmodifieddateProperty = model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContact,
		U.PROPERTY_KEY_LAST_MODIFIED_DATE)
	userPropertyValue, err = U.GetPropertyValueAsFloat64(properitesMap[lastmodifieddateProperty])
	assert.Equal(t, err, nil)
	assert.Equal(t, timestampT2, int64(userPropertyValue)*1000)
	assert.Equal(t, timestampT3, user.PropertiesUpdatedTimestamp)

	user, status = store.GetStore().GetUser(project.ID, userU2)
	assert.Equal(t, http.StatusFound, status)
	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	lastmodifieddateProperty = model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContact,
		U.PROPERTY_KEY_LAST_MODIFIED_DATE)
	userPropertyValue, err = U.GetPropertyValueAsFloat64(properitesMap[lastmodifieddateProperty])
	assert.Equal(t, err, nil)
	assert.Equal(t, timestampT2, int64(userPropertyValue)*1000)
	assert.Equal(t, timestampT3, user.PropertiesUpdatedTimestamp)
}

func TestHubspotGroupUserFix(t *testing.T) {
	// Initialize the project and the user.
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)

	// create hubspot-company record
	timestamp := time.Now().AddDate(0, 0, 0).Unix() * 1000

	companyCreatedDate := time.Now().AddDate(0, 0, -5)
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

	enJSON, err := json.Marshal(company)
	assert.Nil(t, err)
	companyPJson := postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameCompany,
		Value:     &companyPJson,
	}
	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// create hubspot-deal record
	dealCompanyAssociations := [][]int64{{int64(1)}, {int64(2), int64(3)}, {}, {}}
	company1Contact := []int64{1, 2, 3, 4}

	deal := IntHubspot.Deal{
		DealId: int64(1),
		Properties: map[string]IntHubspot.Property{
			"hs_createdate": {
				Value: fmt.Sprintf("%d", time.Now().Unix()*1000),
			},
			"hs_lastmodifieddate": {
				Value: fmt.Sprintf("%d", time.Now().Unix()*1000),
			},
			"stage": {
				Value: fmt.Sprintf("deal%d In Progress", int64(1)),
			},
		},
		Associations: IntHubspot.Associations{
			AssociatedCompanyIds: []int64{dealCompanyAssociations[0][0], int64(2)},
			AssociatedContactIds: []int64{company1Contact[3]},
		},
	}

	enJSON, err = json.Marshal(deal)
	assert.Nil(t, err)
	dealPJson := postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameDeal,
		Value:     &dealPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// Execute sync job to process the contact created above
	syncStatus, _ := IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
	for i := range syncStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, syncStatus[i].Status)
	}

	// verification for groupID.
	createDocument, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", 1)}, model.HubspotDocumentTypeCompany, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	// verify group_user_id in the document
	assert.NotNil(t, createDocument[0].GroupUserId)
	// verify that group user has groupId as document.ID
	user, status = store.GetStore().GetUser(project.ID, createDocument[0].GroupUserId)
	assert.Equal(t, http.StatusFound, status)
	groupID := GetGroupID(user)
	assert.Equal(t, groupID, createDocument[0].ID)

	// verification for groupID.
	createDocument, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", 1)}, model.HubspotDocumentTypeDeal, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	// verify group_user_id in the document
	assert.NotNil(t, createDocument[0].GroupUserId)
	// verify that group user has groupId as document.ID
	user, status = store.GetStore().GetUser(project.ID, createDocument[0].GroupUserId)
	assert.Equal(t, http.StatusFound, status)
	groupID = GetGroupID(user)
	assert.Equal(t, groupID, createDocument[0].ID)

	// create hubspot-company record
	companyCreatedDate = time.Now().AddDate(0, 0, -5)
	companyUpdatedDate = companyCreatedDate.AddDate(0, 0, 1)
	company = IntHubspot.Company{
		CompanyId:  2,
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
	companyPJson = postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameCompany,
		Value:     &companyPJson,
	}
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// get groupName
	groupName := model.GROUP_NAME_HUBSPOT_COMPANY

	// create group user with random groupID
	groupID = U.RandomLowerAphaNumString(5)
	groupUserID, status := store.GetStore().CreateGroupUser(&model.User{
		ProjectId: project.ID, JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceHubspot),
	}, groupName, groupID)
	assert.Equal(t, http.StatusCreated, status)
	// update group_user_id in the account document, and mark it as synced
	status = store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeCompany, "", companyCreatedDate.Unix()*1000, model.HubspotDocumentActionCreated, "", groupUserID)
	assert.Equal(t, http.StatusAccepted, status)
	// create another update on company record
	companyCreatedDate = time.Now().AddDate(0, 0, -4)
	companyUpdatedDate = companyCreatedDate.AddDate(0, 0, 1)
	company = IntHubspot.Company{
		CompanyId:  2,
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
	companyPJson = postgres.Jsonb{json.RawMessage(enJSON)}
	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameCompany,
		Value:     &companyPJson,
	}
	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// Execute sync job to process the contact created above
	syncStatus, _ = IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
	for i := range syncStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, syncStatus[i].Status)
	}

	// verification for groupID.
	createDocument, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{fmt.Sprintf("%v", 2)}, model.HubspotDocumentTypeCompany, []int{model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, groupUserID)
	assert.Equal(t, http.StatusFound, status)
	groupID = GetGroupID(user)
	assert.Equal(t, groupID, createDocument[0].ID)
}

func TestHubspotBatchCreate(t *testing.T) {

	// Initialize the project and the user.
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)

	contactCreatedDate := time.Now().AddDate(0, 0, -5)
	contactUpdatedDate := contactCreatedDate.AddDate(0, 0, 1)
	processDocuments := make([]*model.HubspotDocument, 0)
	for i := 0; i < 10; i++ {
		contact := IntHubspot.Contact{
			Vid: int64(i),
			Properties: map[string]IntHubspot.Property{
				"createdate":       {Value: fmt.Sprintf("%d", contactCreatedDate.Unix()*1000)},
				"lastmodifieddate": {Value: fmt.Sprintf("%d", contactUpdatedDate.Unix()*1000)},
				"lifecyclestage":   {Value: "lead"},
			},
		}

		enJSON, err := json.Marshal(contact)
		assert.Nil(t, err)
		contactPJson := postgres.Jsonb{json.RawMessage(enJSON)}
		hubspotDocument := model.HubspotDocument{
			TypeAlias: model.HubspotDocumentTypeNameContact,
			Value:     &contactPJson,
		}
		processDocuments = append(processDocuments, &hubspotDocument)
	}

	status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeContact, processDocuments, 5)
	assert.Equal(t, http.StatusCreated, status)

	documents, status := store.GetStore().GetHubspotDocumentsByTypeForSync(project.ID, model.HubspotDocumentTypeContact, time.Now().Unix())
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 30)

	// performing another insert with 3 update and 2 duplicate
	processDocuments = []*model.HubspotDocument{processDocuments[0], processDocuments[1]}
	for i := 2; i < 5; i++ {
		contact := IntHubspot.Contact{
			Vid: int64(i),
			Properties: map[string]IntHubspot.Property{
				"createdate":       {Value: fmt.Sprintf("%d", contactCreatedDate.Unix()*1000+10)},
				"lastmodifieddate": {Value: fmt.Sprintf("%d", contactUpdatedDate.Unix()*1000+10)},
				"lifecyclestage":   {Value: "lead"},
			},
		}

		enJSON, err := json.Marshal(contact)
		assert.Nil(t, err)
		contactPJson := postgres.Jsonb{json.RawMessage(enJSON)}
		hubspotDocument := model.HubspotDocument{
			TypeAlias: model.HubspotDocumentTypeNameContact,
			Value:     &contactPJson,
		}
		processDocuments = append(processDocuments, &hubspotDocument)
	}

	status = store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeContact, processDocuments, 5)
	assert.Equal(t, http.StatusCreated, status)

	documents, status = store.GetStore().GetHubspotDocumentsByTypeForSync(project.ID, model.HubspotDocumentTypeContact, time.Now().Unix())
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 33)

	/*
		Delete contact with previous record and with no previous record
	*/
	processDocuments = make([]*model.HubspotDocument, 0)

	contact := map[string]interface{}{
		"id": int64(1),
		"properties": map[string]interface{}{
			"createdate":       contactCreatedDate.Format(model.HubspotDateTimeLayout),
			"lastmodifieddate": contactUpdatedDate.Add(20 * time.Second).Format(model.HubspotDateTimeLayout),
			"lifecyclestage":   "junk",
		},
		"archived": true,
	}

	enJSON, err := json.Marshal(contact)
	assert.Nil(t, err)
	contactPJson1 := postgres.Jsonb{json.RawMessage(enJSON)}
	deleteDocument1 := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson1,
		Action:    model.HubspotDocumentActionDeleted,
	}
	processDocuments = append(processDocuments, &deleteDocument1)

	// not existing record
	contact = map[string]interface{}{
		"id": int64(14),
		"properties": map[string]interface{}{
			"createdate":       contactCreatedDate.Format(model.HubspotDateTimeLayout),
			"lastmodifieddate": contactUpdatedDate.Add(20 * time.Second).Format(model.HubspotDateTimeLayout),
			"lifecyclestage":   "junk",
		},
		"archived": true,
	}

	enJSON, err = json.Marshal(contact)
	assert.Nil(t, err)
	contactPJson2 := postgres.Jsonb{json.RawMessage(enJSON)}
	deleteDocument2 := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson2,
		Action:    model.HubspotDocumentActionDeleted,
	}
	processDocuments = append(processDocuments, &deleteDocument2)
	status = store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeContact, processDocuments, 2)
	assert.Equal(t, http.StatusCreated, status)

	documents, status = store.GetStore().GetHubspotDocumentsByTypeForSync(project.ID, model.HubspotDocumentTypeContact, time.Now().Unix())
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 34)
	documents, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"1"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionDeleted})
	assert.Equal(t, http.StatusFound, status)
	documents, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"14"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionDeleted})
	assert.Equal(t, http.StatusNotFound, status)
}

func TestHubspotEmptyPropertiesUpdated(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	contactCreatedDate := time.Now().AddDate(0, 0, -5)
	customerUserID := getRandomEmail()
	contact := IntHubspot.Contact{
		Vid: 1,
		Properties: map[string]IntHubspot.Property{
			"createdate":       {Value: fmt.Sprintf("%d", contactCreatedDate.Unix()*1000+10)},
			"lastmodifieddate": {Value: fmt.Sprintf("%d", contactCreatedDate.Unix()*1000+10)},
			"lifecyclestage":   {Value: "lead"},
			"Workflow":         {Value: "A"},
		},
		IdentityProfiles: []IntHubspot.ContactIdentityProfile{
			{
				[]IntHubspot.ContactIdentity{{
					Type:  "EMAIL",
					Value: customerUserID,
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
	contactDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}
	processDocuments := []*model.HubspotDocument{&contactDocument}
	status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeContact, processDocuments, 2)
	assert.Equal(t, http.StatusCreated, status)

	enrichStatus, _ := IntHubspot.Sync(project.ID, 2, time.Now().UTC().Unix(), nil, "", 50)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	documents, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"1"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)

	// validated all properties exist in event properties, event user properties and user properties
	for i := range documents {
		user, status := store.GetStore().GetUser(project.ID, documents[i].UserId)
		assert.Equal(t, http.StatusFound, status)
		event, status := store.GetStore().GetEventById(project.ID, documents[i].SyncId, documents[i].UserId)
		assert.Equal(t, http.StatusFound, status)
		var userProperties map[string]interface{}
		var eventProperties map[string]interface{}
		var eventUserProperties map[string]interface{}
		json.Unmarshal(user.Properties.RawMessage, &userProperties)
		json.Unmarshal(event.Properties.RawMessage, &eventProperties)
		json.Unmarshal(event.UserProperties.RawMessage, &eventUserProperties)
		for key, value := range map[string]interface{}{"lifecyclestage": "lead", "Email": customerUserID, "Workflow": "A"} {
			enKey := model.GetCRMEnrichPropertyKeyByType(U.CRM_SOURCE_NAME_HUBSPOT,
				model.HubspotDocumentTypeNameContact, key)
			assert.Equal(t, value, userProperties[enKey])
			assert.Equal(t, value, eventProperties[enKey])
			assert.Equal(t, value, eventUserProperties[enKey])
		}
	}

	contact.Properties = map[string]IntHubspot.Property{
		"createdate":       {Value: fmt.Sprintf("%d", contactCreatedDate.Unix()*1000+10)},
		"lastmodifieddate": {Value: fmt.Sprintf("%d", contactCreatedDate.Unix()*1000+20)},
		"lifecyclestage":   {Value: ""},
		"Workflow":         {Value: ""},
	}

	enJSON, err = json.Marshal(contact)
	assert.Nil(t, err)
	contactPJson = postgres.Jsonb{json.RawMessage(enJSON)}
	contactDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}
	processDocuments = []*model.HubspotDocument{&contactDocument}
	status = store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeContact, processDocuments, 2)
	assert.Equal(t, http.StatusCreated, status)
	enrichStatus, _ = IntHubspot.Sync(project.ID, 2, time.Now().UTC().Unix(), nil, "", 50)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	documents, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"1"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionUpdated})
	assert.Equal(t, http.StatusFound, status)
	documents = documents[len(documents)-1:] // check the latest processed document

	// Empty properties should overwrite previous non empty properties
	for i := range documents {
		user, status := store.GetStore().GetUser(project.ID, documents[i].UserId)
		assert.Equal(t, http.StatusFound, status)
		event, status := store.GetStore().GetEventById(project.ID, documents[i].SyncId, documents[i].UserId)
		assert.Equal(t, http.StatusFound, status)
		var userProperties map[string]interface{}
		var eventProperties map[string]interface{}
		var eventUserProperties map[string]interface{}
		json.Unmarshal(user.Properties.RawMessage, &userProperties)
		json.Unmarshal(event.Properties.RawMessage, &eventProperties)
		json.Unmarshal(event.UserProperties.RawMessage, &eventUserProperties)
		for key, value := range map[string]interface{}{"lifecyclestage": "", "Email": customerUserID, "Workflow": ""} {
			enKey := model.GetCRMEnrichPropertyKeyByType(U.CRM_SOURCE_NAME_HUBSPOT,
				model.HubspotDocumentTypeNameContact, key)
			assert.Equal(t, value, userProperties[enKey])
			assert.Equal(t, value, eventProperties[enKey])
			assert.Equal(t, value, eventUserProperties[enKey])
		}
	}
}

func TestHubspotProjectDistributer(t *testing.T) {
	project1, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	project2, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	intHubspot := true
	_, errCode := store.GetStore().UpdateProjectSettings(project1.ID, &model.ProjectSetting{
		IntHubspot: &intHubspot, IntHubspotApiKey: "1234",
	})
	assert.Equal(t, http.StatusAccepted, errCode)
	_, errCode = store.GetStore().UpdateProjectSettings(project2.ID, &model.ProjectSetting{
		IntHubspot: &intHubspot, IntHubspotApiKey: "1234",
	})
	assert.Equal(t, http.StatusAccepted, errCode)
	contactCreatedDate := time.Now().AddDate(0, 0, -5)
	contact := IntHubspot.Contact{
		Vid: 1,
		Properties: map[string]IntHubspot.Property{
			"createdate":       {Value: fmt.Sprintf("%d", contactCreatedDate.Unix()*1000+10)},
			"lastmodifieddate": {Value: fmt.Sprintf("%d", contactCreatedDate.Unix()*1000+10)},
			"lifecyclestage":   {Value: "lead"},
			"Workflow":         {Value: "A"},
		},
		IdentityProfiles: []IntHubspot.ContactIdentityProfile{
			{
				[]IntHubspot.ContactIdentity{
					{
						Type:  "LEAD_GUID",
						Value: "123-45",
					},
				},
			},
		},
	}
	// project_id 1 - > 20 records, project_id 2 - 40 records
	for projecID, count := range map[int64]int{project1.ID: 10, project2.ID: 20} {
		processDocuments := []*model.HubspotDocument{}
		for i := 0; i < count; i++ {
			contact.Vid = int64(i)
			enJSON, err := json.Marshal(contact)
			assert.Nil(t, err)
			contactPJson := postgres.Jsonb{json.RawMessage(enJSON)}
			contactDocument := model.HubspotDocument{
				TypeAlias: model.HubspotDocumentTypeNameContact,
				Value:     &contactPJson,
			}
			processDocuments = append(processDocuments, &contactDocument)
		}
		status := store.GetStore().CreateHubspotDocumentInBatch(projecID, model.HubspotDocumentTypeContact, processDocuments, 5)
		assert.Equal(t, http.StatusCreated, status)
	}

	// threshold of 20 should distribute project 1 to light and 2 to heavy
	config := map[string]interface{}{
		"light_projects_count_threshold": 20,
		"health_check_ping_id":           "",
		"override_healthcheck_ping_id":   "",
		"max_record_created_at":          time.Now().Unix(),
	}

	jobStatus, success := hubspot_enrich.RunHubspotProjectDistributer(config)
	assert.Equal(t, true, success)
	assert.Contains(t, jobStatus["light_projects"], project1.ID) // project 1 has been marked as light
	assert.NotContains(t, jobStatus["light_projects"], project2.ID)
	assert.Contains(t, jobStatus["heavy_projects"], project2.ID) // project 2 has been marked as heavy_project
	assert.NotContains(t, jobStatus["heavy_projects"], project1.ID)
	crmSetting, status := store.GetStore().GetCRMSetting(project1.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, false, crmSetting.HubspotEnrichHeavy)
	crmSetting, status = store.GetStore().GetCRMSetting(project2.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, true, crmSetting.HubspotEnrichHeavy)

	// running RunHubspotProjectDistributer again shouldn't return the currently marked heavy project
	jobStatus, success = hubspot_enrich.RunHubspotProjectDistributer(config)
	assert.Equal(t, true, success)

	assert.Contains(t, jobStatus["light_projects"], project1.ID)
	assert.NotContains(t, jobStatus["light_projects"], project2.ID)
	assert.NotContains(t, jobStatus["heavy_projects"], project2.ID) // project 2 not present as still marked as heavy project
	assert.NotContains(t, jobStatus["heavy_projects"], project1.ID)
	crmSetting, status = store.GetStore().GetCRMSetting(project1.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, false, crmSetting.HubspotEnrichHeavy)
	crmSetting, status = store.GetStore().GetCRMSetting(project2.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, true, crmSetting.HubspotEnrichHeavy)

	// after marking enrich heavy as false, the project will be consider for re distribution
	status = store.GetStore().UpdateCRMSetting(project2.ID, model.HubspotEnrichHeavy(false, nil))
	assert.Equal(t, http.StatusAccepted, status)
	crmSetting, status = store.GetStore().GetCRMSetting(project2.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, false, crmSetting.HubspotEnrichHeavy)
	jobStatus, success = hubspot_enrich.RunHubspotProjectDistributer(config)
	assert.Equal(t, true, success)
	assert.Contains(t, jobStatus["light_projects"], project1.ID)
	assert.NotContains(t, jobStatus["light_projects"], project2.ID)
	assert.Contains(t, jobStatus["heavy_projects"], project2.ID) // project 2 is present as heavy project is marked a false
	assert.NotContains(t, jobStatus["heavy_projects"], project1.ID)
	crmSetting, status = store.GetStore().GetCRMSetting(project1.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, false, crmSetting.HubspotEnrichHeavy)
	crmSetting, status = store.GetStore().GetCRMSetting(project2.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, true, crmSetting.HubspotEnrichHeavy)

}

func TestHubspotDateTimezone(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	contactCreatedDate := time.Now().AddDate(0, 0, -5)
	contactCreatedDateTimestampMs := contactCreatedDate.Unix()
	contact := IntHubspot.Contact{
		Vid: 1,
		Properties: map[string]IntHubspot.Property{
			"createdate":       {Value: fmt.Sprintf("%d", contactCreatedDateTimestampMs*1000)},
			"lastmodifieddate": {Value: fmt.Sprintf("%d", contactCreatedDateTimestampMs*1000)},
			"lifecyclestage":   {Value: "lead"},
			"Workflow":         {Value: "A"},
			"date":             {Value: "1651363200000"}, // May 1, 2022 GMT daylight saving
		},
		IdentityProfiles: []IntHubspot.ContactIdentityProfile{
			{
				[]IntHubspot.ContactIdentity{
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
	contactDocument1 := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}
	processDocuments := []*model.HubspotDocument{&contactDocument1}
	contact.Vid = 2
	contact.Properties["date"] = IntHubspot.Property{Value: "1646179200000"} // America/Vancouver non daylight saving
	contact.Properties["lastmodifieddate"] = IntHubspot.Property{Value: fmt.Sprintf("%d", contactCreatedDate.Add(1*time.Minute).Unix()*1000+10)}

	enJSON, err = json.Marshal(contact)
	assert.Nil(t, err)
	contactPJson2 := postgres.Jsonb{json.RawMessage(enJSON)}
	contactDocument2 := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson2,
	}
	processDocuments = append(processDocuments, &contactDocument2)

	status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeContact, processDocuments, 2)
	assert.Equal(t, http.StatusCreated, status)
	dateProperties := map[int]*map[string]bool{
		model.HubspotDocumentTypeContact: {
			"date": true,
		},
	}

	enrichStatus, _ := IntHubspot.Sync(project.ID, 2, time.Now().UTC().Unix(), dateProperties, "America/Vancouver", 50)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}
	documents, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"1", "2"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	assert.Equal(t, http.StatusFound, status)
	latestDocument := make(map[string]*model.HubspotDocument)
	for i := range documents {
		if latestDocument[documents[i].ID] == nil {
			latestDocument[documents[i].ID] = &documents[i]
			continue
		}

		if latestDocument[documents[i].ID].Timestamp < documents[i].Timestamp {
			latestDocument[documents[i].ID] = &documents[i]
			continue
		}
	}

	for id, timeZoneTimeStamp := range map[string]string{
		"1": "1651388400", // midnight May 1, 2022  America/Vancouver daylight saving on
		"2": "1646208000", // midnight March 2, 2022  America/Vancouver daylight saving off
	} {
		document := latestDocument[id]
		event, status := store.GetStore().GetEvent(project.ID, document.UserId, document.SyncId)
		assert.Equal(t, http.StatusFound, status)
		var properties map[string]interface{}
		err = json.Unmarshal(event.UserProperties.RawMessage, &properties)
		assert.Nil(t, err)
		enDateKey := model.GetCRMEnrichPropertyKeyByType(U.CRM_SOURCE_NAME_HUBSPOT, model.HubspotDocumentTypeNameContact, "date")
		assert.Equal(t, timeZoneTimeStamp, U.GetPropertyValueAsString(properties[enDateKey]), fmt.Sprintf("Document id %s", id))
		// validate datetime property
		enCreatedAtKey := model.GetCRMEnrichPropertyKeyByType(U.CRM_SOURCE_NAME_HUBSPOT, model.HubspotDocumentTypeNameContact, "createdate")
		assert.Equal(t, U.GetPropertyValueAsString(contactCreatedDateTimestampMs), U.GetPropertyValueAsString(properties[enCreatedAtKey]), fmt.Sprintf("Document id %s", id))
	}
}

func TestHubspotLimitProcessing(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	contactCreatedDate := time.Now().AddDate(0, 0, -5)
	contactCreatedDateTimestampMs := U.GetPropertyValueAsString(contactCreatedDate.Unix() * 1000)
	contact := IntHubspot.Contact{
		Vid: 1,
		Properties: map[string]IntHubspot.Property{
			"createdate":       {Value: contactCreatedDateTimestampMs},
			"lastmodifieddate": {Value: contactCreatedDateTimestampMs},
			"lifecyclestage":   {Value: "lead"},
		},
		IdentityProfiles: []IntHubspot.ContactIdentityProfile{
			{
				[]IntHubspot.ContactIdentity{
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
	contactDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &postgres.Jsonb{json.RawMessage(enJSON)},
	}
	processDocuments := []*model.HubspotDocument{&contactDocument}
	contact.Vid = 2
	enJSON, err = json.Marshal(contact)
	assert.Nil(t, err)
	contactDocument2 := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &postgres.Jsonb{json.RawMessage(enJSON)},
	}
	processDocuments = append(processDocuments, &contactDocument2)
	contact.Vid = 3
	enJSON, err = json.Marshal(contact)
	assert.Nil(t, err)
	contactDocument3 := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &postgres.Jsonb{json.RawMessage(enJSON)},
	}
	processDocuments = append(processDocuments, &contactDocument3)

	status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeContact, processDocuments, 2)
	assert.Equal(t, http.StatusCreated, status)

	// limit processing to 1 contact
	enrichStatus, _ := IntHubspot.Sync(project.ID, 2, time.Now().UTC().Unix(), nil, "", 1)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}
	documents, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"1"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Len(t, documents, 2)
	assert.Equal(t, true, documents[0].Synced)
	assert.Equal(t, true, documents[1].Synced)
	documents, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"2"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Len(t, documents, 2)
	assert.Equal(t, true, documents[0].Synced)
	assert.Equal(t, true, documents[1].Synced)

	// 2nd contact not processed
	documents, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"3"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Len(t, documents, 2)
	assert.Equal(t, false, documents[0].Synced)
	assert.Equal(t, false, documents[1].Synced)

	// process 2nd contact
	enrichStatus, _ = IntHubspot.Sync(project.ID, 2, time.Now().UTC().Unix(), nil, "", 1)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}
	documents, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{"3"}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	assert.Len(t, documents, 2)
	assert.Equal(t, true, documents[0].Synced)
	assert.Equal(t, true, documents[1].Synced)
}

func TestHubspotDisableGroupUserPropertiesFromUserPropertiesCache(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	userID, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: time.Now().Unix() - 1000, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, status, http.StatusCreated)
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"
	w := ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "event2", "auto": true}`, userID)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	assert.Equal(t, http.StatusOK, w.Code)

	updateProperties := &postgres.Jsonb{json.RawMessage(`{"name":"user1","city":"bangalore","$hubspot_company_id":"company1", "$hubspot_contact_id":"contact1","$hubspot_deal_id":"deal1"}`)}
	_, status = store.GetStore().UpdateUserProperties(project.ID, userID, updateProperties, time.Now().Unix())
	assert.Equal(t, status, http.StatusAccepted)

	// execute DoRollUpSortedSet
	configs := make(map[string]interface{})
	configs["rollupLookback"] = 1
	event_user_cache.DoRollUpSortedSet(configs)

	w = sendGetUserProperties(project.ID, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)
	var responsePayload struct {
		Properties map[string][]string `json:"properties"`
	}

	jsonResponse, _ := ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &responsePayload)
	assert.Nil(t, err)

	categoryProperties := responsePayload.Properties
	assert.Contains(t, categoryProperties["categorical"], "name")
	assert.Contains(t, categoryProperties["categorical"], "city")
	// group properties should not be present in response of user properties
	for _, properties := range categoryProperties {
		assert.NotContains(t, properties, "$hubspot_company_id")
		assert.NotContains(t, properties, "$hubspot_deal_id")
	}

	user, status := store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, status)
	userProperties := make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &userProperties)
	assert.Nil(t, err)
	assert.Contains(t, userProperties, "$hubspot_company_id")
	assert.Contains(t, userProperties, "$hubspot_deal_id")
}

func TestHubspotIntegration(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	projectSetting, status := store.GetStore().GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, status)

	assert.Equal(t, false, *projectSetting.IntHubspot)
	assert.Equal(t, "", projectSetting.IntHubspotApiKey)
	assert.Equal(t, "", projectSetting.IntHubspotRefreshToken)
	// only enable api key based integration
	intHubspot := true
	_, status = store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{IntHubspotApiKey: "12-34", IntHubspot: &intHubspot},
	)
	assert.Equal(t, http.StatusAccepted, status)
	projectSetting2, status := store.GetStore().GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, status)
	// only required field should be updated
	portalID := 0
	projectSetting.IntHubspot = &intHubspot
	projectSetting.IntHubspotApiKey = "12-34"
	projectSetting.IntHubspotPortalID = &portalID
	projectSetting.UpdatedAt = projectSetting2.UpdatedAt
	assert.Equal(t, projectSetting, projectSetting2)

	// add refresh token to integration
	refreshToken := U.RandomString(5)
	_, status = store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{IntHubspotRefreshToken: refreshToken, IntHubspot: &intHubspot},
	)
	assert.Equal(t, http.StatusAccepted, status)
	projectSetting2, status = store.GetStore().GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, status)
	// only required field should be updated including previous
	projectSetting.IntHubspotRefreshToken = refreshToken
	projectSetting.UpdatedAt = projectSetting2.UpdatedAt
	assert.Equal(t, projectSetting, projectSetting2)
}
