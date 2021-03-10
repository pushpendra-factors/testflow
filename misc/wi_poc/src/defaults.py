# Project IDs of clients.
CHARGEBEE = '399'
HIPPOVIDEO = '427'
LIVSPACE = '386'
IZOOTO = '446'
DESIGNCAFE = '483'
ATHERENERGY = '498'

# Week information maps. TODO: To make this more dynamic.
AUG_WKS = ['01-07 Aug', '09-15 Aug', '16-22 Aug', '23-29 Aug', '30 Aug-05 Sep']
NOV_WKS = ['01-07 Nov', '08-14 Nov', '15-21 Nov', '22-28 Nov', '29 Nov-05 Dec']
DEC_WKS = ['29 Nov-05 Dec', '06-12 Dec', '13-19 Dec', '20-26 Dec', '27 Dec-02 Jan']
JAN_WKS = ['27 Dec-02 Jan', '03-09 Jan', '10-16 Jan', '17-23 Jan', '24-30 Jan', '31 Jan-06 Feb']
WKS = {'aug': AUG_WKS, 'nov': NOV_WKS, 'dec': DEC_WKS, 'jan': JAN_WKS}
WKS[8] = WKS['aug']
WKS[11] = WKS['nov']
WKS[12] = WKS['dec']
WKS[1] = WKS['jan']

# Some "base" and "target" event examples.
EVENTS_KEY = 'en'
FORM_SUBMITTED_KEY = '$form_submitted'
SESSION_KEY = '$session'
FORM_SUBMITTED_TARGET = (EVENTS_KEY, FORM_SUBMITTED_KEY, True)
SESSION_BASE = (EVENTS_KEY, SESSION_KEY, True)

# Some units of interest (UOI)
NUM_USERS_STR = '#users'
NUM_EVENTS_STR = '#events'

## SOME DEFAULTS
# Data defaults:
DEFAULT_PROJECT_ID = CHARGEBEE
DEFAULT_MODEL_ID = '1' # A sample test model curated for 5000 rows.
DEFAULT_TARGET = FORM_SUBMITTED_TARGET
DEFAULT_BASE = SESSION_BASE
DEFAULT_WK1_KEY = 'sample_wk1' # A sample test model curated for 5000 rows.
DEFAULT_WK2_KEY = 'sample_wk2' # A sample test model curated for 5000 rows.
DEFAULT_WK_PARAMS = {'project_id': DEFAULT_PROJECT_ID,
                     'model_id': DEFAULT_MODEL_ID}

# Feature-selection defaults:
CUSTOM_BLACKLIST = {'$initial_page_domain',
                    '$session_count',
                    '$initial_page_url',
                    '$page_title',
                    '$browser_version',
                    '$os_version',
                    '$user_agent',
                    '$gclid',
                    '$initial_gclid',
                    '$latest_gclid',
                    '$latest_fbclid',
                    '$fbclid',
                    '$initial_fbclid',
                    '$user_id',
                    '$day_of_first_event',
                    '$day_of_week',
                    '$identifiers',
                    '$initial_referrer',
                    '$initial_referrer_domain',
                    '$latest_page_raw_url',
                    '$initial_page_raw_url',
                    '$latest_referrer',
                    '$latest_referrer_domain',
                    '$name',
                    '$page_raw_url',
                    '$page_domain',
                    '$phone',
                    '$referrer',
                    '$referrer_raw_url',
                    '$referrer_domain',
                    '$salesforce_lead_createddate',
                    '$salesforce_lead_customer_whatsapp_optin__c',
                    '$salesforce_lead_date_when_meeting_is_scheduled__c',
                    '$salesforce_lead_email',
                    '$salesforce_lead_follow_up_date_time__c',
                    '$salesforce_lead_gclid__c',
                    '$salesforce_lead_id',
                    '$salesforce_lead_ipaddress__c',
                    '$salesforce_lead_lastmodifiedbyid',
                    '$salesforce_lead_lastmodifieddate',
                    '$salesforce_lead_lastname',
                    '$salesforce_lead_lead_allocation_time__c',
                    '$salesforce_lead_lead_qualified_date__c',
                    'salesforce_lead_mobile_number_external_field__c',
                    '$salesforce_lead_mobilephone',
                    '$salesforce_lead_mobileym__c',
                    '$salesforce_lead_page__c',
                    '$salesforce_lead_page_url__c',
                    '$salesforce_lead_pre_qualified_date__c',
                    '$session_latest_page_raw_url'}
DEFAULT_BLACKLISTS = {'custom', 'date'}
DEFAULT_MAX_NULL_PERC = 50
DEFAULT_MAX_UNIQ_PERC = 5
DEFAULT_MIN_UNIQ_COUNT = 2
DEFAULT_FEAT_SEL_PARAMS = {'max_null_perc': DEFAULT_MAX_NULL_PERC,
                           'max_uniq_perc': DEFAULT_MAX_UNIQ_PERC,
                           'min_uniq_count': DEFAULT_MIN_UNIQ_COUNT,
                           'blacklisted_features': DEFAULT_BLACKLISTS}

