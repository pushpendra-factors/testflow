package model

import (
	"fmt"
)

const (
	LEADSQUARED_LEAD           = "lead"
	LEADSQUARED_SALES_ACTIVITY = "sales_activity"
	LEADSQUARED_EMAIL_SENT     = "email_sent"
	LEADSQUARED_EMAIL_INFO     = "email_info"
	LEADSQUARED_HAD_A_CALL     = "had_a_call"
	LEADSQUARED_CALLED_A_CUST_NEGATIVE_REPLY = "called_a_customer_negative_reply"
	LEADSQUARED_CALLED_A_CUST_POSITIVE_REPLY = "called_a_customer_positive_reply"
	LEADSQUARED_CALLED_TO_COLLECT_REFERRAL = "called_to_collect_referrals"
	LEADSQUARED_EMAIL_BOUNCED = "email_bounced"
	LEADSQUARED_EMAIL_LINK_CLICKED = "email_link_clicked"
	LEADSQUARED_EMAIL_MAILING_PREFERENCE_LINK_CLICKED = "email_mailing_preference_link_clicked"
	LEADSQUARED_EMAIL_MARKED_SPAM = "email_marked_spam"
	LEASQUARED_EMAIL_NEGATIVE_RESPONSE = "email_negative_response"
	LEASQUARED_EMAIL_NEUTRAL_RESPONSE = "email_neutral_response"
	LEASQUARED_EMAIL_POSITIVE_RESPONSE = "email_positive_response"
	LEASQUARED_EMAIL_OPENED = "email_opened"
	LEASQUARED_EMAIL_POSITVE_INBOUND_EMAIL = "email_positive_inbound_email"
	LEASQUARED_EMAIL_RESUBSCRIBED = "email_resubscribed"
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_BOOTCAMP = "email_subscribed_to_bootcamp"
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_COLLECTION = "email_subscribed_to_collection"
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_EVENTS = "email_subscribed_to_events"
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_FESTIVAL = "email_subscribed_to_festival"
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION = "email_subscribed_to_internation_reactivation"
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_NEWSLETTER = "email_subscribed_to_newsletter"
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_REACTIVATION = "email_subscribed_to_reactivation"
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_REFERRAL = "email_subscribed_to_referrral"
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_SURVEY = "email_subscribed_to_survey"
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_TEST = "email_subscribed_to_test"
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_WORKSHOP = "email_subscribed_to_workshop"
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_BOOTCAMP = "email_unsubscribed_to_bootcamp"
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_COLLECTION = "email_unsubscribed_to_collection"
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_EVENTS = "email_unsubscribed_to_events"
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_FESTIVAL = "email_unsubscribed_to_festival"
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION = "email_unsubscribed_to_internation_reactivation"
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_NEWSLETTER = "email_unsubscribed_to_newsletter"
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REACTIVATION = "email_unsubscribed_to_reactivation"
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REFERRAL = "email_unsubscribed_to_referrral"
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_SURVEY = "email_unsubscribed_to_survey"
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_TEST = "email_unsubscribed_to_test"
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_WORKSHOP = "email_unsubscribed_to_workshop"
	LEADSQUARED_EMAIL_UNSUBSCRIBE_LINK_CLICKED = "email_unsubscribe_link_clicked"
	LEADSQUARED_EMAIL_UNSUBSCRIBED = "email_unsubscribed"
	LEADSQUARED_EMAIL_VIEW_IN_BROWSER_LINK_CLICKED = "email_view_in_browser_link_clicked"
	LEADSQUARED_EMAIL_RECEIVED = "email_received"
)

func LeadSquaredDocumentToQuery(docType string)string {
	if(docType == LEADSQUARED_LEAD){
		return "select * FROM `%s.%s."+ LeadSquaredTableName[docType] +"` WHERE %v AND ProspectAutoId > %v order by ProspectAutoId asc LIMIT %v OFFSET 0"
	} else if(ActivityEvents[docType] == true){
		return "select * FROM `%s.%s."+ LeadSquaredTableName[docType] +"` WHERE %v AND CreatedOnUnix > %v order by CreatedOnUnix asc LIMIT %v OFFSET 0"
	}
	return ""
}

