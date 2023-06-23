import datetime
import requests
import time
import json
import copy
import logging as log
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
    
    @staticmethod
    def remove_excluded_projects(linkedin_settings, exclude_project_ids):
        required_linkedin_settings = []
        
        for linkedin_int_setting in linkedin_settings:
            if linkedin_int_setting[PROJECT_ID] not in exclude_project_ids:
                required_linkedin_settings.append(linkedin_int_setting)
        
        return required_linkedin_settings

    
    @staticmethod    
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

    
    @staticmethod    
    def get_separated_date(date):
        date = date.split('-')
        return date[0], date[1], date[2]

    
    @staticmethod    
    def get_split_date_from_timestamp(date):
        new_date = datetime.strptime(str(date), '%Y%m%d').date()
        return new_date.year, int(new_date.month), new_date.day

    
    @staticmethod    
    def get_timestamp(date):
        return int(datetime(date['year'],date['month'],date['day']).strftime('%Y%m%d'))

    
    @staticmethod    
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

    
    @staticmethod    
    def sort_by_timestamp(data):
        date = data['dateRange']['end']
        return Util.get_timestamp(date)

    # if last sync date was 20230101, then date start would be 20230102 adn date end would be today - 1
    # this is inclusive of date start and date end
    @staticmethod
    def get_timestamp_range_from_start_end_date(date_start, date_end):
        timestamps =[]
        num_of_days = (date_end-date_start).days + 1
        if num_of_days <=0:
            return []
        for i in range (0, num_of_days):
            date_required = date_start.strftime("%Y%m%d")
            date_start = date_start + timedelta(days=1)
            timestamps.append(date_required)
        
        return timestamps

    @staticmethod    
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
            # date start = last sync info date + 1 day
            date_start = (datetime.strptime(str(sync_info_with_type[doc_type]), 
                                        '%Y-%m-%d') + timedelta(days=1)).date()

        timestamps = Util.get_timestamp_range_from_start_end_date(date_start, date_end)
        
        #if range greater than max lookback, return error msg
        if len(timestamps) > MAX_LOOKBACK:
            return timestamps[-MAX_LOOKBACK:], 'Range exceeding'
        return timestamps, ''
    
    # in case where from and to timestamps are given, 
    # we only consider from and to timestamps, for the combined range
    # we don't take backfill timestamp into consideration
    @staticmethod
    def get_timestamp_ranges_for_company_insights(doc_type, sync_info_with_type, 
                                                end_timestamp, is_backfill_enable_for_project):
        timerange_for_insights, timerange_for_backfill = [], []
        checkBackfill = (end_timestamp == None and 
                        sync_info_with_type['last_backfill_timestamp'] != 0 
                        and is_backfill_enable_for_project)

        
        timerange_for_insights, errMsg = Util.get_timestamp_range(doc_type, sync_info_with_type, 
                                                end_timestamp)
        
        if checkBackfill:
            timerange_for_backfill = Util.get_timestamp_range_to_be_backfilled(
                                        sync_info_with_type['last_backfill_timestamp'])
        
        combined_range = list(set(timerange_for_insights).union(set(timerange_for_backfill)))
        combined_range.sort()
        timestamp_8_days_ago = (datetime.now() - timedelta(days=BACKFILL_DAY)).strftime("%Y%m%d")
        computed_timerange_insights, computed_timerange_backfill = [], []
        for timestamp in combined_range:
            if timestamp <= timestamp_8_days_ago:
                computed_timerange_backfill.append(timestamp)
            else:
                computed_timerange_insights.append(timestamp)
        
        return computed_timerange_insights, computed_timerange_backfill, errMsg

    @staticmethod
    def get_timestamp_range_to_be_backfilled(last_backfill_timestamp):
        backfill_end_date = (datetime.now() - timedelta(days=BACKFILL_DAY)).date()
        backfill_start_date = (datetime.strptime(str(last_backfill_timestamp), '%Y%m%d')).date()

        backfill_timestamps = []
        num_of_days = (backfill_end_date-backfill_start_date).days + 1
        if num_of_days <=0:
            return []
        for i in range (0, num_of_days):
            date_required = backfill_start_date.strftime("%Y%m%d")
            backfill_timestamps.append(date_required)
            backfill_start_date = backfill_start_date + timedelta(days=1)

        if len(backfill_timestamps) > 0:
            return backfill_timestamps
        else:
            return []

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
    
    @staticmethod    
    def generate_and_update_access_token(options, linkedin_settings):
        failures = []
        settings_with_updated_tokens = []
        is_any_token_updated = False
        for setting in linkedin_settings:
            new_access_token, err_msg = setting.generate_access_token(options)
            if err_msg != '':
                failures.append({'status': 'failed', 'errMsg': err_msg,
                                    PROJECT_ID: setting.project_id, 
                                    AD_ACCOUNT: setting.ad_account})
            else:
                setting.access_token = new_access_token
                token_update_response = DataService(options).update_access_token(
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

    def get_batch_of_ids(records):
        mapIDs = {}
        batch_of_ids = []
        len_of_batch = 500
        for data in records:
            id = data['pivotValues'][0].split(':')[3]
            mapIDs[id]= True

        ids_list = list(mapIDs.keys())
        batch_of_ids = [",".join(ids_list[i:i + len_of_batch]) for i in range(0, len(ids_list), len_of_batch)]
        return batch_of_ids
    
    @staticmethod    
    def org_lookup(access_token, ids):
        url = ORG_LOOKUP_URL.format(ids)
        headers = {'Authorization': 'Bearer ' + access_token, 
                    'X-Restli-Protocol-Version': PROTOCOL_VERSION, 'LinkedIn-Version': LINKEDIN_VERSION}
        return requests.get(url, headers=headers)

    def get_failed_ids(ids, map_id_to_org_data):
        ids_list = ids.split(",")
        keys = map_id_to_org_data.keys()
        set_ids = set(ids_list)
        set_keys = set(keys)
        failed_ids = ",".join(set_ids-set_keys)
        return failed_ids


