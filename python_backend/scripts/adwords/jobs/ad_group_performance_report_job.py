import itertools

from .reports_fetch_job import ReportsFetch


class AdGroupPerformanceReportJob(ReportsFetch):
    FIELDS = ['average_position', 'ad_group_name', 'ad_group_status', 'ad_group_type', 'base_ad_group_id',
              'base_campaign_id', 'campaign_name', 'campaign_status', 'final_url_suffix', 'interaction_types',
              'cpc_bid', 'cpv_bid']

    SEGMENTS = ['campaign_id', 'ad_group_id']

    METRICS = ['active_view_impressions', 'active_view_measurability', 'active_view_measurable_cost',
               'active_view_measurable_impressions', 'active_view_viewability',
               'cost_per_conversion', 'all_conversion_value', 'all_conversions', 'average_cost',
               'click_assisted_conversion_value', 'click_assisted_conversions',
               'click_assisted_conversions_over_last_click_conversions', 'clicks',
               'conversion_value', 'conversions', 'cost', 'engagements', 'gmail_forwards', 'gmail_saves',
               'gmail_secondary_clicks', 'impression_assisted_conversions', 'impression_assisted_conversion_value',
               'impressions', 'interactions', 'value_per_all_conversion',
               'video_quartile_100_rate', 'video_quartile_25_rate', 'video_quartile_50_rate',
               'video_quartile_75_rate', 'video_view_rate', 'video_views', 'view_through_conversions',
               'search_impression_share', 'search_top_impression_share', 'search_absolute_top_impression_share',
               'search_budget_lost_absolute_top_impression_share', 'search_budget_lost_top_impression_share',
               'search_rank_lost_absolute_top_impression_share', 'search_rank_lost_impression_share',
               'search_rank_lost_top_impression_share']

    QUERY_FIELDS = list(itertools.chain(FIELDS, SEGMENTS, METRICS))
    REPORT = 'ADGROUP_PERFORMANCE_REPORT'

    def __init__(self, next_info):
        super().__init__(next_info)
