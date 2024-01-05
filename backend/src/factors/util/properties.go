package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// Common properties type.
type PropertiesMap map[string]interface{}

// Special Event Names used when building patterns and for querying.
const SEN_ALL_ACTIVE_USERS = "$AllActiveUsers"
const SEN_ALL_ACTIVE_USERS_DISPLAY_STRING = "All Active Users"

const SEN_ALL_EVENTS = "$AllEvents"
const SEN_ALL_EVENTS_DISPLAY_STRING = "All Events"

const EVENT_NAME_SESSION = "$session"
const EVENT_NAME_FORM_FILL = "$form_fill"
const EVENT_NAME_OFFLINE_TOUCH_POINT = "$offline_touch_point"
const EVENT_NAME_FORM_SUBMITTED = "$form_submitted"

// Integration: Hubspot event names.
const EVENT_NAME_HUBSPOT_CONTACT_CREATED = "$hubspot_contact_created"
const EVENT_NAME_HUBSPOT_CONTACT_UPDATED = "$hubspot_contact_updated"
const EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED = "$hubspot_deal_state_changed"
const EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION = "$hubspot_form_submission"
const EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED = "$hubspot_engagement_meeting_created"
const EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED = "$hubspot_engagement_meeting_updated"
const EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED = "$hubspot_engagement_call_created"
const EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED = "$hubspot_engagement_call_updated"
const EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL = "$hubspot_engagement_email"
const EVENT_NAME_HUBSPOT_CONTACT_LIST = "$hubspot_contact_list"

// Integration: Salesforce event names.
const EVENT_NAME_SALESFORCE_CONTACT_CREATED = "$sf_contact_created"
const EVENT_NAME_SALESFORCE_CONTACT_UPDATED = "$sf_contact_updated"
const EVENT_NAME_SALESFORCE_LEAD_CREATED = "$sf_lead_created"
const EVENT_NAME_SALESFORCE_LEAD_UPDATED = "$sf_lead_updated"
const EVENT_NAME_SALESFORCE_ACCOUNT_CREATED = "$sf_account_created"
const EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED = "$sf_account_updated"
const EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED = "$sf_opportunity_created"
const EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED = "$sf_opportunity_updated"
const EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED = "$sf_campaign_member_created"
const EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED = "$sf_campaign_member_updated"
const EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN = "$sf_campaign_member_responded_to_campaign"
const EVENT_NAME_SALESFORCE_TASK_CREATED = "$sf_task_created"
const EVENT_NAME_SALESFORCE_TASK_UPDATED = "$sf_task_updated"
const EVENT_NAME_SALESFORCE_EVENT_CREATED = "$sf_event_created"
const EVENT_NAME_SALESFORCE_EVENT_UPDATED = "$sf_event_updated"

// Integration: Marketo
const EVENT_NAME_MARKETO_LEAD_CREATED = "$marketo_lead_created"
const EVENT_NAME_MARKETO_LEAD_UPDATED = "$marketo_lead_updated"
const EVENT_NAME_MARKETO_PROGRAM_MEMBERSHIP_CREATED = "$marketo_program_membership_created"
const EVENT_NAME_MARKETO_PROGRAM_MEMBERSHIP_UPDATED = "$marketo_program_membership_updated"

// Integration: LeadSquared
const EVENT_NAME_LEAD_SQUARED_LEAD_CREATED = "$leadsquared_lead_created"
const EVENT_NAME_LEAD_SQUARED_LEAD_UPDATED = "$leadsquared_lead_updated"
const EVENT_NAME_LEAD_SQUARED_SALES_ACTIVITY_CREATED = "$leadsquared_sales_activity_created"
const EVENT_NAME_LEAD_SQUARED_SALES_ACTIVITY_UPDATED = "$leadsquared_sales_activity_updated"
const EVENT_NAME_LEAD_SQUARED_HAD_A_CALL_ACTIVITY_CREATED = "$leadsquared_had_a_call_activity_created"
const EVENT_NAME_LEAD_SQUARED_HAD_A_CALL_ACTIVITY_UPDATED = "$leadsquared_had_a_call_activity_updated"
const EVENT_NAME_LEAD_SQUARED_EMAIL_SENT_ACTIVITY_CREATED = "$leadsquared_email_sent_activity_created"
const EVENT_NAME_LEAD_SQUARED_EMAIL_INFO_ACTIVITY_CREATED = "$leadsquared_email_info_activity_created"
const EVENT_NAME_LEADSQUARED_CALLED_A_CUST_NEGATIVE_REPLY_UPDATED = "$leadsquared_called_a_customer_negative_reply_activity_updated"
const EVENT_NAME_LEADSQUARED_CALLED_A_CUST_NEGATIVE_REPLY_CREATED = "$leadsquared_called_a_customer_negative_reply_activity_created"
const EVENT_NAME_LEADSQUARED_CALLED_A_CUST_POSITIVE_REPLY_UPDATED = "$leadsquared_called_a_customer_positive_reply_activity_updated"
const EVENT_NAME_LEADSQUARED_CALLED_A_CUST_POSITIVE_REPLY_CREATED = "$leadsquared_called_a_customer_positive_reply_activity_created"
const EVENT_NAME_LEADSQUARED_CALLED_TO_COLLECT_REFERRAL_UPDATED = "$leadsquared_called_to_collect_referrals_activity_updated"
const EVENT_NAME_LEADSQUARED_CALLED_TO_COLLECT_REFERRAL_CREATED = "$leadsquared_called_to_collect_referrals_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_BOUNCED_UPDATED = "$leadsquared_email_bounced_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_BOUNCED_CREATED = "$leadsquared_email_bounced_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_LINK_CLICKED_UPDATED = "$leadsquared_email_link_clicked_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_LINK_CLICKED_CREATED = "$leadsquared_email_link_clicked_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_MAILING_PREFERENCE_LINK_CLICKED_UPDATED = "$leadsquared_email_mailing_preference_link_clicked_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_MAILING_PREFERENCE_LINK_CLICKED_CREATED = "$leadsquared_email_mailing_preference_link_clicked_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_MARKED_SPAM_UPDATED = "$leadsquared_email_marked_spam_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_MARKED_SPAM_CREATED = "$leadsquared_email_marked_spam_activity_created"
const EVENT_NAME_LEASQUARED_EMAIL_NEGATIVE_RESPONSE_UPDATED = "$leadsquared_email_negative_response_activity_updated"
const EVENT_NAME_LEASQUARED_EMAIL_NEGATIVE_RESPONSE_CREATED = "$leadsquared_email_negative_response_activity_created"
const EVENT_NAME_LEASQUARED_EMAIL_NEUTRAL_RESPONSE_UPDATED = "$leadsquared_email_neutral_response_activity_updated"
const EVENT_NAME_LEASQUARED_EMAIL_NEUTRAL_RESPONSE_CREATED = "$leadsquared_email_neutral_response_activity_created"
const EVENT_NAME_LEASQUARED_EMAIL_POSITIVE_RESPONSE_UPDATED = "$leadsquared_email_positive_response_activity_updated"
const EVENT_NAME_LEASQUARED_EMAIL_POSITIVE_RESPONSE_CREATED = "$leadsquared_email_positive_response_activity_created"
const EVENT_NAME_LEASQUARED_EMAIL_OPENED_UPDATED = "$leadsquared_email_opened_activity_updated"
const EVENT_NAME_LEASQUARED_EMAIL_OPENED_CREATED = "$leadsquared_email_opened_activity_created"
const EVENT_NAME_LEASQUARED_EMAIL_POSITVE_INBOUND_EMAIL_UPDATED = "$leadsquared_email_positive_inbound_email_activity_updated"
const EVENT_NAME_LEASQUARED_EMAIL_POSITVE_INBOUND_EMAIL_CREATED = "$leadsquared_email_positive_inbound_email_activity_created"
const EVENT_NAME_LEASQUARED_EMAIL_RESUBSCRIBED_UPDATED = "$leadsquared_email_resubscribed_activity_updated"
const EVENT_NAME_LEASQUARED_EMAIL_RESUBSCRIBED_CREATED = "$leadsquared_email_resubscribed_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_BOOTCAMP_UPDATED = "$leadsquared_email_subscribed_to_bootcamp_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_BOOTCAMP_CREATED = "$leadsquared_email_subscribed_to_bootcamp_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_COLLECTION_UPDATED = "$leadsquared_email_subscribed_to_collection_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_COLLECTION_CREATED = "$leadsquared_email_subscribed_to_collection_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_EVENTS_UPDATED = "$leadsquared_email_subscribed_to_events_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_EVENTS_CREATED = "$leadsquared_email_subscribed_to_events_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_FESTIVAL_UPDATED = "$leadsquared_email_subscribed_to_festival_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_FESTIVAL_CREATED = "$leadsquared_email_subscribed_to_festival_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION_UPDATED = "$leadsquared_email_subscribed_to_internation_reactivation_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION_CREATED = "$leadsquared_email_subscribed_to_internation_reactivation_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_NEWSLETTER_UPDATED = "$leadsquared_email_subscribed_to_newsletter_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_NEWSLETTER_CREATED = "$leadsquared_email_subscribed_to_newsletter_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_REACTIVATION_UPDATED = "$leadsquared_email_subscribed_to_reactivation_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_REACTIVATION_CREATED = "$leadsquared_email_subscribed_to_reactivation_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_REFERRAL_UPDATED = "$leadsquared_email_subscribed_to_referrral_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_REFERRAL_CREATED = "$leadsquared_email_subscribed_to_referrral_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_SURVEY_UPDATED = "$leadsquared_email_subscribed_to_survey_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_SURVEY_CREATED = "$leadsquared_email_subscribed_to_survey_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_TEST_UPDATED = "$leadsquared_email_subscribed_to_test_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_TEST_CREATED = "$leadsquared_email_subscribed_to_test_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_WORKSHOP_UPDATED = "$leadsquared_email_subscribed_to_workshop_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_WORKSHOP_CREATED = "$leadsquared_email_subscribed_to_workshop_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_BOOTCAMP_UPDATED = "$leadsquared_email_unsubscribed_to_bootcamp_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_BOOTCAMP_CREATED = "$leadsquared_email_unsubscribed_to_bootcamp_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_COLLECTION_UPDATED = "$leadsquared_email_unsubscribed_to_collection_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_COLLECTION_CREATED = "$leadsquared_email_unsubscribed_to_collection_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_EVENTS_UPDATED = "$leadsquared_email_unsubscribed_to_events_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_EVENTS_CREATED = "$leadsquared_email_unsubscribed_to_events_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_FESTIVAL_UPDATED = "$leadsquared_email_unsubscribed_to_festival_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_FESTIVAL_CREATED = "$leadsquared_email_unsubscribed_to_festival_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION_UPDATED = "$leadsquared_email_unsubscribed_to_internation_reactivation_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION_CREATED = "$leadsquared_email_unsubscribed_to_internation_reactivation_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_NEWSLETTER_UPDATED = "$leadsquared_email_unsubscribed_to_newsletter_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_NEWSLETTER_CREATED = "$leadsquared_email_unsubscribed_to_newsletter_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REACTIVATION_UPDATED = "$leadsquared_email_unsubscribed_to_reactivation_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REACTIVATION_CREATED = "$leadsquared_email_unsubscribed_to_reactivation_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REFERRAL_UPDATED = "$leadsquared_email_unsubscribed_to_referrral_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REFERRAL_CREATED = "$leadsquared_email_unsubscribed_to_referrral_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_SURVEY_UPDATED = "$leadsquared_email_unsubscribed_to_survey_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_SURVEY_CREATED = "$leadsquared_email_unsubscribed_to_survey_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_TEST_UPDATED = "$leadsquared_email_unsubscribed_to_test_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_TEST_CREATED = "$leadsquared_email_unsubscribed_to_test_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_WORKSHOP_UPDATED = "$leadsquared_email_unsubscribed_to_workshop_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_WORKSHOP_CREATED = "$leadsquared_email_unsubscribed_to_workshop_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBE_LINK_CLICKED_UPDATED = "$leadsquared_email_unsubscribe_link_clicked_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBE_LINK_CLICKED_CREATED = "$leadsquared_email_unsubscribe_link_clicked_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_UPDATED = "$leadsquared_email_unsubscribed_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_CREATED = "$leadsquared_email_unsubscribed_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_VIEW_IN_BROWSER_LINK_CLICKED_UPDATED = "$leadsquared_email_view_in_browser_link_clicked_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_VIEW_IN_BROWSER_LINK_CLICKED_CREATED = "$leadsquared_email_view_in_browser_link_clicked_activity_created"
const EVENT_NAME_LEADSQUARED_EMAIL_RECEIVED_UPDATED = "$leadsquared_email_received_activity_updated"
const EVENT_NAME_LEADSQUARED_EMAIL_RECEIVED_CREATED = "$leadsquared_email_received_activity_created"

const GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED = "$hubspot_company_created"
const GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED = "$hubspot_company_updated"
const GROUP_EVENT_NAME_HUBSPOT_DEAL_CREATED = "$hubspot_deal_created"
const GROUP_EVENT_NAME_HUBSPOT_DEAL_UPDATED = "$hubspot_deal_updated"
const GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED = "$salesforce_account_created"
const GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED = "$salesforce_account_updated"
const GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED = "$salesforce_opportunity_created"
const GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED = "$salesforce_opportunity_updated"

const GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD = "$linkedin_viewed_ad"
const GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD = "$linkedin_clicked_ad"

const GROUP_EVENT_NAME_G2_ALL = "$g2_all"
const GROUP_EVENT_NAME_G2_SPONSORED = "$g2_sponsored"
const GROUP_EVENT_NAME_G2_PRODUCT_PROFILE = "$g2_product_profile"
const GROUP_EVENT_NAME_G2_ALTERNATIVE = "$g2_alternative"
const GROUP_EVENT_NAME_G2_PRICING = "$g2_pricing"
const GROUP_EVENT_NAME_G2_CATEGORY = "$g2_category"
const GROUP_EVENT_NAME_G2_COMPARISON = "$g2_comparison"
const GROUP_EVENT_NAME_G2_REPORT = "$g2_report"
const GROUP_EVENT_NAME_G2_REFERENCE = "$g2_reference"
const GROUP_EVENT_NAME_G2_DEAL = "$g2_deal"

const GROUP_EVENT_NAME_ENGAGEMENT_SCORE = "$engagement_score"
const GROUP_EVENT_NAME_TOTAL_ENGAGEMENT_SCORE = "$total_enagagement_score"

// Integration shopify event names.
const EVENT_NAME_SHOPIFY_CHECKOUT_CREATED = "$shopify_checkout_created"
const EVENT_NAME_SHOPIFY_CHECKOUT_UPDATED = "$shopify_checkout_updated"
const EVENT_NAME_SHOPIFY_ORDER_CREATED = "$shopify_order_created"
const EVENT_NAME_SHOPIFY_ORDER_UPDATED = "$shopify_order_updated"
const EVENT_NAME_SHOPIFY_ORDER_PAID = "$shopify_order_paid"
const EVENT_NAME_SHOPIFY_ORDER_CANCELLED = "$shopify_order_cancelled"
const EVENT_NAME_SHOPIFY_CART_UPDATED = "$shopify_cart_updated"

var ALLOWED_INTERNAL_EVENT_NAMES = [...]string{
	EVENT_NAME_SESSION,
	EVENT_NAME_FORM_FILL,
	EVENT_NAME_FORM_SUBMITTED,
	EVENT_NAME_HUBSPOT_CONTACT_CREATED,
	EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
	EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED,
	EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED,
	EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED,
	EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED,
	EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
	EVENT_NAME_HUBSPOT_CONTACT_LIST,
	EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED,
	EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION,
	GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
	GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED,
	GROUP_EVENT_NAME_HUBSPOT_DEAL_CREATED,
	GROUP_EVENT_NAME_HUBSPOT_DEAL_UPDATED,
	EVENT_NAME_SHOPIFY_CHECKOUT_CREATED,
	EVENT_NAME_SHOPIFY_CHECKOUT_UPDATED,
	EVENT_NAME_SHOPIFY_ORDER_CREATED,
	EVENT_NAME_SHOPIFY_ORDER_UPDATED,
	EVENT_NAME_SHOPIFY_ORDER_PAID,
	EVENT_NAME_SHOPIFY_ORDER_CANCELLED,
	EVENT_NAME_SHOPIFY_CART_UPDATED,
	EVENT_NAME_SALESFORCE_CONTACT_CREATED,
	EVENT_NAME_SALESFORCE_CONTACT_UPDATED,
	EVENT_NAME_SALESFORCE_LEAD_CREATED,
	EVENT_NAME_SALESFORCE_LEAD_UPDATED,
	EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
	EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
	EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED,
	EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED,
	EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
	EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
	EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN,
	EVENT_NAME_SALESFORCE_TASK_CREATED,
	EVENT_NAME_SALESFORCE_TASK_UPDATED,
	EVENT_NAME_SALESFORCE_EVENT_CREATED,
	EVENT_NAME_SALESFORCE_EVENT_UPDATED,
	GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
	GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
	GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED,
	GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED,
	EVENT_NAME_MARKETO_LEAD_CREATED,
	EVENT_NAME_MARKETO_LEAD_UPDATED,
	EVENT_NAME_MARKETO_PROGRAM_MEMBERSHIP_CREATED,
	EVENT_NAME_MARKETO_PROGRAM_MEMBERSHIP_UPDATED,
	EVENT_NAME_LEAD_SQUARED_LEAD_CREATED,
	EVENT_NAME_LEAD_SQUARED_LEAD_UPDATED,
	EVENT_NAME_LEAD_SQUARED_SALES_ACTIVITY_CREATED,
	EVENT_NAME_LEAD_SQUARED_SALES_ACTIVITY_UPDATED,
	EVENT_NAME_LEAD_SQUARED_HAD_A_CALL_ACTIVITY_CREATED,
	EVENT_NAME_LEAD_SQUARED_HAD_A_CALL_ACTIVITY_UPDATED,
	EVENT_NAME_LEAD_SQUARED_EMAIL_SENT_ACTIVITY_CREATED,
	EVENT_NAME_LEAD_SQUARED_EMAIL_INFO_ACTIVITY_CREATED,
	EVENT_NAME_LEADSQUARED_CALLED_A_CUST_NEGATIVE_REPLY_UPDATED,
	EVENT_NAME_LEADSQUARED_CALLED_A_CUST_NEGATIVE_REPLY_CREATED,
	EVENT_NAME_LEADSQUARED_CALLED_A_CUST_POSITIVE_REPLY_UPDATED,
	EVENT_NAME_LEADSQUARED_CALLED_A_CUST_POSITIVE_REPLY_CREATED,
	EVENT_NAME_LEADSQUARED_CALLED_TO_COLLECT_REFERRAL_UPDATED,
	EVENT_NAME_LEADSQUARED_CALLED_TO_COLLECT_REFERRAL_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_BOUNCED_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_BOUNCED_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_LINK_CLICKED_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_LINK_CLICKED_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_MAILING_PREFERENCE_LINK_CLICKED_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_MAILING_PREFERENCE_LINK_CLICKED_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_MARKED_SPAM_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_MARKED_SPAM_CREATED,
	EVENT_NAME_LEASQUARED_EMAIL_NEGATIVE_RESPONSE_UPDATED,
	EVENT_NAME_LEASQUARED_EMAIL_NEGATIVE_RESPONSE_CREATED,
	EVENT_NAME_LEASQUARED_EMAIL_NEUTRAL_RESPONSE_UPDATED,
	EVENT_NAME_LEASQUARED_EMAIL_NEUTRAL_RESPONSE_CREATED,
	EVENT_NAME_LEASQUARED_EMAIL_POSITIVE_RESPONSE_UPDATED,
	EVENT_NAME_LEASQUARED_EMAIL_POSITIVE_RESPONSE_CREATED,
	EVENT_NAME_LEASQUARED_EMAIL_OPENED_UPDATED,
	EVENT_NAME_LEASQUARED_EMAIL_OPENED_CREATED,
	EVENT_NAME_LEASQUARED_EMAIL_POSITVE_INBOUND_EMAIL_UPDATED,
	EVENT_NAME_LEASQUARED_EMAIL_POSITVE_INBOUND_EMAIL_CREATED,
	EVENT_NAME_LEASQUARED_EMAIL_RESUBSCRIBED_UPDATED,
	EVENT_NAME_LEASQUARED_EMAIL_RESUBSCRIBED_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_BOOTCAMP_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_BOOTCAMP_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_COLLECTION_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_COLLECTION_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_EVENTS_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_EVENTS_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_FESTIVAL_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_FESTIVAL_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_NEWSLETTER_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_NEWSLETTER_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_REACTIVATION_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_REACTIVATION_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_REFERRAL_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_REFERRAL_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_SURVEY_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_SURVEY_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_TEST_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_TEST_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_WORKSHOP_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_SUBSCRIBED_TO_WORKSHOP_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_BOOTCAMP_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_BOOTCAMP_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_COLLECTION_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_COLLECTION_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_EVENTS_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_EVENTS_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_FESTIVAL_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_FESTIVAL_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_NEWSLETTER_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_NEWSLETTER_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REACTIVATION_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REACTIVATION_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REFERRAL_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REFERRAL_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_SURVEY_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_SURVEY_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_TEST_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_TEST_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_WORKSHOP_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_WORKSHOP_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBE_LINK_CLICKED_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBE_LINK_CLICKED_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_UNSUBSCRIBED_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_VIEW_IN_BROWSER_LINK_CLICKED_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_VIEW_IN_BROWSER_LINK_CLICKED_CREATED,
	EVENT_NAME_LEADSQUARED_EMAIL_RECEIVED_UPDATED,
	EVENT_NAME_LEADSQUARED_EMAIL_RECEIVED_CREATED,
	GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD,
	GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD,
	GROUP_EVENT_NAME_G2_ALL,
	GROUP_EVENT_NAME_G2_SPONSORED,
	GROUP_EVENT_NAME_G2_PRODUCT_PROFILE,
	GROUP_EVENT_NAME_G2_ALTERNATIVE,
	GROUP_EVENT_NAME_G2_PRICING,
	GROUP_EVENT_NAME_G2_CATEGORY,
	GROUP_EVENT_NAME_G2_COMPARISON,
	GROUP_EVENT_NAME_G2_REPORT,
	GROUP_EVENT_NAME_G2_REFERENCE,
	GROUP_EVENT_NAME_G2_DEAL,
}

const GROUP_NAME_HUBSPOT_COMPANY = "$hubspot_company"
const GROUP_NAME_HUBSPOT_DEAL = "$hubspot_deal"
const GROUP_NAME_SALESFORCE_ACCOUNT = "$salesforce_account"
const GROUP_NAME_SALESFORCE_OPPORTUNITY = "$salesforce_opportunity"
const GROUP_NAME_SIX_SIGNAL = "$6signal"
const GROUP_NAME_DOMAINS = "$domains"
const GROUP_NAME_LINKEDIN_COMPANY = "$linkedin_company"
const GROUP_NAME_G2 = "$g2"

var GROUP_EVENT_NAME_TO_GROUP_NAME_MAPPING = map[string]string{
	GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED:        GROUP_NAME_HUBSPOT_COMPANY,
	GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED:        GROUP_NAME_HUBSPOT_COMPANY,
	GROUP_EVENT_NAME_HUBSPOT_DEAL_CREATED:           GROUP_NAME_HUBSPOT_DEAL,
	GROUP_EVENT_NAME_HUBSPOT_DEAL_UPDATED:           GROUP_NAME_HUBSPOT_DEAL,
	GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED:     GROUP_NAME_SALESFORCE_ACCOUNT,
	GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED:     GROUP_NAME_SALESFORCE_ACCOUNT,
	GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED: GROUP_NAME_SALESFORCE_OPPORTUNITY,
	GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED: GROUP_NAME_SALESFORCE_OPPORTUNITY,
	GROUP_EVENT_NAME_G2_ALL:                         GROUP_NAME_G2,
	GROUP_EVENT_NAME_G2_SPONSORED:                   GROUP_NAME_G2,
	GROUP_EVENT_NAME_G2_PRODUCT_PROFILE:             GROUP_NAME_G2,
	GROUP_EVENT_NAME_G2_ALTERNATIVE:                 GROUP_NAME_G2,
	GROUP_EVENT_NAME_G2_PRICING:                     GROUP_NAME_G2,
	GROUP_EVENT_NAME_G2_CATEGORY:                    GROUP_NAME_G2,
	GROUP_EVENT_NAME_G2_COMPARISON:                  GROUP_NAME_G2,
	GROUP_EVENT_NAME_G2_REPORT:                      GROUP_NAME_G2,
	GROUP_EVENT_NAME_G2_REFERENCE:                   GROUP_NAME_G2,
	GROUP_EVENT_NAME_G2_DEAL:                        GROUP_NAME_G2,
	GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD:             GROUP_NAME_LINKEDIN_COMPANY,
	GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD:            GROUP_NAME_LINKEDIN_COMPANY,
}

// Group/CRM Properties Constants
var GP_HUBSPOT_COMPANY_NAME string = "$hubspot_company_name"
var GP_SALESFORCE_ACCOUNT_NAME string = "$salesforce_account_name"
var GP_SALESFORCE_ACCOUNT_TYPE string = "$salesforce_account_type"
var GP_SALESFORCE_OPPORTUNITY_NAME string = "$salesforce_opportunity_name"
var GP_SALESFORCE_OPPORTUNITY_STAGENAME string = "$salesforce_opportunity_stagename"
var GP_SALESFORCE_OPPORTUNITY_TYPE string = "$salesforce_opportunity_type"
var GP_HUBSPOT_COMPANY_COUNTRY string = "$hubspot_company_country"
var GP_SALESFORCE_ACCOUNT_BILLINGCOUNTRY string = "$salesforce_account_billingcountry"
var GP_HUBSPOT_COMPANY_INDUSTRY string = "$hubspot_company_industry"
var GP_SALESFORCE_ACCOUNT_INDUSTRY string = "$salesforce_account_industry"
var GP_HUBSPOT_COMPANY_NUMBEROFEMPLOYEES string = "$hubspot_company_numberofemployees"
var GP_SALESFORCE_ACCOUNT_NUMBEROFEMPLOYEES string = "$salesforce_account_numberOfEmployees"
var GP_HUBSPOT_COMPANY_DOMAIN string = "$hubspot_company_domain"
var GP_SALESFORCE_ACCOUNT_WEBSITE string = "$salesforce_account_website"
var GP_HUBSPOT_COMPANY_NUM_ASSOCIATED_CONTACTS string = "$hubspot_company_num_associated_contacts"

// Factors API constants
const UserCreated string = "UC"
const AutoTracked string = "AT"

/* Properties Constants */

// Generic Event Properties.
var EP_FIRST_SEEN_OCCURRENCE_COUNT string = "$firstSeenOccurrenceCount"
var EP_LAST_SEEN_OCCURRENCE_COUNT string = "$lastSeenOccurrenceCount"
var EP_FIRST_SEEN_TIME string = "$firstSeenTime"
var EP_LAST_SEEN_TIME string = "$lastSeenTime"
var EP_FIRST_SEEN_SINCE_USER_JOIN string = "$firstSeenSinceUserJoin"
var EP_LAST_SEEN_SINCE_USER_JOIN string = "$lastSeenSinceUserJoin"
var EP_CRM_REFERENCE_EVENT_ID string = "$crm_reference_event_id"

// lastmodifieddate properties.
const PROPERTY_KEY_LAST_MODIFIED_DATE = "lastmodifieddate"
const PROPERTY_KEY_LAST_MODIFIED_DATE_HS = "hs_lastmodifieddate"

