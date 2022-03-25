import logging as log
import sys

import scripts
from lib.data_services.factors_data_service import FactorsDataService
from lib.utils.adwords.sync_util import AdwordsSyncUtil
from scripts.adwords.etl_config import EtlConfig
from scripts.adwords.etl_parser import EtlParser
from scripts.adwords.job_scheduler import JobScheduler


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
        return last_sync_infos

    return [last_sync_info for last_sync_info in last_sync_infos if last_sync_info.get("doc_type_alias") in doc_type]

    # Extract cant be given with from and to.


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
    new_extract_project_id = scripts.adwords.CONFIG.ADWORDS_APP.new_extract_project_id

    final_last_sync_infos = get_last_sync_infos(input_project_ids, input_exclude_project_ids, input_document_type, input_timezone)
    
    for last_sync in final_last_sync_infos:
        next_sync_infos = AdwordsSyncUtil.get_next_sync_infos(last_sync, input_last_timestamp, input_to_timestamp)    
        if next_sync_infos is None:
            continue
        for next_sync in next_sync_infos:
            if JobScheduler.validate(next_sync, skip_today):
                is_new_job = '*' in new_extract_project_id or next_sync["project_id"] in new_extract_project_id
                JobScheduler(next_sync, skip_today).sync(env, is_dry, is_new_job)
            else:
                log.warning("Skipping job scheduler for following project with following properties: "+ str(next_sync))
                continue
            

    metrics_controller.publish()
    log.warning("Successfully synced. End of adwords sync job.")
    sys.exit(0)
