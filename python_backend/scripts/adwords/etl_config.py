from lib.adwords.config.etl.app_config import AppConfig as AdwordsAppConfig
from lib.adwords.config.etl.oauth_config import OauthConfig as AdwordsOauthConfig


class EtlConfig:
    ADWORDS_APP = None
    ADWORDS_OAUTH = None
    # LOGGER = None

    @classmethod
    def build(cls, argv):
        cls.build_adwords(argv)
        # cls.build_logger()

    @classmethod
    def build_adwords(cls, argv):
        AdwordsAppConfig.build(argv)
        AdwordsOauthConfig.build(argv)
        cls.ADWORDS_APP = AdwordsAppConfig
        cls.ADWORDS_OAUTH = AdwordsOauthConfig

    # TODO: Set later
    # @classmethod
    # def build_logger(cls):
    #     logger = logging.getLogger('etl')
    #     logger.setLevel(logging.DEBUG)
    #     handler = logging.StreamHandler()
    #     handler.setLevel(logging.DEBUG)
    #     formatter = logging.Formatter('%(created)f:%(levelname)s:%(name)s:%(module)s:%(message)s')
    #     handler.setFormatter(formatter)
    #     logger.addHandler(handler)
    #     cls.LOGGER = logger
    #     return
