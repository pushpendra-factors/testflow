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
const EVENT_NAME_FORM_SUBMITTED = "$form_submitted"

// Integration: Hubspot event names.
const EVENT_NAME_HUBSPOT_CONTACT_CREATED = "$hubspot_contact_created"
const EVENT_NAME_HUBSPOT_CONTACT_UPDATED = "$hubspot_contact_updated"
const EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED = "$hubspot_deal_state_changed"

// Integration: Salesforce event names.
const EVENT_NAME_SALESFORCE_CONTACT_CREATED = "$sf_contact_created"
const EVENT_NAME_SALESFORCE_CONTACT_UPDATED = "$sf_contact_updated"
const EVENT_NAME_SALESFORCE_LEAD_CREATED = "$sf_lead_created"
const EVENT_NAME_SALESFORCE_LEAD_UPDATED = "$sf_lead_updated"
const EVENT_NAME_SALESFORCE_ACCOUNT_CREATED = "$sf_Account_created"
const EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED = "$sf_Account_updated"
const EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED = "$sf_opportunity_created"
const EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED = "$sf_opportunity_updated"

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
	EVENT_NAME_FORM_SUBMITTED,
	EVENT_NAME_HUBSPOT_CONTACT_CREATED,
	EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
	EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED,
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
}

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
var EP_CAMPAIGN string = "$campaign"
var EP_CAMPAIGN_ID string = "$campaign_id"
var EP_SOURCE string = "$source"
var EP_MEDIUM string = "$medium"
var EP_KEYWORD string = "$keyword"
var EP_KEYWORD_MATCH_TYPE string = "$keyword_match_type"
var EP_CONTENT string = "$content"
var EP_ADGROUP string = "$adgroup"
var EP_ADGROUP_ID string = "$adgroup_id"
var EP_CREATIVE string = "$creative"
var EP_GCLID string = "$gclid"
var EP_FBCLIID string = "$fbclid"
var EP_COST string = "$cost"
var EP_REVENUE string = "$revenue"
var EP_HOUR_OF_DAY string = "$hour_of_day"
var EP_DAY_OF_WEEK string = "$day_of_week"
var EP_SESSION_COUNT string = "$session_count"
var EP_TERM string = "$term"
var EP_CHANNEL string = "$channel" // added at runtime.

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
var UP_CITY string = "$city"
var UP_REGION string = "$region"
var UP_TIMEZONE string = "$timezone"
var UP_SEGMENT_CHANNEL string = "$segment_channel" // from segement (browser, client, etc.,).
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
var UP_INITIAL_CONTENT string = "$initial_content"
var UP_INITIAL_ADGROUP string = "$initial_adgroup"
var UP_INITIAL_ADGROUP_ID string = "$initial_adgroup_id"
var UP_INITIAL_CREATIVE string = "$initial_creative"
var UP_INITIAL_GCLID string = "$initial_gclid"
var UP_INITIAL_FBCLID string = "$initial_fbclid"
var UP_INITIAL_COST string = "$initial_cost"
var UP_TOTAL_COST string = "$total_cost"
var UP_INITIAL_REVENUE string = "$initial_revenue"
var UP_TOTAL_REVENUE string = "$total_revenue"
var UP_INITIAL_REFERRER string = "$initial_referrer"
var UP_INITIAL_REFERRER_URL string = "$initial_referrer_url"
var UP_INITIAL_REFERRER_DOMAIN string = "$initial_referrer_domain"
var UP_DAY_OF_FIRST_EVENT string = "$day_of_first_event"
var UP_HOUR_OF_FIRST_EVENT string = "$hour_of_first_event"
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
var UP_LATEST_CONTENT string = "$latest_content"
var UP_LATEST_ADGROUP string = "$latest_adgroup"
var UP_LATEST_ADGROUP_ID string = "$latest_adgroup_id"
var UP_LATEST_CREATIVE string = "$latest_creative"
var UP_LATEST_GCLID string = "$latest_gclid"
var UP_LATEST_FBCLID string = "$latest_fbclid"
var UP_LATEST_COST string = "$latest_cost"
var UP_LATEST_REVENUE string = "$latest_revenue"
var UP_LATEST_REFERRER string = "$latest_referrer"
var UP_LATEST_REFERRER_URL string = "$latest_referrer_url"
var UP_LATEST_REFERRER_DOMAIN string = "$latest_referrer_domain"

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
var SP_INITIAL_COST string = UP_INITIAL_COST
var SP_INITIAL_REVENUE string = UP_INITIAL_REVENUE

