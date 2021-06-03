from .reports_fetch_job import ReportsFetch


# Note: If the number of custom paths exceed 7 in the subClasses. Move it to strategic pattern.
class GetSearchConsolePageDataJob(ReportsFetch):
    DIMENSIONS = ["page"]

    def __init__(self, next_info):
        super().__init__(next_info)
