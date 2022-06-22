from google.cloud import storage


class GoogleStorageFile:
    client = None
    bucket = None

    def __init__(self, bucket_name):
        self.client = storage.Client()
        self.bucket = self.client.get_bucket(bucket_name)

    def read(self, file_path):
        blob = self.bucket.blob(file_path)
        if blob is None:
            return None
        return blob.download_as_text()

    # Overriding the previously present file.
    def write(self, input_string, file_path):
        blob = self.bucket.blob(file_path)
        blob.upload_from_string(input_string)
        return
