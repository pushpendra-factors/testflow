import logging as log
import sys

import scripts
from lib.data_services.factors_data_service import FactorsDataService
from lib.utils.sync_util import SyncUtil
from scripts.gsc.etl_config import EtlConfig
from scripts.gsc.etl_parser import EtlParser
from scripts.gsc.job_scheduler import JobScheduler


def setup(argv):
    input_args, rem = EtlParser(argv[1::]).parse()
    EtlConfig.build(input_args)
    scripts.gsc.CONFIG = EtlConfig
    FactorsDataService.init(scripts.gsc.CONFIG.GSC_APP.get_data_service_path())
    return


# TODO Error handling not done.
def get_last_sync_infos(include_project_ids, exclude_project_ids, doc_type):
    last_sync_infos = []

    if len(include_project_ids) == 0:
        last_sync_infos = FactorsDataService.get_gsc_last_sync_infos_for_all_projects()
    else:
        for project_id in include_project_ids:
            current_sync_infos = FactorsDataService.get_gsc_last_sync_infos_for_project(project_id)
            last_sync_infos.extend(current_sync_infos)

    last_sync_infos = remove_excluded_project_ids(last_sync_infos, exclude_project_ids)
    last_sync_infos = filter_doc_type(last_sync_infos, doc_type)
    return last_sync_infos


def remove_excluded_project_ids(last_sync_infos, exclude_project_ids):
    if len(exclude_project_ids) == 0:
        return last_sync_infos

    return [last_sync_info for last_sync_info in last_sync_infos if
            last_sync_info.get("project_id") not in exclude_project_ids]

    # Extract cant be given with from and to.

def filter_doc_type(last_sync_infos, doc_type):
    if doc_type is None:
        return last_sync_infos

    return [last_sync_info for last_sync_info in last_sync_infos if str(last_sync_info.get("type")) in doc_type]

if __name__ == "__main__":
    setup(sys.argv)
    log.basicConfig(level=log.INFO)
    log.warning("Started search console sync job.")

    env = scripts.gsc.CONFIG.GSC_APP.env
    is_dry = scripts.gsc.CONFIG.GSC_APP.dry
    skip_today = scripts.gsc.CONFIG.GSC_APP.skip_today
    input_project_ids = scripts.gsc.CONFIG.GSC_APP.project_ids
    input_document_type = scripts.gsc.CONFIG.GSC_APP.document_type
    input_exclude_project_ids = scripts.gsc.CONFIG.GSC_APP.exclude_project_ids
    input_last_timestamp = scripts.gsc.CONFIG.GSC_APP.last_timestamp
    input_to_timestamp = scripts.gsc.CONFIG.GSC_APP.to_timestamp
    metrics_controller = scripts.gsc.CONFIG.GSC_APP.metrics_controller


    final_last_sync_infos = get_last_sync_infos(input_project_ids, input_exclude_project_ids, input_document_type)
    for last_sync in final_last_sync_infos:
        next_sync_infos = SyncUtil.get_gsc_next_sync_infos(last_sync, input_last_timestamp, input_to_timestamp)
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

    metrics_controller.publish_gsc()
    log.warning("Successfully synced. End of search console sync job.")
    sys.exit(0)