var ActivityEvents = map[string]bool {
	LEADSQUARED_SALES_ACTIVITY : true,
	LEADSQUARED_EMAIL_SENT     : true,
	LEADSQUARED_EMAIL_INFO     : true,
	LEADSQUARED_HAD_A_CALL     : true,
	LEADSQUARED_CALLED_A_CUST_NEGATIVE_REPLY : true,
	LEADSQUARED_CALLED_A_CUST_POSITIVE_REPLY : true,
	LEADSQUARED_CALLED_TO_COLLECT_REFERRAL : true,
	LEADSQUARED_EMAIL_BOUNCED : true,
	LEADSQUARED_EMAIL_LINK_CLICKED : true,
	LEADSQUARED_EMAIL_MAILING_PREFERENCE_LINK_CLICKED : true,
	LEADSQUARED_EMAIL_MARKED_SPAM : true,
	LEASQUARED_EMAIL_NEGATIVE_RESPONSE : true,
	LEASQUARED_EMAIL_NEUTRAL_RESPONSE : true,
	LEASQUARED_EMAIL_POSITIVE_RESPONSE : true,
	LEASQUARED_EMAIL_OPENED : true,
	LEASQUARED_EMAIL_POSITVE_INBOUND_EMAIL : true,
	LEASQUARED_EMAIL_RESUBSCRIBED : true,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_BOOTCAMP : true,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_COLLECTION : true,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_EVENTS : true,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_FESTIVAL : true,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION : true,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_NEWSLETTER : true,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_REACTIVATION : true,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_REFERRAL : true,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_SURVEY : true,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_TEST : true,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_WORKSHOP : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_BOOTCAMP : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_COLLECTION : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_EVENTS : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_FESTIVAL : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_NEWSLETTER : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REACTIVATION : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REFERRAL : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_SURVEY : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_TEST : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_WORKSHOP : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBE_LINK_CLICKED : true,
	LEADSQUARED_EMAIL_UNSUBSCRIBED : true,
	LEADSQUARED_EMAIL_VIEW_IN_BROWSER_LINK_CLICKED : true,
	LEADSQUARED_EMAIL_RECEIVED : true,
}

func LeadSquaredDocumentEndpoint(docType string)string {
	if(docType == LEADSQUARED_LEAD){
		return  "/v2/LeadManagement.svc/Leads.RecentlyModified"
	} else if(ActivityEvents[docType] == true){
		return  "/v2/ProspectActivity.svc/CustomActivity/RetrieveByActivityEvent"
	}
	return ""
}

func LeadSquaredDocumentCreatedEventName(docType string)string {
	if(ActivityEvents[docType] == true){
		return  LeadSquaredTableName[docType] + "_activity_created"
	}
	return ""
}

func LeadSquaredDocumentUpdatedEventName(docType string)string {
	if(ActivityEvents[docType] == true){
		return  LeadSquaredTableName[docType] + "_activity_updated"
	}
	return ""
}

func LeadSquaredMetadataEndpoint(docType string)string {
	if(docType == LEADSQUARED_LEAD){
		return  "/v2/LeadManagement.svc/LeadsMetaData.Get"
	} else if(ActivityEvents[docType] == true){
		return  "/v2/ProspectActivity.svc/CustomActivity/GetActivitySetting"
	}
	return ""
}

var LeadSquaredHistoricalSyncEndpoint = map[string]string{
	LEADSQUARED_LEAD: "/v2/LeadManagement.svc/Leads.Get",
}

