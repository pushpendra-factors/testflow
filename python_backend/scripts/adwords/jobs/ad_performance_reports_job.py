from .reports_fetch_job import ReportsFetch


class AdPerformanceReportsJob(ReportsFetch):
    QUERY_FIELDS = ["id", "account_currency_code", "account_descriptive_name", "active_view_impressions",
                    "active_view_measurability", "active_view_measurable_cost", "active_view_measurable_impressions",
                    "active_view_viewability", "ad_group_id", "campaign_id",
                    "cost_per_conversion",
                    "all_conversion_value", "all_conversions", "average_cost", "average_position",
                    "average_time_on_site",
                    "bounce_rate", "click_assisted_conversion_value", "click_assisted_conversions",
                    "click_assisted_conversions_over_last_click_conversions",
                    "clicks", "conversion_value", "conversions", "cost", "date", "engagements", "gmail_forwards",
                    "gmail_saves", "gmail_secondary_clicks", "impression_assisted_conversions", "impressions",
                    "interaction_types",
                    "interactions", "value_per_all_conversion", "video_quartile_100_rate", "video_quartile_25_rate",
                    "video_quartile_50_rate", "video_quartile_75_rate", "video_view_rate", "video_views",
                    "view_through_conversions"]
    REPORT = "AD_PERFORMANCE_REPORT"

    def __init__(self, next_info):
        super().__init__(next_info)
