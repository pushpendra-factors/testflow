from .reports_fetch_job import ReportsFetch


# Note: If the number of custom paths exceed 7 in the subClasses. Move it to strategic pattern.
class KeywordPerformanceReportsJob(ReportsFetch):
    QUERY_FIELDS = ['id', 'ad_group_id', 'all_conversion_rate', 'conversion_rate',
                    'cost_per_conversion', 'all_conversion_value', 'all_conversions',
                    'approval_status', 'average_cost', 'average_cpc', 'average_cpm', 'average_cpv',
                    'average_pageviews', 'average_position', 'average_time_on_site', 'campaign_id',
                    'click_assisted_conversion_value',
                    'click_assisted_conversions', 'clicks', 'conversions', 'cpc_bid', 'cpc_bid_source', 'criteria',
                    'ctr', 'date', 'impression_assisted_conversions', 'impressions', 'keyword_match_type']
    REPORT = 'KEYWORDS_PERFORMANCE_REPORT'

    def __init__(self, next_info):
        super().__init__(next_info)
