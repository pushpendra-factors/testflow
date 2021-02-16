from lib.config import Config

SESSION_COOKIE_NAME = "factors-sid"


class AppConfig(Config):
    env = None
    dry = None
    skip_today = None
    data_service_host = None
    project_id = None
    extract_schema_changed = None

    @classmethod
    def _init(cls, env, dry, extract_schema_changed, skip_today, project_id, data_service_host):
        cls.env = env
        cls.dry = (dry == "True")
        cls.extract_schema_changed = (extract_schema_changed == "True")
        cls.skip_today = (skip_today == "True")
        cls.project_id = project_id
        cls.data_service_host = data_service_host

    @classmethod
    def build(cls, argv):
        cls._init(argv.env, argv.dry, argv.extract_schema_changed, argv.skip_today,
                  argv.project_id, argv.data_service_host)

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
