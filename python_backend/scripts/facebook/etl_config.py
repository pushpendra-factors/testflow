from lib.facebook.config.etl.app_config import AppConfig as FacebookAppConfig


class EtlConfig:
    FACEBOOK_APP = None

    @classmethod
    def build(cls, argv):
        cls.build_facebook(argv)

    @classmethod
    def build_facebook(cls, argv):
        FacebookAppConfig.build(argv)
        cls.FACEBOOK_APP = FacebookAppConfig
