package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	IntSalesforce "factors/integration/salesforce"
	"factors/model/model"
	"factors/model/store"
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

func TestCreateSalesforceDocument(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	refreshToken := U.RandomLowerAphaNumString(5)
	instancURL := U.RandomLowerAphaNumString(5)
	errCode := store.GetStore().UpdateAgentIntSalesforce(agent.UUID,
		refreshToken,
		instancURL,
	)
	assert.Equal(t, http.StatusAccepted, errCode)

	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntSalesforceEnabledAgentUUID: &agent.UUID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	//should return list of supported doc type with timestamp 0
	syncInfo, status := store.GetStore().GetSalesforceSyncInfo()
	assert.Equal(t, http.StatusFound, status)

	assert.Equal(t, refreshToken, syncInfo.ProjectSettings[project.ID].RefreshToken)
	assert.Equal(t, instancURL, syncInfo.ProjectSettings[project.ID].InstanceURL)

	assert.Contains(t, syncInfo.LastSyncInfo[project.ID], model.SalesforceDocumentTypeNameContact)
	assert.Equal(t, int64(0), syncInfo.LastSyncInfo[project.ID][model.SalesforceDocumentTypeNameContact])
	assert.Contains(t, syncInfo.LastSyncInfo[project.ID], model.SalesforceDocumentTypeNameAccount)
	assert.Equal(t, int64(0), syncInfo.LastSyncInfo[project.ID][model.SalesforceDocumentTypeNameAccount])
	assert.Contains(t, syncInfo.LastSyncInfo[project.ID], model.SalesforceDocumentTypeNameLead)
	assert.Equal(t, int64(0), syncInfo.LastSyncInfo[project.ID][model.SalesforceDocumentTypeNameLead])

	//should not contain opportunity by default
	assert.NotContains(t, syncInfo.LastSyncInfo[project.ID], model.SalesforceDocumentTypeNameOpportunity)

	contactID := U.RandomLowerAphaNumString(5)
	name := U.RandomLowerAphaNumString(5)

	createdDate := time.Now()

	// salesforce record with created == updated
	jsonData := fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentTimeLayout))
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	syncInfo, status = store.GetStore().GetSalesforceSyncInfo()

	//should return latest timestamp from the databse
	assert.Equal(t, createdDate.Unix(), syncInfo.LastSyncInfo[project.ID][model.SalesforceDocumentTypeNameContact])

	//should return error on duplicate
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusConflict, status)

	//enrich job, create contact created and contact updated event
	enrichStatus := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)
	assert.Equal(t, "success", enrichStatus[2].Status)

	eventNameCreated := fmt.Sprintf("$sf_%s_created", salesforceDocument.TypeAlias)
	eventNameUpdate := fmt.Sprintf("$sf_%s_updated", salesforceDocument.TypeAlias)
	query := model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       eventNameCreated,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       eventNameUpdate,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassInsights,

		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	// test using query
	result, errCode, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, eventNameCreated, result.Rows[0][0])
	assert.Equal(t, int64(1), result.Rows[0][1])
	assert.Equal(t, eventNameUpdate, result.Rows[1][0])
	assert.Equal(t, int64(1), result.Rows[1][1])

	query = model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       eventNameCreated,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       eventNameUpdate,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassInsights,

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	// test using query
	result, errCode, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, int64(1), result.Rows[0][0])

	/*
		salesforce record1 with createdDate != updatedDate
		salesforce record2 with createdDate != updatedDate
		both same id
	*/
	contactID = U.RandomLowerAphaNumString(5)
	name = U.RandomLowerAphaNumString(5)
	createdDate = createdDate.AddDate(0, 0, -10)
	updatedDate := createdDate.AddDate(0, 0, 1)

	// salesforce record1 with created != updated
	jsonData = fmt.Sprintf(`{"Id":"%s", "name":"%s","MobilePhone":1234567890,"CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentTimeLayout), updatedDate.UTC().Format(model.SalesforceDocumentTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	// salesforce record2 with created != updated same user
	updatedDate = updatedDate.AddDate(0, 0, 1)
	jsonData = fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentTimeLayout), updatedDate.Format(model.SalesforceDocumentTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	//should return conflict on duplicate
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusConflict, status)

	//enrich job, create contact created and contact updated event
	enrichStatus = IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)
	assert.Equal(t, "success", enrichStatus[2].Status)

	// query count of unique users
	query = model.Query{
		From: createdDate.Unix() - 500,
		To:   updatedDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       eventNameCreated,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       eventNameUpdate,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassInsights,

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	// test using query
	result, errCode, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, int64(1), result.Rows[0][0])

	// query count of events
	query = model.Query{
		From: createdDate.Unix() - 500,
		To:   updatedDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       eventNameCreated,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       eventNameUpdate,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassInsights,

		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	// test using query
	result, errCode, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, eventNameCreated, result.Rows[0][0])
	assert.Equal(t, int64(1), result.Rows[0][1])
	assert.Equal(t, eventNameUpdate, result.Rows[1][0])
	assert.Equal(t, int64(3), result.Rows[1][1])

	query.GroupByProperties = []model.QueryGroupByProperty{
		{
			Entity:    model.PropertyEntityUser,
			Property:  "$user_id",
			EventName: model.UserPropertyGroupByPresent,
		},
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, eventNameCreated, result.Rows[0][0])
	assert.Equal(t, "1234567890", result.Rows[0][1])
	assert.Equal(t, eventNameUpdate, result.Rows[1][0])
	assert.Equal(t, "1234567890", result.Rows[1][1])
}

func TestSalesforceCRMSmartEvent(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	contactID := U.RandomLowerAphaNumString(5)
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
	updatedDate := createdAt.AddDate(0, 0, -11)
	propertyDay := "Sunday"
	jsonData := fmt.Sprintf(`{"Id":"%s", "day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, propertyDay, createdAt.UTC().Format(model.SalesforceDocumentTimeLayout), updatedDate.Format(model.SalesforceDocumentTimeLayout))
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().UpdateSalesforceDocumentAsSynced(project.ID, salesforceDocument, "", userID3)
	assert.Equal(t, http.StatusAccepted, status)

	createdAt = time.Now().AddDate(0, 0, -11)
	updatedDate = createdAt.AddDate(0, 0, -10)
	propertyDay = "Monday"
	jsonData = fmt.Sprintf(`{"Id":"%s", "day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, propertyDay, createdAt.UTC().Format(model.SalesforceDocumentTimeLayout), updatedDate.Format(model.SalesforceDocumentTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().UpdateSalesforceDocumentAsSynced(project.ID, salesforceDocument, "", userID1)
	assert.Equal(t, http.StatusAccepted, status)

	createdAt = time.Now().AddDate(0, 0, -11)
	updatedDate = createdAt.AddDate(0, 0, -9)
	propertyDay = "Tuesday"
	jsonData = fmt.Sprintf(`{"Id":"%s", "day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, propertyDay, createdAt.UTC().Format(model.SalesforceDocumentTimeLayout), updatedDate.Format(model.SalesforceDocumentTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().UpdateSalesforceDocumentAsSynced(project.ID, salesforceDocument, "", userID2)
	assert.Equal(t, http.StatusAccepted, status)

	createdAt = time.Now().AddDate(0, 0, -11)
	updatedDate = createdAt.AddDate(0, 0, -8)
	propertyDay = "Wednesday"
	jsonData = fmt.Sprintf(`{"Id":"%s", "day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, propertyDay, createdAt.UTC().Format(model.SalesforceDocumentTimeLayout), updatedDate.Format(model.SalesforceDocumentTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	prevDoc, status := store.GetStore().GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(project.ID, cuid, userID3, salesforceDocument.Type)
	assert.Equal(t, http.StatusFound, status)
	_, prevProperties, err := IntSalesforce.GetSalesforceDocumentProperties(project.ID, prevDoc)
	assert.Nil(t, err)
	assert.Equal(t, "Tuesday", (*prevProperties)["day"])

	prevDoc, status = store.GetStore().GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(project.ID, "", userID3, salesforceDocument.Type)
	assert.Equal(t, http.StatusFound, status)
	_, prevProperties, err = IntSalesforce.GetSalesforceDocumentProperties(project.ID, prevDoc)
	assert.Nil(t, err)
	assert.Equal(t, "Sunday", (*prevProperties)["day"])

	filter := model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "day",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "Saturday",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "Tuesday",
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
	currentProperties["day"] = "Saturday"
	smartEvent, _, ok := IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "test", cuid, userID3, salesforceDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)
	assert.Equal(t, "test", smartEvent.Name)
	assert.Equal(t, "Tuesday", smartEvent.Properties["$prev_salesforce_contact_day"])
	assert.Equal(t, "Saturday", smartEvent.Properties["$curr_salesforce_contact_day"])

	smartEvent, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "test", "", userID3, salesforceDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, false, ok)

	filter = model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "day",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "Saturday",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "Sunday",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	smartEvent, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "test", "", userID3, salesforceDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)
	assert.Equal(t, "test", smartEvent.Name)
	assert.Equal(t, "Sunday", smartEvent.Properties["$prev_salesforce_contact_day"])
	assert.Equal(t, "Saturday", smartEvent.Properties["$curr_salesforce_contact_day"])
}

func TestSalesforceLastSyncedDocument(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	contactID1 := U.RandomLowerAphaNumString(5)
	contactID2 := U.RandomLowerAphaNumString(5)
	contactID3 := U.RandomLowerAphaNumString(5)
	contactID4 := U.RandomLowerAphaNumString(5)
	contactID5 := U.RandomLowerAphaNumString(5)
	contactID6 := U.RandomLowerAphaNumString(5)

	userID1 := U.RandomLowerAphaNumString(5)
	userID2 := U.RandomLowerAphaNumString(5)
	userID3 := U.RandomLowerAphaNumString(5)
	userID4 := U.RandomLowerAphaNumString(5)
	userID5 := U.RandomLowerAphaNumString(5)
	userID6 := U.RandomLowerAphaNumString(5)
	_, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID1})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID2})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID3})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID4})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID5})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID6})
	assert.Equal(t, http.StatusCreated, status)

	userIDs := []string{userID1, userID2, userID3, userID4, userID5, userID6}
	contactIDs := []string{contactID1, contactID2, contactID3, contactID4, contactID5, contactID6}
	characters := []string{"A", "B", "C", "D", "E", "F"}
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Friday", "Saturday"}
	var createdAt time.Time
	var updatedDate time.Time

	/*
		Summary Of synced test document
		U1(day="Sunday", character="A", type = contact)  -> U1(day="Saturday", character="G", type = contact) -> U1(day="Friday", character="H", type = lead)
		U2(day="Monday", character="B",type = contact)
		U3(day="Tuesday", character="C",type = contact)
		U4(day="Wednesday", character="D",type = contact)
		U5(day="Friday", character="E",type = contact)
		U6(day="Saturday", character="F",type = contact)
	*/
	for i := 0; i < 6; i++ {
		createdAt = time.Now().AddDate(0, 0, -20+i)
		updatedDate = createdAt.AddDate(0, 0, -20+i)
		jsonData := fmt.Sprintf(`{"Id":"%s", "character":"%s","day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactIDs[i], characters[i], days[i], createdAt.UTC().Format(model.SalesforceDocumentTimeLayout), updatedDate.Format(model.SalesforceDocumentTimeLayout))
		salesforceDocument := &model.SalesforceDocument{
			ProjectID: project.ID,
			TypeAlias: model.SalesforceDocumentTypeNameContact,
			Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
		}
		status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
		assert.Equal(t, http.StatusCreated, status)
		status = store.GetStore().UpdateSalesforceDocumentAsSynced(project.ID, salesforceDocument, "", userIDs[i])
		assert.Equal(t, http.StatusAccepted, status)
	}

	updatedDate = updatedDate.AddDate(0, 0, -1)
	jsonData := fmt.Sprintf(`{"Id":"%s", "character":"%s","day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID1, "G", "Saturday", createdAt.UTC().Format(model.SalesforceDocumentTimeLayout), updatedDate.Format(model.SalesforceDocumentTimeLayout))
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().UpdateSalesforceDocumentAsSynced(project.ID, salesforceDocument, "", userID1)
	assert.Equal(t, http.StatusAccepted, status)

	updatedDate = updatedDate.AddDate(0, 0, -1)
	leadID1 := U.RandomLowerAphaNumString(5)
	jsonData = fmt.Sprintf(`{"Id":"%s", "character":"%s","day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, leadID1, "H", "Friday", createdAt.UTC().Format(model.SalesforceDocumentTimeLayout), updatedDate.Format(model.SalesforceDocumentTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameLead,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().UpdateSalesforceDocumentAsSynced(project.ID, salesforceDocument, "", userID1)
	assert.Equal(t, http.StatusAccepted, status)

	/*
		Last synced document of U1 and type contact
		U1(day="Saturday", character="G", type = contact)
	*/
	prevDoc, status := store.GetStore().GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(project.ID, "", userID1, model.SalesforceDocumentTypeContact)
	assert.Equal(t, http.StatusFound, status)
	_, prevProperties, err := IntSalesforce.GetSalesforceDocumentProperties(project.ID, prevDoc)
	assert.Nil(t, err)
	assert.Equal(t, "G", (*prevProperties)["character"])
	assert.Equal(t, "Saturday", (*prevProperties)["day"])

	/*
		filter1:
		prev_salesforce_contact_character = "G" AND curr_salesforce_contact_character ="H"
	*/
	var filters []model.SmartCRMEventFilter
	filter := model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "character",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "H",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "G",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	filters = append(filters, filter)
	/*
		filter2:
		prev_salesforce_contact_character = "B" AND curr_salesforce_contact_character ="I"
	*/
	filter = model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "character",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "I",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "B",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	filters = append(filters, filter)

	/*
		filter3:
		prev_salesforce_contact_day = "Sunday" AND curr_salesforce_contact_day ="Sunday"
	*/
	filter = model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "day",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "Sunday",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "Saturday",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	filters = append(filters, filter)

	/*
		filter1(prev_salesforce_contact_character = "G" AND curr_salesforce_contact_character ="H")
		for U1
		New incoming record(salesforce_contact_character = "H")
		Expected previous record U1(day="Sunday", character="G", type = contact)
	*/
	currentProperties := make(map[string]interface{})
	currentProperties["character"] = "H"
	smartEvent, prevProperties, ok := IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter1", "", userID1, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[0])
	assert.Equal(t, true, ok)
	assert.Equal(t, "filter1", smartEvent.Name)
	assert.Equal(t, "G", smartEvent.Properties["$prev_salesforce_contact_character"])
	assert.Equal(t, "H", smartEvent.Properties["$curr_salesforce_contact_character"])
	//prev properties check
	assert.Equal(t, "Saturday", (*prevProperties)["day"])
	assert.Equal(t, "G", (*prevProperties)["character"])

	//Fail Test
	currentProperties = make(map[string]interface{})
	currentProperties["character"] = "G"
	smartEvent, prevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter1", "", userID1, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[0])
	assert.Equal(t, false, ok)
	// prev properties should be nil
	assert.Nil(t, prevProperties)

	/*
		filter2(prev_salesforce_contact_character = "B" AND curr_salesforce_contact_character ="I")
		for U2
		New incoming record(salesforce_contact_character = "I")
		Expected previous record U2(day="Monday", character="B",type = contact)
	*/
	currentProperties = make(map[string]interface{})
	currentProperties["character"] = "I"
	smartEvent, prevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter2", "", userID2, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[1])
	assert.Equal(t, true, ok)
	assert.Equal(t, "filter2", smartEvent.Name)
	assert.Equal(t, "B", smartEvent.Properties["$prev_salesforce_contact_character"])
	assert.Equal(t, "I", smartEvent.Properties["$curr_salesforce_contact_character"])
	// prev properties check
	assert.Equal(t, "B", (*prevProperties)["character"])
	assert.Equal(t, "Monday", (*prevProperties)["day"])

	//Fail Test filter2
	currentProperties["character"] = "J"
	smartEvent, prevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter2", "", userID2, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[1])
	assert.Equal(t, false, ok)
	// prev properties should be nil
	assert.Nil(t, prevProperties)

	/*
		filter3(prev_salesforce_contact_day = "Sunday" AND curr_salesforce_contact_day ="Sunday")
		for U1
		New incoming record(salesforce_contact_day = "Sunday")
		Expected previous record U1(day="Saturday", character="G", type = contact)
	*/
	currentProperties = make(map[string]interface{})
	currentProperties["day"] = "Sunday"
	smartEvent, prevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter3", "", userID1, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[2])
	assert.Equal(t, true, ok)
	assert.Equal(t, "filter3", smartEvent.Name)
	assert.Equal(t, "Saturday", smartEvent.Properties["$prev_salesforce_contact_day"])
	assert.Equal(t, "Sunday", smartEvent.Properties["$curr_salesforce_contact_day"])
	//prev properties check
	assert.Equal(t, "G", (*prevProperties)["character"])
	assert.Equal(t, "Saturday", (*prevProperties)["day"])

	//Fail Test filter2
	currentProperties = make(map[string]interface{})
	currentProperties["day"] = "Monday"
	smartEvent, prevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter2", "", userID1, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[2])
	assert.Equal(t, false, ok)
	// prev properties should be nil
	assert.Nil(t, prevProperties)
}

func TestSameUserSmartEvent(t *testing.T) {

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	filter := model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "character",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "I",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "B",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: model.TimestampReferenceTypeDocument,
	}

	smartEventName := "Event 1"
	requestPayload := make(map[string]interface{})
	requestPayload["name"] = smartEventName
	requestPayload["expr"] = filter

	w := sendCreateSmartEventFilterReq(r, project.ID, agent, &requestPayload)
	assert.Equal(t, http.StatusCreated, w.Code)
	var responsePayload H.APISmartEventFilterResponePayload
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &responsePayload)
	assert.Nil(t, err)
	stringCompEventNameId := responsePayload.EventNameID
	assert.NotEqual(t, 0, stringCompEventNameId)

	contactID := U.RandomLowerAphaNumString(5)
	name := U.RandomLowerAphaNumString(5)

	createdDate := time.Now()

	jsonData := fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s","character":"B"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentTimeLayout))
	salesforceDocumentPrev := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status := store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocumentPrev)
	assert.Equal(t, http.StatusCreated, status)

	userID1 := U.RandomLowerAphaNumString(5)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID1})
	assert.Equal(t, http.StatusCreated, status)
	eventID1 := U.RandomLowerAphaNumString(10)
	store.GetStore().UpdateSalesforceDocumentAsSynced(project.ID, salesforceDocumentPrev, eventID1, userID1)

	currentProperties := make(map[string]interface{})
	currentProperties["character"] = "I"
	jsonData = fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s","character":"%s"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentTimeLayout), createdDate.AddDate(0, 0, 1).UTC().Format(model.SalesforceDocumentTimeLayout), currentProperties["$salesforce_contact_character"])
	currentSalesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		Type:      model.SalesforceDocumentTypeContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	salesforceSmartEventName := &IntSalesforce.SalesforceSmartEventName{
		EventName: smartEventName,
		Filter:    &filter,
		Type:      model.TYPE_CRM_SALESFORCE,
	}

	eventName1 := "ev1"
	eventName, status := store.GetStore().CreateOrGetEventName(&model.EventName{ProjectId: project.ID, Name: eventName1, Type: model.TYPE_USER_CREATED_EVENT_NAME})
	assert.Equal(t, http.StatusCreated, status)
	_, errCode := store.GetStore().CreateEvent(&model.Event{
		ProjectId:   project.ID,
		EventNameId: eventName.ID,
		UserId:      userID1,
		Timestamp:   createdDate.Unix(),
	})
	assert.Equal(t, http.StatusCreated, errCode)

	eventID2 := U.RandomLowerAphaNumString(10)
	IntSalesforce.TrackSalesforceSmartEvent(project.ID, salesforceSmartEventName, eventID2, "", userID1, currentSalesforceDocument.Type, &currentProperties, nil, createdDate.AddDate(0, 0, 2).Unix())

	query := model.Query{
		From: createdDate.Unix(),
		To:   createdDate.AddDate(0, 0, 5).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       eventName1,
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       smartEventName,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.Rows[0][0])
	assert.Equal(t, int64(1), result.Rows[0][1])

	// no previous record will ruturn true for all not equal to any value
	filter = model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce user created",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "day",
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

	cuid := "123-4567"
	userID2 := "123-234-455"
	currentProperties = make(map[string]interface{})
	currentProperties["day"] = "Sunday"
	_, _, ok := IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "test", cuid, userID2, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)

	// if property value is nil
	prevProperties := make(map[string]interface{})
	prevProperties["day"] = nil
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "test", cuid, userID2, model.SalesforceDocumentTypeContact, &currentProperties, &prevProperties, &filter)
	assert.Equal(t, true, ok)
}

