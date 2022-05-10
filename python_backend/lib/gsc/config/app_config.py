from lib.adwords.cors import Cors
from lib.config import Config

SESSION_COOKIE_NAME = "factors-sid"


class AppConfig(Config):
    env = None
    app_host_url = None
    data_service_url = None  # Check: This vs better name.

    @staticmethod
    def __validate(argv):
        pass

    @classmethod
    def __init(cls, env, data_service_url, app_host_url):
        cls.env = env
        cls.data_service_url = data_service_url
        cls.app_host_url = app_host_url
        Cors.set_acceptable_origins(env)
        cls.cors = Cors

    @classmethod
    def build(cls, argv):
        cls.__validate(argv)
        cls.__init(argv.env, argv.api_host_url, argv.app_host_url)

    @classmethod
    def get_session_cookie_key(cls):
        if cls.env == "production":
            return SESSION_COOKIE_NAME
        elif cls.env == "staging":
            return SESSION_COOKIE_NAME + "s"
        else:
            return SESSION_COOKIE_NAME + "d"

    @classmethod
    def get_factors_login_redirect_url(cls):
        return cls.app_host_url + "/login"

    @classmethod
    def get_factors_admin_gsc_redirect_url(cls, status=None):
        url = cls.app_host_url + "/settings/integration"
        if status is not None:
            url = url + "?status=" + status
        return url

    @classmethod
    def get_app_host_url(cls):
        return cls.app_host_url

    @classmethod
    def get_data_service_path(cls):
        return cls.data_service_url + "/data_service"
