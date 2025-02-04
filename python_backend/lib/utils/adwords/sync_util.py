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
        next_sync_infos = []
        is_input_timerange_given = False

        if AdwordsSyncUtil.non_historical_doc_type(sync_doc_type):
            next_timestamp = None
            next_timestamp_end = None
            if input_last_timestamp is not None:
                next_timestamp = SyncUtil.get_next_start_time(input_last_timestamp)
                is_input_timerange_given = True
            else:
                next_timestamp = SyncUtil.get_next_start_time(sync_last_timestamp)

            if input_to_timestamp is not None:
                next_timestamp_end = input_to_timestamp
                is_input_timerange_given = True
            else:
                next_timestamp_end = TimeUtil.get_timestamp_from_datetime(datetime.utcnow())

            next_sync_info = last_sync.copy()
            next_sync_info['next_timestamp'] = next_timestamp
            next_sync_info['next_timestamp_end'] = next_timestamp_end
            next_sync_info['first_run'] = first_run
            next_sync_infos.append(next_sync_info)

        else:
            next_timestamps = []
            if input_last_timestamp is not None and input_to_timestamp is not None:
                is_input_timerange_given = True
                next_timestamps = SyncUtil.get_next_timestamps(input_last_timestamp, input_to_timestamp)
            elif input_last_timestamp is not None and input_to_timestamp is None:
                is_input_timerange_given = True
                next_timestamps = SyncUtil.get_next_timestamps(input_last_timestamp)
            else:
                next_timestamps = SyncUtil.get_next_timestamps(sync_last_timestamp)

            for next_timestamp in next_timestamps:
                next_sync_info = last_sync.copy()
                next_sync_info['next_timestamp'] = next_timestamp
                next_sync_info['first_run'] = first_run
                next_sync_infos.append(next_sync_info)

        return next_sync_infos, is_input_timerange_given

    @staticmethod
    def doesnt_contains_historical_data(last_timestamp, doc_type):
        adwords_timestamp_today = TimeUtil.get_timestamp_from_datetime(datetime.utcnow())
        non_report_related = AdwordsSyncUtil.non_historical_doc_type(doc_type)
        return non_report_related and last_timestamp != adwords_timestamp_today

    @staticmethod
    def non_historical_doc_type(doc_type):
        return doc_type in [CAMPAIGNS, ADS, AD_GROUPS, CUSTOMER_ACCOUNT_PROPERTIES]

    @staticmethod
    def is_token_error(message):
        is_token_failure = False
        if ("invalid_grant" in message.lower() or "PERMISSION_DENIED".lower() in message.lower() 
            or "invalid params" in message.lower() or "access token" in message.lower() 
            or "refresh_token" in message.lower()):
            is_token_failure = True
        
        return is_token_failure