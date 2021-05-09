from .base_info_extract import BaseInfoExtract


class AdInfoExtract(BaseInfoExtract):
    NAME = "Ad Info"
    FIELDS = ["id", "name", "account_id", "campaign_id", "adset_id",
              "configured_status", "effective_status"]
