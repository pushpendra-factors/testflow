from constants.constants import *
from .base_heirarchical_job import BaseHeirarchicalJob

class CampaignGroupJob(BaseHeirarchicalJob):
    meta_doc_type = CAMPAIGN_GROUPS
    insights_doc_type = CAMPAIGN_GROUP_INSIGHTS
    meta_endpoint = URL_ENDPOINT_CAMPAIGN_GROUP_META
    pivot_insights = PIVOT_CAMPAIGN_GROUP

    def __init__(self, linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp):
        super().__init__(self.meta_doc_type, self.insights_doc_type, self.meta_endpoint, self.pivot_insights,
                        linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp)

