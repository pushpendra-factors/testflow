class EtlRecordsStats:
    PROJECT_KEY = 'projects'
    TOTAL_KEY = 'total_by_key'

    def __init__(self):
        self.request_stats = {self.PROJECT_KEY: {}, self.TOTAL_KEY: {}}

    def update(self, project_id, doc_type, count):
        if self.request_stats[self.PROJECT_KEY].get(project_id) is None:
            self.request_stats[self.PROJECT_KEY][project_id] = {}

        if self.request_stats[self.PROJECT_KEY][project_id].get(doc_type) is None:
            self.request_stats[self.PROJECT_KEY][project_id][doc_type] = 0

        if self.request_stats[self.TOTAL_KEY].get(doc_type) is None:
            self.request_stats[self.TOTAL_KEY][doc_type] = 0

        self.request_stats[self.PROJECT_KEY][project_id][doc_type] += count
        self.request_stats[self.TOTAL_KEY][doc_type] += count
