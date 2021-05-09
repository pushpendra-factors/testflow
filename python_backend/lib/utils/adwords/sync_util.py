from datetime import datetime

import scripts
from lib.utils.time import TimeUtil
from scripts.adwords import CAMPAIGNS, ADS, AD_GROUPS, CUSTOMER_ACCOUNT_PROPERTIES
from lib.utils.sync_util import SyncUtil

class AdwordsSyncUtil:
    MAX_LOOK_BACK_DAYS = 30

    # generates next sync info with all missing timestamps for each document type.
    @staticmethod
    def get_next_sync_infos(last_sync, input_last_timestamp, input_to_timestamp):
        sync_last_timestamp = last_sync.get("last_timestamp")
        sync_doc_type = last_sync.get("doc_type_alias")
        first_run = (sync_last_timestamp == 0)
        next_timestamps = AdwordsSyncUtil.get_next_timestamps_for_run(first_run, input_last_timestamp, input_to_timestamp,
                                                                      sync_last_timestamp, sync_doc_type)
        next_sync_infos = []
        for next_timestamp in next_timestamps:
            next_sync_info = last_sync.copy()
            next_sync_info['next_timestamp'] = next_timestamp
            next_sync_info['first_run'] = first_run
            next_sync_infos.append(next_sync_info)
        return next_sync_infos

    @staticmethod
    def get_next_timestamps_for_run(first_run, input_last_timestamp, input_to_timestamp, sync_last_timestamp, sync_doc_type):
        next_timestamps = []
        if first_run or (input_last_timestamp is None and input_to_timestamp is None):
            if AdwordsSyncUtil.doesnt_contains_historical_data(sync_last_timestamp, sync_doc_type):
                next_timestamps = [TimeUtil.get_timestamp_from_datetime(datetime.utcnow())]
            else:
                next_timestamps = SyncUtil.get_next_timestamps(sync_last_timestamp)

        if scripts.adwords.CONFIG.ADWORDS_APP.type_of_run != scripts.adwords.EXTRACT:
            if input_last_timestamp is not None and input_to_timestamp is not None:
                next_timestamps = SyncUtil.get_next_timestamps(input_last_timestamp, input_to_timestamp)
            elif input_last_timestamp is not None:
                next_timestamps = SyncUtil.get_next_timestamp(input_last_timestamp)
        return next_timestamps


    @staticmethod
    def doesnt_contains_historical_data(last_timestamp, doc_type):
        adwords_timestamp_today = TimeUtil.get_timestamp_from_datetime(datetime.utcnow())
        non_report_related = AdwordsSyncUtil.non_historical_doc_type(doc_type)
        return non_report_related and last_timestamp != adwords_timestamp_today

    @staticmethod
    def non_historical_doc_type(doc_type):
        return doc_type in [CAMPAIGNS, ADS, AD_GROUPS, CUSTOMER_ACCOUNT_PROPERTIES]
