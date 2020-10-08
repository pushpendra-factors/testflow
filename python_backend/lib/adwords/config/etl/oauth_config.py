
# TODO: This class is similar to adwords/app/oauth_config.
# It errors when argv.host_url is called but not initialised as parameter.
import json
import sys

from tornado.log import logging as log

from lib.config import Config


class OauthConfig(Config):
    secret_json = None
    client_id = None
    client_secret = None
    developer_token = None

    @classmethod
    def build(cls, argv):
        cls._validate(argv)
        oauth_secret_json = json.loads(argv.oauth_secret.strip())
        cls._init(oauth_secret_json, argv.developer_token)
    
    @staticmethod
    def _validate(argv):
        if argv.developer_token == "":
            log.error("Option: developer_token cannot be empty")
            sys.exit(1)
        
        oauth_secret_str = argv.oauth_secret.strip()
        if oauth_secret_str == "":
            log.error("Option: oauth_secret cannot be empty.")
            sys.exit(1)

        try:
            json.loads(oauth_secret_str)
        except Exception as e:
            log.error("Failed to load oauth_secret JSON: %s", str(e))
            sys.exit(1)

    @classmethod
    def _init(cls, oauth_secret_json, developer_token):
        cls.secret_json = oauth_secret_json
        cls.client_id = oauth_secret_json["web"]["client_id"]
        cls.client_secret = oauth_secret_json["web"]["client_secret"]
        cls.developer_token = developer_token
        return
