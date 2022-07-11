from .fields_mapping import FieldsMapping
from .reports_fetch_job import ReportsFetch


class NewAdPerformanceReportsJob(ReportsFetch):

    # Elements in EXTRACT_FIELDS and HEADERS_VMAX are in sync
    # If field is added in any one then add to other too.
    EXTRACT_FIELDS = [
        "ad_group_ad.ad.id",
        "customer.currency_code",
        "customer.descriptive_name",
        "metrics.active_view_impressions",
        "metrics.active_view_measurability",
        "metrics.active_view_measurable_cost_micros",
        "metrics.active_view_measurable_impressions",
        "metrics.active_view_viewability",
        "ad_group.id",
        "campaign.id",
        "metrics.cost_per_conversion",
        "metrics.all_conversions_value",
        "metrics.all_conversions",
        "metrics.average_cost",
        "metrics.average_time_on_site",
        "metrics.bounce_rate",
        "metrics.clicks",
        "metrics.conversions_value",
        "metrics.conversions",
        "metrics.cost_micros",
        "segments.date",
        "metrics.engagements",
        "metrics.gmail_forwards",
        "metrics.gmail_saves",
        "metrics.gmail_secondary_clicks",
        "metrics.impressions",
        "metrics.interaction_event_types",
        "metrics.interactions",
        "metrics.value_per_all_conversions",
        "metrics.video_quartile_p100_rate",
        "metrics.video_quartile_p25_rate",
        "metrics.video_quartile_p50_rate",
        "metrics.video_quartile_p75_rate",
        "metrics.video_view_rate",
        "metrics.video_views",
        "metrics.view_through_conversions",
    ]

    HEADERS_V01 = [
        "id",
        "account_currency_code",
        "account_descriptive_name",
        "active_view_impressions",
        "active_view_measurability",
        "active_view_measurable_cost",
        "active_view_measurable_impressions",
        "active_view_viewability",
        "ad_group_id",
        "campaign_id",
        "cost_per_conversion",
        "all_conversion_value",
        "all_conversions",
        "average_cost",
        "average_time_on_site",
        "bounce_rate",
        "clicks",
        "conversion_value",
        "conversions",
        "cost",
        "date",
        "engagements",
        "gmail_forwards",
        "gmail_saves",
        "gmail_secondary_clicks",
        "impressions",
        "interaction_types",
        "interactions",
        "value_per_all_conversion",
        "video_quartile_100_rate",
        "video_quartile_25_rate",
        "video_quartile_50_rate",
        "video_quartile_75_rate",
        "video_view_rate",
        "video_views",
        "view_through_conversions",
    ]

    HEADERS_V00 = [
        "id", "account_currency_code", "account_descriptive_name", "active_view_impressions",
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

    HEADERS_V02 = HEADERS_V01
    HEADERS_VMAX = HEADERS_V01

    REPORT = "ad_group_ad"

    FIELDS_TO_PERCENTAGE = [
        "active_view_measurability",
        "active_view_viewability",
        "bounce_rate",
        "video_quartile_100_rate",
        "video_quartile_25_rate",
        "video_quartile_50_rate",
        "video_quartile_75_rate",
        "video_view_rate",
    ]

    FIELDS_WITH_INTERACTION_TYPES = [
        "interaction_types",
    ]

    def __init__(self, next_info):
        super().__init__(next_info)
