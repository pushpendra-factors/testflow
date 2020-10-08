from .reports_fetch_job import ReportsFetch


# Note: If the number of custom paths exceed 7 in the subClasses. Move it to strategic pattern.
class SeachPerformanceReportsJob(ReportsFetch):
    QUERY_FIELDS = ['ad_group_id', 'ad_group_name', 'all_conversion_rate', 'conversion_rate', 'cost_per_conversion',
                    'all_conversion_value', 'all_conversions', 'average_cost', 'average_cpc', 'average_cpe',
                    'average_cpm', 'average_cpv', 'average_position', 'campaign_id', 'clicks', 'conversion_value',
                    'conversions',
                    'cost', 'cost_per_all_conversion', 'cross_device_conversions', 'ctr', 'date',
                    'device', 'engagement_rate', 'engagements', 'external_customer_id',
                    'final_url', 'impressions', 'interaction_rate', 'interaction_types', 'interactions', 'keyword_id',
                    'query', 'query_match_type_with_variant', 'tracking_url_template', 'value_per_all_conversion',
                    'value_per_conversion', 'video_quartile_100_rate', 'video_quartile_25_rate',
                    'video_quartile_50_rate',
                    'video_quartile_75_rate', 'video_view_rate', 'video_views', 'view_through_conversions', 'week',
                    'year']
    REPORT = 'SEARCH_QUERY_PERFORMANCE_REPORT'

    def __init__(self, next_info):
        super().__init__(next_info)
