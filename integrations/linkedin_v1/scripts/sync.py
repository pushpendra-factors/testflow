from optparse import OptionParser
import logging as log
import sys
import traceback
import signal
from constants.constants import *
from custom_exception.custom_exception import CustomException
from linkedin_setting.linkedin_setting import LinkedinSetting
from jobs.ad_account import AdAccountJob
from jobs.campaign_group import CampaignGroupJob
from jobs.campaign import CampaignJob
from global_objects.global_obj_creator import metrics_aggregator_obj, data_service_obj, campaign_group_cache, campaign_cache, creative_cache
from util.util import Util as U



parser = OptionParser()
parser.add_option('--env', dest='env', default='development')
parser.add_option('--dry', dest='dry', help='', default='False')
parser.add_option('--skip_today', dest='skip_today', help='', default='False') 
parser.add_option('--is_dry_run', dest='is_dry_run', help='', default= 'False', type= str)
parser.add_option('--project_ids', dest='project_ids', help='', default="", type=str)
parser.add_option('--exclude_project_ids', dest='exclude_project_ids', help='', default='', type=str)
parser.add_option('--client_id', dest='client_id', help='',default=None, type=str)
parser.add_option('--client_secret', dest='client_secret', help='',default=None, type=str)
parser.add_option('--data_service_host', dest='data_service_host',
    help='Data service host', default='http://localhost:8089')
parser.add_option('--input_start_timestamp', dest='input_start_timestamp', help='', default=None, type=int)
parser.add_option('--input_end_timestamp', dest='input_end_timestamp', help='', default=None, type=int)


def clear_cache():
    campaign_group_cache.reset_campaign_group_data()
    campaign_cache.reset_campaign_data()
    creative_cache.reset_creative_data()

def sync_ads_data(linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp):

    AdAccountJob(linkedin_setting, input_end_timestamp).execute()

    CampaignGroupJob(linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp).execute()

    CampaignJob(linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp).execute()

    clear_cache()
    metrics_aggregator_obj.update_stats(linkedin_setting.project_id, linkedin_setting.ad_account)
        
def handle(signum, frame):
    raise Exception("Function timeout after 20 mins")

if __name__ == '__main__':
    (options, args) = parser.parse_args()

    input_start_timestamp = options.input_start_timestamp
    input_end_timestamp = options.input_end_timestamp

    data_service_obj.data_service_host = options.data_service_host
    is_project_id_flag_given = (options.project_ids != None and options.project_ids != '')
    allProjects = options.project_ids == "*"
    
    linkedin_int_settings =[]
    err_get_settings = ''
    if is_project_id_flag_given and not allProjects:
        linkedin_int_settings, err_get_settings = data_service_obj.get_linkedin_int_settings_for_projects(
                                                                                    options.project_ids)
    else:
        linkedin_int_settings, err_get_settings = data_service_obj.get_linkedin_int_settings()
    if err_get_settings != '':
        notification_payload = {'status': 'failed', 'errMsg': 'Failed to get linkedin settings'}
        U.ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload, '/fail')
        log.error(notification_payload['errMsg'])
        sys.exit(0)
    

    split_linkedin_settings, token_failures = LinkedinSetting.perform_token_inspect_and_split_settings(
                                                                        options, linkedin_int_settings)
    metrics_aggregator_obj.etl_stats['token_failures'] = token_failures
    
    for linkedin_setting in split_linkedin_settings:
        try:
            # timeout this function after 20 mins
            signal.signal(signal.SIGALRM, handle)
            signal.alarm(1200)
            # 
            sync_info_with_type = data_service_obj.get_last_sync_info(linkedin_setting, input_start_timestamp, 
                                                                                        input_end_timestamp)
            sync_ads_data(linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp)
        
        except CustomException as e:
            traceback.print_tb(e.__traceback__)
            if AD_ACCOUNT_FAILURE in str(e):
                metrics_aggregator_obj.etl_stats['token_failures'].append({'status': 'failed', 'errMsg': str(e), 
                                                                        PROJECT_ID: linkedin_setting.project_id, 
                                                                        AD_ACCOUNT: linkedin_setting.ad_account})
            else:
                metrics_aggregator_obj.update_stats(linkedin_setting.project_id, linkedin_setting.ad_account, 
                                                            e.doc_type, e.request_count, 'failed', e.message)
        except Exception as e:
            traceback.print_tb(e.__traceback__)
            if AD_ACCOUNT_FAILURE in str(e):
                metrics_aggregator_obj.etl_stats['token_failures'].append({'status': 'failed', 'errMsg': str(e), 
                                                                        PROJECT_ID: linkedin_setting.project_id, 
                                                                        AD_ACCOUNT: linkedin_setting.ad_account})
            else:
                metrics_aggregator_obj.update_stats(linkedin_setting.project_id, linkedin_setting.ad_account, 
                                                            0, 0, 'failed', str(e))
        metrics_aggregator_obj.reset_request_counter()
    
    metrics_aggregator_obj.ping_notification_services(options.env, HEALTHCHECK_PING_ID)
    log.warning('Successfully synced. End of Linkedin sync job.')
    sys.exit(0)