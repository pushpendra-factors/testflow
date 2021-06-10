import itertools

from lib.utils.time import TimeUtil
from scripts.facebook import *
from scripts.facebook.tasks.context.extract.base_report_extract import BaseReportExtract as BaseExtractContext


class CampaignPerformanceReportExtract(BaseExtractContext):
    NAME = CAMPAIGN_INSIGHTS
    FIELDS = ["account_name", "account_id", "account_currency", "ad_id", "ad_name", "adset_name", "campaign_name",
              "adset_id", "campaign_id", "actions", "action_values", "date_start", "date_stop"]
    SEGMENTS = ["publisher_platform"]
    METRICS = ["clicks", "cost_per_conversion", "cost_per_ad_click", "cpc", "cpm", "cpp", "ctr",
               "frequency", "impressions", "inline_post_engagement", "social_spend", "spend",
               "inline_link_clicks", "unique_clicks", "reach",
               "video_p50_watched_actions", "video_p25_watched_actions", "video_30_sec_watched_actions",
               "video_p100_watched_actions", "video_p75_watched_actions"]
    LEVEL_BREAKDOWN = "campaign"
    UNFORMATTED_URL = 'https://graph.facebook.com/v9.0/{}/insights?' \
                    'action_breakdowns=action_type&&time_range={}&&fields={}&&access_token={}&&level={' \
                    '}&&filtering=[{{\'field\':\'impressions\',\'operator\':\'GREATER_THAN\',\'value\':0}}]&&limit=1000'

    # fields + metrics.
    def get_fields(self):
        return list(itertools.chain(self.FIELDS, self.METRICS))

    def get_url(self):
        curr_timestamp_in_string = TimeUtil.get_string_of_specific_format_from_timestamp(self.curr_timestamp,
                                                                                         '%Y-%m-%d')
        time_range = {'since': curr_timestamp_in_string, 'until': curr_timestamp_in_string}
        url_ = self.UNFORMATTED_URL.format(self.customer_account_id, time_range, self.get_fields(),
                                           self.int_facebook_access_token, self.LEVEL_BREAKDOWN)
        return url_
