import itertools
import logging as log

from lib.task.system.google_storage import GoogleStorage
from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from lib.utils.facebook.storage_decider import StorageDecider as FacebookStorageDecider
from lib.utils.json import JsonUtil
from lib.utils.sync_util import SyncUtil
from lib.utils.time import TimeUtil
from scripts.facebook import *
from .base_extract import BaseExtract

# TODO somehow make this an interface where people refer for methods
# TODO check Also for commonFields?
# TODO decouple task Context from Execution/Running Context.
# BaseExtract is different from BaseLoad/CampaignInfoLoad.
# IMP - CampaignPeformanceExtract is different from CampaignInfoExtract.
# Splitting adperformance extract into 2 different jobs is possible. But because of backward compatibility and time taken, we are not going ahead.
# Duplicated code across adPerformance and AdsetPerformance
class BaseReportExtract(BaseExtract):
    # Currently not treating this as configuration Related Context.
    NAME = ""
    KEY_FIELDS = []
    FIELDS = []
    SEGMENTS = []
    METRICS_1 = []
    METRICS_2 = []
    LEVEL_BREAKDOWN = ""  # Facebook terminology
    TASK_TYPE = EXTRACT
    UNFORMATTED_URL = 'https://graph.facebook.com/v13.0/{}/insights?breakdowns={' \
                      '}&&action_breakdowns=action_type&&time_range={}&&fields={}&&access_token={}&&level={' \
                      '}&&filtering=[{{\'field\':\'impressions\',\'operator\':\'GREATER_THAN_OR_EQUAL\',\'value\':0}}]&&limit=1000'


    def add_job_running_context(self):
        pass

    # Getters start here.
    def get_name(self):
        return self.NAME

    # fields + metrics.
    def get_fields(self):
        pass

    def get_destinations(self):
        return self.destinations

    def get_records(self):
        return self.records

    def get_segments(self):
        return self.SEGMENTS

    def get_url(self):
        pass

    def get_next_timestamps(self):
        if self.input_from_timestamp is not None and self.input_to_timestamp is not None:
            return SyncUtil.get_next_timestamps(self.input_from_timestamp, self.input_to_timestamp)
        elif self.input_from_timestamp is not None and self.input_to_timestamp is None:
            return SyncUtil.get_next_timestamps(self.input_from_timestamp)
        else:
            return SyncUtil.get_next_timestamps(self.last_timestamp)

    # Check if could be done better.
    def add_destination_attributes(self):
        for destination in self.get_destinations():
            if isinstance(destination, GoogleStorage):
                bucket_name = FacebookStorageDecider.get_bucket_name(self.env, self.dry)
                file_path = FacebookStorageDecider.get_file_path(self.curr_timestamp, self.project_id,
                                                                 self.customer_account_id, self.type_alias)
                destination.set_attributes({"bucket_name": bucket_name, "file_path": file_path, "file_override": True})
            else:
                file_path = FacebookStorageDecider.get_file_path(self.curr_timestamp, self.project_id,
                                                                 self.customer_account_id, self.type_alias)
                destination.set_attributes(
                    {"base_path": "/usr/local/var/factors/cloud_storage/", "file_path": file_path, "file_override": True})
        return

    def read_records(self):
        self.add_source_attributes_for_metrics1()
        resp_status = self.read_records_for_current_columns_and_update_metrics()
        if resp_status != "success":
            return resp_status
        records_with_metrics1 = self.records
        self.add_source_attributes_for_metrics2()
        resp_status = self.read_records_for_current_columns_and_update_metrics()
        if resp_status != "success":
            return resp_status
        transformed_records = self.transform_array_metrics_with_action_type(self.records)
        records_with_metrics2 = transformed_records
        self.records = self.merge_records_of_metrics1_and_2(records_with_metrics1, records_with_metrics2)
        return "success"

    def write_records(self):
        # log.warning("Writing to destination for the following date: %s", self.curr_timestamp)
        records_string = JsonUtil.create(self.records)
        for destination in self.get_destinations():
            destination.write(records_string)
        return

    def add_source_attributes_for_metrics1(self):
        url = self.get_url_for_extract1()
        attributes = {"url": url}
        self.source.set_attributes(attributes)
        return

    def get_url_for_extract1(self):
        curr_timestamp_in_string = TimeUtil.get_string_of_specific_format_from_timestamp(self.curr_timestamp,
                                                                                         '%Y-%m-%d')
        time_range = {'since': curr_timestamp_in_string, 'until': curr_timestamp_in_string}
        url_ = self.UNFORMATTED_URL.format(self.customer_account_id, time_range, self.get_fields_for_extract1(),
                                           self.int_facebook_access_token, self.LEVEL_BREAKDOWN)
        return url_

    # fields + metrics1.
    def get_fields_for_extract1(self):
        return list(itertools.chain(self.KEY_FIELDS, self.METRICS_1))

    def add_source_attributes_for_metrics2(self):
        url = self.get_url_for_extract2()
        attributes = {"url": url}
        self.source.set_attributes(attributes)
        return

    def get_url_for_extract2(self):
        curr_timestamp_in_string = TimeUtil.get_string_of_specific_format_from_timestamp(self.curr_timestamp,
                                                                                         '%Y-%m-%d')
        time_range = {'since': curr_timestamp_in_string, 'until': curr_timestamp_in_string}
        url_ = self.UNFORMATTED_URL.format(self.customer_account_id, time_range, self.get_fields_for_extract2(),
                                           self.int_facebook_access_token, self.LEVEL_BREAKDOWN)
        return url_

    # fields + metrics2.
    def get_fields_for_extract2(self):
        return list(itertools.chain(self.KEY_FIELDS, self.FIELDS, self.METRICS_2))

    # Read records gives response of status of message
    def read_records_for_current_columns_and_update_metrics(self):
        records_string, result_response, current_no_of_requests = self.source.read()
        self.total_number_of_records += current_no_of_requests
        if not result_response.ok:
            log.warning(ERROR_MESSAGE.format(self.get_name(), result_response.status_code, result_response.text,
                                 self.project_id))
            MetricsAggregator.update_job_stats(self.project_id, self.customer_account_id,
                                               self.type_alias, "failed", result_response.text)
            return "failed"
        else:
            MetricsAggregator.update_job_stats(self.project_id, self.customer_account_id,
                                               self.type_alias, "success", "")
            self.records = JsonUtil.read(records_string)
            return "success"

    def transform_array_metrics_with_action_type(self, records):
        new_records = []
        keys= [COST_PER_ACTION_TYPE, WEBSITE_PURCHASE_ROAS]
        for record in records:
            for key in keys:
                if key in record:
                    for action_type_value in record[key]:
                        record[key + "_" + action_type_value[ACTION_TYPE]] = action_type_value[VALUE]
            new_records.append(record)
        return new_records
