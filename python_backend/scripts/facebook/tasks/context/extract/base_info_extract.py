from datetime import datetime
import logging as log

from lib.task.system.google_storage import GoogleStorage
from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from lib.utils.facebook.storage_decider import StorageDecider as FacebookStorageDecider
from lib.utils.json import JsonUtil
from lib.utils.sync_util import SyncUtil
from lib.utils.time import TimeUtil
from scripts.facebook import ERROR_MESSAGE
from .base_extract import BaseExtract


class BaseInfoExtract(BaseExtract):
    NAME = ""
    FIELDS = []
    type_alias = ""  # Facebook terminology
    UNFORMATTED_URL = "https://graph.facebook.com/v13.0/{}/{}s?fields={}&&access_token={}&&limit=1000"
    BACKFILL_SUPPORTED = True
    records = None

    # Getters start here.
    def get_name(self):
        return self.NAME

    # fields.
    def get_fields(self):
        return self.FIELDS

    def get_url(self):
        return self.UNFORMATTED_URL.format(self.customer_account_id, self.type_alias.replace('_', ''),
                                           self.get_fields(),
                                           self.int_facebook_access_token)

    def get_destinations(self):
        return self.destinations

    def run_job(self):
        if self.input_from_timestamp is not None or self.input_to_timestamp is not None:
            return True
        elif self.get_next_timestamp() > self.last_timestamp:
            return True
        return False

    def get_next_timestamp(self):
        return TimeUtil.get_timestamp_from_datetime(datetime.utcnow())

    def get_next_timestamps(self):
        if self.input_from_timestamp is not None and self.input_to_timestamp is not None:
            return SyncUtil.get_next_timestamps(self.input_from_timestamp, self.input_to_timestamp)
        elif self.input_from_timestamp is not None and self.input_to_timestamp is None:
            return SyncUtil.get_next_timestamps(self.input_from_timestamp, int(datetime.utcnow().strftime('%Y%m%d')))
        elif self.is_eligible_for_backfill():
            return SyncUtil.get_next_timestamps(0, int(datetime.utcnow().strftime('%Y%m%d')))
        else:
            min_timestamp_to_fill = min(self.project_min_timestamp, self.last_timestamp)
            return SyncUtil.get_next_timestamps(min_timestamp_to_fill, int(datetime.utcnow().strftime('%Y%m%d')))

    def is_first_run(self):
        return self.last_timestamp == 0 or self.last_timestamp < self.get_max_look_back_timestamp()

    def is_eligible_for_backfill(self):
        return self.BACKFILL_SUPPORTED and self.is_first_run()

    def get_max_look_back_timestamp(self):
        return SyncUtil.get_max_look_back_timestamp()

    def add_source_attributes(self):
        url = self.get_url()
        attributes = {"url": url, "access_token": self.int_facebook_access_token}
        self.source.set_attributes(attributes)
        return

    # Check if could be done better.
    def add_destination_attributes(self):
        for destination in self.get_destinations():
            if isinstance(destination, GoogleStorage):
                bucket_name = FacebookStorageDecider.get_bucket_name(self.env, self.dry)
                file_path = FacebookStorageDecider.get_file_path(self.curr_timestamp, self.project_id,
                                                                 self.customer_account_id, self.type_alias)
                destination.set_attributes({"bucket_name": bucket_name, "file_path": file_path, "file_override": False})
            else:
                file_path = FacebookStorageDecider.get_file_path(self.curr_timestamp, self.project_id,
                                                                 self.customer_account_id, self.type_alias)
                destination.set_attributes(
                    {"base_path": "/usr/local/var/factors/cloud_storage/", "file_path": file_path, "file_override": False})
        return

    def read_records(self):
        records_string, result_response, current_no_of_requests, async_requests = self.source.read()
        self.total_number_of_records += current_no_of_requests 
        self.total_number_of_async_requests += async_requests
        if not result_response.ok:
            error_msg = result_response.text
            if 'error' in result_response.json():
                # error_subcode is not always present
                error_msg = 'Message: {}, code: {}'.format(result_response.json()['error']['message'], result_response.json()['error']['code'])
            log.warning(ERROR_MESSAGE.format(self.get_name(), result_response.status_code, error_msg,
                                 self.project_id, self.customer_account_id))
            MetricsAggregator.update_job_stats(self.project_id, self.customer_account_id,
                                               self.type_alias, "failed", error_msg)
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
