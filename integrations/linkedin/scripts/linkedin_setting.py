from constants import *
import requests

class LinkedinSetting:
    project_id = ''
    ad_account = ''
    access_token = ''
    refresh_token = ''

    def __init__(self, linkedin_setting):
        self.project_id = linkedin_setting[PROJECT_ID]
        self.ad_account = linkedin_setting[LINKEDIN_AD_ACCOUNT]
        self.access_token = linkedin_setting[ACCESS_TOKEN]
        self.refresh_token = linkedin_setting[REFRESH_TOKEN]
    
    
    def validate_access_token(self):
        access_token_check_url = ACCESS_TOKEN_CHECK_URL.format(self.access_token)
        response = requests.get(access_token_check_url)
        if response.ok:
            return True
        return False
    
    def generate_access_token(self, options):
        url = TOKEN_GENERATION_URL.format(self.refresh_token, options.client_id, options.client_secret)
        response = requests.get(url)
        response_json = response.json()
        if response.ok:
            return response_json['access_token'], ''
        return '', response.text