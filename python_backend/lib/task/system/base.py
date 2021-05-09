
# BaseSystem - read doesnt know what kind of response will be given and might not even information if its csv or json.
# Read is not having consistent response across different sub classes.
class BaseSystem:
    system_attributes = None

    def set_attributes(self, attributes):
        self.system_attributes = attributes

    def read(self):
        pass
        """ Implement the below for reading from any system """

    def write(self, input_string):
        pass
        """ Implement the below for reading from any system """
