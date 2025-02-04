import logging as log
import sys

import scripts
from lib.data_services.factors_data_service import FactorsDataService
from lib.adwords.oauth_service.fetch_service import FetchService
from lib.utils.adwords.sync_util import AdwordsSyncUtil
from scripts.adwords.etl_config import EtlConfig
from scripts.adwords.etl_parser import EtlParser
from scripts.adwords.job_scheduler import JobScheduler
from scripts.adwords.new_jobs.query import QueryBuilder
from scripts.adwords import *


def setup(argv):
    input_args, rem = EtlParser(argv[1::]).parse()
    EtlConfig.build(input_args)
    scripts.adwords.CONFIG = EtlConfig
    FactorsDataService.init(scripts.adwords.CONFIG.ADWORDS_APP.get_data_service_path())
    return


# TODO Error handling not done.
def get_last_sync_infos(include_project_ids, exclude_project_ids, doc_type, input_timezone):
    last_sync_infos = []
    if len(include_project_ids) == 0:
        last_sync_infos = FactorsDataService.get_last_sync_infos_for_all_projects()
    else:
        for project_id in include_project_ids:
            current_sync_infos = FactorsDataService.get_last_sync_infos_for_project(project_id)
            last_sync_infos.extend(current_sync_infos)

    last_sync_infos = remove_excluded_project_ids(last_sync_infos, exclude_project_ids)
    last_sync_infos = filter_based_on_input_timezone(last_sync_infos, input_timezone)
    last_sync_infos = filter_doc_type(last_sync_infos, doc_type)
    return last_sync_infos


def remove_excluded_project_ids(last_sync_infos, exclude_project_ids):
    if len(exclude_project_ids) == 0:
        return last_sync_infos

    return [last_sync_info for last_sync_info in last_sync_infos if
            last_sync_info.get("project_id") not in exclude_project_ids]


def filter_based_on_input_timezone(last_sync_infos, input_timezone):
    if input_timezone == "":
        return last_sync_infos
    resultant_last_sync_infos = []
    for last_sync_info in last_sync_infos:
        if input_timezone == scripts.adwords.TIMEZONE_IST and last_sync_info.get("timezone") == input_timezone:
            resultant_last_sync_infos.append(last_sync_info)

        if input_timezone != scripts.adwords.TIMEZONE_IST and last_sync_info.get("timezone") != scripts.adwords.TIMEZONE_IST:
            resultant_last_sync_infos.append(last_sync_info)

    return resultant_last_sync_infos

def filter_doc_type(last_sync_infos, doc_type):
    if doc_type is None:
        return [last_sync_info for last_sync_info in last_sync_infos if last_sync_info.get("doc_type_alias") != "search_performance_report"]

    return [last_sync_info for last_sync_info in last_sync_infos if (last_sync_info.get("doc_type_alias") in doc_type and last_sync_info.get("doc_type_alias") != "search_performance_report")]

    # Extract cant be given with from and to.

# format: { project_id1: {customer1: [doc1_sync_info, doc2_sync_info....], customer2: [doc1, doc2...]}, projectId2...}
def get_last_sync_infos_map_by_project_id_customer_id(last_sync_infos):
    last_sync_infos_map = {}
    for last_sync in last_sync_infos:
        customer_acc_id = last_sync.get("customer_acc_id")
        project_id = last_sync.get("project_id")

        if project_id not in last_sync_infos_map:
            last_sync_infos_map[project_id] = {}
            if customer_acc_id not in last_sync_infos_map[project_id]:
                last_sync_infos_map[project_id][customer_acc_id] = []
            last_sync_infos_map[project_id][customer_acc_id].append(last_sync)
        else:
            if customer_acc_id not in last_sync_infos_map[project_id]:
                last_sync_infos_map[project_id][customer_acc_id] = []
            last_sync_infos_map[project_id][customer_acc_id].append(last_sync)

    return last_sync_infos_map

