from lib.parse_options import ParseOptions


class ChatParserOptions(ParseOptions):

    @classmethod
    def add_options_to_parser(cls, parser):
        parser.add_option("--chat_bucket_name", default="", help="bucket name for chat data")
