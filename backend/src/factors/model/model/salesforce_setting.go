package model

import (
	C "factors/config"
)

// Salesforce required fields per project
var (
	designcafeLeadsAllowedFields = map[string]bool{
		"Id":                                 true,
		"IsDeleted":                          true,
		"LastName":                           true,
		"Salutation":                         true,
		"Company":                            true,
		"Street":                             true,
		"City":                               true,
		"State":                              true,
		"Phone":                              true,
		"MobilePhone":                        true,
		"Email":                              true,
		"LeadSource":                         true,
		"Status":                             true,
		"ConvertedDate":                      true,
		"ConvertedOpportunityId":             true,
		"CreatedDate":                        true,
		"CreatedById":                        true,
		"LastModifiedDate":                   true,
		"LastModifiedById":                   true,
		"LastActivityDate":                   true,
		"DC_Lead_Status__c":                  true,
		"Alternate_Contact_Number__c":        true,
		"Approx_Budget__c":                   true,
		"Call_Center_Agent__c":               true,
		"Call_Count__c":                      true,
		"Call_Stage__c":                      true,
		"CMM_Name__c":                        true,
		"Channel__c":                         true,
		"DC_Lead_Source__c":                  true,
		"DSA__c":                             true,
		"DSA_Category__c":                    true,
		"DSAname__c":                         true,
		"Designer__c":                        true,
		"First_Date_of_Contact__c":           true,
		"Follow_Up_Count__c":                 true,
		"Follow_Up_Date_Time__c":             true,
		"GClid__c":                           true,
		"Home_Type__c":                       true,
		"Interior_work_needed_for__c":        true,
		"Is_Designer_Assigned__c":            true,
		"Last_Source__c":                     true,
		"Lead_Allocation_Time__c":            true,
		"Lead_Owner_Name__c":                 true,
		"Lead_Owner_Region__c":               true,
		"Lead_Owner_Team__c":                 true,
		"Lead_Qualified_Date__c":             true,
		"CMM_Team__c":                        true,
		"Has_Designer_Accepted__c":           true,
		"Meeting_Scheduled_on_First_Call__c": true,
		"Meeting_Type__c":                    true,
		"Page_URL__c":                        true,
		"Pre_Qualified_Date__c":              true,
		"Project_Name__c":                    true,
		"Region__c":                          true,
		"Requirement_Details__c":             true,
		"Source_Journey__c":                  true,
		"Source__c":                          true,
		"Lead_Owner_Role__c":                 true,
		"Property_Type__c":                   true,
		"Willingness_For_Meeting__c":         true,
		"Ad_Group__c":                        true,
		"Ad_Name__c":                         true,
		"Ad_Network__c":                      true,
		"Budget__c":                          true,
		"Campagin__c":                        true,
		"DC_Campaign_Source__c":              true,
		"Device_Type__c":                     true,
		"EC_Location__c":                     true,
		"Enquiry_ID__c":                      true,
		"IP_Address__c":                      true,
		"Meeting_Venue__c":                   true,
		"Page_URL_1__c":                      true,
		"Property_Possession_Date__c":        true,
		"Qualified_Status__c":                true,
		"Recontacted__c":                     true,
		"Affiliate_Name__c":                  true,
		"first_date_of_contact_to_Qualified_c__c": true,
		"Keyword__c":                        true,
		"Custom_Lead_ID__c":                 true,
		"Pre_Qualified_FollowUp_Date__c":    true,
		"Date_When_Meeting_is_Scheduled__c": true,
		"Mobile_Number_External_Field__c":   true,
		"Country_Code__c":                   true,
		"OTP_Verified__c":                   true,
		"User_Mobile__c":                    true,
		"Messaging_Source__c":               true,
		"Entry_Url__c":                      true,
		"First_Visit_Time__c":               true,
		"IPAddress__c":                      true,
		"Page__c":                           true,
		"Time_On_Last_page__c":              true,
		"User_Browser__c":                   true,
		"User_Last_Url__c":                  true,
		"User_OS__c":                        true,
		"Lockdown_Survey__c":                true,
		"Civil_Work__c":                     true,
		"Whatsapp_Opt_IN__c":                true,
		"MobileYM__c":                       true,
		"Customer_WhatsApp_OptIN__c":        true,
		"Design_User_Name__c":               true,
		"Design_User__c":                    true,
		"Industry":                          true,
		"Old_Source__c":                     true,
		"Client_Site_visit__c":              true,
		"Client_s_Budget__c":                true,
		"DC_Home_Visit__c":                  true,
		"Property_Name__c":                  true,
	}

	designcafeOpportunityAllowedFields = map[string]bool{
		"Id":                             true, // require for identification purpose
		"IsDeleted":                      true,
		"Name":                           true,
		"StageName":                      true,
		"Amount":                         true,
		"ExpectedRevenue":                true,
		"CloseDate":                      true,
		"UnitType":                       true,
		"LeadSource":                     true,
		"IsWon":                          true,
		"OwnerId":                        true,
		"CreatedDate":                    true,
		"LastModifiedDate":               true,
		"Budget_Confirmed__c":            true,
		"Affiliate_Name__c":              true,
		"Loss_Reason__c":                 true,
		"Call_Center_Agent__c":           true,
		"Call_Stage__c":                  true,
		"CMM_Team__c":                    true,
		"Channel__c":                     true,
		"Client_s_Budget__c":             true,
		"Signup_Amount__c":               true,
		"ConvertedLeadId__c":             true,
		"DSA__c":                         true,
		"Designer__c":                    true,
		"Home_Type__c":                   true,
		"Lead_Stage__c":                  true,
		"Lead_Status__c":                 true,
		"Meeting_Date_and_Time__c":       true,
		"Meeting_Scheduled_Date_Time__c": true,
		"Meeting_Type__c":                true,
		"Meeting_Venue__c":               true,
		"Offer_and_discounts__c":         true,
		"Possession_Status__c":           true,
		"Project_Name__c":                true,
		"Property_Address__c":            true,
		"Proposed_Budget__c":             true,
		"CMM_Name__c":                    true,
		"Region__c":                      true,
		"Source__c":                      true,
		"Total_Amount__c":                true,
		"Approx_Budget__c":               true,
		"Ad_Group__c":                    true,
		"Ad_Name__c":                     true,
		"Campaign__c":                    true,
		"Enquiry_ID__c":                  true,
		"Overall_Sales_Duration__c":      true,
		"Campaign_Source__c":             true,
		"DC_Lead_Source__c":              true,
		"Reason_for_Loss__c":             true,
		"Closing_Offer__c":               true,
		"Mobile__c":                      true,
		"Wohoo_Card__c":                  true,
		"Customer_ID__c":                 true,
		"Phone__c":                       true,
		"Messaging_Source__c":            true,
		"Payment_Mode__c":                true,
		"Packages__c":                    true,
		"MobileYM__c":                    true,
		"Design_User_Name__c":            true,
		"Design_User__c":                 true,
		"Opportunity_status__c":          true,
		"CampaignId":                     true,
		"LastActivityDate":               true,
		"DSA_Code__c":                    true,
		"Interior_work_needed_for__c":    true,
		"Lead_Id__c":                     true,
		"Referral_Code__c":               true,
		"Referred_By__c":                 true,
		"Requirement_Details__c":         true,
		"ST_Referee_Name__c":             true,
		"ST_Referee_Code__c":             true,
		"Whatsapp_Opt_IN__c":             true,
		"RUP_Signup_Amount__c":           true,
		"Proposal_Sent_Date__c":          true,
		"Alternate_Phone__c":             true,
		"DC_Home_Visit__c":               true,
		"Proposed_Value_Dis_Incl_GST__c": true,
		"Reason__c":                      true,
		"Campaign_ID__c":                 true,
		"Ad_Group_ID__c":                 true,
		"Followup_Date_Time__c":          true,
		"Referee_Email_ID__c":            true,
		"Country_Code__c":                true,
	}

	designcafeAllowedObjects = map[string]map[string]bool{
		SalesforceDocumentTypeNameLead:        designcafeLeadsAllowedFields,
		SalesforceDocumentTypeNameOpportunity: designcafeOpportunityAllowedFields,
	}

	SalesforceProjectStore = map[uint64]map[string]map[string]bool{
		483: designcafeAllowedObjects,
	}

	// salesforceStandardIndentificationField standard field for salesforce identification
	salesforceStandardIndentificationField = map[string]map[string][]string{
		SalesforceDocumentTypeNameAccount: {
			IdentificationTypePhone: {"Phone", "PersonMobilePhone"},
			//Account don't have "Email" as standard field, "PersonEmail" and "PersonMobilePhone" is only used when PersonAccout is activate
			//https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/sforce_api_guidelines_personaccounts.htm
			IdentificationTypeEmail: {"PersonEmail"},
		},
		SalesforceDocumentTypeNameContact: {
			IdentificationTypeEmail: {"Email"},
			IdentificationTypePhone: {"Phone", "MobilePhone"},
		},
		SalesforceDocumentTypeNameLead: {
			IdentificationTypeEmail: {"Email"},
			IdentificationTypePhone: {"Phone", "MobilePhone"},
		},
	}

	/*
	 Custom field per customer per object
	 Will overwrite indentity order precedence
	*/
	DesignCafeIdentificationField = map[string][]string{
		SalesforceDocumentTypeNameLead:        {"MobilePhone", "MobileYM__c"},
		SalesforceDocumentTypeNameOpportunity: {"Mobile__c", "MobileYM__c", "Phone__c"},
	}

	MoEngageIdentificationField = map[string][]string{
		SalesforceDocumentTypeNameOpportunity: {"Billing_Contact_Email_ID__c"},
	}

	MercosIdentificationField = map[string][]string{
		SalesforceDocumentTypeNameOpportunity: {"Email__c"},
	}

	SalesforceProjectIdentificationFieldStore = map[uint64]map[string][]string{
		483: DesignCafeIdentificationField,
		566: MoEngageIdentificationField,
		587: MercosIdentificationField,
	}
)

