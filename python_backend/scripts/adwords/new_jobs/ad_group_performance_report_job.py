from .fields_mapping import FieldsMapping
from .reports_fetch_job import ReportsFetch


class NewAdGroupPerformanceReportJob(ReportsFetch):
    
    # Elements in EXTRACT_FIELDS and HEADERS_VMAX are in sync
    # If field is added in any one then add to other too.
    EXTRACT_FIELDS = [
        "ad_group.name",
        "ad_group.status",
        "ad_group.type",
        "ad_group.base_ad_group",
        "campaign.base_campaign",
        "campaign.name",
        "campaign.status",
        "ad_group.final_url_suffix",
        "metrics.interaction_event_types",
        "ad_group.cpc_bid_micros",
        "ad_group.cpv_bid_micros",

        "campaign.id",
        "ad_group.id",
        
        "metrics.active_view_impressions",
        "metrics.active_view_measurability",
        "metrics.active_view_measurable_cost_micros",
        "metrics.active_view_measurable_impressions",
        "metrics.active_view_viewability",
        "metrics.cost_per_conversion",
        "metrics.all_conversions_value",
        "metrics.all_conversions",
        "metrics.average_cost",
        "metrics.clicks",
        "metrics.conversions_value",
        "metrics.conversions",
        "metrics.cost_micros",
        "metrics.engagements",
        "metrics.gmail_forwards",
        "metrics.gmail_saves",
        "metrics.gmail_secondary_clicks",
        "metrics.impressions",
        "metrics.interactions",
        "metrics.value_per_all_conversions",
        "metrics.video_quartile_p100_rate",
        "metrics.video_quartile_p25_rate",
        "metrics.video_quartile_p50_rate",
        "metrics.video_quartile_p75_rate",
        "metrics.video_view_rate",
        "metrics.video_views",
        "metrics.view_through_conversions",
        "metrics.search_impression_share",
        "metrics.search_top_impression_share",
        "metrics.search_absolute_top_impression_share",
        "metrics.search_budget_lost_absolute_top_impression_share",
        "metrics.search_budget_lost_top_impression_share",
        "metrics.search_rank_lost_absolute_top_impression_share",
        "metrics.search_rank_lost_impression_share",
        "metrics.search_rank_lost_top_impression_share",
    ]
    
    HEADERS_VMAX = [
        "ad_group_name",
        "ad_group_status",
        "ad_group_type",
        "base_ad_group_id",
        "base_campaign_id",
        "campaign_name",
        "campaign_status",
        "final_url_suffix",
        "interaction_types",
        "cpc_bid",
        "cpv_bid",

        "campaign_id",
        "ad_group_id",

        "active_view_impressions",
        "active_view_measurability",
        "active_view_measurable_cost",
        "active_view_measurable_impressions",
        "active_view_viewability",
        "cost_per_conversion",
        "all_conversion_value",
        "all_conversions",
        "average_cost",
        "clicks",
        "conversion_value",
        "conversions",
        "cost",
        "engagements",
        "gmail_forwards",
        "gmail_saves",
        "gmail_secondary_clicks",
        "impressions",
        "interactions",
        "value_per_all_conversion",
        "video_quartile_100_rate",
        "video_quartile_25_rate",
        "video_quartile_50_rate",
        "video_quartile_75_rate",
        "video_view_rate",
        "video_views",
        "view_through_conversions",
        "search_impression_share",
        "search_top_impression_share",
        "search_absolute_top_impression_share",
        "search_budget_lost_absolute_top_impression_share",
        "search_budget_lost_top_impression_share",
        "search_rank_lost_absolute_top_impression_share",
        "search_rank_lost_impression_share",
        "search_rank_lost_top_impression_share",
    ]
    
    HEADERS_V00 = [
        "average_position", "ad_group_name", "ad_group_status", "ad_group_type", "base_ad_group_id",
        "base_campaign_id", "campaign_name", "campaign_status", "final_url_suffix", "interaction_types",
        "cpc_bid", "cpv_bid",

        "campaign_id", "ad_group_id",

        "active_view_impressions", "active_view_measurability", "active_view_measurable_cost",
        "active_view_measurable_impressions", "active_view_viewability",
        "cost_per_conversion", "all_conversion_value", "all_conversions", "average_cost",
        "click_assisted_conversion_value", "click_assisted_conversions",
        "click_assisted_conversions_over_last_click_conversions", "clicks",
        "conversion_value", "conversions", "cost", "engagements", "gmail_forwards", "gmail_saves",
        "gmail_secondary_clicks", "impression_assisted_conversions", "impression_assisted_conversion_value",
        "impressions", "interactions", "value_per_all_conversion",
        "video_quartile_100_rate", "video_quartile_25_rate", "video_quartile_50_rate",
        "video_quartile_75_rate", "video_view_rate", "video_views", "view_through_conversions",
        "search_impression_share", "search_top_impression_share", "search_absolute_top_impression_share",
        "search_budget_lost_absolute_top_impression_share", "search_budget_lost_top_impression_share",
        "search_rank_lost_absolute_top_impression_share", "search_rank_lost_impression_share",
        "search_rank_lost_top_impression_share"]

    REPORT = "ad_group"
    
    FIELDS_WITH_STATUS = [
        "ad_group_status",
        "campaign_status",
    ]

    FIELDS_WITH_RESOURCE_NAME = [
        "base_ad_group_id",
        "base_campaign_id",
    ]

    FIELDS_TO_PERCENTAGE = [
        "active_view_measurability",
        "active_view_viewability",
        "video_quartile_100_rate",
        "video_quartile_25_rate",
        "video_quartile_50_rate",
        "video_quartile_75_rate",
        "video_view_rate",
    ]

    FIELDS_WITH_INTERACTION_TYPES = [
        "interaction_types",
    ]

    TRANSFORM_MAP_V01 = [
        {ReportsFetch.FIELD: "ad_group_type", ReportsFetch.MAP: FieldsMapping.AD_GROUP_TYPE_MAPPING},
    ]

    def __init__(self, next_info):
        super().__init__(next_info)
