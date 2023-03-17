from optparse import OptionParser
import logging as log
import datetime
import requests
import copy
from datetime import datetime
import sys
import time
import traceback
from data_service import DataService
from constants import *
from util import Util as U
from data_fetch import DataFetch
from transformations import DataTransformation
from data_insert import DataInsert
from linkedin_setting import LinkedinSetting

parser = OptionParser()
parser.add_option('--env', dest='env', default='development')
parser.add_option('--dry', dest='dry', help='', default='False')
parser.add_option('--skip_today', dest='skip_today', help='', default='False') 
parser.add_option('--project_ids', dest='project_ids', help='', default=None, type=str)
parser.add_option('--exclude_project_ids', dest='exclude_project_ids', help='', default='', type=str)
parser.add_option('--client_id', dest='client_id', help='',default=None, type=str)
parser.add_option('--client_secret', dest='client_secret', help='',default=None, type=str)
parser.add_option('--start_timestamp', dest='start_timestamp', help='', default=None, type=int)
parser.add_option('--end_timestamp', dest='end_timestamp', help='', default=None, type=int)
parser.add_option('--insert_metadata', dest='insert_metadata', help='', default='True')
parser.add_option('--insert_report', dest='insert_report', help='', default='True')
parser.add_option('--run_ads_heirarchical_data', dest='run_ads_heirarchical_data', help='', default='True' )
parser.add_option('--data_service_host', dest='data_service_host',
    help='Data service host', default='http://localhost:8089')
parser.add_option('--run_member_company_insights', dest='run_member_company_insights', help='', default='True')

def ping_healthcheck(successes, failures, token_failures):
        status_msg = ''
        if len(failures) > 0: status_msg = 'Failures on sync.'
        else: status_msg = 'Successfully synced.'
        notification_payload = {
            'status': status_msg, 
            'failures': failures, 
            'success': successes,
        }

        if len(failures) > 0:
            U.ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload, endpoint='/fail')
        else:
            U.ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload)
        if len(token_failures) > 0:
            notification_payload = {
                'status': 'Token failures', 
                'failures': token_failures,
            }
            U.ping_healthcheck(options.env, HEALTHCHECK_TOKEN_FAILURE_PING_ID, notification_payload, endpoint='/fail')


def get_collections(options, linkedin_setting, sync_info_with_type, end_timestamp):
    response = {'status': 'success'}
    status = ''
    errMsgs = []
    skipMsgs = []
    campaign_group_meta = {}
    campaign_meta = {}
    creative_meta = {}
    requests_counter = 0
    run_member_company_insights = (options.run_member_company_insights == 'True' or options.run_member_company_insights == True)
    run_ads_heirarchical_data = (options.run_ads_heirarchical_data == 'True' or options.run_ads_heirarchical_data == True)

    try:
        if run_ads_heirarchical_data:
            res = DataFetch.get_ad_account_data(options, linkedin_setting, end_timestamp)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])

            # don't mutate meta object, return it as a new object from get_campaign_group_function
            res = DataFetch.get_ads_hierarchical_data(options, linkedin_setting,
                                                 sync_info_with_type, campaign_group_meta,
                                                     campaign_meta, creative_meta,
                                                         CAMPAIGN_GROUPS, CAMPAIGN_GROUP_INSIGHTS,
                                                            URL_ENDPOINT_CAMPAIGN_GROUP_META,
                                                                 'CAMPAIGN_GROUP', end_timestamp)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])
            if res['status'] == 'skipped':
                skipMsgs.append(res['errMsg'])
            
            res = DataFetch.get_ads_hierarchical_data(options, linkedin_setting,
                                             sync_info_with_type, campaign_group_meta,
                                                campaign_meta, creative_meta,
                                                    CAMPAIGNS, CAMPAIGN_INSIGHTS,
                                                        URL_ENDPOINT_CAMPAIGN_META,
                                                        'CAMPAIGN', end_timestamp)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])
            if res['status'] == 'skipped':
                skipMsgs.append(res['errMsg'])
            
            res = DataFetch.get_ads_hierarchical_data(options, linkedin_setting,
                                                     sync_info_with_type, campaign_group_meta,
                                                        campaign_meta, creative_meta, CREATIVES,
                                                             CREATIVE_INSIGHTS, URL_ENDPOINT_CREATIVE_META,
                                                                 'CREATIVE', end_timestamp)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])
            if res['status'] == 'skipped':
                skipMsgs.append(res['errMsg'])
        
        if run_member_company_insights:
            res = DataFetch.get_member_company_data(options, linkedin_setting,
                             sync_info_with_type, end_timestamp)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])
            if res['status'] == 'skipped':
                skipMsgs.append(res['errMsg'])
        
    except Exception as e:
        traceback.print_tb(e.__traceback__)
        response['status'] = 'failed'
        response['msg'] = 'Failed with exception '+str(e)
        response[API_REQUESTS]= requests_counter
        return response
    if status == 'failed':
        response['status'] = 'failed'
        response['msg'] = errMsgs
        response[API_REQUESTS]= requests_counter
        return response
    response['status']= 'success'
    response[API_REQUESTS] = requests_counter
    response['msg'] = skipMsgs
    return response


