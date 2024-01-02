class CustomException(Exception):
    message = ''
    request_count = 0
    doc_type = ''
    def __init__(self, message, request_count, doc_type):
        self.message = message
        self.request_count = request_count
        self.doc_type = doc_type