from lib.utils.json import JsonUtil
from scripts.facebook import CAMPAIGN
from .base_report_load import BaseReportLoad


class CampaignPerformanceLoad(BaseReportLoad):
    campaigns = {}
    FIELDS_TO_BE_ADDED = ["daily_budget", "lifetime_budget", "configured_status", "effective_status",
                          "objective", "buying_type", "bid_strategy"]

    def read_dependencies(self, curr_timestamp):
        self.read_campaign(curr_timestamp)

    def read_current_records(self, curr_timestamp):
        self.add_curr_timestamp(curr_timestamp)

        self.add_source_attributes()
        records_string = self.source.read()
        records = JsonUtil.read(records_string)
        records = self.transform(records)
        self.records = records
        return

    def merge_dependencies_and_current_task_records(self):
        records = self.records
        for record in records:
            required_campaign = self.campaigns[record["campaign_id"]]
            for field in self.FIELDS_TO_BE_ADDED:
                record["campaign_" + field] = required_campaign.get(field)
        self.records = records
        return

    def read_campaign(self, curr_timestamp):
        self.add_curr_timestamp(curr_timestamp)
        self.add_source_with_attributes_for_campaign()
        campaigns_string = self.source.read()
        campaigns = JsonUtil.read(campaigns_string)

        result_campaigns = {}
        for campaign in campaigns:
            result_campaigns[campaign["id"]] = campaign
        self.campaigns = result_campaigns
        return

    def add_source_with_attributes_for_campaign(self):
        self.add_source_attributes_for_type_alias(CAMPAIGN)

    def transform(self, records):
        result_records = []
        for record in records:
            record["id"] = record["campaign_id"]
            record = self.transform_video_attributes(record)
            record = self.transform_action_attributes(record)
            result_records.append(record)
        return result_records
