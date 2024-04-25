
from util.linkedin_api_service import LinkedinApiService
from data_service.data_service import DataService
from google_storage.google_storage import GoogleStorage
from constants.constants import *
from _datetime import datetime
class AdAccountJob:
    linkedin_setting = None
    input_timestamp = None

    def __init__(self, linkedin_setting, input_timestamp) -> None:
        self.linkedin_setting, self.input_timestamp = linkedin_setting, input_timestamp
    
    def execute(self):
        metadata = LinkedinApiService.get_instance().get_ad_account_data(self.linkedin_setting)
        GoogleStorage.get_instance().write(str(metadata), "daily", DATA_STATE_RAW, 
                                           timestamp, self.linkedin_setting.project_id, 
                                           self.linkedin_setting.ad_account, AD_ACCOUNT)

        timestamp = int(datetime.now().strftime('%Y%m%d'))
        if self.input_timestamp != None:
            timestamp = self.input_timestamp
        
        DataService.get_instance().add_linkedin_documents(
                        self.linkedin_setting.project_id, self.linkedin_setting.ad_account,
                        AD_ACCOUNT, str(metadata['id']),
                        metadata, timestamp)