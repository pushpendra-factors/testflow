from lib.gsc.config.app_config import AppConfig as GSCAppConfig
from lib.gsc.config.oauth_config import OauthConfig as GSCOauthConfig
from lib.data_services.sentry.sentry_config import SentryConfig


class AppConfig:
    GSC_APP = None
    GSC_OAUTH = None
    SENTRY = None

    @classmethod
    def build(cls, argv):
        cls.build_gsc(argv)

    @classmethod
    def build_gsc(cls, argv):
        GSCAppConfig.build(argv)
        GSCOauthConfig.build(argv)
        SentryConfig.build(argv)
        cls.GSC_APP = GSCAppConfig
        cls.GSC_OAUTH = GSCOauthConfig
        cls.SENTRY = SentryConfig
