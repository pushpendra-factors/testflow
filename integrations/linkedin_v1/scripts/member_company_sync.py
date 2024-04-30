from optparse import OptionParser
import logging as log
import sys
import traceback
import signal
from constants.constants import *
from custom_exception.custom_exception import CustomException
from util.util import Util as U
from linkedin_setting.linkedin_setting import LinkedinSetting
from global_objects.global_obj_creator import GlobalObjects
from job_runners.weekly_member_company_job_runner import WeeklyMemberCompanyJobRunner
from job_runners.member_company_job_runner import MemberCompanyJobRunner
from cache.campaign_group_info import CampaignGroupInfo
from cache.campaign_info import CampaignInfo
from data_service.data_service import DataService
from metrics_aggregator.metrics_aggregator import MetricsAggregator


parser = OptionParser()
parser.add_option('--env', dest='env', default='development')
parser.add_option('--dry', dest='dry', help='', default='False')
parser.add_option('--skip_today', dest='skip_today', help='', default='False') 
parser.add_option('--project_ids', dest='project_ids', help='', default='', type=str)
parser.add_option('--exclude_project_ids', dest='exclude_project_ids', help='', default='', type=str)
parser.add_option('--client_id', dest='client_id', help='',default=None, type=str)
parser.add_option('--client_secret', dest='client_secret', help='',default=None, type=str)
parser.add_option('--data_service_host', dest='data_service_host',
    help='Data service host', default='http://localhost:8089')
parser.add_option('--input_start_timestamp', dest='input_start_timestamp', help='', default=None, type=int)
parser.add_option('--input_end_timestamp', dest='input_end_timestamp', help='', default=None, type=int)
parser.add_option('--job_type', dest='job_type', default='1,2,3', type=str)
parser.add_option('--new_change_project_ids', dest='new_change_project_ids', help='', default='')
# job_type -> 1 for daily job, 2 for t8 job, 3 for t22
# to run combination like t8 and t22 only -> use job_type = "2,3"

def sync_company_data(options, linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp):
    daily_job_req, t8_job_req, t22_job_req = U.check_job_type_req(options.job_type)

    campaign_group_cache = CampaignGroupInfo.get_instance()
    if daily_job_req:
        max_ingestion_timestamp = 0
        if MEMBER_COMPANY_INSIGHTS in sync_info_with_type:
            max_ingestion_timestamp = sync_info_with_type[MEMBER_COMPANY_INSIGHTS]
        MemberCompanyJobRunner(linkedin_setting, max_ingestion_timestamp, 
                                input_start_timestamp, input_end_timestamp).execute()
        campaign_group_cache.reset_campaign_group_data()
    
    if t8_job_req:
        WeeklyMemberCompanyJobRunner(sync_info_with_type[SYNC_INFO_KEY_T8], T8_END_BUFFER, 
                SYNC_STATUS_T8, 't8', linkedin_setting, input_start_timestamp, input_end_timestamp).execute()
        campaign_group_cache.reset_campaign_group_data()
        
    if t22_job_req:
        WeeklyMemberCompanyJobRunner(sync_info_with_type[SYNC_INFO_KEY_T22], T22_END_BUFFER, 
                SYNC_STATUS_T22, 't22', linkedin_setting, input_start_timestamp, input_end_timestamp).execute()
        campaign_group_cache.reset_campaign_group_data()

def sync_company_data_v1(options, linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp):
    daily_job_req, t8_job_req, t22_job_req = U.check_job_type_req(options.job_type)

    campaign_group_cache = CampaignGroupInfo.get_instance()
    campaign_cache = CampaignInfo.get_instance()
    if daily_job_req:
        max_ingestion_timestamp = 0
        if MEMBER_COMPANY_INSIGHTS in sync_info_with_type:
            max_ingestion_timestamp = sync_info_with_type[MEMBER_COMPANY_INSIGHTS]
        MemberCompanyJobRunner(linkedin_setting, max_ingestion_timestamp, 
                                input_start_timestamp, input_end_timestamp).execute_v1()
        campaign_group_cache.reset_campaign_group_data()
        campaign_cache.reset_campaign_data()
    
    if t8_job_req:
        WeeklyMemberCompanyJobRunner(sync_info_with_type[SYNC_INFO_KEY_T8], T8_END_BUFFER, 
                SYNC_STATUS_T8, 't8', linkedin_setting, input_start_timestamp, input_end_timestamp).execute_v1()
        campaign_group_cache.reset_campaign_group_data()
        campaign_cache.reset_campaign_data()
        
    if t22_job_req:
        WeeklyMemberCompanyJobRunner(sync_info_with_type[SYNC_INFO_KEY_T22], T22_END_BUFFER, 
                SYNC_STATUS_T22, 't22', linkedin_setting, input_start_timestamp, input_end_timestamp).execute_v1()
        campaign_group_cache.reset_campaign_group_data()
        campaign_cache.reset_campaign_data()
            

if __name__ == '__main__':
    (options, args) = parser.parse_args()

    input_start_timestamp = options.input_start_timestamp
    input_end_timestamp = options.input_end_timestamp

    globalObject = GlobalObjects(options.data_service_host)
    data_service_obj = DataService.get_instance()
    metrics_aggregator_obj = MetricsAggregator.get_instance()

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
            sync_info_with_type = data_service_obj.get_last_sync_info_for_company_data(
                                                    linkedin_setting, input_start_timestamp, 
                                                                        input_end_timestamp)
            if options.new_change_project_ids == "*" or linkedin_setting.project_id in options.new_change_project_ids:
                sync_company_data_v1(options, linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp)
            else:
                sync_company_data(options, linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp)
        except CustomException as e:
            traceback.print_tb(e.__traceback__)
            metrics_aggregator_obj.update_stats(linkedin_setting.project_id, linkedin_setting.ad_account, 
                                                            e.doc_type, e.request_count, 'failed', e.message)
        except Exception as e:
            traceback.print_tb(e.__traceback__)
            metrics_aggregator_obj.update_stats(linkedin_setting.project_id, linkedin_setting.ad_account, 
                                                            0, 0, 'failed', str(e))
        
        metrics_aggregator_obj.reset_request_counter()
    
    metrics_aggregator_obj.ping_notification_services(options.env, HEALTHCHECK_COMPANY_SYNC_JOB)
    log.warning('Successfully synced. End of Linkedin sync job.')
    sys.exit(0)