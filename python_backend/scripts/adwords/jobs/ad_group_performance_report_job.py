from .reports_fetch_job import ReportsFetch


class AdGroupPerformanceReportJob(ReportsFetch):
    QUERY_FIELDS = ['active_view_impressions', 'active_view_measurability', 'active_view_measurable_cost',
                    'active_view_measurable_impressions', 'active_view_viewability', 'all_conversion_rate',
                    'conversion_rate', 'cost_per_conversion', 'all_conversion_value', 'all_conversions', 'average_cost',
                    'average_position', 'ad_group_id', 'ad_group_name', 'ad_group_status', 'base_ad_group_id',
                    'base_campaign_id', 'campaign_id', 'campaign_name', 'campaign_status', 'final_url_suffix',
                    'click_assisted_conversion_value', 'click_assisted_conversions',
                    'click_assisted_conversions_over_last_click_conversions', 'clicks',
                    'conversion_value', 'conversions', 'cost', 'engagements', 'gmail_forwards', 'gmail_saves',
                    'gmail_secondary_clicks', 'impression_assisted_conversions', 'impression_assisted_conversion_value',
                    'impressions', 'interaction_types', 'interactions', 'value_per_all_conversion',
                    'video_quartile_100_rate', 'video_quartile_25_rate', 'video_quartile_50_rate',
                    'video_quartile_75_rate', 'video_view_rate', 'video_views', 'view_through_conversions']

    REPORT = 'ADGROUP_PERFORMANCE_REPORT'

    def __init__(self, next_info):
        super().__init__(next_info)
