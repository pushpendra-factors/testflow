from .reports_fetch_job import ReportsFetch


# Note: If the number of custom paths exceed 7 in the subClasses. Move it to strategic pattern.
class CampaignPerformanceReportsJob(ReportsFetch):
    QUERY_FIELDS = ['active_view_impressions', 'active_view_measurability', 'active_view_measurable_cost',
                    'active_view_measurable_impressions', 'active_view_viewability', 'advertising_channel_sub_type',
                    'all_conversion_rate', 'conversion_rate', 'cost_per_conversion',
                    'all_conversion_value', 'all_conversions', 'amount', 'average_cost', 'average_position',
                    'average_time_on_site',
                    'base_campaign_id', 'bounce_rate', 'budget_id', 'campaign_id', 'campaign_name', 'campaign_status',
                    'campaign_trial_type', 'click_assisted_conversion_value',
                    'click_assisted_conversions', 'click_assisted_conversions_over_last_click_conversions', 'clicks',
                    'conversion_value', 'conversions',
                    'cost', 'start_date', 'end_date', 'engagements', 'gmail_forwards', 'gmail_saves',
                    'gmail_secondary_clicks', 'impression_assisted_conversions',
                    'impression_reach', 'impressions', 'interaction_types', 'interactions', 'invalid_clicks',
                    'is_budget_explicitly_shared', 'url_custom_parameters',
                    'value_per_all_conversion', 'video_quartile_100_rate', 'video_quartile_25_rate',
                    'video_quartile_50_rate', 'video_quartile_75_rate',
                    'video_view_rate', 'video_views', 'view_through_conversions']
    REPORT = 'CAMPAIGN_PERFORMANCE_REPORT'

    def __init__(self, next_info):
        super().__init__(next_info)
