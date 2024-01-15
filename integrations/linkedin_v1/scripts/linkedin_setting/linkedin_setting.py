from constants.constants import *
import requests
import copy
import time
from global_objects.global_obj_creator import data_service_obj

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
    
    def generate_access_token(self, client_id, client_secret):
        url = TOKEN_GENERATION_URL.format(self.refresh_token, 
                            client_id, client_secret)
        response = requests.get(url)
        response_json = response.json()
        if response.ok:
            return response_json['access_token'], ''
        return '', response.text
    
    def remove_excluded_projects(linkedin_settings, exclude_project_ids):
        required_linkedin_settings = []
        
        for linkedin_int_setting in linkedin_settings:
            if linkedin_int_setting[PROJECT_ID] not in exclude_project_ids:
                required_linkedin_settings.append(linkedin_int_setting)
        
        return required_linkedin_settings
 
    def split_settings_for_multiple_ad_accounts(linkedin_settings):
        split_linkedin_settings = []
        failures = []
        for setting in linkedin_settings:
            if setting.ad_account == '':
                failures.append({'status': 'failed', 'errMsg': 'empty ad account',
                                        PROJECT_ID: setting.project_id, 
                                        AD_ACCOUNT: setting.ad_account})
                continue

            # spliting 1 setting into multiple for multiple ad accounts
            ad_accounts =  setting.ad_account.split(',')
            for account_id in ad_accounts:
                new_setting = copy.deepcopy(setting)
                new_setting.ad_account = account_id
                split_linkedin_settings.append(new_setting)
        
        return split_linkedin_settings, failures

    def separate_valid_and_invalid_tokens(linkedin_settings):
        valid_linkedin_settings = []
        invalid_linkedin_settings = []

        for linkedin_int_setting in linkedin_settings:
            linkedin_setting = LinkedinSetting(linkedin_int_setting)

            is_valid_access_token = linkedin_setting.validate_access_token()
            if is_valid_access_token:
                valid_linkedin_settings.append(linkedin_setting)
            else:
                invalid_linkedin_settings.append(linkedin_setting)
        
        return valid_linkedin_settings, invalid_linkedin_settings
    
    def generate_and_update_access_token(options, linkedin_settings):
        failures = []
        settings_with_updated_tokens = []
        is_any_token_updated = False
        for setting in linkedin_settings:
            new_access_token, err_msg = setting.generate_access_token(options.client_id, options.client_secret)
            if err_msg != '':
                failures.append({'status': 'failed', 'errMsg': err_msg,
                                    PROJECT_ID: setting.project_id, 
                                    AD_ACCOUNT: setting.ad_account})
            else:
                setting.access_token = new_access_token
                token_update_response = data_service_obj.update_access_token(
                                    setting.project_id,
                                    setting.access_token)
                if not token_update_response.ok:
                    failures.append({'status': 'failed', 
                                    'errMsg': 'failed to update access token in db',
                                    PROJECT_ID: setting.project_id, 
                                    AD_ACCOUNT: setting.ad_accoun})
                else:
                    settings_with_updated_tokens.append(setting)   
                    is_any_token_updated = True  
        
        if is_any_token_updated:
            time.sleep(600)
        
        return settings_with_updated_tokens, failures
    
       
    
    def perform_token_inspect_and_split_settings(options, linkedin_settings):
        required_linkedin_settings = LinkedinSetting.remove_excluded_projects(
            linkedin_settings, options.exclude_project_ids)
        valid_linkedin_settings, invalid_linkedin_settings = LinkedinSetting.separate_valid_and_invalid_tokens(
            required_linkedin_settings)
        
        settings_with_updated_tokens, token_failures = LinkedinSetting.generate_and_update_access_token(options,
            invalid_linkedin_settings)

        valid_linkedin_settings.extend(settings_with_updated_tokens)
        
        split_linkedin_settings, split_failures = LinkedinSetting.split_settings_for_multiple_ad_accounts(
            valid_linkedin_settings)
        token_failures.extend(split_failures)

        return split_linkedin_settings, token_failures
        
