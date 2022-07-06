import itertools

from .fields_mapping import FieldsMapping
from .reports_fetch_job import ReportsFetch


# To Add Segments "date", "device", "ad_network_type_1"
# Note: If the number of custom paths exceed 7 in the subClasses. Move it to strategic pattern.


class NewKeywordPerformanceReportsJob(ReportsFetch):
    
    # Elements in EXTRACT_FIELDS and HEADERS_VMAX are in sync
    # If field is added in any one then add to other too.
    EXTRACT_FIELDS = [
        "metrics.cost_per_conversion",
        "metrics.all_conversions_value",
        "metrics.all_conversions",
        "metrics.cost_micros",
        "metrics.average_cost",
        "metrics.average_cpc",
        "metrics.average_cpm",
        "metrics.average_cpv",
        "metrics.average_page_views",
        "metrics.average_time_on_site",
        "metrics.clicks",
        "metrics.conversions",
        "metrics.ctr",
        "metrics.impressions",
        "metrics.search_impression_share",
        "metrics.search_top_impression_share",
        "metrics.search_absolute_top_impression_share",
        "metrics.search_budget_lost_absolute_top_impression_share",
        "metrics.search_budget_lost_top_impression_share",
        "metrics.search_rank_lost_absolute_top_impression_share",
        "metrics.search_rank_lost_impression_share",
        "metrics.search_rank_lost_top_impression_share",
        "metrics.absolute_top_impression_percentage",
        "metrics.top_impression_percentage",

        "ad_group_criterion.criterion_id",
        "ad_group.id",
        "campaign.id",

        "ad_group.name",
        "ad_group.status",
        "campaign.name",
        "campaign.status",
        "ad_group_criterion.labels",
        "ad_group_criterion.approval_status",
        "ad_group_criterion.keyword.text",
        "ad_group_criterion.keyword.match_type",
        "ad_group_criterion.negative",
        "ad_group_criterion.status",
        "ad_group_criterion.effective_cpc_bid_micros",
        "ad_group_criterion.effective_cpc_bid_source",
        "ad_group_criterion.position_estimates.first_position_cpc_micros",
        "ad_group_criterion.position_estimates.first_page_cpc_micros",
        "ad_group_criterion.position_estimates.top_of_page_cpc_micros",
        "ad_group_criterion.quality_info.quality_score",
    ]

    HEADERS_V02 = [
        "cost_per_conversion",
        "all_conversion_value",
        "all_conversions",
        "cost",
        "average_cost",
        "average_cpc",
        "average_cpm",
        "average_cpv",
        "average_pageviews",
        "average_time_on_site",
        "clicks",
        "conversions",
        "ctr",
        "impressions",
        "search_impression_share",
        "search_top_impression_share",
        "search_absolute_top_impression_share",
        "search_budget_lost_absolute_top_impression_share",
        "search_budget_lost_top_impression_share",
        "search_rank_lost_absolute_top_impression_share",
        "search_rank_lost_impression_share",
        "search_rank_lost_top_impression_share",
        "absolute_top_impression_percentage",
        "top_impression_percentage",

        "id",
        "ad_group_id",
        "campaign_id",

        "ad_group_name",
        "ad_group_status",
        "campaign_name",
        "campaign_status",
        "labels",
        "approval_status",
        "criteria",
        "keyword_match_type",
        "is_negative",
        "status",
        "cpc_bid",
        "cpc_bid_source",
        "first_position_cpc",
        "first_page_cpc",
        "top_of_page_cpc",
        "quality_score",
    ]

    HEADERS_V01 = [
        "cost_per_conversion",
        "all_conversion_value",
        "all_conversions",
        "cost",
        "average_cost",
        "average_cpc",
        "average_cpm",
        "average_cpv",
        "average_pageviews",
        "average_time_on_site",
        "clicks",
        "conversions",
        "ctr",
        "impressions",
        "search_impression_share",
        "search_top_impression_share",
        "search_absolute_top_impression_share",
        "search_budget_lost_absolute_top_impression_share",
        "search_budget_lost_top_impression_share",
        "search_rank_lost_absolute_top_impression_share",
        "search_rank_lost_impression_share",
        "search_rank_lost_top_impression_share",

        "id",
        "ad_group_id",
        "campaign_id",

        "ad_group_name",
        "ad_group_status",
        "campaign_name",
        "campaign_status",
        "labels",
        "approval_status",
        "criteria",
        "keyword_match_type",
        "is_negative",
        "status",
        "cpc_bid",
        "cpc_bid_source",
        "first_position_cpc",
        "first_page_cpc",
        "top_of_page_cpc",
        "quality_score",
    ]
    
    HEADERS_V00 = [
        "cost_per_conversion", "all_conversion_value", "all_conversions", "cost", "average_cost", "average_cpc",
        "average_cpm", "average_cpv", "average_pageviews", "average_time_on_site",
        "click_assisted_conversion_value",
        "click_assisted_conversions", "clicks", "conversions", "ctr", "impression_assisted_conversions",
        "impressions",
        "search_impression_share", "search_top_impression_share", "search_absolute_top_impression_share",
        "search_budget_lost_absolute_top_impression_share", "search_budget_lost_top_impression_share",
        "search_rank_lost_absolute_top_impression_share", "search_rank_lost_impression_share",
        "search_rank_lost_top_impression_share",

        "id", "ad_group_id", "campaign_id",

        "ad_group_name", "ad_group_status", "campaign_name", "campaign_status", "labels",
        "approval_status", "average_position", "criteria", "keyword_match_type", "is_negative", "status",
        "cpc_bid", "cpc_bid_source", "first_position_cpc", "first_page_cpc", "top_of_page_cpc", "quality_score"]

    HEADERS_VMAX = HEADERS_V02

    REPORT = "keyword_view"

    FIELDS_WITH_STATUS = [
        "ad_group_status",
        "campaign_status",
        "approval_status",
        "status",
    ]

    FIELDS_WITH_BOOLEAN = [
        "is_negative",
    ]
    
    TRANSFORM_MAP_V01 = [
        {ReportsFetch.FIELD: "keyword_match_type", ReportsFetch.MAP: FieldsMapping.KEYWORD_MAPPING},
        {ReportsFetch.FIELD: "cpc_bid_source", ReportsFetch.MAP: FieldsMapping.BIDDING_SOURCE_MAPPING},
    ]

    def __init__(self, next_info):
        super().__init__(next_info)
