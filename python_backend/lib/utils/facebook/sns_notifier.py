import logging as log

import requests


class SnsNotifier:
    SNS_URL = "https://fjnvg9a8wi.execute-api.us-east-1.amazonaws.com/v1/notify"

    def __init__(self, env, source):
        self.env = env
        self.source = source

    def notify(self, message, name):
        if self.env != "production":
            log.warning("Skipped notification for env %s payload %s", self.env, str(message))
            return

        payload = {"env": self.env, "source": self.source, "name": name, "message": message}
        response = requests.post(self.SNS_URL, json=payload)
        if not response.ok:
            log.error("Failed to notify through sns.")
        return response