// Properties used only for $form_fill events.
const EP_TIME_SPENT_ON_FORM = "time_spent_on_form"             // unit:seconds
const EP_TIME_SPENT_ON_FORM_FIELD = "time_spent_on_form_field" // unit:seconds
const EP_FORM_FIELD_VALUE = "form_field_value"

var GENERIC_NUMERIC_EVENT_PROPERTIES = [...]string{
	EP_FIRST_SEEN_OCCURRENCE_COUNT,
	EP_LAST_SEEN_OCCURRENCE_COUNT,
	EP_FIRST_SEEN_TIME,
	EP_LAST_SEEN_TIME,
	EP_FIRST_SEEN_SINCE_USER_JOIN,
	EP_LAST_SEEN_SINCE_USER_JOIN,
}

// Generic User Properties.
var UP_JOIN_TIME string = "$joinTime"

var GENERIC_NUMERIC_USER_PROPERTIES = [...]string{
	UP_JOIN_TIME,
}

var PROPERTIES_TYPE_DATE_TIME = [...]string{
	UP_JOIN_TIME,
}

// Generic hubspot properties
const CRM_HUBSPOT_DEALSTAGE = "$hubspot_deal_dealstage"

// status for sync job
const (
	CRM_SYNC_STATUS_SUCCESS  = "success"
	CRM_SYNC_STATUS_FAILURES = "failures_seen"
)

// Event Properites
var EP_INTERNAL_IP string = "$ip"
var EP_SKIP_SESSION string = "$skip_session"

var EP_LOCATION_LATITUDE string = "$location_lat"
var EP_LOCATION_LONGITUDE string = "$location_lng"
var EP_IS_PAGE_VIEW string = "$is_page_view" // type:bool
var EP_PAGE_TITLE string = "$page_title"
var EP_PAGE_DOMAIN string = "$page_domain"
var EP_PAGE_RAW_URL string = "$page_raw_url"
var EP_PAGE_URL string = "$page_url"
var EP_REFERRER string = "$referrer"
var EP_REFERRER_DOMAIN string = "$referrer_domain"
var EP_REFERRER_URL string = "$referrer_url"
var EP_PAGE_LOAD_TIME string = "$page_load_time"   // unit:seconds
var EP_PAGE_SPENT_TIME string = "$page_spent_time" // unit:seconds
var EP_PAGE_SCROLL_PERCENT string = "$page_scroll_percent"
var EP_SEGMENT_EVENT_VERSION string = "$segment_event_version"
var EP_SEGMENT_SOURCE_LIBRARY string = "$segment_source_library" // values: analytics.js, analytics-python, analytics-react, etc.,
var EP_SEGMENT_SOURCE_CHANNEL string = "$segment_source_channel" // values: client, server
var EP_RUDDERSTACK_EVENT_VERSION string = "$rudderstack_event_version"
var EP_RUDDERSTACK_SOURCE_LIBRARY string = "$rudderstack_source_library" // values: analytics.js, analytics-python, analytics-react, etc.,
var EP_RUDDERSTACK_SOURCE_CHANNEL string = "$rudderstack_source_channel" // values: client, server
var EP_CAMPAIGN string = "$campaign"
var EP_CAMPAIGN_ID string = "$campaign_id"
var EP_SOURCE string = "$source"
var EP_MEDIUM string = "$medium"
var EP_KEYWORD string = "$keyword"
var EP_KEYWORD_MATCH_TYPE string = "$keyword_match_type"
var EP_TERM string = "$term"
var EP_CONTENT string = "$content"
var EP_ADGROUP string = "$adgroup"
var EP_ADGROUP_ID string = "$adgroup_id"
var EP_AD string = "$ad"
var EP_AD_ID string = "$ad_id"
var EP_CREATIVE string = "$creative"
var EP_GCLID string = "$gclid"
var EP_FBCLID string = "$fbclid"
var EP_COST string = "$cost"
var EP_REVENUE string = "$revenue"
var EP_PAGE_COUNT string = "$page_count"
var EP_TIMESTAMP string = "$timestamp"
var EP_HOUR_OF_DAY string = "$hour_of_day"
var EP_DAY_OF_WEEK string = "$day_of_week"
var EP_SESSION_COUNT string = "$session_count"
var EP_CHANNEL string = "$channel" // added at runtime.
var EP_TYPE string = "$type"
var EP_HUBSPOT_ENGAGEMENT_THREAD_ID string = "$hubspot_engagement_thread_id"
var EP_OTP_RULE_ID string = "$otp_rule_id"
var EP_SALESFORCE_CAMPAIGN_NAME string = "$salesforce_campaign_name"
var EP_HUBSPOT_FORM_SUBMISSION_TITLE string = "$hubspot_form_submission_title"
var EP_HUBSPOT_FORM_SUBMISSION_FORMTYPE string = "$hubspot_form_submission_form-type"
var EP_HUBSPOT_FORM_SUBMISSION_PAGETITLE string = "$hubspot_form_submission_page-title"
var EP_HUBSPOT_FORM_SUBMISSION_PAGEURL string = "$hubspot_form_submission_page-url-no-qp"
var EP_HUBSPOT_FORM_SUBMISSION_TIMESTAMP string = "$hubspot_form_submission_timestamp"
var EP_HUBSPOT_ENGAGEMENT_TITLE string = "$hubspot_engagement_title"
var EP_SALESFORCE_CAMPAIGN_TYPE string = "$salesforce_campaign_type"
var EP_SALESFORCE_CAMPAIGNMEMBER_STATUS string = "$salesforce_campaignmember_status"
var EP_HUBSPOT_ENGAGEMENT_TYPE string = "$hubspot_engagement_type"
var EP_HUBSPOT_ENGAGEMENT_SUBJECT string = "$hubspot_engagement_subject"
var EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME string = "$hubspot_engagement_meetingoutcome"
var EP_HUBSPOT_ENGAGEMENT_STARTTIME string = "$hubspot_engagement_starttime"
var EP_HUBSPOT_ENGAGEMENT_ENDTIME string = "$hubspot_engagement_endtime"
var EP_HUBSPOT_ENGAGEMENT_DURATIONMILLISECONDS string = "$hubspot_engagement_durationmilliseconds"
var EP_HUBSPOT_ENGAGEMENT_STATUS string = "$hubspot_engagement_status"
var EP_HUBSPOT_ENGAGEMENT_SOURCE string = "$hubspot_engagement_source"
var EP_HUBSPOT_ENGAGEMENT_FROM string = "$hubspot_engagement_from"
var EP_HUBSPOT_ENGAGEMENT_TO string = "$hubspot_engagement_to"
var EP_HUBSPOT_ENGAGEMENT_TIMESTAMP string = "$hubspot_engagement_timestamp"
var EP_HUBSPOT_ENGAGEMENT_ID string = "$hubspot_engagement_id"
var EP_HUBSPOT_CONTACT_EMAIL string = "$hubspot_contact_email"
var EP_HUBSPOT_CONTACT_FIRSTNAME string = "$hubspot_contact_firstname"
var EP_HUBSPOT_CONTACT_LASTNAME string = "$hubspot_contact_lastname"
var EP_SALESFORCE_CONTACT_NAME string = "$salesforce_contact_name"
var EP_SALESFORCE_CONTACT_EMAIL string = "$salesforce_contact_email"
var EP_HUBSPOT_CONTACT_LIST_LIST_ID string = "$hubspot_contact_list_list_id"
var EP_HUBSPOT_CONTACT_LIST_LIST_NAME string = "$hubspot_contact_list_list_name"
var EP_HUBSPOT_CONTACT_LIST_LIST_TYPE string = "$hubspot_contact_list_list_type"
var EP_HUBSPOT_CONTACT_LIST_LIST_CREATED_TIMESTAMP string = "$hubspot_contact_create_timestamp"
var EP_HUBSPOT_CONTACT_LIST_CONTACT_EMAIL string = "$hubspot_contact_list_contact_email"
var EP_OTP_UNIQUE_KEY string = "$otp_unique_key"
var EP_SF_LEAD_NAME string = "$salesforce_lead_name"
var EP_SF_LEAD_EMAIL string = "$salesforce_lead_email"
var EP_SF_TASK_ID string = "$salesforce_task_id"
var EP_SF_EVENT_ID string = "$salesforce_event_id"
var EP_SF_TASK_SUBJECT string = "$salesforce_task_subject"
var EP_SF_TASK_TYPE string = "$salesforce_task_type"
var EP_SF_TASK_SUBTYPE string = "$salesforce_task_tasksubtype"
var EP_SF_TASK_STATUS string = "$salesforce_task_status"
var EP_SF_TASK_DESCRIPTION string = "$salesforce_task_description"
var EP_SF_TASK_COMPLETED_DATETIME string = "$salesforce_task_completeddatetime"
var EP_SF_EVENT_SUBJECT string = "$salesforce_event_subject"
var EP_SF_EVENT_TYPE string = "$salesforce_event_type"
var EP_SF_EVENT_SUBTYPE string = "$salesforce_event_eventsubtype"
var EP_SF_EVENT_COMPLETED_DATETIME string = "$salesforce_event_completeddatetime"
var EP_G2_TAG string = "$g2_tag"
var EP_G2_CATEGORY_IDS string = "$g2_category_ids"
var EP_G2_PRODUCT_IDS string = "$g2_product_ids"
var EP_G2_CITY string = "$g2_visitor_city"
var EP_G2_STATE string = "$g2_visitor_state"
var EP_G2_COUNTRY string = "$g2_visitor_country"

// Event Form meta attributes properties
var EP_FORM_ID string = "$form_id"
var EP_FORM_NAME string = "$form_name"
var EP_FORM_CLASS string = "$form_class"
var EP_FORM_TYPE string = "$form_type"
var EP_FORM_METHOD string = "$form_method"
var EP_FORM_TARGET string = "$form_target"
var EP_FORM_ACTION string = "$form_action"

// User Properties
var UP_INITIAL_PAGE_EVENT_ID string = "$initial_page_event_id" // internal. id of initial page event.
var UP_MERGE_TIMESTAMP string = "$merge_timestamp"             // Internal property used in user properties merge.

var UP_PLATFORM string = "$platform"
var UP_BROWSER string = "$browser"
var UP_BROWSER_VERSION string = "$browser_version"
var UP_BROWSER_WITH_VERSION string = "$browser_with_version"
var UP_USER_AGENT string = "$user_agent"
var UP_OS string = "$os"
var UP_OS_VERSION string = "$os_version"
var UP_OS_WITH_VERSION string = "$os_with_version"
var UP_SCREEN_WIDTH string = "$screen_width"
var UP_SCREEN_HEIGHT string = "$screen_height"
var UP_SCREEN_DENSITY string = "$screen_density"
var UP_LANGUAGE string = "$language"
var UP_LOCALE string = "$locale"
var UP_DEVICE_ID string = "$device_id"
var UP_DEVICE_NAME string = "$device_name"
var UP_DEVICE_ADVERTISING_ID string = "$device_advertising_id"
var UP_DEVICE_BRAND string = "$device_brand"
var UP_DEVICE_MODEL string = "$device_model"
var UP_DEVICE_TYPE string = "$device_type"
var UP_DEVICE_FAMILY string = "$device_family"
var UP_DEVICE_MANUFACTURER string = "$device_manufacturer"
var UP_DEVICE_CARRIER string = "$device_carrier"
var UP_DEVICE_ADTRACKING_ENABLED string = "$device_ad_tracking_enabled"
var UP_NETWORK_BLUETOOTH string = "$network_bluetooth"
var UP_NETWORK_CARRIER string = "$network_carrier"
var UP_NETWORK_CELLULAR string = "$network_cellular"
var UP_NETWORK_WIFI string = "$network_wifi"
var UP_APP_NAME string = "$app_name"
var UP_APP_NAMESPACE string = "$app_namespace"
var UP_APP_VERSION string = "$app_version"
var UP_APP_BUILD string = "$app_build"
var UP_COUNTRY string = "$country"
var UP_ISO_CODE string = "$iso_code"
var UP_CITY string = "$city"
var UP_CONTINENT string = "$continent"
var UP_POSTAL_CODE string = "$postal_code"
var UP_REGION string = "$region"
var UP_TIMEZONE string = "$timezone"
var UP_USER_ID string = "$user_id"
var UP_EMAIL string = "$email"
var UP_COMPANY string = "$company"
var UP_NAME string = "$name"
var UP_FIRST_NAME string = "$first_name"
var UP_LAST_NAME string = "$last_name"
var UP_PHONE string = "$phone"
var UP_INITIAL_PAGE_URL string = "$initial_page_url"
var UP_INITIAL_PAGE_DOMAIN string = "$initial_page_domain"
var UP_INITIAL_PAGE_RAW_URL string = "$initial_page_raw_url"
var UP_INITIAL_PAGE_LOAD_TIME string = "$initial_page_load_time"
var UP_INITIAL_PAGE_SPENT_TIME string = "$initial_page_spent_time" // unit:seconds
var UP_INITIAL_PAGE_SCROLL_PERCENT string = "$initial_page_scroll_percent"
var UP_INITIAL_CAMPAIGN string = "$initial_campaign"
var UP_INITIAL_CAMPAIGN_ID string = "$initial_campaign_id"
var UP_INITIAL_SOURCE string = "$initial_source"
var UP_INITIAL_MEDIUM string = "$initial_medium"
var UP_INITIAL_KEYWORD string = "$initial_keyword"
var UP_INITIAL_KEYWORD_MATCH_TYPE string = "$initial_keyword_match_type"
var UP_INITIAL_TERM string = "$initial_term"
var UP_INITIAL_CONTENT string = "$initial_content"
var UP_INITIAL_ADGROUP string = "$initial_adgroup"
var UP_INITIAL_ADGROUP_ID string = "$initial_adgroup_id"
var UP_INITIAL_CREATIVE string = "$initial_creative"
var UP_INITIAL_GCLID string = "$initial_gclid"
var UP_INITIAL_FBCLID string = "$initial_fbclid"
var UP_INITIAL_REFERRER string = "$initial_referrer"
var UP_INITIAL_REFERRER_URL string = "$initial_referrer_url"
var UP_INITIAL_REFERRER_DOMAIN string = "$initial_referrer_domain"
var UP_DAY_OF_FIRST_EVENT string = "$day_of_first_event"
var UP_HOUR_OF_FIRST_EVENT string = "$hour_of_first_event"

// ** INITIAL_CHANNEL is the channel of First session for the user
var UP_INITIAL_CHANNEL string = "$initial_channel"

var UP_SESSION_COUNT string = "$session_count"
var UP_PAGE_COUNT string = "$page_count"
var UP_TOTAL_SPENT_TIME string = "$session_spent_time" // unit:seconds
var UP_META_OBJECT_IDENTIFIER_KEY = "$identifiers"

var UP_LATEST_PAGE_URL string = "$latest_page_url"
var UP_LATEST_PAGE_DOMAIN string = "$latest_page_domain"
var UP_LATEST_PAGE_RAW_URL string = "$latest_page_raw_url"
var UP_LATEST_PAGE_LOAD_TIME string = "$latest_page_load_time"
var UP_LATEST_PAGE_SPENT_TIME string = "$latest_page_spent_time" // unit:seconds
var UP_LATEST_PAGE_SCROLL_PERCENT string = "$latest_page_scroll_percent"
var UP_LATEST_CAMPAIGN string = "$latest_campaign"
var UP_LATEST_CAMPAIGN_ID string = "$latest_campaign_id"
var UP_LATEST_SOURCE string = "$latest_source"
var UP_LATEST_MEDIUM string = "$latest_medium"
var UP_LATEST_KEYWORD string = "$latest_keyword"
var UP_LATEST_KEYWORD_MATCH_TYPE string = "$latest_keyword_match_type"
var UP_LATEST_TERM string = "$latest_term"
var UP_LATEST_CONTENT string = "$latest_content"
var UP_LATEST_ADGROUP string = "$latest_adgroup"
var UP_LATEST_ADGROUP_ID string = "$latest_adgroup_id"
var UP_LATEST_CREATIVE string = "$latest_creative"
var UP_LATEST_GCLID string = "$latest_gclid"
var UP_LATEST_FBCLID string = "$latest_fbclid"
var UP_LATEST_REFERRER string = "$latest_referrer"
var UP_LATEST_REFERRER_URL string = "$latest_referrer_url"
var UP_LATEST_REFERRER_DOMAIN string = "$latest_referrer_domain"

// ** LATEST_CHANNEL is the channel of last session for the user, incase of sessionUserProperties it's the channel of that session
var UP_LATEST_CHANNEL string = "$latest_channel"

// session properties
var SP_IS_FIRST_SESSION = "$is_first_session" // type:bool
var SP_PAGE_VIEWS = "$page_views"
var SP_SESSION_TIME = "$session_time"
var SP_INITIAL_REFERRER = "$initial_referrer"
var SP_INITIAL_REFERRER_URL = "$initial_referrer_url"
var SP_INITIAL_REFERRER_DOMAIN = "$initial_referrer_domain"
var SP_SPENT_TIME string = "$session_spent_time" // unit:seconds
var SP_PAGE_COUNT string = "$page_count"
var SP_LATEST_PAGE_URL = "$session_latest_page_url"
var SP_LATEST_PAGE_RAW_URL = "$session_latest_page_raw_url"

// session properties same as user properties.
var SP_INITIAL_PAGE_URL string = UP_INITIAL_PAGE_URL
var SP_INITIAL_PAGE_RAW_URL string = UP_INITIAL_PAGE_RAW_URL
var SP_INITIAL_PAGE_DOMAIN string = UP_INITIAL_PAGE_DOMAIN
var SP_INITIAL_PAGE_LOAD_TIME string = UP_INITIAL_PAGE_LOAD_TIME
var SP_INITIAL_PAGE_SPENT_TIME string = UP_INITIAL_PAGE_SPENT_TIME // unit:seconds
var SP_INITIAL_PAGE_SCROLL_PERCENT string = UP_INITIAL_PAGE_SCROLL_PERCENT

// 6Signal Properties
var SIX_SIGNAL_PROPERTIES_PREFIX = "$6Signal_"
var SIX_SIGNAL_ZIP = "$6Signal_zip"
var SIX_SIGNAL_NAICS_DESCRIPTION = "$6Signal_naics_description"
var SIX_SIGNAL_EMPLOYEE_COUNT = "$6Signal_employee_count"
var SIX_SIGNAL_COUNTRY = "$6Signal_country"
var SIX_SIGNAL_ADDRESS = "$6Signal_address"
var SIX_SIGNAL_CITY = "$6Signal_city"
var SIX_SIGNAL_EMPLOYEE_RANGE = "$6Signal_employee_range"
var SIX_SIGNAL_INDUSTRY = "$6Signal_industry"
var SIX_SIGNAL_SIC = "$6Signal_sic"
var SIX_SIGNAL_REVENUE_RANGE = "$6Signal_revenue_range"
var SIX_SIGNAL_COUNTRY_ISO_CODE = "$6Signal_country_iso_code"
var SIX_SIGNAL_PHONE = "$6Signal_phone"
var SIX_SIGNAL_DOMAIN = "$6Signal_domain"
var SIX_SIGNAL_NAME = "$6Signal_name"
var SIX_SIGNAL_STATE = "$6Signal_state"
var SIX_SIGNAL_REGION = "$6Signal_region"
var SIX_SIGNAL_NAICS = "$6Signal_naics"
var SIX_SIGNAL_ANNUAL_REVENUE = "$6Signal_annual_revenue"
var SIX_SIGNAL_SIC_DESCRIPTION = "$6Signal_sic_description"

// Enriched Company properties
var ENRICHED_PROPERTIES_PREFIX = "$enriched_"
var ENRICHMENT_SOURCE = "$enrichment_source"
var ENRICHED_COMPANY_TYPE = "$enriched_company_type"
var ENRICHED_COMPANY_ID = "$enriched_company_id"
var ENRICHED_COMPANY_SUB_INDUSTRY = "$enriched_company_sub_industry"
var ENRICHED_COMPANY_SECTOR = "$enriched_company_sector"
var ENRICHED_COMPANY_INDUSTRY_GROUP = "$enriched_company_industry_group"
var ENRICHED_COMPANY_ALEXA_GLOBAL_RANK = "$enriched_company_alexa_global_rank"
var ENRICHED_COMPANY_FOUNDED_YEAR = "$enriched_company_founded_year"
var ENRICHED_COMPANY_LEGAL_NAME = "$enriched_company_legal_name"
var ENRICHED_COMPANY_ALEXA_US_RANK = "$enriched_company_alexa_us_rank"
var ENRICHED_COMPANY_FUNDING_RAISED = "$enriched_company_funding_raised"
var ENRICHED_COMPANY_MARKET_CAP = "$enriched_company_market_cap"
var ENRICHED_COMPANY_TAGS = "$enriched_company_tags"
var ENRICHED_COMPANY_TECH = "$enriched_company_tech"
var ENRICHED_COMPANY_DESCRIPTION = "$enriched_company_description"
var ENRICHED_COMPANY_LINKEDIN_URL = "$enriched_company_linkedin_url"
var ENRICHED_COMPANY_TRAFFIC_RANK = "$enriched_company_traffic_rank"

// LinkedinCompany properties
var LI_PROPERTIES_PREFIX = "$li_"
var LI_DOMAIN = "$li_domain"
var LI_HEADQUARTER = "$li_headquarter"
var LI_PREFERRED_COUNTRY = "$li_preferred_country"
var LI_VANITY_NAME = "$li_vanity_name"
var LI_LOCALIZED_NAME = "$li_localized_name"
var LI_AD_VIEW_COUNT = "$li_ad_view_count"
var LI_AD_CLICK_COUNT = "$li_ad_click_count"
var LI_ORGANIZATION_ID = "$li_org_id"

// Click properties
var EP_CLICK_ELEMENT_TYPE = "element_type"
var EP_CLICK_CLASS = "class"
var EP_CLICK_ID = "id"
var EP_CLICK_REL = "rel"
var EP_CLICK_ROLE = "role"
var EP_CLICK_TARGET = "target"
var EP_CLICK_HREF = "href"
var EP_CLICK_MEDIA = "media"
var EP_CLICK_TYPE = "type"
var EP_CLICK_NAME = "name"

// g2Company properties
var G2_PROPERTIES_PREFIX = "$g2_"
var G2_DOMAIN = "$g2_domain"
var G2_NAME = "$g2_name"
var G2_LEGAL_NAME = "$g2_legal_name"
var G2_COUNTRY = "$g2_country"
var G2_EMPLOYEES_RANGE = "$g2_employees_range"
var G2_EMPLOYEES = "$g2_employees"
var G2_COMPANY_ID = "$g2_company_id"

// account properties
var IN_HUBSPOT = "$in_hubspot"
var IN_G2 = "$in_g2"
var VISITED_WEBSITE = "$visited_website"
var IN_SALESFORCE = "$in_salesforce"
var IN_LINKEDIN = "$in_linkedin"
var IDENTIFIED_USER_ID = "$identified_user_id"

// SQL column as properties
var CUSTOMER_USER_ID = "customer_user_id"

var DP_DOMAIN_NAME = "$domain_name"

var SDK_ALLOWED_EVENT_PROPERTIES = [...]string{
	EP_INTERNAL_IP,
	EP_LOCATION_LATITUDE,
	EP_LOCATION_LONGITUDE,
	EP_SKIP_SESSION,
	EP_SEGMENT_EVENT_VERSION,
	EP_SEGMENT_SOURCE_LIBRARY,
	EP_SEGMENT_SOURCE_CHANNEL,
	EP_RUDDERSTACK_EVENT_VERSION,
	EP_RUDDERSTACK_SOURCE_LIBRARY,
	EP_RUDDERSTACK_SOURCE_CHANNEL,
	EP_IS_PAGE_VIEW,
	EP_PAGE_TITLE,
	EP_PAGE_DOMAIN,
	EP_PAGE_RAW_URL,
	EP_PAGE_URL,
	EP_REFERRER,
	EP_REFERRER_DOMAIN,
	EP_REFERRER_URL,
	EP_PAGE_LOAD_TIME,
	EP_PAGE_SPENT_TIME,
	EP_PAGE_SCROLL_PERCENT,
	EP_CAMPAIGN,
	EP_CAMPAIGN_ID,
	EP_SOURCE,
	EP_MEDIUM,
	EP_KEYWORD,
	EP_KEYWORD_MATCH_TYPE,
	EP_TERM,
	EP_CONTENT,
	EP_ADGROUP,
	EP_ADGROUP_ID,
	EP_AD,
	EP_AD_ID,
	EP_CREATIVE,
	EP_GCLID,
	EP_FBCLID,
	EP_COST,
	EP_REVENUE,

	// event properties part of offline touch points
	EP_CHANNEL,
	EP_TYPE,
	EP_HUBSPOT_ENGAGEMENT_THREAD_ID,
	EP_OTP_RULE_ID,
	EP_OTP_UNIQUE_KEY,

	// user_properties captured on event. i.e form_submit.
	UP_EMAIL,
	UP_PHONE,
	UP_COMPANY,
	UP_NAME,
	UP_FIRST_NAME,
	UP_LAST_NAME,

	// Form meta properties
	EP_FORM_ID,
	EP_FORM_NAME,
	EP_FORM_CLASS,
	EP_FORM_TYPE,
	EP_FORM_METHOD,
	EP_FORM_TARGET,
	EP_FORM_ACTION,
}

var FORM_SUBMIT_USER_PROPERTIES = [...]string{
	UP_EMAIL,
	UP_PHONE,
	UP_COMPANY,
	UP_NAME,
	UP_FIRST_NAME,
	UP_LAST_NAME,
}

// Event properties that are not visible to user for analysis.
var INTERNAL_EVENT_PROPERTIES = [...]string{
	EP_INTERNAL_IP,
	EP_LOCATION_LATITUDE,
	EP_LOCATION_LONGITUDE,
	EP_SKIP_SESSION,
}

