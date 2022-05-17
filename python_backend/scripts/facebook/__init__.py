CONFIG = None
APP_NAME = "facebook_sync"
CAMPAIGN_INSIGHTS = "campaign_insights"
AD_SET_INSIGHTS = "ad_set_insights"
AD_INSIGHTS = "ad_insights"
CAMPAIGN = "campaign"
AD = "ad"
AD_ACCOUNT = "ad_account"
AD_SET = "ad_set"

TIMEZONE_IST = "Asia/Kolkata"

ACCESS_TOKEN = "int_facebook_access_token"
FACEBOOK_AD_ACCOUNT = "int_facebook_ad_account"
TOKEN_EXPIRY = "int_facebook_token_expiry"
DATA = "data"
FACEBOOK = "facebook"
PLATFORM = "platform"
MAX_LOOKBACK = 30
API_REQUESTS = "api_requests"
ERR_MSG = "err_msg"
STATUS = "status"
PAGING = "paging"

PROJECT_ID = "project_id"
CUSTOMER_ACCOUNT_ID = "customer_acc_id"
LAST_TIMESTAMP = "last_timestamp"
TYPE_ALIAS = "type_alias"
INT_FACEBOOK_USER_ID = "int_facebook_user_id"
INT_FACEBOOK_ACCESS_TOKEN = "int_facebook_access_token"
INT_FACEBOOK_EMAIL = "int_facebook_email"

# Tasks - could also represent Workflow if run partial.
EXTRACT = "extract"
LOAD = "load"

DEVELOPMENT = "development"
TEST = "test"
STAGING = "staging"
PRODUCTION = "production"

COST_PER_ACTION_TYPE = 'cost_per_action_type'
WEBSITE_PURCHASE_ROAS = 'website_purchase_roas'
ACTIONS = 'actions'
ACTION_VALUES = 'action_values'

ACTION_TYPE = 'action_type'
VALUE = 'value'


# Workflow
EXTRACT_AND_LOAD_WORKFLOW = "extract_and_load_workflow"
EXTRACT_WORKFLOW = "extract_workflow"
LOAD_WORKFLOW = "load_workflow"

WORKFLOW_TO_TASKS = {
    EXTRACT_AND_LOAD_WORKFLOW: [EXTRACT, LOAD],
    EXTRACT_WORKFLOW: [EXTRACT],
    LOAD_WORKFLOW: [LOAD]
}

STARTED = "started"
COMPlETED = "completed"

ERROR_MESSAGE = "Failed to get {} from facebook. StatusCode: {} Error: {}. Project_id: {}. Customer_account_id: {}"
