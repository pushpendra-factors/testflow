export const Stub =  {
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
}


export const StubOld = 
[{
  "v": "v0 Baliga",
  "id": 13,
  "title": "Factors → HubSpot company",
  "alert": {
    "title": "Factors → HubSpot company",
    "alert_message": "Factors → HubSpot company",
    "alert_name": "An account visited pricing page",
    "description": "Get alerts whenever one of your accounts check out your pricing page so that reach out in real time.",
    "payload_props": {},
    "prepopulate": {},
  },
  "template_constants": {
    "categories": [
      "Hubspot"
    ],
  },
  "workflow_config": {
    published: false,
    draft: true,
    title: 'saved title name',
    desr: 'saved desc',
    integrations: ['hubspot'],
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
    "payload": {
        "mandatory_props_company": [
          {
            factors: 'company-name',
            others: 'HS-company-name'
          },
          {
            factors: 'domain-name',
            others: 'HS-domain-name'
          },
        ],
        "additional_props_company": [
          {
            factors: 'Fa_value_1',
            others: 'hs_comp_field_1'
          },
          {
            factors: 'Fa_value_2',
            others: 'hs_comp_field_2'
          },
        ],
        "mapping_details" : {
          others: 'hubspot'
        },

      },

},
  "is_deleted": false,
  "created_at": "2024-03-21T00:00:00Z",
  "updated_at": "2024-03-21T00:00:00Z"
}]