var SDK_ALLOWED_EVENT_PROPERTIES = [...]string{
	EP_INTERNAL_IP,
	EP_LOCATION_LATITUDE,
	EP_LOCATION_LONGITUDE,
	EP_SKIP_SESSION,
	EP_SEGMENT_EVENT_VERSION,
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
	EP_CONTENT,
	EP_ADGROUP,
	EP_ADGROUP_ID,
	EP_CREATIVE,
	EP_GCLID,
	EP_FBCLIID,
	EP_COST,
	EP_REVENUE,
	EP_TERM,

	// user_properties captured on event. i.e form_submit.
	UP_EMAIL,
	UP_PHONE,
	UP_COMPANY,
	UP_NAME,
	UP_FIRST_NAME,
	UP_LAST_NAME,
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
	UP_SEGMENT_CHANNEL,
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
	UP_REGION,
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
	UP_INITIAL_CONTENT,
	UP_INITIAL_ADGROUP,
	UP_INITIAL_ADGROUP_ID,
	UP_INITIAL_CREATIVE,
	UP_INITIAL_GCLID,
	UP_INITIAL_FBCLID,
	UP_INITIAL_COST,
	UP_TOTAL_COST,
	UP_INITIAL_REVENUE,
	UP_TOTAL_REVENUE,
	UP_DAY_OF_FIRST_EVENT,
	UP_HOUR_OF_FIRST_EVENT,
	UP_LATEST_PAGE_URL,
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
	UP_LATEST_CONTENT,
	UP_LATEST_ADGROUP,
	UP_LATEST_ADGROUP_ID,
	UP_LATEST_CREATIVE,
	UP_LATEST_GCLID,
	UP_LATEST_FBCLID,
	UP_LATEST_COST,
	UP_LATEST_REVENUE,
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

const (
	SMART_EVENT_SALESFORCE_PREV_PROPERTY = "$prev_salesforce_"
	SMART_EVENT_SALESFORCE_CURR_PROPERTY = "$curr_salesforce_"
	SMART_EVENT_HUBSPOT_PREV_PROPERTY    = "$prev_hubspot_"
	SMART_EVENT_HUBSPOT_CURR_PROPERTY    = "$curr_hubspot_"
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
	UP_INITIAL_COST,
	UP_TOTAL_COST,
	UP_INITIAL_REVENUE,
	UP_TOTAL_REVENUE,
	UP_SCREEN_WIDTH,
	UP_SCREEN_HEIGHT,
	UP_SCREEN_DENSITY,
	UP_SESSION_COUNT,
	EP_SESSION_COUNT,
	UP_PAGE_COUNT,
	UP_TOTAL_SPENT_TIME,
	UP_LATEST_PAGE_LOAD_TIME,
	UP_LATEST_PAGE_SPENT_TIME,
}
var CATEGORICAL_PROPERTY_BY_NAME = [...]string{
	EP_CAMPAIGN_ID,
	EP_ADGROUP_ID,
	EP_IS_PAGE_VIEW,
	UP_INITIAL_ADGROUP_ID,
	UP_INITIAL_CAMPAIGN_ID,
	SP_IS_FIRST_SESSION,
}

var DATETIME_PROPERTY_BY_NAME = [...]string{
	UP_JOIN_TIME,
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
	EP_CONTENT:             UP_INITIAL_CONTENT,
	EP_ADGROUP:             UP_INITIAL_ADGROUP,
	EP_ADGROUP_ID:          UP_INITIAL_ADGROUP_ID,
	EP_CREATIVE:            UP_INITIAL_CREATIVE,
	EP_GCLID:               UP_INITIAL_GCLID,
	EP_FBCLIID:             UP_INITIAL_FBCLID,
	EP_COST:                UP_INITIAL_COST,
	EP_REVENUE:             UP_INITIAL_REVENUE,
}

var EVENT_TO_USER_LATEST_PROPERTIES = map[string]string{
	EP_PAGE_URL:            UP_LATEST_PAGE_URL,
	EP_PAGE_RAW_URL:        UP_LATEST_PAGE_RAW_URL,
	EP_PAGE_DOMAIN:         UP_LATEST_PAGE_DOMAIN,
	EP_REFERRER_URL:        UP_LATEST_REFERRER_URL,
	EP_REFERRER_DOMAIN:     UP_LATEST_REFERRER_DOMAIN,
	EP_REFERRER:            UP_LATEST_REFERRER,
	EP_PAGE_LOAD_TIME:      UP_LATEST_PAGE_LOAD_TIME,
	EP_PAGE_SPENT_TIME:     UP_LATEST_PAGE_SPENT_TIME,
	EP_PAGE_SCROLL_PERCENT: UP_LATEST_PAGE_SCROLL_PERCENT,
	EP_CAMPAIGN:            UP_LATEST_CAMPAIGN,
	EP_CAMPAIGN_ID:         UP_LATEST_CAMPAIGN_ID,
	EP_SOURCE:              UP_LATEST_SOURCE,
	EP_MEDIUM:              UP_LATEST_MEDIUM,
	EP_KEYWORD:             UP_LATEST_KEYWORD,
	EP_KEYWORD_MATCH_TYPE:  UP_LATEST_KEYWORD_MATCH_TYPE,
	EP_CONTENT:             UP_LATEST_CONTENT,
	EP_ADGROUP:             UP_LATEST_ADGROUP,
	EP_ADGROUP_ID:          UP_LATEST_ADGROUP_ID,
	EP_CREATIVE:            UP_LATEST_CREATIVE,
	EP_GCLID:               UP_LATEST_GCLID,
	EP_FBCLIID:             UP_LATEST_FBCLID,
	EP_COST:                UP_LATEST_COST,
	EP_REVENUE:             UP_LATEST_REVENUE,
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
	UP_REGION,
	UP_TIMEZONE,
	UP_TOTAL_COST,
	UP_TOTAL_REVENUE,
}

var EVENT_TO_SESSION_PROPERTIES = map[string]string{
	EP_PAGE_URL:            SP_INITIAL_PAGE_URL,
	EP_PAGE_RAW_URL:        SP_INITIAL_PAGE_RAW_URL,
	EP_PAGE_DOMAIN:         SP_INITIAL_PAGE_DOMAIN,
	EP_PAGE_LOAD_TIME:      SP_INITIAL_PAGE_LOAD_TIME,
	EP_PAGE_SPENT_TIME:     SP_INITIAL_PAGE_SPENT_TIME,
	EP_PAGE_SCROLL_PERCENT: SP_INITIAL_PAGE_SCROLL_PERCENT,
	EP_COST:                SP_INITIAL_COST,
	EP_REVENUE:             SP_INITIAL_REVENUE,

	// Uses same name as event properties.
	EP_CAMPAIGN:           EP_CAMPAIGN,
	EP_CAMPAIGN_ID:        EP_CAMPAIGN_ID,
	EP_SOURCE:             EP_SOURCE,
	EP_MEDIUM:             EP_MEDIUM,
	EP_KEYWORD:            EP_KEYWORD,
	EP_KEYWORD_MATCH_TYPE: EP_KEYWORD_MATCH_TYPE,
	EP_CONTENT:            EP_CONTENT,
	EP_ADGROUP:            EP_ADGROUP,
	EP_ADGROUP_ID:         EP_ADGROUP_ID,
	EP_CREATIVE:           EP_CREATIVE,
	EP_GCLID:              EP_GCLID,
	EP_FBCLIID:            EP_FBCLIID,

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
	EP_TERM,
	EP_KEYWORD_MATCH_TYPE,
	EP_CONTENT,
	EP_ADGROUP,
	EP_ADGROUP_ID,
	EP_CREATIVE,
	EP_GCLID,
	EP_FBCLIID,
}

var PREDEFINED_BIN_RANGES_FOR_PROPERTY = map[string][][2]float64{
	EP_PAGE_LOAD_TIME: [][2]float64{
		//[2]float64{0, 1},
		[2]float64{1, 2},
		[2]float64{2, 5},
		[2]float64{5, 10},
		[2]float64{10, 20},
		[2]float64{20, math.MaxFloat64},
	},
	UP_INITIAL_PAGE_LOAD_TIME: [][2]float64{
		//[2]float64{0, 1},
		[2]float64{1, 2},
		[2]float64{2, 5},
		[2]float64{5, 10},
		[2]float64{10, 20},
		[2]float64{20, math.MaxFloat64},
	},
	EP_PAGE_SPENT_TIME: [][2]float64{
		[2]float64{0, 30},
		[2]float64{30, 120},
		[2]float64{120, 300},
		[2]float64{300, 600},
		[2]float64{600, 1800},
		[2]float64{1800, math.MaxFloat64},
	},
	UP_INITIAL_PAGE_SPENT_TIME: [][2]float64{
		[2]float64{0, 30},
		[2]float64{30, 120},
		[2]float64{120, 300},
		[2]float64{300, 600},
		[2]float64{600, 1800},
		[2]float64{1800, math.MaxFloat64},
	},
	EP_PAGE_SCROLL_PERCENT: [][2]float64{
		[2]float64{0, 10},
		[2]float64{10, 30},
		[2]float64{30, 50},
		[2]float64{50, 80},
		[2]float64{80, 100},
	},
	UP_INITIAL_PAGE_SCROLL_PERCENT: [][2]float64{
		[2]float64{0, 10},
		[2]float64{10, 30},
		[2]float64{30, 50},
		[2]float64{50, 80},
		[2]float64{80, 100},
	},
}

// DISABLED_CORE_QUERY_USER_PROPERTIES Less important user properties in core query context.
var DISABLED_CORE_QUERY_USER_PROPERTIES = [...]string{
	UP_DEVICE_ADTRACKING_ENABLED,
	UP_NETWORK_BLUETOOTH,
	UP_NETWORK_CARRIER,
	UP_NETWORK_CELLULAR,
	UP_NETWORK_WIFI,
	UP_SEGMENT_CHANNEL,
	UP_DEVICE_ADVERTISING_ID,
	UP_DEVICE_ID,
	UP_MERGE_TIMESTAMP,
	UP_INITIAL_PAGE_EVENT_ID,
	UP_META_OBJECT_IDENTIFIER_KEY,
	EP_CRM_REFERENCE_EVENT_ID,
}

// DISABLED_CORE_QUERY_EVENT_PROPERTIES Less important event properties in core query context.
var DISABLED_CORE_QUERY_EVENT_PROPERTIES = [...]string{
	EP_INTERNAL_IP,
	EP_LOCATION_LATITUDE,
	EP_LOCATION_LONGITUDE,
	EP_SEGMENT_EVENT_VERSION,
	EP_CRM_REFERENCE_EVENT_ID,
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
	UP_SEGMENT_CHANNEL,
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

// WHITELIST_FACTORS_USER_PROPERTIES USER properties enabled for the factors analysis.
var WHITELIST_FACTORS_USER_PROPERTIES = [...]string{
	UP_BROWSER,
	UP_COUNTRY,
	UP_JOIN_TIME,
	UP_OS,
}
var WHITELIST_FACTORS_EVENT_PROPERTIES = [...]string{
	EP_PAGE_LOAD_TIME,
	EP_PAGE_SPENT_TIME,
	EP_PAGE_SCROLL_PERCENT,
	EP_REVENUE,
	EP_COST,
	EP_CAMPAIGN,
}

// DISABLED_FACTORS_EVENT_PROPERTIES Event properties disabled for the factors analysis.
var DISABLED_FACTORS_EVENT_PROPERTIES = [...]string{
	EP_INTERNAL_IP,
	EP_LOCATION_LATITUDE,
	EP_LOCATION_LONGITUDE,
	EP_SEGMENT_EVENT_VERSION,
	EP_PAGE_RAW_URL,
	EP_REFERRER,
	EP_GCLID,
	EP_FBCLIID,
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
	EP_PAGE_URL:             true,
	EP_PAGE_RAW_URL:         true,
	EP_PAGE_TITLE:           true,
	EP_DAY_OF_WEEK:          true,
	EP_HOUR_OF_DAY:          true,
	// Temporary fix.
	EP_REFERRER:                    true,
	EP_REFERRER_URL:                true,
	EP_REFERRER_DOMAIN:             true,
	SP_INITIAL_REFERRER_DOMAIN:     true,
	SP_INITIAL_REFERRER:            true,
	EP_PAGE_LOAD_TIME:              true,
	UP_INITIAL_PAGE_SPENT_TIME:     true,
	UP_INITIAL_PAGE_SCROLL_PERCENT: true,

	// Session Latest.
	SP_LATEST_PAGE_RAW_URL: true,
	SP_LATEST_PAGE_URL:     true,

	// Counts being seen as categorical.
	UP_PAGE_COUNT:       true,
	SP_PAGE_COUNT:       true,
	UP_SESSION_COUNT:    true,
	EP_SESSION_COUNT:    true,
	SP_SESSION_TIME:     true,
	SP_SPENT_TIME:       true,
	UP_TOTAL_SPENT_TIME: true,
}

var ITREE_NUMERICAL_PROPERTIES_TO_IGNORE = map[string]bool{
	"$campaign":         true,
	"$initial_campaign": true,
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
	UP_INITIAL_COST,
	UP_INITIAL_CREATIVE,
	UP_INITIAL_FBCLID,
	UP_INITIAL_GCLID,
	UP_INITIAL_KEYWORD,
	UP_INITIAL_KEYWORD_MATCH_TYPE,
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
	UP_INITIAL_REVENUE,
	UP_INITIAL_SOURCE,
	UP_JOIN_TIME,
}

var USER_PROPERTIES_MERGE_TYPE_ADD = [...]string{
	UP_PAGE_COUNT,
	UP_SESSION_COUNT,
	UP_TOTAL_SPENT_TIME,
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

// isValidProperty - Validate property type.
func isPropertyTypeValid(value interface{}) error {
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
				!strings.HasPrefix(k, HUBSPOT_PROPERTY_PREFIX) &&
				!strings.HasPrefix(k, SALESFORCE_PROPERTY_PREFIX) &&
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

func GetValidatedEventProperties(properties *PropertiesMap) *PropertiesMap {
	validatedProperties := make(PropertiesMap)
	for k, v := range *properties {
		if err := isPropertyTypeValid(v); err == nil {
			var propertyKey string
			// Escape properties with $ prefix but allow query_params_props
			// with selected prefixes starting with $ and default properties.
			if strings.HasPrefix(k, NAME_PREFIX) &&
				!strings.HasPrefix(k, QUERY_PARAM_PROPERTY_PREFIX) &&
				!strings.HasPrefix(k, HUBSPOT_PROPERTY_PREFIX) &&
				!strings.HasPrefix(k, SALESFORCE_PROPERTY_PREFIX) &&
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
				validatedProperties[propertyKey] = v
			}
		}
	}
	return &validatedProperties
}

func UnEscapeQueryParamProperties(properties *PropertiesMap) {
	for k := range *properties {
		if strings.HasPrefix(k, QUERY_PARAM_PROPERTY_PREFIX) {
			(*properties)[k] = GetUnEscapedPropertyValue((*properties)[k])
		}
	}
}

func MapEventPropertiesToDefinedProperties(properties *PropertiesMap) (*PropertiesMap, bool) {
	mappedProperties := make(PropertiesMap)

	for k, v := range *properties {
		var property string
		switch k {
		case QUERY_PARAM_UTM_PREFIX + "campaign", QUERY_PARAM_UTM_PREFIX + "campaign_name":
			property = EP_CAMPAIGN
		case QUERY_PARAM_UTM_PREFIX + "campaignid", QUERY_PARAM_UTM_PREFIX + "campaign_id":
			property = EP_CAMPAIGN_ID
		case QUERY_PARAM_UTM_PREFIX + "source":
			property = EP_SOURCE
		case QUERY_PARAM_UTM_PREFIX + "medium":
			property = EP_MEDIUM
		case QUERY_PARAM_UTM_PREFIX + "keyword", QUERY_PARAM_UTM_PREFIX + "key_word":
			property = EP_KEYWORD
		case QUERY_PARAM_UTM_PREFIX + "term":
			property = EP_TERM
		case QUERY_PARAM_UTM_PREFIX + "matchtype", QUERY_PARAM_UTM_PREFIX + "match_type":
			property = EP_KEYWORD_MATCH_TYPE
		case QUERY_PARAM_UTM_PREFIX + "content":
			property = EP_CONTENT
		case QUERY_PARAM_UTM_PREFIX + "adgroup", QUERY_PARAM_UTM_PREFIX + "ad_group":
			property = EP_ADGROUP
		case QUERY_PARAM_UTM_PREFIX + "adgroupid", QUERY_PARAM_UTM_PREFIX + "adgroup_id", QUERY_PARAM_UTM_PREFIX + "ad_group_id":
			property = EP_ADGROUP_ID
		case QUERY_PARAM_UTM_PREFIX + "creative", QUERY_PARAM_UTM_PREFIX + "creative_id", QUERY_PARAM_UTM_PREFIX + "creativeid":
			property = EP_CREATIVE
		case QUERY_PARAM_PROPERTY_PREFIX + "gclid":
			property = EP_GCLID
		case QUERY_PARAM_PROPERTY_PREFIX + "fbclid":
			property = EP_FBCLIID
		default:
			property = k
		}

		mappedProperties[property] = v
	}

	return &mappedProperties, HasDefinedMarketingProperty(&mappedProperties)
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

func GetPropertyTypeByKeyValue(propertyKey string, propertyValue interface{}) string {
	// PropertyKey will be set to null if the pre-mentioned classfication behaviour need to be supressed
	if propertyKey != "" {
		if strings.HasPrefix(propertyKey, NAME_PREFIX) {
			if isNumericalPropertyByName(propertyKey) {
				return PropertyTypeNumerical
			}
			if isCategoricalPropertyByName(propertyKey) {
				return PropertyTypeCategorical
			}
			if isDateTimePropertyByName(propertyKey) {
				return PropertyTypeDateTime
			}
		}
		if IsPropertyNameContainsDateOrTime(propertyKey) {
			_, status := ConvertDateTimeValueToNumber(propertyValue)
			if status == true {
				return PropertyTypeDateTime
			}
		}
	}

	switch propertyValue.(type) {
	case int, float64:
		return PropertyTypeNumerical
	case string:
		return PropertyTypeCategorical
	default:
		return PropertyTypeUnknown
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
// user_properites based on the update allowed event_properties.
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

// Add day_of_week and hour_of_day event property
func FillHourAndDayEventProperty(properties *postgres.Jsonb, timestamp int64) (*postgres.Jsonb, error) {
	unixTimeUTC := time.Unix(timestamp, 0)
	weekDay := unixTimeUTC.Weekday().String()
	hr, _, _ := unixTimeUTC.Clock()
	eventPropsJSON, err := DecodePostgresJsonb(properties)
	if err != nil {
		return nil, err
	}
	(*eventPropsJSON)[EP_DAY_OF_WEEK] = weekDay
	(*eventPropsJSON)[EP_HOUR_OF_DAY] = hr
	return EncodeToPostgresJsonb(eventPropsJSON)
}

// ClassifyPropertiesType - Classifies type of properties as categorical and numerical
// properties -> map[propertyKey]map[propertyValue]true
func ClassifyPropertiesType(properties *map[string]map[interface{}]bool) (map[string][]string, error) {
	numProperties := make([]string, 0, 0)
	catProperties := make([]string, 0, 0)

	for propertyKey, v := range *properties {
		isNumericalProperty := true
		for propertyValue := range v {
			propertyType := GetPropertyTypeByKeyValue(propertyKey, propertyValue)
			switch propertyType {
			case PropertyTypeNumerical:
			case PropertyTypeCategorical:
				isNumericalProperty = false
			default:
				return nil, fmt.Errorf("unsupported type %s on property type classification %s - %s", propertyType, propertyKey, propertyValue)
			}
		}

		if isNumericalProperty {
			numProperties = append(numProperties, propertyKey)
		} else {
			catProperties = append(catProperties, propertyKey)
		}
	}

	propsByType := make(map[string][]string, 0)
	propsByType[PropertyTypeNumerical] = numProperties
	propsByType[PropertyTypeCategorical] = catProperties

	return propsByType, nil
}

// Moves datetime properties from numerical properties to type datetime.
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

// Fills default user propertie which should be present on properties list always.
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
		log.Error("Invalid value type on GetPropertyValueAsString : ", valueType)
		return ""
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
func FilterDisabledCoreUserProperties(propertiesByType *map[string][]string) {
	for propertyType, properties := range *propertiesByType {
		(*propertiesByType)[propertyType] = StringSliceDiff(properties, DISABLED_CORE_QUERY_USER_PROPERTIES[:])
	}
}

// FilterDisabledCoreEventProperties Filters out less important properties from the list.
func FilterDisabledCoreEventProperties(propertiesByType *map[string][]string) {
	for propertyType, properties := range *propertiesByType {
		(*propertiesByType)[propertyType] = StringSliceDiff(properties, DISABLED_CORE_QUERY_EVENT_PROPERTIES[:])
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
	}

	return strings.HasSuffix(property, "url") ||
		StringValueIn(property, propertiesWithoutURLSuffix)
}

func SanitizeProperties(properties *PropertiesMap) {
	for k, v := range *properties {
		if v == nil {
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

// SortByTimestampAndCount Sorts the given array by timestamp/count
// Pick all past 24 hours event and sort the remaining by count and return
// No filtering is done in this method
func SortByTimestampAndCount(data []NameCountTimestampCategory) []NameCountTimestampCategory {

	smartEventNames := make([]NameCountTimestampCategory, 0)
	sorted := make([]NameCountTimestampCategory, 0)
	trimmed := make([]NameCountTimestampCategory, 0)
	currentDate := time.Now().UTC()

	sort.Slice(data, func(i, j int) bool {
		return data[i].Count > data[j].Count
	})

	for _, details := range data {
		hoursBeforeLastSeen := currentDate.Sub(time.Unix(details.Timestamp, 0)).Hours()
		if hoursBeforeLastSeen <= float64(24) {
			details.GroupName = MostRecent
			if details.Category == SmartEvent {
				smartEventNames = append(smartEventNames, details)
				continue
			}
			sorted = append(sorted, details)
		} else {
			details.GroupName = FrequentlySeen
			trimmed = append(trimmed, details)
		}
	}

	sorted = append(smartEventNames, sorted...)
	for _, data := range trimmed {
		sorted = append(sorted, data)
	}
	return sorted
}

//AggregatePropertyValuesAcrossDate values are stored by date and this method aggregates the count and last seen value and returns
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

//AggregatePropertyAcrossDate values are stored by date and this method aggregates the count and last seen value and returns
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
		if value.ValueType == "number" {
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
