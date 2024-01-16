from lib.adwords.config.app_config import AppConfig as AdwordsAppConfig
from lib.adwords.config.oauth_config import OauthConfig as AdwordsOauthConfig
from lib.gsc.config.app_config import AppConfig as GSCAppConfig
from lib.gsc.config.oauth_config import OauthConfig as GSCOauthConfig
from lib.data_services.sentry.sentry_config import SentryConfig


class AppConfig:
    ADWORDS_APP = None
    ADWORDS_OAUTH = None
    GSC_APP = None
    GSC_OAUTH = None
    SENTRY = None
    CHAT_BUCKET = None

    @classmethod
    def build(cls, argv):
        cls.build_adwords(argv)
        cls.build_gsc(argv)
        cls.CHAT_BUCKET = argv.chat_bucket_name

    @classmethod
    def build_adwords(cls, argv):
        AdwordsAppConfig.build(argv)
        AdwordsOauthConfig.build(argv)
        SentryConfig.build(argv)
        cls.ADWORDS_APP = AdwordsAppConfig
        cls.ADWORDS_OAUTH = AdwordsOauthConfig
        cls.SENTRY = SentryConfig

    @classmethod
    def build_gsc(cls, argv):
        GSCAppConfig.build(argv)
        GSCOauthConfig.build(argv)
        cls.GSC_APP = GSCAppConfig
        cls.GSC_OAUTH = GSCOauthConfig