func TestSalesforceEventUserPropertiesState(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	refreshToken := U.RandomLowerAphaNumString(5)
	instancURL := U.RandomLowerAphaNumString(5)
	errCode := store.GetStore().UpdateAgentIntSalesforce(agent.UUID,
		refreshToken,
		instancURL,
	)
	assert.Equal(t, http.StatusAccepted, errCode)

	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntSalesforceEnabledAgentUUID: &agent.UUID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	cuID := U.RandomLowerAphaNumString(5)
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

	contactID := U.RandomLowerAphaNumString(7)
	name := U.RandomLowerAphaNumString(3)
	createdDate := time.Now()

	// salesforce record
	jsonData := fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s","Email":"%s"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentTimeLayout), cuID)
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameLead,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	//should return error on duplicate
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusConflict, status)

	//enrich job, create contact created and contact updated event
	enrichStatus := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)
	assert.Equal(t, "success", enrichStatus[2].Status)

	query := model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "$sf_lead_created",
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				Property:       "city",
				EventName:      "$sf_lead_created",
				EventNameIndex: 1,
			},
		},

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
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
				Name:       "$sf_lead_created",
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				Property:       "$user_id",
				EventName:      "$sf_lead_created",
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

func sendGetCRMObjectValuesByPropertyNameReq(r *gin.Engine, projectID uint64, agent *model.Agent, objectSource, objectType, propertyName string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
		return nil
	}

	rb := U.NewRequestBuilder(http.MethodGet, fmt.Sprintf("/projects/%d/v1/crm/%s/%s/properties/%s/values", projectID, objectSource, objectType, propertyName)).
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

