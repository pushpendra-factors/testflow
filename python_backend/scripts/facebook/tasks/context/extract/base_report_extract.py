import itertools
import logging as log

from lib.task.system.google_storage import GoogleStorage
from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from lib.utils.facebook.storage_decider import StorageDecider as FacebookStorageDecider
from lib.utils.json import JsonUtil
from lib.utils.sync_util import SyncUtil
from lib.utils.time import TimeUtil
from scripts.facebook import *
# TODO somehow make this an interface where people refer for methods
# TODO check Also for commonFields?
# TODO decouple task Context from Execution/Running Context.
# BaseExtract is different from BaseLoad/CampaignInfoLoad.
# IMP - CampaignPeformanceExtract is different from CampaignInfoExtract.
from .base_extract import BaseExtract


class BaseReportExtract(BaseExtract):
    # Currently not treating this as configuration Related Context.
    NAME = ""
    FIELDS = []
    SEGMENTS = []
    METRICS = []
    LEVEL_BREAKDOWN = ""  # Facebook terminology
    TASK_TYPE = EXTRACT
    UNFORMATTED_URL = 'https://graph.facebook.com/v9.0/{}/insights?breakdowns={' \
                      '}&&action_breakdowns=action_type&&time_range={}&&fields={}&&access_token={}&&level={' \
                      '}&&filtering=[{{\'field\':\'impressions\',\'operator\':\'GREATER_THAN_OR_EQUAL\',\'value\':0}}]&&limit=1000'


    def add_job_running_context(self):
        pass

    # Getters start here.
    def get_name(self):
        return self.NAME

    # fields + metrics.
    def get_fields(self):
        return list(itertools.chain(self.FIELDS, self.METRICS))

    def get_destinations(self):
        return self.destinations

    def get_records(self):
        return self.records

    def get_segments(self):
        return self.SEGMENTS

    def get_url(self):
        curr_timestamp_in_string = TimeUtil.get_string_of_specific_format_from_timestamp(self.curr_timestamp,
                                                                                         '%Y-%m-%d')
        time_range = {'since': curr_timestamp_in_string, 'until': curr_timestamp_in_string}
        url_ = self.UNFORMATTED_URL.format(self.customer_account_id, self.get_segments(), time_range, self.get_fields(),
                                           self.int_facebook_access_token, self.LEVEL_BREAKDOWN)
        return url_

    def get_next_timestamps(self):
        if self.input_from_timestamp is not None and self.input_to_timestamp is not None:
            return SyncUtil.get_next_timestamps(self.input_from_timestamp, self.input_to_timestamp)
        elif self.input_from_timestamp is not None and self.input_to_timestamp is None:
            return SyncUtil.get_next_timestamps(self.input_from_timestamp)
        else:
            return SyncUtil.get_next_timestamps(self.last_timestamp)

    def add_source_attributes(self):
        url = self.get_url()
        attributes = {"url": url}
        self.source.set_attributes(attributes)
        return

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

    # Read records gives response of status of message
    def read_records(self):
        records_string, result_response = self.source.read()
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

    def write_records(self):
        file_path = FacebookStorageDecider.get_file_path(self.curr_timestamp, self.project_id,
                                                                 self.customer_account_id, self.type_alias)
        # log.warning("Writing to destination for the following date: %s", self.curr_timestamp)
        records_string = JsonUtil.create(self.records)
        for destination in self.get_destinations():
            destination.write(records_string)
        return