# Filter defaults:
DEFAULT_FILTERS_KEYS = ['nzc', 'zmc', 'msmcr', 'mscmcc', 'fifth', 'sixth', 'max_prev']
DEFAULT_FILTER_PARAMS = {'min_fm1': 10, 'min_fm2': 10,
                         'min_f': 50, 'min_crf': 0.10,
                         'min_fc_f': 0.10, 'min_fc_fm': 0.10, 'min_fm': 10,
                         'max_prev': 0.90}

# Metric defaults:
DEFAULT_DIFFERENCE_METRIC_NAMES = ['delta_ratio', 'jsd_f', 'jsd_fm',
                                   'impact', 'change_in_scale', 'change_in_cr']
DEFAULT_NEGATION_PREDICT_MODE = 'diff-based'

# Miscelleneous defaults:
DEFAULT_UNIT_OF_INTEREST = NUM_USERS_STR
DEFAULT_UID_FEAT = 'uid'
DEFAULT_DEDUPLICATE_LOGIC = 'upr'
DEFAULT_UOI = NUM_USERS_STR
DEFAULT_MERGE_UPR_LOGIC = 'cumulative'

# Model information maps. TODO: To make this more dynamic.
MODEL_INFO = {
    CHARGEBEE: {
        DEFAULT_WK1_KEY: ('1', 5000),
        DEFAULT_WK2_KEY: ('2', 5000),
        WKS['dec'][1]: ('1608449537302', 229231),
        WKS['dec'][2]: ('1608452239024', 235741),
        WKS['jan'][2]: ('1610914336618', 488452), 
        WKS['jan'][3]: ('1611470169895', 213435),
        WKS['jan'][4]: ('1612192854122', 235853),
        WKS['jan'][5]: ('1612779643168', None)},
    HIPPOVIDEO: {
        WKS['aug'][-3]: ('1598177830048', 72504),
        WKS['aug'][-2]: ('1598801710139', 85222)},
    LIVSPACE: {
        WKS['dec'][0]: ('1608707973413', 1572767),
        WKS['dec'][1]: ('1608727490184', 1527141),
        WKS['dec'][2]: ('1609346149884', 1909362),
        WKS['dec'][3]: ('1609386848912', 1579761),
        WKS['dec'][4]: (None, None)},
    IZOOTO: {
        WKS['dec'][0]: ('1607644474318', 261452),
        WKS['dec'][1]: ('1609783107212', 271565),
        WKS['dec'][2]: ('1609793254072', 243228),
        WKS['dec'][3]: ('1609802262328', 268491),
        WKS['dec'][4]: ('1609811114491', 302561),
        WKS['jan'][1]: ('1610519262142', 358134),
        WKS['jan'][2]: ('1610921586874', 323462),
        WKS['jan'][3]: ('1611586802465', 334745),
        WKS['jan'][4]: ('1612177506412', 304522),
        WKS['jan'][5]: ('1612786152964', None)},
    DESIGNCAFE: {
        WKS['dec'][0]: ('1607482505880', 98965),
        WKS['dec'][1]: ('1609779343714', 103357),
        WKS['dec'][2]: ('1609780522113', 98972),
        WKS['dec'][3]: ('1609781591132', 70183),
        WKS['dec'][4]: ('1609782315325', 73255),
        WKS['jan'][1]: ('1610439268600', 78418),
        WKS['jan'][2]: ('1610925758595', 99163),
        WKS['jan'][3]: ('1611578965671', 101656),
        WKS['jan'][4]: ('1612201698442', 102122),
        WKS['jan'][5]: ('1612785481047', None)},
    ATHERENERGY: {
        WKS['dec'][0]: (None, None),
        WKS['dec'][1]: (None, None),
        WKS['dec'][2]: (None, None),
        WKS['dec'][3]: (None, None),
        WKS['dec'][4]: ('1611197955978', 517594),
        WKS['jan'][1]: ('1611243664034', 665344),
        WKS['jan'][2]: ('1611245722155', 862609),
        WKS['jan'][3]: ('1611565936247', 1257064),
        WKS['jan'][4]: ('1612288805719', 2060272),
        WKS['jan'][5]: ('1612808039160', None)}}

# Specify criteria to select "base" and "target" users for the insights.
# TODO: Make this more dynamic.
criteria_map = {DESIGNCAFE: {'base': SESSION_BASE,
                             'target': ('en',
                                        '$sf_lead_created',
                                        True)},
                IZOOTO: {'base': SESSION_BASE,
                                #  (['en', '$page_url'],
                                #   ['$session', 'www.izooto.com/campaign/unsubscribing-from-web-push-notifications'],
                                #   [True, False]),
                         'target': (['en', '$page_url'],
                                    ['$form_submitted', 'panel.izooto.com/signup'],
                                    [True, True])},
                CHARGEBEE: {'base': SESSION_BASE,
                            'target': (['en', '$hubspot_contact_demo_booked_on'],
                                       ['$hubspot_contact_updated', None],
                                       [True, False])},
                LIVSPACE: {'base': SESSION_BASE,
                           'target': FORM_SUBMITTED_TARGET},
                ATHERENERGY: {'base': SESSION_BASE,
                              'target': (['en'],
                                         ['app.atherenergy.com/product/450x/testride/book/confirmation'],
                                         [True])}}
