from util.util import Util as U
from constants.constants import *
class CampaignInfo:
    campaign_info_map = {}

    def __init__(self, campaign_info_map={}) -> None:
        self.campaign_info_map = campaign_info_map

    def get_campaign_data(self):
        return self.campaign_info_map
    
    def get_campaign_info_keys(self):
        return self.campaign_info_map.keys()
    
    def update_campaign_data(self, new_campaign_info_map={}):
        self.campaign_info_map = U.merge_2_dictionaries(
                self.campaign_info_map, new_campaign_info_map)
        
    def reset_campaign_data(self):
        self.campaign_info_map = {}
    