from lib.config import Config

SESSION_COOKIE_NAME = "factors-sid"


class AppConfig(Config):
    env = None
    dry = None
    skip_today = None
    project_ids = None
    exclude_project_ids = None
    document_type = None
    last_timestamp = None
    data_service_host = None


    @classmethod
    def _init(cls, env, dry, skip_today, project_ids, exclude_project_ids, document_type, last_timestamp, data_service_host):
        cls.env = env
        cls.dry = (dry == "True")
        cls.skip_today = (skip_today == "True")
        cls.project_ids = project_ids
        cls.exclude_project_ids = exclude_project_ids
        cls.document_type = document_type
        cls.last_timestamp = last_timestamp
        cls.data_service_host = data_service_host

    @classmethod
    def build(cls, argv):
        project_ids = set()
        exclude_project_ids = set()
        if argv.project_id is not None:
            project_ids = set([int(x) for x in argv.project_id.split(",")])
        if argv.exclude_project_id is not None:
            exclude_project_ids = set([int(x) for x in argv.exclude_project_id.split(",")])

        cls._init(argv.env, argv.dry, argv.skip_today,
                  project_ids, exclude_project_ids, argv.document_type, argv.last_timestamp, argv.data_service_host)

    @classmethod
    def get_session_cookie_key(cls):
        if cls.env == "production":
            return SESSION_COOKIE_NAME
        elif cls.env == "staging":
            return SESSION_COOKIE_NAME + "s"
        else:
            return SESSION_COOKIE_NAME + "d"

    # @classmethod
    # def get_factors_login_redirect_url(cls):
    #     return cls._app_host_url + "/#/login"
    #
    # @classmethod
    # def get_factors_admin_adwords_redirect_url(cls, status=None):
    #     url = cls._app_host_url + "/#/settings/adwords"
    #     if status is not None:
    #         url = url + "?status=" + status
    #     return url
    #
    # @classmethod
    # def get_app_host_url(cls):
    #     return cls._app_host_url

    @classmethod
    def get_data_service_path(cls):
        return cls.data_service_host + '/data_service'
