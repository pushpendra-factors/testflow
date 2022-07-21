# from datetime import time
import time
from datetime import datetime, timedelta


class TimeUtil:

    @staticmethod
    def is_today(timestamp):
        today_timestamp = int(time.strftime("%Y%m%d"))
        return timestamp == today_timestamp

    @staticmethod
    def get_timestamp_from_datetime(dt):
        if dt is None:
            return
        return int(dt.strftime('%Y%m%d'))

    @staticmethod
    def get_string_of_specific_format_from_timestamp(timestamp, fmt):
        if timestamp is None:
            return
        curr_date = datetime.strptime(str(timestamp), "%Y%m%d")
        return curr_date.strftime(fmt)

    @staticmethod
    def get_datetime_from_timestamp(timestamp):
        if timestamp is None:
            return
        return datetime.strptime(str(timestamp), "%Y%m%d")

    @staticmethod
    def get_next_day_timestamp(timestamp):
        start_datetime = TimeUtil.get_datetime_from_timestamp(timestamp)
        return TimeUtil.get_timestamp_from_datetime(start_datetime + timedelta(days=1))

    @staticmethod
    def get_timestamp_before_days(days):
        return TimeUtil.get_timestamp_from_datetime(
            datetime.utcnow() - timedelta(days=days))

    @staticmethod
    def get_timestamp_range(from_timestamp, to_timestamp=None):
        date_range = []
        if from_timestamp is None:
            return date_range
        
        # if to_timestamp not given: range till yesterday. 
        if to_timestamp is None:
            to_timestamp = TimeUtil.get_timestamp_before_days(1)

        start_timestamp = from_timestamp
        while start_timestamp <= to_timestamp:
            date_range.append(start_timestamp)
            start_timestamp = TimeUtil.get_next_day_timestamp(start_timestamp)
        
        return date_range

    @staticmethod
    def get_gsc_timestamp_range(from_timestamp, to_timestamp=None):
        date_range = []
        if from_timestamp is None:
            return date_range
        
        # if to_timestamp not given: range 3 days before. 
        if to_timestamp is None:
            to_timestamp = TimeUtil.get_timestamp_before_days(3)

        start_timestamp = from_timestamp
        while start_timestamp <= to_timestamp:
            date_range.append(start_timestamp)
            start_timestamp = TimeUtil.get_next_day_timestamp(start_timestamp)
        return date_range

    @staticmethod
    def convert_timestamp_to_gsc_date_parameter(dt):
        return datetime.strftime(datetime.strptime(str(dt), "%Y%m%d"), "%Y-%m-%d")

    @staticmethod
    def get_difference_from_current_day(dt):
        input_date_obj = datetime.strptime(dt, "%Y-%m-%d")
        current_date_obj = datetime.now()
        return ((current_date_obj-input_date_obj).days)