var LeadSquaredTableName = map[string]string{
	LEADSQUARED_LEAD:           "lead",
	LEADSQUARED_SALES_ACTIVITY: "sales_activity",
	LEADSQUARED_EMAIL_SENT:     "email_sent",
	LEADSQUARED_EMAIL_INFO:     "email_info",
	LEADSQUARED_HAD_A_CALL:     "had_a_call",
	LEADSQUARED_CALLED_A_CUST_NEGATIVE_REPLY : "called_a_customer_negative_reply",
	LEADSQUARED_CALLED_A_CUST_POSITIVE_REPLY : "called_a_customer_positive_reply",
	LEADSQUARED_CALLED_TO_COLLECT_REFERRAL : "called_to_collect_referrals",
	LEADSQUARED_EMAIL_BOUNCED : "email_bounced",
	LEADSQUARED_EMAIL_LINK_CLICKED : "email_link_clicked",
	LEADSQUARED_EMAIL_MAILING_PREFERENCE_LINK_CLICKED : "email_mailing_preference_link_clicked",
	LEADSQUARED_EMAIL_MARKED_SPAM : "email_marked_spam",
	LEASQUARED_EMAIL_NEGATIVE_RESPONSE : "email_negative_response",
	LEASQUARED_EMAIL_NEUTRAL_RESPONSE : "email_neutral_response",
	LEASQUARED_EMAIL_POSITIVE_RESPONSE : "email_positive_response",
	LEASQUARED_EMAIL_OPENED : "email_opened",
	LEASQUARED_EMAIL_POSITVE_INBOUND_EMAIL : "email_positive_inbound_email",
	LEASQUARED_EMAIL_RESUBSCRIBED : "email_resubscribed",
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_BOOTCAMP : "email_subscribed_to_bootcamp",
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_COLLECTION : "email_subscribed_to_collection",
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_EVENTS : "email_subscribed_to_events",
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_FESTIVAL : "email_subscribed_to_festival",
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION : "email_subscribed_to_internation_reactivation",
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_NEWSLETTER : "email_subscribed_to_newsletter",
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_REACTIVATION : "email_subscribed_to_reactivation",
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_REFERRAL : "email_subscribed_to_referrral",
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_SURVEY : "email_subscribed_to_survey",
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_TEST : "email_subscribed_to_test",
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_WORKSHOP : "email_subscribed_to_workshop",
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_BOOTCAMP : "email_unsubscribed_to_bootcamp",
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_COLLECTION : "email_unsubscribed_to_collection",
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_EVENTS : "email_unsubscribed_to_events",
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_FESTIVAL : "email_unsubscribed_to_festival",
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION : "email_unsubscribed_to_internation_reactivation",
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_NEWSLETTER : "email_unsubscribed_to_newsletter",
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REACTIVATION : "email_unsubscribed_to_reactivation",
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REFERRAL : "email_unsubscribed_to_referrral",
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_SURVEY : "email_unsubscribed_to_survey",
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_TEST : "email_unsubscribed_to_test",
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_WORKSHOP : "email_unsubscribed_to_workshop",
	LEADSQUARED_EMAIL_UNSUBSCRIBE_LINK_CLICKED : "email_unsubscribe_link_clicked",
	LEADSQUARED_EMAIL_UNSUBSCRIBED : "email_unsubscribed",
	LEADSQUARED_EMAIL_VIEW_IN_BROWSER_LINK_CLICKED : "email_view_in_browser_link_clicked",
	LEADSQUARED_EMAIL_RECEIVED : "email_received",
}

func LeadSquaredDataObjectColumnsQuery(docType string) string {
	return "SELECT * FROM `%s.%s.INFORMATION_SCHEMA.COLUMNS` WHERE table_name = '"+ LeadSquaredTableName[docType] +"' ORDER by ordinal_position"
}

func LeadSquaredDocTypeIntegrationObjectMap(docType string)string {
	if(docType == LEADSQUARED_LEAD){
		return  "user"
	} else if(ActivityEvents[docType] == true){
		return  "activity"
	}
	return ""
}

var LeadSquaredUserIdMapping = map[string]string{
	LEADSQUARED_LEAD: "ProspectID",
}

