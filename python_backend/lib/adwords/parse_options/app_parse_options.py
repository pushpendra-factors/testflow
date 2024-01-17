from lib.parse_options import ParseOptions


# Check if this should be at WebServer/Application Level
class AppParserOptions(ParseOptions):

    @classmethod
    def add_options_to_parser(cls, parser):
        parser.add_option("--port", default=8091, help="Runs on the given port.")
        parser.add_option("--host_url", default="http://localhost:8091",
                          help="Self host url with protocol to refer on callbacks.")
        parser.add_option("--env", default="development", help="Environment.")
        parser.add_option("--developer_token", default="", help="Adwords developer token.")
        parser.add_option("--api_host_url", default="http://localhost:8089", help="Data service host url")
        parser.add_option("--app_host_url", default="http://factors-dev.com:3000", help="App host url")
        parser.add_option("--oauth_secret", default="", help="OAuth2 client secret JSON string")