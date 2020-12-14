import json
import requests
from tornado.log import logging as log

class HealthchecksUtil:
    
    @staticmethod
    def ping_healthcheck(env, healthcheck_id, message, endpoint=""):
        if env != "production": 
            log.warning("Skipped healthcheck ping for env %s payload %s", env, str(message))
            return

        try:
            requests.post("https://hc-ping.com/" + healthcheck_id + endpoint,
                data=json.dumps(message, indent=1), timeout=10)
        except requests.RequestException as e:
            # Log ping failure here...
            log.error("Ping failed to healthchecks.io: %s" % e)
