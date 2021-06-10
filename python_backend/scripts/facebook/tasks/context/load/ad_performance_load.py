from lib.utils.json import JsonUtil
from scripts.facebook import CAMPAIGN, AD_SET, AD
from .base_report_load import BaseReportLoad


# TODO check if we need to add campaignId + adSetId to form the unique key?
class AdPerformanceLoad(BaseReportLoad):
    campaigns = {}
    ad_set = {}
    ads = {}
    records = None
    FIELDS_OF_CAMPAIGN_TO_BE_ADDED = ["daily_budget", "lifetime_budget", "configured_status", "effective_status",
                                      "objective", "buying_type", "bid_strategy"]
    FIELDS_OF_AD_SET_TO_BE_ADDED = ["daily_budget", "lifetime_budget", "configured_status", "effective_status",
                                    "objective", "buying_type", "bid_strategy"]
    FIELDS_OF_AD_TO_BE_ADDED = ["configured_status", "effective_status"]

    def read_dependencies(self, curr_timestamp):
        self.read_campaign(curr_timestamp)
        self.read_ad_set(curr_timestamp)
        self.read_ad(curr_timestamp)

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
            if (record["campaign_id"] not in self.campaigns) or (record["adset_id"] not in self.ad_set) or (record["ad_id"] not in self.ads):
                continue
            required_campaign = self.campaigns[record["campaign_id"]]
            required_ad_set = self.ad_set[record["adset_id"]]
            required_ad = self.ads[record["ad_id"]]
            for field in self.FIELDS_OF_CAMPAIGN_TO_BE_ADDED:
                record["campaign_" + field] = required_campaign.get(field)
            for field in self.FIELDS_OF_AD_SET_TO_BE_ADDED:
                record["ad_set_" + field] = required_ad_set.get(field)
            for field in self.FIELDS_OF_AD_TO_BE_ADDED:
                record["ad_" + field] = required_ad.get(field)
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

    def read_ad_set(self, curr_timestamp):
        self.add_curr_timestamp(curr_timestamp)
        self.add_source_with_attributes_for_ad_set()
        ad_sets_string = self.source.read()
        ad_sets = JsonUtil.read(ad_sets_string)
        result_ad_sets = {}
        for ad_set in ad_sets:
            result_ad_sets[ad_set["id"]] = ad_set
        self.ad_set = result_ad_sets
        return

    def read_ad(self, curr_timestamp):
        self.add_curr_timestamp(curr_timestamp)
        self.add_source_with_attributes_for_ad()
        ad_string = self.source.read()
        ads = JsonUtil.read(ad_string)
        result_ads = {}
        for ad in ads:
            result_ads[ad["id"]] = ad
        self.ads = result_ads
        return

    def add_source_with_attributes_for_campaign(self):
        self.add_source_attributes_for_type_alias(CAMPAIGN)

    def add_source_with_attributes_for_ad_set(self):
        self.add_source_attributes_for_type_alias(AD_SET)

    def add_source_with_attributes_for_ad(self):
        self.add_source_attributes_for_type_alias(AD)

    def transform(self, records):
        result_records = []
        for record in records:
            record["id"] = record["ad_id"]
            record = self.transform_video_attributes(record)
            result_records.append(record)
        return result_records
