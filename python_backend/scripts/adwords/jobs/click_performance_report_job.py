from .reports_fetch_job import ReportsFetch


# Note: If the number of custom paths exceed 7 in the subClasses. Move it to strategic pattern.
class ClickPerformanceReportsJob(ReportsFetch):
    QUERY_FIELDS = ["ad_format", "ad_group_id", "ad_group_name", "ad_group_status", "ad_network_type_1",
                    "ad_network_type_2",
                    "aoi_most_specific_target_id", "campaign_id", "campaign_location_target_id", "campaign_name",
                    "campaign_status", "clicks",
                    "click_type", "creative_id", "criteria_id", "criteria_parameters", "date", "device",
                    "external_customer_id", "gcl_id",
                    "page", "slot", "user_list_id"]
    REPORT = "CLICK_PERFORMANCE_REPORT"

    def __init__(self, next_info):
        super().__init__(next_info)

    # using transform to dedup.
    def transform_entities(self, rows):
        already_present = {}
        for row in rows:
            if row["gcl_id"] in already_present:
                continue
            already_present[row["gcl_id"]] = row
        return list(already_present.values())
