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

        dt_year = str(dt.year)
        if len(dt_year) == 1:
            dt_year = "0"+dt_year

        dt_month = str(dt.month)
        if len(dt_month) == 1:
            dt_month = "0"+dt_month

        dt_day = str(dt.day)
        if len(dt_day) == 1:
            dt_day = "0"+dt_day

        return int(dt_year+dt_month+dt_day)

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
