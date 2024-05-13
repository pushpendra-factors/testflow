package tests

import (
	"encoding/base64"
	"encoding/json"
	DD "factors/default_data"
	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	"factors/model/store/memsql"
	T "factors/task"
	U "factors/util"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDefaultWidgetGroupsCreation(t *testing.T) {
	project, _, _ := SetupProjectWithAgentDAO()

	areWidgetsAdded, _, statusCode4 := store.GetStore().AreWidgetsAddedToWidgetGroup(project.ID)
	assert.Equal(t, http.StatusFound, statusCode4)
	assert.Equal(t, false, areWidgetsAdded)

	widgetGroup, errCode2, statusCode2 := store.GetStore().AddWidgetsToWidgetGroup(project.ID, memsql.MarketingEngagementWidgetGroup, model.HUBSPOT)
	assert.Equal(t, "", errCode2)
	assert.Equal(t, http.StatusCreated, statusCode2)
	assert.Equal(t, true, widgetGroup.WidgetsAdded)
	store.GetStore().AddWidgetsToWidgetGroup(project.ID, memsql.SalesOppWidgetGroup, model.HUBSPOT)

	// Invalid query Metric but adding just to check.
	// Custom KPI is not created. Just for testing, I am testing a custom KPI.
	widget := model.Widget{
		QueryType:   model.QueryClassKPI,
		QueryMetric: "a",
		DisplayName: "widget 1",
	}
	widget.ID = uuid.New().String()
	widget.CreatedAt = time.Now()
	widget.UpdatedAt = time.Now()
	widgetGroup.DecodeWidgetsAndSetDecodedWidgets()

	_, _, statusCode5 := store.GetStore().AddWidgetToWidgetGroup(widgetGroup, widget)
	assert.Equal(t, http.StatusCreated, statusCode5)
}

func TestWidgetGroupExecution1(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, _, _ := SetupProjectWithAgentDAO()
	factory := DD.GetDefaultDataCustomKPIFactory(DD.HubspotIntegrationName)
	statusCode2 := factory.Build(project.ID)

	widgetGroup, _, _ := store.GetStore().AddWidgetsToWidgetGroup(project.ID, memsql.MarketingEngagementWidgetGroup, model.HUBSPOT)
	_, err1, _ := store.GetStore().AddWidgetsToWidgetGroup(project.ID, memsql.SalesOppWidgetGroup, model.HUBSPOT)
	assert.NotNil(t, err1)

	widgetGroup, _, _ = store.GetStore().GetWidgetGroupByID(project.ID, widgetGroup.ID)

	domaindGroup, _ := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.NotNil(t, domaindGroup)

	domProperties := postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"$domain_name":"%s",
		"$engagement_level":"Hot","$engagement_score":125.300000,"$joinTime":1681211371,
		"$total_enagagement_score":196.000000,"$in_hubspot":true}`, "adpushup.com"))}
	source := model.GetRequestSourcePointer(model.UserSourceDomains)
	groupUser := true
	domID, _ := store.GetStore().CreateUser(&model.User{
		ID:          fmt.Sprintf("dom-%s", base64.StdEncoding.EncodeToString([]byte("adpushup.com"))),
		ProjectId:   project.ID,
		Source:      source,
		Group1ID:    "adpushup.com",
		Properties:  domProperties,
		IsGroupUser: &groupUser,
	})
	domainUser, errCode := store.GetStore().GetUser(project.ID, domID)

	// Create CRM groups
	hubspotGroup, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, hubspotGroup)

	hubspotGroupDeal, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_DEAL, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, hubspotGroupDeal)

	dummyPropsMap := []map[string]interface{}{
		{"$hubspot_company_name": "AdPushup", "$hubspot_company_country": "US", "$hubspot_deal_dealstage": "closedwon",
			"$hubspot_company_hs_object_id": 2, "$hubspot_company_domain": "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_deal_name": "abc2", "$hubspot_deal_domain": "adpushup.com", "$hubspot_deal_region": "B", "$hubspot_deal_createdate": time.Now().Unix()},
	}

	userPropsMap := []map[string]interface{}{
		{"$browser": "Chrome", "$city": "London", "$country": "UK", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2000, U.UP_COMPANY: "XYZ Company"},
	}

	groupUsers := make([]model.User, 0)
	users := make([]model.User, 0)

	propertiesJSON, err := json.Marshal(dummyPropsMap[0])
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	properties := postgres.Jsonb{RawMessage: propertiesJSON}
	isGroupUser := true
	var inHubspot string
	var hsUserID string
	var customerUserID string
	lastEventTime := time.Now()
	source = model.GetRequestSourcePointer(model.UserSourceHubspot)
	inHubspot = "244"
	customerUserID = fmt.Sprintf("hsuser%d@%s", 1, dummyPropsMap[0]["$hubspot_company_domain"])
	createdGroupUserID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:    project.ID,
		Properties:   properties,
		IsGroupUser:  &isGroupUser,
		Group1ID:     domainUser.Group1ID,
		Group1UserID: domainUser.ID,
		Group2ID:     inHubspot,
		Group3ID:     "",
		Group4ID:     "",
		Source:       source,
		LastEventAt:  &lastEventTime,
	})
	account, errCode := store.GetStore().GetUser(project.ID, createdGroupUserID)
	assert.Equal(t, http.StatusFound, errCode)

	hsUserID = createdGroupUserID
	groupUsers = append(groupUsers, *account)

	// user associated to the account
	isGroupUser = false
	propertiesJSON, _ = json.Marshal(userPropsMap[0])
	userProperties := postgres.Jsonb{RawMessage: propertiesJSON}
	createdUserID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Properties:     userProperties,
		IsGroupUser:    &isGroupUser,
		Group1UserID:   domainUser.ID,
		Group2UserID:   hsUserID,
		Group3UserID:   "",
		Group4UserID:   "",
		CustomerUserId: customerUserID,
		Source:         source,
		LastEventAt:    &lastEventTime,
	})
	user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
	users = append(users, *user)

	// Creating hubspot deal.
	propertiesJSON1, err := json.Marshal(dummyPropsMap[1])
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	properties1 := postgres.Jsonb{RawMessage: propertiesJSON1}
	isGroupUser1 := true
	lastEventTime1 := time.Now()
	source = model.GetRequestSourcePointer(model.UserSourceHubspot)
	inHubspot = "367" // Being hard coded.
	customerUserID = fmt.Sprintf("hsuser%d@%s", 1, dummyPropsMap[1]["$hubspot_deal_domain"])
	_, statusCode1 := store.GetStore().CreateUser(&model.User{
		ProjectId:    project.ID,
		Properties:   properties1,
		IsGroupUser:  &isGroupUser1,
		Group1ID:     domainUser.Group1ID,
		Group1UserID: domainUser.ID,
		Group2ID:     "",
		Group3ID:     inHubspot,
		Group4ID:     "",
		Source:       source,
		LastEventAt:  &lastEventTime1,
	})
	log.WithField("statusCode1", statusCode1).Warn("k1")
	segment, _ := store.GetStore().GetSegmentByName(project.ID, "In Hubspot")

	pIDList := map[int64]bool{project.ID: true}
	errCode = T.SegmentMarker(project.ID, pIDList, map[int64][]string{})
	assert.Equal(t, errCode, http.StatusOK)

	requestParamsForExecution := model.RequestSegmentKPI{}
	requestParamsForExecution.From = 1577840461
	requestParamsForExecution.To = time.Now().Unix()
	requestParamsForExecution.Timezone = "Asia/Kolkata"

	// Querying.
	results, statusCode := store.GetStore().ExecuteWidgetGroup(project.ID, widgetGroup, segment.Id, uuid.New().String(), requestParamsForExecution)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, 1, len(results[0].Rows))
	assert.Equal(t, float64(1), results[0].Rows[0][0])
	assert.Equal(t, 1, len(results[0].Rows))
	assert.Equal(t, float64(0), results[1].Rows[0][0])

	widgetGroup2, _, _ := store.GetStore().GetWidgetGroupByName(project.ID, memsql.AccountsWidgetGroup)
	log.WithField("widgetGroup2", widgetGroup2.DecodedWidgets).Warn("kark244")

	results2, statusCode2 := store.GetStore().ExecuteWidgetGroup(project.ID, widgetGroup2, segment.Id, uuid.New().String(), requestParamsForExecution)
	log.WithField("results2", results2).Warn("kark2")
	assert.Equal(t, http.StatusOK, statusCode2)
	assert.Equal(t, 1, len(results2[0].Rows))
	assert.Equal(t, float64(1), results2[0].Rows[0][0])
	assert.Equal(t, 1, len(results2[0].Rows))
	assert.Equal(t, float64(0), results2[1].Rows[0][0])
}

func TestAccountAnalyticsExecution(t *testing.T) {
	widget := model.Widget{QueryMetric: model.HighEngagedAccountsMetric}
	requestParams := model.RequestSegmentKPI{}
	analyticsQuery := store.GetStore().BuildAccountAnalytics(1, widget, "abc", requestParams)

	_, sCode := store.GetStore().ExecuteAccountAnalyticsQuery(1, "", analyticsQuery)
	assert.Equal(t, http.StatusOK, sCode)
}

// func TestWidgetGroupExecution(t *testing.T) {
// 	r := gin.Default()
// 	H.InitAppRoutes(r)

// 	// project, agent, err := SetupProjectWithAgentDAO()
// 	project, _, err := SetupProjectWithAgentDAO()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, project)

// 	startTimestamp := time.Now().Unix()

// 	// create new hubspot document
// 	jsonDealModel := `{
// 		"dealId": %d,
// 		"properties": {
// 			"amount": { "value": "%d" },
// 			"createdate": { "value": "%d" },
// 			"hs_createdate": { "value": "%d" },
// 		  	"dealname": { "value": "%s" },
// 			"latest_source": { "value": "%s" },
// 		  	"hs_lastmodifieddate": { "value": "%d" }
// 		}
// 	}`

// 	latestSources := []string{"ORGANIC_SEARCH", "DIRECT_TRAFFIC", "PAID_SOCIAL"}
// 	hubspotDocuments := make([]*model.HubspotDocument, 0)
// 	for i := 0; i < len(latestSources); i++ {
// 		documentID := i
// 		createdTime := startTimestamp*1000 + int64(i*100)
// 		updatedTime := createdTime + 200
// 		amount := U.RandomIntInRange(1000, 2000)
// 		jsonDeal := fmt.Sprintf(jsonDealModel, documentID, amount, createdTime, createdTime, fmt.Sprintf("Dealname %d", i), latestSources[i], updatedTime)

// 		document := model.HubspotDocument{
// 			TypeAlias: model.HubspotDocumentTypeNameDeal,
// 			Value:     &postgres.Jsonb{json.RawMessage(jsonDeal)},
// 		}
// 		hubspotDocuments = append(hubspotDocuments, &document)
// 	}

// 	status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeDeal, hubspotDocuments, 3)
// 	assert.Equal(t, http.StatusCreated, status)

// 	// execute sync job
// 	allStatus, _ := IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50, 3, "abc.com")
// 	for i := range allStatus {
// 		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, allStatus[i].Status)
// 	}
// }

// -------------------------------------

// store.GetStore().GetWidgetGroupAndWidgetsForConfig(project.ID)
// // abc(project.ID, abc1, abc2, abc3)
// factory := DD.GetDefaultDataCustomKPIFactory(DD.HubspotIntegrationName)
// statusCode2 := factory.Build(project.ID)
// if statusCode2 != http.StatusOK {
// 	_, errMsg := fmt.Printf("Failed during prebuilt custom KPI creation for: %v, %v", project.ID, "Hubspot")
// 	log.WithField("projectID", project.ID).WithField("integration", "hubspot").Warn(errMsg)
// 	// return "", http.StatusInternalServerError, "", "", false
// }

// factory2 := DD.GetDefaultDataCustomKPIFactory(DD.SalesforceIntegrationName)
// statusCode3 := factory2.Build(project.ID)
// if statusCode3 != http.StatusOK {
// 	_, errMsg := fmt.Printf("Failed during prebuilt custom KPI creation for: %v, %v", project.ID, "sa")
// 	log.WithField("projectID", project.ID).WithField("integration", "sa").Warn(errMsg)
// 	// return "", http.StatusInternalServerError, "", "", false
// }

// var abc1 = []model.CustomMetric{
// 	{
// 		Name:        model.HubspotContacts,
// 		Description: "Hubspot Contacts timestamped at Contact Create Date",
// 		TypeOfQuery: 1,
// 		ObjectType:  model.HubspotContactsDisplayCategory,
// 	},
// }
// var abc2 = []model.CustomMetricTransformation{
// 	{
// 		AggregateFunction:     model.UniqueAggregateFunction,
// 		AggregateProperty:     "",
// 		AggregatePropertyType: "categorical",
// 		Filters:               []model.KPIFilter{},
// 		DateField:             "$hubspot_contact_createdate",
// 		EventName:             "",
// 		Entity:                model.UserEntity,
// 	},
// }
// var abc3 = []model.KPIQueryGroup{}

// func abc(projectID int64, customMetrics []model.CustomMetric, customTransformations []model.CustomMetricTransformation,
// 	derivedTransformations []model.KPIQueryGroup) int {
// 	selectedStore := store.GetStore()
// 	postgresTransformations, statusCode := buildJsonTransformationsForCustomKPIs(customMetrics, customTransformations, derivedTransformations)
// 	if statusCode != http.StatusOK {
// 		log.WithField("projectId", projectID).Warn("Failed in building Json Transformations for hubspot")
// 		return http.StatusInternalServerError
// 	}

// 	for index := range customMetrics {
// 		customMetrics[index].ProjectID = projectID
// 		customMetrics[index].Transformations = &postgresTransformations[index]
// 	}

// 	for _, customMetric := range customMetrics {
// 		_, _, statusCode := selectedStore.CreateCustomMetric(customMetric)
// 		if statusCode == http.StatusConflict {
// 			continue
// 		}
// 		if statusCode != http.StatusCreated {
// 			return http.StatusInternalServerError
// 		}
// 	}

// 	return http.StatusOK
// }

// // TODO add error for the methods which are calling.
// func buildJsonTransformationsForCustomKPIs(customMetrics []model.CustomMetric, profileTransformations []model.CustomMetricTransformation, derivedTransformations []model.KPIQueryGroup) ([]postgres.Jsonb, int) {

// 	resTransformations := make([]postgres.Jsonb, 0)
// 	indexOfProfileBasedTransformation := 0
// 	indexOfDerivedTransformation := 0
// 	for _, customMetric := range customMetrics {
// 		if customMetric.TypeOfQuery == 1 {
// 			transformation := profileTransformations[indexOfProfileBasedTransformation]
// 			jsonTransformation, err := json.Marshal(transformation)
// 			if err != nil {
// 				return make([]postgres.Jsonb, 0), http.StatusInternalServerError
// 			}
// 			postgresTransformation := postgres.Jsonb{json.RawMessage(jsonTransformation)}
// 			resTransformations = append(resTransformations, postgresTransformation)

// 			indexOfProfileBasedTransformation++
// 		} else {
// 			transformation := derivedTransformations[indexOfDerivedTransformation]
// 			jsonTransformation, err := json.Marshal(transformation)
// 			if err != nil {
// 				return make([]postgres.Jsonb, 0), http.StatusInternalServerError
// 			}
// 			postgresTransformation := postgres.Jsonb{json.RawMessage(jsonTransformation)}
// 			resTransformations = append(resTransformations, postgresTransformation)

// 			indexOfDerivedTransformation++
// 		}
// 	}
// 	return resTransformations, http.StatusOK
// }
