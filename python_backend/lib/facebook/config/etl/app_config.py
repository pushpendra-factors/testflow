from lib.config import Config
from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from lib.utils.string import StringUtil


class AppConfig(Config):
    env = None
    project_ids = None
    exclude_project_ids = None
    document_type = None
    data_service_host = None

    type_of_run = None
    dry = None
    last_timestamp = None
    to_timestamp = None
    # To check if we need this to be globally available as service.
    metrics_controller = None
    job_storage = None

    @classmethod
    def _init(cls, env, project_ids, exclude_project_ids, document_type, data_service_host,
              type_of_run, dry, last_timestamp, to_timestamp):
        cls.env = env
        cls.project_ids = project_ids
        cls.exclude_project_ids = exclude_project_ids
        cls.document_type = document_type
        cls.data_service_host = data_service_host

        cls.type_of_run = type_of_run
        cls.dry = (dry == "True")
        cls.last_timestamp = last_timestamp
        cls.to_timestamp = to_timestamp
        MetricsAggregator.init(env, type_of_run)
        cls.metrics_controller = MetricsAggregator

    @classmethod
    def build(cls, argv):
        project_ids = set()
        exclude_project_ids = set()
        if argv.project_id is not None:
            project_ids = StringUtil.get_set_from_string_split_by_comma(argv.project_id)
        if argv.exclude_project_id is not None:
            exclude_project_ids = StringUtil.get_set_from_string_split_by_comma(argv.exclude_project_id)

        cls._init(argv.env, project_ids, exclude_project_ids, argv.document_type, argv.data_service_host,
                  argv.type_of_run, argv.dry, argv.last_timestamp, argv.to_timestamp)

    @classmethod
    def get_data_service_path(cls):
        return cls.data_service_host + "/data_service"
