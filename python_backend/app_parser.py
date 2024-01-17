from optparse import OptionParser

from lib.adwords.parse_options.app_parse_options import AppParserOptions as AdwordsParseOptions
# TODO: we are not able to merge the parsed values. Hence moving the parsing functionality outside. This goes bad. Hence remodify.
from lib.data_services.sentry.sentry_parse_options import SentryParseOptions
from chat_factors.chat.chat_parse_options import ChatParserOptions


class AppParser:

    def __init__(self, argv):
        self.argv = argv
        self.parser = OptionParser()

    def parse(self):
        self.add_options_for_parsing_google()
        self.add_options_for_parsing_sentry()
        self.add_options_for_parsing_chat()
        return self.parser.parse_args(self.argv)

    def add_options_for_parsing_google(self):
        AdwordsParseOptions.add_options_to_parser(self.parser)

    def add_options_for_parsing_sentry(self):
        SentryParseOptions.add_options_to_parser(self.parser)

    def add_options_for_parsing_chat(self):
        ChatParserOptions.add_options_to_parser(self.parser)
