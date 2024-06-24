APP_NAME = 'linkedin_sync'
CAMPAIGN_GROUP_INSIGHTS = 'campaign_group_insights'
CAMPAIGN_INSIGHTS = 'campaign_insights'
CREATIVE_INSIGHTS = 'creative_insights'
MEMBER_COMPANY_INSIGHTS = 'member_company_insights'
CAMPAIGN = 'campaign'
CAMPAIGNS = 'campaign'
CREATIVES = 'creative'
CAMPAIGN_GROUPS = 'campaign_group'
AD_ACCOUNT = 'ad_account'
ACCESS_TOKEN = 'int_linkedin_access_token'
REFRESH_TOKEN = 'int_linkedin_refresh_token'
LINKEDIN_AD_ACCOUNT = 'int_linkedin_ad_account'
ELEMENTS = 'elements'
METADATA = 'metadata'
NEXT_PAGE_TOKEN = 'nextPageToken'
PROJECT_ID = 'project_id'
CAMPAIGN_GROUP_ID = 'campaign_group_id'
CAMPAIGN_ID = 'campaign_id'
CREATIVE_ID = 'creative_id'
MAX_LOOKBACK = 30
API_REQUESTS = 'api_requests'
DATA_STATE_RAW = "raw"
DATA_STATE_TRANSFORMED = "transformed"
META_COUNT = 1000
BACKFILL_DAY = 8
REQUESTED_ROWS_LIMIT = 10000
INSIGHTS_COUNT=10000
BATCH_SIZE = 500
ORG_BATCH_SIZE = 100
T22_END_BUFFER = 2
T22_START_BUFFER = 35
SYNC_STATUS_T22 = 2
SYNC_STATUS_T8 = 1
SYNC_STATUS_T0 = 0
T8_END_BUFFER = 0
SYNC_INFO_KEY_T8 = MEMBER_COMPANY_INSIGHTS + ":1"
SYNC_INFO_KEY_T22 = MEMBER_COMPANY_INSIGHTS + ":2"
LINKEDIN_VERSION = '202405'
PROTOCOL_VERSION = '2.0.0'
META_DATA_URL = 'https://api.linkedin.com/rest/adAccounts/{}/{}?q=search&search=(status:(values:List(ACTIVE,PAUSED)))&pageSize={}'
META_DATA_URL_PAGINATED = 'https://api.linkedin.com/rest/adAccounts/{}/{}?q=search&search=(status:(values:List(ACTIVE,PAUSED)))&pageSize={}&pageToken={}'
INSIGHTS_REQUEST_URL_FORMAT = 'https://api.linkedin.com/rest/adAnalytics?q=analytics&pivot={}&dateRange=(start:(day:{},month:{},year:{}),end:(day:{},month:{},year:{}))&timeGranularity=DAILY&fields={}&accounts=List(urn%3Ali%3AsponsoredAccount%3A{})&start={}&count={}'
COMPANY_CAMPAIGN_GROUP_INSIGHTS_REQUEST_URL_FORMAT = 'https://api.linkedin.com/rest/adAnalytics?q=analytics&pivot={}&dateRange=(start:(day:{},month:{},year:{}),end:(day:{},month:{},year:{}))&timeGranularity=ALL&fields={}&accounts=List(urn%3Ali%3AsponsoredAccount%3A{})&campaignGroups=List(urn%3Ali%3AsponsoredCampaignGroup%3A{})&start={}&count={}'
COMPANY_CAMPAIGN_INSIGHTS_REQUEST_URL_FORMAT = 'https://api.linkedin.com/rest/adAnalytics?q=analytics&pivot={}&dateRange=(start:(day:{},month:{},year:{}),end:(day:{},month:{},year:{}))&timeGranularity=ALL&fields={}&accounts=List(urn%3Ali%3AsponsoredAccount%3A{})&campaigns=List(urn%3Ali%3AsponsoredCampaign%3A{})&start={}&count={}'
REQUESTED_FIELDS='totalEngagements,impressions,clicks,dateRange,landingPageClicks,costInUsd,leadGenerationMailContactInfoShares,leadGenerationMailInterestedClicks,opens,videoCompletions,videoFirstQuartileCompletions,videoMidpointCompletions,videoThirdQuartileCompletions,videoViews,externalWebsiteConversions,externalWebsitePostClickConversions,externalWebsitePostViewConversions,costInLocalCurrency,conversionValueInLocalCurrency,pivotValues'
ORG_LOOKUP_URL = 'https://api.linkedin.com/rest/organizationsLookup?ids=List({})'
AD_ACCOUNT_URL = 'https://api.linkedin.com/rest/adAccounts/{}'
FETCH_LOG_WITH_DOC_TYPE = 'Fetching {} started for project {} ad account {} for timestamp {}'
FETCH_CG_LOG_WITH_DOC_TYPE = 'Fetching {} started for campaign group {} for project {} ad account {} for timestamp {}'
FETCH_C_LOG_WITH_DOC_TYPE = 'Fetching {} started for campaign {} for project {} ad account {} for timestamp {}'
NUM_OF_RECORDS_LOG = 'No of {} records to be inserted for project {} ad account {}: {}'
NUM_OF_RECORDS_CG_LOG = 'No of {} records for campaign group {} to be inserted for project {} ad account {}: {}'
NUM_OF_RECORDS_C_LOG = 'No of {} records for campaign {} to be inserted for project {} ad account {}: {}'
API_ERROR_FORMAT = 'Failed to get {} {} from linkedin. StatusCode: {}. Error: {}. Project_id: {}. Ad Account: {}'
DOC_INSERT_ERROR = 'Failed to insert {} {} in database. StatusCode: {}. Error: {}. Project_id: {}. Ad Account: {}. Timestamp: {}'
META_FETCH_START = 'Fetching metadata for {} started for project {} ad account {}'
INSERTION_LOG = 'Inserting {} {} for timestamp {}'
INSERTION_END_LOG = 'Inserting {} {} ended for timestamp {}'
FINAL_INSERTION_END_LOG = 'Inserting {} {} ended for project {} ad account {}'
NO_DATA_MEMBER_COMPANY_LOG = 'No data found for member company insights for project {} and Ad account {}'
RANGE_EXCEED_LOG = "Range exceeded for project_id {} ad account {} for doc_type {}"
ACCESS_TOKEN_CHECK_URL = 'https://api.linkedin.com/v2/me?oauth2_access_token={}'
TOKEN_GENERATION_URL = 'https://www.linkedin.com/oauth/v2/accessToken?grant_type=refresh_token&refresh_token={}&client_id={}&client_secret={}'
METRIC_TYPE_INCR = 'incr'
HEALTHCHECK_PING_ID = '837dce09-92ec-4930-80b3-831b295d1a34'
HEALTHCHECK_COMPANY_SYNC_JOB = 'da24fcd4-6f09-4f29-9326-72ab73c9affb'
HEALTHCHECK_TOKEN_FAILURE_PING_ID = 'b231cf93-ce7e-4df6-9416-1797c5065c22'
URL_ENDPOINT_CAMPAIGN_GROUP_META = 'adCampaignGroups'
URL_ENDPOINT_CAMPAIGN_META = 'adCampaigns'
URL_ENDPOINT_CREATIVE_META = 'adCreatives'
PIVOT_CAMPAIGN_GROUP = 'CAMPAIGN_GROUP'
PIVOT_CAMPAIGN= 'CAMPAIGN'
PIVOT_MEMBER_COMPANY = 'MEMBER_COMPANY'
BACKFILL_NOT_REQUIRED = "Backfill not required for %s for project %s for ad account %s"
AD_ACCOUNT_FAILURE = 'Failed to get ad account metadata from linkedin'
ORG_DATA_FETCH_ERROR = "Failed getting organisation data with error {}"
SLACK_URL ='https://hooks.slack.com/services/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'
NO_CAMPAIGN_ERR = 'No campaign_data found'