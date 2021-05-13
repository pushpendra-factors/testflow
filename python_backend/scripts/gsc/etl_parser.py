from optparse import OptionParser

from lib.gsc.parse_options.etl_parse_options import EtlParserOptions


# TODO: we are not able to merge the parsed values. Hence moving the parsing functionality outside. This goes bad. Hence remodify.
class EtlParser:

    def __init__(self, argv):
        self.argv = argv
        self.parser = OptionParser()

    def parse(self):
        self.add_options_for_parsing_gsc()
        return self.parser.parse_args(self.argv)

    def add_options_for_parsing_gsc(self):
        EtlParserOptions.add_options_to_parser(self.parser)
