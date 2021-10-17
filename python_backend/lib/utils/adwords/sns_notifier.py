import requests
from tornado.log import logging as log

import scripts

# to do @ashhar: merge gsc and adwords functions
class SnsNotifier:
    APP_NAME = "adwords_sync"
    GSC_APP_NAME = "search_console_sync"

    @staticmethod
    def notify(message, name):
        env = scripts.adwords.CONFIG.ADWORDS_APP.env
        source = SnsNotifier.APP_NAME
        if env != "production":
            log.warning("Skipped notification for env %s payload %s", env, str(message))
            return

        sns_url = "https://fjnvg9a8wi.execute-api.us-east-1.amazonaws.com/v1/notify"
        payload = {"env": env, "source": source, "name": name, "message": message}
        response = requests.post(sns_url, json=payload)
        if not response.ok:
            log.error("Failed to notify through sns.")
        return response

    @staticmethod
    def notify_gsc(message, name):
        env = scripts.gsc.CONFIG.GSC_APP.env
        source = SnsNotifier.GSC_APP_NAME
        if env != "production":
            log.warning("Skipped notification for env %s payload %s", env, str(message))
            return

        sns_url = "https://fjnvg9a8wi.execute-api.us-east-1.amazonaws.com/v1/notify"
        payload = {"env": env, "message": message, "name": name, "source": source}
        response = requests.post(sns_url, json=payload)
        if not response.ok:
            log.error("Failed to notify through sns.")
        return response