#   overall flow->
# get integration settings with or without given project_ids
# filter settings based on flag exclude_project_ids
# separate valid and invalid settings based on access token validity
# for invalid tokens, try and update the token using refesh token
# for multi-account, split 1 setting into multiple settings
# for each setting get last sync info
# for each setting get collections
    # get add-account-data
    # get campaign_group data
    # get campaign data
    # get creative data
    # get company insights
    # check for errors and return response
# combine errors and ping healtcheck
if __name__ == '__main__':
    (options, args) = parser.parse_args()

    failures = []
    successes = []
    token_failures = []

    data_service = DataService(options)
    is_project_id_flag_given = options.project_ids != None and options.project_ids != ''
    
    linkedin_int_settings =[]
    err_get_settings = ''
    if is_project_id_flag_given:
        linkedin_int_settings, err_get_settings = data_service.get_linkedin_int_settings_for_projects(options.project_ids)
    else:
        linkedin_int_settings, err_get_settings = data_service.get_linkedin_int_settings()
    if err_get_settings != '':
        failures.append({'status': 'failed', 'msg': 'Failed to get linkedin settings'})
    
    start_timestamp = options.start_timestamp
    end_timestamp = options.end_timestamp


    if(linkedin_int_settings is not None):
        
        required_linkedin_settings = U.remove_excluded_projects(linkedin_int_settings,
                                                         options.exclude_project_ids)
        valid_linkedin_settings, invalid_linkedin_settings = U.separate_valid_and_invalid_tokens(linkedin_int_settings)
        
        settings_with_updated_tokens, token_failures = U.generate_and_update_access_token(options,
                                                             invalid_linkedin_settings)
        valid_linkedin_settings.extend(settings_with_updated_tokens)
        
        split_linkedin_settings, split_failures = U.split_settings_for_multiple_ad_accounts(valid_linkedin_settings)
        token_failures.extend(split_failures)
        
        for setting in split_linkedin_settings:
            sync_info_with_type = {}
            sync_info_with_type, err = data_service.get_last_sync_info(setting, start_timestamp)
            if err != '':
                response['status'] = 'failed'
                response['msg'] = 'Failed to get last sync info'
            else:
                response = get_collections(options, setting, sync_info_with_type, end_timestamp)

            response[PROJECT_ID] = setting.project_id
            response[AD_ACCOUNT] = setting.ad_account
            if(response['status']=='failed'):
                if 'Failed' in response['msg'] and 'ad_accounts metadata' in response['msg']:
                    token_failures.append(response)
                else:
                    failures.append(response)
            else:
                successes.append(response)
        
        ping_healthcheck(successes, failures, token_failures) 
        log.warning('Successfully synced. End of Linkedin sync job.')
        sys.exit(0)