from scripts.facebook import AD_SET_INSIGHTS
from scripts.facebook.tasks.context.extract.base_report_extract import BaseReportExtract as BaseExtractContext


class AdSetPerformanceReportExtract(BaseExtractContext):
    NAME = AD_SET_INSIGHTS
    FIELDS = ["account_name", "account_id", "account_currency", "ad_id", "ad_name", "adset_name", "campaign_name",
              "adset_id", "campaign_id", "actions", "action_values", "date_start", "date_stop"]
    SEGMENTS = ["publisher_platform"]
    METRICS = ["clicks", "conversions", "cost_per_conversion", "cost_per_ad_click", "cpc", "cpm", "cpp", "ctr",
               "frequency", "impressions", "inline_post_engagement", "social_spend", "spend",
               "inline_link_clicks", "unique_clicks", "reach",
               "video_p50_watched_actions", "video_p25_watched_actions", "video_30_sec_watched_actions",
               "video_p100_watched_actions", "video_p75_watched_actions"]
    LEVEL_BREAKDOWN = "adset"
