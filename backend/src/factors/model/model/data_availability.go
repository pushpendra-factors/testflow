package model

import (
	"fmt"
	"time"
)

/* Steps to add new integration
1. Add the integration in the const list
2. Add the integrations in INTEGATIONS map
3. Add the corresponding query in INTEGRATIONS_QUERY map and the only variable parameter in the query should be project_id
4. return object from the query should be one value (timestamp either in unix() format or unixmilli() format or yyyyMMdd format)
- add the condition  to transformTimestampValue() method according to the return value format
5. Add if integration exist for that projectId in the IsIntegrationAvailable() method
*/
/*
This has support for 2 methods
1. GetLatestDataStatus([]string{"*"}, project.ID, false) - Parameters are first one is the list of integrations, projectId, HardRefresh
2. IsDataAvailable(project_id, model.HUBSPOT, 1653865933) - project_id, integration, timestamp for which data is checked for
integration constants support one of the following
3. Add DataAvailabilityExpiry to the config - default is 30 min

Eg: If PullEvents job waits for hubspot and salesforce to be done before starting it for that week, then Call IsDataAvailable(1, 'hubspot', <end of week timestamp>)
This will return a true or false about if its done till that time or not

Other option is to directly call GetLatestDataStatus([]string{"*"}, 1, false]) which returns what is the latest date for which the data is available for that project.
*/
const (
	HUBSPOT        = "hubspot"
	SALESFORCE     = "salesforce"
	MARKETO        = "marketo"
	GOOGLE_ORGANIC = "google-organic"
	ADWORDS        = "adwords"
	LINKEDIN       = "linkedin"
	FACEBOOK       = "facebook"
	BINGADS        = "bingads"
	SESSION        = "session"
)

var INTEGRATIONS_QUERY = map[string]string{
	HUBSPOT:        fmt.Sprintf("select max(timestamp) from hubspot_documents where type in (%v, %v, %v, %v) and project_id = ? and synced=true", HubspotDocumentTypeCompany, HubspotDocumentTypeContact, HubspotDocumentTypeDeal, HubspotDocumentTypeEngagement),
	SALESFORCE:     fmt.Sprintf("select max(timestamp) from salesforce_documents where type not in (%v, %v) and project_id = ? and synced=true", SalesforceDocumentTypeCampaign, SalesforceDocumentTypeOpportunityContactRole),
	MARKETO:        "select max(timestamp) from crm_users where synced = true and project_id = ? and source = 3",
	GOOGLE_ORGANIC: "select max(timestamp) from google_organic_documents where project_id = ?",
	ADWORDS:        "select max(timestamp) from adwords_documents where project_id = ?",
	LINKEDIN:       "select max(timestamp) from linkedin_documents where project_id = ?",
	FACEBOOK:       "select max(timestamp) from facebook_documents where project_id = ?",
	BINGADS:        "select max(timestamp) from integration_documents where project_id = ? and source = 'bingads'",
	SESSION:        "select json_extract_bigint(jobs_metadata,'next_session_start_timestamp') from projects where id = ?",
}

var INTEGRATIONS = map[string]bool{
	HUBSPOT:        true,
	SALESFORCE:     true,
	MARKETO:        true,
	GOOGLE_ORGANIC: true,
	ADWORDS:        true,
	LINKEDIN:       true,
	FACEBOOK:       true,
	BINGADS:        true,
	SESSION:        true,
}

type DataAvailability struct {
	ProjectID           uint64    `gorm:"primary_key:true" json:"project_id"`
	Integration         string    `json:"integration"`
	LatestDataTimestamp int64     `json:"latest_data_timestamp"`
	LastPolled          time.Time `json:"last_polled"`
	Source              string    `json:"source"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type DataAvailabilityStatus struct {
	IntegrationStatus bool   `json:"integration_status`
	LatestData        uint64 `json:"lastest_data"`
	Message           string `json:"message"`
}
