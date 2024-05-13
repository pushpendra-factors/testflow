# A grammatically correct template is "Number of X in this week." when X is "users".  Can you suggest similar grammatically correct templates when X could be one of the following 16 phrases: ['New Users',
#  'Repeat Users',
#  'Engaged Sessions per user',
#  'Avg session duration',
#  'Engaged Users',
#  'Total time on site',
#  'Avg page views per session',
#  'Avg initial page load time',
#  'Engaged Sessions',
#  'Engagement rate',
#  'Total Sessions',
#  'Sessions',
#  'Unique Users',
#  'Users',
#  'Session Per User',
#  'Bounce rate']

# Generate one template each for each of the 16 phrases in the list. Replace "this week" with "Y".
# ChatGPT
# Sure, here are some grammatically correct templates based on the 16 phrases you provided:

# Number of New Users in Y.
# Number of Repeat Users in Y.
# Engaged Sessions per user during Y.
# Average session duration in Y.
# Number of Engaged Users in Y.
# Total time on site in Y.
# Average page views per session during Y.
# Average initial page load time during Y.
# Number of Engaged Sessions during Y.
# Engagement rate during Y.
# Total Sessions during Y.
# Number of Sessions in Y.
# Number of Unique Users in Y.
# Number of Users in Y.
# Average Session Per User in Y.
# Bounce rate during Y.






website_metrics_input_output = {
    'New Users': 'new_users',
    'Repeat Users': 'repeat_users',
    'Engaged Sessions per user': 'engaged_sessions_per_user',
    'Avg session duration': 'average_session_duration',
    'Engaged Users': 'engaged_users',
    'Total time on site': 'total_time_on_site',
    'Avg page views per session': 'average_page_views_per_session',
    'Avg initial page load time': 'average_initial_page_load_time',
    'Engaged Sessions': 'engaged_sessions',
    'Engagement rate': 'engagement_rate',
    'Total Sessions': 'total_sessions',
    'Sessions': 'total_sessions',
    'Unique Users': 'unique_users',
    'Users': 'unique_users',
    'Session Per User': 'sessions_per_user',
    'Bounce rate': 'bounce_rate',
}

website_dimensions_input_output = {
    'Source': '$source',
    'Medium': '$medium',
    'Campaign': '$campaign',
    'Adgroup': '$adgroup',
    'Keyword': '$keyword',
    'Content': '$content',
    'Channel': '$channel',
    'Referrer URL': '$initial_referrer_url',
    'Landing Page URL': '$initial_page_url',
    'Session Exit Page URL': '$session_latest_page_url',
    'OS': '$os',
    'OS Version': '$os_version',
    'Browser': '$browser',
    'Browser Version': '$browser_version',
    'Country': '$country',
    'Region': '$region',
    'City': '$city',
    'Timestamp': '$timestamp',
    'Session Time Spent': '$session_spent_time',
    'Landing Page Load Time': '$initial_page_load_time',
    'Landing Page Scroll Percent': '$initial_page_scroll_percent',
    'Type of content': 'Type of content',
    'Latest Page Scroll Percent': '$latest_page_scroll_percent',
    'Initial Page Domain': '$initial_page_domain',
    '6signal City': '$6Signal_city',
    'Initial Page Load Time': '$initial_page_load_time',
    'Device Manufacturer': '$device_manufacturer',
    'Browser Version': '$browser_version',
    'Email ID': '$email',
    'Latest Referrer Domain': '$latest_referrer_domain',
    'Latest Content': '$latest_content',
    '6signal Sic': '$6Signal_sic',
    'Initial Medium': '$initial_medium',
    'Latest Page URL': '$latest_page_url',
    '6signal Address': '$6Signal_address',
    'Postal Code': '$postal_code',
    'OS Version': '$os_version',
    '6signal Employee Range': '$6Signal_employee_range',
    'Latest Cost': '$latest_cost',
    'Initial Page Raw URL': '$initial_page_raw_url',
    'Initial Content': '$initial_content',
    'Device Type': '$device_type',
    'Device Brand': '$device_brand',
    'Total Revenue': '$total_revenue',
    'Initial Referrer URL': '$initial_referrer_url',
    'Page Count': '$page_count',
    'First Seen Time': '$joinTime',
    'Latest Campaign ID': '$latest_campaign_id',
    'Latest Page Spent Time': '$latest_page_spent_time',
    'Last Name': '$last_name',
    'Device Model': '$device_model',
    'Latest Page Raw URL': '$latest_page_raw_url',
    'Phone': '$phone',
    'Latest Medium': '$latest_medium',
    'Latest FBCLID': '$latest_fbclid',
    'Initial FBCLID': '$initial_fbclid',
    'Initial Campaign ID': '$initial_campaign_id',
    'Initial Source': '$initial_source',
    'Latest Adgroup ID': '$latest_adgroup_id',
    'Continent': '$continent',
    'Screen Height': '$screen_height',
    'Initial Page URL': '$initial_page_url',
    'Latest Source': '$latest_source',
    'Company': '$company',
    'Latest Page Load Time': '$latest_page_load_time',
    'Screen Width': '$screen_width',
    'Latest Adgroup': '$latest_adgroup',
    'Timezone': '$timezone',
    'Initial Campaign': '$initial_campaign',
    'Latest Channel': '$latest_channel',
    '6signal Name': '$6Signal_name',
    'Initial Revenue': '$initial_revenue',
    'Initial GCLID': '$initial_gclid',
    'Latest Term': '$latest_term',
    '6signal Country': '$6Signal_country',
    'City': '$city',
    'Screen Density': '$screen_density',
    'OS': '$os',
    'Initial Page Scroll Percent': '$initial_page_scroll_percent',
    'Latest Revenue': '$latest_revenue',
    'Device Family': '$device_family',
    'Initial Term': '$initial_term',
    'Day of First Event': '$day_of_first_event',
    '6signal Sic Description': '$6Signal_sic_description',
    'Latest Keyword Match UnitType': '$latest_keyword_match_type',
    '6signal Annual Revenue': '$6Signal_annual_revenue',
    'Latest Referrer URL': '$latest_referrer_url',
    'Latest Creative': '$latest_creative',
    '6signal Naics Description': '$6Signal_naics_description',
    'Session Spent Time': '$session_spent_time',
    'Total Cost': '$total_cost',
    'User ID': '$user_id',
    'Browser': '$browser',
    'Initial Adgroup': '$initial_adgroup',
    '6signal Zip': '$6Signal_zip',
    'Country': '$country',
    'Initial Keyword Match UnitType': '$initial_keyword_match_type',
    'Latest Referrer': '$latest_referrer',
    '6signal Employee Count': '$6Signal_employee_count',
    'Platform': '$platform',
    'Initial Referrer Domain': '$initial_referrer_domain',
    '6signal Revenue Range': '$6Signal_revenue_range',
    'Hour of First Event': '$hour_of_first_event',
    'Initial Cost': '$initial_cost',
    'Device Carrier': '$device_carrier',
    'First Name': '$first_name',
    '6signal Region': '$6Signal_region',
    'Latest Campaign': '$latest_campaign',
    '6signal Naics': '$6Signal_naics',
    '6signal State': '$6Signal_state',
    'Latest Page Domain': '$latest_page_domain',
    '6signal Country Iso Code': '$6Signal_country_iso_code',
    '6signal Phone': '$6Signal_phone',
    '6signal Domain': '$6Signal_domain',
    'Initial Channel': '$initial_channel',
    'Initial Creative': '$initial_creative',
    'Latest Keyword': '$latest_keyword',
    'Device Name': '$device_name',
    'Initial Referrer': '$initial_referrer',
    'Language': '$language',
    'Initial Adgroup ID': '$initial_adgroup_id',
    'Locale': '$locale',
    'Latest GCLID': '$latest_gclid',
    'Initial Keyword': '$initial_keyword',
    '6signal Industry': '$6Signal_industry',
    'Region': '$region',
    'Initial Page Spent Time': '$initial_page_spent_time',
    'Name': '$name'
}