var SDK_ALLOWED_USER_PROPERTIES = [...]string{
	UP_PLATFORM,
	UP_BROWSER,
	UP_BROWSER_VERSION,
	UP_BROWSER_WITH_VERSION,
	UP_USER_AGENT,
	UP_OS,
	UP_OS_VERSION,
	UP_OS_WITH_VERSION,
	UP_SCREEN_WIDTH,
	UP_SCREEN_HEIGHT,
	UP_SCREEN_DENSITY,
	UP_LANGUAGE,
	UP_LOCALE,
	UP_DEVICE_ID,
	UP_DEVICE_NAME,
	UP_DEVICE_ADVERTISING_ID,
	UP_DEVICE_BRAND,
	UP_DEVICE_MODEL,
	UP_DEVICE_TYPE,
	UP_DEVICE_FAMILY,
	UP_DEVICE_MANUFACTURER,
	UP_DEVICE_CARRIER,
	UP_DEVICE_ADTRACKING_ENABLED,
	UP_NETWORK_BLUETOOTH,
	UP_NETWORK_CARRIER,
	UP_NETWORK_CELLULAR,
	UP_NETWORK_WIFI,
	UP_APP_NAME,
	UP_APP_NAMESPACE,
	UP_APP_VERSION,
	UP_APP_BUILD,
	UP_COUNTRY,
	UP_CITY,
	UP_CONTINENT,
	UP_POSTAL_CODE,
	UP_TIMEZONE,
	UP_USER_ID,
	UP_EMAIL,
	UP_COMPANY,
	UP_NAME,
	UP_FIRST_NAME,
	UP_LAST_NAME,
	UP_PHONE,
	UP_INITIAL_PAGE_EVENT_ID,
	UP_INITIAL_PAGE_URL,
	UP_INITIAL_PAGE_DOMAIN,
	UP_INITIAL_PAGE_RAW_URL,
	UP_INITIAL_REFERRER,
	UP_INITIAL_REFERRER_DOMAIN,
	UP_INITIAL_REFERRER_URL,
	UP_INITIAL_PAGE_LOAD_TIME,
	UP_INITIAL_PAGE_SPENT_TIME,
	UP_INITIAL_PAGE_SCROLL_PERCENT,
	UP_INITIAL_CAMPAIGN,
	UP_INITIAL_CAMPAIGN_ID,
	UP_INITIAL_SOURCE,
	UP_INITIAL_MEDIUM,
	UP_INITIAL_KEYWORD,
	UP_INITIAL_KEYWORD_MATCH_TYPE,
	UP_INITIAL_TERM,
	UP_INITIAL_CONTENT,
	UP_INITIAL_ADGROUP,
	UP_INITIAL_ADGROUP_ID,
	UP_INITIAL_CREATIVE,
	UP_INITIAL_GCLID,
	UP_INITIAL_FBCLID,
	UP_DAY_OF_FIRST_EVENT,
	UP_HOUR_OF_FIRST_EVENT,
	UP_LATEST_PAGE_URL,
	UP_PAGE_COUNT,
	UP_TOTAL_SPENT_TIME,
	UP_LATEST_PAGE_DOMAIN,
	UP_LATEST_PAGE_RAW_URL,
	UP_LATEST_PAGE_LOAD_TIME,
	UP_LATEST_PAGE_SPENT_TIME,
	UP_LATEST_PAGE_SCROLL_PERCENT,
	UP_LATEST_CAMPAIGN,
	UP_LATEST_CAMPAIGN_ID,
	UP_LATEST_SOURCE,
	UP_LATEST_MEDIUM,
	UP_LATEST_KEYWORD,
	UP_LATEST_KEYWORD_MATCH_TYPE,
	UP_LATEST_TERM,
	UP_LATEST_CONTENT,
	UP_LATEST_ADGROUP,
	UP_LATEST_ADGROUP_ID,
	UP_LATEST_CREATIVE,
	UP_LATEST_GCLID,
	UP_LATEST_FBCLID,
	UP_LATEST_REFERRER,
	UP_LATEST_REFERRER_URL,
	UP_LATEST_REFERRER_DOMAIN,
}

// Event properties that are not visible to user for analysis.
var INTERNAL_USER_PROPERTIES = [...]string{
	UP_DEVICE_ID,
	"_$deviceId", // Here for legacy reason.
}

var UPDATE_ALLOWED_EVENT_PROPERTIES = [...]string{
	EP_PAGE_SPENT_TIME,
	EP_PAGE_SCROLL_PERCENT,
}

const NAME_PREFIX = "$"
const NAME_PREFIX_ESCAPE_CHAR = "_"
const QUERY_PARAM_PROPERTY_PREFIX = "$qp_"
const QUERY_PARAM_UTM_PREFIX = QUERY_PARAM_PROPERTY_PREFIX + "utm_"
const HUBSPOT_PROPERTY_PREFIX = "$hubspot_"
const SALESFORCE_PROPERTY_PREFIX = "$salesforce_"
const MARKETO_PROPERTY_PREFIX = "$marketo_"
const LEADSQUARED_PROPERTY_PREFIX = "$leadsquared_"

var CRMEventPrefixes = [...]string{
	"$hubspot", "$salesforce", "$sf", "$leadsquared", "$marketo",
}
var AllowedCRMPropertyPrefix = map[string]bool{
	HUBSPOT_PROPERTY_PREFIX:     true,
	SALESFORCE_PROPERTY_PREFIX:  true,
	MARKETO_PROPERTY_PREFIX:     true,
	LEADSQUARED_PROPERTY_PREFIX: true,
}

const (
	SMART_EVENT_SALESFORCE_PREV_PROPERTY = "$prev_salesforce_"
	SMART_EVENT_SALESFORCE_CURR_PROPERTY = "$curr_salesforce_"
	SMART_EVENT_HUBSPOT_PREV_PROPERTY    = "$prev_hubspot_"
	SMART_EVENT_HUBSPOT_CURR_PROPERTY    = "$curr_hubspot_"
)

const (
	PROPERTY_OVERRIDE_BLACKLIST = 1
)

// Platforms
const PLATFORM_WEB = "web"

const (
	PropertyTypeNumerical   = "numerical"
	PropertyTypeCategorical = "categorical"
	PropertyTypeDateTime    = "datetime"
	PropertyTypeUnknown     = "unknown"
)

const (
	DateTimeBreakdownHourlyGranularity  = "hour"
	DateTimeBreakdownDailyGranularity   = "day"
	DateTimeBreakdownWeeklyGranularity  = "week"
	DateTimeBreakdownMonthlyGranularity = "month"
	DateTimeBreakdownYearlyGranularity  = "year"
)

// PROPERTY_VALUE_ANY CRM Rule any value constant
const PROPERTY_VALUE_ANY = "value_any"

var NUMERICAL_PROPERTY_BY_NAME = [...]string{
	EP_PAGE_LOAD_TIME,
	EP_PAGE_SPENT_TIME,
	EP_PAGE_SCROLL_PERCENT,
	EP_REVENUE,
	EP_COST,
	EP_HOUR_OF_DAY,
	UP_INITIAL_PAGE_LOAD_TIME,
	UP_INITIAL_PAGE_SPENT_TIME,
	UP_INITIAL_PAGE_SCROLL_PERCENT,
	UP_SCREEN_WIDTH,
	UP_SCREEN_HEIGHT,
	UP_SCREEN_DENSITY,
	EP_SESSION_COUNT,
	UP_PAGE_COUNT,
	UP_TOTAL_SPENT_TIME,
	UP_LATEST_PAGE_LOAD_TIME,
	UP_LATEST_PAGE_SPENT_TIME,
	UP_HOUR_OF_FIRST_EVENT,
	UP_LATEST_PAGE_SCROLL_PERCENT,
	SIX_SIGNAL_ANNUAL_REVENUE,
	SIX_SIGNAL_EMPLOYEE_COUNT,
}
var CATEGORICAL_PROPERTY_BY_NAME = [...]string{
	EP_CAMPAIGN_ID,
	EP_ADGROUP_ID,
	UP_INITIAL_ADGROUP_ID,
	UP_INITIAL_CAMPAIGN_ID,
	SIX_SIGNAL_ADDRESS,
	SIX_SIGNAL_CITY,
	SIX_SIGNAL_COUNTRY,
	SIX_SIGNAL_COUNTRY_ISO_CODE,
	SIX_SIGNAL_DOMAIN,
	SIX_SIGNAL_EMPLOYEE_RANGE,
	SIX_SIGNAL_INDUSTRY,
	SIX_SIGNAL_NAICS,
	SIX_SIGNAL_NAICS_DESCRIPTION,
	SIX_SIGNAL_NAME,
	SIX_SIGNAL_PHONE,
	SIX_SIGNAL_REGION,
	SIX_SIGNAL_REVENUE_RANGE,
	SIX_SIGNAL_SIC,
	SIX_SIGNAL_SIC_DESCRIPTION,
	SIX_SIGNAL_STATE,
	SIX_SIGNAL_ZIP,
}

var DATETIME_PROPERTY_BY_NAME = [...]string{
	UP_JOIN_TIME,
	EP_TIMESTAMP,
}

var EVENT_TO_USER_INITIAL_PROPERTIES = map[string]string{
	EP_PAGE_URL:            UP_INITIAL_PAGE_URL,
	EP_PAGE_RAW_URL:        UP_INITIAL_PAGE_RAW_URL,
	EP_PAGE_DOMAIN:         UP_INITIAL_PAGE_DOMAIN,
	EP_REFERRER_URL:        UP_INITIAL_REFERRER_URL,
	EP_REFERRER_DOMAIN:     UP_INITIAL_REFERRER_DOMAIN,
	EP_REFERRER:            UP_INITIAL_REFERRER,
	EP_PAGE_LOAD_TIME:      UP_INITIAL_PAGE_LOAD_TIME,
	EP_PAGE_SPENT_TIME:     UP_INITIAL_PAGE_SPENT_TIME,
	EP_PAGE_SCROLL_PERCENT: UP_INITIAL_PAGE_SCROLL_PERCENT,
	EP_CAMPAIGN:            UP_INITIAL_CAMPAIGN,
	EP_CAMPAIGN_ID:         UP_INITIAL_CAMPAIGN_ID,
	EP_SOURCE:              UP_INITIAL_SOURCE,
	EP_MEDIUM:              UP_INITIAL_MEDIUM,
	EP_KEYWORD:             UP_INITIAL_KEYWORD,
	EP_KEYWORD_MATCH_TYPE:  UP_INITIAL_KEYWORD_MATCH_TYPE,
	EP_TERM:                UP_INITIAL_TERM,
	EP_CONTENT:             UP_INITIAL_CONTENT,
	EP_ADGROUP:             UP_INITIAL_ADGROUP,
	EP_ADGROUP_ID:          UP_INITIAL_ADGROUP_ID,
	EP_CREATIVE:            UP_INITIAL_CREATIVE,
	EP_GCLID:               UP_INITIAL_GCLID,
	EP_FBCLID:              UP_INITIAL_FBCLID,
}

var EVENT_TO_USER_LATEST_PAGE_PROPERTIES = map[string]string{
	EP_PAGE_URL:            UP_LATEST_PAGE_URL,
	EP_PAGE_RAW_URL:        UP_LATEST_PAGE_RAW_URL,
	EP_PAGE_DOMAIN:         UP_LATEST_PAGE_DOMAIN,
	EP_REFERRER_URL:        UP_LATEST_REFERRER_URL,
	EP_REFERRER_DOMAIN:     UP_LATEST_REFERRER_DOMAIN,
	EP_REFERRER:            UP_LATEST_REFERRER,
	EP_PAGE_LOAD_TIME:      UP_LATEST_PAGE_LOAD_TIME,
	EP_PAGE_SPENT_TIME:     UP_LATEST_PAGE_SPENT_TIME,
	EP_PAGE_SCROLL_PERCENT: UP_LATEST_PAGE_SCROLL_PERCENT,
}

var EVENT_TO_USER_LATEST_PROPERTIES = map[string]string{
	EP_CAMPAIGN:           UP_LATEST_CAMPAIGN,
	EP_CAMPAIGN_ID:        UP_LATEST_CAMPAIGN_ID,
	EP_SOURCE:             UP_LATEST_SOURCE,
	EP_MEDIUM:             UP_LATEST_MEDIUM,
	EP_KEYWORD:            UP_LATEST_KEYWORD,
	EP_KEYWORD_MATCH_TYPE: UP_LATEST_KEYWORD_MATCH_TYPE,
	EP_TERM:               UP_LATEST_TERM,
	EP_CONTENT:            UP_LATEST_CONTENT,
	EP_ADGROUP:            UP_LATEST_ADGROUP,
	EP_ADGROUP_ID:         UP_LATEST_ADGROUP_ID,
	EP_CREATIVE:           UP_LATEST_CREATIVE,
	EP_GCLID:              UP_LATEST_GCLID,
	EP_FBCLID:             UP_LATEST_FBCLID,
}

// Uses same name as source user properties.
var USER_TO_SESSION_PROPERTIES = [...]string{
	UP_PLATFORM,
	UP_BROWSER,
	UP_BROWSER_VERSION,
	UP_BROWSER_WITH_VERSION,
	UP_USER_AGENT,
	UP_OS,
	UP_OS_VERSION,
	UP_OS_WITH_VERSION,
	UP_COUNTRY,
	UP_CITY,
	UP_CONTINENT,
	UP_POSTAL_CODE,
	UP_REGION,
	UP_TIMEZONE,
}

var EVENT_TO_SESSION_PROPERTIES = map[string]string{
	EP_PAGE_URL:            SP_INITIAL_PAGE_URL,
	EP_PAGE_RAW_URL:        SP_INITIAL_PAGE_RAW_URL,
	EP_PAGE_DOMAIN:         SP_INITIAL_PAGE_DOMAIN,
	EP_PAGE_LOAD_TIME:      SP_INITIAL_PAGE_LOAD_TIME,
	EP_PAGE_SPENT_TIME:     SP_INITIAL_PAGE_SPENT_TIME,
	EP_PAGE_SCROLL_PERCENT: SP_INITIAL_PAGE_SCROLL_PERCENT,

	// Uses same name as event properties.
	EP_CAMPAIGN:           EP_CAMPAIGN,
	EP_CAMPAIGN_ID:        EP_CAMPAIGN_ID,
	EP_SOURCE:             EP_SOURCE,
	EP_MEDIUM:             EP_MEDIUM,
	EP_KEYWORD:            EP_KEYWORD,
	EP_KEYWORD_MATCH_TYPE: EP_KEYWORD_MATCH_TYPE,
	EP_TERM:               EP_TERM,
	EP_CONTENT:            EP_CONTENT,
	EP_ADGROUP:            EP_ADGROUP,
	EP_ADGROUP_ID:         EP_ADGROUP_ID,
	EP_AD:                 EP_AD,
	EP_AD_ID:              EP_AD_ID,
	EP_CREATIVE:           EP_CREATIVE,
	EP_GCLID:              EP_GCLID,
	EP_FBCLID:             EP_FBCLID,

	// Uses session property names.
	EP_REFERRER:        SP_INITIAL_REFERRER,
	EP_REFERRER_URL:    SP_INITIAL_REFERRER_URL,
	EP_REFERRER_DOMAIN: SP_INITIAL_REFERRER_DOMAIN,
}

var DEFINED_MARKETING_PROPERTIES = [...]string{
	EP_CAMPAIGN,
	EP_CAMPAIGN_ID,
	EP_SOURCE,
	EP_MEDIUM,
	EP_KEYWORD,
	EP_KEYWORD_MATCH_TYPE,
	EP_TERM,
	EP_CONTENT,
	EP_ADGROUP,
	EP_ADGROUP_ID,
	EP_AD,
	EP_AD_ID,
	EP_CREATIVE,
	EP_GCLID,
	EP_FBCLID,
}

var PREDEFINED_BIN_RANGES_FOR_PROPERTY = map[string][][2]float64{
	EP_PAGE_LOAD_TIME: [][2]float64{
		[2]float64{1, 1},
		[2]float64{2, 2},
		[2]float64{3, 5},
		[2]float64{5, 10},
		[2]float64{11, 20},
		[2]float64{21, 30},
		[2]float64{31, 60},
		[2]float64{61, 120},
		[2]float64{121, 180},
		[2]float64{181, 300},
		[2]float64{301, 600},
		[2]float64{601, math.MaxFloat64},
	},
	EP_PAGE_SPENT_TIME: [][2]float64{
		//[2]float64{0, 1},
		[2]float64{1, 1},
		[2]float64{2, 2},
		[2]float64{3, 10},
		[2]float64{11, 30},
		[2]float64{31, 60},
		[2]float64{61, 180},
		[2]float64{181, 600},
		[2]float64{601, 1800},
		[2]float64{1801, 3600},
		[2]float64{3601, 21600},
		[2]float64{21601, 43200},
		[2]float64{43201, 84600},
		[2]float64{84601, 259200},
		[2]float64{259201, math.MaxFloat64},
	},
	EP_PAGE_SCROLL_PERCENT: [][2]float64{
		[2]float64{1, 1},
		[2]float64{2, 5},
		[2]float64{5, 12.5},
		[2]float64{12.5, 25},
		[2]float64{25, 37.5},
		[2]float64{37.5, 50},
		[2]float64{50, 62.5},
		[2]float64{62.5, 75},
		[2]float64{75, 87.5},
		[2]float64{87.5, 100},
	},
	UP_TOTAL_SPENT_TIME: [][2]float64{
		//[2]float64{0, 1},
		[2]float64{1, 1},
		[2]float64{2, 2},
		[2]float64{3, 10},
		[2]float64{11, 30},
		[2]float64{31, 60},
		[2]float64{61, 180},
		[2]float64{181, 600},
		[2]float64{601, 1800},
		[2]float64{1801, 3600},
		[2]float64{3601, 21600},
		[2]float64{21601, 43200},
		[2]float64{43201, 84600},
		[2]float64{84601, 259200},
		[2]float64{259201, math.MaxFloat64},
	},
	UP_INITIAL_PAGE_LOAD_TIME: [][2]float64{
		[2]float64{1, 1},
		[2]float64{2, 2},
		[2]float64{3, 5},
		[2]float64{5, 10},
		[2]float64{11, 20},
		[2]float64{21, 30},
		[2]float64{31, 60},
		[2]float64{61, 120},
		[2]float64{121, 180},
		[2]float64{181, 300},
		[2]float64{301, 600},
		[2]float64{601, math.MaxFloat64},
	},
	UP_INITIAL_PAGE_SPENT_TIME: [][2]float64{
		//[2]float64{0, 1},
		[2]float64{1, 1},
		[2]float64{2, 2},
		[2]float64{3, 10},
		[2]float64{11, 30},
		[2]float64{31, 60},
		[2]float64{61, 180},
		[2]float64{181, 600},
		[2]float64{601, 1800},
		[2]float64{1801, 3600},
		[2]float64{3601, 21600},
		[2]float64{21601, 43200},
		[2]float64{43201, 84600},
		[2]float64{84601, 259200},
		[2]float64{259201, math.MaxFloat64},
	},
	UP_LATEST_PAGE_LOAD_TIME: [][2]float64{
		[2]float64{1, 1},
		[2]float64{2, 2},
		[2]float64{3, 5},
		[2]float64{5, 10},
		[2]float64{11, 20},
		[2]float64{21, 30},
		[2]float64{31, 60},
		[2]float64{61, 120},
		[2]float64{121, 180},
		[2]float64{181, 300},
		[2]float64{301, 600},
		[2]float64{601, math.MaxFloat64},
	},
	UP_INITIAL_PAGE_SCROLL_PERCENT: [][2]float64{
		[2]float64{1, 1},
		[2]float64{2, 5},
		[2]float64{5, 12.5},
		[2]float64{12.5, 25},
		[2]float64{25, 37.5},
		[2]float64{37.5, 50},
		[2]float64{50, 62.5},
		[2]float64{62.5, 75},
		[2]float64{75, 87.5},
		[2]float64{87.5, 100},
	},
	SP_INITIAL_PAGE_LOAD_TIME: [][2]float64{
		[2]float64{1, 1},
		[2]float64{2, 2},
		[2]float64{3, 5},
		[2]float64{5, 10},
		[2]float64{11, 20},
		[2]float64{21, 30},
		[2]float64{31, 60},
		[2]float64{61, 120},
		[2]float64{121, 180},
		[2]float64{181, 300},
		[2]float64{301, 600},
		[2]float64{601, math.MaxFloat64},
	},
	SP_INITIAL_PAGE_SPENT_TIME: [][2]float64{
		//[2]float64{0, 1},
		[2]float64{1, 1},
		[2]float64{2, 2},
		[2]float64{3, 10},
		[2]float64{11, 30},
		[2]float64{31, 60},
		[2]float64{61, 180},
		[2]float64{181, 600},
		[2]float64{601, 1800},
		[2]float64{1801, 3600},
		[2]float64{3601, 21600},
		[2]float64{21601, 43200},
		[2]float64{43201, 84600},
		[2]float64{84601, 259200},
		[2]float64{259201, math.MaxFloat64},
	},
	SP_INITIAL_PAGE_SCROLL_PERCENT: [][2]float64{
		[2]float64{1, 1},
		[2]float64{2, 5},
		[2]float64{5, 12.5},
		[2]float64{12.5, 25},
		[2]float64{25, 37.5},
		[2]float64{37.5, 50},
		[2]float64{50, 62.5},
		[2]float64{62.5, 75},
		[2]float64{75, 87.5},
		[2]float64{87.5, 100},
	},
	SP_SPENT_TIME: [][2]float64{
		//[2]float64{0, 1},
		[2]float64{1, 1},
		[2]float64{2, 2},
		[2]float64{3, 10},
		[2]float64{11, 30},
		[2]float64{31, 60},
		[2]float64{61, 180},
		[2]float64{181, 600},
		[2]float64{601, 1800},
		[2]float64{1801, 3600},
		[2]float64{3601, 21600},
		[2]float64{21601, 43200},
		[2]float64{43201, 84600},
		[2]float64{84601, 259200},
		[2]float64{259201, math.MaxFloat64},
	},
	SP_PAGE_COUNT: [][2]float64{
		//[2]float64{0, 1},
		[2]float64{1, 1},
		[2]float64{2, 2},
		[2]float64{3, 3},
		[2]float64{4, 4},
		[2]float64{5, 5},
		[2]float64{6, 10},
		[2]float64{11, 25},
		[2]float64{26, 50},
		[2]float64{51, 100},
		[2]float64{101, math.MaxFloat64},
	},
	UP_PAGE_COUNT: [][2]float64{
		//[2]float64{0, 1},
		[2]float64{1, 1},
		[2]float64{2, 2},
		[2]float64{3, 3},
		[2]float64{4, 4},
		[2]float64{5, 5},
		[2]float64{6, 10},
		[2]float64{11, 25},
		[2]float64{26, 50},
		[2]float64{51, 100},
		[2]float64{101, math.MaxFloat64},
	},
}

// DISABLED_CORE_QUERY_USER_PROPERTIES Less important user properties in core query context.
var DISABLED_CORE_QUERY_USER_PROPERTIES = [...]string{
	UP_DEVICE_ADTRACKING_ENABLED,
	UP_NETWORK_BLUETOOTH,
	UP_NETWORK_CARRIER,
	UP_NETWORK_CELLULAR,
	UP_NETWORK_WIFI,
	UP_DEVICE_ADVERTISING_ID,
	UP_DEVICE_ID,
	UP_MERGE_TIMESTAMP,
	UP_INITIAL_PAGE_EVENT_ID,
	UP_META_OBJECT_IDENTIFIER_KEY,
	EP_CRM_REFERENCE_EVENT_ID,
	"$marketo_lead__fivetran_synced",
}

// DISABLED_CORE_QUERY_EVENT_PROPERTIES Less important event properties in core query context.
var DISABLED_CORE_QUERY_EVENT_PROPERTIES = [...]string{
	EP_INTERNAL_IP,
	EP_LOCATION_LATITUDE,
	EP_LOCATION_LONGITUDE,
	EP_SEGMENT_EVENT_VERSION,
	EP_RUDDERSTACK_EVENT_VERSION,
	EP_CRM_REFERENCE_EVENT_ID,
	EP_SKIP_SESSION,
	"$marketo_lead__fivetran_synced",
}

var SESSION_PROPERTIES_SET_IN_UPDATE = map[string]bool{
	EP_CHANNEL:             true,
	SP_SPENT_TIME:          true,
	SP_PAGE_COUNT:          true,
	SP_LATEST_PAGE_RAW_URL: true,
	SP_LATEST_PAGE_URL:     true,
}

var DISABLED_USER_PROPERTIES_UI = [...]string{
	UP_USER_AGENT,
	UP_BROWSER_WITH_VERSION,
	UP_OS_WITH_VERSION,
	UP_SESSION_COUNT,
}

var DISABLED_EVENT_PROPERTIES_UI = [...]string{
	UP_USER_AGENT,
	UP_BROWSER_WITH_VERSION,
	UP_OS_WITH_VERSION,
}

// DISABLED_FACTORS_USER_PROPERTIES User properties disabled for the factors analysis.
var DISABLED_FACTORS_USER_PROPERTIES = [...]string{
	UP_BROWSER_VERSION,
	UP_OS_VERSION,
	UP_DEVICE_ID,
	UP_DEVICE_ADVERTISING_ID,
	UP_DEVICE_ADTRACKING_ENABLED,
	UP_NETWORK_BLUETOOTH,
	UP_NETWORK_CARRIER,
	UP_NETWORK_CELLULAR,
	UP_NETWORK_WIFI,
	UP_APP_BUILD,
	UP_USER_ID,
	UP_INITIAL_GCLID,
	UP_INITIAL_FBCLID,
	UP_LATEST_GCLID,
	UP_LATEST_FBCLID,
	UP_LATEST_REFERRER,
	UP_INITIAL_REFERRER,
	UP_MERGE_TIMESTAMP,
	UP_INITIAL_PAGE_EVENT_ID,
	UP_META_OBJECT_IDENTIFIER_KEY,
}

var DISABLED_EVENT_USER_LEVEL_PROPERTIES = []string{

	UP_SESSION_COUNT,
	UP_PAGE_COUNT,
	UP_TOTAL_SPENT_TIME,
}

// DISABLED_FACTORS_EVENT_PROPERTIES Event properties disabled for the factors analysis.
var DISABLED_FACTORS_EVENT_PROPERTIES = [...]string{
	EP_INTERNAL_IP,
	EP_LOCATION_LATITUDE,
	EP_LOCATION_LONGITUDE,
	EP_SEGMENT_EVENT_VERSION,
	EP_SEGMENT_SOURCE_LIBRARY,
	EP_SEGMENT_SOURCE_CHANNEL,
	EP_RUDDERSTACK_EVENT_VERSION,
	EP_RUDDERSTACK_SOURCE_LIBRARY,
	EP_RUDDERSTACK_SOURCE_CHANNEL,
	EP_PAGE_RAW_URL,
	EP_GCLID,
	EP_FBCLID,
	UP_EMAIL,
	UP_JOIN_TIME,
	UP_OS_WITH_VERSION,
	UP_HOUR_OF_FIRST_EVENT,
	UP_DAY_OF_FIRST_EVENT,
}

var DEFAULT_EVENT_PROPERTY_VALUES = map[string]interface{}{
	EP_PAGE_SPENT_TIME:     1, // 1 second
	EP_PAGE_LOAD_TIME:      1, // 1 second
	EP_PAGE_SCROLL_PERCENT: 0,
}

var DEFAULT_USER_PROPERTY_VALUES = map[string]interface{}{
	UP_INITIAL_PAGE_SPENT_TIME:     DEFAULT_EVENT_PROPERTY_VALUES[EP_PAGE_SPENT_TIME],
	UP_INITIAL_PAGE_LOAD_TIME:      DEFAULT_EVENT_PROPERTY_VALUES[EP_PAGE_LOAD_TIME],
	UP_INITIAL_PAGE_SCROLL_PERCENT: DEFAULT_EVENT_PROPERTY_VALUES[EP_PAGE_SCROLL_PERCENT],
}

// ITREE_PROPERTIES_TO_IGNORE Predefined properties that do not add much insights.
var ITREE_PROPERTIES_TO_IGNORE = map[string]bool{
	UP_BROWSER_VERSION: true,
	"$browserVersion":  true, // Deprecated standard properties.
	"_$browserVersion": true,
	UP_SCREEN_HEIGHT:   true,
	"$screenHeight":    true,
	"_$screenHeight":   true,
	UP_SCREEN_WIDTH:    true,
	"$screenWidth":     true,
	"_$screenWidth":    true,
	UP_OS_VERSION:      true,
	"$osVersion":       true,
	"_$osVersion":      true,
	UP_JOIN_TIME:       true,
	"_$joinTime":       true,
	// Old incorrect property.
	"$session":              true,
	UP_BROWSER_WITH_VERSION: true,
	UP_USER_AGENT:           true,
	UP_BROWSER:              true,
	EP_IS_PAGE_VIEW:         true,

	UP_INITIAL_PAGE_DOMAIN:  true,
	UP_INITIAL_PAGE_URL:     true,
	UP_INITIAL_PAGE_RAW_URL: true,
	EP_PAGE_DOMAIN:          true,
	EP_PAGE_RAW_URL:         true,
	EP_PAGE_TITLE:           true,
	EP_TIMESTAMP:            true,
	// Temporary fix.
	EP_REFERRER_URL:                true,
	EP_REFERRER_DOMAIN:             true,
	SP_INITIAL_REFERRER_DOMAIN:     true,
	SP_INITIAL_REFERRER:            true,
	UP_INITIAL_PAGE_SPENT_TIME:     true,
	UP_INITIAL_PAGE_SCROLL_PERCENT: true,

	// Session Latest.
	SP_LATEST_PAGE_RAW_URL: true,
	SP_LATEST_PAGE_URL:     true,

	// Counts being seen as categorical.
	UP_PAGE_COUNT:       true,
	SP_PAGE_COUNT:       true,
	EP_SESSION_COUNT:    true,
	SP_SESSION_TIME:     true,
	SP_SPENT_TIME:       true,
	UP_TOTAL_SPENT_TIME: true,
}

