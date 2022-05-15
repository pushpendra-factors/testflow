package tests

import (
	"encoding/json"
	enrichment "factors/crm_enrichment"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCRMCreateData(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	t.Run("CreateCRMUser", func(t *testing.T) {
		user1Properties := postgres.Jsonb{json.RawMessage(`{"name":"abc","city":"xyz"}`)}

		user1 := &model.CRMUser{
			ProjectID:  project.ID,
			Source:     U.CRM_SOURCE_HUBSPOT,
			Type:       1,
			ID:         "123",
			Properties: &user1Properties,
			Timestamp:  time.Now().Unix(),
		}
		status, err := store.GetStore().CreateCRMUser(user1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, user1.Action, model.CRMActionCreated)

		status, err = store.GetStore().CreateCRMUser(user1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, user1.Action, model.CRMActionUpdated)

		status, err = store.GetStore().CreateCRMUser(user1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, user1.Action, model.CRMActionUpdated)

		user1.Timestamp = user1.Timestamp + 100
		status, err = store.GetStore().CreateCRMUser(user1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, user1.Action, model.CRMActionUpdated)
	},
	)

	t.Run("CreateCRMGroup", func(t *testing.T) {
		user1Properties := postgres.Jsonb{json.RawMessage(`{"company":"company1","city":"xyz"}`)}

		group1 := &model.CRMGroup{
			ProjectID:  project.ID,
			Source:     U.CRM_SOURCE_HUBSPOT,
			Type:       1,
			ID:         "123",
			Properties: &user1Properties,
			Timestamp:  time.Now().Unix(),
		}
		status, err := store.GetStore().CreateCRMGroup(group1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, group1.Action, model.CRMActionCreated)

		status, err = store.GetStore().CreateCRMGroup(group1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, group1.Action, model.CRMActionUpdated)

		status, err = store.GetStore().CreateCRMGroup(group1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, group1.Action, model.CRMActionUpdated)

		group1.Timestamp = group1.Timestamp + 100
		status, err = store.GetStore().CreateCRMGroup(group1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, group1.Action, model.CRMActionUpdated)
	},
	)

	t.Run("CreateCRMRelationship", func(t *testing.T) {

		relationship := &model.CRMRelationship{
			ProjectID: project.ID,
			Source:    U.CRM_SOURCE_HUBSPOT,
			FromType:  1,
			FromID:    "123",
			ToType:    2,
			ToID:      "234",
			Timestamp: time.Now().Unix(),
		}
		status, err := store.GetStore().CreateCRMRelationship(relationship)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		status, err = store.GetStore().CreateCRMRelationship(relationship)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, status)
	},
	)

	t.Run("CreateCRMActivity", func(t *testing.T) {
		activityProperties := postgres.Jsonb{json.RawMessage(`{"name":"abc","clicked":"true"}`)}
		activity := &model.CRMActivity{
			ProjectID:          project.ID,
			Source:             U.CRM_SOURCE_HUBSPOT,
			Name:               "test1",
			Type:               1,
			ExternalActivityID: "123",
			ActorType:          1,
			ActorID:            "123",
			Properties:         &activityProperties,
			Timestamp:          time.Now().Unix(),
		}
		status, err := store.GetStore().CreateCRMActivity(activity)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		activity.ID = ""
		status, err = store.GetStore().CreateCRMActivity(activity)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, status)
		activity.Timestamp = activity.Timestamp + 100
		status, err = store.GetStore().CreateCRMActivity(activity)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
	},
	)
}

