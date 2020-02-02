package hubspot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	H "factors/handler"
	M "factors/model"
	U "factors/util"
)

type Version struct {
	Name      string `json:"version"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

type Property struct {
	Value     string    `json:"value"`
	Versions  []Version `json:"versions"`
	Timestamp int64     `json:"timestamp"`
}

type Associations struct {
	AssociatedContactIds []int64 `json:"associatedVids"`
	AssociatedCompanyIds []int64 `json:"associatedCompanyIds"`
	AssociatedDealIds    []int64 `json:"associatedDealIds"`
}

type ContactIdentity struct {
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
	IsPrimary bool        `json:"is-primary"`
}

type ContactIdentityProfile struct {
	Identities []ContactIdentity `json:"identities"`
}

type Contact struct {
	Vid              int64                    `json:"vid"`
	Properties       map[string]Property      `json:"properties"`
	IdentityProfiles []ContactIdentityProfile `json:"identity-profiles"`
}

type Deal struct {
	DealId       int64               `json:"dealId"`
	Properties   map[string]Property `json:"properties"`
	Associations Associations        `json:"associations"`
}

type Company struct {
	CompanyId int64 `json:"companyId"`
	// not part of hubspot response. added to company on download.
	ContactIds []int64             `json:"contactIds"`
	Properties map[string]Property `json:"properties"`
}

const propertyNameLeadGUID = "lead_guid"

var syncOrderByType = [...]int{
	M.HubspotDocumentTypeContact,
	M.HubspotDocumentTypeCompany,
	M.HubspotDocumentTypeDeal,
}

func getContactProperties(document *M.HubspotDocument) (map[string]interface{}, error) {
	var properties map[string]interface{}

	if document.Type != M.HubspotDocumentTypeContact {
		return properties, errors.New("invalid type")
	}

	var contact Contact
	err := json.Unmarshal((document.Value).RawMessage, &contact)
	if err != nil {
		return properties, err
	}

	properties = make(map[string]interface{}, 0)

	for ipi := range contact.IdentityProfiles {
		for idi := range contact.IdentityProfiles[ipi].Identities {
			key := getPropertyKeyByType(M.HubspotDocumentTypeNameContact,
				contact.IdentityProfiles[ipi].Identities[idi].Type)
			if _, exists := properties[key]; !exists {
				properties[key] = contact.IdentityProfiles[ipi].Identities[idi].Value
			}
		}
	}

	for pkey, pvalue := range contact.Properties {
		key := getPropertyKeyByType(M.HubspotDocumentTypeNameContact, pkey)

		// give precedence to identity profiles, do not
		// overwrite same key from form.
		if _, exists := properties[key]; exists {
			continue
		}
		properties[key] = pvalue.Value
	}

	return properties, nil
}

func getCustomerUserIdFromProperties(properties map[string]interface{}) string {
	// identify using email if exist on properties.
	emailInt, emailExists := properties["email"]
	if emailExists || emailInt != nil {
		email, ok := emailInt.(string)
		if ok && email != "" {
			return email
		}
	}

	var phoneKey string
	for key := range properties {
		hasPhone := strings.Index(key, "phone")
		if hasPhone > -1 && phoneKey == "" {
			phoneKey = key
		}
	}

	userPhoneIdentified := false
	phoneInt := properties[phoneKey]
	if phoneInt != nil {
		phone := U.GetPropertyValueAsString(phoneInt)
		if phone != "" && !userPhoneIdentified {
			return phone
		}
	}

	return ""
}

func getPropertyKeyByType(typ, key string) string {
	return fmt.Sprintf("$hubspot_%s_%s", typ, strings.ToLower(key))
}

func getEventTimestamp(timestamp int64) int64 {
	if timestamp == 0 {
		return 0
	}

	return timestamp / 1000
}

func syncContact(projectId uint64, document *M.HubspotDocument) int {
	logCtx := log.WithField("project_id",
		projectId).WithField("document_id", document.ID)

	properties, err := getContactProperties(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properites from hubspot contact.")
		return http.StatusInternalServerError
	}

	leadGuid, exists := properties[getPropertyKeyByType(
		M.HubspotDocumentTypeNameContact, propertyNameLeadGUID)]
	if !exists {
		logCtx.Error("Missing lead_guid on hubspot contact properties. Sync failed.")
		return http.StatusInternalServerError
	}

	trackPayload := &H.SDKTrackPayload{
		ProjectId:       projectId,
		EventProperties: properties,
		UserProperties:  properties,
		Timestamp:       getEventTimestamp(document.Timestamp),
	}

	logCtx = logCtx.WithField("action", document.Action).WithField(
		propertyNameLeadGUID, leadGuid)

	var eventId, userId string
	if document.Action == M.HubspotDocumentActionCreated {
		trackPayload.Name = U.EVENT_NAME_HUBSPOT_CONTACT_CREATED

		status, response := H.SDKTrack(projectId, trackPayload, "", "", true)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithField("status", status).Error("Failed to track hubspot contact created event.")
			return http.StatusInternalServerError
		}

		userId = response.UserId
		eventId = response.EventId
	} else if document.Action == M.HubspotDocumentActionUpdated {
		trackPayload.Name = U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED

		userPropertiesRecords, errCode := M.GetUserPropertiesRecordsByProperty(projectId,
			getPropertyKeyByType(M.HubspotDocumentTypeNameContact, propertyNameLeadGUID), leadGuid)
		if errCode != http.StatusFound {
			logCtx.WithField("err_code", errCode).Error(
				"Failed to get user with given lead_guid. Failed to track hubspot contact updated event.")
			return http.StatusInternalServerError
		}

		// use the user_id of same lead_guid done
		// contact created event.
		userId = userPropertiesRecords[0].UserId
		trackPayload.UserId = userId
		status, response := H.SDKTrack(projectId, trackPayload, "", "", true)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithField("status", status).Error("Failed to track hubspot contact updated event.")
			return http.StatusInternalServerError
		}
		eventId = response.EventId
	} else {
		logCtx.Error("Invalid action on hubspot contact sync.")
		return http.StatusInternalServerError
	}

	customerUserId := getCustomerUserIdFromProperties(properties)
	if customerUserId != "" {
		status, _ := H.SDKIdentify(projectId, &H.SDKIdentifyPayload{
			UserId: userId, CustomerUserId: customerUserId})
		if status != http.StatusOK {
			logCtx.WithField("customer_user_id", customerUserId).Error(
				"Failed to identify user on hubspot contact sync.")
			return http.StatusInternalServerError
		}
	} else {
		logCtx.Error("Skipped user identification on hubspot contact sync. No customer_user_id on properties.")
	}

	// Mark as synced, if customer_user_id not present or present and identified.
	errCode := M.UpdateHubspotDocumentAsSynced(projectId, document.ID, eventId)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot contact document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func getDealUserId(projectId uint64, deal *Deal) string {
	logCtx := log.WithField("project_id", projectId)

	contactIds := make([]string, 0, 0)
	// Get directly associated contacts.
	if len(deal.Associations.AssociatedContactIds) > 0 {
		// Considering first contact as primary contact.
		for i := range deal.Associations.AssociatedContactIds {
			contactIds = append(contactIds,
				strconv.FormatInt(deal.Associations.AssociatedContactIds[i], 10))
		}
	}

	// If no directly associated contacts available, get
	// contacts of companies directly associated.
	if len(contactIds) == 0 && len(deal.Associations.AssociatedCompanyIds) > 0 {
		// Considering first company as primary company.
		companyId := strconv.FormatInt(deal.Associations.AssociatedCompanyIds[0], 10)
		companyDocuments, errCode := M.GetHubspotDocumentByTypeAndActions(projectId,
			[]string{companyId}, M.HubspotDocumentTypeCompany,
			[]int{M.HubspotDocumentActionCreated, M.HubspotDocumentActionUpdated})

		if errCode == http.StatusInternalServerError {
			logCtx.Error(
				"Failed to get deal user. Failed to get synced hubspot company documents.")
			return ""
		}

		companyContactIds := make(map[int64]bool, 0)
		for _, companyDocument := range companyDocuments {
			var company Company
			err := json.Unmarshal((companyDocument.Value).RawMessage, &company)
			if err != nil {
				logCtx.WithError(err).Error("Failed to unmarshal company document on get deal user")
			}

			for i := range company.ContactIds {
				companyContactIds[company.ContactIds[i]] = true
			}
		}

		for id := range companyContactIds {
			if id > 0 {
				contactIds = append(contactIds, strconv.FormatInt(id, 10))
			}
		}
	}

	if len(contactIds) == 0 {
		logCtx.Error("Failed to get deal user. No contact associated to deal.")
		return ""
	}

	contactDocuments, errCode := M.GetHubspotDocumentByTypeAndActions(projectId,
		contactIds, M.HubspotDocumentTypeContact, []int{M.HubspotDocumentActionCreated})
	if errCode == http.StatusInternalServerError {
		logCtx.Error(
			"Failed to get deal user. Failed to get synced hubspot contact documents.")
		return ""
	}

	// No synced contact document.
	if errCode == http.StatusNotFound || len(contactDocuments) == 0 {
		return ""
	}

	// Use first contact as primary contact.
	contactDocument := contactDocuments[0]
	if contactDocument.SyncId == "" {
		logCtx.Error("No sync_id on synced hubspot contact document.")
		return ""
	}

	event, errCode := M.GetEventById(projectId, contactDocument.SyncId)
	if errCode != http.StatusFound {
		logCtx.WithField("event_id", contactDocument.SyncId).Error(
			"Failed to get deal user. Failed to get hubspot contact created event using sync_id.")
		return ""
	}

	return event.UserId
}

func syncCompany(projectId uint64, document *M.HubspotDocument) int {
	if document.Type != M.HubspotDocumentTypeCompany {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id",
		projectId).WithField("document_id", document.ID)

	var company Company
	err := json.Unmarshal((document.Value).RawMessage, &company)
	if err != nil {
		logCtx.WithError(err).Error("Falied to unmarshal hubspot company document.")
		return http.StatusInternalServerError
	}

	if len(company.ContactIds) == 0 {
		logCtx.Error("Skipped company sync. No contacts associated to company.")
		return http.StatusOK
	}

	contactIds := make([]string, 0, 0)
	for i := range company.ContactIds {
		contactIds = append(contactIds,
			strconv.FormatInt(company.ContactIds[i], 10))
	}

	contactDocuments, errCode := M.GetHubspotDocumentByTypeAndActions(projectId,
		contactIds, M.HubspotDocumentTypeContact, []int{M.HubspotDocumentActionCreated})
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Failed to get hubspot documents by type and action on sync company.")
		return errCode
	}

	// build company properties json from properties. make sure company name exist.
	companyProperties := make(map[string]interface{}, 0)
	for key, value := range company.Properties {
		propertyKey := getPropertyKeyByType(M.HubspotDocumentTypeNameCompany, key)
		companyProperties[propertyKey] = value.Value
	}

	companyPropertiesJsonb, err := U.EncodeToPostgresJsonb(&companyProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to marshal company properties to Jsonb.")
		return http.StatusInternalServerError
	}

	// update $hubspot_company_name and other company
	// properties on each associated contact user.
	isContactsUpdateFailed := false
	for _, contactDocument := range contactDocuments {
		if contactDocument.SyncId != "" {
			contactSyncEvent, errCode := M.GetEventById(projectId,
				contactDocument.SyncId)
			if errCode == http.StatusFound {
				_, errCode := M.UpdateUserProperties(projectId,
					contactSyncEvent.UserId, companyPropertiesJsonb)
				if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
					logCtx.WithField("user_id", contactSyncEvent.UserId).Error(
						"Failed to update user properites with company properties.")
					isContactsUpdateFailed = true
				}
			}
		}
	}

	if isContactsUpdateFailed {
		logCtx.Error("Failed to update some hubspot company properties on user properties.")
		return http.StatusInternalServerError
	}

	// No sync_id as no event or user or one user property created.
	errCode = M.UpdateHubspotDocumentAsSynced(projectId, document.ID, "")
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot deal document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func syncDeal(projectId uint64, document *M.HubspotDocument) int {
	if document.Type != M.HubspotDocumentTypeDeal {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id",
		projectId).WithField("document_id", document.ID)

	var deal Deal
	err := json.Unmarshal((document.Value).RawMessage, &deal)
	if err != nil {
		logCtx.Error("Failed to unmarshal hubspot document deal.")
		return http.StatusInternalServerError
	}

	properties := make(map[string]interface{}, 0)
	for k, v := range deal.Properties {
		key := getPropertyKeyByType(M.HubspotDocumentTypeNameDeal, k)
		properties[key] = v.Value
	}

	dealStage, exists := properties[getPropertyKeyByType(
		M.HubspotDocumentTypeNameDeal, "dealstage")]
	if !exists || dealStage == nil {
		logCtx.Error("No deal stage property found on hubspot deal.")
		return http.StatusInternalServerError
	}

	userId := getDealUserId(projectId, &deal)
	if userId == "" {
		logCtx.Error("Skipped deal sync. No user associated to hubspot deal.")
		return http.StatusOK
	}

	trackPayload := &H.SDKTrackPayload{
		Name:            U.EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED,
		ProjectId:       projectId,
		UserId:          userId,
		EventProperties: properties,
		UserProperties:  properties,
		Timestamp:       getEventTimestamp(document.Timestamp),
	}

	// Track deal stage change only if, deal with same id and
	// same stage, not synced before.
	dealId := strconv.FormatInt(deal.DealId, 10)
	if dealId == "" {
		logCtx.Error("Invalid deal_id on conversion. Failed to sync deal.")
		return http.StatusInternalServerError
	}

	_, errCode := M.GetSyncedHubspotDealDocumentByIdAndStage(projectId,
		dealId, dealStage.(string))
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		logCtx.Error("Failed to get synced deal document by stage on sync_deal")
		return http.StatusInternalServerError
	}

	// skip sync as deal stage is synced already.
	if errCode == http.StatusFound {
		return http.StatusOK
	}

	status, response := H.SDKTrack(projectId, trackPayload, "", "", true)
	if status != http.StatusOK && status != http.StatusFound &&
		status != http.StatusNotModified {

		logCtx.WithField("status", status).Error(
			"Failed to track hubspot contact deal stage change event.")
		return http.StatusInternalServerError
	}

	errCode = M.UpdateHubspotDocumentAsSynced(projectId,
		document.ID, response.EventId)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot deal document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func syncAll(projectId uint64, documents []M.HubspotDocument) int {
	var seenFailures bool
	for i := range documents {
		switch documents[i].Type {
		case M.HubspotDocumentTypeContact:
			errCode := syncContact(projectId, &documents[i])
			if errCode != http.StatusOK {
				seenFailures = true
			}
		case M.HubspotDocumentTypeCompany:
			errCode := syncCompany(projectId, &documents[i])
			if errCode != http.StatusOK {
				seenFailures = true
			}
		case M.HubspotDocumentTypeDeal:
			errCode := syncDeal(projectId, &documents[i])
			if errCode != http.StatusOK {
				seenFailures = true
			}
		}
	}

	if seenFailures {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

type Status struct {
	ProjectId uint64 `json:"project_id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

// Sync - Syncs hubspot documents in an order of type.
func Sync(projectId uint64) []Status {
	logCtx := log.WithField("project_id", projectId)

	statusByProjectAndType := make([]Status, 0, 0)
	for i := range syncOrderByType {
		logCtx = logCtx.WithField("type", syncOrderByType[i])

		documents, errCode := M.GetHubspotDocumentsByTypeForSync(
			projectId, syncOrderByType[i])
		if errCode != http.StatusFound {
			logCtx.Error("Failed to get hubspot document by type for sync.")
			continue
		}

		status := Status{ProjectId: projectId,
			Type: M.GetHubspotTypeAliasByType(syncOrderByType[i])}

		errCode = syncAll(projectId, documents)
		if errCode == http.StatusOK {
			status.Status = "success"
		} else {
			status.Status = "failures_seen"
		}
		statusByProjectAndType = append(statusByProjectAndType, status)
	}

	return statusByProjectAndType
}
