import json

import requests
import logging as log

from lib.utils.json import JsonUtil

class SlackUtil:


    @staticmethod
    def ping(env, message, slack_url):
        if env != "production":
            return
        
        count = 0
        response = {}
        # retrying
        while count<= 3:
            count += 1
            response = requests.post(slack_url, json=message, timeout=10)
            if response.ok:
                break
        if not response.ok:
            log.error('Ping failed to slack alerts')

    @staticmethod
    def build_slack_block(project_ids, channel_name):
        message = {}
        blocks = [{
			"type": "header",
			"text": {
				"type": "plain_text",
				"text": channel_name + " token failures"
			}
		}]
        project_ids_str = ", ".join(project_ids)

        fields = [{
                "type": "plain_text",
                "text": project_ids_str
            }]
        section = {
            "type" : "section",
            "fields": fields
        }
        blocks.append(section)
        message["blocks"] = blocks

        return message
