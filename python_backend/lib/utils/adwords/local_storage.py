# noinspection PyMethodMayBeStatic
import errno
import os

# Though this is not in used in google context, this is used only for google.
class LocalStorage:
    BASE_FOLDER = "/usr/local/var/factors/cloud_storage/"

    def read(self, file_path):
        file_path = LocalStorage.BASE_FOLDER + file_path
        with open(file_path, 'r') as reader:
            result = reader.read()
        return result

    def write(self, input_string, file_name_with_path):
        file_name_with_path = LocalStorage.BASE_FOLDER + file_name_with_path
        LocalStorage.create_dirs(file_name_with_path)
        with open(file_name_with_path, 'w+') as writer:
            writer.write(input_string)

    @staticmethod
    def create_dirs(file_name_with_path):
        if not os.path.exists(os.path.dirname(file_name_with_path)):
            try:
                os.makedirs(os.path.dirname(file_name_with_path))
            except OSError as exc:  # Guard against race condition
                if exc.errno != errno.EEXIST:
                    raise