func TestCRMMarketoEnrichment(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	//record types
	typeUserLead := 1
	typeActivityProgramMember := 2
	// create crm user records
	leadUpdateTimestampProperty := "updated_at"
	leadTimestamp := time.Now().AddDate(0, 0, -1)
	leadJoinTimestamp := leadTimestamp
	leadProperties := fmt.Sprintf(`{"Name":"name1","city":"city1","%s":%d}`, leadUpdateTimestampProperty, leadTimestamp.Unix())
	user1Properties := postgres.Jsonb{json.RawMessage(leadProperties)}

	user1 := &model.CRMUser{
		ProjectID:  project.ID,
		Source:     U.CRM_SOURCE_MARKETO,
		Type:       typeUserLead,
		ID:         "lead1",
		Properties: &user1Properties,
		Timestamp:  leadTimestamp.Unix(),
	}
	status, err := store.GetStore().CreateCRMUser(user1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, user1.Action, model.CRMActionCreated)
	status, err = store.GetStore().CreateCRMUser(user1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, user1.Action, model.CRMActionUpdated)

	// create crm activity record
	activityUpdateTimestampProperty := "_fivetran_synced"
	activityTimestamp := leadTimestamp.Add(2 * time.Hour)
	externalActivityID := "activity1"
	activityProperties := fmt.Sprintf(`{"Name":"Click event","status":"Responded","%s":%d}`, activityUpdateTimestampProperty, activityTimestamp.Unix())
	activity1Properties := postgres.Jsonb{json.RawMessage(activityProperties)}
	activity1 := &model.CRMActivity{
		ProjectID:          project.ID,
		Source:             U.CRM_SOURCE_MARKETO,
		ExternalActivityID: externalActivityID,
		Type:               typeActivityProgramMember,
		Name:               "program_membership_created",
		ActorType:          typeUserLead,
		ActorID:            "lead1",
		Properties:         &activity1Properties,
		Timestamp:          activityTimestamp.Unix(),
	}

	status, err = store.GetStore().CreateCRMActivity(activity1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	activity1.ID = ""
	status, err = store.GetStore().CreateCRMActivity(activity1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusConflict, status)

	// same actor but different external id should be allowed
	externalActivityID2 := "activity2"
	activityProperties = fmt.Sprintf(`{"Name":"Click event","status":"Responded","%s":%d}`, activityUpdateTimestampProperty, activityTimestamp.Unix())
	activity2Properties := postgres.Jsonb{json.RawMessage(activityProperties)}
	activity2 := &model.CRMActivity{
		ProjectID:          project.ID,
		Source:             U.CRM_SOURCE_MARKETO,
		ExternalActivityID: externalActivityID2,
		Type:               typeActivityProgramMember,
		Name:               "program_membership_created",
		ActorType:          typeUserLead,
		ActorID:            "lead1",
		Properties:         &activity2Properties,
		Timestamp:          activityTimestamp.Unix(),
	}

	status, err = store.GetStore().CreateCRMActivity(activity2)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	activity2.ID = ""
	status, err = store.GetStore().CreateCRMActivity(activity2)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusConflict, status)

	sourceObjectTypeAndAlias := map[int]string{
		typeUserLead:              "lead",
		typeActivityProgramMember: "program_membership",
	}

	// create source config for mapping record type to alias
	userTypes := map[int]bool{
		typeUserLead: true,
	}
	activityTypes := map[int]bool{
		typeActivityProgramMember: true,
	}

	sourceConfig, err := enrichment.NewCRMEnrichmentConfig(U.CRM_SOURCE_NAME_MARKETO, sourceObjectTypeAndAlias, userTypes, nil, activityTypes)
	assert.Nil(t, err)
	enrichStatus := enrichment.Enrich(project.ID, sourceConfig)

	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	sourceAlias, err := model.GetCRMSourceByAliasName(U.CRM_SOURCE_NAME_MARKETO)
	assert.Nil(t, err)
	crmUser, status := store.GetStore().GetCRMUserByTypeAndAction(project.ID, sourceAlias, "lead1", typeUserLead, model.CRMActionCreated)
	assert.Equal(t, http.StatusFound, status)
	assert.NotEqual(t, "", crmUser.UserID)
	createdUserID := crmUser.UserID
	crmUser, status = store.GetStore().GetCRMUserByTypeAndAction(project.ID, sourceAlias, "lead1", typeUserLead, model.CRMActionUpdated)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, createdUserID, crmUser.UserID)

	// validate activity event
	eventName, status := store.GetStore().GetEventName(U.EVENT_NAME_MARKETO_PROGRAM_MEMBERSHIP_CREATED, project.ID)
	assert.Equal(t, http.StatusFound, status)
	events, status := store.GetStore().GetUserEventsByEventNameId(project.ID, createdUserID, eventName.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, events, 2)
	user, status := store.GetStore().GetUser(project.ID, events[0].UserId)
	assert.Equal(t, http.StatusFound, status)
	// activities shouldn't affect the user properties update timestamp
	assert.Equal(t, leadJoinTimestamp.Unix(), user.PropertiesUpdatedTimestamp)
	properties := make(map[string]interface{})
	err = json.Unmarshal(events[0].Properties.RawMessage, &properties)
	assert.Nil(t, err)

	assert.Equal(t, "Click event", properties["$marketo_program_membership_name"])
	assert.Equal(t, "Responded", properties["$marketo_program_membership_status"])

	// validate user event
	eventName, status = store.GetStore().GetEventName(U.EVENT_NAME_MARKETO_LEAD_CREATED, project.ID)
	assert.Equal(t, http.StatusFound, status)
	events, status = store.GetStore().GetUserEventsByEventNameId(project.ID, createdUserID, eventName.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, events, 1)
	assert.Equal(t, events[0].EventNameId, eventName.ID)
	assert.Equal(t, leadTimestamp.Unix(), events[0].Timestamp)
	properties = make(map[string]interface{})
	err = json.Unmarshal(events[0].Properties.RawMessage, &properties)
	assert.Nil(t, err)

	assert.Equal(t, "name1", properties["$marketo_lead_name"])
	assert.Equal(t, "city1", properties["$marketo_lead_city"])

	eventName, status = store.GetStore().GetEventName(U.EVENT_NAME_MARKETO_LEAD_UPDATED, project.ID)
	assert.Equal(t, http.StatusFound, status)
	events, status = store.GetStore().GetUserEventsByEventNameId(project.ID, createdUserID, eventName.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, events, 1)
	assert.Equal(t, events[0].EventNameId, eventName.ID)
	assert.Equal(t, leadTimestamp.Unix(), events[0].Timestamp)
	properties = make(map[string]interface{})
	err = json.Unmarshal(events[0].Properties.RawMessage, &properties)
	assert.Nil(t, err)

	assert.Equal(t, "name1", properties["$marketo_lead_name"])
	assert.Equal(t, "city1", properties["$marketo_lead_city"])

	//created crm user with email
	lead2Timestmap := time.Now().AddDate(0, 0, -1)
	lead2JointTimestamp := lead2Timestmap
	leadProperties = fmt.Sprintf(`{"Name":"name2","city":"city2","%s":%d}`, leadUpdateTimestampProperty, lead2Timestmap.Unix())
	user2Properties := postgres.Jsonb{json.RawMessage(leadProperties)}
	user2 := &model.CRMUser{
		ProjectID:  project.ID,
		Source:     U.CRM_SOURCE_MARKETO,
		Type:       typeUserLead,
		ID:         "lead2",
		Email:      "abc2@abc.com",
		Properties: &user2Properties,
		Timestamp:  lead2Timestmap.Unix(),
	}
	status, err = store.GetStore().CreateCRMUser(user2)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, model.CRMActionCreated, user2.Action)
	status, err = store.GetStore().CreateCRMUser(user2)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, user2.Action, model.CRMActionUpdated)

	enrichStatus = enrichment.Enrich(project.ID, sourceConfig)

	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	// validate 2nd user
	eventName, status = store.GetStore().GetEventName(U.EVENT_NAME_MARKETO_LEAD_CREATED, project.ID)
	assert.Equal(t, http.StatusFound, status)
	crmUser, status = store.GetStore().GetCRMUserByTypeAndAction(project.ID, sourceAlias, "lead2", typeUserLead, model.CRMActionCreated)
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, crmUser.UserID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, "abc2@abc.com", user.CustomerUserId) // validate email association
	properties = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properties)
	assert.Nil(t, err)

	assert.Equal(t, "name2", properties["$marketo_lead_name"])
	assert.Equal(t, "city2", properties["$marketo_lead_city"])
	events, status = store.GetStore().GetUserEventsByEventNameId(project.ID, crmUser.UserID, eventName.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, events, 1)

	eventName, status = store.GetStore().GetEventName(U.EVENT_NAME_MARKETO_LEAD_UPDATED, project.ID)
	assert.Equal(t, http.StatusFound, status)
	events, status = store.GetStore().GetUserEventsByEventNameId(project.ID, crmUser.UserID, eventName.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, events, 1)

	// update on user2
	lead2Timestmap = lead2Timestmap.Add(1 * time.Hour)
	leadProperties = fmt.Sprintf(`{"Name":"name3","city":"city2","%s":%d}`, leadUpdateTimestampProperty, lead2Timestmap.Unix())
	user2Properties = postgres.Jsonb{json.RawMessage(leadProperties)}
	user2 = &model.CRMUser{
		ProjectID:  project.ID,
		Source:     U.CRM_SOURCE_MARKETO,
		Type:       typeUserLead,
		ID:         "lead2",
		Email:      "abc2@abc.com",
		Properties: &user2Properties,
		Timestamp:  lead2Timestmap.Unix(),
	}
	status, err = store.GetStore().CreateCRMUser(user2)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, model.CRMActionUpdated, user2.Action)
	status, err = store.GetStore().CreateCRMUser(user2)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusConflict, status)
	enrichStatus = enrichment.Enrich(project.ID, sourceConfig)

	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}
	crmUser, status = store.GetStore().GetCRMUserByTypeAndAction(project.ID, sourceAlias, "lead2", typeUserLead, model.CRMActionCreated)
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, crmUser.UserID)
	assert.Equal(t, http.StatusFound, status)
	// crm user updates shouldn't affect the user properties update timestamp
	assert.Equal(t, user.PropertiesUpdatedTimestamp, lead2JointTimestamp.Unix())
	properties = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properties)
	assert.Nil(t, err)
	assert.Equal(t, "name3", properties["$marketo_lead_name"])
	assert.Equal(t, "city2", properties["$marketo_lead_city"])

	// update user2 crm properties if user update_timestamp is ahead of crm property timestamp
	crmUser, status = store.GetStore().GetCRMUserByTypeAndAction(project.ID, sourceAlias, "lead2", typeUserLead, model.CRMActionCreated)
	assert.Equal(t, http.StatusFound, status)
	userProperties := &postgres.Jsonb{RawMessage: []byte(`{"name":"user1","city":"bangalore"}`)}
	propertiesUpdateTimestamp := time.Now().Unix()
	_, status = store.GetStore().UpdateUserProperties(project.ID, crmUser.UserID, userProperties, propertiesUpdateTimestamp)
	assert.Equal(t, http.StatusAccepted, status)
	leadProperties = fmt.Sprintf(`{"Name":"name4","city":"city2","%s":%d}`, leadUpdateTimestampProperty, lead2Timestmap.Unix()+100)
	user2Properties = postgres.Jsonb{json.RawMessage(leadProperties)}
	user2 = &model.CRMUser{
		ProjectID:  project.ID,
		Source:     U.CRM_SOURCE_MARKETO,
		Type:       typeUserLead,
		ID:         "lead2",
		Email:      "abc2@abc.com",
		Properties: &user2Properties,
		Timestamp:  lead2Timestmap.Unix() + 100,
	}
	status, err = store.GetStore().CreateCRMUser(user2)
	assert.Nil(t, err)
	enrichStatus = enrichment.Enrich(project.ID, sourceConfig)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}
	user, status = store.GetStore().GetUser(project.ID, crmUser.UserID)
	assert.Equal(t, http.StatusFound, status)
	// crm user updates shouldn't affect the user properties update timestamp
	assert.Equal(t, user.PropertiesUpdatedTimestamp, propertiesUpdateTimestamp)
	properties = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properties)
	assert.Nil(t, err)
	assert.Equal(t, "name4", properties["$marketo_lead_name"])
	assert.Equal(t, "city2", properties["$marketo_lead_city"])
}

