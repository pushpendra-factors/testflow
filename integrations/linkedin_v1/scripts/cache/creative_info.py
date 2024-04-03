from util.util import Util as U
from constants.constants import *
class CreativeInfo:
    creative_info_map = {}
    __instance = None

    @staticmethod
    def get_instance(creative_info={}):
        if CreativeInfo.__instance == None:
            CreativeInfo(creative_info)
        return CreativeInfo.__instance

    def __init__(self, creative_info={}) -> None:
        self.creative_info_map = creative_info
        CreativeInfo.__instance = self

    def get_creative_data(self):
        return self.creative_info_map
    
    def get_creative_info_keys(self):
        return self.creative_info_map.keys()
    
    def update_creative_data(self, new_creative_info_map={}):
        self.creative_info_map = U.merge_2_dictionaries(
                self.creative_info_map, new_creative_info_map)
        
    def reset_creative_data(self):
        self.creative_info_map = {}