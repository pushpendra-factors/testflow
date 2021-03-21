import json


# Standardise the method names.
class JsonUtil:

    @staticmethod
    def read(json_string):
        if json_string is None or len(json_string) < 2:
            return {}
        return json.loads(json_string)

    @staticmethod
    def create(json_data):
        return json.dumps(json_data)

    @staticmethod
    def serialize_sets(obj):
        if isinstance(obj, set):
            return list(obj)
        return obj
