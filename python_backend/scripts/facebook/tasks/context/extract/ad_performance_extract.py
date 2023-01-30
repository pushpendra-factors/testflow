import itertools

from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from lib.utils.json import JsonUtil
from lib.utils.time import TimeUtil
from scripts.facebook import AD_INSIGHTS, ERROR_MESSAGE
from scripts.facebook.tasks.context.extract.base_report_extract import BaseReportExtract as BaseExtractContext
import logging as log

# Splitting adperformance extract into 2 different jobs is possible. But because of backward compatibility and time taken, we are not going ahead.
# Duplicated code across adPerformance and AdsetPerformance
class AdPerformanceReportExtract(BaseExtractContext):
    NAME = AD_INSIGHTS
    KEY_FIELDS = ["ad_id"]
    FIELDS = ["account_name", "account_id", "account_currency", "ad_name", "adset_name", "campaign_name",
              "adset_id", "campaign_id", "date_start", "date_stop"]
    METRICS_1 = ["clicks", "cost_per_conversion", "cost_per_ad_click", "cpc", "cpm", "cpp", "ctr",
               "frequency", "impressions", "inline_post_engagement", "social_spend", "spend", "inline_link_clicks", "unique_clicks", "reach"]
    METRICS_2 = ["video_p50_watched_actions", "video_p25_watched_actions"
               ]
    METRICS_3 = [ "video_30_sec_watched_actions",
               "video_p100_watched_actions", "video_p75_watched_actions"]
    LEVEL_BREAKDOWN = "ad"
    UNFORMATTED_URL = 'https://graph.facebook.com/v15.0/{}/insights?' \
                    'time_range={}&&fields={}&&access_token={}&&level={' \
                    '}&&filtering=[{{\'field\':\'impressions\',\'operator\':\'GREATER_THAN\',\'value\':0}}]&&limit=1000'

    # In place merge of record.
    def merge_records_of_metrics1_and_2(self, records_with_metrics1, records_with_metrics2):
        id_to_records2 = self.get_map_of_id_to_record(records_with_metrics2)
        for record in records_with_metrics1:
            if record["ad_id"] in id_to_records2:
                record.update(id_to_records2[record["ad_id"]])
        return records_with_metrics1

    def get_map_of_id_to_record(self, records):
        result_records = {}
        for record in records:
            result_records[record["ad_id"]] = record
        return result_records
    
    def read_records(self):
        self.add_source_attributes_for_metrics_or_fields(self.METRICS_1)
        resp_status = self.read_records_for_current_columns_and_update_metrics()
        if resp_status != "success":
            return resp_status
        records_with_metrics1 = self.records
        self.add_source_attributes_for_metrics_or_fields(self.METRICS_2)
        resp_status = self.read_records_for_current_columns_and_update_metrics()
        if resp_status != "success":
            return resp_status
        records_with_metrics2 = self.records
        self.add_source_attributes_for_metrics_or_fields(self.METRICS_3)
        resp_status = self.read_records_for_current_columns_and_update_metrics()
        if resp_status != "success":
            return resp_status
        records_with_metrics3 = self.records
        self.add_source_attributes_for_metrics_or_fields(self.FIELDS)
        resp_status = self.read_records_for_current_columns_and_update_metrics()
        if resp_status != "success":
            return resp_status
        records_with_metrics4 = self.records
        self.records = self.merge_records_of_metrics1_and_2(records_with_metrics1, records_with_metrics2)
        self.records = self.merge_records_of_metrics1_and_2(self.records, records_with_metrics3)
        self.records = self.merge_records_of_metrics1_and_2(self.records, records_with_metrics4)
        return "success"

    def add_source_attributes_for_metrics_or_fields(self, metrics):
        url = self.get_url_for_extract(metrics)
        attributes = {"url": url, "access_token": self.int_facebook_access_token}
        self.source.set_attributes(attributes)
        return
    
    def get_url_for_extract(self, metrics):
        curr_timestamp_in_string = TimeUtil.get_string_of_specific_format_from_timestamp(self.curr_timestamp,
                                                                                         '%Y-%m-%d')
        time_range = {'since': curr_timestamp_in_string, 'until': curr_timestamp_in_string}
        url_ = self.UNFORMATTED_URL.format(self.customer_account_id, time_range, self.get_fields_for_extract(metrics),
                                           self.int_facebook_access_token, self.LEVEL_BREAKDOWN)
        return url_

    def get_fields_for_extract(self, metrics):
        return list(itertools.chain(self.KEY_FIELDS, metrics))