import itertools
from .reports_fetch_job import ReportsFetch

# To Add Segments 'date', 'device', 'ad_network_type_1'
# Note: If the number of custom paths exceed 7 in the subClasses. Move it to strategic pattern.
class KeywordPerformanceReportsJob(ReportsFetch):
    FIELDS = ['cost_per_conversion', 'all_conversion_value', 'all_conversions', 'cost', 'average_cost', 'average_cpc',
            'average_cpm', 'average_cpv','average_pageviews', 'average_time_on_site',  'click_assisted_conversion_value',
            'click_assisted_conversions', 'clicks', 'conversions', 'ctr', 'impression_assisted_conversions', 'impressions',
            'search_impression_share', 'search_top_impression_share', 'search_budget_lost_absolute_top_impression_share',
            'search_budget_lost_top_impression_share', 'search_rank_lost_absolute_top_impression_share',
            'search_rank_lost_impression_share', 'search_rank_lost_top_impression_share']
    
    SEGMENTS = ['id', 'ad_group_id', 'campaign_id']

    METRICS  = ['ad_group_name', 'ad_group_status', 'campaign_name', 'campaign_status', 'labels',
                'approval_status', 'average_position', 'criteria', 'keyword_match_type', 'is_negative', 'status',
                'cpc_bid', 'cpc_bid_source', 'first_position_cpc', 'first_page_cpc', 'top_of_page_cpc', 'quality_score']

    QUERY_FIELDS = list(itertools.chain(FIELDS, SEGMENTS, METRICS))
    REPORT = 'KEYWORDS_PERFORMANCE_REPORT'

    def __init__(self, next_info):
        super().__init__(next_info)
