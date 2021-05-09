# Worker Implementation
class BaseLoad:
    INSTANCE = None

    @classmethod
    def get_instance(cls):
        """ Override this to fetch the instance of required sub class Object. """
        pass

    def execute(self, task_context):
        """ Override this to execute the required sub class Object. """
