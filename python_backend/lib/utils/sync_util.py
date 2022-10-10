from lib.utils.time import TimeUtil


class SyncUtil:
    MAX_LOOK_BACK_DAYS = 30

    # generates next sync info with all missing timestamps for each document type.
    @staticmethod
    def get_next_sync_infos(last_sync, input_last_timestamp, input_to_timestamp):
        sync_last_timestamp = last_sync.get("last_timestamp")
        sync_doc_type = last_sync.get("doc_type_alias")
        first_run = (sync_last_timestamp == 0)
        next_timestamps = SyncUtil.get_next_timestamps_for_run(first_run, input_last_timestamp, input_to_timestamp,
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
        first_run = False
        if first_run or (input_last_timestamp is None and input_to_timestamp is None):
            if SyncUtil.doesnt_contains_historical_data(sync_last_timestamp, sync_doc_type):
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
    def get_gsc_next_sync_infos(last_sync, input_last_timestamp, input_to_timestamp):
        sync_last_timestamp = last_sync.get("last_timestamp")
        first_run = (sync_last_timestamp == 0)
        next_timestamps = []
        if first_run or (input_last_timestamp is None and input_to_timestamp is None):
            next_timestamps = SyncUtil.get_gsc_next_timestamps(sync_last_timestamp)
        elif input_last_timestamp is not None and input_to_timestamp is not None:
            next_timestamps = SyncUtil.get_gsc_next_timestamps(input_last_timestamp, input_to_timestamp)
        else:
            next_timestamps = SyncUtil.get_gsc_next_timestamps(input_last_timestamp)
        next_sync_infos = []
        for next_timestamp in next_timestamps:
            next_sync_info = last_sync.copy()
            next_sync_info['next_timestamp'] = next_timestamp
            next_sync_info['first_run'] = first_run
            next_sync_infos.append(next_sync_info)
        return next_sync_infos


    @staticmethod
    def doesnt_contains_historical_data(last_timestamp, doc_type):
        adwords_timestamp_today = TimeUtil.get_timestamp_from_datetime(datetime.utcnow())
        non_report_related = SyncUtil.non_historical_doc_type(doc_type)
        return non_report_related and last_timestamp != adwords_timestamp_today

    @staticmethod
    def non_historical_doc_type(doc_type):
        return doc_type in [CAMPAIGNS, ADS, AD_GROUPS, CUSTOMER_ACCOUNT_PROPERTIES]

    @staticmethod
    def get_next_timestamps(last_timestamp, to_timestamp=None):
        start_timestamp = SyncUtil.get_next_start_time(last_timestamp)
        return TimeUtil.get_timestamp_range(start_timestamp, to_timestamp)

    @staticmethod
    def get_gsc_next_timestamps(last_timestamp, to_timestamp=None):
        start_timestamp = SyncUtil.get_next_start_time(last_timestamp)
        return TimeUtil.get_gsc_timestamp_range(start_timestamp, to_timestamp)

    @staticmethod
    def get_next_timestamp(last_timestamp):
        start_timestamp = SyncUtil.get_next_start_time(last_timestamp)
        return [start_timestamp]

    @staticmethod
    def get_next_start_time(last_timestamp):
        max_look_back_timestamp = SyncUtil.get_max_look_back_timestamp()
        if last_timestamp == 0 or last_timestamp == None:
            start_timestamp = max_look_back_timestamp
        else:
            start_timestamp = TimeUtil.get_next_day_timestamp(last_timestamp)

        return start_timestamp

    @staticmethod
    def get_max_look_back_timestamp():
        return TimeUtil.get_timestamp_before_days(SyncUtil.MAX_LOOK_BACK_DAYS)