func TestCRMEmptyPropertiesUpdated(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	leadTimestmap := time.Now().AddDate(0, 0, -1)
	leadUpdateTimestampProperty := "updated_at"
	leadProperties := fmt.Sprintf(`{"Name":"name2","city":"city2","Stage":"lead","%s":%d}`, leadUpdateTimestampProperty, leadTimestmap.Unix())
	leadEnProperties := postgres.Jsonb{json.RawMessage(leadProperties)}

	typeUserLead := 1
	crmUser := &model.CRMUser{
		ProjectID:  project.ID,
		Source:     U.CRM_SOURCE_MARKETO,
		Type:       typeUserLead,
		ID:         "lead1",
		Email:      "abc2@abc.com",
		Properties: &leadEnProperties,
		Timestamp:  leadTimestmap.Unix(),
	}
	status, err := store.GetStore().CreateCRMUser(crmUser)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, crmUser.Action, model.CRMActionCreated)

	userTypes := map[int]bool{
		typeUserLead: true,
	}
	sourceObjectTypeAndAlias := map[int]string{
		typeUserLead: "lead",
	}

	sourceConfig, err := enrichment.NewCRMEnrichmentConfig(U.CRM_SOURCE_NAME_MARKETO, sourceObjectTypeAndAlias, userTypes, nil, nil)
	assert.Nil(t, err)

	enrichStatus := enrichment.Enrich(project.ID, sourceConfig)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	crmUser, status = store.GetStore().GetCRMUserByTypeAndAction(project.ID, U.CRM_SOURCE_MARKETO, "lead1", typeUserLead, model.CRMActionCreated)
	assert.Equal(t, http.StatusFound, status)
	assert.NotEqual(t, "", crmUser.UserID)

	// validte properties without empty
	user, status := store.GetStore().GetUser(project.ID, crmUser.UserID)
	assert.Equal(t, http.StatusFound, status)
	event, status := store.GetStore().GetEventById(project.ID, crmUser.SyncID, crmUser.UserID)
	assert.Equal(t, http.StatusFound, status)
	var userProperties map[string]interface{}
	var eventProperties map[string]interface{}
	var eventUserProperties map[string]interface{}
	json.Unmarshal(user.Properties.RawMessage, &userProperties)
	json.Unmarshal(event.Properties.RawMessage, &eventProperties)
	json.Unmarshal(event.UserProperties.RawMessage, &eventUserProperties)
	for key, value := range map[string]interface{}{"name": "name2", "city": "city2", "stage": "lead"} {
		enKey := model.GetCRMEnrichPropertyKeyByType(U.CRM_SOURCE_NAME_MARKETO,
			"lead", key)
		assert.Equal(t, value, userProperties[enKey])
		assert.Equal(t, value, eventProperties[enKey])
		assert.Equal(t, value, eventUserProperties[enKey])
	}

	// update with empty and null value. Both should be converted to empty string and overridden
	leadEnProperties = postgres.Jsonb{json.RawMessage(`{"city":"","Stage":null}`)}

	crmUser = &model.CRMUser{
		ProjectID:  project.ID,
		Source:     U.CRM_SOURCE_MARKETO,
		Type:       typeUserLead,
		ID:         "lead1",
		Email:      "abc2@abc.com",
		Properties: &leadEnProperties,
		Timestamp:  leadTimestmap.Unix(),
	}
	status, err = store.GetStore().CreateCRMUser(crmUser)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, crmUser.Action, model.CRMActionUpdated)

	enrichStatus = enrichment.Enrich(project.ID, sourceConfig)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	crmUser, status = store.GetStore().GetCRMUserByTypeAndAction(project.ID, U.CRM_SOURCE_MARKETO, "lead1", typeUserLead, model.CRMActionUpdated)
	assert.Equal(t, http.StatusFound, status)
	assert.NotEqual(t, "", crmUser.UserID)
	user, status = store.GetStore().GetUser(project.ID, crmUser.UserID)
	assert.Equal(t, http.StatusFound, status)
	event, status = store.GetStore().GetEventById(project.ID, crmUser.SyncID, crmUser.UserID)
	assert.Equal(t, http.StatusFound, status)

	json.Unmarshal(user.Properties.RawMessage, &userProperties)
	json.Unmarshal(event.Properties.RawMessage, &eventProperties)
	json.Unmarshal(event.UserProperties.RawMessage, &eventUserProperties)
	for key, value := range map[string]interface{}{"name": "name2", "city": "", "stage": ""} {
		enKey := model.GetCRMEnrichPropertyKeyByType(U.CRM_SOURCE_NAME_MARKETO,
			"lead", key)
		assert.Equal(t, value, userProperties[enKey], enKey)
		assert.Equal(t, value, eventProperties[enKey], enKey)
		assert.Equal(t, value, eventUserProperties[enKey], enKey)
	}
}