var ITREE_NUMERICAL_PROPERTIES_TO_IGNORE = map[string]bool{
	"$campaign":         true,
	"$initial_campaign": true,
}

var STANDARD_EVENTS_DISPLAY_NAMES = map[string]string{
	"$hubspot_contact_created":                  "Contact Created",
	"$hubspot_contact_updated":                  "Contact Updated",
	"$hubspot_deal_state_changed":               "Deal State Changed",
	"$hubspot_form_submission":                  "Hubspot Form Submissions",
	"$hubspot_engagement_email":                 "Engagement Email",
	"$hubspot_engagement_meeting_created":       "Engagement Meeting Created",
	"$hubspot_engagement_meeting_updated":       "Engagement Meeting Updated",
	"$hubspot_engagement_call_created":          "Engagement Call Created",
	"$hubspot_engagement_call_updated":          "Engagement Call Updated",
	"$hubspot_contact_list":                     "Contact List",
	"$sf_contact_created":                       "Contact Created",
	"$sf_contact_updated":                       "Contact Updated",
	"$sf_lead_created":                          "Lead Created",
	"$sf_lead_updated":                          "Lead Updated",
	"$sf_account_created":                       "Account Created",
	"$sf_account_updated":                       "Account Updated",
	"$sf_opportunity_created":                   "Opportunity Created",
	"$sf_opportunity_updated":                   "Opportunity Updated",
	"$sf_campaign_member_created":               "Added to Campaign",
	"$sf_campaign_member_updated":               "Interacted with Campaign",
	"$sf_campaign_member_responded_to_campaign": "Responded to Campaign",
	"$session":                           "Website Session",
	"$form_submitted":                    "Form Button Click",
	"$hubspot_company_created":           "Company Created",
	"$hubspot_company_updated":           "Company Updated",
	"$hubspot_deal_created":              "Deal Created",
	"$hubspot_deal_updated":              "Deal Updated",
	"$salesforce_account_updated":        "Salesforce Account Updated",
	"$salesforce_opportunity_updated":    "Salesforce Opportunity Updated",
	"$salesforce_account_created":        "Salesforce Account Created",
	"$salesforce_opportunity_created":    "Salesforce Opportunity Created",
	"$offline_touch_point":               "Offline Touchpoint",
	"$leadsquared_lead_created":          "Lead Created",
	"$leadsquared_lead_updated":          "Lead Updated",
	EVENT_NAME_FORM_FILL:                 "Form Fills",
	EVENT_NAME_SALESFORCE_TASK_CREATED:   "Salesforce Task Created",
	EVENT_NAME_SALESFORCE_EVENT_CREATED:  "Salesforce Event Created",
	GROUP_EVENT_NAME_G2_ALL:              "G2 All",
	GROUP_EVENT_NAME_G2_SPONSORED:        "Saw ad on competitor's page",
	GROUP_EVENT_NAME_G2_PRODUCT_PROFILE:  "Looked at product page",
	GROUP_EVENT_NAME_G2_ALTERNATIVE:      "Looked at alternatives",
	GROUP_EVENT_NAME_G2_PRICING:          "Looked at pricing",
	GROUP_EVENT_NAME_G2_CATEGORY:         "Looked at product category",
	GROUP_EVENT_NAME_G2_COMPARISON:       "Compared with other products",
	GROUP_EVENT_NAME_G2_REPORT:           "Looked at grid report",
	GROUP_EVENT_NAME_G2_REFERENCE:        "Looked at reference page",
	GROUP_EVENT_NAME_G2_DEAL:             "Looked at deal page",
	GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD: "Linkedin Ad Clicked",
	GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD:  "Linkedin Ad Viewed",
}

var STANDARD_GROUP_DISPLAY_NAMES = map[string]string{
	GROUP_NAME_HUBSPOT_COMPANY:        "Hubspot Companies",
	GROUP_NAME_HUBSPOT_DEAL:           "Hubspot Deals",
	GROUP_NAME_SALESFORCE_ACCOUNT:     "Salesforce Accounts",
	GROUP_NAME_SALESFORCE_OPPORTUNITY: "Salesforce Opportunities",
	GROUP_NAME_SIX_SIGNAL:             "Identified Companies",
	GROUP_NAME_LINKEDIN_COMPANY:       "Linkedin Company Engagements",
	GROUP_NAME_G2:                     "G2 Engagements",
}

var ALL_ACCOUNT_DEFAULT_PROPERTIES = []string{
	IN_LINKEDIN,
	IN_HUBSPOT,
	IN_G2,
	VISITED_WEBSITE,
	IN_SALESFORCE,
}

var GROUP_TO_DEFAULT_SEGMENT_MAP = map[string]string{
	GROUP_NAME_HUBSPOT_COMPANY:    IN_HUBSPOT,
	GROUP_NAME_SALESFORCE_ACCOUNT: IN_SALESFORCE,
	GROUP_NAME_LINKEDIN_COMPANY:   IN_LINKEDIN,
	GROUP_NAME_G2:                 IN_G2,
}

var ALL_ACCOUNT_DEFAULT_PROPERTIES_DISPLAY_NAMES = map[string]string{
	IN_LINKEDIN:                       "Engaged on LinkedIn",
	IN_HUBSPOT:                        "In Hubspot",
	IN_G2:                             "Visited G2",
	VISITED_WEBSITE:                   "Visited Website",
	IN_SALESFORCE:                     "In Salesforce",
	GROUP_EVENT_NAME_ENGAGEMENT_SCORE: "Engagement Score",
}

var USER_PROPERTIES_WITH_COLUMN = []string{
	IDENTIFIED_USER_ID,
}

var USER_PROPERTIES_WITH_COLUMN_DISPLAY_NAMES = map[string]string{
	IDENTIFIED_USER_ID: "Identified User Id",
}

var CRM_USER_EVENT_NAME_LABELS = map[string]string{
	"$hubspot_contact_created":                  "Hubspot Contacts",
	"$hubspot_contact_updated":                  "Hubspot Contacts",
	"$hubspot_engagement_email":                 "Hubspot Contacts",
	"$hubspot_engagement_meeting_created":       "Hubspot Contacts",
	"$hubspot_engagement_meeting_updated":       "Hubspot Contacts",
	"$hubspot_engagement_call_created":          "Hubspot Contacts",
	"$hubspot_engagement_call_updated":          "Hubspot Contacts",
	"$hubspot_contact_list":                     "Hubspot Contacts",
	"$marketo_lead_created":                     "Marketo Person",
	"$marketo_lead_updated":                     "Marketo Person",
	"$leadsquared_lead_created":                 "LeadSquared Person",
	"$leadsquared_lead_updated":                 "LeadSquared Person",
	"$sf_contact_created":                       "Salesforce Users",
	"$sf_contact_updated":                       "Salesforce Users",
	"$sf_lead_created":                          "Salesforce Users",
	"$sf_lead_updated":                          "Salesforce Users",
	"$sf_campaign_member_created":               "Salesforce Users",
	"$sf_campaign_member_updated":               "Salesforce Users",
	"$sf_campaign_member_responded_to_campaign": "Salesforce Users",
	"$sf_task_created":                          "Salesforce Users",
	"$sf_task_updated":                          "Salesforce Users",
	"$sf_event_created":                         "Salesforce Users",
	"$sf_event_updated":                         "Salesforce Users",
}

var STANDARD_EVENTS_GROUP_NAMES = map[string]string{
	"$hubspot_contact_created":                  "Hubspot",
	"$hubspot_contact_updated":                  "Hubspot",
	"$hubspot_deal_state_changed":               "Hubspot",
	"$hubspot_company_created":                  "Hubspot",
	"$hubspot_company_updated":                  "Hubspot",
	"$hubspot_deal_created":                     "Hubspot",
	"$hubspot_deal_updated":                     "Hubspot",
	"$hubspot_form_submission":                  "Hubspot",
	"$hubspot_contact_list":                     "Hubspot",
	"$sf_contact_created":                       "Salesforce",
	"$sf_contact_updated":                       "Salesforce",
	"$sf_lead_created":                          "Salesforce",
	"$sf_lead_updated":                          "Salesforce",
	"$sf_account_created":                       "Salesforce",
	"$sf_account_updated":                       "Salesforce",
	"$sf_opportunity_created":                   "Salesforce",
	"$sf_opportunity_updated":                   "Salesforce",
	"$sf_campaign_member_created":               "Salesforce",
	"$sf_campaign_member_updated":               "Salesforce",
	"$sf_campaign_member_responded_to_campaign": "Salesforce",
	"$salesforce_account_updated":               "Salesforce",
	"$salesforce_opportunity_updated":           "Salesforce",
	"$salesforce_account_created":               "Salesforce",
	"$salesforce_opportunity_created":           "Salesforce",
	"$leadsquared_lead_created":                 "LeadSquared",
	"$leadsquared_lead_updated":                 "LeadSquared",
	"$leadsquared_sales_activity_created":       "LeadSquared",
	"$leadsquared_sales_activity_updated":       "LeadSquared",
	"$leadsquared_email_sent_activity_created":  "LeadSquared",
	"$leadsquared_email_info_activity_created":  "LeadSquared",
	"$leadsquared_had_a_call_activity_updated":  "LeadSquared",
	"$leadsquared_had_a_call_activity_created":  "LeadSquared",
}

var STANDARD_EVENTS_IN_DROPDOWN = []string{
	"$hubspot_contact_created",
	"$hubspot_contact_updated",
	"$hubspot_contact_list",
	"$hubspot_company_created",
	"$hubspot_company_updated",
	"$hubspot_deal_created",
	"$hubspot_deal_updated",
	"$hubspot_form_submission",
	"$hubspot_engagement_email",
	"$hubspot_engagement_meeting_created",
	"$hubspot_engagement_meeting_updated",
	"$hubspot_engagement_call_created",
	"$hubspot_engagement_call_updated",
	"$sf_contact_created",
	"$sf_contact_updated",
	"$sf_lead_created",
	"$sf_lead_updated",
	"$sf_campaign_member_created",
	"$sf_campaign_member_updated",
	"$sf_campaign_member_responded_to_campaign",
	"$sf_task_created",
	"$sf_task_updated",
	"$sf_event_created",
	"$sf_event_updated",
	"$salesforce_account_created",
	"$salesforce_account_updated",
	"$salesforce_opportunity_created",
	"$salesforce_opportunity_updated",
	"$marketo_lead_created",
	"$marketo_lead_updated",
	"$marketo_program_membership_created",
	"$marketo_program_membership_updated",
	"$leadsquared_lead_created",
	"$leadsquared_lead_updated",
	"$leadsquared_sales_activity_created",
	"$leadsquared_sales_activity_updated",
	"$leadsquared_had_a_call_activity_created",
	"$leadsquared_had_a_call_activity_updated",
	"$leadsquared_email_sent_activity_created",
	"$leadsquared_email_info_activity_created",
}

var STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES = map[string]string{
	EP_IS_PAGE_VIEW:                          "Is page view",
	EP_PAGE_TITLE:                            "Page title",
	EP_PAGE_DOMAIN:                           "Page domain",
	EP_PAGE_RAW_URL:                          "Page raw URL",
	EP_PAGE_URL:                              "Page URL",
	EP_REFERRER:                              "Page referrer",
	EP_REFERRER_DOMAIN:                       "Page referrer domain",
	EP_REFERRER_URL:                          "Page referrer URL",
	EP_PAGE_LOAD_TIME:                        "Page load time",
	EP_PAGE_SPENT_TIME:                       "Page active time",
	EP_PAGE_SCROLL_PERCENT:                   "Page scroll percent",
	EP_SEGMENT_EVENT_VERSION:                 "Segment Event Version",
	EP_SEGMENT_SOURCE_LIBRARY:                "Segment Source Library",
	EP_SEGMENT_SOURCE_CHANNEL:                "Segment Source Channel",
	EP_RUDDERSTACK_EVENT_VERSION:             "Rudderstack Event Version",
	EP_RUDDERSTACK_SOURCE_LIBRARY:            "Rudderstack Source Library",
	EP_RUDDERSTACK_SOURCE_CHANNEL:            "Rudderstack Source Channel",
	EP_CAMPAIGN:                              "Campaign",
	EP_CAMPAIGN_ID:                           "Campaign ID",
	EP_SOURCE:                                "Source",
	EP_MEDIUM:                                "Medium",
	EP_KEYWORD:                               "Keyword",
	EP_KEYWORD_MATCH_TYPE:                    "Keyword Match UnitType",
	EP_TERM:                                  "Term",
	EP_CONTENT:                               "Content",
	EP_ADGROUP:                               "Adgroup",
	EP_ADGROUP_ID:                            "Adgroup ID",
	EP_CREATIVE:                              "Creative",
	EP_GCLID:                                 "GCLID",
	EP_FBCLID:                                "FBCLID",
	EP_COST:                                  "Cost",
	EP_REVENUE:                               "Revenue",
	EP_TIMESTAMP:                             "Timestamp",
	EP_HOUR_OF_DAY:                           "Hour of occurrence",
	EP_DAY_OF_WEEK:                           "Day of occurrence",
	EP_SESSION_COUNT:                         "Session Count",
	EP_TERM:                                  "Term",
	EP_CHANNEL:                               "Channel",
	UP_POSTAL_CODE:                           "Postal Code",
	UP_CONTINENT:                             "Continent",
	EP_FORM_ID:                               "Form Id",
	EP_FORM_NAME:                             "Form Name",
	EP_FORM_CLASS:                            "Form Class",
	EP_FORM_TARGET:                           "Form Target",
	EP_FORM_METHOD:                           "Form Method",
	EP_FORM_ACTION:                           "Form Action",
	EP_FORM_TYPE:                             "Form Type",
	EP_G2_TAG:                                "G2 Tag",
	EP_G2_CITY:                               "G2 Visitor's City",
	EP_G2_STATE:                              "G2 Visitor's State",
	EP_G2_COUNTRY:                            "G2 Visitor's Country",
	EP_G2_CATEGORY_IDS:                       "G2 Category IDs",
	EP_G2_PRODUCT_IDS:                        "G2 Product IDs",
	"$hubspot_form_submission_form-type":     "Form Type",
	"$hubspot_form_submission_title":         "Form Title",
	"$hubspot_form_submission_form-id":       "Form ID",
	"$hubspot_form_submission_conversion-id": "Conversion ID",
	"$hubspot_form_submission_email":         "Email",
	"$hubspot_form_submission_page-url":      "Page Raw URL",
	"$hubspot_form_submission_page-url-no-qp": "Page URL",
	"$hubspot_form_submission_page-title":     "Page Title",
	"utm_source":                              "Source",
	"utm_campaign":                            "Campaign",
	"utm_medium":                              "Medium",
	"utm_content":                             "Content",
	"utm_term":                                "Term",
	"utm_name":                                "Name",
	"$hubspot_form_submission_phone":          "Phone",
	"$hubspot_form_submission_timestamp":      "Form Submit Timestamp",
	"$hubspot_form_submission_portal-id":      "Portal ID",
	"Source-Medium":                           "Source Medium",
	"page_url":                                "Page URL",
}

var STANDARD_EVENT_PROPERTIES_CATAGORIZATION = map[string]string{
	EP_IS_PAGE_VIEW:        "Page properties",
	EP_PAGE_TITLE:          "Page properties",
	EP_PAGE_DOMAIN:         "Page properties",
	EP_PAGE_RAW_URL:        "Page properties",
	EP_PAGE_URL:            "Page properties",
	EP_REFERRER:            "Traffic source",
	EP_REFERRER_DOMAIN:     "Traffic source",
	EP_REFERRER_URL:        "Traffic source",
	EP_PAGE_LOAD_TIME:      "Page properties",
	EP_PAGE_SPENT_TIME:     "Page properties",
	EP_PAGE_SCROLL_PERCENT: "Page properties",
	EP_CAMPAIGN:            "Traffic source",
	EP_CAMPAIGN_ID:         "Traffic source",
	EP_SOURCE:              "Traffic source",
	EP_MEDIUM:              "Traffic source",
	EP_KEYWORD:             "Traffic source",
	EP_KEYWORD_MATCH_TYPE:  "Traffic source",
	EP_TERM:                "Traffic source",
	EP_CONTENT:             "Traffic source",
	EP_ADGROUP:             "Traffic source",
	EP_ADGROUP_ID:          "Traffic source",
	EP_CREATIVE:            "Traffic source",
	EP_GCLID:               "Traffic source",
	EP_FBCLID:              "Traffic source",
	EP_TIMESTAMP:           "Session properties",
	EP_HOUR_OF_DAY:         "Session properties",
	EP_DAY_OF_WEEK:         "Session properties",
	EP_SESSION_COUNT:       "Session properties",
	EP_CHANNEL:             "Traffic source",
	UP_POSTAL_CODE:         "User identification",
	UP_CONTINENT:           "User identification",
	SP_IS_FIRST_SESSION:    "Session properties",
	SP_SESSION_TIME:        "Session properties",
	SP_SPENT_TIME:          "Session properties",
	SP_PAGE_COUNT:          "Session properties",
	EP_G2_TAG:              "G2 Properties",
	EP_G2_CITY:             "G2 Properties",
	EP_G2_STATE:            "G2 Properties",
	EP_G2_COUNTRY:          "G2 Properties",
	EP_G2_CATEGORY_IDS:     "G2 Properties",
	EP_G2_PRODUCT_IDS:      "G2 Properties",
}

// GetStandardUserPropertiesBasedOnIntegration is using this.
// Separate logic for integration based properties is there.
var STANDARD_USER_PROPERTIES_DISPLAY_NAMES = map[string]string{
	UP_PLATFORM:                        "User platform",
	UP_BROWSER:                         "User browser",
	UP_BROWSER_VERSION:                 "User browser version",
	UP_OS:                              "User OS",
	UP_OS_VERSION:                      "User OS version",
	UP_SCREEN_WIDTH:                    "Screen width",
	UP_SCREEN_HEIGHT:                   "Screen height",
	UP_SCREEN_DENSITY:                  "Screen density",
	UP_LANGUAGE:                        "Language",
	UP_LOCALE:                          "Locale",
	UP_DEVICE_NAME:                     "Device name",
	UP_DEVICE_BRAND:                    "Device brand",
	UP_DEVICE_MODEL:                    "Device model",
	UP_DEVICE_TYPE:                     "Device type",
	UP_DEVICE_FAMILY:                   "Device family",
	UP_DEVICE_MANUFACTURER:             "Device manufacturer",
	UP_DEVICE_CARRIER:                  "Device carrier",
	UP_COUNTRY:                         "User country",
	UP_CITY:                            "User city",
	UP_REGION:                          "User region",
	UP_TIMEZONE:                        "User timezone",
	UP_USER_ID:                         "User ID",
	UP_EMAIL:                           "User email ID",
	UP_COMPANY:                         "Company",
	UP_NAME:                            "User Name",
	UP_FIRST_NAME:                      "User first Name",
	UP_LAST_NAME:                       "User last Name",
	UP_PHONE:                           "User phone",
	UP_INITIAL_PAGE_URL:                "User first page URL",
	UP_INITIAL_PAGE_DOMAIN:             "User first page domain",
	UP_INITIAL_PAGE_RAW_URL:            "User first page raw URL",
	UP_INITIAL_PAGE_LOAD_TIME:          "User first page load time",
	UP_INITIAL_PAGE_SPENT_TIME:         "User first page active time",
	UP_INITIAL_PAGE_SCROLL_PERCENT:     "User first page scroll percent",
	UP_INITIAL_CAMPAIGN:                "User first campaign",
	UP_INITIAL_CAMPAIGN_ID:             "User first campaign ID",
	UP_INITIAL_SOURCE:                  "User first source",
	UP_INITIAL_MEDIUM:                  "User first medium",
	UP_INITIAL_KEYWORD:                 "User first keyword",
	UP_INITIAL_KEYWORD_MATCH_TYPE:      "User first keyword match type",
	UP_INITIAL_TERM:                    "User first search term",
	UP_INITIAL_CONTENT:                 "User first content",
	UP_INITIAL_ADGROUP:                 "User first adgroup",
	UP_INITIAL_ADGROUP_ID:              "User first adgroup ID",
	UP_INITIAL_CREATIVE:                "User first creative",
	UP_INITIAL_GCLID:                   "User first GCLID",
	UP_INITIAL_FBCLID:                  "User first FBCLID",
	UP_INITIAL_REFERRER:                "User first referrer",
	UP_INITIAL_REFERRER_URL:            "User first referrer URL",
	UP_INITIAL_REFERRER_DOMAIN:         "User first referrer domain",
	UP_INITIAL_CHANNEL:                 "User first channel",
	UP_DAY_OF_FIRST_EVENT:              "First seen day",
	UP_HOUR_OF_FIRST_EVENT:             "First seen hour",
	UP_PAGE_COUNT:                      "User page count",
	UP_TOTAL_SPENT_TIME:                "User total active time",
	UP_LATEST_PAGE_URL:                 "User latest page URL",
	UP_LATEST_PAGE_DOMAIN:              "User latest page domain",
	UP_LATEST_PAGE_RAW_URL:             "User latest page raw URL",
	UP_LATEST_PAGE_LOAD_TIME:           "User latest page load time",
	UP_LATEST_PAGE_SPENT_TIME:          "User latest page active time",
	UP_LATEST_PAGE_SCROLL_PERCENT:      "User latest page scroll percent",
	UP_LATEST_CAMPAIGN:                 "User latest campaign",
	UP_LATEST_CAMPAIGN_ID:              "User latest campaign ID",
	UP_LATEST_SOURCE:                   "User latest source",
	UP_LATEST_MEDIUM:                   "User latest medium",
	UP_LATEST_KEYWORD:                  "User latest keyword",
	UP_LATEST_KEYWORD_MATCH_TYPE:       "User latest keyword match type",
	UP_LATEST_TERM:                     "User latest search term",
	UP_LATEST_CONTENT:                  "User latest content",
	UP_LATEST_ADGROUP:                  "User latest adgroup",
	UP_LATEST_ADGROUP_ID:               "User latest adgroup ID",
	UP_LATEST_CREATIVE:                 "User latest creative",
	UP_LATEST_GCLID:                    "User latest GCLID",
	UP_LATEST_FBCLID:                   "User latest FBCLID",
	UP_LATEST_REFERRER:                 "User latest referrer",
	UP_LATEST_REFERRER_URL:             "User latest referrer URL",
	UP_LATEST_REFERRER_DOMAIN:          "User latest referrer domain",
	UP_LATEST_CHANNEL:                  "User latest channel",
	UP_JOIN_TIME:                       "First seen date",
	UP_POSTAL_CODE:                     "User postal code",
	UP_CONTINENT:                       "User continent",
	SIX_SIGNAL_ADDRESS:                 "Company HQ address",
	SIX_SIGNAL_ANNUAL_REVENUE:          "Company annual revenue",
	SIX_SIGNAL_CITY:                    "Company HQ city",
	SIX_SIGNAL_COUNTRY:                 "Company country",
	SIX_SIGNAL_COUNTRY_ISO_CODE:        "Company country ISO code",
	SIX_SIGNAL_DOMAIN:                  "Company domain",
	SIX_SIGNAL_EMPLOYEE_COUNT:          "Company employee count",
	SIX_SIGNAL_EMPLOYEE_RANGE:          "Company employee range",
	SIX_SIGNAL_INDUSTRY:                "Company industry",
	SIX_SIGNAL_NAICS:                   "Company NAICS code",
	SIX_SIGNAL_NAICS_DESCRIPTION:       "Company NAICS description",
	SIX_SIGNAL_NAME:                    "Company name",
	SIX_SIGNAL_PHONE:                   "Company phone",
	SIX_SIGNAL_REGION:                  "Company region",
	SIX_SIGNAL_REVENUE_RANGE:           "Company annual revenue range",
	SIX_SIGNAL_SIC:                     "Company SIC code",
	SIX_SIGNAL_SIC_DESCRIPTION:         "Company SIC description",
	SIX_SIGNAL_STATE:                   "Company state",
	SIX_SIGNAL_ZIP:                     "Company ZIP code",
	ENRICHED_COMPANY_TYPE:              "Company Type",
	ENRICHED_COMPANY_ID:                "Company ID",
	ENRICHED_COMPANY_SUB_INDUSTRY:      "Company Sub Industry",
	ENRICHED_COMPANY_SECTOR:            "Company Sector",
	ENRICHED_COMPANY_INDUSTRY_GROUP:    "Company Industry Group",
	ENRICHED_COMPANY_ALEXA_GLOBAL_RANK: "Company Alexa Global Rank",
	ENRICHED_COMPANY_FOUNDED_YEAR:      "Company Founded Year",
	ENRICHED_COMPANY_LEGAL_NAME:        "Company Legal Name",
	ENRICHED_COMPANY_ALEXA_US_RANK:     "Company Alexa US Rank",
	ENRICHED_COMPANY_FUNDING_RAISED:    "Company Funding Raised",
	ENRICHED_COMPANY_MARKET_CAP:        "Company Market Cap",
	ENRICHED_COMPANY_TAGS:              "Company Tags",
	ENRICHED_COMPANY_TECH:              "Company Tech",
	ENRICHED_COMPANY_DESCRIPTION:       "Company Description",
	ENRICHED_COMPANY_LINKEDIN_URL:      "Company Linkedin URL",
	ENRICHED_COMPANY_TRAFFIC_RANK:      "Company Traffic Rank",
	G2_DOMAIN:                          "G2 Company Domain",
	G2_NAME:                            "G2 Company Name",
	G2_LEGAL_NAME:                      "G2 Company Legal Name",
	G2_COUNTRY:                         "G2 Company Country",
	G2_EMPLOYEES_RANGE:                 "G2 Company Employee Range",
	G2_EMPLOYEES:                       "G2 No Of Employees",
	G2_COMPANY_ID:                      "G2 Company ID",
}

