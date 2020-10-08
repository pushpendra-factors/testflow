
class Config:
    # _argv = None

    # @class_method
    # def init(cls, argv):
    #     cls._argv = argv

    @classmethod
    def build(cls, argv):
        """ Override this method to build the Configuration Objects """