func TestCRMPropertiesSync(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	userTypeMap := map[int]bool{
		1: true,
	}

	activityTypeMap := map[int]bool{
		2: true,
	}

	typeAlias := map[int]string{
		1: "lead",
		2: "program_member",
	}

	property1 := model.CRMProperty{
		ProjectID:        project.ID,
		Source:           U.CRM_SOURCE_MARKETO,
		Type:             1,
		Name:             "created_at",
		ExternalDataType: "datetime",
		MappedDataType:   U.PropertyTypeDateTime,
		Label:            "Created At",
	}

	status, err := store.GetStore().CreateCRMProperties(&property1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	status, err = store.GetStore().CreateCRMProperties(&property1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusConflict, status)

	leadTimestamp := time.Now().AddDate(0, 0, -1)
	leadProperties := fmt.Sprintf(`{"Name":"name1","city":"city1","%s":%d}`, "updated_at", leadTimestamp.Unix())
	user1Properties := postgres.Jsonb{json.RawMessage(leadProperties)}

	user1 := &model.CRMUser{
		ProjectID:  project.ID,
		Source:     U.CRM_SOURCE_MARKETO,
		Type:       1,
		ID:         "lead1",
		Properties: &user1Properties,
		Timestamp:  leadTimestamp.Unix(),
	}
	status, err = store.GetStore().CreateCRMUser(user1)
	assert.Nil(t, err)

	// create activity record for adding property details
	activityTimestamp := time.Now().Add(2 * time.Hour)
	activityUpdateTimestampProperty := "_fivetran_synced"
	activityProperties := fmt.Sprintf(`{"Name":"Click event","status":"Responded","%s":%d}`, activityUpdateTimestampProperty, activityTimestamp.Unix())
	activity1 := &model.CRMActivity{
		ProjectID:          project.ID,
		Source:             U.CRM_SOURCE_MARKETO,
		ExternalActivityID: "123",
		Type:               2,
		Name:               "program_membership_created",
		ActorType:          1,
		ActorID:            "lead1",
		Properties:         &postgres.Jsonb{[]byte(activityProperties)},
		Timestamp:          activityTimestamp.Unix(),
	}

	status, err = store.GetStore().CreateCRMActivity(activity1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)

	activityProperty1 := model.CRMProperty{
		ProjectID:        project.ID,
		Source:           U.CRM_SOURCE_MARKETO,
		Type:             2,
		Name:             "created_at",
		ExternalDataType: "datetime",
		MappedDataType:   U.PropertyTypeDateTime,
		Label:            "Created At",
	}

	status, err = store.GetStore().CreateCRMProperties(&activityProperty1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	status, err = store.GetStore().CreateCRMProperties(&activityProperty1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusConflict, status)

	sourceConfig, err := enrichment.NewCRMEnrichmentConfig(U.CRM_SOURCE_NAME_MARKETO, typeAlias, userTypeMap, nil, activityTypeMap)
	assert.Nil(t, err)
	enrichStatus := enrichment.Enrich(project.ID, sourceConfig)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	enrichStatus = enrichment.SyncProperties(project.ID, sourceConfig)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}
	propertyDetails, status := store.GetStore().GetAllPropertyDetailsByProjectID(project.ID, "", true)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, *propertyDetails, 1)
	assert.Equal(t, U.PropertyTypeDateTime, (*propertyDetails)[model.GetCRMEnrichPropertyKeyByType(U.CRM_SOURCE_NAME_MARKETO, "lead", "created_at")])

	status, displayNames := store.GetStore().GetDisplayNamesForObjectEntities(project.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, "marketo lead Created At", displayNames[model.GetCRMEnrichPropertyKeyByType(U.CRM_SOURCE_NAME_MARKETO, "lead", "created_at")])

	status, displayNames = store.GetStore().GetDisplayNamesForObjectEntities(project.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, "marketo program_member Created At", displayNames[model.GetCRMEnrichPropertyKeyByType(U.CRM_SOURCE_NAME_MARKETO, "program_member", "created_at")])

	// change data type should be allowed to insert
	property1.ExternalDataType = "int"
	status, err = store.GetStore().CreateCRMProperties(&property1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	property1.ExternalDataType = "datetime"
	status, err = store.GetStore().CreateCRMProperties(&property1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	status, err = store.GetStore().CreateCRMProperties(&property1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusConflict, status)
}
