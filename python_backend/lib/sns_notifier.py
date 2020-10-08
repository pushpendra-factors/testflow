import requests
from tornado.log import logging as log

class SnsNotifier:

    @staticmethod
    def notify(env, source, message):
        if env != "production":
            log.warning("Skipped notification for env %s payload %s", env, str(message))
            return

        sns_url = "https://fjnvg9a8wi.execute-api.us-east-1.amazonaws.com/v1/notify"
        payload = {"env": env, "message": message, "source": source}
        response = requests.post(sns_url, json=payload)
        if not response.ok:
            log.error("Failed to notify through sns.")
        return response
