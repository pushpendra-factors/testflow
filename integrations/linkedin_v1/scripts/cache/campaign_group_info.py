from util.util import Util as U
from constants.constants import *
class CampaignGroupInfo:
    campaign_group_info = {}
    def __init__(self, campaign_group_info={}) -> None:
        self.campaign_group_info = campaign_group_info

    def get_campaign_group_data(self):
        return self.campaign_group_info
    
    def get_campaign_group_ids(self):
        return self.campaign_group_info.keys()
    
    def update_campaign_group_data(self, new_campaign_group_info={}):
        self.campaign_group_info = U.merge_2_dictionaries(
                self.campaign_group_info, new_campaign_group_info)
        
    def reset_campaign_group_data(self):
        self.campaign_group_info = {}

    
    def return_meta_based_on_doc_type(self, doc_type):
        if doc_type == CAMPAIGN_GROUP_INSIGHTS:
            return self.campaign_group_info
        if doc_type == CAMPAIGN_INSIGHTS:
            return self.campaign_map
        if doc_type == CREATIVE_INSIGHTS:
            return self.creative_map