func LeadSquaredUserAutoIdMapping(docType string) (string, bool) {
	if(docType == LEADSQUARED_LEAD){
		return  "ProspectAutoId", true
	} else if(ActivityEvents[docType] == true){
		return  "CreatedOnUnix", true 
	}
	return "", false
}

var LeadSquaredEmailMapping = map[string]string{
	LEADSQUARED_LEAD: "EmailAddress",
}

var LeadSquaredPhoneMapping = map[string]string{
	LEADSQUARED_LEAD: "Phone",
}

var LeadSquaredDocumentTypeAlias = map[string]int{
	LEADSQUARED_LEAD:           1,
	LEADSQUARED_SALES_ACTIVITY: 2,
	LEADSQUARED_EMAIL_SENT:     3,
	LEADSQUARED_EMAIL_INFO:     4,
	LEADSQUARED_HAD_A_CALL:     5,
	LEADSQUARED_CALLED_A_CUST_NEGATIVE_REPLY : 6,
	LEADSQUARED_CALLED_A_CUST_POSITIVE_REPLY : 7,
	LEADSQUARED_CALLED_TO_COLLECT_REFERRAL : 8,
	LEADSQUARED_EMAIL_BOUNCED : 9,
	LEADSQUARED_EMAIL_LINK_CLICKED : 10,
	LEADSQUARED_EMAIL_MAILING_PREFERENCE_LINK_CLICKED : 11,
	LEADSQUARED_EMAIL_MARKED_SPAM : 12,
	LEASQUARED_EMAIL_NEGATIVE_RESPONSE : 13,
	LEASQUARED_EMAIL_NEUTRAL_RESPONSE : 14,
	LEASQUARED_EMAIL_POSITIVE_RESPONSE : 15,
	LEASQUARED_EMAIL_OPENED : 16,
	LEASQUARED_EMAIL_POSITVE_INBOUND_EMAIL : 17,
	LEASQUARED_EMAIL_RESUBSCRIBED : 18,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_BOOTCAMP : 19,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_COLLECTION : 20,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_EVENTS : 21,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_FESTIVAL : 22,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION : 23,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_NEWSLETTER : 24,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_REACTIVATION : 25,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_REFERRAL : 26,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_SURVEY : 27,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_TEST : 28,
	LEADSQUARED_EMAIL_SUBSCRIBED_TO_WORKSHOP : 29,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_BOOTCAMP : 30,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_COLLECTION : 31,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_EVENTS : 32,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_FESTIVAL : 33,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION : 34,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_NEWSLETTER : 35,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REACTIVATION : 36,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REFERRAL : 37,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_SURVEY : 38,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_TEST : 39,
	LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_WORKSHOP : 40,
	LEADSQUARED_EMAIL_UNSUBSCRIBE_LINK_CLICKED : 41,
	LEADSQUARED_EMAIL_UNSUBSCRIBED : 42,
	LEADSQUARED_EMAIL_VIEW_IN_BROWSER_LINK_CLICKED : 43,
	LEADSQUARED_EMAIL_RECEIVED : 44,
}

func LeadSquaredDataObjectFilters(docType string)string {
	if(docType == LEADSQUARED_LEAD){
		return  "DATE(%v) = '%v'"
	} else if(ActivityEvents[docType] == true){
		return  "DATE(%v) = '%v'"
	}
	return ""
}

func LeadSquaredDataObjectFiltersColumn(docType string)string {
	if(docType == LEADSQUARED_LEAD){
		return "synced_at"
	} else if(ActivityEvents[docType] == true){
		return "synced_at"
	}
	return ""
}

func LeadSquaredTimestampMapping(docType string)([]string, bool) {
	if(docType == LEADSQUARED_LEAD){
		return []string{"CreatedOn", "CreatedOn", "ModifiedOn"}, true
	} else if(ActivityEvents[docType] == true){
		return []string{"CreatedOn", "ModifiedOn"}, true
	}
	return nil, false
}

func LeadSquaredProgramIdMapping(docType string) (string, bool) {
	if(ActivityEvents[docType] == true){
		return "ProspectActivityId", true
	}
	return "", false
}

