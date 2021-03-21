import logging

import sentry_sdk
from sentry_sdk.integrations.logging import LoggingIntegration
from sentry_sdk.integrations.tornado import TornadoIntegration


# TODO: Disable development
# This is not capturing 400/no endpoint errors.
class SentryDataService:
    CONFIG = None

    @classmethod
    def init(cls, config):
        if config.environment in ["staging", "development"]:
            return

        cls.CONFIG = config
        sentry_logging = LoggingIntegration(
            level=logging.ERROR,
            event_level=logging.ERROR
        )
        sentry_sdk.init(
            dsn=config.dsn,
            environment=config.environment,
            server_name=config.server_name,
            attach_stacktrace=config.attach_stacktrace,
            integrations=[sentry_logging, TornadoIntegration()]
        )

        with sentry_sdk.configure_scope() as scope:
            scope.set_tag("AppName", config.server_name)
