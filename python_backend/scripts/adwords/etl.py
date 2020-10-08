import sys
from datetime import datetime
import logging as log

import scripts
from lib.data_services.factors_data_service import FactorsDataService
from lib.sns_notifier import SnsNotifier
from lib.utils.time import TimeUtil
from scripts.adwords import STATUS_FAILED, STATUS_SKIPPED, APP_NAME, etl_record_stats
from scripts.adwords.etl_config import EtlConfig
from scripts.adwords.etl_parser import EtlParser
from scripts.adwords.job_scheduler import JobScheduler
from scripts.adwords.jobs.reports_fetch_job import ReportsFetch

# from . import STATUS_FAILED, STATUS_SKIPPED, APP_NAME, CONFIG, etl_record_stats
# from .etl_config import EtlConfig
# from .etl_parser import EtlParser
# from .jobs.reports_fetch import ReportsFetch

def setup(argv):
    input_args, rem = EtlParser(argv[1::]).parse()
    EtlConfig.build(input_args)
    scripts.adwords.CONFIG = EtlConfig
    FactorsDataService.init(config=scripts.adwords.CONFIG)
    return


# generates next sync info with all missing timestamps
# for each document type.
def get_next_sync_info(last_sync):
    next_sync_info = []
    last_timestamp = last_sync.get('last_timestamp')
    doc_type = last_sync.get('doc_type_alias')

    if ReportsFetch.contains_historical_data(last_timestamp, doc_type):
        sync_info = last_sync.copy()
        sync_info['next_timestamp'] = TimeUtil.get_timestamp_from_datetime(datetime.utcnow())
        return [sync_info]
    else:
        start_timestamp = ReportsFetch.get_next_start_time_for_historical_data(last_timestamp)
        next_timestamps = TimeUtil.get_timestamp_range(start_timestamp)

        for timestamp in next_timestamps:
            sync_info = last_sync.copy()
            sync_info['next_timestamp'] = timestamp
            next_sync_info.append(sync_info)

        return next_sync_info


if __name__ == "__main__":
    setup(sys.argv)
    log.basicConfig(level=log.INFO)
    log.warning("Started adwords sync job.")

    is_dry = scripts.adwords.CONFIG.ADWORDS_APP.dry
    skip_today = scripts.adwords.CONFIG.ADWORDS_APP.skip_today
    last_sync_infos = FactorsDataService.get_last_sync_infos()
    input_project_id = scripts.adwords.CONFIG.ADWORDS_APP.project_id

    next_sync_failures = []
    next_sync_skipped = []
    next_sync_success = {}

    for last_sync in last_sync_infos:
        # add next_sync_info only for the selected project.
        if input_project_id is not None:
            project_id = last_sync.get("project_id")
            if project_id != input_project_id:
                continue

        last_timestamp = last_sync.get('last_timestamp')
        if last_timestamp is None:
            log.error("Missing last_timestamp in last sync info.")
            continue

        doc_type = last_sync.get('doc_type_alias')
        if doc_type is None:
            log.error("Missing doc_type_alias name on last_sync_info.")
            continue

        next_sync_infos = get_next_sync_info(last_sync)
        if next_sync_infos is None:
            continue
        for next_sync in next_sync_infos:
            response = JobScheduler(next_sync, skip_today).sync(scripts.adwords.CONFIG.ADWORDS_APP.env, is_dry)
            status = response.get("status")
            if status is None:
                next_sync_failures.append("Sync status is missing on response")
            elif status == STATUS_FAILED:
                next_sync_failures.append(response)
            elif status == STATUS_SKIPPED:
                next_sync_skipped.append(response)
            else:
                next_sync_success[next_sync.get("project_id")] = next_sync.get("customer_acc_id")

    status_msg = ""
    if len(next_sync_failures) > 0:
        status_msg = "Failures on sync."
    else:
        status_msg = "Successfully synced."
    notify_payload = {
        "status": status_msg,
        "failures": next_sync_failures,
        "skipped": next_sync_skipped,
        "success": {"projects": next_sync_success},
        "requests": etl_record_stats,
    }
    SnsNotifier.notify(scripts.adwords.CONFIG.ADWORDS_APP.env, APP_NAME, notify_payload)
    log.warning("Successfully synced. End of adwords sync job.")
    sys.exit(0)
