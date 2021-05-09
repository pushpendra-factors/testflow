# Singleton Equivalence. Is it not a performance issue when used in multi thread?
# TODO adding job execution to task Context or not? revisit base_info_extract
class BaseExtract:
    INSTANCE = None

    @classmethod
    def get_instance(cls):
        """ Override this to fetch the instance of required sub class Object. """
        pass

    def execute(self, task_context):
        """ Override this to execute the Sub task. """
        pass