var STANDARD_USER_PROPERTIES_CATAGORIZATION = map[string]string{
	UP_PLATFORM:                    "Platform/Device",
	UP_BROWSER:                     "Platform/Device",
	UP_BROWSER_VERSION:             "Platform/Device",
	UP_OS:                          "Platform/Device",
	UP_OS_VERSION:                  "Platform/Device",
	UP_SCREEN_WIDTH:                "Platform/Device",
	UP_SCREEN_HEIGHT:               "Platform/Device",
	UP_SCREEN_DENSITY:              "Platform/Device",
	UP_LANGUAGE:                    "User identification",
	UP_LOCALE:                      "User identification",
	UP_DEVICE_NAME:                 "Platform/Device",
	UP_DEVICE_BRAND:                "Platform/Device",
	UP_DEVICE_MODEL:                "Platform/Device",
	UP_DEVICE_TYPE:                 "Platform/Device",
	UP_DEVICE_FAMILY:               "Platform/Device",
	UP_DEVICE_MANUFACTURER:         "Platform/Device",
	UP_DEVICE_CARRIER:              "Platform/Device",
	UP_COUNTRY:                     "User identification",
	UP_CITY:                        "User identification",
	UP_REGION:                      "User identification",
	UP_TIMEZONE:                    "User identification",
	UP_USER_ID:                     "User identification",
	UP_EMAIL:                       "User identification",
	UP_COMPANY:                     "Company identification",
	UP_NAME:                        "User identification",
	UP_FIRST_NAME:                  "User identification",
	UP_LAST_NAME:                   "User identification",
	UP_PHONE:                       "User identification",
	UP_INITIAL_PAGE_URL:            "Page properties",
	UP_INITIAL_PAGE_DOMAIN:         "Page properties",
	UP_INITIAL_PAGE_RAW_URL:        "Page properties",
	UP_INITIAL_PAGE_LOAD_TIME:      "Page properties",
	UP_INITIAL_PAGE_SPENT_TIME:     "Page properties",
	UP_INITIAL_PAGE_SCROLL_PERCENT: "Page properties",
	UP_INITIAL_CAMPAIGN:            "Traffic source",
	UP_INITIAL_CAMPAIGN_ID:         "Traffic source",
	UP_INITIAL_SOURCE:              "Traffic source",
	UP_INITIAL_MEDIUM:              "Traffic source",
	UP_INITIAL_KEYWORD:             "Traffic source",
	UP_INITIAL_KEYWORD_MATCH_TYPE:  "Traffic source",
	UP_INITIAL_TERM:                "Traffic source",
	UP_INITIAL_CONTENT:             "Traffic source",
	UP_INITIAL_ADGROUP:             "Traffic source",
	UP_INITIAL_ADGROUP_ID:          "Traffic source",
	UP_INITIAL_CREATIVE:            "Traffic source",
	UP_INITIAL_GCLID:               "Traffic source",
	UP_INITIAL_FBCLID:              "Traffic source",
	UP_INITIAL_REFERRER:            "Traffic source",
	UP_INITIAL_REFERRER_URL:        "Traffic source",
	UP_INITIAL_REFERRER_DOMAIN:     "Traffic source",
	UP_INITIAL_CHANNEL:             "Traffic source",
	UP_DAY_OF_FIRST_EVENT:          "User identification",
	UP_HOUR_OF_FIRST_EVENT:         "User identification",
	UP_LATEST_PAGE_URL:             "Page properties",
	UP_LATEST_PAGE_DOMAIN:          "Page properties",
	UP_LATEST_PAGE_RAW_URL:         "Page properties",
	UP_LATEST_PAGE_LOAD_TIME:       "Page properties",
	UP_LATEST_PAGE_SPENT_TIME:      "Page properties",
	UP_LATEST_PAGE_SCROLL_PERCENT:  "Page properties",
	UP_LATEST_CAMPAIGN:             "Traffic source",
	UP_LATEST_CAMPAIGN_ID:          "Traffic source",
	UP_LATEST_SOURCE:               "Traffic source",
	UP_LATEST_MEDIUM:               "Traffic source",
	UP_LATEST_KEYWORD:              "Traffic source",
	UP_LATEST_KEYWORD_MATCH_TYPE:   "Traffic source",
	UP_LATEST_TERM:                 "Traffic source",
	UP_LATEST_CONTENT:              "Traffic source",
	UP_LATEST_ADGROUP:              "Traffic source",
	UP_LATEST_ADGROUP_ID:           "Traffic source",
	UP_LATEST_CREATIVE:             "Traffic source",
	UP_LATEST_GCLID:                "Traffic source",
	UP_LATEST_FBCLID:               "Traffic source",
	UP_LATEST_REFERRER:             "Traffic source",
	UP_LATEST_REFERRER_URL:         "Traffic source",
	UP_LATEST_REFERRER_DOMAIN:      "Traffic source",
	UP_LATEST_CHANNEL:              "Traffic source",
	UP_JOIN_TIME:                   "User identification",
	UP_POSTAL_CODE:                 "User identification",
	UP_CONTINENT:                   "User identification",
	EP_HOUR_OF_DAY:                 "User identification",
	EP_DAY_OF_WEEK:                 "User identification",
	SP_SESSION_TIME:                "Session properties",
	SP_SPENT_TIME:                  "Session properties",
	SP_PAGE_COUNT:                  "Session properties",
	SIX_SIGNAL_ADDRESS:             "Company identification",
	SIX_SIGNAL_ANNUAL_REVENUE:      "Company identification",
	SIX_SIGNAL_CITY:                "Company identification",
	SIX_SIGNAL_COUNTRY:             "Company identification",
	SIX_SIGNAL_COUNTRY_ISO_CODE:    "Company identification",
	SIX_SIGNAL_DOMAIN:              "Company identification",
	SIX_SIGNAL_EMPLOYEE_COUNT:      "Company identification",
	SIX_SIGNAL_EMPLOYEE_RANGE:      "Company identification",
	SIX_SIGNAL_INDUSTRY:            "Company identification",
	SIX_SIGNAL_NAICS:               "Company identification",
	SIX_SIGNAL_NAICS_DESCRIPTION:   "Company identification",
	SIX_SIGNAL_NAME:                "Company identification",
	SIX_SIGNAL_PHONE:               "Company identification",
	SIX_SIGNAL_REGION:              "Company identification",
	SIX_SIGNAL_REVENUE_RANGE:       "Company identification",
	SIX_SIGNAL_SIC:                 "Company identification",
	SIX_SIGNAL_SIC_DESCRIPTION:     "Company identification",
	SIX_SIGNAL_STATE:               "Company identification",
	SIX_SIGNAL_ZIP:                 "Company identification",
	G2_DOMAIN:                      "G2 Properties",
	G2_NAME:                        "G2 Properties",
	G2_LEGAL_NAME:                  "G2 Properties",
	G2_COUNTRY:                     "G2 Properties",
	G2_EMPLOYEES_RANGE:             "G2 Properties",
	G2_EMPLOYEES:                   "G2 Properties",
	G2_COMPANY_ID:                  "G2 Properties",
}

var DISABLED_EVENT_USER_PROPERTIES = []string{
	UP_SESSION_COUNT,
	UP_PAGE_COUNT,
	UP_TOTAL_SPENT_TIME,
	IDENTIFIED_USER_ID,
}

var STANDARD_SESSION_PROPERTIES_CATAGORIZATION = map[string]string{
	SP_IS_FIRST_SESSION:            "Session properties",
	SP_SESSION_TIME:                "Session properties",
	SP_INITIAL_REFERRER:            "Session properties",
	SP_INITIAL_REFERRER_URL:        "Session properties",
	SP_INITIAL_REFERRER_DOMAIN:     "Session properties",
	SP_SPENT_TIME:                  "Session properties",
	SP_PAGE_COUNT:                  "Session properties",
	SP_LATEST_PAGE_URL:             "Session properties",
	SP_LATEST_PAGE_RAW_URL:         "Session properties",
	UP_INITIAL_PAGE_URL:            "Session properties",
	UP_INITIAL_PAGE_RAW_URL:        "Session properties",
	UP_INITIAL_PAGE_DOMAIN:         "Session properties",
	UP_INITIAL_PAGE_LOAD_TIME:      "Session properties",
	UP_INITIAL_PAGE_SPENT_TIME:     "Session properties",
	UP_INITIAL_PAGE_SCROLL_PERCENT: "Session properties",
	UP_PLATFORM:                    "Session properties",
	UP_BROWSER:                     "Session properties",
	UP_BROWSER_VERSION:             "Session properties",
	UP_OS:                          "Session properties",
	UP_OS_VERSION:                  "Session properties",
	UP_COUNTRY:                     "Session properties",
	UP_CITY:                        "Session properties",
	UP_REGION:                      "Session properties",
	UP_TIMEZONE:                    "Session properties",
	UP_POSTAL_CODE:                 "Session properties",
	UP_CONTINENT:                   "Session properties",
	EP_CAMPAIGN:                    "Session properties",
	EP_CAMPAIGN_ID:                 "Session properties",
	EP_SOURCE:                      "Session properties",
	EP_MEDIUM:                      "Session properties",
	EP_KEYWORD:                     "Session properties",
	EP_KEYWORD_MATCH_TYPE:          "Session properties",
	EP_TERM:                        "Session properties",
	EP_CONTENT:                     "Session properties",
	EP_ADGROUP:                     "Session properties",
	EP_ADGROUP_ID:                  "Session properties",
	EP_CREATIVE:                    "Session properties",
	EP_GCLID:                       "Session properties",
	EP_FBCLID:                      "Session properties",
}

var STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES = map[string]string{
	SP_IS_FIRST_SESSION:            "Is first session",
	SP_SESSION_TIME:                "Session start time",
	SP_INITIAL_REFERRER:            "Session referrer",
	SP_INITIAL_REFERRER_URL:        "Session referrer URL",
	SP_INITIAL_REFERRER_DOMAIN:     "Session referrer domain",
	SP_SPENT_TIME:                  "Session active time",
	SP_PAGE_COUNT:                  "Session page count",
	SP_LATEST_PAGE_URL:             "Session exit page URL",
	SP_LATEST_PAGE_RAW_URL:         "Session exit page raw URL",
	UP_INITIAL_PAGE_URL:            "Session landing page URL",
	UP_INITIAL_PAGE_RAW_URL:        "Session landing page raw URL",
	UP_INITIAL_PAGE_DOMAIN:         "Session landing page domain",
	UP_INITIAL_PAGE_LOAD_TIME:      "Session landing page load time",
	UP_INITIAL_PAGE_SPENT_TIME:     "Session landing page active time",
	UP_INITIAL_PAGE_SCROLL_PERCENT: "Session landing page scroll percent",
	UP_PLATFORM:                    "Session platform",
	UP_BROWSER:                     "Session browser",
	UP_BROWSER_VERSION:             "Session browser version",
	UP_OS:                          "Session OS",
	UP_OS_VERSION:                  "Session OS version",
	UP_COUNTRY:                     "Session country",
	UP_CITY:                        "Session city",
	UP_REGION:                      "Session region",
	UP_TIMEZONE:                    "Session timezone",
	UP_POSTAL_CODE:                 "Session postal code",
	UP_CONTINENT:                   "Session continent",
	EP_CAMPAIGN:                    "Session campaign",
	EP_CAMPAIGN_ID:                 "Session campaign ID",
	EP_SOURCE:                      "Session source",
	EP_MEDIUM:                      "Session medium",
	EP_KEYWORD:                     "Session keyword",
	EP_KEYWORD_MATCH_TYPE:          "Session keyword match type",
	EP_TERM:                        "Session search term",
	EP_CONTENT:                     "Session content",
	EP_ADGROUP:                     "Session adgroup",
	EP_ADGROUP_ID:                  "Session adgroup ID",
	EP_CREATIVE:                    "Session creative",
	EP_GCLID:                       "Session GCLID",
	EP_FBCLID:                      "Session FBCLID",
}

var CHANNEL_PROPERTIES_DISPLAY_NAMES = map[string]string{
	"$initial_referrer_domain": "Referrer Domain",
	"$campaign":                "Campaign",
	"$source":                  "Source",
	"$medium":                  "Medium",
	"$gclid":                   "GCLID",
	"$fbclid":                  "FBCLID",
}

var PAGE_VIEWS_STANDARD_PROPERTIES_CATEGORICAL = []string{
	EP_CAMPAIGN,
	EP_CAMPAIGN_ID,
	EP_SOURCE,
	EP_MEDIUM,
	EP_KEYWORD,
	EP_KEYWORD_MATCH_TYPE,
	EP_TERM,
	EP_CONTENT,
	EP_ADGROUP,
	EP_ADGROUP_ID,
	EP_AD,
	EP_AD_ID,
	EP_CREATIVE,
	EP_GCLID,
	EP_FBCLID,
	EP_PAGE_TITLE,
	EP_PAGE_DOMAIN,
	EP_PAGE_RAW_URL,
	EP_PAGE_URL,
	EP_REFERRER,
	EP_REFERRER_DOMAIN,
	EP_REFERRER_URL,
}

var BUTTON_CLICKS_STANDARD_PROPERTIES_CATEGORICAL = []string{
	EP_PAGE_TITLE,
	EP_PAGE_DOMAIN,
	EP_PAGE_RAW_URL,
	EP_PAGE_URL,
	EP_REFERRER,
	EP_REFERRER_DOMAIN,
	EP_REFERRER_URL,
	EP_CLICK_ELEMENT_TYPE,
	EP_CLICK_CLASS,
	EP_CLICK_ID,
	EP_CLICK_REL,
	EP_CLICK_ROLE,
	EP_CLICK_TARGET,
	EP_CLICK_HREF,
	EP_CLICK_MEDIA,
	EP_CLICK_TYPE,
	EP_CLICK_NAME,
}

var PAGE_VIEWS_STANDARD_PROPERTIES_NUMERICAL = []string{
	EP_PAGE_LOAD_TIME,
	EP_PAGE_SPENT_TIME,
	EP_PAGE_SCROLL_PERCENT,
}

// USER_PROPERTIES_MERGE_TYPE_INITIAL Properties for which preference will be given to first occurrence while merging.
// For rest all properties, latest user values will prevail on conflict.
var USER_PROPERTIES_MERGE_TYPE_INITIAL = [...]string{
	UP_DAY_OF_FIRST_EVENT,
	UP_HOUR_OF_FIRST_EVENT,
	UP_INITIAL_ADGROUP,
	UP_INITIAL_ADGROUP_ID,
	UP_INITIAL_CAMPAIGN,
	UP_INITIAL_CAMPAIGN_ID,
	UP_INITIAL_CONTENT,
	UP_INITIAL_CREATIVE,
	UP_INITIAL_FBCLID,
	UP_INITIAL_GCLID,
	UP_INITIAL_KEYWORD,
	UP_INITIAL_KEYWORD_MATCH_TYPE,
	UP_INITIAL_TERM,
	UP_INITIAL_MEDIUM,
	UP_INITIAL_PAGE_DOMAIN,
	UP_INITIAL_PAGE_LOAD_TIME,
	UP_INITIAL_PAGE_RAW_URL,
	UP_INITIAL_PAGE_SCROLL_PERCENT,
	UP_INITIAL_PAGE_SPENT_TIME,
	UP_INITIAL_PAGE_URL,
	UP_INITIAL_REFERRER,
	UP_INITIAL_REFERRER_DOMAIN,
	UP_INITIAL_REFERRER_URL,
	UP_INITIAL_SOURCE,
	UP_INITIAL_CHANNEL,
	UP_JOIN_TIME,
}

var USER_PROPERTIES_MERGE_TYPE_ADD = [...]string{
	UP_PAGE_COUNT,
	UP_TOTAL_SPENT_TIME,
}
var CUSTOM_WHITELIST_DELTA = []string{
	"$referrer",
	"$page_url",
	"$source",
	"$campaign",
	"$channel",
}

