import itertools

from lib.utils.time import TimeUtil
from scripts.facebook import AD_SET_INSIGHTS
from scripts.facebook.tasks.context.extract.base_report_extract import BaseReportExtract as BaseExtractContext

class AdSetPerformanceReportExtract(BaseExtractContext):
    NAME = AD_SET_INSIGHTS
    KEY_FIELDS = ["adset_id"]
    FIELDS = ["account_name", "account_id", "account_currency", "adset_name", "campaign_name",
              "campaign_id", "date_start", "date_stop"]
    METRICS_1 = ["clicks", "cost_per_conversion", "cost_per_ad_click", "cpc", "cpm", "cpp", "ctr",
               "frequency", "impressions", "inline_post_engagement", "social_spend", "spend",
               "inline_link_clicks", "unique_clicks", "reach"]
    METRICS_2 = ["video_p50_watched_actions", "video_p25_watched_actions", "video_30_sec_watched_actions",
               "video_p100_watched_actions", "video_p75_watched_actions"]
    LEVEL_BREAKDOWN = "adset"
    UNFORMATTED_URL = 'https://graph.facebook.com/v13.0/{}/insights?' \
                    'time_range={}&&fields={}&&access_token={}&&level={' \
                    '}&&filtering=[{{\'field\':\'impressions\',\'operator\':\'GREATER_THAN\',\'value\':0}}]&&limit=1000'

    # In place merge of record.
    def merge_records_of_metrics1_and_2(self, records_with_metrics1, records_with_metrics2):
        id_to_records2 = self.get_map_of_id_to_record(records_with_metrics2)
        for record in records_with_metrics1:
            if record["adset_id"] in id_to_records2:
                record.update(id_to_records2[record["adset_id"]])
        return records_with_metrics1

    def get_map_of_id_to_record(self, records):
        result_records = {}
        for record in records:
            result_records[record["adset_id"]] = record
        return result_records
