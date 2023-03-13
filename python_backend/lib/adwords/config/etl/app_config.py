from lib.config import Config
from lib.utils.adwords.metrics_controller import MetricsController
from lib.utils.adwords.job_storage import JobStorage
from lib.utils.string import StringUtil

SESSION_COOKIE_NAME = "factors-sid"


class AppConfig(Config):
    env = None
    skip_today = None
    project_ids = None
    exclude_project_ids = None
    document_type = None
    google_project_name = None
    data_service_host = None

    timezone = None
    type_of_run = None
    dry = None
    last_timestamp = None
    to_timestamp = None
    metrics_controller = None
    job_storage = None

    @classmethod
    def _init(cls, env, skip_today, 
              project_ids, exclude_project_ids, document_type, data_service_host,
              timezone, type_of_run, dry, last_timestamp, to_timestamp):
        cls.env = env
        cls.skip_today = (skip_today == "True")
        cls.project_ids = project_ids
        cls.exclude_project_ids = exclude_project_ids
        cls.document_type = document_type
        cls.data_service_host = data_service_host

        cls.timezone = timezone
        cls.type_of_run = type_of_run
        cls.dry = (dry == "True")
        cls.last_timestamp = last_timestamp
        cls.to_timestamp = to_timestamp
        MetricsController.init(type_of_run)
        cls.metrics_controller = MetricsController
        JobStorage.init(cls.env, cls.dry)
        cls.job_storage = JobStorage

    @classmethod
    def build(cls, argv):
        project_ids = set()
        exclude_project_ids = set()
        if argv.project_id is not None:
            project_ids = StringUtil.get_set_from_string_split_by_comma(argv.project_id)
        if argv.exclude_project_id is not None:
            exclude_project_ids = StringUtil.get_set_from_string_split_by_comma(argv.exclude_project_id)
                
        cls._init(argv.env, argv.skip_today,
                  project_ids, exclude_project_ids, argv.document_type, argv.data_service_host,
                  argv.timezone, argv.type_of_run, argv.dry, argv.last_timestamp, argv.to_timestamp)

    @classmethod
    def get_session_cookie_key(cls):
        if cls.env == "production":
            return SESSION_COOKIE_NAME
        elif cls.env == "staging":
            return SESSION_COOKIE_NAME + "s"
        else:
            return SESSION_COOKIE_NAME + "d"

    @classmethod
    def get_data_service_path(cls):
        return cls.data_service_host + "/data_service"
