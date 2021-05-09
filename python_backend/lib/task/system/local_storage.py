import errno
import os

from .base import BaseSystem


class LocalStorage(BaseSystem):

    def read(self):
        base_path = self.system_attributes["base_path"]
        file_path = self.system_attributes["file_path"]
        absolute_file_path = base_path + file_path
        with open(absolute_file_path, 'r') as reader:
            result = reader.read()
        return result

    def write(self, input_string):
        base_path = self.system_attributes["base_path"]
        file_path = self.system_attributes["file_path"]
        absolute_file_path = base_path + file_path
        self.create_dirs(absolute_file_path)
        with open(absolute_file_path, 'w+') as writer:
            writer.write(input_string)
        return

    @staticmethod
    def create_dirs(file_path):
        if not os.path.exists(os.path.dirname(file_path)):
            try:
                os.makedirs(os.path.dirname(file_path))
            except OSError as exc:  # Guard against race condition
                if exc.errno != errno.EEXIST:
                    raise
