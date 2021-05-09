from .base_info_extract import BaseInfoExtract


class CampaignInfoExtract(BaseInfoExtract):
    NAME = "Campaign Info"
    FIELDS = ["id", "name", "account_id", "buying_type", "daily_budget", "lifetime_budget", "configured_status",
              "effective_status", "spend_cap", "start_time", "stop_time", "objective", "bid_strategy"]
