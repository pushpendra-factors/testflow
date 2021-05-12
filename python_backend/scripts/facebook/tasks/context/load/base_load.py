# Context Load
import logging as log

from lib.utils.sync_util import SyncUtil
from scripts.facebook import *


class BaseLoad:
    TASK_TYPE = LOAD

    project_id = None
    customer_account_id = None
    last_timestamp = None
    type_alias = None
    # workflow = None
    env = None
    dry = None
    input_from_timestamp = None
    input_to_timestamp = None
    facebook_data_service_path = None

    source = None
    destinations = None
    curr_timestamp = None

    # Setters start here.
    def add_last_sync_info(self, last_sync_info):
        self.project_id = last_sync_info.get(PROJECT_ID)
        self.customer_account_id = last_sync_info.get(CUSTOMER_ACCOUNT_ID)
        self.type_alias = last_sync_info.get(TYPE_ALIAS)
        self.last_timestamp = last_sync_info.get(LAST_TIMESTAMP)
        return

    def add_facebook_settings(self, facebook_settings):
        return

    def add_env(self, env):
        self.env = env

    def add_dry(self, dry):
        self.dry = dry

    def add_source(self, source):
        self.source = source

    def add_destinations(self, destinations):
        self.destinations = destinations

    def is_first_run(self):
        return self.last_timestamp == 0

    def get_next_timestamps(self):
        return SyncUtil.get_next_timestamps(self.last_timestamp)

    def get_source(self):
        return self.source

    def get_destinations(self):
        return self.destinations

    def add_input_from_timestamp(self, input_from_timestamp):
        self.input_from_timestamp = input_from_timestamp

    def add_input_to_timestamp(self, input_to_timestamp):
        self.input_to_timestamp = input_to_timestamp

    def add_curr_timestamp(self, curr_timestamp):
        self.curr_timestamp = curr_timestamp

    def add_facebook_data_service_path(self, facebook_data_service_path):
        self.facebook_data_service_path = facebook_data_service_path

    def add_log(self, running_status):
        log.warning("%s load of job for Project Id: %s, Timestamp: %d, Doc Type: %s", running_status,
                    self.project_id, self.curr_timestamp, self.type_alias)

    # This is used only in extract
    def add_project_min_timestamp(self, project_min_timestamp):
        pass