func TestSalesforceObjectPropertiesAPI(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	property1 := U.RandomLowerAphaNumString(4)
	property2 := U.RandomLowerAphaNumString(4)
	documentID := U.RandomLowerAphaNumString(4)
	createdDate := time.Now().AddDate(0, 0, -1)

	jsonData := fmt.Sprintf(`{"Id":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, createdDate.UTC().Format(model.SalesforceDocumentTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentTimeLayout))
	salesforceDocumentPrev := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status := store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocumentPrev)
	assert.Equal(t, http.StatusCreated, status)

	limit := 100
	for i := 0; i < limit; i++ {
		createdDate = createdDate.Add(10 * time.Second)
		value1 := fmt.Sprintf("%s_%d", property1, i)
		value2 := fmt.Sprintf("%s_%d", property2, i)
		jsonData = fmt.Sprintf(`{"Id":"%s","%s":"%s", "%s":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, property1, value1, property2, value2, createdDate.UTC().Format(model.SalesforceDocumentTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentTimeLayout))
		salesforceDocumentPrev := &model.SalesforceDocument{
			ProjectID: project.ID,
			TypeAlias: model.SalesforceDocumentTypeNameContact,
			Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
		}
		status := store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocumentPrev)
		assert.Equal(t, http.StatusCreated, status)
	}

	var property1Values []interface{}
	var property2Values []interface{}
	w := sendGetCRMObjectValuesByPropertyNameReq(r, project.ID, agent, model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact, property1)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &property1Values)
	assert.Nil(t, err)

	w = sendGetCRMObjectValuesByPropertyNameReq(r, project.ID, agent, model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact, property2)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &property2Values)
	assert.Nil(t, err)
	for i := 0; i < limit; i++ {
		assert.Contains(t, property1Values, fmt.Sprintf("%s_%d", property1, i))
		assert.Contains(t, property2Values, fmt.Sprintf("%s_%d", property2, i))
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < i+1; j++ {
			createdDate = createdDate.Add(10 * time.Second)
			value1 := fmt.Sprintf("%s_%d", property1, i)
			value2 := fmt.Sprintf("%s_%d", property2, i)
			jsonData = fmt.Sprintf(`{"Id":"%s","%s":"%s", "%s":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, property1, value1, property2, value2, createdDate.UTC().Format(model.SalesforceDocumentTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentTimeLayout))
			salesforceDocumentPrev := &model.SalesforceDocument{
				ProjectID: project.ID,
				TypeAlias: model.SalesforceDocumentTypeNameContact,
				Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
			}
			status := store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocumentPrev)
			assert.Equal(t, http.StatusCreated, status)
		}
	}

	createdDate = createdDate.Add(10 * time.Second)
	value3 := "val3"
	jsonData = fmt.Sprintf(`{"Id":"%s","%s":"%s", "%s":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, property1, value3, property2, value3, createdDate.UTC().Format(model.SalesforceDocumentTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentTimeLayout))
	salesforceDocumentPrev = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocumentPrev)
	assert.Equal(t, http.StatusCreated, status)

	w = sendGetCRMObjectValuesByPropertyNameReq(r, project.ID, agent, model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact, property1)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &property1Values)
	assert.Nil(t, err)

	w = sendGetCRMObjectValuesByPropertyNameReq(r, project.ID, agent, model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact, property2)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &property2Values)
	assert.Nil(t, err)
	for i := range property1Values[:5] {
		assert.Equal(t, fmt.Sprintf("%s_%d", property1, 4-i), property1Values[i])
		assert.Equal(t, fmt.Sprintf("%s_%d", property2, 4-i), property2Values[i])
	}

}
