from datetime import datetime

from googleads import adwords

import scripts
from lib.adwords.oauth_service.fetch_service import FetchService
from lib.utils.csv import CsvUtil
from lib.utils.string import StringUtil
from lib.utils.time import TimeUtil
from .base_job import BaseJob
# Note: If the number of custom paths exceed 5 in the subClasses. Move it to strategic pattern.


class ReportsFetch(BaseJob):
    QUERY_FIELDS = []
    REPORT = ''
    WHERE_IN_COLUMN = 'CampaignStatus'
    WHERE_IN_VALUES = ['ENABLED', 'PAUSED']

    def __init__(self, next_info):
        super().__init__(next_info)

    def start(self):
        str_timestamp = str(self._timestamp)
        during = str_timestamp + ',' + str_timestamp
        downloader = FetchService(scripts.adwords.CONFIG.ADWORDS_OAUTH).get_report_downloader(self._refresh_token, self._customer_account_id)
        fields = StringUtil.snake_to_pascal_case(self.QUERY_FIELDS)

        report_query = (adwords.ReportQueryBuilder()
                        .Select(*fields)
                        .From(self.REPORT)
                        .Where(self.WHERE_IN_COLUMN).In(*self.WHERE_IN_VALUES)
                        .During(during).Build())

        report = downloader.DownloadReportAsStringWithAwql(
            report_query, 'CSV', skip_report_header=True,
            skip_column_header=True, skip_report_summary=True)

        lines = report.split('\n')
        return CsvUtil.csv_to_dict_list(self.QUERY_FIELDS, lines), 1

    @staticmethod
    def contains_historical_data(last_timestamp, doc_type):
        adwords_timestamp_today = TimeUtil.get_timestamp_from_datetime(datetime.utcnow())
        non_report_related = doc_type in ['campaigns', 'ads', 'ad_groups', 'customer_account_properties']
        return non_report_related and last_timestamp != adwords_timestamp_today

    @staticmethod
    def get_next_start_time_for_historical_data(last_timestamp):
        max_look_back_timestamp = TimeUtil.get_timestamp_before_days(30)
        if last_timestamp == 0 or last_timestamp < max_look_back_timestamp:
            start_timestamp = max_look_back_timestamp
        else:
            start_timestamp = TimeUtil.get_next_day_timestamp(last_timestamp)

        return start_timestamp