func LeadSquaredActorTypeMapping(docType string) (string, bool) {
	if(ActivityEvents[docType] == true){
		return LEADSQUARED_LEAD, true
	}
	return "", false
}

func LeadSquaredActorIdMapping(docType string) (string, bool){
	if(ActivityEvents[docType] == true){
		return "RelatedProspectId", true
	}
	return "", false
}

func GetLeadSquaredTypeToAliasMap(aliasType map[string]int) map[int]string {
	typeAlias := make(map[int]string)
	for alias, typ := range aliasType {
		typeAlias[typ] = alias
	}
	return typeAlias
}

func GetLeadSquaredDocumentQuery(bigQueryProjectId string, schemaId string, baseQuery string, executionDate string, docType string, limit int, lastProcessedId int) string {
	if docType == LEADSQUARED_LEAD || ActivityEvents[docType] == true {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, GetLeadSquaredDocumentFilterCondition(docType, false, "", executionDate), lastProcessedId, limit)
	}
	return ""
}

func GetLeadSquaredDocumentActorId(documentType string, data []string, metadataColumns []string) string {
	actorId, exists := LeadSquaredActorIdMapping(documentType)
	if exists {
		dataObjectColumns := GetObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[actorId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

func GetLeadSquaredDocumentFilterCondition(docType string, addPrefix bool, prefix string, executionDate string) string {
	filterColumn := ""
	if addPrefix {
		filterColumn = fmt.Sprintf("%v.%v", prefix, LeadSquaredDataObjectFiltersColumn(docType))
	} else {
		filterColumn = LeadSquaredDataObjectFiltersColumn(docType)
	}
	filterCondition := fmt.Sprintf(LeadSquaredDataObjectFilters(docType), filterColumn, executionDate)

	return filterCondition
}

func GetLeadSquaredDocumentMetadataQuery(docType string, bigQueryProjectId string, schemaId string, baseQuery string) (string, bool) {
	if docType == LEADSQUARED_LEAD || ActivityEvents[docType] == true {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId), true
	}
	return "", false
}

func GetLeadSquaredDocumentDocumentType(documentTypeString string) int {
	docTypeId, exists := LeadSquaredDocumentTypeAlias[documentTypeString]
	if exists {
		return docTypeId
	}
	return 0
}

func GetLeadSquaredObjectDataColumns(docType string, metadataColumns []string) map[string]int {
	dataObjectColumns := make(map[string]int, 0)
	if metadataColumns != nil {
		for index, column := range metadataColumns {
			dataObjectColumns[column] = index
		}
	}
	return dataObjectColumns
}

func GetLeadSquaredActorType(documentTypeString string) int {
	actorType, exists := LeadSquaredActorTypeMapping(documentTypeString)
	if exists {
		actorTypeId, exists_actor := LeadSquaredDocumentTypeAlias[actorType]
		if exists_actor {
			return actorTypeId
		}
		return 0
	}
	return 0
}

func GetLeadSquaredDocumentValues(docType string, data []string, metadataColumns []string, metadataColumnDateTimeType map[string]bool, metadataColumnNumericalType map[string]bool) map[string]interface{} {
	values := make(map[string]interface{})
	dataObjectColumns := GetLeadSquaredObjectDataColumns(docType, metadataColumns)
	for key, index := range dataObjectColumns {
		if metadataColumnDateTimeType != nil && metadataColumnDateTimeType[key] {
			convertedTimestamp := ConvertTimestamp(data[index])
			if convertedTimestamp == 0 {
				values[key] = nil
			} else {
				values[key] = convertedTimestamp
			}
		} else if metadataColumnNumericalType != nil && metadataColumnNumericalType[key] {
			convertedNumber := ConvertToNumber(data[index])
			if convertedNumber == 0 {
				values[key] = nil
			} else {
				values[key] = convertedNumber
			}
		} else {
			if data[index] == "<nil>" {
				values[key] = ""
			} else {
				values[key] = data[index]
			}
		}
	}
	return values
}

func GetLeadSquaredUserId(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := LeadSquaredUserIdMapping[documentType]
	if exists {
		dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[activtyNameId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

func GetLeadSquaredUserAutoId(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := LeadSquaredUserAutoIdMapping(documentType)
	if exists {
		dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[activtyNameId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

func GetLeadSquaredDocumentPhone(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := LeadSquaredPhoneMapping[documentType]
	if exists {
		dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[activtyNameId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

func GetLeadSquaredDocumentEmail(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := LeadSquaredEmailMapping[documentType]
	if exists {
		dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[activtyNameId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

func GetLeadSquaredDocumentAction(documentType string, data []string, metadataColumns []string) CRMAction {
	dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
	created_at_index, exists_created_at_index := dataObjectColumns["CreatedOn"]
	updated_at_index, exists_updated_at_index := dataObjectColumns["ModifiedOn"]
	if !exists_created_at_index || !exists_updated_at_index {
		return 0
	}
	if ConvertTimestamp(data[updated_at_index]) > ConvertTimestamp(data[created_at_index]) {
		return CRMActionUpdated
	} else {
		return CRMActionCreated
	}
}

func GetLeadSquaredDocumentTimestamp(documentType string, data []string, metadataColumns []string) []int64 {
	timestampIds, exists := LeadSquaredTimestampMapping(documentType)
	result := make([]int64, 0)
	for _, timestampId := range timestampIds {
		if exists {
			dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
			index, exists_index := dataObjectColumns[timestampId]
			if exists_index {
				result = append(result, ConvertTimestamp(data[index]))
			}
		}
	}
	return result
}

func GetLeadSquaredDocumentProgramId(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := LeadSquaredProgramIdMapping(documentType)
	if exists {
		dataObjectColumns := GetObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[activtyNameId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

type LeadsByDateRangeRequest struct {
	Parameter ParameterObj `json:"Parameter"`
	Columns   ColumnsObj   `json:"Columns"`
	Paging    PagingObj    `json:"Paging"`
}

type ParameterObj struct {
	FromDate string `json:"FromDate"`
	ToDate   string `json:"ToDate"`
}

type LeadSearchParameterObj struct {
	LookupName  string `json:"LookupName"`
	LookupValue string `json:"LookupValue"`
	SqlOperator string `json:"SqlOperator"`
}

type SalesActivitySearchParameterObj struct {
	FromDate         string `json:"FromDate"`
	ToDate           string `json:"ToDate"`
	ActivityEvent    int    `json:"ActivityEvent"`
	RemoveEmptyValue bool   `json:"RemoveEmptyValue"`
}

type ColumnsObj struct {
	IncludeCSV string `json:"Include_CSV"`
}

type PagingObj struct {
	PageIndex int `json:"PageIndex"`
	PageSize  int `json:"PageSize"`
}

type SortingObj struct {
	ColumnName string `json:"ColumnName"`
	Direction  string `json:"Direction"`
}

type LeadsByDateRangeResponse struct {
	RecordCount int            `json:"RecordCount"`
	Leads       []LeadResponse `json:"Leads"`
}

type LeadResponse struct {
	LeadPropertyList []LeadAttributeValue `json:"LeadPropertyList"`
}

type LeadAttributeValue struct {
	Attribute string `json:"Attribute"`
	Value     string `json:"Value"`
}

type SearchLeadsByCriteriaRequest struct {
	Columns   ColumnsObj             `json:"Columns"`
	Paging    PagingObj              `json:"Paging"`
	Sorting   SortingObj             `json:"Sorting"`
	Parameter LeadSearchParameterObj `json:"Parameter"`
}

type SearchSalesActivityByCriteriaRequest struct {
	Paging    PagingObj                       `json:"Paging"`
	Sorting   SortingObj                      `json:"Sorting"`
	Parameter SalesActivitySearchParameterObj `json:"Parameter"`
}
