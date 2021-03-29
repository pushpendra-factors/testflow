import itertools

from .reports_fetch_job import ReportsFetch


from .reports_fetch_job import ReportsFetch
# Note: If the number of custom paths exceed 7 in the subClasses. Move it to strategic pattern.
# To Add Segments "date", "device", "ad_network_type_1"
class CampaignPerformanceReportsJob(ReportsFetch):
    FIELDS = ["advertising_channel_sub_type", "average_position", "average_time_on_site", "base_campaign_id",
              # "budget_id",
              "campaign_name", "campaign_status",
              "campaign_trial_type", "start_date", "end_date", "interaction_types", "is_budget_explicitly_shared",
              "url_custom_parameters", "labels",
              "advertising_channel_type"]

    SEGMENTS = ["campaign_id"]

    METRICS = ["active_view_impressions", "active_view_measurability", "active_view_measurable_cost",
               "active_view_measurable_impressions", "active_view_viewability", "cost_per_conversion",
               "all_conversion_value", "all_conversions", "amount", "average_cost", "bounce_rate",
               "click_assisted_conversion_value",
               "click_assisted_conversions", "click_assisted_conversions_over_last_click_conversions", "clicks",
               "conversion_value", "conversions", "cost", "engagements", "gmail_forwards", "gmail_saves",
               "gmail_secondary_clicks", "impression_assisted_conversions", "impression_reach", "impressions",
               "interactions", "invalid_clicks", "value_per_all_conversion", "video_quartile_100_rate",
               "video_quartile_25_rate",
               "video_quartile_50_rate", "video_quartile_75_rate", "video_view_rate", "video_views",
               "view_through_conversions",
               "search_click_share",
               "search_impression_share", "search_top_impression_share", "search_absolute_top_impression_share",
               "search_budget_lost_absolute_top_impression_share", "search_budget_lost_impression_share",
               "search_budget_lost_top_impression_share",
               "search_rank_lost_absolute_top_impression_share", "search_rank_lost_impression_share",
               "search_rank_lost_top_impression_share"]

    QUERY_FIELDS = list(itertools.chain(FIELDS, SEGMENTS, METRICS))
    REPORT = "CAMPAIGN_PERFORMANCE_REPORT"

    def __init__(self, next_info):
        super().__init__(next_info)
