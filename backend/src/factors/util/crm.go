package util

import "strings"

type CRMSource int

const (
	CRM_SOURCE_HUBSPOT          CRMSource = 1
	CRM_SOURCE_SALESFORCE       CRMSource = 2
	CRM_SOURCE_MARKETO          CRMSource = 3
	CRM_SOURCE_LEADSQUARED      CRMSource = 4
	CRM_SOURCE_NAME_HUBSPOT               = "hubspot"
	CRM_SOURCE_NAME_SALESFORCE            = "salesforce"
	CRM_SOURCE_NAME_MARKETO               = "marketo"
	CRM_SOURCE_NAME_LEADSQUARED           = "leadsquared"
)

// List of prefix to differentiate CRM property from other properties. Only properties with prefix will overwritten by CRM
var CRMUserPropertiesOverwritePrefixes = map[string]string{
	CRM_SOURCE_NAME_HUBSPOT:     HUBSPOT_PROPERTY_PREFIX,
	CRM_SOURCE_NAME_SALESFORCE:  SALESFORCE_PROPERTY_PREFIX,
	CRM_SOURCE_NAME_MARKETO:     MARKETO_PROPERTY_PREFIX,
	CRM_SOURCE_NAME_LEADSQUARED: LEADSQUARED_PROPERTY_PREFIX,
}

var SourceCRM = map[string]int{
	CRM_SOURCE_NAME_HUBSPOT:     1,
	CRM_SOURCE_NAME_SALESFORCE:  2,
	CRM_SOURCE_NAME_MARKETO:     3,
	CRM_SOURCE_NAME_LEADSQUARED: 4,
}

func IsCRMPropertyKeyBySource(source, key string) bool {
	prefix := CRMUserPropertiesOverwritePrefixes[source]
	if prefix == "" {
		return false
	}

	return strings.HasPrefix(key, prefix)
}

func IsCRMPropertyKey(key string) bool {
	for _, prefix := range CRMUserPropertiesOverwritePrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	return false
}

func IsCRM(source string) bool {
	for crmSource, _ := range SourceCRM {
		if crmSource == source {
			return true
		}
	}
	return false
}
