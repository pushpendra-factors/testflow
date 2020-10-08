from lib.parse_options import ParseOptions


class SentryParseOptions(ParseOptions):

    @classmethod
    def add_options_to_parser(cls, parser):
        parser.add_option("--sentry_dsn", default="", help="Sentry Client DSN.")
        parser.add_option("--server_name", default="python backend", help="Runs on the given server.")
