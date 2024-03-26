from constants.constants import *
from util.util import Util as U

class MetricsAggregator:
    etl_stats = {
        "failures": {},
        "token_failures": [],
        "success": {},
        "warnings": {}
    }
    request_counter = 0
    job_type = 'daily'
    __instance = None

    @staticmethod
    def get_instance():
        if MetricsAggregator.__instance == None:
            MetricsAggregator()
        return MetricsAggregator.__instance

    def __init__(self) -> None:
        MetricsAggregator.__instance = self

    def update_stats(self, project_id, ad_account, doc_type=None, request_counter=0, status='success', err_msg=''):
        self.request_counter += request_counter
        msg_dict = {
            'status': status,
            'err_msg': err_msg,
            'total_api_requests': self.request_counter
        }
        if status == 'failed' or err_msg != '':
            if NO_CAMPAIGN_ERR in err_msg:
                self.etl_stats['warnings'].setdefault(self.job_type, {})
                self.etl_stats['warnings'][self.job_type].setdefault(project_id, {})
                self.etl_stats['warnings'][self.job_type][project_id].setdefault(ad_account, {})
                self.etl_stats['warnings'][self.job_type][project_id][ad_account].setdefault(doc_type, msg_dict)
            else:
                self.etl_stats['failures'].setdefault(self.job_type, {})
                self.etl_stats['failures'][self.job_type].setdefault(project_id, {})
                self.etl_stats['failures'][self.job_type][project_id].setdefault(ad_account, {})
                self.etl_stats['failures'][self.job_type][project_id][ad_account].setdefault(doc_type, msg_dict)
        else:
            self.etl_stats['success'].setdefault(self.job_type, {})
            self.etl_stats['success'][self.job_type].setdefault(project_id, {})
            self.etl_stats['success'][self.job_type][project_id].setdefault(ad_account, {})
            self.etl_stats['success'][self.job_type][project_id][ad_account] = msg_dict
        self.request_counter = 0
    
    def reset_request_counter(self):
        self.request_counter = 0

    def ping_notification_services(self, env, healthcheck_ping_id):
        status_msg = ''
        failures = self.etl_stats['failures']
        successes = self.etl_stats['success']
        warnings = self.etl_stats['warnings']
        token_failures = self.etl_stats['token_failures']
        if len(failures) > 0: status_msg = 'Failures on sync.'
        else: status_msg = 'Successfully synced.'
        notification_payload = {
            'status': status_msg, 
            'failures': failures, 
            'success': successes,
            'warnings': warnings,
        }
        if len(failures) > 0:
            U.ping_healthcheck(env, healthcheck_ping_id,
                notification_payload, endpoint='/fail')
        else:
            U.ping_healthcheck(env, healthcheck_ping_id, notification_payload)
        if len(token_failures) > 0:
            notification_payload = {
                'status': 'Token failures', 
                'failures': token_failures,
            }
            U.ping_healthcheck(env, HEALTHCHECK_TOKEN_FAILURE_PING_ID, 
                notification_payload, endpoint='/fail')
            
            U.build_message_and_ping_slack(env, SLACK_URL, token_failures)

