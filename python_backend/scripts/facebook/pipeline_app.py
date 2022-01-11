import copy
import logging as log
import sys
import traceback
import time
from python_backend.scripts.facebook import TOKEN_EXPIRY

import scripts
from lib.data_services.factors_data_service import FactorsDataService
from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from lib.utils.healthchecks import HealthChecksUtil
from scripts.facebook import FACEBOOK_AD_ACCOUNT, PROJECT_ID, DEVELOPMENT, TEST
from scripts.facebook.etl_config import EtlConfig
from scripts.facebook.etl_parser import EtlParser
from scripts.facebook.job_scheduler_and_runner import JobSchedulerAndRunner

SECONDS_IN_A_DAY= 86400

def setup(argv):
    input_args, rem = EtlParser(argv[1::]).parse()
    EtlConfig.build(input_args)
    scripts.facebook.CONFIG = EtlConfig
    FactorsDataService.init(scripts.facebook.CONFIG.FACEBOOK_APP.get_data_service_path())
    return


def allow_project_ids(infos, include_project_ids):
    if len(include_project_ids) == 0:
        return infos

    return [info for info in infos if info.get("project_id") in include_project_ids]


def remove_project_ids(infos, exclude_project_ids):
    if len(exclude_project_ids) == 0:
        return infos

    return [info for info in infos if info.get("project_id") not in exclude_project_ids]

def filter_based_on_input_timezone(facebook_settings, input_timezone):
    if input_timezone == "":
        return facebook_settings
    resultant_facebook_settings = []
    for last_sync_info in facebook_settings:
        if input_timezone == scripts.facebook.TIMEZONE_IST and last_sync_info.get("timezone") == input_timezone:
            resultant_facebook_settings.append(last_sync_info)

        if input_timezone != scripts.facebook.TIMEZONE_IST and last_sync_info.get("timezone") != scripts.facebook.TIMEZONE_IST:
            resultant_facebook_settings.append(last_sync_info)

    return resultant_facebook_settings

# takes in token expriy timestamp as parameter, returns if token is already expired, will token expire within 10 days, days in which it'll expire
def check_token_expiry(token_expiry):
    current_timestamp = int(time.time())
    if token_expiry == 0:
        return False, False, 0
    if token_expiry-current_timestamp < (SECONDS_IN_A_DAY * 10):
        if token_expiry-current_timestamp <=0:
            return True, False, 0
        else:
            return False, True, int((token_expiry-current_timestamp)/SECONDS_IN_A_DAY)
    return False, False, None

# TODO IMP add notification 10 days before expiry to team@factors.ai.
# TODO: Move the expiry logic to notifications/stats itself.
if __name__ == "__main__":
    setup(sys.argv)
    log.basicConfig(level=log.INFO)
    log.warning("Started facebook sync job.")
    facebook_config = scripts.facebook.CONFIG.FACEBOOK_APP

    facebook_settings: dict = FactorsDataService.get_facebook_settings()
    if facebook_settings is None:
        MetricsAggregator.publish_to_healthcheck_failure()
        log.warning("Failed to get facebook facebook_settings. End of facebook sync job.")
        sys.exit(0)

    facebook_settings = allow_project_ids(facebook_settings, facebook_config.project_ids)
    facebook_settings = remove_project_ids(facebook_settings, facebook_config.exclude_project_ids)
    facebook_settings = filter_based_on_input_timezone(facebook_settings, facebook_config.timezone)
    token_expiry_payload = {"Token about to expire": [], "Token expired": []}

    try:
        for facebook_int_setting in facebook_settings:
            is_token_expired, will_token_expire, days_to_expire = check_token_expiry(facebook_int_setting[TOKEN_EXPIRY])
            if will_token_expire:
                token_expiry_payload["Token about to expire"].append({PROJECT_ID: facebook_int_setting[PROJECT_ID], "days_remaining": days_to_expire})
            if is_token_expired:
                token_expiry_payload["Token expired"].append({PROJECT_ID: facebook_int_setting[PROJECT_ID]})
            else:
                customer_account_ids = facebook_int_setting[FACEBOOK_AD_ACCOUNT].split(',')
                for customer_account_id in customer_account_ids:
                    last_sync_info_with_type: dict = FactorsDataService.get_facebook_last_sync_info(
                        facebook_int_setting[PROJECT_ID], customer_account_id)

                    facebook_int_setting_with_customer_account: dict = copy.deepcopy(facebook_int_setting)
                    facebook_int_setting_with_customer_account[FACEBOOK_AD_ACCOUNT] = customer_account_id
                    JobSchedulerAndRunner.sync(facebook_int_setting_with_customer_account, last_sync_info_with_type)
        if facebook_config.dry != True and facebook_config.env not in [DEVELOPMENT, TEST]:
            MetricsAggregator.publish()
            if len(token_expiry_payload["Token about to expire"]) != 0 or len(token_expiry_payload["Token expired"]) != 0:
                HealthChecksUtil.ping(MetricsAggregator.env, token_expiry_payload, MetricsAggregator.HEALTHCHECK_PING_ID_TOKEN_FAILURE, endpoint="/fail")
        log.warning("Successfully synced. End of facebook sync job.")
        sys.exit(0)            
    except Exception as e:
        traceback.print_tb(e.__traceback__)
        message = str(e)
        log.warning("Failed with exception: %s", message)