var CUSTOM_BLACKLIST_DELTA = []string{
	"$latest_referrer",
	"$latest_referrer_url",
	"$initial_referrer",
	"$initial_referrer_url",
	"$referrer",
	"$referrer_url",
	"$latest_page_url",
	"$latest_page_domain",
	"$latest_page_raw_url",
	"$latest_page_load_time",
	"$latest_page_spent_time",
	"$latest_page_scroll_percent",
	"$ip",
	"$timestamp",
	"$session_latest_page_url",
	"$browser_version",
	"$browser_with_version",
	"$day_of_first_event",
	"$device_brand",
	"$device_model",
	"$email",
	"$first_name",
	"$hour_of_first_event",
	"$hubspot_company_address",
	"$hubspot_company_address2",
	"$hubspot_company_closedate",
	"$hubspot_company_createdate",
	"$hubspot_company_current_plan_test",
	"$hubspot_company_days_to_close",
	"$hubspot_company_description",
	"$hubspot_company_domain",
	"$hubspot_company_engagements_last_meeting_booked",
	"$hubspot_company_facebook_company_page",
	"$hubspot_company_first_contact_createdate",
	"$hubspot_company_first_contact_createdate_timestamp_earliest_value_78b50eea",
	"$hubspot_company_first_conversion_date",
	"$hubspot_company_first_conversion_date_timestamp_earliest_value_61f58f2c",
	"$hubspot_company_first_conversion_event_name",
	"$hubspot_company_first_conversion_event_name_timestamp_earliest_value_68ddae0a",
	"$hubspot_company_first_deal_created_date",
	"$hubspot_company_hs_additional_domains",
	"$hubspot_company_hs_all_accessible_team_ids",
	"$hubspot_company_hs_all_owner_ids",
	"$hubspot_company_hs_all_team_ids",
	"$hubspot_company_hs_analytics_first_timestamp",
	"$hubspot_company_hs_analytics_first_visit_timestamp",
	"$hubspot_company_hs_analytics_last_timestamp",
	"$hubspot_company_hs_analytics_last_visit_timestamp",
	"$hubspot_company_hs_analytics_num_page_views",
	"$hubspot_company_hs_analytics_num_visits",
	"$hubspot_company_hs_created_by_user_id",
	"$hubspot_company_hs_last_booked_meeting_date",
	"$hubspot_company_hs_last_logged_call_date",
	"$hubspot_company_hs_last_open_task_date",
	"$hubspot_company_hs_last_sales_activity_date",
	"$hubspot_company_hs_last_sales_activity_timestamp",
	"$hubspot_company_hs_lastmodifieddate",
	"$hubspot_company_hs_latest_meeting_activity",
	"$hubspot_company_hs_merged_object_ids",
	"$hubspot_company_hs_num_blockers",
	"$hubspot_company_hs_num_child_companies",
	"$hubspot_company_hs_num_open_deals",
	"$hubspot_company_hs_object_id",
	"$hubspot_company_hs_predictivecontactscore_v2",
	"$hubspot_company_hs_sales_email_last_replied",
	"$hubspot_company_hs_target_account_probability",
	"$hubspot_company_hs_total_deal_value",
	"$hubspot_company_hs_updated_by_user_id",
	"$hubspot_company_hs_user_ids_of_all_owners",
	"$hubspot_company_hubspot_owner_assigneddate",
	"$hubspot_company_hubspot_owner_id",
	"$hubspot_company_hubspot_team_id",
	"$hubspot_company_lfapp_latest_visit",
	"$hubspot_company_lfapp_view_in_leadfeeder",
	"$hubspot_company_lifecyclestage",
	"$hubspot_company_linkedin_company_page",
	"$hubspot_company_linkedinbio",
	"$hubspot_company_name",
	"$hubspot_company_notes_last_contacted",
	"$hubspot_company_notes_last_updated",
	"$hubspot_company_notes_next_activity_date",
	"$hubspot_company_num_associated_deals",
	"$hubspot_company_num_contacted_notes",
	"$hubspot_company_num_conversion_events",
	"$hubspot_company_num_notes",
	"$hubspot_company_phone",
	"$hubspot_company_recent_conversion_date",
	"$hubspot_company_recent_conversion_date_timestamp_latest_value_72856da1",
	"$hubspot_company_recent_conversion_event_name",
	"$hubspot_company_recent_conversion_event_name_timestamp_latest_value_66c820bf",
	"$hubspot_company_recent_deal_amount",
	"$hubspot_company_recent_deal_close_date",
	"$hubspot_company_rollworks_clicks",
	"$hubspot_company_rollworks_conversions",
	"$hubspot_company_rollworks_ctc",
	"$hubspot_company_rollworks_impression_cost",
	"$hubspot_company_rollworks_impressions",
	"$hubspot_company_rollworks_last_activity",
	"$hubspot_company_rollworks_page_views",
	"$hubspot_company_rollworks_vtc",
	"$hubspot_company_timezone",
	"$hubspot_company_twitterhandle",
	"$hubspot_company_web_technologies",
	"$hubspot_company_website",
	"$hubspot_company_zip",
	"$hubspot_contact_additional_emails",
	"$hubspot_contact_address",
	"$hubspot_contact_aircall_last_call_at",
	"$hubspot_contact_asset",
	"$hubspot_contact_assigned_sdr",
	"$hubspot_contact_associatedcompanyid",
	"$hubspot_contact_associatedcompanylastupdated",
	"$hubspot_contact_chat_website",
	"$hubspot_contact_closedate",
	"$hubspot_contact_company",
	"$hubspot_contact_company_name_hiver",
	"$hubspot_contact_country_code",
	"$hubspot_contact_createdate",
	"$hubspot_contact_csm_owner",
	"$hubspot_contact_currentlyinworkflow",
	"$hubspot_contact_days_to_close",
	"$hubspot_contact_document_title",
	"$hubspot_contact_drip_owner",
	"$hubspot_contact_dropoff_email_context",
	"$hubspot_contact_email",
	"$hubspot_contact_email_overload_revenue_calculator_value",
	"$hubspot_contact_engagements_last_meeting_booked",
	"$hubspot_contact_engagements_last_meeting_booked_campaign",
	"$hubspot_contact_engagements_last_meeting_booked_medium",
	"$hubspot_contact_engagements_last_meeting_booked_source",
	"$hubspot_contact_facebook_form_id",
	"$hubspot_contact_favourite_song",
	"$hubspot_contact_fax",
	"$hubspot_contact_first_conversion_date",
	"$hubspot_contact_first_deal_created_date",
	"$hubspot_contact_firstname",
	"$hubspot_contact_followercount",
	"$hubspot_contact_full_name",
	"$hubspot_contact_get_a_demo_of_hiver_",
	"$hubspot_contact_google_apps_check_with_builtwith",
	"$hubspot_contact_grexit_admin_link",
	"$hubspot_contact_hs_additional_emails",
	"$hubspot_contact_hs_all_accessible_team_ids",
	"$hubspot_contact_hs_all_assigned_business_unit_ids",
	"$hubspot_contact_hs_all_contact_vids",
	"$hubspot_contact_hs_all_owner_ids",
	"$hubspot_contact_hs_all_team_ids",
	"$hubspot_contact_hs_analytics_average_page_views",
	"$hubspot_contact_hs_analytics_first_timestamp",
	"$hubspot_contact_hs_analytics_first_touch_converting_campaign",
	"$hubspot_contact_hs_analytics_first_url",
	"$hubspot_contact_hs_analytics_first_visit_timestamp",
	"$hubspot_contact_hs_analytics_last_referrer",
	"$hubspot_contact_hs_analytics_last_timestamp",
	"$hubspot_contact_hs_analytics_last_touch_converting_campaign",
	"$hubspot_contact_hs_analytics_last_url",
	"$hubspot_contact_hs_analytics_last_visit_timestamp",
	"$hubspot_contact_hs_analytics_num_page_views",
	"$hubspot_contact_hs_analytics_num_visits",
	"$hubspot_contact_hs_analytics_source",
	"$hubspot_contact_hs_analytics_source_data_1",
	"$hubspot_contact_hs_content_membership_notes",
	"$hubspot_contact_hs_count_is_unworked",
	"$hubspot_contact_hs_count_is_worked",
	"$hubspot_contact_hs_created_by_conversations",
	"$hubspot_contact_hs_document_last_revisited",
	"$hubspot_contact_hs_email_bad_address",
	"$hubspot_contact_hs_email_bounce",
	"$hubspot_contact_hs_email_click",
	"$hubspot_contact_hs_email_delivered",
	"$hubspot_contact_hs_email_domain",
	"$hubspot_contact_hs_email_first_click_date",
	"$hubspot_contact_hs_email_first_open_date",
	"$hubspot_contact_hs_email_first_reply_date",
	"$hubspot_contact_hs_email_first_send_date",
	"$hubspot_contact_hs_email_hard_bounce_reason_enum",
	"$hubspot_contact_hs_email_last_click_date",
	"$hubspot_contact_hs_email_last_email_name",
	"$hubspot_contact_hs_email_last_open_date",
	"$hubspot_contact_hs_email_last_reply_date",
	"$hubspot_contact_hs_email_last_send_date",
	"$hubspot_contact_hs_email_open",
	"$hubspot_contact_hs_email_optout",
	"$hubspot_contact_hs_email_optout_11021605",
	"$hubspot_contact_hs_email_optout_4647003",
	"$hubspot_contact_hs_email_optout_5302455",
	"$hubspot_contact_hs_email_optout_5302456",
	"$hubspot_contact_hs_email_optout_5354102",
	"$hubspot_contact_hs_email_optout_5596517",
	"$hubspot_contact_hs_email_optout_5596768",
	"$hubspot_contact_hs_email_optout_5845738",
	"$hubspot_contact_hs_email_replied",
	"$hubspot_contact_hs_email_sends_since_last_engagement",
	"$hubspot_contact_hs_facebook_ad_clicked",
	"$hubspot_contact_hs_facebook_click_id",
	"$hubspot_contact_hs_facebookid",
	"$hubspot_contact_hs_first_engagement_object_id",
	"$hubspot_contact_hs_google_click_id",
	"$hubspot_contact_hs_ip_timezone",
	"$hubspot_contact_hs_is_contact",
	"$hubspot_contact_hs_is_unworked",
	"$hubspot_contact_hs_last_sales_activity_date",
	"$hubspot_contact_hs_last_sales_activity_timestamp",
	"$hubspot_contact_hs_latest_meeting_activity",
	"$hubspot_contact_hs_latest_sequence_ended_date",
	"$hubspot_contact_hs_latest_sequence_enrolled",
	"$hubspot_contact_hs_latest_sequence_enrolled_date",
	"$hubspot_contact_hs_latest_sequence_finished_date",
	"$hubspot_contact_hs_latest_sequence_unenrolled_date",
	"$hubspot_contact_hs_legal_basis",
	"$hubspot_contact_hs_lifecyclestage_customer_date",
	"$hubspot_contact_hs_lifecyclestage_lead_date",
	"$hubspot_contact_hs_lifecyclestage_marketingqualifiedlead_date",
	"$hubspot_contact_hs_lifecyclestage_opportunity_date",
	"$hubspot_contact_hs_lifecyclestage_other_date",
	"$hubspot_contact_hs_lifecyclestage_salesqualifiedlead_date",
	"$hubspot_contact_hs_lifecyclestage_subscriber_date",
	"$hubspot_contact_hs_marketable_reason_type",
	"$hubspot_contact_hs_marketable_status",
	"$hubspot_contact_hs_marketable_until_renewal",
	"$hubspot_contact_hs_object_id",
	"$hubspot_contact_hs_predictivecontactscore_v2",
	"$hubspot_contact_hs_predictivescoringtier",
	"$hubspot_contact_hs_sa_first_engagement_date",
	"$hubspot_contact_hs_sa_first_engagement_descr",
	"$hubspot_contact_hs_sa_first_engagement_object_type",
	"$hubspot_contact_hs_sales_email_last_clicked",
	"$hubspot_contact_hs_sales_email_last_opened",
	"$hubspot_contact_hs_sales_email_last_replied",
	"$hubspot_contact_hs_searchable_calculated_mobile_number",
	"$hubspot_contact_hs_searchable_calculated_phone_number",
	"$hubspot_contact_hs_sequences_enrolled_count",
	"$hubspot_contact_hs_sequences_is_enrolled",
	"$hubspot_contact_hs_social_facebook_clicks",
	"$hubspot_contact_hs_social_google_plus_clicks",
	"$hubspot_contact_hs_social_last_engagement",
	"$hubspot_contact_hs_social_linkedin_clicks",
	"$hubspot_contact_hs_social_num_broadcast_clicks",
	"$hubspot_contact_hs_social_twitter_clicks",
	"$hubspot_contact_hs_time_between_contact_creation_and_deal_close",
	"$hubspot_contact_hs_time_between_contact_creation_and_deal_creation",
	"$hubspot_contact_hs_time_to_first_engagement",
	"$hubspot_contact_hs_time_to_move_from_lead_to_customer",
	"$hubspot_contact_hs_time_to_move_from_marketingqualifiedlead_to_customer",
	"$hubspot_contact_hs_time_to_move_from_opportunity_to_customer",
	"$hubspot_contact_hs_time_to_move_from_salesqualifiedlead_to_customer",
	"$hubspot_contact_hs_time_to_move_from_subscriber_to_customer",
	"$hubspot_contact_hs_twitterid",
	"$hubspot_contact_hs_user_ids_of_all_owners",
	"$hubspot_contact_hubspot_owner_assigneddate",
	"$hubspot_contact_hubspot_owner_id",
	"$hubspot_contact_hubspot_team_id",
	"$hubspot_contact_hubspotscore",
	"$hubspot_contact_internal_notes",
	"$hubspot_contact_invited_users",
	"$hubspot_contact_ip_city",
	"$hubspot_contact_ip_country",
	"$hubspot_contact_ip_country_code",
	"$hubspot_contact_ip_state",
	"$hubspot_contact_ip_state_code",
	"$hubspot_contact_is_churned_customer",
	"$hubspot_contact_is_dnd",
	"$hubspot_contact_kloutscoregeneral",
	"$hubspot_contact_last_used_aircall_phone_number",
	"$hubspot_contact_lastmodifieddate",
	"$hubspot_contact_lastname",
	"$hubspot_contact_lead_guid",
	"$hubspot_contact_linkedin_consent_checkbox_i_allow_hiver_to_send_me_super_valuable_content_which_i_may_opt_out_from_",
	"$hubspot_contact_linkedin_profile_link",
	"$hubspot_contact_linkedinbio",
	"$hubspot_contact_linkedinconnections",
	"$hubspot_contact_marketing_funnel",
	"$hubspot_contact_message",
	"$hubspot_contact_mobilephone",
	"$hubspot_contact_notes_last_contacted",
	"$hubspot_contact_notes_last_updated",
	"$hubspot_contact_notes_next_activity_date",
	"$hubspot_contact_num_associated_deals",
	"$hubspot_contact_num_contacted_notes",
	"$hubspot_contact_num_notes",
	"$hubspot_contact_oauth_done",
	"$hubspot_contact_om_campaign_name",
	"$hubspot_contact_original_source_internal_use_events",
	"$hubspot_contact_outbound_sdr_cadence",
	"$hubspot_contact_page",
	"$hubspot_contact_partner_attached",
	"$hubspot_contact_ph_country",
	"$hubspot_contact_phone",
	"$hubspot_contact_phone_no",
	"$hubspot_contact_photo",
	"$hubspot_contact_quiz_name",
	"$hubspot_contact_recent_conversion_date",
	"$hubspot_contact_recent_deal_amount",
	"$hubspot_contact_recent_deal_close_date",
	"$hubspot_contact_sdr_qualified",
	"$hubspot_contact_signup_form_submitted",
	"$hubspot_contact_sm_created",
	"$hubspot_contact_sm_validated",
	"$hubspot_contact_subscription_for_marketing_mails",
	"$hubspot_contact_t_shirt_size",
	"$hubspot_contact_tag_id",
	"$hubspot_contact_tags",
	"$hubspot_contact_total_revenue",
	"$hubspot_contact_trial",
	"$hubspot_contact_trial_signup_source",
	"$hubspot_contact_trial_up_source",
	"$hubspot_contact_twitterbio",
	"$hubspot_contact_twitterhandle",
	"$hubspot_contact_twitterprofilephoto",
	"$hubspot_contact_use_case",
	"$hubspot_contact_usergroup",
	"$hubspot_contact_want_a_demo_of_hiver",
	"$hubspot_contact_webinareventlastupdated",
	"$hubspot_contact_website",
	"$hubspot_contact_work_email",
	"$hubspot_contact_zip",
	"$hubspot_contact_zoom_webinar_joinlink",
	"$hubspot_contact_zoom_webinar_registration_count",
	"$hubspot_deal_amount",
	"$hubspot_deal_amount_in_home_currency",
	"$hubspot_deal_closed_lost_reason",
	"$hubspot_deal_closedate",
	"$hubspot_deal_country",
	"$hubspot_deal_createdate",
	"$hubspot_deal_days_to_close",
	"$hubspot_deal_dealname",
	"$hubspot_deal_dealstage",
	"$hubspot_deal_dealtype",
	"$hubspot_deal_engagements_last_meeting_booked",
	"$hubspot_deal_forecasted_deal_amount",
	"$hubspot_deal_hs_all_accessible_team_ids",
	"$hubspot_deal_hs_all_owner_ids",
	"$hubspot_deal_hs_all_team_ids",
	"$hubspot_deal_hs_closed_amount",
	"$hubspot_deal_hs_closed_amount_in_home_currency",
	"$hubspot_deal_hs_created_by_user_id",
	"$hubspot_deal_hs_createdate",
	"$hubspot_deal_hs_date_entered_closedlost",
	"$hubspot_deal_hs_date_entered_closedwon",
	"$hubspot_deal_hs_date_entered_contractsent",
	"$hubspot_deal_hs_date_entered_decisionmakerboughtin",
	"$hubspot_deal_hs_date_entered_f41f27a4_791a_49ff_8a97_0f10745d660f_2143803989",
	"$hubspot_deal_hs_date_entered_presentationscheduled",
	"$hubspot_deal_hs_date_entered_qualifiedtobuy",
	"$hubspot_deal_hs_date_exited_contractsent",
	"$hubspot_deal_hs_date_exited_decisionmakerboughtin",
	"$hubspot_deal_hs_date_exited_f41f27a4_791a_49ff_8a97_0f10745d660f_2143803989",
	"$hubspot_deal_hs_date_exited_presentationscheduled",
	"$hubspot_deal_hs_date_exited_qualifiedtobuy",
	"$hubspot_deal_hs_deal_stage_probability",
	"$hubspot_deal_hs_deal_stage_probability_shadow",
	"$hubspot_deal_hs_forecast_amount",
	"$hubspot_deal_hs_forecast_probability",
	"$hubspot_deal_hs_is_closed",
	"$hubspot_deal_hs_is_closed_won",
	"$hubspot_deal_hs_lastmodifieddate",
	"$hubspot_deal_hs_latest_meeting_activity",
	"$hubspot_deal_hs_manual_forecast_category",
	"$hubspot_deal_hs_num_associated_deal_splits",
	"$hubspot_deal_hs_num_target_accounts",
	"$hubspot_deal_hs_object_id",
	"$hubspot_deal_hs_projected_amount",
	"$hubspot_deal_hs_projected_amount_in_home_currency",
	"$hubspot_deal_hs_sales_email_last_replied",
	"$hubspot_deal_hs_time_in_closedlost",
	"$hubspot_deal_hs_time_in_closedwon",
	"$hubspot_deal_hs_time_in_contractsent",
	"$hubspot_deal_hs_time_in_decisionmakerboughtin",
	"$hubspot_deal_hs_time_in_f41f27a4_791a_49ff_8a97_0f10745d660f_2143803989",
	"$hubspot_deal_hs_time_in_presentationscheduled",
	"$hubspot_deal_hs_time_in_qualifiedtobuy",
	"$hubspot_deal_hs_updated_by_user_id",
	"$hubspot_deal_hs_user_ids_of_all_owners",
	"$hubspot_deal_hubspot_owner_assigneddate",
	"$hubspot_deal_hubspot_owner_id",
	"$hubspot_deal_hubspot_team_id",
	"$hubspot_deal_notes_last_contacted",
	"$hubspot_deal_notes_last_updated",
	"$hubspot_deal_notes_next_activity_date",
	"$hubspot_deal_num_associated_contacts",
	"$hubspot_deal_num_contacted_notes",
	"$hubspot_deal_num_notes",
	"$hubspot_deal_partner_attached",
	"$hubspot_deal_pipeline",
	"$hubspot_deal_source",
	"$identifiers",
	"$initial_content",
	"$initial_fbclid",
	"$initial_gclid",
	"$initial_page_url",
	"$initial_referrer",
	"$initial_source",
	"$merge_timestamp",
	"$name",
	"$os_version",
	"$os_with_version",
	"$phone",
	"$screen_height",
	"$screen_width",
	"$user_agent",
	"$user_id",
	"$hubspot_company_account_rating",
	"$hubspot_company_active_account_size",
	"$hubspot_company_active_customer_",
	"$hubspot_company_adpushup_country",
	"$hubspot_company_adpushup_products",
	"$hubspot_company_adpushup_region",
	"$hubspot_company_ads_dot_txt_line_count",
	"$hubspot_company_ads_txt_file",
	"$hubspot_company_adsense_on_page",
	"$hubspot_company_adsense_pub_id",
	"$hubspot_company_alexa_rank_in_2016",
	"$hubspot_company_amp_additional_detail",
	"$hubspot_company_amp_component_type",
	"$hubspot_company_amp_found",
	"$hubspot_company_amp_page_url",
	"$hubspot_company_amp_score",
	"$hubspot_company_ap_adx_spm_id",
	"$hubspot_company_average_visit_duration",
	"$hubspot_company_bounce_rate",
	"$hubspot_company_click_through_rate",
	"$hubspot_company_company_gross_monthly_revenue",
	"$hubspot_company_company_type",
	"$hubspot_company_competitor_live",
	"$hubspot_company_competitor_script_live",
	"$hubspot_company_competitor_script_live_page2",
	"$hubspot_company_competitor_script_page2_status",
	"$hubspot_company_competitor_script_previous_live",
	"$hubspot_company_competitor_script_status",
	"$hubspot_company_composite_score",
	"$hubspot_company_contract_done",
	"$hubspot_company_ctr_label",
	"$hubspot_company_desktop_share",
	"$hubspot_company_domain_registered_month",
	"$hubspot_company_domain_registered_on",
	"$hubspot_company_domain_registration_date",
	"$hubspot_company_exit_load",
	"$hubspot_company_expiry_date_of_contract",
	"$hubspot_company_frequency_of_customer_engagement",
	"$hubspot_company_g2_crowd_import",
	"$hubspot_company_google_represented_revenue",
	"$hubspot_company_hs_avatar_filemanager_key",
	"$hubspot_company_hs_ideal_customer_profile",
	"$hubspot_company_hs_parent_company_id",
	"$hubspot_company_iab_vertical",
	"$hubspot_company_last_claimed_active_account_size",
	"$hubspot_company_last_enriched_by",
	"$hubspot_company_last_enriched_on",
	"$hubspot_company_last_modified_by",
	"$hubspot_company_lost_customer",
	"$hubspot_company_market_research_associate",
	"$hubspot_company_mcm_name",
	"$hubspot_company_minimum_commitment_months_",
	"$hubspot_company_mobile_share",
	"$hubspot_company_monthly_ad_revenue",
	"$hubspot_company_monthly_visits",
	"$hubspot_company_moved_to_onboarding_date",
	"$hubspot_company_mra_assigned_date",
	"$hubspot_company_news_and_media_new_websites_for_mra",
	"$hubspot_company_news_and_media_pubs_india",
	"$hubspot_company_not_fit_for_enrichment",
	"$hubspot_company_notice_period_days_",
	"$hubspot_company_ops_scope_of_expansion",
	"$hubspot_company_page2_url",
	"$hubspot_company_page_speed_insights_score",
	"$hubspot_company_pages_per_visit",
	"$hubspot_company_payment_terms",
	"$hubspot_company_persona",
	"$hubspot_company_pre_deal_abm_stage",
	"$hubspot_company_primary_adpushup_website_id",
	"$hubspot_company_primary_login_email",
	"$hubspot_company_rank_composite_score",
	"$hubspot_company_revenue_share",
	"$hubspot_company_revenue_share_start_date",
	"$hubspot_company_richa_500",
	"$hubspot_company_richa_q3",
	"$hubspot_company_score_category_experience",
	"$hubspot_company_score_desktop_share",
	"$hubspot_company_score_monthly_visits",
	"$hubspot_company_score_pages_per_visit",
	"$hubspot_company_score_parent_category_experience",
	"$hubspot_company_score_visit_duration",
	"$hubspot_company_score_volatility",
	"$hubspot_company_secondary_websites",
	"$hubspot_company_secondary_websites_2",
	"$hubspot_company_similar_web_category",
	"$hubspot_company_similar_web_enrichment_flag",
	"$hubspot_company_similar_web_rank",
	"$hubspot_company_similar_web_website_category",
	"$hubspot_company_similar_web_website_rank",
	"$hubspot_company_similar_web_website_sub_category",
	"$hubspot_company_similarweb_page_views",
	"$hubspot_company_site_live_date",
	"$hubspot_company_sub_category_group",
	"$hubspot_company_temp_reverseadsensemapped",
	"$hubspot_company_time_on_site",
	"$hubspot_company_top_country_1",
	"$hubspot_company_top_country_1_share",
	"$hubspot_company_top_country_2",
	"$hubspot_company_top_country_2_share",
	"$hubspot_company_twitterbio",
	"$hubspot_company_twitterfollowers",
	"$hubspot_contact_active_customer_",
	"$hubspot_contact_ad_click_timestamp",
	"$hubspot_contact_adpushup_country",
	"$hubspot_contact_adpushup_email_validity",
	"$hubspot_contact_adpushup_persona",
	"$hubspot_contact_adpushup_products",
	"$hubspot_contact_adpushup_region",
	"$hubspot_contact_adsense_on_page",
	"$hubspot_contact_all_ad_interactions",
	"$hubspot_contact_ar_customer",
	"$hubspot_contact_ar_subscriber",
	"$hubspot_contact_average_visit_duration",
	"$hubspot_contact_bidder_live",
	"$hubspot_contact_bidders_not_live",
	"$hubspot_contact_blog_form_submission_url",
	"$hubspot_contact_blog_subscriber",
	"$hubspot_contact_bof_form_submission_url",
	"$hubspot_contact_changed_in_oct_crm_flow",
	"$hubspot_contact_chat_initiation_reason",
	"$hubspot_contact_clearbit_enrichment_flag",
	"$hubspot_contact_company_channel",
	"$hubspot_contact_company_channel_new",
	"$hubspot_contact_competitor_script_live",
	"$hubspot_contact_competitor_script_live_page2",
	"$hubspot_contact_competitor_script_page2_status",
	"$hubspot_contact_competitor_script_status",
	"$hubspot_contact_contact_type",
	"$hubspot_contact_contacted_age",
	"$hubspot_contact_customer_success_poc",
	"$hubspot_contact_date_of_birth",
	"$hubspot_contact_date_of_movement_into_demo_complete",
	"$hubspot_contact_deal_amount",
	"$hubspot_contact_demo_outcome",
	"$hubspot_contact_demo_reject_description",
	"$hubspot_contact_demo_reject_reason",
	"$hubspot_contact_diwali_gift_2020_delivery_date",
	"$hubspot_contact_diwali_gift_2020_dispatch_date",
	"$hubspot_contact_email_id_type",
	"$hubspot_contact_email_snippet",
	"$hubspot_contact_enrichment_type__latest_",
	"$hubspot_contact_event",
	"$hubspot_contact_event_subscriber",
	"$hubspot_contact_external_enrichment_vendor",
	"$hubspot_contact_finalized_for_nal_gifting",
	"$hubspot_contact_first_activity_timestamp",
	"$hubspot_contact_first_webinar_registration_date",
	"$hubspot_contact_gclid",
	"$hubspot_contact_gender",
	"$hubspot_contact_holiday_gift_2020_delivery_date",
	"$hubspot_contact_holiday_gifting_2020_dispatch_date",
	"$hubspot_contact_hs_avatar_filemanager_key",
	"$hubspot_contact_hs_content_membership_status",
	"$hubspot_contact_hs_email_optout_6775018",
	"$hubspot_contact_hs_email_optout_6775208",
	"$hubspot_contact_hs_email_optout_8283879",
	"$hubspot_contact_hs_language",
	"$hubspot_contact_hs_lifecyclestage_evangelist_date",
	"$hubspot_contact_hsa_ad",
	"$hubspot_contact_hsa_grp",
	"$hubspot_contact_hsa_kw",
	"$hubspot_contact_hsa_mt",
	"$hubspot_contact_hsa_tgt",
	"$hubspot_contact_hubspot_score_at_sql",
	"$hubspot_contact_hubspot_score_reached_100_date",
	"$hubspot_contact_hubspot_score_reached_120_date",
	"$hubspot_contact_hubspot_score_reached_20_date",
	"$hubspot_contact_hubspot_score_reached_40_date",
	"$hubspot_contact_hubspot_score_reached_60_date",
	"$hubspot_contact_hubspot_score_reached_80_date",
	"$hubspot_contact_hubspot_score_when_contact_reached_opportunity",
	"$hubspot_contact_hubspot_user_token",
	"$hubspot_contact_i_want_to_see_the_product_demo",
	"$hubspot_contact_inbound_form_submission_timestamp",
	"$hubspot_contact_inbound_owner_zapier_update",
	"$hubspot_contact_inbound_theme",
	"$hubspot_contact_included_in_nal_gift_list_q2_2020",
	"$hubspot_contact_inquiry_type",
	"$hubspot_contact_last_enriched_by",
	"$hubspot_contact_last_enriched_on",
	"$hubspot_contact_last_page_seen_list",
	"$hubspot_contact_lead_response_time_hours_",
	"$hubspot_contact_linkedin_profile_url",
	"$hubspot_contact_lp_first_level_email_opt_in",
	"$hubspot_contact_market_research_associate",
	"$hubspot_contact_marketing_campaign",
	"$hubspot_contact_monthly_ad_revenue",
	"$hubspot_contact_monthly_ad_revenue_hubspot_forms_",
	"$hubspot_contact_monthly_visits",
	"$hubspot_contact_mra_assigned_date",
	"$hubspot_contact_nal_gift_delivered_date",
	"$hubspot_contact_newsletter_subscriber",
	"$hubspot_contact_number_of_sales_emails_replied",
	"$hubspot_contact_outgrow_calculator_name",
	"$hubspot_contact_page2_url",
	"$hubspot_contact_pages_per_visit",
	"$hubspot_contact_paid_ads_subscriber",
	"$hubspot_contact_poptin_form",
	"$hubspot_contact_prospect_research_pointers",
	"$hubspot_contact_qualify_status",
	"$hubspot_contact_re_enrichment_required",
	"$hubspot_contact_relationship_status",
	"$hubspot_contact_resource_subscriber",
	"$hubspot_contact_richa_500",
	"$hubspot_contact_salutation",
	"$hubspot_contact_sdr",
	"$hubspot_contact_seo_demarcation",
	"$hubspot_contact_shortlisted_for_personalised_us_outreach",
	"$hubspot_contact_similar_web_category",
	"$hubspot_contact_similar_web_enrichment_flag",
	"$hubspot_contact_similar_web_pageviews",
	"$hubspot_contact_similar_web_rank",
	"$hubspot_contact_similar_web_website_sub_category",
	"$hubspot_contact_skype_id",
	"$hubspot_contact_sub_category_group",
	"$hubspot_contact_surveymonkeyeventlastupdated",
	"$hubspot_contact_tof_form_submission_url",
	"$hubspot_contact_twitter_handle",
	"$hubspot_contact_utm_campaign",
	"$hubspot_contact_utm_keyword",
	"$hubspot_contact_utm_medium",
	"$hubspot_contact_utm_source",
	"$hubspot_contact_utm_term",
	"$hubspot_contact_utsav",
	"$hubspot_contact_webinar_form_submission_url",
	"$hubspot_contact_webinar_registrations",
	"$hubspot_contact_webinar_subscriber",
	"$hubspot_contact_website_monthly_revenue_choose_closest",
	"$hubspot_contact_website_url_new_",
	"$hubspot_contact_zerobounce_email_status",
	"$hubspot_contact_zerobounce_email_sub_status",
	"$hubspot_contact_zerobounce_enrichment_flag",
	"$hubspot_contact_zoom_webinar_attendance_average_duration",
	"$hubspot_contact_zoom_webinar_attendance_count",
	"$hubspot_deal_adpushup_products",
	"$hubspot_deal_ads_txt",
	"$hubspot_deal_ads_txt_count",
	"$hubspot_deal_bofu_marketing_campaign",
	"$hubspot_deal_closed_lost_reason_jan2021_new",
	"$hubspot_deal_closed_won_reason",
	"$hubspot_deal_competitor_lost_to",
	"$hubspot_deal_competitor_pitted_against",
	"$hubspot_deal_deal_channel",
	"$hubspot_deal_deal_country",
	"$hubspot_deal_deal_create_date",
	"$hubspot_deal_deal_sdr",
	"$hubspot_deal_deal_source",
	"$hubspot_deal_deal_status",
	"$hubspot_deal_description",
	"$hubspot_deal_hs_acv",
	"$hubspot_deal_hs_arr",
	"$hubspot_deal_hs_date_entered_1490828",
	"$hubspot_deal_hs_date_entered_9561448",
	"$hubspot_deal_hs_date_entered_9561449",
	"$hubspot_deal_hs_date_entered_9561450",
	"$hubspot_deal_hs_date_entered_appointmentscheduled",
	"$hubspot_deal_hs_date_exited_1490828",
	"$hubspot_deal_hs_date_exited_9561448",
	"$hubspot_deal_hs_date_exited_9561449",
	"$hubspot_deal_hs_date_exited_9561450",
	"$hubspot_deal_hs_date_exited_appointmentscheduled",
	"$hubspot_deal_hs_merged_object_ids",
	"$hubspot_deal_hs_mrr",
	"$hubspot_deal_hs_predicted_amount",
	"$hubspot_deal_hs_predicted_amount_in_home_currency",
	"$hubspot_deal_hs_tcv",
	"$hubspot_deal_hs_time_in_1490828",
	"$hubspot_deal_hs_time_in_9561448",
	"$hubspot_deal_hs_time_in_9561449",
	"$hubspot_deal_hs_time_in_9561450",
	"$hubspot_deal_hs_time_in_appointmentscheduled",
	"$hubspot_deal_lead_source",
	"$hubspot_deal_minimum_guarantee_",
	"$hubspot_deal_moved_to_onboarding_date",
	"$hubspot_deal_nal_gift_delivered",
	"$hubspot_deal_recurring_revenue_deal_type",
	"$hubspot_deal_sdr_opportunity_created",
	"$hubspot_deal_transfer_to_ae_temporary_",
	"$hubspot_deal_website_domain",
	"$hubspot_deal_zoho_import",
	"$initial_adgroup_id",
	"$initial_campaign_id",
	"$latest_adgroup_id",
	"$latest_campaign_id",
	"$latest_keyword",
	"$latest_term",
	"$salesforce_account_billingcity",
	"$salesforce_account_billingcountry",
	"$salesforce_account_billingpostalcode",
	"$salesforce_account_billingstreet",
	"$salesforce_account_createdbyid",
	"$salesforce_account_createddate",
	"$salesforce_account_current_customer__c",
	"$salesforce_account_description",
	"$salesforce_account_id",
	"$salesforce_account_isdeleted",
	"$salesforce_account_iv__insideview_company_id__c",
	"$salesforce_account_iv__insideview_created__c",
	"$salesforce_account_iv__insideview_data_integrity_status__c",
	"$salesforce_account_iv__insideview_date_last_updated__c",
	"$salesforce_account_iv__insideview_match_status__c",
	"$salesforce_account_iv__insideview_parent_company_id__c",
	"$salesforce_account_iv__insideview_ultimate_parent_company_id__c",
	"$salesforce_account_iv__insideview_user_last_updated__c",
	"$salesforce_account_lastactivitydate",
	"$salesforce_account_lastmodifiedbyid",
	"$salesforce_account_lastmodifieddate",
	"$salesforce_account_name",
	"$salesforce_account_ownerid",
	"$salesforce_account_phone",
	"$salesforce_account_shippingcity",
	"$salesforce_account_shippingpostalcode",
	"$salesforce_account_shippingstate",
	"$salesforce_account_shippingstreet",
	"$salesforce_account_systemmodstamp",
	"$salesforce_account_type",
	"$salesforce_account_website",
	"$salesforce_contact_accountid",
	"$salesforce_contact_createdbyid",
	"$salesforce_contact_createddate",
	"$salesforce_contact_email",
	"$salesforce_contact_firstname",
	"$salesforce_contact_hasoptedoutofemail",
	"$salesforce_contact_id",
	"$salesforce_contact_isdeleted",
	"$salesforce_contact_iv__insideview_created__c",
	"$salesforce_contact_iv__insideview_data_integrity_status__c",
	"$salesforce_contact_iv__insideview_match_status__c",
	"$salesforce_contact_lastactivitydate",
	"$salesforce_contact_lastmodifiedbyid",
	"$salesforce_contact_lastmodifieddate",
	"$salesforce_contact_lastname",
	"$salesforce_contact_mailingcity",
	"$salesforce_contact_mailingcountry",
	"$salesforce_contact_mailingpostalcode",
	"$salesforce_contact_mailingstate",
	"$salesforce_contact_mailingstreet",
	"$salesforce_contact_mc4sf__mc_subscriber__c",
	"$salesforce_contact_mobilephone",
	"$salesforce_contact_name",
	"$salesforce_contact_ownerid",
	"$salesforce_contact_phone",
	"$salesforce_contact_pi__conversion_date__c",
	"$salesforce_contact_pi__conversion_object_name__c",
	"$salesforce_contact_pi__conversion_object_type__c",
	"$salesforce_contact_pi__created_date__c",
	"$salesforce_contact_pi__first_activity__c",
	"$salesforce_contact_pi__grade__c",
	"$salesforce_contact_pi__last_activity__c",
	"$salesforce_contact_pi__needs_score_synced__c",
	"$salesforce_contact_pi__pardot_hard_bounced__c",
	"$salesforce_contact_pi__pardot_last_scored_at__c",
	"$salesforce_contact_pi__url__c",
	"$salesforce_contact_reportstoid",
	"$salesforce_contact_systemmodstamp",
	"$salesforce_lead_calendlycreated__c",
	"$salesforce_lead_company",
	"$salesforce_lead_convertedaccountid",
	"$salesforce_lead_convertedcontactid",
	"$salesforce_lead_converteddate",
	"$salesforce_lead_convertedopportunityid",
	"$salesforce_lead_createdbyid",
	"$salesforce_lead_createddate",
	"$salesforce_lead_description",
	"$salesforce_lead_donotcall",
	"$salesforce_lead_email",
	"$salesforce_lead_email_2__c",
	"$salesforce_lead_emailbounceddate",
	"$salesforce_lead_emailbouncedreason",
	"$salesforce_lead_firstname",
	"$salesforce_lead_hasoptedoutofemail",
	"$salesforce_lead_hqphone__c",
	"$salesforce_lead_id",
	"$salesforce_lead_isconverted",
	"$salesforce_lead_isdeleted",
	"$salesforce_lead_isunreadbyowner",
	"$salesforce_lead_iv__insideview_company_id__c",
	"$salesforce_lead_iv__insideview_created__c",
	"$salesforce_lead_iv__insideview_data_integrity_status__c",
	"$salesforce_lead_iv__insideview_date_last_updated__c",
	"$salesforce_lead_iv__insideview_employment_id__c",
	"$salesforce_lead_iv__insideview_executive_id__c",
	"$salesforce_lead_iv__insideview_match_status__c",
	"$salesforce_lead_iv__insideview_parent_company_id__c",
	"$salesforce_lead_iv__insideview_ultimate_parent_company_id__c",
	"$salesforce_lead_iv__insideview_user_last_updated__c",
	"$salesforce_lead_lastactivitydate",
	"$salesforce_lead_lastmodifiedbyid",
	"$salesforce_lead_lastmodifieddate",
	"$salesforce_lead_lastname",
	"$salesforce_lead_mc4sf__mc_subscriber__c",
	"$salesforce_lead_mobile_number__c",
	"$salesforce_lead_name",
	"$salesforce_lead_notes__c",
	"$salesforce_lead_ownerid",
	"$salesforce_lead_phone",
	"$salesforce_lead_pi__conversion_date__c",
	"$salesforce_lead_pi__conversion_object_name__c",
	"$salesforce_lead_pi__conversion_object_type__c",
	"$salesforce_lead_pi__created_date__c",
	"$salesforce_lead_pi__first_activity__c",
	"$salesforce_lead_pi__first_touch_url__c",
	"$salesforce_lead_pi__grade__c",
	"$salesforce_lead_pi__last_activity__c",
	"$salesforce_lead_pi__needs_score_synced__c",
	"$salesforce_lead_pi__pardot_hard_bounced__c",
	"$salesforce_lead_pi__pardot_last_scored_at__c",
	"$salesforce_lead_pi__score__c",
	"$salesforce_lead_pi__url__c",
	"$salesforce_lead_postalcode",
	"$salesforce_lead_street",
	"$salesforce_lead_systemmodstamp",
	"$salesforce_lead_time__c",
	"$salesforce_lead_title",
	"$salesforce_lead_website",
	"$salesforce_lead_x3rd_party_data_source_provider__c",
	"$salesforce_opportunity_account_sdr__c",
	"$salesforce_opportunity_accountid",
	"$salesforce_opportunity_amount",
	"$salesforce_opportunity_arr_amount__c",
	"$salesforce_opportunity_attributed_sdr__c",
	"$salesforce_opportunity_barriers__c",
	"$salesforce_opportunity_closedate",
	"$salesforce_opportunity_createdbyid",
	"$salesforce_opportunity_createddate",
	"$salesforce_opportunity_cro_commit_level__c",
	"$salesforce_opportunity_description",
	"$salesforce_opportunity_fiscal",
	"$salesforce_opportunity_fiscalquarter",
	"$salesforce_opportunity_fiscalyear",
	"$salesforce_opportunity_forecastcategory",
	"$salesforce_opportunity_forecastcategoryname",
	"$salesforce_opportunity_hasopportunitylineitem",
	"$salesforce_opportunity_high_amount__c",
	"$salesforce_opportunity_id",
	"$salesforce_opportunity_isclosed",
	"$salesforce_opportunity_isdeleted",
	"$salesforce_opportunity_iswon",
	"$salesforce_opportunity_key_opportunity__c",
	"$salesforce_opportunity_lastactivitydate",
	"$salesforce_opportunity_lastmodifiedbyid",
	"$salesforce_opportunity_lastmodifieddate",
	"$salesforce_opportunity_leadsource",
	"$salesforce_opportunity_name",
	"$salesforce_opportunity_nextstep",
	"$salesforce_opportunity_ownerid",
	"$salesforce_opportunity_primary_use_case__c",
	"$salesforce_opportunity_probability",
	"$salesforce_opportunity_sor__c",
	"$salesforce_opportunity_subscription_end_date__c",
	"$salesforce_opportunity_subscription_start_date__c",
	"$salesforce_opportunity_systemmodstamp",
	"$salesforce_opportunity_this_qtr_probability__c",
	"$salesforce_opportunity_ticket_vol_monthly__c",
	"$initial_keyword_match_type",
	"$latest_adgroup",
	"$latest_keyword_match_type",
	"$salesforce_lead_alternate_contact_number__c",
	"$salesforce_lead_call_count__c",
	"$salesforce_lead_country_code__c",
	"$salesforce_lead_custom_lead_id__c",
	"$salesforce_lead_date_when_meeting_is_scheduled__c",
	"$salesforce_lead_ec_location__c",
	"$salesforce_lead_enquiry_id__c",
	"$salesforce_lead_first_date_of_contact__c",
	"$salesforce_lead_first_date_of_contact_to_qualified_c__c",
	"$salesforce_lead_follow_up_count__c",
	"$salesforce_lead_follow_up_date_time__c",
	"$salesforce_lead_gclid__c",
	"$salesforce_lead_has_designer_accepted__c",
	"$salesforce_lead_ipaddress__c",
	"$salesforce_lead_is_designer_assigned__c",
	"$salesforce_lead_lead_allocation_time__c",
	"$salesforce_lead_lead_owner_name__c",
	"$salesforce_lead_lead_qualified_date__c",
	"$salesforce_lead_lockdown_survey__c",
	"$salesforce_lead_messaging_source__c",
	"$salesforce_lead_mobile_number_external_field__c",
	"$salesforce_lead_mobilephone",
	"$salesforce_lead_mobileym__c",
	"$salesforce_lead_otp_verified__c",
	"$salesforce_lead_page__c",
	"$salesforce_lead_pre_qualified_date__c",
	"$salesforce_lead_property_possession_date__c",
	"$salesforce_lead_recontacted__c",
	"$salesforce_lead_requirement_details__c",
	"$salesforce_lead_salutation",
	"$salesforce_lead_time_on_last_page__c",
	"$salesforce_lead_user_browser__c",
	"$salesforce_lead_user_mobile__c",
	"$salesforce_lead_user_os__c",
	"$salesforce_lead_whatsapp_opt_in__c",
	"$salesforce_lead_willingness_for_meeting__c",
	"$salesforce_opportunity_ad_group__c",
	"$salesforce_opportunity_ad_name__c",
	"$salesforce_opportunity_affiliate_name__c",
	"$salesforce_opportunity_call_center_agent__c",
	"$salesforce_opportunity_cmm_name__c",
	"$salesforce_opportunity_cmm_team__c",
	"$salesforce_opportunity_customer_id__c",
	"$salesforce_opportunity_dc_lead_source__c",
	"$salesforce_opportunity_designer__c",
	"$salesforce_opportunity_dsa__c",
	"$salesforce_opportunity_enquiry_id__c",
	"$salesforce_opportunity_expectedrevenue",
	"$salesforce_opportunity_meeting_scheduled_date_time__c",
	"$salesforce_opportunity_meeting_type__c",
	"$salesforce_opportunity_meeting_venue__c",
	"$salesforce_opportunity_messaging_source__c",
	"$salesforce_opportunity_mobile__c",
	"$salesforce_opportunity_mobileym__c",
	"$salesforce_opportunity_payment_mode__c",
	"$salesforce_opportunity_phone__c",
	"$salesforce_opportunity_project_name__c",
	"$salesforce_opportunity_property_address__c",
	"$salesforce_opportunity_proposed_budget__c",
	"$salesforce_opportunity_total_amount__c",
	"$salesforce_opportunity_wohoo_card__c",
	"SF_Ad_Group",
	"SF_Call_Stage",
	"SF_Created_Date",
	"SF_Last_Name",
	"SF_Lead_ID",
	"SF_Mobile",
	"SF_Opportunity_ID",
	"$hubspot_company_abm_campaign",
	"$hubspot_company_about_us",
	"$hubspot_company_account_owner_abm_outbound_",
	"$hubspot_company_allbound_id",
	"$hubspot_company_allbound_status",
	"$hubspot_company_bdr_owner",
	"$hubspot_company_company_tags",
	"$hubspot_company_company_temp_score",
	"$hubspot_company_contact_hs_band",
	"$hubspot_company_country_workflow_",
	"$hubspot_company_dummy_field",
	"$hubspot_company_first_demo_booked_on",
	"$hubspot_company_for_ops_test",
	"$hubspot_company_freshsales_account_id",
	"$hubspot_company_ls_change",
	"$hubspot_company_marketing_conversion_mode",
	"$hubspot_company_open_deal_amount",
	"$hubspot_company_partner_type",
	"$hubspot_company_salesloft_account_id",
	"$hubspot_company_salesloft_last_contacted_at",
	"$hubspot_company_tier_enrollment_date",
	"$hubspot_company_won_deal_amount",
	"$hubspot_company_zoominfo_company_id",
	"$hubspot_contact_abm_lead",
	"$hubspot_contact_account_claimed_date",
	"$hubspot_contact_activation_date",
	"$hubspot_contact_adwords_campaign_id_fs_",
	"$hubspot_contact_adwords_campaign_keyword_fs_",
	"$hubspot_contact_ae_territory_fs_",
	"$hubspot_contact_allbound_id",
	"$hubspot_contact_assignee_email_sdr_",
	"$hubspot_contact_bdr_notes",
	"$hubspot_contact_bdr_owner",
	"$hubspot_contact_became_a_blog_subscriber",
	"$hubspot_contact_behavior_score_hs_",
	"$hubspot_contact_booking_status_cp__c",
	"$hubspot_contact_business_type_id",
	"$hubspot_contact_c_n_l_campaign__lost_reason",
	"$hubspot_contact_calendly_source_fs_",
	"$hubspot_contact_call_booked_via_chat_ldr",
	"$hubspot_contact_call_scheduled_date",
	"$hubspot_contact_canceled_cp__c",
	"$hubspot_contact_cancellation_reason",
	"$hubspot_contact_cb___must_win",
	"$hubspot_contact_cf_average_revenue_per_customer_arpu",
	"$hubspot_contact_chargebee_customer_id",
	"$hubspot_contact_chargebee_merchant_signup_date",
	"$hubspot_contact_chargebee_site_name",
	"$hubspot_contact_chargebee_site_status",
	"$hubspot_contact_chargebee_team",
	"$hubspot_contact_clearbit_must_win",
	"$hubspot_contact_clearbit_reveal_company_name",
	"$hubspot_contact_clearbit_reveal_company_tags",
	"$hubspot_contact_company_tags",
	"$hubspot_contact_company_tech_categories",
	"$hubspot_contact_company_url",
	"$hubspot_contact_contact_owner_fs_",
	"$hubspot_contact_continent_fs_",
	"$hubspot_contact_country_drift",
	"$hubspot_contact_coupons_applied",
	"$hubspot_contact_course_completion_percentage",
	"$hubspot_contact_csm_calendly_link",
	"$hubspot_contact_csm_ces_email_id",
	"$hubspot_contact_csm_owner_cbm_",
	"$hubspot_contact_currencies_enabled__cbm",
	"$hubspot_contact_current_activity_id",
	"$hubspot_contact_current_billing_system",
	"$hubspot_contact_customer_bucket_cbm_",
	"$hubspot_contact_demo_booked_by",
	"$hubspot_contact_demo_booked_on",
	"$hubspot_contact_demo_booked_segment",
	"$hubspot_contact_demo_booked_yes_no_",
	"$hubspot_contact_demo_give_on",
	"$hubspot_contact_demo_scheduled_for_fs_",
	"$hubspot_contact_demo_scheduled_on_fs_",
	"$hubspot_contact_demographic_score_hs_",
	"$hubspot_contact_do_not_disturb",
	"$hubspot_contact_email_deliverability_status",
	"$hubspot_contact_email_invalid_",
	"$hubspot_contact_email_invalid_cause",
	"$hubspot_contact_existing_payment_gateway_fs_",
	"$hubspot_contact_geography__cbm",
	"$hubspot_contact_hosted_region",
	"$hubspot_contact_hs_email_optout_4356543",
	"$hubspot_contact_hs_email_optout_4622904",
	"$hubspot_contact_hs_email_optout_4623032",
	"$hubspot_contact_hs_email_optout_5608281",
	"$hubspot_contact_hs_email_optout_5657505",
	"$hubspot_contact_hs_email_optout_5773070",
	"$hubspot_contact_hs_email_optout_5830713",
	"$hubspot_contact_hs_email_optout_5868516",
	"$hubspot_contact_hs_email_optout_6714938",
	"$hubspot_contact_hs_email_optout_6860878",
	"$hubspot_contact_hs_email_optout_6932217",
	"$hubspot_contact_hs_email_optout_9535844",
	"$hubspot_contact_hs_email_quarantined",
	"$hubspot_contact_hs_email_quarantined_reason",
	"$hubspot_contact_hs_email_recipient_fatigue_recovery_time",
	"$hubspot_contact_hs_emailconfirmationstatus",
	"$hubspot_contact_hs_first_engagement_date",
	"$hubspot_contact_hs_first_engagement_descr",
	"$hubspot_contact_hs_first_engagement_object_type",
	"$hubspot_contact_hs_linkedinid",
	"$hubspot_contact_hubspot_must_win",
	"$hubspot_contact_i_would_like_to_get_a_demo_from_a_chargebee_expert",
	"$hubspot_contact_influ2_contact",
	"$hubspot_contact_inquiry_details",
	"$hubspot_contact_ip_country_fs_",
	"$hubspot_contact_ip_latlon",
	"$hubspot_contact_ip_zipcode",
	"$hubspot_contact_is_3ds_enabled_cbm",
	"$hubspot_contact_is_active_on_other_sites_cbm_",
	"$hubspot_contact_is_active_on_site_cbm_",
	"$hubspot_contact_is_customer_cbm_",
	"$hubspot_contact_is_disposable_email",
	"$hubspot_contact_is_free_email",
	"$hubspot_contact_is_live_site_user_cbm_",
	"$hubspot_contact_is_on_c4e_",
	"$hubspot_contact_is_role_based_email",
	"$hubspot_contact_is_rs_enabled_cbm",
	"$hubspot_contact_is_site_owner_cbm_",
	"$hubspot_contact_is_test_site_user_cbm_",
	"$hubspot_contact_lead_created_date_fs_",
	"$hubspot_contact_lead_description",
	"$hubspot_contact_lead_score_bucket",
	"$hubspot_contact_lead_shift",
	"$hubspot_contact_lead_stage_drift",
	"$hubspot_contact_lead_stage_fs_",
	"$hubspot_contact_mailchimp_subscription_status",
	"$hubspot_contact_marketing_purpose",
	"$hubspot_contact_medium_fs_",
	"$hubspot_contact_meeting_booked_on",
	"$hubspot_contact_meeting_creation_date_cp__c",
	"$hubspot_contact_meeting_type_cp__c",
	"$hubspot_contact_military_status",
	"$hubspot_contact_must_win",
	"$hubspot_contact_n14_day_trial_contact_sales",
	"$hubspot_contact_n14_day_trial_extension_completed",
	"$hubspot_contact_no_show_cp__c",
	"$hubspot_contact_np_449a91a5_328f_41e2_bbcd_d4839221677a",
	"$hubspot_contact_np_79226b5f_0d65_459a_8bde_a6cd128f15cf",
	"$hubspot_contact_np_a197850e_4ae8_478b_966a_c6be8a272937",
	"$hubspot_contact_onboarding_type",
	"$hubspot_contact_owner_email",
	"$hubspot_contact_partnership_type",
	"$hubspot_contact_persona_sl_",
	"$hubspot_contact_pre_hubspot_lead",
	"$hubspot_contact_qresult_og_",
	"$hubspot_contact_quiz_name_og_",
	"$hubspot_contact_recaptcha_score",
	"$hubspot_contact_rev_segment__cb_ranges_",
	"$hubspot_contact_revenue_per_annum",
	"$hubspot_contact_revenue_range_text",
	"$hubspot_contact_revops_cart_items",
	"$hubspot_contact_revops_cart_revenue_operations",
	"$hubspot_contact_revops_cart_subscription_analytics",
	"$hubspot_contact_revops_cart_subscription_management",
	"$hubspot_contact_role_cbm_",
	"$hubspot_contact_router_name_cp__c",
	"$hubspot_contact_rw_abm_contact",
	"$hubspot_contact_sa_implementation_help",
	"$hubspot_contact_salesloft_last_contacted_at",
	"$hubspot_contact_scaleup_tiers_fs_",
	"$hubspot_contact_sdr_call_booked_on__fs_",
	"$hubspot_contact_sdr_call_scheduled_on__fs_",
	"$hubspot_contact_sdr_owner",
	"$hubspot_contact_sector",
	"$hubspot_contact_segment_temp_",
	"$hubspot_contact_self_activated_",
	"$hubspot_contact_sendex_score",
	"$hubspot_contact_sic_code",
	"$hubspot_contact_signed_up_plan",
	"$hubspot_contact_site_time_zone__cbm",
	"$hubspot_contact_source_fs_",
	"$hubspot_contact_sub_sector",
	"$hubspot_contact_subscription_status_cbm_",
	"$hubspot_contact_tbd_from_hubspot_fs_",
	"$hubspot_contact_tpv_buckets_cbm",
	"$hubspot_contact_trial_end_date",
	"$hubspot_contact_trial_nurture_status",
	"$hubspot_contact_utm_campaign_cbwebsite",
	"$hubspot_contact_utm_content_cbwebsite",
	"$hubspot_contact_utm_gclid",
	"$hubspot_contact_utm_keyword_cbwebsite",
	"$hubspot_contact_utm_medium_cbwebsite",
	"$hubspot_contact_utm_source_cbwebsite",
	"$hubspot_contact_vertical__cbm",
	"$hubspot_contact_warmly_monitored_",
	"$hubspot_contact_warmly_new_company_url",
	"$hubspot_contact_warmly_new_email",
	"$hubspot_contact_weighted_score",
	"$hubspot_contact_what_do_you_seek_by_partnering_with_chargebee_",
	"$hubspot_contact_what_does_your_company_do",
	"$hubspot_contact_what_geographies_do_you_work_with_",
	"$hubspot_contact_what_type_of_company_are_you_",
	"$hubspot_contact_where_is_your_company_incorporated_",
	"$hubspot_contact_why_do_you_want_to_partner_with_chargebee_",
	"$hubspot_contact_zi_linkedin_url",
	"$hubspot_contact_zoominfo_company_id",
	"$hubspot_contact_zoominfo_contact_id",
	"$hubspot_deal_ae_territory",
	"$hubspot_deal_expected_close_date",
	"$hubspot_deal_fs_deal_id",
	"$hubspot_deal_hs_date_entered_1817098",
	"$hubspot_deal_hs_date_entered_1817099",
	"$hubspot_deal_hs_date_entered_1817100",
	"$hubspot_deal_hs_date_entered_1817101",
	"$hubspot_deal_hs_date_entered_1817102",
	"$hubspot_deal_hs_date_exited_1817098",
	"$hubspot_deal_hs_date_exited_1817099",
	"$hubspot_deal_hs_date_exited_1817100",
	"$hubspot_deal_hs_date_exited_1817101",
	"$hubspot_deal_hs_time_in_1817098",
	"$hubspot_deal_hs_time_in_1817099",
	"$hubspot_deal_hs_time_in_1817100",
	"$hubspot_deal_hs_time_in_1817101",
	"$hubspot_deal_hs_time_in_1817102",
	"$hubspot_deal_revenue_segment_ae",
}

