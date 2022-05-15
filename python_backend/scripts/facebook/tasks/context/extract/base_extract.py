import logging as log

from scripts.facebook import *


# Currently, No decouple present for read and write.
# Currently not treating this as configuration Related Context.
# Later add job running status.
class BaseExtract:
    TASK_TYPE = EXTRACT
    project_id = None
    customer_account_id = None
    type_alias = None
    last_timestamp = None
    int_facebook_user_id = None
    int_facebook_access_token = None
    int_facebook_email = None
    env = None
    dry = None
    input_from_timestamp = None
    input_to_timestamp = None
    curr_timestamp = None
    facebook_data_service_path = None
    project_min_timestamp = None

    source = None
    destination = None
    total_number_of_records = 0
    total_number_of_async_requests = 0

    # Setters start here.
    def add_last_sync_info(self, last_sync_info):
        self.project_id = last_sync_info.get(PROJECT_ID)
        self.customer_account_id = last_sync_info.get(CUSTOMER_ACCOUNT_ID)
        self.type_alias = last_sync_info.get("type_alias")
        self.last_timestamp = last_sync_info.get(LAST_TIMESTAMP)
        return

    def add_facebook_settings(self, facebook_settings):
        self.int_facebook_user_id = facebook_settings.get(INT_FACEBOOK_USER_ID)
        self.int_facebook_access_token = facebook_settings.get(INT_FACEBOOK_ACCESS_TOKEN)
        self.int_facebook_email = facebook_settings.get(INT_FACEBOOK_EMAIL)
        return

    def add_env(self, env):
        self.env = env

    def add_dry(self, dry):
        self.dry = dry

    def add_source(self, source):
        self.source = source

    def add_destinations(self, destinations):
        self.destinations = destinations

    def add_input_from_timestamp(self, input_from_timestamp):
        self.input_from_timestamp = input_from_timestamp

    def add_input_to_timestamp(self, input_to_timestamp):
        self.input_to_timestamp = input_to_timestamp

    def add_curr_timestamp(self, curr_timestamp):
        self.curr_timestamp = curr_timestamp

    def add_facebook_data_service_path(self, facebook_data_service_path):
        self.facebook_data_service_path = facebook_data_service_path

    # Used in info task. Min of (info, project_reports) is used.
    def add_project_min_timestamp(self, project_min_timestamp):
        self.project_min_timestamp = project_min_timestamp

    def add_log(self, running_status):
        log.warning("%s extract of job for Project Id: %s, Customer account id: %s Timestamp: %d, Doc Type: %s", running_status,
                    self.project_id, self.customer_account_id, self.curr_timestamp, self.type_alias)

    def reset_total_number_of_records(self):
        self.total_number_of_records = 0

    def reset_total_number_of_async_requests(self):
        self.total_number_of_async_requests = 0