from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from lib.utils.json import JsonUtil
from lib.utils.time import TimeUtil
from scripts.facebook import AD_INSIGHTS, ERROR_MESSAGE
from scripts.facebook.tasks.context.extract.base_report_extract import BaseReportExtract as BaseExtractContext
import logging as log


class AdPerformanceReportExtract(BaseExtractContext):
    NAME = AD_INSIGHTS
    FIELDS = ["account_name", "account_id", "account_currency", "ad_id", "ad_name", "adset_name", "campaign_name",
              "adset_id", "campaign_id", "actions", "action_values", "date_start", "date_stop"]
    SEGMENTS = ["publisher_platform"]
    METRICS = ["clicks", "conversions", "cost_per_conversion", "cost_per_ad_click", "cpc", "cpm", "cpp", "ctr",
               "frequency", "impressions", "inline_post_engagement", "social_spend", "spend",
               "inline_link_clicks", "unique_clicks", "reach",
               "video_p50_watched_actions", "video_p25_watched_actions", "video_30_sec_watched_actions",
               "video_p100_watched_actions", "video_p75_watched_actions"]
    LEVEL_BREAKDOWN = "ad"

    def read_records(self):
        adset_records, response_status = self.get_adset_ids()
        total_records = []
        if response_status == "failed":
            return "failed"
        curr_timestamp_in_string = TimeUtil.get_string_of_specific_format_from_timestamp(self.curr_timestamp,
                                                                                         '%Y-%m-%d')
        time_range = {'since': curr_timestamp_in_string, 'until': curr_timestamp_in_string}
        for adset_record in adset_records:
            adset_id = adset_record["id"]
            current_url = self.UNFORMATTED_URL.format(adset_id, self.get_segments(), time_range,
                                                      self.get_fields(), self.int_facebook_access_token,
                                                      self.LEVEL_BREAKDOWN)
            attributes = {"url": current_url}
            self.source.set_attributes(attributes)

            response_status = super().read_records()
            if response_status == "failed":
                return response_status
            total_records.extend(self.records)
        self.records = total_records
        return "success"

    def get_adset_ids(self):
        unformatted_url = "https://graph.facebook.com/v9.0/{}/{}s?fields={}&&access_token={}&&limit=1000"
        fetch_adids_url = unformatted_url.format(self.customer_account_id, 'adset',
                                                 ["ad_id"], self.int_facebook_access_token)
        attributes = {"url": fetch_adids_url}
        source = self.source
        source.set_attributes(attributes)
        records_string, result_response = source.read()
        if not result_response.ok:
            log.warning(ERROR_MESSAGE.format(self.get_name(), result_response.status_code, result_response.text,
                                             self.project_id))
            MetricsAggregator.update_job_stats(self.project_id, self.customer_account_id,
                                               self.type_alias, "failed", result_response.text)
            return [], "failed"
        return JsonUtil.read(records_string), "success"
