from optparse import OptionParser

from lib.adwords.parse_options.etl_parse_options import AdwordsEtlParserOptions as AdwordsParseOptions


# TODO: we are not able to merge the parsed values. Hence moving the parsing functionality outside. This goes bad. Hence remodify.
class EtlParser:

    def __init__(self, argv):
        self.argv = argv
        self.parser = OptionParser()

    def parse(self):
        self.add_options_for_parsing_adwords()
        return self.parser.parse_args(self.argv)

    def add_options_for_parsing_adwords(self):
        AdwordsParseOptions.add_options_to_parser(self.parser)