# Total Sessions
# {"query_type": "kpi", "query_entity_1": "website_session", "query_filter_1":"none", 
# "query_breakdown_1": "none", "time_range": "default", "start_time": "default", "end_time": "default"}
for k, v in website_metrics_input_output.items():
    print('\n')
    print('%s' % k)
    print('{"query_type": "kpi", "query_entity_1": "%s", \
        "query_filter_1":"none", "query_breakdown_1": "none", \
        "time_range": "default", "start_time": "default", "end_time": "default"}' % v)

# Number of Sessions in the last month
# {"query_type": "kpi", "query_entity_1": "website_session",
# "query_filter_1":"none", "query_breakdown_1": "none", "time_range": "last_month", "start_time": "default", "end_time": "default"}
for k, v in website_metrics_input_output.items():
    print('\n')
    print('Number of %s in the last month' % k)
    print('{"query_type": "kpi", "query_entity_1": "%s", \
"query_filter_1":"none", "query_breakdown_1": "none", "time_range": "last_month", \
"start_time": "default", "end_time": "default"}' % v)

# Number of Sessions in the last week
# {"query_type": "kpi", "query_entity_1": "website_session",
# "query_filter_1":"none", "query_breakdown_1": "none", "time_range": "last_week", "start_time": "default", "end_time": "default"}
for k, v in website_metrics_input_output.items():
    print('\n')
    print('Number of %s in the last week' % k)
    print('{"query_type": "kpi", "query_entity_1": "%s", \
"query_filter_1":"none", "query_breakdown_1": "none", "time_range": "last_week", \
"start_time": "default", "end_time": "default"}' % v)

# How many new users we had this week?
# {"query_type": "kpi", "query_entity_1": "new_users", "query_filter_1":"none", "query_breakdown_1": "none", 
# "time_range": "this_week", "start_time": "default", "end_time": "default"}
for k, v in website_metrics_input_output.items():
    print('\n')
    print('How many %s we had this week?' % k)
    print('{"query_type": "kpi", "query_entity_1": "%s", "query_filter_1":"none", \
"query_breakdown_1": "none", "time_range": "this_week", "start_time": "default", \
"end_time": "default"}'% v)

# How many users visited our website this month?
# {"query_type": "kpi", "query_entity_1": "unique_users", "query_filter_1":"none", 
# "query_breakdown_1": "none", "time_range": "this_month", "start_time": "default", "end_time": "default"}
for k, v in website_metrics_input_output.items():
    print('\n')
    print('How many %s visited our website this month?' % k)
    print('{"query_type": "kpi", "query_entity_1": "%s", "query_filter_1":"none", \
"query_breakdown_1": "none", "time_range": "this_month", "start_time": "default", \
"end_time": "default"}'% v)