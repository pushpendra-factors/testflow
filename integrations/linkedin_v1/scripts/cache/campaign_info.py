from util.util import Util as U
from constants.constants import *
from data_service.data_service import DataService
class CampaignInfo:
    campaign_info_map = {}
    __instance = None

    @staticmethod
    def get_instance(campaign_info={}):
        if CampaignInfo.__instance == None:
            CampaignInfo(campaign_info)
        return CampaignInfo.__instance

    def __init__(self, campaign_info={}) -> None:
        self.campaign_info_map = campaign_info
        CampaignInfo.__instance = self

    def get_campaign_data(self):
        return self.campaign_info_map
    
    def get_campaign_ids(self):
        return self.campaign_info_map.keys()
    
    def update_campaign_data(self, new_campaign_info_map={}):
        self.campaign_info_map = U.merge_2_dictionaries(
                self.campaign_info_map, new_campaign_info_map)
        
    def reset_campaign_data(self):
        self.campaign_info_map = {}
    
    def get_campaign_info_from_db(self, project_id, ad_account_id, 
                                                    start_timestamp, input_end_timestamp):
        data_service_obj = DataService.get_instance()

        self.campaign_info_map = data_service_obj.get_campaign_data_for_given_timerange(
                                                    project_id, ad_account_id, 
                                                    start_timestamp, input_end_timestamp)