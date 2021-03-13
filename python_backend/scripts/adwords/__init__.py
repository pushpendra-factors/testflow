from scripts.stats import EtlRecordsStats

CONFIG = None
APP_NAME = "adwords_sync"
STATUS_FAILED = "failed"
STATUS_SKIPPED = "skipped"
PAGE_SIZE = 200


CUSTOMER_ACCOUNT_PROPERTIES = "customer_account_properties"
CAMPAIGNS = "campaigns"
ADS = "ads"
AD_GROUPS = "ad_groups"
CLICK_PERFORMANCE_REPORT = "click_performance_report"
CAMPAIGN_PERFORMANCE_REPORT = "campaign_performance_report"
AD_PERFORMANCE_REPORT = "ad_performance_report"
AD_GROUP_PERFORMANCE_REPORT = "ad_group_performance_report"
SEARCH_PERFORMANCE_REPORT = "search_performance_report"
KEYWORD_PERFORMANCE_REPORT = "keyword_performance_report"

HEALTHCHECKS_ADWORDS_SYNC_PING_ID = "188cbf7c-0ea1-414b-bf5c-eee47c12a0c8"

etl_record_stats = EtlRecordsStats()
DEVELOPMENT = "development"
TEST = "test"
STAGING = "staging"
PRODUCTION = "production"

ENABLED = "ENABLED"
PAUSED = "PAUSED"
