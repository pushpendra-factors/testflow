package model

import "time"

type ProjectSetting struct {
	// Foreign key constraint project_id -> projects(id)
	// Used project_id as primary key also, becase of 1-1 relationship.
	ProjectId uint64 `gorm:"primary_key:true" json:"project_id,omitempty"`
	// Using pointers to avoid update by default value.
	// omit empty to avoid nil(filelds not updated) on resp json.
	AutoTrack       *bool `gorm:"not null;default:false" json:"auto_track,omitempty"`
	AutoFormCapture *bool `gorm:"not null;default:false" json:"auto_form_capture,omitempty"`
	ExcludeBot      *bool `gorm:"not null;default:false" json:"exclude_bot,omitempty"`
	// Segment integration settings.
	IntSegment *bool `gorm:"not null;default:false" json:"int_segment,omitempty"`
	// Adwords integration settings.
	// Foreign key constraint int_adwords_enabled_agent_uuid -> agents(uuid)
	// Todo: Set int_adwords_enabled_agent_uuid, int_adwords_customer_account_id to NULL
	// for disabling adwords integration for the project.
	IntAdwordsEnabledAgentUUID  *string `json:"int_adwords_enabled_agent_uuid,omitempty"`
	IntAdwordsCustomerAccountId *string `json:"int_adwords_customer_account_id,omitempty"`
	// Hubspot integration settings.
	IntHubspot       *bool     `gorm:"not null;default:false" json:"int_hubspot,omitempty"`
	IntHubspotApiKey string    `json:"int_hubspot_api_key,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	//Facebook settings
	IntFacebookEmail       string  `json:"int_facebook_email,omitempty"`
	IntFacebookAccessToken string  `json:"int_facebook_access_token,omitempty"`
	IntFacebookAgentUUID   *string `json:"int_facebook_agent_uuid,omitempty"`
	IntFacebookUserID      string  `json:"int_facebook_user_id,omitempty"`
	IntFacebookAdAccount   string  `json:"int_facebook_ad_account,omitempty"`
	// Archival related fields.
	ArchiveEnabled  *bool `gorm:"default:false" json:"archive_enabled"`
	BigqueryEnabled *bool `gorm:"default:false" json:"bigquery_enabled"`
	//Salesforce settings
	IntSalesforceEnabledAgentUUID *string `json:"int_salesforce_enabled_agent_uuid,omitempty"`
	//Linkedin related fields
	IntLinkedinAdAccount          string  `json:"int_linkedin_ad_account"`
	IntLinkedinAccessToken        string  `json:"int_linkedin_access_token"`
	IntLinkedinRefreshToken       string  `json:"int_linkedin_refresh_token"`
	IntLinkedinRefreshTokenExpiry int64   `json:"int_linkedin_refresh_token_expiry"`
	IntLinkedinAccessTokenExpiry  int64   `json:"int_linkedin_access_token_expiry"`
	IntLinkedinAgentUUID          *string `json:"int_linkedin_agent_uuid"`
	IntDrift                      *bool   `gorm:"not null;default:false" json:"int_drift,omitempty"`
}

const ProjectSettingKeyToken = "token"
const ProjectSettingKeyPrivateToken = "private_token"

var projectSettingKeys = [...]string{
	ProjectSettingKeyToken,
	ProjectSettingKeyPrivateToken,
}

// Salesforce required fields per project
var (
	designcafeLeadsAllowedFields = map[string]bool{
		"Id":                                 true,
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
	}

	designcafeOpportunityAllowedFields = map[string]bool{
		"Id":                             true, // require for identification purpose
		"Name":                           true,
		"StageName":                      true,
		"Amount":                         true,
		"ExpectedRevenue":                true,
		"CloseDate":                      true,
		"Type":                           true,
		"LeadSource":                     true,
		"IsWon":                          true,
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
	}

	designcafeAllowedObjects = map[string]map[string]bool{
		SalesforceDocumentTypeNameLead:        designcafeLeadsAllowedFields,
		SalesforceDocumentTypeNameOpportunity: designcafeOpportunityAllowedFields,
	}

	SalesforceProjectStore = map[uint64]map[string]map[string]bool{
		483: designcafeAllowedObjects,
	}
)

type AdwordsProjectSettings struct {
	ProjectId         uint64
	CustomerAccountId string
	AgentUUID         string
	RefreshToken      string
}

type HubspotProjectSettings struct {
	ProjectId uint64 `json:"-"`
	APIKey    string `json:"api_key"`
}

type FacebookProjectSettings struct {
	ProjectId              string `json:"project_id"`
	IntFacebookUserId      string `json:"int_facebook_user_id"`
	IntFacebookAccessToken string `json:"int_facebook_access_token"`
	IntFacebookAdAccount   string `json:"int_facebook_ad_account"`
	IntFacebookEmail       string `json:"int_facebook_email"`
}
type LinkedinProjectSettings struct {
	ProjectId                     string `json:"project_id"`
	IntLinkedinAdAccount          string `json:"int_linkedin_ad_account"`
	IntLinkedinAccessToken        string `json:"int_linkedin_access_token"`
	IntLinkedinRefreshToken       string `json:"int_linkedin_refresh_token"`
	IntLinkedinRefreshTokenExpiry int64  `json:"int_linkedin_refresh_token_expiry"`
	IntLinkedinAccessTokenExpiry  int64  `json:"int_linkedin_access_token_expiry"`
}

// SalesforceProjectSettings contains refresh_token and instance_url for enabled projects
type SalesforceProjectSettings struct {
	ProjectID    uint64 `json:"-"`
	RefreshToken string `json:"refresh_token"`
	InstanceURL  string `json:"instance_url"`
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
	} else {
		docTypes = SalesforceStandardDocumentType
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
