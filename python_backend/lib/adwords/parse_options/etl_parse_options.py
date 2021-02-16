from lib.parse_options import ParseOptions


# Check if this should be at WebServer/Application Level
class EtlParserOptions(ParseOptions):

    @classmethod
    def add_options_to_parser(cls, parser):
        parser.add_option("--env", dest="env", default="development")
        parser.add_option("--developer_token", dest="developer_token", help="", default="")
        parser.add_option("--dry", dest="dry", help="", default="False")
        parser.add_option("--extract_schema_changed", dest="extract_schema_changed", help="", default="False")
        parser.add_option("--skip_today", dest="skip_today", help="", default="False")
        parser.add_option("--oauth_secret", dest="oauth_secret", help="", default="")
        parser.add_option("--project_id", dest="project_id", help="", default=None, type=int)
        parser.add_option("--data_service_host", dest="data_service_host",
                          help="Data service host", default="http://localhost:8089")
