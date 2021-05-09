from .base_info_extract import BaseInfoExtract


class AdSetInfoExtract(BaseInfoExtract):
    NAME = "Ad Set Info"
    FIELDS = ["id", "name", "account_id", "campaign_id", "daily_budget", "lifetime_budget",
              "configured_status", "effective_status", "start_time", "end_time",
              "stop_time", "objective", "bid_strategy"]
