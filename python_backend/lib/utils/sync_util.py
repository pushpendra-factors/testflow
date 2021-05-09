from lib.utils.time import TimeUtil


class SyncUtil:
    MAX_LOOK_BACK_DAYS = 30

    @staticmethod
    def get_next_timestamps(last_timestamp, to_timestamp=None):
        start_timestamp = SyncUtil.get_next_start_time(last_timestamp)
        return TimeUtil.get_timestamp_range(start_timestamp, to_timestamp)

    @staticmethod
    def get_next_timestamp(last_timestamp):
        start_timestamp = SyncUtil.get_next_start_time(last_timestamp)
        return [start_timestamp]

    @staticmethod
    def get_next_start_time(last_timestamp):
        max_look_back_timestamp = SyncUtil.get_max_look_back_timestamp()
        if last_timestamp == 0 or last_timestamp < max_look_back_timestamp:
            start_timestamp = max_look_back_timestamp
        else:
            start_timestamp = TimeUtil.get_next_day_timestamp(last_timestamp)

        return start_timestamp

    @staticmethod
    def get_max_look_back_timestamp():
        return TimeUtil.get_timestamp_before_days(SyncUtil.MAX_LOOK_BACK_DAYS)