//GetSalesforceEmailFieldByProjectIDAndObjectName return email indentification field by project, returns standard identification field if no configuration found
func GetSalesforceEmailFieldByProjectIDAndObjectName(projectID uint64, objectName string, projectIdentificationFieldStore *map[uint64]map[string][]string) []string {
	if projectID == 0 || objectName == "" {
		return nil
	}

	if objects, exist := (*projectIdentificationFieldStore)[projectID]; exist {
		if fields, exist := objects[objectName]; exist {
			allFields := append(fields, salesforceStandardIndentificationField[objectName][IdentificationTypeEmail]...)
			return allFields
		}
	}
	return salesforceStandardIndentificationField[objectName][IdentificationTypeEmail]
}

//GetSalesforcePhoneFieldByProjectIDAndObjectName return phone indentification field by project, returns standard identification field if no configuration found
func GetSalesforcePhoneFieldByProjectIDAndObjectName(projectID uint64, objectName string, projectIdentificationFieldStore *map[uint64]map[string][]string) []string {
	if projectID == 0 || objectName == "" {
		return nil
	}

	if objects, exist := (*projectIdentificationFieldStore)[projectID]; exist {
		if fields, exist := objects[objectName]; exist {
			allFields := append(fields, salesforceStandardIndentificationField[objectName][IdentificationTypePhone]...)
			return allFields
		}
	}

	return salesforceStandardIndentificationField[objectName][IdentificationTypePhone]
}

// GetSalesforceAllowedObjects return allowed object type for a project
func GetSalesforceAllowedObjects(projectID uint64) []int {
	var docTypes []int
	if projectID == 0 {
		return docTypes
	}

	if objects, exist := SalesforceProjectStore[projectID]; exist {
		for name := range objects {
			docType := GetSalesforceDocTypeByAlias(name)
			docTypes = append(docTypes, docType)
		}

		return docTypes
	}

	docTypes = SalesforceStandardDocumentType
	if C.IsAllowedCampaignEnrichementByProjectID(projectID) {
		docTypes = append(docTypes, SalesforceCampaignDocuments...)
	}

	if C.IsAllowedSalesforceGroupsByProjectID(projectID) {
		docTypes = append(docTypes, SalesforceDocumentTypeOpportunityContactRole)
	}

	return docTypes
}

// GetSalesforceAllowedfiedsByObject return list of allowed field for a project
func GetSalesforceAllowedfiedsByObject(projectID uint64, objectName string) map[string]bool {
	if projectID == 0 {
		return nil
	}

	if objects, exist := SalesforceProjectStore[projectID]; exist {
		if _, objExist := objects[objectName]; objExist {
			return objects[objectName]
		}
	}

	return nil
}
