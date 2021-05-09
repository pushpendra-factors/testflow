import json

import requests
import logging as log

from lib.utils.json import JsonUtil

class HealthChecksUtil:


    @staticmethod
    def ping(env, message, ping_id, endpoint=""):
        if env != "production":
            return

        message = json.dumps(message, indent=1, default=JsonUtil.serialize_sets)
        log.warning("HealthCheck ping for env %s payload %s", env, message)
        try:
            requests.post("https://hc-ping.com/" + ping_id + endpoint,
                          data=message, timeout=10)
        except requests.RequestException as e:
            # Log ping failure here...
            log.error("Ping failed to healthchecks.io: %s" % e)
