from .fields_mapping import FieldsMapping
from .reports_fetch_job import ReportsFetch
# Note: If the number of custom paths exceed 7 in the subClasses. Move it to strategic pattern.
# To Add Segments "date", "device", "ad_network_type_1"
class NewCampaignPerformanceReportsJob(ReportsFetch):
    
    # Elements in EXTRACT_FIELDS and HEADERS_VMAX are in sync
    # If field is added in any one then add to other too.
    EXTRACT_FIELDS = [
        "campaign.advertising_channel_sub_type",
        "metrics.average_time_on_site",
        "campaign.base_campaign",
        "campaign.name",
        "campaign.status",
        "campaign.experiment_type",
        "campaign.start_date",
        "campaign.end_date",
        "metrics.interaction_event_types",
        "campaign_budget.explicitly_shared",
        "campaign.url_custom_parameters",
        "campaign.labels",
        "campaign.advertising_channel_type",

        "campaign.id",

        "metrics.active_view_impressions",
        "metrics.active_view_measurability",
        "metrics.active_view_measurable_cost_micros",
        "metrics.active_view_measurable_impressions",
        "metrics.active_view_viewability",
        "metrics.cost_per_conversion",
        "metrics.all_conversions_value",
        "metrics.all_conversions",
        "campaign_budget.amount_micros",
        "metrics.average_cost",
        "metrics.bounce_rate",
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
        "metrics.invalid_clicks",
        "metrics.value_per_all_conversions",
        "metrics.video_quartile_p100_rate",
        "metrics.video_quartile_p25_rate",
        "metrics.video_quartile_p50_rate",
        "metrics.video_quartile_p75_rate",
        "metrics.video_view_rate",
        "metrics.video_views",
        "metrics.view_through_conversions",
        "metrics.search_click_share",
        "metrics.search_impression_share",
        "metrics.search_top_impression_share",
        "metrics.search_absolute_top_impression_share",
        "metrics.search_budget_lost_absolute_top_impression_share",
        "metrics.search_budget_lost_impression_share",
        "metrics.search_budget_lost_top_impression_share",
        "metrics.search_rank_lost_absolute_top_impression_share",
        "metrics.search_rank_lost_impression_share",
        "metrics.search_rank_lost_top_impression_share",
        "metrics.absolute_top_impression_percentage",
        "metrics.top_impression_percentage",
    ]

    HEADERS_V02 = [
        "advertising_channel_sub_type",
        "average_time_on_site",
        "base_campaign_id",
        "campaign_name",
        "campaign_status",
        "campaign_trial_type",
        "start_date",
        "end_date",
        "interaction_types",
        "is_budget_explicitly_shared",
        "url_custom_parameters",
        "labels",
        "advertising_channel_type",

        "campaign_id",
        
        "active_view_impressions",
        "active_view_measurability",
        "active_view_measurable_cost",
        "active_view_measurable_impressions",
        "active_view_viewability",
        "cost_per_conversion",
        "all_conversion_value",
        "all_conversions",
        "amount",
        "average_cost",
        "bounce_rate",
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
        "invalid_clicks",
        "value_per_all_conversion",
        "video_quartile_100_rate",
        "video_quartile_25_rate",
        "video_quartile_50_rate",
        "video_quartile_75_rate",
        "video_view_rate",
        "video_views",
        "view_through_conversions",
        "search_click_share",
        "search_impression_share",
        "search_top_impression_share",
        "search_absolute_top_impression_share",
        "search_budget_lost_absolute_top_impression_share",
        "search_budget_lost_impression_share",
        "search_budget_lost_top_impression_share",
        "search_rank_lost_absolute_top_impression_share",
        "search_rank_lost_impression_share",
        "search_rank_lost_top_impression_share",
        "absolute_top_impression_percentage",
        "top_impression_percentage",
    ]

    HEADERS_V01 = [
        "advertising_channel_sub_type",
        "average_time_on_site",
        "base_campaign_id",
        "campaign_name",
        "campaign_status",
        "campaign_trial_type",
        "start_date",
        "end_date",
        "interaction_types",
        "is_budget_explicitly_shared",
        "url_custom_parameters",
        "labels",
        "advertising_channel_type",

        "campaign_id",
        
        "active_view_impressions",
        "active_view_measurability",
        "active_view_measurable_cost",
        "active_view_measurable_impressions",
        "active_view_viewability",
        "cost_per_conversion",
        "all_conversion_value",
        "all_conversions",
        "amount",
        "average_cost",
        "bounce_rate",
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
        "invalid_clicks",
        "value_per_all_conversion",
        "video_quartile_100_rate",
        "video_quartile_25_rate",
        "video_quartile_50_rate",
        "video_quartile_75_rate",
        "video_view_rate",
        "video_views",
        "view_through_conversions",
        "search_click_share",
        "search_impression_share",
        "search_top_impression_share",
        "search_absolute_top_impression_share",
        "search_budget_lost_absolute_top_impression_share",
        "search_budget_lost_impression_share",
        "search_budget_lost_top_impression_share",
        "search_rank_lost_absolute_top_impression_share",
        "search_rank_lost_impression_share",
        "search_rank_lost_top_impression_share",
    ]

    HEADERS_V00 = [
        "advertising_channel_sub_type", "average_position", "average_time_on_site", "base_campaign_id",
        # "budget_id",
        "campaign_name", "campaign_status",
        "campaign_trial_type", "start_date", "end_date", "interaction_types", "is_budget_explicitly_shared",
        "url_custom_parameters", "labels",
        "advertising_channel_type",

        "campaign_id",

        "active_view_impressions", "active_view_measurability", "active_view_measurable_cost",
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
    
    HEADERS_VMAX = HEADERS_V02
    
    REPORT = "campaign"

    FIELDS_WITH_BOOLEAN = [
        "is_budget_explicitly_shared",
    ]

    FIELDS_WITH_STATUS = [
        "campaign_status",
    ]

    FIELDS_WITH_RESOURCE_NAME = [
        "base_campaign_id",
    ]

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

    TRANSFORM_MAP_V01 = [
        {ReportsFetch.FIELD: "campaign_trial_type", ReportsFetch.MAP: FieldsMapping.CAMPAIGN_TRIAL_TYPE_MAPPING},
        {ReportsFetch.FIELD: "advertising_channel_type", ReportsFetch.MAP: FieldsMapping.ADVERTISING_CHANNEL_TYPE_MAPPING},
        {ReportsFetch.FIELD: "advertising_channel_sub_type", ReportsFetch.MAP: FieldsMapping.ADVERTISING_CHANNEL_SUB_TYPE_MAPPING},
    ]

    def __init__(self, next_info):
        super().__init__(next_info)
