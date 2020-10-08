# TODO: Handle errors.
# TODO: this can be called as ParserOptions
# TODO: Check where is OptParser module coming from. https://www.tornadoweb.org/en/stable/options.html
class ParseOptions:
    
    @classmethod
    def add_options_to_parser(cls, parser):
        """ Override this method to provide the options required to parse input command line args """
