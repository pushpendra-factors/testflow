from lib.parse_options import ParseOptions


# Check if this should be at WebServer/Application Level
class AdwordsEtlParserOptions(ParseOptions):

    @classmethod
    def add_options_to_parser(cls, parser):
        parser.add_option("--env", dest="env", default="development")
        parser.add_option("--developer_token", dest="developer_token", help="", default="")
        parser.add_option("--oauth_secret", dest="oauth_secret", help="", default="")
        parser.add_option("--skip_today", dest="skip_today", help="", default="False")
        parser.add_option("--data_service_host", dest="data_service_host",
                          help="Data service host", default="http://localhost:8089")

        parser.add_option("--project_id", dest="project_id", help="", default=None)
        parser.add_option("--exclude_project_id", dest="exclude_project_id", help="", default=None)
        parser.add_option("--document_type", dest="document_type", help="", default=None)
        parser.add_option("--timezone", dest="timezone", default="")
        parser.add_option("--type_of_run", dest="type_of_run", default="extract_and_load")
        parser.add_option("--dry", dest="dry", help="", default="False")
        parser.add_option("--last_timestamp", dest="last_timestamp", help="", default=None, type=int)
        parser.add_option("--to_timestamp", dest="to_timestamp", help="", default=None, type=int)
        parser.add_option("--new_extract_project_id", dest="new_extract_project_id", help="", default=None)