var disableGroupUserPropertiesByKeyPrefix = []string{
	"$hubspot_company_",
	"$hubspot_deal_",
	"$salesforce_opportunity_",
	"$salesforce_account_",
}

var explainPropertyWeights = map[string]float64{
	// weight based on git issue : 5849

	UP_INITIAL_CHANNEL:  1.5,
	UP_INITIAL_PAGE_URL: 1.5,
	UP_INITIAL_CAMPAIGN: 1.5,
	UP_LATEST_CHANNEL:   1.5,
	UP_LATEST_CAMPAIGN:  1.5,
	UP_LATEST_SOURCE:    1.5,
	UP_LATEST_MEDIUM:    1.5,
	EP_CHANNEL:          1.5,
	EP_MEDIUM:           1.5,
	EP_SOURCE:           1.5,
	UP_DEVICE_TYPE:      0.5,
	UP_OS_VERSION:       0.5,
	UP_OS:               0.5,
	UP_BROWSER:          0.5,
	UP_PLATFORM:         0.5,
	UP_DEVICE_BRAND:     0.5,
	UP_CONTINENT:        0.5,
	EP_CREATIVE:         0.5,
	EP_CONTENT:          0.5,
	UP_POSTAL_CODE:      0.01,
	EP_CAMPAIGN_ID:      0.01,
	EP_ADGROUP_ID:       0.01,
}

const SamplePropertyValuesLimit = 100

// defined property values.
// single letter bool value alias to save space.
const PROPERTY_VALUE_TRUE = "t"
const PROPERTY_VALUE_FALSE = "f"

// Properties should be present always, mainly for queries.
var MandatoryDefaultUserPropertiesByType = map[string][]string{
	PropertyTypeDateTime: []string{
		UP_JOIN_TIME,
	},
}

