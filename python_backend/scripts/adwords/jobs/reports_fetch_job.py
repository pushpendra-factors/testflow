from datetime import datetime

from googleads import adwords
import operator

import scripts
from lib.adwords.oauth_service.fetch_service import FetchService
from lib.utils.csv import CsvUtil
from lib.utils.format import FormatUtil
from lib.utils.string import StringUtil
from lib.utils.time import TimeUtil
from .base_job import BaseJob
from .. import CAMPAIGNS, ADS, AD_GROUPS, CUSTOMER_ACCOUNT_PROPERTIES


# Note: If the number of custom paths exceed 5 in the subClasses. Move it to strategic pattern.
class ReportsFetch(BaseJob):
    QUERY_FIELDS = []
    REPORT = ''
    WHERE_IN_COLUMN = 'CampaignStatus'
    WHERE_IN_VALUES = ['ENABLED', 'PAUSED']
    MAX_LOOK_BACK_DAYS = 30
    DEFUALT_FLOAT = 0.000
    DEFAULT_NUMERATOR_FLOAT = 0.0
    DEFAULT_DENOMINATOR_FLOAT = 1.0
    DEFAULT_DECIMAL_PLACES = 3

    # Currently only commonFields Transformation is being done and Expression = Total--- = impressions/share---
    OTHER_FIELD_NAME = 'other_field_name'
    RESULT_FIELD = 'result_field'
    OPERATION = 'operation'
    TRANSFORM_TO_FORM_NEW_FIELDS = {
       'search_impression_share': { RESULT_FIELD: 'total_search_impression', OPERATION: operator.truediv, OTHER_FIELD_NAME: 'impressions'},
       'search_click_share': { RESULT_FIELD: 'total_search_click', OPERATION: operator.truediv, OTHER_FIELD_NAME: 'impressions'},
       'search_top_impression_share': { RESULT_FIELD: 'total_search_top_impression', OPERATION: operator.truediv, OTHER_FIELD_NAME: 'impressions'},
       'search_budget_lost_absolute_top_impression_share': { RESULT_FIELD: 'total_search_budget_lost_absolute_top_impression', OPERATION: operator.truediv, OTHER_FIELD_NAME: 'impressions'},
       'search_budget_lost_impression_share': { RESULT_FIELD: 'total_search_budget_lost_impression', OPERATION: operator.truediv, OTHER_FIELD_NAME: 'impressions'},
       'search_budget_lost_top_impression_share': { RESULT_FIELD: 'total_search_budget_lost_top_impression', OPERATION: operator.truediv, OTHER_FIELD_NAME: 'impressions'},
       'search_rank_lost_absolute_top_impression_share': { RESULT_FIELD: 'total_search_rank_lost_absolute_top_impression', OPERATION: operator.truediv, OTHER_FIELD_NAME: 'impressions'},
       'search_rank_lost_impression_share': { RESULT_FIELD: 'total_search_rank_lost_impression', OPERATION: operator.truediv, OTHER_FIELD_NAME: 'impressions'},
       'search_rank_lost_top_impression_share': { RESULT_FIELD: 'total_search_rank_lost_top_impression', OPERATION: operator.truediv, OTHER_FIELD_NAME: 'impressions'}
    }


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
        rows = CsvUtil.csv_to_dict_list(self.QUERY_FIELDS, lines)
        rows = self.transform_entities(rows)
        return rows, 1

    def transform_entities(self, rows):
        transformed_rows = []
        for row in rows:
            transformed_rows.append(self.transform_entity(row))
        return transformed_rows

    def transform_entity(self, row):
        for field1_name in self.TRANSFORM_TO_FORM_NEW_FIELDS:
            value = self.TRANSFORM_TO_FORM_NEW_FIELDS[field1_name]
            field2_name = value[self.OTHER_FIELD_NAME]
            operation = value[self.OPERATION]
            result_field_name = value[self.RESULT_FIELD]
            if field1_name in row and field2_name in row:
                transformed_value = self.get_transformed_value_for_division_operator(row, field1_name, field2_name, operation)
                row[result_field_name] = transformed_value

        return row

    @staticmethod
    def doesnt_contains_historical_data(last_timestamp, doc_type):
        adwords_timestamp_today = TimeUtil.get_timestamp_from_datetime(datetime.utcnow())
        non_report_related = ReportsFetch.non_historical_doc_type(doc_type)
        return non_report_related and last_timestamp != adwords_timestamp_today

    @staticmethod
    def get_next_sync_infos_for_older_date_range(last_timestamp, last_sync):
        next_sync_info = []
        start_timestamp = ReportsFetch.get_next_start_time_for_historical_data(last_timestamp)
        next_timestamps = TimeUtil.get_timestamp_range(start_timestamp)

        for timestamp in next_timestamps:
            sync_info = last_sync.copy()
            sync_info['next_timestamp'] = timestamp
            next_sync_info.append(sync_info)

        return next_sync_info

    @staticmethod
    def non_historical_doc_type(doc_type):
        return doc_type in [CAMPAIGNS, ADS, AD_GROUPS, CUSTOMER_ACCOUNT_PROPERTIES]

    @staticmethod
    def get_next_start_time_for_historical_data(last_timestamp):
        max_look_back_timestamp = ReportsFetch.get_max_look_back_timestamp()
        if last_timestamp == 0 or last_timestamp < max_look_back_timestamp:
            start_timestamp = max_look_back_timestamp
        else:
            start_timestamp = TimeUtil.get_next_day_timestamp(last_timestamp)

        return start_timestamp

    @staticmethod
    def get_max_look_back_timestamp():
        return TimeUtil.get_timestamp_before_days(ReportsFetch.MAX_LOOK_BACK_DAYS)

    @staticmethod
    def get_transformed_value_for_division_operator(row, field1, field2, operation):
        field1_value = FormatUtil.get_numeric_from_percentage_string(row.get(field1, ReportsFetch.DEFAULT_NUMERATOR_FLOAT))
        field2_value = FormatUtil.get_numeric_from_percentage_string(row.get(field2, ReportsFetch.DEFAULT_DENOMINATOR_FLOAT))
        if field1_value == 0 or field2_value == 0:
            return ReportsFetch.DEFUALT_FLOAT
        else:
            return round(operation(field1_value, field2_value), ReportsFetch.DEFAULT_DECIMAL_PLACES)
