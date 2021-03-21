import json
import logging as log

from lib.sns_notifier import SnsNotifier
from lib.utils.json import JsonUtil
from scripts.adwords import REQUEST_COUNT, RECORDS_COUNT, LATENCY_COUNT, EXTRACT_PHASE, \
    LOAD_PHASE


class JobTaskStats:
    PROJECT_KEY = "projects"
    TOTAL_KEY = "total_by_key"
    STATS = {
            REQUEST_COUNT: {PROJECT_KEY: {}, TOTAL_KEY: {}},
            RECORDS_COUNT: {PROJECT_KEY: {}, TOTAL_KEY: {}},
            LATENCY_COUNT: {PROJECT_KEY: {}, TOTAL_KEY: {}},
    }
    task_stats = None

    def __init__(self):
        self.task_stats = {
            EXTRACT_PHASE: self.STATS.copy(),
            LOAD_PHASE: self.STATS.copy()
        }

    # Each type of run has reading from source and pushing to destination. Phase represents this.
    def update_record_stats(self, phase, metric_type, project_id, doc_type, value):
        count_map = self.task_stats[phase][metric_type]
        project_count_map = count_map[self.PROJECT_KEY]
        total_key_map = count_map[self.TOTAL_KEY]

        project_count_map.setdefault(project_id, {})
        per_project_count_map = project_count_map[project_id]
        per_project_count_map.setdefault(doc_type, 0)
        total_key_map.setdefault(doc_type, 0)

        per_project_count_map[doc_type] += value
        total_key_map[doc_type] += value

    def processed_equal_records(self, that):
        if isinstance(that, JobTaskStats):
            return json.dumps(self.STATS[RECORDS_COUNT]) == \
                   json.dumps(that.STATS[RECORDS_COUNT])
        return False

    def publish(self):
        task_stats = json.dumps(self.task_stats)
        SnsNotifier.notify(task_stats)
        log.warning("Metrics for the job Tasks: %s", task_stats)