func DisableGroupUserPropertiesByKeyPrefix(key string) bool {
	for _, prefix := range disableGroupUserPropertiesByKeyPrefix {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

func FilterGroupUserPropertiesKeysByPrefix(propertyKeys []string) []string {
	filteredPropertiesKeys := make([]string, 0)
	for _, key := range propertyKeys {
		if DisableGroupUserPropertiesByKeyPrefix(key) {
			continue
		}
		filteredPropertiesKeys = append(filteredPropertiesKeys, key)
	}
	return filteredPropertiesKeys
}

func FilterPropertiesByKeysByPrefix(properties *PropertiesMap, prefix string) *PropertiesMap {
	if properties == nil {
		return nil
	}

	filteredProperties := make(PropertiesMap)
	for key := range *properties {
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		filteredProperties[key] = (*properties)[key]
	}

	return &filteredProperties
}

// isValidProperty - Validate property type.
func isPropertyTypeValid(value interface{}) error {
	if value == nil {
		return nil
	}

	switch valueType := value.(type) {
	case int:
	case int32:
	case int64:
	case float32:
	case float64:
	case string:
	case bool:
	default:
		log.WithFields(log.Fields{"value": value,
			"valueType": valueType}).Debug("Invalid type used on property")
		return fmt.Errorf("invalid property type")
	}
	return nil
}

func IsFormSubmitUserProperty(key string) bool {
	for _, k := range FORM_SUBMIT_USER_PROPERTIES {
		if k == key {
			return true
		}
	}
	return false
}

func isSDKAllowedUserProperty(key *string) bool {
	for _, k := range SDK_ALLOWED_USER_PROPERTIES {
		if k == *key {
			return true
		}
	}
	return false
}

func isSDKAllowedEventProperty(key *string) bool {
	for _, k := range SDK_ALLOWED_EVENT_PROPERTIES {
		if k == *key {
			return true
		}
	}
	return false
}

func IsInternalEventProperty(key *string) bool {
	for _, k := range INTERNAL_EVENT_PROPERTIES {
		if k == *key {
			return true
		}
	}
	return false
}

func IsInternalUserProperty(key *string) bool {
	for _, k := range INTERNAL_USER_PROPERTIES {
		if k == *key {
			return true
		}
	}
	return false
}

func IsGenericEventProperty(key *string) bool {
	for _, k := range GENERIC_NUMERIC_EVENT_PROPERTIES {
		if k == *key {
			return true
		}
	}
	return false
}

func IsGenericUserProperty(key *string) bool {
	for _, k := range GENERIC_NUMERIC_USER_PROPERTIES {
		if k == *key {
			return true
		}
	}
	return false
}

func IsPageViewEvent(eventPropertiesMap *PropertiesMap) bool {
	if eventPropertiesMap == nil {
		return false
	}

	valueAsInterface, exists := (*eventPropertiesMap)[EP_IS_PAGE_VIEW]
	if !exists {
		return false
	}

	value, ok := valueAsInterface.(bool)
	return ok && value
}

func GetUnEscapedPropertyValue(v interface{}) interface{} {
	switch v.(type) {
	case string:
		strValue := v.(string)
		if escapedPath, err := url.PathUnescape(strValue); err == nil {
			return escapedPath
		}
	}

	return v
}

func GetValidatedUserProperties(properties *PropertiesMap) *PropertiesMap {
	validatedProperties := make(PropertiesMap)
	for k, v := range *properties {
		if err := isPropertyTypeValid(v); err == nil {
			if strings.HasPrefix(k, NAME_PREFIX) &&
				!IsAllowedCRMPropertyPrefix(k) &&
				!isSDKAllowedUserProperty(&k) {

				validatedProperties[fmt.Sprintf("%s%s", NAME_PREFIX_ESCAPE_CHAR, k)] = v
			} else {
				validatedProperties[k] = v
			}
		} else {
			log.WithError(err).Warnf("Invalid type for property %s with value %v", k, v)
		}
	}
	return &validatedProperties
}

func isCRMSmartEventPropertyKey(key *string) bool {
	if !strings.HasPrefix((*key), SMART_EVENT_SALESFORCE_PREV_PROPERTY) &&
		!strings.HasPrefix((*key), SMART_EVENT_SALESFORCE_CURR_PROPERTY) &&
		!strings.HasPrefix((*key), SMART_EVENT_HUBSPOT_PREV_PROPERTY) &&
		!strings.HasPrefix((*key), SMART_EVENT_HUBSPOT_CURR_PROPERTY) &&
		(*key) != EP_CRM_REFERENCE_EVENT_ID {
		return false
	}

	return true
}

func IsAllowedCRMPropertyPrefix(name string) bool {
	for prefix := range AllowedCRMPropertyPrefix {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func GetValidatedEventProperties(properties *PropertiesMap) *PropertiesMap {
	validatedProperties := make(PropertiesMap)
	for k, v := range *properties {
		if err := isPropertyTypeValid(v); err == nil {
			var propertyKey string
			// Escape properties with $ prefix but allow query_params_props
			// with selected prefixes starting with $ and default properties.
			if strings.HasPrefix(k, NAME_PREFIX) &&
				!strings.HasPrefix(k, QUERY_PARAM_PROPERTY_PREFIX) &&
				!IsAllowedCRMPropertyPrefix(k) &&
				!isCRMSmartEventPropertyKey(&k) &&
				!isSDKAllowedEventProperty(&k) {
				propertyKey = fmt.Sprintf("%s%s", NAME_PREFIX_ESCAPE_CHAR, k)
			} else {
				propertyKey = k
			}

			if propertyKey == UP_EMAIL {
				email := GetEmailLowerCase(v)
				if email != "" {
					validatedProperties[propertyKey] = email
				}
			} else if propertyKey == UP_PHONE {
				sPhoneNo := SanitizePhoneNumber(v)
				if sPhoneNo != "" {
					validatedProperties[propertyKey] = sPhoneNo
				}
			} else {
				if k != EP_INTERNAL_IP {
					validatedProperties[propertyKey] = v
				}
			}
		}
	}
	return &validatedProperties
}

func UnEscapeQueryParamProperties(properties *PropertiesMap) {
	UnEscapedProperties := make(PropertiesMap)
	for k, v := range *properties {
		if strings.HasPrefix(k, QUERY_PARAM_PROPERTY_PREFIX) {
			UnEscapedProperties[GetUnEscapedPropertyValue(k).(string)] = GetUnEscapedPropertyValue(v)
		} else {
			UnEscapedProperties[k] = v
		}
	}
	*properties = UnEscapedProperties
}

func HasDefinedMarketingProperty(properties *PropertiesMap) bool {
	for _, marketingProperty := range DEFINED_MARKETING_PROPERTIES {
		if _, exists := (*properties)[marketingProperty]; exists {
			return true
		}
	}

	return false
}

func isNumericalPropertyByName(propertyKey string) bool {
	for _, key := range NUMERICAL_PROPERTY_BY_NAME {
		if key == propertyKey {
			return true
		}
	}

	return false
}

func isCategoricalPropertyByName(propertyKey string) bool {
	for _, key := range CATEGORICAL_PROPERTY_BY_NAME {
		if key == propertyKey {
			return true
		}
	}

	return false
}

func isDateTimePropertyByName(propertyKey string) bool {
	for _, key := range DATETIME_PROPERTY_BY_NAME {
		if key == propertyKey {
			return true
		}
	}

	return false
}

func GetPropertyTypeByName(propertyName string) string {
	// PropertyKey will be set to null if the pre-mentioned classfication behaviour need to be supressed
	if isDateTimePropertyByName(propertyName) {
		return PropertyTypeDateTime
	}
	if isNumericalPropertyByName(propertyName) {
		return PropertyTypeNumerical
	}
	return PropertyTypeCategorical
}

func GetPropertyTypeByKeyORValue(projectID int64, eventName string, propertyKey string, propertyValue interface{}, isUserProperty bool) (string, bool) {
	// PropertyKey will be set to null if the pre-mentioned classfication behaviour need to be supressed
	if propertyKey != "" {

		if strings.HasPrefix(propertyKey, NAME_PREFIX) {
			if isNumericalPropertyByName(propertyKey) {
				return PropertyTypeNumerical, true
			}
			if isCategoricalPropertyByName(propertyKey) {
				return PropertyTypeCategorical, true
			}
			if isDateTimePropertyByName(propertyKey) {
				return PropertyTypeDateTime, true
			}
		}
		if IsPropertyNameContainsDateOrTime(propertyKey) {
			_, status := ConvertDateTimeValueToNumber(propertyValue)
			if status {
				return PropertyTypeDateTime, false
			}
		}
	}

	switch propertyValue.(type) {
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64:
		return PropertyTypeNumerical, false
	case string:
		return PropertyTypeCategorical, false
	default:
		return PropertyTypeUnknown, false
	}
}

func IsPropertyNameContainsDateOrTime(propertyName string) bool {
	propertyNameAllLower := strings.ToLower(propertyName)
	if strings.Contains(propertyNameAllLower, "date") || strings.Contains(propertyNameAllLower, "timestamp") {
		return true
	}
	return false
}

func ConvertDateTimeValueToNumber(propertyValue interface{}) (interface{}, bool) {
	propertyValueFloat64, err := GetPropertyValueAsFloat64(propertyValue)
	if err == nil {
		return propertyValueFloat64, true
	}
	return propertyValue, false
}

func GetUpdateAllowedEventProperties(properties *PropertiesMap) *PropertiesMap {
	allowedProperties := make(PropertiesMap)
	for key, value := range *properties {
		if strings.HasPrefix(key, NAME_PREFIX) {
			for _, allowedKey := range UPDATE_ALLOWED_EVENT_PROPERTIES {
				if key == allowedKey {
					allowedProperties[key] = value
					continue
				}
			}
		} else {
			allowedProperties[key] = value
		}
	}

	return &allowedProperties
}

// GetUpdateAllowedInitialUserProperties - Returns update allowed initial
// user_properties based on the update allowed event_properties.
func GetUpdateAllowedInitialUserProperties(eventProperties *PropertiesMap) *PropertiesMap {
	newInitialUserProperties := make(PropertiesMap, 0)

	if eventProperties == nil {
		return &newInitialUserProperties
	}

	for _, eventProperty := range UPDATE_ALLOWED_EVENT_PROPERTIES {
		eventPropertyValue, exists := (*eventProperties)[eventProperty]
		if !exists {
			continue
		}

		initialUserProperty, exists := EVENT_TO_USER_INITIAL_PROPERTIES[eventProperty]
		if !exists {
			continue
		}

		newInitialUserProperties[initialUserProperty] = eventPropertyValue
	}

	return &newInitialUserProperties
}

func FillInitialUserProperties(newUserProperties *PropertiesMap, eventID string,
	eventProperties *PropertiesMap, existingUserProperties *map[string]interface{},
	isPropertiesDefaultableRequest bool) {

	if existingUserProperties == nil {
		existingUserProperties = &map[string]interface{}{}
	}

	var initialUserPropertiesExists bool
	for _, property := range EVENT_TO_USER_INITIAL_PROPERTIES {
		if _, exists := (*existingUserProperties)[property]; exists {
			initialUserPropertiesExists = true
			break
		}
	}

	if newUserProperties == nil {
		newUserProperties = &PropertiesMap{}
	}

	// Add value, if property doesn't exist already
	// and default value allowed property.
	if isPropertiesDefaultableRequest {
		for k, v := range DEFAULT_USER_PROPERTY_VALUES {
			if _, exists := (*existingUserProperties)[k]; !exists {
				(*newUserProperties)[k] = v
			}
		}
	}

	if initialUserPropertiesExists {
		return
	}

	for k, v := range *eventProperties {
		if userPropertyKey, exists := EVENT_TO_USER_INITIAL_PROPERTIES[k]; exists {
			(*newUserProperties)[userPropertyKey] = v
		}
	}
	(*newUserProperties)[UP_INITIAL_PAGE_EVENT_ID] = eventID
}

func GetSessionProperties(isFirstSession bool, eventProperties,
	userProperties *PropertiesMap) *PropertiesMap {
	sessionProperties := make(PropertiesMap)

	if isFirstSession {
		sessionProperties[SP_IS_FIRST_SESSION] = isFirstSession
	}

	for k, v := range *userProperties {
		for _, property := range USER_TO_SESSION_PROPERTIES {
			if property == k {
				sessionProperties[k] = v
				break
			}
		}
	}

	for k, v := range *eventProperties {
		if property, exists := EVENT_TO_SESSION_PROPERTIES[k]; exists {
			sessionProperties[property] = v
		}
	}

	return &sessionProperties
}

func FillHourDayAndTimestampEventProperty(properties *postgres.Jsonb, timestamp int64, timezoneString TimeZoneString) (*postgres.Jsonb, error) {
	t := ConvertTimeIn(time.Unix(timestamp, 0), timezoneString)
	weekDay := t.Weekday().String()
	hr, _, _ := t.Clock()
	eventPropsJSON, err := DecodePostgresJsonb(properties)
	if err != nil {
		return nil, err
	}
	(*eventPropsJSON)[EP_DAY_OF_WEEK] = weekDay
	(*eventPropsJSON)[EP_HOUR_OF_DAY] = hr
	(*eventPropsJSON)[EP_TIMESTAMP] = timestamp
	return EncodeToPostgresJsonb(eventPropsJSON)
}

// Moves datetime properties from numerical properties to type datetime.
// Few Properties, defined in factors are to be classified into right DataType.
func ClassifyDateTimePropertyKeys(propertiesByType *map[string][]string) map[string][]string {
	cProperties := make(map[string][]string, 0)

	datetime := (*propertiesByType)[PropertyTypeDateTime]
	numerical := make([]string, 0, 0)
	for _, prop := range (*propertiesByType)[PropertyTypeNumerical] {
		isDatetime := false
		for _, dtProp := range PROPERTIES_TYPE_DATE_TIME {
			if prop == dtProp {
				datetime = append(datetime, prop)
				isDatetime = true
				break
			}
		}

		if !isDatetime {
			numerical = append(numerical, prop)
		}
	}
	categorical := make([]string, 0, 0)
	for _, prop := range (*propertiesByType)[PropertyTypeCategorical] {
		isDatetime := false
		for _, dtProp := range PROPERTIES_TYPE_DATE_TIME {
			if prop == dtProp {
				datetime = append(datetime, prop)
				isDatetime = true
				break
			}
		}

		if !isDatetime {
			categorical = append(categorical, prop)
		}
	}
	cProperties[PropertyTypeNumerical] = numerical
	cProperties[PropertyTypeDateTime] = datetime
	cProperties[PropertyTypeCategorical] = categorical
	return cProperties
}

// Fills default user properties which should be present on properties list always.
func FillMandatoryDefaultUserProperties(propertiesByType *map[string][]string) {
	for propType, props := range *propertiesByType {
		if _, exists := MandatoryDefaultUserPropertiesByType[propType]; exists {
			for _, dProp := range MandatoryDefaultUserPropertiesByType[propType] {
				dPropExists := false
				for _, prop := range props {
					if prop == dProp {
						dPropExists = true
						break
					}
				}

				// adds missing default property.
				if !dPropExists {
					(*propertiesByType)[propType] = append((*propertiesByType)[propType], dProp)
				}
			}
		}
	}
}

func FillLatestPageUserProperties(userProperties, eventProperties *PropertiesMap) {
	for k, v := range *eventProperties {
		if userPropertyKey, exists := EVENT_TO_USER_LATEST_PAGE_PROPERTIES[k]; exists {
			(*userProperties)[userPropertyKey] = v
		}
	}
}

func FillLatestTouchUserProperties(userProperties, eventProperties *PropertiesMap) {
	for k, v := range *eventProperties {
		if userPropertyKey, exists := EVENT_TO_USER_LATEST_PROPERTIES[k]; exists {
			(*userProperties)[userPropertyKey] = v
		}
	}
}

func FillPropertiesFromURL(properties *PropertiesMap, url *url.URL) error {
	queryParams := url.Query()
	for k, v := range queryParams {
		// param can have multiple values as array, using 1st alone.
		(*properties)[QUERY_PARAM_PROPERTY_PREFIX+k] = v[0]
	}

	fragmentParams := GetQueryParamsFromURLFragment(url.Fragment)
	for k, v := range fragmentParams {
		(*properties)[QUERY_PARAM_PROPERTY_PREFIX+k] = v
	}

	return nil
}

func GetPropertyValueAsString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch valueType := value.(type) {
	case float32, float64:
		return fmt.Sprintf("%0.0f", value)
	case int, int32, int64:
		return fmt.Sprintf("%v", value)
	case string:
		return value.(string)
	case bool:
		return strconv.FormatBool(value.(bool))
	default:
		log.WithField("value_type", valueType).WithField("value", value).
			Error("Invalid value type on GetPropertyValueAsString")
		return ""
	}
}

func GetPropertyValueAsInt64(value interface{}) (int64, error) {
	if value == nil {
		return 0, nil
	}

	switch valueType := value.(type) {
	case float64:
		return int64(value.(float64)), nil
	case float32:
		return int64(value.(float32)), nil
	case int:
		return int64(value.(int)), nil
	case int32:
		return int64(value.(int32)), nil
	case int64:
		return value.(int64), nil
	case string:
		valueString := value.(string)
		if valueString == "" {
			return 0, nil
		}

		intValue, err := strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			return 0, err
		}
		return intValue, err
	default:
		return 0, fmt.Errorf("invalid property value type %v", valueType)
	}
}

func GetPropertyValueAsFloat64(value interface{}) (float64, error) {
	if value == nil {
		return 0, nil
	}

	switch valueType := value.(type) {
	case float64:
		return value.(float64), nil
	case float32:
		return float64(value.(float32)), nil
	case int:
		return float64(value.(int)), nil
	case int32:
		return float64(value.(int32)), nil
	case int64:
		return float64(value.(int64)), nil
	case string:
		valueString := value.(string)
		if valueString == "" {
			return 0, nil
		}

		floatValue, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			return 0, err
		}
		return floatValue, err
	default:
		return 0, fmt.Errorf("invalid property value type %v", valueType)
	}
}

func GetPredefinedBinRanges(propertyName string) ([][2]float64, bool) {
	predfinedBinRanges, found := PREDEFINED_BIN_RANGES_FOR_PROPERTY[propertyName]
	return predfinedBinRanges, found
}

func FillFirstEventUserPropertiesIfNotExist(existingUserProperties *map[string]interface{},
	newUserProperties *PropertiesMap, eventTimestamp int64) error {

	if eventTimestamp == 0 {
		return errors.New("invalid event timestamp")
	}

	// Should not add first event user properties, even if one of them already available.
	isAnyFirstEventUserPropertiesExist := existingUserProperties != nil &&
		((*existingUserProperties)[UP_HOUR_OF_FIRST_EVENT] != nil || (*existingUserProperties)[UP_DAY_OF_FIRST_EVENT] != nil)

	if !isAnyFirstEventUserPropertiesExist {
		(*newUserProperties)[UP_DAY_OF_FIRST_EVENT] = time.Unix(eventTimestamp, 0).Weekday().String()
		(*newUserProperties)[UP_HOUR_OF_FIRST_EVENT], _, _ = time.Unix(eventTimestamp, 0).Clock()
	}

	return nil
}

// FilterDisabledCoreUserProperties Filters out less important properties from the list.
func FilterDisabledCoreUserProperties(overrides []string, propertiesByType *map[string][]string) {
	overideMap := make(map[string]bool)
	for _, overide := range overrides {
		overideMap[overide] = true
	}
	DISABLED_CORE_QUERY_USER_PROPERTIES_Override := make([]string, 0)
	for _, property := range DISABLED_CORE_QUERY_USER_PROPERTIES {
		if !overideMap[property] {
			DISABLED_CORE_QUERY_USER_PROPERTIES_Override = append(DISABLED_CORE_QUERY_USER_PROPERTIES_Override, property)
		}
	}
	DISABLED_USER_PROPERTIES_UI_Override := make([]string, 0)
	for _, property := range DISABLED_USER_PROPERTIES_UI {
		if !overideMap[property] {
			DISABLED_USER_PROPERTIES_UI_Override = append(DISABLED_USER_PROPERTIES_UI_Override, property)
		}
	}
	for propertyType, properties := range *propertiesByType {
		(*propertiesByType)[propertyType] = StringSliceDiff(properties, DISABLED_CORE_QUERY_USER_PROPERTIES_Override[:])
	}
	for propertyType, properties := range *propertiesByType {
		(*propertiesByType)[propertyType] = StringSliceDiff(properties, DISABLED_USER_PROPERTIES_UI_Override[:])
	}
	for propertyType, properties := range *propertiesByType {
		(*propertiesByType)[propertyType] = FilterGroupUserPropertiesKeysByPrefix(properties)
	}
}

// FilterDisabledCoreEventProperties Filters out less important properties from the list.
func FilterDisabledCoreEventProperties(overrides []string, propertiesByType *map[string][]string) {
	overideMap := make(map[string]bool)
	for _, overide := range overrides {
		overideMap[overide] = true
	}
	DISABLED_CORE_QUERY_EVENT_PROPERTIES_Override := make([]string, 0)
	for _, property := range DISABLED_CORE_QUERY_EVENT_PROPERTIES {
		if !overideMap[property] {
			DISABLED_CORE_QUERY_EVENT_PROPERTIES_Override = append(DISABLED_CORE_QUERY_EVENT_PROPERTIES_Override, property)
		}
	}
	DISABLED_EVENT_PROPERTIES_UI_Override := make([]string, 0)
	for _, property := range DISABLED_EVENT_PROPERTIES_UI {
		if !overideMap[property] {
			DISABLED_EVENT_PROPERTIES_UI_Override = append(DISABLED_EVENT_PROPERTIES_UI_Override, property)
		}
	}
	for propertyType, properties := range *propertiesByType {
		(*propertiesByType)[propertyType] = StringSliceDiff(properties, DISABLED_CORE_QUERY_EVENT_PROPERTIES_Override[:])
	}
	for propertyType, properties := range *propertiesByType {
		(*propertiesByType)[propertyType] = StringSliceDiff(properties, DISABLED_EVENT_PROPERTIES_UI_Override[:])
	}
}

// ShouldIgnoreItreeProperty Checks if property is to be ignored for building ITree.
func ShouldIgnoreItreeProperty(propertyName string) bool {
	if _, found := ITREE_PROPERTIES_TO_IGNORE[propertyName]; found {
		return true
	}

	return IsInternalEventProperty(&propertyName) || IsInternalUserProperty(&propertyName)
}

// ShouldIgnoreItreeProperty Checks if property is to be ignored for building ITree.
func ShouldIgnoreItreeNumericalProperty(propertyName string) bool {
	if _, found := ITREE_NUMERICAL_PROPERTIES_TO_IGNORE[propertyName]; found {
		return true
	}
	return false
}

func SetDefaultValuesToEventProperties(eventProperties *PropertiesMap) {
	for property, defaultValue := range DEFAULT_EVENT_PROPERTY_VALUES {
		var setDefault bool
		if value, exists := (*eventProperties)[property]; exists {
			v, err := GetPropertyValueAsFloat64(value)
			setDefault = err == nil && v == 0
		} else {
			setDefault = true
		}

		var value interface{} = defaultValue
		// Treated default value for page_spent_time,
		// based on page_load_time.
		if setDefault && property == EP_PAGE_SPENT_TIME {
			pageLoadTime, err := GetPropertyValueAsFloat64((*eventProperties)[EP_PAGE_LOAD_TIME])
			if err == nil && pageLoadTime > 0 {
				value = (*eventProperties)[EP_PAGE_LOAD_TIME]
			}
		}

		if setDefault {
			(*eventProperties)[property] = value
		}
	}
}

func isURLProperty(property string) bool {
	propertiesWithoutURLSuffix := []string{
		EP_REFERRER,
		UP_INITIAL_REFERRER,
		UP_LATEST_REFERRER,
		SP_INITIAL_REFERRER,
		UP_INITIAL_PAGE_DOMAIN,
		UP_INITIAL_REFERRER_DOMAIN,
	}

	return strings.HasSuffix(property, "url") ||
		StringValueIn(property, propertiesWithoutURLSuffix)
}

func SanitizeProperties(properties *PropertiesMap) {
	for k, v := range *properties {
		if v == nil && !IsCRMPropertyKey(k) {
			continue
		}
		if isURLProperty(k) {
			(*properties)[k] = strings.TrimSuffix(v.(string), "/")
		}

		if IsPropertyNameContainsDateOrTime(k) {
			(*properties)[k], _ = ConvertDateTimeValueToNumber(v)
		}
	}
}

func SanitizePropertiesJsonb(propertiesJsonb *postgres.Jsonb) *postgres.Jsonb {
	propertiesMap, err := DecodePostgresJsonbAsPropertiesMap(propertiesJsonb)
	if err != nil {
		log.WithError(err).Error("Failed to decode JSON to sanitize properties.")
		return propertiesJsonb
	}

	SanitizeProperties(propertiesMap)

	propertiesJsonMap := map[string]interface{}(*propertiesMap)
	propertiesJsonb, err = EncodeToPostgresJsonb(&propertiesJsonMap)
	if err != nil {
		log.WithError(err).Error("Failed to encode sanitized JSON.")
		return propertiesJsonb
	}

	return propertiesJsonb
}

type CountTimestampTuple struct {
	LastSeenTimestamp int64  `json:"lst"`
	Count             int64  `json:"cnt"`
	Type              string `json:"type"`
}

type CachePropertyWithTimestamp struct {
	Property map[string]PropertyWithTimestamp `json:"pr"`
}

type PropertyWithTimestamp struct {
	Category          string              `json:"ca"`
	CategorywiseCount map[string]int64    `json:"cwc"` // Not to be used by handlers. Only cache set will use it before computing category
	CountTime         CountTimestampTuple `json:"ct"`
}

type CachePropertyValueWithTimestamp struct {
	PropertyValue map[string]CountTimestampTuple `json:"pv"`
}

type NameCountTimestampCategory struct {
	Name      string
	Count     int64
	Timestamp int64
	Category  string
	GroupName string
}

// isElementPresent checks if an element is present in a slice
func isElementPresent(elementsList []string, element string) bool {
	for _, value := range elementsList {
		if value == element {
			return true
		}
	}
	return false
}

// SortByTimestampAndCount Sorts the given array by timestamp/count
// Pick all past 24 hours event and sort the remaining by count and return
// No filtering is done in this method
func SortByTimestampAndCount(data []NameCountTimestampCategory) []NameCountTimestampCategory {
	mandatoryEventsList := []string{EVENT_NAME_SESSION, EVENT_NAME_FORM_SUBMITTED, EVENT_NAME_FORM_FILL}
	smartEventNames := make([]NameCountTimestampCategory, 0)
	pageViewEventNames := make([]NameCountTimestampCategory, 0)
	sorted := make([]NameCountTimestampCategory, 0)
	mandatoryEventNames := make([]NameCountTimestampCategory, 0)
	trimmed := make([]NameCountTimestampCategory, 0)

	sort.Slice(data, func(i, j int) bool {
		return data[i].Count > data[j].Count
	})

	for index := range data {
		if data[index].Category == SmartEvent {
			data[index].GroupName = SmartEvent
		} else if data[index].Category == PageViewEvent {
			data[index].GroupName = PageViewEvent
		} else {
			data[index].GroupName = FrequentlySeen
		}

	}

	for _, details := range data {
		if isElementPresent(mandatoryEventsList, details.Name) {
			mandatoryEventNames = append(mandatoryEventNames, details)
		} else if details.Category == SmartEvent {
			smartEventNames = append(smartEventNames, details)
		} else if details.Category == PageViewEvent {
			pageViewEventNames = append(pageViewEventNames, details)
		} else {
			trimmed = append(trimmed, details)
		}

	}

	sorted = append(smartEventNames, sorted...)
	sorted = append(sorted, mandatoryEventNames...)
	sorted = append(sorted, pageViewEventNames...)

	for _, data := range trimmed {
		sorted = append(sorted, data)
	}
	return sorted
}

// AggregatePropertyValuesAcrossDate values are stored by date and this method aggregates the count and last seen value and returns
// no filtering is done
func AggregatePropertyValuesAcrossDate(values []CachePropertyValueWithTimestamp) []NameCountTimestampCategory {
	valuesAggregated := make(map[string]CountTimestampTuple)
	// Sort Event Properties by timestamp, count and return top n
	for _, valueList := range values {
		for valueName, valueDetails := range valueList.PropertyValue {
			valuesAggregatedInt := valuesAggregated[valueName]
			valuesAggregatedInt.Count += valueDetails.Count
			if valuesAggregatedInt.LastSeenTimestamp < valueDetails.LastSeenTimestamp {
				valuesAggregatedInt.LastSeenTimestamp = valueDetails.LastSeenTimestamp
			}
			valuesAggregated[valueName] = valuesAggregatedInt
		}
	}
	propertyValueAggregatedSlice := make([]NameCountTimestampCategory, 0)
	for k, v := range valuesAggregated {
		propertyValueAggregatedSlice = append(propertyValueAggregatedSlice, NameCountTimestampCategory{
			k, v.Count, v.LastSeenTimestamp, "", ""})
	}
	return propertyValueAggregatedSlice
}

// AggregatePropertyAcrossDate values are stored by date and this method aggregates the count and last seen value and returns
// no filtering is done
func AggregatePropertyAcrossDate(properties []CachePropertyWithTimestamp) []NameCountTimestampCategory {
	propertiesAggregated := make(map[string]PropertyWithTimestamp)
	propertyCategoryAggregated := make(map[string]map[string]int64)
	// Sort Event Properties by timestamp, count and return top n
	for _, PropertyList := range properties {
		for propertyName, propertyDetails := range PropertyList.Property {
			propertiesAggregatedInt := propertiesAggregated[propertyName]
			for cat, count := range propertyDetails.CategorywiseCount {
				if propertyCategoryAggregated[propertyName] == nil {
					propertyCategoryAggregated[propertyName] = make(map[string]int64)
				}
				propertyCategoryAggregated[propertyName][cat] += count
			}
			propertiesAggregatedInt.Category = propertyDetails.Category
			propertiesAggregatedInt.CountTime.Count += propertyDetails.CountTime.Count
			if propertiesAggregatedInt.CountTime.LastSeenTimestamp < propertyDetails.CountTime.LastSeenTimestamp {
				propertiesAggregatedInt.CountTime.LastSeenTimestamp = propertyDetails.CountTime.LastSeenTimestamp
			}
			propertiesAggregated[propertyName] = propertiesAggregatedInt
		}
	}
	for property, details := range propertiesAggregated {
		propAgg := details
		propAgg.Category = DeriveCategory(propertyCategoryAggregated[property], details.CountTime.Count)
		propertiesAggregated[property] = propAgg
	}

	propertiesAggregatedSlice := make([]NameCountTimestampCategory, 0)
	for k, v := range propertiesAggregated {
		propertiesAggregatedSlice = append(propertiesAggregatedSlice, NameCountTimestampCategory{
			k, v.CountTime.Count, v.CountTime.LastSeenTimestamp, v.Category, ""})
	}
	return propertiesAggregatedSlice
}

type Property struct {
	Key      string `json:"key"`
	Count    int64  `json:"count"`
	LastSeen uint64 `json:"last_seen"`
}

type PropertyValue struct {
	Value     string `json:"value"`
	Count     int64  `json:"count"`
	LastSeen  uint64 `json:"last_seen"`
	ValueType string `json:"value_type"`
}

func GetCategoryType(propertyName string, values []PropertyValue) string {
	if len(values) == 0 {
		return ""
	}
	valueType := make(map[string]int64)
	for _, value := range values {
		if IsPropertyNameContainsDateOrTime(propertyName) {
			_, status := ConvertDateTimeValueToNumber(value.Value)
			if status == true {
				valueType[PropertyTypeDateTime]++
				continue
			}
		}
		if value.ValueType == "string" {
			valueType[PropertyTypeCategorical]++
		}
		if value.ValueType == "number" || value.ValueType == "double" {
			valueType[PropertyTypeNumerical]++
		}
	}
	return DeriveCategory(valueType, int64(len(values)))
}

func DeriveCategory(categorySplit map[string]int64, totalCount int64) string {
	acceptablePercentage := int64(95)

	for category, count := range categorySplit {
		if count*100/totalCount >= acceptablePercentage {
			return category
		}
	}
	return PropertyTypeCategorical
}

// FillPropertyKvsFromPropertiesJson - Fills properties key with limited
// no.of of values propertiesKvs -> map[propertyKey]map[propertyValue]true
func FillPropertyKvsFromPropertiesJson(propertiesJson []byte,
	propertiesKvs *map[string]map[interface{}]bool, valuesLimit int) error {
	var rowProperties map[string]interface{}
	err := json.Unmarshal(propertiesJson, &rowProperties)
	if err != nil {
		return err
	}

	for k, v := range rowProperties {
		// allow only string, float and bool valued
		// properties.
		_, strOk := v.(string)
		_, fltOk := v.(float64)
		_, boolOk := v.(bool)
		if !strOk && !fltOk && !boolOk {
			continue
		}

		if _, ok := (*propertiesKvs)[k]; !ok {
			(*propertiesKvs)[k] = make(map[interface{}]bool, 0)
		}
		if len((*propertiesKvs)[k]) < valuesLimit {
			(*propertiesKvs)[k][v] = true
		}
	}
	return nil
}

func IsGroupEventName(eventName string) bool {
	_, exists := GROUP_EVENT_NAME_TO_GROUP_NAME_MAPPING[eventName]
	return exists
}

func GetGroupNameFromGroupEventName(eventName string) string {
	return GROUP_EVENT_NAME_TO_GROUP_NAME_MAPPING[eventName]
}

var groupPropertiesKeyPrefix = map[string]string{
	GROUP_NAME_HUBSPOT_COMPANY:        "$hubspot_company_",
	GROUP_NAME_HUBSPOT_DEAL:           "$hubspot_deal_",
	GROUP_NAME_SALESFORCE_OPPORTUNITY: "$salesforce_opportunity_",
	GROUP_NAME_SALESFORCE_ACCOUNT:     "$salesforce_account_",
}

func GetGroupNameByPropertyName(propertyName string) (string, bool) {
	for groupName, prefix := range groupPropertiesKeyPrefix {
		if strings.HasPrefix(propertyName, prefix) {
			return groupName, true
		}
	}

	return "", false
}

func GetExplainPropertyWeights(propertyName string) float64 {

	prefix_name := map[string]float64{
		"$hubspot":    0.1,
		"$salesforce": 0.1,
		"$marketo":    0.1,
	}

	if val, ok := explainPropertyWeights[propertyName]; ok {
		return val
	}

	for prefix_string, prefix_val := range prefix_name {
		if strings.HasPrefix(propertyName, prefix_string) {
			return prefix_val
		}
	}

	return float64(1)

}

func GetStandardDisplayNameGroups() map[string]string {
	displayNameGroups := make(map[string]string)
	for group := range STANDARD_GROUP_DISPLAY_NAMES {
		displayNameGroups[STANDARD_GROUP_DISPLAY_NAMES[group]] = group
	}
	return displayNameGroups
}

func IsJsonAllowedProperty(k string) bool {
	return k == UP_META_OBJECT_IDENTIFIER_KEY
}

func ValidateAndFillEnrichmentPropsForStringValue(value string, propertyName string, properties *PropertiesMap) {
	if value != "" {
		if c, ok := (*properties)[propertyName]; !ok || c == "" {
			(*properties)[propertyName] = value
		}
	}
}

func ValidateAndFillEnrichmentPropsForIntegerValue(value int, propertyName string, properties *PropertiesMap) {
	if value > 0 {
		if c, ok := (*properties)[propertyName]; !ok || c == "" {
			(*properties)[propertyName] = value
		}
	}
}
