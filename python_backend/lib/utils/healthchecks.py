import json

import requests
import logging as log

class HealthChecksUtil:
    ADWORDS_SYNC_PING_ID = "188cbf7c-0ea1-414b-bf5c-eee47c12a0c8"

    @staticmethod
    def ping(env, message, endpoint=""):
        if env != "production":
            return

        message = json.dumps(message, indent=1)
        log.warning("HealthCheck ping for env %s payload %s", env, message)
        try:
            requests.post("https://hc-ping.com/" + HealthChecksUtil.ADWORDS_SYNC_PING_ID + endpoint,
                          data=message, timeout=10)
        except requests.RequestException as e:
            # Log ping failure here...
            log.error("Ping failed to healthchecks.io: %s" % e)
