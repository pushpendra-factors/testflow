//Rule Types For HUBSPOT
export const RULE_TYPE_HS_CONTACT = 'hs_contact';
export const RULE_TYPE_HS_EMAILS = 'hs_emails';
export const RULE_TYPE_HS_FORM_SUBMISSIONS = 'hs_form_submissions';
export const RULE_TYPE_HS_CALLS = 'hs_calls';
export const RULE_TYPE_HS_MEETINGS = 'hs_meetings';
export const RULE_TYPE_HS_LISTS = 'hs_contact_list';

//Rule Types For SALESFORCE
export const RULE_TYPE_SF_CONTACT = 'sf_contact';
export const RULE_TYPE_SF_CAMPAIGNS = 'sf_campaigns';
export const RULE_TYPE_SF_TASKS = 'sf_tasks';
export const RULE_TYPE_SF_EVENTS = 'sf_events';

//Events To Call For HUBSPOT
export const EVENT_HS_EMAILS = '$hubspot_engagement_email';
export const EVENT_HS_CONTACT = '$hubspot_contact_updated';
export const EVENT_HS_FORM_SUBMISSIONS = '$hubspot_form_submission';
export const EVENT_HS_CALL = '$hubspot_engagement_call_updated';
export const EVENT_HS_MEETINGS = '$hubspot_engagement_meeting_updated';
export const EVENT_HS_LISTS = '$hubspot_contact_list';

//Events To Call For SALESFORCE
export const EVENT_SF_CAMPAIGN = [
  '$sf_campaign_member_created',
  '$sf_campaign_member_updated'
];
export const EVENT_SF_EVENT = ['$sf_event_updated'];
export const EVENT_SF_TASK = ['$sf_task_updated'];

//EVENTS-MAP
export const EVENTS_MAP = {
  [RULE_TYPE_HS_CONTACT]: EVENT_HS_CONTACT,
  [RULE_TYPE_HS_EMAILS]: EVENT_HS_EMAILS,
  [RULE_TYPE_HS_FORM_SUBMISSIONS]: EVENT_HS_FORM_SUBMISSIONS,
  [RULE_TYPE_HS_CALLS]: EVENT_HS_CALL,
  [RULE_TYPE_HS_MEETINGS]: EVENT_HS_MEETINGS,
  [RULE_TYPE_HS_LISTS]: EVENT_HS_LISTS,
  [RULE_TYPE_SF_CAMPAIGNS]: EVENT_SF_CAMPAIGN,
  [RULE_TYPE_SF_TASKS]: EVENT_SF_TASK,
  [RULE_TYPE_SF_EVENTS]: EVENT_SF_EVENT
};

export const PROPERTY_MAP_OPTIONS = [
  ['Type', '$type'],
  ['Source', '$source'],
  ['Campaign', '$campaign'],
  ['Channel', '$channel']
];

export const Extra_PROP_SHOW_OPTIONS = [
  ['Campaign Id', null, 'campaign_id'],
  ['Adgroup', null, 'adgroup'],
  ['Adgroup ID', null, 'adgroup_id'],
  ['Page URL', null, 'page_url']
];
export const ruleTypesNameMappingForHS = {
  [RULE_TYPE_HS_CONTACT]: 'Contact',
  [RULE_TYPE_HS_FORM_SUBMISSIONS]: 'Form Submissions',
  [RULE_TYPE_HS_EMAILS]: 'Email',
  [RULE_TYPE_HS_MEETINGS]: 'Meetings',
  [RULE_TYPE_HS_CALLS]: 'Calls',
  [RULE_TYPE_HS_LISTS]: 'Lists'
};
export const ruleTypesNameMappingForSF = {
  [RULE_TYPE_SF_CONTACT]: 'Contact',
  [RULE_TYPE_SF_CAMPAIGNS]: 'Campaigns',
  [RULE_TYPE_SF_TASKS]: 'Tasks',
  [RULE_TYPE_SF_EVENTS]: 'Events'
};
export const DEFAULT_TIMESTAMPS = {
  [RULE_TYPE_HS_CONTACT]: 'LAST_MODIFIED_TIME_REF',
  [RULE_TYPE_HS_EMAILS]: '$hubspot_engagement_timestamp',
  [RULE_TYPE_HS_FORM_SUBMISSIONS]: '$timestamp',
  [RULE_TYPE_HS_MEETINGS]: '$hubspot_engagement_timestamp',
  [RULE_TYPE_HS_CALLS]: '$hubspot_engagement_timestamp',
  [RULE_TYPE_HS_LISTS]: '$hubspot_contact_list_timestamp',
  [RULE_TYPE_SF_CONTACT]: 'campaign_member_created_date',
  [RULE_TYPE_SF_TASKS]: '$salesforce_task_lastmodifieddate',
  [RULE_TYPE_SF_EVENTS]: '$salesforce_event_lastmodifieddate',
  [RULE_TYPE_SF_CAMPAIGNS]: '$sf_campaign_member_created'
};
