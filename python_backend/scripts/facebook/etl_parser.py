from optparse import OptionParser

from lib.facebook.parse_options.etl_parse_options import FacebookEtlParserOptions


class EtlParser:

    def __init__(self, argv):
        self.argv = argv
        self.parser = OptionParser()

    def parse(self):
        self.add_options_for_parsing_facebook()
        return self.parser.parse_args(self.argv)

    def add_options_for_parsing_facebook(self):
        FacebookEtlParserOptions.add_options_to_parser(self.parser)