def precheck_token(last_sync):
    project_id = last_sync.get("project_id")
    customer_acc_id = last_sync.get("customer_acc_id")
    refresh_token = last_sync.get("refresh_token")
    manager_id = last_sync.get("manager_id")
    doc_type = last_sync.get("doc_type_alias")
    EXTRACT_FIELDS = [
        "customer.id",
        "customer.currency_code",
        "customer.time_zone",
        "customer.manager",
        "customer.test_account",
    ]
    try:
        service_query = (QueryBuilder()
                            .Select(EXTRACT_FIELDS)
                            .From("customer")
                            # .Limit(1)
                            .Build())

        ads_service = FetchService(scripts.adwords.CONFIG.ADWORDS_OAUTH).new_get_service(
                                            "GoogleAdsService", refresh_token, manager_id)
        stream = ads_service.search_stream(customer_id=customer_acc_id, query=service_query)
    
    except Exception as e:
        str_exception = str(e)
        message = str_exception
        if AdwordsSyncUtil.is_token_error(message):
            metrics_controller.update_job_stats(project_id, customer_acc_id, 
                                                    doc_type, "failed", message)
            return True
    return False


if __name__ == "__main__":
    setup(sys.argv)
    log.basicConfig(level=log.INFO)
    log.warning("Started adwords sync job.")

    env = scripts.adwords.CONFIG.ADWORDS_APP.env
    is_dry = scripts.adwords.CONFIG.ADWORDS_APP.dry
    skip_today = scripts.adwords.CONFIG.ADWORDS_APP.skip_today
    input_project_ids = scripts.adwords.CONFIG.ADWORDS_APP.project_ids
    input_exclude_project_ids = scripts.adwords.CONFIG.ADWORDS_APP.exclude_project_ids
    input_document_type = scripts.adwords.CONFIG.ADWORDS_APP.document_type
    input_last_timestamp = scripts.adwords.CONFIG.ADWORDS_APP.last_timestamp
    input_to_timestamp = scripts.adwords.CONFIG.ADWORDS_APP.to_timestamp
    metrics_controller = scripts.adwords.CONFIG.ADWORDS_APP.metrics_controller
    input_timezone = scripts.adwords.CONFIG.ADWORDS_APP.timezone

    final_last_sync_infos = get_last_sync_infos(input_project_ids, input_exclude_project_ids, input_document_type, input_timezone)
    mapped_last_sync_infos = get_last_sync_infos_map_by_project_id_customer_id(final_last_sync_infos)

    for project_id, last_sync_customer_acc_id in mapped_last_sync_infos.items():
        for customer_acc_id, last_sync_infos in last_sync_customer_acc_id.items():
            is_token_invalid = precheck_token(last_sync_infos[0])
            if is_token_invalid:
                continue
            for last_sync in last_sync_infos:
                next_sync_infos, is_input_timerange_given = AdwordsSyncUtil.get_next_sync_infos(
                                                            last_sync, input_last_timestamp, input_to_timestamp) 
                # for normal job run, if data is missing for more than 31 days, we pull data for last 30 days and notify in healthcheck
                # for custom job run, we don't pull any data and we notify healthchecks
                if len(next_sync_infos) > 31:
                    error_message = ("From and to timestamps exceed 31 days: From: " + str(input_last_timestamp) + " To: " 
                                    + str(input_to_timestamp))
                    metrics_controller.update_job_stats(last_sync.get("project_id"), last_sync.get("customer_acc_id"), 
                                                        last_sync.get("doc_type_alias"), "failed", error_message)
                    if is_input_timerange_given:
                        continue
                    else:
                        next_sync_infos = next_sync_infos[-30:]
                if next_sync_infos is None:
                    continue
                for next_sync in next_sync_infos:
                    if JobScheduler.validate(next_sync, skip_today):
                        err = JobScheduler(next_sync, skip_today).sync(env, is_dry)
                        if err != '' and err != None:
                            break
                    else:
                        log.warning("Skipping job scheduler for following project with following properties: "+ str(next_sync))
                        continue
            

    metrics_controller.publish()
    log.warning("Successfully synced. End of adwords sync job.")
    sys.exit(0)
