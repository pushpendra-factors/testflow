import sys

from tornado.log import logging as log
from lib.config import Config


class SentryConfig(Config):
    dsn = None
    environment = None
    server_name = None
    attach_stacktrace = True

    @classmethod
    def build(cls, argv):
        cls._validate(argv)
        cls._init(argv.sentry_dsn, argv.env, argv.server_name)

    @staticmethod
    def _validate(argv):
        sentry_dsn = argv.sentry_dsn
        if sentry_dsn == "":
            log.error("Option: sentry_dsn cannot be empty.")
            sys.exit(1)

    @classmethod
    def _init(cls, dsn, environment, server_name):
        cls.dsn = dsn
        cls.environment = environment
        cls.server_name = server_name
