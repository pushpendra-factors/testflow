import datetime
import requests
import time
import json
import copy
import logging as log
from datetime import datetime
from constants import *
from _datetime import timedelta
from data_service import DataService
from linkedin_setting import LinkedinSetting

class Util:
    
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
                failures.append({'status': 'failed', 'msg': 'empty ad account',
                             PROJECT_ID: setting.project_id, AD_ACCOUNT: setting.ad_account})
                continue

            # spliting 1 setting into multiple for multiple ad accounts
            ad_accounts =  setting.ad_account.split(',')
            for account_id in ad_accounts:
                new_setting = copy.deepcopy(setting)
                new_setting.ad_account = account_id
                split_linkedin_settings.append(new_setting)
        
        return split_linkedin_settings, failures

    
    def get_separated_date(date):
        date = date.split('-')
        return date[0], date[1], date[2]

    
    def get_split_date_from_timestamp(date):
        new_date = datetime.strptime(str(date), '%Y%m%d').date()
        return new_date.year, int(new_date.month), new_date.day

    
    def get_timestamp(date):
        return int(datetime(date['year'],date['month'],date['day']).strftime('%Y%m%d'))

    
    def ping_healthcheck(env, healthcheck_id, message, endpoint=''):
        message = json.dumps(message, indent=1)
        log.warning('Healthcheck ping for env %s payload %s', env, message)
        if env != 'production': 
            return

        try:
            requests.post('https://hc-ping.com/' + healthcheck_id + endpoint,
                data=message, timeout=10)
        except requests.RequestException as e:
            # Log ping failure here...
            log.error('Ping failed to healthchecks.io: %s' % e)

    
    def sort_by_timestamp(data):
        date = data['dateRange']['end']
        return int(datetime(date['year'],date['month'],date['day']).strftime('%Y%m%d'))

    
    def get_timestamp_range(doc_type, sync_info_with_type, end_timestamp):
        timestamps =[]
        date_start = ''
        date_end = ''

        if end_timestamp != None:
            date_end = datetime.strptime(str(end_timestamp), '%Y%m%d').date()
        else:
            date_end = (datetime.now() - timedelta(days=1)).date()
        
        if doc_type not in sync_info_with_type:
            date_start = (datetime.now() - timedelta(days=MAX_LOOKBACK)).date()
        else:
            date_start = datetime.strptime(str(sync_info_with_type[doc_type]), '%Y-%m-%d').date()

        num_of_days = (date_end-date_start).days
        if num_of_days <=0:
            return [], ''
        for i in range (0, num_of_days):
            date_start = date_start + timedelta(days=1)
            date_required = date_start.strftime("%Y%m%d")
            timestamps.append(date_required)
        
        #if range greater than max lookback, return error msg
        if len(timestamps) > MAX_LOOKBACK:
            return [], 'Range exceeding {} days'.format(MAX_LOOKBACK)
        return timestamps, ''

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
            new_access_token, err_msg = setting.generate_access_token(options)
            if err_msg != '':
                failures.append({'status': 'failed', 'msg': err_msg,
                         PROJECT_ID: setting.project_id, AD_ACCOUNT: setting.ad_account})
            else:
                setting.access_token = new_access_token
                token_update_response = DataService(options).update_access_token(setting.project_id,
                             setting.access_token, options)
                if not token_update_response.ok:
                    failures.append({'status': 'failed', 'msg': 'failed to update access token in db',
                                 PROJECT_ID: setting.project_id, AD_ACCOUNT: setting.ad_accoun})
                else:
                    settings_with_updated_tokens.append(setting)   
                    is_any_token_updated = True  
        
        if is_any_token_updated:
            time.sleep(600)
        
        return settings_with_updated_tokens, failures

    def get_batch_of_ids(records):
        mapIDs = {}
        for data in records:
            id = data['pivotValue'].split(':')[3]
            mapIDs[id]= True

        idStr = ''
        idCount = 0
        idStrArray = []
        for key in mapIDs:
            idCount += 1
            if idStr == '':
                idStr += key
            else:
                idStr += (',' + key)
            if idCount >=500:
                idStrArray.append(idStr)
                idStr = ""
                idCount = 0
        
        if idStr != "":
            idStrArray.append(idStr)
        
        return idStrArray