export const filterOptions = [
  {
    "iconName": "BullsEyePointer",
    "label": "All Accounts",
    "values": [
      {
        "value": "$visited_website",
        "label": "Visited Website",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$domains"
        }
      },
      {
        "value": "$in_g2",
        "label": "Visited G2",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$domains"
        }
      },
      {
        "value": "$in_hubspot",
        "label": "In Hubspot",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$domains"
        }
      },
      {
        "value": "$in_linkedin",
        "label": "Engaged on LinkedIn",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$domains"
        }
      },
      {
        "value": "$in_salesforce",
        "label": "In Salesforce",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$domains"
        }
      },
      {
        "value": "$engagement_level",
        "label": "Engagement Level",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$domains"
        }
      },
      {
        "value": "$top_enagagement_signals",
        "label": "Top Engagement Signals",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$domains"
        }
      },
      {
        "value": "$domain_name",
        "label": "Company ID",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$domains"
        }
      },
      {
        "value": "$engagement_score",
        "label": "Engagement Score",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$domains"
        }
      },
      {
        "value": "$total_enagagement_score",
        "label": "Total Engagement Score",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$domains"
        }
      }
    ]
  },
  {
    "iconName": "linkedin_ads",
    "label": "Linkedin Company Engagements",
    "values": [
      {
        "value": "$li_preferred_country",
        "label": "Li Preferred Country",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$linkedin_company"
        }
      },
      {
        "value": "$li_vanity_name",
        "label": "Li Vanity Name",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$linkedin_company"
        }
      },
      {
        "value": "$li_domain",
        "label": "Li Domain",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$linkedin_company"
        }
      },
      {
        "value": "$li_headquarter",
        "label": "Li Headquarter",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$linkedin_company"
        }
      },
      {
        "value": "$li_localized_name",
        "label": "Li Localized Name",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$linkedin_company"
        }
      },
      {
        "value": "$li_org_id",
        "label": "Li Org Id",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$linkedin_company"
        }
      },
      {
        "value": "$li_total_ad_click_count",
        "label": "Li Total Ad Click Count",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$linkedin_company"
        }
      },
      {
        "value": "$li_total_ad_view_count",
        "label": "Li Total Ad View Count",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$linkedin_company"
        }
      }
    ]
  },
  {
    "iconName": "hubspot_ads",
    "label": "Hubspot Deals",
    "values": [
      {
        "value": "$hubspot_deal_hs_object_source_id",
        "label": "Hubspot Deal Record Creation Source ID",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_is_closed_won",
        "label": "Hubspot Deal Is Closed Won",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_object_source",
        "label": "Hubspot Deal Record Creation Source",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_num_associated_deal_splits",
        "label": "Hubspot Deal Hs Num Associated Deal Splits",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_num_associated_active_deal_registrations",
        "label": "Hubspot Deal Hs Num Associated Active Deal Registrations",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_object_source_label",
        "label": "Hubspot Deal Record Source",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_num_associated_deal_registrations",
        "label": "Hubspot Deal Hs Num Associated Deal Registrations",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_is_deal_split",
        "label": "Hubspot Deal Hs Is Deal Split",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_is_active_shared_deal",
        "label": "Hubspot Deal Hs Is Active Shared Deal",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_pipeline",
        "label": "Hubspot Deal Pipeline",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_is_closed",
        "label": "Hubspot Deal Is Deal Closed?",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_dealstage",
        "label": "Hubspot Deal Deal Stage",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_user_ids_of_all_owners",
        "label": "Hubspot Deal User IDs Of All Owners",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_all_owner_ids",
        "label": "Hubspot Deal All Owner Ids",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_dealname",
        "label": "Hubspot Deal Deal Name",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_rocketlane_deal_stage",
        "label": "Hubspot Deal Rocketlane Deal Stage",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hubspot_owner_id",
        "label": "Hubspot Deal Deal Owner",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_latest_source_data_1",
        "label": "Hubspot Deal Latest Source Data 1",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_source",
        "label": "Hubspot Deal Original Source",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_latest_source",
        "label": "Hubspot Deal Latest Source",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_source_data_1",
        "label": "Hubspot Deal Original Source Drill-Down 1",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_latest_source_company",
        "label": "Hubspot Deal Latest Source Company",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_latest_source_data_1_company",
        "label": "Hubspot Deal Latest Source Data 1 Company",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_product",
        "label": "Hubspot Deal Product",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_latest_source_contact",
        "label": "Hubspot Deal Latest Source Contact",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_latest_source_data_1_contact",
        "label": "Hubspot Deal Latest Source Data 1 Contact",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_sales_owner",
        "label": "Hubspot Deal Sales Owner",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_source_data_2",
        "label": "Hubspot Deal Original Source Drill-Down 2",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_latest_source_data_2_company",
        "label": "Hubspot Deal Latest Source Data 2 Company",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_latest_source_data_2",
        "label": "Hubspot Deal Latest Source Data 2",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_latest_source_data_2_contact",
        "label": "Hubspot Deal Latest Source Data 2 Contact",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_deal_score",
        "label": "Hubspot Deal Hs Deal Score",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_was_imported",
        "label": "Hubspot Deal Performed In An Import",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_inbound_outbound",
        "label": "Hubspot Deal Deal Source",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_contract_duration",
        "label": "Hubspot Deal Contract Duration",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_contract_file",
        "label": "Hubspot Deal Contract File",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_renewal_frequency",
        "label": "Hubspot Deal Renewal Frequency",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_secondary_cs",
        "label": "Hubspot Deal Secondary CS",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_trial_type",
        "label": "Hubspot Deal Trial Type",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_value_delivered_checklist",
        "label": "Hubspot Deal Value Delivered Checklist",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_webhooks_setup_",
        "label": "Hubspot Deal Webhooks Setup?",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_slack___ms_team_alerts_setup_",
        "label": "Hubspot Deal Slack / MS Team Alerts Setup?",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_sales_follow_up_status",
        "label": "Hubspot Deal Sales Follow Up Status",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_linkedin_ads_integrated_",
        "label": "Hubspot Deal LinkedIn Ads Integrated?",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_use_cases",
        "label": "Hubspot Deal Use-Cases",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_notes___only_for_sachin",
        "label": "Hubspot Deal Notes - Only For Sachin",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_october_comments__only_for_sachin_",
        "label": "Hubspot Deal October Comments (Only For Sachin)",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_september_comments___only_for_sachin_",
        "label": "Hubspot Deal September Comments ( Only For Sachin) ",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_last_invoice_memo",
        "label": "Hubspot Deal Last Invoice Memo",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_dealtype",
        "label": "Hubspot Deal Deal Type",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_nov_comments___only_for_sachin_",
        "label": "Hubspot Deal Nov Comments ( Only For Sachin)",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_priority",
        "label": "Hubspot Deal Priority",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_all_team_ids",
        "label": "Hubspot Deal All Team Ids",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_all_accessible_team_ids",
        "label": "Hubspot Deal All Teams",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_merged_object_ids",
        "label": "Hubspot Deal Merged Deal IDs",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hubspot_team_id",
        "label": "Hubspot Deal HubSpot Team",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_outbound_source",
        "label": "Hubspot Deal Outbound Source",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_notes_next_activity_type",
        "label": "Hubspot Deal Next Activity Type",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_notes_next_activity",
        "label": "Hubspot Deal Hs Notes Next Activity",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_user_ids_of_all_notification_followers",
        "label": "Hubspot Deal User IDs Of All Notification Followers",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_object_source_detail_3",
        "label": "Hubspot Deal Record Source Detail 3",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_object_source_detail_1",
        "label": "Hubspot Deal Record Source Detail 1",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_object_source_detail_2",
        "label": "Hubspot Deal Record Source Detail 2",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_lastmodifieddate",
        "label": "Hubspot Deal Last Modified Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_createdate",
        "label": "Hubspot Deal HubSpot Create Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_createdate",
        "label": "Hubspot Deal Create Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_today_s_date",
        "label": "Hubspot Deal Today'S Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_entered_deal_stage",
        "label": "Hubspot Deal Entered Deal Stage",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hubspot_owner_assigneddate",
        "label": "Hubspot Deal Owner Assigned Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_latest_source_timestamp",
        "label": "Hubspot Deal Latest Source Timestamp",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_analytics_latest_source_timestamp_contact",
        "label": "Hubspot Deal Latest Source Timestamp Contact",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_closedate",
        "label": "Hubspot Deal Close Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_notes_last_updated",
        "label": "Hubspot Deal Last Activity Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_notes_last_contacted",
        "label": "Hubspot Deal Last Contacted",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_entered_11609175",
        "label": "Hubspot Deal Date Entered \"Trial Started (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_sales_email_last_replied",
        "label": "Hubspot Deal Recent Sales Email Replied Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_time_in__x__deal_stage",
        "label": "Hubspot Deal Time In 'X' Deal Stage",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_latest_meeting_activity",
        "label": "Hubspot Deal Latest Meeting Activity",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_exited_11609175",
        "label": "Hubspot Deal Date Exited \"Trial Started (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_entered_presentationscheduled",
        "label": "Hubspot Deal Date Entered \"SQL (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_free_trial_start_date",
        "label": "Hubspot Deal Free Trial Start Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_entered_closedlost",
        "label": "Hubspot Deal Date Entered \"Closed Lost (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_exited_presentationscheduled",
        "label": "Hubspot Deal Date Exited \"SQL (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_entered_12728897",
        "label": "Hubspot Deal Date Entered \"Trial Blocked (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_notes_next_activity_date",
        "label": "Hubspot Deal Next Activity Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_contract_expiry_date",
        "label": "Hubspot Deal Contract Expiry Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_invoicing_reminder_date__for_sachin_",
        "label": "Hubspot Deal Invoicing Reminder Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_contract_start_date___sachin__",
        "label": "Hubspot Deal Contract Start Date ( Sachin ) ",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_entered_12000581",
        "label": "Hubspot Deal Date Entered \"Contract/Pricing Negotiation (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_exited_12000581",
        "label": "Hubspot Deal Date Exited \"Contract/Pricing Negotiation (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_entered_closedwon",
        "label": "Hubspot Deal Date Entered \"Closed Won (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_entered_70995281",
        "label": "Hubspot Deal Date Entered \"Paid Trial (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_renewal_date___only_for_sachin_",
        "label": "Hubspot Deal Renewal Date ( Only For Sachin) ",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_paid_trial_start_date",
        "label": "Hubspot Deal Paid Trial Start Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_contract_start_date",
        "label": "Hubspot Deal Contract Start Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_closed_won_date",
        "label": "Hubspot Deal Closed Won Date (Internal)",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_contract_pricing_negotiation_start_date",
        "label": "Hubspot Deal Contract/Pricing Negotiation Start Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_entered_decisionmakerboughtin",
        "label": "Hubspot Deal Date Entered \"Value Delivered (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_exited_decisionmakerboughtin",
        "label": "Hubspot Deal Date Exited \"Value Delivered (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_exited_12728897",
        "label": "Hubspot Deal Date Exited \"Trial Blocked (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_exited_70995281",
        "label": "Hubspot Deal Date Exited \"Paid Trial (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_exited_closedwon",
        "label": "Hubspot Deal Date Exited \"Closed Won (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_entered_77060243",
        "label": "Hubspot Deal Date Entered \"Churned (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_exited_closedlost",
        "label": "Hubspot Deal Date Exited \"Closed Lost (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_value_delivered_start_date",
        "label": "Hubspot Deal Value Delivered Start Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_engagements_last_meeting_booked",
        "label": "Hubspot Deal Date Of Last Meeting Booked In Meetings Tool",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_date_exited_77060243",
        "label": "Hubspot Deal Date Exited \"Churned (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_object_id",
        "label": "Hubspot Deal Record ID",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_closed_amount",
        "label": "Hubspot Deal Closed Deal Amount",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_num_of_associated_line_items",
        "label": "Hubspot Deal Number Of Associated Line Items",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_created_by_user_id",
        "label": "Hubspot Deal Created By User ID",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_days_to_close",
        "label": "Hubspot Deal Days To Close",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_object_source_user_id",
        "label": "Hubspot Deal Record Creation Source User ID",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_num_target_accounts",
        "label": "Hubspot Deal Number Of Target Accounts",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_closed_amount_in_home_currency",
        "label": "Hubspot Deal Closed Deal Amount In Home Currency",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_projected_amount_in_home_currency",
        "label": "Hubspot Deal Weighted Amount In Company Currency",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_projected_amount",
        "label": "Hubspot Deal Weighted Amount",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_updated_by_user_id",
        "label": "Hubspot Deal Updated By User ID",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_deal_stage_probability",
        "label": "Hubspot Deal Deal Probability",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_deal_stage_probability_shadow",
        "label": "Hubspot Deal Deal Stage Probability Shadow",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_num_associated_contacts",
        "label": "Hubspot Deal Number Of Associated Contacts",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_is_open_count",
        "label": "Hubspot Deal Is Open (Numeric)",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_time_differennce",
        "label": "Hubspot Deal Time In Stage",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_days_to_close_raw",
        "label": "Hubspot Deal Days To Close (Without Rounding)",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_amount",
        "label": "Hubspot Deal Amount",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_forecast_amount",
        "label": "Hubspot Deal Forecast Amount",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_amount_in_home_currency",
        "label": "Hubspot Deal Amount In Company Currency",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_num_notes",
        "label": "Hubspot Deal Number Of Sales Activities",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_num_contacted_notes",
        "label": "Hubspot Deal Number Of Times Contacted",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_latest_time_in_11609175",
        "label": "Hubspot Deal Latest Time In \"Trial Started (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_cumulative_time_in_11609175",
        "label": "Hubspot Deal Cumulative Time In \"Trial Started (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_closed_won_count",
        "label": "Hubspot Deal Is Closed Won (Numeric)",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_latest_time_in_presentationscheduled",
        "label": "Hubspot Deal Latest Time In \"SQL (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_cumulative_time_in_presentationscheduled",
        "label": "Hubspot Deal Cumulative Time In \"SQL (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_cumulative_time_in_12000581",
        "label": "Hubspot Deal Cumulative Time In \"Contract/Pricing Negotiation (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_latest_time_in_12000581",
        "label": "Hubspot Deal Latest Time In \"Contract/Pricing Negotiation (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_latest_time_in_decisionmakerboughtin",
        "label": "Hubspot Deal Latest Time In \"Value Delivered (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_cumulative_time_in_decisionmakerboughtin",
        "label": "Hubspot Deal Cumulative Time In \"Value Delivered (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_cumulative_time_in_12728897",
        "label": "Hubspot Deal Cumulative Time In \"Trial Blocked (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_latest_time_in_12728897",
        "label": "Hubspot Deal Latest Time In \"Trial Blocked (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_factors_app_project_id",
        "label": "Hubspot Deal Factors App Project ID",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_latest_time_in_70995281",
        "label": "Hubspot Deal Latest Time In \"Paid Trial (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_cumulative_time_in_70995281",
        "label": "Hubspot Deal Cumulative Time In \"Paid Trial (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_latest_time_in_closedwon",
        "label": "Hubspot Deal Latest Time In \"Closed Won (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_cumulative_time_in_closedwon",
        "label": "Hubspot Deal Cumulative Time In \"Closed Won (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_latest_time_in_closedlost",
        "label": "Hubspot Deal Latest Time In \"Closed Lost (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_cumulative_time_in_closedlost",
        "label": "Hubspot Deal Cumulative Time In \"Closed Lost (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_last_invoice_amount",
        "label": "Hubspot Deal Last Invoice Amount",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_predicted_amount_in_home_currency",
        "label": "Hubspot Deal The Predicted Deal Amount In Your Company'S Currency",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_predicted_amount",
        "label": "Hubspot Deal The Predicted Deal Amount",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_cumulative_time_in_77060243",
        "label": "Hubspot Deal Cumulative Time In \"Churned (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_v2_latest_time_in_77060243",
        "label": "Hubspot Deal Latest Time In \"Churned (Sales Pipeline)\"",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_exchange_rate",
        "label": "Hubspot Deal Exchange Rate",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      },
      {
        "value": "$hubspot_deal_hs_pinned_engagement_id",
        "label": "Hubspot Deal Pinned Engagement ID",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_deal"
        }
      }
    ]
  },
  {
    "iconName": "hubspot_ads",
    "label": "Hubspot Companies",
    "values": [
      {
        "value": "$hubspot_company_$object_url",
        "label": "Hubspot Company URL",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_object_source_label",
        "label": "Hubspot Company Record Source",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_object_source",
        "label": "Hubspot Company Record Creation Source",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_name",
        "label": "Hubspot Company Company Name",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$company",
        "label": "Company",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_website",
        "label": "Hubspot Company Website URL",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_domain",
        "label": "Hubspot Company Company Domain Name",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_object_source_id",
        "label": "Hubspot Company Record Creation Source ID",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_is_public",
        "label": "Hubspot Company Is Public",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_description",
        "label": "Hubspot Company Description",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_web_technologies",
        "label": "Hubspot Company Web Technologies",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_linkedinbio",
        "label": "Hubspot Company LinkedIn Bio",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_lifecyclestage",
        "label": "Hubspot Company Lifecycle Stage",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_pipeline",
        "label": "Hubspot Company Pipeline",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_country",
        "label": "Hubspot Company Country/Region",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_timezone",
        "label": "Hubspot Company Time Zone",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_linkedin_company_page",
        "label": "Hubspot Company LinkedIn Company Page",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_city",
        "label": "Hubspot Company City",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_state",
        "label": "Hubspot Company State/Region",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_industry",
        "label": "Hubspot Company Industry",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_founded_year",
        "label": "Hubspot Company Year Founded",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_source",
        "label": "Hubspot Company Original Source Type",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_source_data_1",
        "label": "Hubspot Company Original Source Data 1",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_latest_source_data_1",
        "label": "Hubspot Company Latest Source Data 1",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_latest_source",
        "label": "Hubspot Company Latest Source",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_zip",
        "label": "Hubspot Company Postal Code",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_twitterhandle",
        "label": "Hubspot Company Twitter Handle",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_address",
        "label": "Hubspot Company Street Address",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_latest_source_data_2",
        "label": "Hubspot Company Latest Source Data 2",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_source_data_2",
        "label": "Hubspot Company Original Source Data 2",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_phone",
        "label": "Hubspot Company Phone Number",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_facebook_company_page",
        "label": "Hubspot Company Facebook Company Page",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_recent_conversion_event_name",
        "label": "Hubspot Company Recent Conversion",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_first_conversion_event_name",
        "label": "Hubspot Company First Conversion",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_meeting_status",
        "label": "Hubspot Company Meeting Status",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_last_sales_activity_type",
        "label": "Hubspot Company Last Engagement Type",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_annual_revenue_currency_code",
        "label": "Hubspot Company Annual Revenue Currency Code",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_address2",
        "label": "Hubspot Company Street Address 2",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_object_source_detail_1",
        "label": "Hubspot Company Record Source Detail 1",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_icp",
        "label": "Hubspot Company ICP",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_total_money_raised",
        "label": "Hubspot Company Total Money Raised",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_user_ids_of_all_owners",
        "label": "Hubspot Company User IDs Of All Owners",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_all_owner_ids",
        "label": "Hubspot Company All Owner Ids",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hubspot_owner_id",
        "label": "Hubspot Company Company Owner",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_industry___paragon",
        "label": "Hubspot Company Industry - Paragon",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_naics_description_paragon",
        "label": "Hubspot Company NAICS Description Paragon",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_lead_status",
        "label": "Hubspot Company Lead Status",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_icp_industry_category",
        "label": "Hubspot Company ICP Industry Category",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_revenue_range___factors",
        "label": "Hubspot Company Revenue Range - Paragon",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_employee_range___paragon",
        "label": "Hubspot Company Employee Range - Paragon",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_outbound_company",
        "label": "Hubspot Company Outbound Company",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_was_imported",
        "label": "Hubspot Company Performed In An Import",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_intent",
        "label": "Hubspot Company Intent",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_all_team_ids",
        "label": "Hubspot Company All Team Ids",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hubspot_team_id",
        "label": "Hubspot Company HubSpot Team",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_all_accessible_team_ids",
        "label": "Hubspot Company All Teams",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_about_us",
        "label": "Hubspot Company About Us",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_notes_next_activity",
        "label": "Hubspot Company Hs Notes Next Activity",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_notes_next_activity_type",
        "label": "Hubspot Company Next Activity Type",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_taxonomy",
        "label": "Hubspot Company Taxonomy",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_lfapp_view_in_leadfeeder",
        "label": "Hubspot Company View In Leadfeeder",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_object_source_detail_2",
        "label": "Hubspot Company Record Source Detail 2",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_additional_domains",
        "label": "Hubspot Company Additional Domains",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_object_source_detail_3",
        "label": "Hubspot Company Record Source Detail 3",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_merged_object_ids",
        "label": "Hubspot Company Merged Company IDs",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_last_touch_converting_campaign",
        "label": "Hubspot Company Last Touch Converting Campaign",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_first_touch_converting_campaign",
        "label": "Hubspot Company First Touch Converting Campaign",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_churned",
        "label": "Hubspot Company Churned",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_inbound_outbound",
        "label": "Hubspot Company Inbound/Outbound",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_outbound_source",
        "label": "Hubspot Company Outbound Source",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_identified_by_paragon",
        "label": "Hubspot Company Identified By Paragon",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_is_target_account",
        "label": "Hubspot Company Target Account",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_associated_keytags",
        "label": "Hubspot Company Associated Keytags",
        "extraProps": {
          "valueType": "categorical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_lastmodifieddate",
        "label": "Hubspot Company Last Modified Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_createdate",
        "label": "Hubspot Company Create Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_latest_source_timestamp",
        "label": "Hubspot Company Latest Source Timestamp",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_first_contact_createdate",
        "label": "Hubspot Company First Contact Create Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_first_timestamp",
        "label": "Hubspot Company Time First Seen",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_notes_last_updated",
        "label": "Hubspot Company Last Activity Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_notes_last_contacted",
        "label": "Hubspot Company Last Contacted",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_last_visit_timestamp",
        "label": "Hubspot Company Time Of Last Session",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_recent_conversion_date",
        "label": "Hubspot Company Recent Conversion Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_first_conversion_date",
        "label": "Hubspot Company First Conversion Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_last_sales_activity_date",
        "label": "Hubspot Company Last Sales Activity Date Old",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_first_visit_timestamp",
        "label": "Hubspot Company Time Of First Session",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_last_timestamp",
        "label": "Hubspot Company Time Last Seen",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_last_sales_activity_timestamp",
        "label": "Hubspot Company Last Engagement Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_meeting_status_time_stamp",
        "label": "Hubspot Company Meeting Status Time Stamp",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_notes_next_activity_date",
        "label": "Hubspot Company Next Activity Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_last_booked_meeting_date",
        "label": "Hubspot Company Last Booked Meeting Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_latest_meeting_activity",
        "label": "Hubspot Company Latest Meeting Activity",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hubspot_owner_assigneddate",
        "label": "Hubspot Company Owner Assigned Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_sales_email_last_replied",
        "label": "Hubspot Company Recent Sales Email Replied Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_first_deal_created_date",
        "label": "Hubspot Company First Deal Created Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_closedate",
        "label": "Hubspot Company Close Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_recent_deal_close_date",
        "label": "Hubspot Company Recent Deal Close Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_engagements_last_meeting_booked",
        "label": "Hubspot Company Date Of Last Meeting Booked In Meetings Tool",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_last_logged_call_date",
        "label": "Hubspot Company Last Logged Call Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_lfapp_latest_visit",
        "label": "Hubspot Company Latest Leadfeeder Visit",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_last_open_task_date",
        "label": "Hubspot Company Last Open Task Date",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_pipeline_closedate",
        "label": "Hubspot Company Hs Pipeline Closedate",
        "extraProps": {
          "valueType": "datetime",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_num_associated_contacts",
        "label": "Hubspot Company Number Of Associated Contacts",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_num_open_deals",
        "label": "Hubspot Company Number Of Open Deals",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_num_child_companies",
        "label": "Hubspot Company Number Of Child Companies",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_num_decision_makers",
        "label": "Hubspot Company Number Of Decision Makers",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_num_blockers",
        "label": "Hubspot Company Number Of Blockers",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_object_id",
        "label": "Hubspot Company Record ID",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_num_contacts_with_buying_roles",
        "label": "Hubspot Company Number Of Contacts With A Buying Role",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_target_account_probability",
        "label": "Hubspot Company Target Account Probability",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_numberofemployees",
        "label": "Hubspot Company Number Of Employees",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_annualrevenue",
        "label": "Hubspot Company Annual Revenue",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_num_visits",
        "label": "Hubspot Company Number Of Sessions",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_analytics_num_page_views",
        "label": "Hubspot Company Number Of Pageviews",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_num_conversion_events",
        "label": "Hubspot Company Number Of Form Submissions",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_num_contacted_notes",
        "label": "Hubspot Company Number Of Times Contacted",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_num_notes",
        "label": "Hubspot Company Number Of Sales Activities",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_num_associated_deals",
        "label": "Hubspot Company Number Of Associated Deals",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_total_deal_value",
        "label": "Hubspot Company Total Open Deal Value",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_days_to_close",
        "label": "Hubspot Company Days To Close",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_total_revenue",
        "label": "Hubspot Company Total Revenue",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_recent_deal_amount",
        "label": "Hubspot Company Recent Deal Amount",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_updated_by_user_id",
        "label": "Hubspot Company Updated By User ID",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_arpu",
        "label": "Hubspot Company ARPU",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_number_of_customers",
        "label": "Hubspot Company Number Of Customers",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_object_source_user_id",
        "label": "Hubspot Company Record Creation Source User ID",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_hs_created_by_user_id",
        "label": "Hubspot Company Created By User ID",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_factors_app_project_id",
        "label": "Hubspot Company Factors App Project ID",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      },
      {
        "value": "$hubspot_company_mrr",
        "label": "Hubspot Company MRR",
        "extraProps": {
          "valueType": "numerical",
          "propertyType": "user",
          "groupName": "$hubspot_company"
        }
      }
    ]
  }
]