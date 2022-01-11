package model

import (
	C "factors/config"
)

// Salesforce required fields per project
var (
	designcafeAllowedObjects = map[string]map[string]bool{
		SalesforceDocumentTypeNameLead:        nil, // nil mapping will allow all properties
		SalesforceDocumentTypeNameOpportunity: nil,
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
		if _, objExist := objects[objectName]; objExist && objects[objectName] != nil {
			return objects[objectName]
		}
	}

	return nil
}
