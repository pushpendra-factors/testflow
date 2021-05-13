from lib.gsc.config.etl.app_config import AppConfig
from lib.gsc.config.etl.oauth_config import OauthConfig


class EtlConfig:
    GSC_APP = None
    GSC_OAUTH = None

    # LOGGER = None

    @classmethod
    def build(cls, argv):
        cls.build_gsc(argv)
        # cls.build_logger()

    @classmethod
    def build_gsc(cls, argv):
        AppConfig.build(argv)
        OauthConfig.build(argv)
        cls.GSC_APP = AppConfig
        cls.GSC_OAUTH = OauthConfig

    # TODO: Set later
    # @classmethod
    # def build_logger(cls):
    #     logger = logging.getLogger("etl")
    #     logger.setLevel(logging.DEBUG)
    #     handler = logging.StreamHandler()
    #     handler.setLevel(logging.DEBUG)
    #     formatter = logging.Formatter("%(created)f:%(levelname)s:%(name)s:%(module)s:%(message)s")
    #     handler.setFormatter(formatter)
    #     logger.addHandler(handler)
    #     cls.LOGGER = logger
    #     return
