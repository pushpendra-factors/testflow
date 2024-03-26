package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var jsonResponse string = `
{
	    templates: [
	        {
	            title: 'Add identified companies to HubSpot companies.',
	            short_desc: 'Segment is a Customer Data Platform (CDP) that simplifies collecting and using data from the users of your digital properties and SaaS applications',
	            long_desc: '<p>Segment is a Customer Data Platform (CDP) that simplifies collecting and using data from the users of your digital properties and SaaS applications<p>',
	            image: 'factors.ai/../../image_URL.png',
	            category: ['Hubspot'],
	            tags: ['Website session', 'Company country','Company revenue range'],
	            alert_config: {
	                published: false,
	                draft: true,
	                title: 'saved title name',
	                desr: 'saved desc',
	                integrations: ['apollo', 'hubspot'],
	                trigger: {
	                        "event": "$session",
	                        "event_level": "user",
	                        "breakdown_properties": [
	                            {
	                                "en": "user",
	                                "ena": "$session",
	                                "pr": "$city",
	                                "pty": "categorical"
	                            }
	                        ],
	                        "filter": [
	                            {
	                                "en": "user",
	                                "grpn": "user",
	                                "lop": "AND",
	                                "op": "equals",

	                                "pr": "$hubspot_contact_hs_analytics_source_data_1",
	                                "ty": "categorical",
	                                "va": "API"
	                            },
	                            {
	                                "en": "user",
	                                "grpn": "user",
	                                "lop": "OR",
	                                "op": "equals",
	                                "pr": "$hubspot_contact_hs_analytics_source_data_1",
	                                "ty": "categorical",
	                                "va": "Auto-tagged PPC"
	                            },
	                        ],
	                },
	                expected_payload: {
	                    "mandatory_props_company": {
	                      "company_name": "company_name_value",
	                      "company_domain": "company_domain_value"
	                    },
	                    "additional_props_company": {
	                      "hs_comp_field_1": "Fa_value_1",
	                      "hs_comp_field_2": "Fa_value_2",
	                      "hs_comp_field_3": "Fa_value_3",
	                      "hs_comp_field_n": "Fa_value_n"
	                    },
	                    "apollo_config": {
	                      "api_key": "user_api_key",
	                      "job_titles": "array_of_job_titles",
	                      "job_seniorities": "array_of_job_seniorities",
	                      "max_contacts": "number_of_contacts_per_company"
	                    },
	                    "additional_props_contact_mapping": {
	                      "first_name": "hs_first_name_field",
	                      "last_name": "hs_last_name_field",
	                      "full_name": "hs_full_name_field",
	                      "job_title": "hs_job_title_field",
	                      "job_seniority": "hs_job_seniority_field",
	                      "country": "hs_country_field",
	                      "state": "hs_state_field",
	                      "city": "hs_city_field",
	                      "linkedin_url": "hs_linkedin_url_field"
	                    }
	                  },

	            }

	        },
	    ]
	}`

func GetAllWorkflowTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, jsonResponse)
}

func GetAllSavedWorkflows(c *gin.Context) (interface{}, int, string, string, bool) {
	return nil, http.StatusOK, "", "stub-api", false
}
