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
PROJECT_ID = 'project_id'
CAMPAIGN_GROUP_ID = 'campaign_group_id'
CAMPAIGN_ID = 'campaign_id'
CREATIVE_ID = 'creative_id'
MAX_LOOKBACK = 30
API_REQUESTS = 'api_requests'
META_COUNT = 100
BACKFILL_DAY = 8
REQUESTED_ROWS_LIMIT = 10000
INSIGHTS_COUNT=10000
BATCH_SIZE = 500
ORG_BATCH_SIZE = 100
BACKFILL_END_DAY = 15
BACKFILL_START_DAY = 35
LINKEDIN_VERSION = '202305'
PROTOCOL_VERSION = '2.0.0'
META_DATA_URL = 'https://api.linkedin.com/rest/adAccounts/{}/{}?q=search&search=(status:(values:List(ACTIVE,PAUSED)))&start={}&count={}'
INSIGHTS_REQUEST_URL_FORMAT = 'https://api.linkedin.com/rest/adAnalytics?q=analytics&pivot={}&dateRange=(start:(day:{},month:{},year:{}),end:(day:{},month:{},year:{}))&timeGranularity=DAILY&fields={}&accounts=List(urn%3Ali%3AsponsoredAccount%3A{})&start={}&count={}'
COMPANY_CAMPAIGN_GROUP_INSIGHTS_REQUEST_URL_FORMAT = 'https://api.linkedin.com/rest/adAnalytics?q=analytics&pivot={}&dateRange=(start:(day:{},month:{},year:{}),end:(day:{},month:{},year:{}))&timeGranularity=ALL&fields={}&accounts=List(urn%3Ali%3AsponsoredAccount%3A{})&campaignGroups=List(urn%3Ali%3AsponsoredCampaignGroup%3A{})&start={}&count={}'
REQUESTED_FIELDS='totalEngagements,impressions,clicks,dateRange,landingPageClicks,costInUsd,leadGenerationMailContactInfoShares,leadGenerationMailInterestedClicks,opens,videoCompletions,videoFirstQuartileCompletions,videoMidpointCompletions,videoThirdQuartileCompletions,videoViews,externalWebsiteConversions,externalWebsitePostClickConversions,externalWebsitePostViewConversions,costInLocalCurrency,conversionValueInLocalCurrency,pivotValues'
ORG_LOOKUP_URL = 'https://api.linkedin.com/rest/organizationsLookup?ids=List({})'
AD_ACCOUNT_URL = 'https://api.linkedin.com/rest/adAccounts/{}'
FETCH_LOG_WITH_DOC_TYPE = 'Fetching {} started for project {} for timestamp {}'
FETCH_CG_LOG_WITH_DOC_TYPE = 'Fetching {} started for camapign group {} for project {} for timestamp {}'
NUM_OF_RECORDS_LOG = 'No of {} records to be inserted for project {} : {}'
NUM_OF_RECORDS_CG_LOG = 'No of {} records for campaign group {} to be inserted for project {} : {}'
API_ERROR_FORMAT = 'Failed to get {} {} from linkedin. StatusCode: {}. Error: {}. Project_id: {}. Ad Account: {}'
DOC_INSERT_ERROR = 'Failed to insert {} {} in database. StatusCode: {}. Error: {}. Project_id: {}. Ad Account: {}. Timestamp: {}'
META_FETCH_START = 'Fetching metadata for {} started for project {}'
INSERTION_LOG = 'Inserting {} {} for timestamp {}'
INSERTION_END_LOG = 'Inserting {} {} ended for timestamp {}'
FINAL_INSERTION_END_LOG = 'Inserting {} {} ended for project {}'
NO_DATA_MEMBER_COMPANY_LOG = 'No data found for member company insights for project {} and Ad account {}'
ACCESS_TOKEN_CHECK_URL = 'https://api.linkedin.com/v2/me?oauth2_access_token={}'
TOKEN_GENERATION_URL = 'https://www.linkedin.com/oauth/v2/accessToken?grant_type=refresh_token&refresh_token={}&client_id={}&client_secret={}'
METRIC_TYPE_INCR = 'incr'
HEALTHCHECK_PING_ID = '837dce09-92ec-4930-80b3-831b295d1a34'
HEALTHCHECK_WEEKLY_JOB = 'da24fcd4-6f09-4f29-9326-72ab73c9affb'
HEALTHCHECK_TOKEN_FAILURE_PING_ID = 'b231cf93-ce7e-4df6-9416-1797c5065c22'
URL_ENDPOINT_CAMPAIGN_GROUP_META = 'adCampaignGroups'
URL_ENDPOINT_CAMPAIGN_META = 'adCampaigns'
URL_ENDPOINT_CREATIVE_META = 'adCreatives'
BACKFILL_NOT_REQUIRED = "Backfill not required for %s for project %s for ad account %s"
AD_ACCOUNT_FAILURE = 'Failed to get ad account metadata from linkedin'
ORG_DATA_FETCH_ERROR = "Failed getting organisation data with error {}"